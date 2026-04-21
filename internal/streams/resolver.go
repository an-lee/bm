package streams

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"bm/internal/addons"
	"bm/internal/config"
	"bm/internal/stremio"
)

// ResolvedStream is one stream with its originating addon.
type ResolvedStream struct {
	AddonID   string `json:"addon_id"`
	AddonName string `json:"addon_name"`
	stremio.Stream
	DedupeKey string `json:"dedupe_key,omitempty"`
}

// PlayableURL returns HTTP URL or magnet for clipboard / JSON.
func (s ResolvedStream) PlayableURL() string {
	return s.Stream.PlayableURL()
}

// Resolver fans out stream requests to all capable addons.
type Resolver struct {
	cfg      *config.Config
	registry *addons.Registry
	client   *stremio.Client
}

// NewResolver constructs a resolver.
func NewResolver(cfg *config.Config, reg *addons.Registry, client *stremio.Client) *Resolver {
	return &Resolver{cfg: cfg, registry: reg, client: client}
}

// Resolve returns merged streams from all stream-capable addons.
func (r *Resolver) Resolve(ctx context.Context, imdbID, metaType string, season, episode int) ([]ResolvedStream, error) {
	streamID := stremio.StreamItemID(imdbID, metaType, season, episode)
	list, err := r.registry.StreamAddons(ctx, metaType, streamID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no stream addons available for type %q", metaType)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	out := make([]ResolvedStream, 0)

	for _, addon := range list {
		addon := addon
		wg.Add(1)
		go func() {
			defer wg.Done()
			base, err := stremio.BaseFromManifestURL(addon.ManifestURL)
			if err != nil {
				return
			}
			streams, err := r.client.GetStreams(ctx, base, metaType, streamID)
			if err != nil {
				return
			}
			m, _ := r.client.GetManifest(ctx, addon.ManifestURL)
			addonName := addon.Name
			if m != nil && m.Name != "" {
				addonName = m.Name
			}
			mu.Lock()
			for _, s := range streams {
				rs := ResolvedStream{
					AddonID:   addon.ID,
					AddonName: addonName,
					Stream:    s,
					DedupeKey: dedupeKey(s),
				}
				out = append(out, rs)
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	out = dedupeStreams(out)
	ApplySort(out, r.cfg.UI.StreamOrder)
	return out, nil
}

func dedupeKey(s stremio.Stream) string {
	u := strings.TrimSpace(s.PlayableURL())
	if u != "" && strings.HasPrefix(u, "http") {
		return "url:" + u
	}
	if u != "" && strings.HasPrefix(u, "magnet:") {
		return "magnet:" + strings.TrimPrefix(strings.ToLower(strings.TrimSpace(s.InfoHash)), "")
	}
	u = strings.TrimSpace(s.URL)
	if u != "" {
		return "url:" + u
	}
	h := strings.TrimSpace(s.InfoHash)
	if h != "" {
		idx := ""
		if s.FileIdx != nil {
			idx = fmt.Sprintf(":%d", *s.FileIdx)
		}
		return "hash:" + strings.ToLower(h) + idx
	}
	sum := sha256.Sum256([]byte(strings.Join([]string{s.Name, s.Title, s.URL, s.InfoHash}, "|")))
	return "raw:" + hex.EncodeToString(sum[:])
}

func dedupeStreams(in []ResolvedStream) []ResolvedStream {
	seen := make(map[string]struct{})
	out := make([]ResolvedStream, 0, len(in))
	for _, s := range in {
		k := s.DedupeKey
		if k == "" {
			k = dedupeKey(s.Stream)
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, s)
	}
	return out
}
