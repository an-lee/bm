package streams

import (
	"sort"
	"strings"
)

// Stream list sort modes (CLI --order, config ui.stream_order, TUI).
const (
	OrderRank    = "rank"
	OrderRankAsc = "rank-asc"
	OrderAddon   = "addon"
	OrderTitle   = "title"
)

// NormalizeOrder maps user input to a known mode; empty defaults to rank.
func NormalizeOrder(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", OrderRank, "quality", "score":
		return OrderRank
	case OrderRankAsc, "worst":
		return OrderRankAsc
	case OrderAddon, "source", "addon_name":
		return OrderAddon
	case OrderTitle, "name":
		return OrderTitle
	default:
		return OrderRank
	}
}

// NextStreamOrder cycles rank → rank-asc → addon → title → rank (for TUI).
func NextStreamOrder(current string) string {
	const n = 4
	cycle := [n]string{OrderRank, OrderRankAsc, OrderAddon, OrderTitle}
	cur := NormalizeOrder(current)
	next := 0
	for i := 0; i < n; i++ {
		if cycle[i] == cur {
			next = (i + 1) % n
			break
		}
	}
	return cycle[next]
}

// ApplySort reorders slice in place (after dedupe).
func ApplySort(slice []ResolvedStream, order string) {
	ord := NormalizeOrder(order)
	switch ord {
	case OrderRankAsc:
		sort.Slice(slice, func(i, j int) bool {
			return streamRank(slice[i]) < streamRank(slice[j])
		})
	case OrderAddon:
		sort.Slice(slice, func(i, j int) bool {
			ai := addonSortKey(slice[i])
			aj := addonSortKey(slice[j])
			if ai != aj {
				return ai < aj
			}
			return titleSortKey(slice[i]) < titleSortKey(slice[j])
		})
	case OrderTitle:
		sort.Slice(slice, func(i, j int) bool {
			ti := titleSortKey(slice[i])
			tj := titleSortKey(slice[j])
			if ti != tj {
				return ti < tj
			}
			return addonSortKey(slice[i]) < addonSortKey(slice[j])
		})
	default: // rank
		sort.Slice(slice, func(i, j int) bool {
			return streamRank(slice[i]) > streamRank(slice[j])
		})
	}
}

func streamRank(s ResolvedStream) int {
	score := 0
	if strings.TrimSpace(s.PlayableURL()) != "" {
		score += 10
	}
	t := strings.ToLower(s.Title + " " + s.Name)
	if strings.Contains(t, "1080") {
		score += 3
	}
	if strings.Contains(t, "720") {
		score += 2
	}
	if strings.Contains(t, "4k") || strings.Contains(t, "2160") {
		score += 4
	}
	return score
}

func addonSortKey(s ResolvedStream) string {
	return strings.ToLower(strings.TrimSpace(s.AddonName) + "\x00" + strings.TrimSpace(s.AddonID))
}

func titleSortKey(s ResolvedStream) string {
	t := strings.TrimSpace(s.Title)
	n := strings.TrimSpace(s.Name)
	if t == "" {
		t = n
	}
	return strings.ToLower(t + "\x00" + n)
}
