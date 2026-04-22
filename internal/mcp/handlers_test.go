package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bm/internal/config"
	"bm/internal/testxdg"

	"github.com/mark3labs/mcp-go/mcp"
)

func setupTempConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
}

func reqTool(name string, args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: name, Arguments: args},
	}
}

func TestToolListAddons(t *testing.T) {
	setupTempConfig(t)
	res, err := toolListAddons(context.Background(), reqTool("list_addons", nil))
	if err != nil || res == nil || res.IsError {
		t.Fatalf("err %v res %#v", err, res)
	}
}

func TestToolSearchTitle(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	manifest := `{"id":"cat","name":"C","resources":["catalog"],"types":["movie"],"catalogs":[{"type":"movie","id":"s","extraSupported":["search"]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "manifest.json"):
			_, _ = w.Write([]byte(manifest))
		case strings.Contains(r.URL.Path, "/search="):
			_, _ = w.Write([]byte(`{"metas":[{"id":"tt1","imdb_id":"tt1","type":"movie","name":"X"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	cfg.TMDB.APIKey = ""
	cfg.Addons = []config.AddonEntry{{
		ID: "cat", ManifestURL: srv.URL + "/manifest.json", Enabled: true,
		Resources: []string{"catalog"}, Types: []string{"movie"},
	}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	res, err := toolSearchTitle(context.Background(), reqTool("search_title", map[string]any{
		"query": "q", "type": "movie", "year": float64(0),
	}))
	if err != nil || res == nil || res.IsError {
		t.Fatalf("err %v %#v", err, res)
	}
}

func TestToolGetStreams(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	manifest := `{"id":"s1","name":"S1","resources":[{"name":"stream","types":["movie"],"idPrefixes":["tt"]}],"types":["movie"]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "manifest.json"):
			_, _ = w.Write([]byte(manifest))
		case strings.Contains(r.URL.Path, "/stream/"):
			_, _ = w.Write([]byte(`{"streams":[{"name":"n","url":"https://u"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	cfg.TMDB.APIKey = ""
	cfg.Addons = []config.AddonEntry{{
		ID: "s1", ManifestURL: srv.URL + "/manifest.json", Enabled: true,
		Resources: []string{"stream"}, Types: []string{"movie"},
	}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}
	res, err := toolGetStreams(context.Background(), reqTool("get_streams", map[string]any{
		"imdb_id": "tt1", "type": "movie", "season": float64(0), "episode": float64(0),
	}))
	if err != nil || res == nil || res.IsError {
		t.Fatalf("err %v %#v", err, res)
	}
}

func TestToolInstallAndRemoveAddon(t *testing.T) {
	setupTempConfig(t)
	manifest := `{"id":"com.rm","name":"RM","version":"1","resources":["stream"],"types":["movie"],"catalogs":[]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(manifest))
	}))
	t.Cleanup(srv.Close)
	if _, err := toolInstallAddon(context.Background(), reqTool("install_addon", map[string]any{
		"manifest_url": srv.URL + "/manifest.json",
	})); err != nil {
		t.Fatal(err)
	}
	res, err := toolRemoveAddon(context.Background(), reqTool("remove_addon", map[string]any{"id": "com.rm"}))
	if err != nil || res == nil || res.IsError {
		t.Fatalf("err %v", err)
	}
}

func TestToolGetMeta(t *testing.T) {
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
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	res, err := toolGetMeta(context.Background(), reqTool("get_meta", map[string]any{
		"imdb_id": "tt1", "type": "movie",
	}))
	if err != nil || res == nil || res.IsError {
		t.Fatalf("err %v %#v", err, res)
	}
}

func TestToolResolveIMDB_errors(t *testing.T) {
	setupTempConfig(t)
	res, err := toolResolveIMDB(context.Background(), reqTool("resolve_imdb_id", map[string]any{
		"query": "x", "year": float64(0),
	}))
	if err != nil || res == nil || !res.IsError {
		t.Fatalf("expected error result, got err=%v isError=%v", err, res != nil && res.IsError)
	}
}

func TestToolSearchTitle_appError_emptyQuery(t *testing.T) {
	setupTempConfig(t)
	res, err := toolSearchTitle(context.Background(), reqTool("search_title", map[string]any{
		"query": "", "type": "movie",
	}))
	if err != nil || res == nil || !res.IsError {
		t.Fatal("expected error")
	}
}
