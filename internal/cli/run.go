package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var (
	runMaxAnalyze int
	runForce      bool
	runSafe       bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the complete pipeline",
	Long:  "Fetch, normalize, filter, extract ideas, deduplicate, and rank",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		opts := app.RunOptions{
			MaxAnalyze: runMaxAnalyze,
			Force:      runForce,
			SafeMode:   runSafe,
		}

		return application.Run(ctx, opts)
	},
}

func init() {
	runCmd.Flags().IntVar(&runMaxAnalyze, "max-analyze", 0, "Limit number of items to analyze with LLM")
	runCmd.Flags().BoolVar(&runForce, "force", false, "Force re-run all stages")
	runCmd.Flags().BoolVar(&runSafe, "safe", false, "Enable safe mode (only official APIs)")
}
