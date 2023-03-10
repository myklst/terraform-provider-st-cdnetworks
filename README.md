Terraform Custom Provider for CDNetworks
========================================

This Terraform custom provider is designed for own use case scenario.

Supported Versions
------------------

| Terraform version | minimum provider version |maxmimum provider version
| ---- | ---- | ----|
| >= 1.3.x	| 0.1.0	| latest |

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 1.3.x
-	[Go](https://golang.org/doc/install) 1.19 (to build the provider plugin)

Local Installation
------------------

1. Run `make install-local-custom-provider` to install the provider under ~/.terraform.d/plugins.

2. The provider source should be change to the path that configured in the *Makefile*:

    ```
    terraform {
      required_providers {
        st-cdnetworks = {
          source = "example.local/myklst/st-cdnetworks"
        }
      }
    }

    provider "st-cdnetworks" {}
    ```

Why Custom Provider
-------------------

CDNetworks does not support managing resources with Terraform.


References
----------

- Terraform website: https://www.terraform.io
- Terraform Plugin Framework: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework
- CDNetworks API documentation: https://docs.cdnetworks.com/en/cdn/apidocs
