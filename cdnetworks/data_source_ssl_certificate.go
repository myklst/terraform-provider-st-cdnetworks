package cdnetworks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

var (
	_ datasource.DataSource              = &certDataSource{}
	_ datasource.DataSourceWithConfigure = &certDataSource{}
)

type certificate struct {
	Id   types.String `tfsdk:"certificate_id"`
	Name types.String `tfsdk:"name"`
}

type certDataSourceModel struct {
	CertNameList types.List     `tfsdk:"cert_name_list"`
	CertList     []*certificate `tfsdk:"cert_list"`
}

type certDataSource struct {
	client *cdnetworksapi.Client
}

func NewCertDataSource() datasource.DataSource {
	return &certDataSource{}
}

func (d *certDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl_certificate"
}

func (d *certDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source provides certificates configured in cdnetworks.",
		Attributes: map[string]schema.Attribute{
			"cert_name_list": schema.ListAttribute{
				Description: "List of certificate name.If cert_name_list is null,retrieve all certificates.If cert_name_list is not null (includes empty), retrive certificates with specific name.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"cert_list": schema.ListNestedAttribute{
				Description: "List of certificate",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of certificate",
							Computed:    true,
						},
						"certificate_id": schema.StringAttribute{
							Description: "Certificate ID",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *certDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (d *certDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model, state certDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.CertNameList = model.CertNameList

	queryCertificateListResponse, err := d.client.QueryCertificateList()
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Certificate List", err.Error())
		return
	}

	state.CertList = make([]*certificate, 0)
	if state.CertNameList.IsNull() {
		for _, ssl := range queryCertificateListResponse.SslCertificates {
			state.CertList = append(state.CertList,
				&certificate{
					Id:   types.StringPointerValue(ssl.CertificateId),
					Name: types.StringPointerValue(ssl.Name),
				},
			)
		}
	} else {
		for _, name := range state.CertNameList.Elements() {
			var c *certificate
			for _, ssl := range queryCertificateListResponse.SslCertificates {
				if name.(types.String).ValueString() == *ssl.Name {
					c = &certificate{
						Id:   types.StringPointerValue(ssl.CertificateId),
						Name: types.StringPointerValue(ssl.Name),
					}
					break
				}
			}
			state.CertList = append(state.CertList, c)
		}
	}

	resp.State.Set(ctx, &state)
}
