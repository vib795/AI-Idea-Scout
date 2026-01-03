package fetch

import (
	"context"
	"time"

	"github.com/vib795/AI-Idea-Scout/internal/config"
	"github.com/vib795/AI-Idea-Scout/internal/db"
	"github.com/vib795/AI-Idea-Scout/internal/sources"
)

type Fetcher struct {
	cfg *config.Config
}

func NewFetcher(cfg *config.Config) *Fetcher {
	return &Fetcher{cfg: cfg}
}

func (f *Fetcher) FetchSource(ctx context.Context, source sources.Source, limit int, since time.Duration) ([]*db.RawContent, error) {
	// Build source-specific config
	srcCfg := sources.SourceConfig{
		Limit: limit,
		Since: since,
	}

	// Fetch items from source
	items, err := source.Fetch(ctx, srcCfg)
	if err != nil {
		return nil, err
	}

	// Convert to RawContent
	var rawContents []*db.RawContent
	for _, item := range items {
		rc, err := source.Normalize(item)
		if err != nil {
			// Log error but continue
			continue
		}
		rawContents = append(rawContents, rc)
	}

	return rawContents, nil
}
