// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package loadbalancerresourcevalidator

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/customtype"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"testing"
)

var testOriginOriginType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"type":                  tftypes.String,
		"url":                   tftypes.String,
		"host_header":           tftypes.String,
		"storagezone":           tftypes.Number,
		"script":                tftypes.Number,
		"container_app_id":      tftypes.String,
		"container_endpoint_id": tftypes.String,
	},
}

type testOriginTestCase struct {
	ExpectedError bool
	PlanValues    map[string]tftypes.Value
}

type originPayload struct {
	Url                 string
	HostHeader          string
	Storagezone         int64
	Script              int64
	ContainerAppId      string
	ContainerEndpointId string
}

func TestOrigin(t *testing.T) {
	testCases := []testOriginTestCase{
		// OriginUrl
		// ok
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com"}),
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com", HostHeader: "backend.example.com"}),
		// invalid args
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: ""}),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "invalid-url"}),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "ftp://example.com"}),
		// unexpected args
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", Storagezone: 12345}),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", Script: 12345}),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234-cdn-1"}),

		// StorageZone
		// ok
		testOriginMakeTestCase(false, "StorageZone", originPayload{Storagezone: 12345}),
		// unexpected args
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, Url: "https://example.com"}),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, Script: 12345}),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234-cdn-1"}),

		// ComputeScript
		// ok
		testOriginMakeTestCase(false, "ComputeScript", originPayload{Script: 12345}),
		// unexpected args
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, Url: "https://example.com"}),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, Storagezone: 12345}),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234-cdn-1"}),

		// ComputeContainer
		// ok
		testOriginMakeTestCase(false, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1"}),
		// unexpected args
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Url: "https://192.0.2.3:443/"}),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Storagezone: 12345}),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Script: 12345}),
	}

	configSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{},
			"url": schema.StringAttribute{
				CustomType: customtype.PullzoneOriginUrlType{},
			},
			"host_header":           schema.StringAttribute{},
			"storagezone":           schema.Int64Attribute{},
			"script":                schema.Int64Attribute{},
			"container_app_id":      schema.StringAttribute{},
			"container_endpoint_id": schema.StringAttribute{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			request := resource.ValidateConfigRequest{
				Config: tfsdk.Config{
					Schema: configSchema,
					Raw:    tftypes.NewValue(testOriginOriginType, testCase.PlanValues),
				},
			}

			response := resource.ValidateConfigResponse{}
			originValidator{}.ValidateResource(context.Background(), request, &response)

			if testCase.ExpectedError && !response.Diagnostics.HasError() {
				t.Error("expected error, got none")
			}

			if !testCase.ExpectedError && response.Diagnostics.HasError() {
				t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
			}
		})
	}
}

func testOriginMakeTestCase(err bool, ot string, payload originPayload) testOriginTestCase {
	var url any = nil
	var hostHeader any = nil
	var storagezone any = nil
	var script any = nil
	var containerAppId any = nil
	var containerEndpointId any = nil

	if payload.Url != "" {
		url = payload.Url
	}

	if payload.HostHeader != "" {
		hostHeader = payload.HostHeader
	}

	if payload.Storagezone != 0 {
		storagezone = payload.Storagezone
	}

	if payload.Script != 0 {
		script = payload.Script
	}

	if payload.ContainerAppId != "" {
		containerAppId = payload.ContainerAppId
	}

	if payload.ContainerEndpointId != "" {
		containerEndpointId = payload.ContainerEndpointId
	}

	return testOriginTestCase{
		ExpectedError: err,
		PlanValues: map[string]tftypes.Value{
			"type":                  tftypes.NewValue(tftypes.String, ot),
			"url":                   tftypes.NewValue(tftypes.String, url),
			"host_header":           tftypes.NewValue(tftypes.String, hostHeader),
			"storagezone":           tftypes.NewValue(tftypes.Number, storagezone),
			"script":                tftypes.NewValue(tftypes.Number, script),
			"container_app_id":      tftypes.NewValue(tftypes.String, containerAppId),
			"container_endpoint_id": tftypes.NewValue(tftypes.String, containerEndpointId),
		},
	}
}
