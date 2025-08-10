package gophon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v74/github"
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
func readURLContent(owner string, repo string, path string, tag string) ([]byte, error) {
	githubClient := github.NewClient(&http.Client{})

	// Add GitHub token as Bearer authorization header if environment variable is set
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		githubClient = githubClient.WithAuthToken(token)
	}
	option := &github.RepositoryContentGetOptions{}
	if tag != "" {
		option.Ref = tag
	}
	fileContent, _, resp, err := githubClient.Repositories.GetContents(context.Background(), owner, repo, path, option)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", path, err)
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
		return nil, fmt.Errorf("HTTP request failed with status %d for URL: %s", resp.StatusCode, path)
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from URL %s: %w", path, err)
	}

	return []byte(content), nil
}

func GetTerraformSourceCode(blockType, terraformType, entrypointName, tag string) (string, error) {
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
	path := fmt.Sprintf("%s/%s/%s.json", "index", blockType, terraformType)

	// Use the helper function to read content from the URL
	content, err := readURLContent(remoteIndex.GitHubOwner, remoteIndex.GitHubRepo, path, tag)
	if err != nil {
		return "", fmt.Errorf("failed to read content from URL: %w", err)
	}

	index := make(map[string]string)
	if err = json.Unmarshal(content, &index); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON content from URL %s: %w", path, err)
	}
	entrypointName += "_index"
	entryPoint := index[entrypointName]
	namespace := index["namespace"]
	namespace = strings.TrimPrefix(namespace, remoteIndex.PackagePath)
	sourceCode, err := readURLContent(remoteIndex.GitHubOwner, remoteIndex.GitHubRepo, "index"+namespace+"/"+entryPoint, "")
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
