package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vib795/AI-Idea-Scout/internal/db"
)

type Formatter struct {
	color bool
}

func NewFormatter(color bool) *Formatter {
	return &Formatter{color: color}
}

func (f *Formatter) FormatIdeasList(ideas []*db.Idea, format string) error {
	switch format {
	case "json":
		return f.formatJSON(ideas)
	case "md":
		return f.formatMarkdown(ideas)
	default:
		return f.formatText(ideas)
	}
}

func (f *Formatter) formatText(ideas []*db.Idea) error {
	if len(ideas) == 0 {
		fmt.Println("No ideas found.")
		return nil
	}

	fmt.Printf("\nFound %d ideas:\n\n", len(ideas))

	for i, idea := range ideas {
		fmt.Printf("─────────────────────────────────────────────────────────────\n")
		fmt.Printf("[%d] %s\n", idea.ID, f.bold(idea.Title))

		if idea.OneLiner.Valid {
			fmt.Printf("    %s\n\n", idea.OneLiner.String)
		}

		// Scores
		if idea.OverallScore.Valid {
			fmt.Printf("    Score: %.1f/10", idea.OverallScore.Float64)
		}
		if idea.BuildabilityScore.Valid {
			fmt.Printf(" | Build: %d/10", idea.BuildabilityScore.Int64)
		}
		if idea.DemandScore.Valid {
			fmt.Printf(" | Demand: %d/10", idea.DemandScore.Int64)
		}
		fmt.Println()

		// Metadata
		meta := []string{}
		if idea.Category.Valid {
			meta = append(meta, idea.Category.String)
		}
		if idea.Complexity.Valid {
			meta = append(meta, idea.Complexity.String)
		}
		if idea.Status != "" {
			meta = append(meta, idea.Status)
		}
		if len(meta) > 0 {
			fmt.Printf("    %s\n", strings.Join(meta, " • "))
		}

		fmt.Println()

		// Show only first 3 for brevity
		if i >= 2 && len(ideas) > 5 {
			fmt.Printf("\n... and %d more (use --top N or 'ideascout show <id>' for details)\n\n", len(ideas)-3)
			break
		}
	}

	return nil
}

func (f *Formatter) formatJSON(ideas []*db.Idea) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(ideas)
}

func (f *Formatter) formatMarkdown(ideas []*db.Idea) error {
	fmt.Println("# Ideas\n")
	for _, idea := range ideas {
		fmt.Printf("## %s\n\n", idea.Title)
		if idea.OneLiner.Valid {
			fmt.Printf("*%s*\n\n", idea.OneLiner.String)
		}
		if idea.Problem.Valid {
			fmt.Printf("**Problem:** %s\n\n", idea.Problem.String)
		}
		fmt.Println("---\n")
	}
	return nil
}

func (f *Formatter) FormatIdea(idea *db.Idea, format string) error {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(idea)
	}

	// Detailed text format
	fmt.Printf("\n%s\n", f.bold(idea.Title))
	fmt.Println(strings.Repeat("=", len(idea.Title)))

	if idea.OneLiner.Valid {
		fmt.Printf("\n%s\n", idea.OneLiner.String)
	}

	fmt.Println("\n## Details")
	if idea.Problem.Valid {
		fmt.Printf("\n**Problem:**\n%s\n", idea.Problem.String)
	}
	if idea.TargetUsers.Valid {
		fmt.Printf("\n**Target Users:**\n%s\n", idea.TargetUsers.String)
	}
	if idea.NextSteps.Valid {
		fmt.Printf("\n**Next Steps:**\n%s\n", idea.NextSteps.String)
	}

	fmt.Println("\n## Scores")
	if idea.OverallScore.Valid {
		fmt.Printf("Overall: %.1f/10\n", idea.OverallScore.Float64)
	}
	if idea.BuildabilityScore.Valid {
		fmt.Printf("Buildability: %d/10\n", idea.BuildabilityScore.Int64)
	}
	if idea.DemandScore.Valid {
		fmt.Printf("Demand: %d/10\n", idea.DemandScore.Int64)
	}
	if idea.NoveltyScore.Valid {
		fmt.Printf("Novelty: %d/10\n", idea.NoveltyScore.Int64)
	}

	return nil
}

func (f *Formatter) FormatSearchResults(results []*db.SearchResult, format string) error {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	fmt.Printf("\nFound %d results:\n\n", len(results))
	for _, r := range results {
		fmt.Printf("[%s:%s] %s\n", r.SourceType, r.ID, r.Title)
		if r.Snippet != "" {
			fmt.Printf("  %s\n", r.Snippet)
		}
		fmt.Println()
	}

	return nil
}

func (f *Formatter) FormatDigest(ideas []*db.Idea, since string, format string) error {
	if format == "md" {
		return f.generateMarkdownDigest(ideas, since)
	}

	fmt.Printf("\n# AI Idea Scout - Daily Digest\n")
	fmt.Printf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	return f.formatText(ideas)
}

func (f *Formatter) generateMarkdownDigest(ideas []*db.Idea, since string) error {
	fmt.Printf("# AI Idea Scout - Top Ideas\n\n")
	fmt.Printf("*Generated: %s*\n\n", time.Now().Format("2006-01-02 15:04:05"))

	if len(ideas) == 0 {
		fmt.Println("No ideas found in this period.\n")
		return nil
	}

	for i, idea := range ideas {
		fmt.Printf("## %d. %s\n\n", i+1, idea.Title)

		if idea.OneLiner.Valid {
			fmt.Printf("**%s**\n\n", idea.OneLiner.String)
		}

		if idea.Problem.Valid {
			fmt.Printf("### Problem\n\n%s\n\n", idea.Problem.String)
		}

		if idea.TargetUsers.Valid {
			fmt.Printf("### Target Users\n\n%s\n\n", idea.TargetUsers.String)
		}

		// Scores
		fmt.Printf("### Scores\n\n")
		if idea.BuildabilityScore.Valid {
			fmt.Printf("- **Buildability:** %d/10\n", idea.BuildabilityScore.Int64)
		}
		if idea.DemandScore.Valid {
			fmt.Printf("- **Demand:** %d/10\n", idea.DemandScore.Int64)
		}
		if idea.OverallScore.Valid {
			fmt.Printf("- **Overall:** %.1f/10\n", idea.OverallScore.Float64)
		}

		fmt.Println()
		if idea.Complexity.Valid {
			fmt.Printf("**Complexity:** %s\n\n", idea.Complexity.String)
		}

		fmt.Println("---\n")
	}

	return nil
}

func (f *Formatter) ExportToFile(ideas []*db.Idea, outputPath string, format string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Temporarily redirect stdout
	oldStdout := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = oldStdout }()

	if format == "json" {
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		return enc.Encode(ideas)
	}

	// Markdown format
	return f.generateMarkdownDigest(ideas, "all")
}

func (f *Formatter) WriteDeepDive(content string, outputPath string) error {
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("✓ Deep dive written to: %s\n", outputPath)
	return nil
}

func (f *Formatter) FormatStats(stats map[string]interface{}) error {
	fmt.Println("\n=== AI Idea Scout Statistics ===\n")

	if total, ok := stats["total_raw_content"].(int64); ok {
		fmt.Printf("Total raw content items: %d\n", total)
	}
	if total, ok := stats["total_ideas"].(int64); ok {
		fmt.Printf("Total ideas extracted: %d\n", total)
	}
	if total, ok := stats["total_clusters"].(int64); ok {
		fmt.Printf("Total idea clusters: %d\n", total)
	}

	if bySource, ok := stats["raw_content_by_source"].(map[string]int64); ok {
		fmt.Println("\nContent by source:")
		for source, count := range bySource {
			fmt.Printf("  %s: %d\n", source, count)
		}
	}

	if byStatus, ok := stats["ideas_by_status"].(map[string]int64); ok {
		fmt.Println("\nIdeas by status:")
		for status, count := range byStatus {
			fmt.Printf("  %s: %d\n", status, count)
		}
	}

	fmt.Println()
	return nil
}

func (f *Formatter) bold(s string) string {
	if f.color {
		return "\033[1m" + s + "\033[0m"
	}
	return s
}
