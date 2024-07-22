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

type contentAccelerationDomainResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &contentAccelerationDomainResource{}
	_ resource.ResourceWithConfigure   = &contentAccelerationDomainResource{}
	_ resource.ResourceWithImportState = &contentAccelerationDomainResource{}
	_ resource.ResourceWithModifyPlan  = &contentAccelerationDomainResource{}
)

func NewContentAccelerationDomainResource() resource.Resource {
	return &contentAccelerationDomainResource{}
}

func (r *contentAccelerationDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_acceleration_domain"
}

func (r *contentAccelerationDomainResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = DomainSchema
}

func (r *contentAccelerationDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *contentAccelerationDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *DomainResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addCdnDomainRequest := cdnetworksapi.AddCdnDomainRequest{
		Version:           API_VERSION,
		DomainName:        model.Domain.ValueStringPointer(),
		AccelerateNoChina: model.AccelerateNoChina.ValueStringPointer(),
		ContractId:        model.ContractId.ValueStringPointer(),
		ItemId:            model.ItemId.ValueStringPointer(),
		Comment:           model.Comment.ValueStringPointer(),
		HeaderOfClientIp:  model.HeaderOfClientIp.ValueStringPointer(),
		OriginConfig:      model.BuildApiOriginConfig(),
	}

	addCdnDomainResponse, err := r.client.AddCdnDomain(addCdnDomainRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Add Content Acceleration Domain", err.Error())
		return
	}

	model.DomainId = types.StringValue(*addCdnDomainResponse.DomainId)
	model.Status = types.StringValue("InProgress")

	// Save state after cdn is created, prevent become orphan.
	// But will prompt error for those field that required 'computed' but not inputted.
	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)

	// Append newly added cdn domains to control_group, to bind to specific account.
	common.BindCdnDomainToControlGroup(r.client, model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Bind Control Group", err.Error())
		return
	}

	err = utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}

	// Required as copying computedFields from queryResponse.
	queryCdnDomainResponse, err := r.client.QueryCdnDomain(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Content Acceleration Domain", err.Error())
		return
	}
	model.CopyComputedFields(&queryCdnDomainResponse)

	resp.State.Set(ctx, model)
}

func (r *contentAccelerationDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
					resp.Diagnostics.AddWarning("[Call API] Trying to bind Content Acceleration Domain to Control Group.", fmt.Sprintf("Domain: %s", model.Domain.ValueString()))
					// Bind CDN domains to ControlGroup, in case previous bind action doesn't complete.
					// Prevent error from Read(), Create() might failed to bind into controlGroup.
					common.BindCdnDomainToControlGroup(r.client, model)
					if err != nil {
						return backoff.Permanent(fmt.Errorf("bind control group API error. err: %v", err))
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
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Content Acceleration Domain", err.Error())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *contentAccelerationDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
			resp.Diagnostics.AddError("[API ERROR] Fail to Update Content Acceleration Domain", err.Error())
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
			resp.Diagnostics.AddError("[API ERROR] Fail to Enable/Disable Content Acceleration Domain", err.Error())
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
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Content Acceleration Domain", err.Error())
		return
	}
	plan.CopyComputedFields(&queryCdnDomainResponse)

	resp.State.Set(ctx, plan)
}

func (r *contentAccelerationDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *DomainResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteApiDomain(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Delete Content Acceleration Domain", err.Error())
		return
	}

	err = utils.WaitForDomainDeleted(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}
}

func (r *contentAccelerationDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain"), req, resp)
}

func (r *contentAccelerationDomainResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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
