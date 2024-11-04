resource "st-cdnetworks_anti_hotlinking_config" "test" {
  domain_id = st-cdnetworks_shield_domain.test.domain_id

  ip_control_rule {
    directory      = "/abc/"
    control_action = "302"
    rewrite_to     = "aaa"
    allowed_ips    = ["1.1.1.1", "1.1.1.2", "1.1.1.3"]
  }

  referer_control_rule {
    directory          = "/abc/"
    control_action     = "302"
    rewrite_to         = "aaa"
    valid_urls         = ["https://1.1.1.1", "https://2.2.2.2"]
    allow_null_referer = true
  }

  ua_control_rule {
    directory         = "/abc/"
    control_action    = "302"
    rewrite_to        = "aaa"
    valid_user_agents = ["a|b", "c|d"]
  }
}
