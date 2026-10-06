package ianatlds

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"
)

// ─── TLDsList ────────────────────────────────────────────────────────────────

func TestTLDsList(t *testing.T) {
	if len(TLDsList) <= 1000 {
		t.Fatalf("expected > 1000 TLDs, got %d", len(TLDsList))
	}
	for _, want := range []string{"com", "org", "net", "edu", "gov"} {
		if !slices.Contains(TLDsList, want) {
			t.Errorf("missing common TLD %q", want)
		}
	}
	seen := make(map[string]bool, len(TLDsList))
	for _, tld := range TLDsList {
		if tld != strings.ToLower(tld) {
			t.Errorf("TLD %q is not lowercase", tld)
		}
		if seen[tld] {
			t.Errorf("duplicate TLD %q", tld)
		}
		seen[tld] = true
	}
}

// ─── LastUpdated ─────────────────────────────────────────────────────────────

func TestLastUpdated(t *testing.T) {
	if LastUpdated == "" {
		t.Fatal("LastUpdated is empty")
	}
	parsed, err := time.Parse(time.RFC3339, LastUpdated)
	if err != nil {
		t.Fatalf("not ISO 8601: %v", err)
	}
	if got := parsed.UTC().Format("2006-01-02T15:04:05.000Z"); got != LastUpdated {
		t.Errorf("does not round-trip: %q != %q", got, LastUpdated)
	}
	if parsed.Year() < 2024 {
		t.Errorf("year %d < 2024", parsed.Year())
	}
}

// ─── ValidateEmail ───────────────────────────────────────────────────────────

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  EmailValidation
	}{
		// valid
		{"standard .com", "user@example.com", EmailValidation{Valid: true, Email: "user@example.com", TLD: "com"}},
		{".org", "contact@nonprofit.org", EmailValidation{Valid: true, Email: "contact@nonprofit.org", TLD: "org"}},
		{"ccTLD", "joel@yarigai.mx", EmailValidation{Valid: true, Email: "joel@yarigai.mx", TLD: "mx"}},
		{"uppercase TLD normalised", "user@example.COM", EmailValidation{Valid: true, Email: "user@example.COM", TLD: "com"}},
		{"subdomains", "user@mail.subdomain.example.co.uk", EmailValidation{Valid: true, Email: "user@mail.subdomain.example.co.uk", TLD: "uk"}},
		{"last @ is the boundary (RFC 5321)", "user@host@example.com", EmailValidation{Valid: true, Email: "user@host@example.com", TLD: "com"}},
		{"IDN TLD (xn--p1ai)", "user@example.xn--p1ai", EmailValidation{Valid: true, Email: "user@example.xn--p1ai", TLD: "xn--p1ai"}},
		// unknown_tld
		{"made-up TLD", "user@example.xyzzy", EmailValidation{Email: "user@example.xyzzy", TLD: "xyzzy", Reason: ReasonUnknownTLD}},
		{"parsed TLD included", "a@b.notreal", EmailValidation{Email: "a@b.notreal", TLD: "notreal", Reason: ReasonUnknownTLD}},
		// invalid_format
		{"no @", "not-an-email", EmailValidation{Email: "not-an-email", Reason: ReasonInvalidFormat}},
		{"starts with @", "@domain.com", EmailValidation{Email: "@domain.com", Reason: ReasonInvalidFormat}},
		{"ends with @", "user@", EmailValidation{Email: "user@", Reason: ReasonInvalidFormat}},
		{"domain with no dot", "user@localhost", EmailValidation{Email: "user@localhost", Reason: ReasonInvalidFormat}},
		{"domain starts with dot", "user@.com", EmailValidation{Email: "user@.com", Reason: ReasonInvalidFormat}},
		{"domain ends with dot", "user@example.", EmailValidation{Email: "user@example.", Reason: ReasonInvalidFormat}},
		{"echoes original email", "bad-input", EmailValidation{Email: "bad-input", Reason: ReasonInvalidFormat}},
		{"empty", "", EmailValidation{Email: "", Reason: ReasonInvalidFormat}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidateEmail(tc.email); got != tc.want {
				t.Errorf("ValidateEmail(%q) = %+v, want %+v", tc.email, got, tc.want)
			}
		})
	}
}

// The JSON shape is the cross-language contract with @yarigai/iana-tlds.
func TestValidateEmailJSON(t *testing.T) {
	tests := map[string]string{
		"user@example.com":   `{"valid":true,"email":"user@example.com","tld":"com"}`,
		"user@example.xyzzy": `{"valid":false,"email":"user@example.xyzzy","tld":"xyzzy","reason":"unknown_tld"}`,
		"not-an-email":       `{"valid":false,"email":"not-an-email","tld":null,"reason":"invalid_format"}`,
	}
	for email, want := range tests {
		got, err := json.Marshal(ValidateEmail(email))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("json(%q) = %s, want %s", email, got, want)
		}
		var back EmailValidation
		if err := json.Unmarshal(got, &back); err != nil || back != ValidateEmail(email) {
			t.Errorf("round-trip %q = %+v, %v", email, back, err)
		}
	}
}

// ─── IsValidEmail ────────────────────────────────────────────────────────────

func TestIsValidEmail(t *testing.T) {
	for email, want := range map[string]bool{
		"user@example.com":   true,
		"user@example.xyzzy": false,
		"not-an-email":       false,
		"@":                  false,
		"":                   false,
	} {
		if got := IsValidEmail(email); got != want {
			t.Errorf("IsValidEmail(%q) = %v, want %v", email, got, want)
		}
	}
	for _, email := range []string{"user@example.com", "bad@fake.xyzzy", "notanemail", "user@EXAMPLE.ORG", "a@b."} {
		if IsValidEmail(email) != ValidateEmail(email).Valid {
			t.Errorf("IsValidEmail(%q) inconsistent with ValidateEmail", email)
		}
	}
}

func BenchmarkValidateEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ValidateEmail("user@example.com")
	}
}
