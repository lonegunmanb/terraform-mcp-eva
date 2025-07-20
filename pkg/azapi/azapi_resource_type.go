package azapi

import (
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	"github.com/ms-henglu/go-azure-types/types"
	"github.com/zclconf/go-cty/cty"
)

func GetResourceSchema(resourceType, apiVersion, path string) (string, error) {
	t, err := getResourceType(resourceType, apiVersion, path)
	if err != nil {
		return "", err
	}
	return compactGoType(t.GoString()), nil
}

func getResourceType(resourceType, apiVersion, path string) (cty.Type, error) {
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
	typeDesc, err := toCtyType(blockSchema)
	if err != nil {
		return cty.NilType, fmt.Errorf("failed to convert block schema to cty type: %w", err)
	}
	if path == "" {
		return typeDesc, nil
	}
	t, err := queryTypeInBlock(blockSchema, path)
	if err != nil {
		return typeDesc, fmt.Errorf("failed to query type in block for path %s: %w", path, err)
	}
	return t, nil
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
		attrTypes[name] = attr.AttributeType
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

func queryTypeInBlock(block *tfjson.SchemaBlock, path string) (cty.Type, error) {
	segments := strings.Split(path, ".")
	segment := segments[0]
	attribute, ok := block.Attributes[segment]
	if ok {
		if len(segments) == 1 {
			return attribute.AttributeType, nil
		}
		return queryTypeFromType(attribute.AttributeType, strings.Join(segments[1:], "."))
	}
	nb, ok := block.NestedBlocks[segment]
	if !ok {
		return cty.NilType, fmt.Errorf("type not found for path %s in block", path)
	}
	if len(segments) == 1 {
		return schemaBlockToCtyType(nb.Block), nil
	}
	return queryTypeFromType(schemaBlockToCtyType(nb.Block), strings.Join(segments[1:], "."))
}

func schemaBlockToCtyType(block *tfjson.SchemaBlock) cty.Type {
	attrTypes := make(map[string]cty.Type)

	// Add attributes
	for name, attr := range block.Attributes {
		attrTypes[name] = attr.AttributeType
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
