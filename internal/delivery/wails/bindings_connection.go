package wailsapp

import (
	"github.com/google/uuid"
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/usecase"
)

type ConnectionBindings struct {
	app    *App
	conn   *usecase.ConnectionUseCase
	broker *usecase.HostKeyBroker
}

func NewConnectionBindings(app *App, conn *usecase.ConnectionUseCase, broker *usecase.HostKeyBroker) *ConnectionBindings {
	return &ConnectionBindings{app: app, conn: conn, broker: broker}
}

type UnlockResult struct {
	Profile   domain.Profile `json:"profile"`
	Connected bool           `json:"connected"`
}

func (b *ConnectionBindings) Unlock(profileID, passphrase string) (UnlockResult, error) {
	client, err := b.conn.Connect(b.app.Ctx(), domain.ProfileID(profileID), passphrase)
	if err != nil {
		return UnlockResult{}, err
	}
	return UnlockResult{Profile: client.Profile(), Connected: client.IsAlive()}, nil
}

func (b *ConnectionBindings) Lock() error {
	return b.conn.Disconnect()
}

func (b *ConnectionBindings) IsConnected() bool {
	_, ok := b.conn.Current()
	return ok
}

// ConfirmHostKey is called by the UI in response to a `hostkey:prompt` event.
func (b *ConnectionBindings) ConfirmHostKey(fingerprint string, accept bool) {
	b.broker.Decide(fingerprint, accept)
}

// NewTrashID is a small helper exposed to the frontend so the UI can generate
// stable IDs ahead of time (used when calling Remove with trash semantics).
func (b *ConnectionBindings) NewTrashID() string {
	return uuid.NewString()
}
