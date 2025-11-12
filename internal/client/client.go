package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a HubSpot API client
type Client struct {
	apiToken    string
	baseURL     string
	apiVersion  string
	httpClient  *http.Client
	retryConfig RetryConfig
}

// Config holds the configuration for creating a new Client
type Config struct {
	APIToken   string
	BaseURL    string
	APIVersion string
	Timeout    time.Duration
}

// NewClient creates a new HubSpot API client
func NewClient(config Config) *Client {
	// Set defaults
	if config.BaseURL == "" {
		config.BaseURL = "https://api.hubapi.com"
	}
	if config.APIVersion == "" {
		config.APIVersion = "v3"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		apiToken:   config.APIToken,
		baseURL:    config.BaseURL,
		apiVersion: config.APIVersion,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		retryConfig: DefaultRetryConfig(),
	}
}

// SetRetryConfig allows customizing the retry configuration
func (c *Client) SetRetryConfig(config RetryConfig) {
	c.retryConfig = config
}

// buildURL constructs a full API URL from a path
func (c *Client) buildURL(path string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, path)
}

// addAuthHeader adds the authentication header to the request
func (c *Client) addAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
}

// addCommonHeaders adds common headers to all requests
func (c *Client) addCommonHeaders(req *http.Request) {
	c.addAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
}

// doRequest executes an HTTP request with retry logic and error handling
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.buildURL(path)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.addCommonHeaders(req)

	resp, err := c.doWithRetry(ctx, req)
	if err != nil {
		return nil, err
	}

	// Check for HTTP errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, parseErrorResponse(resp)
	}

	return resp, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil)
}

// DecodeResponse decodes a JSON response into the provided interface
func DecodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
