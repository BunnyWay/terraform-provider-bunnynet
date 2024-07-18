// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"strings"
)

var _ resource.Resource = &StreamCollectionResource{}
var _ resource.ResourceWithImportState = &StreamCollectionResource{}

func NewStreamCollectionResource() resource.Resource {
	return &StreamCollectionResource{}
}

type StreamCollectionResource struct {
	client *api.Client
}

type StreamCollectionResourceModel struct {
	Id      types.String `tfsdk:"id"`
	Library types.Int64  `tfsdk:"library"`
	Name    types.String `tfsdk:"name"`
}

func (r *StreamCollectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream_collection"
}

func (r *StreamCollectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages collections in bunny.net Stream. It is used to create and organize collections of video content.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the stream collection.",
			},
			"library": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: "The ID of the stream library to which the collection belongs.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the stream collection.",
			},
		},
	}
}

func (r *StreamCollectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StreamCollectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf StreamCollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateStreamCollection(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create stream collection", err.Error())
		return
	}

	tflog.Trace(ctx, "created stream collection "+dataApi.Name)
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamCollectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StreamCollectionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetStreamCollection(data.Library.ValueInt64(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching stream collection", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamCollectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StreamCollectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateStreamCollection(dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating stream collection", err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated stream collection %s for library %d", dataApi.Id, dataApi.LibraryId))

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamCollectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StreamCollectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStreamCollection(data.Library.ValueInt64(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting stream collection", err.Error()))
	}
}

func (r *StreamCollectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	libraryIdStr, guid, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding stream collection", "Use \"<streamLibraryId>|<streamCollectionGuid>\" as ID on terraform import command"))
		return
	}

	libraryId, err := strconv.ParseInt(libraryIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding stream collection", "Invalid stream library ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetStreamCollection(libraryId, guid)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching stream collection", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamCollectionResource) convertModelToApi(ctx context.Context, dataTf StreamCollectionResourceModel) api.StreamCollection {
	dataApi := api.StreamCollection{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.LibraryId = dataTf.Library.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()

	return dataApi
}

func (r *StreamCollectionResource) convertApiToModel(dataApi api.StreamCollection) (StreamCollectionResourceModel, diag.Diagnostics) {
	dataTf := StreamCollectionResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.Library = types.Int64Value(dataApi.LibraryId)
	dataTf.Name = types.StringValue(dataApi.Name)

	return dataTf, nil
}
