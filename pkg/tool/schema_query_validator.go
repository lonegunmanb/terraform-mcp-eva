package tool

import (
	"fmt"
)

// SchemaQueryValidator handles validation of schema query parameters
type SchemaQueryValidator struct{}

// NewSchemaQueryValidator creates a new validator instance
func NewSchemaQueryValidator() *SchemaQueryValidator {
	return &SchemaQueryValidator{}
}

// ValidateParams validates the schema query parameters
func (v *SchemaQueryValidator) ValidateParams(category, resourceType, path, namespace, providerName string) error {
	// Check if category is valid
	if _, ok := validCategories[category]; !ok {
		return fmt.Errorf("invalid category: %s", category)
	}

	// For function category, name parameter is required
	if category == "function" && providerName == "" {
		return fmt.Errorf("provider name is required when category is 'function'")
	}

	// For provider category, name parameter is required
	if category == "provider" && providerName == "" {
		return fmt.Errorf("provider name is required when category is 'provider'")
	}

	// For function category, path queries are not supported
	if category == "function" && path != "" {
		return fmt.Errorf("path queries are not supported for %s schemas", category)
	}

	// For non-function and non-provider categories, validate provider name can be inferred if not provided
	if providerName == "" {
		inferredName := inferProviderNameFromType(resourceType)
		if inferredName == "" {
			return fmt.Errorf("could not infer provider name from type '%s', please provide the 'name' parameter", resourceType)
		}
	}

	return nil
}

// NormalizeNamespace sets default namespace if empty
func (v *SchemaQueryValidator) NormalizeNamespace(namespace string) string {
	if namespace == "" {
		return "hashicorp"
	}
	return namespace
}

// InferProviderName attempts to infer provider name from resource type if not provided
func (v *SchemaQueryValidator) InferProviderName(category, resourceType, providerName string) (string, error) {
	if providerName != "" {
		return providerName, nil
	}

	// For function and provider categories, name must be explicitly provided
	if category == "function" || category == "provider" {
		return "", fmt.Errorf("provider name is required when category is '%s'", category)
	}

	// Extract provider name from type (e.g., "aws_ec2_instance" -> "aws")
	inferredName := inferProviderNameFromType(resourceType)
	if inferredName == "" {
		return "", fmt.Errorf("could not infer provider name from type '%s', please provide the 'name' parameter", resourceType)
	}

	return inferredName, nil
}

var validCategories = map[string]struct{}{
	"resource":  {},
	"data":      {},
	"ephemeral": {},
	"function":  {},
	"provider":  {},
}
