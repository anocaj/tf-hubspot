package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure HubSpotProvider satisfies various provider interfaces.
var _ provider.Provider = &HubSpotProvider{}

// HubSpotProvider defines the provider implementation.
type HubSpotProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// HubSpotProviderModel describes the provider data model.
type HubSpotProviderModel struct {
	// Provider configuration fields will be added in task 3
}

// New creates a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HubSpotProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *HubSpotProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hubspot"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *HubSpotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing HubSpot resources.",
		Attributes:  map[string]schema.Attribute{
			// Provider configuration schema will be added in task 3
		},
	}
}

// Configure prepares a HubSpot API client for data sources and resources.
func (p *HubSpotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Provider configuration will be implemented in task 3
}

// Resources defines the resources implemented in the provider.
func (p *HubSpotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Resources will be registered in task 19
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *HubSpotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources will be registered in task 19
	}
}
