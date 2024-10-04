package cdnetworks

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type originRulesRewriteModel struct {
	PathPattern           types.String `tfsdk:"path_pattern"`
	PathPatternHttp       types.String `tfsdk:"path_pattern_http"`
	ExceptPathPattern     types.String `tfsdk:"except_path_pattern"`
	ExceptPathPatternHttp types.String `tfsdk:"except_path_pattern_http"`
	IgnoreLetterCase      types.Bool   `tfsdk:"ignore_letter_case"`
	OriginInfo            types.String `tfsdk:"origin_info"`
	Priority              types.Int64  `tfsdk:"priority"`
	OriginHost            types.String `tfsdk:"origin_host"`
	BeforeRewriteUri      types.String `tfsdk:"before_rewrite_uri"`
	AfterRewriteUri       types.String `tfsdk:"after_rewrite_uri"`
}

type originRulesRewriteConfigModel struct {
	DomainId           types.String               `tfsdk:"domain_id"`
	OriginRulesRewrite []*originRulesRewriteModel `tfsdk:"origin_rules_rewrite"`
}

type originRulesRewriteConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &originRulesRewriteConfigResource{}
	_ resource.ResourceWithConfigure   = &originRulesRewriteConfigResource{}
	_ resource.ResourceWithImportState = &originRulesRewriteConfigResource{}
)

func NewOriginRulesRewriteConfigResource() resource.Resource {
	return &originRulesRewriteConfigResource{}
}

func (r *originRulesRewriteConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_origin_rules_rewrite_config"
}

func (r *originRulesRewriteConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource configures alternative(s) origins for specific URL paths.",
		Attributes: map[string]schema.Attribute{
			"domain_id": &schema.StringAttribute{
				Description: "Domain ID",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"origin_rules_rewrite": &schema.ListNestedBlock{
				Description: "Configures path rewrites, alternate origins and url rewrites.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path_pattern": &schema.StringAttribute{
							Description: "Url matching mode supports regular expression. To match all paths, input can be .*",
							Required:    true,
						},
						"path_pattern_http": &schema.StringAttribute{
							Description: strings.Join([]string{
								"Whether to match only paths with HTTP or HTTPS protocol only.",
								"Default is blank, matches all paths regardless of protocol"}, " "),
							Optional: true,
						},
						"except_path_pattern": &schema.StringAttribute{
							Description: "Ignore url rewrite rules if paths match these patterns.",
							Optional:    true,
						},
						"except_path_pattern_http": &schema.StringAttribute{
							Description: strings.Join([]string{
								"Whether to match only paths with HTTP or HTTPS protocol only.",
								"Default is blank, matches all paths regardless of protocol"}, " "),
							Optional: true,
						},
						"ignore_letter_case": &schema.BoolAttribute{
							Description: "Whether to match the letter casing",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
						"origin_info": &schema.StringAttribute{
							Description: "The information of the back to source. Can be IP or domain name.",
							Required:    true,
						},
						"priority": &schema.Int64Attribute{
							Description: "The priority of the execution order. The bigger the number, the higher the priority.",
							Optional:    true,
							Computed:    true,
							Default:     int64default.StaticInt64(10),
						},
						"origin_host": &schema.StringAttribute{
							Description: "The host/domain name to use when performing back to origin request.",
							Required:    true,
						},
						"before_rewrite_uri": &schema.StringAttribute{
							Description: "The original request URI. Supports regular expression.",
							Optional:    true,
						},
						"after_rewrite_uri": &schema.StringAttribute{
							Description: "The rewritten request URI. Supports regular expression.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *originRulesRewriteConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *originRulesRewriteConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *originRulesRewriteConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	originRulesRewrites := []*cdnetworksapi.OriginRulesRewrite{}
	for _, v := range model.OriginRulesRewrite {
		originRulesRewrites = append(originRulesRewrites, &cdnetworksapi.OriginRulesRewrite{
			PathPattern:           v.PathPattern.ValueStringPointer(),
			PathPatternHttp:       v.PathPatternHttp.ValueStringPointer(),
			ExceptPathPattern:     v.ExceptPathPattern.ValueStringPointer(),
			ExceptPathPatternHttp: v.ExceptPathPatternHttp.ValueStringPointer(),
			IgnoreLetterCase:      v.IgnoreLetterCase.ValueBoolPointer(),
			OriginInfo:            v.OriginInfo.ValueStringPointer(),
			Priority:              v.Priority.ValueInt64Pointer(),
			OriginHost:            v.OriginHost.ValueStringPointer(),
			BeforeRewriteUri:      v.BeforeRewriteUri.ValueStringPointer(),
			AfterRewriteUri:       v.AfterRewriteUri.ValueStringPointer(),
		})
	}

	_, err := r.client.UpdateOriginUriAndOriginHost(model.DomainId.ValueString(), cdnetworksapi.UpdateOriginUriAndOriginHostRequest{
		OriginRulesRewrites: originRulesRewrites,
	})
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to set origin_rules_rewrites", err.Error())
		return
	}
	resp.State.Set(ctx, model)
}

func (r *originRulesRewriteConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *originRulesRewriteConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateModel(model)
	if err != nil {
		resp.Diagnostics.AddError("[API Error]Fail to query origin_rules_rewrites", err.Error())
		return
	}

	resp.State.Set(ctx, model)
}

func (r *originRulesRewriteConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *originRulesRewriteConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API Error]Fail to update origin_rules_rewrites", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *originRulesRewriteConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *originRulesRewriteConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.OriginRulesRewrite = make([]*originRulesRewriteModel, 0)
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete origin_rules_rewrite", err.Error())
	}
}

func (r *originRulesRewriteConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *originRulesRewriteConfigResource) updateConfig(model *originRulesRewriteConfigModel) error {
	rules := make([]*cdnetworksapi.OriginRulesRewrite, 0)
	if model.OriginRulesRewrite != nil {
		for _, rulesRewrite := range model.OriginRulesRewrite {
			rule := &cdnetworksapi.OriginRulesRewrite{
				PathPattern:           rulesRewrite.PathPattern.ValueStringPointer(),
				PathPatternHttp:       rulesRewrite.PathPatternHttp.ValueStringPointer(),
				ExceptPathPattern:     rulesRewrite.ExceptPathPattern.ValueStringPointer(),
				ExceptPathPatternHttp: rulesRewrite.ExceptPathPatternHttp.ValueStringPointer(),
				IgnoreLetterCase:      rulesRewrite.IgnoreLetterCase.ValueBoolPointer(),
				OriginInfo:            rulesRewrite.OriginInfo.ValueStringPointer(),
				Priority:              rulesRewrite.Priority.ValueInt64Pointer(),
				OriginHost:            rulesRewrite.OriginHost.ValueStringPointer(),
				BeforeRewriteUri:      rulesRewrite.BeforeRewriteUri.ValueStringPointer(),
				AfterRewriteUri:       rulesRewrite.AfterRewriteUri.ValueStringPointer(),
			}
			rules = append(rules, rule)
		}
	}

	updateOriginAndOriginHostRequest := cdnetworksapi.UpdateOriginUriAndOriginHostRequest{
		OriginRulesRewrites: rules,
	}

	_, err := r.client.UpdateOriginUriAndOriginHost(model.DomainId.ValueString(), updateOriginAndOriginHostRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}

func (r *originRulesRewriteConfigResource) updateModel(model *originRulesRewriteConfigModel) error {
	originRulesRewritesConfigResponse, err := r.client.QueryOriginUriAndOriginHost(model.DomainId.ValueString())
	if err != nil {
		return err
	}

	model.OriginRulesRewrite = make([]*originRulesRewriteModel, 0)
	for _, ruleEntry := range originRulesRewritesConfigResponse.OriginRulesRewrites {
		model.OriginRulesRewrite = append(model.OriginRulesRewrite, &originRulesRewriteModel{
			PathPattern:           types.StringPointerValue(ruleEntry.PathPattern),
			PathPatternHttp:       types.StringPointerValue(ruleEntry.PathPatternHttp),
			ExceptPathPattern:     types.StringPointerValue(ruleEntry.ExceptPathPattern),
			ExceptPathPatternHttp: types.StringPointerValue(ruleEntry.ExceptPathPatternHttp),
			IgnoreLetterCase:      types.BoolPointerValue(ruleEntry.IgnoreLetterCase),
			OriginInfo:            types.StringPointerValue(ruleEntry.OriginInfo),
			Priority:              types.Int64PointerValue(ruleEntry.Priority),
			OriginHost:            types.StringPointerValue(ruleEntry.OriginHost),
			BeforeRewriteUri:      types.StringPointerValue(ruleEntry.BeforeRewriteUri),
			AfterRewriteUri:       types.StringPointerValue(ruleEntry.AfterRewriteUri),
		})
	}

	return nil
}
