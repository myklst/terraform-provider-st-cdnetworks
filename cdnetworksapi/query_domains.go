package cdnetworksapi

type QueryDomainResponse struct {
	DomainId         *string `json:"domain-id" xml:"domain-id" tfsdk:"domain_id"`
	DomainName       *string `json:"domain-name" xml:"domain-name" tfsdk:"domain_name"`
	CreatedDate      *string `json:"created-date" xml:"created-date" tfsdk:"created_date"`
	LastModified     *string `json:"last-modified" xml:"last-modified" tfsdk:"last_modified"`
	ServiceType      *string `json:"service-type" xml:"service-type" tfsdk:"service_type"`
	Comment          *string `json:"comment" xml:"comment" tfsdk:"comment"`
	Cname            *string `json:"cname" xml:"cname" tfsdk:"cname"`
	Status           *string `json:"status" xml:"status" tfsdk:"status"`
	CdnServiceStatus *string `json:"cdn-service-status" xml:"cdn-service-status" tfsdk:"cdn_service_status"`
	Enabled          *bool   `json:"enabled" xml:"enabled" tfsdk:"enabled"`
	UseRange         *bool   `json:"useRange" xml:"useRange" tfsdk:"useRange"`
	Follow301        *bool   `json:"follow301" xml:"follow301" tfsdk:"follow301"`
	Follow302        *bool   `json:"follow302" xml:"follow302" tfsdk:"follow302"`
	OriginConfig     *struct {
		OriginIps               *string `json:"origin-ips" xml:"origin-ips" tfsdk:"origin_ips"`
		DefaultOriginHostHeader *string `json:"default-origin-host-header" xml:"default-origin-host-header" tfsdk:"default_origin_host_header"`
		AdvOriginConfigs        *struct {
			DetectUrl       *string `json:"detect-url" xml:"detect-url" tfsdk:"detect_url"`
			DetectPeriod    *int    `json:"detect-period" xml:"detect-period" tfsdk:"detect_period"`
			AdvOriginConfig []*struct {
				MasterIps *string `json:"master-ips" xml:"master-ips" tfsdk:"master_ips"`
				BackupIps *string `json:"backup-ips" xml:"backup-ips" tfsdk:"backup_ips"`
			} `json:"adv-origin-config" xml:"adv-origin-config" tfsdk:"adv_origin_config"`
		} `json:"adv-origin-configs" xml:"adv-origin-configs" tfsdk:"adv_origin_configs"`
	} `json:"origin-config" xml:"origin-config" tfsdk:"origin_config"`
	Ssl *struct {
		UseSsl           *bool   `json:"use-ssl" xml:"use-ssl" tfsdk:"use_ssl"`
		UseForSni        *bool   `json:"use-for-sni" xml:"use-for-sni" tfsdk:"use_for_sni"`
		SslCertificateId *string `json:"ssl-certificate-id" xml:"ssl-certificate-id" tfsdk:"ssl_certificate_id"`
	} `json:"ssl" xml:"ssl"`
	CacheBehaviors []*struct {
		PathPattern        *string `json:"path-pattern" xml:"path-pattern" tfsdk:"path_pattern"`
		CacheTtl           *int    `json:"cache-ttl" xml:"cache-ttl" tfsdk:"cache_ttl"`
		IgnoreCacheControl *bool   `json:"ignore-cache-control" xml:"ignore-cache-control" tfsdk:"ignore_cache_control"`
	} `json:"cache-behaviors" xml:"cache-behaviors>cache-behavior" tfsdk:"cache_behaviors"`
	QueryStringSettings []*struct {
		PathPattern       *string `json:"path-pattern" xml:"path-pattern" tfsdk:"path_pattern"`
		IgnoreQueryString *bool   `json:"ignore-query-string" xml:"ignore-query-string" tfsdk:"ignore_query_string"`
	} `json:"query-string-settings" xml:"query-string-settings>query-string-setting" tfsdk:"query_string_settings"`
	CacheKeyRules []*struct {
		PathPattern *string `json:"path-pattern" xml:"path-pattern" tfsdk:"path_pattern"`
		IgnoreCase  *bool   `json:"ignore-case" xml:"ignore-case" tfsdk:"ignore_case"`
		HeaderName  *string `json:"header-name" xml:"header-name" tfsdk:"header_name"`
	} `json:"cache-key-rules" xml:"cache-key-rules>cache-key-rule" tfsdk:"cache_key_rules"`
	VisitControlRules []*struct {
		PathPattern      *string   `json:"path-pattern" xml:"path-pattern" tfsdk:"path_pattern"`
		AllowNullReferer *bool     `json:"allownullreferer" xml:"allownullreferer" tfsdk:"allownullreferer"`
		InvalidReferers  []*string `json:"invalid-referers" xml:"invalid-referers" tfsdk:"invalid_referers"`
		ForbiddenIps     *string   `json:"forbidden-ips" xml:"forbidden-ips" tfsdk:"forbidden_ips"`
	} `json:"visit-control-rules" xml:"visit-control-rules>visit-control-rule" tfsdk:"visit_control_rules"`
	ErrorPageRules []*struct {
		PathPattern *string `json:"path-pattern" xml:"path-pattern" tfsdk:"path_pattern"`
		IgnoreCase  *bool   `json:"ignore-case" xml:"ignore-case" tfsdk:"ignore_case"`
		ErrorCode   *string `json:"error-code" xml:"error-code" tfsdk:"error_code"`
		ForwardUrl  *string `json:"forward-url" xml:"forward-url" tfsdk:"forward_url"`
	} `json:"error-page-rules" xml:"error-page-rules>error-page-rule" tfsdk:"error_page_rules"`
	ClientControlRule *struct {
		AccessSpeedRules []*struct {
			PathPattern *string `json:"path-pattern" xml:"path-pattern" tfsdk:"path_pattern"`
			Speed       *int    `json:"speed" xml:"speed" tfsdk:"speed"`
		} `json:"access-speed-rules" xml:"access-speed-rules>access-speed-rule" tfsdk:"access_speed_rules"`
	} `json:"client-control-rule" xml:"client-control-rule" tfsdk:"client_control_rule"`
	VideoDrags *struct {
		PathPattern *string `json:"path-pattern" xml:"path-pattern" tfsdk:"path_pattern"`
		DragMode    *string `json:"drag-mode" xml:"drag-mode" tfsdk:"drag_mode"`
		StartFlag   *string `json:"start-flag" xml:"start-flag" tfsdk:"start_flag"`
		EndFlag     *string `json:"end-flag" xml:"end-flag" tfsdk:"end_flag"`
	} `json:"videodrags" xml:"videodrags" tfsdk:"videodrags"`
}

func (c *Client) QueryDomain(domainId string) (response QueryDomainResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/cdnw/api/domain/" + domainId,
	}, &response)
	return
}
