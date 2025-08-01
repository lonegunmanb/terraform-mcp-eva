package gophon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSupportedNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name: "should return hardcoded supported golang namespaces",
			expected: []string{
				"github.com/hashicorp/terraform-provider-azurerm/internal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListSupportedNamespaces()

			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, 1, "initially should only have azurerm namespace")
		})
	}
}

func TestGetNamespaceBaseURL(t *testing.T) {
	tests := []struct {
		name          string
		namespace     string
		expectedURL   string
		expectError   bool
	}{
		{
			name:        "should return base URL for azurerm namespace",
			namespace:   "github.com/hashicorp/terraform-provider-azurerm/internal",
			expectedURL: "https://raw.githubusercontent.com/lonegunmanb/terraform-provider-azurerm-index/refs/heads/main/index/internal",
			expectError: false,
		},
		{
			name:        "should return error for unsupported namespace",
			namespace:   "unsupported.namespace",
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := GetNamespaceBaseURL(tt.namespace)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}
