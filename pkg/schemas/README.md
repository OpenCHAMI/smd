
# OpenCHAMI Schema Repository

Welcome to the OpenCHAMI Schema Repository! This repository serves as the central source for all JSON schemas used across the OpenCHAMI consortium. By maintaining a unified set of schemas, we ensure consistency and compatibility across all OpenCHAMI projects.

## Overview

This repository contains JSON schemas that define the structure and validation rules for data used across various OpenCHAMI projects. Each schema is generated from Go structs using reflection, ensuring that the schema remains consistent with the underlying data models.

## Directory Structure

The repository is organized as follows:

- **schemas/**: This directory contains all the Go structs that will become JSON schemas, each in its own file. Schemas are named according to their purpose or associated data structure.
- **examples/**: This directory contains example payloads that conform to the schemas. These examples serve as references for developers implementing or integrating with OpenCHAMI components.
- **docs/**: Documentation related to the schemas, including detailed descriptions and usage guidelines, is found here.

## How to Contribute

We welcome contributions to the schema repository! Here’s how you can get involved:

1. **Fork the Repository**: Start by forking this repository to your own GitHub account.
2. **Create a Branch**: Create a new branch for your changes.
3. **Add/Update Schemas**: Modify or add new Go structs in the appropriate files. Use reflection to generate the corresponding JSON schema.
4. **Test Your Changes**: Ensure that your changes are valid and conform to the repository’s guidelines. Include example payloads in the `examples/` directory.
5. **Submit a Pull Request**: Once your changes are ready, submit a pull request for review.

## Generating JSON Schemas

Schemas in this repository are generated from Go structs using reflection. Here’s an example of how to generate a JSON schema:

```go
package schemas

// Example struct definition
type Node struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    IP   string `json:"ip"`
}
```

## Schema Versioning

Each schema is versioned using an envelope/header format. This allows servers to verify the schema version before processing the contained data. Here’s an example:

```go
package schemas

// Envelope structure for schema versioning
type Envelope struct {
    SchemaID   string      `json:"schema_id"`
    Version    string      `json:"version"`
    Payload    interface{} `json:"payload"`
}
```

## Referencing Schemas

All schemas are published on a webpage for easy access and reference. Servers and clients can use these schemas to validate data and ensure compliance with OpenCHAMI standards.

## Resources

- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Kubernetes API Versioning](https://kubernetes.io/docs/concepts/overview/kubernetes-api/#api-versioning)
- [OpenCHAMI Node Orchestrator](https://github.com/OpenCHAMI/node-orchestrator)

## License

This repository is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

---

Thank you for contributing to the OpenCHAMI Schema Repository! Together, we can maintain a consistent and reliable set of data models for all OpenCHAMI projects.
