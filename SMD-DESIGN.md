# State Management Database(smd)

This document outlines the Architecture and design details of SMD 


## Table of Contents
1. [SMD Capabilities](#smd-capabilities)
2. [SMD Overview](#smd-overview)
3. [SMD API](#smd-api)
   - [API Overview](#api-overview)
   - [Redfish Endpoints](#redfish-endpoints)
   - [Component Redfish Endpoint Information](#component-redfish-endpoint-information)
   - [Component State](#component-state)
   - [Additional Component State Queries](#additional-component-state-queries)
   - [Hardware Inventory and FRUs](#hardware-inventory-and-frus)
   - [NodeMaps](#nodemaps)
   - [Component Groups](#component-groups)
   - [Component Partitions](#component-partitions)
   - [Component Group and Partition Memberships](#component-group-and-partition-memberships)
4. [SMD Features](#smd-features)
   - [Feature Map](#feature-map)
   - [Current Features](#current-features)
   - [Future Features And Updates](#future-features-and-updates)
5. [Design Details](#design-details)
   - [States](#states)
   - [State Transition Rules](#state-transition-rules)
   - [Setting Default Node NIDs and Roles](#setting-default-node-nids-and-roles)
   - [Groups and Partitions](#groups-and-partitions)
6. [State Change Service - Monitoring Redfish Events](#state-change-service---monitoring-redfish-events)
7. [Additional Run Instructions](#additional-run-instructions)
   - [Running on Craystack or Other Node-Less Machine](#running-on-craystack-or-other-node-less-machine)
   - [Accessing Postgres Operator](#accessing-postgres-operator)
   - [Manually Adding State/Component Entries](#manually-adding-statecomponent-entries)
   - [Loading a Database Backup from a Real System](#loading-a-database-backup-from-a-real-system)

---

---


## SMD Capabilities
The SMD monitors and interrogates hardware components
in a HPC system, tracking hardware state and inventory information, and
making it available via REST queries and message bus events when changes occur.

This service provides the following functions:

* Performs inventory discovery of all known Mountain/River controllers in system
  * This process bootstraps everything HSM tracks.
  * Updated in response to component life-cycle changes to keep info current.
  * Interrogates Redfish controllers whenever provided by an endpoint discovery service:
     * REDS
     * MEDS
     * Or manually added to Hardware State Manager via REST
* Tracks hardware state and other high-level component data:
  * Hardware state (e.g. Empty/Off/On/Ready(heart beating))
  * Logical state  (i.e. managed plane software state)
  * Hardware Type  (i.e. Node, Chassis, BMC, Slot), as well as SubType
  * Enabled/Disabled status
  * Role, e.g. Compute vs. NCN type (nodes only)
  * NID, i.e. the numerical Node ID (nodes only)
  * Architecture
* Allows the creation of component groups and partitions
* Tracks default NIDs and Roles to use when discovering new nodes
* Stores Redfish endpoint and component data obtained via inventory discovery
  * Translates global Physical IDs (i.e. xnames) to local Redfish URIs.
  * Stores component-by-component Redfish properties to allow interactions
  * Stores needed boot-time information such as MACs
      * Actions for performing power operations via the managing Redfish EP
      * Service endpoint info, e.g. for performing firmware updates
* Learns, stores and tracks detailed hardware inventory data:
  * Tracks FRU data for manufacturing/service and scheduling (e.g. SLURM)
  * Tracks FRUs by location, showing the present composition of the system.
  * We include the most detailed information available on:
     * The manufacturer, model, revision, and so on describing exactly what is installed.
     * Serial numbers and any other identifying info to track individual pieces of HW
     * Properties of the hardware describing it's resources and abilities, for example: 
         * Speeds
         * Memory/storage capacities
         * Thread and core counts
         * The component family and the specific sub-type in use

## SMD Overview 

<img src="docs/images/StateManagerRedfish.png" alt="drawing" width="1045"/>

## SMD API

### API Overview

The main components of _smd_'s RESTful API are as follows:

#### Redfish Endpoints

```text
/hsm/v2/Inventory/RedfishEndpoints

    POST   A new RedfishEndpoint to be inventoried by state manager
    GET    The collection of all RedfishEndpoints, with optional filters.
    
/hsm/v2/Inventory/RedfishEndpoints/{xname-id}
    
    PUT    Updates to a RedfishEndpoint
    GET    The RedfishEndpoint's details or check its discovery status
    DELETE A RedfishEndpoint that is no longer in the system
```

#### Component Redfish endpoint information

```text
/hsm/v2/Inventory/ComponentsEndpoints?filter-option1=xxx...

    GET    Array of Redfish details for a filtered subset of components

/hsm/v2/Inventory/ComponentsEndpoints/{xname-id}

    GET    Redfish details for a specific component
```

#### Component State

```text
/hsm/v2/State/Components?filter-option1=xxx...

    GET    A filtered subset of all components, or a specific component by its id

/hsm/v2/State/Components

    POST   A list of individual component ids to query, and filtering options.

/hsm/v2/State/Components/{xname-id}

    GET    The HW state, flag, role, enabled status, NID, etc. of the component
    PATCH  The HW state, flag, role, enabled status, NID, etc. of a component
```

#### Additional Component State Queries

```text
/hsm/v2/State/Components/Query

    POST   A list of parents and filtering options

/hsm/v2/State/Components/Query/{parent-id}?filter-option1=xxx...

    GET    Parent and children of selected component parent id, optionally filtering
```

#### Hardware Inventory and FRUs

```text
/hsm/v2/Inventory/Hardware/Query/all

    GET    An xthwinv-like json representation of the system's hardware and FRUs.

/hsm/v2/Inventory/HardwareByFRU/{fru-id}

    GET    Details on a particular FRU by it's ID.

/hsm/v2/Inventory/HardwareByFRU/{fru-id}

    GET    Details on a particular FRU by it's ID.
```

#### NodeMaps

```text
/hsm/v2/Defaults/NodeMaps

    GET    All NodeMaps entries, with default NID and Role per xname
    POST   One or more new NodeMaps entries to be added or overwritten

/hsm/v2/Defaults/NodeMaps/{xname}

    GET    The default NID and Role for {xname}
    PUT    Update the default NID and Role for xname {xname}
```

#### Component Groups

```text
/hsm/v2/groups

    GET    Details on all groups
    POST   A new component group with a list of members
    
/hsm/v2/groups/{group-label}

    PATCH  Metadata for an existing group {group-label}
    GET    Details on the group {group-label}, i.e. it's members and metadata

/hsm/v2/groups/{group-label}/members

    GET    Just the member list for a group
    POST   The id of a new component to add to the group's members list
    
/hsm/v2/groups/{group-label}/members/{xname-id}

    DELETE Component {xname-id} from the members of the group {group-label}
```

#### Component Partitions

```text
/hsm/v2/partitions

    GET    Details on all partitions
    POST   A new partition with a list of members
    
/hsm/v2/partitions/{part-name}

    PATCH  Metadata for an existing partition {part-name}
    GET    Details on the partition {part-name}, i.e. it's members and metadata

/hsm/v2/partitions/{part-name}/members

    GET    Just the members list for a partition
    POST   The id of a new component to add to the partition's members list
    
/hsm/v2/partitions/{part-name}/members/{xname-id}

    DELETE Component {xname-id} from the members of partition {part-name}
```

#### Component Group and Partition Memberships

```text
/hsm/v2/memberships?filter-option1=xxx...

    GET    A filtered list of each system component's group/partition memberships

/hsm/v2/memberships/{xname-id}

    GET    The group and partition memberships (if any) of component {xname-id}
```

_NOTE The above is NOT an exhausive list of the API calls and is intended solely as an overview_

### Additional API Documentation

The complete HSM (smd) API documentation is included in the Cray API docs.
This is the nightly-generated version.  Content is generated in an automated
fashion from the current swagger.yaml file.

http://web.us.cray.com/~ekoen/cray-portal/public

Latest detailed API usage examples:

https://github.com/OpenCHAMI/smd/blob/master/docs/examples.adoc  (current)

Latest swagger.yaml (if you would prefer to use the OpenAPI viewer of your choice):

https://github.com/OpenCHAMI/smd/blob/master/api/swagger_v2.yaml (current)


## SMD Features

__________________________________________________________________

### Feature Map

This is primarily intended to compare XC-Shasta functionality.

| V1 Feature | V1+ Feature | XC Equivalent |
| --- | --- | --- |
| /hsm/v2/State/Components (structure) | - | rs_node_t |
| /hsm/v2/State/Components (GET) | - | xtcli status, but with all fields, including NIDs |
| /hsm/v2/State/Components/Query/<comp> (GET) | - | xtcli status <comp> |
| /hsm/v2/State/Components/<xname>/Enabled (PATCH) | - | xtcli enable/disable <cname> |
| /hsm/v2/State/Components/<xname>/Role (PATCH) | - | xtcli mark <role>(1) |
| /hsm/v2/State/Components/<xname>/StateData (PATCH) | - | xtcli set_empty -f <cname> |
| /hsm/v2/State/Components/<xname>/FlagOnly (PATCH) | - | xtcli set_flag/clr_flag <cname>(2) |
| /hsm/v2/Inventory/Hardware/Query/all (GET) | - | xthwinv s0 |
| /hsm/v2/Inventory/Discover (POST XNames list) | - | xtdiscover --warmswap <xnames>(3) |
| - | Additional Hardware/Query options | xthwinv <other-flags> |
| SCN Events | - | ec_node_(un)available, ec_node_failed |
| - | See Future Features and Updates Below| - |

(1) Role determines Compute vs NCN (Non-Compute Node) type, not Compute vs.
Service as on XC.

(2) Flags cleared automatically on successful state transition, so normally not
needed.

(3) There is no direct equivalent to the full xtdiscover command on Shasta.
Discovery is continuous in response to system events and works in concert with
endpoint discovery performed by MEDS and REDS, as well as system info provided by
(the upcoming) IDEALS.

### Current Features

* Inventory discovery of all known Mountain/River controllers in system
* Tracks hardware state and other high-level component data (like rs_node_t on XC)
* Allows the creation of component groups and partitions
* Tracks default NIDs and Roles to use when discovering new nodes
     * This can be used to distinguish Compute vs Non-Compute Nodes NCNs
* Stores Redfish endpoint and component data obtained via inventory discovery
  * This is used by HMS services such as CAPMC to interact with nodes via Redfish
* Monitors Redfish events from currently supported River/Mountain controller types
     * Looks at BMC alerts/events to track power on/off changes
     * Events vary by manufacturer/firmware version.
* Learns, stores and tracks detailed hardware inventory data.  Currently supported types are:
     * Mountain Chassis and Slots (Compute and Router)
     * Mountain cC and sC and associated board and node subcomponents
     * River BMCs and Nodes for Intel s2600
     * Processors and DIMMs
     * River RackPDUs.
* Sends SCN State Change Notifications via the HMI NFD for Component State and
  other field changes.

### Future Features And Updates

* Automatic warmswaps/removals/additions with no manual rediscovery
* Integration with IDEALS to:
     * Report empty components where missing/undiscovered
     * Retrieve info on manually configured components/endpoints
     * Get information on NID map, roles, etc. from CCD-based data.
     * And so on.
* Performance/scaling optimizations
* Inventory and State monitoring of additional component types
* Support for Gigabyte River nodes (discovery/events)
* Improved FRU tracking
* Historical Location<->FRU tracking
* FRU firmware version tracking
* Full featured push events for all hardware changes for HMS services and
  customer message bus.  (Having services poll HSM doesn't scale.)
* Cease storage of RedfishEndpoint credentials in DB, using vault instead.
* Use HMS specific Kafka bus for Redfish and HSM(smd) events to decouple from
  the telemetry/SMA message bus.
* Improved DB testing capabilities and allowing saving/restoring database
  settings for test/debug
* Dumping state for dump utility(?)



## Design Details
---

### States

Note that these are the States HSM directly has access to.  They are basically just the hardware states, with the Ready, et. al states above On being tracked by the heartbeat monitor is the case of nodes (and in the case of controllers, by HSM directly confirming that a component can be accessed for Redfish operations).  Other hardware types will generally go no higher than on.

A separate field, SoftwareStatus, is intended for any additional state that might exist for a heart beating node.   Note that we have no table of these states, nor a transition diagram, because these are a function of the managed plane and we do not limit what can appear there so that there are no dependencies created.
```text
StateUnknown   HMSState = "Unknown"   // The State is unknown.  Appears missing but has not been confirmed as empty.
StateEmpty     HMSState = "Empty"     // The location is not populated with a component
StatePopulated HMSState = "Populated" // Present (not empty), but no further track can or is being done.
StateOff       HMSState = "Off"       // Present but powered off
StateOn        HMSState = "On"        // Powered on.  If no heartbeat mechanism is available, it's software state may be unknown.

StateStandby   HMSState = "Standby" // No longer Ready and presumed dead.  It typically means HB has been lost (w/alert).
StateHalt      HMSState = "Halt"    // No longer Ready and halted.  OS has been gracefully shutdown or panicked (w/ alert).
StateReady     HMSState = "Ready"   // Both On and Ready to provide its expected services, i.e. used for jobs.
```

#### State Transition Rules

To avoid undesirable behavior (bad ordering, invalid states), only certain state transitions are allowed based upon events or REST operations. 

Note that the inventory discovery process has the ability to perform any state change, e.g. when a new component is added or is powered on after appearing to disappear from the system.

Desired new state    -     Required current state
```text
"Unknown":   {}, // Force/HSM-internal only
"Empty":     {}, // Force/HSM-internal only
"Populated": {}, // Force/HSM-internal only
"Off":       {StateOff, StateOn, StateStandby, StateHalt, StateReady},
"On":        {StateOn, StateOff, StateStandby, StateHalt, StateReady},
"Standby":   {StateStandby, StateReady},
"Halt":      {StateHalt, StateReady},
"Ready":     {StateReady, StateOn},
```

### Setting Default Node NIDs and Roles

***Setting default NIDs***

This document describes and provides examples for the Hardware State Manager 
smd) NodeMaps feature, which allows the installer (or a user via the HSM REST
API) to pre-populate default NID assignments for node locations in the system.
These are then used a node with an xname matches the NodeMaps entry of the same
name, setting the correct NID and Role values from the start, and making
it unnecessary to manually patch them.

See: https://connect.us.cray.com/confluence/display/HSOS/HSM+Documentation+for+SPS2+-+Setting+Default+NIDs

### Groups and Partitions

***Groups (or Labels)***

Are named sets of system components, most commonly nodes.   Each component may
belong to any number of groups.  Groups can be created freely, and smd does not
assign them any particular predetermined meaning. 

***Partitions***

Are essentially a kind of group, but have an established meaning and are
treated as distinct entities from groups.  Each component may belong to at most
one partition, and partitions are used as an access control mechanism. 

#### Goals of Group and Partition functionality

1. To allow the creation, deletion and modification of groups and partitions, as well as the members thereof
2. To allow groups and partitions to be used as filtering mechanisms for existing smd API calls.   
       * Filtering by partition should be optimized for use as an access control mechanism.

#### Groups and Partitions Design Document:

See: https://connect.us.cray.com/confluence/display/HSOS/Hardware+State+Manager+Group+and+Partition+Service


### State Change Service - Monitoring Redfish Events

hmcollector polls smd periodically and establishes event subscriptions when
new RedfishEndpoints are found.  These events are used for power state changes
and they are POSTed to a kafka bus (currently the telemetry bus) that smd
then monitors.

When an event comes in, smd establishes the sending BMC and then looks
up (via the ComponentEndpoints) the path of the subcomponent URI referenced in
the payload in order to establish, for example, which of the two nodes under a
node controller is the one powering on or off.

The event payloads used can vary from Redfish implementation to implementation.
In some cases, these are "Alert" type events that are more or less a
repackaging of the underlying iPMI alert message.  In other cases, the
controller may use the standard ResourceEvent registry, where the intent is
to report on Redfish Status field (and other) changes in a more generic way.

***State Change Notification Infrastructure***

<img src="docs/images/SCNforHSM.png" alt="drawing2" width="900"/>


## Additional Run Instructions 

#### Running on Craystack or Other Node-Less Machine

Running a plain docker container is not really practical in a full helm-based
deployment because of the lack of integration and features that are provided
via helm.

The easiest way to add nodes via discovery is to find nodes that have
externally visible IP addresses for their BMCs.

I've not gotten the socks5 method above to work in kubernetes, but it might
be something simple.  Logging into each cray-smd pod using kubectl exec 
and doing "apk add openssh" will allow you to install ssh and use it to
connect to external hosts, however the -D option gives an error logging in.
In any case, you would have to reroll the values.yaml helm chart for
cray-smd (and incrememnt the version number in Chart.yaml) to add the
SMD_PROXY env variable (see above)

***Accessing Postgres Operator***

You can access the postgres database cluster as follows:

```text
sms-1:~ # kubectl get pods -n services | grep smd
sms-1:~ # kubectl exec -it -n services cray-smd-{id-from-previous} -- /bin/sh

/ # cat /secrets/postgres/hmsdsuser/password   # Copy to clipboard
/ # psql hmsds hmsdsuser -h cray-smd-postgres-0 -W 
(Paste password)
```

At this point, you have two options.

1. Use the following instructions to dump (from a real machine) a db backup and restore it on craystack if the db is fresh and empty.
```text
https://www.postgresql.org/docs/11/backup-dump.html
```
2. Use the instructions below to manually create a few Components (limited functonality) via the pgsl client directly.

#### Manually Adding State/Component Entries

WARNING:  This only works for some kinds of testing as it creates incomplete configurations (but should work with group, partition, and State/Components calls, though remember that normally there will be non-Node components you will need to filter out if that's desired).

Moreover, even in this case, adding entries manually can create a corrupt database even if data is correct but improperly normalized.  Best to not stray from the example below except to change the slot number in the xname and the NID.  

1. Access the database on one of the postgres containers, as described above, from the cray-smd pod:
```text
(cray-smd-container) # psql hmsds hmsdsuser -h cray-smd-postgres-0 -W
<enter password from /secrets/postgres/hmsdsuser/password, see above>
```
2. Create Nodes one at a time with the following template (follow carefully, note '123' is the NID).
```text
hmsds=> insert INTO components (id,type,state,flag,enabled,admin,role,nid,subtype,nettype,arch) VALUES('x0c0s0b0n0','Node','Empty','OK',true,'','Compute',123,'','Sling','X86');
```
3. If you get the following, the insert worked and you are on the primary node for the postgres cluster.
```text
INSERT 0 1  # <- SUCCESS, running on primary
```
4. If you get the following error, repeat steps 1 and 2 with -h cray-smd-postgres-1 and, if needed, -h cray-smd-postgres-2.  Password is the same.
```text
ERROR:  cannot execute INSERT in a read-only transaction   # <- Not on primary postgres pod
```

##### Additional psql commands to help with adjusting the contents of the HSM database

* Quitting the psql editor

```text
hmsds=> \q
```

* Display contents of a  table

```text
hmsds=> SELECT * FROM table_name;
```

* Display contents of the components table

```text
hmsds=> SELECT * FROM components;
```

* Display only components of type 'Node' from the components table

```text
hmsds=> SELECT * FROM components
hmsds-> WHERE type = 'Node';
```

* Delete a node from the components table; Necessary precursor to re-adding the same node with different values

```text
hmsds=> DELETE FROM components                                                                                                                                                                                                                                                WHERE id = 'x3000c0s19b1n0';
```

* Create a node in the components table with the given attributes

```text
hmsds=> insert INTO components (id,type,state,flag,enabled,admin,role,nid,subtype,nettype,arch) VALUES('x0c0s0b0n0','Node','Empty','OK',true,'','Compute',123,'','Sling','X86');
```

* List available tables

```text
hmsds=> \dt
```

* Describe a table

```text
hmsds=> \d table_name
```

* Command history

```text
hmsds=> \s
```

* Get help on psql commands

```text
hmsds=> \h
```

#### Loading a Database Backup from a Real System:

I don't have step-by-step instructions but the basic idea is to:

1. At some point use the above instructions to access postgres on a real system and run pg_dump.
2. On Craystack, access postgres on the primary node after smd is first installed and load the dump

For more info: https://www.postgresql.org/docs/11/backup-dump.html

Note if it is easier, you can run pg_dump/psql on the sms, but you must use 
the IP, not the hostname, of the cray-smd-postgres-[012] service.  Running
kubectl get services -n services will get you this IP.
