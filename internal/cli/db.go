package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/db"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := cfg.DatabasePath
		if dbPath == "" {
			return fmt.Errorf("database_path not configured")
		}

		if err := db.Migrate(dbPath); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		fmt.Println("✓ Database migrations completed successfully")
		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
}
