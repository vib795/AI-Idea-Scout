package llm

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/vib795/AI-Idea-Scout/internal/db"
)

type Client struct {
	apiKey        string
	model         string
	promptVersion string
	client        *anthropic.Client
}

func NewClient(apiKey, model, promptVersion string) *Client {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &Client{
		apiKey:        apiKey,
		model:         model,
		promptVersion: promptVersion,
		client:        &client,
	}
}

// ExtractIdeas extracts buildable AI ideas from normalized content
func (c *Client) ExtractIdeas(ctx context.Context, content *db.RawContent, database *db.DB) ([]ExtractedIdea, error) {
	// Generate cache key
	cacheKey := c.generateCacheKey(content.ContentHash, "extract")

	// Check cache
	if cached, err := database.GetLLMCache(ctx, cacheKey); err == nil && cached != nil {
		// Check if cache is still valid (within TTL)
		var ideas []ExtractedIdea
		if err := json.Unmarshal([]byte(cached.ResponseJSON), &ideas); err == nil {
			return ideas, nil
		}
	}

	// Build prompt
	prompt := c.buildExtractionPrompt(content)

	// Call Claude
	maxTokens := int64(4096)
	systemPrompt := extractionSystemPrompt
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		System: []anthropic.TextBlockParam{
			{Type: "text", Text: systemPrompt},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	// Extract text from response
	var responseText string
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	// Parse JSON response
	ideas, err := c.parseIdeasResponse(responseText)
	if err != nil {
		// Retry once with clarification
		ideas, err = c.retryWithClarification(ctx, responseText, content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ideas after retry: %w", err)
		}
	}

	// Cache the result
	cacheEntry := &db.LLMCacheEntry{
		Key:            cacheKey,
		Model:          c.model,
		PromptVersion:  c.promptVersion,
		ResponseJSON:   mustMarshal(ideas),
		CreatedAt:      time.Now(),
	}
	database.SetLLMCache(ctx, cacheEntry)

	return ideas, nil
}

func (c *Client) buildExtractionPrompt(content *db.RawContent) string {
	// Build the content section
	text := fmt.Sprintf("Title: %s\n", content.Title)
	if content.Body.Valid && content.Body.String != "" {
		text += fmt.Sprintf("\nContent:\n%s\n", content.Body.String)
	}
	if content.URL.Valid {
		text += fmt.Sprintf("\nURL: %s\n", content.URL.String)
	}
	if content.Score.Valid {
		text += fmt.Sprintf("Score/Engagement: %d\n", content.Score.Int64)
	}

	return fmt.Sprintf(extractionUserPrompt, text)
}

func (c *Client) parseIdeasResponse(response string) ([]ExtractedIdea, error) {
	// Try to find JSON in the response
	start := -1
	for i := 0; i < len(response); i++ {
		if response[i] == '[' {
			start = i
			break
		}
	}

	if start == -1 {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := response[start:]

	var ideas []ExtractedIdea
	if err := json.Unmarshal([]byte(jsonStr), &ideas); err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	return ideas, nil
}

func (c *Client) retryWithClarification(ctx context.Context, previousResponse string, content *db.RawContent) ([]ExtractedIdea, error) {
	clarificationPrompt := fmt.Sprintf(`The previous response was not valid JSON. Please return ONLY a valid JSON array with no markdown formatting or explanation.

Previous response:
%s

Please provide the corrected JSON array now:`, previousResponse)

	maxTokens := int64(4096)
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(clarificationPrompt)),
		},
	})

	if err != nil {
		return nil, err
	}

	var responseText string
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	return c.parseIdeasResponse(responseText)
}

// GenerateDeepDive creates a detailed MVP specification for an idea
func (c *Client) GenerateDeepDive(ctx context.Context, idea *db.Idea) (string, error) {
	prompt := c.buildDeepDivePrompt(idea)

	maxTokens := int64(8192)
	systemPrompt := deepDiveSystemPrompt
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.model),
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		System: []anthropic.TextBlockParam{
			{Type: "text", Text: systemPrompt},
		},
	})

	if err != nil {
		return "", fmt.Errorf("claude API error: %w", err)
	}

	var responseText string
	for _, block := range message.Content {
		if block.Type == "text" {
			responseText = block.Text
			break
		}
	}

	return responseText, nil
}

func (c *Client) buildDeepDivePrompt(idea *db.Idea) string {
	ideaJSON, _ := json.MarshalIndent(idea, "", "  ")
	return fmt.Sprintf(deepDiveUserPrompt, string(ideaJSON))
}

func (c *Client) generateCacheKey(contentHash, operation string) string {
	data := fmt.Sprintf("%s:%s:%s:%s", contentHash, operation, c.model, c.promptVersion)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func mustMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}
