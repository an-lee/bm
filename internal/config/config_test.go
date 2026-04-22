package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bm/internal/testxdg"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()
	c := defaultConfig()
	if c.UI.DefaultType != "movie" {
		t.Fatal(c.UI.DefaultType)
	}
	if len(c.Addons) < 2 {
		t.Fatal(len(c.Addons))
	}
	found := false
	for _, a := range c.Addons {
		if a.ID == CinemetaAddonID && a.Enabled {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("cinemeta missing")
	}
}

func TestLoadWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.TMDB.APIKey = "secret"
	cfg.UI.StreamOrder = "title"
	cfg.UI.DefaultType = "series"
	if err := Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.TMDB.APIKey != "secret" || cfg2.UI.StreamOrder != "title" || cfg2.UI.DefaultType != "series" {
		t.Fatalf("%+v", cfg2.UI)
	}
}

func TestLoad_setsDefaultTypeWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	// Minimal TOML without ui.default_type
	if err := os.WriteFile(p, []byte(`[ui]
stream_order = "rank"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UI.DefaultType != "movie" {
		t.Fatalf("got %q", cfg.UI.DefaultType)
	}
}

func TestSetKeyGetKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	if err := SetKey("tmdb.api_key", "k1"); err != nil {
		t.Fatal(err)
	}
	v, err := GetKey("tmdb.api_key")
	if err != nil || v != "k1" {
		t.Fatalf("%v %q", err, v)
	}
	if err := SetKey("ui.default_type", "series"); err != nil {
		t.Fatal(err)
	}
	v2, err := GetKey("ui.default_type")
	if err != nil || v2 != "series" {
		t.Fatalf("%v %q", err, v2)
	}
}

func TestSetKey_unknown(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	err := SetKey("nope", "x")
	if err == nil || !strings.Contains(err.Error(), "unknown config key") {
		t.Fatalf("err %v", err)
	}
}

func TestGetKey_unknown(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	_, err := GetKey("nope")
	if err == nil || !strings.Contains(err.Error(), "unknown config key") {
		t.Fatalf("err %v", err)
	}
}

func TestPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	p, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(p, filepath.Join("bm", "config.toml")) {
		t.Fatalf("path %q", p)
	}
}
