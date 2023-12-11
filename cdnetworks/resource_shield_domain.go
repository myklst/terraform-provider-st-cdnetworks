package cdnetworks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	. "github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/model"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type shieldDomainResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &shieldDomainResource{}
	_ resource.ResourceWithConfigure   = &shieldDomainResource{}
	_ resource.ResourceWithImportState = &shieldDomainResource{}
	_ resource.ResourceWithModifyPlan  = &shieldDomainResource{}
)

func NewShieldDomainResource() resource.Resource {
	return &shieldDomainResource{}
}

func (r *shieldDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shield_domain"
}

func (r *shieldDomainResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainSchema
}

func (r *shieldDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *shieldDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addCdnDomainRequest := cdnetworksapi.AddCdnDomainRequest{
		Version:          API_VERSION,
		DomainName:       model.Domain.ValueStringPointer(),
		ContractId:       model.ContractId.ValueStringPointer(),
		ItemId:           model.ItemId.ValueStringPointer(),
		Comment:          model.Comment.ValueStringPointer(),
		HeaderOfClientIp: model.HeaderOfClientIp.ValueStringPointer(),
		OriginConfig:     model.BuildApiOriginConfig(),
	}
	addCdnDomainResponse, err := r.client.AddCdnDomain(addCdnDomainRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Add Shield Domain", err.Error())
		return
	}

	err = utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}

	model.DomainId = types.StringValue(*addCdnDomainResponse.DomainId)
	queryCdnDomainResponse, err := r.client.QueryCdnDomain(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Shield Domain", err.Error())
		return
	}
	model.CopyComputedFields(&queryCdnDomainResponse)

	resp.State.Set(ctx, model)
}

func (r *shieldDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var domain string
	if !model.DomainId.IsNull() {
		domain = model.DomainId.ValueString()
	} else {
		domain = model.Domain.ValueString()
	}
	queryCdnDomainResponse, err := r.client.QueryCdnDomain(domain)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Shield Domain", err.Error())
		return
	}

	model.UpdateDomainFromApiConfig(ctx, &queryCdnDomainResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *shieldDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.DomainId = state.DomainId

	if state.Enabled.ValueBool() {
		updateCdnDomainRequest := cdnetworksapi.UpdateCdnDomainRequest{
			Version:          API_VERSION,
			DomainName:       plan.Domain.ValueStringPointer(),
			Comment:          plan.Comment.ValueStringPointer(),
			HeaderOfClientIp: plan.HeaderOfClientIp.ValueStringPointer(),
			OriginConfig:     plan.BuildApiOriginConfig(),
		}
		_, err := r.client.UpdateCdnDomain(plan.DomainId.ValueString(), updateCdnDomainRequest)
		if err != nil {
			resp.Diagnostics.AddError("[API ERROR] Fail to Update Shield Domain", err.Error())
			return
		}
	} else if plan.Enabled.Equal(state.Enabled) {
		resp.Diagnostics.AddError("[API ERROR] Update disabled domain is not Allowed", "")
		return
	}

	var err error

	if !plan.Enabled.Equal(state.Enabled) && !plan.Enabled.IsNull() {
		if plan.Enabled.ValueBool() {
			_, err = r.client.EnableDomain(plan.DomainId.ValueString())
		} else {
			_, err = r.client.DisableDomain(plan.DomainId.ValueString())
		}
		if err != nil {
			resp.Diagnostics.AddError("[API ERROR] Fail to Enable/Disable Shield Domain", err.Error())
			return
		}
	}

	err = utils.WaitForDomainDeployed(r.client, plan.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}

	queryCdnDomainResponse, err := r.client.QueryCdnDomain(plan.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Shield Domain", err.Error())
		return
	}
	plan.CopyComputedFields(&queryCdnDomainResponse)

	resp.State.Set(ctx, plan)
}

func (r *shieldDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteApiDomain(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Delete Shield Domain", err.Error())
		return
	}

	err = utils.WaitForDomainDeleted(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}
}

func (r *shieldDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}

func (r *shieldDomainResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if plan == nil {
		return
	}
	plan.Fill()
	err := plan.Check()
	if err != nil {
		resp.Diagnostics.AddError("[Validate Config] Invalid Config", err.Error())
		return
	}
	resp.Plan.Set(ctx, plan)
}
