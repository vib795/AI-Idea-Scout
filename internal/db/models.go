package db

import (
	"database/sql"
	"time"
)

type RawContent struct {
	ID              string
	Source          string
	SourceID        string
	URL             sql.NullString
	Title           string
	Body            sql.NullString
	Author          sql.NullString
	Score           sql.NullInt64
	CommentsCount   sql.NullInt64
	CreatedAt       sql.NullTime
	FetchedAt       time.Time
	RawJSON         sql.NullString
	NormalizedText  sql.NullString
	ContentHash     string
	ProcessedStage  string
}

type Idea struct {
	ID                  int64
	ContentID           string
	Title               string
	OneLiner            sql.NullString
	Problem             sql.NullString
	TargetUsers         sql.NullString
	JobToBeDone         sql.NullString
	DataRequirements    sql.NullString // JSON
	TechnicalApproach   sql.NullString // markdown or JSON
	Complexity          sql.NullString
	Category            sql.NullString
	Tags                sql.NullString // JSON array
	BuildabilityScore   sql.NullInt64
	DemandScore         sql.NullInt64
	NoveltyScore        sql.NullInt64
	DistributionWedge   sql.NullString
	Moat                sql.NullString
	Risks               sql.NullString // JSON array
	Evidence            sql.NullString // JSON array
	NextSteps           sql.NullString // JSON array
	OverallScore        sql.NullFloat64
	Status              string
	CreatedAt           time.Time
}

type IdeaCluster struct {
	ClusterID       string
	CanonicalIdeaID sql.NullInt64
	MemberIdeaIDs   sql.NullString // JSON array
	SimilarityNotes sql.NullString
	UpdatedAt       time.Time
}

type LLMCacheEntry struct {
	Key            string
	Model          string
	PromptVersion  string
	ResponseJSON   string
	CreatedAt      time.Time
}

type SearchResult struct {
	ID         string
	SourceType string
	Title      string
	Snippet    string
	Rank       float64
}
