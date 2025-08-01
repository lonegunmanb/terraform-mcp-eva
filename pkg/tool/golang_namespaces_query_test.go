package tool

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerySupportedGolangNamespaces(t *testing.T) {
	tests := []struct {
		name               string
		expectedNamespaces []string
		expectError        bool
	}{
		{
			name: "should return supported golang namespaces",
			expectedNamespaces: []string{
				"github.com/hashicorp/terraform-provider-azurerm/internal",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			session := &mcp.ServerSession{} // Mock session - actual session not needed for this test
			params := &mcp.CallToolParams{}

			result, err := QuerySupportedGolangNamespaces(ctx, session, params)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok, "expected TextContent")

				// The result should contain the expected namespaces in JSON array format
				assert.Contains(t, textContent.Text, "github.com/hashicorp/terraform-provider-azurerm/internal")
				assert.Contains(t, textContent.Text, "[")
				assert.Contains(t, textContent.Text, "]")
			}
		})
	}
}
