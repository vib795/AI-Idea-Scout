package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  "Create a default configuration file with example values",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetDefaultConfigPath()
		if err != nil {
			return fmt.Errorf("cannot determine config path: %w", err)
		}

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			overwrite, _ := cmd.Flags().GetBool("force")
			if !overwrite {
				return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
			}
		}

		// Ensure config directory exists
		if _, err := config.EnsureConfigDir(); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Write config template
		if err := os.WriteFile(configPath, []byte(config.DefaultConfigTemplate()), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Printf("✓ Created config file at: %s\n\n", configPath)
		fmt.Println("Next steps:")
		fmt.Println("1. Edit the config file and add your Anthropic API key")
		fmt.Println("2. Run 'ideascout db migrate' to initialize the database")
		fmt.Println("3. Run 'ideascout fetch' to start collecting ideas")

		return nil
	},
}

func init() {
	initCmd.Flags().Bool("force", false, "Overwrite existing config file")
	configCmd.AddCommand(initCmd)
}
