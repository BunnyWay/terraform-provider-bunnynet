// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"strings"
)

var _ resource.Resource = &PullzoneOptimizerClassResource{}
var _ resource.ResourceWithImportState = &PullzoneOptimizerClassResource{}

func NewPullzoneOptimizerClassResource() resource.Resource {
	return &PullzoneOptimizerClassResource{}
}

type PullzoneOptimizerClassResource struct {
	client *api.Client
}

type PullzoneOptimizerClassResourceModel struct {
	PullzoneId   types.Int64  `tfsdk:"pullzone"`
	Name         types.String `tfsdk:"name"`
	Width        types.Int64  `tfsdk:"width"`
	Height       types.Int64  `tfsdk:"height"`
	AspectRatio  types.String `tfsdk:"aspect_ratio"`
	Quality      types.Int64  `tfsdk:"quality"`
	Sharpen      types.Bool   `tfsdk:"sharpen"`
	Blur         types.Int64  `tfsdk:"blur"`
	Crop         types.String `tfsdk:"crop"`
	CropGravity  types.String `tfsdk:"crop_gravity"`
	Flip         types.Bool   `tfsdk:"flip"`
	Flop         types.Bool   `tfsdk:"flop"`
	Brightness   types.Int64  `tfsdk:"brightness"`
	Saturation   types.Int64  `tfsdk:"saturation"`
	Hue          types.Int64  `tfsdk:"hue"`
	Contrast     types.Int64  `tfsdk:"contrast"`
	AutoOptimize types.String `tfsdk:"auto_optimize"`
	Sepia        types.Int64  `tfsdk:"sepia"`
}

func (r *PullzoneOptimizerClassResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_optimizer_class"
}

func (r *PullzoneOptimizerClassResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages optimizer classes for a bunny.net pull zone. It is used to define Bunny Optimizer settings for pullzones.",

		Attributes: map[string]schema.Attribute{
			"pullzone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The ID of the linked pull zone.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The name of the optimizer class.",
			},
			"width": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The width to which images should be resized.",
			},
			"height": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The height to which images should be resized.",
			},
			"aspect_ratio": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The aspect ratio for image optimization.",
			},
			"quality": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
				Description: "The quality setting for image optimization.",
			},
			"sharpen": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether to sharpen images.",
			},
			"blur": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
				Description: "The level of blur to apply to images.",
			},
			"crop": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The cropping settings for images.",
			},
			"crop_gravity": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("center", "forget", "east", "north", "south", "west", "northeast", "northwest", "southeast", "southwest"),
				},
				Description: "The gravity setting for cropping.",
			},
			"flip": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether to flip images horizontally.",
			},
			"flop": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether to flip images vertically.",
			},
			"brightness": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(-100, 100),
				},
				Description: "The brightness adjustment for images.",
			},
			"saturation": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(-100, 100),
				},
				Description: "The saturation adjustment for images.",
			},
			"hue": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
				Description: "The hue adjustment for images.",
			},
			"contrast": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(-100, 100),
				},
				Description: "The contrast adjustment for images.",
			},
			"auto_optimize": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("low", "medium", "high"),
				},
				Description: "Indicates whether automatic optimization for images is enabled.",
			},
			"sepia": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 100),
				},
				Description: "The level of sepia tone to apply to images.",
			},
		},
	}
}

func (r *PullzoneOptimizerClassResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneOptimizerClassResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneOptimizerClassResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := dataTf.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneOptimizerClass(dataApi)
	pzMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.AddError("Unable to create Optimizer Image Class", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created Optimizer Image Class for pullzone %d", dataApi.PullzoneId))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneOptimizerClassResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneOptimizerClassResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	dataApi, err := r.client.GetPullzoneOptimizerClass(data.PullzoneId.ValueInt64(), data.Name.ValueString())
	pzMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching Optimizer Image Class", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneOptimizerClassResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneOptimizerClassResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdatePullzoneOptimizerClass(dataApi)
	pzMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating Optimizer Image Class", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneOptimizerClassResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneOptimizerClassResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pullzoneId := data.PullzoneId.ValueInt64()
	pzMutex.Lock(pullzoneId)
	err := r.client.DeletePullzoneOptimizerClass(pullzoneId, data.Name.ValueString())
	pzMutex.Unlock(pullzoneId)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting Optimizer Image Class", err.Error()))
	}
}

func (r *PullzoneOptimizerClassResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pullzoneIdStr, name, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding Optimizer Image Class", "Use \"<pullzoneId>|<optimizerClassName>\" as ID on terraform import command"))
		return
	}

	pullzoneId, err := strconv.ParseInt(pullzoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding Optimizer Image Class", "Invalid pullzone ID: "+err.Error()))
		return
	}

	optimizer_class, err := r.client.GetPullzoneOptimizerClass(pullzoneId, name)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding Optimizer Image Class", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(optimizer_class)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneOptimizerClassResource) convertModelToApi(ctx context.Context, dataTf PullzoneOptimizerClassResourceModel) api.PullzoneOptimizerClass {
	properties := map[string]string{}

	if !dataTf.Width.IsNull() {
		properties["width"] = strconv.FormatInt(dataTf.Width.ValueInt64(), 10)
	}

	if !dataTf.Height.IsNull() {
		properties["height"] = strconv.FormatInt(dataTf.Height.ValueInt64(), 10)
	}

	if !dataTf.AspectRatio.IsNull() {
		properties["aspect_ratio"] = dataTf.AspectRatio.ValueString()
	}

	if !dataTf.Quality.IsNull() {
		properties["quality"] = strconv.FormatInt(dataTf.Quality.ValueInt64(), 10)
	}

	if !dataTf.Sharpen.IsNull() {
		v := dataTf.Sharpen.ValueBool()
		if v {
			properties["sharpen"] = "true"
		} else {
			properties["sharpen"] = "false"
		}
	}

	if !dataTf.Blur.IsNull() {
		properties["blur"] = strconv.FormatInt(dataTf.Blur.ValueInt64(), 10)
	}

	if !dataTf.Crop.IsNull() {
		properties["crop"] = dataTf.Crop.ValueString()
	}

	if !dataTf.CropGravity.IsNull() {
		properties["crop_gravity"] = dataTf.CropGravity.ValueString()
	}

	if !dataTf.Flip.IsNull() {
		v := dataTf.Flip.ValueBool()
		if v {
			properties["flip"] = "true"
		} else {
			properties["flip"] = "false"
		}
	}

	if !dataTf.Flop.IsNull() {
		v := dataTf.Flop.ValueBool()
		if v {
			properties["flop"] = "true"
		} else {
			properties["flop"] = "false"
		}
	}

	if !dataTf.Brightness.IsNull() {
		properties["brightness"] = strconv.FormatInt(dataTf.Brightness.ValueInt64(), 10)
	}

	if !dataTf.Saturation.IsNull() {
		properties["saturation"] = strconv.FormatInt(dataTf.Saturation.ValueInt64(), 10)
	}

	if !dataTf.Hue.IsNull() {
		properties["hue"] = strconv.FormatInt(dataTf.Hue.ValueInt64(), 10)
	}

	if !dataTf.Contrast.IsNull() {
		properties["contrast"] = strconv.FormatInt(dataTf.Contrast.ValueInt64(), 10)
	}

	if !dataTf.AutoOptimize.IsNull() {
		properties["auto_optimize"] = dataTf.AutoOptimize.ValueString()
	}

	if !dataTf.Sepia.IsNull() {
		properties["sepia"] = strconv.FormatInt(dataTf.Sepia.ValueInt64(), 10)
	}

	dataApi := api.PullzoneOptimizerClass{}
	dataApi.PullzoneId = dataTf.PullzoneId.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Properties = properties

	return dataApi
}

func (r *PullzoneOptimizerClassResource) convertApiToModel(dataApi api.PullzoneOptimizerClass) (PullzoneOptimizerClassResourceModel, diag.Diagnostics) {
	dataTf := PullzoneOptimizerClassResourceModel{}
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Name = types.StringValue(dataApi.Name)

	if v, ok := dataApi.Properties["width"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid width value")}
		}

		dataTf.Width = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["height"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid height value")}
		}

		dataTf.Height = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["aspect_ratio"]; ok {
		dataTf.AspectRatio = types.StringValue(v)
	}

	if v, ok := dataApi.Properties["quality"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid quality value")}
		}

		dataTf.Quality = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["sharpen"]; ok {
		dataTf.Sharpen = types.BoolValue(v == "true")
	}

	if v, ok := dataApi.Properties["blur"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid blur value")}
		}

		dataTf.Blur = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["crop"]; ok {
		dataTf.Crop = types.StringValue(v)
	}

	if v, ok := dataApi.Properties["crop_gravity"]; ok {
		dataTf.CropGravity = types.StringValue(v)
	}

	if v, ok := dataApi.Properties["flip"]; ok {
		dataTf.Flip = types.BoolValue(v == "true")
	}

	if v, ok := dataApi.Properties["flop"]; ok {
		dataTf.Flop = types.BoolValue(v == "true")
	}

	if v, ok := dataApi.Properties["brightness"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid brightness value")}
		}

		dataTf.Brightness = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["saturation"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid saturation value")}
		}

		dataTf.Saturation = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["hue"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid hue value")}
		}

		dataTf.Hue = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["contrast"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid contrast value")}
		}

		dataTf.Contrast = types.Int64Value(int64(value))
	}

	if v, ok := dataApi.Properties["auto_optimize"]; ok {
		dataTf.CropGravity = types.StringValue(v)
	}

	if v, ok := dataApi.Properties["sepia"]; ok {
		value, err := strconv.Atoi(v)
		if err != nil {
			return PullzoneOptimizerClassResourceModel{}, diag.Diagnostics{diag.NewErrorDiagnostic("Error converting API data", "Invalid sepia value")}
		}

		dataTf.Sepia = types.Int64Value(int64(value))
	}

	return dataTf, nil
}
