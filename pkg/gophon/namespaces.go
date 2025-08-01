package gophon

import (
	"fmt"
)

// ListSupportedNamespaces returns all supported golang namespaces
func ListSupportedNamespaces() []string {
	namespaces := make([]string, 0, len(RemoteIndexMap))
	for k, _ := range RemoteIndexMap {
		namespaces = append(namespaces, k)
	}
	return namespaces
}

// GetNamespaceBaseURL returns the base URL for a given golang namespace
func GetNamespaceBaseURL(namespace string) (string, error) {
	index, exists := RemoteIndexMap[namespace]
	if !exists {
		return "", fmt.Errorf("unsupported namespace: %s", namespace)
	}
	return index.BaseUrl, nil
}
