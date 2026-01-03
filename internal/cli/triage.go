package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var triageCmd = &cobra.Command{
	Use:   "triage",
	Short: "Interactive triage workflow for reviewing ideas",
	Long:  "Review ideas interactively and mark them as reviewed/shortlisted/archived",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		return application.Triage(ctx)
	},
}
