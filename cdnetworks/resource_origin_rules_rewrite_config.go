package cdnetworks

import (
	"context"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type originRulesRewriteModel struct {
	DataId                types.Int64  `tfsdk:"data_id"`
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
						"data_id": &schema.Int64Attribute{
							Description: "Used by CDNetworks to keep track of the individual configuration",
							Optional:    true,
							Computed:    true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
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

	err := r.updateConfig(model, nil)
	if err != nil {
		resp.Diagnostics.AddError("[API Error]Fail to update origin_rules_rewrites", err.Error())
		return
	}

	err = r.updateModel(model)
	if err != nil {
		resp.Diagnostics.AddError("[API Error]Fail to query origin_rules_rewrites", err.Error())
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

	var state *originRulesRewriteConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CDNetworks's way to perform a delete on a single origin_rewrite_rule
	// is to pass in only the dataId of the rule that has been marked for deletion.

	// Extract the dataIds of the origin_rewrite_rules that have been deleted.
	// Do this by subtracting the plan dataIds from the state dataIds.
	// The remainder of the subtraction is marked as deleted.
	planDataIds := mapset.NewSet[int64]()
	stateDataIds := mapset.NewSet[int64]()

	for _, rule := range plan.OriginRulesRewrite {
		planDataIds.Add(rule.DataId.ValueInt64())
	}
	for _, rule := range state.OriginRulesRewrite {
		stateDataIds.Add(rule.DataId.ValueInt64())
	}

	deletedDataIds := stateDataIds.Difference(planDataIds)
	err := r.updateConfig(plan, deletedDataIds.ToSlice())
	if err != nil {
		resp.Diagnostics.AddError("[API Error]Fail to update origin_rules_rewrites", err.Error())
		return
	}

	err = r.updateModel(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API Error]Fail to query origin_rules_rewrites", err.Error())
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
	err := r.updateConfig(model, nil)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete origin_rules_rewrite", err.Error())
	}
}

func (r *originRulesRewriteConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *originRulesRewriteConfigResource) updateConfig(model *originRulesRewriteConfigModel, deletedDataIds []int64) error {
	rules := make([]*cdnetworksapi.OriginRulesRewrite, 0)
	if model.OriginRulesRewrite != nil {
		for _, rulesRewrite := range model.OriginRulesRewrite {
			rule := &cdnetworksapi.OriginRulesRewrite{
				DataId:                rulesRewrite.DataId.ValueInt64Pointer(),
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

	_, err := r.client.UpdateOriginUriAndOriginHost(model.DomainId.ValueString(), updateOriginAndOriginHostRequest, deletedDataIds)
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

	sort.Slice(originRulesRewritesConfigResponse.OriginRulesRewrites, func(i, j int) bool {
		return *originRulesRewritesConfigResponse.OriginRulesRewrites[i].DataId <
			*originRulesRewritesConfigResponse.OriginRulesRewrites[j].DataId
	})

	model.OriginRulesRewrite = make([]*originRulesRewriteModel, 0)
	for _, ruleEntry := range originRulesRewritesConfigResponse.OriginRulesRewrites {
		model.OriginRulesRewrite = append(model.OriginRulesRewrite, &originRulesRewriteModel{
			DataId:                types.Int64PointerValue(ruleEntry.DataId),
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
