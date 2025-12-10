package sonarcloud

import (
	"fmt"
)

func testAccDataSourceQualityGatesConfig() string {
	return fmt.Sprintf(`
data "sonarcloud_quality_gates" "test_quality_gates" {}
`)
}
