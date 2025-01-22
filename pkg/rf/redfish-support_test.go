// MIT License
//
// (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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
package rf

import (
	"sort"
	"testing"
)

type checkMsgId struct {
	input    *EventRecord
	Registry string
	Version  string
	MsgId    string
}

// Others tested implicitly.

func TestEventRecordMsgId(t *testing.T) {
	var tests = []checkMsgId{{
		input:    &EventRecord{MessageId: " Alert.1.0.0.ShutDown  "},
		Registry: "Alert",
		Version:  "1.0.0",
		MsgId:    "ShutDown",
	}, {
		input:    &EventRecord{MessageId: "Alert.1.0.ShutDown"},
		Registry: "Alert",
		Version:  "1.0",
		MsgId:    "ShutDown",
	}, {
		input:    &EventRecord{MessageId: "Alert.v1.ShutDown"},
		Registry: "Alert",
		Version:  "v1",
		MsgId:    "ShutDown",
	}, {
		input:    &EventRecord{MessageId: "Alert.ShutDown"},
		Registry: "Alert",
		Version:  "",
		MsgId:    "ShutDown",
	}, {
		input:    &EventRecord{MessageId: "ShutDown"},
		Registry: "",
		Version:  "",
		MsgId:    "ShutDown",
	},
	}
	for i, test := range tests {
		reg, ver, id := EventRecordMsgId(test.input)
		if reg != test.Registry || ver != test.Version || id != test.MsgId {
			t.Errorf("Testcase %da: FAIL Got '%s/%s/%s', expected '%s/%s/%s'",
				i, reg, ver, id, test.Registry, test.Version, test.MsgId)
		} else {
			t.Logf("Testcase %da: PASS Got '%s/%s/%s'",
				i, reg, ver, id)
		}
	}
}

type checkVersionFields struct {
	input    string
	delim    string
	num      int
	output   string
	included int
}

func TestVersionFields(t *testing.T) {
	var tests = []checkVersionFields{{
		input:    "1.0.0",
		delim:    ".",
		num:      1,
		output:   "1",
		included: 1,
	}, {
		input:    "1.0.0",
		delim:    ".",
		num:      1,
		output:   "1",
		included: 1,
	}, {
		input:    "1.0.0",
		delim:    ".",
		num:      2,
		output:   "1.0",
		included: 2,
	}, {
		input:    "1.0.0",
		delim:    ".",
		num:      3,
		output:   "1.0.0",
		included: 3,
	}, {
		input:    "1.0.0",
		delim:    ".",
		num:      4,
		output:   "1.0.0",
		included: 3,
	}, {
		input:    "100",
		delim:    ".",
		num:      2,
		output:   "100",
		included: 1,
	},
	}
	for i, test := range tests {
		output, included := VersionFields(test.input, test.delim, test.num)
		if output != test.output || included != test.included {
			t.Errorf("Testcase %da: FAIL Got '%s': %d, expected '%s': %d",
				i, output, included, test.output, test.included)
		} else {
			t.Logf("Testcase %da: PASS Got '%s': %d", i, output, included)
		}
	}
}

func TestSort(t *testing.T) {
	var ids = []ResourceID{{
		Oid: "OUTLET24",
	}, {
		Oid: "OUTLET2",
	}, {
		Oid: "OUTLET1",
	}, {
		Oid: "OUTLET11",
	}, {
		Oid: "OUTLET5",
	},
	}
	sort.Sort(ResourceIDSlice(ids))
	if (ids[0].Oid == "OUTLET1") && (ids[1].Oid == "OUTLET2") && (ids[2].Oid == "OUTLET5") && (ids[3].Oid == "OUTLET11") && (ids[4].Oid == "OUTLET24") {
		t.Logf("SORT PASS Got '%s, %s, %s, %s, %s", ids[0], ids[1], ids[2], ids[3], ids[4])
	} else {
		t.Errorf("SORT FAIL Got '%s, %s, %s, %s, %s", ids[0], ids[1], ids[2], ids[3], ids[4])
	}
	var ids2 = []ResourceID{{
		Oid: "OUTLET24",
	}, {
		Oid: "MACHINE1",
	}, {
		Oid: "ALPHA",
	}, {
		Oid: "QUEUE",
	}, {
		Oid: "AAA",
	},
	}
	sort.Sort(ResourceIDSlice(ids2))
	if (ids2[0].Oid == "AAA") && (ids2[1].Oid == "ALPHA") && (ids2[2].Oid == "MACHINE1") && (ids2[3].Oid == "OUTLET24") && (ids2[4].Oid == "QUEUE") {
		t.Logf("SORT PASS Got '%s, %s, %s, %s, %s", ids2[0], ids2[1], ids2[2], ids2[3], ids2[4])
	} else {
		t.Errorf("SORT FAIL Got '%s, %s, %s, %s, %s", ids2[0], ids2[1], ids2[2], ids2[3], ids2[4])
	}
}
