package stremio

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBaseFromManifestURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want string
	}{
		{"https://addon.example/manifest.json", "https://addon.example"},
		{"https://addon.example/manifest.json/", "https://addon.example"},
		{"https://addon.example/foo/manifest.json", "https://addon.example/foo"},
		{"https://addon.example/", "https://addon.example"},
	}
	for _, tc := range cases {
		got, err := BaseFromManifestURL(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("BaseFromManifestURL(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBaseFromManifestURL_invalid(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "not-a-url", "http:///manifest.json"} {
		if _, err := BaseFromManifestURL(in); err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}

func TestStreamItemID(t *testing.T) {
	t.Parallel()
	if got := StreamItemID("  tt1  ", "series", 2, 3); got != "tt1:2:3" {
		t.Fatalf("series: %q", got)
	}
	if got := StreamItemID("tt1", "series", 0, 3); got != "tt1" {
		t.Fatalf("missing season: %q", got)
	}
	if got := StreamItemID("tt1", "movie", 1, 1); got != "tt1" {
		t.Fatalf("movie: %q", got)
	}
}

func TestJoinBasePath(t *testing.T) {
	t.Parallel()
	got, err := joinBasePath("https://x/y", "/catalog/movie/top.json")
	if err != nil {
		t.Fatal(err)
	}
	// Absolute suffix replaces host path per URL resolution rules.
	if got != "https://x/catalog/movie/top.json" {
		t.Fatalf("got %q", got)
	}
	got2, err := joinBasePath("https://x/y", "meta/foo.json")
	if err != nil || got2 != "https://x/meta/foo.json" {
		t.Fatalf("got %q err %v", got2, err)
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	if truncate("hi", 10) != "hi" {
		t.Fatal()
	}
	s := strings.Repeat("a", 250)
	out := truncate(s, 10)
	if len(out) != 13 || !strings.HasSuffix(out, "...") {
		t.Fatalf("len %d %q", len(out), out)
	}
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	c := NewClient()
	if c == nil || c.HTTP == nil {
		t.Fatal("expected client")
	}
}

func TestClient_GetManifest_success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"id":"x","name":"Test"}`)
	}))
	t.Cleanup(srv.Close)
	c := &Client{HTTP: srv.Client()}
	c.HTTP.Timeout = 5 * time.Second
	ctx := context.Background()
	m, err := c.GetManifest(ctx, srv.URL+"/manifest.json")
	if err != nil || m == nil || m.ID != "x" {
		t.Fatalf("err %v m %#v", err, m)
	}
}

func TestClient_GetManifest_decodeError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{`)
	}))
	t.Cleanup(srv.Close)
	c := &Client{HTTP: srv.Client()}
	_, err := c.GetManifest(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "decode manifest") {
		t.Fatalf("err %v", err)
	}
}

func TestClient_GetManifest_nonOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, strings.Repeat("e", 300), http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	c := &Client{HTTP: srv.Client()}
	_, err := c.GetManifest(context.Background(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("err %v", err)
	}
}

func TestClient_get_retryThenSuccess(t *testing.T) {
	t.Parallel()
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n.Add(1) < 2 {
			http.Error(w, "fail", http.StatusInternalServerError)
			return
		}
		_, _ = io.WriteString(w, `{"metas":[]}`)
	}))
	t.Cleanup(srv.Close)
	c := &Client{HTTP: srv.Client()}
	c.HTTP.Timeout = 5 * time.Second
	base := strings.TrimSuffix(srv.URL, "/")
	_, err := c.CatalogGet(context.Background(), base, "movie", "top")
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_CatalogGet_badJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{`)
	}))
	t.Cleanup(srv.Close)
	c := &Client{HTTP: srv.Client()}
	_, err := c.CatalogGet(context.Background(), srv.URL, "movie", "top")
	if err == nil || !strings.Contains(err.Error(), "decode catalog") {
		t.Fatalf("err %v", err)
	}
}

func TestClient_CatalogSearch_and_GetMeta_GetStreams_GetSubtitles(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/search="):
			_, _ = io.WriteString(w, `{"metas":[{"id":"tt1","name":"A"}]}`)
		case strings.Contains(r.URL.Path, "/meta/"):
			_, _ = io.WriteString(w, `{"meta":{"id":"tt1","type":"movie","name":"A"}}`)
		case strings.Contains(r.URL.Path, "/stream/"):
			_, _ = io.WriteString(w, `{"streams":[{"name":"s"}]}`)
		case strings.Contains(r.URL.Path, "/subtitles/"):
			_, _ = io.WriteString(w, `{"subtitles":[{"id":"1","url":"u","lang":"en"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	c := &Client{HTTP: srv.Client()}
	c.HTTP.Timeout = 5 * time.Second
	base := strings.TrimSuffix(srv.URL, "/")

	cr, err := c.CatalogSearch(context.Background(), base, "movie", "top", "hello world")
	if err != nil || len(cr.Metas) != 1 {
		t.Fatalf("catalog search: %v %#v", err, cr)
	}

	meta, err := c.GetMeta(context.Background(), base, "movie", "tt1")
	if err != nil || meta.Name != "A" {
		t.Fatalf("meta: %v %#v", err, meta)
	}

	streams, err := c.GetStreams(context.Background(), base, "movie", "tt1")
	if err != nil || len(streams) != 1 {
		t.Fatalf("streams: %v %#v", err, streams)
	}

	subs, err := c.GetSubtitles(context.Background(), base, "movie", "tt1")
	if err != nil || len(subs) != 1 || subs[0].Lang != "en" {
		t.Fatalf("subs: %v %#v", err, subs)
	}
}

func TestClient_get_contextCancelDuringBackoff(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	c := &Client{HTTP: srv.Client()}
	c.HTTP.Timeout = 5 * time.Second
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.GetManifest(ctx, srv.URL)
	if err == nil || err != context.Canceled {
		t.Fatalf("want canceled, got %v", err)
	}
}
