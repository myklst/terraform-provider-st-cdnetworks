resource "st-cdnetworks_back_to_origin_protocol_rewrite_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id
  protocol = "https"
  port = "808"
}
