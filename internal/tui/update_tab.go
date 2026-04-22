package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *rootModel) onTabChange() tea.Cmd {
	m.quitConfirm = false

	if m.tab != tabSearch {
		m.searchInput.Blur()
	}
	if m.tab != tabSettings {
		m.settingsInput.Blur()
	}
	if m.tab != tabAddons {
		m.addonURLMode = false
		m.addonURL.Blur()
	}

	if m.tab == tabSearch {
		return textinput.Blink
	}
	if m.tab == tabSettings {
		m.settingsInput.Focus()
		return textinput.Blink
	}
	return nil
}

func (m *rootModel) refreshAddonList() {
	items := make([]list.Item, 0, len(m.app.Config.Addons))
	for _, a := range m.app.Config.Addons {
		st := "off"
		if a.Enabled {
			st = "on"
		}
		items = append(items, addonItem{a: a, label: fmt.Sprintf("%s — %s [%s]", a.ID, a.Name, st)})
	}
	m.addonList.SetItems(items)
}
