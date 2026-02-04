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
		newConditionValues := make(map[string]tftypes.Value, 4)
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

	newValue := tftypes.NewValue(newType, map[string]tftypes.Value{
		"id":              oldState["id"],
		"pullzone":        oldState["pullzone"],
		"name":            oldState["name"],
		"description":     oldState["description"],
		"transformations": tftypes.NewValue(newType.AttributeTypes["transformations"], newStateTransformations),
		"condition":       tftypes.NewValue(newType.AttributeTypes["condition"], newStateCondition),
		"limit":           oldState["limit"],
		"response":        oldState["response"],
	})

	dv, err := tfprotov6.NewDynamicValue(newType, newValue)
	if err != nil {
		resp.Diagnostics.AddError("Failed to construct upgraded state", err.Error())
		return
	}

	resp.DynamicValue = &dv
}
