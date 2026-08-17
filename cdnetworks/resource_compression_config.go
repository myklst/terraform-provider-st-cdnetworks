package cdnetworks

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type compressionSettingsModel struct {
	CompressionEnabled types.Bool   `tfsdk:"compression_enabled"`
	BrTypes            types.Bool   `tfsdk:"br_types"`
	IgnoreLetterCase   types.Bool   `tfsdk:"ignore_letter_case"`
	PathPattern        types.String `tfsdk:"path_pattern"`
	CustomPattern      types.String `tfsdk:"custom_pattern"`
	SpecifyUrlPattern  types.String `tfsdk:"specify_url_pattern"`
	Directory          types.String `tfsdk:"directory"`
	FileTypes          types.List   `tfsdk:"file_types"`
	FileTypeOthers     types.List   `tfsdk:"file_type_others"`
	CustomFileTypes    types.List   `tfsdk:"custom_file_types"`
	UriFileTypes       types.List   `tfsdk:"uri_file_types"`
}

type compressionConfigModel struct {
	DomainId            types.String              `tfsdk:"domain_id"`
	CompressionSettings *compressionSettingsModel `tfsdk:"compression_settings"`
}

type compressionConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &compressionConfigResource{}
	_ resource.ResourceWithConfigure   = &compressionConfigResource{}
	_ resource.ResourceWithModifyPlan  = &compressionConfigResource{}
	_ resource.ResourceWithImportState = &compressionConfigResource{}
)

func NewCompressionConfigResource() resource.Resource {
	return &compressionConfigResource{}
}

func (r *compressionConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compression_config"
}

func (r *compressionConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Compression response configuration. Controls whether the CDN compresses responses. Defining type of files or the matching scope of the request are compressed. When you need to cancel the compression configuration, you can pass in the empty node.",
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
			"compression_settings": &schema.SingleNestedAttribute{
				Description: `Compression configuration`,
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"compression_enabled": &schema.BoolAttribute{
						Description: "Whether to enable compression. The optional values are true and false. If it is empty, the default value is false. True means compression is on; false means compression is off.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"br_types": &schema.BoolAttribute{
						Description: "Whether to enable Brotli (br) compression. The optional values are true and false. If it is empty, the default value is false.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"ignore_letter_case": &schema.BoolAttribute{
						Description: "Whether to ignore the letter case of the file type. The optional values are true and false. If it is empty, the default value is false. True means to ignore case; false means not to ignore case.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"path_pattern": &schema.StringAttribute{
						Description: "The url matching mode supports regularization. If all matches, the input parameters can be configured as: .*",
						Optional:    true,
					},
					"custom_pattern": &schema.StringAttribute{
						Description: "Specify common types. Optional values: all (all files) and homepage (home page).",
						Optional:    true,
					},
					"directory": &schema.StringAttribute{
						Description: "Directory. Specify the directory to be compressed. Enter a legal directory format. Multiple separated by semicolons.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`^/.*/$`), `Must start and end with "/"`),
						},
					},
					"specify_url_pattern": &schema.StringAttribute{
						Description: "Specify the URL to be compressed. The INS format does not support the URI format with http(s)://.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(regexp.MustCompile(`[^(http(s)?://)].*`), "The input parameter does not support the URI format starting with http(s)://"),
						},
					},
					"file_types": &schema.ListAttribute{
						ElementType: types.StringType,
						Description: "(Legacy MIME type) List of File type. Specify the file type to be compressed. File types include: gif png bmp jpeg jpg html htm shtml mp3 wma flv mp4 wmv zip exe rar css txt ico js swf. If you need all types, pass all directly. Multiples are separated by semicolons, and all and specific file types cannot be configured at the same time.",
						Optional:    true,
					},
					"file_type_others": &schema.ListAttribute{
						ElementType: types.StringType,
						Description: "List of MIME file types to be compressed.",
						Optional:    true,
					},
					"custom_file_types": &schema.ListAttribute{
						ElementType: types.StringType,
						Description: "Custom file type. Fill in the appropriate identifiable file type according to your needs outside of the specified file type. Can be used with file_types. If file_types is also configured, the actual file type is the sum of the two parameters. Multiples are separated by semicolons.",
						Optional:    true,
					},
					"uri_file_types": &schema.ListAttribute{
						ElementType: types.StringType,
						Description: "List of URI file types to be compressed.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *compressionConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *compressionConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *compressionConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to create compression_config", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *compressionConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var (
		model *compressionConfigModel
		diags diag.Diagnostics
	)
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queryCompressionConfigResponse, err := r.client.QueryCompressionConfig(model.DomainId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to query compression_settings", err.Error())
		return
	}
	config := queryCompressionConfigResponse.CompressionSetting
	if config != nil {
		model.CompressionSettings = &compressionSettingsModel{
			CompressionEnabled: types.BoolPointerValue(config.CompressionEnabled),
			BrTypes:            types.BoolPointerValue(config.BrTypes),
			IgnoreLetterCase:   types.BoolPointerValue(config.IgnoreLetterCase),
			PathPattern:        types.StringPointerValue(config.PathPattern),
			CustomPattern:      types.StringPointerValue(config.CustomPattern),
			SpecifyUrlPattern:  types.StringPointerValue(config.SpecifyUrlPattern),
			Directory:          types.StringPointerValue(config.Directory),
		}

		model.CompressionSettings.FileTypes, diags = types.ListValueFrom(ctx, types.StringType, config.FileTypes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		model.CompressionSettings.CustomFileTypes, diags = types.ListValueFrom(ctx, types.StringType, config.CustomFileTypes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		model.CompressionSettings.FileTypeOthers, diags = types.ListValueFrom(ctx, types.StringType, config.FileTypeOthers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		model.CompressionSettings.UriFileTypes, diags = types.ListValueFrom(ctx, types.StringType, config.UriFileTypes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.State.Set(ctx, &model)
}

func (r *compressionConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *compressionConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to update compression_config", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *compressionConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *compressionConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.CompressionSettings = &compressionSettingsModel{}
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete compression_config", err.Error())
	}
}

func (r *compressionConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *compressionConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan == nil {
		return
	}

	err := plan.CompressionSettings.check()
	if err != nil {
		resp.Diagnostics.AddError("[Validate Config]Invalid Config", err.Error())
		return
	}

	resp.Plan.Set(ctx, plan)
}

func (r *compressionConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *compressionConfigResource) updateConfig(model *compressionConfigModel) error {
	compressionModel := model.CompressionSettings
	if compressionModel != nil {
		configs := &cdnetworksapi.CompressionSetting{
			CompressionEnabled: compressionModel.CompressionEnabled.ValueBoolPointer(),
			BrTypes:            compressionModel.BrTypes.ValueBoolPointer(),
			IgnoreLetterCase:   compressionModel.IgnoreLetterCase.ValueBoolPointer(),
			SpecifyUrlPattern:  compressionModel.SpecifyUrlPattern.ValueStringPointer(),
			PathPattern:        compressionModel.PathPattern.ValueStringPointer(),
			CustomPattern:      compressionModel.CustomPattern.ValueStringPointer(),
			Directory:          compressionModel.Directory.ValueStringPointer(),
			FileTypes:          utils.ToStrSlice(compressionModel.FileTypes),
			FileTypeOthers:     utils.ToStrSlice(compressionModel.FileTypeOthers),
			CustomFileTypes:    utils.ToStrSlice(compressionModel.CustomFileTypes),
			UriFileTypes:       utils.ToStrSlice(compressionModel.UriFileTypes),
		}

		updateCompressionSettingRequest := cdnetworksapi.UpdateCompressionConfigRequest{
			CompressionSetting: configs,
		}
		_, err := r.client.UpdateCompressionConfig(model.DomainId.ValueString(), updateCompressionSettingRequest)
		if err != nil {
			return err
		}
	}

	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}

func (setting *compressionSettingsModel) check() error {
	// check range
	rangeCount := 0
	if !setting.PathPattern.IsNull() && !setting.PathPattern.IsUnknown() {
		rangeCount++
	}
	if !setting.CustomPattern.IsNull() && setting.CustomPattern.ValueString() != "" {
		rangeCount++
	}
	if !setting.CustomFileTypes.IsNull() && !setting.CustomFileTypes.IsUnknown() {
		rangeCount++
	}
	if !setting.SpecifyUrlPattern.IsNull() && !setting.SpecifyUrlPattern.IsUnknown() {
		rangeCount++
	}
	if !setting.Directory.IsNull() && !setting.Directory.IsUnknown() {
		rangeCount++
	}
	if rangeCount == 0 {
		return errors.New("Pick one of the following items: URL matching patterns, directories, file types - customized file types, specified common types and specified URLs)!")
	} else if rangeCount > 1 {
		return errors.New("One and only one of the following items should have value at the same time: path-pattern, custom-pattern, custom-file-type, specify-url-pattern, directory.")
	}

	// check file type
	if !setting.FileTypes.IsNull() && !setting.FileTypes.IsUnknown() {
		var fileTypes []string
		diags := setting.FileTypes.ElementsAs(nil, &fileTypes, false)
		if diags.HasError() {
			return errors.New("Fail to convert FileTypes to string slice")
		}
		if err := utils.CheckFileTypes(strings.Join(fileTypes, utils.Separator)); err != nil {
			return err
		}
	}

	if !setting.CustomFileTypes.IsNull() && !setting.CustomFileTypes.IsUnknown() {
		var customFileTypes []string
		diags := setting.CustomFileTypes.ElementsAs(nil, &customFileTypes, false)
		if diags.HasError() {
			return errors.New("Fail to convert CustomFileTypes to string slice")
		}
		if err := utils.CheckFileTypes(strings.Join(customFileTypes, utils.Separator)); err != nil {
			return err
		}
	}

	return nil
}
