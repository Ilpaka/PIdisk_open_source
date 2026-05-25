package sshclient

import (
	"context"
	"errors"
	"sync"

	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/rs/zerolog"
)

// ClientFactory is what the manager calls to build new SSH clients. We accept
// it as an injection so tests can swap in a fake (in-memory) SSH transport.
type ClientFactory func(domain.Profile) (ports.SSHClient, error)

// Manager keeps at most one live client per profile and tracks the active one.
type Manager struct {
	mu      sync.Mutex
	clients map[domain.ProfileID]ports.SSHClient
	current ports.SSHClient
	factory ClientFactory
	log     zerolog.Logger
}

func NewManager(log zerolog.Logger, factory ClientFactory) *Manager {
	return &Manager{
		clients: map[domain.ProfileID]ports.SSHClient{},
		factory: factory,
		log:     log.With().Str("component", "ssh-pool").Logger(),
	}
}

func (m *Manager) Acquire(ctx context.Context, id domain.ProfileID) (ports.SSHClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.clients[id]; ok {
		if c.IsAlive() {
			m.current = c
			return c, nil
		}
		_ = c.Close()
		delete(m.clients, id)
	}
	c, err := m.factory(domain.Profile{ID: id})
	if err != nil {
		return nil, err
	}
	if err := c.Connect(ctx); err != nil {
		_ = c.Close()
		return nil, err
	}
	m.clients[id] = c
	m.current = c
	return c, nil
}

func (m *Manager) AcquireWithProfile(ctx context.Context, p domain.Profile) (ports.SSHClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.clients[p.ID]; ok {
		if c.IsAlive() {
			m.current = c
			return c, nil
		}
		_ = c.Close()
		delete(m.clients, p.ID)
	}
	c, err := m.factory(p)
	if err != nil {
		return nil, err
	}
	if err := c.Connect(ctx); err != nil {
		_ = c.Close()
		return nil, err
	}
	m.clients[p.ID] = c
	m.current = c
	return c, nil
}

func (m *Manager) Release(id domain.ProfileID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.clients[id]
	if !ok {
		return nil
	}
	if m.current == c {
		m.current = nil
	}
	delete(m.clients, id)
	return c.Close()
}

func (m *Manager) Current() (ports.SSHClient, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil || !m.current.IsAlive() {
		return nil, false
	}
	return m.current, true
}

func (m *Manager) SetCurrent(id domain.ProfileID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.clients[id]
	if !ok {
		return errors.New("client not in pool")
	}
	m.current = c
	return nil
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, c := range m.clients {
		_ = c.Close()
		delete(m.clients, id)
	}
	m.current = nil
	return nil
}

var _ ports.ClientManager = (*Manager)(nil)
