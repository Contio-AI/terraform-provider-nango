terraform {
  required_providers {
    nango = {
      source = "contio.ai/contio/nango"
    }
  }
}

provider "nango" {
  environment_key = "96e7de6c-4286-4a62-9eab-baf08f9ac849"
}

resource "nango_integration" "platform-google" {
  unique_key     = "platform-google"
  display_name   = "Google"
  nango_provider = "google"
  credentials = {
    client_id     = var.google_client_id
    client_secret = var.google_client_secret
    type          = "OAUTH2"
    scopes = [
      "https://www.googleapis.com/auth/gmail.readonly",
      "https://www.googleapis.com/auth/drive.readonly",
      "https://www.googleapis.com/auth/calendar.readonly",
      "https://www.googleapis.com/auth/documents",
      "https://www.googleapis.com/auth/userinfo.profile",
      "https://www.googleapis.com/auth/userinfo.email",
    ]
  }
}

resource "nango_integration" "service-google-calendar" {
  unique_key     = "service-google-calendar"
  display_name   = "Calendar"
  nango_provider = "google-calendar"
  credentials = {
    client_id     = var.google_client_id
    client_secret = var.google_client_secret
    type          = "OAUTH2"
    scopes = [
      "https://www.googleapis.com/auth/calendar.readonly",
      "https://www.googleapis.com/auth/userinfo.profile",
      "https://www.googleapis.com/auth/userinfo.email"
    ]
  }
}



resource "nango_integration" "service-google-drive" {
  unique_key     = "service-google-drive"
  display_name   = "Drive"
  nango_provider = "google-drive"
  credentials = {
    client_id     = var.google_client_id
    client_secret = var.google_client_secret
    type          = "OAUTH2"
    scopes = [
      "https://www.googleapis.com/auth/drive.readonly",
      "https://www.googleapis.com/auth/userinfo.email",
      "https://www.googleapis.com/auth/userinfo.profile",
      "https://www.googleapis.com/auth/documents",

    ]
  }
}

resource "nango_integration" "service-google-gmail" {
  unique_key     = "service-google-gmail"
  display_name   = "Gmail"
  nango_provider = "google-mail"
  credentials = {
    client_id     = var.google_client_id
    client_secret = var.google_client_secret
    type          = "OAUTH2"
    scopes = [
      "https://www.googleapis.com/auth/gmail.readonly",
      "https://www.googleapis.com/auth/userinfo.profile",
      "https://www.googleapis.com/auth/userinfo.email",

    ]
  }
}

resource "nango_integration" "platform-microsoft" {
  unique_key     = "platform-microsoft"
  display_name   = "Microsoft"
  nango_provider = "microsoft"
  credentials = {
    client_id     = var.microsoft_client_id
    client_secret = var.microsoft_client_secret
    type          = "OAUTH2"
    scopes = [
      "Calendars.Read",
      "Files.Read.All",
      "offline_access",
      "Mail.Read",
      "User.Read"
    ]
  }
}

resource "nango_integration" "service-outlook" {
  unique_key     = "service-outlook"
  display_name   = "Calendar"
  nango_provider = "outlook"
  credentials = {
    client_id     = var.microsoft_client_id
    client_secret = var.microsoft_client_secret
    type          = "OAUTH2"
    scopes = [
      "Calendars.Read",
      "offline_access",
      "Mail.Read",
      "User.Read"
    ]
  }
}

resource "nango_integration" "service-sharepoint-online" {
  unique_key     = "service-sharepoint-online"
  display_name   = "SharePoint Online (v2)"
  nango_provider = "sharepoint-online"
  credentials = {
    client_id     = var.microsoft_client_id
    client_secret = var.microsoft_client_secret
    type          = "OAUTH2"
    scopes = [
      "Files.Read.All",
      "offline_access",
      "User.Read"
    ]
  }
}

data "nango_integrations" "contio" {}

output "google_platform" {
  sensitive = true
  value = resource.nango_integration.platform-google
}

output "service_google_calendar" {
  sensitive = true
  value = resource.nango_integration.service-google-calendar
}

output "service_google_drive" {
  sensitive = true
  value = resource.nango_integration.service-google-drive
}

output "service_google_gmail" {
  sensitive = true
  value = resource.nango_integration.service-google-gmail
}

output "microsoft_platform" {
  sensitive = true
  value = resource.nango_integration.platform-microsoft
}
