package tool

import (
	"context"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/gophon"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TerraformSourceCodeQueryParam struct {
	BlockType      string `json:"block_type" jsonschema:"The terraform block type (e.g. 'resource', 'data', 'ephemeral')"`
	TerraformType  string `json:"terraform_type" jsonschema:"The terraform type (e.g. 'azurerm_resource_group')"`
	EntrypointName string `json:"entrypoint_name" jsonschema:"The function or method name you want to read the source code (for 'resource': 'create', 'read', 'update', 'delete', 'schema', 'attribute'; for 'data': 'read', 'schema', 'attribute'; for 'ephemeral': 'open', 'close', 'renew', 'schema')"`
	Tag            string `json:"tag,omitempty" jsonschema:"Optional tag version, e.g.: v4.0.0 (defaults to latest version if not specified)"`
}

// QueryTerraformSourceCode is an MCP tool that returns terraform source code for a specific block type, terraform type, and entrypoint
func QueryTerraformSourceCode(_ context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[TerraformSourceCodeQueryParam]) (*mcp.CallToolResultFor[any], error) {
	blockType := params.Arguments.BlockType
	terraformType := params.Arguments.TerraformType
	entrypointName := params.Arguments.EntrypointName
	tag := params.Arguments.Tag

	if blockType == "" {
		return nil, fmt.Errorf("block_type parameter is required")
	}
	if terraformType == "" {
		return nil, fmt.Errorf("terraform_type parameter is required")
	}
	if entrypointName == "" {
		return nil, fmt.Errorf("entrypoint_name parameter is required")
	}

	// Get terraform source code using the core business logic
	sourceCode, err := gophon.GetTerraformSourceCode(blockType, terraformType, entrypointName, tag)
	if err != nil {
		return nil, fmt.Errorf("failed to get terraform source code for %s %s.%s: %w", blockType, terraformType, entrypointName, err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: sourceCode,
			},
		},
	}, nil
}
