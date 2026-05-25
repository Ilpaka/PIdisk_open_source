package sftpfs

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
)

// Download streams remotePath into localPath using SFTP concurrent reads.
// Progress is delivered through the same throttled pulse used by Upload.
func (f *FS) Download(ctx context.Context, remotePath, localPath string, progress ports.ProgressFunc) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	src, err := c.Open(remotePath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	total := info.Size()

	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return err
	}
	dst, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()

	if progress == nil {
		progress = func(int64, int64) {}
	}

	done := &atomic.Int64{}
	pr := &progressReader{r: src, done: done}

	tickCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go pulse(tickCtx, done, total, progress)

	if _, err := io.Copy(dst, pr); err != nil {
		return err
	}
	progress(done.Load(), total)
	return nil
}
