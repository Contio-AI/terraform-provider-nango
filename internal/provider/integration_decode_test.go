// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"testing"
)

// TestIntegrationResponseDecodesClientId pins the read-path contract that
// credential drift detection depends on (CON-6127): the client_id returned by
// GET /integrations/{key}?include=credentials must survive deserialization.
// Before this field existed, Read() could never refresh client_id into state,
// so a divergence between Nango and Terraform was invisible in plans.
func TestIntegrationResponseDecodesClientId(t *testing.T) {
	body := `{"data":{"unique_key":"service-outlook","display_name":"Outlook","provider":"outlook",` +
		`"updated_at":"2026-07-29T00:00:00Z","credentials":{"type":"OAUTH2",` +
		`"client_id":"ae7862fc-486d-43a3-9fc8-bfe2ba1e10f2","client_secret":"s","scopes":"a,b"}}}`

	var resp nanogoIntegrationResponse2
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Data.Credentials == nil {
		t.Fatal("credentials not decoded")
	}
	if got, want := resp.Data.Credentials.ClientId, "ae7862fc-486d-43a3-9fc8-bfe2ba1e10f2"; got != want {
		t.Fatalf("client_id = %q, want %q", got, want)
	}
	if got, want := resp.Data.Credentials.Scopes, "a,b"; got != want {
		t.Fatalf("scopes = %q, want %q", got, want)
	}
}
