# Sync

The sync engine mirrors a list of local folders against a remote root over
SFTP. It is designed to be polite to the server (one watch per folder, no
busy loops) and predictable to the user (LWW conflict resolution with a
2-second tolerance window).

## Folder definition

A `SyncFolder` ties a local path to a remote path:

```json
{
  "name": "photos",
  "localPath": "/Users/me/PIdisk/photos",
  "remotePath": "/srv/me/photos",
  "enabled": true,
  "direction": "both"
}
```

Folders are stored per profile in `<DataDir>/sync.db`.

## Loop

For each enabled folder the engine runs:

1. `snapshotLocal(root, ignorer)` walks the local tree and records
   `(rel, size, mtime)` for every file the ignorer does not match.
2. `snapshotRemote(ctx, fs, root)` recursively `ReadDir`s the remote.
3. `Diff(folder, local, remote, ignorer)` returns `Uploads`, `Downloads`
   and `Conflicts`.
4. Uploads/downloads are executed sequentially via the SFTP layer; each
   conflict is emitted as `sync:conflict`.
5. Aggregate counters are bumped and `sync:status` is emitted.

A fsnotify watcher on the local tree feeds incremental triggers (debounced
500ms); a `time.Ticker` runs the full diff on the configured interval
(default 30s, configurable in settings).

## Conflict resolution

`mtimeTolerance = 2s`. When both sides have the same path:

- local mtime is newer by more than the tolerance: upload, log a
  `local-wins` conflict.
- remote mtime is newer by more than the tolerance: download, log a
  `remote-wins` conflict.
- otherwise: do nothing.

The 2-second window guards against filesystems with coarse mtime
precision (FAT, some NFS shares) so we do not ping-pong files forever.

## Ignore rules

The engine reads `.pidiskignore` (gitignore syntax) from the local folder
root and combines it with the default list configured in app settings.

A starter file lives at `.pidiskignore.example` in the repo:

```
.DS_Store
Thumbs.db
*.tmp
*.swp
node_modules/
.git/
```

Files matched by the ignorer are skipped both on upload and on download.

## Events

| Event              | Payload                                       |
| ------------------ | --------------------------------------------- |
| `sync:status`      | `SyncStats` (used by the UI panel)            |
| `sync:job-completed` | `SyncEvent` (folder, files, bytes, duration) |
| `sync:job-error`   | `SyncEvent` with `error`                      |
| `sync:conflict`    | `SyncConflict`                                |
