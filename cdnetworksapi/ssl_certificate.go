package cdnetworksapi

import (
	"encoding/xml"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// Certificate
////////////////////////////////////////////////////////////////////////////////

// QueryCertificateService 查看证书

type CertificateDomain struct {
	DomainId   *string `json:"domain-id" xml:"domain-id"`
	DomainName *string `json:"domain-name" xml:"domain-name"`
}

type QueryCertificateResponse struct {
	Name                    *string              `json:"name" xml:"name"`
	Comment                 *string              `json:"comment" xml:"comment"`
	ShareSsl                *bool                `json:"share-ssl" xml:"share-ssl"`
	CertificateValidityFrom *string              `json:"certificate-validity-from" xml:"certificate-validity-from"`
	CertificateValidityTo   *string              `json:"certificate-validity-to" xml:"certificate-validity-to"`
	CertificateUpdateTime   *string              `json:"certificate-update-time" xml:"certificate-update-time"`
	CrtMd5                  *string              `json:"crt-md5" xml:"crt-md5"`
	KeyMd5                  *string              `json:"key-md5" xml:"key-md5"`
	CaMd5                   *string              `json:"ca-md5" xml:"ca-md5"`
	CertificateIssuer       *string              `json:"certificate-issuer" xml:"certificate-issuer"`
	CertificateSerial       *string              `json:"certificate-serial" xml:"certificate-serial"`
	RelatedDomains          []*CertificateDomain `json:"related-domains" xml:"related-domains>related-domain"`
	DnsNames                []*string            `json:"dns-names" xml:"dns-names"`
}

func (c *Client) QueryCertificate(certificateId string) (response QueryCertificateResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/ssl/certificate/" + certificateId,
	}, &response)
	return
}

// QueryCertificateListService 查看证书列表

type SslCertificate struct {
	CertificateId           *string              `json:"certificate-id" xml:"certificate-id"`
	Name                    *string              `json:"name" xml:"name"`
	Comment                 *string              `json:"comment" xml:"comment"`
	ShareSsl                *bool                `json:"share-ssl" xml:"share-ssl"`
	CertificateValidityFrom *string              `json:"certificate-validity-from" xml:"certificate-validity-from"`
	CertificateValidityTo   *string              `json:"certificate-validity-to" xml:"certificate-validity-to"`
	CrtMd5                  *string              `json:"crt-md5" xml:"crt-md5"`
	CaMd5                   *string              `json:"ca-md5" xml:"ca-md5"`
	KeyMd5                  *string              `json:"key-md5" xml:"key-md5"`
	CertificateIssuer       *string              `json:"certificate-issuer" xml:"certificate-issuer"`
	CertificateSerial       *string              `json:"certificate-serial" xml:"certificate-serial"`
	RelatedDomains          []*CertificateDomain `json:"related-domains" xml:"related-domains>related-domain"`
	DnsNames                []*string            `json:"dns-names" xml:"dns-names>dns-name"`
}

type QueryCertificateListResponse struct {
	SslCertificates []*SslCertificate `json:"ssl-certificate" xml:"ssl-certificate"`
}

func (c *Client) QueryCertificateList() (response QueryCertificateListResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/ssl/certificate",
	}, &response)
	return
}

// UpdateCertificateService 修改证书

type UpdateCertificateV2Request struct {
	Name        *string `json:"name,omitempty" xml:"name,omitempty"`
	Certificate *string `json:"certificate,omitempty" xml:"certificate,omitempty"`
	PrivateKey  *string `json:"privateKey,omitempty" xml:"privateKey,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

type updateCertificateRequest struct {
	XMLName             xml.Name `json:"-" xml:"ssl-certificate"`
	CsrId               *string  `json:"csr-id,omitempty" xml:"csr-id,omitempty"`
	Name                *string  `json:"name,omitempty" xml:"name,omitempty"`
	Comment             *string  `json:"comment,omitempty" xml:"comment,omitempty"`
	Algorithm           *string  `json:"algorithm,omitempty" xml:"algorithm,omitempty"`
	SslCertificate      *string  `json:"ssl-certificate,omitempty" xml:"ssl-certificate,omitempty"`
	SslCertificateChain *string  `json:"ssl-certificate-chain,omitempty" xml:"ssl-certificate-chain,omitempty"`
	SslKey              *string  `json:"ssl-key,omitempty" xml:"ssl-key,omitempty"`
}

type UpdateCertificateResponse struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) UpdateCertificateV2(certificateId string, request UpdateCertificateV2Request) (response UpdateCertificateResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpPut,
		Path:   "/api/certificate/" + certificateId,
		Body:   request,
	}, &response)
	return
}

// AddCertificateServiceV2 新增证书V2

type AddCertificateV2Request struct {
	Name        *string `json:"name,omitempty" xml:"name,omitempty"`
	Certificate *string `json:"certificate,omitempty" xml:"certificate,omitempty"`
	PrivateKey  *string `json:"privateKey,omitempty" xml:"privateKey,omitempty"`
	Comment     *string `json:"comment,omitempty"`
}

type AddCertificateV2Response struct {
	Code          *string `json:"code" xml:"code"`
	Message       *string `json:"message" xml:"message"`
	CertificateId *string `json:"-" xml:"-"`
}

func (c *Client) AddCertificateV2(request AddCertificateV2Request) (response AddCertificateV2Response, err error) {
	res, err := c.DoXmlApiRequest(Request{
		Method: HttpPost,
		Path:   "/api/certificate",
		Body:   request,
	}, &response)
	if err != nil {
		return
	}
	location := res.Header.Get("Location")
	id := location[strings.LastIndex(location, "/")+1:]
	response.CertificateId = &id
	return
}

// QueryCertificateInfo 查看证书详情V2
type QueryCertificateInfoResponseData struct {
	CertificateId           *int64   `json:"certificateId,omitempty" xml:"certificateId,omitempty"`
	Name                    *string  `json:"name,omitempty" xml:"name,omitempty"`
	Comment                 *string  `json:"comment,omitempty" xml:"comment,omitempty"`
	Serial                  *string  `json:"serial,omitempty" xml:"serial,omitempty"`
	NotBefore               *string  `json:"notBefore,omitempty" xml:"notBefore,omitempty"`
	NotAfter                *string  `json:"notAfter,omitempty" xml:"notAfter,omitempty"`
	CommonName              *string  `json:"commonName,omitempty" xml:"commenName,omitempty"`
	SubjectAlternativeNames []string `json:"subjectAlternativeNames" xml:"subjectAlternativeNames"`
}

type QueryCertificateInfoResponse struct {
	Code                             *int64                            `json:"code" xml:"code"`
	Message                          *string                           `json:"message" xml:"message"`
	QueryCertificateInfoResponseData *QueryCertificateInfoResponseData `json:"data,omitempty" xml:"data,omitempty"`
}

func (c *Client) QueryCertificateInfo(certificateId string) (response QueryCertificateInfoResponse, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpGet,
		Path:   "/api/certificate/" + certificateId,
	}, &response)
	return
}

// DeleteCertificate 删除证书V2

type DeleteCertificateV2Response struct {
	Code    *string `json:"code" xml:"code"`
	Message *string `json:"message" xml:"message"`
}

func (c *Client) DeleteCertificateV2(certificateId string) (response DeleteCertificateV2Response, err error) {
	_, err = c.DoXmlApiRequest(Request{
		Method: HttpDelete,
		Path:   "/api/certificate/" + certificateId,
	}, &response)
	return
}
