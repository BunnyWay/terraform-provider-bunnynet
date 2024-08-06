# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

> [!NOTE]
> While we strive to maintain backwards compatibility as much as possible, we can't guarantee semantic versioning will be strictly followed, as this provider depends on the underlying [bunny.net API](https://docs.bunny.net/reference/bunnynet-api-overview).

## [0.3.5] - 2024-08-06
### Added
- Import command example for all resources
### Changed
- resource pullzone_hostname: force_ssl can only be enabled when tls_enabled is true

## [0.3.4] - 2024-08-01
### Added
- resource pullzone_hostname: manage free TLS certificate
### Changed
- resource pullzone_hostname: import via name instead of ID

## [0.3.3] - 2024-07-30
### Added
- resource stream_library: multiple output audio track support
### Changed
- resource stream_library: run Create() in two passes

## [0.3.2] - 2024-07-22
### Added
- Run acceptance tests for every commit
### Changed
- Refactored code according to golangci-lint
- Improved documentation

## [0.3.1] - 2024-07-17
### Changed
- Some examples had the wrong resource name

## [0.3.0] - 2024-07-17

- Initial public release
