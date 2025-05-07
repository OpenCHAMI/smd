// MIT License
//
// (C) Copyright [2019-2024] Hewlett Packard Enterprise Development LP
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
	"fmt"
	"strings"
	"sync"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/OpenCHAMI/smd/v2/internal/hmsds"
	rf "github.com/OpenCHAMI/smd/v2/pkg/redfish"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

var em = base.NewHMSError("sm.msgbus", "internal error")

var ErrSmMsgNoPayload = em.NewChild("empty event payload")
var ErrSmMsgBadDecode = em.NewChild("payload failed to decode")
var ErrSmMsgBadProcess = em.NewChild("event failed processing")
var ErrSmMsgNilEvent = em.NewChild("nil Event")
var ErrSmMsgNilEventRec = em.NewChild("nil EventRecord")
var ErrSmMsgNilProcRFE = em.NewChild("nil processedRFEvent")
var ErrSmMsgNoIDCtx = em.NewChild("no xname ID in any context")
var ErrSmMsgNoMsgId = em.NewChild("no MessageId found")
var ErrSmMsgBadID = em.NewChild("target xname ID is not valid")
var ErrSmMsgNoID = em.NewChild("no target xname ID found")
var ErrSmMsgNoURI = em.NewChild("no target URI found")
var ErrSmMsgNoEP = em.NewChild("no target ComponentEndpoint ID found")
var ErrSmMsgNoCreds = em.NewChild("no target component credentials found")
var ErrSmMsgRFFail = em.NewChild("redfish call failed")
var ErrSmMsgRFNoInfo = em.NewChild("no info from redfish")
var ErrSmMsgCredsStore = em.NewChild("failed to get credentials from secure store")
var ErrSmMsgCredsDB = em.NewChild("failed to get credentials from database")
var ErrSmMsgIgnState = em.NewChild("ignored state change type")
var ErrSmMsgBadCached = em.NewChild("bad decode of cached type from interface")
var ErrSmMsgNoPowerState = em.NewChild("missing power state value")
var ErrSmMsgMissedSync = em.NewChild("unexpectedly missing after sync")
var ErrSmMsgFiltered = em.NewChild("message(s) filtered due to wrong type")

type processedRFEvent struct {
	MessageId     string
	SubLabels     []string
	Registry      string
	RegVersion    string
	RfEndppointID string
	Origin        string
	Severity      string
	Message       string
	MessageArgs   []string
}

// Take a string-encoded Redfish Event from the message bus and take
// the appropriate action (if any), as far as state changes or actions.
// Call once per event.  Do not retry on error unless specifically told
// to do so.
func (s *SmD) doHandleRFEvent(eventRaw string) error {
	if eventRaw == "" {
		return ErrSmMsgNoPayload
	}
	// Decode Redfish Event from raw input.  Should never be nil on non-error.
	e, err := rf.EventDecode([]byte(eventRaw))
	if err != nil {
		// Don't abort.  The error might be in a field we don't need.
		// It will generally fill in everything it can.
		s.Log(LOG_INFO, "warning: doHandleRFEvent: '%s': '%s'", err, eventRaw)
	}
	// Process the Redfish event into normalized form that hides the
	// underlying implementation.
	pes, err := s.processRFEvent(e)
	if err != nil {
		// Should rarely return errors unless something is really screwed up.
		// We expect to get a lot of the wrong kind of error and we normally
		// just ignore (and possibly log) them.
		return err
	}
	if len(pes) > 0 {
		s.Log(LOG_DEBUG, "Received event: '%s'", eventRaw)
	}
	for _, pe := range pes {
		update, err := s.compUpdateFromRFEvent(pe)
		if err != nil {
			s.Log(LOG_INFO, "handleRFEvent: getting CompUpdate: '%s' ", err)
			continue
		} else if update == nil {
			continue
		}
		s.Log(LOG_INFO, "CHANGING STATE: %s->%s: calling doCompUpdate(%s) CompUpdateType=%s",
			pe.RfEndppointID, pe.MessageId, update.ComponentIDs, update.UpdateType)
		err = s.doCompUpdate(update, "handleRFEvent")
		if err != nil {
			s.LogAlways("ERROR: %s->%s: calling doCompUpdate(%s): %s",
				pe.RfEndppointID, pe.MessageId, update.ComponentIDs, err)
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// Normalization of raw Events into processedRFEvent structs that hide
// individual Redfish implementation details.
//
// Most of the event decoding logic goes here and there will probably
// need to be updates when new Redfish stacks are introduced.
/////////////////////////////////////////////////////////////////////////////

// Take a Redfish JSON event, decoded into a struct, and produce a
// decodedRFEvent struct or structs depending on whether multiple independent
// EventRecords require separate actions.
//
// This normalizes the implementation differences of individual events and
// processes them into a form that allows State Manager to treat all events
// the same basic way.
func (s *SmD) processRFEvent(e *rf.Event) ([]*processedRFEvent, error) {
	pes := make([]*processedRFEvent, 0, 1) // Returned array
	if e == nil {
		return pes, ErrSmMsgNilEvent
	}
	for i, erec := range e.Events {
		pe, err := s.processRFEventRecord(e, &erec)
		if err != nil {
			// We ignore errors, but report them unless they're things
			// we don't care about.
			if err == ErrSmMsgNoIDCtx {
				s.Log(LOG_INFO, "Event: %s: '%s'/'%s'",
					err, e.Context, erec.Context)
			} else if err != ErrSmMsgFiltered {
				s.Log(LOG_INFO, "EventRecord %d: %s: '%v'", i, err, e)
			}
			continue
		}
		pes = append(pes, pe)
	}
	return pes, nil
}

// Processes one individual EventRecord 'erec', in an event 'e'.
func (s *SmD) processRFEventRecord(
	e *rf.Event,
	erec *rf.EventRecord,
) (*processedRFEvent, error) {
	if e == nil || erec == nil {
		return nil, ErrSmMsgNilEventRec
	}
	// Start processing the event.
	reg, ver, msgid := rf.EventRecordMsgId(erec)
	if msgid == "" {
		// No message id, can't identify event
		return nil, ErrSmMsgNoMsgId
	} else if ignoredRegistry(reg) {
		return nil, ErrSmMsgFiltered
	}
	id, labels := GetEventIDAndLabels(e.Context, erec.Context)
	if id == "" {
		// No controller ID found.
		return nil, ErrSmMsgNoIDCtx
	}
	// Enough info to process event,
	pe := new(processedRFEvent)
	pe.RfEndppointID = id
	pe.SubLabels = labels
	pe.Registry = reg
	pe.RegVersion = ver
	pe.MessageId = msgid

	pe.Message = erec.MessageId
	pe.MessageArgs = erec.MessageArgs
	pe.Origin = erec.OriginOfCondition.Oid
	pe.Severity = erec.Severity

	return pe, nil
}

// Filter registries - Things we know we don't care about
func ignoredRegistry(reg string) bool {
	switch strings.ToLower(reg) {
	case "craytelemetry":
		return true
	}
	return false
}

// This is not Redfish-driven.  The context can be anything we choose, so we
// must standardize where controller xname ID in it and what the separator is.
// We capture additional tags as an array for now.
// We currently expect the xname to be first, however, we can optionally
// check (anyXName=true) every subfield, and choose the first that is a valid
// xname (for a controller)
func EventContextDecode(
	context string,
	anyXName bool,
) (xnameID string, subLabels []string) {
	subLabels = make([]string, 0, 1)
	// ":" is used after the field separator
	fields := strings.Split(context, ":")
	for i, field := range fields {
		setID := false
		if i == 0 || (xnameID == "" && anyXName == true) {
			normField := xnametypes.NormalizeHMSCompID(field)
			if xnametypes.IsHMSTypeController(xnametypes.GetHMSType(normField)) {
				xnameID = normField
				setID = true
			}
		}
		// If not xname ID, add to generic list
		if !setID {
			subLabels = append(subLabels, field)
		}
	}
	return
}

// This is a version of EventContextDecode that looks at both potential
// context fields, the Event and an invididual EventRecord.  We should
// only have one, but this hides the logic of what's what.
func GetEventIDAndLabels(eCtx, rCtx string) (id string, labels []string) {
	// First time, only check for xname ID of controller in expected place.
	eID, eLabels := EventContextDecode(eCtx, false)
	rID, rLabels := EventContextDecode(rCtx, false)
	if eID == "" && rID == "" {
		// No xname IDs, be less picky about where the xname ID is
		eID, eLabels = EventContextDecode(eCtx, true)
		rID, rLabels = EventContextDecode(rCtx, true)
	}
	if eID != "" {
		id = eID
		labels = eLabels
	} else if rID != "" {
		id = rID
		labels = rLabels
	}
	return id, labels
}

/////////////////////////////////////////////////////////////////////////////
// Generate Component state change, if needed.
/////////////////////////////////////////////////////////////////////////////

// Turn a processed RF event into a state change request that can be
// submitted
func (s *SmD) compUpdateFromRFEvent(pe *processedRFEvent) (*CompUpdate, error) {
	if pe == nil {
		return nil, ErrSmMsgNilProcRFE
	}
	action := s.GetEventActionParser(pe)
	if action == nil {
		return nil, nil
	}
	return action(s, pe)
}

/////////////////////////////////////////////////////////////////////////////
// EventActionParser - Lookup table and parsers
//
// This is where changes in logic due to new implementations and events go.
// we just
/////////////////////////////////////////////////////////////////////////////

// Handler prototype for the lookup function below
type EventActionParser func(*SmD, *processedRFEvent) (*CompUpdate, error)

// Lookups go from most general to most specific.  If a more general
// lookup is ambiguous,  that is the signal to keep trying with a
// more specific key.  Unfortunately, we can't compare functions, except with
// nil, we can't use special function names to describe the next step.  So
// calls that need disambiguation need to return nil for each ambiguous level.
//
// Level 0: Just MessageId, no version or Registry.
// Level 1: if just messageId nil, try again with MessageId:Registry
// Level 2: If messageId + reg also nil, try again with MessageId:Registry:vers
//
//	where vers is :1 for 1.0 or 1.0.1
//
// Level 3: Version is still not specific enough.  They again with
//
//	MessageId:Registry:vers, but this time include maj version, i.e. 1.1
var eventActionParserLookup = map[string]EventActionParser{
	"resourcepowerstatechanged":               nil,
	"resourcepowerstatechanged:resourceevent": ResourcePowerStateChangedParser,
	"resourcepowerstatechanged:crayalerts":    ResourcePowerStateChangedParser,
	"resourcepowerstatechanged:":              ResourcePowerStateChangedParser,
	"systempoweron":                           AlertSystemPowerOnParser,
	"systempoweroff":                          AlertSystemPowerOffParser,
	"alert":                                   AlertSystemPowerParser,
	"powerstatuschange":                       AlertSystemPowerParser,
	"serverpoweredon":                         AlertSystemPowerOnParser,
	"serverpoweredoff":                        AlertSystemPowerOffParser,
	"dcpoweron":                               FoxconnAlertSystemPowerOnParser,
	"dcpoweroff":                              FoxconnAlertSystemPowerOffParser,
}

// Gets the EventActionParser function for the processed event or returns
// noActionFound
func (s *SmD) GetEventActionParser(pe *processedRFEvent) EventActionParser {
	lookup := pe.MessageId

	// We may need to repeat the lookup a small number of times with
	// additional info to disambiguate.  This keeps the table much
	// more compact.
	for level := 1; level <= 3; level++ {
		action, ok := eventActionParserLookup[strings.ToLower(lookup)]
		if !ok {
			// No match at all, not even a nil pointer to keep trying.
			// No action for event.
			return nil
		}
		// Determine if another lookup is needed and include the requested key.
		if action == nil {
			switch level {
			case 1:
				lookup = pe.MessageId + ":" + pe.Registry
			case 2:
				vers, _ := rf.VersionFields(pe.RegVersion, ".", 1)
				lookup = pe.MessageId + ":" + pe.Registry + ":" + vers
			case 3:
				vers, incl := rf.VersionFields(pe.RegVersion, ".", 2)
				if incl != 2 {
					// Don't have this, leave it empty
					vers = ""
				}
				lookup = pe.MessageId + ":" + pe.Registry + ":" + vers
			}
		} else {
			// Match - return parser
			return action
		}
	}
	// Shouldn't get here, to many levels of nil function entries.  But
	// assume that means we didn't find it.
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// ResourceEvents - Right now just for Cray hardware.
/////////////////////////////////////////////////////////////////////////////

type ResourceOp string

const (
	ResourceOn      ResourceOp = "on"
	ResourceOff     ResourceOp = "off"
	ResourceUnknown ResourceOp = "unknown"
	ResourceOther   ResourceOp = "other"
)

// EventActionParser - ResourcePowerStateChanged - Cray addition to standard
//
//	ResourceEvent registry.  Need to look at payload to see
//	new state and component type.
func ResourcePowerStateChangedParser(s *SmD, pe *processedRFEvent) (*CompUpdate, error) {
	// Parse the arguments.  Arg1 should be the URI for the component,
	// and Arg2 the state.  But take them in either order.  And if there
	// is no URI, use the origin.
	uri := pe.Origin
	op := ResourceUnknown
	for _, arg := range pe.MessageArgs {
		if strings.HasPrefix(arg, "/") {
			uri = arg
		} else {
			switch strings.ToLower(strings.TrimSpace(arg)) {
			case string(ResourceOn):
				op = ResourceOn
			case string(ResourceOff):
				op = ResourceOff
			}
		}
	}
	// If the state is not off or on, ignore it.  PoweringOn/Off is
	// not used here.  In HSM, it means something is actively trying to
	// turn it on.  But at the Redfish level, we don't care until it is
	// actually on or off.
	if op == ResourceUnknown {
		return nil, ErrSmMsgIgnState
	} else if uri == "" {
		return nil, ErrSmMsgNoURI
	}
	xname, err := s.getIDForURI(pe.RfEndppointID, uri)
	if err != nil {
		return nil, err
	} else if xname == "" {
		s.Log(LOG_INFO, "ResourcePowerStateChangedParser(%s, %s): Not found.",
			pe.RfEndppointID, uri)
		return nil, ErrSmMsgNoID
	}
	// We already know the target, but we might need to affect other
	// components if it is a power off operation, depending on the
	// type of component affected, e.g. slot, etc.
	switch xnametypes.GetHMSType(pe.RfEndppointID) {
	case xnametypes.ChassisBMC:
		return powerStateCMM(s, pe, xname, op)
	case xnametypes.NodeBMC:
		return powerStateNC(s, pe, xname, op)
	case xnametypes.RouterBMC:
		return powerStateRC(s, pe, xname, op)
	case xnametypes.CabinetPDUController:
		return powerStateCabPDUController(s, pe, xname, op)
	default:
		return nil, ErrSmMsgBadID
	}
}

// Helper for ResourcePowerStateChangedParser - CMM changes
func powerStateCMM(s *SmD, pe *processedRFEvent,
	xname string,
	op ResourceOp,
) (*CompUpdate, error) {
	u := new(CompUpdate)
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	switch op {
	case ResourceOn:
		u.State = base.StateOn.String()
	case ResourceOff:
		u.State = base.StateOff.String()
	default:
		// Should never happen.
		return u, ErrSmMsgNoPowerState
	}
	switch xnametypes.GetHMSType(xname) {
	case xnametypes.ComputeModule:
		ids := generateNcChildIDs(s, xname, op)
		u.ComponentIDs = append(u.ComponentIDs, ids...)
	case xnametypes.RouterModule:
		ids := generateRcChildIDs(s, xname, op)
		u.ComponentIDs = append(u.ComponentIDs, ids...)
	}

	// Kick off rediscovery for any BMCs that are getting powered on.
	// This may fail at first if the BMC isn't ready yet but the
	// LastDiscoveryStatus will get changed to a failed state which will
	// cause a retry later.
	if u.State == base.StateOn.String() {
		for _, id := range u.ComponentIDs {
			if xnametypes.IsHMSTypeController(xnametypes.GetHMSType(id)) {
				rep, err := s.db.GetRFEndpointByID(id)
				if err != nil {
					s.Log(LOG_INFO, "powerStateCMM(): Lookup failure on %s: %s", id, err)
				} else if rep != nil {
					go s.discoverFromEndpoint(rep, 0, false)
				}
			}
		}
	}

	return u, nil
}

// Helper function for ResourcePowerStateChangedParser - nC changes
func powerStateNC(s *SmD, pe *processedRFEvent,
	xname string,
	op ResourceOp,
) (*CompUpdate, error) {
	u := new(CompUpdate)
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	switch op {
	case ResourceOn:
		u.State = base.StateOn.String()
		// Update hwinv for nodes
		if xnametypes.GetHMSType(xname) == xnametypes.Node {
			cep, ep, err := s.getCompEPInfo(xname)
			if err == nil {
				go s.doUpdateCompHWInv(cep, ep)
			}
		}
	case ResourceOff:
		u.State = base.StateOff.String()
	}
	return u, nil
}

// Helper for ResourcePowerStateChangedParser - sC/TOR changes
func powerStateRC(s *SmD, pe *processedRFEvent,
	xname string,
	op ResourceOp,
) (*CompUpdate, error) {
	u := new(CompUpdate)
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	switch op {
	case ResourceOn:
		u.State = base.StateOn.String()
	case ResourceOff:
		u.State = base.StateOff.String()
	}
	return u, nil
}

// Helper for ResourcePowerStateChangedParser - PDU (RTS) changes.
func powerStateCabPDUController(s *SmD, pe *processedRFEvent,
	xname string,
	op ResourceOp,
) (*CompUpdate, error) {
	u := new(CompUpdate)
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	switch op {
	case ResourceOn:
		u.State = base.StateOn.String()
	case ResourceOff:
		u.State = base.StateOff.String()
	}
	return u, nil
}

// The database will ignore any ids that don't exist, do them all.
// If ResourceOn only include up controllers/cards themselves, other
// If ResourceOff include all nodes and other subcomponents.
// This ASSUMES xname is a valid ComputeModule and Mountain slot.
func generateNcChildIDs(s *SmD, xname string, op ResourceOp) []string {
	maxNCs := 4
	ids := []string{}
	if op != ResourceOn && op != ResourceOff {
		// shouldn't happen
		return ids
	}
	for i := 0; i < maxNCs; i++ {
		// Get nth BMC for slot xname
		id := fmt.Sprintf("%sb%d", xname, i) // NodeBMC

		// Get children from database, syncing if necessary.
		children, err := s.getChildIDsForRfEP(id)
		if err != nil {
			if err != ErrSmMsgBadID {
				// Bad ID errors are expected here as we are talking about
				// a child controller.
				s.LogAlways("generateNcChildIDs(%s, %s): DB error: %s",
					xname, id, err)
			}
			continue
		}
		if op == ResourceOn {
			// Only turn the cards on, not the nodes, when slot comes up.
			for _, childID := range children {
				switch xnametypes.GetHMSType(childID) {
				case xnametypes.NodeEnclosure:
					fallthrough
				case xnametypes.NodeBMC:
					fallthrough
				case xnametypes.NodeBMCNic:
					ids = append(ids, childID)
				}
			}
		} else if op == ResourceOff {
			ids = append(ids, children...)
		}
	}
	return ids
}

// Generate all subcomponents for a Mountain router slot.
// The database will ignore any ids that don't exist, do them all.
// This ASSUMES xname is a valid RouterModule and Mountain slot.
func generateRcChildIDs(s *SmD, xname string, op ResourceOp) []string {
	maxControllers := 1
	ids := []string{}
	for i := 0; i < maxControllers; i++ {
		id := fmt.Sprintf("%sb%d", xname, i) // RouterBMC
		// Get children from database, syncing if necessary.
		children, err := s.getChildIDsForRfEP(id)
		if err != nil {
			if err != ErrSmMsgBadID {
				// Bad ID errors are expected here as we are talking about
				// a child controller.
				s.LogAlways("generateRcChildIDs(%s, %s): DB error: %s",
					xname, id, err)
			}
			continue
		}
		ids = append(ids, children...)
	}
	return ids
}

/////////////////////////////////////////////////////////////////////////////
// Intel BMC firmware & HPE iLo
/////////////////////////////////////////////////////////////////////////////

// EventActionParser - Alert, presumably from Intel BMC, that indicates
//
//	System (i.e. node) powered ON.
//	Id in OriginOfCondition (though likely single node).
func AlertSystemPowerOnParser(s *SmD, pe *processedRFEvent) (*CompUpdate, error) {
	u := new(CompUpdate)
	xname, err := s.getIDForURI(pe.RfEndppointID, pe.Origin)
	if err != nil {
		return nil, err
	} else if xname == "" {
		return nil, ErrSmMsgNoID
	}
	// Update hwinv for nodes
	if xnametypes.GetHMSType(xname) == xnametypes.Node {
		cep, ep, err := s.getCompEPInfo(xname)
		if err == nil {
			go s.doUpdateCompHWInv(cep, ep)
		}
	}
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	u.State = base.StateOn.String()
	return u, nil
}

// EventActionParser - Alert, presumably from Intel BMC, that indicates
//
//	System (i.e. node) powered OFF.
//	Id in OriginOfCondition (though likely single node).
func AlertSystemPowerOffParser(s *SmD, pe *processedRFEvent) (*CompUpdate, error) {
	u := new(CompUpdate)
	xname, err := s.getIDForURI(pe.RfEndppointID, pe.Origin)
	if err != nil {
		return nil, err
	} else if xname == "" {
		return nil, ErrSmMsgNoID
	}
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	u.State = base.StateOff.String()
	return u, nil
}

/////////////////////////////////////////////////////////////////////////////
// Gigabyte BMC firmware
/////////////////////////////////////////////////////////////////////////////

// AlertSystemPowerParser - Alert, presumably from Gigabyte BMC, that indicates
//
//	System (i.e. node) powered ON or OFF.
//	Id in OriginOfCondition (though likely single node).
func AlertSystemPowerParser(s *SmD, pe *processedRFEvent) (*CompUpdate, error) {
	var (
		cep *sm.ComponentEndpoint
		ep  *rf.RedfishEP
	)

	// Parse the arguments. Arg1 should be the URI for the component,
	// and Arg2 the state. But take them in either order. And if there
	// is no URI, use the origin. If there is no State reachout to the
	// component via redfish to get it.
	uri := pe.Origin
	op := ResourceUnknown
	for _, arg := range pe.MessageArgs {
		if strings.HasPrefix(arg, "/") {
			uri = arg
		} else {
			switch strings.ToLower(strings.TrimSpace(arg)) {
			case string(ResourceOn):
				op = ResourceOn
			case string(ResourceOff):
				op = ResourceOff
			default:
				op = ResourceOther
			}
		}
	}
	// If the state is not off or on, ignore it.  PoweringOn/Off is
	// not used here.  In HSM, it means something is actively trying to
	// turn it on.  But at the Redfish level, we don't care until it is
	// actually on or off. If there isn't a state in the msg, we'll go
	// ping it.
	if op == ResourceOther {
		return nil, ErrSmMsgIgnState
	} else if uri == "" {
		return nil, ErrSmMsgNoURI
	}
	xname, err := s.getIDForURI(pe.RfEndppointID, uri)
	if err != nil {
		return nil, err
	} else if xname == "" {
		s.Log(LOG_INFO, "AlertSystemPowerParser(%s, %s): Not found.",
			pe.RfEndppointID, uri)
		return nil, ErrSmMsgNoID
	}

	if op == ResourceUnknown || op == ResourceOn {
		cep, ep, err = s.getCompEPInfo(xname)
		if err != nil {
			return nil, err
		}
	}

	// There was no state in the redfish event. We have to go get it.
	if op == ResourceUnknown {
		ps, err := s.getCompEPState(cep, ep)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(ps) {
		case string(ResourceOn):
			op = ResourceOn
		case string(ResourceOff):
			op = ResourceOff
		default:
			return nil, ErrSmMsgIgnState
		}
	}
	u := new(CompUpdate)
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	switch op {
	case ResourceOn:
		u.State = base.StateOn.String()
		if xnametypes.GetHMSType(xname) == xnametypes.Node {
			go s.doUpdateCompHWInv(cep, ep)
		}
	case ResourceOff:
		u.State = base.StateOff.String()
	}
	return u, nil
}

// getCompEPInfo - This gathers the existing ComponentEndpoint and credentials
//
//	present in either the secure store or the database (if
//	secure store is not enabled) for the xname.
func (s *SmD) getCompEPInfo(xname string) (*sm.ComponentEndpoint, *rf.RedfishEP, error) {
	var (
		user string
		pw   string
	)
	cep, err := s.getCompEPbyID(xname)
	if err != nil {
		return nil, nil, err
	} else if cep == nil {
		s.Log(LOG_INFO, "GetCompEPInfo(%s): ComponentEndpoint not found.", xname)
		return nil, nil, ErrSmMsgNoEP
	}
	if s.readVault {
		// Read component endpoint credentials from the secure store.
		cred, err := s.ccs.GetCompCred(xname)
		if err != nil {
			if strings.Contains(err.Error(), "Code: 404") {
				// Ignore if there are no credentials in vault for the component
				s.Log(LOG_INFO, "GetCompEPInfo(%s): No credentials in Vault - %s",
					xname, err)
				return nil, nil, ErrSmMsgNoCreds
			} else {
				s.Log(LOG_INFO, "GetCompEPInfo(%s): Failed to get credentials from Vault - %s",
					xname, err)
				return nil, nil, ErrSmMsgCredsStore
			}
		}
		user = cred.Username
		pw = cred.Password
	} else {
		// get credentials from the database redfishendpoints
		rep, err := s.db.GetRFEndpointByID(cep.RfEndpointID)
		if err != nil {
			s.Log(LOG_INFO, "GetCompEPInfo(%s): Lookup failure: %s", xname, err)
			return nil, nil, ErrSmMsgCredsDB
		}
		if rep == nil {
			s.Log(LOG_INFO, "GetCompEPInfo(%s): No credentials in the database for %s",
				xname, cep.RfEndpointID)
			return nil, nil, ErrSmMsgNoCreds
		}
		user = rep.User
		pw = rep.Password
	}
	// Minimally populate a redfish description struct
	rfEPType := xnametypes.GetHMSType(cep.RfEndpointID)
	epDesc := rf.RedfishEPDescription{
		ID:       cep.RfEndpointID,
		Type:     rfEPType.String(),
		FQDN:     cep.RfEndpointFQDN,
		User:     user,
		Password: pw,
	}
	// Make a RedfishEP struct to make our redfish call against.
	ep, err := rf.NewRedfishEp(&epDesc)
	return cep, ep, err
}

// getCompEPState - Get the redfish powerstate of a component, presumably a
//
//	Node. This uses the existing ComponentEndpoint and
//	credentials as gathered by getCompEPInfo() for the xname to
//	check the power state of the component via redfish.
func (s *SmD) getCompEPState(cep *sm.ComponentEndpoint, ep *rf.RedfishEP) (string, error) {
	if cep == nil || ep == nil {
		return "", ErrSmMsgNoEP
	}
	// Make a RedfishEP struct to make our redfish call against.
	url := cep.RfEndpointFQDN + cep.OdataID

	rfJSON, err := ep.GETRelative(cep.OdataID)
	if err != nil {
		s.Log(LOG_INFO, "GetCompEPState(%s): redfish call failed: %s: %s",
			cep.ID, url, err)
		return "", err
	}
	// Parse the data that we received
	rfData := new(rf.ComputerSystem)
	if err := json.Unmarshal(rfJSON, &rfData); err != nil {
		if rf.IsUnmarshalTypeError(err) {
			s.Log(LOG_INFO, "GetCompEPState(%s): bad field(s) skipped: %s: %s",
				cep.ID, url, err)
		} else {
			s.Log(LOG_INFO, "GetCompEPState(%s): json decode failed: %s: %s",
				cep.ID, url, err)
			return "", err
		}
	}
	return rfData.PowerState, nil
}

// doUpdateCompHWInv - Update the hwinv for a component, presumably a Node.
//
//	This uses the existing ComponentEndpoint and credentials
//	as gathered by getCompEPInfo() for the xname to update
//	the HW Inventory data for the Component with info
//	gathered.
func (s *SmD) doUpdateCompHWInv(cep *sm.ComponentEndpoint, ep *rf.RedfishEP) error {
	if cep == nil || ep == nil {
		return ErrSmMsgNoEP
	}
	// Update the node info under the redfish endpoint
	if xnametypes.GetHMSType(cep.ID) == xnametypes.Node {
		// Read from redfish
		status := ep.GetSystems()
		if status != rf.HTTPsGetOk {
			s.Log(LOG_INFO, "doUpdateCompHWInv(%s): Failed to get system info: %s",
				cep.ID, status)
			return ErrSmMsgRFFail
		}
		// Process redfish data
		if err := ep.VerifySystems(); err != nil {
			s.Log(LOG_INFO, "doUpdateCompHWInv(%s): Systems verification failed: %s",
				cep.ID, err)
			return err
		}
		// Discover hardware inventory from redfish data
		hwlocs, err := s.DiscoverHWInvByLocArray(ep)
		if err != nil {
			if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
				// Non-fatal, one or more components wasn't supported.  Likely to
				// recur if discovery re-run.
				s.Log(LOG_INFO, "DiscoverHWInvByLocArray(%s): One or more: %s",
					cep.ID, err)
			} else {
				s.Log(LOG_INFO, "DiscoverHWInvByLocArray(%s): Fatal error storing: %s",
					cep.ID, err)
				return err
			}
		}
		// Only insert the discovered locs for the node we're concerned about.
		// This will cut down on database concurrency issues.
		nodehls := make([]*sm.HWInvByLoc, 0, 1)
		for _, hl := range hwlocs {
			if strings.Contains(hl.ID, cep.ID) {
				nodehls = append(nodehls, hl)
			}
		}
		err = s.db.InsertHWInvByLocs(nodehls)
		if err != nil {
			s.Log(LOG_INFO, "doUpdateCompHWInv(%s): Failed to update hwinv: %s",
				cep.ID, err)
		}
		// Generate hardware history
		err = s.GenerateHWInvHist(nodehls)
		if err != nil {
			s.Log(LOG_INFO, "doUpdateCompHWInv(%s): Failed to update hwinv history: %s",
				cep.ID, err)
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// Foxconn Paradise OpenBmc firmware
/////////////////////////////////////////////////////////////////////////////

// doUpdateCompFoxconn - Update both the hwinv and the redfish system endpoint data.
//
//      For Foxconn Paradise hardware we want to do more than just call doUpdateCompHWInv().
//      If the last discover was run with the node powered off, the PowerURL and
//      PowerControl system information may not have been updated due to a BMC fw
//      bug (see PRDIS-198).  doUpdateCompHWInv() would indeed update this system
//      information correctly in the endpoint 'ep' since the node is now powered on, but
//      it will not push it into the database.
//
func (s *SmD) doUpdateCompFoxconn(cep *sm.ComponentEndpoint, ep *rf.RedfishEP) error {
	// First update the hardware inventory.  This also updates system info in the
	// component endpoint from the ProcesorModule_0 chassis
	err := s.doUpdateCompHWInv(cep, ep)
	if err != nil {
		return err
	}

	// Prep for writing to the database
	sysceps := new(sm.ComponentEndpointArray)
	for _, sysEP := range ep.Systems.OIDs {
		syscep := s.DiscoverCompEndpointSystem(sysEP)
		if syscep != nil {
			sysceps.ComponentEndpoints = append(sysceps.ComponentEndpoints, syscep)
		}
	}

	// Now push into the database
	err = s.db.UpsertCompEndpoints(sysceps)
	if err != nil {
		s.Log(LOG_INFO, "doUpdateCompFoxconn(%s): Failed to update system component endpoints: %s",
			cep.ID, err)
	}

	return nil
}

// EventActionParser - Alert, presumably from Foxconn Paradise BMC
//
//	The OriginOfXondition (pe.Origin) will likely not be set. Assume n0 in that case
func FoxconnAlertSystemPowerOnParser(s *SmD, pe *processedRFEvent) (*CompUpdate, error) {
	u := new(CompUpdate)
	xname := fmt.Sprintf("%sn0", pe.RfEndppointID)
	if pe.Origin != "" {
		xname, err := s.getIDForURI(pe.RfEndppointID, pe.Origin)
		if err != nil {
			return nil, err
		} else if xname == "" {
			return nil, ErrSmMsgNoID
		}
	}
	// Update hwinv for nodes
	if xnametypes.GetHMSType(xname) == xnametypes.Node {
		cep, ep, err := s.getCompEPInfo(xname)
		if err == nil {
			go s.doUpdateCompFoxconn(cep, ep)
		}
	}
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	u.State = base.StateOn.String()
	return u, nil
}

// EventActionParser - Alert, presumably from Foxconn Paradise BMC
//
//	The OriginOfXondition (pe.Origin) will likely not be set. Assume n0 in that case
func FoxconnAlertSystemPowerOffParser(s *SmD, pe *processedRFEvent) (*CompUpdate, error) {
	u := new(CompUpdate)
	xname := fmt.Sprintf("%sn0", pe.RfEndppointID)
	if pe.Origin != "" {
		xname, err := s.getIDForURI(pe.RfEndppointID, pe.Origin)
		if err != nil {
			return nil, err
		} else if xname == "" {
			return nil, ErrSmMsgNoID
		}
	}
	u.ComponentIDs = append(u.ComponentIDs, xname)
	u.UpdateType = StateDataUpdate.String()
	u.State = base.StateOff.String()
	return u, nil
}

/////////////////////////////////////////////////////////////////////////////
//
// Cached DB lookups
//
/////////////////////////////////////////////////////////////////////////////

// Prototype function for creating a type-specific SyncMap.
type SMapBuildFunc func(*SyncMap) error

// This is a Syncronized cache with two levels.  A direct lookup of
// objects by their primary key, and an indirect lookup that allows a
// primary key to be looked up via some other value that can be
// reduced to a string so we can look it up quickly via a map.
type SyncMap struct {
	rwLock    sync.RWMutex
	valStore  map[string]interface{}
	syncNum   int
	keyLookup map[string]string
	f         SMapBuildFunc
}

// Creates a new SyncMap with the given function for creating the mapping,
// i.e. specific to a particular data type.  This does not actually create
// the mapping.  This will be done when TrySync is first called, typically
// when the first read fails and we want to make sure we try it again with
// synched data.
func NewSyncMap(f SMapBuildFunc) *SyncMap {
	smap := new(SyncMap)
	smap.valStore = make(map[string]interface{})
	smap.syncNum = 0
	smap.keyLookup = make(map[string]string)
	smap.f = f

	return smap
}

// Read the current update number after making sure we are ordered
// before or after a sync.
func (smap *SyncMap) GetSyncNum() int {
	smap.rwLock.RLock()
	defer smap.rwLock.RUnlock()
	return smap.syncNum
}

// Based on last read, try to update the mapping if it has not already
// been updated since that read completed (returning syncNum).
// Should return now-current syncNum (>= input value) or the original
// number plus an error.
//
// TODO: Call this periodically so missing entries are purged.  This is
// less critical since state changes will be ignored if component is
// empty.  This also leads to fewer syncs.
func (smap *SyncMap) TrySync(syncNum int) (bool, int, error) {
	smap.rwLock.Lock()
	defer smap.rwLock.Unlock()
	if smap.syncNum != syncNum {
		return false, smap.syncNum, nil
	}
	err := smap.f(smap)
	if err != nil {
		return false, smap.syncNum, err
	}
	smap.syncNum += 1
	return true, smap.syncNum, nil
}

// Get the primary key for a stored object based on a lookup string
func (smap *SyncMap) LookupKey(queryKey string) (string, int) {
	smap.rwLock.RLock()
	defer smap.rwLock.RUnlock()
	value, ok := smap.keyLookup[queryKey]
	if !ok {
		return "", smap.syncNum
	}
	return value, smap.syncNum
}

// Get the stored data type by it's primary key
func (smap *SyncMap) LookupValue(key string) (interface{}, int) {
	smap.rwLock.RLock()
	defer smap.rwLock.RUnlock()
	value, ok := smap.valStore[key]
	if !ok {
		return nil, smap.syncNum
	}
	return value, smap.syncNum
}

/////////////////////////////////////////////////////////////////////////////
// Cached DB lookups - ComponentEndpoints
/////////////////////////////////////////////////////////////////////////////

const CompEPQueryKeyChildren = "children"

const CompEPSyncRetries = 3

// This creates a ComponentEndpoint mapping, with keys that can be
// looked up by the parent RedfishEndpoint plus the Redfish URI.
// It is used for reverse lookups in events, at least where the
// events contain the proper origin information as the OriginOfCondition
// or as one of the MessageArgs.
//
// Other mappings are possible where needed, but we really want to avoid
// assuming the xname, as we don't necessarily want to depend on the
// index starting at zero or one.  This is even more important for
// things like (Redfish) Chassis, where the name is often ambiguous
// across different implementations.
func ComponentEndpointSMap(s *SmD) SMapBuildFunc {
	return func(smap *SyncMap) error {
		ceps, err := s.db.GetCompEndpointsAll()
		if err != nil {
			return err
		}
		smap.valStore = make(map[string]interface{})
		smap.keyLookup = make(map[string]string)
		for _, cep := range ceps {
			// Add to ComponentEndpoint array, keyed on each member's xname ID
			smap.valStore[cep.ID] = cep

			// Query Key1: Add RedFishEndpoint:URI as key to
			// ComponentEndpoint ID needed to lookup value.
			key := cep.RfEndpointID + ":" + cep.OdataID
			smap.keyLookup[strings.ToLower(key)] = cep.ID

			// Query Key2: Add RedfishEndpoint:children as key to csv list
			// of ComponentEndpoints with that ID as their RfEndpointID.
			// Basically a csv list of child xnames for a controller.
			key2 := cep.RfEndpointID + ":" + CompEPQueryKeyChildren
			current, ok := smap.keyLookup[key2]
			if !ok {
				// New list
				smap.keyLookup[key2] = cep.ID
			} else {
				// Append to list
				smap.keyLookup[key2] = current + "," + cep.ID
			}
		}
		return nil
	}
}

// This avoids a thundering herd when creating the map for the first time,
// but is not especially efficient when there are many failed lookups.
// In theory, we shouldn't see these, because the subscriptions should be
// created in response to State Manager creating the endpoints, but
// there is a degenerate case if something goes wrong.
func (s *SmD) getIDForURI(epID, URI string) (string, error) {
	var found, didUpdate bool = false, false
	var err error

	// Put together key for this endpoint plus the URI in the event.
	uriKey := strings.ToLower(epID + ":" + URI)

	// Lookup the key in the cache.
	id, snum := s.smapCompEP.LookupKey(uriKey)
	for i := 0; i < 2; i++ {
		if id == "" {
			// Not found, try looking up directly in database
			if !didUpdate {
				// If found, make sure a sync occurs so it will be in the
				// cache.
				found, didUpdate, err = s.checkSyncCompEP(epID, snum)
				if err != nil {
					return "", err
				}
				// Not in DB
				if !found {
					return "", nil
				}
				// Found in DB, try the lookup from the cache again.
				id, snum = s.smapCompEP.LookupKey(uriKey)

				// NOTE: If this lookup failed, but found was true after sync,
				// try a second checkSync if didUpdate==false, because
				// someone else did the last sync, and it may have been
				// too early.
			}
		} else {
			// Found
			break
		}
	}
	return id, nil
}

// Get the stored data type (in this case ComponentEndpoint) by it's xname ID
func (s *SmD) getCompEPbyID(epID string) (*sm.ComponentEndpoint, error) {
	var found, didUpdate bool = false, false
	var err error

	// We look up a generic interface, but if found it should always be a
	// componentEndpoint (if we set things up properly.
	cepi, snum := s.smapCompEP.LookupValue(epID)
	for i := 0; i < 2; i++ {
		if cepi == nil {
			// Not found, try looking up directly in database
			if !didUpdate {
				// If found, make sure a sync occurs so it will be in the
				// cache.
				found, didUpdate, err = s.checkSyncCompEP(epID, snum)
				if err != nil {
					return nil, err
				}
				// Not in DB
				if !found {
					return nil, nil
				}
				// Found in DB, try the lookup from the cache again
				// (up to two retries).
				cepi, snum = s.smapCompEP.LookupValue(epID)

				// NOTE: If this lookup failed, but found was true after sync,
				// try a second checkSync if didUpdate==false, because
				// someone else did the last sync, and it may have been
				// too early.
			}
		} else {
			// Found
			break
		}
	}
	if cepi != nil {
		cep, ok := cepi.(*sm.ComponentEndpoint)
		if !ok {
			// Shouldn't happen.  Map was set up wrong if it contains
			// another value type.
			return nil, ErrSmMsgBadCached
		}
		// Non-nil and interface value is a valid *ComponentEndpoint
		return cep, nil
	}
	// Got nil interface, not found even after sync.
	// No error just not there.
	return nil, nil
}

// Look up the child xname ID list for an endpoint (CMM, nC, BMC, etc).
// These are ComponentEndpoints IDs, so they are only created post-discovery.
// If our own entry exists, our children should too.  Does not capture
// components that are not first-order children (i.e. the nC child of a
// CMM slot).
func (s *SmD) getChildIDsForRfEP(epID string) ([]string, error) {
	var found, didUpdate bool = false, false
	var err error

	// Get key to lookup children for this endpoint.
	childrenKey := epID + ":" + CompEPQueryKeyChildren

	idsStr, snum := s.smapCompEP.LookupKey(childrenKey)
	for i := 0; i < 2; i++ {
		if idsStr == "" {
			// Not found, try looking up directly in database
			if !didUpdate {
				// If found, make sure a sync occurs so it will be in the
				// cache.
				found, didUpdate, err = s.checkSyncCompEP(epID, snum)
				if err != nil {
					return []string{}, err
				}
				// Not in DB
				if !found {
					return []string{}, nil
				}
				// Found in DB, try the lookup from the cache again
				// (up to two retries).
				idsStr, snum = s.smapCompEP.LookupKey(childrenKey)

				// NOTE: If this lookup failed, but found was true after sync,
				// try a second checkSync if didUpdate==false, because
				// someone else did the last sync, and it may have been
				// too early.
			}
		} else {
			// Found
			break
		}
	}
	if idsStr == "" {
		// Not found after sync.  Unexpected if we received an event
		// from this endpoint.
		return []string{}, ErrSmMsgBadID
	}
	// Convert csv string into ids and return array.
	ids := strings.Split(idsStr, ",")
	s.Log(LOG_DEBUG, "getChildIDsForRfEP(%s): [%s]", epID, ids)
	return ids, nil
}

// This is done only AFTER a failed cached read.  It will do a single lookup
// for xname, and if it succeeds, it will sync the cache.  This avoids
// syncs that do not have at least one piece of new data and ensures a
// cheap operation rather than an expensive one occurs on a miss.  We
// don't expect a lot of misses, but this eases the pain.
//
// Input: xname is search ID, num is syncNum for Sync.
//
// Returns found=true if found, and didUpdate=true if the sync actually took
// place.
//
// If found=true and didUpdate=false, then a second attempt should be made
// if the entry is not found in the SyncMap, because we don't know for sure
// whether the map was updated after we found the entry in the DB or before.
func (s *SmD) checkSyncCompEP(xname string, snum int) (found, didUpdate bool, err error) {
	found = false
	didUpdate = false
	var ids []string

	for i := 0; i < CompEPSyncRetries; i++ {
		if xnametypes.GetHMSType(xname) == xnametypes.CabinetPDUController {
			// CabinentPDUControllers (xXmM) don't have ComponentEndpoint
			// entries. See if we've discovered any children.
			ids, err = s.db.GetCompEndpointIDs(hmsds.CE_RfEPs([]string{xname}))
		} else {
			ids, err = s.db.GetCompEndpointIDs(hmsds.CE_ID(xname))
		}
		if err == nil {
			if len(ids) == 0 {
				// Not found
				return
			} else {
				// found, do sync
				found = true
				break
			}
		} else if err != nil && base.IsHMSError(err) {
			// Driver error, won't work if retried.
			return
		} else {
			// DB error
			s.Log(LOG_INFO, "DB error: checkSyncCompEP(%s/%d): get: %s",
				xname, snum, err)
		}
	}
	// Maxed out retries
	if err != nil {
		return
	}
	// Ensure a sync has been done.  Doesn't matter who did it, as long
	// as it got done, and try again.
	for i := 0; i < CompEPSyncRetries; i++ {
		didUpdate, _, err = s.smapCompEP.TrySync(snum)
		if err == nil {
			break
		}
		s.Log(LOG_INFO, "DB error: checkSyncCompEP(%s/%d): Sync: %s",
			xname, snum, err)
	}
	return
}
