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
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	UniqueKey   *string `json:"unique_key,omitempty"`
	DisplayName string  `json:"display_name"`
	// NangoProvider is the Nango provider slug (e.g. "salesforce").
	NangoProvider *string `json:"provider,omitempty"`
	// Credentials is a pointer with omitempty so it can be left out entirely.
	//
	// Nango's public integration API accepts credentials only for the OAUTH1,
	// OAUTH2, TBA, APP and CUSTOM auth modes. BASIC and API_KEY integrations
	// hold no integration-level secret at all — the secrets are per-connection
	// — so for those the block must be absent. Sending an empty-string block
	// (which a non-pointer field without omitempty always produced) is rejected.
	Credentials *integrationCredentialsRequestModel `json:"credentials,omitempty"`
}

type integrationCredentialsRequestModel struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Type         string `json:"type"`
	Scopes       string `json:"scopes"` // Changed to string for API
}

// buildCredentialsRequest converts the planned credentials block into its
// request form, returning nil when the block was omitted.
//
// integrationModel.Credentials is a pointer, so it is nil whenever the
// practitioner leaves `credentials` out. Both Create and Update previously
// dereferenced it unconditionally (plan.Credentials.Scopes.ElementsAs), which
// panics for a credential-less integration.
func buildCredentialsRequest(ctx context.Context, credentials *integrationCredentialModel) *integrationCredentialsRequestModel {
	if credentials == nil {
		return nil
	}

	var scopes []string
	credentials.Scopes.ElementsAs(ctx, &scopes, false)

	return &integrationCredentialsRequestModel{
		ClientId:     credentials.ClientId.ValueString(),
		ClientSecret: credentials.ClientSecret.ValueString(),
		Type:         credentials.Type.ValueString(),
		Scopes:       strings.Join(scopes, ","), // Now a comma-delimited string
	}
}

// integrationResource is the resource implementation.
type integrationResource struct {
	client *nangoClient
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
				// Optional, not Required: BASIC and API_KEY integrations hold
				// no integration-level credentials, and Nango rejects an empty
				// credentials block. Widening Required -> Optional is additive,
				// so every existing OAuth2 configuration is unaffected and no
				// state migration is needed.
				Optional:            true,
				MarkdownDescription: "The credentials for this integration. Omit entirely for auth modes that carry no integration-level secret (e.g. BASIC, API_KEY), where credentials are supplied per-connection.",
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The client ID",
					},
					"client_secret": schema.StringAttribute{
						Required:            true,
						Sensitive:           true,
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

	// Populate the request model with data from the plan. Credentials are nil
	// when the block is omitted, which is the only valid form for auth modes
	// that carry no integration-level secret.
	request := integrationRequestModel{
		UniqueKey:     plan.UniqueKey.ValueStringPointer(),
		DisplayName:   plan.DisplayName.ValueString(),
		NangoProvider: plan.NangoProvider.ValueStringPointer(),
		Credentials:   buildCredentialsRequest(ctx, plan.Credentials),
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
	_, err = r.client.client.Post(r.client.baseURL+"/integrations", "application/json", requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Integration",
			err.Error(),
		)
		return
	}

	getResponse, gErr := r.client.client.Get(r.client.baseURL + "/integrations/" + plan.UniqueKey.ValueString() + "?include=webhook&include=credentials")
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

	// Get refreshed integration value from Nango, including credentials/scopes
	integrationResponse, err := r.client.client.Get(r.client.baseURL + "/integrations/" + state.UniqueKey.ValueString() + "?include=credentials")

  if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Nango Integration",
			"Could not read Nango integration "+state.UniqueKey.ValueString()+": "+err.Error(),
		)
		return
	}

	var integrationResp nanogoIntegrationResponse2
	err = json.NewDecoder(integrationResponse.Body).Decode(&integrationResp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode JSON",
			err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state from the API
	if integrationResp.Data.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(integrationResp.Data.UpdatedAt)
	}

	// Refresh client_id and client_secret from the API so Terraform can detect
	// credential drift. Without this, state forever echoes the last-applied
	// values and an out-of-band credential change (or a bad var about to be
	// applied over a hand-corrected one) never shows in a plan — updates send
	// the full credentials object, so the rewrite happens silently (CON-6127).
	// The non-empty guards keep an API contract change (field dropped or
	// masked to empty) from wiping state; a masked-to-placeholder change would
	// surface as a loud perpetual diff rather than a silent clobber.
	if integrationResp.Data.Credentials != nil && state.Credentials != nil {
		if integrationResp.Data.Credentials.ClientId != "" {
			state.Credentials.ClientId = types.StringValue(integrationResp.Data.Credentials.ClientId)
		}
		if integrationResp.Data.Credentials.ClientSecret != "" {
			state.Credentials.ClientSecret = types.StringValue(integrationResp.Data.Credentials.ClientSecret)
		}
	}

	// Parse scopes from API response back into types.List so Terraform can detect drift
	if integrationResp.Data.Credentials != nil && integrationResp.Data.Credentials.Scopes != "" {
		scopeStrings := strings.Split(integrationResp.Data.Credentials.Scopes, ",")
		scopeValues := make([]attr.Value, len(scopeStrings))
		for i, s := range scopeStrings {
			scopeValues[i] = types.StringValue(strings.TrimSpace(s))
		}
		scopesList, scopeDiags := types.ListValue(types.StringType, scopeValues)
		resp.Diagnostics.Append(scopeDiags...)
		if !resp.Diagnostics.HasError() && state.Credentials != nil {
			state.Credentials.Scopes = scopesList
		}
	}

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

	// Populate the request model with data from the plan (excluding unique_key and provider for updates)
	// unique_key must NOT be in the body — Nango interprets it as a rename attempt
	request := integrationRequestModel{
		DisplayName: plan.DisplayName.ValueString(),
		Credentials: buildCredentialsRequest(ctx, plan.Credentials),
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

	// Create a PATCH request to update the integration
	req2, err := retryablehttp.NewRequest("PATCH", r.client.baseURL+"/integrations/"+plan.UniqueKey.ValueString(), requestBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Request",
			err.Error(),
		)
		return
	}
	req2.Header.Set("Content-Type", "application/json")

	response, err := r.client.client.Do(req2)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Update Integration",
			err.Error(),
		)
		return
	}

	if response.StatusCode != 200 {
		bodyBytes, _ := json.Marshal(response.Body)
		resp.Diagnostics.AddError(
			"Nango API Error",
			fmt.Sprintf("PATCH /integrations/%s returned HTTP %d: %s", plan.UniqueKey.ValueString(), response.StatusCode, string(bodyBytes)),
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

	client, ok := req.ProviderData.(*nangoClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *nangoClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
