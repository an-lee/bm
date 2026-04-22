package streams

import (
	"encoding/json"
	"math"
	"testing"

	"bm/internal/stremio"
)

func TestStreamSeedValue(t *testing.T) {
	cases := []struct {
		name  string
		s     ResolvedStream
		want  int
		wantOk bool
	}{
		{
			"behaviorHints float64",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": float64(42)}}},
			42, true,
		},
		{
			"text seeds word",
			ResolvedStream{Stream: stremio.Stream{Name: "1080p 123 seeds something"}},
			123, true,
		},
		{
			"SL format",
			ResolvedStream{Stream: stremio.Stream{Title: "S/L 5/2"}},
			5, true,
		},
		{
			"no data",
			ResolvedStream{Stream: stremio.Stream{Name: "x"}},
			0, false,
		},
		{
			"behaviorHints int",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeders": 7}}},
			7, true,
		},
		{
			"behaviorHints int32",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seed": int32(9)}}},
			9, true,
		},
		{
			"behaviorHints int64",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seedCount": int64(11)}}},
			11, true,
		},
		{
			"behaviorHints json.Number",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": json.Number("13")}}},
			13, true,
		},
		{
			"behaviorHints string int",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": " 15 "}}},
			15, true,
		},
		{
			"behaviorHints string float",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": "2.5"}}},
			2, true,
		},
		{
			"behaviorHints negative int rejected",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": -1}}},
			0, false,
		},
		{
			"behaviorHints float64 NaN rejected",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": math.NaN()}}},
			0, false,
		},
		{
			"behaviorHints int64 overflow rejected",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": int64(maxSortableSeed + 1)}}},
			0, false,
		},
		{
			"behaviorHints unknown type",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": []int{1}}}},
			0, false,
		},
		{
			"behaviorHints json.Number over max caps",
			ResolvedStream{Stream: stremio.Stream{BehaviorHints: map[string]any{"seeds": json.Number("2000000001")}}},
			maxSortableSeed, true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, ok := StreamSeedValue(tc.s)
			if n != tc.want || ok != tc.wantOk {
				t.Fatalf("StreamSeedValue() = (%d, %v), want (%d, %v)", n, ok, tc.want, tc.wantOk)
			}
		})
	}
}
