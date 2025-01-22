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
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	base "github.com/Cray-HPE/hms-base"
	compcreds "github.com/Cray-HPE/hms-compcredentials"
	sstorage "github.com/Cray-HPE/hms-securestorage"
	"github.com/Cray-HPE/hms-smd/v2/internal/hmsds"
	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
	stest "github.com/Cray-HPE/hms-smd/v2/pkg/sharedtest"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"

	"github.com/gorilla/mux"
)

var s *SmD
var results *TestResults
var router *mux.Router

var ffStringMap = map[hmsds.FieldFilter]string{
	hmsds.FLTR_DEFAULT:   "FLTR_DEFAULT",
	hmsds.FLTR_STATEONLY: "FLTR_STATEONLY",
	hmsds.FLTR_FLAGONLY:  "FLTR_FLAGONLY",
	hmsds.FLTR_ROLEONLY:  "FLTR_ROLEONLY",
	hmsds.FLTR_NIDONLY:   "FLTR_NIDONLY",
	hmsds.FLTR_MAX:       "FLTR_MAX",
}

///////////////////////////////////////////////////////////////////////////////
// Test Helper Functions
///////////////////////////////////////////////////////////////////////////////

// Compare one hmsds.ComponentFilter to another
func compareFilter(fltr1 hmsds.ComponentFilter, fltr2 hmsds.ComponentFilter) bool {
	if len(fltr1.NID) != len(fltr2.NID) {
		return false
	}
	if len(fltr1.NIDStart) != len(fltr2.NIDStart) {
		return false
	}
	if len(fltr1.NIDEnd) != len(fltr2.NIDEnd) {
		return false
	}
	if len(fltr1.Type) != len(fltr2.Type) {
		return false
	}
	if len(fltr1.State) != len(fltr2.State) {
		return false
	}
	if len(fltr1.Flag) != len(fltr2.Flag) {
		return false
	}
	if len(fltr1.Role) != len(fltr2.Role) {
		return false
	}
	if len(fltr1.SubRole) != len(fltr2.SubRole) {
		return false
	}
	if len(fltr1.Subtype) != len(fltr2.Subtype) {
		return false
	}
	if len(fltr1.Arch) != len(fltr2.Arch) {
		return false
	}
	if len(fltr1.Class) != len(fltr2.Class) {
		return false
	}
	for i := 0; i < len(fltr1.NID); i++ {
		found := false
		for j := 0; j < len(fltr2.NID); j++ {
			if fltr1.NID[i] == fltr2.NID[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.NIDStart); i++ {
		found := false
		for j := 0; j < len(fltr2.NIDStart); j++ {
			if fltr1.NIDStart[i] == fltr2.NIDStart[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.NIDEnd); i++ {
		found := false
		for j := 0; j < len(fltr2.NIDEnd); j++ {
			if fltr1.NIDEnd[i] == fltr2.NIDEnd[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.Type); i++ {
		found := false
		for j := 0; j < len(fltr2.Type); j++ {
			if fltr1.Type[i] == fltr2.Type[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.State); i++ {
		found := false
		for j := 0; j < len(fltr2.State); j++ {
			if fltr1.State[i] == fltr2.State[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.Flag); i++ {
		found := false
		for j := 0; j < len(fltr2.Flag); j++ {
			if fltr1.Flag[i] == fltr2.Flag[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.Role); i++ {
		found := false
		for j := 0; j < len(fltr2.Role); j++ {
			if fltr1.Role[i] == fltr2.Role[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.SubRole); i++ {
		found := false
		for j := 0; j < len(fltr2.SubRole); j++ {
			if fltr1.SubRole[i] == fltr2.SubRole[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.Subtype); i++ {
		found := false
		for j := 0; j < len(fltr2.Subtype); j++ {
			if fltr1.Subtype[i] == fltr2.Subtype[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.Arch); i++ {
		found := false
		for j := 0; j < len(fltr2.Arch); j++ {
			if fltr1.Arch[i] == fltr2.Arch[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	for i := 0; i < len(fltr1.Class); i++ {
		found := false
		for j := 0; j < len(fltr2.Class); j++ {
			if fltr1.Class[i] == fltr2.Class[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Compare arrays of xnames to make sure every name exists in both lists (order doesn't matter)
func compareIDs(ids1 []string, ids2 []string) bool {
	if len(ids1) != len(ids2) {
		return false
	}
	for i := 0; i < len(ids1); i++ {
		found := false
		for j := 0; j < len(ids2); j++ {
			if ids1[i] == ids2[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Compare arrays of components to make sure every component exists in both lists (order doesn't matter)
func compareComps(comps1 *[]base.Component, comps2 *[]base.Component) bool {
	if len(*comps1) != len(*comps2) {
		return false
	}
	for _, comp1 := range *comps1 {
		found := false
		for _, comp2 := range *comps2 {
			if comp1 == comp2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func compareSCNSubs(subs1, subs2 sm.SCNSubscriptionArray) bool {
	if len(subs1.SubscriptionList) != len(subs2.SubscriptionList) {
		return false
	}
	for i, sub1 := range subs1.SubscriptionList {
		sub2 := subs2.SubscriptionList[i]
		if sub1.ID == sub2.ID &&
			sub1.Subscriber == sub2.Subscriber &&
			sub1.Url == sub2.Url &&
			sub1.Enabled == sub2.Enabled &&
			len(sub1.Roles) == len(sub2.Roles) &&
			len(sub1.SubRoles) == len(sub2.SubRoles) &&
			len(sub1.SoftwareStatus) == len(sub2.SoftwareStatus) &&
			len(sub1.States) == len(sub2.States) {
			if sub1.Enabled != nil && *sub1.Enabled != *sub2.Enabled {
				return false
			}
			for j, role1 := range sub1.Roles {
				role2 := sub2.Roles[j]
				if role1 != role2 {
					return false
				}
			}
			for j, subRole1 := range sub1.SubRoles {
				subRole2 := sub2.SubRoles[j]
				if subRole1 != subRole2 {
					return false
				}
			}
			for j, swStatus1 := range sub1.SoftwareStatus {
				swStatus2 := sub2.SoftwareStatus[j]
				if swStatus1 != swStatus2 {
					return false
				}
			}
			for j, state1 := range sub1.States {
				state2 := sub2.States[j]
				if state1 != state2 {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func compareSCNSubMaps(map1, map2 SCNSubMap) bool {
	for i, subMap1 := range map1 {
		for key, urls1 := range subMap1 {
			if len(urls1) != len(map2[i][key]) {
				return false
			}
			for j, url1 := range urls1 {
				if url1.url != map2[i][key][j].url &&
					url1.refCount != map2[i][key][j].refCount {
					return false
				}
			}
		}
	}
	return true
}

func compareGroup(grp1, grp2 *sm.Group) bool {
	if (grp1 == nil) != (grp2 == nil) {
		return false
	} else if grp1 == nil {
		return true
	}
	if grp1.Label != grp2.Label ||
		grp1.Description != grp2.Description ||
		grp1.ExclusiveGroup != grp2.ExclusiveGroup ||
		len(grp1.Tags) != len(grp2.Tags) ||
		len(grp1.Members.IDs) != len(grp2.Members.IDs) {
		return false
	}
	if len(grp1.Tags) > 0 {
		for i, tag := range grp1.Tags {
			if tag != grp2.Tags[i] {
				return false
			}
		}
	}
	if len(grp1.Members.IDs) > 0 {
		for i, id := range grp1.Members.IDs {
			if id != grp2.Members.IDs[i] {
				return false
			}
		}
	}
	return true
}

func compareGroupPatch(grp1, grp2 *sm.GroupPatch) bool {
	if (grp1 == nil) != (grp2 == nil) {
		return false
	} else if grp1 == nil {
		return true
	}
	if (grp1.Description == nil) != (grp2.Description == nil) ||
		(grp1.Tags == nil) != (grp2.Tags == nil) {
		return false
	}
	if grp1.Description != nil && grp1.Description != grp2.Description {
		return false
	}
	if grp1.Tags != nil {
		if len(*grp1.Tags) != len(*grp2.Tags) {
			return false
		}
		for i, tag := range *grp1.Tags {
			if tag != (*grp2.Tags)[i] {
				return false
			}
		}
	}
	return true
}

func comparePartition(part1, part2 *sm.Partition) bool {
	if (part1 == nil) != (part2 == nil) {
		return false
	} else if part1 == nil {
		return true
	}
	if part1.Name != part2.Name ||
		part1.Description != part2.Description ||
		len(part1.Tags) != len(part2.Tags) ||
		len(part1.Members.IDs) != len(part2.Members.IDs) {
		return false
	}
	if len(part1.Tags) > 0 {
		for i, tag := range part1.Tags {
			if tag != part2.Tags[i] {
				return false
			}
		}
	}
	if len(part1.Members.IDs) > 0 {
		for i, id := range part1.Members.IDs {
			if id != part2.Members.IDs[i] {
				return false
			}
		}
	}
	return true
}

func comparePartitionPatch(part1, part2 *sm.PartitionPatch) bool {
	if (part1 == nil) != (part2 == nil) {
		return false
	} else if part1 == nil {
		return true
	}
	if (part1.Description == nil) != (part2.Description == nil) ||
		(part1.Tags == nil) != (part2.Tags == nil) {
		return false
	}
	if part1.Description != nil && part1.Description != part2.Description {
		return false
	}
	if part1.Tags != nil {
		if len(*part1.Tags) != len(*part2.Tags) {
			return false
		}
		for i, tag := range *part1.Tags {
			if tag != (*part2.Tags)[i] {
				return false
			}
		}
	}
	return true
}

// Compare one hmsds.HWInvLocFilter to another
func compareHWInvLocFilter(fltr1 hmsds.HWInvLocFilter, fltr2 hmsds.HWInvLocFilter) bool {
	if len(fltr1.ID) != len(fltr2.ID) ||
		len(fltr1.Type) != len(fltr2.Type) ||
		len(fltr1.Manufacturer) != len(fltr2.Manufacturer) ||
		len(fltr1.PartNumber) != len(fltr2.PartNumber) ||
		len(fltr1.SerialNumber) != len(fltr2.SerialNumber) ||
		len(fltr1.FruId) != len(fltr2.FruId) ||
		len(fltr1.Partition) != len(fltr2.Partition) ||
		fltr1.Children != fltr2.Children ||
		fltr1.Parents != fltr2.Parents {
		return false
	}

	for i, id := range fltr1.ID {
		if id != fltr2.ID[i] {
			return false
		}
	}
	for i, hmsType := range fltr1.Type {
		if hmsType != fltr2.Type[i] {
			return false
		}
	}
	for i, manufacturer := range fltr1.Manufacturer {
		if manufacturer != fltr2.Manufacturer[i] {
			return false
		}
	}
	for i, partNumber := range fltr1.PartNumber {
		if partNumber != fltr2.PartNumber[i] {
			return false
		}
	}
	for i, serialNumber := range fltr1.SerialNumber {
		if serialNumber != fltr2.SerialNumber[i] {
			return false
		}
	}
	for i, fruId := range fltr1.FruId {
		if fruId != fltr2.FruId[i] {
			return false
		}
	}
	for i, partition := range fltr1.Partition {
		if partition != fltr2.Partition[i] {
			return false
		}
	}
	return true
}

// Compare one hmsds.HWInvLocFilter to another
func compareHWInvHistFilter(fltr1 *hmsds.HWInvHistFilter, fltr2 *hmsds.HWInvHistFilter) bool {
	if (fltr1 == nil) != (fltr2 == nil) {
		return false
	}
	if fltr1 == nil {
		return true
	}
	if len(fltr1.ID) != len(fltr2.ID) ||
		len(fltr1.FruId) != len(fltr2.FruId) ||
		len(fltr1.EventType) != len(fltr2.EventType) ||
		fltr1.StartTime != fltr2.StartTime ||
		fltr1.EndTime != fltr2.EndTime {
		return false
	}

	for i, id := range fltr1.ID {
		if id != fltr2.ID[i] {
			return false
		}
	}
	for i, fruid := range fltr1.FruId {
		if fruid != fltr2.FruId[i] {
			return false
		}
	}
	for i, eventType := range fltr1.EventType {
		if eventType != fltr2.EventType[i] {
			return false
		}
	}
	return true
}

func compareCompEthInterfaceFilter(fltr1 hmsds.CompEthInterfaceFilter, fltr2 hmsds.CompEthInterfaceFilter) bool {
	if len(fltr1.ID) != len(fltr2.ID) {
		return false
	}
	if len(fltr1.MACAddr) != len(fltr2.MACAddr) {
		return false
	}
	if len(fltr1.IPAddr) != len(fltr2.IPAddr) {
		return false
	}
	if fltr1.OlderThan != fltr2.OlderThan {
		return false
	}
	if fltr1.NewerThan != fltr2.NewerThan {
		return false
	}
	if len(fltr1.CompID) != len(fltr2.CompID) {
		return false
	}
	if len(fltr1.CompType) != len(fltr2.CompType) {
		return false
	}
	for i, id := range fltr1.ID {
		if id != fltr2.ID[i] {
			return false
		}
	}
	for i, mac := range fltr1.MACAddr {
		if mac != fltr2.MACAddr[i] {
			return false
		}
	}
	for i, ip := range fltr1.IPAddr {
		if ip != fltr2.IPAddr[i] {
			return false
		}
	}
	for i, cid := range fltr1.CompID {
		if cid != fltr2.CompID[i] {
			return false
		}
	}
	for i, t := range fltr1.CompType {
		if t != fltr2.CompType[i] {
			return false
		}
	}
	return true
}

///////////////////////////////////////////////////////////////////////////////
// Pre-Test Setup
///////////////////////////////////////////////////////////////////////////////

// Sets up the mux router and SmD struct for calling smd functions
func TestMain(m *testing.M) {
	s = new(SmD)

	// v2 APIs
	s.apiRootV2 = "/hsm/v2"
	s.serviceBaseV2 = s.apiRootV2 + "/service"
	s.valuesBaseV2 = s.serviceBaseV2 + "/values"
	s.stateBaseV2 = s.apiRootV2 + "/State"
	s.componentsBaseV2 = s.stateBaseV2 + "/Components"
	s.redfishEPBaseV2 = s.apiRootV2 + "/Inventory/RedfishEndpoints"
	s.nodeMapBaseV2 = s.apiRootV2 + "/Defaults/NodeMaps"
	s.compEPBaseV2 = s.apiRootV2 + "/Inventory/ComponentEndpoints"
	s.serviceEPBaseV2 = s.apiRootV2 + "/Inventory/ServiceEndpoints"
	s.compEthIntBaseV2 = s.apiRootV2 + "/Inventory/EthernetInterfaces"
	s.hsnIntBaseV2 = s.apiRootV2 + "/Inventory/HSNInterfaces"
	s.hwinvByLocBaseV2 = s.apiRootV2 + "/Inventory/Hardware"
	s.hwinvByFRUBaseV2 = s.apiRootV2 + "/Inventory/HardwareByFRU"
	s.invDiscoverBaseV2 = s.apiRootV2 + "/Inventory/Discover"
	s.invDiscStatusBaseV2 = s.apiRootV2 + "/Inventory/DiscoveryStatus"
	s.subscriptionBaseV2 = s.apiRootV2 + "/Subscriptions"
	s.groupsBaseV2 = s.apiRootV2 + "/groups"
	s.partitionsBaseV2 = s.apiRootV2 + "/partitions"
	s.membershipsBaseV2 = s.apiRootV2 + "/memberships"
	s.compLockBaseV2 = s.apiRootV2 + "/locks"
	s.sysInfoBaseV2 = s.apiRootV2 + "/sysinfo"
	s.powerMapBaseV2 = s.sysInfoBaseV2 + "/powermaps"

	s.smapCompEP = NewSyncMap(ComponentEndpointSMap(s))

	s.msgbusHandle = nil

	s.dbDSN = ""
	s.lg = log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags|log.Lmicroseconds)
	s.db, results = NewHMSDB_Test(s.dbDSN, s.lg)
	s.wp = new(base.WorkerPool)

	routes := s.generateRoutes()
	router = s.NewRouter(routes)

	excode := 1
	excode = m.Run()
	os.Exit(excode)
}

///////////////////////////////////////////////////////////////////////////////
// Unit Tests
///////////////////////////////////////////////////////////////////////////////

var jsonErrHMSDSArgMissing = json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"a required argument was missing","status":400}
`)

func TestGetFieldFilterForm(t *testing.T) {
	tests := []struct {
		in          *FieldFltrInForm
		expectedOut hmsds.FieldFilter
	}{{
		&FieldFltrInForm{},
		hmsds.FLTR_DEFAULT,
	}, {
		nil,
		hmsds.FLTR_DEFAULT,
	}, {
		&FieldFltrInForm{
			StateOnly: []string{"true"},
			RoleOnly:  []string{"true"},
		},
		hmsds.FLTR_STATEONLY,
	}, {
		&FieldFltrInForm{
			RoleOnly: []string{"false", "true"},
		},
		hmsds.FLTR_DEFAULT,
	}, {
		&FieldFltrInForm{
			RoleOnly: []string{"foo"},
		},
		hmsds.FLTR_DEFAULT,
	}}
	for i, test := range tests {
		out := getFieldFilterForm(test.in)
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected fieldFilter '%v'; Received fieldFilter '%v'", i, ffStringMap[test.expectedOut], ffStringMap[out])
		}
	}
}

func TestGetFieldFilter(t *testing.T) {
	tests := []struct {
		in          *FieldFltrIn
		expectedOut hmsds.FieldFilter
	}{{
		&FieldFltrIn{false, false, false, false},
		hmsds.FLTR_DEFAULT,
	}, {
		nil,
		hmsds.FLTR_DEFAULT,
	}, {
		&FieldFltrIn{true, false, true, false},
		hmsds.FLTR_STATEONLY,
	}, {
		&FieldFltrIn{false, false, true, false},
		hmsds.FLTR_ROLEONLY,
	}}
	for i, test := range tests {
		out := getFieldFilter(test.in)
		if test.expectedOut != out {
			t.Errorf("Test %v Failed: Expected fieldFilter '%v'; Received fieldFilter '%v'", i, ffStringMap[test.expectedOut], ffStringMap[out])
		}
	}
}

func TestNidRangeToCompFilter(t *testing.T) {
	tests := []struct {
		nidRanges   []string
		f           *hmsds.ComponentFilter
		expectedOut *hmsds.ComponentFilter
		expectedErr bool
	}{{
		[]string{"800"},
		&hmsds.ComponentFilter{},
		&hmsds.ComponentFilter{
			NID: []string{"800"},
		},
		false,
	}, {
		[]string{"800", "16300-16500", "850-900", "700"},
		&hmsds.ComponentFilter{},
		&hmsds.ComponentFilter{
			NID:      []string{"800", "700"},
			NIDStart: []string{"16300", "850"},
			NIDEnd:   []string{"16500", "900"},
		},
		false,
	}, {
		[]string{},
		nil,
		&hmsds.ComponentFilter{},
		false,
	}, {
		[]string{"16300-16500"},
		&hmsds.ComponentFilter{
			NIDStart: []string{"850"},
			NIDEnd:   []string{"900"},
		},
		&hmsds.ComponentFilter{
			NIDStart: []string{"16300", "850"},
			NIDEnd:   []string{"16500", "900"},
		},
		false,
	}, {
		[]string{"-1"},
		&hmsds.ComponentFilter{},
		nil,
		true,
	}}
	for i, test := range tests {
		out, err := nidRangeToCompFilter(test.nidRanges, test.f)
		if err == nil {
			if test.expectedErr {
				t.Errorf("Test %v Failed: Expected an error; Received no error", i)
			} else if !compareFilter(*test.expectedOut, *out) {
				t.Errorf("Test %v Failed: Expected compFilter '%v'; Received compFilter '%v'", i, test.expectedOut, out)
			}
		} else if !test.expectedErr {
			t.Errorf("Test %v Failed: Expected no error; Received error '%s'", i, err.Error())
		}
	}
}

func TestDoReadyGet(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespErr error
		expectedResp []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/service/ready",
		nil,
		json.RawMessage(`{"code":0,"message":"HSM is healthy"}` + "\n"),
	}, {
		"GET",
		"https://localhost/hsm/v2/service/ready",
		hmsds.ErrHMSDSPtrClosed,
		json.RawMessage(`{"type":"about:blank","title":"Service Unavailable","detail":"HSM's database is unhealthy: HMSDS handle is not open.","status":503}` + "\n"),
	}}

	for i, test := range tests {
		results.TestConnection.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 204", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoLivenessGet(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespErr error
		expectedResp []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/service/liveness",
		nil,
		nil,
	}}

	for i, test := range tests {
		results.TestConnection.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Response code was %v; want 204", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentGet(t *testing.T) {
	enabledFlg := true
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespID  *base.Component
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/State/Components/x0c0s27b0n0",
		&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		nil,
		"x0c0s27b0n0",
		json.RawMessage(`{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components/x0c0s27b0n0",
		nil,
		nil,
		"x0c0s27b0n0",
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components/x0c0s27",
		nil,
		hmsds.ErrHMSDSArgMissing, // Could be any HMSError
		"x0c0s27",
		jsonErrHMSDSArgMissing,
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components/x0c0s27",
		nil,
		errors.New("unexpected DB error"), // Could be any non HMSError
		"x0c0s27",
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}
`),
	}}

	for i, test := range tests {
		results.GetComponentByID.Return.id = test.hmsdsRespID
		results.GetComponentByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespID != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespID == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetComponentByID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.GetComponentByID.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentByNIDGet(t *testing.T) {
	enabledFlg := true
	testComp := base.Component{
		ID:       "x0c0s27b0n0",
		Type:     "Node",
		State:    "On",
		Flag:     "OK",
		Enabled:  &enabledFlg,
		SwStatus: "AdminStatus",
		Role:     "Compute",
		SubRole:  "",
		NID:      "864",
		Subtype:  "",
		NetType:  "Sling",
		Arch:     "X86",
		Class:    "",
	}
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespID  *base.Component
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/State/Components/ByNID/864",
		hmsdsRespID:  &testComp,
		hmsdsRespErr: nil,
		expectedID:   "864",
		expectedResp: json.RawMessage(`{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/State/Components/ByNID/864",
		hmsdsRespID:  nil,
		hmsdsRespErr: nil,
		expectedID:   "864",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such NID.","status":404}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/State/Components/ByNID/864",
		hmsdsRespID:  nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing, // Could be any HMS error
		expectedID:   "864",
		expectedResp: jsonErrHMSDSArgMissing,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/State/Components/ByNID/864",
		hmsdsRespID:  nil,
		hmsdsRespErr: errors.New("unexpected DB error"), // Could be any non-HMS error
		expectedID:   "864",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetComponentByNID.Return.id = test.hmsdsRespID
		results.GetComponentByNID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespID != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespID == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetComponentByNID.Input.nid {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.GetComponentByNID.Input.nid)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentDelete(t *testing.T) {
	type testParams struct {
		reqType        string
		reqURI         string
		hmsdsDidDelete bool
		hmsdsRespErr   error
		expectedID     string
		expectedResp   []byte
	}
	tests := []testParams{{
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/State/Components/x0c0s27b0n0",
		hmsdsDidDelete: true,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/State/Components/x0c0s27b0n0",
		hmsdsDidDelete: false,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/State/Components/ ",
		hmsdsDidDelete: false,
		hmsdsRespErr:   hmsds.ErrHMSDSArgBadID,
		expectedID:     " ",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname","status":400}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/State/Components/0c0s27b0n0",
		hmsdsDidDelete: false,
		hmsdsRespErr:   hmsds.ErrHMSDSArgBadID,
		expectedID:     "0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteComponentByID.Return.changed = test.hmsdsDidDelete
		results.DeleteComponentByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsDidDelete && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if w.Code != http.StatusBadRequest && test.expectedID != results.DeleteComponentByID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.DeleteComponentByID.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentDeleteAll(t *testing.T) {
	const numEntries int64 = 42
	type testParams struct {
		reqType          string
		reqURI           string
		hmsdsRespNumRows int64
		hmsdsRespErr     error
		expectedResp     []byte
	}

	// TODO: create test that generates a 400 or update swagger.yml
	tests := []testParams{{

		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/State/Components",
		hmsdsRespNumRows: numEntries,
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted ` + strconv.FormatInt(numEntries, 10) + ` entries"}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/State/Components",
		hmsdsRespNumRows: 0,
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteComponentsAll.Return.numRows = test.hmsdsRespNumRows
		results.DeleteComponentsAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespNumRows != 0 && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentsGet(t *testing.T) {
	enabledFlg := true
	tests := []struct {
		reqType             string
		reqURI              string
		hmsdsRespIDs        []*base.Component
		hmsdsRespErr        error
		expectedFilter      hmsds.ComponentFilter
		expectedFieldFilter hmsds.FieldFilter
		expectedResp        []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/State/Components?type=node",
		[]*base.Component{
			&base.Component{"x0c0s14b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "448", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s15b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "480", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s18b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "576", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s22b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "704", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s24b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "786", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "800", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			Type: []string{"node"},
		},
		hmsds.FLTR_DEFAULT,
		json.RawMessage(`{"Components":[{"ID":"x0c0s14b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":448,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s15b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":480,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s18b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":576,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s22b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":704,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s24b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":786,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s25b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":800,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s26b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":832,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"}]}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components?type=node",
		[]*base.Component{
			&base.Component{"x0c0s14b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "448", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s15b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "480", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s18b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "576", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s22b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "704", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s24b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "786", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "800", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
		hmsds.ErrHMSDSArgBadType,
		hmsds.ComponentFilter{
			Type: []string{"node"},
		},
		hmsds.FLTR_DEFAULT,
		json.RawMessage(`{"type":"about:blank","title":"Method Not Allowed","detail":"allow GET,POST,DELETE","status":405}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components?type=node&type=nodebmc&flagonly=true",
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s26b0n0", "Node", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			Type: []string{"node", "nodebmc"},
		},
		hmsds.FLTR_FLAGONLY,
		json.RawMessage(`{"Components":[{"ID":"x0c0s26b0","Type":"NodeBMC","Flag":"OK"},{"ID":"x0c0s26b0n0","Type":"Node","Flag":"OK"},{"ID":"x0c0s27b0","Type":"NodeBMC","Flag":"OK"},{"ID":"x0c0s27b0n0","Type":"Node","Flag":"OK"}]}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components",
		[]*base.Component{
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27e0", "NodeEnclosure", "On", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{},
		hmsds.FLTR_DEFAULT,
		json.RawMessage(`{"Components":[{"ID":"x0c0s27b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27e0","Type":"NodeEnclosure","State":"On","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"}]}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components?type=foo",
		nil,
		hmsds.ErrHMSDSArgBadType,
		hmsds.ComponentFilter{
			Type: []string{"foo"},
		},
		hmsds.FLTR_DEFAULT,
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not a valid HMS Type","status":400}
`),
	}}

	for i, test := range tests {
		results.GetComponentsFilter.Return.ids = test.hmsdsRespIDs
		results.GetComponentsFilter.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !compareFilter(test.expectedFilter, results.GetComponentsFilter.Input.compFilter) {
			t.Errorf("Test %v Failed: Expected compFilter '%v'; Received compFilter '%v'", i, test.expectedFilter, results.GetComponentsFilter.Input.compFilter)
		}

		if test.expectedFieldFilter != results.GetComponentsFilter.Input.fieldFilter {
			t.Errorf("Test %v Failed: Expected fieldFilter '%v'; Received fieldFilter '%v'", i, ffStringMap[test.expectedFieldFilter], ffStringMap[results.GetComponentsFilter.Input.fieldFilter])
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentsQueryGet(t *testing.T) {
	enabledFlg := true
	tests := []struct {
		reqType             string
		reqURI              string
		hmsdsRespIDs        []*base.Component
		hmsdsRespErr        error
		expectedFilter      hmsds.ComponentFilter
		expectedFieldFilter hmsds.FieldFilter
		expectedIDs         []string
		expectedResp        []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/State/Components/Query/x0c0?type=node",
		[]*base.Component{
			&base.Component{"x0c0s14b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "448", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s15b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "480", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s18b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "576", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s22b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "704", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s24b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "786", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "800", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			Type: []string{"node"},
		},
		hmsds.FLTR_DEFAULT,
		[]string{"x0c0"},
		json.RawMessage(`{"Components":[{"ID":"x0c0s14b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":448,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s15b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":480,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s18b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":576,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s22b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":704,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s24b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":786,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s25b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":800,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s26b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":832,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"}]}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components/Query/x0c0?type=node&type=nodebmc",
		[]*base.Component{
			&base.Component{"x0c0s14b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s14b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "448", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s15b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s15b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "480", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s18b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s18b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "576", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s22b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s22b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "704", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s24b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s24b0n0", "Node", "Off", "OK", &enabledFlg, "AdminStatus", "Compute", "", "786", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s25b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "800", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			Type: []string{"node", "nodebmc"},
		},
		hmsds.FLTR_DEFAULT,
		[]string{"x0c0"},
		json.RawMessage(`{"Components":[{"ID":"x0c0s14b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s14b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":448,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s15b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s15b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":480,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s18b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s18b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":576,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s22b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s22b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":704,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s24b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s24b0n0","Type":"Node","State":"Off","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":786,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s25b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s25b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":800,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s26b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s26b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":832,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"}]}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components/Query/x0c0s27",
		[]*base.Component{
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27e0", "NodeEnclosure", "On", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{},
		hmsds.FLTR_DEFAULT,
		[]string{"x0c0s27"},
		json.RawMessage(`{"Components":[{"ID":"x0c0s27b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27e0","Type":"NodeEnclosure","State":"On","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"}]}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/State/Components/Query/x0c0s27?type=foo",
		nil,
		hmsds.ErrHMSDSArgBadType,
		hmsds.ComponentFilter{
			Type: []string{"foo"},
		},
		hmsds.FLTR_DEFAULT,
		[]string{"x0c0s27"},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not a valid HMS Type","status":400}
`),
	}}

	for i, test := range tests {
		results.GetComponentsQuery.Return.ids = test.hmsdsRespIDs
		results.GetComponentsQuery.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !compareFilter(test.expectedFilter, results.GetComponentsQuery.Input.compFilter) {
			t.Errorf("Test %v Failed: Expected compFilter '%v'; Received compFilter '%v'", i, test.expectedFilter, results.GetComponentsQuery.Input.compFilter)
		}

		if test.expectedFieldFilter != results.GetComponentsQuery.Input.fieldFilter {
			t.Errorf("Test %v Failed: Expected fieldFilter '%v'; Received fieldFilter '%v'", i, ffStringMap[test.expectedFieldFilter], ffStringMap[results.GetComponentsQuery.Input.fieldFilter])
		}

		if !compareIDs(test.expectedIDs, results.GetComponentsQuery.Input.ids) {
			t.Errorf("Test %v Failed: Expected compArray '%v'; Received compArray '%v'", i, test.expectedIDs, results.GetComponentsQuery.Input.ids)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentsQueryPost(t *testing.T) {
	enabledFlg := true
	tests := []struct {
		reqType             string
		reqURI              string
		reqBody             []byte
		hmsdsRespIDs        []*base.Component
		hmsdsRespErr        error
		expectedFilter      hmsds.ComponentFilter
		expectedFieldFilter hmsds.FieldFilter
		expectedIDs         []string
		expectedResp        []byte
	}{{
		"POST",
		"https://localhost/hsm/v2/State/Components/Query",
		json.RawMessage(`{"componentids":["x0c0s27b0","x0c0s25b0"],"type":["node"],"stateonly":true}`),
		[]*base.Component{
			&base.Component{"x0c0s25b0n0", "Node", "On", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			Type: []string{"node"},
		},
		hmsds.FLTR_STATEONLY,
		[]string{"x0c0s27b0", "x0c0s25b0"},
		json.RawMessage(`{"Components":[{"ID":"x0c0s25b0n0","Type":"Node","State":"On","Flag":"OK"},{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK"}]}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/State/Components/Query",
		json.RawMessage(`{"componentids":["x0c0s27","x0c0s25"],"stateonly":false,"type":["node","nodebmc"]}`),
		[]*base.Component{
			&base.Component{"x0c0s25b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "800", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			Type: []string{"node", "nodebmc"},
		},
		hmsds.FLTR_DEFAULT,
		[]string{"x0c0s27", "x0c0s25"},
		json.RawMessage(`{"Components":[{"ID":"x0c0s25b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s25b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":800,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0","Type":"NodeBMC","State":"Ready","Flag":"OK","Enabled":true,"NetType":"Sling","Arch":"X86"},{"ID":"x0c0s27b0n0","Type":"Node","State":"On","Flag":"OK","Enabled":true,"SoftwareStatus":"AdminStatus","Role":"Compute","NID":864,"NetType":"Sling","Arch":"X86"}]}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/State/Components/Query",
		json.RawMessage(`{"componentids":["x0c0s25","x0c0s27"],"flagonly":true}`),
		[]*base.Component{
			&base.Component{"x0c0s25b0", "NodeBMC", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s25e0", "NodeEnclosure", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27e0", "NodeEnclosure", "", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{},
		hmsds.FLTR_FLAGONLY,
		[]string{"x0c0s27", "x0c0s25"},
		json.RawMessage(`{"Components":[{"ID":"x0c0s25b0","Type":"NodeBMC","Flag":"OK"},{"ID":"x0c0s25b0n0","Type":"Node","Flag":"OK"},{"ID":"x0c0s25e0","Type":"NodeEnclosure","Flag":"OK"},{"ID":"x0c0s27b0","Type":"NodeBMC","Flag":"OK"},{"ID":"x0c0s27b0n0","Type":"Node","Flag":"OK"},{"ID":"x0c0s27e0","Type":"NodeEnclosure","Flag":"OK"}]}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/State/Components/Query",
		json.RawMessage(`{"componentids":["x0c0","x0c1"],"type":["foo"]}`),
		nil,
		hmsds.ErrHMSDSArgBadType,
		hmsds.ComponentFilter{
			Type: []string{"foo"},
		},
		hmsds.FLTR_DEFAULT,
		[]string{"x0c0", "x0c1"},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not a valid HMS Type","status":400}
`),
	}}

	for i, test := range tests {
		results.GetComponentsQuery.Return.ids = test.hmsdsRespIDs
		results.GetComponentsQuery.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !compareFilter(test.expectedFilter, results.GetComponentsQuery.Input.compFilter) {
			t.Errorf("Test %v Failed: Expected compFilter '%v'; Received compFilter '%v'", i, test.expectedFilter, results.GetComponentsQuery.Input.compFilter)
		}

		if test.expectedFieldFilter != results.GetComponentsQuery.Input.fieldFilter {
			t.Errorf("Test %v Failed: Expected fieldFilter '%v'; Received fieldFilter '%v'", i, ffStringMap[test.expectedFieldFilter], ffStringMap[results.GetComponentsQuery.Input.fieldFilter])
		}

		if !compareIDs(test.expectedIDs, results.GetComponentsQuery.Input.ids) {
			t.Errorf("Test %v Failed: Expected compArray '%v'; Received compArray '%v'", i, test.expectedIDs, results.GetComponentsQuery.Input.ids)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentByNIDQueryPost(t *testing.T) {
	tests := []struct {
		reqType             string
		reqURI              string
		reqBody             []byte
		hmsdsRespIDs        []*base.Component
		hmsdsRespErr        error
		expectedFilter      hmsds.ComponentFilter
		expectedFieldFilter hmsds.FieldFilter
		expectedResp        []byte
	}{{
		"POST",
		"https://localhost/hsm/v2/State/Components/ByNID/Query",
		json.RawMessage(`{"nidranges":["800","16300-16500"],"nidonly":true}`),
		[]*base.Component{
			&base.Component{"x0c0s25b0n0", "Node", "", "", nil, "", "", "", "800", "", "", "", "", false, false},
			&base.Component{"x1c0s0b0n0", "Node", "", "", nil, "", "", "", "16384", "", "", "", "", false, false},
			&base.Component{"x1c0s1b0n0", "Node", "", "", nil, "", "", "", "16416", "", "", "", "", false, false},
			&base.Component{"x1c0s2b0n0", "Node", "", "", nil, "", "", "", "16448", "", "", "", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			NID:      []string{"800"},
			NIDStart: []string{"16300"},
			NIDEnd:   []string{"16500"},
		},
		hmsds.FLTR_NIDONLY,
		json.RawMessage(`{"Components":[{"ID":"x0c0s25b0n0","Type":"Node","NID":800},{"ID":"x1c0s0b0n0","Type":"Node","NID":16384},{"ID":"x1c0s1b0n0","Type":"Node","NID":16416},{"ID":"x1c0s2b0n0","Type":"Node","NID":16448}]}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/State/Components/ByNID/Query",
		json.RawMessage(`{"nidranges":["800","16300-16500","440-480","768"],"nidonly":true}`),
		[]*base.Component{
			&base.Component{"x0c0s14b0n0", "Node", "", "", nil, "", "", "", "448", "", "", "", "", false, false},
			&base.Component{"x0c0s15b0n0", "Node", "", "", nil, "", "", "", "480", "", "", "", "", false, false},
			&base.Component{"x0c0s24b0n0", "Node", "", "", nil, "", "", "", "768", "", "", "", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "", "", nil, "", "", "", "800", "", "", "", "", false, false},
			&base.Component{"x1c0s0b0n0", "Node", "", "", nil, "", "", "", "16384", "", "", "", "", false, false},
			&base.Component{"x1c0s1b0n0", "Node", "", "", nil, "", "", "", "16416", "", "", "", "", false, false},
			&base.Component{"x1c0s2b0n0", "Node", "", "", nil, "", "", "", "16448", "", "", "", "", false, false},
		},
		nil,
		hmsds.ComponentFilter{
			NID:      []string{"800", "768"},
			NIDStart: []string{"16300", "440"},
			NIDEnd:   []string{"16500", "480"},
		},
		hmsds.FLTR_NIDONLY,
		json.RawMessage(`{"Components":[{"ID":"x0c0s14b0n0","Type":"Node","NID":448},{"ID":"x0c0s15b0n0","Type":"Node","NID":480},{"ID":"x0c0s24b0n0","Type":"Node","NID":768},{"ID":"x0c0s25b0n0","Type":"Node","NID":800},{"ID":"x1c0s0b0n0","Type":"Node","NID":16384},{"ID":"x1c0s1b0n0","Type":"Node","NID":16416},{"ID":"x1c0s2b0n0","Type":"Node","NID":16448}]}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/State/Components/ByNID/Query",
		json.RawMessage(`{"nidonly":true}`),
		nil,
		errors.New("Missing NID ranges"),
		hmsds.ComponentFilter{},
		hmsds.FLTR_DEFAULT,
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing NID ranges","status":400}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/State/Components/ByNID/Query",
		json.RawMessage(`{"nidranges":["foo"],"nidonly":true}`),
		[]*base.Component{},
		nil,
		hmsds.ComponentFilter{
			NID: []string{"foo"},
		},
		hmsds.FLTR_NIDONLY,
		json.RawMessage(`{"Components":[]}
`),
	}}

	for i, test := range tests {
		// Clear the results
		results.GetComponentsFilter.Input.compFilter = hmsds.ComponentFilter{}
		results.GetComponentsFilter.Input.fieldFilter = hmsds.FLTR_DEFAULT
		results.GetComponentsFilter.Return.ids = test.hmsdsRespIDs
		results.GetComponentsFilter.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !compareFilter(test.expectedFilter, results.GetComponentsFilter.Input.compFilter) {
			t.Errorf("Test %v Failed: Expected compFilter '%v'; Received compFilter '%v'", i, test.expectedFilter, results.GetComponentsFilter.Input.compFilter)
		}

		if test.expectedFieldFilter != results.GetComponentsFilter.Input.fieldFilter {
			t.Errorf("Test %v Failed: Expected fieldFilter '%v'; Received fieldFilter '%v'", i, ffStringMap[test.expectedFieldFilter], ffStringMap[results.GetComponentsFilter.Input.fieldFilter])
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

// Expected when SMD layer returns error ErrSMDBadID
var ErrSMDBadIDJSON = json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname ID","status":400}
`)

// Expected output when HMSDS API layer returns error hmsds.ErrHMSDSArgNoMatch
var ErrHMSDSArgNoMatchJSON = json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"a required argument did not match any valid input","status":400}
`)

// Expected output when HMSDS API layer returns error hmsds.ErrHMSDSArgMissing
var ErrHMSDSArgMissingJSON = json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"a required argument was missing","status":400}
`)

// Calls hmsds.UpdateCompStates for both single and bulk updates.

func TestDoCompStateDataPatch(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsRespErr  error
		expectedID    string
		expectedState string
		expectedFlag  string
		expectedForce bool
		expectedResp  []byte
	}{{
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/StateData",
		reqBody:       json.RawMessage(`{"state": "off"}`),
		hmsdsRespErr:  nil,
		expectedID:    "x0c0s27b0n0",
		expectedState: "off",
		expectedFlag:  "OK", // added by front end now
		expectedForce: false,
		expectedResp:  nil,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/State/Components/0c0s27b0n0/StateData",
		reqBody:       json.RawMessage(`{"state": "off"}`),
		hmsdsRespErr:  ErrSMDBadID,
		expectedID:    "", // Missing due to pre-db error.
		expectedState: "",
		expectedFlag:  "",
		expectedForce: false,
		expectedResp:  ErrSMDBadIDJSON,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/StateData",
		reqBody:       json.RawMessage(`{"state": "off","flag": "ok", "UpdateType":"StateData"}`),
		hmsdsRespErr:  nil,
		expectedID:    "x0c0s27b0n0",
		expectedState: "off",
		expectedFlag:  "ok",
		expectedForce: false,
		expectedResp:  nil,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/StateData",
		reqBody:       json.RawMessage(`{"state": "offf","flag": "ok"}`),
		hmsdsRespErr:  hmsds.ErrHMSDSArgNoMatch,
		expectedID:    "x0c0s27b0n0",
		expectedState: "offf",
		expectedFlag:  "ok",
		expectedForce: false,
		expectedResp:  ErrHMSDSArgNoMatchJSON,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/StateData",
		reqBody:       json.RawMessage(`{"state": "foo", "Force":true}`),
		hmsdsRespErr:  hmsds.ErrHMSDSArgNoMatch,
		expectedID:    "x0c0s27b0n0",
		expectedState: "foo",
		expectedFlag:  "OK", // filled in automatically when omitted.
		expectedForce: true,
		expectedResp:  ErrHMSDSArgNoMatchJSON,
	}}

	for i, test := range tests {
		// Clear the results
		results.UpdateCompStates.Input.ids = []string{}
		results.UpdateCompStates.Input.state = ""
		results.UpdateCompStates.Input.flag = ""
		results.UpdateCompStates.Input.force = false
		results.UpdateCompStates.Input.pi = nil

		results.UpdateCompStates.Return.affectedIds = []string{test.expectedID}
		results.UpdateCompStates.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if len(results.UpdateCompStates.Input.ids) > 0 &&
			test.expectedID != results.UpdateCompStates.Input.ids[0] {
			t.Errorf("Test %v Failed: Expected ids '%v'; Received id '%v'",
				i, test.expectedID, results.UpdateCompStates.Input.ids[0])
		}
		if test.expectedState != results.UpdateCompStates.Input.state {
			t.Errorf("Test %v Failed: Expected state '%s'; Received state '%s'",
				i, test.expectedState, results.UpdateCompStates.Input.state)
		}
		if test.expectedFlag != results.UpdateCompStates.Input.flag {
			t.Errorf("Test %v Failed: Expected flag '%v'; Received flag '%v'",
				i, test.expectedFlag, results.UpdateCompStates.Input.flag)
		}
		if test.expectedForce != results.UpdateCompStates.Input.force {
			t.Errorf("Test %v Failed: Expected force '%v'; Received force '%v'",
				i, test.expectedForce, results.UpdateCompStates.Input.force)
		}
		if test.hmsdsRespErr == nil && results.UpdateCompStates.Input.pi == nil {
			t.Errorf("Test %v Failed: Expected non-nil pi; Received nil", i)
		}
	}
}

func TestDoNodeMapGet(t *testing.T) {
	nodeMap, _ := sm.NewNodeMap("x0c0s27b0n0", "Compute", "", 1, nil)
	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespEntry *sm.NodeMap
		hmsdsRespErr   error
		expectedID     string
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27b0n0",
		hmsdsRespEntry: nodeMap,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"ID":"x0c0s27b0n0","NID":1,"Role":"Compute"}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27b0n0",
		hmsdsRespEntry: nil,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27",
		hmsdsRespEntry: nil,
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedID:     "x0c0s27",
		expectedResp:   jsonErrHMSDSArgMissing,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27",
		hmsdsRespEntry: nil,
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedID:     "x0c0s27",
		expectedResp:   jsonErrHMSDSArgMissing,
	}}

	for i, test := range tests {
		results.GetNodeMapByID.Return.entry = test.hmsdsRespEntry
		results.GetNodeMapByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespEntry != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespEntry == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetNodeMapByID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.GetNodeMapByID.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoNodeMapsGet(t *testing.T) {
	var nodeMaps = []*sm.NodeMap{
		{ID: "x0c0s21b0n0", NID: 151, Role: "Compute"},
		{ID: "x0c0s22b0n0", NID: 152, Role: "Compute"},
		{ID: "x0c0s23b0n0", NID: 153, Role: "Compute"},
		{ID: "x0c0s24b0n0", NID: 154, Role: "Compute"},
		{ID: "x0c0s25b0n0", NID: 155, Role: "Compute"},
		{ID: "x0c0s26b0n0", NID: 156, Role: "Compute"},
		{ID: "x0c0s27b0n0", NID: 157, Role: "Compute"},
		{ID: "x0c0s28b0n0", NID: 158, Role: "Compute"},
		{ID: "x0c0s29b0n0", NID: 159, Role: "Compute"},
		{ID: "x0c0s31b0n0", NID: 181, Role: "Service"},
		{ID: "x0c0s32b0n0", NID: 182, Role: "Service"},
		{ID: "x0c0s33b0n0", NID: 183, Role: "Service"},
	}

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespEntry []*sm.NodeMap
		hmsdsRespErr   error
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps",
		hmsdsRespEntry: nodeMaps[0:1],
		hmsdsRespErr:   nil,
		expectedResp:   json.RawMessage(`{"NodeMaps":[{"ID":"x0c0s21b0n0","NID":151,"Role":"Compute"}]}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps",
		hmsdsRespEntry: nodeMaps[4:7],
		hmsdsRespErr:   nil,
		expectedResp:   json.RawMessage(`{"NodeMaps":[{"ID":"x0c0s25b0n0","NID":155,"Role":"Compute"},{"ID":"x0c0s26b0n0","NID":156,"Role":"Compute"},{"ID":"x0c0s27b0n0","NID":157,"Role":"Compute"}]}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps",
		hmsdsRespEntry: nodeMaps[9:12],
		hmsdsRespErr:   nil,
		expectedResp:   json.RawMessage(`{"NodeMaps":[{"ID":"x0c0s31b0n0","NID":181,"Role":"Service"},{"ID":"x0c0s32b0n0","NID":182,"Role":"Service"},{"ID":"x0c0s33b0n0","NID":183,"Role":"Service"}]}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps?ID=foo",
		hmsdsRespEntry: nil,
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedResp:   jsonErrHMSDSArgMissing,
	}}

	for i, test := range tests {
		results.GetNodeMapsAll.Return.entry = test.hmsdsRespEntry
		results.GetNodeMapsAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespEntry != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespEntry == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoNodeMapDelete(t *testing.T) {
	type testParams struct {
		reqType        string
		reqURI         string
		hmsdsDidDelete bool
		hmsdsRespErr   error
		expectedID     string
		expectedResp   []byte
	}
	tests := []testParams{{
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27b0n0",
		hmsdsDidDelete: true,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27b0n0",
		hmsdsDidDelete: false,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/0c0s27b0n0",
		hmsdsDidDelete: false,
		hmsdsRespErr:   hmsds.ErrHMSDSArgBadID,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname","status":400}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Defaults/NodeMaps/0c0s27b0n0",
		hmsdsDidDelete: false,
		hmsdsRespErr:   hmsds.ErrHMSDSArgBadID,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteNodeMapByID.Return.changed = test.hmsdsDidDelete
		results.DeleteNodeMapByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsDidDelete && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if w.Code != http.StatusBadRequest && test.expectedID != results.DeleteNodeMapByID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.DeleteNodeMapByID.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoNodeMapDeleteAll(t *testing.T) {
	type testParams struct {
		reqType       string
		reqURI        string
		hmsdsRespErr  error
		expectedCount int64
		expectedResp  []byte
	}
	tests := []testParams{{
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/Defaults/NodeMaps",
		hmsdsRespErr:  nil,
		expectedCount: 3,
		expectedResp:  json.RawMessage(`{"code":0,"message":"deleted 3 entries"}` + "\n"),
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/Defaults/NodeMaps",
		hmsdsRespErr:  nil,
		expectedCount: 0,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteNodeMapsAll.Return.numRows = test.expectedCount
		results.DeleteNodeMapsAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.expectedCount != 0 && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if w.Code != http.StatusBadRequest && test.expectedCount != results.DeleteNodeMapsAll.Return.numRows {
			t.Errorf("Test %v Failed: Expected count '%v'; Received count '%v'", i, test.expectedCount, results.DeleteNodeMapsAll.Return.numRows)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoNodeMapPost(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedResp []byte
	}{{
		"POST",
		"https://localhost/hsm/v2/Defaults/NodeMaps",
		json.RawMessage(`{ "NodeMaps": [ { "ID": "x0c0s0b0n0", "NID": 1, "Role": "Compute" } ]}`),
		nil,
		json.RawMessage(`{"code":0,"message":"Created or modified 1 entries"}` + "\n"),
	}, {
		"POST",
		"https://localhost/hsm/v2/Defaults/NodeMaps",
		json.RawMessage(`{ "NodeMaps": [ { "ID": "x0c0s21b0n0", "NID": 151, "Role": "Compute" }, { "ID": "x0c0s22b0n0", "NID": 152, "Role": "Compute" }  ]}`),
		nil,
		json.RawMessage(`{"code":0,"message":"Created or modified 2 entries"}` + "\n"),
	}, {
		"POST",
		"https://localhost/hsm/v2/Defaults/NodeMaps",
		json.RawMessage(`{ "NodeMaps": [ { "ID": "x0c0s0b0n0", "NID": 1, "Role": "Compute" } ]}`),
		hmsds.ErrHMSDSDuplicateKey,
		json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing xname ID that has the same NID.","status":409}` + "\n"),
	}, {
		"POST",
		"https://localhost/hsm/v2/Defaults/NodeMaps",
		json.RawMessage(`{ "NodeMaps": [  ]}`),
		nil,
		json.RawMessage(`{"code":0,"message":"Created or modified 0 entries"}` + "\n"),
	}, {
		"POST",
		"https://localhost/hsm/v2/Defaults/NodeMaps",
		nil,
		hmsds.ErrHMSDSArgMissing,
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"error decoding JSON unexpected end of JSON input","status":500}` + "\n"),
	}, {
		"POST",
		"https://localhost/hsm/v2/Defaults/NodeMaps",
		json.RawMessage(`{ "NoteMaps": [ { "ID": "x0c0s0b0n0", "NID": 1, "Role": "Compute" } ]}`),
		hmsds.ErrHMSDSArgNoMatch,
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"operation 'POST' failed during store. ","status":500}` + "\n"),
	}}
	for i, test := range tests {
		// Clear the results
		results.InsertNodeMaps.Input.nnms = nil
		results.InsertNodeMaps.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoNodeMapPut(t *testing.T) {
	nodeMap, _ := sm.NewNodeMap("x0c0s27b0n0", "Compute", "", 1, nil)
	tests := []struct {
		reqType         string
		reqURI          string
		reqBody         []byte
		expectedNodeMap *sm.NodeMap
		hmsdsRespErr    error
		expectedResp    []byte
	}{{
		reqType:         "PUT",
		reqURI:          "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27b0n0",
		reqBody:         json.RawMessage(`{"NID": 1,"Role": "Compute"}`),
		expectedNodeMap: nodeMap,
		hmsdsRespErr:    nil,
		expectedResp:    json.RawMessage(`{"ID":"x0c0s27b0n0","NID":1,"Role":"Compute"}` + "\n"),
	}, {
		reqType:         "PUT",
		reqURI:          "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27b0n0",
		reqBody:         json.RawMessage(`{"NID": 1,"Role": "Compute"}`),
		expectedNodeMap: nodeMap,
		hmsdsRespErr:    hmsds.ErrHMSDSDuplicateKey,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing resource that has the same NID","status":409}` + "\n"),
	}, {
		reqType:         "PUT",
		reqURI:          "https://localhost/hsm/v2/Defaults/NodeMaps/x0c0s27b0n0",
		reqBody:         json.RawMessage(`{}`),
		expectedNodeMap: &sm.NodeMap{},
		hmsdsRespErr:    hmsds.ErrHMSDSArgMissingNID,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate endpoint data: NID '0' is out of range","status":400}` + "\n"),
	}, {
		reqType:         "PUT",
		reqURI:          "https://localhost/hsm/v2/Defaults/NodeMaps/0c0s27b0n0",
		reqBody:         json.RawMessage(`{"NID": 1,"Role": "Compute"}`),
		expectedNodeMap: &sm.NodeMap{},
		hmsdsRespErr:    hmsds.ErrHMSDSArgBadID,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate endpoint data: xname ID '0c0s27b0n0' is invalid or not a node","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.InsertNodeMap.Input.m = &sm.NodeMap{}
		results.InsertNodeMap.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.expectedNodeMap != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.expectedNodeMap == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if *test.expectedNodeMap != *results.InsertNodeMap.Input.m {
			t.Errorf("Test %v Failed: Expected nodeMap '%v'; Received NodeMap '%v'", i, test.expectedNodeMap, results.InsertNodeMap.Input.m)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompFlagOnlyPatch(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedID   string
		expectedFlag string
		expectedRows int64
		expectedResp []byte
	}{{
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/FlagOnly",
		reqBody:      json.RawMessage(`{"flag": "alert"}`),
		hmsdsRespErr: nil,
		expectedID:   "x0c0s27b0n0",
		expectedFlag: "alert", // added by front end now
		expectedRows: 1,
		expectedResp: nil,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/0c0s27b0n0/FlagOnly",
		reqBody:      json.RawMessage(`{"flag": "alert"}`),
		hmsdsRespErr: ErrSMDBadID,
		expectedID:   "", // Missing due to pre-db error.
		expectedFlag: "",
		expectedRows: 0,
		expectedResp: ErrSMDBadIDJSON,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/FlagOnly",
		reqBody:      json.RawMessage(`{"flag": "warn"}`),
		hmsdsRespErr: nil,
		expectedID:   "x0c0s27b0n0",
		expectedFlag: "warn",
		expectedRows: 1,
		expectedResp: nil,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/FlagOnly",
		reqBody:      json.RawMessage(`{"flag": "foo"}`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedID:   "x0c0s27b0n0",
		expectedFlag: "foo",
		expectedRows: 0,
		expectedResp: ErrHMSDSArgNoMatchJSON,
	}}

	for i, test := range tests {
		// Clear the results
		results.UpdateCompFlagOnly.Input.id = ""
		results.UpdateCompFlagOnly.Input.flag = ""

		results.UpdateCompFlagOnly.Return.rowsAffected = test.expectedRows
		results.UpdateCompFlagOnly.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if test.expectedID != results.UpdateCompFlagOnly.Input.id {
			t.Errorf("Test %v Failed: Expected id '%v'; Received id '%v'",
				i, test.expectedID, results.UpdateCompFlagOnly.Input.id)
		}
		if test.expectedFlag != results.UpdateCompFlagOnly.Input.flag {
			t.Errorf("Test %v Failed: Expected flag '%v'; Received flag '%v'",
				i, test.expectedFlag, results.UpdateCompFlagOnly.Input.flag)
		}
		if test.expectedRows != results.UpdateCompFlagOnly.Return.rowsAffected {
			t.Errorf("Test %v Failed: Expected rows '%v'; Received rows '%v'",
				i, test.expectedRows, results.UpdateCompFlagOnly.Return.rowsAffected)
		}
		if test.expectedFlag != results.UpdateCompFlagOnly.Input.flag {
			t.Errorf("Test %v Failed: Expected flag '%v'; Received flag '%v'",
				i, test.expectedFlag, results.UpdateCompFlagOnly.Input.flag)
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompBulkStateDataPatch(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsRespErr  error
		expectedIds   []string
		expectedState string
		expectedFlag  string
		expectedForce bool
		expectedResp  []byte
	}{{
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkStateData",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"state": "off"}`),
		nil,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		"off",
		"OK", // added by front end now
		false,
		nil,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkStateData",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","xx0c0s25b0n0"],"state": "off"}`),
		ErrSMDBadID,
		[]string{}, // Missing due to pre-db error.
		"",
		"",
		false,
		ErrSMDBadIDJSON,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkStateData",
		json.RawMessage(`{"componentids":["x0c0s27b0n0"],"state": "offf","flag": "ok"}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{"x0c0s27b0n0"},
		"offf",
		"ok",
		false,
		ErrHMSDSArgNoMatchJSON,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkStateData",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"state": "foo", "Force":true}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		"foo",
		"OK", // filled in automatically when omitted.
		true,
		ErrHMSDSArgNoMatchJSON,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkStateData",
		json.RawMessage(`{"state": "foo"}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"",
		"",
		false,
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing IDs","status":400}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkStateData",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"]}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"",
		"",
		false,
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing State","status":400}
`),
	}}

	for i, test := range tests {
		// Clear the results
		results.UpdateCompStates.Input.ids = []string{}
		results.UpdateCompStates.Input.state = ""
		results.UpdateCompStates.Input.flag = ""
		results.UpdateCompStates.Input.force = false
		results.UpdateCompStates.Input.pi = nil

		results.UpdateCompStates.Return.affectedIds = test.expectedIds
		results.UpdateCompStates.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !compareIDs(test.expectedIds, results.UpdateCompStates.Input.ids) {
			t.Errorf("Test %v Failed: Expected ids '%v'; Received ids '%v'",
				i, test.expectedIds, results.UpdateCompStates.Input.ids)
		}
		if test.expectedState != results.UpdateCompStates.Input.state {
			t.Errorf("Test %v Failed: Expected state '%s'; Received state '%s'",
				i, test.expectedState, results.UpdateCompStates.Input.state)
		}
		if test.expectedFlag != results.UpdateCompStates.Input.flag {
			t.Errorf("Test %v Failed: Expected flag '%v'; Received flag '%v'",
				i, test.expectedFlag, results.UpdateCompStates.Input.flag)
		}
		if test.expectedForce != results.UpdateCompStates.Input.force {
			t.Errorf("Test %v Failed: Expected force '%v'; Received force '%v'",
				i, test.expectedForce, results.UpdateCompStates.Input.force)
		}
		if test.hmsdsRespErr == nil && results.UpdateCompStates.Input.pi == nil {
			t.Errorf("Test %v Failed: Expected non-nil pi; Received nil", i)
		}
	}
}

func TestDoCompBulkFlagOnlyPatch(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedIds  []string
		expectedFlag string
		expectedResp []byte
	}{{
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkFlagOnly",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"flag": "alert"}`),
		nil,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		"alert",
		nil,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkFlagOnly",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"flag": "foo"}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		"foo",
		ErrHMSDSArgNoMatchJSON,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkFlagOnly",
		json.RawMessage(`{"flag": "warn"}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"",
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing IDs","status":400}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkFlagOnly",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"]}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"",
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing Flag","status":400}
`),
	}}

	for i, test := range tests {
		// Clear the results
		results.BulkUpdateCompFlagOnly.Input.ids = []string{}
		results.BulkUpdateCompFlagOnly.Input.flag = ""
		results.BulkUpdateCompFlagOnly.Return.affectedIds = test.expectedIds
		results.BulkUpdateCompFlagOnly.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !compareIDs(test.expectedIds, results.BulkUpdateCompFlagOnly.Input.ids) {
			t.Errorf("Test %v Failed: Expected ids '%v'; Received ids '%v'", i, test.expectedIds, results.BulkUpdateCompFlagOnly.Input.ids)
		}

		if test.expectedFlag != results.BulkUpdateCompFlagOnly.Input.flag {
			t.Errorf("Test %v Failed: Expected flag '%v'; Received flag '%v'", i, test.expectedFlag, results.BulkUpdateCompFlagOnly.Input.flag)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompBulkEnabledPatch(t *testing.T) {
	tests := []struct {
		reqType         string
		reqURI          string
		reqBody         []byte
		hmsdsRespErr    error
		expectedIds     []string
		expectedEnabled bool
		expectedResp    []byte
	}{{
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkEnabled",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"Enabled": true}`),
		nil,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		true,
		nil,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkEnabled",
		json.RawMessage(`{"Enabled": true}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		true,
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing IDs","status":400}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkEnabled",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"]}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		false,
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing Enabled field","status":400}
`),
	}}

	for i, test := range tests {
		// Clear the results
		results.BulkUpdateCompEnabled.Input.ids = []string{}
		results.BulkUpdateCompEnabled.Input.enabled = test.expectedEnabled
		results.BulkUpdateCompEnabled.Return.affectedIds = test.expectedIds
		results.BulkUpdateCompEnabled.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !compareIDs(test.expectedIds, results.BulkUpdateCompEnabled.Input.ids) {
			t.Errorf("Test %v Failed: Expected ids '%v'; Received ids '%v'",
				i, test.expectedIds, results.BulkUpdateCompEnabled.Input.ids)
		}

		if test.expectedEnabled != results.BulkUpdateCompEnabled.Input.enabled {
			t.Errorf("Test %v Failed: Expected Enabled '%v'; Received Enabled '%v'",
				i, test.expectedEnabled, results.BulkUpdateCompEnabled.Input.enabled)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompBulkSwStatusPatch(t *testing.T) {
	tests := []struct {
		reqType          string
		reqURI           string
		reqBody          []byte
		hmsdsRespErr     error
		expectedIds      []string
		expectedSwStatus string
		expectedResp     []byte
	}{{
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkSoftwareStatus",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"SoftwareStatus": "Something"}`),
		nil,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		"Something", // Arbitrary, we don't verify the states - managed plane does this.
		nil,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkSoftwareStatus",
		json.RawMessage(`{"SoftwareStatus": "Something"}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"Something",
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing IDs","status":400}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkSoftwareStatus",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"]}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"Something",
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing SoftwareStatus","status":400}
`),
	}}

	for i, test := range tests {
		// Clear the results
		results.BulkUpdateCompSwStatus.Input.ids = []string{}
		results.BulkUpdateCompSwStatus.Input.swstatus = test.expectedSwStatus
		results.BulkUpdateCompSwStatus.Return.affectedIds = test.expectedIds
		results.BulkUpdateCompSwStatus.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !compareIDs(test.expectedIds, results.BulkUpdateCompSwStatus.Input.ids) {
			t.Errorf("Test %v Failed: Expected ids '%v'; Received ids '%v'",
				i, test.expectedIds, results.BulkUpdateCompSwStatus.Input.ids)
		}

		if test.expectedSwStatus != results.BulkUpdateCompSwStatus.Input.swstatus {
			t.Errorf("Test %v Failed: Expected SwStatus '%v'; Received SwStatus '%v'",
				i, test.expectedSwStatus, results.BulkUpdateCompSwStatus.Input.swstatus)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompRolePatch(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedID   string
		expectedRole string
		expectedRows int64
		expectedResp []byte
	}{{
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/Role",
		reqBody:      json.RawMessage(`{"Role": "Compute"}`),
		hmsdsRespErr: nil,
		expectedID:   "x0c0s27b0n0",
		expectedRole: "Compute", // added by front end now
		expectedRows: 1,
		expectedResp: nil,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/0c0s27b0n0/Role",
		reqBody:      json.RawMessage(`{"Role": "Service"}`),
		hmsdsRespErr: ErrSMDBadID,
		expectedID:   "", // Missing due to pre-db error.
		expectedRole: "",
		expectedRows: 0,
		expectedResp: ErrSMDBadIDJSON,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/Role",
		reqBody:      json.RawMessage(`{"Role": "Service"}`),
		hmsdsRespErr: nil,
		expectedID:   "x0c0s27b0n0",
		expectedRole: "Service",
		expectedRows: 1,
		expectedResp: nil,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/Role",
		reqBody:      json.RawMessage(`{"Role": ""}`),
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedID:   "x0c0s27b0n0",
		expectedRole: "",
		expectedRows: 0,
		expectedResp: ErrHMSDSArgMissingJSON,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s27b0n0/Role",
		reqBody:      json.RawMessage(`{"Role": "foo"}`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedID:   "x0c0s27b0n0",
		expectedRole: "foo",
		expectedRows: 0,
		expectedResp: ErrHMSDSArgNoMatchJSON,
	}}

	for i, test := range tests {
		// Clear the results
		results.UpdateCompRole.Input.id = ""
		results.UpdateCompRole.Input.role = ""

		results.UpdateCompRole.Return.rowsAffected = test.expectedRows
		results.UpdateCompRole.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if test.expectedID != results.UpdateCompRole.Input.id {
			t.Errorf("Test %v Failed: Expected id '%v'; Received id '%v'",
				i, test.expectedID, results.UpdateCompRole.Input.id)
		}
		if test.expectedRole != results.UpdateCompRole.Input.role {
			t.Errorf("Test %v Failed: Expected flag '%v'; Received flag '%v'",
				i, test.expectedRole, results.UpdateCompRole.Input.role)
		}
		if test.expectedRows != results.UpdateCompRole.Return.rowsAffected {
			t.Errorf("Test %v Failed: Expected rows '%v'; Received rows '%v'",
				i, test.expectedRows, results.UpdateCompRole.Return.rowsAffected)
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompBulkRolePatch(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedIds  []string
		expectedRole string
		expectedResp []byte
	}{{
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkRole",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"role": "compute"}`),
		nil,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		"compute",
		nil,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkRole",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"],"role": "foo"}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{"x0c0s27b0n0", "x0c0s25b0n0"},
		"foo",
		ErrHMSDSArgNoMatchJSON,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkRole",
		json.RawMessage(`{"role": "service"}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"",
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing IDs","status":400}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkRole",
		json.RawMessage(`{"componentids":["x0c0s27b0n0","x0c0s25b0n0"]}`),
		hmsds.ErrHMSDSArgNoMatch,
		[]string{},
		"",
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing Role","status":400}
`),
	}}

	for i, test := range tests {
		// Clear the results
		results.BulkUpdateCompRole.Input.ids = []string{}
		results.BulkUpdateCompRole.Input.role = ""
		results.BulkUpdateCompRole.Return.affectedIds = test.expectedIds
		results.BulkUpdateCompRole.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !compareIDs(test.expectedIds, results.BulkUpdateCompRole.Input.ids) {
			t.Errorf("Test %v Failed: Expected ids '%v'; Received ids '%v'", i, test.expectedIds, results.BulkUpdateCompRole.Input.ids)
		}

		if test.expectedRole != results.BulkUpdateCompRole.Input.role {
			t.Errorf("Test %v Failed: Expected role '%v'; Received role '%v'", i, test.expectedRole, results.BulkUpdateCompRole.Input.role)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompNIDPatch(t *testing.T) {
	testComp1 := base.Component{
		ID:  "x0c0s25b0n0",
		NID: "864",
	}

	testComp2 := base.Component{
		ID:  "x0c0s25b0n0",
		NID: "800",
	}

	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedComp *base.Component
		expectedResp []byte
	}{{
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s25b0n0/NID",
		reqBody:      json.RawMessage(`{"NID": 864}`),
		hmsdsRespErr: nil,
		expectedComp: &testComp1,
		expectedResp: nil,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s25b0n0/NID",
		reqBody:      json.RawMessage(`{"NID": 800}`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedComp: &testComp2,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"a required argument did not match any valid input","status":400}` + "\n"),
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/State/Components/x0c0s25b0n0/NID",
		reqBody:      json.RawMessage(`{"NID": "foo"}`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedComp: &base.Component{},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"error decoding JSON json: cannot unmarshal string into Go struct field compPatchIn.NID of type int64","status":500}` + "\n"),
	}}

	for i, test := range tests {
		// Clear the results
		results.UpdateCompNID.Input.c = &base.Component{}
		results.UpdateCompNID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if *test.expectedComp != *results.UpdateCompNID.Input.c {
			t.Errorf("Test %v Failed: Expected component '%v'; Received component '%v'", i, test.expectedComp, results.UpdateCompNID.Input.c)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompBulkNIDPatch(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsRespErr  error
		expectedComps *[]base.Component
		expectedResp  []byte
	}{{
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkNID",
		json.RawMessage(`{"components":[{"ID":"x0c0s25b0n0","NID":"800"},{"ID":"x0c0s27b0n0","NID":"864"}]}`),
		nil,
		&[]base.Component{
			base.Component{"x0c0s25b0n0", "", "", "", nil, "", "", "", "800", "", "", "", "", false, false},
			base.Component{"x0c0s27b0n0", "", "", "", nil, "", "", "", "864", "", "", "", "", false, false},
		},
		nil,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/State/Components/BulkNID",
		json.RawMessage(`{"components":[]}`),
		hmsds.ErrHMSDSArgNoMatch,
		&[]base.Component{},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing Components","status":400}
`),
	}}

	for i, test := range tests {
		// Clear the results
		results.BulkUpdateCompNID.Input.comps = &[]base.Component{}
		results.BulkUpdateCompNID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusNoContent {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !compareComps(test.expectedComps, results.BulkUpdateCompNID.Input.comps) {
			t.Errorf("Test %v Failed: Expected ids '%v'; Received ids '%v'", i, test.expectedComps, results.BulkUpdateCompNID.Input.comps)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestGetSCNSubscriptionsAll(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsSubs    *sm.SCNSubscriptionArray
		hmsdsErr     error
		expectedResp []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		&sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		nil,
		json.RawMessage(`{"SubscriptionList":[{"ID":2,"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"},{"ID":3,"Subscriber":"hmfd@sms02","States":["Off","Ready"],"Url":"https://foo2/bar"}]}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		nil,
		hmsds.ErrHMSDSPtrClosed,
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		&sm.SCNSubscriptionArray{[]sm.SCNSubscription{}},
		nil,
		json.RawMessage(`{"SubscriptionList":[]}
`),
	}}

	for i, test := range tests {
		results.GetSCNSubscriptionsAll.Return.subs = test.hmsdsSubs
		results.GetSCNSubscriptionsAll.Return.err = test.hmsdsErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPostSCNSubscription(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		reqBody        []byte
		subs           sm.SCNSubscriptionArray
		subMap         SCNSubMap
		hmsdsRespId    int64
		hmsdsRespErr   error
		expectedSub    sm.SCNPostSubscription
		expectedSubs   sm.SCNSubscriptionArray
		expectedSubMap SCNSubMap
		expectedResp   []byte
	}{{
		"POST",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		json.RawMessage(`{"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}`),
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		2,
		nil,
		sm.SCNPostSubscription{
			Subscriber: "hmfd@sms01",
			States:     []string{"On", "Off"},
			Url:        "https://foo/bar",
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"on":  []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"ID":2,"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		json.RawMessage(`{"Subscriber":"hmfd@sms02","States":["Off","Ready"],"Url":"https://foo2/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"on":  []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		3,
		nil,
		sm.SCNPostSubscription{
			Subscriber: "hmfd@sms02",
			States:     []string{"Off", "Ready"},
			Url:        "https://foo2/bar",
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"ID":3,"Subscriber":"hmfd@sms02","States":["Off","Ready"],"Url":"https://foo2/bar"}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		json.RawMessage(`{"Subscriber":"hmfd@sms03","Url":"https://foo3/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		0,
		nil,
		sm.SCNPostSubscription{},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Missing trigger. Must subscribe to atleast one Enabled, Role, SubRole, SoftwareStatus, or State trigger.","status":400}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		json.RawMessage(`{"Subscriber":"hmfd@sms03","States":["On","foo"],"Url":"https://foo3/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		0,
		nil,
		sm.SCNPostSubscription{},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid state 'foo'","status":400}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		json.RawMessage(`{"Subscriber":"hmfd@sms02","States":["On","Ready"],"Url":"https://foo2/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		0,
		hmsds.ErrHMSDSDuplicateKey,
		sm.SCNPostSubscription{
			Subscriber: "hmfd@sms02",
			States:     []string{"On", "Ready"},
			Url:        "https://foo2/bar",
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Subscribe failed","status":400}
`),
	}, {
		"POST",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		json.RawMessage(`{"Subscriber":"hmfd@sms03","States":["On","Off"],"Url":"https://foo3/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		0,
		hmsds.ErrHMSDSPtrClosed,
		sm.SCNPostSubscription{
			Subscriber: "hmfd@sms03",
			States:     []string{"On", "Off"},
			Url:        "https://foo3/bar",
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Subscribe failed","status":400}
`),
	}}

	for i, test := range tests {
		s.scnSubs = test.subs
		s.scnSubMap = test.subMap
		results.InsertSCNSubscription.Input.sub = sm.SCNPostSubscription{}
		results.InsertSCNSubscription.Return.id = test.hmsdsRespId
		results.InsertSCNSubscription.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		match := true
		if test.expectedSub.Subscriber == results.InsertSCNSubscription.Input.sub.Subscriber &&
			test.expectedSub.Url == results.InsertSCNSubscription.Input.sub.Url {
			for i, state := range test.expectedSub.States {
				if state != results.InsertSCNSubscription.Input.sub.States[i] {
					match = false
				}
			}
		}

		if !match {
			t.Errorf("Test %v Failed: Expected sub '%v'; Received sub '%v'", i, test.expectedSub, results.InsertSCNSubscription.Input.sub)
		}

		if !compareSCNSubs(test.expectedSubs, s.scnSubs) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubs, s.scnSubs)
		}

		if !compareSCNSubMaps(test.expectedSubMap, s.scnSubMap) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubMap, s.scnSubMap)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoDeleteSCNSubscriptionsAll(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		subs           sm.SCNSubscriptionArray
		subMap         SCNSubMap
		hmsdsNumDelete int64
		hmsdsErr       error
		expectedSubs   sm.SCNSubscriptionArray
		expectedSubMap SCNSubMap
		expectedResp   []byte
	}{{
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		0,
		nil,
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		json.RawMessage(`{"code":0,"message":"0 Subscriptions deleted"}
`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"on":  []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		1,
		nil,
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		json.RawMessage(`{"code":0,"message":"1 Subscriptions deleted"}
`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		2,
		nil,
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		json.RawMessage(`{"code":0,"message":"2 Subscriptions deleted"}
`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN",
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		0,
		hmsds.ErrHMSDSPtrClosed,
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Unsubscribe failed","status":400}
`),
	}}

	for i, test := range tests {
		s.scnSubs = test.subs
		s.scnSubMap = test.subMap
		results.DeleteSCNSubscriptionsAll.Return.numDelete = test.hmsdsNumDelete
		results.DeleteSCNSubscriptionsAll.Return.err = test.hmsdsErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if !compareSCNSubs(test.expectedSubs, s.scnSubs) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubs, s.scnSubs)
		}

		if !compareSCNSubMaps(test.expectedSubMap, s.scnSubMap) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubMap, s.scnSubMap)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestGetSCNSubscription(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsSub     *sm.SCNSubscription
		hmsdsErr     error
		expectedId   int64
		expectedResp []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		&sm.SCNSubscription{
			ID:         2,
			Subscriber: "hmfd@sms01",
			States:     []string{"On", "Off"},
			Url:        "https://foo/bar",
		},
		nil,
		2,
		json.RawMessage(`{"ID":2,"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		nil,
		hmsds.ErrHMSDSPtrClosed,
		2,
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		nil,
		nil,
		2,
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"Subscription not found","status":404}
`),
	}}

	for i, test := range tests {
		results.GetSCNSubscription.Return.sub = test.hmsdsSub
		results.GetSCNSubscription.Return.err = test.hmsdsErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if results.GetSCNSubscription.Input.id != test.expectedId {
			t.Errorf("Test %v Failed: Expected id is '%v'; Received '%v'", i, test.expectedId, results.GetSCNSubscription.Input.id)
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPutSCNSubscription(t *testing.T) {
	tests := []struct {
		reqType            string
		reqURI             string
		reqBody            []byte
		subs               sm.SCNSubscriptionArray
		subMap             SCNSubMap
		hmsdsRespDidUpdate bool
		hmsdsRespErr       error
		expectedId         int64
		expectedSub        sm.SCNPostSubscription
		expectedSubs       sm.SCNSubscriptionArray
		expectedSubMap     SCNSubMap
		expectedResp       []byte
	}{{
		"PUT",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}`),
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		false,
		nil,
		2,
		sm.SCNPostSubscription{
			Subscriber: "hmfd@sms01",
			States:     []string{"On", "Off"},
			Url:        "https://foo/bar",
		},
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"Subscription not found","status":404}
`),
	}, {
		"PUT",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Subscriber":"hmfd@sms01","States":["Off","Ready"],"Url":"https://foo/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"on":  []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		true,
		nil,
		2,
		sm.SCNPostSubscription{
			Subscriber: "hmfd@sms01",
			States:     []string{"Off", "Ready"},
			Url:        "https://foo/bar",
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		nil,
	}, {
		"PUT",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Subscriber":"hmfd@sms01","States":["On","foo"],"Url":"https://foo/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		false,
		nil,
		0,
		sm.SCNPostSubscription{},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid state 'foo'","status":400}
`),
	}, {
		"PUT",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Subscriber":"hmfd@sms02","States":["On","Ready"],"Url":"https://foo2/bar"}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		false,
		hmsds.ErrHMSDSPtrClosed,
		2,
		sm.SCNPostSubscription{
			Subscriber: "hmfd@sms02",
			States:     []string{"On", "Ready"},
			Url:        "https://foo2/bar",
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Subscription update failed","status":400}
`),
	}}

	for i, test := range tests {
		s.scnSubs = test.subs
		s.scnSubMap = test.subMap
		results.UpdateSCNSubscription.Input.id = 0
		results.UpdateSCNSubscription.Input.sub = sm.SCNPostSubscription{}
		results.UpdateSCNSubscription.Return.didUpdate = test.hmsdsRespDidUpdate
		results.UpdateSCNSubscription.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		match := true
		if test.expectedSub.Subscriber == results.UpdateSCNSubscription.Input.sub.Subscriber &&
			test.expectedSub.Url == results.UpdateSCNSubscription.Input.sub.Url {
			for i, state := range test.expectedSub.States {
				if state != results.UpdateSCNSubscription.Input.sub.States[i] {
					match = false
				}
			}
		}

		if !match {
			t.Errorf("Test %v Failed: Expected sub '%v'; Received sub '%v'", i, test.expectedSub, results.UpdateSCNSubscription.Input.sub)
		}

		if results.UpdateSCNSubscription.Input.id != test.expectedId {
			t.Errorf("Test %v Failed: Expected id '%v'; Received id '%v'", i, test.expectedId, results.UpdateSCNSubscription.Input.id)
		}

		if !compareSCNSubs(test.expectedSubs, s.scnSubs) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubs, s.scnSubs)
		}

		if !compareSCNSubMaps(test.expectedSubMap, s.scnSubMap) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubMap, s.scnSubMap)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPatchSCNSubscription(t *testing.T) {
	tests := []struct {
		reqType           string
		reqURI            string
		reqBody           []byte
		subs              sm.SCNSubscriptionArray
		subMap            SCNSubMap
		hmsdsRespDidPatch bool
		hmsdsRespErr      error
		expectedId        int64
		expectedOp        string
		expectedPatch     sm.SCNPatchSubscription
		expectedSubs      sm.SCNSubscriptionArray
		expectedSubMap    SCNSubMap
		expectedResp      []byte
	}{{
		"PATCH",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Op":"add","States":["On","Off"]}`),
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		false,
		nil,
		2,
		"add",
		sm.SCNPatchSubscription{
			Op:     "add",
			States: []string{"On", "Off"},
		},
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"Subscription not found","status":404}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Op":"add","States":["Off","Ready"]}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"on":  []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		true,
		nil,
		2,
		"add",
		sm.SCNPatchSubscription{
			Op:     "add",
			States: []string{"Off", "Ready"},
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off", "Ready"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		nil,
	}, {
		"PATCH",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Op":"add","States":["On","foo"]}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		false,
		nil,
		0,
		"",
		sm.SCNPatchSubscription{},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid state 'foo'","status":400}
`),
	}, {
		"PATCH",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		json.RawMessage(`{"Op":"add","States":["On","Ready"]}`),
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		false,
		hmsds.ErrHMSDSPtrClosed,
		2,
		"add",
		sm.SCNPatchSubscription{
			Op:     "add",
			States: []string{"On", "Ready"},
		},
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Subscription patch failed","status":400}
`),
	}}

	for i, test := range tests {
		s.scnSubs = test.subs
		s.scnSubMap = test.subMap
		results.PatchSCNSubscription.Input.id = 0
		results.PatchSCNSubscription.Input.op = ""
		results.PatchSCNSubscription.Input.patch = sm.SCNPatchSubscription{}
		results.PatchSCNSubscription.Return.didPatch = test.hmsdsRespDidPatch
		results.PatchSCNSubscription.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		match := true
		if test.expectedPatch.Op == results.PatchSCNSubscription.Input.patch.Op {
			for i, state := range test.expectedPatch.States {
				if state != results.PatchSCNSubscription.Input.patch.States[i] {
					match = false
				}
			}
		}

		if !match {
			t.Errorf("Test %v Failed: Expected patch '%v'; Received patch '%v'", i, test.expectedPatch, results.PatchSCNSubscription.Input.patch)
		}

		if results.PatchSCNSubscription.Input.id != test.expectedId {
			t.Errorf("Test %v Failed: Expected id '%v'; Received id '%v'", i, test.expectedId, results.PatchSCNSubscription.Input.id)
		}

		if results.PatchSCNSubscription.Input.op != test.expectedOp {
			t.Errorf("Test %v Failed: Expected op '%v'; Received op '%v'", i, test.expectedOp, results.PatchSCNSubscription.Input.op)
		}

		if !compareSCNSubs(test.expectedSubs, s.scnSubs) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubs, s.scnSubs)
		}

		if !compareSCNSubMaps(test.expectedSubMap, s.scnSubMap) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubMap, s.scnSubMap)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoDeleteSCNSubscription(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		subs           sm.SCNSubscriptionArray
		subMap         SCNSubMap
		hmsdsDidDelete bool
		hmsdsErr       error
		expectedId     int64
		expectedSubs   sm.SCNSubscriptionArray
		expectedSubMap SCNSubMap
		expectedResp   []byte
	}{{
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		false,
		nil,
		2,
		sm.SCNSubscriptionArray{},
		SCNSubMap{},
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"Subscription not found","status":404}
`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off": []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"on":  []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
			},
		},
		true,
		nil,
		2,
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off": []SCNUrl{},
				"on":  []SCNUrl{},
			},
		},
		json.RawMessage(`{"code":0,"message":"Subscription deleted"}
`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		true,
		nil,
		2,
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"code":0,"message":"Subscription deleted"}
`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Subscriptions/SCN/2",
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		false,
		hmsds.ErrHMSDSPtrClosed,
		2,
		sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "hmfd@sms02",
				States:     []string{"Off", "Ready"},
				Url:        "https://foo2/bar",
			},
		}},
		SCNSubMap{
			SCNMAP_STATE: map[string][]SCNUrl{
				"off":   []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}, SCNUrl{url: "https://foo2/bar", refCount: 1}},
				"on":    []SCNUrl{SCNUrl{url: "https://foo/bar", refCount: 1}},
				"ready": []SCNUrl{SCNUrl{url: "https://foo2/bar", refCount: 1}},
			},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Unsubscribe failed","status":400}
`),
	}}

	for i, test := range tests {
		s.scnSubs = test.subs
		s.scnSubMap = test.subMap
		results.DeleteSCNSubscription.Input.id = 0
		results.DeleteSCNSubscription.Return.didDelete = test.hmsdsDidDelete
		results.DeleteSCNSubscription.Return.err = test.hmsdsErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if results.DeleteSCNSubscription.Input.id != test.expectedId {
			t.Errorf("Test %v Failed: Expected id '%v'; Received id '%v'", i, test.expectedId, results.DeleteSCNSubscription.Input.id)
		}

		if !compareSCNSubs(test.expectedSubs, s.scnSubs) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubs, s.scnSubs)
		}

		if !compareSCNSubMaps(test.expectedSubMap, s.scnSubMap) {
			t.Errorf("Test %v Failed: Expected subs '%v'; Received subs '%v'", i, test.expectedSubMap, s.scnSubMap)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

//////////////////////////////////////////////////////////////////////////////
// HW Inventory
//////////////////////////////////////////////////////////////////////////////

func TestDoHWInvByLocationGet(t *testing.T) {
	xname := stest.NodeHWInvByLoc1.ID
	payload1, _ := json.Marshal(&stest.NodeHWInvByLoc1)

	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespID  *sm.HWInvByLoc
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/Inventory/Hardware/" + xname,
		&stest.NodeHWInvByLoc1,
		nil,
		xname,
		payload1,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/Hardware/" + xname,
		nil,
		nil,
		xname,
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/Hardware/x0c0s27",
		nil,
		errors.New("unexpected error"), // non-HMS error
		"x0c0s27",
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}
`),
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/Hardware/x0c0s27",
		nil,
		hmsds.ErrHMSDSArgMissing, // HMSError
		"x0c0s27",
		jsonErrHMSDSArgMissing,
	}}

	for i, test := range tests {
		results.GetHWInvByLocID.Return.entry = test.hmsdsRespID
		results.GetHWInvByLocID.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespID != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespID == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetHWInvByLocID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'",
				i, test.expectedID, results.GetHWInvByLocID.Input.id)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvByFRUGet(t *testing.T) {
	fruid := stest.NodeHWInvByFRU1.FRUID
	payload1, _ := json.Marshal(&stest.NodeHWInvByFRU1)

	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespID  *sm.HWInvByFRU
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU/" + fruid,
		&stest.NodeHWInvByFRU1,
		nil,
		fruid,
		payload1,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU/" + fruid,
		nil,
		hmsds.ErrHMSDSArgMissing, //HMS error
		fruid,
		jsonErrHMSDSArgMissing,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU/" + fruid,
		nil,
		errors.New("unexpected error"), // non-HMSError
		fruid,
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}
`),
	}}

	for i, test := range tests {
		results.GetHWInvByFRUID.Return.entry = test.hmsdsRespID
		results.GetHWInvByFRUID.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespID != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespID == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetHWInvByFRUID.Input.fruid {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'",
				i, test.expectedID, results.GetHWInvByFRUID.Input.fruid)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvByLocationGetAll(t *testing.T) {
	payload1, _ := json.Marshal(stest.HWInvByLocArray1)

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespIDs   []*sm.HWInvByLoc
		hmsdsRespErr   error
		expectedFilter *hmsds.HWInvLocFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware",
		hmsdsRespIDs:   stest.HWInvByLocArray1,
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvLocFilter{},
		expectedResp:   payload1,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware",
		hmsdsRespIDs:   nil,
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing, // actual error here doesn't really matter- any error.
		expectedFilter: &hmsds.HWInvLocFilter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetHWInvByLocFilter.Input.f = nil
		results.GetHWInvByLocFilter.Return.hwlocs = test.hmsdsRespIDs
		results.GetHWInvByLocFilter.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil &&
			len(test.hmsdsRespIDs) == len(stest.HWInvByLocArray1) &&
			w.Code != http.StatusOK {

			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespIDs == nil) &&
			w.Code == http.StatusOK {

			t.Errorf("Response code was %v; expected an error", w.Code)
		} else {
			if !compareHWInvLocFilter(*test.expectedFilter, *results.GetHWInvByLocFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetHWInvByLocFilter.Input.f)
			}
			if strings.TrimSpace(string(test.expectedResp)) !=
				strings.TrimSpace(string(w.Body.Bytes())) {

				t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
					i, string(test.expectedResp), w.Body)
			}
		}
	}
}

func TestDoHWInvByFRUGetAll(t *testing.T) {
	payload1, _ := json.Marshal(stest.HWInvByFRUArray1)

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespIDs   []*sm.HWInvByFRU
		hmsdsRespErr   error
		expectedFilter *hmsds.HWInvLocFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU",
		hmsdsRespIDs:   stest.HWInvByFRUArray1,
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvLocFilter{},
		expectedResp:   payload1,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU",
		hmsdsRespIDs:   nil,
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing, // actual error here doesn't really matter- any error.
		expectedFilter: &hmsds.HWInvLocFilter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetHWInvByFRUFilter.Input.f = nil
		results.GetHWInvByFRUFilter.Return.hwfrus = test.hmsdsRespIDs
		results.GetHWInvByFRUFilter.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil &&
			len(test.hmsdsRespIDs) == len(stest.HWInvByFRUArray1) &&
			w.Code != http.StatusOK {

			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespIDs == nil) &&
			w.Code == http.StatusOK {

			t.Errorf("Response code was %v; expected an error", w.Code)
		} else {
			if !compareHWInvLocFilter(*test.expectedFilter, *results.GetHWInvByFRUFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetHWInvByFRUFilter.Input.f)
			}
			if strings.TrimSpace(string(test.expectedResp)) !=
				strings.TrimSpace(string(w.Body.Bytes())) {

				t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
					i, string(test.expectedResp), w.Body)
			}
		}
	}
}

func TestDoHWInvByLocationPost(t *testing.T) {
	var HWInvByLocArray1 = []sm.HWInvByLoc{
		stest.NodeHWInvByLoc1,
	}
	var hwIn1 HwInvIn
	hwIn1.Hardware = HWInvByLocArray1
	hwLocs1, _ := sm.NewHWInvByLocs(HWInvByLocArray1)
	payload1, _ := json.Marshal(hwIn1)

	var HWInvByLocArray2 = []sm.HWInvByLoc{
		stest.NodeHWInvByLoc1,
		stest.ProcHWInvByLoc1,
		stest.ProcHWInvByLoc2,
		stest.MemHWInvByLoc1,
		stest.MemHWInvByLoc2,
	}
	var hwIn2 HwInvIn
	hwIn2.Hardware = HWInvByLocArray2
	hwLocs2, _ := sm.NewHWInvByLocs(HWInvByLocArray2)
	payload2, _ := json.Marshal(hwIn2)

	tests := []struct {
		reqType             string
		reqURI              string
		reqBody             []byte
		expectedHWInvByLocs []*sm.HWInvByLoc
		hmsdsRespErr        error
		expectedResp        []byte
	}{{
		reqType:             "POST",
		reqURI:              "https://localhost/hsm/v2/Inventory/Hardware",
		reqBody:             payload1,
		expectedHWInvByLocs: hwLocs1,
		hmsdsRespErr:        nil,
		expectedResp:        json.RawMessage(`{"code":0,"message":"Created 1 entries"}` + "\n"),
	}, {
		reqType:             "POST",
		reqURI:              "https://localhost/hsm/v2/Inventory/Hardware",
		reqBody:             payload2,
		expectedHWInvByLocs: hwLocs2,
		hmsdsRespErr:        nil,
		expectedResp:        json.RawMessage(`{"code":0,"message":"Created 5 entries"}` + "\n"),
	}, {
		reqType:             "POST",
		reqURI:              "https://localhost/hsm/v2/Inventory/Hardware",
		reqBody:             json.RawMessage(`{}`),
		expectedHWInvByLocs: hwLocs2,
		hmsdsRespErr:        nil,
		expectedResp:        json.RawMessage(`{"code":0,"message":"Created 0 entries"}` + "\n"),
	}, {
		reqType:             "POST",
		reqURI:              "https://localhost/hsm/v2/Inventory/Hardware",
		reqBody:             payload2,
		expectedHWInvByLocs: hwLocs2,
		hmsdsRespErr:        errors.New("Unknown Error"),
		expectedResp:        json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"operation 'POST' failed during store.","status":500}` + "\n"),
	}, {
		reqType:             "POST",
		reqURI:              "https://localhost/hsm/v2/Inventory/Hardware",
		reqBody:             payload2,
		expectedHWInvByLocs: hwLocs2,
		hmsdsRespErr:        hmsds.ErrHMSDSArgBadID,
		expectedResp:        json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}` + "\n"),
	}, {
		reqType:             "POST",
		reqURI:              "https://localhost/hsm/v2/Inventory/Hardware",
		reqBody:             payload2,
		expectedHWInvByLocs: hwLocs2,
		hmsdsRespErr:        hmsds.ErrHMSDSDuplicateKey,
		expectedResp:        json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Would create a duplicate key or non-unique field","status":400}` + "\n"),
	}}

	for i, test := range tests {
		if i == 0 {
			results.InsertHWInvByLocs.Input.hls = hwLocs1
		} else {
			results.InsertHWInvByLocs.Input.hls = hwLocs2
		}
		results.InsertHWInvByLocs.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.expectedHWInvByLocs != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.expectedHWInvByLocs == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvByLocationDelete(t *testing.T) {
	data := Response{0, "deleted 1 entry"}
	payload1, _ := json.Marshal(data)

	tests := []struct {
		reqType            string
		reqURI             string
		hmsdsRespDidDelete bool
		hmsdsRespErr       error
		expectedID         string
		expectedResp       []byte
	}{{
		"DELETE",
		"https://localhost/hsm/v2/Inventory/Hardware/x0c0s0b0n0",
		true,
		nil,
		"x0c0s0b0n0",
		payload1,
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/Hardware/x0c0s0b0n0",
		false,
		nil,
		"x0c0s0b0n0",
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/Hardware/x0c0s0b0n0",
		false,
		errors.New("unexpected error"), // actual error here doesn't really matter- any non-HMS error.
		"x0c0s0b0n0",
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/Hardware/x0c0s0b0n0",
		false,
		hmsds.ErrHMSDSArgMissing, // actual error here doesn't really matter- any HMS error.
		"x0c0s0b0n0",
		jsonErrHMSDSArgMissing,
	}}

	for i, test := range tests {
		results.DeleteHWInvByLocID.Return.changed = test.hmsdsRespDidDelete
		results.DeleteHWInvByLocID.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil &&
			test.hmsdsRespDidDelete == true &&
			w.Code != http.StatusOK {

			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespDidDelete == false) &&
			w.Code == http.StatusOK {

			t.Errorf("Response code was %v; expected an error", w.Code)
		}
		if test.expectedID != results.DeleteHWInvByLocID.Input.id {
			t.Errorf("Test %v Failed: Expected xname '%v'; Received xname '%v'",
				i, test.expectedID, results.DeleteHWInvByLocID.Input.id)
		}
		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvByFRUDelete(t *testing.T) {
	data := Response{0, "deleted 1 entry"}
	payload1, _ := json.Marshal(data)

	tests := []struct {
		reqType            string
		reqURI             string
		hmsdsRespDidDelete bool
		hmsdsRespErr       error
		expectedID         string
		expectedResp       []byte
	}{{
		"DELETE",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU/some_fruid",
		true,
		nil,
		"some_fruid",
		payload1,
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU/some_fruid",
		false,
		nil,
		"some_fruid",
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such FRU ID.","status":404}`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU/some_fruid",
		false,
		errors.New("unexpected DB error"), // actual error here doesn't really matter- any non-HMS error.
		"some_fruid",
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}
`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU/some_fruid",
		false,
		hmsds.ErrHMSDSArgMissing, // actual error here doesn't really matter- any HMS error.
		"some_fruid",
		jsonErrHMSDSArgMissing,
	}}

	for i, test := range tests {
		results.DeleteHWInvByFRUID.Return.changed = test.hmsdsRespDidDelete
		results.DeleteHWInvByFRUID.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil &&
			test.hmsdsRespDidDelete == true &&
			w.Code != http.StatusOK {

			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespDidDelete == false) &&
			w.Code == http.StatusOK {

			t.Errorf("Response code was %v; expected an error", w.Code)
		}
		if test.expectedID != results.DeleteHWInvByFRUID.Input.fruid {
			t.Errorf("Test %v Failed: Expected fruid '%v'; Received fruid '%v'",
				i, test.expectedID, results.DeleteHWInvByFRUID.Input.fruid)
		}
		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvByLocationDeleteAll(t *testing.T) {
	data := Response{0, "deleted 2 entries"}
	payload1, _ := json.Marshal(data)

	tests := []struct {
		reqType          string
		reqURI           string
		hmsdsRespNumRows int64
		hmsdsRespErr     error
		expectedResp     []byte
	}{{
		"DELETE",
		"https://localhost/hsm/v2/Inventory/Hardware",
		2,
		nil,
		payload1,
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/Hardware",
		0,
		nil,
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/Hardware",
		0,
		hmsds.ErrHMSDSArgMissing, // actual error here doesn't really matter- any error.
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}`),
	}}

	for i, test := range tests {
		results.DeleteHWInvByLocsAll.Return.numRows = test.hmsdsRespNumRows
		results.DeleteHWInvByLocsAll.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespNumRows > 0 &&
			w.Code != http.StatusOK {

			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespNumRows == 0) &&
			w.Code == http.StatusOK {

			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvByFRUDeleteAll(t *testing.T) {
	data := Response{0, "deleted 2 entries"}
	payload1, _ := json.Marshal(data)

	tests := []struct {
		reqType          string
		reqURI           string
		hmsdsRespNumRows int64
		hmsdsRespErr     error
		expectedResp     []byte
	}{{
		"DELETE",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU",
		2,
		nil,
		payload1,
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU",
		0,
		nil,
		json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}`),
	}, {
		"DELETE",
		"https://localhost/hsm/v2/Inventory/HardwareByFRU",
		0,
		hmsds.ErrHMSDSArgMissing, // actual error here doesn't really matter- any error.
		json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}`),
	}}

	for i, test := range tests {
		results.DeleteHWInvByFRUsAll.Return.numRows = test.hmsdsRespNumRows
		results.DeleteHWInvByFRUsAll.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespNumRows > 0 &&
			w.Code != http.StatusOK {

			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespNumRows == 0) &&
			w.Code == http.StatusOK {

			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvByLocationQueryGet(t *testing.T) {
	payload1, _ := json.Marshal(stest.HWInvByLocQuery)
	payload2, _ := json.Marshal(stest.HWInvByLocQueryFF)
	payload3, _ := json.Marshal(stest.HWInvByLocQueryProc)

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespIDs   []*sm.HWInvByLoc
		hmsdsRespErr   error
		expectedFilter *hmsds.HWInvLocFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/Query/s0",
		hmsdsRespIDs:   stest.HWInvByLocArray1,
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvLocFilter{Children: true},
		expectedResp:   payload1,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/Query/s0?format=fullyflat",
		hmsdsRespIDs:   stest.HWInvByLocArray1,
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvLocFilter{Children: true},
		expectedResp:   payload2,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/Hardware/Query/x0c0s0b0n0?type=processor",
		hmsdsRespIDs: stest.HWInvByLocQueryProcArray,
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.HWInvLocFilter{
			ID:       []string{"x0c0s0b0n0"},
			Type:     []string{base.Processor.String()},
			Children: true,
		},
		expectedResp: payload3,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/Query/s0",
		hmsdsRespIDs:   nil,
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing, // actual error here doesn't really matter- any error.
		expectedFilter: &hmsds.HWInvLocFilter{Children: true},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetHWInvByLocQueryFilter.Input.f = nil
		results.GetHWInvByLocQueryFilter.Return.hwlocs = test.hmsdsRespIDs
		results.GetHWInvByLocQueryFilter.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil &&
			len(test.hmsdsRespIDs) == len(stest.HWInvByLocArray1) &&
			w.Code != http.StatusOK {

			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespIDs == nil) &&
			w.Code == http.StatusOK {

			t.Errorf("Response code was %v; expected an error", w.Code)
		} else {
			if !compareHWInvLocFilter(*test.expectedFilter, *results.GetHWInvByLocQueryFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetHWInvByLocQueryFilter.Input.f)
			}
			if strings.TrimSpace(string(test.expectedResp)) !=
				strings.TrimSpace(string(w.Body.Bytes())) {

				t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
					i, string(test.expectedResp), w.Body)
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////
// HW Inventory History
//////////////////////////////////////////////////////////////////////////////

func TestDoHWInvHistByLocationGet(t *testing.T) {
	testHWInvHist1 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}
	testHWInvHist3 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:00:00",
		EventType: "Removed",
	}
	testHWInvHist5 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_2",
		Timestamp: "2020-01-21 12:10:00",
		EventType: "Added",
	}

	payload1, _ := json.Marshal(sm.HWInvHistArray{
		ID:      testHWInvHist1.ID,
		History: []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist3, &testHWInvHist5},
	})
	payload2, _ := json.Marshal(sm.HWInvHistArray{
		ID:      testHWInvHist1.ID,
		History: []*sm.HWInvHist{&testHWInvHist1},
	})

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.HWInvHist
		hmsdsRespErr   error
		expectedFilter *hmsds.HWInvHistFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/History/" + testHWInvHist1.ID,
		hmsdsResp:      []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist3, &testHWInvHist5},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvHistFilter{ID: []string{testHWInvHist1.ID}},
		expectedResp:   payload1,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/Hardware/History/" + testHWInvHist1.ID + "?starttime=2020-01-21T12:01:00Z&endtime=2020-01-21T11:00:00Z&eventtype=added",
		hmsdsResp:    []*sm.HWInvHist{&testHWInvHist1},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.HWInvHistFilter{
			ID:        []string{testHWInvHist1.ID},
			EventType: []string{"Added"},
			StartTime: "2020-01-21T12:01:00Z",
			EndTime:   "2020-01-21T11:00:00Z",
		},
		expectedResp: payload2,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/History/" + testHWInvHist1.ID + "?eventtype=foo",
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: nil,
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid HWInvHist event type","status":400}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/History/" + testHWInvHist1.ID,
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: &hmsds.HWInvHistFilter{ID: []string{testHWInvHist1.ID}},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetHWInvHistFilter.Input.f = nil
		results.GetHWInvHistFilter.Return.hwhists = test.hmsdsResp
		results.GetHWInvHistFilter.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		} else {
			if !compareHWInvHistFilter(test.expectedFilter, results.GetHWInvHistFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetHWInvHistFilter.Input.f)
			}
			if strings.TrimSpace(string(test.expectedResp)) !=
				strings.TrimSpace(string(w.Body.Bytes())) {
				t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
					i, string(test.expectedResp), w.Body)
			}
		}
	}
}

func TestDoHWInvHistByFRUGet(t *testing.T) {
	testHWInvHist1 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}
	testHWInvHist3 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:00:00",
		EventType: "Removed",
	}
	testHWInvHist6 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p1",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:10:00",
		EventType: "Added",
	}

	payload1, _ := json.Marshal(sm.HWInvHistArray{
		ID:      testHWInvHist1.FruId,
		History: []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist3, &testHWInvHist6},
	})
	payload2, _ := json.Marshal(sm.HWInvHistArray{
		ID:      testHWInvHist1.FruId,
		History: []*sm.HWInvHist{&testHWInvHist1},
	})

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.HWInvHist
		hmsdsRespErr   error
		expectedFilter *hmsds.HWInvHistFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU/History/" + testHWInvHist1.FruId,
		hmsdsResp:      []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist3, &testHWInvHist6},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvHistFilter{FruId: []string{testHWInvHist1.FruId}},
		expectedResp:   payload1,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/HardwareByFRU/History/" + testHWInvHist1.FruId + "?starttime=2020-01-21T12:01:00Z&endtime=2020-01-21T11:00:00Z&eventtype=added",
		hmsdsResp:    []*sm.HWInvHist{&testHWInvHist1},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.HWInvHistFilter{
			FruId:     []string{testHWInvHist1.FruId},
			EventType: []string{"Added"},
			StartTime: "2020-01-21T12:01:00Z",
			EndTime:   "2020-01-21T11:00:00Z",
		},
		expectedResp: payload2,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU/History/" + testHWInvHist1.FruId + "?eventtype=foo",
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: nil,
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid HWInvHist event type","status":400}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU/History/" + testHWInvHist1.FruId,
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: &hmsds.HWInvHistFilter{FruId: []string{testHWInvHist1.FruId}},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetHWInvHistFilter.Input.f = nil
		results.GetHWInvHistFilter.Return.hwhists = test.hmsdsResp
		results.GetHWInvHistFilter.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		} else {
			if !compareHWInvHistFilter(test.expectedFilter, results.GetHWInvHistFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetHWInvHistFilter.Input.f)
			}
			if strings.TrimSpace(string(test.expectedResp)) !=
				strings.TrimSpace(string(w.Body.Bytes())) {
				t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
					i, string(test.expectedResp), w.Body)
			}
		}
	}
}

func TestDoHWInvHistByLocationGetAll(t *testing.T) {
	testHWInvHist1 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}
	testHWInvHist2 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p1",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_2",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}
	testHWInvHist3 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:00:00",
		EventType: "Removed",
	}
	testHWInvHist4 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p1",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_2",
		Timestamp: "2020-01-21 12:00:00",
		EventType: "Removed",
	}
	testHWInvHist5 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_2",
		Timestamp: "2020-01-21 12:10:00",
		EventType: "Added",
	}
	testHWInvHist6 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p1",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:10:00",
		EventType: "Added",
	}

	payload1, _ := json.Marshal(sm.HWInvHistResp{
		Components: []sm.HWInvHistArray{{
			ID:      testHWInvHist1.ID,
			History: []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist3, &testHWInvHist5},
		}, {
			ID:      testHWInvHist2.ID,
			History: []*sm.HWInvHist{&testHWInvHist2, &testHWInvHist4, &testHWInvHist6},
		}},
	})
	payload2, _ := json.Marshal(sm.HWInvHistResp{
		Components: []sm.HWInvHistArray{{
			ID:      testHWInvHist1.ID,
			History: []*sm.HWInvHist{&testHWInvHist1},
		}, {
			ID:      testHWInvHist2.ID,
			History: []*sm.HWInvHist{&testHWInvHist2},
		}},
	})

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.HWInvHist
		hmsdsRespErr   error
		expectedFilter *hmsds.HWInvHistFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/History",
		hmsdsResp:      []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist2, &testHWInvHist3, &testHWInvHist4, &testHWInvHist5, &testHWInvHist6},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvHistFilter{},
		expectedResp:   payload1,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/Hardware/History?id=x5c4s3b2n1p0&id=x5c4s3b2n1p1&fruid=MFR-PARTNUMBER-SERIALNUMBER_1&fruid=MFR-PARTNUMBER-SERIALNUMBER_2&starttime=2020-01-21T12:01:00Z&endtime=2020-01-21T11:00:00Z&eventtype=added",
		hmsdsResp:    []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist2},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.HWInvHistFilter{
			ID:        []string{testHWInvHist1.ID, testHWInvHist2.ID},
			FruId:     []string{testHWInvHist1.FruId, testHWInvHist2.FruId},
			EventType: []string{"Added"},
			StartTime: "2020-01-21T12:01:00Z",
			EndTime:   "2020-01-21T11:00:00Z",
		},
		expectedResp: payload2,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/History?eventtype=foo",
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: nil,
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid HWInvHist event type","status":400}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/Hardware/History",
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: &hmsds.HWInvHistFilter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetHWInvHistFilter.Input.f = nil
		results.GetHWInvHistFilter.Return.hwhists = test.hmsdsResp
		results.GetHWInvHistFilter.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		} else {
			if !compareHWInvHistFilter(test.expectedFilter, results.GetHWInvHistFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetHWInvHistFilter.Input.f)
			}
			if strings.TrimSpace(string(test.expectedResp)) !=
				strings.TrimSpace(string(w.Body.Bytes())) {
				t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
					i, string(test.expectedResp), w.Body)
			}
		}
	}
}

func TestDoHWInvHistByFRUGetAll(t *testing.T) {
	testHWInvHist1 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}
	testHWInvHist2 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p1",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_2",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}
	testHWInvHist3 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:00:00",
		EventType: "Removed",
	}
	testHWInvHist4 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p1",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_2",
		Timestamp: "2020-01-21 12:00:00",
		EventType: "Removed",
	}
	testHWInvHist5 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_2",
		Timestamp: "2020-01-21 12:10:00",
		EventType: "Added",
	}
	testHWInvHist6 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p1",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:10:00",
		EventType: "Added",
	}

	payload1, _ := json.Marshal(sm.HWInvHistResp{
		Components: []sm.HWInvHistArray{{
			ID:      testHWInvHist1.FruId,
			History: []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist3, &testHWInvHist6},
		}, {
			ID:      testHWInvHist2.FruId,
			History: []*sm.HWInvHist{&testHWInvHist2, &testHWInvHist4, &testHWInvHist5},
		}},
	})
	payload2, _ := json.Marshal(sm.HWInvHistResp{
		Components: []sm.HWInvHistArray{{
			ID:      testHWInvHist1.FruId,
			History: []*sm.HWInvHist{&testHWInvHist1},
		}, {
			ID:      testHWInvHist2.FruId,
			History: []*sm.HWInvHist{&testHWInvHist2},
		}},
	})

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.HWInvHist
		hmsdsRespErr   error
		expectedFilter *hmsds.HWInvHistFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU/History",
		hmsdsResp:      []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist2, &testHWInvHist3, &testHWInvHist4, &testHWInvHist5, &testHWInvHist6},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.HWInvHistFilter{},
		expectedResp:   payload1,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/HardwareByFRU/History?id=x5c4s3b2n1p0&id=x5c4s3b2n1p1&fruid=MFR-PARTNUMBER-SERIALNUMBER_1&fruid=MFR-PARTNUMBER-SERIALNUMBER_2&starttime=2020-01-21T12:01:00Z&endtime=2020-01-21T11:00:00Z&eventtype=added",
		hmsdsResp:    []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist2},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.HWInvHistFilter{
			ID:        []string{testHWInvHist1.ID, testHWInvHist2.ID},
			FruId:     []string{testHWInvHist1.FruId, testHWInvHist2.FruId},
			EventType: []string{"Added"},
			StartTime: "2020-01-21T12:01:00Z",
			EndTime:   "2020-01-21T11:00:00Z",
		},
		expectedResp: payload2,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU/History?eventtype=foo",
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: nil,
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid HWInvHist event type","status":400}` + "\n"),
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/HardwareByFRU/History",
		hmsdsResp:      []*sm.HWInvHist{},
		hmsdsRespErr:   hmsds.ErrHMSDSArgMissing,
		expectedFilter: &hmsds.HWInvHistFilter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetHWInvHistFilter.Input.f = nil
		results.GetHWInvHistFilter.Return.hwhists = test.hmsdsResp
		results.GetHWInvHistFilter.Return.err = test.hmsdsRespErr

		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		} else {
			if !compareHWInvHistFilter(test.expectedFilter, results.GetHWInvHistFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetHWInvHistFilter.Input.f)
			}
			if strings.TrimSpace(string(test.expectedResp)) !=
				strings.TrimSpace(string(w.Body.Bytes())) {
				t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
					i, string(test.expectedResp), w.Body)
			}
		}
	}
}

func TestDoHWInvHistByLocationDelete(t *testing.T) {
	type testParams struct {
		reqType          string
		reqURI           string
		hmsdsRespNumRows int64
		hmsdsRespErr     error
		expectedID       string
		expectedResp     []byte
	}

	tests := []testParams{{
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/Hardware/History/x5c4s3b2n1p0",
		hmsdsRespNumRows: 42,
		hmsdsRespErr:     nil,
		expectedID:       "x5c4s3b2n1p0",
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted 42 entries"}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/Hardware/History/x5c4s3b2n1p0",
		hmsdsRespNumRows: 0,
		hmsdsRespErr:     nil,
		expectedID:       "x5c4s3b2n1p0",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/Hardware/History/x5c4s3b2n1p0",
		hmsdsRespNumRows: 0,
		hmsdsRespErr:     errors.New("unexpected DB error"),
		expectedID:       "x5c4s3b2n1p0",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteHWInvHistByLocID.Return.numRows = test.hmsdsRespNumRows
		results.DeleteHWInvHistByLocID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespNumRows != 0 && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if test.expectedID != results.DeleteHWInvHistByLocID.Input.id {
			t.Errorf("Test %v Failed: Expected id is '%v'; Received '%v'", i, test.expectedID, results.DeleteHWInvHistByLocID.Input.id)
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvHistByFRUDelete(t *testing.T) {
	type testParams struct {
		reqType          string
		reqURI           string
		hmsdsRespNumRows int64
		hmsdsRespErr     error
		expectedFruId    string
		expectedResp     []byte
	}

	tests := []testParams{{
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/HardwareByFRU/History/MFR-PARTNUMBER-SERIALNUMBER_1",
		hmsdsRespNumRows: 42,
		hmsdsRespErr:     nil,
		expectedFruId:    "MFR-PARTNUMBER-SERIALNUMBER_1",
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted 42 entries"}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/HardwareByFRU/History/MFR-PARTNUMBER-SERIALNUMBER_1",
		hmsdsRespNumRows: 0,
		hmsdsRespErr:     nil,
		expectedFruId:    "MFR-PARTNUMBER-SERIALNUMBER_1",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/HardwareByFRU/History/MFR-PARTNUMBER-SERIALNUMBER_1",
		hmsdsRespNumRows: 0,
		hmsdsRespErr:     errors.New("unexpected DB error"),
		expectedFruId:    "MFR-PARTNUMBER-SERIALNUMBER_1",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteHWInvHistByFRUID.Return.numRows = test.hmsdsRespNumRows
		results.DeleteHWInvHistByFRUID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespNumRows != 0 && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedFruId != results.DeleteHWInvHistByFRUID.Input.fruid {
			t.Errorf("Test %v Failed: Expected id is '%v'; Received '%v'", i, test.expectedFruId, results.DeleteHWInvHistByFRUID.Input.fruid)
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoHWInvHistDeleteAll(t *testing.T) {
	type testParams struct {
		reqType          string
		reqURI           string
		hmsdsRespNumRows int64
		hmsdsRespErr     error
		expectedResp     []byte
	}

	tests := []testParams{{
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/Hardware/History",
		hmsdsRespNumRows: 42,
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted 42 entries"}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/Hardware/History",
		hmsdsRespNumRows: 0,
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/Hardware/History",
		hmsdsRespNumRows: 0,
		hmsdsRespErr:     errors.New("unexpected DB error"),
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteHWInvHistAll.Return.numRows = test.hmsdsRespNumRows
		results.DeleteHWInvHistAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespNumRows != 0 && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// RedfishEndpoints
//////////////////////////////////////////////////////////////////////////////

func TestDoRedfishEndpointGet(t *testing.T) {
	xname := stest.TestRedfishEndpointNodeBMC1.ID
	payload, _ := json.Marshal(&stest.TestRedfishEndpointNodeBMC1)

	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespEP  *sm.RedfishEndpoint
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/" + xname,
		hmsdsRespEP:  &stest.TestRedfishEndpointNodeBMC1,
		hmsdsRespErr: nil,
		expectedID:   xname,
		expectedResp: payload,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/" + xname,
		hmsdsRespEP:  nil,
		hmsdsRespErr: nil,
		expectedID:   xname,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/" + xname,
		hmsdsRespEP:  nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedID:   xname,
		expectedResp: jsonErrHMSDSArgMissing,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/" + xname,
		hmsdsRespEP:  nil,
		hmsdsRespErr: errors.New("unexpected DB error"),
		expectedID:   xname,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetRFEndpointByID.Return.entry = test.hmsdsRespEP
		results.GetRFEndpointByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an unexpected error '%s' occurred while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespEP != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespEP == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetRFEndpointByID.Input.id {
			t.Errorf("Test %v Failed: Expected endpoint ID is '%v'; Received endpoint ID '%v'", i, test.expectedID, results.GetRFEndpointByID.Input.id)
		}

		if strings.TrimSpace(string(test.expectedResp)) != strings.TrimSpace(string(w.Body.Bytes())) {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoRedfishEndpointsGet(t *testing.T) {
	payload1, _ := json.Marshal(stest.TestRedfishEndpointArray)
	payload2, _ := json.Marshal(stest.TestRedfishEndpointArrayNodeBMCs)
	payload3, _ := json.Marshal(stest.TestRedfishEndpointArrayNodeBMC1)
	payload4, _ := json.Marshal(stest.TestRedfishEndpointArrayNodeBMC2)
	payload5, _ := json.Marshal(stest.TestRedfishEndpointArrayRouterBMCs)
	ssMockLDataRPArray := []sstorage.MockLookup{}
	for _, ep := range stest.TestRedfishEndpointArray.RedfishEndpoints {
		ssDataItem := sstorage.MockLookup{
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{
					Xname:    ep.ID,
					URL:      ep.FQDN + "/redfish/v2",
					Username: ep.User,
					Password: ep.Password,
				},
				Err: nil,
			},
		}
		ssMockLDataRPArray = append(ssMockLDataRPArray, ssDataItem)
	}
	ssMockLDataRPArrayNodeBMCs := []sstorage.MockLookup{}
	for _, ep := range stest.TestRedfishEndpointArrayNodeBMCs.RedfishEndpoints {
		ssDataItem := sstorage.MockLookup{
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{
					Xname:    ep.ID,
					URL:      ep.FQDN + "/redfish/v2",
					Username: ep.User,
					Password: ep.Password,
				},
				Err: nil,
			},
		}
		ssMockLDataRPArrayNodeBMCs = append(ssMockLDataRPArrayNodeBMCs, ssDataItem)
	}
	ssMockLDataRPArrayNodeBMC1 := []sstorage.MockLookup{}
	for _, ep := range stest.TestRedfishEndpointArrayNodeBMC1.RedfishEndpoints {
		ssDataItem := sstorage.MockLookup{
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{
					Xname:    ep.ID,
					URL:      ep.FQDN + "/redfish/v2",
					Username: ep.User,
					Password: ep.Password,
				},
				Err: nil,
			},
		}
		ssMockLDataRPArrayNodeBMC1 = append(ssMockLDataRPArrayNodeBMC1, ssDataItem)
	}
	ssMockLDataRPArrayNodeBMC2 := []sstorage.MockLookup{}
	for _, ep := range stest.TestRedfishEndpointArrayNodeBMC2.RedfishEndpoints {
		ssDataItem := sstorage.MockLookup{
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{
					Xname:    ep.ID,
					URL:      ep.FQDN + "/redfish/v2",
					Username: ep.User,
					Password: ep.Password,
				},
				Err: nil,
			},
		}
		ssMockLDataRPArrayNodeBMC2 = append(ssMockLDataRPArrayNodeBMC2, ssDataItem)
	}
	ssMockLDataRPArrayRouterBMCs := []sstorage.MockLookup{}
	for _, ep := range stest.TestRedfishEndpointArrayRouterBMCs.RedfishEndpoints {
		ssDataItem := sstorage.MockLookup{
			Output: sstorage.OutputLookup{
				Output: &compcreds.CompCredentials{
					Xname:    ep.ID,
					URL:      ep.FQDN + "/redfish/v2",
					Username: ep.User,
					Password: ep.Password,
				},
				Err: nil,
			},
		}
		ssMockLDataRPArrayRouterBMCs = append(ssMockLDataRPArrayRouterBMCs, ssDataItem)
	}

	tests := []struct {
		reqType        string
		reqURI         string
		ssRespLData    []sstorage.MockLookup
		hmsdsRespIDs   []*sm.RedfishEndpoint
		hmsdsRespErr   error
		expectedFilter hmsds.RedfishEPFilter
		expectedResp   []byte
	}{{
		"GET",
		"https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		ssMockLDataRPArray,
		stest.TestRedfishEndpointArray.RedfishEndpoints,
		nil,
		hmsds.RedfishEPFilter{},
		payload1,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/RedfishEndpoints?type=nodeBMC",
		ssMockLDataRPArrayNodeBMCs,
		stest.TestRedfishEndpointArrayNodeBMCs.RedfishEndpoints,
		nil,
		hmsds.RedfishEPFilter{
			Type: []string{"nodeBMC"},
		},
		payload2,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/RedfishEndpoints?fqdn=" +
			stest.TestRedfishEndpointNodeBMC1.FQDN,
		ssMockLDataRPArrayNodeBMC1,
		stest.TestRedfishEndpointArrayNodeBMC1.RedfishEndpoints,
		nil,
		hmsds.RedfishEPFilter{
			FQDN: []string{stest.TestRedfishEndpointNodeBMC1.FQDN},
		},
		payload3,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/RedfishEndpoints?uuid=" +
			stest.TestRedfishEndpointNodeBMC2.UUID,
		ssMockLDataRPArrayNodeBMC2,
		stest.TestRedfishEndpointArrayNodeBMC2.RedfishEndpoints,
		nil,
		hmsds.RedfishEPFilter{
			UUID: []string{stest.TestRedfishEndpointNodeBMC2.UUID},
		},
		payload4,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/RedfishEndpoints?uuid=" +
			stest.TestRedfishEndpointNodeBMC2.UUID + "&type=nodebmc",
		ssMockLDataRPArrayNodeBMC2,
		stest.TestRedfishEndpointArrayNodeBMC2.RedfishEndpoints,
		nil,
		hmsds.RedfishEPFilter{
			Type: []string{"nodebmc"},
			UUID: []string{stest.TestRedfishEndpointNodeBMC2.UUID},
		},
		payload4,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/RedfishEndpoints?type=routerBMC",
		ssMockLDataRPArrayRouterBMCs,
		stest.TestRedfishEndpointArrayRouterBMCs.RedfishEndpoints,
		nil,
		hmsds.RedfishEPFilter{
			Type: []string{"routerBMC"},
		},
		payload5,
	}, {
		"GET",
		"https://localhost/hsm/v2/Inventory/RedfishEndpoints?type=foo",
		[]sstorage.MockLookup{},
		nil,
		hmsds.ErrHMSDSArgBadType,
		hmsds.RedfishEPFilter{
			Type: []string{"foo"},
		},
		json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not a valid HMS Type","status":400}
`),
	}}

	// Use the mock secure storage
	ss, adapter := sstorage.NewMockAdapter()
	ccs := compcreds.NewCompCredStore("secret/hms-cred", ss)
	s.ss = ss
	s.ccs = ccs
	for i, test := range tests {
		adapter.LookupNum = 0
		adapter.LookupData = test.ssRespLData
		results.GetRFEndpointsFilter.Return.entries = test.hmsdsRespIDs
		results.GetRFEndpointsFilter.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
		// Testing filter no valid if there's an error.
		if test.hmsdsRespErr != nil {
			continue
		}

		if len(test.expectedFilter.Type) !=
			len(results.GetRFEndpointsFilter.Input.f.Type) ||
			len(test.expectedFilter.FQDN) !=
				len(results.GetRFEndpointsFilter.Input.f.FQDN) ||
			len(test.expectedFilter.UUID) !=
				len(results.GetRFEndpointsFilter.Input.f.UUID) {

			t.Errorf("Test %v Failed: Bad len: filter '%v'; Received filter '%v'",
				i, test.expectedFilter, results.GetRFEndpointsFilter.Input.f)

		} else if (len(test.expectedFilter.Type) > 0 &&
			test.expectedFilter.Type[0] !=
				results.GetRFEndpointsFilter.Input.f.Type[0]) ||
			(len(test.expectedFilter.FQDN) > 0 &&
				test.expectedFilter.FQDN[0] !=
					results.GetRFEndpointsFilter.Input.f.FQDN[0]) ||
			(len(test.expectedFilter.UUID) > 0 &&
				test.expectedFilter.UUID[0] !=
					results.GetRFEndpointsFilter.Input.f.UUID[0]) {

			t.Errorf("Test %v Failed: Expected filter '%v'; Received filter '%v'",
				i, test.expectedFilter, results.GetRFEndpointsFilter.Input.f)
		}
	}
}

func TestDoRedfishEndpointQueryGet(t *testing.T) {
	redfishEndpointArr := sm.RedfishEndpointArray{RedfishEndpoints: []*sm.RedfishEndpoint{&stest.TestRedfishEndpointNodeBMC1}}
	payload, _ := json.Marshal(redfishEndpointArr)

	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespEPs []*sm.RedfishEndpoint
		hmsdsRespErr error
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/Query/s0",
		hmsdsRespEPs: []*sm.RedfishEndpoint{&stest.TestRedfishEndpointNodeBMC1},
		hmsdsRespErr: nil,
		expectedResp: payload,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/Query/all",
		hmsdsRespEPs: []*sm.RedfishEndpoint{&stest.TestRedfishEndpointNodeBMC1},
		hmsdsRespErr: nil,
		expectedResp: payload,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/Query/s0",
		hmsdsRespEPs: nil,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"RedfishEndpoints":null}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/Query/s0",
		hmsdsRespEPs: nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/RedfishEndpoints/Query/s0",
		hmsdsRespEPs: nil,
		hmsdsRespErr: errors.New("Argument was not a valid xname"),
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetRFEndpointsAll.Return.entries = test.hmsdsRespEPs
		results.GetRFEndpointsAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoRedfishEndpointDelete(t *testing.T) {
	tests := []struct {
		reqType         string
		reqURI          string
		hmsdsDidDelete  bool
		hmsdsExpectedId []string
		hmsdsRespErr    error
		expectedResp    []byte
	}{{
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		hmsdsDidDelete:  true,
		hmsdsExpectedId: []string{"x0c0s14b0"},
		hmsdsRespErr:    nil,
		expectedResp:    json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{"x0c0s14b0"},
		hmsdsRespErr:    nil,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/RedfishEndpoints/ ",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{" "},
		hmsdsRespErr:    hmsds.ErrHMSDSArgBadID,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{"x0c0s14b0"},
		hmsdsRespErr:    hmsds.ErrHMSDSArgBadID,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{"x0c0s14b0"},
		hmsdsRespErr:    errors.New("Unknown error"),
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteRFEndpointByIDSetEmpty.Return.changed = test.hmsdsDidDelete
		results.DeleteRFEndpointByIDSetEmpty.Return.affectedIds = test.hmsdsExpectedId
		results.DeleteRFEndpointByIDSetEmpty.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsDidDelete && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if strings.Join(test.hmsdsExpectedId, "") != results.DeleteRFEndpointByIDSetEmpty.Input.id {
			t.Errorf("Test %v Failed: Expected endpoint ID is '%v'; Received endpoint ID '%v'", i, test.hmsdsExpectedId, results.DeleteRFEndpointByIDSetEmpty.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoRedfishEndpointsDeleteAll(t *testing.T) {
	tests := []struct {
		reqType          string
		reqURI           string
		hmsdsRespCount   int64
		hmsdsExpectedIds []string
		hmsdsRespErr     error
		expectedResp     []byte
	}{{
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		hmsdsRespCount:   1,
		hmsdsExpectedIds: []string{"x0c0s14b0"},
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted 1 entries"}`),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		hmsdsRespCount:   4,
		hmsdsExpectedIds: []string{"x0c0s14b0", "x0c0s15b0", "x0c0s16b0", "x0c0s17b0"},
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted 4 entries"}`),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		hmsdsRespCount:   0,
		hmsdsExpectedIds: []string{""},
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}`),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		hmsdsRespCount:   0,
		hmsdsExpectedIds: []string{""},
		hmsdsRespErr:     hmsds.ErrHMSDSArgBadID,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		hmsdsRespCount:   0,
		hmsdsExpectedIds: []string{""},
		hmsdsRespErr:     errors.New("DB Error"),
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteRFEndpointsAllSetEmpty.Return.numRows = test.hmsdsRespCount
		results.DeleteRFEndpointsAllSetEmpty.Return.affectedIds = test.hmsdsExpectedIds
		results.DeleteRFEndpointsAllSetEmpty.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && test.hmsdsRespCount > 0 && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespCount == 0) && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoRedfishEndpointPut(t *testing.T) {
	var writeVaultInitial = s.writeVault
	defer func() {
		s.writeVault = writeVaultInitial
	}()
	s.writeVault = false

	var rawRFEEnabled = true
	var rawRFEUseSSDP = false
	var rawRFEMACRequired = false
	var rawRFERediscOnUpdate = true
	rawRedfishEndpoint := rf.RawRedfishEP{"x0c0s14b0", "NodeBMC", "", "10.10.255.11", "local", "10.10.255.11", &rawRFEEnabled, "d4c6d22f-6983-42d8-8e6e-e1fd6d675c17", "root", "********", &rawRFEUseSSDP, &rawRFEMACRequired, "", "", &rawRFERediscOnUpdate, ""}
	redfishEndpointDescPtr, _ := rf.NewRedfishEPDescription(&rawRedfishEndpoint)
	redfishEndpointPtr := sm.NewRedfishEndpoint(redfishEndpointDescPtr)

	tests := []struct {
		reqType              string
		reqURI               string
		reqBody              []byte
		expectedRedfishEP    *sm.RedfishEndpoint
		hmsdsRespEntry       *sm.RedfishEndpoint
		hmsdsRespAffectedIds []string
		hmsdsRespErr         error
		expectedResp         []byte
	}{{
		reqType:              "PUT",
		reqURI:               "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		reqBody:              json.RawMessage(`{"Type":"NodeBMC","Hostname":"10.10.255.11","Domain":"local","FQDN":"10.10.255.11","Enabled":true,"UUID":"d4c6d22f-6983-42d8-8e6e-e1fd6d675c17","User":"root","Password":"********","IPAddress":"10.10.255.11","RediscoverOnUpdate":true,"DiscoveryInfo":{"LastDiscoveryStatus":"NotYetQueried"}}`),
		expectedRedfishEP:    redfishEndpointPtr,
		hmsdsRespEntry:       redfishEndpointPtr,
		hmsdsRespAffectedIds: []string{"x0c0s14b0"},
		hmsdsRespErr:         nil,
		expectedResp:         json.RawMessage(`{"ID":"x0c0s14b0","Type":"NodeBMC","Hostname":"10.10.255.11","Domain":"local","FQDN":"10.10.255.11","Enabled":true,"UUID":"d4c6d22f-6983-42d8-8e6e-e1fd6d675c17","User":"root","Password":"********","IPAddress":"10.10.255.11","RediscoverOnUpdate":true,"DiscoveryInfo":{"LastDiscoveryStatus":"NotYetQueried"}}` + "\n"),
	}, {
		reqType:              "PUT",
		reqURI:               "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		reqBody:              json.RawMessage(`{"Type":"NodeBMC","Hostname":"10.10.255.11","Domain":"local","FQDN":"10.10.255.11","Enabled":true,"UUID":"d4c6d22f-6983-42d8-8e6e-e1fd6d675c17","User":"root","Password":"********","IPAddress":"10.10.255.11","RediscoverOnUpdate":true,"DiscoveryInfo":{"LastDiscoveryStatus":"NotYetQueried"}}`),
		expectedRedfishEP:    &sm.RedfishEndpoint{},
		hmsdsRespEntry:       redfishEndpointPtr,
		hmsdsRespAffectedIds: []string{"x0c0s14b0"},
		hmsdsRespErr:         hmsds.ErrHMSDSArgNoMatch,
		expectedResp:         json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"operation 'PUT' failed during store","status":500}` + "\n"),
	}, {
		reqType:              "PUT",
		reqURI:               "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		reqBody:              json.RawMessage(`{"Type":"NodeBMC","Hostname":"10.10.255.11","Domain":"local","FQDN":"10.10.255.11","Enabled":true,"UUID":"d4c6d22f-6983-42d8-8e6e-e1fd6d675c17","User":"root","Password":"********","IPAddress":"10.10.255.11","RediscoverOnUpdate":true,"DiscoveryInfo":{"LastDiscoveryStatus":"NotYetQueried"}}`),
		expectedRedfishEP:    &sm.RedfishEndpoint{},
		hmsdsRespEntry:       redfishEndpointPtr,
		hmsdsRespAffectedIds: []string{"x0c0s14b0"},
		hmsdsRespErr:         hmsds.ErrHMSDSArgBadID,
		expectedResp:         json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"operation 'PUT' failed during store","status":500}` + "\n"),
	}, {
		reqType:              "PUT",
		reqURI:               "https://localhost/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0",
		reqBody:              json.RawMessage(`{"ID":"x0c0s16b1","Type":"NodeBMC","Hostname":"10.10.255.11","Domain":"local","FQDN":"10.10.255.11","Enabled":true,"UUID":"d4c6d22f-6983-42d8-8e6e-e1fd6d675c17","User":"root","Password":"********","IPAddress":"10.10.255.11","RediscoverOnUpdate":true,"DiscoveryInfo":{"LastDiscoveryStatus":"NotYetQueried"}}`),
		expectedRedfishEP:    &sm.RedfishEndpoint{},
		hmsdsRespEntry:       redfishEndpointPtr,
		hmsdsRespAffectedIds: []string{"x0c0s14b0"},
		hmsdsRespErr:         hmsds.ErrHMSDSArgBadID,
		expectedResp:         json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"xname in URL and PUT body do not match","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.UpdateRFEndpointNoDiscInfo.Input.ep = redfishEndpointPtr
		results.UpdateRFEndpointNoDiscInfo.Return.entry = test.hmsdsRespEntry
		results.UpdateRFEndpointNoDiscInfo.Return.affectedIds = test.hmsdsRespAffectedIds
		results.UpdateRFEndpointNoDiscInfo.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.expectedRedfishEP != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.expectedRedfishEP == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoRedfishEndpointsPost(t *testing.T) {
	var writeVaultInitial = s.writeVault
	defer func() {
		s.writeVault = writeVaultInitial
	}()
	s.writeVault = false

	redfishEndpointArr1 := &stest.TestRedfishEndpointArrayNodeBMC1
	payload1, _ := json.Marshal(redfishEndpointArr1)

	redfishEndpointArr2 := &stest.TestRedfishEndpointArray
	payload2, _ := json.Marshal(redfishEndpointArr2)

	tests := []struct {
		reqType            string
		reqURI             string
		reqBody            []byte
		expectedRedfishEPs *sm.RedfishEndpointArray
		hmsdsRespErr       error
		expectedResp       []byte
	}{{
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		reqBody:            payload1,
		expectedRedfishEPs: redfishEndpointArr1,
		hmsdsRespErr:       nil,
		expectedResp:       json.RawMessage(`[{"URI":"/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0"}]` + "\n"),
	}, {
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		reqBody:            payload2,
		expectedRedfishEPs: redfishEndpointArr2,
		hmsdsRespErr:       nil,
		expectedResp:       json.RawMessage(`[{"URI":"/hsm/v2/Inventory/RedfishEndpoints/x0c0s14b0"},{"URI":"/hsm/v2/Inventory/RedfishEndpoints/x11c0s1b0"},{"URI":"/hsm/v2/Inventory/RedfishEndpoints/x11c0r1b0"},{"URI":"/hsm/v2/Inventory/RedfishEndpoints/x6c1b0"}]` + "\n"),
	}, {
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		reqBody:            payload2,
		expectedRedfishEPs: redfishEndpointArr2,
		hmsdsRespErr:       errors.New("Unknown Error"),
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"operation 'POST' failed during store. ","status":500}` + "\n"),
	}, {
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/RedfishEndpoints",
		reqBody:            payload2,
		expectedRedfishEPs: redfishEndpointArr2,
		hmsdsRespErr:       hmsds.ErrHMSDSArgBadID,
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"operation 'POST' failed during store. ","status":500}` + "\n"),
	}}

	for i, test := range tests {
		if i == 0 {
			results.InsertRFEndpoints.Input.eps = redfishEndpointArr1
		} else {
			results.InsertRFEndpoints.Input.eps = redfishEndpointArr2
		}
		results.InsertRFEndpoints.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.expectedRedfishEPs != nil && w.Code != http.StatusCreated {
			t.Errorf("Response code was %v; want 201", w.Code)
		} else if (test.hmsdsRespErr != nil || test.expectedRedfishEPs == nil) && w.Code == http.StatusCreated {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// ComponentEndpoints
//////////////////////////////////////////////////////////////////////////////

func TestDoComponentEndpointGet(t *testing.T) {
	xname := stest.TestCompEndpointNodeBMC.ID
	payload, _ := json.Marshal(&stest.TestCompEndpointNodeBMC)

	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespEP  *sm.ComponentEndpoint
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints/" + xname,
		hmsdsRespEP:  &stest.TestCompEndpointNodeBMC,
		hmsdsRespErr: nil,
		expectedID:   xname,
		expectedResp: payload,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints/" + xname,
		hmsdsRespEP:  nil,
		hmsdsRespErr: nil,
		expectedID:   xname,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints/" + xname,
		hmsdsRespEP:  nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedID:   xname,
		expectedResp: jsonErrHMSDSArgMissing,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints/" + xname,
		hmsdsRespEP:  nil,
		hmsdsRespErr: errors.New("unexpected DB error"),
		expectedID:   xname,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetCompEndpointByID.Return.entry = test.hmsdsRespEP
		results.GetCompEndpointByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an unexpected error '%s' occurred while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespEP != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespEP == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetCompEndpointByID.Input.id {
			t.Errorf("Test %v Failed: Expected endpoint ID is '%v'; Received endpoint ID '%v'", i, test.expectedID, results.GetCompEndpointByID.Input.id)
		}

		if strings.TrimSpace(string(test.expectedResp)) != strings.TrimSpace(string(w.Body.Bytes())) {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentEndpointsGet(t *testing.T) {
	componentEndpointArray := sm.ComponentEndpointArray{ComponentEndpoints: stest.SampleCompEndpoints}
	payload, _ := json.Marshal(componentEndpointArray)

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespEPs   []*sm.ComponentEndpoint
		hmsdsRespErr   error
		expectedFilter *hmsds.CompEPFilter
		expectedResp   []byte
		expectError    bool
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/ComponentEndpoints",
		hmsdsRespEPs:   stest.SampleCompEndpoints,
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.CompEPFilter{},
		expectedResp:   payload,
		expectError:    false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints?id=x1c4",
		hmsdsRespEPs: stest.SampleCompEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEPFilter{
			ID: []string{"x1c4"},
		},
		expectedResp: payload,
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints?redfish_ep=x1c4b0",
		hmsdsRespEPs: stest.SampleCompEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEPFilter{
			RfEndpointID: []string{"x1c4b0"},
		},
		expectedResp: payload,
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints?type=NodeBMC",
		hmsdsRespEPs: stest.SampleCompEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEPFilter{
			Type: []string{"NodeBMC"},
		},
		expectedResp: payload,
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints?redfish_type=Chassis",
		hmsdsRespEPs: stest.SampleCompEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEPFilter{
			RedfishType: []string{"Chassis"},
		},
		expectedResp: payload,
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints?id=x1c4&redfish_ep=x1c4b0&type=Chassis&redfish_type=Chassis",
		hmsdsRespEPs: stest.SampleCompEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEPFilter{
			ID:           []string{"x1c4"},
			RfEndpointID: []string{"x1c4b0"},
			Type:         []string{"Chassis"},
			RedfishType:  []string{"Chassis"},
		},
		expectedResp: payload,
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ComponentEndpoints?type=fake",
		hmsdsRespEPs: stest.SampleCompEndpoints,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadType,
		expectedFilter: &hmsds.CompEPFilter{
			Type: []string{"fake"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not a valid HMS Type","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/ComponentEndpoints",
		hmsdsRespEPs:   stest.SampleCompEndpoints,
		hmsdsRespErr:   errors.New("Unknown error"),
		expectedFilter: &hmsds.CompEPFilter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
		expectError:    true,
	}}

	for i, test := range tests {
		results.GetCompEndpointsFilter.Return.entries = test.hmsdsRespEPs
		results.GetCompEndpointsFilter.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.GetCompEndpointsFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected component endpoint filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetCompEndpointsFilter.Input.f)
			}
		}

		if strings.TrimSpace(string(test.expectedResp)) != strings.TrimSpace(string(w.Body.Bytes())) {
			t.Errorf("Test %v Failed: Expected body is '%v';\n Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentEndpointDelete(t *testing.T) {
	tests := []struct {
		reqType         string
		reqURI          string
		hmsdsDidDelete  bool
		hmsdsExpectedId []string
		hmsdsRespErr    error
		expectedResp    []byte
	}{{
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/ComponentEndpoints/x1c4",
		hmsdsDidDelete:  true,
		hmsdsExpectedId: []string{"x1c4"},
		hmsdsRespErr:    nil,
		expectedResp:    json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/ComponentEndpoints/x1c4",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{"x1c4"},
		hmsdsRespErr:    nil,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/ComponentEndpoints/ ",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{" "},
		hmsdsRespErr:    hmsds.ErrHMSDSArgBadID,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/ComponentEndpoints/x1c4",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{"x1c4"},
		hmsdsRespErr:    hmsds.ErrHMSDSArgBadID,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}` + "\n"),
	}, {
		reqType:         "DELETE",
		reqURI:          "https://localhost/hsm/v2/Inventory/ComponentEndpoints/x1c4",
		hmsdsDidDelete:  false,
		hmsdsExpectedId: []string{"x1c4"},
		hmsdsRespErr:    errors.New("Unknown error"),
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteCompEndpointByIDSetEmpty.Return.changed = test.hmsdsDidDelete
		results.DeleteCompEndpointByIDSetEmpty.Return.affectedIds = test.hmsdsExpectedId
		results.DeleteCompEndpointByIDSetEmpty.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsDidDelete && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if strings.Join(test.hmsdsExpectedId, "") != results.DeleteCompEndpointByIDSetEmpty.Input.id {
			t.Errorf("Test %v Failed: Expected endpoint ID is '%v'; Received endpoint ID '%v'", i, test.hmsdsExpectedId, results.DeleteCompEndpointByIDSetEmpty.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoComponentEndpointsDeleteAll(t *testing.T) {
	tests := []struct {
		reqType          string
		reqURI           string
		hmsdsRespCount   int64
		hmsdsExpectedIds []string
		hmsdsRespErr     error
		expectedResp     []byte
	}{{
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/ComponentEndpoints",
		hmsdsRespCount:   1,
		hmsdsExpectedIds: []string{"x0c0s14b0"},
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted 1 entries"}`),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/ComponentEndpoints",
		hmsdsRespCount:   4,
		hmsdsExpectedIds: []string{"x1c4", "x1c5", "x1c6", "x1c7"},
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"code":0,"message":"deleted 4 entries"}`),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/ComponentEndpoints",
		hmsdsRespCount:   0,
		hmsdsExpectedIds: []string{""},
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}`),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/ComponentEndpoints",
		hmsdsRespCount:   0,
		hmsdsExpectedIds: []string{""},
		hmsdsRespErr:     hmsds.ErrHMSDSArgBadID,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
	}, {
		reqType:          "DELETE",
		reqURI:           "https://localhost/hsm/v2/Inventory/ComponentEndpoints",
		hmsdsRespCount:   0,
		hmsdsExpectedIds: []string{""},
		hmsdsRespErr:     errors.New("DB Error"),
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.DeleteCompEndpointsAllSetEmpty.Return.numRows = test.hmsdsRespCount
		results.DeleteCompEndpointsAllSetEmpty.Return.affectedIds = test.hmsdsExpectedIds
		results.DeleteCompEndpointsAllSetEmpty.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && test.hmsdsRespCount > 0 && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespCount == 0) && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// ServiceEndpoints
///////////////////////////////////////////////////////////////////////////////

func TestDoServiceEndpointGet(t *testing.T) {
	payload1, _ := json.Marshal(stest.TestServiceEndpointUpdate1)

	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsRespID  *sm.ServiceEndpoint
		hmsdsRespErr error
		expectedSVC  string
		expectedID   string
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService/RedfishEndpoints/x0c0s21b0",
		hmsdsRespID:  &stest.TestServiceEndpointUpdate1,
		hmsdsRespErr: nil,
		expectedSVC:  "UpdateService",
		expectedID:   "x0c0s21b0",
		expectedResp: payload1,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService/RedfishEndpoints/foo",
		hmsdsRespID:  nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadID,
		expectedSVC:  "UpdateService",
		expectedID:   "foo",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}`),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService/RedfishEndpoints/foo",
		hmsdsRespID:  nil,
		hmsdsRespErr: errors.New("Argument was not a valid xname"),
		expectedSVC:  "UpdateService",
		expectedID:   "foo",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}`),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/foo/RedfishEndpoints/x0c0s21b0",
		hmsdsRespID:  nil,
		hmsdsRespErr: nil,
		expectedSVC:  "foo",
		expectedID:   "x0c0s21b0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such service under redfish endpoint.","status":404}`),
	}}

	for i, test := range tests {
		results.GetServiceEndpointByID.Return.entry = test.hmsdsRespID
		results.GetServiceEndpointByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && test.hmsdsRespID != nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespID == nil) && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
		// Testing filter not valid if there's an error.
		if test.hmsdsRespErr != nil {
			continue
		}

		if test.expectedSVC != results.GetServiceEndpointByID.Input.svc {
			t.Errorf("Test %v Failed: Expected service '%v'; Received service '%v'",
				i, test.expectedSVC, results.GetServiceEndpointByID.Input.svc)
		}
		if test.expectedID != results.GetServiceEndpointByID.Input.id {
			t.Errorf("Test %v Failed: Expected id '%v'; Received id '%v'",
				i, test.expectedID, results.GetServiceEndpointByID.Input.id)
		}
	}
}

func TestDoServiceEndpointsGetAll(t *testing.T) {
	payload1, _ := json.Marshal(stest.TestServiceEndpointArray)
	payload2, _ := json.Marshal(stest.TestServiceEndpointArrayUpdates)
	payload3, _ := json.Marshal(stest.TestServiceEndpointArrayUpdate1)
	payload4, _ := json.Marshal(stest.TestServiceEndpointArrayRFEP)

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespIDs   []*sm.ServiceEndpoint
		hmsdsRespErr   error
		expectedFilter hmsds.ServiceEPFilter
		expectedResp   []byte
	}{{
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/ServiceEndpoints",
		hmsdsRespIDs:   stest.TestServiceEndpointArray.ServiceEndpoints,
		hmsdsRespErr:   nil,
		expectedFilter: hmsds.ServiceEPFilter{},
		expectedResp:   payload1,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints?service=UpdateService",
		hmsdsRespIDs: stest.TestServiceEndpointArrayUpdates.ServiceEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: hmsds.ServiceEPFilter{
			Service: []string{"UpdateService"},
		},
		expectedResp: payload2,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints?service=UpdateService&redfish_ep=x0c0s21b0",
		hmsdsRespIDs: stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: hmsds.ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"x0c0s21b0"},
		},
		expectedResp: payload3,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints?redfish_ep=x0c0s21b0",
		hmsdsRespIDs: stest.TestServiceEndpointArrayRFEP.ServiceEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: hmsds.ServiceEPFilter{
			RfEndpointID: []string{"x0c0s21b0"},
		},
		expectedResp: payload4,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints?redfish_ep=foo",
		hmsdsRespIDs: nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadID,
		expectedFilter: hmsds.ServiceEPFilter{
			RfEndpointID: []string{"foo"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}`),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints?redfish_ep=foo",
		hmsdsRespIDs: nil,
		hmsdsRespErr: errors.New("Argument was not a valid xname"),
		expectedFilter: hmsds.ServiceEPFilter{
			RfEndpointID: []string{"foo"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}`),
	}}

	for i, test := range tests {
		results.GetServiceEndpointsFilter.Return.entries = test.hmsdsRespIDs
		results.GetServiceEndpointsFilter.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
		// Testing filter not valid if there's an error.
		if test.hmsdsRespErr != nil {
			continue
		}

		if len(test.expectedFilter.Service) !=
			len(results.GetServiceEndpointsFilter.Input.f.Service) ||
			len(test.expectedFilter.RfEndpointID) !=
				len(results.GetServiceEndpointsFilter.Input.f.RfEndpointID) {

			t.Errorf("Test %v Failed: Bad len: filter '%v'; Received filter '%v'",
				i, test.expectedFilter, results.GetServiceEndpointsFilter.Input.f)

		} else if (len(test.expectedFilter.Service) > 0 &&
			test.expectedFilter.Service[0] !=
				results.GetServiceEndpointsFilter.Input.f.Service[0]) ||
			(len(test.expectedFilter.RfEndpointID) > 0 &&
				test.expectedFilter.RfEndpointID[0] !=
					results.GetServiceEndpointsFilter.Input.f.RfEndpointID[0]) {

			t.Errorf("Test %v Failed: Expected filter '%v'; Received filter '%v'",
				i, test.expectedFilter, results.GetServiceEndpointsFilter.Input.f)
		}
	}
}

func TestDoServiceEndpointsGet(t *testing.T) {
	payload2, _ := json.Marshal(stest.TestServiceEndpointArrayUpdates)
	payload3, _ := json.Marshal(stest.TestServiceEndpointArrayUpdate1)

	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespIDs   []*sm.ServiceEndpoint
		hmsdsRespErr   error
		expectedFilter hmsds.ServiceEPFilter
		expectedResp   []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService",
		hmsdsRespIDs: stest.TestServiceEndpointArrayUpdates.ServiceEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: hmsds.ServiceEPFilter{
			Service: []string{"UpdateService"},
		},
		expectedResp: payload2,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService?redfish_ep=x0c0s21b0",
		hmsdsRespIDs: stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints,
		hmsdsRespErr: nil,
		expectedFilter: hmsds.ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"x0c0s21b0"},
		},
		expectedResp: payload3,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService?redfish_ep=foo",
		hmsdsRespIDs: nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadID,
		expectedFilter: hmsds.ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"foo"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}`),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService?redfish_ep=foo",
		hmsdsRespIDs: nil,
		hmsdsRespErr: errors.New("Argument was not a valid xname"),
		expectedFilter: hmsds.ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"foo"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}`),
	}}

	for i, test := range tests {
		results.GetServiceEndpointsFilter.Return.entries = test.hmsdsRespIDs
		results.GetServiceEndpointsFilter.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
		// Testing filter not valid if there's an error.
		if test.hmsdsRespErr != nil {
			continue
		}

		if len(test.expectedFilter.Service) !=
			len(results.GetServiceEndpointsFilter.Input.f.Service) ||
			len(test.expectedFilter.RfEndpointID) !=
				len(results.GetServiceEndpointsFilter.Input.f.RfEndpointID) {

			t.Errorf("Test %v Failed: Bad len: filter '%v'; Received filter '%v'",
				i, test.expectedFilter, results.GetServiceEndpointsFilter.Input.f)

		} else if (len(test.expectedFilter.Service) > 0 &&
			test.expectedFilter.Service[0] !=
				results.GetServiceEndpointsFilter.Input.f.Service[0]) ||
			(len(test.expectedFilter.RfEndpointID) > 0 &&
				test.expectedFilter.RfEndpointID[0] !=
					results.GetServiceEndpointsFilter.Input.f.RfEndpointID[0]) {

			t.Errorf("Test %v Failed: Expected filter '%v'; Received filter '%v'",
				i, test.expectedFilter, results.GetServiceEndpointsFilter.Input.f)
		}
	}
}

func TestDoServiceEndpointDelete(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    bool
		hmsdsRespErr error
		expectedSVC  string
		expectedID   string
		expectedResp []byte
	}{{
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService/RedfishEndpoints/x0c0s21b0",
		hmsdsResp:    true,
		hmsdsRespErr: nil,
		expectedSVC:  "UpdateService",
		expectedID:   "x0c0s21b0",
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 1 entry"}`),
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService/RedfishEndpoints/foo",
		hmsdsResp:    false,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadID,
		expectedSVC:  "UpdateService",
		expectedID:   "foo",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}`),
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/UpdateService/RedfishEndpoints/foo",
		hmsdsResp:    false,
		hmsdsRespErr: errors.New("Argument was not a valid xname"),
		expectedSVC:  "UpdateService",
		expectedID:   "foo",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}`),
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/ServiceEndpoints/foo/RedfishEndpoints/x0c0s21b0",
		hmsdsResp:    false,
		hmsdsRespErr: nil,
		expectedSVC:  "foo",
		expectedID:   "x0c0s21b0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such service under redfish endpoint.","status":404}`),
	}}

	for i, test := range tests {
		results.DeleteServiceEndpointByID.Return.changed = test.hmsdsResp
		results.DeleteServiceEndpointByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && test.hmsdsResp && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if (test.hmsdsRespErr != nil || !test.hmsdsResp) && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
		// Testing filter not valid if there's an error.
		if test.hmsdsRespErr != nil {
			continue
		}

		if test.expectedSVC != results.DeleteServiceEndpointByID.Input.svc {
			t.Errorf("Test %v Failed: Expected service '%v'; Received service '%v'",
				i, test.expectedSVC, results.DeleteServiceEndpointByID.Input.svc)
		}
		if test.expectedID != results.DeleteServiceEndpointByID.Input.id {
			t.Errorf("Test %v Failed: Expected id '%v'; Received id '%v'",
				i, test.expectedID, results.DeleteServiceEndpointByID.Input.id)
		}
	}
}

func TestDoServiceEndpointsDeleteAll(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsRespCount int64
		hmsdsRespErr   error
		expectedResp   []byte
	}{{
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Inventory/ServiceEndpoints",
		hmsdsRespCount: 4,
		hmsdsRespErr:   nil,
		expectedResp:   json.RawMessage(`{"code":0,"message":"deleted 4 entries"}`),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Inventory/ServiceEndpoints",
		hmsdsRespCount: 0,
		hmsdsRespErr:   nil,
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}`),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Inventory/ServiceEndpoints",
		hmsdsRespCount: 0,
		hmsdsRespErr:   hmsds.ErrHMSDSArgBadID,
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid xname ID","status":400}`),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/Inventory/ServiceEndpoints",
		hmsdsRespCount: 0,
		hmsdsRespErr:   errors.New("DB Error"),
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"failed to query DB.","status":500}`),
	}}

	for i, test := range tests {
		results.DeleteServiceEndpointsAll.Return.numRows = test.hmsdsRespCount
		results.DeleteServiceEndpointsAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if test.hmsdsRespErr == nil && test.hmsdsRespCount > 0 && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespCount == 0) && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if strings.TrimSpace(string(test.expectedResp)) !=
			strings.TrimSpace(string(w.Body.Bytes())) {

			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'",
				i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// Component Ethernet Interfaces - V1 API
//////////////////////////////////////////////////////////////////////////////

func TestDoCompEthInterfaceDeleteAll(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    int64
		hmsdsRespErr error
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    1,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 1 entries"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    7,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 7 entries"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    0,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    0,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.DeleteCompEthInterfacesAll.Return.numRows = test.hmsdsResp
		results.DeleteCompEthInterfacesAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// Component Ethernet Interfaces - V2 API
//////////////////////////////////////////////////////////////////////////////

func TestDoCompEthInterfacesGetV2(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.CompEthInterfaceV2
		hmsdsRespErr   error
		expectedFilter *hmsds.CompEthInterfaceFilter
		expectedResp   []byte
		expectError    bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				Desc:    "My description",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{},
		expectedResp:   json.RawMessage(`[{"ID":"a4bf0138ee65","Description":"My description","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[{"IPAddress":"10.254.2.14"}]}]` + "\n"),
		expectError:    false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				Desc:    "My description",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14", Network: "HMN"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{},
		expectedResp:   json.RawMessage(`[{"ID":"a4bf0138ee65","Description":"My description","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[{"IPAddress":"10.254.2.14","Network":"HMN"}]}]` + "\n"),
		expectError:    false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces?componentid=x3000c0s26b0",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:         "a4bf0138ee65",
				MACAddr:    "a4:bf:01:38:ee:65",
				IPAddrs:    []sm.IPAddressMapping{},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			CompID: []string{"x3000c0s26b0"},
		},
		expectedResp: json.RawMessage(`[{"ID":"a4bf0138ee65","Description":"","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[]}]` + "\n"),
		expectError:  false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces?ipaddress=!&olderthan=2020-05-14T21:59:02.363448Z",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
			}},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			IPAddr:    []string{"!"},
			OlderThan: "2020-05-14T21:59:02.363448Z",
		},
		expectedResp: json.RawMessage(`[{"ID":"a4bf0138ee65","Description":"","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"","Type":"","IPAddresses":[{"IPAddress":"10.254.2.14"}]}]` + "\n"),
		expectError:  false,
	}, {
		reqType:        "GET",
		reqURI:         "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:      []*sm.CompEthInterfaceV2{},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{},
		expectedResp:   json.RawMessage(`[]` + "\n"),
		expectError:    false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				Desc:    "My description",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14"},
					{IPAddr: "10.252.1.11"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{},
		expectedResp:   json.RawMessage(`[{"ID":"a4bf0138ee65","Description":"My description","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[{"IPAddress":"10.254.2.14"},{"IPAddress":"10.252.1.11"}]}]` + "\n"),
		expectError:    false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				Desc:    "My description",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14", Network: "HMN"},
					{IPAddr: "10.252.1.11"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{},
		expectedResp:   json.RawMessage(`[{"ID":"a4bf0138ee65","Description":"My description","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[{"IPAddress":"10.254.2.14","Network":"HMN"},{"IPAddress":"10.252.1.11"}]}]` + "\n"),
		expectError:    false,
	}}

	for i, test := range tests {
		results.GetCompEthInterfaceFilter.Return.ceis = test.hmsdsResp
		results.GetCompEthInterfaceFilter.Return.err = test.hmsdsRespErr
		results.GetCompEthInterfaceFilter.Input.f = &hmsds.CompEthInterfaceFilter{}
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr == nil {
			if !compareCompEthInterfaceFilter(*test.expectedFilter, *results.GetCompEthInterfaceFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetCompEthInterfaceFilter.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfacePostV2(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedCEI  *sm.CompEthInterfaceV2
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		reqBody:      json.RawMessage(`{"description":"My description","macaddress":"a4:bf:01:38:ee:65", "IPAddresses":[{"ipaddress":"10.254.2.14"}] }`),
		hmsdsRespErr: nil,
		expectedCEI: &sm.CompEthInterfaceV2{
			ID:      "a4bf0138ee65",
			Desc:    "My description",
			MACAddr: "a4:bf:01:38:ee:65",
			IPAddrs: []sm.IPAddressMapping{
				{IPAddr: "10.254.2.14"},
			},
		},
		expectedResp: json.RawMessage(`{"URI":"/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		reqBody:      json.RawMessage(`{"description":"My description","IPAddresses":[{"ipaddress":"10.254.2.14"}] }`),
		hmsdsRespErr: nil,
		expectedCEI:  nil,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"MAC string has invalid length","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		reqBody:      json.RawMessage(`{"macaddress":"a4:bf:01:38:ee:65","componentid":"foo"}`),
		hmsdsRespErr: nil,
		expectedCEI:  nil,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid CompEthInterface component ID","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		reqBody:      json.RawMessage(`{"macaddress":"a4:bf:01:38:ee:65"}`),
		hmsdsRespErr: hmsds.ErrHMSDSDuplicateKey,
		expectedCEI: &sm.CompEthInterfaceV2{
			ID:      "a4bf0138ee65",
			MACAddr: "a4:bf:01:38:ee:65",
			IPAddrs: []sm.IPAddressMapping{},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing component ethernet interface that has the same MAC address.","status":409}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		reqBody:      json.RawMessage(`{"description":"My description","macaddress":"a4:bf:01:38:ee:65","IPAddresses":[{"IPAddress":""}]}`),
		hmsdsRespErr: nil,
		expectedCEI:  nil,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid CompEthInterface IP Address","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.InsertCompEthInterface.Return.err = test.hmsdsRespErr
		results.InsertCompEthInterface.Input.cei = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusCreated {
			t.Errorf("Test %v Failed: Response code was %v; want 201", i, w.Code)
		} else if test.expectError && w.Code == http.StatusCreated {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedCEI, results.InsertCompEthInterface.Input.cei) {
				t.Errorf("Test %v Failed: Expected component ethernet interface is '%v'; Received '%v'", i, test.expectedCEI, results.InsertCompEthInterface.Input.cei)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfaceDeleteAllV2(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    int64
		hmsdsRespErr error
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    1,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 1 entries"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    8,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 8 entries"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    0,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces",
		hmsdsResp:    0,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.DeleteCompEthInterfacesAll.Return.numRows = test.hmsdsResp
		results.DeleteCompEthInterfacesAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfaceGetV2(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.CompEthInterfaceV2
		hmsdsRespErr   error
		expectedFilter *hmsds.CompEthInterfaceFilter
		expectedResp   []byte
		expectError    bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				Desc:    "My description",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			ID: []string{"a4bf0138ee65"},
		},
		expectedResp: json.RawMessage(`{"ID":"a4bf0138ee65","Description":"My description","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[{"IPAddress":"10.254.2.14"}]}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		hmsdsResp:    []*sm.CompEthInterfaceV2{},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			ID: []string{"a4bf0138ee65"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such component ethernet interface: a4bf0138ee65","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			ID: []string{"a4bf0138ee65"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetCompEthInterfaceFilter.Return.ceis = test.hmsdsResp
		results.GetCompEthInterfaceFilter.Return.err = test.hmsdsRespErr
		results.GetCompEthInterfaceFilter.Input.f = &hmsds.CompEthInterfaceFilter{}
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !compareCompEthInterfaceFilter(*test.expectedFilter, *results.GetCompEthInterfaceFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetCompEthInterfaceFilter.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfaceDeleteV2(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    bool
		hmsdsRespErr error
		expectedId   string
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		hmsdsResp:    true,
		hmsdsRespErr: nil,
		expectedId:   "a4bf0138ee65",
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		hmsdsResp:    false,
		hmsdsRespErr: nil,
		expectedId:   "a4bf0138ee65",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such component ethernet interface.","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		hmsdsResp:    false,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedId:   "a4bf0138ee65",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.DeleteCompEthInterfaceByID.Return.didDelete = test.hmsdsResp
		results.DeleteCompEthInterfaceByID.Return.err = test.hmsdsRespErr
		results.DeleteCompEthInterfaceByID.Input.id = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedId != results.DeleteCompEthInterfaceByID.Input.id {
				t.Errorf("Test %v Failed: Expected id is '%v'; Received '%v'", i, test.expectedId, results.DeleteCompEthInterfaceByID.Input.id)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfacePatchV2(t *testing.T) {
	ipaddr := "10.254.2.14"
	desc := "My description"
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsResp     *sm.CompEthInterfaceV2
		hmsdsRespErr  error
		expectedId    string
		expectedPatch *sm.CompEthInterfaceV2Patch
		expectedResp  []byte
		expectError   bool
	}{{
		reqType: "PATCH",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		reqBody: json.RawMessage(`{"IPAddresses": [{"IPAddress": "10.254.2.14"}] }`),
		hmsdsResp: &sm.CompEthInterfaceV2{
			ID:      "a4bf0138ee65",
			Desc:    "My description",
			MACAddr: "a4:bf:01:38:ee:65",
			IPAddrs: []sm.IPAddressMapping{
				{IPAddr: "10.254.2.14"},
			},
			LastUpdate: "2020-05-13T21:59:02.363448Z",
			CompID:     "x3000c0s26b0",
			Type:       "NodeBMC",
		},
		hmsdsRespErr:  nil,
		expectedId:    "a4bf0138ee65",
		expectedPatch: &sm.CompEthInterfaceV2Patch{IPAddrs: &[]sm.IPAddressMapping{{IPAddr: ipaddr}}},
		expectedResp:  json.RawMessage(`{"ID":"a4bf0138ee65","Description":"My description","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[{"IPAddress":"10.254.2.14"}]}` + "\n"),
		expectError:   false,
	}, {
		reqType: "PATCH",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		reqBody: json.RawMessage(`{"Description":"My description"}`),
		hmsdsResp: &sm.CompEthInterfaceV2{
			ID:      "a4bf0138ee65",
			Desc:    "My description",
			MACAddr: "a4:bf:01:38:ee:65",
			IPAddrs: []sm.IPAddressMapping{
				{IPAddr: "10.254.2.14"},
			},
			LastUpdate: "2020-05-13T21:59:02.363448Z",
			CompID:     "x3000c0s26b0",
			Type:       "NodeBMC",
		},
		hmsdsRespErr:  nil,
		expectedId:    "a4bf0138ee65",
		expectedPatch: &sm.CompEthInterfaceV2Patch{Desc: &desc},
		expectedResp:  json.RawMessage(`{"ID":"a4bf0138ee65","Description":"My description","MACAddress":"a4:bf:01:38:ee:65","LastUpdate":"2020-05-13T21:59:02.363448Z","ComponentID":"x3000c0s26b0","Type":"NodeBMC","IPAddresses":[{"IPAddress":"10.254.2.14"}]}` + "\n"),
		expectError:   false,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		reqBody:       json.RawMessage(`{}`),
		hmsdsRespErr:  nil,
		expectedId:    "a4bf0138ee65",
		expectedPatch: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Request must have at least one patch field.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65",
		reqBody:       json.RawMessage(`{"IPAddresses": [{"IPAddress": "10.254.2.14"}] }`),
		hmsdsResp:     nil,
		hmsdsRespErr:  nil,
		expectedId:    "a4bf0138ee65",
		expectedPatch: &sm.CompEthInterfaceV2Patch{IPAddrs: &[]sm.IPAddressMapping{{IPAddr: ipaddr}}},
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such component ethernet interface.","status":404}` + "\n"),
		expectError:   true,
	}}

	for i, test := range tests {
		results.UpdateCompEthInterface.Return.cei = test.hmsdsResp
		results.UpdateCompEthInterface.Return.err = test.hmsdsRespErr
		results.UpdateCompEthInterface.Input.id = ""
		results.UpdateCompEthInterface.Input.ceip = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 204", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedId != results.UpdateCompEthInterface.Input.id {
				t.Errorf("Test %v Failed: Expected id is '%v'; Received '%v'", i, test.expectedId, results.UpdateCompEthInterface.Input.id)
			}
			if !reflect.DeepEqual(test.expectedPatch, results.UpdateCompEthInterface.Input.ceip) {
				t.Errorf("Test %v Failed: Expected patch is '%v'; Received '%v'", i, test.expectedPatch, results.UpdateCompEthInterface.Input.ceip)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfaceIPAddressesGetV2(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.CompEthInterfaceV2
		hmsdsRespErr   error
		expectedFilter *hmsds.CompEthInterfaceFilter
		expectedResp   []byte
		expectError    bool
	}{{ // Test 0 Perform a Get on valid MAC address with a single IP
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				Desc:    "My description",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			ID: []string{"a4bf0138ee65"},
		},
		expectedResp: json.RawMessage(`[{"IPAddress":"10.254.2.14"}]` + "\n"),
		expectError:  false,
	}, { // Test 1 Perform a Get on valid MAC address with a two IPs
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		hmsdsResp: []*sm.CompEthInterfaceV2{
			&sm.CompEthInterfaceV2{
				ID:      "a4bf0138ee65",
				Desc:    "My description",
				MACAddr: "a4:bf:01:38:ee:65",
				IPAddrs: []sm.IPAddressMapping{
					{IPAddr: "10.254.2.14"},
					{IPAddr: "10.252.2.14"},
				},
				LastUpdate: "2020-05-13T21:59:02.363448Z",
				CompID:     "x3000c0s26b0",
				Type:       "NodeBMC",
			}},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			ID: []string{"a4bf0138ee65"},
		},
		expectedResp: json.RawMessage(`[{"IPAddress":"10.254.2.14"},{"IPAddress":"10.252.2.14"}]` + "\n"),
		expectError:  false,
	}, { // Test 2 Perform a Get on a unknown MAC address
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		hmsdsResp:    []*sm.CompEthInterfaceV2{},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			ID: []string{"a4bf0138ee65"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such component ethernet interface: a4bf0138ee65","status":404}` + "\n"),
		expectError:  true,
	}, { // Test 3 Error from HMSDS
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedFilter: &hmsds.CompEthInterfaceFilter{
			ID: []string{"a4bf0138ee65"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetCompEthInterfaceFilter.Return.ceis = test.hmsdsResp
		results.GetCompEthInterfaceFilter.Return.err = test.hmsdsRespErr
		results.GetCompEthInterfaceFilter.Input.f = &hmsds.CompEthInterfaceFilter{}
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !compareCompEthInterfaceFilter(*test.expectedFilter, *results.GetCompEthInterfaceFilter.Input.f) {
				t.Errorf("Test %v Failed: Expected filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetCompEthInterfaceFilter.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfaceIPAddressPostV2(t *testing.T) {
	tests := []struct {
		reqType            string
		reqURI             string
		reqBody            []byte
		hmsdsResp          string
		hmsdsRespErr       error
		expectedMACAddress string
		expectedIPM        *sm.IPAddressMapping
		expectedResp       []byte
		expectError        bool
	}{{ // Test 0 POST a new IP Address mapping
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		reqBody:            json.RawMessage(`{"IPAddress": "10.254.2.14" }`),
		hmsdsResp:          "10.254.2.14",
		hmsdsRespErr:       nil,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPM: &sm.IPAddressMapping{
			IPAddr: "10.254.2.14",
		},
		expectedResp: json.RawMessage(`{"URI":"/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.2.14"}` + "\n"),
		expectError:  false,
	}, { // Test 1 POST a new IP Address mapping with a network
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		reqBody:            json.RawMessage(`{"IPAddress": "10.254.2.14", "Network": "HMN" }`),
		hmsdsResp:          "10.254.2.14",
		hmsdsRespErr:       nil,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPM: &sm.IPAddressMapping{
			IPAddr:  "10.254.2.14",
			Network: "HMN",
		},
		expectedResp: json.RawMessage(`{"URI":"/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.2.14"}` + "\n"),
		expectError:  false,
	}, { // Test 2 POST a duplicate IP address
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		reqBody:            json.RawMessage(`{"IPAddress": "10.254.2.14", "Network": "HMN" }`),
		hmsdsResp:          "",
		hmsdsRespErr:       hmsds.ErrHMSDSDuplicateKey,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPM: &sm.IPAddressMapping{
			IPAddr:  "10.254.2.14",
			Network: "HMN",
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing IP Address on the same ethernet interface.","status":409}` + "\n"),
		expectError:  true,
	}, { // Test 3 POST to a non-existant MAC address (404)
		reqType:            "POST",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses",
		reqBody:            json.RawMessage(`{"IPAddress": "10.254.2.14", "Network": "HMN" }`),
		hmsdsResp:          "",
		hmsdsRespErr:       hmsds.ErrHMSDSNoCompEthInterface,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPM: &sm.IPAddressMapping{
			IPAddr:  "10.254.2.14",
			Network: "HMN",
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such component ethernet interface: a4bf0138ee65","status":404}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.AddCompEthInterfaceIPAddress.Return.id = test.hmsdsResp
		results.AddCompEthInterfaceIPAddress.Return.err = test.hmsdsRespErr
		results.AddCompEthInterfaceIPAddress.Input.id = ""
		results.AddCompEthInterfaceIPAddress.Input.ipm = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusCreated {
			t.Errorf("Test %v Failed: Response code was %v; want 201", i, w.Code)
		} else if test.expectError && w.Code == http.StatusCreated {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedMACAddress != results.AddCompEthInterfaceIPAddress.Input.id {
				t.Errorf("Test %v Failed: Expected MAC address is '%v'; Received '%v'", i, test.expectedMACAddress, results.AddCompEthInterfaceIPAddress.Input.id)
			}

			if !reflect.DeepEqual(test.expectedIPM, results.AddCompEthInterfaceIPAddress.Input.ipm) {
				t.Errorf("Test %v Failed: Expected IP Address Mapping is '%v'; Received '%v'", i, test.expectedIPM, results.AddCompEthInterfaceIPAddress.Input.ipm)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfaceIPAddressPatchV2(t *testing.T) {
	network := "HMN"

	tests := []struct {
		reqType            string
		reqURI             string
		reqBody            []byte
		hmsdsResp          *sm.IPAddressMapping
		hmsdsRespErr       error
		expectedMACAddress string
		expectedIPAddress  string
		expectedPatch      *sm.IPAddressMappingPatch
		expectedResp       []byte
		expectError        bool
	}{{ // Test 0 PATCH to an existing IP address, update network feild
		reqType: "PATCH",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.0.10",
		reqBody: json.RawMessage(`{ "Network": "HMN" }`),
		hmsdsResp: &sm.IPAddressMapping{
			IPAddr:  "10.254.0.10",
			Network: "HMN",
		},
		hmsdsRespErr:       nil,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPAddress:  "10.254.0.10",
		expectedPatch: &sm.IPAddressMappingPatch{
			Network: &network,
		},
		expectedResp: json.RawMessage(`{"IPAddress":"10.254.0.10","Network":"HMN"}` + "\n"),
		expectError:  false,
	}, { // Test 1 PATCH with empty request (do nothing)
		reqType: "PATCH",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.0.10",
		reqBody: json.RawMessage(`{}`),
		hmsdsResp: &sm.IPAddressMapping{
			IPAddr:  "10.254.0.10",
			Network: "HMN",
		},
		hmsdsRespErr:       nil,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPAddress:  "10.254.0.10",
		expectedPatch:      &sm.IPAddressMappingPatch{},
		expectedResp:       json.RawMessage(`{"IPAddress":"10.254.0.10","Network":"HMN"}` + "\n"),
		expectError:        false,
	}, { // Test 2 PATCH to a nonexistant IP Address (404)
		reqType:            "PATCH",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.0.10",
		reqBody:            json.RawMessage(`{ "Network": "HMN" }`),
		hmsdsResp:          nil,
		hmsdsRespErr:       nil,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPAddress:  "10.254.0.10",
		expectedPatch: &sm.IPAddressMappingPatch{
			Network: &network,
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such IP address in component ethernet interface.","status":404}` + "\n"),
		expectError:  true,
	}, { // Test 3 PATCH to a nonexistant MAC address (404)
		reqType: "PATCH",
		reqURI:  "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.0.10",
		reqBody: json.RawMessage(`{ "Network": "HMN" }`),
		hmsdsResp: &sm.IPAddressMapping{
			IPAddr:  "10.254.0.10",
			Network: "HMN",
		},
		hmsdsRespErr:       hmsds.ErrHMSDSNoCompEthInterface,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPAddress:  "10.254.0.10",
		expectedPatch: &sm.IPAddressMappingPatch{
			Network: &network,
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: no such component ethernet interface","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.UpdateCompEthInterfaceIPAddress.Output.ipm = test.hmsdsResp
		results.UpdateCompEthInterfaceIPAddress.Output.err = test.hmsdsRespErr
		results.UpdateCompEthInterfaceIPAddress.Input.id = ""
		results.UpdateCompEthInterfaceIPAddress.Input.ipAddr = ""
		results.UpdateCompEthInterfaceIPAddress.Input.ipmPatch = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedMACAddress != results.UpdateCompEthInterfaceIPAddress.Input.id {
				t.Errorf("Test %v Failed: Expected MAC address is '%v'; Received '%v'", i, test.expectedMACAddress, results.UpdateCompEthInterfaceIPAddress.Input.id)
			}

			if test.expectedIPAddress != results.UpdateCompEthInterfaceIPAddress.Input.ipAddr {
				t.Errorf("Test %v Failed: Expected IP Address is '%v'; Received '%v'", i, test.expectedIPAddress, results.UpdateCompEthInterfaceIPAddress.Input.ipAddr)
			}

			if !reflect.DeepEqual(test.expectedPatch, results.UpdateCompEthInterfaceIPAddress.Input.ipmPatch) {
				t.Errorf("Test %v Failed: Expected patch is '%v'; Received '%v'", i, test.expectedPatch, results.UpdateCompEthInterfaceIPAddress.Input.ipmPatch)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompEthInterfaceIPAddressDeleteV2(t *testing.T) {
	tests := []struct {
		reqType            string
		reqURI             string
		hmsdsResp          bool
		hmsdsRespErr       error
		expectedMACAddress string
		expectedIPAddress  string
		expectedResp       []byte
		expectError        bool
	}{{ // Test 0 Delete an existing IP Address
		reqType:            "DELETE",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.0.10",
		hmsdsResp:          true,
		hmsdsRespErr:       nil,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPAddress:  "10.254.0.10",
		expectedResp:       json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
		expectError:        false,
	}, { // Test 1 Delete a non-existant IP Address
		reqType:            "DELETE",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.0.11",
		hmsdsResp:          false,
		hmsdsRespErr:       nil,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPAddress:  "10.254.0.11",
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"component ethernet interface has no such ip address.","status":404}` + "\n"),
		expectError:        true,
	}, { // Test 2 Delete on a non-existant MAC address
		reqType:            "DELETE",
		reqURI:             "https://localhost/hsm/v2/Inventory/EthernetInterfaces/a4bf0138ee65/IPAddresses/10.254.0.10",
		hmsdsResp:          false,
		hmsdsRespErr:       hmsds.ErrHMSDSNoCompEthInterface,
		expectedMACAddress: "a4bf0138ee65",
		expectedIPAddress:  "10.254.0.10",
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such component ethernet interface: a4bf0138ee65","status":404}` + "\n"),
		expectError:        true,
	}}

	for i, test := range tests {
		results.DeleteCompEthInterfaceIPAddress.Output.didDelete = test.hmsdsResp
		results.DeleteCompEthInterfaceIPAddress.Output.err = test.hmsdsRespErr
		results.DeleteCompEthInterfaceIPAddress.Input.id = ""
		results.DeleteCompEthInterfaceIPAddress.Input.ipAddr = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedMACAddress != results.DeleteCompEthInterfaceIPAddress.Input.id {
				t.Errorf("Test %v Failed: Expected MAC address is '%v'; Received '%v'", i, test.expectedMACAddress, results.DeleteCompEthInterfaceIPAddress.Input.id)
			}

			if test.expectedIPAddress != results.DeleteCompEthInterfaceIPAddress.Input.ipAddr {
				t.Errorf("Test %v Failed: Expected IP Address is '%v'; Received '%v'", i, test.expectedIPAddress, results.DeleteCompEthInterfaceIPAddress.Input.ipAddr)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Discovery
///////////////////////////////////////////////////////////////////////////////

func TestDiscoveryStatusGet(t *testing.T) {
	discoveryStatus := sm.DiscoveryStatus{ID: 121, Status: "InProgress", LastUpdate: "654654654"}
	tests := []struct {
		reqType         string
		reqURI          string
		hmsdsRespStatus *sm.DiscoveryStatus
		hmsdsRespErr    error
		expectedID      uint
		expectedResp    []byte
	}{{
		reqType:         "GET",
		reqURI:          "https://localhost/hsm/v2/Inventory/DiscoveryStatus/121",
		hmsdsRespStatus: &discoveryStatus,
		hmsdsRespErr:    nil,
		expectedID:      121,
		expectedResp:    json.RawMessage(`{"ID":121,"Status":"InProgress","LastUpdateTime":"654654654"}` + "\n"),
	}, {
		reqType:         "GET",
		reqURI:          "https://localhost/hsm/v2/Inventory/DiscoveryStatus/32",
		hmsdsRespStatus: nil,
		hmsdsRespErr:    nil,
		expectedID:      32,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such DiscoveryStatus ID.","status":404}` + "\n"),
	}, {
		reqType:         "GET",
		reqURI:          "https://localhost/hsm/v2/Inventory/DiscoveryStatus/0",
		hmsdsRespStatus: nil,
		hmsdsRespErr:    hmsds.ErrHMSDSArgMissing,
		expectedID:      0,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"Failed due to DB access issue.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetDiscoveryStatusByID.Return.status = test.hmsdsRespStatus
		results.GetDiscoveryStatusByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespStatus != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespStatus == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetDiscoveryStatusByID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.GetDiscoveryStatusByID.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDiscoveryStatusGetAll(t *testing.T) {
	discoveryStatus := []*sm.DiscoveryStatus{
		{ID: 121, Status: "Complete", LastUpdate: "654654654432"},
		{ID: 123, Status: "InProgress", LastUpdate: "654654655344"},
		{ID: 124, Status: "InProgress", LastUpdate: "654654654343"},
		{ID: 126, Status: "Pending", LastUpdate: "65465465465"},
		{ID: 127, Status: "NotStarted", LastUpdate: "6546546548789"},
	}
	tests := []struct {
		reqType         string
		reqURI          string
		hmsdsRespStatus []*sm.DiscoveryStatus
		hmsdsRespErr    error
		expectedResp    []byte
	}{{
		reqType:         "GET",
		reqURI:          "https://localhost/hsm/v2/Inventory/DiscoveryStatus",
		hmsdsRespStatus: discoveryStatus,
		hmsdsRespErr:    nil,
		expectedResp:    json.RawMessage(`[{"ID":121,"Status":"Complete","LastUpdateTime":"654654654432"},{"ID":123,"Status":"InProgress","LastUpdateTime":"654654655344"},{"ID":124,"Status":"InProgress","LastUpdateTime":"654654654343"},{"ID":126,"Status":"Pending","LastUpdateTime":"65465465465"},{"ID":127,"Status":"NotStarted","LastUpdateTime":"6546546548789"}]` + "\n"),
	}, {
		reqType:         "GET",
		reqURI:          "https://localhost/hsm/v2/Inventory/DiscoveryStatus?id=23",
		hmsdsRespStatus: nil,
		hmsdsRespErr:    hmsds.ErrHMSDSArgTooMany,
		expectedResp:    json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"Failed due to DB access issue.","status":500}` + "\n"),
	}}

	for i, test := range tests {
		results.GetDiscoveryStatusAll.Return.status = test.hmsdsRespStatus
		results.GetDiscoveryStatusAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsRespStatus != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsRespStatus == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// Groups
//////////////////////////////////////////////////////////////////////////////

func TestDoGroupsGet(t *testing.T) {
	tests := []struct {
		reqType            string
		reqURI             string
		hmsdsRespLabels    []string
		hmsdsRespLabelsErr error
		hmsdsRespGroup     *sm.Group
		hmsdsRespGroupErr  error
		expectedLabel      string
		expectedFiltPart   string
		expectedResp       []byte
		expectError        bool
	}{{
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups",
		hmsdsRespLabels:    []string{"my_group"},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup: &sm.Group{
			Label:       "my_group",
			Description: "This is my group",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespGroupErr: nil,
		expectedLabel:     "my_group",
		expectedFiltPart:  "",
		expectedResp:      json.RawMessage(`[{"label":"my_group","description":"This is my group","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}]` + "\n"),
		expectError:       false,
	}, {
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups?group=my_group&tag=foo&partition=p1",
		hmsdsRespLabels:    []string{"my_group"},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup: &sm.Group{
			Label:       "my_group",
			Description: "This is my group",
			Tags:        []string{"foo", "bar"},
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespGroupErr: nil,
		expectedLabel:     "my_group",
		expectedFiltPart:  "p1",
		expectedResp:      json.RawMessage(`[{"label":"my_group","description":"This is my group","tags":["foo","bar"],"members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}]` + "\n"),
		expectError:       false,
	}, {
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups?group=my_group&tag=bar",
		hmsdsRespLabels:    []string{"my_group"},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup: &sm.Group{
			Label:       "my_group",
			Description: "This is my group",
			Tags:        []string{"foo"},
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespGroupErr: nil,
		expectedLabel:     "my_group",
		expectedFiltPart:  "",
		expectedResp:      json.RawMessage(`[]` + "\n"),
		expectError:       false,
	}, {
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups?group=your_group",
		hmsdsRespLabels:    []string{"my_group"},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup:     nil,
		hmsdsRespGroupErr:  nil,
		expectedLabel:      "",
		expectedFiltPart:   "",
		expectedResp:       json.RawMessage(`[]` + "\n"),
		expectError:        false,
	}, {
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups?partition=",
		hmsdsRespLabels:    []string{},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup:     nil,
		hmsdsRespGroupErr:  nil,
		expectedLabel:      "",
		expectedFiltPart:   "",
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:        true,
	}, {
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups?tag=",
		hmsdsRespLabels:    []string{},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup:     nil,
		hmsdsRespGroupErr:  nil,
		expectedLabel:      "",
		expectedFiltPart:   "",
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid tag.","status":400}` + "\n"),
		expectError:        true,
	}, {
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups?group=",
		hmsdsRespLabels:    []string{},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup:     nil,
		hmsdsRespGroupErr:  nil,
		expectedLabel:      "",
		expectedFiltPart:   "",
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid group label.","status":400}` + "\n"),
		expectError:        true,
	}, {
		reqType:            "GET",
		reqURI:             "https://localhost/hsm/v2/groups",
		hmsdsRespLabels:    []string{"my_group"},
		hmsdsRespLabelsErr: nil,
		hmsdsRespGroup:     nil,
		hmsdsRespGroupErr:  hmsds.ErrHMSDSArgBadArg,
		expectedLabel:      "my_group",
		expectedFiltPart:   "",
		expectedResp:       json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:        true,
	}}

	for i, test := range tests {
		results.GetGroupLabels.Return.labels = test.hmsdsRespLabels
		results.GetGroupLabels.Return.err = test.hmsdsRespLabelsErr
		results.GetGroup.Return.group = test.hmsdsRespGroup
		results.GetGroup.Return.err = test.hmsdsRespGroupErr
		results.GetGroup.Input.label = ""
		results.GetGroup.Input.filt_part = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespLabelsErr == nil {
			if test.expectedLabel != results.GetGroup.Input.label {
				t.Errorf("Test %v Failed: Expected label is '%v'; Received '%v'", i, test.expectedLabel, results.GetGroup.Input.label)
			}
			if test.expectedFiltPart != results.GetGroup.Input.filt_part {
				t.Errorf("Test %v Failed: Expected filt_part is '%v'; Received '%v'", i, test.expectedFiltPart, results.GetGroup.Input.filt_part)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupsPost(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsResp     string
		hmsdsRespErr  error
		expectedGroup *sm.Group
		expectedResp  []byte
		expectError   bool
	}{{
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/groups",
		reqBody:      json.RawMessage(`{"label":"my_group","description":"This is my group","tags":["foo","bar"],"exclusivegroup":"my_system","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}`),
		hmsdsResp:    "my_group",
		hmsdsRespErr: nil,
		expectedGroup: &sm.Group{
			Label:          "my_group",
			Description:    "This is my group",
			Tags:           []string{"foo", "bar"},
			ExclusiveGroup: "my_system",
			Members:        sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		expectedResp: json.RawMessage(`[{"URI":"/hsm/v2/groups/my_group"}]` + "\n"),
		expectError:  false,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups",
		reqBody:       json.RawMessage(`{"description":"This is my group","tags":["foo","bar"],"exclusivegroup":"my_system","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}`),
		hmsdsResp:     "",
		hmsdsRespErr:  nil,
		expectedGroup: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate group: group or partition field has invalid characters","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups",
		reqBody:       json.RawMessage(`{"label":"my_group","description":"This is my group","tags":["foo","bar"],"exclusivegroup":"my_system","members":{"ids":["x0c0s1b0n0","foo"]}}`),
		hmsdsResp:     "",
		hmsdsRespErr:  nil,
		expectedGroup: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate group: got HMSTypeInvalid instead of valid type","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/groups",
		reqBody:      json.RawMessage(`{"label":"my_group","description":"This is my group","tags":["foo","bar"],"exclusivegroup":"my_system","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}`),
		hmsdsResp:    "",
		hmsdsRespErr: hmsds.ErrHMSDSDuplicateKey,
		expectedGroup: &sm.Group{
			Label:          "my_group",
			Description:    "This is my group",
			Tags:           []string{"foo", "bar"},
			ExclusiveGroup: "my_system",
			Members:        sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing group that has the same label.","status":409}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.InsertGroup.Return.label = test.hmsdsResp
		results.InsertGroup.Return.err = test.hmsdsRespErr
		results.InsertGroup.Input.g = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusCreated {
			t.Errorf("Response code was %v; want 201", w.Code)
		} else if test.expectError && w.Code == http.StatusCreated {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !compareGroup(test.expectedGroup, results.InsertGroup.Input.g) {
				t.Errorf("Test %v Failed: Expected group is '%v'; Received '%v'", i, test.expectedGroup, results.InsertGroup.Input.g)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupGet(t *testing.T) {
	tests := []struct {
		reqType          string
		reqURI           string
		hmsdsResp        *sm.Group
		hmsdsRespErr     error
		expectedLabel    string
		expectedFiltPart string
		expectedResp     []byte
		expectError      bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/groups/my_group",
		hmsdsResp: &sm.Group{
			Label:       "my_group",
			Description: "This is my group",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespErr:     nil,
		expectedLabel:    "my_group",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"label":"my_group","description":"This is my group","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}` + "\n"),
		expectError:      false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/groups/my_group?partition=p1",
		hmsdsResp: &sm.Group{
			Label:       "my_group",
			Description: "This is my group",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespErr:     nil,
		expectedLabel:    "my_group",
		expectedFiltPart: "p1",
		expectedResp:     json.RawMessage(`{"label":"my_group","description":"This is my group","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}` + "\n"),
		expectError:      false,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/your_group",
		hmsdsResp:        nil,
		hmsdsRespErr:     nil,
		expectedLabel:    "your_group",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such group: your_group","status":404}` + "\n"),
		expectError:      true,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/my_group?partition=",
		hmsdsResp:        nil,
		hmsdsRespErr:     nil,
		expectedLabel:    "",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:      true,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/My Group",
		hmsdsResp:        nil,
		hmsdsRespErr:     nil,
		expectedLabel:    "",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid group label.","status":400}` + "\n"),
		expectError:      true,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/my_group",
		hmsdsResp:        nil,
		hmsdsRespErr:     hmsds.ErrHMSDSArgBadArg,
		expectedLabel:    "my_group",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:      true,
	}}

	for i, test := range tests {
		results.GetGroup.Return.group = test.hmsdsResp
		results.GetGroup.Return.err = test.hmsdsRespErr
		results.GetGroup.Input.label = ""
		results.GetGroup.Input.filt_part = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedLabel != results.GetGroup.Input.label {
				t.Errorf("Test %v Failed: Expected label is '%v'; Received '%v'", i, test.expectedLabel, results.GetGroup.Input.label)
			}
			if test.expectedFiltPart != results.GetGroup.Input.filt_part {
				t.Errorf("Test %v Failed: Expected filt_part is '%v'; Received '%v'", i, test.expectedFiltPart, results.GetGroup.Input.filt_part)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupDelete(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		hmsdsResp     bool
		hmsdsRespErr  error
		expectedLabel string
		expectedResp  []byte
		expectError   bool
	}{{
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/my_group",
		hmsdsResp:     true,
		hmsdsRespErr:  nil,
		expectedLabel: "my_group",
		expectedResp:  json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
		expectError:   false,
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/your_group",
		hmsdsResp:     false,
		hmsdsRespErr:  nil,
		expectedLabel: "your_group",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such group.","status":404}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/My Group",
		hmsdsResp:     false,
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid group label.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/my_group",
		hmsdsResp:     false,
		hmsdsRespErr:  hmsds.ErrHMSDSArgBadArg,
		expectedLabel: "my_group",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
		expectError:   true,
	}}

	for i, test := range tests {
		results.DeleteGroup.Return.didDelete = test.hmsdsResp
		results.DeleteGroup.Return.err = test.hmsdsRespErr
		results.DeleteGroup.Input.label = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedLabel != results.DeleteGroup.Input.label {
				t.Errorf("Test %v Failed: Expected label is '%v'; Received '%v'", i, test.expectedLabel, results.DeleteGroup.Input.label)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupPatch(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsRespErr  error
		expectedLabel string
		expectedPatch *sm.GroupPatch
		expectedResp  []byte
		expectError   bool
	}{{
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/groups/my_group",
		reqBody:       json.RawMessage(`{"tags":["foo","bar"]}`),
		hmsdsRespErr:  nil,
		expectedLabel: "my_group",
		expectedPatch: &sm.GroupPatch{
			Tags: &[]string{"foo", "bar"},
		},
		expectedResp: nil,
		expectError:  false,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/groups/my_group",
		reqBody:       json.RawMessage(`{}`),
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedPatch: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Request must have at least one patch field.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/groups/my_group",
		reqBody:       json.RawMessage(`{"tags":["foo","B!ar"]}`),
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedPatch: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid tag.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/groups/My Group",
		reqBody:       json.RawMessage(`{"tags":["foo","Bar"]}`),
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedPatch: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid group label.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/groups/your_group",
		reqBody:       json.RawMessage(`{"tags":["foo","Bar"]}`),
		hmsdsRespErr:  hmsds.ErrHMSDSNoGroup,
		expectedLabel: "your_group",
		expectedPatch: &sm.GroupPatch{
			Tags: &[]string{"foo", "Bar"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such group.","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/groups/my_group",
		reqBody:       json.RawMessage(`{"tags":["foo","bar"]}`),
		hmsdsRespErr:  hmsds.ErrHMSDSArgBadArg,
		expectedLabel: "my_group",
		expectedPatch: &sm.GroupPatch{
			Tags: &[]string{"foo", "bar"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.UpdateGroup.Return.err = test.hmsdsRespErr
		results.UpdateGroup.Input.label = ""
		results.UpdateGroup.Input.gp = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusNoContent {
			t.Errorf("Response code was %v; want 204", w.Code)
		} else if test.expectError && w.Code == http.StatusNoContent {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedLabel != results.UpdateGroup.Input.label {
				t.Errorf("Test %v Failed: Expected label is '%v'; Received '%v'", i, test.expectedLabel, results.UpdateGroup.Input.label)
			}
			if !compareGroupPatch(test.expectedPatch, results.UpdateGroup.Input.gp) {
				t.Errorf("Test %v Failed: Expected group patch is '%v'; Received '%v'", i, test.expectedPatch, results.UpdateGroup.Input.gp)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupLabelsGet(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    []string
		hmsdsRespErr error
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/groups/labels",
		hmsdsResp:    []string{"my_group"},
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`["my_group"]` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/groups/labels",
		hmsdsResp:    []string{"my_group", "your_group"},
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`["my_group","your_group"]` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/groups/labels",
		hmsdsResp:    []string{},
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.GetGroupLabels.Return.labels = test.hmsdsResp
		results.GetGroupLabels.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupMembersGet(t *testing.T) {
	tests := []struct {
		reqType          string
		reqURI           string
		hmsdsResp        *sm.Group
		hmsdsRespErr     error
		expectedLabel    string
		expectedFiltPart string
		expectedResp     []byte
		expectError      bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/groups/my_group/members",
		hmsdsResp: &sm.Group{
			Label:       "my_group",
			Description: "This is my group",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespErr:     nil,
		expectedLabel:    "my_group",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}` + "\n"),
		expectError:      false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/groups/My_Group/members?partition=P1",
		hmsdsResp: &sm.Group{
			Label:       "my_group",
			Description: "This is my group",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespErr:     nil,
		expectedLabel:    "my_group",
		expectedFiltPart: "p1",
		expectedResp:     json.RawMessage(`{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}` + "\n"),
		expectError:      false,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/your_group/members",
		hmsdsResp:        nil,
		hmsdsRespErr:     nil,
		expectedLabel:    "your_group",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such group: your_group","status":404}` + "\n"),
		expectError:      true,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/my_group/members?partition=",
		hmsdsResp:        nil,
		hmsdsRespErr:     nil,
		expectedLabel:    "",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:      true,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/MyGroup++/members",
		hmsdsResp:        nil,
		hmsdsRespErr:     nil,
		expectedLabel:    "",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid group label.","status":400}` + "\n"),
		expectError:      true,
	}, {
		reqType:          "GET",
		reqURI:           "https://localhost/hsm/v2/groups/my_group/members",
		hmsdsResp:        nil,
		hmsdsRespErr:     hmsds.ErrHMSDSArgBadArg,
		expectedLabel:    "my_group",
		expectedFiltPart: "",
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:      true,
	}}

	for i, test := range tests {
		results.GetGroup.Return.group = test.hmsdsResp
		results.GetGroup.Return.err = test.hmsdsRespErr
		results.GetGroup.Input.label = ""
		results.GetGroup.Input.filt_part = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedLabel != results.GetGroup.Input.label {
				t.Errorf("Test %v Failed: Expected label is '%v'; Received '%v'", i, test.expectedLabel, results.GetGroup.Input.label)
			}
			if test.expectedFiltPart != results.GetGroup.Input.filt_part {
				t.Errorf("Test %v Failed: Expected filt_part is '%v'; Received '%v'", i, test.expectedFiltPart, results.GetGroup.Input.filt_part)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupMembersPost(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsResp     string
		hmsdsRespErr  error
		expectedLabel string
		expectedID    string
		expectedResp  []byte
		expectError   bool
	}{{
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members",
		reqBody:       json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:     "x0c0s1b0n0",
		hmsdsRespErr:  nil,
		expectedLabel: "my_group",
		expectedID:    "x0c0s1b0n0",
		expectedResp:  json.RawMessage(`[{"URI":"/hsm/v2/groups/my_group/members/x0c0s1b0n0"}]` + "\n"),
		expectError:   false,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members",
		reqBody:       json.RawMessage(`{}`),
		hmsdsResp:     "",
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedID:    "",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname ID","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members",
		reqBody:       json.RawMessage(`{"id":"foo"}`),
		hmsdsResp:     "",
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedID:    "",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname ID","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups/~MyGroup/members",
		reqBody:       json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:     "",
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedID:    "",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid group label.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups/your_group/members",
		reqBody:       json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:     "",
		hmsdsRespErr:  hmsds.ErrHMSDSNoGroup,
		expectedLabel: "your_group",
		expectedID:    "x0c0s1b0n0",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such group: your_group","status":404}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members",
		reqBody:       json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:     "",
		hmsdsRespErr:  hmsds.ErrHMSDSExclusiveGroup,
		expectedLabel: "my_group",
		expectedID:    "x0c0s1b0n0",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing member in another exclusive group.","status":409}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "POST",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members",
		reqBody:       json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:     "",
		hmsdsRespErr:  hmsds.ErrHMSDSDuplicateKey,
		expectedLabel: "my_group",
		expectedID:    "x0c0s1b0n0",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing member in the same group.","status":409}` + "\n"),
		expectError:   true,
	}}

	for i, test := range tests {
		results.AddGroupMember.Return.id = test.hmsdsResp
		results.AddGroupMember.Return.err = test.hmsdsRespErr
		results.AddGroupMember.Input.label = ""
		results.AddGroupMember.Input.id = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusCreated {
			t.Errorf("Response code was %v; want 201", w.Code)
		} else if test.expectError && w.Code == http.StatusCreated {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedLabel != results.AddGroupMember.Input.label {
				t.Errorf("Test %v Failed: Expected label is '%v'; Received '%v'", i, test.expectedLabel, results.AddGroupMember.Input.label)
			}
			if test.expectedID != results.AddGroupMember.Input.id {
				t.Errorf("Test %v Failed: Expected id is '%v'; Received '%v'", i, test.expectedID, results.AddGroupMember.Input.id)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoGroupMemberDelete(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		hmsdsResp     bool
		hmsdsRespErr  error
		expectedLabel string
		expectedID    string
		expectedResp  []byte
		expectError   bool
	}{{
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members/x0c0s1b0n0",
		hmsdsResp:     true,
		hmsdsRespErr:  nil,
		expectedLabel: "my_group",
		expectedID:    "x0c0s1b0n0",
		expectedResp:  json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
		expectError:   false,
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/your_group/members/x0c0s1b0n0",
		hmsdsResp:     false,
		hmsdsRespErr:  nil,
		expectedLabel: "your_group",
		expectedID:    "x0c0s1b0n0",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"group has no such member.","status":404}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/$MyGroup/members/x0c0s1b0n0",
		hmsdsResp:     false,
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedID:    "",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid group label.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members/foo",
		hmsdsResp:     false,
		hmsdsRespErr:  nil,
		expectedLabel: "",
		expectedID:    "",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname ID","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/groups/my_group/members/x0c0s1b0n0",
		hmsdsResp:     false,
		hmsdsRespErr:  hmsds.ErrHMSDSArgBadArg,
		expectedLabel: "my_group",
		expectedID:    "x0c0s1b0n0",
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
		expectError:   true,
	}}

	for i, test := range tests {
		results.DeleteGroupMember.Return.didDelete = test.hmsdsResp
		results.DeleteGroupMember.Return.err = test.hmsdsRespErr
		results.DeleteGroupMember.Input.label = ""
		results.DeleteGroupMember.Input.id = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedLabel != results.DeleteGroupMember.Input.label {
				t.Errorf("Test %v Failed: Expected label is '%v'; Received '%v'", i, test.expectedLabel, results.DeleteGroupMember.Input.label)
			}
			if test.expectedID != results.DeleteGroupMember.Input.id {
				t.Errorf("Test %v Failed: Expected xname ID is '%v'; Received '%v'", i, test.expectedID, results.DeleteGroupMember.Input.id)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// Partitions
//////////////////////////////////////////////////////////////////////////////

func TestDoPartitionsGet(t *testing.T) {
	tests := []struct {
		reqType           string
		reqURI            string
		hmsdsRespNames    []string
		hmsdsRespNamesErr error
		hmsdsRespPart     *sm.Partition
		hmsdsRespPartErr  error
		expectedName      string
		expectedResp      []byte
		expectError       bool
	}{{
		reqType:           "GET",
		reqURI:            "https://localhost/hsm/v2/partitions",
		hmsdsRespNames:    []string{"p1"},
		hmsdsRespNamesErr: nil,
		hmsdsRespPart: &sm.Partition{
			Name:        "p1",
			Description: "This is my partition",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespPartErr: nil,
		expectedName:     "p1",
		expectedResp:     json.RawMessage(`[{"name":"p1","description":"This is my partition","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}]` + "\n"),
		expectError:      false,
	}, {
		reqType:           "GET",
		reqURI:            "https://localhost/hsm/v2/partitions?partition=p1&tag=foo",
		hmsdsRespNames:    []string{"p1"},
		hmsdsRespNamesErr: nil,
		hmsdsRespPart: &sm.Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespPartErr: nil,
		expectedName:     "p1",
		expectedResp:     json.RawMessage(`[{"name":"p1","description":"This is my partition","tags":["foo","bar"],"members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}]` + "\n"),
		expectError:      false,
	}, {
		reqType:           "GET",
		reqURI:            "https://localhost/hsm/v2/partitions?partition=p1&tag=bar",
		hmsdsRespNames:    []string{"p1"},
		hmsdsRespNamesErr: nil,
		hmsdsRespPart: &sm.Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{"foo"},
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespPartErr: nil,
		expectedName:     "p1",
		expectedResp:     json.RawMessage(`[]` + "\n"),
		expectError:      false,
	}, {
		reqType:           "GET",
		reqURI:            "https://localhost/hsm/v2/partitions?partition=p2",
		hmsdsRespNames:    []string{"p1"},
		hmsdsRespNamesErr: nil,
		hmsdsRespPart:     nil,
		hmsdsRespPartErr:  nil,
		expectedName:      "",
		expectedResp:      json.RawMessage(`[]` + "\n"),
		expectError:       false,
	}, {
		reqType:           "GET",
		reqURI:            "https://localhost/hsm/v2/partitions?partition=",
		hmsdsRespNames:    []string{},
		hmsdsRespNamesErr: nil,
		hmsdsRespPart:     nil,
		hmsdsRespPartErr:  nil,
		expectedName:      "",
		expectedResp:      json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:       true,
	}, {
		reqType:           "GET",
		reqURI:            "https://localhost/hsm/v2/partitions?tag=",
		hmsdsRespNames:    []string{},
		hmsdsRespNamesErr: nil,
		hmsdsRespPart:     nil,
		hmsdsRespPartErr:  nil,
		expectedName:      "",
		expectedResp:      json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid tag.","status":400}` + "\n"),
		expectError:       true,
	}, {
		reqType:           "GET",
		reqURI:            "https://localhost/hsm/v2/partitions",
		hmsdsRespNames:    []string{"p1"},
		hmsdsRespNamesErr: nil,
		hmsdsRespPart:     nil,
		hmsdsRespPartErr:  hmsds.ErrHMSDSArgBadArg,
		expectedName:      "p1",
		expectedResp:      json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:       true,
	}}

	for i, test := range tests {
		results.GetPartitionNames.Return.pnames = test.hmsdsRespNames
		results.GetPartitionNames.Return.err = test.hmsdsRespNamesErr
		results.GetPartition.Return.partition = test.hmsdsRespPart
		results.GetPartition.Return.err = test.hmsdsRespPartErr
		results.GetPartition.Input.pname = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespNamesErr == nil {
			if test.expectedName != results.GetPartition.Input.pname {
				t.Errorf("Test %v Failed: Expected name is '%v'; Received '%v'", i, test.expectedName, results.GetPartition.Input.pname)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionsPost(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsResp    string
		hmsdsRespErr error
		expectedPart *sm.Partition
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions",
		reqBody:      json.RawMessage(`{"name":"p1","description":"This is my partition","tags":["foo","bar"],"members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}`),
		hmsdsResp:    "p1",
		hmsdsRespErr: nil,
		expectedPart: &sm.Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		expectedResp: json.RawMessage(`[{"URI":"/hsm/v2/partitions/p1"}]` + "\n"),
		expectError:  false,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions",
		reqBody:      json.RawMessage(`{"description":"This is my partition","tags":["foo","bar"],"members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}`),
		hmsdsResp:    "",
		hmsdsRespErr: nil,
		expectedPart: nil,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate partition: Bad partition name. Must be p# or p#.#","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions",
		reqBody:      json.RawMessage(`{"name":"p1","description":"This is my partition","tags":["foo","bar"],"members":{"ids":["x0c0s1b0n0","foo"]}}`),
		hmsdsResp:    "",
		hmsdsRespErr: nil,
		expectedPart: nil,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate partition: got HMSTypeInvalid instead of valid type","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions",
		reqBody:      json.RawMessage(`{"name":"p1","description":"This is my partition","tags":["foo","bar"],"members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}`),
		hmsdsResp:    "",
		hmsdsRespErr: hmsds.ErrHMSDSDuplicateKey,
		expectedPart: &sm.Partition{
			Name:        "p1",
			Description: "This is my partition",
			Tags:        []string{"foo", "bar"},
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing partition that has the same name.","status":409}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.InsertPartition.Return.pname = test.hmsdsResp
		results.InsertPartition.Return.err = test.hmsdsRespErr
		results.InsertPartition.Input.p = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusCreated {
			t.Errorf("Response code was %v; want 201", w.Code)
		} else if test.expectError && w.Code == http.StatusCreated {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !comparePartition(test.expectedPart, results.InsertPartition.Input.p) {
				t.Errorf("Test %v Failed: Expected partition is '%v'; Received '%v'", i, test.expectedPart, results.InsertPartition.Input.p)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionGet(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    *sm.Partition
		hmsdsRespErr error
		expectedName string
		expectedResp []byte
		expectError  bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/partitions/p1",
		hmsdsResp: &sm.Partition{
			Name:        "p1",
			Description: "This is my partition",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespErr: nil,
		expectedName: "p1",
		expectedResp: json.RawMessage(`{"name":"p1","description":"This is my partition","members":{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/p2",
		hmsdsResp:    nil,
		hmsdsRespErr: nil,
		expectedName: "p2",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such partition: p2","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/&p1",
		hmsdsResp:    nil,
		hmsdsRespErr: nil,
		expectedName: "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/p1",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedName: "p1",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetPartition.Return.partition = test.hmsdsResp
		results.GetPartition.Return.err = test.hmsdsRespErr
		results.GetPartition.Input.pname = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedName != results.GetPartition.Input.pname {
				t.Errorf("Test %v Failed: Expected name is '%v'; Received '%v'", i, test.expectedName, results.GetPartition.Input.pname)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionDelete(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    bool
		hmsdsRespErr error
		expectedName string
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/p1",
		hmsdsResp:    true,
		hmsdsRespErr: nil,
		expectedName: "p1",
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/p2",
		hmsdsResp:    false,
		hmsdsRespErr: nil,
		expectedName: "p2",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such partition.","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/,P1,",
		hmsdsResp:    false,
		hmsdsRespErr: nil,
		expectedName: "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/p1",
		hmsdsResp:    false,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedName: "p1",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.DeletePartition.Return.didDelete = test.hmsdsResp
		results.DeletePartition.Return.err = test.hmsdsRespErr
		results.DeletePartition.Input.pname = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedName != results.DeletePartition.Input.pname {
				t.Errorf("Test %v Failed: Expected name is '%v'; Received '%v'", i, test.expectedName, results.DeletePartition.Input.pname)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionPatch(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		reqBody       []byte
		hmsdsRespErr  error
		expectedName  string
		expectedPatch *sm.PartitionPatch
		expectedResp  []byte
		expectError   bool
	}{{
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/partitions/p1",
		reqBody:      json.RawMessage(`{"tags":["foo","bar"]}`),
		hmsdsRespErr: nil,
		expectedName: "p1",
		expectedPatch: &sm.PartitionPatch{
			Tags: &[]string{"foo", "bar"},
		},
		expectedResp: nil,
		expectError:  false,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/partitions/p1",
		reqBody:       json.RawMessage(`{}`),
		hmsdsRespErr:  nil,
		expectedName:  "",
		expectedPatch: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Request must have at least one patch field.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/partitions/p1",
		reqBody:       json.RawMessage(`{"tags":["foo","B^^^ar"]}`),
		hmsdsRespErr:  nil,
		expectedName:  "",
		expectedPatch: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid tag.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:       "PATCH",
		reqURI:        "https://localhost/hsm/v2/partitions/P!1",
		reqBody:       json.RawMessage(`{"tags":["foo","bar"]}`),
		hmsdsRespErr:  nil,
		expectedName:  "",
		expectedPatch: nil,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:   true,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/partitions/p2",
		reqBody:      json.RawMessage(`{"tags":["foo","bar"]}`),
		hmsdsRespErr: hmsds.ErrHMSDSNoPartition,
		expectedName: "p2",
		expectedPatch: &sm.PartitionPatch{
			Tags: &[]string{"foo", "bar"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such partition.","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "PATCH",
		reqURI:       "https://localhost/hsm/v2/partitions/p1",
		reqBody:      json.RawMessage(`{"tags":["foo","bar"]}`),
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedName: "p1",
		expectedPatch: &sm.PartitionPatch{
			Tags: &[]string{"foo", "bar"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.UpdatePartition.Return.err = test.hmsdsRespErr
		results.UpdatePartition.Input.pname = ""
		results.UpdatePartition.Input.pp = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusNoContent {
			t.Errorf("Response code was %v; want 204", w.Code)
		} else if test.expectError && w.Code == http.StatusNoContent {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedName != results.UpdatePartition.Input.pname {
				t.Errorf("Test %v Failed: Expected name is '%v'; Received '%v'", i, test.expectedName, results.UpdatePartition.Input.pname)
			}
			if !comparePartitionPatch(test.expectedPatch, results.UpdatePartition.Input.pp) {
				t.Errorf("Test %v Failed: Expected partition patch is '%v'; Received '%v'", i, test.expectedPatch, results.UpdatePartition.Input.pp)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionNamesGet(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    []string
		hmsdsRespErr error
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/names",
		hmsdsResp:    []string{"p1"},
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`["p1"]` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/names",
		hmsdsResp:    []string{"p1", "p2"},
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`["p1","p2"]` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/names",
		hmsdsResp:    []string{},
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.GetPartitionNames.Return.pnames = test.hmsdsResp
		results.GetPartitionNames.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionMembersGet(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    *sm.Partition
		hmsdsRespErr error
		expectedName string
		expectedResp []byte
		expectError  bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/partitions/p1/members",
		hmsdsResp: &sm.Partition{
			Name:        "p1",
			Description: "This is my partition",
			Members:     sm.Members{IDs: []string{"x0c0s1b0n0", "x0c0s2b0n0"}},
		},
		hmsdsRespErr: nil,
		expectedName: "p1",
		expectedResp: json.RawMessage(`{"ids":["x0c0s1b0n0","x0c0s2b0n0"]}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/P2/members",
		hmsdsResp:    nil,
		hmsdsRespErr: nil,
		expectedName: "p2",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such partition: p2","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/p==1/members",
		hmsdsResp:    nil,
		hmsdsRespErr: nil,
		expectedName: "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedName: "p1",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetPartition.Return.partition = test.hmsdsResp
		results.GetPartition.Return.err = test.hmsdsRespErr
		results.GetPartition.Input.pname = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedName != results.GetPartition.Input.pname {
				t.Errorf("Test %v Failed: Expected name is '%v'; Received '%v'", i, test.expectedName, results.GetPartition.Input.pname)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionMembersPost(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsResp    string
		hmsdsRespErr error
		expectedName string
		expectedID   string
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members",
		reqBody:      json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:    "x0c0s1b0n0",
		hmsdsRespErr: nil,
		expectedName: "p1",
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`[{"URI":"/hsm/v2/partitions/p1/members/x0c0s1b0n0"}]` + "\n"),
		expectError:  false,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members",
		reqBody:      json.RawMessage(`{}`),
		hmsdsResp:    "",
		hmsdsRespErr: nil,
		expectedName: "",
		expectedID:   "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname ID","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members",
		reqBody:      json.RawMessage(`{"id":"foo"}`),
		hmsdsResp:    "",
		hmsdsRespErr: nil,
		expectedName: "",
		expectedID:   "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname ID","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions/p@@1/members",
		reqBody:      json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:    "",
		hmsdsRespErr: nil,
		expectedName: "",
		expectedID:   "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions/p2/members",
		reqBody:      json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:    "",
		hmsdsRespErr: hmsds.ErrHMSDSNoPartition,
		expectedName: "p2",
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such partition: p2","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members",
		reqBody:      json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:    "",
		hmsdsRespErr: hmsds.ErrHMSDSExclusivePartition,
		expectedName: "p1",
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing member in another partition.","status":409}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members",
		reqBody:      json.RawMessage(`{"id":"x0c0s1b0n0"}`),
		hmsdsResp:    "",
		hmsdsRespErr: hmsds.ErrHMSDSDuplicateKey,
		expectedName: "p1",
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Conflict","detail":"operation would conflict with an existing member in the same partition.","status":409}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.AddPartitionMember.Return.id = test.hmsdsResp
		results.AddPartitionMember.Return.err = test.hmsdsRespErr
		results.AddPartitionMember.Input.pname = ""
		results.AddPartitionMember.Input.id = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusCreated {
			t.Errorf("Response code was %v; want 201", w.Code)
		} else if test.expectError && w.Code == http.StatusCreated {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedName != results.AddPartitionMember.Input.pname {
				t.Errorf("Test %v Failed: Expected name is '%v'; Received '%v'", i, test.expectedName, results.AddPartitionMember.Input.pname)
			}
			if test.expectedID != results.AddPartitionMember.Input.id {
				t.Errorf("Test %v Failed: Expected id is '%v'; Received '%v'", i, test.expectedID, results.AddPartitionMember.Input.id)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPartitionMemberDelete(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    bool
		hmsdsRespErr error
		expectedName string
		expectedID   string
		expectedResp []byte
		expectError  bool
	}{{
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members/x0c0s1b0n0",
		hmsdsResp:    true,
		hmsdsRespErr: nil,
		expectedName: "p1",
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/P2/members/x0c0s1b0n0",
		hmsdsResp:    false,
		hmsdsRespErr: nil,
		expectedName: "p2",
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"partition has no such member.","status":404}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/P!1/members/x0c0s1b0n0",
		hmsdsResp:    false,
		hmsdsRespErr: nil,
		expectedName: "",
		expectedID:   "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid partition name.","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members/foo",
		hmsdsResp:    false,
		hmsdsRespErr: nil,
		expectedName: "",
		expectedID:   "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname ID","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "DELETE",
		reqURI:       "https://localhost/hsm/v2/partitions/p1/members/x0c0s1b0n0",
		hmsdsResp:    false,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedName: "p1",
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"DB query failed.","status":500}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.DeletePartitionMember.Return.didDelete = test.hmsdsResp
		results.DeletePartitionMember.Return.err = test.hmsdsRespErr
		results.DeletePartitionMember.Input.pname = ""
		results.DeletePartitionMember.Input.id = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedName != results.DeletePartitionMember.Input.pname {
				t.Errorf("Test %v Failed: Expected name is '%v'; Received '%v'", i, test.expectedName, results.DeletePartitionMember.Input.pname)
			}
			if test.expectedID != results.DeletePartitionMember.Input.id {
				t.Errorf("Test %v Failed: Expected xname ID is '%v'; Received '%v'", i, test.expectedID, results.DeletePartitionMember.Input.id)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// Memberships
//////////////////////////////////////////////////////////////////////////////

func TestDoMembershipsGet(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsResp      []*sm.Membership
		hmsdsRespErr   error
		expectedFilter *hmsds.ComponentFilter
		expectedResp   []byte
		expectError    bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/memberships",
		hmsdsResp: []*sm.Membership{{
			ID:            "x0c0s1b0n0",
			GroupLabels:   []string{"my_group"},
			PartitionName: "p1",
		}, {
			ID:            "x0c0s2b0n0",
			GroupLabels:   []string{"your_group"},
			PartitionName: "p1",
		}},
		hmsdsRespErr:   nil,
		expectedFilter: &hmsds.ComponentFilter{},
		expectedResp:   json.RawMessage(`[{"id":"x0c0s1b0n0","groupLabels":["my_group"],"partitionName":"p1"},{"id":"x0c0s2b0n0","groupLabels":["your_group"],"partitionName":"p1"}]` + "\n"),
		expectError:    false,
	}, {
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/memberships?type=node",
		hmsdsResp: []*sm.Membership{{
			ID:            "x0c0s1b0n0",
			GroupLabels:   []string{"my_group"},
			PartitionName: "p1",
		}},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.ComponentFilter{
			Type: []string{"node"},
		},
		expectedResp: json.RawMessage(`[{"id":"x0c0s1b0n0","groupLabels":["my_group"],"partitionName":"p1"}]` + "\n"),
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/memberships?type=nodebmc",
		hmsdsResp:    []*sm.Membership{},
		hmsdsRespErr: nil,
		expectedFilter: &hmsds.ComponentFilter{
			Type: []string{"nodebmc"},
		},
		expectedResp: json.RawMessage(`[]` + "\n"),
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/memberships?type=foo",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedFilter: &hmsds.ComponentFilter{
			Type: []string{"foo"},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetMemberships.Return.memberships = test.hmsdsResp
		results.GetMemberships.Return.err = test.hmsdsRespErr
		results.GetMemberships.Input.f = nil
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !compareFilter(*test.expectedFilter, *results.GetMemberships.Input.f) {
				t.Errorf("Test %v Failed: Expected comp filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetMemberships.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoMembershipGet(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    *sm.Membership
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
		expectError  bool
	}{{
		reqType: "GET",
		reqURI:  "https://localhost/hsm/v2/memberships/x0c0s1b0n0",
		hmsdsResp: &sm.Membership{
			ID:            "x0c0s1b0n0",
			GroupLabels:   []string{"my_group"},
			PartitionName: "p1",
		},
		hmsdsRespErr: nil,
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"id":"x0c0s1b0n0","groupLabels":["my_group"],"partitionName":"p1"}` + "\n"),
		expectError:  false,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/memberships/foo",
		hmsdsResp:    nil,
		hmsdsRespErr: nil,
		expectedID:   "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/memberships/x0c0s1b0n0",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgBadArg,
		expectedID:   "x0c0s1b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"bad query param: Argument was not valid","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/memberships/x0c0s2b0n0",
		hmsdsResp:    nil,
		hmsdsRespErr: nil,
		expectedID:   "x0c0s2b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"No such xname: x0c0s2b0n0","status":404}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetMembership.Return.membership = test.hmsdsResp
		results.GetMembership.Return.err = test.hmsdsRespErr
		results.GetMembership.Input.id = ""
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if test.expectedID != results.GetMembership.Input.id {
				t.Errorf("Test %v Failed: Expected ID is '%v'; Received '%v'", i, test.expectedID, results.GetMembership.Input.id)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// V2 Component Reservations
//////////////////////////////////////////////////////////////////////////////

func TestDoCompLocksReservationRemove(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/reservations/remove"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      sm.CompLockV2UpdateResult
		hmsdsRespErr   error
		expectedFilter sm.CompLockV2Filter
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"]}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Counts: sm.CompLockV2Count{
				Total:   1,
				Success: 1,
				Failure: 0,
			},
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{"x3000c0s9b0n0"},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"Counts":{"Total":1,"Success":1,"Failure":0},"Success":{"ComponentIDs":["x3000c0s9b0n0"]},"Failure":[]}` + "\n"),
		expectError:  false,
	}, {
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"]}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: sm.ErrCompLockV2NotFound,
		expectedFilter: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Component not found","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"],"ProcessingModel":"foo"}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadProcessingModel,
		expectedFilter: sm.CompLockV2Filter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid Processing Model","status":400}` + "\n"),
		expectError:    true,
	}}

	for i, test := range tests {
		results.DeleteCompReservationsForce.Return.results = test.hmsdsResp
		results.DeleteCompReservationsForce.Return.err = test.hmsdsRespErr
		results.DeleteCompReservationsForce.Input.f = sm.CompLockV2Filter{}
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.DeleteCompReservationsForce.Input.f) {
				t.Errorf("Test %v Failed: Expected component lock filter is '%v'; Received '%v'", i, test.expectedFilter, results.DeleteCompReservationsForce.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksReservationRelease(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/reservations/release"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      sm.CompLockV2UpdateResult
		hmsdsRespErr   error
		expectedFilter sm.CompLockV2ReservationFilter
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"ReservationKeys":[{"ID":"x3000c0s9b0n0","Key":"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"}]}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Counts: sm.CompLockV2Count{
				Total:   1,
				Success: 1,
				Failure: 0,
			},
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{"x3000c0s9b0n0"},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2ReservationFilter{
			ReservationKeys: []sm.CompLockV2Key{
				sm.CompLockV2Key{
					ID:  "x3000c0s9b0n0",
					Key: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
				},
			},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"Counts":{"Total":1,"Success":1,"Failure":0},"Success":{"ComponentIDs":["x3000c0s9b0n0"]},"Failure":[]}` + "\n"),
		expectError:  false,
	}, {
		reqBody: json.RawMessage(`{"ReservationKeys":[{"ID":"x3000c0s9b0n0","Key":"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"}]}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: sm.ErrCompLockV2NotFound,
		expectedFilter: sm.CompLockV2ReservationFilter{
			ReservationKeys: []sm.CompLockV2Key{
				sm.CompLockV2Key{
					ID:  "x3000c0s9b0n0",
					Key: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
				},
			},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Component not found","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqBody: json.RawMessage(`{"ReservationKeys":[{"ID":"x3000c0s9b0n0","Key":"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"}],"ProcessingModel":"foo"}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadProcessingModel,
		expectedFilter: sm.CompLockV2ReservationFilter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid Processing Model","status":400}` + "\n"),
		expectError:    true,
	}, {
		reqBody: json.RawMessage(`{"ReservationKeys":[{"ID":"x3000c0s9b0n0"}]}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: sm.ErrCompLockV2RKey,
		expectedFilter: sm.CompLockV2ReservationFilter{
			ReservationKeys: []sm.CompLockV2Key{
				sm.CompLockV2Key{
					ID: "x3000c0s9b0n0",
				},
			},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Reservation Key required for operation","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.DeleteCompReservations.Return.results = test.hmsdsResp
		results.DeleteCompReservations.Return.err = test.hmsdsRespErr
		results.DeleteCompReservations.Input.f = sm.CompLockV2ReservationFilter{}
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.DeleteCompReservations.Input.f) {
				t.Errorf("Test %v Failed: Expected component reservation filter is '%v'; Received '%v'", i, test.expectedFilter, results.DeleteCompReservations.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksReservationCreate(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/reservations"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      sm.CompLockV2ReservationResult
		hmsdsRespErr   error
		expectedFilter sm.CompLockV2Filter
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"]}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{
				sm.CompLockV2Success{
					ID:             "x3000c0s9b0n0",
					DeputyKey:      "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17",
					ReservationKey: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
				},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"Success":[{"ID":"x3000c0s9b0n0","DeputyKey":"x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17","ReservationKey":"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"}],"Failure":[]}` + "\n"),
		expectError:  false,
	}, {
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"],"ProcessingModel":"foo"}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadProcessingModel,
		expectedFilter: sm.CompLockV2Filter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid Processing Model","status":400}` + "\n"),
		expectError:    true,
	}}

	for i, test := range tests {
		results.InsertCompReservations.Return.results = test.hmsdsResp
		results.InsertCompReservations.Return.err = test.hmsdsRespErr
		results.InsertCompReservations.Input.f = sm.CompLockV2Filter{}
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.InsertCompReservations.Input.f) {
				t.Errorf("Test %v Failed: Expected component lock filter is '%v'; Received '%v'", i, test.expectedFilter, results.InsertCompReservations.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksServiceReservationCreate(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/service/reservations"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      sm.CompLockV2ReservationResult
		hmsdsRespErr   error
		expectedFilter sm.CompLockV2Filter
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"],"ReservationDuration":1}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{
				sm.CompLockV2Success{
					ID:             "x3000c0s9b0n0",
					DeputyKey:      "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17",
					ReservationKey: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
					ExpirationTime: "2020-10-14T20:05:12.086Z",
				},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ID:                  []string{"x3000c0s9b0n0"},
			ProcessingModel:     sm.CLProcessingModelRigid,
			ReservationDuration: 1,
		},
		expectedResp: json.RawMessage(`{"Success":[{"ID":"x3000c0s9b0n0","DeputyKey":"x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17","ReservationKey":"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d","ExpirationTime":"2020-10-14T20:05:12.086Z"}],"Failure":[]}` + "\n"),
		expectError:  false,
	}, {
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"],"ProcessingModel":"foo","ReservationDuration":1}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadProcessingModel,
		expectedFilter: sm.CompLockV2Filter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid Processing Model","status":400}` + "\n"),
		expectError:    true,
	}, {
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"]}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadDuration,
		expectedFilter: sm.CompLockV2Filter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"ReservationDuration must be greater than 0","status":400}` + "\n"),
		expectError:    true,
	}, {
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"],"ReservationDuration":16}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadDuration,
		expectedFilter: sm.CompLockV2Filter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid Reservation Duration","status":400}` + "\n"),
		expectError:    true,
	}}

	for i, test := range tests {
		results.InsertCompReservations.Return.results = test.hmsdsResp
		results.InsertCompReservations.Return.err = test.hmsdsRespErr
		results.InsertCompReservations.Input.f = sm.CompLockV2Filter{}
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.InsertCompReservations.Input.f) {
				t.Errorf("Test %v Failed: Expected component lock filter is '%v'; Received '%v'", i, test.expectedFilter, results.InsertCompReservations.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksServiceReservationRenew(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/service/reservations/renew"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      sm.CompLockV2UpdateResult
		hmsdsRespErr   error
		expectedFilter sm.CompLockV2ReservationFilter
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"ReservationKeys":[{"ID":"x3000c0s9b0n0","Key":"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"}],"ReservationDuration":1}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Counts: sm.CompLockV2Count{
				Total:   1,
				Success: 1,
				Failure: 0,
			},
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{"x3000c0s9b0n0"},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2ReservationFilter{
			ReservationKeys: []sm.CompLockV2Key{
				sm.CompLockV2Key{
					ID:  "x3000c0s9b0n0",
					Key: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
				},
			},
			ProcessingModel:     sm.CLProcessingModelRigid,
			ReservationDuration: 1,
		},
		expectedResp: json.RawMessage(`{"Counts":{"Total":1,"Success":1,"Failure":0},"Success":{"ComponentIDs":["x3000c0s9b0n0"]},"Failure":[]}` + "\n"),
		expectError:  false,
	}, {
		reqBody: json.RawMessage(`{"ReservationKeys":[{"ID":"x3000c0s9b0n0","Key":"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"}]}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadDuration,
		expectedFilter: sm.CompLockV2ReservationFilter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"ReservationDuration must be greater than 0","status":400}` + "\n"),
		expectError:    true,
	}}

	for i, test := range tests {
		results.UpdateCompReservations.Return.results = test.hmsdsResp
		results.UpdateCompReservations.Return.err = test.hmsdsRespErr
		results.UpdateCompReservations.Input.f = sm.CompLockV2ReservationFilter{}
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.UpdateCompReservations.Input.f) {
				t.Errorf("Test %v Failed: Expected component reservation filter is '%v'; Received '%v'", i, test.expectedFilter, results.UpdateCompReservations.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksServiceReservationCheck(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/service/reservations/check"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      sm.CompLockV2ReservationResult
		hmsdsRespErr   error
		expectedFilter []sm.CompLockV2Key
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"DeputyKeys":[{"ID":"x3000c0s9b0n0","Key":"x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17"}]}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{
				sm.CompLockV2Success{
					ID:        "x3000c0s9b0n0",
					DeputyKey: "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17",
				},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: nil,
		expectedFilter: []sm.CompLockV2Key{
			sm.CompLockV2Key{
				ID:  "x3000c0s9b0n0",
				Key: "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17",
			},
		},
		expectedResp: json.RawMessage(`{"Success":[{"ID":"x3000c0s9b0n0","DeputyKey":"x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17"}],"Failure":[]}` + "\n"),
		expectError:  false,
	}, {
		reqBody: json.RawMessage(`{"DeputyKeys":[{"ID":"x3000c0s9b0n0"}],"ReservationDuration":1}`),
		hmsdsResp: sm.CompLockV2ReservationResult{
			Success: []sm.CompLockV2Success{},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr: sm.ErrCompLockV2DKey,
		expectedFilter: []sm.CompLockV2Key{
			sm.CompLockV2Key{
				ID: "x3000c0s9b0n0",
			},
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Deputy Key required for operation","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetCompReservations.Return.results = test.hmsdsResp
		results.GetCompReservations.Return.err = test.hmsdsRespErr
		results.GetCompReservations.Input.dkeys = []sm.CompLockV2Key{}
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.GetCompReservations.Input.dkeys) {
				t.Errorf("Test %v Failed: Expected deputy keys array is '%v'; Received '%v'", i, test.expectedFilter, results.GetCompReservations.Input.dkeys)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksStatus(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/status"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      []sm.CompLockV2
		hmsdsRespErr   error
		expectedFilter sm.CompLockV2Filter
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"]}`),
		hmsdsResp: []sm.CompLockV2{
			sm.CompLockV2{
				ID:                  "x3000c0s9b0n0",
				Locked:              true,
				Reserved:            false,
				ReservationDisabled: false,
			},
		},
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"Components":[{"ID":"x3000c0s9b0n0","Locked":true,"Reserved":false,"ReservationDisabled":false}]}` + "\n"),
		expectError:  false,
	}}

	for i, test := range tests {
		results.GetCompLocksV2.Return.cls = test.hmsdsResp
		results.GetCompLocksV2.Return.err = test.hmsdsRespErr
		results.GetCompLocksV2.Input.f = sm.CompLockV2Filter{}
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.GetCompLocksV2.Input.f) {
				t.Errorf("Test %v Failed: Expected component lock filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetCompLocksV2.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksStatusGet(t *testing.T) {
	reqType := "GET"
	hmsdsResp := []sm.CompLockV2{
		sm.CompLockV2{
			ID:                  "x3000c0s9b0n0",
			Locked:              true,
			Reserved:            false,
			ReservationDisabled: false,
		},
		sm.CompLockV2{
			ID:                  "x3000c0s10b0n0",
			Locked:              false,
			Reserved:            true,
			ReservationDisabled: true,
		},
	}
	expectedRespGood := json.RawMessage(`{"Components":[{"ID":"x3000c0s9b0n0","Locked":true,"Reserved":false,"ReservationDisabled":false},{"ID":"x3000c0s10b0n0","Locked":false,"Reserved":true,"ReservationDisabled":true}]}` + "\n")

	tests := []struct {
		reqURI         string
		hmsdsRespErr   error
		expectedFilter sm.CompLockV2Filter
		expectedResp   []byte
		expectError    bool
	}{{
		reqURI:       "https://localhost/hsm/v2/locks/status",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?type=Node",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			Type:            []string{"Node"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?state=Ready",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			State:           []string{"Ready"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?role=Management",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			Role:            []string{"Management"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?subrole=Master",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			SubRole:         []string{"Master"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?locked=True",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ProcessingModel: sm.CLProcessingModelRigid,
			Locked:          []string{"True"},
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?reserved=False",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ProcessingModel: sm.CLProcessingModelRigid,
			Reserved:        []string{"False"},
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?reservationDisabled=False",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			ProcessingModel:     sm.CLProcessingModelRigid,
			ReservationDisabled: []string{"False"},
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?type=Node&state=Ready&role=Management&subrole=Master&locked=True&reserved=False&reservationDisabled=False",
		hmsdsRespErr: nil,
		expectedFilter: sm.CompLockV2Filter{
			Type:                []string{"Node"},
			State:               []string{"Ready"},
			Role:                []string{"Management"},
			SubRole:             []string{"Master"},
			ProcessingModel:     sm.CLProcessingModelRigid,
			Locked:              []string{"True"},
			Reserved:            []string{"False"},
			ReservationDisabled: []string{"False"},
		},
		expectedResp: expectedRespGood,
		expectError:  false,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status",
		hmsdsRespErr: sm.ErrCompLockV2NotFound,
		expectedFilter: sm.CompLockV2Filter{
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Component not found","status":400}` + "\n"),
		expectError:  true,
	}, {
		reqURI:       "https://localhost/hsm/v2/locks/status?type=Fake",
		hmsdsRespErr: hmsds.ErrHMSDSArgBadType,
		expectedFilter: sm.CompLockV2Filter{
			Type:            []string{"Fake"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Argument was not a valid HMS Type","status":400}` + "\n"),
		expectError:  true,
	}}

	for i, test := range tests {
		results.GetCompLocksV2.Return.cls = hmsdsResp
		results.GetCompLocksV2.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.GetCompLocksV2.Input.f) {
				t.Errorf("Test %v Failed: Expected component lock filter is '%v'; Received '%v'", i, test.expectedFilter, results.GetCompLocksV2.Input.f)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoCompLocksLock(t *testing.T) {
	reqType := "POST"
	reqURI := "https://localhost/hsm/v2/locks/lock"
	tests := []struct {
		reqBody        []byte
		hmsdsResp      sm.CompLockV2UpdateResult
		hmsdsRespErr   error
		expectedAction string
		expectedFilter sm.CompLockV2Filter
		expectedResp   []byte
		expectError    bool
	}{{
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"]}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Counts: sm.CompLockV2Count{
				Total:   1,
				Success: 1,
				Failure: 0,
			},
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{"x3000c0s9b0n0"},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   nil,
		expectedAction: hmsds.CLUpdateActionLock,
		expectedFilter: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		expectedResp: json.RawMessage(`{"Counts":{"Total":1,"Success":1,"Failure":0},"Success":{"ComponentIDs":["x3000c0s9b0n0"]},"Failure":[]}` + "\n"),
		expectError:  false,
	}, {
		reqBody: json.RawMessage(`{"ComponentIDs":["x3000c0s9b0n0"],"ProcessingModel":"foo"}`),
		hmsdsResp: sm.CompLockV2UpdateResult{
			Success: sm.CompLockV2SuccessArray{
				ComponentIDs: []string{},
			},
			Failure: []sm.CompLockV2Failure{},
		},
		hmsdsRespErr:   sm.ErrCompLockV2BadProcessingModel,
		expectedAction: "",
		expectedFilter: sm.CompLockV2Filter{},
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"Invalid Processing Model","status":400}` + "\n"),
		expectError:    true,
	}}

	for i, test := range tests {
		results.UpdateCompLocksV2.Return.results = test.hmsdsResp
		results.UpdateCompLocksV2.Return.err = test.hmsdsRespErr
		results.UpdateCompLocksV2.Input.f = sm.CompLockV2Filter{}
		results.UpdateCompLocksV2.Input.action = ""
		req, err := http.NewRequest(reqType, reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if !test.expectError && w.Code != http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; want 200", i, w.Code)
		} else if test.expectError && w.Code == http.StatusOK {
			t.Errorf("Test %v Failed: Response code was %v; expected an error", i, w.Code)
		}

		if !test.expectError || test.hmsdsRespErr != nil {
			if !reflect.DeepEqual(test.expectedFilter, results.UpdateCompLocksV2.Input.f) {
				t.Errorf("Test %v Failed: Expected component lock filter is '%v'; Received '%v'", i, test.expectedFilter, results.UpdateCompLocksV2.Input.f)
			}
			if test.expectedAction != results.UpdateCompLocksV2.Input.action {
				t.Errorf("Test %v Failed: Expected component lock action is '%v'; Received '%v'", i, test.expectedAction, results.UpdateCompLocksV2.Input.action)
			}
		}
		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

/////////////////////////////////////////////////////////////////////////////
// Power Maps
//////////////////////////////////////////////////////////////////////////////

func TestDoPowerMapGet(t *testing.T) {
	powerMap, _ := sm.NewPowerMap("x0c0s27b0n0", []string{"x0m0p0j10", "x0m0p0j11"})
	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    *sm.PowerMap
		hmsdsRespErr error
		expectedID   string
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27b0n0",
		hmsdsResp:    powerMap,
		hmsdsRespErr: nil,
		expectedID:   "x0c0s27b0n0",
		expectedResp: json.RawMessage(`{"id":"x0c0s27b0n0","poweredBy":["x0m0p0j10","x0m0p0j11"]}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27b0n0",
		hmsdsResp:    nil,
		hmsdsRespErr: nil,
		expectedID:   "x0c0s27b0n0",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedID:   "x0c0s27",
		expectedResp: jsonErrHMSDSArgMissing,
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps/foo",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedID:   "",
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.GetPowerMapByID.Input.id = ""
		results.GetPowerMapByID.Return.m = test.hmsdsResp
		results.GetPowerMapByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsResp != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsResp == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedID != results.GetPowerMapByID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.GetPowerMapByID.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPowerMapsGet(t *testing.T) {
	var powerMaps = []*sm.PowerMap{
		{ID: "x0c0s21b0n0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
		{ID: "x0c0s22b0n0", PoweredBy: []string{"x0m0p0j12", "x0m0p0j13"}},
	}

	tests := []struct {
		reqType      string
		reqURI       string
		hmsdsResp    []*sm.PowerMap
		hmsdsRespErr error
		expectedResp []byte
	}{{
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		hmsdsResp:    powerMaps,
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`[{"id":"x0c0s21b0n0","poweredBy":["x0m0p0j10","x0m0p0j11"]},{"id":"x0c0s22b0n0","poweredBy":["x0m0p0j12","x0m0p0j13"]}]` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		hmsdsResp:    []*sm.PowerMap{},
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`[]` + "\n"),
	}, {
		reqType:      "GET",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps?ID=foo",
		hmsdsResp:    nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedResp: jsonErrHMSDSArgMissing,
	}}

	for i, test := range tests {
		results.GetPowerMapsAll.Return.ms = test.hmsdsResp
		results.GetPowerMapsAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsResp != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.hmsdsResp == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPowerMapDelete(t *testing.T) {
	tests := []struct {
		reqType        string
		reqURI         string
		hmsdsDidDelete bool
		hmsdsRespErr   error
		expectedID     string
		expectedResp   []byte
	}{{
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27b0n0",
		hmsdsDidDelete: true,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"code":0,"message":"deleted 1 entry"}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27b0n0",
		hmsdsDidDelete: false,
		hmsdsRespErr:   nil,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no such xname.","status":404}` + "\n"),
	}, {
		reqType:        "DELETE",
		reqURI:         "https://localhost/hsm/v2/sysinfo/powermaps/0c0s27b0n0",
		hmsdsDidDelete: false,
		hmsdsRespErr:   hmsds.ErrHMSDSArgBadID,
		expectedID:     "x0c0s27b0n0",
		expectedResp:   json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"invalid xname","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.DeletePowerMapByID.Input.id = ""
		results.DeletePowerMapByID.Return.changed = test.hmsdsDidDelete
		results.DeletePowerMapByID.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.hmsdsDidDelete && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if w.Code != http.StatusBadRequest && test.expectedID != results.DeletePowerMapByID.Input.id {
			t.Errorf("Test %v Failed: Expected comp '%v'; Received comp '%v'", i, test.expectedID, results.DeletePowerMapByID.Input.id)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPowerMapDeleteAll(t *testing.T) {
	tests := []struct {
		reqType       string
		reqURI        string
		hmsdsRespErr  error
		expectedCount int64
		expectedResp  []byte
	}{{
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/sysinfo/powermaps",
		hmsdsRespErr:  nil,
		expectedCount: 3,
		expectedResp:  json.RawMessage(`{"code":0,"message":"deleted 3 entries"}` + "\n"),
	}, {
		reqType:       "DELETE",
		reqURI:        "https://localhost/hsm/v2/sysinfo/powermaps",
		hmsdsRespErr:  nil,
		expectedCount: 0,
		expectedResp:  json.RawMessage(`{"type":"about:blank","title":"Not Found","detail":"no entries to delete","status":404}` + "\n"),
	}}

	for i, test := range tests {
		results.DeletePowerMapsAll.Return.numRows = test.expectedCount
		results.DeletePowerMapsAll.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, nil)
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.expectedCount != 0 && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if w.Code != http.StatusBadRequest && test.expectedCount != results.DeletePowerMapsAll.Return.numRows {
			t.Errorf("Test %v Failed: Expected count '%v'; Received count '%v'", i, test.expectedCount, results.DeletePowerMapsAll.Return.numRows)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPowerMapPost(t *testing.T) {
	tests := []struct {
		reqType      string
		reqURI       string
		reqBody      []byte
		hmsdsRespErr error
		expectedResp []byte
	}{{
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      json.RawMessage(`[{"id": "x0c0s0b0n0", "poweredBy":["x0m0p0j10","x0m0p0j11"]}]`),
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"code":0,"message":"Created or modified 1 entries"}` + "\n"),
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      json.RawMessage(`[{"id": "x0c0s21b0n0", "poweredBy":["x0m0p0j10","x0m0p0j11"]}, { "ID": "x0c0s22b0n0", "poweredBy":["x0m0p0j12","x0m0p0j13"]}]`),
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"code":0,"message":"Created or modified 2 entries"}` + "\n"),
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      json.RawMessage(`[]`),
		hmsdsRespErr: nil,
		expectedResp: json.RawMessage(`{"code":0,"message":"Created or modified 0 entries"}` + "\n"),
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      nil,
		hmsdsRespErr: hmsds.ErrHMSDSArgMissing,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"error decoding JSON unexpected end of JSON input","status":500}` + "\n"),
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      json.RawMessage(`[{"id": "x0c0s0b0n0", "poweredBy":["x0m0p0j10","x0m0p0j11"]}]`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Internal Server Error","detail":"operation 'POST' failed during store. ","status":500}` + "\n"),
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      json.RawMessage(`[{"id": "x0c0s0b0n0", "poweredBy":["foo","x0m0p0j11"]}]`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate map data at idx 0: Power supply xname ID 'foo' is invalid","status":400}` + "\n"),
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      json.RawMessage(`[{"id": "foo", "poweredBy":["x0m0p0j10","x0m0p0j11"]}]`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate map data at idx 0: xname ID 'foo' is invalid","status":400}` + "\n"),
	}, {
		reqType:      "POST",
		reqURI:       "https://localhost/hsm/v2/sysinfo/powermaps",
		reqBody:      json.RawMessage(`[{"id": "x0c0s0b0n0"}]`),
		hmsdsRespErr: hmsds.ErrHMSDSArgNoMatch,
		expectedResp: json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"poweredby is required for PowerMaps","status":400}` + "\n"),
	}}
	for i, test := range tests {
		// Clear the results
		results.InsertPowerMaps.Input.ms = nil
		results.InsertPowerMaps.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if test.hmsdsRespErr != nil && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}

func TestDoPowerMapPut(t *testing.T) {
	powerMap, _ := sm.NewPowerMap("x0c0s27b0n0", []string{"x0m0p0j10", "x0m0p0j11"})
	tests := []struct {
		reqType          string
		reqURI           string
		reqBody          []byte
		expectedPowerMap *sm.PowerMap
		hmsdsRespErr     error
		expectedResp     []byte
	}{{
		reqType:          "PUT",
		reqURI:           "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27b0n0",
		reqBody:          json.RawMessage(`{"poweredBy": ["x0m0p0j10","x0m0p0j11"]}`),
		expectedPowerMap: powerMap,
		hmsdsRespErr:     nil,
		expectedResp:     json.RawMessage(`{"id":"x0c0s27b0n0","poweredBy":["x0m0p0j10","x0m0p0j11"]}` + "\n"),
	}, {
		reqType:          "PUT",
		reqURI:           "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27b0n0",
		reqBody:          json.RawMessage(`{}`),
		expectedPowerMap: &sm.PowerMap{},
		hmsdsRespErr:     hmsds.ErrHMSDSArgNoMatch,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"poweredby is required in PUT body","status":400}` + "\n"),
	}, {
		reqType:          "PUT",
		reqURI:           "https://localhost/hsm/v2/sysinfo/powermaps/0c0s27b0n0",
		reqBody:          json.RawMessage(`{"poweredBy": ["x0m0p0j10","x0m0p0j11"]}`),
		expectedPowerMap: &sm.PowerMap{},
		hmsdsRespErr:     hmsds.ErrHMSDSArgBadID,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"couldn't validate PowerMap data: xname ID '0c0s27b0n0' is invalid","status":400}` + "\n"),
	}, {
		reqType:          "PUT",
		reqURI:           "https://localhost/hsm/v2/sysinfo/powermaps/x0c0s27b0n0",
		reqBody:          json.RawMessage(`{"id": "x0c0s27b0n2", "poweredBy": ["x0m0p0j10","x0m0p0j11"]}`),
		expectedPowerMap: &sm.PowerMap{},
		hmsdsRespErr:     hmsds.ErrHMSDSArgBadID,
		expectedResp:     json.RawMessage(`{"type":"about:blank","title":"Bad Request","detail":"xname in URL and PUT body do not match","status":400}` + "\n"),
	}}

	for i, test := range tests {
		results.InsertPowerMap.Input.m = &sm.PowerMap{}
		results.InsertPowerMap.Return.err = test.hmsdsRespErr
		req, err := http.NewRequest(test.reqType, test.reqURI, bytes.NewBuffer(test.reqBody))
		if err != nil {
			t.Fatalf("an error '%s' was not expected while creating request", err)
		}
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		if test.hmsdsRespErr == nil && test.expectedPowerMap != nil && w.Code != http.StatusOK {
			t.Errorf("Response code was %v; want 200", w.Code)
		} else if (test.hmsdsRespErr != nil || test.expectedPowerMap == nil) && w.Code == http.StatusOK {
			t.Errorf("Response code was %v; expected an error", w.Code)
		}

		if test.expectedPowerMap.ID != results.InsertPowerMap.Input.m.ID ||
			len(test.expectedPowerMap.PoweredBy) != len(results.InsertPowerMap.Input.m.PoweredBy) {
			t.Errorf("Test %v Failed: Expected PowerMap '%v'; Received PowerMap '%v'", i, test.expectedPowerMap, results.InsertPowerMap.Input.m)
		} else {
			for i, xname := range test.expectedPowerMap.PoweredBy {
				if xname != results.InsertPowerMap.Input.m.PoweredBy[i] {
					t.Errorf("Test %v Failed: Expected PowerMap '%v'; Received PowerMap '%v'", i, test.expectedPowerMap, results.InsertPowerMap.Input.m)
				}
			}
		}

		if bytes.Compare(test.expectedResp, w.Body.Bytes()) != 0 {
			t.Errorf("Test %v Failed: Expected body is '%v'; Received '%v'", i, string(test.expectedResp), w.Body)
		}
	}
}
