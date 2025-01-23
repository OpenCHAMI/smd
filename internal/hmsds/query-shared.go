// MIT License
//
// (C) Copyright [2019-2023] Hewlett Packard Enterprise Development LP
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

import (
	"fmt"
	"strings"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
)

// Matches generic query sources in external API, but with real table names,
// not just identifiers for the type of data.
const (
	componentsTableDB            = "components"
	redfishEndpointsTableDB      = "rf_endpoints"
	componentEndpointsTableDB    = "comp_endpoints"
	componentEndpointInfoTableDB = "comp_endpoints_info"
	nodeMapTableDB               = "node_nid_mapping"
	powerMapTableDB              = "power_mapping"
	hwInvByLocTableDB            = "hwinv_by_loc"
	hwInvByFRUTableDB            = "hwinv_by_fru"
	discoveryStatusTableDB       = "discovery_status"
	scnSubcriptionsTableDB       = "scn_subscriptions"
)

// Generic id field queries
const getIDPrefix = "SELECT id FROM "

const getCompIDPrefix = getIDPrefix + componentsTableDB + " "
const getRFEndpointIDPrefix = getIDPrefix + redfishEndpointsTableDB + " "
const getCompEndpointIDPrefix = getIDPrefix + componentEndpointsTableDB + " "
const getNodeMapIDPrefix = getIDPrefix + nodeMapTableDB + " "
const getPowerMapIDPrefix = getIDPrefix + powerMapTableDB + " "
const getHWInvByLocIDPrefix = getIDPrefix + hwInvByLocTableDB + " "
const getHWInvByFRUIDPrefix = getIDPrefix + hwInvByFRUTableDB + " "
const getDiscoveryStatusIDPrefix = getIDPrefix + discoveryStatusTableDB + " "

//
// Suffix strings for various select queries
//

const suffixCompFilter = `
) AS filter;`

const suffixByID = `
WHERE id = ?;`

const suffixByIDForUpd = `
WHERE id = ? FOR UPDATE;`

const suffixByFRUID = `
WHERE fru_id = ?;`

const suffixByFRUIDForUpd = `
WHERE fru_id = ? FOR UPDATE;`

const suffixByType = `
WHERE type = ?;`

const suffixByTypeForUpd = `
WHERE type = ? FOR UPDATE;`

const suffixByNID = `
WHERE nid = ?;`

const suffixByNIDForUpd = `
WHERE nid = ? FOR UPDATE;`

const suffixByFQDN = `
WHERE fqdn = ?;`

const suffixByFQDNForUpd = `
WHERE fqdn = ? FOR UPDATE;`

const suffixByRfEP = `
WHERE rf_endpoint_id = ?;`

const suffixByRfEPForUpd = `
WHERE rf_endpoint_id = ? FOR UPDATE;`

const suffixByRfType = `
WHERE redfish_type = ?;`

const suffixByRfTypeForUpd = `
WHERE redfish_type = ? FOR UPDATE;`

const suffixByRfEPandRfType = `
WHERE rf_endpoint_id = ? AND redfish_type = ?;`

const suffixByRfEPandRfTypeForUpd = `
WHERE rf_endpoint_id = ? AND redfish_type = ? FOR UPDATE;`

const suffixByRfSubtype = `
WHERE redfish_subtype = ?;`

const suffixByRfSubtypeForUpd = `
WHERE redfish_subtype = ? FOR UPDATE;`

//
// Component structs
//

const getCompPrefix = `
SELECT
    id,
    type,
    state,
    flag,
    enabled,
    admin,
    role,
    subrole,
    nid,
    subtype,
    nettype,
    arch,
    class,
    reservation_disabled,
    locked
FROM components `

const getCompStatePrefix = `
SELECT
    filter.id,
    filter.type,
    filter.state,
    filter.flag
FROM ( `

const getCompFlagPrefix = `
SELECT
    filter.id,
    filter.type,
    filter.flag
FROM ( `

const getCompRolePrefix = `
SELECT
    filter.id,
    filter.type,
    filter.role,
    filter.subrole
FROM ( `

const getCompNIDPrefix = `
SELECT
    filter.id,
    filter.type,
    filter.nid
FROM ( `

const getCompQueryPrefix = `
SELECT
    comp.id,
    comp.type,
    comp.state,
    comp.flag,
    comp.enabled,
    comp.admin,
    comp.role,
    comp.subrole,
    comp.nid,
    comp.subtype,
    comp.nettype,
    comp.arch,
    comp.class,
    comp.reservation_disabled,
    comp.locked
FROM ( `

const getComponentByIDQuery = getCompPrefix + suffixByID
const getComponentByIDForUpdQuery = getCompPrefix + suffixByIDForUpd
const getComponentByTypeQuery = getCompPrefix + suffixByType
const getComponentByTypeForUpdQuery = getCompPrefix + suffixByTypeForUpd
const getComponentByNIDQuery = getCompPrefix + suffixByNID
const getComponentByNIDForUpdQuery = getCompPrefix + suffixByNIDForUpd

const getComponentsAllQuery = getCompPrefix + ";"

//const getComponentByFQDNQuery = getCompPrefix + suffixByFQDN

//
// Component queries
//

const updateCompPrefix = `UPDATE components SET `

const updateCompStatePrefix = `
UPDATE components SET
    state = ?,
    flag  = ? `

const updateCompFlagOnlyPrefix = `
UPDATE components SET
    flag  = ? `

const updateCompEnabledPrefix = `
UPDATE components SET
    enabled = ? `

const updateCompSwStatusPrefix = `
UPDATE components SET
    admin = ? `

const updateCompRolePrefix = `
UPDATE components SET
    role = ?,
    subrole = ? `

const updateCompClassPrefix = `
UPDATE components SET
    class = ? `

const updateCompNIDPrefix = `
UPDATE components SET
    nid = ? `

const updateCompStateByIDQuery = updateCompStatePrefix + suffixByID
const updateCompFlagOnlyByIDQuery = updateCompFlagOnlyPrefix + suffixByID
const updateCompEnabledByIDQuery = updateCompEnabledPrefix + suffixByID
const updateCompSwStatusByIDQuery = updateCompSwStatusPrefix + suffixByID
const updateCompRoleByIDQuery = updateCompRolePrefix + suffixByID
const updateCompNIDByIDQuery = updateCompNIDPrefix + suffixByID

const deleteComponentPrefix = `
DELETE FROM components `

const deleteComponentByIDQuery = deleteComponentPrefix + suffixByID
const deleteComponentsAllQuery = deleteComponentPrefix + ";"

// getCompIDPrefix
// Node xname->NID mapping
const getNodeMapPrefix = `
SELECT
    id,
    nid,
    role,
    subrole,
    node_info
FROM node_nid_mapping `

const getNodeMapByIDQuery = getNodeMapPrefix + suffixByID
const getNodeMapsAllQuery = getNodeMapPrefix + ";"

//
// Component NID map queries
//

const updateNodeMapPrefix = `
UPDATE node_nid_mapping SET
    nid = ?,
    role = ?,
    subrole = ?,
    node_info = ?`

const deleteNodeMapPrefix = `
DELETE FROM node_nid_mapping `

const deleteNodeMapByIDQuery = deleteNodeMapPrefix + suffixByID
const deleteNodeMapsAllQuery = deleteNodeMapPrefix + ";"

// Power mapping
const getPowerMapPrefix = `
SELECT
    id,
    powered_by
FROM power_mapping `

const getPowerMapByIDQuery = getPowerMapPrefix + suffixByID
const getPowerMapsAllQuery = getPowerMapPrefix + ";"

//
// Component Power map queries
//

const updatePowerMapPrefix = `
UPDATE power_mapping SET
    powered_by = ?`

const deletePowerMapPrefix = `
DELETE FROM power_mapping `

const deletePowerMapByIDQuery = deletePowerMapPrefix + suffixByID
const deletePowerMapsAllQuery = deletePowerMapPrefix + ";"

//
// Hardware Inventory Queries
//

const getHWInvByLocPrefix = `
SELECT
    id,
    type,
    ordinal,
    status,
    location_info,
    fru_id
FROM hwinv_by_loc `

const getHWInvByFRUPrefix = `
SELECT
    fru_id,
    type,
    subtype,
    fru_info
FROM hwinv_by_fru `

const getHWInvByLocWithFRUPrefix = `
SELECT
    id,
    type,
    ordinal,
    status,
    location_info,
    fru_id,
    fru_type,
    fru_subtype,
    fru_info
FROM hwinv_by_loc_with_fru `

const getHWInvByLocByIDQuery = getHWInvByLocPrefix + suffixByID
const getHWInvByFRUByFRUIDQuery = getHWInvByFRUPrefix + suffixByFRUID
const getHWInvByFRUAllQuery = getHWInvByFRUPrefix + ";"
const getHWInvByLocWithFRUByIDQuery = getHWInvByLocWithFRUPrefix + suffixByID
const getHWInvByLocWithFRUAllQuery = getHWInvByLocWithFRUPrefix + ";"

//
// Hardware Inventory - Delete operations
//

const deleteHWInvByLocPrefix = `
DELETE FROM hwinv_by_loc `

const deleteHWInvByLocIDQuery = deleteHWInvByLocPrefix + suffixByID
const deleteHWInvByLocsAllQuery = deleteHWInvByLocPrefix + ";"

const deleteHWInvByFRUPrefix = `
DELETE FROM hwinv_by_fru `

const deleteHWInvByFRUIDQuery = deleteHWInvByFRUPrefix + suffixByFRUID
const deleteHWInvByFRUsAllQuery = deleteHWInvByFRUPrefix + ";"

//
// Redfish Endpoints - Queries
//

const getRFEndpointPrefix = `
SELECT
    rf.id,
    rf.type,
    rf.name,
    rf.hostname,
    rf.domain,
    rf.fqdn,
    rf.enabled,
    rf.uuid,
    rf.user,
    rf.password,
    rf.useSSDP,
    rf.macRequired,
    rf.macAddr,
    rf.ipAddr,
    rf.rediscoverOnUpdate,
    rf.templateID,
    rf.discovery_info
FROM rf_endpoints rf`

// Primary key is ID, other modifiers produce multiple results
const getRFEndpointByIDQuery = getRFEndpointPrefix + suffixByID
const getRFEndpointsByTypeQuery = getRFEndpointPrefix + suffixByType
const getRFEndpointsAllQuery = getRFEndpointPrefix + ";"

//
// RedfishEndpoints - Delete operations
//

const deleteRFEndpointPrefix = `
DELETE FROM rf_endpoints `

const deleteRFEndpointByIDQuery = deleteRFEndpointPrefix + suffixByID
const deleteRFEndpointsAllQuery = deleteRFEndpointPrefix + ";"

//
// Component and Service Endpoints - Queries
//

const getCompEndpointPrefix = `
SELECT
    id,
    type,
    domain,
    redfish_type,
    redfish_subtype,
    mac,
    uuid,
    odata_id,
    rf_endpoint_id,
    rf_endpoint_fqdn,
    component_info,
    enabled
FROM comp_endpoints_info `

const getServiceEndpointPrefix = `
SELECT
    rf_endpoint_id,
    redfish_type,
    redfish_subtype,
    uuid,
    odata_id,
    rf_endpoint_fqdn,
    service_info
FROM service_endpoints_info `

// Primary key is ID, others modifiers produce multiple results
const getCompEndpointByIDQuery = getCompEndpointPrefix + suffixByID
const getCompEndpointsByRfEPQuery = getCompEndpointPrefix + suffixByRfEP
const getCompEndpointsByTypeQuery = getCompEndpointPrefix + suffixByType
const getCompEndpointsByRfSubtypeQuery = getCompEndpointPrefix + suffixByRfSubtype
const getCompEndpointsAllQuery = getCompEndpointPrefix + ";"

// Primary key is Redfish EP id + Redfish type, others produce multiple results
const getServiceEndpointsByRfEPQuery = getServiceEndpointPrefix + suffixByRfEP
const getServiceEndpointsByRfTypeQuery = getServiceEndpointPrefix + suffixByRfType
const getServiceEndpointsByRfSubtypeQuery = getServiceEndpointPrefix + suffixByRfSubtype
const getServiceEndpointByRfEPandRfTypeQuery = getServiceEndpointPrefix + suffixByRfEPandRfType
const getServiceEndpointsAllQuery = getServiceEndpointPrefix + ";"

//
// ComponentEndpoints - Delete operations
//

const deleteCompEndpointPrefix = `
DELETE FROM comp_endpoints `

const deleteCompEndpointByIDQuery = deleteCompEndpointPrefix + suffixByID
const deleteCompEndpointsAllQuery = deleteCompEndpointPrefix + ";"

//
// ServiceEndpoints - Delete operations
//

const deleteServiceEndpointPrefix = `
DELETE FROM service_endpoints `

const deleteServiceByRfEPandRfTypeQuery = deleteServiceEndpointPrefix + suffixByRfEPandRfType
const deleteServiceEndpointsAllQuery = deleteServiceEndpointPrefix + ";"

//
// Discovery status
//

const getDiscoveryStatusPrefix = `
SELECT
    id,
    status,
    last_update,
    details
FROM discovery_status `

const getDiscoveryStatusByIDQuery = getDiscoveryStatusPrefix + suffixByID
const getDiscoveryStatusesAllQuery = getDiscoveryStatusPrefix + ";"

//
// SCNs
//

const getSCNSubsAll = `
SELECT
    id,
    subscription
FROM scn_subscriptions;`

const getSCNSub = `
SELECT
    id,
    subscription
FROM scn_subscriptions WHERE id = ?;`

const getSCNSubUpdate = `
SELECT
    id,
    subscription
FROM scn_subscriptions WHERE id = ? FOR UPDATE;`

const insertSCNSub = `
INSERT INTO scn_subscriptions (
    sub_url,
    subscription)
VALUES (?, ?);`

const updateSCNSub = `
UPDATE scn_subscriptions SET
    sub_url = ?,
    subscription = ?
WHERE id = ?;`

const deleteSCNSubscription = `
DELETE FROM scn_subscriptions WHERE id = ?;`

const deleteSCNSubscriptionsAll = `
DELETE FROM scn_subscriptions;`

////////////////////////////////////////////////////////////////////////////
//
// Helper functions - Query building
//
////////////////////////////////////////////////////////////////////////////

// Puts together custom query of ServiceEndpoints collection to filter based
// on query strings provided by user.
func buildServiceEPQuery(baseQuery string, f *ServiceEPFilter) (string, []interface{}, error) {
	q := newPreparedQuery(baseQuery)

	err := q.doQueryArg("rf_endpoint_id", f.RfEndpointID, validXNameFilter)
	if err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadID
	}
	err = q.doQueryArg("redfish_type", f.Service, strToAlphaNum)
	if err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadType
	}
	// Terminate statement
	if f.writeLock == true {
		q.appendToQuery(" FOR UPDATE;")
	} else {
		q.appendToQuery(";")
	}
	return q.query, q.args, nil
}

// Puts together custom query of ComponentEndpoints collection to filter based
// on query strings provided by user.
func buildCompEPQuery(baseQuery string, f *CompEPFilter) (string, []interface{}, error) {
	q := newPreparedQuery(baseQuery)

	err := q.doQueryArg("id", f.ID, validXNameFilter)
	if err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadID
	}
	err = q.doQueryArg("rf_endpoint_id", f.RfEndpointID, validXNameFilter)
	if err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadID
	}
	err = q.doQueryArg("redfish_subtype", f.RedfishType, strToAlphaNum)
	if err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadRedfishType
	}
	err = q.doQueryArg("type", f.Type, base.VerifyNormalizeType)
	if err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadType
	}
	// Terminate statement
	if f.writeLock == true {
		q.appendToQuery(" FOR UPDATE;")
	} else {
		q.appendToQuery(";")
	}
	return q.query, q.args, nil
}

// Puts together custom query of ComponentEndpoints collection to filter based
// on query strings provided by user.
func buildRedfishEPQuery(baseQuery string, f *RedfishEPFilter) (string, []interface{}, error) {
	q := newPreparedQuery(baseQuery)

	if err := q.doQueryArg("id", f.ID, validXNameFilter); err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadID
	}
	if err := q.doQueryArg("fqdn", f.FQDN, nil); err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadArg
	}
	if err := q.doQueryArg("uuid", f.UUID, nil); err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadArg
	}
	err := q.doQueryArg("type", f.Type, base.VerifyNormalizeType)
	if err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadType
	}
	if err := q.doQueryArg("macAddr", f.MACAddr, nil); err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadArg
	}
	if err := q.doQueryArg("ipAddr", f.IPAddr, nil); err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadArg
	}
	if err := q.doQueryArg("discovery_info ->> 'LastDiscoveryStatus'", f.LastStatus, nil); err != nil {
		return baseQuery, q.args, ErrHMSDSArgBadArg
	}
	// Terminate statement
	if f.writeLock == true {
		q.appendToQuery(" FOR UPDATE;")
	} else {
		q.appendToQuery(";")
	}
	return q.query, q.args, nil
}

// Puts together custom query of HMS Components collection to filter based
// on query strings provided by user.   Intended for prepared statements,
// inserting "?" variables and returning the args as interface array.
func buildComponentQuery(
	baseQuery string,
	f *ComponentFilter,
) (string, []interface{}, error) {
	q := newPreparedQuery(baseQuery)
	err := q.setCompWhereQuery(f, false)
	return q.query, q.args, err
}

// Start a component update with the given filter options and the set of
// ids.  If direct is set, non-id values will not be parameterized.
func startCompUpdate(ids []string, direct bool, f_opts ...CompFiltFunc) (*preparedQuery, error) {
	q := newPreparedQuery(updateCompPrefix)
	q.setDirectQuery(direct)
	f := new(ComponentFilter)
	for _, f_opt := range f_opts {
		f_opt(f)
	}
	if err := q.setCompUpdateArgs(f); err != nil {
		return q, err
	}
	q.setDirectQuery(false)
	fcomp := ComponentFilter{ID: ids}
	if err := q.setCompWhereQuery(&fcomp, true); err != nil {
		return q, err
	}
	q.setDirectQuery(direct)
	return q, nil
}

// Using filter ops, append the existing write query with filtering options
// to restrict the candidate rows based on the options give.
// NOTE: If started with direct=true, do not feed raw user input into these
// as we do not use the args as prepared query arguments.
func finishCompUpdate(
	q *preparedQuery,
	f_opts ...CompFiltFunc,
) (string, []interface{}, error) {
	f := new(ComponentFilter)
	for _, f_opt := range f_opts {
		f_opt(f)
	}
	err := q.setCompWhereQuery(f, false)
	return q.query, q.args, err
}

// Start custom Update ... SET arg1 = val1, arg2 = val2, ... query.
// Does NOT add WHERE clause and is unterminated.
func (q *preparedQuery) setCompUpdateArgs(f *ComponentFilter) error {
	if q == nil {
		return ErrHMSDSArgNil
	}
	err := q.doUpdateArg("type", f.Type, nil, base.VerifyNormalizeType, false)
	if err != nil {
		return ErrHMSDSArgBadType
	}
	// Addtional states to be added as normal AND options
	err = q.doUpdateArg("state", f.State, nil, base.VerifyNormalizeState, false)
	if err != nil {
		return ErrHMSDSArgBadState
	}
	// Ditto flag
	err = q.doUpdateArg("flag", f.Flag, f.flagCondition, base.VerifyNormalizeFlag, false)
	if err != nil {
		return ErrHMSDSArgBadFlag
	}
	err = q.doUpdateArg("admin", f.SwStatus, nil, nil, true)
	if err != nil {
		return ErrHMSDSArgBadArg
	}
	err = q.doUpdateArg("enabled", f.Enabled, nil, strToDbBool, false)
	if err != nil {
		return ErrHMSDSArgNoMatch
	}
	err = q.doUpdateArg("role", f.Role, nil, base.VerifyNormalizeRole, false)
	if err != nil {
		return ErrHMSDSArgBadRole
	}
	err = q.doUpdateArg("subrole", f.SubRole, nil, base.VerifyNormalizeSubRole, false)
	if err != nil {
		return ErrHMSDSArgBadSubRole
	}
	err = q.doUpdateArg("subtype", f.Subtype, nil, strToAlphaNum, true)
	if err != nil {
		return ErrHMSDSArgBadSubtype
	}
	err = q.doUpdateArg("arch", f.Arch, nil, base.VerifyNormalizeArch, false)
	if err != nil {
		return ErrHMSDSArgBadArch
	}
	err = q.doUpdateArg("class", f.Class, nil, base.VerifyNormalizeClass, false)
	if err != nil {
		return ErrHMSDSArgBadClass
	}
	// No NIDStart or NIDEnds. Just NIDs.
	err = q.doUpdateArg("nid", f.NID, nil, nidStrTransform, false)
	if err != nil {
		return ErrHMSDSArgBadType
	}
	return nil
}

// Puts together WHERE query, starting one if cont=true or terminating it
// otherwise.  Used prepared args or direct substitution depending on q
// options.
func (q *preparedQuery) setCompWhereQuery(f *ComponentFilter, cont bool) error {
	// Build up a query string to filter the full results
	// NOTE: if nothing is given, we still should end up with a
	// valid query, ditto when more than one argument.
	if q == nil {
		return ErrHMSDSArgNil
	}
	//
	// Construct 'where' arguments.
	//
	err := q.doQueryArg("id", f.ID, validXNameFilter)
	if err != nil {
		return ErrHMSDSArgBadID
	}
	err = q.doQueryArg("type", f.Type, base.VerifyNormalizeType)
	if err != nil {
		return ErrHMSDSArgBadType
	}
	// Allow state OR flag as a subclause without ORing the whole query
	err = q.doQueryArgsWithOR(
		"state", "flag",
		f.orState, f.orFlag,
		base.VerifyNormalizeState,
		base.VerifyNormalizeFlag)
	if err != nil {
		return ErrHMSDSArgBadState
	}
	// Addtional states to be added as normal AND options
	err = q.doQueryArg("state", f.State, base.VerifyNormalizeState)
	if err != nil {
		return ErrHMSDSArgBadState
	}
	// Ditto flag
	err = q.doQueryArg("flag", f.Flag, base.VerifyNormalizeFlag)
	if err != nil {
		return ErrHMSDSArgBadFlag
	}
	err = q.doQueryArg("admin", f.SwStatus, nil)
	if err != nil {
		return ErrHMSDSArgBadArg
	}
	err = q.doQueryArg("enabled", f.Enabled, strToDbBool)
	if err != nil {
		return ErrHMSDSArgNoMatch
	}
	err = q.doQueryArg("role", f.Role, base.VerifyNormalizeRole)
	if err != nil {
		return ErrHMSDSArgBadRole
	}
	err = q.doQueryArg("subrole", f.SubRole, base.VerifyNormalizeSubRole)
	if err != nil {
		return ErrHMSDSArgBadSubRole
	}
	err = q.doQueryArg("subtype", f.Subtype, strToAlphaNum)
	if err != nil {
		return ErrHMSDSArgBadSubtype
	}
	err = q.doQueryArg("arch", f.Arch, base.VerifyNormalizeArch)
	if err != nil {
		return ErrHMSDSArgBadArch
	}
	err = q.doQueryArg("class", f.Class, base.VerifyNormalizeClass)
	if err != nil {
		return ErrHMSDSArgBadClass
	}
	// Specified NIDStart and NIDEnd keys will be AND paired.
	// EX: (nid >= NIDStart AND nid <= NIDEnd)
	// Multiple ranges need to be OR'd together.
	// EX: (nid >= NIDStart1 AND nid <= NIDEnd1) OR (nid >= NIDStart2 AND nid <= NIDEnd2)
	// Ranges need to be OR'd with NID keys.
	// EX: ((nid >= NIDStart1 AND nid <= NIDEnd1) OR nid = NID)
	if len(f.NIDStart) > 0 || len(f.NIDEnd) > 0 {
		for i := 0; i < len(f.NIDStart); i++ {
			if q.first == true {
				q.appendToQuery(" WHERE ((nid >= ?")
				q.first = false
			} else {
				if i > 0 {
					q.appendToQuery(" OR (nid >= ?")
				} else {
					q.appendToQuery(" AND ((nid >= ?")
				}
			}
			if !base.IsNumeric(f.NIDStart[i]) {
				return ErrHMSDSArgNotAnInt
			}
			q.addArg(f.NIDStart[i])
			// Check for a paired NIDEnd
			if len(f.NIDEnd) >= i+1 {
				if !base.IsNumeric(f.NIDEnd[i]) {
					return ErrHMSDSArgNotAnInt
				}
				q.appendToQuery(" AND nid <= ?)")
				q.addArg(f.NIDEnd[i])
			} else {
				q.appendToQuery(")")
			}
		}
		// OR any leftover NIDEnd values if we had more NIDEnds than NIDStarts
		for i := len(f.NIDStart); i < len(f.NIDEnd); i++ {
			if q.first == true {
				q.appendToQuery(" WHERE (nid <= ?")
				q.first = false
			} else {
				if i > 0 {
					q.appendToQuery(" OR nid <= ?")
				} else {
					q.appendToQuery(" AND (nid <= ?")
				}
			}
			if !base.IsNumeric(f.NIDEnd[i]) {
				return ErrHMSDSArgNotAnInt
			}
			q.addArg(f.NIDEnd[i])
		}
		// OR any NID values if we have them
		if len(f.NID) > 0 {
			for i := 0; i < len(f.NID); i++ {
				if !base.IsNumeric(f.NID[i]) {
					return ErrHMSDSArgNotAnInt
				}
				q.appendToQuery(" OR nid = ?")
				q.addArg(f.NID[i])
			}
		}
		// If no start, presume a start of positive nids.
		if len(f.NIDStart) == 0 {
			q.appendToQuery(") AND nid >= 0 ")
		} else {
			q.appendToQuery(")")
		}
	} else if len(f.NID) > 0 {
		// No NIDStart or NIDEnds. Just NIDs.
		err = q.doQueryArg("nid", f.NID, nidStrTransform)
		if err != nil {
			return ErrHMSDSArgNotAnInt
		}
	}
	// Terminate statement
	if cont == false {
		if f.writeLock == true {
			q.appendToQuery(" FOR UPDATE;")
		} else {
			q.appendToQuery(";")
		}
	}
	return nil
}

// Puts together custom query of HMS Components collection to filter based
// on query strings provided by user and a list of component xnames.
// NOTE: baseQuery ends up as a subquery as a result of this function.
func buildBulkCompUpdateQuery(baseQuery string, ids []string) (string, []interface{}, error) {
	args := make([]interface{}, 0, 1)

	// If no 'type' is specified we'll be using the 'type' of the
	// component (e.g. no component expansion). No need to use REGEXP.
	filterQuery := baseQuery
	for i := 0; i < len(ids); i++ {
		if i > 0 {
			// Multiples of the same key are OR'd together.
			filterQuery += " OR id = ?"
		} else {
			filterQuery += " WHERE (id = ?"
		}
		args = append(args, base.NormalizeHMSCompID(ids[i]))
	}
	filterQuery += ");"
	return filterQuery, args, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Helper functions - PreparedQueries
//
////////////////////////////////////////////////////////////////////////////

// Holds a preparared query string and the corresponding argument interface
// array.  Also holds a tracker to ensure only one WHERE clause is printed
// (and not printed if it is not needed)
type preparedQuery struct {
	query          string
	args           []interface{}
	numArgs        int // Postgres has unique numbered arguments.
	first          bool
	firstUpdateArg bool
	noArgs         bool
	postgresArgs   bool // default false
}

// Creates a new prepared query with base query string baseQuery and
// an allocated but zero-length argument list.
func newPreparedQuery(baseQuery string) *preparedQuery {
	p := new(preparedQuery)
	p.args = make([]interface{}, 0, 1)

	p.numArgs = 0 // Note: postgres starts numbering from one
	p.query = baseQuery
	p.first = true
	p.noArgs = false
	p.firstUpdateArg = true
	return p
}

// Creates a new prepared query with base query string baseQuery that
// already includes baseArgs
func addPreparedQuery(baseQuery string, baseArgs []interface{}) *preparedQuery {
	p := newPreparedQuery(baseQuery)

	p.numArgs = len(baseArgs)
	p.args = append(p.args, baseArgs...)

	return p
}

// Don't parameterize arguments and add them to args array, put them
// directly into query.
func (p *preparedQuery) setDirectQuery(val bool) {
	p.noArgs = val
}

// Appends parameters with name 'name' to mysql prepared query string and
// appends the matching 'params' values to args array.
// If 'parseF' is set, it is run on 'params' members before appending them to
// 'args', unless it returns the empty string with non-empty input, in which
// case it returns an error.
func (p *preparedQuery) doQueryArg(name string,
	params []string,
	parseF func(string) string,
) error {
	names := []string{name}
	paramset := [][]string{params}
	parseFs := []func(string) string{parseF}

	return p.doQueryArgSet(names, paramset, parseFs, false)
}

// Same as doQueryArg, but create a single clause that contains
// two parameter subclauses optionally separated by an OR instead of an
// AND.  Because this OR is only local to this clause, it does not have
// global effect.  For example, if we wanted to see if a new state and
// flag would update the existing state or flag, but while still
// allowing previous filters to work as intended (such as a set of nodes).
func (p *preparedQuery) doQueryArgsWithOR(name1, name2 string,
	params1, params2 []string,
	parseF1, parseF2 func(string) string,
) error {
	names := []string{name1, name2}
	paramset := [][]string{params1, params2}
	parseFs := []func(string) string{parseF1, parseF2}

	return p.doQueryArgSet(names, paramset, parseFs, true)
}

// Bulk version of doQueryArg, allowing for a set of parameters to be
// placed into a single clause.  If we OR the subclauses together, they
// will be subclauses of the same function-wide AND clause, i.e. they don't
// be OR'd with the global query, just the parameters in this call.
//
// For example, if we wanted to see if a new state and flag would
// update the existing state or flag, but while still allowing previous
// filters to work as intended (such as a set of nodes).
func (p *preparedQuery) doQueryArgSet(names []string,
	paramset [][]string,
	parseFs []func(string) string,
	doORnotAND bool,
) error {
	var isMatch, orOper string
	firstForFunc := true
	closeOuter := false

	// Create one AND clause with sub-clauses for each parameter.  Sub
	// clauses are ORed or ANDed depending on config.
	for j, params := range paramset {
		// We do two rounds per set of params.
		// The first will OR together all non-negated args, and the second
		// will AND together all negated args, each parenthezied
		// separately.
		for round := 1; round <= 2; round++ {
			// Round select - negate if we want non-matches for Round 2
			if round == 1 {
				// Check whether we add as a prepared statement arg, or
				// directly to query.
				if p.noArgs == true {
					isMatch = " = "
				} else {
					// Add only variable for prepared argument arg.
					isMatch = " = ?"
				}
				orOper = " OR "
			} else if round == 2 {
				if p.noArgs == true {
					// Don't need placeholder for prepared arg, adding
					// directly.
					isMatch = " != "
				} else {
					isMatch = " != ?"
				}
				orOper = " AND "
			}
			closeParen := false

			// Iterate through first the non-negated, then the negated
			// arugments for this parameter creating a two part sub-clause,
			// i.e. "... (AND (a OR b OR C) AND (!d AND !e AND !f) ..."
			firstForRound := -1
			for i, param := range params {
				// Check for regular or negated option.  idx will be zero
				// if option is negated.
				// Depending on the result, skip round 2 for non-negated
				// options and round 1 for the negated ones.
				idx := strings.IndexAny(param, "!")
				if idx == 0 {
					// Got negation - only for round 2
					if round != 2 {
						continue
					}
				} else if round != 1 {
					// No negation, only round 1
					continue
				}
				val := strings.TrimLeft(param, "!")
				// Mark the first non-skipped index for this round.
				if firstForRound == -1 {
					firstForRound = i
				}
				// If start of new parameter/round, we need to add a keyword
				// and parenthesize appropriately.
				if i == firstForRound {
					if firstForFunc == true {
						if p.first == true {
							// Global first and function first, start WHERE
							p.query += " WHERE ("
							p.first = false
						} else {
							// function first only, do AND
							p.query += " AND ("
						}
						// Starting clause but doing OR for subsequent
						// params.  Need double paren to keep this a
						// single AND clause or the OR will have global
						// effect.
						if doORnotAND == true {
							p.query += "("
							closeOuter = true // Outer statement opened.
						}
						firstForFunc = false
						p.query += names[j] + isMatch
					} else {
						// Starting a second or later sub-clause
						// start AND or OR
						if doORnotAND == false {
							p.query += " AND (" + names[j] + isMatch
						} else {
							p.query += " OR (" + names[j] + isMatch
						}
					}
					closeParen = true // Inner tatement was opened
				} else {
					// Not first value for this round
					// Multiples of the same key are OR'd together by default,
					// but ANDed for negation.
					p.query += orOper + names[j] + isMatch
				}
				// Do we modify this value before adding to the args?
				if parseFs[j] != nil {
					// transform this arg first, unless we get the empty tring
					arg := parseFs[j](val)
					if arg == "" && val != "" {
						return ErrHMSDSArgBadArg
					}
					// We can eiter add prepared parameters to the args
					// array, or put the arguments directly into the
					// query.  The latter makes a good verification function
					// important as there could be SQL injection attacks
					// if not otherwise prechecked.
					if p.noArgs == true {
						p.query += "'" + arg + "'"
					} else {
						// Add as an argument to a prepared statement instead
						// of adding directly to query string.
						p.args = append(p.args, arg)
					}
				} else {
					// No check function: noArgs can be especially dangerous
					// here if params are not pre-checked, i.e. SQL injection
					// attacks
					if p.noArgs == true {
						p.query += "'" + val + "'"
					} else {
						// Using prepared arguments is safer, if we can
						// do so.
						p.args = append(p.args, val)
					}
				}
			}
			// Close open parenthesis
			if closeParen {
				p.query += ")"
			}
		}
	}
	// Close double parenthesis
	if closeOuter == true {
		p.query += ")"
	}
	return nil
}

// Add to query string
func (p *preparedQuery) appendToQuery(q string) {
	p.query += q
}

// Add an arg to the argument list.  Make sure a placeholder ("?") has
// been added to the query string.
func (p *preparedQuery) addArg(arg interface{}) {
	p.args = append(p.args, arg)
	p.numArgs += 1
}

// Get arg placeholder for prepared queries.  This is more complicated
// for postgres queries because these have to be numbered.
// NOTE: Use this version if the arg itself was already added with addArg.
func (p *preparedQuery) argNum() string {
	if p.postgresArgs == true {
		return fmt.Sprintf("$%d", p.numArgs)
	}
	return "?"
}

// Get NEXT arg placeholder for prepared queries.  This is more complicated
// for postgres queries because these have to be numbered.
// NOTE: Use this version if the arg itself still needs to be added with addArg.
func (p *preparedQuery) argNext() string {
	if p.postgresArgs == true {
		return fmt.Sprintf("$%d", p.numArgs+1)
	}
	return "?"
}

// Append a comma-separated col = arg pair to the current query.  No other
// separators are used.  This would be used in a SET option, for example.
// Note that if emptyOK=true, empty values are ok as long as they were empty
// before applying parseF as well.
func (p *preparedQuery) doUpdateArg(
	name string,
	params []string,
	pCond *PCondition,
	parseF func(string) string,
	emptyOK bool,
) error {
	var separator, arg string

	// Process argument
	if len(params) == 0 {
		return nil
	} else if len(params) > 1 {
		// We're using a filter, but there should not be multiple args
		// of a single type
		return ErrHMSDSArgTooMany
	} else {
		param := params[0]
		// See if doUpdateArg has been called before.  If so, we need a
		// comma from the previous value.
		if p.firstUpdateArg == true {
			separator = " "
			p.firstUpdateArg = false
		} else {
			separator = ", "
		}
		p.query += separator + name + " = "
		// Do we modify this value before adding to the args?
		if parseF != nil {
			// transform this arg first, unless we get the empty tring
			arg = parseF(param)
			if arg == "" {
				if param != "" {
					return ErrHMSDSArgBadArg
				} else if !emptyOK {
					return ErrHMSDSArgEmpty
				}
			}
		} else {
			// No check function: noArgs can be especially dangerous
			// here if params are not pre-checked, i.e. SQL injection
			// attacks
			arg = param
		}
		// We can eiter add prepared parameters to the args
		// array, or put the arguments directly into the
		// query.  The latter makes a good verification function
		// important as there could be SQL injection attacks
		// if not otherwise prechecked.
		if p.noArgs == true {
			if pCond == nil {
				p.query += "'" + arg + "'"
			} else {
				p.query += "CASE WHEN " + name + " " + pCond.op +
					" '" + pCond.cond + "' THEN '" + pCond.value +
					"' ELSE '" + arg + "' END"
			}
		} else {
			// Add as an argument to a prepared statement instead
			// of adding directly to query string.
			if pCond == nil {
				p.query += "?"
				p.args = append(p.args, arg)
			} else {
				p.query += "CASE WHEN " + name + " " + pCond.op +
					" ? THEN ? ELSE ? END"
				p.args = append(p.args, pCond.cond, pCond.value, arg)
			}
		}
	}
	return nil
}

// Convert a true/false string, i.e. true/false/1/0/t/f (case insensitive)
// from the user into a boolean representation the DB can use directly.
// if the string is invalid, return the empty string.
func strToDbBool(boolStr string) string {
	//  Get the right arg for enabled.
	lower := strings.ToLower(boolStr)
	switch lower {
	case "true":
		fallthrough
	case "t":
		fallthrough
	case "1":
		return "1"
	case "false":
		fallthrough
	case "f":
		fallthrough
	case "0":
		return "0"
	}
	return ""
}

// Convert alpha-numeric check into a form that returns itself on
// true or returns the empty string on false (to match function
// input type for query parameter parsing).
func strToAlphaNum(checkStr string) string {
	if !base.IsAlphaNum(checkStr) {
		return ""
	}
	return checkStr
}

// Same as the above but also lowercases the result if it is alphanumeric
// else it still returns the empty string.
func strToAlphaNumLower(checkStr string) string {
	if !base.IsAlphaNum(checkStr) {
		return ""
	}
	return strings.ToLower(checkStr)
}

// For parsing nid arguments.
func nidStrTransform(checkStr string) string {
	if checkStr == "" {
		return "-1" // Same as "no entry"
	}
	// Return empty string if not number, i.e. error for non-empty,
	// non numeric value.
	if !base.IsNumeric(checkStr) {
		return ""
	}
	return checkStr
}

// If xname is valid, returns normalized xname, otherwise returns empty
// string.
func validXNameFilter(xname string) string {
	return base.VerifyNormalizeCompID(xname)
}

// If group or partion name is valid, return it normalized, else return
// empty string.
func validGroupField(label string) string {
	if label == "NULL" {
		return label
	}
	if err := sm.VerifyGroupField(label); err != nil {
		return ""
	}
	return strings.ToLower(label)
}
