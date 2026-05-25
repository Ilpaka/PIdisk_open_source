# Profiles

A PIdisk profile describes one SSH endpoint plus the local-side state
needed to use it. Profiles are immutable after creation in terms of their
identity (`id` is a UUID assigned at create time); their fields can be
edited.

## Fields

| Field            | Purpose                                                 |
| ---------------- | ------------------------------------------------------- |
| `name`           | Display label on the login screen.                      |
| `host` / `port`  | SSH endpoint. Default port 22.                          |
| `username`       | Remote user.                                            |
| `privateKeyPath` | Absolute (or `~/`-prefixed) path to the SSH key file.   |
| `rootDir`        | Default remote directory the file manager opens in.     |
| `trashDir`       | Where deletions are moved when trash is enabled.        |
| `localSyncDir`   | Anchor for the sync engine; folders live underneath it. |

## Storage

- Metadata: bbolt at `<DataDir>/profiles.db`, bucket `profiles`, value
  `JSON(domain.Profile)`.
- Passphrase: OS keyring under service `com.pidisk.profiles`, account =
  profile id.
- Last-used timestamp is bumped on every successful unlock.

## Lifecycle

1. **Create.** `CreateProfilePage` collects fields, posts to
   `CreateProfile`. The usecase validates, stores metadata, then writes
   the passphrase to the keyring.
2. **Select.** From `LoginPage` the user picks a profile. If the keyring
   already has a passphrase we try to unlock immediately; otherwise we
   open `PassphrasePrompt` and call `Unlock(id, passphrase)`.
3. **Unlock.** `ConnectionUseCase.Connect` loads the signer, dials, runs
   the TOFU callback (which may prompt the user via `HostKeyDialog`), and
   opens the SFTP subsystem.
4. **Lock.** Closing a session calls `Lock`, which tears down the SSH
   client and clears the active profile; the keyring entry stays.
5. **Delete.** Removes both the bbolt record and the keyring entry.

## Choosing a key

PIdisk supports Ed25519 and RSA keys. Generate one with:

```
ssh-keygen -t ed25519 -f ~/.ssh/pidisk_ed25519
```

Set a passphrase when prompted; PIdisk will remember it in your keyring
after the first connection.
