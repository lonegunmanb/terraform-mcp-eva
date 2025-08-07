package tflint

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		expected string
	}{
		{
			name:     "empty string should default to reusable",
			category: "",
			expected: "reusable",
		},
		{
			name:     "reusable should remain reusable",
			category: "reusable",
			expected: "reusable",
		},
		{
			name:     "example should remain example",
			category: "example",
			expected: "example",
		},
		{
			name:     "invalid category should default to reusable",
			category: "invalid",
			expected: "reusable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDefaultCategory(tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigURL(t *testing.T) {
	tests := []struct {
		name     string
		category string
		expected string
	}{
		{
			name:     "reusable category should return correct URL",
			category: "reusable",
			expected: "https://raw.githubusercontent.com/Azure/avm-terraform-governance/refs/heads/main/tflint-configs/avm.tflint.hcl",
		},
		{
			name:     "example category should return correct URL",
			category: "example",
			expected: "https://raw.githubusercontent.com/Azure/avm-terraform-governance/refs/heads/main/tflint-configs/avm.tflint_example.hcl",
		},
		{
			name:     "empty category should return reusable URL",
			category: "",
			expected: "https://raw.githubusercontent.com/Azure/avm-terraform-governance/refs/heads/main/tflint-configs/avm.tflint.hcl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getConfigURL(tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultTargetPath(t *testing.T) {
	tests := []struct {
		name       string
		targetPath string
		shouldErr  bool
	}{
		{
			name:       "empty target path should return current directory",
			targetPath: "",
			shouldErr:  false,
		},
		{
			name:       "relative path should be converted to absolute",
			targetPath: "./test",
			shouldErr:  false,
		},
		{
			name:       "absolute path should remain absolute",
			targetPath: "/absolute/path",
			shouldErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getDefaultTargetPath(tt.targetPath)

			if tt.shouldErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result)

			// Result should always be an absolute path
			assert.True(t, filepath.IsAbs(result))

			if tt.targetPath != "" && tt.targetPath != "." {
				// For non-empty paths, check specific cases
				if tt.targetPath == "/absolute/path" {
					// On Windows, this gets converted to a Windows absolute path
					assert.True(t, filepath.IsAbs(result))
				} else if tt.targetPath == "./test" {
					assert.Contains(t, result, "test")
				}
			}
		})
	}
}

func TestValidateCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		expected bool
	}{
		{
			name:     "reusable is valid",
			category: "reusable",
			expected: true,
		},
		{
			name:     "example is valid",
			category: "example",
			expected: true,
		},
		{
			name:     "empty string is invalid",
			category: "",
			expected: false,
		},
		{
			name:     "invalid category is invalid",
			category: "invalid",
			expected: false,
		},
		{
			name:     "case sensitive - Reusable is invalid",
			category: "Reusable",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateCategory(tt.category)
			assert.Equal(t, tt.expected, result)
		})
	}
}
