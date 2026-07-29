## 0.1.0 (Unreleased)

BUG FIXES:

* `nango_integration`: `Read()` now refreshes `credentials.client_id` from the
  Nango API, so plans surface credential drift between Terraform and Nango.
  Previously state always echoed the last-applied value, and because updates
  send the full credentials object, any in-place change silently rewrote a
  hand-corrected credential from a stale variable (CON-6127). Environments
  where the dashboard value diverges from the Terraform variable will show a
  one-time `~ client_id` diff — fix the variable/secret rather than applying
  blindly.

FEATURES:
