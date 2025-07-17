package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &nangoProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &nangoProvider{
			version: version,
		}
	}
}

type nangoProviderMdoel struct {
	EnvironmentKey types.String `tfsdk:"environment_key"`
}

// nangoProvider is the provider implementation.
type nangoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *nangoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nango"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *nangoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_key": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// Configure prepares a Nango API client for data sources and resources.
func (p *nangoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config nangoProviderMdoel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.EnvironmentKey.IsUnknown() {
		resp.Diagnostics.AddError(
			"Unable to find environment key",
			"Expected environment key to be set, but it was not.",
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	environmentKey := os.Getenv("NANGO_ENVIRONMENT_KEY")

	if !config.EnvironmentKey.IsNull() {
		environmentKey = config.EnvironmentKey.ValueString()
	}

	if environmentKey == "" {
		resp.Diagnostics.AddError(
			"Unable to find environment key2",
			"Expected environment key to be set, but it was not.",
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3                          // Maximum retry attempts
	retryClient.RetryWaitMin = 1 * time.Second        // Minimum wait time between retries
	retryClient.RetryWaitMax = 5 * time.Second        // Maximum wait time between retries
	retryClient.HTTPClient.Timeout = 30 * time.Second // Set the timeout for the HTTP client
	retryClient.HTTPClient.Transport = &myTransport{authKey: environmentKey, next: retryClient.HTTPClient.Transport}

	resp.DataSourceData = retryClient
	resp.ResourceData = retryClient
}

// DataSources defines the data sources implemented in the provider.
func (p *nangoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIntegrationDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *nangoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIntegrationResource,
	}
}

type myTransport struct {
	authKey string
	next    http.RoundTripper
}

func (t *myTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Println("RoundTrip called")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.authKey))
	startTime := time.Now()

	log.Printf("Request: %s %s %s", req.Method, req.URL.String(), req.Header["Authoriztaion"])
	if req.Body != nil {
		reqBody, _ := io.ReadAll(req.Body)
		log.Printf("Request Body: %s", string(reqBody))
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	resp, err := t.next.RoundTrip(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	elapsedTime := time.Since(startTime)
	log.Printf("Response: %s %s - %d in %s", req.Method, req.URL.String(), resp.StatusCode, elapsedTime)

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Response Body: %s", string(respBody))
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	return resp, nil
	// return http.DefaultTransport.RoundTrip(req)
}
