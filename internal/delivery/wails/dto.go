package wailsapp

import (
	"github.com/pidisk/pidisk/internal/domain"
)

// ProfileInput is the form payload sent from CreateProfilePage.
type ProfileInput struct {
	Name           string `json:"name"`
	Host           string `json:"host"`
	Port           uint16 `json:"port"`
	Username       string `json:"username"`
	PrivateKeyPath string `json:"privateKeyPath"`
	Passphrase     string `json:"passphrase,omitempty"`
	RootDir        string `json:"rootDir"`
	TrashDir       string `json:"trashDir"`
	LocalSyncDir   string `json:"localSyncDir,omitempty"`
}

func (in ProfileInput) toDomain() domain.Profile {
	return domain.Profile{
		Name:           in.Name,
		Host:           in.Host,
		Port:           in.Port,
		Username:       in.Username,
		PrivateKeyPath: in.PrivateKeyPath,
		AuthMethod:     domain.AuthKey,
		RootDir:        in.RootDir,
		TrashDir:       in.TrashDir,
		LocalSyncDir:   in.LocalSyncDir,
	}
}
