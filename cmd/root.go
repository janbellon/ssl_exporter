/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"ssl-exporter/internal/config"
	"ssl-exporter/internal/httpserver"
	"syscall"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ssl-exporter",
	Short: "SSL Exporter for Prometheus",
	Long:  `A SSL certificate expiration delay exporter for Prometheus`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cmd.Flags())
		if err != nil {
			return err
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
		return nil
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
	rootCmd.PersistentFlags().String("bearer", "", "Bearer token for clients authentication")
	rootCmd.PersistentFlags().String("log_level", "", "Logging level (info, warn, error)")
}
