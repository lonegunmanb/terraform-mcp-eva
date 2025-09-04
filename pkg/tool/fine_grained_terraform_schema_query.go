package tool

import (
	"context"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tfschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FineGrainedSchemaQueryParam struct {
	Category          string `json:"category" jsonschema:"Terraform block type, possible values: resource, data_source, ephemeral, function"`
	Type              string `json:"type" jsonschema:"Terraform block type like: azurerm_resource_group or function name like: can"`
	Path              string `json:"path,omitempty" jsonschema:"JSON path to query the resource schema, for example: default_node_pool.upgrade_settings, if not specified, the whole resource schema will be returned. Note: path queries are not supported for function schemas"`
	ProviderNamespace string `json:"namespace" jsonschema:"Provider namespace (e.g., 'hashicorp', 'Azure')"`
	ProviderName      string `json:"name" jsonschema:"Provider name (e.g., 'aws', 'azurerm', 'azapi')"`
	ProviderVersion   string `json:"version,omitempty" jsonschema:"Provider version (e.g., '5.0.0', '4.39.0'). If not specified, the latest version will be used."`
}

var validCategories = map[string]struct{}{
	"resource":    {},
	"data_source": {},
	"ephemeral":   {},
	"function":    {},
}

func QueryFineGrainedSchema(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[FineGrainedSchemaQueryParam]) (*mcp.CallToolResultFor[any], error) {
	category := params.Arguments.Category
	t := params.Arguments.Type
	path := params.Arguments.Path
	namespace := params.Arguments.ProviderNamespace
	name := params.Arguments.ProviderName
	version := params.Arguments.ProviderVersion

	if _, ok := validCategories[category]; !ok {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	if namespace == "" || name == "" {
		return nil, fmt.Errorf("provider_namespace and provider_name are required parameters")
	}

	providerReq := tfschema.ProviderRequest{
		ProviderNamespace: namespace,
		ProviderName:      name,
		ProviderVersion:   version,
	}

	schema, err := tfschema.QuerySchema(category, t, path, providerReq)
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
