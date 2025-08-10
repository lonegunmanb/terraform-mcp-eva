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

			for _, ns := range tt.expected {
				assert.Contains(t, result, ns)
			}
		})
	}
}
