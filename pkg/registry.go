package pkg

import (
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/prompt"
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterMcpServer(s *mcp.Server) {
	// Tool 1: Get Supported Golang Namespaces
	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		Description: "Get all indexed golang namespaces available for analysis. Returns a JSON array of namespace strings like ['github.com/hashicorp/terraform-provider-azurerm/internal']. Use this tool when you are reading Golang source code and need to: 1) Discover what golang projects/packages have been indexed, 2) Find available namespaces before querying specific code symbols, functions, or types, 3) Understand the scope of indexed golang codebases available for analysis.",
		Name:        "golang_source_code_server_get_supported_golang_namespaces",
	}, tool.QuerySupportedGolangNamespaces)

	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		Description: "Get all supported tags/versions for a specific golang namespace. Requires a 'namespace' parameter (string) and returns a JSON array of version tags like ['v4.20.0', 'v4.21.0']. Use this tool when you need to: 1) Discover available versions/tags for a specific golang namespace, 2) Find the latest or specific versions before analyzing code from a particular tag, 3) Understand version history for indexed golang projects.",
		Name:        "golang_source_code_server_get_supported_tags",
	}, tool.QuerySupportedTags)

	// Tool 3: Get Supported Providers
	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		Description: "Get all supported Terraform provider names available for source code query. Returns a JSON array of provider name strings like ['azurerm']. Use this tool when you need to: 1) Discover what Terraform providers have been indexed and are available for golang source query, you can study details of provider's behavior, 2) Find available providers before querying specific golang functions, methods, types, variables.",
		Name:        "terraform_source_code_query_get_supported_providers",
	}, tool.QuerySupportedProviders)

	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		Description: "[You should use this tool before you try resolveProviderDocID]Query fine grained AzAPI resource schema by `resource type`, `api_version` and optional `path`. The returned type is a Go type string, which can be used in Go code to represent the resource schema. If you're querying AzAPI provider resource schema, this tool should have higher priority",
		Name:        "query_azapi_resource_schema",
	}, tool.QueryAzAPIResourceSchema)
	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		Description: "[You should use this tool before you try resolveProviderDocID]Query Azure API versions by `resource type`. The returned value is a list of API versions for the specified resource type, split by comma.",
		Name:        "list_azapi_api_versions",
	}, tool.QueryAzAPIVersions)
	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		Description: "[You should use this tool before you try resolveProviderDocID]Query fine grained AzAPI resource description by `resource type`, `api_version` and optional `path`. The returned value is either description of the property, or json object representing the object, the key is property name the value is the description of the property. Via description you can learn whether a property is id, readonly or writeonly, and possible values. If you're querying AzAPI provider resource description, this tool should have higher priority",
		Name:        "query_azapi_resource_document",
	}, tool.QueryAzAPIDescriptionSchema)
	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		Description: "[You should use this tool before you try resolveProviderDocID]Query fine grained Terraform resource schema by `category`, `name` and optional `path`. The returned value is a json string representing the resource schema, including attribute descriptions, which can be used in Terraform provider schema. If you're querying schema information about specified attribute or nested block schema of a resource from supported provider, this tool should have higher priority. Only support `azurerm`, `azuread`, `aws`, `awscc`, `google` providers now.",
		Name:        "query_terraform_fine_grained_document",
	}, tool.QueryFineGrainedSchema)
	prompt.AddSolveAvmIssuePrompt(s)
}

func p[T any](input T) *T {
	return &input
}
