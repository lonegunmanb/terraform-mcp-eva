package conftest

import (
	"fmt"

	"github.com/spf13/afero"
)

// Global filesystem interface for testing (following tflint pattern)
var fs = afero.NewOsFs()

// Predefined policy configurations following the TODO specification
var predefinedPolicyConfigs = map[string][]string{
	"aprl":   {"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2"},
	"avmsec": {"git::https://github.com/Azure/policy-library-avm.git//policy/avmsec"},
	"all":    {"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2", "git::https://github.com/Azure/policy-library-avm.git//policy/avmsec"},
}

// resolvePolicyUrls resolves policy URLs based on predefined alias or custom URLs
func resolvePolicyUrls(predefinedAlias string, customUrls []string) ([]string, error) {
	// Check for mutually exclusive parameters
	if predefinedAlias != "" && len(customUrls) > 0 {
		return nil, fmt.Errorf("predefined_policy_library_alias and custom_urls are mutually exclusive")
	}

	// If custom URLs are provided, return them
	if len(customUrls) > 0 {
		return customUrls, nil
	}

	// If predefined alias is provided, resolve it
	if predefinedAlias != "" {
		urls, exists := predefinedPolicyConfigs[predefinedAlias]
		if !exists {
			return nil, fmt.Errorf("invalid predefined_policy_library_alias: %s", predefinedAlias)
		}
		return urls, nil
	}

	// Default to "all" when both are empty
	return predefinedPolicyConfigs["all"], nil
}
