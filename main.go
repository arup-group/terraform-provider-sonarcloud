// Package main is the entry point for the terraform-provider-sonarcloud.
package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"terraform-provider-sonarcloud/sonarcloud"
)

// Format examples and generate documentation
//go:generate terraform fmt -recursive ./examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	err := providerserver.Serve(context.Background(), sonarcloud.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/arup-group/sonarcloud",
	})
	if err != nil {
		log.Fatal(err)
	}
}
