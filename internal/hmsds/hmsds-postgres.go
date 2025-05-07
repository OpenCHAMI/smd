// MIT License
//
// (C) Copyright [2018-2024] Hewlett Packard Enterprise Development LP
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
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	rf "github.com/OpenCHAMI/smd/v2/pkg/redfish"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

// MUST be kept in sync with schema installed via smd-init job
const HMSDS_PG_SCHEMA = 20
const HMSDS_PG_SYSTEM_ID = 0

type hmsdbPg struct {
	dsn       string // dataSourceName
	connected bool   // Has db.Open been called?
	db        *sql.DB
	ctx       context.Context
	sc        *sq.StmtCache
	lg        *log.Logger
	lgLvl     LogLevel
}

// Gen DSN for MySQL/MariaDB
func GenDsnHMSDB_PB(name, user, pass, host, opts string, port int) string {
	dsn := ""
	sep := ""
	if name != "" {
		dsn = fmt.Sprintf("dbname=%s", name)
		sep = " "
	}
	if user != "" {
		dsn = fmt.Sprintf("%s%suser=%s", dsn, sep, user)
		sep = " "
	}
	if pass != "" {
		dsn = fmt.Sprintf("%s%spassword=%s", dsn, sep, pass)
		sep = " "
	}
	if host != "" {
		dsn = fmt.Sprintf("%s%shost=%s", dsn, sep, host)
		sep = " "
	}
	if port != 0 {
		dsn = fmt.Sprintf("%s%sport=%d", dsn, sep, port)
		sep = " "
	}
	if opts != "" {
		dsn = fmt.Sprintf("%s%s%s", dsn, sep, opts)
		sep = " "
	}
	return dsn
}

// Variant for Postgres databases.
func NewHMSDB_PG(dsn string, l *log.Logger) HMSDB {
	d := new(hmsdbPg)
	d.dsn = dsn
	d.db = nil
	d.connected = false
	d.lgLvl = LOG_DEFAULT
	d.ctx = context.TODO()

	if l == nil {
		d.lg = log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags|log.Lmicroseconds)
	} else {
		d.lg = l
	}
	return d
}

// Conditional logging function (based on current log level set for conn)
func (d *hmsdbPg) Log(l LogLevel, format string, a ...interface{}) {
	if int(l) <= int(d.lgLvl) {
		// depth=2, get line num of caller, not us
		d.lg.Output(2, fmt.Sprintf(format, a...))
	}
}

// Log to logging infrastructure regardless of current log level.
func (d *hmsdbPg) LogAlways(format string, a ...interface{}) {
	// Use caller's line number (depth=2)
	d.lg.Output(2, fmt.Sprintf(format, a...))
}

// Works like log.Printf, but registers error for function calling the
// function that is printing the error. e.g. instead of always saying
// an error occurred in begin(), we show where begin() was called, so
// we don't have to guess, and the message can make clear what in begin()
// failed.
func (d *hmsdbPg) LogAlwaysParentFunc(format string, a ...interface{}) {
	// Use caller's caller's line number (depth=3)
	d.lg.Output(3, fmt.Sprintf(format, a...))
}

func (d *hmsdbPg) ImplementationName() string {
	return "Postgres"
}

func (d *hmsdbPg) SetLogLevel(lvl LogLevel) error {
	if lvl >= LOG_DEFAULT && lvl < LOG_LVL_MAX {
		d.lgLvl = lvl
		return nil
	} else {
		return errors.New("Warning: verbose level unchanged")
	}
}

// Is error from postgres?
func IsDBErrorPg(err error) bool {
	_, ok := err.(*pq.Error)
	if !ok {
		return false
	} else {
		return true
	}
}

// Is error from postgres and indicating a duplicate key error?
func IsPgDuplicateKeyErr(err error) bool {
	if pgError, ok := err.(*pq.Error); ok {
		if pgError.Code == "23505" {
			// Key already exists, user error.
			return true
		}
	}
	return false
}

// Is error from postgres and indicating a foreign key error?
func IsPgForeignKeyErr(err error) bool {
	if pgError, ok := err.(*pq.Error); ok {
		if pgError.Code == "23503" {
			// Key already exists, user error.
			return true
		}
	}
	return false
}

// Takes an error from the database driver, and if it matches certain
// expected scenarios, we will return an HMSError that gives the user-
// friendly message suitable for returning to the client.
// If not, we will return the original error (which should fail
// base.IsHMSError, and so can be treated as a potentially sensitive
// internal error where we just log it an send something generic to the
// client.
func ParsePgDBError(err error) error {
	if IsPgDuplicateKeyErr(err) {
		if pgErr, ok := err.(*pq.Error); ok {
			// Look at which key - component_id and group_namespace or group_id
			if strings.Contains(pgErr.Detail, compGroupMembersNsCol) {
				if strings.Contains(pgErr.Detail, partGroupNamespace) {
					return ErrHMSDSExclusivePartition
				} else {
					return ErrHMSDSExclusiveGroup
				}
			}
		}
		return ErrHMSDSDuplicateKey
	}
	if IsPgForeignKeyErr(err) {
		return ErrHMSDSNoComponent
	}
	return err
}

////////////////////////////////////////////////////////////////////////////
//
// DB operations - Open, Close, Start Transaction
//
////////////////////////////////////////////////////////////////////////////

func (d *hmsdbPg) Open() error {
	var err error
	if d.connected == true {
		d.LogAlways("Warning: Open(): Already called, but no Close()")
		return nil
	}
	// Create long-lived database handle.  This handle can manage many
	// concurrent DB connections up to the configured limit and is
	// safe for use by multiple Go routes.
	d.db, err = sql.Open("postgres", d.dsn)
	if err != nil {
		d.LogAlways("Error: Open(): sql.Open failed: %s", err)
		return err
	}
	// Ping the new DB handle.  sql.Open does not actually connect to the DB;
	// this is done as-needed, so make sure we can actually do so.
	err = d.db.Ping()
	if err != nil {
		d.LogAlways("Error: Open(): Failed to ping DB: %s", err)
		d.db.Close()
		return err
	}
	// If we can read the DB, we should be able to get the schema version.
	// Make sure the expected version is installed and it's not still updating.
	err = d.checkPgSchemaVersion(HMSDS_PG_SYSTEM_ID, HMSDS_PG_SCHEMA)
	if err != nil {
		d.LogAlways("Error: Open(): Schema check failed: %s", err)
		d.db.Close()
		return err
	}
	//
	// Configure connection here
	//

	// This needs to be less than MariaDB's current global max_connections
	// via a cnf file or the default of 151.  The sql package default is
	// unlimited and this causes the server-side limit to get overwhelmed.
	// This should be kept in sync with the configured value in the
	// mariadb docker container, which presently has it set to 100 (likely
	// way too low).  Note that it is a global value, however, and we should
	// leave a little slack in any case (setting it to 100 exacly causes
	// an occasional failure, even with no other processes connecting).
	d.db.SetMaxOpenConns(70)

	// Workaround for HMS-1080, one of these, so long as a minute is less
	// than wait_timeout.
	//d.db.SetMaxIdleConns(0)
	//d.db.SetConnMaxLifetime(time.Minute)

	// Mark handle as connected, as we've successfully contacted the DB
	// and are ready to perform queries.
	d.connected = true

	// Create statement cache now that we're open.
	d.sc = sq.NewStmtCache(d.db)

	d.LogAlways("Open() completed successfully.")
	return nil
}

// Check the systemId (should only be 0 currently) schema_version and
// if it does not match, return ErrHMSDSBadSchema.  If no other
// error, will return nil if schema is obtained and matches expected
// version.
func (d *hmsdbPg) checkPgSchemaVersion(sysId, expectedVersion int) error {
	var schemaVersion int

	query := selectSystemSchemaVersion(sysId)
	query = query.PlaceholderFormat(sq.Dollar)
	err := query.RunWith(d.db).QueryRow().Scan(&schemaVersion)
	if err != nil {
		d.LogAlways("Error: System table query failed: %s", err)
		return err
	}
	// New schema are backwards compatible
	if schemaVersion < expectedVersion {
		d.LogAlways("Got schema version %d, expected %d+, for id=%d",
			schemaVersion, expectedVersion, sysId)
		return ErrHMSDSBadSchema
	}
	d.LogAlways("Running schema version %d.", expectedVersion)
	return nil
}

// Closes the database connection.  This is a global operation that
// affects all go routines using a hmsdb handle.  It is only used when
// we are done with the DB entirely.  Individual connections are pooled
// and managed transparently by the sql API, so fine-grained management
// is not needed for individual DB calls.
func (d *hmsdbPg) Close() error {
	if d.connected != true {
		d.LogAlways("Warning: Close(): Not open.")
		return nil
	}
	d.connected = false

	err := d.db.Close()
	return err
}

// Starts a new transaction, returning a HMSDBTx handle which allows
// transaction-friendly operations to be invoked in sequence.  You
// MUST close the HMSDBTx handle with one of Commit or Rollback when
// done or the operations will eventually time out and rollback anyways,
// but in such a way that they may block operations on the same DB
// resources until then.
func (d *hmsdbPg) Begin() (HMSDBTx, error) {
	if d.connected == false {
		return nil, ErrHMSDSPtrClosed
	}
	// We back off and retry if we can't create a new transaction.
	// We should keep things tuned so we never run out of connections,
	// but we don't want things just randomly failing unless things are
	// really bad on the server side,
	var err error = nil
	var tx HMSDBTx = nil
	for i := 0; i < 8; i++ {
		tx, err = newHMSDBPgTx(d)
		if err == nil {
			return tx, err
		}
		if i == 0 {
			d.Log(LOG_INFO, "BeginTx failed: DBStats: %+v", d.db.Stats())
		}
		time.Sleep(time.Millisecond * time.Duration(10+(50*i)))
	}
	// Should always have an error here.
	if err == nil {
		err = ErrHMSDSTxFailed
	}
	d.LogAlwaysParentFunc("BeginTx failed even after retries: %s", err)
	return nil, err
}

// Test the database connection to make sure that it is healthy
func (d *hmsdbPg) TestConnection() error {
	if !d.connected {
		d.LogAlways("Warning: TestConnection(): Not open")
		return ErrHMSDSPtrClosed
	}
	// Ping the DB handle. This is done as-needed, so make sure we can
	// actually do so.
	err := d.db.Ping()
	if err != nil {
		d.LogAlways("Error: TestConnection(): Failed to ping DB: %s", err)
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////
//
// HMSDB Interface - Generic ID queries
//
//    Return IDs of given type based on Arguments
//
////////////////////////////////////////////////////////////////////////////

// Build filter query for Component IDs using filter functions and
// then return the list of matching xname IDs as a string array, write
// locking the rows if requested.
func (d *hmsdbPg) GetComponentIDs(f_opts ...CompFiltFunc) ([]string, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	ids, err := t.GetComponentIDsTx(f_opts...)
	if err != nil {
		t.Rollback()
		return ids, err
	}
	err = t.Commit()
	return ids, err
}

// Build filter query for ComponentEndpoints IDs using filter functions and
// then return the list of matching xname IDs as a string array, write
// locking the rows if requested.
func (d *hmsdbPg) GetCompEndpointIDs(f_opts ...CompEPFiltFunc) ([]string, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	ids, err := t.GetCompEndpointIDsTx(f_opts...)
	if err != nil {
		t.Rollback()
		return ids, err
	}
	err = t.Commit()
	return ids, err
}

// Build filter query for RedfishEndpoints IDs using filter functions and
// then return the list of matching xname IDs as a string array, write
// locking the rows if requested.
func (d *hmsdbPg) GetRFEndpointIDs(f_opts ...RedfishEPFiltFunc) ([]string, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	ids, err := t.GetRFEndpointIDsTx(f_opts...)
	if err != nil {
		t.Rollback()
		return ids, err
	}
	err = t.Commit()
	return ids, err
}

////////////////////////////////////////////////////////////////////////////
//
// HMS Components - Managed plane info: State, NID, Role
//
////////////////////////////////////////////////////////////////////////////

// Get a single component entry by its ID/xname
func (d *hmsdbPg) GetComponentByID(id string) (*base.Component, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	comp, err := t.GetComponentByIDTx(id)
	if err != nil {
		t.Rollback()
		return comp, err
	}
	err = t.Commit()
	return comp, err
}

// Get all HMS Components in system.
func (d *hmsdbPg) GetComponentsAll() ([]*base.Component, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	comps, err := t.GetComponentsAllTx()
	if err != nil {
		t.Rollback()
		return comps, err
	}
	err = t.Commit()
	return comps, err
}

// Get some or all HMS Components in system, with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbPg) GetComponentsFilter(f *ComponentFilter, fieldFltr FieldFilter) ([]*base.Component, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	comps, err := t.GetComponentsFilterTx(f, fieldFltr)
	if err != nil {
		t.Rollback()
		return comps, err
	}
	err = t.Commit()
	return comps, err
}

// Get some or all HMS Components in system under
// a set of parent components, with filtering options to possibly
// narrow the returned values. If no filter provided, just get
// the parent components.  Otherwise use it to create a custom
// WHERE... string that filters out entries that do not match ALL
// of the non-empty strings in the filter struct.
func (d *hmsdbPg) GetComponentsQuery(f *ComponentFilter, fieldFltr FieldFilter, ids []string) ([]*base.Component, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	comps, err := t.GetComponentsQueryTx(f, fieldFltr, ids)
	if err != nil {
		t.Rollback()
		return comps, err
	}
	err = t.Commit()
	return comps, err
}

// Get a single component by its NID, if one exists.
func (d *hmsdbPg) GetComponentByNID(nid string) (*base.Component, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	comp, err := t.GetComponentByNIDTx(nid)
	if err != nil {
		t.Rollback()
		return comp, err
	}
	t.Commit()
	return comp, err
}

// Insert HMS Component into database, updating it if it exists.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (d *hmsdbPg) InsertComponent(c *base.Component) (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	rowsAffected, err := t.InsertComponentTx(c)
	if err != nil {
		t.Rollback()
		return 0, err
	}
	if err = t.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Inserts or updates ComponentArray entries in database within a
// single all-or-none transaction.
func (d *hmsdbPg) InsertComponents(comps *base.ComponentArray) ([]string, error) {
	if comps == nil {
		d.LogAlways("Error: InsertComponents(): Component array was nil.")
		return []string{}, ErrHMSDSArgNil
	}
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}
	affectedIDs, err := t.InsertComponentsTx(comps.Components)
	if err != nil {
		t.Rollback()
		return []string{}, err
	}
	if err := t.Commit(); err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Inserts or updates ComponentArray entries in database within a single
// all-or-none transaction. If force=true, only the state, flag, subtype,
// nettype, and arch will be overwritten for existing components. Otherwise,
// this won't overwrite existing components.
func (d *hmsdbPg) UpsertComponents(comps []*base.Component, force bool) (map[string]map[string]bool, error) {
	affectedRowMap := make(map[string]map[string]bool, 0)
	cmap := make(map[string]*base.Component, 0)
	compList := make([]*base.Component, 0, 1)
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(comps))
	for i, comp := range comps {
		ids[i] = comp.ID
	}
	// Lock components for update
	affectedComps, err := t.GetComponentsTx(IDs(ids), From("UpsertComponents"))
	if err != nil {
		t.Rollback()
		return nil, err
	}
	for _, comp := range affectedComps {
		cmap[comp.ID] = comp
	}
	for _, comp := range comps {
		changeMap := make(map[string]bool, 0)
		aComp, ok := cmap[comp.ID]
		if ok {
			if !force {
				// Don't affect existing components
				continue
			} else {
				otherChange := false
				// Replace everything
				if comp.State != aComp.State {
					changeMap["state"] = true
				}
				if comp.Flag != aComp.Flag {
					changeMap["flag"] = true
				}
				if len(comp.Subtype) != 0 && comp.Subtype != aComp.Subtype {
					otherChange = true
				}
				if len(comp.NetType) != 0 && comp.NetType != aComp.NetType {
					otherChange = true
				}
				if len(comp.Arch) != 0 && comp.Arch != aComp.Arch {
					otherChange = true
				}
				if len(comp.Class) != 0 && comp.Class != aComp.Class {
					otherChange = true
				}

				// Move on if there are no changes
				if len(changeMap) == 0 && !otherChange {
					continue
				}
			}
		} else {
			// We are creating a component here
			changeMap["state"] = true
			changeMap["flag"] = true
			changeMap["enabled"] = true
			if len(comp.SwStatus) != 0 {
				changeMap["swStatus"] = true
			}
			changeMap["role"] = true
			changeMap["subRole"] = true
			changeMap["nid"] = true
		}
		compList = append(compList, comp)
		affectedRowMap[comp.ID] = changeMap
	}
	compsAffected, err := t.InsertComponentsTx(compList)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	if err := t.Commit(); err != nil {
		return nil, err
	}
	// Remove changes that didn't happen
	if len(compsAffected) != len(compList) {
		compsAffectedMap := make(map[string]string)
		for _, comp := range compsAffected {
			compsAffectedMap[comp] = comp
		}
		for _, comp := range compList {
			if id, ok := compsAffectedMap[comp.ID]; !ok {
				delete(affectedRowMap, id)
			}
		}
	}
	return affectedRowMap, nil
}

// Update state and flag fields only in DB for a list of components
//
//	Note: If flag is not set, it will be set to OK (i.e. no flag)
func (d *hmsdbPg) BulkUpdateCompState(ids []string, state string, flag string) ([]string, error) {
	return d.UpdateCompStates(ids, state, flag, false, new(PartInfo))
}

// Update state and flag fields only in DB for the given IDs.  If
// len(ids) is > 1 a locking read will be done to ensure the list o
// components that was actually modified is always returned.
//
// If force = true ignores any starting state restrictions and will
// always set ids to 'state', unless it is already set.
//
//	Note: If flag is not set, it will be set to OK (i.e. no flag)
func (d *hmsdbPg) UpdateCompStates(
	ids []string,
	state string,
	flag string,
	force bool,
	pi *PartInfo,
) ([]string, error) {
	// Verify input
	numIds := len(ids)
	fname := "UpdateCompStates"
	if numIds < 1 {
		d.LogAlways("Error: %s(): id list is empty", fname)
		return []string{}, ErrHMSDSArgMissing
	}
	nflag := base.VerifyNormalizeFlagOK(flag)

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}
	// We need to figure out the modified components.  If there are more
	// than one, we need to do a locking read with those that can be
	// updated given their current state and the allowed starting state
	// for the new state requested.
	// If there is just one, we don't need to worry about this because
	// there is no ambiguity as far as what components got updated and
	// which didn't.
	affectedIDs := []string{}
	if numIds == 1 {
		// Normalize the input as it comes from the user.
		idArray := []string{xnametypes.NormalizeHMSCompID(ids[0])}

		// Let the Update itself verify the starting states.
		cnt, err := t.UpdateCompStatesTx(idArray, state, nflag, force, false, pi)
		if err != nil {
			t.Rollback()
			return []string{}, err
		}
		if cnt > 0 {
			affectedIDs = idArray
			if cnt != 1 {
				d.LogAlways("ERROR: %s: cnt(%d) is not one!", fname, cnt)
			}
		}
	} else {
		var startStates []string
		nstate := base.VerifyNormalizeState(state)
		if nstate == "" {
			return []string{}, ErrHMSDSArgBadState
		}
		// The {State:"Ready", Flag:"Warning"} case needs to only affect components in
		// the "Ready" state. This prevents HBTD from accidentally turning nodes back
		// to Ready if the node state changed before HBTD noticed that the heartbeat
		// was late.
		if nstate == base.StateReady.String() && nflag == base.FlagWarning.String() {
			startStates = []string{base.StateReady.String()}
		} else {
			// Get list of required starting states, if any, given the requested
			// start state and the status of the force flag.
			// We only have to do this once when getting the affectedIDs.
			startStates, err = base.GetValidStartStateWForce(state, force)
			if err != nil {
				return []string{}, err
			}
		}
		// Lock components for update and select components we need to change.
		// This should produce normalized affectedIDs.
		affectedIDs, err = t.GetComponentIDsTx(IDs(ids),
			NotStateOrFlag(state, nflag), States(startStates),
			From(fname))
		if err != nil {
			t.Rollback()
			return []string{}, err
		}
		if affectedIDs != nil && len(affectedIDs) != 0 {
			// We already tested the list, so we don't need to do the
			// starting state filtering again, so this always works like force.
			_, err = t.UpdateCompStatesTx(affectedIDs, state, nflag,
				true, true, pi)
			if err != nil {
				t.Rollback()
				return []string{}, err
			}
		}
	}
	if err := t.Commit(); err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Update state and flag fields only in DB from those in c
// Returns the number of affected rows. < 0 means RowsAffected() is not
// supported.
//
//	Note: If flag is not set, it will be set to OK (i.e. no flag)
func (d *hmsdbPg) UpdateCompState(c *base.Component) (int64, error) {
	ids, err := d.UpdateCompStates([]string{c.ID}, c.State, c.Flag,
		false, new(PartInfo))
	return int64(len(ids)), err
}

// Update flag field in DB for a list of components
// Note: Flag cannot be empty/invalid.
func (d *hmsdbPg) BulkUpdateCompFlagOnly(ids []string, flag string) ([]string, error) {
	// Verify input
	if len(ids) < 1 {
		d.LogAlways("Error: BulkUpdateCompFlagOnly(): id list is empty")
		return nil, ErrHMSDSArgMissing
	}
	flag = base.VerifyNormalizeFlag(flag)
	if flag == "" {
		return []string{}, ErrHMSDSArgNoMatch
	}

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}
	// Lock components for update and get components that don't already have
	// flag
	// Lock components for update and select components we need to change.
	affectedIDs, err := t.GetComponentIDsTx(IDs(ids), Flag("!"+flag),
		From("BulkUpdateCompFlagOnly"))
	if err != nil {
		t.Rollback()
		return []string{}, err
	}
	if len(affectedIDs) != 0 {
		if _, err := t.BulkUpdateCompFlagOnlyTx(affectedIDs, flag); err != nil {
			t.Rollback()
			return []string{}, err
		}
	}
	if err := t.Commit(); err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Update Flag field in DB from c's Flag field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Flag cannot be blank/invalid.
func (d *hmsdbPg) UpdateCompFlagOnly(id string, flag string) (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	rowsAffected, err := t.UpdateCompFlagOnlyTx(id, flag)
	if err != nil {
		t.Rollback()
		return 0, err
	}
	if err := t.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Update enabled field in DB from c's Enabled field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: c.Enabled cannot be nil.
func (d *hmsdbPg) UpdateCompEnabled(id string, enabled bool) (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	rowsAffected, err := t.UpdateCompEnabledTx(id, enabled)
	if err != nil {
		t.Rollback()
		return 0, err
	}
	if err := t.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Update Enabled field only in DB for a list of components
func (d *hmsdbPg) BulkUpdateCompEnabled(ids []string, enabled bool) ([]string, error) {
	// Verify input
	if len(ids) < 1 {
		d.LogAlways("Error: BulkUpdateCompEnabled(): id list is empty")
		return nil, ErrHMSDSArgMissing
	}

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}
	// Lock components for update and select those that still need updates.
	affectedIDs, err := t.GetComponentIDsTx(IDs(ids),
		Enabled("!"+strconv.FormatBool(enabled)),
		From("BulkUpdateCompEnabled"))
	if err != nil {
		t.Rollback()
		return []string{}, err
	}
	if len(affectedIDs) != 0 {
		if _, err := t.BulkUpdateCompEnabledTx(affectedIDs, enabled); err != nil {
			t.Rollback()
			return []string{}, err
		}
	}
	if err := t.Commit(); err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Update SwStatus field in DB from c's SwStatus field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (d *hmsdbPg) UpdateCompSwStatus(id string, swStatus string) (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	rowsAffected, err := t.UpdateCompSwStatusTx(id, swStatus)
	if err != nil {
		t.Rollback()
		return 0, err
	}
	if err := t.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Update SwStatus field only in DB for a list of components
func (d *hmsdbPg) BulkUpdateCompSwStatus(ids []string, swstatus string) ([]string, error) {
	// Verify input
	if len(ids) < 1 {
		d.LogAlways("Error: BulkUpdateCompSwStatus(): id list is empty")
		return nil, ErrHMSDSArgMissing
	}

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}
	// Lock components for update
	affectedIDs, err := t.GetComponentIDsTx(IDs(ids), SwStatus("!"+swstatus),
		From("BulkUpdateCompSwStatus"))
	if err != nil {
		t.Rollback()
		return []string{}, err
	}
	if len(affectedIDs) != 0 {
		if _, err := t.BulkUpdateCompSwStatusTx(affectedIDs, swstatus); err != nil {
			t.Rollback()
			return []string{}, err
		}
	}
	if err := t.Commit(); err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Update Role/SubRole field in DB for a list of components
// Note: Role cannot be empty/invalid.
func (d *hmsdbPg) BulkUpdateCompRole(ids []string, role, subRole string) ([]string, error) {
	var affectedIDs []string
	// Verify input
	if len(ids) < 1 {
		d.LogAlways("Error: BulkUpdateCompRole(): id list is empty")
		return nil, ErrHMSDSArgMissing
	}
	role = base.VerifyNormalizeRole(role)
	if role == "" {
		return []string{}, ErrHMSDSArgNoMatch
	}
	// Allow SubRole to be empty
	if subRole != "" {
		subRole = base.VerifyNormalizeSubRole(subRole)
		if subRole == "" {
			return []string{}, ErrHMSDSArgNoMatch
		}
	}

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}
	// Lock components for update that still need changes (i.e. !role)
	if subRole == "" {
		affectedIDs, err = t.GetComponentIDsTx(IDs(ids), Role("!"+role),
			From("BulkUpdateCompRole"))
	} else {
		affectedIDs, err = t.GetComponentIDsTx(IDs(ids), NotRoleOrSubRole(role, subRole),
			From("BulkUpdateCompRole"))
	}
	if err != nil {
		t.Rollback()
		return []string{}, err
	}
	if len(affectedIDs) != 0 {
		if _, err := t.BulkUpdateCompRoleTx(affectedIDs, role, subRole); err != nil {
			t.Rollback()
			return []string{}, err
		}
	}
	if err := t.Commit(); err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Update Role/SubRole field in DB from c's Role/SubRole field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Role cannot be blank/invalid.
func (d *hmsdbPg) UpdateCompRole(id string, role, subRole string) (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	rowsAffected, err := t.UpdateCompRoleTx(id, role, subRole)
	if err != nil {
		t.Rollback()
		return 0, err
	}
	if err := t.Commit(); err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Update Class field only in DB for a list of components
func (d *hmsdbPg) BulkUpdateCompClass(ids []string, class string) ([]string, error) {
	// Verify input
	if len(ids) < 1 {
		d.LogAlways("Error: BulkUpdateCompClass(): id list is empty")
		return nil, ErrHMSDSArgMissing
	}

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}
	// Lock components for update
	affectedIDs, err := t.GetComponentIDsTx(IDs(ids), Class("!"+class),
		From("BulkUpdateCompClass"))
	if err != nil {
		t.Rollback()
		return []string{}, err
	}
	if len(affectedIDs) != 0 {
		if _, err := t.BulkUpdateCompClassTx(affectedIDs, class); err != nil {
			t.Rollback()
			return []string{}, err
		}
	}
	if err := t.Commit(); err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Update NID field in DB for a list of components
// Note: NID cannot be blank.  Should be negative to unset.
func (d *hmsdbPg) BulkUpdateCompNID(comps *[]base.Component) error {
	// Verify input
	if len(*comps) < 1 {
		d.LogAlways("Error: BulkUpdateCompNID(): Component list is empty")
		return ErrHMSDSArgMissing
	}

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return err
	}

	if err := t.BulkUpdateCompNIDTx(*comps); err != nil {
		t.Rollback()
		return err
	}

	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Update NID field in DB from c's NID field.
// Note: NID cannot be blank.  Should be negative to unset.
func (d *hmsdbPg) UpdateCompNID(c *base.Component) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	if err := t.UpdateCompNIDTx(c); err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Delete HMS Component with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteComponentByID(id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteComponentByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all HMS Components from database (atomically)
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteComponentsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteComponentsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
}

/////////////////////////////////////////////////////////////////////////////
//
// Node->NID Mapping
//
/////////////////////////////////////////////////////////////////////////////

// Look up one Node->NID Mapping by id, i.e. node xname.
func (d *hmsdbPg) GetNodeMapByID(id string) (*sm.NodeMap, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	m, err := t.GetNodeMapByIDTx(id)
	if err != nil {
		t.Rollback()
		return m, err
	}
	err = t.Commit()
	return m, err
}

// Look up ALL Node->NID Mappings.
func (d *hmsdbPg) GetNodeMapsAll() ([]*sm.NodeMap, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	nnms, err := t.GetNodeMapsAllTx()
	if err != nil {
		t.Rollback()
		return nnms, err
	}
	err = t.Commit()
	return nnms, err
}

// Insert Node->NID Mapping into database, updating it if it exists.
func (d *hmsdbPg) InsertNodeMap(m *sm.NodeMap) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertNodeMapTx(m)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Inserts or updates Node->NID Mapping Array entries in database within a
// single all-or-none transaction.
func (d *hmsdbPg) InsertNodeMaps(nnms *sm.NodeMapArray) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	for _, nnm := range nnms.NodeMaps {
		err := t.InsertNodeMapTx(nnm)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Delete Node NID Mapping entry with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteNodeMapByID(id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteNodeMapByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all Node NID Mapping entries from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteNodeMapsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteNodeMapsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
}

/////////////////////////////////////////////////////////////////////////////
//
// Power Mapping
//
/////////////////////////////////////////////////////////////////////////////

// Look up one Power Mapping by id, i.e. node xname.
func (d *hmsdbPg) GetPowerMapByID(id string) (*sm.PowerMap, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	m, err := t.GetPowerMapByIDTx(id)
	if err != nil {
		t.Rollback()
		return m, err
	}
	err = t.Commit()
	return m, err
}

// Look up ALL Power Mappings.
func (d *hmsdbPg) GetPowerMapsAll() ([]*sm.PowerMap, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	nnms, err := t.GetPowerMapsAllTx()
	if err != nil {
		t.Rollback()
		return nnms, err
	}
	err = t.Commit()
	return nnms, err
}

// Insert Power Mapping into database, updating it if it exists.
func (d *hmsdbPg) InsertPowerMap(m *sm.PowerMap) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertPowerMapTx(m)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Inserts or updates Power Mapping Array entries in database within a
// single all-or-none transaction.
func (d *hmsdbPg) InsertPowerMaps(ms []sm.PowerMap) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	for _, m := range ms {
		err := t.InsertPowerMapTx(&m)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Delete Power Mapping entry with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeletePowerMapByID(id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeletePowerMapByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all Power Mapping entries from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeletePowerMapsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeletePowerMapsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
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
func (d *hmsdbPg) GetHWInvByLocQueryFilter(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error) {
	query, err := getHWInvByLocQuery(f_opts...)
	if err != nil {
		return nil, err
	}

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	d.Log(LOG_DEBUG, "Debug: GetHWInvByLoc(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(d.sc).QueryContext(d.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwlocs := make([]*sm.HWInvByLoc, 0, 1)
	i := 0
	for rows.Next() {
		hwloc, err := d.scanHwInvByLocWithFRU(rows)
		if err != nil {
			d.LogAlways("Error: GetHWInvByLoc(): Scan failed: %s", err)
			return hwlocs, err
		}
		d.Log(LOG_DEBUG, "Debug: GetHWInvByLoc() scanned[%d]: %v", i, hwloc)
		hwlocs = append(hwlocs, hwloc)
		i += 1
	}
	err = rows.Err()
	d.Log(LOG_INFO, "Info: GetHWInvByLoc() returned %d hwinv items.", len(hwlocs))
	return hwlocs, err
}

// Get some or all Hardware Inventory entries with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct.
func (d *hmsdbPg) GetHWInvByLocFilter(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error) {
	var queryTable string

	// Parse the filter options
	f := new(HWInvLocFilter)
	for _, opts := range f_opts {
		opts(f)
	}

	if len(f.Partition) > 0 {
		queryTable = hwInvPartTable
	} else {
		queryTable = hwInvTable
	}

	query := sq.Select(addAliasToCols(hwInvAlias, hwInvCols, hwInvCols)...).
		From(queryTable + " " + hwInvAlias)
	if len(f.ID) > 0 {
		idCol := hwInvAlias + "." + hwInvIdCol
		idArgs := []string{}
		for _, id := range f.ID {
			idArgs = append(idArgs, xnametypes.NormalizeHMSCompID(id))
		}
		query = query.Where(sq.Eq{idCol: idArgs})
	}
	if len(f.Type) > 0 {
		typeCol := hwInvAlias + "." + hwInvTypeCol
		tArgs := []string{}
		for _, t := range f.Type {
			normType := xnametypes.VerifyNormalizeType(t)
			if normType == "" {
				return nil, ErrHMSDSArgBadType
			}
			tArgs = append(tArgs, normType)
		}
		query = query.Where(sq.Eq{typeCol: tArgs})
	}
	// Add Manufacturer filters if any. Use ILIKE to match
	// to make searches case insensitive.
	if len(f.Manufacturer) > 0 {
		mStr := "("
		mArgs := make([]interface{}, 0, 1)
		mCol := hwInvAlias + "." + hwInvFruInfoCol + " ->> 'Manufacturer'"
		for i, m := range f.Manufacturer {
			// Need to use one ILIKE per manufacturer specified because
			// ILIKE doesn't have alternative grouping specifiers (i.e. "()").
			if i > 0 {
				mStr += " OR " + mCol + " ILIKE ?"
			} else {
				mStr += mCol + " ILIKE ?"
			}
			str := "%" + m + "%"
			mArgs = append(mArgs, str)
		}
		mStr += ")"
		query = query.Where(sq.Expr(mStr, mArgs...))
	}
	if len(f.PartNumber) > 0 {
		pnCol := hwInvAlias + "." + hwInvFruInfoCol + " ->> 'PartNumber'"
		query = query.Where(sq.Eq{pnCol: f.PartNumber})
	}
	if len(f.SerialNumber) > 0 {
		pnCol := hwInvAlias + "." + hwInvFruInfoCol + " ->> 'SerialNumber'"
		query = query.Where(sq.Eq{pnCol: f.SerialNumber})
	}
	if len(f.FruId) > 0 {
		fruIdCol := hwInvAlias + "." + hwInvFruIdCol
		query = query.Where(sq.Eq{fruIdCol: f.FruId})
	}
	if len(f.Partition) > 0 {
		partCol := hwInvAlias + "." + hwInvPartPartitionCol
		query = query.Where(sq.Eq{partCol: f.Partition})
	}

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	d.Log(LOG_DEBUG, "Debug: GetHWInvByFRUFilter(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(d.sc).QueryContext(d.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwlocs := make([]*sm.HWInvByLoc, 0, 1)
	i := 0
	for rows.Next() {
		hwloc, err := d.scanHwInvByLocWithFRU(rows)
		if err != nil {
			d.LogAlways("Error: GetHWInvByLoc(): Scan failed: %s", err)
			return hwlocs, err
		}
		d.Log(LOG_DEBUG, "Debug: GetHWInvByLoc() scanned[%d]: %v", i, hwloc)
		hwlocs = append(hwlocs, hwloc)
		i += 1
	}
	err = rows.Err()
	d.Log(LOG_INFO, "Info: GetHWInvByLoc() returned %d hwinv items.", len(hwlocs))
	return hwlocs, err
}

// Get a single Hardware inventory entry by current xname
// This struct includes the FRU info if the xname is currently populated.
func (d *hmsdbPg) GetHWInvByLocID(id string) (*sm.HWInvByLoc, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	hwloc, err := t.GetHWInvByLocIDTx(id)
	if err != nil {
		t.Rollback()
		return hwloc, err
	}
	t.Commit()
	return hwloc, err
}

// Get HWInvByLoc by primary key (xname) for all entries in the system.
// It also pairs the data with the matching HWInvByFRU if the xname is
// populated.
func (d *hmsdbPg) GetHWInvByLocAll() ([]*sm.HWInvByLoc, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	hwlocs, err := t.GetHWInvByLocAllTx()
	if err != nil {
		t.Rollback()
		return hwlocs, err
	}
	t.Commit()
	return hwlocs, err
}

// Get HW Inventory-by-FRU entry at the provided location FRU ID
func (d *hmsdbPg) GetHWInvByFRUID(fruid string) (*sm.HWInvByFRU, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	hwfru, err := t.GetHWInvByFRUIDTx(fruid)
	if err != nil {
		t.Rollback()
		return hwfru, err
	}
	t.Commit()
	return hwfru, nil
}

// Get some or all Hardware Inventory entries with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct.
func (d *hmsdbPg) GetHWInvByFRUFilter(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByFRU, error) {
	// Parse the filter options
	f := new(HWInvLocFilter)
	for _, opts := range f_opts {
		opts(f)
	}

	query := sq.Select(addAliasToCols(hwInvFruAlias, hwInvFruTblCols, hwInvFruTblCols)...).
		From(hwInvFruTable + " " + hwInvFruAlias)
	if len(f.Type) > 0 {
		typeCol := hwInvFruAlias + "." + hwInvFruTblTypeCol
		tArgs := []string{}
		for _, t := range f.Type {
			normType := xnametypes.VerifyNormalizeType(t)
			if normType == "" {
				return nil, ErrHMSDSArgBadType
			}
			tArgs = append(tArgs, normType)
		}
		query = query.Where(sq.Eq{typeCol: tArgs})
	}
	// Add Manufacturer filters if any. Use ILIKE to match
	// to make searches case insensitive.
	if len(f.Manufacturer) > 0 {
		mStr := "("
		mArgs := make([]interface{}, 0, 1)
		mCol := hwInvFruAlias + "." + hwInvFruTblInfoCol + " ->> 'Manufacturer'"
		for i, m := range f.Manufacturer {
			// Need to use one ILIKE per manufacturer specified because
			// ILIKE doesn't have alternative grouping specifiers (i.e. "()").
			if i > 0 {
				mStr += " OR " + mCol + " ILIKE ?"
			} else {
				mStr += mCol + " ILIKE ?"
			}
			str := "%" + m + "%"
			mArgs = append(mArgs, str)
		}
		mStr += ")"
		query = query.Where(sq.Expr(mStr, mArgs...))
	}
	if len(f.PartNumber) > 0 {
		pnCol := hwInvFruAlias + "." + hwInvFruTblInfoCol + " ->> 'PartNumber'"
		query = query.Where(sq.Eq{pnCol: f.PartNumber})
	}
	if len(f.SerialNumber) > 0 {
		pnCol := hwInvFruAlias + "." + hwInvFruTblInfoCol + " ->> 'SerialNumber'"
		query = query.Where(sq.Eq{pnCol: f.SerialNumber})
	}
	if len(f.FruId) > 0 {
		fruIdCol := hwInvFruAlias + "." + hwInvFruTblIdCol
		query = query.Where(sq.Eq{fruIdCol: f.FruId})
	}

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	d.Log(LOG_DEBUG, "Debug: GetHWInvByFRUFilter(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(d.sc).QueryContext(d.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwfrus := make([]*sm.HWInvByFRU, 0, 1)
	i := 0
	for rows.Next() {
		hwfru, err := d.scanHwInvByFRU(rows)
		if err != nil {
			d.LogAlways("Error: GetHWInvByFRUFilter(): Scan failed: %s", err)
			return hwfrus, err
		}
		d.Log(LOG_DEBUG, "Debug: GetHWInvByFRUFilter() scanned[%d]: %v", i, hwfru)
		hwfrus = append(hwfrus, hwfru)
		i += 1
	}
	err = rows.Err()
	d.Log(LOG_INFO, "Info: GetHWInvByFRUFilter() returned %d hwinv items.", len(hwfrus))
	return hwfrus, err
}

// Get all HW-inventory-by-FRU entries.
func (d *hmsdbPg) GetHWInvByFRUAll() ([]*sm.HWInvByFRU, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	hwfrus, err := t.GetHWInvByFRUAllTx()
	if err != nil {
		t.Rollback()
		return hwfrus, err
	}
	t.Commit()
	return hwfrus, err
}

// Insert or update HWInventoryByLocation struct.
// If PopulatedFRU is present, this is also added to the DB  If
// it is not, this effectively "depopulates" the given location.
// The actual HWInventoryByFRU is stored using within the same
// transaction.
func (d *hmsdbPg) InsertHWInvByLoc(hl *sm.HWInvByLoc) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	if hl.PopulatedFRU != nil {
		err = t.InsertHWInvByFRUTx(hl.PopulatedFRU)
		if err != nil {
			t.Rollback()
			return err
		}
	}
	err = t.InsertHWInvByLocTx(hl)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Insert or update HWInventoryByFRU struct.  This does not associate
// the object with any HW-Inventory-By-Location info so it is
// typically not needed.  InsertHWInvByLoc is typically used to
// store both type of info at once.
func (d *hmsdbPg) InsertHWInvByFRU(hf *sm.HWInvByFRU) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertHWInvByFRUTx(hf)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Insert or update array of HWInventoryByLocation structs.
// If PopulatedFRU is present, these is also added to the DB  If
// it is not, this effectively "depopulates" the given locations.
// The actual HWInventoryByFRU is stored using within the same
// transaction.
func (d *hmsdbPg) InsertHWInvByLocs(hls []*sm.HWInvByLoc) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	hfs := make([]*sm.HWInvByFRU, 0, len(hls))
	// Insert FRUs first because the location info links to them.
	for _, hl := range hls {
		if hl.PopulatedFRU != nil {
			hfs = append(hfs, hl.PopulatedFRU)
		}
	}
	err = t.BulkInsertHWInvByFRUTx(hfs)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.BulkInsertHWInvByLocTx(hls)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Delete HWInvByLoc entry with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteHWInvByLocID(id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteHWInvByLocIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete ALL HWInvByLoc entries from database (atomically)
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteHWInvByLocsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteHWInvByLocsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
}

// Delete HWInvByFRU entry with matching FRU ID from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteHWInvByFRUID(fruid string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteHWInvByFRUIDTx(fruid)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete ALL HWInvByFRU entries from database (atomically)
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteHWInvByFRUsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteHWInvByFRUsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
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
func (d *hmsdbPg) GetHWInvHistFilter(f_opts ...HWInvHistFiltFunc) ([]*sm.HWInvHist, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	hhs, err := t.GetHWInvHistFilterTx(f_opts...)
	if err != nil {
		t.Rollback()
		return hhs, err
	}
	err = t.Commit()
	return hhs, err
}

// Get only the most recent hardware history event for some or all hardware locations.
func (d *hmsdbPg) GetHWInvHistLastEvents(ids []string) ([]*sm.HWInvHist, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	hhs, err := t.GetHWInvHistLastEventsTx(ids)
	if err != nil {
		t.Rollback()
		return hhs, err
	}
	err = t.Commit()
	return hhs, err
}

// Insert a HWInventoryHistory entry.
// If a duplicate is present return an error.
func (d *hmsdbPg) InsertHWInvHist(hh *sm.HWInvHist) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}

	err = t.InsertHWInvHistTx(hh)
	if err != nil {
		t.Rollback()
		return err
	}

	err = t.Commit()
	return err
}

// Insert an array of HWInventoryHistory entries.
// If a duplicate is present return an error.
func (d *hmsdbPg) InsertHWInvHists(hhs []*sm.HWInvHist) error {
	if len(hhs) == 0 {
		// Nothing to do
		return nil
	}
	t, err := d.Begin()
	if err != nil {
		return err
	}
	// Insert HWInvHist entries.
	err = t.InsertHWInvHistsTx(hhs)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Delete all HWInvHist entries with matching xname id from database, if it
// exists.
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteHWInvHistByLocID(id string) (int64, error) {
	// Build query
	query := sq.Delete(hwInvHistTable).
		Where(sq.Eq{hwInvHistIdCol: id})

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(d.sc).ExecContext(d.ctx)
	if err != nil {
		return 0, err
	}
	// See if any rows were affected
	return res.RowsAffected()
}

// Delete all HWInvHist entries with matching FRU id from database, if it
// exists.
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteHWInvHistByFRUID(fruid string) (int64, error) {
	// Build query
	query := sq.Delete(hwInvHistTable).
		Where(sq.Eq{hwInvHistFruIdCol: fruid})

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(d.sc).ExecContext(d.ctx)
	if err != nil {
		return 0, err
	}
	// See if any rows were affected
	return res.RowsAffected()
}

// Delete all HWInvHist entries from database (atomically)
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteHWInvHistAll() (int64, error) {
	// Build query
	query := sq.Delete(hwInvHistTable)

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(d.sc).ExecContext(d.ctx)
	if err != nil {
		return 0, err
	}
	// See if any rows were affected
	return res.RowsAffected()
}

// Delete all HWInvHist entries from database matching a filter.
// Returns the number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteHWInvHistFilter(f_opts ...HWInvHistFiltFunc) (int64, error) {
	// Parse the filter options
	f := new(HWInvHistFilter)
	for _, opts := range f_opts {
		opts(f)
	}

	// Build query
	query := sq.Delete(hwInvHistTable)
	if len(f.ID) > 0 {
		query = query.Where(sq.Eq{hwInvHistIdCol: f.ID})
	}
	if len(f.FruId) > 0 {
		query = query.Where(sq.Eq{hwInvHistFruIdCol: f.FruId})
	}
	if len(f.EventType) > 0 {
		tArgs := []string{}
		for _, evt := range f.EventType {
			normEvt := sm.VerifyNormalizeHWInvHistEventType(evt)
			if normEvt == "" {
				return 0, ErrHMSDSArgBadHWInvHistEventType
			}
			tArgs = append(tArgs, normEvt)
		}
		query = query.Where(sq.Eq{hwInvHistEventTypeCol: tArgs})
	}
	if f.StartTime != "" {
		start, err := time.Parse(time.RFC3339, f.StartTime)
		if err != nil {
			return 0, ErrHMSDSArgBadTimeFormat
		}
		query = query.Where(sq.Gt{hwInvHistTimestampCol: start})
	}
	if f.EndTime != "" {
		end, err := time.Parse(time.RFC3339, f.EndTime)
		if err != nil {
			return 0, ErrHMSDSArgBadTimeFormat
		}
		query = query.Where(sq.Lt{hwInvHistTimestampCol: end})
	}

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(d.sc).ExecContext(d.ctx)
	if err != nil {
		return 0, err
	}
	// See if any rows were affected
	return res.RowsAffected()
}

////////////////////////////////////////////////////////////////////////////
//
// Redfish Endpoints - Top-level Redfish service roots used for discovery
//
////////////////////////////////////////////////////////////////////////////

// Get RedfishEndpoint by ID (xname), i.e. a single entry.
func (d *hmsdbPg) GetRFEndpointByID(id string) (*sm.RedfishEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	ep, err := t.GetRFEndpointByIDTx(id)
	if err != nil {
		t.Rollback()
		return ep, err
	}
	t.Commit()
	return ep, nil
}

// Get all RedfishEndpoints in system.
func (d *hmsdbPg) GetRFEndpointsAll() ([]*sm.RedfishEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	eps, err := t.GetRFEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return eps, err
	}
	t.Commit()
	return eps, nil
}

// Get some or all RedfishEndpoints in system, with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbPg) GetRFEndpointsFilter(f *RedfishEPFilter) ([]*sm.RedfishEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	reps, err := t.GetRFEndpointsFilterTx(f)
	if err != nil {
		t.Rollback()
		return reps, err
	}
	t.Commit()
	return reps, nil
}

// Insert new RedfishEndpoint into database.
// Does not update any ComponentEndpoint children.
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (d *hmsdbPg) InsertRFEndpoint(ep *sm.RedfishEndpoint) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertRFEndpointTx(ep)
	if err != nil {
		t.Rollback()
		return err
	}
	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Insert new RedfishEndpointArray entries into database within a
// single all-or-none transaction.  Does not update any ComponentEndpoint
// children.
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (d *hmsdbPg) InsertRFEndpoints(eps *sm.RedfishEndpointArray) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}

	err = t.InsertRFEndpointsTx(eps.RedfishEndpoints)
	if err != nil {
		t.Rollback()
		return err
	}

	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Update existing RedfishEndpointArray entry in database.
// Does not update any ComponentEndpoint children.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbPg) UpdateRFEndpoint(ep *sm.RedfishEndpoint) (*sm.RedfishEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	didUpdate, err := t.UpdateRFEndpointTx(ep)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	getEP, err := t.GetRFEndpointByIDTx(ep.ID)
	if err != nil {
		if didUpdate != true {
			// No update because there was no entry
			t.Rollback()
			return nil, nil
		} else {
			t.Rollback()
			return nil, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, err
	}
	return getEP, nil
}

// Update existing RedfishEndpointArray entry in database, but only updates
// fields that would be changed by a user-directed operation.
// Does not update any ComponentEndpoint children.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbPg) UpdateRFEndpointNoDiscInfo(ep *sm.RedfishEndpoint) (*sm.RedfishEndpoint, []string, error) {
	affectedIDs := []string{}
	t, err := d.Begin()
	if err != nil {
		return nil, []string{}, err
	}
	if !ep.Enabled {
		// Set all State/Components entries to Empty/OK if the ComponentEndpoint
		// of the same name exists and is a child of RedfishEndpoint id.
		// This also locks the affected rows in all three tables.
		affectedIDs, err = t.SetChildCompStatesRFEndpointsTx([]string{ep.ID},
			base.StateEmpty.String(), base.FlagOK.String(), true, true)
		if err != nil {
			t.Rollback()
			return nil, []string{}, err
		}
	}
	didUpdate, err := t.UpdateRFEndpointNoDiscInfoTx(ep)
	if err != nil {
		t.Rollback()
		return nil, []string{}, err
	}
	getEP, err := t.GetRFEndpointByIDTx(ep.ID)
	if err != nil {
		if didUpdate != true {
			// No update because there was no entry
			t.Rollback()
			return nil, []string{}, nil
		} else {
			t.Rollback()
			return nil, []string{}, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, []string{}, err
	}
	return getEP, affectedIDs, nil
}

// Patch existing RedfishEndpointArray entry in database, but only updates
// specified fields.
// Does not update any ComponentEndpoint children.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbPg) PatchRFEndpointNoDiscInfo(id string, epp sm.RedfishEndpointPatch) (*sm.RedfishEndpoint, []string, error) {
	var (
		rep        rf.RawRedfishEP
		haveUpdate bool
		reps       []*sm.RedfishEndpoint
	)

	affectedIDs := []string{}
	label := "PatchRFEndpointNoDiscInfo"
	t, err := d.Begin()
	if err != nil {
		return nil, []string{}, err
	}
	if epp.Enabled != nil && *epp.Enabled == false {
		// Set all State/Components entries to Empty/OK if the ComponentEndpoint
		// of the same name exists and is a child of RedfishEndpoint id.
		// This also locks the affected rows in all three tables.
		affectedIDs, err = t.SetChildCompStatesRFEndpointsTx([]string{id},
			base.StateEmpty.String(), base.FlagOK.String(), true, true)
		if err != nil {
			t.Rollback()
			return nil, []string{}, err
		}
		reps, err = t.GetRFEndpointsTx(
			RFE_From(label),
			RFE_ID(id))
	} else {
		reps, err = t.GetRFEndpointsTx(
			RFE_From(label),
			RFE_ID(id))
	}
	if err != nil {
		t.Rollback()
		return nil, []string{}, err
	} else if len(reps) != 1 {
		t.Rollback()
		return nil, []string{}, nil
	}
	getEP := reps[0]
	rep.ID = getEP.ID
	if epp.Type != nil && getEP.Type != *epp.Type {
		rep.Type = *epp.Type
		haveUpdate = true
	} else {
		rep.Type = getEP.Type
	}
	if epp.Name != nil && getEP.Name != *epp.Name {
		rep.Name = *epp.Name
		haveUpdate = true
	} else {
		rep.Name = getEP.Name
	}
	if epp.Hostname != nil && getEP.Hostname != *epp.Hostname {
		rep.Hostname = *epp.Hostname
		haveUpdate = true
	} else {
		rep.Hostname = getEP.Hostname
	}
	if epp.Domain != nil && getEP.Domain != *epp.Domain {
		rep.Domain = *epp.Domain
		haveUpdate = true
	} else {
		rep.Domain = getEP.Domain
	}
	if epp.FQDN != nil && getEP.FQDN != *epp.FQDN {
		rep.FQDN = *epp.FQDN
		haveUpdate = true
	} else {
		rep.FQDN = getEP.FQDN
	}
	if epp.Enabled != nil && getEP.Enabled != *epp.Enabled {
		rep.Enabled = epp.Enabled
		haveUpdate = true
	} else {
		rep.Enabled = &getEP.Enabled
	}
	if epp.UUID != nil && getEP.UUID != *epp.UUID {
		rep.UUID = *epp.UUID
		haveUpdate = true
	} else {
		rep.UUID = getEP.UUID
	}
	if epp.User != nil && getEP.User != *epp.User {
		rep.User = *epp.User
		haveUpdate = true
	} else {
		rep.User = getEP.User
	}
	if epp.Password != nil && getEP.Password != *epp.Password {
		rep.Password = *epp.Password
		haveUpdate = true
	} else {
		rep.Password = getEP.Password
	}
	if epp.UseSSDP != nil && getEP.UseSSDP != *epp.UseSSDP {
		rep.UseSSDP = epp.UseSSDP
		haveUpdate = true
	} else {
		rep.UseSSDP = &getEP.UseSSDP
	}
	if epp.MACRequired != nil && getEP.MACRequired != *epp.MACRequired {
		rep.MACRequired = epp.MACRequired
		haveUpdate = true
	} else {
		rep.MACRequired = &getEP.MACRequired
	}
	if epp.MACAddr != nil && getEP.MACAddr != *epp.MACAddr {
		rep.MACAddr = *epp.MACAddr
		haveUpdate = true
	} else {
		rep.MACAddr = getEP.MACAddr
	}
	if epp.IPAddr != nil && getEP.IPAddr != *epp.IPAddr {
		rep.IPAddr = *epp.IPAddr
		// If the hostname/FQDN is an IP address, update it as well
		hostnameIP := rf.GetIPAddressString(rep.Hostname)
		if hostnameIP != "" && hostnameIP != rep.IPAddr {
			rep.Hostname = rep.IPAddr
			rep.FQDN = rep.IPAddr
		}
		haveUpdate = true
	} else {
		rep.IPAddr = getEP.IPAddr
	}
	if epp.RediscOnUpdate != nil && getEP.RediscOnUpdate != *epp.RediscOnUpdate {
		rep.RediscOnUpdate = epp.RediscOnUpdate
		haveUpdate = true
	} else {
		rep.RediscOnUpdate = &getEP.RediscOnUpdate
	}
	if epp.TemplateID != nil && getEP.TemplateID != *epp.TemplateID {
		rep.TemplateID = *epp.TemplateID
		haveUpdate = true
	} else {
		rep.TemplateID = getEP.TemplateID
	}
	if !haveUpdate {
		t.Rollback()
		return getEP, []string{}, nil
	}
	// Validate new RedfishEndpoint data
	epd, err := rf.NewRedfishEPDescription(&rep)
	if err != nil {
		t.Rollback()
		return nil, []string{}, err
	}
	ep := sm.NewRedfishEndpoint(epd)
	didUpdate, err := t.UpdateRFEndpointNoDiscInfoTx(ep)
	if err != nil {
		t.Rollback()
		return nil, []string{}, err
	}
	updEP, err := t.GetRFEndpointByIDTx(ep.ID)
	if err != nil {
		if didUpdate != true {
			// No update because there was no entry
			t.Rollback()
			return nil, []string{}, nil
		} else {
			t.Rollback()
			return nil, []string{}, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, []string{}, err
	}
	return updEP, affectedIDs, nil
}

// Returns: Discoverable endpoint list, with status set appropriately in DB
// and return values.  However this list will omit those RF EPs  who are
// already being discovered, unless forced.
// Error returned on unexpected failure or any entry in eps not existing,
// the latter error being ErrHMSDSNoREP.
func (d *hmsdbPg) UpdateRFEndpointForDiscover(ids []string, force bool) (
	[]*sm.RedfishEndpoint, error) {

	label := "UpdateRFEndpointForDiscover"
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	// Get current endpoints to see if they are already being discovered
	reps, err := t.GetRFEndpointsTx(
		RFE_From(label),
		RFE_IDs(ids))
	if err != nil {
		t.Rollback()
		return nil, err
	} else if len(reps) != len(ids) {
		return nil, ErrHMSDSNoREP
	}
	modEPs := make([]*sm.RedfishEndpoint, 0, len(reps))
	for _, rep := range reps {
		if force || rep.DiscInfo.LastStatus != rf.DiscoveryStarted {
			modEP := sm.NewRedfishEndpoint(&rep.RedfishEPDescription)
			modEP.DiscInfo = rep.DiscInfo
			modEP.DiscInfo.UpdateLastStatusWithTS(rf.DiscoveryStarted)
			modEPs = append(modEPs, modEP)
		}
	}
	if len(modEPs) > 0 {
		_, err := t.UpdateRFEndpointsTx(modEPs)
		if err != nil {
			t.Rollback()
			return nil, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, err
	}
	return modEPs, nil
}

// Update existing RedfishEndpointArray entries in database within a
// single all-or-none transaction.  Does not update any ComponentEndpoint
// children.
// Returns FALSE with err == nil if one or more updated entries do
// not exist.  No updates are performed in this case.
func (d *hmsdbPg) UpdateRFEndpoints(eps *sm.RedfishEndpointArray) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}

	getEPs, err := t.UpdateRFEndpointsTx(eps.RedfishEndpoints)
	if err != nil {
		t.Rollback()
		return false, err
	} else if len(getEPs) != len(eps.RedfishEndpoints) {
		// An entry didn't exist.
		t.Rollback()
		return false, nil
	}

	if err := t.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

// Delete RedfishEndpoint with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteRFEndpointByID(id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteRFEndpointByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all RedfishEndpoints from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteRFEndpointsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteRFEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
}

// Delete RedfishEndpoint with matching xname id from database, if it
// exists.  When dooing so, set all HMS Components to Empty if they
// are children of the RedfishEndpoint.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteRFEndpointByIDSetEmpty(id string) (bool, []string, error) {
	if !base.IsAlphaNum(id) {
		return false, []string{}, ErrHMSDSArgBadID
	}
	t, err := d.Begin()
	if err != nil {
		return false, []string{}, err
	}
	// Set all State/Components entries to Empty/OK if the ComponentEndpoint
	// of the same name exists and is a child of RedfishEndpoint id.
	// This also locks the affected rows in all three tables.
	affectedIDs, err := t.SetChildCompStatesRFEndpointsTx([]string{id},
		base.StateEmpty.String(), base.FlagOK.String(), true, true)
	if err != nil {
		t.Rollback()
		return false, []string{}, err
	}
	// Now, delete the requested RF endpoint.
	didDelete, err := t.DeleteRFEndpointByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, []string{}, err
	}
	if didDelete == false {
		// Shouldn't have modified anything, so don't.
		t.Rollback()
		return false, []string{}, nil
	}
	// OK, commit transaction and release locks
	if err := t.Commit(); err != nil {
		return false, []string{}, err
	}
	// Commit ok, return info since everything went through.
	return true, affectedIDs, nil
}

// Delete all RedfishEndpoints from database.
// This also deletes all child ComponentEndpoints, and in addition,
// sets the State/Components entries for those ComponentEndpoints to Empty/OK
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteRFEndpointsAllSetEmpty() (int64, []string, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, []string{}, err
	}
	// Set all State/Components entries to Empty/OK if a ComponentEndpoint
	// of the same name exists as a child of ANY RedfishEndpoint id.
	// This also locks the affected rows in all three tables.
	affectedIDs, err := t.SetChildCompStatesRFEndpointsTx([]string{"!"},
		base.StateEmpty.String(), base.FlagOK.String(), true, true)
	if err != nil {
		t.Rollback()
		return 0, []string{}, err
	}
	numDeleted, err := t.DeleteRFEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return 0, []string{}, err
	}
	if numDeleted == 0 {
		// Shouldn't have modified anything, so don't.
		t.Rollback()
		return 0, []string{}, nil
	}
	if err := t.Commit(); err != nil {
		return 0, []string{}, err
	}
	return numDeleted, affectedIDs, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Component Endpoints - Component info discovered from parent RedfishEndpoint
//                       Management plane equivalent to HMS Component
//
////////////////////////////////////////////////////////////////////////////

// Get ComponentEndpoint by id (xname), i.e. a single entry.
func (d *hmsdbPg) GetCompEndpointByID(id string) (*sm.ComponentEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	cep, err := t.GetCompEndpointByIDTx(id)
	if err != nil {
		t.Rollback()
		return cep, err
	}
	t.Commit()
	return cep, err
}

// Get all ComponentEndpoints in system.
func (d *hmsdbPg) GetCompEndpointsAll() ([]*sm.ComponentEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	ceps, err := t.GetCompEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return ceps, err
	}
	t.Commit()
	return ceps, nil
}

// Get some or all ComponentEndpoints in system, with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbPg) GetCompEndpointsFilter(f *CompEPFilter) ([]*sm.ComponentEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	ceps, err := t.GetCompEndpointsFilterTx(f)
	if err != nil {
		t.Rollback()
		return ceps, err
	}
	t.Commit()
	return ceps, nil
}

// Upsert ComponentEndpoint into database, updating it if it exists.
func (d *hmsdbPg) UpsertCompEndpoint(cep *sm.ComponentEndpoint) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.UpsertCompEndpointTx(cep)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Upsert ComponentEndpointArray into database within a single all-or-none
// transaction.
func (d *hmsdbPg) UpsertCompEndpoints(ceps *sm.ComponentEndpointArray) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.UpsertCompEndpointsTx(ceps)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Delete ComponentEndpoint with matching xname id from database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteCompEndpointByID(id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteCompEndpointByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all ComponentEndpoints from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteCompEndpointsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteCompEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
}

// Delete ComponentEndpoint with matching xname id from database, if it
// exists.  When dooing so, set the corresponding HMS Component to Empty if it
// is not already in that state.
// Return true if there was a row affected, false if there were zero.  The
// string array returns the single xname ID that changed state or is empty.
func (d *hmsdbPg) DeleteCompEndpointByIDSetEmpty(id string) (bool, []string, error) {
	if !base.IsAlphaNum(id) {
		return false, []string{}, ErrHMSDSArgBadID
	}
	t, err := d.Begin()
	if err != nil {
		return false, []string{}, err
	}
	// Set the matching State/Components entry to Empty/OK if it exists.
	// This also locks the affected rows in all three tables.
	affectedIDs, err := t.SetChildCompStatesCompEndpointsTx([]string{id},
		base.StateEmpty.String(), base.FlagOK.String(), true)
	if err != nil {
		t.Rollback()
		return false, []string{}, err
	}
	// Now, delete the requested ComponentEndpoint.
	didDelete, err := t.DeleteCompEndpointByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, []string{}, err
	}
	if didDelete == false {
		// Shouldn't have modified anything, so don't.
		t.Rollback()
		return false, []string{}, nil
	}
	// OK, commit transaction and release locks
	if err := t.Commit(); err != nil {
		return false, []string{}, err
	}
	// Commit ok, return info since everything went through.
	return true, affectedIDs, nil
}

// Delete all ComponentEndpoints from database. In addition,
// sets the State/Components entry for each ComponentEndpoint to Empty/OK
// Also returns number of deleted rows, if error is nil, and also string array
// of those xname IDs that were set to Empty/OK (i.e. not already Empty/OK)
// as part of the deletion.
func (d *hmsdbPg) DeleteCompEndpointsAllSetEmpty() (int64, []string, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, []string{}, err
	}
	// Set all State/Components entries to Empty/OK if a matching
	// ComponentEndpoint (i.e. with the same xname ID) exists.
	// This also locks the affected rows in all three tables until
	// the transaction is done (SELECT ... FOR UPDATE).
	// Here we use a wildcard that says any non-empty ID, which means all
	// ComponentEndpoints, since it is a mandatory value.
	affectedIDs, err := t.SetChildCompStatesCompEndpointsTx([]string{"!"},
		base.StateEmpty.String(), base.FlagOK.String(), true)
	if err != nil {
		t.Rollback()
		return 0, []string{}, err
	}
	numDeleted, err := t.DeleteCompEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return 0, []string{}, err
	}
	if numDeleted == 0 {
		// Shouldn't have modified anything, so don't.
		t.Rollback()
		return 0, []string{}, nil
	}
	if err := t.Commit(); err != nil {
		return 0, []string{}, err
	}
	return numDeleted, affectedIDs, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Service Endpoints - Service info discovered from parent RedfishEndpoint
//
////////////////////////////////////////////////////////////////////////////

// Get ServiceEndpoint by service type and id (xname), i.e. a single entry.
func (d *hmsdbPg) GetServiceEndpointByID(svc, id string) (*sm.ServiceEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	sep, err := t.GetServiceEndpointByIDTx(svc, id)
	if err != nil {
		t.Rollback()
		return sep, err
	}
	t.Commit()
	return sep, err
}

// Get all ServiceEndpoints in system.
func (d *hmsdbPg) GetServiceEndpointsAll() ([]*sm.ServiceEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	seps, err := t.GetServiceEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return seps, err
	}
	t.Commit()
	return seps, nil
}

// Get some or all ServiceEndpoints in system, with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbPg) GetServiceEndpointsFilter(f *ServiceEPFilter) ([]*sm.ServiceEndpoint, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	seps, err := t.GetServiceEndpointsFilterTx(f)
	if err != nil {
		t.Rollback()
		return seps, err
	}
	t.Commit()
	return seps, nil
}

// Upsert ServiceEndpoint into database, updating it if it exists.
func (d *hmsdbPg) UpsertServiceEndpoint(sep *sm.ServiceEndpoint) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.UpsertServiceEndpointTx(sep)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Upsert ServiceEndpointArray into database within a single all-or-none
// transaction.
func (d *hmsdbPg) UpsertServiceEndpoints(seps *sm.ServiceEndpointArray) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.UpsertServiceEndpointsTx(seps)
	if err != nil {
		t.Rollback()
		return err
	}
	err = t.Commit()
	return err
}

// Delete ServiceEndpoint with matching service type and xname id from
// database, if it exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteServiceEndpointByID(svc, id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteServiceEndpointByIDTx(svc, id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all ServiceEndpoints from database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteServiceEndpointsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteServiceEndpointsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Component Ethernet Interfaces - MAC address to IP address relations for
//     component endpoint ethernet interfaces.
//
////////////////////////////////////////////////////////////////////////////

// Get some or all CompEthInterfaces in the system, with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (d *hmsdbPg) GetCompEthInterfaceFilter(f_opts ...CompEthInterfaceFiltFunc) ([]*sm.CompEthInterfaceV2, error) {
	// Parse the filter options
	f := new(CompEthInterfaceFilter)
	for _, opts := range f_opts {
		opts(f)
	}

	query := sq.Select(addAliasToCols(compEthAlias, compEthCols, compEthCols)...).
		From(compEthTable + " " + compEthAlias)

	if len(f.IPAddr) > 0 || len(f.Network) > 0 {
		// If searching on IP address or network multiple rows could be returned for the same mac address
		query = query.Options("DISTINCT ON(", compEthIdColAlias, ")")
	}
	if len(f.IPAddr) > 0 {
		predicate := fmt.Sprintf("COALESCE(ip->>'%s', '')", compEthJsonIPAddress)
		query = query.JoinClause(fmt.Sprintf("LEFT JOIN LATERAL json_array_elements(%s) ip ON true", compEthIPAddressesAlias)).
			Where(sq.Eq{predicate: f.IPAddr})
	}
	if len(f.Network) > 0 {
		predicate := fmt.Sprintf("COALESCE(ip->>'%s', '')", compEthJsonNetwork)
		query = query.JoinClause(fmt.Sprintf("LEFT JOIN LATERAL json_array_elements(%s) ip ON true", compEthIPAddressesAlias)).
			Where(sq.Eq{predicate: f.Network})
	}

	if len(f.ID) > 0 {
		idCol := compEthAlias + "." + compEthIdCol
		query = query.Where(sq.Eq{idCol: f.ID})
	}
	if len(f.MACAddr) > 0 {
		macCol := compEthAlias + "." + compEthMACAddrCol
		query = query.Where(sq.Eq{macCol: f.MACAddr})
	}
	if f.NewerThan != "" {
		tsCol := compEthAlias + "." + compEthLastUpdateCol
		nt, err := time.Parse(time.RFC3339, f.NewerThan)
		if err != nil {
			return nil, ErrHMSDSArgBadTimeFormat
		}
		query = query.Where(sq.Gt{tsCol: nt})
	}
	if f.OlderThan != "" {
		tsCol := compEthAlias + "." + compEthLastUpdateCol
		ot, err := time.Parse(time.RFC3339, f.OlderThan)
		if err != nil {
			return nil, ErrHMSDSArgBadTimeFormat
		}
		query = query.Where(sq.Lt{tsCol: ot})
	}
	if len(f.CompID) > 0 {
		idCol := compEthAlias + "." + compEthCompIDCol
		query = query.Where(sq.Eq{idCol: f.CompID})
	}
	if len(f.CompType) > 0 {
		typeCol := compEthAlias + "." + compEthTypeCol
		query = query.Where(sq.Eq{typeCol: f.CompType})
	}

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	d.Log(LOG_DEBUG, "Debug: GetCompEthInterfaceFilter(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(d.sc).QueryContext(d.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ceis := make([]*sm.CompEthInterfaceV2, 0, 1)
	i := 0
	for rows.Next() {
		var ipAddresses []byte

		cei := new(sm.CompEthInterfaceV2)
		err := rows.Scan(&cei.ID, &cei.Desc, &cei.MACAddr, &cei.LastUpdate, &cei.CompID, &cei.Type, &ipAddresses)
		if err != nil {
			d.LogAlways("Error: GetCompEthInterfaceFilter(): Scan failed: %s", err)
			return ceis, err
		}

		err = json.Unmarshal(ipAddresses, &cei.IPAddrs)
		if err != nil {
			d.LogAlways("Warning: GetCompEthInterfaceFilter(): Decode IPAddresses: %s", err)
			return nil, err
		}

		d.Log(LOG_DEBUG, "Debug: GetCompEthInterfaceFilter() scanned[%d]: %v", i, cei)
		ceis = append(ceis, cei)
		i += 1
	}
	err = rows.Err()
	d.Log(LOG_INFO, "Info: GetCompEthInterfaceFilter() returned %d CompEthInterface items.", len(ceis))
	return ceis, err
}

// Insert a new CompEthInterface into the database.
// If ID or MAC address already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (d *hmsdbPg) InsertCompEthInterface(cei *sm.CompEthInterfaceV2) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertCompEthInterfaceTx(cei)
	if err != nil {
		t.Rollback()
		return err
	}
	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Insert new CompEthInterfaces into the database within a single
// all-or-none transaction.
// If ID or MAC address already exists, return ErrHMSDSDuplicateKey
// No insertions are done on err != nil
func (d *hmsdbPg) InsertCompEthInterfaces(ceis []*sm.CompEthInterfaceV2) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertCompEthInterfacesTx(ceis)
	if err != nil {
		t.Rollback()
		return err
	}
	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Insert/update a CompEthInterface in the database.
// If ID or MAC address already exists, only overwrite ComponentID
// and Type fields.
// No insertion done on err != nil
func (d *hmsdbPg) InsertCompEthInterfaceCompInfo(cei *sm.CompEthInterfaceV2) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertCompEthInterfaceCompInfoTx(cei)
	if err != nil {
		t.Rollback()
		return err
	}
	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Insert new CompEthInterfaces into database within a single
// all-or-none transaction.
// If ID or MAC address already exists, only overwrite ComponentID
// and Type fields.
// No insertions are done on err != nil
func (d *hmsdbPg) InsertCompEthInterfacesCompInfo(ceis []*sm.CompEthInterfaceV2) error {
	t, err := d.Begin()
	if err != nil {
		return err
	}
	err = t.InsertCompEthInterfacesCompInfoTx(ceis)
	if err != nil {
		t.Rollback()
		return err
	}
	if err := t.Commit(); err != nil {
		return err
	}
	return nil
}

// Patch existing CompEthInterface entry in database, but only updates
// specified fields.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbPg) UpdateCompEthInterface(id string, ceip *sm.CompEthInterfaceV2Patch) (*sm.CompEthInterfaceV2, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	getCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	didUpdate, err := t.UpdateCompEthInterfaceTx(getCEI, ceip)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	updCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		if didUpdate != true {
			// No update because there was no entry
			t.Rollback()
			return nil, nil
		} else {
			t.Rollback()
			return nil, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, err
	}
	return updCEI, nil
}

// Update existing CompEthInterface entry in the database, but only updates
// fields that would be changed by a user-directed operation.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
//
// Special handling is required to use the V1 API Patch on a V2 CompEthInterface.
// If the CEI has more than 2 or more IP addresses associated with it the error
// CompEthInterfacePatch will be ErrHMSDSCompEthInterfaceMultipleIPs returned.
func (d *hmsdbPg) UpdateCompEthInterfaceV1(id string, ceip *sm.CompEthInterfacePatch) (*sm.CompEthInterfaceV2, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	getCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		t.Rollback()
		return nil, err
	}

	// Enforce backward compability rules. We will only allow an Update on a CEI that has
	// 0 or 1 IP Addresses associated with it. When a Update is done on a CEI with muliple
	// IPs it is ambigous which IP Address is being requested to be updated
	ceipV2 := new(sm.CompEthInterfaceV2Patch)
	ceipV2.CompID = ceip.CompID
	ceipV2.Desc = ceip.Desc

	if ceip.IPAddr != nil {
		ipAddrPatch := *ceip.IPAddr
		if len(getCEI.IPAddrs) > 1 {
			t.Rollback()
			return nil, ErrHMSDSCompEthInterfaceMultipleIPs
		} else if len(getCEI.IPAddrs) == 1 {
			// Update the existing IP Address, and perserve network name if present
			network := getCEI.IPAddrs[0].Network
			ceipV2.IPAddrs = &[]sm.IPAddressMapping{{
				IPAddr:  ipAddrPatch,
				Network: network,
			}}
		} else {
			// Add new IP Address to the CEI
			ceipV2.IPAddrs = &[]sm.IPAddressMapping{{IPAddr: ipAddrPatch}}
		}
	}

	didUpdate, err := t.UpdateCompEthInterfaceTx(getCEI, ceipV2)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	updCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		if didUpdate != true {
			// No update because there was no entry
			t.Rollback()
			return nil, nil
		} else {
			t.Rollback()
			return nil, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, err
	}
	return updCEI, nil
}

// Delete CompEthInterface with matching id from the database, if it
// exists.
// Return true if there was a row affected, false if there were zero.
func (d *hmsdbPg) DeleteCompEthInterfaceByID(id string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteCompEthInterfaceByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all CompEthInterfaces from the database.
// Also returns number of deleted rows, if error is nil.
func (d *hmsdbPg) DeleteCompEthInterfacesAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDeleted, err := t.DeleteCompEthInterfacesAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	if err != nil {
		return 0, err
	}
	return numDeleted, nil
}

// Add IP Address mapping to the existing component ethernet interface.
// returns:
//   - ErrHMSDSNoCompEthInterface if the parent component ethernet interface
//   - ErrHMSDSDuplicateKey if the parent component ethernet interface already
//     has that IP address
//
// Returns key of new IP Address Mapping id, should be the IP address
func (d *hmsdbPg) AddCompEthInterfaceIPAddress(id string, ipmIn *sm.IPAddressMapping) (string, error) {
	t, err := d.Begin()
	if err != nil {
		return "", err
	}
	getCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		t.Rollback()
		return "", err
	}

	if getCEI == nil {
		t.Rollback()
		return "", ErrHMSDSNoCompEthInterface
	}

	// Check that this IP address is not a duplicate
	for _, ipm := range getCEI.IPAddrs {
		if ipm.IPAddr == ipmIn.IPAddr {
			t.Rollback()
			return "", ErrHMSDSDuplicateKey
		}
	}

	// Update the JSON Blob
	ipAddrs := append(getCEI.IPAddrs, *ipmIn)

	// Patch the existing CompEthInterface in the database
	ceip := new(sm.CompEthInterfaceV2Patch)
	ceip.IPAddrs = &ipAddrs

	didUpdate, err := t.UpdateCompEthInterfaceTx(getCEI, ceip)
	if err != nil {
		t.Rollback()
		return "", err
	} else if !didUpdate {
		t.Rollback()

		d.LogAlways("Error: AddCompEthInterfaceIPAddress(): failed to patch CompEthInterface!")
		return "", fmt.Errorf("failed to patch CompEthInterface")
	}

	if err := t.Commit(); err != nil {
		return "", nil
	}

	return ipmIn.IPAddr, nil
}

// Update existing IP Address Mapping for a CompEthInterface entry in the database,
// but only updates fields that would be changed by a user-directed operation.
// Returns updated entry or nil/nil if not found.  If an error occurred,
// nil/error will be returned.
func (d *hmsdbPg) UpdateCompEthInterfaceIPAddress(id, ipAddr string, ipmPatch *sm.IPAddressMappingPatch) (*sm.IPAddressMapping, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}

	getCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		t.Rollback()
		return nil, err
	}

	// Update the JSON Blob only if the Network field is provided
	ipAddrs := []sm.IPAddressMapping{}
	var ipmIndex int
	if ipmPatch.Network != nil {
		fmt.Println("Checking IPs to do the patch", id)
		found := false
		for i, ipm := range getCEI.IPAddrs {
			if ipm.IPAddr == ipAddr {
				// Found the matching IP address mapping
				ipm.Network = *ipmPatch.Network
				found = true
				ipmIndex = i
			}

			ipAddrs = append(ipAddrs, ipm)
		}

		if !found {
			t.Rollback()
			return nil, nil
		}
	}

	// Patch the existing CompEthInterface in the database
	ceip := new(sm.CompEthInterfaceV2Patch)
	ceip.IPAddrs = &ipAddrs

	didUpdate, err := t.UpdateCompEthInterfaceTx(getCEI, ceip)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	updCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		if didUpdate != true {
			// No update because there was no entry
			t.Rollback()
			return nil, nil
		} else {
			t.Rollback()
			return nil, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, err
	}

	// Extract the IP Mapping that was updated
	updIPM := updCEI.IPAddrs[ipmIndex]
	return &updIPM, nil
}

// Delete IP Address mapping from the Component Ethernet Interface.
// If no error, bool indicates whether the IP Address Mapping was present to remove.
func (d *hmsdbPg) DeleteCompEthInterfaceIPAddress(id, ipAddr string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}

	getCEI, err := t.GetCompEthInterfaceByIDTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}

	// Remove the IP Address from the JSON Blob
	found := false
	ipAddrs := []sm.IPAddressMapping{}
	for _, ipm := range getCEI.IPAddrs {
		if ipm.IPAddr == ipAddr {
			found = true
			continue
		}

		ipAddrs = append(ipAddrs, ipm)
	}

	// The IP Adress was not present, no reason to update the record in the database
	if !found {
		t.Rollback()
		return false, nil
	}

	// Patch the existing component interface
	ceip := &sm.CompEthInterfaceV2Patch{
		IPAddrs: &ipAddrs,
	}

	didUpdate, err := t.UpdateCompEthInterfaceTx(getCEI, ceip)
	if err != nil {
		t.Rollback()
		return false, err
	} else if !didUpdate {
		t.Rollback()

		d.LogAlways("Error: DeleteCompEthInterfaceIPAddress(): failed to patch CompEthInterface!")
		return false, fmt.Errorf("failed to patch CompEthInterface")
	}

	err = t.Commit()
	return true, err
}

/////////////////////////////////////////////////////////////////////////////
//
// DiscoveryStatus - Discovery status tracking
//
/////////////////////////////////////////////////////////////////////////////

// Get DiscoveryStatus with the given numerical ID.
func (d *hmsdbPg) GetDiscoveryStatusByID(id uint) (*sm.DiscoveryStatus, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	stat, err := t.GetDiscoveryStatusByIDTx(id)
	if err != nil {
		t.Rollback()
		return stat, err
	}
	t.Commit()
	return stat, err
}

// Get all DiscoveryStatus entries.
func (d *hmsdbPg) GetDiscoveryStatusAll() ([]*sm.DiscoveryStatus, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	stats, err := t.GetDiscoveryStatusAllTx()
	if err != nil {
		t.Rollback()
		return stats, err
	}
	t.Commit()
	return stats, err
}

// Update discovery status in DB.
func (d *hmsdbPg) UpsertDiscoveryStatus(stat *sm.DiscoveryStatus) error {
	return nil
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
func (d *hmsdbPg) UpdateAllForRFEndpoint(
	ep *sm.RedfishEndpoint,
	ceps *sm.ComponentEndpointArray,
	hls []*sm.HWInvByLoc,
	comps *base.ComponentArray,
	seps *sm.ServiceEndpointArray,
	ceis []*sm.CompEthInterfaceV2,
) (*[]base.Component, error) {

	discoveredIDs := make([]base.Component, 0, len(comps.Components))

	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	// Do update of RedfishEndpoint
	didUpdate, err := t.UpdateRFEndpointTx(ep)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	// Make sure entry exists or the other stores will fail.
	_, err = t.GetRFEndpointByIDTx(ep.ID)
	if err != nil {
		if didUpdate != true {
			// No update because there was no entry
			t.Rollback()
			return nil, ErrHMSDSArgNoMatch
		} else {
			t.Rollback()
			return nil, err
		}
	}
	// Upsert ComponentEndpointArray into database
	if ceps != nil {
		err = t.UpsertCompEndpointsTx(ceps)
		if err != nil {
			t.Rollback()
			return nil, err
		}
	}
	// Insert FRUs first because the location info links to them.
	if hls != nil {
		hfs := make([]*sm.HWInvByFRU, 0, len(hls))
		for _, hl := range hls {
			if hl.PopulatedFRU != nil {
				hfs = append(hfs, hl.PopulatedFRU)
			}
		}
		err = t.BulkInsertHWInvByFRUTx(hfs)
		if err != nil {
			t.Rollback()
			return nil, err
		}
		// Now insert HWInvByLocation so that the FRU link will exist.
		err = t.BulkInsertHWInvByLocTx(hls)
		if err != nil {
			t.Rollback()
			return nil, err
		}
	}
	// Inserts or updates HMS Components entries
	if comps != nil {
		compMap := make(map[string]*base.Component)
		nodeList := make([]string, 0, 1)
		for _, comp := range comps.Components {
			compMap[comp.ID] = comp
			if comp.Type == xnametypes.Node.String() {
				nodeList = append(nodeList, comp.ID)
			} else {
				d.Log(LOG_INFO,
					"Not node: %s, type %s, new state is %s/%s",
					comp.ID, comp.Type,
					comp.State, comp.Flag,
				)
			}
		}
		if len(nodeList) > 0 {
			nodes, err := t.GetComponentsTx(IDs(nodeList), From("UpdateAllForRFEndpoint"))
			if err != nil {
				t.Rollback()
				return nil, err
			}
			for _, compOld := range nodes {
				compNew, ok := compMap[compOld.ID]
				if !ok {
					continue
				}
				if base.VerifyNormalizeState(compNew.State) ==
					base.StateOn.String() &&
					base.IsPostBootState(compOld.State) &&
					compNew.Flag == base.FlagOK.String() {
					//
					// Keep higher states if ON is Redfish state.
					// since that is the highest one Redfish will report.
					//
					d.Log(LOG_INFO,
						"Keeping old state for %s, old: %s/%s, new: %s/%s",
						compNew.ID,
						compOld.State, compOld.Flag,
						compNew.State, compNew.Flag,
					)
					compNew.State = compOld.State
					compNew.Flag = compOld.Flag
				} else {
					d.Log(LOG_INFO,
						"Updating state for %s, old: %s/%s, new: %s/%s",
						compNew.ID,
						compOld.State, compOld.Flag,
						compNew.State, compNew.Flag,
					)
				}
			}
		}
		rowsAffected, err := t.InsertComponentsTx(comps.Components)
		if err != nil {
			t.Rollback()
			return nil, err
		}
		for _, id := range rowsAffected {
			if comp, ok := compMap[id]; ok {
				discoveredIDs = append(discoveredIDs, *comp)
			}
		}
	}
	// Upsert ServiceEndpointArray into database
	if seps != nil {
		err = t.UpsertServiceEndpointsTx(seps)
		if err != nil {
			t.Rollback()
			return nil, err
		}
	}
	// Insert CompEthInterfaces into the database
	if ceis != nil {
		err = t.InsertCompEthInterfacesCompInfoTx(ceis)
		if err != nil {
			t.Rollback()
			return nil, err
		}
	}
	if err := t.Commit(); err != nil {
		return nil, err
	}
	return &discoveredIDs, nil
}

////////////////////////////////////////////////////////////////////////////
//
// SCN Subscription Operations
//
////////////////////////////////////////////////////////////////////////////

// Get all SCN subscriptions
func (d *hmsdbPg) GetSCNSubscriptionsAll() (*sm.SCNSubscriptionArray, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	subs, err := t.GetSCNSubscriptionsAllTx()
	if err != nil {
		t.Rollback()
		return subs, err
	}
	err = t.Commit()
	return subs, err
}

// Get a SCN subscription
func (d *hmsdbPg) GetSCNSubscription(id int64) (*sm.SCNSubscription, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	sub, err := t.GetSCNSubscriptionTx(id)
	if err != nil {
		t.Rollback()
		return sub, err
	}
	err = t.Commit()
	return sub, err
}

// Insert a new SCN subscription. Existing subscriptions are unaffected
func (d *hmsdbPg) InsertSCNSubscription(sub sm.SCNPostSubscription) (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	id, err := t.InsertSCNSubscriptionTx(sub)
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	return id, err
}

// Update an existing SCN subscription.
func (d *hmsdbPg) UpdateSCNSubscription(id int64, sub sm.SCNPostSubscription) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didUpdate, err := t.UpdateSCNSubscriptionTx(id, sub)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didUpdate, err
}

// Patch an existing SCN subscription.
func (d *hmsdbPg) PatchSCNSubscription(id int64, op string, patch sm.SCNPatchSubscription) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didPatch, err := t.PatchSCNSubscriptionTx(id, op, patch)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didPatch, err
}

// Delete a SCN subscription
func (d *hmsdbPg) DeleteSCNSubscription(id int64) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didDelete, err := t.DeleteSCNSubscriptionTx(id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Delete all SCN subscriptions
func (d *hmsdbPg) DeleteSCNSubscriptionsAll() (int64, error) {
	t, err := d.Begin()
	if err != nil {
		return 0, err
	}
	numDelete, err := t.DeleteSCNSubscriptionsAllTx()
	if err != nil {
		t.Rollback()
		return 0, err
	}
	err = t.Commit()
	return numDelete, err
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
// nil error. Will return ErrHMSDSDuplicateKey if group exits or is
// exclusive and xname id is already in another group in this exclusive set.
// In addition, returns ErrHMSDSNoComponent if a component id doesn't exist.
func (d *hmsdbPg) InsertGroup(g *sm.Group) (string, error) {
	t, err := d.Begin()
	if err != nil {
		return "", err
	}
	// Insert first the group, with no members.
	// Note this also normalizes and verifies data - exgroup won't contain '%'
	uuid, label, exgrp, err := t.InsertEmptyGroupTx(g)
	if err != nil {
		t.Rollback()
		return "", err
	}
	namespace := label // Normal namespace is non-exclusive group name
	if exgrp != "" {
		// exclusive group - uniquified exclusive group as namespace
		namespace = "%" + exgrp + "%"
	}
	err = t.InsertMembersTx(uuid, namespace, &g.Members)
	if err != nil {
		t.Rollback()
		return "", err
	}
	err = t.Commit()
	return label, err
}

// Update group with label
func (d *hmsdbPg) UpdateGroup(label string, gp *sm.GroupPatch) error {
	gp.Normalize()
	if err := gp.Verify(); err != nil {
		return err
	}
	// Start the transaction
	t, err := d.Begin()
	if err != nil {
		return err
	}
	// Get the existing partition in a transaction, without members initially.
	uuid, g, err := t.GetEmptyGroupTx(label)
	if err != nil {
		// Unexpected error - couldn't get partition
		t.Rollback()
		return err
	} else if g == nil || uuid == "" {
		// Lookup returned nothing - 404
		t.Rollback()
		return ErrHMSDSNoGroup
	}
	if err := t.UpdateEmptyGroupTx(uuid, g, gp); err != nil {
		t.Rollback()
		return err
	}
	return t.Commit()
}

// Get Group with given label.  Nil if not found and nil error, otherwise
// nil plus non-nil error (not normally expected)
// If filt_part is non-empty, the partition name is used to filter
// the members list.
func (d *hmsdbPg) GetGroup(label, filt_part string) (*sm.Group, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	uuid, g, err := t.GetEmptyGroupTx(label)
	if err != nil {
		t.Rollback()
		return nil, err
	}
	not_uuid := ""
	null_part := false
	if filt_part == "NULL" {
		// Don't need to look up and_part, we want group members w NO partition
		null_part = true
	} else if filt_part != "" {
		not_uuid, _, err = t.GetEmptyPartitionTx(filt_part)
		if err != nil {
			t.Rollback()
			return nil, err
		} else if not_uuid == "" {
			t.Rollback()
			return nil, ErrHMSDSNoPartition
		}
	}
	// Get Members
	if g != nil && uuid != "" {
		if not_uuid == "" && null_part == false {
			// Just get group members
			ms, err := t.GetMembersTx(uuid)
			if err != nil {
				t.Rollback()
				return nil, err
			}
			g.Members.IDs = ms.IDs
		} else {
			// Filtering group members by part, and filt_part is valid or
			// NULL
			ms, err := t.GetMembersFilterTx(uuid, not_uuid)
			if err != nil {
				t.Rollback()
				return nil, err
			}
			g.Members.IDs = ms.IDs
		}
	}
	t.Commit()
	return g, err
}

// Get list of group labels (names).
func (d *hmsdbPg) GetGroupLabels() ([]string, error) {
	query := sq.Select("name").
		From(compGroupsTable).
		Where("namespace = ?", groupNamespace)

	// Query with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(d.sc).QueryContext(d.ctx)
	if err != nil {
		d.LogAlways("Error: GetGroupLabels(): query failed: %s", err)
		return []string{}, err
	}
	defer rows.Close()

	labels := []string{}
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			d.LogAlways("Error: GetGroupLabels(): scan failed: %s", err)
			return []string{}, err
		}
		labels = append(labels, label)
	}
	return labels, nil
}

// Delete entire group with the given label.  If no error, bool indicates
// whether member was present to remove.
func (d *hmsdbPg) DeleteGroup(label string) (bool, error) {
	// Build query
	query := sq.Delete(compGroupsTable).
		Where("name = ?", sm.NormalizeGroupField(label)).
		Where("namespace = ?", groupNamespace)

	// Execute delete query
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(d.sc).ExecContext(d.ctx)
	if err != nil {
		return false, err
	}
	// See if the rows were affected.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else {
		if num > 0 {
			if num > 1 {
				d.LogAlways("Error: DeleteGroup(): multiple deletions!")
			}
			return true, nil
		}
	}
	return false, nil
}

// Add member xname id to existing group label.  returns ErrHMSDSNoGroup
// if group with label does not exist, or ErrHMSDSDuplicateKey if Group
// is exclusive and xname id is already in another group in this exclusive set.
// In addition, returns ErrHMSDSNoComponent if the component doesn't exist.
//
// Returns key of new member id, should be same as id after normalization,
// if any.  Label should already be normalized.
func (d *hmsdbPg) AddGroupMember(label, id string) (string, error) {
	// Prep id for insertion and verify it.
	ms := new(sm.Members)
	ms.IDs = append(ms.IDs, id)
	ms.Normalize()
	if err := ms.Verify(); err != nil {
		return "", err
	}
	// Start transaction,
	t, err := d.Begin()
	if err != nil {
		return "", err
	}
	// First we need to look up the group, if it exists.
	uuid, g, err := t.GetEmptyGroupTx(label)
	if err != nil {
		t.Rollback()
		return "", err
	} else if g == nil || uuid == "" {
		// Group does not exist
		t.Rollback()
		return "", ErrHMSDSNoGroup
	}
	// Default namespace is non-exclusive group name
	namespace := g.Label
	if g.ExclusiveGroup != "" {
		// exclusive group - uniquified exclusive group as namespace
		namespace = "%" + g.ExclusiveGroup + "%"
	}
	// Do the actual insertion now that we have the group and and namespace.
	err = t.InsertMembersTx(uuid, namespace, ms)
	if err != nil {
		t.Rollback()
		return "", err
	}
	err = t.Commit()
	return ms.IDs[0], err
}

// Delete Group member from label.  If no error, bool indicates whether
// group was present to remove.
func (d *hmsdbPg) DeleteGroupMember(label, id string) (bool, error) {
	// Start transaction, first we need to look up the group, if it exists.
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	uuid, g, err := t.GetEmptyGroupTx(label)
	if err != nil {
		t.Rollback()
		return false, err
	} else if g == nil || uuid == "" {
		// Group does not exist
		t.Rollback()
		return false, ErrHMSDSNoGroup
	}
	didDelete, err := t.DeleteMemberTx(uuid, id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

// Sets membership list of group label to ids. If any xnames in ids already
// exist in group, they stay in the group. If any xnames in ids do not already
// exist in group, they are added. If any xnames that already exist in group are
// not in ids, they are removed. An error is returned if one occurs, otherwise
// nil.
func (d *hmsdbPg) SetGroupMembers(label string, ids []string) ([]string, error) {
	// Prepare membership data and verify member xnames
	ms := new(sm.Members)
	ms.IDs = ids
	ms.Normalize()
	if err := ms.Verify(); err != nil {
		return []string{}, fmt.Errorf("failed to verify member xnames: %w", err)
	}

	// Start transaction
	t, err := d.Begin()
	if err != nil {
		return []string{}, fmt.Errorf("failed to begin postgres transaction: %w", err)
	}

	// Group needs to exist for this to work
	uuid, g, err := t.GetEmptyGroupTx(label)
	if err != nil {
		// A different error occurred
		return []string{}, fmt.Errorf("failed to get group %q: %w", label, err)
	} else if g == nil || uuid == "" {
		// Group does not exist
		t.Rollback()
		return []string{}, ErrHMSDSNoGroup
	}

	// Determine namespace
	//
	// Default is non-exclusive group name
	namespace := g.Label
	if g.ExclusiveGroup != "" {
		// exclusive group - uniquified exclusive group as namespace
		namespace = "%" + g.ExclusiveGroup + "%"
	}

	// Delete old groups
	if _, err := t.DeleteMembersAllTx(uuid); err != nil {
		t.Rollback()
		return []string{}, fmt.Errorf("failed to delete old groups from %q: %w", label, err)
	}

	// Add new groups
	err = t.InsertMembersTx(uuid, namespace, ms)
	if err != nil {
		t.Rollback()
		return []string{}, fmt.Errorf("failed to add new members %v to group %q: %w", ids, label, err)
	}

	// Commit changes
	err = t.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit changes: %w", err)
	}

	return ms.IDs, err
}

//
// Partitions
//

// Create a partition.  Returns new name (should match one in struct,
// unless case-normalized) if successful, otherwise empty string + non
// nil error.  Will return ErrHMSDSDuplicateKey if partition exits or an
// xname id already exists in another partition.
// In addition, returns ErrHMSDSNoComponent if a component doesn't exist.
func (d *hmsdbPg) InsertPartition(p *sm.Partition) (string, error) {
	t, err := d.Begin()
	if err != nil {
		return "", err
	}
	// Insert first the partition, with no members, after
	// verifying/normalizing.
	uuid, pname, err := t.InsertEmptyPartitionTx(p)
	if err != nil {
		t.Rollback()
		return "", err
	}
	// special unique namespace for partitions - can't clash with due to
	// normally disallowed '%' characters.  These were checked in the last
	// call.
	namespace := partGroupNamespace
	err = t.InsertMembersTx(uuid, namespace, &p.Members)
	if err != nil {
		t.Rollback()
		return "", err
	}
	err = t.Commit()
	return pname, err
}

// Update Partition with given name
func (d *hmsdbPg) UpdatePartition(pname string, pp *sm.PartitionPatch) error {
	// Check input before starting any DB actions
	pp.Normalize()
	if err := pp.Verify(); err != nil {
		return err
	}
	// Start the transaction
	t, err := d.Begin()
	if err != nil {
		return err
	}
	// Get the existing partition in a transaction, without members initially.
	uuid, p, err := t.GetEmptyPartitionTx(pname)
	if err != nil {
		// Unexpected error - couldn't get partition
		t.Rollback()
		return err
	} else if p == nil || uuid == "" {
		// Lookup returned nothing - 404
		t.Rollback()
		return ErrHMSDSNoPartition
	}
	if err := t.UpdateEmptyPartitionTx(uuid, p, pp); err != nil {
		t.Rollback()
		return err
	}
	return t.Commit()
}

// Get partition with given name  Nil if not found and nil error, otherwise
// nil plus non-nil error (not normally expected)
func (d *hmsdbPg) GetPartition(pname string) (*sm.Partition, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	uuid, p, err := t.GetEmptyPartitionTx(pname)
	if err != nil {
		t.Rollback()
		return nil, err
	} else if p != nil && uuid != "" {
		ms, err := t.GetMembersTx(uuid)
		if err != nil {
			t.Rollback()
			return nil, err
		}
		p.Members.IDs = ms.IDs
	}
	t.Commit()
	return p, err
}

// Get list of partition names.
func (d *hmsdbPg) GetPartitionNames() ([]string, error) {
	query := sq.Select("name").
		From(compGroupsTable).
		Where("namespace = ?", partNamespace)

	// Query with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(d.sc).QueryContext(d.ctx)
	if err != nil {
		d.LogAlways("Error: GetPartitionNames(): query failed: %s", err)
		return []string{}, err
	}
	defer rows.Close()

	pnames := []string{}
	for rows.Next() {
		var pname string
		if err := rows.Scan(&pname); err != nil {
			d.LogAlways("Error: GetPartitionNames(): scan failed: %s", err)
			return []string{}, err
		}
		pnames = append(pnames, pname)
	}
	return pnames, nil
}

// Delete entire partition with pname.  If no error, bool indicates
// whether partition was present to remove.
func (d *hmsdbPg) DeletePartition(pname string) (bool, error) {
	// Build query
	query := sq.Delete(compGroupsTable).
		Where("name = ?", sm.NormalizeGroupField(pname)).
		Where("namespace = ?", partNamespace)

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(d.sc).ExecContext(d.ctx)
	if err != nil {
		return false, err
	}
	// See if any rows were affected
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else {
		if num > 0 {
			if num > 1 {
				d.LogAlways("Error: DeletePartition(): multiple deletions!")
			}
			return true, nil
		}
	}
	return false, nil
}

// Add member xname id to existing partition.  returns ErrHMSDSNoGroup
// if partition name does not exist, or ErrHMSDSDuplicateKey if xname id
// is already in a different partition.
// Returns key of new member, should be same as id after normalization,
// if any.  pname should already be normalized.
func (d *hmsdbPg) AddPartitionMember(pname, id string) (string, error) {
	// Prep id for insertion and verify it.
	ms := new(sm.Members)
	ms.IDs = append(ms.IDs, id)
	ms.Normalize()
	if err := ms.Verify(); err != nil {
		return "", err
	}
	// Start transaction, first we need to look up the part name, if it exists.
	t, err := d.Begin()
	if err != nil {
		return "", err
	}
	uuid, p, err := t.GetEmptyPartitionTx(pname)
	if err != nil {
		t.Rollback()
		return "", err
	} else if p == nil || uuid == "" {
		// Partition does not exist
		t.Rollback()
		return "", ErrHMSDSNoPartition
	}
	// special unique namespace for partitions - can't clash with due to
	// normally disallowed '%' characters.  These were checked in the last
	// call.
	namespace := partGroupNamespace
	err = t.InsertMembersTx(uuid, namespace, ms)
	if err != nil {
		t.Rollback()
		return "", err
	}
	err = t.Commit()
	return ms.IDs[0], err
}

// Delete partition member from partition.  If no error, bool indicates
// whether member was present to remove.
func (d *hmsdbPg) DeletePartitionMember(pname, id string) (bool, error) {
	// Start transaction, first we need to look up the group, if it exists.
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	uuid, p, err := t.GetEmptyPartitionTx(pname)
	if err != nil {
		t.Rollback()
		return false, err
	} else if p == nil || uuid == "" {
		// Partition does not exist
		t.Rollback()
		return false, ErrHMSDSNoPartition
	}
	didDelete, err := t.DeleteMemberTx(uuid, id)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didDelete, err
}

//
// Memberships
//

// Get the memberships for a particular component xname id
func (d *hmsdbPg) GetMembership(id string) (*sm.Membership, error) {
	f := new(ComponentFilter)
	f.ID = []string{id}
	f.label = "GetMembership"
	mbs, err := d.GetMemberships(f)
	if err != nil {
		return nil, err
	}
	if mbs != nil && len(mbs) != 0 {
		return mbs[0], nil
	}
	return nil, nil
}

// Get all memberships, optionally filtering
// Convenience feature - not needed for initial implementation
func (d *hmsdbPg) GetMemberships(f *ComponentFilter) ([]*sm.Membership, error) {

	fname := "GetMemberships"
	if f != nil && f.label == "" {
		f.label = fname
	}
	query, err := selectComponents(f, FLTR_ID_W_GROUP)
	if err != nil {
		d.LogAlways("Error: %s(): makeComponentQuery failed: %s", fname, err)
		return []*sm.Membership{}, err
	}
	query = query.PlaceholderFormat(sq.Dollar)
	queryString, args, _ := query.ToSql()
	d.Log(LOG_DEBUG, "%s: Submitting '%s' with '%v'", fname, queryString, args)

	rows, err := query.RunWith(d.sc).QueryContext(d.ctx)
	if err != nil {
		d.LogAlways("Error: %s(): query exec failed: %s", fname, err)
		return []*sm.Membership{}, err
	}
	defer rows.Close()

	// We need to consolodate rows into one object per xname id
	lookup := make(map[string]*sm.Membership)

	var name, namespace *string
	var id string = ""
	var mb *sm.Membership
	var ok bool
	for rows.Next() {
		if err := rows.Scan(&id, &name, &namespace); err != nil {
			d.LogAlways("Error: %s(): scan failed: %s", fname, err)
			return []*sm.Membership{}, err
		}
		mb, ok = lookup[id]
		if !ok {
			mb = new(sm.Membership)
			mb.ID = id
			mb.GroupLabels = []string{}
			lookup[id] = mb
		}
		if namespace != nil {
			if *namespace == groupNamespace {
				if name != nil {
					mb.GroupLabels = append(mb.GroupLabels, *name)
				} else {
					d.LogAlways("Warning: %s(): nil group name for id=%s,ns=%s",
						fname, id, *namespace)
				}
			} else if *namespace == partNamespace {
				if name != nil {
					mb.PartitionName = *name
				} else {
					d.LogAlways("Warning: %s(): nil pname for id=%s,n s=%s",
						fname, id, *namespace)
				}
			}
		} else if name != nil {
			d.LogAlways("Warning: %s(): nil namespace for id=%s, name=%s",
				fname, id, *name)
		}
	}
	mbs := make([]*sm.Membership, 0, len(lookup))
	for _, m := range lookup {
		mbs = append(mbs, m)
	}
	return mbs, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Component Lock Management
//
////////////////////////////////////////////////////////////////////////////

//
// Component Locks V2
//

func compLockFilterToCompFilter(clf sm.CompLockV2Filter) (cf ComponentFilter) {
	cf.ID = clf.ID
	cf.NID = clf.NID
	cf.Type = clf.Type
	cf.State = clf.State
	cf.Flag = clf.Flag
	cf.Enabled = clf.Enabled
	cf.SwStatus = clf.SwStatus
	cf.Role = clf.Role
	cf.SubRole = clf.SubRole
	cf.Subtype = clf.Subtype
	cf.Arch = clf.Arch
	cf.Class = clf.Class
	cf.Group = clf.Group
	cf.Partition = clf.Partition
	cf.ReservationDisabled = clf.ReservationDisabled
	cf.Locked = clf.Locked
	return cf
}

// Create component reservations if one doesn't already exist.
// To create reservations without a duration, the component must be locked.
// To create reservations with a duration, the component must be unlocked.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func insertCompReservationsHelper(t HMSDBTx, f sm.CompLockV2Filter) (sm.CompLockV2ReservationResult, error) {
	var result sm.CompLockV2ReservationResult
	var rigid bool
	insertComps := make([]string, 0, 1)
	result.Success = make([]sm.CompLockV2Success, 0, 1)
	result.Failure = make([]sm.CompLockV2Failure, 0, 1)

	if f.ProcessingModel == sm.CLProcessingModelRigid {
		rigid = true
	}

	cf := compLockFilterToCompFilter(f)
	cf.writeLock = true
	cf.label = "InsertCompReservations"
	affectedComps, err := t.GetComponentsFilterTx(&cf, FLTR_DEFAULT)
	if err != nil {
		return result, err
	}
	if len(affectedComps) == 0 {
		return result, sm.ErrCompLockV2NotFound
	}
	// Insert reservations
	for _, comp := range affectedComps {
		lockErr := sm.CLResultSuccess
		if comp.ReservationDisabled {
			// Can't create reservations when reservations are disabled
			lockErr = sm.CLResultDisabled
			err = sm.ErrCompLockV2CompDisabled
		} else if f.ReservationDuration == 0 && !comp.Locked {
			// Can't create non-expiring reservations while the component is unlocked.
			lockErr = sm.CLResultUnlocked
			err = sm.ErrCompLockV2CompUnlocked
		} else if f.ReservationDuration != 0 && comp.Locked {
			// Can't create expiring reservations while the component is locked.
			lockErr = sm.CLResultLocked
			err = sm.ErrCompLockV2CompLocked
		}
		if lockErr != sm.CLResultSuccess {
			if rigid {
				return result, err
			}
			fail := sm.CompLockV2Failure{
				ID:     comp.ID,
				Reason: lockErr,
			}
			result.Failure = append(result.Failure, fail)
			continue
		}
		// Build a list of components that we will attempt to reserve
		insertComps = append(insertComps, comp.ID)
	}
	// This can really only happen if f.ProcessingModel == sm.CLProcessingModelFlexible
	if len(insertComps) == 0 {
		return result, nil
	}
	reservations, lockErr, err := t.InsertCompReservationsTx(insertComps, f.ReservationDuration)
	if err != nil {
		return result, err
	} else if lockErr != sm.CLResultSuccess {
		// The component is already reserved
		if rigid {
			return result, sm.ErrCompLockV2CompReserved
		}
		for _, id := range insertComps {
			fail := sm.CompLockV2Failure{
				ID:     id,
				Reason: lockErr,
			}
			result.Failure = append(result.Failure, fail)
		}
	}

	reservationMap := make(map[string]bool)
	// Add our successful inserts into the results
	for _, reservation := range reservations {
		reservationMap[reservation.ID] = true
		result.Success = append(result.Success, reservation)
	}
	if len(insertComps) != len(reservations) {
		// If we're flexible, generate a list of inserts that didn't happen
		for _, id := range insertComps {
			if _, ok := reservationMap[id]; !ok {
				fail := sm.CompLockV2Failure{
					ID:     id,
					Reason: sm.CLResultReserved,
				}
				result.Failure = append(result.Failure, fail)
			}
		}
		if rigid {
			return result, sm.ErrCompLockV2CompReserved
		}
	}
	return result, nil
}

// Create component reservations if one doesn't already exist.
// To create reservations without a duration, the component must be locked.
// To create reservations with a duration, the component must be unlocked.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbPg) InsertCompReservations(f sm.CompLockV2Filter) (sm.CompLockV2ReservationResult, error) {
	t, err := d.Begin()
	if err != nil {
		return sm.CompLockV2ReservationResult{}, err
	}

	result, err := insertCompReservationsHelper(t, f)
	if err != nil {
		t.Rollback()
		return result, err
	}

	err = t.Commit()
	return result, err
}

// Remove/Release component reservations.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
// Force = true allows reservations to be removed without the reservation key.
func deleteCompReservationsHelper(t HMSDBTx, f sm.CompLockV2ReservationFilter, force bool) (sm.CompLockV2UpdateResult, error) {
	var result sm.CompLockV2UpdateResult
	result.Success.ComponentIDs = make([]string, 0, 1)
	result.Failure = make([]sm.CompLockV2Failure, 0, 1)

	locks, err := t.DeleteCompReservationsTx(f.ReservationKeys, force)
	if err != nil {
		if f.ProcessingModel == sm.CLProcessingModelRigid {
			return result, err
		} else {
			for _, key := range f.ReservationKeys {
				fail := sm.CompLockV2Failure{
					ID:     key.ID,
					Reason: sm.CLResultServerError,
				}
				result.Failure = append(result.Failure, fail)
			}
		}
	} else if len(f.ReservationKeys) != len(locks) {
		// Component reservation does not exist
		if f.ProcessingModel == sm.CLProcessingModelRigid {
			return result, sm.ErrCompLockV2NotFound
		}
		lockMap := make(map[string]bool)
		for _, lock := range locks {
			lockMap[lock] = true
		}
		for _, key := range f.ReservationKeys {
			if _, ok := lockMap[key.ID]; !ok {
				fail := sm.CompLockV2Failure{
					ID:     key.ID,
					Reason: sm.CLResultNotFound,
				}
				result.Failure = append(result.Failure, fail)
			} else {
				result.Success.ComponentIDs = append(result.Success.ComponentIDs, key.ID)
			}
		}
	} else {
		result.Success.ComponentIDs = append(result.Success.ComponentIDs, locks...)
	}

	// Do the counts
	result.Counts.Success = len(result.Success.ComponentIDs)
	result.Counts.Failure = len(result.Failure)
	result.Counts.Total = result.Counts.Success + result.Counts.Failure

	return result, nil
}

// Forcebly remove/release component reservations.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbPg) DeleteCompReservationsForce(f sm.CompLockV2Filter) (sm.CompLockV2UpdateResult, error) {
	var resFilter sm.CompLockV2ReservationFilter

	// Start transaction, first we need to look up the group, if it exists.
	t, err := d.Begin()
	if err != nil {
		return sm.CompLockV2UpdateResult{}, err
	}

	cf := compLockFilterToCompFilter(f)
	affectedComps, err := t.GetComponentsFilterTx(&cf, FLTR_DEFAULT)
	if err != nil {
		t.Rollback()
		return sm.CompLockV2UpdateResult{}, err
	}
	if len(affectedComps) == 0 {
		t.Rollback()
		return sm.CompLockV2UpdateResult{}, sm.ErrCompLockV2NotFound
	}
	resFilter.ProcessingModel = f.ProcessingModel
	for _, comp := range affectedComps {
		key := sm.CompLockV2Key{ID: comp.ID}
		resFilter.ReservationKeys = append(resFilter.ReservationKeys, key)
	}
	result, err := deleteCompReservationsHelper(t, resFilter, true)
	if err != nil {
		t.Rollback()
		return result, err
	}

	err = t.Commit()
	return result, err
}

// Remove/release component reservations.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbPg) DeleteCompReservations(f sm.CompLockV2ReservationFilter) (sm.CompLockV2UpdateResult, error) {
	// Start transaction, first we need to look up the group, if it exists.
	t, err := d.Begin()
	if err != nil {
		return sm.CompLockV2UpdateResult{}, err
	}

	result, err := deleteCompReservationsHelper(t, f, false)
	if err != nil {
		t.Rollback()
		return result, err
	}

	err = t.Commit()
	return result, err
}

// Release all expired reservations
func (d *hmsdbPg) DeleteCompReservationsExpired() ([]string, error) {
	// Start transaction, first we need to look up the group, if it exists.
	t, err := d.Begin()
	if err != nil {
		return []string{}, err
	}

	xnames, err := t.DeleteCompReservationExpiredTx()
	if err != nil {
		t.Rollback()
		return xnames, err
	}

	err = t.Commit()
	return xnames, err
}

// Retrieve the status of reservations. The public key and xname is
// required to address the reservation.
func (d *hmsdbPg) GetCompReservations(dkeys []sm.CompLockV2Key) (sm.CompLockV2ReservationResult, error) {
	var result sm.CompLockV2ReservationResult
	result.Success = make([]sm.CompLockV2Success, 0, 1)
	result.Failure = make([]sm.CompLockV2Failure, 0, 1)

	t, err := d.Begin()
	if err != nil {
		return result, err
	}
	reservations, lockErr, err := t.GetCompReservationsTx(dkeys, false)
	if err != nil {
		t.Rollback()
		return result, err
	} else if lockErr != sm.CLResultSuccess {
		for _, key := range dkeys {
			fail := sm.CompLockV2Failure{
				ID:     key.ID,
				Reason: lockErr,
			}
			result.Failure = append(result.Failure, fail)
		}
	} else {
		reservationMap := make(map[string]bool)
		for _, reservation := range reservations {
			result.Success = append(result.Success, reservation)
			reservationMap[reservation.ID] = true
		}
		// Report the reservations we didn't find
		if len(reservations) != len(dkeys) {
			for _, key := range dkeys {
				if _, ok := reservationMap[key.ID]; !ok {
					fail := sm.CompLockV2Failure{
						ID:     key.ID,
						Reason: sm.CLResultNotFound,
					}
					result.Failure = append(result.Failure, fail)
				}
			}
		}
	}

	err = t.Commit()
	return result, err
}

// Update/renew the expiration time of component reservations with the given
// ID/Key combinations.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbPg) UpdateCompReservations(f sm.CompLockV2ReservationFilter) (sm.CompLockV2UpdateResult, error) {
	var result sm.CompLockV2UpdateResult
	result.Success.ComponentIDs = make([]string, 0, 1)
	result.Failure = make([]sm.CompLockV2Failure, 0, 1)

	// Start the transaction
	t, err := d.Begin()
	if err != nil {
		return result, err
	}

	// Update the reservations and retrieve any v1LockIDs associated with our reservations.
	locks, err := t.UpdateCompReservationsTx(f.ReservationKeys, f.ReservationDuration, false)
	if err != nil {
		if f.ProcessingModel == sm.CLProcessingModelRigid {
			t.Rollback()
			return result, err
		}
		for _, key := range f.ReservationKeys {
			fail := sm.CompLockV2Failure{
				ID:     key.ID,
				Reason: sm.CLResultServerError,
			}
			result.Failure = append(result.Failure, fail)
		}
	} else if len(locks) != len(f.ReservationKeys) {
		// Component reservation does not exist
		if f.ProcessingModel == sm.CLProcessingModelRigid {
			t.Rollback()
			return result, sm.ErrCompLockV2NotFound
		}
		lockMap := make(map[string]bool)
		for _, lock := range locks {
			lockMap[lock] = true
		}
		for _, key := range f.ReservationKeys {
			if _, ok := lockMap[key.ID]; !ok {
				fail := sm.CompLockV2Failure{
					ID:     key.ID,
					Reason: sm.CLResultNotFound,
				}
				result.Failure = append(result.Failure, fail)
			} else {
				result.Success.ComponentIDs = append(result.Success.ComponentIDs, key.ID)
			}
		}
	} else {
		result.Success.ComponentIDs = append(result.Success.ComponentIDs, locks...)
	}

	// Do the counts
	result.Counts.Success = len(result.Success.ComponentIDs)
	result.Counts.Failure = len(result.Failure)
	result.Counts.Total = result.Counts.Success + result.Counts.Failure

	err = t.Commit()
	return result, err
}

// Retrieve component lock information.
func (d *hmsdbPg) GetCompLocksV2(f sm.CompLockV2Filter) ([]sm.CompLockV2, error) {
	var result []sm.CompLockV2
	var keys []sm.CompLockV2Key

	t, err := d.Begin()
	if err != nil {
		return nil, err
	}

	// Get a list of components based on the given filters
	cf := compLockFilterToCompFilter(f)
	affectedComps, err := t.GetComponentsFilterTx(&cf, FLTR_DEFAULT)
	if err != nil {
		t.Rollback()
		return result, err
	}
	if len(affectedComps) == 0 {
		t.Rollback()
		return result, sm.ErrCompLockV2NotFound
	}

	for _, comp := range affectedComps {
		key := sm.CompLockV2Key{ID: comp.ID}
		keys = append(keys, key)
	}

	// Get any reservations associated with the list of components
	reservations, _, err := t.GetCompReservationsTx(keys, true)
	if err != nil {
		t.Rollback()
		return result, err
	}
	t.Commit()

	reservationMap := make(map[string]sm.CompLockV2Success)
	for _, reservation := range reservations {
		reservationMap[reservation.ID] = reservation
	}

	for _, comp := range affectedComps {
		lock := sm.CompLockV2{
			ID:                  comp.ID,
			Locked:              comp.Locked,
			Reserved:            false,
			ReservationDisabled: comp.ReservationDisabled,
		}
		if reservation, ok := reservationMap[comp.ID]; ok {
			lock.Reserved = true
			lock.CreationTime = reservation.CreationTime
			lock.ExpirationTime = reservation.ExpirationTime
		}
		if f.Reserved != nil {
			reservedParam, err := strconv.ParseBool(f.Reserved[0])
			if err != nil {
				return result, ErrHMSDSArgBadArg
			}
			if reservedParam && lock.Reserved {
				result = append(result, lock)
			} else if !reservedParam && !lock.Reserved {
				result = append(result, lock)
			}
		} else {
			result = append(result, lock)
		}
	}

	if len(result) == 0 {
		return result, sm.ErrCompLockV2NotFound
	}

	return result, nil
}

// Update component locks. Valid actions are 'Lock', 'Unlock', 'Disable',
// and 'Repair'.
// 'Lock'\'Unlock' updates the 'locked' status of the components.
// 'Disable'\'Repair' updates the 'reservationsDisabled' status of components.
// ProcessingModel "rigid" is all or nothing. ProcessingModel "flexible" is
// best try.
func (d *hmsdbPg) UpdateCompLocksV2(f sm.CompLockV2Filter, action string) (sm.CompLockV2UpdateResult, error) {
	var (
		result      sm.CompLockV2UpdateResult
		affectedIds []string
		lockKeys    []sm.CompLockV2Key
	)
	result.Success.ComponentIDs = make([]string, 0, 1)
	result.Failure = make([]sm.CompLockV2Failure, 0, 1)

	t, err := d.Begin()
	if err != nil {
		return result, err
	}

	// Get a list of components to update
	cf := compLockFilterToCompFilter(f)
	affectedComps, err := t.GetComponentsFilterTx(&cf, FLTR_DEFAULT)
	if err != nil {
		t.Rollback()
		return result, err
	}
	if len(affectedComps) == 0 {
		t.Rollback()
		return result, sm.ErrCompLockV2NotFound
	}

	switch action {
	case CLUpdateActionDisable:
		resFilter := sm.CompLockV2ReservationFilter{
			ProcessingModel: f.ProcessingModel,
		}
		for _, comp := range affectedComps {
			key := sm.CompLockV2Key{ID: comp.ID}
			resFilter.ReservationKeys = append(resFilter.ReservationKeys, key)
		}
		// Forcibly release reservations for components
		// we are disabling reservations for.
		_, err = deleteCompReservationsHelper(t, resFilter, true)
		if err != nil && err != sm.ErrCompLockV2NotFound {
			t.Rollback()
			return result, err
		}
		fallthrough
	case CLUpdateActionRepair:
		newVal := (action == CLUpdateActionDisable)
		for _, comp := range affectedComps {
			affectedIds = append(affectedIds, comp.ID)
		}
		// Update the ReservationsDisabled flag on the components
		updatedIds, err := t.BulkUpdateCompResDisabledTx(affectedIds, newVal)
		if err != nil {
			t.Rollback()
			return result, err
		}
		updatedIdMap := make(map[string]bool)
		for _, id := range updatedIds {
			updatedIdMap[id] = true
			result.Success.ComponentIDs = append(result.Success.ComponentIDs, id)
		}
		if len(affectedIds) != len(updatedIds) {
			for _, comp := range affectedComps {
				if _, ok := updatedIdMap[comp.ID]; !ok {
					if comp.ReservationDisabled != newVal {
						// Shouldn't really happen unless somehow the component
						// was deleted between our GET and UPDATE.
						if f.ProcessingModel == sm.CLProcessingModelRigid {
							t.Rollback()
							return result, sm.ErrCompLockV2Unknown
						}
						fail := sm.CompLockV2Failure{
							ID:     comp.ID,
							Reason: sm.CLResultServerError,
						}
						result.Failure = append(result.Failure, fail)
					} else {
						// Add updates that were already disabled/enabled to our successes
						result.Success.ComponentIDs = append(result.Success.ComponentIDs, comp.ID)
					}
				}
			}
		}
	case CLUpdateActionLock:
		fallthrough
	case CLUpdateActionUnlock:
		newVal := (action == "Lock")
		lockErr := sm.CLResultSuccess
		affectedMap := make(map[string]bool)
		for _, comp := range affectedComps {
			lockErr = sm.CLResultSuccess
			// Components can't be (un)locked if reservations are disabled.
			if comp.ReservationDisabled {
				lockErr = sm.CLResultDisabled
				err = sm.ErrCompLockV2CompDisabled
			}
			// Components can't be (un)locked if already (un)locked.
			if comp.Locked == newVal {
				if newVal {
					lockErr = sm.CLResultLocked
					err = sm.ErrCompLockV2CompLocked
				} else {
					lockErr = sm.CLResultUnlocked
					err = sm.ErrCompLockV2CompUnlocked
				}
			}
			if lockErr != sm.CLResultSuccess {
				if f.ProcessingModel == sm.CLProcessingModelRigid {
					t.Rollback()
					return result, err
				}
				fail := sm.CompLockV2Failure{
					ID:     comp.ID,
					Reason: lockErr,
				}
				result.Failure = append(result.Failure, fail)
				continue
			}
			key := sm.CompLockV2Key{ID: comp.ID}
			lockKeys = append(lockKeys, key)
			affectedMap[comp.ID] = true
		}
		// Reset err to clear anything from above for ProcessingModel = Flexible requests
		err = nil
		// Check for reservations. Components can't be
		// (un)locked if there are any reservations.
		reservations, lockErr, err := t.GetCompReservationsTx(lockKeys, true)
		if err != nil {
			t.Rollback()
			return result, err
		} else if len(reservations) > 0 {
			// A reservation was found. Components can't be
			// (un)locked if there are any reservations.
			if f.ProcessingModel == sm.CLProcessingModelRigid {
				t.Rollback()
				return result, sm.ErrCompLockV2CompReserved
			}
			// Remove reserved IDs from our list
			for _, reservation := range reservations {
				fail := sm.CompLockV2Failure{
					ID:     reservation.ID,
					Reason: sm.CLResultReserved,
				}
				result.Failure = append(result.Failure, fail)
				delete(affectedMap, reservation.ID)
			}
		}
		for id, _ := range affectedMap {
			affectedIds = append(affectedIds, id)
		}
		if len(affectedIds) == 0 {
			t.Rollback()
			return result, sm.ErrCompLockV2CompReserved
		}
		// No reservation found for these locks. Time to (un)lock the component.
		updatedIds, err := t.BulkUpdateCompResLockedTx(affectedIds, newVal)
		if err != nil {
			t.Rollback()
			return result, err
		}
		updatedIdMap := make(map[string]bool)
		for _, id := range updatedIds {
			updatedIdMap[id] = true
			result.Success.ComponentIDs = append(result.Success.ComponentIDs, id)
		}
		if len(affectedIds) != len(updatedIds) {
			for _, id := range affectedIds {
				if _, ok := updatedIdMap[id]; !ok {
					// Shouldn't really happen unless somehow the component
					// was deleted between our GET and UPDATE.
					if f.ProcessingModel == sm.CLProcessingModelRigid {
						t.Rollback()
						return result, sm.ErrCompLockV2Unknown
					}
					fail := sm.CompLockV2Failure{
						ID:     id,
						Reason: sm.CLResultServerError,
					}
					result.Failure = append(result.Failure, fail)
				} else {
					result.Success.ComponentIDs = append(result.Success.ComponentIDs, id)
				}
			}
		}
	default:
		// Invalid action
		t.Rollback()
		return result, ErrHMSDSInvalidCompLockAction
	}

	// Do the counts
	result.Counts.Success = len(result.Success.ComponentIDs)
	result.Counts.Failure = len(result.Failure)
	result.Counts.Total = result.Counts.Success + result.Counts.Failure

	err = t.Commit()
	return result, err
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
func (d *hmsdbPg) InsertJob(j *sm.Job) (string, error) {
	t, err := d.Begin()
	if err != nil {
		return "", err
	}
	// Insert first the Job, with no info
	jobId, err := t.InsertEmptyJobTx(j)
	if err != nil {
		t.Rollback()
		return "", err
	}
	// Insert info for this job
	switch j.Type {
	case sm.JobTypeSRFP:
		data, ok := j.Data.(*sm.SrfpJobData)
		if !ok {
			// Error: bad Job Data
			t.Rollback()
			return "", ErrHMSDSNoJobData
		}
		err = t.InsertStateRFPollJobTx(jobId, data)
	default:
		// Error: bad JobType
		t.Rollback()
		return "", ErrHMSDSArgBadJobType
	}
	if err != nil {
		t.Rollback()
		return "", err
	}
	err = t.Commit()
	return jobId, err
}

// Update the status of the job with the given jobId.
func (d *hmsdbPg) UpdateJob(jobId, status string) (bool, error) {
	t, err := d.Begin()
	if err != nil {
		return false, err
	}
	didUpdate, err := t.UpdateEmptyJobTx(jobId, status)
	if err != nil {
		t.Rollback()
		return false, err
	}
	err = t.Commit()
	return didUpdate, err
}

// Get the job sync entry with the given job id. Nil if not found and nil
// error, otherwise non-nil error (not normally expected).
func (d *hmsdbPg) GetJob(jobId string) (*sm.Job, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	j, err := t.GetEmptyJobTx(jobId)
	if err != nil {
		t.Rollback()
		return nil, err
	} else if j != nil {
		switch j.Type {
		case sm.JobTypeSRFP:
			j.Data, err = t.GetStateRFPollJobByIdTx(jobId)
		default:
			// Error: bad JobType
			t.Rollback()
			return nil, ErrHMSDSArgBadJobType
		}
		if err != nil {
			t.Rollback()
			return nil, err
		}
		if j.Data == nil {
			t.Rollback()
			// Error: No Job Data Found
			return nil, ErrHMSDSNoJobData
		}
	}
	t.Commit()
	return j, err
}

// Get list of jobs from the job sync.
func (d *hmsdbPg) GetJobs(f_opts ...JobSyncFiltFunc) ([]*sm.Job, error) {
	t, err := d.Begin()
	if err != nil {
		return nil, err
	}
	js, err := t.GetEmptyJobsTx(f_opts...)
	if err != nil {
		t.Rollback()
		return nil, err
	} else if len(js) != 0 {
		for _, j := range js {
			switch j.Type {
			case sm.JobTypeSRFP:
				j.Data, err = t.GetStateRFPollJobByIdTx(j.Id)
			default:
				// Error: bad JobType. Skip
				continue
			}
			if err != nil {
				t.Rollback()
				return nil, err
			}
		}
	}
	t.Commit()
	return js, err
}

// Delete the job entry with the given jobId. If no error, bool indicates
// whether component lock was present to remove.
func (d *hmsdbPg) DeleteJob(jobId string) (bool, error) {
	// Build query
	query := sq.Delete(jobTable).
		Where(jobIdCol+" = ?", jobId)

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(d.sc).ExecContext(d.ctx)
	if err != nil {
		return false, err
	}
	// See if any rows were affected
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else {
		if num > 0 {
			if num > 1 {
				d.LogAlways("Error: DeleteJob(): multiple deletions!")
			}
			return true, nil
		}
	}
	return false, nil
}
