package stremio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxRetries  = 3
	baseTimeout = 45 * time.Second
)

// Client fetches Stremio addon endpoints.
type Client struct {
	HTTP *http.Client
}

// NewClient returns a client with sensible defaults.
func NewClient() *Client {
	return &Client{
		HTTP: &http.Client{Timeout: baseTimeout},
	}
}

// BaseFromManifestURL strips manifest.json (or trailing path) to addon root.
func BaseFromManifestURL(manifestURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(manifestURL))
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid manifest URL: %q", manifestURL)
	}
	u.Path = strings.TrimSuffix(u.Path, "/")
	u.Path = strings.TrimSuffix(u.Path, "/manifest.json")
	// If path was only /manifest.json, path becomes empty -> use /
	if u.Path == "" {
		u.Path = "/"
	}
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimSuffix(u.String(), "/"), nil
}

// GetManifest fetches manifest.json.
func (c *Client) GetManifest(ctx context.Context, manifestURL string) (*Manifest, error) {
	body, err := c.get(ctx, manifestURL)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return &m, nil
}

// CatalogSearch performs GET /catalog/{type}/{catalogID}/search={query}.json
func (c *Client) CatalogSearch(ctx context.Context, baseURL, catalogType, catalogID, query string) (*CatalogResponse, error) {
	u, err := joinBasePath(baseURL, fmt.Sprintf("/catalog/%s/%s/search=%s.json", catalogType, catalogID, escapeSearchSegment(query)))
	if err != nil {
		return nil, err
	}
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var out CatalogResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode catalog: %w", err)
	}
	return &out, nil
}

func escapeSearchSegment(q string) string {
	// Stremio uses encodeURIComponent-style; url.PathEscape encodes spaces as %20
	return url.PathEscape(strings.TrimSpace(q))
}

// GetMeta fetches /meta/{type}/{id}.json
func (c *Client) GetMeta(ctx context.Context, baseURL, metaType, id string) (*Meta, error) {
	u, err := catalogURL(baseURL, fmt.Sprintf("/meta/%s/%s.json", metaType, id))
	if err != nil {
		return nil, err
	}
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp MetaResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode meta: %w", err)
	}
	return &resp.Meta, nil
}

// GetStreams fetches /stream/{type}/{id}.json (id may be tt123:1:2 for episodes).
func (c *Client) GetStreams(ctx context.Context, baseURL, metaType, id string) ([]Stream, error) {
	u, err := catalogURL(baseURL, fmt.Sprintf("/stream/%s/%s.json", metaType, id))
	if err != nil {
		return nil, err
	}
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp StreamResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode streams: %w", err)
	}
	return resp.Streams, nil
}

// GetSubtitles fetches /subtitles/{type}/{id}.json
func (c *Client) GetSubtitles(ctx context.Context, baseURL, metaType, id string) ([]SubtitleTrack, error) {
	u, err := catalogURL(baseURL, fmt.Sprintf("/subtitles/%s/%s.json", metaType, id))
	if err != nil {
		return nil, err
	}
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var resp SubtitlesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode subtitles: %w", err)
	}
	return resp.Subtitles, nil
}

// StreamItemID builds stream endpoint id (movie: tt..., series: tt...:s:e).
func StreamItemID(imdbID string, metaType string, season, episode int) string {
	imdbID = strings.TrimSpace(imdbID)
	if metaType == "series" && season > 0 && episode > 0 {
		return fmt.Sprintf("%s:%d:%d", imdbID, season, episode)
	}
	return imdbID
}

func catalogURL(baseURL, rel string) (string, error) {
	return joinBasePath(baseURL, rel)
}

func joinBasePath(baseURL, suffix string) (string, error) {
	b, err := url.Parse(strings.TrimSuffix(strings.TrimSpace(baseURL), "/"))
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(suffix, "/") {
		suffix = "/" + suffix
	}
	r, err := url.Parse(suffix)
	if err != nil {
		return "", err
	}
	return b.ResolveReference(r).String(), nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 300 * time.Millisecond):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "bm/1.0 (Stremio addon client)")

		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if rerr != nil {
			lastErr = rerr
			continue
		}
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
		}
		return body, nil
	}
	if lastErr == nil {
		lastErr = errors.New("unknown HTTP error")
	}
	return nil, lastErr
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
