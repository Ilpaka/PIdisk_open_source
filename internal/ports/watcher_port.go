package ports

import (
	"context"
	"time"
)

type FSEventKind string

const (
	FSCreate FSEventKind = "create"
	FSModify FSEventKind = "modify"
	FSDelete FSEventKind = "delete"
	FSRename FSEventKind = "rename"
)

type FSEvent struct {
	Path string
	Kind FSEventKind
}

type FSWatcher interface {
	Watch(ctx context.Context, root string, debounce time.Duration) (<-chan FSEvent, error)
	Close() error
}
