package gophon

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSupportedProviders(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name:     "should return all supported provider names",
			expected: []string{"azurerm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSupportedProviders()

			assert.Len(t, result, len(tt.expected))
			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Expected provider %s to be in the list", expected)
			}
		})
	}
}
