resource "st-cdnetworks_ssl_certificate" "test" {
  name            = "test"
  ssl_certificate = file("cert.pem")
  ssl_key         = file("key.pem")
}

