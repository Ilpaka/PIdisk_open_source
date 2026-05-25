package wailsapp

import (
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/usecase"
)

type TrashBindings struct {
	app *App
	uc  *usecase.TrashUseCase
}

func NewTrashBindings(app *App, uc *usecase.TrashUseCase) *TrashBindings {
	return &TrashBindings{app: app, uc: uc}
}

func (b *TrashBindings) ListTrash() ([]domain.TrashEntry, error) {
	return b.uc.List(b.app.Ctx())
}

func (b *TrashBindings) RestoreFromTrash(id string) (domain.TrashEntry, error) {
	return b.uc.Restore(b.app.Ctx(), id)
}

func (b *TrashBindings) ClearAllTrash() error {
	return b.uc.ClearAll(b.app.Ctx())
}
