resource "nango_integration" "google" {
  unique_key     = "google-oauth"
  display_name   = "Google OAuth"
  nango_provider = "google"

  credentials = {
    client_id     = var.google_client_id
    client_secret = var.google_client_secret
    type          = "OAUTH2"
    scopes = [
      "https://www.googleapis.com/auth/userinfo.email",
      "https://www.googleapis.com/auth/userinfo.profile"
    ]
  }
}
