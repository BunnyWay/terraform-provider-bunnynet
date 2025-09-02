// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"strings"
)

var _ resource.Resource = &StreamVideoResource{}
var _ resource.ResourceWithImportState = &StreamVideoResource{}

func NewStreamVideoResource() resource.Resource {
	return &StreamVideoResource{}
}

type StreamVideoResource struct {
	client *api.Client
}

type StreamVideoResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Library     types.Int64  `tfsdk:"library"`
	Collection  types.String `tfsdk:"collection"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Chapters    types.Set    `tfsdk:"chapters"`
	Moments     types.Set    `tfsdk:"moments"`
}

var streamVideoChapterType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"title": types.StringType,
		"start": types.StringType,
		"end":   types.StringType,
	},
}

var streamVideoMomentType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"label":     types.StringType,
		"timestamp": types.StringType,
	},
}

func (r *StreamVideoResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stream_video"
}

func (r *StreamVideoResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages individual video files in bunny.net Stream. It is used to manage individual video files in a stream library.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique ID of the video.",
			},
			"library": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The ID of the stream library to which the video belongs.",
			},
			"collection": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the collection to which the video belongs.",
			},
			"title": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The title of the video.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The description of the video.",
			},
			"chapters": schema.SetAttribute{
				ElementType: streamVideoChapterType,
				Optional:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "The list of chapters available in the video.",
			},
			"moments": schema.SetAttribute{
				ElementType: streamVideoMomentType,
				Optional:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "The list of moments available in the video.",
			},
		},
	}
}

func (r *StreamVideoResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StreamVideoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("Create video not supported", "Upload the video via dash.bunny.net and then use terraform import.")
}

func (r *StreamVideoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StreamVideoResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetStreamVideo(data.Library.ValueInt64(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching stream video", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamVideoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StreamVideoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.convertModelToApi(ctx, data)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating stream video", err.Error()))
		return
	}

	dataApi, err = r.client.UpdateStreamVideo(dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating stream video", err.Error()))
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated stream video %s for library %d", dataApi.Id, dataApi.LibraryId))

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	// handle not-present and empty sets as the same thing
	var chapters types.Set
	req.Plan.GetAttribute(ctx, path.Root("chapters"), &chapters)
	if !chapters.IsNull() && len(chapters.Elements()) == 0 {
		dataTf.Chapters = types.SetValueMust(streamVideoChapterType, []attr.Value{})
	}

	// handle not-present and empty sets as the same thing
	var moments types.Set
	req.Plan.GetAttribute(ctx, path.Root("moments"), &moments)
	if !moments.IsNull() && len(moments.Elements()) == 0 {
		dataTf.Moments = types.SetValueMust(streamVideoMomentType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamVideoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StreamVideoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStreamVideo(data.Library.ValueInt64(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting stream video", err.Error()))
	}
}

func (r *StreamVideoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	libraryIdStr, guid, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding stream video", "Use \"<streamLibraryId>|<streamVideoGuid>\" as ID on terraform import command"))
		return
	}

	libraryId, err := strconv.ParseInt(libraryIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding stream library", "Invalid stream library ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetStreamVideo(libraryId, guid)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching stream video", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StreamVideoResource) convertModelToApi(ctx context.Context, dataTf StreamVideoResourceModel) (api.StreamVideo, error) {
	dataApi := api.StreamVideo{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.LibraryId = dataTf.Library.ValueInt64()
	dataApi.CollectionId = dataTf.Collection.ValueString()
	dataApi.Title = dataTf.Title.ValueString()
	dataApi.MetaTags = []api.StreamVideoMetaTag{
		{Property: "description", Value: dataTf.Description.ValueString()},
	}

	// chapters
	{
		chapterElements := dataTf.Chapters.Elements()
		dataApi.Chapters = make([]api.StreamVideoChapter, len(chapterElements))
		for i, el := range chapterElements {
			var title string
			var start, end uint64
			var err error

			attrValues := el.(types.Object).Attributes()

			if t, ok := attrValues["title"]; ok && !t.(types.String).IsNull() {
				title = t.(types.String).ValueString()
			}

			if t, ok := attrValues["start"]; ok && !t.(types.String).IsNull() {
				start, err = convertTimestampToSeconds(t.(types.String).ValueString())
				if err != nil {
					return dataApi, err
				}
			}

			if t, ok := attrValues["end"]; ok && !t.(types.String).IsNull() {
				end, err = convertTimestampToSeconds(t.(types.String).ValueString())
				if err != nil {
					return dataApi, err
				}
			}

			dataApi.Chapters[i] = api.StreamVideoChapter{
				Title: title,
				Start: start,
				End:   end,
			}
		}
	}

	// moments
	{
		momentsElements := dataTf.Moments.Elements()
		dataApi.Moments = make([]api.StreamVideoMoment, len(momentsElements))
		for i, el := range momentsElements {
			var label string
			var timestamp uint64
			var err error

			attrValues := el.(types.Object).Attributes()

			if t, ok := attrValues["label"]; ok && !t.(types.String).IsNull() {
				label = t.(types.String).ValueString()
			}

			if t, ok := attrValues["timestamp"]; ok && !t.(types.String).IsNull() {
				timestamp, err = convertTimestampToSeconds(t.(types.String).ValueString())
				if err != nil {
					return dataApi, err
				}
			}

			dataApi.Moments[i] = api.StreamVideoMoment{
				Label:     label,
				Timestamp: timestamp,
			}
		}
	}

	return dataApi, nil
}

func (r *StreamVideoResource) convertApiToModel(dataApi api.StreamVideo) (StreamVideoResourceModel, diag.Diagnostics) {
	dataTf := StreamVideoResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.Library = types.Int64Value(dataApi.LibraryId)
	dataTf.Collection = types.StringValue(dataApi.CollectionId)
	dataTf.Title = types.StringValue(dataApi.Title)

	for _, meta := range dataApi.MetaTags {
		if meta.Property == "description" {
			dataTf.Description = types.StringValue(meta.Value)
		}
	}

	// chapters
	if len(dataApi.Chapters) == 0 {
		dataTf.Chapters = types.SetNull(streamVideoChapterType)
	} else {
		chaptersValues := make([]attr.Value, 0, len(dataApi.Chapters))
		for _, chapter := range dataApi.Chapters {
			chapterObj, diags := types.ObjectValue(streamVideoChapterType.AttrTypes, map[string]attr.Value{
				"title": types.StringValue(chapter.Title),
				"start": types.StringValue(convertSecondsToTimestamp(chapter.Start)),
				"end":   types.StringValue(convertSecondsToTimestamp(chapter.End)),
			})

			if diags != nil {
				return StreamVideoResourceModel{}, diags
			}

			chaptersValues = append(chaptersValues, chapterObj)
		}

		chaptersSet, diags := types.SetValue(streamVideoChapterType, chaptersValues)
		if diags != nil {
			return StreamVideoResourceModel{}, diags
		}

		dataTf.Chapters = chaptersSet
	}

	// moments
	if len(dataApi.Moments) == 0 {
		dataTf.Moments = types.SetNull(streamVideoMomentType)
	} else {
		momentsValues := make([]attr.Value, 0, len(dataApi.Moments))
		for _, moment := range dataApi.Moments {
			momentObj, diags := types.ObjectValue(streamVideoMomentType.AttrTypes, map[string]attr.Value{
				"label":     types.StringValue(moment.Label),
				"timestamp": types.StringValue(convertSecondsToTimestamp(moment.Timestamp)),
			})

			if diags != nil {
				return StreamVideoResourceModel{}, diags
			}

			momentsValues = append(momentsValues, momentObj)
		}

		momentsSet, diags := types.SetValue(streamVideoMomentType, momentsValues)
		if diags != nil {
			return StreamVideoResourceModel{}, diags
		}

		dataTf.Moments = momentsSet
	}

	return dataTf, nil
}
