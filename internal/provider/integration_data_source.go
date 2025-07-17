package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected integrations.
var (
	_ datasource.DataSource              = &integrationDataSource{}
	_ datasource.DataSourceWithConfigure = &integrationDataSource{}
)

type integrationDataSource struct {
	client *retryablehttp.Client
}

type nangoIntegrationResponse struct {
	Data []nangoIntegrationModel `json:"data"`
}

type nanogoIntegrationResponse2 struct {
	Data nangoIntegrationModel `json:"data"`
}

type nangoIntegrationModel struct {
	UniqueKey     string `json:"unique_key"`
	DisplayName   string `json:"display_name"`
	NangoProvider string `json:"provider"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Logo          string `json:"logo"`
}

type integrationDataSourceModel struct {
	Integrations []integrationModel `tfsdk:"integrations"`
}
type integrationModel struct {
	UniqueKey     types.String                `tfsdk:"unique_key"`
	DisplayName   types.String                `tfsdk:"display_name"`
	NangoProvider types.String                `tfsdk:"nango_provider"`
	CreatedAt     types.String                `tfsdk:"created_at"`
	UpdatedAt     types.String                `tfsdk:"updated_at"`
	Logo          types.String                `tfsdk:"logo"`
	Credentials   *integrationCredentialModel `tfsdk:"credentials"`
}

type integrationCredentialModel struct {
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Type         types.String `tfsdk:"type"`
	Scopes       types.List   `tfsdk:"scopes"`
}

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewIntegrationDataSource() datasource.DataSource {
	return &integrationDataSource{}
}

// Metadata returns the data source type name.
func (d *integrationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

// Schema defines the schema for the data source.
func (d *integrationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"integrations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"unique_key": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The integration ID that you created in Nango.",
						},
						"display_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The provider display name.",
						},
						"nango_provider": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The nango_provider",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "When it was created",
						},
						"updated_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Last time it was updated",
						},
						"logo": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Absolute path to the logo of this integration (svg)",
						},
						"credentials": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "The credentials for this integration",
							Attributes: map[string]schema.Attribute{
								"client_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The client ID",
								},
								"client_secret": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The client secret",
								},
								"type": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The type of credential",
								},
								"scopes": schema.ListAttribute{
									Computed:            true,
									MarkdownDescription: "The scopes for this credential",
									ElementType:         types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *integrationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state integrationDataSourceModel

	integrationsResponse, err := d.client.Get("https://api.nango.dev/integrations")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Integrations",
			err.Error(),
		)
		return
	}

	//unmarshal response body to integrationDataSourceModel
	var integrations nangoIntegrationResponse
	err = json.NewDecoder(integrationsResponse.Body).Decode(&integrations)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode JSON",
			err.Error(),
		)
		return
	}

	// Set state
	for _, integration := range integrations.Data {
		integ := integrationModel{
			UniqueKey:     types.StringValue(integration.UniqueKey),
			DisplayName:   types.StringValue(integration.DisplayName),
			NangoProvider: types.StringValue(integration.NangoProvider),
			CreatedAt:     types.StringValue(integration.CreatedAt),
			UpdatedAt:     types.StringValue(integration.UpdatedAt),
			Logo:          types.StringValue(integration.Logo),
			Credentials:   nil, // Set to nil when credentials are not available
		}
		state.Integrations = append(state.Integrations, integ)
	}
	fmt.Println(state.Integrations)
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *integrationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*retryablehttp.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *retryablehttp.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}
