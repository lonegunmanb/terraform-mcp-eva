package gophon

import (
	"fmt"
)

// supportedNamespaces maps golang namespaces to their remote base URLs
var supportedNamespaces = map[string]string{
	"github.com/hashicorp/terraform-provider-azurerm/internal": "https://raw.githubusercontent.com/lonegunmanb/terraform-provider-azurerm-index/refs/heads/main/index/internal",
}

// ListSupportedNamespaces returns all supported golang namespaces
func ListSupportedNamespaces() []string {
	namespaces := make([]string, 0, len(supportedNamespaces))
	for namespace := range supportedNamespaces {
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

// GetNamespaceBaseURL returns the base URL for a given golang namespace
func GetNamespaceBaseURL(namespace string) (string, error) {
	url, exists := supportedNamespaces[namespace]
	if !exists {
		return "", fmt.Errorf("unsupported namespace: %s", namespace)
	}
	return url, nil
}
