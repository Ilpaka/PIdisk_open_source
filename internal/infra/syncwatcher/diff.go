package syncwatcher

import (
	"time"

	"github.com/pidisk/pidisk/internal/domain"
	gitignore "github.com/sabhiram/go-gitignore"
)

// FileSnap captures the bits of a file we care about for sync.
type FileSnap struct {
	Rel   string
	Size  int64
	MTime time.Time
	IsDir bool
}

// DiffResult lists the actions sync needs to take.
type DiffResult struct {
	Uploads   []FileSnap
	Downloads []FileSnap
	Conflicts []domain.SyncConflict
}

// Tolerance for "same mtime" comparisons. FAT/ext have 2s precision, NTFS has
// 100ns; 2s is a safe upper bound that avoids endless ping-pong syncs.
const mtimeTolerance = 2 * time.Second

// Diff computes uploads, downloads and conflicts using last-writer-wins with
// the tolerance above. ignorer is consulted on the local set; entries it
// matches are skipped entirely on the upload path and ignored on the download
// path as well (we never want to pull a file the user excluded).
func Diff(folder string, local, remote map[string]FileSnap, ignorer *gitignore.GitIgnore) DiffResult {
	var res DiffResult

	for rel, l := range local {
		if ignorer != nil && ignorer.MatchesPath(rel) {
			continue
		}
		r, ok := remote[rel]
		if !ok {
			res.Uploads = append(res.Uploads, l)
			continue
		}
		if l.MTime.After(r.MTime.Add(mtimeTolerance)) {
			res.Uploads = append(res.Uploads, l)
			res.Conflicts = append(res.Conflicts, domain.SyncConflict{
				Folder:      folder,
				Path:        rel,
				LocalMTime:  l.MTime,
				RemoteMTime: r.MTime,
				Resolution:  "local-wins",
			})
		} else if r.MTime.After(l.MTime.Add(mtimeTolerance)) {
			res.Downloads = append(res.Downloads, r)
			res.Conflicts = append(res.Conflicts, domain.SyncConflict{
				Folder:      folder,
				Path:        rel,
				LocalMTime:  l.MTime,
				RemoteMTime: r.MTime,
				Resolution:  "remote-wins",
			})
		}
	}

	for rel, r := range remote {
		if _, ok := local[rel]; ok {
			continue
		}
		if ignorer != nil && ignorer.MatchesPath(rel) {
			continue
		}
		res.Downloads = append(res.Downloads, r)
	}
	return res
}
