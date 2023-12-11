package cdnetworks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

var http2SettingAttributeTypes = map[string]attr.Type{
	"enable_http2":            types.BoolType,
	"back_to_origin_protocol": types.StringType,
}

type http2SettingsConfigModel struct {
	DomainId      types.String `tfsdk:"domain_id"`
	Http2Settings types.Object `tfsdk:"http2_settings"`
}

type http2SettingsConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &http2SettingsConfigResource{}
	_ resource.ResourceWithConfigure   = &http2SettingsConfigResource{}
	_ resource.ResourceWithImportState = &http2SettingsConfigResource{}
)

func NewHttp2SettingsConfigResource() resource.Resource {
	return &http2SettingsConfigResource{}
}

func (r *http2SettingsConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http2_settings_config"
}

func (r *http2SettingsConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "HTTP2.0 controls whether CDN uses http2.0 protocol to interact with the client, and Back-to-Origin protocol version controls whether to use http2.0 protocol Back-to-Origin. Some products do not support to configure http2.0 Back-to-Origin.",
		Attributes: map[string]schema.Attribute{
			"domain_id": &schema.StringAttribute{
				Description: "Domain ID",
				Required:    true,
			},
			"http2_settings": &schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"enable_http2": &schema.BoolAttribute{
						Description: "Enable http2.0. The optional values are true and false. If it is empty, the default value is false. True means http2.0 is on; false means http2.0 is off.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"back_to_origin_protocol": &schema.StringAttribute{
						Description: `Back-to-origin protocol, the optional value is
                                    http1.1: Use the HTTP1.1 protocol version to back to source. if not filled, use it as default.
                                    follow-request: Same as client request protocol.
                                    http2.0: Use the HTTP2.0 protocol. version to back to source.`,
						Optional: true,
						Computed: true,
						Default:  stringdefault.StaticString("http1.1"),
						Validators: []validator.String{
							stringvalidator.OneOf("http1.1", "follow-request", "http2.0"),
						},
					},
				},
				Required: true,
			},
		},
	}
}

func (r *http2SettingsConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *http2SettingsConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *http2SettingsConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to update http2_setting_config", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *http2SettingsConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *http2SettingsConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryHttp2SettingsConfigResponse, err := r.client.QueryHttp2SettingsConfig(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to query http2_setting_config", err.Error())
		return
	}
	if queryHttp2SettingsConfigResponse.Http2Setting == nil {
		model.Http2Settings = types.ObjectNull(http2SettingAttributeTypes)
	} else {
		model.Http2Settings = types.ObjectValueMust(http2SettingAttributeTypes, map[string]attr.Value{
			"enable_http2":            types.BoolPointerValue(queryHttp2SettingsConfigResponse.Http2Setting.EnableHttp2),
			"back_to_origin_protocol": types.StringPointerValue(queryHttp2SettingsConfigResponse.Http2Setting.BackToOriginProtocol),
		})
	}
	resp.State.Set(ctx, model)
}

func (r *http2SettingsConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *http2SettingsConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to update http2_settings_config", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *http2SettingsConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *http2SettingsConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Http2Settings = types.ObjectNull(http2SettingAttributeTypes)
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete http2_setting_config", err.Error())
	}
}

func (r *http2SettingsConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *http2SettingsConfigResource) updateConfig(model *http2SettingsConfigModel) error {
	setting := &cdnetworksapi.Http2Setting{}
	for k, v := range model.Http2Settings.Attributes() {
		if k == "enable_http2" && !v.IsNull() {
			setting.EnableHttp2 = v.(types.Bool).ValueBoolPointer()
		} else if k == "back_to_origin_protocol" && !v.IsNull() {
			setting.BackToOriginProtocol = v.(types.String).ValueStringPointer()
		}
	}
	updateHttp2SettingsConfigRequest := cdnetworksapi.UpdateHttp2SettingsConfigRequest{
		Http2Setting: setting,
	}
	_, err := r.client.UpdateHttp2SettingsConfig(model.DomainId.ValueString(), updateHttp2SettingsConfigRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}
