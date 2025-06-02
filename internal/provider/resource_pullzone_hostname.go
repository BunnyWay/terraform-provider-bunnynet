// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/boolvalidator"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/pullzonehostnameresourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	Id             types.Int64  `tfsdk:"id"`
	PullzoneId     types.Int64  `tfsdk:"pullzone"`
	Name           types.String `tfsdk:"name"`
	IsInternal     types.Bool   `tfsdk:"is_internal"`
	TLSEnabled     types.Bool   `tfsdk:"tls_enabled"`
	ForceSSL       types.Bool   `tfsdk:"force_ssl"`
	Certificate    types.String `tfsdk:"certificate"`
	CertificateKey types.String `tfsdk:"certificate_key"`
}

func (r *PullzoneHostnameResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_hostname"
}

func (r *PullzoneHostnameResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages custom hostnames for a bunny.net pull zone. It is used to add and configure custom hostnames for pullzones.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique ID of the hostname.",
			},
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
					stringvalidator.RegexMatches(regexp.MustCompile(`(.+)\.(.+)`), "Invalid domain"),
				},
				Description: "The hostname value for the domain name.",
			},
			"is_internal": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the hostname is internal (in the CDN domain) or provided by the user.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"tls_enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: `Indicates whether the hostname should support HTTPS. If a custom certificate is not provided via the <code>certificate</code> attribute, a Domain-validated TLS certificate will be automatically obtained and managed by Bunny. ***Important***: it is not possible to tell managed and custom certificates apart for imported resources.`,
			},
			"force_ssl": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Bool{
					boolvalidator.RequiresAttributeAsTrue("tls_enabled"),
				},
				Description: "Indicates whether SSL should be enforced for the hostname.",
			},
			"certificate": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("certificate_key")),
				},
				MarkdownDescription: `The certificate for the hostname, in PEM format. ***Important***: the Bunny API will not return the certificate data, so you'll have to make sure you're importing the correct certificate.`,
			},
			"certificate_key": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("certificate")),
				},
				MarkdownDescription: `The certificate private key for the hostname, in PEM format. ***Important***: the Bunny API will not return the certificate key, so you'll have to make sure you're importing the correct certificate key.`,
			},
		},
	}
}

func (r *PullzoneHostnameResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		pullzonehostnameresourcevalidator.CustomCertificate(),
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
	hostname := dataApi.Name
	pullzoneId := dataApi.PullzoneId
	certificate := dataApi.Certificate
	certificateKey := dataApi.CertificateKey

	dataApi, err := r.client.CreatePullzoneHostname(dataApi)
	if err != nil {
		re := regexp.MustCompile(`loadFreeCertificate failed: The domain .* is not pointing to our servers\.`)
		if re.MatchString(err.Error()) {
			err2 := r.client.DeletePullzoneHostname(pullzoneId, hostname)
			if err2 != nil {
				tflog.Error(ctx, fmt.Sprintf("Delete hostname for pullzone %d failed: %s", pullzoneId, err2.Error()))
				resp.Diagnostics.AddWarning("pullzone_hostname is in a dirty state", "The hostname creation failed, and unfortunately the cleanup did too. You'll have to manually remove the hostname in dash.bunny.net, or import it in terraform to continue.")
			}
		}

		resp.Diagnostics.AddError("Unable to create hostname", err.Error())
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("created hostname for pullzone %d", pullzoneId))
	dataApi.Certificate = certificate
	dataApi.CertificateKey = certificateKey
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
	dataApi.Certificate = data.Certificate.ValueString()
	dataApi.CertificateKey = data.CertificateKey.ValueString()

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

	var previousData PullzoneHostnameResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &previousData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	previousDataApi := r.convertModelToApi(ctx, previousData)

	dataApiResult, err := r.client.UpdatePullzoneHostname(dataApi, previousDataApi)
	if len(dataApi.Certificate) > 0 {
		dataApiResult.Certificate = dataApi.Certificate
	}

	if len(dataApi.CertificateKey) > 0 {
		dataApiResult.CertificateKey = dataApi.CertificateKey
	}

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating hostname", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApiResult)
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

	// b-cdn.net hostnames cannot be deleted
	if data.IsInternal.ValueBool() {
		return
	}

	err := r.client.DeletePullzoneHostname(data.PullzoneId.ValueInt64(), data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting hostname", err.Error()))
	}
}

func (r *PullzoneHostnameResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	pullzoneIdStr, hostname, ok := strings.Cut(req.ID, "|")
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding hostname", "Use \"<pullzoneId>|<hostname>\" as ID on terraform import command"))
		return
	}

	pullzoneId, err := strconv.ParseInt(pullzoneIdStr, 10, 64)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding hostname", "Invalid pullzone ID: "+err.Error()))
		return
	}

	dataApi, err := r.client.GetPullzoneHostnameByName(pullzoneId, hostname)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error finding hostname", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
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
	dataApi.IsSystemHostname = dataTf.IsInternal.ValueBool()
	dataApi.HasCertificate = dataTf.TLSEnabled.ValueBool()
	dataApi.ForceSSL = dataTf.ForceSSL.ValueBool()
	dataApi.Certificate = dataTf.Certificate.ValueString()
	dataApi.CertificateKey = dataTf.CertificateKey.ValueString()

	return dataApi
}

func (r *PullzoneHostnameResource) convertApiToModel(dataApi api.PullzoneHostname) (PullzoneHostnameResourceModel, diag.Diagnostics) {
	dataTf := PullzoneHostnameResourceModel{}
	dataTf.Id = types.Int64Value(dataApi.Id)
	dataTf.PullzoneId = types.Int64Value(dataApi.PullzoneId)
	dataTf.Name = types.StringValue(dataApi.Name)
	dataTf.IsInternal = types.BoolValue(dataApi.IsSystemHostname)
	dataTf.TLSEnabled = types.BoolValue(dataApi.HasCertificate)
	dataTf.ForceSSL = types.BoolValue(dataApi.ForceSSL)
	dataTf.Certificate = types.StringValue(dataApi.Certificate)
	dataTf.CertificateKey = types.StringValue(dataApi.CertificateKey)

	return dataTf, nil
}
