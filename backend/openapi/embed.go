// Package openapi holds the Senju OpenAPI 3 specification (embedded) and contract tests.
package openapi

import (
	_ "embed"
)

// SpecYAML is the canonical OpenAPI 3 document embedded at build time and served at GET /openapi.yaml.
//
//go:embed openapi.yaml
var SpecYAML []byte
