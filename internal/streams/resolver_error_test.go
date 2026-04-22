package streams

import (
	"context"
	"testing"

	"bm/internal/addons"
	"bm/internal/config"
	"bm/internal/stremio"
)

func TestResolver_Resolve_noStreamAddons(t *testing.T) {
	cfg := &config.Config{}
	cfg.UI.StreamOrder = "rank"
	cfg.Addons = nil
	r := NewResolver(cfg, addons.NewRegistry(cfg, stremio.NewClient()), stremio.NewClient())
	_, err := r.Resolve(context.Background(), "tt1", "movie", 0, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}
