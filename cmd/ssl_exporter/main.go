package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"ssl-exporter/internal/config"
	"ssl-exporter/internal/httpserver"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error(fmt.Sprint(err))
	}
	ctx := context.Background()
	httpListener := httpserver.NewHTTPListener(cfg, ctx)

	stopRoutine := make(chan struct{})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run a background task to shutdown all routines
	go func() {
		<-sigChan
		slog.Info("Shutdown signal received...")
		close(stopRoutine)
		httpListener.Shutdown(context.Background())
	}()

	httpListener.Listen()
}
