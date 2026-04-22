package search

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

func TestService_CinemetaFeatured(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	manifest := `{"id":"` + config.CinemetaAddonID + `","name":"C","resources":["catalog"],"types":["movie"],"catalogs":[{"type":"movie","id":"imdbRating","name":"Featured"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "manifest.json") {
			_, _ = w.Write([]byte(manifest))
			return
		}
		_, _ = w.Write([]byte(`{"metas":[{"id":"tt2","imdb_id":"tt2","type":"movie","name":"F"}]}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	cfg.TMDB.APIKey = ""
	cfg.Addons = []config.AddonEntry{{
		ID: config.CinemetaAddonID, ManifestURL: srv.URL + "/manifest.json",
		Enabled: true, Resources: []string{"catalog"}, Types: []string{"movie"},
	}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, _ := config.Load()
	client := &stremio.Client{HTTP: srv.Client()}
	s := NewService(cfg2, addons.NewRegistry(cfg2, client), client)
	res, err := s.CinemetaFeatured(context.Background(), "movie", 0)
	if err != nil || len(res) != 1 {
		t.Fatalf("%v %#v", err, res)
	}
}

func TestService_searchCatalog_skipsBadBaseURL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	manifest := `{"id":"good","name":"G","resources":["catalog"],"types":["movie"],"catalogs":[{"type":"movie","id":"s","extraSupported":["search"]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "manifest.json"):
			_, _ = w.Write([]byte(manifest))
		case strings.Contains(r.URL.Path, "/search="):
			_, _ = w.Write([]byte(`{"metas":[{"id":"tt1","imdb_id":"tt1","type":"movie","name":"Hit"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	cfg.TMDB.APIKey = ""
	cfg.Addons = []config.AddonEntry{
		{ID: "bad", Name: "B", ManifestURL: "http:// [invalid", Enabled: true, Resources: []string{"catalog"}, Types: []string{"movie"}},
		{ID: "good", Name: "G", ManifestURL: srv.URL + "/manifest.json", Enabled: true, Resources: []string{"catalog"}, Types: []string{"movie"}},
	}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, _ := config.Load()
	client := &stremio.Client{HTTP: srv.Client()}
	s := NewService(cfg2, addons.NewRegistry(cfg2, client), client)
	res, err := s.Search(context.Background(), "q", "movie", 0)
	if err != nil || len(res) != 1 {
		t.Fatalf("%v %#v", err, res)
	}
}
