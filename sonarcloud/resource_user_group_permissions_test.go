package sonarcloud

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"strings"
	"testing"
)

func TestAccUserGroupPermissions(t *testing.T) {
	projectKey := os.Getenv("SONARCLOUD_PROJECT_KEY")
	name := os.Getenv("SONARCLOUD_TEST_GROUP_NAME")

	// Possible values for global permissions: admin, profileadmin, gateadmin, scan, provisioning
	// Possible values for project permissions admin, codeviewer, issueadmin, securityhotspotadmin, scan, user
	// Note: some permissions (like codeviewer) are active by default on public projects, and are not returned when reading
	// these should not be used in tests when using a public test project
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig("", name, []string{
					"provisioning",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "project_key", ""),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "name", name),
					resource.TestCheckResourceAttrSet("sonarcloud_user_group_permissions.test_permission", "description"),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.0", "provisioning"),
				),
			},
			userGroupPermissionsImportCheck("sonarcloud_user_group_permissions.test_permission", name, ""),
			{
				Config: testAccPermissionConfig("", name, []string{
					"provisioning",
					"scan",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "project_key", ""),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "name", name),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.0", "provisioning"),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.1", "scan"),
				),
			},
			userGroupPermissionsImportCheck("sonarcloud_user_group_permissions.test_permission", name, ""),
			{
				Config: testAccPermissionConfig("", name, []string{
					"scan",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "project_key", ""),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "name", name),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.0", "scan"),
				),
			},
			userGroupPermissionsImportCheck("sonarcloud_user_group_permissions.test_permission", name, ""),
			{
				Config: testAccPermissionConfig(projectKey, name, []string{
					"admin",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "project_key", projectKey),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "name", name),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.0", "admin"),
				),
			},
			userGroupPermissionsImportCheck("sonarcloud_user_group_permissions.test_permission", name, projectKey),
			{
				Config: testAccPermissionConfig(projectKey, name, []string{
					"issueadmin",
					"scan",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "project_key", projectKey),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "name", name),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.0", "issueadmin"),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.1", "scan"),
				),
			},
			userGroupPermissionsImportCheck("sonarcloud_user_group_permissions.test_permission", name, projectKey),
			{
				Config: testAccPermissionConfig(projectKey, name, []string{
					"scan",
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "project_key", projectKey),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "name", name),
					resource.TestCheckResourceAttr("sonarcloud_user_group_permissions.test_permission", "permissions.0", "scan"),
				),
			},
			userGroupPermissionsImportCheck("sonarcloud_user_group_permissions.test_permission", name, projectKey),
		},
		CheckDestroy: testAccPermissionDestroy,
	})
}

func testAccPermissionDestroy(s *terraform.State) error {
	return nil
}

func testAccPermissionConfig(project string, name string, permissions []string) string {
	result := fmt.Sprintf(`
resource "sonarcloud_user_group_permissions" "test_permission" {
	project_key = "%s"
	name = "%s"
	permissions = %s
}
`, project, name, permissionsListString(permissions))
	return result
}

func permissionsListString(permissions []string) string {
	return fmt.Sprintf(`["%s"]`, strings.Join(permissions, `","`))
}

func userGroupPermissionsImportCheck(resourceName, name, projectKey string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      resourceName,
		ImportState:       true,
		ImportStateId:     fmt.Sprintf("%s,%s", name, projectKey),
		ImportStateVerify: true,
	}
}
