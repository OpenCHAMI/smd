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
	"os"
	"strings"
	"testing"
)

// inMAC: value to modify
// offset: int offset to use
// expMAC: expected result
//         expMAC == "" implies err != nil
//         expMAC != "" implies err = nil
type getOffsetTest struct {
	inMAC  string
	offset int64
	outMAC string
}

var getOffsetMACStringTests = []getOffsetTest{
	{"a4:bf:01:2e:7f:aa", 0, "a4:bf:01:2e:7f:aa"},
	{"a4-bf-01-2e-7f-aa", 0, "a4-bf-01-2e-7f-aa"},
	{"14-23-33-55-66-77", 0, "14-23-33-55-66-77"},
	{"a4:bf:01:2e:7f:aa", -1, "a4:bf:01:2e:7f:a9"},
	{"a4:bf:01:2e:7f:aa", -10, "a4:bf:01:2e:7f:a0"},
	{"a4:bf:01:2e:7f:aa", -11, "a4:bf:01:2e:7f:9f"},
	{"a4:bf:01:2e:7f:aa", -170, "a4:bf:01:2e:7f:00"},
	{"a4:bf:01:2e:7f:aa", -171, "a4:bf:01:2e:7e:ff"},
	{"a4:bf:01:2e:7f:aa", -256, "a4:bf:01:2e:7e:aa"},
	{"a4:bf:01:2e:7f:aa", -65536, "a4:bf:01:2d:7f:aa"},
	{"a4:bf:01:2e:7f:aa", 5, "a4:bf:01:2e:7f:af"},
	{"a4:bf:01:2e:7f:aa", 6, "a4:bf:01:2e:7f:b0"},
	{"a4:bf:01:2e:7f:aa", 85, "a4:bf:01:2e:7f:ff"},
	{"a4-bf-01-2e-7f-aa", 86, "a4-bf-01-2e-80-00"},
	{"a4:bf:01:2e:7f:AA", 87, "a4:bf:01:2e:80:01"},
	{"a4:bf:01:2e:7f:aa", 256, "a4:bf:01:2e:80:aa"},
	{"a4:bf:01:2E:7f:aa", 65536, "a4:bf:01:2f:7f:aa"},
	{"A4bf012e7faa", 0, "a4:bf:01:2e:7f:aa"},
	{"a4bf012e7faa", -1, "a4:bf:01:2e:7f:a9"},
	{"a4bf012e7faa", -10, "a4:bf:01:2e:7f:a0"},
	{"a4bf012e7faa", -11, "a4:bf:01:2e:7f:9f"},
	{"a4bf012e7faa", -170, "a4:bf:01:2e:7f:00"},
	{"a4bf012e7faa", -171, "a4:bf:01:2e:7e:ff"},
	{"a4BF012e7faa", -171, "a4:bf:01:2e:7e:ff"},
	{"a4bf012e7faa", -256, "a4:bf:01:2e:7e:aa"},
	{"a4bf012e7faa", -65536, "a4:bf:01:2d:7f:aa"},
	{"a4bf012e7faa", 5, "a4:bf:01:2e:7f:af"},
	{"a4bf012e7faa", 6, "a4:bf:01:2e:7f:b0"},
	{"a4bf012e7faa", 87, "a4:bf:01:2e:80:01"},
	{"a4bf012e7faa", 86, "a4:bf:01:2e:80:00"},
	{"a4bf012e7fAA", 85, "a4:bf:01:2e:7f:ff"},
	{"a4bf012e7faa", 256, "a4:bf:01:2e:80:aa"},
	{"a4bf012e7faa", 65536, "a4:bf:01:2f:7f:aa"},
	{"00:00:00:00:00:00", 0, "00:00:00:00:00:00"},
	{"ff:ff:ff:ff:ff:ff", 0, "ff:ff:ff:ff:ff:ff"},
	{"a4:bf:01:2e:7f:aaa", -1, ""},
	{"a4bf0127faa", 0, ""},
	{"0000bf012e7faa", -1, ""},
	{"00:00:00:00:00:00", -1, ""},
	{"ff:ff:ff:ff:ff:ff", 1, ""},
}

// We also use this for MACCompare (just the first value)
var getNormalizeVerifyMACTests = []getOffsetTest{
	{"a4:bf:01:2e:7f:aa", 0, "a4:bf:01:2e:7f:aa"},
	{"a4-bf-01-2e-7f-aa", 0, "a4:bf:01:2e:7f:aa"},
	{"a4bf.012e.7faa", 0, "a4:bf:01:2e:7f:aa"},
	{"a4.bf.01.2e.7f.aa", 0, "a4:bf:01:2e:7f:aa"},
	{"14-23-33-55-66-77", 0, "14:23:33:55:66:77"},
	{"142333556677", 0, "14:23:33:55:66:77"},
	{"1423335566", 0, ""},
	{"A4-BF-01-2E-7F-AA", 0, "a4:bf:01:2e:7f:aa"},
	{"A4:BF:01:2E:7F:AA", 0, "a4:bf:01:2e:7f:aa"},
	{"ff:ff:ff:ff:ff:ff", 0, "ff:ff:ff:ff:ff:ff"},
	{"FF:FF:FF:FF:FF:FF", 0, "ff:ff:ff:ff:ff:ff"},
	{"FF-FF-FF-FF-FF-FF", 0, "ff:ff:ff:ff:ff:ff"},
	{"00:00:00:00:00:00", 0, "00:00:00:00:00:00"},
	{"00-00-00-00-00-00", 0, "00:00:00:00:00:00"},
	{"a4:bf:01:2e:7f:aaa", 0, ""},
	{"a4-bf-01-2e-7f-aa-af", 0, ""},
}

func TestGetOffsetMACString(t *testing.T) {
	t.Logf("Performing lower case tests...")
	for i, test := range getOffsetMACStringTests {
		ans, err := GetOffsetMACString(test.inMAC, test.offset)
		if ans == test.outMAC {
			if ans == "" {
				if err != nil {
					t.Logf("Testcase %d: PASS %s(%d): Got err as expected: %s",
						i, test.inMAC, test.offset, err)
				} else {
					t.Errorf("Testcase %d: FAIL %s(%d): Err was nil.",
						i, test.inMAC, test.offset)
				}
			} else if err != nil {
				t.Errorf("Testcase %d: FAIL %s(%d): Got %s, but err !=nil: %s",
					i, test.inMAC, test.offset, test.outMAC, err)
			} else {
				t.Logf("Testcase %d: PASS %s(%d): Got %s as expected.",
					i, test.inMAC, test.offset, test.outMAC)
			}
		} else {
			t.Errorf("Testcase %d: FAIL %s(%d): Did not get %s as expected: %s",
				i, test.inMAC, test.offset, test.outMAC, ans)
		}
	}
	t.Logf("Performing upper case tests...")
	for i, test := range getOffsetMACStringTests {
		inMAC := strings.ToUpper(test.inMAC)
		outMAC := strings.ToUpper(test.outMAC)
		ans, err := GetOffsetMACString(inMAC, test.offset)
		if ans == outMAC {
			if ans == "" {
				if err != nil {
					t.Logf("Testcase %d: PASS %s(%d): Got err as expected: %s",
						i, inMAC, test.offset, err)
				} else {
					t.Errorf("Testcase %d: FAIL %s(%d): Err unexpectedly nil.",
						i, inMAC, test.offset)
				}
			} else if err != nil {
				t.Errorf("Testcase %d: FAIL  %s(%d): Got %s, but err !=nil: %s",
					i, inMAC, test.offset, outMAC, err)
			} else {
				t.Logf("Testcase %d: PASS %s(%d): Got %s as expected.",
					i, inMAC, test.offset, outMAC)
			}
		} else {
			t.Errorf("Testcase %d: FAIL %s(%d): Did not get %s as expected: %s",
				i, inMAC, test.offset, outMAC, ans)
		}
	}
}

func TestNormalizeVerifyMAC(t *testing.T) {
	t.Logf("Testing NormalizeVerifyMAC")
	for i, test := range getNormalizeVerifyMACTests {
		inMAC := test.inMAC
		outMAC := test.outMAC
		ans, err := NormalizeVerifyMAC(inMAC)
		if ans == outMAC {
			if ans == "" {
				if err != nil {
					t.Logf("Testcase %d: PASS %s: Got err as expected: %s",
						i, inMAC, err)
				} else {
					t.Errorf("Testcase %d: FAIL %s: Err was nil.", i, inMAC)
				}
			} else if err != nil {
				t.Errorf("Testcase %d: FAIL %s: Got %s, but err !=nil: %s",
					i, inMAC, outMAC, err)
			} else {
				t.Logf("Testcase %d: PASS %s: Got %s as expected.",
					i, inMAC, outMAC)
			}
		} else {
			t.Errorf("Testcase %d: FAIL %s: Did not get %s as expected: %s",
				i, inMAC, outMAC, ans)
		}
	}
}

func TestNormalizeMAC(t *testing.T) {
	t.Logf("Testing NormalizeMAC")
	for i, test := range getNormalizeVerifyMACTests {
		inMAC := test.inMAC
		outMAC := test.outMAC
		ans := NormalizeMAC(inMAC)
		if outMAC == "" {
			if ans != inMAC {
				t.Errorf("Testcase %d: FAIL %s: Did not get original string.",
					i, inMAC)
			} else {
				t.Logf("Testcase %d: PASS %s: Got %s as expected.",
					i, inMAC, inMAC)
			}
		} else if ans != outMAC {
			t.Errorf("Testcase %d: FAIL %s: Did not get %s as expected: %s",
				i, inMAC, outMAC, ans)
		} else {
			t.Logf("Testcase %d: PASS %s: Got %s as expected.",
				i, inMAC, outMAC)
		}
	}
}

func TestNormalizeMACIfValid(t *testing.T) {
	t.Logf("Testing NormalizeMACIfValid")
	for i, test := range getNormalizeVerifyMACTests {
		inMAC := test.inMAC
		outMAC := test.outMAC
		ans := NormalizeMACIfValid(inMAC)
		if outMAC == "" {
			if ans != "" {
				t.Errorf("Testcase %d: FAIL %s: Did not get empty string: %s",
					i, inMAC, ans)
			} else {
				t.Logf("Testcase %d: PASS %s: Got empty string as expected.",
					i, inMAC)
			}
		} else if ans != outMAC {
			t.Errorf("Testcase %d: FAIL %s: Did not get %s as expected: %s",
				i, inMAC, outMAC, ans)
		} else {
			t.Logf("Testcase %d: PASS %s: Got %s as expected.",
				i, inMAC, outMAC)
		}
	}
}

func TestMACCompare(t *testing.T) {
	t.Logf("MACCompare")
	for idx, test := range getNormalizeVerifyMACTests {
		for offset := 1; offset <= 256; offset *= 2 {
			inMAC := test.inMAC
			outMAC := test.outMAC
			inMACComp, err := GetOffsetMACString(inMAC, int64(offset))
			if err != nil {
				// Skip overflows
				continue
			}
			// First arg is lower
			ans, err := MACCompare(inMAC, inMACComp)
			if err != nil {
				if outMAC == "" {
					t.Logf("Testcase %d: PASS %s/%s: Got err as expected: %s",
						idx, inMAC, inMACComp, err)
				} else {
					t.Errorf("Testcase %d: FAIL %s/%s: Unexpected error: %s",
						idx, inMAC, inMACComp, err)
				}
			} else {
				if ans == -1 {
					t.Logf("Testcase %d: PASS %s/%s: Got -1 as expected",
						idx, inMAC, inMACComp)
				} else {
					t.Errorf("Testcase %d: FAIL %s/%s: got %d != 1",
						idx, inMAC, inMACComp, ans)
				}
			}
			// First arg is greater
			ans, err = MACCompare(inMACComp, inMAC)
			if err != nil {
				if outMAC == "" {
					t.Logf("Testcase %d: PASS %s/%s: Got err as expected: %s",
						idx, inMACComp, inMAC, err)
				} else {
					t.Errorf("Testcase %d: FAIL %s/%s: Unexpected error: %s",
						idx, inMACComp, inMAC, err)
				}
			} else {
				if ans == 1 {
					t.Logf("Testcase %d: PASS %s/%s: Got 1 as expected",
						idx, inMACComp, inMAC)
				} else {
					t.Errorf("Testcase %d: FAIL %s/%s: got %d != 1",
						idx, inMACComp, inMAC, ans)
				}
			}
			// Args are equal
			ans, err = MACCompare(inMAC, inMAC)
			if err != nil {
				if outMAC == "" {
					t.Logf("Testcase %d: PASS %s/%s: Got err as expected: %s",
						idx, inMAC, inMAC, err)
				} else {
					t.Errorf("Testcase %d: FAIL %s/%s: Unexpected error: %s",
						idx, inMAC, inMAC, err)
				}
			} else {
				if ans == 0 {
					t.Logf("Testcase %d: PASS %s/%s: Got 0 as expected: %s",
						idx, inMAC, inMAC, err)
				} else {
					t.Errorf("Testcase %d: FAIL %s/%s: got %d != 0",
						idx, inMAC, inMAC, ans)
				}
			}
		}
	}
}

func TestEnableGenTestingPayloads(t *testing.T) {
	var argsGood = []string{
		"x0c0b0",
		"x0c0s0b0:Title",
	}
	var argsBad = []string{
		"x0c0s0b0:Title:extra", // extra arg
		"x0.c0:Title",          // bad EP ID 1 - not alpha-numeric
		"x0.c0",                // bad EP ID 2, ditto
		"x0c0b0:Title/Title2",  // bad title, not alpha-numeric.
	}
	for i, arg := range argsGood {
		if err := EnableGenTestingPayloads(arg); err != nil {
			t.Errorf("Testcase %da: FAIL Got unexpected error for '%s': %s",
				i, arg, err)
		} else {
			t.Logf("Testcase %da: PASS Got no error for '%s'", i, arg)
		}
	}
	for i, arg := range argsBad {
		if err := EnableGenTestingPayloads(arg); err != nil {
			t.Logf("Testcase %db: Pass Got expected error for '%s': %s",
				i, arg, err)
		} else {
			t.Errorf("Testcase %db: Fail Got no error for '%s'", i, arg)
		}
	}
}

func TestGenTestingPayloads(t *testing.T) {
	payload := []byte("{}")
	err := GenTestingPayloads(os.Stdout,
		"Title",
		"/redfish/v1/whatever",
		payload)
	if err != nil {
		t.Errorf("Test 1: FAIL Got error '%s'", err)
	}
}

type addrCheck struct {
	in       string
	expected string
}

func TestGetIPAddressString(t *testing.T) {
	var checkAddrs = []addrCheck{
		{"127.0.0.1", "127.0.0.1"},
		{"127.0.0.1:1234", "127.0.0.1:1234"},
		{"127.0.0.1%en0:1234", ""},
		{"127.0.0.1%en0", ""},
		{"[127.0.0.1%en0]", ""},
		{"[127.0.0.1%en0]:234", ""},
		{"asdf:asdfae", ""},
		{"host.domain.com", ""},
		{"host.domain.com:9999", ""},
		{"fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670", "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]"},
		{"fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f670", "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]"},
		{"[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]", "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]"},
		{"[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f670]", "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]"},
		{"[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f670]:1234", "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]:1234"},
		{"[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f670%eno0]:1234", "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670%eno0]:1234"},
		{"FD40:AAAA:BBBB:1007:0240:A6ff:FE82:F670", "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f670]"},
	}
	for i, addr := range checkAddrs {
		check := GetIPAddressString(addr.in)
		if check == addr.expected {
			t.Logf("Testcase %d: Pass Matched as expected: '%s'",
				i, check)
		} else {
			t.Errorf("Testcase %d: FAIL '%s' did not match expected '%s'",
				i, check, addr.expected)
		}
	}
}
