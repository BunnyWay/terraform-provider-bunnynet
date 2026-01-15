// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"strconv"
	"strings"
)

var _ resource.Resource = &PullzoneAccessListResource{}
var _ resource.ResourceWithImportState = &PullzoneAccessListResource{}

func NewPullzoneAccessList() resource.Resource {
	return &PullzoneAccessListResource{}
}

type PullzoneAccessListResource struct {
	client *api.Client
}

type PullzoneAccessListResourceModel struct {
	Id       types.Int64  `tfsdk:"id"`
	Pullzone types.Int64  `tfsdk:"pullzone"`
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Type     types.String `tfsdk:"type"`
	Action   types.String `tfsdk:"action"`
	Entries  types.Set    `tfsdk:"entries"`
}

// curl -sS -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/shield-zone/13939/access-lists/enums | jq -r '.AccessListType'
var pullzoneAccessListTypeMap = map[uint8]string{
	0: "IP",
	1: "CIDR",
	2: "ASN",
	3: "Country",
}

// curl -sS -H "AccessKey: ${BUNNYNET_API_KEY}" https://api.bunny.net/shield/shield-zone/13939/access-lists/enums | jq -r '.AccessListAction'
var pullzoneAccessListActionMap = map[uint8]string{
	1: "Allow",
	2: "Block",
	3: "Challenge",
	4: "Log",
	5: "Bypass",
}

func (r *PullzoneAccessListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_access_list"
}

func (r *PullzoneAccessListResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages an Access List for a bunny.net pullzone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the Access List.",
			},
			"pullzone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				Description: "The ID of the linked pullzone.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The Access List name.",
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicated whether the Access List is enabled.",
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneAccessListTypeMap)...),
				},
				Description: generateMarkdownMapOptions(pullzoneAccessListTypeMap),
			},
			"action": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneAccessListActionMap)...),
				},
				Description: generateMarkdownMapOptions(pullzoneAccessListActionMap),
			},
			"entries": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Description: "The Access List entries.",
			},
		},
	}
}

func (r *PullzoneAccessListResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *PullzoneAccessListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PullzoneAccessListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf PullzoneAccessListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreatePullzoneAccessList(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create access list", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created access list for pullzone %d", dataTf.Pullzone.ValueInt64()))
	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneAccessListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PullzoneAccessListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetPullzoneAccessList(ctx, data.Pullzone.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching access list", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneAccessListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PullzoneAccessListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApiResult, err := r.client.UpdatePullzoneAccessList(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating access list", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApiResult)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneAccessListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PullzoneAccessListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePullzoneAccessList(ctx, data.Pullzone.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting access list", err.Error()))
	}
}

func (r *PullzoneAccessListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pullzoneIdStr, listIdStr, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Use \"<pullzoneId>|<listId>\" as ID on terraform import command"))
		return
	}

	pullzoneId, err := strconv.ParseInt(pullzoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Invalid pullzoneId: "+err.Error()))
		return
	}

	listId, err := strconv.ParseInt(listIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Invalid resource identifier", "Invalid listId: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetPullzoneAccessList(ctx, pullzoneId, listId)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching access list", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(ctx, dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *PullzoneAccessListResource) convertModelToApi(ctx context.Context, dataTf PullzoneAccessListResourceModel) api.PullzoneAccessList {
	dataApi := api.PullzoneAccessList{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.PullzoneId = dataTf.Pullzone.ValueInt64()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.IsEnabled = dataTf.Enabled.ValueBool()
	dataApi.Type = mapValueToKey(pullzoneAccessListTypeMap, dataTf.Type.ValueString())
	dataApi.Action = mapValueToKey(pullzoneAccessListActionMap, dataTf.Action.ValueString())

	// entries
	{
		entriesTf := dataTf.Entries.Elements()
		entries := make([]string, 0, len(entriesTf))
		for _, entry := range entriesTf {
			entries = append(entries, entry.(types.String).ValueString())
		}
		dataApi.Entries = entries
	}

	return dataApi
}

func (r *PullzoneAccessListResource) convertApiToModel(ctx context.Context, dataApi api.PullzoneAccessList) (PullzoneAccessListResourceModel, diag.Diagnostics) {
	dataTf := PullzoneAccessListResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Pullzone = types.Int64Value(dataApi.PullzoneId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Enabled = types.BoolValue(dataApi.IsEnabled)
	dataTf.Action = types.StringValue(mapKeyToValue(pullzoneAccessListActionMap, dataApi.Action))
	dataTf.Type = types.StringValue(mapKeyToValue(pullzoneAccessListTypeMap, dataApi.Type))

	// entries
	entries := make([]attr.Value, 0, len(dataApi.Entries))
	for _, entry := range dataApi.Entries {
		entries = append(entries, types.StringValue(entry))
	}

	entriesSet, diags := types.SetValue(types.StringType, entries)
	if diags.HasError() {
		return dataTf, diags
	}

	dataTf.Entries = entriesSet

	return dataTf, nil
}
