resource "st-cdnetworks_http2_settings_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  http2_settings = {
    enable_http2            = true
    back_to_origin_protocol = "http2.0"
  }
}
