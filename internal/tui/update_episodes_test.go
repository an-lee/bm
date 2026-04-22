package tui

import (
	"testing"

	"bm/internal/search"
	"bm/internal/stremio"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAdvanceEpisodeCursor(t *testing.T) {
	m := newRootModel(testApp(t))
	items := buildEpisodeListItems([]stremio.Video{
		{Season: 1, Episode: 1, Title: "a"},
		{Season: 1, Episode: 2, Title: "b"},
	})
	m.episodesList.SetItems(items)
	m.episodesList.Select(0)
	m.advanceEpisodeCursor(1)
}

func TestPickEpisodeOrSkipHeader_episode(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Type: "series", IMDBID: "tt1"}
	items := buildEpisodeListItems([]stremio.Video{{Season: 1, Episode: 5, Title: "E"}})
	m.episodesList.SetItems(items)
	m.episodesList.Select(len(items) - 1)
	_, cmd := m.pickEpisodeOrSkipHeader()
	if cmd == nil {
		t.Fatal("expected load streams cmd")
	}
}

func TestPickEpisodeOrSkipHeader_seasonHeader(t *testing.T) {
	m := newRootModel(testApp(t))
	items := buildEpisodeListItems([]stremio.Video{
		{Season: 1, Episode: 1, Title: "a"},
		{Season: 1, Episode: 2, Title: "b"},
	})
	m.episodesList.SetItems(items)
	m.episodesList.Select(0)
	_, cmd := m.pickEpisodeOrSkipHeader()
	if cmd != nil {
		t.Fatalf("expected nil cmd for header, got %v", cmd)
	}
}

func TestUpdateEpisodes_keyB(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "series"}
	m.streamsStage = stageEpisodes
	out, _ := m.updateEpisodes(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if out.(*rootModel).tab != tabSearch {
		t.Fatal(out.(*rootModel).tab)
	}
}
