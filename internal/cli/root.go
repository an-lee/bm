package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// osExit is overridden in tests to avoid terminating the test process.
var osExit = os.Exit

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
}

var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "bm",
	Short: "Stremio-compatible movie/TV search and stream resolver",
	Long: `bm searches catalogs via installed Stremio addons, resolves stream URLs,
and copies links to the clipboard. Run with no arguments for the TUI.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI(cmd)
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "machine-readable JSON (where supported)")
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(streamCmd)
	rootCmd.AddCommand(addonsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(mcpCmd)
}
