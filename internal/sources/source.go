package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/vib795/AI-Idea-Scout/internal/config"
	"github.com/vib795/AI-Idea-Scout/internal/db"
)

// Source represents a content source (HN, Reddit, RSS, GitHub)
type Source interface {
	Name() string
	Compliance() ComplianceInfo
	RatePolicy() RatePolicy
	Fetch(ctx context.Context, cfg SourceConfig) ([]RawItem, error)
	Normalize(item RawItem) (*db.RawContent, error)
}

// ComplianceInfo describes compliance characteristics of a source
type ComplianceInfo struct {
	OfficialAPI bool
	PublicFeed  bool
	TOSURL      string
	Notes       string
}

// RatePolicy describes rate limiting requirements
type RatePolicy struct {
	RequestsPerSecond float64
	BurstSize         int
}

// SourceConfig contains fetch configuration
type SourceConfig struct {
	Limit int
	Since time.Duration
}

// RawItem is the generic item returned by a source
type RawItem struct {
	ID            string
	URL           string
	Title         string
	Body          string
	Author        string
	Score         int
	CommentsCount int
	CreatedAt     time.Time
	RawJSON       string
}

// EnabledSources returns list of enabled source names
func EnabledSources(cfg *config.Config) []string {
	var enabled []string
	if cfg.Sources.HackerNews.Enabled {
		enabled = append(enabled, "hackernews")
	}
	if cfg.Sources.Reddit.Enabled {
		enabled = append(enabled, "reddit")
	}
	if cfg.Sources.RSS.Enabled {
		enabled = append(enabled, "rss")
	}
	if cfg.Sources.GitHub.Enabled {
		enabled = append(enabled, "github")
	}
	return enabled
}

// GetSource creates a source instance by name
func GetSource(name string, cfg *config.Config) (Source, error) {
	switch name {
	case "hackernews":
		return NewHackerNewsSource(cfg), nil
	case "reddit":
		return NewRedditSource(cfg), nil
	case "rss":
		return NewRSSSource(cfg), nil
	case "github":
		return NewGitHubSource(cfg), nil
	default:
		return nil, fmt.Errorf("unknown source: %s", name)
	}
}
