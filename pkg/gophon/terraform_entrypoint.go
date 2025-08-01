package gophon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var validEntrypoints = map[string]map[string]struct{}{
	"resource": {
		"create":    {},
		"read":      {},
		"update":    {},
		"delete":    {},
		"schema":    {},
		"attribute": {},
	},
	"data": {
		"read":      {},
		"schema":    {},
		"attribute": {},
	},
	"ephemeral": {
		"open":   {},
		"close":  {},
		"renew":  {},
		"schema": {},
	},
}

var NotFoundError = errors.New("source code not found (404)")

// readURLContent reads content from a URL and returns it as []byte
func readURLContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error if needed, but don't override the main error
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFoundError
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d for URL: %s", resp.StatusCode, url)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from URL %s: %w", url, err)
	}

	return content, nil
}

func GetTerraformSourceCode(blockType, terraformType, entrypointName, tag string) (string, error) {
	tag = formatVersion(tag)
	entryPoints, ok := validEntrypoints[blockType]
	if !ok {
		return "", fmt.Errorf("invalid block type: %s", blockType)
	}
	if _, ok := entryPoints[entrypointName]; !ok {
		return "", fmt.Errorf("invalid entrypoint name: %s for block type: %s", entrypointName, blockType)
	}
	segments := strings.Split(terraformType, "_")
	if len(segments) < 2 {
		return "", fmt.Errorf("invalid terraform type: %s, valid terraform type should be like `azurerm_resource_group`", terraformType)
	}
	providerType := segments[0]
	indexKey, ok := ProviderIndexMap[providerType]
	if !ok {
		return "", fmt.Errorf("unsupported provider type: %s, supported providers are: %v", providerType, GetSupportedProviders())
	}
	remoteIndex := RemoteIndexMap[indexKey]
	if blockType != "ephemeral" {
		blockType += "s"
	}
	baseUrl := strings.ReplaceAll(remoteIndex.BaseUrl, "{version}", tag)
	url := fmt.Sprintf("%s/%s/%s.json", baseUrl, blockType, terraformType)

	// Use the helper function to read content from the URL
	content, err := readURLContent(url)
	if err != nil {
		return "", fmt.Errorf("failed to read content from URL: %w", err)
	}

	index := make(map[string]string)
	if err = json.Unmarshal(content, &index); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON content from URL %s: %w", url, err)
	}
	entrypointName += "_index"
	entryPoint := index[entrypointName]
	namespace := index["namespace"]
	namespace = strings.TrimPrefix(namespace, remoteIndex.PackagePath)
	sourceCode, err := readURLContent(baseUrl + namespace + "/" + entryPoint)
	if err != nil {
		return "", err
	}
	return string(sourceCode), nil
}

func formatVersion(tag string) string {
	if tag == "" {
		tag = "heads/main"
	} else {
		tag = "tags/" + tag
	}
	return tag
}
