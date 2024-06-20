package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunny/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"regexp"
	"strconv"
	"strings"
)

var _ resource.Resource = &PullzoneHostnameResource{}
var _ resource.ResourceWithImportState = &PullzoneHostnameResource{}

func NewPullzoneHostnameResource() resource.Resource {
	return &PullzoneHostnameResource{}
}

type PullzoneHostnameResource struct {
	client *api.Client
}

type PullzoneHostnameResourceModel struct {
	Id         types.Int64  `tfsdk:"id"`
	PullzoneId types.Int64  `tfsdk:"pullzone"`
	Name       types.String `tfsdk:"name"`
	ForceSSL   types.Bool   `tfsdk:"force_ssl"`
}

func (r *PullzoneHostnameResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_hostname"
}

func (r *PullzoneHostnameResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pullzone Hostname",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"pullzone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile("(.+)\\.(.+)"), "Invalid domain"),
				},
			},
			"force_ssl": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PullzoneHostnameResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneHostnameResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneHostnameResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneHostname(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create hostname", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created hostname for pullzone %d", dataApi.PullzoneId))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneHostnameResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneHostnameResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetPullzoneHostname(data.PullzoneId.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching hostname", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneHostnameResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneHostnameResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdatePullzoneHostname(dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating hostname", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneHostnameResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneHostnameResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePullzoneHostname(data.PullzoneId.ValueInt64(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting hostname", err.Error()))
	}
}

func (r *PullzoneHostnameResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pullzoneIdStr, idStr, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding hostname", "Use \"<pullzoneId>|<hostnameId>\" as ID on terraform import command"))
		return
	}

	pullzoneId, err := strconv.ParseInt(pullzoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding hostname", "Invalid pullzone ID: "+err.Error()))
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding hostname", "Invalid ID: "+err.Error()))
		return
	}

	hostname, err := r.client.GetPullzoneHostname(pullzoneId, id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding hostname", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(hostname)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneHostnameResource) convertModelToApi(ctx context.Context, dataTf PullzoneHostnameResourceModel) api.PullzoneHostname {
	dataApi := api.PullzoneHostname{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.PullzoneId = dataTf.PullzoneId.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.ForceSSL = dataTf.ForceSSL.ValueBool()

	return dataApi
}

func (r *PullzoneHostnameResource) convertApiToModel(dataApi api.PullzoneHostname) (PullzoneHostnameResourceModel, diag.Diagnostics) {
	dataTf := PullzoneHostnameResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.ForceSSL = types.BoolValue(dataApi.ForceSSL)

	return dataTf, nil
}
