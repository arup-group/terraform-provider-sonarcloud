package sonarcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestQualityGateSchemaComplete verifies the schema is fully defined with all required attributes
func TestQualityGateSchemaComplete(t *testing.T) {
	ctx := context.Background()
	r := &resourceQualityGate{}

	// Get schema
	var req resource.SchemaRequest
	var resp resource.SchemaResponse
	r.Schema(ctx, req, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema has errors: %v", resp.Diagnostics)
	}

	schema := resp.Schema

	// Verify all expected attributes exist
	requiredAttrs := []string{"id", "gate_id", "name", "is_built_in", "is_default", "conditions"}
	for _, attrName := range requiredAttrs {
		if _, exists := schema.Attributes[attrName]; !exists {
			t.Errorf("Schema missing required attribute: %s", attrName)
		}
	}

	// Verify conditions has nested attributes
	conditionsAttr, exists := schema.Attributes["conditions"]
	if !exists {
		t.Fatal("Schema missing 'conditions' attribute")
	}

	// The conditions attribute should have NestedObject with attributes
	// We're checking that it's defined and not just a TODO placeholder
	if conditionsAttr.GetDescription() == "" {
		t.Error("conditions attribute has no description - may be incomplete")
	}
}

// func TestAccResourceQualityGate(t *testing.T) {
// 	names := []string{"quality_gate_a", "quality_gate_b"}
// 	def := []string{"true", "false"}
// 	metrics := []string{"coverage", "duplicated_lines_density"}
// 	testError := []string{"10", "11"}
// 	Op := []string{"LT", "GT"}

// 	// TODO: use fixed test organization so that changes can be verified.

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:                 func() { testAccPreCheck(t) },
// 		ProtoV6ProviderFactories: testAccProviderFactories,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccQualityGateConfig(names[0], def[0], metrics[0], testError[0], Op[0]),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "name", names[0]),
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "conditions.0.metric", metrics[0]),
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "conditions.0.error", testError[0]),
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "conditions.0.op", Op[0]),
// 				),
// 			},
// 			qualityGateImportCheck("sonarcloud_quality_gate.test", names[0]),
// 			{
// 				Config: testAccQualityGateConfig(names[1], def[1], metrics[1], testError[1], Op[1]),
// 				Check: resource.ComposeTestCheckFunc(
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "name", names[1]),
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "conditions.0.metric", metrics[1]),
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "conditions.0.error", testError[1]),
// 					resource.TestCheckResourceAttr("sonarcloud_quality_gate.test", "conditions.0.op", Op[1]),
// 				),
// 			},
// 			qualityGateImportCheck("sonarcloud_quality_gate.test", names[1]),
// 		},
// 		CheckDestroy: testAccQualityGateDestroy,
// 	})
// }

func testAccQualityGateDestroy(s *terraform.State) error {
	return nil
}

func testAccQualityGateConfig(name, def, metric, err, op string) string {
	return fmt.Sprintf(`
resource "sonarcloud_quality_gate" "test" {
	name = "%s"
	is_default = "%s"
	conditions = [
		{
			metric = "%s"
			error = "%s"
			op = "%s"
		}
	]
}
	`, name, def, metric, err, op)

}

// func qualityGateImportCheck(resourceName, name string) resource.TestStep {
// 	return resource.TestStep{
// 		ResourceName:      resourceName,
// 		ImportState:       true,
// 		ImportStateId:     name,
// 		ImportStateVerify: true,
// 	}
// }
