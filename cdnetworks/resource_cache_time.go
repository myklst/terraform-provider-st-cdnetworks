package cdnetworks

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type cacheTimeBehaviorModel struct {
	PathPattern                types.String `tfsdk:"path_pattern"`
	ExceptPathPattern          types.String `tfsdk:"except_path_pattern"`
	CustomPattern              types.String `tfsdk:"custom_pattern"`
	SpecifyUrlPattern          types.String `tfsdk:"specify_url_pattern"`
	FileType                   types.String `tfsdk:"file_type"`
	CustomFileType             types.String `tfsdk:"custom_file_type"`
	Directory                  types.String `tfsdk:"directory"`
	CacheTtl                   types.Int64  `tfsdk:"cache_ttl"`
	IgnoreCacheControl         types.Bool   `tfsdk:"ignore_cache_control"`
	IsRespectServer            types.Bool   `tfsdk:"is_respect_server"`
	IgnoreLetterCase           types.Bool   `tfsdk:"ignore_letter_case"`
	ReloadManage               types.String `tfsdk:"reload_manage"`
	IgnoreAuthenticationHeader types.Bool   `tfsdk:"ignore_authentication_header"`
	Priority                   types.Int64  `tfsdk:"priority"`
}

type cacheTimeModel struct {
	DomainId           types.String              `tfsdk:"domain_id"`
	CacheTimeBehaviors []*cacheTimeBehaviorModel `tfsdk:"cache_time_behavior"`
}

type cacheTimeResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &cacheTimeResource{}
	_ resource.ResourceWithConfigure   = &cacheTimeResource{}
	_ resource.ResourceWithModifyPlan  = &cacheTimeResource{}
	_ resource.ResourceWithImportState = &cacheTimeResource{}
)

func NewCacheTimeResource() resource.Resource {
	return &cacheTimeResource{}
}

func (r *cacheTimeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache_time"
}

func (r *cacheTimeResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `This resource implementation modifies the domain name cache time configuration, realizes the custom cache function according to the customer's request. Node cache is divided into regular cache and query string URl cache, where you can set the cache time and ignore certain headers that affect the cache, and whether to cache empty files, etc., can the query string Url be set to multiple or to cache the Url after removing the question mark (increasing hit rate)`,
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"cache_time_behavior": &schema.ListNestedBlock{
				Description: `Cache time configuration`,
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path_pattern": &schema.StringAttribute{
							Description: "The url matching mode supports regularization. If all matches, the input parameters can be configured as: .*",
							Optional:    true,
						},
						"except_path_pattern": &schema.StringAttribute{
							Description: `Exceptional url matching mode, except for some URLs: such as abc.jpg, do not do anti-theft chain function.E.g: ^https?://[^/]+/.*\.m3u8`,
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
						"cache_ttl": &schema.Int64Attribute{
							Description: `Cache time: set the time corresponding to the cache object.Time unit is second, no cache is set to 0. This time is set according to the customer's own needs. If the customer feels that some of the files are not changed frequently, then the setting is longer. For example, the text class js, css, html, etc. can be set shorter, the picture, video and audio classes can be set longer (because the cache time will be replaced by the new file due to the file heat algorithm, the longest suggestion Do not exceed one month)`,
							Required:    true,
						},
						"ignore_cache_control": &schema.BoolAttribute{
							Description: `Ignore the source station does not cache the header. The optional values are true and false, which are used to ignore the two configurations of cache-control in the request header (private, no-cache) and the Authorization set by the client.
The true indicates that the source station's settings for the three are ignored. Enables resources to be cached on the service node in the form of cache-control: public, and then our nodes can cache this type of resource and provide acceleration services.
False means that when the source station sets cache-control: private, cache-control: no-cache for a resource or specifies to cache according to authorization, our service node will not cache such files.`,
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"is_respect_server": &schema.BoolAttribute{
							Description: `Respect the server: Accelerate whether to prioritize the source cache time.
Optional values: true and false
True: indicates that the server is time-first
False: The cache time of the CDN configuration takes precedence.`,
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"ignore_letter_case": &schema.BoolAttribute{
							Description: `Ignore case, the optional value is true or false, true means to ignore case; false means not to ignore case;
When adding a new configuration item, the default is not true.`,
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"reload_manage": &schema.StringAttribute{
							Description: `Reload processing rules, optional: ignore or if-modified-since
If-modified-since: indicates that you want to convert to if-modified-since
Ignore: means to ignore client refresh`,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("ignore", "if-modified-since"),
							},
							Computed: true,
							Default:  stringdefault.StaticString("ignore"),
						},
						"ignore_authentication_header": &schema.BoolAttribute{
							Description: `You can set it 'true' to cache
ignoring the http header 'Authentication'. If it is empty, the header is not ignored by default.`,
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"priority": &schema.Int64Attribute{
							Description: `Indicates the priority execution order of multiple sets of redirected content by the customer. The higher the number, the higher the priority.
When adding a new configuration item, the default is 10`,
							Optional: true,
							Computed: true,
							Default:  int64default.StaticInt64(10),
						},
					},
				},
			},
		},
	}
}

func (r *cacheTimeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *cacheTimeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *cacheTimeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to set cache_time", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *cacheTimeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *cacheTimeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateModel(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to read_cache_time", err.Error())
		return
	}

	resp.State.Set(ctx, model)
}

func (r *cacheTimeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *cacheTimeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to set cache_time", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *cacheTimeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *cacheTimeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.CacheTimeBehaviors = make([]*cacheTimeBehaviorModel, 0)
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete cache_time", err.Error())
	}
}

func (r *cacheTimeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *cacheTimeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan == nil {
		return
	}

	if plan.CacheTimeBehaviors == nil || len(plan.CacheTimeBehaviors) == 0 {
		resp.Diagnostics.AddError("[Validate Config]Invalid Config", "cache_time_behavior is required")
		return
	}
	for _, behavior := range plan.CacheTimeBehaviors {
		err := behavior.check()
		if err != nil {
			resp.Diagnostics.AddError("[Validate Config]Invalid Config", err.Error())
			return
		}
	}
	resp.Plan.Set(ctx, plan)
}

func (r *cacheTimeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *cacheTimeResource) updateConfig(model *cacheTimeModel) error {
	behaviors := make([]*cdnetworksapi.CacheTimeBehavior, 0)
	if model.CacheTimeBehaviors != nil {
		for _, behaviorModel := range model.CacheTimeBehaviors {
			behavior := &cdnetworksapi.CacheTimeBehavior{
				PathPattern:                behaviorModel.PathPattern.ValueStringPointer(),
				ExceptPathPattern:          behaviorModel.ExceptPathPattern.ValueStringPointer(),
				CustomPattern:              behaviorModel.CustomPattern.ValueStringPointer(),
				FileType:                   behaviorModel.FileType.ValueStringPointer(),
				CustomFileType:             behaviorModel.CustomFileType.ValueStringPointer(),
				SpecifyUrlPattern:          behaviorModel.SpecifyUrlPattern.ValueStringPointer(),
				Directory:                  behaviorModel.Directory.ValueStringPointer(),
				IgnoreCacheControl:         behaviorModel.IgnoreCacheControl.ValueBoolPointer(),
				IsRespectServer:            behaviorModel.IsRespectServer.ValueBoolPointer(),
				IgnoreLetterCase:           behaviorModel.IgnoreLetterCase.ValueBoolPointer(),
				ReloadManage:               behaviorModel.ReloadManage.ValueStringPointer(),
				IgnoreAuthenticationHeader: behaviorModel.IgnoreAuthenticationHeader.ValueBoolPointer(),
				Priority:                   behaviorModel.Priority.ValueInt64Pointer(),
			}
			if behaviorModel.CacheTtl.IsNull() || behaviorModel.CacheTtl.IsUnknown() {
				behavior.CacheTtl = nil
			} else {
				ttl := fmt.Sprintf("%ds", behaviorModel.CacheTtl.ValueInt64())
				behavior.CacheTtl = &ttl
			}
			behaviors = append(behaviors, behavior)
		}
	}
	updateCacheTimeConfigRequest := cdnetworksapi.UpdateCacheTimeConfigRequest{
		CacheTimeBehaviors: behaviors,
	}
	_, err := r.client.UpdateCacheTimeConfig(model.DomainId.ValueString(), updateCacheTimeConfigRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}

func (r *cacheTimeResource) updateModel(model *cacheTimeModel) error {
	behaviorIndexMap := make(map[string][]int)
	for i, behavior := range model.CacheTimeBehaviors {
		list, ok := behaviorIndexMap[behavior.String()]
		if !ok {
			list = make([]int, 0)
		}
		behaviorIndexMap[behavior.String()] = append(list, i)
	}

	queryCacheTimeConfigResponse, err := r.client.QueryCacheTimeConfig(model.DomainId.ValueString())
	if err != nil {
		return err
	}
	model.CacheTimeBehaviors = make([]*cacheTimeBehaviorModel, 0)
	if queryCacheTimeConfigResponse.CacheTimeBehaviors != nil {
		for _, behavior := range queryCacheTimeConfigResponse.CacheTimeBehaviors {
			behaviorModel := &cacheTimeBehaviorModel{
				PathPattern:                types.StringPointerValue(behavior.PathPattern),
				ExceptPathPattern:          types.StringPointerValue(behavior.ExceptPathPattern),
				CustomPattern:              types.StringPointerValue(behavior.CustomPattern),
				FileType:                   types.StringPointerValue(behavior.FileType),
				CustomFileType:             types.StringPointerValue(behavior.CustomFileType),
				SpecifyUrlPattern:          types.StringPointerValue(behavior.SpecifyUrlPattern),
				Directory:                  types.StringPointerValue(behavior.Directory),
				IgnoreCacheControl:         types.BoolPointerValue(behavior.IgnoreLetterCase),
				IsRespectServer:            types.BoolPointerValue(behavior.IsRespectServer),
				IgnoreLetterCase:           types.BoolPointerValue(behavior.IgnoreLetterCase),
				ReloadManage:               types.StringPointerValue(behavior.ReloadManage),
				IgnoreAuthenticationHeader: types.BoolPointerValue(behavior.IgnoreAuthenticationHeader),
				Priority:                   types.Int64PointerValue(behavior.Priority),
			}
			if behavior.CacheTtl == nil {
				behaviorModel.CacheTtl = types.Int64Null()
			} else {
				ttl, err := convertTtlToSecond(*behavior.CacheTtl)
				if err != nil {
					return err
				}
				behaviorModel.CacheTtl = types.Int64Value(ttl)
			}
			model.CacheTimeBehaviors = append(model.CacheTimeBehaviors, behaviorModel)
		}
	}

	// Sort cache_time_behavior to align with rules in .tf configuration file
	cacheTimeBehaviors := make([]*cacheTimeBehaviorModel, len(model.CacheTimeBehaviors))
	unmatchedBehaviors := make([]*cacheTimeBehaviorModel, 0)
	for i := 0; i < len(model.CacheTimeBehaviors); i++ {
		behavior := model.CacheTimeBehaviors[i]
		list, ok := behaviorIndexMap[behavior.String()]
		if ok && len(list) > 0 {
			cacheTimeBehaviors[list[0]] = behavior
			behaviorIndexMap[behavior.String()] = list[1:]
		} else {
			unmatchedBehaviors = append(unmatchedBehaviors, behavior)
		}
	}
	for i, j := 0, 0; i < len(cacheTimeBehaviors) && j < len(unmatchedBehaviors); i++ {
		if cacheTimeBehaviors[i] != nil {
			continue
		}
		cacheTimeBehaviors[i] = unmatchedBehaviors[j]
		j++
	}
	model.CacheTimeBehaviors = cacheTimeBehaviors

	return nil
}

func (behavior *cacheTimeBehaviorModel) check() error {
	// check range
	rangeCount := 0
	if !behavior.PathPattern.IsNull() && !behavior.PathPattern.IsUnknown() {
		rangeCount++
	}
	if !behavior.CustomPattern.IsNull() && behavior.CustomPattern.ValueString() != "" {
		rangeCount++
	}
	if (!behavior.FileType.IsNull() && !behavior.FileType.IsUnknown()) || (!behavior.CustomFileType.IsNull() && !behavior.CustomFileType.IsUnknown()) {
		rangeCount++
	}
	if !behavior.Directory.IsNull() && !behavior.Directory.IsUnknown() {
		rangeCount++
	}
	if !behavior.SpecifyUrlPattern.IsNull() && !behavior.SpecifyUrlPattern.IsUnknown() {
		rangeCount++
	}
	if rangeCount != 1 {
		return errors.New("One and only one of the following items should have value at the same time: path-pattern, directory, (file-type | custom-file-type), custom-pattern, specify-url-pattern.")
	}
	// check file type
	if !behavior.FileType.IsNull() && !behavior.FileType.IsUnknown() {
		err := utils.CheckFileTypes(behavior.FileType.ValueString())
		if err != nil {
			return err
		}
	}
	return nil
}

func (behavior *cacheTimeBehaviorModel) String() string {
	values := []string{
		behavior.PathPattern.String(),
		behavior.SpecifyUrlPattern.String(),
		behavior.CustomPattern.String(),
		behavior.ExceptPathPattern.String(),
		behavior.FileType.String(),
		behavior.CustomPattern.String(),
		behavior.Directory.String(),
		behavior.CacheTtl.String(),
		behavior.IgnoreCacheControl.String(),
		behavior.IgnoreAuthenticationHeader.String(),
		behavior.IsRespectServer.String(),
		behavior.IgnoreLetterCase.String(),
		behavior.ReloadManage.String(),
		behavior.Priority.String(),
	}
	return strings.Join(values, "$$")
}

func convertTtlToSecond(ttl string) (int64, error) {
	matched, err := regexp.MatchString(`^(\d+[shdm]|\d+)$`, ttl)
	if err != nil || !matched {
		return 0, fmt.Errorf("Invalid ttl %s", ttl)
	}
	sz := len(ttl)
	multiplier := 1

	var n int64
	if ttl[sz-1] == 'm' {
		n, _ = strconv.ParseInt(ttl[0:len(ttl)-1], 10, 64)
		multiplier = 3600 * 24 * 30
	} else if ttl[sz-1] == 'd' {
		n, _ = strconv.ParseInt(ttl[0:len(ttl)-1], 10, 64)
		multiplier = 3600 * 24
	} else if ttl[sz-1] == 'h' {
		n, _ = strconv.ParseInt(ttl[0:len(ttl)-1], 10, 64)
		multiplier = 3600
	} else if ttl[sz-1] == 's' {
		n, _ = strconv.ParseInt(ttl[0:len(ttl)-1], 10, 64)
	} else {
		n, _ = strconv.ParseInt(ttl, 10, 64)
	}
	return n * int64(multiplier), nil
}
