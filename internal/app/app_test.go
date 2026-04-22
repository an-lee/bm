package app

import (
	"testing"

	"bm/internal/testxdg"
)

func TestNew_andReload(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	testxdg.Reload()
	a, err := New()
	if err != nil || a == nil {
		t.Fatalf("New: %v", err)
	}
	if err := a.Reload(); err != nil {
		t.Fatal(err)
	}
}
