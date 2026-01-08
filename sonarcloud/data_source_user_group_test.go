package sonarcloud

import (
	"fmt"
)

func testAccDataSourceUserGroupConfig() string {
	return `
data "sonarcloud_user_group" "test_group" {
	name = "TEST_DONT_REMOVE"
}
`
}
