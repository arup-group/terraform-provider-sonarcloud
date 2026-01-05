# AGENTS.md

This document provides guidance for AI agents working on the `terraform-provider-sonarcloud` codebase.

## Project Overview

This is a **Terraform Provider** for managing SonarCloud resources (user groups, projects, quality gates, webhooks, permissions, etc.). It is written in Go and uses the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

### Tech Stack

- **Language:** Go 1.24+
- **Framework:** HashiCorp Terraform Plugin Framework
- **Build Tool:** GoReleaser
- **API Client:** `github.com/reinoudk/go-sonarcloud`

## Repository Structure

```
├── sonarcloud/           # Main provider code
│   ├── provider.go       # Provider definition and configuration
│   ├── resource_*.go     # Resource implementations (CRUD operations)
│   ├── data_source_*.go  # Data source implementations (read-only)
│   ├── *_test.go         # Acceptance and unit tests
│   ├── models.go         # Shared data models/structs
│   ├── helpers.go        # Utility functions
│   └── validators.go     # Custom validators
├── docs/                 # Generated documentation (do not edit manually)
├── examples/             # Example Terraform configurations
├── templates/            # Documentation templates for tfplugindocs
├── tools/                # Development tools (doc generation)
├── Makefile              # Build and test commands
└── go.mod                # Go module definition
```

## Development Commands

### Building

```bash
# Build the provider binary (uses GoReleaser)
make build

# Build and install to local Terraform plugins directory
make install
```

### Testing

```bash
# Run unit tests (no external dependencies required)
make test

# Run acceptance tests (requires SonarCloud environment)
make testacc

# Debug tests with delve
make debug-test
```

### Code Quality

```bash
# Format Go code
make fmt

# Generate documentation
make docs

# Verify docs are up-to-date (for CI)
make docs-check
```

## Environment Setup for Acceptance Tests

Acceptance tests require a configured SonarCloud organization. Set the following environment variables (you can use a `.env` file which is automatically loaded by the Makefile):

| Variable | Description |
|---|---|
| `SONARCLOUD_ORGANIZATION` | The SonarCloud organization to run tests against |
| `SONARCLOUD_TOKEN` | Admin token for the organization |
| `SONARCLOUD_TEST_USER_LOGIN` | Existing org member login (format: `<github_handle>@github`) |
| `SONARCLOUD_TEST_GROUP_NAME` | Existing group name for member tests |
| `SONARCLOUD_TOKEN_TEST_USER_LOGIN` | Login that owns the `SONARCLOUD_TOKEN` |
| `SONARCLOUD_PROJECT_KEY` | Test project key |
| `SONARCLOUD_QUALITY_GATE_ID` | Test quality gate ID |
| `SONARCLOUD_QUALITY_GATE_NAME` | Test quality gate name |

## Terraform Provider Best Practices

### Resource Implementation Pattern

When implementing a new resource (`resource_*.go`):

1. **Define the struct** with an embedded provider reference:
   ```go
   type resourceMyResource struct {
       p *sonarcloudProvider
   }
   ```

2. **Implement required interfaces**:
   - `resource.Resource` (Metadata, Schema, Create, Read, Update, Delete)
   - `resource.ResourceWithImportState` (if import is supported)

3. **Schema Design**:
    - Use `Required` for mandatory fields
    - Use `Optional` for optional fields
    - Use `Computed` for API-generated values (IDs, timestamps)
    - Always include `Description` for documentation
    - Add validators to attributes whenever possible to catch invalid input early (use helpers in `sonarcloud/validators.go`).

4. **Error Handling**:
   - Use `resp.Diagnostics.AddError()` for errors
   - Use `resp.Diagnostics.AddWarning()` for non-fatal issues
   - Provide clear, actionable error messages

### Data Source Implementation Pattern

Data sources (`data_source_*.go`) follow a similar pattern but are read-only:

1. Implement `datasource.DataSource` interface
2. Only implement `Read` method (no Create/Update/Delete)
3. Use for querying existing resources

### Testing Patterns

- **Unit tests**: Test helper functions and validators without API calls
- **Acceptance tests**: Full integration tests that create/update/delete real resources
- Test files are named `*_test.go` alongside their implementation files
- Use the `resource.Test()` helper with `TestStep` configurations
- Always include `CheckDestroy` to verify cleanup

Example test structure:
```go
func TestAccMyResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccMyResourceConfig(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("sonarcloud_my_resource.test", "name", "expected"),
                ),
            },
        },
        CheckDestroy: testAccMyResourceDestroy,
    })
}
```

### Documentation

- Documentation is auto-generated from schema descriptions using `tfplugindocs`
- Run `make docs` after schema changes
- Edit templates in `templates/` for custom documentation
- Example configurations go in `examples/` directory

## Code Style Guidelines

1. **Naming Conventions**:
   - Resources: `resource_<name>.go` with struct `resource<Name>`
   - Data Sources: `data_source_<name>.go` with struct `dataSource<Name>`
   - Tests: `*_test.go` in the same package

2. **Error Messages**: Be specific and include context:
   ```go
   resp.Diagnostics.AddError(
       "Could not create user group",
       fmt.Sprintf("The CreateRequest returned an error: %+v", err),
   )
   ```

3. **Provider Configuration Check**: Always verify provider is configured:
   ```go
   if !r.p.configured {
       resp.Diagnostics.AddError(
           "Provider not configured",
           "The provider hasn't been configured before apply...",
       )
       return
   }
   ```

## Common Tasks

### Adding a New Resource

1. Create `sonarcloud/resource_<name>.go` implementing the resource
2. Create `sonarcloud/resource_<name>_test.go` with acceptance tests
3. Register in `provider.go` under `Resources()` method
4. Add example in `examples/resources/<name>/resource.tf`
5. Make sure tests pass
6. Run `make docs` to generate documentation

### Adding a New Data Source

1. Create `sonarcloud/data_source_<name>.go` implementing the data source
2. Create `sonarcloud/data_source_<name>_test.go` with acceptance tests
3. Register in `provider.go` under `DataSources()` method
4. Add example in `examples/data-sources/<name>/data-source.tf`
5. Make sure tests pass
6. Run `make docs` to generate documentation

### Debugging

```bash
# Run acceptance tests with delve debugger
make debug-test

# Run specific test
TF_ACC=1 go test -v ./sonarcloud -run TestAccUserGroup -timeout 120m
```

## Important Notes

- The provider uses the Terraform Plugin Framework (not the older SDK v2 for new code)
- The `go-sonarcloud` client library handles SonarCloud API communication
- Always run `make fmt` before committing
- Ensure `make docs-check` passes (documentation is up-to-date)
- Acceptance tests create real resources in SonarCloud - use a dedicated test organization
