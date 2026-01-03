package sources

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vib795/AI-Idea-Scout/internal/config"
	"github.com/vib795/AI-Idea-Scout/internal/db"
	"github.com/vib795/AI-Idea-Scout/internal/httpclient"
)

const (
	githubAPIBase = "https://api.github.com"
	githubSearchURL = githubAPIBase + "/search/repositories"
)

type GitHubSource struct {
	cfg    *config.Config
	client *httpclient.HTTPClient
}

func NewGitHubSource(cfg *config.Config) *GitHubSource {
	client := httpclient.NewHTTPClient(
		"ideascout/1.0 (AI Idea Scout; +https://github.com/vib795/AI-Idea-Scout)",
		0.5, // 10 requests per minute for unauthenticated
		true,
	)
	return &GitHubSource{
		cfg:    cfg,
		client: client,
	}
}

func (s *GitHubSource) Name() string {
	return "github"
}

func (s *GitHubSource) Compliance() ComplianceInfo {
	return ComplianceInfo{
		OfficialAPI: true,
		PublicFeed:  true,
		TOSURL:      "https://docs.github.com/en/site-policy/github-terms/github-terms-of-service",
		Notes:       "Official GitHub Search API. Token recommended for higher rate limits.",
	}
}

func (s *GitHubSource) RatePolicy() RatePolicy {
	return RatePolicy{
		RequestsPerSecond: 0.5,
		BurstSize:         2,
	}
}

func (s *GitHubSource) Fetch(ctx context.Context, cfg SourceConfig) ([]RawItem, error) {
	var allItems []RawItem

	// Calculate date range
	since := time.Now().AddDate(0, 0, -7) // Default to 7 days
	if cfg.Since > 0 {
		since = time.Now().Add(-cfg.Since)
	}

	for _, queryTemplate := range s.cfg.Sources.GitHub.QueryTemplates {
		// Replace template variables
		query := strings.ReplaceAll(queryTemplate, "{{.Since}}", since.Format("2006-01-02"))

		items, err := s.searchRepositories(ctx, query, cfg.Limit)
		if err != nil {
			// Log error but continue
			continue
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

func (s *GitHubSource) searchRepositories(ctx context.Context, query string, limit int) ([]RawItem, error) {
	// Build search URL
	url := fmt.Sprintf("%s?q=%s&sort=stars&order=desc&per_page=30", githubSearchURL, query)

	data, err := s.client.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("github search failed: %w", err)
	}

	var response struct {
		Items []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
			Stars       int    `json:"stargazers_count"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
			Owner       struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse github response: %w", err)
	}

	var items []RawItem
	for _, repo := range response.Items {
		// Filter by stars
		if repo.Stars < s.cfg.Sources.GitHub.MinStars {
			continue
		}

		createdAt, _ := time.Parse(time.RFC3339, repo.CreatedAt)

		items = append(items, RawItem{
			ID:        fmt.Sprintf("%d", repo.ID),
			URL:       repo.HTMLURL,
			Title:     repo.FullName,
			Body:      repo.Description,
			Author:    repo.Owner.Login,
			Score:     repo.Stars,
			CreatedAt: createdAt,
			RawJSON:   string(data),
		})

		// Apply limit
		if limit > 0 && len(items) >= limit {
			break
		}
	}

	return items, nil
}

func (s *GitHubSource) Normalize(item RawItem) (*db.RawContent, error) {
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
		Source:         "github",
		SourceID:       item.ID,
		URL:            sql.NullString{String: item.URL, Valid: true},
		Title:          item.Title,
		Body:           sql.NullString{String: item.Body, Valid: item.Body != ""},
		Author:         sql.NullString{String: item.Author, Valid: true},
		Score:          sql.NullInt64{Int64: int64(item.Score), Valid: true},
		CreatedAt:      sql.NullTime{Time: item.CreatedAt, Valid: !item.CreatedAt.IsZero()},
		FetchedAt:      time.Now(),
		RawJSON:        sql.NullString{String: item.RawJSON, Valid: true},
		NormalizedText: sql.NullString{String: normalizedText, Valid: true},
		ContentHash:    contentHash,
		ProcessedStage: "ingested",
	}, nil
}
