# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.9.1]

### Fixed

- Linter Errors

### Changed

- Support local builds
- Support goreleaser for builds/releases
- Move to github.com/OpenChami/hms-smd
- Container now based on wolfi
- Deprecated built-in kafaka listener for redfish events

## [2.9.0] - 2023-05-19

### Fixed

- Node architecture discovery

## [2.8.0] - 2023-05-03

### Added

- Added RTS management switch discovery

### Changed

- Reduced verbosity of V2 readiness/liveness requests in HSM logs.

## [2.7.0] - 2023-04-12

### Fixed

- Fixed HSM seeing empty memory slots as populated in EX hardware.

## [2.6.0] - 2023-02-27

### Added

- Added CT test coverage for most remaining APIs.
- Fixed several CT test bugs and made various improvements.

## [2.5.0] - 2023-02-03

### Changed

- Moved hardware-sensitive CT tests into separate stage.

## [2.4.0] - 2023-01-30

### Added

- CASMHMS-5863 - Added Reacquire() function to the reservations client library.

## [2.3.0] - 2023-01-26

### Fixed

- CASMHMS-5902: Linting of language in API spec (no content changes); corrected markdown formatting issue in changelog

## [2.2.0] - 2023-01-09

### Removed

- CASMHMS-5626 - Removed v1 API

## [1.62.0] - 2023-01-09

### Fixed

- HSN NIC numbering for devices running proliant iLO redfish.

## [1.61.0] - 2022-12-19

### Changed

- Update service values test to allow custom role and subrole configurations.

## [1.60.0] - 2022-12-01

### Fixed

- Restored previous name of GitHub Actions workflow file for test images.

## [1.59.0] - 2022-11-16

### Changed

CASMHMS-5747 - Refactored HSM CT tests for HMTH, including:
- Update HSM CT tests to use latest hms-test:4.0.0 image
- Break out HSM CT tests into non-disruptive, disruptive, and destructive test buckets
- Add many new API tests that execute in the runCT environment in the build pipeline
- Fixes to Swagger specification to reflect actual API behavior

## [1.58.1] - 2022-08-31

### Fixed

- CASMHMS-5726 - Fixed syntax error in migration down step 22

## [1.58.0] - 2022-08-11

### Fixed

- CASMHMS-5675 - HSM now discovers HSN NICs under '/redfish/v1/Chassis/<sysid>/Devices' for Proliant iLO redfish implementations.
- CASMHMS-5675 - HSN now ignores non-HSN NICs that show up for Proliant iLO redfish implementations.

## [1.57.0] - 2022-07-27

### Fixed

- CASMHMS-5355 - /locks/reservations/check now reports the reservations that were not found.
- CASMHMS-5387 - HSM now discovers the power URL for Intel nodes.

## [1.56.0] - 2022-07-19

### Fixed

- CASMHMS-5625 - Removed much of the redundant v1 API code in preparation for removal in CSM 1.4.
- CASMHMS-5610 - Fixed POST /Inventory/RedfishEndpoints returning 500 instead of 409 for conflicts.
- CASMHMS-5373 - Fixed locking bug preventing 'Flexible' requests from working.

## [1.55.0] - 2022-06-22

### Changed

- Updated CT tests to hms-test:3.1.0 image as part of Helm test coordination.

## [1.54.0] - 2022-06-21

### Added

- CASMHMS-5591 - Added indexes to the hwinv_hist table to speed up migration step 20.

## [1.53.0] - 2022-06-16

### Fixed

- CASMHMS-4972 - HSM now ignores A100 NV Switch and Baseboard when discovering GPUs on Proliant iLO devices.

## [1.52.0] - 2022-06-03

### Changed

- HSM now uses the 'Class' from SLS

## [1.51.0] - 2022-05-10

### Changed

- Converted image builds to be via github actions, updated the image links to be in artifactory.algol60.net
- Added a runCT.sh script that can run the tavern tests and smoke tests in a docker-compose environment.
- Refactored CT tests and their directory structure.
- Renamed dockerfiles.

## [1.50.0] - 2022-05-03

### Fixed

- CASMHMS-5511 - Fixed issue causing FRU data to get improperly populated for empty locations.

## [1.49.0] - 2022-03-04

### Added

- CASMHMS-4974 - Added HSM GET /v2/status/locks API.

## [1.48.0] - 2022-02-18

### Fixed

- CASMINST-4069 - Updated HSM components test for new expected BMC management roles.

## [1.47.0] - 2022-02-10

### Added

- Added flexible-model methods to the service-reservation package.

## [1.46.0] - 2022-02-07

### Added

- CASMHMS-5278 - Added HSM SCN subscription database unit tests.

## [1.45.0] - 2022-02-07

### Changed

- CASMHMS-5365 - Updated HSM CT tests for custom roles and subroles.

## [1.44.0] - 2022-02-02

### Added

- CASMHMS-5353 - Added CheckDeputyKeys() method to service_reservations package.

## [1.43.0] - 2022-01-31

### Added

- CASMHMS-5348 - Fixed NULL value issue with POST /Inventory/EthernetInterfaces

## [1.42.0] - 2022-01-24

### Added

- CASMHMS-633 - Added HSM hardware inventory unit tests.

## [1.41.0] - 2022-01-11

### Added

- CASMHMS-632 - Added HSM component endpoint unit tests.

## [1.40.0] - 2022-01-06

### Added

- CASMHMS-634 - Added HSM Redfish endpoint unit tests.

## [1.39.0] - 2021-12-21

### Changed

Replaced golang Sarama kafka interface with Confluent.

## [1.38.0] - 2021-12-07

### Added

- CASMTRIAGE-2801 - Added support for HPE PDUs to ComponentEndpoints CT test.

- Added vaild state transitions section to the HSM swagger doc.

### Changed

- Corrections to the HSM swagger doc including correcting typos, updating parameter descriptions to include valid values, updating parameter descriptions to include if they can be specified multiple times, and properly marking fields as required.

## [1.37.0] - 2021-11-05

### Added

- Support for HPE PDUs

## [1.36.0] - 2021-11-19

### Changed

- CASMHMS-5205 - Rename HSM CT smoke tests to swap execution order.

## [1.35.0] - 2021-11-18

### Changed

- CASMHMS-5198 - Updated image refs in the chart.

## [1.34.0] - 2021-11-16

### Changed

- CASMHMS-5272 - Added support for AuxiliaryController Redfish subtype to ComponentEndpoints CT test.

## [1.33.0] - 2021-11-12

### Changed

- CASMHMS-5239 - HSM now kicks off re-discovery for nodeBMCs when a power on redfish event is received for its slot.

### Fixed

- CASMHMS-5233 - HSM correctly ignores duplicate xnames given as arguments to `POST /Inventory/Discover`

## [1.32.0] - 2021-10-27

### Added

- CASMHMS-5055 - Added SMD CT test RPM.

## [1.31.0] - 2021-10-19

### Changed

- CASMHMS-5226 - Add priority value to postgres cluster resource

## [1.30.19] - 2021-10-06

### Changed

- CASMHMS-4951 - Changed HSM to use NAME and ProductPartNumber fields in place of empty Model and PartNumber fields for GPUs discovered on HPE hardware.

- CASMHMS-4954 - Changed HSM to use the NAME field in place of an empty Model field for Enclosures.

## [1.30.18] - 2021-09-28

### Changed

- Changed cray-service version to ~6.0.0

## [1.30.17] - 2021-09-10

### Changed

- Changed the docker image to run as the user nobody

## [1.30.16] - 2021-09-07

### Added

- CASMHMS-5039 - Added support for power capping for Bard Peak nodes.
- Workaround for discovery for Bard Peak to correctly discover node BMCs.

### Fixed

- Bulk postgres operations trying to operate on the same row multiple times.

### Changed

- CASMHMS-5041 - Set the 'Name' field in the power control struct for Apollo 6500.

## [1.30.15] - 2021-08-24

### Changed

- CASMHMS-5036 - Updated the discovery status CT smoke test with troubleshooting steps.

## [1.30.14] - 2021-08-19

### Changed

- CASMHMS-4835 - Changed HSM postgres operations to use bulk Inserts and Updates when working with multiple entries.

## [1.30.13] - 2021-08-10

### Changed

- Added GitHub configuration files.

## [1.30.12] - 2021-08-03

### Changed

- CASMTRIAGE-1808 - Updated the ComponentEndpoints CT test for multiple accelerator components.

## [1.30.11] - 2021-08-02

### Changed

- CASMHMS-4885 - Set pod priority for HSM.

## [1.30.10] - 2021-07-30

### Changed

- CASMHMS-4990 - Add "HPE" to the match list for Cray manufacturer.

## [1.30.9] - 2021-07-26

### Added

- github transition phase 3. Remove stash references.

## [1.30.8] - 2021-07-22

### Added

- Added Jenkins file and Makefile for migrating hms-smd to github.

## [1.30.7] - 2021-07-20

### Added

- CASMHMS-4927 - smd-init prunes previously bloated hwinv_hist database tables of redundant hardware history events.

### Changed

- CASMHMS-4927 - FRU history events are only generated if a change occurred.

### Fixed

- CASMHMS-4971 - Fixed HSM crashing when discovering Bard Peak nodes

## [1.30.6] - 2021-07-01

### Added

- CASMHMS-4930 - Enabled automatic postgres backups in the helm chart.

## [1.30.5] - 2021-07-13

### Changed

- CASMINST-2680 - Updated CT tests for when ncn-m001 is not part of the management cluster.

## [1.30.4] - 2021-06-30

### Security

- CASMHMS-4898 - Updated base container images for security updates.

## [1.30.3] - 2021-06-18

### Fixed

- CASMHMS-4884 - Fixed HSM crashing when manually adding power supplies via POST /Inventory/Hardware

## [1.30.2] - 2021-06-18

### Changed

- CASMINST-2511 - Update the ComponentEndpoints CT test to make InterfaceEnabled an optional EthernetNICInfo field and add it to RedfishSystemInfo.

## [1.30.1] - 2021-06-02

### Changed

- CASMHMS-4842 - HSM now joins a client group with its replicas to share at pool of redfish events from the kafka bus

## [1.29.1] - 2021-06-01

### Fixed

- CASMHMS-4865 - Fixed component filtering when locking components.

## [1.29.0] - 2021-05-28

### Added

- CASMHMS-4706 - Added support for power capping HPE Apollo 6500.

## [1.28.19] - 2021-05-28

### Changed

- CASMPET-4148 - Change smd-postgres pvc size to 100GB

## [1.28.18] - 2021-05-13

### Changed

- CASMHMS-4834 - Modifies Insert, Delete, and Update postgres operations on the v2 locking interface use bulk operations.

## [1.28.17] - 2021-05-14

### Added

- CASMHMS-4836 - Support for parsing redfish events from HPE iLo nodes

## [1.28.16] - 2021-05-10

### Changed

- Changed kubernetes values.yaml for podAntiAffinity from istio-ingressgateway

## [1.28.15] - 2021-05-05

### Changed

- Updated docker-compose files to pull images from Artifactory instead of DTR.

## [1.28.14] - 2021-05-04

### Changed

- CASMHMS-4796 - HSM no longer takes out row exclusive locks in postgres.
- CASMHMS-4796 - Reuses http transport whenever possible.
- CASMHMS-4796 - Pod resources are increased for both HSM and postgres.
- CASMHMS-4796 - Readiness probe timeout is increased.
- CASMHMS-4796 - Set GOMAXPROCS to tune HSM to the CPU resource limits.
- CASMHMS-4796 - Unset SetConnMaxLifetime() so postgres connections can be reused.
- CASMHMS-4796 - Set indexs on role/subrole rows in the components table

## [1.28.13] - 2021-05-04

### Changed

- CASMHMS-4810 - Allow valid nodeAccel type xnames for more than 8 GPUs

## [1.28.12] - 2021-05-03

### Changed

- CASMHMS-4811 - Added anti-affinity for HSM to avoid (if possible) scheduling on the same nodes as the Istio gateways.

## [1.28.11] - 2021-04-28

### Removed

- CASMHMS-4794 - Disabled CT test for the DiscoveryStatus API.

## [1.28.10] - 2021-04-20

### Fixed

- CASMHMS-4751 - Increased the wait-for-postgres resource limit

## [1.28.9] - 2021-04-16

### Fixed

- CASMHMS-4719 - Fix HSM postgres slowness during discovery floods on large (2000+ nodes) systems.

### Changed

- CASMHMS-4719 - Changed FRU tracking to be more simple and avoid long running sql queries.

## [1.28.8] - 2021-04-14

### Fixed

- CASMHMS-4700 - HSM now discovers GPUs in PCI slots on HPE hardware

## [1.28.7] - 2021-04-14

### Fixed

- CASMHMS-4713 - Fix HTTP response leaks

## [1.28.6] - 2021-04-12

### Changed

- CASMHMS-4693 - Update HSM Hardware Inventory CT test to allow empty drive bays.
- CASMHMS-4709 - Update HSM Hardware Inventory CT test to allow more ProcessorType data values.

## [1.28.5] - 2021-04-06

### Fixed

- CASMHMS-4593 - PATCH /v2/Inventory/EthernetInterfaces/<id> now allows ComponentID only patches

## [1.28.4] - 2021-04-06

### Changed

- CASMHMS-4579 - Update the cray-service chart to 2.4.5.

## [1.28.3] - 2021-03-31

### Changed

- CASMHMS-4605 - Update the loftsman/docker-kubectl image to use a production version.

## [1.28.2] - 2021-03-08

### Changed

- Added a note in HSM v1 and v2 Swagger about v1 deprecation.

## [1.28.1] - 2021-02-04

### Changed

- Added User-Agent header to outbound HTTP requests.

## [1.28.0] - 2021-02-01

### Changed

- Updated to MIT License
- Updated HMS libraries to latest
- Code changes to test.go code for updates to hms-cert

## [1.27.1] - 2021-01-20

### Changed

- CASMHMS-4334 Fixed issue with Processor discovery

## [1.27.0] - 2021-01-14

### Changed

- Updated license file.

## [1.26.8] - 2020-12-18

### Changed

- CASMHMS-4295 - Changed partitions API to restrict partition names to the p# or p#.# (hard.soft) naming convention for partitions so they will work correctly with other APIs.

## [1.26.7] - 2020-12-03

### Changed

- CASMHMS-4260 - Change NodeHsnNic hardware inventory data to show as NodeHsnNicFRUInfo instead of HSNNICFRUInfo.

## [1.26.6] - 2020-11-25

### Fixed

- CASMHMS-4246 - Fixed HSM using invalid MAC addresses to generate EthernetInterface entries.

## [1.26.5] - 2020-11-24

### Changed

- CASMHMS-4240 - Change NodeAccel hardware inventory data to show as NodeAccelFRUInfo instead of ProcessorFRUInfo.

## [1.26.4] - 2020-11-23

### Changed

- CASMHMS-4237 - Update NodeAccelRiserFRUInfoRF definitions: remove Manufacturer, add Producer and EngineeringChangeLevel

## [1.26.3] - 2020-11-23

### Added

- CASMHMS-4224 Added the discovery for NetworkAdapters (NodeHsnNic HMS types) to HSM

## [1.26.2] - 2020-11-20

### Added

- CASMHMS-4087 Added the NodeAccelRiser type to represent GPUSubsystem baseboards, ie Redstone

## [1.26.1] - 2020-11-18

### Changed

- CASMHMS-4211 - Added final CA bundle configmap handling to Helm chart.

## [1.26.0] - 2020-11-17

### Changed

- CASMHMS-4158 - The V2 API for Component Ethernet Interfaces now supports associating multiple IP addresses to a single MAC Address. The new IP Address structure has a optional Network field to note which network an IP Address is apart of. Added new APIS to manipulate the IPAddresses
- The V1 Component Ethernet Interfaces API remains unchanged, except for new behavior when performing a PATCH on a component ethernet interface with a new IPAddress that has multiple IP addresses it will return a Bad Request status code as this is a ambiguous situation.

## [1.25.6] - 2020-11-13

### Fixed

- CASMHMS-4077 - HSM now periodically updates the timestamp of currently running discovery jobs.

### Changed

- CASMHMS-4077 - Much of the HSM manual rediscovery path has been parallelized

## [1.25.5] - 2020-11-10

### Changed

- CASMHMS-3848 - HSM now queries HBTD for heartbeat status of nodes it discovers in the 'On' state to see if they should be 'Ready'.

## [1.25.4] - 2020-11-05

### Changed

- CASMHMS-3232 - HSM now retries sending failed SCNs.

## [1.25.3] - 2020-10-29

### Security

- CASMHMS-4148 - Update HMS vendor code for security fix.
- Set grpc go module to v1.29.1 to resolve smd-related grpc/etcd incompatibility issue.

## [1.25.2] - 2020-10-27

### Changed

- CASMHMS-4144 - Update to latest cray-service base chart v2.2.0 to pick up postgres labels.

## [1.25.1] - 2020-10-21

### Security

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.25.0] - 2020-10-19

### Added

- Added a V2 of SMD; V1 is now on the deprecation path.  We have added a new locking and reservations API

## [1.24.1] - 2020-10-16

### Added

- CASMHMS-4111 - Added a POST to the /Inventory/Hardware REST endpoint to generically add hw inventory entries from external sources.

### Removed

- CASMHMS-4111 - Removed HSNInterfaces APIs and functionality

## [1.24.0] - 2020-10-13

### Added

- Added support for TLS certs for Redfish endpoint communcations.

## [1.23.1] - 2020-09-16

### Fixed

- CASMHMS-4026 - HSM now correctly resyncs its ComponentEndpoint cache when a redfish event comes from a PDU controller.

## [1.23.0] - 2020-09-16

### Summary and Scope

These are changes to charts in support of:

- moving to Helm v1/Loftsman v1
- the newest 2.x cray-service base chart
  - upgraded to support Helm v3
  - modified containers/init containers, volume, and persistent volume claim value definitions to be objects instead of arrays
- the newest 0.2.x cray-jobs base chart
  - upgraded to support Helm v3

## [1.22.10] - 2020-09-10

### Security

- CASMHMS-3997 - Updated hms-smd to use latest trusted baseOS images.

## [1.22.9] - 2020-09-10

### Added

- CASMHMS-4018 - Added code to process GPU info from redfish correctly

## [1.22.8] - 2020-09-02

### Added

- CASMHMS-3975 - Added a mechanism for restarting orphaned discovery jobs

## [1.22.7] - 2020-08-18

### Added

- CASMHMS-3509 - Added the hms-base config file into the HSM chart

## [1.22.6] - 2020-08-14

### Changed

- CASMHMS-3807 - Changed PDU discovery behavior to discover outlets as CabinetPDUPowerConnector HMS type.

## [1.22.5] - 2020-08-14

### Changed

- CASMHMS-3914 - Changed HSM to skip node discovery for CMCs with special NodeBMC xname xXcCsSb999

## [1.22.4] - 2020-08-07

### Changed
- CASMHMS-3888 - Changed PDU discovery behavior to allow Cabinet PDU controllers to have more than 1 Cabinet PDU.

## [1.22.3] - 2020-08-05

### Added

- CASMHMS-3871 - Added PowerStatusChange to the list of valid redfish event types for HSM to process.

## [1.22.2] - 2020-07-24

### Changed

- CASMHMS-3818 - CT functional test updates for /State/Components SubRoles and /SCN States.

## [1.22.1] - 2020-07-24

### Changed

- CASMHMS-3815 - Bumped the resource limits and made the compose file work.

## [1.22.0] - 2020-07-21

### Added

- CASMHMS-1466 - Added partition query parameters to /Inventory/Hardware
- CASMHMS-1466 - Added the 'parent_node' column to the hwinv_loc table to be able to associate lower components with partitions of their parents
- CASMHMS-1466 - Added a schema view that includes partition information with hwinv data.
- CASMHMS-1466 - Added the 'laststatus' query parameter to /Inventory/RedfishEndpoints to allow queries to be filtered based on discovery status.

## [1.21.3] - 2020-07-14

### Added

- CASMHMS-2921 - Fru Tracking of sC

## [1.21.2] - 2020-07-06

### Added

- CASMHMS-2919 - Fru Tracking of nC

## [1.21.1] - 2020-07-01

### Changed

- CASMHMS-3617 - Changed 'PATCH /Inventory/EthernetInterfaces' to include 'CompID' as a patchable value.

## [1.21.0] - 2020-06-26

### Added

- CASMHMS-3462 - HSNInterfaces REST API which includes GET/POST/DELETE /Inventory/HSNInterfaces and GET/PATCH/DELETE /Inventory/HSNInterfaces/{xname}

## [1.20.10] - 2020-06-15

### Removed

- CASMHMS-3575 - Disabled CT test for /Defaults/NodeMaps since it is deprecated in favor of SLS.

## [1.20.9] - 2020-06-10

### Added

- CASMHMS-3553 - Updated HSM /State/Components CT test cases for optional 'SubRole' and 'Subtype' fields.

## [1.20.8] - 2020-06-10

### Changed

- CASMHMS-3506 - HSM now treats Ready/Warning StateData patches as only affecting components in the Ready state.

## [1.20.7] - 2020-06-08

### Fixed

- CASMHMS-3526 - fixed job cleanup.

## [1.20.6] - 2020-06-05

### Changed

- CASMHMS-3531 - Updated HSM /State/Components CT test case for optional 'SoftwareStatus' field.
- CASMHMS-3532 - Updated HSM /Subscriptions/SCN CT test case for new subscription keys.

## [1.20.5] - 2020-06-05

### Changed

- Re-inventory triggered by redfish events now only generate "Scanned" hardware history events.

## [1.20.4] - 2020-06-03

### Changed

- removed cray-smd-loader job per CASMHMS-3392

## [1.20.3] - 2020-05-26

### Changed

- Added a locking mechanism for the HSM jobList to prevent crashes.

## [1.20.2] - 2020-05-26

### Changed

- Updated the cray-service chart version.
- Changed smd-init to downgrade as well as upgrade schemas
- smd-init is now built in the same container image as HSM

### Added

- Added a job to delete the previously run smd-init and smd-loader jobs for upgrade/downgrade
- Added a persistent storage volume for storing all previously applied schema migration steps

## [1.20.1] - 2020-05-20

### Changed

- replicaCount now set to 3 in helm chart for resiliency

## [1.20.0] - 2020-05-13

### Added

- Added a REST API for storing and querying for component ethernet interfaces

## [1.19.10] - 2020-05-06

### Changed

- CASMHMS-2966 - Update hms-smd build to use trusted baseOS.

## [1.19.9] - 2020-05-01

### Changed

- Update version of hms-base to 1.7.3, which includes changes for CASMHMS-3403: modifications to xname validation for CMMRectifiers

## [1.19.8] - 2020-04-24

### Changed

- Increased the size of the fru_id column from varchar(63) to varchar(255) in the hwinv_by_loc, hwinv_by_fru, and hwinv_hist HSM database tables.
- Added more robust fruid validation to the fruid generation function.

## [1.19.7] - 2020-04-17

### Changed

- CASMHMS-3241 - Update Redfish endpoint CT test for optional IPAddress field.

## [1.19.6] - 2020-04-16

### Changed

- HSM now sets detached FRUs associated with a disabled RedfishEndpoint from their loc.
- HSM generates "removed" events in hardware history when RedfishEndpoints are disabled

### Fixed

- Fixed a bug in the hardware history cleanup logic causing all history to get deleted each day.
- Fixed a bug in node standby polling jobs causing them to match powerstate stings case-sensitively.

## [1.19.5] - 2020-04-15

### Added

- CASMHMS-3096 - added FRU tracking support for power supplies, specifically CMMRectifiers and NodeEnclosurePowerSupplies

## [1.19.4] - 2020-04-06

### Changed

- HSM now sets components associated with a disabled RedfishEndpoint to 'Empty'

## [1.19.3] - 2020-04-02

### Fixed

- HSM now correctly processes the NULL partition parameter correctly for GET /groups/<group>/members

## [1.19.2] - 2020-03-31

### Added

- Added the IPAddress field to the RedfishEndpoints API as a patchable and a queryable field.

## [1.19.1] - 2020-03-30

### Changed

- CASMHMS-3211 - Update Redfish endpoint CT test for chassis and router BMCs.

## [1.19.0] - 2020-03-26

### Added

- Added a configmap volume mount to the cray-smd deployment to mount as an updatable configfile.
- Added a config file watcher to pick up any new roles/subroles defined in the config file.
- Added /service/values/\* REST APIs to list valid values for hms-base enums.

### Changed

- Changed the valid component role and subrole values to be extendable via configfile.

## [1.18.3] - 2020-03-26

### Changed

- CASMHMS-3163 - Add additional cleanup actions for test interrupts to HSM group and partition CT tests.

## [1.18.2] - 2020-03-25

### Fixed

- CASMHMS-3097 - Update Redfish Pkg by standardizing FRUID initialization.

## [1.18.1] - 2020-03-23

### Fixed

- CASMHMS-2929 - Update Redfish Pkg by adding SerialNumber to Processor data.

## [1.18.0] - 2020-03-16

### Fixed

- CASMHMS-3137 - Update HSM CT test for /State/Components to include new 'Class' field.

## [1.17.2] - 2020-03-13

### Changed

- Transitioning a component from Ready to On is no longer a valid state transition
- Redfish events are now processed concurrently

## [1.17.1] - 2020-03-10

### Fixed

- 405 responses to include Allow header with list of allowed HTTP methods

## [1.17.0] - 2020-03-09

### Changed

- Information under the /State/Components REST API now includes the component Class (River/Mountain).

## [1.16.9] - 2020-03-06

### Fixed

- Fixed SLS URL.
- Made Docker compose work. Running `docker-compose up -d` in the root directory now gives you a working HSM with Vault.

## [1.16.8] - 2020-03-03

### Changed

- HSM now delays discovery when processor info is not populated when discovering nodes.

## [1.16.7] - 2020-02-28

### Changed

- Update discovery functions in pkg/redfish to use a default flag of "OK"

## [1.16.6] - 2020-02-26

### Changed

- Create standard FRUID initialization/validation function, apply to Memory and Chassis

## [1.16.5] - 2020-02-26

### Fixed

- HSM segfault when generating hardware history entries.

## [1.16.4] - 2020-02-24

### Changed

- Updated FRUID initialization code for MemoryMods to use unique identifier

## [1.16.3] - 2020-02-21

### Added

- Added SMD_HWINVHIST_AGE_MAX_DAYS environment variable to control when FRU history entries should be cleaned up. This defaults to 365.

### Changed

- HSM generates FRU historical data.

## [1.16.1] - 2020-02-20

### Fixed

- CASMHMS-3007 - redact passwords from redfish struct output.

## [1.16.0] - 2020-02-13

### Added

- Added PATCH /hsm/v1/Inventory/RedfishEndpoints/{xname}

### Changed

- Database version checking now looks for installed schema versions greater than or equal to the expected schema.

## [1.15.0] - 2020-02-11

### Added

- Added functionality to hmsds to store hardware inventory historical data.
- Added /hsm/v1/Inventory/Hardware/History REST endpoint (GET/DELETE)
- Added /hsm/v1/Inventory/Hardware/History/{xname} REST endpoint (GET/DELETE)
- Added /hsm/v1/Inventory/HardwareByFRU/History REST endpoint (GET)
- Added /hsm/v1/Inventory/HardwareByFRU/History/{fruid} REST endpoint (GET/DELETE)

## [1.14.0] - 2020-02-07

### Changed

- CASMHMS-1009 - added support for disks

## [1.13.2] - 2020-02-05

### Changed

- CASMHMS-2908 - RedfishEndpoints API test workaround for Intel firmware v1.93 UANs failing discovery CASMHMS-2767.

## [1.13.1] - 2020-01-31

### Changed

- CASMHMS-2860 - Updated CT test for Hardware FRU tracking API additions.

## [1.13.0] - 2020-01-30

### Changed

- Updated imports to use new hms-base, hms-compcredentials, hms-securestorage, and hms-msgbus repos in place of deprecated hms-common versions.

## [1.12.0] - 2020-01-30

### Added

- Liveness probe & settings

### Changed

- Only log probes when DEBUG or higher
- Increased k8s initialDelaySeconds and periodSeconds

## [1.11.0] - 2020-01-23

### Added

- Added query parameters to /hsm/v1/Inventory/Hardware REST endpoint
- Added query parameters to /hsm/v1/Inventory/HardwareByFRU REST endpoint
- Added query parameters to /hsm/v1/Inventory/Hardware/Query/{xname} REST endpoint
- Implemented /hsm/v1/Inventory/Hardware/Query/{xname} to accept more xnames than just "s0"

## [1.10.6] - 2020-01-22

### Fixed

- Increased size of postgresql volume to 30Gi.

## [1.10.5] - 2020-01-17

### Added

- Additional functional Tavern API tests for CT framework.

## [1.10.4] - 2019-12-20

### Added

- Functional Tavern API tests for CT framework.

## [1.10.3] - 2019-12-12

### Changed

- Updated version of hms-common.

## [1.10.2] - 2019-12-3

### Changed

- Redfish node discovery now waits for all info to be loaded from BIOS

## [1.10.1] - 2019-11-26

### Changed

- Improved retry logic in loader to essentially retry forever.

## [1.10.0] - 2019-11-22

### Added

- Subroles to HSM

## [1.9.5] - 2019-11-20

### Changed

- HSM now reloads node hwinv when nodes power on.

## [1.9.4] - 2019-11-19

### Added

- Added an Enabled field to ComponentEndpoints as a reference to the same field in the parent RedfishEndpoint.

## [1.9.3] - 2019-11-15

### Fixed

- Workaround added for gigabyte nodes with missing Ethernet Interfaces

## [1.9.2] - 2019-11-14

### Fixed

- Istio preventing HSM from receiving redfish events

## [1.9.1] - 2019-11-12

### Changed

- Reduced HSM's default log verbosity

## [1.9.0] - 2019-11-08

### Fixed

- Nodes staying in the Standby state when they don't send redfish events.

## [1.8.8] - 2019-10-28

### Added

- Support for using SLS to get NID and Role assignments for nodes

## [1.8.7] - 2019-10-25

### Added

- The CrayAlerts registry to the list of valid registries for ResourcePowerStateChanged redfish events

## [1.8.6] - 2019-10-23

### Added

- GET /hsm/v1/service/ready REST API for HSM health checks

### Changed

- Liveliness and readiness probes for the HSM deployment now point to GET /hsm/v1/service/ready
- The hmsds log level now gets set to match HSM's log level.

### Fixed

- Missing query parameters, enabled and softwarestatus, in the swagger doc.

## [1.8.5] - 2019-10-17

### Added

- Added Oids to the PowerControl struct

## [1.8.4] - 2019-10-16

### Added

- Discovery of EPO redfish endpoints for chassis.

## [1.8.3] - 2019-10-16

### Added

- POST to /hsm/v1/State/Components
- PUT to /hsm/v1/State/Components/<xname>

## [1.8.2] - 2019-10-11

### Fixed

- PowerControl data discovery for non-mountain components

## [1.8.1] - 2019-10-10

### Fixed

- GET /locks returns all locks instead of get the first.

## [1.8.0] - 2019-10-09

### Added

- REST API for PowerMaps.

## [1.7.2] - 2019-10-03

### Removed

- Redfish credentials from REST API output.

## [1.7.1] - 2019-10-03

### Added

- Power Control Info discovery for mountain nodes

## [1.7.0] - 2019-09-25

### Added

- REST API for component locking.

### Fixed

- Gigabyte node enclosure discovery.

## [1.6.2] - 2019-08-22

### Added

- Support for parsing redfish events from updated Gigabyte nodes

## [1.6.1] - 2019-08-12

### Added

- Added new loader utility which is used to load HSM's Node NID map.

## [1.6.0] - 2019-08-08

### Added

- Changes from hms-common where picked up to include that addition of the 'Management' role.
- The 'Management' role to the HSM swagger document.
- Vault operations were added to smd. Configurable via the 'SMD_RVAULT' and 'SMD_WVAULT' environment variables.
- Vault environment variables, 'VAULT_ADDR' and 'VAULT_SKIP_VERIFY', to values.yaml to point HSM to a Vault instance.
- Product specification to the jenkins file

### Removed

- Unused Mariadb code

### Fixed

- yamllint errors and warnings
- Segfault when if database transactions can't be started
- Temp file creation in testing OS independant.
- AllowableValues for outlet power control
- Schema change for redfish resetTypes

## [1.5.1] - 2019-07-16

### Fixed

- Fixed bug in chart with incorrect `imagesHost` setting.

## [1.5.0] - 2019-07-12

### Changed

- Postgres is now the default and only supported backing store
- cray-smd now uses helm and the Postgres operator.
- cray-smd-init has been re-written to install/upgrade the schema for postgres
  using golang-migrate.

## [1.4.0] - 2019-07-08

### Changed

- Add rediscovery for RedfishEndpoints on PUT updates, with related bug fix.
- Fix xname normalization issues, group/partition normalization issues

## [1.3.1] - 2019-05-29

### Changed

- Fix bad 500 status responses that don't pass through an HMS error and return 400 like they should for a bad request.  These aren't internal DB errors and we don't want to report them that way.

## [1.3.0] - 2019-05-24

### Changed

- Added support for PDU discovery.

## [1.2.0] - 2019-05-16

### Added

- Added /ServiceEndpoints/* REST endpoints to HSM for querying for information on discovered redfish services.
- Added discovery logic to HSM to discover redfish services.
- Added storage logic to hmsds to store discovered redfish service information.
- Added logic to HSM to check for correct schema version.

### Changed

- Changed the table view for service_endpoint_info to correct extracting FQDN info for the redfish endpoint.

### Removed

## [1.1.0] - 2019-05-13

### Added

- Brought in `redfish`, `sharedtest`, and `sm` packages to this repo as they're really specific to HSM.
- Broung in `hmsds` package to the internal part of this repo as it shouldn't even be used by any other services.
- Checked in vendor code for 3rd party dependencies.

### Changed

- Updated Dockerfile to now copy over new `pkg` and `internal` folders when building.

### Removed

- Old version (v1.0.0) of hms-common code.

## [1.0.0] - 2019-05-13

### Added

- This is the initial release. It contains everything that was in `hms-services` at the time with the major exception of being `go mod` based now.

### Changed

### Deprecated

### Removed

### Fixed

### Security
