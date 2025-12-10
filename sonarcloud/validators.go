package sonarcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Copied from https://www.terraform.io/plugin/framework/validation
type stringLengthBetweenValidator struct {
	Min int
	Max int
}

func stringLengthBetween(min int, max int) *stringLengthBetweenValidator {
	return &stringLengthBetweenValidator{Min: min, Max: max}
}

func (v stringLengthBetweenValidator) Description(ctx context.Context) string {
	return fmt.Sprintf("string length must be between %d and %d", v.Min, v.Max)
}

func (v stringLengthBetweenValidator) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("string length must be between `%d` and `%d`", v.Min, v.Max)
}

// Validate checks if the length of the string attribute is between Min and Max
func (v stringLengthBetweenValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// In v1+ API, req.ConfigValue is already the proper type
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	strLen := len(req.ConfigValue.ValueString())

	if strLen < v.Min || strLen > v.Max {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid String Length",
			fmt.Sprintf("String length must be between %d and %d, got: %d.", v.Min, v.Max, strLen),
		)

		return
	}
}

type allowedOptionsValidator struct {
	Options []string
}

func allowedOptions(options ...string) *allowedOptionsValidator {
	return &allowedOptionsValidator{Options: options}
}

func (v allowedOptionsValidator) Description(_ context.Context) string {
	return fmt.Sprintf("option must be one of %v", v.Options)
}

func (v allowedOptionsValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("option must be one of `%v`", v.Options)
}

func (v allowedOptionsValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// In v1+ API, req.ConfigValue is already the proper type
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	valid := false
	strValue := req.ConfigValue.ValueString()
	for _, option := range v.Options {
		if option == strValue {
			valid = true
			break
		}
	}

	if !valid {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid String Value",
			fmt.Sprintf("String must be one of %v, got: %s.", v.Options, strValue),
		)

		return
	}
}

type allowedSetOptionsValidator struct {
	Options []string
}

func allowedSetOptions(options ...string) *allowedSetOptionsValidator {
	return &allowedSetOptionsValidator{Options: options}
}

func (v allowedSetOptionsValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value in set must be one of %v", v.Options)
}

func (v allowedSetOptionsValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("value in set must be one of `%v`", v.Options)
}

func (v allowedSetOptionsValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	// In v1+ API, req.ConfigValue is already the proper type
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var values []string
	diags := req.ConfigValue.ElementsAs(ctx, &values, true)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	options := make(map[string]struct{})
	for _, option := range v.Options {
		options[option] = struct{}{}
	}

	for _, val := range values {
		if _, ok := options[val]; !ok {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid String Element in Set",
				fmt.Sprintf("Element must be one of %v, got: %s.", v.Options, val),
			)

			return
		}
	}
}
