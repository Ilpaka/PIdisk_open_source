package wailsapp

import (
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// DialogBindings exposes native OS file/folder pickers to the frontend.
// We deliberately avoid letting the user type raw paths into a text field
// for upload / download targets; that flow is error-prone (typos, escaping,
// nonexistent dirs) and feels foreign on macOS.
type DialogBindings struct {
	app *App
}

func NewDialogBindings(app *App) *DialogBindings {
	return &DialogBindings{app: app}
}

// SelectFile opens a native open-file dialog and returns the chosen path.
// Returns an empty string if the user cancels.
func (b *DialogBindings) SelectFile(title string) (string, error) {
	if title == "" {
		title = "Choose a file"
	}
	return wailsruntime.OpenFileDialog(b.app.Ctx(), wailsruntime.OpenDialogOptions{
		Title: title,
	})
}

// SelectFolder opens a native folder picker.
func (b *DialogBindings) SelectFolder(title string) (string, error) {
	if title == "" {
		title = "Choose a folder"
	}
	return wailsruntime.OpenDirectoryDialog(b.app.Ctx(), wailsruntime.OpenDialogOptions{
		Title: title,
	})
}

// SaveFile opens a native save-as dialog for a regular file with no filter.
func (b *DialogBindings) SaveFile(defaultName, title string) (string, error) {
	if title == "" {
		title = "Save file as"
	}
	return wailsruntime.SaveFileDialog(b.app.Ctx(), wailsruntime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultName,
	})
}

// SaveArchive opens a native save-as dialog with a .zip filter. The default
// filename gets a .zip suffix appended if missing.
func (b *DialogBindings) SaveArchive(defaultName string) (string, error) {
	if defaultName != "" && !endsWithLower(defaultName, ".zip") {
		defaultName += ".zip"
	}
	return wailsruntime.SaveFileDialog(b.app.Ctx(), wailsruntime.SaveDialogOptions{
		Title:           "Save folder as ZIP archive",
		DefaultFilename: defaultName,
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "ZIP archive (*.zip)", Pattern: "*.zip"},
		},
	})
}

func endsWithLower(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	tail := s[len(s)-len(suffix):]
	for i := 0; i < len(suffix); i++ {
		a := tail[i]
		b := suffix[i]
		if a >= 'A' && a <= 'Z' {
			a += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}
