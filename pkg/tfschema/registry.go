package tfschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/matt-FFFFFF/tfpluginschema"
)

var serverInstance *tfpluginschema.Server
var serverOnce sync.Once

// ProviderRequest represents a request for a specific provider
type ProviderRequest struct {
	ProviderNamespace string `json:"namespace"`
	ProviderName      string `json:"name"`
	ProviderVersion   string `json:"version"`
}

func getServer() *tfpluginschema.Server {
	serverOnce.Do(func() {
		serverInstance = tfpluginschema.NewServer(nil)
	})
	return serverInstance
}

func QuerySchema(category, name, path string, providerReq ProviderRequest) (string, error) {
	server := getServer()

	request := tfpluginschema.Request{
		Namespace: providerReq.ProviderNamespace,
		Name:      providerReq.ProviderName,
		Version:   providerReq.ProviderVersion,
	}

	var schema *tfjson.Schema
	var functionSignature *tfjson.FunctionSignature
	var err error

	switch category {
	case "resource":
		schema, err = server.GetResourceSchema(request, name)
	case "data":
		schema, err = server.GetDataSourceSchema(request, name)
	case "ephemeral":
		schema, err = server.GetEphemeralResourceSchema(request, name)
	case "function":
		functionSignature, err = server.GetFunctionSchema(request, name)
	default:
		return "", errors.New("unknown schema category, must be one of 'resource', 'data', 'ephemeral', or 'function'")
	}

	if err != nil {
		return "", fmt.Errorf("failed to get %s schema for %s: %w", category, name, err)
	}

	// Handle function signatures differently from schemas
	if category == "function" {
		if path != "" {
			return "", errors.New("path queries are not supported for function schemas")
		}
		return toCompactJson(functionSignature)
	}

	if path == "" {
		return toCompactJson(schema)
	}

	// Query the specific path in the schema
	result, err := querySchemaPath(schema.Block, path)
	if err != nil {
		return "", fmt.Errorf("failed to query path %s in schema %s: %w", path, name, err)
	}
	return toCompactJson(result)
}

// ListItems lists available items (resources, data sources, ephemeral resources, or functions) for a provider
func ListItems(category string, providerReq ProviderRequest) ([]string, error) {
	server := getServer()

	request := tfpluginschema.Request{
		Namespace: providerReq.ProviderNamespace,
		Name:      providerReq.ProviderName,
		Version:   providerReq.ProviderVersion,
	}

	var items []string
	var err error
	switch category {
	case "resource":
		items, err = server.ListResources(request)
	case "data":
		items, err = server.ListDataSources(request)
	case "ephemeral":
		items, err = server.ListEphemeralResources(request)
	case "function":
		items, err = server.ListFunctions(request)
	default:
		return nil, errors.New("unknown category, must be one of 'resource', 'data', 'ephemeral', or 'function'")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list %s items for provider %s/%s: %w", category, providerReq.ProviderNamespace, providerReq.ProviderName, err)
	}

	return items, nil
}

// querySchemaPath traverses a schema block following the given dot-separated path
func querySchemaPath(block *tfjson.SchemaBlock, path string) (interface{}, error) {
	if path == "" {
		return block, nil
	}

	segments := strings.Split(path, ".")
	segment := segments[0]
	remainingPath := strings.Join(segments[1:], ".")

	// Check if the segment is an attribute
	if attr, ok := block.Attributes[segment]; ok {
		if remainingPath == "" {
			return attr, nil
		}
		// For attributes, we can't traverse further into the structure
		// since AttributeType is a cty.Type, not a schema block
		return nil, fmt.Errorf("cannot traverse into attribute %s - attributes don't have nested structure", segment)
	}

	// Check if the segment is a nested block
	if nestedBlock, ok := block.NestedBlocks[segment]; ok {
		if remainingPath == "" {
			return nestedBlock, nil
		}
		return querySchemaPath(nestedBlock.Block, remainingPath)
	}

	return nil, fmt.Errorf("path segment '%s' not found in schema block", segment)
}

func toCompactJson(data interface{}) (string, error) {
	marshal, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %+v", err)
	}
	dst := &bytes.Buffer{}
	if err = json.Compact(dst, marshal); err != nil {
		return "", fmt.Errorf("failed to compact data: %+v", err)
	}
	return dst.String(), nil
}
