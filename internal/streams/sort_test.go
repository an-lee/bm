package streams

import (
	"reflect"
	"testing"

	"bm/internal/stremio"
)

func TestNormalizeOrder(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"":               OrderRank,
		"  RANK  ":       OrderRank,
		"quality":        OrderRank,
		"score":          OrderRank,
		"rank-asc":       OrderRankAsc,
		"worst":          OrderRankAsc,
		"addon":          OrderAddon,
		"source":         OrderAddon,
		"addon_name":     OrderAddon,
		"title":          OrderTitle,
		"name":           OrderTitle,
		"seeds":          OrderSeeds,
		"seeders":        OrderSeeds,
		"peer":           OrderSeeds,
		"peers":          OrderSeeds,
		"seeds-asc":      OrderSeedsAsc,
		"seeds-ascend":   OrderSeedsAsc,
		"fewest-seeds":   OrderSeedsAsc,
		"unknown-value":  OrderRank,
	}
	for in, want := range cases {
		if got := NormalizeOrder(in); got != want {
			t.Fatalf("NormalizeOrder(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNextStreamOrder(t *testing.T) {
	t.Parallel()
	cur := OrderRank
	seen := map[string]bool{}
	for i := 0; i < 6; i++ {
		seen[cur] = true
		cur = NextStreamOrder(cur)
	}
	if cur != OrderRank {
		t.Fatalf("expected cycle back to rank after 6 steps, got %q", cur)
	}
	if len(seen) < 6 {
		t.Fatalf("expected 6 distinct orders in cycle, got %v", seen)
	}
}

func sampleStreams() []ResolvedStream {
	return []ResolvedStream{
		{AddonID: "b", AddonName: "B", Stream: stremio.Stream{Name: "720p", Title: "T2", URL: "https://u2"}},
		{AddonID: "a", AddonName: "A", Stream: stremio.Stream{Name: "1080p", Title: "T1", URL: "https://u1"}},
		{AddonID: "a", AddonName: "A", Stream: stremio.Stream{Name: "4k", Title: "T3", URL: "https://u3"}},
	}
}

func TestApplySort_rankDesc(t *testing.T) {
	t.Parallel()
	s := sampleStreams()
	ApplySort(s, OrderRank)
	want := []string{"https://u3", "https://u1", "https://u2"}
	got := urls(s)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rank desc order: %#v", got)
	}
}

func TestApplySort_rankAsc(t *testing.T) {
	t.Parallel()
	s := sampleStreams()
	ApplySort(s, OrderRankAsc)
	want := []string{"https://u2", "https://u1", "https://u3"}
	got := urls(s)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rank asc order: %#v", got)
	}
}

func TestApplySort_addonThenTitle(t *testing.T) {
	t.Parallel()
	s := sampleStreams()
	ApplySort(s, OrderAddon)
	got := make([]string, len(s))
	for i := range s {
		got[i] = s[i].AddonID + ":" + s[i].Title
	}
	want := []string{"a:T1", "a:T3", "b:T2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("addon order: %#v", got)
	}
}

func TestApplySort_title(t *testing.T) {
	t.Parallel()
	s := sampleStreams()
	ApplySort(s, OrderTitle)
	got := titles(s)
	want := []string{"T1", "T2", "T3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("title order: %#v", got)
	}
}

func TestApplySort_seedsDescAndAsc(t *testing.T) {
	t.Parallel()
	s := []ResolvedStream{
		{Stream: stremio.Stream{Name: "a", BehaviorHints: map[string]any{"seeds": 1}}},
		{Stream: stremio.Stream{Name: "b", BehaviorHints: map[string]any{"seeds": 10}}},
		{Stream: stremio.Stream{Name: "c"}},
	}
	ApplySort(s, OrderSeeds)
	got := names(s)
	if got[0] != "b" || got[1] != "a" || got[2] != "c" {
		t.Fatalf("seeds desc: %#v", got)
	}
	s2 := []ResolvedStream{
		{Stream: stremio.Stream{Name: "a", BehaviorHints: map[string]any{"seeds": 1}}},
		{Stream: stremio.Stream{Name: "b", BehaviorHints: map[string]any{"seeds": 10}}},
		{Stream: stremio.Stream{Name: "c"}},
	}
	ApplySort(s2, OrderSeedsAsc)
	got2 := names(s2)
	if got2[0] != "a" || got2[1] != "b" || got2[2] != "c" {
		t.Fatalf("seeds asc: %#v", got2)
	}
}

func urls(s []ResolvedStream) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = s[i].URL
	}
	return out
}

func titles(s []ResolvedStream) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = s[i].Title
	}
	return out
}

func names(s []ResolvedStream) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = s[i].Name
	}
	return out
}
