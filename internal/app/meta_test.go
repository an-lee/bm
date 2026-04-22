package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bm/internal/addons"
	"bm/internal/config"
	"bm/internal/stremio"
	"bm/internal/testxdg"
)

func TestApp_Meta(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	manifest := `{"id":"cat","name":"C","resources":["catalog","meta"],"types":["movie"],"catalogs":[{"type":"movie","id":"s","extraSupported":["search"]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "manifest.json"):
			_, _ = w.Write([]byte(manifest))
		case strings.Contains(r.URL.Path, "/meta/"):
			_, _ = w.Write([]byte(`{"meta":{"id":"tt1","type":"movie","name":"M"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	cfg.Addons = []config.AddonEntry{{
		ID: "cat", ManifestURL: srv.URL + "/manifest.json", Enabled: true,
		Resources: []string{"catalog", "meta"}, Types: []string{"movie"},
	}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	client := &stremio.Client{HTTP: srv.Client()}
	a := &App{
		Config: cfg2,
		Client: client,
		Addons: addons.NewRegistry(cfg2, client),
	}
	meta, err := a.Meta(context.Background(), "tt1", "movie")
	if err != nil || meta.Name != "M" {
		t.Fatalf("%v %#v", err, meta)
	}
}
