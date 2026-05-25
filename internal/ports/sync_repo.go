package ports

import (
	"context"

	"github.com/pidisk/pidisk/internal/domain"
)

type SyncFolderRepository interface {
	List(ctx context.Context, profileID domain.ProfileID) ([]domain.SyncFolder, error)
	Get(ctx context.Context, profileID domain.ProfileID, name string) (domain.SyncFolder, error)
	Save(ctx context.Context, profileID domain.ProfileID, folder domain.SyncFolder) error
	Delete(ctx context.Context, profileID domain.ProfileID, name string) error
	Close() error
}
