package wailsapp

import (
	"path"
	"strings"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/usecase"
)

type FileBindings struct {
	app      *App
	files    *usecase.FilesUseCase
	trash    *usecase.TrashUseCase
	profiles *usecase.ProfilesUseCase
}

func NewFileBindings(app *App, files *usecase.FilesUseCase, trash *usecase.TrashUseCase, profiles *usecase.ProfilesUseCase) *FileBindings {
	return &FileBindings{app: app, files: files, trash: trash, profiles: profiles}
}

func (b *FileBindings) ReadDir(p string) (domain.Listing, error) {
	if p == "" {
		if active, ok := b.profiles.Active(); ok {
			p = active.RootDir
		} else {
			return domain.Listing{}, domain.ErrNoActiveProfile
		}
	}
	return b.files.ReadDir(b.app.Ctx(), p)
}

func (b *FileBindings) Mkdir(parent, name string) (string, error) {
	if parent == "" {
		if active, ok := b.profiles.Active(); ok {
			parent = active.RootDir
		}
	}
	return b.files.Mkdir(b.app.Ctx(), parent, name)
}

func (b *FileBindings) Move(src, dst string) error {
	return b.files.Move(b.app.Ctx(), src, dst)
}

func (b *FileBindings) Rename(cwd, oldName, newName string) (string, error) {
	return b.files.Rename(b.app.Ctx(), cwd, oldName, newName)
}

type RemoveResultDTO struct {
	Trashed      bool   `json:"trashed"`
	OriginalPath string `json:"originalPath,omitempty"`
	TrashedPath  string `json:"trashedPath,omitempty"`
	IsDir        bool   `json:"isDir"`
	Size         int64  `json:"size"`
}

func (b *FileBindings) Remove(target string) (RemoveResultDTO, error) {
	active, ok := b.profiles.Active()
	if !ok {
		return RemoveResultDTO{}, domain.ErrNoActiveProfile
	}
	// Inside the trash directory we delete outright, otherwise we move to trash.
	inTrash := active.TrashDir != "" &&
		(target == active.TrashDir || strings.HasPrefix(target, active.TrashDir+"/"))
	if inTrash {
		res, err := b.files.Remove(b.app.Ctx(), target, usecase.RemoveOptions{})
		if err != nil {
			return RemoveResultDTO{}, err
		}
		return RemoveResultDTO{
			Trashed:      false,
			OriginalPath: target,
			IsDir:        res.IsDir,
			Size:         res.Size,
		}, nil
	}
	entry, err := b.trash.MoveToTrash(b.app.Ctx(), target)
	if err != nil {
		return RemoveResultDTO{}, err
	}
	return RemoveResultDTO{
		Trashed:      true,
		OriginalPath: entry.OriginalPath,
		TrashedPath:  entry.TrashedPath,
		IsDir:        entry.IsDir,
		Size:         entry.Size,
	}, nil
}

func (b *FileBindings) ClearTrash() error {
	return b.trash.ClearAll(b.app.Ctx())
}

func (b *FileBindings) DiskUsage(p string) (domain.DiskUsage, error) {
	if p == "" {
		if active, ok := b.profiles.Active(); ok {
			p = active.RootDir
		}
	}
	return b.files.DiskUsage(b.app.Ctx(), p)
}

// Join is a convenience used by the frontend to keep path joining consistent
// with the server (forward-slash semantics rather than Windows).
func (b *FileBindings) Join(parent, child string) string {
	return path.Join(parent, child)
}
