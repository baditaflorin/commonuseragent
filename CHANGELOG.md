# Changelog

All notable changes to this project are documented here.

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
