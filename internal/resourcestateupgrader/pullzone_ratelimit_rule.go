// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package resourcestateupgrader

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func PullzoneRatelimitRuleV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	oldType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":          tftypes.Number,
			"pullzone":    tftypes.Number,
			"name":        tftypes.String,
			"description": tftypes.String,
			"condition": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"operator":       tftypes.String,
					"value":          tftypes.String,
					"variable":       tftypes.String,
					"variable_value": tftypes.String,
					"transformations": tftypes.List{
						ElementType: tftypes.String,
					},
				},
			},
			"limit": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"requests": tftypes.Number,
					"interval": tftypes.Number,
				},
			},
			"response": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"interval": tftypes.Number,
				},
			},
		},
	}

	newType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":          tftypes.Number,
			"pullzone":    tftypes.Number,
			"name":        tftypes.String,
			"description": tftypes.String,
			"transformations": tftypes.List{
				ElementType: tftypes.String,
			},
			"condition": tftypes.List{
				ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"operator":       tftypes.String,
						"value":          tftypes.String,
						"variable":       tftypes.String,
						"variable_value": tftypes.String,
						"negated":        tftypes.Bool,
					},
				},
			},
			"limit": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"requests":    tftypes.Number,
					"interval":    tftypes.Number,
					"counter_key": tftypes.String,
				},
			},
			"response": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"interval": tftypes.Number,
					"action":   tftypes.String,
				},
			},
		},
	}

	oldRawValue, err := req.RawState.Unmarshal(oldType)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal prior state", err.Error())
		return
	}

	var oldState map[string]tftypes.Value
	if err := oldRawValue.As(&oldState); err != nil {
		resp.Diagnostics.AddError("Failed to convert old state", err.Error())
		return
	}

	var oldCondition map[string]tftypes.Value
	if err := oldState["condition"].As(&oldCondition); err != nil {
		resp.Diagnostics.AddError("Failed to convert old state", err.Error())
		return
	}

	newStateCondition := make([]tftypes.Value, 0, 1)
	var newStateTransformations []tftypes.Value

	{
		newConditionValues := make(map[string]tftypes.Value, 5)
		for _, attribute := range []string{"operator", "variable", "variable_value", "value"} {
			var value string
			err := oldCondition[attribute].As(&value)
			if err != nil {
				resp.Diagnostics.AddError("Failed to convert old state", err.Error())
				return
			}

			if value == "" {
				newConditionValues[attribute] = tftypes.NewValue(tftypes.String, nil)
			} else {
				newConditionValues[attribute] = tftypes.NewValue(tftypes.String, value)
			}
		}

		newConditionValues["negated"] = tftypes.NewValue(tftypes.Bool, false)

		newStateCondition = append(newStateCondition, tftypes.NewValue(newType.AttributeTypes["condition"].(tftypes.List).ElementType, newConditionValues))
	}

	{
		var transformations []tftypes.Value
		err := oldCondition["transformations"].As(&transformations)
		if err != nil {
			resp.Diagnostics.AddError("Failed to convert old state", err.Error())
			return
		}

		newStateTransformations = append(newStateTransformations, transformations...)
	}

	var oldLimit map[string]tftypes.Value
	if err := oldState["limit"].As(&oldLimit); err != nil {
		resp.Diagnostics.AddError("Failed to convert old state", err.Error())
		return
	}

	newStateLimit := tftypes.NewValue(newType.AttributeTypes["limit"], map[string]tftypes.Value{
		"requests":    oldLimit["requests"],
		"interval":    oldLimit["interval"],
		"counter_key": tftypes.NewValue(tftypes.String, "IP"),
	})

	var oldResponse map[string]tftypes.Value
	if err := oldState["response"].As(&oldResponse); err != nil {
		resp.Diagnostics.AddError("Failed to convert old state", err.Error())
		return
	}

	newStateResponse := tftypes.NewValue(newType.AttributeTypes["response"], map[string]tftypes.Value{
		"interval": oldResponse["interval"],
		"action":   tftypes.NewValue(tftypes.String, "RateLimit"),
	})

	newValue := tftypes.NewValue(newType, map[string]tftypes.Value{
		"id":              oldState["id"],
		"pullzone":        oldState["pullzone"],
		"name":            oldState["name"],
		"description":     oldState["description"],
		"transformations": tftypes.NewValue(newType.AttributeTypes["transformations"], newStateTransformations),
		"condition":       tftypes.NewValue(newType.AttributeTypes["condition"], newStateCondition),
		"limit":           newStateLimit,
		"response":        newStateResponse,
	})

	dv, err := tfprotov6.NewDynamicValue(newType, newValue)
	if err != nil {
		resp.Diagnostics.AddError("Failed to construct upgraded state", err.Error())
		return
	}

	resp.DynamicValue = &dv
}
