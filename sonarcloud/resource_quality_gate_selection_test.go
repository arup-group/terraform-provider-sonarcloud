package sonarcloud

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccPreCheckQualityGateSelection(t *testing.T) {
	if v := os.Getenv("SONARCLOUD_PROJECT_KEY"); v == "" {
		t.Fatal("SONARCLOUD_PROJECT_KEY must be set for acceptance tests")
	}
}

func TestAccResourceQualityGateSelection(t *testing.T) {
	projectKey := os.Getenv("SONARCLOUD_PROJECT_KEY")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccPreCheckQualityGateSelection(t) },
		ProtoV6ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQualityGateSelectionConfig(projectKey),
				Check: resource.ComposeTestCheckFunc(
					// compare the gate_id between the created quality gate resource and the selection resource
					resource.TestCheckResourceAttrPair("sonarcloud_quality_gate_selection.test", "gate_id", "sonarcloud_quality_gate.test", "gate_id"),
					resource.TestCheckResourceAttr("sonarcloud_quality_gate_selection.test", "project_keys.0", projectKey),
				),
			},
		},
		CheckDestroy: testAccQualityGateSelectionDestroy,
	})
}

func testAccQualityGateSelectionDestroy(_ *terraform.State) error {
	return nil
}

func testAccQualityGateSelectionConfig(projectKey string) string {
	name := fmt.Sprintf("tf-acceptance-qg-%d", time.Now().Unix())
	return fmt.Sprintf(`
resource "sonarcloud_quality_gate" "test" {
	name = "%s"
}

resource "sonarcloud_quality_gate_selection" "test" {
	gate_id = sonarcloud_quality_gate.test.gate_id
	project_keys = ["%s"]
}
	`, name, projectKey)
}
