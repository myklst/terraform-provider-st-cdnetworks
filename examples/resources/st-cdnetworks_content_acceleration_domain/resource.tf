resource "st-cdnetworks_content_acceleration_domain" "test" {
  domain             = "www.ccflood.com"
  comment            = "test terraform update"
  enabled            = true
  header_of_clientip = "Cdn-Src-Ip"

  origin_config = {
    origin_ips                 = ["2.2.3.2", "2.2.3.1"]
    default_origin_host_header = "b.abc.com"
  }
  contract_id   = ""
  item_id       = ""
  control_group = {}
}


