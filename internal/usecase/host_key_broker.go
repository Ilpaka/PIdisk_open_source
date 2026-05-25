package usecase

import (
	"sync"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/infra/sshclient"
)

// HostKeyBroker bridges the synchronous SSH host-key callback with the async
// UI prompt. The connect path blocks on Resolve until the user calls Decide
// from the frontend.
type HostKeyBroker struct {
	mu       sync.Mutex
	pending  map[string]chan sshclient.HostKeyDecision
	autoTrust bool
}

func NewHostKeyBroker() *HostKeyBroker {
	return &HostKeyBroker{pending: map[string]chan sshclient.HostKeyDecision{}}
}

func (b *HostKeyBroker) SetAutoTrust(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.autoTrust = v
}

// Prompter returns a function compatible with sshclient.HostKeyPrompter that
// blocks until Decide is called for the matching fingerprint.
func (b *HostKeyBroker) Prompter() sshclient.HostKeyPrompter {
	return func(entry domain.KnownHost) sshclient.HostKeyDecision {
		b.mu.Lock()
		if b.autoTrust {
			b.mu.Unlock()
			return sshclient.HostKeyAccept
		}
		ch, ok := b.pending[entry.Fingerprint]
		if !ok {
			ch = make(chan sshclient.HostKeyDecision, 1)
			b.pending[entry.Fingerprint] = ch
		}
		b.mu.Unlock()
		return <-ch
	}
}

func (b *HostKeyBroker) Decide(fingerprint string, accept bool) {
	b.mu.Lock()
	ch, ok := b.pending[fingerprint]
	if !ok {
		ch = make(chan sshclient.HostKeyDecision, 1)
		b.pending[fingerprint] = ch
	}
	delete(b.pending, fingerprint)
	b.mu.Unlock()
	decision := sshclient.HostKeyReject
	if accept {
		decision = sshclient.HostKeyAccept
	}
	select {
	case ch <- decision:
	default:
	}
}
