package app

import (
	"context"
	"fmt"
	"time"

	"github.com/vib795/AI-Idea-Scout/internal/config"
	"github.com/vib795/AI-Idea-Scout/internal/db"
	"github.com/vib795/AI-Idea-Scout/internal/dedupe"
	"github.com/vib795/AI-Idea-Scout/internal/fetch"
	"github.com/vib795/AI-Idea-Scout/internal/llm"
	"github.com/vib795/AI-Idea-Scout/internal/output"
	"github.com/vib795/AI-Idea-Scout/internal/pipeline"
	"github.com/vib795/AI-Idea-Scout/internal/rank"
	"github.com/vib795/AI-Idea-Scout/internal/sources"
	"github.com/vib795/AI-Idea-Scout/internal/triage"
)

type App struct {
	cfg      *config.Config
	db       *db.DB
	fetcher  *fetch.Fetcher
	pipeline *pipeline.Pipeline
	llm      *llm.Client
	ranker   *rank.Ranker
	deduper  *dedupe.Deduper
	output   *output.Formatter
	triage   *triage.Manager
}

func New(cfg *config.Config) (*App, error) {
	// Expand database path
	dbPath := config.ExpandPath(cfg.DatabasePath)

	// Open database
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize LLM client
	llmClient := llm.NewClient(cfg.AnthropicAPIKey, cfg.Model, cfg.Analysis.PromptVersion)

	// Initialize components
	fetcher := fetch.NewFetcher(cfg)
	ranker := rank.NewRanker()
	deduper := dedupe.NewDeduper()
	formatter := output.NewFormatter(cfg.Output.Color)
	triageManager := triage.NewManager(database)

	// Initialize pipeline
	pipelineEngine := pipeline.New(database, llmClient, ranker, deduper, cfg)

	return &App{
		cfg:      cfg,
		db:       database,
		fetcher:  fetcher,
		pipeline: pipelineEngine,
		llm:      llmClient,
		ranker:   ranker,
		deduper:  deduper,
		output:   formatter,
		triage:   triageManager,
	}, nil
}

func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Command implementations

type FetchOptions struct {
	Sources []string
	Limit   int
	Since   time.Duration
	Force   bool
}

func (a *App) Fetch(ctx context.Context, opts FetchOptions) error {
	// Determine which sources to fetch from
	sourceNames := opts.Sources
	if len(sourceNames) == 0 {
		sourceNames = sources.EnabledSources(a.cfg)
	}

	fmt.Printf("Fetching from %d source(s)...\n", len(sourceNames))

	for _, name := range sourceNames {
		fmt.Printf("\n[%s] Fetching...\n", name)
		source, err := sources.GetSource(name, a.cfg)
		if err != nil {
			fmt.Printf("  ✗ Error: %v\n", err)
			continue
		}

		items, err := a.fetcher.FetchSource(ctx, source, opts.Limit, opts.Since)
		if err != nil {
			fmt.Printf("  ✗ Error: %v\n", err)
			continue
		}

		// Store items in database
		stored := 0
		for _, item := range items {
			if err := a.db.InsertRawContent(ctx, item); err != nil {
				fmt.Printf("  ! Warning: failed to store item: %v\n", err)
			} else {
				stored++
			}
		}

		fmt.Printf("  ✓ Fetched %d items, stored %d\n", len(items), stored)
	}

	return nil
}

type RunOptions struct {
	MaxAnalyze int
	Force      bool
	SafeMode   bool
}

func (a *App) Run(ctx context.Context, opts RunOptions) error {
	fmt.Println("Running complete pipeline...")

	// Stage 1: Fetch
	fmt.Println("\n[1/6] Fetching content...")
	fetchOpts := FetchOptions{Force: opts.Force}
	if err := a.Fetch(ctx, fetchOpts); err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	// Stage 2-6: Run pipeline
	fmt.Println("\n[2/6] Normalizing content...")
	if err := a.pipeline.Normalize(ctx, opts.Force); err != nil {
		return fmt.Errorf("normalize failed: %w", err)
	}

	fmt.Println("\n[3/6] Filtering signals...")
	if err := a.pipeline.FilterSignals(ctx, opts.Force); err != nil {
		return fmt.Errorf("filter failed: %w", err)
	}

	fmt.Println("\n[4/6] Extracting ideas...")
	if err := a.pipeline.ExtractIdeas(ctx, opts.MaxAnalyze, opts.Force); err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	fmt.Println("\n[5/6] Deduplicating ideas...")
	if err := a.pipeline.DedupeIdeas(ctx, opts.Force); err != nil {
		return fmt.Errorf("dedupe failed: %w", err)
	}

	fmt.Println("\n[6/6] Ranking ideas...")
	if err := a.pipeline.RankIdeas(ctx, opts.Force); err != nil {
		return fmt.Errorf("rank failed: %w", err)
	}

	fmt.Println("\n✓ Pipeline completed successfully")
	return nil
}

type ListOptions struct {
	Top        int
	Category   string
	Complexity string
	Status     string
	Since      string
	Format     string
}

func (a *App) List(ctx context.Context, opts ListOptions) error {
	filters := make(map[string]interface{})
	if opts.Status != "" {
		filters["status"] = opts.Status
	}
	if opts.Category != "" {
		filters["category"] = opts.Category
	}
	if opts.Complexity != "" {
		filters["complexity"] = opts.Complexity
	}

	ideas, err := a.db.ListIdeas(ctx, filters, opts.Top)
	if err != nil {
		return fmt.Errorf("failed to list ideas: %w", err)
	}

	return a.output.FormatIdeasList(ideas, opts.Format)
}

func (a *App) Show(ctx context.Context, ideaID int64, format string) error {
	idea, err := a.db.GetIdeaByID(ctx, ideaID)
	if err != nil {
		return fmt.Errorf("failed to get idea: %w", err)
	}

	return a.output.FormatIdea(idea, format)
}

func (a *App) Search(ctx context.Context, query string, format string) error {
	results, err := a.db.SearchContent(ctx, query, 50)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	return a.output.FormatSearchResults(results, format)
}

func (a *App) Triage(ctx context.Context) error {
	return a.triage.Run(ctx)
}

type DigestOptions struct {
	Since  string
	Top    int
	Format string
}

func (a *App) Digest(ctx context.Context, opts DigestOptions) error {
	// Get top ideas
	ideas, err := a.db.ListIdeas(ctx, map[string]interface{}{}, opts.Top)
	if err != nil {
		return fmt.Errorf("failed to get ideas: %w", err)
	}

	return a.output.FormatDigest(ideas, opts.Since, opts.Format)
}

type ExportOptions struct {
	Format string
	Output string
}

func (a *App) Export(ctx context.Context, opts ExportOptions) error {
	ideas, err := a.db.ListIdeas(ctx, map[string]interface{}{}, 0)
	if err != nil {
		return fmt.Errorf("failed to get ideas: %w", err)
	}

	return a.output.ExportToFile(ideas, opts.Output, opts.Format)
}

func (a *App) DeepDive(ctx context.Context, ideaID int64, outputPath string) error {
	idea, err := a.db.GetIdeaByID(ctx, ideaID)
	if err != nil {
		return fmt.Errorf("failed to get idea: %w", err)
	}

	content, err := a.llm.GenerateDeepDive(ctx, idea)
	if err != nil {
		return fmt.Errorf("failed to generate deep dive: %w", err)
	}

	return a.output.WriteDeepDive(content, outputPath)
}

func (a *App) Stats(ctx context.Context) error {
	stats, err := a.db.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	return a.output.FormatStats(stats)
}
