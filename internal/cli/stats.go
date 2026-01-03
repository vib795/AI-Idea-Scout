package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about collected data",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		return application.Stats(ctx)
	},
}
