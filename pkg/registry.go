package pkg

import (
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/prompt"
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
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
		InputSchema: &jsonschema.Schema{
			Type: "object",
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
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"namespace": {
					Type:        "string",
					Description: "The golang namespace to get tags for (e.g. 'github.com/hashicorp/terraform-provider-azurerm/internal')",
				},
			},
			Required: []string{"namespace"},
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
		InputSchema: &jsonschema.Schema{
			Type: "object",
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
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"block_type": {
					Type:        "string",
					Description: "The terraform block type (e.g. 'resource', 'data', 'ephemeral')",
				},
				"terraform_type": {
					Type:        "string",
					Description: "The terraform type (e.g. 'azurerm_resource_group')",
				},
				"entrypoint_name": {
					Type:        "string",
					Description: "The function or method name you want to read the source code (for 'resource': 'create', 'read', 'update', 'delete', 'schema', 'attribute'; for 'data': 'read', 'schema', 'attribute'; for 'ephemeral': 'open', 'close', 'renew', 'schema')",
				},
				"tag": {
					Type:        "string",
					Description: "Optional tag version, e.g.: v4.0.0 (defaults to latest version if not specified)",
				},
			},
			Required: []string{"block_type", "terraform_type", "entrypoint_name"},
		},
		Description: "Read Terraform provider source code for a given Terraform block, if you see `source code not found (404)` in error, it implies that maybe the function or method is not implemented in the provider. Use this tool when you need to: 1) Read the source code of a specific Terraform function or method, 2) How a Terraform Provider calls API, 3) Debug issues related to specific Terraform resource.",
		Name:        "query_terraform_block_implementation_source_code",
	}, tool.QueryTerraformSourceCode)
	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"namespace": {
					Type:        "string",
					Description: "[Required] The golang namespace to query (e.g. 'github.com/hashicorp/terraform-provider-azurerm/internal'). When you are reading golang source code and want to read a specific function, method, type or variable, you need to infer the correct namespace first. To infer the namespace of a given symbol, you must read 'package' declaration in the current golang code, along with all imports, then guess the symbol you'd like to read is in which namespace. The symbol could be placed in a different namespace, it's quite common.",
				},
				"symbol": {
					Type:        "string",
					Description: "[Required] The symbol you want to read, possible values: 'func', 'method', 'type', 'var'",
					Enum:        []interface{}{"func", "method", "type", "var"},
				},
				"receiver": {
					Type:        "string",
					Description: "The type of method receiver, e.g.: 'ContainerAppResource'. Can only be set when symbol is 'method'.",
				},
				"name": {
					Type:        "string",
					Description: "[Required] The name of the function, method, type or variable you want to read. For example: 'NewContainerAppResource', 'ContainerAppResource'",
				},
				"tag": {
					Type:        "string",
					Description: "Optional tag version, e.g.: v4.0.0 (defaults to latest version if not specified)",
				},
			},
			Required: []string{"namespace", "symbol", "name"},
		},
		Description: "Read golang source code for given type, variable, constant, function or method definition, if you see `source code not found (404)` in error, it implies that maybe the function or method is not implemented in the provider, or it could be a variable with function type. `symbol` set to `var` for variable or constant, `type` for type definition including struct, interface or type alias, `func` for function without receiver, `method` for method that has receiver. If you want to know how a Terraform resource is implemented, you should call `query_terraform_block_implementation_source_code` before you call this tool. Use this tool when you need to: 1) You want to see other function, method, type, variable's definition while you're reading golang source code, 2) How a Terraform Provider expand or flatten struct, 3) Debug issues related to specific Terraform resource.",
		Name:        "query_golang_source_code",
	}, tool.QueryGolangSourceCode)

	mcp.AddTool(s, &mcp.Tool{
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: p(false),
			IdempotentHint:  true,
			OpenWorldHint:   p(false),
			ReadOnlyHint:    true,
		},
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"resource_type": {
					Type:        "string",
					Description: "Azure resource type, for example: Microsoft.Compute/virtualMachines, combined with api_version to identify the resource schema, like: Microsoft.Compute/virtualMachines@2024-11-01",
				},
				"api_version": {
					Type:        "string",
					Description: "Azure resource api-version, for example: 2024-11-01, combined with resource_type to identify the resource schema, like: Microsoft.Compute/virtualMachines@2024-11-01",
				},
				"path": {
					Type:        "string",
					Description: "JSON path to query the resource schema, for example: body.properties.osProfile.secrets.sourceVault.id, if not specified, the whole resource schema will be returned",
				},
			},
			Required: []string{"resource_type", "api_version"},
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
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"resource_type": {
					Type:        "string",
					Description: "Azure resource type, for example: Microsoft.Compute/virtualMachines",
				},
			},
			Required: []string{"resource_type"},
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
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"resource_type": {
					Type:        "string",
					Description: "Azure resource type, for example: Microsoft.Compute/virtualMachines, combined with api_version to identify the resource schema, like: Microsoft.Compute/virtualMachines@2024-11-01",
				},
				"api_version": {
					Type:        "string",
					Description: "Azure resource api-version, for example: 2024-11-01, combined with resource_type to identify the resource schema, like: Microsoft.Compute/virtualMachines@2024-11-01",
				},
				"path": {
					Type:        "string",
					Description: "JSON path to query the resource schema, for example: body.properties.osProfile.secrets.sourceVault.id, if not specified, the whole resource schema will be returned",
				},
			},
			Required: []string{"resource_type", "api_version"},
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
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"category": {
					Type:        "string",
					Description: "Terraform block type, possible values: resource, data, ephemeral",
					Enum:        []interface{}{"resource", "data", "ephemeral"},
				},
				"type": {
					Type:        "string",
					Description: "Terraform block type like: azurerm_resource_group",
				},
				"path": {
					Type:        "string",
					Description: "JSON path to query the resource schema, for example: default_node_pool.upgrade_settings, if not specified, the whole resource schema will be returned",
				},
			},
			Required: []string{"category", "type"},
		},
		Description: "[You should use this tool before you try resolveProviderDocID]Query fine grained Terraform resource schema by `category`, `name` and optional `path`. The returned value is a json string representing the resource schema, including attribute descriptions, which can be used in Terraform provider schema. If you're querying schema information about specified attribute or nested block schema of a resource from supported provider, this tool should have higher priority. Only support `azurerm`, `azuread`, `aws`, `awscc`, `google` providers now.",
		Name:        "query_terraform_fine_grained_document",
	}, tool.QueryFineGrainedSchema)
	prompt.AddSolveAvmIssuePrompt(s)
}

func p[T any](input T) *T {
	return &input
}
