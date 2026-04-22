package tui

import (
	"errors"
	"strings"
	"testing"

	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestView_addonsURLMode(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 90
	m.height = 28
	m.tab = tabAddons
	m.addonURLMode = true
	m.addonURL.Focus()
	s := m.View()
	if !strings.Contains(s, "Manifest") {
		t.Fatal(s[:min(200, len(s))])
	}
}

func TestView_settingsTab(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 90
	m.height = 28
	m.tab = tabSettings
	s := m.View()
	if !strings.Contains(s, "TMDB") {
		t.Fatal(s[:min(200, len(s))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestUpdate_helpEsc(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.helpOpen = true
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if out.(*rootModel).helpOpen {
		t.Fatal()
	}
}

func TestUpdate_helpCtrlC(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.helpOpen = true
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if out.(*rootModel).helpOpen {
		t.Fatal()
	}
}

func TestUpdate_streamsDone_reSortExisting(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Type: "movie"}
	m.allResolvedStreams = []streams.ResolvedStream{{Stream: stremio.Stream{URL: "https://z"}}}
	m.streamListOrder = "title"
	out, _ := m.Update(statusMsg{text: "saved"})
	rm := out.(*rootModel)
	if len(rm.allResolvedStreams) != 1 {
		t.Fatal()
	}
}

func TestUpdate_streamsErr_nilSelection(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = nil
	_, _ = m.Update(streamsErrMsg{err: errors.New("x")})
}

func TestUpdate_metaDone_emptyEpisodes(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = &search.TitleResult{Type: "series"}
	out, _ := m.Update(metaDoneMsg{meta: &stremio.Meta{Videos: nil}})
	if out.(*rootModel).toast == "" {
		t.Fatal()
	}
}

func TestUpdate_metaErr_nilSelection(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = nil
	_, _ = m.Update(metaErrMsg{err: errors.New("x")})
}

func TestUpdate_streamsDone_nilSelection(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.selected = nil
	_, _ = m.Update(streamsDoneMsg(nil))
}

func TestUpdate_streams_enterNoPlayableURL(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "movie"}
	m.streamsStage = stageStreams
	m.streamsList.SetItems([]list.Item{
		streamItem{s: streams.ResolvedStream{Stream: stremio.Stream{Name: "n"}}},
	})
	m.streamsList.Select(0)
	out, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || out.(*rootModel).toast == "" {
		t.Fatal(cmd, out.(*rootModel).toast)
	}
}

func TestUpdate_streams_backFromSeriesStreams(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabStreams
	m.selected = &search.TitleResult{Type: "series", IMDBID: "tt1"}
	m.streamsStage = stageStreams
	out, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	rm := out.(*rootModel)
	if rm.streamsStage != stageEpisodes {
		t.Fatalf("stage %d", rm.streamsStage)
	}
}
