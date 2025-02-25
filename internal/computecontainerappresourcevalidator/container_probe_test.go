package computecontainerappresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"testing"
)

func TestContainerProbeObject(t *testing.T) {
	computeContainerAppContainerProbeHttpType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"path":            types.StringType,
			"expected_status": types.Int64Type,
		},
	}

	computeContainerAppContainerProbeGrpcType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"service": types.StringType,
		},
	}

	type testCase struct {
		ExpectedError string
		Values        map[string]attr.Value
	}

	attrTypes := map[string]attr.Type{
		"type": types.StringType,
		"port": types.Int64Type,
		"http": types.SetType{ElemType: computeContainerAppContainerProbeHttpType},
		"grpc": types.SetType{ElemType: computeContainerAppContainerProbeGrpcType},
	}

	httpSetNull := types.SetNull(computeContainerAppContainerProbeHttpType)
	gprcSetNull := types.SetNull(computeContainerAppContainerProbeGrpcType)

	httpSetEmpty := types.SetValueMust(computeContainerAppContainerProbeHttpType, []attr.Value{})
	gprcSetEmpty := types.SetValueMust(computeContainerAppContainerProbeGrpcType, []attr.Value{})

	httpSetValid := types.SetValueMust(computeContainerAppContainerProbeHttpType, []attr.Value{
		types.ObjectValueMust(computeContainerAppContainerProbeHttpType.AttrTypes, map[string]attr.Value{
			"path":            types.StringValue("/ping"),
			"expected_status": types.Int64Null(),
		}),
	})

	grpcSetValid := types.SetValueMust(computeContainerAppContainerProbeGrpcType, []attr.Value{
		types.ObjectValueMust(computeContainerAppContainerProbeGrpcType.AttrTypes, map[string]attr.Value{
			"service": types.StringValue("ping"),
		}),
	})

	httpSetMany := types.SetValueMust(computeContainerAppContainerProbeHttpType, []attr.Value{
		types.ObjectValueMust(computeContainerAppContainerProbeHttpType.AttrTypes, map[string]attr.Value{
			"path":            types.StringValue("/ping"),
			"expected_status": types.Int64Null(),
		}),
		types.ObjectValueMust(computeContainerAppContainerProbeHttpType.AttrTypes, map[string]attr.Value{
			"path":            types.StringValue("/health"),
			"expected_status": types.Int64Null(),
		}),
	})

	grpcSetMany := types.SetValueMust(computeContainerAppContainerProbeGrpcType, []attr.Value{
		types.ObjectValueMust(computeContainerAppContainerProbeGrpcType.AttrTypes, map[string]attr.Value{
			"service": types.StringValue("ping"),
		}),
		types.ObjectValueMust(computeContainerAppContainerProbeGrpcType.AttrTypes, map[string]attr.Value{
			"service": types.StringValue("healthcheck"),
		}),
	})

	testCases := []testCase{
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(8080),
				"http": httpSetValid,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"type": types.StringValue("tcp"),
				"port": types.Int64Value(8080),
				"http": httpSetNull,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpSetNull,
				"grpc": grpcSetValid,
			},
		},
		{
			ExpectedError: errProbeTypeNull,
			Values: map[string]attr.Value{
				"type": types.StringNull(),
				"port": types.Int64Null(),
				"http": httpSetNull,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: errProbeType,
			Values: map[string]attr.Value{
				"type": types.StringValue("test"),
				"port": types.Int64Null(),
				"http": httpSetNull,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: errProbePort,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Null(),
				"http": httpSetValid,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: errProbePort,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(-1),
				"http": httpSetValid,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: errProbePort,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(123456),
				"http": httpSetValid,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: errProbeHttpBlockMissing,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(8080),
				"http": httpSetEmpty,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: errProbeGrpcBlockMissing,
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpSetNull,
				"grpc": gprcSetEmpty,
			},
		},
		{
			ExpectedError: errProbeHttpBlockTooMany,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(8080),
				"http": httpSetMany,
				"grpc": gprcSetNull,
			},
		},
		{
			ExpectedError: errProbeGrpcBlockTooMany,
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpSetNull,
				"grpc": grpcSetMany,
			},
		},
		{
			ExpectedError: errProbeGrpcBlockExtra,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(8080),
				"http": httpSetValid,
				"grpc": grpcSetValid,
			},
		},
		{
			ExpectedError: errProbeHttpBlockExtra,
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpSetValid,
				"grpc": grpcSetValid,
			},
		},
	}

	for _, testCase := range testCases {
		value, diags := types.ObjectValue(attrTypes, testCase.Values)
		if diags.HasError() {
			t.Error(diags)
			continue
		}

		request := validator.ObjectRequest{
			Path:        path.Root("test_probe"),
			ConfigValue: value,
		}

		response := validator.ObjectResponse{}
		containerProbe{}.ValidateObject(context.Background(), request, &response)

		if testCase.ExpectedError != "" && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if testCase.ExpectedError == "" && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}

		expectedDiag := diag.NewErrorDiagnostic("Invalid probe configuration", testCase.ExpectedError)
		if testCase.ExpectedError != "" && !response.Diagnostics.Contains(expectedDiag) {
			t.Errorf("expected %s, got %s", expectedDiag, response.Diagnostics.Errors())
		}
	}
}
