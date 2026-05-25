package trashrepo

import (
	"context"
	"testing"
	"time"

	"github.com/pidisk/pidisk/internal/domain"
)

func TestAddListDelete(t *testing.T) {
	dir := t.TempDir()
	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	entry := domain.TrashEntry{
		ID:           "abc",
		OriginalPath: "/srv/foo",
		TrashedPath:  "/srv/.trash/abc_foo",
		DeletedAt:    time.Now(),
		ProfileID:    "profile-1",
	}
	if err := repo.Add(ctx, entry); err != nil {
		t.Fatalf("add: %v", err)
	}
	got, err := repo.Get(ctx, "abc")
	if err != nil || got.OriginalPath != entry.OriginalPath {
		t.Fatalf("get: %+v err=%v", got, err)
	}
	list, err := repo.List(ctx, "profile-1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %+v err=%v", list, err)
	}
	if err := repo.Delete(ctx, "abc"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.Get(ctx, "abc"); err == nil {
		t.Fatalf("expected ErrTrashEntryNotFound")
	}
}

func TestClearByProfile(t *testing.T) {
	dir := t.TempDir()
	repo, err := Open(dir)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer repo.Close()
	ctx := context.Background()

	_ = repo.Add(ctx, domain.TrashEntry{ID: "a", ProfileID: "p1"})
	_ = repo.Add(ctx, domain.TrashEntry{ID: "b", ProfileID: "p2"})
	if err := repo.Clear(ctx, "p1"); err != nil {
		t.Fatalf("clear: %v", err)
	}
	left, _ := repo.List(ctx, "")
	if len(left) != 1 || left[0].ID != "b" {
		t.Fatalf("expected only p2 entry left, got %+v", left)
	}
}
