package tui

import tea "github.com/charmbracelet/bubbletea"

// Run starts the full-screen TUI.
func Run() error {
	a, err := newAppModel()
	if err != nil {
		return err
	}
	p := tea.NewProgram(a, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func newAppModel() (tea.Model, error) {
	ap, err := loadApp()
	if err != nil {
		return nil, err
	}
	return newRootModel(ap), nil // *rootModel implements tea.Model
}
