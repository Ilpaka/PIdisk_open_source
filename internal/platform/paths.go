package platform

import (
	"os"
	"path/filepath"
)

const brand = "PIdisk"

// ConfigDir returns the directory for application configuration files.
// macOS:   ~/Library/Application Support/PIdisk
// Windows: %APPDATA%\PIdisk
// Linux:   $XDG_CONFIG_HOME/PIdisk (or ~/.config/PIdisk)
func ConfigDir() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, brand)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// DataDir returns the directory for application data files such as bbolt stores.
// macOS:   ~/Library/Application Support/PIdisk
// Windows: %APPDATA%\PIdisk
// Linux:   $XDG_DATA_HOME/PIdisk (or ~/.local/share/PIdisk)
func DataDir() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		dir := filepath.Join(xdg, brand)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", err
		}
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", brand)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// LogDir returns the directory for application log files.
func LogDir() (string, error) {
	data, err := DataDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(data, "logs")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}
