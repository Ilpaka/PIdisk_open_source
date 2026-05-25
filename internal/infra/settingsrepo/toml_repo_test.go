package settingsrepo

import (
	"testing"

	"github.com/pidisk/pidisk/internal/domain"
)

func TestLoadDefaultsWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	repo, err := New(dir)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	got, err := repo.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	defaults := domain.DefaultAppSettings()
	if got.Theme != defaults.Theme || got.SyncIntervalSeconds != defaults.SyncIntervalSeconds {
		t.Fatalf("expected defaults, got %+v", got)
	}
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo, err := New(dir)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	want := domain.AppSettings{
		Theme:               "dark",
		Language:            "ru",
		SyncIntervalSeconds: 15,
		PrometheusEnabled:   true,
		PrometheusAddr:      "127.0.0.1:9090",
		LogLevel:            "debug",
	}
	if err := repo.Save(want); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := repo.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Theme != want.Theme || got.SyncIntervalSeconds != want.SyncIntervalSeconds || got.PrometheusAddr != want.PrometheusAddr {
		t.Fatalf("roundtrip mismatch: %+v vs %+v", got, want)
	}
}
