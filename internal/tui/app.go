package tui

import "bm/internal/app"

func loadApp() (*app.App, error) {
	return app.New()
}
