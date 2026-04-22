package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *rootModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchInput.Focused() {
		switch msg.String() {
		case "ctrl+t":
			return m.toggleSearchType()
		case "ctrl+p":
			return m, m.runCinemetaPopular()
		case "ctrl+i":
			return m, m.runCinemetaFeatured()
		case "enter":
			q := strings.TrimSpace(m.searchInput.Value())
			if q == "" {
				return m, nil
			}
			return m, m.runSearch(q)
		case "down":
			m.searchInput.Blur()
		default:
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "ctrl+t", "t":
		return m.toggleSearchType()
	case "ctrl+p":
		return m, m.runCinemetaPopular()
	case "ctrl+i":
		return m, m.runCinemetaFeatured()
	case "up", "down", "k", "j":
		var cmd tea.Cmd
		m.searchList, cmd = m.searchList.Update(msg)
		return m, cmd
	case "enter":
		if it, ok := m.searchList.SelectedItem().(titleItem); ok {
			m.selected = &it.r
			m.tab = tabStreams
			return m, tea.Batch(m.loadStreamsForSelection(), m.onTabChange())
		}
		return m, nil
	default:
		m.searchInput.Focus()
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
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
		return searchDoneMsg(res)
	}
}

func (m *rootModel) runCinemetaPopular() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		res, err := m.app.Search.CinemetaPopular(ctx, m.searchMediaType, 0)
		if err != nil {
			return searchErrMsg{err}
		}
		return searchDoneMsg(res)
	}
}

func (m *rootModel) runCinemetaFeatured() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		res, err := m.app.Search.CinemetaFeatured(ctx, m.searchMediaType, 0)
		if err != nil {
			return searchErrMsg{err}
		}
		return searchDoneMsg(res)
	}
}

func (m *rootModel) toggleSearchType() (tea.Model, tea.Cmd) {
	if m.searchMediaType == "movie" {
		m.searchMediaType = "series"
	} else {
		m.searchMediaType = "movie"
	}
	if m.lastSearchQuery == "" {
		m.toast = "Type: " + m.searchMediaType
		return m, m.tickToast()
	}
	return m, m.runSearch(m.lastSearchQuery)
}

func normalizeSearchMediaType(s string) string {
	if strings.TrimSpace(s) == "series" {
		return "series"
	}
	return "movie"
}
