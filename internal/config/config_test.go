package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultsAndOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "egressd.yaml")
	content := []byte("listen:\n  socks: 127.0.0.1:1901\n  http: 127.0.0.1:1902\nusers:\n  file: my-users.json\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Listen.SOCKS != "127.0.0.1:1901" {
		t.Fatalf("socks = %s", cfg.Listen.SOCKS)
	}
	if cfg.Listen.HTTP != "127.0.0.1:1902" {
		t.Fatalf("http = %s", cfg.Listen.HTTP)
	}
	if cfg.Users.File != "my-users.json" {
		t.Fatalf("users = %s", cfg.Users.File)
	}
	got := cfg.UsersFilePath(path)
	want := filepath.Join(dir, "my-users.json")
	if got != want {
		t.Fatalf("UsersFilePath = %s want %s", got, want)
	}
}

func TestLoadMissingUsesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	d := Default()
	if cfg.Listen.SOCKS != d.Listen.SOCKS || cfg.Listen.HTTP != d.Listen.HTTP {
		t.Fatalf("unexpected defaults: %+v", cfg.Listen)
	}
}
