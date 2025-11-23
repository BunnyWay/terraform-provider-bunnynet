// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/dnszoneresourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"regexp"
	"strconv"
)

var _ resource.Resource = &DnsZoneResource{}
var _ resource.ResourceWithImportState = &DnsZoneResource{}
var _ resource.ResourceWithModifyPlan = &DnsZoneResource{}

func NewDnsZoneResourceResource() resource.Resource {
	return &DnsZoneResource{}
}

type DnsZoneResource struct {
	client *api.Client
}

type DnsZoneResourceModel struct {
	Id                 types.Int64  `tfsdk:"id"`
	Domain             types.String `tfsdk:"domain"`
	NameserverCustom   types.Bool   `tfsdk:"nameserver_custom"`
	Nameserver1        types.String `tfsdk:"nameserver1"`
	Nameserver2        types.String `tfsdk:"nameserver2"`
	SoaEmail           types.String `tfsdk:"soa_email"`
	LogEnabled         types.Bool   `tfsdk:"log_enabled"`
	LogAnonymized      types.Bool   `tfsdk:"log_anonymized"`
	LogAnonymizedStyle types.String `tfsdk:"log_anonymized_style"`
	DnssecEnabled      types.Bool   `tfsdk:"dnssec_enabled"`
	DnssecDigest       types.String `tfsdk:"dnssec_digest"`
	DnssecDigestType   types.Int64  `tfsdk:"dnssec_digest_type"`
	DnssecAlgorithm    types.Int64  `tfsdk:"dnssec_algorithm"`
	DnssecPublicKey    types.String `tfsdk:"dnssec_public_key"`
	DnssecFlags        types.Int64  `tfsdk:"dnssec_flags"`
	DnssecKeytag       types.Int64  `tfsdk:"dnssec_keytag"`
}

var dnsZoneLogAnonymizedStyleMap = map[uint8]string{
	0: "OneDigit",
	1: "Drop",
}

func (r *DnsZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (r *DnsZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a DNS zone in bunny.net. It is used to create and manage DNS zones.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: dnsZoneDescription.Id,
			},
			"domain": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`(.+)\.(.+)`), "Invalid domain"),
				},
				Description: dnsZoneDescription.Domain,
			},
			"nameserver_custom": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: dnsZoneDescription.NameserverCustom,
			},
			"nameserver1": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("kiki.bunny.net"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: dnsZoneDescription.Nameserver1,
			},
			"nameserver2": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("coco.bunny.net"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: dnsZoneDescription.Nameserver2,
			},
			"soa_email": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("hostmaster@bunny.net"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`(.+)@(.+)\.(.+)`), "Invalid email address"),
				},
				Description: dnsZoneDescription.SoaEmail,
			},
			"log_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: dnsZoneDescription.LogEnabled,
			},
			"log_anonymized": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: dnsZoneDescription.LogAnonymized,
			},
			"log_anonymized_style": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("OneDigit"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(maps.Values(pullzoneLogAnonymizedStyleMap)...),
				},
				MarkdownDescription: dnsZoneDescription.LogAnonymizedStyle,
			},
			"dnssec_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: dnsZoneDescription.DnssecEnabled,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_algorithm": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecAlgorithm,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecPublicKey,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_digest": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecDigest,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_digest_type": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecDigestType,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_flags": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecFlags,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_keytag": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecKeyTag,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DnsZoneResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		dnszoneresourcevalidator.CustomNameserver(),
	}
}

func (r *DnsZoneResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	enabledAttr := path.Root("dnssec_enabled")

	var enabledPlan types.Bool
	var enabledState types.Bool
	var diags diag.Diagnostics

	diags = request.Plan.GetAttribute(ctx, enabledAttr, &enabledPlan)
	if diags.HasError() {
		response.Diagnostics = append(response.Diagnostics, diags...)
		return
	}

	diags = request.State.GetAttribute(ctx, enabledAttr, &enabledState)
	if diags.HasError() {
		response.Diagnostics = append(response.Diagnostics, diags...)
		return
	}

	if enabledPlan.IsNull() || enabledState.IsNull() || enabledPlan.IsUnknown() || enabledState.IsUnknown() {
		return
	}

	if enabledPlan.ValueBool() == enabledState.ValueBool() {
		return
	}

	response.Plan.SetAttribute(ctx, path.Root("dnssec_algorithm"), types.Int64Unknown())
	response.Plan.SetAttribute(ctx, path.Root("dnssec_public_key"), types.StringUnknown())
	response.Plan.SetAttribute(ctx, path.Root("dnssec_digest"), types.StringUnknown())
	response.Plan.SetAttribute(ctx, path.Root("dnssec_digest_type"), types.Int64Unknown())
	response.Plan.SetAttribute(ctx, path.Root("dnssec_flags"), types.Int64Unknown())
	response.Plan.SetAttribute(ctx, path.Root("dnssec_keytag"), types.Int64Unknown())
}

func (r *DnsZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DnsZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf DnsZoneResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)
	dataApi, err := r.client.CreateDnsZone(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create DNS zone", err.Error())
		return
	}

	tflog.Trace(ctx, "created dns zone "+dataApi.Domain)
	dataTf, diags = dnsZoneApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DnsZoneResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetDnsZone(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching dns zone", err.Error()))
		return
	}

	dataTf, diags := dnsZoneApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DnsZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateDnsZone(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating dns zone", err.Error()))
		return
	}

	dataTf, diags := dnsZoneApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DnsZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDnsZone(ctx, data.Id.ValueInt64())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting dns zone", err.Error()))
	}
}

func (r *DnsZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error converting ID to integer", err.Error()))
		return
	}

	dataApi, err := r.client.GetDnsZone(ctx, id)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching dns zone", err.Error()))
		return
	}

	dataTf, diags := dnsZoneApiToTf(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *DnsZoneResource) convertModelToApi(ctx context.Context, dataTf DnsZoneResourceModel) api.DnsZone {
	dataApi := api.DnsZone{}
	dataApi.Id = dataTf.Id.ValueInt64()
	dataApi.Domain = dataTf.Domain.ValueString()
	dataApi.CustomNameserversEnabled = dataTf.NameserverCustom.ValueBool()
	dataApi.Nameserver1 = dataTf.Nameserver1.ValueString()
	dataApi.Nameserver2 = dataTf.Nameserver2.ValueString()
	dataApi.SoaEmail = dataTf.SoaEmail.ValueString()
	dataApi.LoggingEnabled = dataTf.LogEnabled.ValueBool()
	dataApi.LoggingIPAnonymizationEnabled = dataTf.LogAnonymized.ValueBool()
	dataApi.LogAnonymizationType = mapValueToKey(dnsZoneLogAnonymizedStyleMap, dataTf.LogAnonymizedStyle.ValueString())
	dataApi.DnssecEnabled = dataTf.DnssecEnabled.ValueBool()

	return dataApi
}

func dnsZoneApiToTf(dataApi api.DnsZone) (DnsZoneResourceModel, diag.Diagnostics) {
	dataTf := DnsZoneResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.Domain = types.StringValue(dataApi.Domain)
	dataTf.NameserverCustom = types.BoolValue(dataApi.CustomNameserversEnabled)
	dataTf.Nameserver1 = types.StringValue(dataApi.Nameserver1)
	dataTf.Nameserver2 = types.StringValue(dataApi.Nameserver2)
	dataTf.SoaEmail = types.StringValue(dataApi.SoaEmail)
	dataTf.LogEnabled = types.BoolValue(dataApi.LoggingEnabled)
	dataTf.LogAnonymized = types.BoolValue(dataApi.LoggingIPAnonymizationEnabled)
	dataTf.LogAnonymizedStyle = types.StringValue(mapKeyToValue(dnsZoneLogAnonymizedStyleMap, dataApi.LogAnonymizationType))
	dataTf.DnssecEnabled = types.BoolValue(dataApi.DnssecEnabled)
	dataTf.DnssecAlgorithm = types.Int64Value(int64(dataApi.DnssecAlgorithm))
	dataTf.DnssecPublicKey = types.StringValue(dataApi.DnssecPublicKey)
	dataTf.DnssecDigest = types.StringValue(dataApi.DnssecDigest)
	dataTf.DnssecDigestType = types.Int64Value(int64(dataApi.DnssecDigestType))
	dataTf.DnssecFlags = types.Int64Value(int64(dataApi.DnssecFlags))
	dataTf.DnssecKeytag = types.Int64Value(int64(dataApi.DnssecKeyTag))

	return dataTf, nil
}
