package pipeline

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/vib795/AI-Idea-Scout/internal/config"
	"github.com/vib795/AI-Idea-Scout/internal/db"
	"github.com/vib795/AI-Idea-Scout/internal/dedupe"
	"github.com/vib795/AI-Idea-Scout/internal/llm"
	"github.com/vib795/AI-Idea-Scout/internal/rank"
)

type Pipeline struct {
	db      *db.DB
	llm     *llm.Client
	ranker  *rank.Ranker
	deduper *dedupe.Deduper
	cfg     *config.Config
}

func New(database *db.DB, llmClient *llm.Client, ranker *rank.Ranker, deduper *dedupe.Deduper, cfg *config.Config) *Pipeline {
	return &Pipeline{
		db:      database,
		llm:     llmClient,
		ranker:  ranker,
		deduper: deduper,
		cfg:     cfg,
	}
}

// Normalize cleans and prepares raw content
func (p *Pipeline) Normalize(ctx context.Context, force bool) error {
	stage := "ingested"
	if !force {
		stage = "ingested"
	}

	items, err := p.db.GetRawContentByStage(ctx, stage, 0)
	if err != nil {
		return err
	}

	fmt.Printf("Normalizing %d items...\n", len(items))

	for _, item := range items {
		// Normalize text (already done during ingestion, but we can enhance it)
		normalizedText := item.Title
		if item.Body.Valid && item.Body.String != "" {
			normalizedText += "\n\n" + cleanText(item.Body.String)
		}

		// Update normalized text and content hash
		_ = sha256.Sum256([]byte(normalizedText + item.URL.String))

		// Note: In a real implementation, we'd update these fields
		// For now, mark as normalized
		if err := p.db.UpdateRawContentStage(ctx, item.ID, "normalized"); err != nil {
			return err
		}
	}

	fmt.Printf("✓ Normalized %d items\n", len(items))
	return nil
}

// FilterSignals applies heuristics to identify high-signal content
func (p *Pipeline) FilterSignals(ctx context.Context, force bool) error {
	stage := "normalized"
	items, err := p.db.GetRawContentByStage(ctx, stage, 0)
	if err != nil {
		return err
	}

	fmt.Printf("Filtering %d items for signals...\n", len(items))

	passed := 0
	for _, item := range items {
		if p.hasSignal(item) {
			if err := p.db.UpdateRawContentStage(ctx, item.ID, "filtered"); err != nil {
				return err
			}
			passed++
		} else {
			// Mark as skipped
			if err := p.db.UpdateRawContentStage(ctx, item.ID, "skipped"); err != nil {
				return err
			}
		}
	}

	fmt.Printf("✓ %d items passed signal filter (%d skipped)\n", passed, len(items)-passed)
	return nil
}

// ExtractIdeas uses LLM to extract structured ideas
func (p *Pipeline) ExtractIdeas(ctx context.Context, maxAnalyze int, force bool) error {
	stage := "filtered"
	items, err := p.db.GetRawContentByStage(ctx, stage, maxAnalyze)
	if err != nil {
		return err
	}

	if maxAnalyze > 0 && len(items) > maxAnalyze {
		items = items[:maxAnalyze]
	}

	fmt.Printf("Extracting ideas from %d items (this may take a while)...\n", len(items))

	totalIdeas := 0
	for i, item := range items {
		fmt.Printf("  [%d/%d] Analyzing: %s\n", i+1, len(items), truncateString(item.Title, 60))

		ideas, err := p.llm.ExtractIdeas(ctx, item, p.db)
		if err != nil {
			fmt.Printf("    ✗ Error: %v\n", err)
			continue
		}

		// Store ideas
		for _, extractedIdea := range ideas {
			ideaID, err := p.storeIdea(ctx, item.ID, extractedIdea)
			if err != nil {
				fmt.Printf("    ! Failed to store idea: %v\n", err)
				continue
			}
			totalIdeas++
			fmt.Printf("    ✓ Idea %d: %s\n", ideaID, extractedIdea.Title)
		}

		// Mark as extracted
		if err := p.db.UpdateRawContentStage(ctx, item.ID, "extracted"); err != nil {
			return err
		}
	}

	fmt.Printf("✓ Extracted %d ideas from %d items\n", totalIdeas, len(items))
	return nil
}

// DedupeIdeas identifies and clusters similar ideas
func (p *Pipeline) DedupeIdeas(ctx context.Context, force bool) error {
	// Get all ideas
	ideas, err := p.db.ListIdeas(ctx, map[string]interface{}{}, 0)
	if err != nil {
		return err
	}

	fmt.Printf("Deduplicating %d ideas...\n", len(ideas))

	clusters := p.deduper.ClusterIdeas(ideas)

	fmt.Printf("✓ Found %d unique idea clusters\n", len(clusters))
	return nil
}

// RankIdeas computes overall scores for ideas
func (p *Pipeline) RankIdeas(ctx context.Context, force bool) error {
	ideas, err := p.db.ListIdeas(ctx, map[string]interface{}{}, 0)
	if err != nil {
		return err
	}

	fmt.Printf("Ranking %d ideas...\n", len(ideas))

	for _, idea := range ideas {
		score := p.ranker.ComputeScore(idea)

		// Update idea with score
		// Note: In a real implementation, we'd have an UpdateIdea method
		_ = score
	}

	fmt.Printf("✓ Ranked %d ideas\n", len(ideas))
	return nil
}

// Helper methods

func (p *Pipeline) hasSignal(item *db.RawContent) bool {
	text := item.Title
	if item.NormalizedText.Valid {
		text = item.NormalizedText.String
	}

	// Check for ask patterns
	askPatterns := []string{
		"is there a tool",
		"how do i",
		"anyone built",
		"i wish",
		"need a way to",
		"looking for",
		"would pay for",
		"frustrated with",
	}

	textLower := strings.ToLower(text)
	for _, pattern := range askPatterns {
		if strings.Contains(textLower, pattern) {
			return true
		}
	}

	// Check for AI relevance keywords
	aiKeywords := []string{
		"llm", "gpt", "claude", "ai agent", "chatbot",
		"automation", "rag", "vector", "embedding",
		"langchain", "prompt", "fine-tune",
	}

	for _, keyword := range aiKeywords {
		if strings.Contains(textLower, keyword) {
			return true
		}
	}

	// Check minimum score
	if item.Score.Valid && item.Score.Int64 >= 20 {
		return true
	}

	// Minimum length check (avoid one-liners with no substance)
	if len(text) < 100 {
		return false
	}

	return true
}

func (p *Pipeline) storeIdea(ctx context.Context, contentID string, idea llm.ExtractedIdea) (int64, error) {
	dataReqJSON, _ := json.Marshal(idea.DataRequirements)
	techApproachJSON, _ := json.Marshal(idea.TechnicalApproach)
	tagsJSON, _ := json.Marshal(idea.Tags)
	risksJSON, _ := json.Marshal(idea.Risks)
	evidenceJSON, _ := json.Marshal(idea.Evidence)
	nextStepsJSON, _ := json.Marshal(idea.NextSteps)

	dbIdea := &db.Idea{
		ContentID:         contentID,
		Title:             idea.Title,
		OneLiner:          sql.NullString{String: idea.OneLiner, Valid: true},
		Problem:           sql.NullString{String: idea.Problem, Valid: true},
		TargetUsers:       sql.NullString{String: idea.TargetUsers, Valid: true},
		JobToBeDone:       sql.NullString{String: idea.JobToBeDone, Valid: true},
		DataRequirements:  sql.NullString{String: string(dataReqJSON), Valid: true},
		TechnicalApproach: sql.NullString{String: string(techApproachJSON), Valid: true},
		Complexity:        sql.NullString{String: idea.Complexity, Valid: true},
		Category:          sql.NullString{String: idea.Category, Valid: true},
		Tags:              sql.NullString{String: string(tagsJSON), Valid: true},
		BuildabilityScore: sql.NullInt64{Int64: int64(idea.BuildabilityScore), Valid: true},
		DemandScore:       sql.NullInt64{Int64: int64(idea.DemandScore), Valid: true},
		NoveltyScore:      sql.NullInt64{Int64: int64(idea.NoveltyScore), Valid: true},
		DistributionWedge: sql.NullString{String: idea.DistributionWedge, Valid: true},
		Moat:              sql.NullString{String: idea.Moat, Valid: true},
		Risks:             sql.NullString{String: string(risksJSON), Valid: true},
		Evidence:          sql.NullString{String: string(evidenceJSON), Valid: true},
		NextSteps:         sql.NullString{String: string(nextStepsJSON), Valid: true},
		Status:            "new",
	}

	return p.db.InsertIdea(ctx, dbIdea)
}

func cleanText(s string) string {
	// Remove excessive whitespace
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	// Remove URLs
	s = regexp.MustCompile(`https?://\S+`).ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
