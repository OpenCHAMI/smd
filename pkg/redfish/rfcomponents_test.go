// MIT License
//
// (C) Copyright [2019-2022,2024-2025] Hewlett Packard Enterprise Development LP
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

package rf

// TODO: Put most of this in sharedtest so it can be used to validate
//       higher-level discovery routines in sm and smd.

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

// RoundTrip method override
type RTFunc func(req *http.Request) *http.Response

// Implement RoundTrip interface by implementing RoundTrip method
func (f RTFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(f RTFunc) *hms_certs.HTTPClientPair {
	cp, _ := hms_certs.CreateHTTPClientPair("", 5)
	cp.InsecureClient.HTTPClient.Transport = RTFunc(f)
	return cp
}

const testXName = "x0c0s16b0"
const testDomain = "local"
const testFQDN = "[fd40:aaaa:bbbb:1000:240:a6ff:ee82:d7c2]"

const serviceRootStr = "/redfish/v1"

////////////////////////////////////////////////////////////////////////////
//
//  Test full discovery
//
///////////////////////////////////////////////////////////////////////////

// Verification data, so type specific fields can be added to a
// generic verification routine for any endpoint.
// *ActionCount should be -1 if no actions are expected.
//
// Fields not relevant should be left blank.  Omit ignored chassis
// types, only ones that should be discovered as HMS components should
// be listed.  Types/IDs(xnames) will be checked.
//
// This assumes for simplicity that multiple instances of the same type
// are mostly identical.  We will skip most checks for empty instances.
type RedfishEPVerifyInfo struct {
	SystemIds                    []string // In order of xnames/ordinals
	SystemActionCount            int
	SystemActionTargets          []string // Parallel array
	SystemExpectPowerInfo        bool
	SystemPowerControl           []*PowerControl
	ManagerId                    string
	ManagerType                  string
	ManagerActionCount           int
	ManagerActionTarget          string
	NodeEnclosureId              string
	NodeEnclosureActionCount     int
	NodeEnclosureActionTarget    string
	HSNBoardId                   string
	HSNBoardActionCount          int
	HSNBoardActionTarget         string
	ChassisEnclosureId           string
	ChassisEnclosureActionCount  int
	ChassisEnclosureActionTarget string
	ComputeBladeIds              []string // In order of xnames/ordinals
	ComputeBladeActionCount      int
	ComputeBladeActionTargets    []string // Parallel array
	RouterBladeIds               []string // In order of xnames/ordinals
	RouterBladeActionCount       int
	RouterBladeActionTargets     []string // Parallel array
	PDUIds                       []string // In order of xnames/ordinals
	PDUActionCount               int
	PDUActionTargets             []string // Parallel array
	OutletPDUIds                 []string // PDU Id for outlet, parallel array
	OutletIds                    []string // In order of xnames/ordinals
	OutletActionCountP           int
	OutletActionTargetsP         []string // Parallel array
	OutletActionCountR           int
	OutletActionTargetsR         []string // Parallel array
}

// Intel dummy endpoint 1
var TestRedfishEPInitIntel = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             testXName,
		Type:           "NodeBMC",
		Hostname:       "x0c0s16b0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		MACRequired:    false,
		RediscOnUpdate: false,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for Intel s2600* dummy endpoint 1
var IntelVerifyInfo = RedfishEPVerifyInfo{
	SystemIds:                 []string{"QSBP74304715"},
	SystemActionCount:         6,
	SystemActionTargets:       []string{"/redfish/v1/Systems/QSBP74304715/Actions/ComputerSystem.Reset"},
	ManagerId:                 "BMC",
	ManagerType:               "NodeBMC",
	ManagerActionCount:        1,
	ManagerActionTarget:       "/redfish/v1/Managers/BMC/Actions/Manager.Reset",
	NodeEnclosureId:           "RackMount",
	NodeEnclosureActionCount:  -1,
	NodeEnclosureActionTarget: "",
}

// Dell dummy endpoint 1
var TestRedfishEPInitDell = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             testXName,
		Type:           "NodeBMC",
		Hostname:       "x0c0s16b0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		MACRequired:    false,
		RediscOnUpdate: false,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for Dell dummy endpoint 1
var DellVerifyInfo = RedfishEPVerifyInfo{
	SystemIds:                 []string{"System.Embedded.1"},
	SystemActionCount:         5,
	SystemActionTargets:       []string{"/redfish/v1/Systems/System.Embedded.1/Actions/ComputerSystem.Reset"},
	ManagerId:                 "iDRAC.Embedded.1",
	ManagerType:               "NodeBMC",
	ManagerActionCount:        1,
	ManagerActionTarget:       "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Manager.Reset",
	NodeEnclosureId:           "System.Embedded.1",
	NodeEnclosureActionCount:  2,
	NodeEnclosureActionTarget: "/redfish/v1/Chassis/System.Embedded.1/Actions/Chassis.Reset",
}

// Cray Mountain chassis controller, aka CMM dummy endpoint
var TestRedfishEPInitCrayCMM = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             "x1c2b0",
		Type:           "ChassisBMC",
		Hostname:       "x1c2b0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		MACRequired:    false,
		RediscOnUpdate: true,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for Cray CMM dummy endpoint
var CrayCMM1VerifyInfo = RedfishEPVerifyInfo{
	SystemIds:                    []string{},
	SystemActionCount:            -1,
	SystemActionTargets:          []string{},
	ManagerId:                    "BMC",
	ManagerType:                  "ChassisBMC",
	ManagerActionCount:           2,
	ManagerActionTarget:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset",
	ChassisEnclosureId:           "Enclosure",
	ChassisEnclosureActionCount:  3,
	ChassisEnclosureActionTarget: "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset",
	ComputeBladeIds: []string{
		"Blade0",
		"Blade1",
		"Blade2",
		"Blade3",
		"Blade4",
		"Blade5",
		"Blade6",
		"Blade7",
	}, // In order of xnames/ordinals
	ComputeBladeActionCount: 3,
	ComputeBladeActionTargets: []string{
		"/redfish/v1/Chassis/Blade0/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Blade1/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Blade2/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Blade3/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Blade4/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Blade5/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Blade6/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Blade7/Actions/Chassis.Reset",
	}, // Parallel array
	RouterBladeIds: []string{
		"Perif0",
		"Perif1",
		"Perif2",
		"Perif3",
		"Perif4",
		"Perif5",
		"Perif6",
		"Perif7",
	},
	RouterBladeActionCount: 3,
	RouterBladeActionTargets: []string{
		"/redfish/v1/Chassis/Perif0/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Perif1/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Perif2/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Perif3/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Perif4/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Perif5/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Perif6/Actions/Chassis.Reset",
		"/redfish/v1/Chassis/Perif7/Actions/Chassis.Reset",
	}, // Parallel array
}

// Cray Mountain Node Card BMC dummy endpoint 1
var TestRedfishEPInitCrayNC = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             testXName,
		Type:           "NodeBMC",
		Hostname:       "x0c0s16b0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		RediscOnUpdate: true,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for Cray Mountain Node Card BMC dummy endpoint 1
var CrayNC1VerifyInfo = RedfishEPVerifyInfo{
	SystemIds: []string{
		"Node0",
		"Node1",
	},
	SystemActionCount: 3,
	SystemActionTargets: []string{
		"/redfish/v1/Systems/Node0/Actions/ComputerSystem.Reset",
		"/redfish/v1/Systems/Node1/Actions/ComputerSystem.Reset",
	},
	ManagerId:                 "BMC",
	ManagerType:               "NodeBMC",
	ManagerActionCount:        2,
	ManagerActionTarget:       "/redfish/v1/Managers/BMC/Actions/Manager.Reset",
	NodeEnclosureId:           "Enclosure",
	NodeEnclosureActionCount:  0,
	NodeEnclosureActionTarget: "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset",
	SystemExpectPowerInfo:     true,
	SystemPowerControl: []*PowerControl{
		&PowerControl{
			Name:               "Node Power Control",
			PowerCapacityWatts: 900,
			OEM: &PwrCtlOEM{
				Cray: &PwrCtlOEMCray{
					PowerIdleWatts: 250,
					PowerLimit: &CrayPwrLimit{
						Min: 350,
						Max: 850,
					},
					PowerResetWatts: 250,
				},
			},
			RelatedItem: []*ResourceID{
				&ResourceID{Oid: ""},
			},
		},
		&PowerControl{
			Name: "Accelerator0 Power Control",
			OEM: &PwrCtlOEM{
				Cray: &PwrCtlOEMCray{
					PowerIdleWatts: 100,
					PowerLimit: &CrayPwrLimit{
						Min: 200,
						Max: 350,
					},
				},
			},
		},
	},
}

// Cray Mountain Node Card BMC dummy endpoint 2
var TestRedfishEPInitCrayNC2 = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             testXName,
		Type:           "NodeBMC",
		Hostname:       "x1000c7s1b0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		RediscOnUpdate: true,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for Cray Mountain Node Card BMC dummy endpoint 2
var CrayNC2VerifyInfo = RedfishEPVerifyInfo{
	SystemIds: []string{
		"Node0",
		"Node1",
	},
	SystemActionCount: 3,
	SystemActionTargets: []string{
		"/redfish/v1/Systems/Node0/Actions/ComputerSystem.Reset",
		"/redfish/v1/Systems/Node1/Actions/ComputerSystem.Reset",
	},
	ManagerId:                 "BMC",
	ManagerType:               "NodeBMC",
	ManagerActionCount:        2,
	ManagerActionTarget:       "/redfish/v1/Managers/BMC/Actions/Manager.Reset",
	NodeEnclosureId:           "Enclosure",
	NodeEnclosureActionCount:  0,
	NodeEnclosureActionTarget: "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset",
	SystemExpectPowerInfo:     true,
	SystemPowerControl: []*PowerControl{
		&PowerControl{
			Name:               "Node Power Control",
			PowerCapacityWatts: 900,
			OEM: &PwrCtlOEM{
				Cray: &PwrCtlOEMCray{
					PowerIdleWatts: 250,
					PowerLimit: &CrayPwrLimit{
						Min: 350,
						Max: 850,
					},
					PowerResetWatts: 250,
				},
			},
			RelatedItem: []*ResourceID{
				&ResourceID{Oid: ""},
			},
		},
		&PowerControl{
			Name: "Accelerator0 Power Control",
			OEM: &PwrCtlOEM{
				Cray: &PwrCtlOEMCray{
					PowerIdleWatts: 100,
					PowerLimit: &CrayPwrLimit{
						Min: 200,
						Max: 350,
					},
				},
			},
		},
	},
}

// Cray Mountain Router Card BMC dummy endpoint 1
var TestRedfishEPInitCrayRC = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             "x0c0r16b0",
		Type:           "RouterBMC",
		Hostname:       "x0c0r16b0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		RediscOnUpdate: true,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for Cray Mountain Router Card/TOR BMC dummy endpoint 1
var CrayRC1VerifyInfo = RedfishEPVerifyInfo{
	ManagerId:            "BMC",
	ManagerType:          "RouterBMC",
	ManagerActionCount:   2,
	ManagerActionTarget:  "/redfish/v1/Managers/BMC/Actions/Manager.Reset",
	HSNBoardId:           "Enclosure",
	HSNBoardActionCount:  3,
	HSNBoardActionTarget: "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset",
}

// RTS PDU BMC dummy endpoint 1
var TestRedfishEPInitRtsCabPDUController = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             "x0m0",
		Type:           "CabinetPDUController",
		Hostname:       "x0m0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		RediscOnUpdate: true,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for Cray Mountain Router Card/TOR BMC dummy endpoint 1
var RtsCabPDUControllerVerifyInfo = RedfishEPVerifyInfo{
	ManagerId:           "BMC",
	ManagerType:         "CabinetPDUController",
	ManagerActionCount:  2,
	ManagerActionTarget: "/redfish/v1/Managers/BMC/Actions/Manager.Reset",
	PDUIds:              []string{"1"}, // In order of xnames/ordinals
	PDUActionCount:      -1,
	PDUActionTargets:    []string{}, // Parallel array
	OutletPDUIds: []string{
		"1", "1", "1", "1", "1",
		"1", "1", "1",
		"1", "1", "1",
	},
	OutletIds: []string{
		"A1", "A2", "A3", "A4", "A5",
		"B1", "B2", "B3",
		"C1", "C2", "C3",
	},
	OutletActionCountP: 0,
	OutletActionTargetsP: []string{
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A2/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A3/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A4/Outlet.PowerControl",
		"",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B1/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B2/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B3/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C1/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C2/Outlet.PowerControl",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C3/Outlet.PowerControl",
	}, // Parallel array
	OutletActionCountR: 0,
	OutletActionTargetsR: []string{
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1/Outlet.ResetStatistics",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A2/Outlet.ResetStatistics",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A3/Outlet.ResetStatistics",
		"",
		"",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B1/Outlet.ResetStatistics",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B2/Outlet.ResetStatistics",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B3/Outlet.ResetStatistics",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C1/Outlet.ResetStatistics",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C2/Outlet.ResetStatistics",
		"/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C3/Outlet.ResetStatistics",
	}, // Parallel array
}

// Intel dummy endpoint 1
var TestRedfishEPInitGBT = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             testXName,
		Type:           "NodeBMC",
		Hostname:       "x3309c0s9b1",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		MACRequired:    false,
		RediscOnUpdate: false,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     0,
}

// Verification data for GBT dummy endpoint
var GBTVerifyInfo = RedfishEPVerifyInfo{
	SystemIds:           []string{"Self"},
	SystemActionCount:   4,
	SystemActionTargets: []string{"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"},
}

// HPE Proliant iLO/BMC dummy endpoint 1
var TestRedfishEPInitPRLT = RedfishEP{
	RedfishEPDescription: RedfishEPDescription{
		ID:             testXName,
		Type:           "NodeBMC",
		Hostname:       "x3000c0s7b0",
		Domain:         testDomain,
		FQDN:           testFQDN,
		Enabled:        true,
		User:           "root",
		Password:       "********",
		UseSSDP:        false,
		RediscOnUpdate: true,
		DiscInfo: DiscoveryInfo{
			LastStatus: NotYetQueried,
		},
	},
	ServiceRootURL: testFQDN + "/redfish/v1",
	RedfishType:    "ServiceRoot",
	OdataID:        "/redfish/v1",
	NumSystems:     1,
}

// Verification data for Cray Mountain Node Card BMC dummy endpoint 1
var PRLTVerifyInfo = RedfishEPVerifyInfo{
	SystemIds:           []string{"1"},
	SystemActionCount:   6,
	SystemActionTargets: []string{"/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"},
}

// Do a mock discovery of the two main HW types we're seen so far,
// which demonstrate all of the existing workarounds needed to discover
// them.  This should touch just about all tv.ManagerIdhe code in rfcomponents.
func TestGetRootInfo(t *testing.T) {

	// GBT Endpoint
	clientGBT1 := NewTestClient(NewRTFuncGBT1())
	gbtEP1 := TestRedfishEPInitGBT
	gbtEP1.client = clientGBT1
	gbtEP1.GetRootInfo()

	if gbtEP1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 1: (GBT): FAILED discovery, LastStatus: %s",
			gbtEP1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 1: (GBT): PASSED discovery, LastStatus: %s",
			gbtEP1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&gbtEP1, GBTVerifyInfo); err != nil {
			t.Errorf("Testcase 1: (GBT): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 1: (GBT): PASSED verification")
		}
	}

	// Intel Endpoint (s2600*)
	clientIntel1 := NewTestClient(NewRTFuncIntel1())
	intelEP1 := TestRedfishEPInitIntel
	intelEP1.client = clientIntel1
	intelEP1.GetRootInfo()

	if intelEP1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 2: (Intel): FAILED discovery, LastStatus: %s",
			intelEP1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 2: (Intel): PASSED discovery, LastStatus: %s",
			intelEP1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&intelEP1, IntelVerifyInfo); err != nil {
			t.Errorf("Testcase 2: (Intel): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 2: (Intel): PASSED verification")
		}
	}

	// Dell endpoint
	clientDell1 := NewTestClient(NewRTFuncDell1())
	dellEP1 := TestRedfishEPInitDell
	dellEP1.client = clientDell1
	dellEP1.GetRootInfo()

	if dellEP1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 3: (Dell):  FAILED discovery, LastStatus: %s",
			dellEP1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 3: (Dell):  PASSED discovery, LastStatus: %s",
			dellEP1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&dellEP1, DellVerifyInfo); err != nil {
			t.Errorf("Testcase 3: (Dell): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 3: (Dell): PASSED verification")
		}
	}

	// Cray Mountain Chassis controller i.e. CMM
	clientCrayCMM1 := NewTestClient(NewRTFuncCrayCMM1())
	crayCMM1 := TestRedfishEPInitCrayCMM
	crayCMM1.client = clientCrayCMM1
	crayCMM1.GetRootInfo()

	if crayCMM1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 4: (CrayChassis):  FAILED discovery, LastStatus: %s",
			crayCMM1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 4: (CrayChassis):  PASSED discovery, LastStatus: %s",
			crayCMM1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&crayCMM1, CrayCMM1VerifyInfo); err != nil {
			t.Errorf("Testcase 4: (CrayChassis): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 4: (CrayChassis): PASSED verification")
		}
	}

	// Cray Mountain Node Card BMC
	clientCrayNC1 := NewTestClient(NewRTFuncCrayNC1())
	CrayNC1 := TestRedfishEPInitCrayNC
	CrayNC1.client = clientCrayNC1
	CrayNC1.GetRootInfo()

	if CrayNC1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 5: (CrayNC1):  FAILED discovery, LastStatus: %s",
			CrayNC1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 5: (CrayNC1):  PASSED discovery, LastStatus: %s",
			CrayNC1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&CrayNC1, CrayNC1VerifyInfo); err != nil {
			t.Errorf("Testcase 5: (CrayNC1): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 5: (CrayNC1): PASSED verification")
		}
	}

	// Cray Mountain Node Card BMC 2
	// with changes to support new Systems/NodeX Actions
	clientCrayNC2 := NewTestClient(NewRTFuncCrayNC2())
	CrayNC2 := TestRedfishEPInitCrayNC2
	CrayNC2.client = clientCrayNC2
	CrayNC2.GetRootInfo()

	if CrayNC2.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 6: (CrayNC2):  FAILED discovery, LastStatus: %s",
			CrayNC2.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 6: (CrayNC2):  PASSED discovery, LastStatus: %s",
			CrayNC2.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&CrayNC2, CrayNC2VerifyInfo); err != nil {
			t.Errorf("Testcase 6: (CrayNC2): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 6: (CrayNC2): PASSED verification")
		}
	}
	// Cray Router Card/TOR BMC
	clientCrayRC1 := NewTestClient(NewRTFuncCrayRC1())
	CrayRC1 := TestRedfishEPInitCrayRC
	CrayRC1.client = clientCrayRC1
	CrayRC1.GetRootInfo()

	if CrayRC1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 7: (CrayRC1):  FAILED discovery, LastStatus: %s",
			CrayRC1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 7: (CrayRC1):  PASSED discovery, LastStatus: %s",
			CrayRC1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&CrayRC1, CrayRC1VerifyInfo); err != nil {
			t.Errorf("Testcase 7: (CrayRC1): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 7: (CrayRC1): PASSED verification")
		}
	}

	// RTS CabinetPDUController
	clientRtsPDU1 := NewTestClient(NewRTFuncRtsPDU1())
	RtsPDU1 := TestRedfishEPInitRtsCabPDUController
	RtsPDU1.client = clientRtsPDU1
	RtsPDU1.GetRootInfo()

	if RtsPDU1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 8: (RtsPDU1):  FAILED discovery, LastStatus: %s",
			RtsPDU1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 8: (RtsPDU1):  PASSED discovery, LastStatus: %s",
			RtsPDU1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&RtsPDU1, RtsCabPDUControllerVerifyInfo); err != nil {
			t.Errorf("Testcase 8: (RtsPDU1): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 8: (RtsPDU1): PASSED verification")
		}
	}

	// HPE Proliant Endpoint
	clientPRLT1 := NewTestClient(NewRTFuncPRLT1())
	prltEP1 := TestRedfishEPInitPRLT
	prltEP1.client = clientPRLT1
	prltEP1.GetRootInfo()

	if prltEP1.DiscInfo.LastStatus != DiscoverOK {
		t.Errorf("Testcase 9: (PRLT): FAILED discovery, LastStatus: %s",
			prltEP1.DiscInfo.LastStatus)
	} else {
		t.Logf("Testcase 9: (PRLT): PASSED discovery, LastStatus: %s",
			prltEP1.DiscInfo.LastStatus)
		if err := VerifyGetRootInfo(&prltEP1, PRLTVerifyInfo); err != nil {
			t.Errorf("Testcase 9: (PRLT): FAILED verfication: %s", err)
		} else {
			t.Logf("Testcase 9: (PRLT): PASSED verification")
		}
	}
}

// Check System, Manager, and Chassis.  Make sure status is OK, actions are
// set, etc.   Uses RedfishEPVerifyInfo as template for endpoint-dependent
// info (e.g. different Id names, actions, etc.).
func VerifyGetRootInfo(e *RedfishEP, v RedfishEPVerifyInfo) error {
	// System tests
	for i, sysId := range v.SystemIds {
		if s, ok := e.Systems.OIDs[sysId]; ok == true {
			if s.LastStatus != DiscoverOK {
				return fmt.Errorf(sysId + ": bad LastStatus: %s", s.LastStatus)
			}
			if s.Actions == nil {
				if v.SystemActionCount != -1 {
					return fmt.Errorf("%s: nil Action struct", sysId)
				}
			} else if len(s.Actions.ComputerSystemReset.AllowableValues) !=
				v.SystemActionCount ||
				s.Actions.ComputerSystemReset.Target !=
					v.SystemActionTargets[i] {
				return fmt.Errorf("SystemId: %s: bad Target/AllowableValues", sysId)
			}
			if v.SystemExpectPowerInfo {
				if len(s.PowerCtl) != len(v.SystemPowerControl) {
					return fmt.Errorf("%s: Bad powerControl info mismatched array length '%s'",
						sysId, s.ID)
				}
				for j, pCtl := range s.PowerCtl {
					vPCtl := v.SystemPowerControl[j]
					if pCtl.Name != vPCtl.Name {
						return fmt.Errorf("%s: Bad powerControl Name %s != %s for '%s'",
							sysId, pCtl.Name, vPCtl.Name, s.ID)
					}
					if pCtl.PowerCapacityWatts != vPCtl.PowerCapacityWatts {
						return fmt.Errorf("%s: Bad powerControl PowerCapacityWatts %d != %d for '%s'",
							sysId, pCtl.PowerCapacityWatts, vPCtl.PowerCapacityWatts, s.ID)
					}
					if (pCtl.OEM == nil) != (vPCtl.OEM == nil) {
						return fmt.Errorf("%s: Bad powerControl OEM nil struct for '%s'",
							sysId, s.ID)
					}
					if pCtl.OEM != nil {
						if (pCtl.OEM.Cray == nil) != (vPCtl.OEM.Cray == nil) {
							return fmt.Errorf("%s: Bad powerControl OEM Cray nil struct for '%s'",
								sysId, s.ID)
						}
						if pCtl.OEM.Cray != nil {
							if pCtl.OEM.Cray.PowerIdleWatts != vPCtl.OEM.Cray.PowerIdleWatts {
								return fmt.Errorf("%s: Bad powerControl OEM Cray PowerIdleWatts %d != %d for '%s'",
									sysId, pCtl.OEM.Cray.PowerIdleWatts, vPCtl.OEM.Cray.PowerIdleWatts, s.ID)
							}
							if pCtl.OEM.Cray.PowerResetWatts != vPCtl.OEM.Cray.PowerResetWatts {
								return fmt.Errorf("%s: Bad powerControl OEM Cray PowerResetWatts %d != %d for '%s'",
									sysId, pCtl.OEM.Cray.PowerResetWatts, vPCtl.OEM.Cray.PowerResetWatts, s.ID)
							}
							if (pCtl.OEM.Cray.PowerLimit == nil) != (vPCtl.OEM.Cray.PowerLimit == nil) {
								return fmt.Errorf("%s: Bad powerControl OEM Cray PowerLimit nil struct for '%s'",
									sysId, s.ID)
							}
							if pCtl.OEM.Cray.PowerLimit != nil {
								if pCtl.OEM.Cray.PowerLimit.Min != vPCtl.OEM.Cray.PowerLimit.Min {
									return fmt.Errorf("%s: Bad powerControl OEM Cray PowerLimit Min %d != %d for '%s'",
										sysId, pCtl.OEM.Cray.PowerLimit.Min, vPCtl.OEM.Cray.PowerLimit.Min, s.ID)
								}
								if pCtl.OEM.Cray.PowerLimit.Max != vPCtl.OEM.Cray.PowerLimit.Max {
									return fmt.Errorf("%s: Bad powerControl OEM Cray PowerLimit Max %d != %d for '%s'",
										sysId, pCtl.OEM.Cray.PowerLimit.Max, vPCtl.OEM.Cray.PowerLimit.Max, s.ID)
								}
							}
						}
					}
					if len(pCtl.RelatedItem) != len(vPCtl.RelatedItem) {
						return fmt.Errorf("%s: Bad powerControl RelatedItem mismatched array length '%s'",
							sysId, s.ID)
					}
				}
			}
			// Verify xname and type
			stype := xnametypes.GetHMSType(s.ID)
			if stype != xnametypes.Node || s.Type != stype.String() {
				return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
					s.ID, s.Type, sysId)
			}
			if s.Ordinal != i {
				return fmt.Errorf("%s: Bad ordinal %d != %d for '%s'",
					sysId, s.Ordinal, i, s.ID)
			}
		} else {
			return fmt.Errorf("systemId %s did not exist", sysId)
		}

	}
	// Manager tests
	if m, ok := e.Managers.OIDs[v.ManagerId]; ok == true {
		if m.LastStatus != DiscoverOK {
			return fmt.Errorf("%s: bad LastStatus: %s", v.ManagerId, m.LastStatus)
		}
		if m.Actions == nil {
			if v.ManagerActionCount != -1 {
				return fmt.Errorf("%s: nil Action struct", v.ManagerId)
			}
		} else if len(m.Actions.ManagerReset.AllowableValues) !=
			v.ManagerActionCount ||
			m.Actions.ManagerReset.Target !=
				v.ManagerActionTarget {
			return fmt.Errorf("ManagerId: %s: bad Target/AllowableValues", v.ManagerId)
		}
		// Verify xname and type
		mtype := xnametypes.GetHMSTypeString(m.ID)
		if m.ID == "" || mtype != v.ManagerType || m.Type != v.ManagerType {
			return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
				m.ID, m.Type, v.ManagerId)
		}
	} else if v.ManagerId != "" {
		return fmt.Errorf("ManagerId %s did not exist", v.ManagerId)
	}
	// Chassis tests
	neId := v.NodeEnclosureId
	if c, ok := e.Chassis.OIDs[neId]; ok == true {
		if c.LastStatus != DiscoverOK {
			return fmt.Errorf("%s: bad LastStatus: %s", neId, c.LastStatus)
		}
		if c.Actions == nil {
			if v.NodeEnclosureActionCount != -1 {
				return fmt.Errorf("%s: nil Action struct", neId)
			}
		} else if len(c.Actions.ChassisReset.AllowableValues) !=
			v.NodeEnclosureActionCount ||
			c.Actions.ChassisReset.Target !=
				v.NodeEnclosureActionTarget {
			return fmt.Errorf("NodeEnclosureId: %s: bad Target/AllowableValues", neId)
		}
		// Verify xname and type
		ctype := xnametypes.GetHMSType(c.ID)
		if ctype != xnametypes.NodeEnclosure || c.Type != ctype.String() {
			return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
				c.ID, c.Type, neId)
		}
	} else if neId != "" {
		return fmt.Errorf("NodeEnclosureId %s did not exist", neId)
	}
	ceId := v.ChassisEnclosureId
	if c, ok := e.Chassis.OIDs[ceId]; ok == true {
		if c.LastStatus != DiscoverOK {
			return fmt.Errorf("%s: bad LastStatus: %s", ceId, c.LastStatus)
		}
		if c.Actions == nil {
			if v.ChassisEnclosureActionCount != -1 {
				return fmt.Errorf("%s: nil Action struct", ceId)
			}
		} else if len(c.Actions.ChassisReset.AllowableValues) !=
			v.ChassisEnclosureActionCount ||
			c.Actions.ChassisReset.Target !=
				v.ChassisEnclosureActionTarget {
			return fmt.Errorf("ChassisEnclosureId: %s: bad Target/AllowableValues", ceId)
		}
		// Verify xname and type
		ctype := xnametypes.GetHMSType(c.ID)
		if ctype != xnametypes.Chassis || c.Type != ctype.String() {
			return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
				c.ID, c.Type, ceId)
		}
	} else if ceId != "" {
		return fmt.Errorf("ChassisEnclosureId %s did not exist", ceId)
	}
	for i, cbId := range v.ComputeBladeIds {
		if cb, ok := e.Chassis.OIDs[cbId]; ok == true {
			if cb.LastStatus != DiscoverOK {
				return fmt.Errorf("%s: bad LastStatus: %s", cbId, cb.LastStatus)
			}
			if cb.Actions == nil {
				if v.ComputeBladeActionCount != -1 {
					return fmt.Errorf("%s: nil Action struct", cbId)
				}
			} else if len(cb.Actions.ChassisReset.AllowableValues) !=
				v.ComputeBladeActionCount ||
				cb.Actions.ChassisReset.Target !=
					v.ComputeBladeActionTargets[i] {
				return fmt.Errorf("ComputeBladeId: %s: bad Target/AllowableValues", cbId)
			}
			// Verify xname and type
			cbtype := xnametypes.GetHMSType(cb.ID)
			if cbtype != xnametypes.ComputeModule || cb.Type != cbtype.String() {
				return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
					cb.ID, cb.Type, cbId)
			}
			if cb.Ordinal != i {
				return fmt.Errorf("%s: Bad ordinal %d != %d for '%s'",
					cbId, cb.Ordinal, i, cb.ID)
			}
		} else if cbId != "" {
			return fmt.Errorf("compBladeId %s did not exist", cbId)
		}
	}
	for i, rbId := range v.RouterBladeIds {
		if rb, ok := e.Chassis.OIDs[rbId]; ok == true {
			if rb.LastStatus != DiscoverOK {
				return fmt.Errorf("%s: bad LastStatus: %s", rbId, rb.LastStatus)
			}
			if rb.Actions == nil {
				if v.RouterBladeActionCount != -1 {
					return fmt.Errorf("%s: nil Action struct", rbId)
				}
			} else if len(rb.Actions.ChassisReset.AllowableValues) !=
				v.RouterBladeActionCount ||
				rb.Actions.ChassisReset.Target !=
					v.RouterBladeActionTargets[i] {
				return fmt.Errorf("RouterBladeId: %s: bad Target/AllowableValues", rbId)
			}
			// Verify xname and type
			rbtype := xnametypes.GetHMSType(rb.ID)
			if rbtype != xnametypes.RouterModule || rb.Type != rbtype.String() {
				return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
					rb.ID, rb.Type, rbId)
			}
			if rb.Ordinal != i {
				return fmt.Errorf("%s: Bad ordinal %d != %d for '%s'",
					rbId, rb.Ordinal, i, rb.ID)
			}
		} else if rbId != "" {
			return fmt.Errorf("routerBladeId %s did not exist", rbId)
		}
	}
	for i, pduId := range v.PDUIds {
		if pdu, ok := e.RackPDUs.OIDs[pduId]; ok == true {
			if pdu.LastStatus != DiscoverOK {
				return fmt.Errorf("%s: bad LastStatus: %s", pduId, pdu.LastStatus)
			}
			if pdu.Actions == nil {
				if v.PDUActionCount != -1 {
					return fmt.Errorf("%s: nil Action struct", pduId)
				}
			}
			// No current actions to verify.

			// Verify xname and type
			pdutype := xnametypes.GetHMSType(pdu.ID)
			if pdutype != xnametypes.CabinetPDU || pdu.Type != pdutype.String() {
				return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
					pdu.ID, pdu.Type, pduId)
			}
			if pdu.Ordinal != i {
				return fmt.Errorf("%s: Bad ordinal %d != %d for '%s'",
					pduId, pdu.Ordinal, i, pdu.ID)
			}
		} else if pduId != "" {
			return fmt.Errorf("PDU %s did not exist", pduId)
		}

	}
	for i, outId := range v.OutletIds {
		pdu, ok := e.RackPDUs.OIDs[v.OutletPDUIds[i]]
		if !ok {
			return fmt.Errorf("PDU %s did not exist for outlet", v.OutletPDUIds[i])
		}
		if out, ok := pdu.Outlets.OIDs[outId]; ok == true {
			if out.LastStatus != DiscoverOK {
				return fmt.Errorf("%s: bad LastStatus: %s", outId, out.LastStatus)
			}
			if out.Actions == nil {
				if v.OutletActionCountP > 0 || v.OutletActionCountR > 0 {
					return fmt.Errorf("%s: nil Action struct", outId)
				}
				if v.OutletActionCountP == 0 &&
					v.OutletActionTargetsP[i] != "" {
					return fmt.Errorf("%s: nil Action struct", outId)
				}
				if v.OutletActionCountR == 0 &&
					v.OutletActionTargetsR[i] != "" {
					return fmt.Errorf("%s: nil Action struct", outId)
				}
			} else {
				if out.Actions.PowerControl == nil {
					if v.OutletActionTargetsP[i] != "" {
						return fmt.Errorf("%s: nil PowerControl struct", outId)
					}
				} else {
					if out.Actions.PowerControl.Target !=
						v.OutletActionTargetsP[i] {
						return fmt.Errorf(outId+": mismatch PowerControl target: %s",
							out.Actions.PowerControl.Target)
					}
					if v.OutletActionCountP > 0 &&
						len(out.Actions.PowerControl.AllowableValues) !=
							v.OutletActionCountP {
						return fmt.Errorf(outId+": bad AllowableValuesP count (%d)",
							len(out.Actions.PowerControl.AllowableValues))
					}
				}
				if out.Actions.ResetStatistics == nil {
					if v.OutletActionTargetsR[i] != "" {
						return fmt.Errorf("%s: nil ResetStats struct", outId)
					}
				} else {
					if out.Actions.ResetStatistics.Target !=
						v.OutletActionTargetsR[i] {
						return fmt.Errorf(outId+": mismatch ResetStatistics target: %s",
							out.Actions.ResetStatistics.Target)
					}
					if v.OutletActionCountR > 0 &&
						len(out.Actions.ResetStatistics.AllowableValues) !=
							v.OutletActionCountR {
						return fmt.Errorf(outId+": bad AllowableValuesR count (%d)",
							len(out.Actions.ResetStatistics.AllowableValues))
					}
				}
			}
			// No current actions to verify.

			// Verify xname and type
			outtype := xnametypes.GetHMSType(out.ID)
			if outtype != xnametypes.CabinetPDUPowerConnector || out.Type != outtype.String() {
				return fmt.Errorf("bad xname ID ('%s') or Type ('%s') for %s\n",
					out.ID, out.Type, outId)
			}
			if out.Ordinal != i {
				return fmt.Errorf("%s: Bad ordinal %d != %d for '%s'",
					outId, out.Ordinal, i, out.ID)
			}
		} else if outId != "" {
			return fmt.Errorf("Outlet %s did not exist", outId)
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////
//
// Path and matching payload variables - Mock responses for the given URI
//
/////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////
//                         Proliant - Mock Client
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncPRLT1() RTFunc {
	return func(req *http.Request) *http.Response {
    defer base.DrainAndCloseRequestBody(req)

		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathPRLT_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_redfish_v1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_account_service:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_account_service)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_session_service:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_session_service)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_event_service:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_event_service)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_task_service:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_task_service)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_update_service:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_update_service)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_managers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_managers)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_managers_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_managers_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_managers_1_ethernet_interfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_managers_1_ethernet_interfaces)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_managers_1_ethernet_interfaces_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_managers_1_ethernet_interfaces_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_managers_1_ethernet_interfaces_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_managers_1_ethernet_interfaces_2)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_managers_1_ethernet_interfaces_3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_managers_1_ethernet_interfaces_3)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis_1_power:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis_1_power)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis_1_network_adapters:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis_1_network_adapters)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis_1_devices:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis_1_devices)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis_1_devices_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis_1_devices_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis_1_devices_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis_1_devices_2)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_chassis_1_devices_3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_chassis_1_devices_3)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_ethernet_interfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_ethernet_interfaces)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_ethernet_interfaces_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_ethernet_interfaces_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_ethernet_interfaces_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_ethernet_interfaces_2)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_processors:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_processors)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_processors_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_processors_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_memory:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_memory)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_memory_proc1dimm1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_memory_proc1dimm1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_memory_proc1dimm2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_memory_proc1dimm2)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_memory_proc1dimm3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_memory_proc1dimm3)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_memory_proc1dimm4:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_memory_proc1dimm4)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathPRLT_systems_1_storage:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadPRLT_systems_1_storage)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString("")),

				Header: make(http.Header),
			}
		}
	}
}

const testPathPRLT_redfish_v1 = "/redfish/v1"

const testPayloadPRLT_redfish_v1 = `{
        "@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
        "@odata.etag": "W/\"F583FFB0\"",
        "@odata.id": "/redfish/v1/",
        "@odata.type": "#ServiceRoot.v1_5_1.ServiceRoot",
        "Id": "RootService",
        "AccountService": {
          "@odata.id": "/redfish/v1/AccountService/"
        },
        "Chassis": {
          "@odata.id": "/redfish/v1/Chassis/"
        },
        "EventService": {
          "@odata.id": "/redfish/v1/EventService/"
        },
        "JsonSchemas": {
          "@odata.id": "/redfish/v1/JsonSchemas/"
        },
        "Links": {
          "Sessions": {
            "@odata.id": "/redfish/v1/SessionService/Sessions/"
          }
        },
        "Managers": {
          "@odata.id": "/redfish/v1/Managers/"
        },
        "Name": "HPE RESTful Root Service",
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeiLOServiceExt.HpeiLOServiceExt",
            "@odata.type": "#HpeiLOServiceExt.v2_3_0.HpeiLOServiceExt",
            "Links": {
              "ResourceDirectory": {
                "@odata.id": "/redfish/v1/ResourceDirectory/"
              }
            },
            "Manager": [
              {
                "DefaultLanguage": "en",
                "FQDN": "ILOMXQ019020H.",
                "HostName": "ILOMXQ019020H",
                "Languages": [
                  {
                    "Language": "en",
                    "TranslationName": "English",
                    "Version": "2.14"
                  }
                ],
                "ManagerFirmwareVersion": "2.14",
                "ManagerType": "iLO 5",
                "Status": {
                  "Health": "OK"
                }
              }
            ],
            "Moniker": {
              "ADVLIC": "iLO Advanced",
              "BMC": "iLO",
              "BSYS": "BladeSystem",
              "CLASS": "Baseboard Management Controller",
              "FEDGRP": "DEFAULT",
              "IPROV": "Intelligent Provisioning",
              "PRODABR": "iLO",
              "PRODFAM": "Integrated Lights-Out",
              "PRODGEN": "iLO 5",
              "PRODNAM": "Integrated Lights-Out 5",
              "PRODTAG": "HPE iLO 5",
              "STDLIC": "iLO Standard",
              "SUMABR": "SUM",
              "SUMGR": "Smart Update Manager",
              "SYSFAM": "ProLiant",
              "VENDABR": "HPE",
              "VENDNAM": "Hewlett Packard Enterprise",
              "WWW": "www.hpe.com",
              "WWWAHSV": "www.hpe.com/servers/ahsv",
              "WWWBMC": "www.hpe.com/info/ilo",
              "WWWDOC": "www.hpe.com/support/ilo-docs",
              "WWWERS": "www.hpe.com/services/getconnected",
              "WWWGLIS": "reserved for liconf URI",
              "WWWIOL": "www.hpe.com/info/insightonline",
              "WWWLIC": "www.hpe.com/info/ilo",
              "WWWLML": "www.hpe.com/support",
              "WWWPASS": "www.hpe.com/support/hpesc",
              "WWWPRV": "www.hpe.com/info/privacy",
              "WWWQSPEC": "www.hpe.com/info/qs",
              "WWWRESTDOC": "www.hpe.com/support/restfulinterface/docs",
              "WWWSUP": "www.hpe.com/support/ilo5",
              "WWWSWLIC": "www.hpe.com/software/SWLicensing"
            },
            "Sessions": {
              "CertCommonName": "ILOMXQ019020H",
              "CertificateLoginEnabled": false,
              "KerberosEnabled": false,
              "LDAPAuthLicenced": true,
              "LDAPEnabled": false,
              "LocalLoginEnabled": true,
              "LoginFailureDelay": 0,
              "LoginHint": {
                "Hint": "POST to /Sessions to login using the following JSON object:",
                "HintPOSTData": {
                  "Password": "password",
                  "UserName": "username"
                }
              },
              "SecurityOverride": false,
              "ServerName": ""
            },
            "System": [
              {
                "Status": {
                  "Health": "OK"
                }
              }
            ],
            "Time": "2021-01-12T16:27:02Z"
          }
        },
        "Product": "ProLiant DL325 Gen10 Plus",
        "ProtocolFeaturesSupported": {
          "ExpandQuery": {
            "ExpandAll": false,
            "Levels": true,
            "Links": false,
            "MaxLevels": 1,
            "NoLinks": true
          },
          "FilterQuery": true,
          "OnlyMemberQuery": true,
          "SelectQuery": false
        },
        "RedfishVersion": "1.6.0",
        "Registries": {
          "@odata.id": "/redfish/v1/Registries/"
        },
        "SessionService": {
          "@odata.id": "/redfish/v1/SessionService/"
        },
        "Systems": {
          "@odata.id": "/redfish/v1/Systems/"
        },
        "Tasks": {
          "@odata.id": "/redfish/v1/TaskService/"
        },
        "UUID": "9bd8d4d5-9904-541e-bee4-997455a4bcf7",
        "UpdateService": {
          "@odata.id": "/redfish/v1/UpdateService/"
        },
        "Vendor": "HPE"
      }
`
const testPathPRLT_managers = "/redfish/v1/Managers/"

const testPayloadPRLT_managers = `{
   "@odata.context": "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
   "@odata.etag": "W/\"AA6D42B0\"",
   "@odata.id": "/redfish/v1/Managers",
   "@odata.type": "#ManagerCollection.ManagerCollection",
   "Description": "Managers view",
   "Name": "Managers",
   "Members": [
     {
       "@odata.id": "/redfish/v1/Managers/1"
     }
   ],
   "Members@odata.count": 1
 }
`
const testPathPRLT_account_service = "/redfish/v1/AccountService/"

const testPayloadPRLT_account_service = `{
        "@odata.context": "/redfish/v1/$metadata#AccountService.AccountService",
        "@odata.etag": "W/\"DF31857A\"",
        "@odata.id": "/redfish/v1/AccountService/",
        "@odata.type": "#AccountService.v1_3_0.AccountService",
        "Id": "AccountService",
        "Accounts": {
          "@odata.id": "/redfish/v1/AccountService/Accounts/"
        },
        "ActiveDirectory": {
          "AccountProviderType": "ActiveDirectoryService",
          "Authentication": {
            "AuthenticationType": "UsernameAndPassword",
            "Password": null,
            "Username": ""
          },
          "RemoteRoleMapping": [
            {
              "LocalRole": "dirgroupb3d8954f6ebbe735764e9f7c",
              "RemoteGroup": "Administrators"
            },
            {
              "LocalRole": "dirgroup9d4546a03a03bb977c03086a",
              "RemoteGroup": "Authenticated Users:S-1-5-11"
            }
          ],
          "ServiceAddresses": [
            ""
          ],
          "ServiceEnabled": false
        },
        "Description": "iLO User Accounts",
        "LDAP": {
          "AccountProviderType": "ActiveDirectoryService",
          "Authentication": {
            "AuthenticationType": "UsernameAndPassword",
            "Password": null,
            "Username": ""
          },
          "LDAPService": {
            "SearchSettings": {}
          },
          "RemoteRoleMapping": [
            {
              "LocalRole": "dirgroupb3d8954f6ebbe735764e9f7c",
              "RemoteGroup": "Administrators"
            },
            {
              "LocalRole": "dirgroup9d4546a03a03bb977c03086a",
              "RemoteGroup": "Authenticated Users:S-1-5-11"
            }
          ],
          "ServiceAddresses": [
            ""
          ],
          "ServiceEnabled": false
        },
        "LocalAccountAuth": "Enabled",
        "Name": "Account Service",
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeiLOAccountService.HpeiLOAccountService",
            "@odata.id": "/redfish/v1/AccountService/",
            "@odata.type": "#HpeiLOAccountService.v2_3_0.HpeiLOAccountService",
            "Id": "AccountService",
            "Actions": {
              "#HpeiLOAccountService.ImportKerberosKeytab": {
                "target": "/redfish/v1/AccountService/Actions/Oem/Hpe/HpeiLOAccountService.ImportKerberosKeytab/"
              }
            },
            "AuthFailureDelayTimeSeconds": 10,
            "AuthFailureLoggingThreshold": 3,
            "AuthFailuresBeforeDelay": 1,
            "DefaultPassword": null,
            "DefaultUserName": null,
            "DirectorySettings": {
              "LdapAuthenticationMode": "Disabled",
              "LdapCaCertificateLoaded": false,
              "LdapCaCertificates": {
                "@odata.id": "/redfish/v1/AccountService/ExternalAccountProviders/LDAP/Certificates/"
              },
              "LdapServerPort": 636
            },
            "DirectoryTest": {
              "@odata.id": "/redfish/v1/AccountService/DirectoryTest/"
            },
            "EnforcePasswordComplexity": false,
            "KerberosSettings": {
              "KDCServerPort": 88,
              "KerberosRealm": ""
            },
            "MinPasswordLength": 8
          }
        },
        "Roles": {
          "@odata.id": "/redfish/v1/AccountService/Roles/"
        }
      }
`
const testPathPRLT_session_service = "/redfish/v1/SessionService/"

const testPayloadPRLT_session_service = `{
        "@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
        "@odata.etag": "W/\"ED13A396\"",
        "@odata.id": "/redfish/v1/SessionService/",
        "@odata.type": "#SessionService.v1_0_0.SessionService",
        "Id": "SessionService",
        "Description": "Session Service",
        "Name": "Session Service",
        "ServiceEnabled": true,
        "Sessions": {
          "@odata.id": "/redfish/v1/SessionService/Sessions/"
        },
        "Status": {
          "Health": "OK",
          "State": "Enabled"
        }
      }
`
const testPathPRLT_event_service = "/redfish/v1/EventService/"

const testPayloadPRLT_event_service = `{
        "@odata.context": "/redfish/v1/$metadata#EventService.EventService",
        "@odata.etag": "W/\"0427F3F5\"",
        "@odata.id": "/redfish/v1/EventService/",
        "@odata.type": "#EventService.v1_0_8.EventService",
        "Id": "EventService",
        "Actions": {
          "#EventService.SubmitTestEvent": {
            "target": "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent/"
          }
        },
        "DeliveryRetryAttempts": 3,
        "DeliveryRetryIntervalSeconds": 30,
        "Description": "Event Subscription service",
        "EventTypesForSubscription": [
          "StatusChange",
          "ResourceUpdated",
          "ResourceAdded",
          "ResourceRemoved",
          "Alert"
        ],
        "Name": "Event Service",
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeEventService.HpeEventService",
            "@odata.id": "/redfish/v1/EventService/",
            "@odata.type": "#HpeEventService.v2_1_0.HpeEventService",
            "Id": "EventService",
            "Actions": {
              "#HpeEventService.ImportCACertificate": {
                "target": "/redfish/v1/EventService/Actions/Oem/Hpe/HpeEventService.ImportCACertificate/"
              }
            },
            "CACertificates": {
              "@odata.id": "/redfish/v1/EventService/CACertificates/"
            },
            "RequestedMaxEventsToQueueDefault": 3,
            "RetireOldEventInMinutesDefault": 10,
            "TTLCountDefault": 999999,
            "TTLUnitsDefault": "minutes"
          }
        },
        "ServiceEnabled": true,
        "Status": {
          "Health": "OK",
          "HealthRollup": "OK",
          "State": "Enabled"
        },
        "Subscriptions": {
          "@odata.id": "/redfish/v1/EventService/Subscriptions/"
        }
      }
`
const testPathPRLT_task_service = "/redfish/v1/TaskService/"

const testPayloadPRLT_task_service = `{
        "@odata.context": "/redfish/v1/$metadata#TaskService.TaskService",
        "@odata.etag": "W/\"538CCB4A\"",
        "@odata.id": "/redfish/v1/TaskService/",
        "@odata.type": "#TaskService.v1_1_1.TaskService",
        "Id": "TaskService",
        "CompletedTaskOverWritePolicy": "Manual",
        "DateTime": "2021-01-13T19:25:53Z",
        "Description": "iLO Task Service",
        "LifeCycleEventOnTaskStateChange": true,
        "Name": "Task Service",
        "ServiceEnabled": true,
        "Status": {
          "Health": "OK",
          "HealthRollup": "OK",
          "State": "Enabled"
        },
        "Tasks": {
          "@odata.id": "/redfish/v1/TaskService/Tasks/"
        }
      }
`
const testPathPRLT_update_service = "/redfish/v1/UpdateService/"

const testPayloadPRLT_update_service = `{
        "@odata.context": "/redfish/v1/$metadata#UpdateService.UpdateService",
        "@odata.etag": "W/\"886F108C\"",
        "@odata.id": "/redfish/v1/UpdateService/",
        "@odata.type": "#UpdateService.v1_1_1.UpdateService",
        "Id": "UpdateService",
        "Actions": {
          "#UpdateService.SimpleUpdate": {
            "TransferProtocol@Redfish.AllowableValues": [
              "HTTP",
              "HTTPS"
            ],
            "target": "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate/"
          }
        },
        "Description": "iLO Update Service",
        "FirmwareInventory": {
          "@odata.id": "/redfish/v1/UpdateService/FirmwareInventory/"
        },
        "HttpPushUri": "/cgi-bin/uploadFile",
        "Name": "Update Service",
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeiLOUpdateServiceExt.HpeiLOUpdateServiceExt",
            "@odata.type": "#HpeiLOUpdateServiceExt.v2_1_4.HpeiLOUpdateServiceExt",
            "Actions": {
              "#HpeiLOUpdateServiceExt.AddFromUri": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.AddFromUri/"
              },
              "#HpeiLOUpdateServiceExt.DeleteInstallSets": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.DeleteInstallSets/"
              },
              "#HpeiLOUpdateServiceExt.DeleteMaintenanceWindows": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.DeleteMaintenanceWindows/"
              },
              "#HpeiLOUpdateServiceExt.DeleteUnlockedComponents": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.DeleteUnlockedComponents/"
              },
              "#HpeiLOUpdateServiceExt.DeleteUpdateTaskQueueItems": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.DeleteUpdateTaskQueueItems/"
              },
              "#HpeiLOUpdateServiceExt.RemoveLanguagePack": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.RemoveLanguagePack/"
              },
              "#HpeiLOUpdateServiceExt.SetDefaultLanguage": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.SetDefaultLanguage/"
              },
              "#HpeiLOUpdateServiceExt.StartFirmwareIntegrityCheck": {
                "target": "/redfish/v1/UpdateService/Actions/Oem/Hpe/HpeiLOUpdateServiceExt.StartFirmwareIntegrityCheck/"
              }
            },
            "ComponentRepository": {
              "@odata.id": "/redfish/v1/UpdateService/ComponentRepository/"
            },
            "CurrentTime": "2021-01-13T19:26:06Z",
            "DowngradePolicy": "AllowDowngrade",
            "FirmwareIntegrity": {
              "EnableBackgroundScan": false,
              "LastScanResult": "OK",
              "LastScanTime": "1970-01-01T00:01:57Z",
              "OnIntegrityFailure": "LogOnly",
              "ScanEveryDays": 7
            },
            "InstallSets": {
              "@odata.id": "/redfish/v1/UpdateService/InstallSets/"
            },
            "InvalidImageRepository": {
              "@odata.id": "/redfish/v1/UpdateService/InvalidImageRepository/"
            },
            "MaintenanceWindows": {
              "@odata.id": "/redfish/v1/UpdateService/MaintenanceWindows/"
            },
            "State": "Idle",
            "UpdateTaskQueue": {
              "@odata.id": "/redfish/v1/UpdateService/UpdateTaskQueue/"
            }
          }
        },
        "ServiceEnabled": true,
        "SoftwareInventory": {
          "@odata.id": "/redfish/v1/UpdateService/SoftwareInventory/"
        }
      }
`
const testPathPRLT_managers_1 = "/redfish/v1/Managers/1"

const testPayloadPRLT_managers_1 = `{
   "@odata.context": "/redfish/v1/$metadata#Manager.Manager",
   "@odata.etag": "W/\"49D4F98F\"",
   "@odata.id": "/redfish/v1/Managers/1",
   "@odata.type": "#Manager.v1_5_1.Manager",
   "Id": "1",
   "Actions": {
     "#Manager.Reset": {
       "ResetType@Redfish.AllowableValues": [
         "ForceRestart",
         "GracefulRestart"
       ],
       "target": "/redfish/v1/Managers/1/Actions/Manager.Reset"
     }
   },
   "CommandShell": {
     "ConnectTypesSupported": [
       "SSH",
       "Oem"
     ],
     "MaxConcurrentSessions": 9,
     "ServiceEnabled": true
   },
   "DateTime": "2021-01-12T16:39:59Z",
   "DateTimeLocalOffset": "+00:00",
   "EthernetInterfaces": {
     "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces"
   },
   "FirmwareVersion": "iLO 5 v2.14",
   "GraphicalConsole": {
     "ConnectTypesSupported": [
       "KVMIP"
     ],
     "MaxConcurrentSessions": 10,
     "ServiceEnabled": true
   },
   "HostInterfaces": {
     "@odata.id": "/redfish/v1/Managers/1/HostInterfaces"
   },
   "Links": {
     "ManagerInChassis": {
       "@odata.id": "/redfish/v1/Chassis/1"
     },
     "ManagerForServers": [
       {
         "@odata.id": "/redfish/v1/Systems/1"
       }
     ],
     "ManagerForChassis": [
       {
         "@odata.id": "/redfish/v1/Chassis/1"
       }
     ]
   },
   "LogServices": {
     "@odata.id": "/redfish/v1/Managers/1/LogServices"
   },
   "ManagerType": "BMC",
   "Model": "iLO 5",
   "Name": "Manager",
   "NetworkProtocol": {
     "@odata.id": "/redfish/v1/Managers/1/NetworkProtocol"
   },
   "Oem": {
     "Hpe": {
       "@odata.context": "/redfish/v1/$metadata#HpeiLO.HpeiLO",
       "@odata.type": "#HpeiLO.v2_7_1.HpeiLO",
       "Actions": {
         "#HpeiLO.ClearHotKeys": {
           "target": "/redfish/v1/Managers/1/Actions/Oem/Hpe/HpeiLO.ClearHotKeys"
         },
         "#HpeiLO.ClearRestApiState": {
           "target": "/redfish/v1/Managers/1/Actions/Oem/Hpe/HpeiLO.ClearRestApiState"
         },
         "#HpeiLO.DisableiLOFunctionality": {
           "target": "/redfish/v1/Managers/1/Actions/Oem/Hpe/HpeiLO.DisableiLOFunctionality"
         },
         "#HpeiLO.RequestFirmwareAndOsRecovery": {
           "target": "/redfish/v1/Managers/1/Actions/Oem/Hpe/HpeiLO.RequestFirmwareAndOsRecovery"
         },
         "#HpeiLO.ResetToFactoryDefaults": {
           "ResetType@Redfish.AllowableValues": [
             "Default"
           ],
           "target": "/redfish/v1/Managers/1/Actions/Oem/Hpe/HpeiLO.ResetToFactoryDefaults"
         }
       },
       "ClearRestApiStatus": "DataPresent",
       "ConfigurationSettings": "Current",
       "FederationConfig": {
         "IPv6MulticastScope": "Site",
         "MulticastAnnouncementInterval": 600,
         "MulticastDiscovery": "Enabled",
         "MulticastTimeToLive": 5,
         "iLOFederationManagement": "Enabled"
       },
       "Firmware": {
         "Current": {
           "Date": "Feb 11 2020",
           "DebugBuild": false,
           "MajorVersion": 2,
           "MinorVersion": 14,
           "VersionString": "iLO 5 v2.14"
         }
       },
       "FrontPanelUSB": {
         "State": "Ready"
       },
       "IdleConnectionTimeoutMinutes": 30,
       "IntegratedRemoteConsole": {
         "HotKeys": [
           {
             "KeySequence": [
               "NONE",
               "NONE",
               "NONE",
               "NONE",
               "NONE"
             ],
             "Name": "Ctrl-T"
           },
           {
             "KeySequence": [
               "NONE",
               "NONE",
               "NONE",
               "NONE",
               "NONE"
             ],
             "Name": "Ctrl-U"
           },
           {
             "KeySequence": [
               "NONE",
               "NONE",
               "NONE",
               "NONE",
               "NONE"
             ],
             "Name": "Ctrl-V"
           },
           {
             "KeySequence": [
               "NONE",
               "NONE",
               "NONE",
               "NONE",
               "NONE"
             ],
             "Name": "Ctrl-W"
           },
           {
             "KeySequence": [
               "NONE",
               "NONE",
               "NONE",
               "NONE",
               "NONE"
             ],
             "Name": "Ctrl-X"
           },
           {
             "KeySequence": [
               "NONE",
               "NONE",
               "NONE",
               "NONE",
               "NONE"
             ],
             "Name": "Ctrl-Y"
           }
         ],
         "LockKey": {
           "CustomKeySequence": [
             "NONE",
             "NONE",
             "NONE",
             "NONE",
             "NONE"
           ],
           "LockOption": "Disabled"
         },
         "TrustedCertificateRequired": false
       },
       "License": {
         "LicenseKey": "XXXXX-XXXXX-XXXXX-XXXXX-QQ4PB",
         "LicenseString": "iLO Advanced",
         "LicenseType": "Perpetual"
       },
       "Links": {
         "DateTimeService": {
           "@odata.id": "/redfish/v1/Managers/1/DateTime"
         },
         "LicenseService": {
           "@odata.id": "/redfish/v1/Managers/1/LicenseService"
         },
         "EmbeddedMediaService": {
           "@odata.id": "/redfish/v1/Managers/1/EmbeddedMedia"
         },
         "FederationPeers": {
           "@odata.id": "/redfish/v1/Managers/1/FederationPeers"
         },
         "FederationGroups": {
           "@odata.id": "/redfish/v1/Managers/1/FederationGroups"
         },
         "ActiveHealthSystem": {
           "@odata.id": "/redfish/v1/Managers/1/ActiveHealthSystem"
         },
         "SecurityService": {
           "@odata.id": "/redfish/v1/Managers/1/SecurityService"
         },
         "BackupRestoreService": {
           "@odata.id": "/redfish/v1/Managers/1/BackupRestoreService"
         },
         "RemoteSupport": {
           "@odata.id": "/redfish/v1/Managers/1/RemoteSupportService"
         },
         "GUIService": {
           "@odata.id": "/redfish/v1/Managers/1/GUIService"
         },
         "FederationDispatch": {
           "extref": "/dispatch"
         },
         "VSPLogLocation": {
           "extref": "/sol.log.gz"
         },
         "Thumbnail": {
           "extref": "/images/thumbnail.bmp"
         }
       },
       "PersistentMouseKeyboardEnabled": false,
       "PhysicalMonitorHealthStatusEnabled": true,
       "RIBCLEnabled": true,
       "RemoteConsoleThumbnailEnabled": true,
       "RequireHostAuthentication": false,
       "RequiredLoginForiLORBSU": false,
       "SerialCLISpeed": 115200,
       "SerialCLIStatus": "EnabledAuthReq",
       "SerialCLIUART": "Present",
       "VSPDlLoggingEnabled": false,
       "VSPLogDownloadEnabled": false,
       "VideoPresenceDetectOverride": true,
       "VideoPresenceDetectOverrideSupported": true,
       "VirtualNICEnabled": true,
       "WebGuiEnabled": true,
       "iLOFunctionalityEnabled": true,
       "iLOFunctionalityRequired": false,
       "iLOIPduringPOSTEnabled": true,
       "iLORBSUEnabled": true,
       "iLOSelfTestResults": [
         {
           "Notes": "",
           "SelfTestName": "NVRAMData",
           "Status": "OK"
         },
         {
           "Notes": "Controller firmware revision  2.11.00  ",
           "SelfTestName": "EmbeddedFlash",
           "Status": "OK"
         },
         {
           "Notes": "",
           "SelfTestName": "EEPROM",
           "Status": "OK"
         },
         {
           "Notes": "",
           "SelfTestName": "HostRom",
           "Status": "OK"
         },
         {
           "Notes": "",
           "SelfTestName": "SupportedHost",
           "Status": "OK"
         },
         {
           "Notes": "Version 1.0.7",
           "SelfTestName": "PowerManagementController",
           "Status": "Informational"
         },
         {
           "Notes": "ProLiant DL325 Gen10 Plus System Programmable Logic Device 0x0D",
           "SelfTestName": "CPLDPAL0",
           "Status": "Informational"
         },
         {
           "Notes": "",
           "SelfTestName": "ASICFuses",
           "Status": "OK"
         }
       ],
       "iLOServicePort": {
         "MassStorageAuthenticationRequired": false,
         "USBEthernetAdaptersEnabled": true,
         "USBFlashDriveEnabled": true,
         "iLOServicePortEnabled": true
       }
     }
   },
   "SerialConsole": {
     "ConnectTypesSupported": [
       "SSH",
       "IPMI",
       "Oem"
     ],
     "MaxConcurrentSessions": 13,
     "ServiceEnabled": true
   },
   "Status": {
     "Health": "OK",
     "State": "Enabled"
   },
   "UUID": "5e623487-8864-51e7-af0a-0d0c365a1fd0",
   "VirtualMedia": {
     "@odata.id": "/redfish/v1/Managers/1/VirtualMedia"
   }
 }
`
const testPathPRLT_managers_1_ethernet_interfaces = "/redfish/v1/Managers/1/EthernetInterfaces"

const testPayloadPRLT_managers_1_ethernet_interfaces = `{
   "@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
   "@odata.etag": "W/\"E589C4BF\"",
   "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces",
   "@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
   "Description": "Configuration of Manager Network Interfaces",
   "Name": "Manager Network Interfaces",
   "Members": [
     {
       "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces/1"
     },
     {
       "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces/2"
     },
     {
       "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces/3"
     }
   ],
   "Members@odata.count": 3
 }
`
const testPathPRLT_managers_1_ethernet_interfaces_1 = "/redfish/v1/Managers/1/EthernetInterfaces/1"

const testPayloadPRLT_managers_1_ethernet_interfaces_1 = `{
   "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
   "@odata.etag": "W/\"0E8D4794\"",
   "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces/1",
   "@odata.type": "#EthernetInterface.v1_4_1.EthernetInterface",
   "Id": "1",
   "AutoNeg": true,
   "DHCPv4": {
     "DHCPEnabled": true,
     "UseDNSServers": true,
     "UseDomainName": true,
     "UseGateway": true,
     "UseNTPServers": true,
     "UseStaticRoutes": true
   },
   "DHCPv6": {
     "OperatingMode": "Stateful",
     "UseDNSServers": true,
     "UseDomainName": true,
     "UseNTPServers": true,
     "UseRapidCommit": false
   },
   "Description": "Configuration of this Manager Network Interface",
   "FQDN": "ILOMXQ019020H.",
   "FullDuplex": true,
   "HostName": "ILOMXQ019020H",
   "IPv4Addresses": [
     {
       "Address": "10.254.2.11",
       "AddressOrigin": "DHCP",
       "Gateway": "10.254.0.1",
       "SubnetMask": "255.255.128.0"
     }
   ],
   "IPv4StaticAddresses": [],
   "IPv6AddressPolicyTable": [
     {
       "Label": null,
       "Precedence": 35,
       "Prefix": "::ffff:0:0/96"
     }
   ],
   "IPv6Addresses": [
     {
       "Address": "FE80::9640:C9FF:FE37:E38A",
       "AddressOrigin": "SLAAC",
       "AddressState": "Preferred",
       "PrefixLength": 64
     }
   ],
   "IPv6DefaultGateway": "::",
   "IPv6StaticAddresses": [
     {
       "Address": "::",
       "PrefixLength": null
     },
     {
       "Address": "::",
       "PrefixLength": null
     },
     {
       "Address": "::",
       "PrefixLength": null
     },
     {
       "Address": "::",
       "PrefixLength": null
     }
   ],
   "IPv6StaticDefaultGateways": [
     {
       "Address": "::"
     }
   ],
   "InterfaceEnabled": true,
   "LinkStatus": "LinkUp",
   "MACAddress": "94:40:C9:37:E3:8A",
   "MaxIPv6StaticAddresses": 4,
   "Name": "Manager Dedicated Network Interface",
   "NameServers": [
     "10.94.100.225"
   ],
   "Oem": {
     "Hpe": {
       "@odata.context": "/redfish/v1/$metadata#HpeiLOEthernetNetworkInterface.HpeiLOEthernetNetworkInterface",
       "@odata.type": "#HpeiLOEthernetNetworkInterface.v2_2_1.HpeiLOEthernetNetworkInterface",
       "ConfigurationSettings": "Current",
       "DHCPv4": {
         "ClientIdType": "Default",
         "Enabled": true,
         "UseDNSServers": true,
         "UseDomainName": true,
         "UseGateway": true,
         "UseNTPServers": true,
         "UseStaticRoutes": true,
         "UseWINSServers": true
       },
       "DHCPv6": {
         "StatefulModeEnabled": true,
         "StatelessModeEnabled": true,
         "UseDNSServers": true,
         "UseDomainName": true,
         "UseNTPServers": true,
         "UseRapidCommit": false
       },
       "DomainName": "",
       "HostName": "ILOMXQ019020H",
       "IPv4": {
         "DDNSRegistration": true,
         "DNSServers": [
           "10.94.100.225",
           "0.0.0.0",
           "0.0.0.0"
         ],
         "StaticRoutes": [
           {
             "Destination": "0.0.0.0",
             "Gateway": "0.0.0.0",
             "SubnetMask": "0.0.0.0"
           },
           {
             "Destination": "0.0.0.0",
             "Gateway": "0.0.0.0",
             "SubnetMask": "0.0.0.0"
           },
           {
             "Destination": "0.0.0.0",
             "Gateway": "0.0.0.0",
             "SubnetMask": "0.0.0.0"
           }
         ],
         "WINSRegistration": true,
         "WINSServers": [
           "0.0.0.0",
           "0.0.0.0"
         ]
       },
       "IPv6": {
         "DDNSRegistration": true,
         "DNSServers": [
           "::",
           "::",
           "::"
         ],
         "SLAACEnabled": true,
         "StaticDefaultGateway": "::",
         "StaticRoutes": [
           {
             "Destination": "::",
             "Gateway": "::",
             "PrefixLength": null,
             "Status": "Unknown"
           },
           {
             "Destination": "::",
             "Gateway": "::",
             "PrefixLength": null,
             "Status": "Unknown"
           },
           {
             "Destination": "::",
             "Gateway": "::",
             "PrefixLength": null,
             "Status": "Unknown"
           }
         ]
       },
       "InterfaceType": "Dedicated",
       "NICEnabled": true,
       "NICSupportsIPv6": true,
       "PingGatewayOnStartup": true
     }
   },
   "PermanentMACAddress": "94:40:C9:37:E3:8A",
   "SpeedMbps": 1000,
   "StatelessAddressAutoConfig": {
     "IPv6AutoConfigEnabled": true
   },
   "StaticNameServers": [
     "0.0.0.0",
     "0.0.0.0",
     "0.0.0.0",
     "::",
     "::",
     "::"
   ],
   "Status": {
     "Health": "OK",
     "State": "Enabled"
   },
   "VLAN": {
     "VLANEnable": false,
     "VLANId": null
   }
 }
`
const testPathPRLT_managers_1_ethernet_interfaces_2 = "/redfish/v1/Managers/1/EthernetInterfaces/2"

const testPayloadPRLT_managers_1_ethernet_interfaces_2 = `{
   "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
   "@odata.etag": "W/\"BBD0FBD8\"",
   "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces/2",
   "@odata.type": "#EthernetInterface.v1_4_1.EthernetInterface",
   "Id": "2",
   "AutoNeg": null,
   "DHCPv4": {
     "DHCPEnabled": true,
     "UseDNSServers": true,
     "UseDomainName": true,
     "UseGateway": true,
     "UseNTPServers": true,
     "UseStaticRoutes": true
   },
   "DHCPv6": {
     "OperatingMode": "Stateful",
     "UseDNSServers": true,
     "UseDomainName": true,
     "UseNTPServers": true,
     "UseRapidCommit": false
   },
   "Description": "Configuration of this Manager Network Interface",
   "FQDN": "ILOMXQ019020H.",
   "FullDuplex": false,
   "HostName": "ILOMXQ019020H",
   "IPv4Addresses": [
     {
       "Address": "0.0.0.0",
       "AddressOrigin": "DHCP",
       "Gateway": "0.0.0.0",
       "SubnetMask": "255.255.255.255"
     }
   ],
   "IPv4StaticAddresses": [],
   "IPv6AddressPolicyTable": [
     {
       "Label": null,
       "Precedence": 35,
       "Prefix": "::ffff:0:0/96"
     }
   ],
   "IPv6Addresses": [],
   "IPv6DefaultGateway": "::",
   "IPv6StaticAddresses": [
     {
       "Address": "::",
       "PrefixLength": null
     },
     {
       "Address": "::",
       "PrefixLength": null
     },
     {
       "Address": "::",
       "PrefixLength": null
     },
     {
       "Address": "::",
       "PrefixLength": null
     }
   ],
   "IPv6StaticDefaultGateways": [
     {
       "Address": "::"
     }
   ],
   "InterfaceEnabled": false,
   "LinkStatus": null,
   "MACAddress": "94:40:C9:37:E3:8B",
   "MaxIPv6StaticAddresses": 4,
   "Name": "Manager Shared Network Interface",
   "NameServers": [],
   "Oem": {
     "Hpe": {
       "@odata.context": "/redfish/v1/$metadata#HpeiLOEthernetNetworkInterface.HpeiLOEthernetNetworkInterface",
       "@odata.type": "#HpeiLOEthernetNetworkInterface.v2_2_1.HpeiLOEthernetNetworkInterface",
       "ConfigurationSettings": "Current",
       "DHCPv4": {
         "ClientIdType": "Default",
         "Enabled": true,
         "UseDNSServers": true,
         "UseDomainName": true,
         "UseGateway": true,
         "UseNTPServers": true,
         "UseStaticRoutes": true,
         "UseWINSServers": true
       },
       "DHCPv6": {
         "StatefulModeEnabled": true,
         "StatelessModeEnabled": true,
         "UseDNSServers": true,
         "UseDomainName": true,
         "UseNTPServers": true,
         "UseRapidCommit": false
       },
       "DomainName": "",
       "HostName": "ILOMXQ019020H",
       "IPv4": {
         "DDNSRegistration": true,
         "DNSServers": [
           "0.0.0.0",
           "0.0.0.0",
           "0.0.0.0"
         ],
         "StaticRoutes": [
           {
             "Destination": "0.0.0.0",
             "Gateway": "0.0.0.0",
             "SubnetMask": "0.0.0.0"
           },
           {
             "Destination": "0.0.0.0",
             "Gateway": "0.0.0.0",
             "SubnetMask": "0.0.0.0"
           },
           {
             "Destination": "0.0.0.0",
             "Gateway": "0.0.0.0",
             "SubnetMask": "0.0.0.0"
           }
         ],
         "WINSRegistration": true,
         "WINSServers": [
           "0.0.0.0",
           "0.0.0.0"
         ]
       },
       "IPv6": {
         "DDNSRegistration": true,
         "DNSServers": [
           "::",
           "::",
           "::"
         ],
         "SLAACEnabled": true,
         "StaticDefaultGateway": "::",
         "StaticRoutes": [
           {
             "Destination": "::",
             "Gateway": "::",
             "PrefixLength": null,
             "Status": "Unknown"
           },
           {
             "Destination": "::",
             "Gateway": "::",
             "PrefixLength": null,
             "Status": "Unknown"
           },
           {
             "Destination": "::",
             "Gateway": "::",
             "PrefixLength": null,
             "Status": "Unknown"
           }
         ]
       },
       "InterfaceType": "Shared",
       "NICEnabled": false,
       "NICSupportsIPv6": true,
       "PingGatewayOnStartup": true,
       "SharedNetworkPortOptions": {
         "Port": 1
       },
       "SupportsFlexibleLOM": true,
       "SupportsLOM": true
     }
   },
   "PermanentMACAddress": "94:40:C9:37:E3:8B",
   "SpeedMbps": null,
   "StatelessAddressAutoConfig": {
     "IPv6AutoConfigEnabled": true
   },
   "StaticNameServers": [
     "0.0.0.0",
     "0.0.0.0",
     "0.0.0.0",
     "::",
     "::",
     "::"
   ],
   "Status": {
     "Health": null,
     "State": "Disabled"
   },
   "VLAN": {
     "VLANEnable": false,
     "VLANId": null
   }
 }
`
const testPathPRLT_managers_1_ethernet_interfaces_3 = "/redfish/v1/Managers/1/EthernetInterfaces/3"

const testPayloadPRLT_managers_1_ethernet_interfaces_3 = `{
   "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
   "@odata.etag": "W/\"BCB18AFB\"",
   "@odata.id": "/redfish/v1/Managers/1/EthernetInterfaces/3",
   "@odata.type": "#EthernetInterface.v1_4_1.EthernetInterface",
   "Id": "3",
   "Description": "Configuration of this Manager USB Ethernet Interface available for access from Host.",
   "IPv4Addresses": [
     {
       "Address": "16.1.15.1",
       "AddressOrigin": "DHCP",
       "SubnetMask": "255.255.255.252"
     }
   ],
   "IPv4StaticAddresses": [],
   "IPv6AddressPolicyTable": [],
   "IPv6Addresses": [],
   "IPv6StaticAddresses": [],
   "IPv6StaticDefaultGateways": [],
   "InterfaceEnabled": true,
   "LinkStatus": "LinkUp",
   "Links": {
     "HostInterface": {
       "@odata.id": "/redfish/v1/Managers/1/HostInterfaces/1"
     }
   },
   "MACAddress": "00:CA:FE:F0:0D:04",
   "Name": "Manager Virtual Network Interface",
   "NameServers": [],
   "Oem": {
     "Hpe": {
       "@odata.context": "/redfish/v1/$metadata#HpeiLOEthernetNetworkInterface.HpeiLOEthernetNetworkInterface",
       "@odata.type": "#HpeiLOEthernetNetworkInterface.v2_2_1.HpeiLOEthernetNetworkInterface",
       "ConfigurationSettings": "Current",
       "InterfaceType": "HostInterface",
       "NICSupportsIPv6": false
     }
   },
   "PermanentMACAddress": "00:CA:FE:F0:0D:04",
   "StaticNameServers": [],
   "Status": {
     "Health": "OK",
     "State": "Enabled"
   }
 }
`
const testPathPRLT_chassis = "/redfish/v1/Chassis/"

const testPayloadPRLT_chassis = `{
   "@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
   "@odata.etag": "W/\"AA6D42B0\"",
   "@odata.id": "/redfish/v1/Chassis",
   "@odata.type": "#ChassisCollection.ChassisCollection",
   "Description": "Computer System Chassis View",
   "Name": "Computer System Chassis",
   "Members": [
     {
       "@odata.id": "/redfish/v1/Chassis/1"
     }
   ],
   "Members@odata.count": 1
 }
`
const testPathPRLT_chassis_1 = "/redfish/v1/Chassis/1"

const testPayloadPRLT_chassis_1 = `{
   "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
   "@odata.etag": "W/\"DD454000\"",
   "@odata.id": "/redfish/v1/Chassis/1",
   "@odata.type": "#Chassis.v1_6_0.Chassis",
   "Id": "1",
   "AssetTag": "",
   "ChassisType": "RackMount",
   "IndicatorLED": "Off",
   "Links": {
     "ManagedBy": [
       {
         "@odata.id": "/redfish/v1/Managers/1"
       }
     ],
     "ComputerSystems": [
       {
         "@odata.id": "/redfish/v1/Systems/1"
       }
     ]
   },
   "Manufacturer": "HPE",
   "Model": "ProLiant DL325 Gen10 Plus",
   "Name": "Computer System Chassis",
   "NetworkAdapters": {
     "@odata.id": "/redfish/v1/Chassis/1/NetworkAdapters"
   },
   "Oem": {
     "Hpe": {
       "@odata.context": "/redfish/v1/$metadata#HpeServerChassis.HpeServerChassis",
       "@odata.type": "#HpeServerChassis.v2_3_1.HpeServerChassis",
       "Actions": {
         "#HpeServerChassis.DisableMCTPOnServer": {
           "target": "/redfish/v1/Chassis/1/Actions/Oem/Hpe/HpeServerChassis.DisableMCTPOnServer"
         },
         "#HpeServerChassis.FactoryResetMCTP": {
           "target": "/redfish/v1/Chassis/1/Actions/Oem/Hpe/HpeServerChassis.FactoryResetMCTP"
         }
       },
       "Firmware": {
         "PlatformDefinitionTable": {
           "Current": {
             "VersionString": "2.5.0 Build 21"
           }
         },
         "PowerManagementController": {
           "Current": {
             "VersionString": "1.0.7"
           }
         },
         "PowerManagementControllerBootloader": {
           "Current": {
             "Family": "25",
             "VersionString": "1.1"
           }
         },
         "SystemProgrammableLogicDevice": {
           "Current": {
             "VersionString": "0x0D"
           }
         }
       },
       "Links": {
         "Devices": {
           "@odata.id": "/redfish/v1/Chassis/1/Devices"
         }
       },
       "MCTPEnabledOnServer": true,
       "SystemMaintenanceSwitches": {
         "Sw1": "Off",
         "Sw10": "Off",
         "Sw11": "Off",
         "Sw12": "Off",
         "Sw2": "Off",
         "Sw3": "Off",
         "Sw4": "Off",
         "Sw5": "Off",
         "Sw6": "Off",
         "Sw7": "Off",
         "Sw8": "Off",
         "Sw9": "Off"
       }
     }
   },
   "Power": {
     "@odata.id": "/redfish/v1/Chassis/1/Power"
   },
   "PowerState": "On",
   "SKU": "P18606-B21",
   "SerialNumber": "MXQ019020H",
   "Status": {
     "Health": "OK",
     "State": "Enabled"
   },
   "Thermal": {
     "@odata.id": "/redfish/v1/Chassis/1/Thermal"
   }
 }
`
const testPathPRLT_chassis_1_power = "/redfish/v1/Chassis/1/Power"

const testPayloadPRLT_chassis_1_power = `{
   "@odata.context": "/redfish/v1/$metadata#Power.Power",
   "@odata.etag": "W/\"88AA866B\"",
   "@odata.id": "/redfish/v1/Chassis/1/Power",
   "@odata.type": "#Power.v1_3_0.Power",
   "Id": "Power",
   "Name": "PowerMetrics",
   "Oem": {
     "Hpe": {
       "@odata.context": "/redfish/v1/$metadata#HpePowerMetricsExt.HpePowerMetricsExt",
       "@odata.type": "#HpePowerMetricsExt.v2_2_0.HpePowerMetricsExt",
       "BrownoutRecoveryEnabled": true,
       "HasCpuPowerMetering": true,
       "HasDimmPowerMetering": true,
       "HasGpuPowerMetering": false,
       "HasPowerMetering": true,
       "HighEfficiencyMode": "Balanced",
       "Links": {
         "PowerMeter": {
           "@odata.id": "/redfish/v1/Chassis/1/Power/PowerMeter"
         },
         "FastPowerMeter": {
           "@odata.id": "/redfish/v1/Chassis/1/Power/FastPowerMeter"
         },
         "FederatedGroupCapping": {
           "@odata.id": "/redfish/v1/Chassis/1/Power/FederatedGroupCapping"
         }
       },
       "MinimumSafelyAchievableCap": null,
       "MinimumSafelyAchievableCapValid": false,
       "SNMPPowerThresholdAlert": {
         "DurationInMin": 0,
         "ThresholdWatts": 0,
         "Trigger": "Disabled"
       }
     }
   },
   "PowerControl": [
     {
       "@odata.id": "/redfish/v1/Chassis/1/Power#PowerControl/0",
       "MemberId": "0",
       "PowerCapacityWatts": 1000,
       "PowerConsumedWatts": 121,
       "PowerLimit": {
         "LimitInWatts": null
       },
       "PowerMetrics": {
         "AverageConsumedWatts": 121,
         "IntervalInMin": 20,
         "MaxConsumedWatts": 137,
         "MinConsumedWatts": 121
       }
     }
   ],
   "PowerSupplies": [
     {
       "@odata.id": "/redfish/v1/Chassis/1/Power#PowerSupplies/0",
       "FirmwareVersion": "1.00",
       "LastPowerOutputWatts": 55,
       "LineInputVoltage": 206,
       "LineInputVoltageType": "ACHighLine",
       "Manufacturer": "LTEON",
       "MemberId": "0",
       "Model": "865408-B21",
       "Name": "HpeServerPowerSupply",
       "Oem": {
         "Hpe": {
           "@odata.context": "/redfish/v1/$metadata#HpeServerPowerSupply.HpeServerPowerSupply",
           "@odata.type": "#HpeServerPowerSupply.v2_0_0.HpeServerPowerSupply",
           "AveragePowerOutputWatts": 55,
           "BayNumber": 1,
           "HotplugCapable": true,
           "MaxPowerOutputWatts": 56,
           "Mismatched": false,
           "PowerSupplyStatus": {
             "State": "Ok"
           },
           "iPDUCapable": false
         }
       },
       "PowerCapacityWatts": 500,
       "PowerSupplyType": "AC",
       "SerialNumber": "5WBXK0FLLDG2YG",
       "SparePartNumber": "866729-001",
       "Status": {
         "Health": "OK",
         "State": "Enabled"
       }
     },
     {
       "@odata.id": "/redfish/v1/Chassis/1/Power#PowerSupplies/1",
       "FirmwareVersion": "1.00",
       "LastPowerOutputWatts": 66,
       "LineInputVoltage": 206,
       "LineInputVoltageType": "ACHighLine",
       "Manufacturer": "LTEON",
       "MemberId": "1",
       "Model": "865408-B21",
       "Name": "HpeServerPowerSupply",
       "Oem": {
         "Hpe": {
           "@odata.context": "/redfish/v1/$metadata#HpeServerPowerSupply.HpeServerPowerSupply",
           "@odata.type": "#HpeServerPowerSupply.v2_0_0.HpeServerPowerSupply",
           "AveragePowerOutputWatts": 66,
           "BayNumber": 2,
           "HotplugCapable": true,
           "MaxPowerOutputWatts": 68,
           "Mismatched": false,
           "PowerSupplyStatus": {
             "State": "Ok"
           },
           "iPDUCapable": false
         }
       },
       "PowerCapacityWatts": 500,
       "PowerSupplyType": "AC",
       "SerialNumber": "5WBXK0FLLDG3N9",
       "SparePartNumber": "866729-001",
       "Status": {
         "Health": "OK",
         "State": "Enabled"
       }
     }
   ],
   "Redundancy": [
     {
       "@odata.id": "/redfish/v1/Chassis/1/Power#Redundancy/0",
       "MaxNumSupported": 2,
       "MemberId": "0",
       "MinNumNeeded": 2,
       "Mode": "Failover",
       "Name": "PowerSupply Redundancy Group 1",
       "RedundancySet": [
         {
           "@odata.id": "/redfish/v1/Chassis/1/Power#PowerSupplies/0"
         },
         {
           "@odata.id": "/redfish/v1/Chassis/1/Power#PowerSupplies/1"
         }
       ],
       "Status": {
         "Health": "OK",
         "State": "Enabled"
       }
     }
   ]
 }
`
const testPathPRLT_chassis_1_network_adapters = "/redfish/v1/Chassis/1/NetworkAdapters"

const testPayloadPRLT_chassis_1_network_adapters = `{
   "@odata.context": "/redfish/v1/$metadata#NetworkAdapterCollection.NetworkAdapterCollection",
   "@odata.etag": "W/\"F303ECE9\"",
   "@odata.id": "/redfish/v1/Chassis/1/NetworkAdapters",
   "@odata.type": "#NetworkAdapterCollection.NetworkAdapterCollection",
   "Description": "The collection of network adapter resource instances available in this chassis.",
   "Name": "NetworkAdapterCollection",
   "Oem": {
     "Hpe": {
       "@odata.context": "/redfish/v1/$metadata#HpeNetworkAdapterStatus.HpeNetworkAdapterStatus",
       "@odata.type": "#HpeNetworkAdapterStatus.v1_0_0.HpeNetworkAdapterStatus",
       "MemberContents": "AllDevices"
     }
   },
   "Members": [],
   "Members@odata.count": 0
 }
`

const testPathPRLT_chassis_1_devices = "/redfish/v1/Chassis/1/Devices"

const testPayloadPRLT_chassis_1_devices = `{
  "@odata.context": "/redfish/v1/$metadata#HpeServerDeviceCollection.HpeServerDeviceCollection",
  "@odata.etag": "W/\"DA402EBA\"",
  "@odata.id": "/redfish/v1/Chassis/1/Devices",
  "@odata.type": "#HpeServerDeviceCollection.HpeServerDeviceCollection",
  "Description": "Devices View",
  "Name": "Devices",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Chassis/1/Devices/1"
    },
    {
      "@odata.id": "/redfish/v1/Chassis/1/Devices/2"
    },
    {
      "@odata.id": "/redfish/v1/Chassis/1/Devices/3"
    }
  ],
  "Members@odata.count": 3
}
`

const testPathPRLT_chassis_1_devices_1 = "/redfish/v1/Chassis/1/Devices/1"

const testPayloadPRLT_chassis_1_devices_1 = `{
  "@odata.context": "/redfish/v1/$metadata#HpeServerDevice.HpeServerDevice",
  "@odata.etag": "W/\"923970AD\"",
  "@odata.id": "/redfish/v1/Chassis/1/Devices/1",
  "@odata.type": "#HpeServerDevice.v2_0_0.HpeServerDevice",
  "Id": "1",
  "DeviceInstances": [
    {
      "@odata.id": "/redfish/v1/Systems/1/PCIDevices/5"
    },
    {
      "@odata.id": "/redfish/v1/Systems/1/PCIDevices/6"
    }
  ],
  "DeviceType": "LOM/NIC",
  "FirmwareVersion": {
    "Current": {
      "VersionString": "08.50.78"
    }
  },
  "Links": {
    "PCISlot": {
      "@odata.id": "/redfish/v1/Systems/1/PCISlots/5"
    }
  },
  "Location": "OCP 3.0 Slot 10",
  "MCTPProtocolDisabled": false,
  "Manufacturer": "Marvell",
  "Name": "Marvell 2P 25GbE SFP28 QL41232HQCU-HC OCP3 Adapter",
  "PartNumber": "AH2010414-31  A",
  "ProductPartNumber": "AH2010414-31  A",
  "ProductVersion": " ",
  "SerialNumber": "REE2020K07000",
  "Status": {
    "Health": "OK",
    "State": "Enabled"
  }
}
`

const testPathPRLT_chassis_1_devices_2 = "/redfish/v1/Chassis/1/Devices/2"

const testPayloadPRLT_chassis_1_devices_2 = `{
  "@odata.context": "/redfish/v1/$metadata#HpeServerDevice.HpeServerDevice",
  "@odata.etag": "W/\"D5CDA23A\"",
  "@odata.id": "/redfish/v1/Chassis/1/Devices/2",
  "@odata.type": "#HpeServerDevice.v2_0_0.HpeServerDevice",
  "Id": "2",
  "DeviceInstances": [
    {
      "@odata.id": "/redfish/v1/Systems/1/PCIDevices/7"
    }
  ],
  "DeviceType": "Smart Storage",
  "FirmwareVersion": {
    "Current": {
      "VersionString": "2.65"
    }
  },
  "Links": {
    "PCISlot": {
      "@odata.id": "/redfish/v1/Systems/1/PCISlots/6"
    }
  },
  "Location": "Storage Slot 12",
  "MCTPProtocolDisabled": false,
  "Manufacturer": "",
  "Name": "HPE Smart Array E208i-a SR Gen10",
  "PartNumber": "836259-001",
  "ProductPartNumber": "804329-001",
  "ProductVersion": "B",
  "SerialNumber": "PEYHB0ERHDV3OT ",
  "Status": {
    "Health": "OK",
    "State": "Enabled"
  }
}
`

const testPathPRLT_chassis_1_devices_3 = "/redfish/v1/Chassis/1/Devices/3"

const testPayloadPRLT_chassis_1_devices_3 = `{
  "@odata.context": "/redfish/v1/$metadata#HpeServerDevice.HpeServerDevice",
  "@odata.etag": "W/\"DF7754E3\"",
  "@odata.id": "/redfish/v1/Chassis/1/Devices/3",
  "@odata.type": "#HpeServerDevice.v2_0_0.HpeServerDevice",
  "Id": "3",
  "DeviceInstances": [
    {
      "@odata.id": "/redfish/v1/Systems/1/PCIDevices/2"
    }
  ],
  "DeviceType": "GPU",
  "FirmwareVersion": {
    "Current": {
      "VersionString": "88.00.48.00.02"
    }
  },
  "Links": {
    "PCISlot": {
      "@odata.id": "/redfish/v1/Systems/1/PCISlots/1"
    }
  },
  "Location": "PCI-E Slot 2",
  "MCTPProtocolDisabled": false,
  "Manufacturer": "NVIDIA",
  "Name": "NVIDIA Tesla V100 32GB PCIE",
  "PartNumber": "1DB6-897-A1",
  "ProductPartNumber": "900-2G500-0010-000",
  "ProductVersion": "A1",
  "SerialNumber": "0322818009821",
  "Status": {
    "Health": "OK",
    "State": "Enabled"
  }
}
`

const testPathPRLT_systems = "/redfish/v1/Systems/"

const testPayloadPRLT_systems = `{
        "@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
        "@odata.etag": "W/\"AA6D42B0\"",
        "@odata.id": "/redfish/v1/Systems",
        "@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
        "Description": "Computer Systems view",
        "Name": "Computer Systems",
        "Members": [
          {
            "@odata.id": "/redfish/v1/Systems/1"
          }
        ],
        "Members@odata.count": 1
      }
`
const testPathPRLT_systems_1 = "/redfish/v1/Systems/1"

const testPayloadPRLT_systems_1 = `{
        "@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
        "@odata.etag": "W/\"F7ACBF79\"",
        "@odata.id": "/redfish/v1/Systems/1",
        "@odata.type": "#ComputerSystem.v1_4_0.ComputerSystem",
        "Id": "1",
        "Actions": {
          "#ComputerSystem.Reset": {
            "ResetType@Redfish.AllowableValues": [
              "On",
              "ForceOff",
              "GracefulShutdown",
              "ForceRestart",
              "Nmi",
              "PushPowerButton"
            ],
            "target": "/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"
          }
        },
        "AssetTag": "",
        "Bios": {
          "@odata.id": "/redfish/v1/systems/1/bios"
        },
        "BiosVersion": "A43 v1.20 (02/28/2020)",
        "Boot": {
          "BootOptions": {
            "@odata.id": "/redfish/v1/Systems/1/BootOptions"
          },
          "BootOrder": [
            "Boot000D",
            "Boot000E",
            "Boot000F",
            "Boot0010",
            "Boot0011",
            "Boot0012",
            "Boot0013",
            "Boot0014",
            "Boot0015",
            "Boot001E",
            "Boot001F",
            "Boot0016",
            "Boot0017",
            "Boot0018",
            "Boot0019"
          ],
          "BootSourceOverrideEnabled": "Disabled",
          "BootSourceOverrideMode": "UEFI",
          "BootSourceOverrideTarget": "None",
          "BootSourceOverrideTarget@Redfish.AllowableValues": [
            "None",
            "Cd",
            "Hdd",
            "Usb",
            "SDCard",
            "Utilities",
            "Diags",
            "BiosSetup",
            "Pxe",
            "UefiShell",
            "UefiHttp",
            "UefiTarget"
          ],
          "UefiTargetBootSourceOverride": "None",
          "UefiTargetBootSourceOverride@Redfish.AllowableValues": [
            "UsbClass(0xFFFF,0xFFFF,0xFF,0xFF,0xFF)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x0,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x1,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x2,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x3,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x4,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x5,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x6,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x3)/Pci(0x0,0x0)/Sata(0x7,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x2)/Pci(0x0,0x0)/Sata(0x6,0x0,0x0)",
            "PciRoot(0x1)/Pci(0x8,0x2)/Pci(0x0,0x0)/Sata(0x7,0x0,0x0)",
            "PciRoot(0x0)/Pci(0x1,0x1)/Pci(0x0,0x0)/MAC(1402ECD93C80,0x1)/IPv4(0.0.0.0)/Uri()",
            "PciRoot(0x0)/Pci(0x1,0x1)/Pci(0x0,0x0)/MAC(1402ECD93C80,0x1)/IPv4(0.0.0.0)",
            "PciRoot(0x0)/Pci(0x1,0x1)/Pci(0x0,0x0)/MAC(1402ECD93C80,0x1)/IPv6(0000:0000:0000:0000:0000:0000:0000:0000)/Uri()",
            "PciRoot(0x0)/Pci(0x1,0x1)/Pci(0x0,0x0)/MAC(1402ECD93C80,0x1)/IPv6(0000:0000:0000:0000:0000:0000:0000:0000)"
          ]
        },
        "EthernetInterfaces": {
          "@odata.id": "/redfish/v1/Systems/1/EthernetInterfaces"
        },
        "HostName": "",
        "IndicatorLED": "Off",
        "Links": {
          "ManagedBy": [
            {
              "@odata.id": "/redfish/v1/Managers/1"
            }
          ],
          "Chassis": [
            {
              "@odata.id": "/redfish/v1/Chassis/1"
            }
          ]
        },
        "LogServices": {
          "@odata.id": "/redfish/v1/Systems/1/LogServices"
        },
        "Manufacturer": "HPE",
        "Memory": {
          "@odata.id": "/redfish/v1/Systems/1/Memory"
        },
        "MemoryDomains": {
          "@odata.id": "/redfish/v1/Systems/1/MemoryDomains"
        },
        "MemorySummary": {
          "Status": {
            "HealthRollup": "OK"
          },
          "TotalSystemMemoryGiB": 128,
          "TotalSystemPersistentMemoryGiB": 0
        },
        "Model": "ProLiant DL325 Gen10 Plus",
        "Name": "Computer System",
        "NetworkInterfaces": {
          "@odata.id": "/redfish/v1/Systems/1/NetworkInterfaces"
        },
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeComputerSystemExt.HpeComputerSystemExt",
            "@odata.type": "#HpeComputerSystemExt.v2_7_0.HpeComputerSystemExt",
            "Actions": {
              "#HpeComputerSystemExt.PowerButton": {
                "PushType@Redfish.AllowableValues": [
                  "Press",
                  "PressAndHold"
                ],
                "target": "/redfish/v1/Systems/1/Actions/Oem/Hpe/HpeComputerSystemExt.PowerButton"
              },
              "#HpeComputerSystemExt.RestoreManufacturingDefaults": {
                "target": "/redfish/v1/Systems/1/Actions/Oem/Hpe/HpeComputerSystemExt.RestoreManufacturingDefaults"
              },
              "#HpeComputerSystemExt.RestoreSystemDefaults": {
                "target": "/redfish/v1/Systems/1/Actions/Oem/Hpe/HpeComputerSystemExt.RestoreSystemDefaults"
              },
              "#HpeComputerSystemExt.SecureSystemErase": {
                "target": "/redfish/v1/Systems/1/Actions/Oem/Hpe/HpeComputerSystemExt.SecureSystemErase"
              },
              "#HpeComputerSystemExt.ServerIntelligentDiagnosticsMode": {
                "target": "/redfish/v1/Systems/1/Actions/Oem/Hpe/HpeComputerSystemExt.ServerIntelligentDiagnosticsMode"
              },
              "#HpeComputerSystemExt.ServerSafeMode": {
                "target": "/redfish/v1/Systems/1/Actions/Oem/Hpe/HpeComputerSystemExt.ServerSafeMode"
              },
              "#HpeComputerSystemExt.SystemReset": {
                "ResetType@Redfish.AllowableValues": [
                  "ColdBoot",
                  "AuxCycle"
                ],
                "target": "/redfish/v1/Systems/1/Actions/Oem/Hpe/HpeComputerSystemExt.SystemReset"
              }
            },
            "AggregateHealthStatus": {
              "AgentlessManagementService": "Unavailable",
              "BiosOrHardwareHealth": {
                "Status": {
                  "Health": "OK"
                }
              },
              "FanRedundancy": "Redundant",
              "Fans": {
                "Status": {
                  "Health": "OK"
                }
              },
              "Memory": {
                "Status": {
                  "Health": "OK"
                }
              },
              "Network": {
                "Status": {
                  "Health": "OK"
                }
              },
              "PowerSupplies": {
                "PowerSuppliesMismatch": false,
                "Status": {
                  "Health": "OK"
                }
              },
              "PowerSupplyRedundancy": "Redundant",
              "Processors": {
                "Status": {
                  "Health": "OK"
                }
              },
              "Storage": {
                "Status": {
                  "Health": "OK"
                }
              },
              "Temperatures": {
                "Status": {
                  "Health": "OK"
                }
              }
            },
            "Bios": {
              "Backup": {
                "Date": "02/28/2020",
                "Family": "A43",
                "VersionString": "A43 v1.20 (02/28/2020)"
              },
              "Current": {
                "Date": "02/28/2020",
                "Family": "A43",
                "VersionString": "A43 v1.20 (02/28/2020)"
              },
              "UefiClass": 2
            },
            "CurrentPowerOnTimeSeconds": 5179955,
            "DeviceDiscoveryComplete": {
              "AMSDeviceDiscovery": "NoAMS",
              "DeviceDiscovery": "vMainDeviceDiscoveryComplete",
              "SmartArrayDiscovery": "Complete"
            },
            "ElapsedEraseTimeInMinutes": 0,
            "EndOfPostDelaySeconds": null,
            "EstimatedEraseTimeInMinutes": 0,
            "IntelligentProvisioningAlwaysOn": true,
            "IntelligentProvisioningIndex": 7,
            "IntelligentProvisioningLocation": "System Board",
            "IntelligentProvisioningVersion": "3.40.192",
            "IsColdBooting": false,
            "Links": {
              "PCIDevices": {
                "@odata.id": "/redfish/v1/Systems/1/PCIDevices"
              },
              "PCISlots": {
                "@odata.id": "/redfish/v1/Systems/1/PCISlots"
              },
              "NetworkAdapters": {
                "@odata.id": "/redfish/v1/Systems/1/BaseNetworkAdapters"
              },
              "SmartStorage": {
                "@odata.id": "/redfish/v1/Systems/1/SmartStorage"
              },
              "USBPorts": {
                "@odata.id": "/redfish/v1/Systems/1/USBPorts"
              },
              "USBDevices": {
                "@odata.id": "/redfish/v1/Systems/1/USBDevices"
              },
              "EthernetInterfaces": {
                "@odata.id": "/redfish/v1/Systems/1/EthernetInterfaces"
              },
              "WorkloadPerformanceAdvisor": {
                "@odata.id": "/redfish/v1/Systems/1/WorkloadPerformanceAdvisor"
              }
            },
            "PCAPartNumber": "P12579-001",
            "PCASerialNumber": "PWWSE0ARHDK0AS",
            "PostDiscoveryCompleteTimeStamp": "2020-11-13T17:49:44Z",
            "PostDiscoveryMode": null,
            "PostMode": null,
            "PostState": "FinishedPost",
            "PowerAllocationLimit": 1000,
            "PowerAutoOn": "Restore",
            "PowerOnDelay": "Minimum",
            "PowerOnMinutes": 159823,
            "PowerRegulatorMode": "Unknown",
            "PowerRegulatorModesSupported": [
              "OSControl",
              "Max"
            ],
            "SMBIOS": {
              "extref": "/smbios"
            },
            "ServerFQDN": "",
            "ServerIntelligentDiagnosticsModeEnabled": false,
            "ServerSafeModeEnabled": false,
            "SystemROMAndiLOEraseComponentStatus": {
              "BIOSSettingsEraseStatus": "Idle",
              "iLOSettingsEraseStatus": "Idle"
            },
            "SystemROMAndiLOEraseStatus": "Idle",
            "UserDataEraseComponentStatus": {},
            "UserDataEraseStatus": "Idle",
            "VirtualProfile": "Inactive"
          }
        },
        "PowerState": "On",
        "ProcessorSummary": {
          "Count": 1,
          "Model": "AMD EPYC 7302P 16-Core Processor               ",
          "Status": {
            "HealthRollup": "OK"
          }
        },
        "Processors": {
          "@odata.id": "/redfish/v1/Systems/1/Processors"
        },
        "SKU": "P18606-B21",
        "SecureBoot": {
          "@odata.id": "/redfish/v1/Systems/1/SecureBoot"
        },
        "SerialNumber": "MXQ019020H",
        "Status": {
          "Health": "OK",
          "HealthRollup": "OK",
          "State": "Enabled"
        },
        "Storage": {
          "@odata.id": "/redfish/v1/Systems/1/Storage"
        },
        "SystemType": "Physical",
        "TrustedModules": [
          {
            "FirmwareVersion": "73.64",
            "InterfaceType": "TPM2_0",
            "Oem": {
              "Hpe": {
                "@odata.context": "/redfish/v1/$metadata#HpeTrustedModuleExt.HpeTrustedModuleExt",
                "@odata.type": "#HpeTrustedModuleExt.v2_0_0.HpeTrustedModuleExt",
                "VendorName": "STMicro"
              }
            },
            "Status": {
              "Health": "OK",
              "State": "Enabled"
            }
          }
        ],
        "UUID": "36383150-3630-584D-5130-313930323048"
      }
`
const testPathPRLT_systems_1_ethernet_interfaces = "/redfish/v1/Systems/1/EthernetInterfaces"

const testPayloadPRLT_systems_1_ethernet_interfaces = `{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "@odata.etag": "W/\"570254F2\"",
        "@odata.id": "/redfish/v1/Systems/1/EthernetInterfaces/",
        "@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "Description": "Collection of System Ethernet Interfaces",
        "Name": "System Ethernet Interfaces",
        "Members": [
          {
            "@odata.id": "/redfish/v1/Systems/1/EthernetInterfaces/1/"
          },
          {
            "@odata.id": "/redfish/v1/Systems/1/EthernetInterfaces/2/"
          }
        ],
        "Members@odata.count": 2
      }
`
const testPathPRLT_systems_1_ethernet_interfaces_1 = "/redfish/v1/Systems/1/EthernetInterfaces/1/"

const testPayloadPRLT_systems_1_ethernet_interfaces_1 = `{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.etag": "W/\"C42D3985\"",
        "@odata.id": "/redfish/v1/Systems/1/EthernetInterfaces/1/",
        "@odata.type": "#EthernetInterface.v1_4_1.EthernetInterface",
        "Id": "1",
        "FullDuplex": true,
        "IPv4Addresses": [],
        "IPv4StaticAddresses": [],
        "IPv6AddressPolicyTable": [],
        "IPv6Addresses": [],
        "IPv6StaticAddresses": [],
        "IPv6StaticDefaultGateways": [],
        "InterfaceEnabled": null,
        "LinkStatus": "LinkUp",
        "MACAddress": "14:02:ec:d9:3c:80",
        "Name": "",
        "NameServers": [],
        "SpeedMbps": null,
        "StaticNameServers": [],
        "Status": {
          "Health": "OK",
          "State": "Enabled"
        },
        "UefiDevicePath": "PciRoot(0x0)/Pci(0x1,0x1)/Pci(0x0,0x0)"
      }
`
const testPathPRLT_systems_1_ethernet_interfaces_2 = "/redfish/v1/Systems/1/EthernetInterfaces/2/"

const testPayloadPRLT_systems_1_ethernet_interfaces_2 = `{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.etag": "W/\"4C8024A2\"",
        "@odata.id": "/redfish/v1/Systems/1/EthernetInterfaces/2/",
        "@odata.type": "#EthernetInterface.v1_4_1.EthernetInterface",
        "Id": "2",
        "FullDuplex": true,
        "IPv4Addresses": [],
        "IPv4StaticAddresses": [],
        "IPv6AddressPolicyTable": [],
        "IPv6Addresses": [],
        "IPv6StaticAddresses": [],
        "IPv6StaticDefaultGateways": [],
        "InterfaceEnabled": null,
        "LinkStatus": "LinkUp",
        "MACAddress": "14:02:ec:d9:3c:81",
        "Name": "",
        "NameServers": [],
        "SpeedMbps": null,
        "StaticNameServers": [],
        "Status": {
          "Health": "OK",
          "State": "Enabled"
        }
      }
`
const testPathPRLT_systems_1_processors = "/redfish/v1/Systems/1/Processors"

const testPayloadPRLT_systems_1_processors = `{
        "@odata.context": "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
        "@odata.etag": "W/\"AA6D42B0\"",
        "@odata.id": "/redfish/v1/Systems/1/Processors",
        "@odata.type": "#ProcessorCollection.ProcessorCollection",
        "Description": "Processors view",
        "Name": "Processors Collection",
        "Members": [
          {
            "@odata.id": "/redfish/v1/Systems/1/Processors/1"
          }
        ],
        "Members@odata.count": 1
      }
`
const testPathPRLT_systems_1_processors_1 = "/redfish/v1/Systems/1/Processors/1"

const testPayloadPRLT_systems_1_processors_1 = `{
        "@odata.context": "/redfish/v1/$metadata#Processor.Processor",
        "@odata.etag": "W/\"772E9943\"",
        "@odata.id": "/redfish/v1/Systems/1/Processors/1",
        "@odata.type": "#Processor.v1_0_0.Processor",
        "Id": "1",
        "InstructionSet": "x86-64",
        "Manufacturer": "Advanced Micro Devices, Inc.",
        "MaxSpeedMHz": 3100,
        "Model": "AMD EPYC 7302P 16-Core Processor               ",
        "Name": "Processors",
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeProcessorExt.HpeProcessorExt",
            "@odata.type": "#HpeProcessorExt.v2_0_0.HpeProcessorExt",
            "AssetTag": "Unknown",
            "Cache": [
              {
                "Associativity": "8waySetAssociative",
                "CacheSpeedns": 1,
                "CurrentSRAMType": [
                  "Pipeline Burst"
                ],
                "EccType": "MultiBitECC",
                "InstalledSizeKB": 1024,
                "Location": "Internal",
                "MaximumSizeKB": 1024,
                "Name": "L1-Cache",
                "Policy": "WriteBack",
                "Socketed": false,
                "SupportedSRAMType": [
                  "Pipeline Burst"
                ],
                "SystemCacheType": "Unified"
              },
              {
                "Associativity": "8waySetAssociative",
                "CacheSpeedns": 1,
                "CurrentSRAMType": [
                  "Pipeline Burst"
                ],
                "EccType": "MultiBitECC",
                "InstalledSizeKB": 8192,
                "Location": "Internal",
                "MaximumSizeKB": 8192,
                "Name": "L2-Cache",
                "Policy": "WriteBack",
                "Socketed": false,
                "SupportedSRAMType": [
                  "Pipeline Burst"
                ],
                "SystemCacheType": "Unified"
              },
              {
                "Associativity": "16waySetAssociative",
                "CacheSpeedns": 1,
                "CurrentSRAMType": [
                  "Pipeline Burst"
                ],
                "EccType": "MultiBitECC",
                "InstalledSizeKB": 131072,
                "Location": "Internal",
                "MaximumSizeKB": 131072,
                "Name": "L3-Cache",
                "Policy": "WriteBack",
                "Socketed": false,
                "SupportedSRAMType": [
                  "Pipeline Burst"
                ],
                "SystemCacheType": "Unified"
              }
            ],
            "Characteristics": [
              "64Bit",
              "MultiCore",
              "HwThread",
              "ExecuteProtection",
              "EnhancedVirtualization",
              "PowerPerfControl"
            ],
            "ConfigStatus": {
              "Populated": true,
              "State": "Enabled"
            },
            "CoresEnabled": 16,
            "ExternalClockMHz": 100,
            "MicrocodePatches": [
              {
                "CpuId": "0x00008310",
                "Date": "2019-10-24T00:00:00Z",
                "PatchId": "0x08301034"
              },
              {
                "CpuId": "0x00008300",
                "Date": "2019-04-01T00:00:00Z",
                "PatchId": "0x08300027"
              }
            ],
            "PartNumber": "",
            "RatedSpeedMHz": 3000,
            "SerialNumber": "",
            "VoltageVoltsX10": 11
          }
        },
        "ProcessorArchitecture": "x86",
        "ProcessorId": {
          "EffectiveFamily": "107",
          "EffectiveModel": "1",
          "IdentificationRegisters": "0x0f100083fbff178b",
          "MicrocodeInfo": null,
          "Step": "0",
          "VendorId": "Advanced Micro Devices, Inc."
        },
        "ProcessorType": "CPU",
        "Socket": "Proc 1",
        "Status": {
          "Health": "OK",
          "State": "Enabled"
        },
        "TotalCores": 16,
        "TotalThreads": 32
      }
`
const testPathPRLT_systems_1_memory = "/redfish/v1/Systems/1/Memory"

const testPayloadPRLT_systems_1_memory = `{
        "@odata.context": "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
        "@odata.etag": "W/\"C687E11B\"",
        "@odata.id": "/redfish/v1/Systems/1/Memory",
        "@odata.type": "#MemoryCollection.MemoryCollection",
        "Description": "Memory DIMM Collection",
        "Name": "Memory DIMM Collection",
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeAdvancedMemoryProtection.HpeAdvancedMemoryProtection",
            "@odata.type": "#HpeAdvancedMemoryProtection.v2_0_0.HpeAdvancedMemoryProtection",
            "AmpModeActive": "AdvancedECC",
            "AmpModeStatus": "AdvancedECC",
            "AmpModeSupported": [
              "AdvancedECC"
            ],
            "MemoryList": [
              {
                "BoardCpuNumber": 1,
                "BoardNumberOfSockets": 16,
                "BoardOperationalFrequency": 3200,
                "BoardOperationalVoltage": 1200,
                "BoardTotalMemorySize": 131072
              }
            ]
          }
        },
        "Members": [
          {
            "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm1"
          },
          {
            "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm2"
          },
          {
            "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm3"
          },
          {
            "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm4"
          }
        ],
        "Members@odata.count": 4
      }
`
const testPathPRLT_systems_1_memory_proc1dimm1 = "/redfish/v1/Systems/1/Memory/proc1dimm1"

const testPayloadPRLT_systems_1_memory_proc1dimm1 = `{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.etag": "W/\"E6EC3A2C\"",
        "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm1",
        "@odata.type": "#Memory.v1_7_1.Memory",
        "Id": "proc1dimm1",
        "BaseModuleType": "RDIMM",
        "BusWidthBits": 72,
        "CacheSizeMiB": 0,
        "CapacityMiB": 16384,
        "DataWidthBits": 64,
        "DeviceLocator": "PROC 1 DIMM 1",
        "ErrorCorrection": "MultiBitECC",
        "LogicalSizeMiB": 0,
        "Manufacturer": "Hynix",
        "MemoryDeviceType": "DDR4",
        "MemoryLocation": {
          "Channel": 3,
          "MemoryController": 3,
          "Slot": 1,
          "Socket": 1
        },
        "MemoryMedia": [
          "DRAM"
        ],
        "MemoryType": "DRAM",
        "Name": "proc1dimm1",
        "NonVolatileSizeMiB": 0,
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeMemoryExt.HpeMemoryExt",
            "@odata.type": "#HpeMemoryExt.v2_5_0.HpeMemoryExt",
            "Attributes": [
              "HpeStandardMemory"
            ],
            "BaseModuleType": "RDIMM",
            "DIMMStatus": "GoodInUse",
            "MaxOperatingSpeedMTs": 3200,
            "MinimumVoltageVoltsX10": 12,
            "VendorName": "SK Hynix"
          }
        },
        "OperatingMemoryModes": [
          "Volatile"
        ],
        "OperatingSpeedMhz": 3200,
        "PartNumber": "P11442-191",
        "PersistentRegionSizeLimitMiB": 0,
        "RankCount": 1,
        "SecurityCapabilities": {},
        "SerialNumber": "438AED20",
        "Status": {
          "Health": "OK",
          "State": "Enabled"
        },
        "VendorID": "44288",
        "VolatileRegionSizeLimitMiB": 16384,
        "VolatileSizeMiB": 16384
      }
`
const testPathPRLT_systems_1_memory_proc1dimm2 = "/redfish/v1/Systems/1/Memory/proc1dimm2"

const testPayloadPRLT_systems_1_memory_proc1dimm2 = `{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.etag": "W/\"6165906D\"",
        "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm2",
        "@odata.type": "#Memory.v1_7_1.Memory",
        "Id": "proc1dimm2",
        "BusWidthBits": 72,
        "CacheSizeMiB": 0,
        "CapacityMiB": 0,
        "DataWidthBits": 64,
        "DeviceLocator": "PROC 1 DIMM 2",
        "ErrorCorrection": "MultiBitECC",
        "LogicalSizeMiB": 0,
        "MemoryLocation": {
          "Channel": 3,
          "MemoryController": 3,
          "Slot": 2,
          "Socket": 1
        },
        "MemoryMedia": [
          "DRAM"
        ],
        "MemoryType": "DRAM",
        "Name": "proc1dimm2",
        "NonVolatileSizeMiB": 0,
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeMemoryExt.HpeMemoryExt",
            "@odata.type": "#HpeMemoryExt.v2_5_0.HpeMemoryExt",
            "DIMMStatus": "NotPresent",
            "MinimumVoltageVoltsX10": 0
          }
        },
        "OperatingMemoryModes": [
          "Volatile"
        ],
        "PersistentRegionSizeLimitMiB": 0,
        "RankCount": null,
        "SecurityCapabilities": {},
        "Status": {
          "Health": "OK",
          "State": "Absent"
        },
        "VendorID": "0",
        "VolatileRegionSizeLimitMiB": 0,
        "VolatileSizeMiB": 0
      }
`
const testPathPRLT_systems_1_memory_proc1dimm3 = "/redfish/v1/Systems/1/Memory/proc1dimm3"

const testPayloadPRLT_systems_1_memory_proc1dimm3 = `{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.etag": "W/\"E6EC3A2C\"",
        "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm3",
        "@odata.type": "#Memory.v1_7_1.Memory",
        "Id": "proc1dimm3",
        "BaseModuleType": "RDIMM",
        "BusWidthBits": 72,
        "CacheSizeMiB": 0,
        "CapacityMiB": 16384,
        "DataWidthBits": 64,
        "DeviceLocator": "PROC 1 DIMM 3",
        "ErrorCorrection": "MultiBitECC",
        "LogicalSizeMiB": 0,
        "Manufacturer": "Hynix",
        "MemoryDeviceType": "DDR4",
        "MemoryLocation": {
          "Channel": 4,
          "MemoryController": 4,
          "Slot": 3,
          "Socket": 1
        },
        "MemoryMedia": [
          "DRAM"
        ],
        "MemoryType": "DRAM",
        "Name": "proc1dimm3",
        "NonVolatileSizeMiB": 0,
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeMemoryExt.HpeMemoryExt",
            "@odata.type": "#HpeMemoryExt.v2_5_0.HpeMemoryExt",
            "Attributes": [
              "HpeStandardMemory"
            ],
            "BaseModuleType": "RDIMM",
            "DIMMStatus": "GoodInUse",
            "MaxOperatingSpeedMTs": 3200,
            "MinimumVoltageVoltsX10": 12,
            "VendorName": "SK Hynix"
          }
        },
        "OperatingMemoryModes": [
          "Volatile"
        ],
        "OperatingSpeedMhz": 3200,
        "PartNumber": "P11442-191",
        "PersistentRegionSizeLimitMiB": 0,
        "RankCount": 1,
        "SecurityCapabilities": {},
        "SerialNumber": "438AED1F",
        "Status": {
          "Health": "OK",
          "State": "Enabled"
        },
        "VendorID": "44288",
        "VolatileRegionSizeLimitMiB": 16384,
        "VolatileSizeMiB": 16384
      }
`
const testPathPRLT_systems_1_memory_proc1dimm4 = "/redfish/v1/Systems/1/Memory/proc1dimm4"

const testPayloadPRLT_systems_1_memory_proc1dimm4 = `{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.etag": "W/\"6165906D\"",
        "@odata.id": "/redfish/v1/Systems/1/Memory/proc1dimm4",
        "@odata.type": "#Memory.v1_7_1.Memory",
        "Id": "proc1dimm4",
        "BusWidthBits": 72,
        "CacheSizeMiB": 0,
        "CapacityMiB": 0,
        "DataWidthBits": 64,
        "DeviceLocator": "PROC 1 DIMM 4",
        "ErrorCorrection": "MultiBitECC",
        "LogicalSizeMiB": 0,
        "MemoryLocation": {
          "Channel": 4,
          "MemoryController": 4,
          "Slot": 4,
          "Socket": 1
        },
        "MemoryMedia": [
          "DRAM"
        ],
        "MemoryType": "DRAM",
        "Name": "proc1dimm4",
        "NonVolatileSizeMiB": 0,
        "Oem": {
          "Hpe": {
            "@odata.context": "/redfish/v1/$metadata#HpeMemoryExt.HpeMemoryExt",
            "@odata.type": "#HpeMemoryExt.v2_5_0.HpeMemoryExt",
            "DIMMStatus": "NotPresent",
            "MinimumVoltageVoltsX10": 0
          }
        },
        "OperatingMemoryModes": [
          "Volatile"
        ],
        "PersistentRegionSizeLimitMiB": 0,
        "RankCount": null,
        "SecurityCapabilities": {},
        "Status": {
          "Health": "OK",
          "State": "Absent"
        },
        "VendorID": "0",
        "VolatileRegionSizeLimitMiB": 0,
        "VolatileSizeMiB": 0
      }
`
const testPathPRLT_systems_1_storage = "/redfish/v1/Systems/1/Storage"

const testPayloadPRLT_systems_1_storage = `{
        "@odata.context": "/redfish/v1/$metadata#StorageCollection.StorageCollection",
        "@odata.etag": "W/\"75983E8D\"",
        "@odata.id": "/redfish/v1/Systems/1/Storage",
        "@odata.type": "#StorageCollection.StorageCollection",
        "Description": "Storage View",
        "Name": "Storage",
        "Members": [],
        "Members@odata.count": 0
      }
`

//////////////////////////////////////////////////////////////////////////////
//                         GBT - Mock Client
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncGBT1() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathGBT_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_redfish_v1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_chassis:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_chassis)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_chassis_self:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_chassis_self)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_managers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_managers)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_managers_self:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_managers_self)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_managers_self_reset_action_info:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_managers_self_reset_action_info)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_managers_self_ethernet_interfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_managers_self_ethernet_interfaces)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_managers_self_ethernet_interfaces_bond0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_managers_self_ethernet_interfaces_bond0)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_managers_self_ethernet_interfaces_usb0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_managers_self_ethernet_interfaces_usb0)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_reset_action_info:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_reset_action_info)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_chassis_self_power:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_chassis_self_power)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_chassis_self_network_adapters:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_chassis_self_network_adapters)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_ethernet_interfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_ethernet_interfaces)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_ethernet_interfaces_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_ethernet_interfaces_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_ethernet_interfaces_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_ethernet_interfaces_2)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_processors:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_processors)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_processors_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_processors_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_processors_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_processors_2)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_memory:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_memory)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_memory_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_memory_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_memory_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_memory_2)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_storage:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_storage)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathGBT_systems_self_storage_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadGBT_systems_self_storage_1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString("")),

				Header: make(http.Header),
			}
		}
	}
}

const testPathGBT_redfish_v1 = "/redfish/v1"

const testPayloadGBT_redfish_v1 = `
{
   "Name" : "Root Service",
   "@odata.type" : "#ServiceRoot.v1_4_0.ServiceRoot",
   "Chassis" : {
      "@odata.id" : "/redfish/v1/Chassis"
   },
   "UUID" : "b42e99b5-d713-d603-0010-debfa0b1536e",
   "Links" : {
      "Sessions" : {
         "@odata.id" : "/redfish/v1/SessionService/Sessions"
      }
   },
   "Product" : "AMI Redfish Server",
   "Systems" : {
      "@odata.id" : "/redfish/v1/Systems"
   },
   "Managers" : {
      "@odata.id" : "/redfish/v1/Managers"
   },
   "@odata.context" : "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
   "ProtocolFeaturesSupported" : {
      "SelectQuery" : true,
      "ExpandQuery" : {
         "MaxLevels" : 5,
         "ExpandAll" : true,
         "NoLinks" : true,
         "Links" : true,
         "Levels" : true
      },
      "FilterQuery" : true
   },
   "Registries" : {
      "@odata.id" : "/redfish/v1/Registries"
   },
   "TelemetryService" : {
      "@odata.id" : "/redfish/v1/TelemetryService"
   },
   "@odata.etag" : "W/\"1584650690\"",
   "RedfishVersion" : "1.6.0",
   "Description" : "The service root for all Redfish requests on this host",
   "Oem" : {
      "Ami" : {
         "@odata.type" : "#AMIServiceRoot.v1_0_0.AMIServiceRoot",
         "RtpVersion" : "1.5.b",
         "Configurations" : {
            "@odata.id" : "/redfish/v1/Configurations"
         }
      },
      "Dre" : {
         "@odata.type" : "#AMIDynamicExtension.v1_0_0.AMIDynamicExtension",
         "DynamicExtension" : {
            "@odata.id" : "/redfish/v1/DynamicExtension"
         }
      },
      "Vendor" : {
         "VendorVersion" : "1.6.0.111",
         "@odata.type" : "#VendorServiceRoot.v1_0_0.VendorServiceRoot"
      }
   },
   "JsonSchemas" : {
      "@odata.id" : "/redfish/v1/JsonSchemas"
   },
   "JobService" : {
      "@odata.id" : "/redfish/v1/JobService"
   },
   "@odata.id" : "/redfish/v1/",
   "Id" : "RootService"
}
`

const testPathGBT_chassis = "/redfish/v1/Chassis"

const testPayloadGBT_chassis = `
{
   "Name" : "Chassis Collection",
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   ],
   "@odata.id" : "/redfish/v1/Chassis",
   "@odata.context" : "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
   "Members@odata.count" : 1,
   "@odata.type" : "#ChassisCollection.ChassisCollection",
   "Description" : "The Collection for Chassis",
   "@odata.etag" : "W/\"1584727014\""
}
`

const testPathGBT_chassis_self = "/redfish/v1/Chassis/Self"

const testPayloadGBT_chassis_self = `
{
   "SKU" : "01234567890123456789AB",
   "IndicatorLED@Redfish.AllowableValues" : [
      "Lit",
      "Blinking",
      "Off"
   ],
   "@odata.id" : "/redfish/v1/Chassis/Self",
   "PartNumber" : "6NH262Z63MR-00-102",
   "IndicatorLED" : "Off",
   "NetworkAdapters" : {
      "@odata.id" : "/redfish/v1/Chassis/Self/NetworkAdapters"
   },
   "Actions" : {
      "#Chassis.Reset" : {
         "target" : "/redfish/v1/Chassis/Self/Actions/Chassis.Reset",
         "@Redfish.ActionInfo" : "/redfish/v1/Chassis/Self/ResetActionInfo"
      }
   },
   "AssetTag" : "01234567890123456789AB",
   "UUID" : "b42e99b5-d713-d603-0010-debfa0b1536e",
   "@odata.etag" : "W/\"1584727014\"",
   "@odata.context" : "/redfish/v1/$metadata#Chassis.Chassis",
   "PowerState" : "On",
   "Links" : {
      "ManagedBy@odata.count" : 1,
      "ManagedBy" : [
         {
            "@odata.id" : "/redfish/v1/Managers/Self"
         }
      ],
      "ManagersInChassis" : [
         {
            "@odata.id" : "/redfish/v1/Managers/Self"
         }
      ],
      "ResourceBlocks" : [
         {
            "@odata.id" : "/redfish/v1/CompositionService/ResourceBlocks/ComputeBlock"
         },
         {
            "@odata.id" : "/redfish/v1/CompositionService/ResourceBlocks/DrivesBlock"
         },
         {
            "@odata.id" : "/redfish/v1/CompositionService/ResourceBlocks/NetworkBlock"
         }
      ],
      "PCIeDevices@odata.count" : 1,
      "ManagersInChassis@odata.count" : 1,
      "ResourceBlocks@odata.count" : 3,
      "ComputerSystems@odata.count" : 1,
      "ComputerSystems" : [
         {
            "@odata.id" : "/redfish/v1/Systems/Self"
         }
      ],
      "Drives@odata.count" : 0,
      "PCIeDevices" : [
         {
            "@odata.id" : "/redfish/v1/Chassis/Self/PCIeDevices/1"
         }
      ]
   },
   "Manufacturer" : "Cray Inc.",
   "SerialNumber" : "GJGAN7012A0142",
   "Name" : "Computer System Chassis",
   "LogServices" : {
      "@odata.id" : "/redfish/v1/Chassis/Self/LogServices"
   },
   "@odata.type" : "#Chassis.v1_8_0.Chassis",
   "Thermal" : {
      "@odata.id" : "/redfish/v1/Chassis/Self/Thermal"
   },
   "Power" : {
      "@odata.id" : "/redfish/v1/Chassis/Self/Power"
   },
   "Oem" : {
      "GBTChassisOemProperty" : {
         "@odata.type" : "#GBTChassisOemProperty.v1_0_0.GBTChassisOemProperty",
         "Board Serial Number" : "JN9N0700685"
      }
   },
   "Status" : {
      "State" : "Enabled",
      "Health" : "OK",
      "HealthRollup" : "OK"
   },
   "Id" : "Self",
   "Model" : "H262-Z63-YF",
   "ChassisType" : "Other",
   "Description" : "Chassis Self"
}
`

const testPathGBT_managers = "/redfish/v1/Managers"

const testPayloadGBT_managers = `
{
   "Members@odata.count" : 1,
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Managers/Self"
      }
   ],
   "Description" : "The collection for Managers",
   "Name" : "Manager Collection",
   "@odata.id" : "/redfish/v1/Managers",
   "@odata.context" : "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
   "@odata.type" : "#ManagerCollection.ManagerCollection",
   "@odata.etag" : "W/\"1584732466\""
}
`

const testPathGBT_managers_self = "/redfish/v1/Managers/Self"

const testPayloadGBT_managers_self = `
{
   "Description" : "BMC",
   "@odata.id" : "/redfish/v1/Managers/Self",
   "FirmwareVersion" : "12.03.3",
   "CommandShell" : {
      "ServiceEnabled" : true,
      "ConnectTypesSupported" : [
         "SSH",
         "IPMI"
      ],
      "MaxConcurrentSessions" : 36
   },
   "DateTime" : "2020-03-20T19:52:46+00:00",
   "Links" : {
      "ManagerInChassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      },
      "ManagerForServers" : [
         {
            "@odata.id" : "/redfish/v1/Systems/Self"
         }
      ],
      "ManagerForChassis" : [
         {
            "@odata.id" : "/redfish/v1/Chassis/Self"
         }
      ],
      "ManagerForChassis@odata.count" : 1,
      "ManagerForServers@odata.count" : 1
   },
   "LogServices" : {
      "@odata.id" : "/redfish/v1/Managers/Self/LogServices"
   },
   "UUID" : "b42e99b5-d713-d603-0010-debfa0b1536e",
   "ServiceEntryPointUUID" : "b42e99b5-d713-d603-0010-debfa0b1536e",
   "PowerState" : "On",
   "NetworkProtocol" : {
      "@odata.id" : "/redfish/v1/Managers/Self/NetworkProtocol"
   },
   "EthernetInterfaces" : {
      "@odata.id" : "/redfish/v1/Managers/Self/EthernetInterfaces"
   },
   "Id" : "Self",
   "@odata.context" : "/redfish/v1/$metadata#Manager.Manager",
   "DateTimeLocalOffset" : "+00:00",
   "ManagerType" : "BMC",
   "Name" : "Manager",
   "SerialInterfaces" : {
      "@odata.id" : "/redfish/v1/Managers/Self/SerialInterfaces"
   },
   "@odata.etag" : "W/\"1584732466\"",
   "VirtualMedia" : {
      "@odata.id" : "/redfish/v1/Managers/Self/VirtualMedia"
   },
   "SerialConsole" : {
      "MaxConcurrentSessions" : 1,
      "ServiceEnabled" : true,
      "ConnectTypesSupported" : [
         "SSH",
         "IPMI"
      ]
   },
   "Redundancy@odata.count" : 0,
   "HostInterfaces" : {
      "@odata.id" : "/redfish/v1/Managers/Self/HostInterfaces"
   },
   "GraphicalConsole" : {
      "ServiceEnabled" : true,
      "ConnectTypesSupported" : [
         "KVMIP"
      ],
      "MaxConcurrentSessions" : 2
   },
   "Actions" : {
      "Oem" : {
         "#RedfishDBReset" : {
            "@Redfish.ActionInfo" : "/redfish/v1/Managers/Self/FactoryResetActionInfo",
            "target" : "/redfish/v1/Managers/Self/Actions/RedfishDBReset"
         }
      },
      "#Manager.Reset" : {
         "target" : "/redfish/v1/Managers/Self/Actions/Manager.Reset",
         "@Redfish.ActionInfo" : "/redfish/v1/Managers/Self/ResetActionInfo"
      }
   },
   "Status" : {
      "Health" : "OK",
      "State" : "Enabled"
   },
   "Model" : "410810600",
   "@odata.type" : "#Manager.v1_5_0.Manager"
}
`

const testPathGBT_managers_self_reset_action_info = "/redfish/v1/Managers/Self/ResetActionInfo"

const testPayloadGBT_managers_self_reset_action_info = `
{
  "@odata.context": "/redfish/v1/$metadata#ActionInfo.ActionInfo",
  "@odata.etag": "W/\"1721044073\"",
  "@odata.id": "/redfish/v1/Managers/Self/ResetActionInfo",
  "@odata.type": "#ActionInfo.v1_1_1.ActionInfo",
  "Description": "This action is used to reset the Managers",
  "Id": "ResetAction",
  "Name": "ResetAction",
  "Parameters": [
    {
      "AllowableValues": [
        "ForceRestart"
      ],
      "DataType": "String",
      "Name": "ResetType",
      "Required": true
    }
  ]
}
`

const testPathGBT_managers_self_ethernet_interfaces = "/redfish/v1/Managers/Self/EthernetInterfaces"

const testPayloadGBT_managers_self_ethernet_interfaces = `
{
   "@odata.type" : "#EthernetInterfaceCollection.EthernetInterfaceCollection",
   "@odata.id" : "/redfish/v1/Managers/Self/EthernetInterfaces",
   "Members@odata.count" : 2,
   "Name" : "Ethernet Network Interface Collection",
   "@odata.etag" : "W/\"1584735955\"",
   "Description" : "Collection of Ethernet Interfaces for this Manager",
   "@odata.context" : "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Managers/Self/EthernetInterfaces/bond0"
      },
      {
         "@odata.id" : "/redfish/v1/Managers/Self/EthernetInterfaces/usb0"
      }
   ]
}
`

const testPathGBT_managers_self_ethernet_interfaces_bond0 = "/redfish/v1/Managers/Self/EthernetInterfaces/bond0"

const testPayloadGBT_managers_self_ethernet_interfaces_bond0 = `
{
   "InterfaceEnabled" : true,
   "HostName" : "AMIB42E99B5D713",
   "PermanentMACAddress" : "B4:2E:99:B5:D7:13",
   "Description" : "Ethernet Interface bond0",
   "FullDuplex" : true,
   "VLAN" : {
      "VLANId" : 0,
      "VLANEnable" : false
   },
   "FQDN" : "AMIB42E99B5D713.bad",
   "IPv6Addresses" : [
      {
         "Address" : "fe80::b62e:99ff:feb5:d713",
         "AddressOrigin" : "LinkLocal",
         "PrefixLength" : 64
      }
   ],
   "Name" : "bond0",
   "@odata.etag" : "W/\"1584735955\"",
   "SpeedMbps" : 1000,
   "NameServers" : [
      "::",
      "10.141.255.254"
   ],
   "Id" : "bond0",
   "DHCPv4" : {
      "DHCPEnabled" : true
   },
   "Status" : {
      "State" : "Enabled",
      "Health" : "OK"
   },
   "MaxIPv6StaticAddresses" : 16,
   "MACAddress" : "B4:2E:99:B5:D7:13",
   "AutoNeg" : true,
   "LinkStatus" : "LinkUp",
   "@odata.id" : "/redfish/v1/Managers/Self/EthernetInterfaces/bond0",
   "MTUSize" : 1500,
   "@odata.context" : "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
   "@odata.type" : "#EthernetInterface.v1_4_1.EthernetInterface",
   "IPv4Addresses" : [
      {
         "Address" : "10.254.3.15",
         "AddressOrigin" : "DHCP",
         "Gateway" : "10.254.0.1",
         "SubnetMask" : "255.255.128.0"
      }
   ]
}
`

const testPathGBT_managers_self_ethernet_interfaces_usb0 = "/redfish/v1/Managers/Self/EthernetInterfaces/usb0"

const testPayloadGBT_managers_self_ethernet_interfaces_usb0 = `
{
   "@odata.etag" : "W/\"1584735955\"",
   "Name" : "usb0",
   "MTUSize" : 1500,
   "NameServers" : [
      "::",
      "10.141.255.254"
   ],
   "LinkStatus" : "LinkUp",
   "IPv6Addresses" : [
      {
         "PrefixLength" : 64,
         "Address" : "fe80::c484:28ff:fee4:125e",
         "AddressOrigin" : "LinkLocal"
      }
   ],
   "IPv4Addresses" : [
      {
         "SubnetMask" : "255.255.240.0",
         "Address" : "169.254.0.17",
         "Gateway" : "0.0.0.0",
         "AddressOrigin" : "Static"
      }
   ],
   "MaxIPv6StaticAddresses" : 16,
   "HostName" : "AMIB42E99B5D713",
   "Status" : {
      "Health" : "OK",
      "State" : "Enabled"
   },
   "FQDN" : "AMIB42E99B5D713.bad",
   "DHCPv4" : {
      "DHCPEnabled" : false
   },
   "MACAddress" : "C6:84:28:E4:12:5E",
   "PermanentMACAddress" : "C6:84:28:E4:12:5E",
   "Description" : "Ethernet Interface usb0",
   "VLAN" : {
      "VLANId" : 0,
      "VLANEnable" : false
   },
   "@odata.type" : "#EthernetInterface.v1_4_1.EthernetInterface",
   "InterfaceEnabled" : true,
   "IPv4StaticAddresses" : [
      {
         "Gateway" : "0.0.0.0",
         "AddressOrigin" : "Static",
         "Address" : "169.254.0.17",
         "SubnetMask" : "255.255.240.0"
      }
   ],
   "Id" : "usb0",
   "@odata.id" : "/redfish/v1/Managers/Self/EthernetInterfaces/usb0",
   "@odata.context" : "/redfish/v1/$metadata#EthernetInterface.EthernetInterface"
}
`

const testPathGBT_systems = "/redfish/v1/Systems"

const testPayloadGBT_systems = `
{
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Self"
      }
   ],
   "@odata.id" : "/redfish/v1/Systems",
   "Description" : "Collection of Computer Systems",
   "@Redfish.CollectionCapabilities" : {
      "Capabilities" : [
         {
            "CapabilitiesObject" : {
               "@odata.id" : "/redfish/v1/Systems/Capabilities"
            },
            "Links" : {
               "TargetCollection" : {
                  "@odata.id" : "/redfish/v1/Systems"
               }
            },
            "UseCase" : "ComputerSystemComposition"
         }
      ],
      "@odata.type" : "#CollectionCapabilities.v1_1_0.CollectionCapabilities"
   },
   "Members@odata.count" : 1,
   "@odata.context" : "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
   "Name" : "Systems Collection",
   "@odata.etag" : "W/\"1584634868\"",
   "@odata.type" : "#ComputerSystemCollection.ComputerSystemCollection"
}
`

const testPathGBT_systems_self = "/redfish/v1/Systems/Self"

const testPayloadGBT_systems_self = `
{
   "Boot" : {
      "BootSourceOverrideTarget@Redfish.AllowableValues" : [
         "None",
         "Pxe",
         "Floppy",
         "Cd",
         "Usb",
         "Hdd",
         "BiosSetup",
         "Utilities",
         "Diags",
         "UefiShell",
         "UefiTarget",
         "SDCard",
         "UefiHttp",
         "RemoteDrive",
         "UefiBootNext"
      ],
      "BootSourceOverrideTarget" : "None",
      "BootOptions" : {
         "@odata.id" : "/redfish/v1/Systems/Self/BootOptions"
      },
      "BootSourceOverrideMode" : "Legacy",
      "BootSourceOverrideEnabled@Redfish.AllowableValues" : [
         "Disabled",
         "Once",
         "Continuous"
      ],
      "BootSourceOverrideMode@Redfish.AllowableValues" : [
         "Legacy",
         "UEFI"
      ],
      "BootSourceOverrideEnabled" : "Disabled"
   },
   "Description" : "System Self",
   "PartNumber" : "000000000001",
   "@odata.type" : "#ComputerSystem.v1_5_1.ComputerSystem",
   "Bios" : {
      "@odata.id" : "/redfish/v1/Systems/Self/Bios"
   },
   "IndicatorLED" : "Off",
   "Memory" : {
      "@odata.id" : "/redfish/v1/Systems/Self/Memory"
   },
   "EthernetInterfaces" : {
      "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces"
   },
   "@odata.etag" : "W/\"1584634868\"",
   "IndicatorLED@Redfish.AllowableValues" : [
      "Lit",
      "Blinking",
      "Off"
   ],
   "Id" : "Self",
   "@odata.id" : "/redfish/v1/Systems/Self",
   "NetworkInterfaces" : {
      "@odata.id" : "/redfish/v1/Systems/Self/NetworkInterfaces"
   },
   "Storage" : {
      "@odata.id" : "/redfish/v1/Systems/Self/Storage"
   },
   "PCIeDevices@odata.count" : 1,
   "@odata.context" : "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
   "Processors" : {
      "@odata.id" : "/redfish/v1/Systems/Self/Processors"
   },
   "BiosVersion" : "C12",
   "Manufacturer" : "Cray Inc.",
   "Status" : {
      "Health" : "OK",
      "State" : "Enabled",
      "HealthRollup" : "OK"
   },
   "MemoryDomains" : {
      "@odata.id" : "/redfish/v1/Systems/Self/MemoryDomains"
   },
   "PowerState" : "On",
   "Model" : "H262-Z63-YF",
   "MemorySummary" : {
      "TotalSystemMemoryGiB" : 244
   },
   "ProcessorSummary" : {
      "Model" : "AMD EPYC 7542 32-Core Processor                ",
      "Count" : 2
   },
   "Links" : {
      "Chassis" : [
         {
            "@odata.id" : "/redfish/v1/Chassis/Self"
         }
      ],
      "ManagedBy@odata.count" : 1,
      "ManagedBy" : [
         {
            "@odata.id" : "/redfish/v1/Managers/Self"
         }
      ],
      "Chassis@odata.count" : 1
   },
   "SimpleStorage" : {
      "@odata.id" : "/redfish/v1/Systems/Self/SimpleStorage"
   },
   "UUID" : "cd210000-3b17-11ea-8000-b42e99b5d711",
   "LogServices" : {
      "@odata.id" : "/redfish/v1/Systems/Self/LogServices"
   },
   "SKU" : "01234567890123456789AB",
   "SystemType" : "Physical",
   "AssetTag" : "Free form asset tag",
   "PCIeDevices" : [
      {
         "@odata.id" : "/redfish/v1/Chassis/Self/PCIeDevices/1"
      }
   ],
   "SerialNumber" : "GJGAN7012A014201",
   "SecureBoot" : {
      "@odata.id" : "/redfish/v1/Systems/Self/SecureBoot"
   },
   "Actions" : {
      "#ComputerSystem.Reset" : {
         "@Redfish.ActionInfo" : "/redfish/v1/Systems/Self/ResetActionInfo",
         "target" : "/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"
      }
   },
   "Name" : "System"
}
`

const testPathGBT_systems_self_reset_action_info = "/redfish/v1/Systems/Self/ResetActionInfo"

const testPayloadGBT_systems_self_reset_action_info = `
{
   "Name" : "ResetAction",
   "@odata.context" : "/redfish/v1/$metadata#ActionInfo.ActionInfo",
   "Id" : "ResetAction",
   "Parameters" : [
      {
         "Required" : true,
         "DataType" : "String",
         "Name" : "ResetType",
         "AllowableValues" : [
            "ForceRestart",
            "On",
            "ForceOff",
            "GracefulShutdown"
         ]
      }
   ],
   "@odata.type" : "#ActionInfo.v1_0_3.ActionInfo",
   "@odata.etag" : "W/\"1583182793\"",
   "Description" : "This action is used to reset the Systems",
   "@odata.id" : "/redfish/v1/Systems/Self/ResetActionInfo"
}
`

const testPathGBT_chassis_self_power = "/redfish/v1/Chassis/Self/Power"

const testPayloadGBT_chassis_self_power = `
{
   "@odata.id" : "/redfish/v1/Chassis/Self/Power",
   "Name" : "Power",
   "PowerControl" : [
      {
         "RelatedItem@odata.count" : 2,
         "PowerMetrics" : {
            "MaxConsumedWatts" : 710,
            "MinConsumedWatts" : 4,
            "AverageConsumedWatts" : 180,
            "IntervalInMin" : 0
         },
         "PowerConsumedWatts" : 195,
         "MemberId" : "0",
         "Name" : "Chassis Power Control",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Status" : {
            "Health" : "Critical",
            "State" : "Disabled"
         },
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/PowerControl/0",
         "PowerCapacityWatts" : 33795,
         "Oem" : {
            "Vendor" : {
               "@odata.type" : "#GbtPowerLimit.v1_0_0.Vendor",
               "PowerLimit" : {
                  "Min" : 15616,
                  "Factor" : 10,
                  "Max" : 33795
               },
               "PowerIdleWatts" : 15616,
               "PowerResetWatts" : 35330
            }
         },
         "PowerLimit" : {
            "LimitException" : "HardPowerOff",
            "CorrectionInMs" : 1000,
            "LimitInWatts" : 500
         },
         "PhysicalContext" : "Intake"
      }
   ],
   "@odata.type" : "#Power.v1_5_1.Power",
   "@odata.context" : "/redfish/v1/$metadata#Power.Power",
   "@odata.etag" : "W/\"1583182793\"",
   "Voltages@odata.count" : 28,
   "PowerControl@odata.count" : 1,
   "PowerSupplies@odata.count" : 2,
   "Id" : "Power",
   "Actions" : {
      "Oem" : {
         "#PowerLimitTrigger" : {
            "@Redfish.ActionInfo" : "/redfish/v1/Chassis/Self/Power/LimitTrigger",
            "target" : "/redfish/v1/Chassis/Self/Power/Actions/LimitTrigger"
         }
      }
   },
   "Voltages" : [
      {
         "MaxReadingRange" : 16.575,
         "RelatedItem@odata.count" : 2,
         "MemberId" : "VoltageSensor64",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P_12V",
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 10.27,
         "MinReadingRange" : 0,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/0",
         "UpperThresholdCritical" : 13.65,
         "UpperThresholdNonCritical" : 13.13,
         "LowerThresholdNonCritical" : 10.79,
         "SensorNumber" : 64,
         "ReadingVolts" : 12.22,
         "PhysicalContext" : "VoltageRegulator"
      },
      {
         "UpperThresholdNonCritical" : 5.4741,
         "UpperThresholdCritical" : 5.6797,
         "LowerThresholdNonCritical" : 4.4975,
         "SensorNumber" : 65,
         "ReadingVolts" : 5.14,
         "PhysicalContext" : "VoltageRegulator",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 6.5535,
         "MemberId" : "VoltageSensor65",
         "Name" : "P_5V",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "MinReadingRange" : 0,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 4.2919,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/1"
      },
      {
         "MemberId" : "VoltageSensor66",
         "Name" : "P_3V3",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "MaxReadingRange" : 4.4115,
         "RelatedItem@odata.count" : 2,
         "MinReadingRange" : 0,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 2.8199,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/2",
         "LowerThresholdNonCritical" : 2.9583,
         "UpperThresholdCritical" : 3.7541,
         "UpperThresholdNonCritical" : 3.6157,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 3.3562,
         "SensorNumber" : 66
      },
      {
         "UpperThresholdNonCritical" : 5.4741,
         "UpperThresholdCritical" : 5.6797,
         "LowerThresholdNonCritical" : 4.4975,
         "SensorNumber" : 67,
         "ReadingVolts" : 5.1657,
         "PhysicalContext" : "VoltageRegulator",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 6.5535,
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P_5V_STBY",
         "MemberId" : "VoltageSensor67",
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/3",
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 4.2919,
         "MinReadingRange" : 0
      },
      {
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 68,
         "ReadingVolts" : 0.651,
         "LowerThresholdNonCritical" : 0.448,
         "UpperThresholdCritical" : 1.456,
         "UpperThresholdNonCritical" : 1.4,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/4",
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 0.399,
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P0_VDDCR_SOC",
         "MemberId" : "VoltageSensor68",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785
      },
      {
         "LowerThresholdNonCritical" : 2.6924,
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 69,
         "ReadingVolts" : 3.074,
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P_VBAT",
         "MemberId" : "VoltageSensor69",
         "MaxReadingRange" : 5.406,
         "RelatedItem@odata.count" : 2,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/5",
         "LowerThresholdCritical" : 2.544,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0
      },
      {
         "LowerThresholdNonCritical" : 0.448,
         "UpperThresholdCritical" : 1.456,
         "UpperThresholdNonCritical" : 1.4,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 0.868,
         "SensorNumber" : 70,
         "MemberId" : "VoltageSensor70",
         "Name" : "P0_VDDCR_CPU",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785,
         "MinReadingRange" : 0,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "LowerThresholdCritical" : 0.399,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/6"
      },
      {
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 0.399,
         "MinReadingRange" : 0,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/7",
         "MemberId" : "VoltageSensor71",
         "Name" : "P1_VDDCR_CPU",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785,
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 71,
         "ReadingVolts" : 0.868,
         "LowerThresholdNonCritical" : 0.448,
         "UpperThresholdCritical" : 1.456,
         "UpperThresholdNonCritical" : 1.4
      },
      {
         "UpperThresholdNonCritical" : 1.323,
         "UpperThresholdCritical" : 1.358,
         "LowerThresholdNonCritical" : 1.071,
         "ReadingVolts" : 1.239,
         "SensorNumber" : 72,
         "PhysicalContext" : "VoltageRegulator",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785,
         "Name" : "P0_VDDIO_ABCD",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "MemberId" : "VoltageSensor72",
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/8",
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "LowerThresholdCritical" : 1.022,
         "MinReadingRange" : 0
      },
      {
         "MinReadingRange" : 0,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 1.022,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/9",
         "MemberId" : "VoltageSensor73",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P0_VDDIO_EFGH",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785,
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 73,
         "ReadingVolts" : 1.246,
         "LowerThresholdNonCritical" : 1.071,
         "UpperThresholdNonCritical" : 1.323,
         "UpperThresholdCritical" : 1.358
      },
      {
         "LowerThresholdNonCritical" : 1.071,
         "UpperThresholdCritical" : 1.358,
         "UpperThresholdNonCritical" : 1.323,
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 74,
         "ReadingVolts" : 1.232,
         "MemberId" : "VoltageSensor74",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P1_VDDIO_ABCD",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 1.022,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/10"
      },
      {
         "MemberId" : "VoltageSensor75",
         "Name" : "P1_VDDIO_EFGH",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 1.022,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/11",
         "LowerThresholdNonCritical" : 1.071,
         "UpperThresholdCritical" : 1.358,
         "UpperThresholdNonCritical" : 1.323,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 1.239,
         "SensorNumber" : 75
      },
      {
         "ReadingVolts" : 0.637,
         "SensorNumber" : 76,
         "PhysicalContext" : "VoltageRegulator",
         "UpperThresholdNonCritical" : 1.4,
         "UpperThresholdCritical" : 1.456,
         "LowerThresholdNonCritical" : 0.448,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/12",
         "MinReadingRange" : 0,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "LowerThresholdCritical" : 0.399,
         "MaxReadingRange" : 1.785,
         "RelatedItem@odata.count" : 2,
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P1_VDDCR_SOC",
         "MemberId" : "VoltageSensor76"
      },
      {
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 77,
         "ReadingVolts" : 1.8228,
         "LowerThresholdNonCritical" : 1.6268,
         "UpperThresholdCritical" : 2.058,
         "UpperThresholdNonCritical" : 1.9796,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/13",
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 1.5484,
         "Name" : "P0_VDD_18",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "MemberId" : "VoltageSensor77",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 2.499
      },
      {
         "MemberId" : "VoltageSensor78",
         "Name" : "P_1V0_AUX_LAN",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 1.785,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "LowerThresholdCritical" : 0.847,
         "MinReadingRange" : 0,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/14",
         "LowerThresholdNonCritical" : 0.896,
         "UpperThresholdNonCritical" : 1.106,
         "UpperThresholdCritical" : 1.155,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 0.994,
         "SensorNumber" : 78
      },
      {
         "UpperThresholdCritical" : 2.058,
         "UpperThresholdNonCritical" : 1.9796,
         "LowerThresholdNonCritical" : 1.6268,
         "ReadingVolts" : 1.8326,
         "SensorNumber" : 79,
         "PhysicalContext" : "VoltageRegulator",
         "MaxReadingRange" : 2.499,
         "RelatedItem@odata.count" : 2,
         "MemberId" : "VoltageSensor79",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "P1_VDD_18",
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "LowerThresholdCritical" : 1.5484,
         "MinReadingRange" : 0,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/15"
      },
      {
         "ReadingVolts" : 12.125,
         "SensorNumber" : 80,
         "PhysicalContext" : "VoltageRegulator",
         "UpperThresholdNonCritical" : 13.25,
         "UpperThresholdCritical" : 13.625,
         "LowerThresholdNonCritical" : 10.75,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 10.375,
         "MinReadingRange" : 0,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/16",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 31.875,
         "MemberId" : "VoltageSensor80",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "VR_P0_VIN"
      },
      {
         "Name" : "VR_P0_VOUT",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "MemberId" : "VoltageSensor81",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 4.08,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/17",
         "LowerThresholdCritical" : 0.4,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0,
         "LowerThresholdNonCritical" : 0.448,
         "UpperThresholdCritical" : 1.456,
         "UpperThresholdNonCritical" : 1.408,
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 81,
         "ReadingVolts" : 1.184
      },
      {
         "MemberId" : "VoltageSensor82",
         "Name" : "VR_P1_VIN",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 31.875,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 10.375,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/18",
         "LowerThresholdNonCritical" : 10.75,
         "UpperThresholdCritical" : 13.625,
         "UpperThresholdNonCritical" : 13.25,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 12,
         "SensorNumber" : 82
      },
      {
         "MemberId" : "VoltageSensor83",
         "Name" : "VR_P1_VOUT",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "MaxReadingRange" : 4.08,
         "RelatedItem@odata.count" : 2,
         "MinReadingRange" : 0,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "LowerThresholdCritical" : 0.4,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/19",
         "LowerThresholdNonCritical" : 0.448,
         "UpperThresholdCritical" : 1.456,
         "UpperThresholdNonCritical" : 1.408,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 1.216,
         "SensorNumber" : 83
      },
      {
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 12.125,
         "SensorNumber" : 84,
         "LowerThresholdNonCritical" : 10.75,
         "UpperThresholdCritical" : 13.625,
         "UpperThresholdNonCritical" : 13.25,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 10.375,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/20",
         "MemberId" : "VoltageSensor84",
         "Name" : "VR_DIMMG0_VIN",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 31.875
      },
      {
         "UpperThresholdNonCritical" : 1.328,
         "UpperThresholdCritical" : 1.392,
         "LowerThresholdNonCritical" : 1.088,
         "SensorNumber" : 85,
         "ReadingVolts" : 1.248,
         "PhysicalContext" : "VoltageRegulator",
         "MaxReadingRange" : 4.08,
         "RelatedItem@odata.count" : 2,
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "VR_DIMMG0_VOUT",
         "MemberId" : "VoltageSensor85",
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/21",
         "MinReadingRange" : 0,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "LowerThresholdCritical" : 1.024
      },
      {
         "MemberId" : "VoltageSensor86",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "VR_DIMMG1_VIN",
         "MaxReadingRange" : 31.875,
         "RelatedItem@odata.count" : 2,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 10.375,
         "MinReadingRange" : 0,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/22",
         "LowerThresholdNonCritical" : 10.75,
         "UpperThresholdCritical" : 13.625,
         "UpperThresholdNonCritical" : 13.25,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 12,
         "SensorNumber" : 86
      },
      {
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 1.024,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/23",
         "MaxReadingRange" : 4.08,
         "RelatedItem@odata.count" : 2,
         "MemberId" : "VoltageSensor87",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "VR_DIMMG1_VOUT",
         "ReadingVolts" : 1.248,
         "SensorNumber" : 87,
         "PhysicalContext" : "VoltageRegulator",
         "UpperThresholdCritical" : 1.392,
         "UpperThresholdNonCritical" : 1.328,
         "LowerThresholdNonCritical" : 1.088
      },
      {
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "VR_DIMMG2_VIN",
         "MemberId" : "VoltageSensor88",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 31.875,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/24",
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 10.375,
         "LowerThresholdNonCritical" : 10.75,
         "UpperThresholdNonCritical" : 13.25,
         "UpperThresholdCritical" : 13.625,
         "PhysicalContext" : "VoltageRegulator",
         "SensorNumber" : 88,
         "ReadingVolts" : 12.125
      },
      {
         "MinReadingRange" : 0,
         "Status" : {
            "Health" : "OK",
            "State" : "Enabled"
         },
         "LowerThresholdCritical" : 1.024,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/25",
         "RelatedItem@odata.count" : 2,
         "MaxReadingRange" : 4.08,
         "MemberId" : "VoltageSensor89",
         "Name" : "VR_DIMMG2_VOUT",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "SensorNumber" : 89,
         "ReadingVolts" : 1.248,
         "PhysicalContext" : "VoltageRegulator",
         "UpperThresholdCritical" : 1.392,
         "UpperThresholdNonCritical" : 1.328,
         "LowerThresholdNonCritical" : 1.088
      },
      {
         "ReadingVolts" : 12.125,
         "SensorNumber" : 90,
         "PhysicalContext" : "VoltageRegulator",
         "UpperThresholdNonCritical" : 13.25,
         "UpperThresholdCritical" : 13.625,
         "LowerThresholdNonCritical" : 10.75,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0,
         "LowerThresholdCritical" : 10.375,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/26",
         "MaxReadingRange" : 31.875,
         "RelatedItem@odata.count" : 2,
         "MemberId" : "VoltageSensor90",
         "Name" : "VR_DIMMG3_VIN",
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ]
      },
      {
         "RelatedItem" : [
            {
               "@odata.id" : "/redfish/v1/Chassis/Self"
            },
            {
               "@odata.id" : "/redfish/v1/Systems/Self"
            }
         ],
         "Name" : "VR_DIMMG3_VOUT",
         "MemberId" : "VoltageSensor91",
         "MaxReadingRange" : 4.08,
         "RelatedItem@odata.count" : 2,
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/Voltages/27",
         "LowerThresholdCritical" : 1.024,
         "Status" : {
            "State" : "Enabled",
            "Health" : "OK"
         },
         "MinReadingRange" : 0,
         "LowerThresholdNonCritical" : 1.088,
         "UpperThresholdCritical" : 1.392,
         "UpperThresholdNonCritical" : 1.328,
         "PhysicalContext" : "VoltageRegulator",
         "ReadingVolts" : 1.248,
         "SensorNumber" : 91
      }
   ],
   "Description" : "Power sensor readings",
   "PowerSupplies" : [
      {
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/PowerSupplies/0",
         "Name" : "nil",
         "MemberId" : "0"
      },
      {
         "@odata.id" : "/redfish/v1/Chassis/Self/Power#/PowerSupplies/1",
         "Name" : "nil",
         "MemberId" : "1"
      }
   ]
}
`

const testPathGBT_chassis_self_network_adapters = "/redfish/v1/Chassis/Self/NetworkAdapters"

const testPayloadGBT_chassis_self_network_adapters = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapterCollection.NetworkAdapterCollection",
  "@odata.etag": "W/\"1604711998\"",
  "@odata.id": "/redfish/v1/Chassis/Self/NetworkAdapters",
  "@odata.type": "#NetworkAdapterCollection.NetworkAdapterCollection",
  "Description": "The Collection of Network Adapters",
  "Members": [],
  "Members@odata.count": 0,
  "Name": "NetworkAdapter Collection"
}`

const testPathGBT_systems_self_ethernet_interfaces = "/redfish/v1/Systems/Self/EthernetInterfaces"

const testPayloadGBT_systems_self_ethernet_interfaces = `
{
   "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces",
   "@odata.type" : "#EthernetInterfaceCollection.EthernetInterfaceCollection",
   "Members@odata.count" : 2,
   "Name" : "Ethernet Interface Collection",
   "@odata.context" : "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
   "@odata.etag" : "W/\"1584739441\"",
   "Description" : "Collection of ethernet interfaces for this system",
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces/1"
      },
      {
         "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces/2"
      }
   ]
}
`

const testPathGBT_systems_self_ethernet_interfaces_1 = "/redfish/v1/Systems/Self/EthernetInterfaces/1"

const testPayloadGBT_systems_self_ethernet_interfaces_1 = `
{
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "MACAddress" : "B4:2E:99:B5:D7:11",
   "Name" : "Lan1",
   "VLANs" : {
      "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces/1/VLANs"
   },
   "Description" : "Ethernet Interface Lan1",
   "@odata.type" : "#EthernetInterface.v1_4_1.EthernetInterface",
   "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces/1",
   "@odata.context" : "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
   "Status" : {
      "Health" : "OK",
      "State" : "Enabled"
   },
   "Id" : "1",
   "@odata.etag" : "W/\"1584739441\""
}
`

const testPathGBT_systems_self_ethernet_interfaces_2 = "/redfish/v1/Systems/Self/EthernetInterfaces/2"

const testPayloadGBT_systems_self_ethernet_interfaces_2 = `
{
   "MACAddress" : "B4:2E:99:B5:D7:12",
   "@odata.context" : "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
   "VLANs" : {
      "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces/2/VLANs"
   },
   "@odata.type" : "#EthernetInterface.v1_4_1.EthernetInterface",
   "@odata.id" : "/redfish/v1/Systems/Self/EthernetInterfaces/2",
   "Status" : {
      "State" : "Enabled",
      "Health" : "OK"
   },
   "Id" : "2",
   "Description" : "Ethernet Interface Lan2",
   "Name" : "Lan2",
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1584739441\""
}
`

const testPathGBT_systems_self_processors = "/redfish/v1/Systems/Self/Processors"

const testPayloadGBT_systems_self_processors = `
{
   "@odata.type" : "#ProcessorCollection.ProcessorCollection",
   "Description" : "Collection of processors",
   "Members@odata.count" : 2,
   "@odata.etag" : "W/\"1583182793\"",
   "@odata.id" : "/redfish/v1/Systems/Self/Processors",
   "@odata.context" : "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Self/Processors/1"
      },
      {
         "@odata.id" : "/redfish/v1/Systems/Self/Processors/2"
      }
   ],
   "Name" : "Processors Collection"
}
`

const testPathGBT_systems_self_processors_1 = "/redfish/v1/Systems/Self/Processors/1"

const testPayloadGBT_systems_self_processors_1 = `
{
   "TotalThreads" : 64,
   "Id" : "1",
   "@odata.type" : "#Processor.v1_3_1.Processor",
   "Socket" : "P0",
   "ProcessorArchitecture" : "x86",
   "Description" : "Processor Instance 1",
   "Oem" : {
      "GBTProcessorOemProperty" : {
         "@odata.type" : "#GBTProcessorOemProperty.v1_0_0.GBTProcessorOemProperty",
         "Processor Serial Number" : "2B493ADB3B3C07D"
      }
   },
   "InstructionSet" : "x86-64",
   "SubProcessors" : {
      "@odata.id" : "/redfish/v1/Systems/Self/Processors/1/SubProcessors"
   },
   "ProcessorId" : {
      "EffectiveFamily" : "AMD Zen Processor Family",
      "Step" : "0x0",
      "VendorId" : "AuthenticAMD",
      "IdentificationRegisters" : "178bfbff00830f10",
      "EffectiveModel" : "0x31"
   },
   "ProcessorType" : "CPU",
   "@odata.id" : "/redfish/v1/Systems/Self/Processors/1",
   "@odata.context" : "/redfish/v1/$metadata#Processor.Processor",
   "TotalCores" : 32,
   "Status" : {
      "State" : "Enabled",
      "Health" : "OK"
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "Name" : "Processor 1",
   "MaxSpeedMHz" : 3400,
   "@odata.etag" : "W/\"1583182793\"",
   "Manufacturer" : "Advanced Micro Devices, Inc.",
   "Model" : "AMD EPYC 7542 32-Core Processor                "
}
`

const testPathGBT_systems_self_processors_2 = "/redfish/v1/Systems/Self/Processors/2"

const testPayloadGBT_systems_self_processors_2 = `
{
   "Oem" : {
      "GBTProcessorOemProperty" : {
         "@odata.type" : "#GBTProcessorOemProperty.v1_0_0.GBTProcessorOemProperty",
         "Processor Serial Number" : "2B493ADB3B3C082"
      }
   },
   "Name" : "Processor 2",
   "TotalThreads" : 64,
   "InstructionSet" : "x86-64",
   "SubProcessors" : {
      "@odata.id" : "/redfish/v1/Systems/Self/Processors/2/SubProcessors"
   },
   "ProcessorArchitecture" : "x86",
   "MaxSpeedMHz" : 3400,
   "Manufacturer" : "Advanced Micro Devices, Inc.",
   "@odata.context" : "/redfish/v1/$metadata#Processor.Processor",
   "@odata.id" : "/redfish/v1/Systems/Self/Processors/2",
   "Description" : "Processor Instance 2",
   "ProcessorId" : {
      "VendorId" : "AuthenticAMD",
      "EffectiveFamily" : "AMD Zen Processor Family",
      "Step" : "0x0",
      "EffectiveModel" : "0x31",
      "IdentificationRegisters" : "178bfbff00830f10"
   },
   "TotalCores" : 32,
   "ProcessorType" : "CPU",
   "Model" : "AMD EPYC 7542 32-Core Processor                ",
   "Id" : "2",
   "Socket" : "P1",
   "@odata.type" : "#Processor.v1_3_1.Processor",
   "@odata.etag" : "W/\"1583182793\"",
   "Status" : {
      "Health" : "OK",
      "State" : "Enabled"
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   }
}
`

const testPathGBT_systems_self_memory = "/redfish/v1/Systems/Self/Memory"

const testPayloadGBT_systems_self_memory = `
{
   "Members@odata.count" : 2,
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Self/Memory/1"
      },
      {
         "@odata.id" : "/redfish/v1/Systems/Self/Memory/2"
      }
   ],
   "@odata.id" : "/redfish/v1/Systems/Self/Memory",
   "Description" : "Collection of Memories for this system",
   "@odata.context" : "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
   "@odata.type" : "#MemoryCollection.MemoryCollection",
   "Name" : "Memory Collection",
   "@odata.etag" : "W/\"1583182793\""
}
`

const testPathGBT_systems_self_memory_1 = "/redfish/v1/Systems/Self/Memory/1"

const testPayloadGBT_systems_self_memory_1 = `
{
   "Regions" : [
      {
         "RegionId" : "4c",
         "SizeMiB" : 1953,
         "OffsetMiB" : 0
      }
   ],
   "@odata.type" : "#Memory.v1_6_0.Memory",
   "Description" : "Memory Instance 1",
   "OperatingSpeedMhz" : 3200,
   "PartNumber" : "M393A2K43DB2-CWE    ",
   "Manufacturer" : "Samsung",
   "@odata.context" : "/redfish/v1/$metadata#Memory.Memory",
   "DataWidthBits" : 40,
   "Actions" : {
      "Oem" : {
         "#AmiBios.ChangeState" : {
            "State@Redfish.AllowableValues" : [
               "Enabled",
               "Disabled"
            ],
            "target" : "/redfish/v1/Systems/Self/Memory/1/Actions/AmiBios.ChangeState"
         }
      }
   },
   "DeviceLocator" : "DIMM_P0_A0",
   "Id" : "1",
   "BusWidthBits" : 48,
   "Name" : "Memory 1",
   "AllowedSpeedsMHz" : [
      3200
   ],
   "ErrorCorrection" : "MultiBitECC",
   "SerialNumber" : "039A9B54",
   "MemoryType" : "DRAM",
   "@odata.id" : "/redfish/v1/Systems/Self/Memory/1",
   "Status" : {
      "State" : "Enabled",
      "Health" : "OK"
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1583182889\"",
   "MemoryDeviceType" : "DDR4",
   "CapacityMiB" : 15625
}
`

const testPathGBT_systems_self_memory_2 = "/redfish/v1/Systems/Self/Memory/2"

const testPayloadGBT_systems_self_memory_2 = `
{
   "Regions" : [
      {
         "OffsetMiB" : 1953,
         "SizeMiB" : 3906,
         "RegionId" : "4f"
      }
   ],
   "MemoryDeviceType" : "DDR4",
   "@odata.etag" : "W/\"1583182889\"",
   "Name" : "Memory 2",
   "@odata.id" : "/redfish/v1/Systems/Self/Memory/2",
   "BusWidthBits" : 48,
   "Status" : {
      "Health" : "OK",
      "State" : "Enabled"
   },
   "ErrorCorrection" : "MultiBitECC",
   "Manufacturer" : "Samsung",
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.context" : "/redfish/v1/$metadata#Memory.Memory",
   "Actions" : {
      "Oem" : {
         "#AmiBios.ChangeState" : {
            "target" : "/redfish/v1/Systems/Self/Memory/2/Actions/AmiBios.ChangeState",
            "State@Redfish.AllowableValues" : [
               "Enabled",
               "Disabled"
            ]
         }
      }
   },
   "PartNumber" : "M393A2K43DB2-CWE    ",
   "OperatingSpeedMhz" : 3200,
   "DeviceLocator" : "DIMM_P0_B0",
   "CapacityMiB" : 15625,
   "AllowedSpeedsMHz" : [
      3200
   ],
   "Id" : "2",
   "Description" : "Memory Instance 2",
   "SerialNumber" : "039A9A80",
   "DataWidthBits" : 40,
   "MemoryType" : "DRAM",
   "@odata.type" : "#Memory.v1_6_0.Memory"
}
`

const testPathGBT_systems_self_storage = "/redfish/v1/Systems/Self/Storage"

const testPayloadGBT_systems_self_storage = `
{
   "@odata.id" : "/redfish/v1/Systems/Self/Storage",
   "@odata.type" : "#StorageCollection.StorageCollection",
   "Description" : "Collection of Storage resource instances",
   "Members" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Self/Storage/1"
      }
   ],
   "@odata.etag" : "W/\"1584741186\"",
   "Name" : "Storage Collection",
   "@odata.context" : "/redfish/v1/$metadata#StorageCollection.StorageCollection",
   "Members@odata.count" : 1
}
`

const testPathGBT_systems_self_storage_1 = "/redfish/v1/Systems/Self/Storage/1"

const testPayloadGBT_systems_self_storage_1 = `
{
   "Name" : "Storage",
   "@odata.id" : "/redfish/v1/Systems/Self/Storage/1",
   "@odata.context" : "/redfish/v1/$metadata#Storage.Storage",
   "Description" : "This resource shall be used to represent resources that represent a storage subsystem in the Redfish specification.",
   "Actions" : {
      "Oem" : {
         "#CreateLogicDevice" : {
            "@Redfish.ActionInfo" : "/redfish/v1/Systems/Self/Storage/RaidConfig/CreateLogicDevice",
            "target" : "/redfish/v1/Systems/Self/Storage/RaidConfig/Actions/CreateLogicDevice"
         },
         "#DeleteLogicDevice" : {
            "@Redfish.ActionInfo" : "/redfish/v1/Systems/Self/Storage/RaidConfig/DeleteLogicDevice",
            "target" : "/redfish/v1/Systems/Self/Storage/RaidConfig/Actions/DeleteLogicDevice"
         }
      }
   },
   "@odata.etag" : "W/\"1584741186\"",
   "@odata.type" : "#Storage.v1_5_0.Storage",
   "Id" : "1"
}
`

//////////////////////////////////////////////////////////////////////////////
//                         Intel s2600* - Mock Client
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncIntel1() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathIntel_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadIntel_redfish_v1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelAccountService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelAccountService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSessionService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSessionService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelEventService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelEventService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_PwrSupply1FRU:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_PwrSupply1FRU)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_PwrSupply2FRU:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_PwrSupply2FRU)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_Baseboard:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_Baseboard)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_Baseboard_Power:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_Baseboard_Power)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_RackMount:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_RackMount)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_HSBackplane:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_HSBackplane)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_PCIeRiser1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_PCIeRiser1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelChassis_PCIeRiser2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelChassis_PCIeRiser2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelManagers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelManagers)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelManagers_BMC:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelManagers_BMC)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelManagers_BMC_NICs:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelManagers_BMC_NICs)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelManagers_BMC_NICs_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelManagers_BMC_NICs_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelManagers_BMC_NICs_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelManagers_BMC_NICs_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelManagers_BMC_NICs_3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelManagers_BMC_NICs_3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Processors:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Processors)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_4:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_4)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_5:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_5)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_6:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_6)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_7:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_7)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_1_Drives_8:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_8)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_4:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_4)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_5:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_5)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_6:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_6)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_7:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_7)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Storage_2_Drives_8:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_8)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Processors_CPU1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Processors_CPU1)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Processors_CPU2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Processors_CPU2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Memory:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Memory)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Memory_Memory1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Memory_Memory1)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Memory_Memory2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Memory_Memory2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Memory_Memory3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Memory_Memory3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Memory_Memory4:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Memory_Memory4)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Memory_Memory5:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Memory_Memory5)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathIntelSystems_QSBP74304715_Memory_Memory6:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadIntelSystems_QSBP74304715_Memory_Memory6)),
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString("")),

				Header: make(http.Header),
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////
//                         Intel s2600* - Payloads
//////////////////////////////////////////////////////////////////////////////

const testPathIntel_redfish_v1 = "/redfish/v1"

const testPayloadIntel_redfish_v1 = `
{
        "@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
        "@odata.id": "/redfish/v1",
        "@odata.type": "#ServiceRoot.v1_1_1.ServiceRoot",
        "Id": "RootService",
        "Name": "Root Service",
        "RedfishVersion": "1.1.0",
        "UUID": "7a24e110-043c-4497-b114-17d5cde4d29f",
        "Systems": {
                "@odata.id": "/redfish/v1/Systems"
        },
        "Chassis": {
                "@odata.id": "/redfish/v1/Chassis"
        },
        "Managers": {
                "@odata.id": "/redfish/v1/Managers"
        },
        "SessionService": {
                "@odata.id": "/redfish/v1/SessionService"
        },
        "AccountService": {
                "@odata.id": "/redfish/v1/AccountService"
        },
        "UpdateService": {
                "@odata.id": "/redfish/v1/UpdateService"
        },
        "EventService": {
                "@odata.id": "/redfish/v1/EventService"
        },
        "Registries": {
                "@odata.id": "/redfish/v1/Registries"
        },
        "Links": {
                "Sessions": {
                        "@odata.id": "/redfish/v1/SessionService/Sessions"
                }
        }
}
`

const testPathIntelAccountService = "/redfish/v1/AccountService"

const testPayloadIntelAccountService = `
{
        "@odata.context": "/redfish/v1/$metadata#AccountService.AccountService",
        "@odata.id": "/redfish/v1/AccountService",
        "@odata.type": "#AccountService.v1_1_0.AccountService",
        "Id": "AccountService",
        "Name": "Account Service",
        "Description": "BMC User Accounts",
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "ServiceEnabled": true,
        "MinPasswordLength": 1,
        "MaxPasswordLength": 20,
        "Accounts": {
                "@odata.id": "/redfish/v1/AccountService/Accounts"
        },
        "Roles": {
                "@odata.id": "/redfish/v1/AccountService/Roles"
        }
}
`

const testPathIntelSessionService = "/redfish/v1/SessionService"

const testPayloadIntelSessionService = `
{
        "@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
        "@odata.id": "/redfish/v1/SessionService",
        "@odata.type": "#SessionService.v1_1_1.SessionService",
        "Id": "SessionService",
        "Name": "SessionService",
        "Description": "SessionService",
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "ServiceEnabled": true,
        "SessionTimeout": 1800,
        "Sessions": {
                "@odata.id": "/redfish/v1/SessionService/Sessions"
        }
}
`

const testPathIntelEventService = "/redfish/v1/EventService"

const testPayloadIntelEventService = `
{
        "@odata.context": "/redfish/v1/$metadata#EventService.EventService",
        "@odata.id": "/redfish/v1/EventService",
        "@odata.type": "#EventService.v1_0_3.EventService",
        "Id": "EventService",
        "Name": "Event Service",
        "ServiceEnabled": true,
        "DeliveryRetryAttempts": 3,
        "EventTypesForSubscription": [
                "Alert"
        ],
        "Subscriptions": {
                "@odata.id": "/redfish/v1/EventService/Subscriptions"
        },
        "Actions": {
                "#EventService.SubmitTestEvent": {
                        "target": "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent",
                        "EventType@Redfish.AllowableValues": [
                                "Alert"
                        ]
                }
        }
}
`

const testPathIntelChassis = "/redfish/v1/Chassis"

// Note last link here is INTENTIONALLY wrong - Should recover
const testPayloadIntelChassis = `
{
        "@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
        "@odata.id": "/redfish/v1/Chassis",
        "@odata.type": "#ChassisCollection.ChassisCollection",
        "Name": "Chassis Collection",
        "Members@odata.count": 7,
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Chassis/Baseboard"
                },
                {
                        "@odata.id": "/redfish/v1/Chassis/RackMount"
                },
                {
                        "@odata.id": "/redfish/v1/Chassis/HSBackplane"
                },
                {
                        "@odata.id": "/redfish/v1/Chassis/PCIeRiser1"
                },
                {
                        "@odata.id": "/redfish/v1/Chassis/PCIeRiser2"
                },
                {
                        "@odata.id": "/redfish/v1/Chassis/PwrSupply1FRU"
                },
                {
                        "@odata.id": "/redfish/v1/Chassis/PwrSupply2FRU"
                },
                {
                        "@odata.id": "/redfish/v1/Chassis/DeadLink404Test"
                }
        ]
}
`

const testPathIntelChassis_PwrSupply1FRU = "/redfish/v1/Chassis/PwrSupply1FRU"

const testPayloadIntelChassis_PwrSupply1FRU = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/PwrSupply1FRU",
        "@odata.type": "#Chassis.v1_4_0.Chassis",
        "Id": "PwrSupply1FRU",
        "Name": "Computer System Card",
        "ChassisType": "Card",
        "Manufacturer": "SOLUM CO., LTD.                  ",
        "Model": "PSSF222201A",
        "SerialNumber": "CNS2221A4AH9R0619",
        "PartNumber": "H66158-010",
        "PowerState": "On",
        "Links": {
                "ComputerSystems@odata.count": 1,
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ],
                "ContainedBy@odata.count": 1,
                "ContainedBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ]
        }
}
`

const testPathIntelChassis_PwrSupply2FRU = "/redfish/v1/Chassis/PwrSupply2FRU"

const testPayloadIntelChassis_PwrSupply2FRU = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/PwrSupply2FRU",
        "@odata.type": "#Chassis.v1_4_0.Chassis",
        "Id": "PwrSupply2FRU",
        "Name": "Computer System Card",
        "ChassisType": "Card",
        "Manufacturer": "SOLUM CO., LTD.                  ",
        "Model": "PSSF222201A",
        "SerialNumber": "CNS2221A4AHAP0374",
        "PartNumber": "H66158-010",
        "PowerState": "On",
        "Links": {
                "ComputerSystems@odata.count": 1,
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ],
                "ContainedBy@odata.count": 1,
                "ContainedBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ]
        }
}
`

const testPathIntelChassis_Baseboard = "/redfish/v1/Chassis/Baseboard"

const testPayloadIntelChassis_Baseboard = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/Baseboard",
        "@odata.type": "#Chassis.v1_4_0.Chassis",
        "Id": "Baseboard",
        "Name": "Computer System Card",
        "ChassisType": "Card",
        "Manufacturer": "Intel Corporation",
        "Model": "S2600BPB",
        "SerialNumber": "QSBP74304715",
        "PartNumber": "H87926-550",
        "PowerState": "On",
        "Thermal": {
                "@odata.id": "/redfish/v1/Chassis/Baseboard/Thermal"
        },
        "Power": {
                "@odata.id": "/redfish/v1/Chassis/Baseboard/Power"
        },
        "Links": {
                "ComputerSystems@odata.count": 1,
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ],
                "ContainedBy@odata.count": 1,
                "ContainedBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ]
        }
}
`

const testPathIntelChassis_Baseboard_Power = "/redfish/v1/Chassis/Baseboard/Power"

const testPayloadIntelChassis_Baseboard_Power = `
{
  "@odata.context": "/redfish/v1/$metadata#Power.Power",
  "@odata.id": "/redfish/v1/Chassis/RackMount/Baseboard/Power",
  "@odata.type": "#Power.v1_5_0.Power",
  "Id": "Power",
  "Name": "Power",
  "PowerControl": [
    {
      "@odata.id": "/redfish/v1/Chassis/RackMount/Baseboard/Power#/PowerControl/0",
      "MemberId": "0",
      "Name": "Server Power Control",
      "PowerConsumedWatts": 124,
      "PowerMetrics": {
        "IntervalInMin": 7120,
        "MinConsumedWatts": 1,
        "MaxConsumedWatts": 157,
        "AverageConsumedWatts": 124
      },
      "RelatedItem": [
        {
          "@odata.id": "/redfish/v1/Systems/QSBP74304715"
        },
        {
          "@odata.id": "/redfish/v1/Chassis/RackMount"
        }
      ]
    }
  ],
  "Voltages": [
    {
      "@odata.id": "/redfish/v1/Chassis/RackMount/Baseboard/Power#/Voltages/0",
      "MemberId": "0",
      "Name": "BB +12.0V",
      "SensorNumber": 208,
      "Status": {
        "State": "Enabled",
        "Health": "OK",
        "HealthRollup": "OK"
      },
      "ReadingVolts": 11.991999626159668,
      "UpperThresholdNonCritical": 13.256999969482422,
      "UpperThresholdCritical": 13.642000198364258,
      "LowerThresholdNonCritical": 11.001999855041504,
      "LowerThresholdCritical": 10.67199993133545,
      "MinReadingRange": -0.21799999475479126,
      "MaxReadingRange": 13.807000160217285,
      "PhysicalContext": "SystemBoard",
      "RelatedItem": [
        {
          "@odata.id": "/redfish/v1/Systems/QSBP74304715"
        },
        {
          "@odata.id": "/redfish/v1/Chassis/RackMount"
        }
      ]
    },
    {
      "@odata.id": "/redfish/v1/Chassis/RackMount/Baseboard/Power#/Voltages/1",
      "MemberId": "1",
      "Name": "BB +3.3V Vbat",
      "SensorNumber": 222,
      "Status": {
        "State": "Enabled",
        "Health": "OK",
        "HealthRollup": "OK"
      },
      "ReadingVolts": 2.8405001163482666,
      "LowerThresholdNonCritical": 2.450500011444092,
      "LowerThresholdCritical": 2.125499963760376,
      "MinReadingRange": 0.006500000134110451,
      "MaxReadingRange": 3.321500062942505,
      "PhysicalContext": "SystemBoard",
      "RelatedItem": [
        {
          "@odata.id": "/redfish/v1/Systems/QSBP74304715"
        },
        {
          "@odata.id": "/redfish/v1/Chassis/RackMount"
        }
      ]
    }
  ],
  "PowerSupplies": [
    {
      "@odata.id": "/redfish/v1/Chassis/RackMount/Baseboard/Power#/PowerSupplies/0",
      "MemberId": "0",
      "Name": "Power Supply Bay",
      "Status": {
        "State": "Enabled",
        "Health": "OK",
        "HealthRollup": "OK"
      },
      "LineInputVoltage": 377,
      "Model": "PSSF222201A",
      "Manufacturer": "SOLUM CO., LTD.                  ",
      "FirmwareVersion": "04A",
      "SerialNumber": "CNS2221A4AH9R0619",
      "PartNumber": "H66158-010",
      "RelatedItem": [
        {
          "@odata.id": "/redfish/v1/Chassis/Baseboard/Power"
        }
      ],
      "Redundancy": [
        {
          "@odata.id": "/redfish/v1/Chassis/Baseboard/Power#/Redundancy/0"
        }
      ]
    },
    {
      "@odata.id": "/redfish/v1/Chassis/Baseboard/Power#/PowerSupplies/1",
      "MemberId": "1",
      "Name": "Power Supply Bay",
      "Status": {
        "State": "Enabled",
        "Health": "OK",
        "HealthRollup": "OK"
      },
      "LineInputVoltage": 13,
      "Model": "PSSF222201A",
      "Manufacturer": "SOLUM CO., LTD.                  ",
      "FirmwareVersion": "04A",
      "SerialNumber": "CNS2221A4AHAP0374",
      "PartNumber": "H66158-010",
      "RelatedItem": [
        {
          "@odata.id": "/redfish/v1/Chassis/Baseboard/Power"
        }
      ],
      "Redundancy": [
        {
          "@odata.id": "/redfish/v1/Chassis/Baseboard/Power#/Redundancy/0"
        }
      ]
    }
  ],
  "Redundancy": [
    {
      "@odata.id": "/redfish/v1/Chassis/Baseboard/Power#/Redundancy/0",
      "MemberId": "0",
      "Name": "Baseboard Power Supply",
      "RedundancySet": [
        {
          "@odata.id": "/redfish/v1/Chassis/Baseboard/Power#/PowerSupplies/0"
        },
        {
          "@odata.id": "/redfish/v1/Chassis/Baseboard/Power#/PowerSupplies/1"
        }
      ],
      "Mode": "N+m",
      "Status": {
        "State": "Disabled",
        "Health": "OK",
        "HealthRollup": "OK"
      },
      "MinNumNeeded": 1,
      "MaxNumSupported": 2
    }
  ]
}
`

const testPathIntelChassis_RackMount = "/redfish/v1/Chassis/RackMount"

const testPayloadIntelChassis_RackMount = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/RackMount",
        "@odata.type": "#Chassis.v1_4_0.Chassis",
        "Id": "RackMount",
        "Name": "Computer System Chassis",
        "Description": "System Chassis",
        "ChassisType": "RackMount",
        "Manufacturer": "Cray Inc.",
        "Model": "S2600BPB",
        "SerialNumber": "QSBP74304715",
        "PartNumber": "..................",
        "PowerState": "On",
        "PhysicalSecurity": {
                "IntrusionSensorNumber": 4,
                "IntrusionSensor": "Normal"
        },
        "IndicatorLED": "Off",
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "Links": {
                "Storage@odata.count": 1,
                "Storage": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1"
                        }
                ],
                "ComputerSystems@odata.count": 1,
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ],
                "Contains@odata.count": 6,
                "Contains": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount/Baseboard"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount/HSBackplane"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount/PCIeRiser1"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount/PCIeRiser2"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount/PwrSupply1FRU"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount/PwrSupply2FRU"
                        }
                ]
        }
}
`

const testPathIntelChassis_HSBackplane = "/redfish/v1/Chassis/HSBackplane"

const testPayloadIntelChassis_HSBackplane = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/HSBackplane",
        "@odata.type": "#Chassis.v1_4_0.Chassis",
        "Id": "HSBackplane",
        "Name": "Computer System Card",
        "ChassisType": "Card",
        "Manufacturer": "Intel Corporation",
        "Model": "FHW12X35HSBP12G",
        "SerialNumber": "BQKP74303187",
        "PartNumber": "G92718-271",
        "PowerState": "On",
        "Links": {
                "ComputerSystems@odata.count": 1,
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "Drives": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD1"
                        },
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD2"
                        },
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD3"
                        },
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD4"
                        },
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD5"
                        },
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD6"
                        },
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD7"
                        },
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD8"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ],
                "ContainedBy@odata.count": 1,
                "ContainedBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ]
        }
}
`

const testPathIntelChassis_PCIeRiser1 = "/redfish/v1/Chassis/PCIeRiser1"

const testPayloadIntelChassis_PCIeRiser1 = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/PCIeRiser1",
        "@odata.type": "#Chassis.v1_4_0.Chassis",
        "Id": "PCIeRiser1",
        "Name": "Computer System Card",
        "ChassisType": "Card",
        "Manufacturer": "Intel Corporation",
        "Model": "AHW1URISER1    ",
        "SerialNumber": "QSBP74206799",
        "PartNumber": "H88058-150",
        "PowerState": "On",
        "Links": {
                "ComputerSystems@odata.count": 1,
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ],
                "ContainedBy@odata.count": 1,
                "ContainedBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ]
        }
}
`

const testPathIntelChassis_PCIeRiser2 = "/redfish/v1/Chassis/PCIeRiser2"

const testPayloadIntelChassis_PCIeRiser2 = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/PCIeRiser2",
        "@odata.type": "#Chassis.v1_4_0.Chassis",
        "Id": "PCIeRiser2",
        "Name": "Computer System Card",
        "ChassisType": "Card",
        "Manufacturer": "Intel Corporation",
        "Model": "AHW1UM2RISER2  ",
        "SerialNumber": "QSBP74210438",
        "PartNumber": "H94856-350",
        "PowerState": "On",
        "Links": {
                "ComputerSystems@odata.count": 1,
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ],
                "ContainedBy@odata.count": 1,
                "ContainedBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ]
        }
}
`

const testPathIntelManagers = "/redfish/v1/Managers"

const testPayloadIntelManagers = `
{
        "@odata.context": "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
        "@odata.id": "/redfish/v1/Managers",
        "@odata.type": "#ManagerCollection.ManagerCollection",
        "Name": "Manager Collection",
        "Members@odata.count": 1,
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Managers/BMC"
                }
        ]
}
`

const testPathIntelManagers_BMC = "/redfish/v1/Managers/BMC"

const testPayloadIntelManagers_BMC = `
{
        "@odata.context": "/redfish/v1/$metadata#Manager.Manager",
        "@odata.id": "/redfish/v1/Managers/BMC",
        "@odata.type": "#Manager.v1_3_0.Manager",
        "Id": "BMC",
        "Name": "Manager",
        "ManagerType": "BMC",
        "Description": "Baseboard Management Controller",
        "UUID": "697fb703-878b-2749-093a-0a6276ea1409",
        "Manufacturer": "Intel",
        "SerialNumber": "12345xyz",
        "Model": "S2600BPB",
        "DateTime": "2018-10-01T22:20:52+00:00",
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "GraphicalConsole": {
                "ServiceEnabled": false,
                "MaxConcurrentSessions": 4,
                "ConnectTypesSupported": [
                        "KVMIP"
                ]
        },
        "SerialConsole": {
                "ServiceEnabled": false,
                "MaxConcurrentSessions": 1,
                "ConnectTypesSupported": [
                        "SSH",
                        "IPMI"
                ]
        },
        "FirmwareVersion": "1.43.660a4315",
        "PowerState": "On",
        "NetworkProtocol": {
                "@odata.id": "/redfish/v1/Managers/BMC/NetworkProtocol"
        },
        "EthernetInterfaces": {
                "@odata.id": "/redfish/v1/Managers/BMC/NICs"
        },
        "LogServices": {
                "@odata.id": "/redfish/v1/Managers/BMC/LogServices"
        },
        "VirtualMedia": {
                "@odata.id": "/redfish/v1/Managers/BMC/VirtualMedia"
        },
        "Links": {
                "ManagerForServers@odata.count": 1,
                "ManagerForServers": [
                        {
                                "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                        }
                ],
                "ManagerForChassis@odata.count": 1,
                "ManagerForChassis": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ]
        },
        "Actions": {
                "#Manager.Reset": {
                        "target": "/redfish/v1/Managers/BMC/Actions/Manager.Reset",
                        "ResetType@Redfish.AllowableValues": [
                                "ForceRestart"
                        ]
                }
        }
}
`

const testPathIntelManagers_BMC_NICs = "/redfish/v1/Managers/BMC/NICs"

const testPayloadIntelManagers_BMC_NICs = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "@odata.id": "/redfish/v1/Managers/BMC/NICs",
        "Name": "Ethernet Network Interface Collection",
        "Description": "Collection of EthernetInterfaces for this Manager",
        "Members@odata.count": 3,
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Managers/BMC/NICs/1"
                },
                {
                        "@odata.id": "/redfish/v1/Managers/BMC/NICs/2"
                },
                {
                        "@odata.id": "/redfish/v1/Managers/BMC/NICs/3"
                }
        ]
}
`

const testPathIntelManagers_BMC_NICs_1 = "/redfish/v1/Managers/BMC/NICs/1"

const testPayloadIntelManagers_BMC_NICs_1 = `
{
        "@odata.type": "#EthernetInterface.v1_2_0.EthernetInterface",
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Managers/BMC/NICs/1",
        "Id": "1",
        "Name": "Manager Ethernet Interface",
        "Description": "Management Network Interface",
        "Status": {
                "State": "Disabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "MACAddress": "a4-bf-01-2c-f7-d0",
        "HostName": "BMCA4BF012CF7D0",
        "NameServers": [
                "",
                ""
        ],
        "MTUSize": 1500,
        "MaxIPv6StaticAddresses": 1,
        "IPv4Addresses": [
                {
                        "Address": "0.0.0.0",
                        "SubnetMask": "0.0.0.0",
                        "AddressOrigin": "Static",
                        "Gateway": "0.0.0.0"
                }
        ],
        "IPv6StaticAddresses": [
                {
                        "Address": "",
                        "PrefixLength": 64
                }
        ],
        "IPv6DefaultGateway": "",
        "IPv6Addresses": [
                {
                        "Address": "",
                        "PrefixLength": 64,
                        "AddressOrigin": "Static"
                }
        ],
        "VLAN": {
                "VLANEnable": false,
                "VLANId": 0
        }
}
`

const testPathIntelManagers_BMC_NICs_2 = "/redfish/v1/Managers/BMC/NICs/2"

const testPayloadIntelManagers_BMC_NICs_2 = `
{
        "@odata.type": "#EthernetInterface.v1_2_0.EthernetInterface",
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Managers/BMC/NICs/2",
        "Id": "2",
        "Name": "Manager Ethernet Interface",
        "Description": "Management Network Interface",
        "Status": {
                "State": "Disabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "MACAddress": "a4-bf-01-2c-f7-d1",
        "HostName": "BMCA4BF012CF7D0",
        "NameServers": [
                "",
                ""
        ],
        "MTUSize": 1500,
        "MaxIPv6StaticAddresses": 1,
        "IPv4Addresses": [
                {
                        "Address": "0.0.0.0",
                        "SubnetMask": "0.0.0.0",
                        "AddressOrigin": "Static",
                        "Gateway": "0.0.0.0"
                }
        ],
        "IPv6StaticAddresses": [
                {
                        "Address": "",
                        "PrefixLength": 64
                }
        ],
        "IPv6DefaultGateway": "",
        "IPv6Addresses": [
                {
                        "Address": "",
                        "PrefixLength": 64,
                        "AddressOrigin": "Static"
                }
        ],
        "VLAN": {
                "VLANEnable": false,
                "VLANId": 0
        }
}
`

const testPathIntelManagers_BMC_NICs_3 = "/redfish/v1/Managers/BMC/NICs/3"

const testPayloadIntelManagers_BMC_NICs_3 = `
{
        "@odata.type": "#EthernetInterface.v1_2_0.EthernetInterface",
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Managers/BMC/NICs/3",
        "Id": "3",
        "Name": "Manager Ethernet Interface",
        "Description": "Management Network Interface",
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "MACAddress": "a4-bf-01-2c-f7-d2",
        "HostName": "BMCA4BF012CF7D0",
        "NameServers": [
                "",
                ""
        ],
        "MTUSize": 1500,
        "MaxIPv6StaticAddresses": 1,
        "IPv4Addresses": [
                {
                        "Address": "10.100.172.20",
                        "SubnetMask": "255.255.240.0",
                        "AddressOrigin": "Static",
                        "Gateway": "10.100.172.1"
                }
        ],
        "IPv6StaticAddresses": [
                {
                        "Address": "",
                        "PrefixLength": 64
                }
        ],
        "IPv6DefaultGateway": "",
        "IPv6Addresses": [
                {
                        "Address": "",
                        "PrefixLength": 64,
                        "AddressOrigin": "Static"
                }
        ],
        "VLAN": {
                "VLANEnable": false,
                "VLANId": 0
        }
}
`

const testPathIntelSystems = "/redfish/v1/Systems"

const testPayloadIntelSystems = `
{
        "@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
        "@odata.id": "/redfish/v1/Systems",
        "@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
        "Name": "Computer System Collection",
        "Members@odata.count": 1,
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715"
                }
        ]
}
`

const testPathIntelSystems_QSBP74304715 = "/redfish/v1/Systems/QSBP74304715"

const testPayloadIntelSystems_QSBP74304715 = `
{
        "@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715",
        "@odata.type": "#ComputerSystem.v1_3_0.ComputerSystem",
        "Id": "QSBP74304715",
        "SerialNumber": "QSBP74304715",
        "Name": "S2600BPB",
        "Model": "S2600BPB",
        "Manufacturer": "Cray Inc.",
        "PartNumber": "102002300",
        "SystemType": "Physical",
        "Description": "Computer System",
        "UUID": "f5f9d048-bb10-11e7-ab21-a4bf012cf7ce",
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        },
        "IndicatorLED": "Off",
        "PowerState": "On",
        "Boot": {
                "BootSourceOverrideEnabled": "Once",
                "BootSourceOverrideTarget": "None",
                "BootSourceOverrideMode": "Legacy",
                "BootSourceOverrideTarget@Redfish.AllowableValues": [
                        "None",
                        "Pxe",
                        "Cd",
                        "Usb",
                        "Hdd",
                        "BiosSetup",
                        "Diags",
                        "UefiTarget"
                ]
        },
        "BiosVersion": "SE5C620.86B.00.01.0013.C0001.030920180427",
        "ProcessorSummary": {
                "Count": 2,
                "Model": "Central Processor",
                "Status": {
                        "State": "Enabled",
                        "Health": "OK",
                        "HealthRollup": "OK"
                }
        },
        "MemorySummary": {
                "TotalSystemMemoryGiB": 48,
                "Status": {
                        "State": "Enabled",
                        "Health": "OK",
                        "HealthRollup": "OK"
                }
        },
        "Processors": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Processors"
        },
        "LogServices": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/LogServices"
        },
        "Memory": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory"
        },
        "Storage": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage"
        },
        "Links": {
                "Chassis": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/RackMount"
                        }
                ],
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/BMC"
                        }
                ]
        },
        "Actions": {
                "#ComputerSystem.Reset": {
                        "target": "/redfish/v1/Systems/QSBP74304715/Actions/ComputerSystem.Reset",
                        "ResetType@Redfish.AllowableValues": [
                                "On",
                                "ForceOff",
                                "GracefulShutdown",
                                "GracefulRestart",
                                "ForceRestart",
                                "Nmi"
                        ]
                }
        }
}
`
const testPathIntelSystems_QSBP74304715_Storage = "/redfish/v1/Systems/QSBP74304715/Storage"

const testPayloadIntelSystems_QSBP74304715_Storage = `
{
    "@odata.context": "/redfish/v1/$metadata#StorageCollection.StorageCollection",
    "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage",
    "@odata.type": "#StorageCollection.StorageCollection",
    "Name": "Storage Collection",
    "Members@odata.count": 2,
    "Members": [
        {
            "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1"
        },
        {
            "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2"
        }
    ],
    "Description" : "Collection of Storage resource instances",
    "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1 = "/redfish/v1/Systems/QSBP74304715/Storage/1"

const testPayloadIntelSystems_QSBP74304715_Storage_1 = `{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1",
   "@odata.type" : "#Storage.v1_5_0.Storage",
   "@odata.context" : "/redfish/v1/$metadata#Storage.Storage",
   "Description" : "This resource shall be used to represent resources that represent a storage subsystem in the Redfish specification.",
   "Name" : "Storage",
   "Drives": [
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD1"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD2"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD3"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD4"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD5"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD6"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD7"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD8"
           }
   ],
   "Drives@odata.count" : 8,
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2 = "/redfish/v1/Systems/QSBP74304715/Storage/2"

const testPayloadIntelSystems_QSBP74304715_Storage_2 = `{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2",
   "@odata.type" : "#Storage.v1_5_0.Storage",
   "@odata.context" : "/redfish/v1/$metadata#Storage.Storage",
   "Description" : "This resource shall be used to represent resources that represent a storage subsystem in the Redfish specification.",
   "Name" : "Storage",
   "Drives": [
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD1"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD2"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD3"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD4"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD5"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD6"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD7"
           },
           {
                   "@odata.id": "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD8"
           }
   ],
   "Drives@odata.count" : 8,
   "@odata.etag" : "W/\"1579629365\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_1 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD1"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_1 = `
{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD1",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400230-HDD1",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_2 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD2"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_2 = `
{
   "Id" : "2",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD2",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400231-HDD2",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_3 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD3"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_3 = `
{
   "Id" : "3",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD3",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400232-HDD2",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_4 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD4"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_4 = `
{
   "Id" : "4",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD4",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400233-HDD4",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_5 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD5"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_5 = `
{
   "Id" : "5",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD5",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400234-HDD5",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_6 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD6"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_6 = `
{
   "Id" : "6",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD6",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400236-HDD6",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_7 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD7"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_7 = `
{
   "Id" : "7",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD7",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400237-HDD7",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_1_Drives_8 = "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD8"

const testPayloadIntelSystems_QSBP74304715_Storage_1_Drives_8 = `
{
   "Id" : "8",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/1/Drives/HDD8",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400238-HDD8",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_1 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD1"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_1 = `
{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD1",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400230-2HDD1",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629365\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_2 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD2"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_2 = `
{
   "Id" : "2",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD2",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400231-2HDD2",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_3 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD3"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_3 = `
{
   "Id" : "3",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD3",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400232-2HDD3",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_4 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD4"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_4 = `
{
   "Id" : "4",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD4",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400233-2HDD4",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_5 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD5"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_5 = `
{
   "Id" : "5",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD5",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400234-2HDD5",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_6 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD6"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_6 = `
{
   "Id" : "6",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD6",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400236-2HDD6",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_7 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD7"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_7 = `
{
   "Id" : "7",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD7",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400237-2HDD7",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Storage_2_Drives_8 = "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD8"

const testPayloadIntelSystems_QSBP74304715_Storage_2_Drives_8 = `
{
   "Id" : "8",
   "@odata.id" : "/redfish/v1/Systems/QSBP74304715/Storage/2/Drives/HDD8",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400238-2HDD8",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathIntelSystems_QSBP74304715_Processors = "/redfish/v1/Systems/QSBP74304715/Processors"

const testPayloadIntelSystems_QSBP74304715_Processors = `
{
        "@odata.context": "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Processors",
        "@odata.type": "#ProcessorCollection.ProcessorCollection",
        "Name": "Processor Collection",
        "Members@odata.count": 2,
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Processors/CPU1"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Processors/CPU2"
                }
        ]
}
`

const testPathIntelSystems_QSBP74304715_Processors_CPU1 = "/redfish/v1/Systems/QSBP74304715/Processors/CPU1"

const testPayloadIntelSystems_QSBP74304715_Processors_CPU1 = `
{
        "@odata.context": "/redfish/v1/$metadata#Processor.Processor",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Processors/CPU1",
        "@odata.type": "#Processor.v1_0_3.Processor",
        "Name": "Processor 1",
        "Id": "CPU1",
        "Socket": "CPU 1",
        "ProcessorType": "CPU",
        "ProcessorArchitecture": "x86",
        "InstructionSet": "x86-64",
        "Manufacturer": "Intel(R) Corporation",
        "Model": "Intel Xeon processor",
        "ProcessorId": {
                "VendorId": "Intel(R) Xeon(R) Silver 4112 CPU @ 2.60GHz",
                "IdentificationRegisters": "50-65-4",
                "EffectiveFamily": "0xb3",
                "EffectiveModel": "0x3"
        },
        "MaxSpeedMHz": 4000,
        "TotalCores": 4,
        "TotalThreads": 4,
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

const testPathIntelSystems_QSBP74304715_Processors_CPU2 = "/redfish/v1/Systems/QSBP74304715/Processors/CPU2"

const testPayloadIntelSystems_QSBP74304715_Processors_CPU2 = `
{
        "@odata.context": "/redfish/v1/$metadata#Processor.Processor",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Processors/CPU2",
        "@odata.type": "#Processor.v1_0_3.Processor",
        "Name": "Processor 2",
        "Id": "CPU2",
        "Socket": "CPU 2",
        "ProcessorType": "CPU",
        "ProcessorArchitecture": "x86",
        "InstructionSet": "x86-64",
        "Manufacturer": "Intel(R) Corporation",
        "Model": "Intel Xeon processor",
        "ProcessorId": {
                "VendorId": "Intel(R) Xeon(R) Silver 4112 CPU @ 2.60GHz",
                "IdentificationRegisters": "50-65-4",
                "EffectiveFamily": "0xb3",
                "EffectiveModel": "0x3"
        },
        "MaxSpeedMHz": 4000,
        "TotalCores": 4,
        "TotalThreads": 4,
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

const testPathIntelSystems_QSBP74304715_Memory = "/redfish/v1/Systems/QSBP74304715/Memory"

const testPayloadIntelSystems_QSBP74304715_Memory = `
{
        "@odata.context": "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory",
        "@odata.type": "#MemoryCollection.MemoryCollection",
        "Name": "MemoryCollection",
        "Members@odata.count": 6,
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory1"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory2"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory3"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory4"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory5"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory6"
                }
        ]
}
`

const testPathIntelSystems_QSBP74304715_Memory_Memory1 = "/redfish/v1/Systems/QSBP74304715/Memory/Memory1"

const testPayloadIntelSystems_QSBP74304715_Memory_Memory1 = `
{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory1",
        "@odata.type": "#Memory.v1_1_0.Memory",
        "Id": "Memory1",
        "Name": "Memory 1",
        "Description": "System Memory",
        "MemoryType": "DRAM",
        "MemoryDeviceType": "DDR4",
        "BaseModuleType": "RDIMM",
        "CapacityMiB": 8192,
        "DataWidthBits": 64,
        "BusWidthBits": 72,
        "Manufacturer": "Hynix",
        "SerialNumber": "102E2435",
        "PartNumber": "HMA41GR7MFR8N-TFTD  ",
        "AllowedSpeedsMHz": [
                2133
        ],
        "MemoryMedia": [
                "DRAM"
        ],
        "RankCount": 2,
        "DeviceLocator": "CPU1_DIMM_A1",
        "ErrorCorrection": "MultiBitECC",
        "OperatingSpeedMhz": 2133,
        "Metrics": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory1/MemoryMetrics"
        },
        "Oem": {
                "Intel_Rackscale": {
                        "@odata.type": "#Intel.Oem.DimmConfig",
                        "VoltageVolt": 1.2
                }
        },
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

const testPathIntelSystems_QSBP74304715_Memory_Memory2 = "/redfish/v1/Systems/QSBP74304715/Memory/Memory2"

const testPayloadIntelSystems_QSBP74304715_Memory_Memory2 = `
{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory2",
        "@odata.type": "#Memory.v1_1_0.Memory",
        "Id": "Memory2",
        "Name": "Memory 2",
        "Description": "System Memory",
        "MemoryType": "DRAM",
        "MemoryDeviceType": "DDR4",
        "BaseModuleType": "RDIMM",
        "CapacityMiB": 8192,
        "DataWidthBits": 64,
        "BusWidthBits": 72,
        "Manufacturer": "Hynix",
        "SerialNumber": "102533E8",
        "PartNumber": "HMA41GR7MFR8N-TFTD  ",
        "AllowedSpeedsMHz": [
                2133
        ],
        "MemoryMedia": [
                "DRAM"
        ],
        "RankCount": 2,
        "DeviceLocator": "CPU1_DIMM_B1",
        "ErrorCorrection": "MultiBitECC",
        "OperatingSpeedMhz": 2133,
        "Metrics": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory2/MemoryMetrics"
        },
        "Oem": {
                "Intel_Rackscale": {
                        "@odata.type": "#Intel.Oem.DimmConfig",
                        "VoltageVolt": 1.2
                }
        },
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

const testPathIntelSystems_QSBP74304715_Memory_Memory3 = "/redfish/v1/Systems/QSBP74304715/Memory/Memory3"

const testPayloadIntelSystems_QSBP74304715_Memory_Memory3 = `
{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory3",
        "@odata.type": "#Memory.v1_1_0.Memory",
        "Id": "Memory3",
        "Name": "Memory 3",
        "Description": "System Memory",
        "MemoryType": "DRAM",
        "MemoryDeviceType": "DDR4",
        "BaseModuleType": "RDIMM",
        "CapacityMiB": 8192,
        "DataWidthBits": 64,
        "BusWidthBits": 72,
        "Manufacturer": "Hynix",
        "SerialNumber": "102CE000",
        "PartNumber": "HMA41GR7MFR8N-TFT1  ",
        "AllowedSpeedsMHz": [
                2133
        ],
        "MemoryMedia": [
                "DRAM"
        ],
        "RankCount": 2,
        "DeviceLocator": "CPU1_DIMM_C1",
        "ErrorCorrection": "MultiBitECC",
        "OperatingSpeedMhz": 2133,
        "Metrics": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory3/MemoryMetrics"
        },
        "Oem": {
                "Intel_Rackscale": {
                        "@odata.type": "#Intel.Oem.DimmConfig",
                        "VoltageVolt": 1.2
                }
        },
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

const testPathIntelSystems_QSBP74304715_Memory_Memory4 = "/redfish/v1/Systems/QSBP74304715/Memory/Memory4"

const testPayloadIntelSystems_QSBP74304715_Memory_Memory4 = `
{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory4",
        "@odata.type": "#Memory.v1_1_0.Memory",
        "Id": "Memory4",
        "Name": "Memory 4",
        "Description": "System Memory",
        "MemoryType": "DRAM",
        "MemoryDeviceType": "DDR4",
        "BaseModuleType": "RDIMM",
        "CapacityMiB": 8192,
        "DataWidthBits": 64,
        "BusWidthBits": 72,
        "Manufacturer": "Hynix",
        "SerialNumber": "102E4933",
        "PartNumber": "HMA41GR7MFR8N-TFTD  ",
        "AllowedSpeedsMHz": [
                2133
        ],
        "MemoryMedia": [
                "DRAM"
        ],
        "RankCount": 2,
        "DeviceLocator": "CPU2_DIMM_A1",
        "ErrorCorrection": "MultiBitECC",
        "OperatingSpeedMhz": 2133,
        "Metrics": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory4/MemoryMetrics"
        },
        "Oem": {
                "Intel_Rackscale": {
                        "@odata.type": "#Intel.Oem.DimmConfig",
                        "VoltageVolt": 1.2
                }
        },
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

const testPathIntelSystems_QSBP74304715_Memory_Memory5 = "/redfish/v1/Systems/QSBP74304715/Memory/Memory5"

const testPayloadIntelSystems_QSBP74304715_Memory_Memory5 = `
{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory5",
        "@odata.type": "#Memory.v1_1_0.Memory",
        "Id": "Memory5",
        "Name": "Memory 5",
        "Description": "System Memory",
        "MemoryType": "DRAM",
        "MemoryDeviceType": "DDR4",
        "BaseModuleType": "RDIMM",
        "CapacityMiB": 8192,
        "DataWidthBits": 64,
        "BusWidthBits": 72,
        "Manufacturer": "Hynix",
        "SerialNumber": "102E2688",
        "PartNumber": "HMA41GR7MFR8N-TFTD  ",
        "AllowedSpeedsMHz": [
                2133
        ],
        "MemoryMedia": [
                "DRAM"
        ],
        "RankCount": 2,
        "DeviceLocator": "CPU2_DIMM_B1",
        "ErrorCorrection": "MultiBitECC",
        "OperatingSpeedMhz": 2133,
        "Metrics": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory5/MemoryMetrics"
        },
        "Oem": {
                "Intel_Rackscale": {
                        "@odata.type": "#Intel.Oem.DimmConfig",
                        "VoltageVolt": 1.2
                }
        },
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

const testPathIntelSystems_QSBP74304715_Memory_Memory6 = "/redfish/v1/Systems/QSBP74304715/Memory/Memory6"

const testPayloadIntelSystems_QSBP74304715_Memory_Memory6 = `
{
        "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
        "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory6",
        "@odata.type": "#Memory.v1_1_0.Memory",
        "Id": "Memory6",
        "Name": "Memory 6",
        "Description": "System Memory",
        "MemoryType": "DRAM",
        "MemoryDeviceType": "DDR4",
        "BaseModuleType": "RDIMM",
        "CapacityMiB": 8192,
        "DataWidthBits": 64,
        "BusWidthBits": 72,
        "Manufacturer": "Hynix",
        "SerialNumber": "102534D7",
        "PartNumber": "HMA41GR7MFR8N-TFTD  ",
        "AllowedSpeedsMHz": [
                2133
        ],
        "MemoryMedia": [
                "DRAM"
        ],
        "RankCount": 2,
        "DeviceLocator": "CPU2_DIMM_C1",
        "ErrorCorrection": "MultiBitECC",
        "OperatingSpeedMhz": 2133,
        "Metrics": {
                "@odata.id": "/redfish/v1/Systems/QSBP74304715/Memory/Memory6/MemoryMetrics"
        },
        "Oem": {
                "Intel_Rackscale": {
                        "@odata.type": "#Intel.Oem.DimmConfig",
                        "VoltageVolt": 1.2
                }
        },
        "Status": {
                "State": "Enabled",
                "Health": "OK",
                "HealthRollup": "OK"
        }
}
`

//////////////////////////////////////////////////////////////////////////////
//                          Dell - Mock Client
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncDell1() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathDell_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadDell_redfish_v1)),
				// Header must always be non-nil or it will cause a panic.
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellManagers_iDRACEmbedded1_AccountService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellManagers_iDRACEmbedded1_AccountService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSessionService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSessionService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellEventService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadDellEventService)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellTaskService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellTaskService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellChassis:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellChassis)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellChassis_SystemEmbedded1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellChassis_SystemEmbedded1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellChassis_SystemEmbedded1Power:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellChassis_SystemEmbedded1Power)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellManagers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellManagers)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellManagers_iDRACEmbedded1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadDellManagers_iDRACEmbedded1)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellManagers_iDRACEmbedded1_EthernetInterfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellManagers_iDRACEmbedded1_EthernetInterfaces)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellManagers_iDRACEmbedded1_EthernetInterfaces_iDRACEmbedded1_23NIC1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellManagers_iDRACEmbedded1_EthernetInterfaces_iDRACEmbedded1_23NIC1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_EthernetInterfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body: ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces)),

				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated131:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated131)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated141:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated141)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated111:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated111)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated121:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated121)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_Processors:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_Processors)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_Processors_CPUSocket1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_Processors_CPUSocket1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathDellSystems_SystemEmbedded1_Processors_CPUSocket2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadDellSystems_SystemEmbedded1_Processors_CPUSocket2)),
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
				Header: make(http.Header),
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////
//                      Dell  - Response payloads
//////////////////////////////////////////////////////////////////////////////

const testPathDell_redfish_v1 = "/redfish/v1"

const testPayloadDell_redfish_v1 = `
{
        "@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
        "@odata.id": "/redfish/v1",
        "@odata.type": "#ServiceRoot.v1_0_2.ServiceRoot",
        "AccountService": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/AccountService"
        },
        "Chassis": {
                "@odata.id": "/redfish/v1/Chassis"
        },
        "Description": "Root Service",
        "EventService": {
                "@odata.id": "/redfish/v1/EventService"
        },
        "Id": "RootService",
        "JsonSchemas": {
                "@odata.id": "/redfish/v1/JSONSchemas"
        },
        "Links": {
                "Sessions": {
                        "@odata.id": "/redfish/v1/Sessions"
                }
        },
        "Managers": {
                "@odata.id": "/redfish/v1/Managers"
        },
        "Name": "Root Service",
        "RedfishVersion": "1.0.2",
        "Registries": {
                "@odata.id": "/redfish/v1/Registries"
        },
        "SessionService": {
                "@odata.id": "/redfish/v1/SessionService"
        },
        "Systems": {
                "@odata.id": "/redfish/v1/Systems"
        },
        "Tasks": {
                "@odata.id": "/redfish/v1/TaskService"
        }
}
`

const testPathDellManagers_iDRACEmbedded1_AccountService = "/redfish/v1/Managers/iDRAC.Embedded.1/AccountService"

const testPayloadDellManagers_iDRACEmbedded1_AccountService = `
{
        "@odata.context": "/redfish/v1/$metadata#AccountService.AccountService",
        "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/AccountService",
        "@odata.type": "#AccountService.v1_0_2.AccountService",
        "AccountLockoutCounterResetAfter": 0,
        "AccountLockoutDuration": 0,
        "AccountLockoutThreshold": 0,
        "Accounts": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Accounts"
        },
        "AuthFailureLoggingThreshold": 2,
        "Description": "BMC User Accounts",
        "Id": "AccountService",
        "MaxPasswordLength": 20,
        "MinPasswordLength": 1,
        "Name": "Account Service",
        "Roles": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/Roles"
        },
        "ServiceEnabled": true,
        "Status": {
                "Health": "Ok",
                "State": "Enabled"
        }
}
`

const testPathDellSessionService = "/redfish/v1/SessionService"

const testPayloadDellSessionService = `
{
        "@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
        "@odata.id": "/redfish/v1/SessionService",
        "@odata.type": "#SessionService.v1_0_2.SessionService",
        "Description": "Session Service",
        "Id": "SessionService",
        "Name": "Session Service",
        "ServiceEnabled": true,
        "SessionTimeout": 1800,
        "Sessions": {
                "@odata.id": "/redfish/v1/Sessions"
        }
}
`

const testPathDellEventService = "/redfish/v1/EventService"

const testPayloadDellEventService = `
{
        "@odata.context": "/redfish/v1/$metadata#EventService.EventService",
        "@odata.id": "/redfish/v1/EventService",
        "@odata.type": "#EventService.v1_0_2.EventService",
        "Actions": {
                "#EventService.SubmitTestEvent": {
                        "EventType@Redfish.AllowableValues": [
                                "StatusChange",
                                "ResourceUpdated",
                                "ResourceAdded",
                                "ResourceRemoved",
                                "Alert"
                        ],
                        "target": "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"
                }
        },
        "DeliveryRetryAttempts": 3,
        "DeliveryRetryIntervalInSeconds": 30,
        "Description": "Event Service represents the properties for the service",
        "EventTypesForSubscription": [
                "StatusChange",
                "ResourceUpdated",
                "ResourceAdded",
                "ResourceRemoved",
                "Alert"
        ],
        "EventTypesForSubscription@odata.count": 5,
        "Id": "EventService",
        "IgnoreCertificateErrors": "Yes",
        "Name": "Event Service",
        "ServiceEnabled": false,
        "Status": {
                "Health": "Ok",
                "HealthRollUp": "Ok",
                "State": "Disabled"
        },
        "Subscriptions": {
                "@odata.id": "/redfish/v1/EventService/Subscriptions"
        }
}
`

const testPathDellTaskService = "/redfish/v1/TaskService"

const testPayloadDellTaskService = `
{
        "@odata.context": "/redfish/v1/$metadata#TaskService.TaskService",
        "@odata.id": "/redfish/v1/TaskService",
        "@odata.type": "#TaskService.v1_0_2.TaskService",
        "DateTime": "2018-10-01T15:31:16-05:00",
        "Description": "Represents the properties for the service itself and has links to the actual list of Tasks",
        "Id": "TaskService",
        "Name": "Task Service",
        "ServiceEnabled": false,
        "Status": {
                "Health": "OK",
                "State": "Disabled"
        },
        "Tasks": {
                "@odata.id": "/redfish/v1/TaskService/Tasks"
        }
}
`

const testPathDellChassis = "/redfish/v1/Chassis"

const testPayloadDellChassis = `
{
        "@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
        "@odata.id": "/redfish/v1/Chassis/",
        "@odata.type": "#ChassisCollection.ChassisCollection",
        "Description": "Collection of Chassis",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Chassis/System.Embedded.1"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Chassis Collection"
}
`

const testPathDellChassis_SystemEmbedded1 = "/redfish/v1/Chassis/System.Embedded.1"

const testPayloadDellChassis_SystemEmbedded1 = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis",
        "@odata.id": "/redfish/v1/Chassis/System.Embedded.1",
        "@odata.type": "#Chassis.v1_0_2.Chassis",
        "Actions": {
                "#Chassis.Reset": {
                        "ResetType@Redfish.AllowableValues": [
                                "On",
                                "ForceOff"
                        ],
                        "Target": "/redfish/v1/Chassis/System.Embedded.1/Actions/Chassis.Reset"
                }
        },
        "AssetTag": "",
        "ChassisType": "Enclosure",
        "Description": "It represents the properties for physical components for any system.It represent racks, rackmount servers, blades, standalone, modular systems,enclosures, and all other containers.The non-cpu/device centric parts of the schema are all accessed either directly or indirectly through this resource.",
        "Id": "System.Embedded.1",
        "IndicatorLED": "Off",
        "Links": {
                "ComputerSystems": [
                        {
                                "@odata.id": "/redfish/v1/Systems/System.Embedded.1"
                        }
                ],
                "ComputerSystems@odata.count": 1,
                "Contains": [],
                "Contains@odata.count": 0,
                "CooledBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.1A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.2A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.3A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.4A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.5A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.6A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.7A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.1B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.2B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.3B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.4B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.5B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.6B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.7B"
                        }
                ],
                "CooledBy@odata.count": 14,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "PoweredBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.1"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2"
                        }
                ],
                "PoweredBy@odata.count": 2
        },
        "Manufacturer": " ",
        "Model": " ",
        "Name": "Computer System Chassis",
        "PartNumber": "02C2CPA06",
        "Power": {
                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power"
        },
        "PowerState": "On",
        "SKU": "9W0WHK2",
        "SerialNumber": "CNIVC0076B0735",
        "Status": {
                "Health": "Ok",
                "HealthRollUp": "Ok",
                "State": "Enabled"
        },
        "Thermal": {
                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Thermal"
        }
}
`

const testPathDellChassis_SystemEmbedded1Power = "/redfish/v1/Chassis/System.Embedded.1/Power"

const testPayloadDellChassis_SystemEmbedded1Power = `
{
    "@odata.context": "/redfish/v1/$metadata#Power.Power",
    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power",
    "@odata.type": "#Power.v1_0_2.Power",
    "Description": "Power",
    "Id": "Power",
    "Name": "Power",
    "PowerControl": [
        {
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerControl",
            "MemberID": "PowerControl",
            "Name": "System Power Control",
            "PowerAllocatedWatts": 750,
            "PowerAvailableWatts": 0,
            "PowerCapacityWatts": 980,
            "PowerConsumedWatts": 129,
            "PowerLimit": {
                "CorrectionInMs": 0,
                "LimitException": "HardPowerOff",
                "LimitInWatts": 355
            },
            "PowerMetrics": {
                "AverageConsumedWatts": 129,
                "IntervalInMin": 60,
                "MaxConsumedWatts": 142,
                "MinConsumedWatts": 128
            },
            "PowerRequestedWatts": 373,
            "RelatedItem": [
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1"
                },
                {
                    "@odata.id": "/redfish/v1/Systems/System.Embedded.1"
                }
            ],
            "RelatedItem@odata.count": 2
        }
    ],
    "PowerControl@odata.count": 1,
    "PowerSupplies": [
        {
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.1",
            "FirmwareVersion": "00.10.37",
            "LastPowerOutputWatts": 750,
            "LineInputVoltage": 206,
            "LineInputVoltageType": "ACMidLine",
            "MemberID": "PSU.Slot.1",
            "Model": "PWR SPLY,750W,RDNT,EMSN       ",
            "Name": "PS1 Status",
            "PartNumber": "0TPJ2XA01",
            "PowerCapacityWatts": 750,
            "PowerSupplyType": "AC",
            "Redundancy": [
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/Redundancy/iDRAC.Embedded.1%23SystemBoardPSRedundancy",
                    "MaxNumSupported": 4,
                    "MemberID": "iDRAC.Embedded.1#SystemBoardPSRedundancy",
                    "MinNumNeeded": 2,
                    "Mode": [
                        {
                            "Member": "N+1"
                        }
                    ],
                    "Name": "System Board PS Redundancy",
                    "RedundancySet": [
                        {
                            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.1"
                        },
                        {
                            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2"
                        }
                    ],
                    "RedundancySet@odata.count": 2,
                    "Status": {
                        "Health": "Ok",
                        "State": "Enabled"
                    }
                }
            ],
            "Redundancy@odata.count": 1,
            "RelatedItem": [
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1"
                }
            ],
            "RelatedItem@odata.count": 1,
            "SerialNumber": "PH1629853H00A7",
            "SparePartNumber": "0TPJ2XA01",
            "Status": {
                "Health": "OK",
                "State": "Enabled"
            }
        },
        {
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2",
            "FirmwareVersion": "00.10.37",
            "LastPowerOutputWatts": 750,
            "LineInputVoltage": 206,
            "LineInputVoltageType": "ACMidLine",
            "MemberID": "PSU.Slot.2",
            "Model": "PWR SPLY,750W,RDNT,EMSN       ",
            "Name": "PS2 Status",
            "PartNumber": "0TPJ2XA01",
            "PowerCapacityWatts": 750,
            "PowerSupplyType": "AC",
            "Redundancy": [
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/Redundancy/iDRAC.Embedded.1%23SystemBoardPSRedundancy",
                    "MaxNumSupported": 4,
                    "MemberID": "iDRAC.Embedded.1#SystemBoardPSRedundancy",
                    "MinNumNeeded": 2,
                    "Mode": [
                        {
                            "Member": "N+1"
                        }
                    ],
                    "Name": "System Board PS Redundancy",
                    "RedundancySet": [
                        {
                            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.1"
                        },
                        {
                            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2"
                        }
                    ],
                    "RedundancySet@odata.count": 2,
                    "Status": {
                        "Health": "Ok",
                        "State": "Enabled"
                    }
                }
            ],
            "Redundancy@odata.count": 1,
            "RelatedItem": [
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1"
                }
            ],
            "RelatedItem@odata.count": 1,
            "SerialNumber": "PH1629853H008C",
            "SparePartNumber": "0TPJ2XA01",
            "Status": {
                "Health": "OK",
                "State": "Enabled"
            }
        }
    ],
    "PowerSupplies@odata.count": 2,
    "Redundancy": [
        {
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/Redundancy/iDRAC.Embedded.1%23SystemBoardPSRedundancy",
            "MaxNumSupported": 4,
            "MemberID": "iDRAC.Embedded.1#SystemBoardPSRedundancy",
            "MinNumNeeded": 2,
            "Mode": [
                {
                    "Member": "N+1"
                }
            ],
            "Name": "System Board PS Redundancy",
            "RedundancySet": [
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.1"
                },
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2"
                }
            ],
            "RedundancySet@odata.count": 2,
            "Status": {
                "Health": "Ok",
                "State": "Enabled"
            }
        }
    ],
    "Redundancy@odata.count": 1,
    "Voltages": [
        {
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23CPU1VCOREPG",
            "LowerThresholdCritical": null,
            "LowerThresholdFatal": null,
            "LowerThresholdNonCritical": null,
            "MaxReadingRange": 0,
            "MemberID": "iDRAC.Embedded.1#CPU1VCOREPG",
            "MinReadingRange": 0,
            "Name": "CPU1 VCORE PG",
            "PhysicalContext": "CPU",
            "ReadingVolts": 1,
            "RelatedItem": [
                {
                    "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"
                }
            ],
            "RelatedItem@odata.count": 1,
            "SensorNumber": 35,
            "Status": {
                "Health": "OK",
                "State": "Enabled"
            },
            "UpperThresholdCritical": null,
            "UpperThresholdFatal": null,
            "UpperThresholdNonCritical": null
        },
        {
            "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Voltages/iDRAC.Embedded.1%23PS2Voltage",
            "LowerThresholdCritical": null,
            "LowerThresholdFatal": null,
            "LowerThresholdNonCritical": null,
            "MaxReadingRange": 0,
            "MemberID": "iDRAC.Embedded.1#PS2Voltage",
            "MinReadingRange": 0,
            "Name": "PS2 Voltage 2",
            "PhysicalContext": "PowerSupply",
            "ReadingVolts": 206.0,
            "RelatedItem": [
                {
                    "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2"
                }
            ],
            "RelatedItem@odata.count": 1,
            "SensorNumber": 109,
            "Status": {
                "Health": "OK",
                "State": "Enabled"
            },
            "UpperThresholdCritical": null,
            "UpperThresholdFatal": null,
            "UpperThresholdNonCritical": null
        }
    ],
    "Voltages@odata.count": 2
}
`

const testPathDellManagers = "/redfish/v1/Managers"

const testPayloadDellManagers = `
{
        "@odata.context": "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
        "@odata.id": "/redfish/v1/Managers",
        "@odata.type": "#ManagerCollection.ManagerCollection",
        "Description": "BMC",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Manager"
}
`

const testPathDellManagers_iDRACEmbedded1 = "/redfish/v1/Managers/iDRAC.Embedded.1"

const testPayloadDellManagers_iDRACEmbedded1 = `
{
        "@odata.context": "/redfish/v1/$metadata#Manager.Manager",
        "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1",
        "@odata.type": "#Manager.v1_0_2.Manager",
        "Actions": {
                "#Manager.Reset": {
                        "ResetType@Redfish.AllowableValues": [
                                "GracefulRestart"
                        ],
                        "target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Manager.Reset"
                },
                "Oem": {
                        "OemManager.v1_0_0#OemManager.ExportSystemConfiguration": {
                                "ExportFormat@Redfish.AllowableValues": [
                                        "XML"
                                ],
                                "ExportUse@Redfish.AllowableValues": [
                                        "Default",
                                        "Clone",
                                        "Replace"
                                ],
                                "IncludeInExport@Redfish.AllowableValues": [
                                        "Default",
                                        "IncludeReadOnly",
                                        "IncludePasswordHashValues"
                                ],
                                "ShareParameters": {
                                        "ShareParameters@Redfish.AllowableValues": [
                                                "IPAddress",
                                                "ShareName",
                                                "FileName",
                                                "UserName",
                                                "Password",
                                                "Workgroup"
                                        ],
                                        "ShareType@Redfish.AllowableValues": [
                                                "NFS",
                                                "CIFS"
                                        ],
                                        "Target@Redfish.AllowableValues": [
                                                "ALL",
                                                "IDRAC",
                                                "BIOS",
                                                "NIC",
                                                "RAID"
                                        ]
                                },
                                "target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ExportSystemConfiguration"
                        },
                        "OemManager.v1_0_0#OemManager.ImportSystemConfiguration": {
                                "HostPowerState@Redfish.AllowableValues": [
                                        "On",
                                        "Off"
                                ],
                                "ImportSystemConfiguration@Redfish.AllowableValues": [
                                        "TimeToWait",
                                        "ImportBuffer"
                                ],
                                "ShareParameters": {
                                        "ShareParameters@Redfish.AllowableValues": [
                                                "IPAddress",
                                                "ShareName",
                                                "FileName",
                                                "UserName",
                                                "Password",
                                                "Workgroup"
                                        ],
                                        "ShareType@Redfish.AllowableValues": [
                                                "NFS",
                                                "CIFS"
                                        ],
                                        "Target@Redfish.AllowableValues": [
                                                "ALL",
                                                "IDRAC",
                                                "BIOS",
                                                "NIC",
                                                "RAID"
                                        ]
                                },
                                "ShutdownType@Redfish.AllowableValues": [
                                        "Graceful",
                                        "Forced",
                                        "NoReboot"
                                ],
                                "target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfiguration"
                        },
                        "OemManager.v1_0_0#OemManager.ImportSystemConfigurationPreview": {
                                "ImportSystemConfigurationPreview@Redfish.AllowableValues": [
                                        "ImportBuffer"
                                ],
                                "ShareParameters": {
                                        "ShareParameters@Redfish.AllowableValues": [
                                                "IPAddress",
                                                "ShareName",
                                                "FileName",
                                                "UserName",
                                                "Password",
                                                "Workgroup"
                                        ],
                                        "ShareType@Redfish.AllowableValues": [
                                                "NFS",
                                                "CIFS"
                                        ],
                                        "Target@Redfish.AllowableValues": [
                                                "ALL"
                                        ]
                                },
                                "target": "/redfish/v1/Managers/iDRAC.Embedded.1/Actions/Oem/EID_674_Manager.ImportSystemConfigurationPreview"
                        }
                }
        },
        "CommandShell": {
                "ConnectTypesSupported": [
                        "SSH",
                        "Telnet",
                        "IPMI"
                ],
                "ConnectTypesSupported@odata.count": 3,
                "MaxConcurrentSessions": 5,
                "ServiceEnabled": true
        },
        "DateTime": "2018-10-01T15:31:33-05:00",
        "DateTimeLocalOffset": "-05:00",
        "Description": "BMC",
        "EthernetInterfaces": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces"
        },
        "FirmwareVersion": "2.40.40.40",
        "GraphicalConsole": {
                "ConnectTypesSupported": [
                        "KVMIP"
                ],
                "ConnectTypesSupported@odata.count": 1,
                "MaxConcurrentSessions": 6,
                "ServiceEnabled": true
        },
        "Id": "iDRAC.Embedded.1",
        "Links": {
                "ManagerForChassis": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1"
                        }
                ],
                "ManagerForChassis@odata.count": 1,
                "ManagerForServers": [
                        {
                                "@odata.id": "/redfish/v1/Systems/System.Embedded.1"
                        }
                ],
                "ManagerForServers@odata.count": 1
        },
        "LogServices": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/LogServices"
        },
        "ManagerType": "BMC",
        "Model": "13G Monolithic",
        "Name": "Manager",
        "NetworkProtocol": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/NetworkProtocol"
        },
        "Redundancy": [],
        "Redundancy@odata.count": 0,
        "RedundancySet": [],
        "RedundancySet@odata.count": 0,
        "SerialConsole": {
                "ConnectTypesSupported": [],
                "ConnectTypesSupported@odata.count": 0,
                "MaxConcurrentSessions": 0,
                "ServiceEnabled": false
        },
        "SerialInterfaces": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/SerialInterfaces"
        },
        "Status": {
                "Health": "Ok",
                "State": "Enabled"
        },
        "UUID": "324b484f-c0b9-5780-3010-00574c4c4544",
        "VirtualMedia": {
                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/VirtualMedia"
        }
}
`

const testPathDellManagers_iDRACEmbedded1_EthernetInterfaces = "/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces"

const testPayloadDellManagers_iDRACEmbedded1_EthernetInterfaces = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces",
        "@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "Description": "Collection of EthernetInterfaces for this Manager",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces/iDRAC.Embedded.1%23NIC.1"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Ethernet Network Interface Collection"
}
`

const testPathDellManagers_iDRACEmbedded1_EthernetInterfaces_iDRACEmbedded1_23NIC1 = "/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces/iDRAC.Embedded.1%23NIC.1"

const testPayloadDellManagers_iDRACEmbedded1_EthernetInterfaces_iDRACEmbedded1_23NIC1 = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1/EthernetInterfaces/iDRAC.Embedded.1%23NIC.1",
        "@odata.type": "#EthernetInterface.v1_0_2.EthernetInterface",
        "AutoNeg": true,
        "Description": "Management Network Interface",
        "FQDN": "x0c0s26.crush.next.cray.com",
        "FullDuplex": true,
        "HostName": "x0c0s26",
        "IPv4Addresses": [
                {
                        "Address": "10.100.16.26",
                        "AddressOrigin": null,
                        "Gateway": "10.100.16.1",
                        "SubnetMask": "255.255.240.0"
                }
        ],
        "IPv4Addresses@odata.count": 1,
        "IPv6AddressPolicyTable": [],
        "IPv6AddressPolicyTable@odata.count": 0,
        "IPv6Addresses": [
                {
                        "Address": "::",
                        "AddressOrigin": "Static",
                        "AddressState": "Preferred",
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": "Failed",
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                },
                {
                        "Address": "::",
                        "AddressOrigin": null,
                        "AddressState": null,
                        "PrefixLength": 64
                }
        ],
        "IPv6Addresses@odata.count": 15,
        "IPv6DefaultGateway": "::",
        "IPv6StaticAddresses": [
                {
                        "Address": "::",
                        "PrefixLength": 64
                }
        ],
        "IPv6StaticAddresses@odata.count": 1,
        "Id": "iDRAC.Embedded.1#NIC.1",
        "InterfaceEnabled": true,
        "MACAddress": "50:9A:4C:A8:6F:CC",
        "MTUSize": 1500,
        "MaxIPv6StaticAddresses": 16,
        "Name": "Manager Ethernet Interface",
        "NameServers": [
                "0.0.0.0",
                "0.0.0.0",
                "0.0.0.0",
                "0.0.0.0",
                "::",
                "::",
                "::",
                "::"
        ],
        "NameServers@odata.count": 8,
        "PermanentMACAddress": "50:9A:4C:A8:6F:CC",
        "SpeedMbps": 1000,
        "Status": {
                "Health": "Ok",
                "State": "Enabled"
        },
        "VLAN": {
                "VLANEnable": false,
                "VLANId": 1
        }
}
`

const testPathDellSystems = "/redfish/v1/Systems"

const testPayloadDellSystems = `
{
        "@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
        "@odata.id": "/redfish/v1/Systems",
        "@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
        "Description": "Collection of Computer Systems",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Systems/System.Embedded.1"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Computer System Collection"
}
`

const testPathDellSystems_SystemEmbedded1 = "/redfish/v1/Systems/System.Embedded.1"

const testPayloadDellSystems_SystemEmbedded1 = `
{
        "@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1",
        "@odata.type": "#ComputerSystem.v1_0_2.ComputerSystem",
        "Actions": {
                "#ComputerSystem.Reset": {
                        "ResetType@Redfish.AllowableValues": [
                                "On",
                                "ForceOff",
                                "GracefulRestart",
                                "PushPowerButton",
                                "Nmi"
                        ],
                        "target": "/redfish/v1/Systems/System.Embedded.1/Actions/ComputerSystem.Reset"
                }
        },
        "AssetTag": "",
        "BiosVersion": "2.4.3",
        "Boot": {
                "BootSourceOverrideEnabled": "Once",
                "BootSourceOverrideTarget": "None",
                "BootSourceOverrideTarget@Redfish.AllowableValues": [
                        "None",
                        "Pxe",
                        "Cd",
                        "Floppy",
                        "Hdd",
                        "BiosSetup",
                        "Utilities",
                        "UefiTarget",
                        "SDCard"
                ],
                "UefiTargetBootSourceOverride": ""
        },
        "Description": "Computer System which represents a machine (physical or virtual) and the local resources such as memory, cpu and other devices that can be accessed from that machine.",
        "EthernetInterfaces": {
                "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces"
        },
        "HostName": "MINWINPC",
        "Id": "System.Embedded.1",
        "IndicatorLED": "Off",
        "Links": {
                "Chassis": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1"
                        }
                ],
                "Chassis@odata.count": 1,
                "CooledBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.1A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.2A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.3A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.4A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.5A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.6A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.7A"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.1B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.2B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.3B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.4B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.5B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.6B"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Sensors/Fans/0x17||Fan.Embedded.7B"
                        }
                ],
                "CooledBy@odata.count": 14,
                "ManagedBy": [
                        {
                                "@odata.id": "/redfish/v1/Managers/iDRAC.Embedded.1"
                        }
                ],
                "ManagedBy@odata.count": 1,
                "PoweredBy": [
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.1"
                        },
                        {
                                "@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Power/PowerSupplies/PSU.Slot.2"
                        }
                ],
                "PoweredBy@odata.count": 2
        },
        "Manufacturer": " ",
        "MemorySummary": {
                "Status": {
                        "Health": "OK",
                        "HealthRollUp": "OK",
                        "State": "Enabled"
                },
                "TotalSystemMemoryGiB": 128.0
        },
        "Model": " ",
        "Name": "System",
        "PartNumber": "02C2CPA06",
        "PowerState": "On",
        "ProcessorSummary": {
                "Count": 2,
                "Model": "Intel(R) Xeon(R) CPU E5-2650 v4 @ 2.20GHz",
                "Status": {
                        "Health": "OK",
                        "HealthRollUp": "OK",
                        "State": "Enabled"
                }
        },
        "Processors": {
                "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Processors"
        },
        "SKU": "9W0WHK2",
        "SerialNumber": "CNIVC0076B0735",
        "SimpleStorage": {
                "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Storage/Controllers"
        },
        "Status": {
                "Health": "OK",
                "HealthRollUp": "OK",
                "State": "Enabled"
        },
        "SystemType": "Physical",
        "UUID": "4c4c4544-0057-3010-8057-b9c04f484b32"
}
`

const testPathDellSystems_SystemEmbedded1_EthernetInterfaces = "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces"

const testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces",
        "@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "Description": "Collection of Ethernet Interfaces for this System",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-3-1"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-4-1"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-1-1"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-2-1"
                }
        ],
        "Members@odata.count": 4,
        "Name": "System Ethernet Interface Collection"
}
`

const testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated131 = "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-3-1"

const testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated131 = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-3-1",
        "@odata.type": "#EthernetInterface.v1_0_2.EthernetInterface",
        "AutoNeg": false,
        "Description": "Integrated NIC 1 Port 3 Partition 1",
        "FQDN": null,
        "FullDuplex": false,
        "HostName": null,
        "IPV6DefaultGateway": null,
        "IPv4Addresses": [],
        "IPv4Addresses@odata.count": 0,
        "IPv6AddressPolicyTable": [],
        "IPv6AddressPolicyTable@odata.count": 0,
        "IPv6Addresses": [],
        "IPv6Addresses@odata.count": 0,
        "IPv6StaticAddresses": [],
        "IPv6StaticAddresses@odata.count": 0,
        "Id": "NIC.Integrated.1-3-1",
        "InterfaceEnabled": null,
        "MTUSize": null,
        "MacAddress": "80:18:44:DE:4F:E6",
        "MaxIPv6StaticAddresses": null,
        "Name": "System Ethernet Interface",
        "NameServers": [],
        "NameServers@odata.count": 0,
        "PermanentMACAddress": "80:18:44:DE:4F:E6",
        "SpeedMbps": 0,
        "Status": {
                "Health": "Ok",
                "State": "StandbyOffline"
        },
        "UefiDevicePath": "PciRoot(0x0)/Pci(0x3,0x0)/Pci(0x0,0x0)",
        "VLAN": null
}
`

const testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated141 = "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-4-1"

const testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated141 = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-4-1",
        "@odata.type": "#EthernetInterface.v1_0_2.EthernetInterface",
        "AutoNeg": false,
        "Description": "Integrated NIC 1 Port 4 Partition 1",
        "FQDN": null,
        "FullDuplex": false,
        "HostName": null,
        "IPV6DefaultGateway": null,
        "IPv4Addresses": [],
        "IPv4Addresses@odata.count": 0,
        "IPv6AddressPolicyTable": [],
        "IPv6AddressPolicyTable@odata.count": 0,
        "IPv6Addresses": [],
        "IPv6Addresses@odata.count": 0,
        "IPv6StaticAddresses": [],
        "IPv6StaticAddresses@odata.count": 0,
        "Id": "NIC.Integrated.1-4-1",
        "InterfaceEnabled": null,
        "MTUSize": null,
        "MacAddress": "80:18:44:DE:4F:E7",
        "MaxIPv6StaticAddresses": null,
        "Name": "System Ethernet Interface",
        "NameServers": [],
        "NameServers@odata.count": 0,
        "PermanentMACAddress": "80:18:44:DE:4F:E7",
        "SpeedMbps": 0,
        "Status": {
                "Health": "Ok",
                "State": "StandbyOffline"
        },
        "UefiDevicePath": "PciRoot(0x0)/Pci(0x3,0x0)/Pci(0x0,0x1)",
        "VLAN": null
}
`

const testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated111 = "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-1-1"

const testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated111 = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-1-1",
        "@odata.type": "#EthernetInterface.v1_0_2.EthernetInterface",
        "AutoNeg": true,
        "Description": "Integrated NIC 1 Port 1 Partition 1",
        "FQDN": null,
        "FullDuplex": true,
        "HostName": null,
        "IPV6DefaultGateway": null,
        "IPv4Addresses": [],
        "IPv4Addresses@odata.count": 0,
        "IPv6AddressPolicyTable": [],
        "IPv6AddressPolicyTable@odata.count": 0,
        "IPv6Addresses": [],
        "IPv6Addresses@odata.count": 0,
        "IPv6StaticAddresses": [],
        "IPv6StaticAddresses@odata.count": 0,
        "Id": "NIC.Integrated.1-1-1",
        "InterfaceEnabled": null,
        "MTUSize": null,
        "MacAddress": "80:18:44:DE:4F:E4",
        "MaxIPv6StaticAddresses": null,
        "Name": "System Ethernet Interface",
        "NameServers": [],
        "NameServers@odata.count": 0,
        "PermanentMACAddress": "80:18:44:DE:4F:E4",
        "SpeedMbps": 1000,
        "Status": {
                "Health": "Ok",
                "State": "Enabled"
        },
        "UefiDevicePath": "PciRoot(0x0)/Pci(0x3,0x1)/Pci(0x0,0x0)",
        "VLAN": null
}
`

const testPathDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated121 = "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-2-1"

const testPayloadDellSystems_SystemEmbedded1_EthernetInterfaces_NICIntegrated121 = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/EthernetInterfaces/NIC.Integrated.1-2-1",
        "@odata.type": "#EthernetInterface.v1_0_2.EthernetInterface",
        "AutoNeg": false,
        "Description": "Integrated NIC 1 Port 2 Partition 1",
        "FQDN": null,
        "FullDuplex": false,
        "HostName": null,
        "IPV6DefaultGateway": null,
        "IPv4Addresses": [],
        "IPv4Addresses@odata.count": 0,
        "IPv6AddressPolicyTable": [],
        "IPv6AddressPolicyTable@odata.count": 0,
        "IPv6Addresses": [],
        "IPv6Addresses@odata.count": 0,
        "IPv6StaticAddresses": [],
        "IPv6StaticAddresses@odata.count": 0,
        "Id": "NIC.Integrated.1-2-1",
        "InterfaceEnabled": null,
        "MTUSize": null,
        "MacAddress": "80:18:44:DE:4F:E5",
        "MaxIPv6StaticAddresses": null,
        "Name": "System Ethernet Interface",
        "NameServers": [],
        "NameServers@odata.count": 0,
        "PermanentMACAddress": "80:18:44:DE:4F:E5",
        "SpeedMbps": 0,
        "Status": {
                "Health": "Ok",
                "State": "StandbyOffline"
        },
        "UefiDevicePath": "PciRoot(0x0)/Pci(0x3,0x1)/Pci(0x0,0x1)",
        "VLAN": null
}
`

const testPathDellSystems_SystemEmbedded1_Processors = "/redfish/v1/Systems/System.Embedded.1/Processors"

const testPayloadDellSystems_SystemEmbedded1_Processors = `
{
        "@odata.context": "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Processors",
        "@odata.type": "#ProcessorCollection.ProcessorCollection",
        "Description": "Collection of Processors for this System",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"
                },
                {
                        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"
                }
        ],
        "Members@odata.count": 2,
        "Name": "ProcessorsCollection"
}
`

const testPathDellSystems_SystemEmbedded1_Processors_CPUSocket1 = "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1"

const testPayloadDellSystems_SystemEmbedded1_Processors_CPUSocket1 = `
{
        "@odata.context": "/redfish/v1/$metadata#Processor.Processor",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.1",
        "@odata.type": "#Processor.v1_0_2.Processor",
        "Description": "Represents the properties of a Processor attached to this System",
        "Id": "CPU.Socket.1",
        "InstructionSet": "x86-64",
        "Manufacturer": "Intel",
        "MaxSpeedMHz": 4000,
        "Model": "Intel(R) Xeon(R) CPU E5-2650 v4 @ 2.20GHz",
        "Name": "CPU 1",
        "ProcessorArchitecture": "x86",
        "ProcessorId": {
                "EffectiveFamily": "6",
                "EffectiveModel": "79",
                "IdentificationRegisters": "0x000406F1",
                "MicrocodeInfo": "0xB00001F",
                "Step": "1",
                "VendorID": "GenuineIntel"
        },
        "ProcessorType": "CPU",
        "Socket": "CPU.Socket.1",
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "TotalCores": 12,
        "TotalThreads": 24
}
`

const testPathDellSystems_SystemEmbedded1_Processors_CPUSocket2 = "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2"

const testPayloadDellSystems_SystemEmbedded1_Processors_CPUSocket2 = `
{
        "@odata.context": "/redfish/v1/$metadata#Processor.Processor",
        "@odata.id": "/redfish/v1/Systems/System.Embedded.1/Processors/CPU.Socket.2",
        "@odata.type": "#Processor.v1_0_2.Processor",
        "Description": "Represents the properties of a Processor attached to this System",
        "Id": "CPU.Socket.2",
        "InstructionSet": [
                {
                        "Member": "x86-64"
                }
        ],
        "Manufacturer": "Intel",
        "MaxSpeedMHz": 4000,
        "Model": "Intel(R) Xeon(R) CPU E5-2650 v4 @ 2.20GHz",
        "Name": "CPU 2",
        "ProcessorArchitecture": [
                {
                        "Member": "x86"
                }
        ],
        "ProcessorId": {
                "EffectiveFamily": "6",
                "EffectiveModel": "79",
                "IdentificationRegisters": "0x000406F1",
                "MicrocodeInfo": "0xB00001F",
                "Step": "1",
                "VendorID": "GenuineIntel"
        },
        "ProcessorType": "CPU",
        "Socket": "CPU.Socket.2",
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "TotalCores": 12,
        "TotalThreads": 24
}
`

//////////////////////////////////////////////////////////////////////////////
//                  Cray Mountain Chassis Controller mock
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncCrayCMM1() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathCrayCMM_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMM_redfish_v1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMAccountService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMAccountService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMSessionService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMSessionService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMEventService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMEventService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMTaskService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMTaskService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade7:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade7)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif5:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif5)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade5:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade5)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif4:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif4)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade6:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade6)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Blade4:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Blade4)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Enclosure:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Enclosure)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Enclosure_Power:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Enclosure_Power)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif6:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif6)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMChassis_Perif7:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMChassis_Perif7)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMManagers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMManagers)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayCMMManagers_BMC:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMManagers_BMC)),
				Header: make(http.Header),
			}
		//case "https://" + testFQDN + testPathCrayCMMSystems:
		//	return &http.Response{
		//		StatusCode: 200,
		//		// Send mock response for rpath
		//		Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayCMMSystems)),
		//		Header: make(http.Header),
		//	}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
				Header: make(http.Header),
			}
		}
	}
}

const testPathCrayCMM_redfish_v1 = "/redfish/v1"

const testPayloadCrayCMM_redfish_v1 = `
{
	"@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
	"@odata.etag": "W/\"1549801610\"",
	"@odata.id": "/redfish/v1/",
	"@odata.type": "#ServiceRoot.v1_2_0.ServiceRoot",
	"AccountService": {
		"@odata.id": "/redfish/v1/AccountService"
	},
	"Chassis": {
		"@odata.id": "/redfish/v1/Chassis"
	},
	"Description": "The service root for all Redfish requests on this host",
	"EventService": {
		"@odata.id": "/redfish/v1/EventService"
	},
	"Id": "RootService",
	"JsonSchemas": {
		"@odata.id": "/redfish/v1/JsonSchemas"
	},
	"Links": {
		"Sessions": {
			"@odata.id": "/redfish/v1/SessionService/Sessions"
		}
	},
	"Managers": {
		"@odata.id": "/redfish/v1/Managers"
	},
	"Name": "Root Service",
	"Oem": {
		"@odata.type": "ServiceRoot.v1_2_0.ServiceRoot",
		"Ami": {
			"Configurations": {
				"@odata.id": "/redfish/v1/configurations"
			}
		}
	},
	"RedfishVersion": "1.2.0",
	"Registries": {
		"@odata.id": "/redfish/v1/Registries"
	},
	"SessionService": {
		"@odata.id": "/redfish/v1/SessionService"
	},
	"Tasks": {
		"@odata.id": "/redfish/v1/TaskService"
	},
	"TelemetryService": {
		"@odata.id": "/redfish/v1/TelemetryService"
	},
	"UpdateService": {
		"@odata.id": "/redfish/v1/UpdateService"
	}
}`

const testPathCrayCMMAccountService = "/redfish/v1/AccountService"

const testPayloadCrayCMMAccountService = `
{
	"@odata.context": "/redfish/v1/$metadata#AccountService.AccountService",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/AccountService",
	"@odata.type": "#AccountService.v1_2_1.AccountService",
	"AccountLockoutCounterResetAfter": 30,
	"AccountLockoutDuration": 30,
	"AccountLockoutThreshold": 5,
	"Accounts": {
		"@odata.id": "/redfish/v1/AccountService/Accounts"
	},
	"AuthFailureLoggingThreshold": 3,
	"Description": "BMC User Accounts",
	"Id": "AccountService",
	"MaxPasswordLength": 12,
	"MinPasswordLength": 8,
	"Name": "Account Service",
	"Oem": {
		"@odata.type": "AccountService.v1_2_1.AccountService",
		"Ami": {
			"Configuration": {
				"@odata.id": "/redfish/v1/AccountService/Configurations"
			}
		}
	},
	"Roles": {
		"@odata.id": "/redfish/v1/AccountService/Roles"
	},
	"ServiceEnabled": true,
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	}
}`

const testPathCrayCMMSessionService = "/redfish/v1/SessionService"

const testPayloadCrayCMMSessionService = `
{
	"@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/SessionService",
	"@odata.type": "#SessionService.v1_1_3.SessionService",
	"Description": "Session Service",
	"Id": "SessionService",
	"Name": "Session Service",
	"ServiceEnabled": true,
	"SessionTimeout": 30,
	"Sessions": {
		"@odata.id": "/redfish/v1/SessionService/Sessions"
	},
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	}
}`

const testPathCrayCMMEventService = "/redfish/v1/EventService"

const testPayloadCrayCMMEventService = `
{
	"@odata.context": "/redfish/v1/$metadata#EventService.EventService",
	"@odata.etag": "W/\"1549491468\"",
	"@odata.id": "/redfish/v1/EventService",
	"@odata.type": "#EventService.v1_0_5.EventService",
	"Actions": {
		"#EventService.SubmitTestEvent": {
			"EventType@Redfish.AllowableValues": [
				"StatusChange",
				"ResourceUpdated",
				"ResourceAdded",
				"ResourceRemoved",
				"Alert"
			],
			"target": "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"
		},
		"Oem": {
			"Ami": {
				"#EventService.SubmitDelayedTestEvent": {
					"EventType@Redfish.AllowableValues": [
						"StatusChange",
						"ResourceUpdated",
						"ResourceAdded",
						"ResourceRemoved",
						"Alert"
					],
					"target": "/redfish/v1/EventService/Actions/EventService.SubmitDelayedTestEvent"
				}
			}
		}
	},
	"DeliveryRetryAttempts": 3,
	"DeliveryRetryIntervalSeconds": 60,
	"Description": "Event Service",
	"EventTypesForSubscription": [
		"StatusChange",
		"ResourceUpdated",
		"ResourceAdded",
		"ResourceRemoved",
		"Alert"
	],
	"Id": "EventService",
	"Name": "Event Service",
	"ServiceEnabled": true,
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	},
	"Subscriptions": {
		"@odata.id": "/redfish/v1/EventService/Subscriptions"
	}
}`

const testPathCrayCMMTaskService = "/redfish/v1/TaskService"

const testPayloadCrayCMMTaskService = `
{
	"@odata.context": "/redfish/v1/$metadata#TaskService.TaskService",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/TaskService",
	"@odata.type": "#TaskService.v1_1_0.TaskService",
	"CompletedTaskOverWritePolicy": "Oldest",
	"DateTime": "2019-02-12T23:59:13Z",
	"Description": "Task Service",
	"Id": "TaskService",
	"LifeCycleEventOnTaskStateChange": true,
	"Name": "Task Service",
	"ServiceEnabled": true,
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	},
	"Tasks": {
		"@odata.id": "/redfish/v1/TaskService/Tasks"
	}
}`

const testPathCrayCMMChassis = "/redfish/v1/Chassis"

const testPayloadCrayCMMChassis = `
{
	"@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
	"@odata.etag": "W/\"1549801610\"",
	"@odata.id": "/redfish/v1/Chassis",
	"@odata.type": "#ChassisCollection.ChassisCollection",
	"Description": "The Collection for Chassis",
	"Members": [
		{
			"@odata.id": "/redfish/v1/Chassis/Blade7"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif0"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Blade5"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif4"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif5"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Enclosure"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif3"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Blade2"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif6"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif1"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Blade3"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Blade0"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Blade6"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif2"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Perif7"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Blade4"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Blade1"
		}
	],
	"Members@odata.count": 17,
	"Name": "Chassis Collection"
}`

const testPathCrayCMMChassis_Blade7 = "/redfish/v1/Chassis/Blade7"

const testPayloadCrayCMMChassis_Blade7 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Blade7",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade7/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade7",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade7",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade7/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Perif5 = "/redfish/v1/Chassis/Perif5"

const testPayloadCrayCMMChassis_Perif5 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif5",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif5/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif5",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif5",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif5/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Blade2 = "/redfish/v1/Chassis/Blade2"

const testPayloadCrayCMMChassis_Blade2 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Blade2",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade2/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade2",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade2",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade2/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Blade5 = "/redfish/v1/Chassis/Blade5"

const testPayloadCrayCMMChassis_Blade5 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549493187\"",
	"@odata.id": "/redfish/v1/Chassis/Blade5",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade5/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade5",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade5",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade5/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Perif4 = "/redfish/v1/Chassis/Perif4"

const testPayloadCrayCMMChassis_Perif4 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif4",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif4/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif4",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif4",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif4/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Blade1 = "/redfish/v1/Chassis/Blade1"

const testPayloadCrayCMMChassis_Blade1 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549493187\"",
	"@odata.id": "/redfish/v1/Chassis/Blade1",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade1/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade1",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade1",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade1/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Blade3 = "/redfish/v1/Chassis/Blade3"

const testPayloadCrayCMMChassis_Blade3 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Blade3",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade3/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade3",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade3",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade3/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Blade6 = "/redfish/v1/Chassis/Blade6"

const testPayloadCrayCMMChassis_Blade6 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549801610\"",
	"@odata.id": "/redfish/v1/Chassis/Blade6",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade6/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade6",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade6",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade6/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Enabled"
	}
}`

const testPathCrayCMMChassis_Perif1 = "/redfish/v1/Chassis/Perif1"

const testPayloadCrayCMMChassis_Perif1 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif1",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif1/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif1",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif1",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif1/NetworkAdapters"
	},
	"PowerState": "On",
	"Status": {
		"State": "Enabled"
	}
}`

const testPathCrayCMMChassis_Blade0 = "/redfish/v1/Chassis/Blade0"

const testPayloadCrayCMMChassis_Blade0 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Blade0",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade0/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade0",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade0",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade0/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Blade4 = "/redfish/v1/Chassis/Blade4"

const testPayloadCrayCMMChassis_Blade4 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Blade4",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Blade4/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Blade4",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Blade4",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Blade4/NetworkAdapters"
	},
	"PowerState": "On",
	"Status": {
		"State": "Enabled"
	}
}`

const testPathCrayCMMChassis_Enclosure = "/redfish/v1/Chassis/Enclosure"

const testPayloadCrayCMMChassis_Enclosure = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Thermal,Id,Status,AssetTag,PowerState,Power,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568702\"",
	"@odata.id": "/redfish/v1/Chassis/Enclosure",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Enclosure",
	"Id": "Enclosure",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Enclosure",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Enclosure/NetworkAdapters"
	},
	"Power": {
		"@odata.id": "/redfish/v1/Chassis/Enclosure/Power"
	},
	"PowerState": "On",
	"Status": {
		"State": "Enabled"
	},
	"Thermal": {
		"@odata.id": "/redfish/v1/Chassis/Enclosure/Thermal"
	}
}`

const testPathCrayCMMChassis_Enclosure_Power = "/redfish/v1/Chassis/Enclosure/Power"

const testPayloadCrayCMMChassis_Enclosure_Power = `
{
  "@odata.context": "/redfish/v1/$metadata#Power.Power(PowerSupplies@odata.count,Id,PowerSupplies,Name,Description)",
  "@odata.etag": "W/\"1585648650\"",
  "@odata.id": "/redfish/v1/Chassis/Enclosure/Power",
  "@odata.type": "#Power.v1_4_0.Power",
  "Description": "Power sensor readings",
  "Id": "Power",
  "Name": "Power",
  "PowerSupplies": [
    {
      "@odata.id": "/redfish/v1/Chassis/Enclosure/Power#/PowerSupplies/200",
      "Manufacturer": "ABB",
      "Model": "CC12500H3C380T",
      "Name": "Rectifier 0",
      "SerialNumber": "LBGEPE18KZ31046456",
      "Status": {
        "State": "Enabled"
      }
    },
    {
      "@odata.id": "/redfish/v1/Chassis/Enclosure/Power#/PowerSupplies/400",
      "Manufacturer": "ABB",
      "Model": "CC12500H3C380T",
      "Name": "Rectifier 2",
      "SerialNumber": "LBGEPE18KZ31046448",
      "Status": {
        "State": "Enabled"
      }
    },
    {
      "@odata.id": "/redfish/v1/Chassis/Enclosure/Power#/PowerSupplies/300",
      "Manufacturer": "ABB",
      "Model": "CC12500H3C380T",
      "Name": "Rectifier 1",
      "SerialNumber": "LBGEPE18KZ31046447",
      "Status": {
        "State": "Enabled"
      }
    }
  ],
  "PowerSupplies@odata.count": 3
}`

const testPathCrayCMMChassis_Perif0 = "/redfish/v1/Chassis/Perif0"

const testPayloadCrayCMMChassis_Perif0 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif0",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif0/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif0",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif0",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif0/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Perif2 = "/redfish/v1/Chassis/Perif2"

const testPayloadCrayCMMChassis_Perif2 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif2",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif2/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif2",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif2",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif2/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Perif3 = "/redfish/v1/Chassis/Perif3"

const testPayloadCrayCMMChassis_Perif3 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif3",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif3/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif3",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif3",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif3/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Perif6 = "/redfish/v1/Chassis/Perif6"

const testPayloadCrayCMMChassis_Perif6 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif6",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif6/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif6",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif6",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif6/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMChassis_Perif7 = "/redfish/v1/Chassis/Perif7"

const testPayloadCrayCMMChassis_Perif7 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Chassis/Perif7",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"On",
				"Off",
				"ForceOff"
			],
			"target": "/redfish/v1/Chassis/Perif7/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Perif7",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Perif7",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Perif7/NetworkAdapters"
	},
	"PowerState": "Off",
	"Status": {
		"State": "Absent"
	}
}`

const testPathCrayCMMManagers = "/redfish/v1/Managers"

const testPayloadCrayCMMManagers = `
{
	"@odata.context": "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
	"@odata.etag": "W/\"0\"",
	"@odata.id": "/redfish/v1/Managers",
	"@odata.type": "#ManagerCollection.ManagerCollection",
	"Description": "The collection for Managers",
	"Members": [
		{
			"@odata.id": "/redfish/v1/Managers/BMC"
		}
	],
	"Members@odata.count": 1,
	"Name": "Manager Collection"
}`

const testPathCrayCMMManagers_BMC = "/redfish/v1/Managers/BMC"

const testPayloadCrayCMMManagers_BMC = `
{
	"@odata.context": "/redfish/v1/$metadata#Manager.Manager(DateTimeLocalOffset,Id,Status,NetworkProtocol,ManagerType,DateTime,Links,Name,LogServices,Description,Actions)",
	"@odata.etag": "W/\"0\"",
	"@odata.id": "/redfish/v1/Managers/BMC",
	"@odata.type": "#Manager.v1_3_2.Manager",
	"Actions": {
		"#Manager.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"ForceRestart",
				"ForceEraseNetworkReload"
			],
			"target": "/redfish/v1/Managers/BMC/Actions/Manager.Reset"
		},
		"Oem": {
			"#CrayProcess.Schedule": {
				"Name@Redfish.AllowableValues": [
					"memtest",
					"cpuburn"
				],
				"target": "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"
			},
			"#Manager.FactoryReset": {
				"FactoryResetType@Redfish.AllowableValues": [
					"ResetAll"
				],
				"target": "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset"
			}
		}
	},
	"DateTime": "2019-02-12T23:59:18Z",
	"DateTimeLocalOffset": "0",
	"Description": "Shasta Manager",
	"Id": "BMC",
	"Links": {
		"ManagerForChassis@odata.count": 0,
		"ManagerForServers@odata.count": 0,
		"ManagerInChassis": {
			"@odata.id": "/redfish/v1/Chassis/Self"
		}
	},
	"LogServices": {
		"@odata.id": "/redfish/v1/Managers/BMC/LogServices"
	},
	"ManagerType": "EnclosureManager",
	"Name": "BMC",
    "Manufacturer": "Cray",
    "SerialNumber": "12345xyz",
	"NetworkProtocol": {
		"@odata.id": "/redfish/v1/Managers/BMC/NetworkProtocol"
	},
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	}
}`

const testPathCrayCMMSystems = "/redfish/v1/Systems"

const testPayloadCrayCMMSystems = `
{
	"@Redfish.CollectionCapabilities": {
		"@odata.type": "#CollectionCapabilities.v1_0_0.CollectionCapabilities",
		"Capabilities": [
			{
				"CapabilitiesObject": {
					"@odata.id": "/redfish/v1/Systems/Capabilities"
				},
				"Links": {
					"RelatedItem": [
						{
							"@odata.id": "/redfish/v1/CompositionService/ResourceZones/1"
						}
					],
					"TargetCollection": {
						"@odata.id": "/redfish/v1/Systems"
					}
				},
				"UseCase": "ComputerSystemComposition"
			}
		]
	},
	"@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
	"@odata.etag": "W/\"1549568748\"",
	"@odata.id": "/redfish/v1/Systems",
	"@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
	"Members@odata.count": 0,
	"Name": "Systems Collection"
}`

//////////////////////////////////////////////////////////////////////////////
//                  Cray Mountain Node Card Controller mock
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncCrayNC1() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathCrayNC_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNC_redfish_v1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCAccountService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCAccountService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSessionService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSessionService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCEventService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCEventService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCTaskService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCTaskService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis_Enclosure:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis_Enclosure)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis_Node0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis_Node0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis_Node1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis_Node1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisPower_Node0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisPower_Node0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisAssembly_Node0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisAssembly_Node0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapters_Node0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapters_Node0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapter0_Node0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapter0_Node0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisPower_Node1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisPower_Node1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisAssembly_Node1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisAssembly_Node1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapters_Node1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapters_Node1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapter0_Node1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapter0_Node1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCManagers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCManagers)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCManagers_BMC:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCManagers_BMC)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_EthernetInterfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_EthernetInterfaces)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Processors:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Processors)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Storage:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Storage)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Storage_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Storage_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Storage_1_Drives_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Storage_1_Drives_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Memory:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Memory)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_EthernetInterfaces:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_EthernetInterfaces)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Processors:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Processors)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Storage:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Storage)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Storage_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Storage_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Storage_1_Drives_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Storage_1_Drives_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Memory:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Memory)),
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
				Header: make(http.Header),
			}
		}
	}
}

const testPathCrayNC_redfish_v1 = "/redfish/v1"

const testPayloadCrayNC_redfish_v1 = `
{
	"@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/",
	"@odata.type": "#ServiceRoot.v1_2_0.ServiceRoot",
	"AccountService": {
		"@odata.id": "/redfish/v1/AccountService"
	},
	"Chassis": {
		"@odata.id": "/redfish/v1/Chassis"
	},
	"Description": "The service root for all Redfish requests on this host",
	"EventService": {
		"@odata.id": "/redfish/v1/EventService"
	},
	"Id": "RootService",
	"JsonSchemas": {
		"@odata.id": "/redfish/v1/JsonSchemas"
	},
	"Links": {
		"Sessions": {
			"@odata.id": "/redfish/v1/SessionService/Sessions"
		}
	},
	"Managers": {
		"@odata.id": "/redfish/v1/Managers"
	},
	"Name": "Root Service",
	"Oem": {
		"@odata.type": "ServiceRoot.v1_2_0.ServiceRoot",
		"Ami": {
			"Configurations": {
				"@odata.id": "/redfish/v1/configurations"
			}
		}
	},
	"RedfishVersion": "1.2.0",
	"Registries": {
		"@odata.id": "/redfish/v1/Registries"
	},
	"SessionService": {
		"@odata.id": "/redfish/v1/SessionService"
	},
	"Systems": {
		"@odata.id": "/redfish/v1/Systems"
	},
	"Tasks": {
		"@odata.id": "/redfish/v1/TaskService"
	},
	"TelemetryService": {
		"@odata.id": "/redfish/v1/TelemetryService"
	},
	"UpdateService": {
		"@odata.id": "/redfish/v1/UpdateService"
	}
}`

const testPathCrayNCAccountService = "/redfish/v1/AccountService"

const testPayloadCrayNCAccountService = `
{
	"@odata.context": "/redfish/v1/$metadata#AccountService.AccountService",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/AccountService",
	"@odata.type": "#AccountService.v1_2_1.AccountService",
	"AccountLockoutCounterResetAfter": 30,
	"AccountLockoutDuration": 30,
	"AccountLockoutThreshold": 5,
	"Accounts": {
		"@odata.id": "/redfish/v1/AccountService/Accounts"
	},
	"AuthFailureLoggingThreshold": 3,
	"Description": "BMC User Accounts",
	"Id": "AccountService",
	"MaxPasswordLength": 12,
	"MinPasswordLength": 8,
	"Name": "Account Service",
	"Oem": {
		"@odata.type": "AccountService.v1_2_1.AccountService",
		"Ami": {
			"Configuration": {
				"@odata.id": "/redfish/v1/AccountService/Configurations"
			}
		}
	},
	"Roles": {
		"@odata.id": "/redfish/v1/AccountService/Roles"
	},
	"ServiceEnabled": true,
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	}
}`

const testPathCrayNCSessionService = "/redfish/v1/SessionService"

const testPayloadCrayNCSessionService = `
{
	"@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/SessionService",
	"@odata.type": "#SessionService.v1_1_3.SessionService",
	"Description": "Session Service",
	"Id": "SessionService",
	"Name": "Session Service",
	"ServiceEnabled": true,
	"SessionTimeout": 30,
	"Sessions": {
		"@odata.id": "/redfish/v1/SessionService/Sessions"
	},
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	}
}`

const testPathCrayNCEventService = "/redfish/v1/EventService"

const testPayloadCrayNCEventService = `
{
	"@odata.context": "/redfish/v1/$metadata#EventService.EventService",
	"@odata.etag": "W/\"1549568836\"",
	"@odata.id": "/redfish/v1/EventService",
	"@odata.type": "#EventService.v1_0_5.EventService",
	"Actions": {
		"#EventService.SubmitTestEvent": {
			"EventType@Redfish.AllowableValues": [
				"StatusChange",
				"ResourceUpdated",
				"ResourceAdded",
				"ResourceRemoved",
				"Alert"
			],
			"target": "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"
		},
		"Oem": {
			"Ami": {
				"#EventService.SubmitDelayedTestEvent": {
					"EventType@Redfish.AllowableValues": [
						"StatusChange",
						"ResourceUpdated",
						"ResourceAdded",
						"ResourceRemoved",
						"Alert"
					],
					"target": "/redfish/v1/EventService/Actions/EventService.SubmitDelayedTestEvent"
				}
			}
		}
	},
	"DeliveryRetryAttempts": 3,
	"DeliveryRetryIntervalSeconds": 60,
	"Description": "Event Service",
	"EventTypesForSubscription": [
		"StatusChange",
		"ResourceUpdated",
		"ResourceAdded",
		"ResourceRemoved",
		"Alert"
	],
	"Id": "EventService",
	"Name": "Event Service",
	"ServiceEnabled": true,
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	},
	"Subscriptions": {
		"@odata.id": "/redfish/v1/EventService/Subscriptions"
	}
}`

const testPathCrayNCTaskService = "/redfish/v1/TaskService"

const testPayloadCrayNCTaskService = `
{
	"@odata.context": "/redfish/v1/$metadata#TaskService.TaskService",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/TaskService",
	"@odata.type": "#TaskService.v1_1_0.TaskService",
	"CompletedTaskOverWritePolicy": "Oldest",
	"DateTime": "2019-02-13T00:44:47Z",
	"Description": "Task Service",
	"Id": "TaskService",
	"LifeCycleEventOnTaskStateChange": true,
	"Name": "Task Service",
	"ServiceEnabled": true,
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	},
	"Tasks": {
		"@odata.id": "/redfish/v1/TaskService/Tasks"
	}
}`

const testPathCrayNCChassis = "/redfish/v1/Chassis"

const testPayloadCrayNCChassis = `
{
	"@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Chassis",
	"@odata.type": "#ChassisCollection.ChassisCollection",
	"Description": "The Collection for Chassis",
	"Members": [
		{
			"@odata.id": "/redfish/v1/Chassis/Node1"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Node0"
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Enclosure"
		}
	],
	"Members@odata.count": 3,
	"Name": "Chassis Collection"
}`

const testPathCrayNCChassis_Enclosure = "/redfish/v1/Chassis/Enclosure"

const testPayloadCrayNCChassis_Enclosure = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(SerialNumber,Id,Status,AssetTag,Name,Links,NetworkAdapters,ChassisType,Manufacturer,Actions)",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Chassis/Enclosure",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [],
			"target": "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Enclosure",
	"Id": "Enclosure",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Enclosure",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Enclosure/NetworkAdapters"
	},
	"SerialNumber": "HA18340007",
	"Status": {
		"State": "Enabled"
	}
}`

const testPathCrayNCChassis_Node0 = "/redfish/v1/Chassis/Node0"

const testPayloadCrayNCChassis_Node0 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Thermal,Id,Status,AssetTag,Name,Power,Links,NetworkAdapters,ChassisType,Manufacturer,Actions)",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Chassis/Node0",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [],
			"target": "/redfish/v1/Chassis/Node0/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Node0",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Node0",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters"
	},
	"Power": {
		"@odata.id": "/redfish/v1/Chassis/Node0/Power"
	},
	"Status": {
		"State": "Enabled"
	},
	"Thermal": {
		"@odata.id": "/redfish/v1/Chassis/Node0/Thermal"
	},
	"Assembly": {
		"@odata.id": "/redfish/v1/Chassis/Node0/Assembly"
	}
}`

const testPathCrayNCChassis_Node1 = "/redfish/v1/Chassis/Node1"

const testPayloadCrayNCChassis_Node1 = `
{
	"@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Thermal,Id,Status,AssetTag,Name,Power,Links,NetworkAdapters,ChassisType,Manufacturer,Actions)",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Chassis/Node1",
	"@odata.type": "#Chassis.v1_5_1.Chassis",
	"Actions": {
		"#Chassis.Reset": {
			"ResetType@Redfish.AllowableValues": [],
			"target": "/redfish/v1/Chassis/Node1/Actions/Chassis.Reset"
		}
	},
	"AssetTag": "",
	"ChassisType": "Blade",
	"Id": "Node1",
	"Links": {
		"ComputerSystems@odata.count": 0,
		"Contains@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Drives@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"ManagersInChassis@odata.count": 0,
		"PCIeDevices@odata.count": 0,
		"PoweredBy@odata.count": 0,
		"ResourceBlock@odata.count": 0,
		"Storage@odata.count": 0
	},
	"Manufacturer": "Cray Inc",
	"Name": "Node1",
	"NetworkAdapters": {
		"@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters"
	},
	"Power": {
		"@odata.id": "/redfish/v1/Chassis/Node1/Power"
	},
	"Status": {
		"State": "Enabled"
	},
	"Thermal": {
		"@odata.id": "/redfish/v1/Chassis/Node1/Thermal"
	},
	"Assembly": {
		"@odata.id": "/redfish/v1/Chassis/Node1/Assembly"
	}
}`

const testPathCrayNCChassisPower_Node0 = "/redfish/v1/Chassis/Node0/Power"

const testPayloadCrayNCChassisPower_Node0 = `
{
    "@odata.context": "/redfish/v1/$metadata#Power.Power(Voltages,Id,Voltages@odata.count,Name,Description)",
    "@odata.etag": "W/\"1569785935\"",
    "@odata.id": "/redfish/v1/Chassis/Node0/Power",
    "@odata.type": "#Power.v1_4_0.Power",
    "Description": "Power sensor readings",
    "Id": "Power",
    "Name": "Power",
	"PowerControl": [
        {
            "RelatedItem@odata.count": 1,
            "PowerCapacityWatts": 900,
            "Name": "Node Power Control",
            "Oem": {
                "Cray": {
                    "PowerAllocatedWatts": 900,
                    "PowerIdleWatts": 250,
                    "PowerLimit": {
                        "Min": 350,
                        "Max": 850,
                        "Factor": 1.02
                    },
                    "PowerFloorTargetWatts": 0,
                    "PowerResetWatts": 250
                }
            },
            "@odata.id": "/redfish/v1/Chassis/Node0/Power#/PowerControl/Node",
            "PowerLimit": {
                "LimitException": "LogEventOnly",
                "CorrectionInMs": 6000,
                "LimitInWatts": 500
            },
            "RelatedItem": [
                {
                    "@odata.id": "/redfish/v1/Chassis/Node0/Power#/PowerControl/Accelerator0"
                }
            ]
        },
        {
            "RelatedItem@odata.count": 0,
            "Name": "Accelerator0 Power Control",
            "Oem": {
                "Cray": {
                    "PowerIdleWatts": 100,
                    "PowerLimit": {
                        "Min": 200,
                        "Max": 350,
                        "Factor": 1.0
                    },
                    "PowerFloorTargetWatts": 0
                }
            },
            "@odata.id": "/redfish/v1/Chassis/Node0/Power#/PowerControl/Accelerator0",
            "PowerLimit": {
                "LimitException": "LogEventOnly",
                "CorrectionInMs": 6000,
                "LimitInWatts": 300
            }
        }
    ],
    "Voltages": [
        {
            "@odata.id": "/redfish/v1/Chassis/Node0/Power#/Voltages/153",
            "Name": "VDD_1V2_S0 Voltage",
            "PhysicalContext": "VoltageRegulator",
            "ReadingVolts": 1.219,
            "RelatedItem@odata.count": 0
        },
        {
            "@odata.id": "/redfish/v1/Chassis/Node0/Power#/Voltages/158",
            "Name": "CPUS VDD_3V3_S0 Voltage Output",
            "PhysicalContext": "VoltageRegulator",
            "ReadingVolts": 3.371,
            "RelatedItem@odata.count": 0
        }
    ],
    "Voltages@odata.count": 2
}`

const testPathCrayNCChassisAssembly_Node0 = "/redfish/v1/Chassis/Node0/Assembly"

const testPayloadCrayNCChassisAssembly_Node0 = `
{
   "@odata.context": "/redfish/v1/$metadata",
   "@odata.etag": "W/\"1605638544\"",
   "@odata.id": "/redfish/v1/Chassis/Node0/Assembly",
   "@odata.type": "#Assembly.v1_3_0.Assembly",
   "Assemblies": [
     {
       "@odata.id": "/redfish/v1/Chassis/Node0/Assembly#/Assemblies/0",
       "Description": "The Nvidia baseboard assembly that houses the node's GPUs.",
       "Producer": "NVIDIA",
       "MemberId": "0",
       "Model": "NVIDIA HGX A100 4 GPU 40",
       "Name": "NVIDIA Redstone",
       "Oem": {
         "PCBSerialNumber": "1572820530361"
       },
       "PartNumber": "935-22687-3830-000",
       "PhysicalContext": "GPUSubsystem",
       "ProductionDate": "2020/07/18-17:28:00",
       "SerialNumber": "1572920001914",
       "Version": "538980400",
       "EngineeringChangeLevel": "538980400"
     }
   ],
   "Id": "Assembly",
   "Name": "System-related Assembly data"
 }`

const testPathCrayNCChassisNetworkAdapters_Node0 = "/redfish/v1/Chassis/Node0/NetworkAdapters"

const testPayloadCrayNCChassisNetworkAdapters_Node0 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapterCollection.NetworkAdapterCollection",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters",
  "@odata.type": "#NetworkAdapterCollection.NetworkAdapterCollection",
  "Description": "The Collection of Network Adapters",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0"
    }
  ],
  "Members@odata.count": 1,
  "Name": "NetworkAdapter Collection"
}`

const testPathCrayNCChassisNetworkAdapter0_Node0 = "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0"

const testPayloadCrayNCChassisNetworkAdapter0_Node0 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapter.NetworkAdapter(SerialNumber,Id,NetworkDeviceFunctions,NetworkPorts,PartNumber,Manufacturer,Model,Description,Actions)",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0",
  "@odata.type": "#NetworkAdapter.v1_0_1.NetworkAdapter",
  "Actions": {
    "#NetworkAdapter.ResetSettingsToDefault": {
      "ResetSettingsToDefaultType@Redfish.AllowableValues": [],
      "target": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0/Actions/NetworkAdapter.ResetSettingsToDefault"
    }
  },
  "Description": "Shasta Timms NMC REV04 (HSN)",
  "Id": "HPCNet0",
  "Manufacturer": "Mellanox Technologies, Ltd.",
  "Model": "ConnectX-5 100Gb/s",
  "NetworkDeviceFunctions": {
    "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0/NetworkDeviceFunctions"
  },
  "NetworkPorts": {
    "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0/NetworkPorts"
  },
  "PartNumber": "102005303",
  "SerialNumber": "HG19501557"
}`

const testPathCrayNCChassisPower_Node1 = "/redfish/v1/Chassis/Node1/Power"

const testPayloadCrayNCChassisPower_Node1 = `
{
    "@odata.context": "/redfish/v1/$metadata#Power.Power(Voltages,Id,Voltages@odata.count,Name,Description)",
    "@odata.etag": "W/\"1569785935\"",
    "@odata.id": "/redfish/v1/Chassis/Node1/Power",
    "@odata.type": "#Power.v1_4_0.Power",
    "Description": "Power sensor readings",
    "Id": "Power",
    "Name": "Power",
	"PowerControl": [
        {
            "RelatedItem@odata.count": 1,
            "PowerCapacityWatts": 900,
            "Name": "Node Power Control",
            "Oem": {
                "Cray": {
                    "PowerAllocatedWatts": 900,
                    "PowerIdleWatts": 250,
                    "PowerLimit": {
                        "Min": 350,
                        "Max": 850,
                        "Factor": 1.02
                    },
                    "PowerFloorTargetWatts": 0,
                    "PowerResetWatts": 250
                }
            },
            "@odata.id": "/redfish/v1/Chassis/Node1/Power#/PowerControl/Node",
            "PowerLimit": {
                "LimitException": "LogEventOnly",
                "CorrectionInMs": 6000,
                "LimitInWatts": 500
            },
            "RelatedItem": [
                {
                    "@odata.id": "/redfish/v1/Chassis/Node1/Power#/PowerControl/Accelerator0"
                }
            ]
        },
        {
            "RelatedItem@odata.count": 0,
            "Name": "Accelerator0 Power Control",
            "Oem": {
                "Cray": {
                    "PowerIdleWatts": 100,
                    "PowerLimit": {
                        "Min": 200,
                        "Max": 350,
                        "Factor": 1.0
                    },
                    "PowerFloorTargetWatts": 0
                }
            },
            "@odata.id": "/redfish/v1/Chassis/Node1/Power#/PowerControl/Accelerator0",
            "PowerLimit": {
                "LimitException": "LogEventOnly",
                "CorrectionInMs": 6000,
                "LimitInWatts": 300
            }
        }
    ],
    "Voltages": [
        {
            "@odata.id": "/redfish/v1/Chassis/Node1/Power#/Voltages/153",
            "Name": "VDD_1V2_S0 Voltage",
            "PhysicalContext": "VoltageRegulator",
            "ReadingVolts": 1.219,
            "RelatedItem@odata.count": 0
        },
        {
            "@odata.id": "/redfish/v1/Chassis/Node1/Power#/Voltages/116",
            "Name": "CPU0 DIMM VDD_VPP_EFGH_S0 Voltage Input",
            "PhysicalContext": "VoltageRegulator",
            "ReadingVolts": 11.969,
            "RelatedItem@odata.count": 0
        }
    ],
    "Voltages@odata.count": 2
}`

const testPathCrayNCChassisAssembly_Node1 = "/redfish/v1/Chassis/Node1/Assembly"

const testPayloadCrayNCChassisAssembly_Node1 = `
{
   "@odata.context": "/redfish/v1/$metadata",
   "@odata.etag": "W/\"1605638544\"",
   "@odata.id": "/redfish/v1/Chassis/Node1/Assembly",
   "@odata.type": "#Assembly.v1_3_0.Assembly",
   "Assemblies": [
     {
       "@odata.id": "/redfish/v1/Chassis/Node1/Assembly#/Assemblies/0",
       "Description": "The Nvidia baseboard assembly that houses the node's GPUs.",
       "Producer": "NVIDIA",
       "MemberId": "0",
       "Model": "NVIDIA HGX A100 4 GPU 40",
       "Name": "NVIDIA Redstone",
       "Oem": {
         "PCBSerialNumber": "1572820530362"
       },
       "PartNumber": "935-22687-3830-000",
       "PhysicalContext": "GPUSubsystem",
       "ProductionDate": "2020/07/18-17:28:00",
       "SerialNumber": "1572920001915",
       "Version": "538980400",
       "EngineeringChangeLevel": "538980400"
     },
     {
       "@odata.id": "/redfish/v1/Chassis/Node1/Assembly#/Assemblies/1",
       "Description": "Arbitrary assembly object for test purposes",
       "MemberId": "1",
       "Model": "",
       "Name": "",
       "PartNumber": "12345",
       "Producer": "Acme",
       "PhysicalContext": "Junk",
       "ProductionDate": "",
       "SerialNumber": "missing",
       "Version": ""
     },
     {
       "@odata.id": "/redfish/v1/Chassis/Node1/Assembly#/Assemblies/2",
       "Description": "The Nvidia baseboard assembly that houses the node's GPUs.",
       "Manufacturer": "NVIDIA",
       "MemberId": "2",
       "Model": "NVIDIA HGX A100 4 GPU 40",
       "Name": "NVIDIA Redstone",
       "Oem": {
         "PCBSerialNumber": "1572820530363"
       },
       "PartNumber": "935-22687-3830-000",
       "PhysicalContext": "GPUSubsystem",
       "ProductionDate": "2020/07/18-17:28:00",
       "SerialNumber": "1572920001916",
       "Version": "538980400"
     }
   ],
   "Id": "Assembly",
   "Name": "System-related Assembly data"
 }`

const testPathCrayNCChassisNetworkAdapters_Node1 = "/redfish/v1/Chassis/Node1/NetworkAdapters"

const testPayloadCrayNCChassisNetworkAdapters_Node1 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapterCollection.NetworkAdapterCollection",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters",
  "@odata.type": "#NetworkAdapterCollection.NetworkAdapterCollection",
  "Description": "The Collection of Network Adapters",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0"
    }
  ],
  "Members@odata.count": 1,
  "Name": "NetworkAdapter Collection"
}`

const testPathCrayNCChassisNetworkAdapter0_Node1 = "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0"

const testPayloadCrayNCChassisNetworkAdapter0_Node1 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapter.NetworkAdapter(SerialNumber,Id,NetworkDeviceFunctions,NetworkPorts,PartNumber,Manufacturer,Model,Description,Actions)",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0",
  "@odata.type": "#NetworkAdapter.v1_0_1.NetworkAdapter",
  "Actions": {
    "#NetworkAdapter.ResetSettingsToDefault": {
      "ResetSettingsToDefaultType@Redfish.AllowableValues": [],
      "target": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0/Actions/NetworkAdapter.ResetSettingsToDefault"
    }
  },
  "Description": "Shasta Timms NMC REV04 (HSN)",
  "Id": "HPCNet0",
  "Manufacturer": "Mellanox Technologies, Ltd.",
  "Model": "ConnectX-5 100Gb/s",
  "NetworkDeviceFunctions": {
    "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0/NetworkDeviceFunctions"
  },
  "NetworkPorts": {
    "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0/NetworkPorts"
  },
  "PartNumber": "102005303",
  "SerialNumber": "HG19501557"
}`

const testPathCrayNCManagers = "/redfish/v1/Managers"

const testPayloadCrayNCManagers = `
{
	"@odata.context": "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
	"@odata.etag": "W/\"0\"",
	"@odata.id": "/redfish/v1/Managers",
	"@odata.type": "#ManagerCollection.ManagerCollection",
	"Description": "The collection for Managers",
	"Members": [
		{
			"@odata.id": "/redfish/v1/Managers/BMC"
		}
	],
	"Members@odata.count": 1,
	"Name": "Manager Collection"
}`

const testPathCrayNCManagers_BMC = "/redfish/v1/Managers/BMC"

const testPayloadCrayNCManagers_BMC = `
{
	"@odata.context": "/redfish/v1/$metadata#Manager.Manager(DateTimeLocalOffset,Id,Status,NetworkProtocol,ManagerType,DateTime,Links,Name,LogServices,Description,Actions)",
	"@odata.etag": "W/\"0\"",
	"@odata.id": "/redfish/v1/Managers/BMC",
	"@odata.type": "#Manager.v1_3_2.Manager",
	"Actions": {
		"#Manager.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"ForceRestart",
				"ForceEraseNetworkReload"
			],
			"target": "/redfish/v1/Managers/BMC/Actions/Manager.Reset"
		},
		"Oem": {
			"#CrayProcess.Schedule": {
				"Name@Redfish.AllowableValues": [
					"memtest",
					"cpuburn"
				],
				"target": "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"
			},
			"#Manager.FactoryReset": {
				"FactoryResetType@Redfish.AllowableValues": [
					"ResetAll"
				],
				"target": "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset"
			}
		}
	},
	"DateTime": "2019-02-13T00:44:49Z",
	"DateTimeLocalOffset": "0",
	"Description": "Shasta Manager",
	"Id": "BMC",
    "Manufacturer": "Cray",
    "SerialNumber": "12345xyz",
	"Links": {
		"ManagerForChassis@odata.count": 0,
		"ManagerForServers@odata.count": 0,
		"ManagerInChassis": {
			"@odata.id": "/redfish/v1/Chassis/Self"
		}
	},
	"LogServices": {
		"@odata.id": "/redfish/v1/Managers/BMC/LogServices"
	},
	"ManagerType": "EnclosureManager",
	"Name": "BMC",
	"NetworkProtocol": {
		"@odata.id": "/redfish/v1/Managers/BMC/NetworkProtocol"
	},
	"Status": {
		"Health": "OK",
		"State": "Enabled"
	}
}`

const testPathCrayNCSystems = "/redfish/v1/Systems"

const testPayloadCrayNCSystems = `
{
	"@Redfish.CollectionCapabilities": {
		"@odata.type": "#CollectionCapabilities.v1_0_0.CollectionCapabilities",
		"Capabilities": [
			{
				"CapabilitiesObject": {
					"@odata.id": "/redfish/v1/Systems/Capabilities"
				},
				"Links": {
					"RelatedItem": [
						{
							"@odata.id": "/redfish/v1/CompositionService/ResourceZones/1"
						}
					],
					"TargetCollection": {
						"@odata.id": "/redfish/v1/Systems"
					}
				},
				"UseCase": "ComputerSystemComposition"
			}
		]
	},
	"@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems",
	"@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
	"Members": [
		{
			"@odata.id": "/redfish/v1/Systems/Node1"
		},
		{
			"@odata.id": "/redfish/v1/Systems/Node0"
		}
	],
	"Members@odata.count": 2,
	"Name": "Systems Collection"
}`

const testPathCrayNCSystems_Node0 = "/redfish/v1/Systems/Node0"

const testPayloadCrayNCSystems_Node0 = `
{
	"@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem(*)",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node0",
	"@odata.type": "#ComputerSystem.v1_4_1.ComputerSystem",
	"Actions": {
		"#ComputerSystem.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"Off",
				"ForceOff",
				"On"
			],
			"target": "/redfish/v1/Systems/Node0/Actions/ComputerSystem.Reset"
		}
	},
	"BiosVersion": "Unknown",
	"Boot": {
		"BootSourceOverrideEnabled": "Disabled",
		"BootSourceOverrideEnabled@Redfish.AllowableValues": [
			"Disabled",
			"Once",
			"Continuous"
		],
		"BootSourceOverrideTarget": "None",
		"BootSourceOverrideTarget@Redfish.AllowableValues": [
			"None",
			"Pxe",
			"Floppy",
			"Cd",
			"Usb",
			"Hdd",
			"BiosSetup",
			"Utilities",
			"Diags",
			"UefiShell",
			"UefiTarget",
			"SDCard",
			"UefiHttp",
			"RemoteDrive"
		]
	},
	"Description": "Node",
	"EthernetInterfaces": {
		"@odata.id": "/redfish/v1/Systems/Node0/EthernetInterfaces"
	},
	"Id": "Node0",
	"Links": {
		"Chassis@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Endpoints@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"PoweredBy@odata.count": 0
	},
	"LogServices": {
		"@odata.id": "/redfish/v1/Systems/Node0/LogServices"
	},
	"Manufacturer": "Cray Inc",
	"Memory": {
		"@odata.id": "/redfish/v1/Systems/Node0/Memory"
	},
	"Name": "Node0",
	"NetworkInterfaces": {
		"@odata.id": "/redfish/v1/Systems/Node0/NetworkInterfaces"
	},
	"PCIeDevices@odata.count": 0,
	"PCIeFunctions@odata.count": 0,
	"PowerState": "On",
	"Processors": {
		"@odata.id": "/redfish/v1/Systems/Node0/Processors"
	},
	"SecureBoot": {
		"@odata.id": "/redfish/v1/Systems/Node0/SecureBoot"
	},
	"SimpleStorage": {
		"@odata.id": "/redfish/v1/Systems/Node0/SimpleStorage"
	},
	"Storage": {
		"@odata.id": "/redfish/v1/Systems/Node0/Storage"
	},
	"SystemType": "Physical"
}`

const testPathCrayNCSystems_Node0_EthernetInterfaces = "/redfish/v1/Systems/Node0/EthernetInterfaces"

const testPayloadCrayNCSystems_Node0_EthernetInterfaces = `
{
	"@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node0/EthernetInterfaces",
	"@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
	"Description": "Collection of ethernet interfaces for this system",
	"Members": [
		{
			"@odata.id": "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet"
		}
	],
	"Members@odata.count": 1,
	"Name": "Ethernet Interface Collection"
}`

const testPathCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet = "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet"

const testPayloadCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet = `
{
	"@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface(Id,MACAddress,Links,VLANs,PermanentMACAddress,Description)",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet",
	"@odata.type": "#EthernetInterface.v1_3_0.EthernetInterface",
	"Description": "Node Maintenance Network",
	"Id": "ManagementEthernet",
	"Links": {
		"Chassis": {
			"@odata.id": "/redfish/v1/Chassis/Self"
		},
		"Endpoints@odata.count": 0
	},
	"MACAddress": "00:40:a6:82:f6:0a",
	"PermanentMACAddress": "00:40:a6:82:f6:0a",
	"VLANs": {
		"@odata.id": "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet/VLANs"
	}
}`

const testPathCrayNCSystems_Node0_Processors = "/redfish/v1/Systems/Node0/Processors"

const testPayloadCrayNCSystems_Node0_Processors = `
{
	"@odata.context": "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node0/Processors",
	"@odata.type": "#ProcessorCollection.ProcessorCollection",
	"Description": "Collection of processors for this system",
	"Members": [],
	"Members@odata.count": 0,
	"Name": "Processors Collection"
}`

const testPathCrayNCSystems_Node0_Storage = "/redfish/v1/Systems/Node0/Storage"
const testPayloadCrayNCSystems_Node0_Storage = `
{
    "@odata.context": "/redfish/v1/$metadata#StorageCollection.StorageCollection",
    "@odata.id": "/redfish/v1/Systems/Node0/Storage",
    "@odata.type": "#StorageCollection.StorageCollection",
    "Name": "Storage Collection",
    "Members@odata.count": 1,
    "Members": [
        {
            "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1"
        }
    ],
    "Description" : "Collection of Storage resource instances",
    "@odata.etag" : "W/\"1579629364\""
}
`
const testPathCrayNCSystems_Node0_Storage_1 = "/redfish/v1/Systems/Node0/Storage/1"
const testPayloadCrayNCSystems_Node0_Storage_1 = `{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1",
   "@odata.type" : "#Storage.v1_5_0.Storage",
   "@odata.context" : "/redfish/v1/$metadata#Storage.Storage",
   "Description" : "This resource shall be used to represent resources that represent a storage subsystem in the Redfish specification.",
   "Name" : "Storage",
   "Drives" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1/Drives/1"
      }
   ],
   "Drives@odata.count" : 1,
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node0_Storage_1_Drives_1 = "/redfish/v1/Systems/Node0/Storage/1/Drives/1"
const testPayloadCrayNCSystems_Node0_Storage_1_Drives_1 = `
{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1/Drives/1",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400233      ",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node0_Memory = "/redfish/v1/Systems/Node0/Memory"

const testPayloadCrayNCSystems_Node0_Memory = `
{
	"@odata.context": "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node0/Memory",
	"@odata.type": "#MemoryCollection.MemoryCollection",
	"Description": "Collection of Memories for this system",
	"Members": [],
	"Members@odata.count": 0,
	"Name": "Memory Collection"
}`

const testPathCrayNCSystems_Node1 = "/redfish/v1/Systems/Node1"

const testPayloadCrayNCSystems_Node1 = `
{
	"@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem(*)",
	"@odata.etag": "W/\"1549568845\"",
	"@odata.id": "/redfish/v1/Systems/Node1",
	"@odata.type": "#ComputerSystem.v1_4_1.ComputerSystem",
	"Actions": {
		"#ComputerSystem.Reset": {
			"ResetType@Redfish.AllowableValues": [
				"Off",
				"ForceOff",
				"On"
			],
			"target": "/redfish/v1/Systems/Node1/Actions/ComputerSystem.Reset"
		}
	},
	"BiosVersion": "Unknown",
	"Boot": {
		"BootSourceOverrideEnabled": "Disabled",
		"BootSourceOverrideEnabled@Redfish.AllowableValues": [
			"Disabled",
			"Once",
			"Continuous"
		],
		"BootSourceOverrideTarget": "None",
		"BootSourceOverrideTarget@Redfish.AllowableValues": [
			"None",
			"Pxe",
			"Floppy",
			"Cd",
			"Usb",
			"Hdd",
			"BiosSetup",
			"Utilities",
			"Diags",
			"UefiShell",
			"UefiTarget",
			"SDCard",
			"UefiHttp",
			"RemoteDrive"
		]
	},
	"Description": "Node",
	"EthernetInterfaces": {
		"@odata.id": "/redfish/v1/Systems/Node1/EthernetInterfaces"
	},
	"Id": "Node1",
	"Links": {
		"Chassis@odata.count": 0,
		"CooledBy@odata.count": 0,
		"Endpoints@odata.count": 0,
		"ManagedBy@odata.count": 0,
		"PoweredBy@odata.count": 0
	},
	"LogServices": {
		"@odata.id": "/redfish/v1/Systems/Node1/LogServices"
	},
	"Manufacturer": "Cray Inc",
	"Memory": {
		"@odata.id": "/redfish/v1/Systems/Node1/Memory"
	},
	"Name": "Node1",
	"NetworkInterfaces": {
		"@odata.id": "/redfish/v1/Systems/Node1/NetworkInterfaces"
	},
	"PCIeDevices@odata.count": 0,
	"PCIeFunctions@odata.count": 0,
	"PowerState": "On",
	"Processors": {
		"@odata.id": "/redfish/v1/Systems/Node1/Processors"
	},
	"SecureBoot": {
		"@odata.id": "/redfish/v1/Systems/Node1/SecureBoot"
	},
	"SimpleStorage": {
		"@odata.id": "/redfish/v1/Systems/Node1/SimpleStorage"
	},
	"Storage": {
		"@odata.id": "/redfish/v1/Systems/Node1/Storage"
	},
	"SystemType": "Physical"
}`

const testPathCrayNCSystems_Node1_EthernetInterfaces = "/redfish/v1/Systems/Node1/EthernetInterfaces"

const testPayloadCrayNCSystems_Node1_EthernetInterfaces = `
{
	"@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node1/EthernetInterfaces",
	"@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
	"Description": "Collection of ethernet interfaces for this system",
	"Members": [
		{
			"@odata.id": "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet"
		}
	],
	"Members@odata.count": 1,
	"Name": "Ethernet Interface Collection"
}`

const testPathCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet = "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet"

const testPayloadCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet = `
{
	"@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface(Id,MACAddress,Links,VLANs,PermanentMACAddress,Description)",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet",
	"@odata.type": "#EthernetInterface.v1_3_0.EthernetInterface",
	"Description": "Node Maintenance Network",
	"Id": "ManagementEthernet",
	"Links": {
		"Chassis": {
			"@odata.id": "/redfish/v1/Chassis/Self"
		},
		"Endpoints@odata.count": 0
	},
	"MACAddress": "00:40:a6:82:f6:0b",
	"PermanentMACAddress": "00:40:a6:82:f6:0b",
	"VLANs": {
		"@odata.id": "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet/VLANs"
	}
}`

const testPathCrayNCSystems_Node1_Processors = "/redfish/v1/Systems/Node1/Processors"

const testPayloadCrayNCSystems_Node1_Processors = `
{
	"@odata.context": "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node1/Processors",
	"@odata.type": "#ProcessorCollection.ProcessorCollection",
	"Description": "Collection of processors for this system",
	"Members": [],
	"Members@odata.count": 0,
	"Name": "Processors Collection"
}`

const testPathCrayNCSystems_Node1_Storage = "/redfish/v1/Systems/Node1/Storage"
const testPayloadCrayNCSystems_Node1_Storage = `
{
    "@odata.context": "/redfish/v1/$metadata#StorageCollection.StorageCollection",
    "@odata.id": "/redfish/v1/Systems/Node1/Storage",
    "@odata.type": "#StorageCollection.StorageCollection",
    "Name": "Storage Collection",
    "Members@odata.count": 1,
    "Members": [
        {
            "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1"
        }
    ],
    "Description" : "Collection of Storage resource instances",
    "@odata.etag" : "W/\"1579629364\""
}
`
const testPathCrayNCSystems_Node1_Storage_1 = "/redfish/v1/Systems/Node1/Storage/1"
const testPayloadCrayNCSystems_Node1_Storage_1 = `{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1",
   "@odata.type" : "#Storage.v1_5_0.Storage",
   "@odata.context" : "/redfish/v1/$metadata#Storage.Storage",
   "Description" : "This resource shall be used to represent resources that represent a storage subsystem in the Redfish specification.",
   "Name" : "Storage",
   "Drives" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1/Drives/1"
      }
   ],
   "Drives@odata.count" : 1,
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node1_Storage_1_Drives_1 = "/redfish/v1/Systems/Node1/Storage/1/Drives/1"
const testPayloadCrayNCSystems_Node1_Storage_1_Drives_1 = `
{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1/Drives/1",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400233      ",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node1_Memory = "/redfish/v1/Systems/Node1/Memory"

const testPayloadCrayNCSystems_Node1_Memory = `
{
	"@odata.context": "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
	"@odata.etag": "W/\"1549572116\"",
	"@odata.id": "/redfish/v1/Systems/Node1/Memory",
	"@odata.type": "#MemoryCollection.MemoryCollection",
	"Description": "Collection of Memories for this system",
	"Members": [],
	"Members@odata.count": 0,
	"Name": "Memory Collection"
}`

//////////////////////////////////////////////////////////////////////////////
//                  Cray Mountain Node Card Controller mock
//                  Support change to Systems/NodeX Actions
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncCrayNC2() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathCrayNC_redfish_v1_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNC_redfish_v1_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCAccountService_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCAccountService_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSessionService_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSessionService_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCEventService_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCEventService_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCTaskService_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCTaskService_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis_Enclosure_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis_Enclosure_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis_Node0_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis_Node0_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassis_Node1_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassis_Node1_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisPower_Node0_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisPower_Node0_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisPower_Node1_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisPower_Node1_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapters_Node0_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapters_Node0_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapters_Node1_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapters_Node1_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapter0_Node0_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapter0_Node0_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCChassisNetworkAdapter0_Node1_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCChassisNetworkAdapter0_Node1_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCManagers_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCManagers_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCManagers_BMC_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCManagers_BMC_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_ResetActionInfo_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_ResetActionInfo_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_EthernetInterfaces_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_EthernetInterfaces_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Processors_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Processors_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Storage_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Storage_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Storage_2_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Storage_2_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Storage_2_1_Drives_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Storage_2_1_Drives_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node0_Memory_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node0_Memory_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_ResetActionInfo_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_ResetActionInfo_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_EthernetInterfaces_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_EthernetInterfaces_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Processors_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Processors_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Storage_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Storage_2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Storage_2_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Storage_2_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Storage_2_1_Drives_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Storage_2_1_Drives_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayNCSystems_Node1_Memory_2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayNCSystems_Node1_Memory_2)),
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
				Header: make(http.Header),
			}
		}
	}
}

const testPathCrayNC_redfish_v1_2 = "/redfish/v1"
const testPayloadCrayNC_redfish_v1_2 = `
{
	"RedfishVersion" : "1.2.0",
		"Name" : "Root Service",
		"@odata.id" : "/redfish/v1/",
		"@odata.context" : "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
		"Managers" : {
			"@odata.id" : "/redfish/v1/Managers"
		},
		"AccountService" : {
			"@odata.id" : "/redfish/v1/AccountService"
		},
		"@odata.type" : "#ServiceRoot.v1_2_0.ServiceRoot",
		"Registries" : {
			"@odata.id" : "/redfish/v1/Registries"
		},
		"JsonSchemas" : {
			"@odata.id" : "/redfish/v1/JsonSchemas"
		},
		"Systems" : {
			"@odata.id" : "/redfish/v1/Systems"
		},
		"Oem" : {
			"@odata.type" : "ServiceRoot.v1_2_0.ServiceRoot",
			"Ami" : {
				"Configurations" : {
					"@odata.id" : "/redfish/v1/configurations"
				}
			}
		},
		"Description" : "The service root for all Redfish requests on this host",
		"TelemetryService" : {
			"@odata.id" : "/redfish/v1/TelemetryService"
		},
		"EventService" : {
			"@odata.id" : "/redfish/v1/EventService"
		},
		"Tasks" : {
			"@odata.id" : "/redfish/v1/TaskService"
		},
		"Links" : {
			"Sessions" : {
				"@odata.id" : "/redfish/v1/SessionService/Sessions"
			}
		},
		"Id" : "RootService",
		"UpdateService" : {
			"@odata.id" : "/redfish/v1/UpdateService"
		},
		"@odata.etag" : "W/\"665\"",
		"SessionService" : {
			"@odata.id" : "/redfish/v1/SessionService"
		},
		"Chassis" : {
			"@odata.id" : "/redfish/v1/Chassis"
		}
}`

const testPathCrayNCAccountService_2 = "/redfish/v1/AccountService"
const testPayloadCrayNCAccountService_2 = `
{
	"@odata.type" : "#AccountService.v1_2_1.AccountService",
		"AuthFailureLoggingThreshold" : 3,
		"@odata.id" : "/redfish/v1/AccountService",
		"@odata.context" : "/redfish/v1/$metadata#AccountService.AccountService",
		"MaxPasswordLength" : 12,
		"Status" : {
			"Health" : "OK",
			"State" : "Enabled"
		},
		"Roles" : {
			"@odata.id" : "/redfish/v1/AccountService/Roles"
		},
		"MinPasswordLength" : 8,
		"@odata.etag" : "W/\"665\"",
		"AccountLockoutCounterResetAfter" : 30,
		"Description" : "BMC User Accounts",
		"Accounts" : {
			"@odata.id" : "/redfish/v1/AccountService/Accounts"
		},
		"Id" : "AccountService",
		"Name" : "Account Service",
		"Oem" : {
			"@odata.type" : "AccountService.v1_2_1.AccountService",
			"Ami" : {
				"Configuration" : {
					"@odata.id" : "/redfish/v1/AccountService/Configurations"
				}
			}
		},
		"AccountLockoutDuration" : 30,
		"AccountLockoutThreshold" : 5,
		"ServiceEnabled" : true
}`

const testPathCrayNCSessionService_2 = "/redfish/v1/SessionService"
const testPayloadCrayNCSessionService_2 = `
{
	"ServiceEnabled" : true,
		"Name" : "Session Service",
		"Description" : "Session Service",
		"@odata.type" : "#SessionService.v1_1_3.SessionService",
		"Id" : "SessionService",
		"Sessions" : {
			"@odata.id" : "/redfish/v1/SessionService/Sessions"
		},
		"@odata.context" : "/redfish/v1/$metadata#SessionService.SessionService",
		"@odata.id" : "/redfish/v1/SessionService",
		"@odata.etag" : "W/\"665\"",
		"Status" : {
			"State" : "Enabled",
			"Health" : "OK"
		},
		"SessionTimeout" : 30
}`

const testPathCrayNCEventService_2 = "/redfish/v1/EventService"
const testPayloadCrayNCEventService_2 = `
{
	"@odata.etag" : "W/\"1562887592\"",
		"Id" : "EventService",
		"DeliveryRetryIntervalSeconds" : 60,
		"DeliveryRetryAttempts" : 3,
		"Description" : "Event Service",
		"EventTypesForSubscription" : [
			"StatusChange",
		"ResourceUpdated",
		"ResourceAdded",
		"ResourceRemoved",
		"Alert"
		],
		"Actions" : {
			"Oem" : {
				"Ami" : {
					"#EventService.SubmitDelayedTestEvent" : {
						"EventType@Redfish.AllowableValues" : [
							"StatusChange",
						"ResourceUpdated",
						"ResourceAdded",
						"ResourceRemoved",
						"Alert"
						],
						"target" : "/redfish/v1/EventService/Actions/EventService.SubmitDelayedTestEvent"
					}
				}
			},
			"#EventService.SubmitTestEvent" : {
				"EventType@Redfish.AllowableValues" : [
					"StatusChange",
				"ResourceUpdated",
				"ResourceAdded",
				"ResourceRemoved",
				"Alert"
				],
				"target" : "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"
			}
		},
		"Status" : {
			"Health" : "OK",
			"State" : "Enabled"
		},
		"@odata.id" : "/redfish/v1/EventService",
		"Subscriptions" : {
			"@odata.id" : "/redfish/v1/EventService/Subscriptions"
		},
		"@odata.context" : "/redfish/v1/$metadata#EventService.EventService",
		"Name" : "Event Service",
		"ServiceEnabled" : true,
		"@odata.type" : "#EventService.v1_0_5.EventService"
}`

const testPathCrayNCTaskService_2 = "/redfish/v1/TaskService"
const testPayloadCrayNCTaskService_2 = `
{
	"@odata.type" : "#TaskService.v1_1_0.TaskService",
		"@odata.context" : "/redfish/v1/$metadata#TaskService.TaskService",
		"CompletedTaskOverWritePolicy" : "Oldest",
		"ServiceEnabled" : true,
		"Name" : "Task Service",
		"Tasks" : {
			"@odata.id" : "/redfish/v1/TaskService/Tasks"
		},
		"Id" : "TaskService",
		"@odata.etag" : "W/\"665\"",
		"DateTime" : "1970-01-01T02:23:51Z",
		"LifeCycleEventOnTaskStateChange" : true,
		"Status" : {
			"Health" : "OK",
			"State" : "Enabled"
		},
		"@odata.id" : "/redfish/v1/TaskService",
		"Description" : "Task Service"
}`

const testPathCrayNCChassis_2 = "/redfish/v1/Chassis"
const testPayloadCrayNCChassis_2 = `
{
	"@odata.context" : "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
		"Name" : "Chassis Collection",
		"@odata.etag" : "W/\"665\"",
		"Description" : "The Collection for Chassis",
		"@odata.type" : "#ChassisCollection.ChassisCollection",
		"@odata.id" : "/redfish/v1/Chassis",
		"Members@odata.count" : 3,
		"Members" : [
		{
			"@odata.id" : "/redfish/v1/Chassis/Node0"
		},
		{
			"@odata.id" : "/redfish/v1/Chassis/Enclosure"
		},
		{
			"@odata.id" : "/redfish/v1/Chassis/Node1"
		}
		]
}`

const testPathCrayNCChassis_Enclosure_2 = "/redfish/v1/Chassis/Enclosure"
const testPayloadCrayNCChassis_Enclosure_2 = `
{
	"Manufacturer" : "Cray Inc",
		"AssetTag" : "",
		"NetworkAdapters" : {
			"@odata.id" : "/redfish/v1/Chassis/Enclosure/NetworkAdapters"
		},
		"Name" : "Enclosure",
		"@odata.id" : "/redfish/v1/Chassis/Enclosure",
		"ChassisType" : "Enclosure",
		"SerialNumber" : "HA19070035",
		"@odata.type" : "#Chassis.v1_5_1.Chassis",
		"Actions" : {
			"#Chassis.Reset" : {
				"ResetType@Redfish.AllowableValues" : [],
				"target" : "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset"
			}
		},
		"@odata.etag" : "W/\"665\"",
		"@odata.context" : "/redfish/v1/$metadata#Chassis.Chassis(SerialNumber,Id,Status,AssetTag,Name,Links,NetworkAdapters,ChassisType,Manufacturer,Actions)",
		"Status" : {
			"State" : "Enabled"
		},
		"Id" : "Enclosure",
		"Links" : {
			"PCIeDevices@odata.count" : 0,
			"Storage@odata.count" : 0,
			"Drives@odata.count" : 0,
			"ManagedBy@odata.count" : 0,
			"ManagersInChassis@odata.count" : 0,
			"Contains@odata.count" : 0,
			"CooledBy@odata.count" : 0,
			"PoweredBy@odata.count" : 0,
			"ComputerSystems@odata.count" : 0,
			"ResourceBlock@odata.count" : 0
		}
}`

const testPathCrayNCChassis_Node0_2 = "/redfish/v1/Chassis/Node0"
const testPayloadCrayNCChassis_Node0_2 = `
{
	"@odata.id" : "/redfish/v1/Chassis/Node0",
		"Name" : "Node0",
		"Thermal" : {
			"@odata.id" : "/redfish/v1/Chassis/Node0/Thermal"
		},
		"Manufacturer" : "Cray Inc",
		"Id" : "Node0",
		"Actions" : {
			"#Chassis.Reset" : {
				"target" : "/redfish/v1/Chassis/Node0/Actions/Chassis.Reset",
				"ResetType@Redfish.AllowableValues" : []
			}
		},
		"ChassisType" : "Blade",
		"@odata.context" : "/redfish/v1/$metadata#Chassis.Chassis(Thermal,Id,Status,AssetTag,Name,Power,Links,NetworkAdapters,ChassisType,Manufacturer,Actions)",
		"Power" : {
			"@odata.id" : "/redfish/v1/Chassis/Node0/Power"
		},
		"@odata.type" : "#Chassis.v1_5_1.Chassis",
		"Links" : {
			"Contains@odata.count" : 0,
			"PCIeDevices@odata.count" : 0,
			"ManagedBy@odata.count" : 0,
			"Drives@odata.count" : 0,
			"ResourceBlock@odata.count" : 0,
			"Storage@odata.count" : 0,
			"PoweredBy@odata.count" : 0,
			"CooledBy@odata.count" : 0,
			"ComputerSystems@odata.count" : 0,
			"ManagersInChassis@odata.count" : 0
		},
		"@odata.etag" : "W/\"665\"",
		"Status" : {
			"State" : "Enabled"
		},
		"NetworkAdapters" : {
			"@odata.id" : "/redfish/v1/Chassis/Node0/NetworkAdapters"
		},
		"AssetTag" : ""
}`

const testPathCrayNCChassis_Node1_2 = "/redfish/v1/Chassis/Node1"
const testPayloadCrayNCChassis_Node1_2 = `
{
	"Manufacturer" : "Cray Inc",
		"AssetTag" : "",
		"@odata.etag" : "W/\"665\"",
		"@odata.id" : "/redfish/v1/Chassis/Node1",
		"Status" : {
			"State" : "Enabled"
		},
		"Links" : {
			"CooledBy@odata.count" : 0,
			"ComputerSystems@odata.count" : 0,
			"ResourceBlock@odata.count" : 0,
			"ManagedBy@odata.count" : 0,
			"ManagersInChassis@odata.count" : 0,
			"Storage@odata.count" : 0,
			"Drives@odata.count" : 0,
			"PoweredBy@odata.count" : 0,
			"PCIeDevices@odata.count" : 0,
			"Contains@odata.count" : 0
		},
		"@odata.context" : "/redfish/v1/$metadata#Chassis.Chassis(Thermal,Id,Status,AssetTag,Name,Power,Links,NetworkAdapters,ChassisType,Manufacturer,Actions)",
		"Id" : "Node1",
		"@odata.type" : "#Chassis.v1_5_1.Chassis",
		"Thermal" : {
			"@odata.id" : "/redfish/v1/Chassis/Node1/Thermal"
		},
		"ChassisType" : "Blade",
		"Name" : "Node1",
		"Power" : {
			"@odata.id" : "/redfish/v1/Chassis/Node1/Power"
		},
		"NetworkAdapters" : {
			"@odata.id" : "/redfish/v1/Chassis/Node1/NetworkAdapters"
		},
		"Actions" : {
			"#Chassis.Reset" : {
				"ResetType@Redfish.AllowableValues" : [],
				"target" : "/redfish/v1/Chassis/Node1/Actions/Chassis.Reset"
			}
		}
}`

const testPathCrayNCChassisPower_Node0_2 = "/redfish/v1/Chassis/Node0/Power"
const testPayloadCrayNCChassisPower_Node0_2 = `
{
	"@odata.context": "/redfish/v1/$metadata#Power.Power(Voltages,Id,Voltages@odata.count,Name,Description)",
	"@odata.etag": "W/\"1569785935\"",
	"@odata.id": "/redfish/v1/Chassis/Node0/Power",
	"@odata.type": "#Power.v1_4_0.Power",
	"Description": "Power sensor readings",
	"Id": "Power",
	"Name": "Power",
	"PowerControl": [
		{
			"RelatedItem@odata.count": 1,
			"PowerCapacityWatts": 900,
			"Name": "Node Power Control",
			"Oem": {
				"Cray": {
					"PowerAllocatedWatts": 900,
					"PowerIdleWatts": 250,
					"PowerLimit": {
						"Min": 350,
						"Max": 850,
						"Factor": 1.02
					},
					"PowerFloorTargetWatts": 0,
					"PowerResetWatts": 250
				}
			},
			"@odata.id": "/redfish/v1/Chassis/Node0/Power#/PowerControl/Node",
			"PowerLimit": {
				"LimitException": "LogEventOnly",
				"CorrectionInMs": 6000,
				"LimitInWatts": 500
			},
			"RelatedItem": [
				{
					"@odata.id": "/redfish/v1/Chassis/Node0/Power#/PowerControl/Accelerator0"
				}
			]
		},
		{
			"RelatedItem@odata.count": 0,
			"Name": "Accelerator0 Power Control",
			"Oem": {
				"Cray": {
					"PowerIdleWatts": 100,
					"PowerLimit": {
						"Min": 200,
						"Max": 350,
						"Factor": 1.0
					},
					"PowerFloorTargetWatts": 0
				}
			},
			"@odata.id": "/redfish/v1/Chassis/Node0/Power#/PowerControl/Accelerator0",
			"PowerLimit": {
				"LimitException": "LogEventOnly",
				"CorrectionInMs": 6000,
				"LimitInWatts": 300
			}
		}
	],
	"Voltages": [
		{
			"@odata.id": "/redfish/v1/Chassis/Node0/Power#/Voltages/153",
			"Name": "VDD_1V2_S0 Voltage",
			"PhysicalContext": "VoltageRegulator",
			"ReadingVolts": 1.219,
			"RelatedItem@odata.count": 0
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Node0/Power#/Voltages/158",
			"Name": "CPUS VDD_3V3_S0 Voltage Output",
			"PhysicalContext": "VoltageRegulator",
			"ReadingVolts": 3.371,
			"RelatedItem@odata.count": 0
		}
	],
	"Voltages@odata.count": 2
}`

const testPathCrayNCChassisPower_Node1_2 = "/redfish/v1/Chassis/Node1/Power"
const testPayloadCrayNCChassisPower_Node1_2 = `
{
	"@odata.context": "/redfish/v1/$metadata#Power.Power(Voltages,Id,Voltages@odata.count,Name,Description)",
	"@odata.etag": "W/\"1569785935\"",
	"@odata.id": "/redfish/v1/Chassis/Node1/Power",
	"@odata.type": "#Power.v1_4_0.Power",
	"Description": "Power sensor readings",
	"Id": "Power",
	"Name": "Power",
	"PowerControl": [
		{
			"RelatedItem@odata.count": 1,
			"PowerCapacityWatts": 900,
			"Name": "Node Power Control",
			"Oem": {
				"Cray": {
					"PowerAllocatedWatts": 900,
					"PowerIdleWatts": 250,
					"PowerLimit": {
						"Min": 350,
						"Max": 850,
						"Factor": 1.02
					},
					"PowerFloorTargetWatts": 0,
					"PowerResetWatts": 250
				}
			},
			"@odata.id": "/redfish/v1/Chassis/Node1/Power#/PowerControl/Node",
			"PowerLimit": {
				"LimitException": "LogEventOnly",
				"CorrectionInMs": 6000,
				"LimitInWatts": 500
			},
			"RelatedItem": [
				{
					"@odata.id": "/redfish/v1/Chassis/Node1/Power#/PowerControl/Accelerator0"
				}
			]
		},
		{
			"RelatedItem@odata.count": 0,
			"Name": "Accelerator0 Power Control",
			"Oem": {
				"Cray": {
					"PowerIdleWatts": 100,
					"PowerLimit": {
						"Min": 200,
						"Max": 350,
						"Factor": 1.0
					},
					"PowerFloorTargetWatts": 0
				}
			},
			"@odata.id": "/redfish/v1/Chassis/Node1/Power#/PowerControl/Accelerator0",
			"PowerLimit": {
				"LimitException": "LogEventOnly",
				"CorrectionInMs": 6000,
				"LimitInWatts": 300
			}
		}
	],
	"Voltages": [
		{
			"@odata.id": "/redfish/v1/Chassis/Node1/Power#/Voltages/153",
			"Name": "VDD_1V2_S0 Voltage",
			"PhysicalContext": "VoltageRegulator",
			"ReadingVolts": 1.219,
			"RelatedItem@odata.count": 0
		},
		{
			"@odata.id": "/redfish/v1/Chassis/Node1/Power#/Voltages/116",
			"Name": "CPU0 DIMM VDD_VPP_EFGH_S0 Voltage Input",
			"PhysicalContext": "VoltageRegulator",
			"ReadingVolts": 11.969,
			"RelatedItem@odata.count": 0
		}
	],
	"Voltages@odata.count": 2
}`

const testPathCrayNCChassisNetworkAdapters_Node0_2 = "/redfish/v1/Chassis/Node0/NetworkAdapters"
const testPayloadCrayNCChassisNetworkAdapters_Node0_2 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapterCollection.NetworkAdapterCollection",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters",
  "@odata.type": "#NetworkAdapterCollection.NetworkAdapterCollection",
  "Description": "The Collection of Network Adapters",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0"
    }
  ],
  "Members@odata.count": 1,
  "Name": "NetworkAdapter Collection"
}`

const testPathCrayNCChassisNetworkAdapter0_Node0_2 = "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0"
const testPayloadCrayNCChassisNetworkAdapter0_Node0_2 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapter.NetworkAdapter(SerialNumber,Id,NetworkDeviceFunctions,NetworkPorts,PartNumber,Manufacturer,Model,Description,Actions)",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0",
  "@odata.type": "#NetworkAdapter.v1_0_1.NetworkAdapter",
  "Actions": {
    "#NetworkAdapter.ResetSettingsToDefault": {
      "ResetSettingsToDefaultType@Redfish.AllowableValues": [],
      "target": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0/Actions/NetworkAdapter.ResetSettingsToDefault"
    }
  },
  "Description": "Shasta Timms NMC REV04 (HSN)",
  "Id": "HPCNet0",
  "Manufacturer": "Mellanox Technologies, Ltd.",
  "Model": "ConnectX-5 100Gb/s",
  "NetworkDeviceFunctions": {
    "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0/NetworkDeviceFunctions"
  },
  "NetworkPorts": {
    "@odata.id": "/redfish/v1/Chassis/Node0/NetworkAdapters/HPCNet0/NetworkPorts"
  },
  "PartNumber": "102005303",
  "SerialNumber": "HG19501557"
}`

const testPathCrayNCChassisNetworkAdapters_Node1_2 = "/redfish/v1/Chassis/Node1/NetworkAdapters"
const testPayloadCrayNCChassisNetworkAdapters_Node1_2 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapterCollection.NetworkAdapterCollection",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters",
  "@odata.type": "#NetworkAdapterCollection.NetworkAdapterCollection",
  "Description": "The Collection of Network Adapters",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0"
    }
  ],
  "Members@odata.count": 1,
  "Name": "NetworkAdapter Collection"
}`

const testPathCrayNCChassisNetworkAdapter0_Node1_2 = "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0"
const testPayloadCrayNCChassisNetworkAdapter0_Node1_2 = `
{
  "@odata.context": "/redfish/v1/$metadata#NetworkAdapter.NetworkAdapter(SerialNumber,Id,NetworkDeviceFunctions,NetworkPorts,PartNumber,Manufacturer,Model,Description,Actions)",
  "@odata.etag": "W/\"1605721941\"",
  "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0",
  "@odata.type": "#NetworkAdapter.v1_0_1.NetworkAdapter",
  "Actions": {
    "#NetworkAdapter.ResetSettingsToDefault": {
      "ResetSettingsToDefaultType@Redfish.AllowableValues": [],
      "target": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0/Actions/NetworkAdapter.ResetSettingsToDefault"
    }
  },
  "Description": "Shasta Timms NMC REV04 (HSN)",
  "Id": "HPCNet0",
  "Manufacturer": "Mellanox Technologies, Ltd.",
  "Model": "ConnectX-5 100Gb/s",
  "NetworkDeviceFunctions": {
    "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0/NetworkDeviceFunctions"
  },
  "NetworkPorts": {
    "@odata.id": "/redfish/v1/Chassis/Node1/NetworkAdapters/HPCNet0/NetworkPorts"
  },
  "PartNumber": "102005303",
  "SerialNumber": "HG19501557"
}`

const testPathCrayNCManagers_2 = "/redfish/v1/Managers"
const testPayloadCrayNCManagers_2 = `
{
	"@odata.context" : "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
		"Description" : "The collection for Managers",
		"@odata.type" : "#ManagerCollection.ManagerCollection",
		"Name" : "Manager Collection",
		"Members@odata.count" : 1,
		"@odata.id" : "/redfish/v1/Managers",
		"Members" : [
		{
			"@odata.id" : "/redfish/v1/Managers/BMC"
		}
		],
		"@odata.etag" : "W/\"0\""
}`

const testPathCrayNCManagers_BMC_2 = "/redfish/v1/Managers/BMC"
const testPayloadCrayNCManagers_BMC_2 = `
{
	"Links" : {
		"ManagerForServers@odata.count" : 0,
			"ManagerInChassis" : {
				"@odata.id" : "/redfish/v1/Chassis/Self"
			},
			"ManagerForChassis@odata.count" : 0
	},
		"LogServices" : {
			"@odata.id" : "/redfish/v1/Managers/BMC/LogServices"
		},
		"Name" : "BMC",
		"Status" : {
			"State" : "Online",
			"Health" : "OK"
		},
		"Actions" : {
			"Oem" : {
				"#Manager.FactoryReset" : {
					"target" : "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset",
					"FactoryResetType@Redfish.AllowableValues" : [
						"ResetAll"
					]
				},
				"#CrayProcess.Schedule" : {
					"target" : "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule",
					"Name@Redfish.AllowableValues" : [
						"memtest",
					"cpuburn"
					]
				}
			},
			"#Manager.Reset" : {
				"ResetType@Redfish.AllowableValues" : [
					"ForceRestart",
				"ForceEraseNetworkReload"
				],
				"target" : "/redfish/v1/Managers/BMC/Actions/Manager.Reset"
			}
		},
		"@odata.etag" : "W/\"0\"",
		"Oem" : {
			"RemoteSyslog" : {
				"ServerUri" : "",
				"ForwardingEnabled" : false,
				"Port" : 514
			}
		},
		"DateTime" : "1970-01-01T02:23:54Z",
		"ManagerType" : "EnclosureManager",
		"NetworkProtocol" : {
			"@odata.id" : "/redfish/v1/Managers/BMC/NetworkProtocol"
		},
		"@odata.context" : "/redfish/v1/$metadata#Manager.Manager(DateTimeLocalOffset,Id,Status,NetworkProtocol,ManagerType,Oem,DateTime,Links,Name,LogServices,Description,Actions)",
		"Id" : "BMC",
		"Description" : "Shasta Manager",
        "Manufacturer": "Cray",
        "SerialNumber": "12345xyz",
		"@odata.id" : "/redfish/v1/Managers/BMC",
		"DateTimeLocalOffset" : "0",
		"@odata.type" : "#Manager.v1_3_2.Manager"
}`

const testPathCrayNCSystems_2 = "/redfish/v1/Systems"
const testPayloadCrayNCSystems_2 = `
{
	"Name" : "Systems Collection",
		"@odata.type" : "#ComputerSystemCollection.ComputerSystemCollection",
		"@odata.etag" : "W/\"665\"",
		"@odata.context" : "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
		"Description" : "Collection of Computer Systems",
		"@odata.id" : "/redfish/v1/Systems",
		"@Redfish.CollectionCapabilities" : {
			"Capabilities" : [
			{
				"UseCase" : "ComputerSystemComposition",
				"Links" : {
					"TargetCollection" : {
						"@odata.id" : "/redfish/v1/Systems"
					},
					"RelatedItem" : [
					{
						"@odata.id" : "/redfish/v1/CompositionService/ResourceZones/1"
					}
					]
				},
				"CapabilitiesObject" : {
					"@odata.id" : "/redfish/v1/Systems/Capabilities"
				}
			}
			],
			"@odata.type" : "#CollectionCapabilities.v1_0_0.CollectionCapabilities"
		},
		"Members@odata.count" : 2,
		"Members" : [
		{
			"@odata.id" : "/redfish/v1/Systems/Node1"
		},
		{
			"@odata.id" : "/redfish/v1/Systems/Node0"
		}
		]
}`

const testPathCrayNCSystems_Node0_2 = "/redfish/v1/Systems/Node0"
const testPayloadCrayNCSystems_Node0_2 = `
{
	"Actions" : {
		"#ComputerSystem.Reset" : {
			"target" : "/redfish/v1/Systems/Node0/Actions/ComputerSystem.Reset",
				"@Redfish.ActionInfo" : "/redfish/v1/Systems/Node0/ResetActionInfo"
		}
	},
		"PowerState" : "On",
		"PCIeDevices@odata.count" : 0,
		"Id" : "Node0",
		"@odata.id" : "/redfish/v1/Systems/Node0",
		"Name" : "Node0",
		"SecureBoot" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/SecureBoot"
		},
		"SimpleStorage" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/SimpleStorage"
		},
		"BiosVersion" : "wnc.bios-0.6.0",
		"Manufacturer" : "Cray Inc",
		"SystemType" : "Physical",
		"Bios" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/Bios"
		},
		"Processors" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/Processors"
		},
		"LogServices" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/LogServices"
		},
		"Description" : "Node",
		"MemoryDomains" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/MemoryDomains"
		},
		"NetworkInterfaces" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/NetworkInterfaces"
		},
		"@odata.context" : "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
		"Links" : {
			"Endpoints@odata.count" : 0,
			"Chassis@odata.count" : 0,
			"CooledBy@odata.count" : 0,
			"ManagedBy@odata.count" : 0,
			"PoweredBy@odata.count" : 0
		},
		"Memory" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/Memory"
		},
		"EthernetInterfaces" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/EthernetInterfaces"
		},
		"@odata.etag" : "W/\"664\"",
		"Boot" : {
			"BootSourceOverrideEnabled" : "Disabled",
			"BootSourceOverrideTarget@Redfish.AllowableValues" : [
				"None",
			"Pxe",
			"Floppy",
			"Cd",
			"Usb",
			"Hdd",
			"BiosSetup",
			"Utilities",
			"Diags",
			"UefiShell",
			"UefiTarget",
			"SDCard",
			"UefiHttp",
			"RemoteDrive",
			"UefiBootNext"
			],
			"BootOptions" : {
				"@odata.id" : "/redfish/v1/Systems/Node0/BootOptions"
			},
			"BootSourceOverrideTarget" : "None",
			"BootSourceOverrideEnabled@Redfish.AllowableValues" : [
				"Disabled",
			"Once",
			"Continuous"
			]
		},
		"Storage" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/Storage"
		},
		"@odata.type" : "#ComputerSystem.v1_5_0.ComputerSystem",
		"PCIeFunctions@odata.count" : 0
}`

const testPathCrayNCSystems_Node0_ResetActionInfo_2 = "/redfish/v1/Systems/Node0/ResetActionInfo"
const testPayloadCrayNCSystems_Node0_ResetActionInfo_2 = `
{
	"Name" : "ResetAction",
		"Id" : "ResetAction",
		"@odata.id" : "/redfish/v1/Systems/Node0/ResetActionInfo",
		"@odata.context" : "/redfish/v1/$metadata#ActionInfo.ActionInfo",
		"@odata.type" : "#ActionInfo.v1_0_3.ActionInfo",
		"Description" : "This action is used to reset the Systems",
		"@odata.etag" : "W/\"665\"",
		"Parameters" : [
		{
			"AllowableValues" : [
				"ForceOff",
			"Off",
			"On"
			],
			"DataType" : "String",
			"Name" : "ResetType",
			"Required" : true
		}
		]
}`

const testPathCrayNCSystems_Node0_EthernetInterfaces_2 = "/redfish/v1/Systems/Node0/EthernetInterfaces"
const testPayloadCrayNCSystems_Node0_EthernetInterfaces_2 = `
{
	"Name" : "Ethernet Interface Collection",
		"Members@odata.count" : 1,
		"Members" : [
		{
			"@odata.id" : "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet"
		}
		],
		"@odata.etag" : "W/\"665\"",
		"@odata.context" : "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
		"@odata.type" : "#EthernetInterfaceCollection.EthernetInterfaceCollection",
		"Description" : "Collection of ethernet interfaces for this system",
		"@odata.id" : "/redfish/v1/Systems/Node0/EthernetInterfaces"
}`

const testPathCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet_2 = "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet"
const testPayloadCrayNCSystems_Node0_EthernetInterfaces_ManagementEthernet_2 = `
{
	"Id" : "ManagementEthernet",
		"VLANs" : {
			"@odata.id" : "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet/VLANs"
		},
		"@odata.type" : "#EthernetInterface.v1_3_0.EthernetInterface",
		"PermanentMACAddress" : "00:40:a6:82:f5:89",
		"Description" : "Node Maintenance Network",
		"@odata.context" : "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
		"@odata.etag" : "W/\"665\"",
		"MACAddress" : "00:40:a6:82:f5:89",
		"Links" : {
			"Endpoints@odata.count" : 0,
			"Chassis" : {
				"@odata.id" : "/redfish/v1/Chassis/Self"
			}
		},
		"@odata.id" : "/redfish/v1/Systems/Node0/EthernetInterfaces/ManagementEthernet"
}`

const testPathCrayNCSystems_Node0_Storage_2 = "/redfish/v1/Systems/Node0/Storage"
const testPayloadCrayNCSystems_Node0_Storage_2 = `
{
    "@odata.context": "/redfish/v1/$metadata#StorageCollection.StorageCollection",
    "@odata.id": "/redfish/v1/Systems/Node0/Storage",
    "@odata.type": "#StorageCollection.StorageCollection",
    "Name": "Storage Collection",
    "Members@odata.count": 1,
    "Members": [
        {
            "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1"
        }
    ],
    "Description" : "Collection of Storage resource instances",
    "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node0_Storage_2_1 = "/redfish/v1/Systems/Node0/Storage/1"
const testPayloadCrayNCSystems_Node0_Storage_2_1 = `{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1",
   "@odata.type" : "#Storage.v1_5_0.Storage",
   "@odata.context" : "/redfish/v1/$metadata#Storage.Storage",
   "Description" : "This resource shall be used to represent resources that represent a storage subsystem in the Redfish specification.",
   "Name" : "Storage",
   "Drives" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1/Drives/1"
      }
   ],
   "Drives@odata.count" : 1,
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node0_Storage_2_1_Drives_1 = "/redfish/v1/Systems/Node0/Storage/1/Drives/1"
const testPayloadCrayNCSystems_Node0_Storage_2_1_Drives_1 = `
{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node0/Storage/1/Drives/1",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400233      ",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node0_Processors_2 = "/redfish/v1/Systems/Node0/Processors"
const testPayloadCrayNCSystems_Node0_Processors_2 = `
{
	"@odata.type" : "#ProcessorCollection.ProcessorCollection",
		"Members@odata.count" : 0,
		"Description" : "Collection of processors",
		"@odata.etag" : "W/\"665\"",
		"Name" : "Processors Collection",
		"@odata.id" : "/redfish/v1/Systems/Node0/Processors",
		"@odata.context" : "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
		"Members" : []
}`

const testPathCrayNCSystems_Node0_Memory_2 = "/redfish/v1/Systems/Node0/Memory"
const testPayloadCrayNCSystems_Node0_Memory_2 = `
{
	"Name" : "Memory Collection",
	"Members" : [],
	"@odata.etag" : "W/\"665\"",
	"Description" : "Collection of Memories for this system",
	"@odata.context" : "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
	"Members@odata.count" : 0,
	"@odata.id" : "/redfish/v1/Systems/Node0/Memory",
	"@odata.type" : "#MemoryCollection.MemoryCollection"
}`

const testPathCrayNCSystems_Node1_2 = "/redfish/v1/Systems/Node1"
const testPayloadCrayNCSystems_Node1_2 = `
{
	"PowerState" : "On",
		"Actions" : {
			"#ComputerSystem.Reset" : {
				"target" : "/redfish/v1/Systems/Node1/Actions/ComputerSystem.Reset",
				"@Redfish.ActionInfo" : "/redfish/v1/Systems/Node1/ResetActionInfo"
			}
		},
		"SecureBoot" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/SecureBoot"
		},
		"Links" : {
			"PoweredBy@odata.count" : 0,
			"Endpoints@odata.count" : 0,
			"CooledBy@odata.count" : 0,
			"ManagedBy@odata.count" : 0,
			"Chassis@odata.count" : 0
		},
		"Bios" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/Bios"
		},
		"SimpleStorage" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/SimpleStorage"
		},
		"Manufacturer" : "Cray Inc",
		"Id" : "Node1",
		"Boot" : {
			"BootSourceOverrideEnabled@Redfish.AllowableValues" : [
				"Disabled",
			"Once",
			"Continuous"
			],
			"BootSourceOverrideTarget@Redfish.AllowableValues" : [
				"None",
			"Pxe",
			"Floppy",
			"Cd",
			"Usb",
			"Hdd",
			"BiosSetup",
			"Utilities",
			"Diags",
			"UefiShell",
			"UefiTarget",
			"SDCard",
			"UefiHttp",
			"RemoteDrive",
			"UefiBootNext"
			],
			"BootSourceOverrideTarget" : "None",
			"BootSourceOverrideEnabled" : "Disabled",
			"BootOptions" : {
				"@odata.id" : "/redfish/v1/Systems/Node1/BootOptions"
			}
		},
		"NetworkInterfaces" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/NetworkInterfaces"
		},
		"Memory" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/Memory"
		},
		"MemoryDomains" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/MemoryDomains"
		},
		"Processors" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/Processors"
		},
		"EthernetInterfaces" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/EthernetInterfaces"
		},
		"SystemType" : "Physical",
		"LogServices" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/LogServices"
		},
		"Description" : "Node",
		"Name" : "Node1",
		"@odata.type" : "#ComputerSystem.v1_5_0.ComputerSystem",
		"@odata.context" : "/redfish/v1/$metadata#ComputerSystem.ComputerSystem",
		"BiosVersion" : "wnc.bios-0.6.0",
		"PCIeDevices@odata.count" : 0,
		"@odata.id" : "/redfish/v1/Systems/Node1",
		"Storage" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/Storage"
		},
		"PCIeFunctions@odata.count" : 0,
		"@odata.etag" : "W/\"665\""
}`

const testPathCrayNCSystems_Node1_ResetActionInfo_2 = "/redfish/v1/Systems/Node1/ResetActionInfo"
const testPayloadCrayNCSystems_Node1_ResetActionInfo_2 = `
{
	"Parameters" : [
	{
		"DataType" : "String",
			"AllowableValues" : [
				"ForceOff",
			"Off",
			"On"
			],
			"Required" : true,
			"Name" : "ResetType"
	}
	],
	"@odata.etag" : "W/\"665\"",
	"@odata.id" : "/redfish/v1/Systems/Node1/ResetActionInfo",
	"@odata.type" : "#ActionInfo.v1_0_3.ActionInfo",
	"Name" : "ResetAction",
	"Id" : "ResetAction",
	"Description" : "This action is used to reset the Systems",
	"@odata.context" : "/redfish/v1/$metadata#ActionInfo.ActionInfo"
}`

const testPathCrayNCSystems_Node1_EthernetInterfaces_2 = "/redfish/v1/Systems/Node1/EthernetInterfaces"
const testPayloadCrayNCSystems_Node1_EthernetInterfaces_2 = `
{
	"@odata.type" : "#EthernetInterfaceCollection.EthernetInterfaceCollection",
		"@odata.context" : "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
		"@odata.etag" : "W/\"665\"",
		"Members@odata.count" : 1,
		"Description" : "Collection of ethernet interfaces for this system",
		"Members" : [
		{
			"@odata.id" : "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet"
		}
		],
		"Name" : "Ethernet Interface Collection",
		"@odata.id" : "/redfish/v1/Systems/Node1/EthernetInterfaces"
}`

const testPathCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet_2 = "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet"
const testPayloadCrayNCSystems_Node1_EthernetInterfaces_ManagementEthernet_2 = `
{
	"Links" : {
		"Chassis" : {
			"@odata.id" : "/redfish/v1/Chassis/Self"
		},
			"Endpoints@odata.count" : 0
	},
		"Description" : "Node Maintenance Network",
		"PermanentMACAddress" : "00:40:a6:82:f5:8a",
		"VLANs" : {
			"@odata.id" : "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet/VLANs"
		},
		"@odata.etag" : "W/\"665\"",
		"MACAddress" : "00:40:a6:82:f5:8a",
		"@odata.type" : "#EthernetInterface.v1_3_0.EthernetInterface",
		"@odata.context" : "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
		"@odata.id" : "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet",
		"Id" : "ManagementEthernet"
}`

const testPathCrayNCSystems_Node1_Processors_2 = "/redfish/v1/Systems/Node1/Processors"
const testPayloadCrayNCSystems_Node1_Processors_2 = `
{
	"@odata.context" : "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
		"@odata.etag" : "W/\"665\"",
		"Members@odata.count" : 0,
		"Name" : "Processors Collection",
		"Members" : [],
		"@odata.id" : "/redfish/v1/Systems/Node1/Processors",
		"@odata.type" : "#ProcessorCollection.ProcessorCollection",
		"Description" : "Collection of processors"
}`

const testPathCrayNCSystems_Node1_Storage_2 = "/redfish/v1/Systems/Node1/Storage"
const testPayloadCrayNCSystems_Node1_Storage_2 = `
{
    "@odata.context": "/redfish/v1/$metadata#StorageCollection.StorageCollection",
    "@odata.id": "/redfish/v1/Systems/Node1/Storage",
    "@odata.type": "#StorageCollection.StorageCollection",
    "Name": "Storage Collection",
    "Members@odata.count": 1,
    "Members": [
        {
            "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1"
        }
    ],
    "Description" : "Collection of Storage resource instances",
    "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node1_Storage_2_1 = "/redfish/v1/Systems/Node1/Storage/1"
const testPayloadCrayNCSystems_Node1_Storage_2_1 = `{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1",
   "@odata.type" : "#Storage.v1_5_0.Storage",
   "@odata.context" : "/redfish/v1/$metadata#Storage.Storage",
   "Description" : "This resource shall be used to represent resources that represent a storage subsystem in the Redfish specification.",
   "Name" : "Storage",
   "Drives" : [
      {
         "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1/Drives/1"
      }
   ],
   "Drives@odata.count" : 1,
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node1_Storage_2_1_Drives_1 = "/redfish/v1/Systems/Node1/Storage/1/Drives/1"
const testPayloadCrayNCSystems_Node1_Storage_2_1_Drives_1 = `
{
   "Id" : "1",
   "@odata.id" : "/redfish/v1/Systems/Node1/Storage/1/Drives/1",
   "@odata.type" : "#Drive.v1_5_0.Drive",
   "@odata.context" : "/redfish/v1/$metadata#Drive.Drive",
   "Name" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "Description" : "This resource shall be used to represent a disk drive or other physical storage medium for a Redfish implementation.",
   "FailurePredicted" : false,
   "Status" : {
      "State" : "Enabled",
      "HealthRollup" : "OK",
      "Health" : "OK"
   },
   "CapacityBytes" : 4027323514880,
   "Model" : "SAMSUNG MZ7LH3T8HMLT-00005",
   "SerialNumber" : "S456NY0M400233      ",
   "Oem" : {
      "GBT" : {
         "SlotNumber" : "ff",
         "@odata.type" : "#GbtOemDrives.v1_0_0.GbtOemDrives"
      }
   },
   "Links" : {
      "Chassis" : {
         "@odata.id" : "/redfish/v1/Chassis/Self"
      }
   },
   "@odata.etag" : "W/\"1579629364\""
}
`

const testPathCrayNCSystems_Node1_Memory_2 = "/redfish/v1/Systems/Node1/Memory"
const testPayloadCrayNCSystems_Node1_Memory_2 = `
{
	"Description" : "Collection of Memories for this system",
		"@odata.type" : "#MemoryCollection.MemoryCollection",
		"@odata.context" : "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
		"Members@odata.count" : 0,
		"Members" : [],
	"@odata.id" : "/redfish/v1/Systems/Node1/Memory",
	"@odata.etag" : "W/\"665\"",
	"Name" : "Memory Collection"
}`

//////////////////////////////////////////////////////////////////////////////
//               Cray Mountain Router Card/BMC Controller mock
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncCrayRC1() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathCrayRC_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRC_redfish_v1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCAccountService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCAccountService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCSessionService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCSessionService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCEventService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCEventService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCTaskService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCTaskService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCChassis:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCChassis)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCChassis_Enclosure:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCChassis_Enclosure)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCManagers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCManagers)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCManagers_BMC:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCManagers_BMC)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathCrayRCSystems:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadCrayRCSystems)),
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
				Header: make(http.Header),
			}
		}
	}
}

const testPathCrayRC_redfish_v1 = "/redfish/v1"

const testPayloadCrayRC_redfish_v1 = `
{
        "@odata.context": "/redfish/v1/$metadata#ServiceRoot.ServiceRoot",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/",
        "@odata.type": "#ServiceRoot.v1_2_0.ServiceRoot",
        "AccountService": {
                "@odata.id": "/redfish/v1/AccountService"
        },
        "Chassis": {
                "@odata.id": "/redfish/v1/Chassis"
        },
        "Description": "The service root for all Redfish requests on this host",
        "EventService": {
                "@odata.id": "/redfish/v1/EventService"
        },
        "Id": "RootService",
        "JsonSchemas": {
                "@odata.id": "/redfish/v1/JsonSchemas"
        },
        "Links": {
                "Sessions": {
                        "@odata.id": "/redfish/v1/SessionService/Sessions"
                }
        },
        "Managers": {
                "@odata.id": "/redfish/v1/Managers"
        },
        "Name": "Root Service",
        "Oem": {
                "@odata.type": "ServiceRoot.v1_2_0.ServiceRoot",
                "Ami": {
                        "Configurations": {
                                "@odata.id": "/redfish/v1/configurations"
                        }
                }
        },
        "RedfishVersion": "1.2.0",
        "Registries": {
                "@odata.id": "/redfish/v1/Registries"
        },
        "SessionService": {
                "@odata.id": "/redfish/v1/SessionService"
        },
        "Tasks": {
                "@odata.id": "/redfish/v1/TaskService"
        },
        "TelemetryService": {
                "@odata.id": "/redfish/v1/TelemetryService"
        },
        "UpdateService": {
                "@odata.id": "/redfish/v1/UpdateService"
        }
}`

const testPathCrayRCAccountService = "/redfish/v1/AccountService"

const testPayloadCrayRCAccountService = `
{
        "@odata.context": "/redfish/v1/$metadata#AccountService.AccountService",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/AccountService",
        "@odata.type": "#AccountService.v1_2_1.AccountService",
        "AccountLockoutCounterResetAfter": 30,
        "AccountLockoutDuration": 30,
        "AccountLockoutThreshold": 5,
        "Accounts": {
                "@odata.id": "/redfish/v1/AccountService/Accounts"
        },
        "AuthFailureLoggingThreshold": 3,
        "Description": "BMC User Accounts",
        "Id": "AccountService",
        "MaxPasswordLength": 12,
        "MinPasswordLength": 8,
        "Name": "Account Service",
        "Oem": {
                "@odata.type": "AccountService.v1_2_1.AccountService",
                "Ami": {
                        "Configuration": {
                                "@odata.id": "/redfish/v1/AccountService/Configurations"
                        }
                }
        },
        "Roles": {
                "@odata.id": "/redfish/v1/AccountService/Roles"
        },
        "ServiceEnabled": true,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        }
}`

const testPathCrayRCSessionService = "/redfish/v1/SessionService"

const testPayloadCrayRCSessionService = `
{
        "@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/SessionService",
        "@odata.type": "#SessionService.v1_1_3.SessionService",
        "Description": "Session Service",
        "Id": "SessionService",
        "Name": "Session Service",
        "ServiceEnabled": true,
        "SessionTimeout": 30,
        "Sessions": {
                "@odata.id": "/redfish/v1/SessionService/Sessions"
        },
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        }
}`

const testPathCrayRCEventService = "/redfish/v1/EventService"

const testPayloadCrayRCEventService = `
{
        "@odata.context": "/redfish/v1/$metadata#EventService.EventService",
        "@odata.etag": "W/\"1550870601\"",
        "@odata.id": "/redfish/v1/EventService",
        "@odata.type": "#EventService.v1_0_5.EventService",
        "Actions": {
                "#EventService.SubmitTestEvent": {
                        "EventType@Redfish.AllowableValues": [
                                "StatusChange",
                                "ResourceUpdated",
                                "ResourceAdded",
                                "ResourceRemoved",
                                "Alert"
                        ],
                        "target": "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"
                },
                "Oem": {
                        "Ami": {
                                "#EventService.SubmitDelayedTestEvent": {
                                        "EventType@Redfish.AllowableValues": [
                                                "StatusChange",
                                                "ResourceUpdated",
                                                "ResourceAdded",
                                                "ResourceRemoved",
                                                "Alert"
                                        ],
                                        "target": "/redfish/v1/EventService/Actions/EventService.SubmitDelayedTestEvent"
                                }
                        }
                }
        },
        "DeliveryRetryAttempts": 3,
        "DeliveryRetryIntervalSeconds": 60,
        "Description": "Event Service",
        "EventTypesForSubscription": [
                "StatusChange",
                "ResourceUpdated",
                "ResourceAdded",
                "ResourceRemoved",
                "Alert"
        ],
        "Id": "EventService",
        "Name": "Event Service",
        "ServiceEnabled": true,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "Subscriptions": {
                "@odata.id": "/redfish/v1/EventService/Subscriptions"
        }
}`

const testPathCrayRCTaskService = "/redfish/v1/TaskService"

const testPayloadCrayRCTaskService = `
{
        "@odata.context": "/redfish/v1/$metadata#TaskService.TaskService",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/TaskService",
        "@odata.type": "#TaskService.v1_1_0.TaskService",
        "CompletedTaskOverWritePolicy": "Oldest",
        "DateTime": "2019-04-03T18:04:02Z",
        "Description": "Task Service",
        "Id": "TaskService",
        "LifeCycleEventOnTaskStateChange": true,
        "Name": "Task Service",
        "ServiceEnabled": true,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "Tasks": {
                "@odata.id": "/redfish/v1/TaskService/Tasks"
        }
}`

const testPathCrayRCChassis = "/redfish/v1/Chassis"

const testPayloadCrayRCChassis = `
{
        "@odata.context": "/redfish/v1/$metadata#ChassisCollection.ChassisCollection",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/Chassis",
        "@odata.type": "#ChassisCollection.ChassisCollection",
        "Description": "The Collection for Chassis",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Chassis/Enclosure"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Chassis Collection"
}`

const testPathCrayRCChassis_Enclosure = "/redfish/v1/Chassis/Enclosure"

const testPayloadCrayRCChassis_Enclosure = `
{
        "@odata.context": "/redfish/v1/$metadata#Chassis.Chassis(Id,Status,AssetTag,PowerState,Links,NetworkAdapters,ChassisType,Manufacturer,Name,Actions)",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/Chassis/Enclosure",
        "@odata.type": "#Chassis.v1_5_1.Chassis",
        "Actions": {
                "#Chassis.Reset": {
                        "ResetType@Redfish.AllowableValues": [
                                "On",
                                "ForceOff",
                                "Off"
                        ],
                        "target": "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset"
                }
        },
        "AssetTag": "",
        "ChassisType": "Enclosure",
        "Id": "Enclosure",
        "Links": {
                "ComputerSystems@odata.count": 0,
                "Contains@odata.count": 0,
                "CooledBy@odata.count": 0,
                "Drives@odata.count": 0,
                "ManagedBy@odata.count": 0,
                "ManagersInChassis@odata.count": 0,
                "PCIeDevices@odata.count": 0,
                "PoweredBy@odata.count": 0,
                "ResourceBlock@odata.count": 0,
                "Storage@odata.count": 0
        },
        "Manufacturer": "Cray Inc",
        "SerialNumber": "12345xyz",
        "Name": "Enclosure",
        "NetworkAdapters": {
                "@odata.id": "/redfish/v1/Chassis/Enclosure/NetworkAdapters"
        },
        "PowerState": "On",
        "Status": {
                "State": "Absent"
        }
}`

const testPathCrayRCManagers = "/redfish/v1/Managers"

const testPayloadCrayRCManagers = `
{
        "@odata.context": "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/Managers",
        "@odata.type": "#ManagerCollection.ManagerCollection",
        "Description": "The collection for Managers",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Managers/BMC"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Manager Collection"
}`

const testPathCrayRCManagers_BMC = "/redfish/v1/Managers/BMC"

const testPayloadCrayRCManagers_BMC = `
{
        "@odata.context": "/redfish/v1/$metadata#Manager.Manager(DateTimeLocalOffset,Id,Status,NetworkProtocol,ManagerType,DateTime,Links,Name,LogServices,Description,Actions)",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/Managers/BMC",
        "@odata.type": "#Manager.v1_3_2.Manager",
        "Actions": {
                "#Manager.Reset": {
                        "ResetType@Redfish.AllowableValues": [
                                "ForceRestart",
                                "ForceEraseNetworkReload"
                        ],
                        "target": "/redfish/v1/Managers/BMC/Actions/Manager.Reset"
                },
                "Oem": {
                        "#CrayProcess.Schedule": {
                                "Name@Redfish.AllowableValues": [
                                        "memtest",
                                        "cpuburn"
                                ],
                                "target": "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"
                        },
                        "#Manager.FactoryReset": {
                                "FactoryResetType@Redfish.AllowableValues": [
                                        "ResetAll"
                                ],
                                "target": "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset"
                        }
                }
        },
        "DateTime": "2019-04-03T18:04:02Z",
        "DateTimeLocalOffset": "0",
        "Description": "Shasta Manager",
        "Id": "BMC",
        "Links": {
                "ManagerForChassis@odata.count": 0,
                "ManagerForServers@odata.count": 0,
                "ManagerInChassis": {
                        "@odata.id": "/redfish/v1/Chassis/Self"
                }
        },
        "LogServices": {
                "@odata.id": "/redfish/v1/Managers/BMC/LogServices"
        },
        "ManagerType": "EnclosureManager",
        "Name": "BMC",
        "NetworkProtocol": {
                "@odata.id": "/redfish/v1/Managers/BMC/NetworkProtocol"
        },
        "Manufacturer": "Cray",
        "SerialNumber": "12345xyz",
        "Status": {
                "Health": "OK",
                "State": "Online"
        }
}`

const testPathCrayRCSystems = "/redfish/v1/Systems"

const testPayloadCrayRCSystems = `
{
        "@Redfish.CollectionCapabilities": {
                "@odata.type": "#CollectionCapabilities.v1_0_0.CollectionCapabilities",
                "Capabilities": [
                        {
                                "CapabilitiesObject": {
                                        "@odata.id": "/redfish/v1/Systems/Capabilities"
                                },
                                "Links": {
                                        "RelatedItem": [
                                                {
                                                        "@odata.id": "/redfish/v1/CompositionService/ResourceZones/1"
                                                }
                                        ],
                                        "TargetCollection": {
                                                "@odata.id": "/redfish/v1/Systems"
                                        }
                                },
                                "UseCase": "ComputerSystemComposition"
                        }
                ]
        },
        "@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
        "@odata.etag": "W/\"0\"",
        "@odata.id": "/redfish/v1/Systems",
        "@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
        "Members@odata.count": 0,
        "Name": "Systems Collection"
}`

//////////////////////////////////////////////////////////////////////////////
//                 RTS CabinetPDU Controller mock
//////////////////////////////////////////////////////////////////////////////

func NewRTFuncRtsPDU1() RTFunc {
	return func(req *http.Request) *http.Response {
		// Test request parameters
		switch req.URL.String() {
		case "https://" + testFQDN + testPathRtsPDU_redfish_v1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDU_redfish_v1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUAccountService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUAccountService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUSessionService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUSessionService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUEventService:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUEventService)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUManagers:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUManagers)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUManagers_BMC:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUManagers_BMC)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUManagers_BMC_NICs:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUManagers_BMC_NICs)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUManagers_BMC_NICs_Dedicated:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUManagers_BMC_NICs_Dedicated)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A5:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A5)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_B1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_B1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_B2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_B2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A4:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A4)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_B3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_B3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_C1:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_C1)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_C2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_C2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_C3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_C3)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A2:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A2)),
				Header: make(http.Header),
			}
		case "https://" + testFQDN + testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A3:
			return &http.Response{
				StatusCode: 200,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString(testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A3)),
				Header: make(http.Header),
			}
		default:
			return &http.Response{
				StatusCode: 404,
				// Send mock response for rpath
				Body:   ioutil.NopCloser(bytes.NewBufferString("")),
				Header: make(http.Header),
			}
		}
	}
}

const testPathRtsPDU_redfish_v1 = "/redfish/v1"

const testPayloadRtsPDU_redfish_v1 = `
{
        "@odata.id": "/redfish/v1/",
        "@odata.type": "#ServiceRoot.v1_6_0.ServiceRoot",
        "AccountService": {
                "@odata.id": "/redfish/v1/AccountService"
        },
        "EventService": {
                "@odata.id": "/redfish/v1/EventService"
        },
        "Facilities": {
                "@odata.id": "/redfish/v1/Facilities"
        },
        "Id": "RootService",
        "Links": {
                "Sessions": {
                        "@odata.id": "/redfish/v1/SessionService/Sessions"
                }
        },
        "Managers": {
                "@odata.id": "/redfish/v1/Managers"
        },
        "Name": "Root Service",
        "Oem": {},
        "PowerEquipment": {
                "@odata.id": "/redfish/v1/PowerEquipment"
        },
        "RedfishVersion": "1.6.0",
        "SessionService": {
                "@odata.id": "/redfish/v1/SessionService"
        },
        "UUID": "92384634-2938-2342-8820-489239905423"
}`

const testPathRtsPDUAccountService = "/redfish/v1/AccountService"

const testPayloadRtsPDUAccountService = `
{
        "@odata.context": "/redfish/v1/$metadata#AccountService.AccountService",
        "@odata.id": "/redfish/v1/AccountService",
        "@odata.type": "#AccountService.v1_2_0.AccountService",
        "AccountLockoutCounterResetAfter": 30,
        "AccountLockoutDuration": 30,
        "AccountLockoutThreshold": 5,
        "Accounts": {
                "@odata.id": "/redfish/v1/AccountService/Accounts"
        },
        "AuthFailureLoggingThreshold": 3,
        "Description": "Account Service",
        "Id": "AccountService",
        "MinPasswordLength": 8,
        "Name": "Account Service",
        "Roles": {
                "@odata.id": "/redfish/v1/AccountService/Roles"
        },
        "ServiceEnabled": true,
        "Status": {
                "@odata.type": "#Resource.Status",
                "Health": "OK",
                "State": "Enabled"
        }
}`

const testPathRtsPDUSessionService = "/redfish/v1/SessionService"

const testPayloadRtsPDUSessionService = `
{
        "@odata.context": "/redfish/v1/$metadata#SessionService.SessionService",
        "@odata.id": "/redfish/v1/SessionService",
        "@odata.type": "#SessionService.v1_1_2.SessionService",
        "Description": "Session Service",
        "Id": "SessionService",
        "Name": "Session Service",
        "ServiceEnabled": true,
        "SessionTimeout": 30,
        "Sessions": {
                "@odata.id": "/redfish/v1/SessionService/Sessions"
        },
        "Status": {
                "@odata.type": "#Resource.Status",
                "Health": "OK",
                "State": "Enabled"
        }
}`

const testPathRtsPDUEventService = "/redfish/v1/EventService"

const testPayloadRtsPDUEventService = `
{
        "@odata.context": "/redfish/v1/$metadata#EventService.EventService",
        "@odata.id": "/redfish/v1/EventService",
        "@odata.type": "#EventService.v1_0_4.EventService",
        "Actions": {
                "#EventService.SubmitTestEvent": {
                        "EventType@Redfish.AllowableValues": [
                                "StatusChange",
                                "ResourceUpdated",
                                "ResourceAdded",
                                "ResourceRemoved",
                                "Alert"
                        ],
                        "target": "/redfish/v1/EventService/Actions/EventService.SubmitTestEvent"
                },
                "@odata.type": "#EventService.v1_0_0.Actions",
                "Oem": {}
        },
        "DeliveryRetryAttempts": 3,
        "DeliveryRetryIntervalSeconds": 60,
        "EventTypesForSubscription": [
                "StatusChange",
                "ResourceUpdated",
                "ResourceAdded",
                "ResourceRemoved",
                "Alert"
        ],
        "Id": "EventService",
        "Name": "Event Service",
        "Oem": {},
        "ServiceEnabled": true,
        "Status": {
                "@odata.type": "#Resource.Status",
                "Health": "OK",
                "State": "Enabled"
        },
        "Subscriptions": {
                "@odata.id": "/redfish/v1/EventService/Subscriptions"
        }
}`

const testPathRtsPDUManagers = "/redfish/v1/Managers"

const testPayloadRtsPDUManagers = `
{
        "@odata.context": "/redfish/v1/$metadata#ManagerCollection.ManagerCollection",
        "@odata.id": "/redfish/v1/Managers",
        "@odata.type": "#ManagerCollection.ManagerCollection",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Managers/BMC"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Manager Collection",
        "Oem": {}
}`

const testPathRtsPDUManagers_BMC = "/redfish/v1/Managers/BMC"

const testPayloadRtsPDUManagers_BMC = `
{
        "@odata.context": "/redfish/v1/$metadata#Manager.Manager",
        "@odata.id": "/redfish/v1/Managers/BMC",
        "@odata.type": "#Manager.v1_3_1.Manager",
        "Actions": {
                "#Manager.Reset": {
                        "ResetType@Redfish.AllowableValues": [
                                "ForceRestart",
                                "GracefulRestart"
                        ],
                        "target": "/redfish/v1/Managers/BMC/Actions/Manager.Reset"
                },
                "@odata.type": "#Manager.v1_0_0.Actions",
                "Oem": {}
        },
        "CommandShell": {
                "@odata.type": "#Manager.v1_0_0.CommandShell",
                "ConnectTypesSupported": [
                        "Telnet",
                        "SSH"
                ],
                "MaxConcurrentSessions": 4,
                "ServiceEnabled": true
        },
        "DateTime": "2015-03-13T04:14:33+06:00",
        "DateTimeLocalOffset": "+06:00",
        "Description": "Contoso BMC",
        "EthernetInterfaces": {
                "@odata.id": "/redfish/v1/Managers/BMC/NICs"
        },
        "FirmwareVersion": "1.00",
        "GraphicalConsole": {
                "@odata.type": "#Manager.v1_0_0.GraphicalConsole",
                "ConnectTypesSupported": [
                        "KVMIP"
                ],
                "MaxConcurrentSessions": 2,
                "ServiceEnabled": true
        },
        "Id": "BMC",
        "Links": {
                "@odata.type": "#Manager.v1_1_0.Links",
                "ManagerForChassis": [],
                "ManagerForServers": [],
                "Oem": {}
        },
        "LogServices": {
                "@odata.id": "/redfish/v1/Managers/BMC/LogServices"
        },
        "ManagerType": "BMC",
        "Model": "Joo Janta 200",
        "Name": "Manager",
        "NetworkProtocol": {
                "@odata.id": "/redfish/v1/Managers/BMC/NetworkProtocol"
        },
        "Oem": {},
        "PowerState": "On",
        "SerialConsole": {
                "@odata.type": "#Manager.v1_0_0.SerialConsole",
                "ConnectTypesSupported": [
                        "Telnet",
                        "SSH",
                        "IPMI"
                ],
                "MaxConcurrentSessions": 1,
                "ServiceEnabled": true
        },
        "SerialInterfaces": {
                "@odata.id": "/redfish/v1/Managers/BMC/SerialInterfaces"
        },
        "ServiceEntryPointUUID": "92384634-2938-2342-8820-489239905423",
        "Status": {
                "@odata.type": "#Resource.Status",
                "Health": "OK",
                "State": "Enabled"
        },
        "UUID": "58893887-8974-2487-2389-841168418919",
        "VirtualMedia": {
                "@odata.id": "/redfish/v1/Managers/BMC/VirtualMedia"
        }
}`

const testPathRtsPDUManagers_BMC_NICs = "/redfish/v1/Managers/BMC/NICs"

const testPayloadRtsPDUManagers_BMC_NICs = `
{
        "@odata.context": "/redfish/v1/$metadata#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "@odata.id": "/redfish/v1/Managers/BMC/NICs",
        "@odata.type": "#EthernetInterfaceCollection.EthernetInterfaceCollection",
        "Description": "Collection of EthernetInterfaces for this Manager",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/Managers/BMC/NICs/Dedicated"
                }
        ],
        "Members@odata.count": 1,
        "Name": "Ethernet Network Interface Collection",
        "Oem": {}
}`

const testPathRtsPDUManagers_BMC_NICs_Dedicated = "/redfish/v1/Managers/BMC/NICs/Dedicated"

const testPayloadRtsPDUManagers_BMC_NICs_Dedicated = `
{
        "@Redfish.Settings": {
                "@odata.type": "#Settings.v1_0_0.Settings",
                "ETag": "84ffcbb050ddc7fa9cddb59014546e59",
                "Messages": [
                        {
                                "@odata.type": "#Message.v1_0_0.Message",
                                "MessageId": "Base.1.0.SettingsFailed",
                                "RelatedProperties": [
                                        "#/IPv6Addresses/PrefixLength"
                                ]
                        }
                ],
                "SettingsObject": {
                        "@odata.id": "/redfish/v1/Managers/BMC/NICs/Dedicated/SD"
                },
                "Time": "2012-03-07T14:44.30-05:00"
        },
        "@odata.context": "/redfish/v1/$metadata#EthernetInterface.EthernetInterface",
        "@odata.id": "/redfish/v1/Managers/BMC/NICs/Dedicated",
        "@odata.type": "#EthernetInterface.v1_3_0.EthernetInterface",
        "AutoNeg": true,
        "Description": "Management Network Interface",
        "FQDN": "web483-bmc.dmtf.org",
        "FullDuplex": true,
        "HostName": "web483-bmc",
        "IPv4Addresses": [
                {
                        "@odata.type": "#IPAddresses.v1_0_0.IPv4Address",
                        "Address": "192.168.0.10",
                        "AddressOrigin": "DHCP",
                        "Gateway": "192.168.0.1",
                        "Oem": {},
                        "SubnetMask": "255.255.252.0"
                }
        ],
        "IPv6AddressPolicyTable": [
                {
                        "@odata.type": "#EthernetInterface.v1_0_0.IPv6AddressPolicyEntry",
                        "Label": 0,
                        "Precedence": 50,
                        "Prefix": "::1/128"
                }
        ],
        "IPv6Addresses": [
                {
                        "@odata.type": "#IPAddresses.v1_0_0.IPv6Address",
                        "Address": "fe80::1ec1:deff:fe6f:1e24",
                        "AddressOrigin": "SLAAC",
                        "AddressState": "Preferred",
                        "Oem": {},
                        "PrefixLength": 64
                }
        ],
        "IPv6DefaultGateway": "fe80::1ec1:deff:fe6f:1e24",
        "IPv6StaticAddresses": [
                {
                        "@odata.type": "#IPAddresses.v1_0_0.IPv6StaticAddress",
                        "Address": "fe80::1ec1:deff:fe6f:1e24",
                        "PrefixLength": 16
                }
        ],
        "Id": "Dedicated",
        "InterfaceEnabled": true,
        "LinkStatus": "LinkUp",
        "MACAddress": "23:11:8A:33:CF:EA",
        "MTUSize": 1500,
        "MaxIPv6StaticAddresses": 1,
        "Name": "Manager Ethernet Interface",
        "NameServers": [
                "names.dmtf.org"
        ],
        "Oem": {},
        "PermanentMACAddress": "23:11:8A:33:CF:EA",
        "SpeedMbps": 100,
        "Status": {
                "@odata.type": "#Resource.Status",
                "Health": "OK",
                "State": "Enabled"
        },
        "VLAN": {
                "@odata.type": "#VLanNetworkInterface.v1_0_0.VLAN",
                "VLANEnable": true,
                "VLANId": 101
        }
}`

const testPathRtsPDUPowerEquipment = "/redfish/v1/PowerEquipment"

const testPayloadRtsPDUPowerEquipment = `
{
        "@odata.id": "/redfish/v1/PowerEquipment",
        "@odata.type": "#PowerEquipment.v0_9_0.PowerEquipment",
        "FloorPDUs": {
                "@odata.id": "/redfish/v1/PowerEquipment/FloorPDUs"
        },
        "Id": "EnergyEquipment",
        "Links": {},
        "Name": "DCIM Energy Equipment",
        "RackPDUs": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs"
        },
        "Status": {
                "HealthRollup": "OK",
                "State": "Enabled"
        },
        "TransferSwitches": {
                "@odata.id": "/redfish/v1/PowerEquipment/TransferSwitches"
        },
        "UPSs": {
                "@odata.id": "/redfish/v1/PowerEquipment/UPSs"
        }
}`

const testPathRtsPDUPowerEquipment_RackPDUs = "/redfish/v1/PowerEquipment/RackPDUs"

const testPayloadRtsPDUPowerEquipment_RackPDUs = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs",
        "@odata.type": "#PowerDistributionCollection.PowerDistributionCollection",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1"
                }
        ],
        "Members@odata.count": 1,
        "Name": "RackPDU Collection"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1 = "/redfish/v1/PowerEquipment/RackPDUs/1"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1",
        "@odata.type": "#PowerDistribution.v0_9_0.PowerDistribution",
        "Alarms": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Alarms"
        },
        "AssetTag": "PDX-92381",
        "Branches": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches"
        },
        "CircuitSummary": {
                "ControlledOutlets": 8,
                "MonitoredBranches": 4,
                "MonitoredOutlets": 12,
                "MonitoredPhases": 3,
                "TotalBranches": 4,
                "TotalOutlets": 16,
                "TotalPhases": 3
        },
        "DateOfManufacture": "2017-01-11T08:00:00Z",
        "EquipmentType": "RackPDU",
        "FirmwareVersion": "4.3.0",
        "HardwareRevision": "1.03b",
        "Id": "1",
        "Location": {
                "Latitude": 122.67,
                "Longitude": 45.52,
                "Placement": {
                        "Row": "North"
                },
                "PostalAddress": {
                        "Building": "3A",
                        "City": "Portland",
                        "Country": "US",
                        "Floor": "2",
                        "Name": "DMTF Headquarters",
                        "PostalCode": "97204",
                        "Room": "213",
                        "Street": "1001 SW 5th Avenue",
                        "Territory": "OR"
                }
        },
        "Mains": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Mains"
        },
        "Manufacturer": "Contoso",
        "Metrics": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Metrics"
        },
        "Model": "ZAP4000",
        "Name": "RackPDU1",
        "OutletGroups": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/OutletGroups"
        },
        "Outlets": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets"
        },
        "PartNumber": "AA-23",
        "Sensors": {
                "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors"
        },
        "SerialNumber": "29347ZT536",
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "TriggeredAlarms": [
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Alarms/Overload",
                        "@odata.type": "#Alarm.v0_8_0.Alarm",
                        "Acknowledged": false,
                        "AlarmState": "Triggered",
                        "AutomaticReArm": true,
                        "Id": "Overload",
                        "Links": {
                                "Oem": {},
                                "RelatedSensor": {
                                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/ACMainPower"
                                }
                        },
                        "Message": "Rack PDU Overload Condition",
                        "MessageArgs": [
                                "58703"
                        ],
                        "MessageId": "DCIM.0.1.0.Overload",
                        "Name": "PDU Unit Overload",
                        "Oem": {},
                        "Severity": "Critical",
                        "TriggerTime": "2018-08-07T14:44:00Z"
                }
        ],
        "UUID": "32354641-4135-4332-4a35-313735303734"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets",
        "@odata.type": "#CircuitCollection.CircuitCollection",
        "Members": [
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A2"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A3"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A4"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A5"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B1"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B2"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B3"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C1"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C2"
                },
                {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C3"
                }
        ],
        "Members@odata.count": 11,
        "Name": "Outlet Circuit Collection"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A5 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A5"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A5 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A5",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Id": "A5",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/A"
                }
        },
        "Name": "Un-managed, Un-monitored Outlet A5, Branch Circuit A",
        "NominalVoltage": "AC120V",
        "OutletType": "NEMA_5_15R",
        "PhaseWiringType": "OnePhase3Wire",
        "RatedCurrentAmps": 20,
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_B1 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B1"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_B1 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B1",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B1/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B1/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentB1",
                "Name": "Outlet B1 Current",
                "PeakReading": 4.86,
                "Reading": 3.08,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyB1",
                "Name": "Outlet B1 Energy",
                "Reading": 36166,
                "ReadingUnits": "kW"
        },
        "Id": "B1",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/B"
                }
        },
        "Name": "Outlet B1, Branch Circuit B",
        "NominalVoltage": "AC120V",
        "OutletType": "NEMA_5_20R",
        "PhaseWiringType": "OnePhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentB1",
                        "Name": "Outlet B1 Current",
                        "PeakReading": 4.86,
                        "Reading": 3.08,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line2ToNeutral": {
                        "ApparentVA": 361.7,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerB1",
                        "Name": "Outlet B1 Power",
                        "PeakReading": 357.0,
                        "PowerFactor": 0.63,
                        "ReactiveVAR": 280.5,
                        "Reading": 228.3,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line2ToNeutral": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageB1",
                        "Name": "Outlet B1 Voltage L2N",
                        "Reading": 116.6,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 4,
        "PowerSensor": {
                "ApparentVA": 361.7,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerB1",
                "Name": "Outlet B1 Power",
                "PeakReading": 357.0,
                "PowerFactor": 0.63,
                "ReactiveVAR": 280.5,
                "Reading": 228.3,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 20,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageB1",
                "Name": "Outlet B1 Voltage L2N",
                "Reading": 116.6,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_B2 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B2"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_B2 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B2",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B2/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B2/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentB2",
                "Name": "Outlet B2 Current",
                "PeakReading": 2.25,
                "Reading": 1.58,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyB2",
                "Name": "Outlet B2 Energy",
                "Reading": 52941,
                "ReadingUnits": "kW"
        },
        "Id": "B2",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/B"
                }
        },
        "Name": "Outlet B2, Branch Circuit B",
        "NominalVoltage": "AC240V",
        "OutletType": "NEMA_5_15R",
        "PhaseWiringType": "TwoPhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentB2",
                        "Name": "Outlet B2 Current",
                        "PeakReading": 2.25,
                        "Reading": 1.58,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line2ToLine3": {
                        "ApparentVA": 321.6,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerB2",
                        "Name": "Outlet B2 Power",
                        "PeakReading": 458.1,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 321.6,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line2ToLine3": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageB2",
                        "Name": "Outlet B2 Voltage L23",
                        "Reading": 203.6,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 6,
        "PowerSensor": {
                "ApparentVA": 321.6,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerB2",
                "Name": "Outlet B2 Power",
                "PeakReading": 458.1,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 321.6,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 12,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageB2",
                "Name": "Outlet A2 Voltage L23",
                "Reading": 203.6,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A1 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A1 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentA1",
                "Name": "Outlet A1 Current",
                "PeakReading": 2.86,
                "Reading": 1.68,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyA1",
                "Name": "Outlet A1 Energy",
                "Reading": 36166,
                "ReadingUnits": "kW"
        },
        "FrequencySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/FrequencyA1",
                "Name": "Outlet A1 Frequency",
                "Reading": 60.0,
                "ReadingUnits": "Hz"
        },
        "Id": "A1",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/A"
                }
        },
        "Name": "Outlet A1, Branch Circuit A",
        "NominalVoltage": "AC120V",
        "OutletType": "NEMA_5_20R",
        "PhaseWiringType": "OnePhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentA1",
                        "Name": "Outlet A1 Current",
                        "PeakReading": 2.86,
                        "Reading": 1.68,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line1ToNeutral": {
                        "ApparentVA": 197.4,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerA1",
                        "Name": "Outlet A1 Power",
                        "PeakReading": 336.0,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 197.4,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line1ToNeutral": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageA1",
                        "Name": "Outlet A1 Voltage L1N",
                        "Reading": 117.5,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 4,
        "PowerSensor": {
                "ApparentVA": 197.4,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerA1",
                "Name": "Outlet A1 Power",
                "PeakReading": 336.0,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 197.4,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 20,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageA1",
                "Name": "Outlet A1 Voltage L1N",
                "Reading": 117.5,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A4 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A4"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A4 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A4",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A4/Outlet.PowerControl"
                }
        },
        "Id": "A4",
        "IndicatorLED": "Off",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/A"
                }
        },
        "Name": "Managed, Un-monitored Outlet A4, Branch Circuit A",
        "NominalVoltage": "AC120V",
        "OutletType": "NEMA_5_15R",
        "PhaseWiringType": "OnePhase3Wire",
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 10,
        "PowerState": "On",
        "RatedCurrentAmps": 20,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_B3 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B3"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_B3 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B3",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B3/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/B3/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentB3",
                "Name": "Outlet B3 Current",
                "PeakReading": 2.77,
                "Reading": 2.09,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyB3",
                "Name": "Outlet B3 Energy",
                "Reading": 61848,
                "ReadingUnits": "kW"
        },
        "Id": "B3",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/B"
                }
        },
        "Name": "Outlet B3, Branch Circuit B",
        "NominalVoltage": "AC240V",
        "OutletType": "NEMA_5_15R",
        "PhaseWiringType": "TwoPhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentB3",
                        "Name": "Outlet B3 Current",
                        "PeakReading": 2.77,
                        "Reading": 2.09,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line2ToLine3": {
                        "ApparentVA": 426.5,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerB3",
                        "Name": "Outlet B3 Power",
                        "PeakReading": 565.3,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 426.5,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line2ToLine3": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageB3",
                        "Name": "Outlet B3 Voltage L23",
                        "Reading": 204.1,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 10,
        "PowerSensor": {
                "ApparentVA": 426.5,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerB3",
                "Name": "Outlet B3 Power",
                "PeakReading": 565.3,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 426.5,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 12,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageB3",
                "Name": "Outlet B3 Voltage L23",
                "Reading": 204.1,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_C1 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C1"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_C1 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C1",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C1/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C1/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentC1",
                "Name": "Outlet C1 Current",
                "PeakReading": 0.62,
                "Reading": 0.48,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyC1",
                "Name": "Outlet C1 Energy",
                "Reading": 36166,
                "ReadingUnits": "kW"
        },
        "Id": "C1",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/C"
                }
        },
        "Name": "Outlet C1, Branch Circuit C",
        "NominalVoltage": "AC120V",
        "OutletType": "NEMA_L6_20R",
        "PhaseWiringType": "OnePhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentC1",
                        "Name": "Outlet C1 Current",
                        "PeakReading": 0.62,
                        "Reading": 0.48,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line3ToNeutral": {
                        "ApparentVA": 56.8,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerC1",
                        "Name": "Outlet C1 Power",
                        "PeakReading": 73.4,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 56.8,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line3ToNeutral": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageC1",
                        "Name": "Outlet C1 Voltage L3N",
                        "Reading": 118.4,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 4,
        "PowerSensor": {
                "ApparentVA": 56.8,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerC1",
                "Name": "Outlet C1 Power",
                "PeakReading": 73.4,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 56.8,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 20,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageC1",
                "Name": "Outlet C1 Voltage L3N",
                "Reading": 118.4,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_C2 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C2"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_C2 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C2",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C2/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C2/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentC2",
                "Name": "Outlet C2 Current",
                "PeakReading": 2.5,
                "Reading": 1.93,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyC2",
                "Name": "Outlet C2 Energy",
                "Reading": 52941,
                "ReadingUnits": "kW"
        },
        "Id": "C2",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/C"
                }
        },
        "Name": "Outlet C2, Branch Circuit C",
        "NominalVoltage": "AC240V",
        "OutletType": "NEMA_L6_20R",
        "PhaseWiringType": "TwoPhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentC2",
                        "Name": "Outlet C2 Current",
                        "PeakReading": 2.5,
                        "Reading": 1.93,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line3ToLine1": {
                        "ApparentVA": 394.1,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerC2",
                        "Name": "Outlet C2 Power",
                        "PeakReading": 510.5,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 394.1,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line3ToLine1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageC2",
                        "Name": "Outlet C2 Voltage L31",
                        "Reading": 204.2,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 6,
        "PowerSensor": {
                "ApparentVA": 394.1,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerC2",
                "Name": "Outlet C2 Power",
                "PeakReading": 510.5,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 394.1,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 12,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageC2",
                "Name": "Outlet C2 Voltage L31",
                "Reading": 204.2,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_C3 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C3"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_C3 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C3",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C3/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/C3/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentC3",
                "Name": "Outlet C3 Current",
                "PeakReading": 2.85,
                "Reading": 1.74,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyC3",
                "Name": "Outlet C3 Energy",
                "Reading": 61848,
                "ReadingUnits": "kW"
        },
        "Id": "C3",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/C"
                }
        },
        "Name": "Outlet C3, Branch Circuit C",
        "NominalVoltage": "AC240V",
        "OutletType": "NEMA_L6_20R",
        "PhaseWiringType": "TwoPhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line3": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentC3",
                        "Name": "Outlet C3 Current",
                        "PeakReading": 2.85,
                        "Reading": 1.74,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line3ToLine1": {
                        "ApparentVA": 355.4,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerC3",
                        "Name": "Outlet C3 Power",
                        "PeakReading": 582.3,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 355.4,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line3ToLine1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageC3",
                        "Name": "Outlet C3 Voltage L31",
                        "Reading": 204.3,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 10,
        "PowerSensor": {
                "ApparentVA": 355.4,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerC3",
                "Name": "Outlet C3 Power",
                "PeakReading": 582.3,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 355.4,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 12,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageC3",
                "Name": "Outlet C3 Voltage L31",
                "Reading": 204.3,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A2 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A2"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A2 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A2",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A2/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A2/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentA2",
                "Name": "Outlet A2 Current",
                "PeakReading": 2.0,
                "Reading": 1.9,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyA2",
                "Name": "Outlet A2 Energy",
                "Reading": 52941,
                "ReadingUnits": "kW"
        },
        "FrequencySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/FrequencyA2",
                "Name": "Outlet A2 Frequency",
                "Reading": 60.0,
                "ReadingUnits": "Hz"
        },
        "Id": "A2",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/A"
                }
        },
        "Name": "Outlet A2, Branch Circuit A",
        "NominalVoltage": "AC240V",
        "OutletType": "C13",
        "PhaseWiringType": "TwoPhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentA2",
                        "Name": "Outlet A2 Current",
                        "PeakReading": 2.0,
                        "Reading": 1.9,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line1ToLine2": {
                        "ApparentVA": 384.5,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerA2",
                        "Name": "Outlet A2 Power",
                        "PeakReading": 404.8,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 384.5,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line1ToLine2": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageA2",
                        "Name": "Outlet A2 Voltage L12",
                        "Reading": 202.4,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 6,
        "PowerSensor": {
                "ApparentVA": 384.5,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerA2",
                "Name": "Outlet A2 Power",
                "PeakReading": 404.8,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 384.5,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 12,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageA2",
                "Name": "Outlet A2 Voltage L12",
                "Reading": 202.4,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`

const testPathRtsPDUPowerEquipment_RackPDUs_1_Outlets_A3 = "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A3"

const testPayloadRtsPDUPowerEquipment_RackPDUs_1_Outlets_A3 = `
{
        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A3",
        "@odata.type": "#Outlet.v0_9_0.Outlet",
        "Actions": {
                "#Outlet.PowerControl": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A3/Outlet.PowerControl"
                },
                "#Outlet.ResetStatistics": {
                        "target": "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A3/Outlet.ResetStatistics"
                }
        },
        "CurrentSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentA3",
                "Name": "Outlet A3 Current",
                "PeakReading": 2.5,
                "Reading": 1.73,
                "ReadingUnits": "A"
        },
        "EnergySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/EnergyA3",
                "Name": "Outlet A3 Energy",
                "Reading": 61848,
                "ReadingUnits": "kW"
        },
        "FrequencySensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/FrequencyA3",
                "Name": "Outlet A3 Frequency",
                "Reading": 60.0,
                "ReadingUnits": "Hz"
        },
        "Id": "A3",
        "IndicatorLED": "Lit",
        "Links": {
                "BranchCircuit": {
                        "@odata.id": "/redfish/v1/PowerEquipment/RackPDUs/1/Branches/A"
                }
        },
        "Name": "Outlet A3, Branch Circuit A",
        "NominalVoltage": "AC240V",
        "OutletType": "C13",
        "PhaseWiringType": "TwoPhase3Wire",
        "PolyPhaseCurrentSensors": {
                "Line1": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/CurrentA3",
                        "Name": "Outlet A3 Current",
                        "PeakReading": 2.5,
                        "Reading": 1.73,
                        "ReadingUnits": "A"
                }
        },
        "PolyPhasePowerSensors": {
                "Line1ToLine2": {
                        "ApparentVA": 349.9,
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerA3",
                        "Name": "Outlet A3 Power",
                        "PeakReading": 505.7,
                        "PowerFactor": 1.0,
                        "ReactiveVAR": 0.0,
                        "Reading": 349.9,
                        "ReadingUnits": "W"
                }
        },
        "PolyPhaseVoltageSensors": {
                "Line1ToLine2": {
                        "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageA3",
                        "Name": "Outlet A3 Voltage L12",
                        "Reading": 202.3,
                        "ReadingUnits": "V"
                }
        },
        "PowerEnabled": true,
        "PowerOffDelaySeconds": 0,
        "PowerOnDelaySeconds": 10,
        "PowerSensor": {
                "ApparentVA": 349.9,
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/PowerA3",
                "Name": "Outlet A3 Power",
                "PeakReading": 505.7,
                "PowerFactor": 1.0,
                "ReactiveVAR": 0.0,
                "Reading": 349.9,
                "ReadingUnits": "W"
        },
        "PowerState": "On",
        "RatedCurrentAmps": 12,
        "Status": {
                "Health": "OK",
                "State": "Enabled"
        },
        "VoltageSensor": {
                "DataSourceUri": "/redfish/v1/PowerEquipment/RackPDUs/1/Sensors/VoltageA3",
                "Name": "Outlet A3 Voltage L12",
                "Reading": 202.3,
                "ReadingUnits": "V"
        },
        "VoltageType": "AC"
}`
