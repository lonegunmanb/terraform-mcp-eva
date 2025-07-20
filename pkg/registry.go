package pkg

import (
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/prompt"
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterMcpServer(s *mcp.Server) {
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
