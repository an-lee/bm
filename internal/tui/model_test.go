package tui

import (
	"errors"
	"testing"

	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRootModel(t *testing.T) {
	m := newRootModel(testApp(t))
	if m == nil || m.tab != tabSearch {
		t.Fatal()
	}
}

func TestRootModel_globalKeysBlocked(t *testing.T) {
	m := newRootModel(testApp(t))
	if m.globalKeysBlocked() {
		t.Fatal()
	}
	m.tab = tabSearch
	m.searchActive = true
	m.searchInput.Focus()
	if !m.globalKeysBlocked() {
		t.Fatal()
	}
}

func TestRootModel_effectiveMetaType(t *testing.T) {
	m := newRootModel(testApp(t))
	if got := m.effectiveMetaType(); got != "movie" {
		t.Fatal(got)
	}
	m.selected = &search.TitleResult{Type: "series"}
	if got := m.effectiveMetaType(); got != "series" {
		t.Fatal(got)
	}
}

func TestRootModel_closeSearchInput(t *testing.T) {
	m := newRootModel(testApp(t))
	m.searchActive = true
	m.searchInput.Focus()
	m.searchInput.SetValue("x")
	m.closeSearchInput()
	if m.searchActive || m.searchInput.Focused() || m.searchInput.Value() != "" {
		t.Fatal()
	}
}

func TestRootModel_WindowSizeMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	out, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if cmd != nil {
		t.Fatal("expected nil cmd")
	}
	rm := out.(*rootModel)
	if rm.width != 100 || rm.height != 40 {
		t.Fatal(rm.width, rm.height)
	}
}

func TestRootModel_helpToggle(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	rm := out.(*rootModel)
	if !rm.helpOpen {
		t.Fatal()
	}
	out2, _ := rm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if out2.(*rootModel).helpOpen {
		t.Fatal()
	}
}

func TestRootModel_quit(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	_ = out.(*rootModel)
}

func TestRootModel_ctrlC(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal()
	}
}

func TestRootModel_tabCycle(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if out.(*rootModel).tab != tabStreams {
		t.Fatal(out.(*rootModel).tab)
	}
	out2, _ := out.(*rootModel).Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if out2.(*rootModel).tab != tabSearch {
		t.Fatal(out2.(*rootModel).tab)
	}
}

func TestRootModel_numberTabs(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if out.(*rootModel).tab != tabAddons {
		t.Fatal()
	}
}

func TestRootModel_searchDoneMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	items := []search.TitleResult{{Title: "A", IMDBID: "tt1", Type: "movie"}}
	out, _ := m.Update(searchDoneMsg{items: items, source: browseFeatured})
	rm := out.(*rootModel)
	if len(rm.searchItems) != 1 {
		t.Fatal(len(rm.searchItems))
	}
}

func TestRootModel_searchErrMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(searchErrMsg{err: errors.New("boom")})
	if out.(*rootModel).toast == "" {
		t.Fatal()
	}
}

func TestRootModel_metaDoneMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Title: "S", IMDBID: "tt1", Type: "series"}
	meta := &stremio.Meta{Videos: []stremio.Video{{Season: 1, Episode: 1, Title: "E1"}}}
	out, _ := m.Update(metaDoneMsg{meta: meta})
	rm := out.(*rootModel)
	if rm.seriesMeta == nil {
		t.Fatal()
	}
}

func TestRootModel_metaDoneMsg_noSelection(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(metaDoneMsg{meta: &stremio.Meta{}})
	if out.(*rootModel).seriesMeta != nil {
		t.Fatal()
	}
}

func TestRootModel_streamsDoneMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Title: "M", IMDBID: "tt9", Type: "movie"}
	rows := []streams.ResolvedStream{{AddonID: "a", AddonName: "A", Stream: stremio.Stream{URL: "https://u"}}}
	out, _ := m.Update(streamsDoneMsg(rows))
	rm := out.(*rootModel)
	if len(rm.allResolvedStreams) != 1 {
		t.Fatal(len(rm.allResolvedStreams))
	}
}

func TestRootModel_streamsErrMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Title: "M", IMDBID: "tt9", Type: "movie"}
	out, _ := m.Update(streamsErrMsg{err: errors.New("e")})
	if out.(*rootModel).toast == "" {
		t.Fatal()
	}
}

func TestRootModel_statusMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(statusMsg{text: "ok"})
	if out.(*rootModel).toast != "ok" {
		t.Fatal(out.(*rootModel).toast)
	}
}

func TestRootModel_statusMsg_err(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	out, _ := m.Update(statusMsg{err: errors.New("bad")})
	if out.(*rootModel).toast == "" {
		t.Fatal()
	}
}

func TestRootModel_toastClearMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.toast = "x"
	out, _ := m.Update(toastClearMsg{})
	if out.(*rootModel).toast != "" {
		t.Fatal()
	}
}

func TestRootModel_metaErrMsg(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Type: "series"}
	out, _ := m.Update(metaErrMsg{err: errors.New("meta")})
	if out.(*rootModel).toast == "" {
		t.Fatal()
	}
}

func TestRootModel_esc_addonURLMode(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabAddons
	m.addonURLMode = true
	m.addonURL.Focus()
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	rm := out.(*rootModel)
	if rm.addonURLMode {
		t.Fatal()
	}
}

func TestRootModel_esc_searchActive(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	m.searchActive = true
	m.searchInput.Focus()
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	rm := out.(*rootModel)
	if rm.searchActive {
		t.Fatal()
	}
}
