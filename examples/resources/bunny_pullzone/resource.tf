resource "bunny_pullzone" "test" {
  name = "test"

  origin {
    type                = "OriginUrl"
    url                 = "https://192.0.2.1"
    follow_redirects    = true
    forward_host_header = true
    verify_ssl          = true
  }

  routing {
    tier    = "Standard"
    zones   = ["US", "EU", "ASIA"]
    filters = ["all"]
  }

  cors_enabled    = true
  cors_extensions = ["eot", "ttf", "woff", "woff2", "css"]

  optimizer_enabled                     = true
  optimizer_webp                        = true
  optimizer_minify_css                  = true
  optimizer_minify_js                   = true
  optimizer_smartimage                  = true
  optimizer_smartimage_desktop_maxwidth = 1600
  optimizer_smartimage_desktop_quality  = 85
  optimizer_smartimage_mobile_maxwidth  = 800
  optimizer_smartimage_mobile_quality   = 70
  optimizer_watermark                   = false
  optimizer_watermark_url               = "https://bunny.net/icons/icon-72x72.png"
  optimizer_watermark_position          = "BottomRight"
  optimizer_classes_force               = false

  limit_download_speed = 10000
  limit_requests       = 0
  limit_after          = 0
  limit_burst          = 0
  limit_connections    = 0
  limit_bandwidth      = 1000000000

  safehop_enabled            = true
  safehop_retry_count        = 1
  safehop_retry_delay        = 5
  safehop_retry_reasons      = ["connectionTimeout", "5xxResponse", "responseTimeout"]
  safehop_connection_timeout = 10
  safehop_response_timeout   = 60

  cache_enabled                 = true
  cache_expiration_time         = -1
  cache_expiration_time_browser = -1
  sort_querystring              = true
  cache_errors                  = false
  cache_vary = [
    "querystring",
    "webp",
    "country",
    "hostname",
    "mobile",
    "avif",
    "cookie",
  ]
  cache_vary_querystring = ["ver", "q"]
  cache_vary_cookie      = ["JSESSIONID"]
  strip_cookies          = true
  cache_chunked          = false
  cache_stale = [
    "offline",
    "updating",
  ]

  permacache_storagezone = bunny_storage_zone.test.id

  originshield_enabled              = true
  originshield_concurrency_limit    = true
  originshield_concurrency_requests = 200
  originshield_queue_requests       = 5000
  originshield_queue_wait           = 45

  request_coalescing_enabled = true
  request_coalescing_timeout = 15

  block_root_path     = true
  block_post_requests = true
  allow_referers      = ["*.example.com", "example.com"]
  block_referers      = ["example.org", "*.example.org"]
  block_ips           = ["1.1.1.1", "192.0.2.*"]

  log_enabled          = true
  log_anonymized       = true
  log_anonymized_style = "Drop"
  log_forward_enabled  = true
  log_forward_server   = "192.0.2.254"
  log_forward_port     = 1234
  log_forward_token    = "my-log-secret"
  log_forward_protocol = "udp|tcp|tcp_encrypted|datadog"
  log_forward_format   = "json|plain"
  log_storage_enabled  = true
  log_storage_zone     = bunny_storage_zone.logs.id

  tls_support = ["TLSv1.0", "TLSv1.1"]

  errorpage_whitelabel         = true
  errorpage_statuspage_enabled = true
  errorpage_statuspage_code    = "abc1234d"
  errorpage_custom_enabled     = true
  errorpage_custom_content     = "error {{status_code}}"

  s3_auth_enabled = true
  s3_auth_key     = "key-goes-here"
  s3_auth_secret  = "secret-goes-here"
  s3_auth_region  = "us-east-1"

  token_auth_enabled       = true
  token_auth_ip_validation = true

  blocked_countries    = ["KP"]
  redirected_countries = ["CN"]
}
