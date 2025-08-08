package tool

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/azapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type AzAPIVersionQueryParam struct {
	ResourceType string `json:"resource_type" jsonschema:"Azure resource type, for example: Microsoft.Compute/virtualMachines""`
}

func QueryAzAPIVersions(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[AzAPIVersionQueryParam]) (*mcp.CallToolResultFor[any], error) {
	resourceType := params.Arguments.ResourceType
	if resourceType == "" {
		return nil, errors.New("`resource_type` are required parameters")
	}

	versions, err := azapi.GetApiVersions(resourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions for %s: %w", resourceType, err)
	}
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("[%s]", strings.Join(versions, ",")),
			},
		},
	}, nil
}
