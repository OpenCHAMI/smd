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
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-smd/v2/internal/hmsds"
	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
)

var e = base.NewHMSError("sm", "server error")

var ErrSMDNoUpType = e.NewChild("missing UpdateType")
var ErrSMDBadUpType = e.NewChild("invalid UpdateType")
var ErrSMDUpTypeConf = e.NewChild("conflicting UpdateType")

var ErrSMDInternal = e.NewChild("internal server error")
var ErrSMDBadID = e.NewChild("invalid xname ID")
var ErrSMDNoIDs = e.NewChild("Missing IDs")
var ErrSMDNoID = e.NewChild("Missing ID")
var ErrSMDIDConf = e.NewChild("ID in URL and PATCH body do not match")
var ErrSMDNoState = e.NewChild("Missing State")
var ErrSMDNoFlag = e.NewChild("Missing Flag")
var ErrSMDNoEnabled = e.NewChild("Missing Enabled field")
var ErrSMDNoSwStatus = e.NewChild("Missing SoftwareStatus")
var ErrSMDNoRole = e.NewChild("Missing Role")
var ErrSMDNoNID = e.NewChild("Missing NID")
var ErrSMDTooManyIDs = e.NewChild("too many IDs")

type CompUpdateType string

// CompUpdateType
const (
	StateDataUpdate   CompUpdateType = "StateData"
	FlagOnlyUpdate    CompUpdateType = "FlagOnly"
	EnabledUpdate     CompUpdateType = "Enabled"
	SwStatusUpdate    CompUpdateType = "SoftwareStatus"
	RoleUpdate        CompUpdateType = "Role"
	SingleNIDUpdate   CompUpdateType = "NID"
	CompUpdateInvalid CompUpdateType = "INVALID" // Not an actual type, invalid
)

// For case-insensitive verification and normalization of state strings
var compUpdateTypeMap = map[string]CompUpdateType{
	"statedata":      StateDataUpdate,
	"flagonly":       FlagOnlyUpdate,
	"enabled":        EnabledUpdate,
	"softwarestatus": SwStatusUpdate,
	"role":           RoleUpdate,
	"nid":            SingleNIDUpdate,
}

// if a valid type is given, return the normalized string form.
// Else, return the empty string.
func VerifyNormalizeCompUpdateType(utype string) string {
	utypeLower := strings.ToLower(utype)
	utypeTrimmed := strings.TrimPrefix(utypeLower, "update")
	value, ok := compUpdateTypeMap[utypeTrimmed]
	if !ok {
		return ""
	} else {
		return value.String()
	}
}

// Turn a case-insensitive string into the corresponding CompUpdateType.
// If there is no match, return CompUpdateInvalid
func GetCompUpdateType(utype string) CompUpdateType {
	utypeLower := strings.ToLower(utype)
	utypeTrimmed := strings.TrimPrefix(utypeLower, "update")
	value, ok := compUpdateTypeMap[utypeTrimmed]
	if !ok {
		return CompUpdateInvalid
	} else {
		return value
	}
}

// Allow CompUpdateType to be treated as a standard string type.
func (ut CompUpdateType) String() string { return string(ut) }

type CompUpdate struct {
	ComponentIDs []string `json:"ComponentIDs"`
	State        string   `json:"State,omitempty"`
	Flag         string   `json:"Flag,omitempty"`
	Enabled      *bool    `json:"Enabled,omitempty"`
	SwStatus     *string  `json:"SoftwareStatus,omitempty"`
	Role         *string  `json:"Role,omitempty"`
	SubRole      *string  `json:"SubRole,omitempty"`
	NID          *int64   `json:"NID,omitempty"`
	Subtype      string   `json:"Subtype,omitempty"`
	NetType      string   `json:"NetType,omitempty"`
	Arch         string   `json:"Arch,omitempty"`

	Group        []string        `json:"Group"`
	Partition    []string        `json:"Partition"`
	UpdateType   string          `json:"UpdateType,omitempty"`
	Force        bool            `json:"Force,omitempty"`
	ExtendedInfo json.RawMessage `json:"ExtendedInfo,omitempty"`
}

// Update the database based on the input fields and the selected operation.
// Then send any SCN messages required.  This is intended to be used
// for REST operations and operations that occur due to message bus events.
func (s *SmD) doCompUpdate(u *CompUpdate, name string) error {
	var data base.Component
	pi := new(hmsds.PartInfo)
	compIDs := []string{}
	scnIDs := []string{}
	skipSCNs := false

	if u == nil {
		s.LogAlways("WARNING: %s: got nil pointer", name)
		return ErrSMDInternal
	}

	// Validate arguments, should be at least one ID and all should
	// be valid, normalized xnames (remove leading zeroes and stuff).
	if len(u.ComponentIDs) == 0 {
		return ErrSMDNoIDs
	} else {
		for _, id := range u.ComponentIDs {
			normID := base.VerifyNormalizeCompID(id)
			if normID == "" {
				return ErrSMDBadID
			}
			compIDs = append(compIDs, normID)
		}
	}
	// This isn't used currently, but will need to be used to filter
	// every operation so we don't have things touching data they
	// shouldn't.
	pi.Group = append(pi.Group, u.Group...)
	pi.Partition = append(pi.Partition, u.Partition...)

	var err error
	switch GetCompUpdateType(u.UpdateType) {
	case StateDataUpdate:
		nflag := u.Flag
		if u.State == "" {
			return ErrSMDNoState
		}
		if u.Flag == "" {
			nflag = base.FlagOK.String()
		}
		data.State = base.VerifyNormalizeState(u.State)
		data.Flag = base.VerifyNormalizeFlag(nflag)
		scnIDs, err = s.dbUpdateCompState(compIDs, u.State, nflag, u.Force, pi)
		if err == nil {
			if data.State == base.StateStandby.String() {
				// Start State Redfish Polling jobs for any nodes
				// transitioning to standby.
				for _, id := range scnIDs {
					if base.GetHMSTypeString(id) == base.Node.String() {
						s.doStateRFPoll(id, 30)
					}
				}
			} else {
				// Find and cancel State Redfish Polling jobs for any component
				// transitioning to a state other than standby.
				for _, id := range scnIDs {
					s.cancelStateRFPoll(id)
				}
			}
		}
	case FlagOnlyUpdate:
		// This should work, but we don't support it as a valid SCN type
		// now.
		skipSCNs = true
		if u.Flag == "" {
			return ErrSMDNoFlag
		}
		data.Flag = base.VerifyNormalizeFlag(u.Flag)
		scnIDs, err = s.dbUpdateCompFlagOnly(compIDs, u.Flag, pi)
	case EnabledUpdate:
		if u.Enabled == nil {
			return ErrSMDNoEnabled
		}
		data.Enabled = u.Enabled
		scnIDs, err = s.dbUpdateCompEnabled(compIDs, u.Enabled, pi)
	case SwStatusUpdate:
		if u.SwStatus == nil {
			return ErrSMDNoSwStatus
		}
		data.SwStatus = *u.SwStatus
		scnIDs, err = s.dbUpdateCompSwStatus(compIDs, *u.SwStatus, pi)
	case RoleUpdate:
		subRole := ""
		if u.Role == nil {
			return ErrSMDNoRole
		}
		data.Role = base.VerifyNormalizeRole(*u.Role)
		if u.SubRole != nil {
			subRole = *u.SubRole
			data.SubRole = base.VerifyNormalizeSubRole(subRole)
		}
		scnIDs, err = s.dbUpdateCompRole(compIDs, *u.Role, subRole, pi)
	case SingleNIDUpdate:
		if u.NID == nil {
			return ErrSMDNoNID
		}
		// No SCN ever for NID updates (at the moment)
		skipSCNs = true
		err = s.dbUpdateCompSingleNID(compIDs, *u.NID, pi)
	default:
		s.LogAlways("Error: %s: doCompUpdate: bad CompUpdateType: '%s'",
			name, u.UpdateType)
		return ErrSMDBadUpType
	}
	if err != nil {
		return err
	}
	// Send SCN if there were changes.
	if len(scnIDs) != 0 && !skipSCNs {
		scn := NewJobSCN(scnIDs, data, s)
		s.wp.Queue(scn)
	}
	return nil
}

// For either single or bulk State/Flag updates.  Single updates are faster
// because we only have one target and don't need a second query to see if it
// needs to be changed.  We can just see what happens.
func (s *SmD) dbUpdateCompState(
	ids []string,
	state, flag string,
	force bool,
	pi *hmsds.PartInfo,
) ([]string, error) {
	return s.db.UpdateCompStates(ids, state, flag, force, pi)
}

// For either single or bulk Flag-only updates (state is not affected).  Single
// updates are faster because we only have one target and don't need a second
// query to see if it needs to be changed.  We can just see what happens.
func (s *SmD) dbUpdateCompFlagOnly(
	ids []string,
	flag string,
	pi *hmsds.PartInfo,
) ([]string, error) {
	if len(ids) == 1 {
		rowsAffected, err := s.db.UpdateCompFlagOnly(ids[0], flag)
		if rowsAffected != 0 {
			return []string{ids[0]}, err
		} else {
			return []string{}, err
		}
	} else if len(ids) > 1 {
		return s.db.BulkUpdateCompFlagOnly(ids, flag)
	}
	return []string{}, ErrSMDNoIDs
}

// For either single or bulk enabledSoftwareStatus field updates.  Single updates are faster
// because we only have one target and don't need a second query to see if it
// needs to be changed.  We can just see what happens.
func (s *SmD) dbUpdateCompEnabled(
	ids []string,
	enabled *bool,
	pi *hmsds.PartInfo,
) ([]string, error) {
	if len(ids) == 1 {
		rowsAffected, err := s.db.UpdateCompEnabled(ids[0], *enabled)
		if rowsAffected != 0 {
			return []string{ids[0]}, err
		}
		return []string{}, err
	} else if len(ids) > 1 {
		return s.db.BulkUpdateCompEnabled(ids, *enabled)
	}
	return []string{}, ErrSMDNoIDs
}

// For either single or bulk SoftwareStatus field updates.  Single updates are
// faster because we only have one target and don't need a second query to
// see if it needs to be changed, we can just see what happens.
func (s *SmD) dbUpdateCompSwStatus(
	ids []string,
	swstatus string,
	pi *hmsds.PartInfo,
) ([]string, error) {
	if len(ids) == 1 {
		rowsAffected, err := s.db.UpdateCompSwStatus(ids[0], swstatus)
		if rowsAffected != 0 {
			return []string{ids[0]}, err
		}
		return []string{}, err
	} else if len(ids) > 1 {
		return s.db.BulkUpdateCompSwStatus(ids, swstatus)
	}
	return []string{}, ErrSMDNoIDs
}

// For either single or bulk Role/SubRole field updates.  Single updates are
// faster because we only have one target and don't need a second query to
// see if it needs to be changed, we can just see what happens.
func (s *SmD) dbUpdateCompRole(
	ids []string,
	role string,
	subRole string,
	pi *hmsds.PartInfo,
) ([]string, error) {
	if len(ids) == 1 {
		rowsAffected, err := s.db.UpdateCompRole(ids[0], role, subRole)
		if rowsAffected != 0 {
			return []string{ids[0]}, err
		}
		return []string{}, err
	} else if len(ids) > 1 {
		return s.db.BulkUpdateCompRole(ids, role, subRole)
	}
	return []string{}, ErrSMDNoIDs
}

// For single node NID updates only.  Obviously we cannot assign the
// same NID to more than one component.
func (s *SmD) dbUpdateCompSingleNID(
	ids []string,
	nid int64,
	pi *hmsds.PartInfo,
) error {
	if len(ids) == 1 {
		comp := base.Component{
			ID:  ids[0],
			NID: json.Number(strconv.FormatInt(nid, 10)),
		}
		err := s.db.UpdateCompNID(&comp)
		return err
	} else if len(ids) > 1 {
		return ErrSMDTooManyIDs
	}
	return ErrSMDNoIDs
}

// Starts a State Redfish Poll job for a component.
func (s *SmD) doStateRFPoll(id string, delay int) error {
	var err error
	job := new(Job)
	job.job, err = sm.NewStateRFPollJob(id, delay, 10, 30, 20)
	if err != nil {
		return err
	}
	s.jobLock.Lock()
	job.job.Id, err = s.db.InsertJob(job.job)
	if err != nil {
		s.jobLock.Unlock()
		return err
	}
	job.cancelChan = make(chan bool)
	s.jobList[job.job.Id] = job
	s.srfpJobList[id] = job
	s.jobLock.Unlock()
	go s.doPollRFState(job)
	return nil
}

// Periodically poll for the redfish powerstate of a component. This gathers
// the existing ComponentEndpoint and credentials present in either the secure
// store or the database (if the secure store is not enabled) for the xname to
// check the power state of the component via redfish.
// This polls for component redfish state every 10 seconds with an initial
// start delay of 30 seconds until the redfish state changes to "off" or
// cancelled.
func (s *SmD) doPollRFState(job *Job) {
	var (
		user string
		pw   string
		cep  *sm.ComponentEndpoint
		ep   *rf.RedfishEP
		url  string
		err  error
	)

	data, ok := job.job.Data.(*sm.SrfpJobData)
	if !ok {
		return
	}

	// Start the job's keep alive goroutine
	kaStopChan := make(chan bool)
	go s.doSRFPKeepAlive(job.job, kaStopChan)
	defer func() { kaStopChan <- true }()

	// This initial ticker will delay the start of the polling for 30 seconds
	// to allow for the possibility of redfish events.
	ticker := time.NewTicker(time.Duration(data.Delay) * time.Second)
	defer ticker.Stop()

	delayed := false
	haveData := false
	setErrStatus := false
	// Initially true to cause an "InProgress" update.
	updatedErrStatus := true
	for {
		// Update the job status only when it changes
		if setErrStatus && !updatedErrStatus {
			s.db.UpdateJob(job.job.Id, sm.JobError)
			updatedErrStatus = true
		} else if !setErrStatus && updatedErrStatus {
			s.db.UpdateJob(job.job.Id, sm.JobInProgress)
			updatedErrStatus = false
		}
		setErrStatus = false
		select {
		case <-job.cancelChan:
			return
		case <-ticker.C:
			// No redfish event came in. Gather the data for polling.
			// If there are errors gathering the info, we'll retry
			// every 10 seconds.
			if !delayed {
				ticker.Stop()
				ticker = time.NewTicker(time.Duration(data.Poll) * time.Second)
				delayed = true
			}
			if !haveData {
				cep, err = s.db.GetCompEndpointByID(data.CompId)
				if err != nil {
					continue
				} else if cep == nil {
					s.Log(LOG_INFO, "doPollRFState(%s): ComponentEndpoint not found.", data.CompId)
					setErrStatus = true
					continue
				}
				if s.readVault {
					// Read component endpoint credentials from the secure store.
					cred, err := s.ccs.GetCompCred(data.CompId)
					if err != nil {
						s.Log(LOG_INFO, "doPollRFState(%s): Failed to get credentials from Vault - %s",
							data.CompId, err)
						setErrStatus = true
						// Retry
						continue
					}
					if len(cred.Username) == 0 {
						// Receive empty credentials from vault for the component
						s.Log(LOG_INFO, "doPollRFState(%s): No credentials in Vault - %s",
							data.CompId, err)
						setErrStatus = true
						// Retry
						continue
					}
					user = cred.Username
					pw = cred.Password
				} else {
					// Get credentials from the database redfishendpoints if Vault isn't enabled
					rep, err := s.db.GetRFEndpointByID(cep.RfEndpointID)
					if err != nil {
						s.Log(LOG_INFO, "doPollRFState(%s): Lookup failure: %s", data.CompId, err)
						setErrStatus = true
						// Retry
						continue
					}
					if rep == nil {
						s.Log(LOG_INFO, "doPollRFState(%s): No credentials in the database for %s",
							data.CompId, cep.RfEndpointID)
						setErrStatus = true
						// Retry
						continue
					}
					user = rep.User
					pw = rep.Password
				}
				// Minimally populate a redfish description struct
				rfEPType := base.GetHMSType(cep.RfEndpointID)
				epDesc := rf.RedfishEPDescription{
					ID:       cep.RfEndpointID,
					Type:     rfEPType.String(),
					FQDN:     cep.RfEndpointFQDN,
					User:     user,
					Password: pw,
				}
				// Make a RedfishEP struct to make our redfish call against.
				url = cep.RfEndpointFQDN + cep.OdataID
				ep, _ = rf.NewRedfishEp(&epDesc)
				haveData = true
			}
			// The data needed to start polling has been gathered.
			// Poll the component for redfish state.
			if haveData {
				rfJSON, err := ep.GETRelative(cep.OdataID)
				if err != nil {
					s.Log(LOG_INFO, "doPollRFState(%s): redfish call failed: %s: %s",
						data.CompId, url, err)
					setErrStatus = true
					// Maybe the redfishEndpoint data changed?
					haveData = false
					continue
				}
				// Parse the data that we received
				rfData := new(rf.ComputerSystem)
				if err := json.Unmarshal(rfJSON, &rfData); err != nil {
					if rf.IsUnmarshalTypeError(err) {
						s.Log(LOG_INFO, "doPollRFState(%s): bad field(s) skipped: %s: %s",
							data.CompId, url, err)
					} else {
						s.Log(LOG_INFO, "doPollRFState(%s): json decode failed: %s: %s",
							data.CompId, url, err)
						setErrStatus = true
						continue
					}
				}
				if strings.ToLower(rfData.PowerState) == "off" {
					s.db.UpdateJob(job.job.Id, sm.JobComplete)
					update := new(CompUpdate)
					update.ComponentIDs = []string{data.CompId}
					update.UpdateType = StateDataUpdate.String()
					update.State = base.StateOff.String()
					s.doCompUpdate(update, "doPollRFState")
					// No return here because doCompUpdate() will signal our
					// cancelChan. We'll wait to stop that way.
				}
			}
		}
	}
}

// This is the keep alive function for State Redfish Polling jobs. It updates
// the job's status every 30 seconds until cancelled. It will cancel the job
// itself if the component either does not exist or the component's state is
// no longer standby.
func (s *SmD) doSRFPKeepAlive(job *sm.Job, stopChan chan bool) {
	data, ok := job.Data.(*sm.SrfpJobData)
	if !ok {
		// Shouldn't happen
		return
	}

	keepAlive := time.NewTicker(time.Duration(job.KeepAlive) * time.Second)
	defer keepAlive.Stop()

	for {
		cancelled := false
		select {
		case <-stopChan:
			return
		case <-keepAlive.C:
			if cancelled {
				continue
			}
			comp, err := s.db.GetComponentByID(data.CompId)
			if err != nil {
				// Error. Retry?
				continue
			}
			if comp == nil {
				// Not found. Stop.
				s.cancelStateRFPoll(data.CompId)
				cancelled = true
			} else if comp.State == base.StateStandby.String() {
				// Still in Standby
				s.db.UpdateJob(job.Id, "")
			} else {
				// State was changed. We're done
				s.cancelStateRFPoll(data.CompId)
				cancelled = true
				// No return here because cancelStateRFPoll() will cause our
				// parent signal our stopChan. We'll wait to stop that way.
			}
		}
	}
}

// Signals State Redfish Polling jobs to stop.
func (s *SmD) cancelStateRFPoll(id string) {
	// Check to see if we have a polling job for this id. If not, there
	// is either no job or another HSM instance owns the job. The other
	// HSM instance will cancel its job when it sees that the state has
	// changed.
	s.jobLock.Lock()
	defer s.jobLock.Unlock()
	if hsmJob, ok := s.srfpJobList[id]; ok {
		go func(cancelChan chan bool) {
			cancelChan <- true
		}(hsmJob.cancelChan)
		s.db.DeleteJob(hsmJob.job.Id)
		delete(s.srfpJobList, id)
		delete(s.jobList, hsmJob.job.Id)
	}
}
