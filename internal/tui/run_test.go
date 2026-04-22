package tui

import (
	"testing"

	"bm/internal/testxdg"
)

func TestNewAppModel(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	m, err := newAppModel()
	if err != nil {
		t.Fatal(err)
	}
	if m == nil {
		t.Fatal("nil model")
	}
}
