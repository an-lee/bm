package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *rootModel) renderHelpPanel() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Keyboard shortcuts")
	text := strings.Join([]string{
		"",
		"Tabs: 1–4 jump · Tab / Shift+Tab cycle (blocked while search or manifest URL field is focused)",
		"",
		"Browse: / search · Enter run search when search is open · Esc closes search",
		"        t movie/series · p popular · f featured (Cinemeta) · Enter opens Streams",
		"",
		"Streams: series → pick episode then streams · Enter copies URL · Esc or b back",
		"         r reload · o sort · h / l addon filter (when multiple addons)",
		"",
		"Addons: a add manifest · d remove · c configure in browser",
		"",
		"Settings: Enter save TMDB key",
		"",
		"Quit: q or ctrl+c · ? or F1 toggles this help · Esc or q closes help",
	}, "\n")
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(max(40, m.width-8)).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, dim.Render(text)))
	return panel
}

func (m *rootModel) renderStreamsBreadcrumb() string {
	if m.selected == nil {
		return ""
	}
	parts := []string{
		fmt.Sprintf("%s (%s)", m.selected.Title, m.selected.IMDBID),
	}
	if m.effectiveMetaType() == "series" {
		switch m.streamsStage {
		case stageEpisodes:
			parts = append(parts, "Episodes")
		case stageStreams:
			if m.seasonPick > 0 || m.episodePick > 0 {
				parts = append(parts, fmt.Sprintf("S%02d", m.seasonPick), fmt.Sprintf("E%02d", m.episodePick))
			}
		}
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Join(parts, " › "))
}

func (m *rootModel) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	tabBar := m.renderTabs()
	var body string
	if !m.helpOpen {
		switch m.tab {
		case tabSearch:
			searchLine := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
				"Press / to search · t type · p popular · f featured (Cinemeta)")
			if m.searchActive {
				searchLine = m.searchInput.View()
			}
			typeLine := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
				fmt.Sprintf("Type: %s", m.searchMediaType))
			body = lipgloss.JoinVertical(lipgloss.Left,
				searchLine,
				typeLine,
				"",
				m.searchList.View(),
			)
		case tabStreams:
			body = m.renderStreamsBody()
		case tabAddons:
			extra := ""
			if m.addonURLMode {
				extra = lipgloss.JoinVertical(lipgloss.Left,
					"",
					"Manifest URL:",
					m.addonURL.View(),
					"(Enter to install, esc cancel)",
				)
			}
			body = lipgloss.JoinVertical(lipgloss.Left,
				"[a] add  [d] remove selected  [c] configure in browser",
				m.addonList.View(),
				extra,
			)
		case tabSettings:
			body = lipgloss.JoinVertical(lipgloss.Left,
				"TMDB API key (optional, improves search):",
				m.settingsInput.View(),
				"",
				"Enter to save. Keys are stored in config.toml.",
			)
		}
	} else {
		body = m.renderHelpPanel()
	}

	toast := ""
	if m.toast != "" {
		toast = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(m.toast)
	}
	helpStr := "? help · tab / 1–4 · q quit"
	if m.tab == tabSearch && !m.helpOpen {
		helpStr = "? help · tab / 1–4 · / search · t p f · q quit"
	}
	if m.tab == tabStreams && !m.helpOpen {
		helpStr = "? help · tab / 1–4 · esc/b back · r reload · o sort · q quit"
		if m.selected != nil && m.streamsStage == stageStreams && len(m.streamAddonTabs) > 1 {
			helpStr += " · h/l addons"
		}
	}
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(helpStr)

	frame := lipgloss.NewStyle().
		MaxWidth(m.width).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			tabBar,
			"",
			body,
			"",
			toast,
			help,
		))
	return frame
}

func (m *rootModel) renderStreamsBody() string {
	if m.selected == nil {
		return "No title selected. Open Browse (1), pick a title, press Enter."
	}
	crumb := m.renderStreamsBreadcrumb()
	if crumb != "" {
		crumb = crumb + "\n\n"
	}

	if m.effectiveMetaType() == "series" && m.streamsStage == stageEpisodes {
		head := "Pick an episode · Enter to load streams · esc or b back to Browse."
		if m.episodesBusy {
			head += "\nLoading episodes…"
		}
		return lipgloss.JoinVertical(lipgloss.Left, crumb+head, "", m.episodesList.View())
	}

	head := "Pick a stream · Enter copies URL · esc or b back."
	if m.streamsBusy {
		head += "\nLoading streams…"
	}
	sortHint := ""
	if !m.streamsBusy && len(m.allResolvedStreams) > 0 {
		sortHint = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
			fmt.Sprintf("Sort: %s · o cycle", m.streamListOrder))
	}
	addonStrip := ""
	if len(m.streamAddonTabs) > 1 {
		addonStrip = m.renderStreamsAddonTabs() + "\n"
	}
	sections := []string{crumb + head}
	if sortHint != "" {
		sections = append(sections, sortHint)
	}
	if addonStrip != "" {
		sections = append(sections, addonStrip)
	}
	sections = append(sections, m.streamsList.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *rootModel) renderTabs() string {
	names := []string{"Browse", "Streams", "Addons", "Settings"}
	var parts []string
	for i, n := range names {
		st := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("252"))
		if i == m.tab {
			st = st.Foreground(lipgloss.Color("205")).Bold(true)
		}
		parts = append(parts, st.Render(fmt.Sprintf("%d:%s", i+1, n)))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
