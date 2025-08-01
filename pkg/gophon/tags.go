package gophon

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

// ListSupportedTags returns all supported tags/versions for a given golang namespace
func ListSupportedTags(namespace string) ([]string, error) {
	// Get the remote index configuration for the namespace
	remoteIndex, exists := RemoteIndexMap[namespace]
	if !exists {
		return nil, fmt.Errorf("unsupported namespace: %s", namespace)
	}

	// Create GitHub client with authentication if token is available
	var client *github.Client
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		// Create authenticated client using OAuth2
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(context.Background(), ts)
		client = github.NewClient(tc)
	} else {
		// Create unauthenticated client for public repositories
		client = github.NewClient(nil)
	}

	// List all tags from the repository
	var allTags []string

	// Use pagination to get all tags
	opts := &github.ListOptions{PerPage: 100}
	for {
		tags, resp, err := client.Repositories.ListTags(context.Background(), remoteIndex.GitHubOwner, remoteIndex.GitHubRepo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list tags from GitHub repository %s/%s: %w",
				remoteIndex.GitHubOwner, remoteIndex.GitHubRepo, err)
		}

		// Extract tag names
		for _, tag := range tags {
			if tag.Name != nil {
				allTags = append(allTags, *tag.Name)
			}
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Sort tags in ascending order (lexicographic sorting)
	// Note: For proper semantic version sorting, we would need a library like semver
	// For now, lexicographic sorting will work reasonably well for most version tags
	sort.Strings(allTags)

	return allTags, nil
}
