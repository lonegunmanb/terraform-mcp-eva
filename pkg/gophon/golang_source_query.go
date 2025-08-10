package gophon

import (
	"fmt"
	"strings"
)

var validSymbols = map[string]struct{}{
	"func":   {},
	"method": {},
	"type":   {},
	"var":    {},
}

func GetGolangSourceCode(namespace, symbol, receiver, name, tag string) (string, error) {
	var remoteKey string
	for _, n := range Namespaces {
		if strings.HasPrefix(namespace, n) {
			remoteKey = n
			break
		}
	}
	if remoteKey == "" {
		return "", fmt.Errorf("unsupported namespace: %s", namespace)
	}
	if _, ok := validSymbols[symbol]; !ok {
		return "", fmt.Errorf("unsupported symbol: %s", symbol)
	}
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}
	if receiver != "" && symbol != "method" {
		return "", fmt.Errorf("receiver is only valid for methods")
	}
	remoteIndex := RemoteIndexMap[remoteKey]
	//baseUrl := strings.ReplaceAll(remoteIndex.BaseUrl, "{version}", version)
	namespace = strings.TrimPrefix(namespace, remoteIndex.PackagePath)
	path := fmt.Sprintf("%s%s/%s.%s.%s.goindex", "index", namespace, symbol, receiver, name)
	if receiver == "" {
		path = fmt.Sprintf("%s%s/%s.%s.goindex", "index", namespace, symbol, name)
	}
	content, err := readURLContent(remoteIndex.GitHubOwner, remoteIndex.GitHubRepo, path, tag)
	if err != nil {
		return "", fmt.Errorf("failed to read content from URL: %w", err)
	}
	return string(content), nil
}
