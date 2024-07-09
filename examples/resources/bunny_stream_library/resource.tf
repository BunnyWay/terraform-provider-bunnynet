resource "bunny_stream_library" "example" {
  name = "test"

  # player
  player_language                  = "en"
  player_font_family               = "Rubik"
  player_primary_color             = "#ff7755"
  player_controls                  = ["airplay", "play", "volume", "captions", "current-time", "fullscreen", "mute", "pip", "play-large", "progress", "settings"]
  player_captions_font_color       = "#fff"
  player_captions_font_size        = 20
  player_captions_background_color = "#000"
  player_custom_head               = ""
  player_watchtime_heatmap_enabled = false

  # advertising
  vast_tag_url = ""

  # encoding
  original_files_keep     = true
  early_play_enabled      = false
  content_tagging_enabled = true
  mp4_fallback_enabled    = true
  resolutions             = ["240p", "360p", "480p", "720p", "1080p"]
  bitrate_240p            = 600
  bitrate_360p            = 800
  bitrate_480p            = 1400
  bitrate_720p            = 2800
  bitrate_1080p           = 5000
  bitrate_1440p           = 8000
  bitrate_2160p           = 25000
  watermark_position_left = 0
  watermark_position_top  = 0
  watermark_width         = 200
  watermark_height        = 50

  # transcribing
  transcribing_enabled                   = false
  transcribing_smart_title_enabled       = false
  transcribing_smart_description_enabled = false
  transcribing_languages                 = ["en"]

  # security
  direct_play_enabled                = true
  referers_allowed                   = ["*.example.com", "example.com"]
  referers_blocked                   = ["example.org", "*.example.org"]
  direct_url_file_access_blocked     = true
  view_token_authentication_required = true
  cdn_token_authentication_required  = true
  drm_mediacage_basic_enabled        = true

  # api
  webhook_url = ""
}
