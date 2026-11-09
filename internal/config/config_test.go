package config

import (
	"testing"

	"github.com/spf13/pflag"
)

func newFlags() *pflag.FlagSet {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("listen", "", "")
	fs.Int("port", 0, "")
	fs.Int("timeout", 0, "")
	fs.String("bearer", "", "")
	fs.String("log_level", "", "")
	return fs
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(newFlags())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Listen.Host != "127.0.0.1" || cfg.Listen.Port != 9123 || cfg.Client.Timeout != 30 || cfg.LogLevel != "info" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadFlagsOverrideEnv(t *testing.T) {
	t.Setenv("SSL_EXPORTER_LISTEN_PORT", "9000")
	t.Setenv("SSL_EXPORTER_LISTEN_BEARER", "fromenv")

	fs := newFlags()
	if err := fs.Parse([]string{"--port", "9001", "--listen", "::1", "--timeout", "5", "--log_level", "debug"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(fs)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Listen.Port != 9001 || cfg.Listen.Host != "::1" || cfg.Client.Timeout != 5 || cfg.LogLevel != "debug" || cfg.Listen.Bearer != "fromenv" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

func TestLoadRejectsInvalid(t *testing.T) {
	for _, args := range [][]string{
		{"--port", "70000"},
		{"--timeout", "-1"},
		{"--log_level", "verbose"},
	} {
		fs := newFlags()
		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(fs); err == nil {
			t.Errorf("Load(%v): expected an error", args)
		}
	}
}
