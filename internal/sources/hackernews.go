package sources

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vib795/AI-Idea-Scout/internal/config"
	"github.com/vib795/AI-Idea-Scout/internal/db"
	"github.com/vib795/AI-Idea-Scout/internal/httpclient"
)

const (
	hnAPIBase      = "https://hacker-news.firebaseio.com/v0"
	hnItemURL      = hnAPIBase + "/item/%d.json"
	hnTopStoriesURL = hnAPIBase + "/topstories.json"
	hnBestStoriesURL = hnAPIBase + "/beststories.json"
)

type HackerNewsSource struct {
	cfg    *config.Config
	client *httpclient.HTTPClient
}

func NewHackerNewsSource(cfg *config.Config) *HackerNewsSource {
	return &HackerNewsSource{
		cfg: cfg,
		client: httpclient.NewHTTPClient(
			"ideascout/1.0 (AI Idea Scout; +https://github.com/vib795/AI-Idea-Scout)",
			1.0, // 1 req/sec for HN
			true,
		),
	}
}

func (s *HackerNewsSource) Name() string {
	return "hackernews"
}

func (s *HackerNewsSource) Compliance() ComplianceInfo {
	return ComplianceInfo{
		OfficialAPI: true,
		PublicFeed:  true,
		TOSURL:      "https://github.com/HackerNews/API",
		Notes:       "Official Firebase API, no authentication required",
	}
}

func (s *HackerNewsSource) RatePolicy() RatePolicy {
	return RatePolicy{
		RequestsPerSecond: 1.0,
		BurstSize:         5,
	}
}

func (s *HackerNewsSource) Fetch(ctx context.Context, cfg SourceConfig) ([]RawItem, error) {
	// Get top stories IDs
	data, err := s.client.Get(ctx, hnBestStoriesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch best stories: %w", err)
	}

	var storyIDs []int
	if err := json.Unmarshal(data, &storyIDs); err != nil {
		return nil, fmt.Errorf("failed to parse story IDs: %w", err)
	}

	// Determine how many to fetch
	limit := s.cfg.Sources.HackerNews.MaxItems
	if cfg.Limit > 0 && cfg.Limit < limit {
		limit = cfg.Limit
	}
	if limit > len(storyIDs) {
		limit = len(storyIDs)
	}

	// Fetch individual stories
	var items []RawItem
	for i := 0; i < limit; i++ {
		item, err := s.fetchItem(ctx, storyIDs[i])
		if err != nil {
			// Log error but continue
			continue
		}

		// Filter by score
		if item.Score < s.cfg.Sources.HackerNews.MinScore {
			continue
		}

		// Filter by time if specified
		if cfg.Since > 0 {
			itemAge := time.Since(item.CreatedAt)
			if itemAge > cfg.Since {
				continue
			}
		}

		items = append(items, *item)
	}

	return items, nil
}

func (s *HackerNewsSource) fetchItem(ctx context.Context, id int) (*RawItem, error) {
	url := fmt.Sprintf(hnItemURL, id)
	data, err := s.client.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	var hnItem struct {
		ID          int    `json:"id"`
		Type        string `json:"type"`
		By          string `json:"by"`
		Time        int64  `json:"time"`
		Text        string `json:"text"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		Score       int    `json:"score"`
		Descendants int    `json:"descendants"`
	}

	if err := json.Unmarshal(data, &hnItem); err != nil {
		return nil, err
	}

	// Skip non-story items
	if hnItem.Type != "story" {
		return nil, fmt.Errorf("not a story")
	}

	return &RawItem{
		ID:            fmt.Sprintf("%d", hnItem.ID),
		URL:           hnItem.URL,
		Title:         hnItem.Title,
		Body:          hnItem.Text,
		Author:        hnItem.By,
		Score:         hnItem.Score,
		CommentsCount: hnItem.Descendants,
		CreatedAt:     time.Unix(hnItem.Time, 0),
		RawJSON:       string(data),
	}, nil
}

func (s *HackerNewsSource) Normalize(item RawItem) (*db.RawContent, error) {
	// Build normalized text
	normalizedText := item.Title
	if item.Body != "" {
		normalizedText += "\n\n" + item.Body
	}

	// Compute content hash
	hash := sha256.Sum256([]byte(normalizedText + item.URL))
	contentHash := fmt.Sprintf("%x", hash)

	// Build HN URL if no external URL
	url := item.URL
	if url == "" {
		url = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", item.ID)
	}

	return &db.RawContent{
		ID:             uuid.New().String(),
		Source:         "hackernews",
		SourceID:       item.ID,
		URL:            sql.NullString{String: url, Valid: true},
		Title:          item.Title,
		Body:           sql.NullString{String: item.Body, Valid: item.Body != ""},
		Author:         sql.NullString{String: item.Author, Valid: true},
		Score:          sql.NullInt64{Int64: int64(item.Score), Valid: true},
		CommentsCount:  sql.NullInt64{Int64: int64(item.CommentsCount), Valid: true},
		CreatedAt:      sql.NullTime{Time: item.CreatedAt, Valid: true},
		FetchedAt:      time.Now(),
		RawJSON:        sql.NullString{String: item.RawJSON, Valid: true},
		NormalizedText: sql.NullString{String: normalizedText, Valid: true},
		ContentHash:    contentHash,
		ProcessedStage: "ingested",
	}, nil
}
