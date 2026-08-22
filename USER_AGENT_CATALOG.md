# User-agent catalog policy

The embedded catalog contains current, representative browser user-agent
strings for application testing. It is not a device fingerprint database and
must not be used to impersonate real users or evade a service's access rules.

## Update policy

- Review the catalog at least once per browser release cycle, and record the
  refresh date in `CHANGELOG.md`.
- Use stable browser releases and the current user-agent grammar. Chrome's
  reduced Android string intentionally reports `Android 10; K` and a frozen
  `x.0.0.0` version; do not replace it with an invented device model or build.
- Keep the `pct` values as relative selection weights within each catalog.
  They are rounded, representative weights based on worldwide usage rather
  than a claim of exact telemetry.
- Include the principal browser families on their relevant platforms. Do not
  add crawlers, libraries, or bots to these end-user catalogs.
- Do not create a distinct Brave entry: Brave intentionally uses Chromium's
  User-Agent on most platforms. Likewise, Vivaldi identifies itself only for
  a small allowlist of partner sites, so it is not representative of ordinary
  browser traffic.

## Sources for the August 2026 refresh

- [Chrome stable release notes](https://chromereleases.googleblog.com/2026/08/stable-channel-update-for-desktop.html)
- [MDN user-agent reference](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/User-Agent)
- [Apple Safari 26.6 release notes](https://developer.apple.com/documentation/safari-release-notes/safari-26_6-release-notes)
- [Mozilla Firefox 152 announcement](https://blog.mozilla.org/en/firefox/firefox-roadmap-152/)
- [StatCounter worldwide browser share, July 2026](https://gs.statcounter.com/browser-market-share/all-worldwide)
- [Brave browser detection policy](https://github.com/brave/brave-browser/wiki/Detecting-Brave-%28for-Websites%29)
- [Vivaldi User-Agent policy](https://help.vivaldi.com/developers/web/vivaldi-user-agent-and-client-hints-user-agent/)

For servers that need browser, platform, or architecture details beyond the
reduced User-Agent header, prefer User-Agent Client Hints where supported.
