package ports

import (
	"context"
	"io"

	"github.com/pidisk/pidisk/internal/domain"
)

type RemoteSession interface {
	Output(cmd string) ([]byte, error)
	Close() error
}

// SSHClient abstracts a single live SSH connection.
type SSHClient interface {
	Profile() domain.Profile
	Connect(ctx context.Context) error
	Reconnect(ctx context.Context) error
	Close() error
	IsAlive() bool
	NewSession(ctx context.Context) (RemoteSession, error)
	RemoteFS() RemoteFS
}

// ClientManager keeps at most one live SSHClient per profile.
type ClientManager interface {
	Acquire(ctx context.Context, profileID domain.ProfileID) (SSHClient, error)
	Release(profileID domain.ProfileID) error
	Current() (SSHClient, bool)
	SetCurrent(profileID domain.ProfileID) error
	Close() error
}

// Ensure io is referenced in case future helpers extend the file.
var _ = io.EOF
