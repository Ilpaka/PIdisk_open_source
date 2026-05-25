# Security model

PIdisk is built to be opened publicly, so this document lays out the
threat model and the boundaries we draw.

## Authentication

- **Keys only.** Password authentication is intentionally disabled.
  Profiles store a path to a private key plus an optional passphrase. The
  passphrase lives in the OS keyring (`com.pidisk.profiles`); the private
  key file itself stays on disk where the user put it.
- **TOFU host keys.** The first connection to a host writes its key
  fingerprint to `known_hosts` in the user's config directory after the
  user confirms it in a dialog. Subsequent connections compare the
  incoming key against the saved one and refuse on mismatch
  (`domain.ErrHostKeyMismatch`).

## Path safety

- All file-manager operations go through `pkg/sftp` calls
  (`Mkdir`, `Rename`, `Remove`, `RemoveAll`). We never join user input
  into a shell command for these.
- The single place a shell command is constructed is the `df -k` fallback
  used by `DiskUsage` when the server does not implement
  `statvfs@openssh.com`. The path is escaped via
  `platform.ShellEscape`, which single-quotes the argument and escapes
  any embedded single quote.
- `usecase/files_uc.go` rejects paths containing `..` or NUL/CR/LF before
  forwarding them to the SFTP layer.

## At-rest data

- Profiles, trash and sync metadata sit in bbolt files inside the user
  data directory. The on-disk values are not encrypted; they contain no
  secret material (no passwords, no keys, no tokens).
- The keyring is the only secret store. We do not write encrypted
  fallback files; on Linux a Secret Service daemon
  (gnome-keyring/kwallet/libsecret) is required.

## Network surface

- The only outbound network connection is the SSH/SFTP session for the
  active profile.
- Prometheus metrics, when enabled in settings, expose `/metrics` on the
  loopback interface (`127.0.0.1:9101` by default). Change the bind
  address through `config.toml` if you want it reachable from elsewhere.

## What we will not ship

- Password-based SSH authentication.
- A "trust all hosts" toggle in the UI. TOFU plus explicit confirmation
  is the only happy path.
- Any telemetry that leaves the machine. The metrics endpoint is local
  and opt-in.

## Reporting issues

Please open a private security advisory on GitHub if you find a
vulnerability. Avoid filing public issues with reproduction details.
