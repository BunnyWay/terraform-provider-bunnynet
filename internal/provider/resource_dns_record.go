// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"strconv"
	"strings"
)

var _ resource.Resource = &DnsRecordResource{}
var _ resource.ResourceWithImportState = &DnsRecordResource{}

func NewDnsRecordResourceResource() resource.Resource {
	return &DnsRecordResource{}
}

type DnsRecordResource struct {
	client *api.Client
}

type DnsRecordResourceModel struct {
	Id                    types.Int64   `tfsdk:"id"`
	Zone                  types.Int64   `tfsdk:"zone"`
	Enabled               types.Bool    `tfsdk:"enabled"`
	Type                  types.String  `tfsdk:"type"`
	Ttl                   types.Int64   `tfsdk:"ttl"`
	Value                 types.String  `tfsdk:"value"`
	Name                  types.String  `tfsdk:"name"`
	Weight                types.Int64   `tfsdk:"weight"`
	Priority              types.Int64   `tfsdk:"priority"`
	Port                  types.Int64   `tfsdk:"port"`
	Flags                 types.Int64   `tfsdk:"flags"`
	Tag                   types.String  `tfsdk:"tag"`
	Accelerated           types.Bool    `tfsdk:"accelerated"`
	AcceleratedPullZoneId types.Int64   `tfsdk:"accelerated_pullzone"`
	LinkName              types.String  `tfsdk:"link_name"`
	MonitorType           types.String  `tfsdk:"monitor_type"`
	GeolocationLatitude   types.Float64 `tfsdk:"geolocation_lat"`
	GeolocationLongitude  types.Float64 `tfsdk:"geolocation_long"`
	LatencyZone           types.String  `tfsdk:"latency_zone"`
	SmartRoutingType      types.String  `tfsdk:"smart_routing_type"`
	Comment               types.String  `tfsdk:"comment"`
}

var dnsRecordTypeMap = map[uint8]string{
	0:  "A",
	1:  "AAAA",
	2:  "CNAME",
	3:  "TXT",
	4:  "MX",
	5:  "Redirect",
	6:  "Flatten",
	7:  "PullZone",
	8:  "SRV",
	9:  "CAA",
	10: "PTR",
	11: "Script",
	12: "NS",
}

var dnsRecordMonitorTypeMap = map[uint8]string{
	0: "None",
	1: "Ping",
	2: "Http",
	3: "Monitor",
}

var dnsRecordSmartRoutingTypeMap = map[uint8]string{
	0: "None",
	1: "Latency",
	2: "Geolocation",
}

func (r *DnsRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DnsRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages DNS records in a bunny.net DNS zone. It is used to create, update, and delete DNS records within a specific DNS zone managed by bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the DNS record.",
			},
			"zone": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "ID of the related DNS zone.",
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the DNS record is enabled.",
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(dnsRecordTypeMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(dnsRecordTypeMap),
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(300),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The time-to-live value for the DNS record.",
			},
			"value": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The value of the DNS record.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the DNS record.",
			},
			"weight": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The weight of the DNS record. It is used in load balancing scenarios to distribute traffic based on the specified weight.",
			},
			"priority": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "The priority of the DNS record.",
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
				Description: "The port number for services that require a specific port.",
			},
			"flags": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				Description: "Flags for advanced DNS settings.",
			},
			"tag": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "A tag for the DNS record.",
			},
			"accelerated": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "Indicates whether the DNS record should utilize bunny.net’s acceleration services.",
			},
			"accelerated_pullzone": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the accelerated pull zone.",
			},
			"link_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The name of the linked resource.",
			},
			"monitor_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("None"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(dnsRecordMonitorTypeMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(dnsRecordMonitorTypeMap),
			},
			"geolocation_lat": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(0.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Float64{
					float64validator.Between(-180.0, 180.0),
				},
				Description: "The latitude for geolocation-based routing.",
			},
			"geolocation_long": schema.Float64Attribute{
				Optional: true,
				Computed: true,
				Default:  float64default.StaticFloat64(0.0),
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Float64{
					float64validator.Between(-180.0, 180.0),
				},
				Description: "The longitude for geolocation-based routing.",
			},
			"latency_zone": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The latency zone for latency-based routing.",
			},
			"smart_routing_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("None"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(dnsRecordSmartRoutingTypeMap)...),
				},
				MarkdownDescription: generateMarkdownMapOptions(dnsRecordSmartRoutingTypeMap),
			},
			"comment": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "This property allows users to add descriptive notes for documentation and management purposes.",
			},
		},
	}
}

func (r *DnsRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf DnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateDnsRecord(dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DNS record", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created dns record %s %s", mapKeyToValue(dnsRecordTypeMap, dataApi.Type), dataApi.Name))
	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DnsRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetDnsRecord(data.Zone.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching dns record", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateDnsRecord(dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating dns record", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDnsRecord(data.Zone.ValueInt64(), data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting dns record", err.Error()))
	}
}

func (r *DnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	zoneIdStr, idStr, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding dns record", "Use \"<zoneId>|<recordId>\" as ID on terraform import command"))
		return
	}

	zoneId, err := strconv.ParseInt(zoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding dns record", "Invalid DNS zone ID: "+err.Error()))
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error converting ID to integer", err.Error()))
		return
	}

	dataApi, err := r.client.GetDnsRecord(zoneId, id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching dns record", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsRecordResource) convertModelToApi(ctx context.Context, dataTf DnsRecordResourceModel) api.DnsRecord {
	dataApi := api.DnsRecord{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.Zone = dataTf.Zone.ValueInt64()
	dataApi.Type = mapValueToKey(dnsRecordTypeMap, dataTf.Type.ValueString())
	dataApi.Ttl = dataTf.Ttl.ValueInt64()
	dataApi.Value = dataTf.Value.ValueString()
	dataApi.Name = dataTf.Name.ValueString()
	dataApi.Weight = dataTf.Weight.ValueInt64()
	dataApi.Priority = dataTf.Priority.ValueInt64()
	dataApi.Port = dataTf.Port.ValueInt64()
	dataApi.Flags = dataTf.Flags.ValueInt64()
	dataApi.Tag = dataTf.Tag.ValueString()
	dataApi.Accelerated = dataTf.Accelerated.ValueBool()
	dataApi.AcceleratedPullZoneId = dataTf.AcceleratedPullZoneId.ValueInt64()
	dataApi.LinkName = dataTf.LinkName.ValueString()
	dataApi.MonitorType = mapValueToKey(dnsRecordMonitorTypeMap, dataTf.MonitorType.ValueString())
	dataApi.GeolocationLatitude = dataTf.GeolocationLatitude.ValueFloat64()
	dataApi.GeolocationLongitude = dataTf.GeolocationLongitude.ValueFloat64()
	dataApi.LatencyZone = dataTf.LatencyZone.ValueString()
	dataApi.SmartRoutingType = mapValueToKey(dnsRecordSmartRoutingTypeMap, dataTf.SmartRoutingType.ValueString())
	dataApi.Comment = dataTf.Comment.ValueString()
	dataApi.Disabled = !dataTf.Enabled.ValueBool()

	return dataApi
}

func (r *DnsRecordResource) convertApiToModel(dataApi api.DnsRecord) (DnsRecordResourceModel, diag.Diagnostics) {
	dataTf := DnsRecordResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Zone = types.Int64Value(dataApi.Zone)
	dataTf.Type = types.StringValue(mapKeyToValue(dnsRecordTypeMap, dataApi.Type))
	dataTf.Ttl = types.Int64Value(dataApi.Ttl)
	dataTf.Value = types.StringValue(dataApi.Value)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.Weight = types.Int64Value(dataApi.Weight)
	dataTf.Priority = types.Int64Value(dataApi.Priority)
	dataTf.Port = types.Int64Value(dataApi.Port)
	dataTf.Flags = types.Int64Value(dataApi.Flags)
	dataTf.Tag = types.StringValue(dataApi.Tag)
	dataTf.Accelerated = types.BoolValue(dataApi.Accelerated)
	dataTf.AcceleratedPullZoneId = types.Int64Value(dataApi.AcceleratedPullZoneId)
	dataTf.LinkName = types.StringValue(dataApi.LinkName)
	dataTf.MonitorType = types.StringValue(mapKeyToValue(dnsRecordMonitorTypeMap, dataApi.MonitorType))
	dataTf.GeolocationLatitude = types.Float64Value(dataApi.GeolocationLatitude)
	dataTf.GeolocationLongitude = types.Float64Value(dataApi.GeolocationLongitude)
	dataTf.LatencyZone = types.StringValue(dataApi.LatencyZone)
	dataTf.SmartRoutingType = types.StringValue(mapKeyToValue(dnsRecordSmartRoutingTypeMap, dataApi.SmartRoutingType))
	dataTf.Comment = types.StringValue(dataApi.Comment)
	dataTf.Enabled = types.BoolValue(!dataApi.Disabled)

	return dataTf, nil
}
