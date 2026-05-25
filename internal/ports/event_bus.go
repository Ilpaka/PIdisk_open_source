package ports

// EventEmitter is the thin abstraction over the Wails runtime event channel.
// Usecases depend on this rather than the runtime package directly so they
// can be exercised in unit tests with a recording fake.
type EventEmitter interface {
	Emit(name string, payload any)
}

type EventEmitterFunc func(name string, payload any)

func (f EventEmitterFunc) Emit(name string, payload any) { f(name, payload) }
