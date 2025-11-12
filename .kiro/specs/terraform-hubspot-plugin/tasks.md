# Implementation Plan

- [x] 1. Initialize project structure and dependencies
  - Create Go module with `go mod init terraform-provider-hubspot`
  - Add Terraform Plugin Framework dependency (v1.4+)
  - Create directory structure: internal/provider, internal/resources, internal/datasources, internal/client
  - Create main.go entry point with provider registration
  - Set up .gitignore for Go projects
  - _Requirements: All requirements depend on proper project setup_

- [x] 2. Implement HubSpot API client foundation
  - Create client package with base HTTP client configuration
  - Implement authentication header injection for API token
  - Create error types and error parsing for HubSpot API responses
  - Implement retry logic with exponential backoff for rate limiting (429) and server errors (5xx)
  - Add context support for cancellation and timeouts
  - _Requirements: 1.3, 7.1, 7.3, 7.4_

- [ ]* 2.1 Write unit tests for client error handling and retry logic
  - Create mock HTTP responses for various error scenarios
  - Test retry behavior with 429 and 5xx responses
  - Test authentication error handling
  - _Requirements: 7.1, 7.3, 7.4_

- [x] 3. Implement provider configuration
  - Create provider schema with api_token, api_url, and api_version fields
  - Mark api_token as sensitive in schema
  - Implement environment variable support for HUBSPOT_API_TOKEN
  - Add validation for API token format
  - Initialize HubSpot client with provider configuration
  - Implement Configure method to validate credentials
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 9.1, 9.2, 9.3, 9.4_

- [ ]* 3.1 Write unit tests for provider configuration
  - Test provider initialization with valid configuration
  - Test environment variable configuration
  - Test validation errors for invalid credentials
  - _Requirements: 1.3, 1.4, 1.5_

- [x] 4. Implement Contact API client methods
  - Create Contact model struct with JSON tags
  - Implement CreateContact method with POST to /crm/v3/objects/contacts
  - Implement GetContact method with GET to /crm/v3/objects/contacts/{id}
  - Implement UpdateContact method with PATCH to /crm/v3/objects/contacts/{id}
  - Implement DeleteContact method with DELETE to /crm/v3/objects/contacts/{id}
  - Implement GetContactByEmail method with search API
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 7.5_

- [ ]* 4.1 Write unit tests for Contact API methods
  - Mock HTTP responses for each CRUD operation
  - Test JSON marshaling/unmarshaling
  - Test error handling for each method
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [x] 5. Implement Contact resource
  - Create contact_resource.go with resource schema (id, email, firstname, lastname, properties)
  - Implement Create method calling client.CreateContact
  - Implement Read method calling client.GetContact with 404 handling
  - Implement Update method calling client.UpdateContact
  - Implement Delete method calling client.DeleteContact
  - Implement ImportState method for importing by contact ID
  - Add email format validation in schema validators
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 7.5, 8.1, 10.1, 10.2, 10.4_

- [ ]* 5.1 Write acceptance tests for Contact resource
  - Test create and read operations
  - Test update operations
  - Test delete operations
  - Test import functionality
  - _Requirements: 2.1, 2.2, 2.3, 8.1_

- [ ] 6. Implement Company API client methods
  - Create Company model struct with JSON tags
  - Implement CreateCompany method with POST to /crm/v3/objects/companies
  - Implement GetCompany method with GET to /crm/v3/objects/companies/{id}
  - Implement UpdateCompany method with PATCH to /crm/v3/objects/companies/{id}
  - Implement DeleteCompany method with DELETE to /crm/v3/objects/companies/{id}
  - Implement GetCompanyByDomain method with search API
  - _Requirements: 3.1, 3.3, 3.4, 3.5_

- [ ]* 6.1 Write unit tests for Company API methods
  - Mock HTTP responses for each CRUD operation
  - Test property validation
  - Test error handling
  - _Requirements: 3.1, 3.3, 3.5_

- [ ] 7. Implement Company resource
  - Create company_resource.go with resource schema (id, name, domain, properties)
  - Implement Create method calling client.CreateCompany
  - Implement Read method calling client.GetCompany
  - Implement Update method calling client.UpdateCompany with partial updates
  - Implement Delete method calling client.DeleteCompany
  - Implement ImportState method for importing by company ID
  - Add domain/URL format validation in schema validators
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 8.2, 10.1, 10.2, 10.5_

- [ ]* 7.1 Write acceptance tests for Company resource
  - Test create and read operations
  - Test update with partial property changes
  - Test delete operations
  - Test import functionality
  - _Requirements: 3.1, 3.3, 3.4, 8.2_

- [ ] 8. Implement Deal API client methods
  - Create Deal model struct with JSON tags
  - Implement CreateDeal method with POST to /crm/v3/objects/deals
  - Implement GetDeal method with GET to /crm/v3/objects/deals/{id}
  - Implement UpdateDeal method with PATCH to /crm/v3/objects/deals/{id}
  - Implement DeleteDeal method with DELETE to /crm/v3/objects/deals/{id}
  - Implement association methods for linking deals to contacts and companies
  - _Requirements: 4.1, 4.2, 4.5_

- [ ]* 8.1 Write unit tests for Deal API methods
  - Mock HTTP responses for CRUD operations
  - Test association creation
  - Test error handling
  - _Requirements: 4.1, 4.2, 4.5_

- [ ] 9. Implement Deal resource
  - Create deal_resource.go with resource schema (id, dealname, amount, dealstage, pipeline, properties, associations)
  - Implement Create method calling client.CreateDeal and creating associations
  - Implement Read method calling client.GetDeal
  - Implement Update method calling client.UpdateDeal and updating associations
  - Implement Delete method calling client.DeleteDeal
  - Implement ImportState method for importing by deal ID
  - Add validation for dealstage existence in pipeline
  - Add validation error for non-existent pipeline references
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 8.3, 10.1, 10.2, 10.3_

- [ ]* 9.1 Write acceptance tests for Deal resource
  - Test create with associations
  - Test update operations
  - Test dealstage validation
  - Test import functionality
  - _Requirements: 4.1, 4.2, 4.3, 8.3_

- [ ] 10. Implement Property API client methods
  - Create PropertyDefinition model struct with JSON tags
  - Implement CreateProperty method with POST to /crm/v3/properties/{objectType}
  - Implement GetProperty method with GET to /crm/v3/properties/{objectType}/{propertyName}
  - Implement UpdateProperty method with PATCH to /crm/v3/properties/{objectType}/{propertyName}
  - Implement DeleteProperty method with DELETE to /crm/v3/properties/{objectType}/{propertyName}
  - _Requirements: 5.1, 5.2, 5.5_

- [ ]* 10.1 Write unit tests for Property API methods
  - Mock HTTP responses for CRUD operations
  - Test property type handling
  - Test error handling
  - _Requirements: 5.1, 5.2_

- [ ] 11. Implement Property resource
  - Create property_resource.go with resource schema (object_type, name, label, type, field_type, group_name, description)
  - Implement Create method calling client.CreateProperty
  - Implement Read method calling client.GetProperty
  - Implement Update method calling client.UpdateProperty
  - Implement Delete method calling client.DeleteProperty with in-use validation
  - Implement ImportState method using object_type and property name
  - Add ForceNew for type changes to trigger recreation
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 8.4_

- [ ]* 11.1 Write acceptance tests for Property resource
  - Test property creation for different object types
  - Test property type changes trigger recreation
  - Test property group assignment
  - Test import functionality
  - _Requirements: 5.1, 5.2, 5.4, 5.5, 8.4_

- [ ] 12. Implement Contact data source
  - Create contact_data_source.go with schema (email as input, all contact fields as outputs)
  - Implement Read method calling client.GetContactByEmail
  - Add email format validation
  - Handle not found errors gracefully
  - _Requirements: 6.1, 6.2, 10.4_

- [ ]* 12.1 Write acceptance tests for Contact data source
  - Test lookup by email
  - Test not found handling
  - Test data source output attributes
  - _Requirements: 6.1, 6.2_

- [ ] 13. Implement Company data source
  - Create company_data_source.go with schema (domain as input, all company fields as outputs)
  - Implement Read method calling client.GetCompanyByDomain
  - Add domain format validation
  - Handle not found errors gracefully
  - _Requirements: 6.1, 6.3, 10.5_

- [ ]* 13.1 Write acceptance tests for Company data source
  - Test lookup by domain
  - Test not found handling
  - Test data source output attributes
  - _Requirements: 6.1, 6.3_

- [ ] 14. Implement Deal data source
  - Create deal_data_source.go with schema (deal_id as input, all deal fields as outputs)
  - Implement Read method calling client.GetDeal
  - Handle not found errors gracefully
  - _Requirements: 6.1, 6.4_

- [ ]* 14.1 Write acceptance tests for Deal data source
  - Test lookup by deal ID
  - Test not found handling
  - Test data source output attributes
  - _Requirements: 6.1, 6.4_

- [ ] 15. Implement Property data source
  - Create property_data_source.go with schema (object_type and name as inputs, property definition as outputs)
  - Implement Read method calling client.GetProperty
  - Handle not found errors gracefully
  - _Requirements: 6.1, 6.5_

- [ ]* 15.1 Write acceptance tests for Property data source
  - Test lookup by object type and property name
  - Test not found handling
  - Test data source output attributes
  - _Requirements: 6.1, 6.5_

- [ ] 16. Create provider documentation
  - Create docs/index.md with provider overview, installation, and authentication setup
  - Document api_token, api_url, and api_version configuration options
  - Document HUBSPOT_API_TOKEN environment variable
  - Document supported API version
  - Create examples/provider/provider.tf with configuration examples
  - _Requirements: 1.1, 1.2, 1.4, 9.3, 9.5_

- [ ] 17. Create resource documentation
  - Create docs/resources/contact.md with schema reference and examples
  - Create docs/resources/company.md with schema reference and examples
  - Create docs/resources/deal.md with schema reference and examples
  - Create docs/resources/property.md with schema reference and examples
  - Include import instructions for each resource
  - Create example .tf files in examples/resources/ directory
  - _Requirements: 2.1, 3.1, 4.1, 5.1, 8.1, 8.2, 8.3, 8.4_

- [ ] 18. Create data source documentation
  - Create docs/data-sources/contact.md with schema reference and examples
  - Create docs/data-sources/company.md with schema reference and examples
  - Create docs/data-sources/deal.md with schema reference and examples
  - Create docs/data-sources/property.md with schema reference and examples
  - Create example .tf files demonstrating data source usage
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 19. Wire provider with all resources and data sources
  - Register all resources in provider Resources method (contact, company, deal, property)
  - Register all data sources in provider DataSources method (contact, company, deal, property)
  - Verify provider metadata and schema
  - Test provider initialization with all components
  - _Requirements: All requirements - final integration_

- [ ]* 20. Create integration test suite
  - Set up test HubSpot account configuration
  - Create test helper functions for acceptance tests
  - Implement testAccPreCheck function for environment validation
  - Run full acceptance test suite against test account
  - Verify all resources and data sources work end-to-end
  - _Requirements: All requirements - validation_
