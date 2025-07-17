// MIT License
//
// (C) Copyright [2025] Hewlett Packard Enterprise Development LP
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
	"testing"
)

func TestSmdFlavor(t *testing.T) {
	flavor, moduleName := getSmdFlavor()
	t.Logf("TestSmdFlavor values: flavor: %s, moduleName: %s", flavor, moduleName)

	if flavor == UnknownSmdFlavor {
		t.Fatalf("TestSmdFlavor unknown smd flavor: %s, moduleName: %s", flavor, moduleName)
	}

	if flavor == CSM {
		if !MSG_BUS_BUILD {
			t.Errorf("TestSmdFlavor exepected the msg bus to be enabled. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if !RF_EVENT_MONITOR_BUILD {
			t.Errorf("TestSmdFlavor exepected the rf event monitor to be enabled. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if !ENABLE_DISCOVERY_DEFAULT {
			t.Errorf("TestSmdFlavor exepected ENABLE_DISCOVERY_DEFAULT to be true. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if OPENCHAMI_DEFAULT {
			t.Errorf("TestSmdFlavor exepected OPENCHAMI_DEFAULT to be false. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if ZEROLOG_DEFAULT {
			t.Errorf("TestSmdFlavor exepected ZEROLOG_DEFAULT to be false. flavor: %s, moduleName: %s", flavor, moduleName)
		}
	} else if flavor == OpenCHAMI {
		if MSG_BUS_BUILD {
			t.Errorf("TestSmdFlavor exepected the msg bus to be disabled. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if RF_EVENT_MONITOR_BUILD {
			t.Errorf("TestSmdFlavor exepected the rf event monitor to be disabled. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if ENABLE_DISCOVERY_DEFAULT {
			t.Errorf("TestSmdFlavor exepected ENABLE_DISCOVERY_DEFAULT to be false. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if !OPENCHAMI_DEFAULT {
			t.Errorf("TestSmdFlavor exepected OPENCHAMI_DEFAULT to be true. flavor: %s, moduleName: %s", flavor, moduleName)
		}
		if !ZEROLOG_DEFAULT {
			t.Errorf("TestSmdFlavor exepected ZEROLOG_DEFAULT to be true. flavor: %s, moduleName: %s", flavor, moduleName)
		}
	}
}
