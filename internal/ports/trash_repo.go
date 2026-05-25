package ports

import (
	"context"

	"github.com/pidisk/pidisk/internal/domain"
)

type TrashRepository interface {
	Add(ctx context.Context, e domain.TrashEntry) error
	Get(ctx context.Context, id string) (domain.TrashEntry, error)
	List(ctx context.Context, profileID domain.ProfileID) ([]domain.TrashEntry, error)
	Delete(ctx context.Context, id string) error
	Clear(ctx context.Context, profileID domain.ProfileID) error
	Close() error
}
