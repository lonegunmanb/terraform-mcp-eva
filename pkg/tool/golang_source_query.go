package tool

import (
	"context"
	"fmt"
	"github.com/lonegunmanb/terraform-mcp-eva/pkg/gophon"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"strings"
)

type GolangSourceCodeQueryParam struct {
	Namespace string `json:"namespace" jsonschema:"[Required] The golang namespace to query (e.g. 'github.com/hashicorp/terraform-provider-azurerm/internal'). When you are reading golang source code and want to read a specific function, method, type or variable, you need to infer the correct namespace first. To infer the namespace of a given symbol, you must read 'package' declaration in the current golang code, along with all imports, then guess the symbol you'd like to read is in which namespace. The symbol could be placed in a different namespace, it's quite common."`
	Symbol    string `json:"symbol" jsonschema:"[Required] The symbol you want to read, possible values: 'func', 'method', 'type', 'var'"`
	Receiver  string `json:"receiver,omitempty" jsonschema:"The type of method receiver, e.g.: 'ContainerAppResource'. Can only be set when symbol is 'method'."`
	Name      string `json:"name" jsonschema:"[Required] The name of the function, method, type or variable you want to read. For example: 'NewContainerAppResource', 'ContainerAppResource'"`
	Tag       string `json:"tag,omitempty" jsonschema:"Optional tag version, e.g.: v4.0.0 (defaults to latest version if not specified)"`
}

func QueryGolangSourceCode(_ context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[GolangSourceCodeQueryParam]) (*mcp.CallToolResultFor[any], error) {
	symbol := params.Arguments.Symbol
	code, err := gophon.GetGolangSourceCode(params.Arguments.Namespace, symbol, params.Arguments.Receiver, params.Arguments.Name, params.Arguments.Tag)
	if err != nil && strings.Contains(err.Error(), gophon.NotFoundError.Error()) && symbol == "func" {
		return nil, fmt.Errorf("cannot find function %s, maybe it's a variable with function type?", symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get golang source code for %s %s: %w", symbol, params.Arguments.Name, err)
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: code,
			},
		},
	}, nil
}
