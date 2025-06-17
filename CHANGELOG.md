# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

> [!NOTE]
> While we strive to maintain backwards compatibility as much as possible, we can't guarantee semantic versioning will be strictly followed, as this provider depends on the underlying [bunny.net API](https://docs.bunny.net/reference/bunnynet-api-overview).

## [Unreleased]
### Fixed
- resource pullzone_shield: missing realtime_threat_intelligence causes a panic

## [0.7.3] - 2025-06-17
### Added
- resource pullzone_edgerule: added `priority` to control the execution order of the edge rules ([#37](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/37));

## [0.7.2] - 2025-06-04
### Added
- resource pullzone_edgerule: add missing actions and trigger types ([#38](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/38));

## [0.7.1] - 2025-06-03
### Fixed
- resource pullzone_hostname: create: clean up after a loadFreeCertificate error ([#36](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/36));

## [0.7.0] - 2025-04-15
### Added
- Support for Bunny Shield
- Build releases for openbsd
### Fixed
- resource storage_zone: handle empty set for `replication_regions`
- resource storage_zone: validate `name`

## [0.6.2] - 2025-03-18
### Changed
- updated dependencies

## [0.6.1] - 2025-02-26
### Fixed
- resource storage_zone: support variables during validation ([#31](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/31))

## [0.6.0] - 2025-02-25
### Added
- Support for Magic Containers
- resource storage_zone: validate regions
### Changed
- Bumped minimum Go version to 1.23
- Bumped minimum Terraform version to 1.4

## [0.5.6] - 2025-01-23
### Added
- resource stream_library: support for HEVC and AV1 on output_codecs
### Fixed
- resource dns_record: create CNAME with a defined weight ([#29](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/29))

## [0.5.5] - 2025-01-21
### Added
- resource stream_library: Premium Encoding support

## [0.5.4] - 2025-01-20
### Added
- resource compute_script_secret ([#27](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/27))

## [0.5.3] - 2025-01-17
### Added
- resource compute_script: expose deployment_key and release attributes ([#26](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/26))

## [0.5.2] - 2025-01-14
### Fixed
- resource compute_script: workaround for updating code in published scripts ([#25](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/25))

## [0.5.1] - 2025-01-07
### Changed
- updated dependencies
### Fixed
- resource pullzone: fix panic when origin or routing blocks are missing ([#20](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/20))
- resource dns_record: support PZ type ([#22](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/23))
- resource pullzone_edgerule: fix action params validation ([#23](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/23))

## [0.5.0] - 2024-12-12
### Changed
- Bumped minimum Go version to 1.22
- update dependencies
### Added
- resource pullzone_edgerule: support ActionParameter3
- resource pullzone_edgerule: validate Redirect action parameters

## [0.4.2] - 2024-12-09
### Fixed
- resource pullzone: create multiple at once throws 502 Bad Gateway

## [0.4.1] - 2024-11-22
### Fixed
- ValidateResource has unknown values on first pass if they use variables ([#18](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/18))

## [0.4.0] - 2024-11-07
### Added
- Support for Compute Script

## [0.3.16] - 2024-10-31
### Added
- negative number validation for all integer attributes
### Fixed
- resource pullzone_edgerule: validate trigger fields ([#17](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/17))
- concurrent resource creation with set attributes causes intermixed values

## [0.3.15] - 2024-09-12
### Added
- resource pullzone_edgerule: multi-actions support

## [0.3.14] - 2024-09-10
### Changed
- resource pullzone_hostname: deleting an internal hostname (`*.b-cdn.net`) will remove the resource from state without deleting it, as internal hostnames cannot be deleted;
- resource pullzone_hostname: creating an internal hostname (`*.b-cdn.net`) will adopt the pre-existing default hostname instead of creating a new one, as the default hostname is automatically created with the `pullzone` resource;

## [0.3.13] - 2024-09-09
### Fixed
- resource pullzone_optimizer_class: some fields were wrongly escaped ([#11](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/11))

## [0.3.12] - 2024-09-06
### Added
- data source bunnynet_dns_zone
- data source bunnynet_dns_record
### Changed
- resource storage_zone: custom_404_file_path can be unset

## [0.3.11] - 2024-09-02
### Added
- resource pullzone_hostname: custom certificates support
### Changed
- Expose API error messages for all resources

## [0.3.10] - 2024-08-28
### Changed
- resource pullzone: autoupdate use_background_update when changing cache_stale ([#4](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/4))
- resource dns_record: accelerated_pullzone cannot be set ([#5](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/5))
- resource pullzone: avoid deadlocks when changing multiple resources linked to the same pullzone ([#6](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/6))

## [0.3.9] - 2024-08-20
### Changed
- resource pullzone: autoupdate cache_expiration_time when changing permacache_storagezone

## [0.3.8] - 2024-08-19
### Added
- resource pullzone: originshield_zone attribute
- resource dns_record: document how to set an apex domain record

## [0.3.7] - 2024-08-15
### Fixed
- resource pullzone_edgerule: creating multiple edgerules for the same pullzone causes inconsistencies in tfstate
- resource storage_zone: zone_tier cannot be changed
### Changed
- resource dns_zone: run Create() in two passes
- resource dns_zone: refactored custom nameserver validation
- resource storage_zone: run Create() in two passes

## [0.3.6] - 2024-08-14
### Fixed
- resource pullzone_hostname: cannot create resource with tls_enabled=true and force_ssl=true ([#2](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/2))

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
