package cdnetworksapi

// EditControlGroup 修改ControlGroup接口, 域名才会显示在对应的账号上.
type EditControlGroupRequest struct {
	ControlGroupName *string    `json:"controlGroupName,omitempty" xml:"controlGroupName,omitempty"`
	DomainList       []*string  `json:"domainList,omitempty" xml:"domainList,omitempty"`
	AccountList      []*Account `json:"accountList,omitempty" xml:"accountList,omitempty"`
	IsAdd            bool       `json:"isAdd,omitempty" xml:"isAdd,omitempty"`
}

type Account struct {
	LoginName *string `json:"loginName,omitempty" xml:"loginName,omitempty" require:"true"`
}

type EditControlGroupResponse struct {
	Message   *string `json:"msg" xml:"msg"`
	RequestId *string `json:"requestId" xml:"requestId"`
}

func (c *Client) EditControlGroup(controlGroupCode string, request *EditControlGroupRequest) (response EditControlGroupResponse, err error) {
	_, err = c.DoJsonApiRequest(Request{
		Method: HttpPut,
		Path:   "/user/control-groups/" + controlGroupCode,
		Body:   request,
	}, &response)

	return
}

type GetDomainListOfControlGroupRequest struct {
	ControlGroupCode []string `json:"controlGroupCode,omitempty" xml:"controlGroupCode,omitempty"`
}

type GetDomainListOfControlGroupResponse struct {
	Message   *string `json:"msg" xml:"msg"`
	RequestId *string `json:"requestId" xml:"requestId"`
	Data      *struct {
		ControlGroupDetails []*controlGroupDetail `json:"controlGroupDetail" xml:"controlGroupDetail"`
	} `json:"data" xml:"data"`
}

type controlGroupDetail struct {
	ControlGroupCode *string  `json:"controlGroupCode" xml:"controlGroupCode"`
	ControlGroupName *string  `json:"controlGroupName" xml:"controlGroupName"`
	DomainList       []string `json:"domainList" xml:"domainList"`
}

func (c *Client) GetDomainListOfControlGroup(request *GetDomainListOfControlGroupRequest) (response *GetDomainListOfControlGroupResponse, err error) {
	_, err = c.DoJsonApiRequest(Request{
		Method: HttpPost,
		Path:   "/user/cgdomainlist",
		Body:   request,
	}, &response)

	return
}
