package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-hubspot/internal/client"
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
	APIToken   types.String `tfsdk:"api_token"`
	APIURL     types.String `tfsdk:"api_url"`
	APIVersion types.String `tfsdk:"api_version"`
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
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "HubSpot API token for authentication. Can also be set via HUBSPOT_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_url": schema.StringAttribute{
				Description: "HubSpot API base URL. Defaults to https://api.hubapi.com",
				Optional:    true,
			},
			"api_version": schema.StringAttribute{
				Description: "HubSpot API version to use. Defaults to v3",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a HubSpot API client for data sources and resources.
func (p *HubSpotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config HubSpotProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get API token from config or environment variable
	apiToken := config.APIToken.ValueString()
	if apiToken == "" {
		apiToken = os.Getenv("HUBSPOT_API_TOKEN")
	}

	// Validate API token is provided
	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider requires an API token for authentication. "+
				"Set the api_token attribute in the provider configuration or "+
				"set the HUBSPOT_API_TOKEN environment variable.",
		)
		return
	}

	// Validate API token format (basic validation - should not be empty or just whitespace)
	if strings.TrimSpace(apiToken) == "" {
		resp.Diagnostics.AddError(
			"Invalid API Token",
			"The API token cannot be empty or contain only whitespace.",
		)
		return
	}

	// Validate API token format (HubSpot tokens typically start with "pat-" for private app tokens)
	// This is a basic format check - actual validation happens when making API calls
	if len(apiToken) < 10 {
		resp.Diagnostics.AddError(
			"Invalid API Token Format",
			"The API token appears to be too short. HubSpot API tokens are typically longer than 10 characters.",
		)
		return
	}

	// Get API URL with default
	apiURL := config.APIURL.ValueString()
	if apiURL == "" {
		apiURL = "https://api.hubapi.com"
	}

	// Get API version with default
	apiVersion := config.APIVersion.ValueString()
	if apiVersion == "" {
		apiVersion = "v3"
	}

	// Create HubSpot client
	hubspotClient := client.NewClient(client.Config{
		APIToken:   apiToken,
		BaseURL:    apiURL,
		APIVersion: apiVersion,
	})

	// Make the client available to resources and data sources
	resp.DataSourceData = hubspotClient
	resp.ResourceData = hubspotClient
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
