package celestrak

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Format string

const (
	FormatTLE        Format = "TLE"
	Format3LE        Format = "3LE"
	Format2LE        Format = "2LE"
	FormatXML        Format = "XML"
	FormatKVN        Format = "KVN"
	FormatJSON       Format = "JSON"
	FormatJSONPretty Format = "JSON-PRETTY"
	FormatCSV        Format = "CSV"
)

const defaultBaseURL = "https://celestrak.org"

// Cache interface for ETag-aware caching.
type Cache interface {
	// Returns cached bytes
	Get(key string) (data []byte, etag string, ok bool)

	// Stores bytes
	Put(key string, data []byte, etag string)
}

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	userAgent  string
	cache      Cache

	// Retry configuration
	maxRetries int           // Maximum number of retries (default: 3)
	retryDelay time.Duration // Initial retry delay (default: 1s)
}

func NewClient(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	u, err := url.Parse(defaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid default base URL: %w", err)
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    u,
		userAgent:  "celestrak-go/0.0.1",
		maxRetries: 3,
		retryDelay: 1 * time.Second,
	}, nil
}

// WithRetries configures retry behavior for transient failures.
// maxRetries: maximum number of retry attempts (default: 3)
// retryDelay: initial delay between retries, uses exponential backoff (default: 1s)
func (c *Client) WithRetries(maxRetries int, retryDelay time.Duration) *Client {
	c.maxRetries = maxRetries
	c.retryDelay = retryDelay
	return c
}

// WithCache sets an optional cache (ETag-aware).
func (c *Client) WithCache(cache Cache) *Client {
	c.cache = cache
	return c
}

// WithUserAgent sets the User-Agent header.
func (c *Client) WithUserAgent(ua string) *Client {
	c.userAgent = ua
	return c
}

const (
	// maxResponseSize limits response size to 100MB to prevent memory exhaustion
	maxResponseSize = 100 * 1024 * 1024
)

// shouldRetry determines if an error is retryable.
func shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Don't retry query errors (client-side validation issues)
	if IsQueryError(err) {
		return false
	}

	// Retry on network errors (connection refused, timeout, etc.)
	// These are typically transient

	// Check for server errors (5xx) - these are retryable
	if errResp, ok := err.(*ErrorResponse); ok {
		return errResp.IsServerError()
	}

	// Retry on context deadline exceeded (might be transient network issue)
	// But not on context cancelled (user intent)

	// For other errors (network failures, etc.), retry
	return true
}

// fetchOnce performs a single fetch attempt.
func (c *Client) fetchOnce(ctx context.Context, q Query, endpoint string) ([]byte, error) {
	if ctx == nil {
		return nil, &QueryError{Message: "context must be non-nil"}
	}

	// Check context before expensive operations
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	fullURL, err := q.BuildURL(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	// Cache key = full URL (deterministic and already encodes params)
	cacheKey := fullURL

	var cached []byte
	var etag string
	var hasCache bool
	if c.cache != nil {
		cached, etag, hasCache = c.cache.Get(cacheKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if hasCache && etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check if error is due to context cancellation/timeout
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 304: use cached body
	if resp.StatusCode == http.StatusNotModified {
		if hasCache {
			return cached, nil
		}
		// If server says not modified but we don't have cached bytes, treat as error.
		return nil, &ErrorResponse{
			Response: resp,
			Message:  "304 Not Modified but no cached body available",
		}
	}

	// Non-2xx: return structured error
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		message := strings.TrimSpace(string(b))
		if message == "" {
			message = resp.Status
		}
		return nil, &ErrorResponse{
			Response: resp,
			Message:  message,
		}
	}

	// Read response with size limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check if response was truncated (indicates response was too large)
	if len(body) == maxResponseSize {
		// Try to read one more byte to see if there's more data
		var extra [1]byte
		if n, _ := resp.Body.Read(extra[:]); n > 0 {
			return nil, &ErrorResponse{
				Response: resp,
				Message:  fmt.Sprintf("response too large (exceeds %d bytes)", maxResponseSize),
			}
		}
	}

	// Validate non-empty response
	if len(body) == 0 {
		return nil, &ErrorResponse{
			Response: resp,
			Message:  "empty response body",
		}
	}

	if c.cache != nil {
		newETag := strings.TrimSpace(resp.Header.Get("ETag"))
		c.cache.Put(cacheKey, body, newETag)
	}

	return body, nil
}

// fetch fetches data from the specified endpoint as raw bytes with automatic retries.
// This is the internal method used by all public Fetch methods.
func (c *Client) fetch(ctx context.Context, q Query, endpoint string) ([]byte, error) {
	var lastErr error
	delay := c.retryDelay

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Check context before each attempt
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled: %w", err)
		}

		// Wait before retry (skip on first attempt)
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(delay):
				// Exponential backoff: double the delay each time
				delay *= 2
			}
		}

		data, err := c.fetchOnce(ctx, q, endpoint)
		if err == nil {
			return data, nil
		}

		lastErr = err

		// Don't retry if error is not retryable
		if !shouldRetry(err) {
			return nil, err
		}
	}

	// All retries exhausted
	return nil, fmt.Errorf("max retries (%d) exceeded: %w", c.maxRetries, lastErr)
}
