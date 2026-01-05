package sonarcloud

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/joho/godotenv"
)

var testAccProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	testAccProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"sonarcloud": providerserver.NewProtocol6WithError(New()),
	}
}

func repoRoot(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Cannot determine repository root")
	}
	dir := filepath.Dir(file)        // directory containing this file
	parent := filepath.Dir(dir)      // one directory up
	abs, err := filepath.Abs(parent) // normalize to absolute path
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
func testAccPreCheck(t *testing.T) {
	// get config variables either from a `.env` file or process env
	envFile := filepath.Join(repoRoot(t), ".env")
	if _, err := os.Stat(envFile); err == nil {
		err := godotenv.Load(envFile)
		if err != nil {
			errMsg := "Could not load .env" + envFile + " file from repo root: " + err.Error()
			t.Error(errMsg)
			t.Fatal("Error loading .env file from repo root")
		}
	}
	if v := os.Getenv("SONARCLOUD_ORGANIZATION"); v == "" {
		t.Fatal("SONARCLOUD_ORGANIZATION must be set for acceptance tests")
	}
	if v := os.Getenv("SONARCLOUD_TOKEN"); v == "" {
		t.Fatal("SONARCLOUD_TOKEN must be set for acceptance tests")
	}
}
