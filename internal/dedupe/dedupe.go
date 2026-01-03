package dedupe

import (
	"strings"

	"github.com/vib795/AI-Idea-Scout/internal/db"
)

type Deduper struct{}

func NewDeduper() *Deduper {
	return &Deduper{}
}

type Cluster struct {
	CanonicalID int64
	MemberIDs   []int64
	Similarity  float64
}

// ClusterIdeas groups similar ideas together
func (d *Deduper) ClusterIdeas(ideas []*db.Idea) []Cluster {
	if len(ideas) == 0 {
		return nil
	}

	// Simple clustering: group ideas with similar titles and problems
	clusters := make(map[int64]*Cluster)
	processed := make(map[int64]bool)

	for i, idea1 := range ideas {
		if processed[idea1.ID] {
			continue
		}

		cluster := &Cluster{
			CanonicalID: idea1.ID,
			MemberIDs:   []int64{idea1.ID},
		}

		// Find similar ideas
		for j := i + 1; j < len(ideas); j++ {
			idea2 := ideas[j]
			if processed[idea2.ID] {
				continue
			}

			similarity := d.computeSimilarity(idea1, idea2)
			if similarity > 0.7 { // 70% similarity threshold
				cluster.MemberIDs = append(cluster.MemberIDs, idea2.ID)
				processed[idea2.ID] = true
			}
		}

		clusters[idea1.ID] = cluster
		processed[idea1.ID] = true
	}

	// Convert to slice
	result := make([]Cluster, 0, len(clusters))
	for _, cluster := range clusters {
		if len(cluster.MemberIDs) > 1 {
			result = append(result, *cluster)
		}
	}

	return result
}

func (d *Deduper) computeSimilarity(idea1, idea2 *db.Idea) float64 {
	// Title similarity
	titleSim := d.stringSimilarity(idea1.Title, idea2.Title)

	// Problem similarity
	problemSim := 0.0
	if idea1.Problem.Valid && idea2.Problem.Valid {
		problemSim = d.stringSimilarity(idea1.Problem.String, idea2.Problem.String)
	}

	// Category match
	categoryMatch := 0.0
	if idea1.Category.Valid && idea2.Category.Valid && idea1.Category.String == idea2.Category.String {
		categoryMatch = 1.0
	}

	// Weighted average
	return 0.5*titleSim + 0.3*problemSim + 0.2*categoryMatch
}

func (d *Deduper) stringSimilarity(s1, s2 string) float64 {
	// Simple word-based Jaccard similarity
	words1 := tokenize(s1)
	words2 := tokenize(s2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	intersection := 0
	set1 := make(map[string]bool)
	for _, w := range words1 {
		set1[w] = true
	}

	for _, w := range words2 {
		if set1[w] {
			intersection++
		}
	}

	union := len(words1) + len(words2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	words := strings.Fields(s)

	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
	}

	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if !stopWords[w] && len(w) > 2 {
			filtered = append(filtered, w)
		}
	}

	return filtered
}
