package tui

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/list"

	"bm/internal/config"
	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
)

type titleItem struct{ r search.TitleResult }

func (i titleItem) Title() string       { return i.r.Title }
func (i titleItem) Description() string { return i.r.IMDBID + "  " + i.r.Year + "  " + i.r.Type }
func (i titleItem) FilterValue() string { return i.r.Title }

type seasonHeaderItem struct{ season int }

func (i seasonHeaderItem) Title() string {
	return fmt.Sprintf("── Season %d ──", i.season)
}
func (i seasonHeaderItem) Description() string { return "" }
func (i seasonHeaderItem) FilterValue() string {
	return "season:" + strconv.Itoa(i.season)
}

type episodeItem struct{ v stremio.Video }

func (i episodeItem) Title() string {
	s, e := episodeSE(i.v)
	return fmt.Sprintf("S%02dE%02d — %s", s, e, i.v.Title)
}

func (i episodeItem) Description() string { return i.v.Released }
func (i episodeItem) FilterValue() string {
	s, e := episodeSE(i.v)
	return fmt.Sprintf("%d %d %s", s, e, i.v.Title)
}

func episodeSE(v stremio.Video) (season, episode int) {
	season, episode = v.Season, v.Episode
	if season == 0 && v.IMDBSeason != 0 {
		season = v.IMDBSeason
	}
	if episode == 0 && v.IMDBEpisode != 0 {
		episode = v.IMDBEpisode
	}
	return season, episode
}

func buildEpisodeListItems(videos []stremio.Video) []list.Item {
	vids := append([]stremio.Video(nil), videos...)
	sort.Slice(vids, func(i, j int) bool {
		si, ei := episodeSE(vids[i])
		sj, ej := episodeSE(vids[j])
		if si != sj {
			return si < sj
		}
		if ei != ej {
			return ei < ej
		}
		return vids[i].Title < vids[j].Title
	})
	var items []list.Item
	const unset = int(^uint(0) >> 1)
	lastSeason := unset
	for _, v := range vids {
		s, _ := episodeSE(v)
		if s != lastSeason {
			items = append(items, seasonHeaderItem{season: s})
			lastSeason = s
		}
		items = append(items, episodeItem{v: v})
	}
	return items
}

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
