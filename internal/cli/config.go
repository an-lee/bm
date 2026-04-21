package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"bm/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read or write bm configuration",
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		p, err := config.Path()
		if err != nil {
			return err
		}
		fmt.Println(p)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value (e.g. tmdb.api_key)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.SetKey(args[0], args[1]); err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]string{args[0]: args[1]})
		}
		cmd.Println("ok")
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Print a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := config.GetKey(args[0])
		if err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]string{args[0]: v})
		}
		fmt.Println(v)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configPathCmd, configSetCmd, configGetCmd)
}
