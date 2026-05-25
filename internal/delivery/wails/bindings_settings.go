package wailsapp

import (
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/usecase"
)

type SettingsBindings struct {
	app *App
	uc  *usecase.SettingsUseCase
}

func NewSettingsBindings(app *App, uc *usecase.SettingsUseCase) *SettingsBindings {
	return &SettingsBindings{app: app, uc: uc}
}

func (b *SettingsBindings) GetAppSettings() domain.AppSettings {
	return b.uc.Get()
}

func (b *SettingsBindings) UpdateAppSettings(next domain.AppSettings) (domain.AppSettings, error) {
	return b.uc.Update(next)
}
