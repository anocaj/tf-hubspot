package client

import (
	"context"
	"fmt"
	"time"
)

// Contact represents a HubSpot contact
type Contact struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
	Archived   bool                   `json:"archived"`
}

// ContactRequest represents the request body for creating/updating contacts
type ContactRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

// ContactSearchRequest represents a search request for contacts
type ContactSearchRequest struct {
	FilterGroups []FilterGroup `json:"filterGroups"`
	Properties   []string      `json:"properties,omitempty"`
}

// FilterGroup represents a group of filters
type FilterGroup struct {
	Filters []Filter `json:"filters"`
}

// Filter represents a single filter condition
type Filter struct {
	PropertyName string `json:"propertyName"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
}

// ContactSearchResponse represents the response from a contact search
type ContactSearchResponse struct {
	Results []Contact `json:"results"`
	Total   int       `json:"total"`
}

// CreateContact creates a new contact in HubSpot
func (c *Client) CreateContact(ctx context.Context, properties map[string]interface{}) (*Contact, error) {
	reqBody := ContactRequest{
		Properties: properties,
	}

	resp, err := c.Post(ctx, "crm/v3/objects/contacts", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	var contact Contact
	if err := DecodeResponse(resp, &contact); err != nil {
		return nil, fmt.Errorf("failed to decode contact response: %w", err)
	}

	return &contact, nil
}

// GetContact retrieves a contact by ID
func (c *Client) GetContact(ctx context.Context, id string) (*Contact, error) {
	path := fmt.Sprintf("crm/v3/objects/contacts/%s", id)
	
	resp, err := c.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	var contact Contact
	if err := DecodeResponse(resp, &contact); err != nil {
		return nil, fmt.Errorf("failed to decode contact response: %w", err)
	}

	return &contact, nil
}

// UpdateContact updates an existing contact
func (c *Client) UpdateContact(ctx context.Context, id string, properties map[string]interface{}) (*Contact, error) {
	path := fmt.Sprintf("crm/v3/objects/contacts/%s", id)
	
	reqBody := ContactRequest{
		Properties: properties,
	}

	resp, err := c.Patch(ctx, path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	var contact Contact
	if err := DecodeResponse(resp, &contact); err != nil {
		return nil, fmt.Errorf("failed to decode contact response: %w", err)
	}

	return &contact, nil
}

// DeleteContact deletes a contact by ID
func (c *Client) DeleteContact(ctx context.Context, id string) error {
	path := fmt.Sprintf("crm/v3/objects/contacts/%s", id)
	
	resp, err := c.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// GetContactByEmail retrieves a contact by email address using the search API
func (c *Client) GetContactByEmail(ctx context.Context, email string) (*Contact, error) {
	searchReq := ContactSearchRequest{
		FilterGroups: []FilterGroup{
			{
				Filters: []Filter{
					{
						PropertyName: "email",
						Operator:     "EQ",
						Value:        email,
					},
				},
			},
		},
	}

	resp, err := c.Post(ctx, "crm/v3/objects/contacts/search", searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to search contact by email: %w", err)
	}

	var searchResp ContactSearchResponse
	if err := DecodeResponse(resp, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	if searchResp.Total == 0 || len(searchResp.Results) == 0 {
		return nil, &HubSpotError{
			Status:  "404",
			Message: fmt.Sprintf("contact with email %s not found", email),
		}
	}

	return &searchResp.Results[0], nil
}
