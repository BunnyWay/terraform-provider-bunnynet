// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

var _ datasource.DataSource = &PullzoneDataSource{}

func NewPullzoneDataSource() datasource.DataSource {
	return &PullzoneDataSource{}
}

type PullzoneDataSource struct {
	client *api.Client
}

func (d *PullzoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pullzone"
}

func (d *PullzoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	rReq := resource.SchemaRequest{}
	rResp := resource.SchemaResponse{}

	r := PullzoneResource{}
	r.Schema(ctx, rReq, &rResp)

	if rResp.Diagnostics.HasError() {
		resp.Diagnostics = rResp.Diagnostics
		return
	}

	schemaAttributes := make(map[string]dschema.Attribute, len(rResp.Schema.Attributes))
	schemaBlocks := make(map[string]dschema.Block, len(rResp.Schema.Blocks))

	for k, v := range rResp.Schema.Attributes {
		if k == "id" {
			schemaAttributes[k] = dschema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Description:         v.GetDescription(),
				MarkdownDescription: v.GetMarkdownDescription(),
			}
			continue
		}

		if k == "name" {
			schemaAttributes[k] = dschema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         v.GetDescription(),
				MarkdownDescription: v.GetMarkdownDescription(),
			}
			continue
		}

		schemaAttributes[k] = resourceAttrToDatasourceAttr(v)
	}

	for k, v := range rResp.Schema.Blocks {
		rBlock := v.(rschema.SingleNestedBlock)
		blockAttributes := make(map[string]dschema.Attribute, len(rBlock.Attributes))

		for attrK, attrV := range rBlock.Attributes {
			blockAttributes[attrK] = resourceAttrToDatasourceAttr(attrV)
		}

		schemaBlocks[k] = dschema.SingleNestedBlock{
			Attributes: blockAttributes,
		}
	}

	resp.Schema = dschema.Schema{
		MarkdownDescription: "This data source represents a bunny.net Pullzone.",
		Attributes:          schemaAttributes,
		Blocks:              schemaBlocks,
	}
}

func (d *PullzoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PullzoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PullzoneResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueInt64()
	name := data.Name.ValueString()

	if id == 0 && name == "" {
		resp.Diagnostics.AddError("Missing identifier attribute", "Either `id` or `name` attribute must be specified.")
		return
	}

	if id > 0 && name != "" {
		resp.Diagnostics.AddError("Ambiguous identifier attribute", "Only one of `id` or `name` attribute must be specified.")
		return
	}

	var zone api.Pullzone
	var err error

	if id > 0 {
		zone, err = d.client.GetPullzone(id)
	} else {
		zone, err = d.client.GetPullzoneByName(ctx, name)
	}

	if err != nil {
		resp.Diagnostics.AddError("Could not fetch pullzone", err.Error())
		return
	}

	dataResult, diags := pullzoneApiToTf(zone)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataResult)...)
}
