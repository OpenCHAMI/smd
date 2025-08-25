package schemas

type EthernetInterface struct {
	URI         string `json:"uri,omitempty"`         // URI of the interface
	MAC         string `json:"mac,omitempty"`         // MAC address of the interface
	IP          string `json:"ip,omitempty"`          // IP address of the interface
	Name        string `json:"name,omitempty"`        // Name of the interface
	Description string `json:"description,omitempty"` // Description of the interface
	Enabled     bool   `json:"enabled,omitempty"`     // Whether interface is enabled
}

type NetworkAdapter struct {
	URI          string `json:"uri,omitempty"`          // URI of the adapter
	Manufacturer string `json:"manufacturer,omitempty"` // Manufacturer of the adapter
	Name         string `json:"name,omitempty"`         // Name of the adapter
	Model        string `json:"model,omitempty"`        // Model of the adapter
	Serial       string `json:"serial,omitempty"`       // Serial number of the adapter
	Description  string `json:"description,omitempty"`  // Description of the adapter
}

type NetworkInterface struct {
	URI         string         `json:"uri,omitempty"`         // URI of the interface
	Name        string         `json:"name,omitempty"`        // Name of the interface
	Description string         `json:"description,omitempty"` // Description of the interface
	Adapter     NetworkAdapter `json:"adapter,omitempty"`     // Adapter of the interface
}

type ResourceID struct {
	OdataID string `json:"@odata.id"`
}

type ActionReset struct {
	AllowableValues []string `json:"ResetType@Redfish.AllowableValues"`
	RFActionInfo    string   `json:"@Redfish.ActionInfo"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}

type ComputerSystemActions struct {
	ComputerSystemReset ActionReset `json:"#ComputerSystem.Reset"`
}

type PowerControl struct {
	ResourceID
	MemberId           string        `json:"MemberId,omitempty"`
	Name               string        `json:"Name,omitempty"`
	PowerCapacityWatts int           `json:"PowerCapacityWatts,omitempty"`
	RelatedItem        []*ResourceID `json:"RelatedItem,omitempty"`
}

type Links struct {
	Chassis  []string `json:"chassis,omitempty"`
	Managers []string `json:"managers,omitempty"`
}

type Power struct {
	State           string   `json:"state,omitempty"`
	PowerControlIDS []string `json:"power_control_ids,omitempty"`
}

type InventoryDetail struct {
	URI                  string              `json:"uri,omitempty"`                  // URI of the BMC
	UUID                 string              `json:"uuid,omitempty"`                 // UUID of Node
	Manufacturer         string              `json:"manufacturer,omitempty"`         // Manufacturer of the Node
	SystemType           string              `json:"system_type,omitempty"`          // System type of the Node
	Name                 string              `json:"name,omitempty"`                 // Name of the Node
	Model                string              `json:"model,omitempty"`                // Model of the Node
	Serial               string              `json:"serial,omitempty"`               // Serial number of the Node
	BiosVersion          string              `json:"bios_version,omitempty"`         // Version of the BIOS
	EthernetInterfaces   []EthernetInterface `json:"ethernet_interfaces,omitempty"`  // Ethernet interfaces of the Node
	NetworkInterfaces    []NetworkInterface  `json:"network_interfaces,omitempty"`   // Network interfaces of the Node
	Power                *Power              `json:"power,omitempty"`                // Power state of the Node
	ProcessorCount       int                 `json:"processor_count,omitempty"`      // Processors of the Node
	ProcessorType        string              `json:"processor_type,omitempty"`       // Processor type of the Node
	MemoryTotal          float32             `json:"memory_total,omitempty"`         // Total memory of the Node in Gigabytes
	TrustedModules       []string            `json:"trusted_modules,omitempty"`      // Trusted modules of the Node
	TrustedComponents    []string            `json:"trusted_components,omitempty"`   // Trusted components of the Chassis
	Chassis_SKU          string              `json:"chassis_sku,omitempty"`          // SKU of the Chassis
	Chassis_Serial       string              `json:"chassis_serial,omitempty"`       // Serial number of the Chassis
	Chassis_AssetTag     string              `json:"chassis_asset_tag,omitempty"`    // Asset tag of the Chassis
	Chassis_Manufacturer string              `json:"chassis_manufacturer,omitempty"` // Manufacturer of the Chassis
	Chassis_Model        string              `json:"chassis_model,omitempty"`        // Model of the Chassis
	OdataId              string              `json:"@odata.id,omitempty"`            // OData ID for the computer system
	PowerURL             string              `json:"PowerURL,omitempty"`             // URL for power control
	PowerControl         []*PowerControl     `json:"PowerControl,omitempty"`         // Power control actions data
	Actions              []string            `json:"actions,omitempty"`              // Actions for the hardware
	Links                *Links              `json:"links,omitempty"`                // Links to related resources
}
