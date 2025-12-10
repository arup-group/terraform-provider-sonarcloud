package sonarcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reinoudk/go-sonarcloud/sonarcloud/qualitygates"
)

type resourceQualityGateSelection struct {
	p *sonarcloudProvider
}

var _ resource.Resource = &resourceQualityGateSelection{}

func (r *resourceQualityGateSelection) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quality_gate_selection"
}

func (r *resourceQualityGateSelection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "This resource selects a quality gate for one or more projects",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The implicit ID of the resource",
				Computed:    true,
			},
			"gate_id": schema.StringAttribute{
				Description: "The ID of the quality gate that is selected for the project(s).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_keys": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The Keys of the projects which have been selected on the referenced quality gate",
				Required:    true,
			},
		},
	}
}

func (r *resourceQualityGateSelection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unkown value from another resource. "+
				"This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan Selection
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract project keys from Set
	var projectKeys []string
	diags = plan.ProjectKeys.ElementsAs(ctx, &projectKeys, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, projectKey := range projectKeys {
		// Fill in api action struct for Quality Gates
		request := qualitygates.SelectRequest{
			GateId:       plan.GateId.ValueString(),
			ProjectKey:   projectKey,
			Organization: r.p.organization,
		}
		err := r.p.client.Qualitygates.Select(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not create Quality Gate Selection",
				fmt.Sprintf("The Select request returned an error: %+v", err),
			)
			return
		}
	}

	// Query for selection
	searchRequest := qualitygates.SearchRequest{
		GateId:       plan.GateId.ValueString(),
		Organization: r.p.organization,
	}

	res, err := r.p.client.Qualitygates.Search(searchRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read Quality Gate Selection",
			fmt.Sprintf("The Search request returned an error: %+v", err),
		)
		return
	}

	if result, ok := findSelection(res, projectKeys); ok {
		result.GateId = types.StringValue(plan.GateId.ValueString())
		result.ID = types.StringValue(plan.GateId.ValueString())
		diags = resp.State.Set(ctx, result)
		resp.Diagnostics.Append(diags...)
	} else {
		resp.Diagnostics.AddError(
			"Could not find Quality Gate Selection",
			fmt.Sprintf("The findSelection function was unable to find the project keys: %+v in the response: %+v", projectKeys, res),
		)
		return
	}
}

func (r *resourceQualityGateSelection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Selection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract project keys from Set
	var stateKeys []string
	diags = state.ProjectKeys.ElementsAs(ctx, &stateKeys, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	searchRequest := qualitygates.SearchRequest{
		GateId:       state.GateId.ValueString(),
		Organization: r.p.organization,
	}
	res, err := r.p.client.Qualitygates.Search(searchRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not Read the Quality Gate Selection",
			fmt.Sprintf("The Search request returned an error: %+v", err),
		)
		return
	}
	if result, ok := findSelection(res, stateKeys); ok {
		result.GateId = types.StringValue(state.GateId.ValueString())
		result.ID = types.StringValue(state.GateId.ValueString())
		diags = resp.State.Set(ctx, result)
		resp.Diagnostics.Append(diags...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (r *resourceQualityGateSelection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state Selection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan Selection
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sel, rem, diags := diffSelection(ctx, state, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, s := range rem {
		deselectRequest := qualitygates.DeselectRequest{
			Organization: r.p.organization,
			ProjectKey:   s,
		}
		err := r.p.client.Qualitygates.Deselect(deselectRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not Deselect the Quality Gate selection",
				fmt.Sprintf("The Deselect request returned an error: %+v", err),
			)
			return
		}
	}
	for _, s := range sel {
		selectRequest := qualitygates.SelectRequest{
			GateId:       state.GateId.ValueString(),
			Organization: r.p.organization,
			ProjectKey:   s,
		}
		err := r.p.client.Qualitygates.Select(selectRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not Select the Quality Gate selection",
				fmt.Sprintf("The Select request returned an error: %+v", err),
			)
			return
		}
	}

	// Extract project keys from plan for search
	var planKeys []string
	diags = plan.ProjectKeys.ElementsAs(ctx, &planKeys, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := qualitygates.SearchRequest{
		GateId:       plan.GateId.ValueString(),
		Organization: r.p.organization,
	}
	res, err := r.p.client.Qualitygates.Search(request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not Read the Quality Gate Selection",
			fmt.Sprintf("The Search request returned an error: %+v", err),
		)
		return
	}
	if result, ok := findSelection(res, planKeys); ok {
		result.GateId = types.StringValue(state.GateId.ValueString())
		result.ID = types.StringValue(state.GateId.ValueString())
		diags = resp.State.Set(ctx, result)
		resp.Diagnostics.Append(diags...)
	} else {
		resp.Diagnostics.AddError(
			"Could not find Quality Gate Selection",
			fmt.Sprintf("The findSelection function was unable to find the project keys: %+v in the response: %+v", planKeys, res),
		)
		return
	}
}

func (r *resourceQualityGateSelection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Selection
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract project keys from Set
	var stateKeys []string
	diags = state.ProjectKeys.ElementsAs(ctx, &stateKeys, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, s := range stateKeys {
		request := qualitygates.DeselectRequest{
			Organization: r.p.organization,
			ProjectKey:   s,
		}
		err := r.p.client.Qualitygates.Deselect(request)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not Deselect the Quality Gate Selection",
				fmt.Sprintf("The Deselect request returned an error: %+v", err),
			)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func diffSelection(ctx context.Context, state, plan Selection) (sel, rem []string, diags diag.Diagnostics) {
	var stateKeys, planKeys []string
	diags = state.ProjectKeys.ElementsAs(ctx, &stateKeys, false)
	if diags.HasError() {
		return
	}
	diags = plan.ProjectKeys.ElementsAs(ctx, &planKeys, false)
	if diags.HasError() {
		return
	}

	for _, old := range stateKeys {
		if !containSelection(planKeys, old) {
			rem = append(rem, old)
		}
	}
	for _, new := range planKeys {
		if !containSelection(stateKeys, new) {
			sel = append(sel, new)
		}
	}

	return sel, rem, diags
}

// Check if a condition is contained in a condition list
func containSelection(list []string, item string) bool {
	for _, c := range list {
		if c == item {
			return true
		}
	}
	return false
}
