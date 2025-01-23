resource "st-cdnetworks_url_sign" "test" {
  domain_id     = "5048000"
  primary_key   = "abc123"
  secondary_key = "def456"
  ttl           = 120
}
