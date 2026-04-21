package cli

import (
	"context"

	"github.com/spf13/cobra"

	"bm/internal/app"
)

var searchType string
var searchYear int

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search titles via TMDB (if configured) or catalog addons",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := app.New()
		if err != nil {
			return err
		}
		ctx := context.Background()
		res, err := a.Search.Search(ctx, args[0], searchType, searchYear)
		if err != nil {
			return err
		}
		if jsonOutput {
			return printJSON(res)
		}
		for _, r := range res {
			year := r.Year
			if year != "" {
				year = " (" + year + ")"
			}
			cmd.Printf("%s\t%s%s\t%s\n", r.IMDBID, r.Title, year, r.Type)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchType, "type", "", "movie or series (default: config ui.default_type)")
	searchCmd.Flags().IntVar(&searchYear, "year", 0, "filter by release year (TMDB mode only)")
	_ = searchCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"movie", "series"}, cobra.ShellCompDirectiveNoFileComp
	})
}
