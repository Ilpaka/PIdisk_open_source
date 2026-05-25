package sftpfs

import (
	"context"
	"io"
	"os"
	"path"
	"sync/atomic"
	"time"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
)

// Upload streams localPath to remotePath in 32 KiB chunks and reports progress
// to the supplied callback. The callback is invoked at most every 200ms; the
// final call (with done == total) is always delivered.
func (f *FS) Upload(ctx context.Context, localPath, remotePath string, progress ports.ProgressFunc) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}
	total := info.Size()

	if dir := path.Dir(remotePath); dir != "." && dir != "/" {
		_ = c.MkdirAll(dir)
	}
	dst, err := c.OpenFile(remotePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer dst.Close()

	if progress == nil {
		progress = func(int64, int64) {}
	}

	done := &atomic.Int64{}
	pw := &progressWriter{w: dst, done: done}

	tickCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go pulse(tickCtx, done, total, progress)

	if _, err := io.Copy(pw, src); err != nil {
		return err
	}
	progress(done.Load(), total)
	return nil
}

type progressWriter struct {
	w    io.Writer
	done *atomic.Int64
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	if n > 0 {
		p.done.Add(int64(n))
	}
	return n, err
}

type progressReader struct {
	r    io.Reader
	done *atomic.Int64
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.done.Add(int64(n))
	}
	return n, err
}

func pulse(ctx context.Context, done *atomic.Int64, total int64, cb ports.ProgressFunc) {
	t := time.NewTicker(200 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cb(done.Load(), total)
		}
	}
}
