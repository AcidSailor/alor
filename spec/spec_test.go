package spec

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSpecDerefJSON guards the embedded MCP artifact's structural invariants:
// non-empty, valid JSON, carrying the operations and component schemas a
// downstream tool-schema builder reads, and — the point of the deref branch —
// no "$ref" for a resolver to chase. The analog of the runtime-free guard on
// models.gen.go: a cheap check that the artifact still holds the promised shape.
func TestSpecDerefJSON(t *testing.T) {
	require.NotEmpty(
		t,
		SpecDerefJSON,
		"SpecDerefJSON is empty; run `task deref`",
	)

	var doc struct {
		OpenAPI    string                     `json:"openapi"`
		Paths      map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	require.NoError(
		t,
		json.Unmarshal(SpecDerefJSON, &doc),
		"SpecDerefJSON is not valid JSON",
	)

	assert.NotEmpty(t, doc.OpenAPI, "missing top-level openapi version")
	assert.NotEmpty(
		t,
		doc.Paths,
		"paths is empty; the curated operations did not survive deref",
	)
	assert.NotEmpty(
		t,
		doc.Components.Schemas,
		"components.schemas is empty; name->schema lookup would fail downstream",
	)

	// Full dereferencing is the contract: a residual $ref means the bundle left
	// a reference uninlined, and the downstream resolver would dangle.
	assert.False(
		t,
		bytes.Contains(SpecDerefJSON, []byte(`"$ref"`)),
		`SpecDerefJSON contains "$ref"; expected a fully dereferenced document`,
	)
}
