-- +goose Up
-- +goose StatementBegin

-- Raw content from sources (immutable snapshots)
CREATE TABLE IF NOT EXISTS raw_content (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    source_id TEXT NOT NULL,
    url TEXT,
    title TEXT NOT NULL,
    body TEXT,
    author TEXT,
    score INTEGER,
    comments_count INTEGER,
    created_at DATETIME,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    raw_json TEXT,
    normalized_text TEXT,
    content_hash TEXT NOT NULL,
    processed_stage TEXT NOT NULL DEFAULT 'ingested',
    UNIQUE(source, source_id)
);

CREATE INDEX IF NOT EXISTS idx_raw_content_hash ON raw_content(content_hash);
CREATE INDEX IF NOT EXISTS idx_raw_content_stage ON raw_content(processed_stage);
CREATE INDEX IF NOT EXISTS idx_raw_content_source ON raw_content(source);
CREATE INDEX IF NOT EXISTS idx_raw_content_created ON raw_content(created_at DESC);

-- Extracted ideas
CREATE TABLE IF NOT EXISTS ideas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_id TEXT NOT NULL REFERENCES raw_content(id),
    title TEXT NOT NULL,
    one_liner TEXT,
    problem TEXT,
    target_users TEXT,
    job_to_be_done TEXT,
    data_requirements TEXT,
    technical_approach TEXT,
    complexity TEXT,
    category TEXT,
    tags TEXT,
    buildability_score INTEGER,
    demand_score INTEGER,
    novelty_score INTEGER,
    distribution_wedge TEXT,
    moat TEXT,
    risks TEXT,
    evidence TEXT,
    next_steps TEXT,
    overall_score REAL,
    status TEXT DEFAULT 'new',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ideas_content ON ideas(content_id);
CREATE INDEX IF NOT EXISTS idx_ideas_status ON ideas(status);
CREATE INDEX IF NOT EXISTS idx_ideas_score ON ideas(overall_score DESC);
CREATE INDEX IF NOT EXISTS idx_ideas_category ON ideas(category);
CREATE INDEX IF NOT EXISTS idx_ideas_complexity ON ideas(complexity);
CREATE INDEX IF NOT EXISTS idx_ideas_created ON ideas(created_at DESC);

-- Idea clusters for deduplication
CREATE TABLE IF NOT EXISTS idea_clusters (
    cluster_id TEXT PRIMARY KEY,
    canonical_idea_id INTEGER REFERENCES ideas(id),
    member_idea_ids TEXT,
    similarity_notes TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_clusters_canonical ON idea_clusters(canonical_idea_id);

-- LLM response cache
CREATE TABLE IF NOT EXISTS llm_cache (
    key TEXT PRIMARY KEY,
    model TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    response_json TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_llm_cache_created ON llm_cache(created_at);

-- FTS5 virtual table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS content_fts USING fts5(
    id UNINDEXED,
    source_type UNINDEXED,
    title,
    content,
    content='',
    tokenize='porter unicode61'
);

-- Triggers to keep FTS5 in sync
CREATE TRIGGER IF NOT EXISTS raw_content_fts_insert AFTER INSERT ON raw_content BEGIN
    INSERT INTO content_fts(id, source_type, title, content)
    VALUES (new.id, 'raw', new.title, new.normalized_text);
END;

CREATE TRIGGER IF NOT EXISTS raw_content_fts_update AFTER UPDATE ON raw_content BEGIN
    UPDATE content_fts SET title = new.title, content = new.normalized_text
    WHERE id = new.id AND source_type = 'raw';
END;

CREATE TRIGGER IF NOT EXISTS raw_content_fts_delete AFTER DELETE ON raw_content BEGIN
    DELETE FROM content_fts WHERE id = old.id AND source_type = 'raw';
END;

CREATE TRIGGER IF NOT EXISTS ideas_fts_insert AFTER INSERT ON ideas BEGIN
    INSERT INTO content_fts(id, source_type, title, content)
    VALUES (CAST(new.id AS TEXT), 'idea', new.title, new.one_liner || ' ' || new.problem);
END;

CREATE TRIGGER IF NOT EXISTS ideas_fts_update AFTER UPDATE ON ideas BEGIN
    UPDATE content_fts SET title = new.title, content = new.one_liner || ' ' || new.problem
    WHERE id = CAST(new.id AS TEXT) AND source_type = 'idea';
END;

CREATE TRIGGER IF NOT EXISTS ideas_fts_delete AFTER DELETE ON ideas BEGIN
    DELETE FROM content_fts WHERE id = CAST(old.id AS TEXT) AND source_type = 'idea';
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS ideas_fts_delete;
DROP TRIGGER IF EXISTS ideas_fts_update;
DROP TRIGGER IF EXISTS ideas_fts_insert;
DROP TRIGGER IF EXISTS raw_content_fts_delete;
DROP TRIGGER IF EXISTS raw_content_fts_update;
DROP TRIGGER IF EXISTS raw_content_fts_insert;
DROP TABLE IF EXISTS content_fts;
DROP INDEX IF EXISTS idx_llm_cache_created;
DROP TABLE IF EXISTS llm_cache;
DROP INDEX IF EXISTS idx_clusters_canonical;
DROP TABLE IF EXISTS idea_clusters;
DROP INDEX IF EXISTS idx_ideas_created;
DROP INDEX IF EXISTS idx_ideas_complexity;
DROP INDEX IF EXISTS idx_ideas_category;
DROP INDEX IF EXISTS idx_ideas_score;
DROP INDEX IF EXISTS idx_ideas_status;
DROP INDEX IF EXISTS idx_ideas_content;
DROP TABLE IF EXISTS ideas;
DROP INDEX IF EXISTS idx_raw_content_created;
DROP INDEX IF EXISTS idx_raw_content_source;
DROP INDEX IF EXISTS idx_raw_content_stage;
DROP INDEX IF EXISTS idx_raw_content_hash;
DROP TABLE IF NOT EXISTS raw_content;

-- +goose StatementEnd
