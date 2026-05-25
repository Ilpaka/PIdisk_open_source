package ports

import "github.com/pidisk/pidisk/internal/domain"

// SecretStore wraps the OS keyring. Implementations must be safe for
// concurrent use; the keyring backends already serialise on their own mutex.
type SecretStore interface {
	SetPassphrase(id domain.ProfileID, passphrase string) error
	GetPassphrase(id domain.ProfileID) (string, error)
	Delete(id domain.ProfileID) error
}
