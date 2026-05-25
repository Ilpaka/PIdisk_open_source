package usecase

import (
	"context"
	"path"
	"sort"

	"github.com/google/uuid"
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/rs/zerolog"
)

type TrashUseCase struct {
	conn     *ConnectionUseCase
	repo     ports.TrashRepository
	profiles *ProfilesUseCase
	bus      ports.EventEmitter
	log      zerolog.Logger
}

func NewTrashUseCase(conn *ConnectionUseCase, repo ports.TrashRepository, profiles *ProfilesUseCase, bus ports.EventEmitter, log zerolog.Logger) *TrashUseCase {
	return &TrashUseCase{
		conn:     conn,
		repo:     repo,
		profiles: profiles,
		bus:      bus,
		log:      log.With().Str("component", "trash-uc").Logger(),
	}
}

// MoveToTrash renames the target into the active profile's trash directory and
// records the move in the trash bucket. The caller (FilesUseCase) handles the
// actual fs.Rename to keep all rename logic in one place; trash is just a
// metadata wrapper that lets Restore put files back later.
func (uc *TrashUseCase) MoveToTrash(ctx context.Context, original string) (domain.TrashEntry, error) {
	active, ok := uc.profiles.Active()
	if !ok {
		return domain.TrashEntry{}, domain.ErrNoActiveProfile
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return domain.TrashEntry{}, err
	}
	if err := fs.MkdirAll(ctx, active.TrashDir); err != nil {
		return domain.TrashEntry{}, err
	}
	stat, err := fs.Stat(ctx, original)
	if err != nil {
		return domain.TrashEntry{}, err
	}
	id := uuid.NewString()
	trashedName := id + "_" + path.Base(original)
	trashedPath := path.Join(active.TrashDir, trashedName)
	if err := fs.Rename(ctx, original, trashedPath); err != nil {
		return domain.TrashEntry{}, err
	}
	entry := domain.TrashEntry{
		ID:           id,
		OriginalPath: original,
		TrashedPath:  trashedPath,
		DeletedAt:    stat.Modified,
		ProfileID:    active.ID,
		IsDir:        stat.IsDir,
		Size:         stat.Size,
	}
	if err := uc.repo.Add(ctx, entry); err != nil {
		return domain.TrashEntry{}, err
	}
	uc.bus.Emit("trash:updated", struct{}{})
	return entry, nil
}

// Restore moves a trashed entry back to its original path. If the parent
// directory no longer exists, it is created.
func (uc *TrashUseCase) Restore(ctx context.Context, id string) (domain.TrashEntry, error) {
	entry, err := uc.repo.Get(ctx, id)
	if err != nil {
		return domain.TrashEntry{}, err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return entry, err
	}
	if parent := path.Dir(entry.OriginalPath); parent != "." && parent != "/" {
		_ = fs.MkdirAll(ctx, parent)
	}
	if err := fs.Rename(ctx, entry.TrashedPath, entry.OriginalPath); err != nil {
		return entry, err
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return entry, err
	}
	uc.bus.Emit("trash:updated", struct{}{})
	return entry, nil
}

// List returns the trash entries for the active profile, newest first.
func (uc *TrashUseCase) List(ctx context.Context) ([]domain.TrashEntry, error) {
	active, ok := uc.profiles.Active()
	if !ok {
		return nil, domain.ErrNoActiveProfile
	}
	entries, err := uc.repo.List(ctx, active.ID)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].DeletedAt.After(entries[j].DeletedAt) })
	if entries == nil {
		entries = []domain.TrashEntry{}
	}
	return entries, nil
}

func (uc *TrashUseCase) ClearAll(ctx context.Context) error {
	active, ok := uc.profiles.Active()
	if !ok {
		return domain.ErrNoActiveProfile
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return err
	}
	if err := fs.RemoveAll(ctx, active.TrashDir); err != nil {
		return err
	}
	if err := uc.repo.Clear(ctx, active.ID); err != nil {
		return err
	}
	uc.bus.Emit("trash:updated", struct{}{})
	return nil
}
