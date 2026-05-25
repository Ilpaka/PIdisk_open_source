package syncwatcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pidisk/pidisk/internal/ports"
)

// Watcher wraps fsnotify with recursive watch (we walk subdirectories at
// startup) and debouncing.
type Watcher struct {
	mu      sync.Mutex
	watcher *fsnotify.Watcher
}

func NewWatcher() (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{watcher: w}, nil
}

func (w *Watcher) Close() error {
	if w.watcher == nil {
		return nil
	}
	return w.watcher.Close()
}

func (w *Watcher) Watch(ctx context.Context, root string, debounce time.Duration) (<-chan ports.FSEvent, error) {
	if debounce <= 0 {
		debounce = 500 * time.Millisecond
	}
	if err := w.addRecursive(root); err != nil {
		return nil, err
	}

	out := make(chan ports.FSEvent, 64)
	go func() {
		defer close(out)
		buffer := map[string]ports.FSEventKind{}
		flush := time.NewTimer(debounce)
		flush.Stop()
		flushPending := false

		emit := func() {
			for p, kind := range buffer {
				select {
				case <-ctx.Done():
					return
				case out <- ports.FSEvent{Path: p, Kind: kind}:
				}
			}
			buffer = map[string]ports.FSEventKind{}
			flushPending = false
		}

		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-w.watcher.Events:
				if !ok {
					return
				}
				kind := kindFromOp(ev.Op)
				buffer[ev.Name] = kind
				if !flushPending {
					flushPending = true
					flush.Reset(debounce)
				}
				if kind == ports.FSCreate {
					if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
						_ = w.addRecursive(ev.Name)
					}
				}
			case <-w.watcher.Errors:
				// errors are best-effort; the next event will retrigger normal flow
			case <-flush.C:
				emit()
			}
		}
	}()
	return out, nil
}

func (w *Watcher) addRecursive(root string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			_ = w.watcher.Add(path)
		}
		return nil
	})
}

func kindFromOp(op fsnotify.Op) ports.FSEventKind {
	switch {
	case op&fsnotify.Create != 0:
		return ports.FSCreate
	case op&fsnotify.Write != 0:
		return ports.FSModify
	case op&fsnotify.Remove != 0:
		return ports.FSDelete
	case op&fsnotify.Rename != 0:
		return ports.FSRename
	default:
		return ports.FSModify
	}
}

var _ ports.FSWatcher = (*Watcher)(nil)
