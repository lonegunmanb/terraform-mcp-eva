package tool

import (
	"fmt"
)

// ListProviderItemsValidator handles validation of list provider items parameters
type ListProviderItemsValidator struct{}

// NewListProviderItemsValidator creates a new validator instance
func NewListProviderItemsValidator() *ListProviderItemsValidator {
	return &ListProviderItemsValidator{}
}

// ValidateParams validates the list provider items parameters
func (v *ListProviderItemsValidator) ValidateParams(category, namespace, providerName, version string) error {
	// Check if category is valid
	if _, ok := validCategories[category]; !ok {
		return fmt.Errorf("invalid category: %s", category)
	}

	// Provider name is always required
	if providerName == "" {
		return fmt.Errorf("provider name is required")
	}

	return nil
}

// NormalizeNamespace sets default namespace if empty
func (v *ListProviderItemsValidator) NormalizeNamespace(namespace string) string {
	if namespace == "" {
		return "hashicorp"
	}
	return namespace
}
