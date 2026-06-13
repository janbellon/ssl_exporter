package config

import (
	"fmt"
	"strings"

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

func Load() (*Config, error) {
	v := viper.New()
	v.SetDefault("log_level", "info")
	v.SetDefault("listen.host", "127.0.0.1")
	v.SetDefault("listen.port", 9123)
	v.SetDefault("listen.bearer", "")
	v.SetDefault("client.timeout", 30)

	configPath := viper.GetString("SSL_EXPORTER_CONFIG")
	if configPath != "" {
		v.SetConfigFile(configPath)

		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("Could not read config file, %w", err)
		}
	}

	v.SetEnvPrefix("SSL_EXPORTER")

	replacer := strings.NewReplacer(".", "_")

	v.SetEnvKeyReplacer(replacer)

	v.AutomaticEnv()

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode configuration, %w", err)
	}

	return &cfg, nil
}
