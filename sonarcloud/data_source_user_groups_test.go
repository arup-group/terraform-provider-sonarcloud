package sonarcloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUserGroups(t *testing.T) {
	numberOfDefaultGroups := "2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserGroupsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sonarcloud_user_groups.test_groups", "groups.#", numberOfDefaultGroups),
				),
			},
		},
	})
}

func testAccDataSourceUserGroupsConfig() string {
	return fmt.Sprintf(`
data "sonarcloud_user_groups" "test_groups" {}
`)
}
