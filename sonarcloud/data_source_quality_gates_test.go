package sonarcloud

//nolint:unused // Kept for future test expansion
func testAccDataSourceQualityGatesConfig() string {
	return `
data "sonarcloud_quality_gates" "test_quality_gates" {}
`
}
