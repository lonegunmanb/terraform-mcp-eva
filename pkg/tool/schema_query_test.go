package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
