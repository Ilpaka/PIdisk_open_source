package usecase

import (
	"sync"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
)

type SettingsUseCase struct {
	repo   ports.SettingsRepository
	metric ports.MetricsRecorder
	sync   *SyncUseCase

	mu      sync.RWMutex
	current domain.AppSettings
}

func NewSettingsUseCase(repo ports.SettingsRepository, metric ports.MetricsRecorder, syncUC *SyncUseCase) (*SettingsUseCase, error) {
	uc := &SettingsUseCase{repo: repo, metric: metric, sync: syncUC}
	loaded, err := repo.Load()
	if err != nil {
		return nil, err
	}
	uc.current = loaded
	uc.applyLocked()
	return uc, nil
}

func (uc *SettingsUseCase) Get() domain.AppSettings {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.current
}

func (uc *SettingsUseCase) Update(next domain.AppSettings) (domain.AppSettings, error) {
	if next.SyncIntervalSeconds == 0 {
		next.SyncIntervalSeconds = 30
	}
	if next.LogLevel == "" {
		next.LogLevel = "info"
	}
	if next.Theme == "" {
		next.Theme = "system"
	}
	if err := uc.repo.Save(next); err != nil {
		return domain.AppSettings{}, err
	}
	uc.mu.Lock()
	uc.current = next
	uc.applyLocked()
	uc.mu.Unlock()
	return next, nil
}

// applyLocked propagates settings to downstream components. Callers must hold
// the write lock (or be inside the constructor before any goroutines see uc).
func (uc *SettingsUseCase) applyLocked() {
	if uc.sync != nil {
		uc.sync.Configure(uc.current.DefaultIgnoredPaths, uc.current.SyncIntervalSeconds)
	}
}
