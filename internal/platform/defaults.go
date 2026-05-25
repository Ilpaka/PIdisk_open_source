package platform

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// commonKeyNames are tried in order; first existing file wins.
var commonKeyNames = []string{"id_ed25519", "id_ecdsa", "id_rsa", "id_dsa"}

// DetectPrivateKey scans ~/.ssh for the well-known SSH private key files and
// returns the first match. Returns "" if the home directory cannot be
// resolved or none of the candidates exist.
func DetectPrivateKey() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, name := range commonKeyNames {
		p := filepath.Join(home, ".ssh", name)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

// DefaultRemoteRoot returns a sensible remote starting directory based on the
// SSH username.
func DefaultRemoteRoot(username string) string {
	switch strings.TrimSpace(username) {
	case "":
		return "/"
	case "root":
		return "/root"
	default:
		return "/home/" + username
	}
}

// DefaultRemoteTrash returns the conventional trash path nested under root.
func DefaultRemoteTrash(remoteRoot string) string {
	if remoteRoot == "" {
		remoteRoot = "/"
	}
	return path.Join(remoteRoot, ".pidisk-trash")
}

// DefaultLocalSyncDir returns ~/PIdiskSync/<safeProfileName>. The caller is
// responsible for creating the directory if it does not exist.
func DefaultLocalSyncDir(profileName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	safe := SanitizeFileName(profileName)
	if safe == "" {
		safe = "profile"
	}
	return filepath.Join(home, "PIdiskSync", safe)
}

// SanitizeFileName converts an arbitrary string into something safe to use as
// a path segment: ASCII letters/digits/`-`/`_` stay as-is, spaces become `_`,
// the rest is dropped.
func SanitizeFileName(s string) string {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_':
			out = append(out, byte(r))
		case r == ' ':
			out = append(out, '_')
		}
	}
	return string(out)
}
