package sonarcloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceProjects(t *testing.T) {
	numberOfDefaultProjects := "2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceProjectsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sonarcloud_projects.test_projects", "projects.#", numberOfDefaultProjects),
				),
			},
		},
	})
}

func testAccDataSourceProjectsConfig() string {
	return `
data "sonarcloud_projects" "test_projects" {}
`
}
