---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "st-cdnetworks_shield_domain Resource - st-cdnetworks"
subcategory: ""
description: |-
  This resource provides the configuration of acceleration domain
---

# st-cdnetworks_shield_domain (Resource)

This resource provides the configuration of acceleration domain

## Example Usage

```terraform
resource "st-cdnetworks_shield_domain" "test" {
  domain = "www.ccflood.com"
  comment = "test terraform update"
  enabled = true
  header_of_clientip = "Cdn-Src-Ip"

  origin_config = {
    origin_ips = ["2.2.3.2", "2.2.3.1"]
    default_origin_host_header = "b.abc.com"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain` (String) CDN accelerated domain name.
- `origin_config` (Attributes) Back-to-origin policy setting, which is used to set the origin site information and the back-to-origin policy of the accelerated domain name (see [below for nested schema](#nestedatt--origin_config))

### Optional

- `comment` (String) Remarks, up to 1000 characters
- `enabled` (Boolean) Speed up the activation of the domain name. This is false when the accelerated domain name service is disabled; true when the accelerated domain name service is enabled.
- `header_of_clientip` (String) Pass the response header of client IP. The optional values are Cdn-Src-Ip and X-Forwarded-For. The default value is Cdn-Src-Ip.

### Read-Only

- `cdn_service_status` (String) Accelerate the CDN service status of the domain name, true means to enable the CDN acceleration service; false means to cancel the CDN acceleration service.
- `cname` (String) Cname
- `contract_id` (String) The id of contract
- `domain_id` (String) Id of acceleration domain, generated by cdnetworks.
- `item_id` (String) The id of item
- `service_type` (String) Accelerated domain name service types, including the following: 1028 : Content Acceleration; 1115 : Dynamic Web Acceleration; 1369 : Media Acceleration - RTMP 1391 : Download Acceleration 1348 : Media Acceleration Live Broadcast 1551 : Flood Shield
- `status` (String) The deployment status of the accelerate domain name. Deployed indicates that the accelerated domain name configuration is complete. InProgress indicates that the deployment task of the accelerated domain name configuration is still in progress, and may be in queue, deployed, or failed.

<a id="nestedatt--origin_config"></a>
### Nested Schema for `origin_config`

Required:

- `origin_ips` (List of String) Origin site address, which can be an IP or a domain name.
						1. Only one domain name can be entered. IP and domain names cannot be entered at the same time.
						2. Maximum character limit is 500.

Optional:

- `default_origin_host_header` (String) The Origin HOST for changing the HOST field in the return source HTTP request header.
						Note: It should be domain or IP format. For domain name format, each segement separated by a dot, does not exceed 62 characters, the total length should not exceed 128 characters.
