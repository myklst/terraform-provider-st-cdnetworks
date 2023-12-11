package cdnetworks

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type queryStringSettingModel struct {
	PathPattern        types.String `tfsdk:"path_pattern"`
	CustomPattern      types.String `tfsdk:"custom_pattern"`
	SpecifyUrlPattern  types.String `tfsdk:"specify_url_pattern"`
	FileType           types.String `tfsdk:"file_type"`
	CustomFileType     types.String `tfsdk:"custom_file_type"`
	Directory          types.String `tfsdk:"directory"`
	IgnoreLetterCase   types.Bool   `tfsdk:"ignore_letter_case"`
	IgnoreQueryString  types.Bool   `tfsdk:"ignore_query_string"`
	QueryStringKept    types.String `tfsdk:"query_string_kept"`
	QueryStringRemoved types.String `tfsdk:"query_string_removed"`
	SourceWithQuery    types.Bool   `tfsdk:"source_with_query"`
	SourceKeyKept      types.String `tfsdk:"source_key_kept"`
	SourceKeyRemoved   types.String `tfsdk:"source_key_removed"`
}

type queryStringUrlConfigModel struct {
	DomainId            types.String               `tfsdk:"domain_id"`
	QueryStringSettings []*queryStringSettingModel `tfsdk:"query_string_setting"`
}

type queryStringUrlConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &queryStringUrlConfigResource{}
	_ resource.ResourceWithConfigure   = &queryStringUrlConfigResource{}
	_ resource.ResourceWithModifyPlan  = &queryStringUrlConfigResource{}
	_ resource.ResourceWithImportState = &queryStringUrlConfigResource{}
)

func NewQueryStringUrlConfigResource() resource.Resource {
	return &queryStringUrlConfigResource{}
}

func (r *queryStringUrlConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_query_string_url_config"
}

func (r *queryStringUrlConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `With the query string URL, you can set whether to cache multiple copies or cache the URL after removing the question mark (to increase the hit rate), and you can set whether to use the original request to return to the source, etc.`,
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"query_string_setting": &schema.ListNestedBlock{
				Description: `Query String Settings Configuration`,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path_pattern": &schema.StringAttribute{
							Description: "The url matching mode supports regularization. If all matches, the input parameters can be configured as: .*",
							Optional:    true,
						},
						"custom_pattern": &schema.StringAttribute{
							Description: `Specify common types: Select the domain name that requires the anti-theft chain to be all files or the home page. :
                                E.g:
                                All: all files
                                Homepage: homepage`,
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"file_type": &schema.StringAttribute{
							Description: `
                            File Type: Specify the file type for anti-theft chain settings.
    File types include: gif png bmp jpeg jpg html htm shtml mp3 wma flv mp4 wmv zip exe rar css txt ico js swf
    If you need all types, pass all directly. Multiples are separated by semicolons, and all and specific file types cannot be configured at the same time.
                            `,
							Optional: true,
						},
						"custom_file_type": &schema.StringAttribute{
							Description: `Custom file type: Fill in the appropriate identifiable file type according to your needs outside of the specified file type. Can be used with file-type. If the file-type is also configured, the actual file type is the sum of the two parameters.`,
							Optional:    true,
						},
						"specify_url_pattern": &schema.StringAttribute{
							Description: `Specify URL cache: Specify url according to requirements for anti-theft chain setting
    INS format does not support URI format with http(s)://`,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`[^(http(s)?://)].*`), "The input parameter does not support the URI format starting with http(s)://"),
							},
						},
						"directory": &schema.StringAttribute{
							Description: "Directory: Specify the directory for anti-theft chain settings.Enter a legal directory format. Multiple separated by semicolons",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^/.*/$`), `Must start and end with "/"`),
							},
						},
						"ignore_letter_case": &schema.BoolAttribute{
							Description: `Whether to ignore letter case.`,
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"ignore_query_string": &schema.BoolAttribute{
							Description: `Define the file types to be compressed. 'text/' will be compressed by default.`,
							Optional:    true,
						},
						"query_string_kept": &schema.StringAttribute{
							Description: `Cache with the specified query string parameters. If the kept parameter values are the same, one copy will be cached.
Note:
1. query-string-kept and query-string-removed are mutually exclusive, and only one of them has a value.
2. query-string-kept and ignore-query-string are mutually exclusive, and only one has a value.`,
							Optional: true,
						},
						"query_string_removed": &schema.StringAttribute{
							Description: `Cache without the specified query string parameters. After deleting the specified parameter, if the other parameter values are the same, one copy will be cached.
1. query-string-kept and query string removed are mutually exclusive, and only one has a value.
2. query-string-removed and ignore-query-string are mutually exclusive.`,
							Optional: true,
						},
						"source_with_query": &schema.BoolAttribute{
							Description: `Whether to use the original URL back source, the allowable values are true and false.
When ignore-query-string is true or not set, source-with-query is true to indicate that the source is returned according to the original request, and false to indicate that the question mark is returned.
When ignore-query-string is false, this default setting is empty (input is invalid).`,
							Optional: true,
						},
						"source_key_kept": &schema.StringAttribute{
							Description: `Return to the source after specifying the reserved parameter value. Please separate them with semicolons, if no parameters reserved, please fill in:- . 1. Source-key-kept and ignore-query-string are mutually exclusive, and only one of them has a value. 2. Source-key-kept and source-key-removed are mutually exclusive, and only one of them has a value.`,
							Optional:    true,
						},
						"source_key_removed": &schema.StringAttribute{
							Description: `Return to the source after specifying the deleted parameter value. Please separate them with semicolons, and if you do not delete any parameters, please fill in:- . 1. Source-key-removed and ignore-query-string are mutually exclusive, and only one of them has a value. 2. Source-key-kept and source-key-removed are mutually exclusive, and only one of them has a value.`,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *queryStringUrlConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *queryStringUrlConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *queryStringUrlConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to set query_string_url_config", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *queryStringUrlConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *queryStringUrlConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateModel(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to read_query_string_url_config", err.Error())
		return
	}

	resp.State.Set(ctx, model)
}

func (r *queryStringUrlConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *queryStringUrlConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to set query_string_url_config", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *queryStringUrlConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *queryStringUrlConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.QueryStringSettings = make([]*queryStringSettingModel, 0)
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete query_string_url_config", err.Error())
	}
}

func (r *queryStringUrlConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *queryStringUrlConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan == nil {
		return
	}

	if plan.QueryStringSettings != nil {
		for _, setting := range plan.QueryStringSettings {
			err := setting.check()
			if err != nil {
				resp.Diagnostics.AddError("[Validate Config]Invalid Config", err.Error())
				return
			}
		}
	}
	resp.Plan.Set(ctx, plan)
}

func (r *queryStringUrlConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *queryStringUrlConfigResource) updateConfig(model *queryStringUrlConfigModel) error {
	settings := make([]*cdnetworksapi.QueryStringSetting, 0)
	if model.QueryStringSettings != nil {
		for _, settingModel := range model.QueryStringSettings {
			setting := &cdnetworksapi.QueryStringSetting{
				PathPattern:        settingModel.PathPattern.ValueStringPointer(),
				CustomPattern:      settingModel.CustomPattern.ValueStringPointer(),
				FileTypes:          settingModel.FileType.ValueStringPointer(),
				CustomFileTypes:    settingModel.CustomFileType.ValueStringPointer(),
				SpecifyUrlPattern:  settingModel.SpecifyUrlPattern.ValueStringPointer(),
				Directories:        settingModel.Directory.ValueStringPointer(),
				IgnoreLetterCase:   settingModel.IgnoreLetterCase.ValueBoolPointer(),
				QueryStringKept:    settingModel.QueryStringKept.ValueStringPointer(),
				QueryStringRemoved: settingModel.QueryStringRemoved.ValueStringPointer(),
				SourceWithQuery:    settingModel.SourceWithQuery.ValueBoolPointer(),
				SourceKeyKept:      settingModel.SourceKeyKept.ValueStringPointer(),
				SourceKeyRemoved:   settingModel.SourceKeyRemoved.ValueStringPointer(),
				IgnoreQueryString:  settingModel.IgnoreQueryString.ValueBoolPointer(),
			}
			settings = append(settings, setting)
		}
	}
	updateQueryStringConfigRequest := cdnetworksapi.UpdateQueryStringConfigRequest{
		QueryStringSettings: settings,
	}
	_, err := r.client.UpdateQueryStringConfig(model.DomainId.ValueString(), updateQueryStringConfigRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}

func (r *queryStringUrlConfigResource) updateModel(model *queryStringUrlConfigModel) error {
	settingIndexMap := make(map[string][]int)
	for i, setting := range model.QueryStringSettings {
		list, ok := settingIndexMap[setting.String()]
		if !ok {
			list = make([]int, 0)
		}
		settingIndexMap[setting.String()] = append(list, i)
	}

	queryQueryStringConfigResponse, err := r.client.QueryQueryStringConfig(model.DomainId.ValueString())
	if err != nil {
		return err
	}
	model.QueryStringSettings = make([]*queryStringSettingModel, 0)
	if queryQueryStringConfigResponse.QueryStringSetting != nil {
		for _, setting := range queryQueryStringConfigResponse.QueryStringSetting {
			settingModel := &queryStringSettingModel{
				PathPattern:        types.StringPointerValue(setting.PathPattern),
				CustomPattern:      types.StringPointerValue(setting.CustomPattern),
				FileType:           types.StringPointerValue(setting.FileTypes),
				CustomFileType:     types.StringPointerValue(setting.CustomFileTypes),
				SpecifyUrlPattern:  types.StringPointerValue(setting.SpecifyUrlPattern),
				Directory:          types.StringPointerValue(setting.Directories),
				IgnoreLetterCase:   types.BoolPointerValue(setting.IgnoreLetterCase),
				QueryStringKept:    types.StringPointerValue(setting.QueryStringKept),
				QueryStringRemoved: types.StringPointerValue(setting.QueryStringRemoved),
				SourceWithQuery:    types.BoolPointerValue(setting.SourceWithQuery),
				SourceKeyKept:      types.StringPointerValue(setting.SourceKeyKept),
				SourceKeyRemoved:   types.StringPointerValue(setting.SourceKeyRemoved),
			}
			if (setting.QueryStringKept != nil || setting.QueryStringRemoved != nil || setting.SourceWithQuery != nil || setting.SourceKeyRemoved != nil) && (setting.IgnoreQueryString != nil && *setting.IgnoreQueryString == false) {
				settingModel.IgnoreQueryString = types.BoolNull()
			} else {
				settingModel.IgnoreQueryString = types.BoolPointerValue(setting.IgnoreQueryString)
			}
			model.QueryStringSettings = append(model.QueryStringSettings, settingModel)
		}
	}

	// Sort settings to align with the configuaration file
	queryStringSettings := make([]*queryStringSettingModel, len(model.QueryStringSettings))
	unmatchedSettings := make([]*queryStringSettingModel, 0)
	for i := 0; i < len(model.QueryStringSettings); i++ {
		setting := model.QueryStringSettings[i]
		list, ok := settingIndexMap[setting.String()]
		if ok && len(list) > 0 {
			queryStringSettings[list[0]] = setting
			settingIndexMap[setting.String()] = list[1:]
		} else {
			unmatchedSettings = append(unmatchedSettings, setting)
		}
	}
	for i, j := 0, 0; i < len(queryStringSettings) && j < len(unmatchedSettings); i++ {
		if queryStringSettings[i] != nil {
			continue
		}
		queryStringSettings[i] = unmatchedSettings[j]
		j++
	}
	model.QueryStringSettings = queryStringSettings

	return nil
}

func (setting *queryStringSettingModel) check() error {
	// check range
	rangeCount := 0
	if !setting.PathPattern.IsNull() && !setting.PathPattern.IsUnknown() {
		rangeCount++
	}
	if !setting.CustomPattern.IsNull() && setting.CustomPattern.ValueString() != "" {
		rangeCount++
	}
	if (!setting.FileType.IsNull() && !setting.FileType.IsUnknown()) || (!setting.CustomFileType.IsNull() && !setting.CustomFileType.IsUnknown()) {
		rangeCount++
	}
	if !setting.Directory.IsNull() && !setting.Directory.IsUnknown() {
		rangeCount++
	}
	if !setting.SpecifyUrlPattern.IsNull() && !setting.SpecifyUrlPattern.IsUnknown() {
		rangeCount++
	}
	if rangeCount != 1 {
		return errors.New("One and only one of the following items should have value at the same time: path-pattern, directory, (file-type | custom-file-type), custom-pattern, specify-url-pattern.")
	}
	// check file type
	if !setting.FileType.IsNull() && !setting.FileType.IsUnknown() {
		err := utils.CheckFileTypes(setting.FileType.ValueString())
		if err != nil {
			return err
		}
	}
	if (!setting.QueryStringKept.IsNull() && !setting.QueryStringKept.IsUnknown()) &&
		(!setting.QueryStringRemoved.IsNull() && !setting.QueryStringRemoved.IsUnknown()) {
		return errors.New("query_string_kept and query_string_removed are mutually exclusive, and only one of them has a value.")
	}

	if (!setting.IgnoreQueryString.IsNull() && !setting.IgnoreQueryString.IsUnknown()) &&
		(!setting.QueryStringKept.IsNull() && !setting.QueryStringKept.IsUnknown()) {
		return errors.New("query_string_kept and ignore_query_string are mutually exclusive, and only one has a value.")
	}
	if (!setting.IgnoreQueryString.IsNull() && !setting.IgnoreQueryString.IsUnknown()) &&
		(!setting.QueryStringRemoved.IsNull() && !setting.QueryStringRemoved.IsUnknown()) {
		return errors.New("query_string_removed and ignore_query_string are mutually exclusive, and only one has a value.")
	}
	if (!setting.SourceKeyKept.IsNull() && !setting.SourceKeyKept.IsUnknown()) &&
		(!setting.SourceKeyRemoved.IsNull() && !setting.SourceKeyRemoved.IsUnknown()) {
		return errors.New("source_key_kept and source_key_removed are mutually exclusive, and only one of them has a value.")
	}
	if (!setting.IgnoreQueryString.IsNull() && !setting.IgnoreQueryString.IsUnknown()) &&
		(!setting.SourceKeyKept.IsNull() && !setting.SourceKeyKept.IsUnknown()) {
		return errors.New("source_key_kept and ignore_query_string are mutually exclusive, and only one of them has a value.")
	}
	if (!setting.IgnoreQueryString.IsNull() && !setting.IgnoreQueryString.IsUnknown()) &&
		(!setting.SourceKeyRemoved.IsNull() && !setting.SourceKeyRemoved.IsUnknown()) {
		return errors.New("source_key_kept and ignore_query_string are mutually exclusive, and only one of them has a value.")
	}
	if (!setting.IgnoreQueryString.IsNull() && !setting.IgnoreQueryString.IsUnknown()) &&
		(!setting.SourceWithQuery.IsNull()) && !setting.SourceWithQuery.IsUnknown() {
		return errors.New("When ignore_query_string is false, source_with_query should be empty.")
	}

	if setting.IgnoreQueryString.IsNull() && setting.QueryStringKept.IsNull() && setting.QueryStringRemoved.IsNull() {
		return errors.New("One of the following field should have a value:ignore_query_string,query_string_kept,query_string_removed.")
	}

	return nil
}

func (setting *queryStringSettingModel) String() string {
	values := []string{
		setting.PathPattern.String(),
		setting.CustomPattern.String(),
		setting.SpecifyUrlPattern.String(),
		setting.FileType.String(),
		setting.CustomFileType.String(),
		setting.Directory.String(),
		setting.IgnoreLetterCase.String(),
		setting.IgnoreQueryString.String(),
		setting.QueryStringKept.String(),
		setting.QueryStringRemoved.String(),
		setting.SourceKeyKept.String(),
		setting.SourceKeyRemoved.String(),
		setting.SourceKeyKept.String(),
	}
	return strings.Join(values, "$$")
}
