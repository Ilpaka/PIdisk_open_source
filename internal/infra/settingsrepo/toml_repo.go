package settingsrepo

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
)

const fileName = "config.toml"

type Repository struct {
	mu   sync.Mutex
	path string
}

func New(configDir string) (*Repository, error) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return nil, err
	}
	return &Repository{path: filepath.Join(configDir, fileName)}, nil
}

func (r *Repository) Path() string { return r.path }

func (r *Repository) Load() (domain.AppSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := domain.DefaultAppSettings()
	raw, err := os.ReadFile(r.path)
	if errors.Is(err, os.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	if _, err := toml.Decode(string(raw), &out); err != nil {
		return out, err
	}
	return out, nil
}

func (r *Repository) Save(settings domain.AppSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tmp, err := os.CreateTemp(filepath.Dir(r.path), "config-*.toml")
	if err != nil {
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := toml.NewEncoder(tmp).Encode(settings); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), r.path)
}

var _ ports.SettingsRepository = (*Repository)(nil)
