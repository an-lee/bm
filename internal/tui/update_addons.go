package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/browser"

	"bm/internal/config"
)

func (m *rootModel) updateAddons(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addonURLMode {
		switch msg.String() {
		case "enter":
			url := strings.TrimSpace(m.addonURL.Value())
			if url == "" {
				return m, nil
			}
			return m, m.installAddon(url)
		default:
			var cmd tea.Cmd
			m.addonURL, cmd = m.addonURL.Update(msg)
			return m, cmd
		}
	}
	switch msg.String() {
	case "a":
		m.addonURLMode = true
		m.addonURL.Focus()
		return m, textinput.Blink
	case "d":
		if it, ok := m.addonList.SelectedItem().(addonItem); ok {
			id := it.a.ID
			if id == "com.linvo.cinemeta" {
				m.toast = "Refusing to remove Cinemeta"
				return m, m.tickToast()
			}
			return m, m.removeAddon(id)
		}
	case "c":
		if it, ok := m.addonList.SelectedItem().(addonItem); ok {
			return m, m.openAddonConfig(it.a)
		}
	}
	var cmd tea.Cmd
	m.addonList, cmd = m.addonList.Update(msg)
	return m, cmd
}

func (m *rootModel) installAddon(url string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		entry, err := m.app.Addons.Install(ctx, url)
		if err != nil {
			return statusMsg{err: err}
		}
		return statusMsg{text: "installed " + entry.ID}
	}
}

func (m *rootModel) removeAddon(id string) tea.Cmd {
	return func() tea.Msg {
		err := m.app.Addons.Remove(id)
		if err != nil {
			return statusMsg{err: err}
		}
		_ = m.app.Reload()
		return statusMsg{text: "removed " + id}
	}
}

func (m *rootModel) openAddonConfig(a config.AddonEntry) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		confURL := strings.TrimSpace(a.ConfigurationURL)
		if confURL == "" {
			man, err := m.app.Client.GetManifest(ctx, a.ManifestURL)
			if err == nil && man.BehaviorHints.OpenURLTemplate != "" {
				confURL = man.BehaviorHints.OpenURLTemplate
			}
		}
		if confURL == "" {
			return statusMsg{err: fmt.Errorf("no configuration URL for this addon")}
		}
		_ = browser.OpenURL(confURL)
		return statusMsg{text: "opened configuration in browser"}
	}
}
