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
)

var _ datasource.DataSource = &ComputeContainerAppContainerDataSource{}

func NewComputeContainerAppContainerDataSource() datasource.DataSource {
	return &ComputeContainerAppContainerDataSource{}
}

type ComputeContainerAppContainerDataSource struct {
	client *api.Client
}

type ComputeContainerAppContainerDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	App  types.String `tfsdk:"app"`
	Name types.String `tfsdk:"name"`
}

func (d *ComputeContainerAppContainerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_container_app_container"
}

func (d *ComputeContainerAppContainerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This data source represents a container in a Magic Containers application.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The container unique identifier.",
			},
			"app": schema.StringAttribute{
				Required:    true,
				Description: "The application unique identifier.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The container name.",
			},
		},
	}
}

func (d *ComputeContainerAppContainerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ComputeContainerAppContainerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ComputeContainerAppContainerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetComputeContainerApp(ctx, data.App.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not fetch Compute Container App", err.Error())
		return
	}

	containerFound := false
	for _, container := range app.ContainerTemplates {
		if container.Name != data.Name.ValueString() {
			continue
		}

		containerFound = true
		data.Id = types.StringValue(container.Id)
	}

	if !containerFound {
		resp.Diagnostics.AddError("Could not find Compute Container App container", fmt.Sprintf("A container with ID \"%s\" could not be found for this App.", data.Name.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
