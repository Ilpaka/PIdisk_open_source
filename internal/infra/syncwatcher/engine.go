package syncwatcher

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/rs/zerolog"
)

// Engine runs the sync loop for the active profile. Each enabled folder gets
// its own goroutine with a fsnotify watcher and a periodic full diff.
type Engine struct {
	conn     RemoteFSProvider
	bus      ports.EventEmitter
	log      zerolog.Logger
	metric   ports.MetricsRecorder
	defaults []string
	interval time.Duration

	mu      sync.Mutex
	running atomic.Bool
	stop    context.CancelFunc
	stats   domain.SyncStats
	folders map[string]*folderJob
}

// RemoteFSProvider is what the engine needs from the connection layer: a way
// to obtain the current SFTP filesystem. We use an interface so tests can
// pull in a memory FS.
type RemoteFSProvider interface {
	RemoteFS() (ports.RemoteFS, error)
}

type folderJob struct {
	folder  domain.SyncFolder
	cancel  context.CancelFunc
	watcher *Watcher
}

type Config struct {
	Provider         RemoteFSProvider
	Bus              ports.EventEmitter
	Logger           zerolog.Logger
	Metric           ports.MetricsRecorder
	DefaultIgnored   []string
	IntervalSeconds  uint64
}

func NewEngine(cfg Config) *Engine {
	if cfg.Metric == nil {
		cfg.Metric = ports.NoopMetrics{}
	}
	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Engine{
		conn:     cfg.Provider,
		bus:      cfg.Bus,
		log:      cfg.Logger.With().Str("component", "sync").Logger(),
		metric:   cfg.Metric,
		defaults: cfg.DefaultIgnored,
		interval: interval,
		folders:  map[string]*folderJob{},
	}
}

func (e *Engine) Start(parent context.Context, folders []domain.SyncFolder) error {
	if e.running.Load() {
		return domain.ErrSyncAlreadyRunning
	}
	ctx, cancel := context.WithCancel(parent)
	e.mu.Lock()
	e.stop = cancel
	e.stats = domain.SyncStats{IsRunning: true, LastSyncTime: time.Now().UTC()}
	e.mu.Unlock()
	e.running.Store(true)

	for _, f := range folders {
		if !f.Enabled {
			continue
		}
		if err := e.startFolder(ctx, f); err != nil {
			e.log.Warn().Err(err).Str("folder", f.Name).Msg("failed to start folder")
		}
	}
	e.emitStats()
	return nil
}

func (e *Engine) Stop(_ context.Context) error {
	if !e.running.Load() {
		return domain.ErrSyncNotRunning
	}
	e.mu.Lock()
	for name, job := range e.folders {
		job.cancel()
		if job.watcher != nil {
			_ = job.watcher.Close()
		}
		delete(e.folders, name)
	}
	if e.stop != nil {
		e.stop()
		e.stop = nil
	}
	e.stats.IsRunning = false
	e.mu.Unlock()
	e.running.Store(false)
	e.emitStats()
	return nil
}

func (e *Engine) IsRunning() bool { return e.running.Load() }

func (e *Engine) Stats() domain.SyncStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := e.stats
	out.SyncedFolders = append([]string{}, e.stats.SyncedFolders...)
	out.Errors = append([]string{}, e.stats.Errors...)
	out.Conflicts = append([]domain.SyncConflict{}, e.stats.Conflicts...)
	return out
}

func (e *Engine) startFolder(ctx context.Context, folder domain.SyncFolder) error {
	if folder.LocalPath == "" || folder.RemotePath == "" {
		return errors.New("sync folder requires both local and remote path")
	}
	watcher, err := NewWatcher()
	if err != nil {
		return err
	}
	jobCtx, cancel := context.WithCancel(ctx)
	job := &folderJob{folder: folder, cancel: cancel, watcher: watcher}
	e.mu.Lock()
	e.folders[folder.Name] = job
	e.mu.Unlock()
	go e.runFolder(jobCtx, job)
	return nil
}

func (e *Engine) runFolder(ctx context.Context, job *folderJob) {
	defer func() {
		if job.watcher != nil {
			_ = job.watcher.Close()
		}
	}()
	if err := os.MkdirAll(job.folder.LocalPath, 0o755); err != nil {
		e.recordError(job.folder.Name, err)
		return
	}
	events, err := job.watcher.Watch(ctx, job.folder.LocalPath, 500*time.Millisecond)
	if err != nil {
		e.recordError(job.folder.Name, err)
		return
	}
	tick := time.NewTicker(e.interval)
	defer tick.Stop()

	e.runOnce(ctx, job)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			e.runOnce(ctx, job)
		case _, ok := <-events:
			if !ok {
				return
			}
			e.runOnce(ctx, job)
		}
	}
}

func (e *Engine) runOnce(ctx context.Context, job *folderJob) {
	start := time.Now()
	e.metric.IncSyncRun()

	fs, err := e.conn.RemoteFS()
	if err != nil {
		e.recordError(job.folder.Name, err)
		return
	}

	ignorer := LoadIgnore(job.folder.LocalPath, e.defaults)
	localMap, err := snapshotLocal(job.folder.LocalPath, ignorer)
	if err != nil {
		e.recordError(job.folder.Name, err)
		return
	}
	remoteMap, err := snapshotRemote(ctx, fs, job.folder.RemotePath)
	if err != nil {
		e.recordError(job.folder.Name, err)
		return
	}

	result := Diff(job.folder.Name, localMap, remoteMap, ignorer)
	var (
		filesSynced uint32
		bytesSynced uint64
	)

	for _, up := range result.Uploads {
		localPath := filepath.Join(job.folder.LocalPath, filepath.FromSlash(up.Rel))
		remotePath := path.Join(job.folder.RemotePath, up.Rel)
		if err := fs.Upload(ctx, localPath, remotePath, nil); err != nil {
			e.recordError(job.folder.Name, err)
			continue
		}
		filesSynced++
		bytesSynced += uint64(up.Size)
	}
	for _, dl := range result.Downloads {
		localPath := filepath.Join(job.folder.LocalPath, filepath.FromSlash(dl.Rel))
		remotePath := path.Join(job.folder.RemotePath, dl.Rel)
		if err := fs.Download(ctx, remotePath, localPath, nil); err != nil {
			e.recordError(job.folder.Name, err)
			continue
		}
		filesSynced++
		bytesSynced += uint64(dl.Size)
	}
	for _, c := range result.Conflicts {
		e.bus.Emit("sync:conflict", c)
	}

	dur := time.Since(start)
	e.recordRun(job.folder.Name, filesSynced, bytesSynced, dur)
	e.bus.Emit("sync:job-completed", domain.SyncEvent{
		Folder:      job.folder.Name,
		FilesSynced: filesSynced,
		BytesSynced: bytesSynced,
		DurationMs:  dur.Milliseconds(),
		StartedAt:   start,
	})
	e.emitStats()
}

func (e *Engine) recordRun(name string, files uint32, bytes uint64, _ time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !contains(e.stats.SyncedFolders, name) {
		e.stats.SyncedFolders = append(e.stats.SyncedFolders, name)
		sort.Strings(e.stats.SyncedFolders)
	}
	e.stats.FilesSynced += files
	e.stats.BytesSynced += bytes
	e.stats.LastSyncTime = time.Now().UTC()
}

func (e *Engine) recordError(folder string, err error) {
	if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
		return
	}
	e.metric.IncSyncError()
	e.mu.Lock()
	defer e.mu.Unlock()
	msg := folder + ": " + err.Error()
	e.stats.Errors = append(e.stats.Errors, msg)
	if len(e.stats.Errors) > 50 {
		e.stats.Errors = e.stats.Errors[len(e.stats.Errors)-50:]
	}
	e.bus.Emit("sync:job-error", domain.SyncEvent{
		Folder: folder,
		Error:  err.Error(),
	})
	e.emitStatsLocked()
}

func (e *Engine) emitStats() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.emitStatsLocked()
}

func (e *Engine) emitStatsLocked() {
	if e.bus == nil {
		return
	}
	snapshot := e.stats
	snapshot.SyncedFolders = append([]string{}, e.stats.SyncedFolders...)
	snapshot.Errors = append([]string{}, e.stats.Errors...)
	snapshot.Conflicts = append([]domain.SyncConflict{}, e.stats.Conflicts...)
	e.bus.Emit("sync:status", snapshot)
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
