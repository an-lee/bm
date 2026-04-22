package streams

import (
	"testing"

	"bm/internal/stremio"
)

func TestResolvedStream_PlayableURL(t *testing.T) {
	t.Parallel()
	rs := ResolvedStream{Stream: stremio.Stream{URL: "https://x"}}
	if rs.PlayableURL() != "https://x" {
		t.Fatal()
	}
}

func TestDedupeKey_httpPlayable(t *testing.T) {
	t.Parallel()
	k := dedupeKey(stremio.Stream{URL: "https://x", InfoHash: "ab"})
	if k != "url:https://x" {
		t.Fatalf("got %q", k)
	}
}

func TestDedupeKey_magnet(t *testing.T) {
	t.Parallel()
	k := dedupeKey(stremio.Stream{InfoHash: "AB", Name: "n"})
	if k != "magnet:ab" {
		t.Fatalf("got %q", k)
	}
}

func TestDedupeKey_rawURLWhenNoHTTPPrefix(t *testing.T) {
	t.Parallel()
	k := dedupeKey(stremio.Stream{URL: "ftp://x"})
	if k != "url:ftp://x" {
		t.Fatalf("got %q", k)
	}
}

func TestDedupeKey_rawFallback(t *testing.T) {
	t.Parallel()
	k := dedupeKey(stremio.Stream{Name: "a", Title: "b"})
	if len(k) < 10 || k[:4] != "raw:" {
		t.Fatalf("got %q", k)
	}
}

func TestDedupeStreams(t *testing.T) {
	t.Parallel()
	in := []ResolvedStream{
		{DedupeKey: "k1", Stream: stremio.Stream{URL: "https://a"}},
		{DedupeKey: "k1", Stream: stremio.Stream{URL: "https://b"}},
		{Stream: stremio.Stream{URL: "https://c"}},
	}
	out := dedupeStreams(in)
	if len(out) != 2 {
		t.Fatalf("len %d", len(out))
	}
	if out[0].URL != "https://a" || out[1].URL != "https://c" {
		t.Fatalf("order/contents: %#v", out)
	}
}

func TestDedupeStreams_emptyDedupeKeyUsesDedupeKeyFromStream(t *testing.T) {
	t.Parallel()
	same := stremio.Stream{URL: "https://dup"}
	in := []ResolvedStream{{Stream: same}, {Stream: same}}
	out := dedupeStreams(in)
	if len(out) != 1 {
		t.Fatalf("len %d", len(out))
	}
}
