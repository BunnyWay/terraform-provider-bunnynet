// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

var _ datasource.DataSource = &ComputeContainerImageRegistryDataSource{}

func NewComputeContainerImageRegistryDataSource() datasource.DataSource {
	return &ComputeContainerImageRegistryDataSource{}
}

type ComputeContainerImageRegistryDataSource struct {
	client *api.Client
}

type ComputeContainerImageRegistryDataSourceModel struct {
	Id       types.Int64  `tfsdk:"id"`
	Registry types.String `tfsdk:"registry"`
	Username types.String `tfsdk:"username"`
}

func (d *ComputeContainerImageRegistryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compute_container_imageregistry"
}

func (d *ComputeContainerImageRegistryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source represents an Image Registry connection for Magic Containers in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier for the image registry.",
			},
			"registry": schema.StringAttribute{
				Required:    true,
				Description: generateMarkdownSliceOptions(computeContainerImageRegistryRegistryOptions),
				Validators: []validator.String{
					stringvalidator.OneOf(computeContainerImageRegistryRegistryOptions...),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username used to authenticate to the registry.",
			},
		},
	}
}

func (d *ComputeContainerImageRegistryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ComputeContainerImageRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ComputeContainerImageRegistryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := d.client.FindComputeContainerImageregistry(data.Registry.ValueString(), data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not find the Compute Image Registry", err.Error())
		return
	}

	id, err := strconv.ParseInt(dataApi.Id, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Error converting ID to integer", err.Error())
		return
	}

	data.Id = types.Int64Value(id)
	data.Registry = types.StringValue(dataApi.DisplayName)
	data.Username = types.StringValue(dataApi.UserName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
