// Binary neon-framework serves the Plugin Framework provider for Neon.
//
// This is the pilot binary for the SDK v2 -> Plugin Framework migration.
// It serves only neon_api_key under the registry.terraform.io/neon/neon
// address while the rest of the resources stay on SDK v2 under
// terraform-provider-neon (the binary built from the package root main.go).
//
// Build:
//
//	go build -o terraform-provider-neon_v0.16.0-pre.1 ./cmd/neon-framework
//
// Users opt in by configuring their Terraform provider source:
//
//	terraform {
//	  required_providers {
//	    neon = {
//	      source  = "neon/neon"
//	      version = ">= 0.16.0-pre.1"
//	    }
//	  }
//	}
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/neon/terraform-provider-neon/internal/provider"
)

// ldflags-set at release time (see .goreleaser.yml).
var version string = "0.0.0-alpha.0"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/neon/neon",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
