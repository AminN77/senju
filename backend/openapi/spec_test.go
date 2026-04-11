package openapi

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestSpecLoadsAndValidates(t *testing.T) {
	t.Parallel()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(SpecYAML)
	if err != nil {
		t.Fatalf("load openapi: %v", err)
	}
	if err := doc.Validate(loader.Context); err != nil {
		t.Fatalf("validate openapi: %v", err)
	}
	paths := []string{
		"/",
		"/health/live",
		"/health/ready",
		"/version",
		"/metrics",
		"/v1/jobs/fastq-upload/metadata",
		"/v1/objects/multipart",
		"/v1/objects/multipart/{upload_id}/parts",
		"/v1/objects/multipart/{upload_id}/complete",
	}
	for _, p := range paths {
		if doc.Paths == nil || doc.Paths.Find(p) == nil {
			t.Errorf("missing path %q in OpenAPI spec", p)
		}
	}
}
