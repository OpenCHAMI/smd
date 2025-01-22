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

package sharedtest

import (
	"encoding/json"

	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
)

//////////////////////////////////////////////////////////////////////////////
// RedfishEndpoints - Dummy examples for testing
//////////////////////////////////////////////////////////////////////////////

var TestRedfishEndpointNodeBMC1 = sm.RedfishEndpoint{
	RedfishEPDescription: rf.RedfishEPDescription{ // Note: Embedded struct
		ID:             "x0c0s14b0",
		Type:           "NodeBMC",
		Hostname:       "10.10.255.11",
		Domain:         "local",
		FQDN:           "10.10.255.11",
		Enabled:        true,
		UUID:           "d4c6d22f-6983-42d8-8e6e-e1fd6d675c17",
		User:           "root",
		Password:       "********",
		RediscOnUpdate: true,
		DiscInfo: rf.DiscoveryInfo{
			LastStatus: rf.DiscoverOK,
		},
	},
}

var TestRedfishEndpointNodeBMC2 = sm.RedfishEndpoint{
	RedfishEPDescription: rf.RedfishEPDescription{
		ID:             "x11c0s1b0",
		Type:           "NodeBMC",
		Hostname:       "10.10.0.1",
		Domain:         "local",
		FQDN:           "10.10.0.1",
		Enabled:        true,
		UUID:           "d446d22f-ffff-ffff-8e6e-e1fd6d675c17",
		User:           "root",
		Password:       "********",
		RediscOnUpdate: true,
		DiscInfo: rf.DiscoveryInfo{
			LastStatus: rf.NotYetQueried,
		},
	},
}

var TestRedfishEndpointRouterBMC1 = sm.RedfishEndpoint{
	RedfishEPDescription: rf.RedfishEPDescription{
		ID:             "x11c0r1b0",
		Type:           "RouterBMC",
		Hostname:       "10.10.99.1",
		Domain:         "here.com",
		FQDN:           "10.10.99.1",
		Enabled:        true,
		UUID:           "d446d22f-aaaa-aaaa-8e6e-e1fd6d675c17",
		User:           "root",
		Password:       "********",
		RediscOnUpdate: true,
		DiscInfo: rf.DiscoveryInfo{
			LastStatus: rf.DiscoverOK,
		},
	},
}

var TestRedfishEndpointChassisBMC1 = sm.RedfishEndpoint{
	RedfishEPDescription: rf.RedfishEPDescription{
		ID:             "x6c1b0",
		Type:           "ChassisBMC",
		Hostname:       "10.6.0.1",
		Domain:         "local",
		FQDN:           "10.6.0.1",
		Enabled:        false,
		UUID:           "d446d22f-bbbb-bbbb-8e6e-e1fd6d675c17",
		User:           "root",
		Password:       "********",
		RediscOnUpdate: true,
		DiscInfo: rf.DiscoveryInfo{
			LastStatus: rf.EndpointNotEnabled,
		},
	},
}

var TestRedfishEndpointArray = sm.RedfishEndpointArray{
	RedfishEndpoints: []*sm.RedfishEndpoint{
		&TestRedfishEndpointNodeBMC1,
		&TestRedfishEndpointNodeBMC2,
		&TestRedfishEndpointRouterBMC1,
		&TestRedfishEndpointChassisBMC1,
	},
}

var TestRedfishEndpointArrayNodeBMCs = sm.RedfishEndpointArray{
	RedfishEndpoints: []*sm.RedfishEndpoint{
		&TestRedfishEndpointNodeBMC1,
		&TestRedfishEndpointNodeBMC2,
	},
}

var TestRedfishEndpointArrayNodeBMC1 = sm.RedfishEndpointArray{
	RedfishEndpoints: []*sm.RedfishEndpoint{
		&TestRedfishEndpointNodeBMC1,
	},
}

var TestRedfishEndpointArrayNodeBMC2 = sm.RedfishEndpointArray{
	RedfishEndpoints: []*sm.RedfishEndpoint{
		&TestRedfishEndpointNodeBMC2,
	},
}

var TestRedfishEndpointArrayRouterBMCs = sm.RedfishEndpointArray{
	RedfishEndpoints: []*sm.RedfishEndpoint{
		&TestRedfishEndpointRouterBMC1,
	},
}

var TestRedfishEndpointsChassisBMCs = sm.RedfishEndpointArray{
	RedfishEndpoints: []*sm.RedfishEndpoint{
		&TestRedfishEndpointChassisBMC1,
	},
}

//////////////////////////////////////////////////////////////////////////////
// ServiceEndpoints - Dummy examples for testing
//////////////////////////////////////////////////////////////////////////////

var TestServiceEndpointUpdate1 = sm.ServiceEndpoint{
	ServiceDescription: rf.ServiceDescription{ // Note: Embedded struct
		RfEndpointID: "x0c0s21b0",
		RedfishType:  "UpdateService",
		UUID:         "",
		OdataID:      "/redfish/v1/UpdateService",
	},
	RfEndpointFQDN: "10.4.0.8",
	URL:            "10.4.0.8/redfish/v1/UpdateService",
	ServiceInfo:    json.RawMessage(`{}`),
}

var TestServiceEndpointUpdate2 = sm.ServiceEndpoint{
	ServiceDescription: rf.ServiceDescription{ // Note: Embedded struct
		RfEndpointID: "x0c0s22b0",
		RedfishType:  "UpdateService",
		UUID:         "",
		OdataID:      "/redfish/v1/UpdateService",
	},
	RfEndpointFQDN: "10.4.0.9",
	URL:            "10.4.0.9/redfish/v1/UpdateService",
	ServiceInfo:    json.RawMessage(`{}`),
}

var TestServiceEndpointEvent1 = sm.ServiceEndpoint{
	ServiceDescription: rf.ServiceDescription{ // Note: Embedded struct
		RfEndpointID: "x0c0s21b0",
		RedfishType:  "EventService",
		UUID:         "",
		OdataID:      "/redfish/v1/EventService",
	},
	RfEndpointFQDN: "10.4.0.8",
	URL:            "10.4.0.8/redfish/v1/EventService",
	ServiceInfo:    json.RawMessage(`{}`),
}

var TestServiceEndpointEvent2 = sm.ServiceEndpoint{
	ServiceDescription: rf.ServiceDescription{ // Note: Embedded struct
		RfEndpointID: "x0c0s22b0",
		RedfishType:  "EventService",
		UUID:         "",
		OdataID:      "/redfish/v1/EventService",
	},
	RfEndpointFQDN: "10.4.0.9",
	URL:            "10.4.0.9/redfish/v1/EventService",
	ServiceInfo:    json.RawMessage(`{}`),
}

var TestServiceEndpointArray = sm.ServiceEndpointArray{
	ServiceEndpoints: []*sm.ServiceEndpoint{
		&TestServiceEndpointUpdate1,
		&TestServiceEndpointUpdate2,
		&TestServiceEndpointEvent1,
		&TestServiceEndpointEvent2,
	},
}

var TestServiceEndpointArrayUpdates = sm.ServiceEndpointArray{
	ServiceEndpoints: []*sm.ServiceEndpoint{
		&TestServiceEndpointUpdate1,
		&TestServiceEndpointUpdate2,
	},
}

var TestServiceEndpointArrayUpdate1 = sm.ServiceEndpointArray{
	ServiceEndpoints: []*sm.ServiceEndpoint{
		&TestServiceEndpointUpdate1,
	},
}

var TestServiceEndpointArrayRFEP = sm.ServiceEndpointArray{
	ServiceEndpoints: []*sm.ServiceEndpoint{
		&TestServiceEndpointUpdate1,
		&TestServiceEndpointEvent1,
	},
}

//////////////////////////////////////////////////////////////////////////////
// ComponentEndpoints - Dummy examples for testing
//////////////////////////////////////////////////////////////////////////////

var trueVal = true
var falseVal = false

var TestCompEndpointNodeBMC = sm.ComponentEndpoint{
	ComponentDescription: rf.ComponentDescription{
		ID:             "x666c0s46b0",
		Type:           "NodeBMC",
		RedfishType:    "Manager",
		RedfishSubtype: "BMC",
		MACAddr:        "a4:bf:01:3f:8d:ae",
		UUID:           "ce705079-e8d9-e111-8534-cb6990e467cc",
		OdataID:        "/redfish/v1/Managers/BMC",
		RfEndpointID:   "x666c0s46b0",
	},
	RfEndpointFQDN:        "10.100.164.164",
	URL:                   "10.100.164.164/redfish/v1/Managers/BMC",
	ComponentEndpointType: "ComponentEndpointManager",
	RedfishManagerInfo: &rf.ComponentManagerInfo{
		Name: "Manager",
		Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
			AllowableValues: []string{"ForceRestart"},
			Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
		EthNICInfo: []*rf.EthernetNICInfo{
			&rf.EthernetNICInfo{RedfishId: "1",
				Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
				Description:      "Network Interface on the Baseboard Management Controller",
				Hostname:         "BMCA4BF013F8DAC",
				InterfaceEnabled: &falseVal,
				MACAddress:       "a4:bf:01:3f:8d:ac"},
			&rf.EthernetNICInfo{RedfishId: "2",
				Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
				Description:      "Network Interface on the Baseboard Management Controller",
				Hostname:         "BMCA4BF013F8DAC",
				InterfaceEnabled: &falseVal,
				MACAddress:       "a4:bf:01:3f:8d:ad"},
			&rf.EthernetNICInfo{RedfishId: "3",
				Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
				Description:      "Network Interface on the Baseboard Management Controller",
				Hostname:         "BMCA4BF013F8DAC",
				InterfaceEnabled: &trueVal,
				MACAddress:       "a4:bf:01:3f:8d:ae"}}},
}

var SampleCompEndpoints = []*sm.ComponentEndpoint{
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4",
			Type:           "Chassis",
			RedfishType:    "Chassis",
			RedfishSubtype: "Enclosure",
			OdataID:        "/redfish/v1/Chassis/Enclosure",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Enclosure",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Enclosure",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4b0",
			Type:           "ChassisBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "EnclosureManager",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "BMC",
			Actions: &rf.ManagerActions{
				ManagerReset: rf.ActionReset{
					AllowableValues: []string{
						"ForceRestart",
						"ForceEraseNetworkReload"},
					Target: "/redfish/v1/Managers/BMC/Actions/Manager.Reset"},
				OEM: &rf.ManagerActionsOEM{
					ManagerFactoryReset: &rf.ActionFactoryReset{
						AllowableValues: []string{"ResetAll"},
						Target:          "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset"},
					CrayProcessSchedule: &rf.ActionNamed{
						AllowableValues: []string{
							"memtest",
							"cpuburn"},
						Target: "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"}}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r0",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif0",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif0",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif0",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif0/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r1",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif1",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif1",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif1",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif1/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r2",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif2",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif2",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif2",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif2/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r3",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif3",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif3",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif3",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif3/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r4",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif4",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif4",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif4",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif4/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r5",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif5",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif5",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif5",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif5/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r6",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif6",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif6",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif6",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif6/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r7",
			Type:           "RouterModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Perif7",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Perif7",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Perif7",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Perif7/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s0",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade0",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade0",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade0",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade0/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s1",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade1",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade1",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade1",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade1/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s2",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade2",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade2",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade2",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade2/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s3",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade3",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade3",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade3",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade3/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s4",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade4",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade4",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade4",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade4/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s5",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade5",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade5",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade5",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade5/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s6",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade6",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade6",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade6",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade6/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s7",
			Type:           "ComputeModule",
			RedfishType:    "Chassis",
			RedfishSubtype: "Blade",
			OdataID:        "/redfish/v1/Chassis/Blade7",
			RfEndpointID:   "x1c4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]",
		URL:                   "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]/redfish/v1/Chassis/Blade7",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Blade7",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Blade7/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r3b0",
			Type:           "RouterBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "EnclosureManager",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x1c4r3b0",
		},
		RfEndpointFQDN:        "[fc00:0:100:10:0:2ff:fe02:6100]",
		URL:                   "[fc00:0:100:10:0:2ff:fe02:6100]/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "BMC",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{
					"ForceRestart",
					"ForceEraseNetworkReload"},
				Target: "/redfish/v1/Managers/BMC/Actions/Manager.Reset"},
				OEM: &rf.ManagerActionsOEM{ManagerFactoryReset: &rf.ActionFactoryReset{
					AllowableValues: []string{"ResetAll"},
					Target:          "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset"},
					CrayProcessSchedule: &rf.ActionNamed{
						AllowableValues: []string{
							"memtest",
							"cpuburn"},
						Target: "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"}}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4r3e0",
			Type:           "HSNBoard",
			RedfishType:    "Chassis",
			RedfishSubtype: "Enclosure",
			OdataID:        "/redfish/v1/Chassis/Enclosure",
			RfEndpointID:   "x1c4r3b0",
		},
		RfEndpointFQDN:        "[fc00:0:100:10:0:2ff:fe02:6100]",
		URL:                   "[fc00:0:100:10:0:2ff:fe02:6100]/redfish/v1/Chassis/Enclosure",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Enclosure",
			Actions: &rf.ChassisActions{ChassisReset: rf.ActionReset{
				AllowableValues: []string{
					"On",
					"ForceOff",
					"Off"},
				Target: "/redfish/v1/Chassis/Enclosure/Actions/Chassis.Reset"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s4b0n0",
			Type:           "Node",
			OdataID:        "/redfish/v1/Systems/Node0",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "00:40:a6:82:f6:2b",
			RfEndpointID:   "x1c4s4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f609]",
		URL:                   "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f609]/redfish/v1/Systems/Node0",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "Node0",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{
					"Off",
					"ForceOff",
					"On",
				}}},
			EthNICInfo: []*rf.EthernetNICInfo{
				{
					Oid:                 "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet",
					MACAddress:          "00:40:a6:82:f6:2b",
					PermanentMACAddress: "00:40:a6:82:f6:2b",
					Description:         "Node Maintenance Network",
					RedfishId:           "ManagementEthernet"},
			},
		}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s4b0n1",
			Type:           "Node",
			OdataID:        "/redfish/v1/Systems/Node1",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "00:40:a6:82:f6:0b",
			RfEndpointID:   "x1c4s4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f609]",
		URL:                   "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f609]/redfish/v1/Systems/Node1",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "Node1",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{
					"Off",
					"ForceOff",
					"On",
				}}},
			EthNICInfo: []*rf.EthernetNICInfo{
				{
					Oid:                 "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet",
					MACAddress:          "00:40:a6:82:f6:0b",
					PermanentMACAddress: "00:40:a6:82:f6:0b",
					Description:         "Node Maintenance Network",
					RedfishId:           "ManagementEthernet"},
			},
		}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s4b1n0",
			Type:           "Node",
			OdataID:        "/redfish/v1/Systems/Node0",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "00:40:a6:82:f9:8a",
			RfEndpointID:   "x1c4s4b1",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f001]",
		URL:                   "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f001]/redfish/v1/Systems/Node0",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "Node0",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{
					"Off",
					"ForceOff",
					"On",
				}}},
			EthNICInfo: []*rf.EthernetNICInfo{
				{
					Oid:                 "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet",
					MACAddress:          "00:40:a6:82:f9:8a",
					PermanentMACAddress: "00:40:a6:82:f9:8a",
					Description:         "Node Maintenance Network",
					RedfishId:           "ManagementEthernet"},
			},
		}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s4b1n1",
			Type:           "Node",
			OdataID:        "/redfish/v1/Systems/Node1",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "00:40:a6:82:f7:7b",
			RfEndpointID:   "x1c4s4b1",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f001]",
		URL:                   "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f001]/redfish/v1/Systems/Node1",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "Node1",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{
					"Off",
					"ForceOff",
					"On",
				}}},
			EthNICInfo: []*rf.EthernetNICInfo{
				{
					Oid:                 "/redfish/v1/Systems/Node1/EthernetInterfaces/ManagementEthernet",
					MACAddress:          "00:40:a6:82:f7:7b",
					PermanentMACAddress: "00:40:a6:82:f7:7b",
					Description:         "Node Maintenance Network",
					RedfishId:           "ManagementEthernet"},
			},
		}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s4b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "EnclosureManager",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x1c4s4b0",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f609]",
		URL:                   "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f609]/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "BMC",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{
					"ForceRestart",
					"ForceEraseNetworkReload"},
				Target: "/redfish/v1/Managers/BMC/Actions/Manager.Reset"},
				OEM: &rf.ManagerActionsOEM{ManagerFactoryReset: &rf.ActionFactoryReset{
					AllowableValues: []string{"ResetAll"},
					Target:          "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset"},
					CrayProcessSchedule: &rf.ActionNamed{
						AllowableValues: []string{
							"memtest",
							"cpuburn"},
						Target: "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"},
				}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x1c4s4b1",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "EnclosureManager",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x1c4s4b1",
		},
		RfEndpointFQDN:        "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f001]",
		URL:                   "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f001]/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "BMC",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{
					"ForceRestart",
					"ForceEraseNetworkReload"},
				Target: "/redfish/v1/Managers/BMC/Actions/Manager.Reset"},
				OEM: &rf.ManagerActionsOEM{ManagerFactoryReset: &rf.ActionFactoryReset{
					AllowableValues: []string{"ResetAll"},
					Target:          "/redfish/v1/Managers/Self/Actions/Manager.FactoryReset"},
					CrayProcessSchedule: &rf.ActionNamed{
						AllowableValues: []string{
							"memtest",
							"cpuburn"},
						Target: "/redfish/v1/Managers/BMC/Actions/Oem/CrayProcess.Schedule"},
				}}}},
	//
	// Intel Nodes and components
	//
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s28e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x99c0s28b0",
		},
		RfEndpointFQDN:        "10.4.0.6",
		URL:                   "10.4.0.6/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x666c0s46b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:3f:8d:ae",
			UUID:           "ce705079-e8d9-e111-8534-cb6990e467cc",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x666c0s46b0",
		},
		RfEndpointFQDN:        "10.100.164.164",
		URL:                   "10.100.164.164/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
					Description:      "Network Interface on the Baseboard Management Controller",
					Hostname:         "BMCA4BF013F8DAC",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:3f:8d:ac"},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
					Description:      "Network Interface on the Baseboard Management Controller",
					Hostname:         "BMCA4BF013F8DAC",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:3f:8d:ad"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
					Description:      "Network Interface on the Baseboard Management Controller",
					Hostname:         "BMCA4BF013F8DAC",
					InterfaceEnabled: &trueVal,
					MACAddress:       "a4:bf:01:3f:8d:ae"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x666c0s46b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:3f:8d:aa",
			UUID:           "0a96c51a-8ee8-11e8-ab21-a4bf013f8daa",
			OdataID:        "/redfish/v1/Systems/............",
			RfEndpointID:   "x666c0s46b0",
		},
		RfEndpointFQDN:        "10.100.164.164",
		URL:                   "10.100.164.164/redfish/v1/Systems/............",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"PushPowerButton",
					"On",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"ForceOn",
					"ForceOff"},
				Target: "/redfish/v1/Systems/............/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:3f:8d:aa"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:3f:8d:ab"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x666c0s46e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x666c0s46b0",
		},
		RfEndpointFQDN:        "10.100.164.164",
		URL:                   "10.100.164.164/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s12b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2e:ac:f2",
			UUID:           "bcd8b346-4b13-a174-780d-0583b4bccd94",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x99c0s12b0",
		},
		RfEndpointFQDN:        "10.4.0.17",
		URL:                   "10.4.0.17/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
					Hostname:         "BMCA4BF012EACF0",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2e:ac:f0"},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
					Hostname:         "BMCA4BF012EACF0",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2e:ac:f1"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
					Hostname:         "BMCA4BF012EACF0",
					InterfaceEnabled: &trueVal,
					MACAddress:       "a4:bf:01:2e:ac:f2"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s12b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:2e:ac:ee",
			UUID:           "033e4bc4-1feb-11e8-ab21-a4bf012eacee",
			OdataID:        "/redfish/v1/Systems/QSBP80904410",
			RfEndpointID:   "x99c0s12b0",
		},
		RfEndpointFQDN:        "10.4.0.17",
		URL:                   "10.4.0.17/redfish/v1/Systems/QSBP80904410",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"PushPowerButton",
					"On",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"ForceOn",
					"ForceOff"},
				Target: "/redfish/v1/Systems/QSBP80904410/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2e:ac:ee"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2e:ac:ef"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s12e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x99c0s12b0",
		},
		RfEndpointFQDN:        "10.4.0.17",
		URL:                   "10.4.0.17/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s13b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2e:a0:13",
			UUID:           "48680f98-e6d1-9f61-5f64-8523c2e07332",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x99c0s13b0",
		},
		RfEndpointFQDN:        "10.4.0.16",
		URL:                   "10.4.0.16/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
					Hostname:         "BMCA4BF012EA011",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2e:a0:12"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
					Hostname:         "BMCA4BF012EA011",
					InterfaceEnabled: &trueVal,
					MACAddress:       "a4:bf:01:2e:a0:13"},
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
					Hostname:         "BMCA4BF012EA011",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2e:a0:11"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s13b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:2e:a0:0f",
			UUID:           "67dfa9ad-1db4-11e8-ab21-a4bf012ea00f",
			OdataID:        "/redfish/v1/Systems/QSBP80903751",
			RfEndpointID:   "x99c0s13b0",
		},
		RfEndpointFQDN:        "10.4.0.16",
		URL:                   "10.4.0.16/redfish/v1/Systems/QSBP80903751",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"PushPowerButton",
					"On",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"ForceOn",
					"ForceOff"},
				Target: "/redfish/v1/Systems/QSBP80903751/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2e:a0:0f"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2e:a0:10"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s13e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x99c0s13b0",
		},
		RfEndpointFQDN:        "10.4.0.16",
		URL:                   "10.4.0.16/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s15b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2b:74:ce",
			UUID:           "f26e71b5-84c3-3047-31e4-affc75e8a841",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x99c0s15b0",
		},
		RfEndpointFQDN:        "10.4.0.14",
		URL:                   "10.4.0.14/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
					Hostname:         "BMCA4BF012B74CC",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:74:cd"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
					Hostname:         "BMCA4BF012B74CC",
					InterfaceEnabled: &trueVal,
					MACAddress:       "a4:bf:01:2b:74:ce"},
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
					Hostname:         "BMCA4BF012B74CC",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:74:cc"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s15b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:2b:74:ca",
			UUID:           "98ed4694-e111-11e7-ab21-a4bf012b74ca",
			OdataID:        "/redfish/v1/Systems/QSBP75002224",
			RfEndpointID:   "x99c0s15b0",
		},
		RfEndpointFQDN:        "10.4.0.14",
		URL:                   "10.4.0.14/redfish/v1/Systems/QSBP75002224",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"PushPowerButton",
					"On",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"ForceOn",
					"ForceOff"},
				Target: "/redfish/v1/Systems/QSBP75002224/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:74:ca"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:74:cb"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s15e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x99c0s15b0",
		},
		RfEndpointFQDN:        "10.4.0.14",
		URL:                   "10.4.0.14/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s16b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2b:72:94",
			UUID:           "e14c6e67-06b3-fb9e-1ff8-2a52510edc6b",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x99c0s16b0",
		},
		RfEndpointFQDN:        "10.4.0.13",
		URL:                   "10.4.0.13/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
					Hostname:         "BMCA4BF012B7292",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:72:92"},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
					Hostname:         "BMCA4BF012B7292",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:72:93"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
					Hostname:         "BMCA4BF012B7292",
					InterfaceEnabled: &trueVal,
					MACAddress:       "a4:bf:01:2b:72:94"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s16b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:2b:72:90",
			UUID:           "1b31cc86-e0cf-11e7-ab21-a4bf012b7290",
			OdataID:        "/redfish/v1/Systems/QSBP75002110",
			RfEndpointID:   "x99c0s16b0",
		},
		RfEndpointFQDN:        "10.4.0.13",
		URL:                   "10.4.0.13/redfish/v1/Systems/QSBP75002110",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"PushPowerButton",
					"On",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"ForceOn",
					"ForceOff"},
				Target: "/redfish/v1/Systems/QSBP75002110/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:72:90"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:72:91"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s16e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x99c0s16b0",
		},
		RfEndpointFQDN:        "10.4.0.13",
		URL:                   "10.4.0.13/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s17b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2b:6a:e7",
			UUID:           "672bc159-e349-24b0-f83f-750441811728",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x99c0s17b0",
		},
		RfEndpointFQDN:        "10.4.0.12",
		URL:                   "10.4.0.12/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
					Hostname:         "BMCA4BF012B6AE5",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:6a:e6"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
					Hostname:         "BMCA4BF012B6AE5",
					InterfaceEnabled: &trueVal,
					MACAddress:       "a4:bf:01:2b:6a:e7"},
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
					Hostname:         "BMCA4BF012B6AE5",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:6a:e5"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s17b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:2b:6a:e3",
			UUID:           "7371837e-e03f-11e7-ab21-a4bf012b6ae3",
			OdataID:        "/redfish/v1/Systems/QSBP75001717",
			RfEndpointID:   "x99c0s17b0",
		},
		RfEndpointFQDN:        "10.4.0.12",
		URL:                   "10.4.0.12/redfish/v1/Systems/QSBP75001717",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"PushPowerButton",
					"On",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"ForceOn",
					"ForceOff"},
				Target: "/redfish/v1/Systems/QSBP75001717/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:6a:e3"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:6a:e4"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s17e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x99c0s17b0",
		},
		RfEndpointFQDN:        "10.4.0.12",
		URL:                   "10.4.0.12/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s18b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2b:6a:c9",
			UUID:           "a25ca215-3612-c20c-3dce-31b4142cb90a",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x99c0s18b0",
		},
		RfEndpointFQDN:        "10.4.0.11",
		URL:                   "10.4.0.11/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/1",
					Hostname:         "BMCA4BF012B6AC7",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:6a:c7"},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/2",
					Hostname:         "BMCA4BF012B6AC7",
					InterfaceEnabled: &falseVal,
					MACAddress:       "a4:bf:01:2b:6a:c8"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:              "/redfish/v1/Managers/BMC/EthernetInterfaces/3",
					Hostname:         "BMCA4BF012B6AC7",
					InterfaceEnabled: &trueVal,
					MACAddress:       "a4:bf:01:2b:6a:c9"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s18b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:2b:6a:c5",
			UUID:           "73718375-e03f-11e7-ab21-a4bf012b6ac5",
			OdataID:        "/redfish/v1/Systems/QSBP75001711",
			RfEndpointID:   "x99c0s18b0",
		},
		RfEndpointFQDN:        "10.4.0.11",
		URL:                   "10.4.0.11/redfish/v1/Systems/QSBP75001711",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"PushPowerButton",
					"On",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"ForceOn",
					"ForceOff"},
				Target: "/redfish/v1/Systems/QSBP75001711/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:6a:c5"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:6a:c6"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s18e0",
			Type:           "NodeEnclosure",
			RedfishType:    "Chassis",
			RedfishSubtype: "RackMount",
			OdataID:        "/redfish/v1/Chassis/RackMount",
			RfEndpointID:   "x99c0s18b0",
		},
		RfEndpointFQDN:        "10.4.0.11",
		URL:                   "10.4.0.11/redfish/v1/Chassis/RackMount",
		ComponentEndpointType: "ComponentEndpointChassis",
		RedfishChassisInfo: &rf.ComponentChassisInfo{
			Name: "Computer System Chassis"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s28b0",
			Type:           "NodeBMC",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2b:75:8c",
			UUID:           "2ea2b238-7290-319c-5e71-13f2e5044433",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x99c0s28b0",
		},
		RfEndpointFQDN:        "10.4.0.6",
		URL:                   "10.4.0.6/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:         "/redfish/v1/Managers/BMC/NICs/1",
					Description: "Management Network Interface",
					Hostname:    "BMCA4BF012B758A",
					MACAddress:  "a4:bf:01:2b:75:8a"},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:         "/redfish/v1/Managers/BMC/NICs/2",
					Description: "Management Network Interface",
					Hostname:    "BMCA4BF012B758A",
					MACAddress:  "a4:bf:01:2b:75:8b"},
				&rf.EthernetNICInfo{RedfishId: "3",
					Oid:         "/redfish/v1/Managers/BMC/NICs/3",
					Description: "Management Network Interface",
					Hostname:    "BMCA4BF012B758A",
					MACAddress:  "a4:bf:01:2b:75:8c"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x99c0s28b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			MACAddr:        "a4:bf:01:2b:75:88",
			UUID:           "98ed467b-e111-11e7-ab21-a4bf012b7588",
			OdataID:        "/redfish/v1/Systems/QSBP75002262",
			RfEndpointID:   "x99c0s28b0",
		},
		RfEndpointFQDN:        "10.4.0.6",
		URL:                   "10.4.0.6/redfish/v1/Systems/QSBP75002262",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "S2600BPB",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"GracefulShutdown",
					"GracefulRestart",
					"ForceRestart",
					"Nmi"},
				Target: "/redfish/v1/Systems/QSBP75002262/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 1, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:75:88"},
				&rf.EthernetNICInfo{RedfishId: "",
					Oid:         "",
					Description: "Missing interface 2, MAC computed via workaround",
					MACAddress:  "a4:bf:01:2b:75:89"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x0m0",
			Type:           "CabinetPDUController",
			RedfishType:    "Manager",
			RedfishSubtype: "BMC",
			MACAddr:        "a4:bf:01:2b:75:8c",
			UUID:           "58893887-8974-2487-2389-841168418919",
			OdataID:        "/redfish/v1/Managers/BMC",
			RfEndpointID:   "x0m0",
		},
		RfEndpointFQDN:        "127.0.0.1",
		URL:                   "127.0.0.1/redfish/v1/Managers/BMC",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishManagerInfo: &rf.ComponentManagerInfo{
			Name: "Manager",
			Actions: &rf.ManagerActions{ManagerReset: rf.ActionReset{
				AllowableValues: []string{"ForceRestart", "GracefulRestart"},
				Target:          "/redfish/v1/Managers/BMC/Actions/Manager.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{
					RedfishId:   "Dedicated",
					FQDN:        "web483-bmc.dmtf.org",
					Oid:         "/redfish/v1/Managers/BMC/NICs/Dedicated",
					Description: "Management Network Interface",
					Hostname:    "web483-bmc",
					MACAddress:  "23:11:8a:33:cf:ea"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x0m0p0",
			Type:           "CabinetPDU",
			RedfishType:    "PowerDistribution",
			RedfishSubtype: "RackPDU",
			UUID:           "32354641-4135-4332-4a35-313735303734",
			OdataID:        "/redfish/v1/PowerEquipment/RackPDUs/1",
			RfEndpointID:   "x0m0",
		},
		RfEndpointFQDN:        "127.0.0.1",
		URL:                   "127.0.0.1/redfish/v1/PowerEquipment/RackPDUs/1",
		ComponentEndpointType: "ComponentEndpointPowerDistribution",
		RedfishPDUInfo:        &rf.ComponentPDUInfo{Name: "1"}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x0m0p0v1",
			Type:           "CabinetPDUPowerConnector",
			RedfishType:    "Outlet",
			RedfishSubtype: "NEMA_5_20R",
			OdataID:        "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1",
			RfEndpointID:   "x0m0",
		},
		RfEndpointFQDN:        "127.0.0.1",
		URL:                   "127.0.0.1/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1",
		ComponentEndpointType: "ComponentEndpointManager",
		RedfishOutletInfo: &rf.ComponentOutletInfo{
			Name: "A1",
			Actions: &rf.OutletActions{
				PowerControl: &rf.ActionPowerControl{
					Target: "redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1/Outlet.PowerControl"},
				ResetStatistics: &rf.ActionResetStatistics{
					Target: "/redfish/v1/PowerEquipment/RackPDUs/1/Outlets/A1/Outlet.ResetStatistics"}}}},
}
