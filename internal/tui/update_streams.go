package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"bm/internal/clipboard"
	"bm/internal/streams"
)

func (m *rootModel) backFromStreams() (tea.Model, tea.Cmd) {
	m.tab = tabSearch
	m.selected = nil
	m.streamsBusy = false
	m.allResolvedStreams = nil
	m.streamAddonTabs = nil
	m.streamsAddonTabIdx = 0
	m.streamsList.SetItems(nil)
	m.streamsList.Title = "Streams"
	return m, m.onTabChange()
}

func (m *rootModel) loadStreamsForSelection() tea.Cmd {
	if m.selected == nil {
		return nil
	}
	sel := *m.selected
	m.streamsBusy = true
	m.allResolvedStreams = nil
	m.streamAddonTabs = nil
	m.streamsAddonTabIdx = 0
	m.streamsList.SetItems(nil)
	m.streamsList.Title = "Streams"
	imdb := sel.IMDBID
	metaType := sel.Type
	if metaType == "" {
		metaType = m.app.Config.UI.DefaultType
	}
	season, episode := 0, 0
	if metaType == "series" {
		season, episode = 1, 1
	}
	return func() tea.Msg {
		ctx := context.Background()
		list, err := m.app.Streams.Resolve(ctx, imdb, metaType, season, episode)
		if err != nil {
			return streamsErrMsg{err}
		}
		return streamsDoneMsg(list)
	}
}

func (m *rootModel) cycleStreamSortOrder() (tea.Model, tea.Cmd) {
	if m.streamsBusy || len(m.allResolvedStreams) == 0 {
		return m, nil
	}
	m.streamListOrder = streams.NextStreamOrder(m.streamListOrder)
	streams.ApplySort(m.allResolvedStreams, m.streamListOrder)
	n := m.applyStreamsAddonFilter()
	tabLabel := "All"
	if len(m.streamAddonTabs) > 0 && m.streamsAddonTabIdx < len(m.streamAddonTabs) {
		tabLabel = m.streamAddonTabs[m.streamsAddonTabIdx].label
	}
	m.toast = fmt.Sprintf("sort: %s · %s · %d streams", m.streamListOrder, tabLabel, n)
	return m, m.tickToast()
}

func (m *rootModel) cycleStreamsAddon(delta int) (tea.Model, tea.Cmd) {
	if len(m.streamAddonTabs) <= 1 {
		return m, nil
	}
	n := len(m.streamAddonTabs)
	m.streamsAddonTabIdx = (n + m.streamsAddonTabIdx + delta) % n
	cnt := m.applyStreamsAddonFilter()
	tab := m.streamAddonTabs[m.streamsAddonTabIdx]
	m.toast = fmt.Sprintf("%s · %d streams", tab.label, cnt)
	return m, m.tickToast()
}

func (m *rootModel) updateStreams(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b":
		return m.backFromStreams()
	case "r":
		if m.selected != nil {
			return m, m.loadStreamsForSelection()
		}
	case "o":
		return m.cycleStreamSortOrder()
	case "[", "h":
		return m.cycleStreamsAddon(-1)
	case "]", "l":
		return m.cycleStreamsAddon(1)
	case "enter":
		if it, ok := m.streamsList.SelectedItem().(streamItem); ok {
			u := it.s.PlayableURL()
			if u != "" {
				_ = clipboard.WriteAll(u)
				m.toast = "Copied to clipboard"
				return m, m.tickToast()
			}
			m.toast = "No URL for this stream"
			return m, m.tickToast()
		}
	}
	var cmd tea.Cmd
	m.streamsList, cmd = m.streamsList.Update(msg)
	return m, cmd
}

func buildStreamAddonTabs(rows []streams.ResolvedStream) []streamAddonTab {
	byID := make(map[string]string)
	for _, r := range rows {
		id := strings.TrimSpace(r.AddonID)
		if id == "" {
			continue
		}
		name := strings.TrimSpace(r.AddonName)
		if name == "" {
			name = id
		}
		byID[id] = name
	}
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	tabs := []streamAddonTab{{label: "All", addonID: ""}}
	for _, id := range ids {
		tabs = append(tabs, streamAddonTab{label: byID[id], addonID: id})
	}
	return tabs
}

func (m *rootModel) applyStreamsAddonFilter() int {
	if len(m.streamAddonTabs) == 0 {
		m.streamsList.SetItems(nil)
		m.streamsList.Title = "Streams"
		return 0
	}
	if m.streamsAddonTabIdx < 0 || m.streamsAddonTabIdx >= len(m.streamAddonTabs) {
		m.streamsAddonTabIdx = 0
	}
	tab := m.streamAddonTabs[m.streamsAddonTabIdx]
	var filtered []streams.ResolvedStream
	if tab.addonID == "" {
		filtered = m.allResolvedStreams
	} else {
		for _, s := range m.allResolvedStreams {
			if s.AddonID == tab.addonID {
				filtered = append(filtered, s)
			}
		}
	}
	items := make([]list.Item, 0, len(filtered))
	for _, s := range filtered {
		items = append(items, streamItem{s: s})
	}
	m.streamsList.SetItems(items)
	m.streamsList.Title = "Streams · " + tab.label
	return len(filtered)
}

func (m *rootModel) renderStreamsAddonTabs() string {
	var cells []string
	for i, tab := range m.streamAddonTabs {
		if i > 0 {
			cells = append(cells, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" | "))
		}
		st := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("252"))
		if i == m.streamsAddonTabIdx {
			st = st.Foreground(lipgloss.Color("205")).Bold(true)
		}
		cells = append(cells, st.Render(tab.label))
	}
	prefix := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Addons: ")
	line := prefix + lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	if len(m.streamAddonTabs) > 1 {
		hint := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  ·  [ ] or h l  filter")
		line += hint
	}
	return line
}
