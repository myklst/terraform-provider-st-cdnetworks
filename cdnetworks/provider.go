package cdnetworks

import (
	"context"
	"os"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_     provider.Provider = &cdnetworksProvider{}
	mutex sync.Mutex
)

// New is a helper function to simplify provider server
func New() provider.Provider {
	return &cdnetworksProvider{}
}

type cdnetworksProvider struct{}

type cdnetworksProviderModel struct {
	Username types.String `tfsdk:"username"`
	ApiKey   types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *cdnetworksProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "st-cdnetworks"
}

// Schema defines the provider-level schema for configuration data.
func (p *cdnetworksProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The CDNetworks provider is used to interact with the many resources supported by CDNetworks. " +
			"The provider needs to be configured with the proper credentials before it can be used.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "URI for CDNetworks API. May also be provided via CDNETWORKS_USERNAME environment variable",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for CDNetworks API. May also be provided via CDNETWORKS_API_KEY environment variable",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a CDNetworks API client for data sources and resources.
func (p *cdnetworksProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config cdnetworksProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown CDNetworks Username",
			"The provider cannot create the CDNetworks API client as there is an "+
				"unknown configuration value for the CDNetworks API username. Set "+
				"the value statically in the configuration, or use the CDNETWORKS_USERNAME "+
				"environment variable.",
		)
	}

	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown CDNetworks API key",
			"The provider cannot create the CDNetworks API client as there is an "+
				"unknown configuration value for the CDNetworks API key. Set the "+
				"value statically in the configuration, or use the CDNETWORKS_API_KEY "+
				"environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	var username, apiKey string
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	} else {
		username = os.Getenv("CDNETWORKS_USERNAME")
	}

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	} else {
		apiKey = os.Getenv("CDNETWORKS_API_KEY")
	}

	// If any of the expected configuration are missing, return
	// errors with provider-specific guidance.
	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing CDNetworks API Username",
			"The provider cannot create the CDNetworks API client as there is a "+
				"missing or empty value for the CDNetworks API username. Set the "+
				"username value in the configuration or use the CDNETWORKS_USERNAME "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing CDNetworks API key",
			"The provider cannot create the CDNetworks API client as there is a "+
				"missing or empty value for the CDNetworks API key. Set the API "+
				"key value in the configuration or use the CDNETWORKS_API_KEY environment "+
				"variable. If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := cdnetworksapi.NewClient(username, apiKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create CDNetworks API Client",
			"An unexpected error occurred when creating the CDNetworks API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"CDNetworks Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *cdnetworksProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
		NewCertDataSource,
	}
}

func (p *cdnetworksProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSslCertificateResource,
		NewContentAccelerationDomainResource,
		NewFloodShieldDomainResource,
		NewDomainSslAssociationResource,
		NewHttpHeaderConfigResource,
		NewHttp2SettingsConfigResource,
		NewAntiHotlinkingConfigResource,
		NewBackToOriginProtocolRewriteConfigResource,
		NewCacheTimeResource,
		NewQueryStringUrlConfigResource,
		NewHttpCodeCacheConfigResource,
		NewIgnoreProtocolResource,
		NewIpv6Resource,
		NewOriginRulesRewriteConfigResource,
		NewUrlSignResource,
	}
}
