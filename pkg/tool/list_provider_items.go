package tool

import (
	"context"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tfschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListItemsParam struct {
	Category          string `json:"category" jsonschema:"Terraform item type to list, possible values: resource, data, ephemeral, function"`
	ProviderNamespace string `json:"namespace" jsonschema:"Provider namespace (e.g., 'hashicorp', 'Azure'). If not set, defaults to 'hashicorp'."`
	ProviderName      string `json:"name" jsonschema:"Provider name (e.g., 'aws', 'azurerm', 'azapi'). Required parameter."`
	ProviderVersion   string `json:"version,omitempty" jsonschema:"Provider version or version constraint (e.g., '5.0.0', '~> 4.0', '>= 3.0, < 5.0'). If not specified, the latest version will be used."`
}

func ListProviderItems(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ListItemsParam]) (*mcp.CallToolResultFor[any], error) {
	category := params.Arguments.Category
	namespace := params.Arguments.ProviderNamespace
	name := params.Arguments.ProviderName
	version := params.Arguments.ProviderVersion

	if _, ok := validCategories[category]; !ok {
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	// Default namespace to hashicorp if not provided
	if namespace == "" {
		namespace = "hashicorp"
	}

	// Provider name is always required
	if name == "" {
		return nil, fmt.Errorf("provider name is required")
	}

	providerReq := tfschema.ProviderRequest{
		ProviderNamespace: namespace,
		ProviderName:      name,
		ProviderVersion:   version,
	}

	items, err := tfschema.ListItems(category, providerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list %s items: %w", category, err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Found %d %s items for provider %s/%s:\n%v", len(items), category, namespace, name, items),
				Annotations: &mcp.Annotations{
					Audience: []mcp.Role{
						"assistant",
					},
				},
			},
		},
	}, nil
}
