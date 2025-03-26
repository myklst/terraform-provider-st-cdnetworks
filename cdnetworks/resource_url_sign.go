package cdnetworks

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type urlSignResourceModel struct {
	DomainId                 types.String `tfsdk:"domain_id"`
	PrimaryKey               types.String `tfsdk:"primary_key"`
	SecondaryKey             types.String `tfsdk:"secondary_key"`
	Ttl                      types.Int64  `tfsdk:"ttl"`
	PathPattern              types.String `tfsdk:"path_pattern"`
	CipherCombination        types.String `tfsdk:"cipher_combination"`
	CipherParam              types.String `tfsdk:"cipher_param"`
	TimeParam                types.String `tfsdk:"time_param"`
	TimeFormat               types.String `tfsdk:"time_format"`
	RequestUrlStyle          types.String `tfsdk:"request_url_style"`
	DstStyle                 types.Int64  `tfsdk:"dst_style"`
	EncryptMethod            types.String `tfsdk:"encrypt_method"`
	LogFormat                types.Bool   `tfsdk:"log_format"`
	IgnoreUriSlash           types.Bool   `tfsdk:"ignore_uri_slash"`
	IgnoreKeyAndTimePosition types.Bool   `tfsdk:"ignore_key_and_time_position"`
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
				Description: "Domain id.",
				Required:    true,
			},
			"primary_key": &schema.StringAttribute{
				Description: "Primary key of the URL Signature.",
				Required:    true,
				Sensitive:   true,
			},
			"secondary_key": &schema.StringAttribute{
				Description: "Backup key of the URL Signature.",
				Required:    true,
				Sensitive:   true,
			},
			"ttl": &schema.Int64Attribute{
				Description: `TTL of the URL Signature, in seconds.`,
				Required:    true,
			},
			"path_pattern": &schema.StringAttribute{
				Description: "URL matching mode, supports regular expressions. Timestamp anti-hotlink verification is performed on the matched URLs; unmatched URLs are rejected.",
				Required:    true,
			},
			"cipher_combination": &schema.StringAttribute{
				Description: "Anti-hotlink generation method, parameters involved in MD5 calculation and combination order.",
				Required:    true,
			},
			"cipher_param": &schema.StringAttribute{
				Description: "Parameter name of the anti-hotlink string.",
				Required:    true,
			},
			"time_param": &schema.StringAttribute{
				Description: "Parameter name of the time string.",
				Required:    true,
			},
			"time_format": &schema.StringAttribute{
				Description: "Anti-hotlink encryption string time format, multiple selections are allowed, separated by semicolons (;).",
				Required:    true,
			},
			"request_url_style": &schema.StringAttribute{
				Description: "Anti-hotlink request URL format.",
				Required:    true,
			},
			"dst_style": &schema.Int64Attribute{
				Description: "Anti-hotlink return method. Values: 1 (use unencrypted URL to return to the source), 2 (use the URL with encrypted string requested by the customer to return to the source).",
				Required:    true,
			},
			"encrypt_method": &schema.StringAttribute{
				Description: "Encryption algorithm. Currently supported parameters: md5sum.",
				Required:    true,
			},
			"log_format": &schema.BoolAttribute{
				Description: "Logging original url.",
				Required:    true,
			},
			"ignore_uri_slash": &schema.BoolAttribute{
				Description: "Remove / from $url in hotlink protection.",
				Required:    true,
			},
			"ignore_key_and_time_position": &schema.BoolAttribute{
				Description: "Key and time can be interchanged.",
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
		resp.Diagnostics.AddError("[API ERROR] Failed to Create URL Sign", err.Error())
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
	state.PathPattern = types.StringValue(*queryURLSignResponse.TimestampVisitControlRule.PathPattern)
	state.CipherCombination = types.StringValue(*queryURLSignResponse.TimestampVisitControlRule.CipherCombination)
	state.CipherParam = types.StringValue(*queryURLSignResponse.TimestampVisitControlRule.CipherParam)
	state.TimeParam = types.StringValue(*queryURLSignResponse.TimestampVisitControlRule.TimeParam)
	state.TimeFormat = types.StringValue(*queryURLSignResponse.TimestampVisitControlRule.TimeFormat)
	state.RequestUrlStyle = types.StringValue(*queryURLSignResponse.TimestampVisitControlRule.RequestUrlStyle)
	state.DstStyle = types.Int64Value(*queryURLSignResponse.TimestampVisitControlRule.DstStyle)
	state.EncryptMethod = types.StringValue(*queryURLSignResponse.TimestampVisitControlRule.EncryptMethod)
	state.LogFormat = types.BoolValue(*queryURLSignResponse.TimestampVisitControlRule.LogFormat == "true")
	state.IgnoreUriSlash = types.BoolValue(*queryURLSignResponse.TimestampVisitControlRule.IgnoreUriSlash == "true")
	state.IgnoreKeyAndTimePosition = types.BoolValue(*queryURLSignResponse.TimestampVisitControlRule.IgnoreKeyAndTimePosition == "true")

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
	state.PathPattern = plan.PathPattern
	state.CipherCombination = plan.CipherCombination
	state.CipherParam = plan.CipherParam
	state.TimeParam = plan.TimeParam
	state.TimeFormat = plan.TimeFormat
	state.RequestUrlStyle = plan.RequestUrlStyle
	state.DstStyle = plan.DstStyle
	state.EncryptMethod = plan.EncryptMethod
	state.LogFormat = plan.LogFormat
	state.IgnoreUriSlash = plan.IgnoreUriSlash
	state.IgnoreKeyAndTimePosition = plan.IgnoreKeyAndTimePosition

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
		resp.Diagnostics.AddError("[API ERROR] Failed to Delete URL Sign", err.Error())
		return
	}
}

func (r *urlSignResource) updateUrlSign(model *urlSignResourceModel) error {
	updateUrlSignRequest := cdnetworksapi.UpdateURLSignRequest{
		TimestampVisitControlRule: &cdnetworksapi.TimestampVisitControlRule{
			PathPattern:              model.PathPattern.ValueStringPointer(),
			CipherCombination:        model.CipherCombination.ValueStringPointer(),
			CipherParam:              model.CipherParam.ValueStringPointer(),
			TimeParam:                model.TimeParam.ValueStringPointer(),
			LowerLimitExpireTime:     model.Ttl.ValueInt64Pointer(),
			UpperLimitExpireTime:     model.Ttl.ValueInt64Pointer(),
			MultipleSecretKeys:        types.StringValue(model.PrimaryKey.ValueString() + ";" + model.SecondaryKey.ValueString()).ValueStringPointer(),
			TimeFormat:               model.TimeFormat.ValueStringPointer(),
			RequestUrlStyle:          model.RequestUrlStyle.ValueStringPointer(),
			DstStyle:                 model.DstStyle.ValueInt64Pointer(),
			EncryptMethod:            model.EncryptMethod.ValueStringPointer(),
			LogFormat:                types.StringValue(strconv.FormatBool(*model.LogFormat.ValueBoolPointer())).ValueStringPointer(),
			IgnoreUriSlash:           types.StringValue(strconv.FormatBool(*model.IgnoreUriSlash.ValueBoolPointer())).ValueStringPointer(),
			IgnoreKeyAndTimePosition: types.StringValue(strconv.FormatBool(*model.IgnoreKeyAndTimePosition.ValueBoolPointer())).ValueStringPointer(),
		},
	}

	_, err := r.client.UpdateURLSign(model.DomainId.ValueString(), updateUrlSignRequest)

	if err != nil {
		return err
	}

	return nil
}
