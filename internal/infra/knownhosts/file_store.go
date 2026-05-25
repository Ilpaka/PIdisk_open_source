package knownhosts

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"golang.org/x/crypto/ssh"
)

const fileName = "known_hosts"

// FileStore persists trusted host keys in an OpenSSH-compatible file.
// We do not use golang.org/x/crypto/ssh/knownhosts directly because we need
// to expose entries to the UI (for inspection and removal), which the stdlib
// helper does not.
type FileStore struct {
	mu   sync.Mutex
	path string
}

func New(configDir string) (*FileStore, error) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return nil, err
	}
	return &FileStore{path: filepath.Join(configDir, fileName)}, nil
}

func (s *FileStore) Path() string { return s.path }

func (s *FileStore) List() ([]domain.KnownHost, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readLocked()
}

func (s *FileStore) Find(host string, port uint16) (domain.KnownHost, bool, error) {
	entries, err := s.List()
	if err != nil {
		return domain.KnownHost{}, false, err
	}
	target := normaliseHost(host, port)
	for _, e := range entries {
		if normaliseHost(e.Host, e.Port) == target {
			return e, true, nil
		}
	}
	return domain.KnownHost{}, false, nil
}

func (s *FileStore) Add(entry domain.KnownHost) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.readLocked()
	if err != nil {
		return err
	}
	target := normaliseHost(entry.Host, entry.Port)
	filtered := existing[:0]
	for _, e := range existing {
		if normaliseHost(e.Host, e.Port) != target {
			filtered = append(filtered, e)
		}
	}
	if entry.AddedAt.IsZero() {
		entry.AddedAt = time.Now().UTC()
	}
	filtered = append(filtered, entry)
	return s.writeLocked(filtered)
}

func (s *FileStore) Remove(host string, port uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, err := s.readLocked()
	if err != nil {
		return err
	}
	target := normaliseHost(host, port)
	filtered := existing[:0]
	for _, e := range existing {
		if normaliseHost(e.Host, e.Port) != target {
			filtered = append(filtered, e)
		}
	}
	return s.writeLocked(filtered)
}

func (s *FileStore) readLocked() ([]domain.KnownHost, error) {
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []domain.KnownHost
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entry, err := parseLine(line)
		if err != nil {
			continue
		}
		out = append(out, entry)
	}
	return out, scanner.Err()
}

func (s *FileStore) writeLocked(entries []domain.KnownHost) error {
	tmp, err := os.CreateTemp(filepath.Dir(s.path), "known_hosts-*")
	if err != nil {
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	for _, e := range entries {
		fmt.Fprintln(tmp, formatLine(e))
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

// HostPattern in OpenSSH lets the form [host]:port mark non-default ports.
func normaliseHost(host string, port uint16) string {
	if port == 0 || port == 22 {
		return strings.ToLower(host)
	}
	return strings.ToLower(net.JoinHostPort("["+host+"]", fmt.Sprintf("%d", port)))
}

func formatLine(e domain.KnownHost) string {
	host := e.Host
	if e.Port != 0 && e.Port != 22 {
		host = fmt.Sprintf("[%s]:%d", e.Host, e.Port)
	}
	return fmt.Sprintf("%s %s %s # %s", host, e.KeyType, e.PublicKey, e.AddedAt.Format(time.RFC3339))
}

func parseLine(line string) (domain.KnownHost, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return domain.KnownHost{}, errors.New("malformed known_hosts line")
	}
	host, port := parseHostPort(parts[0])
	key, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return domain.KnownHost{}, err
	}
	hash := sha256.Sum256(key)
	added := time.Now().UTC()
	for i, p := range parts {
		if p == "#" && i+1 < len(parts) {
			if t, err := time.Parse(time.RFC3339, parts[i+1]); err == nil {
				added = t
			}
			break
		}
	}
	return domain.KnownHost{
		Host:        host,
		Port:        port,
		KeyType:     parts[1],
		PublicKey:   parts[2],
		Fingerprint: "SHA256:" + strings.TrimRight(base64.StdEncoding.EncodeToString(hash[:]), "="),
		AddedAt:     added,
	}, nil
}

func parseHostPort(raw string) (string, uint16) {
	if strings.HasPrefix(raw, "[") {
		closing := strings.Index(raw, "]")
		if closing == -1 {
			return raw, 22
		}
		host := raw[1:closing]
		rest := raw[closing+1:]
		if strings.HasPrefix(rest, ":") {
			var p int
			_, err := fmt.Sscanf(rest[1:], "%d", &p)
			if err == nil && p > 0 && p < 65536 {
				return host, uint16(p)
			}
		}
		return host, 22
	}
	return raw, 22
}

// FingerprintFromKey is a convenience used by the SSH client when no entry
// exists yet and we need to display the fingerprint to the user.
func FingerprintFromKey(key ssh.PublicKey) string {
	return ssh.FingerprintSHA256(key)
}

var _ ports.KnownHostsStore = (*FileStore)(nil)
