package tui

import (
	"testing"

	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_F1TogglesHelp(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyF1})
	if !out.(*rootModel).helpOpen {
		t.Fatal()
	}
}

func TestUpdate_jumpTabs234(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if out.(*rootModel).tab != tabStreams {
		t.Fatal(out.(*rootModel).tab)
	}
	out2, _ := out.(*rootModel).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if out2.(*rootModel).tab != tabAddons {
		t.Fatal()
	}
	out3, _ := out2.(*rootModel).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	if out3.(*rootModel).tab != tabSettings {
		t.Fatal()
	}
}

func TestUpdate_searchPopularKey(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if cmd == nil {
		t.Fatal()
	}
	_ = out
}

func TestUpdate_searchFeaturedKey(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	if cmd == nil {
		t.Fatal()
	}
	_ = out
}

func TestUpdate_searchToggleTypeWhenActiveWithQuery(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	m.searchActive = true
	m.searchInput.Focus()
	m.searchInput.SetValue("abc")
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	if cmd == nil {
		t.Fatal()
	}
	_ = out
}

func TestUpdate_searchEnterEmptyCloses(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	m.searchActive = true
	m.searchInput.Focus()
	m.searchInput.SetValue("   ")
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	rm := out.(*rootModel)
	if rm.searchActive {
		t.Fatal()
	}
}

func TestUpdate_searchDownBlursInput(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	m.searchActive = true
	m.searchInput.Focus()
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if out.(*rootModel).searchInput.Focused() {
		t.Fatal()
	}
}

func TestUpdate_settingsRunePassthrough(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSettings
	_, _ = m.updateSettings(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
}

func TestView_streamsEpisodesBusy(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 90
	m.height = 30
	m.tab = tabStreams
	m.selected = &search.TitleResult{Title: "S", IMDBID: "tt1", Type: "series"}
	m.streamsStage = stageEpisodes
	m.episodesBusy = true
	s := m.View()
	if len(s) < 20 {
		t.Fatal()
	}
}

func TestView_streamsWithSortHint(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 90
	m.height = 30
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "movie"}
	m.streamsStage = stageStreams
	m.streamsBusy = false
	m.allResolvedStreams = []streams.ResolvedStream{{Stream: stremio.Stream{URL: "https://u"}}}
	_ = m.applyStreamsAddonFilter()
	s := m.View()
	if len(s) < 30 {
		t.Fatal(len(s))
	}
}
