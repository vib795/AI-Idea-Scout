package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var showFormat string

var showCmd = &cobra.Command{
	Use:   "show <idea-id>",
	Short: "Show detailed information about an idea",
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

		return application.Show(ctx, ideaID, showFormat)
	},
}

func init() {
	showCmd.Flags().StringVar(&showFormat, "format", "text", "Output format (text/json)")
}
