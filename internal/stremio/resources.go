package stremio

import (
	"encoding/json"
	"fmt"
)

// jsonResources unmarshals Stremio "resources" which may be []string or []objects.
type jsonResources []ResourceSpec

// ResourceSpec is a normalized resource entry.
type ResourceSpec struct {
	Name       string   `json:"name"`
	Types      []string `json:"types"`
	IDPrefixes []string `json:"idPrefixes"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (jr *jsonResources) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch v := raw.(type) {
	case []any:
		out := make([]ResourceSpec, 0, len(v))
		for _, item := range v {
			switch t := item.(type) {
			case string:
				out = append(out, ResourceSpec{Name: t})
			case map[string]any:
				var rs ResourceSpec
				if n, ok := t["name"].(string); ok {
					rs.Name = n
				}
				if types, ok := t["types"].([]any); ok {
					for _, x := range types {
						if s, ok := x.(string); ok {
							rs.Types = append(rs.Types, s)
						}
					}
				}
				if prefs, ok := t["idPrefixes"].([]any); ok {
					for _, x := range prefs {
						if s, ok := x.(string); ok {
							rs.IDPrefixes = append(rs.IDPrefixes, s)
						}
					}
				}
				out = append(out, rs)
			default:
				return fmt.Errorf("unexpected resource item type %T", item)
			}
		}
		*jr = out
		return nil
	default:
		return fmt.Errorf("unexpected resources type %T", raw)
	}
}

// Names returns resource names like "catalog", "stream".
func (jr jsonResources) Names() []string {
	names := make([]string, 0, len(jr))
	for _, r := range jr {
		if r.Name != "" {
			names = append(names, r.Name)
		}
	}
	return names
}

// SupportsStream returns true if stream resource exists and matches type/prefix when specified.
func (jr jsonResources) SupportsStream(metaType, imdbID string) bool {
	for _, r := range jr {
		if r.Name != "stream" {
			continue
		}
		if len(r.Types) > 0 && !contains(r.Types, metaType) {
			continue
		}
		if len(r.IDPrefixes) == 0 {
			return true
		}
		ok := false
		for _, p := range r.IDPrefixes {
			if p != "" && len(imdbID) >= len(p) && imdbID[:len(p)] == p {
				ok = true
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func contains[S ~[]E, E comparable](s S, v E) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
