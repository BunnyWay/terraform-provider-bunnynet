package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		MarkdownDescription: "Storage File",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"content": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
			"content_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"date_created": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"date_modified": schema.StringAttribute{
				Computed: true,
			},
			"checksum": schema.StringAttribute{
				Computed: true,
			},
		},
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

func (r *StorageFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf StorageFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateStorageFile(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create storage file", err.Error())
		return
	}

	tflog.Trace(ctx, "created storage file "+dataApi.Path)
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StorageFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageFileResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetStorageFile(data.Zone.ValueInt64(), data.Path.ValueString())
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

func (r *StorageFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StorageFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateStorageFile(dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating storage file", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *StorageFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StorageFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteStorageFile(data.Zone.ValueInt64(), data.Path.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting storage file", err.Error()))
	}
}

func (r *StorageFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	zoneIdStr, path, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding storage file", "Use \"<storageZoneId>|<storageFilePath>\" as ID on terraform import command"))
		return
	}

	zoneId, err := strconv.ParseInt(zoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding storage file", "Invalid storage zone ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetStorageFile(zoneId, path)
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

func (r *StorageFileResource) convertModelToApi(ctx context.Context, dataTf StorageFileResourceModel) api.StorageFile {
	dataApi := api.StorageFile{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.Zone = dataTf.Zone.ValueInt64()
	dataApi.Path = dataTf.Path.ValueString()
	dataApi.Content = dataTf.Content.ValueString()
	dataApi.Length = uint64(dataTf.Size.ValueInt64())
	dataApi.ContentType = dataTf.ContentType.ValueString()
	dataApi.DateCreated = dataTf.DateCreated.ValueString()
	dataApi.LastChanged = dataTf.DateModified.ValueString()
	dataApi.Checksum = dataTf.Checksum.ValueString()

	return dataApi
}

func (r *StorageFileResource) convertApiToModel(dataApi api.StorageFile) (StorageFileResourceModel, diag.Diagnostics) {
	dataTf := StorageFileResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.Zone = types.Int64Value(dataApi.Zone)
	dataTf.Path = types.StringValue(dataApi.Path)
	dataTf.Content = types.StringValue(dataApi.Content)
	dataTf.Size = types.Int64Value(int64(dataApi.Length))
	dataTf.ContentType = types.StringValue(dataApi.ContentType)
	dataTf.DateCreated = types.StringValue(dataApi.DateCreated)
	dataTf.DateModified = types.StringValue(dataApi.LastChanged)
	dataTf.Checksum = types.StringValue(dataApi.Checksum)

	return dataTf, nil
}