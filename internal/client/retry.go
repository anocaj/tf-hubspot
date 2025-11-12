package client

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// RetryConfig defines the configuration for retry logic
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// shouldRetry determines if a request should be retried based on the response
func shouldRetry(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on rate limit errors (429)
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}

	// Retry on server errors (5xx)
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		return true
	}

	return false
}

// getRetryAfter extracts the Retry-After header value from the response
func getRetryAfter(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	// Try to parse as seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try to parse as HTTP date
	if t, err := http.ParseTime(retryAfter); err == nil {
		return time.Until(t)
	}

	return 0
}

// calculateBackoff calculates the backoff duration for a retry attempt
func calculateBackoff(attempt int, config RetryConfig, resp *http.Response) time.Duration {
	// If we have a Retry-After header, use it
	if resp != nil {
		if retryAfter := getRetryAfter(resp); retryAfter > 0 {
			return retryAfter
		}
	}

	// Calculate exponential backoff
	backoff := float64(config.InitialBackoff) * math.Pow(config.Multiplier, float64(attempt))
	
	// Add jitter (random value between 0 and 25% of backoff)
	jitter := rand.Float64() * backoff * 0.25
	backoff += jitter

	// Cap at max backoff
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	return time.Duration(backoff)
}

// doWithRetry executes an HTTP request with retry logic
func (c *Client) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Execute the request
		resp, err = c.httpClient.Do(req.WithContext(ctx))

		// If successful (2xx), return immediately
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Check if we should retry
		if !shouldRetry(resp, err) {
			// Don't retry, return the error
			if err != nil {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			return resp, nil
		}

		// If this was the last attempt, return the error
		if attempt == c.retryConfig.MaxRetries {
			if err != nil {
				return nil, fmt.Errorf("request failed after %d retries: %w", c.retryConfig.MaxRetries, err)
			}
			return resp, nil
		}

		// Calculate backoff and wait
		backoff := calculateBackoff(attempt, c.retryConfig, resp)
		
		// Close the response body before retrying
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		// Wait for backoff duration
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return resp, err
}
