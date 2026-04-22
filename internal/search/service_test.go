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

func TestTitleResultsFromMetas(t *testing.T) {
	t.Parallel()
	out := titleResultsFromMetas([]stremio.Meta{
		{ID: "id1", IMDBID: "", Name: "", Type: "movie", Description: "d", ReleaseInfo: "2020"},
		{ID: "tt2", IMDBID: "tt2", Name: "N", Type: "series", Year: "2021"},
	})
	if len(out) != 2 || out[0].Title != "id1" || out[0].IMDBID != "id1" {
		t.Fatalf("%+v", out[0])
	}
	if out[1].Title != "N" || out[1].Year != "2021" {
		t.Fatalf("%+v", out[1])
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Parallel()
	if firstNonEmpty(" a ", "") != " a " {
		t.Fatal()
	}
	if firstNonEmpty("", "b") != "b" {
		t.Fatal()
	}
}

func TestYearString(t *testing.T) {
	t.Parallel()
	if yearString("2020-01-01") != "2020" {
		t.Fatal(yearString("2020-01-01"))
	}
	if yearString("ab") != "" {
		t.Fatal()
	}
}

func TestStrconvYear(t *testing.T) {
	t.Parallel()
	y, err := strconvYear("2020")
	if err != nil || y != 2020 {
		t.Fatalf("%v %d", err, y)
	}
	if _, err := strconvYear(""); err == nil {
		t.Fatal()
	}
}

func TestManifestHasCatalog(t *testing.T) {
	t.Parallel()
	m := &stremio.Manifest{
		Catalogs: []stremio.CatalogDef{{Type: "movie", ID: "top"}},
	}
	if !manifestHasCatalog(m, "movie", "top") {
		t.Fatal()
	}
	if manifestHasCatalog(m, "series", "top") {
		t.Fatal()
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	if truncate("hi", 10) != "hi" {
		t.Fatal()
	}
	s := strings.Repeat("x", 300)
	if len(truncate(s, 20)) <= 20 {
		t.Fatal()
	}
}

func TestService_Search_emptyQuery(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	_, err := s.Search(context.Background(), "  ", "movie", 0)
	if err == nil {
		t.Fatal()
	}
}

func TestService_findEnabledCinemeta_error(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.Addons = []config.AddonEntry{{ID: "other", Enabled: true}}
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	_, err := s.CinemetaPopular(context.Background(), "movie", 0)
	if err == nil || !strings.Contains(err.Error(), "Cinemeta") {
		t.Fatalf("err %v", err)
	}
}

func TestService_cinemetaCatalog_success(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	t.Setenv("TMDB_API_KEY", "")
	t.Setenv("BM_TMDB_API_KEY", "")
	manifest := `{"id":"` + config.CinemetaAddonID + `","name":"C","resources":["catalog"],"types":["movie"],"catalogs":[{"type":"movie","id":"top","name":"Popular"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "manifest.json") {
			_, _ = w.Write([]byte(manifest))
			return
		}
		_, _ = w.Write([]byte(`{"metas":[{"id":"tt1","imdb_id":"tt1","type":"movie","name":"A"}]}`))
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	cfg.TMDB.APIKey = ""
	cfg.Addons = []config.AddonEntry{{
		ID:          config.CinemetaAddonID,
		Name:        "Cinemeta",
		ManifestURL: srv.URL + "/manifest.json",
		Enabled:     true,
		Resources:   []string{"catalog", "meta"},
		Types:       []string{"movie"},
		IDPrefixes:  []string{"tt"},
	}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg2.TMDB.APIKey = ""
	client := &stremio.Client{HTTP: srv.Client()}
	s := NewService(cfg2, addons.NewRegistry(cfg2, client), client)
	res, err := s.CinemetaPopular(context.Background(), "movie", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].IMDBID != "tt1" {
		t.Fatalf("%#v", res)
	}
}

func TestService_searchCatalog_viaRegistry(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	t.Setenv("TMDB_API_KEY", "")
	t.Setenv("BM_TMDB_API_KEY", "")
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	cfg.TMDB.APIKey = ""
	cfg.Addons = nil
	manifest := `{"id":"cat","name":"C","resources":["catalog"],"types":["movie"],"catalogs":[{"type":"movie","id":"s","extraSupported":["search"]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "manifest.json"):
			_, _ = w.Write([]byte(manifest))
		case strings.Contains(r.URL.Path, "/search="):
			_, _ = w.Write([]byte(`{"metas":[{"id":"tt9","imdb_id":"tt9","type":"movie","name":"Hit"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	cfg.Addons = []config.AddonEntry{{ID: "cat", Enabled: true, ManifestURL: srv.URL + "/manifest.json"}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, _ := config.Load()
	client := &stremio.Client{HTTP: srv.Client()}
	s := NewService(cfg2, addons.NewRegistry(cfg2, client), client)
	res, err := s.Search(context.Background(), "q", "movie", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].Title != "Hit" {
		t.Fatalf("%#v err %v", res, err)
	}
}

func TestService_ResolveIMDBID_noKey(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	_, err := s.ResolveIMDBID(context.Background(), "x", 0)
	if err == nil || !strings.Contains(err.Error(), "TMDB") {
		t.Fatalf("%v", err)
	}
}
