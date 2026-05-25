package sshclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pidisk/pidisk/internal/domain"
	"github.com/pidisk/pidisk/internal/infra/sftpfs"
	"github.com/pidisk/pidisk/internal/ports"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"
)

const (
	dialTimeout       = 15 * time.Second
	keepaliveInterval = 30 * time.Second
	keepaliveTimeout  = 10 * time.Second
)

// Client wraps a single live SSH connection plus its SFTP subsystem.
// All operations are guarded by mu so background reconnect cannot race with
// in-flight RPCs invoked from Wails bindings.
type Client struct {
	profile domain.Profile
	signer  ssh.Signer
	store   ports.KnownHostsStore
	bus     ports.EventEmitter
	log     zerolog.Logger
	prompt  HostKeyPrompter

	mu       sync.RWMutex
	ssh      *ssh.Client
	sftpFS   *sftpfs.FS
	closeKA  context.CancelFunc
	alive    atomic.Bool
}

type Config struct {
	Profile  domain.Profile
	Signer   ssh.Signer
	Store    ports.KnownHostsStore
	Bus      ports.EventEmitter
	Logger   zerolog.Logger
	Prompter HostKeyPrompter
}

func New(cfg Config) *Client {
	return &Client{
		profile: cfg.Profile,
		signer:  cfg.Signer,
		store:   cfg.Store,
		bus:     cfg.Bus,
		log:     cfg.Logger.With().Str("component", "ssh").Str("profile", string(cfg.Profile.ID)).Logger(),
		prompt:  cfg.Prompter,
	}
}

func (c *Client) Profile() domain.Profile { return c.profile }

func (c *Client) IsAlive() bool { return c.alive.Load() }

func (c *Client) RemoteFS() ports.RemoteFS {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.sftpFS == nil {
		return nil
	}
	return c.sftpFS
}

func (c *Client) Connect(ctx context.Context) error {
	cfg := &ssh.ClientConfig{
		User:            c.profile.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(c.signer)},
		HostKeyCallback: newTOFUCallback(c.store, c.bus, c.prompt).callback,
		Timeout:         dialTimeout,
	}
	addr := net.JoinHostPort(c.profile.Host, strconv.Itoa(int(c.profile.Port)))
	dialer := &net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		c.emitDisconnect(err)
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, cfg)
	if err != nil {
		_ = conn.Close()
		c.emitDisconnect(wrapHostKeyError(err))
		return wrapHostKeyError(err)
	}
	sshClient := ssh.NewClient(sshConn, chans, reqs)
	fs, err := sftpfs.New(sshClient)
	if err != nil {
		_ = sshClient.Close()
		c.emitDisconnect(err)
		return err
	}

	c.mu.Lock()
	c.ssh = sshClient
	c.sftpFS = fs
	c.alive.Store(true)
	c.mu.Unlock()

	kaCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.closeKA = cancel
	c.mu.Unlock()
	go c.runKeepalive(kaCtx)

	c.log.Info().Str("addr", addr).Msg("ssh connected")
	c.emitConnect()
	return nil
}

func (c *Client) Reconnect(ctx context.Context) error {
	c.Close()
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 0
	bo.InitialInterval = time.Second
	bo.MaxInterval = 30 * time.Second
	return backoff.RetryNotify(func() error {
		return c.Connect(ctx)
	}, backoff.WithContext(bo, ctx), func(err error, d time.Duration) {
		c.log.Warn().Err(err).Dur("retry_in", d).Msg("reconnect attempt failed")
	})
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closeKA != nil {
		c.closeKA()
		c.closeKA = nil
	}
	if c.sftpFS != nil {
		_ = c.sftpFS.Close()
		c.sftpFS = nil
	}
	if c.ssh != nil {
		_ = c.ssh.Close()
		c.ssh = nil
	}
	c.alive.Store(false)
	return nil
}

func (c *Client) NewSession(ctx context.Context) (ports.RemoteSession, error) {
	c.mu.RLock()
	sshClient := c.ssh
	c.mu.RUnlock()
	if sshClient == nil {
		return nil, domain.ErrNotConnected
	}
	sess, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	return &sessionAdapter{sess: sess}, nil
}

type sessionAdapter struct{ sess *ssh.Session }

func (s *sessionAdapter) Output(cmd string) ([]byte, error) { return s.sess.Output(cmd) }
func (s *sessionAdapter) Close() error                      { return s.sess.Close() }

func (c *Client) runKeepalive(ctx context.Context) {
	t := time.NewTicker(keepaliveInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := c.ping(ctx); err != nil {
				c.alive.Store(false)
				c.emitDisconnect(err)
				if err := c.Reconnect(ctx); err != nil {
					c.log.Error().Err(err).Msg("reconnect aborted")
					return
				}
			}
		}
	}
}

func (c *Client) ping(ctx context.Context) error {
	c.mu.RLock()
	sshClient := c.ssh
	c.mu.RUnlock()
	if sshClient == nil {
		return domain.ErrNotConnected
	}
	done := make(chan error, 1)
	go func() {
		_, _, err := sshClient.SendRequest("keepalive@openssh.com", true, nil)
		done <- err
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(keepaliveTimeout):
		return errors.New("keepalive timed out")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) emitConnect() {
	if c.bus == nil {
		return
	}
	c.bus.Emit("connection:state", domain.ConnectionState{
		ProfileID: c.profile.ID,
		Connected: true,
		LastPing:  time.Now().UTC(),
	})
}

func (c *Client) emitDisconnect(err error) {
	if c.bus == nil {
		return
	}
	state := domain.ConnectionState{
		ProfileID: c.profile.ID,
		Connected: false,
		LastPing:  time.Now().UTC(),
	}
	if err != nil {
		state.LastError = err.Error()
	}
	c.bus.Emit("connection:state", state)
}

var _ ports.SSHClient = (*Client)(nil)
