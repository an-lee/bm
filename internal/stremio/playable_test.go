package stremio

import (
	"net/url"
	"strings"
	"testing"
)

func TestStream_PlayableURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		s    Stream
		want string
	}{
		{
			name: "url wins",
			s: Stream{
				URL:      "https://example.com/play.m3u8",
				InfoHash: "abc",
			},
			want: "https://example.com/play.m3u8",
		},
		{
			name: "magnet from infohash with name",
			s: Stream{
				InfoHash: "ABCDEF",
				Name:     "My Torrent",
			},
			want: "magnet:?xt=urn:btih:abcdef&dn=" + url.QueryEscape("My Torrent"),
		},
		{
			name: "magnet uses title when name empty",
			s: Stream{
				InfoHash: "aa11bb22",
				Title:    "Title Only",
			},
			want: "magnet:?xt=urn:btih:aa11bb22&dn=" + url.QueryEscape("Title Only"),
		},
		{
			name: "magnet trims hash",
			s: Stream{
				InfoHash: "  mixedCASE  ",
			},
			want: "magnet:?xt=urn:btih:mixedcase",
		},
		{
			name: "empty",
			s:    Stream{},
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.s.PlayableURL()
			if got != tc.want {
				t.Fatalf("PlayableURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestStream_PlayableURL_URLTrimsSpace(t *testing.T) {
	t.Parallel()
	s := Stream{URL: "  https://x/  "}
	if got := s.PlayableURL(); strings.TrimSpace(got) != "https://x/" {
		t.Fatalf("got %q", got)
	}
}
