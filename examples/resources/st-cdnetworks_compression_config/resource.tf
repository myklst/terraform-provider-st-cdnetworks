resource "st-cdnetworks_compression_config" "test" {
  domain_id = st-cdnetworks_flood_shield_domain.test.domain_id

  compression_settings = {
    compression_enabled = true
    br_types            = false
    ignore_letter_case  = true
    file_type_others    = ["jpg", "txt", "png"]
    custom_file_types   = ["html"]
  }
}
