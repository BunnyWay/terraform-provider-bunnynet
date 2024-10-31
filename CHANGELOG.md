# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

> [!NOTE]
> While we strive to maintain backwards compatibility as much as possible, we can't guarantee semantic versioning will be strictly followed, as this provider depends on the underlying [bunny.net API](https://docs.bunny.net/reference/bunnynet-api-overview).

## [Unreleased]
## Added
- negative number validation for all integer attributes
## Fixed
- resource pullzone_edgerule: validate trigger fields ([#17](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/17))
- concurrent resource creation with set attributes causes intermixed values

## [0.3.15] - 2024-09-12
## Added
- resource pullzone_edgerule: multi-actions support

## [0.3.14] - 2024-09-10
## Changed
- resource pullzone_hostname: deleting an internal hostname (`*.b-cdn.net`) will remove the resource from state without deleting it, as internal hostnames cannot be deleted;
- resource pullzone_hostname: creating an internal hostname (`*.b-cdn.net`) will adopt the pre-existing default hostname instead of creating a new one, as the default hostname is automatically created with the `pullzone` resource;

## [0.3.13] - 2024-09-09
### Fixed
- resource pullzone_optimizer_class: some fields were wrongly escaped ([#11](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/11))

## [0.3.12] - 2024-09-06
## Added
- data source bunnynet_dns_zone
- data source bunnynet_dns_record
### Changed
- resource storage_zone: custom_404_file_path can be unset

## [0.3.11] - 2024-09-02
## Added
- resource pullzone_hostname: custom certificates support
### Changed
- Expose API error messages for all resources

## [0.3.10] - 2024-08-28
## Changed
- resource pullzone: autoupdate use_background_update when changing cache_stale ([#4](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/4))
- resource dns_record: accelerated_pullzone cannot be set ([#5](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/5))
- resource pullzone: avoid deadlocks when changing multiple resources linked to the same pullzone ([#6](https://github.com/BunnyWay/terraform-provider-bunnynet/issues/6))

## [0.3.9] - 2024-08-20
## Changed
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
