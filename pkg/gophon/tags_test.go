package gophon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSupportedTags(t *testing.T) {
	tests := []struct {
		name         string
		namespace    string
		expected     []string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "should return tags for azurerm namespace",
			namespace:   "github.com/hashicorp/terraform-provider-azurerm/internal",
			expectError: false,
			expected: []string{
				"v4.20.0",
				"v4.21.0",
			},
		},
		{
			name:         "should return error for unsupported namespace",
			namespace:    "unsupported.namespace",
			expectError:  true,
			errorMessage: "unsupported namespace: unsupported.namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, err := ListSupportedTags(tt.namespace)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tags)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
				return
			}
			assert.NoError(t, err)
			for _, expected := range tt.expected {
				assert.Contains(t, tags, expected, "Expected tag %s not found in %v", expected, tags)
			}
		})
	}
}
