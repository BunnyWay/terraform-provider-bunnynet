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

var _ datasource.DataSource = &DnsRecordDataSource{}

func NewDnsRecordDataSource() datasource.DataSource {
	return &DnsRecordDataSource{}
}

type DnsRecordDataSource struct {
	client *api.Client
}

func (d *DnsRecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (d *DnsRecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source represents a DNS record in [Bunny DNS](https://bunny.net/dns/).",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: dnsRecordDescription.Id,
			},
			"zone": schema.Int64Attribute{
				Required:    true,
				Description: dnsRecordDescription.Zone,
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: dnsRecordDescription.Enabled,
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: dnsRecordDescription.Type,
			},
			"ttl": schema.Int64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.TTL,
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Description: dnsRecordDescription.Value,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: dnsRecordDescription.Name,
			},
			"weight": schema.Int64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.Weight,
			},
			"priority": schema.Int64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.Priority,
			},
			"port": schema.Int64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.Port,
			},
			"flags": schema.Int64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.Flags,
			},
			"tag": schema.StringAttribute{
				Computed:    true,
				Description: dnsRecordDescription.Tag,
			},
			"accelerated": schema.BoolAttribute{
				Computed:    true,
				Description: dnsRecordDescription.Accelerated,
			},
			"accelerated_pullzone": schema.Int64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.AcceleratedPullzone,
			},
			"link_name": schema.StringAttribute{
				Computed:    true,
				Description: dnsRecordDescription.Link,
			},
			"monitor_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: dnsRecordDescription.MonitorType,
			},
			"geolocation_lat": schema.Float64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.GeolocationLat,
			},
			"geolocation_long": schema.Float64Attribute{
				Computed:    true,
				Description: dnsRecordDescription.GeolocationLong,
			},
			"latency_zone": schema.StringAttribute{
				Computed:    true,
				Description: dnsRecordDescription.LatencyZone,
			},
			"smart_routing_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: dnsRecordDescription.SmartRoutingType,
			},
			"comment": schema.StringAttribute{
				Computed:    true,
				Description: dnsRecordDescription.Comment,
			},
		},
	}
}

func (d *DnsRecordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DnsRecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DnsRecordResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := d.client.GetDnsZone(data.Zone.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Could not fetch DNS zone", err.Error())
		return
	}

	for _, record := range zone.Records {
		if !data.Id.IsNull() && record.Id != data.Id.ValueInt64() {
			continue
		}

		if record.Name != data.Name.ValueString() || record.Type != mapValueToKey(dnsRecordTypeMap, data.Type.ValueString()) {
			continue
		}

		record.Zone = zone.Id
		dataResult, diags := dnsRecordApiToTf(record)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &dataResult)...)
		return
	}

	resp.Diagnostics.AddError("Could not find DNS record", fmt.Sprintf("DNS zone does not have record %s %s", data.Type.ValueString(), data.Name.ValueString()))
}
