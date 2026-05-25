package sshclient

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"golang.org/x/crypto/ssh"
)

// HostKeyDecision is what the UI returns when prompted: accept the new key
// (TOFU) or reject the connection.
type HostKeyDecision int

const (
	HostKeyAccept HostKeyDecision = iota + 1
	HostKeyReject
)

// HostKeyPrompter is called when an unknown host key is encountered. The
// implementation is expected to surface the prompt to the user and block until
// they answer.
type HostKeyPrompter func(entry domain.KnownHost) HostKeyDecision

type tofuCallback struct {
	store    ports.KnownHostsStore
	prompter HostKeyPrompter
	bus      ports.EventEmitter
}

func newTOFUCallback(store ports.KnownHostsStore, bus ports.EventEmitter, prompter HostKeyPrompter) *tofuCallback {
	return &tofuCallback{store: store, bus: bus, prompter: prompter}
}

func (t *tofuCallback) callback(hostname string, remote net.Addr, key ssh.PublicKey) error {
	host, port := splitHostPort(hostname, remote)
	entry := domain.KnownHost{
		Host:        host,
		Port:        port,
		KeyType:     key.Type(),
		PublicKey:   base64.StdEncoding.EncodeToString(key.Marshal()),
		Fingerprint: ssh.FingerprintSHA256(key),
		AddedAt:     time.Now().UTC(),
	}
	stored, found, err := t.store.Find(host, port)
	if err != nil {
		return err
	}
	if found {
		if stored.Fingerprint == entry.Fingerprint {
			return nil
		}
		if t.bus != nil {
			t.bus.Emit("hostkey:mismatch", entry)
		}
		return fmt.Errorf("%w: stored=%s incoming=%s", domain.ErrHostKeyMismatch, stored.Fingerprint, entry.Fingerprint)
	}
	if t.prompter == nil {
		return domain.ErrHostKeyPending
	}
	if t.bus != nil {
		t.bus.Emit("hostkey:prompt", entry)
	}
	decision := t.prompter(entry)
	if decision != HostKeyAccept {
		return domain.ErrHostKeyRejected
	}
	return t.store.Add(entry)
}

func splitHostPort(hostname string, remote net.Addr) (string, uint16) {
	host, portStr, err := net.SplitHostPort(hostname)
	if err == nil {
		if p, err := strconv.Atoi(portStr); err == nil {
			return host, uint16(p)
		}
	}
	if remote != nil {
		if h, ps, err := net.SplitHostPort(remote.String()); err == nil {
			if p, err := strconv.Atoi(ps); err == nil {
				return h, uint16(p)
			}
		}
	}
	if strings.Contains(hostname, ":") {
		idx := strings.LastIndex(hostname, ":")
		host = hostname[:idx]
	} else {
		host = hostname
	}
	return host, 22
}

func wrapHostKeyError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrHostKeyRejected) ||
		errors.Is(err, domain.ErrHostKeyMismatch) ||
		errors.Is(err, domain.ErrHostKeyPending) {
		return err
	}
	return err
}
