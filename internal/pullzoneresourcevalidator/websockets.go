package pullzoneresourcevalidator

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func WebsocketsMaxConnections() validator.Int64 {
	return websocketsMaxConnectionsValidator{}
}

type websocketsMaxConnectionsValidator struct{}

func (v websocketsMaxConnectionsValidator) Description(ctx context.Context) string {
	return "Validation for websockets_max_connections"
}

func (v websocketsMaxConnectionsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v websocketsMaxConnectionsValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	value := req.ConfigValue.ValueInt64()

	if value < 1 {
		resp.Diagnostics.AddAttributeError(path.Root("websockets_max_connections"), "Invalid value", "Value must be positive and a multiple of one hundred.")
		return
	}

	rounded := (value + 99) / 100 * 100

	if value != rounded {
		resp.Diagnostics.AddAttributeError(path.Root("websockets_max_connections"), "Invalid value", fmt.Sprintf("Value must be a multiple of one hundred. Use %d instead.", rounded))
	}
}
