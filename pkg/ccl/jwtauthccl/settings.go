// Copyright 2022 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://github.com/cockroachdb/cockroach/blob/master/licenses/CCL.txt

package jwtauthccl

import (
	"bytes"
	"encoding/json"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/errors"
	"github.com/lestrrat-go/jwx/jwk"
)

// All cluster settings necessary for the JWT authentication feature.
const (
	baseJWTAuthSettingName     = "server.jwt_authentication."
	JWTAuthAudienceSettingName = baseJWTAuthSettingName + "audience"
	JWTAuthEnabledSettingName  = baseJWTAuthSettingName + "enabled"
	JWTAuthIssuersSettingName  = baseJWTAuthSettingName + "issuers"
	JWTAuthJWKSSettingName     = baseJWTAuthSettingName + "jwks"
)

// JWTAuthAudience sets accepted audience values for JWT logins over the SQL interface.
var JWTAuthAudience = func() *settings.StringSetting {
	s := settings.RegisterStringSetting(
		settings.TenantReadOnly,
		JWTAuthAudienceSettingName,
		"sets accepted audience values for JWT logins over the SQL interface",
		"",
	)
	return s
}()

// JWTAuthEnabled enables or disabled JWT login over the SQL interface.
var JWTAuthEnabled = func() *settings.BoolSetting {
	s := settings.RegisterBoolSetting(
		settings.TenantReadOnly,
		JWTAuthEnabledSettingName,
		"enables or disabled JWT login for the SQL interface",
		false,
	)
	s.SetReportable(true)
	return s
}()

// JWTAuthJWKS is the public key set for JWT logins over the SQL interface.
var JWTAuthJWKS = func() *settings.StringSetting {
	s := settings.RegisterValidatedStringSetting(
		settings.TenantReadOnly,
		JWTAuthJWKSSettingName,
		"sets the public key set for JWT logins over the SQL interface (JWKS format)",
		"{\"keys\":[]}",
		validateJWTAuthJWKS,
	)
	return s
}()

// JWTAuthIssuers is the list of "issuer" values that are accepted for JWT logins over the SQL interface.
var JWTAuthIssuers = func() *settings.StringSetting {
	s := settings.RegisterValidatedStringSetting(
		settings.TenantReadOnly,
		JWTAuthIssuersSettingName,
		"sets accepted issuer values for JWT logins over the SQL interface either as a string or as a JSON "+
			"string with an array of issuer strings in it",
		"",
		validateJWTAuthIssuers,
	)
	return s
}()

func validateJWTAuthIssuers(values *settings.Values, s string) error {
	var issuers []string

	var jsonCheck json.RawMessage
	if json.Unmarshal([]byte(s), &jsonCheck) != nil {
		// If we know the string is *not* valid JSON, fall back to assuming basic
		// string to use a single valid issuer
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader([]byte(s)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&issuers); err != nil {
		return errors.Wrap(err, "JWT authentication issuers JSON not valid")
	}
	return nil
}

func validateJWTAuthJWKS(values *settings.Values, s string) error {
	if _, err := jwk.Parse([]byte(s)); err != nil {
		return errors.Wrap(err, "JWT authentication JWKS not a valid JWKS")
	}
	return nil
}

func mustParseIssuers(issuersJSON string) []string {
	var issuersConf []string

	var jsonCheck json.RawMessage
	if json.Unmarshal([]byte(issuersJSON), &jsonCheck) != nil {
		// If we know the string is *not* valid JSON, fall back to assuming basic
		// string to use a single valid issuer.
		return []string{issuersJSON}
	}

	decoder := json.NewDecoder(bytes.NewReader([]byte(issuersJSON)))
	if err := decoder.Decode(&issuersConf); err != nil {
		return []string{issuersJSON}
	}
	return issuersConf
}

func mustParseJWKS(jwks string) jwk.Set {
	keySet, err := jwk.Parse([]byte(jwks))
	if err != nil {
		return jwk.NewSet()
	}
	return keySet
}
