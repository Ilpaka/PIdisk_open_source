package sftpfs

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/pkg/sftp"
	"golang.org/x/sync/errgroup"
)

// DownloadFolder mirrors remoteRoot into localRoot. It first walks the tree
// to compute total bytes, then downloads files with up to opts.MaxWorkers
// parallel streams. progress receives the running total across all files.
func (f *FS) DownloadFolder(ctx context.Context, remoteRoot, localRoot string, progress ports.ProgressFunc, opts ports.FolderOptions) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	if progress == nil {
		progress = func(int64, int64) {}
	}
	if opts.MaxWorkers <= 0 {
		opts.MaxWorkers = 8
	}

	type job struct {
		remote string
		local  string
	}
	var (
		jobs      []job
		totalSize int64
	)

	walker := c.Walk(remoteRoot)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}
		info := walker.Stat()
		rel := relativise(remoteRoot, walker.Path())
		dest := filepath.Join(localRoot, filepath.FromSlash(rel))
		if info.IsDir() {
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
			continue
		}
		jobs = append(jobs, job{remote: walker.Path(), local: dest})
		totalSize += info.Size()
	}

	var done atomic.Int64
	progress(0, totalSize)

	group, gctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, opts.MaxWorkers)
	for _, j := range jobs {
		j := j
		select {
		case <-gctx.Done():
			return gctx.Err()
		case sem <- struct{}{}:
		}
		group.Go(func() error {
			defer func() { <-sem }()
			return downloadOne(gctx, c, j.remote, j.local, &done, totalSize, progress)
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}
	progress(totalSize, totalSize)
	return nil
}

func downloadOne(ctx context.Context, c *sftp.Client, remote, local string, done *atomic.Int64, total int64, progress ports.ProgressFunc) error {
	src, err := c.Open(remote)
	if err != nil {
		return err
	}
	defer src.Close()
	if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
		return err
	}
	dst, err := os.OpenFile(local, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()

	buf := make([]byte, 64*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, rerr := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
			done.Add(int64(n))
			progress(done.Load(), total)
		}
		if rerr == io.EOF {
			return nil
		}
		if rerr != nil {
			return rerr
		}
	}
}

func relativise(root, child string) string {
	root = path.Clean(root)
	child = path.Clean(child)
	if root == child {
		return path.Base(child)
	}
	if len(child) <= len(root) || child[:len(root)+1] != root+"/" {
		return path.Base(child)
	}
	return child[len(root)+1:]
}
