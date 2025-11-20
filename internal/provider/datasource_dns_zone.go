// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &DnsZoneDataSource{}

func NewDnsZoneDataSource() datasource.DataSource {
	return &DnsZoneDataSource{}
}

type DnsZoneDataSource struct {
	client *api.Client
}

func (d *DnsZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (d *DnsZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source represents a DNS zone in [Bunny DNS](https://bunny.net/dns/).",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: dnsZoneDescription.Id,
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: dnsZoneDescription.Domain,
			},
			"nameserver_custom": schema.BoolAttribute{
				Computed:    true,
				Description: dnsZoneDescription.NameserverCustom,
			},
			"nameserver1": schema.StringAttribute{
				Computed:    true,
				Description: dnsZoneDescription.Nameserver1,
			},
			"nameserver2": schema.StringAttribute{
				Computed:    true,
				Description: dnsZoneDescription.Nameserver2,
			},
			"soa_email": schema.StringAttribute{
				Computed:    true,
				Description: dnsZoneDescription.SoaEmail,
			},
			"log_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: dnsZoneDescription.LogEnabled,
			},
			"log_anonymized": schema.BoolAttribute{
				Computed:    true,
				Description: dnsZoneDescription.LogAnonymized,
			},
			"log_anonymized_style": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.LogAnonymizedStyle,
			},
			"dnssec_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: dnsZoneDescription.DnssecEnabled,
			},
			"dnssec_algorithm": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecAlgorithm,
			},
			"dnssec_digest": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecDigest,
			},
			"dnssec_digest_type": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecDigestType,
			},
			"dnssec_flags": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecFlags,
			},
			"dnssec_keytag": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: dnsZoneDescription.DnssecKeyTag,
			},
		},
	}
}

func (d *DnsZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DnsZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DnsZoneResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := d.client.GetDnsZoneByDomain(ctx, data.Domain.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not fetch DNS zone", err.Error())
		return
	}

	dataResult, diags := dnsZoneApiToTf(zone)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataResult)...)
}
