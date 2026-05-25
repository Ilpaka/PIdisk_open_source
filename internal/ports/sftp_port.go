package ports

import (
	"context"
	"io"

	"github.com/pidisk/pidisk/internal/domain"
)

type WalkFunc func(path string, entry domain.FileEntry, err error) error

type ProgressFunc func(done, total int64)

type RemoteFS interface {
	Stat(ctx context.Context, p string) (domain.FileEntry, error)
	ReadDir(ctx context.Context, p string) ([]domain.FileEntry, error)
	Mkdir(ctx context.Context, p string) error
	MkdirAll(ctx context.Context, p string) error
	Rename(ctx context.Context, src, dst string) error
	Remove(ctx context.Context, p string) error
	RemoveAll(ctx context.Context, p string) error
	Open(ctx context.Context, p string) (io.ReadCloser, int64, error)
	Create(ctx context.Context, p string) (io.WriteCloser, error)
	Walk(ctx context.Context, root string, fn WalkFunc) error
	StatVFS(ctx context.Context, p string) (domain.DiskUsage, error)
	Upload(ctx context.Context, localPath, remotePath string, progress ProgressFunc) error
	Download(ctx context.Context, remotePath, localPath string, progress ProgressFunc) error
	DownloadFolder(ctx context.Context, remoteRoot, localRoot string, progress ProgressFunc, opts FolderOptions) error
}

type FolderOptions struct {
	MaxWorkers int
}
