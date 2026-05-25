package keystore

import (
	"errors"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/zalando/go-keyring"
)

const ServiceName = "com.pidisk.profiles"

type Keyring struct{}

func New() *Keyring { return &Keyring{} }

func (k *Keyring) SetPassphrase(id domain.ProfileID, passphrase string) error {
	if id == "" {
		return errors.New("empty profile id")
	}
	return keyring.Set(ServiceName, string(id), passphrase)
}

func (k *Keyring) GetPassphrase(id domain.ProfileID) (string, error) {
	v, err := keyring.Get(ServiceName, string(id))
	if errors.Is(err, keyring.ErrNotFound) {
		return "", domain.ErrProfileLocked
	}
	return v, err
}

func (k *Keyring) Delete(id domain.ProfileID) error {
	err := keyring.Delete(ServiceName, string(id))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

var _ ports.SecretStore = (*Keyring)(nil)
