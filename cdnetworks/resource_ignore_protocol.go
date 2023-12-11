package cdnetworks

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type ignoreProtocolRuleModel struct {
	PathPattern         types.String `tfsdk:"path_pattern"`
	CacheIgnoreProtocol types.Bool   `tfsdk:"cache_ignore_protocol"`
	PurgeIgnoreProtocol types.Bool   `tfsdk:"purge_ignore_protocol"`
}

type ignoreProtocolModel struct {
	DomainId            types.String               `tfsdk:"domain_id"`
	IgnoreProtocolRules []*ignoreProtocolRuleModel `tfsdk:"ignore_protocol_rule"`
}

type ignoreProtocolResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &ignoreProtocolResource{}
	_ resource.ResourceWithConfigure   = &ignoreProtocolResource{}
	_ resource.ResourceWithModifyPlan  = &ignoreProtocolResource{}
	_ resource.ResourceWithImportState = &ignoreProtocolResource{}
)

func NewIgnoreProtocolResource() resource.Resource {
	return &ignoreProtocolResource{}
}

func (r *ignoreProtocolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ignore_protocol"
}

func (r *ignoreProtocolResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Ignore protocol caching and push configuration, parent tags
1. This must be filled when protocol cache and push configuration need to be ignored
2. Clear the configuration ignore about protocol cache and pushing`,
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"ignore_protocol_rule": &schema.ListNestedBlock{
				Description: `Ignore protocol configuration`,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path_pattern": &schema.StringAttribute{
							Description: "Url matching pattern, support regular, if all matches, input parameters can be configured as:.*",
							Optional:    true,
						},
						"cache_ignore_protocol": &schema.BoolAttribute{
							Description: `Whether to ignore the protocol cache, with allowable values of true and false. True turns on the HTTP/HTTPS Shared cache. Not on by default.`,
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"purge_ignore_protocol": &schema.BoolAttribute{
							Description: `It is recommended to use with cache-ignore protocol to avoid push failure.
Note:
1. Once configured, the global effect is not applied to the matched path-pattern.
2. Directory push does not distinguish protocols, while url push can distinguish protocols`,
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
					},
				},
			},
		},
	}
}

func (r *ignoreProtocolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *ignoreProtocolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *ignoreProtocolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to set ignore protocol", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *ignoreProtocolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *ignoreProtocolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateModel(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to read ignore protocol", err.Error())
		return
	}

	resp.State.Set(ctx, model)
}

func (r *ignoreProtocolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *ignoreProtocolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to set ignore protocol", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *ignoreProtocolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *ignoreProtocolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.IgnoreProtocolRules = make([]*ignoreProtocolRuleModel, 0)
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete ignore protocol", err.Error())
	}
}

func (r *ignoreProtocolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *ignoreProtocolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan == nil {
		return
	}

	if plan.IgnoreProtocolRules == nil || len(plan.IgnoreProtocolRules) == 0 {
		resp.Diagnostics.AddError("[Validate Config]Invalid Config", "ignore_protocol_rule is required")
		return
	}

	for _, ruleModel := range plan.IgnoreProtocolRules {
		if ruleModel.CacheIgnoreProtocol.ValueBool() && !ruleModel.PurgeIgnoreProtocol.ValueBool() {
			resp.Diagnostics.AddError("[Validate Config]Invalid Config", "If the protocol cache will be ignored，the protocol push must be ignored")
			return
		}
	}

	resp.Plan.Set(ctx, plan)
}

func (r *ignoreProtocolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *ignoreProtocolResource) updateConfig(model *ignoreProtocolModel) error {
	rules := make([]*cdnetworksapi.IgnoreProtocolRule, 0)
	if model.IgnoreProtocolRules != nil {
		for _, ruleModel := range model.IgnoreProtocolRules {
			rule := &cdnetworksapi.IgnoreProtocolRule{
				PathPattern:         ruleModel.PathPattern.ValueStringPointer(),
				CacheIgnoreProtocol: ruleModel.CacheIgnoreProtocol.ValueBoolPointer(),
				PurgeIgnoreProtocol: ruleModel.PurgeIgnoreProtocol.ValueBoolPointer(),
			}
			rules = append(rules, rule)
		}
	}
	updateIgnoreProtocolRequest := cdnetworksapi.UpdateIgnoreProtocolRequest{
		IgnoreProtocolRules: rules,
	}
	_, err := r.client.UpdateIgnoreProtocol(model.DomainId.ValueString(), updateIgnoreProtocolRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}

func (r *ignoreProtocolResource) updateModel(model *ignoreProtocolModel) error {
	ruleIndexMap := make(map[string][]int)
	for i, rule := range model.IgnoreProtocolRules {
		list, ok := ruleIndexMap[rule.String()]
		if !ok {
			list = make([]int, 0)
		}
		ruleIndexMap[rule.String()] = append(list, i)
	}

	queryIgnoreProtocolResponse, err := r.client.QueryIgnoreProtocol(model.DomainId.ValueString())
	if err != nil {
		return err
	}
	model.IgnoreProtocolRules = make([]*ignoreProtocolRuleModel, 0)
	if queryIgnoreProtocolResponse.IgnoreProtocolRules != nil {
		for _, rule := range queryIgnoreProtocolResponse.IgnoreProtocolRules {
			ruleModel := &ignoreProtocolRuleModel{
				PathPattern:         types.StringPointerValue(rule.PathPattern),
				CacheIgnoreProtocol: types.BoolPointerValue(rule.CacheIgnoreProtocol),
				PurgeIgnoreProtocol: types.BoolPointerValue(rule.PurgeIgnoreProtocol),
			}
			model.IgnoreProtocolRules = append(model.IgnoreProtocolRules, ruleModel)
		}
	}

	// Sort rules to align with .tf configuration file
	ignoreProtocolRules := make([]*ignoreProtocolRuleModel, len(model.IgnoreProtocolRules))
	unmatchedIgnoreProtocolRules := make([]*ignoreProtocolRuleModel, 0)
	for i := 0; i < len(model.IgnoreProtocolRules); i++ {
		rule := model.IgnoreProtocolRules[i]
		list, ok := ruleIndexMap[rule.String()]
		if ok && len(list) > 0 {
			ignoreProtocolRules[list[0]] = rule
			ruleIndexMap[rule.String()] = list[1:]
		} else {
			unmatchedIgnoreProtocolRules = append(unmatchedIgnoreProtocolRules, rule)
		}
	}
	for i, j := 0, 0; i < len(ignoreProtocolRules) && j < len(unmatchedIgnoreProtocolRules); i++ {
		if ignoreProtocolRules[i] != nil {
			continue
		}
		ignoreProtocolRules[i] = unmatchedIgnoreProtocolRules[j]
		j++
	}
	model.IgnoreProtocolRules = ignoreProtocolRules

	return nil
}

func (rule *ignoreProtocolRuleModel) String() string {
	attrs := []string{
		rule.PathPattern.String(),
		rule.CacheIgnoreProtocol.String(),
		rule.PurgeIgnoreProtocol.String(),
	}
	return strings.Join(attrs, "$$")
}
