package addons

import (
	"context"
	"fmt"
	"strings"

	"bm/internal/config"
	"bm/internal/stremio"
)

// Registry manages installed addons and manifest metadata.
type Registry struct {
	cfg    *config.Config
	client *stremio.Client
}

// NewRegistry returns a registry backed by config.
func NewRegistry(cfg *config.Config, client *stremio.Client) *Registry {
	return &Registry{cfg: cfg, client: client}
}

// List returns enabled addons (or all with includeDisabled).
func (r *Registry) List(includeDisabled bool) []config.AddonEntry {
	out := make([]config.AddonEntry, 0, len(r.cfg.Addons))
	for _, a := range r.cfg.Addons {
		if !includeDisabled && !a.Enabled {
			continue
		}
		out = append(out, a)
	}
	return out
}

// GetManifest fetches live manifest for an addon entry.
func (r *Registry) GetManifest(ctx context.Context, a config.AddonEntry) (*stremio.Manifest, error) {
	return r.client.GetManifest(ctx, a.ManifestURL)
}

// Install fetches manifest, merges metadata into config, and saves.
func (r *Registry) Install(ctx context.Context, manifestURL string) (*config.AddonEntry, error) {
	manifestURL = strings.TrimSpace(manifestURL)
	if manifestURL == "" {
		return nil, fmt.Errorf("manifest URL is required")
	}
	m, err := r.client.GetManifest(ctx, manifestURL)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	entry := config.AddonEntry{
		ID:          m.ID,
		Name:        m.Name,
		ManifestURL: manifestURL,
		Enabled:     true,
		Resources:   m.Resources.Names(),
		Types:       append([]string(nil), m.Types...),
		IDPrefixes:  append([]string(nil), m.IDPrefixes...),
	}
	if m.BehaviorHints.OpenURLTemplate != "" {
		entry.ConfigurationURL = m.BehaviorHints.OpenURLTemplate
	}
	// De-dupe by id
	addons := make([]config.AddonEntry, 0, len(r.cfg.Addons)+1)
	for _, existing := range r.cfg.Addons {
		if existing.ID == entry.ID {
			continue
		}
		addons = append(addons, existing)
	}
	addons = append(addons, entry)
	r.cfg.Addons = addons
	if err := config.Write(r.cfg); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Remove drops an addon by id.
func (r *Registry) Remove(id string) error {
	id = strings.TrimSpace(id)
	filtered := make([]config.AddonEntry, 0, len(r.cfg.Addons))
	for _, a := range r.cfg.Addons {
		if a.ID != id {
			filtered = append(filtered, a)
		}
	}
	if len(filtered) == len(r.cfg.Addons) {
		return fmt.Errorf("addon not found: %s", id)
	}
	r.cfg.Addons = filtered
	return config.Write(r.cfg)
}

// RefreshMetadata re-fetches manifests for all addons and updates stored fields.
func (r *Registry) RefreshMetadata(ctx context.Context) error {
	for i := range r.cfg.Addons {
		m, err := r.client.GetManifest(ctx, r.cfg.Addons[i].ManifestURL)
		if err != nil {
			continue
		}
		r.cfg.Addons[i].Name = m.Name
		r.cfg.Addons[i].Resources = m.Resources.Names()
		r.cfg.Addons[i].Types = append([]string(nil), m.Types...)
		r.cfg.Addons[i].IDPrefixes = append([]string(nil), m.IDPrefixes...)
		if m.BehaviorHints.OpenURLTemplate != "" {
			r.cfg.Addons[i].ConfigurationURL = m.BehaviorHints.OpenURLTemplate
		}
	}
	return config.Write(r.cfg)
}

// StreamAddons returns addons that can resolve streams for the given meta.
func (r *Registry) StreamAddons(ctx context.Context, metaType, imdbID string) ([]config.AddonEntry, error) {
	baseimdb := imdbID
	if i := strings.Index(imdbID, ":"); i >= 0 {
		baseimdb = imdbID[:i]
	}
	var out []config.AddonEntry
	for _, a := range r.cfg.Addons {
		if !a.Enabled {
			continue
		}
		m, err := r.client.GetManifest(ctx, a.ManifestURL)
		if err != nil {
			continue
		}
		if !m.Resources.SupportsStream(metaType, baseimdb) {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

// CatalogAddons returns addons that expose catalog search for mediaType.
func (r *Registry) CatalogAddons(ctx context.Context, mediaType string) ([]config.AddonEntry, error) {
	var out []config.AddonEntry
	for _, a := range r.cfg.Addons {
		if !a.Enabled {
			continue
		}
		m, err := r.client.GetManifest(ctx, a.ManifestURL)
		if err != nil {
			continue
		}
		if !contains(m.Resources.Names(), "catalog") {
			continue
		}
		if _, ok := PickSearchCatalog(m, mediaType); ok {
			out = append(out, a)
		}
	}
	return out, nil
}

// PickSearchCatalog returns the catalog id that supports search for the media type.
func PickSearchCatalog(m *stremio.Manifest, mediaType string) (catalogID string, ok bool) {
	for _, c := range m.Catalogs {
		if c.Type != mediaType {
			continue
		}
		for _, ex := range c.ExtraSupported {
			if ex == "search" {
				return c.ID, true
			}
		}
	}
	return "", false
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
