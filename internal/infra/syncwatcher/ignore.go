package syncwatcher

import (
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

const ignoreFile = ".pidiskignore"

// LoadIgnore reads .pidiskignore from the folder root (if present) and merges
// it with the supplied defaults. Returns nil if nothing was loaded.
func LoadIgnore(root string, defaults []string) *gitignore.GitIgnore {
	lines := append([]string{}, defaults...)
	path := filepath.Join(root, ignoreFile)
	if raw, err := os.ReadFile(path); err == nil {
		for _, line := range splitLines(raw) {
			if line == "" {
				continue
			}
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return nil
	}
	return gitignore.CompileIgnoreLines(lines...)
}

func splitLines(raw []byte) []string {
	var (
		lines []string
		acc   []byte
	)
	for _, b := range raw {
		if b == '\n' || b == '\r' {
			if len(acc) > 0 {
				lines = append(lines, string(acc))
				acc = acc[:0]
			}
			continue
		}
		acc = append(acc, b)
	}
	if len(acc) > 0 {
		lines = append(lines, string(acc))
	}
	return lines
}
