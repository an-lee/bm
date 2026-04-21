package cli

import (
	"context"

	"github.com/spf13/cobra"

	"bm/internal/app"
)

var addonsCmd = &cobra.Command{
	Use:   "addons",
	Short: "Manage installed Stremio addons",
}

var addonsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List addons",
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := app.New()
		if err != nil {
			return err
		}
		list := a.Addons.List(true)
		if jsonOutput {
			return printJSON(list)
		}
		for _, x := range list {
			en := "disabled"
			if x.Enabled {
				en = "enabled"
			}
			cmd.Printf("%s\t%s\t%s\n", x.ID, x.Name, en)
		}
		return nil
	},
}

var addonsAddCmd = &cobra.Command{
	Use:   "add <manifest_url>",
	Short: "Install an addon by manifest URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := app.New()
		if err != nil {
			return err
		}
		ctx := context.Background()
		entry, err := a.Addons.Install(ctx, args[0])
		if err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(entry)
		}
		cmd.Printf("installed %s (%s)\n", entry.Name, entry.ID)
		return nil
	},
}

var addonsRemoveCmd = &cobra.Command{
	Use:   "remove <addon_id>",
	Short: "Remove an addon by id",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := app.New()
		if err != nil {
			return err
		}
		if err := a.Addons.Remove(args[0]); err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(map[string]string{"removed": args[0]})
		}
		cmd.Println("removed", args[0])
		return nil
	},
}

func init() {
	addonsCmd.AddCommand(addonsListCmd, addonsAddCmd, addonsRemoveCmd)
}
