package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Raw Content Operations

func (db *DB) InsertRawContent(ctx context.Context, rc *RawContent) error {
	query := `
		INSERT INTO raw_content (
			id, source, source_id, url, title, body, author, score, comments_count,
			created_at, fetched_at, raw_json, normalized_text, content_hash, processed_stage
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source, source_id) DO UPDATE SET
			score = excluded.score,
			comments_count = excluded.comments_count,
			fetched_at = excluded.fetched_at
	`
	_, err := db.ExecContext(ctx, query,
		rc.ID, rc.Source, rc.SourceID, rc.URL, rc.Title, rc.Body, rc.Author,
		rc.Score, rc.CommentsCount, rc.CreatedAt, rc.FetchedAt, rc.RawJSON,
		rc.NormalizedText, rc.ContentHash, rc.ProcessedStage,
	)
	return err
}

func (db *DB) GetRawContentByID(ctx context.Context, id string) (*RawContent, error) {
	query := `
		SELECT id, source, source_id, url, title, body, author, score, comments_count,
			created_at, fetched_at, raw_json, normalized_text, content_hash, processed_stage
		FROM raw_content WHERE id = ?
	`
	var rc RawContent
	err := db.QueryRowContext(ctx, query, id).Scan(
		&rc.ID, &rc.Source, &rc.SourceID, &rc.URL, &rc.Title, &rc.Body, &rc.Author,
		&rc.Score, &rc.CommentsCount, &rc.CreatedAt, &rc.FetchedAt, &rc.RawJSON,
		&rc.NormalizedText, &rc.ContentHash, &rc.ProcessedStage,
	)
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (db *DB) UpdateRawContentStage(ctx context.Context, id, stage string) error {
	_, err := db.ExecContext(ctx, "UPDATE raw_content SET processed_stage = ? WHERE id = ?", stage, id)
	return err
}

func (db *DB) GetRawContentByStage(ctx context.Context, stage string, limit int) ([]*RawContent, error) {
	query := `
		SELECT id, source, source_id, url, title, body, author, score, comments_count,
			created_at, fetched_at, raw_json, normalized_text, content_hash, processed_stage
		FROM raw_content WHERE processed_stage = ? ORDER BY fetched_at DESC
	`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.QueryContext(ctx, query, stage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*RawContent
	for rows.Next() {
		var rc RawContent
		if err := rows.Scan(
			&rc.ID, &rc.Source, &rc.SourceID, &rc.URL, &rc.Title, &rc.Body, &rc.Author,
			&rc.Score, &rc.CommentsCount, &rc.CreatedAt, &rc.FetchedAt, &rc.RawJSON,
			&rc.NormalizedText, &rc.ContentHash, &rc.ProcessedStage,
		); err != nil {
			return nil, err
		}
		results = append(results, &rc)
	}
	return results, rows.Err()
}

// Idea Operations

func (db *DB) InsertIdea(ctx context.Context, idea *Idea) (int64, error) {
	query := `
		INSERT INTO ideas (
			content_id, title, one_liner, problem, target_users, job_to_be_done,
			data_requirements, technical_approach, complexity, category, tags,
			buildability_score, demand_score, novelty_score, distribution_wedge,
			moat, risks, evidence, next_steps, overall_score, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.ExecContext(ctx, query,
		idea.ContentID, idea.Title, idea.OneLiner, idea.Problem, idea.TargetUsers,
		idea.JobToBeDone, idea.DataRequirements, idea.TechnicalApproach, idea.Complexity,
		idea.Category, idea.Tags, idea.BuildabilityScore, idea.DemandScore,
		idea.NoveltyScore, idea.DistributionWedge, idea.Moat, idea.Risks,
		idea.Evidence, idea.NextSteps, idea.OverallScore, idea.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) GetIdeaByID(ctx context.Context, id int64) (*Idea, error) {
	query := `
		SELECT id, content_id, title, one_liner, problem, target_users, job_to_be_done,
			data_requirements, technical_approach, complexity, category, tags,
			buildability_score, demand_score, novelty_score, distribution_wedge,
			moat, risks, evidence, next_steps, overall_score, status, created_at
		FROM ideas WHERE id = ?
	`
	var idea Idea
	err := db.QueryRowContext(ctx, query, id).Scan(
		&idea.ID, &idea.ContentID, &idea.Title, &idea.OneLiner, &idea.Problem,
		&idea.TargetUsers, &idea.JobToBeDone, &idea.DataRequirements, &idea.TechnicalApproach,
		&idea.Complexity, &idea.Category, &idea.Tags, &idea.BuildabilityScore,
		&idea.DemandScore, &idea.NoveltyScore, &idea.DistributionWedge, &idea.Moat,
		&idea.Risks, &idea.Evidence, &idea.NextSteps, &idea.OverallScore,
		&idea.Status, &idea.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &idea, nil
}

func (db *DB) UpdateIdeaStatus(ctx context.Context, id int64, status string) error {
	_, err := db.ExecContext(ctx, "UPDATE ideas SET status = ? WHERE id = ?", status, id)
	return err
}

func (db *DB) ListIdeas(ctx context.Context, filters map[string]interface{}, limit int) ([]*Idea, error) {
	query := `
		SELECT id, content_id, title, one_liner, problem, target_users, job_to_be_done,
			data_requirements, technical_approach, complexity, category, tags,
			buildability_score, demand_score, novelty_score, distribution_wedge,
			moat, risks, evidence, next_steps, overall_score, status, created_at
		FROM ideas WHERE 1=1
	`
	args := []interface{}{}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if category, ok := filters["category"].(string); ok && category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	if complexity, ok := filters["complexity"].(string); ok && complexity != "" {
		query += " AND complexity = ?"
		args = append(args, complexity)
	}

	query += " ORDER BY overall_score DESC, created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ideas []*Idea
	for rows.Next() {
		var idea Idea
		if err := rows.Scan(
			&idea.ID, &idea.ContentID, &idea.Title, &idea.OneLiner, &idea.Problem,
			&idea.TargetUsers, &idea.JobToBeDone, &idea.DataRequirements, &idea.TechnicalApproach,
			&idea.Complexity, &idea.Category, &idea.Tags, &idea.BuildabilityScore,
			&idea.DemandScore, &idea.NoveltyScore, &idea.DistributionWedge, &idea.Moat,
			&idea.Risks, &idea.Evidence, &idea.NextSteps, &idea.OverallScore,
			&idea.Status, &idea.CreatedAt,
		); err != nil {
			return nil, err
		}
		ideas = append(ideas, &idea)
	}
	return ideas, rows.Err()
}

// LLM Cache Operations

func (db *DB) GetLLMCache(ctx context.Context, key string) (*LLMCacheEntry, error) {
	query := "SELECT key, model, prompt_version, response_json, created_at FROM llm_cache WHERE key = ?"
	var entry LLMCacheEntry
	err := db.QueryRowContext(ctx, query, key).Scan(
		&entry.Key, &entry.Model, &entry.PromptVersion, &entry.ResponseJSON, &entry.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (db *DB) SetLLMCache(ctx context.Context, entry *LLMCacheEntry) error {
	query := `
		INSERT INTO llm_cache (key, model, prompt_version, response_json, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			response_json = excluded.response_json,
			created_at = excluded.created_at
	`
	_, err := db.ExecContext(ctx, query,
		entry.Key, entry.Model, entry.PromptVersion, entry.ResponseJSON, entry.CreatedAt,
	)
	return err
}

// FTS5 Search

func (db *DB) SearchContent(ctx context.Context, query string, limit int) ([]*SearchResult, error) {
	sql := `
		SELECT id, source_type, title, snippet(content_fts, 2, '<b>', '</b>', '...', 32) as snippet, rank
		FROM content_fts
		WHERE content_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`
	rows, err := db.QueryContext(ctx, sql, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ID, &r.SourceType, &r.Title, &r.Snippet, &r.Rank); err != nil {
			return nil, err
		}
		results = append(results, &r)
	}
	return results, rows.Err()
}

// Stats

func (db *DB) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalRaw, totalIdeas, totalClusters int64
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM raw_content").Scan(&totalRaw)
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ideas").Scan(&totalIdeas)
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM idea_clusters").Scan(&totalClusters)

	stats["total_raw_content"] = totalRaw
	stats["total_ideas"] = totalIdeas
	stats["total_clusters"] = totalClusters

	// Ideas by status
	rows, err := db.QueryContext(ctx, "SELECT status, COUNT(*) FROM ideas GROUP BY status")
	if err == nil {
		defer rows.Close()
		byStatus := make(map[string]int64)
		for rows.Next() {
			var status string
			var count int64
			rows.Scan(&status, &count)
			byStatus[status] = count
		}
		stats["ideas_by_status"] = byStatus
	}

	// Raw content by source
	rows, err = db.QueryContext(ctx, "SELECT source, COUNT(*) FROM raw_content GROUP BY source")
	if err == nil {
		defer rows.Close()
		bySource := make(map[string]int64)
		for rows.Next() {
			var source string
			var count int64
			rows.Scan(&source, &count)
			bySource[source] = count
		}
		stats["raw_content_by_source"] = bySource
	}

	return stats, nil
}
