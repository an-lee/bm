package cli

import (
	"github.com/spf13/cobra"

	"bm/internal/tui"
)

// tuiRun is swapped in tests so the root command can be exercised without a terminal.
var tuiRun = tui.Run

func runTUI(cmd *cobra.Command) error {
	return tuiRun()
}
