package stremio

import (
	"fmt"
	"net/url"
	"strings"
)

// PlayableURL returns HTTP URL or a magnet link for torrent-backed streams.
func (s Stream) PlayableURL() string {
	if u := strings.TrimSpace(s.URL); u != "" {
		return u
	}
	h := strings.ToLower(strings.TrimSpace(s.InfoHash))
	if h == "" {
		return ""
	}
	dn := strings.TrimSpace(s.Name)
	if dn == "" {
		dn = strings.TrimSpace(s.Title)
	}
	m := fmt.Sprintf("magnet:?xt=urn:btih:%s", h)
	if dn != "" {
		m += "&dn=" + url.QueryEscape(dn)
	}
	return m
}
