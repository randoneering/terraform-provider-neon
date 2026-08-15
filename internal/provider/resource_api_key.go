// Package provider implements the Neon Terraform provider on
// terraform-plugin-framework. Resources migrated off terraform-plugin-sdk/v2
// land here.
package provider

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	neon "github.com/kislerdm/neon-sdk-go"
)

// neonAPIKeyResource is the Plugin Framework implementation of neon_api_key.
//
// This is the pilot resource for the SDK v2 -> Plugin Framework migration
// (see openspec/changes/migrate-sdk-v2-to-framework/design.md). The schema
// is intentionally identical to provider/resource_api_key.go so existing
// state files apply without re-import.
type neonAPIKeyResource struct {
	provider *neonProvider
}

func (r *neonAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *neonAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A key to access the Neon API.\n\n~>**WARNING** The resource does not support import.",
		Version:     1,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the API Key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The API key ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The generated 64-bit token required to access the Neon API.",
			},
		},
	}
}

type neonAPIKeyResourceModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.String `tfsdk:"id"`
	Key  types.String `tfsdk:"key"`
}

func (r *neonAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan neonAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.provider == nil || r.provider.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The Neon provider has not been configured.")
		return
	}

	createResp, err := r.provider.client.CreateApiKey(neon.ApiKeyCreateRequest{
		KeyName: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating API key", err.Error())
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(createResp.ID, 10))
	plan.Key = types.StringValue(createResp.Key)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *neonAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state neonAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.provider == nil || r.provider.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The Neon provider has not been configured.")
		return
	}

	keys, err := r.provider.client.ListApiKeys()
	if err != nil {
		resp.Diagnostics.AddError("Error listing API keys", err.Error())
		return
	}

	keyName := state.Name.ValueString()
	found := slices.ContainsFunc(keys, func(k neon.ApiKeysListResponseItem) bool {
		return keyName == k.Name
	})
	if !found {
		tflog.Debug(ctx, "API key not found, removing from state", map[string]interface{}{"name": keyName})
		resp.State.RemoveResource(ctx)
		return
	}

	for _, k := range keys {
		if keyName == k.Name {
			state.ID = types.StringValue(strconv.FormatInt(k.ID, 10))
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *neonAPIKeyResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// No-op: all attributes are ForceNew or Computed.
}

func (r *neonAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state neonAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.provider == nil || r.provider.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The Neon provider has not been configured.")
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid API key ID", err.Error())
		return
	}

	if _, err := r.provider.client.RevokeApiKey(id); err != nil {
		resp.Diagnostics.AddError(
			"Error revoking API key",
			fmt.Sprintf("id=%d: %s", id, err.Error()),
		)
		return
	}
}