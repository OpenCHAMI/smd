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

package sharedtest

import (
	"strings"
)

////////////////////////////////////////////////////////////////////////////
// Event customization functions that will fill in values at runtime so
// the same template can be used for many different tests
////////////////////////////////////////////////////////////////////////////

// Any function that can take an event string, transform it, and return it
// can be used as an argument to GenEvent() below
type EventTemplateArg func(string) string

// Replace xname in context, i.e. the xname of the Chassis/Node/RouterBMC
// i.e. 'x0c0s0b0'
func EpID(epXNameID string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%XNAME%", epXNameID, -1)
		return enew
	}
}

// Replace Redfish ID wildcard with rfId for endpont with xname epXName
// i.e. call with 'Node0'
func RfId(rfId string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%REDFISH_ID%", rfId, -1)
		return enew
	}
}

// Replace path portion i.e. /redfish/v1/XXXX/... to switch type referenced,
// if e contains a wildcard for it.
func Path(path string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%PATH%", path, -1)
		return enew
	}
}

// Replace Severity field in event if there is a wildcard for it.
func Severity(severity string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%SEVERITY%", severity, -1)
		return enew
	}
}

// Replace full MessageId field in event if there is a wildcard for it.
func MsgId(messageId string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%MESSAGE_ID%", messageId, -1)
		return enew
	}
}

// Replace Registry/Prefix portion of MessageId field in event if there is
// a wildcard for it.  Must include trailing separator (.).
func Registry(regId string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%REGISTRY_ID%", regId, -1)
		return enew
	}
}

// The above functions can be passed as a variable number of arguments,
// with only the needed ones required.
func GenEvent(e string, opts ...EventTemplateArg) string {
	enew := e
	for _, opt := range opts {
		enew = opt(enew)
	}
	return enew
}

////////////////////////////////////////////////////////////////////////////
// Event templates - Cray firmware
////////////////////////////////////////////////////////////////////////////

// Defaults if left blank in below.
const (
	CrayPwrStateMessageId = "ResourceEvent.1.0.ResourcePowerStateChanged"
	CrayPwrStatePath      = "Chassis"
	CrayPwrStateSeverity  = "OK"
)

// Same as GenEvent above, but with defaults if there are template fields
// that the event template specified but were not filled in by the opts.
func GenEventResource(e string, opts ...EventTemplateArg) string {
	enew := GenEvent(e, opts...)
	// Default if there are still template values that were not set.
	enew = strings.Replace(enew, "%PATH%", CrayPwrStatePath, -1)
	enew = strings.Replace(enew, "%MESSAGE_ID%", CrayPwrStateMessageId, -1)
	enew = strings.Replace(enew, "%SEVERITY%", CrayPwrStateSeverity, -1)
	return enew
}

// Replace PowerState if that is a field in the message args with a
// template value in e.
func PowState(pow string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%POWERSTATE%", pow, -1)
		return enew
	}
}

var EventCrayOnOKChassis = `{
   "Context": "%XNAME%:PowerState_tests",
   "Events": [{
       "@odata.id": "/redfish/v1/EventService/Events/140",
       "EventId": "140",
       "Message": "The power state of resource /redfish/v1/Chassis/Enclosure has changed to type On.",
       "MessageArgs": [
           "/redfish/v1/Chassis/Enclosure",
           "On"
       ],
       "MessageId": "ResourceEvent.1.0.ResourcePowerStateChanged",
       "OriginOfCondition": {
           "@odata.id": "/redfish/v1/Chassis/Enclosure"
       },
       "Severity": "OK"
   }]
}
`

var EventCrayOffOKChassis = `{
   "Context": "%XNAME%:PowerState_tests",
   "Events": [{
       "@odata.id": "/redfish/v1/EventService/Events/141",
       "EventId": "141",
       "Message": "The power state of resource /redfish/v1/Chassis/Enclosure has changed to type Off.",
       "MessageArgs": [
           "/redfish/v1/Chassis/Enclosure",
           "Off"
       ],
       "MessageId": "ResourceEvent.1.0.ResourcePowerStateChanged",
       "OriginOfCondition": {
           "@odata.id": "/redfish/v1/Chassis/Enclosure"
       },
       "Severity": "OK"
   }]
}
`

// Blade[0-7] for ComputeModule or Perif[0-7] for RouterModule
var EventCrayOnOKSlotX = `{
   "Context": "%XNAME%:PowerState_tests",
   "Events": [{
       "@odata.id": "/redfish/v1/EventService/Events/143",
       "EventId": "143",
       "Message": "The power state of resource /redfish/v1/Chassis/%REDFISH_ID% has changed to type On.",
       "MessageArgs": [
           "/redfish/v1/Chassis/%REDFISH_ID%",
           "On"
       ],
       "MessageId": "ResourceEvent.1.0.ResourcePowerStateChanged",
       "OriginOfCondition": {
           "@odata.id": "/redfish/v1/Chassis/%REDFISH_ID%"
       },
       "Severity": "OK"
   }]
}
`

// Blade[0-7] for ComputeModule or Perif[0-7] for RouterModule
var EventCrayOffOKSlotX = `{
   "Context": "%XNAME%:PowerState_tests",
   "Events": [{
       "@odata.id": "/redfish/v1/EventService/Events/144",
       "EventId": "144",
       "Message": "The power state of resource /redfish/v1/Chassis/%REDFISH_ID% has changed to type Off.",
       "MessageArgs": [
           "/redfish/v1/Chassis/%REDFISH_ID%",
           "Off"
       ],
       "MessageId": "ResourceEvent.1.0.ResourcePowerStateChanged",
       "OriginOfCondition": {
           "@odata.id": "/redfish/v1/Chassis/%REDFISH_ID%"
       },
       "Severity": "OK"
   }]
}
`

// Pretty much everything is a wildcard here.  Can be used for Systems nodes,
// change format of MessageId, etc.   Can set usual defaults with
// GenEventResource and then override the ones of interest.
var EventCrayXXPathX = `{
   "Context": "%XNAME%:PowerState_tests",
   "Events": [{
       "@odata.id": "/redfish/v1/EventService/Events/200",
       "EventId": "200",
       "Message": "The power state of resource /redfish/v1/%PATH%/%REDFISH_ID% has changed to type %POWERSTATE%.",
       "MessageArgs": [
           "/redfish/v1/%PATH%/%REDFISH_ID%",
           "%POWERSTATE%"
       ],
       "MessageId": "%MESSAGE_ID%",
       "OriginOfCondition": {
           "@odata.id": "/redfish/v1/%PATH%/%REDFISH_ID%"
       },
       "Severity": "%SEVERITY%"
   }]
}
`

////////////////////////////////////////////////////////////////////////////
// Event templates - RTS
////////////////////////////////////////////////////////////////////////////

// Defaults if left blank in below.
const (
	RtsPwrStateMessageId  = "ResourceEvent.1.0.ResourcePowerStateChanged"
	RtsPwrStatePath1      = "PowerEquipment/RackPDUs/"
	RtsPwrStatePathParent = "1"
	RtsPwrStatePath2      = "/Outlets"
	RtsPwrStateSeverity   = "OK"
)

// Same as GenEvent above, but with defaults if there are template fields
// that the event template specified but were not filled in by the opts.
func GenRtsEventResource(e string, opts ...EventTemplateArg) string {
	enew := GenEvent(e, opts...)
	// Default if there are still template values that were not set.
	enew = strings.Replace(enew, "%PATH1%", RtsPwrStatePath1, -1)
	enew = strings.Replace(enew, "%PATH_PARENT%", RtsPwrStatePathParent, -1)
	enew = strings.Replace(enew, "%PATH2%", RtsPwrStatePath2, -1)
	enew = strings.Replace(enew, "%MESSAGE_ID%", RtsPwrStateMessageId, -1)
	enew = strings.Replace(enew, "%SEVERITY%", RtsPwrStateSeverity, -1)
	return enew
}

// Replace first path portion i.e. /redfish/v1/XXXX/... to provided string,
// if e contains a wildcard for it.
func Path1(path string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%PATH1%", path, -1)
		return enew
	}
}

// Replace parent name portion of path i.e. /redfish/v1/path1/XXXX/... to
// provided string, if e contains a wildcard for it.
func PathParent(pathp string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%PATH_PARENT%", pathp, -1)
		return enew
	}
}

// Replace second path portion i.e. /redfish/v1/{path1}/{path_parent}/XXXX to
// provided string, if e contains a wildcard for it.
func Path2(path string) EventTemplateArg {
	return func(e string) string {
		enew := strings.Replace(e, "%PATH2%", path, -1)
		return enew
	}
}

// Pretty much everything is a wildcard here.  Can be used for Systems nodes,
// change format of MessageId, etc.   Can set usual defaults with
// GenEventResource and then override the ones of interest.
var EventRtsXXPathX = `{
   "Context": "%XNAME%:PowerState_tests",
   "Events": [{
       "@odata.id": "/redfish/v1/EventService/Events/200",
       "EventId": "200",
       "Message": "The power state of resource /redfish/v1/%PATH1%%PATH_PARENT%%PATH2%/%REDFISH_ID% has changed to type %POWERSTATE%.",
       "MessageArgs": [
           "/redfish/v1/%PATH1%%PATH_PARENT%%PATH2%/%REDFISH_ID%",
           "%POWERSTATE%"
       ],
       "MessageId": "%MESSAGE_ID%",
       "OriginOfCondition": {
           "@odata.id": "/redfish/v1/%PATH1%%PATH_PARENT%%PATH2%/%REDFISH_ID%"
       },
       "Severity": "%SEVERITY%"
   }]
}
`

////////////////////////////////////////////////////////////////////////////
// Event templates - Intel firmware - as of 1.90
////////////////////////////////////////////////////////////////////////////

// Defaults if left blank in below.
const (
	IntelRegistryId = "Alert.1.0.0."
	IntelSeverity   = "OK"
)

// Same as GenEvent above, but with defaults if there are template fields
// that the event template specified but were not filled in by the opts.
func GenEventIntel(e string, opts ...EventTemplateArg) string {
	enew := GenEvent(e, opts...)
	// Default if there are still template values that were not set.
	enew = strings.Replace(enew, "%REGISTRY_ID%", IntelRegistryId, -1)
	enew = strings.Replace(enew, "%SEVERITY%", IntelSeverity, -1)
	return enew
}

// This is a specific event type for a System power up event
var EventIntelSystemOnOK = `{
	"@odata.context": "/redfish/v1/$metadata#Event.Event",
	"@odata.type": "#Event.v1_2_1.Event",
	"@odata.id": "/redfish/v1/EventService/Events",
	"Id": "Event4",
	"Name": "Event Array",
	"Description": "Events",
	"Events":[{
			"EventType": "Alert",
			"EventId": "Event4",
			"Severity": "OK",
			"EventTimestamp": "2019-03-05T20:49:20+00:00",
			"Message": "System Power Turned On.",
			"MessageId": "%REGISTRY_ID%SystemPowerOn",
			"MessageArgs": [null],
			"OriginOfCondition": {
				"@odata.id": "/redfish/v1/Systems/%REDFISH_ID%"
			}
		}],
	"Context": "%XNAME%:whatever"
}
`

// This is a specific event type for a System power down event
var EventIntelSystemOffOK = `{
	"@odata.context": "/redfish/v1/$metadata#Event.Event",
	"@odata.type": "#Event.v1_2_1.Event",
	"@odata.id": "/redfish/v1/EventService/Events",
	"Id": "Event3",
	"Name": "Event Array",
	"Description": "Events",
	"Events": [{
			"EventType": "Alert",
			"EventId": "Event3",
			"Severity": "%SEVERITY%",
			"EventTimestamp": "2019-03-05T20:41:17+00:00",
			"Message": "System Power Turned Off.",
			"MessageId": "%REGISTRY_ID%SystemPowerOff",
			"MessageArgs": [null],
			"OriginOfCondition": {
				"@odata.id": "/redfish/v1/Systems/%REDFISH_ID%"
			}
		}],
	"Context": "%XNAME%:whatever"
}
`

// This is also interesting because it is the only thing we
// get on an immediate power cycle, e.g. pressing the reset button.
var EventIntelDriveInserted = `{
	"@odata.context": "/redfish/v1/$metadata#Event.Event",
	"@odata.type": "#Event.v1_2_1.Event",
	"@odata.id": "/redfish/v1/EventService/Events",
	"Id": "Event8",
	"Name": "Event Array",
	"Description": "Events",
	"Events": [{
			"EventType": "Alert",
			"EventId": "Event8",
			"Severity": "%SEVERITY%",
			"EventTimestamp": "2019-03-06T18:21:25+00:00",
			"Message": "Drive HDD255 has been inserted.",
			"MessageId": "Alert.1.0.0.DriveInserted",
			"MessageArgs": ["HDD255"],
			"OriginOfCondition": {
				"@odata.id": "/redfish/v1/Systems/%REDFISH_ID%/Storage/1/Drives/HDD255"
			}
		}],
	"Context": "%XNAME%:whatever"
}
`

////////////////////////////////////////////////////////////////////////////
// Event templates - Gigabyte firmware
////////////////////////////////////////////////////////////////////////////

// Defaults if left blank in below.
const (
	GigabyteRegistryId = "EventLog.1.0."
	GigabyteSeverity   = "OK"
)

// Same as GenEvent above, but with defaults if there are template fields
// that the event template specified but were not filled in by the opts.
func GenEventGigabyte(e string, opts ...EventTemplateArg) string {
	enew := GenEvent(e, opts...)
	// Default if there are still template values that were not set.
	enew = strings.Replace(enew, "%REGISTRY_ID%", GigabyteRegistryId, -1)
	enew = strings.Replace(enew, "%SEVERITY%", GigabyteSeverity, -1)
	enew = strings.Replace(enew, "%XNAME%", "", -1)
	enew = strings.Replace(enew, "%REDFISH_ID%", "Self", -1)
	return enew
}

// This is a specific event type for a System power up event
var EventGigabyteSystemOK = `{
	"@odata.context": "/redfish/v1/$metadata#EventService.EventService",
	"@odata.type": "#EventService.1.0.0.Event",
	"@odata.id": "/redfish/v1/EventService/Events/1",
	"Id": "1",
	"Name": "Event Array",
	"Events": [{
			"Context": "%XNAME%:whatever",
			"Created": "2019-08-22T22:57:33+00:00",
			"EntryCode": "Informational",
			"EntryType": "Event",
			"EventId": "/redfish/v1/Systems/%REDFISH_ID% - 1566514653",
			"EventTimestamp": "2019-08-22T22:57:33+00:00",
			"EventType": "Alert",
			"Message": "A condition exists on the resource at /redfish/v1/Systems/%REDFISH_ID% which requires attention",
			"MessageId": "%REGISTRY_ID%Alert",
			"Name": "Log entry 18",
			"OriginOfCondition": {
					"@odata.id": "/redfish/v1/Systems/%REDFISH_ID%"
			},
			"Severity": "%SEVERITY%"
		}]
}
`

////////////////////////////////////////////////////////////////////////////
// Event templates - HPE iLo firmware
////////////////////////////////////////////////////////////////////////////

// Defaults if left blank in below.
const (
	HPEiLORegistryId = "iLOEvents.2.1."
	HPEiLOSeverity   = "OK"
)

// Same as GenEvent above, but with defaults if there are template fields
// that the event template specified but were not filled in by the opts.
func GenEventHPEiLO(e string, opts ...EventTemplateArg) string {
	enew := GenEvent(e, opts...)
	// Default if there are still template values that were not set.
	enew = strings.Replace(enew, "%REGISTRY_ID%", HPEiLORegistryId, -1)
	return enew
}

// This is a specific event type for a System power up event
var EventHPEiLOServerPoweredOn = `{
	"@odata.context": "/redfish/v1/$metadata#Event.Event",
	"@odata.type": "#Event.v1_0_0.Event",
	"Events":[{
			"Context": "%XNAME%:whatever",
			"EventId": "558851bc-c386-d8d0-70a9-900e4a067d84",
			"EventTimestamp": "2021-04-13T03:29:51Z",
			"EventType": "Alert",
			"MemberId": "0",
			"MessageId": "%REGISTRY_ID%ServerPoweredOn",
			"Oem": {
				"Hpe": {
					"@odata.context": "/redfish/v1/$metadata#HpeEvent.HpeEvent",
					"@odata.type": "#HpeEvent.v2_1_0.HpeEvent",
					"CorrelatedEventNumber": 133896,
					"CorrelatedEventTimeStamp": "2021-04-13T03:31:28Z",
					"CorrelatedEventType": "Hpe-iLOEventLog",
					"CorrelatedIndications": [
						"HP:SNMP:1.3.6.1.4.1.232:6:9017:2914671648"
					],
					"Resource": "/redfish/v1/Systems/%REDFISH_ID%"
				}
			},
			"OriginOfCondition": {
				"@odata.id": "/redfish/v1/Systems/%REDFISH_ID%"
			},
			"Severity": "OK"
		}
	],
	"Name": "Events"
}
`

// This is a specific event type for a System power down event
var EventHPEiLOServerPoweredOff = `{
	"@odata.context": "/redfish/v1/$metadata#Event.Event",
	"@odata.type": "#Event.v1_0_0.Event",
	"Events":[{
			"Context": "%XNAME%:whatever",
			"EventId": "558851bc-c386-d8d0-70a9-900e4a067d84",
			"EventTimestamp": "2021-04-13T03:29:51Z",
			"EventType": "Alert",
			"MemberId": "0",
			"MessageId": "%REGISTRY_ID%ServerPoweredOff",
			"Oem": {
				"Hpe": {
					"@odata.context": "/redfish/v1/$metadata#HpeEvent.HpeEvent",
					"@odata.type": "#HpeEvent.v2_1_0.HpeEvent",
					"CorrelatedEventNumber": 133896,
					"CorrelatedEventTimeStamp": "2021-04-13T03:31:28Z",
					"CorrelatedEventType": "Hpe-iLOEventLog",
					"CorrelatedIndications": [
						"HP:SNMP:1.3.6.1.4.1.232:6:9017:2914671648"
					],
					"Resource": "/redfish/v1/Systems/%REDFISH_ID%"
				}
			},
			"OriginOfCondition": {
				"@odata.id": "/redfish/v1/Systems/%REDFISH_ID%"
			},
			"Severity": "OK"
		}
	],
	"Name": "Events"
}
`

////////////////////////////////////////////////////////////////////////////
// Event templates - Foxconn Paradise OpenBmc firmware
////////////////////////////////////////////////////////////////////////////

var EventFoxconnServerPoweredOn = `{
	"@odata.type": "#Event.v1_7_0.Event",
	"Context": "%XNAME%:whatever",
	"Id": "2",
	"Name": "Event Log",
	"Events": [
	  {
		"EventTimeStamp": "2024-05-20T17:57:23+00:00",
		"Message": "Host system DC power is on.",
		"MessageId": "Alert.1.0.0.DCPowerOn",
		"MessageSeverity": "Critical"
	  }
	]
  }
`

var EventFoxconnServerPoweredOff = `{
	"@odata.type": "#Event.v1_7_0.Event",
	"Context": "%XNAME%:whatever",
	"Id": "2",
	"Name": "Event Log",
	"Events": [
	  {
		"EventTimeStamp": "2024-05-20T17:57:23+00:00",
		"Message": "Host system DC power is off.",
		"MessageId": "Alert.1.0.0.DCPowerOff",
		"MessageSeverity": "Critical"
	  }
	]
  }
`
