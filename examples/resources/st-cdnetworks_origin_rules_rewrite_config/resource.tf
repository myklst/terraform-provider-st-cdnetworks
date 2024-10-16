resource "st-cdnetworks_origin_rules_rewrite_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  oigin_rules_rewrites {
    path_pattern       = "images/*"
    origin_info        = "alternate.example.com"
    origin_host        = "alternate.example.com"
  }
}
