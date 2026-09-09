// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"testing"
)

func strPtr(s string) *string { return &s }

// Nango's public integration API types `credentials` as optional and its type
// union admits only OAUTH1 | OAUTH2 | TBA | APP | CUSTOM. Auth modes such as
// BASIC and API_KEY carry no integration-level credentials at all — the secrets
// are per-connection — so for those the block must be omitted entirely.
// Emitting {"client_id":"","client_secret":"","type":"BASIC"} fails that union.
func TestIntegrationRequestMarshalOmitsCredentialsWhenNil(t *testing.T) {
	request := integrationRequestModel{
		UniqueKey:     strPtr("service-agencybloc"),
		DisplayName:   "AgencyBloc",
		NangoProvider: strPtr("private-api-basic"),
		Credentials:   nil,
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if _, present := decoded["credentials"]; present {
		t.Errorf("credentials key must be absent for a credential-less integration, got %s", body)
	}
	if decoded["display_name"] != "AgencyBloc" {
		t.Errorf("display_name = %v, want AgencyBloc", decoded["display_name"])
	}
	if decoded["provider"] != "private-api-basic" {
		t.Errorf("provider = %v, want private-api-basic", decoded["provider"])
	}
}

// Backward compatibility: every existing OAuth2 resource must marshal exactly
// as it did before credentials became optional.
func TestIntegrationRequestMarshalUnchangedForOAuth2(t *testing.T) {
	request := integrationRequestModel{
		UniqueKey:     strPtr("service-salesforce"),
		DisplayName:   "Salesforce",
		NangoProvider: strPtr("salesforce"),
		Credentials: &integrationCredentialsRequestModel{
			ClientId:     "abc",
			ClientSecret: "shh",
			Type:         "OAUTH2",
			Scopes:       "read,write",
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	want := `{"unique_key":"service-salesforce","display_name":"Salesforce","provider":"salesforce","credentials":{"client_id":"abc","client_secret":"shh","type":"OAUTH2","scopes":"read,write"}}`
	if string(body) != want {
		t.Errorf("marshalled body diverged from the v0.1.0 shape:\n got: %s\nwant: %s", body, want)
	}
}

// The Update path deliberately omits unique_key and provider — Nango reads
// unique_key in the body as a rename attempt.
func TestIntegrationUpdateRequestOmitsUniqueKeyAndProvider(t *testing.T) {
	request := integrationRequestModel{
		DisplayName: "AgencyBloc",
		Credentials: nil,
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	for _, key := range []string{"unique_key", "provider", "credentials"} {
		if _, present := decoded[key]; present {
			t.Errorf("%s must be absent from the update body, got %s", key, body)
		}
	}
}

// buildCredentialsRequest is the guard that keeps Create and Update from
// dereferencing a nil plan.Credentials. integrationModel.Credentials is a
// pointer, so it is nil whenever the credentials block is omitted.
func TestBuildCredentialsRequestHandlesNilPlanCredentials(t *testing.T) {
	if got := buildCredentialsRequest(nil, nil); got != nil {
		t.Errorf("buildCredentialsRequest(nil) = %+v, want nil", got)
	}
}
