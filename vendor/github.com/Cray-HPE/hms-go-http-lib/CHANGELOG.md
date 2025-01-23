# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.4] - 2021-08-10

### Changed

- Added GitHub configuration files.

## [1.5.3] - 2021-07-22

### Changed

- Fixed forgotten stash reference.

## [1.5.2] - 2021-07-22

### Changed

- Replaced all references to stash.us.cray.com to github.

## [1.5.1] - 2021-07-20

### Changed

- Add support for building within the CSM Jenkins.

## [1.5.0] - 2021-06-28

### Security

- CASMHMS-4898 - Updated base container images for security updates.

## [1.4.1] - 2021-04-14

### Changed

- Updated Dockerfiles to pull base images from Artifactory instead of DTR.

## [1.4.0] - 2021-02-01

### Changed

- Updated license to MIT.
- Updated http.go for code changes in latest version of hms_certs

## [1.3.0] - 2021-01-14

### Changed

- Updated license file.

## [1.2.3] - 2020-10-29

### Security

- CASMHMS-4148 - Update go module vendor code for security fix.

## [1.2.2] - 2020-10-20

### Security

- CASMHMS-4105 - Updated hms-base and hms-securestorage vendor code to resolve libcrypto vulnerability.

## [1.2.1] - 2020-10-16

### Security

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.2.0] - 2020-08-31

### Changed

- Added the ability to use TLS cert-verified HTTP clients.

## [1.1.1] - 2020-04-30

### Changed

- CASMHMS-2974 - Updated hms-go-http-lib to use trusted baseOS.

## [1.1.0] - 2020-02-07

### Changed

- No longer require list of expected status codes.  Empty list means don't check status codes.

## [1.0.0] - 2019-08-22

### Added

- Initial implementation.

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security
