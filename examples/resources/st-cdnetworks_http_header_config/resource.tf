resource "st-cdnetworks_http_header_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  header_rule {
    action           = "add"
    header_direction = "cache2origin"
    header_name      = "x-header-test"
    header_value     = "test"
    path_pattern     = "/*"
  }
}

