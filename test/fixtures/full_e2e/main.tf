terraform {
  required_providers {
    sonarcloud = {
      source = "arup.com/platform/sonarcloud"
    }
  }
}

provider "sonarcloud" {}

variable "test_prefix" {
  description = "Unique prefix for test resources to avoid conflicts"
  type        = string
}

variable "test_user_login" {
  description = "Existing user login in the organization for group membership tests"
  type        = string
}

# ==============================================================================
# RESOURCES
# ==============================================================================

# 1. sonarcloud_project - Create a test project
resource "sonarcloud_project" "test" {
  name       = "${var.test_prefix}-terratest-project"
  key        = "${var.test_prefix}-terratest-project-key"
  visibility = "public"
}

# 2. sonarcloud_project_link - Add a link to the project
resource "sonarcloud_project_link" "test" {
  project_key = sonarcloud_project.test.key
  name        = "${var.test_prefix}-docs-link"
  url         = "https://example.com/docs"
}

# 3. sonarcloud_project_main_branch - Set the main branch name
resource "sonarcloud_project_main_branch" "test" {
  project_key = sonarcloud_project.test.key
  name        = "main"
}

# 4. sonarcloud_user_group - Create a user group
resource "sonarcloud_user_group" "test" {
  name        = "${var.test_prefix}-terratest-group"
  description = "Test group created by Terratest E2E"
}

# 5. sonarcloud_user_group_member - Add a member to the group
resource "sonarcloud_user_group_member" "test" {
  group = sonarcloud_user_group.test.name
  login = var.test_user_login
}

# 6. sonarcloud_quality_gate - Create a quality gate with conditions
resource "sonarcloud_quality_gate" "test" {
  name       = "${var.test_prefix}-terratest-gate"
  is_default = false
  conditions = [
    {
      metric = "coverage"
      op     = "LT"
      error  = "80"
    },
    {
      metric = "duplicated_lines_density"
      op     = "GT"
      error  = "3"
    }
  ]
}

# 7. sonarcloud_quality_gate_selection - Link quality gate to project
resource "sonarcloud_quality_gate_selection" "test" {
  gate_id     = sonarcloud_quality_gate.test.gate_id
  project_keys = [sonarcloud_project.test.key]
}

# 8. sonarcloud_user_permissions - Set user permissions on project
resource "sonarcloud_user_permissions" "test" {
  project_key = sonarcloud_project.test.key
  login       = var.test_user_login
  permissions = ["user", "codeviewer"]
}

# 9. sonarcloud_user_group_permissions - Set group permissions on project
resource "sonarcloud_user_group_permissions" "test" {
  name        = sonarcloud_user_group.test.name
  project_key = sonarcloud_project.test.key
  permissions = ["user", "codeviewer"]
}

# 10. sonarcloud_webhook - Create a webhook for the project
resource "sonarcloud_webhook" "test" {
  name   = "${var.test_prefix}-terratest-webhook"
  url    = "https://example.com/webhook"
  project = sonarcloud_project.test.key
}

# 11. sonarcloud_user_token - Create a user token
# Note: Token value is write-only and cannot be read back after creation
resource "sonarcloud_user_token" "test" {
  name  = "${var.test_prefix}-terratest-token"
  login = var.test_user_login
}

# ==============================================================================
# DATA SOURCES
# ==============================================================================

# 1. sonarcloud_projects - List all projects
data "sonarcloud_projects" "all" {
  depends_on = [sonarcloud_project.test]
}

# 2. sonarcloud_project_links - Get links for the test project
data "sonarcloud_project_links" "test" {
  project_key = sonarcloud_project.test.key
  depends_on  = [sonarcloud_project_link.test]
}

# 3. sonarcloud_user_group - Get the test group by name
data "sonarcloud_user_group" "test" {
  name       = sonarcloud_user_group.test.name
  depends_on = [sonarcloud_user_group.test]
}

# 4. sonarcloud_user_groups - List all user groups
data "sonarcloud_user_groups" "all" {
  depends_on = [sonarcloud_user_group.test]
}

# 5. sonarcloud_user_group_members - Get members of the test group
data "sonarcloud_user_group_members" "test" {
  group       = sonarcloud_user_group.test.name
  depends_on  = [sonarcloud_user_group_member.test]
}

# 6. sonarcloud_user_group_permissions - Get group permissions on project
data "sonarcloud_user_group_permissions" "test" {
  project_key = sonarcloud_project.test.key
  depends_on  = [sonarcloud_user_group_permissions.test]
}

# 7. sonarcloud_user_permissions - Get user permissions on project
data "sonarcloud_user_permissions" "test" {
  project_key = sonarcloud_project.test.key
  depends_on  = [sonarcloud_user_permissions.test]
}

# 8. sonarcloud_quality_gate - Get the test quality gate by name
data "sonarcloud_quality_gate" "test" {
  name       = sonarcloud_quality_gate.test.name
  depends_on = [sonarcloud_quality_gate.test]
}

# 9. sonarcloud_quality_gates - List all quality gates
data "sonarcloud_quality_gates" "all" {
  depends_on = [sonarcloud_quality_gate.test]
}

# 10. sonarcloud_webhooks - Get webhooks for the test project
data "sonarcloud_webhooks" "test" {
  project = sonarcloud_project.test.key
  depends_on  = [sonarcloud_webhook.test]
}

# ==============================================================================
# OUTPUTS
# ==============================================================================

# Resource outputs
output "project_key" {
  value = sonarcloud_project.test.key
}

output "project_name" {
  value = sonarcloud_project.test.name
}

output "project_link_id" {
  value = sonarcloud_project_link.test.id
}

output "project_main_branch_name" {
  value = sonarcloud_project_main_branch.test.name
}

output "user_group_name" {
  value = sonarcloud_user_group.test.name
}

output "user_group_id" {
  value = sonarcloud_user_group.test.id
}

output "quality_gate_name" {
  value = sonarcloud_quality_gate.test.name
}

output "quality_gate_id" {
  value = sonarcloud_quality_gate.test.gate_id
}

output "webhook_name" {
  value = sonarcloud_webhook.test.name
}

output "user_token_name" {
  value = sonarcloud_user_token.test.name
}

# Data source outputs for validation
output "data_projects_count" {
  value = length(data.sonarcloud_projects.all.projects)
}

output "data_project_links_count" {
  value = length(data.sonarcloud_project_links.test.links)
}

output "data_user_group_name" {
  value = data.sonarcloud_user_group.test.name
}

output "data_user_groups_count" {
  value = length(data.sonarcloud_user_groups.all.groups)
}

output "data_user_group_members_count" {
  value = length(data.sonarcloud_user_group_members.test.users)
}

output "data_quality_gate_name" {
  value = data.sonarcloud_quality_gate.test.name
}

output "data_quality_gate_conditions_count" {
  value = length(data.sonarcloud_quality_gate.test.conditions)
}

output "data_quality_gates_count" {
  value = length(data.sonarcloud_quality_gates.all.quality_gates)
}

output "data_webhooks_count" {
  value = length(data.sonarcloud_webhooks.test.webhooks)
}
