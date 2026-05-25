package wailsapp

import (
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/usecase"
)

type SyncBindings struct {
	app *App
	uc  *usecase.SyncUseCase
}

func NewSyncBindings(app *App, uc *usecase.SyncUseCase) *SyncBindings {
	return &SyncBindings{app: app, uc: uc}
}

func (b *SyncBindings) ListSyncFolders() ([]domain.SyncFolder, error) {
	return b.uc.ListFolders(b.app.Ctx())
}

func (b *SyncBindings) AddSyncFolder(folder domain.SyncFolder) (domain.SyncFolder, error) {
	return b.uc.AddFolder(b.app.Ctx(), folder)
}

func (b *SyncBindings) RemoveSyncFolder(name string) error {
	return b.uc.RemoveFolder(b.app.Ctx(), name)
}

func (b *SyncBindings) ToggleSyncFolder(name string, enabled bool) (domain.SyncFolder, error) {
	return b.uc.ToggleFolder(b.app.Ctx(), name, enabled)
}

func (b *SyncBindings) StartSync() error {
	return b.uc.Start(b.app.Ctx())
}

func (b *SyncBindings) StopSync() error {
	return b.uc.Stop(b.app.Ctx())
}

func (b *SyncBindings) GetSyncStatus() domain.SyncStats {
	return b.uc.Status()
}
