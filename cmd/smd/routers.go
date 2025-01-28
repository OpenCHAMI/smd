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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/handlers"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func (s *SmD) NewRouter(publicRoutes []Route, protectedRoutes []Route) *chi.Mux {
	// create router and use recommended middleware
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.StripSlashes)
	// todo resolve openchami source usage
	// router.Use(openchami_logger.OpenCHAMILogger(logger))
	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Logger(http.NotFoundHandler(), "NotFoundHandler")
	}))

	router.Use(middleware.Timeout(60 * time.Second))
	// todo make jwksURL check and changes

	routes := append(publicRoutes, protectedRoutes...)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		if s.lgLvl >= LOG_DEBUG ||
			(!strings.Contains(route.Name, "doReadyGet") &&
				!strings.Contains(route.Name, "doLivenessGet")) {
			handler = handlers.CombinedLoggingHandler(os.Stdout, handler)
		}
		router.Method(
			route.Method,
			route.Pattern,
			handler,
		)
	}

	router.MethodNotAllowed(http.HandlerFunc(s.doMethodNotAllowedHandler))
	s.router = router
	return router
}

func (s *SmD) getAllMethodsForRequest(req *http.Request) []string {
	var methods []string
	smdRoutes := s.generateRoutes()
	path := req.URL.Path
	for _, smdRoute := range smdRoutes {
		// toto match patterns
		// for example be able to match a url to the pattern: /hsm/v2/State/Components/{xname}
		if strings.EqualFold(path, smdRoute.Pattern) {
			methods = append(methods, smdRoute.Method)
		}
	}
	return methods
}

func (s *SmD) doMethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	allowString := ""
	//need to dynamically build the list of HTTP methods supported for the failed request
	allMethods := s.getAllMethodsForRequest(r)
	if len(allMethods) > 0 {
		allowString = strings.Join(allMethods[:], ",")
	}
	if allowString != "" {
		w.Header().Set("allow", allowString)
	}
	sendJsonError(w, http.StatusMethodNotAllowed, "allow "+allowString)
}

func (s *SmD) generatePublicRoutes() Routes {
	return Routes{

		///////////////////////////////////////////////////////////////////////
		// v2 API routes
		///////////////////////////////////////////////////////////////////////

		// HSM Service State
		Route{
			"doReadyGetV2",
			strings.ToUpper("Get"),
			s.serviceBaseV2 + "/ready",
			s.doReadyGet,
		},
		Route{
			"doLivenessGetV2",
			strings.ToUpper("Get"),
			s.serviceBaseV2 + "/liveness",
			s.doLivenessGet,
		},
		Route{
			"doValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2,
			s.doValuesGet,
		},
		Route{
			"doArchValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/arch",
			s.doArchValuesGet,
		},
		Route{
			"doClassValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/class",
			s.doClassValuesGet,
		},
		Route{
			"doFlagValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/flag",
			s.doFlagValuesGet,
		},
		Route{
			"doNetTypeValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/nettype",
			s.doNetTypeValuesGet,
		},
		Route{
			"doRoleValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/role",
			s.doRoleValuesGet,
		},
		Route{
			"doSubRoleValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/subrole",
			s.doSubRoleValuesGet,
		},
		Route{
			"doStateValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/state",
			s.doStateValuesGet,
		},
		Route{
			"doTypeValuesGetV2",
			strings.ToUpper("Get"),
			s.valuesBaseV2 + "/type",
			s.doTypeValuesGet,
		},
		// other
		Route{
			"doComponentsGetV2",
			strings.ToUpper("Get"),
			s.componentsBaseV2,
			s.doComponentsGet,
		},
		Route{
			"doCompEthInterfacesGetV2",
			strings.ToUpper("Get"),
			s.compEthIntBaseV2,
			s.doCompEthInterfacesGetV2,
		},
	}
}

func (s *SmD) generateProtectedRoutes() Routes {
	return Routes{

		// Components
		Route{
			"doComponentGetV2",
			strings.ToUpper("Get"),
			s.componentsBaseV2 + "/{xname}",
			s.doComponentGet,
		},
		Route{
			"doComponentPutV2",
			strings.ToUpper("Put"),
			s.componentsBaseV2 + "/{xname}",
			s.doComponentPut,
		},
		Route{
			"doComponentDeleteV2",
			strings.ToUpper("Delete"),
			s.componentsBaseV2 + "/{xname}",
			s.doComponentDelete,
		},
		Route{
			"doComponentsPostV2",
			strings.ToUpper("Post"),
			s.componentsBaseV2,
			s.doComponentsPost,
		},
		Route{
			"doComponentsDeleteAllV2",
			strings.ToUpper("Delete"),
			s.componentsBaseV2,
			s.doComponentsDeleteAll,
		},
		Route{
			"doCompBulkStateDataPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/BulkStateData",
			s.doCompBulkStateDataPatch,
		},
		Route{
			"doCompStateDataPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/{xname}/StateData",
			s.doCompStateDataPatch,
		},
		Route{
			"doCompBulkFlagOnlyPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/BulkFlagOnly",
			s.doCompBulkFlagOnlyPatch,
		},
		Route{
			"doCompFlagOnlyPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/{xname}/FlagOnly",
			s.doCompFlagOnlyPatch,
		},
		Route{
			"doCompBulkEnabledPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/BulkEnabled",
			s.doCompBulkEnabledPatch,
		},
		Route{
			"doCompEnabledV2",
			"PATCH",
			s.componentsBaseV2 + "/{xname}/Enabled",
			s.doCompEnabledPatch,
		},
		Route{
			"doCompBulkSwStatusPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/BulkSoftwareStatus",
			s.doCompBulkSwStatusPatch,
		},
		Route{
			"doCompSwStatusV2",
			"PATCH",
			s.componentsBaseV2 + "/{xname}/SoftwareStatus",
			s.doCompSwStatusPatch,
		},
		Route{
			"doCompBulkRolePatchV2",
			"PATCH",
			s.componentsBaseV2 + "/BulkRole",
			s.doCompBulkRolePatch,
		},
		Route{
			"doCompRoleV2",
			"PATCH",
			s.componentsBaseV2 + "/{xname}/Role",
			s.doCompRolePatch,
		},
		Route{
			"doCompBulkNIDPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/BulkNID",
			s.doCompBulkNIDPatch,
		},
		Route{
			"doCompNIDPatchV2",
			"PATCH",
			s.componentsBaseV2 + "/{xname}/NID",
			s.doCompNIDPatch,
		},
		Route{
			"doComponentByNIDGetV2",
			strings.ToUpper("Get"),
			s.componentsBaseV2 + "/ByNID/{nid}",
			s.doComponentByNIDGet,
		},
		Route{
			"doComponentByNIDQueryPostV2",
			strings.ToUpper("Post"),
			s.componentsBaseV2 + "/ByNID/Query",
			s.doComponentByNIDQueryPost,
		},
		Route{
			"doComponentsQueryPostV2",
			strings.ToUpper("Post"),
			s.componentsBaseV2 + "/Query",
			s.doComponentsQueryPost,
		},
		Route{
			"doComponentsQueryGetV2",
			strings.ToUpper("Get"),
			s.componentsBaseV2 + "/Query/{xname}",
			s.doComponentsQueryGet,
		},

		// ComponentEndpoints
		Route{
			"doComponentEndpointGetV2", // Individual entry
			strings.ToUpper("Get"),
			s.compEPBaseV2 + "/{xname}",
			s.doComponentEndpointGet,
		},
		Route{
			"doComponentEndpointDeleteV2", // Individual entry
			strings.ToUpper("Delete"),
			s.compEPBaseV2 + "/{xname}",
			s.doComponentEndpointDelete,
		},
		Route{
			"doComponentEndpointsGetV2", // Whole collection
			strings.ToUpper("Get"),
			s.compEPBaseV2,
			s.doComponentEndpointsGet,
		},
		Route{
			"doComponentEndpointsDeleteAllV2", // Whole collection
			strings.ToUpper("Delete"),
			s.compEPBaseV2,
			s.doComponentEndpointsDeleteAll,
		},
		//Route{
		//	"doComponentEndpointQueryGetV2",
		//	strings.ToUpper("Get"),
		//	s.compEPBaseV2 + "/Query/{xname}",
		//	s.doComponentEndpointQueryGet,
		//},

		// ServiceEndpoints
		Route{
			"doServiceEndpointGetV2", // Individual entry
			strings.ToUpper("Get"),
			s.serviceEPBaseV2 + "/{service}/RedfishEndpoints/{xname}",
			s.doServiceEndpointGet,
		},
		Route{
			"doServiceEndpointDeleteV2", // Individual entry
			strings.ToUpper("Delete"),
			s.serviceEPBaseV2 + "/{service}/RedfishEndpoints/{xname}",
			s.doServiceEndpointDelete,
		},
		Route{
			"doServiceEndpointsGetV2", // Collection by service
			strings.ToUpper("Get"),
			s.serviceEPBaseV2 + "/{service}",
			s.doServiceEndpointsGet,
		},
		Route{
			"doServiceEndpointsGetAllV2", // Whole collection
			strings.ToUpper("Get"),
			s.serviceEPBaseV2,
			s.doServiceEndpointsGetAll,
		},
		Route{
			"doServiceEndpointsDeleteAllV2", // Whole collection
			strings.ToUpper("Delete"),
			s.serviceEPBaseV2,
			s.doServiceEndpointsDeleteAll,
		},

		// Component Ethernet Interfaces - V2
		Route{
			"doCompEthInterfacePostV2",
			strings.ToUpper("Post"),
			s.compEthIntBaseV2,
			s.doCompEthInterfacePostV2,
		},
		Route{
			"doCompEthInterfaceDeleteAllV2",
			strings.ToUpper("Delete"),
			s.compEthIntBaseV2,
			s.doCompEthInterfaceDeleteAll,
		},
		Route{
			"doCompEthInterfaceGetV2",
			strings.ToUpper("Get"),
			s.compEthIntBaseV2 + "/{id}",
			s.doCompEthInterfaceGetV2,
		},
		Route{
			"doCompEthInterfaceDeleteV2",
			strings.ToUpper("Delete"),
			s.compEthIntBaseV2 + "/{id}",
			s.doCompEthInterfaceDelete,
		},
		Route{
			"doCompEthInterfacePatchV2",
			strings.ToUpper("Patch"),
			s.compEthIntBaseV2 + "/{id}",
			s.doCompEthInterfacePatchV2,
		},
		Route{
			"doCompEthInterfaceIPAddressesGetV2",
			strings.ToUpper("Get"),
			s.compEthIntBaseV2 + "/{id}/IPAddresses",
			s.doCompEthInterfaceIPAddressesGetV2,
		},
		Route{
			"doCompEthInterfaceIPAddressPostV2",
			strings.ToUpper("Post"),
			s.compEthIntBaseV2 + "/{id}/IPAddresses",
			s.doCompEthInterfaceIPAddressPostV2,
		},
		Route{
			"doCompEthInterfaceMemberPatchV2",
			strings.ToUpper("Patch"),
			s.compEthIntBaseV2 + "/{id}/IPAddresses/{ipaddr}",
			s.doCompEthInterfaceIPAddressPatchV2,
		},
		Route{
			"doCompEthInterfaceMemberDeleteV2",
			strings.ToUpper("Delete"),
			s.compEthIntBaseV2 + "/{id}/IPAddresses/{ipaddr}",
			s.doCompEthInterfaceIPAddressDeleteV2,
		},

		// NodeMaps
		Route{
			"doNodeMapGetV2",
			strings.ToUpper("Get"),
			s.nodeMapBaseV2 + "/{xname}",
			s.doNodeMapGet,
		},
		Route{
			"doNodeMapsGetV2",
			strings.ToUpper("Get"),
			s.nodeMapBaseV2,
			s.doNodeMapsGet,
		},
		Route{
			"doNodeMapDeleteV2",
			strings.ToUpper("Delete"),
			s.nodeMapBaseV2 + "/{xname}",
			s.doNodeMapDelete,
		},
		Route{
			"doNodeMapPutV2",
			strings.ToUpper("Put"),
			s.nodeMapBaseV2 + "/{xname}",
			s.doNodeMapPut,
		},
		Route{
			"doNodeMapsPostV2",
			strings.ToUpper("Post"),
			s.nodeMapBaseV2,
			s.doNodeMapsPost,
		},
		Route{
			"doNodeMapsDeleteAllV2",
			strings.ToUpper("Delete"),
			s.nodeMapBaseV2,
			s.doNodeMapsDeleteAll,
		},

		// Hardware Inventory History
		Route{
			"doHWInvHistByLocationGetV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "/History/{xname}",
			s.doHWInvHistByLocationGet,
		},
		Route{
			"doHWInvHistByLocationGetAllV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "/History",
			s.doHWInvHistByLocationGetAll,
		},
		Route{
			"doHWInvHistByFRUGetV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "ByFRU/History/{fruid}",
			s.doHWInvHistByFRUGet,
		},
		Route{
			"doHWInvHistByFRUGetAllV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "ByFRU/History",
			s.doHWInvHistByFRUGetAll,
		},
		Route{
			"doHWInvHistByLocationDeleteV2",
			strings.ToUpper("Delete"),
			s.hwinvByLocBaseV2 + "/History/{xname}",
			s.doHWInvHistByLocationDelete,
		},
		Route{
			"doHWInvHistDeleteAllV2",
			strings.ToUpper("Delete"),
			s.hwinvByLocBaseV2 + "/History",
			s.doHWInvHistDeleteAll,
		},
		Route{
			"doHWInvHistByFRUDeleteV2",
			strings.ToUpper("Delete"),
			s.hwinvByLocBaseV2 + "ByFRU/History/{fruid}",
			s.doHWInvHistByFRUDelete,
		},

		// Hardware Inventory
		Route{
			"doHWInvByLocationQueryGetV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "/Query/{xname}",
			s.doHWInvByLocationQueryGet,
		},
		Route{
			"doHWInvByFRUGetV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "ByFRU/{fruid}",
			s.doHWInvByFRUGet,
		},
		Route{
			"doHWInvByFRUGetAllV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "ByFRU",
			s.doHWInvByFRUGetAll,
		},
		Route{
			"doHWInvByLocationGetV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2 + "/{xname}",
			s.doHWInvByLocationGet,
		},

		Route{
			"doHWInvByLocationPostV2",
			strings.ToUpper("Post"),
			s.hwinvByLocBaseV2,
			s.doHWInvByLocationPost,
		},

		Route{
			"doHWInvByLocationGetAllV2",
			strings.ToUpper("Get"),
			s.hwinvByLocBaseV2,
			s.doHWInvByLocationGetAll,
		},
		Route{
			"doHWInvByFRUDeleteV2",
			strings.ToUpper("Delete"),
			s.hwinvByLocBaseV2 + "ByFRU/{fruid}",
			s.doHWInvByFRUDelete,
		},
		Route{
			"doHWInvByFRUDeleteAllV2",
			strings.ToUpper("Delete"),
			s.hwinvByLocBaseV2 + "ByFRU",
			s.doHWInvByFRUDeleteAll,
		},
		Route{
			"doHWInvByLocationDeleteV2",
			strings.ToUpper("Delete"),
			s.hwinvByLocBaseV2 + "/{xname}",
			s.doHWInvByLocationDelete,
		},
		Route{
			"doHWInvByLocationDeleteAllV2",
			strings.ToUpper("Delete"),
			s.hwinvByLocBaseV2,
			s.doHWInvByLocationDeleteAll,
		},

		// RefishEndpoints
		Route{
			"doRedfishEndpointGetV2",
			strings.ToUpper("Get"),
			s.redfishEPBaseV2 + "/{xname}",
			s.doRedfishEndpointGet,
		},
		Route{
			"doRedfishEndpointsGetV2",
			strings.ToUpper("Get"),
			s.redfishEPBaseV2,
			s.doRedfishEndpointsGet,
		},
		Route{
			"doRedfishEndpointDeleteV2",
			strings.ToUpper("Delete"),
			s.redfishEPBaseV2 + "/{xname}",
			s.doRedfishEndpointDelete,
		},
		Route{
			"doRedfishEndpointPutV2",
			strings.ToUpper("Put"),
			s.redfishEPBaseV2 + "/{xname}",
			s.doRedfishEndpointPut,
		},
		Route{
			"doRedfishEndpointPatchV2",
			strings.ToUpper("Patch"),
			s.redfishEPBaseV2 + "/{xname}",
			s.doRedfishEndpointPatch,
		},
		Route{
			"doRedfishEndpointsPostV2",
			strings.ToUpper("Post"),
			s.redfishEPBaseV2,
			s.doRedfishEndpointsPost,
		},
		Route{
			"doRedfishEndpointsDeleteAllV2",
			strings.ToUpper("Delete"),
			s.redfishEPBaseV2,
			s.doRedfishEndpointsDeleteAll,
		},
		Route{
			"doRedfishEndpointQueryGetV2",
			strings.ToUpper("Get"),
			s.redfishEPBaseV2 + "/Query/{xname}",
			s.doRedfishEndpointQueryGet,
		},
		Route{
			"doInventoryDiscoverPostV2",
			strings.ToUpper("Post"),
			s.invDiscoverBaseV2,
			s.doInventoryDiscoverPost,
		},
		Route{
			"doDiscoveryStatusGetAllV2",
			strings.ToUpper("Get"),
			s.invDiscStatusBaseV2,
			s.doDiscoveryStatusGetAll,
		},
		Route{
			"doDiscoveryStatusGetV2",
			strings.ToUpper("Get"),
			s.invDiscStatusBaseV2 + "/{id}",
			s.doDiscoveryStatusGet,
		},

		Route{
			"doGetSCNSubscriptionV2",
			strings.ToUpper("Get"),
			s.subscriptionBaseV2 + "/SCN",
			s.doGetSCNSubscriptionsAll,
		},
		Route{
			"doPostSCNSubscriptionV2",
			strings.ToUpper("Post"),
			s.subscriptionBaseV2 + "/SCN",
			s.doPostSCNSubscription,
		},
		Route{
			"doDeleteSCNSubscriptionsV2",
			strings.ToUpper("Delete"),
			s.subscriptionBaseV2 + "/SCN",
			s.doDeleteSCNSubscriptionsAll,
		},
		Route{
			"doGetSCNSubscriptionV2",
			strings.ToUpper("Get"),
			s.subscriptionBaseV2 + "/SCN/{id}",
			s.doGetSCNSubscription,
		},
		Route{
			"doPutSCNSubscriptionV2",
			strings.ToUpper("Put"),
			s.subscriptionBaseV2 + "/SCN/{id}",
			s.doPutSCNSubscription,
		},
		Route{
			"doPatchSCNSubscriptionV2",
			strings.ToUpper("Patch"),
			s.subscriptionBaseV2 + "/SCN/{id}",
			s.doPatchSCNSubscription,
		},
		Route{
			"doDeleteSCNSubscriptionV2",
			strings.ToUpper("Delete"),
			s.subscriptionBaseV2 + "/SCN/{id}",
			s.doDeleteSCNSubscription,
		},

		// Groups
		Route{
			"doGroupsGetV2",
			strings.ToUpper("Get"),
			s.groupsBaseV2,
			s.doGroupsGet,
		},
		Route{
			"doGroupsPostV2",
			strings.ToUpper("Post"),
			s.groupsBaseV2,
			s.doGroupsPost,
		},
		Route{
			"doGroupLabelsGetV2",
			strings.ToUpper("Get"),
			s.groupsBaseV2 + "/labels",
			s.doGroupLabelsGet,
		},
		Route{
			"doGroupGetV2",
			strings.ToUpper("Get"),
			s.groupsBaseV2 + "/{group_label}",
			s.doGroupGet,
		},
		Route{
			"doGroupDeleteV2",
			strings.ToUpper("Delete"),
			s.groupsBaseV2 + "/{group_label}",
			s.doGroupDelete,
		},
		Route{
			"doGroupPatchV2",
			strings.ToUpper("Patch"),
			s.groupsBaseV2 + "/{group_label}",
			s.doGroupPatch,
		},
		Route{
			"doGroupMembersGetV2",
			strings.ToUpper("Get"),
			s.groupsBaseV2 + "/{group_label}/members",
			s.doGroupMembersGet,
		},
		Route{
			"doGroupMembersPostV2",
			strings.ToUpper("Post"),
			s.groupsBaseV2 + "/{group_label}/members",
			s.doGroupMembersPost,
		},
		Route{
			"doGroupMemberDeleteV2",
			strings.ToUpper("Delete"),
			s.groupsBaseV2 + "/{group_label}/members/{xname_id}",
			s.doGroupMemberDelete,
		},

		// Partitions
		Route{
			"doPartitionsGetV2",
			strings.ToUpper("Get"),
			s.partitionsBaseV2,
			s.doPartitionsGet,
		},
		Route{
			"doPartitionsPostV2",
			strings.ToUpper("Post"),
			s.partitionsBaseV2,
			s.doPartitionsPost,
		},
		Route{
			"doPartitionNamesGetV2",
			strings.ToUpper("Get"),
			s.partitionsBaseV2 + "/names",
			s.doPartitionNamesGet,
		},
		Route{
			"doPartitionGetV2",
			strings.ToUpper("Get"),
			s.partitionsBaseV2 + "/{partition_name}",
			s.doPartitionGet,
		},
		Route{
			"doPartitionDeleteV2",
			strings.ToUpper("Delete"),
			s.partitionsBaseV2 + "/{partition_name}",
			s.doPartitionDelete,
		},
		Route{
			"doPartitionPatchV2",
			strings.ToUpper("Patch"),
			s.partitionsBaseV2 + "/{partition_name}",
			s.doPartitionPatch,
		},
		Route{
			"doPartitionMembersGetV2",
			strings.ToUpper("Get"),
			s.partitionsBaseV2 + "/{partition_name}/members",
			s.doPartitionMembersGet,
		},
		Route{
			"doPartitionMembersPostV2",
			strings.ToUpper("Post"),
			s.partitionsBaseV2 + "/{partition_name}/members",
			s.doPartitionMembersPost,
		},
		Route{
			"doPartitionMemberDeleteV2",
			strings.ToUpper("Delete"),
			s.partitionsBaseV2 + "/{partition_name}/members/{xname_id}",
			s.doPartitionMemberDelete,
		},

		// Memberships
		Route{
			"doMembershipsGetV2",
			strings.ToUpper("Get"),
			s.membershipsBaseV2,
			s.doMembershipsGet,
		},
		Route{
			"doMembershipGetV2",
			strings.ToUpper("Get"),
			s.membershipsBaseV2 + "/{xname}",
			s.doMembershipGet,
		},

		// V2 Component Locks

		//Admin reservedMap
		Route{
			"doCompLocksReservationRemoveV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/reservations/remove",
			s.doCompLocksReservationRemove,
		},
		Route{
			"doCompLocksReservationReleaseV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/reservations/release",
			s.doCompLocksReservationRelease,
		},
		Route{
			"doCompLocksReservationCreateV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/reservations",
			s.doCompLocksReservationCreate,
		},

		//Service reservedMap
		Route{
			"doCompLocksServiceReservationRenewV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/service/reservations/renew",
			s.doCompLocksServiceReservationRenew,
		},
		Route{
			"doCompLocksServiceReservationReleaseV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/service/reservations/release",
			s.doCompLocksServiceReservationRelease,
		},
		Route{
			"doCompLocksServiceReservationCreateV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/service/reservations",
			s.doCompLocksServiceReservationCreate,
		},
		Route{
			"doCompLocksServiceReservationCheckV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/service/reservations/check",
			s.doCompLocksServiceReservationCheck,
		},

		//Admin Locks
		Route{
			"doCompLocksStatusV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/status",
			s.doCompLocksStatus,
		},
		Route{
			"doCompLocksStatusGetV2",
			strings.ToUpper("Get"),
			s.compLockBaseV2 + "/status",
			s.doCompLocksStatusGet,
		},
		Route{
			"doCompLocksLockV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/lock",
			s.doCompLocksLock,
		},
		Route{
			"doCompLocksUnlockV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/unlock",
			s.doCompLocksUnlock,
		},
		Route{
			"doCompLocksRepairV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/repair",
			s.doCompLocksRepair,
		},
		Route{
			"doCompLocksDisableV2",
			strings.ToUpper("Post"),
			s.compLockBaseV2 + "/disable",
			s.doCompLocksDisable,
		},

		// PowerMaps
		Route{
			"doPowerMapGetV2",
			strings.ToUpper("Get"),
			s.powerMapBaseV2 + "/{xname}",
			s.doPowerMapGet,
		},
		Route{
			"doPowerMapsGetV2",
			strings.ToUpper("Get"),
			s.powerMapBaseV2,
			s.doPowerMapsGet,
		},
		Route{
			"doPowerMapDeleteV2",
			strings.ToUpper("Delete"),
			s.powerMapBaseV2 + "/{xname}",
			s.doPowerMapDelete,
		},
		Route{
			"doPowerMapPutV2",
			strings.ToUpper("Put"),
			s.powerMapBaseV2 + "/{xname}",
			s.doPowerMapPut,
		},
		Route{
			"doPowerMapsPostV2",
			strings.ToUpper("Post"),
			s.powerMapBaseV2,
			s.doPowerMapsPost,
		},
		Route{
			"doPowerMapsDeleteAllV2",
			strings.ToUpper("Delete"),
			s.powerMapBaseV2,
			s.doPowerMapsDeleteAll,
		},
	}
}

func (s *SmD) generateRoutes() Routes {
	return append(s.generatePublicRoutes(), s.generateProtectedRoutes()...)
}
