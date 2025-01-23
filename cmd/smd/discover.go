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
	"sync"

	base "github.com/Cray-HPE/hms-base"
	compcreds "github.com/Cray-HPE/hms-compcredentials"
	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
)

// When we discover a Redfish Endpoint, the data retrieved is processed
// in the rf package to associate it with basic data needed to place it
// within the system and extra HMS-level metadata.
// We do this discovery and then use the resulting data to create HMS-level
// structures and place them in the database.
//
// Args:
//
//	eps is a set of RedfishEndpoints retrieved from the database.
//	id is the id of the DiscoveryStatus object to write status to.
func (s *SmD) discoverFromEndpoints(eps []*sm.RedfishEndpoint, id uint, update, force bool) {
	idsFiltered := make([]string, 0, len(eps))
	for _, ep := range eps {
		if update && !ep.RediscOnUpdate {
			s.LogAlways("Skipping discovery for %s since !RediscoverOnUpdate",
				ep.ID)
			continue
		}
		if !ep.Enabled {
			s.LogAlways("Skipping discovery for %s since !Enabled",
				ep.ID)
			continue
		}
		idsFiltered = append(idsFiltered, base.VerifyNormalizeCompID(ep.ID))
	}
	// This should not fail in practice unless eps have not been inserted yet
	// This will "lock" the LastStatus to in-progress so it can't be started
	// twice.
	discEPs, err := s.db.UpdateRFEndpointForDiscover(idsFiltered, force)
	if err != nil {
		s.lg.Printf("Discovery: UpdateRFEndpointForDiscover() returned %s", err)
		return
	}
	if len(discEPs) != len(eps) {
		s.LogAlways("%d/%d endpoints are already being discovered and will "+
			"be skipped (not forced)",
			len(eps)-len(discEPs), len(eps))
	}

	// Create RedfishEPDescription array from the raw data
	rfEpds := new(rf.RedfishEPDescriptions)
	numEPs := 0
	for _, ep := range discEPs {
		rfEpd := &ep.RedfishEPDescription
		rfEpds.RfEPDescriptions = append(rfEpds.RfEPDescriptions, *rfEpd)
		numEPs = numEPs + 1
	}
	if numEPs == 0 {
		return
	}

	// Create an array of expanded objects that will contain the
	// discovered info.
	rfEps, err := rf.NewRedfishEps(rfEpds)
	if err != nil {
		// This shouldn't happen as it means an entry was not created
		// correctly.
		s.LogAlways("%d/%d endpoints are invalid and will be skipped: %s",
			(numEPs - rfEps.Num), numEPs, err)
	}

	// Write that discovery has started.
	stat := new(sm.DiscoveryStatus)
	stat.Status = sm.DiscInProgress
	stat.ID = id
	err = s.db.UpsertDiscoveryStatus(stat)
	if err != nil {
		s.lg.Printf("UpsertDiscoveryStatus start: %s", err)
	}

	var wGrp sync.WaitGroup
	for _, rfEp := range rfEps.IDs {
		wGrp.Add(1)
		// Start each endpoint as a separate thread
		go func(e *rf.RedfishEP) {
			defer wGrp.Done()
			s.doDiscovery(e)
		}(rfEp)
	}
	wGrp.Wait()

	// Write discovery status - we're done.
	stat.Status = sm.DiscComplete
	err = s.db.UpsertDiscoveryStatus(stat)
	if err != nil {
		s.lg.Printf("UpsertDiscoveryStatus end: %s", err)
	}
}

// Single-endpoint version of the above.
//
// Args:
//
//	ep is a single RedfishEndpoint retrieved from the database.
//	id is the id of the DiscoveryStatus object to write status to.
func (s *SmD) discoverFromEndpoint(ep *sm.RedfishEndpoint, id uint, force bool) {
	if !ep.RediscOnUpdate {
		s.LogAlways("Skipping discovery for %s: !RediscoverOnUpdate", ep.ID)
		return
	}
	if !ep.Enabled {
		s.LogAlways("Skipping discovery for %s since !Enabled", ep.ID)
		return
	}
	// This will "lock" the LastStatus to in-progress so it can't be started
	// twice.
	discEPs, err := s.db.UpdateRFEndpointForDiscover([]string{ep.ID}, force)
	if err != nil {
		s.lg.Printf("Discovery: UpdateRFEndpointForDiscover() returned %s", err)
		return
	} else if len(discEPs) == 0 {
		s.lg.Printf("Discovery: already in progress for %s", ep.ID)
		return
	}
	rfEP, err := rf.NewRedfishEp(&discEPs[0].RedfishEPDescription)
	if err != nil {
		// This shouldn't happen as it means an entry was not created
		// correctly.
		s.LogAlways("Endpoint is invalid and will be skipped")
	}

	// Write that discovery has started.
	stat := new(sm.DiscoveryStatus)
	stat.Status = sm.DiscInProgress
	stat.ID = id
	err = s.db.UpsertDiscoveryStatus(stat)
	if err != nil {
		s.lg.Printf("UpsertDiscoveryStatus start: %s", err)
	}

	s.doDiscovery(rfEP)

	// Write discovery status - we're done.
	stat.Status = sm.DiscComplete
	err = s.db.UpsertDiscoveryStatus(stat)
	if err != nil {
		s.lg.Printf("UpsertDiscoveryStatus end: %s", err)
	}
}

func (s *SmD) doDiscovery(rfEP *rf.RedfishEP) {

	// Add the xname to the list of discovery jobs for this HSM instance to periodically update.
	s.discoveryMapAdd(rfEP.ID)
	// Get redfish endpoint credentials from Vault
	if s.readVault {
		cred, err := s.ccs.GetCompCred(rfEP.ID)
		if err != nil {
			// Ignore we'll let it naturally fail without credentials later.
			s.LogAlways("Warning: Failed to get credentials from Vault for %s - %s", rfEP.ID, err)
		} else {
			// Don't read empty credentials
			if len(cred.Password) > 0 {
				rfEP.User = cred.Username
				rfEP.Password = cred.Password
			}
		}
	}

	// Do the actual discovery, including contacting the remote endpoint.
	rfEP.GetRootInfo()

	// Create/update HMS-level components from the retrieved discovery data
	// from Redfish.  This also inserts the data into the database.
	s.updateFromRfEndpoint(rfEP)
}

// Back end that writes one RedfishEndpoint's worth of structs to the DB
// provided they can be generated properly from the data we get from the
// RedfishEndpoint.
//
// Args:
//
//	rfEP - A discovered rf.RedfishEP, created from a RedfishEndpoint
//	       and then queried via gets to the specified destination.
func (s *SmD) updateFromRfEndpoint(rfEP *rf.RedfishEP) error {
	ep := sm.NewRedfishEndpoint(&rfEP.RedfishEPDescription)
	var savedErr error = nil
	var savedPw string
	var savedUn string

	// Check if children should be updated.
	if ep.DiscInfo.LastStatus == rf.EndpointTypeNotSupported ||
		ep.DiscInfo.LastStatus == rf.EndpointNotEnabled {
		//
		// Update endpoint only to reflect being skipped.
		//
		s.LogAlways("Discover of RedfishEndpoint %s skipped: %s",
			ep.ID, ep.DiscInfo.LastStatus)
		if s.readVault {
			ep.Password = ""
		}
		s.discoveryMapRemove(ep.ID)
		_, err := s.db.UpdateRFEndpoint(ep)
		return err
	} else if ep.DiscInfo.LastStatus != rf.DiscoverOK {
		s.LogAlways("Discover of RedfishEndpoint %s failed: %s",
			ep.ID, ep.DiscInfo.LastStatus)
		if s.readVault {
			ep.Password = ""
		}
		s.discoveryMapRemove(ep.ID)
		// Update endpoint only to reflect failed state.
		_, err := s.db.UpdateRFEndpoint(ep)
		return err
	}
	// Add/update component endpoints
	ceps, err := s.DiscoverComponentEndpointArray(rfEP)
	if err != nil {
		// These error types shouldn't happen, but may fail every time
		// so better to skip them and store the remaining, valid components.
		if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
			s.LogAlways("DiscoverComponentEndpointArray(%s): One or more: %s",
				rfEP.ID, err)
		} else {
			s.LogAlways("DiscoverComponentEndpointArray(%s): Fatal storing: %s",
				rfEP.ID, err)
			ep.DiscInfo.LastStatus = rf.UnexpectedErrorPreStore
			savedErr = err
		}
	}
	//Add/update component ethernet interface
	ceis := s.DiscoverCompEthInterfaceArray(ep, ceps)
	// Add/update service endpoints
	seps := s.DiscoverServiceEndpointArray(rfEP)
	// Add/update Hardware Inventory (FRU info, etc.) entries
	hwlocs, err := s.DiscoverHWInvByLocArray(rfEP)
	if err != nil {
		if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
			// Non-fatal, one or more components wasn't supported.  Likely to
			// recur if discovery re-run.
			s.LogAlways("DiscoverHWInvByLocArray(%s): One or more: %s",
				rfEP.ID, err)
		} else {
			s.LogAlways("DiscoverHWInvByLocArray(%s): Fatal error storing: %s",
				rfEP.ID, err)
			ep.DiscInfo.LastStatus = rf.UnexpectedErrorPreStore
			savedErr = err
		}
	}
	// Add HMS component entries (NID, state, role, etc.)
	comps, err := s.DiscoverComponentArray(rfEP)
	if err != nil {
		if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
			// Non-fatal, one or more components wasn't supported.  Likely to
			// recur if discovery re-run.
			s.LogAlways("DiscoverComponentArray(%s): One or more: %s",
				rfEP.ID, err)
		} else {
			s.LogAlways("DiscoverComponentArray(%s): Fatal error storing: %s",
				rfEP.ID, err)
			ep.DiscInfo.LastStatus = rf.UnexpectedErrorPreStore
			savedErr = err
		}
	}
	// Unrecoverable error - just save errored state for endpoint.
	if ep.DiscInfo.LastStatus != rf.DiscoverOK {
		if s.readVault {
			ep.Password = ""
		}
		s.discoveryMapRemove(ep.ID)
		_, err = s.db.UpdateAllForRFEndpoint(ep, nil, nil, nil, nil, nil)
		if err == nil {
			// Return initial reason for failure.
			return savedErr
		} else {
			return err
		}
	}

	if s.readVault || s.writeVault {
		savedPw = ep.Password
		savedUn = ep.User
		if s.readVault {
			// Empty the password so it doesn't get stored in the database.
			ep.Password = ""
		}
	}

	// Get hartbeating status from HBTD to make sure nodes previously set to 'Ready'
	// go back to 'Ready' if they are still heartbeating.
	if s.hbtd != nil && comps != nil && len(comps.Components) > 0 {
		compMap := make(map[string]*base.Component)
		compList := make([]string, 0, 1)
		for _, comp := range comps.Components {
			if comp.Type == base.Node.String() && comp.State == base.StateOn.String() {
				compMap[comp.ID] = comp
				compList = append(compList, comp.ID)
			}
		}
		results, err := s.hbtd.GetHeartbeatStatus(compList)
		if err != nil {
			s.LogAlways("GetHeartbeatStatus(): Could not retrieve heartbeat status: %s", err)
		} else {
			for _, stat := range results {
				comp, ok := compMap[stat.XName]
				if ok && stat.Heartbeating {
					comp.State = base.StateReady.String()
				}
			}
		}
	}

	s.discoveryMapRemove(ep.ID)
	// Data looks good - store it
	discoveredComps, err := s.db.UpdateAllForRFEndpoint(ep, ceps, hwlocs, comps, seps, ceis)
	if err != nil {
		// Unexpected error storing endpoint's data.
		s.LogAlways("UpdateAllForRFEndpoint(%s): Fatal error storing: %s",
			rfEP.ID, err)
		// Try to update just the endpoint to store this failed status.
		ep.DiscInfo.LastStatus = rf.StoreFailed
		savedErr = err
		_, err = s.db.UpdateRFEndpoint(ep)
		if err != nil {
			s.LogAlways("UpdateRFEndpoint(%s): Second fatal error storing: %s",
				rfEP.ID, err)
		}
		return savedErr
	}
	if discoveredComps != nil {
		scnMap := make(map[string][]string)
		// Send a SCN for each state for all of the new components and components that have updated states.
		for _, comp := range *discoveredComps {
			if _, ok := scnMap[comp.State]; !ok {
				// Make a SCN for each state
				scnMap[comp.State] = make([]string, 0, 1)
			}
			scnMap[comp.State] = append(scnMap[comp.State], comp.ID)
		}
		for state, ids := range scnMap {
			data := base.Component{State: state}
			scn := NewJobSCN(ids, data, s)
			s.wp.Queue(scn)
		}
	}
	// Store Credentials in Vault for the discovered componentEndpoints. This
	// is done if either readVault or writeVault is true because HSM is the one
	// discovering these components and thus, if Vault is being used, HSM must
	// be the one to populate the entries.
	if s.readVault || s.writeVault {
		// Don't store empty credentials
		if len(savedPw) > 0 {
			for _, cep := range ceps.ComponentEndpoints {
				cred := compcreds.CompCredentials{
					Xname:    cep.ID,
					URL:      cep.URL,
					Username: savedUn,
					Password: savedPw,
				}
				err := s.ccs.StoreCompCred(cred)
				if err != nil {
					// If we fail to store credentials in vault, we'll lose the
					// credentials and the component endpoints associated with
					// them will still be successfully in the database.
					s.LogAlways("Failed to store credentials for %s in Vault - %s", cep.ID, err)
					savedErr = err
				}
			}
		}
	}

	// Generate HWInv History Entries
	err = s.GenerateHWInvHist(hwlocs)
	if err != nil {
		// Unexpected error storing HWInv history entries.
		s.LogAlways("GenerateHWInvHist(): Fatal error storing: %s", err)
		if savedErr == nil {
			return err
		}
	}

	// Return "main" error as far as whether discovered info could be written.
	return savedErr
}

////////////////////////////////////////////////////////////////////////////
//
// Discovery/creation of ComponentEndpoints from Redfish Endpoint data
//
////////////////////////////////////////////////////////////////////////////

// Create a new array of ComponentEndpoints based on a post-discover
// redfish endpoint discovery struct.
func (s *SmD) DiscoverComponentEndpointArray(rfEP *rf.RedfishEP) (*sm.ComponentEndpointArray, error) {
	ceps := new(sm.ComponentEndpointArray)
	for _, chEP := range rfEP.Chassis.OIDs {
		cep := s.DiscoverCompEndpointChassis(chEP)
		if cep != nil {
			ceps.ComponentEndpoints = append(ceps.ComponentEndpoints, cep)
		}
	}
	for _, sysEP := range rfEP.Systems.OIDs {
		cep := s.DiscoverCompEndpointSystem(sysEP)
		if cep != nil {
			ceps.ComponentEndpoints = append(ceps.ComponentEndpoints, cep)
		}
	}
	for _, mEP := range rfEP.Managers.OIDs {
		cep := s.DiscoverCompEndpointManager(mEP)
		if cep != nil {
			ceps.ComponentEndpoints = append(ceps.ComponentEndpoints, cep)
		}
	}
	for _, pduEP := range rfEP.RackPDUs.OIDs {
		cep := s.DiscoverCompEndpointRackPDU(pduEP)
		if cep != nil {
			ceps.ComponentEndpoints = append(ceps.ComponentEndpoints, cep)
			for _, outEP := range pduEP.Outlets.OIDs {
				cout := s.DiscoverCompEndpointOutlet(outEP)
				if cout != nil {
					ceps.ComponentEndpoints =
						append(ceps.ComponentEndpoints, cout)
				}
			}
		}
	}
	return ceps, nil
}

// Use discovered data on a Redfish (not HMS) Chassis type to create
// an HMS ComponentEndpoint representation.
func (s *SmD) DiscoverCompEndpointChassis(chEP *rf.EpChassis) *sm.ComponentEndpoint {
	if chEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentChassis: EP: %s RF Subtype %s "+
			"not supported.", chEP.RfEndpointID, chEP.RedfishSubtype)
		return nil
	} else if chEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentChassis: Saw EP with bad status: %s",
			chEP.LastStatus)
		return nil
	}
	cep := new(sm.ComponentEndpoint)

	cep.ComponentDescription = chEP.ComponentDescription
	cep.URL = chEP.ChassisURL
	cep.ComponentEndpointType = sm.CompEPTypeChassis
	cep.RedfishChassisInfo = &chEP.ComponentChassisInfo

	return cep
}

// Use discovered data on a Redfish (not HMS) System type to create
// an HMS ComponentEndpoint representation.
func (s *SmD) DiscoverCompEndpointSystem(sysEP *rf.EpSystem) *sm.ComponentEndpoint {
	if sysEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentSystem: EP: %s RF Subtype %s "+
			"not supported.", sysEP.RfEndpointID, sysEP.RedfishSubtype)
		return nil
	} else if sysEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentSystem: Saw EP with bad status: %s",
			sysEP.LastStatus)
		return nil
	}
	cep := new(sm.ComponentEndpoint)

	cep.ComponentDescription = sysEP.ComponentDescription
	cep.URL = sysEP.SystemURL
	cep.ComponentEndpointType = sm.CompEPTypeSystem
	cep.RedfishSystemInfo = &sysEP.ComponentSystemInfo

	return cep
}

// Use discovered data on a Redfish (not HMS) Manager type to create
// an HMS ComponentEndpoint representation.
func (s *SmD) DiscoverCompEndpointManager(mEP *rf.EpManager) *sm.ComponentEndpoint {
	if mEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentManager: EP: %s RF Subtype %s "+
			"not supported.", mEP.RfEndpointID, mEP.RedfishSubtype)
		return nil
	} else if mEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentManager: Saw EP with bad status: %s",
			mEP.LastStatus)
		return nil
	}
	cep := new(sm.ComponentEndpoint)

	cep.ComponentDescription = mEP.ComponentDescription
	cep.URL = mEP.ManagerURL
	cep.ComponentEndpointType = sm.CompEPTypeManager
	cep.RedfishManagerInfo = &mEP.ComponentManagerInfo

	return cep
}

// Use discovered data on a Redfish (not HMS) PowerDistribution (PDU) type
// to create an HMS ComponentEndpoint representation.
func (s *SmD) DiscoverCompEndpointRackPDU(pduEP *rf.EpPDU) *sm.ComponentEndpoint {
	if pduEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentRackPDU: EP: %s RF Subtype %s "+
			"not supported.", pduEP.RfEndpointID, pduEP.RedfishSubtype)
		return nil
	} else if pduEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentRackPDU: Saw EP with bad status: %s",
			pduEP.LastStatus)
		return nil
	}
	cep := new(sm.ComponentEndpoint)

	cep.ComponentDescription = pduEP.ComponentDescription
	cep.URL = pduEP.PDUURL
	cep.ComponentEndpointType = sm.CompEPTypePDU
	cep.RedfishPDUInfo = &pduEP.ComponentPDUInfo

	return cep
}

// Use discovered data on a Redfish (not HMS) Outlet (e.g. on a PDU) type
// to create an HMS ComponentEndpoint representation.
func (s *SmD) DiscoverCompEndpointOutlet(outEP *rf.EpOutlet) *sm.ComponentEndpoint {
	if outEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentOutlet: EP: %s RF Subtype %s "+
			"not supported.", outEP.RfEndpointID, outEP.RedfishSubtype)
		return nil
	} else if outEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentOutlet: Saw EP with bad status: %s",
			outEP.LastStatus)
		return nil
	}
	cep := new(sm.ComponentEndpoint)

	cep.ComponentDescription = outEP.ComponentDescription
	cep.URL = outEP.OutletURL
	cep.ComponentEndpointType = sm.CompEPTypeOutlet
	cep.RedfishOutletInfo = &outEP.ComponentOutletInfo

	return cep
}

func (s *SmD) DiscoverCompEthInterfaceArray(ep *sm.RedfishEndpoint, ceps *sm.ComponentEndpointArray) []*sm.CompEthInterfaceV2 {
	if ceps == nil || ep == nil {
		return nil
	}
	epMAC := strings.ToLower(strings.ReplaceAll(ep.MACAddr, ":", ""))
	ceis := make([]*sm.CompEthInterfaceV2, 0, 1)
	for _, cep := range ceps.ComponentEndpoints {
		var ethInfo []*rf.EthernetNICInfo
		switch cep.ComponentEndpointType {
		case sm.CompEPTypeSystem:
			if cep.RedfishSystemInfo != nil {
				ethInfo = cep.RedfishSystemInfo.EthNICInfo
			}
		case sm.CompEPTypeManager:
			if cep.RedfishManagerInfo != nil {
				ethInfo = cep.RedfishManagerInfo.EthNICInfo
			}
		}
		if len(ethInfo) == 0 {
			continue
		}
		for _, ei := range ethInfo {
			mac := rf.NormalizeMACIfValid(ei.MACAddress)
			pmac := rf.NormalizeMACIfValid(ei.PermanentMACAddress)
			if mac == "" && pmac == "" {
				continue
			} else if mac == "" {
				mac = pmac
			}
			cei, err := sm.NewCompEthInterfaceV2(ei.Description, mac, cep.ID, []sm.IPAddressMapping{})
			if err != nil {
				s.LogAlways("DiscoverCompEthInterfaceArray: Bad ethernet info for %s: %s", cep.ID, err)
				continue
			}
			// If this is the MAC address of our redfishEndpoint,
			// we might have already been given an IP address.
			if cep.ID == ep.ID && cei.ID == epMAC {
				cei.IPAddrs = []sm.IPAddressMapping{{IPAddr: ep.IPAddr}}
			}
			ceis = append(ceis, cei)
		}
	}
	return ceis
}

////////////////////////////////////////////////////////////////////////////
//
// Discovery: HW Inventory location info
//
////////////////////////////////////////////////////////////////////////////

// When we discover a Redfish Endpoint, the data retrieved is processed
// in the rf package to associate it with basic data needed to place it
// within the system and extra HMS-level metadata.
// Here use use the data from the Redfish base types to create hardware
// inventory entries.  If the location is populated, the FRU info is
// generated as well.
func (s *SmD) DiscoverHWInvByLocArray(rfEP *rf.RedfishEP) ([]*sm.HWInvByLoc, error) {
	var (
		save_err        error          = nil
		savedNodeLoc    *sm.HWInvByLoc = nil
		savedNodeLocFRU string
		modifiedFRUID   bool
	)

	hwlocs := make([]*sm.HWInvByLoc, 0, 1)
	for _, chEP := range rfEP.Chassis.OIDs {
		hwloc, err := s.DiscoverHWInvByLocChassis(chEP)
		if err != nil {
			if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
				if err != base.ErrHMSTypeInvalid {
					save_err = err
				}
				continue
			}
			return nil, err
		}
		hwlocs = append(hwlocs, hwloc)
		for _, powerSupplyEP := range chEP.PowerSupplies.OIDs {
			if powerSupplyEP.Type == base.CMMRectifier.String() {
				hwloc, err := s.DiscoverHWInvByLocCMMRectifier(powerSupplyEP)
				if err != nil {
					if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
						if err != base.ErrHMSTypeInvalid {
							save_err = err
						}
						continue
					}
					return nil, err
				}
				hwlocs = append(hwlocs, hwloc)
			} else if powerSupplyEP.Type == base.NodeEnclosurePowerSupply.String() {
				hwloc, err := s.DiscoverHWInvByLocNodeEnclosurePowerSupply(powerSupplyEP)
				if err != nil {
					if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
						if err != base.ErrHMSTypeInvalid {
							save_err = err
						}
						continue
					}
					return nil, err
				}
				hwlocs = append(hwlocs, hwloc)
			}
		}
	}
	// Nodes, from Redfish "System" objects
	for _, sysEP := range rfEP.Systems.OIDs {
		hwloc, err := s.DiscoverHWInvByLocSystem(sysEP)
		if err != nil {
			if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
				if err != base.ErrHMSTypeInvalid {
					save_err = err
				}
				continue
			}
			return nil, err
		}
		// Sometimes a single FRU will have multiple nodes and thus multiple
		// nodes with the same FRUID. Modify the FRUID so they will be different.
		// This should still end up creating the same FRUID for the nodes if
		// they get moved since they'll be moving together and will generally
		// be in the same order.
		if hwloc != nil && hwloc.PopulatedFRU != nil {
			if savedNodeLoc == nil {
				savedNodeLoc = hwloc
				savedNodeLocFRU = hwloc.PopulatedFRU.FRUID
			} else if savedNodeLocFRU == hwloc.PopulatedFRU.FRUID {
				hwloc.PopulatedFRU.FRUID = hwloc.PopulatedFRU.FRUID + "_" + strconv.Itoa(hwloc.Ordinal)
				if !modifiedFRUID {
					savedNodeLoc.PopulatedFRU.FRUID = savedNodeLoc.PopulatedFRU.FRUID + "_" + strconv.Itoa(savedNodeLoc.Ordinal)
					modifiedFRUID = true
				}
			}
		}
		hwlocs = append(hwlocs, hwloc)
		// Now do node subcomponents
		for _, hpeDeviceEP := range sysEP.HpeDevices.OIDs {
			hwloc, err := s.DiscoverHWInvByLocHpeDevice(hpeDeviceEP)
			if err != nil {
				if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
					if err != base.ErrHMSTypeInvalid {
						save_err = err
					}
					continue
				}
				return nil, err
			}
			hwlocs = append(hwlocs, hwloc)
		}

		for _, procEP := range sysEP.Processors.OIDs {
			hwloc, err := s.DiscoverHWInvByLocProcessor(procEP)
			if err != nil {
				if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
					if err != base.ErrHMSTypeInvalid {
						save_err = err
					}
					continue
				}
				return nil, err
			}
			hwlocs = append(hwlocs, hwloc)
		}
		for _, memEP := range sysEP.MemoryMods.OIDs {
			hwloc, err := s.DiscoverHWInvByLocMemory(memEP)
			if err != nil {
				if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
					if err != base.ErrHMSTypeInvalid {
						save_err = err
					}
					continue
				}
				return nil, err
			}
			hwlocs = append(hwlocs, hwloc)
		}
		for _, driveEP := range sysEP.Drives.OIDs {
			hwloc, err := s.DiscoverHWInvByLocDrive(driveEP)
			if err != nil {
				if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
					if err != base.ErrHMSTypeInvalid {
						save_err = err
					}
					continue
				}
				return nil, err
			}
			hwlocs = append(hwlocs, hwloc)
		}
		for _, nodeAccelRiserEP := range sysEP.NodeAccelRisers.OIDs {
			if nodeAccelRiserEP.Type == base.NodeAccelRiser.String() {
				hwloc, err := s.DiscoverHWInvByLocNodeAccelRiser(nodeAccelRiserEP)
				if err != nil {
					if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
						if err != base.ErrHMSTypeInvalid {
							save_err = err
						}
						continue
					}
					return nil, err
				}
				hwlocs = append(hwlocs, hwloc)
			}
		}
		for _, networkAdapterEP := range sysEP.NetworkAdapters.OIDs {
			if networkAdapterEP.Type == base.NodeHsnNic.String() {
				hwloc, err := s.DiscoverHWInvByLocNodeHsnNic(networkAdapterEP)
				if err != nil {
					if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
						if err != base.ErrHMSTypeInvalid {
							save_err = err
						}
						continue
					}
					return nil, err
				}
				hwlocs = append(hwlocs, hwloc)
			}
		}
	}
	// RackPDUs, from Redfish "PowerDistribution" objects
	for _, pduEP := range rfEP.RackPDUs.OIDs {
		hwloc, err := s.DiscoverHWInvByLocPDU(pduEP)
		if err != nil {
			if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
				if err != base.ErrHMSTypeInvalid {
					save_err = err
				}
				continue
			}
			return nil, err
		}
		hwlocs = append(hwlocs, hwloc)
		// Now do outlet subcomponents
		for _, outEP := range pduEP.Outlets.OIDs {
			hwloc, err := s.DiscoverHWInvByLocOutlet(outEP)
			if err != nil {
				if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
					if err != base.ErrHMSTypeInvalid {
						save_err = err
					}
					continue
				}
				return nil, err
			}
			hwlocs = append(hwlocs, hwloc)
		}
	}

	// Managers from Redfish "Manager" objects
	for _, managerEP := range rfEP.Managers.OIDs {
		if managerEP.Type == base.NodeBMC.String() {
			hwloc, err := s.DiscoverHWInvByLocNodeBMC(managerEP)
			if err != nil {
				if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
					if err != base.ErrHMSTypeInvalid {
						save_err = err
					}
					continue
				}
				return nil, err
			}
			hwlocs = append(hwlocs, hwloc)
		}
		if managerEP.Type == base.RouterBMC.String() {
			hwloc, err := s.DiscoverHWInvByLocRouterBMC(managerEP)
			if err != nil {
				if err == base.ErrHMSTypeInvalid || err == base.ErrHMSTypeUnsupported {
					if err != base.ErrHMSTypeInvalid {
						save_err = err
					}
					continue
				}
				return nil, err
			}
			hwlocs = append(hwlocs, hwloc)
		}
	}

	// TODO: Managers, i.e. BMCs and other controllers represented by
	// Redfish "Manager" types.  Mostly this is not inventory info,
	// however, which comes from the containing enclosure.

	return hwlocs, save_err
}

// Generate HWInv history entries. This determines the historical event type
// based on the most recent entry for that FRU.
//   - No history means we are adding.
//   - Previously removed means we are adding.
//   - Same location means we are scanning.
//   - Different location mean we are removing from the previous location and
//     adding to the new location. (Removal event wasn't generated)
//
// If a eventType is specified, all generated entries will be forced to that
// event type.
func (s *SmD) GenerateHWInvHist(hwlocs []*sm.HWInvByLoc) error {
	hwhists := make([]*sm.HWInvHist, 0, 1)
	locIDs := make([]string, 0, len(hwlocs))
	lhsMap := make(map[string]string, 0)

	// Get a list of the LocIDs
	for _, hwloc := range hwlocs {
		if hwloc == nil || hwloc.PopulatedFRU == nil {
			continue
		}
		locIDs = append(locIDs, hwloc.ID)
	}
	// Get the most recent event for each locID
	lhs, err := s.db.GetHWInvHistLastEvents(locIDs)
	if err != nil {
		return err
	}
	// Create a map linking the locIDs to the last recorded history entry
	for _, lh := range lhs {
		lhsMap[lh.ID] = lh.FruId
	}
	for _, hwloc := range hwlocs {
		// Skip hwlocs that have no FRU
		if hwloc == nil || hwloc.PopulatedFRU == nil {
			continue
		}
		newHist := sm.HWInvHist{
			ID:        hwloc.ID,
			FruId:     hwloc.PopulatedFRU.FRUID,
			EventType: sm.HWInvHistEventTypeDetected,
		}
		if fruID, ok := lhsMap[hwloc.ID]; !ok || fruID != hwloc.PopulatedFRU.FRUID {
			hwhists = append(hwhists, &newHist)
		}
	}
	if len(hwhists) > 0 {
		// Insert the history events into the database
		err = s.db.InsertHWInvHists(hwhists)
	}
	return err
}

// Most components above nodes except controllers/BMCs are
// Redfish "Chassis", objects a catch all for most physical enclosure
// types.  Use the annotated data retrieved from the parent Redfish
// entry point to create a HW inventory-by-location object.
func (s *SmD) DiscoverHWInvByLocChassis(chEP *rf.EpChassis) (*sm.HWInvByLoc, error) {
	if chEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocChassis: EP: %s RF Subtype %s "+
			"not supported.", chEP.RfEndpointID, chEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if chEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocChassis: Saw EP with bad status: %s",
			chEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = chEP.ID
	hwloc.Type = chEP.Type
	hwloc.Ordinal = chEP.Ordinal
	hwloc.Status = chEP.Status
	if hwloc.Status != "Empty" && chEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUChassis(chEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	//rfChassisLocationInfo := &chEP.ChassisRF.ChassisLocationInfoRF
	switch base.ToHMSType(hwloc.Type) {
	case base.Cabinet:
		hwloc.HMSCabinetLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocCabinet
	case base.Chassis:
		hwloc.HMSChassisLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocChassis
	case base.ComputeModule:
		hwloc.HMSComputeModuleLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocComputeModule
	case base.RouterModule:
		hwloc.HMSRouterModuleLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocRouterModule
	case base.NodeEnclosure:
		hwloc.HMSNodeEnclosureLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocNodeEnclosure
	case base.HSNBoard:
		hwloc.HMSHSNBoardLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocHSNBoard
	case base.MgmtSwitch:
		hwloc.HMSMgmtSwitchLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocMgmtSwitch
	case base.MgmtHLSwitch:
		hwloc.HMSMgmtHLSwitchLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocMgmtHLSwitch
	case base.CDUMgmtSwitch:
		hwloc.HMSCDUMgmtSwitchLocationInfo = &chEP.ChassisRF.ChassisLocationInfoRF
		hwloc.HWInventoryByLocationType = sm.HWInvByLocCDUMgmtSwitch
	case base.HMSTypeInvalid:
		err := base.ErrHMSTypeInvalid
		return nil, err
	default:
		err := base.ErrHMSTypeUnsupported
		return nil, err
	}
	return hwloc, nil
}

// HMS nodes, based on info retrieved from Redfish "System" objects
func (s *SmD) DiscoverHWInvByLocSystem(sysEP *rf.EpSystem) (*sm.HWInvByLoc, error) {
	if sysEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocSystem: EP: %s RF Subtype %s "+
			"not supported.", sysEP.RfEndpointID, sysEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if sysEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocSystem: Saw EP with bad status: %s",
			sysEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = sysEP.ID
	hwloc.Type = sysEP.Type
	hwloc.Ordinal = sysEP.Ordinal
	hwloc.Status = sysEP.Status
	if hwloc.Status != "Empty" && sysEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUSystem(sysEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSNodeLocationInfo = &sysEP.SystemRF.SystemLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocNode
	return hwloc, nil
}

// HMS GPUs, etc, based on info retrieved by HPE Device Redfish objects
func (s *SmD) DiscoverHWInvByLocHpeDevice(hpeDeviceEP *rf.EpHpeDevice) (*sm.HWInvByLoc, error) {
	if hpeDeviceEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocHpeDevice: EP: %s RF Subtype %s "+
			"not supported.", hpeDeviceEP.RfEndpointID, hpeDeviceEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if hpeDeviceEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocHpeDevice: Saw EP with bad status: %s",
			hpeDeviceEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = hpeDeviceEP.ID
	hwloc.Type = hpeDeviceEP.Type
	hwloc.Ordinal = hpeDeviceEP.Ordinal
	hwloc.Status = hpeDeviceEP.Status
	if hwloc.Status != "Empty" && hpeDeviceEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUHpeDevice(hpeDeviceEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	switch base.ToHMSType(hwloc.Type) {
	case base.NodeAccel:
		accelInfo := rf.ProcessorLocationInfoRF{
			Id:          hpeDeviceEP.DeviceRF.Id,
			Name:        hpeDeviceEP.DeviceRF.Name,
			Description: hpeDeviceEP.DeviceRF.Location,
		}
		hwloc.HMSNodeAccelLocationInfo = &accelInfo
		hwloc.HWInventoryByLocationType = sm.HWInvByLocNodeAccel
	case base.NodeHsnNic:
		nicInfo := rf.NALocationInfoRF{
			Id:          hpeDeviceEP.DeviceRF.Id,
			Name:        hpeDeviceEP.DeviceRF.Name,
			Description: hpeDeviceEP.DeviceRF.Location,
		}
		hwloc.HMSHSNNICLocationInfo = &nicInfo
		hwloc.HWInventoryByLocationType = sm.HWInvByLocHSNNIC
	case base.HMSTypeInvalid:
		err := base.ErrHMSTypeInvalid
		return nil, err
	default:
		err := base.ErrHMSTypeUnsupported
		return nil, err
	}
	return hwloc, nil
}

// HMS Processors, based on info retrieved by Redfish object of the same name
func (s *SmD) DiscoverHWInvByLocProcessor(procEP *rf.EpProcessor) (*sm.HWInvByLoc, error) {
	if procEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocProcessor: EP: %s RF Subtype %s "+
			"not supported.", procEP.RfEndpointID, procEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if procEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocProcessor: Saw EP with bad status: %s",
			procEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = procEP.ID
	hwloc.Type = procEP.Type
	hwloc.Ordinal = procEP.Ordinal
	hwloc.Status = procEP.Status
	if hwloc.Status != "Empty" && procEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUProcessor(procEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSProcessorLocationInfo = &procEP.ProcessorRF.ProcessorLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocProcessor
	return hwloc, nil
}

// HMS Memory modules, based on info retrieved by Redfish object of the same
// name
func (s *SmD) DiscoverHWInvByLocMemory(memEP *rf.EpMemory) (*sm.HWInvByLoc, error) {
	if memEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocMemory: EP: %s RF Subtype %s "+
			"not supported.", memEP.RfEndpointID, memEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if memEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocMemory: Saw EP with bad status: %s",
			memEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = memEP.ID
	hwloc.Type = memEP.Type
	hwloc.Ordinal = memEP.Ordinal
	hwloc.Status = memEP.Status
	if hwloc.Status != "Empty" && memEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUMemory(memEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSMemoryLocationInfo = &memEP.MemoryRF.MemoryLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocMemory
	return hwloc, nil
}

// HMS Drives, based on info retrieved by Redfish object of the same name
func (s *SmD) DiscoverHWInvByLocDrive(driveEP *rf.EpDrive) (*sm.HWInvByLoc, error) {
	if driveEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocDrive: EP: %s RF Subtype %s "+
			"not supported.", driveEP.RfEndpointID, driveEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if driveEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocDrive: Saw EP with bad status: %s",
			driveEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = driveEP.ID
	hwloc.Type = driveEP.Type
	hwloc.Ordinal = driveEP.Ordinal
	hwloc.Status = driveEP.Status
	if hwloc.Status != "Empty" && driveEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUDrive(driveEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSDriveLocationInfo = &driveEP.DriveRF.DriveLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocDrive
	return hwloc, nil
}

// HMS PowerDistribution modules, based on info retrieved by Redfish object
// of the same name
func (s *SmD) DiscoverHWInvByLocPDU(pduEP *rf.EpPDU) (*sm.HWInvByLoc, error) {
	if pduEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocPDU: EP: %s RF Subtype %s "+
			"not supported.", pduEP.RfEndpointID, pduEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if pduEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocPDU: Saw EP with bad status: %s",
			pduEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = pduEP.ID
	hwloc.Type = pduEP.Type
	hwloc.Ordinal = pduEP.Ordinal
	hwloc.Status = pduEP.Status
	if hwloc.Status != "Empty" && pduEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUPDU(pduEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSPDULocationInfo = &pduEP.PowerDistributionRF.PowerDistributionLocationInfo
	hwloc.HWInventoryByLocationType = sm.HWInvByLocPDU
	return hwloc, nil
}

// HMS PowerDistribution modules, based on info retrieved by Redfish object
// of the same name
func (s *SmD) DiscoverHWInvByLocOutlet(outEP *rf.EpOutlet) (*sm.HWInvByLoc, error) {
	if outEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocOutlet: EP: %s RF Subtype %s "+
			"not supported.", outEP.RfEndpointID, outEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if outEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocOutlet: Saw EP with bad status: %s",
			outEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = outEP.ID
	hwloc.Type = outEP.Type
	hwloc.Ordinal = outEP.Ordinal
	hwloc.Status = outEP.Status
	if hwloc.Status != "Empty" && outEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUOutlet(outEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSOutletLocationInfo = &outEP.OutletRF.OutletLocationInfo
	hwloc.HWInventoryByLocationType = sm.HWInvByLocOutlet
	return hwloc, nil
}

// HMS CMMRectifier, based on info retrieved by a Redfish PowerSupply
func (s *SmD) DiscoverHWInvByLocCMMRectifier(powerSupplyEP *rf.EpPowerSupply) (*sm.HWInvByLoc, error) {
	if powerSupplyEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocCMMRectifier: EP: %s RF Subtype %s "+
			"not supported.", powerSupplyEP.RfEndpointID, powerSupplyEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if powerSupplyEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocCMMRectifier: Saw EP with bad status: %s",
			powerSupplyEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = powerSupplyEP.ID
	hwloc.Type = powerSupplyEP.Type
	hwloc.Ordinal = powerSupplyEP.Ordinal
	hwloc.Status = powerSupplyEP.Status
	if hwloc.Status != "Empty" && powerSupplyEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUCMMRectifier(powerSupplyEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSCMMRectifierLocationInfo = &powerSupplyEP.PowerSupplyRF.PowerSupplyLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocCMMRectifier
	return hwloc, nil
}

// HMS NodeEnclosurePowerSupply, based on info retrieved by a Redfish PowerSupply
func (s *SmD) DiscoverHWInvByLocNodeEnclosurePowerSupply(powerSupplyEP *rf.EpPowerSupply) (*sm.HWInvByLoc, error) {
	if powerSupplyEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocNodeEnclosurePowerSupply: EP: %s RF Subtype %s "+
			"not supported.", powerSupplyEP.RfEndpointID, powerSupplyEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if powerSupplyEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocNodeEnclosurePowerSupply: Saw EP with bad status: %s",
			powerSupplyEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = powerSupplyEP.ID
	hwloc.Type = powerSupplyEP.Type
	hwloc.Ordinal = powerSupplyEP.Ordinal
	hwloc.Status = powerSupplyEP.Status
	if hwloc.Status != "Empty" && powerSupplyEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUNodeEnclosurePowerSupply(powerSupplyEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSNodeEnclosurePowerSupplyLocationInfo = &powerSupplyEP.PowerSupplyRF.PowerSupplyLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocNodeEnclosurePowerSupply
	return hwloc, nil
}

func (s *SmD) DiscoverHWInvByLocNodeBMC(managerEP *rf.EpManager) (*sm.HWInvByLoc, error) {
	if managerEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocNodeBMC: EP: %s RF Subtype %s "+
			"not supported.", managerEP.RfEndpointID, managerEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if managerEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocNodeBMC: Saw EP with bad status: %s",
			managerEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = managerEP.ID
	hwloc.Type = managerEP.Type
	hwloc.Ordinal = managerEP.Ordinal
	hwloc.Status = managerEP.Status
	if hwloc.Status != "Empty" && managerEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUNodeBMC(managerEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSNodeBMCLocationInfo = &managerEP.ManagerRF.ManagerLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocNodeBMC
	return hwloc, nil
}

func (s *SmD) DiscoverHWInvByLocRouterBMC(managerEP *rf.EpManager) (*sm.HWInvByLoc, error) {
	if managerEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocRouterBMC: EP: %s RF Subtype %s "+
			"not supported.", managerEP.RfEndpointID, managerEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if managerEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocRouterBMC: Saw EP with bad status: %s",
			managerEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = managerEP.ID
	hwloc.Type = managerEP.Type
	hwloc.Ordinal = managerEP.Ordinal
	hwloc.Status = managerEP.Status
	if hwloc.Status != "Empty" && managerEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRURouterBMC(managerEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSRouterBMCLocationInfo = &managerEP.ManagerRF.ManagerLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocRouterBMC
	return hwloc, nil
}

// HMS NodeAccelRiser, based on info retrieved by a Redfish NodeAccelRiser
func (s *SmD) DiscoverHWInvByLocNodeAccelRiser(nodeAccelRiserEP *rf.EpNodeAccelRiser) (*sm.HWInvByLoc, error) {
	if nodeAccelRiserEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocNodeAccelRiser: EP: %s RF Subtype %s "+
			"not supported.", nodeAccelRiserEP.RfEndpointID, nodeAccelRiserEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if nodeAccelRiserEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocNodeAccelRiser: Saw EP with bad status: %s",
			nodeAccelRiserEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = nodeAccelRiserEP.ID
	hwloc.Type = nodeAccelRiserEP.Type
	hwloc.Ordinal = nodeAccelRiserEP.Ordinal
	hwloc.Status = nodeAccelRiserEP.Status
	if hwloc.Status != "Empty" && nodeAccelRiserEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUNodeAccelRiser(nodeAccelRiserEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSNodeAccelRiserLocationInfo = &nodeAccelRiserEP.NodeAccelRiserRF.NodeAccelRiserLocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocNodeAccelRiser
	return hwloc, nil
}

// HMS NodeHSNNIC, based on info retrieved by a Redfish NetworkAdapter
func (s *SmD) DiscoverHWInvByLocNodeHsnNic(networkAdapterEP *rf.EpNetworkAdapter) (*sm.HWInvByLoc, error) {
	if networkAdapterEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByLocNodeHsnNic: EP: %s RF Subtype %s "+
			"not supported.", networkAdapterEP.RfEndpointID, networkAdapterEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if networkAdapterEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByLocNodeHsnNic: Saw EP with bad status: %s",
			networkAdapterEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwloc := new(sm.HWInvByLoc)
	hwloc.ID = networkAdapterEP.ID
	hwloc.Type = networkAdapterEP.Type
	hwloc.Ordinal = networkAdapterEP.Ordinal
	hwloc.Status = networkAdapterEP.Status
	if hwloc.Status != "Empty" && networkAdapterEP.FRUID != "" {
		hwfru, err := s.DiscoverHWInvByFRUNodeHsnNic(networkAdapterEP)
		if err != nil {
			return nil, err
		}
		hwloc.PopulatedFRU = hwfru
	}
	hwloc.HMSHSNNICLocationInfo = &networkAdapterEP.NetworkAdapterRF.NALocationInfoRF
	hwloc.HWInventoryByLocationType = sm.HWInvByLocHSNNIC
	return hwloc, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Discovery: HW Inventory FRU info
//
////////////////////////////////////////////////////////////////////////////

// Use collected and annotated Redfish data retried from remote endpoints
// to create FRU inventory entries

// HMS types represented by Redfish "Component" objects
func (s *SmD) DiscoverHWInvByFRUChassis(chEP *rf.EpChassis) (*sm.HWInvByFRU, error) {
	if chEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUChassis: EP: %s RF Subtype %s "+
			"not supported.", chEP.RfEndpointID, chEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if chEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUChassis: Saw EP with bad status: %s",
			chEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if chEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = chEP.FRUID
	hwfru.Type = chEP.Type
	hwfru.Subtype = chEP.Subtype

	rfChassisFRUInfo := &chEP.ChassisRF.ChassisFRUInfoRF
	switch base.ToHMSType(hwfru.Type) {
	case base.Cabinet:
		hwfru.HMSCabinetFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUCabinet
	case base.Chassis:
		hwfru.HMSChassisFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUChassis
	case base.ComputeModule:
		hwfru.HMSComputeModuleFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUComputeModule
	case base.RouterModule:
		hwfru.HMSRouterModuleFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRURouterModule
	case base.NodeEnclosure:
		hwfru.HMSNodeEnclosureFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUNodeEnclosure
	case base.HSNBoard:
		hwfru.HMSHSNBoardFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUHSNBoard
	case base.MgmtSwitch:
		hwfru.HMSMgmtSwitchFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUMgmtSwitch
	case base.MgmtHLSwitch:
		hwfru.HMSMgmtHLSwitchFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUMgmtHLSwitch
	case base.CDUMgmtSwitch:
		hwfru.HMSCDUMgmtSwitchFRUInfo = rfChassisFRUInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUCDUMgmtSwitch
	case base.HMSTypeInvalid:
		err := base.ErrHMSTypeInvalid
		return nil, err
	default:
		err := base.ErrHMSTypeUnsupported
		return nil, err
	}
	return hwfru, nil
}

// HMS nodes, represented by Redfish "System" objects.  Generate FRU-level
// HW inventory data based on Redfish properties that are persistent even
// if the component moves.
func (s *SmD) DiscoverHWInvByFRUSystem(sysEP *rf.EpSystem) (*sm.HWInvByFRU, error) {
	if sysEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUSystem: EP: %s RF Subtype %s "+
			"not supported.", sysEP.RfEndpointID, sysEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if sysEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUSystem: Saw EP with bad status: %s",
			sysEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if sysEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = sysEP.FRUID
	hwfru.Type = sysEP.Type
	hwfru.Subtype = sysEP.Subtype

	hwfru.HMSNodeFRUInfo = &sysEP.SystemRF.SystemFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUNode

	return hwfru, nil
}

// HMS GPU, etc FRU info, based on info retrieved by HPE Device Redfish objects.
func (s *SmD) DiscoverHWInvByFRUHpeDevice(hpeDeviceEP *rf.EpHpeDevice) (*sm.HWInvByFRU, error) {
	if hpeDeviceEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUHpeDevice: EP: %s RF Subtype %s "+
			"not supported.", hpeDeviceEP.RfEndpointID, hpeDeviceEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if hpeDeviceEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUHpeDevice: Saw EP with bad status: %s",
			hpeDeviceEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if hpeDeviceEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = hpeDeviceEP.FRUID
	hwfru.Type = hpeDeviceEP.Type
	hwfru.Subtype = hpeDeviceEP.Subtype

	switch base.ToHMSType(hwfru.Type) {
	case base.NodeAccel:
		accelInfo := rf.ProcessorFRUInfoRF{
			Manufacturer:  hpeDeviceEP.DeviceRF.Manufacturer,
			Model:         hpeDeviceEP.DeviceRF.Model,
			SerialNumber:  hpeDeviceEP.DeviceRF.SerialNumber,
			PartNumber:    hpeDeviceEP.DeviceRF.PartNumber,
			ProcessorType: hpeDeviceEP.DeviceRF.DeviceType,
		}
		hwfru.HMSNodeAccelFRUInfo = &accelInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUNodeAccel
	case base.NodeHsnNic:
		nicInfo := rf.NAFRUInfoRF{
			Manufacturer: hpeDeviceEP.DeviceRF.Manufacturer,
			Model:        hpeDeviceEP.DeviceRF.Model,
			SerialNumber: hpeDeviceEP.DeviceRF.SerialNumber,
			PartNumber:   hpeDeviceEP.DeviceRF.PartNumber,
		}
		hwfru.HMSHSNNICFRUInfo = &nicInfo
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUHSNNIC
	case base.HMSTypeInvalid:
		err := base.ErrHMSTypeInvalid
		return nil, err
	default:
		err := base.ErrHMSTypeUnsupported
		return nil, err
	}

	return hwfru, nil
}

// HMS Processor FRU info, based on info retrieved by Redfish object of the
// same name under the parent node.
func (s *SmD) DiscoverHWInvByFRUProcessor(procEP *rf.EpProcessor) (*sm.HWInvByFRU, error) {
	if procEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUProcessor: EP: %s RF Subtype %s "+
			"not supported.", procEP.RfEndpointID, procEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if procEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUProcessor: Saw EP with bad status: %s",
			procEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if procEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = procEP.FRUID
	hwfru.Type = procEP.Type
	hwfru.Subtype = procEP.Subtype

	if procEP.Type == base.NodeAccel.String() {
		hwfru.HMSNodeAccelFRUInfo = &procEP.ProcessorRF.ProcessorFRUInfoRF
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUNodeAccel
	} else {
		hwfru.HMSProcessorFRUInfo = &procEP.ProcessorRF.ProcessorFRUInfoRF
		hwfru.HWInventoryByFRUType = sm.HWInvByFRUProcessor
	}

	return hwfru, nil
}

// HMS Memory module FRU info, based on info retrieved via the Redfish object
// of the same name under the parent node.
func (s *SmD) DiscoverHWInvByFRUMemory(memEP *rf.EpMemory) (*sm.HWInvByFRU, error) {
	if memEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUMemory: EP: %s RF Subtype %s "+
			"not supported.", memEP.RfEndpointID, memEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if memEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUMemory: Saw EP with bad status: %s",
			memEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if memEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = memEP.FRUID
	hwfru.Type = memEP.Type
	hwfru.Subtype = memEP.Subtype

	hwfru.HMSMemoryFRUInfo = &memEP.MemoryRF.MemoryFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUMemory

	return hwfru, nil
}

// HMS Drive FRU info, based on info retrieved by Redfish object of the
// same name under the parent node.
func (s *SmD) DiscoverHWInvByFRUDrive(driveEP *rf.EpDrive) (*sm.HWInvByFRU, error) {
	if driveEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUDrive: EP: %s RF Subtype %s "+
			"not supported.", driveEP.RfEndpointID, driveEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if driveEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUDrive: Saw EP with bad status: %s",
			driveEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if driveEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = driveEP.FRUID
	hwfru.Type = driveEP.Type
	hwfru.Subtype = driveEP.Subtype

	hwfru.HMSDriveFRUInfo = &driveEP.DriveRF.DriveFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUDrive

	return hwfru, nil
}

// HMS PowerDistribution module FRU info, based on info retrieved via the
// Redfish object of the same name.
func (s *SmD) DiscoverHWInvByFRUPDU(pduEP *rf.EpPDU) (*sm.HWInvByFRU, error) {
	if pduEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUPDU: EP: %s RF Subtype %s "+
			"not supported.", pduEP.RfEndpointID, pduEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if pduEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUPDU: Saw EP with bad status: %s",
			pduEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if pduEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = pduEP.FRUID
	hwfru.Type = pduEP.Type
	hwfru.Subtype = pduEP.Subtype

	hwfru.HMSPDUFRUInfo = &pduEP.PowerDistributionRF.PowerDistributionFRUInfo
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUPDU

	return hwfru, nil
}

// HMS Outlet module FRU info, based on info retrieved via the
// Redfish object of the same name under a parent PDU
func (s *SmD) DiscoverHWInvByFRUOutlet(outEP *rf.EpOutlet) (*sm.HWInvByFRU, error) {
	if outEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUOutlet: EP: %s RF Subtype %s "+
			"not supported.", outEP.RfEndpointID, outEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if outEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUOutlet: Saw EP with bad status: %s",
			outEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if outEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = outEP.FRUID
	hwfru.Type = outEP.Type
	hwfru.Subtype = outEP.Subtype

	hwfru.HMSOutletFRUInfo = &outEP.OutletRF.OutletFRUInfo
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUOutlet

	return hwfru, nil
}

// HMS PowerSupply FRU info, based on info retrieved by Redfish object of the
// same name.
func (s *SmD) DiscoverHWInvByFRUCMMRectifier(powerSupplyEP *rf.EpPowerSupply) (*sm.HWInvByFRU, error) {
	if powerSupplyEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUCMMRectifier: EP: %s RF Subtype %s "+
			"not supported.", powerSupplyEP.RfEndpointID, powerSupplyEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if powerSupplyEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUCMMRectifier: Saw EP with bad status: %s",
			powerSupplyEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if powerSupplyEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = powerSupplyEP.FRUID
	hwfru.Type = powerSupplyEP.Type
	hwfru.Subtype = powerSupplyEP.Subtype

	hwfru.HMSCMMRectifierFRUInfo = &powerSupplyEP.PowerSupplyRF.PowerSupplyFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUCMMRectifier

	return hwfru, nil
}

// HMS PowerSupply FRU info, based on info retrieved by Redfish object of the
// same name.
func (s *SmD) DiscoverHWInvByFRUNodeEnclosurePowerSupply(powerSupplyEP *rf.EpPowerSupply) (*sm.HWInvByFRU, error) {
	if powerSupplyEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUNodeEnclosurePowerSupply: EP: %s RF Subtype %s "+
			"not supported.", powerSupplyEP.RfEndpointID, powerSupplyEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if powerSupplyEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUNodeEnclosurePowerSupply: Saw EP with bad status: %s",
			powerSupplyEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if powerSupplyEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = powerSupplyEP.FRUID
	hwfru.Type = powerSupplyEP.Type
	hwfru.Subtype = powerSupplyEP.Subtype

	hwfru.HMSNodeEnclosurePowerSupplyFRUInfo = &powerSupplyEP.PowerSupplyRF.PowerSupplyFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUNodeEnclosurePowerSupply

	return hwfru, nil
}

// HMS NodeBMC FRU info, based on info retrieved by Redfish Manager object with subtype of BMC
// same name.
func (s *SmD) DiscoverHWInvByFRUNodeBMC(managerEP *rf.EpManager) (*sm.HWInvByFRU, error) {
	if managerEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUNodeBMC: EP: %s RF Subtype %s "+
			"not supported.", managerEP.RfEndpointID, managerEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if managerEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUNodeBMC: Saw EP with bad status: %s",
			managerEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if managerEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = managerEP.FRUID
	hwfru.Type = managerEP.Type
	hwfru.Subtype = managerEP.Subtype

	hwfru.HMSNodeBMCFRUInfo = &managerEP.ManagerRF.ManagerFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUNodeBMC

	return hwfru, nil
}

// HMS RouterBMC FRU info, based on info retrieved by Redfish Manager object with subtype of BMC
// same name.
func (s *SmD) DiscoverHWInvByFRURouterBMC(managerEP *rf.EpManager) (*sm.HWInvByFRU, error) {
	if managerEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRURouterBMC: EP: %s RF Subtype %s "+
			"not supported.", managerEP.RfEndpointID, managerEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if managerEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRURouterBMC: Saw EP with bad status: %s",
			managerEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if managerEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = managerEP.FRUID
	hwfru.Type = managerEP.Type
	hwfru.Subtype = managerEP.Subtype

	hwfru.HMSRouterBMCFRUInfo = &managerEP.ManagerRF.ManagerFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRURouterBMC

	return hwfru, nil
}

// HMS NodeAccelRiser FRU info, based on info retrieved by Redfish object of the
// same name.
func (s *SmD) DiscoverHWInvByFRUNodeAccelRiser(nodeAccelRiserEP *rf.EpNodeAccelRiser) (*sm.HWInvByFRU, error) {
	if nodeAccelRiserEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUNodeAccelRiser: EP: %s RF Subtype %s "+
			"not supported.", nodeAccelRiserEP.RfEndpointID, nodeAccelRiserEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if nodeAccelRiserEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUNodeAccelRiser: Saw EP with bad status: %s",
			nodeAccelRiserEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if nodeAccelRiserEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = nodeAccelRiserEP.FRUID
	hwfru.Type = nodeAccelRiserEP.Type
	hwfru.Subtype = nodeAccelRiserEP.Subtype

	hwfru.HMSNodeAccelRiserFRUInfo = &nodeAccelRiserEP.NodeAccelRiserRF.NodeAccelRiserFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUNodeAccelRiser

	return hwfru, nil
}

// HMS NodeAccelRiser FRU info, based on info retrieved by Redfish object of the
// same name.
func (s *SmD) DiscoverHWInvByFRUNodeHsnNic(networkAdapterEP *rf.EpNetworkAdapter) (*sm.HWInvByFRU, error) {
	if networkAdapterEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverHWInvByFRUNodeHsnNic: EP: %s RF Subtype %s "+
			"not supported.", networkAdapterEP.RfEndpointID, networkAdapterEP.RedfishSubtype)
		return nil, base.ErrHMSTypeUnsupported
	} else if networkAdapterEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverHWInvByFRUNodeHsnNic: Saw EP with bad status: %s",
			networkAdapterEP.LastStatus)
		return nil, base.ErrHMSTypeInvalid
	}
	hwfru := new(sm.HWInvByFRU)
	if networkAdapterEP.FRUID == "" {
		return nil, sm.ErrHWFRUIDInvalid
	}
	hwfru.FRUID = networkAdapterEP.FRUID
	hwfru.Type = networkAdapterEP.Type
	hwfru.Subtype = networkAdapterEP.Subtype

	hwfru.HMSHSNNICFRUInfo = &networkAdapterEP.NetworkAdapterRF.NAFRUInfoRF
	hwfru.HWInventoryByFRUType = sm.HWInvByFRUHSNNIC

	return hwfru, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Discovery/creation of ServiceEndpoints from Redfish Endpoint data
//
////////////////////////////////////////////////////////////////////////////

// Create a new array of ServiceEndpoints based on a post-discover
// redfish endpoint discovery struct.
func (s *SmD) DiscoverServiceEndpointArray(rfEP *rf.RedfishEP) *sm.ServiceEndpointArray {
	seps := new(sm.ServiceEndpointArray)

	if rfEP.AccountService != nil {
		sep := new(sm.ServiceEndpoint)

		sep.ServiceDescription = rfEP.AccountService.ServiceDescription
		sep.RfEndpointFQDN = rfEP.AccountService.RootFQDN
		sep.URL = rfEP.AccountService.AccountServiceURL
		infoJSON, err := json.Marshal(rfEP.AccountService.AccountServiceRF)
		if err != nil {
			// This should never fail
			s.LogAlways("DiscoverServiceEndpointArray: decode AccountServiceInfo: %s", err)
		} else {
			sep.ServiceInfo = json.RawMessage(infoJSON)
			seps.ServiceEndpoints = append(seps.ServiceEndpoints, sep)
		}
	}
	if rfEP.SessionService != nil {
		sep := new(sm.ServiceEndpoint)

		sep.ServiceDescription = rfEP.SessionService.ServiceDescription
		sep.RfEndpointFQDN = rfEP.SessionService.RootFQDN
		sep.URL = rfEP.SessionService.SessionServiceURL
		infoJSON, err := json.Marshal(rfEP.SessionService.SessionServiceRF)
		if err != nil {
			// This should never fail
			s.LogAlways("DiscoverServiceEndpointArray: decode SessionServiceInfo: %s", err)
		} else {
			sep.ServiceInfo = json.RawMessage(infoJSON)
			seps.ServiceEndpoints = append(seps.ServiceEndpoints, sep)
		}
	}
	if rfEP.EventService != nil {
		sep := new(sm.ServiceEndpoint)

		sep.ServiceDescription = rfEP.EventService.ServiceDescription
		sep.RfEndpointFQDN = rfEP.EventService.RootFQDN
		sep.URL = rfEP.EventService.EventServiceURL
		infoJSON, err := json.Marshal(rfEP.EventService.EventServiceRF)
		if err != nil {
			// This should never fail
			s.LogAlways("DiscoverServiceEndpointArray: decode EventServiceInfo: %s", err)
		} else {
			sep.ServiceInfo = json.RawMessage(infoJSON)
			seps.ServiceEndpoints = append(seps.ServiceEndpoints, sep)
		}
	}
	if rfEP.TaskService != nil {
		sep := new(sm.ServiceEndpoint)

		sep.ServiceDescription = rfEP.TaskService.ServiceDescription
		sep.RfEndpointFQDN = rfEP.TaskService.RootFQDN
		sep.URL = rfEP.TaskService.TaskServiceURL
		infoJSON, err := json.Marshal(rfEP.TaskService.TaskServiceRF)
		if err != nil {
			// This should never fail
			s.LogAlways("DiscoverServiceEndpointArray: decode TaskServiceInfo: %s", err)
		} else {
			sep.ServiceInfo = json.RawMessage(infoJSON)
			seps.ServiceEndpoints = append(seps.ServiceEndpoints, sep)
		}
	}
	if rfEP.UpdateService != nil {
		sep := new(sm.ServiceEndpoint)

		sep.ServiceDescription = rfEP.UpdateService.ServiceDescription
		sep.RfEndpointFQDN = rfEP.UpdateService.RootFQDN
		sep.URL = rfEP.UpdateService.UpdateServiceURL
		infoJSON, err := json.Marshal(rfEP.UpdateService.UpdateServiceRF)
		if err != nil {
			// This should never fail
			s.LogAlways("DiscoverServiceEndpointArray: decode UpdateServiceInfo: %s", err)
		} else {
			sep.ServiceInfo = json.RawMessage(infoJSON)
			seps.ServiceEndpoints = append(seps.ServiceEndpoints, sep)
		}
	}
	return seps
}
