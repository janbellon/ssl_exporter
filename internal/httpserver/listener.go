package httpserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"ssl-exporter/internal/config"
	"time"

	"github.com/go-chi/chi/v5"
)

type HTTPListener struct {
	config *config.Config
	server *http.Server
	client *http.Client
	ctx    context.Context
}

func NewHTTPListener(cfg *config.Config, ctx context.Context) *HTTPListener {
	return &HTTPListener{
		config: cfg,
		ctx:    ctx,
	}
}

func (l *HTTPListener) Listen() error {
	r := chi.NewRouter()
	r.Get("/", l.GetRoot)
	r.Get("/health", l.GetHealth)
	r.Get("/metrics", l.GetMetrics)

	listen := fmt.Sprintf("%s:%d", l.config.Listen.Host, l.config.Listen.Port)

	l.server = &http.Server{
		Addr:    listen,
		Handler: r,
	}

	l.client = &http.Client{
		Timeout: time.Duration(l.config.Client.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	slog.Info(fmt.Sprintf("Listening on %s ...", listen))

	err := l.server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (l *HTTPListener) Shutdown(ctx context.Context) error {
	if l.server == nil {
		return nil
	}

	return l.server.Shutdown(ctx)
}
