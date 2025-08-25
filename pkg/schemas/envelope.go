package schemas

// Envelope structure for schema versioning
type Envelope struct {
	SchemaID string      `json:"schema_id"`
	Version  string      `json:"version"`
	Payload  interface{} `json:"payload"`
}
