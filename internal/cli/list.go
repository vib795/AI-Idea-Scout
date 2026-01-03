package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var (
	listTop        int
	listCategory   string
	listComplexity string
	listStatus     string
	listSince      string
	listFormat     string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List extracted ideas",
	Long:  "Display ideas sorted by overall score with filtering options",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		opts := app.ListOptions{
			Top:        listTop,
			Category:   listCategory,
			Complexity: listComplexity,
			Status:     listStatus,
			Since:      listSince,
			Format:     listFormat,
		}

		return application.List(ctx, opts)
	},
}

func init() {
	listCmd.Flags().IntVar(&listTop, "top", 20, "Show top N ideas")
	listCmd.Flags().StringVar(&listCategory, "category", "", "Filter by category")
	listCmd.Flags().StringVar(&listComplexity, "complexity", "", "Filter by complexity (weekend/week/month/quarter)")
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (new/reviewed/shortlisted/built/archived)")
	listCmd.Flags().StringVar(&listSince, "since", "", "Show ideas since duration (e.g., 24h)")
	listCmd.Flags().StringVar(&listFormat, "format", "text", "Output format (text/json/md)")
}
