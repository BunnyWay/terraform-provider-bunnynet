// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ComputeContainerAppContainerEndpointDataSource{}

func NewComputeContainerAppContainerEndpointDataSource() datasource.DataSource {
	return &ComputeContainerAppContainerEndpointDataSource{}
}

type ComputeContainerAppContainerEndpointDataSource struct {
	client *api.Client
}

type ComputeContainerAppContainerEndpointDataSourceModel struct {
	App        types.String `tfsdk:"app"`
	Container  types.String `tfsdk:"container"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	PublicHost types.String `tfsdk:"public_host"`
	CDN        types.Object `tfsdk:"cdn"`
}

func (d *ComputeContainerAppContainerEndpointDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_container_app_container_endpoint"
}

func (d *ComputeContainerAppContainerEndpointDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source represents a container endpoint in a Magic Containers application.",

		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Required:    true,
				Description: "The application unique identifier.",
			},
			"container": schema.StringAttribute{
				Required:    true,
				Description: "The container unique identifier.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The endpoint name.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The endpoint type.",
			},
			"public_host": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The publicly accessible `IP:port` or hostname for the endpoint.",
			},
			"cdn": schema.ObjectAttribute{
				Computed:       true,
				Description:    "Configurations for CDN endpoints.",
				AttributeTypes: computeContainerAppContainerEndpointCdnType.AttrTypes,
			},
		},
	}
}

func (d *ComputeContainerAppContainerEndpointDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ComputeContainerAppContainerEndpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ComputeContainerAppContainerEndpointDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetComputeContainerApp(ctx, data.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not fetch Compute Container App", err.Error())
		return
	}

	endpointTypeMap := utils.MapInvert(computeContainerAppContainerEndpointTypeMap)
	containerFound := false

	for _, container := range app.ContainerTemplates {
		if container.Id != data.Container.ValueString() {
			continue
		}

		containerFound = true
		endpointFound := false

		for _, endpoint := range container.Endpoints {
			if endpoint.DisplayName != data.Name.ValueString() {
				continue
			}

			endpointFound = true
			data.Type = types.StringValue(endpointTypeMap[endpoint.Type])
			data.PublicHost = types.StringValue(endpoint.PublicHost)
			data.CDN = types.ObjectNull(computeContainerAppContainerEndpointCdnType.AttrTypes)

			if endpoint.Type == "CDN" {
				cdnSet, diags := convertContainerAppContainerEndpointCdnApiToTf(endpoint)
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}

				cdnSetElements := cdnSet.Elements()
				if len(cdnSetElements) != 1 {
					resp.Diagnostics.AddError("Unexpected CDN set length", "CDN set should have one item")
					return
				}

				data.CDN = cdnSetElements[0].(types.Object)
			}

			break
		}

		if !endpointFound {
			resp.Diagnostics.AddError("Could not find Compute Container App container endpoint", fmt.Sprintf("An endpoint named \"%s\" could not be found for this container.", data.Name.ValueString()))
			return
		}
	}

	if !containerFound {
		resp.Diagnostics.AddError("Could not find Compute Container App container", fmt.Sprintf("A container with ID \"%s\" could not be found for this App.", data.Container.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
