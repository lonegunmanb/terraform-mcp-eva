package tflint

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Global filesystem interface for testing
var fs = afero.NewOsFs()

// downloadConfigContent now uses go-getter for all remote config downloads
var downloadConfigContent = func(url string) (string, error) {
	// Create temporary directory for download
	tempDir, err := afero.TempDir(fs, "", "tflint-download-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer fs.RemoveAll(tempDir)

	// Use go-getter to download the file directly (timeout handled in getter)
	configFile := filepath.Join(tempDir, "config.hcl")
	if err := remoteConfigGetter.Get(configFile, url); err != nil {
		return "", fmt.Errorf("failed to download config from %s: %w", url, err)
	}

	// Read the downloaded file
	content, err := afero.ReadFile(fs, configFile)
	if err != nil {
		return "", fmt.Errorf("failed to read downloaded config: %w", err)
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
func setupConfig(category string) (*ConfigData, func(), error) {
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
	return createConfigData(tempDir, baseConfigPath, configURL), tempCleanup, nil
}

// setupRemoteConfig sets up configuration when a remote_config_url is provided.
// Downloads the remote config file directly to the temp directory as remote.tflint.hcl.
func setupRemoteConfig(remoteURL string) (*ConfigData, func(), error) {
	// Create temporary directory first
	tempDir, tempCleanup, err := setupTempConfigDir()
	if err != nil {
		return nil, nil, err
	}

	// Preliminary git root heuristic: if git:: and no .git// treat as root (must specify file path)
	if strings.HasPrefix(remoteURL, "git::") {
		core := remoteURL
		if i := strings.Index(core, "?"); i >= 0 {
			core = core[:i]
		}
		if !strings.Contains(core, ".git//") {
			return nil, tempCleanup, fmt.Errorf("remote_config_url must point to a single file (git repository root detected): %s", remoteURL)
		}
	}

	// Remote getter downloads directly to specified file path (timeout handled in getter)
	baseConfigPath := filepath.Join(tempDir, "remote.tflint.hcl")
	if err := remoteConfigGetter.Get(baseConfigPath, remoteURL); err != nil {
		return nil, tempCleanup, fmt.Errorf("failed to fetch remote config: %w", err)
	}

	// File is now at the expected location
	return createConfigData(tempDir, baseConfigPath, remoteURL), tempCleanup, nil
}

// mergeOptionalCustomConfig merges customConfigFile into baseConfigPath if provided, returning final path.
// mergeOptionalCustomConfig removed after simplification: custom config merging no longer supported.

// createConfigData centralizes creation of ConfigData to avoid duplication
func createConfigData(tempDir, configPath, baseURL string) *ConfigData {
	return &ConfigData{
		TempDir:    tempDir,
		ConfigPath: configPath,
	}
}
