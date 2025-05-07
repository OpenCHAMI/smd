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
	"net/http"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func sendJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil && code != http.StatusNoContent {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			fmt.Printf("Couldn't encode JSON: %s\n", err)
		}
	}
}

func sendJsonError(w http.ResponseWriter, code int, msg string) {
	if code < 400 {
		sendJSON(w, code, Response{0, msg})
	} else {
		base.SendProblemDetailsGeneric(w, code, msg)
	}
}

func sendJsonResponse(w http.ResponseWriter, code int, msg string) {
	sendJSON(w, code, Response{code, msg})
}

func sendJsonDBError(w http.ResponseWriter, prefix, internalErr string, err error) {
	if base.IsHMSError(err) {
		sendJsonError(w, http.StatusBadRequest, prefix+err.Error())
	} else {
		if internalErr != "" {
			sendJsonError(w, http.StatusInternalServerError, internalErr)
		} else {
			sendJsonError(w, http.StatusInternalServerError, "failed to query DB.")
		}
	}
}

func sendResourceArray(w http.ResponseWriter, code int, collectionURI string, uris []*sm.ResourceURI) {
	if len(uris) == 1 {
		w.Header().Set("Location", uris[0].URI)
	} else if len(uris) > 1 {
		w.Header().Set("Location", collectionURI)
	}
	sendJSON(w, code, uris)
}

func sendResource(w http.ResponseWriter, code int, uri *sm.ResourceURI) {
	if uri != nil {
		w.Header().Set("Location", uri.URI)
	}
	sendJSON(w, code, uri)
}

func sendJsonNewResourceIDArray(w http.ResponseWriter, collectionURI string, uris []*sm.ResourceURI) {
	if len(uris) == 0 {
		sendJSON(w, http.StatusNoContent, nil)
	} else {
		sendResourceArray(w, http.StatusCreated, collectionURI, uris)
	}
}

func sendJsonNewResourceID(w http.ResponseWriter, uri *sm.ResourceURI) {
	if uri == nil {
		sendJSON(w, http.StatusNoContent, nil)
	} else {
		sendResource(w, http.StatusCreated, uri)
	}
}

func sendJsonResourceIDArray(w http.ResponseWriter, uris []*sm.ResourceURI) {
	if len(uris) == 0 {
		sendJSON(w, http.StatusNoContent, nil)
	} else {
		sendJSON(w, http.StatusOK, uris)
	}
}

func sendJsonObject(w http.ResponseWriter, code int, obj interface{}) {
	if obj == nil {
		sendJSON(w, http.StatusNoContent, nil)
	} else {
		sendJSON(w, code, obj)
	}
}

func sendJsonCompRsp(w http.ResponseWriter, comp *base.Component) {
	sendJsonObject(w, http.StatusOK, comp)
}

func sendJsonCompArrayRsp(w http.ResponseWriter, comps *base.ComponentArray) {
	sendJsonObject(w, http.StatusOK, comps)
}

func sendJsonNodeMapRsp(w http.ResponseWriter, m *sm.NodeMap) {
	sendJsonObject(w, http.StatusOK, m)
}

func sendJsonNodeMapArrayRsp(w http.ResponseWriter, nnms *sm.NodeMapArray) {
	sendJsonObject(w, http.StatusOK, nnms)
}

func sendJsonHWInvByLocRsp(w http.ResponseWriter, hl *sm.HWInvByLoc) {
	sendJsonObject(w, http.StatusOK, hl)
}

func sendJsonHWInvByLocsRsp(w http.ResponseWriter, hl []*sm.HWInvByLoc) {
	sendJsonObject(w, http.StatusOK, hl)
}

func sendJsonHWInvByFRURsp(w http.ResponseWriter, hf *sm.HWInvByFRU) {
	sendJsonObject(w, http.StatusOK, hf)
}

func sendJsonHWInvByFRUsRsp(w http.ResponseWriter, hf []*sm.HWInvByFRU) {
	sendJsonObject(w, http.StatusOK, hf)
}

func sendJsonSystemHWInvRsp(w http.ResponseWriter, hw *sm.SystemHWInventory) {
	sendJsonObject(w, http.StatusOK, hw)
}

func sendJsonHWInvHistRsp(w http.ResponseWriter, hh *sm.HWInvHistArray) {
	sendJsonObject(w, http.StatusOK, hh)
}

func sendJsonHWInvHistArrayRsp(w http.ResponseWriter, hhs *sm.HWInvHistResp) {
	sendJsonObject(w, http.StatusOK, hhs)
}

func sendJsonRFEndpointRsp(w http.ResponseWriter, ep *sm.RedfishEndpoint) {
	sendJsonObject(w, http.StatusOK, ep)
}

func sendJsonRFEndpointArrayRsp(w http.ResponseWriter, eps *sm.RedfishEndpointArray) {
	sendJsonObject(w, http.StatusOK, eps)
}

func sendJsonCompEndpointRsp(w http.ResponseWriter, cep *sm.ComponentEndpoint) {
	sendJsonObject(w, http.StatusOK, cep)
}

func sendJsonCompEndpointArrayRsp(w http.ResponseWriter, ceps *sm.ComponentEndpointArray) {
	sendJsonObject(w, http.StatusOK, ceps)
}

func sendJsonServiceEndpointRsp(w http.ResponseWriter, sep *sm.ServiceEndpoint) {
	sendJsonObject(w, http.StatusOK, sep)
}

func sendJsonServiceEndpointArrayRsp(w http.ResponseWriter, seps *sm.ServiceEndpointArray) {
	sendJsonObject(w, http.StatusOK, seps)
}

func sendJsonCompEthInterfaceRsp(w http.ResponseWriter, cei *sm.CompEthInterface) {
	sendJsonObject(w, http.StatusOK, cei)
}

func sendJsonCompEthInterfaceArrayRsp(w http.ResponseWriter, ceis []*sm.CompEthInterface) {
	sendJsonObject(w, http.StatusOK, ceis)
}

func sendJsonCompEthInterfaceV2Rsp(w http.ResponseWriter, cei *sm.CompEthInterfaceV2) {
	sendJsonObject(w, http.StatusOK, cei)
}

func sendJsonCompEthInterfaceV2ArrayRsp(w http.ResponseWriter, ceis []*sm.CompEthInterfaceV2) {
	sendJsonObject(w, http.StatusOK, ceis)
}

func sendJsonCompEthInterfaceIPAddressMappingsRsp(w http.ResponseWriter, ipm *sm.IPAddressMapping) {
	sendJsonObject(w, http.StatusOK, ipm)
}

func sendJsonCompEthInterfaceIPAddressMappingsArrayRsp(w http.ResponseWriter, ceis []sm.IPAddressMapping) {
	sendJsonObject(w, http.StatusOK, ceis)
}

func sendJsonDiscoveryStatusRsp(w http.ResponseWriter, stat *sm.DiscoveryStatus) {
	sendJsonObject(w, http.StatusOK, stat)
}

func sendJsonDiscoveryStatusArrayRsp(w http.ResponseWriter, stats []*sm.DiscoveryStatus) {
	sendJsonObject(w, http.StatusOK, stats)
}

func sendJsonSCNSubscriptionArrayRsp(w http.ResponseWriter, subs *sm.SCNSubscriptionArray) {
	sendJsonObject(w, http.StatusOK, subs)
}

func sendJsonSCNSubscriptionRsp(w http.ResponseWriter, sub *sm.SCNSubscription) {
	sendJsonObject(w, http.StatusOK, sub)
}

func sendJsonGroupArrayRsp(w http.ResponseWriter, groups *[]sm.Group) {
	sendJsonObject(w, http.StatusOK, groups)
}

func sendJsonGroupRsp(w http.ResponseWriter, group *sm.Group) {
	sendJsonObject(w, http.StatusOK, group)
}

func sendJsonStringArrayRsp(w http.ResponseWriter, strs *[]string) {
	sendJsonObject(w, http.StatusOK, strs)
}

func sendJsonMembersRsp(w http.ResponseWriter, members *sm.Members) {
	sendJsonObject(w, http.StatusOK, members)
}

func sendJsonPartitionArrayRsp(w http.ResponseWriter, parts *[]sm.Partition) {
	sendJsonObject(w, http.StatusOK, parts)
}

func sendJsonPartitionRsp(w http.ResponseWriter, part *sm.Partition) {
	sendJsonObject(w, http.StatusOK, part)
}

func sendJsonMembershipArrayRsp(w http.ResponseWriter, memberships []*sm.Membership) {
	sendJsonObject(w, http.StatusOK, memberships)
}

func sendJsonMembershipRsp(w http.ResponseWriter, membership *sm.Membership) {
	sendJsonObject(w, http.StatusOK, membership)
}

func sendJsonCompLockV2Rsp(w http.ResponseWriter, cls sm.CompLockV2Status) {
	sendJsonObject(w, http.StatusOK, cls)
}

func sendJsonCompLockV2UpdateRsp(w http.ResponseWriter, clu sm.CompLockV2UpdateResult) {
	sendJsonObject(w, http.StatusOK, clu)
}

func sendJsonCompReservationRsp(w http.ResponseWriter, crs sm.CompLockV2ReservationResult) {
	sendJsonObject(w, http.StatusOK, crs)
}

func sendJsonPowerMapRsp(w http.ResponseWriter, m *sm.PowerMap) {
	sendJsonObject(w, http.StatusOK, m)
}

func sendJsonPowerMapArrayRsp(w http.ResponseWriter, ms []*sm.PowerMap) {
	sendJsonObject(w, http.StatusOK, ms)
}

func sendJsonValueRsp(w http.ResponseWriter, vals *HMSValues) {
	sendJsonObject(w, http.StatusOK, vals)
}
