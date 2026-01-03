package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/config"
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
		// Load config just for this command
		var dbPath string
		if cfg != nil {
			dbPath = cfg.DatabasePath
		} else {
			// Load config without validation
			tempCfg, err := config.Load(cfgFile)
			if err != nil {
				// Use default path if config not available
				home, _ := os.UserHomeDir()
				dbPath = filepath.Join(home, ".local", "share", "ideascout", "ideascout.db")
			} else {
				dbPath = tempCfg.DatabasePath
			}
		}

		dbPath = config.ExpandPath(dbPath)
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
