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
		"Tabs: 1–4 jump · Tab / Shift+Tab cycle (disabled while search or manifest URL field is focused)",
		"",
		"Search: Enter run search · ↓ move to results · Enter open streams · ctrl+t or t toggle movie/series",
		"        ctrl+p Cinemeta popular · ctrl+i Cinemeta featured (current type)",
		"",
		"Streams: Enter copy URL · esc or b back · r reload · o cycle sort (rank · rank-asc · addon · title) · [ ] or h l addon filter (when multiple addons)",
		"",
		"Addons: a add manifest · d remove · c configure in browser",
		"",
		"Settings: Enter save TMDB key",
		"",
		"Quit: esc or ctrl+c once to confirm, then y or ctrl+c again · n or esc cancels",
		"",
		"? or F1 toggles this help · esc or q closes",
	}, "\n")
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(max(40, m.width-8)).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, dim.Render(text)))
	return panel
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
			typeLine := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
				fmt.Sprintf("Type: %s · ctrl+t / t toggle · ctrl+p popular · ctrl+i featured (Cinemeta)", m.searchMediaType))
			body = lipgloss.JoinVertical(lipgloss.Left,
				m.searchInput.View(),
				typeLine,
				"",
				m.searchList.View(),
			)
		case tabStreams:
			head := "Pick a stream, Enter copies URL. esc or b → back to search."
			if m.selected != nil {
				head = fmt.Sprintf("%s (%s) — %s", m.selected.Title, m.selected.IMDBID, m.selected.Type)
			}
			if m.streamsBusy {
				head += "\nLoading…"
			}
			sortHint := ""
			if !m.streamsBusy && len(m.allResolvedStreams) > 0 {
				sortHint = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
					fmt.Sprintf("Sort: %s · o cycle (rank · rank-asc · addon · title)", m.streamListOrder))
			}
			addonStrip := ""
			if len(m.streamAddonTabs) > 1 {
				addonStrip = m.renderStreamsAddonTabs() + "\n"
			}
			streamSections := []string{head}
			if sortHint != "" {
				streamSections = append(streamSections, sortHint)
			}
			if addonStrip != "" {
				streamSections = append(streamSections, addonStrip)
			}
			streamSections = append(streamSections, m.streamsList.View())
			body = lipgloss.JoinVertical(lipgloss.Left, streamSections...)
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

	if m.quitConfirm && !m.helpOpen {
		banner := lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true).
			Render("Quit?  y  confirm  ·  n / esc  cancel  ·  ctrl+c  confirm quit")
		body = lipgloss.JoinVertical(lipgloss.Left, banner, "", body)
	}

	toast := ""
	if m.toast != "" {
		toast = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(m.toast)
	}
	helpStr := "? help · tab / 1–4 tabs · esc or ctrl+c to quit (confirm)"
	if m.tab == tabStreams && !m.helpOpen {
		helpStr = "? help · tab / 1–4 tabs · esc/b back · ctrl+c quit (confirm) · o sort"
		if len(m.streamAddonTabs) > 1 {
			helpStr += " · [ ] / h l addon filter"
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

func (m *rootModel) renderTabs() string {
	names := []string{"Search", "Streams", "Addons", "Settings"}
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
