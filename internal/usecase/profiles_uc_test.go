package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/pidisk/pidisk/internal/domain"
)

type memRepo struct {
	mu   sync.Mutex
	rows map[domain.ProfileID]domain.Profile
}

func newMemRepo() *memRepo { return &memRepo{rows: map[domain.ProfileID]domain.Profile{}} }

func (r *memRepo) List(_ context.Context) ([]domain.Profile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Profile, 0, len(r.rows))
	for _, p := range r.rows {
		out = append(out, p)
	}
	return out, nil
}

func (r *memRepo) Get(_ context.Context, id domain.ProfileID) (domain.Profile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.rows[id]
	if !ok {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	return p, nil
}

func (r *memRepo) Save(_ context.Context, p domain.Profile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows[p.ID] = p
	return nil
}

func (r *memRepo) Delete(_ context.Context, id domain.ProfileID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rows, id)
	return nil
}

func (r *memRepo) Close() error { return nil }

type memSecrets struct {
	mu   sync.Mutex
	rows map[domain.ProfileID]string
}

func newMemSecrets() *memSecrets { return &memSecrets{rows: map[domain.ProfileID]string{}} }

func (s *memSecrets) SetPassphrase(id domain.ProfileID, p string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows[id] = p
	return nil
}

func (s *memSecrets) GetPassphrase(id domain.ProfileID) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.rows[id]
	if !ok {
		return "", domain.ErrProfileLocked
	}
	return v, nil
}

func (s *memSecrets) Delete(id domain.ProfileID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.rows, id)
	return nil
}

func sampleInput(name string) domain.Profile {
	return domain.Profile{
		Name:           name,
		Host:           "example.com",
		Port:           22,
		Username:       "user",
		PrivateKeyPath: "/key",
		RootDir:        "/srv",
		TrashDir:       "/srv/.trash",
	}
}

func TestCreateRejectsDuplicateName(t *testing.T) {
	uc := NewProfilesUseCase(newMemRepo(), newMemSecrets())
	if _, err := uc.Create(context.Background(), sampleInput("alpha"), "p"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := uc.Create(context.Background(), sampleInput("ALPHA"), "p"); !errors.Is(err, domain.ErrProfileExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestCreateStoresSecret(t *testing.T) {
	secrets := newMemSecrets()
	uc := NewProfilesUseCase(newMemRepo(), secrets)
	res, err := uc.Create(context.Background(), sampleInput("alpha"), "topsecret")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := uc.Passphrase(res.Profile.ID)
	if err != nil || got != "topsecret" {
		t.Fatalf("passphrase = %q err=%v", got, err)
	}
}

func TestDeleteWipesActive(t *testing.T) {
	uc := NewProfilesUseCase(newMemRepo(), newMemSecrets())
	res, _ := uc.Create(context.Background(), sampleInput("alpha"), "x")
	uc.SetActive(res.Profile)
	if err := uc.Delete(context.Background(), res.Profile.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := uc.Active(); ok {
		t.Fatalf("active should be cleared after delete")
	}
}

func TestValidateMissingHost(t *testing.T) {
	uc := NewProfilesUseCase(newMemRepo(), newMemSecrets())
	bad := sampleInput("alpha")
	bad.Host = ""
	if _, err := uc.Create(context.Background(), bad, "x"); !errors.Is(err, domain.ErrInvalidHost) {
		t.Fatalf("expected ErrInvalidHost, got %v", err)
	}
}
