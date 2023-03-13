package cdnetworksapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
)

////////////////////////////////////////////////////////////////////////////////
// Client
////////////////////////////////////////////////////////////////////////////////

const ApiEndpoint = "https://api.cdnetworks.com"

type Client struct {
	Username   string
	ApiKey     string
	httpClient *http.Client
}

func NewClient(username, apiKey string) (*Client, error) {
	var emptyVars []string
	if username == "" {
		emptyVars = append(emptyVars, "username")
	}
	if apiKey == "" {
		emptyVars = append(emptyVars, "apiKey")
	}
	if len(emptyVars) > 0 {
		return nil, fmt.Errorf("cdnetworks client missing: %v", emptyVars)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		Username:   username,
		ApiKey:     apiKey,
		httpClient: httpClient,
	}, nil
}

func NewClientFromEnv() (*Client, error) {
	username := os.Getenv("CDNETWORKS_USERNAME")
	apiKey := os.Getenv("CDNETWORKS_API_KEY")
	return NewClient(username, apiKey)
}

////////////////////////////////////////////////////////////////////////////////
// BaseRequest & BaseResponse
////////////////////////////////////////////////////////////////////////////////

type HttpMethod string

const (
	HttpGet    HttpMethod = http.MethodGet
	HttpPost   HttpMethod = http.MethodPost
	HttpPut    HttpMethod = http.MethodPut
	HttpDelete HttpMethod = http.MethodDelete
)

type BaseRequest struct {
	Method HttpMethod
	Path   string
	Query  map[string]string
	Header http.Header
	Body   []byte
}

type BaseResponse struct {
	Url        string
	StatusCode int
	Header     http.Header
	Body       []byte
}

func getHttpDateNow() string {
	return time.Now().Format(http.TimeFormat)
}

func (c *Client) getPassword(httpDate string) string {
	mac := hmac.New(sha1.New, []byte(c.ApiKey))
	mac.Write([]byte(httpDate))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (c *Client) doApiRequest(request BaseRequest) (*BaseResponse, error) {
	url := ApiEndpoint + request.Path
	body := bytes.NewBuffer(request.Body)
	req, err := http.NewRequest(string(request.Method), url, body)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, value := range request.Query {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	req.Header = request.Header
	if req.Header.Get("Date") == "" {
		req.Header.Set("Date", getHttpDateNow())
	}

	req.SetBasicAuth(c.Username, c.getPassword(req.Header.Get("Date")))

	var res *http.Response
	var resBody []byte

	// Exponentially retry sending the HTTP request until a response is
	// received and its body is read successfully, default max 15 minutes.
	operation := func() error {
		res, err = c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		resBody, err = io.ReadAll(res.Body)
		if err == nil {
			return err
		}
		return nil
	}
	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	return &BaseResponse{
		Url:        url,
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       resBody,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// ErrorResponse
////////////////////////////////////////////////////////////////////////////////

type ErrorResponse struct {
	Url             string `json:"-" xml:"-"`
	RequestId       string `json:"-" xml:"-"`
	StatusCode      int    `json:"-" xml:"-"`
	ResponseCode    string `json:"code" xml:"code"`
	ResponseMessage string `json:"message" xml:"message"`
	RequestBody     string `json:"-" xml:"-"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf(
		"url: %s, request-id: %s, status: %d, code: %s, message: %s, request-body: %s",
		e.Url, e.RequestId, e.StatusCode, e.ResponseCode, e.ResponseMessage, e.RequestBody,
	)
}

////////////////////////////////////////////////////////////////////////////////
// Request, Encoding & Response
////////////////////////////////////////////////////////////////////////////////

type Request struct {
	Method HttpMethod
	Path   string
	Query  map[string]string
	Header http.Header
	Body   interface{}
}

type Encoding struct {
	Name          string
	MarshalFunc   func(interface{}) ([]byte, error)
	UnmarshalFunc func([]byte, interface{}) error
}

type Response = BaseResponse

func (c *Client) doApiRequestWithEncoding(encoding Encoding, request Request, responseBody interface{}) (*Response, error) {
	body, err := encoding.MarshalFunc(request.Body)
	if err != nil {
		return nil, err
	}

	if request.Header == nil {
		request.Header = make(http.Header)
	}
	request.Header.Set("Accept", "application/"+encoding.Name)
	request.Header.Set("Content-Type", "application/"+encoding.Name)

	res, err := c.doApiRequest(BaseRequest{
		Method: request.Method,
		Path:   request.Path,
		Query:  request.Query,
		Header: request.Header,
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		requestId := res.Header.Get("X-Cnc-Request-Id")
		errorResponse := &ErrorResponse{
			Url:         res.Url,
			RequestId:   requestId,
			StatusCode:  res.StatusCode,
			RequestBody: string(body),
		}
		if encoding.UnmarshalFunc(res.Body, errorResponse) == nil &&
			errorResponse.ResponseCode != "" &&
			errorResponse.ResponseMessage != "" {
			return nil, errorResponse
		}
		return nil, fmt.Errorf(
			"request-id: %s, status: %d, response-body: %s, request-body: %s",
			requestId, res.StatusCode, res.Body, body,
		)
	}

	if err = encoding.UnmarshalFunc(res.Body, responseBody); err != nil {
		return nil, err
	}

	return res, nil
}

////////////////////////////////////////////////////////////////////////////////
// Do JSON & XML Request
////////////////////////////////////////////////////////////////////////////////

func (c *Client) DoJsonApiRequest(request Request, responseBody interface{}) (*Response, error) {
	encoding := Encoding{
		Name:          "json",
		MarshalFunc:   json.Marshal,
		UnmarshalFunc: json.Unmarshal,
	}
	return c.doApiRequestWithEncoding(encoding, request, responseBody)
}

func (c *Client) DoXmlApiRequest(request Request, responseBody interface{}) (*Response, error) {
	encoding := Encoding{
		Name:          "xml",
		MarshalFunc:   xml.Marshal,
		UnmarshalFunc: xml.Unmarshal,
	}
	return c.doApiRequestWithEncoding(encoding, request, responseBody)
}
