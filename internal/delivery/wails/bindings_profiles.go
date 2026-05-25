package wailsapp

import (
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/platform"
	"github.com/pidisk/pidisk/internal/usecase"
)

// ActiveProfileResult is the structured response returned by ActiveProfile.
// We avoid (Profile, bool) tuples because the Wails generator only handles a
// single value plus an error.
type ActiveProfileResult struct {
	Profile domain.Profile `json:"profile"`
	Present bool           `json:"present"`
}

type ProfileBindings struct {
	app *App
	uc  *usecase.ProfilesUseCase
}

func NewProfileBindings(app *App, uc *usecase.ProfilesUseCase) *ProfileBindings {
	return &ProfileBindings{app: app, uc: uc}
}

func (b *ProfileBindings) ListProfiles() ([]domain.Profile, error) {
	return b.uc.List(b.app.Ctx())
}

func (b *ProfileBindings) GetProfile(id string) (domain.Profile, error) {
	return b.uc.Get(b.app.Ctx(), domain.ProfileID(id))
}

func (b *ProfileBindings) CreateProfile(in ProfileInput) (usecase.CreateResult, error) {
	return b.uc.Create(b.app.Ctx(), in.toDomain(), in.Passphrase)
}

func (b *ProfileBindings) DeleteProfile(id string) error {
	return b.uc.Delete(b.app.Ctx(), domain.ProfileID(id))
}

func (b *ProfileBindings) SetActiveProfile(id string) (domain.Profile, error) {
	p, err := b.uc.Get(b.app.Ctx(), domain.ProfileID(id))
	if err != nil {
		return domain.Profile{}, err
	}
	b.uc.SetActive(p)
	_ = b.uc.MarkUsed(b.app.Ctx(), p.ID)
	return p, nil
}

func (b *ProfileBindings) ClearActiveProfile() {
	b.uc.ClearActive()
}

func (b *ProfileBindings) ActiveProfile() ActiveProfileResult {
	p, ok := b.uc.Active()
	return ActiveProfileResult{Profile: p, Present: ok}
}

func (b *ProfileBindings) HasPassphrase(id string) bool {
	_, err := b.uc.Passphrase(domain.ProfileID(id))
	return err == nil
}

// ProfileDefaults previews the values that would be auto-applied if the
// corresponding fields in CreateProfile are left empty. The UI uses it to
// hint the user without forcing them to leave the form to look anything up.
type ProfileDefaults struct {
	PrivateKeyPath string `json:"privateKeyPath"`
	RemoteRoot     string `json:"remoteRoot"`
	RemoteTrash    string `json:"remoteTrash"`
	LocalSyncDir   string `json:"localSyncDir"`
}

// SuggestDefaults returns the values that Create would fill given the
// supplied profile name and username. Either argument can be empty.
func (b *ProfileBindings) SuggestDefaults(profileName, username string) ProfileDefaults {
	root := platform.DefaultRemoteRoot(username)
	return ProfileDefaults{
		PrivateKeyPath: platform.DetectPrivateKey(),
		RemoteRoot:     root,
		RemoteTrash:    platform.DefaultRemoteTrash(root),
		LocalSyncDir:   platform.DefaultLocalSyncDir(profileName),
	}
}
