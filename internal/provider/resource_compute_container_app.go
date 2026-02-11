// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/computecontainerappresourcevalidator"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"regexp"
	"strconv"
)

var _ resource.Resource = &ComputeContainerAppResource{}
var _ resource.ResourceWithImportState = &ComputeContainerAppResource{}
var _ resource.ResourceWithModifyPlan = &ComputeContainerAppResource{}
var _ resource.ResourceWithConfigValidators = &ComputeContainerAppResource{}

func NewComputeContainerAppResource() resource.Resource {
	return &ComputeContainerAppResource{}
}

type ComputeContainerAppResource struct {
	client *api.Client
}

type ComputeContainerAppResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Version           types.Int64  `tfsdk:"version"`
	Name              types.String `tfsdk:"name"`
	AutoscalingMin    types.Int64  `tfsdk:"autoscaling_min"`
	AutoscalingMax    types.Int64  `tfsdk:"autoscaling_max"`
	RegionsAllowed    types.Set    `tfsdk:"regions_allowed"`
	RegionsRequired   types.Set    `tfsdk:"regions_required"`
	RegionsMaxAllowed types.Int64  `tfsdk:"regions_max_allowed"`
	Containers        types.List   `tfsdk:"container"`
	Volumes           types.List   `tfsdk:"volume"`
}

var computeContainerAppImagePullPolicyOptions = []string{"Always", "IfNotPresent"}

// format: key: terraform, value: api
var computeContainerAppContainerEndpointTypeMap = map[string]string{
	"CDN":        "CDN",
	"Anycast":    "Anycast",
	"InternalIP": "PublicIp", // @TODO replace with InternalIp
}

// format: key: terraform, value: api
var computeContainerAppContainerEndpointProtocolMap = map[string]string{
	"TCP": "Tcp",
	"UDP": "Udp",
}

func (r *ComputeContainerAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_container_app"
}

var computeContainerAppContainerEndpointCdnType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"origin_ssl":  types.BoolType,
		"pullzone_id": types.Int64Type,
		"sticky_sessions": types.ListType{
			ElemType: computeContainerAppContainerEndpointCdnStickySessionsType,
		},
	},
}

var computeContainerAppContainerEndpointCdnStickySessionsType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"headers": types.SetType{ElemType: types.StringType},
	},
}

var computeContainerAppContainerEndpointPortType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"container": types.Int64Type,
		"exposed":   types.Int64Type,
		"protocols": types.SetType{ElemType: types.StringType},
	},
}

var computeContainerAppContainerEndpointType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name": types.StringType,
		"type": types.StringType,
		"cdn": types.ListType{
			ElemType: computeContainerAppContainerEndpointCdnType,
		},
		"port": types.ListType{
			ElemType: computeContainerAppContainerEndpointPortType,
		},
	},
}

var computeContainerAppContainerEnvType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	},
}

var computeContainerAppContainerVolumeMountType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name": types.StringType,
		"path": types.StringType,
	},
}

var containerAppContainerProbeTypeOptions = []string{
	"http",
	"tcp",
	"grpc",
}

var computeContainerAppContainerProbeHttpStatusMap = map[int64]string{
	100: "Continue",
	101: "SwitchingProtocols",
	102: "Processing",
	103: "EarlyHints",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "NonAuthoritativeInformation",
	204: "NoContent",
	205: "ResetContent",
	206: "PartialContent",
	207: "MultiStatus",
	208: "AlreadyReported",
	226: "IMUsed",
	300: "MultipleChoices",
	301: "MovedPermanently",
	302: "Found",
	303: "SeeOther",
	304: "NotModified",
	305: "UseProxy",
	306: "Unused",
	307: "TemporaryRedirect",
	308: "PermanentRedirect",
	400: "BadRequest",
	401: "Unauthorized",
	402: "PaymentRequired",
	403: "Forbidden",
	404: "NotFound",
	405: "MethodNotAllowed",
	406: "NotAcceptable",
	407: "ProxyAuthenticationRequired",
	408: "RequestTimeout",
	409: "Conflict",
	410: "Gone",
	411: "LengthRequired",
	412: "PreconditionFailed",
	413: "RequestEntityTooLarge",
	414: "RequestUriTooLong",
	415: "UnsupportedMediaType",
	416: "RequestedRangeNotSatisfiable",
	417: "ExpectationFailed",
	421: "MisdirectedRequest",
	422: "UnprocessableEntity",
	423: "Locked",
	424: "FailedDependency",
	426: "UpgradeRequired",
	428: "PreconditionRequired",
	429: "TooManyRequests",
	431: "RequestHeaderFieldsTooLarge",
	451: "UnavailableForLegalReasons",
	500: "InternalServerError",
	501: "NotImplemented",
	502: "BadGateway",
	503: "ServiceUnavailable",
	504: "GatewayTimeout",
	505: "HttpVersionNotSupported",
	506: "VariantAlsoNegotiates",
	507: "InsufficientStorage",
	508: "LoopDetected",
	510: "NotExtended",
	511: "NetworkAuthenticationRequired",
}

var computeContainerAppContainerProbeHttpType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"path":            types.StringType,
		"expected_status": types.Int64Type,
	},
}

var computeContainerAppContainerProbeGrpcType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"service": types.StringType,
	},
}

var computeContainerAppContainerProbeType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"type":              types.StringType,
		"port":              types.Int64Type,
		"initial_delay":     types.Int64Type,
		"period":            types.Int64Type,
		"timeout":           types.Int64Type,
		"failure_threshold": types.Int64Type,
		"success_threshold": types.Int64Type,
		"http": types.ListType{
			ElemType: computeContainerAppContainerProbeHttpType,
		},
		"grpc": types.ListType{
			ElemType: computeContainerAppContainerProbeGrpcType,
		},
	},
}

var computeContainerAppContainerType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":                types.StringType,
		"name":              types.StringType,
		"image_registry":    types.Int64Type,
		"image_namespace":   types.StringType,
		"image_name":        types.StringType,
		"image_tag":         types.StringType,
		"image_pull_policy": types.StringType,
		"command":           types.StringType,
		"arguments":         types.StringType,
		"working_dir":       types.StringType,
		"endpoint":          types.ListType{ElemType: computeContainerAppContainerEndpointType},
		"env":               types.ListType{ElemType: computeContainerAppContainerEnvType},
		"volumemount":       types.ListType{ElemType: computeContainerAppContainerVolumeMountType},
		"startup_probe":     types.ListType{ElemType: computeContainerAppContainerProbeType},
		"readiness_probe":   types.ListType{ElemType: computeContainerAppContainerProbeType},
		"liveness_probe":    types.ListType{ElemType: computeContainerAppContainerProbeType},
	},
}

var computeContainerAppVolumeType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name": types.StringType,
		"size": types.Int64Type,
	},
}

func (r *ComputeContainerAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	containerAppContainerProbeSchema := schema.NestedBlockObject{
		Validators: []validator.Object{
			computecontainerappresourcevalidator.ContainerProbe(),
		},
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required:    true,
				Description: generateMarkdownSliceOptions(containerAppContainerProbeTypeOptions),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(containerAppContainerProbeTypeOptions...),
				},
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "The port within the container the probe will connect to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"initial_delay": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The amount of time in seconds after the container is started to wait before the first probe is sent.",
				// @TODO nestedObjects don't support default values: https://github.com/hashicorp/terraform-plugin-framework/issues/726
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"period": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				// @TODO nestedObjects don't support default values: https://github.com/hashicorp/terraform-plugin-framework/issues/726
				Description: "The amount of time in seconds between each probe.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				// @TODO nestedObjects don't support default values: https://github.com/hashicorp/terraform-plugin-framework/issues/726
				Description: "The amount of time in seconds the probe will wait for a response before considering it a failure.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"failure_threshold": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				// @TODO nestedObjects don't support default values: https://github.com/hashicorp/terraform-plugin-framework/issues/726
				Description: "The number of failed probes to consider the container unhealthy.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"success_threshold": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				// @TODO nestedObjects don't support default values: https://github.com/hashicorp/terraform-plugin-framework/issues/726
				Description: "The number of successful probes to consider the container healthy.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
		},
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.UseStateForUnknown(),
		},
		Blocks: map[string]schema.Block{
			"http": schema.ListNestedBlock{
				Description: "HTTP-specific configurations.",
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Required:    true,
							Description: "The HTTP path to be requested.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^/`), ""),
							},
						},
						"expected_status": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The expected HTTP response status code.",
							Validators: []validator.Int64{
								int64validator.OneOf(maps.Keys(computeContainerAppContainerProbeHttpStatusMap)...),
							},
						},
					},
				},
			},
			"grpc": schema.ListNestedBlock{
				Description: "gRPC-specific configurations.",
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"service": schema.StringAttribute{
							Required:    true,
							Description: "The gRPC service name.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "This resource manages a Magic Containers application in bunny.net.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the application.",
			},
			"version": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					computecontainerappresourcevalidator.Version(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9-\s]+$`), ""),
				},
				Description: "The name of the application.",
			},
			"autoscaling_min": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The minimum number of instances that will be provisioned per active region.",
			},
			"autoscaling_max": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The maximum number of instances that will be provisioned per active region.",
			},
			"regions_allowed": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				Description: "The regions that will be dynamically provisionable based on the user latency.",
			},
			"regions_required": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
				Description: "The regions that will be statically provisioned and will always be running and available to users.",
			},
			"regions_max_allowed": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				Description: "The maximum amount of regions to be deployed at any given time.",
			},
		},
		Blocks: map[string]schema.Block{
			// a set does not work here, as the "id" attribute will not be copied over from state to plan (even though we have useStateForUnknown)
			// @see https://github.com/hashicorp/terraform-plugin-framework/issues/726
			"container": schema.ListNestedBlock{
				Description: "Defines a container for the application.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					PlanModifiers: []planmodifier.Object{
						objectplanmodifier.UseStateForUnknown(),
					},
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseNonNullStateForUnknown(),
							},
							Description: "The unique identifier for the container.",
						},
						"name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
								stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9-]+$"), "name should only contain letters, numbers and dash"),
							},
							Description: "The name of the container.",
						},
						"image_registry": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
							Description: "The image registry for the container.",
						},
						"image_namespace": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Description: "The image namespace within the registry, without the domain prefix (i.e.: `my-org`).",
						},
						"image_name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Description: "The image name within the registry, without the domain prefix (i.e.: `my-app`).",
						},
						"image_tag": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							MarkdownDescription: "The image tag (i.e.: `2.9-alpine`).",
						},
						"image_pull_policy": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("Always"),
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(computeContainerAppImagePullPolicyOptions...),
							},
							Description: generateMarkdownSliceOptions(computeContainerAppImagePullPolicyOptions),
						},
						"command": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
							Description: "A custom startup command that will execute once the container is launched.",
						},
						"arguments": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
							Description: "The arguments that will be added to the container entry point when starting the image.",
						},
						"working_dir": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
							Description: "The working directory of the container runtime.",
						},
					},
					Blocks: map[string]schema.Block{
						"endpoint": schema.ListNestedBlock{
							Description: "Defines a public endpoint for the application.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.Object{
									computecontainerappresourcevalidator.ContainerEndpoint(),
								},
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9-]+$"), "endpoint.name should only contain letters, numbers and dash"),
										},
										Description: "The name of the endpoint.",
									},
									"type": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.String{
											stringvalidator.OneOf(maps.Keys(computeContainerAppContainerEndpointTypeMap)...),
										},
										Description: generateMarkdownMapOptions(utils.MapInvert(computeContainerAppContainerEndpointTypeMap)),
									},
								},
								Blocks: map[string]schema.Block{
									"cdn": schema.ListNestedBlock{
										Description: "Configurations for CDN endpoint.",
										Validators: []validator.List{
											listvalidator.SizeBetween(0, 1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"origin_ssl": schema.BoolAttribute{
													Optional: true,
													Computed: true,
													Default:  booldefault.StaticBool(false),
													PlanModifiers: []planmodifier.Bool{
														boolplanmodifier.UseStateForUnknown(),
													},
													Description: "Indicates whether the container will handle TLS termination.",
												},
												"pullzone_id": schema.Int64Attribute{
													Computed:    true,
													Description: "The ID of the pullzone associated with the endpoint.",
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseNonNullStateForUnknown(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"sticky_sessions": schema.ListNestedBlock{
													Description: "Indicates whether sticky sessions is enabled.",
													Validators: []validator.List{
														listvalidator.SizeBetween(0, 1),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.UseStateForUnknown(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"headers": schema.SetAttribute{
																ElementType: types.StringType,
																Required:    true,
																PlanModifiers: []planmodifier.Set{
																	setplanmodifier.UseStateForUnknown(),
																},
																Validators: []validator.Set{
																	setvalidator.SizeAtLeast(1),
																	setvalidator.ValueStringsAre(
																		stringvalidator.LengthAtLeast(1),
																	),
																},
																Description: "Incoming request headers used to select a pod for sticky sessions.",
															},
														},
													},
												},
											},
										},
									},
									"port": schema.ListNestedBlock{
										Description: "Endpoint port configuration.",
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"container": schema.Int64Attribute{
													Required: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
													Validators: []validator.Int64{
														int64validator.Between(0, 65535),
													},
													Description: "The container port number.",
												},
												"exposed": schema.Int64Attribute{
													Optional: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.UseStateForUnknown(),
													},
													Validators: []validator.Int64{
														int64validator.Between(1, 65535),
													},
													Description: "The exposed port number.",
												},
												"protocols": schema.SetAttribute{
													Optional:    true,
													ElementType: types.StringType,
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.UseStateForUnknown(),
													},
													Validators: []validator.Set{
														setvalidator.SizeAtLeast(1),
														setvalidator.ValueStringsAre(
															stringvalidator.OneOf(maps.Keys(computeContainerAppContainerEndpointProtocolMap)...),
														),
													},
													Description: generateMarkdownMapOptions(computeContainerAppContainerEndpointProtocolMap),
												},
											},
											PlanModifiers: []planmodifier.Object{
												objectplanmodifier.UseStateForUnknown(),
											},
										},
									},
								},
							},
						},
						"env": schema.ListNestedBlock{
							Description: "Defines an environment variable for the container",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(),
								},
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$"), "env.name should only contain letters, numbers and underline"),
										},
										Description: "The name of the environment variable.",
									},
									"value": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Description: "The value of the environment variable.",
									},
								},
							},
						},
						"volumemount": schema.ListNestedBlock{
							Description: "Mounts a volume within a container",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: schema.NestedBlockObject{
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(),
								},
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$"), "Invalid name"),
										},
										Description: "The name of the volume.",
									},
									"path": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.UseStateForUnknown(),
										},
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
											stringvalidator.RegexMatches(regexp.MustCompile(`^/[^\t\n\f\r ]*$`), "Invalid path"),
										},
										Description: "The path within the container where the volume will be mounted.",
									},
								},
							},
						},
						"startup_probe": schema.ListNestedBlock{
							Description: "Checks if the application has successfully started. No requests will be routed to the application until this check is successful.",
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: containerAppContainerProbeSchema,
						},
						"readiness_probe": schema.ListNestedBlock{
							Description: "Checks if the application is fully prepared to handle incoming requests. No requests will be routed to the application until this check is successful.",
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: containerAppContainerProbeSchema,
						},
						"liveness_probe": schema.ListNestedBlock{
							Description: "Checks that the application is actively running without issues. It the check fails, the container will be automatically restarted",
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							NestedObject: containerAppContainerProbeSchema,
						},
					},
				},
			},
			"volume": schema.ListNestedBlock{
				Description: "Defines a persistent volume to be used by the application.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					PlanModifiers: []planmodifier.Object{
						objectplanmodifier.UseStateForUnknown(),
					},
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
								stringvalidator.RegexMatches(regexp.MustCompile("^[a-zA-Z0-9-]+$"), "name should only contain letters, numbers and dash"),
							},
							Description: "The name of the volume.",
						},
						"size": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
								int64validator.AtMost(100),
							},
							Description: "The size of the volume, in Gigabytes (10^9 bytes).",
						},
					},
				},
			},
		},
	}
}

func (r *ComputeContainerAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ComputeContainerAppResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		computecontainerappresourcevalidator.RegionRequiredMustAlsoBeAllowed(),
		computecontainerappresourcevalidator.EndpointNameShouldBeUnique(),
		computecontainerappresourcevalidator.ContainerVolumeMounts(),
		computecontainerappresourcevalidator.VolumeNamesShouldBeUnique(),
	}
}

func (r *ComputeContainerAppResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var dataTf ComputeContainerAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	containerElements := dataTf.Containers.Elements()
	for i, container := range containerElements {
		containerAttr := container.(types.Object).Attributes()

		containerObj, diags := types.ObjectValue(computeContainerAppContainerType.AttrTypes, containerAttr)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		containerElements[i] = containerObj
	}

	containerList, diags := types.ListValue(computeContainerAppContainerType, containerElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	dataTf.Containers = containerList
	resp.Plan.Set(ctx, dataTf)
}

func (r *ComputeContainerAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf ComputeContainerAppResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, diags := r.convertModelToApi(ctx, dataTf)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	dataApi, err := r.client.CreateComputeContainerApp(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create container app", err.Error())
		return
	}

	tflog.Trace(ctx, "created container app "+dataApi.Name)
	dataTf, diags = r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComputeContainerAppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetComputeContainerApp(ctx, data.Id.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching container app", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ComputeContainerAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, diags := r.convertModelToApi(ctx, data)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	dataApi, err := r.client.UpdateComputeContainerApp(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating container app", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComputeContainerAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteComputeContainerApp(data.Id.ValueString())
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting container app", err.Error()))
	}
}

func (r *ComputeContainerAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	dataApi, err := r.client.GetComputeContainerApp(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching container app", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *ComputeContainerAppResource) convertModelToApi(ctx context.Context, dataTf ComputeContainerAppResourceModel) (api.ComputeContainerApp, diag.Diagnostics) {
	dataApi := api.ComputeContainerApp{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.AutoScaling.Min = dataTf.AutoscalingMin.ValueInt64()
	dataApi.AutoScaling.Max = dataTf.AutoscalingMax.ValueInt64()
	dataApi.RegionSettings.AllowedRegionIds = utils.ConvertSetToStringSlice(dataTf.RegionsAllowed)
	dataApi.RegionSettings.RequiredRegionIds = utils.ConvertSetToStringSlice(dataTf.RegionsRequired)
	dataApi.RegionSettings.MaxAllowedRegions = dataTf.RegionsMaxAllowed.ValueInt64Pointer()
	dataApi.ContainerTemplates = make([]api.ComputeContainerAppContainer, len(dataTf.Containers.Elements()))
	dataApi.Volumes = make([]api.ComputeContainerAppVolume, len(dataTf.Volumes.Elements()))

	for i, container := range dataTf.Containers.Elements() {
		cAttr := container.(types.Object).Attributes()
		endpointElements := make([]api.ComputeContainerAppContainerEndpoint, 0, len(cAttr["endpoint"].(types.List).Elements()))
		envElements := make([]api.ComputeContainerAppContainerEnv, len(cAttr["env"].(types.List).Elements()))
		volumeMountElements := make([]api.ComputeContainerAppContainerVolumeMount, len(cAttr["volumemount"].(types.List).Elements()))

		for _, endpoint := range cAttr["endpoint"].(types.List).Elements() {
			endpointAttr := endpoint.(types.Object).Attributes()
			endpointType := endpointAttr["type"].(types.String).ValueString()

			originSsl := false
			stickySessions := false
			stickySessionsHeaders := []string{}

			if endpointType == "CDN" {
				cdnList := endpointAttr["cdn"].(types.List).Elements()
				if len(cdnList) > 0 {
					cdnAttr := cdnList[0].(types.Object).Attributes()
					if cdnAttr["origin_ssl"].(types.Bool).ValueBool() {
						originSsl = true
					}

					if !cdnAttr["sticky_sessions"].(types.List).IsNull() {
						stickySessionsList := cdnAttr["sticky_sessions"].(types.List).Elements()
						stickySessionsAttr := stickySessionsList[0].(types.Object).Attributes()
						stickySessions = true

						for _, header := range stickySessionsAttr["headers"].(types.Set).Elements() {
							stickySessionsHeaders = append(stickySessionsHeaders, header.(types.String).ValueString())
						}
					}
				}
			}

			portElements := endpointAttr["port"].(types.List).Elements()
			portMappings := make([]api.ComputeContainerAppContainerEndpointPortMapping, 0, len(portElements))

			for _, port := range portElements {
				portAttr := port.(types.Object).Attributes()
				protocolsElements := portAttr["protocols"].(types.Set).Elements()
				protocolsValues := make([]string, 0, len(protocolsElements))

				for _, protocol := range protocolsElements {
					protocolsValues = append(protocolsValues, computeContainerAppContainerEndpointProtocolMap[protocol.(types.String).ValueString()])
				}

				if endpointType == "CDN" {
					protocolsValues = []string{"Tcp"}
				}

				portMappings = append(portMappings, api.ComputeContainerAppContainerEndpointPortMapping{
					ContainerPort: portAttr["container"].(types.Int64).ValueInt64(),
					ExposedPort:   portAttr["exposed"].(types.Int64).ValueInt64(),
					Protocols:     protocolsValues,
				})
			}

			var stickySessionsElement *api.ComputeContainerAppContainerEndpointStickySessions
			if stickySessions {
				stickySessionsElement = &api.ComputeContainerAppContainerEndpointStickySessions{
					Enabled:        stickySessions,
					SessionHeaders: stickySessionsHeaders,
				}
			}

			endpointElements = append(endpointElements, api.ComputeContainerAppContainerEndpoint{
				DisplayName:    endpointAttr["name"].(types.String).ValueString(),
				Type:           computeContainerAppContainerEndpointTypeMap[endpointType],
				PortMappings:   portMappings,
				IsSslEnabled:   originSsl,
				StickySessions: stickySessionsElement,
			})
		}

		for iEnv, env := range cAttr["env"].(types.List).Elements() {
			envElements[iEnv] = api.ComputeContainerAppContainerEnv{
				Name:  env.(types.Object).Attributes()["name"].(types.String).ValueString(),
				Value: env.(types.Object).Attributes()["value"].(types.String).ValueString(),
			}
		}

		for iMount, mount := range cAttr["volumemount"].(types.List).Elements() {
			volumeMountElements[iMount] = api.ComputeContainerAppContainerVolumeMount{
				Name: mount.(types.Object).Attributes()["name"].(types.String).ValueString(),
				Path: mount.(types.Object).Attributes()["path"].(types.String).ValueString(),
			}
		}

		startupProbe, diags := r.convertModelContainerProbeToApi(cAttr["startup_probe"].(types.List))
		if diags != nil {
			return dataApi, diags
		}

		readinessProbe, diags := r.convertModelContainerProbeToApi(cAttr["readiness_probe"].(types.List))
		if diags != nil {
			return dataApi, diags
		}

		livenessProbe, diags := r.convertModelContainerProbeToApi(cAttr["liveness_probe"].(types.List))
		if diags != nil {
			return dataApi, diags
		}

		probes := api.ComputeContainerAppContainerProbes{
			Startup:   startupProbe,
			Readiness: readinessProbe,
			Liveness:  livenessProbe,
		}

		dataApi.ContainerTemplates[i] = api.ComputeContainerAppContainer{
			Id:              cAttr["id"].(types.String).ValueString(),
			Name:            cAttr["name"].(types.String).ValueString(),
			PackageId:       "d2c1b95e-65e7-42a3-adb3-d27d9d1a4f72",
			ImageRegistryId: fmt.Sprintf("%d", cAttr["image_registry"].(types.Int64).ValueInt64()),
			ImageNamespace:  cAttr["image_namespace"].(types.String).ValueString(),
			ImageName:       cAttr["image_name"].(types.String).ValueString(),
			ImageTag:        cAttr["image_tag"].(types.String).ValueString(),
			ImagePullPolicy: cAttr["image_pull_policy"].(types.String).ValueString(),
			EntryPoint: api.ComputeContainerAppContainerEntrypoint{
				Command:          cAttr["command"].(types.String).ValueString(),
				Arguments:        cAttr["arguments"].(types.String).ValueString(),
				WorkingDirectory: cAttr["working_dir"].(types.String).ValueString(),
			},
			Probes:               probes,
			Endpoints:            endpointElements,
			EnvironmentVariables: envElements,
			VolumeMounts:         volumeMountElements,
		}
	}

	for i, volume := range dataTf.Volumes.Elements() {
		vAttr := volume.(types.Object).Attributes()

		dataApi.Volumes[i] = api.ComputeContainerAppVolume{
			Name: vAttr["name"].(types.String).ValueString(),
			Size: vAttr["size"].(types.Int64).ValueInt64(),
		}
	}

	return dataApi, nil
}

func (r *ComputeContainerAppResource) convertModelContainerProbeToApi(dataTf types.List) (*api.ComputeContainerAppContainerProbe, diag.Diagnostics) {
	if dataTf.IsNull() {
		return nil, nil
	}

	items := dataTf.Elements()
	if len(items) != 1 {
		diags := diag.Diagnostics{}
		diags.AddError("Unexpected number of probes defined", "A single probe of each type should be defined")
		return nil, diags
	}

	attrs := items[0].(types.Object).Attributes()
	probeType := attrs["type"].(types.String).ValueString()
	port := attrs["port"].(types.Int64).ValueInt64()

	result := &api.ComputeContainerAppContainerProbe{
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		TimeoutSeconds:      3,
		FailureThreshold:    3,
		SuccessThreshold:    1,
	}

	if v := attrs["initial_delay"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		result.InitialDelaySeconds = v.ValueInt64()
	}

	if v := attrs["period"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		result.PeriodSeconds = v.ValueInt64()
	}

	if v := attrs["timeout"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		result.TimeoutSeconds = v.ValueInt64()
	}

	if v := attrs["failure_threshold"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		result.FailureThreshold = v.ValueInt64()
	}

	if v := attrs["success_threshold"].(types.Int64); !v.IsNull() && !v.IsUnknown() {
		result.SuccessThreshold = v.ValueInt64()
	}

	switch computecontainerappresourcevalidator.ContainerProbeType(probeType) {
	case computecontainerappresourcevalidator.ContainerProbeTypeHttp:
		httpBlocks := attrs["http"].(types.List).Elements()
		if len(httpBlocks) != 1 {
			diags := diag.Diagnostics{}
			diags.AddError("Unexpected number of http blocks defined", "A single http block should be defined")
			return nil, diags
		}

		httpAttr := httpBlocks[0].(types.Object).Attributes()

		var expectedStatus *string
		if httpAttr["expected_status"].(types.Int64).IsNull() || httpAttr["expected_status"].(types.Int64).IsUnknown() {
			expectedStatus = nil
		} else {
			value := mapKeyToValue(computeContainerAppContainerProbeHttpStatusMap, httpAttr["expected_status"].(types.Int64).ValueInt64())
			expectedStatus = &value
		}

		result.HttpGet = &api.ComputeContainerAppContainerProbeHttp{
			Request: api.ComputeContainerAppContainerProbeHttpRequest{
				PortNumber: port,
				Path:       httpAttr["path"].(types.String).ValueString(),
			},
			Response: api.ComputeContainerAppContainerProbeHttpResponse{
				ExpectedStatusCode: expectedStatus,
			},
		}
	case computecontainerappresourcevalidator.ContainerProbeTypeTcp:
		result.TcpSocket = &api.ComputeContainerAppContainerProbeTcp{
			Request: api.ComputeContainerAppContainerProbeTcpRequest{
				PortNumber: port,
			},
		}
	case computecontainerappresourcevalidator.ContainerProbeTypeGrpc:
		grpcBlocks := attrs["grpc"].(types.List).Elements()
		if len(grpcBlocks) != 1 {
			diags := diag.Diagnostics{}
			diags.AddError("Unexpected number of grpc blocks defined", "A single grpc block should be defined")
			return nil, diags
		}

		grpcAttr := grpcBlocks[0].(types.Object).Attributes()
		result.Grpc = &api.ComputeContainerAppContainerProbeGrpc{
			Request: api.ComputeContainerAppContainerProbeGrpcRequest{
				PortNumber:  port,
				ServiceName: grpcAttr["service"].(types.String).ValueString(),
			},
		}
	default:
		diags := diag.Diagnostics{}
		diags.AddError("Invalid probe type", probeType)
		return nil, diags
	}

	return result, nil
}

func (r *ComputeContainerAppResource) convertApiToModel(ctx context.Context, dataApi api.ComputeContainerApp) (ComputeContainerAppResourceModel, diag.Diagnostics) {
	endpointProtocolMap := utils.MapInvert(computeContainerAppContainerEndpointProtocolMap)
	endpointTypeMap := utils.MapInvert(computeContainerAppContainerEndpointTypeMap)

	var diags diag.Diagnostics
	dataTf := ComputeContainerAppResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.Version = types.Int64Value(2)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.AutoscalingMin = types.Int64Value(dataApi.AutoScaling.Min)
	dataTf.AutoscalingMax = types.Int64Value(dataApi.AutoScaling.Max)

	if dataApi.RegionSettings.MaxAllowedRegions == nil {
		dataTf.RegionsMaxAllowed = types.Int64Null()
	} else {
		dataTf.RegionsMaxAllowed = types.Int64Value(*dataApi.RegionSettings.MaxAllowedRegions)
	}

	{
		regions, diags := utils.ConvertStringSliceToSet(dataApi.RegionSettings.AllowedRegionIds)
		if diags.HasError() {
			return dataTf, diags
		}

		dataTf.RegionsAllowed = regions
	}

	{
		regions, diags := utils.ConvertStringSliceToSet(dataApi.RegionSettings.RequiredRegionIds)
		if diags.HasError() {
			return dataTf, diags
		}

		dataTf.RegionsRequired = regions
	}

	if len(dataApi.ContainerTemplates) == 0 {
		dataTf.Containers = types.ListNull(computeContainerAppContainerType)
	} else {
		containers := make([]attr.Value, len(dataApi.ContainerTemplates))
		for i, c := range dataApi.ContainerTemplates {
			imageRegistryId, err := strconv.ParseInt(c.ImageRegistryId, 10, 64)
			if err != nil {
				diags := diag.Diagnostics{}
				diags.AddError("Error converting image registry id", err.Error())
				return dataTf, diags
			}

			// endpoint
			endpointListValues := make([]attr.Value, len(c.Endpoints))
			for iEndpoint, endpoint := range c.Endpoints {
				cdnList := types.ListNull(computeContainerAppContainerEndpointCdnType)

				if endpoint.Type == "CDN" {
					cdnList, diags = convertContainerAppContainerEndpointCdnApiToTf(endpoint)
					if diags.HasError() {
						return dataTf, diags
					}
				}

				portListValues := make([]attr.Value, 0, len(endpoint.PortMappings))
				for _, port := range endpoint.PortMappings {
					var exposedPort types.Int64
					if port.ExposedPort > 0 {
						exposedPort = types.Int64Value(port.ExposedPort)
					} else {
						exposedPort = types.Int64Null()
					}

					protocols := types.SetNull(types.StringType)
					if endpoint.Type != "CDN" {
						protocolsValues := make([]attr.Value, 0, len(port.Protocols))
						for _, protocol := range port.Protocols {
							protocolsValues = append(protocolsValues, types.StringValue(endpointProtocolMap[protocol]))
						}

						protocols, diags = types.SetValue(types.StringType, protocolsValues)
						if diags.HasError() {
							return dataTf, diags
						}
					}

					portObj, diags := types.ObjectValue(computeContainerAppContainerEndpointPortType.AttrTypes, map[string]attr.Value{
						"container": types.Int64Value(port.ContainerPort),
						"exposed":   exposedPort,
						"protocols": protocols,
					})

					if diags.HasError() {
						return dataTf, diags
					}

					portListValues = append(portListValues, portObj)
				}

				portList, diags := types.ListValue(computeContainerAppContainerEndpointPortType, portListValues)
				if diags.HasError() {
					return dataTf, diags
				}

				endpointAttrs := map[string]attr.Value{
					"name": types.StringValue(endpoint.DisplayName),
					"type": types.StringValue(endpointTypeMap[endpoint.Type]),
					"cdn":  cdnList,
					"port": portList,
				}

				endpointObject, diags := types.ObjectValue(computeContainerAppContainerEndpointType.AttrTypes, endpointAttrs)
				if diags.HasError() {
					return dataTf, diags
				}

				endpointListValues[iEndpoint] = endpointObject
			}

			containerEndpoints, diags := types.ListValue(computeContainerAppContainerEndpointType, endpointListValues)
			if diags.HasError() {
				return dataTf, diags
			}

			// env
			envListValues := make([]attr.Value, len(c.EnvironmentVariables))
			for iEnv, env := range c.EnvironmentVariables {
				envObject, diags := types.ObjectValue(computeContainerAppContainerEnvType.AttrTypes, map[string]attr.Value{
					"name":  types.StringValue(env.Name),
					"value": types.StringValue(env.Value),
				})

				if diags.HasError() {
					return dataTf, diags
				}

				envListValues[iEnv] = envObject
			}

			containerEnvs, diags := types.ListValue(computeContainerAppContainerEnvType, envListValues)
			if diags.HasError() {
				return dataTf, diags
			}

			// volumemounts
			volumeMountValues := make([]attr.Value, len(c.VolumeMounts))
			for iMount, mount := range c.VolumeMounts {
				mountObject, diags := types.ObjectValue(computeContainerAppContainerVolumeMountType.AttrTypes, map[string]attr.Value{
					"name": types.StringValue(mount.Name),
					"path": types.StringValue(mount.Path),
				})

				if diags.HasError() {
					return dataTf, diags
				}

				volumeMountValues[iMount] = mountObject
			}

			containerVolumeMounts, diags := types.ListValue(computeContainerAppContainerVolumeMountType, volumeMountValues)
			if diags.HasError() {
				return dataTf, diags
			}

			// probes
			startupProbe, diags := r.convertApiContainerProbeToModel(c.Probes.Startup)
			if diags.HasError() {
				return dataTf, diags
			}

			readinessProbe, diags := r.convertApiContainerProbeToModel(c.Probes.Readiness)
			if diags.HasError() {
				return dataTf, diags
			}

			livenessProbe, diags := r.convertApiContainerProbeToModel(c.Probes.Liveness)
			if diags.HasError() {
				return dataTf, diags
			}

			// container
			container, diags := types.ObjectValue(computeContainerAppContainerType.AttrTypes, map[string]attr.Value{
				"id":                types.StringValue(c.Id),
				"name":              types.StringValue(c.Name),
				"image_registry":    types.Int64Value(imageRegistryId),
				"image_namespace":   types.StringValue(c.ImageNamespace),
				"image_name":        types.StringValue(c.ImageName),
				"image_tag":         types.StringValue(c.ImageTag),
				"image_pull_policy": types.StringValue(c.ImagePullPolicy),
				"command":           typeStringOrNull(c.EntryPoint.Command),
				"arguments":         typeStringOrNull(c.EntryPoint.Arguments),
				"working_dir":       typeStringOrNull(c.EntryPoint.WorkingDirectory),
				"endpoint":          containerEndpoints,
				"env":               containerEnvs,
				"volumemount":       containerVolumeMounts,
				"startup_probe":     startupProbe,
				"readiness_probe":   readinessProbe,
				"liveness_probe":    livenessProbe,
			})

			if diags.HasError() {
				return dataTf, diags
			}

			containers[i] = container
		}

		containerTf, diags := types.ListValue(computeContainerAppContainerType, containers)
		if diags.HasError() {
			return dataTf, diags
		}

		dataTf.Containers = containerTf
	}

	if len(dataApi.Volumes) == 0 {
		dataTf.Volumes = types.ListNull(computeContainerAppVolumeType)
	} else {
		volumes := make([]attr.Value, len(dataApi.Volumes))
		for i, v := range dataApi.Volumes {
			volume, diags := types.ObjectValue(computeContainerAppVolumeType.AttrTypes, map[string]attr.Value{
				"name": types.StringValue(v.Name),
				"size": types.Int64Value(v.Size),
			})

			if diags.HasError() {
				return dataTf, diags
			}

			volumes[i] = volume
		}

		volumeTf, diags := types.ListValue(computeContainerAppVolumeType, volumes)
		if diags.HasError() {
			return dataTf, diags
		}

		dataTf.Volumes = volumeTf
	}

	return dataTf, nil
}

func convertContainerAppContainerEndpointCdnApiToTf(endpoint api.ComputeContainerAppContainerEndpoint) (types.List, diag.Diagnostics) {
	stickySessionsList := types.ListNull(computeContainerAppContainerEndpointCdnStickySessionsType)

	if endpoint.StickySessions != nil {
		headerValues := make([]attr.Value, 0, len(endpoint.StickySessions.SessionHeaders))
		for _, header := range endpoint.StickySessions.SessionHeaders {
			headerValues = append(headerValues, types.StringValue(header))
		}

		headersSet, diags := types.SetValue(types.StringType, headerValues)
		if diags.HasError() {
			return basetypes.ListValue{}, diags
		}

		stickySessionsObj, diags := types.ObjectValue(computeContainerAppContainerEndpointCdnStickySessionsType.AttrTypes, map[string]attr.Value{
			"headers": headersSet,
		})

		if diags.HasError() {
			return basetypes.ListValue{}, diags
		}

		stickySessionsList, diags = types.ListValue(computeContainerAppContainerEndpointCdnStickySessionsType, []attr.Value{stickySessionsObj})
		if diags.HasError() {
			return basetypes.ListValue{}, diags
		}
	}

	pullzoneId, err := strconv.ParseInt(endpoint.PullZoneId, 10, 64)
	if err != nil {
		diags := diag.Diagnostics{}
		diags.AddError("Could not convert pullzone ID to integer", err.Error())
		return basetypes.ListValue{}, diags
	}

	cdnObj, diags := types.ObjectValue(computeContainerAppContainerEndpointCdnType.AttrTypes, map[string]attr.Value{
		"origin_ssl":      types.BoolValue(endpoint.IsSslEnabled),
		"pullzone_id":     types.Int64Value(pullzoneId),
		"sticky_sessions": stickySessionsList,
	})

	if diags.HasError() {
		return basetypes.ListValue{}, diags
	}

	return types.ListValue(computeContainerAppContainerEndpointCdnType, []attr.Value{cdnObj})
}

func (r *ComputeContainerAppResource) convertApiContainerProbeToModel(probe *api.ComputeContainerAppContainerProbe) (basetypes.ListValue, diag.Diagnostics) {
	if probe == nil {
		return types.ListNull(computeContainerAppContainerProbeType), diag.Diagnostics{}
	}

	objAttr := map[string]attr.Value{
		"initial_delay":     types.Int64Value(probe.InitialDelaySeconds),
		"period":            types.Int64Value(probe.PeriodSeconds),
		"timeout":           types.Int64Value(probe.TimeoutSeconds),
		"failure_threshold": types.Int64Value(probe.FailureThreshold),
		"success_threshold": types.Int64Value(probe.SuccessThreshold),
		"http":              types.ListNull(computeContainerAppContainerProbeHttpType),
		"grpc":              types.ListNull(computeContainerAppContainerProbeGrpcType),
	}

	if probe.HttpGet != nil {
		var expectedStatus types.Int64
		if probe.HttpGet.Response.ExpectedStatusCode == nil || *probe.HttpGet.Response.ExpectedStatusCode == "" {
			expectedStatus = types.Int64Null()
		} else {
			expectedStatus = types.Int64Value(mapValueToKey(computeContainerAppContainerProbeHttpStatusMap, *probe.HttpGet.Response.ExpectedStatusCode))
		}

		objAttr["type"] = types.StringValue(string(computecontainerappresourcevalidator.ContainerProbeTypeHttp))
		objAttr["port"] = types.Int64Value(probe.HttpGet.Request.PortNumber)
		objAttr["http"] = types.ListValueMust(computeContainerAppContainerProbeHttpType, []attr.Value{
			types.ObjectValueMust(computeContainerAppContainerProbeHttpType.AttrTypes, map[string]attr.Value{
				"path":            types.StringValue(probe.HttpGet.Request.Path),
				"expected_status": expectedStatus,
			}),
		})
	}

	if probe.TcpSocket != nil {
		objAttr["type"] = types.StringValue(string(computecontainerappresourcevalidator.ContainerProbeTypeTcp))
		objAttr["port"] = types.Int64Value(probe.TcpSocket.Request.PortNumber)
	}

	if probe.Grpc != nil {
		objAttr["type"] = types.StringValue(string(computecontainerappresourcevalidator.ContainerProbeTypeGrpc))
		objAttr["port"] = types.Int64Value(probe.Grpc.Request.PortNumber)
		objAttr["grpc"] = types.ListValueMust(computeContainerAppContainerProbeGrpcType, []attr.Value{
			types.ObjectValueMust(computeContainerAppContainerProbeGrpcType.AttrTypes, map[string]attr.Value{
				"service": types.StringValue(probe.Grpc.Request.ServiceName),
			}),
		})
	}

	obj, diags := types.ObjectValue(computeContainerAppContainerProbeType.AttrTypes, objAttr)

	if diags.HasError() {
		return types.ListNull(computeContainerAppContainerProbeType), diags
	}

	return types.ListValue(computeContainerAppContainerProbeType, []attr.Value{obj})
}
