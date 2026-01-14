package sonarcloud

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testProjectLinkURL = "https://example.com"

func testAccPreCheckDataSourceProjectLinks(t *testing.T) {
	t.Helper()
	if v := os.Getenv("SONARCLOUD_PROJECT_KEY"); v == "" {
		t.Fatal("SONARCLOUD_PROJECT_KEY must be set for acceptance tests")
	}
}

func TestAccDataSourceProjectLinks(t *testing.T) {
	project := os.Getenv("SONARCLOUD_PROJECT_KEY")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccPreCheckDataSourceProjectLinks(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceProjectLinksConfigWithLink(project, testProjectLinkURL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sonarcloud_project_links.test", "links.#", "1"),
					resource.TestCheckResourceAttrSet("data.sonarcloud_project_links.test", "links.0.id"),
					resource.TestCheckResourceAttr("data.sonarcloud_project_links.test", "links.0.name", "Terraform Test Link"),
					resource.TestCheckResourceAttr("data.sonarcloud_project_links.test", "links.0.url", testProjectLinkURL),
				),
			},
		},
	})
}

func testAccDataSourceProjectLinksConfigWithLink(projectKey, linkURL string) string {
	return fmt.Sprintf(`
resource "sonarcloud_project_link" "test" {
  project_key = "%[1]s"
  name        = "Terraform Test Link"
  url         = "%[2]s"
}

data "sonarcloud_project_links" "test" {
	project_key = "%[1]s"
	depends_on = [sonarcloud_project_link.test]
}
`, projectKey, linkURL)
}
