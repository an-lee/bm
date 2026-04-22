package tui

import (
	"strings"
	"testing"

	"bm/internal/config"
	"bm/internal/search"
	"bm/internal/streams"
	"bm/internal/stremio"
)

func TestTitleItem_listMethods(t *testing.T) {
	t.Parallel()
	i := titleItem{r: search.TitleResult{Title: "T", IMDBID: "tt1", Year: "2020", Type: "movie"}}
	if i.Title() != "T" || !strings.Contains(i.Description(), "tt1") || i.FilterValue() != "T" {
		t.Fatal()
	}
}

func TestSeasonHeaderItem(t *testing.T) {
	t.Parallel()
	i := seasonHeaderItem{season: 2}
	if !strings.Contains(i.Title(), "2") || i.Description() != "" || i.FilterValue() != "season:2" {
		t.Fatal()
	}
}

func TestEpisodeSE(t *testing.T) {
	t.Parallel()
	s, e := episodeSE(stremio.Video{Season: 1, Episode: 2})
	if s != 1 || e != 2 {
		t.Fatal()
	}
	s2, e2 := episodeSE(stremio.Video{IMDBSeason: 3, IMDBEpisode: 4})
	if s2 != 3 || e2 != 4 {
		t.Fatal()
	}
}

func TestEpisodeItem_listMethods(t *testing.T) {
	t.Parallel()
	i := episodeItem{v: stremio.Video{Season: 1, Episode: 2, Title: "Ep"}}
	if !strings.Contains(i.Title(), "S01E02") || i.FilterValue() == "" {
		t.Fatal()
	}
}

func TestBuildEpisodeListItems(t *testing.T) {
	t.Parallel()
	items := buildEpisodeListItems([]stremio.Video{
		{Season: 2, Episode: 1, Title: "b"},
		{Season: 1, Episode: 2, Title: "a"},
	})
	if len(items) < 4 {
		t.Fatalf("len %d", len(items))
	}
}

func TestStreamItem_Description(t *testing.T) {
	t.Parallel()
	rs := streams.ResolvedStream{
		AddonName: "A",
		Stream: stremio.Stream{
			Title:         "T",
			URL:           "https://example.com/" + strings.Repeat("x", 200),
			BehaviorHints: map[string]any{"seeds": 5},
		},
	}
	i := streamItem{s: rs}
	desc := i.Description()
	if !strings.Contains(desc, "seeds") {
		t.Fatal(desc)
	}
}

func TestStreamItem_Description_noURL(t *testing.T) {
	t.Parallel()
	rs := streams.ResolvedStream{
		AddonName: "A",
		Stream:    stremio.Stream{Title: "T", BehaviorHints: map[string]any{"seeds": 3}},
	}
	i := streamItem{s: rs}
	if i.Description() != "3 seeds" {
		t.Fatal(i.Description())
	}
}

func TestAddonItem(t *testing.T) {
	t.Parallel()
	i := addonItem{
		a:     config.AddonEntry{ID: "id", ManifestURL: "https://m"},
		label: "label",
	}
	if i.Title() != "label" || i.Description() != "https://m" || i.FilterValue() != "id" {
		t.Fatal()
	}
}
