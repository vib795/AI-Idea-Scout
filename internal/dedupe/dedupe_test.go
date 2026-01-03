package dedupe

import (
	"database/sql"
	"testing"

	"github.com/vib795/AI-Idea-Scout/internal/db"
)

func TestStringSimilarity(t *testing.T) {
	d := NewDeduper()

	tests := []struct {
		name     string
		s1       string
		s2       string
		minScore float64
	}{
		{
			name:     "identical strings",
			s1:       "AI agent for email automation",
			s2:       "AI agent for email automation",
			minScore: 0.9,
		},
		{
			name:     "similar strings",
			s1:       "Build an AI agent for automating emails",
			s2:       "Create AI automation for email tasks",
			minScore: 0.3,
		},
		{
			name:     "completely different",
			s1:       "AI agent for email",
			s2:       "Database migration tool",
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := d.stringSimilarity(tt.s1, tt.s2)
			if score < tt.minScore {
				t.Errorf("similarity too low: got %.2f, want >= %.2f", score, tt.minScore)
			}
		})
	}
}

func TestClusterIdeas(t *testing.T) {
	d := NewDeduper()

	ideas := []*db.Idea{
		{
			ID:    1,
			Title: "AI email automation agent",
			Problem: sql.NullString{
				String: "People spend too much time on email",
				Valid:  true,
			},
			Category: sql.NullString{String: "automation", Valid: true},
		},
		{
			ID:    2,
			Title: "Email automation with AI",
			Problem: sql.NullString{
				String: "Email takes too long to manage",
				Valid:  true,
			},
			Category: sql.NullString{String: "automation", Valid: true},
		},
		{
			ID:    3,
			Title: "Code review automation tool",
			Problem: sql.NullString{
				String: "Code reviews are time-consuming",
				Valid:  true,
			},
			Category: sql.NullString{String: "automation", Valid: true},
		},
	}

	clusters := d.ClusterIdeas(ideas)

	// Should find at least one cluster (ideas 1 and 2 are similar)
	if len(clusters) == 0 {
		t.Error("expected at least one cluster, got none")
	}
}
