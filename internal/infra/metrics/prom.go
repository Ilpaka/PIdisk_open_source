package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/pidisk/pidisk/internal/ports"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Recorder struct {
	registry *prometheus.Registry

	bytesUploaded   prometheus.Counter
	bytesDownloaded prometheus.Counter
	syncRuns        prometheus.Counter
	syncErrors      prometheus.Counter
	activeTransfers prometheus.Gauge

	mu  sync.Mutex
	srv *http.Server
}

func New() *Recorder {
	reg := prometheus.NewRegistry()
	auto := promauto.With(reg)
	return &Recorder{
		registry:        reg,
		bytesUploaded:   auto.NewCounter(prometheus.CounterOpts{Name: "pidisk_bytes_uploaded_total"}),
		bytesDownloaded: auto.NewCounter(prometheus.CounterOpts{Name: "pidisk_bytes_downloaded_total"}),
		syncRuns:        auto.NewCounter(prometheus.CounterOpts{Name: "pidisk_sync_runs_total"}),
		syncErrors:      auto.NewCounter(prometheus.CounterOpts{Name: "pidisk_sync_errors_total"}),
		activeTransfers: auto.NewGauge(prometheus.GaugeOpts{Name: "pidisk_active_transfers"}),
	}
}

func (r *Recorder) IncBytesUploaded(n int64) {
	if n <= 0 {
		return
	}
	r.bytesUploaded.Add(float64(n))
}

func (r *Recorder) IncBytesDownloaded(n int64) {
	if n <= 0 {
		return
	}
	r.bytesDownloaded.Add(float64(n))
}

func (r *Recorder) IncSyncRun()         { r.syncRuns.Inc() }
func (r *Recorder) IncSyncError()       { r.syncErrors.Inc() }
func (r *Recorder) SetActiveTransfers(n int) { r.activeTransfers.Set(float64(n)) }

func (r *Recorder) Serve(ctx context.Context, addr string) error {
	r.mu.Lock()
	if r.srv != nil {
		r.mu.Unlock()
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{}))
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	r.srv = srv
	r.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()
	select {
	case <-ctx.Done():
		return r.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (r *Recorder) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	srv := r.srv
	r.srv = nil
	r.mu.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

var _ ports.MetricsRecorder = (*Recorder)(nil)
