package sonarcloud

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testDataAccPreCheckQualityGate(t *testing.T) {
	if v := os.Getenv("SONARCLOUD_PROJECT_KEY"); v == "" {
		t.Fatal("SONARCLOUD_PROJECT_KEY must be set for acceptance tests")
	}
}

func TestAccDataSourceQualityGate(t *testing.T) {
	// create a unique name for the quality gate resource used by the test
	name := fmt.Sprintf("tf-acceptance-qg-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testDataAccPreCheckQualityGate(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceQualityGateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.sonarcloud_quality_gate.test_quality_gate", "name", name),
				),
			},
		},
	})
}

func testAccDataSourceQualityGateConfig(qualityGateName string) string {
	return fmt.Sprintf(`
resource "sonarcloud_quality_gate" "test_quality_gate" {
  name = "%s"
}

data "sonarcloud_quality_gate" "test_quality_gate" {
  name = sonarcloud_quality_gate.test_quality_gate.name
}
`, qualityGateName)
}
