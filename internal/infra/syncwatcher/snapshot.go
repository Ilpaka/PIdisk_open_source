package syncwatcher

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/pidisk/pidisk/internal/ports"
	gitignore "github.com/sabhiram/go-gitignore"
)

func snapshotLocal(root string, ignorer *gitignore.GitIgnore) (map[string]FileSnap, error) {
	out := map[string]FileSnap{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if p == root {
			return nil
		}
		rel, rerr := filepath.Rel(root, p)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if ignorer != nil && ignorer.MatchesPath(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			return nil
		}
		out[rel] = FileSnap{Rel: rel, Size: info.Size(), MTime: info.ModTime()}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func snapshotRemote(ctx context.Context, fs ports.RemoteFS, root string) (map[string]FileSnap, error) {
	if err := fs.MkdirAll(ctx, root); err != nil {
		return nil, err
	}
	out := map[string]FileSnap{}
	return walkRemote(ctx, fs, root, root, out)
}

func walkRemote(ctx context.Context, fs ports.RemoteFS, root, current string, out map[string]FileSnap) (map[string]FileSnap, error) {
	entries, err := fs.ReadDir(ctx, current)
	if err != nil {
		return out, err
	}
	for _, e := range entries {
		if e.IsDir {
			if _, err := walkRemote(ctx, fs, root, e.Path, out); err != nil {
				return out, err
			}
			continue
		}
		rel := strings.TrimPrefix(e.Path, root+"/")
		rel = strings.TrimPrefix(rel, "/")
		out[rel] = FileSnap{Rel: rel, Size: e.Size, MTime: e.Modified}
	}
	return out, nil
}
