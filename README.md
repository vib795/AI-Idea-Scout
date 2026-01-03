# AI Idea Scout

**AI Idea Scout** is a production-quality CLI tool that monitors high-signal public sources (HackerNews, Reddit, RSS, GitHub) and extracts buildable AI/agent product ideas using Claude AI.

## Features

- **Multi-source aggregation**: HackerNews, Reddit, RSS feeds, GitHub Search API
- **AI-powered extraction**: Uses Claude to extract structured, actionable ideas
- **Smart deduplication**: Identifies similar ideas across sources
- **Multi-factor ranking**: Scores ideas by buildability, demand, novelty, and distribution
- **Deep dive generation**: Creates detailed MVP specifications for top ideas
- **Interactive triage**: Review and categorize ideas efficiently
- **Full-text search**: FTS5-powered search across all content
- **Compliance-first**: Rate limiting, caching, retries, respectful scraping

## Installation

### Homebrew (macOS/Linux)

```bash
# Coming soon
brew tap vib795/tap
brew install ideascout
```

### From source

```bash
git clone https://github.com/vib795/AI-Idea-Scout.git
cd AI-Idea-Scout
make build
make install
```

### Download binary

Download the latest release from [GitHub Releases](https://github.com/vib795/AI-Idea-Scout/releases).

## Quick Start

### 1. Initialize configuration

```bash
ideascout config init
```

This creates `~/.config/ideascout/config.yaml`. Edit it to add your Anthropic API key:

```yaml
anthropic_api_key: "sk-ant-..."
```

Get your API key at [console.anthropic.com](https://console.anthropic.com/).

### 2. Initialize database

```bash
ideascout db migrate
```

### 3. Fetch and analyze ideas

```bash
# Fetch from all enabled sources
ideascout fetch

# Run the complete pipeline
ideascout run

# List top ideas
ideascout list --top 10
```

### 4. Review ideas

```bash
# Show detailed view of an idea
ideascout show 42

# Interactive triage
ideascout triage

# Generate deep-dive MVP spec
ideascout deepdive 42 --output mvp-spec.md
```

## Commands

### Core Commands

- `ideascout fetch` - Fetch content from sources
- `ideascout run` - Run complete pipeline (fetch + analyze + rank)
- `ideascout list` - List ideas with filters
- `ideascout show <id>` - Show detailed idea
- `ideascout search "<query>"` - Full-text search

### Workflow Commands

- `ideascout triage` - Interactive review workflow
- `ideascout digest` - Generate daily/weekly digest
- `ideascout export --output ideas.md` - Export ideas to file
- `ideascout deepdive <id> --output spec.md` - Generate MVP specification

### Admin Commands

- `ideascout config init` - Initialize configuration
- `ideascout db migrate` - Run database migrations
- `ideascout stats` - Show statistics
- `ideascout version` - Show version info

## Configuration

Example `config.yaml`:

```yaml
anthropic_api_key: "sk-ant-..."
model: "claude-sonnet-4-5-20250929"
safe_mode: false

sources:
  hackernews:
    enabled: true
    min_score: 50
    max_items: 50

  reddit:
    enabled: true
    subreddits:
      - "LocalLLaMA"
      - "MachineLearning"
      - "ArtificialIntelligence"
    sort: "top"
    time: "week"
    min_score: 20

  rss:
    enabled: true
    feeds:
      - "https://example.com/feed.xml"

  github:
    enabled: false
    token: ""  # Optional, increases rate limits
    min_stars: 100

analysis:
  max_analyze: 100
  batch_size: 10
  min_overall_score_to_show: 5
```

Environment variables override config:

```bash
export IDEASCOUT_ANTHROPIC_API_KEY="sk-ant-..."
export IDEASCOUT_SAFE_MODE=true
```

## How It Works

### Pipeline Stages

1. **Ingest**: Fetch raw content from sources with rate limiting
2. **Normalize**: Clean and hash content
3. **Filter**: Apply heuristics to identify high-signal content
4. **Extract**: Use Claude to extract structured ideas
5. **Dedupe**: Cluster similar ideas
6. **Rank**: Compute multi-factor scores

### Ranking Algorithm

```
overall_score =
  0.30 * buildability +
  0.25 * demand +
  0.20 * clarity +
  0.15 * distribution +
  0.10 * novelty
```

- **Buildability**: Can a solo dev/small team build this?
- **Demand**: Clear evidence of user pain and willingness to pay?
- **Clarity**: Well-defined problem, users, and approach?
- **Distribution**: Strong go-to-market wedge?
- **Novelty**: Unique approach or underserved niche?

## Compliance & Safety

- **Official APIs**: Uses HackerNews Firebase API, GitHub Search API
- **Public feeds**: Reddit JSON endpoints, RSS/Atom feeds
- **Rate limiting**: Token bucket per source (respects ToS)
- **Caching**: ETag support, 1-hour TTL cache
- **Retries**: Exponential backoff with jitter
- **User-Agent**: Identifies as AI Idea Scout with contact info

**Safe Mode**: When enabled, only uses sources marked `official_api: true`.

## Development

### Prerequisites

- Go 1.24+
- SQLite (via modernc.org/sqlite, pure Go)
- Anthropic API key

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

### Release

```bash
# Tag a release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GoReleaser will build and publish
```

## Examples

### Daily workflow

```bash
# Morning: fetch new content
ideascout run --since 24h

# Review top ideas
ideascout list --top 20

# Triage new ideas
ideascout triage

# Generate spec for promising idea
ideascout deepdive 42 --output mvp-spec.md
```

### Weekly digest

```bash
ideascout digest --since 7d --top 10 --format md > weekly-digest.md
```

### Search for specific topics

```bash
ideascout search "LLM automation"
ideascout search "agent framework"
```

## Architecture

```
ideascout/
├── cmd/ideascout/         # CLI entry point
├── internal/
│   ├── app/               # Application wiring
│   ├── cli/               # Cobra commands
│   ├── config/            # Viper configuration
│   ├── db/                # SQLite + migrations
│   ├── fetch/             # HTTP client with retry/cache
│   ├── sources/           # Source adapters (HN, Reddit, RSS, GitHub)
│   ├── pipeline/          # Stage orchestration
│   ├── llm/               # Claude client + prompts
│   ├── rank/              # Scoring algorithm
│   ├── dedupe/            # Clustering
│   ├── output/            # Formatting
│   └── triage/            # Interactive workflow
├── migrations/            # SQL migrations
└── testdata/fixtures/     # Test fixtures
```

## Troubleshooting

### "Config file not found"

Run `ideascout config init` to create the default config.

### "HTTP 429: Too Many Requests"

Adjust rate limits in config or wait. The tool respects rate limits automatically.

### "Anthropic API key is required"

Set your API key in `config.yaml` or via `IDEASCOUT_ANTHROPIC_API_KEY` environment variable.

### Network errors

Check your internet connection. The tool will retry with exponential backoff.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Run `make lint` and `make test`
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Uses [Anthropic Claude](https://anthropic.com) for AI extraction
- Powered by [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) for pure Go SQLite

## Contact

- Issues: [GitHub Issues](https://github.com/vib795/AI-Idea-Scout/issues)
- Discussions: [GitHub Discussions](https://github.com/vib795/AI-Idea-Scout/discussions)

---

**Note**: This tool is for research and ideation purposes. Always validate demand before building. Respect all source Terms of Service.
