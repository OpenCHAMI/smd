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
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/OpenCHAMI/smd/v2/pkg/sm"
	"github.com/Cray-HPE/hms-xname/xnametypes"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

//
// hmsdbPgTx struct - This is a private implementation of the HMSDBTx
// interface.  HMSDBTx represents an abstracted notion of a database
// transaction.  This is used when the package user needs to combine
// multiple operations in an atomic fashion.
//

type hmsdbPgTx struct {
	hdb   *hmsdbPg
	tx    *sql.Tx
	ctx   context.Context
	stmt  *sql.Stmt
	sc    *sq.StmtCache
	query string
}

// This should only be called by hdb.Begin()
func newHMSDBPgTx(hdb *hmsdbPg) (HMSDBTx, error) {
	var err error

	t := new(hmsdbPgTx)
	t.hdb = hdb

	// Note this is just a placeholder so we don't have a null pointer
	// Later we will do more useful things, so we are just using the context
	// aware versions of the sql functions so they are plumbed.
	t.ctx = context.TODO()

	// Create a new transaction from from using the exiting DB connection pool
	t.tx, err = t.hdb.db.BeginTx(t.ctx, nil)
	if err != nil {
		return nil, err
	}
	t.sc = sq.NewStmtCache(t.tx)
	return t, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Helper functions
//
////////////////////////////////////////////////////////////////////////////

// Prepare the given query string and return a pointer to the statement
// handle.  If the same query is being done (i.e. in a transaction), return
// a cached pointer so it does not need to be reprepared multiple times.
// If another query is being done, close the cached stmt handle and prepare
// the new one, caching it in the hmsdbTx.  Ditto the latter when the
// cached statment will be the first for the transaction.
func (t *hmsdbPgTx) conditionalPrepare(qname, query string) (*sql.Stmt, error) {
	var err error
	if query == "" {
		t.LogAlways("Error: %s(): query was empty.", qname)
		return nil, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	// Replace ? placeholders for prepared query args with postgres type.
	// These have to be numbered ($1, $2,...), but since the number of args
	// matches the number of placeholders, this is no problem, though we
	// don't take advantage of not needing to reuse things that appear
	// more than once.  NOTE THAT THIS DOES NOT IGNORE QUOTED ?
	// VALUES.  So don't build queries like that (the arguments
	// themselves can have whatever).
	origQuery := query
	query = ToPGQueryArgs(origQuery)
	if t.query == "" {
		t.stmt, err = t.tx.PrepareContext(t.ctx, query)
		if err != nil {
			t.LogAlways("Error: Prepare(%s) Failed: %s", qname, err)
			return nil, err
		}
		t.query = query
	} else if t.query != query {
		if t.stmt != nil {
			if err := t.stmt.Close(); err != nil {
				t.LogAlways("Warning: %s: Failed to close old stmt: %s",
					qname, err)
			}
		}
		t.query = ""
		t.stmt, err = t.tx.PrepareContext(t.ctx, query)
		if err != nil {
			t.LogAlways("Error: Re-prepare(%s) Failed: %s", qname, err)
			return nil, err
		}
		t.query = query
	}
	return t.stmt, nil
}

// Helper function.  Prepares, and issues query with the provided arguments.
// Close *Rows when operation is complete.
func (t *hmsdbPgTx) getRowsForQuery(qname, query string, args ...interface{}) (*sql.Rows, error) {
	t.Log(LOG_DEBUG, "%s(%v) starting....", qname, args)

	stmt, err := t.conditionalPrepare(qname, query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.QueryContext(t.ctx, args...)
	return rows, err
}

//
// Logging
//

// Produces an object implementing a log interface.
func (t *hmsdbPgTx) GetLog() *log.Logger {
	return t.hdb.lg
}

// Abstraction layer to logging infrastructure.  Works like log.Printf.
func (t *hmsdbPgTx) LogAlways(format string, a ...interface{}) {
	// Call depth of 2 - report caller's line number.
	t.hdb.lg.Output(2, fmt.Sprintf(format, a...))
}

// Works like log.Printf, but registers error for function calling the
// function that is printing the error. e.g. instead of always saying
// an error occurred in begin(), we show where begin() was called, so
// we don't have to guess, and the message can make clear what in begin()
// failed.
func (t *hmsdbPgTx) LogAlwaysParentFunc(format string, a ...interface{}) {
	// Call depth of 3 - report caller's caller's line number.
	t.hdb.lg.Output(3, fmt.Sprintf(format, a...))
}

// Conditional logging based on current log level
func (t *hmsdbPgTx) Log(l LogLevel, format string, a ...interface{}) {
	if int(l) <= int(t.hdb.lgLvl) {
		// Call depth of 2 - report caller's line number.
		t.hdb.lg.Output(2, fmt.Sprintf(format, a...))
	}
}

// Checks log level of parent connection pool vs arg
func (t *hmsdbPgTx) IsLogLevel(lvl LogLevel) bool {
	if int(t.hdb.lgLvl) <= int(lvl) {
		return true
	} else {
		return false
	}
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface Implementation
//
/////////////////////////////////////////////////////////////////////////////

// Terminates transaction, reversing all changes made prior to Begin()
func (t *hmsdbPgTx) Rollback() error {
	if t.stmt != nil {
		if err := t.stmt.Close(); err != nil {
			t.LogAlways("Warning: Rollback(): Failed to close old stmt: %s", err)
		}
	}
	return t.tx.Rollback()
}

// Terminates transaction successfully, committing all operations
// performed against it in an atomic fashion.  Closes any non-nil
// statement handles that may still be open.
func (t *hmsdbPgTx) Commit() error {
	if t.stmt != nil {
		t.stmt.Close()
		if err := t.stmt.Close(); err != nil {
			t.LogAlways("Warning: Commit(): Failed to close old stmt: %s", err)
		}
	}
	return t.tx.Commit()
}

// Checks to see if parent connection pool is still healthy.
func (t *hmsdbPgTx) IsConnected() bool {
	return t.hdb.connected
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - Generic value queries
//
/////////////////////////////////////////////////////////////////////////////

// For queries that obtain a single string value such as an ID/xname.  The
// entry type does not matter as long as the query returns one string per row.
func (t *hmsdbPgTx) querySingleStringValue(qname, query string, args ...interface{}) ([]string, error) {
	vals := make([]string, 0, 1)
	rows, err := t.getRowsForQuery(qname, query, args...)
	if err != nil {
		return vals, err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		val, err := t.hdb.scanSingleStringValue(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return vals, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, val)
		if val != nil {
			vals = append(vals, *val)
		}
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d values.", qname, args, len(vals))
	return vals, err
}

// Get the id values for either all labels in the given table, or a
// filtered set based on filter f (*ComponentFilter, *RedfishEPFilter,
// *CompEPFilter)
// Use one of the *Table values in hmsds-api for 'table' arg,
// e.g. ComponentsTable, RedfishEndpointsTable, etc.
func (t *hmsdbPgTx) GetIDListTx(tbl string, f interface{}) ([]string, error) {
	query := ""
	var args []interface{}

	// Name of function to use in errors, if not given by filter.label
	label := "GetIDListTx"

	var err error = nil
	switch tbl {
	case ComponentsTable:
		queryBase := getCompIDPrefix
		if filter, ok := f.(*ComponentFilter); ok {
			query, args, err = buildComponentQuery(queryBase, filter)
			if filter.label != "" {
				label = filter.label
			}
		} else {
			query = queryBase + ";"
		}
	case RedfishEndpointsTable:
		queryBase := getRFEndpointIDPrefix
		if filter, ok := f.(*RedfishEPFilter); ok {
			query, args, err = buildRedfishEPQuery(queryBase, filter)
			if filter.label != "" {
				label = filter.label
			}
		} else {
			query = queryBase + ";"
		}
	case ComponentEndpointsTable:
		queryBase := getCompEndpointIDPrefix
		if filter, ok := f.(*CompEPFilter); ok {
			query, args, err = buildCompEPQuery(queryBase, filter)
			if filter.label != "" {
				label = filter.label
			}
		} else {
			query = queryBase + ";"
		}
	case NodeMapTable:
		query = getNodeMapIDPrefix + ";"
	case HWInvByLocTable:
		query = getHWInvByLocIDPrefix + ";"
	case HWInvByFRUTable:
		query = getHWInvByFRUIDPrefix + ";"
	case DiscoveryStatusTable:
		query = getDiscoveryStatusIDPrefix + ";"
	default:
		t.LogAlways("Error: %s(): Bad input: %s", label, err)
		return []string{}, ErrHMSDSArgMissing
	}
	// Check for error building query
	if err != nil {
		t.LogAlways("Error: %s(): Got error '%s' building query: %s",
			label, err, query)
		return []string{}, err
	}
	// Call DB and get IDs per table contents and optional filter.
	return t.querySingleStringValue(label, query, args...)
}

// Build filter query for Component IDs using filter functions and
// then return the list of matching xname IDs as a string array, write locking
// the rows if requested.  The IDs will be normalized.
func (t *hmsdbPgTx) GetComponentIDsTx(f_opts ...CompFiltFunc) ([]string, error) {
	f := new(ComponentFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	affectedIDs, err := t.GetIDListTx(ComponentsTable, f)
	if err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Build filter query for ComponentEndpoints IDs using filter functions and
// then return the list of matching xname IDs as a string array, write locking
// the rows if requested.
func (t *hmsdbPgTx) GetCompEndpointIDsTx(
	f_opts ...CompEPFiltFunc,
) ([]string, error) {
	f := new(CompEPFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	affectedIDs, err := t.GetIDListTx(ComponentEndpointsTable, f)
	if err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

// Build filter query for RedfishEndpoints IDs using filter functions and
// then return the list of matching xname IDs as a string array, write locking
// the rows if requested.
func (t *hmsdbPgTx) GetRFEndpointIDsTx(
	f_opts ...RedfishEPFiltFunc,
) ([]string, error) {

	f := new(RedfishEPFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	affectedIDs, err := t.GetIDListTx(RedfishEndpointsTable, f)
	if err != nil {
		return []string{}, err
	}
	return affectedIDs, nil
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - Component queries
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more HMS Component rows in
// the result.
func (t *hmsdbPgTx) queryComponent(qname string, fltr FieldFilter, query string, args ...interface{}) ([]*base.Component, error) {
	// Add a row filter to only get the rows that we want
	switch fltr {
	case FLTR_DEFAULT:
	case FLTR_STATEONLY:
		query = strings.TrimSuffix(query, ";")
		query = getCompStatePrefix + query + suffixCompFilter
	case FLTR_FLAGONLY:
		query = strings.TrimSuffix(query, ";")
		query = getCompFlagPrefix + query + suffixCompFilter
	case FLTR_ROLEONLY:
		query = strings.TrimSuffix(query, ";")
		query = getCompRolePrefix + query + suffixCompFilter
	case FLTR_NIDONLY:
		query = strings.TrimSuffix(query, ";")
		query = getCompNIDPrefix + query + suffixCompFilter
	default:
		// Default to using the default filter
		fltr = FLTR_DEFAULT
	}
	t.Log(LOG_DEBUG, "Debug: %s(%v) starting query '%s'",
		qname, args, strings.Replace(query, "\n", " ", -1))
	stmt, err := t.conditionalPrepare(qname, query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.QueryContext(t.ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comps := make([]*base.Component, 0, 1)
	i := 0
	for rows.Next() {
		comp, err := t.hdb.scanComponent(rows, fltr)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return comps, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, comp)
		comps = append(comps, comp)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d comps.", qname, args, len(comps))
	return comps, err
}

// Back end for all queries that produce one or more HMS Component rows in
// the result.
func (t *hmsdbPgTx) sqQueryComponent(q sq.SelectBuilder,
	qname string, fltr FieldFilter) ([]*base.Component, error) {

	queryString, args, _ := q.ToSql()
	t.Log(LOG_DEBUG, "%s(): Submitting '%s' with '%v'",
		qname, queryString, args)

	// Run provided query to get rows to scan.
	q = q.PlaceholderFormat(sq.Dollar)
	rows, err := q.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: %s(): query failed: %s", qname, err)
		return nil, err
	}
	defer rows.Close()

	comps := make([]*base.Component, 0, 1)
	i := 0
	for rows.Next() {
		comp, err := t.hdb.scanComponent(rows, fltr)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return comps, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, comp)
		comps = append(comps, comp)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d comps.", qname, args, len(comps))
	return comps, err
}

// Build filter query for State/Components using filter functions and
// then return the set of matching components as an array, write locking
// the rows if requested.
//
// NOTE: Most args allow negated arguments, i.e. "!x0c0s0b0", so be careful about
// passing in user data if the query should only return a single result. ID()
// does not and takes a single arg.
func (t *hmsdbPgTx) GetComponentsTx(f_opts ...CompFiltFunc) ([]*base.Component, error) {

	f := new(ComponentFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	return t.GetComponentsFilterTx(f, FLTR_DEFAULT)
}

// Same as above, but allows only certain fields to be returned
// via FieldFilter
func (t *hmsdbPgTx) GetComponentsFieldFilterTx(
	fieldFltr FieldFilter,
	f_opts ...CompFiltFunc,
) ([]*base.Component, error) {

	f := new(ComponentFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	return t.GetComponentsFilterTx(f, fieldFltr)
}

// Look up a single HMS Component by id, i.e. xname (in transaction).
func (t *hmsdbPgTx) GetComponentByIDTx(id string) (*base.Component, error) {
	return t.getComponentByIDRawTx(id, false)
}

// Look up a single HMS Component by id, i.e. xname (in transaction).
// THIS WILL CREATE A WRITE LOCK ON THE ENTRY, so the transaction should
// not be kept open longer than needed.
func (t *hmsdbPgTx) GetComponentByIDForUpdateTx(id string) (*base.Component, error) {
	return t.getComponentByIDRawTx(id, true)
}

// Helper... get comp by ID, with or without SELECT ... FOR UPDATE
func (t *hmsdbPgTx) getComponentByIDRawTx(id string, for_update bool) (*base.Component, error) {
	fname := "GetComponentByIDTx"
	query := getComponentByIDQuery
	if for_update == true {
		fname = "GetComponentByIDForUpdateTx"
		query = getComponentByIDForUpdQuery
	}
	if id == "" {
		t.LogAlways("Error: %s(): xname was empty", fname)
		return nil, ErrHMSDSArgMissing
	}
	// Perform corresponding query on DB
	comps, err := t.queryComponent(fname, FLTR_DEFAULT, query,
		xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(comps) == 0 {
		t.Log(LOG_INFO, "Info: %s(%s) matched no comps.",
			fname, xnametypes.NormalizeHMSCompID(id))
		return nil, nil
	} else if len(comps) > 1 {
		t.LogAlways("WARNING: %s(%s): multiple comps!.",
			fname, xnametypes.NormalizeHMSCompID(id))
	}
	return comps[0], nil
}

// Get all HMS Components in system (in transaction).
func (t *hmsdbPgTx) GetComponentsAllTx() ([]*base.Component, error) {
	return t.GetComponentsTx(From("GetComponentsAllTx"))
}

// Get some or all HMS Components in system (in transaction), with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (t *hmsdbPgTx) GetComponentsFilterTx(f *ComponentFilter, fieldFltr FieldFilter) ([]*base.Component, error) {
	var comps []*base.Component
	var err error
	label := "GetComponentsFilterTx"

	query, err := selectComponents(f, fieldFltr)
	if err != nil {
		t.LogAlways("Error: %s(): makeComponentQuery failed: %s", label, err)
		return comps, err
	}
	// Perform corresponding query on DB
	comps, err = t.sqQueryComponent(query, label, fieldFltr)
	if err != nil {
		return comps, err
	}
	// Query succeeded.
	if len(comps) == 0 {
		t.Log(LOG_INFO, "Info: %s(): no matches", label)
		return comps, nil
	}
	return comps, nil
}

// Get some or all HMS Components in system (in transaction) under
// a set of parent components, with filtering options to possibly
// narrow the returned values. If no filter provided, just get
// the parent components.  Otherwise use it to create a custom
// WHERE... string that filters out entries that do not match ALL
// of the non-empty strings in the filter struct.
func (t *hmsdbPgTx) GetComponentsQueryTx(f *ComponentFilter, fieldFltr FieldFilter, ids []string) ([]*base.Component, error) {
	var comps []*base.Component
	var err error
	label := "GetComponentsQueryTx"

	for i := 0; i < len(ids); i++ {
		if ids[i] == "s0" || ids[i] == "all" {
			// Reseting the length of our xname slice to 0
			// will cause the query to not be filtered by xname.
			ids = ids[:0]
			break
		}
	}
	if f == nil {
		f = new(ComponentFilter)
	}
	// Get query string
	query, err := selectComponentsHierarchy(f, fieldFltr, ids)
	if err != nil {
		return nil, err
	}
	// Perform corresponding query on DB
	comps, err = t.sqQueryComponent(query, label, fieldFltr)
	if err != nil {
		return nil, err
	}
	// Query succeeded.
	if len(comps) == 0 {
		t.Log(LOG_INFO, "Info: %s(): no matches: %v", label, ids)
		return comps, nil
	}
	return comps, nil
}

// Get a single HMS Component by its NID, if the NID exists (in transaction)
func (t *hmsdbPgTx) GetComponentByNIDTx(nid string) (*base.Component, error) {
	if nid == "" {
		t.LogAlways("Error: GetComponentByNID(): NID was empty.")
		return nil, ErrHMSDSArgMissing
	}
	// Verify nid string is an int and is positive.
	if i, err := strconv.Atoi(nid); err != nil {
		return nil, ErrHMSDSArgNotAnInt
	} else if i < 0 {
		return nil, ErrHMSDSArgBadRange
	}
	// Perform corresponding query on DB
	comps, err := t.queryComponent("GetComponentByNID", FLTR_DEFAULT,
		getComponentByNIDQuery, nid)
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(comps) == 0 {
		t.Log(LOG_INFO, "Info: GetComponentByNID(%s) matched no comps.", nid)
		return nil, nil
	} else if len(comps) > 1 {
		t.LogAlways("WARNING: GetComponentByNID(%s): multiple comps!.", nid)
	}
	return comps[0], nil
}

// Insert HMS Component into database, updating it if it exists.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (t *hmsdbPgTx) InsertComponentTx(c *base.Component) (int64, error) {
	var enabledFlg bool
	if c == nil {
		t.LogAlways("Error: InsertComponentTx(): Component was nil.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// If NID is not a valid number (e.g. empty string), set to -1.
	//
	var rawNID int64
	if num, err := c.NID.Int64(); err != nil {
		rawNID = -1
	} else {
		rawNID = num
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("InsertComponentTx", insertPgCompQuery)
	if err != nil {
		return 0, err
	}
	// Default to enabled.
	if c.Enabled == nil {
		enabledFlg = true
	} else {
		enabledFlg = *c.Enabled
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(c.ID)

	// Perform insert
	result, err := stmt.ExecContext(t.ctx,
		&normID,
		&c.Type,
		&c.State,
		&c.Flag,
		&enabledFlg,
		&c.SwStatus,
		&c.Role,
		&c.SubRole,
		&rawNID,
		&c.Subtype,
		&c.NetType,
		&c.Arch,
		&c.Class,
		&c.ReservationDisabled,
		&c.Locked)
	if err != nil {
		t.LogAlways("Error: InsertComponentTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_DEBUG, "Debug: InsertComponentTx() - %v", c)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Insert HMS Components into the database, updating it if it exists.
// Returns the IDs of the affected components.
//
// Example Query:
// INSERT INTO components (
//
//	id,
//	type,
//	state,
//	flag,
//	enabled,
//	admin,
//	role,
//	subrole,
//	nid,
//	subtype,
//	nettype,
//	arch,
//	class,
//	reservation_disabled,
//	locked)
//
// VALUES
//
//	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15),
//	...
//	($#, $#, $#, $#, $#, $#, $#, $#, $#, $#, $#, $#, $#, $#, $#)
//
// ON CONFLICT(id) DO UPDATE SET
//
//	state = EXCLUDED.state,
//	flag = EXCLUDED.flag,
//	subtype = EXCLUDED.subtype,
//	nettype = EXCLUDED.nettype,
//	arch = EXCLUDED.arch,
//	class = EXCLUDED.class
//
// RETURNING *;
func (t *hmsdbPgTx) InsertComponentsTx(comps []*base.Component) ([]string, error) {
	results := []string{}
	if len(comps) == 0 {
		return []string{}, nil
	}
	if !t.IsConnected() {
		return []string{}, ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(compTable).
		Columns(compColsDefault...)

	for _, c := range comps {
		// Normalize key
		normID := xnametypes.NormalizeHMSCompID(c.ID)
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[normID]; ok {
			continue
		} else {
			valueMap[normID] = true
		}
		// If NID is not a valid number (e.g. empty string), set to -1.
		var rawNID int64
		if num, err := c.NID.Int64(); err != nil {
			rawNID = -1
		} else {
			rawNID = num
		}

		// Default to enabled.
		var enabledFlg bool
		if c.Enabled == nil {
			enabledFlg = true
		} else {
			enabledFlg = *c.Enabled
		}

		// Set fields for the INSERT
		query = query.Values(
			normID,
			c.Type,
			c.State,
			c.Flag,
			enabledFlg,
			c.SwStatus,
			c.Role,
			c.SubRole,
			rawNID,
			c.Subtype,
			c.NetType,
			c.Arch,
			c.Class,
			c.ReservationDisabled,
			c.Locked)
	}
	query = query.Suffix("ON CONFLICT(" + compIdCol + ") DO UPDATE SET " +
		compStateCol + " = EXCLUDED." + compStateCol + ", " +
		compFlagCol + " = EXCLUDED." + compFlagCol + ", " +
		compSubTypeCol + " = EXCLUDED." + compSubTypeCol + ", " +
		compNetTypeCol + " = EXCLUDED." + compNetTypeCol + ", " +
		compArchCol + " = EXCLUDED." + compArchCol + ", " +
		compClassCol + " = EXCLUDED." + compClassCol +
		" RETURNING " + compIdCol)

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: InsertComponentsTx(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: InsertComponentsTx(): QueryContext: %s", err)
		return []string{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return []string{}, err
		}
		results = append(results, id)
	}
	return results, nil
}

// Update state and flag fields only in DB for xname IDs 'ids'
// If force = true ignores any starting state restrictions and will always
// set ids to state, unless it is already set.
//
// If noVerify = true, don't add extra where clauses to ensure only the
// rows that should change do.  If true, either we already verified that
// the ids list will be changed, and have locked the rows, or else we don't
// care to know the exact ids that actually changed.
//
// Returns the number of affected rows. < 0 means RowsAffected() is not
// supported.
//
//	Note: If flag is not set, it will be set to OK (i.e. no flag)
func (t *hmsdbPgTx) UpdateCompStatesTx(
	ids []string,
	state, flag string,
	force, noVerify bool,
	pi *PartInfo,
) (int64, error) {
	var filterQuery string
	var args []interface{}
	var fname string = "UpdateCompStatesTx"
	var startStates []string
	var err error

	if ids == nil || len(ids) == 0 {
		t.LogAlways("Error: %s(): ID list was nil.", fname)
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// If flag is not provided, default is to unset flag, i.e. set to OK.
	nflag := base.VerifyNormalizeFlagOK(flag)

	nstate := base.VerifyNormalizeState(state)
	if nstate == "" {
		return 0, ErrHMSDSArgBadState
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
		startStates, err = base.GetValidStartStateWForce(state, force)
		if err != nil {
			return 0, err
		}
	}
	// Start an State and Flag update for c.ID.  Note:  this normalizes all
	// of the fields and verifies them.
	q, err := startCompUpdate(ids, true, State(state), Flag(nflag), FlagCondNoChange(base.FlagLocked.String()))
	if err != nil {
		return 0, err
	}
	// Construct rest of WHERE clause, if needed.
	if noVerify == true {
		// We don't care if the state actually changes, either because we
		// already locked the rows, or we don't need detailed info on
		// who updated what.
		filterQuery, args, err = finishCompUpdate(q, From(fname))
	} else {
		// Finish the where clause so we only change states if value
		// will actually be updated.
		filterQuery, args, err = finishCompUpdate(
			q,
			NotStateOrFlag(state, nflag),
			States(startStates),
			From(fname))
	}
	if err != nil {
		return 0, err
	}
	// Prepare statement if this query is not already prepared
	stmt, err := t.conditionalPrepare(fname, filterQuery)
	if err != nil {
		return 0, err
	}
	// Update entry
	result, err := stmt.ExecContext(t.ctx, args...)
	if err != nil {
		t.LogAlways("Error: %s(): stmt.Exec: %s", fname, err)
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update Flag field in DB from c's Flag field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Flag cannot be blank/invalid.
func (t *hmsdbPgTx) UpdateCompFlagOnlyTx(id string, flag string) (int64, error) {
	if len(id) == 0 {
		t.LogAlways("Error: UpdateCompFlagOnlyTx(): ID was empty.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Flag is mandatory
	if len(flag) == 0 {
		return 0, ErrHMSDSArgMissing
	} else {
		// Verify input flag and normalize any capitalization differences.
		flag = base.VerifyNormalizeFlag(flag)
		if flag == "" {
			return 0, ErrHMSDSArgNoMatch
		}
	}
	// Prepare statement
	stmt, err := t.conditionalPrepare("UpdateCompFlagOnlyTx",
		updateCompFlagOnlyByIDQuery)
	if err != nil {
		return 0, err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(id)

	// Make update in database.
	result, err := stmt.ExecContext(t.ctx,
		&flag,
		&normID)
	if err != nil {
		t.LogAlways("Error: UpdateCompFlagOnlyTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_DEBUG, "Debug: UpdateCompFlagOnlyTx() - %v", normID)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update flag field in DB for a list of components.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Flag cannot be empty/invalid.
func (t *hmsdbPgTx) BulkUpdateCompFlagOnlyTx(ids []string, flag string) (int64, error) {
	var filterQuery string
	var err error
	args := make([]interface{}, 0, 1)
	newArgs := make([]interface{}, 0, 1)

	if ids == nil || len(ids) == 0 {
		t.LogAlways("Error: BulkUpdateCompFlagOnlyTx(): ID list was nil.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}

	// Verify input flag and normalize any capitalization differences.
	flag = base.VerifyNormalizeFlag(flag)
	if flag == "" {
		return 0, ErrHMSDSArgNoMatch
	}

	args = append(args, flag)
	filterQuery, newArgs, err = buildBulkCompUpdateQuery(updateCompFlagOnlyPrefix, ids)
	if err != nil {
		return 0, err
	}
	args = append(args, newArgs...)

	// Prepare statement if this query is not already prepared
	stmt, err := t.conditionalPrepare("BulkUpdateCompFlagOnlyTx", filterQuery)
	if err != nil {
		return 0, err
	}
	// Update entry
	result, err := stmt.ExecContext(t.ctx, args...)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompFlagOnlyTx: stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: BulkUpdateCompFlagOnlyTx(len=%d) - %s",
		len(ids), flag)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update enabled field in DB from c's Enabled field (in transaction).
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: c.Enabled cannot be nil.
func (t *hmsdbPgTx) UpdateCompEnabledTx(id string, enabled bool) (int64, error) {
	var enabledFlg bool
	if len(id) == 0 {
		t.LogAlways("Error: UpdateCompEnabledTx(): ID was empty.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Enabled is mandatory
	enabledFlg = enabled
	stmt, err := t.conditionalPrepare("UpdateCompEnabledTx",
		updateCompEnabledByIDQuery)
	if err != nil {
		return 0, err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(id)

	// Make update in database.
	result, err := stmt.ExecContext(t.ctx,
		&enabledFlg,
		&normID)
	if err != nil {
		t.LogAlways("Error: UpdateCompEnabledTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: UpdateCompEnabledTx() - %s, %v", normID, enabled)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update Enabled field only in DB for a list of components (in transaction)
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (t *hmsdbPgTx) BulkUpdateCompEnabledTx(ids []string, enabled bool) (int64, error) {
	var filterQuery string
	var err error
	args := make([]interface{}, 0, 1)
	newArgs := make([]interface{}, 0, 1)

	if ids == nil || len(ids) == 0 {
		t.LogAlways("Error: BulkUpdateCompEnabledTx(): ID list was nil.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	args = append(args, enabled)
	filterQuery, newArgs, err = buildBulkCompUpdateQuery(updateCompEnabledPrefix, ids)
	if err != nil {
		return 0, err
	}
	args = append(args, newArgs...)

	// Prepare statement if this query is not already prepared
	stmt, err := t.conditionalPrepare("BulkUpdateCompEnabledTx", filterQuery)
	if err != nil {
		return 0, err
	}
	// Update entry
	result, err := stmt.ExecContext(t.ctx, args...)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompEnabledTx: stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: BulkUpdateCompEnabledTx(len=%d) - %t",
		len(ids), enabled)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update SwStatus field in DB from c's SwStatus field (in transaction).
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (t *hmsdbPgTx) UpdateCompSwStatusTx(id string, swStatus string) (int64, error) {
	if len(id) == 0 {
		t.LogAlways("Error: UpdateCompSwStatusTx(): ID was empty.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// NOTE: Managed plane is expected to be responsible for verifying
	// input.  TODO: Should empty string be allowed?
	stmt, err := t.conditionalPrepare("UpdateCompSwStatusTx",
		updateCompSwStatusByIDQuery)
	if err != nil {
		return 0, err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(id)

	// Make update in database.
	result, err := stmt.ExecContext(t.ctx,
		&swStatus,
		&normID)
	if err != nil {
		t.LogAlways("Error: UpdateCompSwStatusTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_DEBUG, "Debug: UpdateCompSwStatusTx() - %s, %s", normID, swStatus)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update SwStatus field only in DB for a list of components
// (In transaction.)
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (t *hmsdbPgTx) BulkUpdateCompSwStatusTx(ids []string, swstatus string) (int64, error) {
	var filterQuery string
	var err error
	args := make([]interface{}, 0, 1)
	newArgs := make([]interface{}, 0, 1)

	if ids == nil || len(ids) == 0 {
		t.LogAlways("Error: BulkUpdateCompSwStatusTx(): ID list was nil.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	args = append(args, swstatus)
	filterQuery, newArgs, err = buildBulkCompUpdateQuery(updateCompSwStatusPrefix, ids)
	if err != nil {
		return 0, err
	}
	args = append(args, newArgs...)

	// Prepare statement if this query is not already prepared
	stmt, err := t.conditionalPrepare("BulkUpdateCompSwStatusTx", filterQuery)
	if err != nil {
		return 0, err
	}
	// Update entry
	result, err := stmt.ExecContext(t.ctx, args...)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompSwStatusTx: stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: BulkUpdateCompSwStatusTx(len=%d) - %s",
		len(ids), swstatus)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update Role/SubRole field in DB from c's Role/SubRole field.
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Role cannot be blank/invalid.
func (t *hmsdbPgTx) UpdateCompRoleTx(id string, role, subRole string) (int64, error) {
	if len(id) == 0 {
		t.LogAlways("Error: UpdateCompRoleTx(): ID was empty.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Role is mandatory
	if len(role) == 0 {
		return 0, ErrHMSDSArgMissing
	} else {
		// Verify input role and normalize any capitalization differences.
		role = base.VerifyNormalizeRole(role)
		if role == "" {
			return 0, ErrHMSDSArgNoMatch
		}
	}
	// SubRole can be empty
	if len(subRole) != 0 {
		// Verify input subRole and normalize any capitalization differences.
		subRole = base.VerifyNormalizeSubRole(subRole)
		if subRole == "" {
			return 0, ErrHMSDSArgNoMatch
		}
	}
	stmt, err := t.conditionalPrepare("UpdateCompRoleTx",
		updateCompRoleByIDQuery)
	if err != nil {
		return 0, err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(id)

	// Make update in database.
	result, err := stmt.ExecContext(t.ctx,
		&role,
		&subRole,
		&normID)
	if err != nil {
		t.LogAlways("Error: UpdateCompRoleTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_DEBUG, "Debug: UpdateCompRoleTx(): - %s, %s, %s", normID, role, subRole)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update Role/SubRole field only in DB for a list of components
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
// Note: Role cannot be blank/invalid.
func (t *hmsdbPgTx) BulkUpdateCompRoleTx(ids []string, role, subRole string) (int64, error) {
	var filterQuery string
	var err error
	args := make([]interface{}, 0, 1)
	newArgs := make([]interface{}, 0, 1)

	if ids == nil || len(ids) == 0 {
		t.LogAlways("Error: BulkUpdateCompRoleTx(): ID list was nil.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}

	// Verify input role and normalize any capitalization differences.
	role = base.VerifyNormalizeRole(role)
	if role == "" {
		return 0, ErrHMSDSArgNoMatch
	}
	// Verify input subRole and normalize any capitalization differences.
	if len(subRole) != 0 {
		subRole = base.VerifyNormalizeSubRole(subRole)
		if subRole == "" {
			return 0, ErrHMSDSArgNoMatch
		}
	}

	args = append(args, role, subRole)
	filterQuery, newArgs, err = buildBulkCompUpdateQuery(updateCompRolePrefix, ids)
	if err != nil {
		return 0, err
	}
	args = append(args, newArgs...)

	// Prepare statement if this query is not already prepared
	stmt, err := t.conditionalPrepare("BulkUpdateCompRoleTx", filterQuery)
	if err != nil {
		return 0, err
	}
	// Update entry
	result, err := stmt.ExecContext(t.ctx, args...)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompRoleTx: stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: BulkUpdateCompRoleTx(len=%d) - %s, %s",
		len(ids), role, subRole)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update Class field only in DB for a list of components
// (In transaction.)
// Returns the number of affected rows. < 0 means RowsAffected() is not supported.
func (t *hmsdbPgTx) BulkUpdateCompClassTx(ids []string, class string) (int64, error) {
	var filterQuery string
	var err error
	args := make([]interface{}, 0, 1)
	newArgs := make([]interface{}, 0, 1)

	if ids == nil || len(ids) == 0 {
		t.LogAlways("Error: BulkUpdateCompClassTx(): ID list was nil.")
		return 0, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	args = append(args, class)
	filterQuery, newArgs, err = buildBulkCompUpdateQuery(updateCompClassPrefix, ids)
	if err != nil {
		return 0, err
	}
	args = append(args, newArgs...)

	// Prepare statement if this query is not already prepared
	stmt, err := t.conditionalPrepare("BulkUpdateCompClassTx", filterQuery)
	if err != nil {
		return 0, err
	}
	// Update entry
	result, err := stmt.ExecContext(t.ctx, args...)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompClassTx: stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: BulkUpdateCompClassTx(len=%d) - %s",
		len(ids), class)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This likely means that RowsAffected() is unsupported.
		// Default to reporting that an update happened by returning non-zero.
		return -1, nil
	}
	return rowsAffected, nil
}

// Update NID.  If NID is not set or negative, it is set to -1 which
// effectively unsets it and suppresses its output.
func (t *hmsdbPgTx) UpdateCompNIDTx(c *base.Component) error {
	if c == nil {
		t.LogAlways("Error: UpdateCompNIDTx(): Component was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// NID must be given.  DB uses value for -1 for no NID, client should
	// send negative value to unset NID.
	var rawNID int64
	if num, err := c.NID.Int64(); err != nil {
		return ErrHMSDSArgMissingNID
	} else if num < -1 {
		rawNID = -1
	} else {
		rawNID = num
	}
	stmt, err := t.conditionalPrepare("UpdateCompNIDTx", updateCompNIDByIDQuery)
	if err != nil {
		return err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(c.ID)

	// Make update in database.
	_, err = stmt.ExecContext(t.ctx,
		&rawNID,
		&normID)
	if err != nil {
		t.LogAlways("Error: UpdateCompNIDTx(): stmt.Exec: %s", err)
		return err
	}
	t.Log(LOG_DEBUG, "DEBUG: UpdateCompNIDTx(%s): - %d",
		normID, rawNID)
	return nil
}

// Update the NIDs for a list of components.  If NID is not set or negative, it is set to -1 which
// effectively unsets it and suppresses its output.
func (t *hmsdbPgTx) BulkUpdateCompNIDTx(comps []base.Component) error {
	if len(comps) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	idColAlias := compTableJoinAlias + "." + compIdCol
	idFromCol := compTableSubAlias + "." + compIdCol
	nidFromCol := compTableSubAlias + "." + compNIDCol
	// Generate query
	// Make the column name a sq.Sqlizer so sq will set it as a column name and not a value.
	query := sq.Update(compTable+" "+compTableJoinAlias).
		Set(compNIDCol, sq.Expr(nidFromCol))

	// sq doesn't have a way to add a FROM statement to an UPDATE.
	// We'll just construct the 2nd half of the query manually and
	// add it as a suffix.
	args := make([]interface{}, 0, 1)
	valStr := ""
	for i, c := range comps {
		// Normalize key
		normID := xnametypes.NormalizeHMSCompID(c.ID)
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[normID]; ok {
			continue
		} else {
			valueMap[normID] = true
		}
		// NID must be given. DB uses value for -1 for no NID, client should
		// send negative value to unset NID.
		var rawNID int64
		if num, err := c.NID.Int64(); err != nil {
			return ErrHMSDSArgMissingNID
		} else if num < -1 {
			rawNID = -1
		} else {
			rawNID = num
		}
		// Add the values to our values table
		if i == 0 {
			valStr += "(?,?::BIGINT)"
		} else {
			valStr += ",(?,?::BIGINT)"
		}
		args = append(args,
			normID,
			rawNID)
	}
	// This FROM statement builds us a values table to pull update values
	// from so an update with different values for each row can be done.
	qstr := "FROM (VALUES " + valStr + ") AS " + compTableSubAlias +
		"(" + compIdCol + ", " + compNIDCol + ") " +
		"WHERE " + idColAlias + " = " + idFromCol
	query = query.Suffix(qstr, args...)

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: BulkUpdateCompNIDTx(): Query: %s - With args: %v", qStr, qArgs)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompNIDTx(): ExecContext: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: BulkUpdateCompNIDTx() - %s", res)
	return nil
}

// Delete HMS Component with matching xname id from database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteComponentByIDTx(id string) (bool, error) {
	if id == "" {
		t.LogAlways("Error: DeleteComponentByIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteComponentByIDTx",
		deleteComponentByIDQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		t.LogAlways("Error: DeleteComponentByIDTx(%s): stmt.Exec: %s", id, err)
		return false, err
	}
	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	t.Log(LOG_INFO, "Info: DeleteComponentByIDTx(%s) - %d", id, num)
	if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all HMS Components from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteComponentsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteComponentsAllTx",
		deleteComponentsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteComponentsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_DEBUG, "Debug: DeleteComponentsAllTx() - OK")

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - Node NID Mapping queries
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more Node->NID Mapping rows in
// the result.
func (t *hmsdbPgTx) queryNodeMap(qname, query string, args ...interface{}) ([]*sm.NodeMap, error) {
	t.Log(LOG_DEBUG, "Debug: %s(%v) starting....", qname, args)

	stmt, err := t.conditionalPrepare(qname, query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.QueryContext(t.ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nnms := make([]*sm.NodeMap, 0, 1)
	i := 0
	for rows.Next() {
		nnm, err := t.hdb.scanNodeMap(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return nnms, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, nnm)
		nnms = append(nnms, nnm)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d entries.", qname, args, len(nnms))
	return nnms, err
}

// Look up one Node->NID Mapping by id, i.e. node xname (in transaction).
func (t *hmsdbPgTx) GetNodeMapByIDTx(id string) (*sm.NodeMap, error) {
	if id == "" {
		t.LogAlways("Error: GetNodeMapByIDTx(): xname was empty")
		return nil, ErrHMSDSArgMissing
	}
	// Perform corresponding query on DB
	nnms, err := t.queryNodeMap("GetNodeMapByIDTx",
		getNodeMapByIDQuery, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(nnms) == 0 {
		t.Log(LOG_INFO, "Info: GetNodeMapByIDTx(%s) matched 0.",
			xnametypes.NormalizeHMSCompID(id))
		return nil, nil
	} else if len(nnms) > 1 {
		t.LogAlways("WARNING: GetNodeMapByIDTx(%s): matched >1!",
			xnametypes.NormalizeHMSCompID(id))
	}
	return nnms[0], nil
}

// Look up ALL Node->NID Mappings (in transaction).
func (t *hmsdbPgTx) GetNodeMapsAllTx() ([]*sm.NodeMap, error) {
	// Perform corresponding query on DB
	nnms, err := t.queryNodeMap("GetNodeMapsAllTx",
		getNodeMapsAllQuery)
	if err != nil {
		return nil, err
	}
	// Query succeeded.
	t.Log(LOG_INFO, "Info: GetNodeMapByIDTx() matched %d.", len(nnms))
	return nnms, nil
}

// Insert Node->NID Mapping into database, updating it if it exists.
func (t *hmsdbPgTx) InsertNodeMapTx(m *sm.NodeMap) error {
	if m == nil {
		t.LogAlways("Error: GetNodeMapByIDTx(): Component was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	rawNID := 0
	if m.NID > 0 {
		rawNID = m.NID
	}
	nodeInfoJSON, err := json.Marshal(m.NodeInfo)
	if err != nil {
		// This should never fail
		t.LogAlways("InsertNodeMapTx: decode Details: %s", err)
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("InsertNodeMapTx",
		insertPgNodeMapQuery)
	if err != nil {
		return err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(m.ID)

	// Perform insert
	_, err = stmt.ExecContext(t.ctx,
		&normID,
		&rawNID,
		&m.Role,
		&m.SubRole,
		&nodeInfoJSON)
	if err != nil {
		t.LogAlways("Error: InsertNodeMapTx(): stmt.Exec: %s", err)
		if IsPgDuplicateKeyErr(err) == true {
			// Key already exists, user error. Set false and clear error.
			return ErrHMSDSDuplicateKey
		}
		return err
	}
	return nil
}

// Delete Node NID Mapping entry with matching xname id from database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteNodeMapByIDTx(id string) (bool, error) {
	if id == "" {
		t.LogAlways("Error: DeleteNodeMapByIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteNodeMapByIDTx",
		deleteNodeMapByIDQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		t.LogAlways("Error: DeleteNodeMapByIDTx(%s): stmt.Exec: %s",
			xnametypes.NormalizeHMSCompID(id), err)
		return false, err
	}

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all Node NID Mapping entries from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteNodeMapsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteNodeMapsAllTx",
		deleteNodeMapsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteNodeMapsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - Power Mapping queries
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more Power Mapping rows in
// the result.
func (t *hmsdbPgTx) queryPowerMap(qname, query string, args ...interface{}) ([]*sm.PowerMap, error) {
	t.Log(LOG_DEBUG, "Debug: %s(%v) starting....", qname, args)

	stmt, err := t.conditionalPrepare(qname, query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.QueryContext(t.ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ms := make([]*sm.PowerMap, 0, 1)
	i := 0
	for rows.Next() {
		m, err := t.hdb.scanPowerMap(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return ms, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, m)
		ms = append(ms, m)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d entries.", qname, args, len(ms))
	return ms, err
}

// Look up one Power Mapping by id, i.e. node xname (in transaction).
func (t *hmsdbPgTx) GetPowerMapByIDTx(id string) (*sm.PowerMap, error) {
	if id == "" {
		t.LogAlways("Error: GetPowerMapByIDTx(): xname was empty")
		return nil, ErrHMSDSArgMissing
	}
	// Perform corresponding query on DB
	ms, err := t.queryPowerMap("GetPowerMapByIDTx",
		getPowerMapByIDQuery, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(ms) == 0 {
		t.Log(LOG_INFO, "Info: GetPowerMapByIDTx(%s) matched 0.",
			xnametypes.NormalizeHMSCompID(id))
		return nil, nil
	} else if len(ms) > 1 {
		t.LogAlways("WARNING: GetPowerMapByIDTx(%s): matched >1!",
			xnametypes.NormalizeHMSCompID(id))
	}
	return ms[0], nil
}

// Look up ALL Power Mappings (in transaction).
func (t *hmsdbPgTx) GetPowerMapsAllTx() ([]*sm.PowerMap, error) {
	// Perform corresponding query on DB
	ms, err := t.queryPowerMap("GetPowerMapsAllTx",
		getPowerMapsAllQuery)
	if err != nil {
		return nil, err
	}
	// Query succeeded.
	t.Log(LOG_INFO, "Info: GetPowerMapByIDTx() matched %d.", len(ms))
	return ms, nil
}

// Insert Power Mapping into database, updating it if it exists.
func (t *hmsdbPgTx) InsertPowerMapTx(m *sm.PowerMap) error {
	if m == nil {
		t.LogAlways("Error: GetPowerMapByIDTx(): Component was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("InsertPowerMapTx",
		insertPgPowerMapQuery)
	if err != nil {
		return err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(m.ID)

	normPwrIds := make([]string, 0, len(m.PoweredBy))
	for _, pwrId := range m.PoweredBy {
		normPwrIds = append(normPwrIds, xnametypes.NormalizeHMSCompID(pwrId))
	}

	// Perform insert
	_, err = stmt.ExecContext(t.ctx,
		&normID,
		pq.Array(normPwrIds))
	if err != nil {
		t.LogAlways("Error: InsertPowerMapTx(): stmt.Exec: %s", err)
		return err
	}
	return nil
}

// Delete Power Mapping entry with matching xname id from database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeletePowerMapByIDTx(id string) (bool, error) {
	if id == "" {
		t.LogAlways("Error: DeletePowerMapByIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeletePowerMapByIDTx",
		deletePowerMapByIDQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		t.LogAlways("Error: DeletePowerMapByIDTx(%s): stmt.Exec: %s",
			xnametypes.NormalizeHMSCompID(id), err)
		return false, err
	}

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all Power Mapping entries from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeletePowerMapsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeletePowerMapsAllTx",
		deletePowerMapsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeletePowerMapsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - HWInventory queries
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more HWInvByLoc rows in
// the result.  It also pairs the data with the matching HWInvByFRU if the
// xname is populated.   It does not do any nesting and performs only one
// query.
func (t *hmsdbPgTx) queryHWInvByLoc(qname, query string, args ...interface{}) ([]*sm.HWInvByLoc, error) {
	rows, err := t.getRowsForQuery(qname, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwlocs := make([]*sm.HWInvByLoc, 0, 1)
	i := 0
	for rows.Next() {
		hwloc, err := t.hdb.scanHwInvByLocWithFRU(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return hwlocs, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, hwloc)
		hwlocs = append(hwlocs, hwloc)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d hwinv items.", qname, args, len(hwlocs))
	return hwlocs, err
}

// Back end for all queries that produce one or more HWInvByLoc rows in
// the result.  It also pairs the data with the matching HWInvByFRU if the
// xname is populated.   It does not do any nesting and performs only one
// query.
func (t *hmsdbPgTx) queryHWInvByFRU(qname, query string, args ...interface{}) ([]*sm.HWInvByFRU, error) {
	rows, err := t.getRowsForQuery(qname, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwfrus := make([]*sm.HWInvByFRU, 0, 1)
	i := 0
	for rows.Next() {
		hwfru, err := t.hdb.scanHwInvByFRU(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return hwfrus, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, hwfru)
		hwfrus = append(hwfrus, hwfru)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d hwinv items.",
		qname, args, len(hwfrus))
	return hwfrus, err
}

// Get HWInvByLoc by primary key (xname), i.e. a single entry. It also pairs
// the data with the matching HWInvByFRU if the xname is populated.
func (t *hmsdbPgTx) GetHWInvByLocIDTx(id string) (*sm.HWInvByLoc, error) {
	if id == "" {
		t.LogAlways("Error: GetHWInvByLocIDTx(): xname was empty")
		return nil, ErrHMSDSArgNil
	}
	// Perform corresponding query on DB
	hwlocs, err := t.queryHWInvByLoc("GetHWInvByLocIDTx",
		getHWInvByLocWithFRUByIDQuery, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(hwlocs) == 0 {
		t.Log(LOG_INFO, "Info: GetHWInvByLocIDTx(%s) matched no entry.",
			xnametypes.NormalizeHMSCompID(id))
		return nil, nil
	} else if len(hwlocs) > 1 {
		t.LogAlways("Warning: GetHWInvByLocIDTx(%s): multiple entries!.",
			xnametypes.NormalizeHMSCompID(id))
	}
	return hwlocs[0], nil
}

// Get HWInvByLoc by primary key (xname) for all entries in the system.
// It also pairs the data with the matching HWInvByFRU if the xname is
// populated. (In transaction)
func (t *hmsdbPgTx) GetHWInvByLocAllTx() ([]*sm.HWInvByLoc, error) {
	// Perform corresponding query on DB
	hwlocs, err := t.queryHWInvByLoc("GetHWInvByLocAllTx",
		getHWInvByLocWithFRUAllQuery)
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(hwlocs) == 0 {
		t.Log(LOG_INFO, "Info: GetHWInvByLocAllTx() matched no entry.")
		return nil, nil
	}
	return hwlocs, nil
}

// Get a single HW-inventory-by-FRU entry by its FRUID. (in transaction).
func (t *hmsdbPgTx) GetHWInvByFRUIDTx(fruid string) (*sm.HWInvByFRU, error) {
	if fruid == "" {
		t.LogAlways("Error: GetHWInvByFRUIDTx(): xname was empty")
		return nil, ErrHMSDSArgNil
	}
	// Perform corresponding query on DB
	hwfrus, err := t.queryHWInvByFRU("GetHWInvByFRUIDTx",
		getHWInvByFRUByFRUIDQuery, fruid)
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(hwfrus) == 0 {
		t.Log(LOG_INFO, "Info: GetHWInvByFRUIDTx(%s) matched no entry.", fruid)
		return nil, nil
	} else if len(hwfrus) > 1 {
		t.LogAlways("Warning: GetHWInvByFRUIDTx(%s): multiple entries!.", fruid)
	}
	return hwfrus[0], nil
}

// Get all HW-inventory-by-FRU entries. (in transaction).
func (t *hmsdbPgTx) GetHWInvByFRUAllTx() ([]*sm.HWInvByFRU, error) {
	// Perform corresponding query on DB
	hwfrus, err := t.queryHWInvByFRU("GetHWInvByFRUAllTx",
		getHWInvByFRUAllQuery)
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(hwfrus) == 0 {
		t.Log(LOG_INFO, "Info: GetHWInvByFRUAllTx() matched no entry.")
		return nil, nil
	}
	return hwfrus, nil
}

// Insert or update HWInventoryByLocation struct (in transaction)
// If PopulatedFRU is present, only the FRUID is added to the database.  If
// it is not, this effectively "depopulates" the given location.
// The actual HWInventoryByFRU struct must be stored FIRST using the
// corresponding function (presumably within the same transaction).
func (t *hmsdbPgTx) InsertHWInvByLocTx(hl *sm.HWInvByLoc) error {
	if hl == nil {
		t.LogAlways("Error:  InsertHWInvByLocTx(): Component was nil.")
		return ErrHMSDSArgNil
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("InsertHWInvByLocTx", insertPgHWInvByLocQuery)
	if err != nil {
		return err
	}
	// If a location is empty, the fru_id field will be NULL.
	var fruIdPtr *string = nil
	if hl.PopulatedFRU != nil {
		if hl.PopulatedFRU.FRUID == "" {
			t.LogAlways("WARNING: InsertHWInvByLocTx(): FRUID is empty")
		} else {
			fruIdPtr = &hl.PopulatedFRU.FRUID
		}
	}
	infoJSON, err := hl.EncodeLocationInfo()
	if err != nil {
		t.LogAlways("Error: InsertHWInvByLocTx(): EncodeLocationInfo: %s", err)
		return err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(hl.ID)

	// Get the parent node xname for use with partition queries. Components under nodes
	// (processors, memory, etc.) get the parent_node set to the node above them. For
	// all others parent_node == id
	pnID := normID
	// Don't bother checking if the component isn't under a node
	if strings.Contains(pnID, "n") {
		for xnametypes.GetHMSType(pnID) != xnametypes.Node {
			pnID = xnametypes.GetHMSCompParent(pnID)
			// This is to catch components that are not under nodes
			// but have 'n' in the xname.
			if pnID == "" {
				pnID = normID
				break
			}
		}
	}

	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&normID,
		&hl.Type,
		&hl.Ordinal,
		&hl.Status,
		&pnID,
		&infoJSON,
		fruIdPtr)
	if err != nil {
		t.LogAlways("Error: InsertHWInvByLocTx(): stmt.Exec: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: InsertHWInvByLocTx(): - %v", res)
	return nil
}

// Insert or update HWInventoryByLocation struct (in transaction)
// If PopulatedFRU is present, only the FRUID is added to the database.  If
// it is not, this effectively "depopulates" the given location.
// The actual HWInventoryByFRU struct must be stored FIRST using the
// corresponding function (presumably within the same transaction).
func (t *hmsdbPgTx) BulkInsertHWInvByLocTx(hls []*sm.HWInvByLoc) error {
	var fruId sql.NullString
	if len(hls) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(hwInvLocTable).
		Columns(hwInvLocCols...)

	for _, hl := range hls {
		// Normalize key
		normID := xnametypes.NormalizeHMSCompID(hl.ID)

		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[normID]; ok {
			continue
		} else {
			valueMap[normID] = true
		}

		// If a location is empty, the fru_id field will be NULL.
		if hl.PopulatedFRU != nil {
			if hl.PopulatedFRU.FRUID == "" {
				t.LogAlways("WARNING: BulkInsertHWInvByLocTx(): FRUID is empty")
				fruId.Valid = false
			} else {
				fruId.String = hl.PopulatedFRU.FRUID
				fruId.Valid = true
			}
		} else {
			fruId.Valid = false
		}
		infoJSON, err := hl.EncodeLocationInfo()
		if err != nil {
			t.LogAlways("Error: BulkInsertHWInvByLocTx(): EncodeLocationInfo: %s", err)
			return err
		}

		// Get the parent node xname for use with partition queries. Components under nodes
		// (processors, memory, etc.) get the parent_node set to the node above them. For
		// all others parent_node == id
		pnID := normID
		// Don't bother checking if the component isn't under a node
		if strings.Contains(pnID, "n") {
			for xnametypes.GetHMSType(pnID) != xnametypes.Node {
				pnID = xnametypes.GetHMSCompParent(pnID)
				// This is to catch components that are not under nodes
				// but have 'n' in the xname.
				if pnID == "" {
					pnID = normID
					break
				}
			}
		}

		// Set fields for the INSERT
		query = query.Values(
			normID,
			hl.Type,
			hl.Ordinal,
			hl.Status,
			pnID,
			infoJSON,
			fruId)
	}
	query = query.Suffix("ON CONFLICT(" + hwInvLocIdCol + ") DO UPDATE SET " +
		hwInvLocOrdCol + " = EXCLUDED." + hwInvLocOrdCol + ", " +
		hwInvLocStatusCol + " = EXCLUDED." + hwInvLocStatusCol + ", " +
		hwInvLocNodeCol + " = EXCLUDED." + hwInvLocNodeCol + ", " +
		hwInvLocLocInfoCol + " = EXCLUDED." + hwInvLocLocInfoCol + ", " +
		hwInvLocFruIdCol + " = EXCLUDED." + hwInvLocFruIdCol)

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: BulkInsertHWInvByLocTx(): Query: %s - With args: %v", qStr, qArgs)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: BulkInsertHWInvByLocTx(): ExecContext: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: BulkInsertHWInvByLocTx() - %s", res)
	return nil
}

// Insert or update HWInventoryByFRU struct (in transaction)
func (t *hmsdbPgTx) InsertHWInvByFRUTx(hf *sm.HWInvByFRU) error {
	if hf == nil {
		t.LogAlways("Error:  InsertHWInvByFRUTx(): Component was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("InsertHWInvByFRUTx", insertPgHWInvByFRUQuery)
	if err != nil {
		return err
	}
	infoJSON, err := hf.EncodeFRUInfo()
	if err != nil {
		t.LogAlways("Error: InsertHWInvByFRUTx(): EncodeLocationInfo: %s", err)
		return err
	}
	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&hf.FRUID,
		&hf.Type,
		&hf.Subtype,
		&infoJSON)
	if err != nil {
		t.LogAlways("Error: InsertHWInvByFRUTx(): stmt.Exec: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: InsertHWInvByFRUTx(): - %v", res)
	return nil
}

// Insert or update HWInventoryByFRU structs (in transaction)
func (t *hmsdbPgTx) BulkInsertHWInvByFRUTx(hfs []*sm.HWInvByFRU) error {
	if len(hfs) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(hwInvFruTable).
		Columns(hwInvFruTblCols...)

	for _, hf := range hfs {
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[hf.FRUID]; ok {
			continue
		} else {
			valueMap[hf.FRUID] = true
		}

		infoJSON, err := hf.EncodeFRUInfo()
		if err != nil {
			t.LogAlways("Error: BulkInsertHWInvByFRUTx(): EncodeLocationInfo: %s", err)
			return err
		}

		// Set fields for the INSERT
		query = query.Values(
			hf.FRUID,
			hf.Type,
			hf.Subtype,
			infoJSON)
	}
	query = query.Suffix("ON CONFLICT(" + hwInvFruTblIdCol + ") DO UPDATE SET " +
		hwInvFruTblSubTypeCol + " = EXCLUDED." + hwInvFruTblSubTypeCol + ", " +
		hwInvFruTblInfoCol + " = EXCLUDED." + hwInvFruTblInfoCol)

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: BulkInsertHWInvByFRUTx(): Query: %s - With args: %v", qStr, qArgs)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: BulkInsertHWInvByFRUTx(): ExecContext: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: BulkInsertHWInvByFRUTx() - %s", res)
	return nil
}

// Delete HWInvByLoc entry with matching FRU ID from database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteHWInvByLocIDTx(id string) (bool, error) {
	if id == "" {
		t.LogAlways("Error: DeleteHWInvByLocIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteHWInvByLocIDTx",
		deleteHWInvByLocIDQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		t.LogAlways("Error: DeleteHWInvByLocIDTx(%s): stmt.Exec: %s",
			xnametypes.NormalizeHMSCompID(id), err)
		return false, err
	}
	t.Log(LOG_INFO, "Info: DeleteHWInvByLocIDTx(%s) - %s",
		xnametypes.NormalizeHMSCompID(id), res)

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all HWInvByLoc entries from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteHWInvByLocsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteHWInvByLocsAllTx",
		deleteHWInvByLocsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteHWInvByLocsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: DeleteHWInvByLocsAllTx() - %s", res)

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

// Delete HWInvByFRU entry with matching FRU ID from database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteHWInvByFRUIDTx(fruid string) (bool, error) {
	if fruid == "" {
		t.LogAlways("Error: DeleteHWInvByFRUIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteHWInvByFRUIDTx",
		deleteHWInvByFRUIDQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, fruid)
	if err != nil {
		t.LogAlways("Error: DeleteHWInvByFRUIDTx(%s): stmt.Exec: %s",
			fruid, err)
		return false, err
	}
	t.Log(LOG_INFO, "Info: DeleteHWInvByFRUIDTx(%s) - %s", res, fruid)

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all HWInvByFRU entries from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteHWInvByFRUsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteHWInvByFRUsAllTx",
		deleteHWInvByFRUsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteHWInvByFRUsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: DeleteHWInvByFRUsAllTx() - %s", res)

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

// Get some or all Hardware Inventory entries with filtering
// options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct.
func (t *hmsdbPgTx) GetHWInvByLocQueryFilterTx(f_opts ...HWInvLocFiltFunc) ([]*sm.HWInvByLoc, error) {
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	query, err := getHWInvByLocQuery(f_opts...)
	if err != nil {
		return nil, err
	}

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: GetHWInvByLoc(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwlocs := make([]*sm.HWInvByLoc, 0, 1)
	i := 0
	for rows.Next() {
		hwloc, err := t.hdb.scanHwInvByLocWithFRU(rows)
		if err != nil {
			t.LogAlways("Error: GetHWInvByLoc(): Scan failed: %s", err)
			return hwlocs, err
		}
		t.Log(LOG_DEBUG, "Debug: GetHWInvByLoc() scanned[%d]: %v", i, hwloc)
		hwlocs = append(hwlocs, hwloc)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: GetHWInvByLoc() returned %d hwinv items.", len(hwlocs))
	return hwlocs, err
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - HWInventoryHistory queries
//
/////////////////////////////////////////////////////////////////////////////

// Get hardware history for some or all Hardware Inventory entries with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
// (in transaction)
func (t *hmsdbPgTx) GetHWInvHistFilterTx(f_opts ...HWInvHistFiltFunc) ([]*sm.HWInvHist, error) {
	// Parse the filter options
	f := new(HWInvHistFilter)
	for _, opts := range f_opts {
		opts(f)
	}

	query := sq.Select(addAliasToCols(hwInvHistAlias, hwInvHistCols, hwInvHistCols)...).
		From(hwInvHistTable + " " + hwInvHistAlias)
	if len(f.ID) > 0 {
		idCol := hwInvHistAlias + "." + hwInvHistIdCol
		query = query.Where(sq.Eq{idCol: f.ID})
	}
	if len(f.FruId) > 0 {
		fruIdCol := hwInvHistAlias + "." + hwInvHistFruIdCol
		query = query.Where(sq.Eq{fruIdCol: f.FruId})
	}
	if len(f.EventType) > 0 {
		evtCol := hwInvHistAlias + "." + hwInvHistEventTypeCol
		tArgs := []string{}
		for _, evt := range f.EventType {
			normEvt := sm.VerifyNormalizeHWInvHistEventType(evt)
			if normEvt == "" {
				return nil, ErrHMSDSArgBadHWInvHistEventType
			}
			tArgs = append(tArgs, normEvt)
		}
		query = query.Where(sq.Eq{evtCol: tArgs})
	}
	if f.StartTime != "" {
		tsCol := hwInvHistAlias + "." + hwInvHistTimestampCol
		start, err := time.Parse(time.RFC3339, f.StartTime)
		if err != nil {
			return nil, ErrHMSDSArgBadTimeFormat
		}
		query = query.Where(sq.Gt{tsCol: start})
	}
	if f.EndTime != "" {
		tsCol := hwInvHistAlias + "." + hwInvHistTimestampCol
		end, err := time.Parse(time.RFC3339, f.EndTime)
		if err != nil {
			return nil, ErrHMSDSArgBadTimeFormat
		}
		query = query.Where(sq.Lt{tsCol: end})
	}
	query = query.OrderBy("timestamp ASC")

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: GetHWInvHistFilterTx(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwhists := make([]*sm.HWInvHist, 0, 1)
	i := 0
	for rows.Next() {
		hwhist, err := t.hdb.scanHwInvHist(rows)
		if err != nil {
			t.LogAlways("Error: GetHWInvHistFilterTx(): Scan failed: %s", err)
			return hwhists, err
		}
		t.Log(LOG_DEBUG, "Debug: GetHWInvHistFilterTx() scanned[%d]: %v", i, hwhist)
		hwhists = append(hwhists, hwhist)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: GetHWInvHistFilterTx() returned %d hwinvhist items.", len(hwhists))
	return hwhists, err
}

func (t *hmsdbPgTx) GetHWInvHistLastEventsTx(ids []string) ([]*sm.HWInvHist, error) {

	query := sq.Select(addAliasToCols(hwInvHistAlias, hwInvHistCols, hwInvHistCols)...).
		From(hwInvHistTable + " " + hwInvHistAlias)
	query = query.Options("DISTINCT ON (", hwInvHistIdColAlias, ")")
	if len(ids) > 0 {
		query = query.Where(sq.Eq{hwInvHistIdColAlias: ids})
	}

	query = query.OrderBy("" + hwInvHistIdColAlias + ", " + hwInvHistTimestampColAlias + " DESC")

	// Execute
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: GetHWInvHistLastEventsTx(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hwhists := make([]*sm.HWInvHist, 0, 1)
	i := 0
	for rows.Next() {
		hwhist, err := t.hdb.scanHwInvHist(rows)
		if err != nil {
			t.LogAlways("Error: GetHWInvHistLastEventsTx(): Scan failed: %s", err)
			return hwhists, err
		}
		t.Log(LOG_DEBUG, "Debug: GetHWInvHistLastEventsTx() scanned[%d]: %v", i, hwhist)
		hwhists = append(hwhists, hwhist)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: GetHWInvHistLastEventsTx() returned %d hwinvhist items.", len(hwhists))
	return hwhists, err
}

// Insert a HWInventoryHistory struct (in transaction)
func (t *hmsdbPgTx) InsertHWInvHistTx(hh *sm.HWInvHist) error {
	var err error
	if hh == nil {
		t.LogAlways("Error: InsertHWInvHistTx(): Struct was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// Normalize and verify fields (note these functions track if this
	// has been done and only does each once.)
	eventType := sm.VerifyNormalizeHWInvHistEventType(hh.EventType)
	if eventType == "" {
		return ErrHMSDSArgBadHWInvHistEventType
	}
	loc := xnametypes.VerifyNormalizeCompID(hh.ID)
	if loc == "" {
		return ErrHMSDSArgBadID
	}
	if hh.FruId == "" {
		return ErrHMSDSArgMissing
	}

	// Generate query
	query := sq.Insert(hwInvHistTable).
		Columns(hwInvHistColsNoTS...).
		Values(loc, hh.FruId, eventType)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

// Insert an array of HWInventoryHistory entries. (in transaction)
// If a duplicate is present return an error.
func (t *hmsdbPgTx) InsertHWInvHistsTx(hhs []*sm.HWInvHist) error {
	var err error
	if len(hhs) == 0 {
		// Nothing to do
		return nil
	}

	// Generate query
	query := sq.Insert(hwInvHistTable).
		Columns(hwInvHistColsNoTS...)

	for _, hh := range hhs {
		// Normalize and verify fields (note these functions track if this
		// has been done and only does each once.)
		eventType := sm.VerifyNormalizeHWInvHistEventType(hh.EventType)
		if eventType == "" {
			return ErrHMSDSArgBadHWInvHistEventType
		}
		loc := xnametypes.VerifyNormalizeCompID(hh.ID)
		if loc == "" {
			return ErrHMSDSArgBadID
		}
		if hh.FruId == "" {
			return ErrHMSDSArgMissing
		}
		query = query.Values(loc, hh.FruId, eventType)
	}

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: GetHWInvHistLastEventsTx(): Query: %s - With args: %v", qStr, qArgs)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - RedfishEndpoint queries
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more RedfishEndpoint rows in
// the result.
func (t *hmsdbPgTx) queryRedfishEndpoint(qname, query string, args ...interface{}) ([]*sm.RedfishEndpoint, error) {
	rows, err := t.getRowsForQuery(qname, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	eps := make([]*sm.RedfishEndpoint, 0, 1)
	i := 0
	for rows.Next() {
		ep, err := t.hdb.scanRedfishEndpoint(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return eps, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, ep)
		eps = append(eps, ep)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d EPs.", qname, args, len(eps))
	return eps, err
}

// Build filter query for RedfishEndpoints using filter functions and
// then return the set of matching components as an array.
//
// NOTE: Most args allow negated arguments, i.e. "!x0c0s0b0", so be careful about
// passing in user data if the query should only return a single result. RFE_ID
// does not.
func (t *hmsdbPgTx) GetRFEndpointsTx(f_opts ...RedfishEPFiltFunc) ([]*sm.RedfishEndpoint, error) {

	f := new(RedfishEPFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	return t.GetRFEndpointsFilterTx(f)
}

// Get RedfishEndpoint by ID (xname), i.e. a single entry.
func (t *hmsdbPgTx) GetRFEndpointByIDTx(id string) (*sm.RedfishEndpoint, error) {
	if id == "" {
		t.LogAlways("Error: GetRFEndpointByIDTx(): xname was empty")
		return nil, ErrHMSDSArgNil
	}
	// Perform corresponding query on DB
	eps, err := t.queryRedfishEndpoint("GetRFEndpointByIDTx",
		getRFEndpointByIDQuery, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		return nil, err
	}
	// Query succeeded.  There should be at most 1 row returned...
	if len(eps) == 0 {
		t.Log(LOG_INFO, "Info: GetRFEndpointByIDTx(%s) matched no entry.", id)
		return nil, nil
	} else if len(eps) > 1 {
		t.LogAlways("Warning: GetRFEndpointByIDTx(%s): multiple entries!.", id)
	}
	return eps[0], nil
}

// Get all RedfishEndpoints in system
func (t *hmsdbPgTx) GetRFEndpointsAllTx() ([]*sm.RedfishEndpoint, error) {
	return t.GetRFEndpointsTx(RFE_From("GetRFEndpointsAllTx"))
}

// Get some or all RedfishEndpoints in system (in transaction), with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (t *hmsdbPgTx) GetRFEndpointsFilterTx(f *RedfishEPFilter) ([]*sm.RedfishEndpoint, error) {
	var filterQuery string
	var reps []*sm.RedfishEndpoint
	var err error
	label := "GetRFEndpointsFilterTx" // Use for errors and debug output

	// If no filter provided, just get everything.  Otherwise use it
	// to create a custom WHERE... string.
	if f == nil {
		filterQuery = getRFEndpointsAllQuery
		// Perform corresponding query on DB
		reps, err = t.queryRedfishEndpoint(label, filterQuery)
		if err != nil {
			return nil, err
		}
	} else {
		if f.label != "" {
			label = f.label
		}
		filterQuery, args, err := buildRedfishEPQuery(getRFEndpointPrefix, f)
		if err != nil {
			return nil, err
		}
		// Perform corresponding query on DB
		reps, err = t.queryRedfishEndpoint(label, filterQuery, args...)
		if err != nil {
			return nil, err
		}
	}
	// Query succeeded.
	if len(reps) == 0 {
		t.Log(LOG_INFO, "Info: %s(): no matches: %s", label, filterQuery)
		return reps, nil
	}
	return reps, nil
}

// Insert new RedfishEndpoint into database. Does not insert any
// ComponentEndpoint children.(In transaction.)
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (t *hmsdbPgTx) InsertRFEndpointTx(ep *sm.RedfishEndpoint) error {
	if ep == nil {
		t.LogAlways("Error: InsertRFEndpointTx(): EP was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("InsertRFEndpointTx",
		insertPgRFEndpointQuery)
	if err != nil {
		return err
	}
	discInfoJSON, err := json.Marshal(ep.DiscInfo)
	if err != nil {
		// This should never fail
		t.LogAlways(" InsertRFEndpointTx: decode DiscoveryInfo: %s", err)
	}
	// Ensure endpoint name is normalized and valid
	normID := xnametypes.VerifyNormalizeCompID(ep.ID)
	if normID == "" {
		t.LogAlways("InsertRFEndpointTx(%s): %s", ep.ID, ErrHMSDSArgBadID)
		return ErrHMSDSArgBadID
	}

	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&normID,
		&ep.Type,
		&ep.Name,
		&ep.Hostname,
		&ep.Domain,
		&ep.FQDN,
		&ep.Enabled,
		&ep.UUID,
		&ep.User,
		&ep.Password,
		&ep.UseSSDP,
		&ep.MACRequired,
		&ep.MACAddr,
		&ep.IPAddr,
		&ep.RediscOnUpdate,
		&ep.TemplateID,
		&discInfoJSON)
	if err != nil {
		t.LogAlways("Error: InsertRFEndpointTx(): stmt.Exec: %s", err)
		if IsPgDuplicateKeyErr(err) == true {
			// Key already exists, user error. Set false and clear error.
			return ErrHMSDSDuplicateKey
		}
		// Unexpected internal error.
		return err
	}
	t.Log(LOG_INFO, "Info: InsertRFEndpointTx() - %s", res)
	return nil
}

// Insert new RedfishEndpoints into database. Does not insert any
// ComponentEndpoint children.(In transaction.)
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (t *hmsdbPgTx) InsertRFEndpointsTx(eps []*sm.RedfishEndpoint) error {
	if len(eps) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(rfEPsTable).
		Columns(rfEPsAllCols...)

	for _, ep := range eps {
		// Ensure endpoint name is normalized and valid
		normID := xnametypes.VerifyNormalizeCompID(ep.ID)
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[normID]; ok {
			continue
		} else {
			valueMap[normID] = true
		}
		discInfoJSON, err := json.Marshal(ep.DiscInfo)
		if err != nil {
			// This should never fail
			t.LogAlways(" InsertRFEndpointsTx: decode DiscoveryInfo: %s", err)
		}
		if normID == "" {
			t.LogAlways("InsertRFEndpointsTx(%s): %s", ep.ID, ErrHMSDSArgBadID)
			return ErrHMSDSArgBadID
		}

		// Set fields for the INSERT
		query = query.Values(
			normID,
			ep.Type,
			ep.Name,
			ep.Hostname,
			ep.Domain,
			ep.FQDN,
			ep.Enabled,
			ep.UUID,
			ep.User,
			ep.Password,
			ep.UseSSDP,
			ep.MACRequired,
			ep.MACAddr,
			ep.IPAddr,
			ep.RediscOnUpdate,
			ep.TemplateID,
			discInfoJSON)
	}

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: InsertRFEndpointsTx(): Query: %s - With args: %v", qStr, qArgs)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: InsertRFEndpointsTx(): ExecContext: %s", err)
		if IsPgDuplicateKeyErr(err) == true {
			// Key already exists, user error. Set false and clear error.
			return ErrHMSDSDuplicateKey
		}
		// Unexpected internal error.
		return err
	}
	t.Log(LOG_INFO, "Info: InsertRFEndpointsTx() - %s", res)
	return nil
}

// Update RedfishEndpoint already in DB. Does not update any
// ComponentEndpoint children. (In transaction.)
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
func (t *hmsdbPgTx) UpdateRFEndpointTx(ep *sm.RedfishEndpoint) (bool, error) {
	if ep == nil {
		t.LogAlways("Error: UpdateRFEndpointTx(): EP was nil.")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("UpdateRFEndpointTx",
		updatePgRFEndpointQuery)
	if err != nil {
		return false, err
	}
	discInfoJSON, err := json.Marshal(ep.DiscInfo)
	if err != nil {
		// This should never fail
		t.LogAlways("UpdateRFEndpointTx: decode DiscoveryInfo: %s", err)
	}
	// Normalized key
	normID := xnametypes.NormalizeHMSCompID(ep.ID)

	// Perform update
	res, err := stmt.ExecContext(t.ctx,
		&ep.Type,
		&ep.Name,
		&ep.Hostname,
		&ep.Domain,
		&ep.FQDN,
		&ep.Enabled,
		&ep.UUID,
		&ep.User,
		&ep.Password,
		&ep.UseSSDP,
		&ep.MACRequired,
		&ep.MACAddr,
		&ep.IPAddr,
		&ep.RediscOnUpdate,
		&ep.TemplateID,
		&discInfoJSON,
		&normID) // Key
	if err != nil {
		t.LogAlways("Error: UpdateRFEndpointTx(): stmt.Exec: %s", err)
		if IsPgDuplicateKeyErr(err) == true {
			// Key already exists, user error. Set false and clear error.
			return false, ErrHMSDSDuplicateKey
		}
		return false, err
	}
	t.Log(LOG_INFO, "Info: UpdateRFEndpointTx() - %s", res)
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Update RedfishEndpoint already in DB. Does not update any
// ComponentEndpoint children. (In transaction.)
// If ID or FQDN already exists, return ErrHMSDSDuplicateKey
func (t *hmsdbPgTx) UpdateRFEndpointsTx(eps []*sm.RedfishEndpoint) ([]*sm.RedfishEndpoint, error) {
	newEPs := make([]*sm.RedfishEndpoint, 0, 1)
	if len(eps) == 0 {
		return newEPs, nil
	}
	if !t.IsConnected() {
		return newEPs, ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	// Make the column name a sq.Sqlizer so sq will set it as a column name and not a value.
	query := sq.Update(rfEPsTable+" r").
		Set(rfEPsTypeCol, sq.Expr(rfEPsTypeColAlias)).
		Set(rfEPsNameCol, sq.Expr(rfEPsNameColAlias)).
		Set(rfEPsHostnameCol, sq.Expr(rfEPsHostnameColAlias)).
		Set(rfEPsDomainCol, sq.Expr(rfEPsDomainColAlias)).
		Set(rfEPsFQDNCol, sq.Expr(rfEPsFQDNColAlias)).
		Set(rfEPsEnabledCol, sq.Expr(rfEPsEnabledColAlias)).
		Set(rfEPsUUIDCol, sq.Expr(rfEPsUUIDColAlias)).
		Set(rfEPsUserCol, sq.Expr(rfEPsUserColAlias)).
		Set(rfEPsPasswordCol, sq.Expr(rfEPsPasswordColAlias)).
		Set(rfEPsUseSSDPCol, sq.Expr(rfEPsUseSSDPColAlias)).
		Set(rfEPsMacRequiredCol, sq.Expr(rfEPsMacRequiredColAlias)).
		Set(rfEPsMacAddrCol, sq.Expr(rfEPsMacAddrColAlias)).
		Set(rfEPsIPAddrCol, sq.Expr(rfEPsIPAddrColAlias)).
		Set(rfEPsRediscOnUpdateCol, sq.Expr(rfEPsRediscOnUpdateColAlias)).
		Set(rfEPsTemplateIDCol, sq.Expr(rfEPsTemplateIDColAlias)).
		Set(rfEPsDiscInfoCol, sq.Expr(rfEPsDiscInfoColAlias))

	// sq doesn't have a way to add a FROM statement to an UPDATE.
	// We'll just construct the 2nd half of the query manually and
	// add it as a suffix.
	args := make([]interface{}, 0, 1)
	valStr := ""
	for i, ep := range eps {
		// Normalized key
		normID := xnametypes.NormalizeHMSCompID(ep.ID)
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[normID]; ok {
			continue
		} else {
			valueMap[normID] = true
		}

		discInfoJSON, err := json.Marshal(ep.DiscInfo)
		if err != nil {
			// This should never fail
			t.LogAlways("UpdateRFEndpointsTx: decode DiscoveryInfo: %s", err)
		}
		// Add the values to our values table
		if i == 0 {
			valStr += "(?,?,?,?,?,?,?::BOOL,?,?,?,?::BOOL,?::BOOL,?,?,?::BOOL,?,?::JSON)"
		} else {
			valStr += ",(?,?,?,?,?,?,?::BOOL,?,?,?,?::BOOL,?::BOOL,?,?,?::BOOL,?,?::JSON)"
		}
		args = append(args,
			normID,
			ep.Type,
			ep.Name,
			ep.Hostname,
			ep.Domain,
			ep.FQDN,
			ep.Enabled,
			ep.UUID,
			ep.User,
			ep.Password,
			ep.UseSSDP,
			ep.MACRequired,
			ep.MACAddr,
			ep.IPAddr,
			ep.RediscOnUpdate,
			ep.TemplateID,
			discInfoJSON)
	}
	// This FROM statement builds us a values table to pull update values
	// from so an update with different values for each row can be done.
	colStr := strings.Join(rfEPsAllCols, ",")
	aliasColStr := strings.Join(rfEPsAllCols, ",r.")
	qstr := "FROM (VALUES " + valStr + ") AS " + rfEPsAlias +
		"(" + colStr +
		") WHERE r." + rfEPsIdCol + " = " + rfEPsIdColAlias +
		" RETURNING r." + aliasColStr
	query = query.Suffix(qstr, args...)

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: UpdateRFEndpointsTx(): Query: %s - With args: %v", qStr, qArgs)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: UpdateRFEndpointsTx(): QueryContext: %s", err)
		return newEPs, err
	}
	defer rows.Close()
	for rows.Next() {
		ep, err := t.hdb.scanRedfishEndpoint(rows)
		if err != nil {
			return newEPs, err
		}
		newEPs = append(newEPs, ep)
	}
	return newEPs, nil
}

// Update RedfishEndpoint already in DB, leaving DiscoveryInfo
// unmodifed.  Does not update any ComponentEndpoint children.
// (In transaction.)
func (t *hmsdbPgTx) UpdateRFEndpointNoDiscInfoTx(ep *sm.RedfishEndpoint) (bool, error) {
	if ep == nil {
		t.LogAlways("Error: UpdateRFEndpointNoDiscInfoTx(): EP was nil.")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("UpdateRFEndpointNoDiscInfoTx",
		updatePgRFEndpointNoDiscInfoQuery)
	if err != nil {
		return false, err
	}
	// Normalize key
	normID := xnametypes.NormalizeHMSCompID(ep.ID)

	// Perform update
	res, err := stmt.ExecContext(t.ctx,
		&ep.Type,
		&ep.Name,
		&ep.Hostname,
		&ep.Domain,
		&ep.FQDN,
		&ep.Enabled,
		&ep.UUID,
		&ep.User,
		&ep.Password,
		&ep.UseSSDP,
		&ep.MACRequired,
		&ep.MACAddr,
		&ep.IPAddr,
		&ep.RediscOnUpdate,
		&ep.TemplateID,
		&normID) // Key
	if err != nil {
		t.LogAlways("Error: UpdateRFEndpointNoDiscInfoTx(): stmt.Exec: %s", err)
		if IsPgDuplicateKeyErr(err) == true {
			// Key already exists, user error. Set false and clear error.
			return false, ErrHMSDSDuplicateKey
		}
		return false, err
	}
	t.Log(LOG_INFO, "Info: UpdateRFEndpointNoDiscInfoTx() - %s", res)
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete RedfishEndpoint with matching xname id from database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteRFEndpointByIDTx(id string) (bool, error) {
	if id == "" {
		t.LogAlways("Error: DeleteRFEndpointByIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteRFEndpointByIDTx",
		deleteRFEndpointByIDQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		t.LogAlways("Error: DeleteRFEndpointByIDTx(%s): stmt.Exec: %s", id, err)
		return false, err
	}
	t.Log(LOG_INFO, "Info: DeleteRFEndpointByIDTx(%s) - %v", id, res)

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all RedfishEndpoints from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteRFEndpointsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteRFEndpointsAllTx",
		deleteRFEndpointsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteRFEndpointsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: DeleteRFEndpointsAllTx() - %s", res)

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

// Given the id of a RedfishEndpoint, set the states of all children
// with State/Components entries to state and flag, returning a list of
// xname IDs were at least state or flag was updated.
//
// This code relies on the fact that if an child Inventory/ComponentEndpoint
// entry exists, there should always be exactly one corresponding entry
// in State/Components.  In the other direction, there should always be
// zero or one ComponentEndpoint per State/Component.  The xnames (i.e. ID
// fields should always match.
//
// CREATES A WRITE LOCK ON Redfish/ComponentEndpoints and Components tables
// (all three) until transaction is committed if wrLock == true
//
// Detaches FRUs from locations if detachFRUs == true
func (t *hmsdbPgTx) SetChildCompStatesRFEndpointsTx(
	ids []string,
	state, flag string,
	wrLock bool,
	detachFRUs bool,
) ([]string, error) {
	// Nothing to do.
	if len(ids) < 1 {
		return []string{}, nil
	}
	// Get ComponentEndpoints IDs for these parent RedfishEndpoints in 'ids',
	// locking both tables
	cids, err := t.GetCompEndpointIDsTx(CE_RfEPs(ids), CE_WRLock,
		CE_From("SetChildCompStatesRFEndpointsTx"))
	if err != nil || len(cids) == 0 {
		return []string{}, err
	}
	mIDs, err := t.GetComponentIDsTx(IDs(cids), NotStateOrFlag(state, flag),
		WRLock, From("SetChildCompStatesRFEndpointsTx"))
	if err != nil {
		return []string{}, err
	}
	if len(mIDs) > 0 {
		// Update the states of the locked IDs so that they are state and flag
		// Force and skip verifying that changes were made since we already
		// filtered and locked the rows.
		_, err := t.UpdateCompStatesTx(mIDs, state, flag, true, true, new(PartInfo))
		if err != nil {
			return []string{}, err
		}
		// When setting component states to empty, we may want to detach
		// the FRUs from their location because they are being removed
		if detachFRUs {
			// Get the updated components and their children
			hwlocs, err := t.GetHWInvByLocQueryFilterTx(HWInvLoc_IDs(mIDs), HWInvLoc_Child)
			if err != nil {
				return []string{}, err
			}
			for _, hwloc := range hwlocs {
				// Delete just the loc, this will detach the FRU
				didDelete, err := t.DeleteHWInvByLocIDTx(hwloc.ID)
				if err != nil {
					return []string{}, err
				}
				if !didDelete || hwloc.PopulatedFRU == nil {
					continue
				}
				// Generate a history event for removing the FRU from the loc
				hwHist := sm.HWInvHist{
					ID:        hwloc.ID,
					FruId:     hwloc.PopulatedFRU.FRUID,
					EventType: sm.HWInvHistEventTypeRemoved,
				}
				t.InsertHWInvHistTx(&hwHist)
				if err != nil {
					return []string{}, err
				}
			}
		}
	}
	return mIDs, nil
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - RedfishEndpoint/ByComponent queries - ComponentEndpoint
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more ComponentEndpoint rows in
// the result.
func (t *hmsdbPgTx) queryComponentEndpoint(qname, query string, args ...interface{}) ([]*sm.ComponentEndpoint, error) {
	rows, err := t.getRowsForQuery(qname, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ceps := make([]*sm.ComponentEndpoint, 0, 1)
	i := 0
	for rows.Next() {
		cep, err := t.hdb.scanComponentEndpoint(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return ceps, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, cep)
		ceps = append(ceps, cep)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d EPs.", qname, args, len(ceps))
	return ceps, err
}

// Build filter query for ComponentEndpoints using filter functions and
// then return the set of matching components as an array.
//
// NOTE: Most args allow negated arguments, i.e. "!x0c0s0b0", so be careful about
// passing in user data if the query should only return a single result. CE_ID
// does not.
func (t *hmsdbPgTx) GetCompEndpointsTx(f_opts ...CompEPFiltFunc) ([]*sm.ComponentEndpoint, error) {
	f := new(CompEPFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	return t.GetCompEndpointsFilterTx(f)
}

// Get ComponentEndpoint by ID (xname), i.e. a single entry.
func (t *hmsdbPgTx) GetCompEndpointByIDTx(id string) (*sm.ComponentEndpoint, error) {
	label := "GetCompEndpointByIDTx"
	if id == "" {
		t.LogAlways("Error: GetCompEndpointByIDTx(): xname was empty")
		return nil, ErrHMSDSArgNil
	}
	// Perform corresponding query on DB
	ceps, err := t.GetCompEndpointsTx(CE_ID(id), CE_From(label))
	if err != nil {
		return nil, err
	}
	// Query succeeded.
	if len(ceps) == 0 {
		t.Log(LOG_INFO, "Info: %s(%s) matched no entry.", label, id)
		return nil, nil
	} else if len(ceps) > 1 {
		t.LogAlways("Warning: %s(%s): multiple entries!.", label, id)
	}
	return ceps[0], err
}

// Get all ComponentEndpoints in system (in transaction).
func (t *hmsdbPgTx) GetCompEndpointsAllTx() ([]*sm.ComponentEndpoint, error) {
	return t.GetCompEndpointsTx(CE_From("GetCompEndpointsAllTx"))
}

// Get some or all ComponentEndpoints in system (in transaction), with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (t *hmsdbPgTx) GetCompEndpointsFilterTx(f *CompEPFilter) ([]*sm.ComponentEndpoint, error) {
	var filterQuery string
	var ceps []*sm.ComponentEndpoint
	var err error
	label := "GetCompEndpointsFilterTx"

	// Use filter provided to to create a custom WHERE... string if non-nil.
	// Default uninitialized f or nil f will both select everything
	if f == nil {
		filterQuery = getCompEndpointsAllQuery
		// Perform corresponding query on DB
		ceps, err = t.queryComponentEndpoint(label, filterQuery)
		if err != nil {
			return nil, err
		}
	} else {
		if f.label != "" {
			label = f.label
		}
		filterQuery, args, err := buildCompEPQuery(getCompEndpointPrefix, f)
		if err != nil {
			return nil, err
		}
		// Perform corresponding query on DB
		ceps, err = t.queryComponentEndpoint(label, filterQuery, args...)
		if err != nil {
			return nil, err
		}
	}
	// Query succeeded.
	if len(ceps) == 0 {
		t.Log(LOG_INFO, "Info: %s(): no matches: %s", label, filterQuery)
		return ceps, nil
	}
	return ceps, nil
}

// Insert ComponentEndpoint into database, updating it if it exists
// (in transaction)
func (t *hmsdbPgTx) UpsertCompEndpointTx(cep *sm.ComponentEndpoint) error {
	if cep == nil {
		t.LogAlways("Error: UpsertCompEndpointTx(): Component was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("UpsertCompEndpointTx",
		upsertPgCompEndpointQuery)
	if err != nil {
		return err
	}
	compInfoJSON, err := cep.EncodeComponentInfo()
	if err != nil {
		// This should never fail
		t.LogAlways("UpsertCompEndpointTx: decode CompInfo: %s", err)
	}
	// Ensure endpoint name is normalized and valid
	normID := xnametypes.VerifyNormalizeCompID(cep.ID)
	if normID == "" {
		t.LogAlways("UpsertCompEndpointTx(%s): %s", normID, ErrHMSDSArgBadID)
		return ErrHMSDSArgBadID
	}
	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&normID,
		&cep.Type,
		&cep.Domain,
		&cep.RedfishType,
		&cep.RedfishSubtype,
		&cep.MACAddr,
		&cep.UUID,
		&cep.OdataID,
		&cep.RfEndpointID,
		&compInfoJSON)
	if err != nil {
		t.LogAlways("Error: UpsertCompEndpointTx(): stmt.Exec: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: UpsertCompEndpointTx() - %s", res)
	return nil
}

// Upsert ComponentEndpoints into database, updating them if they exist
// (in transaction)
func (t *hmsdbPgTx) UpsertCompEndpointsTx(ceps *sm.ComponentEndpointArray) error {
	if ceps == nil {
		t.LogAlways("Error: UpsertCompEndpointsTx(): Component array was nil.")
		return ErrHMSDSArgNil
	}
	if len(ceps.ComponentEndpoints) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(compEPsTable).
		Columns(compEPsAllCols...)

	for _, cep := range ceps.ComponentEndpoints {
		// Ensure endpoint name is normalized and valid
		normID := xnametypes.VerifyNormalizeCompID(cep.ID)
		if normID == "" {
			t.LogAlways("UpsertCompEndpointTx(%s): %s", normID, ErrHMSDSArgBadID)
			return ErrHMSDSArgBadID
		}

		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[normID]; ok {
			continue
		} else {
			valueMap[normID] = true
		}

		compInfoJSON, err := cep.EncodeComponentInfo()
		if err != nil {
			// This should never fail
			t.LogAlways("UpsertCompEndpointTx: decode CompInfo: %s", err)
		}

		// Set fields for the INSERT
		query = query.Values(
			normID,
			cep.Type,
			cep.Domain,
			cep.RedfishType,
			cep.RedfishSubtype,
			cep.RfEndpointID,
			cep.MACAddr,
			cep.UUID,
			cep.OdataID,
			compInfoJSON)
	}
	query = query.Suffix("ON CONFLICT(" + compEPsIdCol + ") DO UPDATE SET " +
		compEPsDomainCol + " = EXCLUDED." + compEPsDomainCol + ", " +
		compEPsRedfishTypeCol + " = EXCLUDED." + compEPsRedfishTypeCol + ", " +
		compEPsRedfishSubtypeCol + " = EXCLUDED." + compEPsRedfishSubtypeCol + ", " +
		compEPsRFEndpointIDCol + " = EXCLUDED." + compEPsRFEndpointIDCol + ", " +
		compEPsMACCol + " = EXCLUDED." + compEPsMACCol + ", " +
		compEPsUUIDCol + " = EXCLUDED." + compEPsUUIDCol + ", " +
		compEPsODataIDCol + " = EXCLUDED." + compEPsODataIDCol + ", " +
		compEPsComponentInfoCol + " = EXCLUDED." + compEPsComponentInfoCol)

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: UpsertCompEndpointsTx(): Query: %s - With args: %v", qStr, qArgs)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: UpsertCompEndpointsTx(): ExecContext: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: UpsertCompEndpointsTx() - %s", res)
	return nil
}

// Delete ComponentEndpoint with matching xname id from database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteCompEndpointByIDTx(id string) (bool, error) {
	if id == "" {
		t.LogAlways("Error: DeleteCompEndpointByIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteCompEndpointByIDTx",
		deleteCompEndpointByIDQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, xnametypes.NormalizeHMSCompID(id))
	if err != nil {
		t.LogAlways("Error: DeleteCompEndpointByIDTx(%s): stmt.Exec: %s",
			xnametypes.NormalizeHMSCompID(id), err)
		return false, err
	}
	t.Log(LOG_INFO, "Info: DeleteCompEndpointByIDTx(%s) - %s",
		xnametypes.NormalizeHMSCompID(id), res)

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all ComponentEndpoints from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteCompEndpointsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteCompEndpointsAllTx",
		deleteCompEndpointsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteCompEndpointsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: DeleteCompEndpointsAllTx() - %s", res)

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

// Given the id of a ComponentEndpoint, set the states of matching
// State/Components entries to state and flag, returning a list of
// xname IDs were at least state or flag was updated.
//
// This code relies on the fact that if an child Inventory/ComponentEndpoint
// entry exists, there should always be exactly one corresponding entry
// in State/Components.  In the other direction, there should always be
// zero or one ComponentEndpoint per State/Component.  The xnames (i.e. ID
// fields should always match.
//
// CREATES A WRITE LOCK ON Redfish/ComponentEndpoints and Components tables
// (all three) until transaction is committed if wrLock == true
func (t *hmsdbPgTx) SetChildCompStatesCompEndpointsTx(
	ids []string,
	state, flag string,
	wrLock bool,
) ([]string, error) {
	if len(ids) < 1 {
		return []string{}, nil
	}
	// Should return the same set of xname IDs as 'ids', assuming they are
	// actually in the database. Also locks both ComponentEP/RedfishEP tables
	// if wrLock is true, which is the main thing.
	cids, err := t.GetCompEndpointIDsTx(CE_IDs(ids), CE_WRLock)
	if err != nil || len(cids) == 0 {
		return []string{}, err
	}
	// Get xname IDs that are in the corresponding State/Components table
	// and do NOT have both state and flag already set - these are the ones
	// we can/should change.
	mIDs, err := t.GetComponentIDsTx(IDs(cids), NotStateOrFlag(state, flag), WRLock)
	if err != nil {
		return []string{}, err
	}
	// Update the states of the locked IDs so that they are state and flag
	// Force and skip verifying that changes were made since we already
	// filtered and locked the rows.
	if len(mIDs) > 0 {
		_, err := t.UpdateCompStatesTx(mIDs, state, flag, true, true, new(PartInfo))
		if err != nil {
			return []string{}, err
		}
	}
	return mIDs, nil
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - Service Endpoints
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more ServiceEndpoint rows in
// the result.
func (t *hmsdbPgTx) queryServiceEndpoint(qname, query string, args ...interface{}) ([]*sm.ServiceEndpoint, error) {
	rows, err := t.getRowsForQuery(qname, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seps := make([]*sm.ServiceEndpoint, 0, 1)
	i := 0
	for rows.Next() {
		sep, err := t.hdb.scanServiceEndpoint(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return seps, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, sep)
		seps = append(seps, sep)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d EPs.", qname, args, len(seps))
	return seps, err
}

// Build filter query for ServiceEndpoints using filter functions and
// then return the set of matching components as an array.
func (t *hmsdbPgTx) GetServiceEndpointsTx(f_opts ...ServiceEPFiltFunc) ([]*sm.ServiceEndpoint, error) {
	f := new(ServiceEPFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	return t.GetServiceEndpointsFilterTx(f)
}

// Get ServiceEndpoint by service type and ID (xname), i.e. a single entry.
func (t *hmsdbPgTx) GetServiceEndpointByIDTx(svc, id string) (*sm.ServiceEndpoint, error) {
	label := "GetServiceEndpointByIDTx"
	if svc == "" {
		t.LogAlways("Error: GetServiceEndpointByIDTx(): service type was empty")
		return nil, ErrHMSDSArgNil
	}
	if id == "" {
		t.LogAlways("Error: GetServiceEndpointByIDTx(): xname was empty")
		return nil, ErrHMSDSArgNil
	}
	// Perform corresponding query on DB
	seps, err := t.GetServiceEndpointsTx(SE_RfSvc(svc), SE_RfEP(id), SE_From(label))
	if err != nil {
		return nil, err
	}
	// Query succeeded.
	if len(seps) == 0 {
		t.Log(LOG_INFO, "Info: %s(%s,%s) matched no entry.", label, svc, id)
		return nil, nil
	} else if len(seps) > 1 {
		t.LogAlways("Warning: %s(%s,%s): multiple entries!.", label, svc, id)
	}
	return seps[0], err
}

// Get all ServiceEndpoints in system (in transaction).
func (t *hmsdbPgTx) GetServiceEndpointsAllTx() ([]*sm.ServiceEndpoint, error) {
	return t.GetServiceEndpointsTx(SE_From("GetServiceEndpointsAllTx"))
}

// Get some or all ServiceEndpoints in system (in transaction), with
// filtering options to possibly narrow the returned values.
// If no filter provided, just get everything.  Otherwise use it
// to create a custom WHERE... string that filters out entries that
// do not match ALL of the non-empty strings in the filter struct
func (t *hmsdbPgTx) GetServiceEndpointsFilterTx(f *ServiceEPFilter) ([]*sm.ServiceEndpoint, error) {
	var filterQuery string
	var seps []*sm.ServiceEndpoint
	var err error
	label := "GetServiceEndpointsFilterTx"

	// Use filter provided to to create a custom WHERE... string if non-nil.
	// Default uninitialized f or nil f will both select everything
	if f == nil {
		filterQuery = getServiceEndpointsAllQuery
		// Perform corresponding query on DB
		seps, err = t.queryServiceEndpoint(label, filterQuery)
		if err != nil {
			return nil, err
		}
	} else {
		if f.label != "" {
			label = f.label
		}
		filterQuery, args, err := buildServiceEPQuery(getServiceEndpointPrefix, f)
		if err != nil {
			return nil, err
		}
		// Perform corresponding query on DB
		seps, err = t.queryServiceEndpoint(label, filterQuery, args...)
		if err != nil {
			return nil, err
		}
	}
	// Query succeeded.
	if len(seps) == 0 {
		t.Log(LOG_INFO, "Info: %s(): no matches: %s", label, filterQuery)
		return seps, nil
	}
	return seps, nil
}

// Insert ServiceEndpoint into database, updating it if it exists
// (in transaction)
func (t *hmsdbPgTx) UpsertServiceEndpointTx(sep *sm.ServiceEndpoint) error {
	if sep == nil {
		t.LogAlways("Error: UpsertServiceEndpointTx(): Service was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("UpsertServiceEndpointTx",
		upsertPgServiceEndpointQuery)
	if err != nil {
		return err
	}
	// Normalize key
	normRFID := xnametypes.NormalizeHMSCompID(sep.RfEndpointID)

	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&normRFID,
		&sep.RedfishType,
		&sep.RedfishSubtype,
		&sep.UUID,
		&sep.OdataID,
		&sep.ServiceInfo)
	if err != nil {
		t.LogAlways("Error: UpsertServiceEndpointTx(): stmt.Exec: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: UpsertServiceEndpointTx() - %s", res)
	return nil
}

// Insert ServiceEndpoints into database, updating them if they exist
// (in transaction)
func (t *hmsdbPgTx) UpsertServiceEndpointsTx(seps *sm.ServiceEndpointArray) error {
	if seps == nil {
		t.LogAlways("Error: UpsertServiceEndpointsTx(): Service array was nil.")
		return ErrHMSDSArgNil
	}
	if len(seps.ServiceEndpoints) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(serviceEPsTable).
		Columns(serviceEPsCols...)

	for _, sep := range seps.ServiceEndpoints {
		if sep == nil {
			t.LogAlways("Error: UpsertServiceEndpointsTx(): Service Endpoint was nil.")
			return ErrHMSDSArgNil
		}
		// Normalize key
		normRFID := xnametypes.NormalizeHMSCompID(sep.RfEndpointID)
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		key := normRFID + sep.RedfishType
		if _, ok := valueMap[key]; ok {
			continue
		} else {
			valueMap[key] = true
		}
		// Set fields for the INSERT
		query = query.Values(
			normRFID,
			sep.RedfishType,
			sep.RedfishSubtype,
			sep.UUID,
			sep.OdataID,
			sep.ServiceInfo)
	}
	query = query.Suffix("ON CONFLICT(" + serviceEPsRFEndpointIDCol + ", " + serviceEPsRedfishTypeCol + ") DO UPDATE SET " +
		serviceEPsRedfishSubtypeCol + " = EXCLUDED." + serviceEPsRedfishSubtypeCol + ", " +
		serviceEPsUUIDCol + " = EXCLUDED." + serviceEPsUUIDCol + ", " +
		serviceEPsODataIDCol + " = EXCLUDED." + serviceEPsODataIDCol + ", " +
		serviceEPsServiceInfoCol + " = EXCLUDED." + serviceEPsServiceInfoCol)

	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: UpsertServiceEndpointsTx(): Query: %s - With args: %v", qStr, qArgs)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: UpsertServiceEndpointsTx(): ExecContext: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: UpsertServiceEndpointsTx() - %s", res)
	return nil
}

// Delete ServiceEndpoint with matching service type and xname id from
// database, if it exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteServiceEndpointByIDTx(svc, id string) (bool, error) {
	if svc == "" {
		t.LogAlways("Error: DeleteServiceEndpointByIDTx(): service type was empty")
		return false, ErrHMSDSArgNil
	}
	if id == "" {
		t.LogAlways("Error: DeleteServiceEndpointByIDTx(): xname was empty")
		return false, ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteServiceEndpointByIDTx",
		deleteServiceByRfEPandRfTypeQuery)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, xnametypes.NormalizeHMSCompID(id), svc)
	if err != nil {
		t.LogAlways("Error: DeleteServiceEndpointByIDTx(%s,%s): stmt.Exec: %s",
			svc, id, err)
		return false, err
	}
	t.Log(LOG_INFO, "Info: DeleteServiceEndpointByIDTx(%s,%s) - %s", svc, id, res)

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all ServiceEndpoints from database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteServiceEndpointsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteServiceEndpointsAllTx",
		deleteServiceEndpointsAllQuery)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteServiceEndpointsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: DeleteServiceEndpointsAllTx() - %s", res)

	// Return rows affected (if no error) and nil error, or else
	// undefined number + error from RowsAffected.
	return res.RowsAffected()
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - CompEthInterface queries
//
/////////////////////////////////////////////////////////////////////////////

// Get CompEthInterface by ID, i.e. a single entry for UPDATE (in transaction).
func (t *hmsdbPgTx) GetCompEthInterfaceByIDTx(id string) (*sm.CompEthInterfaceV2, error) {
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	// Generate query
	query := sq.Select(compEthCols...).
		From(compEthTable).
		Where(sq.Eq{compEthIdCol: id})

	// Query with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetCompEthInterfaceByIDTx(%s): query failed: %s", id, err)
		return nil, err
	}
	defer rows.Close()

	cei := new(sm.CompEthInterfaceV2)
	if rows.Next() {
		var ipAddresses []byte

		err := rows.Scan(&cei.ID, &cei.Desc, &cei.MACAddr, &cei.LastUpdate, &cei.CompID, &cei.Type, &ipAddresses)
		if err != nil {
			t.LogAlways("Error: GetCompEthInterfaceByIDTx(%s): scan failed: %s", id, err)
			return nil, err
		}

		err = json.Unmarshal(ipAddresses, &cei.IPAddrs)
		if err != nil {
			t.LogAlways("Warning: GetCompEthInterfaceByIDTx(): Decode IPAddresses: %s", err)
			return nil, err

		}
	}
	return cei, err
}

// Insert a new CompEthInterface into database (in transaction)
// If ID or MAC already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (t *hmsdbPgTx) InsertCompEthInterfaceTx(cei *sm.CompEthInterfaceV2) error {
	var err error
	if cei == nil {
		t.LogAlways("Error: InsertCompEthInterfaceTx(): Struct was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	cei.MACAddr = strings.ToLower(cei.MACAddr)
	cei.ID = strings.ReplaceAll(cei.MACAddr, ":", "")
	if cei.ID == "" {
		return ErrHMSDSArgBadArg
	}
	if cei.CompID != "" {
		cei.CompID = xnametypes.VerifyNormalizeCompID(cei.CompID)
		if cei.CompID == "" {
			return ErrHMSDSArgBadID
		}
	}
	if cei.Type != "" {
		cei.Type = xnametypes.VerifyNormalizeType(cei.Type)
		if cei.Type == "" {
			return ErrHMSDSArgBadType
		}
	}

	ipAddrs, err := json.Marshal(cei.IPAddrs)
	if err != nil {
		// This should never fail
		t.LogAlways("InsertCompEthInterfaceTx: decode Details: %s", err)
		return err
	}

	// Generate query
	query := sq.Insert(compEthTable).
		Columns(compEthCols...).
		Values(cei.ID, cei.Desc, cei.MACAddr, "NOW()", cei.CompID, cei.Type, ipAddrs)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

// Insert new CompEthInterfaces into database (in transaction)
// If ID or MAC already exists, return ErrHMSDSDuplicateKey
// No insertion done on err != nil
func (t *hmsdbPgTx) InsertCompEthInterfacesTx(ceis []*sm.CompEthInterfaceV2) error {
	var err error
	if len(ceis) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(compEthTable).
		Columns(compEthCols...)

	for _, cei := range ceis {
		cei.MACAddr = strings.ToLower(cei.MACAddr)
		cei.ID = strings.ReplaceAll(cei.MACAddr, ":", "")
		if cei.ID == "" {
			return ErrHMSDSArgBadArg
		}
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[cei.ID]; ok {
			continue
		} else {
			valueMap[cei.ID] = true
		}
		if cei.CompID != "" {
			cei.CompID = xnametypes.VerifyNormalizeCompID(cei.CompID)
			if cei.CompID == "" {
				return ErrHMSDSArgBadID
			}
		}
		if cei.Type != "" {
			cei.Type = xnametypes.VerifyNormalizeType(cei.Type)
			if cei.Type == "" {
				return ErrHMSDSArgBadType
			}
		}

		ipAddrs, err := json.Marshal(cei.IPAddrs)
		if err != nil {
			// This should never fail
			t.LogAlways("InsertCompEthInterfacesTx: decode Details: %s", err)
			return err
		}
		query = query.Values(
			cei.ID,
			cei.Desc,
			cei.MACAddr,
			"NOW()",
			cei.CompID,
			cei.Type,
			ipAddrs)
	}

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: InsertCompEthInterfacesTx(): Query: %s - With args: %v", qStr, qArgs)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

// Insert/update a new CompEthInterface into the database (in transaction)
// If ID or FQDN already exists, only overwrite ComponentID
// and Type fields.
// No insertion done on err != nil
func (t *hmsdbPgTx) InsertCompEthInterfaceCompInfoTx(cei *sm.CompEthInterfaceV2) error {
	var err error
	if cei == nil {
		t.LogAlways("Error: InsertCompEthInterfaceCompInfoTx(): Struct was nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	cei.MACAddr = strings.ToLower(cei.MACAddr)
	cei.ID = strings.ReplaceAll(cei.MACAddr, ":", "")
	if cei.ID == "" {
		return ErrHMSDSArgBadArg
	}
	if cei.CompID != "" {
		cei.CompID = xnametypes.VerifyNormalizeCompID(cei.CompID)
		if cei.CompID == "" {
			return ErrHMSDSArgBadID
		}
	}
	if cei.Type != "" {
		cei.Type = xnametypes.VerifyNormalizeType(cei.Type)
		if cei.Type == "" {
			return ErrHMSDSArgBadType
		}
	}

	ipAddrs, err := json.Marshal(cei.IPAddrs)
	if err != nil {
		// This should never fail
		t.LogAlways("InsertCompEthInterfaceCompInfoTx: decode Details: %s", err)
	}

	// Generate query
	query := sq.Insert(compEthTable).
		Columns(compEthCols...).
		Values(cei.ID, cei.Desc, cei.MACAddr, "NOW()", cei.CompID, cei.Type, ipAddrs).
		Suffix("ON CONFLICT(id) DO UPDATE SET " +
			compEthCompIDCol + " = EXCLUDED." + compEthCompIDCol + ", " +
			compEthTypeCol + " = EXCLUDED." + compEthTypeCol)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

// Insert/update new CompEthInterfaces into the database (in transaction)
// If ID or FQDN already exists, only overwrite ComponentID
// and Type fields.
// No insertion done on err != nil
func (t *hmsdbPgTx) InsertCompEthInterfacesCompInfoTx(ceis []*sm.CompEthInterfaceV2) error {
	var err error
	if len(ceis) == 0 {
		return nil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	valueMap := make(map[string]bool)

	// Generate query
	query := sq.Insert(compEthTable).
		Columns(compEthCols...)

	for _, cei := range ceis {
		cei.MACAddr = strings.ToLower(cei.MACAddr)
		cei.ID = strings.ReplaceAll(cei.MACAddr, ":", "")
		if cei.ID == "" {
			return ErrHMSDSArgBadArg
		}
		// Take out duplicates so that we don't get errors for modifying a row multiple times.
		if _, ok := valueMap[cei.ID]; ok {
			continue
		} else {
			valueMap[cei.ID] = true
		}
		if cei.CompID != "" {
			cei.CompID = xnametypes.VerifyNormalizeCompID(cei.CompID)
			if cei.CompID == "" {
				return ErrHMSDSArgBadID
			}
		}
		if cei.Type != "" {
			cei.Type = xnametypes.VerifyNormalizeType(cei.Type)
			if cei.Type == "" {
				return ErrHMSDSArgBadType
			}
		}

		ipAddrs, err := json.Marshal(cei.IPAddrs)
		if err != nil {
			// This should never fail
			t.LogAlways("InsertCompEthInterfacesCompInfoTx: decode Details: %s", err)
		}
		query = query.Values(
			cei.ID,
			cei.Desc,
			cei.MACAddr,
			"NOW()",
			cei.CompID,
			cei.Type,
			ipAddrs)
	}
	query = query.Suffix("ON CONFLICT(" + compEthIdCol + ") DO UPDATE SET " +
		compEthCompIDCol + " = EXCLUDED." + compEthCompIDCol + ", " +
		compEthTypeCol + " = EXCLUDED." + compEthTypeCol)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	qStr, qArgs, _ := query.ToSql()
	t.Log(LOG_DEBUG, "Debug: InsertCompEthInterfacesCompInfoTx(): Query: %s - With args: %v", qStr, qArgs)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

// Update CompEthInterface already in the DB. (In transaction.)
// If err == nil, but FALSE is returned, then no changes were made.
func (t *hmsdbPgTx) UpdateCompEthInterfaceTx(cei *sm.CompEthInterfaceV2, ceip *sm.CompEthInterfaceV2Patch) (bool, error) {
	var doUpdate bool

	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}

	if ceip == nil || cei == nil {
		return false, nil
	}

	// Start update query string
	update := sq.Update(compEthTable).
		Where(sq.Eq{compEthIdCol: cei.ID})

	// Check to see if there are any fields set in the update and then
	// see if they need to be updated.
	if ceip.Desc != nil && cei.Desc != *ceip.Desc {
		update = update.Set(compEthDescCol, *ceip.Desc)
		doUpdate = true
	}
	// We might want to update even if the IPAddress
	// did not change just to update the timestamp.
	if ceip.IPAddrs != nil {
		ipAddrs, err := json.Marshal(*ceip.IPAddrs)
		if err != nil {
			// This should never fail
			t.LogAlways("UpdateCompEthInterfaceTx: decode Details: %s", err)
			return false, err
		}

		update = update.Set(compEthIPAddressesCol, ipAddrs)
		update = update.Set(compEthLastUpdateCol, "NOW()")
		doUpdate = true
	}

	if ceip.CompID != nil {
		compNorm := xnametypes.VerifyNormalizeCompID(*ceip.CompID)
		if compNorm == "" {
			return false, ErrHMSDSArgBadID
		}
		update = update.Set(compEthCompIDCol, compNorm)
		ctype := xnametypes.GetHMSTypeString(compNorm)
		update = update.Set(compEthTypeCol, ctype)
		doUpdate = true
	}

	// Have a change to make...
	if doUpdate == true {
		// Exec with statement cache for caching prepared statements
		update = update.PlaceholderFormat(sq.Dollar)
		res, err := update.RunWith(t.sc).ExecContext(t.ctx)
		if err != nil {
			return false, err
		}
		num, err := res.RowsAffected()
		if err != nil {
			return false, err
		} else if num > 0 {
			return true, nil
		}
	}
	return false, nil
}

// Delete a CompEthInterface with matching id from the database, if it
// exists (in transaction)
// Return true if there was a row affected, false if there were zero.
func (t *hmsdbPgTx) DeleteCompEthInterfaceByIDTx(id string) (bool, error) {
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}

	// Build query - works like AND
	query := sq.Delete(compEthTable).
		Where(sq.Eq{compEthIdCol: id})

	// Execute - Should delete one row.
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
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
				t.LogAlways("Error: DeleteCompEthInterfaceByIDTx(): multiple deletions!")
			}
			return true, nil
		}
	}
	return false, nil
}

// Delete all CompEthInterfaces from the database (in transaction).
// Also returns number of deleted rows, if error is nil.
func (t *hmsdbPgTx) DeleteCompEthInterfacesAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}

	// Build query
	query := sq.Delete(compEthTable)

	// Execute.
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		return 0, err
	}
	// See if any rows were affected
	return res.RowsAffected()
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - Discovery status
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more DiscoveryStatus rows in
// the result.
func (t *hmsdbPgTx) queryDiscoveryStatus(qname, query string, args ...interface{}) ([]*sm.DiscoveryStatus, error) {
	rows, err := t.getRowsForQuery(qname, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ceps := make([]*sm.DiscoveryStatus, 0, 1)
	i := 0
	for rows.Next() {
		cep, err := t.hdb.scanDiscoveryStatus(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return ceps, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, cep)
		ceps = append(ceps, cep)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_INFO, "Info: %s(%v) returned %d EPs.", qname, args, len(ceps))
	return ceps, err
}

// Get DiscoveryStatus with the given numerical ID (in transaction).
func (t *hmsdbPgTx) GetDiscoveryStatusByIDTx(id uint) (*sm.DiscoveryStatus, error) {
	if id < 0 {
		t.LogAlways("Error: GetDiscoveryStatusByIDTx(): id is invalid")
		return nil, ErrHMSDSArgBadRange
	}
	// Perform corresponding query on DB
	stats, err := t.queryDiscoveryStatus("GetDiscoveryStatusByIDTx",
		getDiscoveryStatusByIDQuery, id)
	if err != nil {
		return nil, err
	}
	// Query succeeded.
	if len(stats) == 0 {
		t.Log(LOG_INFO, "Info: GetDiscoveryStatusByIDTx(%d) matched no entry.", id)
		return nil, nil
	} else if len(stats) > 1 {
		t.LogAlways("Warning: GetDiscoveryStatusByIDTx(%d): multiple entries!.", id)
	}
	return stats[0], err
}

// Get all DiscoveryStatus entries (in transaction).
func (t *hmsdbPgTx) GetDiscoveryStatusAllTx() ([]*sm.DiscoveryStatus, error) {
	// Perform corresponding query on DB
	stats, err := t.queryDiscoveryStatus("GetDiscoveryStatusAllTx",
		getDiscoveryStatusesAllQuery)
	if err != nil {
		return stats, err
	}
	// Query succeeded.
	if len(stats) == 0 {
		t.Log(LOG_INFO, "Info: GetDiscoveryStatusAllTx() matched no entry.")
	}
	return stats, nil
}

// Update discovery status in DB (in transaction)
func (t *hmsdbPgTx) UpsertDiscoveryStatusTx(stat *sm.DiscoveryStatus) error {
	if stat == nil {
		t.LogAlways("Error: UpsertDiscoveryStatusTx(): DiscoveryStatus = nil.")
		return ErrHMSDSArgNil
	}
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("UpsertDiscoveryStatusTx",
		upsertPgDiscoveryStatusQuery)
	if err != nil {
		return err
	}
	detailsJSON, err := json.Marshal(stat.Details)
	if err != nil {
		// This should never fail
		t.LogAlways("UpsertDiscoveryStatusTx: decode Details: %s", err)
	}
	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&stat.ID,
		&stat.Status,
		&detailsJSON)
	if err != nil {
		t.LogAlways("Error: UpsertDiscoveryStatusTx(): stmt.Exec: %s", err)
		return err
	}
	t.Log(LOG_INFO, "Info: UpsertDiscoveryStatusTx() - %+v", res)
	return nil
}

/////////////////////////////////////////////////////////////////////////////
//
// HMSDBTx Interface - SCN subscription operations
//
/////////////////////////////////////////////////////////////////////////////

// Back end for all queries that produce one or more SCN subscription rows in
// the result.
func (t *hmsdbPgTx) querySCNSubscription(qname, query string, args ...interface{}) (*sm.SCNSubscriptionArray, error) {
	var subs sm.SCNSubscriptionArray

	stmt, err := t.conditionalPrepare(qname, query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.QueryContext(t.ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		sub, err := t.hdb.scanSCNSubscription(rows)
		if err != nil {
			t.LogAlways("Error: %s(%v): Scan failed: %s", qname, args, err)
			return &subs, err
		}
		t.Log(LOG_DEBUG, "Debug: %s() scanned[%d]: %v", qname, i, *sub)
		subs.SubscriptionList = append(subs.SubscriptionList, *sub)
		i += 1
	}
	err = rows.Err()
	t.Log(LOG_DEBUG, "Debug: %s(%v) returned %d entries.",
		qname, args, len(subs.SubscriptionList))
	return &subs, err
}

// Get all SCN subscriptions
func (t *hmsdbPgTx) GetSCNSubscriptionsAllTx() (*sm.SCNSubscriptionArray, error) {
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	// Perform corresponding query on DB
	subs, err := t.querySCNSubscription("GetSCNSubscriptionsAllTx", getSCNSubsAll)
	if err != nil {
		return nil, err
	}

	// Query succeeded.
	// Note: no reason to log no subscriptions - redundant.
	return subs, nil
}

// Get a SCN subscription
func (t *hmsdbPgTx) GetSCNSubscriptionTx(id int64) (*sm.SCNSubscription, error) {
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	// Perform corresponding query on DB
	subs, err := t.querySCNSubscription("GetSCNSubscriptionTx", getSCNSub, id)
	if err != nil {
		return nil, err
	}

	if subs.SubscriptionList == nil || len(subs.SubscriptionList) == 0 {
		// Not Found
		return nil, nil
	}
	// Query succeeded.
	// Note: no reason to log no subscriptions - redundant.
	return &subs.SubscriptionList[0], nil
}

// Insert a new SCN subscription. Existing subscriptions are unaffected
func (t *hmsdbPgTx) InsertSCNSubscriptionTx(sub sm.SCNPostSubscription) (int64, error) {
	var id int64
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("InsertSCNSubscriptionTx", insertSCNSub)
	if err != nil {
		return 0, err
	}
	jsonSub, err := json.Marshal(sub)
	if err != nil {
		t.LogAlways("InsertSCNSubscriptionTx: encode SCNPostSubscription: %s", err)
		return 0, err
	}
	key := sub.Subscriber + sub.Url
	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&key,
		&jsonSub)
	if err != nil {
		t.LogAlways("Error: InsertSCNSubscriptionTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: InsertSCNSubscriptionTx() - %v", res)

	//Get the generated ID
	stmt, err = t.conditionalPrepare("GetSCNSubIDTx", getPgSCNSubID)
	if err != nil {
		return 0, err
	}
	rows, err := stmt.QueryContext(t.ctx)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	rows.Next()
	err = rows.Scan(&id)
	if err != nil {
		t.LogAlways("Error: GetSCNSubIDTx(): Scan failed: %s", err)
		return 0, err
	}
	t.Log(LOG_DEBUG, "Debug: GetSCNSubIDTx() scanned: %d", id)
	err = rows.Err()
	t.Log(LOG_DEBUG, "Debug: GetSCNSubIDTx() returned id=%d entries.", id)
	return id, nil
}

// Update an existing SCN subscription.
func (t *hmsdbPgTx) UpdateSCNSubscriptionTx(id int64, sub sm.SCNPostSubscription) (bool, error) {
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("UpdateSCNSubscriptionTx", updateSCNSub)
	if err != nil {
		return false, err
	}
	jsonSub, err := json.Marshal(sub)
	if err != nil {
		t.LogAlways("UpdateSCNSubscriptionTx: encode SCNPostSubscription: %s", err)
		return false, err
	}
	key := sub.Subscriber + sub.Url
	// Perform insert
	res, err := stmt.ExecContext(t.ctx,
		&key,
		&jsonSub,
		&id)
	if err != nil {
		t.LogAlways("Error: UpdateSCNSubscriptionTx(): stmt.Exec: %s", err)
		return false, err
	}
	t.Log(LOG_INFO, "Info: UpdateSCNSubscriptionTx() - %v", res)

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Patch an existing SCN subscription.
func (t *hmsdbPgTx) PatchSCNSubscriptionTx(id int64, op string, patch sm.SCNPatchSubscription) (bool, error) {
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	if len(op) == 0 {
		t.LogAlways("Error: PatchSCNSubscriptionTx(): Missing Patch Op")
		return false, ErrHMSDSArgBadArg
	}
	opInt, ok := hmsdsPatchOpMap[strings.ToLower(op)]
	if !ok {
		t.LogAlways("Error: PatchSCNSubscriptionTx(): Invalid Patch Op - %s", op)
		return false, ErrHMSDSArgBadArg
	}
	// Perform corresponding query on DB
	subs, err := t.querySCNSubscription("PatchSCNSubscriptionTx", getSCNSubUpdate, id)
	if err != nil {
		return false, err
	}
	if subs.SubscriptionList == nil || len(subs.SubscriptionList) == 0 {
		// Not Found
		return false, nil
	}
	sub := subs.SubscriptionList[0]

	switch opInt {
	case PatchOpAdd:
		// Find out which values in the request are not already in our
		// current subscription and add them.
		for _, newState := range patch.States {
			match := false
			for _, state := range sub.States {
				if state == newState {
					match = true
					break
				}
			}
			if !match {
				sub.States = append(sub.States, newState)
			}
		}
		for _, newRole := range patch.Roles {
			match := false
			for _, role := range sub.Roles {
				if role == newRole {
					match = true
					break
				}
			}
			if !match {
				sub.Roles = append(sub.Roles, newRole)
			}
		}
		for _, newSubRole := range patch.SubRoles {
			match := false
			for _, subRole := range sub.SubRoles {
				if subRole == newSubRole {
					match = true
					break
				}
			}
			if !match {
				sub.SubRoles = append(sub.SubRoles, newSubRole)
			}
		}
		for _, newSoftwareStatus := range patch.SoftwareStatus {
			match := false
			for _, SoftwareStatus := range sub.SoftwareStatus {
				if SoftwareStatus == newSoftwareStatus {
					match = true
					break
				}
			}
			if !match {
				sub.SoftwareStatus = append(sub.SoftwareStatus, newSoftwareStatus)
			}
		}
		// The add patch op will only ever change the enabled field from false to true.
		// Only show a change if our request has Enabled=true and our current subscription is enabled=false
		if patch.Enabled != nil && *patch.Enabled &&
			sub.Enabled != nil && !*sub.Enabled {
			sub.Enabled = patch.Enabled
		}
	case PatchOpRemove:
		// Find out which values in the request are in our
		// current subscription and remove them.
		for _, newState := range patch.States {
			for j, state := range sub.States {
				if state == newState {
					sub.States = append(sub.States[:j], sub.States[j+1:]...)
					break
				}
			}
		}
		for _, newRole := range patch.Roles {
			for j, role := range sub.Roles {
				if role == newRole {
					sub.Roles = append(sub.Roles[:j], sub.Roles[j+1:]...)
					break
				}
			}
		}
		for _, newSubRole := range patch.SubRoles {
			for j, subRole := range sub.SubRoles {
				if subRole == newSubRole {
					sub.SubRoles = append(sub.SubRoles[:j], sub.SubRoles[j+1:]...)
					break
				}
			}
		}
		for _, newSoftwareStatus := range patch.SoftwareStatus {
			for j, SoftwareStatus := range sub.SoftwareStatus {
				if SoftwareStatus == newSoftwareStatus {
					sub.SoftwareStatus = append(sub.SoftwareStatus[:j], sub.SoftwareStatus[j+1:]...)
					break
				}
			}
		}
		// The remove patch op will only ever change the enabled field from true to false.
		// Only show a change if our request has Enabled=true and our current subscription is Enabled=true
		if patch.Enabled != nil && *patch.Enabled &&
			sub.Enabled != nil && *sub.Enabled {
			*sub.Enabled = false
		}
	case PatchOpReplace:
		if len(patch.States) > 0 {
			sub.States = patch.States
		}
		if len(patch.Roles) > 0 {
			sub.Roles = patch.Roles
		}
		if len(patch.SubRoles) > 0 {
			sub.SubRoles = patch.SubRoles
		}
		if len(patch.SoftwareStatus) > 0 {
			sub.SoftwareStatus = patch.SoftwareStatus
		}
		if patch.Enabled != nil {
			sub.Enabled = patch.Enabled
		}
	default:
		// Shouldn't happen
		t.LogAlways("Error: PatchSCNSubscriptionTx(): Invalid Patch Op - %s", op)
		return false, ErrHMSDSArgBadArg
	}
	newSub := sm.SCNPostSubscription{
		Subscriber:     sub.Subscriber,
		Enabled:        sub.Enabled,
		Roles:          sub.Roles,
		SubRoles:       sub.SubRoles,
		SoftwareStatus: sub.SoftwareStatus,
		States:         sub.States,
		Url:            sub.Url,
	}

	didUpdate, err := t.UpdateSCNSubscriptionTx(id, newSub)
	return didUpdate, err
}

// Delete a SCN subscription
func (t *hmsdbPgTx) DeleteSCNSubscriptionTx(id int64) (bool, error) {
	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteSCNSubscriptionTx", deleteSCNSubscription)
	if err != nil {
		return false, err
	}
	res, err := stmt.ExecContext(t.ctx, id)
	if err != nil {
		t.LogAlways("Error: DeleteSCNSubscriptionTx(%d): stmt.Exec: %s", id, err)
		return false, err
	}
	t.Log(LOG_INFO, "Info: DeleteSCNSubscriptionTx(%d) - %s", id, res)

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Delete all SCN subscriptions
func (t *hmsdbPgTx) DeleteSCNSubscriptionsAllTx() (int64, error) {
	if !t.IsConnected() {
		return 0, ErrHMSDSPtrClosed
	}
	// Prepare query
	stmt, err := t.conditionalPrepare("DeleteSCNSubscriptionsAllTx", deleteSCNSubscriptionsAll)
	if err != nil {
		return 0, err
	}
	res, err := stmt.ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: DeleteSCNSubscriptionsAllTx(): stmt.Exec: %s", err)
		return 0, err
	}
	t.Log(LOG_INFO, "Info: DeleteSCNSubscriptionsAllTx() - %s", res)

	num, err := res.RowsAffected()

	return num, err
}

////////////////////////////////////////////////////////////////////////////
//
// Group and Partition  Management
//
////////////////////////////////////////////////////////////////////////////

//
// Groups
//

// Creates new group in component groups, but adds nothing to the members
// table (in tx, so this can be done in separate query)
//
// Returns: (new UUID string, new group's label, excl group name, error)
func (t *hmsdbPgTx) InsertEmptyGroupTx(g *sm.Group) (
	string, string, string, error,
) {
	var err error
	gi := new(compGroupsInsert)

	if !t.IsConnected() {
		return "", "", "", ErrHMSDSPtrClosed
	}
	// Normalize and verify fields (note these functions track if this
	// has been done and only does each once.)
	g.Normalize()
	if err = g.Verify(); err != nil {
		return "", "", "", err
	}
	// Set fields for update
	gi.id = uuid.New().String()    // Used only internally for joins
	gi.name = g.Label              // Must be unique accross all groups
	gi.description = g.Description // Free-form shortish string

	gi.tags = g.Tags // For MariaDB tags are a json array.

	if g.ExclusiveGroup == "" { // Should be valid chars or empty.
		gi.gtype = groupType
	} else {
		gi.gtype = exclGroupType
	}
	gi.namespace = groupNamespace          // 'Group' enum val - vs. Partition
	gi.exclusiveGroupId = g.ExclusiveGroup // empty string == no exclusive group

	// Generate query
	query := sq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(gi.id, gi.name, gi.description,
			pq.Array(gi.tags), gi.gtype, gi.namespace, gi.exclusiveGroupId)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return gi.id, gi.name, gi.exclusiveGroupId, ParsePgDBError(err)
}

// Update fields in GroupPatch on the returned Group object provided
// (in transaction).
func (t *hmsdbPgTx) UpdateEmptyGroupTx(
	uuid string,
	g *sm.Group,
	gp *sm.GroupPatch,
) error {
	var err error
	var doUpdate bool

	if g == nil || gp == nil {
		return nil
	}
	// Start update query string
	update := sq.Update("").
		Table(compGroupsTable).
		Where(sq.Eq{compGroupIdCol: uuid})

	// Check to see if there are any fields set in the update and then
	// see if they need to be updated.
	if gp.Description != nil && g.Description != *gp.Description {
		update = update.Set(compGroupDescCol, *gp.Description)
		doUpdate = true
	}
	if gp.Tags != nil {
		inTagLen := len(*gp.Tags)
		if inTagLen != len(g.Tags) {
			// Different array lengths - don't need to check contents, update.
			doUpdate = true
			update = update.Set(compGroupTagsCol, pq.Array(gp.Tags))
		} else {
			// Same array length - check individual entries and update if
			// they don't match.
			gotMismatch := false
			for i, str := range *gp.Tags {
				if g.Tags[i] != str {
					gotMismatch = true
				}
			}
			// One or more tags did not match -  note we will update if
			// the order changes, but the ordering should always get returned
			// the same way, so the client should have to go out of it's way
			// to change it.
			if gotMismatch == true {
				update = update.Set(compGroupTagsCol, pq.Array(gp.Tags))
				doUpdate = true
			}
		}
	}
	// Have a change to make...
	if doUpdate == true {
		// Exec with statement cache for caching prepared statements
		update = update.PlaceholderFormat(sq.Dollar)
		_, err = update.RunWith(t.sc).ExecContext(t.ctx)
	}
	return err
}

// Get the user-readable fields in a group entry and it's internal uuid but
// don't fetch its members (done in transaction, so we can fetch them as part
// of the same one).
func (t *hmsdbPgTx) GetEmptyGroupTx(label string) (
	uuid string, g *sm.Group, err error,
) {
	if !t.IsConnected() {
		err = ErrHMSDSPtrClosed
		return
	}
	// Generate query
	query := sq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", sm.NormalizeGroupField(label)).
		Where("namespace = ?", groupNamespace)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetEmptyGroupTx(%s): query failed: %s", label, err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		uuid, g, err = t.hdb.scanPgGroup(rows)
		if err != nil {
			t.LogAlways("Error: GetEmptyGroupTx(%s): Scan failed: %s",
				label, err)
			return
		}
		t.Log(LOG_DEBUG, "Debug: GetEmptyGroupTx() scanned (%s, %v)", uuid, g)
	}
	return
}

//
// Partitions
//

// Creates new partition  in component groups, but adds nothing to the members
// table (in tx, so this can be done in separate query)
//
// Returns: (new UUID string, new partition's official name, error)
func (t *hmsdbPgTx) InsertEmptyPartitionTx(p *sm.Partition) (
	string, string, error,
) {
	var err error
	pi := new(compGroupsInsert)

	if !t.IsConnected() {
		return "", "", ErrHMSDSPtrClosed
	}
	// Normalize and verify fields (note these functions track if this
	// has been done and only does each once.)
	p.Normalize()
	if err = p.Verify(); err != nil {
		return "", "", err
	}
	// Set fields for update
	pi.id = uuid.New().String()    // Used only internally for joins
	pi.name = p.Name               // Must be unique accross all partitions
	pi.description = p.Description // Free-form shortish string

	pi.tags = p.Tags
	pi.gtype = partType          // 'Partition' type (vs. grp/exGrp)
	pi.namespace = partNamespace // 'Partition' vs. 'Group'
	pi.exclusiveGroupId = ""     // Implicitly exclusive

	// Generate query
	query := sq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(pi.id, pi.name, pi.description,
			pq.Array(pi.tags), pi.gtype, pi.namespace, pi.exclusiveGroupId)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return pi.id, pi.name, ParsePgDBError(err)
}

// Update fields in PartitionPatch on the returned partition object provided
// (in transaction).
func (t *hmsdbPgTx) UpdateEmptyPartitionTx(
	uuid string,
	p *sm.Partition,
	pp *sm.PartitionPatch,
) error {
	var err error
	var doUpdate bool

	if p == nil || pp == nil {
		return nil
	}
	// Start update query string
	update := sq.Update("").
		Table(compGroupsTable).
		Where(sq.Eq{compGroupIdCol: uuid})

	// Check to see if there are any fields set in the update and then
	// see if they need to be updated.
	if pp.Description != nil && p.Description != *pp.Description {
		update = update.Set(compGroupDescCol, *pp.Description)
		doUpdate = true
	}
	if pp.Tags != nil {
		inTagLen := len(*pp.Tags)
		if inTagLen != len(p.Tags) {
			doUpdate = true
			update = update.Set(compGroupTagsCol, pq.Array(pp.Tags))
		} else {
			// Same array length - check individual entries and update if
			// they don't match.
			gotMismatch := false
			for i, str := range *pp.Tags {
				if p.Tags[i] != str {
					gotMismatch = true
				}
			}
			// One or more tags did not match -  note we will update if
			// the order changes, but the ordering should always get returned
			// the same way, so the client should have to go out of it's way
			// to change it.
			if gotMismatch == true {
				update = update.Set(compGroupTagsCol, pq.Array(pp.Tags))
				doUpdate = true
			}
		}
	}
	// Have a change to make...
	if doUpdate == true {
		// Exec with statement cache for caching prepared statements
		update = update.PlaceholderFormat(sq.Dollar)
		_, err = update.RunWith(t.sc).ExecContext(t.ctx)
	}
	return err
}

// Get the user-readable fields in a partition entry and it's internal uuid but
// don't fetch its members (done in transaction, so we can fetch them as part
// of the same one).
func (t *hmsdbPgTx) GetEmptyPartitionTx(name string) (
	uuid string, p *sm.Partition, err error,
) {
	if !t.IsConnected() {
		err = ErrHMSDSPtrClosed
		return
	}
	// Generate query
	query := sq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", sm.NormalizeGroupField(name)).
		Where("namespace = ?", partNamespace)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetEmptyPartitionTx(%s): query failed: %s",
			name, err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		uuid, p, err = t.hdb.scanPgPartition(rows)
		if err != nil {
			t.LogAlways("Error: GetEmptyPartitionTx(%s): Scan failed: %s",
				name, err)
			return
		}
		t.Log(LOG_DEBUG, "Debug: GetEmptyPartitionTx() scanned (%s, %v)",
			uuid, p)
	}
	return
}

//
// Members (for either Group/Partition)
//

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
func (t *hmsdbPgTx) InsertMembersTx(uuid, namespace string, ms *sm.Members) error {
	mi := new(compGroupMembersInsertNoTS)

	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	if len(ms.IDs) == 0 {
		return nil
	}
	// Normalize and verify xname ids (note: these functions track if this
	// has been done and only does each once.)
	ms.Normalize()
	if err := ms.Verify(); err != nil {
		return err
	}
	// Use uuid of group for joins, etc.
	mi.group_id = uuid

	// group_namespace is either the (clash-avoidant) exclusive group or the
	// regular group if there is no exclusive one.  Each xname should be
	// unique in this namespace or a unique-constraint conflict error will
	// occur.
	mi.group_namespace = namespace

	// Generate query
	query := sq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...)

	// Append members
	for _, id := range ms.IDs {
		query = query.Values(id, mi.group_id, mi.group_namespace)
	}
	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err := query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

// UUID string should be as retried from one of the group/partition calls.  No
// guarantees made about alternate formatting of the underlying binary value.
func (t *hmsdbPgTx) GetMembersTx(uuid string) (*sm.Members, error) {
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	// Generate query
	ms := sm.NewMembers()
	query := sq.Select(compGroupMembersColsUser...).
		From(compGroupMembersTable).
		Where("group_id = ?", uuid)

	// Query with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetMembersTx(%s): query failed: %s", uuid, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.LogAlways("Error: GetMembersTx(%s): scan failed: %s", uuid, err)
			return nil, err
		}
		ms.IDs = append(ms.IDs, id)
	}
	return ms, err
}

// UUID string should be as retried from one of the group/partition calls.  No
// guarantees made about alternate formatting of the underlying binary value.
// Result is uuid members, but with entries not also in and_uuid removed.
func (t *hmsdbPgTx) GetMembersFilterTx(uuid, and_uuid string) (*sm.Members, error) {
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	// Generate query
	ms := sm.NewMembers()
	query := sq.Select("a.component_id").
		From(compGroupMembersTable+" a").
		Join(compGroupMembersTable+" b ON a.component_id = b.component_id").
		Where("a.group_id = ?", uuid)

	if and_uuid != "" {
		query = query.Where("b.group_id IN (?,?)", uuid, and_uuid).
			GroupBy("a.component_id").
			Having("COUNT(*) = 2")
	} else {
		query = query.Where(sq.Or{sq.Expr("b.group_id = ?", uuid),
			sq.Expr("b.group_namespace = ?", partGroupNamespace)}).
			GroupBy("a.component_id").
			Having("COUNT(*) = 1")
	}
	// Query with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetMembersFilterTx(%s): query failed: %s",
			uuid, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.LogAlways("Error: GetMembersFilterTx(%s): scan failed: %s",
				uuid, err)
			return nil, err
		}
		ms.IDs = append(ms.IDs, id)
	}
	return ms, err
}

// Given an internal group_id uuid, delete the given id, if it exists.
// if it does not, result will be false, nil vs. true,nil on deletion.
func (t *hmsdbPgTx) DeleteMemberTx(uuid, id string) (bool, error) {
	// Build query - works like AND
	query := sq.Delete(compGroupMembersTable).
		Where("group_id = ?", uuid).
		Where("component_id = ?", xnametypes.NormalizeHMSCompID(id))

	// Execute - Should delete one row.
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
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
				t.LogAlways("Error: DeleteMemberTx(): multiple deletions!")
			}
			return true, nil
		}
	}
	return false, nil
}

// Given an internal group_id uuid, delete all of its members. Returns the
// number of deletions, if any, and any error that may have occurred.
func (t *hmsdbPgTx) DeleteMembersAllTx(guuid string) (int64, error) {
	// Build query
	query := sq.Delete(compGroupMembersTable).Where("group_id = ?", guuid)

	// Execute - Can delete one or more rows
	query = query.PlaceholderFormat(sq.Dollar)
	res, err := query.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		return 0, fmt.Errorf("Error: DeleteMembersAllTx(): failed to exec delete query: %w", err)
	}

	// Calculate number of deletions
	num, err := res.RowsAffected()
	if err != nil {
		err = fmt.Errorf("Error: DeleteMembersAllTx(): failed to get number of affected rows: %w", err)
	}

	return num, err
}

////////////////////////////////////////////////////////////////////////////
//
// Component Lock Management
//
////////////////////////////////////////////////////////////////////////////

// Insert component reservations into the database.
// To Insert reservations without a duration, the component must be locked.
// To Insert reservations with a duration, the component must be unlocked.
func (t *hmsdbPgTx) InsertCompReservationsTx(ids []string, duration int) ([]sm.CompLockV2Success, string, error) {
	var err error
	var expiration_timestamp sql.NullTime
	var results []sm.CompLockV2Success

	if !t.IsConnected() {
		return results, sm.CLResultServerError, ErrHMSDSPtrClosed
	}

	create_timestamp := time.Now()

	// Expiration timestamp is only added if it is an expiring reservation
	if duration > 0 {
		expiration_timestamp.Time = create_timestamp.Add(time.Duration(duration) * time.Minute)
		expiration_timestamp.Valid = true
	} else {
		expiration_timestamp.Valid = false
	}

	// Generate query
	query := sq.Insert(compResTable).
		Columns(compResCols...)

	for _, id := range ids {
		// Set fields for update
		deputy_key := id + ":dk:" + uuid.New().String()      // The new unique public key
		reservation_key := id + ":rk:" + uuid.New().String() // The new unique private key
		query = query.Values(id, create_timestamp, expiration_timestamp, deputy_key, reservation_key)
	}

	query = query.Suffix("ON CONFLICT DO NOTHING RETURNING " + compResCompIdCol + ", " + compResDKCol + ", " + compResRKCol)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		if IsPgDuplicateKeyErr(err) {
			return results, sm.CLResultReserved, nil
		}
		return results, sm.CLResultServerError, err
	}
	defer rows.Close()
	for rows.Next() {
		var result sm.CompLockV2Success
		err = rows.Scan(
			&result.ID,
			&result.DeputyKey,
			&result.ReservationKey,
		)
		if expiration_timestamp.Valid {
			result.ExpirationTime = expiration_timestamp.Time.Format(time.RFC3339)
		}
		results = append(results, result)
	}

	return results, sm.CLResultSuccess, nil
}

// Remove/release component reservations.
// Both a component ID and reservation key are required for these operations unless force = true.
// Returns an array of xnames associated with the reservations.
func (t *hmsdbPgTx) DeleteCompReservationsTx(rKeys []sm.CompLockV2Key, force bool) ([]string, error) {
	var results []string
	var ids []string
	var keys []string

	if !t.IsConnected() {
		return results, ErrHMSDSPtrClosed
	}

	for _, rKey := range rKeys {
		ids = append(ids, rKey.ID)
		if rKey.Key == "" && !force {
			return results, sm.ErrCompLockV2DKey
		}
		keys = append(keys, rKey.Key)
	}

	// Build query - works like AND
	query := sq.Delete(compResTable)
	if force {
		if len(ids) > 0 {
			query = query.Where(sq.Eq{compResCompIdCol: ids})
		}
	} else {
		if len(keys) > 0 {
			// Only need the keys because the id is part of the key.
			query = query.Where(sq.Eq{compResRKCol: keys})
		} else {
			return results, sm.ErrCompLockV2RKey
		}
	}
	query = query.Suffix("RETURNING " + compResCompIdCol)

	// Execute - Should delete one row.
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	// See if there was a v1LockId associated with
	// the reservation we just deleted.
	for rows.Next() {
		var xname string
		err = rows.Scan(&xname)
		if err != nil {
			return results, err
		}
		results = append(results, xname)
	}
	return results, nil
}

// Release all expired component reservations
func (t *hmsdbPgTx) DeleteCompReservationExpiredTx() ([]string, error) {
	xnames := make([]string, 0, 1)
	if !t.IsConnected() {
		return xnames, ErrHMSDSPtrClosed
	}

	// Build query - works like AND
	query := sq.Delete(compResTable).
		Where(compResExpireCol + " IS NOT NULL AND NOW() >= " + compResExpireCol).
		Suffix("RETURNING " + compResCompIdCol)

	// Execute - Should delete one row.
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		return xnames, err
	}
	defer rows.Close()

	// Process the xnames associated with the reservations we deleted.
	for rows.Next() {
		id := ""

		err = rows.Scan(&id)
		if err != nil {
			return xnames, err
		}

		xnames = append(xnames, id)
	}
	return xnames, nil
}

// Retrieve the status of reservations. The public key and xname is
// required to address the reservation unless force = true.
func (t *hmsdbPgTx) GetCompReservationsTx(dKeys []sm.CompLockV2Key, force bool) ([]sm.CompLockV2Success, string, error) {
	var results []sm.CompLockV2Success
	var ids []string
	var keys []string
	var err error
	if !t.IsConnected() {
		err = ErrHMSDSPtrClosed
		return results, sm.CLResultServerError, err
	}

	for _, dKey := range dKeys {
		ids = append(ids, dKey.ID)
		if dKey.Key == "" && !force {
			return results, sm.CLResultServerError, sm.ErrCompLockV2DKey
		}
		keys = append(keys, dKey.Key)
	}

	// Generate query
	query := sq.Select(addAliasToCols(compResAlias, compResPubCols, compResPubCols)...).
		From(compResTable + " " + compResAlias)
	if force {
		if len(ids) > 0 {
			query = query.Where(sq.Eq{compResCompIdColAlias: ids})
		}
	} else {
		if len(keys) > 0 {
			// Only need the keys because the id is part of the key.
			query = query.Where(sq.Eq{compResDKColAlias: keys})
		} else {
			return results, sm.CLResultServerError, sm.ErrCompLockV2DKey
		}
	}

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetCompReservationsTx(): query failed: %s", err)
		return results, sm.CLResultServerError, err
	}
	defer rows.Close()

	for rows.Next() {
		var cr compReservation
		err = rows.Scan(
			&cr.component_id,
			&cr.create_timestamp,
			&cr.expiration_timestamp,
			&cr.deputy_key,
		)
		if err != nil {
			t.LogAlways("Error: GetCompReservationsTx(): Scan failed: %s", err)
			return results, sm.CLResultServerError, err
		}
		result := sm.CompLockV2Success{
			ID:        cr.component_id,
			DeputyKey: cr.deputy_key,
		}
		if cr.create_timestamp.Valid {
			result.CreationTime = cr.create_timestamp.Time.Format(time.RFC3339)
		}
		if cr.expiration_timestamp.Valid {
			result.ExpirationTime = cr.expiration_timestamp.Time.Format(time.RFC3339)
		}
		results = append(results, result)
	}
	return results, sm.CLResultSuccess, nil
}

// Update/renew the expiration time of component reservations with the given
// ID/Key combinations.
func (t *hmsdbPgTx) UpdateCompReservationsTx(rKeys []sm.CompLockV2Key, duration int, force bool) ([]string, error) {
	var results []string
	var ids []string
	var keys []string
	var err error

	if !t.IsConnected() {
		return results, ErrHMSDSPtrClosed
	}

	for _, rKey := range rKeys {
		ids = append(ids, rKey.ID)
		if rKey.Key == "" && !force {
			return results, sm.ErrCompLockV2RKey
		}
		keys = append(keys, rKey.Key)
	}

	// Start update query string
	update := sq.Update("").
		Table(compResTable).
		Where(compResExpireCol + " IS NOT NULL")
	if force {
		if len(ids) > 0 {
			update = update.Where(sq.Eq{compResCompIdCol: ids})
		}
	} else {
		if len(keys) > 0 {
			// Only need the keys because the id is part of the key.
			update = update.Where(sq.Eq{compResRKCol: keys})
		} else {
			return results, sm.ErrCompLockV2RKey
		}
	}

	expiration_timestamp := time.Now().Add(time.Duration(duration) * time.Minute)
	update = update.Set(compResExpireCol, expiration_timestamp).
		Suffix("RETURNING " + compResCompIdCol)

	// Exec with statement cache for caching prepared statements
	update = update.PlaceholderFormat(sq.Dollar)
	rows, err := update.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	// See if there was a v1LockId associated with
	// the reservation we just updated.
	for rows.Next() {
		var xname string
		err = rows.Scan(&xname)
		if err != nil {
			return results, err
		}
		results = append(results, xname)
	}
	return results, nil
}

// Update component 'ReservationDisabled' field.
func (t *hmsdbPgTx) BulkUpdateCompResDisabledTx(ids []string, disabled bool) ([]string, error) {
	var results []string

	if !t.IsConnected() {
		return results, ErrHMSDSPtrClosed
	}

	update := sq.Update("").
		Table(compTable)
	if len(ids) > 0 {
		update = update.Where(sq.Eq{compIdCol: ids})
	}
	update = update.Set(compResDisabledCol, disabled).
		Suffix("RETURNING " + compIdCol)

	// Exec with statement cache for caching prepared statements
	update = update.PlaceholderFormat(sq.Dollar)
	rows, err := update.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompResDisabledTx(): query failed: %s", err)
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			t.LogAlways("Error: BulkUpdateCompResDisabledTx(): Scan failed: %s", err)
			return results, err
		}
		results = append(results, id)
	}

	return results, nil
}

// Update component 'locked' field.
func (t *hmsdbPgTx) BulkUpdateCompResLockedTx(ids []string, locked bool) ([]string, error) {
	var results []string

	if !t.IsConnected() {
		return results, ErrHMSDSPtrClosed
	}

	update := sq.Update("").
		Table(compTable)
	if len(ids) > 0 {
		update = update.Where(sq.Eq{compIdCol: ids})
	}
	update = update.Set(compLockedCol, locked).
		Suffix("RETURNING " + compIdCol)

	// Exec with statement cache for caching prepared statements
	update = update.PlaceholderFormat(sq.Dollar)
	rows, err := update.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: BulkUpdateCompResLockedTx(): query failed: %s", err)
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			t.LogAlways("Error: BulkUpdateCompResLockedTx(): Scan failed: %s", err)
			return results, err
		}
		results = append(results, id)
	}
	return results, nil
}

////////////////////////////////////////////////////////////////////////////
//
// Job Sync Management
//
////////////////////////////////////////////////////////////////////////////

//
// Jobs
//

// Creates new job entry in the job sync, but adds nothing to the job's
// type table (in tx, so this can be done in separate query)
//
// Returns: (new jobId string, error)
func (t *hmsdbPgTx) InsertEmptyJobTx(j *sm.Job) (string, error) {
	var err error
	ji := new(jobInsertNoTS)

	if !t.IsConnected() {
		return "", ErrHMSDSPtrClosed
	}

	// Set fields for update
	ji.id = uuid.New().String() // The new unique jobId
	ji.jobType = j.Type         // Job type
	ji.status = j.Status        // Job status
	ji.lifetime = j.Lifetime    // Expiration time for the job

	// Generate query
	query := sq.Insert(jobTable).
		Columns(jobColsNoTS...).
		Values(ji.id, ji.jobType, ji.status, ji.lifetime)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err = query.RunWith(t.sc).ExecContext(t.ctx)
	return ji.id, ParsePgDBError(err)
}

// Update the status and LastUpdated fields for a Job entry (in transaction).
func (t *hmsdbPgTx) UpdateEmptyJobTx(jobId string, status string) (bool, error) {
	var err error

	if !t.IsConnected() {
		return false, ErrHMSDSPtrClosed
	}

	if len(jobId) == 0 {
		return false, ErrHMSDSArgEmpty
	}

	// Start update query string
	update := sq.Update("").
		Table(jobTable).
		Where(sq.Eq{jobIdCol: jobId})

	// Check to see if there are any fields set in the update and then
	// see if they need to be updated.
	if len(status) > 0 {
		update = update.Set(jobStatusCol, status)
	}

	// Always update the timestamp
	update = update.Set(jobLastUpdateCol, "NOW()")

	// Exec with statement cache for caching prepared statements
	update = update.PlaceholderFormat(sq.Dollar)
	res, err := update.RunWith(t.sc).ExecContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: UpdateEmptyJobTx(): stmt.Exec: %s", err)
		return false, err
	}

	// Return true if there was a row affected, false if there were zero.
	num, err := res.RowsAffected()
	if err != nil {
		return false, err
	} else if num > 0 {
		return true, nil
	}
	return false, nil
}

// Get the user-readable fields in a job entry but don't fetch its job type
// specific data (done in transaction, so we can fetch them as part of the
// same one).
func (t *hmsdbPgTx) GetEmptyJobTx(jobId string) (j *sm.Job, err error) {
	if !t.IsConnected() {
		err = ErrHMSDSPtrClosed
		return
	}

	if len(jobId) == 0 {
		err = ErrHMSDSArgEmpty
		return
	}

	// Generate query
	query := sq.Select(jobCols...).
		From(jobTable).
		Where(jobIdCol+" = ?", jobId)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetEmptyJobTx(%s): query failed: %s",
			jobId, err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		j, err = t.hdb.scanPgJob(rows)
		if err != nil {
			t.LogAlways("Error: GetEmptyJobTx(%s): Scan failed: %s",
				jobId, err)
			return
		}
		t.Log(LOG_DEBUG, "Debug: GetEmptyJobTx(%s) scanned (%v)",
			jobId, j)
	}
	return
}

// Get the user-readable fields in a job entry but don't fetch its job type
// specific data (done in transaction, so we can fetch them as part of the
// same one).
func (t *hmsdbPgTx) GetEmptyJobsTx(f_opts ...JobSyncFiltFunc) (js []*sm.Job, err error) {
	var j *sm.Job
	if !t.IsConnected() {
		err = ErrHMSDSPtrClosed
		return
	}
	f := new(JobSyncFilter)
	for _, opts := range f_opts {
		opts(f)
	}
	js = make([]*sm.Job, 0, 1)
	// Generate query
	query := sq.Select(addAliasToCols(jobAlias, jobCols, jobCols)...).
		From(jobTable + " " + jobAlias)
	// Filter by jobId
	if f.ID != nil && len(f.ID) != 0 {
		query = query.Where(sq.Eq{jobIdColAlias: f.ID})
	}
	// Filter by jobType
	if f.Type != nil && len(f.Type) != 0 {
		query = query.Where(sq.Eq{jobTypeColAlias: f.Type})
	}

	// Filter by job status
	if f.Status != nil && len(f.Status) != 0 {
		query = query.Where(sq.Eq{jobStatusColAlias: f.Status})
	}

	if f.isExpired {
		query = query.Where("NOW()-" + jobLastUpdateColAlias +
			" >= (" + jobLifetimeColAlias + " * '1 sec'::interval)")
	}

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetEmptyJobsTx(): query failed: %s", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		j, err = t.hdb.scanPgJob(rows)
		if err != nil {
			t.LogAlways("Error: GetEmptyJobsTx(): Scan failed: %s", err)
			return
		}
		t.Log(LOG_DEBUG, "Debug: GetEmptyJobsTx() scanned (%v)", j)
		js = append(js, j)
	}
	return
}

//
// State Redfish Poll Jobs
//

// Insert job specific info for the given jobId. The jobId parameter should
// be as-returned by InsertEmptyJobTx()/InsertEmptyJobTx().
func (t *hmsdbPgTx) InsertStateRFPollJobTx(jobId string, data *sm.SrfpJobData) error {
	if !t.IsConnected() {
		return ErrHMSDSPtrClosed
	}
	if len(jobId) == 0 || data == nil || len(data.CompId) == 0 {
		return ErrHMSDSArgMissing
	}

	// Generate query
	query := sq.Insert(stateRfPollTable).
		Columns(stateRfPollCols...).
		Values(data.CompId, jobId)

	// Exec with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	_, err := query.RunWith(t.sc).ExecContext(t.ctx)
	return ParsePgDBError(err)
}

// Get the job specific info associated with the jobId. The jobId string should
// be as retried from one of the Job calls.  No guarantees made about
// alternate formatting of the underlying binary value.
func (t *hmsdbPgTx) GetStateRFPollJobByIdTx(jobId string) (*sm.SrfpJobData, error) {
	if !t.IsConnected() {
		return nil, ErrHMSDSPtrClosed
	}
	// Generate query
	data := new(sm.SrfpJobData)
	query := sq.Select(stateRfPollCmpIdCol).
		From(stateRfPollTable).
		Where(stateRfPollJobIdCol+" = ?", jobId)

	// Query with statement cache for caching prepared statements (local to tx)
	query = query.PlaceholderFormat(sq.Dollar)
	rows, err := query.RunWith(t.sc).QueryContext(t.ctx)
	if err != nil {
		t.LogAlways("Error: GetStateRFPollJobByIdTx(%s): query failed: %s", jobId, err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&data.CompId); err != nil {
			t.LogAlways("Error: GetStateRFPollJobByIdTx(%s): scan failed: %s", jobId, err)
			return nil, err
		}
	}
	return data, err
}
