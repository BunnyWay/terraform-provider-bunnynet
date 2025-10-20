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
		"http": types.ListType{ElemType: computeContainerAppContainerProbeHttpType},
		"grpc": types.ListType{ElemType: computeContainerAppContainerProbeGrpcType},
	}

	httpListNull := types.ListNull(computeContainerAppContainerProbeHttpType)
	grpcListNull := types.ListNull(computeContainerAppContainerProbeGrpcType)

	httpListEmpty := types.ListValueMust(computeContainerAppContainerProbeHttpType, []attr.Value{})
	gprcListEmpty := types.ListValueMust(computeContainerAppContainerProbeGrpcType, []attr.Value{})

	httpListValid := types.ListValueMust(computeContainerAppContainerProbeHttpType, []attr.Value{
		types.ObjectValueMust(computeContainerAppContainerProbeHttpType.AttrTypes, map[string]attr.Value{
			"path":            types.StringValue("/ping"),
			"expected_status": types.Int64Null(),
		}),
	})

	grpcListValid := types.ListValueMust(computeContainerAppContainerProbeGrpcType, []attr.Value{
		types.ObjectValueMust(computeContainerAppContainerProbeGrpcType.AttrTypes, map[string]attr.Value{
			"service": types.StringValue("ping"),
		}),
	})

	httpListMany := types.ListValueMust(computeContainerAppContainerProbeHttpType, []attr.Value{
		types.ObjectValueMust(computeContainerAppContainerProbeHttpType.AttrTypes, map[string]attr.Value{
			"path":            types.StringValue("/ping"),
			"expected_status": types.Int64Null(),
		}),
		types.ObjectValueMust(computeContainerAppContainerProbeHttpType.AttrTypes, map[string]attr.Value{
			"path":            types.StringValue("/health"),
			"expected_status": types.Int64Null(),
		}),
	})

	grpcListMany := types.ListValueMust(computeContainerAppContainerProbeGrpcType, []attr.Value{
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
				"http": httpListValid,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"type": types.StringValue("tcp"),
				"port": types.Int64Value(8080),
				"http": httpListNull,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpListNull,
				"grpc": grpcListValid,
			},
		},
		{
			ExpectedError: errProbeTypeNull,
			Values: map[string]attr.Value{
				"type": types.StringNull(),
				"port": types.Int64Null(),
				"http": httpListNull,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: errProbeType,
			Values: map[string]attr.Value{
				"type": types.StringValue("test"),
				"port": types.Int64Null(),
				"http": httpListNull,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: errProbePort,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Null(),
				"http": httpListValid,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: errProbePort,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(-1),
				"http": httpListValid,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: errProbePort,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(123456),
				"http": httpListValid,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: errProbeHttpBlockMissing,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(8080),
				"http": httpListEmpty,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: errProbeGrpcBlockMissing,
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpListNull,
				"grpc": gprcListEmpty,
			},
		},
		{
			ExpectedError: errProbeHttpBlockTooMany,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(8080),
				"http": httpListMany,
				"grpc": grpcListNull,
			},
		},
		{
			ExpectedError: errProbeGrpcBlockTooMany,
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpListNull,
				"grpc": grpcListMany,
			},
		},
		{
			ExpectedError: errProbeGrpcBlockExtra,
			Values: map[string]attr.Value{
				"type": types.StringValue("http"),
				"port": types.Int64Value(8080),
				"http": httpListValid,
				"grpc": grpcListValid,
			},
		},
		{
			ExpectedError: errProbeHttpBlockExtra,
			Values: map[string]attr.Value{
				"type": types.StringValue("grpc"),
				"port": types.Int64Value(8080),
				"http": httpListValid,
				"grpc": grpcListValid,
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
