package gophon

// ListSupportedNamespaces returns all supported golang namespaces
func ListSupportedNamespaces() []string {
	namespaces := make([]string, 0, len(RemoteIndexMap))
	for k, _ := range RemoteIndexMap {
		namespaces = append(namespaces, k)
	}
	return namespaces
}
