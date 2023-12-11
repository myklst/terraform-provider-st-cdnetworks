package utils

import (
	"errors"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

func WaitForDomainDeployed(client *cdnetworksapi.Client, domainId string) error {
	checkStatus := func() error {
		queryCdnDomainResponse, err := client.QueryCdnDomain(domainId)
		if err != nil {
			return err
		}
		if *queryCdnDomainResponse.Status == "Deployed" {
			return nil
		}
		return errors.New("deployment is in progress")
	}

	return backoff.Retry(checkStatus, backoff.NewExponentialBackOff())
}

func WaitForDomainDeleted(client *cdnetworksapi.Client, domainId string) error {
	checkStatus := func() error {
		_, err := client.QueryCdnDomain(domainId)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				return nil
			}
			return err
		}
		return errors.New("deployment is in progress")
	}
	return backoff.Retry(checkStatus, backoff.NewExponentialBackOff())
}
