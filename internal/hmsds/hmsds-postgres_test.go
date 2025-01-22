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
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
	stest "github.com/Cray-HPE/hms-smd/v2/pkg/sharedtest"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

//
// Unit Tests
//

// Global setup of dPG and mockPG in hmsds-api_test.go runs first via TestMain()

const tGetCompBaseQuery = "SELECT c.id AS id, c.type AS type, c.state AS state," +
	" c.flag AS flag, c.enabled AS enabled, c.admin AS admin, c.role AS role," +
	" c.subrole AS subrole, c.nid AS nid, c.subtype AS subtype, c.nettype AS nettype," +
	" c.arch AS arch, c.class AS class, c.reservation_disabled AS reservation_disabled," +
	" c.locked AS locked FROM components c"

const tGetCompNidOnlyQuery = "SELECT c.id AS id, c.type AS type, c.nid AS nid" +
	" FROM components c"

const tGetCompStateOnlyQuery = "SELECT c.id AS id, c.type AS type, " +
	"c.state AS state, c.flag AS flag FROM components c"

const tGetCompQueryHierAll = "SELECT comp.id AS id, comp.type AS type, " +
	"comp.state AS state, comp.flag AS flag, comp.enabled AS enabled, " +
	"comp.admin AS admin, comp.role AS role, comp.subrole AS subrole, " +
	"comp.nid AS nid, comp.subtype AS subtype, comp.nettype AS nettype, " +
	"comp.arch AS arch, comp.class AS class, comp.reservation_disabled AS reservation_disabled, " +
	"comp.locked AS locked FROM ("

const tGetCompQueryHierStateOnly = "SELECT comp.id AS id, comp.type AS type, " +
	"comp.state AS state, comp.flag AS flag FROM ("

const tGetMembershipsBaseQuery = "SELECT c.id AS id, cg.name AS name, " +
	"cg.namespace AS namespace FROM components c"

const tGetMembershipsBaseQueryWSub = "SELECT comp.id AS id, compg.name AS name, " +
	"compg.namespace AS namespace FROM components comp"

const tGetMembershipsBaseSuffixWSub = " GROUP BY c.id )"

const tGetMembershipsGroupSuffixWSub = " GROUP BY c.id HAVING COUNT(*) = 2 )"

const tGetMembershipsSubQuery = " LEFT JOIN component_group_members compgm " +
	"ON compgm.component_id = comp.id LEFT JOIN component_groups compg " +
	"ON compg.id = compgm.group_id WHERE comp.id IN ( SELECT c.id AS id FROM " +
	"components c"

const tGetCompJoinGroupsQuery = " LEFT JOIN component_group_members cgm " +
	"ON cgm.component_id = c.id " +
	"LEFT JOIN component_groups cg ON cg.id = cgm.group_id"

const tGetCompJoinGroupsSuffix = " GROUP BY c.id"

const tGetCompJoinGroupsSuffixAnd = " GROUP BY c.id HAVING COUNT(*) = 2"

const tGetSCNSubscriptionQueryId = "SELECT id, subscription FROM scn_subscriptions WHERE id = $1"

const tGetSCNSubscriptionQueryAll = "SELECT id, subscription FROM scn_subscriptions"

const tGetSCNSubscriptionUpdate = "SELECT id, subscription FROM scn_subscriptions WHERE id = $1 FOR UPDATE"

const tInsertSCNSubscription = "INSERT INTO scn_subscriptions ( sub_url, subscription) VALUES ($1, $2)"

const tInsertSCNSubscriptionQueryLastVal = "SELECT LASTVAL()"

const tUpdateSCNSubscription = "UPDATE scn_subscriptions SET sub_url = $1, subscription = $2 WHERE id = $3"

const tDeleteSCNSubscription = "DELETE FROM scn_subscriptions WHERE id = $1"

const tDeleteSCNSubscriptionAll = "DELETE FROM scn_subscriptions"

func TestPgGetComponentsFilter(t *testing.T) {
	enabledFlg := true
	tests := []struct {
		f               *ComponentFilter
		fieldFltr       FieldFilter
		dbColumns       []string
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedComps   []*base.Component
	}{{
		&ComponentFilter{
			Type: []string{"node"},
		},
		FLTR_DEFAULT,
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.type IN ($1)"),
		[]driver.Value{"Node"},
		[]*base.Component{
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{
			Partition: []string{"part1"},
		},
		FLTR_DEFAULT,
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompBaseQuery + tGetCompJoinGroupsQuery + " WHERE (cg.name IN ($1) AND cg.namespace = $2)"),
		[]driver.Value{"part1", partNamespace},
		[]*base.Component{
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{
			Group: []string{"grp1"},
		},
		FLTR_DEFAULT,
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompBaseQuery + tGetCompJoinGroupsQuery + " WHERE (cg.name IN ($1) AND cg.namespace = $2)"),
		[]driver.Value{"grp1", groupNamespace},
		[]*base.Component{
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{
			Group: []string{"grp1", "grp2"},
		},
		FLTR_DEFAULT,
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompBaseQuery + tGetCompJoinGroupsQuery + " WHERE (cg.name IN ($1,$2) AND cg.namespace = $3)"),
		[]driver.Value{"grp1", "grp2", groupNamespace},
		[]*base.Component{
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{
			NID:      []string{"800"},
			NIDStart: []string{"16300", "440"},
			NIDEnd:   []string{"16500", "480"},
		},
		FLTR_NIDONLY,
		[]string{"id", "type", "nid"},
		[][]driver.Value{
			[]driver.Value{"x0c0s14b0n0", "Node", 448},
			[]driver.Value{"x0c0s15b0n0", "Node", 480},
			[]driver.Value{"x0c0s25b0n0", "Node", 800},
			[]driver.Value{"x1c0s0b0n0", "Node", 16384},
			[]driver.Value{"x1c0s1b0n0", "Node", 16416},
			[]driver.Value{"x1c0s2b0n0", "Node", 16448},
		},
		nil,
		regexp.QuoteMeta(tGetCompNidOnlyQuery + " WHERE ((c.nid >= $1 AND c.nid <= $2) OR (c.nid >= $3 AND c.nid <= $4) OR c.nid = $5)"),
		[]driver.Value{"16300", "16500", "440", "480", "800"},
		[]*base.Component{
			&base.Component{"x0c0s14b0n0", "Node", "", "", nil, "", "", "", "448", "", "", "", "", false, false},
			&base.Component{"x0c0s15b0n0", "Node", "", "", nil, "", "", "", "480", "", "", "", "", false, false},
			&base.Component{"x0c0s25b0n0", "Node", "", "", nil, "", "", "", "800", "", "", "", "", false, false},
			&base.Component{"x1c0s0b0n0", "Node", "", "", nil, "", "", "", "16384", "", "", "", "", false, false},
			&base.Component{"x1c0s1b0n0", "Node", "", "", nil, "", "", "", "16416", "", "", "", "", false, false},
			&base.Component{"x1c0s2b0n0", "Node", "", "", nil, "", "", "", "16448", "", "", "", "", false, false},
		},
	}, {
		nil,
		FLTR_DEFAULT,
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s14b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s14b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 448, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s15b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s15b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 480, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s18b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s18b0n0", "Node", "Off", "OK", true, "AdminStatus", "Compute", "", 576, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s22b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s22b0n0", "Node", "Off", "OK", true, "AdminStatus", "Compute", "", 704, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s24b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s24b0n0", "Node", "Off", "OK", true, "AdminStatus", "Compute", "", 786, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s25b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s25b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 800, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		tGetCompBaseQuery,
		nil,
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
	}, {
		&ComponentFilter{},
		FLTR_STATEONLY,
		[]string{"id", "type", "state", "flag"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK"},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK"},
		},
		sql.ErrNoRows,
		regexp.QuoteMeta(tGetCompStateOnlyQuery),
		nil,
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
	}, {
		nil,
		FLTR_STATEONLY,
		[]string{"id", "type", "state", "flag"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK"},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK"},
		},
		nil,
		regexp.QuoteMeta(tGetCompStateOnlyQuery),
		nil,
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
	}, {
		&ComponentFilter{
			Group:     []string{"grp1"},
			Partition: []string{"part1"},
		},
		FLTR_STATEONLY,
		[]string{"id", "type", "state", "flag"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK"},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK"},
		},
		nil,
		regexp.QuoteMeta(tGetCompStateOnlyQuery + tGetCompJoinGroupsQuery +
			" WHERE ((cg.name IN ($1) AND cg.namespace = $2) OR (cg.name IN ($3) AND cg.namespace = $4))" +
			tGetCompJoinGroupsSuffixAnd),
		[]driver.Value{"grp1", groupNamespace, "part1", partNamespace},
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			if test.expectedArgs != nil {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
			mockPG.ExpectCommit()
		}

		comps, err := dPG.GetComponentsFilter(test.f, test.fieldFltr)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !stest.CompareComps(test.expectedComps, comps) {
				t.Errorf("Test %v Failed: Expected compArray '%v'; Recieved compArray '%v'", i, test.expectedComps, comps)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetComponentsQuery(t *testing.T) {
	enabledFlg := true
	tests := []struct {
		f               *ComponentFilter
		fieldFltr       FieldFilter
		ids             []string
		dbColumns       []string
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedComps   []*base.Component
	}{{
		&ComponentFilter{
			Type: []string{"node"},
		},
		FLTR_DEFAULT,
		[]string{"x0c0s26", "x0c0s27"},
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompQueryHierAll + tGetCompBaseQuery + " WHERE c.type IN ($1)) AS comp WHERE (comp.id SIMILAR TO $2) OR (comp.id SIMILAR TO $3)"),
		[]driver.Value{"Node", "x0c0s26([[:alpha:]][[:alnum:]]*)?", "x0c0s27([[:alpha:]][[:alnum:]]*)?"},
		[]*base.Component{
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{
			Type:  []string{"node"},
			Group: []string{"grp1"},
		},
		FLTR_DEFAULT,
		[]string{"x0c0s26", "x0c0s27"},
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompQueryHierAll + tGetCompBaseQuery + tGetCompJoinGroupsQuery +
			" WHERE c.type IN ($1) AND (cg.name IN ($2) AND cg.namespace = $3) GROUP BY c.id) AS comp" +
			" WHERE (comp.id SIMILAR TO $4) OR (comp.id SIMILAR TO $5)"),
		[]driver.Value{"Node", "grp1", groupNamespace, "x0c0s26([[:alpha:]][[:alnum:]]*)?", "x0c0s27([[:alpha:]][[:alnum:]]*)?"},
		[]*base.Component{
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{
			Type:      []string{"node"},
			Group:     []string{"grp1"},
			Partition: []string{"part1"},
		},
		FLTR_DEFAULT,
		[]string{"x0c0s26", "x0c0s27"},
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s27b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 864, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompQueryHierAll + tGetCompBaseQuery + tGetCompJoinGroupsQuery +
			" WHERE c.type IN ($1) AND ((cg.name IN ($2) AND cg.namespace = $3) OR (cg.name IN ($4) AND cg.namespace = $5))" +
			" GROUP BY c.id HAVING COUNT(*) = 2) AS comp" +
			" WHERE (comp.id SIMILAR TO $6) OR (comp.id SIMILAR TO $7)"),
		[]driver.Value{"Node", "grp1", groupNamespace, "part1", partNamespace, "x0c0s26([[:alpha:]][[:alnum:]]*)?", "x0c0s27([[:alpha:]][[:alnum:]]*)?"},
		[]*base.Component{
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s27b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "864", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{
			Type: []string{"node", "nodebmc"},
		},
		FLTR_DEFAULT,
		[]string{"x0c0s26"},
		[]string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK", true, "", "", "", -1, "", "Sling", "X86", "", false, false},
			[]driver.Value{"x0c0s26b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
		},
		nil,
		regexp.QuoteMeta(tGetCompQueryHierAll + tGetCompBaseQuery + " WHERE c.type IN ($1,$2)) AS comp WHERE (comp.id SIMILAR TO $3)"),
		[]driver.Value{"Node", "NodeBMC", "x0c0s26([[:alpha:]][[:alnum:]]*)?"},
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", &enabledFlg, "", "", "", "", "", "Sling", "X86", "", false, false},
			&base.Component{"x0c0s26b0n0", "Node", "On", "OK", &enabledFlg, "AdminStatus", "Compute", "", "832", "", "Sling", "X86", "", false, false},
		},
	}, {
		&ComponentFilter{},
		FLTR_STATEONLY,
		[]string{"x0c0s26b0", "x0c0s27b0"},
		[]string{"id", "type", "state", "flag"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK"},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK"},
		},
		nil,
		regexp.QuoteMeta(tGetCompQueryHierStateOnly + tGetCompStateOnlyQuery + ") AS comp WHERE comp.id IN ($1,$2)"),
		[]driver.Value{"x0c0s26b0", "x0c0s27b0"},
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
	}, {
		&ComponentFilter{},
		FLTR_STATEONLY,
		[]string{"x0c0s26b0", "x0c0s27b0"},
		[]string{"id", "type", "state", "flag"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK"},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK"},
		},
		sql.ErrNoRows,
		regexp.QuoteMeta(tGetCompQueryHierStateOnly + tGetCompStateOnlyQuery + ") AS comp WHERE comp.id IN ($1,$2)"),
		[]driver.Value{"x0c0s26b0", "x0c0s27b0"},
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
	}, {
		nil,
		FLTR_STATEONLY,
		[]string{},
		[]string{"id", "type", "state", "flag"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK"},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK"},
		},
		nil,
		//regexp.QuoteMeta(getCompStatePrefix + getCompPrefix + suffixCompFilter),
		regexp.QuoteMeta(tGetCompQueryHierStateOnly + tGetCompStateOnlyQuery + ") AS comp"),
		nil,
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
	}, {
		&ComponentFilter{
			Group: []string{"grp1"},
		},
		FLTR_STATEONLY,
		[]string{},
		[]string{"id", "type", "state", "flag"},
		[][]driver.Value{
			[]driver.Value{"x0c0s26b0", "NodeBMC", "Ready", "OK"},
			[]driver.Value{"x0c0s27b0", "NodeBMC", "Ready", "OK"},
		},
		nil,
		regexp.QuoteMeta(tGetCompQueryHierStateOnly + tGetCompStateOnlyQuery + tGetCompJoinGroupsQuery +
			" WHERE (cg.name IN ($1) AND cg.namespace = $2) GROUP BY c.id) AS comp"),
		[]driver.Value{"grp1", groupNamespace},
		[]*base.Component{
			&base.Component{"x0c0s26b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
			&base.Component{"x0c0s27b0", "NodeBMC", "Ready", "OK", nil, "", "", "", "", "", "", "", "", false, false},
		},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			if test.expectedArgs != nil {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
			mockPG.ExpectCommit()
		}

		comps, err := dPG.GetComponentsQuery(test.f, test.fieldFltr, test.ids)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !stest.CompareComps(test.expectedComps, comps) {
				t.Errorf("Test %v Failed: Expected compArray '%v'; Recieved compArray '%v'", i, test.expectedComps, comps)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgUpsertComponents(t *testing.T) {
	insertComponentsSuffix := "ON CONFLICT(" + compIdCol + ") DO UPDATE SET " +
		compStateCol + " = EXCLUDED." + compStateCol + ", " +
		compFlagCol + " = EXCLUDED." + compFlagCol + ", " +
		compSubTypeCol + " = EXCLUDED." + compSubTypeCol + ", " +
		compNetTypeCol + " = EXCLUDED." + compNetTypeCol + ", " +
		compArchCol + " = EXCLUDED." + compArchCol + ", " +
		compClassCol + " = EXCLUDED." + compClassCol +
		" RETURNING " + compIdCol

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	insert1, _, _ := sqq.Insert(compTable).
		Columns(compColsDefault...).
		Values("x0c0s0b0n0", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false).
		Suffix(insertComponentsSuffix).ToSql()
	insert2, _, _ := sqq.Insert(compTable).
		Columns(compColsDefault...).
		Values("x0c0s0b0n0", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false).
		Values("x0c0s0b0n1", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false).
		Suffix(insertComponentsSuffix).ToSql()

	tests := []struct {
		comps           []*base.Component
		force           bool
		dbColumns       []string
		dbRows          [][]driver.Value
		dbError         error
		dbInsertErr     error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedInsert  string
		expectedRows    [][]driver.Value
		expectedInArgs  []driver.Value
		expectedMap     map[string]map[string]bool
	}{{
		comps: []*base.Component{&base.Component{
			ID:    "x0c0s0b0n0",
			Type:  base.Node.String(),
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}},
		force:           false,
		dbColumns:       []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbRows:          [][]driver.Value{},
		dbError:         nil,
		dbInsertErr:     nil,
		expectedPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedArgs:    []driver.Value{"x0c0s0b0n0"},
		expectedInsert:  regexp.QuoteMeta(insert1),
		expectedRows: [][]driver.Value{
			[]driver.Value{"x0c0s0b0n0"},
		},
		expectedInArgs: []driver.Value{
			"x0c0s0b0n0", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false,
		},
		expectedMap: map[string]map[string]bool{
			"x0c0s0b0n0": map[string]bool{
				"state":   true,
				"flag":    true,
				"enabled": true,
				"role":    true,
				"subRole": true,
				"nid":     true,
			},
		},
	}, {
		comps: []*base.Component{&base.Component{
			ID:    "x0c0s0b0n0",
			Type:  base.Node.String(),
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}},
		force:     false,
		dbColumns: []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s0b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
		},
		dbError:         nil,
		dbInsertErr:     nil,
		expectedPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedArgs:    []driver.Value{"x0c0s0b0n0"},
		expectedInsert:  "",
		expectedRows:    [][]driver.Value{},
		expectedInArgs:  []driver.Value{},
		expectedMap:     map[string]map[string]bool{},
	}, {
		comps: []*base.Component{&base.Component{
			ID:    "x0c0s0b0n0",
			Type:  base.Node.String(),
			State: base.StateEmpty.String(),
			Flag:  base.FlagOK.String(),
		}},
		force:     true,
		dbColumns: []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s0b0n0", "Node", "On", "OK", true, "AdminStatus", "Compute", "", 832, "", "Sling", "X86", "", false, false},
		},
		dbError:         nil,
		dbInsertErr:     nil,
		expectedPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedArgs:    []driver.Value{"x0c0s0b0n0"},
		expectedInsert:  regexp.QuoteMeta(insert1),
		expectedRows: [][]driver.Value{
			[]driver.Value{"x0c0s0b0n0"},
		},
		expectedInArgs: []driver.Value{
			"x0c0s0b0n0", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false,
		},
		expectedMap: map[string]map[string]bool{
			"x0c0s0b0n0": map[string]bool{
				"state": true,
			},
		},
	}, {
		comps: []*base.Component{
			&base.Component{
				ID:    "x0c0s0b0n0",
				Type:  base.Node.String(),
				State: base.StateEmpty.String(),
				Flag:  base.FlagOK.String(),
			},
			&base.Component{
				ID:    "x0c0s0b0n1",
				Type:  base.Node.String(),
				State: base.StateEmpty.String(),
				Flag:  base.FlagOK.String(),
			},
		},
		force:           false,
		dbColumns:       []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", ""},
		dbRows:          [][]driver.Value{},
		dbError:         nil,
		dbInsertErr:     nil,
		expectedPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1,$2)"),
		expectedArgs:    []driver.Value{"x0c0s0b0n0", "x0c0s0b0n1"},
		expectedInsert:  regexp.QuoteMeta(insert2),
		expectedRows: [][]driver.Value{
			[]driver.Value{"x0c0s0b0n0"},
			[]driver.Value{"x0c0s0b0n1"},
		},
		expectedInArgs: []driver.Value{
			"x0c0s0b0n0", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false,
			"x0c0s0b0n1", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false,
		},
		expectedMap: map[string]map[string]bool{
			"x0c0s0b0n0": map[string]bool{
				"state":   true,
				"flag":    true,
				"enabled": true,
				"role":    true,
				"subRole": true,
				"nid":     true,
			},
			"x0c0s0b0n1": map[string]bool{
				"state":   true,
				"flag":    true,
				"enabled": true,
				"role":    true,
				"subRole": true,
				"nid":     true,
			},
		},
	}, {
		comps: []*base.Component{
			&base.Component{
				ID:    "x0c0s0b0n0",
				Type:  base.Node.String(),
				State: base.StateEmpty.String(),
				Flag:  base.FlagOK.String(),
			},
		},
		force:           false,
		dbColumns:       []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbRows:          [][]driver.Value{},
		dbError:         sql.ErrNoRows,
		dbInsertErr:     nil,
		expectedPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedArgs:    []driver.Value{"x0c0s0b0n0"},
		expectedInsert:  "",
		expectedRows:    [][]driver.Value{},
		expectedInArgs:  []driver.Value{},
		expectedMap:     nil,
	}, {
		comps: []*base.Component{
			&base.Component{
				ID:    "x0c0s0b0n0",
				Type:  base.Node.String(),
				State: base.StateEmpty.String(),
				Flag:  base.FlagOK.String(),
			},
		},
		force:           false,
		dbColumns:       []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbRows:          [][]driver.Value{},
		dbError:         nil,
		dbInsertErr:     sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedArgs:    []driver.Value{"x0c0s0b0n0"},
		expectedInsert:  regexp.QuoteMeta(insert1),
		expectedRows: [][]driver.Value{
			[]driver.Value{"x0c0s0b0n0"},
		},
		expectedInArgs: []driver.Value{
			"x0c0s0b0n0", "Node", "Empty", "OK", true, "", "", "", -1, "", "", "", "", false, false,
		},
		expectedMap: nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}
		insertRows := sqlmock.NewRows([]string{"id"})
		for _, row := range test.expectedRows {
			insertRows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			if test.expectedInsert != "" {
				if test.dbInsertErr != nil {
					mockPG.ExpectPrepare(test.expectedInsert).ExpectQuery().WillReturnError(test.dbInsertErr)
					mockPG.ExpectRollback()
				} else {
					mockPG.ExpectPrepare(test.expectedInsert).ExpectQuery().WithArgs(test.expectedInArgs...).WillReturnRows(insertRows)
				}
			}
			if test.dbInsertErr == nil {
				mockPG.ExpectCommit()
			}
		}

		changeMap, err := dPG.UpsertComponents(test.comps, test.force)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.dbInsertErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedMap, changeMap) {
				t.Errorf("Test %v Failed: Expected changeMap '%v'; Received changeMap '%v'", i, test.expectedMap, changeMap)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgUpdateCompStates(t *testing.T) {
	tests := []struct {
		ids                   []string
		state                 string
		flag                  string
		force                 bool
		pi                    *PartInfo
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedAffectedIDs   []string
	}{{
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
		"on",
		"",
		false,
		new(PartInfo),
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND ((state != $3) OR (flag != $4)) AND (state = $5 OR state = $6 OR state = $7 OR state = $8);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "On", "OK", "On", "Off", "Standby", "Halt"},
		nil,
		regexp.QuoteMeta(updateCompPrefix +
			" state = 'On', flag = CASE WHEN flag = 'Locked' THEN 'Locked' ELSE 'OK' END WHERE (id = $1);"),
		[]driver.Value{"x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}, {
		[]string{"x0c0s27b0n0"},
		"Halt",
		"alert",
		false,
		new(PartInfo),
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		"", // not called
		[]driver.Value{"x0c0s27b0n0", "Halt", "alert", "Halt", "Ready"},
		nil,
		regexp.QuoteMeta(updateCompPrefix +
			" state = 'Halt', flag = CASE WHEN flag = 'Locked' THEN 'Locked' ELSE 'Alert' END WHERE (id = $1) AND ((state != 'Halt') OR (flag != 'Alert')) AND (state = 'Halt' OR state = 'Ready');"),
		[]driver.Value{"x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}, {
		[]string{"x0c0s27b0n0s0"},
		"Ready",
		"",
		false,
		new(PartInfo),
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,              //
		"",               // Skipped, single arg
		[]driver.Value{}, //
		ErrHMSDSArgBadID, //
		"",               // Not used due to error preparing update.
		[]driver.Value{}, //
		[]string{},
	}, {
		[]string{"x0c0s25b0n0", "x0c0s27b0n0s0"}, // second value is bad.
		"Ready",
		"",
		false,
		new(PartInfo),
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s25b0n0"},
			[]driver.Value{"x0c0s27b0n0"},
		},
		ErrHMSDSArgBadID, //
		"",               // query prepare fails, one of input is bad.
		[]driver.Value{}, //
		nil,              //
		"",               // Don't get here, earlier error
		[]driver.Value{}, //
		[]string{},
	}, {
		[]string{"x0000c0s27b0n0"},
		"ready",
		"",
		false,
		new(PartInfo),
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		"", // not called
		[]driver.Value{},
		nil,
		regexp.QuoteMeta(updateCompPrefix +
			" state = 'Ready', flag = CASE WHEN flag = 'Locked' THEN 'Locked' ELSE 'OK' END WHERE (id = $1) AND ((state != 'Ready') OR (flag != 'OK')) AND (state = 'Ready' OR state = 'On' OR state = 'Off' OR state = 'Standby' OR state = 'Halt');"),
		[]driver.Value{"x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}, {
		[]string{"x0c0s25b0n0", "x00c0s27b0n0"},
		"on",
		"",
		false,
		new(PartInfo),
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s25b0n0"},
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND ((state != $3) OR (flag != $4)) AND (state = $5 OR state = $6 OR state = $7 OR state = $8);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "On", "OK", "On", "Off", "Standby", "Halt"},
		nil,
		regexp.QuoteMeta(updateCompPrefix + " state = 'On', flag = CASE WHEN flag = 'Locked' THEN 'Locked' ELSE 'OK' END WHERE (id = $1 OR id = $2);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0"},
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.expectedUpdatePrepare == "" && test.dbUpdateError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			if test.expectedQueryPrepare != "" {
				mockPG.ExpectPrepare(
					test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			}
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			if test.expectedQueryPrepare != "" {
				mockPG.ExpectPrepare(
					test.expectedQueryPrepare).ExpectQuery().WithArgs(
					test.expectedQueryArgs...).WillReturnRows(rows)
			}
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			if test.expectedQueryPrepare != "" {
				mockPG.ExpectPrepare(
					test.expectedQueryPrepare).ExpectQuery().WithArgs(
					test.expectedQueryArgs...).WillReturnRows(rows)
			}
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WithArgs(
				test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		affectedIDs, err := dPG.UpdateCompStates(test.ids, test.state, test.flag, test.force, test.pi)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s",
					i, err)
			}
			if !compareIDs(test.expectedAffectedIDs, affectedIDs) {
				t.Errorf("Test %v Failed: Expected affectedIDs '%v'; Recieved affectedIDs '%v'",
					i, test.expectedAffectedIDs, affectedIDs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgBulkUpdateCompFlagOnly(t *testing.T) {
	getCompIDPrefix := "SELECT id FROM components "
	tests := []struct {
		ids                   []string
		flag                  string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedAffectedIDs   []string
	}{{
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
		"alert",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s25b0n0"},
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (flag != $3);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "Alert"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompFlagOnlyPrefix) +
			"WHERE (id = $2 OR id = $3);"),
		[]driver.Value{"Alert", "x0c0s25b0n0", "x0c0s27b0n0"},
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
	}, {
		[]string{"x0c0s27b0n0s0"},
		"ok",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		ErrHMSDSArgBadID, //
		"",               // Error during query building due to bad xname
		[]driver.Value{}, //
		nil,
		regexp.QuoteMeta(updateCompFlagOnlyPrefix + "WHERE (id = $1);"),
		[]driver.Value{"OK", "x0c0s27b0n0s0"},
		[]string{},
	}, {
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
		"ok",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (flag != $3);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "OK"},
		sql.ErrNoRows,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompFlagOnlyPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{"OK", "x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WithArgs(
				test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		affectedIDs, err := dPG.BulkUpdateCompFlagOnly(test.ids, test.flag)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if !compareIDs(test.expectedAffectedIDs, affectedIDs) {
				t.Errorf("Test %v Failed: Expected affectedIDs '%v'; Recieved affectedIDs '%v'",
					i, test.expectedAffectedIDs, affectedIDs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgBulkUpdateCompEnabled(t *testing.T) {
	tests := []struct {
		ids                   []string
		enabled               bool
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedAffectedIDs   []string
	}{{
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
		true,
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (enabled != $3);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "1"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompEnabledPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{true, "x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}, {
		[]string{"x0c0s27b0n0s0"},
		true,
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s25b0n0"},
		},
		ErrHMSDSArgBadID, //
		"",               // Error during query building due to bad xname
		[]driver.Value{}, //
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompEnabledPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{true, "x0c0s27b0n0s0"},
		[]string{},
	}, {
		[]string{"x0c0s27b0n0", "x0c0s28b0n0"},
		false,
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (enabled != $3);"),
		[]driver.Value{"x0c0s27b0n0", "x0c0s28b0n0", "0"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompEnabledPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{false, "x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to
		// expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WithArgs(
				test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		affectedIDs, err := dPG.BulkUpdateCompEnabled(test.ids, test.enabled)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s",
					i, err)
			}
			if !compareIDs(test.expectedAffectedIDs, affectedIDs) {
				t.Errorf("Test %v Failed: Expected affectedIDs '%v'; Recieved affectedIDs '%v'",
					i, test.expectedAffectedIDs, affectedIDs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgBulkUpdateCompSwStatus(t *testing.T) {
	tests := []struct {
		ids                   []string
		swstatus              string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedAffectedIDs   []string
	}{{
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
		"OtherStatus", // arbitrary, we don't check values for this field.
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s25b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (admin != $3);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "OtherStatus"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompSwStatusPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{"OtherStatus", "x0c0s25b0n0"},
		[]string{"x0c0s25b0n0"},
	}, {
		[]string{"x0c0s27b0n0s0"},
		"AdminStatus",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		ErrHMSDSArgBadID, //
		"",               // Error during query building due to bad xname
		[]driver.Value{}, //
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompSwStatusPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{"AdminStatus", "x0c0s27b0n0s0"},
		[]string{},
	}, {
		[]string{"x0c0s27b0n0", "x0c0s28b0n0"},
		"OtherStatus",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (admin != $3);"),
		[]driver.Value{"x0c0s27b0n0", "x0c0s28b0n0", "OtherStatus"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompSwStatusPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{"OtherStatus", "x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to
		// expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WithArgs(
				test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		affectedIDs, err := dPG.BulkUpdateCompSwStatus(test.ids, test.swstatus)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if !compareIDs(test.expectedAffectedIDs, affectedIDs) {
				t.Errorf("Test %v Failed: Expected affectedIDs '%v'; Recieved affectedIDs '%v'",
					i, test.expectedAffectedIDs, affectedIDs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgBulkUpdateCompRole(t *testing.T) {
	tests := []struct {
		ids                   []string
		role                  string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedAffectedIDs   []string
	}{{
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
		"compute",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix + " WHERE (id = $1 OR id = $2) AND (role != $3);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "Compute"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompRolePrefix) +
			"WHERE (id = $3);"),
		[]driver.Value{"Compute", "", "x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}, {
		[]string{"x0c0s27b0n0s0"},
		"service",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		ErrHMSDSArgBadID, //
		"",               // Error during query building due to bad xname
		[]driver.Value{}, //
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompRolePrefix) +
			"WHERE (id = $3);"),
		[]driver.Value{"Service", "", "x0c0s27b0n0s0"},
		[]string{},
	}, {
		[]string{"x0c0s27b0n0"},
		"Service",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1) AND (role != $2);"),
		[]driver.Value{"x0c0s27b0n0", "Service"},
		sql.ErrNoRows,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompRolePrefix) +
			"WHERE (id = $3);"),
		[]driver.Value{"Service", "", "x0c0s27b0n0"},
		[]string{},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WithArgs(
				test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		affectedIDs, err := dPG.BulkUpdateCompRole(test.ids, test.role, "")
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s",
					i, err)
			}
			if !compareIDs(test.expectedAffectedIDs, affectedIDs) {
				t.Errorf("Test %v Failed: Expected affectedIDs '%v'; Recieved affectedIDs '%v'",
					i, test.expectedAffectedIDs, affectedIDs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgBulkUpdateCompClass(t *testing.T) {
	tests := []struct {
		ids                   []string
		class                 string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedAffectedIDs   []string
	}{{
		[]string{"x0c0s25b0n0", "x0c0s27b0n0"},
		"River",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s25b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (class != $3);"),
		[]driver.Value{"x0c0s25b0n0", "x0c0s27b0n0", "River"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompClassPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{"River", "x0c0s25b0n0"},
		[]string{"x0c0s25b0n0"},
	}, {
		[]string{"x0c0s27b0n0s0"},
		"Mountain",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		ErrHMSDSArgBadID, //
		"",               // Error during query building due to bad xname
		[]driver.Value{}, //
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompClassPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{"Mountain", "x0c0s27b0n0s0"},
		[]string{},
	}, {
		[]string{"x0c0s27b0n0", "x0c0s28b0n0"},
		"River",
		[]string{"id"},
		[][]driver.Value{
			[]driver.Value{"x0c0s27b0n0"},
		},
		nil,
		regexp.QuoteMeta(getCompIDPrefix +
			" WHERE (id = $1 OR id = $2) AND (class != $3);"),
		[]driver.Value{"x0c0s27b0n0", "x0c0s28b0n0", "River"},
		nil,
		regexp.QuoteMeta(ToPGQueryArgs(updateCompClassPrefix) +
			"WHERE (id = $2);"),
		[]driver.Value{"River", "x0c0s27b0n0"},
		[]string{"x0c0s27b0n0"},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to
		// expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(
				test.expectedQueryPrepare).ExpectQuery().WithArgs(
				test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(
				test.expectedUpdatePrepare).ExpectExec().WithArgs(
				test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		affectedIDs, err := dPG.BulkUpdateCompClass(test.ids, test.class)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if !compareIDs(test.expectedAffectedIDs, affectedIDs) {
				t.Errorf("Test %v Failed: Expected affectedIDs '%v'; Recieved affectedIDs '%v'",
					i, test.expectedAffectedIDs, affectedIDs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgBulkUpdateCompNID(t *testing.T) {

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	update, _, _ := sqq.Update(compTable+" "+compTableJoinAlias).
		Set(compNIDCol, sq.Expr(compTableSubAlias+"."+compNIDCol)).
		Suffix("FROM (VALUES (?,?::BIGINT),(?,?::BIGINT)) AS "+compTableSubAlias+
			"("+compIdCol+", "+compNIDCol+
			") WHERE "+compTableJoinAlias+"."+compIdCol+" = "+compTableSubAlias+"."+compIdCol,
			"x0c0s25b0n0", 800, "x0c0s27b0n0", 864).ToSql()
	tests := []struct {
		comps                 *[]base.Component
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
	}{{
		comps: &[]base.Component{
			base.Component{"x0c0s25b0n0", "", "", "", nil, "", "", "", "800", "", "", "", "", false, false},
			base.Component{"x0c0s27b0n0", "", "", "", nil, "", "", "", "864", "", "", "", "", false, false},
		},
		dbUpdateError:         nil,
		expectedUpdatePrepare: regexp.QuoteMeta(update),
		expectedUpdateArgs:    []driver.Value{"x0c0s25b0n0", 800, "x0c0s27b0n0", 864},
	}, {
		comps: &[]base.Component{
			base.Component{"x0c0s25b0n0", "", "", "", nil, "", "", "", "-1", "", "", "", "", false, false},
			base.Component{"x0c0s27b0n0", "", "", "", nil, "", "", "", "864", "", "", "", "", false, false},
		},
		dbUpdateError:         nil,
		expectedUpdatePrepare: regexp.QuoteMeta(update),
		expectedUpdateArgs:    []driver.Value{"x0c0s25b0n0", -1, "x0c0s27b0n0", 864},
	}, {
		comps: &[]base.Component{
			base.Component{"x0c0s25b0n0", "", "", "", nil, "", "", "", "800", "", "", "", "", false, false},
			base.Component{"x0c0s27b0n0", "", "", "", nil, "", "", "", "864", "", "", "", "", false, false},
		},
		dbUpdateError:         sql.ErrNoRows,
		expectedUpdatePrepare: regexp.QuoteMeta(update),
		expectedUpdateArgs:    []driver.Value{"x0c0s25b0n0", 800, "x0c0s27b0n0", 864},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to
		// expect required DB actions

		mockPG.ExpectBegin()
		if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, int64(len(*test.comps))))
			mockPG.ExpectCommit()
		}

		err := dPG.BulkUpdateCompNID(test.comps)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Hardware Inventory Tests
///////////////////////////////////////////////////////////////////////////////

func TestPgGetHWInvByLocQueryFilter(t *testing.T) {
	columns := addAliasToCols(hwInvAlias, hwInvCols, hwInvCols)
	preQueryCols := []string{hwInvLocAlias + "." + hwInvLocIdCol}
	node1 := stest.NodeHWInvByLoc1
	node1LocInfo, _ := json.Marshal(node1.HMSNodeLocationInfo)
	node1FruInfo, _ := json.Marshal(node1.PopulatedFRU.HMSNodeFRUInfo)
	proc1 := stest.ProcHWInvByLoc1
	proc1LocInfo, _ := json.Marshal(proc1.HMSProcessorLocationInfo)
	proc1FruInfo, _ := json.Marshal(proc1.PopulatedFRU.HMSProcessorFRUInfo)

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query1, _, _ := sqq.Select(columns...).
		From(hwInvTable + " " + hwInvAlias).ToSql()

	preQuery2, preQuery2Args, _ := sq.Select(preQueryCols...).
		From(hwInvLocTable + " " + hwInvLocAlias).
		Where(sq.Expr("("+hwInvLocAlias+"."+hwInvLocIdCol+" SIMILAR TO ?)", node1.ID+"([[:alpha:]][[:alnum:]]*)?")).
		Where(sq.Eq{hwInvLocAlias + "." + hwInvLocTypeCol: []string{base.Processor.String()}}).ToSql()
	preQuery2Args = append(preQuery2Args, ")([[:alpha:]][[:alnum:]]*)?")
	query2, _, _ := sqq.Select(columns...).
		From(hwInvTable+" "+hwInvAlias).
		Where(hwInvAlias+"."+hwInvIdCol+" SIMILAR TO (SELECT id_str FROM id_sel)").
		Prefix("WITH id_sel AS (SELECT ('('||array_to_string(array("+preQuery2+"),'|')||?) AS id_str)", preQuery2Args...).ToSql()

	filterQuery3, filterQuery3Args, _ := sq.Eq{hwInvAlias + "." + hwInvIdCol: []string{"x0"}}.ToSql()
	filterArgs := make([]interface{}, 0, 1)
	filterArgs = append(filterArgs, "x0c0([[:alpha:]][[:alnum:]]*)?")
	filterQuery3Args = append(filterArgs, filterQuery3Args...)
	query3, _, _ := sqq.Select(columns...).
		From(hwInvTable + " " + hwInvAlias).
		Where(sq.Expr("(("+hwInvAlias+"."+hwInvIdCol+" SIMILAR TO ?)) OR "+filterQuery3, filterQuery3Args...)).ToSql()

	query4, _, _ := sqq.Select(columns...).
		From(hwInvTable + " " + hwInvAlias).
		Where(sq.Eq{hwInvAlias + "." + hwInvTypeCol: []string{base.Processor.String()}}).ToSql()

	tests := []struct {
		f_opts          []HWInvLocFiltFunc
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedHwLocs  []*sm.HWInvByLoc
		expectedErr     error
	}{{
		f_opts: []HWInvLocFiltFunc{},
		dbRows: [][]driver.Value{
			[]driver.Value{node1.ID, node1.Type, node1.Ordinal, node1.Status, node1LocInfo, node1.PopulatedFRU.FRUID, node1.PopulatedFRU.Type, node1.PopulatedFRU.Subtype, node1FruInfo},
			[]driver.Value{proc1.ID, proc1.Type, proc1.Ordinal, proc1.Status, proc1LocInfo, proc1.PopulatedFRU.FRUID, proc1.PopulatedFRU.Type, proc1.PopulatedFRU.Subtype, proc1FruInfo},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query1),
		expectedArgs:    []driver.Value{},
		expectedHwLocs:  []*sm.HWInvByLoc{&node1, &proc1},
		expectedErr:     nil,
	}, {
		f_opts: []HWInvLocFiltFunc{HWInvLoc_ID(node1.ID), HWInvLoc_Type(base.Processor.String()), HWInvLoc_Child},
		dbRows: [][]driver.Value{
			[]driver.Value{proc1.ID, proc1.Type, proc1.Ordinal, proc1.Status, proc1LocInfo, proc1.PopulatedFRU.FRUID, proc1.PopulatedFRU.Type, proc1.PopulatedFRU.Subtype, proc1FruInfo},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query2),
		expectedArgs:    []driver.Value{node1.ID + "([[:alpha:]][[:alnum:]]*)?", base.Processor.String(), ")([[:alpha:]][[:alnum:]]*)?"},
		expectedHwLocs:  []*sm.HWInvByLoc{&proc1},
		expectedErr:     nil,
	}, {
		f_opts: []HWInvLocFiltFunc{HWInvLoc_ID("x0c0"), HWInvLoc_Parent, HWInvLoc_Child},
		dbRows: [][]driver.Value{
			[]driver.Value{node1.ID, node1.Type, node1.Ordinal, node1.Status, node1LocInfo, node1.PopulatedFRU.FRUID, node1.PopulatedFRU.Type, node1.PopulatedFRU.Subtype, node1FruInfo},
			[]driver.Value{proc1.ID, proc1.Type, proc1.Ordinal, proc1.Status, proc1LocInfo, proc1.PopulatedFRU.FRUID, proc1.PopulatedFRU.Type, proc1.PopulatedFRU.Subtype, proc1FruInfo},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query3),
		expectedArgs:    []driver.Value{"x0c0([[:alpha:]][[:alnum:]]*)?", "x0"},
		expectedHwLocs:  []*sm.HWInvByLoc{&node1, &proc1},
		expectedErr:     nil,
	}, {
		f_opts:          []HWInvLocFiltFunc{HWInvLoc_Type("foo")},
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedHwLocs:  nil,
		expectedErr:     ErrHMSDSArgBadRedfishType,
	}, {
		f_opts:          []HWInvLocFiltFunc{HWInvLoc_Type(base.Processor.String())},
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(query4),
		expectedArgs:    []driver.Value{base.Processor.String()},
		expectedHwLocs:  nil,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
		} else if test.expectedErr == nil {
			if len(test.expectedArgs) > 0 {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
		}

		hwlocs, err := dPG.GetHWInvByLocQueryFilter(test.f_opts...)
		if test.expectedErr == nil {
			// ensure all expectations have been met
			if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
				t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
			}
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedHwLocs, hwlocs) {
				t.Errorf("Test %v Failed: Expected HWLocs '%v'; Recieved HWLocs '%v'", i, test.expectedHwLocs, hwlocs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetHWInvByLocFilter(t *testing.T) {
	columns := addAliasToCols(hwInvAlias, hwInvCols, hwInvCols)

	node1 := stest.NodeHWInvByLoc1
	node1LocInfo, _ := json.Marshal(node1.HMSNodeLocationInfo)
	node1FruInfo, _ := json.Marshal(node1.PopulatedFRU.HMSNodeFRUInfo)

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query1, _, _ := sqq.Select(columns...).
		From(hwInvTable + " " + hwInvAlias).
		Where(sq.Eq{hwInvAlias + "." + hwInvFruIdCol: []string{node1.PopulatedFRU.FRUID}}).ToSql()

	query2, _, _ := sqq.Select(columns...).
		From(hwInvTable + " " + hwInvAlias).
		Where(sq.Eq{hwInvAlias + "." + hwInvIdCol: []string{node1.ID}}).
		Where(sq.Eq{hwInvAlias + "." + hwInvTypeCol: []string{base.Node.String()}}).
		Where(sq.Expr("("+hwInvAlias+"."+hwInvFruInfoCol+" ->> 'Manufacturer' ILIKE ?)", "%cray%")).
		Where(sq.Eq{hwInvAlias + "." + hwInvFruInfoCol + " ->> 'PartNumber'": []string{node1.PopulatedFRU.HMSNodeFRUInfo.PartNumber}}).
		Where(sq.Eq{hwInvAlias + "." + hwInvFruInfoCol + " ->> 'SerialNumber'": []string{node1.PopulatedFRU.HMSNodeFRUInfo.SerialNumber}}).
		Where(sq.Eq{hwInvAlias + "." + hwInvFruIdCol: []string{node1.PopulatedFRU.FRUID}}).ToSql()

	tests := []struct {
		f_opts          []HWInvLocFiltFunc
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedHwLocs  []*sm.HWInvByLoc
		expectedErr     error
	}{{
		f_opts: []HWInvLocFiltFunc{HWInvLoc_ID(node1.ID), HWInvLoc_Type(node1.Type), HWInvLoc_Manufacturers([]string{"cray"}), HWInvLoc_PartNumbers([]string{node1.PopulatedFRU.HMSNodeFRUInfo.PartNumber}), HWInvLoc_SerialNumbers([]string{node1.PopulatedFRU.HMSNodeFRUInfo.SerialNumber}), HWInvLoc_FruIDs([]string{node1.PopulatedFRU.FRUID})},
		dbRows: [][]driver.Value{
			[]driver.Value{node1.ID, node1.Type, node1.Ordinal, node1.Status, node1LocInfo, node1.PopulatedFRU.FRUID, node1.PopulatedFRU.Type, node1.PopulatedFRU.Subtype, node1FruInfo},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query2),
		expectedArgs:    []driver.Value{node1.ID, base.Node.String(), "%cray%", node1.PopulatedFRU.HMSNodeFRUInfo.PartNumber, node1.PopulatedFRU.HMSNodeFRUInfo.SerialNumber, node1.PopulatedFRU.FRUID},
		expectedHwLocs:  []*sm.HWInvByLoc{&node1},
		expectedErr:     nil,
	}, {
		f_opts:          []HWInvLocFiltFunc{HWInvLoc_Type("foo")},
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedHwLocs:  nil,
		expectedErr:     ErrHMSDSArgBadRedfishType,
	}, {
		f_opts:          []HWInvLocFiltFunc{HWInvLoc_FruIDs([]string{node1.PopulatedFRU.FRUID})},
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(query1),
		expectedArgs:    []driver.Value{node1.PopulatedFRU.FRUID},
		expectedHwLocs:  nil,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
		} else if test.expectedErr == nil {
			if len(test.expectedArgs) > 0 {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
		}

		hwlocs, err := dPG.GetHWInvByLocFilter(test.f_opts...)
		if test.expectedErr == nil {
			// ensure all expectations have been met
			if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
				t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
			}
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedHwLocs, hwlocs) {
				t.Errorf("Test %v Failed: Expected HWLocs '%v'; Recieved HWLocs '%v'", i, test.expectedHwLocs, hwlocs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetHWInvByFRUFilter(t *testing.T) {
	columns := addAliasToCols(hwInvFruAlias, hwInvFruTblCols, hwInvFruTblCols)

	node1 := stest.NodeHWInvByFRU1
	node1FruInfo, _ := json.Marshal(node1.HMSNodeFRUInfo)
	proc1 := stest.ProcHWInvByFRU1
	proc1FruInfo, _ := json.Marshal(proc1.HMSProcessorFRUInfo)

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query1, _, _ := sqq.Select(columns...).
		From(hwInvFruTable + " " + hwInvFruAlias).ToSql()

	query2, _, _ := sqq.Select(columns...).
		From(hwInvFruTable + " " + hwInvFruAlias).
		Where(sq.Eq{hwInvFruAlias + "." + hwInvFruTblTypeCol: []string{base.Node.String()}}).
		Where(sq.Expr("("+hwInvFruAlias+"."+hwInvFruTblInfoCol+" ->> 'Manufacturer' ILIKE ?)", "%cray%")).
		Where(sq.Eq{hwInvFruAlias + "." + hwInvFruTblInfoCol + " ->> 'PartNumber'": []string{node1.HMSNodeFRUInfo.PartNumber}}).
		Where(sq.Eq{hwInvFruAlias + "." + hwInvFruTblInfoCol + " ->> 'SerialNumber'": []string{node1.HMSNodeFRUInfo.SerialNumber}}).
		Where(sq.Eq{hwInvFruAlias + "." + hwInvFruTblIdCol: []string{node1.FRUID}}).ToSql()

	query3, _, _ := sqq.Select(columns...).
		From(hwInvFruTable + " " + hwInvFruAlias).
		Where(sq.Eq{hwInvFruAlias + "." + hwInvFruTblTypeCol: []string{base.Processor.String()}}).ToSql()

	tests := []struct {
		f_opts          []HWInvLocFiltFunc
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedHwFrus  []*sm.HWInvByFRU
		expectedErr     error
	}{{
		f_opts: []HWInvLocFiltFunc{},
		dbRows: [][]driver.Value{
			[]driver.Value{node1.FRUID, node1.Type, node1.Subtype, node1FruInfo},
			[]driver.Value{proc1.FRUID, proc1.Type, proc1.Subtype, proc1FruInfo},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query1),
		expectedArgs:    []driver.Value{},
		expectedHwFrus:  []*sm.HWInvByFRU{&node1, &proc1},
		expectedErr:     nil,
	}, {
		f_opts: []HWInvLocFiltFunc{HWInvLoc_Type(node1.Type), HWInvLoc_Manufacturers([]string{"cray"}), HWInvLoc_PartNumbers([]string{node1.HMSNodeFRUInfo.PartNumber}), HWInvLoc_SerialNumbers([]string{node1.HMSNodeFRUInfo.SerialNumber}), HWInvLoc_FruIDs([]string{node1.FRUID})},
		dbRows: [][]driver.Value{
			[]driver.Value{node1.FRUID, node1.Type, node1.Subtype, node1FruInfo},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query2),
		expectedArgs:    []driver.Value{base.Node.String(), "%cray%", node1.HMSNodeFRUInfo.PartNumber, node1.HMSNodeFRUInfo.SerialNumber, node1.FRUID},
		expectedHwFrus:  []*sm.HWInvByFRU{&node1},
		expectedErr:     nil,
	}, {
		f_opts:          []HWInvLocFiltFunc{HWInvLoc_Type("foo")},
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedHwFrus:  nil,
		expectedErr:     ErrHMSDSArgBadRedfishType,
	}, {
		f_opts:          []HWInvLocFiltFunc{HWInvLoc_Type(base.Processor.String())},
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(query3),
		expectedArgs:    []driver.Value{base.Processor.String()},
		expectedHwFrus:  nil,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
		} else if test.expectedErr == nil {
			if len(test.expectedArgs) > 0 {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
		}

		hwfrus, err := dPG.GetHWInvByFRUFilter(test.f_opts...)
		if test.expectedErr == nil {
			// ensure all expectations have been met
			if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
				t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
			}
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedHwFrus, hwfrus) {
				t.Errorf("Test %v Failed: Expected HWFrus '%v'; Recieved HWFrus '%v'", i, test.expectedHwFrus, hwfrus)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Hardware Inventory History Tests
///////////////////////////////////////////////////////////////////////////////

func TestPgGetHWInvHistFilter(t *testing.T) {
	columns := addAliasToCols(hwInvHistAlias, hwInvHistCols, hwInvHistCols)

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
	timeStartArg, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	timeEndArg, _ := time.Parse(time.RFC3339, "2020-02-01T00:00:00Z")

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query1, _, _ := sqq.Select(columns...).
		From(hwInvHistTable + " " + hwInvHistAlias).
		OrderBy("timestamp ASC").ToSql()

	query2, _, _ := sqq.Select(columns...).
		From(hwInvHistTable + " " + hwInvHistAlias).
		Where(sq.Eq{hwInvHistAlias + "." + hwInvHistIdCol: []string{testHWInvHist1.ID}}).
		Where(sq.Gt{hwInvHistAlias + "." + hwInvHistTimestampCol: timeStartArg}).
		Where(sq.Lt{hwInvHistAlias + "." + hwInvHistTimestampCol: timeEndArg}).
		OrderBy("timestamp ASC").ToSql()

	query3, _, _ := sqq.Select(columns...).
		From(hwInvHistTable + " " + hwInvHistAlias).
		Where(sq.Eq{hwInvHistAlias + "." + hwInvHistEventTypeCol: []string{sm.HWInvHistEventTypeAdded}}).
		OrderBy("timestamp ASC").ToSql()

	tests := []struct {
		f_opts          []HWInvHistFiltFunc
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedHwHists []*sm.HWInvHist
		expectedErr     error
	}{{
		f_opts: []HWInvHistFiltFunc{},
		dbRows: [][]driver.Value{
			[]driver.Value{testHWInvHist1.ID, testHWInvHist1.FruId, testHWInvHist1.EventType, testHWInvHist1.Timestamp},
			[]driver.Value{testHWInvHist2.ID, testHWInvHist2.FruId, testHWInvHist2.EventType, testHWInvHist2.Timestamp},
			[]driver.Value{testHWInvHist3.ID, testHWInvHist3.FruId, testHWInvHist3.EventType, testHWInvHist3.Timestamp},
			[]driver.Value{testHWInvHist4.ID, testHWInvHist4.FruId, testHWInvHist4.EventType, testHWInvHist4.Timestamp},
			[]driver.Value{testHWInvHist5.ID, testHWInvHist5.FruId, testHWInvHist5.EventType, testHWInvHist5.Timestamp},
			[]driver.Value{testHWInvHist6.ID, testHWInvHist6.FruId, testHWInvHist6.EventType, testHWInvHist6.Timestamp},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query1),
		expectedArgs:    []driver.Value{},
		expectedHwHists: []*sm.HWInvHist{
			&testHWInvHist1,
			&testHWInvHist2,
			&testHWInvHist3,
			&testHWInvHist4,
			&testHWInvHist5,
			&testHWInvHist6,
		},
		expectedErr: nil,
	}, {
		f_opts: []HWInvHistFiltFunc{
			HWInvHist_ID(testHWInvHist1.ID),
			HWInvHist_StartTime("2020-01-01T00:00:00Z"),
			HWInvHist_EndTime("2020-02-01T00:00:00Z"),
		},
		dbRows: [][]driver.Value{
			[]driver.Value{testHWInvHist1.ID, testHWInvHist1.FruId, testHWInvHist1.EventType, testHWInvHist1.Timestamp},
			[]driver.Value{testHWInvHist3.ID, testHWInvHist3.FruId, testHWInvHist3.EventType, testHWInvHist3.Timestamp},
			[]driver.Value{testHWInvHist5.ID, testHWInvHist5.FruId, testHWInvHist5.EventType, testHWInvHist5.Timestamp},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query2),
		expectedArgs:    []driver.Value{testHWInvHist1.ID, timeStartArg, timeEndArg},
		expectedHwHists: []*sm.HWInvHist{
			&testHWInvHist1,
			&testHWInvHist3,
			&testHWInvHist5,
		},
		expectedErr: nil,
	}, {
		f_opts:          []HWInvHistFiltFunc{HWInvHist_EventTypes([]string{"foo"})},
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedHwHists: nil,
		expectedErr:     ErrHMSDSArgBadHWInvHistEventType,
	}, {
		f_opts:          []HWInvHistFiltFunc{HWInvHist_EventTypes([]string{sm.HWInvHistEventTypeAdded})},
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(query3),
		expectedArgs:    []driver.Value{sm.HWInvHistEventTypeAdded},
		expectedHwHists: nil,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr == nil {
			if len(test.expectedArgs) > 0 {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
			mockPG.ExpectCommit()
		}

		hwhists, err := dPG.GetHWInvHistFilter(test.f_opts...)
		if test.expectedErr == nil {
			// ensure all expectations have been met
			if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
				t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
			}
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedHwHists, hwhists) {
				t.Errorf("Test %v Failed: Expected HWHists '%v'; Recieved HWHists '%v'", i, test.expectedHwHists, hwhists)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestInsertHWInvHists(t *testing.T) {
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
	testHWInvHistBad := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 12:00:00",
		EventType: "foo",
	}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	insert1, _, _ := sqq.Insert(hwInvHistTable).
		Columns(hwInvHistColsNoTS...).
		Values("x0c0s0b0n0", "MFR-PARTNUMBER-SERIALNUMBER_1", "Scanned").
		Values("x0c0s0b0n0", "MFR-PARTNUMBER-SERIALNUMBER_1", "Scanned").ToSql()

	tests := []struct {
		hhs             []*sm.HWInvHist
		expectedPrepare string
		expectedArgs    []driver.Value
		dbError         error
	}{{
		hhs:             []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist2},
		expectedPrepare: regexp.QuoteMeta(insert1),
		expectedArgs: []driver.Value{
			testHWInvHist1.ID, testHWInvHist1.FruId, testHWInvHist1.EventType,
			testHWInvHist2.ID, testHWInvHist2.FruId, testHWInvHist2.EventType,
		},
		dbError: nil,
	}, {
		hhs:             []*sm.HWInvHist{&testHWInvHistBad},
		expectedPrepare: "",
		expectedArgs:    []driver.Value{},
		dbError:         ErrHMSDSArgBadHWInvHistEventType,
	}, {
		hhs:             []*sm.HWInvHist{&testHWInvHist1, &testHWInvHist2},
		expectedPrepare: regexp.QuoteMeta(insert1),
		expectedArgs: []driver.Value{
			testHWInvHist1.ID, testHWInvHist1.FruId, testHWInvHist1.EventType,
			testHWInvHist2.ID, testHWInvHist2.FruId, testHWInvHist2.EventType,
		},
		dbError: sql.ErrNoRows,
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.expectedPrepare == "" && test.dbError != nil {
			mockPG.ExpectRollback()
		} else if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		err := dPG.InsertHWInvHists(test.hhs)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.dbError != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestDeleteHWInvHistByLocID(t *testing.T) {
	testHWInvHist1 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	delete1, _, _ := sqq.Delete(hwInvHistTable).
		Where(sq.Eq{hwInvHistIdCol: "x0c0s0b0n0"}).ToSql()

	tests := []struct {
		id              string
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedError   error
	}{{
		id:              testHWInvHist1.ID,
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedArgs:    []driver.Value{testHWInvHist1.ID},
		expectedError:   nil,
	}, {
		id:              testHWInvHist1.ID,
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedArgs:    []driver.Value{testHWInvHist1.ID},
		expectedError:   sql.ErrNoRows,
	}}

	for i, test := range tests {
		ResetMockDB()
		if test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.expectedError)
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
		}

		_, err := dPG.DeleteHWInvHistByLocID(test.id)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.expectedError == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.expectedError != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error (%s).", i, test.expectedError)
		}
	}
}

func TestDeleteHWInvHistByFRUID(t *testing.T) {
	testHWInvHist1 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	delete1, _, _ := sqq.Delete(hwInvHistTable).
		Where(sq.Eq{hwInvHistFruIdCol: "MFR-PARTNUMBER-SERIALNUMBER_1"}).ToSql()

	tests := []struct {
		fruid           string
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedError   error
	}{{
		fruid:           testHWInvHist1.FruId,
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedArgs:    []driver.Value{testHWInvHist1.FruId},
		expectedError:   nil,
	}, {
		fruid:           testHWInvHist1.FruId,
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedArgs:    []driver.Value{testHWInvHist1.FruId},
		expectedError:   sql.ErrNoRows,
	}}

	for i, test := range tests {
		ResetMockDB()
		if test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.expectedError)
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
		}

		_, err := dPG.DeleteHWInvHistByFRUID(test.fruid)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.expectedError == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.expectedError != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error (%s).", i, test.expectedError)
		}
	}
}

func TestDeleteHWInvHistAll(t *testing.T) {

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	delete1, _, _ := sqq.Delete(hwInvHistTable).ToSql()

	tests := []struct {
		expectedPrepare string
		expectedError   error
	}{{
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedError:   nil,
	}, {
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedError:   sql.ErrNoRows,
	}}

	for i, test := range tests {
		ResetMockDB()
		if test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.expectedError)
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		}

		_, err := dPG.DeleteHWInvHistAll()
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.expectedError == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.expectedError != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error (%s).", i, test.expectedError)
		}
	}
}

func TestDeleteHWInvHistFilter(t *testing.T) {
	testHWInvHist1 := sm.HWInvHist{
		ID:        "x5c4s3b2n1p0",
		FruId:     "MFR-PARTNUMBER-SERIALNUMBER_1",
		Timestamp: "2020-01-21 11:36:00",
		EventType: "Added",
	}
	timeStartArg, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	timeEndArg, _ := time.Parse(time.RFC3339, "2020-02-01T00:00:00Z")

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	delete1, _, _ := sqq.Delete(hwInvHistTable).ToSql()

	delete2, _, _ := sqq.Delete(hwInvHistTable).
		Where(sq.Eq{hwInvHistIdCol: []string{testHWInvHist1.ID}}).
		Where(sq.Gt{hwInvHistTimestampCol: timeStartArg}).
		Where(sq.Lt{hwInvHistTimestampCol: timeEndArg}).ToSql()

	delete3, _, _ := sqq.Delete(hwInvHistTable).
		Where(sq.Eq{hwInvHistEventTypeCol: []string{sm.HWInvHistEventTypeAdded}}).ToSql()

	tests := []struct {
		f_opts          []HWInvHistFiltFunc
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedErr     error
	}{{
		f_opts:          []HWInvHistFiltFunc{},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedArgs:    []driver.Value{},
		expectedErr:     nil,
	}, {
		f_opts: []HWInvHistFiltFunc{
			HWInvHist_ID(testHWInvHist1.ID),
			HWInvHist_StartTime("2020-01-01T00:00:00Z"),
			HWInvHist_EndTime("2020-02-01T00:00:00Z"),
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(delete2),
		expectedArgs:    []driver.Value{testHWInvHist1.ID, timeStartArg, timeEndArg},
		expectedErr:     nil,
	}, {
		f_opts:          []HWInvHistFiltFunc{HWInvHist_EventTypes([]string{"foo"})},
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedErr:     ErrHMSDSArgBadHWInvHistEventType,
	}, {
		f_opts:          []HWInvHistFiltFunc{HWInvHist_EventTypes([]string{sm.HWInvHistEventTypeAdded})},
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(delete3),
		expectedArgs:    []driver.Value{sm.HWInvHistEventTypeAdded},
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
		} else if test.expectedErr == nil {
			if len(test.expectedArgs) > 0 {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
				//mockPG.ExpectExec(test.expectedPrepare).WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
				//mockPG.ExpectExec(test.expectedPrepare).WillReturnResult(sqlmock.NewResult(0, 1))
			}
		}

		_, err := dPG.DeleteHWInvHistFilter(test.f_opts...)
		if test.expectedErr == nil {
			if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
				t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
			}
		}
		if test.dbError == nil && test.expectedErr == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.dbError != nil && test.expectedErr != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error (%s).", i, test.expectedErr)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Redfish Endpoints Tests
///////////////////////////////////////////////////////////////////////////////

func TestPgPatchRFEndpointNoDiscInfo(t *testing.T) {
	trueBool := true
	//falseBool := false
	columns := rfEPsAllCols
	emptyDiscInfo := rf.DiscoveryInfo{}
	badHost := "x0c0s0b0"
	ipAddr := "10.254.2.13"

	tests := []struct {
		id                  string
		epp                 sm.RedfishEndpointPatch
		dbErrorGet1         error
		dbRowsGet1          [][]driver.Value
		expectedPrepareGet1 string
		expectedArgsGet1    []driver.Value
		dbError             error
		expectedPrepare     string
		expectedArgs        []driver.Value
		dbErrorGet2         error
		dbRowsGet2          [][]driver.Value
		expectedPrepareGet2 string
		expectedArgsGet2    []driver.Value
		expectedResp        *sm.RedfishEndpoint
		expectedIDs         []string
		expectedErr         error
		expectErr           bool
	}{{ // Test 0 Success
		id: "x0c0s1b1",
		epp: sm.RedfishEndpointPatch{
			Enabled:        &trueBool,
			RediscOnUpdate: &trueBool,
		},
		dbErrorGet1: nil,
		dbRowsGet1: [][]driver.Value{
			[]driver.Value{"x0c0s1b1", "NodeBMC", "", "10.254.2.12", "", "10.254.2.12", false, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.12", false, "", json.RawMessage(`{}`)},
		},
		expectedPrepareGet1: regexp.QuoteMeta(getRFEndpointPrefix + " WHERE (id = $1);"),
		expectedArgsGet1:    []driver.Value{"x0c0s1b1"},
		dbError:             nil,
		expectedPrepare:     regexp.QuoteMeta(updatePgRFEndpointNoDiscInfoQuery),
		expectedArgs:        []driver.Value{"NodeBMC", "", "10.254.2.12", "", "10.254.2.12", true, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.12", true, "", "x0c0s1b1"},
		dbErrorGet2:         nil,
		dbRowsGet2: [][]driver.Value{
			[]driver.Value{"x0c0s1b1", "NodeBMC", "", "10.254.2.12", "", "10.254.2.12", true, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.12", true, "", json.RawMessage(`{}`)},
		},
		expectedPrepareGet2: regexp.QuoteMeta(getRFEndpointByIDQuery),
		expectedArgsGet2:    []driver.Value{"x0c0s1b1"},
		expectedResp: &sm.RedfishEndpoint{
			RedfishEPDescription: rf.RedfishEPDescription{
				ID:             "x0c0s1b1",
				Type:           "NodeBMC",
				Name:           "",
				Hostname:       "10.254.2.12",
				Domain:         "",
				FQDN:           "10.254.2.12",
				Enabled:        true,
				UUID:           "da4faffe-6491-4f3f-ab54-3bf8fce57531",
				User:           "root",
				Password:       "",
				UseSSDP:        false,
				MACRequired:    false,
				MACAddr:        "a4bf012e85b5",
				IPAddr:         "10.254.2.12",
				RediscOnUpdate: true,
				TemplateID:     "",
				DiscInfo:       emptyDiscInfo,
			},
		},
		expectedIDs: []string{},
		expectedErr: nil,
		expectErr:   false,
	}, { // Test 1 No Update
		id:          "x0c0s1b1",
		epp:         sm.RedfishEndpointPatch{},
		dbErrorGet1: nil,
		dbRowsGet1: [][]driver.Value{
			[]driver.Value{"x0c0s1b1", "NodeBMC", "", "10.254.2.12", "", "10.254.2.12", false, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.12", false, "", json.RawMessage(`{}`)},
		},
		expectedPrepareGet1: regexp.QuoteMeta(getRFEndpointPrefix + " WHERE (id = $1);"),
		expectedArgsGet1:    []driver.Value{"x0c0s1b1"},
		dbError:             nil,
		expectedPrepare:     "",
		expectedArgs:        []driver.Value{},
		dbErrorGet2:         nil,
		dbRowsGet2:          [][]driver.Value{},
		expectedPrepareGet2: "",
		expectedArgsGet2:    []driver.Value{},
		expectedResp: &sm.RedfishEndpoint{
			RedfishEPDescription: rf.RedfishEPDescription{
				ID:             "x0c0s1b1",
				Type:           "NodeBMC",
				Name:           "",
				Hostname:       "10.254.2.12",
				Domain:         "",
				FQDN:           "10.254.2.12",
				Enabled:        false,
				UUID:           "da4faffe-6491-4f3f-ab54-3bf8fce57531",
				User:           "root",
				Password:       "",
				UseSSDP:        false,
				MACRequired:    false,
				MACAddr:        "a4bf012e85b5",
				IPAddr:         "10.254.2.12",
				RediscOnUpdate: false,
				TemplateID:     "",
				DiscInfo:       emptyDiscInfo,
			},
		},
		expectedIDs: []string{},
		expectedErr: nil,
		expectErr:   true,
	}, { // Test 2 Update would result in bad RFEndpoint
		id: "x0c0s1b1",
		epp: sm.RedfishEndpointPatch{
			Hostname: &badHost,
		},
		dbErrorGet1: nil,
		dbRowsGet1: [][]driver.Value{
			[]driver.Value{"x0c0s1b1", "NodeBMC", "", "10.254.2.12", "", "10.254.2.12", false, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.12", false, "", json.RawMessage(`{}`)},
		},
		expectedPrepareGet1: regexp.QuoteMeta(getRFEndpointPrefix + " WHERE (id = $1);"),
		expectedArgsGet1:    []driver.Value{"x0c0s1b1"},
		dbError:             nil,
		expectedPrepare:     "",
		expectedArgs:        []driver.Value{},
		dbErrorGet2:         nil,
		dbRowsGet2:          [][]driver.Value{},
		expectedPrepareGet2: "",
		expectedArgsGet2:    []driver.Value{},
		expectedResp:        nil,
		expectedErr:         fmt.Errorf("Some Error"),
		expectErr:           true,
	}, { // Test 3 No RFEndpoint
		id:                  "x0c0s1b1",
		epp:                 sm.RedfishEndpointPatch{},
		dbErrorGet1:         nil,
		dbRowsGet1:          [][]driver.Value{},
		expectedPrepareGet1: regexp.QuoteMeta(getRFEndpointPrefix + " WHERE (id = $1);"),
		expectedArgsGet1:    []driver.Value{"x0c0s1b1"},
		dbError:             nil,
		expectedPrepare:     "",
		expectedArgs:        []driver.Value{},
		dbErrorGet2:         nil,
		dbRowsGet2:          [][]driver.Value{},
		expectedPrepareGet2: "",
		expectedArgsGet2:    []driver.Value{},
		expectedResp:        nil,
		expectedIDs:         []string{},
		expectedErr:         nil,
		expectErr:           true,
	}, { // Test 4 IPAddress update
		id: "x0c0s1b1",
		epp: sm.RedfishEndpointPatch{
			IPAddr: &ipAddr,
		},
		dbErrorGet1: nil,
		dbRowsGet1: [][]driver.Value{
			[]driver.Value{"x0c0s1b1", "NodeBMC", "", "10.254.2.12", "", "10.254.2.12", false, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.12", false, "", json.RawMessage(`{}`)},
		},
		expectedPrepareGet1: regexp.QuoteMeta(getRFEndpointPrefix + " WHERE (id = $1);"),
		expectedArgsGet1:    []driver.Value{"x0c0s1b1"},
		dbError:             nil,
		expectedPrepare:     regexp.QuoteMeta(updatePgRFEndpointNoDiscInfoQuery),
		expectedArgs:        []driver.Value{"NodeBMC", "", "10.254.2.13", "", "10.254.2.13", false, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.13", false, "", "x0c0s1b1"},
		dbErrorGet2:         nil,
		dbRowsGet2: [][]driver.Value{
			[]driver.Value{"x0c0s1b1", "NodeBMC", "", "10.254.2.13", "", "10.254.2.13", false, "da4faffe-6491-4f3f-ab54-3bf8fce57531", "root", "", false, false, "a4bf012e85b5", "10.254.2.13", false, "", json.RawMessage(`{}`)},
		},
		expectedPrepareGet2: regexp.QuoteMeta(getRFEndpointByIDQuery),
		expectedArgsGet2:    []driver.Value{"x0c0s1b1"},
		expectedResp: &sm.RedfishEndpoint{
			RedfishEPDescription: rf.RedfishEPDescription{
				ID:             "x0c0s1b1",
				Type:           "NodeBMC",
				Name:           "",
				Hostname:       "10.254.2.13",
				Domain:         "",
				FQDN:           "10.254.2.13",
				Enabled:        false,
				UUID:           "da4faffe-6491-4f3f-ab54-3bf8fce57531",
				User:           "root",
				Password:       "",
				UseSSDP:        false,
				MACRequired:    false,
				MACAddr:        "a4bf012e85b5",
				IPAddr:         "10.254.2.13",
				RediscOnUpdate: false,
				TemplateID:     "",
				DiscInfo:       emptyDiscInfo,
			},
		},
		expectedIDs: []string{},
		expectedErr: nil,
		expectErr:   false,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows1 := sqlmock.NewRows(columns)
		for _, row := range test.dbRowsGet1 {
			rows1.AddRow(row...)
		}
		rows2 := sqlmock.NewRows(columns)
		for _, row := range test.dbRowsGet2 {
			rows2.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbErrorGet1 != nil {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepareGet1)).ExpectQuery().WillReturnError(test.dbErrorGet1)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepareGet1)).ExpectQuery().WithArgs(test.expectedArgsGet1...).WillReturnRows(rows1)
			if test.expectErr {
				mockPG.ExpectRollback()
			} else if test.dbError != nil {
				mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WillReturnError(test.dbError)
				mockPG.ExpectRollback()
			} else {
				mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
				if test.dbErrorGet2 != nil {
					mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepareGet2)).ExpectQuery().WillReturnError(test.dbErrorGet2)
					mockPG.ExpectRollback()
				} else {
					mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepareGet2)).ExpectQuery().WithArgs(test.expectedArgsGet2...).WillReturnRows(rows2)
					mockPG.ExpectCommit()
				}
			}
		}

		resp, affectedIDs, err := dPG.PatchRFEndpointNoDiscInfo(test.id, test.epp)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbErrorGet1 == nil && test.dbError == nil && test.dbErrorGet2 == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else {
				if !reflect.DeepEqual(test.expectedResp, resp) {
					t.Errorf("Test %v Failed: Expected RFEndpoint is %v - Received %v", i, test.expectedResp, resp)
				}
				if !reflect.DeepEqual(test.expectedIDs, affectedIDs) {
					t.Errorf("Test %v Failed: Expected affectedIDs is %v - Received %v", i, test.expectedIDs, affectedIDs)
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Service Endpoint Query Tests
///////////////////////////////////////////////////////////////////////////////

func TestPgGetServiceEndpointByID(t *testing.T) {
	columns := []string{"rf_endpoint_id", "redfish_type", "redfish_subtype", "uuid", "odata_id", "rf_endpoint_fqdn", "service_info"}

	tests := []struct {
		svc             string
		id              string
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedEP      *sm.ServiceEndpoint
		expectedErr     error
	}{{
		svc: "UpdateService",
		id:  "x0c0s21b0",
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", "UpdateService", "", "", "/redfish/v1/UpdateService", "10.4.0.8", json.RawMessage(`{}`)},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointPrefix + " WHERE (rf_endpoint_id = $1) AND (redfish_type = $2)"),
		expectedArgs:    []driver.Value{"x0c0s21b0", "UpdateService"},
		expectedEP:      &stest.TestServiceEndpointUpdate1,
		expectedErr:     nil,
	}, {
		svc:             "UpdateService",
		id:              "foo",
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedEP:      nil,
		expectedErr:     ErrHMSDSArgBadID,
	}, {
		svc:             "Update:Service",
		id:              "x0c0s21b0",
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedEP:      nil,
		expectedErr:     ErrHMSDSArgBadRedfishType,
	}, {
		svc:             "UpdateService",
		id:              "x0c0s21b0",
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointPrefix + " WHERE (rf_endpoint_id = $1) AND (redfish_type = $2)"),
		expectedArgs:    []driver.Value{"x0c0s21b0", "UpdateService"},
		expectedEP:      nil,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		sep, err := dPG.GetServiceEndpointByID(test.svc, test.id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedEP, sep) {
				t.Errorf("Test %v Failed: Expected service endpoint '%v'; Recieved service endpoint '%v'", i, test.expectedEP, sep)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetServiceEndpointsByAll(t *testing.T) {
	columns := []string{"rf_endpoint_id", "redfish_type", "redfish_subtype", "uuid", "odata_id", "rf_endpoint_fqdn", "service_info"}

	tests := []struct {
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedEPs     []*sm.ServiceEndpoint
	}{{
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", "UpdateService", "", "", "/redfish/v1/UpdateService", "10.4.0.8", json.RawMessage(`{}`)},
			[]driver.Value{"x0c0s22b0", "UpdateService", "", "", "/redfish/v1/UpdateService", "10.4.0.9", json.RawMessage(`{}`)},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointsAllQuery),
		expectedEPs:     stest.TestServiceEndpointArrayUpdates.ServiceEndpoints,
	}, {
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointsAllQuery),
		expectedEPs:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		seps, err := dPG.GetServiceEndpointsAll()
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedEPs, seps) {
				t.Errorf("Test %v Failed: Expected service endpoints '%v'; Recieved service endpoints '%v'", i, test.expectedEPs, seps)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetServiceEndpointsFilter(t *testing.T) {
	columns := []string{"rf_endpoint_id", "redfish_type", "redfish_subtype", "uuid", "odata_id", "rf_endpoint_fqdn", "service_info"}

	tests := []struct {
		f               *ServiceEPFilter
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedEPs     []*sm.ServiceEndpoint
		expectedErr     error
	}{{
		f: &ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"x0c0s21b0"},
		},
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", "UpdateService", "", "", "/redfish/v1/UpdateService", "10.4.0.8", json.RawMessage(`{}`)},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointPrefix + " WHERE (rf_endpoint_id = $1) AND (redfish_type = $2)"),
		expectedArgs:    []driver.Value{"x0c0s21b0", "UpdateService"},
		expectedEPs:     stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints,
		expectedErr:     nil,
	}, {
		f: &ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"x0c0s21b0", "x0c0s22b0"},
		},
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", "UpdateService", "", "", "/redfish/v1/UpdateService", "10.4.0.8", json.RawMessage(`{}`)},
			[]driver.Value{"x0c0s22b0", "UpdateService", "", "", "/redfish/v1/UpdateService", "10.4.0.9", json.RawMessage(`{}`)},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointPrefix + " WHERE (rf_endpoint_id = $1 OR rf_endpoint_id = $2) AND (redfish_type = $3)"),
		expectedArgs:    []driver.Value{"x0c0s21b0", "x0c0s22b0", "UpdateService"},
		expectedEPs:     stest.TestServiceEndpointArrayUpdates.ServiceEndpoints,
		expectedErr:     nil,
	}, {
		f: nil,
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", "UpdateService", "", "", "/redfish/v1/UpdateService", "10.4.0.8", json.RawMessage(`{}`)},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointsAllQuery),
		expectedArgs:    nil,
		expectedEPs:     stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints,
		expectedErr:     nil,
	}, {
		f: &ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"foo"},
		},
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedEPs:     nil,
		expectedErr:     ErrHMSDSArgBadID,
	}, {
		f: &ServiceEPFilter{
			Service:      []string{"Update:Service"},
			RfEndpointID: []string{"x0c0s21b0"},
		},
		dbRows:          nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedEPs:     nil,
		expectedErr:     ErrHMSDSArgBadRedfishType,
	}, {
		f: &ServiceEPFilter{
			Service:      []string{"UpdateService"},
			RfEndpointID: []string{"x0c0s21b0"},
		},
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(getServiceEndpointPrefix + " WHERE (rf_endpoint_id = $1) AND (redfish_type = $2)"),
		expectedArgs:    []driver.Value{"x0c0s21b0", "UpdateService"},
		expectedEPs:     nil,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			if test.expectedArgs != nil {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
			mockPG.ExpectCommit()
		}

		seps, err := dPG.GetServiceEndpointsFilter(test.f)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedEPs, seps) {
				t.Errorf("Test %v Failed: Expected service endpoint '%v'; Recieved service endpoint '%v'", i, test.expectedEPs, seps)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgUpsertServiceEndpoint(t *testing.T) {
	tests := []struct {
		sep             *sm.ServiceEndpoint
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedErr     error
	}{{
		sep:             &stest.TestServiceEndpointUpdate1,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(upsertPgServiceEndpointQuery),
		expectedArgs: []driver.Value{
			stest.TestServiceEndpointUpdate1.RfEndpointID,
			stest.TestServiceEndpointUpdate1.RedfishType,
			stest.TestServiceEndpointUpdate1.RedfishSubtype,
			stest.TestServiceEndpointUpdate1.UUID,
			stest.TestServiceEndpointUpdate1.OdataID,
			stest.TestServiceEndpointUpdate1.ServiceInfo,
		},
		expectedErr: nil,
	}, {
		sep:             &stest.TestServiceEndpointUpdate1,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(upsertPgServiceEndpointQuery),
		expectedArgs: []driver.Value{
			stest.TestServiceEndpointUpdate1.RfEndpointID,
			stest.TestServiceEndpointUpdate1.RedfishType,
			stest.TestServiceEndpointUpdate1.RedfishSubtype,
			stest.TestServiceEndpointUpdate1.UUID,
			stest.TestServiceEndpointUpdate1.OdataID,
			stest.TestServiceEndpointUpdate1.ServiceInfo,
		},
		expectedErr: nil,
	}, {
		sep:             nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedErr:     ErrHMSDSArgNil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		err := dPG.UpsertServiceEndpoint(test.sep)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgUpsertServiceEndpoints(t *testing.T) {
	upsertSuffix := "ON CONFLICT(" + serviceEPsRFEndpointIDCol + ", " + serviceEPsRedfishTypeCol + ") DO UPDATE SET " +
		serviceEPsRedfishSubtypeCol + " = EXCLUDED." + serviceEPsRedfishSubtypeCol + ", " +
		serviceEPsUUIDCol + " = EXCLUDED." + serviceEPsUUIDCol + ", " +
		serviceEPsODataIDCol + " = EXCLUDED." + serviceEPsODataIDCol + ", " +
		serviceEPsServiceInfoCol + " = EXCLUDED." + serviceEPsServiceInfoCol

	var servEP sm.ServiceEndpoint

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	insert1, _, _ := sqq.Insert(serviceEPsTable).
		Columns(serviceEPsCols...).
		Values(servEP.RfEndpointID, servEP.RedfishType, servEP.RedfishSubtype, servEP.UUID, servEP.OdataID, servEP.ServiceInfo).
		Suffix(upsertSuffix).ToSql()
	insert2, _, _ := sqq.Insert(serviceEPsTable).
		Columns(serviceEPsCols...).
		Values(servEP.RfEndpointID, servEP.RedfishType, servEP.RedfishSubtype, servEP.UUID, servEP.OdataID, servEP.ServiceInfo).
		Values(servEP.RfEndpointID, servEP.RedfishType, servEP.RedfishSubtype, servEP.UUID, servEP.OdataID, servEP.ServiceInfo).
		Suffix(upsertSuffix).ToSql()
	tests := []struct {
		seps            *sm.ServiceEndpointArray
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedErr     error
	}{{
		seps:            &stest.TestServiceEndpointArrayUpdates,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(insert2),
		expectedArgs: []driver.Value{
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[0].RfEndpointID,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[0].RedfishType,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[0].RedfishSubtype,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[0].UUID,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[0].OdataID,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[0].ServiceInfo,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[1].RfEndpointID,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[1].RedfishType,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[1].RedfishSubtype,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[1].UUID,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[1].OdataID,
			stest.TestServiceEndpointArrayUpdates.ServiceEndpoints[1].ServiceInfo,
		},
		expectedErr: nil,
	}, {
		seps:            &stest.TestServiceEndpointArrayUpdate1,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(insert1),
		expectedArgs: []driver.Value{
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].RfEndpointID,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].RedfishType,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].RedfishSubtype,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].UUID,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].OdataID,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].ServiceInfo,
		},
		expectedErr: nil,
	}, {
		seps:            &stest.TestServiceEndpointArrayUpdate1,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(insert1),
		expectedArgs: []driver.Value{
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].RfEndpointID,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].RedfishType,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].RedfishSubtype,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].UUID,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].OdataID,
			stest.TestServiceEndpointArrayUpdate1.ServiceEndpoints[0].ServiceInfo,
		},
		expectedErr: nil,
	}, {
		seps:            &sm.ServiceEndpointArray{ServiceEndpoints: []*sm.ServiceEndpoint{nil}},
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedErr:     ErrHMSDSArgNil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, int64(len(test.seps.ServiceEndpoints))))
			mockPG.ExpectCommit()
		}

		err := dPG.UpsertServiceEndpoints(test.seps)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgDeleteServiceEndpointByID(t *testing.T) {
	tests := []struct {
		svc             string
		id              string
		dbResult        int64
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedResult  bool
		expectedErr     error
	}{{
		svc:             "UpdateService",
		id:              "x0c0s21b0",
		dbResult:        1,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(deleteServiceByRfEPandRfTypeQuery),
		expectedArgs:    []driver.Value{"x0c0s21b0", "UpdateService"},
		expectedResult:  true,
		expectedErr:     nil,
	}, {
		svc:             "",
		id:              "x0c0s21b0",
		dbResult:        0,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedResult:  false,
		expectedErr:     ErrHMSDSArgNil,
	}, {
		svc:             "UpdateService",
		id:              "",
		dbResult:        0,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedResult:  false,
		expectedErr:     ErrHMSDSArgNil,
	}, {
		svc:             "foo",
		id:              "x0c0s21b0",
		dbResult:        0,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(deleteServiceByRfEPandRfTypeQuery),
		expectedArgs:    []driver.Value{"x0c0s21b0", "foo"},
		expectedResult:  false,
		expectedErr:     nil,
	}, {
		svc:             "UpdateService",
		id:              "x0c0s21b0",
		dbResult:        0,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(deleteServiceByRfEPandRfTypeQuery),
		expectedArgs:    []driver.Value{"x0c0s21b0", "UpdateService"},
		expectedResult:  false,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, test.dbResult))
			mockPG.ExpectCommit()
		}

		didDelete, err := dPG.DeleteServiceEndpointByID(test.svc, test.id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedResult != didDelete {
				t.Errorf("Test %v Failed: Expected didDelete '%v'; Recieved didDelete '%v'", i, test.expectedResult, didDelete)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgDeleteServiceEndpointsAll(t *testing.T) {
	tests := []struct {
		svc             string
		id              string
		dbResult        int64
		dbError         error
		expectedPrepare string
		expectedResult  int64
		expectedErr     error
	}{{
		dbResult:        1,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(deleteServiceEndpointsAllQuery),
		expectedResult:  1,
	}, {
		dbResult:        0,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(deleteServiceEndpointsAllQuery),
		expectedResult:  0,
	}, {
		dbResult:        0,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(deleteServiceEndpointsAllQuery),
		expectedResult:  0,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnResult(sqlmock.NewResult(0, test.dbResult))
			mockPG.ExpectCommit()
		}

		numDelete, err := dPG.DeleteServiceEndpointsAll()
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedResult != numDelete {
				t.Errorf("Test %v Failed: Expected numDelete '%v'; Recieved numDelete '%v'", i, test.expectedResult, numDelete)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//
// Component Ethernet Interface Management
//
////////////////////////////////////////////////////////////////////////////

func TestPgGetCompEthInterfaceFilter(t *testing.T) {
	columns := addAliasToCols(compEthAlias, compEthCols, compEthCols)

	// TODO: Add Network test for testing, test IPAddress/Network Filters
	testCompEth1 := sm.CompEthInterfaceV2{
		ID:      "a4bf0138ee65",
		Desc:    "My description",
		MACAddr: "a4:bf:01:38:ee:65",
		IPAddrs: []sm.IPAddressMapping{
			{IPAddr: "10.254.2.14"},
		},
		LastUpdate: "2020-05-13T21:59:02.363448Z",
		CompID:     "x3000c0s26b0",
		Type:       "NodeBMC",
	}
	testCompEth1IPAddrsRaw, _ := json.Marshal(testCompEth1.IPAddrs)

	testCompEth2 := sm.CompEthInterfaceV2{
		ID:      "a4bf0138ee67",
		Desc:    "My description 2",
		MACAddr: "a4:bf:01:38:ee:67",
		IPAddrs: []sm.IPAddressMapping{
			{IPAddr: "10.254.2.14"},
		},
		LastUpdate: "2020-05-13T21:59:02.363448Z",
		CompID:     "x3000c0s26b0",
		Type:       "NodeBMC",
	}
	testCompEth2IPAddrsRaw, _ := json.Marshal(testCompEth2.IPAddrs)

	newerThanArg, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	olderThanArg, _ := time.Parse(time.RFC3339, "2020-02-01T00:00:00Z")

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query1, _, _ := sqq.Select(columns...).
		From(compEthTable + " " + compEthAlias).ToSql()

	query2, _, _ := sqq.Select(columns...).
		From(compEthTable + " " + compEthAlias).
		Where(sq.Eq{compEthAlias + "." + compEthIdCol: []string{testCompEth1.ID}}).
		Where(sq.Gt{compEthAlias + "." + compEthLastUpdateCol: newerThanArg}).
		Where(sq.Lt{compEthAlias + "." + compEthLastUpdateCol: olderThanArg}).ToSql()

	tests := []struct {
		f_opts          []CompEthInterfaceFiltFunc
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedOut     []*sm.CompEthInterfaceV2
		expectedErr     error
	}{{
		f_opts: []CompEthInterfaceFiltFunc{},
		dbRows: [][]driver.Value{
			[]driver.Value{testCompEth1.ID, testCompEth1.Desc, testCompEth1.MACAddr, testCompEth1.LastUpdate, testCompEth1.CompID, testCompEth1.Type, testCompEth1IPAddrsRaw},
			[]driver.Value{testCompEth2.ID, testCompEth2.Desc, testCompEth2.MACAddr, testCompEth2.LastUpdate, testCompEth2.CompID, testCompEth2.Type, testCompEth2IPAddrsRaw},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query1),
		expectedArgs:    []driver.Value{},
		expectedOut: []*sm.CompEthInterfaceV2{
			&testCompEth1,
			&testCompEth2,
		},
		expectedErr: nil,
	}, {
		f_opts: []CompEthInterfaceFiltFunc{
			CEI_ID(testCompEth1.ID),
			CEI_NewerThan("2020-01-01T00:00:00Z"),
			CEI_OlderThan("2020-02-01T00:00:00Z"),
		},
		dbRows: [][]driver.Value{
			[]driver.Value{testCompEth1.ID, testCompEth1.Desc, testCompEth1.MACAddr, testCompEth1.LastUpdate, testCompEth1.CompID, testCompEth1.Type, testCompEth1IPAddrsRaw},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(query2),
		expectedArgs:    []driver.Value{testCompEth1.ID, newerThanArg, olderThanArg},
		expectedOut: []*sm.CompEthInterfaceV2{
			&testCompEth1,
		},
		expectedErr: nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
		} else if test.expectedErr == nil {
			if len(test.expectedArgs) > 0 {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
		}

		out, err := dPG.GetCompEthInterfaceFilter(test.f_opts...)
		if test.expectedErr == nil {
			// ensure all expectations have been met
			if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
				t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
			}
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedOut, out) {
				t.Errorf("Test %v Failed: Expected CompEthInterfaces '%v'; Received CompEthInterfaces '%v'", i, test.expectedOut, out)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestInsertCompEthInterfaces(t *testing.T) {
	testCompEth1 := sm.CompEthInterfaceV2{
		ID:      "a4bf0138ee65",
		Desc:    "My description",
		MACAddr: "a4:bf:01:38:ee:65",
		IPAddrs: []sm.IPAddressMapping{
			{IPAddr: "10.254.2.14"},
		},
		CompID: "x3000c0s26b0",
		Type:   "NodeBMC",
	}
	testCompEth1IPAddrsRaw, _ := json.Marshal(testCompEth1.IPAddrs)

	testCompEth2 := sm.CompEthInterfaceV2{
		MACAddr: "a4:bf:01:38:ee:67",
	}
	testCompEth2IPAddrsRaw, _ := json.Marshal(testCompEth2.IPAddrs)

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	insert1, _, _ := sqq.Insert(compEthTable).
		Columns(compEthCols...).
		Values("a4bf0138ee65", "My description", "a4:bf:01:38:ee:65", "NOW()", "x3000c0s26b0", "NodeBMC", `[{"IPAddress":"10.254.2.14","Network":""}]`).
		Values("a4bf0138ee65", "My description", "a4:bf:01:38:ee:65", "NOW()", "x3000c0s26b0", "NodeBMC", `[{"IPAddress":"10.254.2.14","Network":""}]`).ToSql()

	tests := []struct {
		in              []*sm.CompEthInterfaceV2
		expectedPrepare string
		expectedArgs    []driver.Value
		dbError         error
	}{{ // Test 0 - Insert 2 new rows
		in:              []*sm.CompEthInterfaceV2{&testCompEth1, &testCompEth2},
		expectedPrepare: regexp.QuoteMeta(insert1),
		expectedArgs: []driver.Value{
			testCompEth1.ID, testCompEth1.Desc, testCompEth1.MACAddr, "NOW()", testCompEth1.CompID, testCompEth1.Type, testCompEth1IPAddrsRaw,
			"a4bf0138ee67", testCompEth2.Desc, testCompEth2.MACAddr, "NOW()", testCompEth2.CompID, testCompEth2.Type, testCompEth2IPAddrsRaw,
		},
		dbError: nil,
	}, { // Test 1 - Test that database error is passed back
		in:              []*sm.CompEthInterfaceV2{&testCompEth1, &testCompEth2},
		expectedPrepare: regexp.QuoteMeta(insert1),
		expectedArgs: []driver.Value{
			testCompEth1.ID, testCompEth1.Desc, testCompEth1.MACAddr, "NOW()", testCompEth1.CompID, testCompEth1.Type, testCompEth1IPAddrsRaw,
			"a4bf0138ee67", testCompEth2.Desc, testCompEth2.MACAddr, "NOW()", testCompEth2.CompID, testCompEth2.Type, testCompEth2IPAddrsRaw,
		},
		dbError: sql.ErrNoRows,
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.expectedPrepare == "" && test.dbError != nil {
			mockPG.ExpectRollback()
		} else if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, int64(len(test.in))))
			mockPG.ExpectCommit()
		}

		err := dPG.InsertCompEthInterfaces(test.in)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.dbError != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestDeleteCompEthInterfaceByID(t *testing.T) {

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	delete1, _, _ := sqq.Delete(compEthTable).
		Where(sq.Eq{compEthIdCol: "a4bf0138ee65"}).ToSql()

	tests := []struct {
		id              string
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedError   error
	}{{
		id:              "a4bf0138ee65",
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedArgs:    []driver.Value{"a4bf0138ee65"},
		expectedError:   nil,
	}, {
		id:              "a4bf0138ee65",
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedArgs:    []driver.Value{"a4bf0138ee65"},
		expectedError:   sql.ErrNoRows,
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.expectedError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		_, err := dPG.DeleteCompEthInterfaceByID(test.id)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.expectedError == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.expectedError != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error (%s).", i, test.expectedError)
		}
	}
}

func TestDeleteCompEthInterfacesAll(t *testing.T) {

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	delete1, _, _ := sqq.Delete(compEthTable).ToSql()

	tests := []struct {
		expectedPrepare string
		expectedError   error
	}{{
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedError:   nil,
	}, {
		expectedPrepare: regexp.QuoteMeta(delete1),
		expectedError:   sql.ErrNoRows,
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.expectedError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		_, err := dPG.DeleteCompEthInterfacesAll()
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.expectedError == nil && err != nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if test.expectedError != nil && err == nil {
			t.Errorf("Test %v Failed: Expected an error (%s).", i, test.expectedError)
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//
// State Change Notification (SCN) Subscriptions
//
////////////////////////////////////////////////////////////////////////////

func TestGetSCNSubscriptionsAll(t *testing.T) {
	tests := []struct {
		dbColumns       []string
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedSCNSubs *sm.SCNSubscriptionArray
	}{{
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{2, `{"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}`},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(tGetSCNSubscriptionQueryAll),
		expectedSCNSubs: &sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "hmfd@sms01",
				States:     []string{"On", "Off"},
				Url:        "https://foo/bar",
			},
		}},
	}, {
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{2, `{"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}`},
			[]driver.Value{3, `{"Subscriber":"hmfd@sms02","States":["Off","Ready"],"Url":"https://foo2/bar"}`},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(tGetSCNSubscriptionQueryAll),
		expectedSCNSubs: &sm.SCNSubscriptionArray{[]sm.SCNSubscription{
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
	}, {
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{1, `{"Subscriber":"cray-hmnfd-798784bd66-s69mg_4","Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`},
			[]driver.Value{2, `{"Subscriber":"cray-hmnfd-798784bd66-rkvzn_4","Roles":["service"],"States":["Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`},
			[]driver.Value{3, `{"Subscriber":"cray-hmnfd-798784bd66-bp8rc_23","Roles":["compute"],"States":["Empty","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(tGetSCNSubscriptionQueryAll),
		expectedSCNSubs: &sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{
				ID:         1,
				Subscriber: "cray-hmnfd-798784bd66-s69mg_4",
				Roles:      []string{"compute", "service"},
				States:     []string{"Empty", "Populated", "Off", "On", "Standby", "Halt", "Ready"},
				Url:        "http://cray-hmnfd/hmi/v1/scn",
			},
			sm.SCNSubscription{
				ID:         2,
				Subscriber: "cray-hmnfd-798784bd66-rkvzn_4",
				Roles:      []string{"service"},
				States:     []string{"Populated", "Off", "On", "Standby", "Halt", "Ready"},
				Url:        "http://cray-hmnfd/hmi/v1/scn",
			},
			sm.SCNSubscription{
				ID:         3,
				Subscriber: "cray-hmnfd-798784bd66-bp8rc_23",
				Roles:      []string{"compute"},
				States:     []string{"Empty", "Off", "On", "Standby", "Halt", "Ready"},
				Url:        "http://cray-hmnfd/hmi/v1/scn",
			},
		}},
	}, {
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{"", `{}`},
		},
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(tGetSCNSubscriptionQueryAll),
		expectedSCNSubs: &sm.SCNSubscriptionArray{[]sm.SCNSubscription{
			sm.SCNSubscription{},
		}},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		subs, err := dPG.GetSCNSubscriptionsAll()
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedSCNSubs, subs) {
				t.Errorf("Test %v Failed: Expected subArray '%v'; Recieved subArray '%v'", i, test.expectedSCNSubs, subs)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestGetSCNSubscription(t *testing.T) {
	tests := []struct {
		id              int64
		dbColumns       []string
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedSCNSub  *sm.SCNSubscription
	}{{
		id:        2,
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{2, `{"Subscriber":"cray-hmnfd-69d99579c5-shpmk_3","States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(tGetSCNSubscriptionQueryId),
		expectedArgs:    []driver.Value{2},
		expectedSCNSub: &sm.SCNSubscription{
			ID:         2,
			Subscriber: "cray-hmnfd-69d99579c5-shpmk_3",
			States:     []string{"Empty", "Populated", "Off", "On", "Standby", "Halt", "Ready"},
			Url:        "http://cray-hmnfd/hmi/v1/scn",
		},
	}, {
		id:        3,
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{3, `{"Subscriber":"cray-hmnfd-798784bd66-s69mg_4","Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(tGetSCNSubscriptionQueryId),
		expectedArgs:    []driver.Value{3},
		expectedSCNSub: &sm.SCNSubscription{
			ID:         3,
			Subscriber: "cray-hmnfd-798784bd66-s69mg_4",
			Roles:      []string{"compute", "service"},
			States:     []string{"Empty", "Populated", "Off", "On", "Standby", "Halt", "Ready"},
			Url:        "http://cray-hmnfd/hmi/v1/scn",
		},
	}, {
		id:        0,
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{"", `{}`},
		},
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(tGetSCNSubscriptionQueryId),
		expectedArgs:    []driver.Value{},
		expectedSCNSub:  &sm.SCNSubscription{},
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		sub, err := dPG.GetSCNSubscription(test.id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedSCNSub, sub) {
				t.Errorf("Test %v Failed: Expected subArray '%v'; Recieved subArray '%v'", i, test.expectedSCNSub, sub)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestInsertSCNSubscription(t *testing.T) {
	tests := []struct {
		sub             sm.SCNPostSubscription
		dbColumns       []string
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedID      int64
	}{{
		sub: sm.SCNPostSubscription{
			Subscriber: "hmfd@sms01",
			States:     []string{"On", "Off"},
			Url:        "https://foo/bar",
		},
		dbColumns: []string{"id"},
		dbRows: [][]driver.Value{
			[]driver.Value{4},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(tInsertSCNSubscription),
		expectedArgs:    []driver.Value{"hmfd@sms01https://foo/bar", json.RawMessage(`{"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}`)},
		expectedID:      4,
	}, {
		sub: sm.SCNPostSubscription{
			Subscriber: "cray-hmnfd-798784bd66-s69mg_4",
			Roles:      []string{"compute", "service"},
			States:     []string{"Empty", "Populated", "Off", "On", "Standby", "Halt", "Ready"},
			Url:        "http://cray-hmnfd/hmi/v1/scn",
		},
		dbColumns: []string{"id"},
		dbRows: [][]driver.Value{
			[]driver.Value{5},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(tInsertSCNSubscription),
		expectedArgs:    []driver.Value{"cray-hmnfd-798784bd66-s69mg_4http://cray-hmnfd/hmi/v1/scn", json.RawMessage(`{"Subscriber":"cray-hmnfd-798784bd66-s69mg_4","Roles":["compute","service"],"States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`)},
		expectedID:      5,
	}, {
		sub:       sm.SCNPostSubscription{},
		dbColumns: []string{"id"},
		dbRows: [][]driver.Value{
			[]driver.Value{0},
		},
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(tInsertSCNSubscription),
		expectedArgs:    []driver.Value{"", json.RawMessage(`{}`)},
		expectedID:      0,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectPrepare(regexp.QuoteMeta(tInsertSCNSubscriptionQueryLastVal)).ExpectQuery().WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		id, err := dPG.InsertSCNSubscription(test.sub)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedID != id {
				t.Errorf("Test %v Failed: Expected id '%v'; Recieved id '%v'", i, test.expectedID, id)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestUpdateSCNSubscription(t *testing.T) {
	tests := []struct {
		id                int64
		sub               sm.SCNPostSubscription
		dbError           error
		expectedPrepare   string
		expectedArgs      []driver.Value
		expectedDidUpdate bool
	}{{
		id: 4,
		sub: sm.SCNPostSubscription{
			Subscriber: "hmfd@sms01",
			States:     []string{"On", "Off"},
			Url:        "https://foo/bar",
		},
		dbError:           nil,
		expectedPrepare:   regexp.QuoteMeta(tUpdateSCNSubscription),
		expectedArgs:      []driver.Value{"hmfd@sms01https://foo/bar", json.RawMessage(`{"Subscriber":"hmfd@sms01","States":["On","Off"],"Url":"https://foo/bar"}`), 4},
		expectedDidUpdate: true,
	}, {
		id:                0,
		sub:               sm.SCNPostSubscription{},
		dbError:           sql.ErrNoRows,
		expectedPrepare:   regexp.QuoteMeta(tUpdateSCNSubscription),
		expectedArgs:      []driver.Value{"", json.RawMessage(`{}`), 0},
		expectedDidUpdate: false,
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		didUpdate, err := dPG.UpdateSCNSubscription(test.id, test.sub)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedDidUpdate && !didUpdate {
				t.Errorf("Test %v Failed: Expected update; no update occurred", i)
			} else if !test.expectedDidUpdate && didUpdate {
				t.Errorf("Test %v Failed: Did not expect update; update occurred", i)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPatchSCNSubscription(t *testing.T) {
	tests := []struct {
		id                    int64
		op                    string
		sub                   sm.SCNPatchSubscription
		dbColumns             []string
		dbRows                [][]driver.Value
		dbError               error
		expectedQueryPrepare  string
		expectedUpdatePrepare string
		expectedQueryArgs     []driver.Value
		expectedUpdateArgs    []driver.Value
		expectedDidPatch      bool
	}{{
		id: 4,
		op: "add",
		sub: sm.SCNPatchSubscription{
			Op:     "add",
			States: []string{"On", "Off"},
		},
		dbColumns: []string{"id", "subscription"},
		dbRows: [][]driver.Value{
			[]driver.Value{4, `{"Subscriber":"cray-hmnfd-69d99579c5-shpmk_3","States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`},
		},
		dbError:               nil,
		expectedQueryPrepare:  regexp.QuoteMeta(tGetSCNSubscriptionUpdate),
		expectedUpdatePrepare: regexp.QuoteMeta(tUpdateSCNSubscription),
		expectedQueryArgs:     []driver.Value{4},
		expectedUpdateArgs:    []driver.Value{"cray-hmnfd-69d99579c5-shpmk_3http://cray-hmnfd/hmi/v1/scn", json.RawMessage(`{"Subscriber":"cray-hmnfd-69d99579c5-shpmk_3","States":["Empty","Populated","Off","On","Standby","Halt","Ready"],"Url":"http://cray-hmnfd/hmi/v1/scn"}`), 4},
		expectedDidPatch:      true,
	}, {
		id:        0,
		op:        "remove",
		sub:       sm.SCNPatchSubscription{},
		dbColumns: []string{"", ""},
		dbRows: [][]driver.Value{
			[]driver.Value{0, `{}`},
		},
		dbError:               sql.ErrNoRows,
		expectedQueryPrepare:  regexp.QuoteMeta(tGetSCNSubscriptionUpdate),
		expectedUpdatePrepare: regexp.QuoteMeta(tUpdateSCNSubscription),
		expectedQueryArgs:     []driver.Value{},
		expectedUpdateArgs:    []driver.Value{"", json.RawMessage(`{}`), 0},
		expectedDidPatch:      false,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		didPatch, err := dPG.PatchSCNSubscription(test.id, test.op, test.sub)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedDidPatch && !didPatch {
				t.Errorf("Test %v Failed: Expected patch; no patch occurred", i)
			} else if !test.expectedDidPatch && didPatch {
				t.Errorf("Test %v Failed: Did not expect patch; patch occurred", i)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestDeleteSCNSubscription(t *testing.T) {
	tests := []struct {
		id                int64
		dbError           error
		expectedPrepare   string
		expectedArgs      []driver.Value
		expectedDidDelete bool
	}{{
		id:                5,
		dbError:           nil,
		expectedPrepare:   regexp.QuoteMeta(tDeleteSCNSubscription),
		expectedArgs:      []driver.Value{5},
		expectedDidDelete: true,
	}, {
		id:                0,
		dbError:           sql.ErrNoRows,
		expectedPrepare:   regexp.QuoteMeta(tDeleteSCNSubscription),
		expectedArgs:      []driver.Value{},
		expectedDidDelete: false,
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		didDelete, err := dPG.DeleteSCNSubscription(test.id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedDidDelete && !didDelete {
				t.Errorf("Test %v Failed: Expected deletion; no deletion occurred", i)
			} else if !test.expectedDidDelete && didDelete {
				t.Errorf("Test %v Failed: Did not expect deletion; deletion occurred", i)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestDeleteSCNSubscriptionsAll(t *testing.T) {
	tests := []struct {
		dbError                error
		expectedPrepare        string
		expectedNumSubsDeleted int64
	}{{
		dbError:                nil,
		expectedPrepare:        regexp.QuoteMeta(tDeleteSCNSubscriptionAll),
		expectedNumSubsDeleted: 1,
	}, {
		dbError:                nil,
		expectedPrepare:        regexp.QuoteMeta(tDeleteSCNSubscriptionAll),
		expectedNumSubsDeleted: 5,
	}, {
		dbError:                sql.ErrNoRows,
		expectedPrepare:        regexp.QuoteMeta(tDeleteSCNSubscriptionAll),
		expectedNumSubsDeleted: 0,
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnResult(sqlmock.NewResult(0, test.expectedNumSubsDeleted))
			mockPG.ExpectCommit()
		}

		numSubsDeleted, err := dPG.DeleteSCNSubscriptionsAll()
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedNumSubsDeleted != numSubsDeleted {
				t.Errorf("Test %v Failed: Expected deletion of '%v' subs; Deleted '%v'", i, test.expectedNumSubsDeleted, numSubsDeleted)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//
// Group and Partition  Management
//
////////////////////////////////////////////////////////////////////////////

//
// Mock entries and arguments to share amongst tests
//

var uuid1 = uuid.New().String()
var uuid2 = uuid.New().String()
var uuid3 = uuid.New().String()
var uuid4 = uuid.New().String()
var uuid5 = uuid.New().String()
var uuid6 = uuid.New().String()

var tags1 = []string{}
var tags2 = []string{"tag1", "otag1"}
var tags3 = []string{"otag1"}
var tags4 = []string{"tag1"}
var tags5 = []string{"otag1", "pp1"}
var tags6 = []string{"tag1", "pp1"}

var dgrp1 = &sm.Group{
	Label:          "grp1",
	Description:    "desc1",
	Tags:           tags1,
	ExclusiveGroup: "",
	Members: sm.Members{
		IDs: []string{"x0c0s0b0n0", "x0c0s0b0n1"},
	},
}

var dgrp1c1 = []driver.Value{"x0c0s0b0n0"}
var dgrp1c2 = []driver.Value{"x0c0s0b0n1"}

var dval1c1 = []driver.Value{"x0c0s0b0n0", uuid1, "grp1"}
var dval1c2 = []driver.Value{"x0c0s0b0n1", uuid1, "grp1"}

// Member list filtered on dgrp5p
var dgrp1f = &sm.Group{
	Label:          "grp1",
	Description:    "desc1",
	Tags:           tags1,
	ExclusiveGroup: "",
	Members: sm.Members{
		IDs: []string{"x0c0s0b0n1"},
	},
}

var dgrp1fc1 = []driver.Value{"x0c0s0b0n1"}

var dval1fc1 = []driver.Value{"x0c0s0b0n1", uuid1, "grp1"}

var dgrp2 = &sm.Group{
	Label:          "grp2",
	Description:    "desc2",
	Tags:           tags2,
	ExclusiveGroup: "",
	Members: sm.Members{
		IDs: []string{"x0c0s0b0n0", "x0c0s0b1n0"},
	},
}

var dgrp2c1 = []driver.Value{"x0c0s0b0n0"}
var dgrp2c2 = []driver.Value{"x0c0s0b1n0"}

var dval2c1 = []driver.Value{"x0c0s0b0n0", uuid2, "grp2"}
var dval2c2 = []driver.Value{"x0c0s0b1n0", uuid2, "grp2"}

var dgrp3x = &sm.Group{
	Label:          "grp3x",
	Description:    "desc3",
	Tags:           tags3,
	ExclusiveGroup: "exgrp1",
	Members: sm.Members{
		IDs: []string{"x0c0s0b0n1"},
	},
}

var dgrp3c1 = []driver.Value{"x0c0s0b0n1"}

var dval3c1 = []driver.Value{"x0c0s0b0n1", uuid3, "%exgrp1%"}

var dgrp4x = &sm.Group{
	Label:          "grp4x",
	Description:    "desc4",
	Tags:           tags4,
	ExclusiveGroup: "exgrp1",
	Members: sm.Members{
		IDs: []string{"x0c0s0b1n0"},
	},
}

var dgrp4c1 = []driver.Value{"x0c0s0b1n0"}

var dval4c1 = []driver.Value{"x0c0s0b1n0", uuid4, "%exgrp1%"}

var dgrp5p = &sm.Partition{
	Name:        "p1",
	Description: "desc5",
	Tags:        tags5,
	Members: sm.Members{
		IDs: []string{"x0c0s0b0n0"},
	},
}

var dgrp5c1 = []driver.Value{"x0c0s0b0n0"}

var dval5c1 = []driver.Value{"x0c0s0b0n0", uuid5, partGroupNamespace}

var dgrp6p = &sm.Partition{
	Name:        "p1.4",
	Description: "desc6",
	Tags:        tags6,
	Members: sm.Members{
		IDs: []string{"x0c0s0b0n0", "x0c0s0b1n0"},
	},
}

var dgrp6c1 = []driver.Value{"x0c0s0b0n0"}
var dgrp6c2 = []driver.Value{"x0c0s0b1n0"}

var dval6c1 = []driver.Value{"x0c0s0b0n0", uuid6, partGroupNamespace}
var dval6c2 = []driver.Value{"x0c0s0b1n0", uuid6, partGroupNamespace}

//
// Helper functions
//

// Define UUID dummy type and implement sqlmocj Match interface.
type AnyUUID struct{}

func (a AnyUUID) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok
}

//
// API Tests
//

func TestPgGetGroup(t *testing.T) {
	columns := compGroupsColsSMGroup // "id", "name", "description", "tags", "exclusive_group_identifier"
	columns2 := compGroupsColsSMPart

	dval1 := []driver.Value{uuid1, dgrp1.Label, dgrp1.Description, pq.Array(&dgrp1.Tags), dgrp1.ExclusiveGroup}
	dval2 := []driver.Value{uuid2, dgrp2.Label, dgrp2.Description, pq.Array(&dgrp2.Tags), dgrp2.ExclusiveGroup}
	dval3 := []driver.Value{uuid3, dgrp3x.Label, dgrp3x.Description, pq.Array(&dgrp3x.Tags), dgrp3x.ExclusiveGroup}
	dval4 := []driver.Value{uuid4, dgrp4x.Label, dgrp4x.Description, pq.Array(&dgrp4x.Tags), dgrp4x.ExclusiveGroup}
	dval5 := []driver.Value{uuid5, dgrp5p.Name, dgrp5p.Description, pq.Array(&dgrp5p.Tags)}
	//dval6 := []driver.Value{uuid6, dgrp6p.Name, dgrp5p.Description, dgrp6p.Tags}

	memberCols := []string{"component_id"}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp1Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp1.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp1MQuery, _, _ := sqq.Select(compGroupMembersColsUser...).
		From(compGroupMembersTable).
		Where("group_id = ?", uuid1).ToSql()

	dgrp2Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp2.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp2MQuery, _, _ := sqq.Select(compGroupMembersColsUser...).
		From(compGroupMembersTable).
		Where("group_id = ?", uuid2).ToSql()

	dgrp3Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp3x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp3MQuery, _, _ := sqq.Select(compGroupMembersColsUser...).
		From(compGroupMembersTable).
		Where("group_id = ?", uuid3).ToSql()

	dgrp4Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp4x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp4MQuery, _, _ := sqq.Select(compGroupMembersColsUser...).
		From(compGroupMembersTable).
		Where("group_id = ?", uuid4).ToSql()

	dgrp1Query2, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp5p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp1MQuery2, _, _ := sqq.Select("a.component_id").
		From(compGroupMembersTable+" a").
		Join(compGroupMembersTable+" b ON a.component_id = b.component_id").
		Where("a.group_id = ?", uuid1).
		Where("b.group_id IN (?,?)", uuid1, uuid1, uuid5).
		GroupBy("a.component_id").
		Having("COUNT(*) = 2").ToSql()

	tests := []struct {
		label                 string
		filt_part             string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbColumns2            []string
		dbRows2               [][]driver.Value
		dbQuery2Error         error
		expectedQuery2Prepare string
		expectedQuery2Args    []driver.Value
		dbMColumns            []string
		dbMRows               [][]driver.Value
		dbMQueryError         error
		expectedMQueryPrepare string
		expectedMQueryArgs    []driver.Value
		expGroup              *sm.Group
		expError              error
	}{{
		dgrp1.Label, // label
		"",          // filt_part
		columns,
		[][]driver.Value{dval1},
		nil,
		regexp.QuoteMeta(dgrp1Query),
		[]driver.Value{dgrp1.Label, groupNamespace},
		columns2,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		memberCols,
		[][]driver.Value{
			dgrp1c1,
			dgrp1c2,
		},
		nil,
		regexp.QuoteMeta(dgrp1MQuery),
		[]driver.Value{uuid1},
		dgrp1,
		nil,
	}, {
		dgrp2.Label, // label
		"",          // filt_part
		columns,
		[][]driver.Value{dval2},
		nil,
		regexp.QuoteMeta(dgrp2Query),
		[]driver.Value{dgrp2.Label, groupNamespace},
		columns2,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		memberCols,
		[][]driver.Value{
			dgrp2c1,
			dgrp2c2,
		},
		nil,
		regexp.QuoteMeta(dgrp2MQuery),
		[]driver.Value{uuid2},
		dgrp2,
		nil,
	}, {
		dgrp3x.Label, // label
		"",           // filt_part
		columns,
		[][]driver.Value{dval3},
		nil,
		regexp.QuoteMeta(dgrp3Query),
		[]driver.Value{dgrp3x.Label, groupNamespace},
		columns2,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		memberCols,
		[][]driver.Value{
			dgrp3c1,
		},
		nil,
		regexp.QuoteMeta(dgrp3MQuery),
		[]driver.Value{uuid3},
		dgrp3x,
		nil,
	}, {
		dgrp4x.Label, // label
		"",           // filt_part
		columns,
		[][]driver.Value{dval4},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{dgrp4x.Label, groupNamespace},
		columns2,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		memberCols,
		[][]driver.Value{
			dgrp4c1,
		},
		nil,
		regexp.QuoteMeta(dgrp4MQuery),
		[]driver.Value{uuid4},
		dgrp4x,
		nil,
	}, {
		dgrp1.Label, // label
		dgrp5p.Name, // filt_part
		columns,
		[][]driver.Value{dval1},
		nil,
		regexp.QuoteMeta(dgrp1Query),
		[]driver.Value{dgrp1.Label, groupNamespace},
		columns2,
		[][]driver.Value{dval5},
		nil,
		regexp.QuoteMeta(dgrp1Query2),
		[]driver.Value{dgrp5p.Name, partNamespace},
		memberCols,
		[][]driver.Value{
			dgrp1fc1,
		},
		nil,
		regexp.QuoteMeta(dgrp1MQuery2),
		[]driver.Value{uuid1, uuid1, uuid5},
		dgrp1f,
		nil,
	}, {
		dgrp1.Label, // label
		dgrp5p.Name, // filt_part
		columns,
		[][]driver.Value{dval1},
		nil,
		regexp.QuoteMeta(dgrp1Query),
		[]driver.Value{dgrp1.Label, groupNamespace},
		columns2,
		[][]driver.Value{},
		sql.ErrNoRows, // Simulates unexpected error
		regexp.QuoteMeta(dgrp1Query2),
		[]driver.Value{dgrp5p.Name, partNamespace},
		memberCols,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		nil,
		sql.ErrNoRows,
	}, {
		dgrp1.Label, // label
		dgrp5p.Name, // filt_part
		columns,
		[][]driver.Value{dval1},
		nil,
		regexp.QuoteMeta(dgrp1Query),
		[]driver.Value{dgrp1.Label, groupNamespace},
		columns2,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp1Query2),
		[]driver.Value{dgrp5p.Name, partNamespace},
		memberCols,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp1MQuery2),
		[]driver.Value{uuid1, uuid1, uuid1},
		nil,
		ErrHMSDSNoPartition,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}
		rows2 := sqlmock.NewRows(test.dbColumns2)
		for _, row := range test.dbRows2 {
			rows2.AddRow(row...)
		}
		mrows := sqlmock.NewRows(test.dbMColumns)
		for _, mrow := range test.dbMRows {
			mrows.AddRow(mrow...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbQuery2Error != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedQuery2Prepare).ExpectQuery().WithArgs(test.expectedQuery2Args...).WillReturnError(test.dbQuery2Error)
			mockPG.ExpectRollback()
		} else if test.dbMQueryError != nil || test.expError != nil { // expError caues MQuery to be skipped.
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.filt_part != "" && len(test.dbRows) > 0 {
				// We skip this if no rows were returned by first query and GetGroup returns an error
				mockPG.ExpectPrepare(test.expectedQuery2Prepare).ExpectQuery().WithArgs(test.expectedQuery2Args...).WillReturnRows(rows2)
			}
			if test.dbMQueryError != nil {
				mockPG.ExpectPrepare(test.expectedMQueryPrepare).ExpectExec().WillReturnError(test.dbMQueryError)
			}
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.filt_part != "" {
				mockPG.ExpectPrepare(test.expectedQuery2Prepare).ExpectQuery().WithArgs(test.expectedQuery2Args...).WillReturnRows(rows2)
			}
			mockPG.ExpectPrepare(test.expectedMQueryPrepare).ExpectQuery().WithArgs(test.expectedMQueryArgs...).WillReturnRows(mrows)
			mockPG.ExpectCommit()
		}

		g, err := dPG.GetGroup(test.label, test.filt_part)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		//if test.dbQueryError == nil && test.dbUpdateError == nil {
		if test.dbQueryError == nil && test.dbQuery2Error == nil &&
			test.dbMQueryError == nil && test.expError == nil {

			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if g == nil {
				t.Errorf("Test %v Failed: Expected non-nil group", i)
			} else {
				if g.Label != test.expGroup.Label {
					t.Errorf("Test %v Failed: Expected label %s (got %s)",
						i, g.Label, test.expGroup.Label)
				}
				if g.Description != test.expGroup.Description {
					t.Errorf("Test %v Failed: Expected description %s (got %s)",
						i, g.Description, test.expGroup.Description)
				}
				if g.ExclusiveGroup != test.expGroup.ExclusiveGroup {
					t.Errorf("Test %v Failed: Expected excl grp %s (got %s)",
						i, g.ExclusiveGroup, test.expGroup.ExclusiveGroup)
				}
				if !reflect.DeepEqual(g.Tags, test.expGroup.Tags) {
					t.Errorf("Test %v Failed: Expected Tags %s (got %s)",
						i, g.Tags, test.expGroup.Tags)
				}
				if !reflect.DeepEqual(g.Members.IDs, test.expGroup.Members.IDs) {
					t.Errorf("Test %v Failed: Expected Members %s (got %s)",
						i, g.Members.IDs, test.expGroup.Members.IDs)
				}
			}
		} else if err != test.expError {
			t.Errorf("Test %v Failed: Expected error %v.", i, err)
		}
	}
}

func TestInsertPgGroup(t *testing.T) {
	dval1 := []driver.Value{AnyUUID{}, dgrp1.Label, dgrp1.Description, pq.Array(&dgrp1.Tags), groupType, groupNamespace, dgrp1.ExclusiveGroup}
	dval2 := []driver.Value{AnyUUID{}, dgrp2.Label, dgrp2.Description, pq.Array(&dgrp2.Tags), groupType, groupNamespace, dgrp2.ExclusiveGroup}
	dval3 := []driver.Value{AnyUUID{}, dgrp3x.Label, dgrp3x.Description, pq.Array(&dgrp3x.Tags), exclGroupType, groupNamespace, dgrp3x.ExclusiveGroup}
	dval4 := []driver.Value{AnyUUID{}, dgrp4x.Label, dgrp4x.Description, pq.Array(&dgrp4x.Tags), exclGroupType, groupNamespace, dgrp4x.ExclusiveGroup}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note we don't use the args here.  We can't in any case because
	// we need driver.Value because it supports AnyUUID{}.  Can't predict
	// the value used by the function because it is generated for each new
	// group.
	dgrp1UpdateGrp, _, _ := sqq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(sq.Expr("?", uuid1), dgrp1.Label, dgrp1.Description,
			pq.Array(&dgrp1.Tags), groupType, groupNamespace,
			dgrp1.ExclusiveGroup).ToSql()

	dgrp1Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values(dgrp1.Members.IDs[0], sq.Expr("?", uuid1),
			dgrp1.Label).
		Values(dgrp1.Members.IDs[1], sq.Expr("?", uuid1),
			dgrp1.Label).ToSql()

	dgrp2UpdateGrp, _, _ := sqq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(sq.Expr("?", uuid2), dgrp2.Label, dgrp2.Description,
			pq.Array(&dgrp2.Tags), groupType, groupNamespace,
			dgrp2.ExclusiveGroup).ToSql()

	dgrp2Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values(dgrp2.Members.IDs[0], sq.Expr("?", uuid2),
			dgrp2.Label).
		Values(dgrp2.Members.IDs[1], sq.Expr("?", uuid2),
			dgrp2.Label).ToSql()

	dgrp3UpdateGrp, _, _ := sqq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(sq.Expr("?", uuid3), dgrp3x.Label, dgrp3x.Description,
			pq.Array(&dgrp3x.Tags), exclGroupType, groupNamespace,
			dgrp3x.ExclusiveGroup).ToSql()

	dgrp3Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values(dgrp3x.Members.IDs[0], sq.Expr("?", uuid3),
			"%"+dgrp3x.ExclusiveGroup+"%").ToSql()

	dgrp4UpdateGrp, _, _ := sqq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(sq.Expr("?", uuid4), dgrp4x.Label, dgrp4x.Description,
			pq.Array(&dgrp4x.Tags), exclGroupType, groupNamespace,
			dgrp4x.ExclusiveGroup).ToSql()

	dgrp4Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values(dgrp4x.Members.IDs[0], sq.Expr("?", uuid4),
			"%"+dgrp4x.ExclusiveGroup+"%").ToSql()

	tests := []struct {
		group                    *sm.Group
		dbUpdateGrpError         error
		expectedUpdateGrpPrepare string
		expectedUpdateGrpArgs    []driver.Value
		dbUpdateError            error
		expectedUpdatePrepare    string
		expectedUpdateArgs       []driver.Value
		expectedMembers          int64
		expectedLabel            string
	}{{
		dgrp1,
		nil,
		regexp.QuoteMeta(dgrp1UpdateGrp),
		dval1,
		nil,
		regexp.QuoteMeta(dgrp1Update),
		[]driver.Value{dgrp1.Members.IDs[0], AnyUUID{}, dgrp1.Label,
			dgrp1.Members.IDs[1], AnyUUID{}, dgrp1.Label},
		int64(len(dgrp1.Members.IDs)),
		strings.ToLower(dgrp1.Label), // label
	}, {
		dgrp2,
		nil,
		regexp.QuoteMeta(dgrp2UpdateGrp),
		dval2,
		nil,
		regexp.QuoteMeta(dgrp2Update),
		[]driver.Value{dgrp2.Members.IDs[0], AnyUUID{}, dgrp2.Label,
			dgrp2.Members.IDs[1], AnyUUID{}, dgrp2.Label},
		int64(len(dgrp2.Members.IDs)),
		strings.ToLower(dgrp2.Label),
	}, {
		dgrp3x,
		nil,
		regexp.QuoteMeta(dgrp3UpdateGrp),
		dval3,
		nil,
		regexp.QuoteMeta(dgrp3Update),
		[]driver.Value{dgrp3x.Members.IDs[0], AnyUUID{}, "%" + dgrp3x.ExclusiveGroup + "%"},
		int64(len(dgrp3x.Members.IDs)),
		strings.ToLower(dgrp3x.Label), // label
	}, {
		dgrp4x,
		nil,
		regexp.QuoteMeta(dgrp4UpdateGrp),
		dval4,
		nil,
		regexp.QuoteMeta(dgrp4Update),
		[]driver.Value{dgrp4x.Members.IDs[0], AnyUUID{}, "%" + dgrp4x.ExclusiveGroup + "%"},
		int64(len(dgrp4x.Members.IDs)),
		strings.ToLower(dgrp4x.Label), // label
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.expectedUpdateGrpPrepare == "" && test.dbUpdateGrpError != nil {
			mockPG.ExpectRollback()
		} else if test.dbUpdateGrpError != nil {
			mockPG.ExpectPrepare(test.expectedUpdateGrpPrepare).ExpectQuery().WillReturnError(test.dbUpdateGrpError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(test.expectedUpdateGrpPrepare).ExpectExec().WithArgs(test.expectedUpdateGrpArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedUpdateGrpPrepare).ExpectExec().WithArgs(test.expectedUpdateGrpArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, test.expectedMembers))
			mockPG.ExpectCommit()
		}

		label, err := dPG.InsertGroup(test.group)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbUpdateGrpError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if label == "" {
				t.Errorf("Test %v Failed: Expected non-empty label", i)
			} else {
				if label != test.expectedLabel {
					t.Errorf("Test %v Failed: Expected label %s (got %s)",
						i, label, test.expectedLabel)
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgUpdateGroup(t *testing.T) {
	newDescription := "newDescription" // shouldn't match any existing desc

	columns := compGroupsColsSMGroup // "id", "name", "description", "tags", "exclusive_group_identifier"
	//
	dval1 := []driver.Value{uuid1, dgrp1.Label, dgrp1.Description, pq.Array(&dgrp1.Tags), dgrp1.ExclusiveGroup}
	dval2 := []driver.Value{uuid2, dgrp2.Label, dgrp2.Description, pq.Array(&dgrp2.Tags), dgrp2.ExclusiveGroup}
	dval3 := []driver.Value{uuid3, dgrp3x.Label, dgrp3x.Description, pq.Array(&dgrp3x.Tags), dgrp3x.ExclusiveGroup}
	dval4 := []driver.Value{uuid4, dgrp4x.Label, dgrp4x.Description, pq.Array(&dgrp4x.Tags), dgrp4x.ExclusiveGroup}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp1Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp1.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp1Update, _, _ := sqq.Update("").Table(compGroupsTable).
		Where("id = ?", uuid1).
		Set(compGroupDescCol, newDescription).
		Set(compGroupTagsCol, pq.Array(&dgrp2.Tags)).ToSql()

	dgrp2Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp2.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	//dgrp2Update, _, _ := sqq.Update("").Table(compGroupsTable).
	//	Where("id = ?", uuid2).ToSql()

	dgrp3Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp3x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	//dgrp3Update, _, _ := sqq.Update("").Table(compGroupsTable).
	//	Where("id = ?", uuid3).ToSql()

	dgrp4Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp4x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp4Update, _, _ := sqq.Update("").Table(compGroupsTable).
		Where("id = ?", uuid4).
		Set(compGroupTagsCol, pq.Array(&dgrp3x.Tags)).ToSql()

	tests := []struct {
		label                 string
		gp                    *sm.GroupPatch
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedError         error
	}{{
		dgrp1.Label, // label
		&sm.GroupPatch{Description: &newDescription, Tags: &dgrp2.Tags},
		columns,
		[][]driver.Value{dval1},
		nil,
		regexp.QuoteMeta(dgrp1Query),
		[]driver.Value{dgrp1.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp1Update),
		[]driver.Value{newDescription, pq.Array(&dgrp2.Tags), uuid1},
		nil,
	}, {
		strings.ToUpper(dgrp2.Label), // label
		&sm.GroupPatch{},
		columns,
		[][]driver.Value{dval2},
		nil,
		regexp.QuoteMeta(dgrp2Query),
		[]driver.Value{dgrp2.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		nil,
	}, {
		strings.ToUpper(dgrp3x.Label), // label
		&sm.GroupPatch{Description: &dgrp3x.Description, Tags: &dgrp3x.Tags},
		columns,
		[][]driver.Value{dval3},
		nil,
		regexp.QuoteMeta(dgrp3Query),
		[]driver.Value{dgrp3x.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		nil,
	}, {
		dgrp4x.Label, // label
		&sm.GroupPatch{Tags: &dgrp3x.Tags},
		columns,
		[][]driver.Value{dval4},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{dgrp4x.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp4Update),
		[]driver.Value{pq.Array(&dgrp3x.Tags), uuid4},
		nil,
	}, {
		dgrp4x.Label, // label
		&sm.GroupPatch{Tags: &dgrp3x.Tags},
		columns,
		[][]driver.Value{dval4},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{dgrp4x.Label, groupNamespace},
		sql.ErrNoRows,
		regexp.QuoteMeta(dgrp4Update),
		[]driver.Value{pq.Array(&dgrp3x.Tags), uuid4},
		sql.ErrNoRows,
	}, {
		"badlabel", // label
		&sm.GroupPatch{},
		columns,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{"badlabel", groupNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		ErrHMSDSNoGroup,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil || test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.dbUpdateError != nil {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			}
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.expectedUpdatePrepare != "" {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			mockPG.ExpectCommit()
		}

		err := dPG.UpdateGroup(test.label, test.gp)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil && test.expectedError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
		} else if err != test.expectedError {
			t.Errorf("Test %v Failed: Expected an error (%s).",
				i, test.expectedError)
		}
	}
}

func TestPgAddGroupMember(t *testing.T) {
	columns := compGroupsColsSMGroup // "id", "name", "description", "tags", "exclusive_group_identifier"
	//
	dval1 := []driver.Value{uuid1, dgrp1.Label, dgrp1.Description, pq.Array(&dgrp1.Tags), dgrp1.ExclusiveGroup}
	dval2 := []driver.Value{uuid2, dgrp2.Label, dgrp2.Description, pq.Array(&dgrp2.Tags), dgrp2.ExclusiveGroup}
	dval3 := []driver.Value{uuid3, dgrp3x.Label, dgrp3x.Description, pq.Array(&dgrp3x.Tags), dgrp3x.ExclusiveGroup}
	dval4 := []driver.Value{uuid4, dgrp4x.Label, dgrp4x.Description, pq.Array(&dgrp4x.Tags), dgrp4x.ExclusiveGroup}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp1Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp1.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp1Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values("x0c0s22b0n0", sq.Expr("?", uuid1),
			dgrp1.Label).ToSql()

	dgrp2Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp2.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp2Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values("x0c0s2b0n0", sq.Expr("?", uuid2),
			dgrp2.Label).ToSql()

	dgrp3Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp3x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp3Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values("x0c0s22b1n1", sq.Expr("?", uuid3),
			"%"+dgrp3x.ExclusiveGroup+"%").ToSql()

	dgrp4Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp4x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp4Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values("x0c0s4b0n0", sq.Expr("?", uuid4),
			"%"+dgrp4x.ExclusiveGroup+"%").ToSql()

	tests := []struct {
		label                 string
		new_id                string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedId            string
		expectedError         error
	}{{
		dgrp1.Label,    // label
		"X00c0s22b0n0", // new_id
		columns,
		[][]driver.Value{dval1},
		nil,
		regexp.QuoteMeta(dgrp1Query),
		[]driver.Value{dgrp1.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp1Update),
		[]driver.Value{"x0c0s22b0n0", uuid1, dgrp1.Label},
		"x0c0s22b0n0",
		nil,
	}, {
		dgrp2.Label,   // label
		"x0c0s2b0n00", // new_id
		columns,
		[][]driver.Value{dval2},
		nil,
		regexp.QuoteMeta(dgrp2Query),
		[]driver.Value{dgrp2.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp2Update),
		[]driver.Value{"x0c0s2b0n0", uuid2, dgrp2.Label},
		"x0c0s2b0n0",
		nil,
	}, {
		dgrp3x.Label,   // label
		"x0c0s022b1n1", // new_id
		columns,
		[][]driver.Value{dval3},
		nil,
		regexp.QuoteMeta(dgrp3Query),
		[]driver.Value{dgrp3x.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp3Update),
		[]driver.Value{"x0c0s22b1n1", uuid3, "%" + dgrp3x.ExclusiveGroup + "%"},
		"x0c0s22b1n1",
		nil,
	}, {
		dgrp4x.Label, // label
		"x0c0s4b0n0", // new_id
		columns,
		[][]driver.Value{dval4},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{dgrp4x.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp4Update),
		[]driver.Value{"x0c0s4b0n0", uuid4, "%" + dgrp4x.ExclusiveGroup + "%"},
		"x0c0s4b0n0",
		nil,
	}, {
		"badlabel",   // label
		"x0c0s4b0n0", // new_id
		columns,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{"badlabel", groupNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		"",
		ErrHMSDSNoGroup,
	}, {
		dgrp4x.Label, // label
		"x0c0s4b0n0", // new_id
		columns,
		[][]driver.Value{dval4},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{dgrp4x.Label, groupNamespace},
		ErrHMSDSDuplicateKey,
		regexp.QuoteMeta(dgrp4Update),
		[]driver.Value{"x0c0s4b0n0", uuid4, "%" + dgrp4x.ExclusiveGroup + "%"},
		"",
		ErrHMSDSDuplicateKey,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil || test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.dbUpdateError != nil {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			}
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		id, err := dPG.AddGroupMember(test.label, test.new_id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil &&
			test.dbUpdateError == nil && test.expectedError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if id == "" {
				t.Errorf("Test %v Failed: Expected non-empty id", i)
			} else {
				if id != test.expectedId {
					t.Errorf("Test %v Failed: Expected label %s (got %s)",
						i, id, test.expectedId)
				}
			}
		} else if err == nil && err != test.expectedError {
			t.Errorf("Test %v Failed: Expected an error (%s).",
				i, test.expectedError)
		}
	}
}

func TestPgDeleteGroupMember(t *testing.T) {
	columns := compGroupsColsSMGroup // "id", "name", "description", "tags", "exclusive_group_identifier"
	//
	dval1 := []driver.Value{uuid1, dgrp1.Label, dgrp1.Description, pq.Array(&dgrp1.Tags), dgrp1.ExclusiveGroup}
	dval2 := []driver.Value{uuid2, dgrp2.Label, dgrp2.Description, pq.Array(&dgrp2.Tags), dgrp2.ExclusiveGroup}
	dval3 := []driver.Value{uuid3, dgrp3x.Label, dgrp3x.Description, pq.Array(&dgrp3x.Tags), dgrp3x.ExclusiveGroup}
	dval4 := []driver.Value{uuid4, dgrp4x.Label, dgrp4x.Description, pq.Array(&dgrp4x.Tags), dgrp4x.ExclusiveGroup}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp1Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp1.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp1Update, _, _ := sqq.Delete(compGroupMembersTable).
		Where("group_id = ?", uuid1).
		Where("component_id = ?",
			base.NormalizeHMSCompID(dgrp1.Members.IDs[0])).ToSql()

	dgrp2Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp2.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp2Update, _, _ := sqq.Delete(compGroupMembersTable).
		Where("group_id = ?", uuid2).
		Where("component_id = ?",
			base.NormalizeHMSCompID(dgrp2.Members.IDs[0])).ToSql()

	dgrp3Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp3x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp3Update, _, _ := sqq.Delete(compGroupMembersTable).
		Where("group_id = ?", uuid3).
		Where("component_id = ?",
			base.NormalizeHMSCompID(dgrp3x.Members.IDs[0])).ToSql()

	dgrp4Query, _, _ := sqq.Select(compGroupsColsSMGroup...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp4x.Label)).
		Where("namespace = ?", groupNamespace).ToSql()

	dgrp4Update, _, _ := sqq.Delete(compGroupMembersTable).
		Where("group_id = ?", uuid4).
		Where("component_id = ?",
			base.NormalizeHMSCompID(dgrp4x.Members.IDs[0])).ToSql()

	tests := []struct {
		label                 string
		del_id                string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedDeleted       bool
		expectedId            string
		expectedError         error
	}{{
		dgrp1.Label, // label
		strings.ToUpper(dgrp1.Members.IDs[0]),
		columns,
		[][]driver.Value{dval1},
		nil,
		regexp.QuoteMeta(dgrp1Query),
		[]driver.Value{dgrp1.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp1Update),
		[]driver.Value{uuid1, dgrp1.Members.IDs[0]},
		true,
		dgrp1.Members.IDs[0],
		nil,
	}, {
		strings.ToUpper(dgrp2.Label), // label
		dgrp2.Members.IDs[1],
		columns,
		[][]driver.Value{dval2},
		nil,
		regexp.QuoteMeta(dgrp2Query),
		[]driver.Value{dgrp2.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp2Update),
		[]driver.Value{uuid2, dgrp2.Members.IDs[1]},
		true,
		dgrp2.Members.IDs[1],
		nil,
	}, {
		strings.ToUpper(dgrp3x.Label), // label
		strings.ToUpper(dgrp3x.Members.IDs[0]),
		columns,
		[][]driver.Value{dval3},
		nil,
		regexp.QuoteMeta(dgrp3Query),
		[]driver.Value{dgrp3x.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp3Update),
		[]driver.Value{uuid3, dgrp3x.Members.IDs[0]},
		true,
		dgrp3x.Members.IDs[0],
		nil,
	}, {
		dgrp4x.Label, // label
		dgrp4x.Members.IDs[0],
		columns,
		[][]driver.Value{dval4},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{dgrp4x.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp4Update),
		[]driver.Value{uuid4, dgrp4x.Members.IDs[0]},
		true,
		dgrp4x.Members.IDs[0],
		nil,
	}, {
		dgrp4x.Label, // label
		dgrp1.Members.IDs[1],
		columns,
		[][]driver.Value{dval4},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{dgrp4x.Label, groupNamespace},
		nil,
		regexp.QuoteMeta(dgrp4Update),
		[]driver.Value{uuid4, dgrp1.Members.IDs[1]},
		false,
		"",
		nil,
	}, {
		"badlabel", // label
		dgrp4x.Members.IDs[0],
		columns,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp4Query),
		[]driver.Value{"badlabel", groupNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{uuid4, dgrp4x.Members.IDs[0]},
		true,
		dgrp4x.Members.IDs[0],
		ErrHMSDSNoGroup,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil || test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.dbUpdateError != nil {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			}
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.expectedDeleted {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			} else {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 0))
			}
			mockPG.ExpectCommit()
		}

		didDelete, err := dPG.DeleteGroupMember(test.label, test.del_id)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil && test.expectedError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if didDelete != test.expectedDeleted {
				t.Errorf("Test %v Failed: Expected %t",
					i, test.expectedDeleted)
			}
		} else if err != test.expectedError {
			t.Errorf("Test %v Failed: Expected an error (%s).",
				i, test.expectedError)
		}
	}
}

//
// Partitions
//

func TestPgGetPartition(t *testing.T) {
	columns := compGroupsColsSMPart // "id", "name", "description", "tags"

	dval5 := []driver.Value{uuid5, dgrp5p.Name, dgrp5p.Description, pq.Array(&dgrp5p.Tags)}
	dval6 := []driver.Value{uuid6, dgrp6p.Name, dgrp6p.Description, pq.Array(&dgrp6p.Tags)}

	memberCols := []string{"component_id"}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp5Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp5p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp5MQuery, _, _ := sqq.Select(compGroupMembersColsUser...).
		From(compGroupMembersTable).
		Where("group_id = ?", uuid5).ToSql()

	dgrp6Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp6p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp6MQuery, _, _ := sqq.Select(compGroupMembersColsUser...).
		From(compGroupMembersTable).
		Where("group_id = ?", uuid6).ToSql()

	tests := []struct {
		name                  string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbMColumns            []string
		dbMRows               [][]driver.Value
		dbMQueryError         error
		expectedMQueryPrepare string
		expectedMQueryArgs    []driver.Value
		expPart               *sm.Partition
	}{{
		dgrp5p.Name, // label
		columns,
		[][]driver.Value{dval5},
		nil,
		regexp.QuoteMeta(dgrp5Query),
		[]driver.Value{dgrp5p.Name, partNamespace},
		memberCols,
		[][]driver.Value{
			dgrp5c1,
		},
		nil,
		regexp.QuoteMeta(dgrp5MQuery),
		[]driver.Value{uuid5},
		dgrp5p,
	}, {
		dgrp6p.Name, // label
		columns,
		[][]driver.Value{dval6},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{dgrp6p.Name, partNamespace},
		memberCols,
		[][]driver.Value{
			dgrp6c1,
			dgrp6c2,
		},
		nil,
		regexp.QuoteMeta(dgrp6MQuery),
		[]driver.Value{uuid6},
		dgrp6p,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}
		mrows := sqlmock.NewRows(test.dbMColumns)
		for _, mrow := range test.dbMRows {
			mrows.AddRow(mrow...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbMQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedMQueryPrepare).ExpectExec().WillReturnError(test.dbMQueryError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedMQueryPrepare).ExpectQuery().WithArgs(test.expectedMQueryArgs...).WillReturnRows(mrows)
			mockPG.ExpectCommit()
		}

		p, err := dPG.GetPartition(test.name)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		//if test.dbQueryError == nil && test.dbUpdateError == nil {
		if test.dbQueryError == nil && test.dbMQueryError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if p == nil {
				t.Errorf("Test %v Failed: Expected non-nil group", i)
			} else {
				if p.Name != test.expPart.Name {
					t.Errorf("Test %v Failed: Expected name %s (got %s)",
						i, p.Name, test.expPart.Name)
				}
				if p.Description != test.expPart.Description {
					t.Errorf("Test %v Failed: Expected description %s (got %s)",
						i, p.Description, test.expPart.Description)
				}
				if !reflect.DeepEqual(p.Tags, test.expPart.Tags) {
					t.Errorf("Test %v Failed: Expected Tags %s (got %s)",
						i, p.Tags, test.expPart.Tags)
				}
				if !reflect.DeepEqual(p.Members.IDs, test.expPart.Members.IDs) {
					t.Errorf("Test %v Failed: Expected Members %s (got %s)",
						i, p.Members.IDs, test.expPart.Members.IDs)
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgInsertPartition(t *testing.T) {
	dval5 := []driver.Value{AnyUUID{}, dgrp5p.Name, dgrp5p.Description, pq.Array(&dgrp5p.Tags), partType, partNamespace, ""}
	dval6 := []driver.Value{AnyUUID{}, dgrp6p.Name, dgrp6p.Description, pq.Array(&dgrp6p.Tags), partType, partNamespace, ""}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note we only use the query here so the args values don't really matter.
	dgrp5UpdateGrp, _, _ := sqq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(sq.Expr("?", uuid5), dgrp5p.Name, dgrp5p.Description,
			pq.Array(&dgrp5p.Tags), partType, partNamespace, "").ToSql()

	dgrp5Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values(dgrp5p.Members.IDs[0], sq.Expr("?", uuid5),
			partGroupNamespace).ToSql()

	dgrp6UpdateGrp, _, _ := sqq.Insert(compGroupsTable).
		Columns(compGroupsColsAll7...).
		Values(sq.Expr("?", uuid6), dgrp6p.Name, dgrp6p.Description,
			pq.Array(&dgrp6p.Tags), partType, partNamespace, "").ToSql()

	dgrp6Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values(dgrp6p.Members.IDs[0], sq.Expr("?", uuid6),
			partGroupNamespace).
		Values(dgrp6p.Members.IDs[1], sq.Expr("?", uuid6),
			partGroupNamespace).ToSql()

	tests := []struct {
		part                     *sm.Partition
		dbUpdateGrpError         error
		expectedUpdateGrpPrepare string
		expectedUpdateGrpArgs    []driver.Value
		dbUpdateError            error
		expectedUpdatePrepare    string
		expectedUpdateArgs       []driver.Value
		expectedMembers          int64
		expectedName             string
	}{{
		dgrp5p,
		nil,
		regexp.QuoteMeta(dgrp5UpdateGrp),
		dval5,
		nil,
		regexp.QuoteMeta(dgrp5Update),
		[]driver.Value{dgrp5p.Members.IDs[0], AnyUUID{}, partGroupNamespace},
		int64(len(dgrp5p.Members.IDs)),
		strings.ToLower(dgrp5p.Name), // Name
	}, {
		dgrp6p,
		nil,
		regexp.QuoteMeta(dgrp6UpdateGrp),
		dval6,
		nil,
		regexp.QuoteMeta(dgrp6Update),
		[]driver.Value{dgrp6p.Members.IDs[0], AnyUUID{}, partGroupNamespace,
			dgrp6p.Members.IDs[1], AnyUUID{}, partGroupNamespace},
		int64(len(dgrp6p.Members.IDs)),
		strings.ToLower(dgrp6p.Name),
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.expectedUpdateGrpPrepare == "" && test.dbUpdateGrpError != nil {
			mockPG.ExpectRollback()
		} else if test.dbUpdateGrpError != nil {
			mockPG.ExpectPrepare(test.expectedUpdateGrpPrepare).ExpectQuery().WillReturnError(test.dbUpdateGrpError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(test.expectedUpdateGrpPrepare).ExpectExec().WithArgs(test.expectedUpdateGrpArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedUpdateGrpPrepare).ExpectExec().WithArgs(test.expectedUpdateGrpArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, test.expectedMembers))
			mockPG.ExpectCommit()
		}

		name, err := dPG.InsertPartition(test.part)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbUpdateGrpError == nil && test.dbUpdateError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if name == "" {
				t.Errorf("Test %v Failed: Expected non-empty label", i)
			} else {
				if name != test.expectedName {
					t.Errorf("Test %v Failed: Expected label %s (got %s)",
						i, name, test.expectedName)
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgUpdatePartition(t *testing.T) {
	newDescription := "newDescription" // shouldn't match any existing desc

	columns := compGroupsColsSMPart // "id", "name", "description", "tags"

	dval5 := []driver.Value{uuid5, dgrp5p.Name, dgrp5p.Description, pq.Array(&dgrp5p.Tags)}
	dval6 := []driver.Value{uuid6, dgrp6p.Name, dgrp5p.Description, pq.Array(&dgrp6p.Tags)}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp5Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp5p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp5Update, _, _ := sqq.Update("").Table(compGroupsTable).
		Where("id = ?", uuid5).
		Set(compGroupDescCol, newDescription).
		Set(compGroupTagsCol, pq.Array(&dgrp6p.Tags)).ToSql()

	dgrp6Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp6p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp6Update, _, _ := sqq.Update("").Table(compGroupsTable).
		Where("id = ?", uuid6).
		Set(compGroupTagsCol, pq.Array(&dgrp5p.Tags)).ToSql()

	tests := []struct {
		label                 string
		gp                    *sm.PartitionPatch
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedError         error
	}{{
		dgrp5p.Name,
		&sm.PartitionPatch{Description: &newDescription, Tags: &dgrp6p.Tags},
		columns,
		[][]driver.Value{dval5},
		nil,
		regexp.QuoteMeta(dgrp5Query),
		[]driver.Value{dgrp5p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(dgrp5Update),
		[]driver.Value{newDescription, pq.Array(&dgrp6p.Tags), uuid5},
		nil,
	}, {
		strings.ToUpper(dgrp5p.Name),
		&sm.PartitionPatch{}, // No updates - do nothing
		columns,
		[][]driver.Value{dval5},
		nil,
		regexp.QuoteMeta(dgrp5Query),
		[]driver.Value{dgrp5p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		nil,
	}, {
		strings.ToUpper(dgrp5p.Name), // No update- should already match
		&sm.PartitionPatch{Description: &dgrp5p.Description, Tags: &dgrp5p.Tags},
		columns,
		[][]driver.Value{dval5},
		nil,
		regexp.QuoteMeta(dgrp5Query),
		[]driver.Value{dgrp5p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		nil,
	}, {
		dgrp6p.Name, // Just tags are different, not description.
		&sm.PartitionPatch{Tags: &dgrp5p.Tags},
		columns,
		[][]driver.Value{dval6},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{dgrp6p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(dgrp6Update),
		[]driver.Value{pq.Array(&dgrp5p.Tags), uuid6},
		nil,
	}, {
		dgrp6p.Name, // Unexpected DB error during update.
		&sm.PartitionPatch{Tags: &dgrp5p.Tags},
		columns,
		[][]driver.Value{dval6},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{dgrp6p.Name, partNamespace},
		sql.ErrNoRows,
		regexp.QuoteMeta(dgrp6Update),
		[]driver.Value{pq.Array(&dgrp5p.Tags), uuid6},
		sql.ErrNoRows,
	}, {
		"badName",
		&sm.PartitionPatch{},
		columns,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{"badname", partNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		ErrHMSDSNoPartition,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil || test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.dbUpdateError != nil {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			}
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.expectedUpdatePrepare != "" {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			}
			mockPG.ExpectCommit()
		}

		err := dPG.UpdatePartition(test.label, test.gp)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil && test.expectedError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
		} else if err != test.expectedError {
			t.Errorf("Test %v Failed: Expected an error (%s).",
				i, test.expectedError)
		}
	}
}

func TestPgAddPartitionMember(t *testing.T) {
	columns := compGroupsColsSMPart // "id", "name", "description", "tags"

	dval5 := []driver.Value{uuid5, dgrp5p.Name, dgrp5p.Description, pq.Array(&dgrp5p.Tags)}
	dval6 := []driver.Value{uuid6, dgrp6p.Name, dgrp6p.Description, pq.Array(&dgrp6p.Tags)}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp5Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp5p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp5Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values("x0c0s22b1n1", sq.Expr("?", uuid5),
			partGroupNamespace).ToSql()

	dgrp6Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp6p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp6Update, _, _ := sqq.Insert(compGroupMembersTable).
		Columns(compGroupMembersColsNoTS...).
		Values("x0c0s4b0n0", sq.Expr("?", uuid6),
			partGroupNamespace).ToSql()

	tests := []struct {
		name                  string
		new_id                string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedId            string
		expectedError         error
	}{{
		dgrp5p.Name,    // name
		"x0c0s022b1n1", // new_id
		columns,
		[][]driver.Value{dval5},
		nil,
		regexp.QuoteMeta(dgrp5Query),
		[]driver.Value{dgrp5p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(dgrp5Update),
		[]driver.Value{"x0c0s22b1n1", uuid5, partGroupNamespace},
		"x0c0s22b1n1",
		nil,
	}, {
		dgrp6p.Name,  // name
		"x0c0s4b0n0", // new_id
		columns,
		[][]driver.Value{dval6},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{dgrp6p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(dgrp6Update),
		[]driver.Value{"x0c0s4b0n0", uuid6, partGroupNamespace},
		"x0c0s4b0n0",
		nil,
	}, {
		"badname",    // name
		"x0c0s4b0n0", // new_id
		columns,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{"badname", partNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		"",
		ErrHMSDSNoPartition,
	}, {
		dgrp6p.Name,  // name
		"x0c0s4b0n0", // new_id
		columns,
		[][]driver.Value{dval6},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{dgrp6p.Name, partNamespace},
		ErrHMSDSDuplicateKey,
		regexp.QuoteMeta(dgrp6Update),
		[]driver.Value{"x0c0s4b0n0", uuid6, partGroupNamespace},
		"",
		ErrHMSDSDuplicateKey,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil || test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.dbUpdateError != nil {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			}
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		id, err := dPG.AddPartitionMember(test.name, test.new_id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil &&
			test.dbUpdateError == nil && test.expectedError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if id == "" {
				t.Errorf("Test %v Failed: Expected non-empty id", i)
			} else {
				if id != test.expectedId {
					t.Errorf("Test %v Failed: Expected name %s (got %s)",
						i, id, test.expectedId)
				}
			}
		} else if err == nil && err != test.expectedError {
			t.Errorf("Test %v Failed: Expected an error (%s).",
				i, test.expectedError)
		}
	}
}

func TestPgDeletePartitionMember(t *testing.T) {
	columns := compGroupsColsSMPart // "id", "name", "description", "tags"

	dval5 := []driver.Value{uuid5, dgrp5p.Name, dgrp5p.Description, pq.Array(&dgrp5p.Tags)}
	dval6 := []driver.Value{uuid6, dgrp6p.Name, dgrp6p.Description, pq.Array(&dgrp6p.Tags)}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dgrp5Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp5p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp5Update, _, _ := sqq.Delete(compGroupMembersTable).
		Where("group_id = ?", uuid5).
		Where("component_id = ?",
			base.NormalizeHMSCompID(dgrp5p.Members.IDs[0])).ToSql()

	dgrp6Query, _, _ := sqq.Select(compGroupsColsSMPart...).
		From(compGroupsTable).
		Where("name = ?", strings.ToLower(dgrp6p.Name)).
		Where("namespace = ?", partNamespace).ToSql()

	dgrp6Update, _, _ := sqq.Delete(compGroupMembersTable).
		Where("group_id = ?", uuid6).
		Where("component_id = ?",
			base.NormalizeHMSCompID(dgrp6p.Members.IDs[0])).ToSql()

	tests := []struct {
		name                  string
		del_id                string
		dbColumns             []string
		dbRows                [][]driver.Value
		dbQueryError          error
		expectedQueryPrepare  string
		expectedQueryArgs     []driver.Value
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
		expectedDeleted       bool
		expectedId            string
		expectedError         error
	}{{
		dgrp5p.Name, // name
		dgrp5p.Members.IDs[0],
		columns,
		[][]driver.Value{dval5},
		nil,
		regexp.QuoteMeta(dgrp5Query),
		[]driver.Value{dgrp5p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(dgrp5Update),
		[]driver.Value{uuid5, dgrp5p.Members.IDs[0]},
		true,
		dgrp5p.Members.IDs[0],
		nil,
	}, {
		dgrp6p.Name, // name
		dgrp6p.Members.IDs[0],
		columns,
		[][]driver.Value{dval6},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{dgrp6p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(dgrp6Update),
		[]driver.Value{uuid6, dgrp6p.Members.IDs[0]},
		true,
		dgrp6p.Members.IDs[0],
		nil,
	}, {
		dgrp6p.Name, // name
		dgrp6p.Members.IDs[1],
		columns,
		[][]driver.Value{dval6},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{dgrp6p.Name, partNamespace},
		nil,
		regexp.QuoteMeta(dgrp6Update),
		[]driver.Value{uuid6, dgrp6p.Members.IDs[1]},
		false,
		"",
		nil,
	}, {
		"badname", // name
		dgrp6p.Members.IDs[0],
		columns,
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(dgrp6Query),
		[]driver.Value{"badname", partNamespace},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		false,
		dgrp6p.Members.IDs[0],
		ErrHMSDSNoPartition,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil || test.expectedError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.dbUpdateError != nil {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			}
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			if test.expectedDeleted {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			} else {
				mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 0))
			}
			mockPG.ExpectCommit()
		}

		didDelete, err := dPG.DeletePartitionMember(test.name, test.del_id)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if test.dbQueryError == nil && test.dbUpdateError == nil && test.expectedError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if didDelete != test.expectedDeleted {
				t.Errorf("Test %v Failed: Expected %t",
					i, test.expectedDeleted)
			}
		} else if err != test.expectedError {
			t.Errorf("Test %v Failed: Expected an error (%s).",
				i, test.expectedError)
		}
	}
}

// Note: GetMembership is basically a special case where f contains ID={id}
func TestPgGetMemberships(t *testing.T) {
	tests := []struct {
		f               *ComponentFilter
		dbColumns       []string
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedMembers []*sm.Membership
		expectedError   error
	}{{
		&ComponentFilter{
			Type: []string{"node"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{
			[]driver.Value{"x0c0s14b0n0", "grp1", "group"},
			[]driver.Value{"x0c0s24b0n0", "grp1", "group"},
			[]driver.Value{"x0c0s24b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp3", "group"},
			[]driver.Value{"x1c0s0b0n0", "part1", "partition"},
		},
		nil,
		regexp.QuoteMeta(tGetMembershipsBaseQuery + tGetCompJoinGroupsQuery + " WHERE c.type IN ($1)"),
		[]driver.Value{"Node"},
		[]*sm.Membership{
			&sm.Membership{ID: "x0c0s14b0n0", GroupLabels: []string{"grp1"}},
			&sm.Membership{ID: "x0c0s24b0n0", GroupLabels: []string{"grp1", "grp2"}},
			&sm.Membership{ID: "x1c0s0b0n0", GroupLabels: []string{"grp2", "grp3"}, PartitionName: "part1"},
		},
		nil,
	}, {
		&ComponentFilter{
			Type: []string{},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{
			[]driver.Value{"x0c0s14b0n0", nil, nil},
			[]driver.Value{"x0c0s24b0n0", "grp1", "group"},
			[]driver.Value{"x0c0s24b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp3", "group"},
			[]driver.Value{"x1c0s0b0n0", "part1", "partition"},
		},
		nil,
		regexp.QuoteMeta(tGetMembershipsBaseQuery + tGetCompJoinGroupsQuery),
		[]driver.Value{},
		[]*sm.Membership{
			&sm.Membership{ID: "x0c0s14b0n0", GroupLabels: []string{}},
			&sm.Membership{ID: "x0c0s24b0n0", GroupLabels: []string{"grp1", "grp2"}},
			&sm.Membership{ID: "x1c0s0b0n0", GroupLabels: []string{"grp2", "grp3"}, PartitionName: "part1"},
		},
		nil,
	}, {
		&ComponentFilter{
			ID: []string{"x1c0s0b0n0"}, // Same as GetMembership
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{
			[]driver.Value{"x1c0s0b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp3", "group"},
			[]driver.Value{"x1c0s0b0n0", "part1", "partition"},
		},
		nil,
		regexp.QuoteMeta(tGetMembershipsBaseQuery + tGetCompJoinGroupsQuery + " WHERE c.id IN ($1)"),
		[]driver.Value{"x1c0s0b0n0"},
		[]*sm.Membership{
			&sm.Membership{ID: "x1c0s0b0n0", GroupLabels: []string{"grp2", "grp3"}, PartitionName: "part1"},
		},
		nil,
	}, {
		&ComponentFilter{
			NID:      []string{"800"},
			NIDStart: []string{"16300", "440"},
			NIDEnd:   []string{"16500", "480"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{
			[]driver.Value{"x0c0s14b0n0", "grp1", "group"},
			[]driver.Value{"x0c0s24b0n0", "grp1", "group"},
			[]driver.Value{"x0c0s24b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp3", "group"},
			[]driver.Value{"x1c0s0b0n0", "part1", "partition"},
		},
		nil,
		regexp.QuoteMeta(tGetMembershipsBaseQuery + tGetCompJoinGroupsQuery + " WHERE ((c.nid >= $1 AND c.nid <= $2) OR (c.nid >= $3 AND c.nid <= $4) OR c.nid = $5)"),
		[]driver.Value{"16300", "16500", "440", "480", "800"},
		[]*sm.Membership{
			&sm.Membership{ID: "x0c0s14b0n0", GroupLabels: []string{"grp1"}},
			&sm.Membership{ID: "x0c0s24b0n0", GroupLabels: []string{"grp1", "grp2"}},
			&sm.Membership{ID: "x1c0s0b0n0", GroupLabels: []string{"grp2", "grp3"}, PartitionName: "part1"},
		},
		nil,
	}, {
		&ComponentFilter{
			Group: []string{"grp2", "grp3"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{
			[]driver.Value{"x0c0s24b0n0", "grp1", "group"},
			[]driver.Value{"x0c0s24b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp3", "group"},
			[]driver.Value{"x1c0s0b0n0", "part1", "partition"},
		},
		nil,
		regexp.QuoteMeta(tGetMembershipsBaseQueryWSub + tGetMembershipsSubQuery +
			tGetCompJoinGroupsQuery +
			" WHERE (cg.name IN ($1,$2) AND cg.namespace = $3)" +
			tGetMembershipsBaseSuffixWSub),
		[]driver.Value{"grp2", "grp3", "group"},
		[]*sm.Membership{
			&sm.Membership{ID: "x0c0s24b0n0", GroupLabels: []string{"grp1", "grp2"}},
			&sm.Membership{ID: "x1c0s0b0n0", GroupLabels: []string{"grp2", "grp3"}, PartitionName: "part1"},
		},
		nil,
	}, {
		&ComponentFilter{
			Group:     []string{"grp2"},
			Partition: []string{"part1"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{
			[]driver.Value{"x1c0s0b0n0", "grp2", "group"},
			[]driver.Value{"x1c0s0b0n0", "grp3", "group"},
			[]driver.Value{"x1c0s0b0n0", "part1", "partition"},
		},
		nil,
		regexp.QuoteMeta(tGetMembershipsBaseQueryWSub + tGetMembershipsSubQuery +
			tGetCompJoinGroupsQuery +
			" WHERE ((cg.name IN ($1) AND cg.namespace = $2) OR (cg.name IN ($3) AND cg.namespace = $4))" +
			tGetMembershipsGroupSuffixWSub),
		[]driver.Value{"grp2", "group", "part1", "partition"},
		[]*sm.Membership{
			&sm.Membership{ID: "x1c0s0b0n0", GroupLabels: []string{"grp2", "grp3"}, PartitionName: "part1"},
		},
		nil,
	}, {
		&ComponentFilter{
			Group:     []string{"NULL"},
			Partition: []string{"NULL"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{
			[]driver.Value{"x0c0s14b0n0", nil, nil},
		},
		nil,
		regexp.QuoteMeta(tGetMembershipsBaseQueryWSub + tGetMembershipsSubQuery +
			tGetCompJoinGroupsQuery +
			" WHERE c" + compGroupMembersNsColAlias + " IS NULL" +
			tGetMembershipsBaseSuffixWSub),
		[]driver.Value{},
		[]*sm.Membership{
			&sm.Membership{ID: "x0c0s14b0n0", GroupLabels: []string{}},
		},
		nil,
	}, {
		&ComponentFilter{
			Group:     []string{"NULL"},
			Partition: []string{"part1"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		[]*sm.Membership{},
		ErrHMSDSNullGroupBadPart,
	}, {
		&ComponentFilter{
			Group:     []string{"grp1"},
			Partition: []string{"NULL"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		[]*sm.Membership{},
		ErrHMSDSNullPartBadGroup,
	}, {
		&ComponentFilter{
			Group: []string{"NULL", "grp1", "grp2"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		[]*sm.Membership{},
		ErrHMSDSNullBadMixGroup,
	}, {
		&ComponentFilter{
			Partition: []string{"p1", "NULL", "p2"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		[]*sm.Membership{},
		ErrHMSDSNullBadMixPart,
	}, {
		&ComponentFilter{
			Group:     []string{"grp1", "grp2"},
			Partition: []string{"part1"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		[]*sm.Membership{},
		ErrHMSDSMultipleGroupAndPart,
	}, {
		&ComponentFilter{
			Group:     []string{"grp1"},
			Partition: []string{"part1", "part2"},
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		[]*sm.Membership{},
		ErrHMSDSMultipleGroupAndPart,
	}, {
		&ComponentFilter{
			Group:     []string{"grp1"},
			Partition: []string{"part$^#$^#$@1"}, // invalid part
		},
		append(compColsIdOnly, compGroupPartCols...),
		[][]driver.Value{},
		nil,
		regexp.QuoteMeta(""),
		[]driver.Value{},
		[]*sm.Membership{},
		ErrHMSDSNoPartition,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
		} else if test.expectedError == nil {
			if len(test.dbColumns) == 0 {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery()
			} else if test.expectedArgs != nil {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			}
		}
		mbs, err := dPG.GetMemberships(test.f)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		// Note if dbError is != nil test.expectedError should be non-nil also
		if test.dbError == nil && test.expectedError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if len(mbs) != len(test.expectedMembers) {
				t.Errorf("Test %v Failed: Expected members '%v'; Got members '%v'",
					i, test.expectedMembers, mbs)
			} else {
				for _, em := range test.expectedMembers {
					found := false
					for _, m := range mbs {
						if reflect.DeepEqual(*m, *em) {
							found = true
							break
						}
					}
					if found != true {
						t.Errorf("Test %v Failed: Expected member '%v'", i, *em)
					}
				}
			}
		} else if err != test.expectedError {
			t.Errorf("Test %v Failed: Expected error '%s', got '%s'",
				i, err, test.expectedError)
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//
// Component Lock Management
//
////////////////////////////////////////////////////////////////////////////

// Define UUID dummy type and implement sqlmocj Match interface.
type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	if _, ok := v.(string); ok {
		return true
	} else if _, ok := v.(time.Time); ok {
		return true
	} else if _, ok := v.(sql.NullTime); ok {
		return true
	} else {
		return false
	}
}

type AnyString struct{}

func (a AnyString) Match(v driver.Value) bool {
	if v == nil {
		return true
	} else if _, ok := v.(string); ok {
		return true
	} else if _, ok := v.(sql.NullString); ok {
		return true
	} else {
		return false
	}
}

func TestPgInsertCompReservations(t *testing.T) {
	resInsert := compReservation{
		component_id: "x3000c0s9b0n0",
		create_timestamp: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		expiration_timestamp: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		deputy_key:      "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17",
		reservation_key: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
	}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note we only use the query here so the args values don't really matter.
	resInsertReservation, _, _ := sqq.Insert(compResTable).
		Columns(compResCols...).
		Values(resInsert.component_id,
			sq.Expr("?", resInsert.create_timestamp),
			sq.Expr("?", resInsert.expiration_timestamp),
			sq.Expr("?", resInsert.deputy_key),
			sq.Expr("?", resInsert.reservation_key)).
		Suffix("ON CONFLICT DO NOTHING RETURNING " + compResCompIdCol + ", " + compResDKCol + ", " + compResRKCol).ToSql()

	tests := []struct {
		f                         sm.CompLockV2Filter
		dbGetCompIDsError         error
		expectedGetCompIDsPrepare string
		expectedGetCompIDsArgs    []driver.Value
		dbGetCompIDsReturnCols    []string
		dbGetCompIDsReturnRows    [][]driver.Value
		dbInsertError             error
		expectedInsertPrepare     string
		expectedInsertArgs        []driver.Value
		dbInsertV2ResReturnCols   []string
		dbInsertV2ResReturnRows   [][]driver.Value
		expectedSuccess           int
		expectedFailure           int
		expectErr                 bool
	}{{
		f: sm.CompLockV2Filter{
			ID:                  []string{"x3000c0s9b0n0"},
			ProcessingModel:     sm.CLProcessingModelRigid,
			ReservationDuration: 1,
		},
		dbGetCompIDsError:         nil,
		expectedGetCompIDsPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedGetCompIDsArgs:    []driver.Value{"x3000c0s9b0n0"},
		dbGetCompIDsReturnCols:    []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbGetCompIDsReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0", "Node", "Ready", "OK", true, "", "Compute", "", 42, "", "Sling", "X86", "Mountain", false, false},
		},
		dbInsertError:           nil,
		expectedInsertPrepare:   regexp.QuoteMeta(resInsertReservation),
		expectedInsertArgs:      []driver.Value{"x3000c0s9b0n0", AnyTime{}, AnyTime{}, AnyUUID{}, AnyUUID{}},
		dbInsertV2ResReturnCols: []string{"component_id", "deputy_key", "reservation_key"},
		dbInsertV2ResReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0", "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17", "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"},
		},
		expectedSuccess: 1,
		expectedFailure: 0,
		expectErr:       false,
	}, {
		f: sm.CompLockV2Filter{
			ID:                  []string{"x3000c0s9b0n0"},
			ProcessingModel:     sm.CLProcessingModelRigid,
			ReservationDuration: 1,
		},
		dbGetCompIDsError:         nil,
		expectedGetCompIDsPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedGetCompIDsArgs:    []driver.Value{"x3000c0s9b0n0"},
		dbGetCompIDsReturnCols:    []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbGetCompIDsReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0", "Node", "Ready", "OK", true, "", "Compute", "", 42, "", "Sling", "X86", "Mountain", false, true},
		},
		dbInsertError:           nil,
		expectedInsertPrepare:   "",
		expectedInsertArgs:      []driver.Value{},
		dbInsertV2ResReturnCols: []string{},
		dbInsertV2ResReturnRows: [][]driver.Value{},
		expectedSuccess:         0,
		expectedFailure:         0,
		expectErr:               true,
	}, {
		f: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		dbGetCompIDsError:         nil,
		expectedGetCompIDsPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedGetCompIDsArgs:    []driver.Value{"x3000c0s9b0n0"},
		dbGetCompIDsReturnCols:    []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbGetCompIDsReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0", "Node", "Ready", "OK", true, "", "Compute", "", 42, "", "Sling", "X86", "Mountain", false, false},
		},
		dbInsertError:           nil,
		expectedInsertPrepare:   "",
		expectedInsertArgs:      []driver.Value{},
		dbInsertV2ResReturnCols: []string{},
		dbInsertV2ResReturnRows: [][]driver.Value{},
		expectedSuccess:         0,
		expectedFailure:         0,
		expectErr:               true,
	}, {
		f: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		dbGetCompIDsError:         nil,
		expectedGetCompIDsPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedGetCompIDsArgs:    []driver.Value{"x3000c0s9b0n0"},
		dbGetCompIDsReturnCols:    []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbGetCompIDsReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0", "Node", "Ready", "OK", true, "", "Compute", "", 42, "", "Sling", "X86", "Mountain", true, false},
		},
		dbInsertError:           nil,
		expectedInsertPrepare:   "",
		expectedInsertArgs:      []driver.Value{},
		dbInsertV2ResReturnCols: []string{},
		dbInsertV2ResReturnRows: [][]driver.Value{},
		expectedSuccess:         0,
		expectedFailure:         0,
		expectErr:               true,
	}}

	for i, test := range tests {
		ResetMockDB()
		rows := sqlmock.NewRows(test.dbGetCompIDsReturnCols)
		for _, row := range test.dbGetCompIDsReturnRows {
			rows.AddRow(row...)
		}
		v2rows := sqlmock.NewRows(test.dbInsertV2ResReturnCols)
		for _, row := range test.dbInsertV2ResReturnRows {
			v2rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedGetCompIDsPrepare == "" && test.dbGetCompIDsError == nil {
			mockPG.ExpectRollback()
		} else if test.dbGetCompIDsError != nil {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WillReturnError(test.dbGetCompIDsError)
			mockPG.ExpectRollback()
		} else if test.expectedInsertPrepare == "" && test.dbInsertError == nil {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WithArgs(test.expectedGetCompIDsArgs...).WillReturnRows(rows)
			mockPG.ExpectRollback()
		} else if test.dbInsertError != nil {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WithArgs(test.expectedGetCompIDsArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedInsertPrepare).ExpectQuery().WillReturnError(test.dbInsertError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WithArgs(test.expectedGetCompIDsArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedInsertPrepare).ExpectQuery().WithArgs(test.expectedInsertArgs...).WillReturnRows(v2rows)
			mockPG.ExpectCommit()
		}

		results, err := dPG.InsertCompReservations(test.f)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if !test.expectErr {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else {
				if len(results.Success) != test.expectedSuccess {
					t.Errorf("Test %v Failed: Expected %v Successes. Got %v", i, test.expectedSuccess, len(results.Success))
				} else if len(results.Failure) != test.expectedFailure {
					t.Errorf("Test %v Failed: Expected %v Failures. Got %v", i, test.expectedFailure, len(results.Failure))
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgDeleteCompReservationsForce(t *testing.T) {
	res := compReservation{
		component_id: "x3000c0s9b0n0",
		create_timestamp: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		expiration_timestamp: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		deputy_key:      "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17",
		reservation_key: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
	}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note we only use the query here so the args values don't really matter.
	resDeleteReservation, _, _ := sqq.Delete(compResTable).
		Where(sq.Eq{compResCompIdCol: []string{res.component_id}}).
		Suffix("RETURNING " + compResCompIdCol).ToSql()

	tests := []struct {
		f                         sm.CompLockV2Filter
		dbGetCompIDsError         error
		expectedGetCompIDsPrepare string
		expectedGetCompIDsArgs    []driver.Value
		dbGetCompIDsReturnCols    []string
		dbGetCompIDsReturnRows    [][]driver.Value
		dbDeleteError             error
		expectedDeletePrepare     string
		expectedDeleteArgs        []driver.Value
		dbDeleteReturnCols        []string
		dbDeleteReturnRows        [][]driver.Value
		expectedSuccess           int
		expectedFailure           int
		expectErr                 bool
	}{{
		f: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		dbGetCompIDsError:         nil,
		expectedGetCompIDsPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedGetCompIDsArgs:    []driver.Value{"x3000c0s9b0n0"},
		dbGetCompIDsReturnCols:    []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbGetCompIDsReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0", "Node", "Ready", "OK", true, "", "Compute", "", 42, "", "Sling", "X86", "Mountain", false, false},
		},
		dbDeleteError:         nil,
		expectedDeletePrepare: regexp.QuoteMeta(resDeleteReservation),
		expectedDeleteArgs:    []driver.Value{"x3000c0s9b0n0"},
		dbDeleteReturnCols:    []string{"component_id"},
		dbDeleteReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0"},
		},
		expectedSuccess: 1,
		expectedFailure: 0,
		expectErr:       false,
	}, {
		f: sm.CompLockV2Filter{
			ID:              []string{"x3000c0s9b0n0"},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		dbGetCompIDsError:         sql.ErrNoRows,
		expectedGetCompIDsPrepare: regexp.QuoteMeta(tGetCompBaseQuery + " WHERE c.id IN ($1)"),
		expectedGetCompIDsArgs:    []driver.Value{"x3000c0s9b0n0"},
		dbGetCompIDsReturnCols:    []string{"id", "type", "state", "flag", "enabled", "admin", "role", "subrole", "nid", "subtype", "nettype", "arch", "class", "reservation_disabled", "locked"},
		dbGetCompIDsReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0", "Node", "Ready", "OK", true, "", "Compute", "", 42, "", "Sling", "X86", "Mountain", false, false},
		},
		dbDeleteError:         nil,
		expectedDeletePrepare: "",
		expectedDeleteArgs:    []driver.Value{},
		dbDeleteReturnCols:    []string{},
		dbDeleteReturnRows:    [][]driver.Value{},
		expectedSuccess:       0,
		expectedFailure:       0,
		expectErr:             true,
	}}

	for i, test := range tests {
		ResetMockDB()
		rows := sqlmock.NewRows(test.dbGetCompIDsReturnCols)
		for _, row := range test.dbGetCompIDsReturnRows {
			rows.AddRow(row...)
		}
		drows := sqlmock.NewRows(test.dbDeleteReturnCols)
		for _, row := range test.dbDeleteReturnRows {
			drows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.expectedGetCompIDsPrepare == "" && test.dbGetCompIDsError == nil {
			mockPG.ExpectRollback()
		} else if test.dbGetCompIDsError != nil {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WillReturnError(test.dbGetCompIDsError)
			mockPG.ExpectRollback()
		} else if test.expectedDeletePrepare == "" && test.dbDeleteError == nil {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WithArgs(test.expectedGetCompIDsArgs...).WillReturnRows(rows)
			mockPG.ExpectRollback()
		} else if test.dbDeleteError != nil {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WithArgs(test.expectedGetCompIDsArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedDeletePrepare).ExpectQuery().WillReturnError(test.dbDeleteError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedGetCompIDsPrepare).ExpectQuery().WithArgs(test.expectedGetCompIDsArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedDeletePrepare).ExpectQuery().WithArgs(test.expectedDeleteArgs...).WillReturnRows(drows)
			mockPG.ExpectCommit()
		}

		results, err := dPG.DeleteCompReservationsForce(test.f)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if !test.expectErr {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else {
				if len(results.Success.ComponentIDs) != test.expectedSuccess {
					t.Errorf("Test %v Failed: Expected %v Successes. Got %v", i, test.expectedSuccess, len(results.Success.ComponentIDs))
				} else if len(results.Failure) != test.expectedFailure {
					t.Errorf("Test %v Failed: Expected %v Failures. Got %v", i, test.expectedFailure, len(results.Failure))
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgDeleteCompReservations(t *testing.T) {
	res := compReservation{
		component_id: "x3000c0s9b0n0",
		create_timestamp: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		expiration_timestamp: sql.NullTime{
			Valid: true,
			Time:  time.Now(),
		},
		deputy_key:      "x3000c0s9b0n0:dk:de1a20c2-efc9-41ad-b839-1e3cef197d17",
		reservation_key: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
	}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note we only use the query here so the args values don't really matter.
	resDeleteReservation, _, _ := sqq.Delete(compResTable).
		Where(sq.Eq{compResRKCol: []string{res.reservation_key}}).
		Suffix("RETURNING " + compResCompIdCol).ToSql()

	tests := []struct {
		f                     sm.CompLockV2ReservationFilter
		dbDeleteError         error
		expectedDeletePrepare string
		expectedDeleteArgs    []driver.Value
		dbDeleteReturnCols    []string
		dbDeleteReturnRows    [][]driver.Value
		expectedSuccess       int
		expectedFailure       int
		expectErr             bool
	}{{
		f: sm.CompLockV2ReservationFilter{
			ReservationKeys: []sm.CompLockV2Key{
				sm.CompLockV2Key{
					ID:  "x3000c0s9b0n0",
					Key: "x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d",
				},
			},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		dbDeleteError:         nil,
		expectedDeletePrepare: regexp.QuoteMeta(resDeleteReservation),
		expectedDeleteArgs:    []driver.Value{"x3000c0s9b0n0:rk:cbff2077-952f-4536-a102-c442227fdc5d"},
		dbDeleteReturnCols:    []string{"component_id"},
		dbDeleteReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0"},
		},
		expectedSuccess: 1,
		expectedFailure: 0,
		expectErr:       false,
	}, {
		f: sm.CompLockV2ReservationFilter{
			ReservationKeys: []sm.CompLockV2Key{
				sm.CompLockV2Key{
					ID: "x3000c0s9b0n0",
				},
			},
			ProcessingModel: sm.CLProcessingModelRigid,
		},
		dbDeleteError:         nil,
		expectedDeletePrepare: "",
		expectedDeleteArgs:    []driver.Value{},
		dbDeleteReturnCols:    []string{},
		dbDeleteReturnRows:    [][]driver.Value{},
		expectedSuccess:       0,
		expectedFailure:       0,
		expectErr:             true,
	}}

	for i, test := range tests {
		ResetMockDB()
		rows := sqlmock.NewRows(test.dbDeleteReturnCols)
		for _, row := range test.dbDeleteReturnRows {
			rows.AddRow(row...)
		}

		v1rows := sqlmock.NewRows([]string{"component_id"})
		v1rows.AddRow([]driver.Value{"x3000c0s9b0n0"}...)

		compRows := sqlmock.NewRows([]string{"id"})
		compRows.AddRow([]driver.Value{"x3000c0s9b0n0"}...)

		mockPG.ExpectBegin()
		if test.expectedDeletePrepare == "" && test.dbDeleteError == nil {
			mockPG.ExpectRollback()
		} else if test.dbDeleteError != nil {
			mockPG.ExpectPrepare(test.expectedDeletePrepare).ExpectQuery().WillReturnError(test.dbDeleteError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedDeletePrepare).ExpectQuery().WithArgs(test.expectedDeleteArgs...).WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		results, err := dPG.DeleteCompReservations(test.f)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if !test.expectErr {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else {
				if len(results.Success.ComponentIDs) != test.expectedSuccess {
					t.Errorf("Test %v Failed: Expected %v Successes. Got %v", i, test.expectedSuccess, len(results.Success.ComponentIDs))
				} else if len(results.Failure) != test.expectedFailure {
					t.Errorf("Test %v Failed: Expected %v Failures. Got %v", i, test.expectedFailure, len(results.Failure))
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgDeleteCompReservationsExpired(t *testing.T) {

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note we only use the query here so the args values don't really matter.
	resDeleteReservation, _, _ := sqq.Delete(compResTable).
		Where(compResExpireCol + " IS NOT NULL AND NOW() >= " + compResExpireCol).
		Suffix("RETURNING " + compResCompIdCol).ToSql()

	tests := []struct {
		dbDeleteError         error
		expectedDeletePrepare string
		dbDeleteReturnCols    []string
		dbDeleteReturnRows    [][]driver.Value
		expectedSuccess       int
		expectErr             bool
	}{{
		dbDeleteError:         nil,
		expectedDeletePrepare: regexp.QuoteMeta(resDeleteReservation),
		dbDeleteReturnCols:    []string{"component_id"},
		dbDeleteReturnRows: [][]driver.Value{
			[]driver.Value{"x3000c0s9b0n0"},
		},
		expectedSuccess: 1,
		expectErr:       false,
	}}

	for i, test := range tests {
		ResetMockDB()
		rows := sqlmock.NewRows(test.dbDeleteReturnCols)
		for _, row := range test.dbDeleteReturnRows {
			rows.AddRow(row...)
		}

		v1rows := sqlmock.NewRows([]string{"component_id"})
		v1rows.AddRow([]driver.Value{"x3000c0s9b0n0"}...)

		compRows := sqlmock.NewRows([]string{"id"})
		compRows.AddRow([]driver.Value{"x3000c0s9b0n0"}...)

		mockPG.ExpectBegin()
		if test.expectedDeletePrepare == "" && test.dbDeleteError == nil {
			mockPG.ExpectRollback()
		} else if test.dbDeleteError != nil {
			mockPG.ExpectPrepare(test.expectedDeletePrepare).ExpectQuery().WillReturnError(test.dbDeleteError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedDeletePrepare).ExpectQuery().WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		results, err := dPG.DeleteCompReservationsExpired()
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if !test.expectErr {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else {
				if len(results) != test.expectedSuccess {
					t.Errorf("Test %v Failed: Expected %v Successes. Got %v", i, test.expectedSuccess, len(results))
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// Power Map Query Tests
///////////////////////////////////////////////////////////////////////////////

func TestPgGetPowerMapByID(t *testing.T) {
	columns := []string{"id", "powered_by"}

	tests := []struct {
		id               string
		dbRows           [][]driver.Value
		dbError          error
		expectedPrepare  string
		expectedArgs     []driver.Value
		expectedPowerMap *sm.PowerMap
		expectedErr      error
	}{{
		id: "x0c0s21b0",
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", pq.Array([]string{"x0m0p0j10", "x0m0p0j11"})},
		},
		dbError:          nil,
		expectedPrepare:  regexp.QuoteMeta(getPowerMapByIDQuery),
		expectedArgs:     []driver.Value{"x0c0s21b0"},
		expectedPowerMap: &sm.PowerMap{ID: "x0c0s21b0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
		expectedErr:      nil,
	}, {
		id:               "",
		dbRows:           nil,
		dbError:          nil,
		expectedPrepare:  "",
		expectedArgs:     nil,
		expectedPowerMap: nil,
		expectedErr:      ErrHMSDSArgMissing,
	}, {
		id:               "x0c0s21b0",
		dbRows:           nil,
		dbError:          sql.ErrNoRows,
		expectedPrepare:  regexp.QuoteMeta(getPowerMapByIDQuery),
		expectedArgs:     []driver.Value{"x0c0s21b0"},
		expectedPowerMap: nil,
		expectedErr:      nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectQuery().WithArgs(test.expectedArgs...).WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		m, err := dPG.GetPowerMapByID(test.id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedPowerMap, m) {
				t.Errorf("Test %v Failed: Expected power map '%v'; Recieved power map '%v'", i, test.expectedPowerMap, m)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetPowerMapsAll(t *testing.T) {
	columns := []string{"id", "powered_by"}

	tests := []struct {
		dbRows          [][]driver.Value
		dbError         error
		expectedPrepare string
		expectedPMs     []*sm.PowerMap
	}{{
		dbRows: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", pq.Array([]string{"x0m0p0j10", "x0m0p0j11"})},
			[]driver.Value{"x0c0s22b0", pq.Array([]string{"x0m0p0j12", "x0m0p0j13"})},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(getPowerMapsAllQuery),
		expectedPMs: []*sm.PowerMap{
			&sm.PowerMap{ID: "x0c0s21b0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
			&sm.PowerMap{ID: "x0c0s22b0", PoweredBy: []string{"x0m0p0j12", "x0m0p0j13"}},
		},
	}, {
		dbRows:          nil,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(getPowerMapsAllQuery),
		expectedPMs:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(columns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}

		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectQuery().WillReturnRows(rows)
			mockPG.ExpectCommit()
		}

		ms, err := dPG.GetPowerMapsAll()
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if !reflect.DeepEqual(test.expectedPMs, ms) {
				t.Errorf("Test %v Failed: Expected power maps '%v'; Recieved power maps '%v'", i, test.expectedPMs, ms)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgInsertPowerMap(t *testing.T) {
	tests := []struct {
		m               *sm.PowerMap
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedErr     error
	}{{
		m:               &sm.PowerMap{ID: "x0c0s21b0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(insertPgPowerMapQuery),
		expectedArgs:    []driver.Value{"x0c0s21b0", pq.Array([]string{"x0m0p0j10", "x0m0p0j11"})},
		expectedErr:     nil,
	}, {
		m:               &sm.PowerMap{ID: "x0c0s21b0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(insertPgPowerMapQuery),
		expectedArgs:    []driver.Value{"x0c0s21b0", pq.Array([]string{"x0m0p0j10", "x0m0p0j11"})},
		expectedErr:     nil,
	}, {
		m:               nil,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedErr:     ErrHMSDSArgNil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		err := dPG.InsertPowerMap(test.m)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgInsertPowerMaps(t *testing.T) {
	tests := []struct {
		ms              []sm.PowerMap
		dbError         error
		expectedPrepare string
		expectedArgs    [][]driver.Value
		expectedErr     error
	}{{
		ms: []sm.PowerMap{
			sm.PowerMap{ID: "x0c0s21b0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
			sm.PowerMap{ID: "x0c0s22b0", PoweredBy: []string{"x0m0p0j12", "x0m0p0j13"}},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(insertPgPowerMapQuery),
		expectedArgs: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", pq.Array([]string{"x0m0p0j10", "x0m0p0j11"})},
			[]driver.Value{"x0c0s22b0", pq.Array([]string{"x0m0p0j12", "x0m0p0j13"})},
		},
		expectedErr: nil,
	}, {
		ms: []sm.PowerMap{
			sm.PowerMap{ID: "x0c0s21b0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
		},
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(insertPgPowerMapQuery),
		expectedArgs: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", pq.Array([]string{"x0m0p0j10", "x0m0p0j11"})},
		},
		expectedErr: nil,
	}, {
		ms: []sm.PowerMap{
			sm.PowerMap{ID: "x0c0s21b0", PoweredBy: []string{"x0m0p0j10", "x0m0p0j11"}},
		},
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(insertPgPowerMapQuery),
		expectedArgs: [][]driver.Value{
			[]driver.Value{"x0c0s21b0", pq.Array([]string{"x0m0p0j10", "x0m0p0j11"})},
		},
		expectedErr: nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			for j, args := range test.expectedArgs {
				if j == 0 {
					mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WithArgs(args...).WillReturnResult(sqlmock.NewResult(0, 1))
				} else {
					mockPG.ExpectExec(ToPGQueryArgs(test.expectedPrepare)).WithArgs(args...).WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}
			mockPG.ExpectCommit()
		}

		err := dPG.InsertPowerMaps(test.ms)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgDeletePowerMapByID(t *testing.T) {
	tests := []struct {
		id              string
		dbResult        int64
		dbError         error
		expectedPrepare string
		expectedArgs    []driver.Value
		expectedResult  bool
		expectedErr     error
	}{{
		id:              "x0c0s21b0",
		dbResult:        1,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(deletePowerMapByIDQuery),
		expectedArgs:    []driver.Value{"x0c0s21b0"},
		expectedResult:  true,
		expectedErr:     nil,
	}, {
		id:              "",
		dbResult:        0,
		dbError:         nil,
		expectedPrepare: "",
		expectedArgs:    nil,
		expectedResult:  false,
		expectedErr:     ErrHMSDSArgNil,
	}, {
		id:              "x0c0s21b0",
		dbResult:        0,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(deletePowerMapByIDQuery),
		expectedArgs:    []driver.Value{"x0c0s21b0"},
		expectedResult:  false,
		expectedErr:     nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else if test.expectedErr != nil {
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(ToPGQueryArgs(test.expectedPrepare)).ExpectExec().WithArgs(test.expectedArgs...).WillReturnResult(sqlmock.NewResult(0, test.dbResult))
			mockPG.ExpectCommit()
		}

		didDelete, err := dPG.DeletePowerMapByID(test.id)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil && test.expectedErr == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedResult != didDelete {
				t.Errorf("Test %v Failed: Expected didDelete '%v'; Recieved didDelete '%v'", i, test.expectedResult, didDelete)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgDeletePowerMapsAll(t *testing.T) {
	tests := []struct {
		dbResult        int64
		dbError         error
		expectedPrepare string
		expectedResult  int64
		expectedErr     error
	}{{
		dbResult:        1,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(deletePowerMapsAllQuery),
		expectedResult:  1,
	}, {
		dbResult:        0,
		dbError:         nil,
		expectedPrepare: regexp.QuoteMeta(deletePowerMapsAllQuery),
		expectedResult:  0,
	}, {
		dbResult:        0,
		dbError:         sql.ErrNoRows,
		expectedPrepare: regexp.QuoteMeta(deletePowerMapsAllQuery),
		expectedResult:  0,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		mockPG.ExpectBegin()
		if test.dbError != nil {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnError(test.dbError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedPrepare).ExpectExec().WillReturnResult(sqlmock.NewResult(0, test.dbResult))
			mockPG.ExpectCommit()
		}

		numDelete, err := dPG.DeletePowerMapsAll()
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else if test.expectedResult != numDelete {
				t.Errorf("Test %v Failed: Expected numDelete '%v'; Recieved numDelete '%v'", i, test.expectedResult, numDelete)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

////////////////////////////////////////////////////////////////////////////
//
// Job Sync Management
//
////////////////////////////////////////////////////////////////////////////

func TestPgInsertJob(t *testing.T) {
	var dJob1 = &sm.Job{
		JobData: sm.JobData{
			Id:       uuid.New().String(),
			Type:     sm.JobTypeSRFP,
			Status:   sm.JobNotStarted,
			Lifetime: 30,
		},
		Data: &sm.SrfpJobData{
			CompId: "x0c0s0b0n0",
		},
	}
	var dJob2 = &sm.Job{
		JobData: sm.JobData{
			Id:       uuid.New().String(),
			Type:     sm.JobTypeSRFP,
			Status:   sm.JobNotStarted,
			Lifetime: 30,
		},
	}

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	// Note we only use the query here so the args values don't really matter.
	dJob1InsertJob, _, _ := sqq.Insert(jobTable).
		Columns(jobColsNoTS...).
		Values(sq.Expr("?", dJob1.Id), dJob1.Type, dJob1.Status, dJob1.Lifetime).ToSql()

	dJob1InsertStateRFPollJob, _, _ := sqq.Insert(stateRfPollTable).
		Columns(stateRfPollCols...).
		Values(dJob1.Data.(*sm.SrfpJobData).CompId, sq.Expr("?", dJob1.Id)).ToSql()

	tests := []struct {
		j                       *sm.Job
		dbInsertJError          error
		expectedInsertJPrepare  string
		expectedInsertJArgs     []driver.Value
		dbInsertJDError         error
		expectedInsertJDPrepare string
		expectedInsertJDArgs    []driver.Value
		expectErr               bool
	}{{
		j:                       dJob1,
		dbInsertJError:          nil,
		expectedInsertJPrepare:  regexp.QuoteMeta(dJob1InsertJob),
		expectedInsertJArgs:     []driver.Value{AnyUUID{}, dJob1.Type, dJob1.Status, dJob1.Lifetime},
		dbInsertJDError:         nil,
		expectedInsertJDPrepare: regexp.QuoteMeta(dJob1InsertStateRFPollJob),
		expectedInsertJDArgs:    []driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId, AnyUUID{}},
		expectErr:               false,
	}, {
		j:                       dJob1,
		dbInsertJError:          nil,
		expectedInsertJPrepare:  regexp.QuoteMeta(dJob1InsertJob),
		expectedInsertJArgs:     []driver.Value{AnyUUID{}, dJob1.Type, dJob1.Status, dJob1.Lifetime},
		dbInsertJDError:         ErrHMSDSArgBadID,
		expectedInsertJDPrepare: regexp.QuoteMeta(dJob1InsertStateRFPollJob),
		expectedInsertJDArgs:    []driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId, AnyUUID{}},
		expectErr:               true,
	}, {
		j:                       dJob1,
		dbInsertJError:          ErrHMSDSArgBadID,
		expectedInsertJPrepare:  regexp.QuoteMeta(dJob1InsertJob),
		expectedInsertJArgs:     []driver.Value{AnyUUID{}, dJob1.Type, dJob1.Status, dJob1.Lifetime},
		dbInsertJDError:         nil,
		expectedInsertJDPrepare: "",
		expectedInsertJDArgs:    []driver.Value{},
		expectErr:               true,
	}, {
		j:                       dJob2,
		dbInsertJError:          nil,
		expectedInsertJPrepare:  regexp.QuoteMeta(dJob1InsertJob),
		expectedInsertJArgs:     []driver.Value{AnyUUID{}, dJob2.Type, dJob2.Status, dJob2.Lifetime},
		dbInsertJDError:         ErrHMSDSArgBadID,
		expectedInsertJDPrepare: "",
		expectedInsertJDArgs:    []driver.Value{},
		expectErr:               true,
	}}

	for i, test := range tests {
		ResetMockDB()

		mockPG.ExpectBegin()
		if test.expectedInsertJPrepare == "" && test.dbInsertJError != nil {
			mockPG.ExpectRollback()
		} else if test.dbInsertJError != nil {
			mockPG.ExpectPrepare(test.expectedInsertJPrepare).ExpectExec().WillReturnError(test.dbInsertJError)
			mockPG.ExpectRollback()
		} else if test.expectedInsertJDPrepare == "" && test.dbInsertJDError != nil {
			mockPG.ExpectPrepare(test.expectedInsertJPrepare).ExpectExec().WithArgs(test.expectedInsertJArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectRollback()
		} else if test.dbInsertJDError != nil {
			mockPG.ExpectPrepare(test.expectedInsertJPrepare).ExpectExec().WithArgs(test.expectedInsertJArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectPrepare(test.expectedInsertJDPrepare).ExpectExec().WillReturnError(test.dbInsertJDError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedInsertJPrepare).ExpectExec().WithArgs(test.expectedInsertJArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectPrepare(test.expectedInsertJDPrepare).ExpectExec().WithArgs(test.expectedInsertJDArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		jobId, err := dPG.InsertJob(test.j)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if !test.expectErr {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			} else {
				if jobId == "" {
					t.Errorf("Test %v Failed: Expected a new job id", i)
				}
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgUpdateJob(t *testing.T) {
	id := uuid.New().String()
	status := sm.JobInProgress

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dJob1Update1, _, _ := sqq.Update("").
		Table(jobTable).
		Where("id = ?", id).
		Set(jobStatusCol, status).
		Set(jobLastUpdateCol, "NOW()").ToSql()

	dJob1Update2, _, _ := sqq.Update("").
		Table(jobTable).
		Where("id = ?", id).
		Set(jobLastUpdateCol, "NOW()").ToSql()

	tests := []struct {
		id                    string
		status                string
		dbUpdateError         error
		expectedUpdatePrepare string
		expectedUpdateArgs    []driver.Value
	}{{
		id:                    id,
		status:                status,
		dbUpdateError:         nil,
		expectedUpdatePrepare: regexp.QuoteMeta(dJob1Update1),
		expectedUpdateArgs:    []driver.Value{status, "NOW()", id},
	}, {
		id:                    id,
		status:                "",
		dbUpdateError:         nil,
		expectedUpdatePrepare: regexp.QuoteMeta(dJob1Update2),
		expectedUpdateArgs:    []driver.Value{"NOW()", id},
	}, {
		id:                    id,
		status:                "",
		dbUpdateError:         ErrHMSDSArgBadID,
		expectedUpdatePrepare: regexp.QuoteMeta(dJob1Update2),
		expectedUpdateArgs:    []driver.Value{"NOW()", id},
	}, {
		id:                    "",
		status:                "",
		dbUpdateError:         ErrHMSDSArgBadID,
		expectedUpdatePrepare: "",
		expectedUpdateArgs:    []driver.Value{},
	}}

	for i, test := range tests {
		ResetMockDB()
		mockPG.ExpectBegin()
		if test.expectedUpdatePrepare == "" && test.dbUpdateError != nil {
			mockPG.ExpectRollback()
		} else if test.dbUpdateError != nil {
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WillReturnError(test.dbUpdateError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedUpdatePrepare).ExpectExec().WithArgs(test.expectedUpdateArgs...).WillReturnResult(sqlmock.NewResult(0, 1))
			mockPG.ExpectCommit()
		}

		_, err := dPG.UpdateJob(test.id, test.status)
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s",
				i, mock_err)
		}
		if err != nil && test.dbUpdateError == nil {
			t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
		} else if err == nil && test.dbUpdateError != nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetJob(t *testing.T) {
	var dJob1 = &sm.Job{
		JobData: sm.JobData{
			Id:         uuid.New().String(),
			Type:       sm.JobTypeSRFP,
			Status:     sm.JobNotStarted,
			LastUpdate: "",
			Lifetime:   30,
		},
		Data: &sm.SrfpJobData{
			CompId: "x0c0s0b0n0",
		},
	}

	columns := jobCols

	dataCols := stateRfPollColsId

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dJob1Query1, _, _ := sqq.Select(jobCols...).
		From(jobTable).
		Where("id = ?", dJob1.Id).ToSql()

	dJob1JDQuery1, _, _ := sqq.Select(stateRfPollColsId...).
		From(stateRfPollTable).
		Where("job_id = ?", dJob1.Id).ToSql()

	tests := []struct {
		jobId                  string
		dbColumns              []string
		dbRows                 [][]driver.Value
		dbQueryError           error
		expectedQueryPrepare   string
		expectedQueryArgs      []driver.Value
		dbJDColumns            []string
		dbJDRows               [][]driver.Value
		dbJDQueryError         error
		expectedJDQueryPrepare string
		expectedJDQueryArgs    []driver.Value
		expectedJ              *sm.Job
	}{{
		jobId:     dJob1.Id,
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         nil,
		expectedQueryPrepare: regexp.QuoteMeta(dJob1Query1),
		expectedQueryArgs:    []driver.Value{dJob1.Id},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         nil,
		expectedJDQueryPrepare: regexp.QuoteMeta(dJob1JDQuery1),
		expectedJDQueryArgs:    []driver.Value{dJob1.Id},
		expectedJ:              dJob1,
	}, {
		jobId:     dJob1.Id,
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         nil,
		expectedQueryPrepare: regexp.QuoteMeta(dJob1Query1),
		expectedQueryArgs:    []driver.Value{dJob1.Id},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         ErrHMSDSArgBadID,
		expectedJDQueryPrepare: regexp.QuoteMeta(dJob1JDQuery1),
		expectedJDQueryArgs:    []driver.Value{dJob1.Id},
		expectedJ:              nil,
	}, {
		jobId:     dJob1.Id,
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         ErrHMSDSArgBadID,
		expectedQueryPrepare: regexp.QuoteMeta(dJob1Query1),
		expectedQueryArgs:    []driver.Value{dJob1.Id},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         nil,
		expectedJDQueryPrepare: "",
		expectedJDQueryArgs:    []driver.Value{},
		expectedJ:              nil,
	}, {
		jobId:     "",
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         ErrHMSDSArgEmpty,
		expectedQueryPrepare: "",
		expectedQueryArgs:    []driver.Value{},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         nil,
		expectedJDQueryPrepare: "",
		expectedJDQueryArgs:    []driver.Value{},
		expectedJ:              nil,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}
		jdrows := sqlmock.NewRows(test.dbJDColumns)
		for _, jdrow := range test.dbJDRows {
			jdrows.AddRow(jdrow...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbJDQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedJDQueryPrepare).ExpectQuery().WillReturnError(test.dbJDQueryError)
			mockPG.ExpectRollback()
		} else {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			mockPG.ExpectPrepare(test.expectedJDQueryPrepare).ExpectQuery().WithArgs(test.expectedJDQueryArgs...).WillReturnRows(jdrows)
			mockPG.ExpectCommit()
		}

		j, err := dPG.GetJob(test.jobId)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbQueryError == nil && test.dbJDQueryError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if j == nil {
				t.Errorf("Test %v Failed: Expected non-nil Job", i)
			} else if !reflect.DeepEqual(j, test.expectedJ) {
				t.Errorf("Test %v Failed: Expected Job %v (got %v)",
					i, j, test.expectedJ)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}

func TestPgGetJobs(t *testing.T) {
	var dJob1 = &sm.Job{
		JobData: sm.JobData{
			Id:         uuid.New().String(),
			Type:       sm.JobTypeSRFP,
			Status:     sm.JobNotStarted,
			LastUpdate: "",
			Lifetime:   30,
		},
		Data: &sm.SrfpJobData{
			CompId: "x0c0s0b0n0",
		},
	}

	columns := addAliasToCols(jobAlias, jobCols, jobCols)

	dataCols := stateRfPollColsId

	sqq := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dJob1Query1, _, _ := sqq.Select(columns...).
		From(jobTable + " " + jobAlias).
		Where(sq.Eq{jobIdColAlias: []string{dJob1.Id}}).ToSql()

	dJob1JDQuery1, _, _ := sqq.Select(stateRfPollColsId...).
		From(stateRfPollTable).
		Where("job_id = ?", dJob1.Id).ToSql()

	dJob1Query2, _, _ := sqq.Select(columns...).
		From(jobTable + " " + jobAlias).
		Where(sq.Eq{jobTypeColAlias: []string{dJob1.Type}}).ToSql()

	dJob1JDQuery2, _, _ := sqq.Select(stateRfPollColsId...).
		From(stateRfPollTable).
		Where("job_id = ?", dJob1.Id).ToSql()

	dJob1Query3, _, _ := sqq.Select(columns...).
		From(jobTable + " " + jobAlias).
		Where(sq.Eq{jobStatusColAlias: []string{dJob1.Status}}).ToSql()

	dJob1JDQuery3, _, _ := sqq.Select(stateRfPollColsId...).
		From(stateRfPollTable).
		Where("job_id = ?", dJob1.Id).ToSql()

	dJob1Query4, _, _ := sqq.Select(columns...).
		From(jobTable + " " + jobAlias).
		Where("NOW()-" + jobLastUpdateColAlias +
			" >= (" + jobLifetimeColAlias + " * '1 sec'::interval)").ToSql()

	dJob1JDQuery4, _, _ := sqq.Select(stateRfPollColsId...).
		From(stateRfPollTable).
		Where("job_id = ?", dJob1.Id).ToSql()

	tests := []struct {
		f_opts                 []JobSyncFiltFunc
		dbColumns              []string
		dbRows                 [][]driver.Value
		dbQueryError           error
		expectedQueryPrepare   string
		expectedQueryArgs      []driver.Value
		dbJDColumns            []string
		dbJDRows               [][]driver.Value
		dbJDQueryError         error
		expectedJDQueryPrepare string
		expectedJDQueryArgs    []driver.Value
		expectedJ              *sm.Job
	}{{
		f_opts:    []JobSyncFiltFunc{JS_ID(dJob1.Id)},
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         nil,
		expectedQueryPrepare: regexp.QuoteMeta(dJob1Query1),
		expectedQueryArgs:    []driver.Value{dJob1.Id},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         nil,
		expectedJDQueryPrepare: regexp.QuoteMeta(dJob1JDQuery1),
		expectedJDQueryArgs:    []driver.Value{dJob1.Id},
		expectedJ:              dJob1,
	}, {
		f_opts:    []JobSyncFiltFunc{JS_Type(dJob1.Type)},
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         nil,
		expectedQueryPrepare: regexp.QuoteMeta(dJob1Query2),
		expectedQueryArgs:    []driver.Value{dJob1.Type},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         nil,
		expectedJDQueryPrepare: regexp.QuoteMeta(dJob1JDQuery2),
		expectedJDQueryArgs:    []driver.Value{dJob1.Id},
		expectedJ:              dJob1,
	}, {
		f_opts:    []JobSyncFiltFunc{JS_Status(dJob1.Status)},
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         nil,
		expectedQueryPrepare: regexp.QuoteMeta(dJob1Query3),
		expectedQueryArgs:    []driver.Value{dJob1.Status},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         nil,
		expectedJDQueryPrepare: regexp.QuoteMeta(dJob1JDQuery3),
		expectedJDQueryArgs:    []driver.Value{dJob1.Id},
		expectedJ:              dJob1,
	}, {
		f_opts:    []JobSyncFiltFunc{JS_Expired},
		dbColumns: columns,
		dbRows: [][]driver.Value{
			[]driver.Value{dJob1.Id, dJob1.Type, dJob1.Status, dJob1.LastUpdate, dJob1.Lifetime},
		},
		dbQueryError:         nil,
		expectedQueryPrepare: regexp.QuoteMeta(dJob1Query4),
		expectedQueryArgs:    []driver.Value{},
		dbJDColumns:          dataCols,
		dbJDRows: [][]driver.Value{
			[]driver.Value{dJob1.Data.(*sm.SrfpJobData).CompId},
		},
		dbJDQueryError:         nil,
		expectedJDQueryPrepare: regexp.QuoteMeta(dJob1JDQuery4),
		expectedJDQueryArgs:    []driver.Value{dJob1.Id},
		expectedJ:              dJob1,
	}}

	for i, test := range tests {
		ResetMockDB()
		// before we actually execute our api function, we need to expect required DB actions
		rows := sqlmock.NewRows(test.dbColumns)
		for _, row := range test.dbRows {
			rows.AddRow(row...)
		}
		jdrows := sqlmock.NewRows(test.dbJDColumns)
		for _, jdrow := range test.dbJDRows {
			jdrows.AddRow(jdrow...)
		}

		mockPG.ExpectBegin()
		if test.expectedQueryPrepare == "" && test.dbQueryError != nil {
			mockPG.ExpectRollback()
		} else if test.dbQueryError != nil {
			mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnError(test.dbQueryError)
			mockPG.ExpectRollback()
		} else if test.dbJDQueryError != nil {
			if len(test.expectedQueryArgs) == 0 {
				mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			}
			mockPG.ExpectPrepare(test.expectedJDQueryPrepare).ExpectQuery().WillReturnError(test.dbJDQueryError)
			mockPG.ExpectRollback()
		} else {
			if len(test.expectedQueryArgs) == 0 {
				mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WillReturnRows(rows)
			} else {
				mockPG.ExpectPrepare(test.expectedQueryPrepare).ExpectQuery().WithArgs(test.expectedQueryArgs...).WillReturnRows(rows)
			}
			mockPG.ExpectPrepare(test.expectedJDQueryPrepare).ExpectQuery().WithArgs(test.expectedJDQueryArgs...).WillReturnRows(jdrows)
			mockPG.ExpectCommit()
		}

		js, err := dPG.GetJobs(test.f_opts...)
		// ensure all expectations have been met
		if mock_err := mockPG.ExpectationsWereMet(); mock_err != nil {
			t.Errorf("Test %v Failed: Sql expectations were not met: %s", i, mock_err)
		}
		if test.dbQueryError == nil && test.dbJDQueryError == nil {
			if err != nil {
				t.Errorf("Test %v Failed: Unexpected error received: %s", i, err)
			}
			if len(js) == 0 {
				t.Errorf("Test %v Failed: Expected non-empty list of Jobs", i)
			} else if !reflect.DeepEqual(js[0], test.expectedJ) {
				t.Errorf("Test %v Failed: Expected Job %v (got %v)",
					i, js[0], test.expectedJ)
			}
		} else if err == nil {
			t.Errorf("Test %v Failed: Expected an error.", i)
		}
	}
}
