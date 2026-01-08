# Issue: testAccUserTokenDestroy function does not perform proper verification

## Description

The `testAccUserTokenDestroy` function in `sonarcloud/resource_user_token_test.go` currently always returns `nil` without performing any verification. This means the test cannot detect if resources fail to be properly cleaned up after the test completes.

## Current Implementation

```go
func testAccUserTokenDestroy(s *terraform.State) error {
	return nil
}
```

**Location:** `sonarcloud/resource_user_token_test.go:39-41`

## Expected Behavior

The destroy check function should verify that user token resources no longer exist in the Terraform state after the test completes. It should iterate through the state resources and return an error if any `sonarcloud_user_token` resources are still present.

## Suggested Fix

```go
func testAccUserTokenDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "sonarcloud_user_token" {
			return fmt.Errorf("user token resource %s still exists in state", rs.Primary.ID)
		}
	}
	return nil
}
```

## Impact

- **Severity:** Medium
- **Type:** Test Quality
- The current implementation reduces the effectiveness of acceptance tests by not validating proper resource cleanup
- Could mask bugs related to resource deletion or state management

## Related

- Original PR: #32
- Review comment: https://github.com/arup-group/terraform-provider-sonarcloud/pull/32#discussion_r2672177492
- Commit: ce52fb77de8b2080e1af62400e96504a331f1a3c

## Labels

- `testing`
- `enhancement`
- `good first issue`
