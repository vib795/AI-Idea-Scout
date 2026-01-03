package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var deepdiveOutput string

var deepdiveCmd = &cobra.Command{
	Use:   "deepdive <idea-id>",
	Short: "Generate a deep-dive MVP specification for an idea",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		ideaID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid idea ID: %w", err)
		}

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		return application.DeepDive(ctx, ideaID, deepdiveOutput)
	},
}

func init() {
	deepdiveCmd.Flags().StringVar(&deepdiveOutput, "output", "", "Output file path")
	deepdiveCmd.MarkFlagRequired("output")
}
