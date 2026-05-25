package ports

import "context"

// MetricsRecorder is intentionally narrow. The Prometheus implementation
// satisfies it; a no-op implementation satisfies it for tests and for users
// who keep metrics disabled in settings.
type MetricsRecorder interface {
	IncBytesUploaded(n int64)
	IncBytesDownloaded(n int64)
	IncSyncRun()
	IncSyncError()
	SetActiveTransfers(n int)
	Serve(ctx context.Context, addr string) error
	Shutdown(ctx context.Context) error
}

type NoopMetrics struct{}

func (NoopMetrics) IncBytesUploaded(int64)                    {}
func (NoopMetrics) IncBytesDownloaded(int64)                  {}
func (NoopMetrics) IncSyncRun()                               {}
func (NoopMetrics) IncSyncError()                             {}
func (NoopMetrics) SetActiveTransfers(int)                    {}
func (NoopMetrics) Serve(context.Context, string) error       { return nil }
func (NoopMetrics) Shutdown(context.Context) error            { return nil }
