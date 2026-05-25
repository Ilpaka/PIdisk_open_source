package usecase

import (
	"context"
	"errors"
	"sync"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/infra/sshclient"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/rs/zerolog"
)

// ConnectionUseCase wires signer loading + keyring lookup + host-key prompt so
// callers can Connect/Disconnect by profile ID. It owns at most one live SSH
// client at a time (single-profile session model).
type ConnectionUseCase struct {
	profiles *ProfilesUseCase
	store    ports.KnownHostsStore
	bus      ports.EventEmitter
	broker   *HostKeyBroker
	log      zerolog.Logger

	mu      sync.RWMutex
	current ports.SSHClient
}

func NewConnectionUseCase(
	profiles *ProfilesUseCase,
	store ports.KnownHostsStore,
	bus ports.EventEmitter,
	broker *HostKeyBroker,
	log zerolog.Logger,
) *ConnectionUseCase {
	return &ConnectionUseCase{
		profiles: profiles,
		store:    store,
		bus:      bus,
		broker:   broker,
		log:      log.With().Str("component", "conn-uc").Logger(),
	}
}

// Connect activates the given profile: load passphrase, parse the key, dial,
// run the host-key callback. Returns once the session is fully established or
// the user rejected the host key.
func (uc *ConnectionUseCase) Connect(ctx context.Context, id domain.ProfileID, passphrase string) (ports.SSHClient, error) {
	profile, err := uc.profiles.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if passphrase == "" {
		stored, err := uc.profiles.Passphrase(id)
		if err == nil {
			passphrase = stored
		}
	}
	signer, err := sshclient.LoadSigner(profile.PrivateKeyPath, passphrase)
	if err != nil {
		if errors.Is(err, sshclient.ErrPassphraseRequired) {
			return nil, domain.ErrProfileLocked
		}
		return nil, err
	}
	if passphrase != "" {
		_ = uc.profiles.RememberPassphrase(id, passphrase)
	}

	client := sshclient.New(sshclient.Config{
		Profile:  profile,
		Signer:   signer,
		Store:    uc.store,
		Bus:      uc.bus,
		Logger:   uc.log,
		Prompter: uc.broker.Prompter(),
	})
	if err := client.Connect(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}

	uc.mu.Lock()
	if uc.current != nil {
		_ = uc.current.Close()
	}
	uc.current = client
	uc.mu.Unlock()

	uc.profiles.SetActive(profile)
	_ = uc.profiles.MarkUsed(ctx, profile.ID)
	return client, nil
}

func (uc *ConnectionUseCase) Disconnect() error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.current == nil {
		return nil
	}
	err := uc.current.Close()
	uc.current = nil
	uc.profiles.ClearActive()
	return err
}

func (uc *ConnectionUseCase) Current() (ports.SSHClient, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	if uc.current == nil || !uc.current.IsAlive() {
		return nil, false
	}
	return uc.current, true
}

func (uc *ConnectionUseCase) RemoteFS() (ports.RemoteFS, error) {
	c, ok := uc.Current()
	if !ok {
		return nil, domain.ErrNotConnected
	}
	fs := c.RemoteFS()
	if fs == nil {
		return nil, domain.ErrNotConnected
	}
	return fs, nil
}
