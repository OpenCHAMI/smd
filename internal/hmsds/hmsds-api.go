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

package hmsds

import (
	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
)

var e = base.NewHMSError("hmsds", "GenericError")

var ErrHMSDSBadSchema = e.NewChild("Not yet running the expected schema version")

var ErrHMSDSArgNil = e.NewChild("HMSDS method arg is nil")
var ErrHMSDSPtrClosed = e.NewChild("HMSDS handle is not open.")
var ErrHMSDSTxFailed = e.NewChild("HMSDS transaction could not be started")
var ErrHMSDSArgMissing = e.NewChild("a required argument was missing")
var ErrHMSDSArgNoMatch = e.NewChild("a required argument did not match any valid input")
var ErrHMSDSArgNotAnInt = e.NewChild("a required argument was not an integer")
var ErrHMSDSArgMissingNID = e.NewChild("NID value missing.  Use NID < 0 to unset.")
var ErrHMSDSArgBadRange = e.NewChild("An argument was out of range")
var ErrHMSDSArgTooMany = e.NewChild("too many arguments")
var ErrHMSDSArgEmpty = e.NewChild("an argument was the empty string")

var ErrHMSDSArgBadArg = e.NewChild("Argument was not valid")
var ErrHMSDSArgBadID = e.NewChild("Argument was not a valid xname ID")
var ErrHMSDSArgBadType = e.NewChild("Argument was not a valid HMS Type")
var ErrHMSDSArgBadState = e.NewChild("Argument was not a valid HMS State")
var ErrHMSDSArgBadFlag = e.NewChild("Argument was not a valid HMS Flag")
var ErrHMSDSArgBadRole = e.NewChild("Argument was not a valid HMS Role")
var ErrHMSDSArgBadSubRole = e.NewChild("Argument was not a valid HMS SubRole")
var ErrHMSDSArgBadArch = e.NewChild("Argument was not a valid HMS Arch")
var ErrHMSDSArgBadClass = e.NewChild("Argument was not a valid HMS Class")
var ErrHMSDSArgBadSubtype = e.NewChild("Argument was not a valid HMS Subtype")
var ErrHMSDSArgBadRedfishType = e.NewChild("Argument was not a valid Redfish type")
var ErrHMSDSArgBadJobType = e.NewChild("Argument was not a valid job Type")
var ErrHMSDSArgBadHWInvHistEventType = e.NewChild("Argument was not a HWInvHist event Type")
var ErrHMSDSArgBadTimeFormat = e.NewChild("Argument was not in a valid RFC3339 time format")

var ErrHMSDSDuplicateKey = e.NewChild("Would create a duplicate key or non-unique field")
var ErrHMSDSNoComponent = e.NewChild("linked component does not exist")
var ErrHMSDSNoREP = e.NewChild("One or more RedfishEndpoints do not exist")

var ErrHMSDSNoGroup = e.NewChild("no such group")
var ErrHMSDSNoPartition = e.NewChild("no such partition")
var ErrHMSDSExclusiveGroup = e.NewChild("Would create a duplicate key in another exclusive group")
var ErrHMSDSExclusivePartition = e.NewChild("Would create a duplicate key in another partition")

var ErrHMSDSMultipleGroupAndPart = e.NewChild("group and partition cannot both have more than one value")
var ErrHMSDSNullGroupBadPart = e.NewChild("NULL group and non-NULL partition arg not permitted")
var ErrHMSDSNullPartBadGroup = e.NewChild("NULL partition and non-NULL group arg not permitted")
var ErrHMSDSNullBadMixGroup = e.NewChild("NULL and non-NULL group arguments not permitted")
var ErrHMSDSNullBadMixPart = e.NewChild("NULL and non-NULL partrition arguments not permitted")

var ErrHMSDSNoCompLock = e.NewChild("no such component lock")
var ErrHMSDSExclusiveCompLock = e.NewChild("Would create a lock on an already locked component")
var ErrHMSDSInvalidCompLockAction = e.NewChild("Invalid action for updating component locks")

var ErrHMSDSNoCompEthInterface = e.NewChild("no such component ethernet interface")
var ErrHMSDSCompEthInterfaceMultipleIPs = e.NewChild("component ethernet interface with multiple IP Addresses")

var ErrHMSDSNoJobData = e.NewChild("Job has no data")

type LogLevel int

const (
	LOG_DEFAULT LogLevel = 0
	LOG_NOTICE  LogLevel = 1
	LOG_INFO    LogLevel = 2
	LOG_DEBUG   LogLevel = 3
	LOG_LVL_MAX LogLevel = 4
)

type HMSDSPatchOp int

const (
	PatchOpInvalid HMSDSPatchOp = 0
	PatchOpAdd     HMSDSPatchOp = 1
	PatchOpRemove  HMSDSPatchOp = 2
	PatchOpReplace HMSDSPatchOp = 3
)

var hmsdsPatchOpMap = map[string]HMSDSPatchOp{
	"add":     PatchOpAdd,
	"remove":  PatchOpRemove,
	"replace": PatchOpReplace,
}

const (
	CLUpdateActionLock    = "Lock"
	CLUpdateActionUnlock  = "Unlock"
	CLUpdateActionDisable = "Disable"
	CLUpdateActionRepair  = "Repair"
)

type HMSDSErrInfo struct {
	UserErr     string
	UserErrArgs string
}

// Identify group, partition info
type PartInfo struct {
	Group     []string `json:"Group"`
	Partition []string `json:"Partition"`
}

type HMSDB interface {

	// Return implementation name as a string
	ImplementationName() string

	// Open connection.  Normally only needs to be done once, as it maintains
	// a collection pool that can be used by multiple Go routines.
	// Details required to establish the connection are implementation
	// dependent and are supplied during the creation of the interface
	// object.  Hence, we don't need to supply any connection details here.
	Open() error

	// Closes the database connection.  This is a global operation that
	// affects all go routines using a hmsdb handle.  It is only used when
	// we are done with the DB entirely (fine-grained management is
	// not needed for individual DB calls).
	Close() error

	// Create a new transaction for multi-part atomic operations.
	// The individual operations are defined under the HMSDBTx interface
	// below (remember to call rollback or commit on the HMSDSTx struct
	// when done).
	Begin() (HMSDBTx, error)

	// Test the database connection to make sure that it is healthy
	TestConnection() error

	// Increase verbosity for debugging, etc.
	SetLogLevel(lvl LogLevel) error

	//                                                                    //
	//          Generic single-valed queries - get id/key, etc.           //
	//                                                                    //

	// Build filter query for Component IDs using filter functions and
	// then return the list of matching xname IDs as a string array, write
	// locking the rows if requested.
	GetComponentIDs(f_opts ...CompFiltFunc) ([]string, error)

	// Build filter query for ComponentEndpoints IDs using filter functions and
	// then return the list of matching xname IDs as a string array, write
	// locking the rows if requested.
	GetCompEndpointIDs(f_opts ...CompEPFiltFunc) ([]string, error)

	// Build filter query for RedfishEndpoints IDs using filter functions and
	// then return the list of matching xname IDs as a string array, write
	// locking the rows if requested.
	GetRFEndpointIDs(f_opts ...RedfishEPFiltFunc) ([]string, error)

	//                                                                    //
	//        HMS Components - Managed plane info: State, NID, Role       //
	//                                                                    //

	// Look up a single component by id, i.e. xname
	GetComponentByID(id string) (*base.Component, error)

	// Get all HMS Components in system.
	GetComponentsAll() ([]*base.Component, error)

	// Get some or all HMS Components in system, with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetComponentsFilter(f *ComponentFilter, fieldFltr FieldFilter) ([]*base.Component, error)

	// Get some or all HMS Components in system under
	// a set of parent components, with filtering options to possibly
	// narrow the returned values. If no filter provided, just get
	// the parent components.  Otherwise use it to create a custom
	// WHERE... string that filters out entries that do not match ALL
	// of the non-empty strings in the filter struct.
	GetComponentsQuery(f *ComponentFilter, fieldfltr FieldFilter, ids []string) ([]*base.Component, error)

	// Get a single component by its NID, if the NID exists.
	GetComponentByNID(nid string) (*base.Component, error)

	// Insert HMS Component into database, updating it if it exists.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	InsertComponent(c *base.Component) (int64, error)

	// Inserts or updates HMS ComponentArray entries in database within a
	// single all-or-none transaction.
	InsertComponents(comps *base.ComponentArray) ([]string, error)

	// Inserts or updates ComponentArray entries in database within a single
	// all-or-none transaction. If force=true, only the state, flag, subtype,
	// nettype, and arch will be overwritten for existing components. Otherwise,
	// this won't overwrite existing components.
	UpsertComponents(comps []*base.Component, force bool) (map[string]map[string]bool, error)

	// Update state and flag fields only in DB for the given IDs.  If
	// len(ids) is > 1 a locking read will be done to ensure the list o
	// components that was actually modified is always returned.
	//
	// If force = true ignores any starting state restrictions and will
	// always set ids to 'state', unless it is already set.
	//   Note: If flag is not set, it will be set to OK (i.e. no flag)
	UpdateCompStates(ids []string, state string, flag string, force bool, pi *PartInfo) ([]string, error)

	// Update Flag field in DB from c's Flag field.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: Flag cannot be empty/invalid.
	UpdateCompFlagOnly(id string, flag string) (int64, error)

	// Update flag field in DB for a list of components
	// Note: Flag cannot be empty/invalid.
	BulkUpdateCompFlagOnly(ids []string, flag string) ([]string, error)

	// Update enabled field in DB from c's Enabled field.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: c.Enabled cannot be nil.
	UpdateCompEnabled(id string, enabled bool) (int64, error)

	// Update Enabled field only in DB for a list of components
	BulkUpdateCompEnabled(ids []string, enabled bool) ([]string, error)

	// Update SwStatus field in DB from c's SwStatus field.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	UpdateCompSwStatus(id string, swStatus string) (int64, error)

	// Update SwStatus field only in DB for a list of components
	BulkUpdateCompSwStatus(ids []string, swstatus string) ([]string, error)

	// Update Role/SubRole field in DB from c's Role/SubRole field.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: Role cannot be blank/invalid.
	UpdateCompRole(id string, role, subRole string) (int64, error)

	// Update Role/SubRole field in DB for a list of components
	// Note: Role cannot be blank/invalid.
	BulkUpdateCompRole(ids []string, role, subRole string) ([]string, error)

	// Update Class field only in DB for a list of components
	BulkUpdateCompClass(ids []string, class string) ([]string, error)

	// Update NID field in DB from c's NID field.
	// Note: NID cannot be blank.  Should be negative to unset.
	UpdateCompNID(c *base.Component) error

	// Update NID field in DB for a list of components
	// Note: NID cannot be blank.  Should be negative to unset.
	BulkUpdateCompNID(comps *[]base.Component) error

	// Delete HMS Component with matching xname id from database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteComponentByID(id string) (bool, error)

	// Delete all HMS Components from database (atomically)
	// Also returns number of deleted rows, if error is nil.
	DeleteComponentsAll() (int64, error)

	//                                                                    //
	//              Node to Default NID, role, etc. mapping               //
	//                                                                    //

	// Look up one Node->NID Mapping by id, i.e. node xname.
	GetNodeMapByID(id string) (*sm.NodeMap, error)

	// Look up ALL Node->NID Mappings.
	GetNodeMapsAll() ([]*sm.NodeMap, error)

	// Insert Node->NID Mapping into database, updating it if it exists.
	InsertNodeMap(m *sm.NodeMap) error

	// Inserts or updates Node->NID Mapping Array entries in database within a
	// single all-or-none transaction.
	InsertNodeMaps(nnms *sm.NodeMapArray) error

	// Delete Node NID Mapping entry with matching xname id from database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteNodeMapByID(id string) (bool, error)

	// Delete all Node NID Mapping entries from database.
	// Also returns number of deleted rows, if error is nil.
	DeleteNodeMapsAll() (int64, error)

	//                                                                    //
	//                           Power mapping                            //
	//                                                                    //

	// Look up one Power Mapping by id, i.e. node xname.
	GetPowerMapByID(id string) (*sm.PowerMap, error)

	// Look up ALL Power Mappings.
	GetPowerMapsAll() ([]*sm.PowerMap, error)

	// Insert Power Mapping into database, updating it if it exists.
	InsertPowerMap(m *sm.PowerMap) error

	// Inserts or updates Power Mapping Array entries in database within a
	// single all-or-none transaction.
	InsertPowerMaps(ms []sm.PowerMap) error

	// Delete Power Mapping entry with matching xname id from database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeletePowerMapByID(id string) (bool, error)

	// Delete all Power Mapping entries from database.
	// Also returns number of deleted rows, if error is nil.
	DeletePowerMapsAll() (int64, error)

	//                                                                    //
	//        Hardware Inventory - Detailed location and FRU info         //
	//                                                                    //

	// Get some or all Hardware Inventory entries with filtering
	// options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct.
	// This does hierarchy searches.
	GetHWInvByLocQueryFilter(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error)

	// Get some or all Hardware Inventory entries with filtering
	// options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetHWInvByLocFilter(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error)

	// Get a single Hardware inventory entry by current xname
	// This struct includes the FRU info if the xname is currently populated.
	GetHWInvByLocID(id string) (*sm.HWInvByLoc, error)

	// Get HWInvByLoc by primary key (xname) for all entries in the system.
	// It also pairs the data with the matching HWInvByFRU if the xname is
	// populated.
	GetHWInvByLocAll() ([]*sm.HWInvByLoc, error)

	// Get HW Inventory-by-FRU entry at the provided location FRU ID
	GetHWInvByFRUID(fruid string) (*sm.HWInvByFRU, error)

	// Get some or all HW-inventory-by-FRU entries with filtering
	// options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetHWInvByFRUFilter(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByFRU, error)

	// Get all HW-inventory-by-FRU entries.
	GetHWInvByFRUAll() ([]*sm.HWInvByFRU, error)

	// Insert or update HWInventoryByLocation struct.
	// If PopulatedFRU is present, this is also added to the DB  If
	// it is not, this effectively "depopulates" the given location.
	// The actual HWInventoryByFRU is stored using within the same
	// transaction.
	InsertHWInvByLoc(hl *sm.HWInvByLoc) error

	// Insert or update HWInventoryByFRU struct.  This does not associate
	// the object with any HW-Inventory-By-Location info so it is
	// typically not needed.  InsertHWInvByLoc is typically used to
	// store both type of info at once.
	InsertHWInvByFRU(hf *sm.HWInvByFRU) error

	// Insert or update array of HWInventoryByLocation structs.
	// If PopulatedFRU is present, these is also added to the DB  If
	// it is not, this effectively "depopulates" the given locations.
	// The actual HWInventoryByFRU is stored using within the same
	// transaction.
	InsertHWInvByLocs(hls []*sm.HWInvByLoc) error

	// Delete HWInvByLoc entry with matching xname id from database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteHWInvByLocID(id string) (bool, error)

	// Delete ALL HWInvByLoc entries from database (atomically)
	// Also returns number of deleted rows, if error is nil.
	DeleteHWInvByLocsAll() (int64, error)

	// Delete HWInvByFRU entry with matching FRU ID from database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteHWInvByFRUID(fruid string) (bool, error)

	// Delete ALL HWInvByFRU entries from database (atomically)
	// Also returns number of deleted rows, if error is nil.
	DeleteHWInvByFRUsAll() (int64, error)

	//                                                                    //
	//   Hardware Inventory History - Detailed history of hardware FRU    //
	//                                location.                           //
	//                                                                    //

	// Get hardware history for some or all Hardware Inventory entries with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetHWInvHistFilter(f_opts ...HWInvHistFiltFunc) ([]*sm.HWInvHist, error)

	// Get only the most recent hardware history event for some or all hardware
	// locations.
	GetHWInvHistLastEvents(ids []string) ([]*sm.HWInvHist, error)

	// Insert a HWInventoryHistory entry.
	// If a duplicate is present return an error.
	InsertHWInvHist(hh *sm.HWInvHist) error

	// Insert an array of HWInventoryHistory entries.
	// If a duplicate is present return an error.
	InsertHWInvHists(hhs []*sm.HWInvHist) error

	// Delete all HWInvHist entries with matching xname id from database, if it
	// exists.
	// Returns the number of deleted rows, if error is nil.
	DeleteHWInvHistByLocID(id string) (int64, error)

	// Delete all HWInvHist entries with matching FRU id from database, if it
	// exists.
	// Returns the number of deleted rows, if error is nil.
	DeleteHWInvHistByFRUID(fruid string) (int64, error)

	// Delete all HWInvHist entries from database (atomically)
	// Returns the number of deleted rows, if error is nil.
	DeleteHWInvHistAll() (int64, error)

	// Delete all HWInvHist entries from database matching a filter.
	// Returns the number of deleted rows, if error is nil.
	DeleteHWInvHistFilter(f_opts ...HWInvHistFiltFunc) (int64, error)

	//                                                                    //
	//    Redfish Endpoints - Redfish service roots used for discovery    //
	//                                                                    //

	// Get RedfishEndpoint by ID (xname), i.e. a single entry.
	GetRFEndpointByID(id string) (*sm.RedfishEndpoint, error)

	// Get all RedfishEndpoints in system.
	GetRFEndpointsAll() ([]*sm.RedfishEndpoint, error)

	// Get some or all RedfishEndpoints in system, with filtering
	// options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetRFEndpointsFilter(f *RedfishEPFilter) ([]*sm.RedfishEndpoint, error)

	// Insert new RedfishEndpoint into database.
	// Does not insert any ComponentEndpoint children.
	// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
	// No insertion done on err != nil
	InsertRFEndpoint(ep *sm.RedfishEndpoint) error

	// Insert new RedfishEndpointArray into database within a single
	// all-or-none transaction.  Does not insert any ComponentEndpoint
	// children.
	// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
	// No insertions are done on err != nil
	InsertRFEndpoints(eps *sm.RedfishEndpointArray) error

	// Update existing RedfishEndpointArray entry in database.
	// Does not update any ComponentEndpoint children.
	// Returns updated entry or nil/nul if not found.  If an error occurred,
	// nil/error will be returned.
	UpdateRFEndpoint(ep *sm.RedfishEndpoint) (*sm.RedfishEndpoint, error)

	// Update existing RedfishEndpointArray entry in database, but only updates
	// fields that would be changed by a user-directed operation.
	// Does not update any ComponentEndpoint children.
	// Returns updated entry or nil/nil if not found.  If an error occurred,
	// nil/error will be returned.
	UpdateRFEndpointNoDiscInfo(ep *sm.RedfishEndpoint) (*sm.RedfishEndpoint, []string, error)

	// Patch existing RedfishEndpointArray entry in database, but only updates
	// specified fields.
	// Does not update any ComponentEndpoint children.
	// Returns updated entry or nil/nil if not found.  If an error occurred,
	// nil/error will be returned.
	PatchRFEndpointNoDiscInfo(id string, epp sm.RedfishEndpointPatch) (*sm.RedfishEndpoint, []string, error)

	// Returns: Discoverable endpoint list, with status set appropriately in DB
	// and return values.  However this list will omit those RF EPs  who are
	// already being discovered, unless forced.
	// Error returned on unexpected failure or any entry in eps not existing,
	// the latter error being ErrHMSDSNoREP.
	UpdateRFEndpointForDiscover(ids []string, force bool) ([]*sm.RedfishEndpoint, error)

	// Update existing RedfishEndpointArray entries in database within a
	// single all-or-none transaction.  Does not update any ComponentEndpoint
	// children.
	// Returns FALSE with err == nil if one or more updated entries do
	// not exist.  No updates are performed in this case.
	UpdateRFEndpoints(eps *sm.RedfishEndpointArray) (bool, error)

	// Delete RedfishEndpoint with matching xname id from database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteRFEndpointByID(id string) (bool, error)

	// Delete all RedfishEndpoints from database.
	// Also returns number of deleted rows, if error is nil.
	DeleteRFEndpointsAll() (int64, error)

	// Delete RedfishEndpoint with matching xname id from database, if it
	// exists.  When dooing so, set all HMS Components to Empty if they
	// are children of the RedfishEndpoint.
	// Return true if there was a row affected, false if there were zero.
	DeleteRFEndpointByIDSetEmpty(id string) (bool, []string, error)

	// Delete all RedfishEndpoints from database.
	// This also deletes all child ComponentEndpoints, and in addition,
	// sets the State/Components entries for those ComponentEndpoints to Empty/OK
	// Also returns number of deleted rows, if error is nil.
	DeleteRFEndpointsAllSetEmpty() (int64, []string, error)

	//                                                                    //
	// ComponentEndpoints: Component info discovered from Parent          //
	//                     RedfishEndpoint.  Management plane equivalent  //
	//                     to HMS Component.                              //
	//                                                                    //

	// Get ComponentEndpoint by id (xname), i.e. a single entry.
	GetCompEndpointByID(id string) (*sm.ComponentEndpoint, error)

	// Get all ComponentEndpoints in system.
	GetCompEndpointsAll() ([]*sm.ComponentEndpoint, error)

	// Get some or all ComponentEndpoints in system, with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetCompEndpointsFilter(f *CompEPFilter) ([]*sm.ComponentEndpoint, error)

	// Upsert ComponentEndpoint into database, updating it if it exists.
	UpsertCompEndpoint(cep *sm.ComponentEndpoint) error

	// Upsert ComponentEndpointArray into database within a single all-or-none
	// transaction.
	UpsertCompEndpoints(ceps *sm.ComponentEndpointArray) error

	// Delete ComponentEndpoint with matching xname id from database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteCompEndpointByID(id string) (bool, error)

	// Delete all ComponentEndpoints from database.
	// Also returns number of deleted rows, if error is nil.
	DeleteCompEndpointsAll() (int64, error)

	// Delete ComponentEndpoint with matching xname id from database, if it
	// exists.  When dooing so, set the corresponding HMS Component to Empty if it
	// is not already in that state.
	// Return true if there was a row affected, false if there were zero.  The
	// string array returns the single xname ID that changed state or is empty.
	DeleteCompEndpointByIDSetEmpty(id string) (bool, []string, error)

	// Delete all ComponentEndpoints from database. In addition,
	// sets the State/Components entry for each ComponentEndpoint to Empty/OK
	// Also returns number of deleted rows, if error is nil, and also string array
	// of those xname IDs that were set to Empty/OK (i.e. not already Empty/OK)
	// as part of the deletion.
	DeleteCompEndpointsAllSetEmpty() (int64, []string, error)

	//                                                                    //
	// ServiceEndpoints: Service info discovered from Parent              //
	//                   RedfishEndpoint.                                 //
	//                                                                    //

	// Get ServiceEndpoint by service and id (xname), i.e. a single entry.
	GetServiceEndpointByID(svc, id string) (*sm.ServiceEndpoint, error)

	// Get all ServiceEndpoints in system.
	GetServiceEndpointsAll() ([]*sm.ServiceEndpoint, error)

	// Get some or all ServiceEndpoints in system, with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetServiceEndpointsFilter(f *ServiceEPFilter) ([]*sm.ServiceEndpoint, error)

	// Upsert ServiceEndpoint into database, updating it if it exists.
	UpsertServiceEndpoint(sep *sm.ServiceEndpoint) error

	// Upsert ServiceEndpointArray into database within a single all-or-none
	// transaction.
	UpsertServiceEndpoints(seps *sm.ServiceEndpointArray) error

	// Delete ServiceEndpoint with matching service type and xname id from
	// database, if it exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteServiceEndpointByID(svc, id string) (bool, error)

	// Delete all ServiceEndpoints from database.
	// Also returns number of deleted rows, if error is nil.
	DeleteServiceEndpointsAll() (int64, error)

	//                                                                    //
	//    Component Ethernet Interfaces - MAC address to IP address       //
	//        relations for component endpoint ethernet interfaces        //
	//                                                                    //

	// Get some or all CompEthInterfaces in the system, with filtering
	// options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetCompEthInterfaceFilter(f_opts ...CompEthInterfaceFiltFunc) ([]*sm.CompEthInterfaceV2, error)

	// Insert a new CompEthInterface into the database.
	// If ID or MAC address already exists, return ErrHMSDSDuplicateKey
	// No insertion done on err != nil
	InsertCompEthInterface(cei *sm.CompEthInterfaceV2) error

	// Insert new CompEthInterfaces into the database within a single
	// all-or-none transaction.
	// If ID or MAC address already exists, return ErrHMSDSDuplicateKey
	// No insertions are done on err != nil
	InsertCompEthInterfaces(ceis []*sm.CompEthInterfaceV2) error

	// Insert/update a CompEthInterface in the database.
	// If ID or MAC address already exists, only overwrite ComponentID
	// and Type fields.
	// No insertion done on err != nil
	InsertCompEthInterfaceCompInfo(cei *sm.CompEthInterfaceV2) error

	// Insert new CompEthInterfaces into database within a single
	// all-or-none transaction.
	// If ID or MAC address already exists, only overwrite ComponentID
	// and Type fields.
	// No insertions are done on err != nil
	InsertCompEthInterfacesCompInfo(ceis []*sm.CompEthInterfaceV2) error

	// Update existing CompEthInterface entry in the database, but only updates
	// fields that would be changed by a user-directed operation.
	// Returns updated entry or nil/nil if not found.  If an error occurred,
	// nil/error will be returned.
	UpdateCompEthInterface(id string, ceip *sm.CompEthInterfaceV2Patch) (*sm.CompEthInterfaceV2, error)

	// Update existing CompEthInterface entry in the database, but only updates
	// fields that would be changed by a user-directed operation.
	// Returns updated entry or nil/nil if not found.  If an error occurred,
	// nil/error will be returned.
	//
	// Special handling is required to use the V1 API Patch on a V2 CompEthInterface.
	// If the CEI has more than 2 or more IP addresses associated with it the error
	// CompEthInterfacePatch will be ErrHMSDSCompEthInterfaceMultipleIPs returned.
	UpdateCompEthInterfaceV1(id string, ceip *sm.CompEthInterfacePatch) (*sm.CompEthInterfaceV2, error)

	// Delete CompEthInterface with matching id from the database, if it
	// exists.
	// Return true if there was a row affected, false if there were zero.
	DeleteCompEthInterfaceByID(id string) (bool, error)

	// Delete all CompEthInterfaces from the database.
	// Also returns number of deleted rows, if error is nil.
	DeleteCompEthInterfacesAll() (int64, error)

	// Add IP Address mapping to the existing component ethernet interface.
	// returns:
	//	- ErrHMSDSNoCompEthInterface if the parent component ethernet interface
	// 	- ErrHMSDSDuplicateKey if the parent component ethernet interface already
	//    has that IP address
	//
	// Returns key of new IP Address Mapping id, should be the IP address
	AddCompEthInterfaceIPAddress(id string, ipm *sm.IPAddressMapping) (string, error)

	// Update existing IP Address Mapping for a CompEthInterface entry in the database,
	// but only updates fields that would be changed by a user-directed operation.
	// Returns updated entry or nil/nil if not found.  If an error occurred,
	// nil/error will be returned.
	UpdateCompEthInterfaceIPAddress(id, ipAddr string, ipmPatch *sm.IPAddressMappingPatch) (*sm.IPAddressMapping, error)

	// Delete IP Address mapping from the Component Ethernet Interface.
	// If no error, bool indicates whether the IP Address Mapping was present to remove.
	DeleteCompEthInterfaceIPAddress(id, ipAddr string) (bool, error)

	//                                                                    //
	//           DiscoveryStatus - Discovery Status tracking               //
	//                                                                    //

	// Get DiscoveryStatus with the given numerical ID.
	GetDiscoveryStatusByID(id uint) (*sm.DiscoveryStatus, error)

	// Get all DiscoveryStatus entries.
	GetDiscoveryStatusAll() ([]*sm.DiscoveryStatus, error)

	// Update discovery status in DB.
	UpsertDiscoveryStatus(stat *sm.DiscoveryStatus) error

	//                                                                    //
	//        Discovery operations - Multi-type atomic operations.        //
	//                                                                    //

	// Atomically:
	//
	// 1. Update discovery-writable fields for RedfishEndpoint
	// 2. Upsert ComponentEndpointArray entries into database within the
	//    same transaction.
	// 3. Insert or update array of HWInventoryByLocation structs.
	//    If PopulatedFRU is present, these is also added to the DB  If
	//    it is not, this effectively "depopulates" the given locations.
	//    The actual HWInventoryByFRU is stored using within the same
	//    transaction.
	// 4. Inserts or updates HMS Components entries in ComponentArray
	//
	UpdateAllForRFEndpoint(
		ep *sm.RedfishEndpoint,
		ceps *sm.ComponentEndpointArray,
		hls []*sm.HWInvByLoc,
		comps *base.ComponentArray,
		seps *sm.ServiceEndpointArray,
		ceis []*sm.CompEthInterfaceV2,
	) (*[]base.Component, error)

	//                                                                    //
	//           SCNSubscription: SCN subscription management             //
	//                                                                    //

	// Get all SCN subscriptions
	GetSCNSubscriptionsAll() (*sm.SCNSubscriptionArray, error)

	// Get all SCN subscriptions
	GetSCNSubscription(id int64) (*sm.SCNSubscription, error)

	// Insert a new SCN subscription. Existing subscriptions are unaffected.
	InsertSCNSubscription(sub sm.SCNPostSubscription) (int64, error)

	// Update an existing SCN subscription.
	UpdateSCNSubscription(id int64, sub sm.SCNPostSubscription) (bool, error)

	// Patch an existing SCN subscription
	PatchSCNSubscription(id int64, op string, patch sm.SCNPatchSubscription) (bool, error)

	// Delete a SCN subscription
	DeleteSCNSubscription(id int64) (bool, error)

	// Delete all SCN subscriptions
	DeleteSCNSubscriptionsAll() (int64, error)

	//                                                                    //
	//                 Group and Partition  Management                    //
	//                                                                    //

	//                           Groups

	// Create a group.  Returns new label (should match one in struct,
	// unless case-normalized) if successful, otherwise empty string + non
	// nil error. Will return ErrHMSDSDuplicateKey if group exits or is
	// exclusive and xname id is already in another group in this exclusive set.
	// In addition, returns ErrHMSDSNoComponent if a component doesn't exist.
	InsertGroup(g *sm.Group) (string, error)

	// Update group with label
	UpdateGroup(label string, gp *sm.GroupPatch) error

	// Get Group with given label.  Nil if not found and nil error, otherwise
	// nil plus non-nil error (not normally expected)
	// If filt_part is non-empty, the partition name is used to filter
	// the members list.
	GetGroup(label, filt_part string) (*sm.Group, error)

	// Get list of group labels (names).
	GetGroupLabels() ([]string, error)

	// Delete entire group with the given label.  If no error, bool indicates
	// whether member was present to remove.
	DeleteGroup(label string) (bool, error)

	// Add member xname id to existing group label.  returns ErrHMSDSNoGroup
	// if group with label does not exist, or ErrHMSDSDuplicateKey if Group
	// is exclusive and xname id is already in another group in this exclusive set.
	// In addition, returns ErrHMSDSNoComponent if the component doesn't exist.
	//
	// Returns key of new member id, should be same as id after normalization,
	// if any.  Label should already be normalized.
	AddGroupMember(label, id string) (string, error)

	// Delete Group member from label.  If no error, bool indicates whether
	// group was present to remove.
	DeleteGroupMember(label, id string) (bool, error)

	//                        Partitions

	// Create a partition.  Returns new name (should match one in struct,
	// unless case-normalized) if successful, otherwise empty string + non
	// nil error.  Will return ErrHMSDSDuplicateKey if partition exits or an
	// xname id already exists in another partition.
	// In addition, returns ErrHMSDSNoComponent if a component doesn't exist.
	InsertPartition(p *sm.Partition) (string, error)

	// Update Partition with given name
	UpdatePartition(pname string, pp *sm.PartitionPatch) error

	// Get partition with given name  Nil if not found and nil error, otherwise
	// nil plus non-nil error (not normally expected)
	GetPartition(pname string) (*sm.Partition, error)

	// Get list of partition names.
	GetPartitionNames() ([]string, error)

	// Delete entire partition with pname.  If no error, bool indicates
	// whether partition was present to remove.
	DeletePartition(pname string) (bool, error)

	// Add member xname id to existing partition.  returns ErrHMSDSNoGroup
	// if partition name does not exist, or ErrHMSDSDuplicateKey if xname id
	// is already in a different partition.
	// Returns key of new member, should be same as id after normalization,
	// if any.  pname should already be normalized.
	AddPartitionMember(pname, id string) (string, error)

	// Delete partition member from partition.  If no error, bool indicates
	// whether member was present to remove.
	DeletePartitionMember(pname, id string) (bool, error)

	//                        Memberships

	// Get the memberships for a particular component xname id
	GetMembership(id string) (*sm.Membership, error)

	// Get all memberships, optionally filtering
	// Convenience feature - not needed for initial implementation
	GetMemberships(f *ComponentFilter) ([]*sm.Membership, error)

	//                                                                    //
	//                    Component Lock Management                       //
	//                                                                    //

	// Create component reservations if one doesn't already exist.
	// To create reservations without a duration, the component must be locked.
	// To create reservations with a duration, the component must be unlocked.
	// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
	// best try.
	InsertCompReservations(f sm.CompLockV2Filter) (sm.CompLockV2ReservationResult, error)

	// Forcebly remove/release component reservations.
	// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
	// best try.
	DeleteCompReservationsForce(f sm.CompLockV2Filter) (sm.CompLockV2UpdateResult, error)

	// Remove/release component reservations.
	// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
	// best try.
	DeleteCompReservations(f sm.CompLockV2ReservationFilter) (sm.CompLockV2UpdateResult, error)

	// Release all expired reservations
	DeleteCompReservationsExpired() ([]string, error)

	// Retrieve the status of reservations. The public key and xname is
	// required to address the reservation.
	GetCompReservations(dkeys []sm.CompLockV2Key) (sm.CompLockV2ReservationResult, error)

	// Update/renew the expiration time of component reservations with the given
	// ID/Key combinations.
	// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
	// best try.
	UpdateCompReservations(f sm.CompLockV2ReservationFilter) (sm.CompLockV2UpdateResult, error)

	// Retrieve component lock information.
	GetCompLocksV2(f sm.CompLockV2Filter) ([]sm.CompLockV2, error)

	// Update component locks. Valid actions are 'Lock', 'Unlock', 'Disable',
	// and 'Repair'.
	// 'Lock'\'Unlock' updates the 'locked' status of the components.
	// 'Disable'\'Repair' updates the 'reservationsDisabled' status of components.
	// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
	// best try.
	UpdateCompLocksV2(f sm.CompLockV2Filter, action string) (sm.CompLockV2UpdateResult, error)

	//                                                                    //
	//                        Job Sync Management                         //
	//                                                                    //

	//                            Job Sync

	// Create a job entry in the job sync. Returns new jobId if successful,
	// otherwise non-nil error.
	InsertJob(j *sm.Job) (string, error)

	// Update the status of the job with the given jobId.
	UpdateJob(jobId, status string) (bool, error)

	// Get the job sync entry with the given job id. Nil if not found and nil
	// error, otherwise non-nil error (not normally expected).
	GetJob(jobId string) (*sm.Job, error)

	// Get list of jobs from the job sync.
	GetJobs(f_opts ...JobSyncFiltFunc) ([]*sm.Job, error)

	// Delete the job entry with the given jobId. If no error, bool indicates
	// whether component lock was present to remove.
	DeleteJob(jobId string) (bool, error)
}

// Table identifiers for generic queries
const (
	ComponentsTable         = "Components"
	RedfishEndpointsTable   = "RedfishEndpoints"
	ComponentEndpointsTable = "ComponentEndpoints"
	ServiceEndpointsTable   = "ServiceEndpoints"
	NodeMapTable            = "NodeMaps"
	HWInvByLocTable         = "HWInventoryByLocation"
	HWInvByFRUTable         = "HWInventoryByFRU"
	DiscoveryStatusTable    = "DiscoveryStatus"
	ScnSubcriptionsTable    = "ScnSubscriptions"
)

type HMSDBTx interface {
	// Terminates transaction, reversing all changes made prior to Begin()
	Rollback() error

	// Terminates transaction successfully, committing all operations
	// performed against it in an atomic fashion.
	Commit() error

	//                                                                    //
	//          Generic single-valed queries - get id/key, etc.           //
	//                                                                    //

	// Get the id values for either all labels in the given table, or a
	// filtered set based on filter f (*ComponentFilter, *RedfishEPFilter,
	// *CompEPFilter)
	// Use one of the above *Table values for 'table' arg,
	// e.g. ComponentsTable
	GetIDListTx(table string, f interface{}) ([]string, error)

	// Build filter query for Component IDs using filter functions and
	// then return the list of matching xname IDs as a string array, write
	// locking the rows if requested (within transaction).
	GetComponentIDsTx(f_opts ...CompFiltFunc) ([]string, error)

	// Build filter query for ComponentEndpoints IDs using filter functions and
	// then return the list of matching xname IDs as a string array, write
	// locking the rows if requested (within transaction).
	GetCompEndpointIDsTx(f_opts ...CompEPFiltFunc) ([]string, error)

	// Build filter query for RedfishEndpoints IDs using filter functions and
	// then return the list of matching xname IDs as a string array, write
	// locking the rows if requested (within transaction).
	GetRFEndpointIDsTx(f_opts ...RedfishEPFiltFunc) ([]string, error)

	//                                                                    //
	//        HMS Components - Managed plane info: State, NID, Role       //
	//                                                                    //

	// Build filter query for State/Components using filter functions and
	// then return the set of matching components as an array, write locking
	// the rows if requested (within transaction).
	//
	// NOTE: Most args allow negated arguments, i.e. "!x0c0s0b0", so be careful
	// passing in user data if the query should only return a single result. ID
	// does not and takes a single ID xname.
	GetComponentsTx(f_opts ...CompFiltFunc) ([]*base.Component, error)

	// Same as above, but also allows only certain fields to be returned
	// via FieldFilter
	GetComponentsFieldFilterTx(
		fieldFltr FieldFilter,
		f_opts ...CompFiltFunc,
	) ([]*base.Component, error)

	// Look up a single HMS Component by id, i.e. xname (in transaction).
	GetComponentByIDTx(id string) (*base.Component, error)

	// Look up a single HMS Component by id, i.e. xname (in transaction).
	// THIS WILL CREATE A WRITE LOCK ON THE ENTRY, so the transaction
	// should not be kept open longer than needed.
	GetComponentByIDForUpdateTx(id string) (*base.Component, error)

	// Get all HMS Components in system (in transaction).
	GetComponentsAllTx() ([]*base.Component, error)

	// Get some or all HMS Components in system (in transaction), with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetComponentsFilterTx(f *ComponentFilter, fieldFltr FieldFilter) ([]*base.Component, error)

	// Get some or all HMS Components in system (in transaction) under
	// a set of parent components, with filtering options to possibly
	// narrow the returned values. If no filter provided, just get
	// the parent components.  Otherwise use it to create a custom
	// WHERE... string that filters out entries that do not match ALL
	// of the non-empty strings in the filter struct.
	GetComponentsQueryTx(f *ComponentFilter, fieldFltr FieldFilter, ids []string) ([]*base.Component, error)

	// Get a single HMS Component by its NID, if the NID exists (in transaction)
	GetComponentByNIDTx(nid string) (*base.Component, error)

	// Retrieve all HMS Components of the given type (in transaction).
	//GetAllCompByTypeTx(t string) ([]*base.Component, error)

	// Insert HMS Component into database, updating it if it exists.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	InsertComponentTx(c *base.Component) (int64, error)

	// Insert HMS Components into the database, updating it if it exists.
	// Returns the IDs of the affected components.
	InsertComponentsTx(comps []*base.Component) ([]string, error)

	// Update state and flag fields only in DB for xname IDs 'ids'
	// If force = true ignores any starting state restrictions and will always
	// set ids to state, unless that state is already set.
	//
	// If noVerify = true, don't add extra where clauses to ensure only the
	// rows that should change do.  If true, either we already verified that
	// the ids list will be changed, and have locked the rows, or else we don't
	// care to know the exact ids that actually changed.
	//
	// Returns the number of affected rows. < 0 means RowsAffected() is not
	// supported.
	//   Note: If flag is not set, it will be set to OK (i.e. no flag)
	UpdateCompStatesTx(ids []string, state, flag string, force, noVerify bool, pi *PartInfo) (int64, error)

	// Update Flag field in DB from c's Flag field.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: Flag cannot be empty/invalid.
	UpdateCompFlagOnlyTx(id string, flag string) (int64, error)

	// Update flag field in DB for a list of components.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: Flag cannot be empty/invalid.
	BulkUpdateCompFlagOnlyTx(ids []string, flag string) (int64, error)

	// Update enabled field in DB from c's Enabled field (in transaction).
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: c.Enabled cannot be nil.
	UpdateCompEnabledTx(id string, enabled bool) (int64, error)

	// Update Enabled field only in DB for a list of components (in transaction)
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	BulkUpdateCompEnabledTx(ids []string, enabled bool) (int64, error)

	// Update SwStatus field in DB from c's SwStatus field (in transaction).
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	UpdateCompSwStatusTx(id string, swStatus string) (int64, error)

	// Update SwStatus field only in DB for a list of components
	// (In transaction.)
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	BulkUpdateCompSwStatusTx(ids []string, swstatus string) (int64, error)

	// Update Role/SubRole field in DB from c's Role/SubRole field.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: Role cannot be blank/invalid.
	UpdateCompRoleTx(id string, role, subRole string) (int64, error)

	// Update Role/SubRole field in DB for a list of components.
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	// Note: Role cannot be empty/invalid.
	BulkUpdateCompRoleTx(ids []string, role, subRole string) (int64, error)

	// Update Class field only in DB for a list of components
	// (In transaction.)
	// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
	BulkUpdateCompClassTx(ids []string, class string) (int64, error)

	// Update NID.  If NID is not set or negative, it is set to -1 which
	// effectively unsets it and suppresses its output.
	UpdateCompNIDTx(c *base.Component) error

	// Update NID.  If NID is not set or negative, it is set to -1 which
	// effectively unsets it and suppresses its output.
	BulkUpdateCompNIDTx(comps []base.Component) error

	// Delete HMS Component with matching xname id from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteComponentByIDTx(id string) (bool, error)

	// Delete all HMS Components from database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteComponentsAllTx() (int64, error)

	//                                                                    //
	//              Node to Default NID, role, etc. mapping               //
	//                                                                    //

	// Look up one Node->NID Mapping by id, i.e. node xname (in transaction).
	GetNodeMapByIDTx(id string) (*sm.NodeMap, error)

	// Look up ALL Node->NID Mappings (in transaction).
	GetNodeMapsAllTx() ([]*sm.NodeMap, error)

	// Insert Node->NID Mapping into database, updating it if it exists.
	InsertNodeMapTx(m *sm.NodeMap) error

	// Delete Node NID Mapping entry with matching xname id from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteNodeMapByIDTx(id string) (bool, error)

	// Delete all Node NID Mapping entries from database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteNodeMapsAllTx() (int64, error)

	//                                                                    //
	//                 Component to power supply mapping                  //
	//                                                                    //

	// Look up one Power Mapping by id, i.e. node xname (in transaction).
	GetPowerMapByIDTx(id string) (*sm.PowerMap, error)

	// Look up ALL Power Mappings (in transaction).
	GetPowerMapsAllTx() ([]*sm.PowerMap, error)

	// Insert Power Mapping into database, updating it if it exists.
	InsertPowerMapTx(m *sm.PowerMap) error

	// Delete Power Mapping entry with matching xname id from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeletePowerMapByIDTx(id string) (bool, error)

	// Delete all Power Mapping entries from database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeletePowerMapsAllTx() (int64, error)

	//                                                                    //
	//        Hardware Inventory - Detailed location and FRU info         //
	//                                                                    //

	// Get a single Hardware inventory entry by current xname (in transaction).
	// This struct includes the FRU info if the xname is currently populated.
	GetHWInvByLocIDTx(id string) (*sm.HWInvByLoc, error)

	// Get HWInvByLoc by primary key (xname) for all entries in the system.
	// It also pairs the data with the matching HWInvByFRU if the xname is
	// populated. (In transaction)
	GetHWInvByLocAllTx() ([]*sm.HWInvByLoc, error)

	// Get a single HW-inventory-by-FRU entry by its FRUID. (in transaction).
	GetHWInvByFRUIDTx(fruid string) (*sm.HWInvByFRU, error)

	// Get all HW-inventory-by-FRU entries. (in transaction).
	GetHWInvByFRUAllTx() ([]*sm.HWInvByFRU, error)

	// Insert or update HWInventoryByLocation struct (in transaction)
	// If PopulatedFRU is present, only the FRUID is added to the database.  If
	// it is not, this effectively "depopulates" the given location.
	// The actual HWInventoryByFRU struct must be stored FIRST using the
	// corresponding function (presumably within the same transaction), as
	// the location info will need to link to it.
	InsertHWInvByLocTx(hl *sm.HWInvByLoc) error

	// Insert or update HWInventoryByLocation structs (in transaction)
	// If PopulatedFRU is present, only the FRUID is added to the database. If
	// it is not, this effectively "depopulates" the given location.
	// The actual HWInventoryByFRU struct must be stored FIRST using the
	// corresponding function (presumably within the same transaction), as
	// the location info will need to link to it.
	BulkInsertHWInvByLocTx(hl []*sm.HWInvByLoc) error

	// Insert or update HWInventoryByFRU struct (in transaction)
	InsertHWInvByFRUTx(hf *sm.HWInvByFRU) error

	// Insert or update HWInventoryByFRU struct (in transaction)
	BulkInsertHWInvByFRUTx(hf []*sm.HWInvByFRU) error

	// Delete HWInvByLoc entry with matching ID from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteHWInvByLocIDTx(id string) (bool, error)

	// Delete all HWInvByLoc entriesfrom database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteHWInvByLocsAllTx() (int64, error)

	// Delete HWInvByFRU entry with matching FRU ID from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteHWInvByFRUIDTx(fruid string) (bool, error)

	// Delete all HWInvByFRU entries from database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteHWInvByFRUsAllTx() (int64, error)

	// Get some or all Hardware Inventory entries with filtering
	// options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct.
	GetHWInvByLocQueryFilterTx(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error)

	//                                                                    //
	//   Hardware Inventory History - Detailed history of hardware FRU    //
	//                                location.                           //
	//                                                                    //

	// Get hardware history for some or all Hardware Inventory entries with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct.
	// (in transaction)
	GetHWInvHistFilterTx(f_opts ...HWInvHistFiltFunc) ([]*sm.HWInvHist, error)

	// Get only the most recent hardware history event for some or all hardware
	// locations. (in transaction)
	GetHWInvHistLastEventsTx(ids []string) ([]*sm.HWInvHist, error)

	// Insert a HWInventoryHistory struct (in transaction)
	InsertHWInvHistTx(hh *sm.HWInvHist) error

	// Insert an array of HWInventoryHistory entries. (in transaction)
	// If a duplicate is present return an error.
	InsertHWInvHistsTx(hhs []*sm.HWInvHist) error

	//                                                                    //
	//    Redfish Endpoints - Redfish service roots used for discovery    //
	//                                                                    //

	// Build filter query for RedfishEndpoints using filter functions and
	// then return the set of matching components as an array.
	//
	// NOTE: Most args allow negated arguments, i.e. "!x0c0s0b0", so be careful
	// passing in user data if the query should only return a single result.
	// RFE_ID() does not and takes a single arg.
	GetRFEndpointsTx(f_opts ...RedfishEPFiltFunc) ([]*sm.RedfishEndpoint, error)

	// Get RedfishEndpoint by ID (xname), i.e. a single entry (in transaction).
	GetRFEndpointByIDTx(id string) (*sm.RedfishEndpoint, error)

	// Get all RedfishEndpoints in system (in transaction)
	GetRFEndpointsAllTx() ([]*sm.RedfishEndpoint, error)

	// Get some or all RedfishEndpoints in system (in transaction), with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetRFEndpointsFilterTx(f *RedfishEPFilter) ([]*sm.RedfishEndpoint, error)

	// Insert new RedfishEndpoint into database (in transaction)
	// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
	// No insertion done on err != nil
	InsertRFEndpointTx(ep *sm.RedfishEndpoint) error

	// Insert new RedfishEndpoints into database (in transaction)
	// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
	// No insertion done on err != nil
	InsertRFEndpointsTx(eps []*sm.RedfishEndpoint) error

	// Update RedfishEndpoint already in DB. Does not update any
	// ComponentEndpoint children. (In transaction.)
	// If err == nil, but FALSE is returned, then no changes were made.
	UpdateRFEndpointTx(ep *sm.RedfishEndpoint) (bool, error)

	// Update RedfishEndpoint already in DB. Does not update any
	// ComponentEndpoint children. (In transaction.)
	// If err == nil, but FALSE is returned, then no changes were made.
	UpdateRFEndpointsTx(ep []*sm.RedfishEndpoint) ([]*sm.RedfishEndpoint, error)

	// Update RedfishEndpoint already in DB, leaving DiscoveryInfo
	// unmodifed.  Does not update any ComponentEndpoint children.
	// If err == nil, but FALSE is returned, then no changes were made.
	// (In transaction.)
	UpdateRFEndpointNoDiscInfoTx(ep *sm.RedfishEndpoint) (bool, error)

	// Delete RedfishEndpoint with matching xname id from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteRFEndpointByIDTx(id string) (bool, error)

	// Delete all RedfishEndpoints from database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteRFEndpointsAllTx() (int64, error)

	// Given the id of a RedfishEndpoint, set the states of all children
	// with State/Components entries to state and flag, returning a list of
	// xname IDs were at least state or flag was updated.
	//
	// CREATES A WRITE LOCK ON Redfish/ComponentEndpoints and Components tables
	// (all three) until transaction is committed if wrLock == true
	//
	// Detaches FRUs from locations if detachFRUs == true
	SetChildCompStatesRFEndpointsTx(ids []string, state, flag string, wrLock bool, detachFRUs bool) ([]string, error)

	//                                                                    //
	// ComponentEndpoints: Component info discovered from Parent          //
	//                     RedfishEndpoint.  Management plane equivalent  //
	//                     to HMS Component.                              //
	//                                                                    //

	// Build filter query for ComponentEndpoints using filter functions and
	// then return the set of matching components as an array.
	//
	// NOTE: Most args allow negated arguments, i.e. "!x0c0s0b0", so be careful
	// passing in user data if the query should only return a single result.
	// CE_ID does not and takes a single xname ID.
	GetCompEndpointsTx(f_opts ...CompEPFiltFunc) ([]*sm.ComponentEndpoint, error)

	// Get ComponentEndpoint by id (xname), i.e. a single entry (in transaction).
	GetCompEndpointByIDTx(id string) (*sm.ComponentEndpoint, error)

	// Get all ComponentEndpoints in system (in transaction)
	GetCompEndpointsAllTx() ([]*sm.ComponentEndpoint, error)

	// Get some or all ComponentEndpoints in system (in transaction), with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetCompEndpointsFilterTx(f *CompEPFilter) ([]*sm.ComponentEndpoint, error)

	// Upsert ComponentEndpoint into database, updating it if it exists
	// (in transaction)
	UpsertCompEndpointTx(cep *sm.ComponentEndpoint) error

	// Upsert ComponentEndpoints into database, updating them if they exist
	// (in transaction)
	UpsertCompEndpointsTx(ceps *sm.ComponentEndpointArray) error

	// Delete ComponentEndpoint with matching xname id from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteCompEndpointByIDTx(id string) (bool, error)

	// Delete all ComponentEndpoints from database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteCompEndpointsAllTx() (int64, error)

	// Given the id of a ComponentEndpoint, set the states of matching
	// State/Components entries to state and flag, returning a list of
	// xname IDs were at least state or flag was updated.
	//
	// CREATES A WRITE LOCK ON Redfish/ComponentEndpoints and Components tables
	// (all three) until transaction is committed if wrLock == true
	SetChildCompStatesCompEndpointsTx(ids []string, state, flag string, wrLock bool) ([]string, error)

	//                                                                    //
	// ServiceEndpoints: Redfish service info discovered from the parent  //
	//                   RedfishEndpoint.                                 //
	//                                                                    //

	// Build filter query for ServiceEndpoints using filter functions and
	// then return the set of matching endpoints as an array.
	GetServiceEndpointsTx(f_opts ...ServiceEPFiltFunc) ([]*sm.ServiceEndpoint, error)

	// Get ServiceEndpoint by id (xname), i.e. a single entry (in transaction).
	GetServiceEndpointByIDTx(svc, id string) (*sm.ServiceEndpoint, error)

	// Get all ServiceEndpoints in system (in transaction)
	GetServiceEndpointsAllTx() ([]*sm.ServiceEndpoint, error)

	// Get some or all ServiceEndpoints in system (in transaction), with
	// filtering options to possibly narrow the returned values.
	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string that filters out entries that
	// do not match ALL of the non-empty strings in the filter struct
	GetServiceEndpointsFilterTx(f *ServiceEPFilter) ([]*sm.ServiceEndpoint, error)

	// Upsert ServiceEndpoint into database, updating it if it exists
	// (in transaction)
	UpsertServiceEndpointTx(sep *sm.ServiceEndpoint) error

	// Upsert ServiceEndpoints into database, updating them if they exist
	// (in transaction)
	UpsertServiceEndpointsTx(seps *sm.ServiceEndpointArray) error

	// Delete ServiceEndpoint with matching xname id from database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteServiceEndpointByIDTx(svc, id string) (bool, error)

	// Delete all ServiceEndpoints from database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteServiceEndpointsAllTx() (int64, error)

	//                                                                    //
	//    Component Ethernet Interfaces - MAC address to IP address       //
	//        relations for component endpoint ethernet interfaces        //
	//                                                                    //

	// Get CompEthInterface by ID, i.e. a single entry for UPDATE (in transaction).
	GetCompEthInterfaceByIDTx(id string) (*sm.CompEthInterfaceV2, error)

	// Insert a new CompEthInterface into database (in transaction)
	// If ID or MAC already exists, return ErrHMSDSDuplicateKey
	// No insertion done on err != nil
	InsertCompEthInterfaceTx(cei *sm.CompEthInterfaceV2) error

	// Insert a new CompEthInterface into database (in transaction)
	// If ID or MAC already exists, return ErrHMSDSDuplicateKey
	// No insertion done on err != nil
	InsertCompEthInterfacesTx(ceis []*sm.CompEthInterfaceV2) error

	// Insert/update a new CompEthInterface into the database (in transaction)
	// If ID or FQDN already exists, only overwrite ComponentID
	// and Type fields.
	// No insertion done on err != nil
	InsertCompEthInterfaceCompInfoTx(cei *sm.CompEthInterfaceV2) error

	// Insert/update new CompEthInterfaces into the database (in transaction)
	// If ID or FQDN already exists, only overwrite ComponentID
	// and Type fields.
	// No insertion done on err != nil
	InsertCompEthInterfacesCompInfoTx(ceis []*sm.CompEthInterfaceV2) error

	// Update CompEthInterface already in the DB. (In transaction.)
	// If err == nil, but FALSE is returned, then no changes were made.
	UpdateCompEthInterfaceTx(cei *sm.CompEthInterfaceV2, ceip *sm.CompEthInterfaceV2Patch) (bool, error)

	// Delete a CompEthInterface with matching id from the database, if it
	// exists (in transaction)
	// Return true if there was a row affected, false if there were zero.
	DeleteCompEthInterfaceByIDTx(id string) (bool, error)

	// Delete all CompEthInterfaces from the database (in transaction).
	// Also returns number of deleted rows, if error is nil.
	DeleteCompEthInterfacesAllTx() (int64, error)

	//                                                                    //
	//           DiscoveryStatus: Discovery Status tracking               //
	//                                                                    //

	// Get DiscoveryStatus with the given numerical ID(in transaction).
	GetDiscoveryStatusByIDTx(id uint) (*sm.DiscoveryStatus, error)

	// Get all DiscoveryStatus entries (in transaction).
	GetDiscoveryStatusAllTx() ([]*sm.DiscoveryStatus, error)

	// Update discovery status in DB (in transaction)
	UpsertDiscoveryStatusTx(stat *sm.DiscoveryStatus) error

	//                                                                    //
	//           SCNSubscription: SCN subscription management             //
	//                                                                    //

	// Get all SCN subscriptions
	GetSCNSubscriptionsAllTx() (*sm.SCNSubscriptionArray, error)

	// Get a SCN subscription
	GetSCNSubscriptionTx(id int64) (*sm.SCNSubscription, error)

	// Insert a new SCN subscription. Existing subscriptions are unaffected
	InsertSCNSubscriptionTx(sm.SCNPostSubscription) (int64, error)

	// Update an existing SCN subscription.
	UpdateSCNSubscriptionTx(id int64, sub sm.SCNPostSubscription) (bool, error)

	// Patch a SCN subscription
	PatchSCNSubscriptionTx(id int64, op string, patch sm.SCNPatchSubscription) (bool, error)

	// Delete a SCN subscription
	DeleteSCNSubscriptionTx(id int64) (bool, error)

	// Delete all SCN subscriptions
	DeleteSCNSubscriptionsAllTx() (int64, error)

	//                                                                    //
	//                 Group and Partition  Management                    //
	//                                                                    //

	//                           Groups

	// Creates new group in component groups, but adds nothing to the members
	// table (in tx, so this can be done in separate query)
	//
	// Returns: (new UUID string, new group's label, excl group name, error)
	InsertEmptyGroupTx(g *sm.Group) (string, string, string, error)

	// Update fields in GroupPatch on the returned Group object provided
	// (in transaction).
	UpdateEmptyGroupTx(uuid string, g *sm.Group, gp *sm.GroupPatch) error

	// Get the user-readable fields in a group entry and it's internal uuid but
	// don't fetch its members (done in transaction, so we can fetch them as part
	// of the same one).
	GetEmptyGroupTx(label string) (uuid string, g *sm.Group, err error)

	//                         Partitions

	// Creates new partition  in component groups, but adds nothing to the members
	// table (in tx, so this can be done in separate query)
	//
	// Returns: (new UUID string, new partition's official name, error)
	InsertEmptyPartitionTx(p *sm.Partition) (string, string, error)

	// Update fields in PartitionPatch on the returned partition object provided
	// (in transaction).
	UpdateEmptyPartitionTx(uuid string, p *sm.Partition, pp *sm.PartitionPatch) error

	// Get the user-readable fields in a partition entry and it's internal uuid but
	// don't fetch its members (done in transaction, so we can fetch them as part
	// of the same one).
	GetEmptyPartitionTx(name string) (uuid string, p *sm.Partition, err error)

	//                  Members (for either Group/Partition)

	// Insert memberlist for group/part.  The uuid parameter should be
	// as-returned by  InsertEmptyGroupTx()/InsertEmptyPartitionTx().
	//
	// Namespace should be either the group name or a non-conflicting (using
	// normally disallowed characters) string based on the exclusiveGroup if
	// there is one, or a single namespace string for all partition (again
	// non-conflicting, also with other exclusive groups, by using different)
	// characters).
	//
	// Namespace is not used for any user output, just to allow uniqueness to be
	// enforced by the DB.
	InsertMembersTx(uuid, namespace string, ms *sm.Members) error

	// UUID string should be as retried from one of the group/partition calls.  No
	// guarantees made about alternate formatting of the underlying binary value.
	GetMembersTx(uuid string) (*sm.Members, error)

	// UUID string should be as retried from one of the group/partition calls.  No
	// guarantees made about alternate formatting of the underlying binary value.
	// Result is uuid members, but with entries not also in and_uuid removed.
	GetMembersFilterTx(uuid, and_uuid string) (*sm.Members, error)

	// Given an internal group_id uuid, delete the given id, if it exists.
	// if it does not, result will be false, nil vs. true,nil on deletion.
	DeleteMemberTx(uuid, id string) (bool, error)

	//                                                                    //
	//                    Component Lock Management                       //
	//                                                                    //

	// Insert component reservations into the database.
	// To Insert reservations without a duration, the component must be locked.
	// To Insert reservations with a duration, the component must be unlocked.
	InsertCompReservationsTx(ids []string, duration int) ([]sm.CompLockV2Success, string, error)

	// Remove/release component reservations.
	// Both a component ID and reservation key are required for these operations unless force = true.
	DeleteCompReservationsTx(rKeys []sm.CompLockV2Key, force bool) ([]string, error)

	// Release all expired component reservations
	DeleteCompReservationExpiredTx() ([]string, error)

	// Retrieve the status of reservations. The public key and xname is
	// required to address the reservation unless force = true.
	GetCompReservationsTx(dKeys []sm.CompLockV2Key, force bool) ([]sm.CompLockV2Success, string, error)

	// Update/renew the expiration time of component reservations with the given
	// ID/Key combinations.
	UpdateCompReservationsTx(rKeys []sm.CompLockV2Key, duration int, force bool) ([]string, error)

	// Update component 'ReservationsDisabled' field.
	BulkUpdateCompResDisabledTx(ids []string, disabled bool) ([]string, error)

	// Update component 'locked' field.
	BulkUpdateCompResLockedTx(ids []string, locked bool) ([]string, error)

	//                                                                    //
	//                        Job Sync Management                         //
	//                                                                    //

	//                             Jobs

	// Creates new job entry in the job sync, but adds nothing to the job's
	// type table (in tx, so this can be done in separate query)
	//
	// Returns: (new jobId string, error)
	InsertEmptyJobTx(j *sm.Job) (string, error)

	// Update the status and LastUpdated fields for a Job entry (in transaction).
	UpdateEmptyJobTx(jobId string, status string) (bool, error)

	// Get the user-readable fields in a job entry but don't fetch its job type
	// specific data (done in transaction, so we can fetch them as part of the
	// same one).
	GetEmptyJobTx(jobId string) (j *sm.Job, err error)

	// Get the user-readable fields in a job entry but don't fetch its job type
	// specific data (done in transaction, so we can fetch them as part of the
	// same one).
	GetEmptyJobsTx(f_opts ...JobSyncFiltFunc) (js []*sm.Job, err error)

	//                    State Redfish Poll Jobs

	// Insert job specific info for the given jobId. The jobId parameter should
	// be as-returned by InsertEmptyJobTx()/InsertEmptyJobTx().
	InsertStateRFPollJobTx(jobId string, data *sm.SrfpJobData) error

	// Get the job specific info associated with the jobId. The jobId string should
	// be as retried from one of the Job calls.  No guarantees made about
	// alternate formatting of the underlying binary value.
	// GetStateRFPollJobTx(jobId string) (*sm.SrfpJobData, error)

	// Get the job specific info associated with the jobId. The jobId string should
	// be as retried from one of the Job calls.  No guarantees made about
	// alternate formatting of the underlying binary value.
	GetStateRFPollJobByIdTx(jobId string) (*sm.SrfpJobData, error)
}
