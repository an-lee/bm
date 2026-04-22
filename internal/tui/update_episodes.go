package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *rootModel) loadSeriesMeta() tea.Cmd {
	if m.selected == nil {
		return nil
	}
	imdb := m.selected.IMDBID
	m.episodesBusy = true
	return func() tea.Msg {
		ctx := context.Background()
		meta, err := m.app.Meta(ctx, imdb, "series")
		if err != nil {
			return metaErrMsg{err}
		}
		return metaDoneMsg{meta: meta}
	}
}

func (m *rootModel) updateEpisodes(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b":
		return m.backToBrowse()
	case "enter":
		if m.episodesBusy {
			return m, nil
		}
		return m.pickEpisodeOrSkipHeader()
	default:
		var cmd tea.Cmd
		m.episodesList, cmd = m.episodesList.Update(msg)
		return m, cmd
	}
}

func (m *rootModel) pickEpisodeOrSkipHeader() (tea.Model, tea.Cmd) {
	it := m.episodesList.SelectedItem()
	switch it := it.(type) {
	case episodeItem:
		m.seasonPick, m.episodePick = episodeSE(it.v)
		m.streamsStage = stageStreams
		return m, m.loadStreamsForEpisode(m.seasonPick, m.episodePick)
	case seasonHeaderItem:
		m.advanceEpisodeCursor(1)
		return m, nil
	default:
		return m, nil
	}
}

// advanceEpisodeCursor moves the cursor to the next (delta=+1) or previous episode row, skipping headers.
func (m *rootModel) advanceEpisodeCursor(delta int) {
	items := m.episodesList.Items()
	if len(items) == 0 {
		return
	}
	idx := m.episodesList.Index()
	for step := 0; step < len(items); step++ {
		idx += delta
		for idx < 0 {
			idx += len(items)
		}
		idx %= len(items)
		if _, ok := items[idx].(episodeItem); ok {
			m.episodesList.Select(idx)
			return
		}
	}
}
