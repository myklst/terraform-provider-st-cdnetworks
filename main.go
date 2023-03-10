package main

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/myklst/terraform-provider-st-cdnetworks/cdnetworks"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name st-cdnetworks

func main() {
	providerAddress := os.Getenv("PROVIDER_LOCAL_PATH")
	if providerAddress == "" {
		providerAddress = "registry.terraform.io/myklst/st-cdnetworks"
	}
	providerserver.Serve(context.Background(), cdnetworks.New, providerserver.ServeOpts{
		Address: providerAddress,
	})
}
