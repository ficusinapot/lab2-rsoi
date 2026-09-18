package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestCommandYAML(t *testing.T) {
	t.Parallel()
	const configFlag = "--config"
	const migrateCommand = "migrate"
	type settings struct {
		Timeout time.Duration `mapstructure:"timeout"`
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("timeout: 3s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	cmd := NewCommand("test", func(_ *cobra.Command, cfg settings) error {
		called = true
		if cfg.Timeout != 3*time.Second {
			t.Fatalf("timeout: %s", cfg.Timeout)
		}
		return nil
	})
	cmd.SetArgs([]string{configFlag, path})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("command callback was not called")
	}
	for _, args := range [][]string{{configFlag, path, migrateCommand}, {migrateCommand, configFlag, path}} {
		invoked := false
		cmd := NewCommand("test", func(_ *cobra.Command, _ settings) error { t.Fatal("wrong command invoked"); return nil },
			Action[settings]{Name: migrateCommand, Run: func(_ *cobra.Command, cfg settings) error {
				invoked = true
				if cfg.Timeout != 3*time.Second {
					t.Fatal("subcommand ignored YAML")
				}
				return nil
			}})
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		if !invoked {
			t.Fatal("subcommand was not invoked")
		}
	}
	if err := os.WriteFile(path, []byte("unknown: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load[settings](path); err == nil {
		t.Fatal("unknown configuration key accepted")
	}
}
