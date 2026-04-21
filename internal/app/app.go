package app

import (
	"context"
	"fmt"

	"bm/internal/addons"
	"bm/internal/config"
	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
)

// App wires core services for CLI, TUI, and MCP.
type App struct {
	Config  *config.Config
	Client  *stremio.Client
	Addons  *addons.Registry
	Search  *search.Service
	Streams *streams.Resolver
}

// New loads config and constructs services.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	client := stremio.NewClient()
	reg := addons.NewRegistry(cfg, client)
	return &App{
		Config:  cfg,
		Client:  client,
		Addons:  reg,
		Search:  search.NewService(cfg, reg, client),
		Streams: streams.NewResolver(cfg, reg, client),
	}, nil
}

// Reload refreshes configuration from disk.
func (a *App) Reload() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	a.Config = cfg
	a.Addons = addons.NewRegistry(cfg, a.Client)
	a.Search = search.NewService(cfg, a.Addons, a.Client)
	a.Streams = streams.NewResolver(cfg, a.Addons, a.Client)
	return nil
}

// Meta fetches meta for a title from the first catalog-capable addon (Cinemeta).
func (a *App) Meta(ctx context.Context, imdbID, metaType string) (*stremio.Meta, error) {
	list, err := a.Addons.CatalogAddons(ctx, metaType)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("no catalog addons for type %q", metaType)
	}
	base, err := stremio.BaseFromManifestURL(list[0].ManifestURL)
	if err != nil {
		return nil, err
	}
	return a.Client.GetMeta(ctx, base, metaType, imdbID)
}
