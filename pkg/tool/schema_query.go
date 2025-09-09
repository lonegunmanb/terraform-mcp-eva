package tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tfschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SchemaQueryParam struct {
	Category          string `json:"category" jsonschema:"Terraform block type, possible values: resource, data, ephemeral, function, provider"`
	Type              string `json:"type" jsonschema:"Terraform block type like: azurerm_resource_group or function name like: can. Not required for provider category."`
	Path              string `json:"path,omitempty" jsonschema:"JSON path to query the resource schema, for example: default_node_pool.upgrade_settings, if not specified, the whole resource schema will be returned. Note: path queries are not supported for function schemas"`
	ProviderNamespace string `json:"namespace" jsonschema:"Provider namespace (e.g., 'hashicorp', 'Azure'). If not set, defaults to 'hashicorp'."`
	ProviderName      string `json:"name" jsonschema:"Provider name (e.g., 'aws', 'azurerm', 'azapi'). Required for provider category. For other categories, if not provided, will be inferred from the type parameter (except for functions)."`
	ProviderVersion   string `json:"version,omitempty" jsonschema:"Provider version or version constraint (e.g., '5.0.0', '~> 4.0', '>= 3.0, < 5.0'). If not specified, the latest version will be used."`
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

	validator := NewSchemaQueryValidator()

	// Validate parameters
	if err := validator.ValidateParams(category, t, path, namespace, name); err != nil {
		return nil, err
	}

	// Normalize namespace
	namespace = validator.NormalizeNamespace(namespace)

	// Infer provider name if needed
	var err error
	name, err = validator.InferProviderName(category, t, name)
	if err != nil {
		return nil, err
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
