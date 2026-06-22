package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveConfigPath(t *testing.T) {
	t.Run("falls back to flag value", func(t *testing.T) {
		t.Setenv("GONDOLA_CONFIG", "")
		if got := resolveConfigPath("flag.yaml"); got != "flag.yaml" {
			t.Errorf("expected flag.yaml, got %q", got)
		}
	})

	t.Run("env overrides flag", func(t *testing.T) {
		t.Setenv("GONDOLA_CONFIG", "env.yaml")
		if got := resolveConfigPath("flag.yaml"); got != "env.yaml" {
			t.Errorf("expected env.yaml, got %q", got)
		}
	})
}

func TestOpenConfig(t *testing.T) {
	if _, err := openConfig(""); err == nil {
		t.Error("expected error for empty path")
	}

	if _, err := openConfig("invalid_file"); err == nil {
		t.Error("expected error for non-existent file")
	}

	f, err := openConfig("../testdata/config/config.yml")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if f != nil {
		_ = f.Close()
	}
}

func TestVersion(t *testing.T) {
	if version() == "" {
		t.Error("expected non-empty version")
	}
}

func TestRunVersion(t *testing.T) {
	if code := run([]string{"-version"}); code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunMissingConfig(t *testing.T) {
	t.Setenv("GONDOLA_CONFIG", "")
	if code := run([]string{"-config", "does_not_exist.yaml"}); code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestRunInvalidConfig(t *testing.T) {
	t.Setenv("GONDOLA_CONFIG", "")
	path := filepath.Join(t.TempDir(), "bad.yaml")
	// Valid YAML but an invalid upstream target -> NewGondola fails.
	content := "proxy:\n  port: \"8080\"\nupstreams:\n  - host_name: x\n    target: \"://\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"-config", path}); code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}
