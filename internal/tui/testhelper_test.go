package tui

import (
	"testing"

	"bm/internal/app"
	"bm/internal/testxdg"
)

func testApp(t *testing.T) *app.App {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	a, err := app.New()
	if err != nil {
		t.Fatal(err)
	}
	return a
}
