package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var (
	fetchSource []string
	fetchLimit  int
	fetchSince  string
	fetchForce  bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch content from configured sources",
	Long:  "Fetch content from HackerNews, Reddit, RSS feeds, and GitHub",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		var since time.Duration
		if fetchSince != "" {
			since, err = time.ParseDuration(fetchSince)
			if err != nil {
				return fmt.Errorf("invalid --since duration: %w", err)
			}
		}

		opts := app.FetchOptions{
			Sources: fetchSource,
			Limit:   fetchLimit,
			Since:   since,
			Force:   fetchForce,
		}

		return application.Fetch(ctx, opts)
	},
}

func init() {
	fetchCmd.Flags().StringSliceVar(&fetchSource, "source", nil, "Source to fetch from (hackernews, reddit, rss, github)")
	fetchCmd.Flags().IntVar(&fetchLimit, "limit", 0, "Limit number of items per source")
	fetchCmd.Flags().StringVar(&fetchSince, "since", "", "Fetch items since duration (e.g., 24h, 7d)")
	fetchCmd.Flags().BoolVar(&fetchForce, "force", false, "Force re-fetch even if cached")
}
