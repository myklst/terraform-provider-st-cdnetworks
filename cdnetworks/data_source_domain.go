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
	DomainName   types.String  `tfsdk:"domain_name"`
	DomainCname  types.String  `tfsdk:"domain_cname"`
	OriginConfig *originConfig `tfsdk:"origin_config"`
}

type originConfig struct {
	OriginIps types.List `tfsdk:"origin_ips"`
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
				Description: "Domain Cname of domain.",
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
	}
}

func (d *domainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var plan, state domainDataSourceModel
	diags := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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
	createGtmInstance := func() error {
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

	// Retry backoff
	reconnectBackoff := backoff.NewExponentialBackOff()
	reconnectBackoff.MaxElapsedTime = 10 * time.Minute
	err = backoff.Retry(createGtmInstance, reconnectBackoff)
	if err != nil {
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
