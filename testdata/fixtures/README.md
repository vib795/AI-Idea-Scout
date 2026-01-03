# Test Fixtures

This directory contains test fixtures for offline testing.

## Structure

```
fixtures/
├── hn/          # HackerNews API responses
├── reddit/      # Reddit JSON responses
├── rss/         # RSS/Atom feed samples
└── github/      # GitHub Search API responses
```

## Usage

Fixtures are used in tests to avoid hitting live APIs. To record new fixtures:

```bash
# Set RECORD_FIXTURES=true to record HTTP responses
RECORD_FIXTURES=true go test ./...
```

## Format

Each fixture is a JSON file containing:
- HTTP status code
- Headers
- Response body
- Timestamp

Example:

```json
{
  "status": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "...",
  "recorded_at": "2026-01-03T12:00:00Z"
}
```
