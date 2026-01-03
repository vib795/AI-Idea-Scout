package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AnthropicAPIKey string         `mapstructure:"anthropic_api_key"`
	Model           string         `mapstructure:"model"`
	SafeMode        bool           `mapstructure:"safe_mode"`
	DatabasePath    string         `mapstructure:"database_path"`
	Sources         SourcesConfig  `mapstructure:"sources"`
	Analysis        AnalysisConfig `mapstructure:"analysis"`
	Output          OutputConfig   `mapstructure:"output"`
}

type SourcesConfig struct {
	HackerNews HackerNewsConfig `mapstructure:"hackernews"`
	Reddit     RedditConfig     `mapstructure:"reddit"`
	RSS        RSSConfig        `mapstructure:"rss"`
	GitHub     GitHubConfig     `mapstructure:"github"`
}

type HackerNewsConfig struct {
	Enabled         bool `mapstructure:"enabled"`
	MinScore        int  `mapstructure:"min_score"`
	MaxItems        int  `mapstructure:"max_items"`
	IncludeComments bool `mapstructure:"include_comments"`
}

type RedditConfig struct {
	Enabled    bool     `mapstructure:"enabled"`
	Subreddits []string `mapstructure:"subreddits"`
	Sort       string   `mapstructure:"sort"`
	Time       string   `mapstructure:"time"`
	MinScore   int      `mapstructure:"min_score"`
	MaxItems   int      `mapstructure:"max_items"`
}

type RSSConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Feeds   []string `mapstructure:"feeds"`
}

type GitHubConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	Token          string   `mapstructure:"token"`
	QueryTemplates []string `mapstructure:"query_templates"`
	MinStars       int      `mapstructure:"min_stars"`
	MaxItems       int      `mapstructure:"max_items"`
}

type AnalysisConfig struct {
	MaxAnalyze          int `mapstructure:"max_analyze"`
	BatchSize           int `mapstructure:"batch_size"`
	MinOverallScore     int `mapstructure:"min_overall_score_to_show"`
	CacheTTLHours       int `mapstructure:"cache_ttl_hours"`
	PromptVersion       string `mapstructure:"prompt_version"`
}

type OutputConfig struct {
	Color         bool   `mapstructure:"color"`
	DefaultFormat string `mapstructure:"default_format"`
}

func Load(cfgFile string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}

		// Search config in home directory with name ".ideascout" (without extension)
		configDir := filepath.Join(home, ".config", "ideascout")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variable overrides
	viper.SetEnvPrefix("IDEASCOUT")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found, run 'ideascout config init' to create one")
		}
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func setDefaults() {
	home, _ := os.UserHomeDir()

	viper.SetDefault("model", "claude-sonnet-4-5-20250929")
	viper.SetDefault("safe_mode", false)
	viper.SetDefault("database_path", filepath.Join(home, ".local", "share", "ideascout", "ideascout.db"))

	// HackerNews defaults
	viper.SetDefault("sources.hackernews.enabled", true)
	viper.SetDefault("sources.hackernews.min_score", 50)
	viper.SetDefault("sources.hackernews.max_items", 50)
	viper.SetDefault("sources.hackernews.include_comments", false)

	// Reddit defaults
	viper.SetDefault("sources.reddit.enabled", true)
	viper.SetDefault("sources.reddit.subreddits", []string{"LocalLLaMA", "MachineLearning", "ArtificialIntelligence", "LangChain"})
	viper.SetDefault("sources.reddit.sort", "top")
	viper.SetDefault("sources.reddit.time", "week")
	viper.SetDefault("sources.reddit.min_score", 20)
	viper.SetDefault("sources.reddit.max_items", 25)

	// RSS defaults
	viper.SetDefault("sources.rss.enabled", true)
	viper.SetDefault("sources.rss.feeds", []string{})

	// GitHub defaults
	viper.SetDefault("sources.github.enabled", false)
	viper.SetDefault("sources.github.min_stars", 100)
	viper.SetDefault("sources.github.max_items", 20)
	viper.SetDefault("sources.github.query_templates", []string{
		"topic:ai stars:>100 created:>{{.Since}}",
		"topic:agent stars:>100 created:>{{.Since}}",
	})

	// Analysis defaults
	viper.SetDefault("analysis.max_analyze", 100)
	viper.SetDefault("analysis.batch_size", 10)
	viper.SetDefault("analysis.min_overall_score_to_show", 5)
	viper.SetDefault("analysis.cache_ttl_hours", 72)
	viper.SetDefault("analysis.prompt_version", "v2-idea-extract-2026-01-03")

	// Output defaults
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.default_format", "text")
}

func (c *Config) Validate() error {
	if c.AnthropicAPIKey == "" {
		return fmt.Errorf("anthropic_api_key is required")
	}

	if c.Model == "" {
		return fmt.Errorf("model is required")
	}

	if c.DatabasePath == "" {
		return fmt.Errorf("database_path is required")
	}

	// Validate at least one source is enabled
	if !c.Sources.HackerNews.Enabled && !c.Sources.Reddit.Enabled &&
	   !c.Sources.RSS.Enabled && !c.Sources.GitHub.Enabled {
		return fmt.Errorf("at least one source must be enabled")
	}

	return nil
}

func DefaultConfigTemplate() string {
	return `# AI Idea Scout Configuration

# Anthropic API key (required)
# Get one at: https://console.anthropic.com/
anthropic_api_key: ""

# Claude model to use
model: "claude-sonnet-4-5-20250929"

# Safe mode: only allow sources with official APIs
safe_mode: false

# Database path
database_path: "~/.local/share/ideascout/ideascout.db"

# Source configurations
sources:
  hackernews:
    enabled: true
    min_score: 50
    max_items: 50
    include_comments: false

  reddit:
    enabled: true
    subreddits:
      - "LocalLLaMA"
      - "MachineLearning"
      - "ArtificialIntelligence"
      - "LangChain"
    sort: "top"          # top, hot, new
    time: "week"         # hour, day, week, month, year, all
    min_score: 20
    max_items: 25

  rss:
    enabled: true
    feeds:
      # Add your RSS feeds here
      # - "https://example.com/feed.xml"

  github:
    enabled: false       # Requires token
    token: ""            # GitHub personal access token
    min_stars: 100
    max_items: 20
    query_templates:
      - "topic:ai stars:>100 created:>{{.Since}}"
      - "topic:agent stars:>100 created:>{{.Since}}"

# Analysis settings
analysis:
  max_analyze: 100
  batch_size: 10
  min_overall_score_to_show: 5
  cache_ttl_hours: 72
  prompt_version: "v2-idea-extract-2026-01-03"

# Output settings
output:
  color: true
  default_format: "text"  # text, json, md
`
}

func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ideascout", "config.yaml"), nil
}

func EnsureConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "ideascout")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

func ExpandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// Helper to get cache TTL as duration
func (c *Config) CacheTTL() time.Duration {
	return time.Duration(c.Analysis.CacheTTLHours) * time.Hour
}
