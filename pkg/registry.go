package pkg

import "github.com/modelcontextprotocol/go-sdk/mcp"

func RegisterMcpServer(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Description: "Query Azure resource schema by `resource type`, `api_version` and optional `path`. The `resource type` and `api_version` are required parameters, while `path` is optional. If `path` is not specified, the whole resource schema will be returned. The returned type is a Go type string, which can be used in Go code to represent the resource schema.",
		Name:        "azure_resource_schema_query",
	}, QueryAzAPIResourceSchema)
	mcp.AddTool(s, &mcp.Tool{
		Description: "Query Azure API versions by `resource type`. The `resource type` is required parameter, which is the Azure resource type, for example: `Microsoft.Compute/virtualMachines`. The returned value is a list of API versions for the specified resource type, split by comma.",
		Name:        "azure_api_versions_query",
	}, QueryAzAPIVersions)
	mcp.AddTool(s, &mcp.Tool{
		Description: "Query Azure resource description by `resource type`, `api_version` and optional `path`. The `resource type` and `api_version` are required parameters, while `path` is optional. If `path` is not specified, the whole resource description will be returned. The returned value is either description of the property, or json object representing the object, the key is property name the value is the description of the property. Via description you can learn whether a property is id, readonly or writeonly, and possible values.",
		Name:        "azure_resource_description_query",
	}, QueryAzAPIDescriptionSchema)
}
