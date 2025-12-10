package sonarcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reinoudk/go-sonarcloud/sonarcloud/webhooks"
)

type dataSourceWebhooks struct {
	p *sonarcloudProvider
}

var _ datasource.DataSource = &dataSourceWebhooks{}

func (d *dataSourceWebhooks) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhooks"
}

func (d *dataSourceWebhooks) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This datasource retrieves the list of webhooks for a project or the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project": schema.StringAttribute{
				Optional:    true,
				Description: "The key of the project. If empty, the webhooks of the organization are returned.",
			},
			"webhooks": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The webhooks of this project or organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed:    true,
							Description: "The key of the webhook.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name the webhook.",
						},
						"url": schema.StringAttribute{
							Computed:    true,
							Description: "The url of the webhook.",
						},
						"has_secret": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the webhook has a secret.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceWebhooks) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DataWebhooks
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fill in api action struct
	request := webhooks.ListRequest{
		Organization: d.p.organization,
		Project:      config.Project.ValueString(),
	}

	response, err := d.p.client.Webhooks.List(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read the webhooks",
			fmt.Sprintf("The List request returned an error: %+v", err),
		)
		return
	}

	hooks := make([]DataWebhook, len(response.Webhooks))
	for i, webhook := range response.Webhooks {
		hooks[i] = DataWebhook{
			Key:       types.StringValue(webhook.Key),
			Name:      types.StringValue(webhook.Name),
			HasSecret: types.BoolValue(webhook.HasSecret),
			Url:       types.StringValue(webhook.Url),
		}
	}

	result := DataWebhooks{
		ID:       types.StringValue(fmt.Sprintf("%s-%s", d.p.organization, config.Project.ValueString())),
		Project:  config.Project,
		Webhooks: hooks,
	}

	diags = resp.State.Set(ctx, result)

	resp.Diagnostics.Append(diags...)
}
