package tfschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	tfjson "github.com/hashicorp/terraform-json"
	aws_v6 "github.com/lonegunmanb/terraform-aws-schema/v6/generated"
	awscc "github.com/lonegunmanb/terraform-awscc-schema/generated"
	azuread_v3 "github.com/lonegunmanb/terraform-azuread-schema/v3/generated"
	azurerm_v4 "github.com/lonegunmanb/terraform-azurerm-schema/v4/generated"
	google_v6 "github.com/lonegunmanb/terraform-google-schema/v6/generated"
	"strings"
	"sync"
)

var resourceSchemas = make(map[string]*tfjson.Schema)
var dataSourceSchemas = make(map[string]*tfjson.Schema)
var ephemerals = make(map[string]*tfjson.Schema)
var ensureSchemas = sync.OnceFunc(initSchema)

func QuerySchema(category, name, path string) (string, error) {
	ensureSchemas()
	var schemas map[string]*tfjson.Schema
	switch category {
	case "resource":
		schemas = resourceSchemas
	case "data_source":
		schemas = dataSourceSchemas
	case "ephemeral":
		schemas = ephemerals
	default:
		return "", errors.New("unknown schema category, must be one of 'resource', 'data_source', or 'ephemeral'")
	}
	schema, ok := schemas[name]
	if !ok {
		return "", fmt.Errorf("schema %s %s not found", category, name)
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

func initSchema() {
	if len(resourceSchemas) == 0 {
		resources := []map[string]*tfjson.Schema{
			awscc.Resources,
			aws_v6.Resources,
			azurerm_v4.Resources,
			google_v6.Resources,
			azuread_v3.Resources,
		}
		for _, schemas := range resources {
			mergeSchemas(resourceSchemas, schemas)
		}
		dataSources := []map[string]*tfjson.Schema{
			awscc.DataSources,
			aws_v6.DataSources,
			azurerm_v4.DataSources,
			google_v6.DataSources,
			azuread_v3.DataSources,
		}
		for _, ds := range dataSources {
			mergeSchemas(dataSourceSchemas, ds)
		}
		ephemeralResources := []map[string]*tfjson.Schema{
			azurerm_v4.EphemeralResources,
			aws_v6.EphemeralResources,
			google_v6.EphemeralResources,
		}
		for _, e := range ephemeralResources {
			mergeSchemas(ephemerals, e)
		}
	}
}

func mergeSchemas(s1, s2 map[string]*tfjson.Schema) {
	for k, v := range s2 {
		s1[k] = v
	}
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
