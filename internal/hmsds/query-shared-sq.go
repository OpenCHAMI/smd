// MIT License
//
// (C) Copyright [2019-2022] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package hmsds

// This is a newer version of query-shared that used squirrel to form
// queries.   Over time this should take over and replace query-shared.go
// as its content as no longer needed.

import (
	"database/sql"
	"strings"

	"github.com/Cray-HPE/hms-xname/xnametypes"

	sq "github.com/Masterminds/squirrel"
)

//                                                                          //
//                              System Table                                //
//                                                                          //

const sysTable = "system"

const (
	sysIdCol            = "id"
	sysSchemaVersionCol = "schema_version"
	sysSystemInfoCol    = "system_info"
)

//                                                                          //
//                           Component structs                              //
//                                                                          //

const compTable = `components`

const compTableJoinAlias = `c`
const compTableSubAlias = `comp`

const (
	compIdCol          = `id`
	compTypeCol        = `type`
	compStateCol       = `state`
	compFlagCol        = `flag`
	compEnabledCol     = `enabled`
	compSwStatusCol    = `admin`
	compRoleCol        = `role`
	compSubRoleCol     = `subrole`
	compNIDCol         = `nid`
	compSubTypeCol     = `subtype`
	compNetTypeCol     = `nettype`
	compArchCol        = `arch`
	compClassCol       = `class`
	compResDisabledCol = `reservation_disabled`
	compLockedCol      = `locked`
)

var compColsNamesAll = []string{
	compIdCol,
	compTypeCol,
	compStateCol,
	compFlagCol,
	compEnabledCol,
	compSwStatusCol,
	compRoleCol,
	compSubRoleCol,
	compNIDCol,
	compSubTypeCol,
	compNetTypeCol,
	compArchCol,
	compClassCol,
	compResDisabledCol,
	compLockedCol,
}

// With added group fields..
var compGroupPartCols = []string{
	compGroupNameCol,
	compGroupNamespaceCol,
}

//
// Queries for various Components column filter options
//

//  FLTR_DEFAULT
var compColsDefault = []string{
	compIdCol,
	compTypeCol,
	compStateCol,
	compFlagCol,
	compEnabledCol,
	compSwStatusCol,
	compRoleCol,
	compSubRoleCol,
	compNIDCol,
	compSubTypeCol,
	compNetTypeCol,
	compArchCol,
	compClassCol,
	compResDisabledCol,
	compLockedCol,
}

//	FLTR_STATEONLY
var compColsStateOnly = []string{
	compIdCol,
	compTypeCol,
	compStateCol,
	compFlagCol,
}

//	FLTR_FLAGONLY
var compColsFlagOnly = []string{
	compIdCol,
	compTypeCol,
	compFlagCol,
}

//	FLTR_ROLEONLY
var compColsRoleOnly = []string{
	compIdCol,
	compTypeCol,
	compRoleCol,
	compSubRoleCol,
}

//	FLTR_NIDONLY
var compColsNIDOnly = []string{
	compIdCol,
	compTypeCol,
	compNIDCol,
}

//	FLTR_ID_ONLY
var compColsIdOnly = []string{
	compIdCol,
}

// These two combine group-related columns in addition to the standard
// Component ones.

//	FLTR_ALL_W_GROUP
var compColsAllWithGroup1 []string = compColsDefault
var compColsAllWithGroup2 []string = compGroupPartCols

//	FLTR_ID_W_GROUP
var compColsIdWithGroup1 []string = compColsIdOnly
var compColsIdWithGroup2 []string = compGroupPartCols

//                                                                          //
//                        Groups and partitions                             //
//                                                                          //

// component_groups table

const compGroupsTable = `component_groups`
const compGroupsAlias = `g` // used during joins, i.e. g.name

const (
	compGroupIdCol        = `id`
	compGroupNameCol      = `name`
	compGroupDescCol      = `description`
	compGroupTagsCol      = `tags`
	compGroupTypeCol      = `type`
	compGroupNamespaceCol = `namespace`
	compGroupExGrpCol     = `exclusive_group_identifier`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	compGroupIdColAlias        = compGroupsAlias + "." + compGroupIdCol
	compGroupNameColAlias      = compGroupsAlias + "." + compGroupNameCol
	compGroupDescColAlias      = compGroupsAlias + "." + compGroupDescCol
	compGroupTagsColAlias      = compGroupsAlias + "." + compGroupTagsCol
	compGroupTypeColAlias      = compGroupsAlias + "." + compGroupTypeCol
	compGroupNamespaceColAlias = compGroupsAlias + "." + compGroupNamespaceCol
	compGroupExGrpColAlias     = compGroupsAlias + "." + compGroupExGrpCol
)

// These are the namespace enums used in the DB
const groupNamespace = `group`
const partNamespace = `partition`

// These are the group type enums used in the DB
const groupType = `shared`
const exclGroupType = `exclusive`
const partType = `partition`

// component_groups table columns.
// Skip 'annotations' for now, future use only
var compGroupsColsAll7 = []string{compGroupIdCol, compGroupNameCol,
	compGroupDescCol, compGroupTagsCol, compGroupTypeCol,
	compGroupNamespaceCol, compGroupExGrpCol}

// Columns that go in the group structure plus (uu)id
var compGroupsColsSMGroup = []string{compGroupIdCol, compGroupNameCol,
	compGroupDescCol, compGroupTagsCol, compGroupExGrpCol}

// Columns that go in the partition structure plus (uu)id
var compGroupsColsSMPart = []string{compGroupIdCol, compGroupNameCol,
	compGroupDescCol, compGroupTagsCol}

type compGroupsInsert struct {
	id               string
	name             string
	description      string
	tags             []string // json encoded array
	gtype            string   // from group type enum
	namespace        string   // from namespace enum
	exclusiveGroupId string
}

// component_group_members table

const compGroupMembersTable = `component_group_members`
const compGroupMembersAlias = `gm` // used during joins, i.e. gm.group_id

const (
	compGroupMembersCmpIdCol = `component_id`
	compGroupMembersGrpIdCol = `group_id`
	compGroupMembersNsCol    = `group_namespace`
	compGroupMembersJTimeCol = `joined_at`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	compGroupMembersCmpIdColAlias = compGroupMembersAlias + "." + compGroupMembersCmpIdCol
	compGroupMembersGrpIdColAlias = compGroupMembersAlias + "." + compGroupMembersGrpIdCol
	compGroupMembersNsColAlias    = compGroupMembersAlias + "." + compGroupMembersNsCol
	compGroupMembersJTimeColAlias = compGroupMembersAlias + "." + compGroupMembersJTimeCol
)

// component_group_members table - all columns
var compGroupMembersColsAll4 = []string{compGroupMembersCmpIdCol,
	compGroupMembersGrpIdCol, compGroupMembersNsCol, compGroupMembersJTimeCol}

// component_group_members table - required columns.
// joined_at (timestamp) omitted.  Not needed for insert, and some queries
var compGroupMembersColsNoTS = []string{compGroupMembersCmpIdCol,
	compGroupMembersGrpIdCol, compGroupMembersNsCol}

const partGroupNamespace = `%%partition%%`

type compGroupMembersInsertNoTS struct {
	component_id    string
	group_id        string
	group_namespace string
}

// Get only user visible columns from component_group_members
var compGroupMembersColsUser = []string{compGroupMembersCmpIdCol}

//                                                                           //
//                            Component Locks V2                             //
//                                                                           //

// reservations table

const compResTable = `reservations`
const compResAlias = `cr` // used during joins, i.e. cr.component.id

const (
	compResCompIdCol   = `component_id`
	compResCreatedCol  = `create_timestamp`
	compResExpireCol   = `expiration_timestamp`
	compResDKCol       = `deputy_key`
	compResRKCol       = `reservation_key`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	compResCompIdColAlias   = compResAlias + "." + compResCompIdCol
	compResCreatedColAlias  = compResAlias + "." + compResCreatedCol
	compResExpireColAlias   = compResAlias + "." + compResExpireCol
	compResDKColAlias       = compResAlias + "." + compResDKCol
	compResRKColAlias       = compResAlias + "." + compResRKCol
)

// reservations table columns.
var compResCols = []string{compResCompIdCol, compResCreatedCol,
	compResExpireCol, compResDKCol, compResRKCol}

// reservations table public columns.
var compResPubCols = []string{compResCompIdCol, compResCreatedCol,
	compResExpireCol, compResDKCol}

type compReservation struct {
	component_id         string
	create_timestamp     sql.NullTime
	expiration_timestamp sql.NullTime
	deputy_key           string
	reservation_key      string
}

//                                                                          //
//                        RedfishEndpoint structs                           //
//                                                                          //

const rfEPsTable = `rf_endpoints`
const rfEPsAlias = `rf`

const rfEPsTableAlias = rfEPsAlias + "." + rfEPsTable

const (
	rfEPsIdCol             = `id`
	rfEPsTypeCol           = `type`
	rfEPsNameCol           = `name`
	rfEPsHostnameCol       = `hostname`
	rfEPsDomainCol         = `domain`
	rfEPsFQDNCol           = `fqdn`
	rfEPsEnabledCol        = `enabled`
	rfEPsUUIDCol           = `uuid`
	rfEPsUserCol           = `"user"`
	rfEPsPasswordCol       = `password`
	rfEPsUseSSDPCol        = `usessdp`
	rfEPsMacRequiredCol    = `macrequired`
	rfEPsMacAddrCol        = `macaddr`
	rfEPsIPAddrCol         = `ipaddr`
	rfEPsRediscOnUpdateCol = `rediscoveronupdate`
	rfEPsTemplateIDCol     = `templateid`
	rfEPsDiscInfoCol       = `discovery_info`
)

const (
	rfEPsIdColAlias             = rfEPsAlias + "." + rfEPsIdCol
	rfEPsTypeColAlias           = rfEPsAlias + "." + rfEPsTypeCol
	rfEPsNameColAlias           = rfEPsAlias + "." + rfEPsNameCol
	rfEPsHostnameColAlias       = rfEPsAlias + "." + rfEPsHostnameCol
	rfEPsDomainColAlias         = rfEPsAlias + "." + rfEPsDomainCol
	rfEPsFQDNColAlias           = rfEPsAlias + "." + rfEPsFQDNCol
	rfEPsEnabledColAlias        = rfEPsAlias + "." + rfEPsEnabledCol
	rfEPsUUIDColAlias           = rfEPsAlias + "." + rfEPsUUIDCol
	rfEPsUserColAlias           = rfEPsAlias + "." + rfEPsUserCol
	rfEPsPasswordColAlias       = rfEPsAlias + "." + rfEPsPasswordCol
	rfEPsUseSSDPColAlias        = rfEPsAlias + "." + rfEPsUseSSDPCol
	rfEPsMacRequiredColAlias    = rfEPsAlias + "." + rfEPsMacRequiredCol
	rfEPsMacAddrColAlias        = rfEPsAlias + "." + rfEPsMacAddrCol
	rfEPsIPAddrColAlias         = rfEPsAlias + "." + rfEPsIPAddrCol
	rfEPsRediscOnUpdateColAlias = rfEPsAlias + "." + rfEPsRediscOnUpdateCol
	rfEPsTemplateIDColAlias     = rfEPsAlias + "." + rfEPsTemplateIDCol
	rfEPsDiscInfoColAlias       = rfEPsAlias + "." + rfEPsDiscInfoCol
)

var rfEPsAllColsNoStatus = []string{
	rfEPsIdCol,
	rfEPsTypeCol,
	rfEPsNameCol,
	rfEPsHostnameCol,
	rfEPsDomainCol,
	rfEPsFQDNCol,
	rfEPsEnabledCol,
	rfEPsUUIDCol,
	rfEPsUserCol,
	rfEPsPasswordCol,
	rfEPsUseSSDPCol,
	rfEPsMacRequiredCol,
	rfEPsMacAddrCol,
	rfEPsIPAddrCol,
	rfEPsRediscOnUpdateCol,
	rfEPsTemplateIDCol,
}

var rfEPsAllCols = append(rfEPsAllColsNoStatus, rfEPsDiscInfoCol)

//                                                                          //
//                        ComponentEndpoint structs                         //
//                                                                          //

const compEPsTable = `comp_endpoints`
const compEPsAlias = `ce`

const (
	compEPsIdCol             = `id`
	compEPsTypeCol           = `type`
	compEPsDomainCol         = `domain`
	compEPsRedfishTypeCol    = `redfish_type`
	compEPsRedfishSubtypeCol = `redfish_subtype`
	compEPsRFEndpointIDCol   = `rf_endpoint_id`
	compEPsMACCol            = `mac`
	compEPsUUIDCol           = `uuid`
	compEPsODataIDCol        = `odata_id`
	compEPsComponentInfoCol  = `component_info`
)

const (
	compEPsIdColAlias             = compEPsAlias + "." + compEPsIdCol
	compEPsTypeColAlias           = compEPsAlias + "." + compEPsTypeCol
	compEPsDomainColAlias         = compEPsAlias + "." + compEPsDomainCol
	compEPsRedfishTypeColAlias    = compEPsAlias + "." + compEPsRedfishTypeCol
	compEPsRedfishSubtypeColAlias = compEPsAlias + "." + compEPsRedfishSubtypeCol
	compEPsRFEndpointIDColAlias   = compEPsAlias + "." + compEPsRFEndpointIDCol
	compEPsMACColAlias            = compEPsAlias + "." + compEPsMACCol
	compEPsUUIDColAlias           = compEPsAlias + "." + compEPsUUIDCol
	compEPsODataIDColAlias        = compEPsAlias + "." + compEPsODataIDCol
	compEPsComponentInfoColAlias  = compEPsAlias + "." + compEPsComponentInfoCol
)

var compEPsAllCols = []string{
	compEPsIdCol,
	compEPsTypeCol,
	compEPsDomainCol,
	compEPsRedfishTypeCol,
	compEPsRedfishSubtypeCol,
	compEPsRFEndpointIDCol,
	compEPsMACCol,
	compEPsUUIDCol,
	compEPsODataIDCol,
	compEPsComponentInfoCol,
}

//                                                                          //
//                         ServiceEndpoint structs                          //
//                                                                          //

const serviceEPsTable = `service_endpoints`
const serviceEPsAlias = `se`

const (
	serviceEPsRFEndpointIDCol   = `rf_endpoint_id`
	serviceEPsRedfishTypeCol    = `redfish_type`
	serviceEPsRedfishSubtypeCol = `redfish_subtype`
	serviceEPsUUIDCol           = `uuid`
	serviceEPsODataIDCol        = `odata_id`
	serviceEPsServiceInfoCol    = `service_info`
)

const (
	serviceEPsRFEndpointIDColAlias   = serviceEPsAlias + "." + serviceEPsRFEndpointIDCol
	serviceEPsRedfishTypeColAlias    = serviceEPsAlias + "." + serviceEPsRedfishTypeCol
	serviceEPsRedfishSubtypeColAlias = serviceEPsAlias + "." + serviceEPsRedfishSubtypeCol
	serviceEPsUUIDColAlias           = serviceEPsAlias + "." + serviceEPsUUIDCol
	serviceEPsODataIDColAlias        = serviceEPsAlias + "." + serviceEPsODataIDCol
	serviceEPsServiceInfoColAlias    = serviceEPsAlias + "." + serviceEPsServiceInfoCol
)

var serviceEPsCols = []string{
	serviceEPsRFEndpointIDCol,
	serviceEPsRedfishTypeCol,
	serviceEPsRedfishSubtypeCol,
	serviceEPsUUIDCol,
	serviceEPsODataIDCol,
	serviceEPsServiceInfoCol,
}

//                                                                          //
//                      Component Ethernet Interfaces                       //
//                                                                          //

const compEthTable = `comp_eth_interfaces`
const compEthAlias = `cei`

const (
	compEthIdCol          = `id`
	compEthDescCol        = `description`
	compEthMACAddrCol     = `macaddr`
	compEthLastUpdateCol  = `last_update`
	compEthCompIDCol      = `compid`
	compEthTypeCol        = `comptype`
	compEthIPAddressesCol = `ip_addresses`

	// JSON Blob keys
	compEthJsonIPAddress = `IPAddress`
	compEthJsonNetwork   = `Network`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	compEthIdColAlias         = compEthAlias + "." + compEthIdCol
	compEthDescColAlias       = compEthAlias + "." + compEthDescCol
	compEthMACAddrColAlias    = compEthAlias + "." + compEthMACAddrCol
	compEthLastUpdateColAlias = compEthAlias + "." + compEthLastUpdateCol
	compEthCompIDColAlias     = compEthAlias + "." + compEthCompIDCol
	compEthTypeColAlias       = compEthAlias + "." + compEthTypeCol
	compEthIPAddressesAlias   = compEthAlias + "." + compEthIPAddressesCol
)

// compEthTable table columns.
var compEthCols = []string{compEthIdCol, compEthDescCol,
	compEthMACAddrCol, compEthLastUpdateCol,
	compEthCompIDCol, compEthTypeCol, compEthIPAddressesCol}

var compEthColsNoTS = []string{compEthIdCol, compEthDescCol,
	compEthMACAddrCol, compEthCompIDCol,
	compEthTypeCol, compEthIPAddressesCol}

//                                                                          //
//                             HwInv structs                                //
//                                                                          //

const hwInvTable = `hwinv_by_loc_with_fru`
const hwInvAlias = `loc`

const (
	hwInvIdCol         = `id`
	hwInvTypeCol       = `type`
	hwInvOrdCol        = `ordinal`
	hwInvStatusCol     = `status`
	hwInvLocInfoCol    = `location_info`
	hwInvFruIdCol      = `fru_id`
	hwInvFruTypeCol    = `fru_type`
	hwInvFruSubTypeCol = `fru_subtype`
	hwInvFruInfoCol    = `fru_info`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	hwInvIdColAlias         = hwInvAlias + "." + hwInvIdCol
	hwInvTypeColAlias       = hwInvAlias + "." + hwInvTypeCol
	hwInvOrdColAlias        = hwInvAlias + "." + hwInvOrdCol
	hwInvStatusColAlias     = hwInvAlias + "." + hwInvStatusCol
	hwInvLocInfoColAlias    = hwInvAlias + "." + hwInvLocInfoCol
	hwInvFruIdColAlias      = hwInvAlias + "." + hwInvFruIdCol
	hwInvFruTypeColAlias    = hwInvAlias + "." + hwInvFruTypeCol
	hwInvFruSubTypeColAlias = hwInvAlias + "." + hwInvFruSubTypeCol
	hwInvFruInfoColAlias    = hwInvAlias + "." + hwInvFruInfoCol
)

// hwInv table columns.
var hwInvCols = []string{hwInvIdCol, hwInvTypeCol,
	hwInvOrdCol, hwInvStatusCol, hwInvLocInfoCol,
	hwInvFruIdCol, hwInvFruTypeCol, hwInvFruSubTypeCol,
	hwInvFruInfoCol}

//
// HwInv with partition info
//

const hwInvPartTable = `hwinv_by_loc_with_partition`
const hwInvPartAlias = hwInvAlias

const (
	hwInvPartIdCol         = hwInvIdCol
	hwInvPartTypeCol       = hwInvTypeCol
	hwInvPartOrdCol        = hwInvOrdCol
	hwInvPartStatusCol     = hwInvStatusCol
	hwInvPartLocInfoCol    = hwInvLocInfoCol
	hwInvPartFruIdCol      = hwInvFruIdCol
	hwInvPartFruTypeCol    = hwInvFruTypeCol
	hwInvPartFruSubTypeCol = hwInvFruSubTypeCol
	hwInvPartFruInfoCol    = hwInvFruInfoCol
	hwInvPartPartitionCol  = `partition`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	hwInvPartIdColAlias         = hwInvPartAlias + "." + hwInvPartIdCol
	hwInvPartTypeColAlias       = hwInvPartAlias + "." + hwInvPartTypeCol
	hwInvPartOrdColAlias        = hwInvPartAlias + "." + hwInvPartOrdCol
	hwInvPartStatusColAlias     = hwInvPartAlias + "." + hwInvPartStatusCol
	hwInvPartLocInfoColAlias    = hwInvPartAlias + "." + hwInvPartLocInfoCol
	hwInvPartFruIdColAlias      = hwInvPartAlias + "." + hwInvPartFruIdCol
	hwInvPartFruTypeColAlias    = hwInvPartAlias + "." + hwInvPartFruTypeCol
	hwInvPartFruSubTypeColAlias = hwInvPartAlias + "." + hwInvPartFruSubTypeCol
	hwInvPartFruInfoColAlias    = hwInvPartAlias + "." + hwInvPartFruInfoCol
	hwInvPartPartitionColAlias  = hwInvPartAlias + "." + hwInvPartPartitionCol
)

// hwInv table columns.
var hwInvPartCols = []string{hwInvPartIdCol, hwInvPartTypeCol,
	hwInvPartOrdCol, hwInvPartStatusCol, hwInvPartLocInfoCol,
	hwInvPartFruIdCol, hwInvPartFruTypeCol, hwInvPartFruSubTypeCol,
	hwInvPartFruInfoCol, hwInvPartPartitionCol}

//
// HwInv by loc only
//

const hwInvLocTable = `hwinv_by_loc`
const hwInvLocAlias = `l`

const (
	hwInvLocIdCol      = `id`
	hwInvLocTypeCol    = `type`
	hwInvLocOrdCol     = `ordinal`
	hwInvLocStatusCol  = `status`
	hwInvLocParentCol  = `parent`
	hwInvLocLocInfoCol = `location_info`
	hwInvLocFruIdCol   = `fru_id`
	hwInvLocNodeCol    = `parent_node`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	hwInvLocIdColAlias      = hwInvLocAlias + "." + hwInvLocIdCol
	hwInvLocTypeColAlias    = hwInvLocAlias + "." + hwInvLocTypeCol
	hwInvLocOrdColAlias     = hwInvLocAlias + "." + hwInvLocOrdCol
	hwInvLocStatusColAlias  = hwInvLocAlias + "." + hwInvLocStatusCol
	hwInvLocParentColAlias  = hwInvLocAlias + "." + hwInvLocParentCol
	hwInvLocLocInfoColAlias = hwInvLocAlias + "." + hwInvLocLocInfoCol
	hwInvLocFruIdColAlias   = hwInvLocAlias + "." + hwInvLocFruIdCol
	hwInvLocNodeColAlias    = hwInvLocAlias + "." + hwInvLocNodeCol
)

// hwInv by loc only table columns.
var hwInvLocCols = []string{hwInvLocIdCol, hwInvLocTypeCol,
	hwInvLocOrdCol, hwInvLocStatusCol, hwInvLocNodeCol,
	hwInvLocLocInfoCol, hwInvLocFruIdCol}

//
// HwInv by FRU only
//

const hwInvFruTable = `hwinv_by_fru`
const hwInvFruAlias = `fru`

const (
	hwInvFruTblIdCol           = `fru_id`
	hwInvFruTblTypeCol         = `type`
	hwInvFruTblSubTypeCol      = `subtype`
	hwInvFruTblSerialCol       = `serial_number`
	hwInvFruTblPartCol         = `part_number`
	hwInvFruTblManufacturerCol = `manufacturer`
	hwInvFruTblInfoCol         = `fru_info`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	hwInvFruTblIdColAlias           = hwInvFruAlias + "." + hwInvFruTblIdCol
	hwInvFruTblTypeColAlias         = hwInvFruAlias + "." + hwInvFruTblTypeCol
	hwInvFruTblSubTypeColAlias      = hwInvFruAlias + "." + hwInvFruTblSubTypeCol
	hwInvFruTblSerialColAlias       = hwInvFruAlias + "." + hwInvFruTblSerialCol
	hwInvFruTblPartColAlias         = hwInvFruAlias + "." + hwInvFruTblPartCol
	hwInvFruTblManufacturerColAlias = hwInvFruAlias + "." + hwInvFruTblManufacturerCol
	hwInvFruTblInfoColAlias         = hwInvFruAlias + "." + hwInvFruTblInfoCol
)

// hwInv by loc only table columns.
var hwInvFruTblCols = []string{hwInvFruTblIdCol, hwInvFruTblTypeCol,
	hwInvFruTblSubTypeCol, hwInvFruTblInfoCol}

var hwInvFruTblColsAll = []string{hwInvFruTblIdCol, hwInvFruTblTypeCol,
	hwInvFruTblSubTypeCol, hwInvFruTblSerialCol, hwInvFruTblPartCol,
	hwInvFruTblManufacturerCol, hwInvFruTblInfoCol}

//                                                                          //
//                         HwInv History structs                            //
//                                                                          //

const hwInvHistTable = `hwinv_hist`
const hwInvHistAlias = `h`

const (
	hwInvHistIdCol        = `id`
	hwInvHistFruIdCol     = `fru_id`
	hwInvHistEventTypeCol = `event_type`
	hwInvHistTimestampCol = `timestamp`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	hwInvHistIdColAlias        = hwInvHistAlias + "." + hwInvHistIdCol
	hwInvHistFruIdColAlias     = hwInvHistAlias + "." + hwInvHistFruIdCol
	hwInvHistEventTypeColAlias = hwInvHistAlias + "." + hwInvHistEventTypeCol
	hwInvHistTimestampColAlias = hwInvHistAlias + "." + hwInvHistTimestampCol
)

// hwInvHist table columns.
var hwInvHistCols = []string{hwInvHistIdCol, hwInvHistFruIdCol,
	hwInvHistEventTypeCol, hwInvHistTimestampCol}

var hwInvHistColsNoTS = []string{hwInvHistIdCol, hwInvHistFruIdCol,
	hwInvHistEventTypeCol}

//                                                                           //
//                                 Job Sync                                  //
//                                                                           //

// job_sync table

const jobTable = `job_sync`
const jobAlias = `jq` // used during joins, i.e. jq.id

const (
	jobIdCol         = `id`
	jobTypeCol       = `type`
	jobStatusCol     = `status`
	jobLastUpdateCol = `last_update`
	jobLifetimeCol   = `lifetime`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	jobIdColAlias         = jobAlias + "." + jobIdCol
	jobTypeColAlias       = jobAlias + "." + jobTypeCol
	jobStatusColAlias     = jobAlias + "." + jobStatusCol
	jobLastUpdateColAlias = jobAlias + "." + jobLastUpdateCol
	jobLifetimeColAlias   = jobAlias + "." + jobLifetimeCol
)

// job_sync table columns.
var jobCols = []string{jobIdCol, jobTypeCol,
	jobStatusCol, jobLastUpdateCol, jobLifetimeCol}

// job_sync table - required columns.
// last_update (timestamp) omitted.  Not needed for insert
var jobColsNoTS = []string{jobIdCol, jobTypeCol,
	jobStatusCol, jobLifetimeCol}

type jobInsert struct {
	id         string
	jobType    string
	status     string
	lastUpdate string
	lifetime   int
}

type jobInsertNoTS struct {
	id       string
	jobType  string
	status   string
	lifetime int
}

//                                                                           //
//                       State Redfish Polling Jobs                          //
//                                                                           //

// Get only the component id column from job_comp_state_rf_poll
var stateRfPollColsId = []string{stateRfPollCmpIdCol}

// job_comp_state_rf_poll table

const stateRfPollTable = `job_state_rf_poll`
const stateRfPollAlias = `srfp` // used during joins, i.e. srfp.job_id

const (
	stateRfPollCmpIdCol = `comp_id`
	stateRfPollJobIdCol = `job_id`
)

// This adds the base table alias to each column.  it can later be appended to.
const (
	stateRfPollCmpIdColAlias = stateRfPollAlias + "." + stateRfPollCmpIdCol
	stateRfPollJobIdColAlias = stateRfPollAlias + "." + stateRfPollJobIdCol
)

// job_comp_state_rf_poll table - all columns
var stateRfPollCols = []string{stateRfPollCmpIdCol,
	stateRfPollJobIdCol}

type stateRfPollInsert struct {
	comp_id string
	job_id  string
}

////////////////////////////////////////////////////////////////////////////
//
// Helper functions - Query building
//
////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////
// SystemTable queries
////////////////////////////////////////////////////////////////////////////

// Return sq.SelectBuilder for querying System table for schema_version
func selectSystemSchemaVersion(id int) sq.SelectBuilder {
	query := sq.Select(sysSchemaVersionCol).From(sysTable).
		Where(sq.Eq{sysIdCol: []int{id}})

	return query
}

////////////////////////////////////////////////////////////////////////////
// Components/Memberships queries
////////////////////////////////////////////////////////////////////////////

// Get correct initial select statement containing the columns requested
// by the FieldFilter.  Add the appropriate alias to the default values.
func selectComponentCols(fltr FieldFilter, alias, from string) sq.SelectBuilder {
	var query sq.SelectBuilder

	columns := []string{}
	aliasCG := alias + compGroupsAlias

	switch fltr {
	case FLTR_STATEONLY:
		columns = addAliasToCols(alias, compColsStateOnly, compColsStateOnly)
	case FLTR_FLAGONLY:
		columns = addAliasToCols(alias, compColsFlagOnly, compColsFlagOnly)
	case FLTR_ROLEONLY:
		columns = addAliasToCols(alias, compColsRoleOnly, compColsRoleOnly)
	case FLTR_NIDONLY:
		columns = addAliasToCols(alias, compColsNIDOnly, compColsNIDOnly)
	case FLTR_ID_ONLY:
		columns = addAliasToCols(alias, compColsIdOnly, compColsIdOnly)
	case FLTR_ALL_W_GROUP:
		// Two sets of columns, Components and ComponentGroups
		columns = addAliasToCols(alias, compColsAllWithGroup1, compColsAllWithGroup1)
		gcs := addAliasToCols(aliasCG, compColsAllWithGroup2, compColsAllWithGroup2)
		columns = append(columns, gcs...)
	case FLTR_ID_W_GROUP:
		// two sets of columns, Components and ComponentGroups
		columns = addAliasToCols(alias, compColsIdWithGroup1, compColsIdWithGroup1)
		gcs := addAliasToCols(aliasCG, compColsIdWithGroup2, compColsIdWithGroup2)
		columns = append(columns, gcs...)
	case FLTR_DEFAULT:
		fallthrough
	default:
		columns = addAliasToCols(alias, compColsDefault, compColsDefault)
	}
	if from != "" {
		query = sq.Select(columns...).From(from + " " + alias)
	} else {
		query = sq.Select(columns...)
	}
	return query
}

// Change colNames to <alias>.<colname>.  If len(asNames) > 0, also add
// " AS asName[i]" (should have same length as colNames, if we are to assign
// them all
func addAliasToCols(alias string, colNames, asNames []string) []string {
	asLen := len(asNames)
	columns := []string{}

	for i, col := range colNames {
		colNew := alias + "." + col
		if asLen > 0 {
			if i < asLen {
				colNew += " AS " + asNames[i]
			}
		}
		columns = append(columns, colNew)
	}
	return columns
}

// Select statement from Component with optional join with
// component_group/component_group_members as per fieldFilter and f.
func selectComponents(f *ComponentFilter, fltr FieldFilter) (
	sq.SelectBuilder, error,
) {
	// Any query that will do a group by needs to be searched again for the
	// full set of rows - i.e. for membership data unless we don't filter
	// on partition or part
	if (fltr == FLTR_ID_W_GROUP || fltr == FLTR_ALL_W_GROUP) &&
		f != nil && (len(f.Partition) > 0 || len(f.Group) > 0) {

		// First select table to get list of matching ids
		q, err := makeComponentQuery(compTableJoinAlias, f, FLTR_ID_ONLY)
		if err != nil {
			return q, err
		}
		// Create a subquery
		selectCol := compTableSubAlias + "." + compIdCol
		query := selectComponentCols(fltr, compTableSubAlias, compTable).
			Where(q.Prefix(selectCol + " IN (").Suffix(")"))

		// Do another join so we get one row per membership entry with the
		// selected ids, and return the results
		query, err = joinComponentsWithGroups(query,
			compTableSubAlias, nil, nil, false) // false = don't group by id
		return query, err
	}
	return makeComponentQuery(compTableJoinAlias, f, fltr)
}

// Puts together custom query of HMS Components collection to filter based
// on query strings provided by user and a list of component xnames.
// NOTE: baseQuery ends up as a subquery as a result of this function.
func selectComponentsHierarchy(
	f *ComponentFilter,
	ff FieldFilter,
	ids []string,
) (sq.SelectBuilder, error) {
	args := make([]interface{}, 0, 1)

	// First select table to get list of matching ids
	q, err := makeComponentQuery(compTableJoinAlias, f, ff)
	if err != nil {
		return q, err
	}
	// Start second query as base for hierarchical query
	query := selectComponentCols(ff, compTableSubAlias, "")

	// As the first query as the From argument.
	query = query.FromSelect(q, compTableSubAlias)

	// Do the Where arguments on the main query to filter the subquery.
	// We add the regexs as a where clause after building them up.
	filterQuery := ""
	idCol := compTableSubAlias + "." + compIdCol
	if len(ids) > 0 {
		if len(f.Type) > 0 {
			for i := 0; i < len(ids); i++ {
				if i > 0 {
					// Multiples of the same key are OR'd together.
					filterQuery += " OR (" + idCol + " SIMILAR TO ?)"
				} else {
					filterQuery += "(" + idCol + " SIMILAR TO ?)"
				}
				// Build a regex for the id.
				// TODO: Explore the performance implications of making all
				//       the ids one big regex instead of a regex per id.
				arg := xnametypes.NormalizeHMSCompID(ids[i]) + "([[:alpha:]][[:alnum:]]*)?"
				args = append(args, arg)
			}
			query = query.Where(sq.Expr(filterQuery, args...))
		} else {
			// If no 'type' the string format expected by the database is
			// specified literally, (e.g. no component expansion).
			// No need to use REGEXP.
			for i := 0; i < len(ids); i++ {
				args = append(args, xnametypes.NormalizeHMSCompID(ids[i]))
			}
			query = query.Where(sq.Eq{idCol: args})
		}
	}
	return query, nil
}

// Create a select builder based on fieldfilter and the component filter.
// If needed, join with group/partition table for additional fields/filtering.
// returns squirrel.SelectBuilder for building onto or sending to database.
func makeComponentQuery(alias string, f *ComponentFilter, fltr FieldFilter) (
	sq.SelectBuilder, error,
) {
	// Get the base query:
	query := selectComponentCols(fltr, alias, compTable)
	if f != nil {
		// Check and normalize filter inputs, skipping if this has
		// already been done.
		if err := f.VerifyNormalize(); err != nil {
			return query, err
		}
	}
	// Add the base query opts - Note the order doesn't have to match the
	// sql statement.
	query = whereComponentCols(query, alias, f)

	// Determine if we really need a join
	needJoin := false
	groupAfterJoin := true
	if f != nil && (len(f.Partition) > 0 || len(f.Group) > 0) {
		needJoin = true
	}
	if fltr == FLTR_ID_W_GROUP || fltr == FLTR_ALL_W_GROUP {
		needJoin = true
		groupAfterJoin = false // No group on id since we have multiple rows
	}
	// Yes - need a join
	if needJoin {
		var err error
		query, err = joinComponentsWithGroups(query, alias, f.Group,
			f.Partition, groupAfterJoin)
		if err != nil {
			return query, err
		}
	}
	// Do other options...
	if f != nil && f.writeLock == true {
		query = query.Suffix("FOR UPDATE")
	}
	return query, nil
}

// Given an existing sq.Select build, fills in query parameters based
// on ComponentFilter, except for Group and Partition which have
// special handling that means they can't just be plopped in.
func whereComponentCols(q sq.SelectBuilder, alias string, f *ComponentFilter) sq.SelectBuilder {
	if f == nil {
		return q
	}
	q = whereComponentCol(q, alias+"."+compIdCol, f.ID)
	q = whereComponentCol(q, alias+"."+compTypeCol, f.Type)
	q = whereComponentCol(q, alias+"."+compStateCol, f.State)
	q = whereComponentCol(q, alias+"."+compFlagCol, f.Flag)
	q = whereComponentCol(q, alias+"."+compEnabledCol, f.Enabled)
	q = whereComponentCol(q, alias+"."+compSwStatusCol, f.SwStatus)
	q = whereComponentCol(q, alias+"."+compRoleCol, f.Role)
	q = whereComponentCol(q, alias+"."+compSubRoleCol, f.SubRole)
	q = whereComponentCol(q, alias+"."+compSubTypeCol, f.Subtype)
	q = whereComponentCol(q, alias+"."+compArchCol, f.Arch)
	q = whereComponentCol(q, alias+"."+compClassCol, f.Class)
	q = whereComponentCol(q, alias+"."+compResDisabledCol, f.ReservationDisabled)
	q = whereComponentCol(q, alias+"."+compLockedCol, f.Locked)

	// Special handling for NIDStart, NIDEnd and NID because of the
	// interaction between them
	q = whereComponentNIDCol(q, alias, f)

	return q
}

// Does an individual set of filter parameters in the where clause of an
// existing query.   Allows negated options.
func whereComponentCol(q sq.SelectBuilder, col string, args []string) sq.SelectBuilder {
	if args == nil {
		return q
	}
	pos, neg := splitSliceWithNegations(args)
	if pos != nil && len(pos) > 0 {
		q = q.Where(sq.Eq{col: pos})
	}
	if neg != nil && len(neg) > 0 {
		q = q.Where(sq.NotEq{col: neg})
	}
	return q
}

// Split a single array with negated string arguments (if any), i.e. !ready
// into separate pos and neg arguments with the ! negation prefix removed for
// neg.
func splitSliceWithNegations(args []string) (pos []string, neg []string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "!") {
			neg = append(neg, strings.TrimLeft(arg, "!"))
		} else {
			pos = append(pos, arg)
		}
	}
	return
}

// Special worker function to do join between Components table and the
// Group/Partition membership information.  The raw result is either the
// component with NULL group/part info, OR one entry for each part and group
// membership the component has.
func joinComponentsWithGroups(
	q sq.SelectBuilder,
	alias string,
	group_args []string,
	part_args []string,
	groupAfterJoin bool,
) (sq.SelectBuilder, error) {
	//
	// Set constants, using alias to modify the default table alias by adding
	//
	// Column names, e.g. alias = "c": "g.namespace" => "cg.namespace"
	grpMbNsColAlias := alias + compGroupMembersNsColAlias
	grpMbCmpIdColAlias := alias + compGroupMembersCmpIdColAlias
	grpMbGrpIdColAlias := alias + compGroupMembersGrpIdColAlias
	grpIdColAlias := alias + compGroupIdColAlias

	// Tables  i.e. "component_groups" g => "component_groups cg"

	grpTable := compGroupsTable + " " + alias + compGroupsAlias
	grpMbTable := compGroupMembersTable + " " + alias + compGroupMembersAlias

	// Determine join options.
	intersection := false // group and partition intersection query
	groupNULL := false    // Get
	partNULL := false
	if len(part_args) == 1 && len(group_args) == 1 {
		// We can only have one group-partition intersection query.
		intersection = true
	} else if len(part_args) >= 1 && len(group_args) >= 1 {
		return q, ErrHMSDSMultipleGroupAndPart
	}
	// Look at possible NULL group or partition arguments before
	// adding these options.  We treat these specially, but don't allow
	// mixing NULL and non-NULL args.  (This is a lot more efficient.)
	joinMod := ""
	if len(part_args) == 1 && part_args[0] == "NULL" {
		if len(group_args) == 0 || group_args[0] == "NULL" {
			partNULL = true
			// Only join non-partition, so components without partitions
			// will have NULL namespace
			joinMod = " AND " + grpMbNsColAlias +
				" = '" + partGroupNamespace + "'"
		} else {
			return q, ErrHMSDSNullPartBadGroup
		}
	}
	if len(group_args) == 1 && group_args[0] == "NULL" {
		if len(part_args) == 0 || partNULL == true {
			groupNULL = true
			if partNULL == false {
				// Only join members with partitions, so components
				// without groups will have NULL namespace.
				joinMod = " AND " + grpMbNsColAlias +
					" != '" + partGroupNamespace + "'"
			} else {
				// Both NULL, they cancel each other out and we don't
				// need to filter anything - should have no membership info at
				// all.
				joinMod = ""
				intersection = false
			}
		} else if partNULL != true {
			return q, ErrHMSDSNullGroupBadPart
		}
	}
	// Do join of components to their group members with group info
	q = q.
		LeftJoin(grpMbTable + " ON " +
			grpMbCmpIdColAlias + " = " + alias + "." + compIdCol + joinMod).
		LeftJoin(grpTable + " ON " + grpIdColAlias + " = " + grpMbGrpIdColAlias)

	//
	// Set Where arguments for groups and partitions
	//
	if !partNULL && !groupNULL {
		//  Do OR of list of groups OR partitions or a single group AND part
		var err error
		q, err = whereComponentGrpCol(q, alias, group_args, part_args)
		if err != nil {
			return q, err
		}
	} else {
		// Do checks for NO groups and/or partition membership info, i.e.
		// not assigned to any group and/or partition.  Assumes setting
		// joinMod as above during join.
		q = q.Where(grpMbNsColAlias + " IS NULL")
	}
	// We aren't returning group-related info and expect only a single line
	// per component.
	if groupAfterJoin {
		q = q.GroupBy(alias + "." + compIdCol)
	}
	// Since we allow intersection with exactly one partition and group, we
	// can filter those entries by looking for components with two rows.
	if intersection == true {
		q = q.Having("COUNT(*) = 2")
	}
	return q, nil
}

// Does group and partition related operations.  Does not allow negation.
func whereComponentGrpCol(
	q sq.SelectBuilder,
	alias string,
	group_args []string,
	part_args []string,
) (sq.SelectBuilder, error) {
	// Check for stray "NULL" args that aren't where they're supposed to be
	for _, arg := range group_args {
		if len(group_args) > 1 && arg == "NULL" {
			return q, ErrHMSDSNullBadMixGroup
		}
	}
	for _, arg := range part_args {
		if len(part_args) > 1 && arg == "NULL" {
			return q, ErrHMSDSNullBadMixPart
		}
	}
	// Do a compound IN (grp1|grp1|etc.) AND namespace = group
	// AND/OR a compound IN (part1|part1|etc.) AND namespace = partition
	// Note we add alias to the existing default alias, i.e. g->cg
	if group_args != nil && part_args != nil {
		q = q.Where(sq.Or{
			sq.And{sq.Eq{alias + compGroupNameColAlias: group_args},
				sq.Eq{alias + compGroupNamespaceColAlias: groupNamespace}},
			sq.And{sq.Eq{alias + compGroupNameColAlias: part_args},
				sq.Eq{alias + compGroupNamespaceColAlias: partNamespace}}})
	} else if group_args != nil {
		q = q.Where(sq.And{sq.Eq{alias + compGroupNameColAlias: group_args},
			sq.Eq{alias + compGroupNamespaceColAlias: groupNamespace}})
	} else if part_args != nil {
		q = q.Where(sq.And{sq.Eq{alias + compGroupNameColAlias: part_args},
			sq.Eq{alias + compGroupNamespaceColAlias: partNamespace}})
	}
	return q, nil
}

// Special handling for NIDStart, NIDEnd and NID because of the
// interaction between them.  Adds to where clause of an existing query.
func whereComponentNIDCol(q sq.SelectBuilder, alias string, f *ComponentFilter) sq.SelectBuilder {
	if f == nil {
		return q
	}
	// NIDs - This is ugly because you can't string together GtOrEq in squirrel
	nid_query := ""
	nidColAlias := alias + "." + compNIDCol
	nid_args := make([]interface{}, 0, 1)
	if len(f.NIDStart) > 0 || len(f.NIDEnd) > 0 {
		for i := 0; i < len(f.NIDStart); i++ {
			if i > 0 {
				nid_query += " OR (" + nidColAlias + " >= ?"
			} else {
				nid_query += "((" + nidColAlias + " >= ?"
			}
			nid_args = append(nid_args, f.NIDStart[i])
			// Check for a paired NIDEnd
			if len(f.NIDEnd) >= i+1 {
				nid_query += " AND " + nidColAlias + " <= ?)"
				nid_args = append(nid_args, f.NIDEnd[i])
			} else {
				nid_query += ")"
			}
		}
		// OR any leftover NIDEnd values if we had more NIDEnds than NIDStarts
		for i := len(f.NIDStart); i < len(f.NIDEnd); i++ {
			if i > 0 {
				nid_query += " OR " + nidColAlias + " <= ?"
			} else {
				nid_query += " (" + nidColAlias + " <= ?"
			}
			nid_args = append(nid_args, f.NIDEnd[i])
		}
		// OR any NID values if we have them
		if len(f.NID) > 0 {
			for i := 0; i < len(f.NID); i++ {
				nid_query += " OR " + nidColAlias + " = ?"
				nid_args = append(nid_args, f.NID[i])
			}
		}
		// If no start, presume a start of positive nids.
		if len(f.NIDStart) == 0 {
			nid_query += ") AND " + nidColAlias + " >= 0"
		} else {
			nid_query += ")"
		}
		if nid_query != "" {
			q = q.Where(sq.Expr(nid_query, nid_args...))
		}
	} else if len(f.NID) > 0 {
		// Do NIDs as a regular array
		q = whereComponentCol(q, nidColAlias, f.NID)
	}
	return q
}

// Get a query for some or all Hardware Inventory entries with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct.
func getHWInvByLocQuery(f_opts ...HWInvLocFiltFunc) (sq.SelectBuilder, error) {
	var (
		outerQuery bool
		hierarchy  bool
		queryTable string
	)
	parentIDs := make(map[string]bool)

	// Parse the filter options
	f := new(HWInvLocFilter)
	for _, opts := range f_opts {
		opts(f)
	}

	if len(f.Partition) > 0 {
		queryTable = hwInvPartTable
	} else {
		queryTable = hwInvTable
	}

	query := sq.Select(addAliasToCols(hwInvAlias, hwInvCols, hwInvCols)...).
		From(queryTable + " " + hwInvAlias)
	queryOuter := query
	idCol := hwInvAlias + "." + hwInvIdCol
	typeCol := hwInvAlias + "." + hwInvTypeCol
	// Set up a subquery if we have both children and types.
	// This will result in us getting the children of the child
	// components of the ID of the specified type. For example if
	// type=node and id=x0c0s0, we would get all the nodes under
	// x0c0s0 as well as their processors and memory.
	// If parents was also specified, we'll ignore types because
	// we'll just end up with a gap between children and parents.
	if f.Children && len(f.Type) > 0 && !f.Parents {
		outerQuery = true
		hierarchy = true
		idCol = hwInvLocAlias + "." + hwInvLocIdCol
		typeCol = hwInvLocAlias + "." + hwInvLocTypeCol
		// Use the hwinv by location only table in the subquery
		// because we only need IDs and this avoids doing a join.
		query = sq.Select(hwInvLocAlias + "." + hwInvLocIdCol).
			From(hwInvLocTable + " " + hwInvLocAlias)
	} else if f.Children || (len(f.Type) > 0 && !f.Parents) {
		// Use "id SIMILAR TO..." instead of "id IN..." to include
		// child components
		hierarchy = true
	}

	args := make([]interface{}, 0, 1)
	filterQuery := ""
	if len(f.ID) > 0 {
		if hierarchy {
			for i, id := range f.ID {
				// Form a list of parent IDs to look for.
				if f.Parents {
					childId := id
					for {
						childId = xnametypes.GetHMSCompParent(childId)
						// Do not include the system as a parent.
						// The hms-xname package gives back proper parents when GetHMSCompParent is called on a Cabinet or CDU.
						// The older hms-base implementation of GetHMSCompParent would return empty string for Cabinet or CDU types.
						if childId != "s0" && len(childId) > 0 {
							// Put it in a map to ensure uniqueness
							parentIDs[childId] = true
						} else {
							break
						}
					}
				}
				// Squirrel doesn't have a way of adding "SIMILAR TO"
				// we have to form it manually.
				if i > 0 {
					// Multiples of the same key are OR'd together.
					filterQuery += " OR (" + idCol + " SIMILAR TO ?)"
				} else {
					filterQuery += "(" + idCol + " SIMILAR TO ?)"
				}
				// Build a regex for the id.
				arg := xnametypes.NormalizeHMSCompID(id) + "([[:alpha:]][[:alnum:]]*)?"
				args = append(args, arg)
			}
		} else {
			// If no children, the string format expected by the database
			// is specified literally, (e.g. no component expansion). No
			// need to use REGEXP.
			for _, id := range f.ID {
				// Form a list of parent IDs to look for.
				if f.Parents {
					childId := id
					for {
						childId = xnametypes.GetHMSCompParent(childId)
						if len(childId) > 0 {
							// Put it in a map to ensure uniqueness
							parentIDs[childId] = true
						} else {
							break
						}
					}
					// If we got here then we either have 'type' with parents
					// and no children or just parents with no children. Add
					// the parent IDs to the array of IDs here so we can get
					// parents of 'type'.
					for parentID, _ := range parentIDs {
						args = append(args, parentID)
					}
					parentIDs = make(map[string]bool)
				}
				args = append(args, xnametypes.NormalizeHMSCompID(id))
			}
			filterQuery, args, _ = sq.Eq{idCol: args}.ToSql()
		}
	}
	// Form the type query filter only if we're not looking
	// for both parents and children.
	if len(f.Type) > 0 && !(f.Parents && f.Children) {
		targs := []string{}
		for _, t := range f.Type {
			normType := xnametypes.VerifyNormalizeType(t)
			if normType == "" {
				return query, ErrHMSDSArgBadType
			}
			targs = append(targs, normType)
		}
		// Use squirrel to form the sql then append it to our running custom query
		typeQuery, typeArgs, _ := sq.Eq{typeCol: targs}.ToSql()
		if len(filterQuery) == 0 {
			filterQuery = typeQuery
		} else {
			filterQuery += " AND " + typeQuery
		}
		args = append(args, typeArgs...)
	}
	// Here we attach the subquery to the begining of the main query.
	// The subquery gives us a list of IDs in the form of a regex string
	// to be used in the main query with "id SIMILAR TO...". This way we
	// can query for a set of IDs then get the children of those IDs. Only
	// the the ID and type queries go in the subquery. The parent id query
	// always goes in the main query.
	if outerQuery {
		query = query.Where(sq.Expr(filterQuery, args...))
		iQuery, iArgs, _ := query.ToSql()
		prefixQuery := "WITH id_sel AS (SELECT ('('||array_to_string(array(" + iQuery + "),'|')||?) AS id_str)"
		filterQuery = hwInvAlias + "." + hwInvIdCol + " SIMILAR TO (SELECT id_str FROM id_sel)"
		iArgs = append(iArgs, ")([[:alpha:]][[:alnum:]]*)?")
		query = queryOuter
		query = query.Prefix(prefixQuery, iArgs...)
		// Clear the running args so we don't add the privious args twice.
		args = make([]interface{}, 0, 1)
	}
	// Add the parent ID query last. This will get OR'd as we want
	// these IDs regardless of the other filters (expect the partition filter).
	if f.Parents && len(parentIDs) > 0 {
		pArray := []string{}
		for parentID, _ := range parentIDs {
			pArray = append(pArray, parentID)
		}
		pQuery, pArgs, _ := sq.Eq{hwInvAlias + "." + hwInvIdCol: pArray}.ToSql()
		if len(filterQuery) == 0 {
			filterQuery = pQuery
		} else {
			filterQuery = "(" + filterQuery + ") OR " + pQuery
		}
		args = append(args, pArgs...)
	}
	if len(filterQuery) > 0 {
		query = query.Where(sq.Expr(filterQuery, args...))
	}
	// Memberships to partitions only include entries in the components table
	// but we'll want to still include child components when filtering hwinv by
	// partition. To do this, we'll use the hwinv_by_loc_with_partition view that
	// uses the 'parent_node' column to associate lower components with parent
	// partitions.
	// NOTE: parent components will only be included if they
	// or one of their parent components are in the partition.
	if len(f.Partition) > 0 {
		partCol := hwInvAlias + "." + hwInvPartPartitionCol
		query = query.Where(sq.Eq{partCol: f.Partition})
	}
	return query, nil
}
