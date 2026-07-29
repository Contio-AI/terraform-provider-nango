## 0.1.0 (Unreleased)

BUG FIXES:

* `nango_integration`: `Read()` now refreshes `credentials.client_id` and
  `credentials.client_secret` from the Nango API, so plans surface credential
  drift between Terraform and Nango. Previously state always echoed the
  last-applied values, and because updates send the full credentials object,
  any in-place change silently rewrote a hand-corrected credential from a
  stale variable (CON-6127). Environments where the dashboard values diverge
  from the Terraform variables will show a one-time credentials diff — fix the
  variable/secret rather than applying blindly.
* `nango_integration`: `credentials.client_secret` is now marked sensitive, so
  drift diffs render as `(sensitive value)` instead of printing the secret in
  plan output. `client_id` intentionally stays visible — it is public in every
  OAuth authorize URL, and seeing the actual value in a drift diff is what
  makes credential incidents diagnosable.

FEATURES:
