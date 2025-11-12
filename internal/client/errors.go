package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HubSpotError represents an error response from the HubSpot API
type HubSpotError struct {
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Category    string                 `json:"category"`
	SubCategory string                 `json:"subCategory,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	StatusCode  int                    `json:"-"`
}

// Error implements the error interface
func (e *HubSpotError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("HubSpot API error (HTTP %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("HubSpot API error (%s): %s", e.Status, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found error
func (e *HubSpotError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if the error is a 429 Rate Limit error
func (e *HubSpotError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServerError returns true if the error is a 5xx server error
func (e *HubSpotError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsAuthError returns true if the error is a 401 Unauthorized error
func (e *HubSpotError) IsAuthError() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// parseErrorResponse parses an HTTP error response into a HubSpotError
func parseErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HubSpotError{
			Status:     "error",
			Message:    fmt.Sprintf("Failed to read error response: %v", err),
			StatusCode: resp.StatusCode,
		}
	}

	var hubspotErr HubSpotError
	if err := json.Unmarshal(body, &hubspotErr); err != nil {
		// If we can't parse the error response, return a generic error
		return &HubSpotError{
			Status:     "error",
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	hubspotErr.StatusCode = resp.StatusCode
	return &hubspotErr
}
