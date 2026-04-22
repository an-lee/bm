package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"bm/internal/config"
	"bm/internal/testxdg"
)

func execCLI(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	jsonOutput = false
	var bufOut, bufErr bytes.Buffer
	rootCmd.SetOut(&bufOut)
	rootCmd.SetErr(&bufErr)
	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	rootCmd.SetArgs(nil)
	return bufOut.String(), bufErr.String(), err
}

func TestCLI_configPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	_, _, err = execCLI(t, "config", "path")
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatal(err)
	}
	outB, _ := io.ReadAll(r)
	_ = r.Close()
	out := string(outB)
	if !strings.Contains(out, "config.toml") {
		t.Fatalf("out %q", out)
	}
}

func TestCLI_addonsList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	out, _, err := execCLI(t, "addons", "list")
	if err != nil || !strings.Contains(out, "Cinemeta") {
		t.Fatalf("err %v out %q", err, out)
	}
}

func TestCLI_configSetGet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	_, _, err := execCLI(t, "config", "set", "ui.default_type", "series")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	_, _, err = execCLI(t, "config", "get", "ui.default_type")
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatal(err)
	}
	outB, _ := io.ReadAll(r)
	_ = r.Close()
	out := string(outB)
	if strings.TrimSpace(out) != "series" {
		t.Fatalf("out %q", out)
	}
}

func TestCLI_search_JSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	t.Setenv("TMDB_API_KEY", "")
	t.Setenv("BM_TMDB_API_KEY", "")
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

	cfg := &config.Config{}
	cfg.UI.DefaultType = "movie"
	cfg.TMDB.APIKey = ""
	cfg.Addons = []config.AddonEntry{{
		ID: "cat", Name: "C", ManifestURL: srv.URL + "/manifest.json",
		Enabled: true, Resources: []string{"catalog"}, Types: []string{"movie"},
	}}
	if err := config.Write(cfg); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	rootCmd.SetArgs([]string{"search", "hello", "--type", "movie", "--json"})
	rootErr := rootCmd.Execute()
	_ = w.Close()
	os.Stdout = old
	jsonOutput = false
	if rootErr != nil {
		t.Fatal(rootErr)
	}
	body, _ := io.ReadAll(r)
	_ = r.Close()
	var parsed []map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("json %s err %v", body, err)
	}
	if len(parsed) != 1 {
		t.Fatalf("%s", body)
	}
}

func TestCLI_stream_seriesValidation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	if _, err := config.Load(); err != nil {
		t.Fatal(err)
	}
	_, _, err := execCLI(t, "stream", "tt123", "--type", "series")
	if err == nil || !strings.Contains(err.Error(), "season") {
		t.Fatalf("err %v", err)
	}
}

func TestCLI_search_wrongArgCount(t *testing.T) {
	_, _, err := execCLI(t, "search")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCLI_stream_movie_JSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	t.Setenv("TMDB_API_KEY", "")
	manifest := `{"id":"s1","name":"S1","resources":[{"name":"stream","types":["movie"],"idPrefixes":["tt"]}],"types":["movie"]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "manifest.json"):
			_, _ = w.Write([]byte(manifest))
		case strings.Contains(r.URL.Path, "/stream/"):
			_, _ = w.Write([]byte(`{"streams":[{"name":"n","title":"t","url":"https://play.example/"}]}`))
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
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	rootCmd.SetArgs([]string{"stream", "tt1", "--type", "movie", "--json"})
	rootErr := rootCmd.Execute()
	_ = w.Close()
	os.Stdout = old
	jsonOutput = false
	rootCmd.SetArgs(nil)
	if rootErr != nil {
		t.Fatal(rootErr)
	}
	body, _ := io.ReadAll(r)
	_ = r.Close()
	var parsed []map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("%s %v", body, err)
	}
	if len(parsed) != 1 {
		t.Fatalf("%s", body)
	}
}
