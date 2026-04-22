package mcp

import (
	"testing"

	"bm/internal/config"
)

func TestBuildServer(t *testing.T) {
	t.Parallel()
	s := buildServer()
	if s == nil {
		t.Fatal()
	}
}

func TestRedactedConfigSnapshot(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "secret"
	cfg.UI.DefaultType = "series"
	cfg.Addons = make([]config.AddonEntry, 3)
	m := redactedConfigSnapshot(cfg)
	tmdb := m["tmdb"].(map[string]any)
	if tmdb["api_key_configured"] != true {
		t.Fatal(m)
	}
	if m["addons_count"] != 3 {
		t.Fatal(m)
	}
}

func TestMustJSON_ok(t *testing.T) {
	t.Parallel()
	res, err := mustJSON(map[string]string{"a": "b"})
	if err != nil || res == nil {
		t.Fatalf("%v %#v", err, res)
	}
}

func TestMustJSON_error(t *testing.T) {
	t.Parallel()
	res, err := mustJSON(make(chan int))
	if err != nil || res == nil || !res.IsError {
		t.Fatalf("err %v res %#v", err, res)
	}
}
