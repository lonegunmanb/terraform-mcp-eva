package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProviderItemsValidator_ValidateParams(t *testing.T) {
	validator := NewListProviderItemsValidator()

	tests := []struct {
		name          string
		category      string
		namespace     string
		providerName  string
		version       string
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid ephemeral resources query",
			category:     "ephemeral",
			namespace:    "hashicorp",
			providerName: "azurerm",
			version:      "~> 4.0",
			expectError:  false,
		},
		{
			name:         "valid resources query",
			category:     "resource",
			namespace:    "hashicorp",
			providerName: "azurerm",
			version:      "~> 4.0",
			expectError:  false,
		},
		{
			name:         "valid data sources query",
			category:     "data",
			namespace:    "hashicorp",
			providerName: "azurerm",
			version:      "~> 4.0",
			expectError:  false,
		},
		{
			name:         "valid functions query",
			category:     "function",
			namespace:    "hashicorp",
			providerName: "azurerm",
			version:      "~> 4.0",
			expectError:  false,
		},
		{
			name:          "invalid category",
			category:      "invalid",
			namespace:     "hashicorp",
			providerName:  "azurerm",
			version:       "~> 4.0",
			expectError:   true,
			errorContains: "invalid category",
		},
		{
			name:          "missing provider name",
			category:      "resource",
			namespace:     "hashicorp",
			providerName:  "",
			version:       "~> 4.0",
			expectError:   true,
			errorContains: "provider name is required",
		},
		{
			name:         "missing namespace is valid (defaults to hashicorp)",
			category:     "resource",
			namespace:    "",
			providerName: "azurerm",
			version:      "~> 4.0",
			expectError:  false,
		},
		{
			name:         "missing version is valid (uses latest)",
			category:     "resource",
			namespace:    "hashicorp",
			providerName: "azurerm",
			version:      "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateParams(tt.category, tt.namespace, tt.providerName, tt.version)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestListProviderItemsValidator_NormalizeNamespace(t *testing.T) {
	validator := NewListProviderItemsValidator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty namespace defaults to hashicorp",
			input:    "",
			expected: "hashicorp",
		},
		{
			name:     "existing namespace preserved",
			input:    "azure",
			expected: "azure",
		},
		{
			name:     "hashicorp namespace preserved",
			input:    "hashicorp",
			expected: "hashicorp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.NormalizeNamespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestListProviderItemsValidator_Integration tests the validator in combination with the validation logic
func TestListProviderItemsValidator_Integration(t *testing.T) {
	validator := NewListProviderItemsValidator()

	// Test a complete validation flow
	category := "ephemeral"
	namespace := "" // Empty, should default to "hashicorp"
	providerName := "azurerm"
	version := "~> 4.0"

	// Validate parameters
	err := validator.ValidateParams(category, namespace, providerName, version)
	require.NoError(t, err, "Validation should succeed for valid parameters")

	// Normalize namespace
	normalizedNamespace := validator.NormalizeNamespace(namespace)
	assert.Equal(t, "hashicorp", normalizedNamespace, "Empty namespace should default to hashicorp")

	// This would be the point where the actual ListItems call would happen
	// but we're testing the validator in isolation to avoid plugin downloads
}
