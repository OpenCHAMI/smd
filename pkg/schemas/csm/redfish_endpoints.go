package csm

import (
	"time"

	"github.com/google/uuid"
)

type DiscoveryInfo struct {
	LastAttempt    time.Time `json:"LastAttempt,omitempty" jsonschema:"description=The time the last discovery attempt took place,format=date-time,readOnly=true"`
	LastStatus     string    `json:"LastStatus,omitempty" jsonschema:"description=Describes the outcome of the last discovery attempt,enum=EndpointInvalid,enum=EPResponseFailedDecode,enum=HTTPsGetFailed,enum=NotYetQueried,enum=VerificationFailed,enum=ChildVerificationFailed,enum=DiscoverOK,readOnly=true"`
	RedfishVersion string    `json:"RedfishVersion,omitempty" jsonschema:"description=Version of Redfish as reported by the RF service root,readOnly=true"`
}

type RedfishDiscovery struct {
	EntrypointID string          `json:"EntrypointID,omitempty" jsonschema:"description=ID of the entrypoint that was used to discover the endpoint"`
	UID          uuid.UUID       `json:"UID,omitempty" jsonschema:"$ref=#/definitions/UUID.1.0.0"`
	URI          string          `json:"EndpointID,omitempty" jsonschema:"description=ID of the endpoint that was discovered"`
	Attempted    time.Time       `json:"Attempted,omitempty" jsonschema:"description=Time the discovery was started,format=date-time"`
	Completed    time.Time       `json:"Completed,omitempty" jsonschema:"description=Time the discovery was completed,format=date-time"`
	Status       string          `json:"Status,omitempty" jsonschema:"description=Describes the outcome of the discovery attempt,enum=EndpointInvalid,enum=EPResponseFailedDecode,enum=HTTPsGetFailed,enum=NotYetQueried,enum=VerificationFailed,enum=ChildVerificationFailed,enum=DiscoverOK"`
	Payload      RedfishEndpoint `json:"Payload,omitempty" jsonschema:"description=The discovered endpoint"`
}

type RedfishEndpoint struct {
	ID                 string        `json:"ID" jsonschema:"description=HMS Logical component type e.g. NodeBMC, ChassisBMC.,$ref=#/definitions/HMSType.1.0.0"`
	Type               ComponentType `json:"Type,omitempty"`
	Name               string        `json:"Name,omitempty" jsonschema:"description=This is an arbitrary, user-provided name for the endpoint. It can describe anything that is not captured by the ID/xname."`
	Hostname           string        `json:"Hostname,omitempty" jsonschema:"description=Hostname of the endpoint's FQDN, will always be the host portion of the fully-qualified domain name. Note that the hostname should normally always be the same as the ID field (i.e. xname) of the endpoint."`
	Domain             string        `json:"Domain,omitempty" jsonschema:"description=Domain of the endpoint's FQDN. Will always match remaining non-hostname portion of fully-qualified domain name (FQDN)."`
	FQDN               string        `json:"FQDN,omitempty" jsonschema:"description=Fully-qualified domain name of RF endpoint on management network. This is not writable because it is made up of the Hostname and Domain."`
	Enabled            bool          `json:"Enabled,omitempty" jsonschema:"description=To disable a component without deleting its data from the database, can be set to false,example=true"`
	URI                string        `json:"URI,omitempty" jsonschema:"description=URI of the Redfish service root"`
	UID                uuid.UUID     `json:"UUID,omitempty" jsonschema:"$ref=#/definitions/UUID.1.0.0"`
	User               string        `json:"User,omitempty" jsonschema:"description=Username to use when interrogating endpoint"`
	Password           string        `json:"Password,omitempty" jsonschema:"description=Password to use when interrogating endpoint, normally suppressed in output."`
	UseSSDP            bool          `json:"UseSSDP,omitempty" jsonschema:"description=Whether to use SSDP for discovery if the EP supports it."`
	MacRequired        bool          `json:"MacRequired,omitempty" jsonschema:"description=Whether the MAC must be used (e.g. in River) in setting up geolocation info so the endpoint's location in the system can be determined. The MAC does not need to be provided when creating the endpoint if the endpoint type can arrive at a geolocated hostname on its own."`
	MACAddr            string        `json:"MACAddr,omitempty" jsonschema:"description=This is the MAC on the of the Redfish Endpoint on the management network, i.e. corresponding to the FQDN field's Ethernet interface where the root service is running. Not the HSN MAC. This is a MAC address in the standard colon-separated 12 byte hex format.,pattern=^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$,example=ae:12:e2:ff:89:9d"`
	IPAddress          string        `json:"IPAddress,omitempty" jsonschema:"description=This is the IP of the Redfish Endpoint on the management network, i.e. corresponding to the FQDN field's Ethernet interface where the root service is running. This may be IPv4 or IPv6,example=10.254.2.10"`
	RediscoverOnUpdate bool          `json:"RediscoverOnUpdate,omitempty" jsonschema:"description=Trigger a rediscovery when endpoint info is updated."`
	TemplateID         string        `json:"TemplateID,omitempty" jsonschema:"description=Links to a discovery template defining how the endpoint should be discovered."`
	DiscoveryInfo      DiscoveryInfo `json:"DiscoveryInfo,omitempty" jsonschema:"description=Contains info about the discovery status of the given endpoint,readOnly=true"`
}
