package tui

import (
	"testing"

	"bm/internal/search"
	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_streams_reloadMovie(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "movie", IMDBID: "tt9"}
	m.streamsStage = stageStreams
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal()
	}
	_ = out
}
