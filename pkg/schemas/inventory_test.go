package schemas

import (
	"encoding/json"
	"testing"

	"os"
	"path/filepath"

	"log"

	"github.com/invopop/jsonschema"
	"github.com/openchami/schemas/schemas/csm"
)

type InventoryRequest struct {
	Header               Envelope          `json:"header"`
	InventoryDetailArray []InventoryDetail `json:"inventory_detail_array"`
}

func generateAndWriteSchemas(path string) {
	schemas := map[string]interface{}{

		"Component.json":              &csm.Component{},
		"RedfishEndpoint.json":        &csm.RedfishEndpoint{},
		"InventoryDetailRequest.json": &InventoryRequest{},
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatal("Failed to create schema directory")
	}

	for filename, model := range schemas {
		schema := jsonschema.Reflect(model)
		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			log.Fatal("Failed to generate JSON schema")
		}
		fullpath := filepath.Join(path, filename)
		if err := os.WriteFile(fullpath, data, 0644); err != nil {
			log.Fatal("Failed to write JSON schema to file")
		}
		log.Println("Schema written")
	}
}

func TestGenerateAndWriteSchemas(t *testing.T) {
	var (
		path    = "jsonschemas"
		schemas = map[string]interface{}{
			"Component.json":              &csm.Component{},
			"RedfishEndpoint.json":        &csm.RedfishEndpoint{},
			"InventoryDetailRequest.json": &InventoryRequest{},
		}
	)

	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to find workding dir. err: %s\n", err)
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("Failed to create schema directory: %s/%s", dir, path)
	} else {
		log.Printf("created: %s/%s\n", dir, path)
	}

	for filename, model := range schemas {
		schema := jsonschema.Reflect(model)
		data, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			t.Fatal("Failed to generate JSON schema")
		}
		fullpath := filepath.Join(path, filename)
		if err := os.WriteFile(fullpath, data, 0644); err != nil {
			t.Fatalf("Failed to write JSON schema to file: %s", fullpath)
		}
		log.Println("Schema written:", fullpath)
	}

	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("Failed to remove schema directory: %s/%s", dir, path)
	} else {
		log.Printf("Removed: %s/%s\n", dir, path)
	}
}
