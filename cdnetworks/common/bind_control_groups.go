package common

import (
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks/model"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

func BindCdnDomainToControlGroup(client *cdnetworksapi.Client, model *model.DomainResourceModel) (err error) {
edit:
	_, err = client.EditControlGroup(model.BuildEditControlGroupRequest())
	if err != nil {
		return err
	}

	// Due to concurrent of EditControlGroup(), some domains doesn't bind successfully.
	// Query once and find if domains already bind into control_group.
	resp, err := client.GetDomainListOfControlGroup(&cdnetworksapi.GetDomainListOfControlGroupRequest{
		ControlGroupCode: []string{model.ControlGroup.Code.ValueString()},
	})
	if err != nil {
		return err
	}

	found := false
	for _, detail := range resp.Data.ControlGroupDetails {
		for _, domain := range detail.DomainList {
			if domain == model.Domain.ValueString() {
				found = true
				break
			}
		}
	}

	if !found {
		goto edit
	}

	return nil
}
