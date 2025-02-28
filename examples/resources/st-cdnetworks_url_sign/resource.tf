resource "st-cdnetworks_url_sign" "test" {
  domain_id                    = "5048000"
  primary_key                  = "abc123"
  secondary_key                = "def456"
  ttl                          = 120
  path_pattern                 = ".*"
  cipher_combination           = "$uri$ourkey$time$args{rand}$args{uid}"
  cipher_param                 = "auth_key"
  time_param                   = "tname"
  time_format                  = "7s"
  request_url_style            = "https://$domain/Suri?$args&tname=$time&auth_key=$key"
  dst_style                    = 1
  encrypt_method               = "md5sum"
  log_format                   = false
  ignore_uri_slash             = false
  ignore_key_and_time_position = false
}
