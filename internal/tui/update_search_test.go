package tui

import (
	"testing"

	"bm/internal/search"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNormalizeSearchMediaType(t *testing.T) {
	t.Parallel()
	if normalizeSearchMediaType("series") != "series" {
		t.Fatal()
	}
	if normalizeSearchMediaType("other") != "movie" {
		t.Fatal()
	}
}

func TestUpdateSearch_slashActivates(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	rm := out.(*rootModel)
	if !rm.searchActive || cmd == nil {
		t.Fatal(rm.searchActive, cmd)
	}
}

func TestUpdateSearch_toggleType(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	rm := out.(*rootModel)
	if rm.searchMediaType != "series" {
		t.Fatal(rm.searchMediaType)
	}
}

func TestUpdateSearch_enterSelectsTitle(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	m.searchList.SetItems(m.searchItems)
	// inject item via searchDoneMsg first
	items := []search.TitleResult{{Title: "X", IMDBID: "tt1", Type: "movie"}}
	m2, _ := m.Update(searchDoneMsg{items: items, source: browseSearch})
	rm := m2.(*rootModel)
	if len(rm.searchItems) == 0 {
		t.Fatal()
	}
	out, cmd := rm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if out.(*rootModel).tab != tabStreams || cmd == nil {
		t.Fatalf("tab %d cmd %v", out.(*rootModel).tab, cmd)
	}
}
