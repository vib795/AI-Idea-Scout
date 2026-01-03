package rank

import (
	"github.com/vib795/AI-Idea-Scout/internal/db"
)

type Ranker struct{}

func NewRanker() *Ranker {
	return &Ranker{}
}

// ComputeScore calculates the overall score for an idea
// Formula: 0.30*buildability + 0.25*demand + 0.20*clarity + 0.15*distribution + 0.10*novelty
func (r *Ranker) ComputeScore(idea *db.Idea) float64 {
	buildability := float64(0)
	if idea.BuildabilityScore.Valid {
		buildability = float64(idea.BuildabilityScore.Int64)
	}

	demand := float64(0)
	if idea.DemandScore.Valid {
		demand = float64(idea.DemandScore.Int64)
	}

	novelty := float64(0)
	if idea.NoveltyScore.Valid {
		novelty = float64(idea.NoveltyScore.Int64)
	}

	// Clarity: assess completeness of problem statement, target users, and job to be done
	clarity := r.assessClarity(idea)

	// Distribution: score the quality of distribution wedge
	distribution := r.assessDistribution(idea)

	// Weighted score
	score := 0.30*buildability + 0.25*demand + 0.20*clarity + 0.15*distribution + 0.10*novelty

	return score
}

func (r *Ranker) assessClarity(idea *db.Idea) float64 {
	score := 0.0

	if idea.Problem.Valid && len(idea.Problem.String) > 50 {
		score += 3.0
	}

	if idea.TargetUsers.Valid && len(idea.TargetUsers.String) > 20 {
		score += 3.0
	}

	if idea.JobToBeDone.Valid && len(idea.JobToBeDone.String) > 20 {
		score += 2.0
	}

	if idea.DataRequirements.Valid && len(idea.DataRequirements.String) > 10 {
		score += 2.0
	}

	return score
}

func (r *Ranker) assessDistribution(idea *db.Idea) float64 {
	if !idea.DistributionWedge.Valid {
		return 3.0
	}

	wedge := idea.DistributionWedge.String

	// High-value distribution channels
	if containsAny(wedge, []string{"viral", "network effect", "marketplace", "community"}) {
		return 9.0
	}

	// Medium-value channels
	if containsAny(wedge, []string{"seo", "integration", "api", "directory"}) {
		return 6.0
	}

	// Basic channels
	return 4.0
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
