package cli

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	err = printJSON(map[string]string{"hello": "world"})
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(r)
	_ = r.Close()
	var m map[string]string
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	if m["hello"] != "world" {
		t.Fatal(m)
	}
}
