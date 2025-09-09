package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaQueryValidator_ValidateParams(t *testing.T) {
	validator := NewSchemaQueryValidator()

	tests := []struct {
		name         string
		category     string
		resourceType string
		path         string
		namespace    string
		providerName string
		expectError  bool
		errorMsg     string
	}{
		{
			name:         "valid resource query",
			category:     "resource",
			resourceType: "azurerm_resource_group",
			path:         "",
			namespace:    "hashicorp",
			providerName: "azurerm",
			expectError:  false,
		},
		{
			name:         "valid provider query",
			category:     "provider",
			resourceType: "",
			path:         "",
			namespace:    "hashicorp",
			providerName: "azurerm",
			expectError:  false,
		},
		{
			name:         "valid provider query with path",
			category:     "provider",
			resourceType: "",
			path:         "client_certificate",
			namespace:    "hashicorp",
			providerName: "azurerm",
			expectError:  false,
		},
		{
			name:         "valid function query",
			category:     "function",
			resourceType: "",
			path:         "",
			namespace:    "hashicorp",
			providerName: "random",
			expectError:  false,
		},
		{
			name:         "invalid category",
			category:     "invalid",
			resourceType: "azurerm_resource_group",
			path:         "",
			namespace:    "hashicorp",
			providerName: "azurerm",
			expectError:  true,
			errorMsg:     "invalid category: invalid",
		},
		{
			name:         "provider category missing provider name",
			category:     "provider",
			resourceType: "",
			path:         "",
			namespace:    "hashicorp",
			providerName: "",
			expectError:  true,
			errorMsg:     "provider name is required when category is 'provider'",
		},
		{
			name:         "function category missing provider name",
			category:     "function",
			resourceType: "",
			path:         "",
			namespace:    "hashicorp",
			providerName: "",
			expectError:  true,
			errorMsg:     "provider name is required when category is 'function'",
		},
		{
			name:         "function category with path not supported",
			category:     "function",
			resourceType: "",
			path:         "some.path",
			namespace:    "hashicorp",
			providerName: "random",
			expectError:  true,
			errorMsg:     "path queries are not supported for function schemas",
		},
		{
			name:         "resource category with invalid type and no provider name",
			category:     "resource",
			resourceType: "invalidtypewithoutnounderscore",
			path:         "",
			namespace:    "hashicorp",
			providerName: "",
			expectError:  true,
			errorMsg:     "could not infer provider name from type 'invalidtypewithoutnounderscore', please provide the 'name' parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateParams(tt.category, tt.resourceType, tt.path, tt.namespace, tt.providerName)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSchemaQueryValidator_NormalizeNamespace(t *testing.T) {
	validator := NewSchemaQueryValidator()

	tests := []struct {
		name      string
		namespace string
		expected  string
	}{
		{
			name:      "empty namespace defaults to hashicorp",
			namespace: "",
			expected:  "hashicorp",
		},
		{
			name:      "existing namespace preserved",
			namespace: "Azure",
			expected:  "Azure",
		},
		{
			name:      "hashicorp namespace preserved",
			namespace: "hashicorp",
			expected:  "hashicorp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.NormalizeNamespace(tt.namespace)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSchemaQueryValidator_InferProviderName(t *testing.T) {
	validator := NewSchemaQueryValidator()

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
			result, err := validator.InferProviderName(tt.category, tt.resourceType, tt.providerName)

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
