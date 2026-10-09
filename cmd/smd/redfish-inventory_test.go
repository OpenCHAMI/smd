// MIT License
//
// (C) Copyright [2018-2023,2025] Hewlett Packard Enterprise Development LP
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
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/openchami/smd/v2/internal/hmsds"
	"github.com/openchami/smd/v2/pkg/sm"
)

// Embedding the datastore interface makes unexpected calls fail immediately.
// Interface insertion enforces uniqueness, unlike the general response mock.
type inventoryTestDB struct {
	hmsds.HMSDB
	interfaces   map[string]*sm.CompEthInterfaceV2
	insertIDs    []string
	deleteIDs    []string
	failID       string
	interfaceErr error
	deleteErr    error
	componentErr error
	endpointErr  error
	componentIDs []string
	endpointIDs  []string
	rfStored     bool
}

func (d *inventoryTestDB) InsertRFEndpoints(eps *sm.RedfishEndpointArray) error {
	d.rfStored = true
	return nil
}
func (d *inventoryTestDB) InsertComponent(c *base.Component) (int64, error) {
	d.componentIDs = append(d.componentIDs, c.ID)
	return 1, d.componentErr
}
func (d *inventoryTestDB) UpsertComponents(comps []*base.Component, force bool) (map[string]map[string]bool, error) {
	return nil, nil
}
func (d *inventoryTestDB) UpsertCompEndpoints(eps *sm.ComponentEndpointArray) error {
	for _, ep := range eps.ComponentEndpoints {
		d.endpointIDs = append(d.endpointIDs, ep.ID)
	}
	return d.endpointErr
}
func (d *inventoryTestDB) UpsertCompEndpoint(ep *sm.ComponentEndpoint) error {
	d.endpointIDs = append(d.endpointIDs, ep.ID)
	return d.endpointErr
}
func (d *inventoryTestDB) InsertCompEthInterface(cei *sm.CompEthInterfaceV2) error {
	d.insertIDs = append(d.insertIDs, cei.ID)
	if cei.ID == d.failID {
		return d.interfaceErr
	}
	if _, exists := d.interfaces[cei.ID]; exists {
		return hmsds.ErrHMSDSDuplicateKey
	}
	d.interfaces[cei.ID] = cei
	return nil
}
func (d *inventoryTestDB) DeleteCompEthInterfaceByID(id string) (bool, error) {
	d.deleteIDs = append(d.deleteIDs, id)
	if d.deleteErr != nil {
		return false, d.deleteErr
	}
	_, exists := d.interfaces[id]
	delete(d.interfaces, id)
	return exists, nil
}

type inventoryResponseRecorder struct {
	*httptest.ResponseRecorder
	headerWrites int
}

func (w *inventoryResponseRecorder) WriteHeader(status int) {
	w.headerWrites++
	w.ResponseRecorder.WriteHeader(status)
}

const inventoryTestPayload = `{
    "ID":"x2000c0s0b0", "Type":"NodeBMC",
    "FQDN":"x2000c0s0b0.openchami.cluster", "SchemaVersion":1,
    "Managers":[{"name":"x2000c0s0b0", "uri":"/redfish/v1/Managers/1",
        "ethernet_interfaces":[
            {"mac":"52:54:00:16:a9:2a","ip":"127.0.0.1"},
            {"mac":"52:54:00:16:a9:2b","ip":"127.0.0.2"}]}],
    "Systems":[{"uuid":"node-uuid", "uri":"/redfish/v1/Systems/1",
        "ethernet_interfaces":[{"mac":"52:54:00:62:65:ad","ip":"172.16.0.1"}]}]
}`

func inventoryTestService(db *inventoryTestDB) *SmD {
	return &SmD{db: db, lg: log.New(io.Discard, "", 0), openchami: true,
		redfishEPBaseV2: "/hsm/v2/Inventory/RedfishEndpoints"}
}

func TestRedfishInventoryPost(t *testing.T) {
	const managerID = "52540016a92a"
	const secondManagerID = "52540016a92b"
	const nodeID = "5254006265ad"
	tests := []struct {
		name          string
		conflictID    string
		failID        string
		interfaceErr  error
		componentErr  error
		endpointErr   error
		missingMAC    bool
		status        int
		inserts       []string
		endpointCount int
		detail        string
	}{
		{name: "fresh discovery", status: http.StatusCreated,
			inserts: []string{managerID, secondManagerID, nodeID}, endpointCount: 2},
		{name: "manager conflict stops before next interface", conflictID: managerID,
			status: http.StatusConflict, inserts: []string{managerID}, detail: "same MAC address"},
		{name: "node conflict stops before node endpoint", conflictID: nodeID,
			status: http.StatusConflict, inserts: []string{managerID, secondManagerID, nodeID}, endpointCount: 1, detail: "same MAC address"},
		{name: "invalid interface", missingMAC: true, status: http.StatusBadRequest},
		{name: "interface storage failure", failID: managerID, interfaceErr: errors.New("database unavailable"),
			status: http.StatusInternalServerError, inserts: []string{managerID}, detail: "operation failed during store"},
		{name: "interface argument rejected by datastore", failID: managerID, interfaceErr: sm.ErrCompEthInterfaceBadIPAddress,
			status: http.StatusBadRequest, inserts: []string{managerID}, detail: "Invalid CompEthInterface IP Address"},
		{name: "component storage failure", componentErr: errors.New("database unavailable"),
			status: http.StatusInternalServerError, detail: "failed to insert"},
		{name: "endpoint storage failure", endpointErr: errors.New("database unavailable"),
			status: http.StatusInternalServerError, inserts: []string{managerID, secondManagerID}, endpointCount: 1, detail: "failed to upsert"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &inventoryTestDB{interfaces: map[string]*sm.CompEthInterfaceV2{},
				failID: test.failID, interfaceErr: test.interfaceErr,
				componentErr: test.componentErr, endpointErr: test.endpointErr}
			if test.conflictID != "" {
				db.interfaces[test.conflictID] = &sm.CompEthInterfaceV2{ID: test.conflictID}
			}
			payload := inventoryTestPayload
			if test.missingMAC {
				payload = strings.Replace(payload, "52:54:00:16:a9:2a", "", 1)
			}
			w := &inventoryResponseRecorder{ResponseRecorder: httptest.NewRecorder()}
			inventoryTestService(db).doRedfishEndpointsPost(w,
				httptest.NewRequest(http.MethodPost, "/hsm/v2/Inventory/RedfishEndpoints", strings.NewReader(payload)))
			if w.Code != test.status {
				t.Fatalf("status=%d, want %d; body=%s", w.Code, test.status, w.Body.String())
			}
			if w.headerWrites != 1 {
				t.Errorf("wrote %d responses, want one", w.headerWrites)
			}
			if !json.Valid(w.Body.Bytes()) {
				t.Fatalf("response is not a single JSON value: %s", w.Body.String())
			}
			if !reflect.DeepEqual(db.insertIDs, test.inserts) {
				t.Errorf("inserts=%v, want %v", db.insertIDs, test.inserts)
			}
			if len(db.endpointIDs) != test.endpointCount {
				t.Errorf("endpoints=%v, want count %d", db.endpointIDs, test.endpointCount)
			}
			if !db.rfStored {
				t.Error("Redfish endpoint was not stored")
			}
			if test.status == http.StatusCreated {
				var uris []sm.ResourceURI
				if err := json.Unmarshal(w.Body.Bytes(), &uris); err != nil {
					t.Fatal(err)
				}
				if len(uris) != 1 || uris[0].URI != "/hsm/v2/Inventory/RedfishEndpoints/x2000c0s0b0" {
					t.Errorf("unexpected URIs: %v", uris)
				}
				if len(db.interfaces) != 3 {
					t.Errorf("stored %d interfaces, want 3", len(db.interfaces))
				}
			} else {
				var problem struct {
					Status int
					Detail string
				}
				if err := json.Unmarshal(w.Body.Bytes(), &problem); err != nil {
					t.Fatal(err)
				}
				if problem.Status != test.status || !strings.Contains(problem.Detail, test.detail) {
					t.Errorf("unexpected problem: %+v", problem)
				}
				if strings.Contains(w.Body.String(), `"URI"`) {
					t.Error("success URI appended to error")
				}
			}
		})
	}
}

func TestRedfishInventoryForcedUpdate(t *testing.T) {
	for _, test := range []struct {
		name       string
		deleteErr  error
		retryErr   error
		wantStatus int
	}{
		{name: "replace existing manager and node interfaces"},
		{name: "deletion failure", deleteErr: errors.New("delete failed"), wantStatus: http.StatusInternalServerError},
		{name: "reinsertion conflict", retryErr: hmsds.ErrHMSDSDuplicateKey, wantStatus: http.StatusConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			db := &inventoryTestDB{interfaces: map[string]*sm.CompEthInterfaceV2{
				"52540016a92a": {ID: "52540016a92a"}, "5254006265ad": {ID: "5254006265ad"}},
				componentErr: hmsds.ErrHMSDSDuplicateKey, deleteErr: test.deleteErr}
			if test.retryErr != nil {
				db.failID = "52540016a92a"
				db.interfaceErr = test.retryErr
			}
			err := inventoryTestService(db).parseRedfishEndpointDataV2([]byte(inventoryTestPayload), true)
			if test.wantStatus == 0 {
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(db.deleteIDs, []string{"52540016a92a", "5254006265ad"}) {
					t.Errorf("deleted interfaces: %v", db.deleteIDs)
				}
				if len(db.insertIDs) != 5 || len(db.interfaces) != 3 || len(db.endpointIDs) != 2 {
					t.Errorf("unexpected update: %+v", db)
				}
				if db.interfaces["52540016a92a"].CompID != "x2000c0s0b0" || db.interfaces["5254006265ad"].CompID != "x2000c0s0b0n0" {
					t.Error("replacement interfaces lack component information")
				}
			} else {
				if err == nil {
					t.Fatal("expected update failure")
				}
				w := &inventoryResponseRecorder{ResponseRecorder: httptest.NewRecorder()}
				sendRedfishInventoryError(w, err)
				if w.Code != test.wantStatus || w.headerWrites != 1 || !json.Valid(w.Body.Bytes()) {
					t.Errorf("unexpected response: %d %s", w.Code, w.Body.String())
				}
				if len(db.endpointIDs) != 0 {
					t.Error("continued after interface update failure")
				}
			}
		})
	}
}
