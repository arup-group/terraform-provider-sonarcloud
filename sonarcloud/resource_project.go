package sonarcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reinoudk/go-sonarcloud/sonarcloud/projects"
)

type resourceProject struct {
	p *sonarcloudProvider
}

var _ resource.Resource = &resourceProject{}
var _ resource.ResourceWithImportState = &resourceProject{}

func (r *resourceProject) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *resourceProject) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource manages a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the project. **Warning:** forces project recreation when changed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringLengthBetween(1, 255),
				},
			},
			"key": schema.StringAttribute{
				Required:    true,
				Description: "The key of the project. **Warning**: must be globally unique.",
				Validators: []validator.String{
					stringLengthBetween(1, 400),
				},
			},
			"visibility": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The visibility of the project. Use `private` to only share it with your organization." +
					" Use `public` if the project should be visible to everyone. Defaults to the organization's default visibility." +
					" **Note:** private projects are only available when you have a SonarCloud subscription.",
				Validators: []validator.String{
					allowedOptions("public", "private"),
				},
			},
		},
	}
}

func (r *resourceProject) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. "+
				"This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan Project
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fill in api action struct
	request := projects.CreateRequest{
		Name:         plan.Name.ValueString(),
		Organization: r.p.organization,
		Project:      plan.Key.ValueString(),
		Visibility:   plan.Visibility.ValueString(),
	}

	res, err := r.p.client.Projects.Create(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create the project",
			fmt.Sprintf("The Create request returned an error: %+v", err),
		)
		return
	}

	var result = Project{
		ID:         types.StringValue(res.Project.Key),
		Name:       types.StringValue(res.Project.Name),
		Key:        types.StringValue(res.Project.Key),
		Visibility: types.StringValue(plan.Visibility.ValueString()),
	}
	diags = resp.State.Set(ctx, result)

	resp.Diagnostics.Append(diags...)
}

func (r *resourceProject) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state
	var state Project
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fill in api action struct
	request := projects.SearchRequest{
		Projects: state.Key.ValueString(),
	}

	response, err := r.p.client.Projects.SearchAll(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read the project",
			fmt.Sprintf("The SearchAll request returned an error: %+v", err),
		)
		return
	}

	// Check if the resource exists the list of retrieved resources
	if result, ok := findProject(response, state.Key.ValueString()); ok {
		diags = resp.State.Set(ctx, result)
		resp.Diagnostics.Append(diags...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (r *resourceProject) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from state
	var state Project
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan
	var plan Project
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if key changed
	if !state.Key.Equal(plan.Key) {
		request := projects.UpdateKeyRequest{
			From: state.Key.ValueString(),
			To:   plan.Key.ValueString(),
		}

		err := r.p.client.Projects.UpdateKey(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not update the project key",
				fmt.Sprintf("The UpdateKey request returned an error: %+v", err),
			)
			return
		}
	}

	// Check if visibility changed
	if !state.Visibility.Equal(plan.Visibility) {
		request := projects.UpdateVisibilityRequest{
			Project:    plan.Key.ValueString(),
			Visibility: plan.Visibility.ValueString(),
		}

		err := r.p.client.Projects.UpdateVisibility(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not update the project visibility",
				fmt.Sprintf("The UpdateVisibility request returned an error: %+v", err),
			)
			return
		}
	}

	// We don't have a return value, so we have to query it again
	// Fill in api action struct
	searchRequest := projects.SearchRequest{}

	response, err := r.p.client.Projects.SearchAll(searchRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read the project",
			fmt.Sprintf("The SearchAll request returned an error: %+v", err),
		)
		return
	}

	// Check if the resource exists the list of retrieved resources
	if result, ok := findProject(response, plan.Key.ValueString()); ok {
		diags = resp.State.Set(ctx, result)
		resp.Diagnostics.Append(diags...)
	}
}

func (r *resourceProject) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state Project
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := projects.DeleteRequest{
		Project: state.Key.ValueString(),
	}

	err := r.p.client.Projects.Delete(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete the project",
			fmt.Sprintf("The Delete request returned an error: %+v", err),
		)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceProject) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
