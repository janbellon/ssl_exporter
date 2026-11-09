package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"ssl-exporter/internal/config"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type HTTPListener struct {
	config  *config.Config
	server  *http.Server
	timeout time.Duration
}

func NewHTTPListener(cfg *config.Config) *HTTPListener {
	l := &HTTPListener{
		config:  cfg,
		timeout: time.Duration(cfg.Client.Timeout) * time.Second,
	}

	r := chi.NewRouter()
	r.Get("/", l.GetRoot)
	r.Get("/health", l.GetHealth)
	r.Get("/metrics", l.GetMetrics)

	l.server = &http.Server{
		Addr:    net.JoinHostPort(cfg.Listen.Host, strconv.Itoa(cfg.Listen.Port)),
		Handler: r,
		// Protect against slow clients holding connections open. There is
		// no WriteTimeout because a scrape legitimately lasts as long as
		// the slowest target; it is bounded by the client timeout and
		// cancelled as soon as the scraper disconnects.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	return l
}

// Listen serves HTTP requests until Shutdown is called. It returns nil on a
// graceful shutdown.
func (l *HTTPListener) Listen() error {
	slog.Info("Listening", "address", l.server.Addr)

	if l.config.Listen.Bearer == "" {
		if ip := net.ParseIP(l.config.Listen.Host); ip == nil || !ip.IsLoopback() {
			slog.Warn("No bearer token configured while listening on a non-loopback address: anyone able to reach this port can make the exporter connect to arbitrary hosts")
		}
	} else if l.config.Listen.Bearer == "changeme" {
		slog.Warn("The bearer token is the sample value 'changeme': change it")
	}

	err := l.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (l *HTTPListener) Shutdown(ctx context.Context) error {
	return l.server.Shutdown(ctx)
}
