package domain

import "time"

type ProfileID string

type AuthMethod string

const (
	AuthKey AuthMethod = "key"
)

// Profile describes how to connect to a single SFTP server.
// Secrets (passphrase) live outside this struct, inside the OS keyring.
type Profile struct {
	ID             ProfileID  `json:"id"`
	Name           string     `json:"name"`
	Host           string     `json:"host"`
	Port           uint16     `json:"port"`
	Username       string     `json:"username"`
	PrivateKeyPath string     `json:"privateKeyPath"`
	AuthMethod     AuthMethod `json:"authMethod"`
	RootDir        string     `json:"rootDir"`
	TrashDir       string     `json:"trashDir"`
	LocalSyncDir   string     `json:"localSyncDir"`
	CreatedAt      time.Time  `json:"createdAt"`
	LastUsedAt     time.Time  `json:"lastUsedAt"`
}

func (p Profile) Validate() error {
	if p.Name == "" {
		return ErrInvalidProfileName
	}
	if p.Host == "" {
		return ErrInvalidHost
	}
	if p.Port == 0 {
		return ErrInvalidPort
	}
	if p.Username == "" {
		return ErrInvalidUsername
	}
	if p.PrivateKeyPath == "" {
		return ErrInvalidKeyPath
	}
	if p.RootDir == "" {
		return ErrInvalidRootDir
	}
	if p.TrashDir == "" {
		return ErrInvalidTrashDir
	}
	return nil
}
