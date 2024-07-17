resource "bunnynet_stream_library" "example" {
  name = "example"

  player_controls                  = ["airplay", "play", "volume", "captions", "current-time", "fullscreen", "mute", "pip", "play-large", "progress", "settings"]
  player_captions_font_color       = "#000"
  player_captions_font_size        = 20
  player_captions_background_color = "#fff"
  resolutions                      = ["240p", "360p", "480p", "720p", "1080p"]
  transcribing_enabled             = true
  transcribing_languages           = ["en", "fr", "es", "de"]
  direct_play_enabled              = true
}
