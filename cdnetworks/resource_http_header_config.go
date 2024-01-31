package cdnetworks

import (
	"context"
	"fmt"
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

type headerRuleModel struct {
	PathPattern       types.String `tfsdk:"path_pattern"`
	ExceptPathPattern types.String `tfsdk:"except_path_pattern"`
	CustomPattern     types.String `tfsdk:"custom_pattern"`
	SpecifyUrl        types.String `tfsdk:"specify_url_pattern"`
	FileType          types.String `tfsdk:"file_type"`
	CustomFileType    types.String `tfsdk:"custom_file_type"`
	Directory         types.String `tfsdk:"directory"`
	Action            types.String `tfsdk:"action"`
	AllowRegexp       types.Bool   `tfsdk:"allow_regexp"`
	HeaderDirection   types.String `tfsdk:"header_direction"`
	HeaderName        types.String `tfsdk:"header_name"`
	HeaderValue       types.String `tfsdk:"header_value"`
	RequestMethod     types.String `tfsdk:"request_method"`
	RequestHeader     types.String `tfsdk:"request_header"`
}

type httpHeaderConfigModel struct {
	DomainId types.String       `tfsdk:"domain_id"`
	Rules    []*headerRuleModel `tfsdk:"header_rule"`
}

type httpHeaderConfigResource struct {
	client *cdnetworksapi.Client
}

var (
	_ resource.Resource                = &httpHeaderConfigResource{}
	_ resource.ResourceWithConfigure   = &httpHeaderConfigResource{}
	_ resource.ResourceWithModifyPlan  = &httpHeaderConfigResource{}
	_ resource.ResourceWithImportState = &httpHeaderConfigResource{}
)

func NewHttpHeaderConfigResource() resource.Resource {
	return &httpHeaderConfigResource{}
}

func (r *httpHeaderConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_header_config"
}

func (r *httpHeaderConfigResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Http header configuration",
		Attributes: map[string]schema.Attribute{
			"domain_id": &schema.StringAttribute{
				Description: "Domain ID",
				Required:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"header_rule": &schema.ListNestedBlock{
				Description: "Header rule",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"except_path_pattern": &schema.StringAttribute{
							Description: "Exception url matching pattern, support regular. Example:",
							Optional:    true,
						},
						"custom_pattern": &schema.StringAttribute{
							Description: "Matching conditions: specify common types, optional values are all or homepage. 1. all: all files 2. homepage: home page",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("all", "homepage", ""),
							},
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"file_type": &schema.StringAttribute{
							Description: "Matching conditions: file type, please separate by semicolon, optional values: gif png bmp jpeg jpg html htm shtml mp3 wma flv mp4 wmv zip exe rar css txt ico js swf m3u8 xml f4m bootstarp ts.",
							Optional:    true,
						},
						"custom_file_type": &schema.StringAttribute{
							Description: "Matching condition: Custom file type, separate by semicolon.",
							Optional:    true,
						},
						"directory": &schema.StringAttribute{
							Description: "directory",
							Optional:    true,
						},
						"specify_url_pattern": &schema.StringAttribute{
							Description: "Matching Condition: Specify URL. The input parameter does not support the URI format starting with http(s)://",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`[^(http(s)?://)].*`), "The input parameter does not support the URI format starting with http(s)://"),
							},
						},
						"request_method": &schema.StringAttribute{
							Description: "The matching request method, the optional values are: GET, POST, PUT, HEAD, DELETE, OPTIONS, separate by semicolons.",
							Optional:    true,
						},
						"path_pattern": &schema.StringAttribute{
							Description: "The url matching mode supports fuzzy regularization. If all matches, the input parameters can be configured as: *",
							Optional:    true,
						},
						"header_direction": &schema.StringAttribute{
							Description: `
                            The control direction of the http header, the optional value is cache2visitor/cache2origin/visitor2cache/origin2cache, single-select.
                            Cache2origin refers to the source direction---corresponding to the configuration item return source request;
                            Cache2visitor refers to the direction of the client back - the corresponding configuration item returns to the client response;
                            Visitor2cache refers to receiving client requests
                            Origin2cache refers to the receiving source response
                            `,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("cache2origin", "cache2visitor", "visitor2cache", "origin2cache"),
							},
						},
						"action": &schema.StringAttribute{
							Description: `
                            The control type of the http header supports the addition and deletion of the http header value. The optional value is add|set|delete, which is single-selected. Corresponding to the header-name and header-value parameters
                            Add: add a header
                            Set: modify the header
                            Delete: delete the header
                            Note: priority is delete>set>add
                            `,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.OneOf("add", "set", "delete"),
							},
						},
						"allow_regexp": &schema.BoolAttribute{
							Description: `
                            Http header regular match, optional value: true / false.
                            True: indicates that the value of the header-name is handled as a regular match.
                            False: indicates that the value of the header-name is processed according to the actual parameters, and no regular match is made.
                            Do not pass the default is false
                            `,
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"header_name": &schema.StringAttribute{
							Description: `
                            Http header name, add or modify the http header, only one is allowed; delete the http header to allow multiple entries, separated by a semicolon ';'.
                            Note: The operation of the special http header is limited, and the http header and operation type of the operation are allowed.
                            This item is required and cannot be empty
                            When the action is add: indicates that the header-name header is added.
                            When the action is set: modify the header-name header
                            When the action is delete: delete the header-name header
                            `,
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"header_value": &schema.StringAttribute{
							Description: `
            The value corresponding to the HTTP header field, for example: mytest.example.com

            Note:

            1. When the action is add or set, the input parameter must be passed a value

            2. When the action is delete, the input parameter is not passed

            Support to get the value of specified variable by keyword, such as client IP, including:

            Key words: meaning

#timestamp: current time, timestamp as 1559124945
#request-host: host in the request header
#request-url: request url, which contains the full path of the protocol domain name, etc., such as http://aaa.aa.com/a.html
#request-uri: request uri, relative path format, such as /index.html
#origin- IP: return source IP
#cache-ip: edge node IP
#server-ip: external service IP
#client-ip: client IP, or visitor IP
#response-header{XXX} : get the value in the response header, such as #response-header{etag}, get the etag value in response-header

#header{XXX} : to get the value in the HTTP header of the request, such as #header{user-agent}, is to get the user-agent value in the header

#cookie{XXX} : get the value in the cookie, such as #cookie{account}, is to get the value of the account set in the cookie
                            `,
							Optional: true,
						},
						"request_header": &schema.StringAttribute{
							Description: "Match request header, header values support regular, header and header values separated by Spaces, e.g. : Range bytes=[0-9]{9,}",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (r *httpHeaderConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*cdnetworksapi.Client)
}

func (r *httpHeaderConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model *httpHeaderConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR] Fail to update http header", err.Error())
	}
	resp.State.Set(ctx, model)
}

func (r *httpHeaderConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model *httpHeaderConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateModel(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to read_http_header_config", err.Error())
		return
	}

	resp.State.Set(ctx, &model)
}

func (r *httpHeaderConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *httpHeaderConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	err := r.updateConfig(plan)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to update http header config", err.Error())
		return
	}
	resp.State.Set(ctx, plan)
}

func (r *httpHeaderConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model *httpHeaderConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Rules = nil
	err := r.updateConfig(model)
	if err != nil {
		resp.Diagnostics.AddError("[API ERROR]Fail to delete http head configs", err.Error())
	}
}

func (r *httpHeaderConfigResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan *httpHeaderConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan == nil {
		return
	}

	if plan.Rules == nil {
		resp.Diagnostics.AddError("[Validate Config]", "Rules are required")
		return
	}

	for _, rule := range plan.Rules {
		if rule == nil {
			continue
		}
		err := rule.check()
		if err != nil {
			resp.Diagnostics.AddError("[Validate Config]Invalid config", err.Error())
			return
		}
	}

	resp.Plan.Set(ctx, &plan)
}

func (r *httpHeaderConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("domain_id"), req, resp)
}

func (r *httpHeaderConfigResource) updateConfig(model *httpHeaderConfigModel) error {
	rules := make([]*cdnetworksapi.HeaderModifyRule, 0)
	if model.Rules != nil {
		for _, ruleModel := range model.Rules {
			rule := &cdnetworksapi.HeaderModifyRule{
				PathPattern:       ruleModel.PathPattern.ValueStringPointer(),
				ExceptPathPattern: ruleModel.ExceptPathPattern.ValueStringPointer(),
				CustomPattern:     ruleModel.CustomPattern.ValueStringPointer(),
				FileType:          ruleModel.FileType.ValueStringPointer(),
				CustomFileType:    ruleModel.CustomFileType.ValueStringPointer(),
				Directory:         ruleModel.Directory.ValueStringPointer(),
				SpecifyUrl:        ruleModel.SpecifyUrl.ValueStringPointer(),
				RequestMethod:     ruleModel.RequestMethod.ValueStringPointer(),
				HeaderDirection:   ruleModel.HeaderDirection.ValueStringPointer(),
				Action:            ruleModel.Action.ValueStringPointer(),
				AllowRegexp:       ruleModel.AllowRegexp.ValueBoolPointer(),
				HeaderName:        ruleModel.HeaderName.ValueStringPointer(),
				HeaderValue:       ruleModel.HeaderValue.ValueStringPointer(),
			}
			rules = append(rules, rule)
		}
	}
	updateHttpConfigRequest := cdnetworksapi.UpdateHttpConfigRequest{
		HeaderModifyRules: rules,
	}
	_, err := r.client.UpdateHttpConfig(model.DomainId.ValueString(), updateHttpConfigRequest)
	if err != nil {
		return err
	}
	return utils.WaitForDomainDeployed(r.client, model.DomainId.ValueString())
}

func (r *httpHeaderConfigResource) updateModel(model *httpHeaderConfigModel) error {
	ruleIndexMap := make(map[string][]int)
	for i, rule := range model.Rules {
		list, ok := ruleIndexMap[rule.String()]
		if !ok {
			list = make([]int, 0)
		}
		ruleIndexMap[rule.String()] = append(list, i)
	}

	queryHttpConfigResponse, err := r.client.QueryHttpConfig(model.DomainId.ValueString())
	if err != nil {
		return err
	}
	model.Rules = make([]*headerRuleModel, 0)
	if queryHttpConfigResponse.HeaderModifyRules != nil {
		for _, rule := range queryHttpConfigResponse.HeaderModifyRules {
			ruleModel := &headerRuleModel{
				ExceptPathPattern: types.StringPointerValue(rule.ExceptPathPattern),
				CustomPattern:     types.StringPointerValue(rule.CustomPattern),
				FileType:          types.StringPointerValue(rule.FileType),
				CustomFileType:    types.StringPointerValue(rule.CustomFileType),
				Directory:         types.StringPointerValue(rule.Directory),
				SpecifyUrl:        types.StringPointerValue(rule.SpecifyUrl),
				RequestMethod:     types.StringPointerValue(rule.RequestMethod),
				PathPattern:       types.StringPointerValue(rule.PathPattern),
				HeaderDirection:   types.StringPointerValue(rule.HeaderDirection),
				Action:            types.StringPointerValue(rule.Action),
				AllowRegexp:       types.BoolPointerValue(rule.AllowRegexp),
				HeaderName:        types.StringPointerValue(rule.HeaderName),
				HeaderValue:       types.StringPointerValue(rule.HeaderValue),
				RequestHeader:     types.StringPointerValue(rule.RequestHeader),
			}
			model.Rules = append(model.Rules, ruleModel)
		}
	}

	// Sort rules to align with .tf configuaration file
	rules := make([]*headerRuleModel, len(model.Rules))
	unmatchedRules := make([]*headerRuleModel, 0)
	for i := 0; i < len(model.Rules); i++ {
		rule := model.Rules[i]
		list, ok := ruleIndexMap[rule.String()]
		if ok && len(list) > 0 {
			rules[list[0]] = rule
			ruleIndexMap[rule.String()] = list[1:]
		} else {
			unmatchedRules = append(unmatchedRules, rule)
		}
	}
	for i, j := 0, 0; i < len(rules) && j < len(unmatchedRules); i++ {
		if rules[i] != nil {
			continue
		}
		rules[i] = unmatchedRules[j]
		j++
	}
	model.Rules = rules

	return nil
}

func (rule *headerRuleModel) check() error {
	if !rule.RequestMethod.IsNull() && !rule.RequestMethod.IsUnknown() {
		values := strings.Split(rule.RequestMethod.ValueString(), utils.Separator)
		for _, v := range values {
			valid := false
			for _, method := range utils.ValidHttpMethods {
				if strings.TrimSpace(v) == method {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid http method:%s", v)
			}
		}
	}
	// check range
	if (rule.CustomPattern.IsUnknown() || rule.CustomPattern.IsNull()) &&
		(rule.Directory.IsUnknown() || rule.Directory.IsNull()) &&
		(rule.SpecifyUrl.IsUnknown() || rule.SpecifyUrl.IsNull()) &&
		(rule.CustomFileType.IsUnknown() || rule.CustomFileType.IsNull()) &&
		(rule.PathPattern.IsUnknown() || rule.PathPattern.IsNull()) &&
		(rule.FileType.IsUnknown() || rule.FileType.IsNull()) {
		return fmt.Errorf("Pick one of the following items: path_pattern, custom_pattern, file_type, custom_file_type, directory and specify_url_pattern")
	}
	// check action
	if rule.Action.IsNull() || rule.Action.IsUnknown() {
		return fmt.Errorf("`action` is required")
	}
	// check header_name and header_value
	if rule.HeaderName.IsNull() || rule.HeaderName.IsUnknown() || rule.HeaderName.ValueString() == "" {
		return fmt.Errorf("`header_name` is required")
	}
	if rule.Action.ValueString() == "add" || rule.Action.ValueString() == "set" {
		values := strings.Split(rule.HeaderName.ValueString(), utils.Separator)
		if len(values) > 1 {
			return fmt.Errorf("Only ONE header_name is allowed when `action` is add or set")
		}
		if rule.HeaderValue.IsNull() || rule.HeaderValue.IsUnknown() || rule.HeaderValue.ValueString() == "" {
			return fmt.Errorf("header_value is required when `action` is add or set")
		}
	} else {
		if !rule.HeaderValue.IsNull() && !rule.HeaderValue.IsUnknown() {
			return fmt.Errorf("No header-value was not need when `action` is delete")
		}
	}
	if rule.HeaderDirection.IsNull() || rule.HeaderDirection.IsUnknown() {
		return fmt.Errorf("`header_direction` is required")
	}
	// check file_type
	if !rule.FileType.IsNull() && !rule.FileType.IsUnknown() {
		err := utils.CheckFileTypes(rule.FileType.ValueString())
		if err != nil {
			return err
		}
	}
	return nil

}

func (rule *headerRuleModel) String() string {
	values := []string{
		rule.PathPattern.String(),
		rule.ExceptPathPattern.String(),
		rule.CustomPattern.String(),
		rule.SpecifyUrl.String(),
		rule.FileType.String(),
		rule.CustomFileType.String(),
		rule.Directory.String(),
		rule.Action.String(),
		rule.AllowRegexp.String(),
		rule.HeaderDirection.String(),
		rule.HeaderName.String(),
		rule.HeaderValue.String(),
		rule.RequestMethod.String(),
		rule.RequestHeader.String(),
	}
	return strings.Join(values, "$$")
}
