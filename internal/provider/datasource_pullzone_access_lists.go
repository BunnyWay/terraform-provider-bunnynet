// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PullzoneAccessListsDataSource{}

func NewPullzoneAccessListsDataSource() datasource.DataSource {
	return &PullzoneAccessListsDataSource{}
}

type PullzoneAccessListsDataSource struct {
	client *api.Client
}

type PullzoneAccessListsDataSourceModel struct {
	Pullzone types.Int64 `tfsdk:"pullzone"`
	Custom   types.Bool  `tfsdk:"custom"`
	Data     types.Map   `tfsdk:"data"`
}

func (d *PullzoneAccessListsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone_access_lists"
}

var pullzoneAccessListDataSourceType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id":   types.Int64Type,
		"name": types.StringType,
		"type": types.StringType,
	},
}

func (d *PullzoneAccessListsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource represents an Access List for a bunny.net pullzone.",

		Attributes: map[string]schema.Attribute{
			"pullzone": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the linked pullzone.",
			},
			"custom": schema.BoolAttribute{
				Required:    true,
				Description: "Select custom or curated Access Lists.",
			},
			"data": schema.MapAttribute{
				ElementType: pullzoneAccessListDataSourceType,
				Computed:    true,
			},
		},
	}
}

func (d *PullzoneAccessListsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PullzoneAccessListsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PullzoneAccessListsDataSourceModel
	req.Config.Get(ctx, &data)

	var query api.PullzoneAccessListQuery = api.PullzoneAccessListQueryCurated
	if data.Custom.ValueBool() {
		query = api.PullzoneAccessListQueryCustom
	}

	lists, err := d.client.GetPullzoneAccessLists(ctx, data.Pullzone.ValueInt64(), query)
	if err != nil {
		resp.Diagnostics.AddError("Could not fetch Access Lists", err.Error())
		return
	}

	objs := map[string]attr.Value{}

	for _, list := range lists {
		obj, diags := types.ObjectValue(pullzoneAccessListDataSourceType.AttrTypes, map[string]attr.Value{
			"id":   types.Int64Value(list.Id),
			"name": types.StringValue(list.Name),
			"type": types.StringValue(mapKeyToValue(pullzoneAccessListTypeMap, list.Type)),
		})

		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		objs[list.Name] = obj
	}

	dataMap, diags := types.MapValue(pullzoneAccessListDataSourceType, objs)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Data = dataMap
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
