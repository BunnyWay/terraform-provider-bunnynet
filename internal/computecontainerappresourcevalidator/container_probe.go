package computecontainerappresourcevalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

func ContainerProbe() validator.Object {
	return containerProbe{}
}

type ContainerProbeType string

const ContainerProbeTypeHttp ContainerProbeType = "http"
const ContainerProbeTypeTcp ContainerProbeType = "tcp"
const ContainerProbeTypeGrpc ContainerProbeType = "grpc"

const errProbeTypeNull = "probe type must be set"
const errProbeType = "probe type must be one of: \"http\", \"tcp\", \"grpc\""
const errProbePort = "Attribute \"port\" must be between 0 and 65535"
const errProbeHttpBlockMissing = "Missing http block"
const errProbeGrpcBlockMissing = "Missing grpc block"
const errProbeHttpBlockExtra = "Unexpected http block"
const errProbeGrpcBlockExtra = "Unexpected grpc block"
const errProbeHttpBlockTooMany = "Too many http blocks"
const errProbeGrpcBlockTooMany = "Too many grpc blocks"

var probeTypes = []string{
	string(ContainerProbeTypeHttp),
	string(ContainerProbeTypeTcp),
	string(ContainerProbeTypeGrpc),
}

type containerProbe struct{}

func (c containerProbe) Description(ctx context.Context) string {
	return "Validates probe configuration"
}

func (c containerProbe) MarkdownDescription(ctx context.Context) string {
	return c.Description(ctx)
}

func (c containerProbe) ValidateObject(ctx context.Context, request validator.ObjectRequest, response *validator.ObjectResponse) {
	if request.ConfigValue.IsNull() {
		return
	}

	obj, _ := request.ConfigValue.ToObjectValue(ctx)
	attrs := obj.Attributes()

	if attrs["type"].IsNull() {
		response.Diagnostics.AddError("Invalid probe configuration", errProbeTypeNull)
		return
	}

	probeType := attrs["type"].(types.String).ValueString()
	probePort := attrs["port"].(types.Int64).ValueInt64()

	if !slices.Contains(probeTypes, probeType) {
		response.Diagnostics.AddError("Invalid probe configuration", errProbeType)
		return
	}

	if attrs["port"].IsNull() || probePort < 0 || probePort > 65535 {
		response.Diagnostics.AddError("Invalid probe configuration", errProbePort)
	}

	httpElements := attrs["http"].(types.Set).Elements()
	grpcElements := attrs["grpc"].(types.Set).Elements()

	if probeType == "http" {
		if len(httpElements) == 0 {
			response.Diagnostics.AddError("Invalid probe configuration", errProbeHttpBlockMissing)
		}

		if len(httpElements) > 1 {
			response.Diagnostics.AddError("Invalid probe configuration", errProbeHttpBlockTooMany)
		}
	}

	if probeType == "grpc" {
		if len(grpcElements) == 0 {
			response.Diagnostics.AddError("Invalid probe configuration", errProbeGrpcBlockMissing)
		}

		if len(grpcElements) > 1 {
			response.Diagnostics.AddError("Invalid probe configuration", errProbeGrpcBlockTooMany)
		}
	}

	if probeType != "http" && len(httpElements) > 0 {
		response.Diagnostics.AddError("Invalid probe configuration", errProbeHttpBlockExtra)
	}

	if probeType != "grpc" && len(grpcElements) > 0 {
		response.Diagnostics.AddError("Invalid probe configuration", errProbeGrpcBlockExtra)
	}
}
