// Copyright (c) ideaaiplus
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ideaaiplus/terraform-provider-clouding/internal/client"
)

var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithConfigure   = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
)

type serverResource struct {
	client *client.Client
}

type serverResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Hostname types.String `tfsdk:"hostname"`
	ImageID  types.String `tfsdk:"image_id"`
	FlavorID types.String `tfsdk:"flavor_id"`
	SSHKeyID types.String `tfsdk:"ssh_key_id"`
}

// NewServerResource is the factory registered in provider.Resources().
func NewServerResource() resource.Resource {
	return &serverResource{}
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Clouding.io server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Server identifier assigned by the Clouding.io API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name for the server.",
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "Fully-qualified hostname assigned to the server.",
			},
			"image_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the OS image to deploy on the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flavor_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the hardware flavor (vCPU/RAM profile).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the SSH public key to inject into the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.ServerCreateRequest{
		Name:     plan.Name.ValueString(),
		Hostname: plan.Hostname.ValueString(),
		FlavorID: plan.FlavorID.ValueString(),
		Volume: client.VolumeConfig{
			Source: "Image",
			ID:     plan.ImageID.ValueString(),
			SsdGB:  20,
		},
	}

	if !plan.SSHKeyID.IsNull() && !plan.SSHKeyID.IsUnknown() && plan.SSHKeyID.ValueString() != "" {
		createReq.AccessCfg = &client.AccessConfig{
			SSHKeyID: plan.SSHKeyID.ValueString(),
		}
	}

	server, err := r.client.CreateServer(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating clouding_server", err.Error())
		return
	}

	plan.ID = types.StringValue(server.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(ctx, state.ID.ValueString())
	if err != nil {
		var notFound *client.NotFoundError
		if errors.As(err, &notFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading clouding_server", err.Error())
		return
	}

	state.Name = types.StringValue(server.Name)
	state.Hostname = types.StringValue(server.Hostname)
	state.ImageID = types.StringValue(server.Image.ID)
	state.FlavorID = types.StringValue(server.Flavor)

	if server.SSHKeyID != "" {
		state.SSHKeyID = types.StringValue(server.SSHKeyID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.ServerUpdateRequest{
		Name:     plan.Name.ValueString(),
		Hostname: plan.Hostname.ValueString(),
	}

	_, err := r.client.UpdateServer(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating clouding_server", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteServer(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting clouding_server", err.Error())
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
