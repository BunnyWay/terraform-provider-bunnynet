package pullzoneshieldresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AccessListSetUniqueValues() validator.Set {
	return accessListSet{}
}

type accessListSet struct{}

func (v accessListSet) Description(ctx context.Context) string {
	return "Each access_list can only be defined once"
}

func (v accessListSet) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v accessListSet) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	elements := req.ConfigValue.Elements()
	seen := make(map[int64]struct{}, len(elements))

	for _, element := range elements {
		if element.IsUnknown() {
			return
		}

		attrs := element.(types.Object).Attributes()
		idAttr, ok := attrs["id"]
		if !ok {
			return
		}

		if idAttr.IsUnknown() || idAttr.IsNull() {
			return
		}

		id := idAttr.(types.Int64).ValueInt64()
		if _, ok := seen[id]; ok {
			resp.Diagnostics.AddError("Duplicate access list", fmt.Sprintf("There are multiple access_list entries with id = \"%d\"", id))
			return
		}

		seen[id] = struct{}{}
	}
}
