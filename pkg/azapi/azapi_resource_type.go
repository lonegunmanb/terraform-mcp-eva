package azapi

import (
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	azapi_resource "github.com/lonegunmanb/terraform-azapi-schema/v2/generated"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

func GetResourceSchema(resourceType, apiVersion, path string) (string, error) {
	t, err := getSwaggerResourceType(resourceType, apiVersion)
	if err != nil {
		return "", err
	}
	schema := azapi_resource.Resources["azapi_resource"]
	schemaType, err := toCtyType(schema.Block)
	if err != nil {
		return "", fmt.Errorf("failed to convert azapi resource schema to cty type: %w", err)
	}
	attributeTypes := schemaType.AttributeTypes()
	for n, at := range t.AttributeTypes() {
		attributeTypes[n] = at
	}
	mergedType := cty.Object(attributeTypes)

	if path == "" {
		return compactGoType(mergedType.GoString()), nil
	}
	subType, err := queryTypeFromType(mergedType, path)
	if err != nil {
		return "", fmt.Errorf("failed to query type from path %s: %w", path, err)
	}
	return compactGoType(subType.GoString()), nil
}

func getSwaggerResourceType(resourceType, apiVersion string) (cty.Type, error) {
	apiType, err := azapi.GetAzApiType(resourceType, apiVersion)
	if err != nil {
		return cty.NilType, fmt.Errorf("failed to get azapi type for resource %s api-version %s: %w", resourceType, apiVersion, err)
	}
	bodyType, ok := apiType.Body.Type.(*types.ObjectType)
	if !ok {
		return cty.NilType, fmt.Errorf("resource body type is not an object type")
	}
	blockSchema, err := azapi.ConvertAzApiObjectTypeToTerraformJsonSchemaAttribute(types.ObjectProperty{
		Type: &types.TypeReference{
			Type: bodyType,
		},
	})
	if err != nil {
		return cty.NilType, fmt.Errorf("failed to convert az api object type to terraform json schema: %w", err)
	}
	return toCtyType(blockSchema)
}

func compactGoType(goType string) string {
	return strings.ReplaceAll(goType, "cty.", "")
}

func toCtyType(block *tfjson.SchemaBlock) (cty.Type, error) {
	if block == nil {
		return cty.NilType, fmt.Errorf("block is nil")
	}
	attrTypes := make(map[string]cty.Type)

	// Add attributes
	for name, attr := range block.Attributes {
		at := attr.AttributeType
		if at == cty.NilType && attr.AttributeNestedType != nil {
			nestedCtyType, err := attributeNestedTypeToCtyType(attr.AttributeNestedType)
			if err != nil {
				return cty.NilType, fmt.Errorf("failed to convert nested attribute type for %s: %w", name, err)
			}
			at = nestedCtyType
		}
		attrTypes[name] = at
	}

	// Add nested blocks as object types
	for name, nestedBlock := range block.NestedBlocks {
		blockType := schemaBlockToCtyType(nestedBlock.Block)
		switch nestedBlock.NestingMode {
		case tfjson.SchemaNestingModeSingle:
			attrTypes[name] = blockType
		case tfjson.SchemaNestingModeList:
			attrTypes[name] = cty.List(blockType)
		case tfjson.SchemaNestingModeSet:
			attrTypes[name] = cty.Set(blockType)
		case tfjson.SchemaNestingModeMap:
			attrTypes[name] = cty.Map(blockType)
		default:
			attrTypes[name] = blockType
		}
	}

	return cty.Object(attrTypes), nil
}

func schemaBlockToCtyType(block *tfjson.SchemaBlock) cty.Type {
	attrTypes := make(map[string]cty.Type)

	// Add attributes
	for name, attr := range block.Attributes {
		at := attr.AttributeType
		if at == cty.NilType && attr.AttributeNestedType != nil {
			nestedCtyType, _ := attributeNestedTypeToCtyType(attr.AttributeNestedType)
			at = nestedCtyType
		}
		attrTypes[name] = at
	}

	// Add nested blocks as object types
	for name, nestedBlock := range block.NestedBlocks {
		blockType := schemaBlockToCtyType(nestedBlock.Block)
		switch nestedBlock.NestingMode {
		case tfjson.SchemaNestingModeSingle:
			attrTypes[name] = blockType
		case tfjson.SchemaNestingModeList:
			attrTypes[name] = cty.List(blockType)
		case tfjson.SchemaNestingModeSet:
			attrTypes[name] = cty.Set(blockType)
		case tfjson.SchemaNestingModeMap:
			attrTypes[name] = cty.Map(blockType)
		default:
			attrTypes[name] = blockType
		}
	}

	return cty.Object(attrTypes)
}

func queryTypeFromType(t cty.Type, path string) (cty.Type, error) {
	segments := strings.Split(path, ".")
	if len(segments) == 0 {
		return t, nil
	}
	segment := segments[0]
	if t.IsObjectType() {
		objType := t
		if attrType, ok := objType.AttributeTypes()[segment]; ok {
			if len(segments) == 1 {
				return attrType, nil
			}
			return queryTypeFromType(attrType, strings.Join(segments[1:], "."))
		}
	} else if t.IsMapType() || t.IsListType() || t.IsSetType() {
		mapType := t.ElementType()
		if attrType, ok := mapType.AttributeTypes()[segment]; ok {
			if len(segments) == 1 {
				return attrType, nil
			}
			return queryTypeFromType(attrType, strings.Join(segments[1:], "."))
		}
	}
	return cty.NilType, fmt.Errorf("type not found for path %s in type %s", path, t.FriendlyName())
}

func queryTypeFromAttributeTypes(attributeTypes map[string]cty.Type, path string) (cty.Type, error) {
	if path == "" {
		return cty.NilType, fmt.Errorf("empty path")
	}

	segments := strings.Split(path, ".")
	segment := segments[0]

	attrType, ok := attributeTypes[segment]
	if !ok {
		return cty.NilType, fmt.Errorf("type not found for path %s in attributeTypes", path)
	}

	if len(segments) == 1 {
		return attrType, nil
	}

	// Continue querying the nested path
	return queryTypeFromType(attrType, strings.Join(segments[1:], "."))
}

func attributeNestedTypeToCtyType(nestedType *tfjson.SchemaNestedAttributeType) (cty.Type, error) {
	if nestedType == nil {
		return cty.NilType, fmt.Errorf("nested type is nil")
	}

	attrTypes := make(map[string]cty.Type)

	// Add attributes from nested type
	for name, attr := range nestedType.Attributes {
		at := attr.AttributeType
		if at == cty.NilType && attr.AttributeNestedType != nil {
			nestedCtyType, err := attributeNestedTypeToCtyType(attr.AttributeNestedType)
			if err != nil {
				return cty.NilType, fmt.Errorf("failed to convert nested attribute type for %s: %w", name, err)
			}
			at = nestedCtyType
		}
		attrTypes[name] = at
	}

	objType := cty.Object(attrTypes)

	// Handle nesting mode
	switch nestedType.NestingMode {
	case tfjson.SchemaNestingModeSingle:
		return objType, nil
	case tfjson.SchemaNestingModeList:
		return cty.List(objType), nil
	case tfjson.SchemaNestingModeSet:
		return cty.Set(objType), nil
	case tfjson.SchemaNestingModeMap:
		return cty.Map(objType), nil
	default:
		return objType, nil
	}
}
