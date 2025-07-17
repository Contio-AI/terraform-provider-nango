// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &integrationResource{}
	_ resource.ResourceWithConfigure = &integrationResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewIntegrationResource() resource.Resource {
	return &integrationResource{}
}

type integrationRequestModel struct {
	UniqueKey     *string                            `json:"unique_key,omitempty"`
	DisplayName   string                             `json:"display_name"`
	NangoProvider *string                            `json:"provider,omitempty"`
	Credentials   integrationCredentialsRequestModel `json:"credentials"`
}

type integrationCredentialsRequestModel struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Type         string `json:"type"`
	Scopes       string `json:"scopes"` // Changed to string for API
}

// integrationResource is the resource implementation.
type integrationResource struct {
	client *retryablehttp.Client
}

// Metadata returns the resource type name.
func (r *integrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

// Schema defines the schema for the resource.
func (r *integrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"unique_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The integration ID that you created in Nango.",
			},
			"display_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The provider display name.",
			},
			"nango_provider": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The nango_provider",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last time it was updated",
			},
			"credentials": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The credentials for this integration",
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The client ID",
					},
					"client_secret": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The client secret",
					},
					"type": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The type of credential",
					},
					"scopes": schema.ListAttribute{
						Optional:            true,
						MarkdownDescription: "The scopes for this credential",
						ElementType:         types.StringType,
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *integrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert scopes from types.List to []string, then to comma-delimited string
	var scopes []string
	plan.Credentials.Scopes.ElementsAs(ctx, &scopes, false)
	scopesString := strings.Join(scopes, ",")

	// Populate the request model with data from the plan
	request := integrationRequestModel{
		UniqueKey:     plan.UniqueKey.ValueStringPointer(),
		DisplayName:   plan.DisplayName.ValueString(),
		NangoProvider: plan.NangoProvider.ValueStringPointer(),
		Credentials: integrationCredentialsRequestModel{
			ClientId:     plan.Credentials.ClientId.ValueString(),
			ClientSecret: plan.Credentials.ClientSecret.ValueString(),
			Type:         plan.Credentials.Type.ValueString(),
			Scopes:       scopesString, // Now a comma-delimited string
		},
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Marshal JSON",
			err.Error(),
		)
		return
	}

	// Create a new request with the JSON body
	_, err = r.client.Post("https://api.nango.dev/integrations", "application/json", requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Integration",
			err.Error(),
		)
		return
	}

	getResponse, gErr := r.client.Get("https://api.nango.dev/integrations/" + plan.UniqueKey.ValueString() + "?include=webhook&include=credentials")
	if gErr != nil {
		resp.Diagnostics.AddError(
			"Unable to Get Integration",
			gErr.Error(),
		)
		return
	}

	//unmarshal response body to integrationModel
	var integration nanogoIntegrationResponse2
	err = json.NewDecoder(getResponse.Body).Decode(&integration)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode JSON",
			err.Error(),
		)
		return
	}

	plan.UniqueKey = types.StringValue(integration.Data.UniqueKey)
	plan.UpdatedAt = types.StringValue(integration.Data.UpdatedAt)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state integrationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from HashiCups
	integrationResponse, err := r.client.Get("https://api.nango.dev/integrations/" + state.UniqueKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HashiCups Order",
			"Could not read HashiCups order ID "+state.UniqueKey.ValueString()+": "+err.Error(),
		)
		return
	}

	var integrations nangoIntegrationModel
	err = json.NewDecoder(integrationResponse.Body).Decode(&integrations)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode JSON",
			err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	// state.DisplayName = types.StringValue(integrations.DisplayName)
	// state.NangoProvider = types.StringValue(integrations.NangoProvider)
	// state.CreatedAt = types.StringValue(integrations.CreatedAt)
	// state.UpdatedAt = types.StringValue(integrations.UpdatedAt)
	// state.Logo = types.StringValue(integrations.Logo)
	// state.WebhookUrl = types.StringValue(integrations.WebhookUrl)
	// state.Credentials = integrationCredentialModel{
	// 	ClientId:     types.StringValue(integrations.Credentials.ClientId),
	// 	ClientSecret: types.StringValue(integrations.Credentials.ClientSecret),
	// 	Type:         types.StringValue(integrations.Credentials.Type),
	// }

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *integrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan integrationModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert scopes from types.List to []string, then to comma-delimited string
	var scopes []string
	plan.Credentials.Scopes.ElementsAs(ctx, &scopes, false)
	scopesString := strings.Join(scopes, ",")

	// Populate the request model with data from the plan (excluding provider for updates)
	request := integrationRequestModel{
		UniqueKey:   plan.UniqueKey.ValueStringPointer(),
		DisplayName: plan.DisplayName.ValueString(),
		// NangoProvider omitted for PATCH requests
		Credentials: integrationCredentialsRequestModel{
			ClientId:     plan.Credentials.ClientId.ValueString(),
			ClientSecret: plan.Credentials.ClientSecret.ValueString(),
			Type:         plan.Credentials.Type.ValueString(),
			Scopes:       scopesString, // Now a comma-delimited string
		},
	}

	// Convert request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Marshal JSON",
			err.Error(),
		)
		return
	}

	// Debug: Log the update request body
	fmt.Printf("Update Request Body: %s\n", string(requestBody))

	// Create a PATCH request to update the integration
	req2, err := retryablehttp.NewRequest("PATCH", "https://api.nango.dev/integrations/"+plan.UniqueKey.ValueString(), requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Request",
			err.Error(),
		)
		return
	}
	req2.Header.Set("Content-Type", "application/json")

	response, err := r.client.Do(req2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Integration",
			err.Error(),
		)
		return
	}

	// Unmarshal response body to integrationModel
	var integration nangoIntegrationModel
	err = json.NewDecoder(response.Body).Decode(&integration)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode JSON",
			err.Error(),
		)
		return
	}

	// Debug: Log the response
	fmt.Printf("Update Response: %+v\n", integration)

	// Update the plan with response data if available, otherwise use plan values
	if integration.DisplayName != "" {
		plan.DisplayName = types.StringValue(integration.DisplayName)
	}
	if integration.UpdatedAt != "" {
		plan.UpdatedAt = types.StringValue(integration.UpdatedAt)
	} else {
		plan.UpdatedAt = types.StringValue(time.Now().Format(time.RFC3339))
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *integrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// Configure adds the provider configured client to the resource.
func (r *integrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*retryablehttp.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// ImportState imports the resource into Terraform state.
func (r *integrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID should be the unique_key of the integration
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("unique_key"), req.ID)...)
}
