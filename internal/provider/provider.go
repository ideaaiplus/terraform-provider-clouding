// Copyright (c) ideaaiplus
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ideaaiplus/terraform-provider-clouding/internal/client"
)

var _ provider.Provider = &cloudingProvider{}

type cloudingProvider struct {
	version string
}

type cloudingProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

// New returns a provider factory consumed by providerserver.Serve.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &cloudingProvider{version: version}
	}
}

func (p *cloudingProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "clouding"
	resp.Version = p.version
}

func (p *cloudingProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Clouding.io resources.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The X-API-KEY used to authenticate with the Clouding.io API.",
			},
		},
	}
}

func (p *cloudingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data cloudingProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIKey.IsNull() || data.APIKey.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires api_key to be set.",
		)
		return
	}

	c := client.New(data.APIKey.ValueString())
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *cloudingProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
	}
}

func (p *cloudingProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
