resource "st-cdnetworks_http_code_cache_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  http_code_cache_rule {
    http_codes = [401]
    cache_ttl  = 100
  }
}
