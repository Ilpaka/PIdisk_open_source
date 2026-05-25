package usecase

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/platform"
	"github.com/pidisk/pidisk/internal/ports"
)

// ProfilesUseCase owns the catalogue of SSH profiles and the notion of an
// "active" profile (the one the file manager is bound to).
type ProfilesUseCase struct {
	repo    ports.ProfileRepository
	secrets ports.SecretStore

	mu      sync.RWMutex
	current *domain.Profile
}

func NewProfilesUseCase(repo ports.ProfileRepository, secrets ports.SecretStore) *ProfilesUseCase {
	return &ProfilesUseCase{repo: repo, secrets: secrets}
}

func (uc *ProfilesUseCase) List(ctx context.Context) ([]domain.Profile, error) {
	out, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []domain.Profile{}
	}
	return out, nil
}

func (uc *ProfilesUseCase) Get(ctx context.Context, id domain.ProfileID) (domain.Profile, error) {
	return uc.repo.Get(ctx, id)
}

// CreateResult is the full outcome of profile creation. If a fresh SSH
// keypair was generated, GeneratedPublicKey carries the line the user needs
// to install on the server. It is empty when the user supplied an existing
// key path.
type CreateResult struct {
	Profile            domain.Profile `json:"profile"`
	GeneratedPublicKey string         `json:"generatedPublicKey,omitempty"`
	GeneratedKeyPath   string         `json:"generatedKeyPath,omitempty"`
}

// Create validates the input, ensures uniqueness by name, stores the profile
// and persists the passphrase in the keyring. Missing optional fields are
// filled with sensible defaults: remote root and trash are derived from the
// username; the local sync directory is created under ~/PIdiskSync; if no
// private key path was supplied a brand-new Ed25519 key is generated with a
// strong random passphrase, both saved to ~/.ssh and the OS keyring.
func (uc *ProfilesUseCase) Create(ctx context.Context, in domain.Profile, passphrase string) (CreateResult, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.AuthMethod == "" {
		in.AuthMethod = domain.AuthKey
	}
	if in.Port == 0 {
		in.Port = 22
	}

	// Pre-flight checks first so we do not leak generated key files on the
	// happy-but-rejected path (duplicate name, invalid host, etc.).
	if in.Name == "" {
		return CreateResult{}, domain.ErrInvalidProfileName
	}
	existing, err := uc.repo.List(ctx)
	if err != nil {
		return CreateResult{}, err
	}
	for _, p := range existing {
		if strings.EqualFold(p.Name, in.Name) {
			return CreateResult{}, domain.ErrProfileExists
		}
	}

	var generated platform.GeneratedKey
	keyWasGenerated := false
	if in.PrivateKeyPath == "" {
		gen, err := platform.GenerateProfileKey(in.Name)
		if err != nil {
			return CreateResult{}, err
		}
		generated = gen
		keyWasGenerated = true
		in.PrivateKeyPath = gen.PrivateKeyPath
		if passphrase == "" {
			passphrase = gen.Passphrase
		}
	}

	cleanupKey := func() {
		if !keyWasGenerated {
			return
		}
		_ = os.Remove(generated.PrivateKeyPath)
		_ = os.Remove(generated.PrivateKeyPath + ".pub")
	}

	if in.RootDir == "" {
		in.RootDir = platform.DefaultRemoteRoot(in.Username)
	}
	if in.TrashDir == "" {
		in.TrashDir = platform.DefaultRemoteTrash(in.RootDir)
	}
	if in.LocalSyncDir == "" {
		in.LocalSyncDir = platform.DefaultLocalSyncDir(in.Name)
	}
	if in.LocalSyncDir != "" {
		_ = os.MkdirAll(in.LocalSyncDir, 0o755)
	}
	if err := in.Validate(); err != nil {
		cleanupKey()
		return CreateResult{}, err
	}
	in.ID = domain.ProfileID(uuid.NewString())
	in.CreatedAt = time.Now().UTC()
	if err := uc.repo.Save(ctx, in); err != nil {
		cleanupKey()
		return CreateResult{}, err
	}
	if passphrase != "" {
		if err := uc.secrets.SetPassphrase(in.ID, passphrase); err != nil {
			_ = uc.repo.Delete(ctx, in.ID)
			cleanupKey()
			return CreateResult{}, err
		}
	}
	result := CreateResult{Profile: in}
	if keyWasGenerated {
		result.GeneratedPublicKey = generated.PublicKey
		result.GeneratedKeyPath = generated.PrivateKeyPath
	}
	return result, nil
}

// Update overwrites an existing profile. The ID must match.
func (uc *ProfilesUseCase) Update(ctx context.Context, in domain.Profile) (domain.Profile, error) {
	if in.ID == "" {
		return domain.Profile{}, errors.New("profile id required")
	}
	current, err := uc.repo.Get(ctx, in.ID)
	if err != nil {
		return domain.Profile{}, err
	}
	in.CreatedAt = current.CreatedAt
	if in.AuthMethod == "" {
		in.AuthMethod = domain.AuthKey
	}
	if err := in.Validate(); err != nil {
		return domain.Profile{}, err
	}
	if err := uc.repo.Save(ctx, in); err != nil {
		return domain.Profile{}, err
	}
	return in, nil
}

func (uc *ProfilesUseCase) Delete(ctx context.Context, id domain.ProfileID) error {
	// Look up first so we can clean up the auto-generated key file if any.
	profile, getErr := uc.repo.Get(ctx, id)
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = uc.secrets.Delete(id)
	if getErr == nil {
		platform.RemoveAutoGeneratedKey(profile.PrivateKeyPath)
	}
	uc.mu.Lock()
	if uc.current != nil && uc.current.ID == id {
		uc.current = nil
	}
	uc.mu.Unlock()
	return nil
}

// MarkUsed bumps LastUsedAt and persists.
func (uc *ProfilesUseCase) MarkUsed(ctx context.Context, id domain.ProfileID) error {
	p, err := uc.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	p.LastUsedAt = time.Now().UTC()
	return uc.repo.Save(ctx, p)
}

func (uc *ProfilesUseCase) Passphrase(id domain.ProfileID) (string, error) {
	return uc.secrets.GetPassphrase(id)
}

func (uc *ProfilesUseCase) RememberPassphrase(id domain.ProfileID, passphrase string) error {
	if passphrase == "" {
		return nil
	}
	return uc.secrets.SetPassphrase(id, passphrase)
}

func (uc *ProfilesUseCase) ForgetPassphrase(id domain.ProfileID) error {
	return uc.secrets.Delete(id)
}

func (uc *ProfilesUseCase) SetActive(p domain.Profile) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	cp := p
	uc.current = &cp
}

func (uc *ProfilesUseCase) ClearActive() {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.current = nil
}

func (uc *ProfilesUseCase) Active() (domain.Profile, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	if uc.current == nil {
		return domain.Profile{}, false
	}
	return *uc.current, true
}
