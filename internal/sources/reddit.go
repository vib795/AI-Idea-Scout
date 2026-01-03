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
	"github.com/vib795/AI-Idea-Scout/internal/fetch"
)

type RedditSource struct {
	cfg    *config.Config
	client *fetch.HTTPClient
}

func NewRedditSource(cfg *config.Config) *RedditSource {
	return &RedditSource{
		cfg: cfg,
		client: fetch.NewHTTPClient(
			"ideascout/1.0 (AI Idea Scout; +https://github.com/vib795/AI-Idea-Scout)",
			0.3, // Reddit asks for < 60 requests per minute = 1 per second, so we use 0.3 to be safe
			true,
		),
	}
}

func (s *RedditSource) Name() string {
	return "reddit"
}

func (s *RedditSource) Compliance() ComplianceInfo {
	return ComplianceInfo{
		OfficialAPI: false,
		PublicFeed:  true,
		TOSURL:      "https://www.redditinc.com/policies/data-api-terms",
		Notes:       "Public JSON endpoints, no authentication. Respect rate limits.",
	}
}

func (s *RedditSource) RatePolicy() RatePolicy {
	return RatePolicy{
		RequestsPerSecond: 0.3,
		BurstSize:         1,
	}
}

func (s *RedditSource) Fetch(ctx context.Context, cfg SourceConfig) ([]RawItem, error) {
	var allItems []RawItem

	for _, subreddit := range s.cfg.Sources.Reddit.Subreddits {
		items, err := s.fetchSubreddit(ctx, subreddit, cfg)
		if err != nil {
			// Log error but continue with other subreddits
			continue
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

func (s *RedditSource) fetchSubreddit(ctx context.Context, subreddit string, cfg SourceConfig) ([]RawItem, error) {
	sort := s.cfg.Sources.Reddit.Sort
	timeFilter := s.cfg.Sources.Reddit.Time

	url := fmt.Sprintf("https://www.reddit.com/r/%s/%s.json?t=%s&limit=100", subreddit, sort, timeFilter)

	data, err := s.client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subreddit %s: %w", subreddit, err)
	}

	var response struct {
		Data struct {
			Children []struct {
				Data struct {
					ID          string  `json:"id"`
					Title       string  `json:"title"`
					Selftext    string  `json:"selftext"`
					Author      string  `json:"author"`
					Score       int     `json:"score"`
					NumComments int     `json:"num_comments"`
					Created     float64 `json:"created_utc"`
					Permalink   string  `json:"permalink"`
					URL         string  `json:"url"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse reddit response: %w", err)
	}

	var items []RawItem
	for _, child := range response.Data.Children {
		post := child.Data

		// Filter by score
		if post.Score < s.cfg.Sources.Reddit.MinScore {
			continue
		}

		createdAt := time.Unix(int64(post.Created), 0)

		// Filter by time if specified
		if cfg.Since > 0 {
			itemAge := time.Since(createdAt)
			if itemAge > cfg.Since {
				continue
			}
		}

		// Build permalink
		permalink := post.Permalink
		if !contains(permalink, "reddit.com") {
			permalink = "https://www.reddit.com" + permalink
		}

		items = append(items, RawItem{
			ID:            post.ID,
			URL:           permalink,
			Title:         post.Title,
			Body:          post.Selftext,
			Author:        post.Author,
			Score:         post.Score,
			CommentsCount: post.NumComments,
			CreatedAt:     createdAt,
			RawJSON:       string(data),
		})

		// Apply limit
		if cfg.Limit > 0 && len(items) >= cfg.Limit {
			break
		}
	}

	return items, nil
}

func (s *RedditSource) Normalize(item RawItem) (*db.RawContent, error) {
	// Build normalized text
	normalizedText := item.Title
	if item.Body != "" {
		normalizedText += "\n\n" + item.Body
	}

	// Compute content hash
	hash := sha256.Sum256([]byte(normalizedText + item.URL))
	contentHash := fmt.Sprintf("%x", hash)

	return &db.RawContent{
		ID:             uuid.New().String(),
		Source:         "reddit",
		SourceID:       item.ID,
		URL:            sql.NullString{String: item.URL, Valid: true},
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
