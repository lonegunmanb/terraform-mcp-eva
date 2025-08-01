package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/gophon"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GolangTagsQueryParam struct {
	Namespace string `json:"namespace" jsonschema:"The golang namespace to get tags for (e.g. 'github.com/hashicorp/terraform-provider-azurerm/internal')"`
}

// QuerySupportedTags is an MCP tool that returns all supported tags for a specific golang namespace
func QuerySupportedTags(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GolangTagsQueryParam]) (*mcp.CallToolResultFor[any], error) {
	namespace := params.Arguments.Namespace
	if namespace == "" {
		return nil, fmt.Errorf("namespace parameter is required")
	}

	// Get supported tags using the core business logic
	tags, err := gophon.ListSupportedTags(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get supported tags for namespace %q: %w", namespace, err)
	}

	// Convert to JSON array format for consistent API response
	jsonBytes, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags to JSON: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonBytes),
			},
		},
	}, nil
}
