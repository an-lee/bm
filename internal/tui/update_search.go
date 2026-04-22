package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *rootModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchActive && m.searchInput.Focused() {
		switch msg.String() {
		case "enter":
			q := strings.TrimSpace(m.searchInput.Value())
			if q == "" {
				m.closeSearchInput()
				return m, nil
			}
			m.searchActive = false
			m.searchInput.Blur()
			return m, m.runSearch(q)
		case "down":
			m.searchInput.Blur()
			return m, nil
		case "esc":
			m.closeSearchInput()
			return m, nil
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "/":
		if !m.searchActive {
			m.searchActive = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, textinput.Blink
		}
	case "t":
		return m.toggleSearchType()
	case "p":
		m.closeSearchInput()
		m.lastSearchQuery = ""
		return m, m.runCinemetaPopular()
	case "f":
		m.closeSearchInput()
		m.lastSearchQuery = ""
		return m, m.runCinemetaFeatured()
	case "up", "down", "k", "j":
		var cmd tea.Cmd
		m.searchList, cmd = m.searchList.Update(msg)
		return m, cmd
	case "enter":
		if it, ok := m.searchList.SelectedItem().(titleItem); ok {
			sel := it.r
			m.selected = &sel
			m.tab = tabStreams
			return m, tea.Batch(m.enterStreamsForSelection(), m.onTabChange())
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.searchList, cmd = m.searchList.Update(msg)
	return m, cmd
}

func (m *rootModel) enterStreamsForSelection() tea.Cmd {
	if m.selected == nil {
		return nil
	}
	if m.effectiveMetaType() == "series" {
		m.streamsStage = stageEpisodes
		m.seriesMeta = nil
		m.episodesList.SetItems(nil)
		return m.loadSeriesMeta()
	}
	m.streamsStage = stageStreams
	return m.loadStreamsForEpisode(0, 0)
}

func (m *rootModel) runSearch(q string) tea.Cmd {
	q = strings.TrimSpace(q)
	m.lastSearchQuery = q
	return func() tea.Msg {
		ctx := context.Background()
		res, err := m.app.Search.Search(ctx, q, m.searchMediaType, 0)
		if err != nil {
			return searchErrMsg{err}
		}
		return searchDoneMsg{items: res, source: browseSearch}
	}
}

func (m *rootModel) runCinemetaPopular() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		res, err := m.app.Search.CinemetaPopular(ctx, m.searchMediaType, 0)
		if err != nil {
			return searchErrMsg{err}
		}
		return searchDoneMsg{items: res, source: browsePopular}
	}
}

func (m *rootModel) runCinemetaFeatured() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		res, err := m.app.Search.CinemetaFeatured(ctx, m.searchMediaType, 0)
		if err != nil {
			return searchErrMsg{err}
		}
		return searchDoneMsg{items: res, source: browseFeatured}
	}
}

func (m *rootModel) toggleSearchType() (tea.Model, tea.Cmd) {
	if m.searchMediaType == "movie" {
		m.searchMediaType = "series"
	} else {
		m.searchMediaType = "movie"
	}
	if m.searchActive {
		q := strings.TrimSpace(m.searchInput.Value())
		if q != "" {
			return m, m.runSearch(q)
		}
		m.toast = "Type: " + m.searchMediaType
		return m, m.tickToast()
	}
	return m, m.refreshBrowseAfterTypeChange()
}

func (m *rootModel) refreshBrowseAfterTypeChange() tea.Cmd {
	switch m.browseMode {
	case browseFeatured:
		return m.runCinemetaFeatured()
	case browseSearch:
		if strings.TrimSpace(m.lastSearchQuery) != "" {
			return m.runSearch(m.lastSearchQuery)
		}
		fallthrough
	default:
		return m.runCinemetaPopular()
	}
}

func normalizeSearchMediaType(s string) string {
	if strings.TrimSpace(s) == "series" {
		return "series"
	}
	return "movie"
}
