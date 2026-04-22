package tui

import (
	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
)

const tabSearch, tabStreams, tabAddons, tabSettings = 0, 1, 2, 3

// Browse catalog source (after a successful catalog/search load).
const (
	browsePopular = iota
	browseFeatured
	browseSearch
)

const (
	stageEpisodes = iota
	stageStreams
)

type searchDoneMsg struct {
	items  []search.TitleResult
	source int
}

type searchErrMsg struct{ err error }

type metaDoneMsg struct{ meta *stremio.Meta }
type metaErrMsg struct{ err error }

type streamsDoneMsg []streams.ResolvedStream
type streamsErrMsg struct{ err error }
type statusMsg struct {
	err  error
	text string
}
type toastClearMsg struct{}

// streamAddonTab is one filter chip on the Streams tab ("All" or a single addon).
type streamAddonTab struct {
	label   string
	addonID string // empty means all addons (first tab only)
}
