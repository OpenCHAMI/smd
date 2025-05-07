// MIT License
//
// (C) Copyright [2019-2022, 2024] Hewlett Packard Enterprise Development LP
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
	"strings"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

type FieldFilter int

const (
	FLTR_DEFAULT     FieldFilter = 0
	FLTR_STATEONLY   FieldFilter = 1
	FLTR_FLAGONLY    FieldFilter = 2
	FLTR_ROLEONLY    FieldFilter = 3
	FLTR_NIDONLY     FieldFilter = 4
	FLTR_ALL_W_GROUP FieldFilter = 5
	FLTR_ID_W_GROUP  FieldFilter = 6
	FLTR_ID_ONLY     FieldFilter = 7
	FLTR_MAX         FieldFilter = 8
)

type PCondition struct {
	op    string // Operator (EX '=' '!=')
	cond  string // Value to match
	value string // Value to set if the condition is met
}

type ComponentFilter struct {
	// User-writable options
	ID        []string `json:"id"`
	NID       []string `json:"nid"`
	NIDStart  []string `json:"nid_start"`
	NIDEnd    []string `json:"nid_end"`
	Type      []string `json:"type"`
	State     []string `json:"state"`
	Flag      []string `json:"flag"`
	Enabled   []string `json:"enabled"`
	SwStatus  []string `json:"softwarestatus"`
	Role      []string `json:"role"`
	SubRole   []string `json:"subrole"`
	Subtype   []string `json:"subtype"`
	Arch      []string `json:"arch"`
	Class     []string `json:"class"`
	Group     []string `json:"group"` // Arbitrary Groups have UUID ids as well as globally unique names
	Partition []string `json:"partition"`
	Locked    []string `json:"locked"`
	ReservationDisabled []string `json:"reservation_disabled"`

	// private options
	writeLock bool   // default is false
	label     string // Labels query for logging, etc.

	// State OR flag subclause without ORing the whole query.  For the
	// target state and clause, since one or the other can be right but
	// the other still needs to be changed (done as !TargetState OR !TargetFlag
	// since otherwise we don't need to change either).
	//
	// Non OR'd state arguments are for formal filtering, or for
	// restricting the starting state to keep to a state transition
	// graph.  If we don't mix the types we can use both in the same query.
	orState []string
	orFlag  []string

	// Same thing is needed to find components whose role or subrole differ
	// from the desired values
	orRole []string
	orSubRole []string

	flagCondition *PCondition

	// Has VerifyAndNormalize been run?
	verified bool
}

type CompEPFilter struct {
	// User-writable options
	ID           []string `json:"id"`
	RfEndpointID []string `json:"redfish_ep"`
	Type         []string `json:"type"`
	RedfishType  []string `json:"redfish_type"`

	// private options
	writeLock bool   // default is false
	label     string // Labels query for logging, etc.
}

type RedfishEPFilter struct {
	// User-writable options
	ID         []string `json:"id"`
	FQDN       []string `json:"fqdn"`
	Type       []string `json:"type"`
	UUID       []string `json:"uuid"`
	MACAddr    []string `json:"macaddr"`
	IPAddr     []string `json:"ipaddress"`
	LastStatus []string `json:"laststatus"`

	// private options
	writeLock bool   // default is false
	label     string // Labels query for logging, etc.
}

type ServiceEPFilter struct {
	// User-writable options
	Service      []string `json:"service"`
	RfEndpointID []string `json:"redfish_ep"`

	// private options
	writeLock bool   // default is false
	label     string // Labels query for logging, etc.
}

type JobSyncFilter struct {
	// User-writable options
	ID     []string `json:"id"`
	Type   []string `json:"type"`
	Status []string `json:"status"`

	// private options
	isExpired bool
	label     string // Labels query for logging, etc.
}

type HWInvLocFilter struct {
	// User-writable options
	ID           []string `json:"id"`
	Type         []string `json:"type"`
	Manufacturer []string `json:"manufacturer"`
	PartNumber   []string `json:"partnumber"`
	SerialNumber []string `json:"serialnumber"`
	FruId        []string `json:"fruid"`
	Children     bool     `json:"children"`
	Parents      bool     `json:"parents"`
	Partition    []string `json:"partition"`

	// private options
	label string // Labels query for logging, etc.
}

type HWInvHistFilter struct {
	// User-writable options
	ID        []string `json:"id"`
	FruId     []string `json:"fruid"`
	EventType []string `json:"eventtype"`
	StartTime string   `json:"starttime"`
	EndTime   string   `json:"endtime"`

	// private options
	label string // Labels query for logging, etc.
}

type CompEthInterfaceFilter struct {
	// User-writable options
	ID        []string `json:"id"`
	MACAddr   []string `json:"macaddr"`
	IPAddr    []string `json:"ipaddr"`
	Network   []string `json:"network"`
	NewerThan string   `json:"newerthan"`
	OlderThan string   `json:"olderthan"`
	CompID    []string `json:"compID"`
	CompType  []string `json:"type"`

	// private options
	label string // Labels query for logging, etc.
}

//
//  Helper functions
//

func filtStringArray(strs []string, f *[]string) {
	if len(strs) == 0 {
		*f = []string{""}
	} else {
		*f = strs
	}
}

////////////////////////////////////////////////////////////////////////////
//  Component Filter options - These are intended to be used as variadic
//                             function arguments, i.e. they are combined to
//                             generate a filter but in a much more flexible
//                             way that doesn't require unneeded arguments or
//                             require changing the function def to add more.
////////////////////////////////////////////////////////////////////////////

// Filter functions: must take a pointer to a ComponentFilter presumed to be
// already initialized and modify the filter accordingly.
type CompFiltFunc func(*ComponentFilter)

// Filter includes just these ids.  Overwrites previous ID call.  IDs can be
// negated with "!" and all such ids will be excluded.
//
// NOTE: will add the empty string if ids is zero length to select no ids.
//       The assumption is that this isn't being used to select any ID as
//       this option would be unneccessary otherwise.
func IDs(ids []string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.ID = []string{""}
			} else {
				f.ID = ids
			}
		}
	}
}

// Filter includes just this id.  Overwrites other ID calls.
// No negated/wildcard options allowed.  Non-alpha-num strings will be
// converted into the empty string that will match nothing
func ID(id string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			if !base.IsAlphaNum(id) {
				f.ID = []string{""}
			} else {
				f.ID = []string{id}
			}
		}
	}
}

// Filter should include this type. Appends to earlier call.  Types can be
// negated with "!", and all such negated types are excluded.
func Type(hmsType string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.Type = append(f.Type, hmsType)
		}
	}
}

// Select xnames with either state OR with flag. Replaces any earlier
// call.  State/flag can be negated with "!", and any lines already set to
// state AND flag are excluded from the query result and possible update.
//
// Use this to see if a State/Flag change would be redundant for both values.
func StateOrFlag(state, flag string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.orState = []string{"!" + state}
			f.orFlag = []string{"!" + flag}
		}
	}
}

// Like the above, but with the negation done for you.  Example, really.
func NotStateOrFlag(state, flag string) CompFiltFunc {
	return StateOrFlag("!"+state, "!"+flag)
}

// Select xnames with either role OR with subrole. Replaces any earlier
// call.  Role/subrole can be negated with "!", and any lines already set to
// role AND subrole are excluded from the query result and possible update.
//
// Use this to see if a Role/SubRole change would be redundant for both values.
func RoleOrSubRole(role, subRole string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.orRole = []string{"!" + role}
			f.orSubRole = []string{"!" + subRole}
		}
	}
}

// Like the above, but with the negation done for you.  Example, really.
func NotRoleOrSubRole(role, subRole string) CompFiltFunc {
	return RoleOrSubRole("!"+role, "!"+subRole)
}

// Filter selects rows with just these states.  Overwrites any previous State,
// or States call.  States can be negated with "!" and any lines with the
// negated state(s) will be excluded from consideration for the update.
func States(states []string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.State = states
		}
	}
}

// Filter should select entries with this State.  Appends to any previous
// State* call.  State can be negated with "!" and rows with the State will be
// excluded.  Negated values are ANDED together, non-negated values
func State(state string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.State = append(f.State, state)
		}
	}
}

// Filter should select entries with this flag.  Appends to any previous
// Flag call.  Flag can be negated with "!" and rows with the flag will be
// excluded.
func Flag(flag string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.Flag = append(f.Flag, flag)
		}
	}
}

func FlagCondNoChange(flag string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			if len(flag) != 0 {
				idx := strings.IndexAny(flag, "!")
				if idx > 0 {
					// Avoid overwriting if invalid
					return
				}
				f.flagCondition = new(PCondition)
				if idx == 0 {
					f.flagCondition.op = " != "
					f.flagCondition.cond = strings.TrimLeft(flag, "!")
					f.flagCondition.value = f.flagCondition.cond
				} else if idx < 0 {
					f.flagCondition.op = " = "
					f.flagCondition.cond = flag
					f.flagCondition.value = flag
				}
			}
		}
	}
}

// Filter on enabled=true/false (in string form).  Replaces earlier call.
func Enabled(enabled string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.Enabled = []string{enabled}
		}
	}
}

// Filter should include this SoftwareStatus. Appends to earlier call.
// SoftwareStatuses can be negated with "!", and all such negated values
// are excluded.
func SwStatus(swstatus string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.SwStatus = append(f.SwStatus, swstatus)
		}
	}
}

// Filter should include this Role.  Appends to earlier call.
// Roles can be negated with "!", and all such negated roles
// are excluded.
func Role(role string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.Role = append(f.Role, role)
		}
	}
}

// Filter should include this SubRole.  Appends to earlier call.
// SubRoles can be negated with "!", and all such negated subroles
// are excluded.
func SubRole(subrole string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.SubRole = append(f.SubRole, subrole)
		}
	}
}

// Filter should include this Class. Appends to earlier call.
// Classes can be negated with "!", and all such negated classes
// are excluded.
func Class(class string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.Class = append(f.Class, class)
		}
	}
}

// Filter should include this Group.  Appends to earlier call.
// Groups are arbitrary and will potentially overlap
func Group(group string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.Group = append(f.Group, group)
		}
	}
}

// Filter should limit themselves to this partition.
// Partitions are exclusive. A node can only be part of one
// at a time.
func Partition(partition string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.Partition = append(f.Partition, partition)
		}
	}
}

// If PartInfo is non-nil, replace the Partition and Group
// arrays in the filter with the ones in pi.  Overwrites previous
// calls to Partition/Group
func PI(pi *PartInfo) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil && pi != nil {
			f.Partition = pi.Partition
			f.Group = pi.Group
		}
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func From(callingFunc string) CompFiltFunc {
	return func(f *ComponentFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}

// Sets write lock for transaction on rows hit by filter by using
// FOR UPDATE.
func WRLock(f *ComponentFilter) {
	if f != nil {
		f.writeLock = true
	}
}

//                                                                           //
//            ComponentFilter - Verification and normalization               //
//                                                                           //

// Verify and normalize f for querying (modifying it), but skipping if this has
// already been done.
func (f *ComponentFilter) VerifyNormalize() error {
	if f == nil {
		return nil
	}
	if f.verified == true {
		return nil
	}
	f.verified = true

	// Verify and normalize each field.
	err := checkFilterField(f.ID, validXNameFilter, false)
	if err != nil {
		return ErrHMSDSArgBadID
	}
	err = checkFilterField(f.Type, xnametypes.VerifyNormalizeType, true)
	if err != nil {
		return ErrHMSDSArgBadType
	}
	err = checkFilterField(f.State, base.VerifyNormalizeState, false)
	if err != nil {
		return ErrHMSDSArgBadState
	}
	err = checkFilterField(f.Flag, base.VerifyNormalizeFlag, false)
	if err != nil {
		return ErrHMSDSArgBadFlag
	}
	err = checkFilterField(f.orState, base.VerifyNormalizeState, false)
	if err != nil {
		return ErrHMSDSArgBadState
	}
	err = checkFilterField(f.orFlag, base.VerifyNormalizeFlag, false)
	if err != nil {
		return ErrHMSDSArgBadFlag
	}
	err = checkFilterField(f.orRole,  base.VerifyNormalizeRole, true)
	if err != nil {
		return ErrHMSDSArgBadRole
	}
	err = checkFilterField(f.orSubRole, base.VerifyNormalizeSubRole, true)
	if err != nil {
		return ErrHMSDSArgBadSubRole
	}
	// Skip SwStatus
	//
	err = checkFilterField(f.Enabled, strToDbBool, false)
	if err != nil {
		return ErrHMSDSArgNoMatch
	}
	err = checkFilterField(f.Role, base.VerifyNormalizeRole, true)
	if err != nil {
		return ErrHMSDSArgBadRole
	}
	err = checkFilterField(f.SubRole, base.VerifyNormalizeSubRole, true)
	if err != nil {
		return ErrHMSDSArgBadSubRole
	}
	err = checkFilterField(f.Subtype, strToAlphaNum, true)
	if err != nil {
		return ErrHMSDSArgBadSubtype
	}
	err = checkFilterField(f.Arch, base.VerifyNormalizeArch, true)
	if err != nil {
		return ErrHMSDSArgBadArch
	}
	err = checkFilterField(f.Class, base.VerifyNormalizeClass, true)
	if err != nil {
		return ErrHMSDSArgBadClass
	}
	err = checkFilterField(f.NID, nidStrTransform, true)
	if err != nil {
		return ErrHMSDSArgNotAnInt
	}
	err = checkFilterField(f.NIDStart, nidStrTransform, false)
	if err != nil {
		return ErrHMSDSArgNotAnInt
	}
	err = checkFilterField(f.NIDEnd, nidStrTransform, false)
	if err != nil {
		return ErrHMSDSArgNotAnInt
	}
	// No negation allowed - won't work right.
	err = checkFilterFieldStrict(f.Group, validGroupField)
	if err != nil {
		return ErrHMSDSNoGroup
	}
	err = checkFilterFieldStrict(f.Partition, validGroupField)
	if err != nil {
		return ErrHMSDSNoPartition
	}
	return nil
}

// Worker for above with plug-in function for verification.
func checkFilterField(field []string, parseF func(string) string, emptyOk bool) error {
	if field == nil {
		return nil
	}
	for i, str := range field {
		val := ""
		replaceNeg := false
		if strings.HasPrefix(str, "!") {
			replaceNeg = true
			val = parseF(strings.TrimLeft(str, "!"))
		} else {
			val = parseF(str)
		}
		if val == "" {
			if emptyOk == false || strings.TrimLeft(str, "!") != "" {
				return ErrHMSDSArgBadArg
			}
		}
		if replaceNeg {
			field[i] = "!" + val
		} else {
			field[i] = val
		}
	}
	return nil
}

// Strict worker for above with plug-in function for verification.  No
// negation, no empty fields.
func checkFilterFieldStrict(field []string, parseF func(string) string) error {
	if field == nil {
		return nil
	}
	for i, str := range field {
		val := parseF(str)
		if val == "" {
			return ErrHMSDSArgBadArg
		}
		field[i] = val
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////
//  CompEP (ComponentEndpoint) Filter options
////////////////////////////////////////////////////////////////////////////

// Filter func args: must take a pointer to a CompEPFilter (ComponentEndpoint
// filter) presumed to be already initialized and modify the filter accordingly.
type CompEPFiltFunc func(*CompEPFilter)

// Filter includes just these ids.  Overwrites previous ID call.  IDs can be
// negated with "!" and all such ids will be excluded.
//
// NOTE: will add the empty string if ids is zero length to select no ids.
//       The assumption is that this isn't being used to select any ID as
//       this option would be unneccessary otherwise.
func CE_IDs(ids []string) CompEPFiltFunc {
	return func(f *CompEPFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.ID = []string{""}
			} else {
				f.ID = ids
			}
		}
	}
}

// Filter includes just this id.  Overwrites other ID calls.
// No negated/wildcard options allowed.  Non-alpha-num strings will be
// converted into the empty string that will match nothing.
func CE_ID(id string) CompEPFiltFunc {
	return func(f *CompEPFilter) {
		if f != nil {
			if !base.IsAlphaNum(id) {
				f.ID = []string{""}
			} else {
				f.ID = []string{id}
			}
		}
	}
}

// Filter should include this type. Appends to earlier call.  Types can be
// negated with "!", and all such negated types are excluded.
func CE_Type(hmsType string) CompEPFiltFunc {
	return func(f *CompEPFilter) {
		if f != nil {
			f.Type = append(f.Type, hmsType)
		}
	}
}

// Filter includes just those ComponentEndpoints with the given parent
// Redfish EP ids.  Overwrites previous call.
// rfEndpointIDs can be negated with "!" and all such ids will be excluded.
func CE_RfEPs(rfEndpointIDs []string) CompEPFiltFunc {
	return func(f *CompEPFilter) {
		if f != nil {
			f.RfEndpointID = rfEndpointIDs
		}
	}
}

// Filter should include this RedfishType. Appends to earlier call.  RFTypes
// can be negated with "!", and all such negated values are excluded.
func CE_RfType(redfishType string) CompEPFiltFunc {
	return func(f *CompEPFilter) {
		if f != nil {
			f.RedfishType = append(f.RedfishType, redfishType)
		}
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func CE_From(callingFunc string) CompEPFiltFunc {
	return func(f *CompEPFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}

// Sets write lock for transaction on rows hit by filter by using
// FOR UPDATE.
func CE_WRLock(f *CompEPFilter) {
	if f != nil {
		f.writeLock = true
	}
}

////////////////////////////////////////////////////////////////////////////
//  RedfishEP (ComponentEndpoint) Filter options
////////////////////////////////////////////////////////////////////////////

// Filter functions: must take a pointer to a RedfishEPFilter presumed to be
// already initialized and modify the filter accordingly.
type RedfishEPFiltFunc func(*RedfishEPFilter)

// Filter includes just these ids.  Overwrites previous ID call.  IDs can be
// negated with "!" and all such ids will be excluded.
//
// NOTE: will add the empty string if ids is zero length to select no ids.
//       The assumption is that this isn't being used to select any ID as
//       this option would be unneccessary otherwise.
func RFE_IDs(ids []string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.ID = []string{""}
			} else {
				f.ID = ids
			}
		}
	}
}

// Filter includes just this id.  Overwrites other ID calls.
// No negated/wildcard options allowed.  Non-alpha-num strings will be
// converted into the empty string that will match nothing.
func RFE_ID(id string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			if !base.IsAlphaNum(id) {
				f.ID = []string{""}
			} else {
				f.ID = []string{id}
			}
		}
	}
}

// Filter should include this type. Appends to earlier call.  Types can be
// negated with "!", and all such negated types are excluded.
func RFE_Type(hmsType string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			f.Type = append(f.Type, hmsType)
		}
	}
}

// Filter should include this FQDN. Appends to earlier call.  Values can be
// negated with "!", and all such negated values are excluded.
func RFE_FQDN(fqdn string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			f.FQDN = append(f.FQDN, fqdn)
		}
	}
}

// Filter should include this UUID. Appends to earlier call.  Values can be
// negated with "!", and all such negated values are excluded.
func RFE_UUID(uuid string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			f.UUID = append(f.UUID, uuid)
		}
	}
}

// Filter should include this MACAddr. Appends to earlier call.  Values can be
// negated with "!", and all such negated values are excluded.
func RFE_MACAddr(macAddr string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			f.MACAddr = append(f.MACAddr, macAddr)
		}
	}
}

// Filter should include this IPAddr. Appends to earlier call.  Values can be
// negated with "!", and all such negated values are excluded.
func RFE_IPAddr(ipAddr string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			f.IPAddr = append(f.IPAddr, ipAddr)
		}
	}
}

// Filter should include this DiscoveryInfo.LastStatus. Appends to earlier call.
// Values can be negated with "!", and all such negated values are excluded.
func RFE_LastStatus(status string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			f.LastStatus = append(f.LastStatus, status)
		}
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func RFE_From(callingFunc string) RedfishEPFiltFunc {
	return func(f *RedfishEPFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}

// Sets write lock for transaction on rows hit by filter by using
// FOR UPDATE.
func RFE_WRLock(f *RedfishEPFilter) {
	if f != nil {
		f.writeLock = true
	}
}

////////////////////////////////////////////////////////////////////////////
//  ServiceEP (ServiceEndpoint) Filter options
////////////////////////////////////////////////////////////////////////////

// Filter func args: must take a pointer to a ServiceEPFilter (ServiceEndpoint
// filter) presumed to be already initialized and modify the filter accordingly.
type ServiceEPFiltFunc func(*ServiceEPFilter)

// Filter should include this redfish service. Appends to earlier call.
// RFTypes can be negated with "!", and all such negated values are excluded.
func SE_RfEP(rfEndpointID string) ServiceEPFiltFunc {
	return func(f *ServiceEPFilter) {
		if f != nil {
			f.RfEndpointID = append(f.RfEndpointID, rfEndpointID)
		}
	}
}

// Filter includes just those ServiceEndpoints with the given parent
// Redfish EP ids.  Overwrites previous call.
// rfEndpointIDs can be negated with "!" and all such ids will be excluded.
func SE_RfEPs(rfEndpointIDs []string) ServiceEPFiltFunc {
	return func(f *ServiceEPFilter) {
		if f != nil {
			f.RfEndpointID = rfEndpointIDs
		}
	}
}

// Filter should include this redfish service. Appends to earlier call.
// RFTypes can be negated with "!", and all such negated values are excluded.
func SE_RfSvc(rfService string) ServiceEPFiltFunc {
	return func(f *ServiceEPFilter) {
		if f != nil {
			f.Service = append(f.Service, rfService)
		}
	}
}

// Filter should include this redfish service. Overwrites previous call.
// RFTypes can be negated with "!", and all such negated values are excluded.
func SE_RfSvcs(rfServices []string) ServiceEPFiltFunc {
	return func(f *ServiceEPFilter) {
		if f != nil {
			f.Service = rfServices
		}
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func SE_From(callingFunc string) ServiceEPFiltFunc {
	return func(f *ServiceEPFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}

// Sets write lock for transaction on rows hit by filter by using
// FOR UPDATE.
func SE_WRLock(f *ServiceEPFilter) {
	if f != nil {
		f.writeLock = true
	}
}

////////////////////////////////////////////////////////////////////////////
//  Job Sync filter options
//  - These are intended to be used as variadic function arguments, i.e. they
//    are combined to generate a filter but in a much more flexible way that
//    doesn't require unneeded arguments or require changing the function def
//    to add more.
////////////////////////////////////////////////////////////////////////////

// Filter functions: must take a pointer to a JobSyncFilter presumed to be
// already initialized and modify the filter accordingly.
type JobSyncFiltFunc func(*JobSyncFilter)

// Filter includes just these job ids.
func JS_IDs(ids []string) JobSyncFiltFunc {
	return func(f *JobSyncFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.ID = []string{}
			} else {
				f.ID = ids
			}
		}
	}
}

// Filter includes just this job id.  Overwrites other ID calls.
func JS_ID(id string) JobSyncFiltFunc {
	return func(f *JobSyncFilter) {
		if f != nil {
			f.ID = []string{id}
		}
	}
}

// Filter includes just jobs with these types.
func JS_Types(jobTypes []string) JobSyncFiltFunc {
	return func(f *JobSyncFilter) {
		if f != nil {
			if len(jobTypes) == 0 {
				f.Type = []string{}
			} else {
				f.Type = jobTypes
			}
		}
	}
}

// Filter includes just jobs that have this type. Overwrites other Type calls.
func JS_Type(jobType string) JobSyncFiltFunc {
	return func(f *JobSyncFilter) {
		if f != nil {
			f.Type = []string{jobType}
		}
	}
}

// Filter includes just jobs with these statuses.
func JS_Status_List(statusList []string) JobSyncFiltFunc {
	return func(f *JobSyncFilter) {
		if f != nil {
			if len(statusList) == 0 {
				f.Status = []string{}
			} else {
				f.Status = statusList
			}
		}
	}
}

// Filter includes just jobs that contain this status. Overwrites other status calls.
func JS_Status(status string) JobSyncFiltFunc {
	return func(f *JobSyncFilter) {
		if f != nil {
			f.Status = []string{status}
		}
	}
}

// Filter for expired jobs.
func JS_Expired(f *JobSyncFilter) {
	if f != nil {
		f.isExpired = true
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func JS_From(callingFunc string) JobSyncFiltFunc {
	return func(f *JobSyncFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//  HWInvLoc (by loc) Filter options
////////////////////////////////////////////////////////////////////////////

// Filter functions: must take a pointer to a HWInvLocFilter presumed to be
// already initialized and modify the filter accordingly.
type HWInvLocFiltFunc func(*HWInvLocFilter)

// Filter includes just these ids.  Overwrites previous ID call.  IDs can be
// negated with "!" and all such ids will be excluded.
//
// NOTE: will add the empty string if ids is zero length to select no ids.
//       The assumption is that this isn't being used to select any ID as
//       this option would be unneccessary otherwise.
func HWInvLoc_IDs(ids []string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.ID = []string{""}
			} else {
				f.ID = ids
			}
		}
	}
}

// Filter includes just this id.  Overwrites other ID calls.
// No negated/wildcard options allowed.  Non-alpha-num strings will be
// converted into the empty string that will match nothing.
func HWInvLoc_ID(id string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			if !base.IsAlphaNum(id) {
				f.ID = []string{""}
			} else {
				f.ID = []string{id}
			}
		}
	}
}

// Filter should include this type. Appends to earlier call.  Types can be
// negated with "!", and all such negated types are excluded.
func HWInvLoc_Type(hmsType string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			f.Type = append(f.Type, hmsType)
		}
	}
}

func HWInvLoc_Types(hmsTypes []string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			if len(hmsTypes) == 0 {
				f.Type = []string{}
			} else {
				f.Type = hmsTypes
			}
		}
	}
}

func HWInvLoc_Manufacturers(manufacturers []string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			if len(manufacturers) == 0 {
				f.Manufacturer = []string{}
			} else {
				f.Manufacturer = manufacturers
			}
		}
	}
}

func HWInvLoc_PartNumbers(partNumbers []string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			if len(partNumbers) == 0 {
				f.PartNumber = []string{}
			} else {
				f.PartNumber = partNumbers
			}
		}
	}
}

func HWInvLoc_SerialNumbers(serialNumbers []string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			if len(serialNumbers) == 0 {
				f.SerialNumber = []string{}
			} else {
				f.SerialNumber = serialNumbers
			}
		}
	}
}

func HWInvLoc_FruIDs(fruIds []string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			if len(fruIds) == 0 {
				f.FruId = []string{}
			} else {
				f.FruId = fruIds
			}
		}
	}
}

// Filter to include child components.
func HWInvLoc_Child(f *HWInvLocFilter) {
	if f != nil {
		f.Children = true
	}
}

// Filter to include parent components.
func HWInvLoc_Parent(f *HWInvLocFilter) {
	if f != nil {
		f.Parents = true
	}
}

// Filter should include only components in this partition. Appends to earlier call.
// Values can be negated with "!", and all such negated values are excluded.
func HWInvLoc_Part(part string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			f.Partition = append(f.Partition, part)
		}
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func HWInvLoc_From(callingFunc string) HWInvLocFiltFunc {
	return func(f *HWInvLocFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//  HWInvHist Filter options
////////////////////////////////////////////////////////////////////////////

// Filter functions: must take a pointer to a HWInvLocFilter presumed to be
// already initialized and modify the filter accordingly.
type HWInvHistFiltFunc func(*HWInvHistFilter)

// Filter includes just these ids.  Overwrites previous ID call.  IDs can be
// negated with "!" and all such ids will be excluded.
//
// NOTE: will add the empty string if ids is zero length to select no ids.
//       The assumption is that this isn't being used to select any ID as
//       this option would be unneccessary otherwise.
func HWInvHist_IDs(ids []string) HWInvHistFiltFunc {
	return func(f *HWInvHistFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.ID = []string{""}
			} else {
				f.ID = ids
			}
		}
	}
}

// Filter includes just this id.  Overwrites other ID calls.
// No negated/wildcard options allowed.  Non-alpha-num strings will be
// converted into the empty string that will match nothing.
func HWInvHist_ID(id string) HWInvHistFiltFunc {
	return func(f *HWInvHistFilter) {
		if f != nil {
			if !base.IsAlphaNum(id) {
				f.ID = []string{""}
			} else {
				f.ID = []string{id}
			}
		}
	}
}

func HWInvHist_FruIDs(fruIds []string) HWInvHistFiltFunc {
	return func(f *HWInvHistFilter) {
		if f != nil {
			if len(fruIds) == 0 {
				f.FruId = []string{}
			} else {
				f.FruId = fruIds
			}
		}
	}
}

func HWInvHist_EventTypes(eventTypes []string) HWInvHistFiltFunc {
	return func(f *HWInvHistFilter) {
		if f != nil {
			if len(eventTypes) == 0 {
				f.EventType = []string{}
			} else {
				f.EventType = eventTypes
			}
		}
	}
}

// Filter should include entries that occur after this time window start.
func HWInvHist_StartTime(startTime string) HWInvHistFiltFunc {
	return func(f *HWInvHistFilter) {
		if f != nil {
			f.StartTime = startTime
		}
	}
}

// Filter should include entries that occur before this time window start.
func HWInvHist_EndTime(endTime string) HWInvHistFiltFunc {
	return func(f *HWInvHistFilter) {
		if f != nil {
			f.EndTime = endTime
		}
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func HWInvHist_From(callingFunc string) HWInvHistFiltFunc {
	return func(f *HWInvHistFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//  CompEthInterface Filter options
////////////////////////////////////////////////////////////////////////////

// Filter functions: must take a pointer to a CompEthInterfaceFilter presumed to be
// already initialized and modify the filter accordingly.
type CompEthInterfaceFiltFunc func(*CompEthInterfaceFilter)

// Filter includes just these ids.  Overwrites previous ID call.  IDs can be
// negated with "!" and all such ids will be excluded.
//
// NOTE: will add the empty string if ids is zero length to select no ids.
//       The assumption is that this isn't being used to select any ID as
//       this option would be unneccessary otherwise.
func CEI_IDs(ids []string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.ID = []string{""}
			} else {
				f.ID = ids
			}
		}
	}
}

// Filter includes just this id.  Overwrites other ID calls.
// No negated/wildcard options allowed.  Non-alpha-num strings will be
// converted into the empty string that will match nothing.
func CEI_ID(id string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			if !base.IsAlphaNum(id) {
				f.ID = []string{""}
			} else {
				f.ID = []string{id}
			}
		}
	}
}

func CEI_MACAddrs(macAddrs []string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			if len(macAddrs) == 0 {
				f.MACAddr = []string{}
			} else {
				f.MACAddr = macAddrs
			}
		}
	}
}

func CEI_IPAddrs(ipAddrs []string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			if len(ipAddrs) == 0 {
				f.IPAddr = []string{}
			} else {
				f.IPAddr = ipAddrs
			}
		}
	}
}

func CEI_Networks(networks []string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			if len(networks) == 0 {
				f.Network = []string{}
			} else {
				f.Network = networks
			}
		}
	}
}

// Filter should include entries that occur after this time.
func CEI_NewerThan(newerThan string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			f.NewerThan = newerThan
		}
	}
}

// Filter should include entries that occur before this time.
func CEI_OlderThan(olderThan string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			f.OlderThan = olderThan
		}
	}
}

func CEI_CompIDs(ids []string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			if len(ids) == 0 {
				f.CompID = []string{""}
			} else {
				f.CompID = ids
			}
		}
	}
}

func CEI_CompTypes(compTypes []string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			if len(compTypes) == 0 {
				f.CompType = []string{}
			} else {
				f.CompType = compTypes
			}
		}
	}
}

// Set label field so any errors during the query can be attributed
// to the calling func
func CEI_From(callingFunc string) CompEthInterfaceFiltFunc {
	return func(f *CompEthInterfaceFilter) {
		if f != nil {
			f.label = callingFunc
		}
	}
}
