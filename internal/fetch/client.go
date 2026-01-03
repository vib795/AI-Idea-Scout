package fetch

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// HTTPClient wraps http.Client with rate limiting, retries, and caching
type HTTPClient struct {
	client      *http.Client
	rateLimiter *RateLimiter
	cache       *Cache
	userAgent   string
	maxRetries  int
}

func NewHTTPClient(userAgent string, reqPerSecond float64, cacheEnabled bool) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
			},
		},
		rateLimiter: NewRateLimiter(reqPerSecond),
		cache:       NewCache(cacheEnabled, 1*time.Hour),
		userAgent:   userAgent,
		maxRetries:  3,
	}
}

func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	// Check cache first
	if cached := c.cache.Get(url); cached != nil {
		return cached, nil
	}

	// Rate limit
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.calculateBackoff(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		body, err := c.doRequest(ctx, url)
		if err == nil {
			// Cache successful response
			c.cache.Set(url, body)
			return body, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		if isClientError(err) {
			break
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *HTTPClient) doRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, text/html, */*")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}

	return io.ReadAll(resp.Body)
}

func (c *HTTPClient) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff with jitter
	base := math.Pow(2, float64(attempt))
	backoff := time.Duration(base) * time.Second
	jitter := time.Duration(rand.Float64() * float64(time.Second))
	return backoff + jitter
}

func isClientError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode >= 400 && httpErr.StatusCode < 500
	}
	return false
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

// RateLimiter implements token bucket algorithm
type RateLimiter struct {
	mu            sync.Mutex
	tokens        float64
	maxTokens     float64
	refillRate    float64 // tokens per second
	lastRefill    time.Time
}

func NewRateLimiter(reqPerSecond float64) *RateLimiter {
	return &RateLimiter{
		tokens:     reqPerSecond,
		maxTokens:  reqPerSecond,
		refillRate: reqPerSecond,
		lastRefill: time.Now(),
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		rl.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(rl.lastRefill).Seconds()
		rl.tokens = math.Min(rl.maxTokens, rl.tokens+elapsed*rl.refillRate)
		rl.lastRefill = now

		if rl.tokens >= 1.0 {
			rl.tokens -= 1.0
			rl.mu.Unlock()
			return nil
		}

		waitTime := time.Duration((1.0-rl.tokens)/rl.refillRate) * time.Second
		rl.mu.Unlock()

		select {
		case <-time.After(waitTime):
			// Continue loop
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Cache provides simple in-memory caching with TTL
type Cache struct {
	mu      sync.RWMutex
	enabled bool
	ttl     time.Duration
	data    map[string]*cacheEntry
}

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewCache(enabled bool, ttl time.Duration) *Cache {
	c := &Cache{
		enabled: enabled,
		ttl:     ttl,
		data:    make(map[string]*cacheEntry),
	}

	if enabled {
		go c.cleanup()
	}

	return c
}

func (c *Cache) Get(key string) []byte {
	if !c.enabled {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.data
}

func (c *Cache) Set(key string, data []byte) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.data {
			if now.After(entry.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}
