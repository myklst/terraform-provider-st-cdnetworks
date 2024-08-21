package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworksapi"
)

func WaitForDomainDeployed(client *cdnetworksapi.Client, domainId string) error {
	checkStatus := func() error {
		// QueryCdnDomains ratelimit is 300/s, normally take ~10mins to successfully update.
		queryCdnDomainResponse, err := client.QueryCdnDomain(domainId)
		if err != nil {
			return err
		}

		if *queryCdnDomainResponse.Status == "Deployed" {
			return nil
		}

		if *queryCdnDomainResponse.Status == "Reviewing" {
			return backoff.Permanent(fmt.Errorf("status is in reviewing, please contact to vendor"))
		}

		return errors.New("deployment is in progress")
	}

	r := backoff.NewExponentialBackOff()
	r.InitialInterval = 30 * time.Second
	r.MaxElapsedTime = 0 // set as infinite retries.

	return backoff.Retry(checkStatus, r)
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

	r := backoff.NewExponentialBackOff()
	r.MaxElapsedTime = 0 // set as infinite retries.

	return backoff.Retry(checkStatus, r)
}
