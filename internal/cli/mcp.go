package cli

import (
	"github.com/spf13/cobra"

	"bm/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run MCP server on stdio for AI clients",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return mcp.ServeStdio()
	},
}
