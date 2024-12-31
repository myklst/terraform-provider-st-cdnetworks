package cdnetworks

import (
	"context"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/common"
	. "github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/model"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type floodShieldDomainResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &floodShieldDomainResource{}
	_ resource.ResourceWithConfigure   = &floodShieldDomainResource{}
	_ resource.ResourceWithImportState = &floodShieldDomainResource{}
	_ resource.ResourceWithModifyPlan  = &floodShieldDomainResource{}
)

func NewFloodShieldDomainResource() resource.Resource {
	return &floodShieldDomainResource{}
}

func (r *floodShieldDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flood_shield_domain"
}

func (r *floodShieldDomainResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainSchema
}

func (r *floodShieldDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *floodShieldDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Add mutex lock to make sure the resources is created 1 by 1.
	// As BindControlGroup API might get overwritten.
	mutex.Lock()
	defer func() {
		mutex.Unlock()
	}()

	var model *DomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addCdnDomainRequest := cdnetworksapi.AddCdnDomainRequest{
		Version:           API_VERSION,
		DomainName:        model.Domain.ValueStringPointer(),
		ConfigFormId:      model.ConfigFormId.ValueStringPointer(),
		AccelerateNoChina: model.AccelerateNoChina.ValueStringPointer(),
		ContractId:        model.ContractId.ValueStringPointer(),
		ItemId:            model.ItemId.ValueStringPointer(),
		Comment:           model.Comment.ValueStringPointer(),
		HeaderOfClientIp:  model.HeaderOfClientIp.ValueStringPointer(),
		OriginConfig:      model.BuildApiOriginConfig(),
	}

	addCdnDomainResponse, err := r.client.AddCdnDomain(addCdnDomainRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Add Flood Shield Domain", err.Error())
		return
	}

	model.DomainId = types.StringValue(*addCdnDomainResponse.DomainId)
	model.Status = types.StringValue("InProgress")

	// Save state after cdn is created, prevent become orphan.
	// But will prompt error for those field that required 'computed' but not inputted.
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)

	// Append newly added cdn domains to control_group, to bind to specific account.
	if model.ControlGroup != nil {
		err = common.BindCdnDomainToControlGroup(r.client, model)
		if err != nil {
			resp.Diagnostics.AddError("[API ERROR] Fail to Bind Control Group", err.Error())
			return
		}
	}

	err = utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}

	// Only trigger UpdateCdnDomain() when cache_host is not "".
	if model.CacheHost.ValueString() != "" {
		updateCdnDomainRequest := cdnetworksapi.UpdateCdnDomainRequest{
			Version:    API_VERSION,
			DomainName: model.Domain.ValueStringPointer(),
			CacheHost:  model.CacheHost.ValueStringPointer(),
		}
		_, err := r.client.UpdateCdnDomain(model.DomainId.ValueString(), updateCdnDomainRequest)
		if err != nil {
			resp.Diagnostics.AddError("[API ERROR] Fail to Update Flood Shield Cache-host for Domain", err.Error())
			return
		}
	}

	// Required as copying computedFields from queryResponse.
	queryCdnDomainResponse, err := r.client.QueryCdnDomain(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Flood Shield Domain", err.Error())
		return
	}
	model.CopyComputedFields(&queryCdnDomainResponse)

	resp.State.Set(ctx, model)
}

func (r *floodShieldDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *DomainResourceModel
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

	queryCdnDomain := func() error {
		queryCdnDomainResponse, err := r.client.QueryCdnDomain(domain)
		if err != nil {
			if cdnErr, ok := err.(*cdnetworksapi.ErrorResponse); ok {
				if cdnErr.ResponseCode == "WRONG_OPERATOR" {
					if model.ControlGroup != nil {
						resp.Diagnostics.AddWarning("[Call API] Trying to bind CDN Domain to Control Group.", fmt.Sprintf("Domain: %s", model.Domain.ValueString()))
						// Bind CDN domains to ControlGroup, in case previous bind action doesn't complete.
						// Prevent error from Read(), Create() might failed to bind into controlGroup.
						err = common.BindCdnDomainToControlGroup(r.client, model)
						if err != nil {
							return backoff.Permanent(fmt.Errorf("bind control group API error. err: %v", err))
						}
					}

					// Retry for queryCdnDomain action
					return cdnErr
				}
				return backoff.Permanent(fmt.Errorf("unexpected error code. code: %s err: %v", cdnErr.ResponseCode, err))
			}
			return backoff.Permanent(fmt.Errorf("queryCdnDomain API error, err: %v", err))
		}

		model.UpdateDomainFromApiConfig(ctx, &queryCdnDomainResponse)
		return nil
	}

	err := backoff.Retry(queryCdnDomain, backoff.NewExponentialBackOff())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Flood Shield Domain", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *floodShieldDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			CacheHost:        plan.CacheHost.ValueStringPointer(),
			HeaderOfClientIp: plan.HeaderOfClientIp.ValueStringPointer(),
			OriginConfig:     plan.BuildApiOriginConfig(),
		}
		_, err := r.client.UpdateCdnDomain(plan.DomainId.ValueString(), updateCdnDomainRequest)
		if err != nil {
			resp.Diagnostics.AddError("[API ERROR] Fail to Update Flood Shield Domain", err.Error())
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
			resp.Diagnostics.AddError("[API ERROR] Fail to Enable/Disable Flood Shield Domain", err.Error())
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
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Flood Shield Domain", err.Error())
		return
	}
	plan.CopyComputedFields(&queryCdnDomainResponse)

	resp.State.Set(ctx, plan)
}

func (r *floodShieldDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteApiDomain(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Delete Flood Shield Domain", err.Error())
		return
	}

	err = utils.WaitForDomainDeleted(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}
}

func (r *floodShieldDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}

func (r *floodShieldDomainResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
