package tui

import (
	"bm/internal/config"
	"bm/internal/search"
	"bm/internal/streams"
)

type titleItem struct{ r search.TitleResult }

func (i titleItem) Title() string       { return i.r.Title }
func (i titleItem) Description() string { return i.r.IMDBID + "  " + i.r.Year + "  " + i.r.Type }
func (i titleItem) FilterValue() string { return i.r.Title }

type streamItem struct{ s streams.ResolvedStream }

func (i streamItem) Title() string {
	t := i.s.Title
	if t == "" {
		t = i.s.Name
	}
	return "[" + i.s.AddonName + "] " + t
}

func (i streamItem) Description() string {
	u := i.s.PlayableURL()
	if len(u) > 120 {
		return u[:117] + "..."
	}
	return u
}

func (i streamItem) FilterValue() string { return i.s.Title + i.s.Name }

type addonItem struct {
	a     config.AddonEntry
	label string
}

func (i addonItem) Title() string       { return i.label }
func (i addonItem) Description() string { return i.a.ManifestURL }
func (i addonItem) FilterValue() string { return i.a.ID }
