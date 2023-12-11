package cdnetworks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type httpCodeCacheRuleModel struct {
	HttpCodes types.List  `tfsdk:"http_codes"`
	CacheTtl  types.Int64 `tfsdk:"cache_ttl"`
}

type httpCodeCacheConfigModel struct {
	DomainId           types.String              `tfsdk:"domain_id"`
	HttpCodeCacheRules []*httpCodeCacheRuleModel `tfsdk:"http_code_cache_rule"`
}

type httpCodeCacheConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &httpCodeCacheConfigResource{}
	_ resource.ResourceWithConfigure   = &httpCodeCacheConfigResource{}
	_ resource.ResourceWithImportState = &httpCodeCacheConfigResource{}
)

func NewHttpCodeCacheConfigResource() resource.Resource {
	return &httpCodeCacheConfigResource{}
}

func (r *httpCodeCacheConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_code_cache_config"
}

func (r *httpCodeCacheConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Http Code Cache Configuration`,
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"http_code_cache_rule": &schema.ListNestedBlock{
				Description: `State Code Caching Rule Configuration, parent node
1. When you need to set state code caching rules, this must be filled in.
2. Configuration of Clear State Code Caching Rules for .`,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"http_codes": &schema.ListAttribute{
							Description: "Configure HTTP status code list",
							ElementType: types.Int64Type,
							Required:    true,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
						"cache_ttl": &schema.Int64Attribute{
							Description: "Define the caching time of the specified status code in units s, 0 to indicate no caching",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *httpCodeCacheConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *httpCodeCacheConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *httpCodeCacheConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to set http_code_cache", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *httpCodeCacheConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		model *httpCodeCacheConfigModel
		diags diag.Diagnostics
	)
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryHttpCodeCacheConfigResponse, err := r.client.QueryHttpCodeCacheConfig(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to query http_code_cache", err.Error())
		return
	}
	model.HttpCodeCacheRules = make([]*httpCodeCacheRuleModel, 0)
	if queryHttpCodeCacheConfigResponse.HttpCodeCacheRules != nil {
		for _, rule := range queryHttpCodeCacheConfigResponse.HttpCodeCacheRules {
			ruleModel := &httpCodeCacheRuleModel{
				CacheTtl: types.Int64PointerValue(rule.CacheTtl),
			}
			ruleModel.HttpCodes, diags = types.ListValueFrom(ctx, types.Int64Type, rule.HttpCodes)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			model.HttpCodeCacheRules = append(model.HttpCodeCacheRules, ruleModel)
		}
	}
	resp.State.Set(ctx, model)
}

func (r *httpCodeCacheConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *httpCodeCacheConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to set http_code_cache", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *httpCodeCacheConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *httpCodeCacheConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.HttpCodeCacheRules = make([]*httpCodeCacheRuleModel, 0)
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete http_code_cache", err.Error())
	}
}

func (r *httpCodeCacheConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *httpCodeCacheConfigResource) updateConfig(model *httpCodeCacheConfigModel) error {
	rules := make([]*cdnetworksapi.HttpCodeCacheRule, 0)
	if model.HttpCodeCacheRules != nil {
		for _, ruleModel := range model.HttpCodeCacheRules {
			rule := &cdnetworksapi.HttpCodeCacheRule{
				CacheTtl: ruleModel.CacheTtl.ValueInt64Pointer(),
			}
			ruleModel.HttpCodes.ElementsAs(nil, &rule.HttpCodes, false)
			rules = append(rules, rule)
		}
	}
	updateHttpCodeCacheConfigRequest := cdnetworksapi.UpdateHttpCodeCacheConfigRequest{
		HttpCodeCacheRules: rules,
	}
	_, err := r.client.UpdateHttpCodeCacheConfig(model.DomainId.ValueString(), updateHttpCodeCacheConfigRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}
