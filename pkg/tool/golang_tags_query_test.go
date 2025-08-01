package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerySupportedTags(t *testing.T) {
	tests := []struct {
		name           string
		arguments      GolangTagsQueryParam
		expectedError  bool
		expectedTags   []string
		errorSubstring string
	}{
		{
			name: "valid namespace",
			arguments: GolangTagsQueryParam{
				Namespace: "github.com/hashicorp/terraform-provider-azurerm/internal",
			},
			expectedError: false,
			expectedTags:  []string{"v4.20.0", "v4.21.0"}, // These might change over time
		},
		{
			name:           "missing namespace parameter",
			arguments:      GolangTagsQueryParam{},
			expectedError:  true,
			errorSubstring: "namespace parameter is required",
		},
		{
			name: "empty namespace parameter",
			arguments: GolangTagsQueryParam{
				Namespace: "",
			},
			expectedError:  true,
			errorSubstring: "namespace parameter is required",
		},
		{
			name: "unsupported namespace",
			arguments: GolangTagsQueryParam{
				Namespace: "unsupported/namespace",
			},
			expectedError:  true,
			errorSubstring: "unsupported namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &mcp.CallToolParamsFor[GolangTagsQueryParam]{
				Arguments: tt.arguments,
			}

			// Call the function
			result, err := QuerySupportedTags(context.Background(), nil, params)

			// Check error expectations
			if tt.expectedError {
				assert.Error(t, err, "Expected an error but got none")
				if tt.errorSubstring != "" && !containsSubstring(err.Error(), tt.errorSubstring) {
					t.Errorf("Expected error to contain %q, but got: %v", tt.errorSubstring, err)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Content, 1)
			textContent, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)

			// Parse the JSON response
			var tags []string
			require.NoError(t, json.Unmarshal([]byte(textContent.Text), &tags))
			assert.True(t, tt.name != "valid namespace" || len(tags) > 0, "Expected tags to be present for valid namespace")
		})
	}
}

func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || containsSubstring(str[1:], substr) || (len(str) > 0 && containsSubstring(str[:len(str)-1], substr)))
}
