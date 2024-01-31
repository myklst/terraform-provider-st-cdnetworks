package cdnetworksapi

import "strings"

////////////////////////////////////////////////////////////////////////////////
// Domain Property
////////////////////////////////////////////////////////////////////////////////

type CacheBehavior struct {
	PathPattern        *string `json:"path-pattern" xml:"path-pattern"`
	CacheTtl           *int64  `json:"cache-ttl" xml:"cache-ttl"`
	IgnoreCacheControl *bool   `json:"ignore-cache-control" xml:"ignore-cache-control"`
}

// AddCdnDomainService CDNW创建加速域名

type AddCdnDomainRequest struct {
	Version          string        `json:"version,omitempty" xml:"version,omitempty"`
	DomainName       *string       `json:"domain-name,omitempty" xml:"domain-name,omitempty"`
	ContractId       *string       `json:"contract-id,omitempty" xml:"contract-id,omitempty"`
	ItemId           *string       `json:"item-id,omitempty" xml:"item-id,omitempty"`
	Comment          *string       `json:"comment,omitempty" xml:"comment,omitempty"`
	HeaderOfClientIp *string       `json:"header-of-clientip,omitempty" xml:"header-of-clientip,omitempty"`
	OriginConfig     *OriginConfig `json:"origin-config,omitempty" xml:"origin-config,omitempty"`
}

type AddCdnDomainResponse struct {
	Message  *string `json:"message" xml:"message"`
	DomainId *string `json:"-" xml:"-"`
}

func (c *Client) AddCdnDomain(request AddCdnDomainRequest) (response AddCdnDomainResponse, err error) {
	res, err := c.DoJsonApiRequest(Request{
		Method: HttpPost,
		Path:   "/cdnw/api/domain",
		Body:   request,
	}, &response)
	if err != nil {
		return
	}
	location := res.Header.Get("Location")
	id := location[strings.LastIndex(location, "/")+1:]
	response.DomainId = &id
	return
}

// QueryCdnDomainService CDNW查询加速域名

type QueryCdnDomainResponse struct {
	DomainId         *string          `json:"domain-id" xml:"domain-id"`
	DomainName       *string          `json:"domain-name" xml:"domain-name"`
	ContractId       *string          `json:"contract-id" xml:"contract-id"`
	ItemId           *string          `json:"item-id" xml:"item-id"`
	ServiceType      *string          `json:"service-type" xml:"service-type"`
	Comment          *string          `json:"comment" xml:"comment"`
	Cname            *string          `json:"cname" xml:"cname"`
	Status           *string          `json:"status" xml:"status"`
	CdnServiceStatus *string          `json:"cdn-service-status" xml:"cdn-service-status"`
	HeaderOfClientIp *string          `json:"header-of-clientip" xml:"header-of-clientip"`
	Enabled          *bool            `json:"enabled" xml:"enabled"`
	CacheHost        *string          `json:"cache-host" xml:"cache-host"`
	OriginConfig     *OriginConfig    `json:"origin-config" xml:"origin-config"`
	Ssl              *Ssl             `json:"ssl" xml:"ssl"`
	CacheBehaviors   []*CacheBehavior `json:"cache-behaviors" xml:"cache-behaviors>cache-behavior"`
}

func (c *Client) QueryCdnDomain(domainId string) (response QueryCdnDomainResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/cdnw/api/domain/" + domainId,
	}, &response)
	return
}

// QueryApiDomainListService 获取域名列表

type DomainSummary struct {
	DomainId         *string `json:"domain-id" xml:"domain-id"`
	DomainName       *string `json:"domain-name" xml:"domain-name"`
	ServiceType      *string `json:"service-type" xml:"service-type"`
	Cname            *string `json:"cname" xml:"cname"`
	Status           *string `json:"status" xml:"status"`
	CdnServiceStatus *string `json:"cdn-service-status" xml:"cdn-service-status"`
	Enabled          *bool   `json:"enabled" xml:"enabled"`
}

/*
func (c *Client) QueryCdnDomain(domainName string) (response DomainSummary, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/domain/" + domainName,
	}, &response)
	return
}
*/

type QueryApiDomainListResponse struct {
	DomainSummaries []*DomainSummary `json:"domain-summary" xml:"domain-summary"`
}

func (c *Client) QueryApiDomainList(cnameLabel *string) (response QueryApiDomainListResponse, err error) {
	var query map[string]string
	if cnameLabel != nil {
		query = map[string]string{"cname-label": *cnameLabel}
	}
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/domain",
		Query:  query,
	}, &response)
	return
}

// UpdateCdnDomainService CDNW修改加速域名

type UpdateCdnDomainRequest struct {
	Version          string        `json:"version,omitempty" xml:"version,omitempty"`
	DomainName       *string       `json:"domain-name,omitempty" xml:"domain-name,omitempty"`
	Comment          *string       `json:"comment,omitempty" xml:"comment,omitempty"`
	CacheHost        *string       `json:"cache-host,omitempty" xml:"cache-host,omitempty"`
	HeaderOfClientIp *string       `json:"header-of-clientip,omitempty" xml:"header-of-clientip,omitempty"`
	OriginConfig     *OriginConfig `json:"origin-config,omitempty" xml:"origin-config,omitempty"`
	Ssl              *Ssl          `json:"ssl,omitempty" xml:"ssl,omitempty"`
}

type UpdateCdnDomainResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateCdnDomain(domainId string, request UpdateCdnDomainRequest) (response UpdateCdnDomainResponse, err error) {
	_, err = c.DoJsonApiRequest(Request{
		Method: HttpPut,
		Path:   "/cdnw/api/domain/" + domainId,
		Body:   request,
	}, &response)
	return
}

// DeleteApiDomainService 删除单加速域名

type DeleteApiDomainResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) DeleteApiDomain(domainId string) (response DeleteApiDomainResponse, err error) {
	_, err = c.DoJsonApiRequest(Request{
		Method: HttpDelete,
		Path:   "/api/domain/" + domainId,
	}, &response)
	return
}

type EnableDomainResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) EnableDomain(domainId string) (response EnableDomainResponse, err error) {
	_, err = c.DoJsonApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/domain/" + domainId + "/enable",
	}, &response)
	return
}

type DisableDomainResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) DisableDomain(domainId string) (response DisableDomainResponse, err error) {
	_, err = c.DoJsonApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/domain/" + domainId + "/disable",
	}, &response)
	return
}
