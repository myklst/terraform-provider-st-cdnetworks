resource "st-cdnetworks_origin_rules_rewrite_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  origin_rules_rewrite {
    path_pattern = "images/*"
    origin_info  = "alternate.example.com"
    origin_host  = "alternate.example.com"
  }
}
