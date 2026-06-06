package apispec

import _ "embed"

// Spec holds the raw OpenAPI 3.0 YAML, embedded at compile time.
//
//go:embed openapi.yaml
var Spec []byte
