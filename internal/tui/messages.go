package tui

import (
	"bm/internal/search"
	"bm/internal/streams"
)

const tabSearch, tabStreams, tabAddons, tabSettings = 0, 1, 2, 3

type searchDoneMsg []search.TitleResult
type searchErrMsg struct{ err error }
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
