# Architecture

PIdisk is structured as a hexagonal-ish Go application with a Wails-bound
React UI. The runtime is a single process; the only cross-process boundary
is the WebKit/WebView2 webview that hosts the frontend.

## Layers

```
domain     pure types, no dependencies
usecase    business logic, depends on ports only
ports      interfaces implemented by infra
infra      adapters: SSH, SFTP, bbolt, keyring, fsnotify, Prometheus
delivery   Wails bindings, DTOs
platform   OS-aware paths and shell escaping
version    build-time metadata
```

The frontend mirrors the same separation. `src/api/*` wraps the generated
Wails bindings, `src/stores/*` holds Zustand stores, `src/pages/*` and
`src/components/*` hold view code.

## Dependency direction

```
delivery -> usecase -> ports -> domain
infra ---^                    ^
                              |
                              + (only usecase imports ports)
```

`usecase` never imports `infra`. The DI graph is wired manually in `main.go`,
which is intentionally the only place that knows about every layer.

## Event flow

Wails events fan out from `infra/eventbus` (which wraps
`runtime.EventsEmit`). The frontend subscribes via `src/api/events.ts`. We
deliberately do not use polling for connection state, transfers, sync status
or trash refreshes; everything flows through events.

## Storage layout

| Data           | Location                          | Format        |
| -------------- | --------------------------------- | ------------- |
| Profiles       | `<DataDir>/profiles.db`           | bbolt + JSON  |
| Trash entries  | `<DataDir>/trash.db`              | bbolt + JSON  |
| Sync folders   | `<DataDir>/sync.db`               | bbolt + JSON  |
| App settings   | `<ConfigDir>/config.toml`         | TOML          |
| Known hosts    | `<ConfigDir>/known_hosts`         | OpenSSH lines |
| Logs           | `<LogDir>/pidisk.log`             | JSON lines    |
| Secrets        | OS keyring (`com.pidisk.profiles`)| native        |

`DataDir`, `ConfigDir`, `LogDir` are resolved via `internal/platform`. macOS
points all three at `~/Library/Application Support/PIdisk`, Linux follows
XDG, Windows uses `%APPDATA%\PIdisk`.

## SFTP transport

`infra/sftpfs` configures `pkg/sftp` with `MaxPacket=32k`,
`MaxConcurrentRequestsPerFile=64`, concurrent reads/writes enabled. Uploads
stream via `io.Copy` with a `progressWriter`; downloads use a
`progressReader`. `DownloadFolder` walks the remote tree, runs an errgroup
with a worker semaphore (8), and reports an aggregate progress callback.

## SSH lifecycle

`infra/sshclient/Client` dials the profile, registers a TOFU host-key
callback, opens the SFTP subsystem, and starts a keepalive ticker
(`keepalive@openssh.com` every 30s, 10s timeout). On failure the ticker calls
`Reconnect`, which closes the existing connection and retries with
exponential backoff (1s start, 30s cap).

## Sync engine

`infra/syncwatcher/Engine` runs one goroutine per enabled folder. Each
folder performs:

1. an immediate full diff (`snapshotLocal` + `snapshotRemote` + `Diff`),
2. fsnotify-driven incremental triggers (debounced 500ms),
3. a periodic full diff on a `time.Ticker` (configurable, default 30s).

Conflicts are resolved last-writer-wins with a 2s mtime tolerance and are
emitted as `sync:conflict` events for the UI to inspect.
