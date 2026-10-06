// Package ianatlds validates email TLDs against the official IANA list,
// bundled at build time and synced daily. Zero dependencies, no network
// calls at runtime.
//
// Go port of the npm package @yarigai/iana-tlds; both expose the same
// behavior and the same JSON shape for validation results.
package ianatlds

//go:generate go run ./internal/updatetlds

import (
	"encoding/json"
	"strings"
)

// EmailValidationReason explains why a TLD validation failed.
type EmailValidationReason string

const (
	// ReasonInvalidFormat means no TLD could be extracted from the email.
	ReasonInvalidFormat EmailValidationReason = "invalid_format"
	// ReasonUnknownTLD means a TLD was parsed but is not in the IANA list.
	ReasonUnknownTLD EmailValidationReason = "unknown_tld"
)

// EmailValidation is the result returned by [ValidateEmail].
//
//   - Valid == true  → TLD is the matched IANA TLD and Reason is empty.
//   - Valid == false → Reason explains the failure; TLD is "" when it could
//     not be parsed (ReasonInvalidFormat), or the parsed TLD when it was not
//     recognised (ReasonUnknownTLD).
//
// Marshals to the same JSON as the JS package: "tld" is null when not
// parsed, and "reason" is omitted when valid.
type EmailValidation struct {
	Valid  bool                  `json:"valid"`
	Email  string                `json:"email"`
	TLD    string                `json:"tld"`
	Reason EmailValidationReason `json:"reason,omitempty"`
}

// MarshalJSON emits "tld": null when the TLD could not be parsed.
func (v EmailValidation) MarshalJSON() ([]byte, error) {
	var tld *string
	if v.TLD != "" {
		tld = &v.TLD
	}
	// Same key order as the JS object: valid, email, tld, reason.
	return json.Marshal(struct {
		Valid  bool                  `json:"valid"`
		Email  string                `json:"email"`
		TLD    *string               `json:"tld"`
		Reason EmailValidationReason `json:"reason,omitempty"`
	}{v.Valid, v.Email, tld, v.Reason})
}

// TLDsList holds all valid TLDs registered with IANA, in lowercase.
// Mutating it does not affect validation.
var TLDsList = strings.Fields(rawTLDs)

var tldsSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(TLDsList))
	for _, t := range TLDsList {
		m[t] = struct{}{}
	}
	return m
}()

// ValidateEmail validates whether an email address has a valid
// IANA-registered TLD.
//
// Only the TLD portion of the domain is checked — local-part syntax and DNS
// resolution are outside the scope of this function.
func ValidateEmail(email string) EmailValidation {
	lastAt := strings.LastIndexByte(email, '@')
	if lastAt <= 0 || lastAt == len(email)-1 {
		return EmailValidation{Email: email, Reason: ReasonInvalidFormat}
	}
	domain := email[lastAt+1:]
	lastDot := strings.LastIndexByte(domain, '.')
	if lastDot <= 0 || lastDot == len(domain)-1 {
		return EmailValidation{Email: email, Reason: ReasonInvalidFormat}
	}
	tld := strings.ToLower(domain[lastDot+1:])
	if _, ok := tldsSet[tld]; ok {
		return EmailValidation{Valid: true, Email: email, TLD: tld}
	}
	return EmailValidation{Email: email, TLD: tld, Reason: ReasonUnknownTLD}
}

// IsValidEmail reports whether the email address has a valid IANA-registered
// TLD.
func IsValidEmail(email string) bool {
	return ValidateEmail(email).Valid
}
