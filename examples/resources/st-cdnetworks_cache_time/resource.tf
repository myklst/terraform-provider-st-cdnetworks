resource "st-cdnetworks_cache_time" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  cache_time_behavior {
    directory = "/abc/"
    cache_ttl = 100
  }
}
