package cli

import (
	"github.com/spf13/cobra"

	"bm/internal/tui"
)

func runTUI(cmd *cobra.Command) error {
	return tui.Run()
}
