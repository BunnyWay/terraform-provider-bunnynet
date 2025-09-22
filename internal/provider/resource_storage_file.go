// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/storageplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"io"
	"os"
	"strconv"
	"strings"
)

var _ resource.Resource = &StorageFileResource{}
var _ resource.ResourceWithImportState = &StorageFileResource{}

func NewStorageFileResource() resource.Resource {
	return &StorageFileResource{}
}

type StorageFileResource struct {
	client *api.Client
}

type StorageFileResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Zone         types.Int64  `tfsdk:"zone"`
	Path         types.String `tfsdk:"path"`
	Content      types.String `tfsdk:"content"`
	Source       types.String `tfsdk:"source"`
	Size         types.Int64  `tfsdk:"size"`
	ContentType  types.String `tfsdk:"content_type"`
	DateCreated  types.String `tfsdk:"date_created"`
	DateModified types.String `tfsdk:"date_modified"`
	Checksum     types.String `tfsdk:"checksum"`
}

func (r *StorageFileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_file"
}

func (r *StorageFileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages files in a bunny.net storage zone. It is used to upload, update, and delete files within a storage zone.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the file.",
			},
			"zone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The ID of the storage zone where the file is stored.",
			},
			"path": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The path of the file within the storage zone.",
			},
			"content": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The to be stored in the file. Use <code>source</code> to upload files from the local disk.",
			},
			"source": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "The path in the local disk for the file to be uploaded to the storage zone. Use <code>content</code> to define the content directly.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "The size of the file in bytes.",
			},
			"content_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Specifies the content type of the file.",
			},
			"date_created": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The date and time when the file was created.",
			},
			"date_modified": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the file was last modified.",
			},
			"checksum": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					storageplanmodifier.DetectFileContentsChange(),
				},
				Description: "The SHA-256 hash of the stored file.",
			},
		},
	}
}

func (r *StorageFileResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("content"),
			path.MatchRoot("source"),
		),
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("content"),
			path.MatchRoot("source"),
		),
	}
}

func (r *StorageFileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageFileResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var checksumState string
	var checksumPlan string

	req.State.GetAttribute(ctx, path.Root("checksum"), &checksumState)
	req.Plan.GetAttribute(ctx, path.Root("checksum"), &checksumPlan)

	if checksumState == "" || checksumPlan == "" || checksumState == checksumPlan {
		return
	}

	resp.Plan.SetAttribute(ctx, path.Root("date_modified"), types.StringUnknown())
	resp.Plan.SetAttribute(ctx, path.Root("size"), types.Int64Unknown())
}

func (r *StorageFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf StorageFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.convertModelToApi(ctx, dataTf)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create storage file", err.Error())
		return
	}

	dataApi, err = r.client.CreateStorageFile(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create storage file", err.Error())
		return
	}

	tflog.Trace(ctx, "created storage file "+dataApi.Path)
	dataTfResult, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	if !dataTf.Content.IsNull() {
		dataTfResult.Content = dataTf.Content
	}

	if !dataTf.Source.IsNull() {
		dataTfResult.Source = dataTf.Source
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTfResult)...)
}

func (r *StorageFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageFileResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetStorageFile(ctx, data.Zone.ValueInt64(), data.Path.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching storage file", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	var content string
	req.State.GetAttribute(ctx, path.Root("content"), &content)
	if len(content) > 0 {
		dataTf.Content = types.StringValue(content)
	}

	var source string
	req.State.GetAttribute(ctx, path.Root("source"), &source)
	if len(source) > 0 {
		dataTf.Source = types.StringValue(source)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StorageFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StorageFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.convertModelToApi(ctx, data)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating storage file", err.Error()))
		return
	}

	dataApi, err = r.client.UpdateStorageFile(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating storage file", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	var content string
	req.Plan.GetAttribute(ctx, path.Root("content"), &content)
	if len(content) > 0 {
		dataTf.Content = types.StringValue(content)
	}

	var source string
	req.Plan.GetAttribute(ctx, path.Root("source"), &source)
	if len(source) > 0 {
		dataTf.Source = types.StringValue(source)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StorageFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StorageFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStorageFile(ctx, data.Zone.ValueInt64(), data.Path.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting storage file", err.Error()))
	}
}

func (r *StorageFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	zoneIdStr, filePath, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding storage file", "Use \"<storageZoneId>|<storageFilePath>\" as ID on terraform import command"))
		return
	}

	zoneId, err := strconv.ParseInt(zoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding storage file", "Invalid storage zone ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetStorageFile(ctx, zoneId, filePath)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching storage file", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StorageFileResource) convertModelToApi(ctx context.Context, dataTf StorageFileResourceModel) (api.StorageFile, error) {
	dataApi := api.StorageFile{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.Zone = dataTf.Zone.ValueInt64()
	dataApi.Path = dataTf.Path.ValueString()
	dataApi.Length = uint64(dataTf.Size.ValueInt64())
	dataApi.ContentType = dataTf.ContentType.ValueString()
	dataApi.DateCreated = dataTf.DateCreated.ValueString()
	dataApi.LastChanged = dataTf.DateModified.ValueString()
	dataApi.Checksum = dataTf.Checksum.ValueString()

	// content or source
	var body io.Reader
	if !dataTf.Content.IsNull() {
		body = strings.NewReader(dataTf.Content.ValueString())
	}

	if !dataTf.Source.IsNull() {
		var err error
		body, err = os.Open(dataTf.Source.ValueString())
		if err != nil {
			return api.StorageFile{}, err
		}
	}

	dataApi.FileContents = body

	return dataApi, nil
}

func (r *StorageFileResource) convertApiToModel(dataApi api.StorageFile) (StorageFileResourceModel, diag.Diagnostics) {
	dataTf := StorageFileResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.Zone = types.Int64Value(dataApi.Zone)
	dataTf.Path = types.StringValue(dataApi.Path)
	dataTf.Size = types.Int64Value(int64(dataApi.Length))
	dataTf.ContentType = types.StringValue(dataApi.ContentType)
	dataTf.DateCreated = types.StringValue(dataApi.DateCreated)
	dataTf.DateModified = types.StringValue(dataApi.LastChanged)
	dataTf.Checksum = types.StringValue(dataApi.Checksum)

	return dataTf, nil
}
