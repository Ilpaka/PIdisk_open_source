package usecase

import (
	"context"
	"errors"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type TransferUseCase struct {
	conn   *ConnectionUseCase
	bus    ports.EventEmitter
	metric ports.MetricsRecorder
	log    zerolog.Logger

	mu       sync.RWMutex
	inflight map[domain.TransferID]*transferEntry
}

type transferEntry struct {
	progress domain.TransferProgress
	cancel   context.CancelFunc
	limiter  *rate.Limiter

	mu    sync.Mutex
	last  time.Time
	bytes int64
}

func NewTransferUseCase(conn *ConnectionUseCase, bus ports.EventEmitter, metric ports.MetricsRecorder, log zerolog.Logger) *TransferUseCase {
	if metric == nil {
		metric = ports.NoopMetrics{}
	}
	return &TransferUseCase{
		conn:     conn,
		bus:      bus,
		metric:   metric,
		log:      log.With().Str("component", "transfer-uc").Logger(),
		inflight: map[domain.TransferID]*transferEntry{},
	}
}

func (uc *TransferUseCase) Upload(ctx context.Context, localPath, remotePath string) (domain.TransferID, error) {
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return "", err
	}
	id := domain.TransferID(uuid.NewString())
	jobCtx, cancel := context.WithCancel(context.Background())
	entry := uc.register(id, domain.TransferProgress{
		ID:         id,
		Kind:       domain.KindUpload,
		Name:       filepath.Base(localPath),
		LocalPath:  localPath,
		RemotePath: remotePath,
		State:      domain.StateRunning,
		StartedAt:  time.Now().UTC(),
	}, cancel)

	go uc.runJob(jobCtx, entry, func() error {
		return fs.Upload(jobCtx, localPath, remotePath, uc.makeProgressCB(entry, domain.KindUpload))
	})
	// jobCtx is owned by runJob; ctx parameter is intentionally unused.
	_ = ctx
	return id, nil
}

func (uc *TransferUseCase) Download(ctx context.Context, remotePath, localPath string) (domain.TransferID, error) {
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return "", err
	}
	id := domain.TransferID(uuid.NewString())
	jobCtx, cancel := context.WithCancel(context.Background())
	entry := uc.register(id, domain.TransferProgress{
		ID:         id,
		Kind:       domain.KindDownload,
		Name:       path.Base(remotePath),
		LocalPath:  localPath,
		RemotePath: remotePath,
		State:      domain.StateRunning,
		StartedAt:  time.Now().UTC(),
	}, cancel)

	go uc.runJob(jobCtx, entry, func() error {
		return fs.Download(jobCtx, remotePath, localPath, uc.makeProgressCB(entry, domain.KindDownload))
	})
	_ = ctx
	return id, nil
}

func (uc *TransferUseCase) DownloadFolder(ctx context.Context, remoteRoot, localRoot string) (domain.TransferID, error) {
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return "", err
	}
	id := domain.TransferID(uuid.NewString())
	jobCtx, cancel := context.WithCancel(context.Background())
	entry := uc.register(id, domain.TransferProgress{
		ID:         id,
		Kind:       domain.KindDownloadDir,
		Name:       path.Base(remoteRoot),
		LocalPath:  localRoot,
		RemotePath: remoteRoot,
		State:      domain.StateRunning,
		StartedAt:  time.Now().UTC(),
	}, cancel)

	go uc.runJob(jobCtx, entry, func() error {
		return fs.DownloadFolder(jobCtx, remoteRoot, localRoot, uc.makeProgressCB(entry, domain.KindDownloadDir), ports.FolderOptions{})
	})
	_ = ctx
	return id, nil
}

func (uc *TransferUseCase) Cancel(id domain.TransferID) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	entry, ok := uc.inflight[id]
	if !ok {
		return domain.ErrTransferNotFound
	}
	entry.cancel()
	return nil
}

func (uc *TransferUseCase) Get(id domain.TransferID) (domain.TransferProgress, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	e, ok := uc.inflight[id]
	if !ok {
		return domain.TransferProgress{}, false
	}
	return e.progress, true
}

func (uc *TransferUseCase) List() []domain.TransferProgress {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	out := make([]domain.TransferProgress, 0, len(uc.inflight))
	for _, e := range uc.inflight {
		out = append(out, e.progress)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out
}

func (uc *TransferUseCase) register(id domain.TransferID, prog domain.TransferProgress, cancel context.CancelFunc) *transferEntry {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	entry := &transferEntry{
		progress: prog,
		cancel:   cancel,
		limiter:  rate.NewLimiter(rate.Every(200*time.Millisecond), 1),
		last:     time.Now(),
	}
	uc.inflight[id] = entry
	uc.metric.SetActiveTransfers(len(uc.inflight))
	uc.bus.Emit("transfer:started", prog)
	return entry
}

func (uc *TransferUseCase) finalise(entry *transferEntry, finalErr error) {
	uc.mu.Lock()
	entry.progress.CompletedAt = time.Now().UTC()
	switch {
	case finalErr == nil:
		entry.progress.State = domain.StateDone
		if entry.progress.BytesTotal > 0 {
			entry.progress.BytesDone = entry.progress.BytesTotal
		}
	case errors.Is(finalErr, context.Canceled):
		entry.progress.State = domain.StateCancelled
		entry.progress.Error = "cancelled"
	default:
		entry.progress.State = domain.StateError
		entry.progress.Error = finalErr.Error()
	}
	finalProgress := entry.progress
	delete(uc.inflight, entry.progress.ID)
	uc.metric.SetActiveTransfers(len(uc.inflight))
	uc.mu.Unlock()
	uc.bus.Emit("transfer:done", finalProgress)
}

func (uc *TransferUseCase) makeProgressCB(entry *transferEntry, kind domain.TransferKind) ports.ProgressFunc {
	var doneCounter atomic.Int64
	return func(done, total int64) {
		entry.mu.Lock()
		entry.progress.BytesDone = done
		entry.progress.BytesTotal = total
		now := time.Now()
		delta := done - entry.bytes
		dt := now.Sub(entry.last)
		if dt > 0 {
			entry.progress.Speed = int64(float64(delta) / dt.Seconds())
		}
		entry.bytes = done
		entry.last = now
		snapshot := entry.progress
		entry.mu.Unlock()

		switch kind {
		case domain.KindUpload:
			uc.metric.IncBytesUploaded(delta)
		case domain.KindDownload, domain.KindDownloadDir:
			uc.metric.IncBytesDownloaded(delta)
		}

		if done == total || entry.limiter.Allow() {
			uc.bus.Emit("transfer:progress", snapshot)
		}
		doneCounter.Store(done)
	}
}

func (uc *TransferUseCase) runJob(_ context.Context, entry *transferEntry, run func() error) {
	defer func() {
		if r := recover(); r != nil {
			uc.log.Error().Interface("panic", r).Msg("transfer panic")
		}
	}()
	err := run()
	uc.finalise(entry, err)
}
