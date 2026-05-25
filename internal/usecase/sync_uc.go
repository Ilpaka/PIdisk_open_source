package usecase

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/infra/syncwatcher"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/rs/zerolog"
)

type SyncUseCase struct {
	conn     *ConnectionUseCase
	profiles *ProfilesUseCase
	repo     ports.SyncFolderRepository
	bus      ports.EventEmitter
	metric   ports.MetricsRecorder
	log      zerolog.Logger

	defaults []string
	interval uint64

	mu     sync.Mutex
	engine *syncwatcher.Engine
}

func NewSyncUseCase(
	conn *ConnectionUseCase,
	profiles *ProfilesUseCase,
	repo ports.SyncFolderRepository,
	bus ports.EventEmitter,
	metric ports.MetricsRecorder,
	log zerolog.Logger,
) *SyncUseCase {
	return &SyncUseCase{
		conn:     conn,
		profiles: profiles,
		repo:     repo,
		bus:      bus,
		metric:   metric,
		log:      log.With().Str("component", "sync-uc").Logger(),
	}
}

// Configure is called by SettingsUseCase whenever the user updates global
// sync settings (interval, default ignored paths).
func (uc *SyncUseCase) Configure(defaults []string, intervalSeconds uint64) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.defaults = defaults
	uc.interval = intervalSeconds
}

func (uc *SyncUseCase) ListFolders(ctx context.Context) ([]domain.SyncFolder, error) {
	active, ok := uc.profiles.Active()
	if !ok {
		return nil, domain.ErrNoActiveProfile
	}
	folders, err := uc.repo.List(ctx, active.ID)
	if err != nil {
		return nil, err
	}
	if folders == nil {
		folders = []domain.SyncFolder{}
	}
	return folders, nil
}

func (uc *SyncUseCase) AddFolder(ctx context.Context, folder domain.SyncFolder) (domain.SyncFolder, error) {
	folder.Name = strings.TrimSpace(folder.Name)
	if folder.Name == "" {
		return domain.SyncFolder{}, errors.New("sync folder name required")
	}
	if folder.LocalPath == "" || folder.RemotePath == "" {
		return domain.SyncFolder{}, errors.New("local and remote path required")
	}
	if folder.Direction == "" {
		folder.Direction = domain.SyncBoth
	}
	active, ok := uc.profiles.Active()
	if !ok {
		return domain.SyncFolder{}, domain.ErrNoActiveProfile
	}
	existing, err := uc.repo.List(ctx, active.ID)
	if err != nil {
		return domain.SyncFolder{}, err
	}
	for _, e := range existing {
		if strings.EqualFold(e.Name, folder.Name) {
			return domain.SyncFolder{}, domain.ErrSyncFolderExists
		}
	}
	if folder.LastSync.IsZero() {
		folder.LastSync = time.Time{}
	}
	if err := uc.repo.Save(ctx, active.ID, folder); err != nil {
		return domain.SyncFolder{}, err
	}
	return folder, nil
}

func (uc *SyncUseCase) RemoveFolder(ctx context.Context, name string) error {
	active, ok := uc.profiles.Active()
	if !ok {
		return domain.ErrNoActiveProfile
	}
	return uc.repo.Delete(ctx, active.ID, name)
}

func (uc *SyncUseCase) ToggleFolder(ctx context.Context, name string, enabled bool) (domain.SyncFolder, error) {
	active, ok := uc.profiles.Active()
	if !ok {
		return domain.SyncFolder{}, domain.ErrNoActiveProfile
	}
	folder, err := uc.repo.Get(ctx, active.ID, name)
	if err != nil {
		return domain.SyncFolder{}, err
	}
	folder.Enabled = enabled
	if err := uc.repo.Save(ctx, active.ID, folder); err != nil {
		return domain.SyncFolder{}, err
	}
	return folder, nil
}

func (uc *SyncUseCase) Start(ctx context.Context) error {
	uc.mu.Lock()
	if uc.engine != nil && uc.engine.IsRunning() {
		uc.mu.Unlock()
		return domain.ErrSyncAlreadyRunning
	}
	uc.mu.Unlock()

	folders, err := uc.ListFolders(ctx)
	if err != nil {
		return err
	}
	engine := syncwatcher.NewEngine(syncwatcher.Config{
		Provider:        uc.conn,
		Bus:             uc.bus,
		Logger:          uc.log,
		Metric:          uc.metric,
		DefaultIgnored:  uc.defaults,
		IntervalSeconds: uc.interval,
	})
	if err := engine.Start(ctx, folders); err != nil {
		return err
	}
	uc.mu.Lock()
	uc.engine = engine
	uc.mu.Unlock()
	return nil
}

func (uc *SyncUseCase) Stop(ctx context.Context) error {
	uc.mu.Lock()
	engine := uc.engine
	uc.engine = nil
	uc.mu.Unlock()
	if engine == nil {
		return domain.ErrSyncNotRunning
	}
	return engine.Stop(ctx)
}

func (uc *SyncUseCase) Status() domain.SyncStats {
	uc.mu.Lock()
	engine := uc.engine
	uc.mu.Unlock()
	if engine == nil {
		return domain.SyncStats{}
	}
	return engine.Stats()
}
