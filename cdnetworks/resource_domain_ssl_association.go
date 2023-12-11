package cdnetworks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type domainSslAssociationModel struct {
	DomainId         types.String `tfsdk:"domain_id"`
	UseSsl           types.Bool   `tfsdk:"use_ssl"`
	SslCertificateId types.String `tfsdk:"ssl_certificate_id"`
}

type domainSslAssociationResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &domainSslAssociationResource{}
	_ resource.ResourceWithConfigure   = &domainSslAssociationResource{}
	_ resource.ResourceWithModifyPlan  = &domainSslAssociationResource{}
	_ resource.ResourceWithImportState = &domainSslAssociationResource{}
)

func NewDomainSslAssociationResource() resource.Resource {
	return &domainSslAssociationResource{}
}

func (r *domainSslAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_ssl_association"
}

func (r *domainSslAssociationResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "",
		Attributes: map[string]schema.Attribute{
			"domain_id": &schema.StringAttribute{
				Description: "",
				Required:    true,
			},
			"use_ssl": &schema.BoolAttribute{
				Description: "",
				Required:    true,
			},
			"ssl_certificate_id": &schema.StringAttribute{
				Description: "",
				Optional:    true,
			},
		},
	}
}

func (r *domainSslAssociationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *domainSslAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *domainSslAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateCdnDomainRequest := cdnetworksapi.UpdateCdnDomainRequest{
		Ssl: &cdnetworksapi.Ssl{
			UseSsl:           model.UseSsl.ValueBoolPointer(),
			SslCertificateId: model.SslCertificateId.ValueStringPointer(),
		},
	}
	_, err := r.client.UpdateCdnDomain(model.DomainId.ValueString(), updateCdnDomainRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Add DomainSslAssociation", err.Error())
		return
	}
	err = utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}
	resp.State.Set(ctx, model)
}

func (r *domainSslAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *domainSslAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	queryCdnDomainResponse, err := r.client.QueryCdnDomain(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query DomainSslAssociation", err.Error())
		return
	}
	if queryCdnDomainResponse.Ssl != nil {
		model.UseSsl = types.BoolPointerValue(queryCdnDomainResponse.Ssl.UseSsl)
		model.SslCertificateId = types.StringPointerValue(queryCdnDomainResponse.Ssl.SslCertificateId)
	} else {
		model.UseSsl = types.BoolValue(false)
		model.SslCertificateId = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *domainSslAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *domainSslAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateCdnDomainRequest := cdnetworksapi.UpdateCdnDomainRequest{
		Ssl: &cdnetworksapi.Ssl{
			UseSsl:           plan.UseSsl.ValueBoolPointer(),
			SslCertificateId: plan.SslCertificateId.ValueStringPointer(),
		},
	}
	_, err := r.client.UpdateCdnDomain(plan.DomainId.ValueString(), updateCdnDomainRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Update DomainSslAssociation", err.Error())
		return
	}
	err = utils.WaitForDomainDeployed(r.client, plan.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainSslAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model domainSslAssociationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	useSsl := false
	updateCdnDomainRequest := cdnetworksapi.UpdateCdnDomainRequest{
		Ssl: &cdnetworksapi.Ssl{
			UseSsl: &useSsl,
		},
	}
	_, err := r.client.UpdateCdnDomain(model.DomainId.ValueString(), updateCdnDomainRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Delete DomainSslAssociation", err.Error())
		return
	}
	err = utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Check DomainStatus", err.Error())
		return
	}
}

func (r *domainSslAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *domainSslAssociationResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *domainSslAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if plan == nil {
		return
	}
	if plan.UseSsl.ValueBool() && plan.SslCertificateId.IsNull() {
		resp.Diagnostics.AddError("[Validate Config] Invalid config", "`ssl_certificate_id` is required when `use_ssl` is true")
		return
	}
}
