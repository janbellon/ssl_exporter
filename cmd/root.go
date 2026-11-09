/*
Copyright © 2026 Jan Bellon <jan@nobell.fr>
*/
package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"ssl-exporter/internal/config"
	"ssl-exporter/internal/httpserver"
	"ssl-exporter/internal/logging"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const shutdownTimeout = 10 * time.Second

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "ssl-exporter",
	Short:        "SSL Exporter for Prometheus",
	Long:         `A SSL certificate expiration delay exporter for Prometheus`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cmd.Flags())
		if err != nil {
			return err
		}
		logging.SetupLogger(cfg.LogLevel)

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		httpListener := httpserver.NewHTTPListener(cfg)

		listenErr := make(chan error, 1)
		go func() {
			listenErr <- httpListener.Listen()
		}()

		select {
		case err := <-listenErr:
			return err
		case <-ctx.Done():
		}

		slog.Info("Shutdown signal received...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := httpListener.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-listenErr
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("listen", "", "Address to listen on")
	rootCmd.PersistentFlags().Int("port", 0, "Port to listen on")
	rootCmd.PersistentFlags().Int("timeout", 0, "Maximum client timeout in seconds")
	rootCmd.PersistentFlags().String("bearer", "", "Bearer token for clients authentication (visible in the process list: prefer SSL_EXPORTER_LISTEN_BEARER or the config file)")
	rootCmd.PersistentFlags().String("log_level", "", "Logging level (debug, info, warn, error)")
}
