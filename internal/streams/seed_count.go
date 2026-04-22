package streams

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	reSeedsWord = regexp.MustCompile(`(?i)\b(\d{1,7})\s*seeds?\b`)
	reSLSlash   = regexp.MustCompile(`(?i)\b[Ss]\s*[:/]?\s*L\s*[:/]?\s*(\d{1,7})\s*/\s*(\d{1,7})\b`)
)

// StreamSeedValue returns a torrent seed/peer count when the addon encodes it in
// behaviorHints and/or in name, title, or description text. The second return is
// false when no value could be parsed.
func StreamSeedValue(s ResolvedStream) (int, bool) {
	if c, ok := seedFromBehaviorHints(s.BehaviorHints); ok {
		return c, true
	}
	if c, ok := parseSeedText(streamSeedText(s)); ok {
		return c, true
	}
	return 0, false
}

func streamSeedText(s ResolvedStream) string {
	var parts []string
	for _, p := range []string{
		strings.TrimSpace(s.Name),
		strings.TrimSpace(s.Title),
		strings.TrimSpace(s.Description),
	} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, " ")
}

var behaviorHintKeys = []string{
	"seeds",
	"seeders",
	"seed",
	"seedCount",
}

func seedFromBehaviorHints(h map[string]any) (int, bool) {
	if len(h) == 0 {
		return 0, false
	}
	for _, k := range behaviorHintKeys {
		if v, ok := h[k]; ok {
			if c, ok := intFromAny(v); ok {
				return c, true
			}
		}
	}
	return 0, false
}

func intFromAny(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, t >= 0
	case int32:
		return int(t), t >= 0
	case int64:
		if t < 0 || t > int64(maxSortableSeed) {
			return 0, false
		}
		return int(t), true
	case float64:
		if t < 0 || t > float64(maxSortableSeed) || math.IsNaN(t) {
			return 0, false
		}
		return int(t), true
	case float32:
		if t < 0 {
			return 0, false
		}
		return int(t), true
	case json.Number:
		n, err := t.Int64()
		if err != nil || n < 0 {
			return 0, false
		}
		if n > int64(maxSortableSeed) {
			return maxSortableSeed, true
		}
		return int(n), true
	case string:
		t = strings.TrimSpace(t)
		if t == "" {
			return 0, false
		}
		n, err := strconv.Atoi(t)
		if err != nil {
			if f, err2 := strconv.ParseFloat(t, 64); err2 == nil && f >= 0 {
				return int(f), true
			}
			return 0, false
		}
		return n, n >= 0
	default:
		return 0, false
	}
}

// Keep parsed seed counts bounded for sane sort keys and regex captures.
const maxSortableSeed = 1_000_000_000

func parseSeedText(text string) (int, bool) {
	if text == "" {
		return 0, false
	}
	// e.g. "S: 12 L: 3" / "S/L 12/3" — use first number as seed estimate.
	if m := reSLSlash.FindStringSubmatchIndex(text); len(m) >= 4 {
		s := text[m[2]:m[3]]
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			if n > maxSortableSeed {
				return maxSortableSeed, true
			}
			return n, true
		}
	}
	if m := reSeedsWord.FindStringSubmatch(text); len(m) >= 2 {
		if n, err := strconv.Atoi(m[1]); err == nil && n >= 0 {
			if n > maxSortableSeed {
				return maxSortableSeed, true
			}
			return n, true
		}
	}
	return 0, false
}
