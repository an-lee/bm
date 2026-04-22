package clipboard

import (
	"strings"
	"testing"
)

func TestWriteAll_empty(t *testing.T) {
	t.Parallel()
	err := WriteAll("   ")
	if err == nil || !strings.Contains(err.Error(), "nothing to copy") {
		t.Fatalf("err %v", err)
	}
}

func TestWriteAll_nonEmpty(t *testing.T) {
	t.Parallel()
	err := WriteAll("hello")
	if err != nil {
		t.Skip("clipboard unavailable:", err)
	}
}
