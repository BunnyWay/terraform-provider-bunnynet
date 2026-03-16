package pullzoneresourcevalidator

import (
	"context"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/customtype"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"testing"
)

func TestOrigin(t *testing.T) {
	type testCase struct {
		ExpectedError bool
		PlanValues    map[string]tftypes.Value
	}

	originType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"type":              tftypes.String,
			"url":               tftypes.String,
			"storagezone":       tftypes.Number,
			"script":            tftypes.Number,
			"middleware_script": tftypes.Number,
		},
	}

	routingFiltersElementType := tftypes.Set{
		ElementType: tftypes.String,
	}

	routingType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"filters": routingFiltersElementType,
		},
	}

	testCases := []testCase{
		// OriginUrl
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, 12345),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, nil),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, 12345),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, 12345),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "all"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "OriginUrl"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, 12345),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, nil),
				}),
			},
		},

		// ComputeScript
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":               tftypes.NewValue(tftypes.String, "https://bunnycdn.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, 12345),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":               tftypes.NewValue(tftypes.String, "https://bunnycdn.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, 12345),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "all"),
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":               tftypes.NewValue(tftypes.String, nil),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, 12345),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, nil),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "ComputeScript"),
					"url":               tftypes.NewValue(tftypes.String, "https://bunnycdn.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, 12345),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},

		// StorageZone
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "StorageZone"),
					"url":               tftypes.NewValue(tftypes.String, nil),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "StorageZone"),
					"url":               tftypes.NewValue(tftypes.String, "https://example.com"),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, nil),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "StorageZone"),
					"url":               tftypes.NewValue(tftypes.String, nil),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, 12345),
					"middleware_script": tftypes.NewValue(tftypes.Number, nil),
				}),
				"routing": tftypes.NewValue(routingType, nil),
			},
		},
		{
			ExpectedError: false,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "StorageZone"),
					"url":               tftypes.NewValue(tftypes.String, nil),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, 12345),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "scripting"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "StorageZone"),
					"url":               tftypes.NewValue(tftypes.String, nil),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, 12345),
				}),
				"routing": tftypes.NewValue(routingType, map[string]tftypes.Value{
					"filters": tftypes.NewValue(routingFiltersElementType, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "all"),
					}),
				}),
			},
		},
		{
			ExpectedError: true,
			PlanValues: map[string]tftypes.Value{
				"origin": tftypes.NewValue(originType, map[string]tftypes.Value{
					"type":              tftypes.NewValue(tftypes.String, "StorageZone"),
					"url":               tftypes.NewValue(tftypes.String, nil),
					"storagezone":       tftypes.NewValue(tftypes.Number, 12345),
					"script":            tftypes.NewValue(tftypes.Number, nil),
					"middleware_script": tftypes.NewValue(tftypes.Number, 12345),
				}),
				"routing": tftypes.NewValue(routingType, nil),
			},
		},
	}

	configSchema := schema.Schema{
		Blocks: map[string]schema.Block{
			"origin": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{},
					"url": schema.StringAttribute{
						CustomType: customtype.PullzoneOriginUrlType{},
					},
					"storagezone":       schema.Int64Attribute{},
					"script":            schema.Int64Attribute{},
					"middleware_script": schema.Int64Attribute{},
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
			"origin":  originType,
			"routing": routingType,
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
