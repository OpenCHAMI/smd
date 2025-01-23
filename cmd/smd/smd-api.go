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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	base "github.com/Cray-HPE/hms-base"
	compcreds "github.com/Cray-HPE/hms-compcredentials"
	"github.com/Cray-HPE/hms-smd/v2/internal/hmsds"
	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
	"github.com/go-chi/chi/v5"
	"github.com/openchami/schemas/schemas"
	redfish "github.com/openchami/schemas/schemas/csm"
)

type componentArrayIn struct {
	Components   []base.Component `json:"Components"`
	ExtendedInfo json.RawMessage  `json:"ExtendedInfo,omitempty"`
}

type CompQueryIn struct {
	ComponentIDs []string `json:"ComponentIDs"`
}

type NIDQueryIn struct {
	NIDRanges []string `json:"NIDRanges"`
}

type FieldFltrIn struct {
	StateOnly bool `json:"stateonly"`
	FlagOnly  bool `json:"flagonly"`
	RoleOnly  bool `json:"roleonly"`
	NIDOnly   bool `json:"nidonly"`
}

type FieldFltrInForm struct {
	StateOnly []string `json:"stateonly"`
	FlagOnly  []string `json:"flagonly"`
	RoleOnly  []string `json:"roleonly"`
	NIDOnly   []string `json:"nidonly"`
}

type HwInvIn struct {
	Hardware []sm.HWInvByLoc `json:"Hardware"`
}

type HwInvQueryIn struct {
	ID           []string `json:"id"`
	Type         []string `json:"type"`
	Manufacturer []string `json:"manufacturer"`
	PartNumber   []string `json:"partnumber"`
	SerialNumber []string `json:"serialnumber"`
	FruId        []string `json:"fruid"`
	Children     []string `json:"children"`
	Parents      []string `json:"parents"`
	Partition    []string `json:"partition"`
	Format       []string `json:"format"`
}

type HwInvHistIn struct {
	ID        []string `json:"id"`
	FruId     []string `json:"fruid"`
	EventType []string `json:"eventtype"`
	StartTime []string `json:"starttime"`
	EndTime   []string `json:"endtime"`
}

type GrpPartFltr struct {
	Group     []string `json:"group"`
	Tag       []string `json:"tag"`
	Partition []string `json:"partition"`
}

type CompLockFltr struct {
	ID    []string `json:"id"`
	Owner []string `json:"owner"`
	Xname []string `json:"xname"`
}

type CompGetLockFltr struct {
	Type                []string `json:"Type"`
	State               []string `json:"State"`
	Role                []string `json:"Role"`
	SubRole             []string `json:"Subrole"`
	Locked              []string `json:"Locked"`
	Reserved            []string `json:"Reserved"`
	ReservationDisabled []string `json:"ReservationDisabled"`
}

type CompEthInterfaceFltr struct {
	ID        []string `json:"id"`
	MACAddr   []string `json:"macaddress"`
	IPAddr    []string `json:"ipaddress"`
	Network   []string `json:"network"`
	OlderThan []string `json:"olderthan"`
	NewerThan []string `json:"newerthan"`
	CompID    []string `json:"componentid"`
	Type      []string `json:"type"`
}

type HMSValueSelect int

const (
	HMSValAll HMSValueSelect = iota
	HMSValArch
	HMSValClass
	HMSValFlag
	HMSValNetType
	HMSValRole
	HMSValSubRole
	HMSValState
	HMSValType
)

type HMSValues struct {
	Arch    []string `json:"Arch,omitempty"`
	Class   []string `json:"Class,omitempty"`
	Flag    []string `json:"Flag,omitempty"`
	NetType []string `json:"NetType,omitempty"`
	Role    []string `json:"Role,omitempty"`
	SubRole []string `json:"SubRole,omitempty"`
	State   []string `json:"State,omitempty"`
	Type    []string `json:"Type,omitempty"`
}

/////////////////////////////////////////////////////////////////////////////
// Helper Fuctions
/////////////////////////////////////////////////////////////////////////////

// Translate form input into a FieldFilter
func getFieldFilterForm(f *FieldFltrInForm) hmsds.FieldFilter {
	/* Deal with the component field filters (i.e. "stateonly"). Due to the way that
	 * the query parameter parsing works, these values will be coming to us as strings.
	 * Convert them to bool. Take the first one that is true.
	 */
	if f == nil {
		return hmsds.FLTR_DEFAULT
	}
	if len(f.StateOnly) > 0 {
		compFltr, _ := strconv.ParseBool(f.StateOnly[0])
		if compFltr {
			return hmsds.FLTR_STATEONLY
		}
	}
	if len(f.FlagOnly) > 0 {
		compFltr, _ := strconv.ParseBool(f.FlagOnly[0])
		if compFltr {
			return hmsds.FLTR_FLAGONLY
		}
	}
	if len(f.RoleOnly) > 0 {
		compFltr, _ := strconv.ParseBool(f.RoleOnly[0])
		if compFltr {
			return hmsds.FLTR_ROLEONLY
		}
	}
	if len(f.NIDOnly) > 0 {
		compFltr, _ := strconv.ParseBool(f.NIDOnly[0])
		if compFltr {
			return hmsds.FLTR_NIDONLY
		}
	}
	return hmsds.FLTR_DEFAULT
}

// Translate POST input into a FieldFilter
func getFieldFilter(f *FieldFltrIn) hmsds.FieldFilter {
	/* Deal with the component field filters (i.e. "stateonly"). Due to the way that
	 * the query parameter parsing works, these values will be coming to us as strings.
	 * Convert them to bool. Take the first one that is true.
	 */
	if f == nil {
		return hmsds.FLTR_DEFAULT
	}
	if f.StateOnly {
		return hmsds.FLTR_STATEONLY
	}
	if f.FlagOnly {
		return hmsds.FLTR_FLAGONLY
	}
	if f.RoleOnly {
		return hmsds.FLTR_ROLEONLY
	}
	if f.NIDOnly {
		return hmsds.FLTR_NIDONLY
	}
	return hmsds.FLTR_DEFAULT
}

// Parse an array of NIDs and NID ranges into the NIDStart, NIDEnd, NID fields
// of a ComponentFilter. This function will prepend the parsed values to the
// NID, NIDStart, and NIDEnd arrays in the given ComponentFilter. This way
// pre-existing values in NIDStart and NIDEnd do not affect the reletive index
// of parsed NIDStart-NIDEnd pairs. A ComponentFilter will be created if one
// is not specified.
func nidRangeToCompFilter(nidRanges []string, f *hmsds.ComponentFilter) (*hmsds.ComponentFilter, error) {
	NIDStart := make([]string, 0, 1)
	NIDEnd := make([]string, 0, 1)
	NID := make([]string, 0, 1)

	// Create a ComponentFilter if one was not provided
	if f == nil {
		f = new(hmsds.ComponentFilter)
	}
	// Parse the NID ranges
	for _, nid := range nidRanges {
		nidRange := strings.Split(nid, "-")
		if len(nidRange) > 1 {
			// NID Range
			if len(nidRange[0]) == 0 || len(nidRange[len(nidRange)-1]) == 0 {
				return f, errors.New("argument was not a valid NID Range")
			}
			NIDStart = append(NIDStart, nidRange[0])
			NIDEnd = append(NIDEnd, nidRange[len(nidRange)-1])
		} else {
			// Single NID
			if len(nid) > 0 {
				NID = append(NID, nid)
			}
		}
	}
	// Append any values from the given ComponentFilter effectively prepending the parsed values.
	if len(f.NIDStart) > 0 {
		NIDStart = append(NIDStart, f.NIDStart...)
	}
	if len(f.NIDEnd) > 0 {
		NIDEnd = append(NIDEnd, f.NIDEnd...)
	}
	if len(f.NID) > 0 {
		NID = append(NID, f.NID...)
	}
	f.NIDStart = NIDStart
	f.NIDEnd = NIDEnd
	f.NID = NID
	return f, nil
}

func compGetLockFltrToCompLockV2Filter(cglf CompGetLockFltr) (clf sm.CompLockV2Filter) {
	clf.Type = cglf.Type
	clf.State = cglf.State
	clf.Role = cglf.Role
	clf.SubRole = cglf.SubRole
	clf.Locked = cglf.Locked
	clf.Reserved = cglf.Reserved
	clf.ReservationDisabled = cglf.ReservationDisabled
	return clf
}

/////////////////////////////////////////////////////////////////////////////
// HSM Service Info
/////////////////////////////////////////////////////////////////////////////

// Get the readiness state of HSM
func (s *SmD) doReadyGet(w http.ResponseWriter, r *http.Request) {
	// If we got here then the initial database connection was successful
	// Check that the DB connection is still available
	err := s.db.TestConnection()
	if err != nil {
		s.LogAlways("doReadyGet(): Database failed health check: %s", err)
		sendJsonError(w, http.StatusServiceUnavailable, "HSM's database is unhealthy: "+err.Error())
		return
	}
	// Tell them we are up and healthy
	sendJsonError(w, http.StatusOK, "HSM is healthy")
}

// Get the liveness state of HSM
func (s *SmD) doLivenessGet(w http.ResponseWriter, r *http.Request) {
	// Let the caller know we are accepting HTTP requests.
	w.WriteHeader(http.StatusNoContent)
}

// Get all HMS base enum values
func (s *SmD) doValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValAll, w, r)
}

// Get HMS base enum values for arch
func (s *SmD) doArchValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValArch, w, r)
}

// Get HMS base enum values for class
func (s *SmD) doClassValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValClass, w, r)
}

// Get HMS base enum values for flag
func (s *SmD) doFlagValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValFlag, w, r)
}

// Get HMS base enum values for nettype
func (s *SmD) doNetTypeValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValNetType, w, r)
}

// Get HMS base enum values for role
func (s *SmD) doRoleValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValRole, w, r)
}

// Get HMS base enum values for subrole
func (s *SmD) doSubRoleValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValSubRole, w, r)
}

// Get HMS base enum values for state
func (s *SmD) doStateValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValState, w, r)
}

// Get HMS base enum values for type
func (s *SmD) doTypeValuesGet(w http.ResponseWriter, r *http.Request) {
	s.getHMSValues(HMSValType, w, r)
}

func (s *SmD) getHMSValues(valSelect HMSValueSelect, w http.ResponseWriter, r *http.Request) {
	values := new(HMSValues)
	switch valSelect {
	case HMSValArch:
		values.Arch = base.GetHMSArchList()
	case HMSValClass:
		values.Class = base.GetHMSClassList()
	case HMSValFlag:
		values.Flag = base.GetHMSFlagList()
	case HMSValNetType:
		values.NetType = base.GetHMSNetTypeList()
	case HMSValRole:
		values.Role = base.GetHMSRoleList()
	case HMSValSubRole:
		values.SubRole = base.GetHMSSubRoleList()
	case HMSValState:
		values.State = base.GetHMSStateList()
	case HMSValType:
		values.Type = base.GetHMSTypeList()
	case HMSValAll:
		values.Arch = base.GetHMSArchList()
		values.Class = base.GetHMSClassList()
		values.Flag = base.GetHMSFlagList()
		values.NetType = base.GetHMSNetTypeList()
		values.Role = base.GetHMSRoleList()
		values.SubRole = base.GetHMSSubRoleList()
		values.State = base.GetHMSStateList()
		values.Type = base.GetHMSTypeList()
	}
	sendJsonValueRsp(w, values)
}

/////////////////////////////////////////////////////////////////////////////
// Component Status
/////////////////////////////////////////////////////////////////////////////

// Get single HMS component by xname ID
func (s *SmD) doComponentGet(w http.ResponseWriter, r *http.Request) {
	// TODO: get route variables using chi instead of mux

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	cmp, err := s.db.GetComponentByID(xname)
	if err != nil {
		s.LogAlways("doComponentGet(): Lookup failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if cmp == nil {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	// Over all summary error code needs to be computed...
	sendJsonCompRsp(w, cmp)
}

// Delete single ComponentEndpoint, by its xname ID.
func (s *SmD) doComponentDelete(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doComponentDelete(): trying...")

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	if !base.IsHMSCompIDValid(xname) {
		sendJsonError(w, http.StatusBadRequest, "invalid xname")
		return
	}

	didDelete, err := s.db.DeleteComponentByID(xname)
	if err != nil {
		s.LogAlways("doComponentDelete(): delete failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Get all HMS Components as named array
func (s *SmD) doComponentsGet(w http.ResponseWriter, r *http.Request) {
	comps := new(base.ComponentArray)
	var err error

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doComponentsGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doComponentsGet(): Marshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	compFilter := new(hmsds.ComponentFilter)
	if err = json.Unmarshal(formJSON, compFilter); err != nil {
		s.lg.Printf("doComponentsGet(): Unmarshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	// Get the component field filter options (i.e. stateonly)
	fieldFltrIn := new(FieldFltrInForm)
	if err = json.Unmarshal(formJSON, fieldFltrIn); err != nil {
		s.lg.Printf("doComponentsGet(): Unmarshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	fieldFltr := getFieldFilterForm(fieldFltrIn)
	comps.Components, err = s.db.GetComponentsFilter(compFilter, fieldFltr)
	if err != nil {
		s.LogAlways("doComponentsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonCompArrayRsp(w, comps)
}

// CREATE/Update components. If the component already exists it will not be
// overwritten unless force=true in which case State, Flag, Subtype, NetType,
// Arch, and Class will get overwritten.
func (s *SmD) doComponentsPost(w http.ResponseWriter, r *http.Request) {
	var err error
	var compsIn sm.ComponentsPost

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &compsIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if len(compsIn.Components) < 1 || len(compsIn.Components[0].ID) == 0 {
		sendJsonError(w, http.StatusBadRequest, "Missing Components")
		return
	}
	err = compsIn.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doComponentsPost(): Couldn't validate components: %s", err)
		sendJsonError(w, http.StatusBadRequest,
			"couldn't validate components: "+err.Error())
		return
	}
	// Get the nid and role defaults for all node types
	for _, comp := range compsIn.Components {
		if comp.Type == base.Node.String() {
			if len(comp.Role) == 0 || len(comp.NID) == 0 || len(comp.Class) == 0 {
				newNID, defRole, defSubRole, defClass := s.GetCompDefaults(comp.ID, base.RoleCompute.String(), "", "")
				if len(comp.Role) == 0 {
					comp.Role = defRole
				}
				if len(comp.SubRole) == 0 {
					comp.SubRole = defSubRole
				}
				if len(comp.NID) == 0 {
					comp.NID = json.Number(strconv.FormatUint(newNID, 10))
				}
				if len(comp.Class) == 0 {
					comp.Class = defClass
				}
			}
		}
	}
	changeMap, err := s.db.UpsertComponents(compsIn.Components, compsIn.Force)
	if err != nil {
		sendJsonDBError(w, "operation 'Post Components' failed: ", "", err)
		s.LogAlways("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		return
	}

	scnIds := make(map[string]map[string][]string, 0)
	// Group component ids by change type and new value for generating SCNs
	for _, comp := range compsIn.Components {
		changes, ok := changeMap[comp.ID]
		if !ok {
			continue
		}
		for change, value := range changes {
			// Skip if the type of change didn't happen.
			if !value {
				continue
			}
			switch change {
			case "state":
				if _, ok := scnIds[change]; !ok {
					scnIds[change] = make(map[string][]string, 0)
				}
				scnIds[change][comp.State] = append(scnIds[change][comp.State], comp.ID)
			case "enabled":
				if _, ok := scnIds[change]; !ok {
					scnIds[change] = make(map[string][]string, 0)
				}
				if comp.Enabled == nil {
					enabled := true
					comp.Enabled = &enabled
				}
				str := strconv.FormatBool(*comp.Enabled)
				scnIds[change][str] = append(scnIds[change][str], comp.ID)
			case "swStatus":
				if _, ok := scnIds[change]; !ok {
					scnIds[change] = make(map[string][]string, 0)
				}
				scnIds[change][comp.SwStatus] = append(scnIds[change][comp.SwStatus], comp.ID)
			case "role":
				if _, ok := scnIds[change]; !ok {
					scnIds[change] = make(map[string][]string, 0)
				}
				changeVal := comp.Role + "." + comp.SubRole
				scnIds[change][changeVal] = append(scnIds[change][changeVal], comp.ID)
			}
		}
	}
	// Send out a SCN for each unique combination of change type and new value
	for change, valMap := range scnIds {
		for val, list := range valMap {
			switch change {
			case "state":
				scn := NewJobSCN(list, base.Component{State: val}, s)
				s.wp.Queue(scn)
			case "enabled":
				enabled, _ := strconv.ParseBool(val)
				scn := NewJobSCN(list, base.Component{Enabled: &enabled}, s)
				s.wp.Queue(scn)
			case "swStatus":
				scn := NewJobSCN(list, base.Component{SwStatus: val}, s)
				s.wp.Queue(scn)
			case "role":
				roles := strings.Split(val, ".")
				scn := NewJobSCN(list, base.Component{Role: roles[0], SubRole: roles[1]}, s)
				s.wp.Queue(scn)
			}
		}
	}

	// Send 204 status (success, no content in response)
	sendJsonError(w, http.StatusNoContent, "operation completed")
}

// Get all HMS Components under multiple parent components as named array
func (s *SmD) doComponentsQueryPost(w http.ResponseWriter, r *http.Request) {
	comps := new(base.ComponentArray)
	var err error

	body, err := ioutil.ReadAll(r.Body)
	// Get the component list
	compQuery := new(CompQueryIn)
	err = json.Unmarshal(body, compQuery)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if len(compQuery.ComponentIDs) < 1 {
		sendJsonError(w, http.StatusBadRequest, "Missing IDs")
		return
	}
	// Get the query parameters
	compFilter := new(hmsds.ComponentFilter)
	err = json.Unmarshal(body, compFilter)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	// Get the component field filter options (i.e. stateonly)
	fieldFltrIn := new(FieldFltrIn)
	err = json.Unmarshal(body, fieldFltrIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	fieldFltr := getFieldFilter(fieldFltrIn)
	comps.Components, err = s.db.GetComponentsQuery(compFilter, fieldFltr, compQuery.ComponentIDs)
	if err != nil {
		s.LogAlways("doComponentsQueryPost(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonCompArrayRsp(w, comps)
}

// Get all HMS Components under a single parent component as named array
func (s *SmD) doComponentsQueryGet(w http.ResponseWriter, r *http.Request) {
	comps := new(base.ComponentArray)
	ids := make([]string, 0, 1)
	var err error

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doComponentsQueryGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doComponentsQueryGet(): Marshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	// Get the query parameters
	compFilter := new(hmsds.ComponentFilter)
	if err = json.Unmarshal(formJSON, compFilter); err != nil {
		s.lg.Printf("doComponentsQueryGet(): Unmarshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	// Get the component field filter options (i.e. stateonly)
	fieldFltrIn := new(FieldFltrInForm)
	if err = json.Unmarshal(formJSON, fieldFltrIn); err != nil {
		s.lg.Printf("doComponentsQueryGet(): Unmarshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	fieldFltr := getFieldFilterForm(fieldFltrIn)
	ids = append(ids, xname)
	comps.Components, err = s.db.GetComponentsQuery(compFilter, fieldFltr, ids)
	if err != nil {
		s.LogAlways("doComponentsQueryGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonCompArrayRsp(w, comps)
}

// Delete entire collection of ComponentEndpoints, undoing discovery.
func (s *SmD) doComponentsDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeleteComponentsAll()
	if err != nil {
		s.lg.Printf("doCompEndpointsDelete(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

// Get single HMS component by NID, if it exists and is a type that has a
// NID (i.e. a node)
func (s *SmD) doComponentByNIDGet(w http.ResponseWriter, r *http.Request) {

	xname := chi.URLParam(r, "nid")

	cmp, err := s.db.GetComponentByNID(xname)
	if err != nil {
		s.LogAlways("doStateComponent(): Lookup failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if cmp == nil {
		sendJsonError(w, http.StatusNotFound, "no such NID.")
		return
	}
	// Over all summary error code needs to be computed...
	sendJsonCompRsp(w, cmp)
}

// Get an array of HMS component by NID, if it exists and is a type that has a
// NID (i.e. a node)
func (s *SmD) doComponentByNIDQueryPost(w http.ResponseWriter, r *http.Request) {
	comps := new(base.ComponentArray)
	var err error

	body, err := ioutil.ReadAll(r.Body)
	// Get the component list
	nidQuery := new(NIDQueryIn)
	err = json.Unmarshal(body, nidQuery)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if len(nidQuery.NIDRanges) < 1 {
		sendJsonError(w, http.StatusBadRequest, "Missing NID ranges")
		return
	}
	// Get the query parameters. This is mostly
	// just to pick up a partition if specified.
	compFilter := new(hmsds.ComponentFilter)
	err = json.Unmarshal(body, compFilter)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	// Get the component field filter options (i.e. stateonly)
	fieldFltrIn := new(FieldFltrIn)
	err = json.Unmarshal(body, fieldFltrIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	fieldFltr := getFieldFilter(fieldFltrIn)
	// Reset the NID query arrays in the component filter just in case something weird happened
	compFilter.NID = compFilter.NID[:0]
	compFilter.NIDStart = compFilter.NIDStart[:0]
	compFilter.NIDEnd = compFilter.NIDEnd[:0]
	// Parse the NID ranges and add them to the compFilter
	compFilter, err = nidRangeToCompFilter(nidQuery.NIDRanges, compFilter)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, "bad query param: "+err.Error())
		return
	}
	comps.Components, err = s.db.GetComponentsFilter(compFilter, fieldFltr)
	if err != nil {
		s.LogAlways("doComponentsQueryGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonCompArrayRsp(w, comps)
}

// Bulk NID patch.  Unlike other patch methods, there needs to be a
// NID for each ID given, so the handling has to be a little different.
func (s *SmD) doCompBulkNIDPatch(w http.ResponseWriter, r *http.Request) {
	var err error

	var compsIn componentArrayIn
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &compsIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	components := &compsIn.Components
	if len(*components) < 1 || len((*components)[0].ID) < 1 {
		sendJsonError(w, http.StatusBadRequest, "Missing Components")
		return
	}
	err = s.db.BulkUpdateCompNID(components)
	if err != nil {
		sendJsonDBError(w, "operation 'Bulk Update NID' failed: ",
			"", err)
		s.LogAlways("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		return
	}
	s.lg.Printf("succeeded: %s %s", r.RemoteAddr, string(body))

	// Send 204 status (success, no content in response)
	sendJsonError(w, http.StatusNoContent, "")
}

// Update component state and flag for a list of components
func (s *SmD) doCompBulkStateDataPatch(w http.ResponseWriter, r *http.Request) {
	s.compBulkPatch(w, r, StateDataUpdate, "doCompBulkStateDataPatch")
}

// Update component state and flag for a list of components
func (s *SmD) doCompBulkFlagOnlyPatch(w http.ResponseWriter, r *http.Request) {
	s.compBulkPatch(w, r, FlagOnlyUpdate, "doCompBulkFlagOnlyPatch")

}

// Update component 'Enabled' boolean for a list of components
func (s *SmD) doCompBulkEnabledPatch(w http.ResponseWriter, r *http.Request) {
	s.compBulkPatch(w, r, EnabledUpdate, "doCompBulkEnabledPatch")

}

// Update component SoftwareStatus field for a list of components
func (s *SmD) doCompBulkSwStatusPatch(w http.ResponseWriter, r *http.Request) {
	s.compBulkPatch(w, r, SwStatusUpdate, "doCompBulkSwStatusPatch")

}

// Update component state and flag for a list of components
func (s *SmD) doCompBulkRolePatch(w http.ResponseWriter, r *http.Request) {
	s.compBulkPatch(w, r, RoleUpdate, "doCompBulkRolePatch")

}

// Helper function for doing a bulk patch via http.  CompUpdateInvalid
// is equivalent to no default.  We don't really want a default unless
// it's for backwards compatibility, or the API specifies a specific
// operation type.
func (s *SmD) compBulkPatch(
	w http.ResponseWriter,
	r *http.Request,
	t CompUpdateType,
	name string,
) {
	var err error
	body, err := ioutil.ReadAll(r.Body)
	bulkUpdate := new(CompUpdate)
	err = json.Unmarshal(body, bulkUpdate)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	s.compPatchHelper(w, r, t, name, bulkUpdate, true, body)
}

// Backend for all patch operations
func (s *SmD) compPatchHelper(
	w http.ResponseWriter,
	r *http.Request,
	t CompUpdateType,
	name string,
	update *CompUpdate,
	isBulk bool,
	body []byte,
) {
	if t != CompUpdateInvalid {
		update.UpdateType = t.String()
	} else {
		sendJsonError(w, http.StatusBadRequest, ErrSMDNoUpType.Error())
		return
	}

	//
	// Update Database
	//
	err := s.doCompUpdate(update, name)
	if err != nil {
		op := VerifyNormalizeCompUpdateType(update.UpdateType)
		if base.IsHMSError(err) {
			// HMS error, ok to send directly
			sendJsonError(w, http.StatusBadRequest, err.Error())
		} else {
			// Non-HMS error, print generic error message.
			if isBulk {
				// print generic message for bulk error
				sendJsonError(w, http.StatusBadRequest, "operation 'Bulk "+
					op+" Update' failed")
			} else {
				// Print generic error for non-bulk operations.
				id := "nil"
				if len(update.ComponentIDs) > 0 {
					id = update.ComponentIDs[0]
				}
				sendJsonError(w, http.StatusBadRequest,
					"operation '"+op+"' failed for "+id)
			}
		}
		// Either way we log the real error, we just don't want to leak
		// internals via user reported errors.
		s.Log(LOG_INFO, "%s(%s) failed: %s %s, Err: %s",
			name, op, r.RemoteAddr, string(body), err)
		return
	} else {
		s.Log(LOG_DEBUG, "%s() succeeded: %s %s",
			name, r.RemoteAddr, string(body))
	}
	// Send 204 status (success, no content in response)
	sendJsonError(w, http.StatusNoContent, "")

}

// Patch the State and Flag field (latter defaults to OK) for a single
// component.
func (s *SmD) doCompStateDataPatch(w http.ResponseWriter, r *http.Request) {
	s.componentPatch(w, r, StateDataUpdate, "doCompStateDataPatch")

}

// Patch the Flag field only (state does not change) for a single component.
func (s *SmD) doCompFlagOnlyPatch(w http.ResponseWriter, r *http.Request) {
	s.componentPatch(w, r, FlagOnlyUpdate, "doCompFlagOnlyPatch")

}

// Patch the Enabled boolean for a single component, leaving other fields
// in place.
func (s *SmD) doCompEnabledPatch(w http.ResponseWriter, r *http.Request) {
	s.componentPatch(w, r, EnabledUpdate, "doCompEnabledPatch")

}

// Patch the SoftwareStatus field for a single component, leaving other
// fields in place.
func (s *SmD) doCompSwStatusPatch(w http.ResponseWriter, r *http.Request) {
	s.componentPatch(w, r, SwStatusUpdate, "doCompSwStatusPatch")

}

// Patch the Role field for a single component, leaving other fields in place.
func (s *SmD) doCompRolePatch(w http.ResponseWriter, r *http.Request) {
	s.componentPatch(w, r, RoleUpdate, "doCompRolePatch")

}

// Update the NID (Node ID) for a single component, leaving other fields
// in place.
func (s *SmD) doCompNIDPatch(w http.ResponseWriter, r *http.Request) {
	s.componentPatch(w, r, SingleNIDUpdate, "doCompNIDPatch")

}

// Scan function to keep compatibility with API, though we don't really
// need the ID field as it is filled in by the URL for single component
// patch operations.
type compPatchIn struct {
	ID         string `json:"ID"`
	CompUpdate        // embedded struct
}

// Helper function to swap the state-change API in HTTP for single
// component updates.
func (s *SmD) componentPatch(
	w http.ResponseWriter,
	r *http.Request, t CompUpdateType,
	name string,
) {

	xname := chi.URLParam(r, "xname")

	var update compPatchIn
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &update)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		s.Log(LOG_INFO, "%s() Got json error: '%s'", name, err)
		return
	}
	// Get the ID from the path.  It should never be empty.
	if update.ID == "" {
		if xname == "" {
			sendJsonError(w, http.StatusBadRequest, ErrSMDNoID.Error())
			return
		} else {
			update.ComponentIDs = []string{xname}
		}
	} else if update.ID != xname {
		// We really do not care about the ID in the body, but for
		// compatibility with earlier versions, if it's there it should
		// not contradict the path ID.
		sendJsonError(w, http.StatusBadRequest, ErrSMDIDConf.Error())
		return
	} else {
		update.ComponentIDs = []string{update.ID}
	}
	// Back-end handling is the same as bulk updates from this point.
	s.compPatchHelper(w, r, t, name, &update.CompUpdate, false, body)
}

// CREATE/Update a component. Force = true causes full replacement of an
// already existing component. Otherwise, only NID and Role fields are updated.
// In any case, it should not be needed except to force changes to what should
// otherwise be write-only fields.
func (s *SmD) doComponentPut(w http.ResponseWriter, r *http.Request) {

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	var compIn sm.ComponentPut
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &compIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	component := &compIn.Component
	if component.ID == "" {
		if xname == "" {
			sendJsonError(w, http.StatusBadRequest, "Missing ID")
			return
		} else {
			component.ID = xname
		}
	} else if base.NormalizeHMSCompID(component.ID) != xname {
		sendJsonError(w, http.StatusBadRequest,
			"ID in URL and POST body do not match")
		return
	}
	err = compIn.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doComponentPut(): Couldn't validate component: %s", err)
		sendJsonError(w, http.StatusBadRequest,
			"couldn't validate component: "+err.Error())
		return
	}
	// Get the nid and role defaults for all node types
	if component.Type == base.Node.String() {
		if len(component.Role) == 0 || len(component.NID) == 0 || len(component.Class) == 0 {
			newNID, defRole, defSubRole, defClass := s.GetCompDefaults(component.ID, base.RoleCompute.String(), "", "")
			if len(component.Role) == 0 {
				component.Role = defRole
			}
			if len(component.SubRole) == 0 {
				component.SubRole = defSubRole
			}
			if len(component.NID) == 0 {
				component.NID = json.Number(strconv.FormatUint(newNID, 10))
			}
			if len(component.Class) == 0 {
				component.Class = defClass
			}
		}
	}
	changeMap, err := s.db.UpsertComponents([]*base.Component{component}, compIn.Force)
	if err != nil {
		sendJsonDBError(w, "operation 'PUT' failed: ", "", err)
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		return
	}
	if changes, ok := changeMap[component.ID]; ok {
		scnIds := make([]string, 0, 1)
		scnIds = append(scnIds, component.ID)
		for change, value := range changes {
			if !value {
				continue
			}
			switch change {
			case "state":
				scn := NewJobSCN(scnIds, base.Component{State: component.State}, s)
				s.wp.Queue(scn)
			case "enabled":
				scn := NewJobSCN(scnIds, base.Component{Enabled: component.Enabled}, s)
				s.wp.Queue(scn)
			case "swStatus":
				scn := NewJobSCN(scnIds, base.Component{SwStatus: component.SwStatus}, s)
				s.wp.Queue(scn)
			case "role":
				scn := NewJobSCN(scnIds, base.Component{Role: component.Role, SubRole: component.SubRole}, s)
				s.wp.Queue(scn)
			}
		}
	}

	// Send 204 status (success, no content in response)
	sendJsonError(w, http.StatusNoContent, "operation completed")

}

/////////////////////////////////////////////////////////////////////////////
// Node NID Mappings
/////////////////////////////////////////////////////////////////////////////

// Get one specific NodeMap entry, previously created, by its xname ID.
func (s *SmD) doNodeMapGet(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doNodeMapGet(): trying...")

	xname := chi.URLParam(r, "xname")
	m, err := s.db.GetNodeMapByID(xname)
	if err != nil {
		s.LogAlways("doNodeMapGet(): Lookup failure: (%s) %s",
			xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if m == nil {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonNodeMapRsp(w, m)
}

// Get all NodeMap entries in database, by doing a GET against the
// entire collection.
func (s *SmD) doNodeMapsGet(w http.ResponseWriter, r *http.Request) {
	nnms := new(sm.NodeMapArray)
	var err error

	nnms.NodeMaps, err = s.db.GetNodeMapsAll()
	if err != nil {
		s.LogAlways("doNodeMapsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "", "", err)
		return
	}
	sendJsonNodeMapArrayRsp(w, nnms)
}

// Polymorphic type that takes either a single (scan-friendly) RedfishEndpoint
// or a named array of them.
type scanableNodeMap struct {
	*sm.NodeMap
	NodeMaps *[]sm.NodeMap `json:"NodeMaps"`
}

// CREATE new or UPDATE EXISTING Node->NID mapping
// Accept either a named array or a single entry.
func (s *SmD) doNodeMapsPost(w http.ResponseWriter, r *http.Request) {
	var scanMap scanableNodeMap
	nnms := new(sm.NodeMapArray)

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &scanMap)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if scanMap.NodeMap != nil {
		nnm, err := sm.NewNodeMap(
			scanMap.NodeMap.ID,
			scanMap.NodeMap.Role,
			scanMap.NodeMap.SubRole,
			scanMap.NodeMap.NID,
			scanMap.NodeMap.NodeInfo)
		if err != nil {
			sendJsonError(w, http.StatusBadRequest,
				"couldn't validate mapping data: "+err.Error())
			return
		}
		nnms.NodeMaps = append(nnms.NodeMaps, nnm)
	} else if scanMap.NodeMaps != nil {
		for i, m := range *scanMap.NodeMaps {
			// Attempt to create a valid NodeMap from the
			// raw data.  If we do not get any errors, it should be sane enough
			// to put into the data store.
			nnm, err := sm.NewNodeMap(
				m.ID,
				m.Role,
				m.SubRole,
				m.NID,
				m.NodeInfo)
			if err != nil {
				idx := strconv.Itoa(i)
				sendJsonError(w, http.StatusBadRequest,
					"couldn't validate map data at idx "+idx+": "+err.Error())
				return
			}
			nnms.NodeMaps = append(nnms.NodeMaps, nnm)
		}
	}
	err = s.db.InsertNodeMaps(nnms)
	if err != nil {
		s.lg.Printf("failed: %s %s Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing xname ID that has the same NID.")
		} else {
			sendJsonError(w, http.StatusInternalServerError,
				"operation 'POST' failed during store. ")
		}
		return
	}
	s.lg.Printf("succeeded: %s %s", r.RemoteAddr, string(body))

	numStr := strconv.FormatInt(int64(len(nnms.NodeMaps)), 10)
	sendJsonError(w, http.StatusOK, "Created or modified "+numStr+" entries")

}

// UPDATE EXISTING Node->NID mapping by it's xname URI.
func (s *SmD) doNodeMapPut(w http.ResponseWriter, r *http.Request) {

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	var m sm.NodeMap
	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &m)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if m.ID == "" {
		if xname != "" {
			m.ID = xname
		}
	} else if base.NormalizeHMSCompID(m.ID) != xname {
		sendJsonError(w, http.StatusBadRequest,
			"xname in URL and PUT body do not match")
		return
	}
	// Make sure the information submitted is a proper endpoint and will
	// not update the entry with invalid data.
	nnm, err := sm.NewNodeMap(
		m.ID,
		m.Role,
		m.SubRole,
		m.NID,
		m.NodeInfo)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest,
			"couldn't validate endpoint data: "+err.Error())
		return
	}
	err = s.db.InsertNodeMap(nnm)
	if err != nil {
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing resource that has the same NID")
		} else {
			// Unexpected error on update
			sendJsonError(w, http.StatusInternalServerError,
				"operation 'PUT' failed during store")
		}
		return
	}
	sendJsonNodeMapRsp(w, nnm)

}

// Delete single NodeMap, by its xname ID.
func (s *SmD) doNodeMapDelete(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doNodeMapDelete(): trying...")

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	if !base.IsHMSCompIDValid(xname) {
		sendJsonError(w, http.StatusBadRequest, "invalid xname")
		return
	}
	didDelete, err := s.db.DeleteNodeMapByID(xname)
	if err != nil {
		s.LogAlways("doNodeMapDelete(): delete failure: (%s) %s",
			xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Delete collection containing all NodeMap entries.
func (s *SmD) doNodeMapsDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeleteNodeMapsAll()
	if err != nil {
		s.lg.Printf("doNodeMapsDelete(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

/////////////////////////////////////////////////////////////////////////////
// Hardware Inventory
/////////////////////////////////////////////////////////////////////////////

// Get single HWInvByLocation entry by it's xname
func (s *SmD) doHWInvByLocationGet(w http.ResponseWriter, r *http.Request) {

	xname := chi.URLParam(r, "xname")
	hl, err := s.db.GetHWInvByLocID(xname)
	if err != nil {
		s.LogAlways("doHWInvByLocationGet(): Lookup failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if hl == nil {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonHWInvByLocRsp(w, hl)
}

// Get all HWInvByLocation entries in database, by doing a GET against the
// entire collection.
func (s *SmD) doHWInvByLocationGetAll(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doHWInvByLocationGetAll(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doHWInvByLocationGetAll(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	hwInvIn := new(HwInvQueryIn)
	if err = json.Unmarshal(formJSON, hwInvIn); err != nil {
		s.lg.Printf("doHWInvByLocationGetAll(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}

	hwInvLocFilter := []hmsds.HWInvLocFiltFunc{}

	if len(hwInvIn.ID) > 0 {
		for i, id := range hwInvIn.ID {
			normId := base.VerifyNormalizeCompID(id)
			if normId == "" {
				s.lg.Printf("doHWInvByLocationGetAll(): Invalid xname: %s", id)
				sendJsonError(w, http.StatusBadRequest, "Invalid xname")
				return
			}
			hwInvIn.ID[i] = normId
		}
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_IDs(hwInvIn.ID))
	}

	// Validate types
	if len(hwInvIn.Type) > 0 {
		for i, cType := range hwInvIn.Type {
			normType := base.VerifyNormalizeType(cType)
			if normType == "" {
				s.lg.Printf("doHWInvByLocationGetAll(): Invalid HMS type: %s", cType)
				sendJsonError(w, http.StatusBadRequest, "Invalid HMS type")
				return
			}
			hwInvIn.Type[i] = normType
		}
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Types(hwInvIn.Type))
	}

	// Manufacturer
	if len(hwInvIn.Manufacturer) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Manufacturers(hwInvIn.Manufacturer))
	}

	// Part Number
	if len(hwInvIn.PartNumber) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_PartNumbers(hwInvIn.PartNumber))
	}

	// Serial Number
	if len(hwInvIn.SerialNumber) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_SerialNumbers(hwInvIn.SerialNumber))
	}

	// FRU Id
	if len(hwInvIn.FruId) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_FruIDs(hwInvIn.FruId))
	}

	// Partition
	if len(hwInvIn.Partition) > 0 {
		for _, p := range hwInvIn.Partition {
			normP := sm.NormalizeGroupField(p)
			if sm.VerifyGroupField(normP) != nil {
				s.lg.Printf("doHWInvByLocationGetAll(): Invalid partition: %s", p)
				sendJsonError(w, http.StatusBadRequest, "Invalid partition")
				return
			}
			hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Part(normP))
		}
	}

	hwlocs, err := s.db.GetHWInvByLocFilter(hwInvLocFilter...)
	if err != nil {
		s.lg.Printf("doHWInvByLocationGetAll(): Lookup failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
		return
	}
	sendJsonHWInvByLocsRsp(w, hwlocs)
}

// Create/update HWInv entries.
func (s *SmD) doHWInvByLocationPost(w http.ResponseWriter, r *http.Request) {
	var hwIn HwInvIn

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &hwIn)
	if err != nil {
		s.lg.Printf("doHWInvByLocationPost(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	hwlocs, err := sm.NewHWInvByLocs(hwIn.Hardware)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = s.db.InsertHWInvByLocs(hwlocs)
	if err != nil {
		s.lg.Printf("doHWInvByLocationPost(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}
	s.GenerateHWInvHist(hwlocs)

	numStr := strconv.Itoa(len(hwlocs))
	sendJsonError(w, http.StatusOK, "Created "+numStr+" entries")

}

// Get single HWInvByFRU entry by its FRU ID
func (s *SmD) doHWInvByFRUGet(w http.ResponseWriter, r *http.Request) {

	fruID := chi.URLParam(r, "fruid")
	hf, err := s.db.GetHWInvByFRUID(fruID)
	if err != nil {
		s.LogAlways("doHWInvByFRUGet(): Lookup failure: (%s) %s", fruID, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if hf == nil {
		sendJsonError(w, http.StatusNotFound, "no such FRU ID.")
		return
	}
	sendJsonHWInvByFRURsp(w, hf)
}

// Get all HWInvByFRU entries in database, by doing a GET against the
// entire collection.
func (s *SmD) doHWInvByFRUGetAll(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doHWInvByFRUGetAll(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doHWInvByFRUGetAll(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	hwInvIn := new(HwInvQueryIn)
	if err = json.Unmarshal(formJSON, hwInvIn); err != nil {
		s.lg.Printf("doHWInvByFRUGetAll(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}

	hwInvLocFilter := []hmsds.HWInvLocFiltFunc{}

	// Validate types
	if len(hwInvIn.Type) > 0 {
		for i, cType := range hwInvIn.Type {
			normType := base.VerifyNormalizeType(cType)
			if normType == "" {
				s.lg.Printf("doHWInvByFRUGetAll(): Invalid HMS type: %s", cType)
				sendJsonError(w, http.StatusBadRequest, "Invalid HMS type")
				return
			}
			hwInvIn.Type[i] = normType
		}
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Types(hwInvIn.Type))
	}

	// Manufacturer
	if len(hwInvIn.Manufacturer) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Manufacturers(hwInvIn.Manufacturer))
	}

	// Part Number
	if len(hwInvIn.PartNumber) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_PartNumbers(hwInvIn.PartNumber))
	}

	// Serial Number
	if len(hwInvIn.SerialNumber) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_SerialNumbers(hwInvIn.SerialNumber))
	}

	// FRU Id
	if len(hwInvIn.FruId) > 0 {
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_FruIDs(hwInvIn.FruId))
	}

	hwfrus, err := s.db.GetHWInvByFRUFilter(hwInvLocFilter...)
	if err != nil {
		s.lg.Printf("doHWInvByFRUGetAll(): Lookup failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
		return
	}
	sendJsonHWInvByFRUsRsp(w, hwfrus)
}

// Provides a xthwinv-type collection of system components, sorted by type
// and optionally nested (fully, or node subcomponents only).
func (s *SmD) doHWInvByLocationQueryGet(w http.ResponseWriter, r *http.Request) {
	var (
		compType    base.HMSType
		parentQuery bool
	)

	xname := chi.URLParam(r, "xname")
	format := sm.HWInvFormatNestNodesOnly

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doHWInvByLocationQueryGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doHWInvByLocationQueryGet(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	hwInvIn := new(HwInvQueryIn)
	if err = json.Unmarshal(formJSON, hwInvIn); err != nil {
		s.lg.Printf("doHWInvByLocationQueryGet(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}

	hwInvLocFilter := []hmsds.HWInvLocFiltFunc{}

	// Treat blanks as s0
	if xname == "" {
		compType = base.System
	} else {
		compType = base.GetHMSType(xname)
	}

	// Validate xnames
	if compType == base.HMSTypeInvalid {
		s.lg.Printf("doHWInvByLocationQueryGet(): Invalid xname: %s", xname)
		sendJsonError(w, http.StatusBadRequest, "Invalid xname")
		return
	} else if compType == base.Partition {
		hwInvIn.Partition = append(hwInvIn.Partition, xname)
	} else if !(compType == base.System || compType == base.HMSTypeAll) {
		// Add anything other than s0 or "all". If it is s0 or "all" we
		// leave ID empty so it will cause the filter to query everything.
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_ID(base.NormalizeHMSCompID(xname)))
	}

	// Validate types
	if len(hwInvIn.Type) > 0 {
		for i, cType := range hwInvIn.Type {
			normType := base.VerifyNormalizeType(cType)
			if normType == "" {
				s.lg.Printf("doHWInvByLocationQueryGet(): Invalid HMS type: %s", cType)
				sendJsonError(w, http.StatusBadRequest, "Invalid HMS type")
				return
			}
			hwInvIn.Type[i] = normType
		}
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Types(hwInvIn.Type))
	}

	// Partition
	if len(hwInvIn.Partition) > 0 {
		for _, p := range hwInvIn.Partition {
			normP := sm.NormalizeGroupField(p)
			if sm.VerifyGroupField(normP) != nil {
				s.lg.Printf("doHWInvByLocationGetAll(): Invalid partition: %s", p)
				sendJsonError(w, http.StatusBadRequest, "Invalid partition")
				return
			}
			hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Part(normP))
		}
	}

	// Search for parents?
	if len(hwInvIn.Parents) > 0 {
		parents, err := strconv.ParseBool(hwInvIn.Parents[0])
		if err != nil {
			s.lg.Printf("doHWInvByLocationQueryGet(): Invalid string for parents: %s", hwInvIn.Parents[0])
			sendJsonError(w, http.StatusBadRequest, "Invalid boolean for parents")
			return
		}
		if parents {
			hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Parent)
			parentQuery = true
		}
	}

	// Search for children?
	if len(hwInvIn.Children) > 0 {
		children, err := strconv.ParseBool(hwInvIn.Children[0])
		if err != nil {
			s.lg.Printf("doHWInvByLocationQueryGet(): Invalid string for children: %s", hwInvIn.Children[0])
			sendJsonError(w, http.StatusBadRequest, "Invalid boolean for children")
			return
		}
		if children {
			hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Child)
		}
	} else if !parentQuery {
		// Get children by default if we aren't looking up parents
		hwInvLocFilter = append(hwInvLocFilter, hmsds.HWInvLoc_Child)
	}

	// Validate format
	if len(hwInvIn.Format) > 0 {
		switch strings.ToLower(hwInvIn.Format[0]) {
		case strings.ToLower(sm.HWInvFormatFullyFlat):
			format = sm.HWInvFormatFullyFlat
		case strings.ToLower(sm.HWInvFormatNestNodesOnly):
			format = sm.HWInvFormatNestNodesOnly
		case strings.ToLower(sm.HWInvFormatHierarchical):
			// Not implemented yet
			fallthrough
		default:
			s.lg.Printf("doHWInvByLocationQueryGet(): Invalid format: %s", hwInvIn.Format)
			sendJsonError(w, http.StatusBadRequest, "Invalid format")
			return
		}
	}

	// Do the query
	hwlocs, err := s.db.GetHWInvByLocQueryFilter(hwInvLocFilter...)
	if err != nil {
		s.LogAlways("doHWInvByLocationQueryGet(%s): Lookup failure: %s",
			xname, err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to query DB.")
		return
	}

	// Sort the results
	hwinv, err := sm.NewSystemHWInventory(hwlocs, xname, format)
	if err != nil {
		s.LogAlways("doHWInvByLocationQueryGet(%s): HWInv parse: %s",
			xname, err)
		if err != base.ErrHMSTypeInvalid &&
			err != base.ErrHMSTypeUnsupported {
			// Ignore bad types so we don't break the whole thing.
			// We should generate everything else.
			sendJsonError(w, http.StatusInternalServerError,
				"Couldn't format response.")
			return
		}
	}
	sendJsonSystemHWInvRsp(w, hwinv)
}

// Delete a single HWInvByLocation by its xname ID.
func (s *SmD) doHWInvByLocationDelete(w http.ResponseWriter, r *http.Request) {

	xname := chi.URLParam(r, "xname")
	didDelete, err := s.db.DeleteHWInvByLocID(xname)
	if err != nil {
		s.LogAlways("doHWInvByLocationDelete(): delete failure: (%s) %s",
			xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Delete collection containing all HWInvByLocation entries
func (s *SmD) doHWInvByLocationDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeleteHWInvByLocsAll()
	if err != nil {
		s.lg.Printf("doHWInvByLocationDeleteAll(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

// Delete a single HWInvByFRUD entry, by its FRU ID.
func (s *SmD) doHWInvByFRUDelete(w http.ResponseWriter, r *http.Request) {

	fruID := chi.URLParam(r, "fruid")
	didDelete, err := s.db.DeleteHWInvByFRUID(fruID)
	if err != nil {
		s.LogAlways("doHWInvByFRUDelete(): delete failure: (%s) %s",
			fruID, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such FRU ID.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Delete collection containing all HWInvByFRU entries
func (s *SmD) doHWInvByFRUDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeleteHWInvByFRUsAll()
	if err != nil {
		s.lg.Printf("doHWInvByFRUDeleteAll(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

/////////////////////////////////////////////////////////////////////////////
// Hardware Inventory History
/////////////////////////////////////////////////////////////////////////////

// Get all HWInvHist entries in the database for a single locational xname.
func (s *SmD) doHWInvHistByLocationGet(w http.ResponseWriter, r *http.Request) {
	s.hwInvHistGet(w, r, sm.HWInvHistFmtByLoc)
}

// Get all HWInvHist entries in the database for a single FRU ID.
func (s *SmD) doHWInvHistByFRUGet(w http.ResponseWriter, r *http.Request) {
	s.hwInvHistGet(w, r, sm.HWInvHistFmtByFRU)
}

// Get all HWInvHist entries in the database for a single component
// (locational xname or FRU ID).
func (s *SmD) hwInvHistGet(w http.ResponseWriter, r *http.Request, format sm.HWInvHistFmt) {
	var id string

	switch format {
	case sm.HWInvHistFmtByLoc:
		id = chi.URLParam(r, "xname")
	case sm.HWInvHistFmtByFRU:
		id = chi.URLParam(r, "fruid")
	default:
		// Shouldn't happen
		return
	}

	if err := r.ParseForm(); err != nil {
		s.lg.Printf("hwInvHistGet(%s): ParseForm: %s", id, err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("hwInvHistGet(%s): Marshal form: %s", id, err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	hwInvHistIn := new(HwInvHistIn)
	if err = json.Unmarshal(formJSON, hwInvHistIn); err != nil {
		s.lg.Printf("hwInvHistGet(%s): Unmarshal form: %s", id, err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}

	hwInvHistFilter := []hmsds.HWInvHistFiltFunc{}

	switch format {
	case sm.HWInvHistFmtByLoc:
		normId := base.VerifyNormalizeCompID(id)
		if normId == "" {
			s.lg.Printf("hwInvHistGet(%s): Invalid xname: %s", id, id)
			sendJsonError(w, http.StatusBadRequest, "Invalid xname")
			return
		}
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_ID(normId))
	case sm.HWInvHistFmtByFRU:
		if id == "" {
			s.lg.Printf("hwInvHistGet(%s): Invalid FRU ID: %s", id, id)
			sendJsonError(w, http.StatusBadRequest, "Invalid FRU ID")
			return
		}
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_FruIDs([]string{id}))
	default:
		// Shouldn't happen
		return
	}

	// Validate event types
	if len(hwInvHistIn.EventType) > 0 {
		for i, evType := range hwInvHistIn.EventType {
			normEvType := sm.VerifyNormalizeHWInvHistEventType(evType)
			if normEvType == "" {
				s.lg.Printf("hwInvHistGet(%s): Invalid HWInvHist event type: %s", id, evType)
				sendJsonError(w, http.StatusBadRequest, "Invalid HWInvHist event type")
				return
			}
			hwInvHistIn.EventType[i] = normEvType
		}
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_EventTypes(hwInvHistIn.EventType))
	}

	// Start Time
	if len(hwInvHistIn.StartTime) > 0 {
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_StartTime(hwInvHistIn.StartTime[0]))
	}

	// End Time
	if len(hwInvHistIn.EndTime) > 0 {
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_EndTime(hwInvHistIn.EndTime[0]))
	}

	hwhists, err := s.db.GetHWInvHistFilter(hwInvHistFilter...)
	if err != nil {
		s.lg.Printf("hwInvHistGet(%s)(): Lookup failure: %s", id, err)
		sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
		return
	}
	hwHistoryResp := sm.HWInvHistArray{
		ID:      id,
		History: hwhists,
	}
	sendJsonHWInvHistRsp(w, &hwHistoryResp)
}

// Get all HWInvHist entries in the database, by doing a GET against the
// entire collection. Sorted by location xname associated with the HWInvHist
// entries.
func (s *SmD) doHWInvHistByLocationGetAll(w http.ResponseWriter, r *http.Request) {
	s.hwInvHistGetAll(w, r, sm.HWInvHistFmtByLoc)
}

// Get all HWInvHist entries in the database, by doing a GET against the
// entire collection. Sorted by FRU ID associated with the HWInvHist entries.
func (s *SmD) doHWInvHistByFRUGetAll(w http.ResponseWriter, r *http.Request) {
	s.hwInvHistGetAll(w, r, sm.HWInvHistFmtByFRU)
}

// Get all HWInvHist entries in the database, by doing a GET against the
// entire collection. Sorted by location xname or FRU ID associated with the
// HWInvHist entries.
func (s *SmD) hwInvHistGetAll(w http.ResponseWriter, r *http.Request, format sm.HWInvHistFmt) {
	var fmtStr string
	switch format {
	case sm.HWInvHistFmtByLoc:
		fmtStr = "HWInvHistFmtByLoc"
	case sm.HWInvHistFmtByFRU:
		fmtStr = "HWInvHistFmtByLoc"
	default:
		// Shouldn't happen
		return
	}
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("hwInvHistGetAll(%s): ParseForm: %s", fmtStr, err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("hwInvHistGetAll(%s): Marshal form: %s", fmtStr, err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	hwInvHistIn := new(HwInvHistIn)
	if err = json.Unmarshal(formJSON, hwInvHistIn); err != nil {
		s.lg.Printf("hwInvHistGetAll(%s): Unmarshal form: %s", fmtStr, err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}

	hwInvHistFilter := []hmsds.HWInvHistFiltFunc{}

	if len(hwInvHistIn.ID) > 0 {
		for i, id := range hwInvHistIn.ID {
			normId := base.VerifyNormalizeCompID(id)
			if normId == "" {
				s.lg.Printf("hwInvHistGetAll(%s): Invalid xname: %s", fmtStr, id)
				sendJsonError(w, http.StatusBadRequest, "Invalid xname")
				return
			}
			hwInvHistIn.ID[i] = normId
		}
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_IDs(hwInvHistIn.ID))
	}

	// FRU Id
	if len(hwInvHistIn.FruId) > 0 {
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_FruIDs(hwInvHistIn.FruId))
	}

	// Validate event types
	if len(hwInvHistIn.EventType) > 0 {
		for i, evType := range hwInvHistIn.EventType {
			normEvType := sm.VerifyNormalizeHWInvHistEventType(evType)
			if normEvType == "" {
				s.lg.Printf("hwInvHistGetAll(%s): Invalid HWInvHist event type: %s", fmtStr, evType)
				sendJsonError(w, http.StatusBadRequest, "Invalid HWInvHist event type")
				return
			}
			hwInvHistIn.EventType[i] = normEvType
		}
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_EventTypes(hwInvHistIn.EventType))
	}

	// Start Time
	if len(hwInvHistIn.StartTime) > 0 {
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_StartTime(hwInvHistIn.StartTime[0]))
	}

	// End Time
	if len(hwInvHistIn.EndTime) > 0 {
		hwInvHistFilter = append(hwInvHistFilter, hmsds.HWInvHist_EndTime(hwInvHistIn.EndTime[0]))
	}

	hwhists, err := s.db.GetHWInvHistFilter(hwInvHistFilter...)
	if err != nil {
		s.lg.Printf("hwInvHistGetAll(%s): Lookup failure: %s", fmtStr, err)
		sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
		return
	}
	historyResp, err := sm.NewHWInvHistResp(hwhists, format)
	if err != nil {
		s.LogAlways("hwInvHistGetAll(%s): HWInvHist parse: %s", fmtStr, err)
		sendJsonError(w, http.StatusInternalServerError, "Couldn't format response.")
		return
	}
	sendJsonHWInvHistArrayRsp(w, historyResp)
}

// Delete the HWInvHist entries for a single HWInvByLocation by its xname ID.
func (s *SmD) doHWInvHistByLocationDelete(w http.ResponseWriter, r *http.Request) {

	xname := chi.URLParam(r, "xname")
	normId := base.VerifyNormalizeCompID(xname)
	if normId == "" {
		s.lg.Printf("doHWInvHistByLocationDelete(%s): Invalid xname: %s", xname, xname)
		sendJsonError(w, http.StatusBadRequest, "Invalid xname")
		return
	}
	numDeleted, err := s.db.DeleteHWInvHistByLocID(normId)
	if err != nil {
		s.LogAlways("doHWInvHistByLocationDelete(): delete failure: (%s) %s", normId, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

// Delete collection containing all HWInvHist entries
func (s *SmD) doHWInvHistDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeleteHWInvHistAll()
	if err != nil {
		s.LogAlways("doHWInvHistDeleteAll(): Delete failure: %s", err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

// Delete the HWInvHist entries for a single HWInvByFRUD entry, by its FRU ID.
func (s *SmD) doHWInvHistByFRUDelete(w http.ResponseWriter, r *http.Request) {

	fruID := chi.URLParam(r, "fruid")
	numDeleted, err := s.db.DeleteHWInvHistByFRUID(fruID)
	if err != nil {
		s.LogAlways("doHWInvByFRUDelete(): Delete failure: (%s) %s", fruID, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

/////////////////////////////////////////////////////////////////////////////
// Redfish endpoints
/////////////////////////////////////////////////////////////////////////////

// Get one specific RedfishEndpoint, previously created, by its xname ID.
func (s *SmD) doRedfishEndpointGet(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doRedfishEndpointGet(): trying...")

	xname := chi.URLParam(r, "xname")
	ep, err := s.db.GetRFEndpointByID(xname)
	if err != nil {
		s.LogAlways("doRedfishEndpointGet(): Lookup failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if ep == nil {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonRFEndpointRsp(w, ep)
}

// Get all RedfishEndpoint entries in database, by doing a GET against the
// entire collection.
func (s *SmD) doRedfishEndpointsGet(w http.ResponseWriter, r *http.Request) {
	eps := new(sm.RedfishEndpointArray)
	var err error

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doRedfishEndpointsGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doRedfishEndpointsGet(): Marshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	rfEPFilter := new(hmsds.RedfishEPFilter)
	if err = json.Unmarshal(formJSON, rfEPFilter); err != nil {
		s.lg.Printf("doRedfishEndpointsGet(): Unmarshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	eps.RedfishEndpoints, err = s.db.GetRFEndpointsFilter(rfEPFilter)
	if err != nil {
		s.LogAlways("doRedfishEndpointsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonRFEndpointArrayRsp(w, eps)
}

// We may not need this.  But need a post version for getting an arbitrary
// list of endpoints.  If we can do these queries via filters on the
// RedfishEndpoints collection then it is superflouous.  That would be
// more REST-like.
func (s *SmD) doRedfishEndpointQueryGet(w http.ResponseWriter, r *http.Request) {
	eps := new(sm.RedfishEndpointArray)

	xname := chi.URLParam(r, "xname")
	if xname == "" || xname == "all" || xname == "s0" {
		var err error
		eps.RedfishEndpoints, err = s.db.GetRFEndpointsAll()
		if err != nil {
			s.lg.Printf("doRedfishEndpointQueryGet(): Lookup failure: %s", err)
			sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
			return
		}
		sendJsonRFEndpointArrayRsp(w, eps)
		return
	} else {
		sendJsonError(w, http.StatusBadRequest, "not yet implemented")
		return
	}
}

// Delete single RedfishEndpoint, by its xname ID.  This also deletes any
// child ComponentEndpoints, though not other structures.
func (s *SmD) doRedfishEndpointDelete(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doRedfishEndpointDelete(): trying...")

	xname := chi.URLParam(r, "xname")
	didDelete, affectedIDs, err := s.db.DeleteRFEndpointByIDSetEmpty(xname)
	if err != nil {
		s.LogAlways("doRedfishEndpointDelete(): delete failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	if len(affectedIDs) != 0 {
		data := base.Component{
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}
		scn := NewJobSCN(affectedIDs, data, s)
		s.wp.Queue(scn)
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Delete collection containing all RedfishEndoint entries.
func (s *SmD) doRedfishEndpointsDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, affectedIDs, err := s.db.DeleteRFEndpointsAllSetEmpty()
	if err != nil {
		s.lg.Printf("doRedfishEndpointsDelete(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	if len(affectedIDs) != 0 {
		data := base.Component{
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}
		scn := NewJobSCN(affectedIDs, data, s)
		s.wp.Queue(scn)
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

// UPDATE existing RedfishEndpoint entry in full (or all least all
// user-writable portions) or CREATE if it does not exists.
func (s *SmD) doRedfishEndpointPut(w http.ResponseWriter, r *http.Request) {
	var (
		xname = base.NormalizeHMSCompID(chi.URLParam(r, "xname"))
		rep   rf.RawRedfishEP
		cred  compcreds.CompCredentials
		body  []byte
		err   error
	)

	body, err = io.ReadAll(r.Body)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error reading response body "+err.Error())
	}

	// We expect the RedfishEndpoint to be in JSON format in the request body.
	err = json.Unmarshal(body, &rep)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if rep.ID == "" {
		if xname != "" {
			rep.ID = xname
		}
	} else if base.NormalizeHMSCompID(rep.ID) != xname {
		sendJsonError(w, http.StatusBadRequest,
			"xname in URL and PUT body do not match")
		return
	}

	// Make sure the information submitted is a proper endpoint and will
	// not update the entry with invalid data.
	epd, err := rf.NewRedfishEPDescription(&rep)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest,
			"couldn't validate endpoint data: "+err.Error())
		return
	}

	// Package it up into a SM RedfishEndpoint representation and send to DB.
	ep := sm.NewRedfishEndpoint(epd)
	if s.writeVault {
		cred = compcreds.CompCredentials{
			Xname:    ep.ID,
			URL:      ep.FQDN + "/redfish/v1",
			Username: ep.User,
			Password: ep.Password,
		}
		if s.readVault {
			ep.Password = ""
		}
	}

	// TODO: Change behavior to create a new RedfishEndpoint entry
	retEP, affectedIDs, err := s.db.UpdateRFEndpointNoDiscInfo(ep)
	if err != nil {
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing resource that has the same FQDN")
		} else {
			// Unexpected error on update
			sendJsonError(w, http.StatusInternalServerError,
				"operation 'PUT' failed during store")
		}
		return
	} else if retEP == nil {
		// Create a new RedfishEndpoint in DB.
		err = s.db.InsertRFEndpoint(ep)
		if err != nil {
			s.lg.Printf("failed: %s Err: %s", r.RemoteAddr, err)
			if err == hmsds.ErrHMSDSDuplicateKey {
				sendJsonError(w, http.StatusConflict, "operation would conflict "+
					"with an existing resource that has the same FQDN or xname ID.")
			} else {
				sendJsonError(w, http.StatusInternalServerError,
					"operation 'POST' failed during store. ")
			}
			return
		}

		// check for the data format sent via the schema version
		var (
			schemaVersion = s.getSchemaVersion(w, body)
			eps           = &sm.RedfishEndpointArray{
				RedfishEndpoints: []*sm.RedfishEndpoint{ep},
			}
		)
		if schemaVersion <= 0 {
			// parse data and populate component endpoints before inserting into db
			err = s.parseRedfishEndpointData(w, eps, body)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					fmt.Sprintf("failed parsing post data: %v", err))
			}
		} else {
			// parse data using the new inventory data format (will conform to schema)
			err = s.parseRedfishEndpointDataV2(w, body, true)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					fmt.Sprintf("failed parsing post data (V2): %v", err))
			}
		}

		// TODO: This is too be removed since we don't want this behavior.
		//
		// No error, but no update: Resource was not found.
		// s.lg.Printf("doRedfishEndpointPut: No such entry %s", rep.ID)
		// sendJsonError(w, http.StatusNotFound, "No such entry: "+rep.ID)
	}

	// Store credentials that are given in vault
	if s.writeVault {
		// Don't store empty credentials
		if len(cred.Password) > 0 {
			err = s.ccs.StoreCompCred(cred)
			if err != nil {
				// Something else should happen here. Maybe try reverting the
				// redfish endpoint changes we just made in the database? If we
				// fail to store credentials in vault, we'll lose the credentials
				// and the redfish endpoints associated with them will still be
				// successfully in the database. I think this is ok for now
				// since the future plan is for HSM to only read credentials
				// from Vault. Other services like REDS should be writing the
				// credentials to Vault.
				s.lg.Printf("failed: %s Err: %s", r.RemoteAddr, err)
				sendJsonError(w, http.StatusInternalServerError,
					"operation 'PUT' failed during secure store")
				return
			}
		}
	}
	if len(affectedIDs) != 0 {
		data := base.Component{
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}
		scn := NewJobSCN(affectedIDs, data, s)
		s.wp.Queue(scn)
	}

	// Do discovery if needed on new Endpoints.  Should never want to
	// force this since it can cause both the new and old discovery to
	// fail.  A manual discovery would be the recovery mechanism.
	// TODO:  Add auto-force based on time delta.
	//
	// Discovery can optionally be disabled with the --disable-discovery
	// flag from the CLI.
	if !s.disableDiscovery {
		go s.discoverFromEndpoint(ep, 0, false)
	}

	//
	// Create RedfishEndpoints, Components, and ComponentEndpoints from
	// the "systems" and "managers" properties found in the request body
	// in JSON format.
	//

	s.lg.Printf("succeeded: %s %s", r.RemoteAddr, string(body))

	// Send 200 status (success
	sendJsonRFEndpointRsp(w, retEP)
}

// PATCH existing RedfishEndpoint entry but only the fields specified.
func (s *SmD) doRedfishEndpointPatch(w http.ResponseWriter, r *http.Request) {

	xname := base.VerifyNormalizeCompID(chi.URLParam(r, "xname"))

	var rep sm.RedfishEndpointPatch
	var epUser string
	var epPassword string

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error reading body "+err.Error())
		return
	}
	err = json.Unmarshal(body, &rep)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if xname == "" {
		sendJsonError(w, http.StatusBadRequest,
			"xname in URL is not valid")
		return
	}

	if s.writeVault {
		if rep.User != nil {
			epUser = *rep.User
		}
		if rep.Password != nil {
			epPassword = *rep.Password
		}
		if s.readVault {
			rep.Password = nil
		}
	}
	retEP, affectedIDs, err := s.db.PatchRFEndpointNoDiscInfo(xname, rep)
	if err != nil {
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing resource that has the same FQDN")
		} else {
			// Unexpected error on update
			sendJsonError(w, http.StatusInternalServerError,
				"operation 'PATCH' failed during store")
		}
		return
	} else if retEP == nil {
		// No error, but no update: Resource was not found.
		s.lg.Printf("doRedfishEndpointPatch: No such entry %s", xname)
		sendJsonError(w, http.StatusNotFound, "No such entry: "+xname)
		return
	}
	// Store credentials that are given in vault
	if s.writeVault {
		// Don't store empty credentials
		if len(epUser) > 0 && len(epPassword) > 0 {
			cred := compcreds.CompCredentials{
				Xname:    retEP.ID,
				URL:      retEP.FQDN + "/redfish/v1",
				Username: epUser,
				Password: epPassword,
			}
			err = s.ccs.StoreCompCred(cred)
			if err != nil {
				// Something else should happen here. Maybe try reverting the
				// redfish endpoint changes we just made in the database? If we
				// fail to store credentials in vault, we'll lose the credentials
				// and the redfish endpoints associated with them will still be
				// successfully in the database. I think this is ok for now
				// since the future plan is for HSM to only read credentials
				// from Vault. Other services like REDS should be writing the
				// credentials to Vault.
				s.lg.Printf("failed: %s Err: %s", r.RemoteAddr, err)
				sendJsonError(w, http.StatusInternalServerError,
					"operation 'PATCH' failed during secure store")
				return
			}
		}
	}
	if len(affectedIDs) != 0 {
		data := base.Component{
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}
		scn := NewJobSCN(affectedIDs, data, s)
		s.wp.Queue(scn)
	}
	// Do discovery if needed on new Endpoints.  Should never want to
	// force this since it can cause both the new and old discovery to
	// fail.  A manual discovery would be the recovery mechanism.
	// TODO:  Add auto-force based on time delta.
	go s.discoverFromEndpoint(retEP, 0, false)

	s.lg.Printf("succeeded: %s %s", r.RemoteAddr, string(body))

	// Send 200 status (success
	sendJsonRFEndpointRsp(w, retEP)

}

// Polymorphic type that takes either a single (scan-friendly) RedfishEndpoint
// or a named array of them.
type scanableRedfishEndpoint struct {
	*rf.RawRedfishEP
	RedfishEndpoints *[]rf.RawRedfishEP `json:"RedfishEndpoints"`
}

// CREATE new RedfishEndpoint or Endpoints if there is a named array provided
// instead of a single entry.
func (s *SmD) doRedfishEndpointsPost(w http.ResponseWriter, r *http.Request) {
	var scanEPs scanableRedfishEndpoint
	eps := new(sm.RedfishEndpointArray)
	creds := []compcreds.CompCredentials{}

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &scanEPs)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if scanEPs.RawRedfishEP != nil {
		epd, err := rf.NewRedfishEPDescription(scanEPs.RawRedfishEP)
		if err != nil {
			sendJsonError(w, http.StatusBadRequest,
				"couldn't validate endpoint data: "+err.Error())
			return
		}
		ep := sm.NewRedfishEndpoint(epd)
		if s.writeVault {
			cred := compcreds.CompCredentials{
				Xname:    ep.ID,
				URL:      ep.FQDN + "/redfish/v1",
				Username: ep.User,
				Password: ep.Password,
			}
			if s.readVault {
				ep.Password = ""
			}
			creds = append(creds, cred)
		}
		eps.RedfishEndpoints = append(eps.RedfishEndpoints, ep)
	} else if scanEPs.RedfishEndpoints != nil {
		for i, rep := range *scanEPs.RedfishEndpoints {
			// Attempt to create a valid RedfishEndpointDescription from the
			// raw data.  If we do not get any errors, it should be sane enough
			// to put into the data store.
			epd, err := rf.NewRedfishEPDescription(&rep)
			if err != nil {
				idx := strconv.Itoa(i)
				sendJsonError(w, http.StatusBadRequest,
					"couldn't validate endpoint data at idx "+idx+": "+err.Error())
				return
			}
			// Package it up into a SM RedfishEndpoint representation and send to DB.
			ep := sm.NewRedfishEndpoint(epd)
			if s.writeVault {
				cred := compcreds.CompCredentials{
					Xname:    ep.ID,
					URL:      ep.FQDN + "/redfish/v1",
					Username: ep.User,
					Password: ep.Password,
				}
				ep.Password = ""
				creds = append(creds, cred)
			}
			eps.RedfishEndpoints = append(eps.RedfishEndpoints, ep)
		}
	}

	err = s.db.InsertRFEndpoints(eps)
	if err != nil {
		s.lg.Printf("failed: %s Err: %s", r.RemoteAddr, err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing resource that has the same FQDN or xname ID.")
		} else {
			sendJsonError(w, http.StatusInternalServerError,
				"operation 'POST' failed during store. ")
		}
		return
	}
	// Store credentials that are given in vault
	if s.writeVault {
		for _, cred := range creds {
			// Don't store empty credentials
			if len(cred.Password) > 0 {
				err = s.ccs.StoreCompCred(cred)
				if err != nil {
					// Something else should happen here. Maybe remove the redfish
					// endpoints we just inserted into the database? If we fail to
					// store credentials in vault, we'll lose the credentials and
					// the redfish endpoints associated with them will still be
					// successfully in the database. I think this is ok for now
					// since the future plan is for HSM to only read credentials
					// from Vault. Other services like REDS should be writing the
					// credentials to Vault.
					s.lg.Printf("failed: %s Err: %s", r.RemoteAddr, err)
					sendJsonError(w, http.StatusInternalServerError,
						"operation 'POST' failed during secure store. ")
					return
				}
			}
		}
	}
	s.lg.Printf("succeeded: %s %s", r.RemoteAddr, string(body))

	// Do discovery if needed on new Endpoints.  Should never need to
	// force this because the endpoint should always be new, else we would
	// have already failed the operation.
	//
	// Discovery can optionally be disabled with the --disable-discovery
	// flag from the CLI.
	if !s.disableDiscovery {
		go s.discoverFromEndpoints(eps.RedfishEndpoints, 0, true, false)
	}

	//
	// Create RedfishEndpoints, Components, and ComponentEndpoints from
	// the "systems" and "managers" properties found in the request body
	// in JSON format.
	//

	if s.ochami {
		// check for the data format sent via the schema version
		schemaVersion := s.getSchemaVersion(w, body)
		if schemaVersion <= 0 {
			// parse data and populate component endpoints before inserting into db
			err = s.parseRedfishEndpointData(w, eps, body)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					fmt.Sprintf("failed parsing post data: %v", err))
			}
		} else {
			// parse data using the new inventory data format (will conform to schema)
			err = s.parseRedfishEndpointDataV2(w, body, false)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					fmt.Sprintf("failed parsing post data (V2): %v", err))
			}
		}
	}

	// Send a URI array of the created resources, along with 201 (created).
	uris := eps.GetResourceURIArray(s.redfishEPBaseV2)
	sendJsonNewResourceIDArray(w, s.redfishEPBaseV2, uris)
}

// Parse the incoming JSON data, extracts specific keys, and writes the data
// to the database
// Used for ochami where discovery is diabled
func (s *SmD) parseRedfishEndpointData(w http.ResponseWriter, eps *sm.RedfishEndpointArray, data []byte) error {
	s.lg.Printf("parsing request data using default parsing method...")
	var obj map[string]any
	err := json.Unmarshal(data, &obj)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to unmarshal data: %v", err))
	}

	// systems
	systems, foundSystems := obj["Systems"]
	if foundSystems && systems != nil {
		for _, system := range systems.([]any) {
			// component
			data, foundData := system.(map[string]any)["Data"]
			status, ok := data.(map[string]any)["Status"].(map[string]any)["State"]
			enabled := ok && status == "Enabled"
			if foundData {
				// get ethernet interface link status
				component := base.Component{
					ID: obj["ID"].(string),
					// State: "On",
					Type:    "Node",
					Enabled: &enabled,
				}
				_, err := s.db.InsertComponent(&component)
				if err != nil {
					sendJsonError(w, http.StatusInternalServerError,
						fmt.Sprintf("failed to insert component: %v", err))
					return err
				}
			}

			// component endpoints
			uuid, ok := data.(map[string]any)["UUID"]
			if !ok {
				uuid = ""
			}
			// get system status (specifically if it is enabled?)

			cep := sm.ComponentEndpoint{
				ComponentDescription: rf.ComponentDescription{
					ID:             obj["ID"].(string),
					Type:           "Node",
					RedfishType:    "ComputerSystem",
					RedfishSubtype: data.(map[string]any)["SystemType"].(string),
					UUID:           uuid.(string),
					RfEndpointID:   obj["ID"].(string),
				},
				RfEndpointFQDN:        "",
				URL:                   data.(map[string]any)["@odata.id"].(string),
				ComponentEndpointType: "ComponentEndpointComputerSystem",
				Enabled:               enabled,
				RedfishSystemInfo:     nil,
			}

			// add ethernet interfaces to component endpoint
			interfaces, foundInterfaces := system.(map[string]any)["EthernetInterfaces"]
			if foundInterfaces {
				nicInfo := []*rf.EthernetNICInfo{}
				for _, i := range interfaces.([]any) {
					in := i.(map[string]any)
					enabled := in["InterfaceEnabled"].(bool)
					nicInfo = append(nicInfo, &rf.EthernetNICInfo{
						InterfaceEnabled: &enabled,
						RedfishId:        in["Id"].(string),
						Oid:              in["@odata.id"].(string),
						Description:      in["Description"].(string),
						MACAddress:       strings.ToLower(in["MACAddress"].(string)),
					})
				}
				cep.RedfishSystemInfo = &rf.ComponentSystemInfo{
					Actions:    nil,
					EthNICInfo: nicInfo,
				}
			}

			// finally, insert component endpoint into DB
			err = s.db.UpsertCompEndpoint(&cep)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					fmt.Sprintf("failed to upsert component endpoint: %v", err))
				return err
			}
		}
	}
	return nil
}

// Used for ochami where discovery is diabled
func (s *SmD) parseRedfishEndpointDataV2(w http.ResponseWriter, data []byte, forceUpdate bool) error {
	s.lg.Printf("parsing request data using V2 parsing method...")

	// NOTE: temporary definition for manager
	type Manager struct {
		URI                string                      `json:"uri,omitempty"`
		UUID               string                      `json:"uuid,omitempty"`
		Name               string                      `json:"name,omitempty"`
		Description        string                      `json:"description,omitempty"`
		Model              string                      `json:"model,omitempty"`
		Type               string                      `json:"type,omitempty"`
		FirmwareVersion    string                      `json:"firmware_version,omitempty"`
		EthernetInterfaces []schemas.EthernetInterface `json:"ethernet_interfaces,omitempty"`
	}

	type Root struct {
		redfish.RedfishEndpoint
		Systems  []schemas.InventoryDetail
		Managers []Manager
	}
	var (
		root Root
		err  error
	)

	// unmarshal the root JSON object from data
	err = json.Unmarshal(data, &root)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to unmarshal Redfish data: %v", err))
		return fmt.Errorf("failed to unmarshal Redfish data: %v", err)
	}

	var addEthernetInterfacesToNICInfo = func(eths []schemas.EthernetInterface, enabled bool) []*rf.EthernetNICInfo {
		// append NIC info to component endpoint
		nicInfo := make([]*rf.EthernetNICInfo, len(eths))
		for i, eth := range eths {
			nicInfo[i] = &rf.EthernetNICInfo{
				InterfaceEnabled: &enabled,        // NOTE: get via RF "InterfaceEnabled"
				RedfishId:        eth.URI,         // NOTE: what should this value be from RF?
				Oid:              eth.URI,         // NOTE: what should this value be from RF?
				Description:      eth.Description, // NOTE: intentionally set explicitly since this is included in V1
				MACAddress:       eth.MAC,
			}
		}
		return nicInfo
	}

	// iterate over all of the managers to create NodeBMC components and component endpoints
	for _, manager := range root.Managers {
		var (
			enabled   = true
			component = base.Component{
				ID: root.ID,
				// State:   manager.PowerState,
				Type:    base.NodeBMC.String(),
				Enabled: &enabled,
			}
		)
		// components
		rowsAffected, err := s.db.InsertComponent(&component)
		if err != nil {
			sendJsonError(w, http.StatusInternalServerError,
				fmt.Sprintf("failed to insert %d component(s): %v", rowsAffected, err))
			if forceUpdate {
				// upsert here to keep allow returning error for duplicates when not forcing updates
				_, err := s.db.UpsertComponents([]*base.Component{&component}, false)
				if err != nil {
					return fmt.Errorf("failed to update component: %v", rowsAffected, err)
				}
			} else {
				return fmt.Errorf("failed to insert %d component(s): %v", rowsAffected, err)
			}
		}

		// create a new ethernet interface with reference to the component above
		for _, eth := range manager.EthernetInterfaces {
			// convert IP address from manager ethernet interface to IPAddressMapping
			ips := []sm.IPAddressMapping{sm.IPAddressMapping{IPAddr: eth.IP}}
			cei, err := sm.NewCompEthInterfaceV2(eth.Description, eth.MAC, component.ID, ips)
			if err != nil {
				sendJsonError(w, http.StatusBadRequest, err.Error())
				continue
			}
			err = s.db.InsertCompEthInterface(cei)
			if err != nil {
				if err == hmsds.ErrHMSDSDuplicateKey {
					sendJsonError(w, http.StatusConflict, "operation would conflict "+
						"with an existing component ethernet interface that has the same MAC address.")
				} else {
					// Send this message as 500 or 400 plus error message if it is
					// an HMSError and not, e.g. an internal DB error code.
					sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
				}
				if forceUpdate {
					// try deleting and reinserting the CompEthInterface
					rowAffected, err := s.db.DeleteCompEthInterfaceByID(cei.ID)
					if err != nil {

					}
					if rowAffected {
						err = s.db.InsertCompEthInterface(cei)
						if err == hmsds.ErrHMSDSDuplicateKey {
							sendJsonError(w, http.StatusConflict, "operation would conflict "+
								"with an existing component ethernet interface that has the same MAC address.")
						} else {
							// Send this message as 500 or 400 plus error message if it is
							// an HMSError and not, e.g. an internal DB error code.
							sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
						}
					}
				}
				continue
			}
		}
	}

	// iterate over all of the systems to create components and component endpoints
	knownCEs := make(map[string]string)
	ceNum := 0
	for _, system := range root.Systems {
		var nid json.Number
		nidJNum, err := json.Marshal(ceNum + 1)
		if err != nil {
			s.Log(LOG_NOTICE, "failed to marshal NID %d into json: %v", ceNum+1, err)
		} else {
			err = json.Unmarshal(nidJNum, &nid)
			if err != nil {
				s.Log(LOG_NOTICE, "failed to unmarshal NID %d into json.Number: %v", ceNum+1, err)
			}
		}
		// use map to store known component endpoints by UUID to avoid adding duplicates
		if _, gotten := knownCEs[system.UUID]; !gotten {
			var (
				enabled   = strings.ToLower(system.PowerState) == "on"
				component = base.Component{
					ID:      root.ID + fmt.Sprintf("n%d", ceNum),
					NID:     nid,
					State:   system.PowerState,
					Type:    base.Node.String(),
					Enabled: &enabled,
				}
				componentEndpoint = sm.ComponentEndpoint{
					ComponentDescription: rf.ComponentDescription{
						ID:             root.ID + fmt.Sprintf("n%d", ceNum),
						Type:           base.Node.String(),
						RedfishType:    rf.ComputerSystemType, // TODO: need to get the RF type
						RedfishSubtype: system.SystemType,     // TODO: need to get the RF subtype (SystemType)
						UUID:           system.UUID,           // TODO: need to get the UUID (UUID)
						RfEndpointID:   root.ID,
					},
					RfEndpointFQDN:        "",
					URL:                   system.URI,
					ComponentEndpointType: "ComponentEndpointComputerSystem",
					Enabled:               enabled,
					RedfishSystemInfo: &rf.ComponentSystemInfo{
						Actions:    nil,
						EthNICInfo: addEthernetInterfacesToNICInfo(system.EthernetInterfaces, enabled),
					},
				}
			)
			knownCEs[system.UUID] = componentEndpoint.ComponentDescription.ID
			ceNum++

			// components
			rowsAffected, err := s.db.InsertComponent(&component)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					fmt.Sprintf("failed to insert %d component(s): %v", rowsAffected, err))

				// upsert here to keep allow returning error for duplicates when not forcing updates
				_, err := s.db.UpsertComponents([]*base.Component{&component}, false)
				if err != nil {
					return fmt.Errorf("failed to update component: %v", rowsAffected, err)
				}
				return fmt.Errorf("failed to insert %d component(s): %v", rowsAffected, err)
			}

			// component endpoints
			err = s.db.UpsertCompEndpoint(&componentEndpoint)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					fmt.Sprintf("failed to upsert component endpoint: %v", err))
				return fmt.Errorf("failed to upsert component endpoint: %v", err)
			}
		}
	}

	return nil
}

// getSchemaVersion() tries to extract the schema version from the JSON data.
// Used for ochami where discovery is diabled
func (s *SmD) getSchemaVersion(w http.ResponseWriter, data []byte) int {
	var (
		schemaVersion int = 0 // default to 0
		root          map[string]any
		ok            bool
		err           error
	)

	// unmarshal JSON to root
	err = json.Unmarshal(data, &root)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			fmt.Sprintf("failed to unmarshal data: %v", err))
	}

	// try and extract schema version and set if valid
	version, ok := root["SchemaVersion"]
	if ok {
		schemaVersion = int(version.(float64))
	}
	return schemaVersion
}

/////////////////////////////////////////////////////////////////////////////
// Component endpoints
/////////////////////////////////////////////////////////////////////////////

// Retrieves a single ComponentEndpoint (discovered info from Redfish on a
// component underneath a RedfishEndpoint).
func (s *SmD) doComponentEndpointGet(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doComponentEndpointGet(): trying...")

	xname := chi.URLParam(r, "xname")
	cep, err := s.db.GetCompEndpointByID(xname)
	if err != nil {
		s.LogAlways("doComponentEndpointGet(): Lookup failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if cep == nil {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonCompEndpointRsp(w, cep)
}

// Get collection of all ComponentEndpoints
func (s *SmD) doComponentEndpointsGet(w http.ResponseWriter, r *http.Request) {
	ceps := new(sm.ComponentEndpointArray)
	var err error

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doComponentEndpointsGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doComponentEndpointsGet(): Marshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	compEPFilter := new(hmsds.CompEPFilter)
	if err = json.Unmarshal(formJSON, compEPFilter); err != nil {
		s.lg.Printf("doComponentEndpointsGet(): Unmarshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	ceps.ComponentEndpoints, err = s.db.GetCompEndpointsFilter(compEPFilter)
	if err != nil {
		s.LogAlways("doComponentEndpointsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonCompEndpointArrayRsp(w, ceps)
}

// Delete single ComponentEndpoint, by its xname ID.
func (s *SmD) doComponentEndpointDelete(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doComponentEndpointDelete(): trying...")

	xname := chi.URLParam(r, "xname")
	didDelete, affectedIDs, err := s.db.DeleteCompEndpointByIDSetEmpty(xname)
	if err != nil {
		s.lg.Printf("doComponentEndpointDelete(): delete failure: (%s) %s", xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	if len(affectedIDs) != 0 {
		data := base.Component{
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}
		scn := NewJobSCN(affectedIDs, data, s)
		s.wp.Queue(scn)
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Delete entire collection of ComponentEndpoints, undoing discovery.
func (s *SmD) doComponentEndpointsDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, affectedIDs, err := s.db.DeleteCompEndpointsAllSetEmpty()
	if err != nil {
		s.lg.Printf("doCompEndpointsDelete(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	if len(affectedIDs) != 0 {
		data := base.Component{
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}
		scn := NewJobSCN(affectedIDs, data, s)
		s.wp.Queue(scn)
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

/////////////////////////////////////////////////////////////////////////////
// Service endpoints
/////////////////////////////////////////////////////////////////////////////

// Retrieves a single ServiceEndpoint (discovered info from Redfish on a
// service underneath a RedfishEndpoint).
func (s *SmD) doServiceEndpointGet(w http.ResponseWriter, r *http.Request) {

	svc := chi.URLParam(r, "service")
	xname := chi.URLParam(r, "xname")
	sep, err := s.db.GetServiceEndpointByID(svc, xname)
	if err != nil {
		s.lg.Printf("doServiceEndpointGet(): Lookup failure: (%s,%s) %s", svc, xname, err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "", err)
		return
	}
	if sep == nil {
		sendJsonError(w, http.StatusNotFound, "no such service under redfish endpoint.")
		return
	}
	sendJsonServiceEndpointRsp(w, sep)
}

// Get collection of all ServiceEndpoints
func (s *SmD) doServiceEndpointsGetAll(w http.ResponseWriter, r *http.Request) {
	seps := new(sm.ServiceEndpointArray)
	var err error

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doServiceEndpointsGetAll(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doServiceEndpointsGetAll(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	serviceEPFilter := new(hmsds.ServiceEPFilter)
	if err = json.Unmarshal(formJSON, serviceEPFilter); err != nil {
		s.lg.Printf("doServiceEndpointsGetAll(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	seps.ServiceEndpoints, err = s.db.GetServiceEndpointsFilter(serviceEPFilter)
	if err != nil {
		s.lg.Printf("doServiceEndpointsGetAll(): Lookup failure: %s", err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "", err)
		return
	}
	sendJsonServiceEndpointArrayRsp(w, seps)
}

// Get collection of all ServiceEndpoints by service
func (s *SmD) doServiceEndpointsGet(w http.ResponseWriter, r *http.Request) {
	seps := new(sm.ServiceEndpointArray)
	var err error

	svc := chi.URLParam(r, "service")

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doServiceEndpointsGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doServiceEndpointsGet(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	serviceEPFilter := new(hmsds.ServiceEPFilter)
	if err = json.Unmarshal(formJSON, serviceEPFilter); err != nil {
		s.lg.Printf("doServiceEndpointsGet(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	serviceEPFilter.Service = []string{svc}
	seps.ServiceEndpoints, err = s.db.GetServiceEndpointsFilter(serviceEPFilter)
	if err != nil {
		s.lg.Printf("doServiceEndpointsGet(): Lookup failure: %s", err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "", err)
		return
	}
	sendJsonServiceEndpointArrayRsp(w, seps)
}

// Delete single ServiceEndpoint, by its service type and xname ID.
func (s *SmD) doServiceEndpointDelete(w http.ResponseWriter, r *http.Request) {

	svc := chi.URLParam(r, "service")
	xname := chi.URLParam(r, "xname")
	didDelete, err := s.db.DeleteServiceEndpointByID(svc, xname)
	if err != nil {
		s.lg.Printf("doServiceEndpointDelete(): delete failure: (%s,%s) %s", svc, xname, err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such service under redfish endpoint.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Delete entire collection of ServiceEndpoints, undoing discovery.
func (s *SmD) doServiceEndpointsDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeleteServiceEndpointsAll()
	if err != nil {
		s.lg.Printf("doServiceEndpointsDeleteAll(): Delete failure: %s", err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "", err)
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

/////////////////////////////////////////////////////////////////////////////
// Component Ethernet Interfaces
/////////////////////////////////////////////////////////////////////////////

// Delete collection containing all component ethernet interface entries.
func (s *SmD) doCompEthInterfaceDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeleteCompEthInterfacesAll()
	if err != nil {
		s.lg.Printf("doCompEthInterfaceDeleteAll(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}

// Delete component ethernet interface {id}.
func (s *SmD) doCompEthInterfaceDelete(w http.ResponseWriter, r *http.Request) {

	id := strings.ToLower(chi.URLParam(r, "id"))

	if len(id) == 0 {
		s.lg.Printf("doCompEthInterfaceDelete(): Invalid id.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid id.")
		return
	}
	didDelete, err := s.db.DeleteCompEthInterfaceByID(id)
	if err != nil {
		s.lg.Printf("doCompEthInterfaceDelete(): delete failure: (%s) %s", id, err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if !didDelete {
		s.lg.Printf("doCompEthInterfaceDelete(): No such component ethernet interface, %s", id)
		sendJsonError(w, http.StatusNotFound, "no such component ethernet interface.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")

}

// Get all component ethernet interfaces that currently exist, optionally filtering the set,
// returning an array of component ethernet interface records.
func (s *SmD) doCompEthInterfacesGetV2(w http.ResponseWriter, r *http.Request) {
	var err error
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doCompEthInterfacesGetV2(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doCompEthInterfacesGetV2(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	filter := new(CompEthInterfaceFltr)
	if err = json.Unmarshal(formJSON, filter); err != nil {
		s.lg.Printf("doCompEthInterfacesGetV2(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}

	ceiFilter := []hmsds.CompEthInterfaceFiltFunc{}

	if len(filter.ID) > 0 {
		for i, id := range filter.ID {
			if len(id) == 0 {
				s.lg.Printf("doCompEthInterfacesGetV2(): Invalid component ethernet interface ID.")
				sendJsonError(w, http.StatusBadRequest, "Invalid component ethernet interface ID.")
				return
			}
			filter.ID[i] = strings.ToLower(id)
		}
		ceiFilter = append(ceiFilter, hmsds.CEI_IDs(filter.ID))
	}
	if len(filter.MACAddr) > 0 {
		for i, mac := range filter.MACAddr {
			if len(mac) == 0 {
				s.lg.Printf("doCompEthInterfacesGetV2(): Invalid component ethernet interface MAC address.")
				sendJsonError(w, http.StatusBadRequest, "Invalid component ethernet interface MAC address.")
				return
			}
			filter.MACAddr[i] = strings.ToLower(mac)
		}
		ceiFilter = append(ceiFilter, hmsds.CEI_MACAddrs(filter.MACAddr))
	}
	if len(filter.IPAddr) > 0 {
		ceiFilter = append(ceiFilter, hmsds.CEI_IPAddrs(filter.IPAddr))
	}
	if len(filter.Network) > 0 {
		ceiFilter = append(ceiFilter, hmsds.CEI_Networks(filter.Network))
	}
	if len(filter.OlderThan) > 0 {
		ceiFilter = append(ceiFilter, hmsds.CEI_OlderThan(filter.OlderThan[0]))
	}
	if len(filter.NewerThan) > 0 {
		ceiFilter = append(ceiFilter, hmsds.CEI_NewerThan(filter.NewerThan[0]))
	}
	if len(filter.CompID) > 0 {
		for i, xname := range filter.CompID {
			xnameNorm := base.VerifyNormalizeCompID(xname)
			if len(xnameNorm) == 0 && len(xname) != 0 {
				s.lg.Printf("doCompEthInterfacesGetV2(): Invalid CompID.")
				sendJsonError(w, http.StatusBadRequest, "Invalid CompID.")
				return
			}
			filter.CompID[i] = xnameNorm
		}
		ceiFilter = append(ceiFilter, hmsds.CEI_CompIDs(filter.CompID))
	}

	if len(filter.Type) > 0 {
		for i, compType := range filter.Type {
			compTypeNorm := base.VerifyNormalizeType(compType)
			if len(compTypeNorm) == 0 {
				s.lg.Printf("doCompEthInterfacesGetV2(): Invalid HMS type.")
				sendJsonError(w, http.StatusBadRequest, "Invalid HMS type.")
				return
			}
			filter.Type[i] = compTypeNorm
		}
		ceiFilter = append(ceiFilter, hmsds.CEI_CompTypes(filter.Type))
	}
	ceis, err := s.db.GetCompEthInterfaceFilter(ceiFilter...)
	if err != nil {
		s.lg.Printf("doCompEthInterfacesGetV2(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonCompEthInterfaceV2ArrayRsp(w, ceis)

}

// Create a new component ethernet interface.
func (s *SmD) doCompEthInterfacePostV2(w http.ResponseWriter, r *http.Request) {
	var ceiIn sm.CompEthInterfaceV2

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &ceiIn)
	if err != nil {
		s.lg.Printf("doCompEthInterfacePostV2(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}

	mac, err := rf.NormalizeVerifyMAC(ceiIn.MACAddr)
	if err != nil {
		s.lg.Printf("doCompEthInterfacePostV2(): Invalid MAC address: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	cei, err := sm.NewCompEthInterfaceV2(ceiIn.Desc, mac, ceiIn.CompID, ceiIn.IPAddrs)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = s.db.InsertCompEthInterface(cei)
	if err != nil {
		s.lg.Printf("doCompEthInterfacePostV2(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing component ethernet interface that has the same MAC address.")
		} else {
			// Send this message as 500 or 400 plus error message if it is
			// an HMSError and not, e.g. an internal DB error code.
			sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		}
		return
	}

	uri := &sm.ResourceURI{URI: s.compEthIntBaseV2 + "/" + cei.ID}
	sendJsonNewResourceID(w, uri)

}

// Retrieve the component ethernet interface which was created with the given {id}.
func (s *SmD) doCompEthInterfaceGetV2(w http.ResponseWriter, r *http.Request) {
	var err error

	id := strings.ToLower(chi.URLParam(r, "id"))

	if len(id) == 0 {
		s.lg.Printf("doCompEthInterfaceGetV2(): Invalid id.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid id.")
		return
	}

	ceis, err := s.db.GetCompEthInterfaceFilter(hmsds.CEI_ID(id))
	if err != nil {
		s.lg.Printf("doCompEthInterfaceGetV2(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	if len(ceis) == 0 {
		s.lg.Printf("doCompEthInterfaceGetV2(): No such component ethernet interface, %s", id)
		sendJsonError(w, http.StatusNotFound, "No such component ethernet interface: "+id)
		return
	}

	sendJsonCompEthInterfaceV2Rsp(w, ceis[0])

}

// To update the IP address and/or description of a component ethernet interface,
// a PATCH operation can be used. Omitted fields are not updated. LastUpdate is
// only updated if an IP address is specified.
func (s *SmD) doCompEthInterfacePatchV2(w http.ResponseWriter, r *http.Request) {
	var ceip sm.CompEthInterfaceV2Patch

	id := chi.URLParam(r, "id")

	if len(id) == 0 {
		s.lg.Printf("doCompEthInterfacePatchV2(): Invalid id.")
		sendJsonError(w, http.StatusBadRequest, "Invalid id.")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &ceip)
	if err != nil {
		s.lg.Printf("doCompEthInterfacePatchV2(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if ceip.Desc == nil && ceip.IPAddrs == nil && ceip.CompID == nil {
		s.lg.Printf("doCompEthInterfacePatchV2(): Request must have at least one patch field.")
		sendJsonError(w, http.StatusBadRequest, "Request must have at least one patch field.")
		return
	}
	cei, err := s.db.UpdateCompEthInterface(id, &ceip)
	if err != nil {
		s.lg.Printf("doCompEthInterfacePatchV2(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	} else if cei == nil {
		s.lg.Printf("doCompEthInterfacePatchV2(): no such component ethernet interface.")
		sendJsonError(w, http.StatusNotFound, "no such component ethernet interface.")
		return
	}

	sendJsonCompEthInterfaceV2Rsp(w, cei)

}

// Get a array of all IP Addresses mappings that are currently
// associated with this Component Ethernet Interface
func (s *SmD) doCompEthInterfaceIPAddressesGetV2(w http.ResponseWriter, r *http.Request) {
	var err error

	id := chi.URLParam(r, "id")

	if len(id) == 0 {
		s.lg.Printf("doCompEthInterfaceIPAddressesGetV2(): Invalid id.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid id.")
		return
	}

	// Lets reuse the normal DB method to get the component ethernet interface, but only use the IPAddrs
	// field and ignore everything else
	ceis, err := s.db.GetCompEthInterfaceFilter(hmsds.CEI_ID(id))
	if err != nil {
		s.lg.Printf("doCompEthInterfaceIPAddressesGetV2(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	if len(ceis) == 0 {
		s.lg.Printf("doCompEthInterfaceIPAddressesGetV2(): No such component ethernet interface, %s", id)
		sendJsonError(w, http.StatusNotFound, "No such component ethernet interface: "+id)
		return
	}

	ipAddresses := ceis[0].IPAddrs
	sendJsonCompEthInterfaceIPAddressMappingsArrayRsp(w, ipAddresses)
}

// Create a new IP Addresses of Component Ethernet Interface {id} with the IP address {ipaddr} provided
// in the payload. New IP Addresses should not already exist in the given Component Ethernet Interface.
func (s *SmD) doCompEthInterfaceIPAddressPostV2(w http.ResponseWriter, r *http.Request) {
	var ipAddressIn sm.IPAddressMapping

	id := chi.URLParam(r, "id")

	if len(id) == 0 {
		s.lg.Printf("doCompEthInterfaceIPAddressesGetV2(): Invalid id.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid id.")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &ipAddressIn)
	if err != nil {
		s.lg.Printf("doCompEthInterfaceIPAddressPostV2(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}

	ipm, err := sm.NewIPAddressMapping(ipAddressIn.IPAddr, ipAddressIn.Network)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	ipID, err := s.db.AddCompEthInterfaceIPAddress(id, ipm)
	if err != nil {
		s.lg.Printf("doCompEthInterfaceIPAddressPostV2(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSNoCompEthInterface {
			sendJsonError(w, http.StatusNotFound, "No such component ethernet interface: "+id)
		} else if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing IP Address on the same ethernet interface.")
		} else {
			// Send this message as 500 or 400 plus error message if it is
			// an HMSError and not, e.g. an internal DB error code.
			sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		}
		return
	}

	uri := &sm.ResourceURI{URI: s.compEthIntBaseV2 + "/" + id + "/IPAddresses/" + ipID}
	sendJsonNewResourceID(w, uri)
}

// Patch the field fields of an IP Address {ipaddr} associated with Component Ethernet interface {id}
func (s *SmD) doCompEthInterfaceIPAddressPatchV2(w http.ResponseWriter, r *http.Request) {
	var ipmPatch sm.IPAddressMappingPatch

	id := chi.URLParam(r, "id")
	ipaddr := chi.URLParam(r, "ipaddr")

	if len(id) == 0 {
		s.lg.Printf("doCompEthInterfaceMembersDelete(): Invalid id.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid id.")
		return
	}

	if len(ipaddr) == 0 {
		s.lg.Printf("doCompEthInterfaceMembersDelete(): Invalid ip address.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid ip address.")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &ipmPatch)
	if err != nil {
		s.lg.Printf("doCompEthInterfaceIPAddressPostV2(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	ipm, err := s.db.UpdateCompEthInterfaceIPAddress(id, ipaddr, &ipmPatch)
	if err != nil {
		s.lg.Printf("doCompEthInterfaceIPAddressPatchV2(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	} else if ipm == nil {
		s.lg.Printf("doCompEthInterfaceIPAddressPatchV2(): no such IP address in component ethernet interface.")
		sendJsonError(w, http.StatusNotFound, "no such IP address in component ethernet interface.")
		return
	}

	sendJsonCompEthInterfaceIPAddressMappingsRsp(w, ipm)
}

// Remove IP Address {ipaddr} from the IP Addresses of the Component Ethernet Interface {id}.
func (s *SmD) doCompEthInterfaceIPAddressDeleteV2(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	ipAddr := chi.URLParam(r, "ipaddr")

	if len(id) == 0 {
		s.lg.Printf("doCompEthInterfaceMembersDelete(): Invalid id.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid id.")
		return
	}

	if len(ipAddr) == 0 {
		s.lg.Printf("doCompEthInterfaceMembersDelete(): Invalid ip address.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid ip address.")
		return
	}

	didDelete, err := s.db.DeleteCompEthInterfaceIPAddress(id, ipAddr)
	if err != nil {
		s.lg.Printf("doCompEthInterfaceMembersDelete(): delete failure: (%s, %s) %s", id, ipAddr, err)
		if err == hmsds.ErrHMSDSNoCompEthInterface {
			sendJsonError(w, http.StatusNotFound, "No such component ethernet interface: "+id)
		} else {
			sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		}
		return
	}

	if !didDelete {
		s.lg.Printf("doCompEthInterfaceMembersDelete(): No such ip address, %s, in component ethernet interface, %s", ipAddr, id)
		sendJsonError(w, http.StatusNotFound, "component ethernet interface has no such ip address.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

/////////////////////////////////////////////////////////////////////////////
// Discovery
/////////////////////////////////////////////////////////////////////////////

func (s *SmD) doDiscoveryStatusGet(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest,
			"DiscoveryStatus ID not an unsigned integer")
		return
	}
	stat, err := s.db.GetDiscoveryStatusByID(uint(id))
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"Failed due to DB access issue.")
		s.lg.Printf("GetDiscoveryStatusByID failed: %s: %s", r.RemoteAddr, err)
		return
	}
	if stat == nil {
		sendJsonError(w, http.StatusNotFound, "no such DiscoveryStatus ID.")
		return
	}
	sendJsonDiscoveryStatusRsp(w, stat)
}

func (s *SmD) doDiscoveryStatusGetAll(w http.ResponseWriter, r *http.Request) {
	stats, err := s.db.GetDiscoveryStatusAll()
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"Failed due to DB access issue.")
		s.lg.Printf("GetDiscoveryStatusAll failed: %s: %s", r.RemoteAddr, err)
		return
	}
	sendJsonDiscoveryStatusArrayRsp(w, stats)
}

// Do discovery.
func (s *SmD) doInventoryDiscoverPost(w http.ResponseWriter, r *http.Request) {
	var discIn sm.DiscoverIn
	var id uint = 0

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &discIn)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, "POST body was not understood")
		return
	}

	// We got an array of one or more xnames.  If they are valid
	// RedfishEndpoints, discover just this set.
	if len(discIn.XNames) > 0 {
		epsTrimmed := make([]*sm.RedfishEndpoint, 0, 1)
		idMap := make(map[string]bool)
		for _, xname := range discIn.XNames {
			if _, ok := idMap[xname]; ok {
				// Ignore duplicates
				continue
			}
			idMap[xname] = true
			ep, err := s.db.GetRFEndpointByID(xname)
			if err != nil {
				sendJsonError(w, http.StatusInternalServerError,
					"Failed due to DB access issue.")
				s.lg.Printf("GetDiscoveryStatusByID failed: %s: %s",
					r.RemoteAddr, err)
				return
			} else if ep == nil {
				sendJsonError(w, http.StatusNotFound,
					"No such RedfishEndpoint: "+xname)
				return
			}
			epsTrimmed = append(epsTrimmed, ep)
		}
		go s.discoverFromEndpoints(epsTrimmed, id, false, discIn.Force)
	} else {
		// We had no array, default to discovering all RedfishEndpoints
		eps, err := s.db.GetRFEndpointsAll()
		if err != nil {
			sendJsonError(w, http.StatusInternalServerError,
				"operation 'POST' failed due to retrieval from DB")
			s.lg.Printf("GetRFEndpointsAll failed: %s: %s", r.RemoteAddr, err)
			return
		}
		if len(eps) == 0 {
			sendJsonError(w, http.StatusNotFound,
				"RedfishEndpoints collection is empty")
			return
		}
		go s.discoverFromEndpoints(eps, id, false, discIn.Force)
	}
	// We return a link to a set of DiscoveryStatus records.  For now,
	// we only allow one discovery at once and the entry number is
	// always fixed.
	uris := make([]*sm.ResourceURI, 0, 1)
	uri := new(sm.ResourceURI)
	uri.URI = s.invDiscStatusBaseV2 + "/" + strconv.FormatUint(uint64(id), 10)
	uris = append(uris, uri)

	sendJsonResourceIDArray(w, uris)
}

/*
 * SCN Subscription API
 */

// Get all currently held SCN subscriptions. This returns the list of SCN
// subscriptions that this instance of HSM has.
func (s *SmD) doGetSCNSubscriptionsAll(w http.ResponseWriter, r *http.Request) {
	subs, err := s.db.GetSCNSubscriptionsAll()
	if err != nil {
		s.lg.Printf("doGetSCNSubscriptionsAll(): Lookup failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
		return
	}
	sendJsonSCNSubscriptionArrayRsp(w, subs)
}

// Create a new SCN subscription
func (s *SmD) doPostSCNSubscription(w http.ResponseWriter, r *http.Request) {
	var err error
	var found bool

	body, err := ioutil.ReadAll(r.Body)
	// Get the subscriptions
	subIn := new(sm.SCNPostSubscription)
	err = json.Unmarshal(body, subIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if len(subIn.Subscriber) == 0 {
		sendJsonError(w, http.StatusBadRequest, "Missing name of subscriber")
		return
	}
	if len(subIn.Url) == 0 {
		sendJsonError(w, http.StatusBadRequest, "Missing url")
		return
	}
	foundTrigger := false
	if subIn.Enabled != nil && *subIn.Enabled {
		foundTrigger = true
	}
	if len(subIn.Roles) != 0 {
		foundTrigger = true
		for _, rl := range subIn.Roles {
			if role := base.VerifyNormalizeRole(rl); role == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid role '"+rl+"'")
				return
			}
		}
	}
	if len(subIn.SubRoles) != 0 {
		foundTrigger = true
		for _, srl := range subIn.SubRoles {
			if subRole := base.VerifyNormalizeSubRole(srl); subRole == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid subRole '"+srl+"'")
				return
			}
		}
	}
	if len(subIn.SoftwareStatus) != 0 {
		foundTrigger = true
		for _, swStatus := range subIn.SoftwareStatus {
			if len(swStatus) == 0 {
				sendJsonError(w, http.StatusBadRequest, "SoftwareStatus can not be an empty string")
				return
			}
		}
	}
	if len(subIn.States) != 0 {
		foundTrigger = true
		for _, st := range subIn.States {
			if state := base.VerifyNormalizeState(st); state == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid state '"+st+"'")
				return
			}
		}
	}
	if !foundTrigger {
		sendJsonError(w, http.StatusBadRequest, "Missing trigger. Must subscribe to atleast one Enabled, Role, SubRole, SoftwareStatus, or State trigger.")
		return
	}

	s.scnSubLock.Lock()
	// Insert the subscription into the database.
	// Existing subscriptions will be updated.
	id, err := s.db.InsertSCNSubscription(*subIn)
	if err != nil {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusBadRequest, "Subscribe failed")
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		return
	}
	newSub := sm.SCNSubscription{
		ID:             id,
		Subscriber:     subIn.Subscriber,
		Enabled:        subIn.Enabled,
		Roles:          subIn.Roles,
		SubRoles:       subIn.SubRoles,
		SoftwareStatus: subIn.SoftwareStatus,
		States:         subIn.States,
		Url:            subIn.Url,
	}
	// Add or update the cached subscription table.
	// Look for an existing subscription. Update it.
	for i, sub := range s.scnSubs.SubscriptionList {
		if sub.ID == newSub.ID {
			// Remove the old subscription from the scnSubMap
			removeSCNMapSubscription(&s.scnSubMap, &sub)
			// Add the new subscription to the scnSubMap
			addSCNMapSubscription(&s.scnSubMap, &newSub)
			// Update the subscription array.
			s.scnSubs.SubscriptionList[i].States = newSub.States
			s.scnSubs.SubscriptionList[i].Enabled = newSub.Enabled
			s.scnSubs.SubscriptionList[i].Roles = newSub.Roles
			s.scnSubs.SubscriptionList[i].SubRoles = newSub.SubRoles
			s.scnSubs.SubscriptionList[i].SoftwareStatus = newSub.SoftwareStatus
			found = true
			break
		}
	}
	if !found {
		addSCNMapSubscription(&s.scnSubMap, &newSub)
		s.scnSubs.SubscriptionList = append(s.scnSubs.SubscriptionList, newSub)
	}
	s.scnSubLock.Unlock()

	sendJsonSCNSubscriptionRsp(w, &newSub)
}

// Delete all SCN subscriptions.
func (s *SmD) doDeleteSCNSubscriptionsAll(w http.ResponseWriter, r *http.Request) {
	var err error

	s.scnSubLock.Lock()
	// Delete all subscriptions from the db
	numDelete, err := s.db.DeleteSCNSubscriptionsAll()
	if err != nil {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusBadRequest, "Unsubscribe failed")
		s.lg.Printf("failed: %s, Err: %s", r.RemoteAddr, err)
		return
	}
	// Delete all subscriptions from our cached subscription table.
	s.scnSubs.SubscriptionList = s.scnSubs.SubscriptionList[:0]
	s.scnSubMap = SCNSubMap{}
	s.scnSubLock.Unlock()
	sendJsonError(w, http.StatusOK, strconv.FormatInt(numDelete, 10)+" Subscriptions deleted")
}

// Get a currently held SCN subscription. This returns the specified SCN subscription.
func (s *SmD) doGetSCNSubscription(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}
	if id < 1 {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}

	sub, err := s.db.GetSCNSubscription(id)
	if err != nil {
		s.lg.Printf("doGetSCNSubscription(): Lookup failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
		return
	} else if sub == nil {
		sendJsonError(w, http.StatusNotFound, "Subscription not found")
		return
	}
	sendJsonSCNSubscriptionRsp(w, sub)
}

// Update a SCN subscription entirely.
func (s *SmD) doPutSCNSubscription(w http.ResponseWriter, r *http.Request) {
	var err error

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}
	if id < 1 {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	// Get the subscriptions
	subIn := new(sm.SCNPostSubscription)
	err = json.Unmarshal(body, subIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if len(subIn.Subscriber) == 0 {
		sendJsonError(w, http.StatusBadRequest, "Missing name of subscriber")
		return
	}
	if len(subIn.Url) == 0 {
		sendJsonError(w, http.StatusBadRequest, "Missing url")
		return
	}
	foundTrigger := false
	if subIn.Enabled != nil && *subIn.Enabled {
		foundTrigger = true
	}
	if len(subIn.Roles) != 0 {
		foundTrigger = true
		for _, rl := range subIn.Roles {
			if role := base.VerifyNormalizeRole(rl); role == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid role '"+rl+"'")
				return
			}
		}
	}
	if len(subIn.SubRoles) != 0 {
		foundTrigger = true
		for _, srl := range subIn.SubRoles {
			if subRole := base.VerifyNormalizeSubRole(srl); subRole == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid subRole '"+srl+"'")
				return
			}
		}
	}
	if len(subIn.SoftwareStatus) != 0 {
		foundTrigger = true
		for _, swStatus := range subIn.SoftwareStatus {
			if len(swStatus) == 0 {
				sendJsonError(w, http.StatusBadRequest, "SoftwareStatus can not be an empty string")
				return
			}
		}
	}
	if len(subIn.States) != 0 {
		foundTrigger = true
		for _, st := range subIn.States {
			if state := base.VerifyNormalizeState(st); state == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid state '"+st+"'")
				return
			}
		}
	}
	if !foundTrigger {
		sendJsonError(w, http.StatusBadRequest, "Missing trigger. Must subscribe to atleast one Enabled, Role, SubRole, SoftwareStatus, or State trigger.")
		return
	}

	s.scnSubLock.Lock()
	// Update the subscription in the database.
	didUpdate, err := s.db.UpdateSCNSubscription(id, *subIn)
	if err != nil {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusBadRequest, "Subscription update failed")
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		return
	} else if !didUpdate {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusNotFound, "Subscription not found")
		return
	}
	newSub := sm.SCNSubscription{
		ID:             id,
		Subscriber:     subIn.Subscriber,
		Enabled:        subIn.Enabled,
		Roles:          subIn.Roles,
		SubRoles:       subIn.SubRoles,
		SoftwareStatus: subIn.SoftwareStatus,
		States:         subIn.States,
		Url:            subIn.Url,
	}
	// Add or update the cached subscription table.
	// Look for an existing subscription. Update it.
	for i, sub := range s.scnSubs.SubscriptionList {
		if sub.ID == newSub.ID {
			// Remove the old subscription from the scnSubMap
			removeSCNMapSubscription(&s.scnSubMap, &sub)
			// Add the new subscription to the scnSubMap
			addSCNMapSubscription(&s.scnSubMap, &newSub)
			// Update the subscription array.
			s.scnSubs.SubscriptionList[i].States = newSub.States
			break
		}
	}
	s.scnSubLock.Unlock()

	// Send 204 status (success, no content in response)
	sendJsonError(w, http.StatusNoContent, "Success")
}

// Patch update a SCN subscription.
func (s *SmD) doPatchSCNSubscription(w http.ResponseWriter, r *http.Request) {
	var err error

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}
	if id < 1 {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	// Get the subscriptions
	patchIn := new(sm.SCNPatchSubscription)
	err = json.Unmarshal(body, patchIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if len(patchIn.Op) == 0 {
		sendJsonError(w, http.StatusBadRequest, "Missing Patch Op")
		return
	}
	op := sm.GetPatchOp(patchIn.Op)
	if op == sm.PatchOpInvalid {
		sendJsonError(w, http.StatusBadRequest, "Invalid Patch Op - "+patchIn.Op)
		return
	}
	foundTrigger := false
	if patchIn.Enabled != nil && *patchIn.Enabled {
		foundTrigger = true
	}
	if len(patchIn.Roles) != 0 {
		foundTrigger = true
		for _, rl := range patchIn.Roles {
			if role := base.VerifyNormalizeRole(rl); role == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid role '"+rl+"'")
				return
			}
		}
	}
	if len(patchIn.SubRoles) != 0 {
		foundTrigger = true
		for _, srl := range patchIn.SubRoles {
			if subRole := base.VerifyNormalizeSubRole(srl); subRole == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid subRole '"+srl+"'")
				return
			}
		}
	}
	if len(patchIn.SoftwareStatus) != 0 {
		foundTrigger = true
		for _, swStatus := range patchIn.SoftwareStatus {
			if len(swStatus) == 0 {
				sendJsonError(w, http.StatusBadRequest, "SoftwareStatus can not be an empty string")
				return
			}
		}
	}
	if len(patchIn.States) != 0 {
		foundTrigger = true
		for _, st := range patchIn.States {
			if state := base.VerifyNormalizeState(st); state == "" {
				sendJsonError(w, http.StatusBadRequest, "Invalid state '"+st+"'")
				return
			}
		}
	}
	if !foundTrigger {
		sendJsonError(w, http.StatusBadRequest, "Missing trigger. Subscriptions must have atleast one Enabled, Role, SubRole, SoftwareStatus, or State trigger.")
		return
	}

	s.scnSubLock.Lock()
	// Patch the subscription in the database.
	didPatch, err := s.db.PatchSCNSubscription(id, patchIn.Op, *patchIn)
	if err != nil {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusBadRequest, "Subscription patch failed")
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		return
	} else if !didPatch {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusNotFound, "Subscription not found")
		return
	}
	// Patch the cached subscription table.
	// Look for an existing subscription. Patch it.
	// Note: There is a possibility that the cached subscription table is out
	//       of sync with the database. Patch what we have anyway. We'll get
	//       corrected by the SCNSubscriptionRefresh() thread.
	for i, sub := range s.scnSubs.SubscriptionList {
		if sub.ID == id {
			newSub := sm.SCNSubscription{
				ID:         id,
				Subscriber: sub.Subscriber,
				Url:        sub.Url,
			}
			switch op {
			case sm.PatchOpAdd:
				// Find out which values in the request are not already in our
				// current subscription and add them.
				for _, newState := range patchIn.States {
					match := false
					for _, state := range sub.States {
						if state == newState {
							match = true
							break
						}
					}
					if !match {
						newSub.States = append(newSub.States, newState)
						s.scnSubs.SubscriptionList[i].States = append(s.scnSubs.SubscriptionList[i].States, newState)
					}
				}
				for _, newRole := range patchIn.Roles {
					match := false
					for _, role := range sub.Roles {
						if role == newRole {
							match = true
							break
						}
					}
					if !match {
						newSub.Roles = append(newSub.Roles, newRole)
						s.scnSubs.SubscriptionList[i].Roles = append(s.scnSubs.SubscriptionList[i].Roles, newRole)
					}
				}
				for _, newSubRole := range patchIn.SubRoles {
					match := false
					for _, subRole := range sub.SubRoles {
						if subRole == newSubRole {
							match = true
							break
						}
					}
					if !match {
						newSub.SubRoles = append(newSub.SubRoles, newSubRole)
						s.scnSubs.SubscriptionList[i].SubRoles = append(s.scnSubs.SubscriptionList[i].SubRoles, newSubRole)
					}
				}
				for _, newSoftwareStatus := range patchIn.SoftwareStatus {
					match := false
					for _, SoftwareStatus := range sub.SoftwareStatus {
						if SoftwareStatus == newSoftwareStatus {
							match = true
							break
						}
					}
					if !match {
						newSub.SoftwareStatus = append(newSub.SoftwareStatus, newSoftwareStatus)
						s.scnSubs.SubscriptionList[i].SoftwareStatus = append(s.scnSubs.SubscriptionList[i].SoftwareStatus, newSoftwareStatus)
					}
				}
				// The add patch op will only ever change the enabled field from false to true.
				// Only show a change if our request has Enabled=true and our current subscription is enabled=false
				if patchIn.Enabled != nil && *patchIn.Enabled &&
					sub.Enabled != nil && !*sub.Enabled {
					newSub.Enabled = patchIn.Enabled
					s.scnSubs.SubscriptionList[i].Enabled = patchIn.Enabled
				}
				addSCNMapSubscription(&s.scnSubMap, &newSub)
			case sm.PatchOpRemove:
				// Find out which values in the request are in our
				// current subscription and remove them.
				for _, newState := range patchIn.States {
					for j, state := range sub.States {
						if state == newState {
							newSub.States = append(newSub.States, newState)
							s.scnSubs.SubscriptionList[i].States = append(s.scnSubs.SubscriptionList[i].States[:j], s.scnSubs.SubscriptionList[i].States[j+1:]...)
							break
						}
					}
				}
				for _, newRole := range patchIn.Roles {
					for j, role := range sub.Roles {
						if role == newRole {
							newSub.Roles = append(newSub.Roles, newRole)
							s.scnSubs.SubscriptionList[i].Roles = append(s.scnSubs.SubscriptionList[i].Roles[:j], s.scnSubs.SubscriptionList[i].Roles[j+1:]...)
							break
						}
					}
				}
				for _, newSubRole := range patchIn.SubRoles {
					for j, subRole := range sub.SubRoles {
						if subRole == newSubRole {
							newSub.SubRoles = append(newSub.SubRoles, newSubRole)
							s.scnSubs.SubscriptionList[i].SubRoles = append(s.scnSubs.SubscriptionList[i].SubRoles[:j], s.scnSubs.SubscriptionList[i].SubRoles[j+1:]...)
							break
						}
					}
				}
				for _, newSoftwareStatus := range patchIn.SoftwareStatus {
					for j, SoftwareStatus := range sub.SoftwareStatus {
						if SoftwareStatus == newSoftwareStatus {
							newSub.SoftwareStatus = append(newSub.SoftwareStatus, newSoftwareStatus)
							s.scnSubs.SubscriptionList[i].SoftwareStatus = append(s.scnSubs.SubscriptionList[i].SoftwareStatus[:j], s.scnSubs.SubscriptionList[i].SoftwareStatus[j+1:]...)
							break
						}
					}
				}
				// The remove patch op will only ever change the enabled field from true to false.
				// Only show a change if our request has Enabled=true and our current subscription is Enabled=true
				if patchIn.Enabled != nil && *patchIn.Enabled &&
					sub.Enabled != nil && *sub.Enabled {
					newSub.Enabled = patchIn.Enabled
					*s.scnSubs.SubscriptionList[i].Enabled = false
				}
				removeSCNMapSubscription(&s.scnSubMap, &newSub)
			case sm.PatchOpReplace:
				removeSCNMapSubscription(&s.scnSubMap, &sub)
				if len(patchIn.States) > 0 {
					s.scnSubs.SubscriptionList[i].States = patchIn.States
				}
				if len(patchIn.Roles) > 0 {
					s.scnSubs.SubscriptionList[i].Roles = patchIn.Roles
				}
				if len(patchIn.SubRoles) > 0 {
					s.scnSubs.SubscriptionList[i].SubRoles = patchIn.SubRoles
				}
				if len(patchIn.SoftwareStatus) > 0 {
					s.scnSubs.SubscriptionList[i].SoftwareStatus = patchIn.SoftwareStatus
				}
				if patchIn.Enabled != nil {
					s.scnSubs.SubscriptionList[i].Enabled = patchIn.Enabled
				}
				addSCNMapSubscription(&s.scnSubMap, &s.scnSubs.SubscriptionList[i])
			default:
				// Shouldn't happen
				sendJsonError(w, http.StatusBadRequest, "Invalid Patch Op - "+patchIn.Op)
				return
			}
			break
		}
	}
	s.scnSubLock.Unlock()

	// Send 204 status (success, no content in response)
	sendJsonError(w, http.StatusNoContent, "Success")
}

// Delete a SCN subscription.
func (s *SmD) doDeleteSCNSubscription(w http.ResponseWriter, r *http.Request) {
	var err error

	idStr := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}
	if id < 1 {
		sendJsonError(w, http.StatusBadRequest, "Invalid id - "+idStr)
		return
	}

	s.scnSubLock.Lock()
	// Delete the subscription from the db
	didDelete, err := s.db.DeleteSCNSubscription(id)
	if err != nil {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusBadRequest, "Unsubscribe failed")
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, idStr, err)
		return
	}
	// Delete the subscription from our cached subscription table.
	if didDelete {
		// Find the subscription.
		for i, sub := range s.scnSubs.SubscriptionList {
			if sub.ID == id {
				// Found a subscription. Remove it from the map.
				removeSCNMapSubscription(&s.scnSubMap, &sub)
				// Remove the subscription from the subscription array.
				s.scnSubs.SubscriptionList = append(s.scnSubs.SubscriptionList[:i], s.scnSubs.SubscriptionList[i+1:]...)
				break
			}
		}
	} else {
		s.scnSubLock.Unlock()
		sendJsonError(w, http.StatusNotFound, "Subscription not found")
		return
	}
	s.scnSubLock.Unlock()
	sendJsonError(w, http.StatusOK, "Subscription deleted")
}

/*
 * HSM Groups API
 */

// Get all groups that currently exist, optionally filtering the set, returning
// an array of groups.
func (s *SmD) doGroupsGet(w http.ResponseWriter, r *http.Request) {
	var err error
	groups := make([]sm.Group, 0)
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doGroupsGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doGroupsGet(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	groupFilter := new(GrpPartFltr)
	if err = json.Unmarshal(formJSON, groupFilter); err != nil {
		s.lg.Printf("doGroupsGet(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	part := ""
	if len(groupFilter.Partition) > 0 {
		part = groupFilter.Partition[0]
		if part != "NULL" {
			part = sm.NormalizeGroupField(part)
			if sm.VerifyGroupField(part) != nil {
				s.lg.Printf("doGroupsGet(): Invalid partition name.")
				sendJsonError(w, http.StatusBadRequest,
					"Invalid partition name.")
				return
			}
		}
	}
	for i, tag := range groupFilter.Tag {
		tag = sm.NormalizeGroupField(tag)
		if sm.VerifyGroupField(tag) != nil {
			s.lg.Printf("doGroupsGet(): Invalid tag.")
			sendJsonError(w, http.StatusBadRequest,
				"Invalid tag.")
			return
		}
		groupFilter.Tag[i] = tag
	}
	for i, label := range groupFilter.Group {
		label = sm.NormalizeGroupField(label)
		if sm.VerifyGroupField(label) != nil {
			s.lg.Printf("doGroupsGet(): Invalid group label.")
			sendJsonError(w, http.StatusBadRequest,
				"Invalid group label.")
			return
		}
		groupFilter.Group[i] = label
	}
	// TODO: Make this one db call. Not in the initial implementation.
	labels, err := s.db.GetGroupLabels()
	if err != nil {
		s.lg.Printf("doGroupsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	for _, label := range labels {
		foundLabel := false
		if len(groupFilter.Group) > 0 {
			for _, labelMatch := range groupFilter.Group {
				if labelMatch == label {
					foundLabel = true
					break
				}
			}
		} else {
			foundLabel = true
		}
		if !foundLabel {
			continue
		}
		group, err := s.db.GetGroup(label, part)
		if err != nil {
			s.lg.Printf("doGroupsGet(): Lookup failure: %s", err)
			sendJsonDBError(w, "bad query param: ", "", err)
			return
		}
		if group == nil {
			// Shouldn't happen but ignore if it does.
			continue
		}
		foundTag := false
		if len(groupFilter.Tag) > 0 {
			for _, tag := range group.Tags {
				for _, tagMatch := range groupFilter.Tag {
					if tagMatch == tag {
						foundTag = true
						break
					}
				}
			}
		} else {
			foundTag = true
		}
		if !foundTag {
			continue
		}
		groups = append(groups, *group)
	}
	sendJsonGroupArrayRsp(w, &groups)

}

// Create a new group identified by the label field. Label should be given
// explicitly, and should not conflict with any existing group, or an error
// will occur.
func (s *SmD) doGroupsPost(w http.ResponseWriter, r *http.Request) {
	var groupIn sm.Group

	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &groupIn)
	if err != nil {
		s.lg.Printf("doGroupsPost(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	group, err := sm.NewGroup(
		groupIn.Label,
		groupIn.Description,
		groupIn.ExclusiveGroup,
		groupIn.Tags,
		groupIn.Members.IDs)
	if err != nil {
		s.lg.Printf("doGroupsPost(): Couldn't validate group: %s", err)
		sendJsonError(w, http.StatusBadRequest,
			"couldn't validate group: "+err.Error())
		return
	}
	label, err := s.db.InsertGroup(group)
	if err != nil {
		s.lg.Printf("doGroupsPost(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing group that has the same label.")
		} else if err == hmsds.ErrHMSDSExclusiveGroup {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing member in another exclusive group.")
		} else {
			// Send this message as 500 or 400 plus error message if it is
			// an HMSError and not, e.g. an internal DB error code.
			sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		}
		return
	}

	uris := []*sm.ResourceURI{{URI: s.groupsBaseV2 + "/" + label}}
	sendJsonNewResourceIDArray(w, s.groupsBaseV2, uris)

}

// Retrieve the group which was created with the given {group_label}.
func (s *SmD) doGroupGet(w http.ResponseWriter, r *http.Request) {
	var err error

	label := sm.NormalizeGroupField(chi.URLParam(r, "group_label"))

	if sm.VerifyGroupField(label) != nil {
		s.lg.Printf("doGroupGet(): Invalid group label.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid group label.")
		return
	}
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doGroupGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doGroupGet(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	groupFilter := new(GrpPartFltr)
	if err = json.Unmarshal(formJSON, groupFilter); err != nil {
		s.lg.Printf("doGroupGet(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	part := ""
	if len(groupFilter.Partition) > 0 {
		part = groupFilter.Partition[0]
		if part != "NULL" {
			part = sm.NormalizeGroupField(part)
			if sm.VerifyGroupField(part) != nil {
				s.lg.Printf("doGroupGet(): Invalid partition name.")
				sendJsonError(w, http.StatusBadRequest,
					"Invalid partition name.")
				return
			}
		}
	}
	group, err := s.db.GetGroup(label, part)
	if err != nil {
		s.lg.Printf("doGroupGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	if group == nil {
		s.lg.Printf("doGroupGet(): No such group, %s", label)
		sendJsonError(w, http.StatusNotFound, "No such group: "+label)
		return
	}

	sendJsonGroupRsp(w, group)

}

// Delete the given group label. Any members previously in the group will no
// longer have the deleted group label associated with them.
func (s *SmD) doGroupDelete(w http.ResponseWriter, r *http.Request) {

	label := sm.NormalizeGroupField(chi.URLParam(r, "group_label"))

	if sm.VerifyGroupField(label) != nil {
		s.lg.Printf("doGroupDelete(): Invalid group label.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid group label.")
		return
	}
	didDelete, err := s.db.DeleteGroup(label)
	if err != nil {
		s.lg.Printf("doGroupDelete(): delete failure: (%s) %s", label, err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if !didDelete {
		s.lg.Printf("doGroupDelete(): No such group, %s", label)
		sendJsonError(w, http.StatusNotFound, "no such group.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")

}

// To update the tags array and/or description, a PATCH operation can be used.
// Omitted fields are not updated.
// NOTE: This cannot be used to completely replace the members list. Rather,
//
//	individual members can be removed or added with the
//	POST/DELETE {group_label}/members API.
func (s *SmD) doGroupPatch(w http.ResponseWriter, r *http.Request) {
	var groupPatch sm.GroupPatch

	label := sm.NormalizeGroupField(chi.URLParam(r, "group_label"))

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &groupPatch)
	if err != nil {
		s.lg.Printf("doGroupPatch(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if groupPatch.Description == nil && groupPatch.Tags == nil {
		s.lg.Printf("doGroupPatch(): Request must have at least one patch field.")
		sendJsonError(w, http.StatusBadRequest,
			"Request must have at least one patch field.")
		return
	}
	if sm.VerifyGroupField(label) != nil {
		s.lg.Printf("doGroupPatch(): Invalid group label.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid group label.")
		return
	}
	if groupPatch.Tags != nil {
		for _, tag := range *groupPatch.Tags {
			tagNorm := sm.NormalizeGroupField(tag)
			if sm.VerifyGroupField(tagNorm) != nil {
				s.lg.Printf("doGroupPatch(): Invalid tag.")
				sendJsonError(w, http.StatusBadRequest,
					"Invalid tag.")
				return
			}
		}
	}
	err = s.db.UpdateGroup(label, &groupPatch)
	if err != nil {
		s.lg.Printf("doGroupPatch(): Lookup failure: %s", err)
		if err == hmsds.ErrHMSDSNoGroup {
			sendJsonError(w, http.StatusNotFound, "no such group.")
		} else {
			sendJsonDBError(w, "bad query param: ", "", err)
		}
		return
	}

	sendJsonError(w, http.StatusNoContent, "Success")

}

// Get a string array of all group labels (i.e. group names) that currently
// exist in HSM.
func (s *SmD) doGroupLabelsGet(w http.ResponseWriter, r *http.Request) {
	labels, err := s.db.GetGroupLabels()
	if err != nil {
		s.lg.Printf("doGroupLabelsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonStringArrayRsp(w, &labels)

}

// Get all members of an existing group {group_label}, optionally filtering the set,
// returning a members set containing the component xname IDs.
func (s *SmD) doGroupMembersGet(w http.ResponseWriter, r *http.Request) {
	var err error

	label := sm.NormalizeGroupField(chi.URLParam(r, "group_label"))

	if sm.VerifyGroupField(label) != nil {
		s.lg.Printf("doGroupMembersGet(): Invalid group label.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid group label.")
		return
	}

	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doGroupMembersGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doGroupMembersGet(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	groupFilter := new(GrpPartFltr)
	if err = json.Unmarshal(formJSON, groupFilter); err != nil {
		s.lg.Printf("doGroupMembersGet(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	part := ""
	if len(groupFilter.Partition) > 0 {
		part = groupFilter.Partition[0]
		if part != "NULL" {
			part = sm.NormalizeGroupField(groupFilter.Partition[0])
			if sm.VerifyGroupField(part) != nil {
				s.lg.Printf("doGroupsGet(): Invalid partition name.")
				sendJsonError(w, http.StatusBadRequest,
					"Invalid partition name.")
				return
			}
		}
	}
	group, err := s.db.GetGroup(label, part)
	if err != nil {
		s.lg.Printf("doGroupMembersGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	if group == nil {
		s.lg.Printf("doGroupMembersGet(): No such group, %s", label)
		sendJsonError(w, http.StatusNotFound, "No such group: "+label)
		return
	}

	sendJsonMembersRsp(w, &group.Members)

}

// Create a new member of group {group_label} with the component xname id provided
// in the payload. New member should not already exist in the given group.
func (s *SmD) doGroupMembersPost(w http.ResponseWriter, r *http.Request) {
	var memberIn sm.MemberAddBody

	label := sm.NormalizeGroupField(chi.URLParam(r, "group_label"))

	if sm.VerifyGroupField(label) != nil {
		s.lg.Printf("doGroupMemberPost(): Invalid group label.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid group label.")
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &memberIn)
	if err != nil {
		s.lg.Printf("doGroupMemberPost(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	normID := base.NormalizeHMSCompID(memberIn.ID)
	if !base.IsHMSCompIDValid(normID) {
		s.lg.Printf("doGroupMemberPost(): Invalid xname ID.")
		sendJsonError(w, http.StatusBadRequest, "invalid xname ID")
		return
	}
	id, err := s.db.AddGroupMember(label, normID)
	if err != nil {
		s.lg.Printf("doGroupMemberPost(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSNoGroup {
			sendJsonError(w, http.StatusNotFound, "No such group: "+label)
		} else if err == hmsds.ErrHMSDSExclusiveGroup {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing member in another exclusive group.")
		} else if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing member in the same group.")
		} else {
			// Send this message as 500 or 400 plus error message if it is
			// an HMSError and not, e.g. an internal DB error code.
			sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		}
		return
	}

	uris := []*sm.ResourceURI{{URI: s.groupsBaseV2 + "/" + label + "/members/" + id}}
	sendJsonNewResourceIDArray(w, s.groupsBaseV2, uris)

}

// Remove component {xname_id} from the members of group {group_label}.
func (s *SmD) doGroupMemberDelete(w http.ResponseWriter, r *http.Request) {

	label := sm.NormalizeGroupField(chi.URLParam(r, "group_label"))
	id := base.NormalizeHMSCompID(chi.URLParam(r, "xname_id"))

	if sm.VerifyGroupField(label) != nil {
		s.lg.Printf("doGroupMemberDelete(): Invalid group label.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid group label.")
		return
	}
	if !base.IsHMSCompIDValid(id) {
		s.lg.Printf("doGroupMemberDelete(): Invalid xname ID.")
		sendJsonError(w, http.StatusBadRequest, "invalid xname ID")
		return
	}
	didDelete, err := s.db.DeleteGroupMember(label, id)
	if err != nil {
		s.lg.Printf("doGroupMemberDelete(): delete failure: (%s, %s) %s", label, id, err)
		if err == hmsds.ErrHMSDSNoGroup {
			sendJsonError(w, http.StatusNotFound, "No such group: "+label)
		} else {
			sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		}
		return
	}
	if !didDelete {
		s.lg.Printf("doGroupMemberDelete(): No such member, %s, in group, %s", id, label)
		sendJsonError(w, http.StatusNotFound, "group has no such member.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")

}

/*
 * HSM Partitions API
 */

// Get all partitions that currently exist, optionally filtering the set,
// returning an array of partition records.
func (s *SmD) doPartitionsGet(w http.ResponseWriter, r *http.Request) {
	var err error
	partitions := make([]sm.Partition, 0)
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doPartitionsGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doPartitionsGet(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	filter := new(GrpPartFltr)
	if err = json.Unmarshal(formJSON, filter); err != nil {
		s.lg.Printf("doPartitionsGet(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	for i, tag := range filter.Tag {
		tagNorm := sm.NormalizeGroupField(tag)
		if sm.VerifyGroupField(tagNorm) != nil {
			s.lg.Printf("doPartitionsGet(): Invalid tag.")
			sendJsonError(w, http.StatusBadRequest,
				"Invalid tag.")
			return
		}
		filter.Tag[i] = tagNorm
	}
	for i, partition := range filter.Partition {
		partNorm := sm.NormalizeGroupField(partition)
		if sm.VerifyGroupField(partNorm) != nil {
			s.lg.Printf("doPartitionsGet(): Invalid partition name.")
			sendJsonError(w, http.StatusBadRequest,
				"Invalid partition name.")
			return
		}
		filter.Partition[i] = partNorm
	}
	// TODO: Make this one db call. Not in the initial implementation.
	pnames, err := s.db.GetPartitionNames()
	if err != nil {
		s.lg.Printf("doPartitionsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	for _, pname := range pnames {
		foundName := false
		if len(filter.Partition) > 0 {
			for _, pnameMatch := range filter.Partition {
				if pnameMatch == pname {
					foundName = true
					break
				}
			}
		} else {
			foundName = true
		}
		if !foundName {
			continue
		}
		partition, err := s.db.GetPartition(pname)
		if err != nil {
			s.lg.Printf("doPartitionsGet(): Lookup failure: %s", err)
			sendJsonDBError(w, "bad query param: ", "", err)
			return
		}
		if partition == nil {
			// Shouldn't happen but ignore if it does.
			continue
		}
		foundTag := false
		if len(filter.Tag) > 0 {
			for _, tag := range partition.Tags {
				for _, tagMatch := range filter.Tag {
					if tagMatch == tag {
						foundTag = true
						break
					}
				}
			}
		} else {
			foundTag = true
		}
		if !foundTag {
			continue
		}
		partitions = append(partitions, *partition)
	}
	sendJsonPartitionArrayRsp(w, &partitions)

}

// Create a new partition identified by the name field. Partition name
// should be given explicitly, and should not conflict with any existing
// partition, or an error will occur. In addition, the member list must not
// overlap with any existing partition.
func (s *SmD) doPartitionsPost(w http.ResponseWriter, r *http.Request) {
	var partIn sm.Partition

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &partIn)
	if err != nil {
		s.lg.Printf("doPartitionsPost(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	part, err := sm.NewPartition(
		partIn.Name,
		partIn.Description,
		partIn.Tags,
		partIn.Members.IDs)
	if err != nil {
		s.lg.Printf("doPartitionsPost(): Couldn't validate partition: %s", err)
		sendJsonError(w, http.StatusBadRequest,
			"couldn't validate partition: "+err.Error())
		return
	}
	name, err := s.db.InsertPartition(part)
	if err != nil {
		s.lg.Printf("doPartitionsPost(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing partition that has the same name.")
		} else if err == hmsds.ErrHMSDSExclusivePartition {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing member in another partition.")
		} else {
			// Send this message as 500 or 400 plus error message if it is
			// an HMSError and not, e.g. an internal DB error code.
			sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		}
		return
	}

	uris := []*sm.ResourceURI{{URI: s.partitionsBaseV2 + "/" + name}}
	sendJsonNewResourceIDArray(w, s.partitionsBaseV2, uris)

}

// Retrieve the partition which was created with the given {partition_name}.
func (s *SmD) doPartitionGet(w http.ResponseWriter, r *http.Request) {
	var err error

	name := sm.NormalizeGroupField(chi.URLParam(r, "partition_name"))

	if sm.VerifyGroupField(name) != nil {
		s.lg.Printf("doPartitionGet(): Invalid partition name.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid partition name.")
		return
	}

	part, err := s.db.GetPartition(name)
	if err != nil {
		s.lg.Printf("doPartitionGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	if part == nil {
		s.lg.Printf("doPartitionGet(): No such partition, %s", name)
		sendJsonError(w, http.StatusNotFound, "No such partition: "+name)
		return
	}

	sendJsonPartitionRsp(w, part)

}

// Delete partition {partition_name}. Any members previously in the partition
// will no longer have the deleted partition name associated with them.
func (s *SmD) doPartitionDelete(w http.ResponseWriter, r *http.Request) {

	name := sm.NormalizeGroupField(chi.URLParam(r, "partition_name"))

	if sm.VerifyGroupField(name) != nil {
		s.lg.Printf("doPartitionDelete(): Invalid partition name.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid partition name.")
		return
	}
	didDelete, err := s.db.DeletePartition(name)
	if err != nil {
		s.lg.Printf("doPartitionDelete(): delete failure: (%s) %s", name, err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if !didDelete {
		s.lg.Printf("doPartitionDelete(): No such partition, %s", name)
		sendJsonError(w, http.StatusNotFound, "no such partition.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")

}

// To update the tags array and/or description, a PATCH operation can be used.
// Omitted fields are not updated.
// NOTE: This cannot be used to completely replace the members list. Rather,
//
//	individual members can be removed or added with the POST/DELETE
//	{partition_name}/members API.
func (s *SmD) doPartitionPatch(w http.ResponseWriter, r *http.Request) {
	var partPatch sm.PartitionPatch

	name := sm.NormalizeGroupField(chi.URLParam(r, "partition_name"))

	if sm.VerifyGroupField(name) != nil {
		s.lg.Printf("doPartitionPatch(): Invalid partition name.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid partition name.")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &partPatch)
	if err != nil {
		s.lg.Printf("doPartitionPatch(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if partPatch.Description == nil && partPatch.Tags == nil {
		s.lg.Printf("doPartitionPatch(): Request must have at least one patch field.")
		sendJsonError(w, http.StatusBadRequest,
			"Request must have at least one patch field.")
		return
	}
	if sm.VerifyGroupField(name) != nil {
		s.lg.Printf("doPartitionPatch(): Invalid partition name.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid partition name.")
		return
	}
	if partPatch.Tags != nil {
		for _, tag := range *partPatch.Tags {
			tagNorm := sm.NormalizeGroupField(tag)
			if sm.VerifyGroupField(tagNorm) != nil {
				s.lg.Printf("doPartitionPatch(): Invalid tag.")
				sendJsonError(w, http.StatusBadRequest,
					"Invalid tag.")
				return
			}
		}
	}
	err = s.db.UpdatePartition(name, &partPatch)
	if err != nil {
		s.lg.Printf("doPartitionPatch(): Lookup failure: %s", err)
		if err == hmsds.ErrHMSDSNoPartition {
			sendJsonError(w, http.StatusNotFound, "no such partition.")
		} else {
			sendJsonDBError(w, "bad query param: ", "", err)
		}
		return
	}

	sendJsonError(w, http.StatusNoContent, "Success")

}

// Get a string array of all partition names that currently exist in HSM.
// These are just the names, not the complete partition records.
func (s *SmD) doPartitionNamesGet(w http.ResponseWriter, r *http.Request) {
	names, err := s.db.GetPartitionNames()
	if err != nil {
		s.lg.Printf("doPartitionNamesGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonStringArrayRsp(w, &names)

}

// Get all members of existing partition {partition_name}, optionally filtering
// the set, returning a members set that includes the component xname IDs.
func (s *SmD) doPartitionMembersGet(w http.ResponseWriter, r *http.Request) {
	var err error

	name := sm.NormalizeGroupField(chi.URLParam(r, "partition_name"))

	if sm.VerifyGroupField(name) != nil {
		s.lg.Printf("doPartitionMembersGet(): Invalid partition name.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid partition name.")
		return
	}

	part, err := s.db.GetPartition(name)
	if err != nil {
		s.lg.Printf("doPartitionMembersGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	if part == nil {
		s.lg.Printf("doPartitionMembersGet(): No such partition, %s", name)
		sendJsonError(w, http.StatusNotFound, "No such partition: "+name)
		return
	}

	sendJsonMembersRsp(w, &part.Members)

}

// Create a new member of partition {partition_name} with the component xname
// id provided in the payload. New member should not already exist in the given
// partition
func (s *SmD) doPartitionMembersPost(w http.ResponseWriter, r *http.Request) {
	var memberIn sm.MemberAddBody

	name := sm.NormalizeGroupField(chi.URLParam(r, "partition_name"))

	if sm.VerifyGroupField(name) != nil {
		s.lg.Printf("doPartitionMembersPost(): Invalid partition name.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid partition name.")
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &memberIn)
	if err != nil {
		s.lg.Printf("doPartitionMembersPost(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	normID := memberIn.ID
	if !s.ochami {
		// CSM requires that the ID is an xname.
		// OCHAMI allows for any string.
		normID = base.NormalizeHMSCompID(memberIn.ID)
		if !base.IsHMSCompIDValid(normID) {
			s.lg.Printf("doPartitionMembersPost(): Invalid xname ID.")
			sendJsonError(w, http.StatusBadRequest, "invalid xname ID")
			return
		}
	}
	id, err := s.db.AddPartitionMember(name, normID)
	if err != nil {
		s.lg.Printf("doPartitionMembersPost(): %s %s Err: %s", r.RemoteAddr,
			string(body), err)
		if err == hmsds.ErrHMSDSNoPartition {
			sendJsonError(w, http.StatusNotFound, "No such partition: "+name)
		} else if err == hmsds.ErrHMSDSExclusivePartition {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing member in another partition.")
		} else if err == hmsds.ErrHMSDSDuplicateKey {
			sendJsonError(w, http.StatusConflict, "operation would conflict "+
				"with an existing member in the same partition.")
		} else {
			// Send this message as 500 or 400 plus error message if it is
			// an HMSError and not, e.g. an internal DB error code.
			sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		}
		return
	}

	uris := []*sm.ResourceURI{{URI: s.partitionsBaseV2 + "/" + name + "/members/" + id}}
	sendJsonNewResourceIDArray(w, s.partitionsBaseV2, uris)

}

// Remove component {xname_id} from the members of partition {partition_name}.
func (s *SmD) doPartitionMemberDelete(w http.ResponseWriter, r *http.Request) {

	name := sm.NormalizeGroupField(chi.URLParam(r, "partition_name"))
	id := base.NormalizeHMSCompID(chi.URLParam(r, "xname_id"))

	if sm.VerifyGroupField(name) != nil {
		s.lg.Printf("doPartitionMemberDelete(): Invalid partition name.")
		sendJsonError(w, http.StatusBadRequest,
			"Invalid partition name.")
		return
	}
	if !base.IsHMSCompIDValid(id) {
		s.lg.Printf("doPartitionMemberDelete(): Invalid xname ID.")
		sendJsonError(w, http.StatusBadRequest, "invalid xname ID")
		return
	}
	didDelete, err := s.db.DeletePartitionMember(name, id)
	if err != nil {
		s.lg.Printf("doPartitionMemberDelete(): delete failure: (%s, %s) %s", name, id, err)
		if err == hmsds.ErrHMSDSNoPartition {
			sendJsonError(w, http.StatusNotFound, "No such partition: "+name)
		} else {
			sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		}
		return
	}
	if !didDelete {
		s.lg.Printf("doPartitionMemberDelete(): No such member, %s, in partition, %s", id, name)
		sendJsonError(w, http.StatusNotFound, "partition has no such member.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")

}

/*
 * HSM Memberships API
 */

func (s *SmD) doMembershipsGet(w http.ResponseWriter, r *http.Request) {
	var err error

	// Parse arguments
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doMembershipsGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doMembershipsGet(): Marshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	compFilter := new(hmsds.ComponentFilter)
	if err = json.Unmarshal(formJSON, compFilter); err != nil {
		s.lg.Printf("doMembershipsGet(): Unmarshal form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	memberships, err := s.db.GetMemberships(compFilter)
	if err != nil {
		s.lg.Printf("doMembershipsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	sendJsonMembershipArrayRsp(w, memberships)
}

func (s *SmD) doMembershipGet(w http.ResponseWriter, r *http.Request) {
	var err error

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	if !base.IsHMSCompIDValid(xname) {
		s.lg.Printf("doMembershipGet(): Invalid xname.")
		sendJsonError(w, http.StatusBadRequest, "invalid xname")
		return
	}
	membership, err := s.db.GetMembership(xname)
	if err != nil {
		s.lg.Printf("doMembershipGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "bad query param: ", "", err)
		return
	}
	if membership == nil {
		s.lg.Printf("doMembershipGet(): No such xname, %s", xname)
		sendJsonError(w, http.StatusNotFound, "No such xname: "+xname)
		return
	}
	sendJsonMembershipRsp(w, membership)

}

/*
 * HSM Component Lock API
 */

func (s *SmD) compLocksV2Helper(w http.ResponseWriter, r *http.Request, action string) {
	var filter sm.CompLockV2Filter

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksV2%s(): Unmarshal body: %s", action, err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksV2%s(): Couldn't validate component lock filter: %s", action, err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	results, err := s.db.UpdateCompLocksV2(filter, action)
	if err != nil {
		s.lg.Printf("doCompLocksV2%s(): %s %s Err: %s", action, r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}

	sendJsonCompLockV2UpdateRsp(w, results)

}

func (s *SmD) doCompLocksReservationRemove(w http.ResponseWriter, r *http.Request) {
	var filter sm.CompLockV2Filter

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksReservationRemove(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksReservationRemove(): Couldn't validate component reservation filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	results, err := s.db.DeleteCompReservationsForce(filter)
	if err != nil {
		s.lg.Printf("doCompLocksReservationRemove(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}

	sendJsonCompLockV2UpdateRsp(w, results)

}

func (s *SmD) doCompLocksReservationRelease(w http.ResponseWriter, r *http.Request) {
	var filter sm.CompLockV2ReservationFilter

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksReservationRelease(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksReservationRelease(): Couldn't validate component reservation filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	results, err := s.db.DeleteCompReservations(filter)
	if err != nil {
		s.lg.Printf("doCompLocksReservationRelease(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}

	sendJsonCompLockV2UpdateRsp(w, results)

}

func (s *SmD) doCompLocksReservationCreate(w http.ResponseWriter, r *http.Request) {
	var filter sm.CompLockV2Filter

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksReservationCreate(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksReservationCreate(): Couldn't validate component reservation filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	filter.ReservationDuration = 0
	results, err := s.db.InsertCompReservations(filter)
	if err != nil {
		s.lg.Printf("doCompLocksReservationCreate(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}

	sendJsonCompReservationRsp(w, results)

}

func (s *SmD) doCompLocksServiceReservationRenew(w http.ResponseWriter, r *http.Request) {
	var filter sm.CompLockV2ReservationFilter

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationRenew(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationRenew(): Couldn't validate component reservation filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if filter.ReservationDuration <= 0 {
		s.lg.Printf("doCompLocksServiceReservationRenew(): ReservationDuration must be greater than 0")
		sendJsonError(w, http.StatusBadRequest, "ReservationDuration must be greater than 0")
		return
	}
	results, err := s.db.UpdateCompReservations(filter)
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationRenew(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}

	sendJsonCompLockV2UpdateRsp(w, results)

}

func (s *SmD) doCompLocksServiceReservationRelease(w http.ResponseWriter, r *http.Request) {
	s.doCompLocksReservationRelease(w, r)

}

func (s *SmD) doCompLocksServiceReservationCreate(w http.ResponseWriter, r *http.Request) {
	var filter sm.CompLockV2Filter

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationCreate(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationCreate(): Couldn't validate component reservation filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if filter.ReservationDuration <= 0 {
		s.lg.Printf("doCompLocksServiceReservationCreate(): ReservationDuration must be greater than 0")
		sendJsonError(w, http.StatusBadRequest, "ReservationDuration must be greater than 0")
		return
	}
	results, err := s.db.InsertCompReservations(filter)
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationCreate(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}

	sendJsonCompReservationRsp(w, results)

}

func (s *SmD) doCompLocksServiceReservationCheck(w http.ResponseWriter, r *http.Request) {
	var filter sm.CompLockV2DeputyKeyArray

	body, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationCheck(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationCheck(): Couldn't validate component reservation filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	results, err := s.db.GetCompReservations(filter.DeputyKeys)
	if err != nil {
		s.lg.Printf("doCompLocksServiceReservationCheck(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}

	sendJsonCompReservationRsp(w, results)

}

func (s *SmD) doCompLocksStatus(w http.ResponseWriter, r *http.Request) {
	var filter sm.CompLockV2Filter
	var results sm.CompLockV2Status
	results.Components = make([]sm.CompLockV2, 0, 1)
	results.NotFound = make([]string, 0, 1)

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &filter)
	if err != nil {
		s.lg.Printf("doCompLocksStatus(): Unmarshal body: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksStatus(): Couldn't validate component lock filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	locks, err := s.db.GetCompLocksV2(filter)
	if err != nil {
		s.lg.Printf("doCompLocksStatus(): %s %s Err: %s", r.RemoteAddr, string(body), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'POST' failed during store.", err)
		return
	}
	results.Components = locks
	lockMap := make(map[string]bool)
	if len(locks) != len(filter.ID) {
		for _, lock := range locks {
			lockMap[lock.ID] = true
		}
		for _, id := range filter.ID {
			if _, ok := lockMap[id]; !ok {
				results.NotFound = append(results.NotFound, id)
			}
		}
	}

	sendJsonCompLockV2Rsp(w, results)

}

func (s *SmD) doCompLocksStatusGet(w http.ResponseWriter, r *http.Request) {
	var inFilter CompGetLockFltr
	var filter sm.CompLockV2Filter
	var results sm.CompLockV2Status
	results.Components = make([]sm.CompLockV2, 0, 1)

	// Parse query parameters
	if err := r.ParseForm(); err != nil {
		s.lg.Printf("doCompLocksStatusGet(): ParseForm: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	formJSON, err := json.Marshal(r.Form)
	if err != nil {
		s.lg.Printf("doCompLocksStatusGet(): Marshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	if err = json.Unmarshal(formJSON, &inFilter); err != nil {
		s.lg.Printf("doCompLocksStatusGet(): Unmarshall form: %s", err)
		sendJsonError(w, http.StatusInternalServerError,
			"failed to decode query parameters.")
		return
	}
	filter = compGetLockFltrToCompLockV2Filter(inFilter)
	err = filter.VerifyNormalize()
	if err != nil {
		s.lg.Printf("doCompLocksStatusGet(): Couldn't validate component lock filter: %s", err)
		sendJsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if filter.Reserved != nil {
		// Only use the first Reserved query parameter supplied since
		// asking for reserved and unreserved components with multiple
		// query parameters in one call doesn't make sense.
		_, err = strconv.ParseBool(filter.Reserved[0])
		if err != nil {
			reservedParamBoolErrMsg := "bad 'Reserved' query parameter: " + filter.Reserved[0]
			s.lg.Printf("doCompLocksStatusGet(): " + reservedParamBoolErrMsg)
			sendJsonError(w, http.StatusBadRequest, reservedParamBoolErrMsg)
			return
		}
	}
	locks, err := s.db.GetCompLocksV2(filter)
	if err != nil {
		s.lg.Printf("doCompLocksStatus(): %s %s Err: %s", r.RemoteAddr, string(formJSON), err)
		// Send this message as 500 or 400 plus error message if it is
		// an HMSError and not, e.g. an internal DB error code.
		sendJsonDBError(w, "", "operation 'GET' failed during query.", err)
		return
	}
	results.Components = locks

	sendJsonCompLockV2Rsp(w, results)

}

func (s *SmD) doCompLocksLock(w http.ResponseWriter, r *http.Request) {
	s.compLocksV2Helper(w, r, hmsds.CLUpdateActionLock)

}

func (s *SmD) doCompLocksUnlock(w http.ResponseWriter, r *http.Request) {
	s.compLocksV2Helper(w, r, hmsds.CLUpdateActionUnlock)

}

func (s *SmD) doCompLocksRepair(w http.ResponseWriter, r *http.Request) {
	s.compLocksV2Helper(w, r, hmsds.CLUpdateActionRepair)

}

func (s *SmD) doCompLocksDisable(w http.ResponseWriter, r *http.Request) {
	s.compLocksV2Helper(w, r, hmsds.CLUpdateActionDisable)

}

/////////////////////////////////////////////////////////////////////////////
// Power Mappings
/////////////////////////////////////////////////////////////////////////////

// Get one specific PowerMap entry, previously created, by its xname ID.
func (s *SmD) doPowerMapGet(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doPowerMapGet(): trying...")

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))
	if !base.IsHMSCompIDValid(xname) {
		s.lg.Printf("doPowerMapGet(): Invalid xname.")
		sendJsonError(w, http.StatusBadRequest, "invalid xname")
		return
	}
	m, err := s.db.GetPowerMapByID(xname)
	if err != nil {
		s.LogAlways("doPowerMapGet(): Lookup failure: (%s) %s",
			xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if m == nil {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonPowerMapRsp(w, m)
}

// Get all PowerMap entries in database, by doing a GET against the
// entire collection.
func (s *SmD) doPowerMapsGet(w http.ResponseWriter, r *http.Request) {
	ms, err := s.db.GetPowerMapsAll()
	if err != nil {
		s.LogAlways("doPowerMapsGet(): Lookup failure: %s", err)
		sendJsonDBError(w, "", "", err)
		return
	}
	sendJsonPowerMapArrayRsp(w, ms)
}

// CREATE new or UPDATE EXISTING Power mapping
func (s *SmD) doPowerMapsPost(w http.ResponseWriter, r *http.Request) {
	msIn := make([]sm.PowerMap, 0, 1)
	ms := make([]sm.PowerMap, 0, 1)

	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &msIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if len(msIn) > 0 {
		for i, mIn := range msIn {
			if len(mIn.PoweredBy) == 0 {
				sendJsonError(w, http.StatusBadRequest,
					"poweredby is required for PowerMaps")
				return
			}
			// Attempt to create a valid PowerMaps from the
			// raw data.  If we do not get any errors, it should be sane enough
			// to put into the data store.
			m, err := sm.NewPowerMap(mIn.ID, mIn.PoweredBy)
			if err != nil {
				idx := strconv.Itoa(i)
				sendJsonError(w, http.StatusBadRequest,
					"couldn't validate map data at idx "+idx+": "+err.Error())
				return
			}
			ms = append(ms, *m)
		}
	}
	err = s.db.InsertPowerMaps(ms)
	if err != nil {
		s.lg.Printf("failed: %s %s Err: %s", r.RemoteAddr, string(body), err)
		sendJsonError(w, http.StatusInternalServerError,
			"operation 'POST' failed during store. ")
		return
	}

	numStr := strconv.FormatInt(int64(len(ms)), 10)
	sendJsonError(w, http.StatusOK, "Created or modified "+numStr+" entries")

}

// UPDATE EXISTING Power mapping by it's xname URI.
func (s *SmD) doPowerMapPut(w http.ResponseWriter, r *http.Request) {

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	var mIn sm.PowerMap
	body, err := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(body, &mIn)
	if err != nil {
		sendJsonError(w, http.StatusInternalServerError,
			"error decoding JSON "+err.Error())
		return
	}
	if mIn.ID == "" {
		if xname != "" {
			mIn.ID = xname
		}
	} else if base.NormalizeHMSCompID(mIn.ID) != xname {
		sendJsonError(w, http.StatusBadRequest,
			"xname in URL and PUT body do not match")
		return
	}

	if len(mIn.PoweredBy) == 0 {
		sendJsonError(w, http.StatusBadRequest,
			"poweredby is required in PUT body")
		return
	}
	// Make sure the information submitted is a proper PowerMap and will
	// not update the entry with invalid data.
	m, err := sm.NewPowerMap(mIn.ID, mIn.PoweredBy)
	if err != nil {
		sendJsonError(w, http.StatusBadRequest,
			"couldn't validate PowerMap data: "+err.Error())
		return
	}
	err = s.db.InsertPowerMap(m)
	if err != nil {
		s.lg.Printf("failed: %s %s, Err: %s", r.RemoteAddr, string(body), err)
		// Unexpected error on update
		sendJsonError(w, http.StatusInternalServerError,
			"operation 'PUT' failed during store")
		return
	}
	sendJsonPowerMapRsp(w, m)

}

// Delete single PowerMap, by its xname ID.
func (s *SmD) doPowerMapDelete(w http.ResponseWriter, r *http.Request) {
	s.lg.Printf("doPowerMapDelete(): trying...")

	xname := base.NormalizeHMSCompID(chi.URLParam(r, "xname"))

	if !base.IsHMSCompIDValid(xname) {
		sendJsonError(w, http.StatusBadRequest, "invalid xname")
		return
	}
	didDelete, err := s.db.DeletePowerMapByID(xname)
	if err != nil {
		s.LogAlways("doPowerMapDelete(): delete failure: (%s) %s",
			xname, err)
		sendJsonDBError(w, "", "", err)
		return
	}
	if !didDelete {
		sendJsonError(w, http.StatusNotFound, "no such xname.")
		return
	}
	sendJsonError(w, http.StatusOK, "deleted 1 entry")
}

// Delete collection containing all PowerMap entries.
func (s *SmD) doPowerMapsDeleteAll(w http.ResponseWriter, r *http.Request) {
	var err error
	numDeleted, err := s.db.DeletePowerMapsAll()
	if err != nil {
		s.lg.Printf("doPowerMapsDelete(): Delete failure: %s", err)
		sendJsonError(w, http.StatusInternalServerError, "DB query failed.")
		return
	}
	if numDeleted == 0 {
		sendJsonError(w, http.StatusNotFound, "no entries to delete")
		return
	}
	numStr := strconv.FormatInt(numDeleted, 10)
	sendJsonError(w, http.StatusOK, "deleted "+numStr+" entries")
}
