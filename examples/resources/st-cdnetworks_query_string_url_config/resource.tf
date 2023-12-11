resource "st-cdnetworks_query_string_url_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  query_string_setting {
    path_pattern = "*"
    query_string_removed = "test_query"
    source_key_kept = "test_key"
  }
}
