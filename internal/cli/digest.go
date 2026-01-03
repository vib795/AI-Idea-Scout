package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var (
	digestSince  string
	digestTop    int
	digestFormat string
)

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Generate a digest of top ideas",
	Long:  "Create a summary report of the best ideas from a time period",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		opts := app.DigestOptions{
			Since:  digestSince,
			Top:    digestTop,
			Format: digestFormat,
		}

		return application.Digest(ctx, opts)
	},
}

func init() {
	digestCmd.Flags().StringVar(&digestSince, "since", "24h", "Time period for digest")
	digestCmd.Flags().IntVar(&digestTop, "top", 10, "Number of top ideas to include")
	digestCmd.Flags().StringVar(&digestFormat, "format", "md", "Output format (md/text/json)")
}
