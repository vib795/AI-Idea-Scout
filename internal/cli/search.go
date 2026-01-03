package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var searchFormat string

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search ideas and raw content using FTS5",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		return application.Search(ctx, args[0], searchFormat)
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchFormat, "format", "text", "Output format (text/json)")
}
