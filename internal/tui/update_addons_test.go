package tui

import (
	"testing"

	"bm/internal/config"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateAddons_openURLMode(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabAddons
	out, cmd := m.updateAddons(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	rm := out.(*rootModel)
	if !rm.addonURLMode || cmd == nil {
		t.Fatal(rm.addonURLMode, cmd)
	}
}

func TestUpdateAddons_removeCinemetaRefused(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabAddons
	m.addonList.SetItems([]list.Item{
		addonItem{a: config.AddonEntry{ID: config.CinemetaAddonID, Name: "Cinemeta", Enabled: true}, label: "cinemeta"},
	})
	m.addonList.Select(0)
	out, cmd := m.updateAddons(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil || out.(*rootModel).toast == "" {
		t.Fatal()
	}
}
