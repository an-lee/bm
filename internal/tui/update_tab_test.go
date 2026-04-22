package tui

import (
	"testing"
)

func TestOnTabChange_settings(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSettings
	cmd := m.onTabChange()
	if cmd == nil {
		t.Fatal()
	}
}

func TestOnTabChange_addonsClearsURLModeWhenLeavingAddons(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabAddons
	m.addonURLMode = true
	m.addonURL.Focus()
	m.tab = tabSearch
	_ = m.onTabChange()
	if m.addonURLMode {
		t.Fatal()
	}
}

func TestRefreshAddonList(t *testing.T) {
	m := newRootModel(testApp(t))
	m.refreshAddonList()
	if len(m.addonList.Items()) == 0 {
		t.Fatal()
	}
}

func TestOnTabChange_searchActive(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSearch
	m.searchActive = true
	cmd := m.onTabChange()
	if cmd == nil {
		t.Fatal()
	}
	_ = cmd
}
