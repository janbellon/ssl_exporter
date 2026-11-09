package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Listen   ListenConfig `mapstructure:"listen"`
	Client   ClientConfig `mapstructure:"client"`
	LogLevel string       `mapstructure:"log_level"`
}

type ListenConfig struct {
	Host   string `mapstructure:"host"`
	Port   int    `mapstructure:"port"`
	Bearer string `mapstructure:"bearer"`
}

type ClientConfig struct {
	Timeout int `mapstructure:"timeout"`
}

// MaxClientTimeout is the upper bound, in seconds, of client.timeout.
const MaxClientTimeout = 300

func Load(flags *pflag.FlagSet) (*Config, error) {
	v := viper.New()

	v.SetDefault("log_level", "info")
	v.SetDefault("listen.host", "127.0.0.1")
	v.SetDefault("listen.port", 9123)
	v.SetDefault("listen.bearer", "")
	v.SetDefault("client.timeout", 30)

	v.SetEnvPrefix("SSL_EXPORTER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	keys := []string{
		"log_level",
		"listen.host",
		"listen.port",
		"listen.bearer",
		"client.timeout",
	}

	for _, key := range keys {
		if err := v.BindEnv(key); err != nil {
			return nil, err
		}
	}

	if configPath := os.Getenv("SSL_EXPORTER_CONFIG"); configPath != "" {
		v.SetConfigFile(configPath)

		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("could not read config file: %w", err)
		}
	}

	// Configuration key -> command line flag (see cmd/root.go)
	flagMapping := map[string]string{
		"log_level":      "log_level",
		"listen.host":    "listen",
		"listen.port":    "port",
		"listen.bearer":  "bearer",
		"client.timeout": "timeout",
	}

	if flags != nil {
		for vKey, flagName := range flagMapping {
			f := flags.Lookup(flagName)
			if f == nil {
				return nil, fmt.Errorf("unknown flag %q", flagName)
			}
			if err := v.BindPFlag(vKey, f); err != nil {
				return nil, fmt.Errorf("unable to bind flag %q: %w", flagName, err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Listen.Host == "" {
		return fmt.Errorf("listen.host must not be empty")
	}
	if c.Listen.Port < 1 || c.Listen.Port > 65535 {
		return fmt.Errorf("listen.port must be between 1 and 65535")
	}
	if c.Client.Timeout < 1 || c.Client.Timeout > MaxClientTimeout {
		return fmt.Errorf("client.timeout must be between 1 and %d seconds", MaxClientTimeout)
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log_level must be one of debug, info, warn, error")
	}
	return nil
}
