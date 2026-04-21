package stremio

// Manifest describes a Stremio addon (manifest.json).
type Manifest struct {
	ID              string        `json:"id"`
	Version         string        `json:"version"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Resources       jsonResources `json:"resources"`
	Types           []string      `json:"types"`
	IDPrefixes      []string      `json:"idPrefixes"`
	Catalogs        []CatalogDef  `json:"catalogs"`
	BehaviorHints   BehaviorHints `json:"behaviorHints"`
	CatalogBehavior *CatalogHints `json:"catalogBehavior,omitempty"`
}

// CatalogHints may contain configuration URL pattern.
type CatalogHints struct {
	Configurable bool `json:"configurable"`
}

// BehaviorHints on manifest.
type BehaviorHints struct {
	Configurable          bool   `json:"configurable"`
	ConfigurationRequired bool   `json:"configurationRequired"`
	OpenURLTemplate       string `json:"openUrlTemplate,omitempty"`
}

// CatalogDef is one catalog entry from manifest.
type CatalogDef struct {
	Type           string   `json:"type"`
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ExtraSupported []string `json:"extraSupported"`
}

// Meta is a Stremio meta document (movie/series).
type Meta struct {
	ID          string   `json:"id"`
	IMDBID      string   `json:"imdb_id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ReleaseInfo string   `json:"releaseInfo"`
	Year        string   `json:"year,omitempty"`
	Poster      string   `json:"poster,omitempty"`
	Background  string   `json:"background,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Videos      []Video  `json:"videos,omitempty"`
}

// Video is an episode or special in a series meta.
type Video struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Released    string `json:"released,omitempty"`
	Season      int    `json:"season"`
	Episode     int    `json:"episode"`
	IMDBSeason  int    `json:"imdbSeason,omitempty"`
	IMDBEpisode int    `json:"imdbEpisode,omitempty"`
}

// CatalogResponse is returned from /catalog/...json.
type CatalogResponse struct {
	Metas []Meta `json:"metas"`
}

// MetaResponse wraps a single meta.
type MetaResponse struct {
	Meta Meta `json:"meta"`
}

// StreamResponse is returned from /stream/...json.
type StreamResponse struct {
	Streams []Stream `json:"streams"`
}

// Stream is one playable source.
type Stream struct {
	Name          string         `json:"name"`
	Title         string         `json:"title"`
	URL           string         `json:"url"`
	InfoHash      string         `json:"infoHash"`
	FileIdx       *int           `json:"fileIdx"`
	BehaviorHints map[string]any `json:"behaviorHints"`
}

// SubtitlesResponse from /subtitles/...json.
type SubtitlesResponse struct {
	Subtitles []SubtitleTrack `json:"subtitles"`
}

// SubtitleTrack is one subtitle entry.
type SubtitleTrack struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Lang string `json:"lang"`
}
