# Changelog

All notable changes to this project are documented here.

## Unreleased

### Fixed

- Return a JSON `503 Service Unavailable` response from monitoring endpoints
  when request logging is unavailable, rather than panicking.
- Return a JSON rate-limit response with `X-Content-Type-Options: nosniff`.

### Testing

- Add API-handler coverage for selection, request logging, unavailable
  monitoring dependencies, invalid limits, and rate-limit responses.

## v0.2.1 - 2026-08-22

### Documentation

- Align the README's release reference with the repository's versioned release
  series and changelog.

## v0.2.0 - 2026-08-22

### Changed

- Refreshed the desktop and mobile browser catalog for August 2026.
- Rebalanced catalog entries against current worldwide browser usage and added
  current Chrome, Edge, Firefox, Safari, Samsung Internet, Opera, and mobile
  Edge representations.
- Replaced the stale GitHub Actions workflow with Woodpecker CI, matching the
  module's Go 1.25 toolchain requirement.

### Fixed

- Reject non-positive, non-finite, and duplicate user-agent weights during
  catalog validation, preventing invalid weighted selections.

### Documentation

- Added the catalog maintenance policy and source links.
