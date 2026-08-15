// Package provider implements the Plugin Framework provider for Neon.
//
// This file is the pilot scaffolding for the SDK v2 -> Plugin Framework
// migration. It registers neon_api_key only; the remaining resources stay
// on SDK v2 in provider/ until they migrate. Once all resources are
// migrated, this package will replace provider/ entirely.
package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	neon "github.com/kislerdm/neon-sdk-go"
)

const Name = "neon/neon"

// neonProvider is the framework Provider implementation. During the pilot
// it lists only the neon_api_key resource; data sources remain on SDK v2.
//
// The configured *neon.Client is stored on the struct so resources can
// reach it through the provider reference (the framework does not pass
// ProviderData to resource methods; this is the standard pattern).
type neonProvider struct {
	client *neon.Client
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &neonProvider{}
	}
}

type neonProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (p *neonProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "neon"
}

func (p *neonProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Description: "API access key. Default is read from the environment variable `NEON_API_KEY`.",
			},
		},
	}
}

func (p *neonProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config neonProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("NEON_API_KEY")
	}

	client, err := neon.NewClient(neon.Config{Key: apiKey})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Neon API client",
			fmt.Sprintf("ensure api_key is set or NEON_API_KEY env var is exported: %s", err.Error()),
		)
		return
	}

	p.client = client
}

func (p *neonProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return &neonAPIKeyResource{provider: p}
		},
	}
}

func (p *neonProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	// Data sources stay on SDK v2 in provider/ until they migrate.
	return nil
}