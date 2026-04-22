package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"bm/internal/addons"
	"bm/internal/config"
	"bm/internal/stremio"
)

// rewriteToServer sends requests to srv instead of the original host (for TMDB tests).
type rewriteToServer struct {
	srv    *httptest.Server
	inner  http.RoundTripper
}

func (rw *rewriteToServer) RoundTrip(req *http.Request) (*http.Response, error) {
	base, err := url.Parse(rw.srv.URL)
	if err != nil {
		return nil, err
	}
	u := *req.URL
	u.Scheme = base.Scheme
	u.Host = base.Host
	req2 := req.Clone(req.Context())
	req2.URL = &u
	if rw.inner == nil {
		rw.inner = http.DefaultTransport
	}
	return rw.inner.RoundTrip(req2)
}

func TestService_searchTMDB_roundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/search/multi"):
			_, _ = w.Write([]byte(`{"results":[{"id":42,"media_type":"movie","title":"Inception","release_date":"2010-01-01","overview":"o"}]}`))
		case strings.Contains(r.URL.Path, "/movie/42/external_ids"):
			_, _ = w.Write([]byte(`{"imdb_id":"tt1375666"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	s.http = &http.Client{Transport: &rewriteToServer{srv: srv}}

	out, err := s.searchTMDB(context.Background(), "inc", "movie", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].IMDBID != "tt1375666" {
		t.Fatalf("%#v", out)
	}
}

func TestService_searchTMDB_yearFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search/multi") {
			_, _ = w.Write([]byte(`{"results":[{"id":1,"media_type":"movie","title":"Old","release_date":"1999-01-01"}]}`))
			return
		}
		if strings.Contains(r.URL.Path, "/movie/1/external_ids") {
			_, _ = w.Write([]byte(`{"imdb_id":"tt1"}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	s.http = &http.Client{Transport: &rewriteToServer{srv: srv}}
	out, err := s.searchTMDB(context.Background(), "x", "movie", 1999)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("%#v", out)
	}
}

func TestService_searchTMDB_skipsPersonResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search/multi") {
			_, _ = w.Write([]byte(`{"results":[{"id":1,"media_type":"person","name":"Actor"},{"id":2,"media_type":"movie","title":"Film","release_date":"2021-01-01"}]}`))
			return
		}
		if strings.Contains(r.URL.Path, "/movie/2/external_ids") {
			_, _ = w.Write([]byte(`{"imdb_id":"tt222"}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	s.http = &http.Client{Transport: &rewriteToServer{srv: srv}}
	out, err := s.searchTMDB(context.Background(), "x", "movie", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].IMDBID != "tt222" {
		t.Fatalf("%#v", out)
	}
}

func TestService_ResolveIMDBID_viaTMDB(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/search/multi"):
			_, _ = w.Write([]byte(`{"results":[{"id":7,"media_type":"tv","name":"Show","first_air_date":"2020-01-01"}]}`))
		case strings.Contains(r.URL.Path, "/tv/7/external_ids"):
			_, _ = w.Write([]byte(`{"imdb_id":"tt777"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "k"
	s := NewService(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	s.http = &http.Client{Transport: &rewriteToServer{srv: srv}}
	id, err := s.ResolveIMDBID(context.Background(), "show", 0)
	if err != nil || id != "tt777" {
		t.Fatalf("%q %v", id, err)
	}
}
