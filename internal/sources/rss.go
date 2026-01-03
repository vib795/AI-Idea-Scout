package sources

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vib795/AI-Idea-Scout/internal/config"
	"github.com/vib795/AI-Idea-Scout/internal/db"
	"github.com/vib795/AI-Idea-Scout/internal/fetch"
)

type RSSSource struct {
	cfg    *config.Config
	client *fetch.HTTPClient
}

func NewRSSSource(cfg *config.Config) *RSSSource {
	return &RSSSource{
		cfg: cfg,
		client: fetch.NewHTTPClient(
			"ideascout/1.0 (AI Idea Scout; +https://github.com/vib795/AI-Idea-Scout)",
			0.5, // 0.5 req/sec for RSS feeds
			true,
		),
	}
}

func (s *RSSSource) Name() string {
	return "rss"
}

func (s *RSSSource) Compliance() ComplianceInfo {
	return ComplianceInfo{
		OfficialAPI: false,
		PublicFeed:  true,
		TOSURL:      "",
		Notes:       "Public RSS/Atom feeds, no authentication required",
	}
}

func (s *RSSSource) RatePolicy() RatePolicy {
	return RatePolicy{
		RequestsPerSecond: 0.5,
		BurstSize:         2,
	}
}

func (s *RSSSource) Fetch(ctx context.Context, cfg SourceConfig) ([]RawItem, error) {
	var allItems []RawItem

	for _, feedURL := range s.cfg.Sources.RSS.Feeds {
		items, err := s.fetchFeed(ctx, feedURL, cfg)
		if err != nil {
			// Log error but continue with other feeds
			continue
		}
		allItems = append(allItems, items...)
	}

	return allItems, nil
}

func (s *RSSSource) fetchFeed(ctx context.Context, feedURL string, cfg SourceConfig) ([]RawItem, error) {
	data, err := s.client.Get(ctx, feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}

	// Try to parse as RSS or Atom
	feed, err := s.parseFeed(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	var items []RawItem
	for _, entry := range feed.Entries {
		// Filter by time if specified
		if cfg.Since > 0 && !entry.Published.IsZero() {
			itemAge := time.Since(entry.Published)
			if itemAge > cfg.Since {
				continue
			}
		}

		items = append(items, RawItem{
			ID:        entry.ID,
			URL:       entry.Link,
			Title:     entry.Title,
			Body:      entry.Content,
			Author:    entry.Author,
			CreatedAt: entry.Published,
			RawJSON:   "", // RSS doesn't have JSON
		})

		// Apply limit
		if cfg.Limit > 0 && len(items) >= cfg.Limit {
			break
		}
	}

	return items, nil
}

type Feed struct {
	Entries []FeedEntry
}

type FeedEntry struct {
	ID        string
	Title     string
	Link      string
	Content   string
	Author    string
	Published time.Time
}

func (s *RSSSource) parseFeed(data []byte) (*Feed, error) {
	// Try RSS 2.0 first
	var rss struct {
		Channel struct {
			Items []struct {
				GUID        string `xml:"guid"`
				Title       string `xml:"title"`
				Link        string `xml:"link"`
				Description string `xml:"description"`
				Author      string `xml:"author"`
				PubDate     string `xml:"pubDate"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	if err := xml.Unmarshal(data, &rss); err == nil && len(rss.Channel.Items) > 0 {
		feed := &Feed{}
		for _, item := range rss.Channel.Items {
			published, _ := parseDate(item.PubDate)
			feed.Entries = append(feed.Entries, FeedEntry{
				ID:        item.GUID,
				Title:     item.Title,
				Link:      item.Link,
				Content:   stripHTML(item.Description),
				Author:    item.Author,
				Published: published,
			})
		}
		return feed, nil
	}

	// Try Atom
	var atom struct {
		Entries []struct {
			ID      string `xml:"id"`
			Title   string `xml:"title"`
			Link    struct {
				Href string `xml:"href,attr"`
			} `xml:"link"`
			Summary string `xml:"summary"`
			Content string `xml:"content"`
			Author  struct {
				Name string `xml:"name"`
			} `xml:"author"`
			Published string `xml:"published"`
		} `xml:"entry"`
	}

	if err := xml.Unmarshal(data, &atom); err == nil && len(atom.Entries) > 0 {
		feed := &Feed{}
		for _, entry := range atom.Entries {
			published, _ := parseDate(entry.Published)
			content := entry.Content
			if content == "" {
				content = entry.Summary
			}
			feed.Entries = append(feed.Entries, FeedEntry{
				ID:        entry.ID,
				Title:     entry.Title,
				Link:      entry.Link.Href,
				Content:   stripHTML(content),
				Author:    entry.Author.Name,
				Published: published,
			})
		}
		return feed, nil
	}

	return nil, fmt.Errorf("unknown feed format")
}

func (s *RSSSource) Normalize(item RawItem) (*db.RawContent, error) {
	// Build normalized text
	normalizedText := item.Title
	if item.Body != "" {
		normalizedText += "\n\n" + item.Body
	}

	// Compute content hash
	hash := sha256.Sum256([]byte(normalizedText + item.URL))
	contentHash := fmt.Sprintf("%x", hash)

	// Generate stable ID if missing
	sourceID := item.ID
	if sourceID == "" {
		sourceID = contentHash[:16]
	}

	return &db.RawContent{
		ID:             uuid.New().String(),
		Source:         "rss",
		SourceID:       sourceID,
		URL:            sql.NullString{String: item.URL, Valid: item.URL != ""},
		Title:          item.Title,
		Body:           sql.NullString{String: item.Body, Valid: item.Body != ""},
		Author:         sql.NullString{String: item.Author, Valid: item.Author != ""},
		FetchedAt:      time.Now(),
		CreatedAt:      sql.NullTime{Time: item.CreatedAt, Valid: !item.CreatedAt.IsZero()},
		NormalizedText: sql.NullString{String: normalizedText, Valid: true},
		ContentHash:    contentHash,
		ProcessedStage: "ingested",
	}, nil
}

// Helper functions

func stripHTML(s string) string {
	// Decode HTML entities
	s = html.UnescapeString(s)

	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")

	// Clean up whitespace
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")

	return s
}

func parseDate(s string) (time.Time, error) {
	// Try common date formats
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"Mon, 02 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}
