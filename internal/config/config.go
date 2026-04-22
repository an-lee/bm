package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

const (
	appDirName     = "bm"
	configFileName = "config"
)

// AddonEntry is one installed Stremio addon persisted in TOML.
type AddonEntry struct {
	ID          string   `mapstructure:"id"`
	Name        string   `mapstructure:"name"`
	ManifestURL string   `mapstructure:"manifest_url"`
	Enabled     bool     `mapstructure:"enabled"`
	Resources   []string `mapstructure:"resources"`
	Types       []string `mapstructure:"types"`
	IDPrefixes  []string `mapstructure:"id_prefixes"`
	// ConfigurationURL is set from manifest when present (not always persisted; refetched on demand).
	ConfigurationURL string `mapstructure:"configuration_url,omitempty"`
}

// Config is the full bm configuration.
type Config struct {
	TMDB struct {
		APIKey string `mapstructure:"api_key"`
	} `mapstructure:"tmdb"`
	UI struct {
		DefaultType string `mapstructure:"default_type"`
		// StreamOrder is one of: rank, rank-asc, addon, title, seeds, seeds-asc (see streams.NormalizeOrder).
		StreamOrder string `mapstructure:"stream_order"`
	} `mapstructure:"ui"`
	Addons []AddonEntry `mapstructure:"addons"`
}

// DefaultCinemetaManifest is the official Cinemeta catalog addon.
const DefaultCinemetaManifest = "https://v3-cinemeta.strem.io/manifest.json"

// CinemetaAddonID is Stremio's official Cinemeta catalog addon id.
const CinemetaAddonID = "com.linvo.cinemeta"

func configDir() (string, error) {
	dir := filepath.Join(xdg.ConfigHome, appDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Path returns the absolute path to config.toml.
func Path() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName+".toml"), nil
}

// Load reads configuration from disk, creating defaults (including Cinemeta) on first run.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(p)

	if _, statErr := os.Stat(p); errors.Is(statErr, os.ErrNotExist) {
		cfg := defaultConfig()
		if werr := Write(cfg); werr != nil {
			return nil, fmt.Errorf("write default config: %w", werr)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.UI.DefaultType == "" {
		cfg.UI.DefaultType = "movie"
	}
	return &cfg, nil
}

func defaultConfig() *Config {
	c := &Config{}
	c.UI.DefaultType = "movie"
	c.Addons = []AddonEntry{
		{
			ID:          CinemetaAddonID,
			Name:        "Cinemeta",
			ManifestURL: DefaultCinemetaManifest,
			Enabled:     true,
			Resources:   []string{"catalog", "meta", "addon_catalog"},
			Types:       []string{"movie", "series"},
			IDPrefixes:  []string{"tt"},
		},
		{
			ID:               "com.stremio.torrentio.addon",
			Name:             "Torrentio",
			ManifestURL:      "https://torrentio.strem.fun/manifest.json",
			Enabled:          true,
			Resources:        []string{"stream"},
			Types:            []string{"movie", "series", "anime", "other"},
			IDPrefixes:       []string{"tt", "kitsu"},
			ConfigurationURL: "https://torrentio.strem.fun/configure",
		},
	}
	return c
}

// Write replaces the entire config file.
func Write(cfg *Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(p)

	v.Set("tmdb", map[string]any{
		"api_key": cfg.TMDB.APIKey,
	})
	if cfg.UI.DefaultType == "" {
		cfg.UI.DefaultType = "movie"
	}
	v.Set("ui", map[string]any{
		"default_type": cfg.UI.DefaultType,
		"stream_order": cfg.UI.StreamOrder,
	})
	addonMaps := make([]map[string]any, 0, len(cfg.Addons))
	for _, a := range cfg.Addons {
		m := map[string]any{
			"id":           a.ID,
			"name":         a.Name,
			"manifest_url": a.ManifestURL,
			"enabled":      a.Enabled,
			"resources":    a.Resources,
			"types":        a.Types,
			"id_prefixes":  a.IDPrefixes,
		}
		if a.ConfigurationURL != "" {
			m["configuration_url"] = a.ConfigurationURL
		}
		addonMaps = append(addonMaps, m)
	}
	v.Set("addons", addonMaps)
	return v.WriteConfigAs(p)
}

// SetKey updates a single dotted key in config (e.g. tmdb.api_key) and saves.
func SetKey(key, value string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	switch key {
	case "tmdb.api_key":
		cfg.TMDB.APIKey = value
	case "ui.default_type":
		cfg.UI.DefaultType = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return Write(cfg)
}

// GetKey returns a config value for display.
func GetKey(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}
	switch key {
	case "tmdb.api_key":
		return cfg.TMDB.APIKey, nil
	case "ui.default_type":
		return cfg.UI.DefaultType, nil
	default:
		return "", fmt.Errorf("unknown config key: %s", key)
	}
}
