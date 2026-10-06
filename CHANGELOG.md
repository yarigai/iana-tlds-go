# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-10-06

### Added

- Initial release, Go port of [`@yarigai/iana-tlds`](https://github.com/yarigai/iana-tlds) 1.1.1 with the same behavior.
- `TLDsList` - all IANA-registered TLDs in lowercase.
- `LastUpdated` - ISO 8601 timestamp of the IANA source file bundled in this release.
- `ValidateEmail(email)` - returns an `EmailValidation` that marshals to the same JSON as the JS result.
- `IsValidEmail(email)` - boolean convenience wrapper.
- `ReasonInvalidFormat` / `ReasonUnknownTLD` typed constants.
- Automated daily sync pipeline via GitHub Actions that tags a patch release when the IANA list changes.

[Unreleased]: https://github.com/yarigai/iana-tlds-go/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/yarigai/iana-tlds-go/releases/tag/v1.0.0
