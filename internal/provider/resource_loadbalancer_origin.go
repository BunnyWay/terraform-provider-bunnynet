// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/customtype"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/loadbalancerresourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"strconv"
	"strings"
)

var _ resource.Resource = &LoadbalancerOriginResource{}
var _ resource.ResourceWithConfigure = &LoadbalancerOriginResource{}
var _ resource.ResourceWithImportState = &LoadbalancerOriginResource{}

func NewLoadbalancerOriginResource() resource.Resource {
	return &LoadbalancerOriginResource{}
}

type LoadbalancerOriginResource struct {
	client *api.Client
}

type LoadbalancerOriginResourceModel struct {
	Id                  types.Int64                       `tfsdk:"id"`
	LoadbalancerId      types.Int64                       `tfsdk:"loadbalancer"`
	GroupId             types.Int64                       `tfsdk:"group"`
	GroupWeight         types.Int64                       `tfsdk:"group_weight"`
	Enabled             types.Bool                        `tfsdk:"enabled"`
	Name                types.String                      `tfsdk:"name"`
	Type                types.String                      `tfsdk:"type"`
	Url                 customtype.PullzoneOriginUrlValue `tfsdk:"url"`
	HostHeader          types.String                      `tfsdk:"host_header"`
	VerifySsl           types.Bool                        `tfsdk:"verify_ssl"`
	StoragezoneId       types.Int64                       `tfsdk:"storagezone"`
	ScriptId            types.Int64                       `tfsdk:"script"`
	ContainerAppId      types.String                      `tfsdk:"container_app_id"`
	ContainerEndpointId types.String                      `tfsdk:"container_endpoint_id"`
}

func (r *LoadbalancerOriginResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_loadbalancer_origin"
}

func (r *LoadbalancerOriginResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a Load Balancer origin in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique ID of the Load Balancer origin.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"loadbalancer": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Load Balancer.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"group": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Load Balancer group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"group_weight": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(100),
				Description: `The weight of the Load Balancer origin. Used in Load Balancers groups with method = "Weight"`,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 255),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Indicates whether the Load Balancer origin is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Load Balancer origin.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: generateMarkdownMapOptions(loadbalancerOriginTypeMap),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(loadbalancerOriginTypeMap)...),
				},
			},
			"url": schema.StringAttribute{
				CustomType:  customtype.PullzoneOriginUrlType{},
				Optional:    true,
				Description: "The origin URL from where the files are fetched.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 500),
				},
			},
			"host_header": schema.StringAttribute{
				Optional:    true,
				Description: "The host header that will be sent to the origin.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"verify_ssl": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Indicates whether the Origin's TLS certificate should be verified.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"storagezone": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the linked storage zone.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"script": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the linked compute script.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"container_app_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID if the compute container app.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"container_endpoint_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID if the compute container app endpoint.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (r *LoadbalancerOriginResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		loadbalancerresourcevalidator.Origin(),
	}
}

func (r *LoadbalancerOriginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LoadbalancerOriginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf LoadbalancerOriginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateLoadbalancerOrigin(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.AddError("Unable to create loadbalancer origin", err.Error())
		return
	}

	tflog.Trace(ctx, "created loadbalancer "+dataApi.Name)
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerOriginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LoadbalancerOriginResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetLoadbalancerOrigin(ctx, data.LoadbalancerId.ValueInt64(), data.GroupId.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching loadbalancer origin", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerOriginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LoadbalancerOriginResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateLoadbalancerOrigin(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating loadbalancer origin", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *LoadbalancerOriginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LoadbalancerOriginResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteLoadbalancerOrigin(ctx, data.LoadbalancerId.ValueInt64(), data.GroupId.ValueInt64(), data.Id.ValueInt64())
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting loadbalancer origin", err.Error()))
	}
}

func (r *LoadbalancerOriginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "|", 3)
	if len(parts) != 3 {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Use \"<loadbalancerId>|<groupName>|<originName>\" as ID on terraform import command"))
		return
	}

	loadbalancerId, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Invalid loadbalancerId: "+err.Error()))
		return
	}

	groupName := parts[1]
	originName := parts[2]

	loadbalancer, err := r.client.GetLoadbalancer(ctx, loadbalancerId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching loadbalancer", err.Error()))
		return
	}

	for _, group := range loadbalancer.Groups {
		if group.Name != groupName {
			continue
		}

		for _, origin := range group.Origins {
			if origin.Name != originName {
				continue
			}

			origin.LoadbalancerId = loadbalancer.Id
			origin.GroupId = group.Id
			dataTf, diags := r.convertApiToModel(origin)
			if diags != nil {
				resp.Diagnostics.Append(diags...)
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("loadbalancer group does have origin", groupName))
		return
	}

	resp.Diagnostics.Append(diag.NewErrorDiagnostic("loadbalancer does have group", groupName))
}

func (r *LoadbalancerOriginResource) convertModelToApi(ctx context.Context, dataTf LoadbalancerOriginResourceModel) api.LoadbalancerOrigin {
	dataApi := api.LoadbalancerOrigin{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.LoadbalancerId = dataTf.LoadbalancerId.ValueInt64()
	dataApi.GroupId = dataTf.GroupId.ValueInt64()
	dataApi.Weight = uint8(dataTf.GroupWeight.ValueInt64())

	dataApi.Enabled = dataTf.Enabled.ValueBool()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.OriginType = mapValueToKey(loadbalancerOriginTypeMap, dataTf.Type.ValueString())

	switch dataApi.OriginType {
	case api.LoadbalancerOriginTypeOriginUrl:
		dataApi.OriginUrl = &api.LoadbalancerOriginOriginUrl{
			OriginUrl:  dataTf.Url.ValueString(),
			VerifySsl:  dataTf.VerifySsl.ValueBool(),
			HostHeader: dataTf.HostHeader.ValueString(),
		}

	case api.LoadbalancerOriginTypeStorageZone:
		storagezoneId := dataTf.StoragezoneId.ValueInt64()
		dataApi.StorageZoneId = &storagezoneId

	case api.LoadbalancerOriginTypeComputeScript:
		edgescriptId := dataTf.ScriptId.ValueInt64()
		dataApi.EdgeScriptId = &edgescriptId

	case api.LoadbalancerOriginTypeComputeContainer:
		dataApi.MagicContainersOrigin = &api.LoadbalancerOriginOriginMC{
			AppId:      dataTf.ContainerAppId.ValueString(),
			EndpointId: dataTf.ContainerEndpointId.ValueString(),
		}
	}

	return dataApi
}

func (r *LoadbalancerOriginResource) convertApiToModel(dataApi api.LoadbalancerOrigin) (LoadbalancerOriginResourceModel, diag.Diagnostics) {
	dataTf := LoadbalancerOriginResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.LoadbalancerId = types.Int64Value(dataApi.LoadbalancerId)
	dataTf.GroupId = types.Int64Value(dataApi.GroupId)
	dataTf.GroupWeight = types.Int64Value(int64(dataApi.Weight))

	dataTf.Enabled = types.BoolValue(dataApi.Enabled)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Type = types.StringValue(mapKeyToValue(loadbalancerOriginTypeMap, dataApi.OriginType))
	dataTf.VerifySsl = types.BoolValue(true)

	switch dataApi.OriginType {
	case api.LoadbalancerOriginTypeOriginUrl:
		dataTf.Url = customtype.PullzoneOriginUrlValue{StringValue: types.StringValue(dataApi.OriginUrl.OriginUrl)}
		dataTf.VerifySsl = types.BoolValue(dataApi.OriginUrl.VerifySsl)
		if dataApi.OriginUrl.HostHeader != "" {
			dataTf.HostHeader = types.StringValue(dataApi.OriginUrl.HostHeader)
		}

	case api.LoadbalancerOriginTypeStorageZone:
		dataTf.StoragezoneId = types.Int64Value(*dataApi.StorageZoneId)

	case api.LoadbalancerOriginTypeComputeScript:
		dataTf.ScriptId = types.Int64Value(*dataApi.EdgeScriptId)

	case api.LoadbalancerOriginTypeComputeContainer:
		dataTf.ContainerAppId = types.StringValue(dataApi.MagicContainersOrigin.AppId)
		dataTf.ContainerEndpointId = types.StringValue(dataApi.MagicContainersOrigin.EndpointId)
	}

	return dataTf, nil
}
