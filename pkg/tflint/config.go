package tflint

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	hclmerge "github.com/lonegunmanb/hclmerge/pkg"
	"github.com/spf13/afero"
)

// Global filesystem interface for testing
var fs = afero.NewOsFs()

// Global HTTP download function for testing
var downloadConfigContent = func(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download config from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download config from %s: status %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read config content: %w", err)
	}

	return string(content), nil
}

// Global temp config dir setup function for testing
var setupTempConfigDir = func() (string, func(), error) {
	tempDir, err := afero.TempDir(fs, "", "tflint-config-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanup := func() {
		_ = fs.RemoveAll(tempDir)
	}

	return tempDir, cleanup, nil
}

// setupConfig sets up the complete TFLint configuration
func setupConfig(category, customConfigFile string) (*ConfigData, func(), error) {
	// Create temporary directory
	tempDir, tempCleanup, err := setupTempConfigDir()
	if err != nil {
		return nil, nil, err
	}

	// Get the appropriate config URL
	normalizedCategory := getDefaultCategory(category)
	configURL := getConfigURL(normalizedCategory)

	// Always download the base config first and save it to temp directory
	baseConfigContent, err := downloadConfigContent(configURL)
	if err != nil {
		return nil, tempCleanup, err
	}

	// Write base config to temp directory
	baseConfigPath := filepath.Join(tempDir, "base.tflint.hcl")
	err = afero.WriteFile(fs, baseConfigPath, []byte(baseConfigContent), 0644)
	if err != nil {
		return nil, tempCleanup, fmt.Errorf("failed to write base config file: %w", err)
	}
	finalConfigPath := baseConfigPath

	// If custom config file is provided, merge it with the base config
	if customConfigFile != "" {
		// Merge custom config with base config using hclmerge.MergeFile
		mergedConfigPath := filepath.Join(tempDir, "final.tflint.hcl")
		err = hclmerge.MergeFile(baseConfigPath, customConfigFile, mergedConfigPath)
		if err != nil {
			return nil, tempCleanup, fmt.Errorf("failed to merge custom config with base config: %w", err)
		}
		finalConfigPath = mergedConfigPath
	}

	config := &ConfigData{
		TempDir:    tempDir,
		ConfigPath: finalConfigPath,
		BaseURL:    configURL,
	}

	return config, tempCleanup, nil
}
