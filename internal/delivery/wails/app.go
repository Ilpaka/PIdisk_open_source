package wailsapp

import (
	"context"
	"sync"

	"github.com/pidisk/pidisk/internal/infra/eventbus"
	"github.com/pidisk/pidisk/internal/version"
	"github.com/rs/zerolog"
)

// App is the root Wails-bound struct. Sub-bindings (Files, Transfer, Sync, etc.)
// are added under the same Bind() in main.go.
type App struct {
	mu  sync.RWMutex
	ctx context.Context
	log zerolog.Logger
	bus *eventbus.Bus
}

func NewApp(log zerolog.Logger, bus *eventbus.Bus) *App {
	return &App{log: log, bus: bus, ctx: context.Background()}
}

func (a *App) OnStartup(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()
	a.bus.SetContext(ctx)
	a.log.Info().Str("version", version.Full()).Msg("PIdisk starting")
}

func (a *App) OnShutdown(_ context.Context) {
	a.log.Info().Msg("PIdisk shutting down")
}

// Ctx returns the runtime context captured at startup. Bindings use this
// instead of accepting context.Context as an argument because the Wails
// generator would otherwise surface it as a frontend parameter.
func (a *App) Ctx() context.Context {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ctx
}

// Logger exposes the shared zerolog instance to sibling bindings.
func (a *App) Logger() zerolog.Logger { return a.log }

// Bus exposes the event emitter to sibling bindings.
func (a *App) Bus() *eventbus.Bus { return a.bus }

// Version is exposed to the frontend so the About page can render it.
func (a *App) Version() string {
	return version.Full()
}

// Ping is a sanity check used during scaffolding.
func (a *App) Ping() string {
	return "pong"
}
