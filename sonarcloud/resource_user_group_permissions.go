package sonarcloud

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reinoudk/go-sonarcloud/sonarcloud"
	"github.com/reinoudk/go-sonarcloud/sonarcloud/permissions"
	"strings"
	"sync"
	"time"
)

type resourceUserGroupPermissions struct {
	p *sonarcloudProvider
}

var _ resource.Resource = &resourceUserGroupPermissions{}

func (r *resourceUserGroupPermissions) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group_permissions"
}

func (r *resourceUserGroupPermissions) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// TODO: Manually convert old schema
	}
}


func (r *resourceUserGroupPermissions) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unkown value from another resource. "+
				"This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan UserGroupPermissions
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add permissions one by one
	wg := sync.WaitGroup{}
	for _, elem := range plan.Permissions.Elems {
		permission := elem.(types.String).Value

		go func() {
			wg.Add(1)
			defer wg.Done()

			request := permissions.AddGroupRequest{
				GroupName:    plan.Name.ValueString(),
				Permission:   permission,
				ProjectKey:   plan.ProjectKey.ValueString(),
				Organization: r.p.organization,
			}
			if err := r.p.client.Permissions.AddGroup(request); err != nil {
				resp.Diagnostics.AddError(
					"Could not add group permissions",
					fmt.Sprintf("The AddGroup request returned an error: %+v", err),
				)
				return
			}
		}()
	}

	// Async wait for all requests to be done
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	// Set ID on success and return error diag on timeout
	select {
	case <-c:
	case <-time.After(30 * time.Second):
		resp.Diagnostics.AddError("Could not set user group permissions",
			"The requests to set the permissions timed out.",
		)
	}

	plannedPermissions := make([]string, len(plan.Permissions.Elems))
	diags = plan.Permissions.ElementsAs(ctx, &plannedPermissions, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backoffConfig := defaultBackoffConfig()

	group, err := backoff.RetryWithData(
		func() (*UserGroupPermissions, error) {
			group, err := findUserGroupWithPermissionsSet(r.p.client, plan.Name.ValueString(), plan.ProjectKey.ValueString(), plan.Permissions)
			return group, err
		}, backoffConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Could not find the user group with the planned permissions",
			fmt.Sprintf("The findUserGroupWithPermissionsSet call returned an error: %+v ", err),
		)
	} else {
		diags = resp.State.Set(ctx, group)
		resp.Diagnostics.Append(diags...)
	}
}

func (r *resourceUserGroupPermissions) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserGroupPermissions
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query for permissions
	searchRequest := UserGroupPermissionsSearchRequest{ProjectKey: state.ProjectKey.Value}
	groups, err := sonarcloud.GetAll[UserGroupPermissionsSearchRequest, UserGroupPermissionsSearchResponseGroup](r.p.client, "/permissions/groups", searchRequest, "groups")
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not get user group permissions",
			fmt.Sprintf("The request returned an error: %+v", err),
		)
		return
	}

	if group, ok := findUserGroup(groups, state.Name.ValueString()); ok {
		permissionsElems := make([]attr.ValueString(), len(group.Permissions))

		for i, permission := range group.Permissions {
			permissionsElems[i] = types.StringValue(permission)
		}

		result := UserGroupPermissions{
			ID:          types.StringValue(group.Id),
			ProjectKey:  state.ProjectKey,
			Name:        types.StringValue(group.Name),
			Description: types.StringValue(group.Description),
			Permissions: types.Set{Elems: permissionsElems, ElemType: types.StringType},
		}
		diags = resp.State.Set(ctx, result)
		resp.Diagnostics.Append(diags...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (r *resourceUserGroupPermissions) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state UserGroupPermissions
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan UserGroupPermissions
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	toAdd, toRemove := diffAttrSets(state.Permissions, plan.Permissions)

	for _, remove := range toRemove {
		removeRequest := permissions.RemoveGroupRequest{
			GroupName:    state.Name.ValueString(),
			Permission:   remove.(types.String).Value,
			ProjectKey:   state.ProjectKey.ValueString(),
			Organization: r.p.organization,
		}
		err := r.p.client.Permissions.RemoveGroup(removeRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not remove the user group permission",
				fmt.Sprintf("The RemoveGroup request returned an error: %+v", err),
			)
			return
		}
	}
	for _, add := range toAdd {
		addRequest := permissions.AddGroupRequest{
			GroupName:    plan.Name.ValueString(),
			Permission:   add.(types.String).Value,
			ProjectKey:   plan.ProjectKey.ValueString(),
			Organization: r.p.organization,
		}
		if err := r.p.client.Permissions.AddGroup(addRequest); err != nil {
			resp.Diagnostics.AddError(
				"Could not add the user group permission",
				fmt.Sprintf("The AddGroup request returned an error: %+v", err),
			)
			return
		}
	}

	plannedPermissions := make([]string, len(plan.Permissions.Elems))
	diags = plan.Permissions.ElementsAs(ctx, &plannedPermissions, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backoffConfig := defaultBackoffConfig()

	group, err := backoff.RetryWithData(
		func() (*UserGroupPermissions, error) {
			group, err := findUserGroupWithPermissionsSet(r.p.client, plan.Name.ValueString(), plan.ProjectKey.ValueString(), plan.Permissions)
			return group, err
		}, backoffConfig)

	if err != nil {
		resp.Diagnostics.AddError(
			"Could not find the user group with the planned permissions",
			fmt.Sprintf("The findUserGroupWithPermissionsSet call returned an error: %+v ", err),
		)
	} else {
		diags = resp.State.Set(ctx, group)
		resp.Diagnostics.Append(diags...)
	}
}

func (r *resourceUserGroupPermissions) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserGroupPermissions
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, remove := range state.Permissions.Elems {
		removeRequest := permissions.RemoveGroupRequest{
			GroupName:    state.Name.ValueString(),
			Permission:   remove.(types.String).Value,
			ProjectKey:   state.ProjectKey.ValueString(),
			Organization: r.p.organization,
		}
		err := r.p.client.Permissions.RemoveGroup(removeRequest)
		if err != nil {
			resp.Diagnostics.AddError(
				"Could not remove the user group permission",
				fmt.Sprintf("The RemoveGroup request returned an error: %+v", err),
			)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceUserGroupPermissions) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) < 1 || len(idParts) > 2 || idParts[0] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name OR name,project_key. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_key"), idParts[1])...)
	}
}

type UserGroupPermissionsSearchRequest struct {
	ProjectKey string
}

type UserGroupPermissionsSearchResponseGroup struct {
	Id          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// findUserGroupWithPermissionsSet tries to find a user group with the given name and the expected permissions
func findUserGroupWithPermissionsSet(client *sonarcloud.Client, name, projectKey string, expectedPermissions types.Set) (*UserGroupPermissions, error) {
	searchRequest := UserGroupPermissionsSearchRequest{ProjectKey: projectKey}
	groups, err := sonarcloud.GetAll[UserGroupPermissionsSearchRequest, UserGroupPermissionsSearchResponseGroup](client, "/permissions/groups", searchRequest, "groups")
	if err != nil {
		return nil, err
	}

	group, ok := findUserGroup(groups, name)
	if !ok {
		return nil, fmt.Errorf("group not found in response (name='%s',projectKey='%s')", name, projectKey)
	}

	permissionsElems := make([]attr.ValueString(), len(group.Permissions))
	for i, permission := range group.Permissions {
		permissionsElems[i] = types.StringValue(permission)
	}

	foundPermissions := types.Set{Elems: permissionsElems, ElemType: types.StringType}

	if !foundPermissions.Equal(expectedPermissions) {
		return nil, fmt.Errorf("the returned permissions do not match the expected permissions (name='%s',projectKey='%s, expected='%v', got='%v')",
			name,
			projectKey,
			expectedPermissions,
			foundPermissions)
	}

	return &UserGroupPermissions{
		ID:          types.StringValue(projectKey + "-" + name),
		ProjectKey:  types.StringValue(projectKey),
		Name:        types.StringValue(group.Name),
		Description: types.StringValue(group.Description),
		Permissions: foundPermissions,
	}, nil
}

// findUserGroup returns the user group with the given name, if it exists
func findUserGroup(groups []UserGroupPermissionsSearchResponseGroup, name string) (*UserGroupPermissionsSearchResponseGroup, bool) {
	for _, group := range groups {
		if group.Name == name {
			return &group, true
		}
	}
	return nil, false
}
