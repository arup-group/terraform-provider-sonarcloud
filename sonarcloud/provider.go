package sonarcloud

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reinoudk/go-sonarcloud/sonarcloud"
)

func New() provider.Provider {
	return &sonarcloudProvider{}
}

type sonarcloudProvider struct {
	configured   bool
	client       *sonarcloud.Client
	organization string
}

func (p *sonarcloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sonarcloud"
}

func (p *sonarcloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Optional: true,
				Description: "The SonarCloud organization to manage the resources for. This value must be set in the" +
					" `SONARCLOUD_ORGANIZATION` environment variable if left empty.",
			},
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "The token of a user with admin permissions in the organization. This value must be set in" +
					" the `SONARCLOUD_TOKEN` environment variable if left empty.",
			},
		},
	}
}

func (p *sonarcloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var organization string
	if config.Organization.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as organization",
		)
		return
	}

	if config.Organization.IsNull() {
		organization = os.Getenv("SONARCLOUD_ORGANIZATION")
	} else {
		organization = config.Organization.ValueString()
	}

	var token string
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as token",
		)
	}

	if config.Token.IsNull() {
		token = os.Getenv("SONARCLOUD_TOKEN")
	} else {
		token = config.Token.ValueString()
	}

	c := sonarcloud.NewClient(organization, token, nil)
	p.client = c
	p.organization = organization
	p.configured = true
}

func (p *sonarcloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &resourceUserGroup{p: p} },
		func() resource.Resource { return &resourceUserGroupMember{p: p} },
		func() resource.Resource { return &resourceProject{p: p} },
		func() resource.Resource { return &resourceProjectLink{p: p} },
		func() resource.Resource { return &resourceProjectMainBranch{p: p} },
		func() resource.Resource { return &resourceUserToken{p: p} },
		func() resource.Resource { return &resourceQualityGate{p: p} },
		func() resource.Resource { return &resourceQualityGateSelection{p: p} },
		func() resource.Resource { return &resourceUserPermissions{p: p} },
		func() resource.Resource { return &resourceUserGroupPermissions{p: p} },
		func() resource.Resource { return &resourceWebhook{p: p} },
	}
}

func (p *sonarcloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &dataSourceProjects{p: p} },
		func() datasource.DataSource { return &dataSourceProjectLinks{p: p} },
		func() datasource.DataSource { return &dataSourceUserGroup{p: p} },
		func() datasource.DataSource { return &dataSourceUserGroups{p: p} },
		func() datasource.DataSource { return &dataSourceUserGroupMembers{p: p} },
		func() datasource.DataSource { return &dataSourceUserGroupPermissions{p: p} },
		func() datasource.DataSource { return &dataSourceUserPermissions{p: p} },
		func() datasource.DataSource { return &dataSourceQualityGate{p: p} },
		func() datasource.DataSource { return &dataSourceQualityGates{p: p} },
		func() datasource.DataSource { return &dataSourceWebhooks{p: p} },
	}
}

type providerData struct {
	Organization types.String `tfsdk:"organization"`
	Token        types.String `tfsdk:"token"`
}
