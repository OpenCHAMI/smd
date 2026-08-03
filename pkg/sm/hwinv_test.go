// MIT License
//
// (C) Copyright [2019-2023] Hewlett Packard Enterprise Development LP
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

package sm_test

import (
	"testing"

	stest "github.com/openchami/smd/v2/pkg/sharedtest"
	"github.com/openchami/smd/v2/pkg/sm"
)

///////////////////////////////////////////////////////////////////////////////
// Tests.  NOTE: INTENDED TO BE RUN IN THIS ORDER.  Some tests modify global
// data.
///////////////////////////////////////////////////////////////////////////////

func TestNewSystemHWInventory(t *testing.T) {
	hwinv, err := sm.NewSystemHWInventory(
		stest.HWInvByLocArray1,
		"s0",
		sm.HWInvFormatFullyFlat)
	if err != nil {
		t.Errorf("Test 1 Failed: Got error '%s'", err)
	} else if hwinv == nil {
		t.Errorf("Test 1 Failed: Got nil hwinv")
	}
	t.Log("Test 1 PASS")
	hwinv, err = sm.NewSystemHWInventory(
		stest.HWInvByLocArray1,
		"s0",
		sm.HWInvFormatNestNodesOnly)
	if err != nil {
		t.Errorf("Test 2 Failed: Got error '%s'", err)
	} else if hwinv == nil {
		t.Errorf("Test 2 Failed: Got nil hwinv")
	}
	t.Log("Test 2 PASS")
	hwinv, err = sm.NewSystemHWInventory(
		stest.HWInvByLocArray1,
		"s0",
		sm.HWInvFormatHierarchical)
	if err == nil {
		t.Errorf("Test 3 Failed: Got nil err")
	} else if err != sm.ErrHWInvFmtNI {
		t.Errorf("Test 3 Failed: Got wrong error '%s'", err)
	}
	t.Log("Test 3 PASS")
	hwinv, err = sm.NewSystemHWInventory(
		stest.HWInvByLocArray1,
		"s0",
		"SomeInvalidFormatType11111")
	if err == nil || hwinv != nil {
		t.Errorf("Test 4 Failed: Got nil err or hwinv != nil")
	} else if err != sm.ErrHWInvFmtInvalid {
		t.Errorf("Test 4 Failed: Got wrong error '%s'", err)
	}
	t.Log("Test 4 PASS")
}

func TestEncodeLocationInfo(t *testing.T) {
	for i, hwloc := range stest.HWInvByLocArray1 {
		bytes, err := hwloc.EncodeLocationInfo()
		if err != nil {
			t.Errorf("Test %d Failed: Got error '%s'", i, err)
		} else if len(bytes) == 0 {
			t.Errorf("Test %d Failed: Zero bytes returned", i)
		} else {
			t.Logf("Test %d PASS", i)
		}
	}
}

func TestDecodeLocationInfo(t *testing.T) {
	for i, hwloc := range stest.HWInvByLocArray1 {
		bytes, err := hwloc.EncodeLocationInfo()
		if err != nil {
			t.Errorf("Test %d Failed (pre-test): Got error '%s'", i, err)
			continue
		}
		err = hwloc.DecodeLocationInfo(bytes)
		if err != nil {
			t.Errorf("Test %d Failed: Got error '%s'", i, err)
			continue
		}
		t.Logf("Test %d PASS", i)
	}
}

func TestEncodeFRUInfo(t *testing.T) {
	for i, hwfru := range stest.HWInvByFRUArray1 {
		bytes, err := hwfru.EncodeFRUInfo()
		if err != nil {
			t.Errorf("Test %d Failed: Got error '%s'", i, err)
		} else if len(bytes) == 0 {
			t.Errorf("Test %d Failed: Zero bytes returned", i)
		} else {
			t.Logf("Test %d PASS", i)
		}
	}
}

func TestDecodeFRUInfo(t *testing.T) {
	for i, hwfru := range stest.HWInvByFRUArray1 {
		bytes, err := hwfru.EncodeFRUInfo()
		if err != nil {
			t.Errorf("Test %d Failed (pre-test): Got error '%s'", i, err)
			continue
		}
		err = hwfru.DecodeFRUInfo(bytes)
		if err != nil {
			t.Errorf("Test %d Failed: Got error '%s'", i, err)
			continue
		}
		t.Logf("Test %d PASS", i)
	}
}
