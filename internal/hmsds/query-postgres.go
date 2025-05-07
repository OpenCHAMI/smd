// MIT License
//
// (C) Copyright [2018-2023] Hewlett Packard Enterprise Development LP
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
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"

	"github.com/lib/pq"
)

const pgSchema = `hmsdsuser`

//
// Component queries
//

const insertPgCompQuery = `
INSERT INTO components (
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
    locked)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    state = EXCLUDED.state,
    flag = EXCLUDED.flag,
    subtype = EXCLUDED.subtype,
    nettype = EXCLUDED.nettype,
    arch = EXCLUDED.arch,
    class = EXCLUDED.class;`

const insertPgCompReplaceAllQuery = `
INSERT INTO components (
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
    locked)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    state = EXCLUDED.state,
    flag = EXCLUDED.flag,
    enabled = EXCLUDED.enabled,
    admin = EXCLUDED.enabled,
    role = EXCLUDED.role,
    subrole = EXCLUDED.subrole,
    nid = EXCLUDED.nid,
    subtype = EXCLUDED.subtype,
    nettype = EXCLUDED.nettype,
    arch = EXCLUDED.arch,
    class = EXCLUDED.class;`

const insertPgNodeMapQuery = `
INSERT INTO node_nid_mapping (
    id,
    nid,
    role,
    subrole,
    node_info)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    id = EXCLUDED.id,
    nid = EXCLUDED.nid,
    role = EXCLUDED.role,
    subrole = EXCLUDED.subrole,
    node_info = EXCLUDED.node_info;`

const insertPgPowerMapQuery = `
INSERT INTO power_mapping (
    id,
    powered_by)
VALUES (?, ?)
ON CONFLICT(id) DO UPDATE SET
    id = EXCLUDED.id,
    powered_by = EXCLUDED.powered_by;`

//
// Hardware Inventory - Insert and update
//

const insertPgHWInvByLocQuery = `
INSERT INTO hwinv_by_loc (
    id,
    type,
    ordinal,
    status,
	parent_node,
    location_info,
    fru_id)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    ordinal = EXCLUDED.ordinal,
    status = EXCLUDED.status,
	parent_node = EXCLUDED.parent_node,
    location_info = EXCLUDED.location_info,
    fru_id = EXCLUDED.fru_id;`

const insertPgHWInvByFRUQuery = `
INSERT INTO hwinv_by_fru (
    fru_id,
    type,
    subtype,
    fru_info)
VALUES (?, ?, ?, ?)
ON CONFLICT(fru_id) DO UPDATE SET
    subtype = EXCLUDED.subtype,
    fru_info = EXCLUDED.fru_info;`

// RedfishEndpoints - Update operations
const updatePgRFEndpointPrefix = `
UPDATE rf_endpoints SET
    "type" = ?,
    "name" = ?,
    "hostname" = ?,
    "domain" = ?,
    "fqdn" = ?,
    "enabled" = ?,
    "uuid" = ?,
    "user" = ?,
    "password" = ?,
    usessdp = ?,
    macrequired = ?,
    macaddr = ?,
    ipaddr = ?,
    rediscoveronupdate = ?,
    templateid = ?,
    discovery_info = ? `

const updatePgRFEndpointNoDiscInfoPrefix = `
UPDATE rf_endpoints SET
    "type" = ?,
    "name" = ?,
    "hostname" = ?,
    "domain" = ?,
    "fqdn" = ?,
    "enabled" = ?,
    "uuid" = ?,
    "user" = ?,
    "password" = ?,
    usessdp = ?,
    macrequired = ?,
    macaddr = ?,
    ipaddr = ?,
    rediscoveronupdate = ?,
    templateid = ? `

const updatePgRFEndpointQuery = updatePgRFEndpointPrefix + suffixByID
const updatePgRFEndpointNoDiscInfoQuery = updatePgRFEndpointNoDiscInfoPrefix + suffixByID

//
// RedfishEndpoints - Insert operations
//

const insertPgRFEndpointPrefix = `
INSERT INTO rf_endpoints (
    "id",
    "type",
    "name",
    "hostname",
    "domain",
    "fqdn",
    "enabled",
    "uuid",
    "user",
    "password",
    usessdp,
    macrequired,
    macaddr,
    ipaddr,
    rediscoveronupdate,
    templateid,
    discovery_info)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) `

const upsertPgRFEndpointModifier = `
ON CONFLICT(id) DO UPDATE SET
    type = EXCLUDED.type,
    name = EXCLUDED.name,
    hostname = EXCLUDED.hostname,
    domain = EXCLUDED.domain,
    fqdn = EXCLUDED.fqdn,
    enabled = EXCLUDED.enabled,
    uuid = EXCLUDED.uuid,
    user = EXCLUDED."user",
    password = EXCLUDED.password,
    usessdp = EXCLUDED.useSSDP,
    macrequired = EXCLUDED.macRequired,
    macaddr = EXCLUDED.macAddr,
    ipaddr = EXCLUDED.ipAddr,
    rediscoveronupdate = EXCLUDED.rediscoverOnUpdate,
    templateid = EXCLUDED.templateID `

const upsertPgRFEndpointPrefix = insertPgRFEndpointPrefix + upsertPgRFEndpointModifier

// If we are just updating metadata for an existing entry we may not want
// to clobber the last discovery status info.
const updatePgDiscInfoOnDupSuffix = `,
    discovery_info = EXCLUDED.discovery_info;`

const insertPgRFEndpointQuery = insertPgRFEndpointPrefix + ";"
const upsertPgRFEndpointQuery = upsertPgRFEndpointPrefix + updatePgDiscInfoOnDupSuffix
const upsertPgRFEndpointNoDiscInfoQuery = upsertPgRFEndpointPrefix + ";"

//
// Component Endpoints - Insert/Upsert/Update
//

const upsertPgCompEndpointQuery = `
INSERT INTO comp_endpoints (
    id,
    type,
    domain,
    redfish_type,
    redfish_subtype,
    mac,
    uuid,
    odata_id,
    rf_endpoint_id,
    component_info)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    domain = EXCLUDED.domain,
    redfish_type = EXCLUDED.redfish_type,
    redfish_subtype = EXCLUDED.redfish_subtype,
    rf_endpoint_id = EXCLUDED.rf_endpoint_id,
    mac = EXCLUDED.mac,
    odata_id = EXCLUDED.odata_id,
    uuid = EXCLUDED.uuid,
    component_info = EXCLUDED.component_info;`

const upsertPgServiceEndpointQuery = `
INSERT INTO service_endpoints (
    rf_endpoint_id,
    redfish_type,
    redfish_subtype,
    uuid,
    odata_id,
    service_info)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(rf_endpoint_id, redfish_type) DO UPDATE SET
    redfish_subtype = EXCLUDED.redfish_subtype,
    odata_id = EXCLUDED.odata_id,
    uuid = EXCLUDED.uuid,
    service_info = EXCLUDED.service_info;`

//
// Discovery status
//

const upsertPgDiscoveryStatusQuery = `
INSERT INTO discovery_status (
    id,
    status,
    last_update,
    details)
VALUES (?, ?, NOW(), ?)
ON CONFLICT(id) DO UPDATE SET
    status = EXCLUDED.status,
    last_update = EXCLUDED.last_update,
    details = EXCLUDED.details;`

//
// SCNs
//

const getPgSCNSubID = `
SELECT LASTVAL();`

// Groups and partitions
const compGroupsTablePg = pgSchema + "." + compGroupsTable
const compGroupMembersTablePg = pgSchema + "." + compGroupMembersTable

////////////////////////////////////////////////////////////////////////////
//
// Row parsing routines by object type
//
///////////////////////////////////////////////////////////////////////////

// This is used for all routines that produce one string per row.  The
// record type doesn't matter, as long as the query selects one value
// per row that can be represented as a string.
func (d *hmsdbPg) scanSingleStringValue(rows *sql.Rows) (*string, error) {
	val := new(string)
	err := rows.Scan(val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// This is used for all routines that read Component struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanComponent(rows *sql.Rows, fltr FieldFilter) (*base.Component, error) {
	// Capture NID as raw int64, real field is a json.Number and won't
	// decode properly as it is not a supported type for Scan (even if
	// it is basically a string, which would normally be ok).
	var rawNID int64
	var err error

	c := new(base.Component)
	switch fltr {
	case FLTR_STATEONLY:
		err = rows.Scan(
			&c.ID,
			&c.Type,
			&c.State,
			&c.Flag)
	case FLTR_FLAGONLY:
		err = rows.Scan(
			&c.ID,
			&c.Type,
			&c.Flag)
	case FLTR_ROLEONLY:
		err = rows.Scan(
			&c.ID,
			&c.Type,
			&c.Role,
			&c.SubRole)
	case FLTR_NIDONLY:
		err = rows.Scan(
			&c.ID,
			&c.Type,
			&rawNID)
	case FLTR_ID_ONLY:
		err = rows.Scan(
			&c.ID)
	case FLTR_DEFAULT:
		c.Enabled = new(bool)
		err = rows.Scan(
			&c.ID,
			&c.Type,
			&c.State,
			&c.Flag,
			c.Enabled,
			&c.SwStatus,
			&c.Role,
			&c.SubRole,
			&rawNID,
			&c.Subtype,
			&c.NetType,
			&c.Arch,
			&c.Class,
			&c.ReservationDisabled,
			&c.Locked)
	default:
		//Not going to let an invalid filter choice ruin our day. Just get all rows.
		c.Enabled = new(bool)
		err = rows.Scan(
			&c.ID,
			&c.Type,
			&c.State,
			&c.Flag,
			c.Enabled,
			&c.SwStatus,
			&c.Role,
			&c.SubRole,
			&rawNID,
			&c.Subtype,
			&c.NetType,
			&c.Arch,
			&c.Class,
			&c.ReservationDisabled,
			&c.Locked)
	}
	if err != nil {
		return nil, err
	}
	// NID is valid if not -1.  Otherwise leave as default empty-string
	// json.Number, as this will omit the field from the produced json.
	if (fltr == FLTR_DEFAULT || fltr == FLTR_NIDONLY) && rawNID >= 0 {
		c.NID = json.Number(strconv.FormatInt(rawNID, 10))
	}
	return c, nil
}

// This is used for all routines that read NodeMap struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanNodeMap(rows *sql.Rows) (*sm.NodeMap, error) {
	var nodeInfo []byte
	var err error

	m := new(sm.NodeMap)

	err = rows.Scan(
		&m.ID,
		&m.NID,
		&m.Role,
		&m.SubRole,
		&nodeInfo)
	if err != nil {
		return nil, err
	}
	if len(nodeInfo) > 0 {
		err = json.Unmarshal(nodeInfo, &m.NodeInfo)
		if err != nil {
			d.LogAlways("Warning: scanNodeMap(): Encode details: %s", err)
		}
	}
	return m, nil
}

// This is used for all routines that read PowerMap struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanPowerMap(rows *sql.Rows) (*sm.PowerMap, error) {
	var err error

	m := new(sm.PowerMap)

	err = rows.Scan(
		&m.ID,
		pq.Array(&m.PoweredBy))
	if err != nil {
		return nil, err
	}
	// Ensure we get an empty array instead of null.
	if m.PoweredBy == nil {
		m.PoweredBy = make([]string, 0, 1)
	}
	return m, nil
}

// Replaces Scan() call when expected data type is sm.HWInvByLoc
func (d *hmsdbPg) scanHwInvByLocWithFRU(rows *sql.Rows) (*sm.HWInvByLoc, error) {
	var location_info, fru_info []byte
	var fru_id, fru_type, fru_subtype sql.NullString

	// There should be only 1 row returned.
	hwloc := new(sm.HWInvByLoc)
	err := rows.Scan(
		&hwloc.ID,
		&hwloc.Type,
		&hwloc.Ordinal,
		&hwloc.Status,
		&location_info,
		&fru_id,      // May be NULL due to outer join, i.e. if location empty
		&fru_type,    //
		&fru_subtype, //
		&fru_info)
	if err != nil {
		return nil, err
	}
	// Did location join with a corresponding FRU entry (populated location)
	if fru_id.Valid {
		hwfru := new(sm.HWInvByFRU)
		hwfru.FRUID = fru_id.String
		hwfru.Type = fru_type.String
		hwfru.Subtype = fru_subtype.String

		err = hwfru.DecodeFRUInfo(fru_info)
		if err != nil {
			d.LogAlways("Warning: scanHwInvByLocWithFRU(): DecodeFRUInfo: %s", err)
		}
		hwloc.PopulatedFRU = hwfru
	}
	err = hwloc.DecodeLocationInfo(location_info)
	if err != nil {
		d.LogAlways("Warning: scanHwInvByLocWithFRU(): DecodeLocationInfo: %s", err)
	}
	return hwloc, nil
}

// Replaces Scan() call when expected data type is sm.HWInvByFRU
func (d *hmsdbPg) scanHwInvByFRU(rows *sql.Rows) (*sm.HWInvByFRU, error) {
	var fru_info []byte

	// There should be only 1 row returned.
	hwfru := new(sm.HWInvByFRU)
	err := rows.Scan(
		&hwfru.FRUID,
		&hwfru.Type,
		&hwfru.Subtype,
		&fru_info)
	if err != nil {
		return nil, err
	}
	err = hwfru.DecodeFRUInfo(fru_info)
	if err != nil {
		d.LogAlways("Warning: scanHwInvByFRU(): DecodeFRUInfo: %s", err)
	}
	return hwfru, nil
}

// Replaces Scan() call when expected data type is sm.HWInvHist
func (d *hmsdbPg) scanHwInvHist(rows *sql.Rows) (*sm.HWInvHist, error) {
	// There should be only 1 row returned.
	hwhist := new(sm.HWInvHist)
	err := rows.Scan(
		&hwhist.ID,
		&hwhist.FruId,
		&hwhist.EventType,
		&hwhist.Timestamp)
	if err != nil {
		return nil, err
	}
	return hwhist, nil
}

// This is used for all routines that read RedfishEndpoint struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanRedfishEndpoint(rows *sql.Rows) (*sm.RedfishEndpoint, error) {
	var discovery_info []byte

	ep := new(sm.RedfishEndpoint)
	err := rows.Scan(
		&ep.ID,
		&ep.Type,
		&ep.Name,
		&ep.Hostname,
		&ep.Domain,
		&ep.FQDN,
		&ep.Enabled,
		&ep.UUID,
		&ep.User,
		&ep.Password,
		&ep.UseSSDP,
		&ep.MACRequired,
		&ep.MACAddr,
		&ep.IPAddr,
		&ep.RediscOnUpdate,
		&ep.TemplateID,
		&discovery_info)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(discovery_info, &ep.DiscInfo)
	if err != nil {
		d.LogAlways("Warning: scanRedfishEndpoint(): Decode DiscoveryInfo: %s", err)
	}
	return ep, nil
}

// This is used for all routines that read ComponentEndpoint struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanComponentEndpoint(rows *sql.Rows) (*sm.ComponentEndpoint, error) {
	var component_info []byte

	cep := new(sm.ComponentEndpoint)
	err := rows.Scan(
		&cep.ID,
		&cep.Type,
		&cep.Domain,
		&cep.RedfishType,
		&cep.RedfishSubtype,
		&cep.MACAddr,
		&cep.UUID,
		&cep.OdataID,
		&cep.RfEndpointID,
		&cep.RfEndpointFQDN,
		&component_info,
		&cep.Enabled)
	if err != nil {
		return nil, err
	}
	// Set FQDN if domain is set.
	if cep.Domain != "" {
		cep.FQDN = cep.ID + "." + cep.Domain
	}
	cep.URL = cep.RfEndpointFQDN + cep.OdataID
	err = cep.DecodeComponentInfo(component_info)
	if err != nil {
		d.LogAlways("Warning: scanComponentEndpoint(): DecodeComponentInfo: %s", err)
	}
	return cep, nil
}

// This is used for all routines that read ServiceEndpoint struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanServiceEndpoint(rows *sql.Rows) (*sm.ServiceEndpoint, error) {

	sep := new(sm.ServiceEndpoint)
	err := rows.Scan(
		&sep.RfEndpointID,
		&sep.RedfishType,
		&sep.RedfishSubtype,
		&sep.UUID,
		&sep.OdataID,
		&sep.RfEndpointFQDN,
		//&sep.URL,
		&sep.ServiceInfo)
	if err != nil {
		return nil, err
	}
	sep.URL = sep.RfEndpointFQDN + sep.OdataID
	// err = cep.DecodeComponentInfo(component_info)
	// if err != nil {
	// d.LogAlways("Warning: scanComponentEndpoint(): DecodeComponentInfo: %s", err)
	// }
	return sep, nil
}

// This is used for all routines that read ComponentEndpoint struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanDiscoveryStatus(rows *sql.Rows) (*sm.DiscoveryStatus, error) {
	var details []byte

	st := new(sm.DiscoveryStatus)
	err := rows.Scan(
		&st.ID,
		&st.Status,
		&st.LastUpdate,
		&details)
	if err != nil {
		return nil, err
	}
	if len(details) > 0 {
		err = json.Unmarshal(details, &st.Details)
		if err != nil {
			d.LogAlways("Warning: scanDiscoveryStatus(): Encode details: %s", err)
		}
	}
	return st, nil
}

// This is used for all routines that read SCN subscription struct as rows and
// replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanSCNSubscription(rows *sql.Rows) (*sm.SCNSubscription, error) {
	var id int64
	var jsonSub []byte
	var err error

	sub := new(sm.SCNSubscription)

	err = rows.Scan(
		&id,
		&jsonSub)
	if err != nil {
		return nil, err
	}
	if len(jsonSub) > 0 {
		err = json.Unmarshal(jsonSub, sub)
		if err != nil {
			d.LogAlways("Warning: scanSCNSubscription(): Decode details: %s", err)
		}
	}
	sub.ID = id
	return sub, nil
}

//
// Groups and partitions
//

// This is used for all routines that read group entries (sans members) as
// rows and replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanPgGroup(rows *sql.Rows) (uuid string, g *sm.Group, err error) {
	g = new(sm.Group)
	err = rows.Scan(
		&uuid,
		&g.Label,
		&g.Description,
		pq.Array(&g.Tags), // tags
		&g.ExclusiveGroup)
	if err != nil {
		uuid = ""
		g = nil
		return
	}
	// Ensure we get an empty array instead of null.
	if g.Tags == nil {
		g.Tags = make([]string, 0, 1)
	}
	return
}

// This is used for all routines that read group entries (sans members) as
// rows and replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanPgPartition(rows *sql.Rows) (
	uuid string, p *sm.Partition, err error,
) {
	p = new(sm.Partition)
	err = rows.Scan(
		&uuid,
		&p.Name,
		&p.Description,
		pq.Array(&p.Tags)) // tags
	if err != nil {
		uuid = ""
		p = nil
		return
	}
	// Ensure we get an empty array instead of null.
	if p.Tags == nil {
		p.Tags = make([]string, 0, 1)
	}
	return
}

// This is used for all routines that read job sync entries (sans job data) as
// rows and replaces rows.Scan in normal usage.
func (d *hmsdbPg) scanPgJob(rows *sql.Rows) (
	j *sm.Job, err error,
) {
	j = new(sm.Job)
	err = rows.Scan(
		&j.Id,
		&j.Type,
		&j.Status,
		&j.LastUpdate,
		&j.Lifetime)
	if err != nil {
		j = nil
	}
	return
}

////////////////////////////////////////////////////////////////////////////
//
// Helper functions - Postgres support
//
////////////////////////////////////////////////////////////////////////////

// Replaces each '?' (MySQL-style query arg) with an ordered $1, $2, ...
// Postgres equivalent.  NOTE: This function is dumb and assumes every
// ? is a variable.  In other words, escaping/quoting/whatever ? does nothing
// and the caller must be OK with this.
func ToPGQueryArgs(query string) string {
	// NOTE: Split always includes <#separators>+1 substrings, so we will
	// always have a leading and trailing string, even if they're "".
	qSplit := strings.Split(query, "?")

	// If there are no '?'s the for loop won't run and we're done here.
	newQuery := qSplit[0]

	// Replace where the ? was with ${i}, then add back the next portion of
	// the string up to, but not including the next ? (or string end if i=len)
	for i := 1; i < len(qSplit); i++ {
		newQuery += fmt.Sprintf("$%d", i) + qSplit[i]
	}
	return newQuery
}
