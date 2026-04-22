package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"bm/internal/config"
)

func (m *rootModel) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		key := strings.TrimSpace(m.settingsInput.Value())
		if key != "" {
			_ = config.SetKey("tmdb.api_key", key)
			m.settingsInput.SetValue("")
			m.toast = "TMDB key saved"
			_ = m.app.Reload()
			return m, m.tickToast()
		}
	}
	var cmd tea.Cmd
	m.settingsInput, cmd = m.settingsInput.Update(msg)
	return m, cmd
}
