package tool

import (
	"context"
	"fmt"
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tfschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FineGrainedSchemaQueryParam struct {
	Category string `json:"category" jsonschema:"Terraform block type, possible values: resource, data, ephemeral"`
	Type     string `json:"type" jsonschema:"Terraform block type like: azurerm_resource_group"`
	Path     string `json:"path,omitempty" jsonschema:"JSON path to query the resource schema, for example: default_node_pool.upgrade_settings, if not specified, the whole resource schema will be returned"`
}

var validCategories = map[string]struct{}{
	"resource":  {},
	"data":      {},
	"ephemeral": {},
}

func QueryFineGrainedSchema(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[FineGrainedSchemaQueryParam]) (*mcp.CallToolResultFor[any], error) {
	category := params.Arguments.Category
	t := params.Arguments.Type
	path := params.Arguments.Path
	if _, ok := validCategories[category]; !ok {
		return nil, fmt.Errorf("invalid category: %s", category)
	}
	schema, err := tfschema.QuerySchema(category, t, path)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema for %s %s: %w", category, t, err)
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: schema,
				Annotations: &mcp.Annotations{
					Audience: []mcp.Role{
						"assistant",
					},
				},
			},
		},
	}, nil
}
