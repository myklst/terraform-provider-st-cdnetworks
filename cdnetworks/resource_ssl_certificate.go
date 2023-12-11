package cdnetworks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type sslCertificateResourceModel struct {
	Id             types.String `tfsdk:"ssl_certificate_id"`
	Name           types.String `tfsdk:"name"`
	Comment        types.String `tfsdk:"comment"`
	SslCertificate types.String `tfsdk:"ssl_certificate"`
	SslKey         types.String `tfsdk:"ssl_key"`
}

type sslCertificateResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &sslCertificateResource{}
	_ resource.ResourceWithConfigure   = &sslCertificateResource{}
	_ resource.ResourceWithImportState = &sslCertificateResource{}
)

func NewSslCertificateResource() resource.Resource {
	return &sslCertificateResource{}
}

func (r *sslCertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl_certificate"
}

func (r *sslCertificateResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource provides a SSL certificate for domain",
		Attributes: map[string]schema.Attribute{
			"ssl_certificate_id": &schema.StringAttribute{
				Description: "certificate Id",
				Computed:    true,
			},
			"name": &schema.StringAttribute{
				Description: "Certificate name",
				Required:    true,
			},
			"comment": &schema.StringAttribute{
				Description: "comment",
				Optional:    true,
			},
			"ssl_certificate": &schema.StringAttribute{
				Description: "Certificate, PEM certificate, including CRT file and CA file.",
				Required:    true,
			},
			"ssl_key": &schema.StringAttribute{
				Description: "Private key of the certificate, PEM certificate.",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *sslCertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *sslCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model sslCertificateResourceModel
	diags := req.Plan.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	addCertificateRequest := cdnetworksapi.AddCertificateV2Request{
		Name:        model.Name.ValueStringPointer(),
		Certificate: model.SslCertificate.ValueStringPointer(),
		PrivateKey:  model.SslKey.ValueStringPointer(),
		Comment:     model.Comment.ValueStringPointer(),
	}

	addCertificateResponse, err := r.client.AddCertificateV2(addCertificateRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Add Certificate", err.Error())
		return
	}
	if *addCertificateResponse.Code != "0" {
		resp.Diagnostics.AddError("[API ERROR] Fail to Add Certificate", *addCertificateResponse.Message)
		return
	}
	model.Id = types.StringValue(*addCertificateResponse.CertificateId)
	resp.State.Set(ctx, &model)
}

func (r *sslCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *sslCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryCertificateInfoResponse, err := r.client.QueryCertificateInfo(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Certificate", err.Error())
		return
	}
	if *queryCertificateInfoResponse.Code == 19638021 {
		resp.State.RemoveResource(ctx)
		return
	} else if *queryCertificateInfoResponse.Code != 0 {
		resp.Diagnostics.AddError("[API ERROR] Fail to Query Certificate", *queryCertificateInfoResponse.Message)
		return
	}

	state.Name = types.StringPointerValue(queryCertificateInfoResponse.QueryCertificateInfoResponseData.Name)
	state.Comment = types.StringPointerValue(queryCertificateInfoResponse.QueryCertificateInfoResponseData.Comment)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *sslCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan sslCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateCertificateRequest := cdnetworksapi.UpdateCertificateV2Request{
		Name:        plan.Name.ValueStringPointer(),
		Certificate: plan.SslCertificate.ValueStringPointer(),
		PrivateKey:  plan.SslKey.ValueStringPointer(),
		Comment:     plan.Comment.ValueStringPointer(),
	}

	updateCertificateResponse, err := r.client.UpdateCertificateV2(state.Id.ValueString(), updateCertificateRequest)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Update Certificate", err.Error())
		return
	}
	if *updateCertificateResponse.Code != "0" {
		resp.Diagnostics.AddError("[API ERROR] Failed to Update Certificate", *updateCertificateResponse.Message)
		return
	}
	state.Name = plan.Name
	state.SslCertificate = plan.SslCertificate
	state.SslKey = plan.SslKey

	resp.State.Set(ctx, state)
}

func (r *sslCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model sslCertificateResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteCertificateResponse, err := r.client.DeleteCertificateV2(model.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Del Certificate", err.Error())
		return
	}
	if *deleteCertificateResponse.Code != "0" {
		resp.Diagnostics.AddError("[API ERROR] Failed to Del Certificate", *deleteCertificateResponse.Message)
		return
	}
}

func (r *sslCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("ssl_certificate_id"), req, resp)
}
