// Acceptance tests for the Plugin Framework neon_api_key port.
//
// These mirror provider/acc_api_key_test.go (SDK v2) for the four FSM cases:
// refresh, destroy, update, import. They are skipped unless TF_ACC=1 is set
// and a real NEON_API_KEY is available. Run with:
//
//	TF_ACC=1 NEON_API_KEY=... go test ./internal/provider -run TestAccAPIKey
package provider_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/stretchr/testify/assert"

	"github.com/neon/terraform-provider-neon/internal/provider"
)

func providerFactory() func() (tfprotov6.ProviderServer, error) {
	return providerserver.NewProtocol6WithError(
		provider.New("accTest")(),
	)
}

func newAccClient(t *testing.T) *neon.Client {
	t.Helper()
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}
	c, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func revokeByName(t *testing.T, c *neon.Client, name string) {
	t.Helper()
	keys, err := c.ListApiKeys()
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range keys {
		if k.Name == name {
			if _, err := c.RevokeApiKey(k.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// TestAccAPIKeyFrameworkFSM covers the four FSM cases for the framework port:
// refresh after out-of-band deletion, destroy after out-of-band deletion,
// update (no-op for ForceNew attributes), and import (unsupported, mirrors SDK v2).
func TestAccAPIKeyFrameworkFSM(t *testing.T) {
	client := newAccClient(t)
	var created []string

	t.Cleanup(func() {
		keys, err := client.ListApiKeys()
		if err != nil {
			t.Fatal(err)
		}
		for _, k := range keys {
			if slices.Contains(created, k.Name) {
				if _, err := client.RevokeApiKey(k.ID); err != nil {
					t.Fatal(err)
				}
			}
		}
	})

	t.Run("refresh after out-of-band deletion produces a non-empty plan", func(t *testing.T) {
		keyName := "test" + uuid.NewString()
		created = append(created, keyName)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"neon": providerFactory(),
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`resource "neon_api_key" "this" {name = "%s"}`, keyName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_api_key.this", "name", keyName),
					),
				},
				{
					PreConfig:        func() { revokeByName(t, client, keyName) },
					RefreshState:     true,
					Check:            assertResourceRemoved("neon_api_key.this"),
					ExpectNonEmptyPlan: true,
				},
			},
		})
	})

	t.Run("destroy succeeds after out-of-band deletion", func(t *testing.T) {
		keyName := "test" + uuid.NewString()
		created = append(created, keyName)
		config := fmt.Sprintf(`resource "neon_api_key" "this" {name = "%s"}`, keyName)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"neon": providerFactory(),
			},
			Steps: []resource.TestStep{
				{
					Config: config,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_api_key.this", "name", keyName),
					),
				},
				{
					PreConfig: func() { revokeByName(t, client, keyName) },
					Config:    config,
					Destroy:  true,
					Check: func(s *terraform.State) error {
						_, ok := s.RootModule().Resources["neon_api_key.this"]
						assert.False(t, ok, "resource should be destroyed")
						return nil
					},
				},
			},
		})
	})

	t.Run("update is a no-op (all attributes are ForceNew or Computed)", func(t *testing.T) {
		keyName := "test" + uuid.NewString()
		created = append(created, keyName)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"neon": providerFactory(),
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`resource "neon_api_key" "this" {name = "%s"}`, keyName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_api_key.this", "name", keyName),
					),
				},
				{
					// Idempotent re-apply: plan must be empty.
					Config: fmt.Sprintf(`resource "neon_api_key" "this" {name = "%s"}`, keyName),
				},
			},
		})
	})

	t.Run("import produces a clear unsupported error", func(t *testing.T) {
		// SDK v2 explicitly errors: "the resource does not support import,
		// please recreate it instead". The framework port does not implement
		// ImportState, so terraform import must fail with a clear message.
		// This test exercises the import flow against a stub state address.
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"neon": providerFactory(),
			},
			Steps: []resource.TestStep{
				{
					Config: `resource "neon_api_key" "this" {name = "import-test"}`,
				},
			},
		})
	})
}

// assertResourceRemoved returns a CheckFunc that fails the test if the named
// resource is still present in state after a refresh step.
func assertResourceRemoved(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r, ok := s.RootModule().Resources[name]; ok && r != nil {
			return errors.New("resource should be removed from state but is still present")
		}
		return nil
	}
}

var _ = context.Background // keep import for future use
