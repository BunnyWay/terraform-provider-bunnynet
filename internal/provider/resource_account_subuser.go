// Copyright (c) BunnyWay d.o.o.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/api"
	"github.com/bunnyway/terraform-provider-bunnynet/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/maps"
	"regexp"
)

var _ resource.Resource = &AccountSubuserResource{}
var _ resource.ResourceWithConfigure = &AccountSubuserResource{}
var _ resource.ResourceWithImportState = &AccountSubuserResource{}

func NewAccountSubuserResource() resource.Resource {
	return &AccountSubuserResource{}
}

type AccountSubuserResource struct {
	client *api.Client
}

type AccountSubuserResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Firstname   types.String `tfsdk:"firstname"`
	Lastname    types.String `tfsdk:"lastname"`
	Email       types.String `tfsdk:"email"`
	Password    types.String `tfsdk:"password"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func (r *AccountSubuserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_subuser"
}

func (r *AccountSubuserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages an account sub-user in bunny.net.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The user's identifier.",
			},
			"firstname": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The user's first name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"lastname": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The user's last name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"email": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The user's email address.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[^@]+@[^.]+\..+[^.]+$`), "must be a public email address"),
				},
			},
			"password": schema.StringAttribute{
				Optional:  true,
				WriteOnly: true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The user's password. This is a write-only field: to update it after creation, re-create the resource.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(12),
					stringvalidator.RegexMatches(regexp.MustCompile("[a-z]+"), "must contain a lower-case character"),
					stringvalidator.RegexMatches(regexp.MustCompile("[A-Z]+"), "must contain an upper-case character"),
					stringvalidator.RegexMatches(regexp.MustCompile("[0-9]+"), "must contain a digit"),
					stringvalidator.RegexMatches(regexp.MustCompile("[^a-zA-Z0-9]+"), "must contain a special character"),
				},
			},
			"permissions": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: generateMarkdownMapOptions(accountSubuserPermissionsMap),
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(maps.Keys(accountSubuserPermissionsMap)...),
					),
				},
			},
		},
	}
}

func (r *AccountSubuserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AccountSubuserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var dataTf AccountSubuserResourceModel
	var diags diag.Diagnostics
	resp.Diagnostics.Append(req.Plan.Get(ctx, &dataTf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, dataTf)

	// read write-only password
	var password types.String
	req.Config.GetAttribute(ctx, path.Root("password"), &password)
	dataApi.Password = password.ValueString()

	if dataApi.Password == "" {
		resp.Diagnostics.AddAttributeError(path.Root("password"), "Missing attribute value", "The password cannot be empty")
		return
	}

	dataApi, err := r.client.CreateAccountSubuser(ctx, dataApi)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create account sub-user", err.Error())
		return
	}

	tflog.Trace(ctx, "created account sub-user "+dataApi.Email)
	dataTf, diags = r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *AccountSubuserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccountSubuserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi, err := r.client.GetAccountSubuser(ctx, data.Id.ValueString())
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching account sub-user", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *AccountSubuserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AccountSubuserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataApi := r.convertModelToApi(ctx, data)
	dataApi, err := r.client.UpdateAccountSubuser(ctx, dataApi)

	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error updating account sub-user", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *AccountSubuserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccountSubuserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAccountSubuser(ctx, data.Id.ValueString())
	if err != nil && !errors.Is(err, api.ErrNotFound) {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error deleting account sub-user", err.Error()))
	}
}

func (r *AccountSubuserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	dataApi, err := r.client.GetAccountSubuserByEmail(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("Error fetching account sub-user", err.Error()))
		return
	}

	dataTf, diags := r.convertApiToModel(dataApi)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &dataTf)...)
}

func (r *AccountSubuserResource) convertModelToApi(ctx context.Context, dataTf AccountSubuserResourceModel) api.AccountSubuser {
	dataApi := api.AccountSubuser{}
	dataApi.Id = dataTf.Id.ValueString()
	dataApi.Firstname = dataTf.Firstname.ValueString()
	dataApi.Lastname = dataTf.Lastname.ValueString()
	dataApi.Email = dataTf.Email.ValueString()

	{
		permList := dataTf.Permissions.Elements()
		values := make([]string, 0, len(permList))
		for _, perm := range permList {
			values = append(values, accountSubuserPermissionsMap[perm.(types.String).ValueString()])
		}
		dataApi.Roles = values
	}

	return dataApi
}

func (r *AccountSubuserResource) convertApiToModel(dataApi api.AccountSubuser) (AccountSubuserResourceModel, diag.Diagnostics) {
	dataTf := AccountSubuserResourceModel{}
	dataTf.Id = types.StringValue(dataApi.Id)
	dataTf.Firstname = types.StringValue(dataApi.Firstname)
	dataTf.Lastname = types.StringValue(dataApi.Lastname)
	dataTf.Email = types.StringValue(dataApi.Email)

	{
		permissionMap := utils.MapInvert(accountSubuserPermissionsMap)
		permissions := make([]attr.Value, 0, len(dataApi.Roles))
		for _, perm := range dataApi.Roles {
			permissions = append(permissions, types.StringValue(permissionMap[perm]))
		}

		set, diags := types.SetValue(types.StringType, permissions)
		if diags != nil {
			return dataTf, diags
		}

		dataTf.Permissions = set
	}

	return dataTf, nil
}
