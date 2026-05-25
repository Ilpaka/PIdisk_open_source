package syncwatcher

import (
	"testing"
	"time"
)

func snap(rel string, age time.Duration) FileSnap {
	return FileSnap{Rel: rel, Size: 1, MTime: time.Now().Add(age)}
}

func TestDiffNewLocalUploaded(t *testing.T) {
	local := map[string]FileSnap{"a.txt": snap("a.txt", 0)}
	remote := map[string]FileSnap{}
	res := Diff("default", local, remote, nil)
	if len(res.Uploads) != 1 || res.Uploads[0].Rel != "a.txt" {
		t.Fatalf("expected one upload, got %+v", res)
	}
	if len(res.Downloads) != 0 {
		t.Fatalf("expected no downloads")
	}
}

func TestDiffRemoteOnlyDownloaded(t *testing.T) {
	local := map[string]FileSnap{}
	remote := map[string]FileSnap{"b.txt": snap("b.txt", -5*time.Second)}
	res := Diff("default", local, remote, nil)
	if len(res.Downloads) != 1 || res.Downloads[0].Rel != "b.txt" {
		t.Fatalf("expected one download, got %+v", res)
	}
}

func TestDiffSameMTimeNoAction(t *testing.T) {
	now := time.Now()
	local := map[string]FileSnap{"c.txt": {Rel: "c.txt", MTime: now}}
	remote := map[string]FileSnap{"c.txt": {Rel: "c.txt", MTime: now}}
	res := Diff("default", local, remote, nil)
	if len(res.Uploads) != 0 || len(res.Downloads) != 0 {
		t.Fatalf("expected no actions, got %+v", res)
	}
}

func TestDiffLocalNewerWins(t *testing.T) {
	now := time.Now()
	local := map[string]FileSnap{"d.txt": {Rel: "d.txt", MTime: now}}
	remote := map[string]FileSnap{"d.txt": {Rel: "d.txt", MTime: now.Add(-10 * time.Second)}}
	res := Diff("default", local, remote, nil)
	if len(res.Uploads) != 1 {
		t.Fatalf("expected upload, got %+v", res)
	}
	if len(res.Conflicts) != 1 || res.Conflicts[0].Resolution != "local-wins" {
		t.Fatalf("expected local-wins conflict, got %+v", res.Conflicts)
	}
}

func TestDiffRemoteNewerWins(t *testing.T) {
	now := time.Now()
	local := map[string]FileSnap{"e.txt": {Rel: "e.txt", MTime: now.Add(-10 * time.Second)}}
	remote := map[string]FileSnap{"e.txt": {Rel: "e.txt", MTime: now}}
	res := Diff("default", local, remote, nil)
	if len(res.Downloads) != 1 {
		t.Fatalf("expected download, got %+v", res)
	}
	if len(res.Conflicts) != 1 || res.Conflicts[0].Resolution != "remote-wins" {
		t.Fatalf("expected remote-wins conflict, got %+v", res.Conflicts)
	}
}
