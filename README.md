# State Management Database (SMD)

## Table of Contents
1. [About](#about)
2. [Overview](#overview)
3. [Build & Install](#build--install)
4. [Testing](#testing)
5. [Running](#running)
6. [More Reading](#more-reading)

## About

The State Management Database (SMD) is a robust service designed for monitoring, tracking, and managing hardware components in high-performance computing (HPC) environments. It performs dynamic inventory discovery, interrogates hardware controllers, and maintains real-time state and lifecycle data. SMD captures essential component details such as hardware status, logical roles, architecture, and resource capabilities, making this information accessible via REST queries and event-driven notifications. Additionally, it facilitates component grouping, partitioning, power operations, firmware management, and boot-time configuration. By maintaining a comprehensive hardware inventory and tracking system changes, SMD ensures efficient resource management, operational continuity, and streamlined troubleshooting across diverse HPC infrastructures.

- SMD provides inventory management services for HPC systems based on BMC discovery and enumeration.

## Overview
SMD is responsible for:
- Discovering hardware inventory.
- Tracking hardware state, logical roles, and enabled/disabled status.
- Creating and managing component groups and partitions.
- Storing and retrieving Redfish endpoint and component data.
- Monitoring hardware state transitions via Redfish events.


## Build & Install

### Using GoReleaser
SMD employs [GoReleaser](https://goreleaser.com/) for automated releases and build metadata tracking. 

To build locally:
#### Set Environment Variables
```bash
export GIT_STATE=$(if git diff-index --quiet HEAD --; then echo 'clean'; else echo 'dirty'; fi)
export BUILD_HOST=$(hostname)
export GO_VERSION=$(go version | awk '{print $3}')
export BUILD_USER=$(whoami)
```

#### Install GoReleaser
Follow [GoReleaser’s installation guide](https://goreleaser.com/install/).

#### Build Locally
```bash
goreleaser release --snapshot --clean
```
Built binaries will be located in the `dist/` directory.

## Testing
SMD includes continuous test (CT) verification for live HPC systems. The test image is invoked via `helm test`.
In addition to the service itself, this repository builds and publishes cray-smd-test images containing tests that verify HSM on live Shasta systems. The tests are invoked via helm test as part of the Continuous Test (CT) framework during CSM installs and upgrades. The version of the cray-smd-test image (vX.Y.Z) should match the version of the cray-smd image being tested, both of which are specified in the helm chart for the service.


## Running

### Runtime Options
Environment variables can be set for runtime configurations:
```bash
RF_MSG_HOST   # Kafka host:port:topic
SMD_PROXY     # socks5 proxy for Redfish endpoint interrogation
SMD_DBTYPE    # Database type (default: postgres)
SMD_DBNAME    # Database name (default: hmsds)
SMD_DBUSER    # Database user (default: hmsdsuser)
SMD_DBHOST    # Database hostname (e.g., cray-smd-postgres in Kubernetes)
SMD_DBPORT    # Database port (default: 5432)
SMD_DBPASS    # Database password
SMD_DBOPTS    # Additional DB parameters
LOGLEVEL      # Logging level (0-4)
```

### Running Outside Kubernetes
To run SMD locally with a PostgreSQL database:

1. Start a local PostgreSQL container:
   ```bash
   sudo docker run --rm --name cray-smd-postgres -e POSTGRES_PASSWORD=hmsdsuser \
   -e POSTGRES_USER=hmsdsuser -e POSTGRES_DB=hmsds -d -p 5432:5432 postgres:10.8
   ```
2. Initialize the database schema:
   ```bash
   sudo docker run --name smd-init --link cray-smd-postgres:cray-smd-postgres \
   -e SMD_DBHOST=cray-smd-postgres -e SMD_DBOPTS="sslmode=disable" -e SMD_DBPASS=hmsdsuser \
   -d dtr.dev.cray.com:443/cray/cray-smd-init:latest
   ```
3. Start the SMD service:
   ```bash
   sudo docker run --name smd --net host -p 27779:27779 -e SMD_DBHOST=127.0.0.1 \
   -e SMD_DBPASS=hmsdsuser -e SMD_DBOPTS="sslmode=disable" \
   -e SMD_PROXY="socks5://127.0.0.1:9999" -d dtr.dev.cray.com:443/cray/cray-smd:latest
   ```
4. Verify the service is running:
   ```bash
   curl -k https://localhost:27779/smd/hsm/v2/groups
   ```
---
#### Using Proxy Mode to Get Access to Non-Local BMCs

Find the machine you wish to discover and ssh to it with dynamic port
forwarding enabled on the local port you gave for SMD_PROXY:

```bash
ssh -D 9999 root@example-sms.us.cray.com
```

Leave this window open until you are finished with the discovery.

Double check /etc/hosts for the BMC IP addresses that are assigned to the nodes
you wish to discover, in case they are non-standard ones

#### Discovering Nodes Once HSM is Properly Running

If the proxy has been set up (or you are running locally on an SMS), then
you can then create endpoints for every BMC you wish to discover using their
native BMC IP addresses.

_NOTE: If you need particular NIDs and Roles, you will need to set up xname_
_entries in /hsm/v2/Defaults/NodeMaps BEFORE discovery OR patch the NID and/or_
_Role fields after discovery:_

See: https://connect.us.cray.com/confluence/display/HSOS/HSM+Documentation+for+SPS2+-+Setting+Default+NIDs

***Example creation and discovery of preview system computes***

These are the usual computes found on a standard preview system, but you can easily adapt this example for whatever is in /etc/hosts.  Just make sure you use the BMC xname and a raw IP (if using a socks5 proxy):

```text
curl -k -d '{"ID": "x0c0s28b0", "RediscoverOnUpdate":true, "Hostname":"10.4.0.5", "User": "root","Password": "somePassword"}' -H "Content-Type: application/json" -X POST https://localhost:27779/hsm/v2/Inventory/RedfishEndpoints
curl -k -d '{"ID": "x0c0s26b0", "RediscoverOnUpdate":true, "Hostname":"10.4.0.6", "User": "root","Password": "somePassword"}' -H "Content-Type: application/json" -X POST https://localhost:27779/hsm/v2/Inventory/RedfishEndpoints
curl -k -d '{"ID": "x0c0s24b0", "RediscoverOnUpdate":true, "Hostname":"10.4.0.7", "User": "root","Password": "somePassword"}' -H "Content-Type: application/json" -X POST https://localhost:27779/hsm/v2/Inventory/RedfishEndpoints
curl -k -d '{"ID": "x0c0s21b0", "RediscoverOnUpdate":true, "Hostname":"10.4.0.8", "User": "root","Password": "somePassword"}' -H "Content-Type: application/json" -X POST https://localhost:27779/hsm/v2/Inventory/RedfishEndpoints
```
> [!NOTE]
> the above path is assuming you are running docker in a bare container (see above).  Otherwise use 'https://<standard-api-gateway-host>/apis/smd/hsm/v2/... instead of 'https://localhost:27779/hsm/...'

> [!NOTE]
> Also note that inventory discovery is a read-only operation and should not do anything to the endpoints besides walk them via GETs.   The "RediscoverOnUpdate":true field is important because it will automatically kick off inventory discovery.


### API Overview
Please refer to [Architecture and Design Details](SMD-DESIGN.md) for API overview 

### Additional API Documentation

The complete HSM (smd) API documentation is included in the Cray API docs.
This is the nightly-generated version.  Content is generated in an automated
fashion from the current swagger.yaml file.

http://web.us.cray.com/~ekoen/cray-portal/public

Latest detailed API usage examples:

https://github.com/OpenCHAMI/smd/blob/master/docs/examples.adoc  (current)

Latest swagger.yaml (if you would prefer to use the OpenAPI viewer of your choice):

https://github.com/OpenCHAMI/smd/blob/master/api/swagger_v2.yaml (current)

## More Reading
- [Architecture and Design Details](SMD-DESIGN.md)
- [API Definitions](api/swagger_v2.yaml)
- [Full API Documentation](https://github.com/OpenCHAMI/smd/blob/master/api/swagger_v2.yaml)
- [HPE’s Original SMD Documentation](https://github.com/Cray-HPE/hms-smd)

For advanced configuration, troubleshooting, and database management, refer to additional documentation in the `docs/` directory.

