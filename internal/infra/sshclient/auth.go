package sshclient

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// LoadSigner reads a private key file and returns a Signer. Plain (unencrypted)
// keys are accepted with passphrase="". If the key is encrypted but no
// passphrase was supplied, ErrPassphraseRequired is returned so the caller can
// prompt the user instead of blowing up.
func LoadSigner(keyPath, passphrase string) (ssh.Signer, error) {
	if keyPath == "" {
		return nil, errors.New("private key path is empty")
	}
	if strings.HasPrefix(keyPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			keyPath = home + keyPath[1:]
		}
	}
	raw, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	if passphrase != "" {
		signer, err := ssh.ParsePrivateKeyWithPassphrase(raw, []byte(passphrase))
		if err != nil {
			return nil, fmt.Errorf("decrypt private key: %w", err)
		}
		return signer, nil
	}
	signer, err := ssh.ParsePrivateKey(raw)
	if err != nil {
		var missing *ssh.PassphraseMissingError
		if errors.As(err, &missing) {
			return nil, ErrPassphraseRequired
		}
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return signer, nil
}

// ErrPassphraseRequired signals that the supplied private key is encrypted and
// the caller has not provided a passphrase.
var ErrPassphraseRequired = errors.New("private key is encrypted, passphrase required")
