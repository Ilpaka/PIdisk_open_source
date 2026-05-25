package eventbus

import (
	"context"
	"sync/atomic"

	"github.com/pidisk/pidisk/internal/ports"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Bus is the Wails-backed implementation of ports.EventEmitter.
// We hold the context captured at app startup so that emitting from
// background goroutines does not race with OnShutdown.
type Bus struct {
	ctx atomic.Pointer[context.Context]
}

func New(ctx context.Context) *Bus {
	b := &Bus{}
	b.SetContext(ctx)
	return b
}

func (b *Bus) SetContext(ctx context.Context) {
	if ctx == nil {
		return
	}
	b.ctx.Store(&ctx)
}

func (b *Bus) Emit(name string, payload any) {
	ptr := b.ctx.Load()
	if ptr == nil {
		return
	}
	ctx := *ptr
	if ctx == nil {
		return
	}
	wailsruntime.EventsEmit(ctx, name, payload)
}

var _ ports.EventEmitter = (*Bus)(nil)
