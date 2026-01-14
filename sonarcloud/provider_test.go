package sonarcloud

import (
	"fmt"
	"log"
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

func repoRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Cannot determine repository root")
	}
	thisDir := filepath.Dir(file)
	parent := filepath.Dir(thisDir)
	absolutePath, err := filepath.Abs(parent)
	if err != nil {
		log.Fatal(err)
	}
	return absolutePath
}
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("SONARCLOUD_ORGANIZATION"); v == "" {
		t.Fatal("SONARCLOUD_ORGANIZATION must be set for acceptance tests")
	}
	if v := os.Getenv("SONARCLOUD_TOKEN"); v == "" {
		t.Fatal("SONARCLOUD_TOKEN must be set for acceptance tests")
	}
}
func TestMain(m *testing.M) {
	// make sure .env loading is tried before test functions are defined
	envFile := filepath.Join(repoRoot(), ".env")
	if _, err := os.Stat(envFile); err == nil {
		err := godotenv.Load(envFile)
		if err != nil {
			fmt.Printf("Could not load .env file %s from repo root: %v", envFile, err)
			// the file exists, but we can't load it, better fail early
			os.Exit(2)
		}
	}
	os.Exit(m.Run())
}
