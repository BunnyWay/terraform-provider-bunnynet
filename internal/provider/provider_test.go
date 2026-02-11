// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bunnynet": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
}

func newApiClient() *api.Client {
	apiKey := os.Getenv("BUNNYNET_API_KEY")
	apiUrl := "https://api.bunny.net"
	streamApiUrl := "https://video.bunnycdn.com"

	envApiUrl := os.Getenv("BUNNYNET_API_URL")
	if envApiUrl != "" {
		apiUrl = envApiUrl
	}

	envStreamApiUrl := os.Getenv("BUNNYNET_STREAM_API_URL")
	if envStreamApiUrl != "" {
		streamApiUrl = envStreamApiUrl
	}

	return api.NewClient(
		apiKey,
		apiUrl,
		streamApiUrl,
		fmt.Sprintf("Terraform/%s BunnynetProvider/%s", "test", "test"),
	)
}
