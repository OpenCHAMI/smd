// MIT License
//
// (C) Copyright [2019-2021,2024] Hewlett Packard Enterprise Development LP
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
	"fmt"
	"strings"
	"testing"
)

var goodRawEPs = RawRedfishEPs{
	RedfishEndpoints: []RawRedfishEP{
		{
			ID:       "x0c0s14b0",
			FQDN:     "x0c0s14.crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			FQDN:     "x0c0s16b0.crush.next.cray.com:443",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s18b0",
			Hostname: "x0c0s18",
			Domain:   "crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s20b0",
			Hostname: "10.100.16.22",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s21b0",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s24b0",
			Hostname: "x0c0s24",
			FQDN:     "x0c0s24.crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s25b0",
			Hostname: "x0c0s25",
			Domain:   "crush.next.cray.com",
			FQDN:     "x0c0s25.crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			Hostname: "x0c0s26b0",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s27b0",
			Domain:   "crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x1c0s2b0",
			FQDN:     "172.30.79.230",
			User:     "root",
			Password: "********",
		}, {
			ID:       "x1c0b0",
			FQDN:     "172.30.79.1",
			User:     "root",
			Password: "********",
		}, {
			ID:       "x0c0s22b0",
			Hostname: "10.100.16.23",
			Domain:   "crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s30b0",
			Hostname: "10.100.16.24",
			Domain:   "crush.next.cray.com",
			FQDN:     "10.100.16.24",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s31b0",
			FQDN:     "fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s32b0",
			Hostname: "fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2",
			Domain:   "crush.next.cray.com",
			FQDN:     "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2]",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s33b0",
			Hostname: "fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2",
			Domain:   "crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s34b0",
			FQDN:     "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2]",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s35b0",
			FQDN:     "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2]",
			Hostname: "fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2",
			Domain:   "crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s36b0",
			FQDN:     "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2]",
			Hostname: "[fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2]",
			Domain:   "crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s37b0",
			FQDN:     "fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f7c2",
			Hostname: "fd40:aaaa:bbbb:1007:240:a6ff:fe82:f7c2",
			Domain:   "crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s38b0",
			Hostname: "fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f7c2",
			FQDN:     "[fd40:aaaa:bbbb:1007:0240:a6ff:fe82:f7c2]",
			Domain:   "",
			User:     "root",
			Password: "calvin",
		},
	},
}

var badRawEPs = RawRedfishEPs{
	RedfishEndpoints: []RawRedfishEP{
		{
			ID:       "x0c0s14", // Not a valid XName for controller
			FQDN:     "x0c0s14.crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			FQDN:     "x0c0s16.crush.next.cray.com", // Does not contain xname
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s18b0",
			Hostname: "x0c0s18", // FQDN is in conflict with Domain/Hostname
			Domain:   "crush.next.cray.com",
			FQDN:     "badhostname.crush.next.cray.com",
			User:     "root",
			Password: "calvin",
		}, {
			Hostname: "x0c0s24", // No ID, Hostname is not controller xname
			Domain:   "crush.next.cray.com",
			FQDN:     "10.100.16.22",
			User:     "root",
			Password: "calvin",
		}, {
			ID:       "x0c0s25b0",
			Hostname: "x0c0s25.crush", // No ID, Hostname is not controller xname
			Domain:   "next.cray.com",
			User:     "root",
			Password: "calvin",
		},
	},
}

func TestNewRedfishEPDescription(t *testing.T) {
	epd, err := NewRedfishEPDescription(nil)
	if err == nil {
		t.Errorf("Unexpectedly got nil error for nil argument")
	} else if epd != nil {
		t.Errorf("Unexpectedly got non-nil return with nil argument")
	} else {
		t.Logf("Got error as expected for nil arg: %s", err)
	}
	t.Logf("Performing positive tests...")
	for i, rep := range goodRawEPs.RedfishEndpoints {
		// make sure the string version of the RawRedfishEP does not include password
		if strings.Contains(fmt.Sprintf("%v", rep), "calvin") {
			t.Errorf("Testcase %d: String result of RawRedfishEP contains password", i)
		}

		epd, err := NewRedfishEPDescription(&rep)
		if err != nil {
			t.Errorf("Testcase %d: Got unexpected error: %s", i, err)
		} else if epd == nil {
			t.Errorf("Testcase %d: Got unexpected nil pointer.", i)
		} else if epd.ID == "" || epd.FQDN == "" {
			t.Errorf("Testcase %d: ID or FQDN unexpectedly empty", i)
		} else {
			// make sure the string function redacts the password
			if strings.Contains(fmt.Sprintf("%v", epd), "calvin") {
				t.Errorf("Testcase %d: String result of RedfishEPDescription contains password", i)
			}
			t.Logf("Testcase %d: Result is ok.", i)
		}
	}
	t.Logf("Performing negative tests...")
	for i, rep := range badRawEPs.RedfishEndpoints {
		epd, err := NewRedfishEPDescription(&rep)
		if err == nil {
			t.Errorf("Testcase %d: err was unexpectedly nil: %v", i, epd)
		} else if epd != nil {
			t.Errorf("Testcase %d: Got unexpected non-nil pointer.", i)
		} else {
			t.Logf("Testcase %d: Got error as expected: %s", i, err)
		}
	}
}

// This is the bulk-create version of the former
func TestNewRedfishEPDescriptions(t *testing.T) {
	epds, err := NewRedfishEPDescriptions(nil)
	if err == nil {
		t.Errorf("Unexpectedly got nil error for nil argument")
	} else {
		t.Logf("Got error as expected for nil arg: %s", err)
	}
	t.Logf("Performing positive tests...")
	epds, err = NewRedfishEPDescriptions(&goodRawEPs)
	if err != nil {
		t.Errorf("Testcase 1: Got unexpected error: %s", err)
	} else if epds == nil {
		t.Errorf("Testcase 1: Got unexpected nil pointer.")
	} else {
		t.Logf("Testcase 1: Results are ok.")
		if len(epds.RfEPDescriptions) != len(goodRawEPs.RedfishEndpoints) {
			t.Errorf("Testcase 2: Got %d descs, expected %d",
				len(epds.RfEPDescriptions), len(goodRawEPs.RedfishEndpoints))
		} else {
			t.Logf("Testcase 2: Got all %d entries.",
				len(epds.RfEPDescriptions))
		}
	}
	t.Logf("Performing negative tests...")
	epds, err = NewRedfishEPDescriptions(&badRawEPs)
	if err == nil {
		t.Errorf("Testcase 1: err was unexpectedly nil")
	} else if epds == nil {
		t.Errorf("Testcase 1: Got unexpected nil pointer.")
	} else {
		t.Logf("Testcase 1: Got error return as expected: %s", err)
		if len(epds.RfEPDescriptions) != 0 {
			t.Errorf("Testcase 2: Length of descriptions was non-zero: %d",
				len(epds.RfEPDescriptions))
		} else {
			t.Logf("Testcase 2: Length of descriptions was zero as expected")
		}
	}
}

// Create endpoints for discover from valid descriptions generated from
// valid raw endpoints.
func TestNewRedfishEp(t *testing.T) {
	_, err := NewRedfishEp(nil)
	if err == nil {
		t.Errorf("Unexpectedly got nil error for nil argument")
	} else {
		t.Logf("Got error as expected for nil arg: %s", err)
	}

	// Setup endpoint descriptions from good raw endpoint array
	epds, err := NewRedfishEPDescriptions(&goodRawEPs)
	if err != nil {
		t.Errorf("Setup failure: Got unexpected error: %s", err)
	} else if epds == nil {
		t.Errorf("Setup failure: Got unexpected nil pointer.")
	}

	// Test generation of discover-ready endpoints from each description
	t.Logf("Performing positive tests...")
	for i, epd := range epds.RfEPDescriptions {
		ep, err := NewRedfishEp(&epd)
		if err != nil {
			t.Errorf("Testcase %d: Got unexpected error: %s", i, err)
		} else if ep == nil {
			t.Errorf("Testcase %d: Got unexpected nil return value", i)
		} else {
			t.Logf("Testcase %d: No error.", i)
		}
	}
}

// Bulk/collection version of NewRedfishEp
func TestNewRedfishEps(t *testing.T) {
	eps, err := NewRedfishEps(nil)
	if err == nil {
		t.Errorf("Unexpectedly got nil error for nil argument")
	} else {
		t.Logf("Got error as expected for nil arg: %s", err)
	}

	// Setup endpoint descriptions from good raw endpoint array
	epds, err := NewRedfishEPDescriptions(&goodRawEPs)
	if err != nil {
		t.Errorf("Setup failure: Got unexpected error: %s", err)
	} else if epds == nil {
		t.Errorf("Setup failure: Got unexpected nil pointer.")
	}
	expectedLen := len(epds.RfEPDescriptions)

	// Test generation of discover-ready endpoints from descriptions
	eps, err = NewRedfishEps(epds)
	if err != nil {
		t.Errorf("Testcase 1: Got unexpected error: %s", err)
	} else if eps == nil {
		t.Errorf("Testcase 1: Got unexpected nil return value")
	} else {
		// make sure no passwords are entered into string version of struct
		for epName, ep := range eps.IDs {
			if strings.Contains(fmt.Sprintf("%v", ep), "calvin") {
				t.Errorf("Testcase %s: String result of RedfishEP contains password", epName)
			}
		}
		t.Logf("Testcase 1: Got no errors.")
		if eps.Num != len(eps.IDs) {
			t.Errorf("Testcase 2: Num doesn't match len: %d/%d",
				eps.Num, len(eps.IDs))
		} else {
			t.Logf("Testcase 2: Num matches len: %d/%d",
				eps.Num, len(eps.IDs))
			if eps.Num != expectedLen {
				t.Errorf("Testcase 3: Expected len != eps count: %d/%d",
					eps.Num, len(eps.IDs))
			} else {
				t.Logf("Testcase 3: Expected len matches eps count: %d/%d",
					eps.Num, len(eps.IDs))
			}
		}
	}
}

func TestGetMemoryFRUID(t *testing.T) {

	//create a RF endpoint
	rep := RawRedfishEP{
		ID:       "x0c0s14b0",
		FQDN:     "x0c0s14.crush.next.cray.com",
		User:     "root",
		Password: "calvin",
	}

	epd, err := NewRedfishEPDescription(&rep)
	if err != nil {
		t.Errorf("TestGetMemoryFRUID: Got unexpected error: %s", err)
	} else if epd == nil {
		t.Error("TestGetMemoryFRUID: Got unexpected nil pointer.")
	} else if epd.ID == "" || epd.FQDN == "" {
		t.Error("TestGetMemoryFRUID: ID or FQDN unexpectedly empty")
	} else {
		t.Log("TestGetMemoryFRUID: Successfully created NewRedfishEPDescription.")
	}

	ep, err := NewRedfishEp(epd)
	if err != nil {
		t.Errorf("TestGetMemoryFRUID: Got unexpected error: %s", err)
	} else if ep == nil {
		t.Error("TestGetMemoryFRUID: Got unexpected nil return value")
	} else {
		t.Log("TestGetMemoryFRUID: Successfully created NewRedfishEp.")
	}
	sys1 := NewEpSystem(ep, ResourceID{"/some/path/to/a/system"}, 0)
	sys1.ID = ep.ID
	mep1 := NewEpMemory(sys1, ResourceID{"/some/path/to/a/memory"}, 0)
	mep1.ID = sys1.ID + "d1"
	mep1.MemoryRF = Memory{
		MemoryFRUInfoRF: MemoryFRUInfoRF{
			Manufacturer: "test1Manufacturer",
			PartNumber:   "test1PartNumber",
			SerialNumber: "test1SerialNumber",
		},
	}
	mep2 := NewEpMemory(sys1, ResourceID{"/some/path/to/a/memory"}, 0)
	mep2.ID = sys1.ID + "d2"
	mep2.MemoryRF = Memory{
		MemoryFRUInfoRF: MemoryFRUInfoRF{
			Manufacturer: "",
			PartNumber:   "test2PartNumber",
			SerialNumber: "test2SerialNumber",
		},
	}
	mep3 := NewEpMemory(sys1, ResourceID{"/some/path/to/a/memory"}, 0)
	mep3.ID = sys1.ID + "d3"
	mep3.MemoryRF = Memory{
		MemoryFRUInfoRF: MemoryFRUInfoRF{
			Manufacturer: "test3Manufacturer",
			PartNumber:   "",
			SerialNumber: "test3SerialNumber",
		},
	}
	mep4 := NewEpMemory(sys1, ResourceID{"/some/path/to/a/memory"}, 0)
	mep4.ID = sys1.ID + "d4"
	mep4.MemoryRF = Memory{
		MemoryFRUInfoRF: MemoryFRUInfoRF{
			Manufacturer: "",
			PartNumber:   "",
			SerialNumber: "test4SerialNumber",
		},
	}
	mep5 := NewEpMemory(sys1, ResourceID{"/some/path/to/a/memory"}, 0)
	mep5.ID = sys1.ID + "d5"
	mep5.MemoryRF = Memory{
		MemoryFRUInfoRF: MemoryFRUInfoRF{
			Manufacturer: "test5Manufacturer",
			PartNumber:   "test5PartNumber",
			SerialNumber: "",
		},
	}
	mep6 := NewEpMemory(sys1, ResourceID{"/some/path/to/a/memory"}, 0)
	mep6.ID = sys1.ID + "d6"

	tests := []struct {
		mep           *EpMemory
		expectedFRUID string
		expectedError bool
	}{
		{
			mep:           mep1,
			expectedFRUID: "Memory.test1Manufacturer.test1PartNumber.test1SerialNumber",
			expectedError: false,
		},
		{
			mep:           mep2,
			expectedFRUID: "Memory.test2PartNumber.test2SerialNumber",
			expectedError: false,
		},
		{
			mep:           mep3,
			expectedFRUID: "Memory.test3Manufacturer.test3SerialNumber",
			expectedError: false,
		},
		{
			mep:           mep4,
			expectedFRUID: "FRUIDforx0c0s14b0d4",
			expectedError: true,
		},
		{
			mep:           mep5,
			expectedFRUID: "FRUIDforx0c0s14b0d5",
			expectedError: true,
		},
		{
			mep:           mep6,
			expectedFRUID: "FRUIDforx0c0s14b0d6",
			expectedError: true,
		},
	}

	for n, test := range tests {
		fruid, err := GetMemoryFRUID(test.mep)
		if err != nil {
			if test.expectedError == true {
				t.Logf("TestGetMemoryFRUID 1 Test Case %d: fruid %s, got error as expected: %s", n, fruid, err.Error())
				if test.expectedFRUID != fruid {
					t.Errorf("TestGetMemoryFRUID 2 Test Case %d: FAIL: FRUID Expected %s but instead got %s", n, test.expectedFRUID, fruid)
				}
			} else {
				t.Errorf("TestGetMemoryFRUID 3 Test Case %d: got unexpected error: %s", n, err.Error())
			}
		} else {
			if test.expectedError {
				t.Errorf("TestGetMemoryFRUID 4 Test Case %d: FAIL: Error Expected but instead got fruid %s", n, fruid)
			} else if test.expectedFRUID != fruid {
				t.Errorf("TestGetMemoryFRUID 5 Test Case %d: FAIL: Error Expected %s but instead got %s", n, test.expectedFRUID, fruid)
			} else {
				t.Logf("TestGetMemoryFRUID 6 Test Case %d: got expected fruid: %s", n, fruid)
			}
		}
	}
}

func TestGetStandardFRUID(t *testing.T) {
	tests := []struct {
		compType      string
		id            string
		manufacturer  string
		partnum       string
		serialnum     string
		expectedFRUID string
		expectedError bool
	}{{ // Test 0 - Success all fields
		compType:      "Node",
		id:            "x0c0s0b0n0",
		manufacturer:  "Intel Corporation",
		partnum:       "testPartNumber",
		serialnum:     "testSerialNumber",
		expectedFRUID: "Node.IntelCorporation.testPartNumber.testSerialNumber",
		expectedError: false,
	}, { // Test 1 - Success no manufacturer
		compType:      "Node",
		id:            "x0c0s0b0n0",
		manufacturer:  "",
		partnum:       "testPartNumber",
		serialnum:     "testSerialNumber",
		expectedFRUID: "Node.testPartNumber.testSerialNumber",
		expectedError: false,
	}, { // Test 2 - Success no part number
		compType:      "Node",
		id:            "x0c0s0b0n0",
		manufacturer:  "Intel Corporation",
		partnum:       "",
		serialnum:     "testSerialNumber",
		expectedFRUID: "Node.IntelCorporation.testSerialNumber",
		expectedError: false,
	}, { // Test 3 - Fail no serial number
		compType:      "Node",
		id:            "x0c0s0b0n0",
		manufacturer:  "Intel Corporation",
		partnum:       "testPartNumber",
		serialnum:     "",
		expectedFRUID: "FRUIDforx0c0s0b0n0",
		expectedError: true,
	}, { // Test 4 - Fail no part number or manufacturer
		compType:      "Node",
		id:            "x0c0s0b0n0",
		manufacturer:  "",
		partnum:       "",
		serialnum:     "testSerialNumber",
		expectedFRUID: "FRUIDforx0c0s0b0n0",
		expectedError: true,
	}, { // Test 5 - Fail detect "..." as empty fields
		compType:      "Node",
		id:            "x0c0s0b0n0",
		manufacturer:  "Intel Corporation",
		partnum:       "..................",
		serialnum:     "..................",
		expectedFRUID: "FRUIDforx0c0s0b0n0",
		expectedError: true,
	}, { // Test 6 - Fail too long (260 length)
		compType:      "Node",
		id:            "x0c0s0b0n0",
		manufacturer:  "Reeeeeeeeeeeeeeeeeally Loooooooooooooong Maaaaaaaaaaaaaaaaaaaaaaaaaaaaaanufacturer Naaaaaaaaaaaaaaaaaaaaaaaaame",
		partnum:       "WhoUsesPartNumbersThatAreThisLoooooooooooooooooooooooooooooooooooooooooong",
		serialnum:     "WhyIsThisSerialNumberSoooooooooooooooooooooooooooooooooooooooooooooLong",
		expectedFRUID: "FRUIDforx0c0s0b0n0",
		expectedError: true,
	}}
	for i, test := range tests {
		fruid, err := getStandardFRUID(test.compType, test.id, test.manufacturer, test.partnum, test.serialnum)
		if err != nil && !test.expectedError {
			t.Errorf("Test %v Failed: Unexpected error received '%s'", i, err)
		} else if err == nil && test.expectedError {
			t.Errorf("Test %v Failed: Expected an error but got FRUID '%s'", i, fruid)
		} else {
			if test.expectedFRUID != fruid {
				t.Errorf("Test %v Failed: Expected FRUID '%s'; Received %s", i, test.expectedFRUID, fruid)
			}
		}
	}
}

func TestGetSystemArch(t *testing.T) {
	type args struct {
		s EpSystem
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "Test Cray ex235a",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "ex235a",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "bardpeaknc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray ex420",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "ex420",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "cnc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray ex425",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "ex425",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "wnc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray ex254n",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "ex254n",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "blancapeaknc",
					},
				},
			},
		},
		want: "ARM",
	}, {
		name: "Test Cray ex255a",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "ex255a",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "parrypeaknc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray ex4252",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "ex4252",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "antero",
					},
				},
			},
		},
		want: "X86",
	}, {
                name: "Test Foxconn HPE Cray Supercomputing XD224 (Paradise)",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Foxconn",
						Model:        "HPE Cray Supercomputing XD224",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "Computer System",
					},
				},
			},
		},
		want: "ARM",
	}, {
		name: "Test Foxconn 1A62WCB00-600-G (Paradise)",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Foxconn",
						Model: "1A62WCB00-600-G",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "Computer System",
					},
				},
			},
		},
		want: "ARM",
	}, {
		name: "Test Cray bardpeaknc",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "bardpeaknc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray cnc",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "cnc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray wnc",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "wnc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray blancapeaknc",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "blancapeaknc",
					},
				},
			},
		},
		want: "ARM",
	}, {
		name: "Test Cray parrypeaknc",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "parrypeaknc",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Cray antero",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "antero",
					},
				},
			},
		},
		want: "X86",
	}, {
		name: "Test Foxconn HPE Cray Supercomputing XD224",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Foxconn",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "Computer System",
					},
				},
			},
		},
		want: "UNKNOWN",
	}, {
		name: "Test Bad",
		args: args{
			EpSystem{
				SystemRF: ComputerSystem{
					SystemFRUInfoRF: SystemFRUInfoRF{
						Manufacturer: "Cray",
						Model:        "xxxxx",
					},
					SystemLocationInfoRF: SystemLocationInfoRF{
						Description: "xxxxx",
					},
				},
			},
		},
		want: "UNKNOWN",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSystemArch(&tt.args.s); got != tt.want {
				t.Errorf("GetSystemArch() = %v, want %v", got, tt.want)
			}
		})
	}
}
