## 0.2.0 (Unreleased)

ENHANCEMENTS:

* `nango_integration`: `credentials` is now **optional**. Nango's integration
  API accepts a credentials object only for the `OAUTH1`, `OAUTH2`, `TBA`,
  `APP` and `CUSTOM` auth modes; `BASIC` and `API_KEY` integrations hold no
  integration-level secret at all, because those credentials are supplied
  per-connection. Previously the attribute was `Required` and marshalled
  without `omitempty`, so the provider always sent
  `{"client_id":"","client_secret":"","type":"BASIC"}`, which Nango rejects.
  Omitting the block now sends no `credentials` key, making credential-less
  integrations expressible (CON-7018). Widening `Required` to `Optional` is
  additive: existing OAuth2 configurations marshal byte-identically and need
  no state migration.
* `nango_integration`: `Create` and `Update` no longer dereference a nil
  `credentials` block. `integrationModel.Credentials` is a pointer, so both
  paths previously panicked when the block was omitted.

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
