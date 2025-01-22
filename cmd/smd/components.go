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
	"fmt"
	"strconv"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-smd/v2/internal/hmsds"
	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
)

////////////////////////////////////////////////////////////////////////////
//
// HMS Components- discovery/creation from Redfish Endpoint data
//
////////////////////////////////////////////////////////////////////////////

// Create a new array of Components based on a post-discover
// redfish endpoint discovery struct.
func (s *SmD) DiscoverComponentArray(rfEP *rf.RedfishEP) (*base.ComponentArray, error) {
	comps := new(base.ComponentArray)
	for _, chEP := range rfEP.Chassis.OIDs {
		comp := s.DiscoverComponentChassis(chEP)
		if comp != nil {
			comps.Components = append(comps.Components, comp)
		}
	}
	for _, sysEP := range rfEP.Systems.OIDs {
		comp := s.DiscoverComponentSystem(sysEP)
		if comp != nil {
			comps.Components = append(comps.Components, comp)
		}
	}
	for _, mEP := range rfEP.Managers.OIDs {
		comp := s.DiscoverComponentManager(mEP)
		if comp != nil {
			comps.Components = append(comps.Components, comp)
		}
	}
	for _, pduEP := range rfEP.RackPDUs.OIDs {
		comp := s.DiscoverComponentRackPDU(pduEP)
		if comp != nil {
			comps.Components = append(comps.Components, comp)
			for _, outEP := range pduEP.Outlets.OIDs {
				cout := s.DiscoverComponentOutlet(outEP)
				if cout != nil {
					comps.Components = append(comps.Components, cout)
				}
			}
		}
	}
	return comps, nil
}

// Use discovered data on a Redfish (not HMS) Chassis type to create
// an HMS Component representation.
func (s *SmD) DiscoverComponentChassis(chEP *rf.EpChassis) *base.Component {
	if chEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentChassis: EP: %s RF Subtype %s "+
			"not supported.", chEP.RfEndpointID, chEP.RedfishSubtype)
		return nil
	} else if chEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentChassis: Saw EP with bad status: %s",
			chEP.LastStatus)
		return nil
	}
	comp := new(base.Component)

	comp.ID = chEP.ID
	comp.Type = chEP.Type
	comp.State = chEP.State
	comp.Flag = chEP.Flag
	comp.Subtype = chEP.Subtype
	comp.Arch = chEP.Arch
	comp.NetType = chEP.NetType
	comp.Class = chEP.DefaultClass

	if comp.Class == "" {
		if comp.Type == base.Chassis.String() {
			comp.Class = base.ClassMountain.String()
			// Just incase our redfish endpoint didn't exist when our child
			// components were discovered, update them to be Mountain too.
			f := hmsds.ComponentFilter{
				Type: []string{
					base.NodeBMC.String(),
					base.NodeEnclosure.String(),
					base.Node.String(),
					base.RouterBMC.String(),
				},
			}
			children, err := s.db.GetComponentsQuery(&f, hmsds.FLTR_ID_ONLY, []string{comp.ID})
			if err != nil {
				s.LogAlways("DiscoverComponentChassis: Could not get child components for %s: %s", comp.ID, err)
			} else {
				if children != nil || len(children) > 0 {
					ids := make([]string, 0, len(children))
					for _, child := range children {
						ids = append(ids, child.ID)
					}
					_, err := s.db.BulkUpdateCompClass(ids, base.ClassMountain.String())
					if err != nil {
						s.LogAlways("DiscoverComponentChassis: Could not update child Class for %s: %s", comp.ID, err)
					}
				}
			}
		} else {
			// Get the ID for the parent Chassis BMC. Need to find the
			// Chassis first then go down a level to the chassisBMC
			p := comp.ID
			for {
				p = base.GetHMSCompParent(p)
				if p == "" {
					s.LogAlways("DiscoverComponentChassis: Could not determine ChassisBMC ID for %s", comp.ID)
					break
				}
				if base.GetHMSType(p) != base.Chassis {
					continue
				}
				chassisBmc := p + "b0"
				// Use the ChassisBMC RedfishEndpoint for existence
				// so we don't miss it if its discovery failed.
				cmp, err := s.db.GetRFEndpointByID(chassisBmc)
				if err != nil {
					s.LogAlways("DiscoverComponentChassis: Could not determine Class for %s: %s", comp.ID, err)
					break
				}
				// If the chassis BMC redfishEndpoint exists we are Mountain
				if cmp != nil {
					comp.Class = base.ClassMountain.String()
				} else {
					comp.Class = base.ClassRiver.String()
				}
				break
			}
		}
	}

	return comp
}

// Use discovered data on a Redfish (not HMS) System type to create
// an HMS Component representation.
func (s *SmD) DiscoverComponentSystem(sysEP *rf.EpSystem) *base.Component {
	if sysEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentSystem: EP: %s RF Subtype %s "+
			"not supported.", sysEP.RfEndpointID, sysEP.RedfishSubtype)
		return nil
	} else if sysEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentSystem: Saw EP with bad status: %s",
			sysEP.LastStatus)
		return nil
	}
	comp := new(base.Component)

	comp.ID = sysEP.ID
	comp.Type = sysEP.Type
	comp.State = sysEP.State
	comp.Flag = sysEP.Flag
	comp.Subtype = sysEP.Subtype
	comp.Arch = sysEP.Arch
	comp.NetType = sysEP.NetType

	newNID, defRole, defSubRole, defClass := s.GetCompDefaults(comp.ID, sysEP.DefaultRole, sysEP.DefaultSubRole, sysEP.DefaultClass)
	comp.Class = defClass
	comp.Role = defRole
	comp.SubRole = defSubRole
	comp.NID = json.Number(strconv.FormatUint(newNID, 10))

	if comp.Class == "" {
		// Get the ID for the parent Chassis BMC. Need to find the
		// Chassis first then go down a level to the chassisBMC
		p := comp.ID
		for {
			p = base.GetHMSCompParent(p)
			if p == "" {
				s.LogAlways("DiscoverComponentChassis: Could not determine ChassisBMC ID for %s", comp.ID)
				break
			}
			if base.GetHMSType(p) != base.Chassis {
				continue
			}
			chassisBmc := p + "b0"
			// Use the ChassisBMC RedfishEndpoint for existence
			// so we don't miss it if its discovery failed.
			cmp, err := s.db.GetRFEndpointByID(chassisBmc)
			if err != nil {
				s.LogAlways("DiscoverComponentChassis: Could not determine Class for %s: %s", comp.ID, err)
				break
			}
			// If the chassis BMC redfishEndpoint exists we are Mountain
			if cmp != nil {
				comp.Class = base.ClassMountain.String()
			} else {
				comp.Class = base.ClassRiver.String()
			}
			break
		}
	}
	return comp
}

// Use discovered data on a Redfish (not HMS) Manager type to create
// an HMS Component representation.
func (s *SmD) DiscoverComponentManager(mEP *rf.EpManager) *base.Component {
	if mEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentManager: EP: %s RF Subtype %s "+
			"not supported.", mEP.RfEndpointID, mEP.RedfishSubtype)
		return nil
	} else if mEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentManager: Saw EP with bad status: %s",
			mEP.LastStatus)
		return nil
	}
	comp := new(base.Component)

	comp.ID = mEP.ID
	comp.Type = mEP.Type
	comp.State = mEP.State
	comp.Flag = mEP.Flag
	comp.Subtype = mEP.Subtype
	comp.Arch = mEP.Arch
	comp.NetType = mEP.NetType
	comp.Class = mEP.DefaultClass

	if comp.Class == "" {
		if comp.Type == base.ChassisBMC.String() {
			comp.Class = base.ClassMountain.String()
		} else {
			// Get the ID for the parent Chassis BMC. Need to find the
			// Chassis first then go down a level to the chassisBMC
			p := comp.ID
			for {
				p = base.GetHMSCompParent(p)
				if p == "" {
					s.LogAlways("DiscoverComponentChassis: Could not determine ChassisBMC ID for %s", comp.ID)
					break
				}
				if base.GetHMSType(p) != base.Chassis {
					continue
				}
				chassisBmc := p + "b0"
				// Use the ChassisBMC RedfishEndpoint for existence
				// so we don't miss it if its discovery failed.
				cmp, err := s.db.GetRFEndpointByID(chassisBmc)
				if err != nil {
					s.LogAlways("DiscoverComponentChassis: Could not determine Class for %s: %s", comp.ID, err)
					break
				}
				// If the chassis BMC redfishEndpoint exists we are Mountain
				if cmp != nil {
					comp.Class = base.ClassMountain.String()
				} else {
					comp.Class = base.ClassRiver.String()
				}
				break
			}
		}
	}

	return comp
}

// Use discovered data on a Redfish (not HMS) PowerDistribution type to create
// an HMS Component representation.
func (s *SmD) DiscoverComponentRackPDU(pduEP *rf.EpPDU) *base.Component {
	if pduEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentRackPDU: EP: %s RF Subtype %s "+
			"not supported.", pduEP.RfEndpointID, pduEP.RedfishSubtype)
		return nil
	} else if pduEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentRackPDU: Saw EP with bad status: %s",
			pduEP.LastStatus)
		return nil
	}
	comp := new(base.Component)

	comp.ID = pduEP.ID
	comp.Type = pduEP.Type
	comp.State = pduEP.State
	comp.Flag = pduEP.Flag
	comp.Subtype = pduEP.Subtype
	comp.Arch = pduEP.Arch
	comp.NetType = pduEP.NetType
	comp.Class = pduEP.DefaultClass

	return comp
}

// Use discovered data on a Redfish (not HMS) Outlet type to create
// an HMS Component representation.
func (s *SmD) DiscoverComponentOutlet(outEP *rf.EpOutlet) *base.Component {
	if outEP.LastStatus == rf.RedfishSubtypeNoSupport {
		s.LogAlways("DiscoverComponentOutlet: EP: %s RF Subtype %s "+
			"not supported.", outEP.RfEndpointID, outEP.RedfishSubtype)
		return nil
	} else if outEP.LastStatus != rf.DiscoverOK {
		s.LogAlways("DiscoverComponentOutlet: Saw EP with bad status: %s",
			outEP.LastStatus)
		return nil
	}
	comp := new(base.Component)

	comp.ID = outEP.ID
	comp.Type = outEP.Type
	comp.State = outEP.State
	comp.Flag = outEP.Flag
	comp.Subtype = outEP.Subtype
	comp.Arch = outEP.Arch
	comp.NetType = outEP.NetType
	comp.Class = outEP.DefaultClass

	return comp
}

// Get default NID and Role from SLS or the uploaded NodeMaps in priority order.
func (s *SmD) GetCompDefaults(xname, defaultRole, defaultSubRole, defaultClass string) (uint64, string, string, string) {
	var (
		nid     uint64
		role    string
		subRole string
		class   string
	)
	if s.sls != nil {
		nodeInfo, err := s.sls.GetNodeInfo(xname)
		if err != nil {
			s.lg.Printf("GetCompDefaults(%s): SLS Lookup failure: %s", xname, err)
		} else {
			nid = uint64(nodeInfo.NID)
			role = nodeInfo.Role
			subRole = nodeInfo.SubRole
			class = nodeInfo.Class
		}
	}
	if nid == 0 || len(role) == 0 {
		m, err := s.db.GetNodeMapByID(xname)
		if err == nil && m != nil {
			if nid == 0 {
				nid = uint64(m.NID)
			}
			if len(role) == 0 {
				role = m.Role
				subRole = m.SubRole
			}
		}
		if err != nil {
			s.lg.Printf("GetCompDefaults(%s): Lookup failure: %s", xname, err)
		} else if m == nil {
			s.lg.Printf("GetCompDefaults(%s): No NodeMap", xname)
		}
	}
	if nid == 0 {
		nid = GetBogusNID(xname)
	}
	if len(role) == 0 {
		role = defaultRole
		subRole = defaultSubRole
	}
	if class == "" {
		class = defaultClass
	}
	return nid, role, subRole, class
}

// Get a bogus nid for an xname
func GetBogusNID(xname string) uint64 {
	var cab, chassis, slot, controller, node uint
	fmt.Sscanf(xname, "x%dc%ds%db%dn%d", &cab, &chassis, &slot, &controller, &node)
	nid := ((cab + 1) * 16384) + (chassis * 2048) + (slot * 32) + (controller * 4) + node
	return uint64(nid)
}
