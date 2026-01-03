package triage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/vib795/AI-Idea-Scout/internal/db"
)

type Manager struct {
	db *db.DB
}

func NewManager(database *db.DB) *Manager {
	return &Manager{db: database}
}

func (m *Manager) Run(ctx context.Context) error {
	// Get ideas with status 'new'
	ideas, err := m.db.ListIdeas(ctx, map[string]interface{}{"status": "new"}, 50)
	if err != nil {
		return fmt.Errorf("failed to get ideas: %w", err)
	}

	if len(ideas) == 0 {
		fmt.Println("No new ideas to triage.")
		return nil
	}

	fmt.Printf("\n=== Interactive Triage ===\n")
	fmt.Printf("Found %d new ideas to review.\n\n", len(ideas))

	reader := bufio.NewReader(os.Stdin)

	for i, idea := range ideas {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(ideas), idea.Title)
		fmt.Println(strings.Repeat("-", 60))

		if idea.OneLiner.Valid {
			fmt.Printf("\n%s\n", idea.OneLiner.String)
		}

		if idea.Problem.Valid {
			fmt.Printf("\nProblem: %s\n", truncate(idea.Problem.String, 200))
		}

		if idea.BuildabilityScore.Valid || idea.DemandScore.Valid {
			fmt.Print("\nScores: ")
			if idea.BuildabilityScore.Valid {
				fmt.Printf("Build: %d/10 ", idea.BuildabilityScore.Int64)
			}
			if idea.DemandScore.Valid {
				fmt.Printf("Demand: %d/10 ", idea.DemandScore.Int64)
			}
			fmt.Println()
		}

		fmt.Println("\nActions:")
		fmt.Println("  [s] Shortlist")
		fmt.Println("  [r] Mark as reviewed")
		fmt.Println("  [a] Archive")
		fmt.Println("  [n] Next (keep as new)")
		fmt.Println("  [q] Quit")
		fmt.Print("\nYour choice: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "s":
			if err := m.db.UpdateIdeaStatus(ctx, idea.ID, "shortlisted"); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("✓ Shortlisted")
			}
		case "r":
			if err := m.db.UpdateIdeaStatus(ctx, idea.ID, "reviewed"); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("✓ Marked as reviewed")
			}
		case "a":
			if err := m.db.UpdateIdeaStatus(ctx, idea.ID, "archived"); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("✓ Archived")
			}
		case "q":
			fmt.Println("\nExiting triage.")
			return nil
		case "n":
			fmt.Println("Keeping as new")
		default:
			fmt.Println("Invalid choice, keeping as new")
		}
	}

	fmt.Println("\n✓ Triage complete")
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
