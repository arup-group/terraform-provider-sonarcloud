// Package test contains end-to-end tests using Terratest
package test

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoRoot returns the absolute path to the repository root
func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Cannot determine repository root")
	}
	testDir := filepath.Dir(file)
	parent := filepath.Dir(testDir)
	absolutePath, err := filepath.Abs(parent)
	if err != nil {
		log.Fatal(err)
	}
	return absolutePath
}

// TestMain loads the .env file before running tests
func TestMain(m *testing.M) {
	// Load .env file from repo root before test functions run
	envFile := filepath.Join(repoRoot(), ".env")
	if _, err := os.Stat(envFile); err == nil {
		err := godotenv.Load(envFile)
		if err != nil {
			fmt.Printf("Could not load .env file %s from repo root: %v\n", envFile, err)
			os.Exit(2)
		}
	}
	os.Exit(m.Run())
}

// TestFullE2E runs a comprehensive end-to-end test that exercises all
// SonarCloud provider resources and data sources in a single integrated scenario.
//
// Required environment variables:
//   - SONARCLOUD_ORGANIZATION: The SonarCloud organization to use
//   - SONARCLOUD_TOKEN: API token with admin permissions
//   - SONARCLOUD_TEST_USER_LOGIN: An existing user login in the organization for group membership tests
func TestFullE2E(t *testing.T) {
	t.Parallel()

	// Validate required environment variables
	checkEnvVars(t)

	// Generate a unique prefix to avoid resource conflicts
	uniqueID := strings.ToLower(random.UniqueId())
	testPrefix := fmt.Sprintf("tt%s", uniqueID)

	// Get the path to the Terraform fixture
	fixturePath := getFixturePath(t, "full_e2e")

	// Get test user login from environment
	testUserLogin := os.Getenv("SONARCLOUD_TEST_USER_LOGIN")

	// Prepare a fixture-local plugins directory and install the built provider there
	pluginRelPath := filepath.Join("plugins", "arup.com", "platform", "sonarcloud", "0.1.0-local", fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH))
	pluginDir := filepath.Join(fixturePath, pluginRelPath)
	require.NoError(t, os.MkdirAll(pluginDir, 0o750)) // G301: Restrict dir perms

	// Find built provider binary in the repository dist/ directory
	repo := repoRoot()
	var builtBinary string

	// Try to find a local manual build first, as it guarantees the correct architecture
	manualBuildPath := filepath.Join(repo, fmt.Sprintf("terraform-provider-sonarcloud_%s", runtime.GOOS))
	if _, err := os.Stat(manualBuildPath); err == nil {
		builtBinary = manualBuildPath
	} else {
		// Fallback to searching goreleaser dist/ directory
		_ = filepath.Walk(filepath.Join(repo, "dist"), func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return err // nilerr: propagate error if not nil
			}
			base := filepath.Base(p)
			if strings.HasPrefix(base, "terraform-provider-sonarcloud") && strings.Contains(p, runtime.GOOS) && strings.Contains(p, runtime.GOARCH) {
				builtBinary = p
				return filepath.SkipDir
			}
			return nil
		})
	}

	require.NotEmpty(t, builtBinary, "Could not find built provider binary under dist/; run `make build` or `make install` first")

	// Copy it into the fixture plugin dir with the expected name using atomic rename to avoid "text file busy"
	targetPath := filepath.Join(pluginDir, "terraform-provider-sonarcloud")
	tmpPath := targetPath + ".tmp"
	src, err := os.Open(builtBinary) // G304: Path is controlled in test
	require.NoError(t, err)
	// Close src and check error

	defer func() {
		err := src.Close()
		if err != nil {
			t.Errorf("error closing src: %v", err)
		}
	}()
	dst, err := os.Create(tmpPath) // G304: Path is controlled in test
	require.NoError(t, err)
	_, err = io.Copy(dst, src)
	require.NoError(t, err)
	// ensure file is written and closed before rename
	require.NoError(t, dst.Sync())
	require.NoError(t, dst.Close())
	// make executable
	require.NoError(t, os.Chmod(tmpPath, 0o700)) // G302: Restrict file perms
	// atomic rename into final location
	require.NoError(t, os.Rename(tmpPath, targetPath))

	// Write a .terraformrc in the fixture that uses a relative path to the fixture-local plugin dir
	terraformrcPath := filepath.Join(fixturePath, ".terraformrc")
	relativePluginPath := "./" + filepath.ToSlash(pluginRelPath)
	terraformrcContent := fmt.Sprintf("provider_installation {\n  dev_overrides {\n    \"arup.com/platform/sonarcloud\" = \"%s\"\n  }\n  direct {}\n}\n", relativePluginPath)
	require.NoError(t, os.WriteFile(terraformrcPath, []byte(terraformrcContent), 0o600)) // G306: Restrict file perms
	// Remove .terraformrc and check error

	defer func() {
		err := os.Remove(terraformrcPath)
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("error removing .terraformrc: %v", err)
		}
	}()

	// Configure Terraform options
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: fixturePath,
		Vars: map[string]interface{}{
			"test_prefix":     testPrefix,
			"test_user_login": testUserLogin,
		},
		// Set environment variables for the provider and point the CLI config to the fixture .terraformrc
		EnvVars: map[string]string{
			"SONARCLOUD_ORGANIZATION": os.Getenv("SONARCLOUD_ORGANIZATION"),
			"SONARCLOUD_TOKEN":        os.Getenv("SONARCLOUD_TOKEN"),
			"TF_CLI_CONFIG_FILE":      terraformrcPath,
		},
		NoColor: true,
	})

	// Ensure cleanup on test completion
	defer terraform.Destroy(t, terraformOptions)

	// When using dev_overrides, terraform init is not needed and may error.
	// Just run terraform apply directly.
	terraform.Apply(t, terraformOptions)

	// =========================================================================
	// Validate Resource Outputs
	// =========================================================================

	t.Run("ValidateResources", func(t *testing.T) {
		// Project
		projectKey := terraform.Output(t, terraformOptions, "project_key")
		assert.Equal(t, fmt.Sprintf("%s-terratest-project-key", testPrefix), projectKey)

		projectName := terraform.Output(t, terraformOptions, "project_name")
		assert.Equal(t, fmt.Sprintf("%s-terratest-project", testPrefix), projectName)

		// Project Link
		projectLinkID := terraform.Output(t, terraformOptions, "project_link_id")
		assert.NotEmpty(t, projectLinkID, "Project link ID should not be empty")

		// Project Main Branch
		mainBranchName := terraform.Output(t, terraformOptions, "project_main_branch_name")
		assert.Equal(t, "main", mainBranchName)

		// User Group
		userGroupName := terraform.Output(t, terraformOptions, "user_group_name")
		assert.Equal(t, fmt.Sprintf("%s-terratest-group", testPrefix), userGroupName)

		userGroupID := terraform.Output(t, terraformOptions, "user_group_id")
		assert.NotEmpty(t, userGroupID, "User group ID should not be empty")

		// Quality Gate
		qualityGateName := terraform.Output(t, terraformOptions, "quality_gate_name")
		assert.Equal(t, fmt.Sprintf("%s-terratest-gate", testPrefix), qualityGateName)

		qualityGateID := terraform.Output(t, terraformOptions, "quality_gate_id")
		assert.NotEmpty(t, qualityGateID, "Quality gate ID should not be empty")

		// Webhook
		webhookName := terraform.Output(t, terraformOptions, "webhook_name")
		assert.Equal(t, fmt.Sprintf("%s-terratest-webhook", testPrefix), webhookName)

		// User Token
		userTokenName := terraform.Output(t, terraformOptions, "user_token_name")
		assert.Equal(t, fmt.Sprintf("%s-terratest-token", testPrefix), userTokenName)
	})

	// =========================================================================
	// Validate Data Source Outputs
	// =========================================================================

	t.Run("ValidateDataSources", func(t *testing.T) {
		// Projects data source - should contain at least our test project
		projectsCount := terraform.Output(t, terraformOptions, "data_projects_count")
		assert.NotEqual(t, "0", projectsCount, "Should have at least one project")

		// Project links data source - should have at least our test link
		projectLinksCount := terraform.Output(t, terraformOptions, "data_project_links_count")
		assert.NotEqual(t, "0", projectLinksCount, "Should have at least one project link")

		// User group data source - should match our created group
		dataUserGroupName := terraform.Output(t, terraformOptions, "data_user_group_name")
		assert.Equal(t, fmt.Sprintf("%s-terratest-group", testPrefix), dataUserGroupName)

		// User groups data source - should have at least our test group
		userGroupsCount := terraform.Output(t, terraformOptions, "data_user_groups_count")
		assert.NotEqual(t, "0", userGroupsCount, "Should have at least one user group")

		// User group members data source - should have at least our test member
		userGroupMembersCount := terraform.Output(t, terraformOptions, "data_user_group_members_count")
		assert.NotEqual(t, "0", userGroupMembersCount, "Should have at least one group member")

		// Quality gate data source - should match our created gate
		dataQualityGateName := terraform.Output(t, terraformOptions, "data_quality_gate_name")
		assert.Equal(t, fmt.Sprintf("%s-terratest-gate", testPrefix), dataQualityGateName)

		// Quality gate conditions - should have 2 conditions
		qualityGateConditionsCount := terraform.Output(t, terraformOptions, "data_quality_gate_conditions_count")
		assert.Equal(t, "2", qualityGateConditionsCount, "Quality gate should have 2 conditions")

		// Quality gates data source - should have at least our test gate
		qualityGatesCount := terraform.Output(t, terraformOptions, "data_quality_gates_count")
		assert.NotEqual(t, "0", qualityGatesCount, "Should have at least one quality gate")

		// Webhooks data source - should have at least our test webhook
		webhooksCount := terraform.Output(t, terraformOptions, "data_webhooks_count")
		assert.NotEqual(t, "0", webhooksCount, "Should have at least one webhook")
	})

	// =========================================================================
	// Validate Terraform Plan (Idempotency Check)
	// =========================================================================

	t.Run("ValidateIdempotency", func(t *testing.T) {
		// Run plan to verify no changes are detected (idempotency)
		planOutput := terraform.Plan(t, terraformOptions)

		// Check that no changes are planned
		assert.Contains(t, planOutput, "No changes",
			"Terraform plan should show no changes after apply (idempotency check)")
	})
}

// checkEnvVars validates that all required environment variables are set
func checkEnvVars(t *testing.T) {
	t.Helper()

	required := []string{
		"SONARCLOUD_ORGANIZATION",
		"SONARCLOUD_TOKEN",
		"SONARCLOUD_TEST_USER_LOGIN",
	}

	var missing []string
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		require.Fail(t, fmt.Sprintf(
			"Required environment variables not set: %s. "+
				"Please set these before running E2E tests.",
			strings.Join(missing, ", ")))
	}
}

// getFixturePath returns the absolute path to a test fixture directory
func getFixturePath(t *testing.T, fixtureName string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "Failed to get current file path")

	testDir := filepath.Dir(currentFile)
	fixturePath := filepath.Join(testDir, "fixtures", fixtureName)

	// Verify the fixture directory exists
	_, err := os.Stat(fixturePath)
	require.NoError(t, err, "Fixture directory does not exist: %s", fixturePath)

	return fixturePath
}
