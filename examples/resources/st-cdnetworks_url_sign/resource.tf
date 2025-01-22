resource "st-cdnetworks_url_sign" "test" {
  domain_id   = "5048000"
  primary_key = "abc123"
  backup_key  = "def456"
  cache_ttl   = 120
}
