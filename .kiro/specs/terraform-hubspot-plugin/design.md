# Design Document: Terraform HubSpot Provider

## Overview

The Terraform HubSpot Provider is a plugin that enables infrastructure-as-code management of HubSpot resources. Built using the Terraform Plugin Framework (the modern successor to Terraform Plugin SDK), it provides a declarative interface for managing HubSpot contacts, companies, deals, custom properties, and other resources through Terraform configuration files.

The provider will be written in Go, following Terraform's plugin architecture and best practices. It will communicate with the HubSpot REST API v3 and handle authentication, rate limiting, error handling, and state management.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Terraform Core                          │
│                  (Plan, Apply, State)                       │
└────────────────────────┬────────────────────────────────────┘
                         │ gRPC Protocol
                         │
┌────────────────────────▼────────────────────────────────────┐
│              Terraform HubSpot Provider                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Provider Configuration & Authentication             │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Resources (Contact, Company, Deal, Property)        │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Data Sources (Contact, Company, Deal, Property)     │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  HubSpot API Client (HTTP, Auth, Rate Limiting)     │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS
                         │
┌────────────────────────▼────────────────────────────────────┐
│                   HubSpot REST API v3                       │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Terraform Plugin Framework v1.4+
- **HTTP Client**: Standard Go net/http with custom retry logic
- **Testing**: Go testing package with Terraform acceptance test framework
- **Build**: Go modules for dependency management

### Directory Structure

```
terraform-provider-hubspot/
├── main.go                          # Provider entry point
├── go.mod                           # Go module definition
├── go.sum                           # Go module checksums
├── internal/
│   ├── provider/
│   │   ├── provider.go              # Provider implementation
│   │   ├── provider_test.go         # Provider tests
│   │   └── config.go                # Provider configuration
│   ├── resources/
│   │   ├── contact_resource.go      # Contact resource
│   │   ├── company_resource.go      # Company resource
│   │   ├── deal_resource.go         # Deal resource
│   │   └── property_resource.go     # Property resource
│   ├── datasources/
│   │   ├── contact_data_source.go   # Contact data source
│   │   ├── company_data_source.go   # Company data source
│   │   ├── deal_data_source.go      # Deal data source
│   │   └── property_data_source.go  # Property data source
│   └── client/
│       ├── client.go                # HubSpot API client
│       ├── contacts.go              # Contact API methods
│       ├── companies.go             # Company API methods
│       ├── deals.go                 # Deal API methods
│       ├── properties.go            # Property API methods
│       ├── retry.go                 # Retry logic
│       └── errors.go                # Error handling
├── examples/
│   ├── provider/
│   │   └── provider.tf              # Provider configuration examples
│   └── resources/
│       ├── contact.tf               # Contact resource examples
│       ├── company.tf               # Company resource examples
│       └── deal.tf                  # Deal resource examples
└── docs/
    ├── index.md                     # Provider documentation
    ├── resources/
    │   ├── contact.md               # Contact resource docs
    │   ├── company.md               # Company resource docs
    │   ├── deal.md                  # Deal resource docs
    │   └── property.md              # Property resource docs
    └── data-sources/
        ├── contact.md               # Contact data source docs
        ├── company.md               # Company data source docs
        ├── deal.md                  # Deal data source docs
        └── property.md              # Property data source docs
```

## Components and Interfaces

### 1. Provider Configuration

The provider configuration handles authentication and global settings.

**Configuration Schema:**
```hcl
provider "hubspot" {
  api_token   = "your-api-token"           # Required, can use env var HUBSPOT_API_TOKEN
  api_url     = "https://api.hubapi.com"   # Optional, defaults to production
  api_version = "v3"                       # Optional, defaults to v3
}
```

**Go Interface:**
```go
type HubSpotProviderModel struct {
    APIToken   types.String `tfsdk:"api_token"`
    APIURL     types.String `tfsdk:"api_url"`
    APIVersion types.String `tfsdk:"api_version"`
}
```

### 2. HubSpot API Client

A centralized client for all HubSpot API interactions.

**Client Interface:**
```go
type Client interface {
    // Contacts
    CreateContact(ctx context.Context, properties map[string]interface{}) (*Contact, error)
    GetContact(ctx context.Context, id string) (*Contact, error)
    UpdateContact(ctx context.Context, id string, properties map[string]interface{}) (*Contact, error)
    DeleteContact(ctx context.Context, id string) error
    GetContactByEmail(ctx context.Context, email string) (*Contact, error)
    
    // Companies
    CreateCompany(ctx context.Context, properties map[string]interface{}) (*Company, error)
    GetCompany(ctx context.Context, id string) (*Company, error)
    UpdateCompany(ctx context.Context, id string, properties map[string]interface{}) (*Company, error)
    DeleteCompany(ctx context.Context, id string) error
    GetCompanyByDomain(ctx context.Context, domain string) (*Company, error)
    
    // Deals
    CreateDeal(ctx context.Context, properties map[string]interface{}) (*Deal, error)
    GetDeal(ctx context.Context, id string) (*Deal, error)
    UpdateDeal(ctx context.Context, id string, properties map[string]interface{}) (*Deal, error)
    DeleteDeal(ctx context.Context, id string) error
    
    // Properties
    CreateProperty(ctx context.Context, objectType string, property *PropertyDefinition) (*PropertyDefinition, error)
    GetProperty(ctx context.Context, objectType, propertyName string) (*PropertyDefinition, error)
    UpdateProperty(ctx context.Context, objectType, propertyName string, property *PropertyDefinition) (*PropertyDefinition, error)
    DeleteProperty(ctx context.Context, objectType, propertyName string) error
    
    // Associations
    CreateAssociation(ctx context.Context, fromObjectType, fromID, toObjectType, toID, associationType string) error
    DeleteAssociation(ctx context.Context, fromObjectType, fromID, toObjectType, toID, associationType string) error
}
```

**Client Implementation Details:**
- Uses Go's `net/http` package with custom transport for logging
- Implements exponential backoff for rate limiting (429 responses)
- Retries transient errors (5xx responses) up to 3 times
- Adds authentication header to all requests
- Marshals/unmarshals JSON payloads
- Provides context-aware cancellation

### 3. Resource Implementations

Each resource implements the Terraform Plugin Framework's `resource.Resource` interface.

**Contact Resource Schema:**
```hcl
resource "hubspot_contact" "example" {
  email     = "contact@example.com"
  firstname = "John"
  lastname  = "Doe"
  
  properties = {
    phone = "+1234567890"
    company = "Example Corp"
  }
}
```

**Contact Resource Model:**
```go
type ContactResourceModel struct {
    ID         types.String `tfsdk:"id"`
    Email      types.String `tfsdk:"email"`
    Firstname  types.String `tfsdk:"firstname"`
    Lastname   types.String `tfsdk:"lastname"`
    Properties types.Map    `tfsdk:"properties"`
}
```

**Company Resource Schema:**
```hcl
resource "hubspot_company" "example" {
  name   = "Example Corp"
  domain = "example.com"
  
  properties = {
    city = "San Francisco"
    state = "CA"
  }
}
```

**Deal Resource Schema:**
```hcl
resource "hubspot_deal" "example" {
  dealname  = "New Deal"
  amount    = "10000"
  dealstage = "appointmentscheduled"
  pipeline  = "default"
  
  properties = {
    closedate = "2025-12-31"
  }
  
  associations = {
    contacts  = [hubspot_contact.example.id]
    companies = [hubspot_company.example.id]
  }
}
```

**Property Resource Schema:**
```hcl
resource "hubspot_property" "example" {
  object_type = "contacts"
  name        = "custom_field"
  label       = "Custom Field"
  type        = "string"
  field_type  = "text"
  group_name  = "contactinformation"
  
  description = "A custom field for contacts"
}
```

### 4. Data Source Implementations

Data sources allow reading existing HubSpot resources without managing them.

**Contact Data Source Schema:**
```hcl
data "hubspot_contact" "example" {
  email = "contact@example.com"
}

output "contact_id" {
  value = data.hubspot_contact.example.id
}
```

**Company Data Source Schema:**
```hcl
data "hubspot_company" "example" {
  domain = "example.com"
}
```

## Data Models

### Contact Model
```go
type Contact struct {
    ID         string                 `json:"id"`
    Properties map[string]interface{} `json:"properties"`
    CreatedAt  time.Time              `json:"createdAt"`
    UpdatedAt  time.Time              `json:"updatedAt"`
    Archived   bool                   `json:"archived"`
}
```

### Company Model
```go
type Company struct {
    ID         string                 `json:"id"`
    Properties map[string]interface{} `json:"properties"`
    CreatedAt  time.Time              `json:"createdAt"`
    UpdatedAt  time.Time              `json:"updatedAt"`
    Archived   bool                   `json:"archived"`
}
```

### Deal Model
```go
type Deal struct {
    ID           string                 `json:"id"`
    Properties   map[string]interface{} `json:"properties"`
    Associations []Association          `json:"associations,omitempty"`
    CreatedAt    time.Time              `json:"createdAt"`
    UpdatedAt    time.Time              `json:"updatedAt"`
    Archived     bool                   `json:"archived"`
}
```

### Property Definition Model
```go
type PropertyDefinition struct {
    Name        string   `json:"name"`
    Label       string   `json:"label"`
    Type        string   `json:"type"`
    FieldType   string   `json:"fieldType"`
    GroupName   string   `json:"groupName"`
    Description string   `json:"description,omitempty"`
    Options     []Option `json:"options,omitempty"`
}

type Option struct {
    Label string `json:"label"`
    Value string `json:"value"`
}
```

### Association Model
```go
type Association struct {
    FromObjectType string `json:"from_object_type"`
    FromID         string `json:"from_id"`
    ToObjectType   string `json:"to_object_type"`
    ToID           string `json:"to_id"`
    Type           string `json:"type"`
}
```

## Error Handling

### Error Types

1. **Authentication Errors (401)**
   - Return clear message about invalid API token
   - Suggest checking credentials and permissions

2. **Rate Limit Errors (429)**
   - Implement exponential backoff with jitter
   - Respect `Retry-After` header from HubSpot
   - Maximum 5 retry attempts

3. **Validation Errors (400)**
   - Parse HubSpot error response for specific field errors
   - Return detailed validation messages to user

4. **Not Found Errors (404)**
   - Mark resource as deleted in Terraform state
   - Allow Terraform to recreate if needed

5. **Server Errors (5xx)**
   - Retry up to 3 times with exponential backoff
   - Return error if all retries fail

### Error Response Structure

```go
type HubSpotError struct {
    Status    string `json:"status"`
    Message   string `json:"message"`
    Category  string `json:"category"`
    SubCategory string `json:"subCategory,omitempty"`
    Context   map[string]interface{} `json:"context,omitempty"`
}

func (e *HubSpotError) Error() string {
    return fmt.Sprintf("HubSpot API error (%s): %s", e.Status, e.Message)
}
```

### Retry Logic

```go
type RetryConfig struct {
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    Multiplier     float64
}

// Default: 3 retries, 1s initial, 30s max, 2x multiplier
```

## Testing Strategy

### Unit Tests

- Test each API client method with mocked HTTP responses
- Test resource CRUD operations with mocked client
- Test data source read operations with mocked client
- Test error handling and retry logic
- Test validation functions
- Target: 80%+ code coverage

### Integration Tests

- Test against HubSpot sandbox/test account
- Verify actual API interactions
- Test rate limiting behavior
- Test authentication flows

### Acceptance Tests

- Use Terraform's acceptance test framework
- Test full resource lifecycle (create, read, update, delete)
- Test import functionality
- Test data source lookups
- Test resource dependencies and associations
- Run against real HubSpot test account

**Example Acceptance Test Structure:**
```go
func TestAccContactResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccContactResourceConfig("test@example.com"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("hubspot_contact.test", "email", "test@example.com"),
                    resource.TestCheckResourceAttrSet("hubspot_contact.test", "id"),
                ),
            },
            // Update testing
            {
                Config: testAccContactResourceConfig("updated@example.com"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("hubspot_contact.test", "email", "updated@example.com"),
                ),
            },
            // Import testing
            {
                ResourceName:      "hubspot_contact.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

### Manual Testing

- Test with real Terraform configurations
- Verify documentation examples work correctly
- Test edge cases and error scenarios
- Performance testing with large numbers of resources

## Security Considerations

1. **API Token Storage**
   - Never log API tokens
   - Support environment variables for sensitive data
   - Mark API token as sensitive in schema

2. **TLS/HTTPS**
   - All API communication over HTTPS
   - Verify SSL certificates

3. **Input Validation**
   - Validate all user inputs before API calls
   - Sanitize error messages to avoid leaking sensitive data

4. **Rate Limiting**
   - Respect HubSpot API rate limits
   - Implement client-side throttling if needed

## Performance Considerations

1. **Batch Operations**
   - Use HubSpot batch APIs where available
   - Reduce number of API calls for bulk operations

2. **Caching**
   - Cache property definitions during provider initialization
   - Avoid redundant API calls for static data

3. **Concurrent Operations**
   - Terraform handles parallelism
   - Ensure client is thread-safe
   - Use proper locking for shared state

4. **Pagination**
   - Handle paginated responses for list operations
   - Implement efficient pagination for large result sets

## Documentation Requirements

1. **Provider Documentation**
   - Installation instructions
   - Authentication setup
   - Configuration reference
   - Environment variables

2. **Resource Documentation**
   - Schema reference for each resource
   - Example configurations
   - Import instructions
   - Attribute descriptions

3. **Data Source Documentation**
   - Schema reference for each data source
   - Example usage
   - Attribute descriptions

4. **Guides**
   - Getting started guide
   - Common use cases
   - Migration from manual HubSpot management
   - Troubleshooting guide

## Future Enhancements

- Support for additional HubSpot objects (tickets, products, line items)
- Workflow management
- Email template management
- Form management
- List management
- Custom object support
- Bulk import/export utilities
- Terraform Cloud integration
