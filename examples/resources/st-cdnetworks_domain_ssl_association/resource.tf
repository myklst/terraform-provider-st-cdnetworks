resource "st-cdnetworks_domain_ssl_association" "test" {
  domain_id          = st-cdnetworks_shield_domain.test.domain_id
  use_ssl            = true
  ssl_certificate_id = st-cdnetworks_ssl_certificate.test.ssl_certificate_id
}
