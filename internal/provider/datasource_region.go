// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &RegionDataSource{}

func NewRegionDataSource() datasource.DataSource {
	return &RegionDataSource{}
}

type RegionDataSource struct {
	client *api.Client
}

type RegionDataSourceModel struct {
	Id                  types.Int64   `tfsdk:"id"`
	Name                types.String  `tfsdk:"name"`
	PricePerGigabyte    types.Float64 `tfsdk:"price_per_gigabyte"`
	RegionCode          types.String  `tfsdk:"region_code"`
	ContinentCode       types.String  `tfsdk:"continent_code"`
	CountryCode         types.String  `tfsdk:"country_code"`
	Latitude            types.Float64 `tfsdk:"latitude"`
	Longitude           types.Float64 `tfsdk:"longitude"`
	AllowLatencyRouting types.Bool    `tfsdk:"allow_latency_routing"`
}

func (d *RegionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

func (d *RegionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source provides information about the regions supported by bunny.net. It is used to retrieve a list of regions where bunny.net services are available, which can then be used to configure other resources within those specific regions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Id",
				Computed:            true,
				Description:         "The unique identifier for the region.",
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name",
				Computed:            true,
				Description:         "The name of the region.",
			},
			"price_per_gigabyte": schema.Float64Attribute{
				MarkdownDescription: "Price per Gigabyte",
				Computed:            true,
				Description:         "The cost per gigabyte of data transfer in this region.",
			},
			"region_code": schema.StringAttribute{
				MarkdownDescription: "Region Code",
				Required:            true,
				Description:         "The code representing a specific region.",
			},
			"continent_code": schema.StringAttribute{
				MarkdownDescription: "Continent Code",
				Computed:            true,
				Description:         "The code representing the continent where the region is located.",
			},
			"country_code": schema.StringAttribute{
				MarkdownDescription: "Country Code",
				Computed:            true,
				Description:         "The code representing the country where the region is located.",
			},
			"latitude": schema.Float64Attribute{
				MarkdownDescription: "Latitude",
				Computed:            true,
				Description:         "The latitude coordinate of the region.",
			},
			"longitude": schema.Float64Attribute{
				MarkdownDescription: "Longitude",
				Computed:            true,
				Description:         "The longitude coordinate of the region.",
			},
			"allow_latency_routing": schema.BoolAttribute{
				MarkdownDescription: "Allow Latency Routing",
				Computed:            true,
				Description:         "Indicates whether latency routing is allowed for this region.",
			},
		},
	}
}

func (d *RegionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RegionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RegionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	region, err := d.client.GetRegion(data.RegionCode.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read region, got error: %s", err))
		return
	}

	data.Id = types.Int64Value(region.Id)
	data.Name = types.StringValue(region.Name)
	data.PricePerGigabyte = types.Float64Value(region.PricePerGigabyte)
	data.RegionCode = types.StringValue(region.RegionCode)
	data.ContinentCode = types.StringValue(region.ContinentCode)
	data.CountryCode = types.StringValue(region.CountryCode)
	data.Latitude = types.Float64Value(region.Latitude)
	data.Longitude = types.Float64Value(region.Longitude)
	data.AllowLatencyRouting = types.BoolValue(region.AllowLatencyRouting)

	tflog.Trace(ctx, "read region "+region.RegionCode)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
