package pullzoneresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/customtype"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"strings"

	"testing"
)

var testOriginOriginType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"type":                  tftypes.String,
		"url":                   tftypes.String,
		"host_header":           tftypes.String,
		"storagezone":           tftypes.Number,
		"script":                tftypes.Number,
		"middleware_script":     tftypes.Number,
		"container_app_id":      tftypes.String,
		"container_endpoint_id": tftypes.String,
		"dns_port":              tftypes.Number,
		"dns_scheme":            tftypes.String,
	},
}

var testOriginRoutingFiltersElementType = tftypes.Set{
	ElementType: tftypes.String,
}

var testOriginRoutingType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"filters": testOriginRoutingFiltersElementType,
	},
}

type testOriginTestCase struct {
	ExpectedError bool
	PlanValues    map[string]tftypes.Value
}

func TestOrigin(t *testing.T) {
	testCases := []testOriginTestCase{
		// OriginUrl
		// ok
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com"}, nil),
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com", HostHeader: "backend.example.com"}, nil),
		// invalid args
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: ""}, nil),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "invalid-url"}, nil),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "ftp://example.com"}, nil),
		// unexpected args
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", DnsScheme: "https", DnsPort: 443}, nil),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", Storagezone: 12345}, nil),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", Script: 12345}, nil),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234-cdn-1"}, nil),
		// middleware
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com", Middleware: 12345}, "scripting"),
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com", Middleware: 12345}, "scripting,eu"),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", Middleware: 12345}, "all"),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com", Middleware: 12345}, nil),
		// routing filters
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com"}, "all"),
		testOriginMakeTestCase(false, "OriginUrl", originPayload{Url: "https://example.com"}, "eu"),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com"}, "scripting"),
		testOriginMakeTestCase(true, "OriginUrl", originPayload{Url: "https://example.com"}, "scripting,eu"),

		// DnsAccelerate
		// ok
		testOriginMakeTestCase(false, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443}, nil),
		testOriginMakeTestCase(false, "DnsAccelerate", originPayload{DnsScheme: "http", DnsPort: 443}, nil),
		// unexpected args
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, Url: "https://192.0.2.3:443/"}, nil),
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, Storagezone: 12345}, nil),
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, Script: 12345}, nil),
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1"}, nil),
		// middleware
		testOriginMakeTestCase(false, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, Middleware: 12345}, "scripting"),
		testOriginMakeTestCase(false, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, Middleware: 12345}, "scripting,eu"),
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, Middleware: 12345}, "all"),
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443, Middleware: 12345}, nil),
		// routing filters
		testOriginMakeTestCase(false, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443}, "all"),
		testOriginMakeTestCase(false, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443}, "eu"),
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443}, "scripting"),
		testOriginMakeTestCase(true, "DnsAccelerate", originPayload{DnsScheme: "https", DnsPort: 443}, "scripting,eu"),

		// StorageZone
		// ok
		testOriginMakeTestCase(false, "StorageZone", originPayload{Storagezone: 12345}, nil),
		// unexpected args
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, Url: "https://example.com"}, nil),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, DnsScheme: "https", DnsPort: 443}, nil),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, Script: 12345}, nil),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234-cdn-1"}, nil),
		// middleware
		testOriginMakeTestCase(false, "StorageZone", originPayload{Storagezone: 12345, Middleware: 12345}, "scripting"),
		testOriginMakeTestCase(false, "StorageZone", originPayload{Storagezone: 12345, Middleware: 12345}, "scripting,eu"),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, Middleware: 12345}, "all"),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345, Middleware: 12345}, nil),
		// routing filters
		testOriginMakeTestCase(false, "StorageZone", originPayload{Storagezone: 12345}, "all"),
		testOriginMakeTestCase(false, "StorageZone", originPayload{Storagezone: 12345}, "eu"),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345}, "scripting"),
		testOriginMakeTestCase(true, "StorageZone", originPayload{Storagezone: 12345}, "scripting,eu"),

		// ComputeScript
		// ok
		testOriginMakeTestCase(false, "ComputeScript", originPayload{Script: 12345}, "scripting"),
		// unexpected args
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, Url: "https://example.com"}, "scripting"),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, DnsScheme: "https", DnsPort: 443}, "scripting"),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, Storagezone: 12345}, "scripting"),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234-cdn-1"}, "scripting"),
		// middleware
		testOriginMakeTestCase(false, "ComputeScript", originPayload{Script: 12345, Middleware: 12345}, "scripting"),
		testOriginMakeTestCase(false, "ComputeScript", originPayload{Script: 12345, Middleware: 12345}, "scripting,eu"),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, Middleware: 12345}, "all"),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345, Middleware: 12345}, nil),
		// routing filters
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345}, nil),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345}, "all"),
		testOriginMakeTestCase(true, "ComputeScript", originPayload{Script: 12345}, "eu"),
		testOriginMakeTestCase(false, "ComputeScript", originPayload{Script: 12345}, "scripting,eu"),

		// ComputeContainer
		// ok
		testOriginMakeTestCase(false, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1"}, nil),
		// unexpected args
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Url: "https://192.0.2.3:443/"}, nil),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", DnsScheme: "https", DnsPort: 443}, nil),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Storagezone: 12345}, nil),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Script: 12345}, nil),
		// middleware
		testOriginMakeTestCase(false, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Middleware: 12345}, "scripting"),
		testOriginMakeTestCase(false, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Middleware: 12345}, "scripting,eu"),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Middleware: 12345}, "all"),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1", Middleware: 12345}, nil),
		// routing filters
		testOriginMakeTestCase(false, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1"}, "all"),
		testOriginMakeTestCase(false, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1"}, "eu"),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1"}, "scripting"),
		testOriginMakeTestCase(true, "ComputeContainer", originPayload{ContainerAppId: "abc1234d", ContainerEndpointId: "abc1234d-cdn-1"}, "scripting,eu"),
	}

	configSchema := schema.Schema{
		Blocks: map[string]schema.Block{
			"origin": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{},
					"url": schema.StringAttribute{
						CustomType: customtype.PullzoneOriginUrlType{},
					},
					"host_header":           schema.StringAttribute{},
					"storagezone":           schema.Int64Attribute{},
					"script":                schema.Int64Attribute{},
					"middleware_script":     schema.Int64Attribute{},
					"container_app_id":      schema.StringAttribute{},
					"container_endpoint_id": schema.StringAttribute{},
					"dns_port":              schema.Int64Attribute{},
					"dns_scheme":            schema.StringAttribute{},
				},
			},
			"routing": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"filters": schema.SetAttribute{
						ElementType: types.StringType,
					},
				},
			},
		},
	}

	configTypes := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"origin":  testOriginOriginType,
			"routing": testOriginRoutingType,
		},
	}

	for _, testCase := range testCases {
		request := resource.ValidateConfigRequest{
			Config: tfsdk.Config{
				Schema: configSchema,
				Raw:    tftypes.NewValue(configTypes, testCase.PlanValues),
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
	}
}

func testOriginMakeTestCase(err bool, ot string, op originPayload, filters any) testOriginTestCase {
	var ft []tftypes.Value = nil

	if f, ok := filters.(string); ok {
		elements := strings.Split(f, ",")
		if len(elements) > 0 {
			ft = make([]tftypes.Value, 0, len(elements))

			for _, v := range elements {
				ft = append(ft, tftypes.NewValue(tftypes.String, v))
			}
		}
	}

	return testOriginTestCase{
		ExpectedError: err,
		PlanValues: map[string]tftypes.Value{
			"origin": testOriginMakeOrigin(ot, op),
			"routing": tftypes.NewValue(testOriginRoutingType, map[string]tftypes.Value{
				"filters": tftypes.NewValue(testOriginRoutingFiltersElementType, ft),
			}),
		},
	}
}

type originPayload struct {
	Url                 string
	HostHeader          string
	Storagezone         int64
	Script              int64
	Middleware          int64
	ContainerAppId      string
	ContainerEndpointId string
	DnsPort             int64
	DnsScheme           string
}

func testOriginMakeOrigin(ot string, payload originPayload) tftypes.Value {
	var url any = nil
	var hostHeader any = nil
	var storagezone any = nil
	var script any = nil
	var middleware any = nil
	var containerAppId any = nil
	var containerEndpointId any = nil
	var dnsPort any = nil
	var dnsScheme any = nil

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

	if payload.Middleware != 0 {
		middleware = payload.Middleware
	}

	if payload.ContainerAppId != "" {
		containerAppId = payload.ContainerAppId
	}

	if payload.ContainerEndpointId != "" {
		containerEndpointId = payload.ContainerEndpointId
	}

	if payload.DnsPort != 0 {
		dnsPort = payload.DnsPort
	}

	if payload.DnsScheme != "" {
		dnsScheme = payload.DnsScheme
	}

	return tftypes.NewValue(testOriginOriginType, map[string]tftypes.Value{
		"type":                  tftypes.NewValue(tftypes.String, ot),
		"url":                   tftypes.NewValue(tftypes.String, url),
		"host_header":           tftypes.NewValue(tftypes.String, hostHeader),
		"storagezone":           tftypes.NewValue(tftypes.Number, storagezone),
		"script":                tftypes.NewValue(tftypes.Number, script),
		"middleware_script":     tftypes.NewValue(tftypes.Number, middleware),
		"container_app_id":      tftypes.NewValue(tftypes.String, containerAppId),
		"container_endpoint_id": tftypes.NewValue(tftypes.String, containerEndpointId),
		"dns_port":              tftypes.NewValue(tftypes.Number, dnsPort),
		"dns_scheme":            tftypes.NewValue(tftypes.String, dnsScheme),
	})
}
