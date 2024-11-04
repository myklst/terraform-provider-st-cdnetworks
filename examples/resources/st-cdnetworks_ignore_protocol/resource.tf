resource "st-cdnetworks_ignore_protocol" "test" {
  domain_id = st-cdnetworks_cdn_domain.test.domain_id

  ignore_protocol_rule {
    path_pattern          = "/test"
    cache_ignore_protocol = true
    purge_ignore_protocol = true
  }
}
