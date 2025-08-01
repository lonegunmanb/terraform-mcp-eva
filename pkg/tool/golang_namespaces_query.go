package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/gophon"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// QuerySupportedGolangNamespaces is an MCP tool that returns all supported golang namespaces
func QuerySupportedGolangNamespaces(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResultFor[any], error) {
	// Get supported namespaces using the core business logic
	namespaces := gophon.ListSupportedNamespaces()

	// Convert to JSON array format for consistent API response
	jsonBytes, err := json.Marshal(namespaces)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal namespaces to JSON: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil
}
