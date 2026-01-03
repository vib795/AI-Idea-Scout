package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/app"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export ideas to file",
	Long:  "Export all ideas or filtered ideas to markdown or JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		application, err := app.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		defer application.Close()

		opts := app.ExportOptions{
			Format: exportFormat,
			Output: exportOutput,
		}

		return application.Export(ctx, opts)
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "md", "Output format (md/json)")
	exportCmd.Flags().StringVar(&exportOutput, "output", "", "Output file path")
	exportCmd.MarkFlagRequired("output")
}
