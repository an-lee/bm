package streams

import (
	"bm/internal/stremio"
	"testing"
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
