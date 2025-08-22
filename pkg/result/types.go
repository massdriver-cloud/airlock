package result

import "github.com/massdriver-cloud/airlock/pkg/schema"

type SchemaResult struct {
	Schema *schema.Schema
	Diags  []Diagnostic
}

type CodeResult struct {
	Code  []byte
	Diags []Diagnostic
}

type Severity string

const (
	Warning Severity = "warning"
	Error   Severity = "error"
)

type Diagnostic struct {
	Path    string
	Code    string
	Message string
	Level   Severity
}
