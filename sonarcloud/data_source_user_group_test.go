package sonarcloud

import (
	"fmt"
)

func testAccDataSourceUserGroupConfig() string {
	return fmt.Sprintf(`
data "sonarcloud_user_group" "test_group" {
	name = "TEST_DONT_REMOVE"
}
`)
}
