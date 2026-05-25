package usecase

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/platform"
	"github.com/rs/zerolog"
)

type FilesUseCase struct {
	conn *ConnectionUseCase
	log  zerolog.Logger
}

func NewFilesUseCase(conn *ConnectionUseCase, log zerolog.Logger) *FilesUseCase {
	return &FilesUseCase{conn: conn, log: log.With().Str("component", "files-uc").Logger()}
}

func validateName(name string) error {
	if name == "" {
		return domain.ErrInvalidPath
	}
	if strings.ContainsAny(name, "\x00\n\r/") {
		return domain.ErrInvalidPath
	}
	return nil
}

func validatePath(p string) error {
	if p == "" {
		return domain.ErrInvalidPath
	}
	if strings.ContainsAny(p, "\x00\n\r") {
		return domain.ErrInvalidPath
	}
	for _, segment := range strings.Split(p, "/") {
		if segment == ".." {
			return domain.ErrPathTraversal
		}
	}
	return nil
}

func (uc *FilesUseCase) ReadDir(ctx context.Context, p string) (domain.Listing, error) {
	if err := validatePath(p); err != nil {
		return domain.Listing{}, err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return domain.Listing{}, err
	}
	entries, err := fs.ReadDir(ctx, p)
	if err != nil {
		return domain.Listing{}, err
	}
	return domain.Listing{Path: p, Entries: entries}, nil
}

func (uc *FilesUseCase) Stat(ctx context.Context, p string) (domain.FileEntry, error) {
	if err := validatePath(p); err != nil {
		return domain.FileEntry{}, err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return domain.FileEntry{}, err
	}
	return fs.Stat(ctx, p)
}

func (uc *FilesUseCase) Mkdir(ctx context.Context, parent, name string) (string, error) {
	if err := validatePath(parent); err != nil {
		return "", err
	}
	if err := validateName(name); err != nil {
		return "", err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return "", err
	}
	full := path.Join(parent, name)
	if err := fs.Mkdir(ctx, full); err != nil {
		return "", err
	}
	return full, nil
}

func (uc *FilesUseCase) Move(ctx context.Context, src, dst string) error {
	if err := validatePath(src); err != nil {
		return err
	}
	if err := validatePath(dst); err != nil {
		return err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return err
	}
	return fs.Rename(ctx, src, dst)
}

func (uc *FilesUseCase) Rename(ctx context.Context, cwd, oldName, newName string) (string, error) {
	if err := validatePath(cwd); err != nil {
		return "", err
	}
	if err := validateName(oldName); err != nil {
		return "", err
	}
	if err := validateName(newName); err != nil {
		return "", err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return "", err
	}
	src := path.Join(cwd, oldName)
	dst := path.Join(cwd, newName)
	if err := fs.Rename(ctx, src, dst); err != nil {
		return "", err
	}
	return dst, nil
}

// Remove either deletes the target outright or, if a trash directory is
// configured for the active profile and the target lives outside trash,
// moves it there instead. The caller hands back trashed=true so the UI can
// distinguish the two outcomes.
func (uc *FilesUseCase) Remove(ctx context.Context, target string, opts RemoveOptions) (RemoveResult, error) {
	if err := validatePath(target); err != nil {
		return RemoveResult{}, err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return RemoveResult{}, err
	}
	if opts.TrashDir == "" || strings.HasPrefix(target, opts.TrashDir+"/") || target == opts.TrashDir {
		stat, err := fs.Stat(ctx, target)
		if err != nil {
			return RemoveResult{}, err
		}
		if stat.IsDir {
			return RemoveResult{Trashed: false}, fs.RemoveAll(ctx, target)
		}
		return RemoveResult{Trashed: false}, fs.Remove(ctx, target)
	}
	if err := fs.MkdirAll(ctx, opts.TrashDir); err != nil {
		return RemoveResult{}, err
	}
	stat, err := fs.Stat(ctx, target)
	if err != nil {
		return RemoveResult{}, err
	}
	trashedName := opts.TrashEntryID + "_" + path.Base(target)
	dst := path.Join(opts.TrashDir, trashedName)
	if err := fs.Rename(ctx, target, dst); err != nil {
		return RemoveResult{}, err
	}
	return RemoveResult{
		Trashed:     true,
		OriginalPath: target,
		TrashedPath:  dst,
		IsDir:        stat.IsDir,
		Size:         stat.Size,
	}, nil
}

type RemoveOptions struct {
	TrashDir     string
	TrashEntryID string
}

type RemoveResult struct {
	Trashed      bool
	OriginalPath string
	TrashedPath  string
	IsDir        bool
	Size         int64
}

// ClearTrash empties the active profile's trash directory.
func (uc *FilesUseCase) ClearTrash(ctx context.Context, trashDir string) error {
	if err := validatePath(trashDir); err != nil {
		return err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return err
	}
	entries, err := fs.ReadDir(ctx, trashDir)
	if errors.Is(err, domain.ErrNotConnected) {
		return err
	}
	if err != nil {
		// directory might not exist yet; nothing to clear
		return nil
	}
	for _, e := range entries {
		if e.IsDir {
			if err := fs.RemoveAll(ctx, e.Path); err != nil {
				return err
			}
			continue
		}
		if err := fs.Remove(ctx, e.Path); err != nil {
			return err
		}
	}
	return nil
}

// DiskUsage prefers the StatVFS extension and falls back to a `df -k`
// invocation when the server does not advertise it.
func (uc *FilesUseCase) DiskUsage(ctx context.Context, p string) (domain.DiskUsage, error) {
	if err := validatePath(p); err != nil {
		return domain.DiskUsage{}, err
	}
	fs, err := uc.conn.RemoteFS()
	if err != nil {
		return domain.DiskUsage{}, err
	}
	if du, err := fs.StatVFS(ctx, p); err == nil {
		return du, nil
	}
	client, ok := uc.conn.Current()
	if !ok {
		return domain.DiskUsage{}, domain.ErrNotConnected
	}
	sess, err := client.NewSession(ctx)
	if err != nil {
		return domain.DiskUsage{}, err
	}
	defer sess.Close()
	out, err := sess.Output("df -k " + platform.ShellEscape(p))
	if err != nil {
		return domain.DiskUsage{}, err
	}
	du, err := parseDf(string(out))
	if err != nil {
		return domain.DiskUsage{}, err
	}
	du.Path = p
	du.Raw = string(out)
	return du, nil
}

func parseDf(out string) (domain.DiskUsage, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return domain.DiskUsage{}, fmt.Errorf("unexpected df output")
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return domain.DiskUsage{}, fmt.Errorf("unexpected df fields")
	}
	total := atoiKB(fields[1]) * 1024
	used := atoiKB(fields[2]) * 1024
	free := atoiKB(fields[3]) * 1024
	percent := 0.0
	if total > 0 {
		percent = float64(used) / float64(total) * 100
	}
	return domain.DiskUsage{Used: used, Free: free, Total: total, Percent: percent}, nil
}

func atoiKB(s string) uint64 {
	var v uint64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			break
		}
		v = v*10 + uint64(ch-'0')
	}
	return v
}
