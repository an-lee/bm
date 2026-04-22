package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bm/internal/addons"
	"bm/internal/config"
	"bm/internal/stremio"
)

// TitleResult is a unified search hit for UI, CLI, and MCP.
type TitleResult struct {
	Title    string `json:"title"`
	Year     string `json:"year,omitempty"`
	Type     string `json:"type"` // movie | series
	IMDBID   string `json:"imdb_id"`
	TMDBID   int    `json:"tmdb_id,omitempty"`
	Overview string `json:"overview,omitempty"`
}

// Service resolves titles via TMDB (when configured) or Stremio catalog addons.
type Service struct {
	cfg      *config.Config
	registry *addons.Registry
	client   *stremio.Client
	http     *http.Client
}

// NewService constructs a search service.
func NewService(cfg *config.Config, reg *addons.Registry, client *stremio.Client) *Service {
	return &Service{
		cfg:      cfg,
		registry: reg,
		client:   client,
		http:     &http.Client{Timeout: 30 * time.Second},
	}
}

// Search returns ranked title results.
func (s *Service) Search(ctx context.Context, query, mediaType string, year int) ([]TitleResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if mediaType == "" {
		mediaType = s.cfg.UI.DefaultType
	}
	if s.cfg.TMDB.APIKey != "" {
		return s.searchTMDB(ctx, query, mediaType, year)
	}
	return s.searchCatalog(ctx, query, mediaType)
}

// CinemetaPopular returns titles from Cinemeta's "Popular" catalog (id top).
func (s *Service) CinemetaPopular(ctx context.Context, mediaType string, skip int) ([]TitleResult, error) {
	return s.cinemetaCatalog(ctx, mediaType, "top", skip)
}

// CinemetaFeatured returns titles from Cinemeta's "Featured" catalog (id imdbRating).
func (s *Service) CinemetaFeatured(ctx context.Context, mediaType string, skip int) ([]TitleResult, error) {
	return s.cinemetaCatalog(ctx, mediaType, "imdbRating", skip)
}

func (s *Service) cinemetaCatalog(ctx context.Context, mediaType, catalogID string, skip int) ([]TitleResult, error) {
	if mediaType != "movie" && mediaType != "series" {
		mediaType = s.cfg.UI.DefaultType
	}
	entry, err := s.findEnabledCinemeta()
	if err != nil {
		return nil, err
	}
	m, err := s.client.GetManifest(ctx, entry.ManifestURL)
	if err != nil {
		return nil, fmt.Errorf("cinemeta manifest: %w", err)
	}
	if !manifestHasCatalog(m, mediaType, catalogID) {
		return nil, fmt.Errorf("cinemeta has no %s catalog %q", mediaType, catalogID)
	}
	base, err := stremio.BaseFromManifestURL(entry.ManifestURL)
	if err != nil {
		return nil, err
	}
	var extras []string
	if skip > 0 {
		extras = append(extras, fmt.Sprintf("skip=%d", skip))
	}
	resp, err := s.client.CatalogGet(ctx, base, mediaType, catalogID, extras...)
	if err != nil {
		return nil, err
	}
	return titleResultsFromMetas(resp.Metas), nil
}

func (s *Service) findEnabledCinemeta() (config.AddonEntry, error) {
	for _, a := range s.cfg.Addons {
		if a.ID == config.CinemetaAddonID && a.Enabled {
			return a, nil
		}
	}
	return config.AddonEntry{}, fmt.Errorf("Cinemeta addon is not installed or disabled; enable it to browse catalogs")
}

func manifestHasCatalog(m *stremio.Manifest, mediaType, catalogID string) bool {
	for _, c := range m.Catalogs {
		if c.Type == mediaType && c.ID == catalogID {
			return true
		}
	}
	return false
}

func titleResultsFromMetas(metas []stremio.Meta) []TitleResult {
	out := make([]TitleResult, 0, len(metas))
	for _, meta := range metas {
		title := meta.Name
		if title == "" {
			title = meta.ID
		}
		imdb := meta.IMDBID
		if imdb == "" {
			imdb = meta.ID
		}
		out = append(out, TitleResult{
			Title:    title,
			Year:     firstNonEmpty(meta.ReleaseInfo, meta.Year),
			Type:     meta.Type,
			IMDBID:   imdb,
			Overview: meta.Description,
		})
	}
	return out
}

// ResolveIMDBID uses TMDB multi-search + external_ids (requires API key).
func (s *Service) ResolveIMDBID(ctx context.Context, query string, year int) (string, error) {
	if s.cfg.TMDB.APIKey == "" {
		return "", fmt.Errorf("TMDB API key is not configured; set tmdb.api_key")
	}
	res, err := s.searchTMDB(ctx, query, "", year)
	if err != nil {
		return "", err
	}
	if len(res) == 0 {
		return "", fmt.Errorf("no results for %q", query)
	}
	return res[0].IMDBID, nil
}

func (s *Service) searchCatalog(ctx context.Context, query, mediaType string) ([]TitleResult, error) {
	list, err := s.registry.CatalogAddons(ctx, mediaType)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no catalog addons support search for type %q", mediaType)
	}
	var lastErr error
	for _, addon := range list {
		base, err := stremio.BaseFromManifestURL(addon.ManifestURL)
		if err != nil {
			lastErr = err
			continue
		}
		m, err := s.client.GetManifest(ctx, addon.ManifestURL)
		if err != nil {
			lastErr = err
			continue
		}
		catID, ok := addons.PickSearchCatalog(m, mediaType)
		if !ok {
			continue
		}
		resp, err := s.client.CatalogSearch(ctx, base, mediaType, catID, query)
		if err != nil {
			lastErr = err
			continue
		}
		return titleResultsFromMetas(resp.Metas), nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("catalog search failed")
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

type tmdbMultiResponse struct {
	Results []struct {
		ID          int64  `json:"id"`
		MediaType   string `json:"media_type"`
		Title       string `json:"title"`
		Name        string `json:"name"`
		ReleaseDate string `json:"release_date"`
		FirstAir    string `json:"first_air_date"`
		Overview    string `json:"overview"`
	} `json:"results"`
}

type tmdbExternalIDs struct {
	IMDBID string `json:"imdb_id"`
}

func (s *Service) searchTMDB(ctx context.Context, query, mediaType string, year int) ([]TitleResult, error) {
	key := s.cfg.TMDB.APIKey
	u, _ := url.Parse("https://api.themoviedb.org/3/search/multi")
	q := u.Query()
	q.Set("api_key", key)
	q.Set("query", query)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var parsed tmdbMultiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	out := make([]TitleResult, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		if r.MediaType == "person" {
			continue
		}
		mt := r.MediaType
		if mediaType != "" && mt != mediaType {
			continue
		}
		title := r.Title
		if title == "" {
			title = r.Name
		}
		date := r.ReleaseDate
		if date == "" {
			date = r.FirstAir
		}
		y := yearString(date)
		if year > 0 {
			ry, _ := strconvYear(y)
			if ry != year {
				continue
			}
		}
		imdb, err := s.tmdbExternalIDs(ctx, key, mt, r.ID)
		if err != nil || imdb == "" {
			continue
		}
		out = append(out, TitleResult{
			Title:    title,
			Year:     y,
			Type:     mt,
			IMDBID:   imdb,
			TMDBID:   int(r.ID),
			Overview: r.Overview,
		})
	}
	return out, nil
}

func (s *Service) tmdbExternalIDs(ctx context.Context, apiKey, media string, id int64) (string, error) {
	var path string
	switch media {
	case "movie":
		path = fmt.Sprintf("https://api.themoviedb.org/3/movie/%d/external_ids?api_key=%s", id, url.QueryEscape(apiKey))
	case "tv":
		path = fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/external_ids?api_key=%s", id, url.QueryEscape(apiKey))
	default:
		return "", fmt.Errorf("unsupported media type %q", media)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("TMDB external_ids %d", resp.StatusCode)
	}
	var ext tmdbExternalIDs
	if err := json.Unmarshal(body, &ext); err != nil {
		return "", err
	}
	return ext.IMDBID, nil
}

func yearString(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return ""
}

func strconvYear(y string) (int, error) {
	if len(y) < 4 {
		return 0, fmt.Errorf("short year")
	}
	return strconv.Atoi(y[:4])
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
