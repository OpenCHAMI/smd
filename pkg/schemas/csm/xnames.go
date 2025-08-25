package csm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/invopop/jsonschema"
)

type NodeXname struct {
	Value string
}

func (n NodeXname) Cabinet() (int, error) {
	if n.Value == "" {
		return 0, fmt.Errorf("node does not have an XName")
	}
	return extractXNameComponents(n.Value).Cabinet, nil
}

func (n NodeXname) Chassis() (int, error) {
	if n.Value == "" {
		return 0, fmt.Errorf("node does not have an XName")
	}
	return extractXNameComponents(n.Value).Chassis, nil
}

func (n NodeXname) Slot() (int, error) {
	if n.Value == "" {
		return 0, fmt.Errorf("node does not have an XName")
	}
	return extractXNameComponents(n.Value).Slot, nil
}

func (n NodeXname) NodePosition() (int, error) {
	if n.Value == "" {
		return 0, fmt.Errorf("node does not have an XName")
	}
	return extractXNameComponents(n.Value).NodePosition, nil
}

func (n NodeXname) BMCPosition() (int, error) {
	if n.Value == "" {
		return 0, fmt.Errorf("node does not have an XName")
	}
	return extractXNameComponents(n.Value).BMCPosition, nil
}

func (n NodeXname) String() string {
	return n.Value
}

type XNameComponents struct {
	Cabinet      int    `json:"cabinet"`
	Chassis      int    `json:"chassis"`
	Slot         int    `json:"slot"`
	BMCPosition  int    `json:"bmc_position"`
	NodePosition int    `json:"node_position"`
	Type         string `json:"type"` // 'n' for node, 'b' for BMC
}

func extractXNameComponents(xname string) XNameComponents {
	var components XNameComponents
	_, err := fmt.Sscanf(xname, "x%dc%ds%db%dn%d", &components.Cabinet, &components.Chassis, &components.Slot, &components.BMCPosition, &components.NodePosition)
	if err == nil {
		components.Type = "n"
		return components
	}
	_, err = fmt.Sscanf(xname, "x%dc%ds%db%d", &components.Cabinet, &components.Chassis, &components.Slot, &components.BMCPosition)
	if err == nil {
		components.Type = "b"
		return components
	}
	return components
}

func (NodeXname) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Title:       "NodeXName",
		Description: "XName for a compute node",
		Pattern:     `^x(\d{3,5})c(\d{1,3})s(\d{1,3})b(\d{1,3})n(\d{1,3})$`,
	}
}

func (xname NodeXname) MarshalJSON() ([]byte, error) {
	return json.Marshal(xname.Value)
}

func (xname *NodeXname) UnmarshalJSON(data []byte) error {
	xname.Value = string(data)
	// Remove quotation marks if they exist
	if len(xname.Value) >= 2 && xname.Value[0] == '"' && xname.Value[len(xname.Value)-1] == '"' {
		xname.Value = xname.Value[1 : len(xname.Value)-1]
	}
	return nil
}

func (xname NodeXname) Valid() (bool, error) {
	nodeXnameRegex := regexp.MustCompile(`^x(?P<cabinet>\d{3,5})c(?P<chassis>\d{1,3})s(?P<slot>\d{1,3})b(?P<bmc>\d{1,3})n(?P<node>\d{1,3})$`)
	if !nodeXnameRegex.MatchString(xname.Value) {
		return false, fmt.Errorf("XName does not match regex")
	}

	// Extract the named groups
	match := nodeXnameRegex.FindStringSubmatch(xname.Value)
	result := make(map[string]string)
	for i, name := range nodeXnameRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			result[name] = match[i]
		}
	}

	// Convert and check chassis number
	chassis, err := strconv.Atoi(result["chassis"])
	if err != nil {
		return false, fmt.Errorf("chassis is not a valid number: %s", result["chassis"])
	}
	if chassis >= 256 {
		return false, fmt.Errorf("chassis number %d exceeds the maximum allowed value of 255", chassis)
	}

	return true, nil
}

// XnameSliceString converts a slice of NodeCollectionType to a slice of strings.
func XnameSliceString(slice []NodeXname) []string {
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		strSlice[i] = v.String()
	}
	return strSlice
}

func NewNodeXname(xname string) NodeXname {
	return NodeXname{Value: xname}
}

type BMCXname struct {
	Value string
}

func NewBMCXname(xname string) BMCXname {
	return BMCXname{Value: xname}
}

func (b BMCXname) String() string {
	return b.Value
}

func (b BMCXname) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Title:       "BMCXName",
		Description: "XName for a BMC",
		Pattern:     `^x(\d{3,5})c(\d{1,3})s(\d{1,3})b(\d{1,3})$`,
	}
}

func (b BMCXname) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Value)
}

func (b *BMCXname) UnmarshalJSON(data []byte) error {
	b.Value = string(data)
	// Remove quotation marks if they exist
	if len(b.Value) >= 2 && b.Value[0] == '"' && b.Value[len(b.Value)-1] == '"' {
		b.Value = b.Value[1 : len(b.Value)-1]
	}
	return nil
}

func (b BMCXname) Valid() (bool, error) {
	bmcXnameRegex := regexp.MustCompile(`^x(?P<cabinet>\d{3,5})c(?P<chassis>\d{1,3})s(?P<slot>\d{1,3})b(?P<bmc>\d{1,3})$`)
	if !bmcXnameRegex.MatchString(b.Value) {
		return false, fmt.Errorf("XName does not match regex")
	}

	// Extract the named groups
	match := bmcXnameRegex.FindStringSubmatch(b.Value)
	result := make(map[string]string)
	for i, name := range bmcXnameRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			result[name] = match[i]
		}
	}

	// Convert and check chassis number
	chassis, err := strconv.Atoi(result["chassis"])
	if err != nil {
		return false, fmt.Errorf("chassis is not a valid number: %s", result["chassis"])
	}
	if chassis >= 256 {
		return false, fmt.Errorf("chassis number %d exceeds the maximum allowed value of 255", chassis)
	}
	return true, nil
}

func IsValidBMCXName(xname string) bool {
	// Compile the regular expression. This is the pattern from your requirement.
	re := regexp.MustCompile(`^x(?P<cabinet>\d{3,5})c(?P<chassis>\d{1,3})s(?P<slot>\d{1,3})b(?P<bmc>\d{1,3})$`)

	// Use FindStringSubmatch to capture the parts of the xname.
	matches := re.FindStringSubmatch(xname)
	if matches == nil {
		return false
	}

	// Since the cabinet can go up to 100,000 and others up to 255, we need to check these values.
	// The order of subexpressions in matches corresponds to the groups in the regex.
	cabinet, _ := strconv.Atoi(matches[1])
	chassis, _ := strconv.Atoi(matches[2])
	slot, _ := strconv.Atoi(matches[3])
	bmc, _ := strconv.Atoi(matches[4])

	if cabinet > 100000 || chassis >= 256 || slot >= 256 || bmc >= 256 {
		return false
	}

	return true
}
