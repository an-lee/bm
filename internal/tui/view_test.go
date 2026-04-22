package tui

import (
	"strings"
	"testing"

	"bm/internal/search"
)

func TestView_loadingWidthZero(t *testing.T) {
	m := newRootModel(testApp(t))
	out := m.View()
	if out != "Loading…" {
		t.Fatalf("%q", out)
	}
}

func TestView_searchTab(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.height = 24
	s := m.View()
	if !strings.Contains(s, "Browse") {
		t.Fatal(s)
	}
}

func TestView_streamsTab(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.height = 24
	m.tab = tabStreams
	s := m.View()
	if !strings.Contains(s, "Streams") {
		t.Fatal(s)
	}
}

func TestView_helpOpen(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.height = 24
	m.helpOpen = true
	s := m.View()
	if !strings.Contains(s, "Keyboard") {
		t.Fatal(s)
	}
}

func TestRenderTabs(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSettings
	s := m.renderTabs()
	if !strings.Contains(s, "Settings") {
		t.Fatal(s)
	}
}

func TestRenderStreamsBreadcrumb_nil(t *testing.T) {
	m := newRootModel(testApp(t))
	if m.renderStreamsBreadcrumb() != "" {
		t.Fatal()
	}
}

func TestRenderStreamsBreadcrumb_movie(t *testing.T) {
	m := newRootModel(testApp(t))
	m.selected = &search.TitleResult{Title: "M", IMDBID: "tt1", Type: "movie"}
	s := m.renderStreamsBreadcrumb()
	if !strings.Contains(s, "tt1") {
		t.Fatal(s)
	}
}

func TestRenderStreamsBreadcrumb_series(t *testing.T) {
	m := newRootModel(testApp(t))
	m.selected = &search.TitleResult{Title: "S", IMDBID: "tt2", Type: "series"}
	m.streamsStage = stageEpisodes
	s1 := m.renderStreamsBreadcrumb()
	if !strings.Contains(s1, "Episodes") {
		t.Fatal(s1)
	}
	m.streamsStage = stageStreams
	m.seasonPick, m.episodePick = 2, 3
	s2 := m.renderStreamsBreadcrumb()
	if !strings.Contains(s2, "S02E03") {
		t.Fatal(s2)
	}
}

func TestRenderStreamsBody_noSelection(t *testing.T) {
	m := newRootModel(testApp(t))
	m.selected = nil
	s := m.renderStreamsBody()
	if !strings.Contains(s, "No title selected") {
		t.Fatal(s)
	}
}

func TestRenderStreamsBody_streamsBusy(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Type: "movie"}
	m.streamsBusy = true
	s := m.renderStreamsBody()
	if !strings.Contains(s, "Loading streams") {
		t.Fatal(s)
	}
}

func TestRenderHelpPanel(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	s := m.renderHelpPanel()
	if !strings.Contains(s, "shortcuts") {
		t.Fatal(s)
	}
}

func TestRenderStreamsAddonTabs(t *testing.T) {
	m := newRootModel(testApp(t))
	m.streamAddonTabs = []streamAddonTab{{label: "All"}, {label: "A", addonID: "a"}}
	m.streamsAddonTabIdx = 1
	s := m.renderStreamsAddonTabs()
	if !strings.Contains(s, "A") {
		t.Fatal(s)
	}
}

func TestView_streamsWithAddonTabsHint(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 100
	m.height = 30
	m.tab = tabStreams
	m.selected = &search.TitleResult{Title: "T", IMDBID: "tt1", Type: "movie"}
	m.streamsStage = stageStreams
	m.streamAddonTabs = []streamAddonTab{{label: "All"}, {label: "X", addonID: "x"}}
	m.streamsAddonTabIdx = 0
	_ = m.applyStreamsAddonFilter()
	s := m.View()
	if !strings.Contains(s, "h/l") {
		t.Fatal(s)
	}
}
