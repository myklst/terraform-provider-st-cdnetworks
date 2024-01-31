package cdnetworks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &domainDataSource{}
	_ datasource.DataSourceWithConfigure = &domainDataSource{}
)

func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

type domainDataSource struct {
	client *cdnetworksapi.Client
}

type domainDataSourceModel struct {
	ClientConfig *clientConfig `tfsdk:"client_config"`
	DomainName   types.String  `tfsdk:"domain_name"`
	DomainCname  types.String  `tfsdk:"domain_cname"`
	OriginConfig *originConfig `tfsdk:"origin_config"`
}

type originConfig struct {
	OriginIps types.List `tfsdk:"origin_ips"`
}

type clientConfig struct {
	Username types.String `tfsdk:"username"`
	ApiKey   types.String `tfsdk:"api_key"`
}

func (d *domainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *domainDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a CDNetworks " +
			"CDN instance or a CDNetworks Flood Shield instance.",
		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				Description: "Domain name of domain.",
				Required:    true,
			},
			"domain_cname": schema.StringAttribute{
				Description: "Domain CNAME of domain.",
				Computed:    true,
			},
			"origin_config": schema.ObjectAttribute{
				Description: "Origin configuration of domain.",
				Computed:    true,
				AttributeTypes: map[string]attr.Type{
					"origin_ips": types.ListType{
						ElemType: types.StringType,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"client_config": schema.SingleNestedBlock{
				Description: "Config to override default client created in Provider. " +
					"This block will not be recorded in state file.",
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						Description: "The username of CDNetworks account. Default " +
							"to use username configured in the provider.",
						Optional: true,
					},
					"api_key": schema.StringAttribute{
						Description: "The api key of CDNetworks account. Default " +
							"to use api key configured in the provider.",
						Optional: true,
					},
				},
			},
		},
	}
}

func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan, state *domainDataSourceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ClientConfig == nil {
		plan.ClientConfig = &clientConfig{}
	}

	initClient := false
	username := plan.ClientConfig.Username.ValueString()
	apiKey := plan.ClientConfig.ApiKey.ValueString()

	if username != "" || apiKey != "" {
		initClient = true
	}

	if initClient {
		if username == "" {
			username = d.client.Username
		}
		if apiKey == "" {
			apiKey = d.client.ApiKey
		}
		var err error
		if d.client, err = cdnetworksapi.NewClient(username, apiKey); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Reinitialize CDNetworks API Client",
				"This is an error in provider, please contact the provider developers.\n\n"+
					"Error: "+err.Error(),
			)
			return
		}
	}

	domainName := plan.DomainName.ValueString()

	if domainName == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("domain_name"),
			"Missing domain name",
			"Domain name must not be empty",
		)
		return
	}

	var domainResp cdnetworksapi.QueryDomainResponse
	var err error
	queryDomainFunc := func() error {
		domainResp, err = d.client.QueryDomain(domainName)
		if err != nil {
			if _t, ok := err.(*cdnetworksapi.ErrorResponse); ok {
				// Ignore domain not found error and escape the backoff retry.
				if _t.ResponseCode == "NoSuchDomain" {
					return backoff.Permanent(fmt.Errorf("NoSuchDomain"))
				}
				if isAbleToRetry(_t.ResponseCode) {
					return err
				} else {
					return backoff.Permanent(err)
				}
			} else {
				return err
			}
		}
		return nil
	}

	state = &domainDataSourceModel{
		OriginConfig: &originConfig{
			OriginIps: types.ListNull(types.StringType),
		},
	}
	// Retry backoff
	reconnectBackoff := backoff.NewExponentialBackOff()
	reconnectBackoff.MaxElapsedTime = 10 * time.Minute
	if err := backoff.Retry(queryDomainFunc, reconnectBackoff); err != nil {
		if err.Error() == "NoSuchDomain" {
			state.DomainName = types.StringNull()
			state.DomainCname = types.StringNull()
			state.OriginConfig = &originConfig{
				OriginIps: types.ListNull(types.StringType),
			}
		} else {
			resp.Diagnostics.AddError(
				"[API ERROR] Unable to query domains",
				err.Error(),
			)
			return
		}
	} else {
		state.DomainName = types.StringValue(*domainResp.DomainName)
		state.DomainCname = types.StringValue(*domainResp.Cname)

		originIpsRawList := strings.Split(*domainResp.OriginConfig.OriginIps, ";")
		originIpsList, diags := types.ListValueFrom(ctx, types.StringType, originIpsRawList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.OriginConfig = &originConfig{
			OriginIps: originIpsList,
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
