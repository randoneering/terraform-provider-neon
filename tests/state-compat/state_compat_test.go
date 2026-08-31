// State-compatibility integration test for the SDK v2 -> Plugin Framework
// migration pilot.
//
// Captures a live API key with the SDK v2 binary, swaps the provider
// implementation to the framework binary via dev_overrides, and asserts
// `terraform plan` exits 0 (no diff). Skipped unless TF_ACC=1 and
// NEON_API_KEY are set.
//
// Run with:
//
//	TF_ACC=1 NEON_API_KEY=... go test ./tests/state-compat/...
package compat_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

const (
	providerAddr = "registry.terraform.io/neon/neon"
)

func TestStateCompat_SDKv2ToFramework(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}
	if os.Getenv("NEON_API_KEY") == "" {
		t.Skip("NEON_API_KEY must be set")
	}
	terraformBin := os.Getenv("TERRAFORM_BIN")
	if terraformBin == "" {
		terraformBin = "terraform"
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	workDir := t.TempDir()
	binDir := filepath.Join(workDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(binDir, "terraform-provider-neon")
	sdkv2Build := filepath.Join(workDir, "sdkv2-build", "terraform-provider-neon")
	fwBuild := filepath.Join(workDir, "fw-build", "terraform-provider-neon")
	if err := os.MkdirAll(filepath.Dir(sdkv2Build), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(fwBuild), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Log("building SDK v2 provider binary")
	runGo(t, repoRoot, "build", "-o", sdkv2Build, ".")

	t.Log("building framework provider binary")
	runGo(t, repoRoot, "build", "-o", fwBuild, "./cmd/neon-framework")

	applyDir := filepath.Join(workDir, "apply")
	if err := os.MkdirAll(applyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	keyName := "state-compat-" + uuid.NewString()
	config := fmt.Sprintf(`terraform {
  required_providers {
    neon = {
      source  = "%s"
      version = ">= 0.0.0"
    }
  }
}

resource "neon_api_key" "this" {
  name = %q
}
`, providerAddr, keyName)
	if err := os.WriteFile(filepath.Join(applyDir, "main.tf"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	// Single dev_overrides directory; we swap the binary inside it between
	// phases rather than re-initializing. Re-init triggers a registry
	// version query that fails because neon/neon has no published releases
	// yet (this is the pilot).
	rcPath := filepath.Join(applyDir, ".terraformrc")
	if err := os.WriteFile(rcPath, []byte(devOverridesRC(t, binDir)), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Logf("applying with SDK v2 binary (key=%s)", keyName)
	copyFile(t, sdkv2Build, binPath)
	runTerraform(t, terraformBin, applyDir, rcPath, "init")
	runTerraform(t, terraformBin, applyDir, rcPath, "apply", "-auto-approve")

	// Destroy via SDK v2 regardless of framework plan outcome so the live
	// API key is always revoked.
	t.Cleanup(func() {
		copyFile(t, sdkv2Build, binPath)
		runTerraform(t, terraformBin, applyDir, rcPath, "destroy", "-auto-approve")
	})

	t.Log("swapping dev_overrides to framework binary")
	copyFile(t, fwBuild, binPath)

	t.Log("planning with framework binary against SDK v2 state")
	out, exit := runTerraformCapture(t, terraformBin, applyDir, rcPath, "plan", "-detailed-exitcode")
	if exit != 0 {
		t.Fatalf("framework plan returned non-zero exit %d; expected 0 (no diff)\n%s", exit, out)
	}
	t.Logf("plan output:\n%s", out)
}

func devOverridesRC(t *testing.T, binDir string) string {
	t.Helper()
	return fmt.Sprintf(`provider_installation {
  dev_overrides {
    %q = %q
  }
}
`, providerAddr, binDir)
}

func runGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %v failed: %v\n%s", args, err, out)
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatal(err)
	}
}

func runTerraform(t *testing.T, bin, dir, rcPath string, args ...string) {
	t.Helper()
	out, exit := runTerraformCapture(t, bin, dir, rcPath, args...)
	if exit != 0 {
		t.Fatalf("terraform %v exited %d\n%s", args, exit, out)
	}
}

func runTerraformCapture(t *testing.T, bin, dir, rcPath string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TF_CLI_CONFIG_FILE="+rcPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exit = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("terraform %v: %v\nstdout: %s\nstderr: %s", args, err, stdout.String(), stderr.String())
	}
	return stdout.String() + stderr.String(), exit
}
