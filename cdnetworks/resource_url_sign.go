package cdnetworks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type urlSignResourceModel struct {
	DomainId   types.String `tfsdk:"domain_id"`
	PrimaryKey types.String `tfsdk:"primary_key"`
	BackupKey  types.String `tfsdk:"backup_key"`
	Ttl        types.Int64  `tfsdk:"ttl"`
}

type urlSignResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource              = &urlSignResource{}
	_ resource.ResourceWithConfigure = &urlSignResource{}
)

func NewUrlSignResource() resource.Resource {
	return &urlSignResource{}
}

func (r *urlSignResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_url_sign"
}

func (r *urlSignResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource enable URL Signature for domain",
		Attributes: map[string]schema.Attribute{
			"domain_id": &schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
			"primary_key": &schema.StringAttribute{
				Description: "Primary key of the URL Signature.",
				Required:    true,
				Sensitive:   true,
			},
			"backup_key": &schema.StringAttribute{
				Description: "Backup key of the URL Signature.",
				Required:    true,
				Sensitive:   true,
			},
			"ttl": &schema.Int64Attribute{
				Description: `TTL of the URL Signature, in seconds.`,
				Required:    true,
			},
		},
	}
}

func (r *urlSignResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *urlSignResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *urlSignResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateUrlSign(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Enable URL Sign", err.Error())
		return
	}

	resp.State.Set(ctx, model)
}

func (r *urlSignResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *urlSignResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryURLSignResponse, err := r.client.QueryURLSign(state.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Read URL Sign", err.Error())
		return
	}

	state.DomainId = types.StringValue(*queryURLSignResponse.DomainId)
	state.Ttl = types.Int64Value(*queryURLSignResponse.TimestampVisitControlRule.LowerLimitExpireTime)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *urlSignResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan *urlSignResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateUrlSign(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Update URL Sign", err.Error())
		return
	}

	state.DomainId = plan.DomainId
	state.Ttl = plan.Ttl
	resp.State.Set(ctx, state)
}

func (r *urlSignResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *urlSignResourceModel
	diags := req.State.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UpdateURLSign(model.DomainId.ValueString(), cdnetworksapi.UpdateURLSignRequest{
		TimestampVisitControlRule: &cdnetworksapi.TimestampVisitControlRule{},
	})

	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Failed to Del URL Sign", err.Error())
		return
	}
}

func (r *urlSignResource) updateUrlSign(model *urlSignResourceModel) error {
	updateUrlSignRequest := cdnetworksapi.UpdateURLSignRequest{
		TimestampVisitControlRule: &cdnetworksapi.TimestampVisitControlRule{
			PathPattern:              types.StringValue("/*").ValueStringPointer(),
			CipherCombination:        types.StringValue("$uri$ourkey$time$args{rand}$args{uid}").ValueStringPointer(),
			CipherParam:              types.StringValue("auth_key").ValueStringPointer(),
			LowerLimitExpireTime:     model.Ttl.ValueInt64Pointer(),
			UpperLimitExpireTime:     model.Ttl.ValueInt64Pointer(),
			MultipleSecretKey:        types.StringValue(model.PrimaryKey.ValueString() + ";" + model.BackupKey.ValueString()).ValueStringPointer(),
			TimeFormat:               types.StringValue("7s").ValueStringPointer(),
			RequestUrlStyle:          types.StringValue("http://$domain/$uri?$args&auth_key=$time-$args{rand}-$args{uid}-$key").ValueStringPointer(),
			DstStyle:                 types.Int64Value(1).ValueInt64Pointer(),
			EncryptMethod:            types.StringValue("md5sum").ValueStringPointer(),
			LogFormat:                types.StringValue("false").ValueStringPointer(),
			IgnoreUriSlash:           types.StringValue("false").ValueStringPointer(),
			IgnoreKeyAndTimePosition: types.StringValue("false").ValueStringPointer(),
		},
	}

	_, err := r.client.UpdateURLSign(model.DomainId.ValueString(), updateUrlSignRequest)

	if err != nil {
		return err
	}

	return nil
}
