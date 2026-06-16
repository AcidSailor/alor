// Package spec embeds the dereferenced OpenAPI document from the MCP branch of
// the generation pipeline so downstream modules (e.g. an Alor MCP server) can
// assemble per-tool JSON Schemas without re-running Docker.
//
// SpecDerefJSON is the curated spec/spec.yml (WebSocket/deprecated ops dropped,
// oneOf responses collapsed to Heavy) run through Redocly bundle --dereferenced:
// every "$ref" is inlined while components.schemas is retained, so name->schema
// lookup works and no reference is left for a resolver to chase. Stays OpenAPI
// 3.0.1.
//
// The ref-based spec.yml is the oapi-codegen input only, not embedded. See task
// deref in taskfile.yml.
package spec

import _ "embed"

// SpecDerefJSON is the dereferenced Alor OpenAPI document. See task deref.
//
//go:embed spec-deref.json
var SpecDerefJSON []byte
