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
}
