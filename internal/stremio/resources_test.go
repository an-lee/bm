package stremio

import (
	"encoding/json"
	"testing"
)

func TestJsonResources_UnmarshalJSON_stringArray(t *testing.T) {
	t.Parallel()
	var jr jsonResources
	data := []byte(`["catalog","stream","meta"]`)
	if err := json.Unmarshal(data, &jr); err != nil {
		t.Fatal(err)
	}
	names := jr.Names()
	if len(names) != 3 || names[0] != "catalog" || names[1] != "stream" || names[2] != "meta" {
		t.Fatalf("Names() = %#v", names)
	}
}

func TestJsonResources_UnmarshalJSON_objectArray(t *testing.T) {
	t.Parallel()
	raw := `[{"name":"stream","types":["movie"],"idPrefixes":["tt","kitsu"]},{"name":"catalog"}]`
	var jr jsonResources
	if err := json.Unmarshal([]byte(raw), &jr); err != nil {
		t.Fatal(err)
	}
	if len(jr) != 2 || jr[0].Name != "stream" || len(jr[0].Types) != 1 || jr[0].Types[0] != "movie" {
		t.Fatalf("first spec: %#v", jr[0])
	}
	if len(jr[0].IDPrefixes) != 2 {
		t.Fatalf("prefixes: %#v", jr[0].IDPrefixes)
	}
}

func TestJsonResources_UnmarshalJSON_errors(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{}`,
		`["catalog",123]`,
		`not-json`,
	} {
		var jr jsonResources
		if err := json.Unmarshal([]byte(raw), &jr); err == nil {
			t.Fatalf("expected error for %s", raw)
		}
	}
}

func TestJsonResources_Names_skipsEmpty(t *testing.T) {
	t.Parallel()
	jr := jsonResources{{Name: "a"}, {Name: ""}, {Name: "b"}}
	n := jr.Names()
	if len(n) != 2 || n[0] != "a" || n[1] != "b" {
		t.Fatalf("got %#v", n)
	}
}

func TestJsonResources_SupportsStream(t *testing.T) {
	t.Parallel()
	jr := jsonResources{
		{Name: "catalog"},
		{Name: "stream", Types: []string{"movie"}, IDPrefixes: []string{"tt"}},
	}
	if !jr.SupportsStream("movie", "tt123") {
		t.Fatal("expected match for movie tt prefix")
	}
	if jr.SupportsStream("series", "tt123") {
		t.Fatal("type mismatch should not match")
	}
	if jr.SupportsStream("movie", "nm123") {
		t.Fatal("prefix mismatch")
	}
	any := jsonResources{{Name: "stream"}}
	if !any.SupportsStream("anime", "anything") {
		t.Fatal("empty types/prefixes should match any")
	}
}
