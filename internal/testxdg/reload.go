package testxdg

import "github.com/adrg/xdg"

// Reload refreshes github.com/adrg/xdg cached paths after changing XDG_* env vars in tests.
func Reload() {
	xdg.Reload()
}
