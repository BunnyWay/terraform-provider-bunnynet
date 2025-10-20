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

func TestContainerEndpointObject(t *testing.T) {
	type testCase struct {
		ExpectedError string
		Values        map[string]attr.Value
	}

	var cdnType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"origin_ssl":              types.BoolType,
			"sticky_sessions":         types.BoolType,
			"sticky_sessions_headers": types.SetType{ElemType: types.StringType},
		},
	}

	var portType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"container": types.Int64Type,
			"exposed":   types.Int64Type,
			"protocols": types.SetType{ElemType: types.StringType},
		},
	}

	attrTypes := map[string]attr.Type{
		"name": types.StringType,
		"type": types.StringType,
		"port": types.ListType{ElemType: portType},
		"cdn":  types.ListType{ElemType: cdnType},
	}

	testCases := []testCase{
		// cdn
		{
			ExpectedError: errEndpointPortMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListNull(portType),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},
		{
			ExpectedError: errEndpointCdnMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointCdnTooMany,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(true),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},
		{
			ExpectedError: errEndpointCdnSslMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolNull(),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},

		{
			ExpectedError: errEndpointPortExposedNotNeeded,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Value(8080),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},
		{
			ExpectedError: errEndpointPortTooMany,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8081),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},
		{
			ExpectedError: errEndpointPortContainerMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Null(),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},
		{
			ExpectedError: errEndpointPortContainerMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Null(),
						"exposed":   types.Int64Value(1234),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},
		{
			ExpectedError: errEndpointPortProtocolNotNeeded,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("CDN"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
						}),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},

		// anycast
		{
			ExpectedError: errEndpointPortMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListNull(portType),
				"cdn":  types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Value(8080),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortExposedMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Value(8080),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8081),
						"exposed":   types.Int64Value(8081),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortContainerMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Null(),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortContainerMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Null(),
						"exposed":   types.Int64Value(1234),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortProtocolRequired,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Value(8080),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortProtocolRequired,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Value(8080),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointCdnNotNeeded,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("Anycast"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Value(8080),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
			},
		},

		// internalIp
		{
			ExpectedError: errEndpointPortMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListNull(portType),
				"cdn":  types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: "",
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortExposedNotNeeded,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Value(8080),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortTooMany,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8081),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortContainerMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Null(),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortContainerMissing,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Null(),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortProtocolRequired,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetNull(types.StringType),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointPortProtocolRequired,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{}),
					}),
				}),
				"cdn": types.ListNull(cdnType),
			},
		},
		{
			ExpectedError: errEndpointCdnNotNeeded,
			Values: map[string]attr.Value{
				"name": types.StringValue("test"),
				"type": types.StringValue("InternalIP"),
				"port": types.ListValueMust(portType, []attr.Value{
					types.ObjectValueMust(portType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(8080),
						"exposed":   types.Int64Null(),
						"protocols": types.SetValueMust(types.StringType, []attr.Value{
							types.StringValue("TCP"),
							types.StringValue("UDP"),
						}),
					}),
				}),
				"cdn": types.ListValueMust(cdnType, []attr.Value{
					types.ObjectValueMust(cdnType.AttrTypes, map[string]attr.Value{
						"origin_ssl":              types.BoolValue(false),
						"sticky_sessions":         types.BoolValue(true),
						"sticky_sessions_headers": types.SetValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
					}),
				}),
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
			Path:        path.Root("endpoint"),
			ConfigValue: value,
		}

		response := validator.ObjectResponse{}
		containerEndpoint{}.ValidateObject(context.Background(), request, &response)

		if testCase.ExpectedError != "" && !response.Diagnostics.HasError() {
			t.Error("expected error, got none")
		}

		if testCase.ExpectedError == "" && response.Diagnostics.HasError() {
			t.Errorf("expected no errors, got %s", response.Diagnostics.Errors())
		}

		expectedDiag := diag.NewErrorDiagnostic("Invalid endpoint.test configuration", testCase.ExpectedError)
		if testCase.ExpectedError != "" && !response.Diagnostics.Contains(expectedDiag) {
			t.Errorf("expected %s, got %s", expectedDiag, response.Diagnostics.Errors())
		}
	}
}
