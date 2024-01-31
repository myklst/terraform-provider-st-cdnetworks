package cdnetworks

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type backToOriginProtocolRewriteConfigModel struct {
	DomainId types.String `tfsdk:"domain_id"`
	Protocol types.String `tfsdk:"protocol"`
	Port     types.String `tfsdk:"port"`
}

type backToOriginProtocolRewriteConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &backToOriginProtocolRewriteConfigResource{}
	_ resource.ResourceWithConfigure   = &backToOriginProtocolRewriteConfigResource{}
	_ resource.ResourceWithModifyPlan  = &backToOriginProtocolRewriteConfigResource{}
	_ resource.ResourceWithImportState = &backToOriginProtocolRewriteConfigResource{}
)

func NewBackToOriginProtocolRewriteConfigResource() resource.Resource {
	return &backToOriginProtocolRewriteConfigResource{}
}

func (r *backToOriginProtocolRewriteConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_back_to_origin_protocol_rewrite_config"
}

func (r *backToOriginProtocolRewriteConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Set the CDN back-to-origin protocol. By default, it follows the protocol requested by the client. If choose HTTPS-->HTTP, port 80 will be used by default; if choose HTTP-->HTTPS, port 443 will be used by default..",
		Attributes: map[string]schema.Attribute{
			"domain_id": &schema.StringAttribute{
				Description: "Domain ID",
				Required:    true,
			},
			"protocol": &schema.StringAttribute{
				Description: "The specified protocol is either http or https.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("https", "http"),
				},
			},
			"port": &schema.StringAttribute{
				Description: "If the protocol is http, the default is 80. If the protocol is https, the default is 443.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *backToOriginProtocolRewriteConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *backToOriginProtocolRewriteConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *backToOriginProtocolRewriteConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to set back_to_origin_protocol_rewrite_config", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *backToOriginProtocolRewriteConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *backToOriginProtocolRewriteConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryBackToOriginRewriteConfigResponse, err := r.client.QueryBackToOriginRewriteConfig(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to query back_to_origin_protocol_rewrite_config", err.Error())
		return
	}
	if queryBackToOriginRewriteConfigResponse.BackToOriginRewriteRule.Protocol == nil {
		resp.State.RemoveResource(ctx)
	} else {
		model.Protocol = types.StringPointerValue(queryBackToOriginRewriteConfigResponse.BackToOriginRewriteRule.Protocol)
		if queryBackToOriginRewriteConfigResponse.BackToOriginRewriteRule.Port == nil {
			if model.Protocol.ValueString() == "http" {
				model.Port = types.StringValue("80")
			} else if model.Protocol.ValueString() == "https" {
				model.Port = types.StringValue("443")
			}
		} else {
			model.Port = types.StringPointerValue(queryBackToOriginRewriteConfigResponse.BackToOriginRewriteRule.Port)
		}
		resp.State.Set(ctx, model)
	}
}

func (r *backToOriginProtocolRewriteConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *backToOriginProtocolRewriteConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to set back_to_origin_protocol_rewrite_config", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *backToOriginProtocolRewriteConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *backToOriginProtocolRewriteConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Protocol = types.StringNull()
	model.Port = types.StringNull()
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete back_to_origin_protocol_rewrite_config", err.Error())
	}
}

func (r *backToOriginProtocolRewriteConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *backToOriginProtocolRewriteConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *backToOriginProtocolRewriteConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan == nil {
		return
	}
	if plan.Port.IsUnknown() {
		if plan.Protocol.ValueString() == "http" {
			plan.Port = types.StringValue("80")
		} else if plan.Protocol.ValueString() == "https" {
			plan.Port = types.StringValue("443")
		}
	}
	resp.Plan.Set(ctx, plan)
}

func (r *backToOriginProtocolRewriteConfigResource) updateConfig(model *backToOriginProtocolRewriteConfigModel) error {
	if model == nil {
		return errors.New("model is nil")
	}
	updateBackToOriginRewriteConfigRequest := cdnetworksapi.UpdateBackToOriginRewriteConfigRequest{
		BackToOriginRewriteRule: cdnetworksapi.BackToOriginRewriteRule{
			Protocol: model.Protocol.ValueStringPointer(),
			Port:     model.Port.ValueStringPointer(),
		},
	}
	_, err := r.client.UpdateBackToOriginRewriteConfig(model.DomainId.ValueString(), updateBackToOriginRewriteConfigRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}
