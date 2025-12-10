package sonarcloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestChangedAttrs tests that changedAttrs properly detects changed attributes
func TestChangedAttrs(t *testing.T) {
	ctx := context.Background()

	// Create test schema
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional: true,
			},
			"key": schema.StringAttribute{
				Optional: true,
			},
			"visibility": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	// Test data structures for plan and state
	type TestData struct {
		Name       types.String `tfsdk:"name"`
		Key        types.String `tfsdk:"key"`
		Visibility types.String `tfsdk:"visibility"`
	}

	// Create plan with updated values
	planData := TestData{
		Name:       types.StringValue("updated_name"),
		Key:        types.StringValue("same_key"),
		Visibility: types.StringValue("public"),
	}

	// Create state with original values
	stateData := TestData{
		Name:       types.StringValue("original_name"),
		Key:        types.StringValue("same_key"),
		Visibility: types.StringValue("private"),
	}

	// Create tfsdk States and Plans
	var planObj tfsdk.Plan
	var stateObj tfsdk.State

	planObj.Schema = testSchema
	stateObj.Schema = testSchema

	diags := planObj.Set(ctx, planData)
	if diags.HasError() {
		t.Fatalf("Failed to set plan: %v", diags)
	}

	diags = stateObj.Set(ctx, stateData)
	if diags.HasError() {
		t.Fatalf("Failed to set state: %v", diags)
	}

	// Create UpdateRequest
	req := resource.UpdateRequest{
		Plan:  planObj,
		State: stateObj,
	}

	// Test changedAttrs
	changed := changedAttrs(ctx, req)

	// Verify name and visibility changed but key did not
	if _, ok := changed["name"]; !ok {
		t.Error("Expected 'name' to be in changed attributes")
	}
	if _, ok := changed["visibility"]; !ok {
		t.Error("Expected 'visibility' to be in changed attributes")
	}
	if _, ok := changed["key"]; ok {
		t.Error("Expected 'key' to NOT be in changed attributes")
	}

	// Verify we detected exactly 2 changes
	if len(changed) != 2 {
		t.Errorf("Expected 2 changed attributes, got %d", len(changed))
	}
}
