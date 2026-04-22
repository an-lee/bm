package streams

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

func TestResolver_Resolve_skipsBadManifestAddon(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	manifest := `{"id":"ok","name":"OK","resources":[{"name":"stream","types":["movie"],"idPrefixes":["tt"]}],"types":["movie"]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/bad-manifest.json":
			http.NotFound(w, r)
		case strings.HasSuffix(r.URL.Path, "/manifest.json"):
			_, _ = w.Write([]byte(manifest))
		case strings.Contains(r.URL.Path, "/stream/"):
			_, _ = w.Write([]byte(`{"streams":[{"name":"n","url":"https://ok"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.UI.StreamOrder = "rank"
	cfg.TMDB.APIKey = ""
	cfg.Addons = []config.AddonEntry{
		{ID: "bad", ManifestURL: srv.URL + "/bad-manifest.json", Enabled: true, Resources: []string{"stream"}, Types: []string{"movie"}},
		{ID: "ok", ManifestURL: srv.URL + "/manifest.json", Enabled: true, Resources: []string{"stream"}, Types: []string{"movie"}},
	}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	client := &stremio.Client{HTTP: srv.Client()}
	reg := addons.NewRegistry(cfg2, client)
	r := NewResolver(cfg2, reg, client)
	out, err := r.Resolve(context.Background(), "tt1", "movie", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("%#v", out)
	}
}
