package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateSettings_enterSaves(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSettings
	m.settingsInput.SetValue("mykey")
	out, cmd := m.updateSettings(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal()
	}
	if out.(*rootModel).settingsInput.Value() != "" {
		t.Fatal("expected cleared input")
	}
}

func TestUpdateSettings_enterEmpty(t *testing.T) {
	m := newRootModel(testApp(t))
	m.width = 80
	m.tab = tabSettings
	m.settingsInput.SetValue("  ")
	_, cmd := m.updateSettings(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal()
	}
}
