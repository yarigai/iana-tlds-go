# iana-tlds-go

[![Go Reference](https://pkg.go.dev/badge/github.com/yarigai/iana-tlds-go.svg)](https://pkg.go.dev/github.com/yarigai/iana-tlds-go)
[![CI](https://github.com/yarigai/iana-tlds-go/actions/workflows/ci.yml/badge.svg)](https://github.com/yarigai/iana-tlds-go/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yarigai/iana-tlds-go)](https://goreportcard.com/report/github.com/yarigai/iana-tlds-go)
[![Zero dependencies](https://img.shields.io/badge/dependencies-0-brightgreen.svg)](go.mod)
[![TLD list updated](https://img.shields.io/github/last-commit/yarigai/iana-tlds-go/main?path=tlds_data.go&label=TLD%20list%20updated)](https://github.com/yarigai/iana-tlds-go/commits/main/tlds_data.go)
[![License: MIT](https://img.shields.io/github/license/yarigai/iana-tlds-go.svg)](LICENSE)

Validate email TLDs against the official IANA list, checked daily. Zero dependencies, standard library only, no network calls at runtime.

Go port of [`@yarigai/iana-tlds`](https://github.com/yarigai/iana-tlds) (npm). Both packages share the same behavior, the same bundled TLD list, and the same JSON shape for validation results, so a Go backend and a JS frontend agree on every email.

## Why

Most email and domain validators decide what a TLD is with a regex or a list someone copied years ago. Both fail in production. A pattern like `\.[a-z]{2,6}$` rejects real TLDs such as `.photography` or the internationalized `.xn--p1ai` (Russian `.рф`), and happily accepts typos like `.con` or `.comm`. Hardcoded lists go stale: IANA added `.web` in 2026, and it retires TLDs too (`.goo` is gone). This package checks the TLD against IANA's [root zone list](https://data.iana.org/TLD/tlds-alpha-by-domain.txt), compiled into your binary.

```go
ianatlds.IsValidEmail("ana@studio.photography") // true
ianatlds.IsValidEmail("ana@example.xn--p1ai")   // true
ianatlds.IsValidEmail("ana@example.con")        // false, unknown_tld
```

`ValidateEmail` does a single map lookup with zero allocations (~13 ns/op on Apple M-series).

This package only answers "is this TLD real?". Combine it with [`net/mail.ParseAddress`](https://pkg.go.dev/net/mail#ParseAddress) for full syntax checks. IDN TLDs are matched in Punycode form (`xn--p1ai`), so convert Unicode domains first with [`golang.org/x/net/idna`](https://pkg.go.dev/golang.org/x/net/idna) (`idna.ToASCII`).

## Installation

```bash
go get github.com/yarigai/iana-tlds-go@latest
```

Pin a version like any Go module (`go get github.com/yarigai/iana-tlds-go@v1.0.0`). Every release is a git tag that has passed CI.

Requires Go 1.21+.

## Usage

### Validate an email's TLD

```go
import ianatlds "github.com/yarigai/iana-tlds-go"

r := ianatlds.ValidateEmail(email)
if r.Valid {
	fmt.Println(r.TLD) // e.g. "com", "mx", "org"
} else {
	fmt.Println(r.Reason) // ianatlds.ReasonInvalidFormat | ianatlds.ReasonUnknownTLD
	fmt.Println(r.TLD)    // parsed TLD, or "" if the format was invalid
}

// Boolean shorthand
ianatlds.IsValidEmail(email)          // true  - TLD is IANA-registered
ianatlds.IsValidEmail(unknownTLDEmail) // false - TLD not in IANA list
ianatlds.IsValidEmail("not-an-email") // false - invalid format
```

### Return the result as JSON

`EmailValidation` marshals to exactly the same JSON as the JS package's result:

```go
json.Marshal(ianatlds.ValidateEmail("user@example.com"))
// {"valid":true,"email":"user@example.com","tld":"com"}

json.Marshal(ianatlds.ValidateEmail("user@example.xyzzy"))
// {"valid":false,"email":"user@example.xyzzy","tld":"xyzzy","reason":"unknown_tld"}

json.Marshal(ianatlds.ValidateEmail("not-an-email"))
// {"valid":false,"email":"not-an-email","tld":null,"reason":"invalid_format"}
```

### Work with the raw TLD list

```go
ianatlds.TLDsList    // []string of lowercase TLDs, e.g. ["aaa", "aarp", ..., "zw"]
ianatlds.LastUpdated // ISO 8601 timestamp of the IANA source file bundled in this release

slices.Contains(ianatlds.TLDsList, "com") // true
fmt.Println(ianatlds.LastUpdated)         // "2026-10-06T07:07:01.000Z"
```

## API Reference

### `func ValidateEmail(email string) EmailValidation`

Validates whether an email address has a valid IANA-registered TLD. Only the
TLD portion of the domain is checked - local-part syntax and DNS resolution
are outside the scope of this function. The last `@` is the boundary
(RFC 5321) and the TLD is lowercased.

| `Valid` | `TLD`         | `Reason`              | When                            |
| ------- | ------------- | --------------------- | ------------------------------- |
| `true`  | matched TLD   | `""`                  | TLD is recognised by IANA       |
| `false` | parsed TLD    | `ReasonUnknownTLD`    | TLD parsed but not in IANA list |
| `false` | `""` (`null`) | `ReasonInvalidFormat` | No TLD could be extracted       |

### `func IsValidEmail(email string) bool`

Convenience wrapper around `ValidateEmail` that returns `true` when the email
has a valid IANA TLD, `false` otherwise.

### Types

```go
type EmailValidationReason string

const (
	ReasonInvalidFormat EmailValidationReason = "invalid_format"
	ReasonUnknownTLD    EmailValidationReason = "unknown_tld"
)

type EmailValidation struct {
	Valid  bool                  `json:"valid"`
	Email  string                `json:"email"`
	TLD    string                `json:"tld"`              // null in JSON when ""
	Reason EmailValidationReason `json:"reason,omitempty"`
}
```

### Data exports

| Export        | Type       | Description                             |
| ------------- | ---------- | --------------------------------------- |
| `TLDsList`    | `[]string` | All valid TLDs in lowercase             |
| `LastUpdated` | `string`   | ISO 8601 timestamp from the IANA source |

## Data source & sync

The TLD list is sourced directly from IANA's root zone database:
`https://data.iana.org/TLD/tlds-alpha-by-domain.txt`

A scheduled CI pipeline runs daily. If IANA adds or removes a TLD, the
pipeline automatically:

1. Fetches the new list and compares it against the bundled version.
2. Runs `gofmt`, `go vet` and the test suite.
3. Commits, tags the next patch version, and pushes back to the repository.
4. Creates a GitHub release and publishes the tag to the Go module proxy.

To sync locally:

```bash
go generate ./...
go test ./...
```

## Contributing

Bug reports and pull requests are welcome at
[github.com/yarigai/iana-tlds-go](https://github.com/yarigai/iana-tlds-go).

Please open an issue first for non-trivial changes. Behavior changes must land
in both this repo and [`yarigai/iana-tlds`](https://github.com/yarigai/iana-tlds).

## License

MIT - see [LICENSE](LICENSE).

---

Maintained by [Yarigai](https://yarigai.mx) - Enterprise Technology Architecture & Strategic IT Consulting.
