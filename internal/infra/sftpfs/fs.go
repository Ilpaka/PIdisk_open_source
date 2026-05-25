package sftpfs

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"sync"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// FS wraps a *sftp.Client and implements ports.RemoteFS.
type FS struct {
	mu     sync.RWMutex
	client *sftp.Client
}

func New(sshClient *ssh.Client) (*FS, error) {
	c, err := sftp.NewClient(sshClient,
		sftp.MaxPacket(SFTPMaxPacket),
		sftp.MaxConcurrentRequestsPerFile(SFTPMaxConcurrentReqs),
		sftp.UseConcurrentReads(true),
		sftp.UseConcurrentWrites(true),
	)
	if err != nil {
		return nil, err
	}
	return &FS{client: c}, nil
}

// Raw returns the underlying *sftp.Client for advanced callers (transfer, walk).
func (f *FS) Raw() *sftp.Client {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.client
}

func (f *FS) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.client == nil {
		return nil
	}
	err := f.client.Close()
	f.client = nil
	return err
}

func (f *FS) Stat(_ context.Context, p string) (domain.FileEntry, error) {
	c := f.Raw()
	if c == nil {
		return domain.FileEntry{}, domain.ErrNotConnected
	}
	info, err := c.Stat(p)
	if err != nil {
		return domain.FileEntry{}, err
	}
	return entryFromFileInfo(p, info), nil
}

func (f *FS) ReadDir(_ context.Context, p string) ([]domain.FileEntry, error) {
	c := f.Raw()
	if c == nil {
		return nil, domain.ErrNotConnected
	}
	infos, err := c.ReadDir(p)
	if err != nil {
		return nil, err
	}
	out := make([]domain.FileEntry, 0, len(infos))
	for _, info := range infos {
		out = append(out, entryFromFileInfo(path.Join(p, info.Name()), info))
	}
	return out, nil
}

func (f *FS) Mkdir(_ context.Context, p string) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	return c.Mkdir(p)
}

func (f *FS) MkdirAll(_ context.Context, p string) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	return c.MkdirAll(p)
}

func (f *FS) Rename(_ context.Context, src, dst string) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	return c.Rename(src, dst)
}

func (f *FS) Remove(_ context.Context, p string) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	return c.Remove(p)
}

func (f *FS) RemoveAll(_ context.Context, p string) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	info, err := c.Stat(p)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return c.Remove(p)
	}
	entries, err := c.ReadDir(p)
	if err != nil {
		return err
	}
	for _, e := range entries {
		child := path.Join(p, e.Name())
		if e.IsDir() {
			if err := f.RemoveAll(context.Background(), child); err != nil {
				return err
			}
			continue
		}
		if err := c.Remove(child); err != nil {
			return err
		}
	}
	return c.RemoveDirectory(p)
}

func (f *FS) Open(_ context.Context, p string) (io.ReadCloser, int64, error) {
	c := f.Raw()
	if c == nil {
		return nil, 0, domain.ErrNotConnected
	}
	file, err := c.Open(p)
	if err != nil {
		return nil, 0, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, 0, err
	}
	return file, info.Size(), nil
}

func (f *FS) Create(_ context.Context, p string) (io.WriteCloser, error) {
	c := f.Raw()
	if c == nil {
		return nil, domain.ErrNotConnected
	}
	return c.Create(p)
}

func (f *FS) Walk(ctx context.Context, root string, fn ports.WalkFunc) error {
	c := f.Raw()
	if c == nil {
		return domain.ErrNotConnected
	}
	walker := c.Walk(root)
	for walker.Step() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := walker.Err(); err != nil {
			if cont := fn(walker.Path(), domain.FileEntry{Path: walker.Path()}, err); cont != nil {
				return cont
			}
			continue
		}
		info := walker.Stat()
		entry := entryFromFileInfo(walker.Path(), info)
		if err := fn(walker.Path(), entry, nil); err != nil {
			if errors.Is(err, errStopWalk) {
				return nil
			}
			return err
		}
	}
	return nil
}

var errStopWalk = errors.New("stop walk")

// StatVFS tries the OpenSSH-only statvfs@openssh.com extension; if the server
// does not advertise it the caller should fall back to a plain `df` session.
func (f *FS) StatVFS(_ context.Context, p string) (domain.DiskUsage, error) {
	c := f.Raw()
	if c == nil {
		return domain.DiskUsage{}, domain.ErrNotConnected
	}
	st, err := c.StatVFS(p)
	if err != nil {
		return domain.DiskUsage{}, err
	}
	total := st.TotalSpace()
	free := st.FreeSpace()
	used := uint64(0)
	if total > free {
		used = total - free
	}
	percent := 0.0
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}
	return domain.DiskUsage{
		Path:    p,
		Used:    used,
		Free:    free,
		Total:   total,
		Percent: percent,
	}, nil
}

// Upload and Download live in upload.go / download.go.

func entryFromFileInfo(absPath string, info os.FileInfo) domain.FileEntry {
	return domain.FileEntry{
		Name:     info.Name(),
		Path:     absPath,
		IsDir:    info.IsDir(),
		Size:     info.Size(),
		Modified: info.ModTime().UTC(),
		Mode:     uint32(info.Mode()),
	}
}
