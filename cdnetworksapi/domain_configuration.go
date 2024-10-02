package cdnetworksapi

import (
	"encoding/xml"
)

var intValue10 = int64(10)

////////////////////////////////////////////////////////////////////////////////
// HTTP Code Cache Config
////////////////////////////////////////////////////////////////////////////////

type HttpCodeCacheRule struct {
	DataId    *string  `json:"data-id,omitempty" xml:"data-id,omitempty"`
	CacheTtl  *int64   `json:"cache-ttl,omitempty" xml:"cache-ttl,omitempty"`
	HttpCodes []*int64 `json:"http-codes,omitempty" xml:"http-codes>http-code,omitempty"`
}

// QueryHttpCodeCacheConfig 查询状态码缓存配置

type QueryHttpCodeCacheConfigResponse struct {
	XMLName            xml.Name             `json:"-" xml:"domain"`
	HttpCodeCacheRules []*HttpCodeCacheRule `json:"http-code-cache-rules" xml:"http-code-cache-rules>http-code-cache-rule"`
}

func (c *Client) QueryHttpCodeCacheConfig(domainId string) (response QueryHttpCodeCacheConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/httpcodecache/" + domainId,
	}, &response)
	return
}

// UpdateHttpCodeCacheConfig 修改状态码缓存配置

type UpdateHttpCodeCacheConfigRequest struct {
	XMLName            xml.Name             `json:"-" xml:"domain"`
	HttpCodeCacheRules []*HttpCodeCacheRule `json:"http-code-cache-rules,omitempty" xml:"http-code-cache-rules>http-code-cache-rule,omitempty"`
}

type UpdateHttpCodeCacheConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateHttpCodeCacheConfig(domainId string, request UpdateHttpCodeCacheConfigRequest) (response UpdateHttpCodeCacheConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/httpcodecache/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Domain Property
////////////////////////////////////////////////////////////////////////////////

type OriginConfig struct {
	OriginIps               *string `json:"origin-ips,omitempty" xml:"origin-ips,omitempty"`
	OriginPort              *string `json:"origin-port,omitempty" xml:"origin-port,omitempty"`
	OriginHost              *string `json:"origin-host,omitempty" xml:"origin-host,omitempty"`
	DefaultOriginHostHeader *string `json:"default-origin-host-header,omitempty" xml:"default-origin-host-header,omitempty"`
}

// UpdateDomainProperty 修改域名属性

type UpdateDomainPropertyRequest struct {
	XMLName      xml.Name      `json:"-" xml:"domain"`
	OriginConfig *OriginConfig `json:"origin-config,omitempty" xml:"origin-config,omitempty"`
}

type UpdateDomainPropertyResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateDomainProperty(domainId string, request UpdateDomainPropertyRequest) (response UpdateDomainPropertyResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/domain/property/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Origin Uri And Origin Host
////////////////////////////////////////////////////////////////////////////////

type OriginRulesRewrite struct {
	DataId                *string `json:"dataId,omitempty" xml:"dataId,omitempty"`
	PathPattern           *string `json:"pathPattern,omitempty" xml:"pathPattern,omitempty"`
	PathPatternHttp       *string `json:"pathPatternHttp,omitempty" xml:"pathPatternHttp,omitempty"`
	ExceptPathPattern     *string `json:"exceptPathPattern,omitempty" xml:"exceptPathPattern,omitempty"`
	ExceptPathPatternHttp *string `json:"exceptPathPatternHttp,omitempty" xml:"exceptPathPatternHttp,omitempty"`
	IgnoreLetterCase      *bool   `json:"ignoreLetterCase,omitempty" xml:"ignoreLetterCase,omitempty"`
	OriginInfo            *string `json:"originInfo,omitempty" xml:"originInfo,omitempty"`
	Priority              *int64  `json:"priority,omitempty" xml:"priority,omitempty"`
	OriginHost            *string `json:"originHost,omitempty" xml:"originHost,omitempty"`
	BeforeRewriteUri      *string `json:"beforeRewriteUri,omitempty" xml:"beforeRewriteUri,omitempty"`
	AfterRewriteUri       *string `json:"afterRewriteUri,omitempty" xml:"afterRewriteUri,omitempty"`
}

// QueryOriginUriAndOriginHost 查询回源uri和host改写

type QueryOriginUriAndOriginHostResponse struct {
	DomainId            *string               `json:"domain-id" xml:"data>domain-id"`
	DomainName          *string               `json:"domain-name" xml:"data>domain-name"`
	OriginRulesRewrites []*OriginRulesRewrite `json:"originRulesRewrites" xml:"data>originRulesRewrites>originRulesRewrite"`
}

func (c *Client) QueryOriginUriAndOriginHost(domainId string) (response QueryOriginUriAndOriginHostResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/originrulesrewrites/" + domainId,
	}, &response)
	return
}

// UpdateOriginUriAndOriginHost 修改回源uri和host改写

type UpdateOriginUriAndOriginHostRequest struct {
	XMLName             xml.Name              `json:"-" xml:"domain"`
	OriginRulesRewrites []*OriginRulesRewrite `json:"originRulesRewrites,omitempty" xml:"originRulesRewrites>originRulesRewrite,omitempty"`
}

type UpdateOriginUriAndOriginHostResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateOriginUriAndOriginHost(domainId string, request UpdateOriginUriAndOriginHostRequest) (response UpdateOriginUriAndOriginHostResponse, err error) {
	for _, v := range request.OriginRulesRewrites {
		if v.Priority == nil {
			v.Priority = &intValue10
		}
	}
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/originrulesrewrites/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// BackToOrigin Protocol Rewrite
////////////////////////////////////////////////////////////////////////////////

type BackToOriginRewriteRule struct {
	Protocol *string `json:"protocol,omitempty" xml:"protocol,omitempty"`
	Port     *string `json:"port,omitempty" xml:"port,omitempty"`
}

// QueryBackToOriginRewriteConfig 查询回源协议接口配置

type QueryBackToOriginRewriteConfigResponse struct {
	DomainId                *string                 `json:"domainId" xml:"data>domainId"`
	DomainName              *string                 `json:"domainName" xml:"data>domainName"`
	BackToOriginRewriteRule BackToOriginRewriteRule `json:"backToOriginRewriteRule" xml:"data>backToOriginRewriteRule"`
}

func (c *Client) QueryBackToOriginRewriteConfig(domainId string) (response QueryBackToOriginRewriteConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/back2originrewrite/" + domainId,
	}, &response)
	return
}

// UpdateBackToOriginRewriteConfig 修改回源协议接口配置

type UpdateBackToOriginRewriteConfigRequest struct {
	XMLName                 xml.Name                `json:"-" xml:"domain"`
	BackToOriginRewriteRule BackToOriginRewriteRule `json:"backToOriginRewriteRule,omitempty" xml:"backToOriginRewriteRule,omitempty"`
}

type UpdateBackToOriginRewriteConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateBackToOriginRewriteConfig(domainId string, request UpdateBackToOriginRewriteConfigRequest) (response UpdateBackToOriginRewriteConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/back2originrewrite/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// IPv6 Config
////////////////////////////////////////////////////////////////////////////////

// QueryIPv6Config 查询域名是否使用ipv6资源

type QueryIPv6ConfigResponse struct {
	DomainId   *string `json:"domain-id" xml:"domain-id"`
	DomainName *string `json:"domain-name" xml:"domain-name"`
	UseIpv6    *bool   `json:"use-ipv6" xml:"use-ipv6"`
}

func (c *Client) QueryIPv6Config(domainId string) (response QueryIPv6ConfigResponse, err error) {
	var baseResp *BaseResponse
	baseResp, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/domain/ipv6/" + domainId,
	}, &response)
	_ = baseResp
	return
}

// UpdateIPv6Config 修改域名是否使用IPv6配置

type UpdateIPv6ConfigRequest struct {
	XMLName   xml.Name `json:"-" xml:"domain"`
	IpVersion []string `json:"ipVersion,omitempty" xml:"ipVersion,omitempty"`
}

type UpdateIPv6ConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

// In Version 2024-09-20 16:30:05, the API only supports JSON format.
func (c *Client) UpdateIPv6Config(domainId string, request UpdateIPv6ConfigRequest) (response UpdateIPv6ConfigResponse, err error) {
	_, err = c.DoJsonApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/ipversion/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Http2 Settings
////////////////////////////////////////////////////////////////////////////////

type Http2Setting struct {
	EnableHttp2          *bool   `json:"enableHttp2,omitempty" xml:"enableHttp2,omitempty"`
	BackToOriginProtocol *string `json:"backToOriginProtocol,omitempty" xml:"backToOriginProtocol,omitempty"`
}

// QueryHttp2SettingsConfigForWplus 查询http2.0开关配置

type QueryHttp2SettingsConfigResponse struct {
	DomainId     *string       `json:"domainId" xml:"data>domainId"`
	DomainName   *string       `json:"domainName" xml:"data>domainName"`
	Http2Setting *Http2Setting `json:"http2Settings" xml:"data>http2Settings"`
}

func (c *Client) QueryHttp2SettingsConfig(domainId string) (response QueryHttp2SettingsConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/http2/" + domainId,
	}, &response)
	return
}

// UpdateHttp2SettingsConfigForWplus 修改http2.0开关配置

type UpdateHttp2SettingsConfigRequest struct {
	XMLName      xml.Name      `json:"-" xml:"domain"`
	Http2Setting *Http2Setting `json:"http2Settings,omitempty" xml:"http2Settings,omitempty"`
}

type UpdateHttp2SettingsConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateHttp2SettingsConfig(domainId string, request UpdateHttp2SettingsConfigRequest) (response UpdateHttp2SettingsConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/http2/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Cache Time
////////////////////////////////////////////////////////////////////////////////

type CacheTimeBehavior struct {
	DataId                     *string `json:"data-id,omitempty" xml:"data-id,omitempty"`
	PathPattern                *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	ExceptPathPattern          *string `json:"except-path-pattern,omitempty" xml:"except-path-pattern,omitempty"`
	CustomPattern              *string `json:"custom-pattern,omitempty" xml:"custom-pattern,omitempty"`
	FileType                   *string `json:"file-type,omitempty" xml:"file-type,omitempty"`
	CustomFileType             *string `json:"custom-file-type,omitempty" xml:"custom-file-type,omitempty"`
	SpecifyUrlPattern          *string `json:"specify-url-pattern,omitempty" xml:"specify-url-pattern,omitempty"`
	Directory                  *string `json:"directory,omitempty" xml:"directory,omitempty"`
	CacheTtl                   *string `json:"cache-ttl,omitempty" xml:"cache-ttl,omitempty"`
	IgnoreCacheControl         *bool   `json:"ignore-cache-control,omitempty" xml:"ignore-cache-control,omitempty"`
	IsRespectServer            *bool   `json:"is-respect-server,omitempty" xml:"is-respect-server,omitempty"`
	IgnoreLetterCase           *bool   `json:"ignore-letter-case,omitempty" xml:"ignore-letter-case,omitempty"`
	ReloadManage               *string `json:"reload-manage,omitempty" xml:"reload-manage,omitempty"`
	IgnoreAuthenticationHeader *bool   `json:"ignore-authentication-header,omitempty" xml:"ignore-authentication-header,omitempty"`
	Priority                   *int64  `json:"priority,omitempty" xml:"priority,omitempty"`
}

// QueryCacheTimeConfig 查询缓存时间配置接口

type QueryCacheTimeConfigResponse struct {
	DomainId           *string              `json:"domain-id" xml:"domain-id"`
	DomainName         *string              `json:"domain-name" xml:"domain-name"`
	CacheTimeBehaviors []*CacheTimeBehavior `json:"cache-time-behaviors" xml:"cache-time-behaviors>cache-time-behavior"`
}

func (c *Client) QueryCacheTimeConfig(domainId string) (response QueryCacheTimeConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/cachetime/" + domainId,
	}, &response)
	return
}

// UpdateCacheTimeConfig 修改缓存时间配置接口

type UpdateCacheTimeConfigRequest struct {
	XMLName            xml.Name             `json:"-" xml:"domain"`
	CacheTimeBehaviors []*CacheTimeBehavior `json:"cache-time-behaviors,omitempty" xml:"cache-time-behaviors>cache-time-behavior"`
}

type UpdateCacheTimeConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateCacheTimeConfig(domainId string, request UpdateCacheTimeConfigRequest) (response UpdateCacheTimeConfigResponse, err error) {
	for _, v := range request.CacheTimeBehaviors {
		if v.Priority == nil {
			v.Priority = &intValue10
		}
	}
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/cachetime/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Redirect Config
////////////////////////////////////////////////////////////////////////////////

type RewriteRuleSetting struct {
	DataId                 *string `json:"data-id,omitempty" xml:"data-id,omitempty"`
	PathPattern            *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	ExceptPathPattern      *string `json:"except-path-pattern,omitempty" xml:"except-path-pattern,omitempty"`
	IgnoreLetterCase       *bool   `json:"ignore-letter-case,omitempty" xml:"ignore-letter-case,omitempty"`
	PublishType            *string `json:"publish-type,omitempty" xml:"publish-type,omitempty"`
	Priority               *int64  `json:"priority,omitempty" xml:"priority,omitempty"`
	BeforeValue            *string `json:"before-value,omitempty" xml:"before-value,omitempty"`
	AfterValue             *string `json:"after-value,omitempty" xml:"after-value,omitempty"`
	RewriteType            *string `json:"rewrite-type,omitempty" xml:"rewrite-type,omitempty"`
	RequestHeader          *string `json:"request-header,omitempty" xml:"request-header,omitempty"`
	ExceptionRequestHeader *string `json:"exception-request-header,omitempty" xml:"exception-request-header,omitempty"`
}

// QueryRedirectConfig 查看域名内部重定向配置

type QueryRedirectConfigResponse struct {
	DomainId            *string               `json:"domain-id" xml:"domain-id"`
	DomainName          *string               `json:"domain-name" xml:"domain-name"`
	RewriteRuleSettings []*RewriteRuleSetting `json:"rewrite-rule-settings" xml:"rewrite-rule-settings>rewrite-rule-setting"`
}

func (c *Client) QueryRedirectConfig(domainId string) (response QueryRedirectConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/InnerRedirect/" + domainId,
	}, &response)
	return
}

// UpdateRedirectConfig 修改域名内部重定向配置

type UpdateRedirectConfigRequest struct {
	XMLName             xml.Name              `json:"-" xml:"domain"`
	RewriteRuleSettings []*RewriteRuleSetting `json:"rewrite-rule-settings,omitempty" xml:"rewrite-rule-settings>rewrite-rule-setting,omitempty"`
}

type UpdateRedirectConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateRedirectConfig(domainId string, request UpdateRedirectConfigRequest) (response UpdateRedirectConfigResponse, err error) {
	for _, v := range request.RewriteRuleSettings {
		if v.Priority == nil {
			v.Priority = &intValue10
		}
	}
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/InnerRedirect/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// HTTP Config
////////////////////////////////////////////////////////////////////////////////

type HeaderModifyRule struct {
	DataId            *int64  `json:"data-id,omitempty" xml:"data-id,omitempty"`
	PathPattern       *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	ExceptPathPattern *string `json:"except-path-pattern,omitempty" xml:"except-path-pattern,omitempty"`
	CustomPattern     *string `json:"custom-pattern,omitempty" xml:"custom-pattern,omitempty"`
	FileType          *string `json:"file-type,omitempty" xml:"file-type,omitempty"`
	CustomFileType    *string `json:"custom-file-type,omitempty" xml:"custom-file-type,omitempty"`
	Directory         *string `json:"directory,omitempty" xml:"directory,omitempty"`
	SpecifyUrl        *string `json:"specify-url,omitempty" xml:"specify-url,omitempty"`
	RequestMethod     *string `json:"request-method,omitempty" xml:"request-method,omitempty"`
	RequestHeader     *string `json:"request-header,omitempty" xml:"request-header,omitempty"`
	HeaderDirection   *string `json:"header-direction,omitempty" xml:"header-direction,omitempty"`
	Action            *string `json:"action,omitempty" xml:"action,omitempty"`
	AllowRegexp       *bool   `json:"allow-regexp,omitempty" xml:"allow-regexp,omitempty"`
	HeaderName        *string `json:"header-name,omitempty" xml:"header-name,omitempty"`
	HeaderValue       *string `json:"header-value,omitempty" xml:"header-value,omitempty"`
}

// QueryHttpConfig 查询http头配置接口

type QueryHttpConfigResponse struct {
	DomainId          *string             `json:"domain-id" xml:"domain-id"`
	DomainName        *string             `json:"domain-name" xml:"domain-name"`
	HeaderModifyRules []*HeaderModifyRule `json:"header-modify-rules" xml:"header-modify-rules>header-modify-rule"`
}

func (c *Client) QueryHttpConfig(domainId string) (response QueryHttpConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/headermodify/" + domainId,
	}, &response)
	return
}

// UpdateHttpConfig 修改http头配置接口

type UpdateHttpConfigRequest struct {
	XMLName           xml.Name            `json:"-" xml:"domain"`
	HeaderModifyRules []*HeaderModifyRule `json:"header-modify-rules,omitempty" xml:"header-modify-rules>header-modify-rule,omitempty"`
}

type UpdateHttpConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateHttpConfig(domainId string, request UpdateHttpConfigRequest) (response UpdateHttpConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/headermodify/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Control Config
////////////////////////////////////////////////////////////////////////////////

type IpControlRule struct {
	ForbiddenIps *string `json:"forbidden-ips,omitempty" xml:"forbidden-ips,omitempty"`
	AllowedIps   *string `json:"allowed-ips,omitempty" xml:"allowed-ips,omitempty"`
}

type RefererControlRule struct {
	AllowNullReferer *bool   `json:"allow-null-referer,omitempty" xml:"allow-null-referer,omitempty"`
	ValidReferer     *string `json:"valid-referers,omitempty" xml:"valid-referers,omitempty"`
	ValidUrl         *string `json:"valid-url,omitempty" xml:"valid-url,omitempty"`
	ValidDomain      *string `json:"valid-domain,omitempty" xml:"valid-domain,omitempty"`
	InvalidReferer   *string `json:"invalid-referers,omitempty" xml:"invalid-referers,omitempty"`
	InvalidUrl       *string `json:"invalid-url,omitempty" xml:"invalid-url,omitempty"`
	InvalidDomain    *string `json:"invalid-domain,omitempty" xml:"invalid-domain,omitempty"`
}

type UaControlRule struct {
	ValidUserAgents   *string `json:"valid-user-agents,omitempty" xml:"valid-user-agents,omitempty"`
	InvalidUserAgents *string `json:"invalid-user-agents,omitempty" xml:"invalid-user-agents,omitempty"`
}

type AdvanceControlRule struct {
	VisitorRegion        *string `json:"visitor-region,omitempty" xml:"visitor-region,omitempty"`
	InvalidVisitorRegion *string `json:"invalid-visitor-region,omitempty" xml:"invalid-visitor-region,omitempty"`
}

type VisitControlRule struct {
	DataId               *string             `json:"data-id,omitempty" xml:"data-id,omitempty"`
	PathPattern          *string             `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	ExceptPathPattern    *string             `json:"except-path-pattern,omitempty" xml:"except-path-pattern,omitempty"`
	CustomPattern        *string             `json:"custom-pattern,omitempty" xml:"custom-pattern,omitempty"`
	FileType             *string             `json:"file-type,omitempty" xml:"file-type,omitempty"`
	CustomFileType       *string             `json:"custom-file-type,omitempty" xml:"custom-file-type,omitempty"`
	SpecifyUrlPattern    *string             `json:"specify-url-pattern,omitempty" xml:"specify-url-pattern,omitempty"`
	Directory            *string             `json:"directory,omitempty" xml:"directory,omitempty"`
	ExceptFileType       *string             `json:"except-file-type,omitempty" xml:"except-file-type,omitempty"`
	ExceptCustomFileType *string             `json:"except-custom-file-type,omitempty" xml:"except-custom-file-type,omitempty"`
	ExceptDirectory      *string             `json:"except-directory,omitempty" xml:"except-directory,omitempty"`
	ControlAction        *string             `json:"control-action,omitempty" xml:"control-action,omitempty"`
	Priority             *int64              `json:"priority,omitempty" xml:"priority,omitempty"`
	RewriteTo            *string             `json:"rewrite-to,omitempty" xml:"rewrite-to,omitempty"`
	IpControlRule        *IpControlRule      `json:"ip-control-rule,omitempty" xml:"ip-control-rule,omitempty"`
	RefererControlRule   *RefererControlRule `json:"referer-control-rule,omitempty" xml:"referer-control-rule,omitempty"`
	UaControlRule        *UaControlRule      `json:"ua-control-rule,omitempty" xml:"ua-control-rule,omitempty"`
	AdvanceControlRule   *AdvanceControlRule `json:"advance-control-rules,omitempty" xml:"advance-control-rules,omitempty"`
}

// QueryControlConfig 查询防盗链配置

type QueryControlConfigResponse struct {
	DomainId          *string             `json:"domain-id" xml:"domain-id"`
	DomainName        *string             `json:"domain-name" xml:"domain-name"`
	VisitControlRules []*VisitControlRule `json:"visit-control-rules" xml:"visit-control-rules>visit-control-rule"`
}

func (c *Client) QueryControlConfig(domainId string) (response QueryControlConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/visitcontrol/" + domainId,
	}, &response)
	return
}

// UpdateControlIpConfig 修改防盗链配置

type UpdateControlConfigRequest struct {
	XMLName           xml.Name            `json:"-" xml:"domain"`
	VisitControlRules []*VisitControlRule `json:"visit-control-rules,omitempty" xml:"visit-control-rules>visit-control-rule,omitempty"`
}

type UpdateControlConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateControlConfig(domainId string, request UpdateControlConfigRequest) (response UpdateControlConfigResponse, err error) {
	for _, v := range request.VisitControlRules {
		if v.Priority == nil {
			v.Priority = &intValue10
		}
	}
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/visitcontrol/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Compression Config
////////////////////////////////////////////////////////////////////////////////

type CompressionSetting struct {
	CompressionEnabled *bool     `json:"compression-enabled,omitempty" xml:"compression-enabled,omitempty"`
	PathPattern        *string   `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	IgnoreLetterCase   *bool     `json:"ignore-letter-case,omitempty" xml:"ignore-letter-case,omitempty"`
	FileTypes          []*string `json:"file-types,omitempty" xml:"file-types>file-type,omitempty"`
}

// QueryCompressionConfig 查询压缩响应配置

type QueryCompressionConfigResponse struct {
	DomainId           *string             `json:"domain-id" xml:"domain-id"`
	DomainName         *string             `json:"domain-name" xml:"domain-name"`
	CompressionSetting *CompressionSetting `json:"compression-settings" xml:"compression-settings"`
}

func (c *Client) QueryCompressionConfig(domainId string) (response QueryCompressionConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/compresssetting/" + domainId,
	}, &response)
	return
}

// UpdateCompressionConfig 修改压缩响应配置

type UpdateCompressionConfigRequest struct {
	XMLName            xml.Name            `json:"-" xml:"domain"`
	CompressionSetting *CompressionSetting `json:"compression-settings,omitempty" xml:"compression-settings,omitempty"`
}

type UpdateCompressionConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateCompressionConfig(domainId string, request UpdateCompressionConfigRequest) (response UpdateCompressionConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/compresssetting/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Query String Config
////////////////////////////////////////////////////////////////////////////////

type QueryStringSetting struct {
	DataId             *string `json:"data-id,omitempty" xml:"data-id,omitempty"`
	PathPattern        *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	IgnoreLetterCase   *bool   `json:"ignore-letter-case,omitempty" xml:"ignore-letter-case,omitempty"`
	IgnoreQueryString  *bool   `json:"ignore-query-string,omitempty" xml:"ignore-query-string,omitempty"`
	QueryStringKept    *string `json:"query-string-kept,omitempty" xml:"query-string-kept,omitempty"`
	QueryStringRemoved *string `json:"query-string-removed,omitempty" xml:"query-string-removed,omitempty"`
	SourceWithQuery    *bool   `json:"source-with-query,omitempty" xml:"source-with-query,omitempty"`
	SourceKeyKept      *string `json:"source-key-kept,omitempty" xml:"source-key-kept,omitempty"`
	SourceKeyRemoved   *string `json:"source-key-removed,omitempty" xml:"source-key-removed,omitempty"`
	FileTypes          *string `json:"file-types,omitempty" xml:"file-types,omitempty"`
	CustomFileTypes    *string `json:"custom-file-types,omitempty" xml:"custom-file-types,omitempty"`
	CustomPattern      *string `json:"custom-pattern,omitempty" xml:"custom-pattern,omitempty"`
	SpecifyUrlPattern  *string `json:"specify-url-pattern,omitempty" xml:"specify-url-pattern,omitempty"`
	Directories        *string `json:"directories,omitempty" xml:"directories,omitempty"`
	Priority           *int64  `json:"priority,omitempty" xml:"priority,omitempty"`
}

// QueryQueryStringConfig 查询去问号缓存配置

type QueryQueryStringConfigResponse struct {
	DomainId           *string               `json:"domain-id" xml:"domain-id"`
	DomainName         *string               `json:"domain-name" xml:"domain-name"`
	QueryStringSetting []*QueryStringSetting `json:"query-string-settings" xml:"query-string-settings>query-string-setting"`
}

func (c *Client) QueryQueryStringConfig(domainId string) (response QueryQueryStringConfigResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/querystring/" + domainId,
	}, &response)
	return
}

// UpdateQueryStringConfig 修改去问号缓存配置

type UpdateQueryStringConfigRequest struct {
	XMLName             xml.Name              `json:"-" xml:"domain"`
	QueryStringSettings []*QueryStringSetting `json:"query-string-settings,omitempty" xml:"query-string-settings>query-string-setting,omitempty"`
}

type UpdateQueryStringConfigResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateQueryStringConfig(domainId string, request UpdateQueryStringConfigRequest) (response UpdateQueryStringConfigResponse, err error) {
	for _, v := range request.QueryStringSettings {
		if v.Priority == nil {
			v.Priority = &intValue10
		}
	}
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/querystring/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Cache Ignore Protocol
////////////////////////////////////////////////////////////////////////////////

type IgnoreProtocolRule struct {
	DataId              *string `json:"data-id,omitempty" xml:"data-id,omitempty"`
	PathPattern         *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	ExceptPathPattern   *string `json:"except-path-pattern,omitempty" xml:"except-path-pattern,omitempty"`
	CacheIgnoreProtocol *bool   `json:"cache-ignore-protocol,omitempty" xml:"cache-ignore-protocol,omitempty"`
	PurgeIgnoreProtocol *bool   `json:"purge-ignore-protocol,omitempty" xml:"purge-ignore-protocol,omitempty"`
}

// QueryIgnoreProtocol 查询忽略协议缓存和推送配置

type QueryIgnoreProtocolResponse struct {
	DomainId            *string               `json:"domain-id" xml:"domain-id"`
	DomainName          *string               `json:"domain-name" xml:"domain-name"`
	IgnoreProtocolRules []*IgnoreProtocolRule `json:"ignore-protocol-rules" xml:"ignore-protocol-rules>ignore-protocol-rule"`
}

func (c *Client) QueryIgnoreProtocol(domainId string) (response QueryIgnoreProtocolResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/config/ignoreprotocol/" + domainId,
	}, &response)
	return
}

// UpdateIgnoreProtocol 修改忽略协议缓存和推送配置

type UpdateIgnoreProtocolRequest struct {
	XMLName             xml.Name              `json:"-" xml:"domain"`
	IgnoreProtocolRules []*IgnoreProtocolRule `json:"ignore-protocol-rules,omitempty" xml:"ignore-protocol-rules>ignore-protocol-rule,omitempty"`
}

type UpdateIgnoreProtocolResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateIgnoreProtocol(domainId string, request UpdateIgnoreProtocolRequest) (response UpdateIgnoreProtocolResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/config/ignoreprotocol/" + domainId,
		Body:   request,
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// Ban Urls
////////////////////////////////////////////////////////////////////////////////

type IllegalInformation struct {
	Url    string   `json:"url,omitempty" xml:"url,omitempty"`
	Method string   `json:"method,omitempty" xml:"method,omitempty"`
	Areas  []string `json:"areas,omitempty" xml:"areas>area,omitempty"`
}

// QueryDomainBanUrls 查询域名URL屏蔽

type QueryDomainBanUrlsResponse struct {
	DomainName          string               `json:"domain-name" xml:"domain-name"`
	DomainId            string               `json:"domain-id" xml:"domain-id"`
	CustomerCode        string               `json:"customer-code" xml:"customer-code"`
	IllegalInformations []IllegalInformation `json:"illegal-informations" xml:"illegal-informations>illegal-information"`
}

func (c *Client) QueryDomainBanUrls(domainId string) (response QueryDomainBanUrlsResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/basicconfig/illegalinformation/" + domainId,
	}, &response)
	return
}

// AddDomainBanURLs 域名新增URL屏蔽
// URL: https://api.cdnetworks.com/api/basicconfig/illegalinformation
// Method: PUT
// ERROR: The Requested URL could not be retrieved.

// DeleteBanURLs 删除Url屏蔽接口

type DeleteDomainBanUrlsRequest struct {
	DomainName string   `json:"domainName,omitempty" xml:"domainName,omitempty"`
	BanUrls    []string `json:"banUrls,omitempty" xml:"banUrls>url,omitempty"`
	DeleteAll  bool     `json:"deleteAll,omitempty" xml:"deleteAll,omitempty"`
}

type DeleteDomainBanUrlsResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) DeleteDomainBanUrls(domainId string) (response DeleteDomainBanUrlsResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpDelete,
		Path:   "/api/basicconfig/illegalinformation",
	}, &response)
	return
}

////////////////////////////////////////////////////////////////////////////////
// API Domain
////////////////////////////////////////////////////////////////////////////////

type AdvOriginConfig struct {
	MasterIps *string `json:"master-ips,omitempty" xml:"master-ips,omitempty"`
	BackupIps *string `json:"backup-ips,omitempty" xml:"backup-ips,omitempty"`
}

type AdvOriginConfigs struct {
	DetectUrl       *string          `json:"detect-url,omitempty" xml:"detect-url,omitempty"`
	DetectPeriod    *int             `json:"detect-period,omitempty" xml:"detect-period,omitempty"`
	AdvOriginConfig *AdvOriginConfig `json:"adv-origin-config,omitempty" xml:"adv-origin-config,omitempty"`
}

type OriginConfigInApiDomain struct {
	OriginIps               *string           `json:"origin-ips,omitempty" xml:"origin-ips,omitempty"`
	DefaultOriginHostHeader *string           `json:"default-origin-host-header,omitempty" xml:"default-origin-host-header,omitempty"`
	OriginPort              *string           `json:"origin-port,omitempty" xml:"origin-port,omitempty"`
	AdvOriginConfigs        *AdvOriginConfigs `json:"adv-origin-configs,omitempty" xml:"adv-origin-configs,omitempty"`
}

type Ssl struct {
	UseSsl           *bool   `json:"use-ssl,omitempty" xml:"use-ssl,omitempty"`
	UseForSni        *bool   `json:"use-for-sni,omitempty" xml:"use-for-sni,omitempty"`
	SslCertificateId *string `json:"ssl-certificate-id,omitempty" xml:"ssl-certificate-id,omitempty"`
}

type ErrorPageRule struct {
	PathPattern *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	IgnoreCase  *bool   `json:"ignore-case,omitempty" xml:"ignore-case,omitempty"`
	ErrorCode   *string `json:"error-code,omitempty" xml:"error-code,omitempty"`
	ForwardUrl  *string `json:"forward-url,omitempty" xml:"forward-url,omitempty"`
}

type AccessSpeedRule struct {
	PathPattern *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	Speed       *int    `json:"speed,omitempty" xml:"speed,omitempty"`
}

type ClientControlRule struct {
	AccessSpeedRules []*AccessSpeedRule `json:"access-speed-rules,omitempty" xml:"access-speed-rules>access-speed-rule,omitempty"`
}

type Videodrags struct {
	PathPattern *string `json:"path-pattern,omitempty" xml:"path-pattern,omitempty"`
	DragMode    *string `json:"drag-mode,omitempty" xml:"drag-mode,omitempty"`
	StartFlag   *string `json:"start-flag,omitempty" xml:"start-flag,omitempty"`
	EndFlag     *string `json:"end-flag,omitempty" xml:"end-flag,omitempty"`
}

// QueryApiDomainService 获取域名基础配置

type QueryApiDomainResponse struct {
	DomainId          *string                  `json:"domain-id" xml:"domain-id"`
	DomainName        *string                  `json:"domain-name" xml:"domain-name"`
	CreatedDate       *string                  `json:"created-date" xml:"created-date"`
	LastModified      *string                  `json:"last-modified" xml:"last-modified"`
	ServiceType       *string                  `json:"service-type" xml:"service-type"`
	ServiceAreas      *string                  `json:"service-areas" xml:"service-areas"`
	ContractId        *string                  `json:"contract-id" xml:"contract-id"`
	ItemId            *string                  `json:"item-id" xml:"item-id"`
	Cname             *string                  `json:"cname" xml:"cname"`
	Comment           *string                  `json:"comment" xml:"comment"`
	Status            *string                  `json:"status" xml:"status"`
	CdnServiceStatus  *string                  `json:"cdn-service-status" xml:"cdn-service-status"`
	Enabled           *bool                    `json:"enabled" xml:"enabled"`
	CacheHost         *string                  `json:"cache-host" xml:"cache-host"`
	HeaderOfClientIp  *string                  `json:"header-of-clientip" xml:"header-of-clientip"`
	OriginConfig      *OriginConfigInApiDomain `json:"origin-config" xml:"origin-config"`
	Ssl               *Ssl                     `json:"ssl" xml:"ssl"`
	ErrorPageRules    []*ErrorPageRule         `json:"error-page-rules" xml:"error-page-rules>error-page-rule"`
	ClientControlRule *ClientControlRule       `json:"client-control-rule" xml:"client-control-rule"`
	Videodrags        *Videodrags              `json:"videodrags" xml:"videodrags"`
}

func (c *Client) QueryApiDomain(domainId string) (response QueryApiDomainResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/domain/" + domainId,
	}, &response)
	return
}

// UpdateApiDomainService 修改域名配置

type UpdateApiDomainRequest struct {
	XMLName           xml.Name                 `xml:"domain"`
	Version           *string                  `json:"version,omitempty" xml:"version,omitempty"`
	Comment           *string                  `json:"comment,omitempty" xml:"comment,omitempty"`
	ServiceAreas      *string                  `json:"service-areas,omitempty" xml:"service-areas,omitempty"`
	CnameLabel        *string                  `json:"cname-label,omitempty" xml:"cname-label,omitempty"`
	HeaderOfClientIp  *string                  `json:"header-of-clientip,omitempty" xml:"header-of-clientip,omitempty"`
	OriginConfig      *OriginConfigInApiDomain `json:"origin-config,omitempty" xml:"origin-config,omitempty"`
	Ssl               *Ssl                     `json:"ssl,omitempty" xml:"ssl,omitempty"`
	ErrorPageRules    []*ErrorPageRule         `json:"error-page-rules,omitempty" xml:"error-page-rules>error-page-rule,omitempty"`
	ClientControlRule *ClientControlRule       `json:"access-speed-rules,omitempty" xml:"client-control-rule,omitempty"`
	Videodrags        *Videodrags              `json:"videodrags,omitempty" xml:"videodrags,omitempty"`
}

type UpdateApiDomainResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateApiDomain(domainId string, request UpdateApiDomainRequest) (response UpdateApiDomainResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/domain/" + domainId,
		Body:   request,
	}, &response)
	return
}
