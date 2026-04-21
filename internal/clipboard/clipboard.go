package clipboard

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
)

// WriteAll copies text to the system clipboard.
func WriteAll(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return fmt.Errorf("nothing to copy")
	}
	return clipboard.WriteAll(text)
}
