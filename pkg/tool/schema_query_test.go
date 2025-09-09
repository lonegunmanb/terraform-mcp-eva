package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInferProviderNameFromType(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expected     string
	}{
		{
			name:         "AWS resource",
			resourceType: "aws_instance",
			expected:     "aws",
		},
		{
			name:         "Azure resource",
			resourceType: "azurerm_resource_group",
			expected:     "azurerm",
		},
		{
			name:         "Google Cloud resource",
			resourceType: "google_compute_instance",
			expected:     "google",
		},
		{
			name:         "Multiple underscores",
			resourceType: "aws_ec2_instance",
			expected:     "aws",
		},
		{
			name:         "No underscore",
			resourceType: "invalidtype",
			expected:     "",
		},
		{
			name:         "Empty string",
			resourceType: "",
			expected:     "",
		},
		{
			name:         "Starts with underscore",
			resourceType: "_resource",
			expected:     "",
		},
		{
			name:         "Only underscore",
			resourceType: "_",
			expected:     "",
		},
		{
			name:         "Ends with underscore",
			resourceType: "provider_",
			expected:     "provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferProviderNameFromType(tt.resourceType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidCategories(t *testing.T) {
	expectedCategories := []string{"resource", "data", "ephemeral", "function", "provider"}

	for _, category := range expectedCategories {
		t.Run(category, func(t *testing.T) {
			_, ok := validCategories[category]
			assert.True(t, ok, "Category %s should be valid", category)
		})
	}

	// Test invalid category
	_, ok := validCategories["invalid"]
	assert.False(t, ok, "Invalid category should not be in validCategories")
}

func TestSchemaQueryValidator_InferProviderName(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		resourceType string
		providerName string
		expected     string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "existing provider name preserved",
			category:     "resource",
			resourceType: "azurerm_resource_group",
			providerName: "azurerm",
			expected:     "azurerm",
			expectError:  false,
		},
		{
			name:         "infer provider name from resource type",
			category:     "resource",
			resourceType: "azurerm_resource_group",
			providerName: "",
			expected:     "azurerm",
			expectError:  false,
		},
		{
			name:         "infer provider name from data source type",
			category:     "data",
			resourceType: "aws_availability_zones",
			providerName: "",
			expected:     "aws",
			expectError:  false,
		},
		{
			name:         "provider category requires explicit provider name",
			category:     "provider",
			resourceType: "",
			providerName: "",
			expected:     "",
			expectError:  true,
			errorMsg:     "provider name is required when category is 'provider'",
		},
		{
			name:         "function category requires explicit provider name",
			category:     "function",
			resourceType: "",
			providerName: "",
			expected:     "",
			expectError:  true,
			errorMsg:     "provider name is required when category is 'function'",
		},
		{
			name:         "cannot infer provider name from invalid type",
			category:     "resource",
			resourceType: "invalidtype",
			providerName: "",
			expected:     "",
			expectError:  true,
			errorMsg:     "could not infer provider name from type 'invalidtype', please provide the 'name' parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := inferProviderName(tt.category, tt.resourceType, tt.providerName)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Equal(t, tt.expected, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
