package ports

import "github.com/pidisk/pidisk/internal/domain"

type SettingsRepository interface {
	Load() (domain.AppSettings, error)
	Save(settings domain.AppSettings) error
	Path() string
}
