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
)

func TestService_tmdbExternalIDs_badStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	s.http = &http.Client{Transport: &rewriteToServer{srv: srv}}
	_, err := s.tmdbExternalIDs(context.Background(), "k", "movie", 1)
	if err == nil {
		t.Fatal()
	}
}

func TestService_tmdbExternalIDs_unsupportedMedia(t *testing.T) {
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	_, err := s.tmdbExternalIDs(context.Background(), "k", "person", 1)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("%v", err)
	}
}

func TestService_searchTMDB_nonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusTeapot)
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	s.http = &http.Client{Transport: &rewriteToServer{srv: srv}}
	_, err := s.searchTMDB(context.Background(), "q", "movie", 0)
	if err == nil || !strings.Contains(err.Error(), "TMDB HTTP") {
		t.Fatalf("%v", err)
	}
}

func TestService_searchTMDB_badJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{`))
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	s.http = &http.Client{Transport: &rewriteToServer{srv: srv}}
	_, err := s.searchTMDB(context.Background(), "q", "movie", 0)
	if err == nil {
		t.Fatal()
	}
}
