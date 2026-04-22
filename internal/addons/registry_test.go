package addons

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"bm/internal/config"
	"bm/internal/stremio"
	"bm/internal/testxdg"
)

func TestPickSearchCatalog(t *testing.T) {
	t.Parallel()
	m := &stremio.Manifest{
		Catalogs: []stremio.CatalogDef{
			{Type: "movie", ID: "other", ExtraSupported: []string{"genre"}},
			{Type: "movie", ID: "search", ExtraSupported: []string{"search"}},
		},
	}
	id, ok := PickSearchCatalog(m, "movie")
	if !ok || id != "search" {
		t.Fatalf("got %q %v", id, ok)
	}
	_, ok = PickSearchCatalog(m, "series")
	if ok {
		t.Fatal("expected false")
	}
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	cfg.Addons = []config.AddonEntry{
		{ID: "a", Enabled: true},
		{ID: "b", Enabled: false},
	}
	r := NewRegistry(cfg, stremio.NewClient())
	if len(r.List(false)) != 1 {
		t.Fatal(r.List(false))
	}
	if len(r.List(true)) != 2 {
		t.Fatal(r.List(true))
	}
}

func TestRegistry_Remove(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Addons = []config.AddonEntry{{ID: "x", Enabled: true, ManifestURL: "https://example/m"}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, _ := config.Load()
	r := NewRegistry(cfg2, stremio.NewClient())
	if err := r.Remove("x"); err != nil {
		t.Fatal(err)
	}
	if err := r.Remove("missing"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRegistry_Install(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()

	manifest := `{"id":"com.test","name":"TestAddon","version":"1","resources":["stream"],"types":["movie"],"idPrefixes":["tt"],"catalogs":[]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(srv.Close)

	client := &stremio.Client{HTTP: srv.Client()}
	r := NewRegistry(cfg, client)
	entry, err := r.Install(context.Background(), srv.URL+"/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if entry.ID != "com.test" {
		t.Fatalf("%+v", entry)
	}
}

func TestRegistry_StreamAddons(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	cfg.Addons = []config.AddonEntry{{
		ID: "x", Enabled: true, ManifestURL: "MANIFEST_URL",
	}}

	manifest := `{"id":"x","name":"X","resources":[{"name":"stream","types":["movie"],"idPrefixes":["tt"]}],"types":["movie"]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(srv.Close)
	cfg.Addons[0].ManifestURL = srv.URL + "/manifest.json"

	client := &stremio.Client{HTTP: srv.Client()}
	r := NewRegistry(cfg, client)
	list, err := r.StreamAddons(context.Background(), "movie", "tt123")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("%#v", list)
	}
}

func TestContains(t *testing.T) {
	t.Parallel()
	if !contains([]string{"a", "b"}, "b") {
		t.Fatal()
	}
	if contains([]string{"a", "b"}, "z") {
		t.Fatal()
	}
}

func TestRegistry_GetManifest(t *testing.T) {
	manifest := `{"id":"z","name":"Z","resources":["stream"],"types":["movie"]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	r := NewRegistry(cfg, &stremio.Client{HTTP: srv.Client()})
	m, err := r.GetManifest(context.Background(), config.AddonEntry{ManifestURL: srv.URL + "/m.json"})
	if err != nil || m == nil || m.ID != "z" {
		t.Fatalf("err %v m %#v", err, m)
	}
}

func TestRegistry_RefreshMetadata(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	manifest := `{"id":"a","name":"Fresh","resources":["catalog","stream"],"types":["movie"],"idPrefixes":["tt"],"behaviorHints":{"openUrlTemplate":"https://cfg.example/"},"catalogs":[]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(srv.Close)
	cfg.Addons = []config.AddonEntry{{
		ID: "a", Enabled: true, ManifestURL: srv.URL + "/manifest.json", Name: "Old",
	}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	cfg2, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	r := NewRegistry(cfg2, &stremio.Client{HTTP: srv.Client()})
	if err := r.RefreshMetadata(context.Background()); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Addons) != 1 || loaded.Addons[0].Name != "Fresh" {
		t.Fatalf("%+v", loaded.Addons)
	}
	if loaded.Addons[0].ConfigurationURL != "https://cfg.example/" {
		t.Fatalf("cfg url %q", loaded.Addons[0].ConfigurationURL)
	}
}

func TestRegistry_StreamAddons_imdbWithColon(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	cfg, _ := config.Load()
	cfg.Addons = []config.AddonEntry{{
		ID: "x", Enabled: true, ManifestURL: "MANIFEST_URL",
	}}
	manifest := `{"id":"x","name":"X","resources":[{"name":"stream","types":["movie"],"idPrefixes":["tt"]}],"types":["movie"]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(srv.Close)
	cfg.Addons[0].ManifestURL = srv.URL + "/manifest.json"
	r := NewRegistry(cfg, &stremio.Client{HTTP: srv.Client()})
	list, err := r.StreamAddons(context.Background(), "movie", "tt999:1:2")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("%#v", list)
	}
}

func TestRegistry_CatalogAddons(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	cfg, _ := config.Load()
	cfg.Addons = []config.AddonEntry{{
		ID: "cat", Enabled: true, ManifestURL: "MANIFEST_URL",
	}}

	manifest := `{"id":"cat","name":"C","resources":["catalog"],"types":["movie"],"catalogs":[{"type":"movie","id":"top","extraSupported":["search"]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(srv.Close)
	cfg.Addons[0].ManifestURL = srv.URL + "/manifest.json"

	client := &stremio.Client{HTTP: srv.Client()}
	r := NewRegistry(cfg, client)
	list, err := r.CatalogAddons(context.Background(), "movie")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("%#v", list)
	}
}
