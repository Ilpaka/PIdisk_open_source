package ports

import (
	"context"

	"github.com/pidisk/pidisk/internal/domain"
)

type ProfileRepository interface {
	List(ctx context.Context) ([]domain.Profile, error)
	Get(ctx context.Context, id domain.ProfileID) (domain.Profile, error)
	Save(ctx context.Context, p domain.Profile) error
	Delete(ctx context.Context, id domain.ProfileID) error
	Close() error
}
