# Changelog

## v0.1.1 - 2026-11-10 (security release)

Upgrading is recommended for all users, especially if the exporter is
reachable from other hosts.

### Security
- **Authentication was not enforced with a proper status**: a missing or wrong
  bearer token returned `200 OK` with a JSON error. `/metrics` now answers
  `401 Unauthorized` with a `WWW-Authenticate` header and a generic message.
- **Timing attack on the bearer token**: the token is now compared in constant time.
- **Request forgery through the `target` parameter (SSRF)**: a target such as
  `host/path?query` or `user@host` was turned into an arbitrary HTTPS `GET`,
  and redirects were followed to any other host. Targets are now strictly
  validated (`host`, `host:port`, `[ipv6]:port`) and the exporter only performs
  a TLS handshake: no HTTP request is sent, nothing is redirected.
- **Metric injection**: label values (the `target` parameter and the
  certificate common name, both attacker controlled) were written unescaped,
  allowing forged samples in the scrape output. They are now escaped.
- **Denial of service**:
  - the HTTP server had no timeouts (slowloris); read/header/idle timeouts and
    a header size limit are now set;
  - the number of targets per request was unbounded; it is now limited to 100
    and each target to 260 characters;
  - checks kept running after the scraper disconnected; they are now cancelled
    with the request;
  - `client.timeout: 0` meant "no timeout"; it must now be between 1 and 300 s.
- **Vulnerable Go toolchain and dependencies**: Go 1.26.4 -> 1.26.10 (fixes 12
  reachable standard library vulnerabilities in `crypto/tls`, `crypto/x509`,
  `net/http`, ...), `golang.org/x/text` 0.40.0 -> 0.44.0 (GO-2026-6629).
  `govulncheck` now reports no vulnerability.
- **Docker image**: runs as an unprivileged user instead of root, base image
  pinned (`alpine:3.22` instead of `latest`).
- **docker-compose**: no more default `changeme` token (the token must be
  provided), read-only filesystem, all capabilities dropped, `no-new-privileges`.
- **systemd unit**: sandboxing options added, `CAP_NET_BIND_SERVICE` removed.
- A warning is logged at startup when no bearer token is configured on a
  non-loopback address, or when the token is still `changeme`.

### Fixed
- The command line flags (`--listen`, `--port`, `--timeout`, `--bearer`,
  `--log_level`) were silently ignored.
- `log_level` was ignored (the logger was never configured).
- Invalid exposition format: the `# HELP` and `# TYPE` lines were written on
  the same line.
- Duplicate targets produced duplicate series, making Prometheus reject the scrape.
- Startup errors (e.g. port already in use) were swallowed and the process
  exited with code 0; graceful shutdown did not wait for in-flight requests.
- IPv6 listen addresses were not supported.
- The Docker image could not be built (wrong package path).
- systemd unit: wrong `[Install]` section (`RequiredBy=network.target`).
- `go.mod`: direct dependencies marked as indirect, unused dependencies removed.
- README: Prometheus configuration example had an invalid YAML indentation.

### Changed
- Targets that do not speak HTTP (LDAPS, SMTPS, ...) can now be checked, since
  only a TLS handshake is performed.
- A request without `target`, or with an invalid one, now returns `400 Bad Request`.
- Label order in the output is now stable (`domain`, then `target`).
- The configuration is validated at startup (port, timeout, log level).
- Unit tests added (`make test`).

### Upgrade notes
- Monitoring that relied on `200` for failed authentication must expect `401`.
- systemd: if you listen on a port below 1024, re-add
  `AmbientCapabilities=CAP_NET_BIND_SERVICE` and
  `CapabilityBoundingSet=CAP_NET_BIND_SERVICE` to the unit.
- docker-compose: define `SSL_EXPORTER_LISTEN_BEARER` (e.g. in a `.env` file).
