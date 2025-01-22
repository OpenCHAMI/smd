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

package sharedtest

import (
	"encoding/json"

	"github.com/Cray-HPE/hms-smd/v2/pkg/rf"
	"github.com/Cray-HPE/hms-smd/v2/pkg/sm"
)

///////////////////////////////////////////////////////////////////////////////
// HW Inventory - Dummy payloads to be copied and modified as needed for tests.
///////////////////////////////////////////////////////////////////////////////

// Node
var NodeHWInvByFRU1 = sm.HWInvByFRU{
	FRUID:                "Dell-99999-1234-1234-2345",
	Type:                 "Node",
	Subtype:              "River",
	HWInventoryByFRUType: "HWInvByFRUNode",
	HMSNodeFRUInfo: &rf.SystemFRUInfoRF{
		AssetTag:     "AdminAssignedAssetTag",
		BiosVersion:  "v1.0.2.9999",
		Model:        "OKS0P2354",
		Manufacturer: "Dell",
		PartNumber:   "99999",
		SerialNumber: "1234-1234-2345",
		SKU:          "as213234",
		SystemType:   "Physical",
		UUID:         "26276e2a-29dd-43eb-8ca6-8186bbc3d971",
	},
}

var NodeHWInvByLoc1 = sm.HWInvByLoc{
	ID:                        "x0c0s0b0n0",
	Type:                      "Node",
	Ordinal:                   0,
	Status:                    "Populated",
	HWInventoryByLocationType: "HWInvByLocNode",
	HMSNodeLocationInfo: &rf.SystemLocationInfoRF{
		Id:          "System.Embedded.1",
		Name:        "Name describing system or where it is located, per mgfr",
		Description: "Description of system/node type, per mgfr",
		Hostname:    "if_defined_in_Redfish",
		ProcessorSummary: rf.ComputerSystemProcessorSummary{
			Count: json.Number("2"),
			Model: "Multi-Core Intel(R) Xeon(R) processor E5-16xx Series",
		},
		MemorySummary: rf.ComputerSystemMemorySummary{
			TotalSystemMemoryGiB: json.Number("64"),
		},
	},
	PopulatedFRU: &NodeHWInvByFRU1,
}

// Processors for node (two)
var ProcHWInvByFRU1 = sm.HWInvByFRU{
	FRUID:                "HOW-TO-ID-CPUS-FROM-REDFISH-IF-AT-ALL1",
	Type:                 "Processor",
	Subtype:              "SKL24",
	HWInventoryByFRUType: "HWInvByFRUProcessor",
	HMSProcessorFRUInfo: &rf.ProcessorFRUInfoRF{
		InstructionSet:        "x86-64",
		Manufacturer:          "Intel",
		MaxSpeedMHz:           json.Number("2600"),
		Model:                 "Intel(R) Xeon(R) CPU E5-2623 v4 @ 2.60GHz",
		ProcessorArchitecture: "x86",
		ProcessorId: rf.ProcessorIdRF{
			EffectiveFamily:         "6",
			EffectiveModel:          "79",
			IdentificationRegisters: "263921",
			MicrocodeInfo:           "184549399",
			Step:                    "1",
			VendorID:                "GenuineIntel",
		},
		ProcessorType: "CPU",
		TotalCores:    json.Number("24"),
		TotalThreads:  json.Number("48"),
	},
}

var ProcHWInvByLoc1 = sm.HWInvByLoc{
	ID:                        "x0c0s0b0n0p0",
	Type:                      "Processor",
	Ordinal:                   0,
	Status:                    "Populated",
	HWInventoryByLocationType: "HWInvByLocProcessor",
	HMSProcessorLocationInfo: &rf.ProcessorLocationInfoRF{
		Id:          "CPU1",
		Name:        "Processor",
		Description: "Socket 1 Processor",
		Socket:      "CPU 1",
	},
	PopulatedFRU: &ProcHWInvByFRU1,
}

var ProcHWInvByFRU2 = sm.HWInvByFRU{
	FRUID:                "HOW-TO-ID-CPUS-FROM-REDFISH-IF-AT-ALL2",
	Type:                 "Processor",
	Subtype:              "SKL24",
	HWInventoryByFRUType: "HWInvByFRUProcessor",
	HMSProcessorFRUInfo: &rf.ProcessorFRUInfoRF{
		InstructionSet:        "x86-64",
		Manufacturer:          "Intel",
		MaxSpeedMHz:           json.Number("2600"),
		Model:                 "Intel(R) Xeon(R) CPU E5-2623 v4 @ 2.60GHz",
		ProcessorArchitecture: "x86",
		ProcessorId: rf.ProcessorIdRF{
			EffectiveFamily:         "6",
			EffectiveModel:          "79",
			IdentificationRegisters: "263921",
			MicrocodeInfo:           "184549399",
			Step:                    "1",
			VendorID:                "GenuineIntel",
		},
		ProcessorType: "CPU",
		TotalCores:    json.Number("24"),
		TotalThreads:  json.Number("48"),
	},
}

var ProcHWInvByLoc2 = sm.HWInvByLoc{
	ID:                        "x0c0s0b0n0p1",
	Type:                      "Processor",
	Ordinal:                   1,
	Status:                    "Populated",
	HWInventoryByLocationType: "HWInvByLocProcessor",
	HMSProcessorLocationInfo: &rf.ProcessorLocationInfoRF{
		Id:          "CPU2",
		Name:        "Processor",
		Description: "Socket 2 Processor",
		Socket:      "CPU 2",
	},
	PopulatedFRU: &ProcHWInvByFRU2,
}

// Memory for node (two)
var MemHWInvByFRU1 = sm.HWInvByFRU{
	FRUID:                "MFR-PARTNUMBER-SERIALNUMBER1",
	Type:                 "Memory",
	Subtype:              "DIMM2400G32",
	HWInventoryByFRUType: "HWInvByFRUMemory",
	HMSMemoryFRUInfo: &rf.MemoryFRUInfoRF{
		BaseModuleType:    "RDIMM",
		BusWidthBits:      json.Number("72"),
		CapacityMiB:       json.Number("32768"),
		DataWidthBits:     json.Number("64"),
		ErrorCorrection:   "MultiBitECC",
		Manufacturer:      "Micron",
		MemoryType:        "DRAM",
		MemoryDeviceType:  "DDR4",
		OperatingSpeedMhz: json.Number("2400"),
		PartNumber:        "XYZ-123-1232",
		RankCount:         json.Number("2"),
		SerialNumber:      "sn12344567689",
	},
}

var MemHWInvByLoc1 = sm.HWInvByLoc{
	ID:                        "x0c0s0b0n0d0",
	Type:                      "Memory",
	Ordinal:                   0,
	Status:                    "Populated",
	HWInventoryByLocationType: "HWInvByLocMemory",
	HMSMemoryLocationInfo: &rf.MemoryLocationInfoRF{
		Id:   "DIMM1",
		Name: "DIMM Slot 1",
		MemoryLocation: rf.MemoryLocationRF{
			Socket:           json.Number("1"),
			MemoryController: json.Number("1"),
			Channel:          json.Number("1"),
			Slot:             json.Number("1"),
		},
	},
	PopulatedFRU: &MemHWInvByFRU1,
}

var MemHWInvByFRU2 = sm.HWInvByFRU{
	FRUID:                "MFR-PARTNUMBER-SERIALNUMBER2",
	Type:                 "Memory",
	Subtype:              "DIMM2400G32",
	HWInventoryByFRUType: "HWInvByFRUMemory",
	HMSMemoryFRUInfo: &rf.MemoryFRUInfoRF{
		BaseModuleType:    "RDIMM",
		BusWidthBits:      json.Number("72"),
		CapacityMiB:       json.Number("32768"),
		DataWidthBits:     json.Number("64"),
		ErrorCorrection:   "MultiBitECC",
		Manufacturer:      "Micron",
		MemoryType:        "DRAM",
		MemoryDeviceType:  "DDR4",
		OperatingSpeedMhz: json.Number("2400"),
		PartNumber:        "XYZ-123-1232",
		RankCount:         json.Number("2"),
		SerialNumber:      "sn12344567680",
	},
}

var MemHWInvByLoc2 = sm.HWInvByLoc{
	ID:                        "x0c0s0b0n0d1",
	Type:                      "Memory",
	Ordinal:                   1,
	Status:                    "Populated",
	HWInventoryByLocationType: "HWInvByLocMemory",
	HMSMemoryLocationInfo: &rf.MemoryLocationInfoRF{
		Id:   "DIMM2",
		Name: "DIMM Slot 2",
		MemoryLocation: rf.MemoryLocationRF{
			Socket:           json.Number("2"),
			MemoryController: json.Number("2"),
			Channel:          json.Number("2"),
			Slot:             json.Number("2"),
		},
	},
	PopulatedFRU: &MemHWInvByFRU2,
}

var HWInvByLocArray1 = []*sm.HWInvByLoc{
	&NodeHWInvByLoc1,
	&ProcHWInvByLoc1,
	&ProcHWInvByLoc2,
	&MemHWInvByLoc1,
	&MemHWInvByLoc2,
}

var HWInvByFRUArray1 = []*sm.HWInvByFRU{
	&NodeHWInvByFRU1,
	&ProcHWInvByFRU1,
	&ProcHWInvByFRU2,
	&MemHWInvByFRU1,
	&MemHWInvByFRU2,
}

// NewSystemHWInventory() is destructive to parent structs.
// Make copies to use with creating example SystemHWInventory
// structs so we leave the originals intact
var NodeHWInvByLocQuery1 = NodeHWInvByLoc1

var HWInvByLocQueryArray1 = []*sm.HWInvByLoc{
	&NodeHWInvByLocQuery1,
	&ProcHWInvByLoc1,
	&ProcHWInvByLoc2,
	&MemHWInvByLoc1,
	&MemHWInvByLoc2,
}

var HWInvByLocQuery, _ = sm.NewSystemHWInventory(HWInvByLocQueryArray1, "s0", sm.HWInvFormatNestNodesOnly)

var HWInvByLocQueryFF, _ = sm.NewSystemHWInventory(HWInvByLocQueryArray1, "s0", sm.HWInvFormatFullyFlat)

var HWInvByLocQueryProcArray = []*sm.HWInvByLoc{
	&ProcHWInvByLoc1,
	&ProcHWInvByLoc2,
}
var HWInvByLocQueryProc, _ = sm.NewSystemHWInventory(HWInvByLocQueryProcArray, "x0c0s0b0n0", sm.HWInvFormatNestNodesOnly)
