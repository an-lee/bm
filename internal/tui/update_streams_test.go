package tui

import (
	"testing"

	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
	tea "github.com/charmbracelet/bubbletea"
)

func TestBuildStreamAddonTabs(t *testing.T) {
	t.Parallel()
	tabs := buildStreamAddonTabs([]streams.ResolvedStream{
		{AddonID: "b", AddonName: "B"},
		{AddonID: "a", AddonName: "A"},
		{AddonID: "", AddonName: "skip"},
	})
	if len(tabs) != 3 || tabs[0].label != "All" {
		t.Fatalf("%#v", tabs)
	}
}

func TestRootModel_applyStreamsAddonFilter(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Title: "T", IMDBID: "tt1", Type: "movie"}
	m.allResolvedStreams = []streams.ResolvedStream{
		{AddonID: "a", AddonName: "A", Stream: stremio.Stream{URL: "https://a"}},
		{AddonID: "b", AddonName: "B", Stream: stremio.Stream{URL: "https://b"}},
	}
	m.streamAddonTabs = buildStreamAddonTabs(m.allResolvedStreams)
	m.streamsAddonTabIdx = 1
	n := m.applyStreamsAddonFilter()
	if n != 1 {
		t.Fatal(n)
	}
	m.streamsAddonTabIdx = 0
	n2 := m.applyStreamsAddonFilter()
	if n2 != 2 {
		t.Fatal(n2)
	}
}

func TestRootModel_clearStreamsListState(t *testing.T) {
	m := newRootModel(testApp(t))
	m.allResolvedStreams = []streams.ResolvedStream{{}}
	m.streamAddonTabs = []streamAddonTab{{label: "All"}}
	m.clearStreamsListState()
	if len(m.allResolvedStreams) != 0 || m.streamAddonTabs != nil {
		t.Fatal()
	}
}

func TestRootModel_streamsListTitle(t *testing.T) {
	m := newRootModel(testApp(t))
	m.selected = &search.TitleResult{Title: "T", IMDBID: "tt1", Type: "movie"}
	if m.streamsListTitle("All") != "Streams · All" {
		t.Fatal(m.streamsListTitle("All"))
	}
	m.selected.Type = "series"
	m.streamsStage = stageStreams
	m.seasonPick, m.episodePick = 1, 2
	got := m.streamsListTitle("Addon")
	if len(got) < 8 {
		t.Fatal(got)
	}
}

func TestRootModel_backToBrowse(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "movie"}
	out, _ := m.backToBrowse()
	rm := out.(*rootModel)
	if rm.tab != tabSearch || rm.selected != nil {
		t.Fatalf("tab %d sel %v", rm.tab, rm.selected)
	}
}

func TestRootModel_backFromStreams_nilSelection(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = nil
	out, _ := m.backFromStreams()
	if out.(*rootModel).tab != tabSearch {
		t.Fatal(out.(*rootModel).tab)
	}
}

func TestRootModel_cycleStreamSortOrder_viaUpdate(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "movie"}
	m.streamsBusy = false
	m.allResolvedStreams = []streams.ResolvedStream{{Stream: stremio.Stream{URL: "https://x"}}}
	m.streamAddonTabs = buildStreamAddonTabs(m.allResolvedStreams)
	_ = m.applyStreamsAddonFilter()
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	if cmd == nil {
		t.Fatal("expected toast cmd")
	}
	if out.(*rootModel).streamListOrder == "" {
		t.Fatal()
	}
}

func TestRootModel_cycleStreamsAddon(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "movie"}
	m.allResolvedStreams = []streams.ResolvedStream{
		{AddonID: "a", Stream: stremio.Stream{URL: "https://a"}},
		{AddonID: "b", Stream: stremio.Stream{URL: "https://b"}},
	}
	m.streamAddonTabs = buildStreamAddonTabs(m.allResolvedStreams)
	_ = m.applyStreamsAddonFilter()
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd == nil {
		t.Fatal()
	}
	_ = out
}
