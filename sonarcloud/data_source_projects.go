package sonarcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reinoudk/go-sonarcloud/sonarcloud/projects"
)

type dataSourceProjects struct {
	p *sonarcloudProvider
}

var _ datasource.DataSource = &dataSourceProjects{}

func (d *dataSourceProjects) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *dataSourceProjects) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This data source retrieves a list of projects for the configured organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"projects": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The projects of this organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the project. Equals to the project name.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the project.",
						},
						"key": schema.StringAttribute{
							Computed:    true,
							Description: "The key of the project.",
						},
						"visibility": schema.StringAttribute{
							Computed:    true,
							Description: "The visibility of the project.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceProjects) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	request := projects.SearchRequest{}

	response, err := d.p.client.Projects.SearchAll(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read the project",
			fmt.Sprintf("The SearchAll request returned an error: %+v", err),
		)
		return
	}

	result := Projects{}
	allProjects := make([]Project, len(response.Components))
	for i, component := range response.Components {
		allProjects[i] = Project{
			ID:         types.StringValue(component.Name),
			Name:       types.StringValue(component.Name),
			Key:        types.StringValue(component.Key),
			Visibility: types.StringValue(component.Visibility),
		}
	}
	result.Projects = allProjects
	result.ID = types.StringValue(d.p.organization)

	diags := resp.State.Set(ctx, result)

	resp.Diagnostics.Append(diags...)
}
