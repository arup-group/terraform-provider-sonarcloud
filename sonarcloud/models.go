package sonarcloud

import "github.com/hashicorp/terraform-plugin-framework/types"

// Groups represents a collection of SonarCloud groups.
type Groups struct {
	ID     types.String `tfsdk:"id"`
	Groups []Group      `tfsdk:"groups"`
}

// Group represents a single SonarCloud group with its properties.
type Group struct {
	ID           types.String `tfsdk:"id"`
	Default      types.Bool   `tfsdk:"default"`
	Description  types.String `tfsdk:"description"`
	MembersCount types.Number `tfsdk:"members_count"`
	Name         types.String `tfsdk:"name"`
}

// GroupMember represents a member of a SonarCloud group.
type GroupMember struct {
	ID    types.String `tfsdk:"id"`
	Group types.String `tfsdk:"group"`
	Login types.String `tfsdk:"login"`
}

// User represents a SonarCloud user.
type User struct {
	Login types.String `tfsdk:"login"`
	Name  types.String `tfsdk:"name"`
}

// Users represents a collection of SonarCloud users.
type Users struct {
	ID    types.String `tfsdk:"id"`
	Group types.String `tfsdk:"group"`
	Users []User       `tfsdk:"users"`
}

// Token represents a SonarCloud user token.
type Token struct {
	ID    types.String `tfsdk:"id"`
	Login types.String `tfsdk:"login"`
	Name  types.String `tfsdk:"name"`
	Token types.String `tfsdk:"token"`
}

// Projects represents a collection of SonarCloud projects.
type Projects struct {
	ID       types.String `tfsdk:"id"`
	Projects []Project    `tfsdk:"projects"`
}

// Project represents a single SonarCloud project.
type Project struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Key        types.String `tfsdk:"key"`
	Visibility types.String `tfsdk:"visibility"`
}

// ProjectMainBranch represents the main branch configuration for a project.
type ProjectMainBranch struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ProjectKey types.String `tfsdk:"project_key"`
}

// Condition represents a quality gate condition.
type Condition struct {
	Error  types.String  `tfsdk:"error"`
	ID     types.Float64 `tfsdk:"id"`
	Metric types.String  `tfsdk:"metric"`
	Op     types.String  `tfsdk:"op"`
}

// Conditions represents a collection of quality gate conditions.
type Conditions struct {
	ID         types.Float64 `tfsdk:"id"`
	Conditions []Condition   `tfsdk:"condition"`
}

// QualityGate represents a SonarCloud quality gate with its conditions.
type QualityGate struct {
	ID         types.String  `tfsdk:"id"`
	GateId     types.Float64 `tfsdk:"gate_id"` //nolint:revive // Field name matches Terraform schema
	Conditions []Condition   `tfsdk:"conditions"`
	IsBuiltIn  types.Bool    `tfsdk:"is_built_in"`
	IsDefault  types.Bool    `tfsdk:"is_default"`
	Name       types.String  `tfsdk:"name"`
}

// QualityGates represents a collection of quality gates.
type QualityGates struct {
	ID           types.String  `tfsdk:"id"`
	QualityGates []QualityGate `tfsdk:"quality_gates"`
}

// Selection represents a quality gate selection for projects.
type Selection struct {
	ID          types.String `tfsdk:"id"`
	GateId      types.String `tfsdk:"gate_id"` //nolint:revive // Field name matches Terraform schema
	ProjectKeys types.Set    `tfsdk:"project_keys"`
}

// DataUserGroupPermissionsGroup represents group permissions data.
type DataUserGroupPermissionsGroup struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
}

// DataUserGroupPermissions represents a collection of group permissions.
type DataUserGroupPermissions struct {
	ID         types.String                    `tfsdk:"id"`
	ProjectKey types.String                    `tfsdk:"project_key"`
	Groups     []DataUserGroupPermissionsGroup `tfsdk:"groups"`
}

// UserGroupPermissions represents permissions for a specific user group.
type UserGroupPermissions struct {
	ID          types.String `tfsdk:"id"`
	ProjectKey  types.String `tfsdk:"project_key"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
}

// DataUserPermissionsUser represents user permissions data.
type DataUserPermissionsUser struct {
	Login       types.String `tfsdk:"login"`
	Name        types.String `tfsdk:"name"`
	Permissions types.Set    `tfsdk:"permissions"`
	Avatar      types.String `tfsdk:"avatar"`
}

// DataUserPermissions represents a collection of user permissions.
type DataUserPermissions struct {
	ID         types.String              `tfsdk:"id"`
	ProjectKey types.String              `tfsdk:"project_key"`
	Users      []DataUserPermissionsUser `tfsdk:"users"`
}

// UserPermissions represents permissions for a specific user.
type UserPermissions struct {
	ID          types.String `tfsdk:"id"`
	ProjectKey  types.String `tfsdk:"project_key"`
	Login       types.String `tfsdk:"login"`
	Name        types.String `tfsdk:"name"`
	Permissions types.Set    `tfsdk:"permissions"`
	Avatar      types.String `tfsdk:"avatar"`
}

// DataProjectLinks represents a collection of project links.
type DataProjectLinks struct {
	ID         types.String      `tfsdk:"id"`
	ProjectKey types.String      `tfsdk:"project_key"`
	Links      []DataProjectLink `tfsdk:"links"`
}

// DataProjectLink represents a single project link data.
type DataProjectLink struct {
	Id   types.String `tfsdk:"id"` //nolint:revive // Field name matches Terraform schema
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	Url  types.String `tfsdk:"url"` //nolint:revive // Field name matches Terraform schema
}

// ProjectLink represents a project link resource.
type ProjectLink struct {
	ID         types.String `tfsdk:"id"`
	ProjectKey types.String `tfsdk:"project_key"`
	Name       types.String `tfsdk:"name"`
	Url        types.String `tfsdk:"url"` //nolint:revive // Field name matches Terraform schema
}

// DataWebhooks represents a collection of webhooks.
type DataWebhooks struct {
	ID       types.String  `tfsdk:"id"`
	Project  types.String  `tfsdk:"project"`
	Webhooks []DataWebhook `tfsdk:"webhooks"`
}

// DataWebhook represents a single webhook data.
type DataWebhook struct {
	Key       types.String `tfsdk:"key"`
	Name      types.String `tfsdk:"name"`
	HasSecret types.Bool   `tfsdk:"has_secret"`
	Url       types.String `tfsdk:"url"` //nolint:revive // Field name matches Terraform schema
}

// Webhook represents a webhook resource.
type Webhook struct {
	ID      types.String `tfsdk:"id"`
	Key     types.String `tfsdk:"key"`
	Project types.String `tfsdk:"project"`
	Name    types.String `tfsdk:"name"`
	Secret  types.String `tfsdk:"secret"`
	Url     types.String `tfsdk:"url"` //nolint:revive // Field name matches Terraform schema
}
