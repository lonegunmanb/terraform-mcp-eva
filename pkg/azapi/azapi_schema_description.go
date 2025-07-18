package azapi

import (
	"fmt"
	"github.com/lonegunmanb/newres/v3/pkg/azapi"
	"github.com/ms-henglu/go-azure-types/types"
	"strings"
)

func GetResourceSchemaDescription(resourceType, apiVersion, path string) (any, error) {
	apiType, err := azapi.GetAzApiType(resourceType, apiVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get azapi type for resource %s api-version %s: %w", resourceType, apiVersion, err)
	}
	bodyType, ok := apiType.Body.Type.(*types.ObjectType)
	if !ok {
		return nil, fmt.Errorf("resource body type is not an object type")
	}
	result := make(map[string]any)
	for n, p := range bodyType.Properties {
		desc, err := ConvertAzApiObjectPropertyToMap(p)
		if err != nil {
			return nil, fmt.Errorf("failed to convert property %s: %w", n, err)
		}
		result[n] = desc
	}
	result = map[string]any{
		"body": result,
	}
	if path == "" {
		return result, nil
	}
	return queryDescriptionInObject(result, path)
}

func queryDescriptionInObject(result map[string]any, path string) (any, error) {
	parts := strings.Split(path, ".")
	current := result

	for i, part := range parts {
		value, ok := current[part]
		if !ok {
			return nil, fmt.Errorf("property '%s' not found at path '%s'", part, strings.Join(parts[:i+1], "."))
		}

		// If this is the last part of the path, return the value
		if i == len(parts)-1 {
			return value, nil
		}

		// Otherwise, the value should be a map for further navigation
		nestedMap, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot navigate further: property '%s' is not an object", part)
		}
		current = nestedMap
	}

	return current, nil
}

// ConvertAzApiObjectPropertyToMap converts types.ObjectProperty to map[string]any
// where values are property descriptions, or nested maps for object properties
func ConvertAzApiObjectPropertyToMap(property types.ObjectProperty) (any, error) {
	objType, ok := property.Type.Type.(*types.ObjectType)
	if !ok {
		// If it's not an object type, return a simple map with description
		description := "[Description not available]"
		if property.Description != nil {
			description = *property.Description
		}

		// Append flag-based descriptions
		for _, flag := range property.Flags {
			switch flag {
			case types.WriteOnly:
				description += " (WriteOnly)"
			case types.Required:
				description += " (Required)"
			case types.ReadOnly:
				description += " (ReadOnly)"
			case types.Identifier:
				description += " (Identifier)"
			case types.DeployTimeConstant:
				description += " (DeployTimeConstant)"
			}
		}
		if possibleValues := getPossibleValues(property); len(possibleValues) > 0 {
			description += fmt.Sprintf(" (Possible values: %s)", strings.Join(possibleValues, ","))
		}

		return description, nil
	}

	return convertObjectTypeToMap(objType)
}

func getPossibleValues(property types.ObjectProperty) []string {
	if ut, ok := property.Type.Type.(*types.UnionType); ok {
		values := make([]string, 0, len(ut.Elements))
		for _, element := range ut.Elements {
			if literalType, ok := element.Type.(*types.StringLiteralType); ok {
				values = append(values, literalType.Value)
			}
		}
		return values
	}
	return nil
}

// convertObjectTypeToMap converts an ObjectType to map[string]any recursively
func convertObjectTypeToMap(objType *types.ObjectType) (map[string]any, error) {
	result := make(map[string]any)

	for name, prop := range objType.Properties {
		descs, err := ConvertAzApiObjectPropertyToMap(prop)
		if err != nil {
			return nil, err
		}
		result[name] = descs
	}

	return result, nil
}
