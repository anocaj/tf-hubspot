package resources

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-hubspot/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ContactResource{}
var _ resource.ResourceWithImportState = &ContactResource{}

// NewContactResource creates a new contact resource.
func NewContactResource() resource.Resource {
	return &ContactResource{}
}

// ContactResource defines the resource implementation.
type ContactResource struct {
	client *client.Client
}

// ContactResourceModel describes the resource data model.
type ContactResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Firstname  types.String `tfsdk:"firstname"`
	Lastname   types.String `tfsdk:"lastname"`
	Properties types.Map    `tfsdk:"properties"`
}

// Metadata returns the resource type name.
func (r *ContactResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

// Schema defines the schema for the resource.
func (r *ContactResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a HubSpot contact.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the contact.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "The email address of the contact.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
						"must be a valid email address",
					),
				},
			},
			"firstname": schema.StringAttribute{
				Description: "The first name of the contact.",
				Optional:    true,
			},
			"lastname": schema.StringAttribute{
				Description: "The last name of the contact.",
				Optional:    true,
			},
			"properties": schema.MapAttribute{
				Description: "Additional custom properties for the contact.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ContactResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new contact resource.
func (r *ContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContactResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build properties map
	properties := make(map[string]interface{})
	
	// Add required and optional fields
	if !data.Email.IsNull() {
		properties["email"] = data.Email.ValueString()
	}
	if !data.Firstname.IsNull() {
		properties["firstname"] = data.Firstname.ValueString()
	}
	if !data.Lastname.IsNull() {
		properties["lastname"] = data.Lastname.ValueString()
	}

	// Add custom properties
	if !data.Properties.IsNull() {
		customProps := make(map[string]string)
		diags := data.Properties.ElementsAs(ctx, &customProps, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		for key, value := range customProps {
			properties[key] = value
		}
	}

	// Create contact via API
	contact, err := r.client.CreateContact(ctx, properties)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Contact",
			fmt.Sprintf("Could not create contact: %s", err.Error()),
		)
		return
	}

	// Set the ID
	data.ID = types.StringValue(contact.ID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read reads the contact resource.
func (r *ContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContactResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get contact from API
	contact, err := r.client.GetContact(ctx, data.ID.ValueString())
	if err != nil {
		// Check if this is a 404 error
		if hubspotErr, ok := err.(*client.HubSpotError); ok && hubspotErr.Status == "404" {
			// Contact no longer exists, remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		
		resp.Diagnostics.AddError(
			"Error Reading Contact",
			fmt.Sprintf("Could not read contact ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update model with API response
	if email, ok := contact.Properties["email"].(string); ok {
		data.Email = types.StringValue(email)
	}
	if firstname, ok := contact.Properties["firstname"].(string); ok {
		data.Firstname = types.StringValue(firstname)
	}
	if lastname, ok := contact.Properties["lastname"].(string); ok {
		data.Lastname = types.StringValue(lastname)
	}

	// Handle custom properties
	if !data.Properties.IsNull() {
		// Get the list of custom property keys from the current state
		customProps := make(map[string]string)
		diags := data.Properties.ElementsAs(ctx, &customProps, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Update custom properties from API response
		updatedProps := make(map[string]string)
		for key := range customProps {
			if val, ok := contact.Properties[key]; ok {
				if strVal, ok := val.(string); ok {
					updatedProps[key] = strVal
				}
			}
		}

		propsMap, diags := types.MapValueFrom(ctx, types.StringType, updatedProps)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Properties = propsMap
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the contact resource.
func (r *ContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContactResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build properties map
	properties := make(map[string]interface{})
	
	// Add required and optional fields
	if !data.Email.IsNull() {
		properties["email"] = data.Email.ValueString()
	}
	if !data.Firstname.IsNull() {
		properties["firstname"] = data.Firstname.ValueString()
	}
	if !data.Lastname.IsNull() {
		properties["lastname"] = data.Lastname.ValueString()
	}

	// Add custom properties
	if !data.Properties.IsNull() {
		customProps := make(map[string]string)
		diags := data.Properties.ElementsAs(ctx, &customProps, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		
		for key, value := range customProps {
			properties[key] = value
		}
	}

	// Update contact via API
	_, err := r.client.UpdateContact(ctx, data.ID.ValueString(), properties)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Contact",
			fmt.Sprintf("Could not update contact ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the contact resource.
func (r *ContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContactResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete contact via API
	err := r.client.DeleteContact(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Contact",
			fmt.Sprintf("Could not delete contact ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports an existing contact resource by ID.
func (r *ContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Use the ID provided in the import command
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
