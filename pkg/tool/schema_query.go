package tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tfschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SchemaQueryParam struct {
	Category          string `json:"category" jsonschema:"Terraform block type, possible values: resource, data, ephemeral, function"`
	Type              string `json:"type" jsonschema:"Terraform block type like: azurerm_resource_group or function name like: can"`
	Path              string `json:"path,omitempty" jsonschema:"JSON path to query the resource schema, for example: default_node_pool.upgrade_settings, if not specified, the whole resource schema will be returned. Note: path queries are not supported for function schemas"`
	ProviderNamespace string `json:"namespace" jsonschema:"Provider namespace (e.g., 'hashicorp', 'Azure'). If not set, defaults to 'hashicorp'."`
	ProviderName      string `json:"name" jsonschema:"Provider name (e.g., 'aws', 'azurerm', 'azapi'). If not provided, will be inferred from the type parameter (except for functions)."`
	ProviderVersion   string `json:"version,omitempty" jsonschema:"Provider version (e.g., '5.0.0', '4.39.0'). If not specified, the latest version will be used."`
}

var validCategories = map[string]struct{}{
	"resource":  {},
	"data":      {},
	"ephemeral": {},
	"function":  {},
}

// inferProviderNameFromType extracts the provider name from a resource/data/ephemeral type
// Examples: "aws_ec2_instance" -> "aws", "azurerm_resource_group" -> "azurerm"
func inferProviderNameFromType(resourceType string) string {
	// Split by underscore and take the first part as provider name
	parts := strings.Split(resourceType, "_")
	if len(parts) >= 2 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func QuerySchema(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[SchemaQueryParam]) (*mcp.CallToolResultFor[any], error) {
	category := params.Arguments.Category
	t := params.Arguments.Type
	path := params.Arguments.Path
	namespace := params.Arguments.ProviderNamespace
	name := params.Arguments.ProviderName
	version := params.Arguments.ProviderVersion

	if _, ok := validCategories[category]; !ok {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	// For function category, name parameter is required
	if category == "function" && name == "" {
		return nil, fmt.Errorf("provider name is required when category is 'function'")
	}

	// For function category, path queries are not supported
	if category == "function" && path != "" {
		return nil, fmt.Errorf("path queries are not supported for function schemas")
	}

	// Default namespace to hashicorp if not provided
	if namespace == "" {
		namespace = "hashicorp"
	}

	// Infer provider name from type if not provided (except for functions)
	if name == "" {
		if category == "function" {
			return nil, fmt.Errorf("provider name is required when category is 'function'")
		}

		// Extract provider name from type (e.g., "aws_ec2_instance" -> "aws")
		inferredName := inferProviderNameFromType(t)
		if inferredName == "" {
			return nil, fmt.Errorf("could not infer provider name from type '%s', please provide the 'name' parameter", t)
		}
		name = inferredName
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
