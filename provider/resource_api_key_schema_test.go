// Schema-compatibility test for the SDK v2 -> Plugin Framework pilot.
//
// Asserts that the framework port of neon_api_key exposes the same logical
// schema as the SDK v2 implementation: identical attribute names, types, and
// Required/Optional/Computed/Sensitive flags. State files written by either
// implementation must apply cleanly against the other; this test catches
// drift before the change ships.
//
// ForceNew (SDK v2) and RequiresReplace plan modifiers (framework) are
// exercised end-to-end by the acceptance tests under
// internal/provider/resource_api_key_acc_test.go, so they are intentionally
// not asserted here — the framework planmodifier package hides the
// RequiresReplace type behind an unexported struct, and reaching for it via
// string-matching the modifier description would couple the test to internal
// wording.

package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"

	fwprovider "github.com/neon/terraform-provider-neon/internal/provider"
)

func TestResourceAPIKey_SchemaCompatibility(t *testing.T) {
	sdkSchema := resourceAPIKey().Schema
	fwSchema := frameworkAPISchema(t)

	assert.Equal(t, len(sdkSchema), len(fwSchema.Attributes),
		"attribute count mismatch between SDK v2 and framework schemas")

	for name, sdkAttr := range sdkSchema {
		t.Run(name, func(t *testing.T) {
			fwAttr, ok := fwSchema.Attributes[name]
			if !ok {
				t.Fatalf("attribute %q present in SDK v2 schema but missing from framework schema", name)
			}

			assert.Equal(t, normalizeType(sdkAttr.Type.String()), normalizeType(fwAttr.GetType().String()),
				"type mismatch for %q", name)
			assert.Equal(t, sdkAttr.Required, fwAttr.IsRequired(),
				"Required mismatch for %q", name)
			assert.Equal(t, sdkAttr.Optional, fwAttr.IsOptional(),
				"Optional mismatch for %q", name)
			assert.Equal(t, sdkAttr.Computed, fwAttr.IsComputed(),
				"Computed mismatch for %q", name)
			assert.Equal(t, sdkAttr.Sensitive, fwAttr.IsSensitive(),
				"Sensitive mismatch for %q", name)
		})
	}

	for name := range fwSchema.Attributes {
		if _, ok := sdkSchema[name]; !ok {
			t.Errorf("attribute %q present in framework schema but missing from SDK v2 schema", name)
		}
	}
}

func frameworkAPISchema(t *testing.T) fwschema.Schema {
	t.Helper()
	var resp resource.SchemaResponse
	fwprovider.NewNeonAPIKeyResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("framework schema returned diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

// normalizeType reduces both SDK v2 and framework type names to the bare
// primitive (e.g. "String", "Bool", "Int64") so they can be compared.
// SDK v2 uses "TypeString", the framework uses "basetypes.StringType".
func normalizeType(s string) string {
	s = strings.TrimPrefix(s, "basetypes.")
	s = strings.TrimPrefix(s, "Type")
	s = strings.TrimSuffix(s, "Type")
	return s
}
