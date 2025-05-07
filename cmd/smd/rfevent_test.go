// MIT License
//
// (C) Copyright [2019-2025] Hewlett Packard Enterprise Development LP
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
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	base "github.com/Cray-HPE/hms-base/v2"
	compcreds "github.com/Cray-HPE/hms-compcredentials"
	sstorage "github.com/Cray-HPE/hms-securestorage"
	"github.com/OpenCHAMI/smd/v2/internal/hmsds"
	rf "github.com/OpenCHAMI/smd/v2/pkg/redfish"
	st "github.com/OpenCHAMI/smd/v2/pkg/sharedtest"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"
)

var testFQDN = "[fd40:aaaa:bbbb:1000:240:a6ff:ee82:d7c2]"

//////////////////////////////////////////////////////////////////////////////
//                         Utils
//////////////////////////////////////////////////////////////////////////////

// Takes arguments and gets a key that is used to lookup the function
// response.
func GetCompEpIDsGenGetID(f_opts ...hmsds.CompEPFiltFunc) string {
	id := ""
	f := new(hmsds.CompEPFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	if len(f.ID) > 0 {
		id = f.ID[0]
	}
	return id
}

// Generates a mapping function that returns a slice containing one xname
// that is present is the list of component endpoints used to create the
// mapping.   Basically, if an id is in ceps, it is returned as a one entry
// slice. If it is not, the function will return ErrSmMsgBadID.
// This simulates things always being found in the database
func GetCompEpIDsGenReturnIDs(ceps []*sm.ComponentEndpoint) func(string) ([]string, error) {
	returnVals := make(map[string][]string)
	for _, cep := range ceps {
		returnVals[cep.ID] = []string{cep.ID}
	}
	return func(id string) ([]string, error) {
		ids, ok := returnVals[id]
		if !ok {
			return []string{}, ErrSmMsgBadID
		}
		return ids, nil
	}
}

//////////////////////////////////////////////////////////////////////////////
//                         Tests
//////////////////////////////////////////////////////////////////////////////

func TestDoHandleRFEvent(t *testing.T) {
	ssMockLData := []sstorage.MockLookup{
		{
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{
					Xname:    "",
					URL:      "",
					Username: "root",
					Password: "********",
				},
				Err: nil,
			},
		}, {
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{
					Xname:    "",
					URL:      "",
					Username: "root",
					Password: "********",
				},
				Err: nil,
			},
		},
	}
	// These all reference sharedtest.SampleCompEndpoints and those that
	// not not present will not generate a change whether the event is
	// otherwise ok or not.
	tests := []struct {
		event         string
		mainErr       error    // Error from main call.
		expectedIds   []string // Multi-arg calls.
		expectedState string
		expectedFlag  string
		expectedForce bool
		hmsdsRespErr  error // Error from Update call
	}{{
		//
		// Mountain HW, using ResourceEvent schema and PowerStateChange
		//
		st.GenEvent(
			st.EventCrayOnOKChassis,
			st.EpID("x1c4b0:telemetry")),
		nil,
		[]string{"x1c4"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKChassis,
			st.EpID("wrongorder:x1c4b0:context:morecontext")), // should find.
		nil,
		[]string{"x1c4"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKChassis,
			st.EpID(":x1c4b0:")), // should find
		nil,
		[]string{"x1c4"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKChassis,
			st.EpID("")), // No context
		nil,
		[]string{},
		"",
		"",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKChassis,
			st.EpID("CrayTelemetry:")), // No xname in context
		nil,
		[]string{},
		"",
		"",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKChassis,
			st.EpID("x123c4b0")), // Bad ID
		nil,
		[]string{},
		"",
		"",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOffOKChassis,
			st.EpID("x1c4b0")),
		nil,
		[]string{"x1c4"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("Blade4")),
		nil,
		[]string{"x1c4s4", "x1c4s4b0", "x1c4s4b1"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOffOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("Blade4")),
		nil,
		[]string{"x1c4s4", "x1c4s4b0", "x1c4s4b1", "x1c4s4b0n0", "x1c4s4b1n0", "x1c4s4b0n1", "x1c4s4b1n1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("Perif7")),
		nil,
		[]string{"x1c4r7"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOffOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("Perif1")),
		nil,
		[]string{"x1c4r1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOffOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("Perif3")),
		nil,
		[]string{"x1c4r3", "x1c4r3b0", "x1c4r3e0"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOnOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("Perif3")),
		nil,
		[]string{"x1c4r3", "x1c4r3b0", "x1c4r3e0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOffOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("Blade6")),
		nil,
		[]string{"x1c4s6"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventCrayOffOKSlotX,
			st.EpID("x1c4b0"),
			st.RfId("BladeBadNoExist")),
		nil,
		[]string{},
		"",
		"",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b0"),
			st.RfId("Node0"),
			st.Path("Systems"),
			st.PowState("On")),
		nil,
		[]string{"x1c4s4b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b0"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("Off")),
		nil,
		[]string{"x1c4s4b0n1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b1"),
			st.RfId("Node0"),
			st.Path("Systems"),
			st.PowState("Off")),
		nil,
		[]string{"x1c4s4b1n0"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b1"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("Off"),
			st.MsgId("asdfasdfasdf.1.0.ResourcePowerStateChanged")),
		nil,
		[]string{}, // Bad registry
		"",
		"",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b1"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("Off"),
			st.MsgId("CrayTelemetry.ResourcePowerStateChanged")),
		nil,
		[]string{}, // Filtered Registry
		"",
		"",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b1"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("Off"),
			st.MsgId("")),
		nil,
		[]string{}, // Bad registry
		"",
		"",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b0"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("Off"),
			st.MsgId("ResourcePowerStateChanged")), // No prefix
		nil,
		[]string{"x1c4s4b0n1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b1"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("On"),
			st.MsgId("ResourceEvent.2.2.ResourcePowerStateChanged")), // New version
		nil,
		[]string{"x1c4s4b1n1"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b1"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("On"),
			st.Severity("Informational"),
			st.MsgId("resourceEvent.1.0.resourcepowerstatechanged")), // case insensitive
		nil,
		[]string{"x1c4s4b1n1"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventResource(
			st.EventCrayXXPathX,
			st.EpID("x1c4s4b1"),
			st.RfId("Node1"),
			st.Path("Systems"),
			st.PowState("On"),
			st.MsgId("CrayAlerts.1.0.ResourcePowerStateChanged")), // CrayAlerts registery
		nil,
		[]string{"x1c4s4b1n1"},
		"On",
		"OK",
		false,
		nil,
	}, {
		//
		//  Intel Firmware - Alert Registry
		//
		st.GenEventIntel(
			st.EventIntelSystemOnOK,
			st.EpID("x99c0s15b0"),
			st.RfId("QSBP75002224")),
		nil,
		[]string{"x99c0s15b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventIntel(
			st.EventIntelSystemOffOK,
			st.EpID("x99c0s13b0"),
			st.RfId("QSBP80903751")),
		nil,
		[]string{"x99c0s13b0n0"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEventIntel(
			st.EventIntelSystemOnOK,
			st.EpID("x99c0s15b0"),
			st.RfId("QSBP75002224"),
			st.Registry("")), // No prefix, just "SystemPowerOn"
		nil,
		[]string{"x99c0s15b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventIntel(
			st.EventIntelSystemOnOK,
			st.EpID("x99c0s13b0"),
			st.RfId("QSBP80903751"),
			st.Registry("alert.1.0.0.")), // case insensitive
		nil,
		[]string{"x99c0s13b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventIntel(
			st.EventIntelSystemOffOK,
			st.EpID("x99c0s13b0"),
			st.RfId("QSBP80903751"),
			st.Registry("alert.")), // No version
		nil,
		[]string{"x99c0s13b0n0"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEventIntel(
			st.EventIntelSystemOffOK,
			st.EpID("x99c0s13b0"),
			st.RfId("QSBP80903751"),
			st.Registry("Messages.")), // Unexpected registry
		nil,
		[]string{"x99c0s13b0n0"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		//
		//  Gigabyte Firmware - Alert Registry
		// 30
		st.GenEventGigabyte(
			st.EventGigabyteSystemOK,
			st.EpID("x0c0s11b0"),
			st.RfId("SelfOn")), // Power On
		nil,
		[]string{"x0c0s11b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventGigabyte(
			st.EventGigabyteSystemOK,
			st.EpID("x0c0s11b0"),
			st.RfId("SelfOff")), // Power Off
		nil,
		[]string{"x0c0s11b0n1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEventGigabyte(
			st.EventGigabyteSystemOK,
			st.EpID("x0c0s11b0"),
			st.RfId("SelfOn"),
			st.Registry("")), // No prefix, just "SystemPowerOn"
		nil,
		[]string{"x0c0s11b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventGigabyte(
			st.EventGigabyteSystemOK,
			st.EpID("x0c0s11b0"),
			st.RfId("SelfOn"),
			st.Registry("alert.1.0.0.")), // case insensitive
		nil,
		[]string{"x0c0s11b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventGigabyte(
			st.EventGigabyteSystemOK,
			st.EpID("x0c0s11b0"),
			st.RfId("SelfOff"),
			st.Registry("alert.")), // No version
		nil,
		[]string{"x0c0s11b0n1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEventGigabyte(
			st.EventGigabyteSystemOK,
			st.EpID("x0c0s11b0"),
			st.RfId("SelfOff"),
			st.Registry("Messages.")), // Unexpected registry
		nil,
		[]string{"x0c0s11b0n1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenRtsEventResource(
			st.EventRtsXXPathX,
			st.EpID("x0m0"),
			st.RfId("A1"),
			st.PathParent("1"),
			st.PowState("On"),
			st.Severity("Informational")),
		nil,
		[]string{"x0m0p0v1"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenRtsEventResource(
			st.EventRtsXXPathX,
			st.EpID("x0m0"),
			st.RfId("A1"),
			st.PathParent("1"),
			st.PowState("Off")),
		nil,
		[]string{"x0m0p0v1"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		//
		//  HPE - iLO Registry
		//
		st.GenEventHPEiLO(
			st.EventHPEiLOServerPoweredOn,
			st.EpID("x3000c0s1b0"),
			st.RfId("1")),
		nil,
		[]string{"x3000c0s1b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEventHPEiLO(
			st.EventHPEiLOServerPoweredOff,
			st.EpID("x3000c0s2b0"),
			st.RfId("1")),
		nil,
		[]string{"x3000c0s2b0n0"},
		"Off",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventFoxconnServerPoweredOn,
			st.EpID("x3000c0s2b0")),
		nil,
		[]string{"x3000c0s2b0n0"},
		"On",
		"OK",
		false,
		nil,
	}, {
		st.GenEvent(
			st.EventFoxconnServerPoweredOff,
			st.EpID("x3000c0s2b0")),
		nil,
		[]string{"x3000c0s2b0n0"},
		"Off",
		"OK",
		false,
		nil,
	}}

	// Start a local HTTP server
	server := httptest.NewTLSServer(http.HandlerFunc(GigabyteHandler))
	// Close the server when test finishes
	defer server.Close()
	u, _ := url.Parse(server.URL)
	testFQDN = u.Host

	// Start a local HTTP server
	server2 := httptest.NewTLSServer(http.HandlerFunc(DummyHandler))
	// Close the server when test finishes
	defer server2.Close()
	u, _ = url.Parse(server2.URL)
	test2FQDN := u.Host

	// Use the mock secure storage
	ss, adapter := sstorage.NewMockAdapter()
	ccs := compcreds.NewCompCredStore("secret/hms-cred", ss)
	s.ss = ss
	s.ccs = ccs
	s.writeVault = true
	s.readVault = true

	results.GetCompEndpointsAll.Return.entries = st.SampleCompEndpoints
	results.GetCompEndpointsAll.Return.err = nil

	// Add our customized gigabyte nodes. They need to be customized because
	// their URLs need to point to our test server.
	for _, cep := range GigabyteCompEPs {
		cep.RfEndpointFQDN = testFQDN
		cep.URL = testFQDN + cep.URL
	}
	for _, cep := range results.GetCompEndpointsAll.Return.entries {
		cep.RfEndpointFQDN = test2FQDN
		cep.URL = test2FQDN + cep.URL
	}
	for _, cep := range HPEiLOCompEPs {
		cep.RfEndpointFQDN = test2FQDN
		cep.URL = test2FQDN + cep.URL
	}

	results.GetCompEndpointsAll.Return.entries = append(results.GetCompEndpointsAll.Return.entries, GigabyteCompEPs...)
	results.GetCompEndpointsAll.Return.entries = append(results.GetCompEndpointsAll.Return.entries, HPEiLOCompEPs...)

	// We mock up the behavior of this function.  It always finds things
	// that are in the sample endpoints, and never finds things that are not.
	results.GetCompEndpointIDs.Funcs.getID = GetCompEpIDsGenGetID
	results.GetCompEndpointIDs.Funcs.returnIDs = GetCompEpIDsGenReturnIDs(results.GetCompEndpointsAll.Return.entries)

	for i, test := range tests {
		adapter.LookupNum = 0
		adapter.LookupData = ssMockLData
		results.UpdateCompStates.Input.ids = []string{}
		results.UpdateCompStates.Input.state = ""
		results.UpdateCompStates.Input.flag = ""
		results.UpdateCompStates.Input.force = false
		results.UpdateCompStates.Input.pi = nil

		results.UpdateCompStates.Return.affectedIds = test.expectedIds
		results.UpdateCompStates.Return.err = test.hmsdsRespErr

		t.Logf("Test %d Running... event is '%s'", i, test.event)
		// Run test - Check return value
		err := s.doHandleRFEvent(test.event)
		if err != nil {
			if test.mainErr == nil {
				t.Errorf("Test %d FAIL: Unexpected error: '%s'", i, err)
			} else if test.mainErr != err {
				t.Errorf("Test %d FAIL: Expected err '%s'; Received: '%s'",
					i, test.mainErr, err)
			}
		} else if test.mainErr != nil {
			t.Errorf("Test %d FAIL: No error, expected '%s'", i, test.mainErr)
		}
		// Now check to see if the update function (one or the other)
		// got the expected IDs and state/flag values

		if !compareIDs(test.expectedIds, results.UpdateCompStates.Input.ids) {
			t.Errorf("Test %d FAIL: Expected ids '%s'; Received ids '%s'",
				i, test.expectedIds, results.UpdateCompStates.Input.ids)
		}
		if test.expectedState != results.UpdateCompStates.Input.state {
			t.Errorf("Test %d FAIL: Expected state '%s'; Received state '%s'",
				i, test.expectedState, results.UpdateCompStates.Input.state)
		}
		if test.expectedFlag != results.UpdateCompStates.Input.flag {
			t.Errorf("Test %d FAIL: Expected flag '%s'; Received flag '%s'",
				i, test.expectedFlag, results.UpdateCompStates.Input.flag)
		}
		if test.expectedFlag != results.UpdateCompStates.Input.flag {
			t.Errorf("Test %d FAIL: Expected flag '%s'; Received flag '%s'",
				i, test.expectedFlag, results.UpdateCompStates.Input.flag)
		}
		if test.expectedForce != results.UpdateCompStates.Input.force {
			t.Errorf("Test %d FAIL: Expected force '%v'; Received force '%v'",
				i, test.expectedForce, results.UpdateCompStates.Input.force)
		}
		if len(test.expectedIds) > 0 && results.UpdateCompStates.Input.pi == nil {
			t.Errorf("Test %d FAIL: Expected non-nil pi; Received nil", i)
		}
	}
}

//////////////////////////////////////////////////////////////////////////////
//                         Dummy Mock Server
//////////////////////////////////////////////////////////////////////////////

func DummyHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(404)
	base.DrainAndCloseRequestBody(req)
	return
}

//////////////////////////////////////////////////////////////////////////////
//                         Gigabyte - Mock Server
//////////////////////////////////////////////////////////////////////////////

func GigabyteHandler(w http.ResponseWriter, req *http.Request) {
	defer base.DrainAndCloseRequestBody(req)

	// Test request parameters
	switch req.URL.RequestURI() {

	case testPathGigabyteSystems:
		w.WriteHeader(200)
		w.Write(json.RawMessage(testPayloadGigabyteSystems))
		return
	case testPathGigabyteSystems_SelfOn:
		w.WriteHeader(200)
		w.Write(json.RawMessage(testPayloadGigabyteSystems_SelfOn))
		return
	case testPathGigabyteSystems_SelfOn_Processors:
		w.WriteHeader(200)
		w.Write(json.RawMessage(testPayloadGigabyteSystems_SelfOn_Processors))
		return
	case testPathGigabyteSystems_SelfOn_Processors_1:
		w.WriteHeader(200)
		w.Write(json.RawMessage(testPayloadGigabyteSystems_SelfOn_Processors_1))
		return
	case testPathGigabyteSystems_SelfOn_Memory:
		w.WriteHeader(200)
		w.Write(json.RawMessage(testPayloadGigabyteSystems_SelfOn_Memory))
		return
	case testPathGigabyteSystems_SelfOn_Memory_1:
		w.WriteHeader(200)
		w.Write(json.RawMessage(testPayloadGigabyteSystems_SelfOn_Memory_1))
		return
	case testPathGigabyteSystems_SelfOff:
		w.WriteHeader(200)
		w.Write(json.RawMessage(testPayloadGigabyteSystems_SelfOff))
		return
	default:
		w.WriteHeader(404)
		return
	}
}

//////////////////////////////////////////////////////////////////////////////
//                         Gigabyte - Payloads
//////////////////////////////////////////////////////////////////////////////

var GigabyteCompEPs = []*sm.ComponentEndpoint{
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x0c0s11b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			UUID:           "cd210000-3b17-11ea-8000-e0d55e659162",
			OdataID:        "/redfish/v1/Systems/SelfOn",
			RfEndpointID:   "x0c0s11b0",
		},
		RfEndpointFQDN:        "",
		URL:                   "/redfish/v1/Systems/SelfOn",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "System",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"GracefulShutdown",
					"ForceRestart"},
				Target: "/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				{RedfishId: "1",
					Oid:         "/redfish/v1/Systems/Self/EthernetInterfaces/1",
					Description: "Ethernet Interface Lan1",
					MACAddress:  "e0:d5:5e:65:91:62"},
				{RedfishId: "2",
					Oid:         "/redfish/v1/Systems/Self/EthernetInterfaces/2",
					Description: "Ethernet Interface Lan2",
					MACAddress:  "e0:d5:5e:65:91:63"}}}},
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x0c0s11b0n1",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			UUID:           "cd210000-3b17-11ea-8000-e0d55e659162",
			OdataID:        "/redfish/v1/Systems/SelfOff",
			RfEndpointID:   "x0c0s11b0",
		},
		RfEndpointFQDN:        "",
		URL:                   "/redfish/v1/Systems/SelfOff",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "System",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{"On",
					"ForceOff",
					"GracefulShutdown",
					"ForceRestart"},
				Target: "/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:         "/redfish/v1/Systems/Self/EthernetInterfaces/1",
					Description: "Ethernet Interface Lan1",
					MACAddress:  "e0:d5:5e:65:91:62"},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:         "/redfish/v1/Systems/Self/EthernetInterfaces/2",
					Description: "Ethernet Interface Lan2",
					MACAddress:  "e0:d5:5e:65:91:63"}}}},
}

const testPathGigabyteSystems = "/redfish/v1/Systems"

const testPayloadGigabyteSystems = `
{
  "@Redfish.CollectionCapabilities": {
    "@odata.type": "#CollectionCapabilities.v1_1_0.CollectionCapabilities",
    "Capabilities": [
      {
        "CapabilitiesObject": {
          "@odata.id": "/redfish/v1/Systems/Capabilities"
        },
        "Links": {
          "TargetCollection": {
            "@odata.id": "/redfish/v1/Systems"
          }
        },
        "UseCase": "ComputerSystemComposition"
      }
    ]
  },
  "@odata.context": "/redfish/v1/$metadata#ComputerSystemCollection.ComputerSystemCollection",
  "@odata.etag": "W/\"1574110568\"",
  "@odata.id": "/redfish/v1/Systems",
  "@odata.type": "#ComputerSystemCollection.ComputerSystemCollection",
  "Description": "Collection of Computer Systems",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Systems/SelfOn"
    }
  ],
  "Members@odata.count": 1,
  "Name": "Systems Collection"
}
`

const testPathGigabyteSystems_SelfOn = "/redfish/v1/Systems/SelfOn"

const testPayloadGigabyteSystems_SelfOn = `
{
    "@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem(*)",
    "@odata.etag": "W/\"1566514653\"",
    "@odata.id": "/redfish/v1/Systems/SelfOn",
    "@odata.type": "#ComputerSystem.v1_3_0.ComputerSystem",
    "Actions": {
        "#ComputerSystem.Reset": {
            "ResetType@Redfish.AllowableValues": [
                "On",
                "ForceOff",
                "ForceRestart",
                "GracefulShutdown"
            ],
            "target": "/redfish/v1/Systems/SelfOn/Actions/ComputerSystem.Reset"
        }
    },
    "AssetTag": "Free form asset tag",
    "BiosVersion": "R04",
    "Boot": {
        "BootSourceOverrideEnabled": "Disabled",
        "BootSourceOverrideEnabled@Redfish.AllowableValues": [
            "Disabled",
            "Once",
            "Continuous"
        ],
        "BootSourceOverrideMode": "Legacy",
        "BootSourceOverrideMode@Redfish.AllowableValues": [
            "Legacy",
            "UEFI"
        ],
        "BootSourceOverrideTarget": "None",
        "BootSourceOverrideTarget@Redfish.AllowableValues": [
            "None",
            "Pxe",
            "Cd",
            "Usb",
            "Hdd",
            "BiosSetup"
        ]
    },
    "Description": "System SelfOn",
    "Id": "SelfOn",
    "IndicatorLED": "Off",
    "Links": {
        "Chassis": [
            {
                "@odata.id": "/redfish/v1/Chassis/Self"
            }
        ],
        "ManagedBy": [
            {
                "@odata.id": "/redfish/v1/Managers/Self"
            }
        ]
    },
    "LogServices": {
        "@odata.id": "/redfish/v1/Systems/SelfOn/LogServices"
    },
    "Manufacturer": "GIGABYTE",
    "Memory": {
        "@odata.id": "/redfish/v1/Systems/SelfOn/Memory"
    },
    "MemorySummary": {
        "TotalSystemMemoryGiB": 244
    },
    "Model": "R272-Z30-00",
    "Name": "System",
    "NetworkInterfaces": {
        "@odata.id": "/redfish/v1/Systems/SelfOn/NetworkInterfaces"
    },
    "PartNumber": "0100",
    "PowerState": "On",
    "ProcessorSummary": {
        "Count": 1,
        "Model": "AMD Eng Sample: 100-000000053-04_32/20_N       "
    },
    "Processors": {
        "@odata.id": "/redfish/v1/Systems/SelfOn/Processors"
    },
    "SKU": "01234567890123456789AB",
    "SecureBoot": {
        "@odata.id": "/redfish/v1/Systems/SelfOn/SecureBoot"
    },
    "SerialNumber": "01234567890123456789AB",
    "SimpleStorage": {
        "@odata.id": "/redfish/v1/Systems/SelfOn/SimpleStorage"
    },
    "Status": {
        "Health": "OK",
        "HealthRollup": "OK",
        "State": "Enabled"
    },
    "Storage": {
        "@odata.id": "/redfish/v1/Systems/SelfOn/Storage"
    },
    "SystemType": "Physical",
    "UUID": "cd210000-3b17-11ea-8000-e0d55e659162"
}
`

const testPathGigabyteSystems_SelfOn_Processors = "/redfish/v1/Systems/SelfOn/Processors"

const testPayloadGigabyteSystems_SelfOn_Processors = `
{
  "@odata.context": "/redfish/v1/$metadata#ProcessorCollection.ProcessorCollection",
  "@odata.etag": "W/\"1574289802\"",
  "@odata.id": "/redfish/v1/Systems/SelfOn/Processors",
  "@odata.type": "#ProcessorCollection.ProcessorCollection",
  "Description": "Collection of processors",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Systems/SelfOn/Processors/1"
    }
  ],
  "Members@odata.count": 1,
  "Name": "Processors Collection"
}
`

const testPathGigabyteSystems_SelfOn_Processors_1 = "/redfish/v1/Systems/SelfOn/Processors/1"

const testPayloadGigabyteSystems_SelfOn_Processors_1 = `
{
  "@odata.context": "/redfish/v1/$metadata#Processor.Processor",
  "@odata.etag": "W/\"1574289802\"",
  "@odata.id": "/redfish/v1/Systems/SelfOn/Processors/1",
  "@odata.type": "#Processor.v1_3_1.Processor",
  "Description": "Processor Instance 1",
  "Id": "1",
  "InstructionSet": "x86-64",
  "Links": {
    "Chassis": {
      "@odata.id": "/redfish/v1/Chassis/SelfOn"
    }
  },
  "Manufacturer": "Advanced Micro Devices, Inc.",
  "MaxSpeedMHz": 3400,
  "Model": "AMD EPYC 7742 64-Core Processor                ",
  "Name": "Processor 1",
  "ProcessorArchitecture": "x86",
  "ProcessorId": {
    "EffectiveFamily": "AMD Zen Processor Family",
    "EffectiveModel": "0x31",
    "IdentificationRegisters": "178bfbff00830f10",
    "Step": "0x0",
    "VendorId": "AuthenticAMD"
  },
  "ProcessorType": "CPU",
  "Socket": "P0",
  "Status": {
    "Health": "OK",
    "State": "Enabled"
  },
  "TotalCores": 64,
  "TotalThreads": 128
}
`

const testPathGigabyteSystems_SelfOn_Memory = "/redfish/v1/Systems/SelfOn/Memory"

const testPayloadGigabyteSystems_SelfOn_Memory = `
{
  "@odata.context": "/redfish/v1/$metadata#MemoryCollection.MemoryCollection",
  "@odata.etag": "W/\"1574289802\"",
  "@odata.id": "/redfish/v1/Systems/SelfOn/Memory",
  "@odata.type": "#MemoryCollection.MemoryCollection",
  "Description": "Collection of Memories for this system",
  "Members": [
    {
      "@odata.id": "/redfish/v1/Systems/SelfOn/Memory/1"
    }
  ],
  "Members@odata.count": 1,
  "Name": "Memory Collection"
}
`

const testPathGigabyteSystems_SelfOn_Memory_1 = "/redfish/v1/Systems/SelfOn/Memory/1"

const testPayloadGigabyteSystems_SelfOn_Memory_1 = `
{
  "@odata.context": "/redfish/v1/$metadata#Memory.Memory",
  "@odata.etag": "W/\"1574289802\"",
  "@odata.id": "/redfish/v1/Systems/SelfOn/Memory/1",
  "@odata.type": "#Memory.v1_6_0.Memory",
  "Actions": {
    "Oem": {
      "#AmiBios.ChangeState": {
        "State@Redfish.AllowableValues": [
          "Enabled",
          "Disabled"
        ],
        "target": "/redfish/v1/Systems/SelfOn/Memory/1/Actions/AmiBios.ChangeState"
      }
    }
  },
  "AllowedSpeedsMHz": [
    0
  ],
  "BusWidthBits": 0,
  "CapacityMiB": 0,
  "DataWidthBits": 0,
  "Description": "Memory Instance 1",
  "DeviceLocator": "DIMM_P0_A0",
  "ErrorCorrection": "MultiBitECC",
  "Id": "1",
  "Links": {
    "Chassis": {
      "@odata.id": "/redfish/v1/Chassis/SelfOn"
    }
  },
  "Manufacturer": "Unknown",
  "MemoryType": "DRAM",
  "Name": "Memory 1",
  "OperatingSpeedMhz": 0,
  "PartNumber": "Unknown",
  "SerialNumber": "Unknown",
  "Status": {
    "State": "Absent"
  }
}
`

const testPathGigabyteSystems_SelfOff = "/redfish/v1/Systems/SelfOff"

const testPayloadGigabyteSystems_SelfOff = `
{
    "@odata.context": "/redfish/v1/$metadata#ComputerSystem.ComputerSystem(*)",
    "@odata.etag": "W/\"1566514653\"",
    "@odata.id": "/redfish/v1/Systems/Self",
    "@odata.type": "#ComputerSystem.v1_3_0.ComputerSystem",
    "Actions": {
        "#ComputerSystem.Reset": {
            "ResetType@Redfish.AllowableValues": [
                "On",
                "ForceOff",
                "ForceRestart",
                "GracefulShutdown"
            ],
            "target": "/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"
        }
    },
    "AssetTag": "Free form asset tag",
    "BiosVersion": "R04",
    "Boot": {
        "BootSourceOverrideEnabled": "Disabled",
        "BootSourceOverrideEnabled@Redfish.AllowableValues": [
            "Disabled",
            "Once",
            "Continuous"
        ],
        "BootSourceOverrideMode": "Legacy",
        "BootSourceOverrideMode@Redfish.AllowableValues": [
            "Legacy",
            "UEFI"
        ],
        "BootSourceOverrideTarget": "None",
        "BootSourceOverrideTarget@Redfish.AllowableValues": [
            "None",
            "Pxe",
            "Cd",
            "Usb",
            "Hdd",
            "BiosSetup"
        ]
    },
    "Description": "System Self",
    "EthernetInterfaces": {
        "@odata.id": "/redfish/v1/Systems/Self/EthernetInterfaces"
    },
    "Id": "Self",
    "IndicatorLED": "Off",
    "Links": {
        "Chassis": [
            {
                "@odata.id": "/redfish/v1/Chassis/Self"
            }
        ],
        "ManagedBy": [
            {
                "@odata.id": "/redfish/v1/Managers/Self"
            }
        ]
    },
    "LogServices": {
        "@odata.id": "/redfish/v1/Systems/Self/LogServices"
    },
    "Manufacturer": "GIGABYTE",
    "Memory": {
        "@odata.id": "/redfish/v1/Systems/Self/Memory"
    },
    "MemorySummary": {
        "TotalSystemMemoryGiB": 244
    },
    "Model": "R272-Z30-00",
    "Name": "System",
    "NetworkInterfaces": {
        "@odata.id": "/redfish/v1/Systems/Self/NetworkInterfaces"
    },
    "PartNumber": "0100",
    "PowerState": "Off",
    "ProcessorSummary": {
        "Count": 1,
        "Model": "AMD Eng Sample: 100-000000053-04_32/20_N       "
    },
    "Processors": {
        "@odata.id": "/redfish/v1/Systems/Self/Processors"
    },
    "SKU": "01234567890123456789AB",
    "SecureBoot": {
        "@odata.id": "/redfish/v1/Systems/Self/SecureBoot"
    },
    "SerialNumber": "01234567890123456789AB",
    "SimpleStorage": {
        "@odata.id": "/redfish/v1/Systems/Self/SimpleStorage"
    },
    "Status": {
        "Health": "OK",
        "HealthRollup": "OK",
        "State": "Enabled"
    },
    "Storage": {
        "@odata.id": "/redfish/v1/Systems/Self/Storage"
    },
    "SystemType": "Physical",
    "UUID": "cd210000-3b17-11ea-8000-e0d55e659162"
}
`

//////////////////////////////////////////////////////////////////////////////
//                         HPE iLO - Payloads
//////////////////////////////////////////////////////////////////////////////

var HPEiLOCompEPs = []*sm.ComponentEndpoint{
	{
		ComponentDescription: rf.ComponentDescription{
			ID:             "x3000c0s1b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			UUID:           "34364C58-6435-4732-3037-505030303235",
			OdataID:        "/redfish/v1/Systems/1",
			RfEndpointID:   "x3000c0s1b0",
		},
		RfEndpointFQDN:        "",
		URL:                   "/redfish/v1/Systems/1",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "Computer System",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{
					"On",
					"ForceOff",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"PushPowerButton",
					"GracefulRestart",
				},
				Target: "/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"},
			},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:        "/redfish/v1/Systems/1/EthernetInterfaces/1",
					MACAddress: "68:05:ca:bb:c1:82",
				},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:        "/redfish/v1/Systems/1/EthernetInterfaces/2",
					MACAddress: "68:05:ca:bb:c1:83",
				},
			},
		},
	}, {
		ComponentDescription: rf.ComponentDescription{
			ID:             "x3000c0s2b0n0",
			Type:           "Node",
			RedfishType:    "ComputerSystem",
			RedfishSubtype: "Physical",
			UUID:           "34364C58-6435-4732-3037-505030303235",
			OdataID:        "/redfish/v1/Systems/1",
			RfEndpointID:   "x3000c0s2b0",
		},
		RfEndpointFQDN:        "",
		URL:                   "/redfish/v1/Systems/1",
		ComponentEndpointType: "ComponentEndpointComputerSystem",
		RedfishSystemInfo: &rf.ComponentSystemInfo{
			Name: "Computer System",
			Actions: &rf.ComputerSystemActions{ComputerSystemReset: rf.ActionReset{
				AllowableValues: []string{
					"On",
					"ForceOff",
					"GracefulShutdown",
					"ForceRestart",
					"Nmi",
					"PushPowerButton",
					"GracefulRestart",
				},
				Target: "/redfish/v1/Systems/1/Actions/ComputerSystem.Reset"},
			},
			EthNICInfo: []*rf.EthernetNICInfo{
				&rf.EthernetNICInfo{RedfishId: "1",
					Oid:        "/redfish/v1/Systems/1/EthernetInterfaces/1",
					MACAddress: "68:05:ca:bb:c1:82",
				},
				&rf.EthernetNICInfo{RedfishId: "2",
					Oid:        "/redfish/v1/Systems/1/EthernetInterfaces/2",
					MACAddress: "68:05:ca:bb:c1:83",
				},
			},
		},
	},
}
