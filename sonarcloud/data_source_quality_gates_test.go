package sonarcloud

func testAccDataSourceQualityGatesConfig() string {
	return `
data "sonarcloud_quality_gates" "test_quality_gates" {}
`
}
