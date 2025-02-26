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
	"log"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-smd/v2/internal/hmsds"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
)

type TestResults struct {
	TestConnection struct {
		Return struct {
			err error
		}
	}
	GetComponentIDs struct {
		Funcs struct {
			getID     func(...hmsds.CompFiltFunc) string
			returnIDs func(string) ([]string, error)
		}
		Input struct {
			id string
		}
		Return struct {
			err error
		}
	}
	GetCompEndpointIDs struct {
		Funcs struct {
			getID     func(...hmsds.CompEPFiltFunc) string
			returnIDs func(string) ([]string, error)
		}
		Input struct {
			id string
		}
		Return struct {
			err error
		}
	}
	GetRFEndpointIDs struct {
		Funcs struct {
			getID     func(...hmsds.RedfishEPFiltFunc) string
			returnIDs func(string) ([]string, error)
		}
		Input struct {
			id string
		}
		Return struct {
			err error
		}
	}
	GetComponentByID struct {
		Input struct {
			id string
		}
		Return struct {
			id  *base.Component
			err error
		}
	}
	GetComponentsAll struct {
		Return struct {
			ids []*base.Component
			err error
		}
	}
	GetComponentsFilter struct {
		Input struct {
			compFilter  hmsds.ComponentFilter
			fieldFilter hmsds.FieldFilter
		}
		Return struct {
			ids []*base.Component
			err error
		}
	}
	GetComponentsQuery struct {
		Input struct {
			compFilter  hmsds.ComponentFilter
			fieldFilter hmsds.FieldFilter
			ids         []string
		}
		Return struct {
			ids []*base.Component
			err error
		}
	}
	GetComponentByNID struct {
		Input struct {
			nid string
		}
		Return struct {
			id  *base.Component
			err error
		}
	}
	InsertComponent struct {
		Input struct {
			c *base.Component
		}
		Return struct {
			rowsAffected int64
			err          error
		}
	}
	InsertComponents struct {
		Input struct {
			comps *base.ComponentArray
		}
		Return struct {
			affectedIds []string
			err         error
		}
	}
	UpsertComponents struct {
		Input struct {
			comps []*base.Component
			force bool
		}
		Return struct {
			changeMap map[string]map[string]bool
			err       error
		}
	}
	UpdateCompStates struct {
		Input struct {
			ids   []string
			state string
			flag  string
			force bool
			pi    *hmsds.PartInfo
		}
		Return struct {
			affectedIds []string
			err         error
		}
	}
	UpdateCompFlagOnly struct {
		Input struct {
			id   string
			flag string
		}
		Return struct {
			rowsAffected int64
			err          error
		}
	}
	BulkUpdateCompFlagOnly struct {
		Input struct {
			ids  []string
			flag string
		}
		Return struct {
			affectedIds []string
			err         error
		}
	}
	UpdateCompEnabled struct {
		Input struct {
			id      string
			enabled bool
		}
		Return struct {
			rowsAffected int64
			err          error
		}
	}
	BulkUpdateCompEnabled struct {
		Input struct {
			ids     []string
			enabled bool
		}
		Return struct {
			affectedIds []string
			err         error
		}
	}
	UpdateCompSwStatus struct {
		Input struct {
			id       string
			swStatus string
		}
		Return struct {
			rowsAffected int64
			err          error
		}
	}
	BulkUpdateCompSwStatus struct {
		Input struct {
			ids      []string
			swstatus string
		}
		Return struct {
			affectedIds []string
			err         error
		}
	}
	UpdateCompRole struct {
		Input struct {
			id      string
			role    string
			subRole string
		}
		Return struct {
			rowsAffected int64
			err          error
		}
	}
	BulkUpdateCompRole struct {
		Input struct {
			ids     []string
			role    string
			subRole string
		}
		Return struct {
			affectedIds []string
			err         error
		}
	}
	BulkUpdateCompClass struct {
		Input struct {
			ids   []string
			class string
		}
		Return struct {
			affectedIds []string
			err         error
		}
	}
	UpdateCompNID struct {
		Input struct {
			c *base.Component
		}
		Return struct {
			err error
		}
	}
	BulkUpdateCompNID struct {
		Input struct {
			comps *[]base.Component
		}
		Return struct {
			err error
		}
	}
	DeleteComponentByID struct {
		Input struct {
			id string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteComponentsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	// NodeMaps
	GetNodeMapByID struct {
		Input struct {
			id string
		}
		Return struct {
			entry *sm.NodeMap
			err   error
		}
	}
	GetNodeMapsAll struct {
		Return struct {
			entry []*sm.NodeMap
			err   error
		}
	}
	InsertNodeMap struct {
		Input struct {
			m *sm.NodeMap
		}
		Return struct {
			err error
		}
	}
	InsertNodeMaps struct {
		Input struct {
			nnms *sm.NodeMapArray
		}
		Return struct {
			err error
		}
	}
	DeleteNodeMapByID struct {
		Input struct {
			id string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteNodeMapsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	// PowerMaps
	GetPowerMapByID struct {
		Input struct {
			id string
		}
		Return struct {
			m   *sm.PowerMap
			err error
		}
	}
	GetPowerMapsAll struct {
		Return struct {
			ms  []*sm.PowerMap
			err error
		}
	}
	InsertPowerMap struct {
		Input struct {
			m *sm.PowerMap
		}
		Return struct {
			err error
		}
	}
	InsertPowerMaps struct {
		Input struct {
			ms []sm.PowerMap
		}
		Return struct {
			err error
		}
	}
	DeletePowerMapByID struct {
		Input struct {
			id string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeletePowerMapsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	// Hardware Inventory
	GetHWInvByLocQueryFilter struct {
		Input struct {
			f *hmsds.HWInvLocFilter
		}
		Return struct {
			hwlocs []*sm.HWInvByLoc
			err    error
		}
	}
	GetHWInvByLocFilter struct {
		Input struct {
			f *hmsds.HWInvLocFilter
		}
		Return struct {
			hwlocs []*sm.HWInvByLoc
			err    error
		}
	}
	GetHWInvByLocID struct {
		Input struct {
			id string
		}
		Return struct {
			entry *sm.HWInvByLoc
			err   error
		}
	}
	GetHWInvByLocAll struct {
		Return struct {
			entries []*sm.HWInvByLoc
			err     error
		}
	}
	GetHWInvByFRUID struct {
		Input struct {
			fruid string
		}
		Return struct {
			entry *sm.HWInvByFRU
			err   error
		}
	}
	GetHWInvByFRUFilter struct {
		Input struct {
			f *hmsds.HWInvLocFilter
		}
		Return struct {
			hwfrus []*sm.HWInvByFRU
			err    error
		}
	}
	GetHWInvByFRUAll struct {
		Return struct {
			entries []*sm.HWInvByFRU
			err     error
		}
	}
	InsertHWInvByLoc struct {
		Input struct {
			hl *sm.HWInvByLoc
		}
		Return struct {
			err error
		}
	}
	InsertHWInvByFRU struct {
		Input struct {
			hf *sm.HWInvByFRU
		}
		Return struct {
			err error
		}
	}
	InsertHWInvByLocs struct {
		Input struct {
			hls []*sm.HWInvByLoc
		}
		Return struct {
			err error
		}
	}
	DeleteHWInvByLocID struct {
		Input struct {
			id string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteHWInvByLocsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	DeleteHWInvByFRUID struct {
		Input struct {
			fruid string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteHWInvByFRUsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	// HWInv History
	GetHWInvHistFilter struct {
		Input struct {
			f *hmsds.HWInvHistFilter
		}
		Return struct {
			hwhists []*sm.HWInvHist
			err     error
		}
	}
	GetHWInvHistLastEvents struct {
		Input struct {
			ids []string
		}
		Return struct {
			hwhists []*sm.HWInvHist
			err     error
		}
	}
	InsertHWInvHist struct {
		Input struct {
			hh *sm.HWInvHist
		}
		Return struct {
			err error
		}
	}
	InsertHWInvHists struct {
		Input struct {
			hhs []*sm.HWInvHist
		}
		Return struct {
			err error
		}
	}
	DeleteHWInvHistByLocID struct {
		Input struct {
			id string
		}
		Return struct {
			numRows int64
			err     error
		}
	}
	DeleteHWInvHistByFRUID struct {
		Input struct {
			fruid string
		}
		Return struct {
			numRows int64
			err     error
		}
	}
	DeleteHWInvHistAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	DeleteHWInvHistFilter struct {
		Input struct {
			f *hmsds.HWInvHistFilter
		}
		Return struct {
			numRows int64
			err     error
		}
	}
	// Redfish Endpoints
	GetRFEndpointByID struct {
		Input struct {
			id string
		}
		Return struct {
			entry *sm.RedfishEndpoint
			err   error
		}
	}
	GetRFEndpointsAll struct {
		Return struct {
			entries []*sm.RedfishEndpoint
			err     error
		}
	}
	GetRFEndpointsFilter struct {
		Input struct {
			f *hmsds.RedfishEPFilter
		}
		Return struct {
			entries []*sm.RedfishEndpoint
			err     error
		}
	}
	InsertRFEndpoint struct {
		Input struct {
			ep *sm.RedfishEndpoint
		}
		Return struct {
			err error
		}
	}
	InsertRFEndpoints struct {
		Input struct {
			eps *sm.RedfishEndpointArray
		}
		Return struct {
			err error
		}
	}
	UpdateRFEndpoint struct {
		Input struct {
			ep *sm.RedfishEndpoint
		}
		Return struct {
			entry *sm.RedfishEndpoint
			err   error
		}
	}
	UpdateRFEndpointNoDiscInfo struct {
		Input struct {
			ep *sm.RedfishEndpoint
		}
		Return struct {
			entry       *sm.RedfishEndpoint
			affectedIds []string
			err         error
		}
	}
	PatchRFEndpointNoDiscInfo struct {
		Input struct {
			id  string
			epp sm.RedfishEndpointPatch
		}
		Return struct {
			entry       *sm.RedfishEndpoint
			affectedIds []string
			err         error
		}
	}
	UpdateRFEndpointForDiscover struct {
		Input struct {
			ids   []string
			force bool
		}
		Return struct {
			entries []*sm.RedfishEndpoint
			err     error
		}
	}
	UpdateRFEndpoints struct {
		Input struct {
			eps *sm.RedfishEndpointArray
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteRFEndpointByID struct {
		Input struct {
			id string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteRFEndpointsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	DeleteRFEndpointByIDSetEmpty struct {
		Input struct {
			id string
		}
		Return struct {
			changed     bool
			affectedIds []string
			err         error
		}
	}
	DeleteRFEndpointsAllSetEmpty struct {
		Return struct {
			numRows     int64
			affectedIds []string
			err         error
		}
	}
	// Component Endpoints
	GetCompEndpointByID struct {
		Input struct {
			id string
		}
		Return struct {
			entry *sm.ComponentEndpoint
			err   error
		}
	}
	GetCompEndpointsAll struct {
		Return struct {
			entries []*sm.ComponentEndpoint
			err     error
		}
	}
	GetCompEndpointsFilter struct {
		Input struct {
			f *hmsds.CompEPFilter
		}
		Return struct {
			entries []*sm.ComponentEndpoint
			err     error
		}
	}
	UpsertCompEndpoint struct {
		Input struct {
			cep *sm.ComponentEndpoint
		}
		Return struct {
			err error
		}
	}
	UpsertCompEndpoints struct {
		Input struct {
			ceps *sm.ComponentEndpointArray
		}
		Return struct {
			err error
		}
	}
	DeleteCompEndpointByID struct {
		Input struct {
			id string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteCompEndpointsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	DeleteCompEndpointByIDSetEmpty struct {
		Input struct {
			id string
		}
		Return struct {
			changed     bool
			affectedIds []string
			err         error
		}
	}
	DeleteCompEndpointsAllSetEmpty struct {
		Return struct {
			numRows     int64
			affectedIds []string
			err         error
		}
	}
	// Service Endpoints
	GetServiceEndpointByID struct {
		Input struct {
			svc string
			id  string
		}
		Return struct {
			entry *sm.ServiceEndpoint
			err   error
		}
	}
	GetServiceEndpointsAll struct {
		Return struct {
			entries []*sm.ServiceEndpoint
			err     error
		}
	}
	GetServiceEndpointsFilter struct {
		Input struct {
			f *hmsds.ServiceEPFilter
		}
		Return struct {
			entries []*sm.ServiceEndpoint
			err     error
		}
	}
	UpsertServiceEndpoint struct {
		Input struct {
			sep *sm.ServiceEndpoint
		}
		Return struct {
			err error
		}
	}
	UpsertServiceEndpoints struct {
		Input struct {
			seps *sm.ServiceEndpointArray
		}
		Return struct {
			err error
		}
	}
	DeleteServiceEndpointByID struct {
		Input struct {
			svc string
			id  string
		}
		Return struct {
			changed bool
			err     error
		}
	}
	DeleteServiceEndpointsAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	// Component Ethernet Interfaces
	GetCompEthInterfaceFilter struct {
		Input struct {
			f *hmsds.CompEthInterfaceFilter
		}
		Return struct {
			ceis []*sm.CompEthInterfaceV2
			err  error
		}
	}
	InsertCompEthInterface struct {
		Input struct {
			cei *sm.CompEthInterfaceV2
		}
		Return struct {
			err error
		}
	}
	InsertCompEthInterfaces struct {
		Input struct {
			ceis []*sm.CompEthInterfaceV2
		}
		Return struct {
			err error
		}
	}
	InsertCompEthInterfaceCompInfo struct {
		Input struct {
			cei *sm.CompEthInterfaceV2
		}
		Return struct {
			err error
		}
	}
	InsertCompEthInterfacesCompInfo struct {
		Input struct {
			ceis []*sm.CompEthInterfaceV2
		}
		Return struct {
			err error
		}
	}
	UpdateCompEthInterface struct {
		Input struct {
			id   string
			ceip *sm.CompEthInterfaceV2Patch
		}
		Return struct {
			cei *sm.CompEthInterfaceV2
			err error
		}
	}
	UpdateCompEthInterfaceV1 struct {
		Input struct {
			id   string
			ceip *sm.CompEthInterfacePatch
		}
		Return struct {
			cei *sm.CompEthInterfaceV2
			err error
		}
	}
	DeleteCompEthInterfaceByID struct {
		Input struct {
			id string
		}
		Return struct {
			didDelete bool
			err       error
		}
	}
	DeleteCompEthInterfacesAll struct {
		Return struct {
			numRows int64
			err     error
		}
	}
	AddCompEthInterfaceIPAddress struct {
		Input struct {
			id  string
			ipm *sm.IPAddressMapping
		}
		Return struct {
			id  string
			err error
		}
	}
	UpdateCompEthInterfaceIPAddress struct {
		Input struct {
			id       string
			ipAddr   string
			ipmPatch *sm.IPAddressMappingPatch
		}
		Output struct {
			ipm *sm.IPAddressMapping
			err error
		}
	}
	DeleteCompEthInterfaceIPAddress struct {
		Input struct {
			id     string
			ipAddr string
		}
		Output struct {
			didDelete bool
			err       error
		}
	}
	// Discovery Status
	GetDiscoveryStatusByID struct {
		Input struct {
			id uint
		}
		Return struct {
			status *sm.DiscoveryStatus
			err    error
		}
	}
	GetDiscoveryStatusAll struct {
		Return struct {
			status []*sm.DiscoveryStatus
			err    error
		}
	}
	UpsertDiscoveryStatus struct {
		Input struct {
			stat *sm.DiscoveryStatus
		}
		Return struct {
			err error
		}
	}
	// Discovery operations
	UpdateAllForRFEndpoint struct {
		Input struct {
			ep    *sm.RedfishEndpoint
			ceps  *sm.ComponentEndpointArray
			hls   []*sm.HWInvByLoc
			comps *base.ComponentArray
			seps  *sm.ServiceEndpointArray
			ceis  []*sm.CompEthInterfaceV2
		}
		Return struct {
			discoveredIds *[]base.Component
			err           error
		}
	}
	// SCN subscriptions operations
	GetSCNSubscriptionsAll struct {
		Return struct {
			subs *sm.SCNSubscriptionArray
			err  error
		}
	}
	GetSCNSubscription struct {
		Input struct {
			id int64
		}
		Return struct {
			sub *sm.SCNSubscription
			err error
		}
	}
	InsertSCNSubscription struct {
		Input struct {
			sub sm.SCNPostSubscription
		}
		Return struct {
			id  int64
			err error
		}
	}
	UpdateSCNSubscription struct {
		Input struct {
			id  int64
			sub sm.SCNPostSubscription
		}
		Return struct {
			didUpdate bool
			err       error
		}
	}
	PatchSCNSubscription struct {
		Input struct {
			id    int64
			op    string
			patch sm.SCNPatchSubscription
		}
		Return struct {
			didPatch bool
			err      error
		}
	}
	DeleteSCNSubscription struct {
		Input struct {
			id int64
		}
		Return struct {
			didDelete bool
			err       error
		}
	}
	DeleteSCNSubscriptionsAll struct {
		Return struct {
			numDelete int64
			err       error
		}
	}
	// Groups
	InsertGroup struct {
		Input struct {
			g *sm.Group
		}
		Return struct {
			label string
			err   error
		}
	}
	UpdateGroup struct {
		Input struct {
			label string
			gp    *sm.GroupPatch
		}
		Return struct {
			err error
		}
	}
	GetGroup struct {
		Input struct {
			label     string
			filt_part string
		}
		Return struct {
			group *sm.Group
			err   error
		}
	}
	GetGroupLabels struct {
		Return struct {
			labels []string
			err    error
		}
	}
	DeleteGroup struct {
		Input struct {
			label string
		}
		Return struct {
			didDelete bool
			err       error
		}
	}
	AddGroupMember struct {
		Input struct {
			label string
			id    string
		}
		Return struct {
			id  string
			err error
		}
	}
	DeleteGroupMember struct {
		Input struct {
			label string
			id    string
		}
		Return struct {
			didDelete bool
			err       error
		}
	}
	// Partitions
	InsertPartition struct {
		Input struct {
			p *sm.Partition
		}
		Return struct {
			pname string
			err   error
		}
	}
	UpdatePartition struct {
		Input struct {
			pname string
			pp    *sm.PartitionPatch
		}
		Return struct {
			err error
		}
	}
	GetPartition struct {
		Input struct {
			pname string
		}
		Return struct {
			partition *sm.Partition
			err       error
		}
	}
	GetPartitionNames struct {
		Return struct {
			pnames []string
			err    error
		}
	}
	DeletePartition struct {
		Input struct {
			pname string
		}
		Return struct {
			didDelete bool
			err       error
		}
	}
	AddPartitionMember struct {
		Input struct {
			pname string
			id    string
		}
		Return struct {
			id  string
			err error
		}
	}
	DeletePartitionMember struct {
		Input struct {
			pname string
			id    string
		}
		Return struct {
			didDelete bool
			err       error
		}
	}
	// Memberships
	GetMembership struct {
		Input struct {
			id string
		}
		Return struct {
			membership *sm.Membership
			err        error
		}
	}
	GetMemberships struct {
		Input struct {
			f *hmsds.ComponentFilter
		}
		Return struct {
			memberships []*sm.Membership
			err         error
		}
	}
	// Component Locks
	InsertCompReservations struct {
		Input struct {
			f sm.CompLockV2Filter
		}
		Return struct {
			results sm.CompLockV2ReservationResult
			err     error
		}
	}
	DeleteCompReservationsForce struct {
		Input struct {
			f sm.CompLockV2Filter
		}
		Return struct {
			results sm.CompLockV2UpdateResult
			err     error
		}
	}
	DeleteCompReservations struct {
		Input struct {
			f sm.CompLockV2ReservationFilter
		}
		Return struct {
			results sm.CompLockV2UpdateResult
			err     error
		}
	}
	DeleteCompReservationsExpired struct {
		Return struct {
			ids []string
			err error
		}
	}
	GetCompReservations struct {
		Input struct {
			dkeys []sm.CompLockV2Key
		}
		Return struct {
			results sm.CompLockV2ReservationResult
			err     error
		}
	}
	UpdateCompReservations struct {
		Input struct {
			f sm.CompLockV2ReservationFilter
		}
		Return struct {
			results sm.CompLockV2UpdateResult
			err     error
		}
	}
	GetCompLocksV2 struct {
		Input struct {
			f sm.CompLockV2Filter
		}
		Return struct {
			cls []sm.CompLockV2
			err error
		}
	}
	UpdateCompLocksV2 struct {
		Input struct {
			f      sm.CompLockV2Filter
			action string
		}
		Return struct {
			results sm.CompLockV2UpdateResult
			err     error
		}
	}
	// Job Sync
	InsertJob struct {
		Input struct {
			j *sm.Job
		}
		Return struct {
			id  string
			err error
		}
	}
	UpdateJob struct {
		Input struct {
			jobId  string
			status string
		}
		Return struct {
			didUpdate bool
			err       error
		}
	}
	GetJob struct {
		Input struct {
			jobId string
		}
		Return struct {
			j   *sm.Job
			err error
		}
	}
	GetJobs struct {
		Input struct {
			f *hmsds.JobSyncFilter
		}
		Return struct {
			js  []*sm.Job
			err error
		}
	}
	DeleteJob struct {
		Input struct {
			jobId string
		}
		Return struct {
			didDelete bool
			err       error
		}
	}
}

type hmsdbtest struct {
	t TestResults
}

// Note that this is safe for multiple go routines and maintains its own
// connection pool and hence should be shared for best results.
func NewHMSDB_Test(dsn string, l *log.Logger) (hmsds.HMSDB, *TestResults) {
	d := new(hmsdbtest)
	return d, &d.t
}

// Conditional logging function (based on current log level set for conn)
func (d *hmsdbtest) Log(l hmsds.LogLevel, format string, a ...interface{}) {}

// Log to logging infrastructure regardless of current log level.
func (d *hmsdbtest) LogAlways(format string, a ...interface{}) {}

// Works like log.Printf, but registers error for function calling the
// function that is printing the error. e.g. instead of always saying
// an error occurred in begin(), we show where begin() was called, so
// we don't have to guess, and the message can make clear what in begin()
// failed.
func (d *hmsdbtest) LogAlwaysParentFunc(format string, a ...interface{}) {
	return
}

func (d *hmsdbtest) ImplementationName() string {
	return ""
}

func (d *hmsdbtest) SetLogLevel(lvl hmsds.LogLevel) error {
	return nil
}

////////////////////////////////////////////////////////////////////////////
//
// DB operations - Open, Close, Start Transaction
//
//////////////////////////////////////////////////////////////////////////////

func (d *hmsdbtest) Open() error {
	return nil
}

// Closes the database connection.  This is a global operation that
// affects all go routines using a hmsdb handle.  It is only used when
// we are done with the DB entirely.  Individual connections are pooled
// and managed transparently by the sql API, so fine-grained management
// is not needed for individual DB calls.
func (d *hmsdbtest) Close() error {
	return nil
}

// Starts a new transaction, returning a HMSDBTx handle which allows
// transaction-friendly operations to be invoked in sequence.  You
// MUST close the HMSDBTx handle with one of Commit or Rollback when
// done or the operations will eventually time out and rollback anyways,
// but in such a way that they may block operations on the same DB
// resources until then.
func (d *hmsdbtest) Begin() (hmsds.HMSDBTx, error) {
	return nil, nil
}

// Test the database connection to make sure that it is healthy
func (d *hmsdbtest) TestConnection() error {
	return d.t.TestConnection.Return.err
}

// Build filter query for Component IDs using filter functions and
// then return the list of matching xname IDs as a string array, write
// locking the rows if requested.
func (d *hmsdbtest) GetComponentIDs(f_opts ...hmsds.CompFiltFunc) ([]string, error) {
	d.t.GetComponentIDs.Input.id = d.t.GetComponentIDs.Funcs.getID(f_opts...)
	ids, err := d.t.GetComponentIDs.Funcs.returnIDs(d.t.GetComponentIDs.Input.id)
	d.t.GetComponentIDs.Return.err = err
	return ids, err
}

// Build filter query for ComponentEndpoints IDs using filter functions and
// then return the list of matching xname IDs as a string array, write
// locking the rows if requested.
func (d *hmsdbtest) GetCompEndpointIDs(f_opts ...hmsds.CompEPFiltFunc) ([]string, error) {
	d.t.GetCompEndpointIDs.Input.id = d.t.GetCompEndpointIDs.Funcs.getID(f_opts...)
	ids, err := d.t.GetCompEndpointIDs.Funcs.returnIDs(d.t.GetCompEndpointIDs.Input.id)
	d.t.GetCompEndpointIDs.Return.err = err
	return ids, err
}

// Build filter query for RedfishEndpoints IDs using filter functions and
// then return the list of matching xname IDs as a string array, write
// locking the rows if requested.
func (d *hmsdbtest) GetRFEndpointIDs(f_opts ...hmsds.RedfishEPFiltFunc) ([]string, error) {
	d.t.GetRFEndpointIDs.Input.id = d.t.GetRFEndpointIDs.Funcs.getID(f_opts...)
	ids, err := d.t.GetRFEndpointIDs.Funcs.returnIDs(d.t.GetRFEndpointIDs.Input.id)
	d.t.GetRFEndpointIDs.Return.err = err
	return ids, err
}

////////////////////////////////////////////////////////////////////////////
//
// HMS Components - Managed plane info: State, NID, Role
//
////////////////////////////////////////////////////////////////////////////

// Get a single component entry by its ID/xname
func (d *hmsdbtest) GetComponentByID(id string) (*base.Component, error) {
	d.t.GetComponentByID.Input.id = id
	return d.t.GetComponentByID.Return.id, d.t.GetComponentByID.Return.err
}

// Get all HMS Components in system.
func (d *hmsdbtest) GetComponentsAll() ([]*base.Component, error) {
	return d.t.GetComponentsAll.Return.ids, d.t.GetComponentsAll.Return.err
}

// Get some or all HMS Components in system, with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetComponentsFilter(f *hmsds.ComponentFilter, fieldFltr hmsds.FieldFilter) ([]*base.Component, error) {
	d.t.GetComponentsFilter.Input.compFilter = *f
	d.t.GetComponentsFilter.Input.fieldFilter = fieldFltr
	return d.t.GetComponentsFilter.Return.ids, d.t.GetComponentsFilter.Return.err
}

// Get some or all HMS Components in system under
// a set of parent components, with filtering options to possibly
// narrow the returned values. If no filter provided, just get
// the parent components.  Otherwise use it to create a custom
// WHERE... string that filters out entries that do not match ALL
// of the non-empty strings in the filter struct.
func (d *hmsdbtest) GetComponentsQuery(f *hmsds.ComponentFilter, fieldFltr hmsds.FieldFilter, ids []string) ([]*base.Component, error) {
	d.t.GetComponentsQuery.Input.compFilter = *f
	d.t.GetComponentsQuery.Input.fieldFilter = fieldFltr
	d.t.GetComponentsQuery.Input.ids = ids
	return d.t.GetComponentsQuery.Return.ids, d.t.GetComponentsQuery.Return.err
}

// Get a single component by its NID, if one exists.
func (d *hmsdbtest) GetComponentByNID(nid string) (*base.Component, error) {
	d.t.GetComponentByNID.Input.nid = nid
	return d.t.GetComponentByNID.Return.id, d.t.GetComponentByNID.Return.err
}

// Insert HMS Component into database, updating it if it exists.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (d *hmsdbtest) InsertComponent(c *base.Component) (int64, error) {
	d.t.InsertComponent.Input.c = c
	return d.t.InsertComponent.Return.rowsAffected, d.t.InsertComponent.Return.err
}

// Inserts or updates ComponentArray entries in database within a
// single all-or-none transaction.
func (d *hmsdbtest) InsertComponents(comps *base.ComponentArray) ([]string, error) {
	d.t.InsertComponents.Input.comps = comps
	return d.t.InsertComponents.Return.affectedIds, d.t.InsertComponents.Return.err
}

// Inserts or updates ComponentArray entries in database within a
// single all-or-none transaction. This will only overwrite the NID
// and Role fields (if set) for existing components unless force=true.
func (d *hmsdbtest) UpsertComponents(comps []*base.Component, force bool) (map[string]map[string]bool, error) {
	d.t.UpsertComponents.Input.comps = comps
	d.t.UpsertComponents.Input.force = force
	return d.t.UpsertComponents.Return.changeMap, d.t.UpsertComponents.Return.err
}

// Update state and flag fields only in DB for the given IDs.  If
// len(ids) is > 1 a locking read will be done to ensure the list o
// components that was actually modified is always returned.
//
// If force = true ignores any starting state restrictions and will
// always set ids to 'state', unless it is already set.
//
//	Note: If flag is not set, it will be set to OK (i.e. no flag)
func (d *hmsdbtest) UpdateCompStates(ids []string, state string, flag string, force bool, pi *hmsds.PartInfo) ([]string, error) {
	d.t.UpdateCompStates.Input.ids = ids
	d.t.UpdateCompStates.Input.state = state
	d.t.UpdateCompStates.Input.flag = flag
	d.t.UpdateCompStates.Input.force = force
	d.t.UpdateCompStates.Input.pi = pi
	return d.t.UpdateCompStates.Return.affectedIds, d.t.UpdateCompStates.Return.err
}

// Update Flag field in DB from c's Flag field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Flag cannot be blank/invalid.
func (d *hmsdbtest) UpdateCompFlagOnly(id string, flag string) (int64, error) {
	d.t.UpdateCompFlagOnly.Input.id = id
	d.t.UpdateCompFlagOnly.Input.flag = flag
	return d.t.UpdateCompFlagOnly.Return.rowsAffected, d.t.UpdateCompFlagOnly.Return.err
}

// Update flag field in DB for a list of components
// Note: Flag cannot be empty/invalid.
func (d *hmsdbtest) BulkUpdateCompFlagOnly(ids []string, flag string) ([]string, error) {
	d.t.BulkUpdateCompFlagOnly.Input.ids = ids
	d.t.BulkUpdateCompFlagOnly.Input.flag = flag
	return d.t.BulkUpdateCompFlagOnly.Return.affectedIds, d.t.BulkUpdateCompFlagOnly.Return.err
}

// Update Enabled field in DB from c's Enabled field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: c.Enabled cannot be nil
func (d *hmsdbtest) UpdateCompEnabled(id string, enabled bool) (int64, error) {
	d.t.UpdateCompEnabled.Input.id = id
	d.t.UpdateCompEnabled.Input.enabled = enabled
	return d.t.UpdateCompEnabled.Return.rowsAffected, d.t.UpdateCompEnabled.Return.err
}

// Update Enabled field in DB for a list of components
func (d *hmsdbtest) BulkUpdateCompEnabled(ids []string, enabled bool) ([]string, error) {
	d.t.BulkUpdateCompEnabled.Input.ids = ids
	d.t.BulkUpdateCompEnabled.Input.enabled = enabled
	return d.t.BulkUpdateCompEnabled.Return.affectedIds, d.t.BulkUpdateCompEnabled.Return.err
}

// Update SwStatus field in DB from c's SwStatus field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (d *hmsdbtest) UpdateCompSwStatus(id string, swStatus string) (int64, error) {
	d.t.UpdateCompSwStatus.Input.id = id
	d.t.UpdateCompSwStatus.Input.swStatus = swStatus
	return d.t.UpdateCompSwStatus.Return.rowsAffected, d.t.UpdateCompSwStatus.Return.err
}

// Update SwStatus field in DB for a list of components
func (d *hmsdbtest) BulkUpdateCompSwStatus(ids []string, swstatus string) ([]string, error) {
	d.t.BulkUpdateCompSwStatus.Input.ids = ids
	d.t.BulkUpdateCompSwStatus.Input.swstatus = swstatus
	return d.t.BulkUpdateCompSwStatus.Return.affectedIds, d.t.BulkUpdateCompSwStatus.Return.err
}

// Update Role field in DB from c's Role field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Role cannot be blank/invalid.
func (d *hmsdbtest) UpdateCompRole(id string, role, subRole string) (int64, error) {
	d.t.UpdateCompRole.Input.id = id
	d.t.UpdateCompRole.Input.role = role
	d.t.UpdateCompRole.Input.subRole = subRole
	return d.t.UpdateCompRole.Return.rowsAffected, d.t.UpdateCompRole.Return.err
}

// Update Role field in DB for a list of components
// Note: Role cannot be empty/invalid.
func (d *hmsdbtest) BulkUpdateCompRole(ids []string, role, subRole string) ([]string, error) {
	d.t.BulkUpdateCompRole.Input.ids = ids
	d.t.BulkUpdateCompRole.Input.role = role
	d.t.BulkUpdateCompRole.Input.subRole = subRole
	return d.t.BulkUpdateCompRole.Return.affectedIds, d.t.BulkUpdateCompRole.Return.err
}

// Update Class field in DB for a list of components
func (d *hmsdbtest) BulkUpdateCompClass(ids []string, class string) ([]string, error) {
	d.t.BulkUpdateCompClass.Input.ids = ids
	d.t.BulkUpdateCompClass.Input.class = class
	return d.t.BulkUpdateCompClass.Return.affectedIds, d.t.BulkUpdateCompClass.Return.err
}

// Update NID field in DB from c's NID field.
// Note: NID cannot be blank.  Should be negative to unset.
func (d *hmsdbtest) UpdateCompNID(c *base.Component) error {
	d.t.UpdateCompNID.Input.c = c
	return d.t.UpdateCompNID.Return.err
}

// Update NID field in DB for a list of components
// Note: NID cannot be blank.  Should be negative to unset.
func (d *hmsdbtest) BulkUpdateCompNID(comps *[]base.Component) error {
	d.t.BulkUpdateCompNID.Input.comps = comps
	return d.t.BulkUpdateCompNID.Return.err
}

// Delete HMS Component with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteComponentByID(id string) (bool, error) {
	d.t.DeleteComponentByID.Input.id = id
	return d.t.DeleteComponentByID.Return.changed, d.t.DeleteComponentByID.Return.err
}

// Delete all HMS Components from database (atomically)
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteComponentsAll() (int64, error) {
	return d.t.DeleteComponentsAll.Return.numRows, d.t.DeleteComponentsAll.Return.err
}

/////////////////////////////////////////////////////////////////////////////
//
// Node->NID Mapping
//
/////////////////////////////////////////////////////////////////////////////

// Look up one Node->NID Mapping by id, i.e. node xname.
func (d *hmsdbtest) GetNodeMapByID(id string) (*sm.NodeMap, error) {
	d.t.GetNodeMapByID.Input.id = id
	return d.t.GetNodeMapByID.Return.entry, d.t.GetNodeMapByID.Return.err
}

// Look up ALL Node->NID Mappings.
func (d *hmsdbtest) GetNodeMapsAll() ([]*sm.NodeMap, error) {
	return d.t.GetNodeMapsAll.Return.entry, d.t.GetNodeMapsAll.Return.err
}

// Insert Node->NID Mapping into database, updating it if it exists.
func (d *hmsdbtest) InsertNodeMap(m *sm.NodeMap) error {
	d.t.InsertNodeMap.Input.m = m
	return d.t.InsertNodeMap.Return.err
}

// Inserts or updates Node->NID Mapping Array entries in database within a
// single all-or-none transaction.
func (d *hmsdbtest) InsertNodeMaps(nnms *sm.NodeMapArray) error {
	d.t.InsertNodeMaps.Input.nnms = nnms
	return d.t.InsertNodeMaps.Return.err
}

// Delete Node NID Mapping entry with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteNodeMapByID(id string) (bool, error) {
	d.t.DeleteNodeMapByID.Input.id = id
	return d.t.DeleteNodeMapByID.Return.changed, d.t.DeleteNodeMapByID.Return.err
}

// Delete all Node NID Mapping entries from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteNodeMapsAll() (int64, error) {
	return d.t.DeleteNodeMapsAll.Return.numRows, d.t.DeleteNodeMapsAll.Return.err
}

/////////////////////////////////////////////////////////////////////////////
//
// Power Mapping
//
/////////////////////////////////////////////////////////////////////////////

// Look up one Power Mapping by id, i.e. node xname.
func (d *hmsdbtest) GetPowerMapByID(id string) (*sm.PowerMap, error) {
	d.t.GetPowerMapByID.Input.id = id
	return d.t.GetPowerMapByID.Return.m, d.t.GetPowerMapByID.Return.err
}

// Look up ALL Power Mappings.
func (d *hmsdbtest) GetPowerMapsAll() ([]*sm.PowerMap, error) {
	return d.t.GetPowerMapsAll.Return.ms, d.t.GetPowerMapsAll.Return.err
}

// Insert Node->NID Mapping into database, updating it if it exists.
func (d *hmsdbtest) InsertPowerMap(m *sm.PowerMap) error {
	d.t.InsertPowerMap.Input.m = m
	return d.t.InsertPowerMap.Return.err
}

// Inserts or updates Power Mapping Array entries in database within a
// single all-or-none transaction.
func (d *hmsdbtest) InsertPowerMaps(ms []sm.PowerMap) error {
	d.t.InsertPowerMaps.Input.ms = ms
	return d.t.InsertPowerMaps.Return.err
}

// Delete Power Mapping entry with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeletePowerMapByID(id string) (bool, error) {
	d.t.DeletePowerMapByID.Input.id = id
	return d.t.DeletePowerMapByID.Return.changed, d.t.DeletePowerMapByID.Return.err
}

// Delete all Power Mapping entries from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeletePowerMapsAll() (int64, error) {
	return d.t.DeletePowerMapsAll.Return.numRows, d.t.DeletePowerMapsAll.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Hardware Inventory - Detailed location and FRU info
//
////////////////////////////////////////////////////////////////////////////

// Get some or all Hardware Inventory entries with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct.
// This does hierarchy searches.
func (d *hmsdbtest) GetHWInvByLocQueryFilter(f_opts ...hmsds.HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error) {
	f := new(hmsds.HWInvLocFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	d.t.GetHWInvByLocQueryFilter.Input.f = f
	return d.t.GetHWInvByLocQueryFilter.Return.hwlocs, d.t.GetHWInvByLocQueryFilter.Return.err
}

// Get some or all Hardware Inventory entries with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetHWInvByLocFilter(f_opts ...hmsds.HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error) {
	f := new(hmsds.HWInvLocFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	d.t.GetHWInvByLocFilter.Input.f = f
	return d.t.GetHWInvByLocFilter.Return.hwlocs, d.t.GetHWInvByLocFilter.Return.err
}

// Get a single Hardware inventory entry by current xname
// This struct includes the FRU info if the xname is currently populated.
func (d *hmsdbtest) GetHWInvByLocID(id string) (*sm.HWInvByLoc, error) {
	d.t.GetHWInvByLocID.Input.id = id
	return d.t.GetHWInvByLocID.Return.entry, d.t.GetHWInvByLocID.Return.err
}

// Get HWInvByLoc by primary key (xname) for all entries in the system.
// It also pairs the data with the matching HWInvByFRU if the xname is
// populated.
func (d *hmsdbtest) GetHWInvByLocAll() ([]*sm.HWInvByLoc, error) {
	return d.t.GetHWInvByLocAll.Return.entries, d.t.GetHWInvByLocAll.Return.err
}

// Get HW Inventory-by-FRU entry at the provided location FRU ID
func (d *hmsdbtest) GetHWInvByFRUID(fruid string) (*sm.HWInvByFRU, error) {
	d.t.GetHWInvByFRUID.Input.fruid = fruid
	return d.t.GetHWInvByFRUID.Return.entry, d.t.GetHWInvByFRUID.Return.err
}

// Get some or all HW-inventory-by-FRU entries with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetHWInvByFRUFilter(f_opts ...hmsds.HWInvLocFiltFunc) ([]*sm.HWInvByFRU, error) {
	f := new(hmsds.HWInvLocFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	d.t.GetHWInvByFRUFilter.Input.f = f
	return d.t.GetHWInvByFRUFilter.Return.hwfrus, d.t.GetHWInvByFRUFilter.Return.err
}

// Get HW Inventory-by-FRU entry at the provided location FRU ID
func (d *hmsdbtest) GetHWInvByFRUAll() ([]*sm.HWInvByFRU, error) {
	return d.t.GetHWInvByFRUAll.Return.entries, d.t.GetHWInvByFRUAll.Return.err
}

// Insert or update HWInventoryByLocation struct.
// If PopulatedFRU is present, this is also added to the DB  If
// it is not, this effectively "depopulates" the given location.
// The actual HWInventoryByFRU is stored using within the same
// transaction.
func (d *hmsdbtest) InsertHWInvByLoc(hl *sm.HWInvByLoc) error {
	d.t.InsertHWInvByLoc.Input.hl = hl
	return d.t.InsertHWInvByLoc.Return.err
}

// Insert or update HWInventoryByFRU struct.  This does not associate
// the object with any HW-Inventory-By-Location info so it is
// typically not needed.  InsertHWInvByLoc is typically used to
// store both type of info at once.
func (d *hmsdbtest) InsertHWInvByFRU(hf *sm.HWInvByFRU) error {
	d.t.InsertHWInvByFRU.Input.hf = hf
	return d.t.InsertHWInvByFRU.Return.err
}

// Insert or update array of HWInventoryByLocation structs.
// If PopulatedFRU is present, these is also added to the DB  If
// it is not, this effectively "depopulates" the given locations.
// The actual HWInventoryByFRU is stored using within the same
// transaction.
func (d *hmsdbtest) InsertHWInvByLocs(hls []*sm.HWInvByLoc) error {
	d.t.InsertHWInvByLocs.Input.hls = hls
	return d.t.InsertHWInvByLocs.Return.err
}

// Delete HWInvByLoc entry with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteHWInvByLocID(id string) (bool, error) {
	d.t.DeleteHWInvByLocID.Input.id = id
	return d.t.DeleteHWInvByLocID.Return.changed, d.t.DeleteHWInvByLocID.Return.err
}

// Delete ALL HWInvByLoc entries from database (atomically)
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteHWInvByLocsAll() (int64, error) {
	return d.t.DeleteHWInvByLocsAll.Return.numRows, d.t.DeleteHWInvByLocsAll.Return.err
}

// Delete HWInvByFRU entry with matching FRU ID from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteHWInvByFRUID(fruid string) (bool, error) {
	d.t.DeleteHWInvByFRUID.Input.fruid = fruid
	return d.t.DeleteHWInvByFRUID.Return.changed, d.t.DeleteHWInvByFRUID.Return.err
}

// Delete ALL HWInvByFRU entries from database (atomically)
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteHWInvByFRUsAll() (int64, error) {
	return d.t.DeleteHWInvByFRUsAll.Return.numRows, d.t.DeleteHWInvByFRUsAll.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Hardware Inventory History - Detailed history of hardware FRU location.
//
////////////////////////////////////////////////////////////////////////////

// Get hardware history for some or all Hardware Inventory entries with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetHWInvHistFilter(f_opts ...hmsds.HWInvHistFiltFunc) ([]*sm.HWInvHist, error) {
	f := new(hmsds.HWInvHistFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	d.t.GetHWInvHistFilter.Input.f = f
	return d.t.GetHWInvHistFilter.Return.hwhists, d.t.GetHWInvHistFilter.Return.err
}

// Get only the most recent hardware history event for some or all hardware locations.
func (d *hmsdbtest) GetHWInvHistLastEvents(ids []string) ([]*sm.HWInvHist, error) {
	d.t.GetHWInvHistLastEvents.Input.ids = ids
	return d.t.GetHWInvHistLastEvents.Return.hwhists, d.t.GetHWInvHistLastEvents.Return.err
}

// Insert a HWInventoryHistory entry.
// If a duplicate is present return an error.
func (d *hmsdbtest) InsertHWInvHist(hh *sm.HWInvHist) error {
	d.t.InsertHWInvHist.Input.hh = hh
	return d.t.InsertHWInvHist.Return.err
}

// Insert an array of HWInventoryHistory entries.
// If a duplicate is present return an error.
func (d *hmsdbtest) InsertHWInvHists(hhs []*sm.HWInvHist) error {
	d.t.InsertHWInvHists.Input.hhs = hhs
	return d.t.InsertHWInvHists.Return.err
}

// Delete all HWInvHist entries with matching xname id from database, if it
// exists.
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteHWInvHistByLocID(id string) (int64, error) {
	d.t.DeleteHWInvHistByLocID.Input.id = id
	return d.t.DeleteHWInvHistByLocID.Return.numRows, d.t.DeleteHWInvHistByLocID.Return.err
}

// Delete all HWInvHist entries with matching FRU id from database, if it
// exists.
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteHWInvHistByFRUID(fruid string) (int64, error) {
	d.t.DeleteHWInvHistByFRUID.Input.fruid = fruid
	return d.t.DeleteHWInvHistByFRUID.Return.numRows, d.t.DeleteHWInvHistByFRUID.Return.err
}

// Delete all HWInvHist entries from database (atomically)
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteHWInvHistAll() (int64, error) {
	return d.t.DeleteHWInvHistAll.Return.numRows, d.t.DeleteHWInvHistAll.Return.err
}

// Delete all HWInvHist entries from database matching a filter.
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteHWInvHistFilter(f_opts ...hmsds.HWInvHistFiltFunc) (int64, error) {
	f := new(hmsds.HWInvHistFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	d.t.DeleteHWInvHistFilter.Input.f = f
	return d.t.DeleteHWInvHistFilter.Return.numRows, d.t.DeleteHWInvHistFilter.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Redfish Endpoints - Top-level Redfish service roots used for discovery
//
////////////////////////////////////////////////////////////////////////////

// Get RedfishEndpoint by ID (xname), i.e. a single entry.
func (d *hmsdbtest) GetRFEndpointByID(id string) (*sm.RedfishEndpoint, error) {
	d.t.GetRFEndpointByID.Input.id = id
	return d.t.GetRFEndpointByID.Return.entry, d.t.GetRFEndpointByID.Return.err
}

// Get all RedfishEndpoints in system.
func (d *hmsdbtest) GetRFEndpointsAll() ([]*sm.RedfishEndpoint, error) {
	return d.t.GetRFEndpointsAll.Return.entries, d.t.GetRFEndpointsAll.Return.err
}

// Get some or all RedfishEndpoints in system, with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetRFEndpointsFilter(f *hmsds.RedfishEPFilter) ([]*sm.RedfishEndpoint, error) {
	d.t.GetRFEndpointsFilter.Input.f = f
	return d.t.GetRFEndpointsFilter.Return.entries, d.t.GetRFEndpointsFilter.Return.err
}

// Insert new RedfishEndpoint into database.
// Does not update any ComponentEndpoint children.
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (d *hmsdbtest) InsertRFEndpoint(ep *sm.RedfishEndpoint) error {
	d.t.InsertRFEndpoint.Input.ep = ep
	return d.t.InsertRFEndpoint.Return.err
}

// Insert new RedfishEndpointArray entries into database within a
// single all-or-none transaction.  Does not update any ComponentEndpoint
// children.
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (d *hmsdbtest) InsertRFEndpoints(eps *sm.RedfishEndpointArray) error {
	d.t.InsertRFEndpoints.Input.eps = eps
	return d.t.InsertRFEndpoints.Return.err
}

// Update existing RedfishEndpointArray entry in database.
// Does not update any ComponentEndpoint children.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbtest) UpdateRFEndpoint(ep *sm.RedfishEndpoint) (*sm.RedfishEndpoint, error) {
	d.t.UpdateRFEndpoint.Input.ep = ep
	return d.t.UpdateRFEndpoint.Return.entry, d.t.UpdateRFEndpoint.Return.err
}

// Update existing RedfishEndpointArray entry in database, but only updates
// fields that would be changed by a user-directed operation.
// Does not update any ComponentEndpoint children.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbtest) UpdateRFEndpointNoDiscInfo(ep *sm.RedfishEndpoint) (*sm.RedfishEndpoint, []string, error) {
	d.t.UpdateRFEndpointNoDiscInfo.Input.ep = ep
	return d.t.UpdateRFEndpointNoDiscInfo.Return.entry, d.t.UpdateRFEndpointNoDiscInfo.Return.affectedIds, d.t.UpdateRFEndpointNoDiscInfo.Return.err
}

// Patch existing RedfishEndpointArray entry in database, but only updates
// specified fields.
// Does not update any ComponentEndpoint children.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbtest) PatchRFEndpointNoDiscInfo(id string, epp sm.RedfishEndpointPatch) (*sm.RedfishEndpoint, []string, error) {
	d.t.PatchRFEndpointNoDiscInfo.Input.id = id
	d.t.PatchRFEndpointNoDiscInfo.Input.epp = epp
	return d.t.PatchRFEndpointNoDiscInfo.Return.entry, d.t.PatchRFEndpointNoDiscInfo.Return.affectedIds, d.t.PatchRFEndpointNoDiscInfo.Return.err
}

// Returns: Discoverable endpoint list, with status set appropriately in DB
// and return values.  However this list will omit those RF EPs  who are
// already being discovered, unless forced.
// Error returned on unexpected failure or any entry in eps not existing,
// the latter error being ErrHMSDSNoREP.
func (d *hmsdbtest) UpdateRFEndpointForDiscover(ids []string, force bool) (
	[]*sm.RedfishEndpoint, error) {

	d.t.UpdateRFEndpointForDiscover.Input.ids = ids
	d.t.UpdateRFEndpointForDiscover.Input.force = force
	return d.t.UpdateRFEndpointForDiscover.Return.entries,
		d.t.UpdateRFEndpointForDiscover.Return.err
}

// Update existing RedfishEndpointArray entries in database within a
// single all-or-none transaction.  Does not update any ComponentEndpoint
// children.
// Returns FALSE with err == nil if one or more updated entries do
// not exist.  No updates are performed in this case.
func (d *hmsdbtest) UpdateRFEndpoints(eps *sm.RedfishEndpointArray) (bool, error) {
	d.t.UpdateRFEndpoints.Input.eps = eps
	return d.t.UpdateRFEndpoints.Return.changed, d.t.UpdateRFEndpoints.Return.err
}

// Delete RedfishEndpoint with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteRFEndpointByID(id string) (bool, error) {
	d.t.DeleteRFEndpointByID.Input.id = id
	return d.t.DeleteRFEndpointByID.Return.changed, d.t.DeleteRFEndpointByID.Return.err
}

// Delete all RedfishEndpoints from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteRFEndpointsAll() (int64, error) {
	return d.t.DeleteRFEndpointsAll.Return.numRows, d.t.DeleteRFEndpointsAll.Return.err
}

// Delete RedfishEndpoint with matching xname id from database, if it
// exists.  When dooing so, set all HMS Components to Empty if they
// are children of the RedfishEndpoint.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteRFEndpointByIDSetEmpty(id string) (bool, []string, error) {
	d.t.DeleteRFEndpointByIDSetEmpty.Input.id = id
	return d.t.DeleteRFEndpointByIDSetEmpty.Return.changed,
		d.t.DeleteRFEndpointByIDSetEmpty.Return.affectedIds,
		d.t.DeleteRFEndpointByIDSetEmpty.Return.err
}

// Delete all RedfishEndpoints from database.
// This also deletes all child ComponentEndpoints, and in addition,
// sets the State/Components entries for those ComponentEndpoints to Empty/OK
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteRFEndpointsAllSetEmpty() (int64, []string, error) {
	return d.t.DeleteRFEndpointsAllSetEmpty.Return.numRows,
		d.t.DeleteRFEndpointsAllSetEmpty.Return.affectedIds,
		d.t.DeleteRFEndpointsAllSetEmpty.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Component Endpoints - Component info discovered from parent RedfishEndpoint
//                       Management plane equivalent to HMS Component
//
////////////////////////////////////////////////////////////////////////////

// Get ComponentEndpoint by id (xname), i.e. a single entry.
func (d *hmsdbtest) GetCompEndpointByID(id string) (*sm.ComponentEndpoint, error) {
	d.t.GetCompEndpointByID.Input.id = id
	return d.t.GetCompEndpointByID.Return.entry, d.t.GetCompEndpointByID.Return.err
}

// Get all ComponentEndpoints in system.
func (d *hmsdbtest) GetCompEndpointsAll() ([]*sm.ComponentEndpoint, error) {
	return d.t.GetCompEndpointsAll.Return.entries, d.t.GetCompEndpointsAll.Return.err
}

// Get some or all ComponentEndpoints in system, with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetCompEndpointsFilter(f *hmsds.CompEPFilter) ([]*sm.ComponentEndpoint, error) {
	d.t.GetCompEndpointsFilter.Input.f = f
	return d.t.GetCompEndpointsFilter.Return.entries, d.t.GetCompEndpointsFilter.Return.err
}

// Upsert ComponentEndpoint into database, updating it if it exists.
func (d *hmsdbtest) UpsertCompEndpoint(cep *sm.ComponentEndpoint) error {
	d.t.UpsertCompEndpoint.Input.cep = cep
	return d.t.UpsertCompEndpoint.Return.err
}

// Upsert ComponentEndpointArray into database within a single all-or-none
// transaction.
func (d *hmsdbtest) UpsertCompEndpoints(ceps *sm.ComponentEndpointArray) error {
	d.t.UpsertCompEndpoints.Input.ceps = ceps
	return d.t.UpsertCompEndpoints.Return.err
}

// Delete ComponentEndpoint with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteCompEndpointByID(id string) (bool, error) {
	d.t.DeleteCompEndpointByID.Input.id = id
	return d.t.DeleteCompEndpointByID.Return.changed, d.t.DeleteCompEndpointByID.Return.err
}

// Delete all ComponentEndpoints from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteCompEndpointsAll() (int64, error) {
	return d.t.DeleteCompEndpointsAll.Return.numRows, d.t.DeleteCompEndpointsAll.Return.err
}

// Delete ComponentEndpoint with matching xname id from database, if it
// exists.  When dooing so, set the corresponding HMS Component to Empty if it
// is not already in that state.
// Return true if there was a row affected, false if there were zero.  The
// string array returns the single xname ID that changed state or is empty.
func (d *hmsdbtest) DeleteCompEndpointByIDSetEmpty(id string) (bool, []string, error) {
	d.t.DeleteCompEndpointByIDSetEmpty.Input.id = id
	return d.t.DeleteCompEndpointByIDSetEmpty.Return.changed,
		d.t.DeleteCompEndpointByIDSetEmpty.Return.affectedIds,
		d.t.DeleteCompEndpointByIDSetEmpty.Return.err
}

// Delete all ComponentEndpoints from database. In addition,
// sets the State/Components entry for each ComponentEndpoint to Empty/OK
// Also returns number of deleted rows, if error is nil, and also string array
// of those xname IDs that were set to Empty/OK (i.e. not already Empty/OK)
// as part of the deletion.
func (d *hmsdbtest) DeleteCompEndpointsAllSetEmpty() (int64, []string, error) {
	return d.t.DeleteCompEndpointsAllSetEmpty.Return.numRows,
		d.t.DeleteCompEndpointsAllSetEmpty.Return.affectedIds,
		d.t.DeleteCompEndpointsAllSetEmpty.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Service Endpoints - Component info discovered from parent RedfishEndpoint
//
////////////////////////////////////////////////////////////////////////////

// Get ServiceEndpoint by service type and id (xname), i.e. a single entry.
func (d *hmsdbtest) GetServiceEndpointByID(svc, id string) (*sm.ServiceEndpoint, error) {
	d.t.GetServiceEndpointByID.Input.svc = svc
	d.t.GetServiceEndpointByID.Input.id = id
	return d.t.GetServiceEndpointByID.Return.entry, d.t.GetServiceEndpointByID.Return.err
}

// Get all ServiceEndpoints in system.
func (d *hmsdbtest) GetServiceEndpointsAll() ([]*sm.ServiceEndpoint, error) {
	return d.t.GetServiceEndpointsAll.Return.entries, d.t.GetServiceEndpointsAll.Return.err
}

// Get some or all ServiceEndpoints in system, with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetServiceEndpointsFilter(f *hmsds.ServiceEPFilter) ([]*sm.ServiceEndpoint, error) {
	d.t.GetServiceEndpointsFilter.Input.f = f
	return d.t.GetServiceEndpointsFilter.Return.entries, d.t.GetServiceEndpointsFilter.Return.err
}

// Upsert ServiceEndpoint into database, updating it if it exists.
func (d *hmsdbtest) UpsertServiceEndpoint(sep *sm.ServiceEndpoint) error {
	d.t.UpsertServiceEndpoint.Input.sep = sep
	return d.t.UpsertServiceEndpoint.Return.err
}

// Upsert ServiceEndpointArray into database within a single all-or-none
// transaction.
func (d *hmsdbtest) UpsertServiceEndpoints(seps *sm.ServiceEndpointArray) error {
	d.t.UpsertServiceEndpoints.Input.seps = seps
	return d.t.UpsertServiceEndpoints.Return.err
}

// Delete ServiceEndpoint with matching service type and xname id from
// database, if it exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteServiceEndpointByID(svc, id string) (bool, error) {
	d.t.DeleteServiceEndpointByID.Input.svc = svc
	d.t.DeleteServiceEndpointByID.Input.id = id
	return d.t.DeleteServiceEndpointByID.Return.changed, d.t.DeleteServiceEndpointByID.Return.err
}

// Delete all ServiceEndpoints from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteServiceEndpointsAll() (int64, error) {
	return d.t.DeleteServiceEndpointsAll.Return.numRows, d.t.DeleteServiceEndpointsAll.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Component Ethernet Interfaces - MAC address to IP address relations
//     for component endpoint ethernet interfaces
//
////////////////////////////////////////////////////////////////////////////

// Get some or all CompEthInterfaces in the system, with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbtest) GetCompEthInterfaceFilter(f_opts ...hmsds.CompEthInterfaceFiltFunc) ([]*sm.CompEthInterfaceV2, error) {
	f := new(hmsds.CompEthInterfaceFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	d.t.GetCompEthInterfaceFilter.Input.f = f
	return d.t.GetCompEthInterfaceFilter.Return.ceis, d.t.GetCompEthInterfaceFilter.Return.err
}

// Insert a new CompEthInterface into the database.
// If ID or MAC address already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (d *hmsdbtest) InsertCompEthInterface(cei *sm.CompEthInterfaceV2) error {
	d.t.InsertCompEthInterface.Input.cei = cei
	return d.t.InsertCompEthInterface.Return.err
}

// Insert new CompEthInterfaces into the database within a single
// all-or-none transaction.
// If ID or MAC address already exists, return ErrHMSDSDuplicateKey
// No insertions are done on err != nil
func (d *hmsdbtest) InsertCompEthInterfaces(ceis []*sm.CompEthInterfaceV2) error {
	d.t.InsertCompEthInterfaces.Input.ceis = ceis
	return d.t.InsertCompEthInterfaces.Return.err
}

// Insert/update a CompEthInterface in the database.
// If ID or MAC address already exists, only overwrite ComponentID
// and Type fields.
// No insertion done on err != nil
func (d *hmsdbtest) InsertCompEthInterfaceCompInfo(cei *sm.CompEthInterfaceV2) error {
	d.t.InsertCompEthInterfaceCompInfo.Input.cei = cei
	return d.t.InsertCompEthInterfaceCompInfo.Return.err
}

// Insert new CompEthInterfaces into database within a single
// all-or-none transaction.
// If ID or MAC address already exists, only overwrite ComponentID
// and Type fields.
// No insertions are done on err != nil
func (d *hmsdbtest) InsertCompEthInterfacesCompInfo(ceis []*sm.CompEthInterfaceV2) error {
	d.t.InsertCompEthInterfacesCompInfo.Input.ceis = ceis
	return d.t.InsertCompEthInterfacesCompInfo.Return.err
}

// Update existing CompEthInterface entry in the database, but only updates
// fields that would be changed by a user-directed operation.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbtest) UpdateCompEthInterface(id string, ceip *sm.CompEthInterfaceV2Patch) (*sm.CompEthInterfaceV2, error) {
	d.t.UpdateCompEthInterface.Input.id = id
	d.t.UpdateCompEthInterface.Input.ceip = ceip
	return d.t.UpdateCompEthInterface.Return.cei, d.t.UpdateCompEthInterface.Return.err
}

// Update existing CompEthInterface entry in the database, but only updates
// fields that would be changed by a user-directed operation.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
//
// Special handling is required to use the V1 API Patch on a V2 CompEthInterface.
// If the CEI has more than 2 or more IP addresses associated with it the error
// CompEthInterfacePatch will be ErrHMSDSCompEthInterfaceMultipleIPs returned.
func (d *hmsdbtest) UpdateCompEthInterfaceV1(id string, ceip *sm.CompEthInterfacePatch) (*sm.CompEthInterfaceV2, error) {
	d.t.UpdateCompEthInterfaceV1.Input.id = id
	d.t.UpdateCompEthInterfaceV1.Input.ceip = ceip
	return d.t.UpdateCompEthInterfaceV1.Return.cei, d.t.UpdateCompEthInterfaceV1.Return.err
}

// Delete CompEthInterface with matching id from the database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbtest) DeleteCompEthInterfaceByID(id string) (bool, error) {
	d.t.DeleteCompEthInterfaceByID.Input.id = id
	return d.t.DeleteCompEthInterfaceByID.Return.didDelete, d.t.DeleteCompEthInterfaceByID.Return.err
}

// Delete all CompEthInterfaces from the database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbtest) DeleteCompEthInterfacesAll() (int64, error) {
	return d.t.DeleteCompEthInterfacesAll.Return.numRows, d.t.DeleteCompEthInterfacesAll.Return.err
}

// Add IP Address mapping to the existing component ethernet interface.
// returns:
//   - ErrHMSDSNoCompEthInterface if the parent component ethernet interface
//   - ErrHMSDSDuplicateKey if the parent component ethernet interface already
//     has that IP address
//
// Returns key of new IP Address Mapping id, should be the IP address
func (d *hmsdbtest) AddCompEthInterfaceIPAddress(id string, ipm *sm.IPAddressMapping) (string, error) {
	d.t.AddCompEthInterfaceIPAddress.Input.id = id
	d.t.AddCompEthInterfaceIPAddress.Input.ipm = ipm

	return d.t.AddCompEthInterfaceIPAddress.Return.id, d.t.AddCompEthInterfaceIPAddress.Return.err
}

// Update existing IP Address Mapping for a CompEthInterface entry in the database,
// but only updates fields that would be changed by a user-directed operation.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbtest) UpdateCompEthInterfaceIPAddress(id, ipAddr string, ipmPatch *sm.IPAddressMappingPatch) (*sm.IPAddressMapping, error) {
	d.t.UpdateCompEthInterfaceIPAddress.Input.id = id
	d.t.UpdateCompEthInterfaceIPAddress.Input.ipAddr = ipAddr
	d.t.UpdateCompEthInterfaceIPAddress.Input.ipmPatch = ipmPatch

	return d.t.UpdateCompEthInterfaceIPAddress.Output.ipm, d.t.UpdateCompEthInterfaceIPAddress.Output.err

}

// Delete IP Address mapping from the Component Ethernet Interface.
// If no error, bool indicates whether the IP Address Mapping was present to remove.
func (d *hmsdbtest) DeleteCompEthInterfaceIPAddress(id, ipAddr string) (bool, error) {
	d.t.DeleteCompEthInterfaceIPAddress.Input.id = id
	d.t.DeleteCompEthInterfaceIPAddress.Input.ipAddr = ipAddr

	return d.t.DeleteCompEthInterfaceIPAddress.Output.didDelete, d.t.DeleteCompEthInterfaceIPAddress.Output.err
}

/////////////////////////////////////////////////////////////////////////////
//
// DiscoveryStatus - Discovery status tracking
//
/////////////////////////////////////////////////////////////////////////////

// Get DiscoveryStatus with the given numerical ID.
func (d *hmsdbtest) GetDiscoveryStatusByID(id uint) (*sm.DiscoveryStatus, error) {
	d.t.GetDiscoveryStatusByID.Input.id = id
	return d.t.GetDiscoveryStatusByID.Return.status, d.t.GetDiscoveryStatusByID.Return.err
}

// Get all DiscoveryStatus entries.
func (d *hmsdbtest) GetDiscoveryStatusAll() ([]*sm.DiscoveryStatus, error) {
	return d.t.GetDiscoveryStatusAll.Return.status, d.t.GetDiscoveryStatusAll.Return.err
}

// Update discovery status in DB.
func (d *hmsdbtest) UpsertDiscoveryStatus(stat *sm.DiscoveryStatus) error {
	d.t.UpsertDiscoveryStatus.Input.stat = stat
	return d.t.UpsertDiscoveryStatus.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Discovery operations - Multi-type atomic operations.
//
////////////////////////////////////////////////////////////////////////////

// Atomically:
//
//  1. Update discovery-writable fields for RedfishEndpoint
//  2. Upsert ComponentEndpointArray into database within the
//     same transaction.
//  3. Insert or update array of HWInventoryByLocation structs.
//     If PopulatedFRU is present, these is also added to the DB  If
//     it is not, this effectively "depopulates" the given locations.
//     The actual HWInventoryByFRU is stored using within the same
//     transaction.
//  4. Inserts or updates HMS Components entries in ComponentArray
func (d *hmsdbtest) UpdateAllForRFEndpoint(
	ep *sm.RedfishEndpoint,
	ceps *sm.ComponentEndpointArray,
	hls []*sm.HWInvByLoc,
	comps *base.ComponentArray,
	seps *sm.ServiceEndpointArray,
	ceis []*sm.CompEthInterfaceV2,
) (*[]base.Component, error) {
	d.t.UpdateAllForRFEndpoint.Input.ep = ep
	d.t.UpdateAllForRFEndpoint.Input.ceps = ceps
	d.t.UpdateAllForRFEndpoint.Input.hls = hls
	d.t.UpdateAllForRFEndpoint.Input.comps = comps
	d.t.UpdateAllForRFEndpoint.Input.seps = seps
	d.t.UpdateAllForRFEndpoint.Input.ceis = ceis
	return d.t.UpdateAllForRFEndpoint.Return.discoveredIds, d.t.UpdateAllForRFEndpoint.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// SCN subscription operations
//
////////////////////////////////////////////////////////////////////////////

// Get all SCN subscriptions
func (d *hmsdbtest) GetSCNSubscriptionsAll() (*sm.SCNSubscriptionArray, error) {
	return d.t.GetSCNSubscriptionsAll.Return.subs, d.t.GetSCNSubscriptionsAll.Return.err
}

// Get a SCN subscription
func (d *hmsdbtest) GetSCNSubscription(id int64) (*sm.SCNSubscription, error) {
	d.t.GetSCNSubscription.Input.id = id
	return d.t.GetSCNSubscription.Return.sub, d.t.GetSCNSubscription.Return.err
}

// Insert a new SCN subscription. Existing subscriptions are unaffected
func (d *hmsdbtest) InsertSCNSubscription(sub sm.SCNPostSubscription) (int64, error) {
	d.t.InsertSCNSubscription.Input.sub = sub
	return d.t.InsertSCNSubscription.Return.id, d.t.InsertSCNSubscription.Return.err
}

// Update an existing SCN subscription.
func (d *hmsdbtest) UpdateSCNSubscription(id int64, sub sm.SCNPostSubscription) (bool, error) {
	d.t.UpdateSCNSubscription.Input.id = id
	d.t.UpdateSCNSubscription.Input.sub = sub
	return d.t.UpdateSCNSubscription.Return.didUpdate, d.t.UpdateSCNSubscription.Return.err
}

// Patch an existing SCN subscription.
func (d *hmsdbtest) PatchSCNSubscription(id int64, op string, patch sm.SCNPatchSubscription) (bool, error) {
	d.t.PatchSCNSubscription.Input.id = id
	d.t.PatchSCNSubscription.Input.op = op
	d.t.PatchSCNSubscription.Input.patch = patch
	return d.t.PatchSCNSubscription.Return.didPatch, d.t.PatchSCNSubscription.Return.err
}

// Delete a SCN subscription
func (d *hmsdbtest) DeleteSCNSubscription(id int64) (bool, error) {
	d.t.DeleteSCNSubscription.Input.id = id
	return d.t.DeleteSCNSubscription.Return.didDelete, d.t.DeleteSCNSubscription.Return.err
}

// Delete all SCN subscriptions
func (d *hmsdbtest) DeleteSCNSubscriptionsAll() (int64, error) {
	return d.t.DeleteSCNSubscriptionsAll.Return.numDelete, d.t.DeleteSCNSubscriptionsAll.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Group and Partition  Management
//
////////////////////////////////////////////////////////////////////////////

//
// Groups
//

// Create a group.  Returns new label (should match one in struct,
// unless case-normalized) if successful, otherwise empty string + non
// nil error.
func (d *hmsdbtest) InsertGroup(g *sm.Group) (string, error) {
	d.t.InsertGroup.Input.g = g
	return d.t.InsertGroup.Return.label, d.t.InsertGroup.Return.err
}

// Update group with label
func (d *hmsdbtest) UpdateGroup(label string, gp *sm.GroupPatch) error {
	d.t.UpdateGroup.Input.label = label
	d.t.UpdateGroup.Input.gp = gp
	return d.t.UpdateGroup.Return.err
}

// Get Group with given label.  Nil if not found and nil error, otherwise
// nil plus non-nil error (not normally expected)
// If filt_part is non-empty, the partition name is used to filter
// the members list.
func (d *hmsdbtest) GetGroup(label, filt_part string) (*sm.Group, error) {
	d.t.GetGroup.Input.label = label
	d.t.GetGroup.Input.filt_part = filt_part
	return d.t.GetGroup.Return.group, d.t.GetGroup.Return.err
}

// Get list of group labels (names).
func (d *hmsdbtest) GetGroupLabels() ([]string, error) {
	return d.t.GetGroupLabels.Return.labels, d.t.GetGroupLabels.Return.err
}

// Delete entire group with the given label.  If no error, bool indicates
// whether member was present to remove.
func (d *hmsdbtest) DeleteGroup(label string) (bool, error) {
	d.t.DeleteGroup.Input.label = label
	return d.t.DeleteGroup.Return.didDelete, d.t.DeleteGroup.Return.err
}

// Add member xname id to existing group label.  returns ErrHMSDSNoGroup
// if group with label does not exist, or ErrHMSDSDuplicateKey if Group
// is exclusive and xname id is already in a different group in this exclusive
// set.
// Returns key of new member, should be same as id after normalization,
// if any.  Label should already be normalized.
func (d *hmsdbtest) AddGroupMember(label, id string) (string, error) {
	d.t.AddGroupMember.Input.label = label
	d.t.AddGroupMember.Input.id = id
	return d.t.AddGroupMember.Return.id, d.t.AddGroupMember.Return.err
}

// Delete Group member from label.  If no error, bool indicates whether
// group was present to remove.
func (d *hmsdbtest) DeleteGroupMember(label, id string) (bool, error) {
	d.t.DeleteGroupMember.Input.label = label
	d.t.DeleteGroupMember.Input.id = id
	return d.t.DeleteGroupMember.Return.didDelete, d.t.DeleteGroupMember.Return.err
}

func (d *hmsdbtest) SetGroupMembers(label string, ids []string) ([]string, error) {
	return nil, nil
}

//
// Partitions
//

// Create a partition.  Returns new name (should match one in struct,
// unless case-normalized) if successful, otherwise empty string + non
// nil error.
func (d *hmsdbtest) InsertPartition(p *sm.Partition) (string, error) {
	d.t.InsertPartition.Input.p = p
	return d.t.InsertPartition.Return.pname, d.t.InsertPartition.Return.err
}

// Update Partition with given name
func (d *hmsdbtest) UpdatePartition(pname string, pp *sm.PartitionPatch) error {
	d.t.UpdatePartition.Input.pname = pname
	d.t.UpdatePartition.Input.pp = pp
	return d.t.UpdatePartition.Return.err
}

// Get partition with given name  Nil if not found and nil error, otherwise
// nil plus non-nil error (not normally expected)
func (d *hmsdbtest) GetPartition(pname string) (*sm.Partition, error) {
	d.t.GetPartition.Input.pname = pname
	return d.t.GetPartition.Return.partition, d.t.GetPartition.Return.err
}

// Get list of partition names.
func (d *hmsdbtest) GetPartitionNames() ([]string, error) {
	return d.t.GetPartitionNames.Return.pnames, d.t.GetPartitionNames.Return.err
}

// Delete entire partition with pname.  If no error, bool indicates
// whether partition was present to remove.
func (d *hmsdbtest) DeletePartition(pname string) (bool, error) {
	d.t.DeletePartition.Input.pname = pname
	return d.t.DeletePartition.Return.didDelete, d.t.DeletePartition.Return.err
}

// Add member xname id to existing partition.  returns ErrHMSDSNoGroup
// if partition name does not exist, or ErrHMSDSDuplicateKey if xname id
// is already in a different partition.
// Returns key of new member, should be same as id after normalization,
// if any.  pname should already be normalized.
func (d *hmsdbtest) AddPartitionMember(pname, id string) (string, error) {
	d.t.AddPartitionMember.Input.pname = pname
	d.t.AddPartitionMember.Input.id = id
	return d.t.AddPartitionMember.Return.id, d.t.AddPartitionMember.Return.err
}

// Delete partition member from partition.  If no error, bool indicates
// whether member was present to remove.
func (d *hmsdbtest) DeletePartitionMember(pname, id string) (bool, error) {
	d.t.DeletePartitionMember.Input.pname = pname
	d.t.DeletePartitionMember.Input.id = id
	return d.t.DeletePartitionMember.Return.didDelete, d.t.DeletePartitionMember.Return.err
}

//
// Memberships
//

// Get the memberships for a particular component xname id
func (d *hmsdbtest) GetMembership(id string) (*sm.Membership, error) {
	d.t.GetMembership.Input.id = id
	return d.t.GetMembership.Return.membership, d.t.GetMembership.Return.err
}

// Get all memberships, optionally filtering
// Convenience feature - not needed for initial implementation
func (d *hmsdbtest) GetMemberships(f *hmsds.ComponentFilter) ([]*sm.Membership, error) {
	d.t.GetMemberships.Input.f = f
	return d.t.GetMemberships.Return.memberships, d.t.GetMemberships.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Component Lock Management
//
////////////////////////////////////////////////////////////////////////////

// Create component reservations if one doesn't already exist.
// To create reservations without a duration, the component must be locked.
// To create reservations with a duration, the component must be unlocked.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbtest) InsertCompReservations(f sm.CompLockV2Filter) (sm.CompLockV2ReservationResult, error) {
	d.t.InsertCompReservations.Input.f = f
	return d.t.InsertCompReservations.Return.results, d.t.InsertCompReservations.Return.err
}

// Forcebly remove/release component reservations.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbtest) DeleteCompReservationsForce(f sm.CompLockV2Filter) (sm.CompLockV2UpdateResult, error) {
	d.t.DeleteCompReservationsForce.Input.f = f
	return d.t.DeleteCompReservationsForce.Return.results, d.t.DeleteCompReservationsForce.Return.err
}

// Remove/release component reservations.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbtest) DeleteCompReservations(f sm.CompLockV2ReservationFilter) (sm.CompLockV2UpdateResult, error) {
	d.t.DeleteCompReservations.Input.f = f
	return d.t.DeleteCompReservations.Return.results, d.t.DeleteCompReservations.Return.err
}

// Release all expired reservations
func (d *hmsdbtest) DeleteCompReservationsExpired() ([]string, error) {
	return d.t.DeleteCompReservationsExpired.Return.ids, d.t.DeleteCompReservationsExpired.Return.err
}

// Retrieve the status of reservations. The public key and xname is
// required to address the reservation.
func (d *hmsdbtest) GetCompReservations(dkeys []sm.CompLockV2Key) (sm.CompLockV2ReservationResult, error) {
	d.t.GetCompReservations.Input.dkeys = dkeys
	return d.t.GetCompReservations.Return.results, d.t.GetCompReservations.Return.err
}

// Update/renew the expiration time of component reservations with the given
// ID/Key combinations.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbtest) UpdateCompReservations(f sm.CompLockV2ReservationFilter) (sm.CompLockV2UpdateResult, error) {
	d.t.UpdateCompReservations.Input.f = f
	return d.t.UpdateCompReservations.Return.results, d.t.UpdateCompReservations.Return.err
}

// Retrieve component lock information.
func (d *hmsdbtest) GetCompLocksV2(f sm.CompLockV2Filter) ([]sm.CompLockV2, error) {
	d.t.GetCompLocksV2.Input.f = f
	return d.t.GetCompLocksV2.Return.cls, d.t.GetCompLocksV2.Return.err
}

// Update component locks. Valid actions are 'Lock', 'Unlock', 'Disable',
// and 'Repair'.
// 'Lock'\'Unlock' updates the 'locked' status of the components.
// 'Disable'\'Repair' updates the 'reservationsDisabled' status of components.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbtest) UpdateCompLocksV2(f sm.CompLockV2Filter, action string) (sm.CompLockV2UpdateResult, error) {
	d.t.UpdateCompLocksV2.Input.f = f
	d.t.UpdateCompLocksV2.Input.action = action
	return d.t.UpdateCompLocksV2.Return.results, d.t.UpdateCompLocksV2.Return.err
}

////////////////////////////////////////////////////////////////////////////
//
// Job Sync Management
//
////////////////////////////////////////////////////////////////////////////

//
// Jobs
//

// Create a job entry in the job sync. Returns new jobId if successful,
// otherwise non-nil error.
func (d *hmsdbtest) InsertJob(j *sm.Job) (string, error) {
	d.t.InsertJob.Input.j = j
	return d.t.InsertJob.Return.id, d.t.InsertJob.Return.err
}

// Update the status of the job with the given jobId.
func (d *hmsdbtest) UpdateJob(jobId, status string) (bool, error) {
	d.t.UpdateJob.Input.jobId = jobId
	d.t.UpdateJob.Input.status = status
	return d.t.UpdateJob.Return.didUpdate, d.t.UpdateJob.Return.err
}

// Get the job sync entry with the given job id. Nil if not found and nil
// error, otherwise non-nil error (not normally expected).
func (d *hmsdbtest) GetJob(jobId string) (*sm.Job, error) {
	d.t.GetJob.Input.jobId = jobId
	return d.t.GetJob.Return.j, d.t.GetJob.Return.err
}

// Get list of jobs from the job sync.
func (d *hmsdbtest) GetJobs(f_opts ...hmsds.JobSyncFiltFunc) ([]*sm.Job, error) {
	f := new(hmsds.JobSyncFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	d.t.GetJobs.Input.f = f
	return d.t.GetJobs.Return.js, d.t.GetJobs.Return.err
}

// Delete the job entry with the given jobId. If no error, bool indicates
// whether component lock was present to remove.
func (d *hmsdbtest) DeleteJob(jobId string) (bool, error) {
	d.t.DeleteJob.Input.jobId = jobId
	return d.t.DeleteJob.Return.didDelete, d.t.DeleteJob.Return.err
}
