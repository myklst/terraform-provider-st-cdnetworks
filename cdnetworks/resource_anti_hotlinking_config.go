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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/utils"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

type ipControlRuleModel struct {
	PathPattern          types.String `tfsdk:"path_pattern"`
	ExceptPathPattern    types.String `tfsdk:"except_path_pattern"`
	CustomPattern        types.String `tfsdk:"custom_pattern"`
	FileType             types.String `tfsdk:"file_type"`
	ExceptFileType       types.String `tfsdk:"except_file_type"`
	CustomFileType       types.String `tfsdk:"custom_file_type"`
	ExceptCustomFileType types.String `tfsdk:"except_custom_file_type"`
	SpecifyUrlPattern    types.String `tfsdk:"specify_url_pattern"`
	Directory            types.String `tfsdk:"directory"`
	ExceptDirectory      types.String `tfsdk:"except_directory"`
	ControlAction        types.String `tfsdk:"control_action"`
	RewriteTo            types.String `tfsdk:"rewrite_to"`
	Priority             types.Int64  `tfsdk:"priority"`

	AllowedIps   types.List `tfsdk:"allowed_ips"`
	ForbiddenIps types.List `tfsdk:"forbidden_ips"`
}

type refererControlRuleModel struct {
	PathPattern          types.String `tfsdk:"path_pattern"`
	ExceptPathPattern    types.String `tfsdk:"except_path_pattern"`
	CustomPattern        types.String `tfsdk:"custom_pattern"`
	FileType             types.String `tfsdk:"file_type"`
	ExceptFileType       types.String `tfsdk:"except_file_type"`
	CustomFileType       types.String `tfsdk:"custom_file_type"`
	ExceptCustomFileType types.String `tfsdk:"except_custom_file_type"`
	SpecifyUrlPattern    types.String `tfsdk:"specify_url_pattern"`
	Directory            types.String `tfsdk:"directory"`
	ExceptDirectory      types.String `tfsdk:"except_directory"`
	ControlAction        types.String `tfsdk:"control_action"`
	RewriteTo            types.String `tfsdk:"rewrite_to"`
	Priority             types.Int64  `tfsdk:"priority"`

	AllowNullReferer types.Bool `tfsdk:"allow_null_referer"`
	ValidUrls        types.List `tfsdk:"valid_urls"`
	ValidDomains     types.List `tfsdk:"valid_domains"`
	InvalidUrls      types.List `tfsdk:"invalid_urls"`
	InvalidDomains   types.List `tfsdk:"invalid_domains"`
}

type uaControlRuleModel struct {
	PathPattern          types.String `tfsdk:"path_pattern"`
	ExceptPathPattern    types.String `tfsdk:"except_path_pattern"`
	CustomPattern        types.String `tfsdk:"custom_pattern"`
	FileType             types.String `tfsdk:"file_type"`
	ExceptFileType       types.String `tfsdk:"except_file_type"`
	CustomFileType       types.String `tfsdk:"custom_file_type"`
	ExceptCustomFileType types.String `tfsdk:"except_custom_file_type"`
	SpecifyUrlPattern    types.String `tfsdk:"specify_url_pattern"`
	Directory            types.String `tfsdk:"directory"`
	ExceptDirectory      types.String `tfsdk:"except_directory"`
	ControlAction        types.String `tfsdk:"control_action"`
	RewriteTo            types.String `tfsdk:"rewrite_to"`
	Priority             types.Int64  `tfsdk:"priority"`

	ValidUserAgents   types.List `tfsdk:"valid_user_agents"`
	InvalidUserAgents types.List `tfsdk:"invalid_user_agents"`
}

type antiHotlinkingConfigModel struct {
	DomainId            types.String               `tfsdk:"domain_id"`
	IpControlRules      []*ipControlRuleModel      `tfsdk:"ip_control_rule"`
	RefererControlRules []*refererControlRuleModel `tfsdk:"referer_control_rule"`
	UaControlRules      []*uaControlRuleModel      `tfsdk:"ua_control_rule"`
}

type antiHotlinkingConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &antiHotlinkingConfigResource{}
	_ resource.ResourceWithConfigure   = &antiHotlinkingConfigResource{}
	_ resource.ResourceWithModifyPlan  = &antiHotlinkingConfigResource{}
	_ resource.ResourceWithImportState = &antiHotlinkingConfigResource{}
)

func NewAntiHotlinkingConfigResource() resource.Resource {
	return &antiHotlinkingConfigResource{}
}

func (r *antiHotlinkingConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_anti_hotlinking_config"
}

func (r *antiHotlinkingConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Anti-theft chain configuration
                    note:
                    1. When you need to cancel the anti-theft chain configuration settings, you can pass in the empty node .
                    2. When it is necessary to set the anti-theft chain configuration, this item is required.`,
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Description: "Domain id",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"ip_control_rule": &schema.ListNestedBlock{
				Description: `Identify IP black and white list anti-theft chain
            note:
            1. a set of black and white list anti-theft chain, only one set under a data-id
            2. When the air interface label indicates the exception of the IP segment configuration and the forbidden IP segment configuration.`,
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
							Description: `File Type: Specify the file type for anti-theft chain settings.
    File types include: gif png bmp jpeg jpg html htm shtml mp3 wma flv mp4 wmv zip exe rar css txt ico js swf
    If you need all types, pass all directly. Multiples are separated by semicolons, and all and specific file types cannot be configured at the same time.`,
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
						"except_file_type": &schema.StringAttribute{
							Description: `Exception file type: Specify the file type that does not require anti-theft chain function
        File types include: gif png bmp jpeg jpg html htm shtml mp3 wma flv mp4 wmv zip exe rar css txt ico js swf
        If you need all types, pass all directly. Multiple separated by semicolons, all and specific file types cannot be configured at the same time
        If file-type=all, except-file-type=all means that the task file type is not matched.`,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`[^(http(s)?://)].*`), "The input parameter does not support the URI format starting with http(s)://"),
							},
						},
						"except_custom_file_type": &schema.StringAttribute{
							Description: `Exceptional custom file types: Fill in the appropriate identifiable file types based on your needs, outside of the specified file type. Can be used with the exception-file-type. If the except-file-type is also configured, the actual file type is the sum of the two parameters.`,
							Optional:    true,
						},
						"except_directory": &schema.StringAttribute{
							Description: `Exceptional directory: Specify a directory that does not require anti-theft chain settings
    Enter a legal directory format. Multiple separated by semicolons`,
							Optional: true,
						},
						"control_action": &schema.StringAttribute{
							Description: `control direction. Available values: 403 and 302
    1) 403 means to return a specific error status code to reject the service (the default mode, the status code can be specified, generally 403).
    2) 302 means to return 302 the redirect url of the Found, the redirected url can be specified. If pass 302, rewrite-to is required`,
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("403"),
							Validators: []validator.String{
								stringvalidator.OneOf("403", "302"),
							},
						},
						"rewrite_to": &schema.StringAttribute{
							Description: `Specify the url after the 302 jump. This field is required if the control-action value is 302.`,
							Optional:    true,
						},
						"priority": &schema.Int64Attribute{
							Description: `Indicates the priority execution order of multiple sets of redirected content by the customer. The higher the number, the higher the priority.
When adding a new configuration item, the default is 10`,
							Optional: true,
							Computed: true,
							Default:  int64default.StaticInt64(10),
						},
						"forbidden_ips": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `Prohibited IP segment.`,
							Optional:    true,
							Computed:    true,
						},
						"allowed_ips": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `The exception IP segment supports input IP or IP segment, and the IP segments are separated by a semicolon (;), such as 1.1.1.0/24; 2.2.2.2, some IP exceptions, no anti-theft chain.`,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"referer_control_rule": &schema.ListNestedBlock{
				Description: `Identify referer anti-theft chain`,
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
						"except_file_type": &schema.StringAttribute{
							Description: `Exception file type: Specify the file type that does not require anti-theft chain function
        File types include: gif png bmp jpeg jpg html htm shtml mp3 wma flv mp4 wmv zip exe rar css txt ico js swf
        If you need all types, pass all directly. Multiple separated by semicolons, all and specific file types cannot be configured at the same time
        If file-type=all, except-file-type=all means that the task file type is not matched.`,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`[^(http(s)?://)].*`), "The input parameter does not support the URI format starting with http(s)://"),
							},
						},
						"except_custom_file_type": &schema.StringAttribute{
							Description: `Exceptional custom file types: Fill in the appropriate identifiable file types based on your needs, outside of the specified file type. Can be used with the exception-file-type. If the except-file-type is also configured, the actual file type is the sum of the two parameters.`,
							Optional:    true,
						},
						"except_directory": &schema.StringAttribute{
							Description: `Exceptional directory: Specify a directory that does not require anti-theft chain settings
    Enter a legal directory format. Multiple separated by semicolons`,
							Optional: true,
						},
						"control_action": &schema.StringAttribute{
							Description: `control direction. Available values: 403 and 302
    1) 403 means to return a specific error status code to reject the service (the default mode, the status code can be specified, generally 403).
    2) 302 means to return 302 the redirect url of the Found, the redirected url can be specified. If pass 302, rewrite-to is required`,
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("403"),
							Validators: []validator.String{
								stringvalidator.OneOf("403", "302"),
							},
						},
						"rewrite_to": &schema.StringAttribute{
							Description: `Specify the url after the 302 jump. This field is required if the control-action value is 302.`,
							Optional:    true,
						},
						"priority": &schema.Int64Attribute{
							Description: `Indicates the priority execution order of multiple sets of redirected content by the customer. The higher the number, the higher the priority.
When adding a new configuration item, the default is 10`,
							Optional: true,
							Computed: true,
							Default:  int64default.StaticInt64(10),
						},
						"allow_null_referer": schema.BoolAttribute{
							Description: `If any of the four terms 'nullreferer: legal referer, (legal domain name, legal URL), illegal referer, (illegal domain name, illegal URL)' is allowed, then 'nullreferer' cannot be null.If the four terms 'legal refer', 'legal domain name, legal URL', 'illegal refer', 'illegal domain name, illegal URL' are all null values, then 'whether to allow a null referer' must be null`,
							Optional:    true,
						},
						"valid_urls": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `Legal url, enter the correct url format.`,
							Optional:    true,
							Computed:    true,
						},
						"valid_domains": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `Legal domain name.`,
							Optional:    true,
							Computed:    true,
						},
						"invalid_urls": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `Invalid url, enter the correct url format.`,
							Optional:    true,
							Computed:    true,
						},
						"invalid_domains": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `Illegal domain name`,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"ua_control_rule": &schema.ListNestedBlock{
				Description: `UA head protection against hotlinking,
                    Note:
                    1. Represents a group of UA head defense hotlinking
                    2. when empty label means clear UA head protection hotlinking`,
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
						"except_file_type": &schema.StringAttribute{
							Description: `Exception file type: Specify the file type that does not require anti-theft chain function
        File types include: gif png bmp jpeg jpg html htm shtml mp3 wma flv mp4 wmv zip exe rar css txt ico js swf
        If you need all types, pass all directly. Multiple separated by semicolons, all and specific file types cannot be configured at the same time
        If file-type=all, except-file-type=all means that the task file type is not matched.`,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`[^(http(s)?://)].*`), "The input parameter does not support the URI format starting with http(s)://"),
							},
						},
						"except_custom_file_type": &schema.StringAttribute{
							Description: `Exceptional custom file types: Fill in the appropriate identifiable file types based on your needs, outside of the specified file type. Can be used with the exception-file-type. If the except-file-type is also configured, the actual file type is the sum of the two parameters.`,
							Optional:    true,
						},
						"except_directory": &schema.StringAttribute{
							Description: `Exceptional directory: Specify a directory that does not require anti-theft chain settings
    Enter a legal directory format. Multiple separated by semicolons`,
							Optional: true,
						},
						"control_action": &schema.StringAttribute{
							Description: `control direction. Available values: 403 and 302
    1) 403 means to return a specific error status code to reject the service (the default mode, the status code can be specified, generally 403).
    2) 302 means to return 302 the redirect url of the Found, the redirected url can be specified. If pass 302, rewrite-to is required`,
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("403"),
							Validators: []validator.String{
								stringvalidator.OneOf("403", "302"),
							},
						},
						"rewrite_to": &schema.StringAttribute{
							Description: `Specify the url after the 302 jump. This field is required if the control-action value is 302.`,
							Optional:    true,
						},
						"priority": &schema.Int64Attribute{
							Description: `Indicates the priority execution order of multiple sets of redirected content by the customer. The higher the number, the higher the priority.
When adding a new configuration item, the default is 10`,
							Optional: true,
							Computed: true,
							Default:  int64default.StaticInt64(10),
						},
						"valid_user_agents": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `Allowed clients, regular matching, no spaces allowed, to configure multiple UA such as:
    Android|iPhone`,
							Optional: true,
							Computed: true,
						},
						"invalid_user_agents": schema.ListAttribute{
							ElementType: types.StringType,
							Description: `Forbidden client, regular match, no spaces allowed, configure multiple UA such as:
    Android|iPhone`,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *antiHotlinkingConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *antiHotlinkingConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *antiHotlinkingConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to update anti_hotlinking_config", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *antiHotlinkingConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *antiHotlinkingConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateModel(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to query anti_hotlinking_config", err.Error())
		return
	}

	resp.State.Set(ctx, &model)
}

func (r *antiHotlinkingConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *antiHotlinkingConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to update anti_hotlinking_config", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *antiHotlinkingConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *antiHotlinkingConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.IpControlRules = make([]*ipControlRuleModel, 0)
	model.RefererControlRules = make([]*refererControlRuleModel, 0)
	model.UaControlRules = make([]*uaControlRuleModel, 0)
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete anti_hotlinking_config", err.Error())
	}
}

func (r *antiHotlinkingConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *antiHotlinkingConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan == nil {
		return
	}

	if plan.IpControlRules == nil {
		plan.IpControlRules = make([]*ipControlRuleModel, 0)
	}
	if plan.RefererControlRules == nil {
		plan.RefererControlRules = make([]*refererControlRuleModel, 0)
	}
	if plan.UaControlRules == nil {
		plan.UaControlRules = make([]*uaControlRuleModel, 0)
	}

	whitelistCount := 0
	blacklistCount := 0
	for _, rule := range plan.IpControlRules {
		if rule == nil {
			continue
		}
		if rule.AllowedIps.IsNull() || rule.AllowedIps.IsUnknown() {
			rule.AllowedIps, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		if rule.ForbiddenIps.IsNull() || rule.ForbiddenIps.IsUnknown() {
			rule.ForbiddenIps, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		err := rule.check()
		if err != nil {
			resp.Diagnostics.AddError("[Validate Config]Invalid config", err.Error())
			return
		}
		if len(rule.AllowedIps.Elements()) > 0 {
			whitelistCount++
		}
		if len(rule.ForbiddenIps.Elements()) > 0 {
			blacklistCount++
		}
	}
	if whitelistCount > 1 {
		resp.Diagnostics.AddError("[Validate Config]Invalid config", "The configuration of whitelist or exception IP/IP Segment already exists. It is not allowed to add a new one.")
		return
	}
	if whitelistCount > 0 && blacklistCount > 0 {
		resp.Diagnostics.AddError("[Validate Config]Invalid config", "The confliction of Blacklist and Whitelist configuration may affect your domain service.")
		return
	}

	whitelistCount = 0
	blacklistCount = 0
	for _, rule := range plan.RefererControlRules {
		if rule == nil {
			continue
		}
		if rule.ValidDomains.IsNull() || rule.ValidDomains.IsUnknown() {
			rule.ValidDomains, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		if rule.InvalidDomains.IsNull() || rule.InvalidDomains.IsUnknown() {
			rule.InvalidDomains, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		if rule.ValidUrls.IsNull() || rule.ValidUrls.IsUnknown() {
			rule.ValidUrls, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		if rule.InvalidUrls.IsNull() || rule.InvalidUrls.IsUnknown() {
			rule.InvalidUrls, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		err := rule.check()
		if err != nil {
			resp.Diagnostics.AddError("[Validate Config]Invalid config", err.Error())
			return
		}
		if len(rule.ValidDomains.Elements()) > 0 || len(rule.ValidUrls.Elements()) > 0 {
			whitelistCount++
		}
		if len(rule.InvalidDomains.Elements()) > 0 || len(rule.InvalidUrls.Elements()) > 0 {
			blacklistCount++
		}
	}
	if whitelistCount > 1 {
		resp.Diagnostics.AddError("[Validate Config]Invalid config", "The configuration of allowed referer already exists. It is not allowed to add a new one.")
		return
	}
	if whitelistCount > 0 && blacklistCount > 0 {
		resp.Diagnostics.AddError("[Validate Config]Invalid config", "The confliction of Blacklist and Whitelist configuration may affect your domain service.")
		return
	}

	whitelistCount = 0
	blacklistCount = 0
	for _, rule := range plan.UaControlRules {
		if rule == nil {
			continue
		}
		if rule.ValidUserAgents.IsNull() || rule.ValidUserAgents.IsUnknown() {
			rule.ValidUserAgents, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		if rule.InvalidUserAgents.IsNull() || rule.InvalidUserAgents.IsUnknown() {
			rule.InvalidUserAgents, _ = types.ListValueFrom(nil, types.StringType, []string{})
		}
		err := rule.check()
		if err != nil {
			resp.Diagnostics.AddError("[Validate Config]Invalid config", err.Error())
			return
		}
		if len(rule.ValidUserAgents.Elements()) > 0 {
			whitelistCount++
		}
		if len(rule.InvalidUserAgents.Elements()) > 0 {
			blacklistCount++
		}
	}
	if whitelistCount > 1 {
		resp.Diagnostics.AddError("[Validate Config]Invalid config", "The configuration of allowed User-Agent already exists. It is not allowed to add a new one.")
		return
	}
	if whitelistCount > 0 && blacklistCount > 0 {
		resp.Diagnostics.AddError("[Validate Config]Invalid config", "The confliction of Blacklist and Whitelist configuration may affect your domain service.")
		return
	}

	resp.Plan.Set(ctx, plan)
}

func (r *antiHotlinkingConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *antiHotlinkingConfigResource) updateConfig(model *antiHotlinkingConfigModel) error {
	rules := make([]*cdnetworksapi.VisitControlRule, 0)
	if model.IpControlRules != nil {
		for _, ruleModel := range model.IpControlRules {
			rule := &cdnetworksapi.VisitControlRule{
				PathPattern:          ruleModel.PathPattern.ValueStringPointer(),
				ExceptPathPattern:    ruleModel.ExceptPathPattern.ValueStringPointer(),
				CustomPattern:        ruleModel.CustomPattern.ValueStringPointer(),
				FileType:             ruleModel.FileType.ValueStringPointer(),
				CustomFileType:       ruleModel.CustomFileType.ValueStringPointer(),
				SpecifyUrlPattern:    ruleModel.SpecifyUrlPattern.ValueStringPointer(),
				Directory:            ruleModel.Directory.ValueStringPointer(),
				ExceptFileType:       ruleModel.ExceptFileType.ValueStringPointer(),
				ExceptCustomFileType: ruleModel.ExceptCustomFileType.ValueStringPointer(),
				ExceptDirectory:      ruleModel.ExceptDirectory.ValueStringPointer(),
				ControlAction:        ruleModel.ControlAction.ValueStringPointer(),
				RewriteTo:            ruleModel.RewriteTo.ValueStringPointer(),
				Priority:             ruleModel.Priority.ValueInt64Pointer(),
				RefererControlRule:   &cdnetworksapi.RefererControlRule{},
				UaControlRule:        &cdnetworksapi.UaControlRule{},
				AdvanceControlRule:   &cdnetworksapi.AdvanceControlRule{},
				IpControlRule:        &cdnetworksapi.IpControlRule{},
			}
			if !ruleModel.AllowedIps.IsNull() && !ruleModel.AllowedIps.IsUnknown() && len(ruleModel.AllowedIps.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.AllowedIps.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.IpControlRule.AllowedIps = &s
			}
			if !ruleModel.ForbiddenIps.IsNull() && !ruleModel.ForbiddenIps.IsUnknown() && len(ruleModel.ForbiddenIps.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.ForbiddenIps.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.IpControlRule.ForbiddenIps = &s
			}
			rules = append(rules, rule)
		}
	}
	if model.RefererControlRules != nil {
		for _, ruleModel := range model.RefererControlRules {
			rule := &cdnetworksapi.VisitControlRule{
				PathPattern:          ruleModel.PathPattern.ValueStringPointer(),
				ExceptPathPattern:    ruleModel.ExceptPathPattern.ValueStringPointer(),
				CustomPattern:        ruleModel.CustomPattern.ValueStringPointer(),
				FileType:             ruleModel.FileType.ValueStringPointer(),
				CustomFileType:       ruleModel.CustomFileType.ValueStringPointer(),
				SpecifyUrlPattern:    ruleModel.SpecifyUrlPattern.ValueStringPointer(),
				Directory:            ruleModel.Directory.ValueStringPointer(),
				ExceptFileType:       ruleModel.ExceptFileType.ValueStringPointer(),
				ExceptCustomFileType: ruleModel.ExceptCustomFileType.ValueStringPointer(),
				ExceptDirectory:      ruleModel.ExceptDirectory.ValueStringPointer(),
				ControlAction:        ruleModel.ControlAction.ValueStringPointer(),
				RewriteTo:            ruleModel.RewriteTo.ValueStringPointer(),
				Priority:             ruleModel.Priority.ValueInt64Pointer(),
				IpControlRule:        &cdnetworksapi.IpControlRule{},
				UaControlRule:        &cdnetworksapi.UaControlRule{},
				AdvanceControlRule:   &cdnetworksapi.AdvanceControlRule{},
				RefererControlRule: &cdnetworksapi.RefererControlRule{
					AllowNullReferer: ruleModel.AllowNullReferer.ValueBoolPointer(),
				},
			}
			if !ruleModel.ValidUrls.IsNull() && !ruleModel.ValidUrls.IsUnknown() && len(ruleModel.ValidUrls.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.ValidUrls.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.RefererControlRule.ValidUrl = &s
			}
			if !ruleModel.InvalidUrls.IsNull() && !ruleModel.InvalidUrls.IsUnknown() && len(ruleModel.InvalidUrls.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.InvalidUrls.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.RefererControlRule.InvalidUrl = &s
			}
			if !ruleModel.ValidDomains.IsNull() && !ruleModel.ValidDomains.IsUnknown() && len(ruleModel.ValidDomains.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.ValidDomains.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.RefererControlRule.ValidDomain = &s
			}
			if !ruleModel.InvalidDomains.IsNull() && !ruleModel.InvalidDomains.IsUnknown() && len(ruleModel.InvalidDomains.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.InvalidDomains.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.RefererControlRule.InvalidDomain = &s
			}
			rules = append(rules, rule)
		}
	}
	if model.UaControlRules != nil {
		for _, ruleModel := range model.UaControlRules {
			rule := &cdnetworksapi.VisitControlRule{
				PathPattern:          ruleModel.PathPattern.ValueStringPointer(),
				ExceptPathPattern:    ruleModel.ExceptPathPattern.ValueStringPointer(),
				CustomPattern:        ruleModel.CustomPattern.ValueStringPointer(),
				FileType:             ruleModel.FileType.ValueStringPointer(),
				CustomFileType:       ruleModel.CustomFileType.ValueStringPointer(),
				SpecifyUrlPattern:    ruleModel.SpecifyUrlPattern.ValueStringPointer(),
				Directory:            ruleModel.Directory.ValueStringPointer(),
				ExceptFileType:       ruleModel.ExceptFileType.ValueStringPointer(),
				ExceptCustomFileType: ruleModel.ExceptCustomFileType.ValueStringPointer(),
				ExceptDirectory:      ruleModel.ExceptDirectory.ValueStringPointer(),
				ControlAction:        ruleModel.ControlAction.ValueStringPointer(),
				RewriteTo:            ruleModel.RewriteTo.ValueStringPointer(),
				Priority:             ruleModel.Priority.ValueInt64Pointer(),
				IpControlRule:        &cdnetworksapi.IpControlRule{},
				RefererControlRule:   &cdnetworksapi.RefererControlRule{},
				AdvanceControlRule:   &cdnetworksapi.AdvanceControlRule{},
				UaControlRule:        &cdnetworksapi.UaControlRule{},
			}
			if !ruleModel.ValidUserAgents.IsNull() && !ruleModel.ValidUserAgents.IsUnknown() && len(ruleModel.ValidUserAgents.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.ValidUserAgents.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.UaControlRule.ValidUserAgents = &s
			}
			if !ruleModel.InvalidUserAgents.IsNull() && !ruleModel.InvalidUserAgents.IsUnknown() && len(ruleModel.InvalidUserAgents.Elements()) > 0 {
				list := make([]string, 0)
				ruleModel.InvalidUserAgents.ElementsAs(nil, &list, false)
				s := strings.Join(list, utils.Separator)
				rule.UaControlRule.InvalidUserAgents = &s
			}
			rules = append(rules, rule)
		}
	}
	updateHttpConfigRequest := cdnetworksapi.UpdateControlConfigRequest{
		VisitControlRules: rules,
	}
	_, err := r.client.UpdateControlConfig(model.DomainId.ValueString(), updateHttpConfigRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}

func (r *antiHotlinkingConfigResource) updateModel(model *antiHotlinkingConfigModel) error {
	ipRuleIndexMap := make(map[string][]int)
	for i, rule := range model.IpControlRules {
		list, ok := ipRuleIndexMap[rule.String()]
		if !ok {
			list = make([]int, 0)
		}
		ipRuleIndexMap[rule.String()] = append(list, i)
	}
	refererRuleIndexMap := make(map[string][]int)
	for i, rule := range model.RefererControlRules {
		list, ok := refererRuleIndexMap[rule.String()]
		if !ok {
			list = make([]int, 0)
		}
		refererRuleIndexMap[rule.String()] = append(list, i)
	}
	uaRuleIndexMap := make(map[string][]int)
	for i, rule := range model.UaControlRules {
		list, ok := uaRuleIndexMap[rule.String()]
		if !ok {
			list = make([]int, 0)
		}
		uaRuleIndexMap[rule.String()] = append(list, i)
	}

	queryControlConfigResponse, err := r.client.QueryControlConfig(model.DomainId.ValueString())
	if err != nil {
		return err
	}
	model.IpControlRules = make([]*ipControlRuleModel, 0)
	model.RefererControlRules = make([]*refererControlRuleModel, 0)
	model.UaControlRules = make([]*uaControlRuleModel, 0)
	if queryControlConfigResponse.VisitControlRules != nil {
		for _, rule := range queryControlConfigResponse.VisitControlRules {
			if rule.IpControlRule != nil && (rule.IpControlRule.ForbiddenIps != nil || rule.IpControlRule.AllowedIps != nil) {
				ruleModel := &ipControlRuleModel{
					PathPattern:          types.StringPointerValue(rule.PathPattern),
					ExceptPathPattern:    types.StringPointerValue(rule.ExceptPathPattern),
					CustomPattern:        types.StringPointerValue(rule.CustomPattern),
					FileType:             types.StringPointerValue(rule.FileType),
					CustomFileType:       types.StringPointerValue(rule.CustomFileType),
					SpecifyUrlPattern:    types.StringPointerValue(rule.SpecifyUrlPattern),
					Directory:            types.StringPointerValue(rule.Directory),
					ExceptFileType:       types.StringPointerValue(rule.ExceptFileType),
					ExceptCustomFileType: types.StringPointerValue(rule.ExceptCustomFileType),
					ExceptDirectory:      types.StringPointerValue(rule.ExceptDirectory),
					ControlAction:        types.StringPointerValue(rule.ControlAction),
					RewriteTo:            types.StringPointerValue(rule.RewriteTo),
					Priority:             types.Int64PointerValue(rule.Priority),
					AllowedIps:           types.ListNull(types.StringType),
					ForbiddenIps:         types.ListNull(types.StringType),
				}
				if rule.IpControlRule.AllowedIps != nil {
					ips := strings.Split(*rule.IpControlRule.AllowedIps, utils.Separator)
					ruleModel.AllowedIps, _ = types.ListValueFrom(nil, types.StringType, ips)
				} else {
					ruleModel.AllowedIps, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				if rule.IpControlRule.ForbiddenIps != nil {
					ips := strings.Split(*rule.IpControlRule.ForbiddenIps, utils.Separator)
					ruleModel.ForbiddenIps, _ = types.ListValueFrom(nil, types.StringType, ips)
				} else {
					ruleModel.ForbiddenIps, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				model.IpControlRules = append(model.IpControlRules, ruleModel)
			} else if rule.RefererControlRule != nil &&
				(rule.RefererControlRule.AllowNullReferer != nil ||
					rule.RefererControlRule.ValidUrl != nil || rule.RefererControlRule.ValidDomain != nil ||
					rule.RefererControlRule.InvalidUrl != nil ||
					rule.RefererControlRule.InvalidDomain != nil) {
				ruleModel := &refererControlRuleModel{
					PathPattern:          types.StringPointerValue(rule.PathPattern),
					ExceptPathPattern:    types.StringPointerValue(rule.ExceptPathPattern),
					CustomPattern:        types.StringPointerValue(rule.CustomPattern),
					FileType:             types.StringPointerValue(rule.FileType),
					CustomFileType:       types.StringPointerValue(rule.CustomFileType),
					SpecifyUrlPattern:    types.StringPointerValue(rule.SpecifyUrlPattern),
					Directory:            types.StringPointerValue(rule.Directory),
					ExceptFileType:       types.StringPointerValue(rule.ExceptFileType),
					ExceptCustomFileType: types.StringPointerValue(rule.ExceptCustomFileType),
					ExceptDirectory:      types.StringPointerValue(rule.ExceptDirectory),
					ControlAction:        types.StringPointerValue(rule.ControlAction),
					RewriteTo:            types.StringPointerValue(rule.RewriteTo),
					Priority:             types.Int64PointerValue(rule.Priority),
					AllowNullReferer:     types.BoolPointerValue(rule.RefererControlRule.AllowNullReferer),
					ValidUrls:            types.ListNull(types.StringType),
					InvalidUrls:          types.ListNull(types.StringType),
					ValidDomains:         types.ListNull(types.StringType),
					InvalidDomains:       types.ListNull(types.StringType),
				}
				if rule.RefererControlRule.ValidDomain != nil {
					domains := strings.Split(*rule.RefererControlRule.ValidDomain, utils.Separator)
					ruleModel.ValidDomains, _ = types.ListValueFrom(nil, types.StringType, domains)
				} else {
					ruleModel.ValidDomains, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				if rule.RefererControlRule.InvalidDomain != nil {
					domains := strings.Split(*rule.RefererControlRule.InvalidDomain, utils.Separator)
					ruleModel.InvalidDomains, _ = types.ListValueFrom(nil, types.StringType, domains)
				} else {
					ruleModel.InvalidDomains, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				if rule.RefererControlRule.ValidUrl != nil {
					urls := strings.Split(*rule.RefererControlRule.ValidUrl, utils.Separator)
					ruleModel.ValidUrls, _ = types.ListValueFrom(nil, types.StringType, urls)
				} else {
					ruleModel.ValidUrls, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				if rule.RefererControlRule.InvalidUrl != nil {
					urls := strings.Split(*rule.RefererControlRule.InvalidUrl, utils.Separator)
					ruleModel.InvalidUrls, _ = types.ListValueFrom(nil, types.StringType, urls)
				} else {
					ruleModel.InvalidUrls, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				model.RefererControlRules = append(model.RefererControlRules, ruleModel)
			} else if rule.UaControlRule != nil && (rule.UaControlRule.ValidUserAgents != nil || rule.UaControlRule.InvalidUserAgents != nil) {
				ruleModel := &uaControlRuleModel{
					PathPattern:          types.StringPointerValue(rule.PathPattern),
					ExceptPathPattern:    types.StringPointerValue(rule.ExceptPathPattern),
					CustomPattern:        types.StringPointerValue(rule.CustomPattern),
					FileType:             types.StringPointerValue(rule.FileType),
					CustomFileType:       types.StringPointerValue(rule.CustomFileType),
					SpecifyUrlPattern:    types.StringPointerValue(rule.SpecifyUrlPattern),
					Directory:            types.StringPointerValue(rule.Directory),
					ExceptFileType:       types.StringPointerValue(rule.ExceptFileType),
					ExceptCustomFileType: types.StringPointerValue(rule.ExceptCustomFileType),
					ExceptDirectory:      types.StringPointerValue(rule.ExceptDirectory),
					ControlAction:        types.StringPointerValue(rule.ControlAction),
					RewriteTo:            types.StringPointerValue(rule.RewriteTo),
					Priority:             types.Int64PointerValue(rule.Priority),
					ValidUserAgents:      types.ListNull(types.StringType),
					InvalidUserAgents:    types.ListNull(types.StringType),
				}
				if rule.UaControlRule.ValidUserAgents != nil {
					agents := strings.Split(*rule.UaControlRule.ValidUserAgents, utils.Separator)
					ruleModel.ValidUserAgents, _ = types.ListValueFrom(nil, types.StringType, agents)
				} else {
					ruleModel.ValidUserAgents, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				if rule.UaControlRule.InvalidUserAgents != nil {
					agents := strings.Split(*rule.UaControlRule.InvalidUserAgents, utils.Separator)
					ruleModel.InvalidUserAgents, _ = types.ListValueFrom(nil, types.StringType, agents)
				} else {
					ruleModel.InvalidUserAgents, _ = types.ListValueFrom(nil, types.StringType, []string{})
				}
				model.UaControlRules = append(model.UaControlRules, ruleModel)
			}
		}
	}

	// Sort rules to align with .tf configuration file
	ipControlRules := make([]*ipControlRuleModel, len(model.IpControlRules))
	unmatchedIpRules := make([]*ipControlRuleModel, 0)
	for i := 0; i < len(model.IpControlRules); i++ {
		rule := model.IpControlRules[i]
		list, ok := ipRuleIndexMap[rule.String()]
		if ok && len(list) > 0 {
			ipControlRules[list[0]] = rule
			ipRuleIndexMap[rule.String()] = list[1:]
		} else {
			unmatchedIpRules = append(unmatchedIpRules, rule)
		}
	}
	for i, j := 0, 0; i < len(ipControlRules) && j < len(unmatchedIpRules); i++ {
		if ipControlRules[i] != nil {
			continue
		}
		ipControlRules[i] = unmatchedIpRules[j]
		j++
	}
	model.IpControlRules = ipControlRules

	refererControlRules := make([]*refererControlRuleModel, len(model.RefererControlRules))
	unmatchedRefererRules := make([]*refererControlRuleModel, 0)
	for i := 0; i < len(model.RefererControlRules); i++ {
		rule := model.RefererControlRules[i]
		list, ok := refererRuleIndexMap[rule.String()]
		if ok && len(list) > 0 {
			refererControlRules[list[0]] = rule
			refererRuleIndexMap[rule.String()] = list[1:]
		} else {
			unmatchedRefererRules = append(unmatchedRefererRules, rule)
		}
	}
	for i, j := 0, 0; i < len(refererControlRules) && j < len(unmatchedRefererRules); i++ {
		if refererControlRules[i] != nil {
			continue
		}
		refererControlRules[i] = unmatchedRefererRules[j]
		j++
	}
	model.RefererControlRules = refererControlRules

	uaControlRules := make([]*uaControlRuleModel, len(model.UaControlRules))
	unmatchedUaRules := make([]*uaControlRuleModel, 0)
	for i := 0; i < len(model.UaControlRules); i++ {
		rule := model.UaControlRules[i]
		list, ok := uaRuleIndexMap[rule.String()]
		if ok && len(list) > 0 {
			uaControlRules[list[0]] = rule
			uaRuleIndexMap[rule.String()] = list[1:]
		} else {
			unmatchedUaRules = append(unmatchedUaRules, model.UaControlRules[i])
		}
	}
	for i, j := 0, 0; i < len(uaControlRules) && j < len(unmatchedUaRules); i++ {
		if uaControlRules[i] != nil {
			continue
		}
		uaControlRules[i] = unmatchedUaRules[j]
		j++
	}
	model.UaControlRules = uaControlRules

	return nil
}

func (rule *ipControlRuleModel) check() error {
	// check range
	rangeCount := 0
	if !rule.PathPattern.IsNull() && !rule.PathPattern.IsUnknown() {
		rangeCount++
	}
	if !rule.CustomPattern.IsNull() && rule.CustomPattern.ValueString() != "" {
		rangeCount++
	}
	if (!rule.FileType.IsNull() && !rule.FileType.IsUnknown()) || (!rule.CustomFileType.IsNull() && !rule.CustomFileType.IsUnknown()) {
		rangeCount++
	}
	if !rule.Directory.IsNull() && !rule.Directory.IsUnknown() {
		rangeCount++
	}
	if !rule.SpecifyUrlPattern.IsNull() && !rule.SpecifyUrlPattern.IsUnknown() {
		rangeCount++
	}
	if rangeCount == 0 {
		return errors.New("Pick one of the following items: URL matching patterns, directories, file types - customized file types, specified common types and specified URLs)!")
	} else if rangeCount > 1 {
		return errors.New("One and only one of the following items should have value at the same time: path-pattern, directory, (file-type | custom-file-type), custom-pattern, specify-url-pattern.")
	}
	// check file type
	if !rule.FileType.IsNull() && !rule.FileType.IsUnknown() {
		err := utils.CheckFileTypes(rule.FileType.ValueString())
		if err != nil {
			return err
		}
	}
	if !rule.ExceptFileType.IsNull() && !rule.ExceptFileType.IsUnknown() {
		err := utils.CheckFileTypes(rule.ExceptFileType.ValueString())
		if err != nil {
			return err
		}
	}
	// check action
	if !rule.ControlAction.IsNull() && !rule.ControlAction.IsUnknown() && rule.ControlAction.ValueString() == "302" &&
		(rule.RewriteTo.IsNull() || rule.RewriteTo.IsUnknown()) {
		return errors.New("`rerwite_to` is required if the control-action value is 302")
	}

	isPermitListSet := false
	isForbidListSet := false

	if !rule.AllowedIps.IsNull() && !rule.AllowedIps.IsUnknown() && len(rule.AllowedIps.Elements()) > 0 {
		isPermitListSet = true
	}
	if !rule.ForbiddenIps.IsNull() && !rule.ForbiddenIps.IsUnknown() && len(rule.ForbiddenIps.Elements()) > 0 {
		isForbidListSet = true
	}
	if !isPermitListSet && !isForbidListSet {
		return errors.New("One of allowed_ips and forbidden_ips is required.")
	}
	if isPermitListSet && isForbidListSet {
		return errors.New("allowed_ips and forbidden_ips cannot be configured at the same time.")
	}
	return nil
}

func (rule *ipControlRuleModel) String() string {
	attrs := []string{
		rule.PathPattern.String(),
		rule.CustomPattern.String(),
		rule.SpecifyUrlPattern.String(),
		rule.ExceptPathPattern.String(),
		rule.FileType.String(),
		rule.ExceptFileType.String(),
		rule.CustomFileType.String(),
		rule.ExceptCustomFileType.String(),
		rule.Directory.String(),
		rule.ExceptDirectory.String(),
		rule.ControlAction.String(),
		rule.RewriteTo.String(),
		rule.Priority.String(),
		rule.AllowedIps.String(),
		rule.ForbiddenIps.String(),
	}
	return strings.Join(attrs, "$$")
}

func (rule *refererControlRuleModel) check() error {
	// check range
	rangeCount := 0
	if !rule.PathPattern.IsNull() && !rule.PathPattern.IsUnknown() {
		rangeCount++
	}
	if !rule.CustomPattern.IsNull() && rule.CustomPattern.ValueString() != "" {
		rangeCount++
	}
	if (!rule.FileType.IsNull() && !rule.FileType.IsUnknown()) || (!rule.CustomFileType.IsNull() && !rule.CustomFileType.IsUnknown()) {
		rangeCount++
	}
	if !rule.Directory.IsNull() && !rule.Directory.IsUnknown() {
		rangeCount++
	}
	if !rule.SpecifyUrlPattern.IsNull() && !rule.SpecifyUrlPattern.IsUnknown() {
		rangeCount++
	}
	if rangeCount == 0 {
		return errors.New("Pick one of the following items: URL matching patterns, directories, file types - customized file types, specified common types and specified URLs)!")
	} else if rangeCount > 1 {
		return errors.New("One and only one of the following items should have value at the same time: path-pattern, directory, (file-type | custom-file-type), custom-pattern, specify-url-pattern.")
	}
	// check file type
	if !rule.FileType.IsNull() && !rule.FileType.IsUnknown() {
		err := utils.CheckFileTypes(rule.FileType.ValueString())
		if err != nil {
			return err
		}
	}
	if !rule.ExceptFileType.IsNull() && !rule.ExceptFileType.IsUnknown() {
		err := utils.CheckFileTypes(rule.ExceptFileType.ValueString())
		if err != nil {
			return err
		}
	}
	// check action
	if !rule.ControlAction.IsNull() && !rule.ControlAction.IsUnknown() && rule.ControlAction.ValueString() == "302" &&
		(rule.RewriteTo.IsNull() || rule.RewriteTo.IsUnknown()) {
		return errors.New("`rerwite_to` is required if the control-action value is 302")
	}

	cnt := 0
	isNullReferSet := false
	isValidDomainSet := false
	isInvalidDomainSet := false
	isValidUrlSet := false
	isInvalidUrlSet := false
	if !rule.AllowNullReferer.IsNull() && !rule.AllowNullReferer.IsUnknown() {
		isNullReferSet = true
	}
	if !rule.ValidDomains.IsNull() && !rule.ValidDomains.IsUnknown() && len(rule.ValidDomains.Elements()) > 0 {
		cnt++
		isValidDomainSet = true
	}
	if !rule.InvalidDomains.IsNull() && !rule.InvalidDomains.IsUnknown() && len(rule.InvalidDomains.Elements()) > 0 {
		cnt++
		isInvalidDomainSet = true
	}
	if !rule.ValidUrls.IsNull() && !rule.ValidUrls.IsUnknown() && len(rule.ValidUrls.Elements()) > 0 {
		cnt++
		isValidUrlSet = true
	}
	if !rule.InvalidUrls.IsNull() && !rule.InvalidUrls.IsUnknown() && len(rule.InvalidUrls.Elements()) > 0 {
		cnt++
		isInvalidUrlSet = true
	}
	if cnt > 0 && !isNullReferSet {
		return errors.New("If any of the four terms 'nullreferer: legal referer, (legal domain name, legal URL), illegal referer, (illegal domain name, illegal URL)' is allowed, then 'allow_null_referer' cannot be null")
	} else if cnt == 0 && isNullReferSet {
		return errors.New("If the four terms 'legal refer', 'legal domain name, legal URL', 'illegal refer', 'illegal domain name, illegal URL' are all null values, then 'allow_null_referer' must be null")
	}
	if isValidDomainSet && isInvalidDomainSet {
		return errors.New("Only one of the following item should have value at the same time:valid_domain, invalid_domain.")
	}
	if isValidUrlSet && isInvalidUrlSet {
		return errors.New("Only one of the following item should have value at the same time:valid_url, invalid_url.")
	}
	return nil
}

func (rule *refererControlRuleModel) String() string {
	attrs := []string{
		rule.PathPattern.String(),
		rule.CustomPattern.String(),
		rule.SpecifyUrlPattern.String(),
		rule.ExceptPathPattern.String(),
		rule.FileType.String(),
		rule.ExceptFileType.String(),
		rule.CustomFileType.String(),
		rule.ExceptCustomFileType.String(),
		rule.Directory.String(),
		rule.ExceptDirectory.String(),
		rule.ControlAction.String(),
		rule.RewriteTo.String(),
		rule.Priority.String(),
		rule.AllowNullReferer.String(),
		rule.ValidUrls.String(),
		rule.InvalidUrls.String(),
		rule.ValidDomains.String(),
		rule.InvalidDomains.String(),
	}
	return strings.Join(attrs, "$$")
}

func (rule *uaControlRuleModel) check() error {
	// check range
	rangeCount := 0
	if !rule.PathPattern.IsNull() && !rule.PathPattern.IsUnknown() {
		rangeCount++
	}
	if !rule.CustomPattern.IsNull() && rule.CustomPattern.ValueString() != "" {
		rangeCount++
	}
	if (!rule.FileType.IsNull() && !rule.FileType.IsUnknown()) || (!rule.CustomFileType.IsNull() && !rule.CustomFileType.IsUnknown()) {
		rangeCount++
	}
	if !rule.Directory.IsNull() && !rule.Directory.IsUnknown() {
		rangeCount++
	}
	if !rule.SpecifyUrlPattern.IsNull() && !rule.SpecifyUrlPattern.IsUnknown() {
		rangeCount++
	}
	if rangeCount == 0 {
		return errors.New("Pick one of the following items: URL matching patterns, directories, file types - customized file types, specified common types and specified URLs)!")
	} else if rangeCount > 1 {
		return errors.New("One and only one of the following items should have value at the same time: path-pattern, directory, (file-type | custom-file-type), custom-pattern, specify-url-pattern.")
	}
	// check file type
	if !rule.FileType.IsNull() && !rule.FileType.IsUnknown() {
		err := utils.CheckFileTypes(rule.FileType.ValueString())
		if err != nil {
			return err
		}
	}
	if !rule.ExceptFileType.IsNull() && !rule.ExceptFileType.IsUnknown() {
		err := utils.CheckFileTypes(rule.ExceptFileType.ValueString())
		if err != nil {
			return err
		}
	}
	// check action
	if !rule.ControlAction.IsNull() && !rule.ControlAction.IsUnknown() && rule.ControlAction.ValueString() == "302" &&
		(rule.RewriteTo.IsNull() || rule.RewriteTo.IsUnknown()) {
		return errors.New("`rerwite_to` is required if the control-action value is 302")
	}

	isValidAgentSet := false
	isInvalidAgentSet := false
	if !rule.ValidUserAgents.IsNull() && !rule.ValidUserAgents.IsUnknown() && len(rule.ValidUserAgents.Elements()) > 0 {
		isValidAgentSet = true
	}
	if !rule.InvalidUserAgents.IsNull() && !rule.InvalidUserAgents.IsUnknown() && len(rule.InvalidUserAgents.Elements()) > 0 {
		isInvalidAgentSet = true
	}
	if !isValidAgentSet && !isInvalidAgentSet {
		return errors.New("Only one of the following item should have value at the same time:valid_user_agents, invalid_user_agents.")
	}
	return nil
}

func (rule *uaControlRuleModel) String() string {
	attrs := []string{
		rule.PathPattern.String(),
		rule.CustomPattern.String(),
		rule.SpecifyUrlPattern.String(),
		rule.ExceptPathPattern.String(),
		rule.FileType.String(),
		rule.ExceptFileType.String(),
		rule.CustomFileType.String(),
		rule.ExceptCustomFileType.String(),
		rule.Directory.String(),
		rule.ExceptDirectory.String(),
		rule.ControlAction.String(),
		rule.RewriteTo.String(),
		rule.Priority.String(),
		rule.ValidUserAgents.String(),
		rule.InvalidUserAgents.String(),
	}
	return strings.Join(attrs, "$$")
}
