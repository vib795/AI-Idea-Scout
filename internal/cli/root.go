package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vib795/AI-Idea-Scout/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
	version string
	commit  string
	date    string
)

var rootCmd = &cobra.Command{
	Use:   "ideascout",
	Short: "AI Idea Scout - Monitor and extract buildable AI/agent product ideas",
	Long: `AI Idea Scout monitors high-signal public sources (HackerNews, Reddit, RSS, GitHub)
and extracts buildable AI/agent product ideas using Claude AI.

It provides:
- Automated fetching from multiple sources with rate limiting and caching
- AI-powered idea extraction with structured analysis
- Smart deduplication and ranking
- Deep-dive MVP specifications
- Interactive triage workflow`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config loading for commands that don't need it
		if cmd.Name() == "version" || cmd.Name() == "help" {
			return nil
		}

		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			// For init command, we allow missing config
			if cmd.Name() == "init" {
				return nil
			}
			return fmt.Errorf("failed to load config: %w", err)
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/ideascout/config.yaml)")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(triageCmd)
	rootCmd.AddCommand(digestCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(deepdiveCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(statsCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ideascout %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}
