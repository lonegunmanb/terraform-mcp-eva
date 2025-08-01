package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/gophon"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// QuerySupportedProviders is an MCP tool that returns all supported provider names
func QuerySupportedProviders(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParams) (*mcp.CallToolResultFor[any], error) {
	// Get supported providers using the core business logic
	providers := gophon.GetSupportedProviders()

	// Convert to JSON array format for consistent API response
	jsonBytes, err := json.Marshal(providers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal providers to JSON: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil
}
