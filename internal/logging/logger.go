package logging

import (
	"log/slog"
	"os"
)

func SetupLogger(level string) {
	slog.Debug("Setting up logging...")

	logLevel := strToLogLevel(level)

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}),
	)

	slog.SetDefault(logger)
}

func strToLogLevel(str string) slog.Level {
	logLevel := slog.LevelInfo
	switch str {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}
	return logLevel
}
