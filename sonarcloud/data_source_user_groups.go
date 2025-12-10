package sonarcloud

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reinoudk/go-sonarcloud/sonarcloud/user_groups"
)

type dataSourceUserGroups struct {
	p *sonarcloudProvider
}

var _ datasource.DataSource = &dataSourceUserGroups{}

func (d *dataSourceUserGroups) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}

func (d *dataSourceUserGroups) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source retrieves a list of user groups for the configured organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"groups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The groups of this organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the user group.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the user group.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "The description of the user group.",
						},
						"members_count": schema.Float64Attribute{
							Computed:    true,
							Description: "The number of members in this user group.",
						},
						"default": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether new members are added to this user group per default or not.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceUserGroups) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	request := user_groups.SearchRequest{}

	res, err := d.p.client.UserGroups.SearchAll(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read user_groups",
			fmt.Sprintf("The SearchAll request returned an error: %+v", err),
		)
		return
	}

	result := Groups{}
	allGroups := make([]Group, len(res.Groups))
	for i, group := range res.Groups {
		allGroups[i] = Group{
			ID:           types.StringValue(big.NewFloat(group.Id).String()),
			Default:      types.BoolValue(group.Default),
			Description:  types.StringValue(group.Description),
			MembersCount: types.NumberValue(big.NewFloat(group.MembersCount)),
			Name:         types.StringValue(group.Name),
		}
	}
	result.Groups = allGroups
	result.ID = types.StringValue(d.p.organization)

	diags := resp.State.Set(ctx, result)

	resp.Diagnostics.Append(diags...)
}
