// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package computecontainerimageresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestRegistry(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		Value         string
	}

	testCases := []testCase{
		{false, "DockerHub"},
		{false, "GitHub"},
		{false, "quay.io"},
		{false, "server01.example.com"},
		{false, "server01.example.com:5000"},
		{true, "server01.example.com:5000/subpath"},
		{true, "https://myserver.example.com"},
		{true, "localhost"},
		{true, "something."},
		{true, "-"},
		{true, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.Value, func(t *testing.T) {
			response := validator.StringResponse{}
			request := validator.StringRequest{
				Path:        path.Root("registry"),
				ConfigValue: types.StringValue(tc.Value),
			}

			registryValidator{}.ValidateString(context.Background(), request, &response)

			if tc.ExpectedError && !response.Diagnostics.HasError() {
				t.Error("expected error, got none")
			}

			if !tc.ExpectedError && response.Diagnostics.HasError() {
				t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
			}
		})
	}
}
