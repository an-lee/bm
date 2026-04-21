package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"bm/internal/app"
	"bm/internal/clipboard"
	"bm/internal/streams"
)

var streamSeason, streamEpisode int
var streamType string
var streamCopy bool
var streamOrder string

var streamCmd = &cobra.Command{
	Use:   "stream <imdb_id>",
	Short: "Resolve streams from all stream-capable addons",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := app.New()
		if err != nil {
			return err
		}
		rawID := strings.TrimSpace(args[0])
		metaType := streamType
		imdbID := rawID
		if metaType == "" {
			metaType = a.Config.UI.DefaultType
		}
		season, episode := streamSeason, streamEpisode
		if strings.Contains(rawID, ":") {
			parts := strings.Split(rawID, ":")
			if len(parts) >= 3 {
				imdbID = parts[0]
				metaType = "series"
				_, _ = fmt.Sscanf(parts[1], "%d", &season)
				_, _ = fmt.Sscanf(parts[2], "%d", &episode)
			}
		}
		if metaType == "series" && (season < 1 || episode < 1) {
			return fmt.Errorf("series requires --season and --episode (or id format tt12345:1:1)")
		}
		ctx := context.Background()
		list, err := a.Streams.Resolve(ctx, imdbID, metaType, season, episode)
		if err != nil {
			return err
		}
		if strings.TrimSpace(streamOrder) != "" {
			streams.ApplySort(list, streamOrder)
		}
		if jsonOutput {
			out := make([]streamJSONRow, 0, len(list))
			for _, s := range list {
				out = append(out, streamJSONRow{
					AddonID:     s.AddonID,
					AddonName:   s.AddonName,
					Name:        s.Name,
					Title:       s.Title,
					URL:         s.URL,
					InfoHash:    s.InfoHash,
					PlayableURL: s.PlayableURL(),
				})
			}
			return printJSON(out)
		}
		for _, s := range list {
			u := s.PlayableURL()
			if u == "" {
				continue
			}
			cmd.Printf("[%s] %s %s\n  %s\n", s.AddonName, s.Name, s.Title, u)
		}
		if streamCopy && len(list) > 0 {
			u := list[0].PlayableURL()
			if u == "" {
				return fmt.Errorf("first stream has no URL or torrent hash")
			}
			if err := clipboard.WriteAll(u); err != nil {
				return err
			}
			cmd.Printf("copied: %s\n", u)
		}
		return nil
	},
}

type streamJSONRow struct {
	AddonID     string `json:"addon_id"`
	AddonName   string `json:"addon_name"`
	Name        string `json:"name,omitempty"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	InfoHash    string `json:"info_hash,omitempty"`
	PlayableURL string `json:"playable_url"`
}

func init() {
	streamCmd.Flags().IntVar(&streamSeason, "season", 0, "season (series)")
	streamCmd.Flags().IntVar(&streamEpisode, "episode", 0, "episode (series)")
	streamCmd.Flags().StringVar(&streamType, "type", "", "movie or series (default from config)")
	streamCmd.Flags().BoolVar(&streamCopy, "copy", false, "copy first playable URL to clipboard")
	streamCmd.Flags().StringVar(&streamOrder, "order", "", "sort streams: rank, rank-asc, addon, title (overrides config ui.stream_order)")
}
