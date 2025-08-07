package tflint

import (
	"os"
	"path/filepath"
	"testing"

	hclmerge "github.com/lonegunmanb/hclmerge/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTempConfigDir(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "should create temporary directory successfully",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use memory filesystem for testing
			memFs := afero.NewMemMapFs()
			stubs := gostub.Stub(&fs, memFs)
			defer stubs.Reset()

			tempDir, cleanup, err := setupTempConfigDir()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			defer cleanup()

			// Verify directory exists and is accessible
			assert.True(t, filepath.IsAbs(tempDir))

			info, err := fs.Stat(tempDir)
			require.NoError(t, err)
			assert.True(t, info.IsDir())

			// Verify we can write to the directory
			testFile := filepath.Join(tempDir, "test.txt")
			err = afero.WriteFile(fs, testFile, []byte("test"), 0644)
			assert.NoError(t, err)

			// Cleanup should remove the directory
			cleanup()
			_, err = fs.Stat(tempDir)
			assert.True(t, os.IsNotExist(err))
		})
	}
}

func TestSetupConfigWithMemoryFs(t *testing.T) {
	tests := []struct {
		name          string
		category      string
		customConfig  string
		configContent string
		wantErr       bool
	}{
		{
			name:          "should setup config for reusable category with memory fs",
			category:      "reusable",
			configContent: `rule "terraform_deprecated_syntax" { enabled = true }`,
			wantErr:       false,
		},
		{
			name:          "should setup config with custom config file",
			category:      "reusable",
			customConfig:  "/test/custom.hcl",
			configContent: `# Custom config`,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use memory filesystem for testing
			memFs := afero.NewMemMapFs()
			stubs := gostub.Stub(&fs, memFs).Stub(&hclmerge.Fs, memFs)
			defer stubs.Reset()

			// Create a mock custom config file if specified
			if tt.customConfig != "" {
				err := afero.WriteFile(fs, tt.customConfig, []byte(tt.configContent), 0644)
				require.NoError(t, err)
			}

			// Mock the download function to return test content
			originalDownload := downloadConfigContent
			downloadConfigContent = func(url string) (string, error) {
				return tt.configContent, nil
			}
			defer func() { downloadConfigContent = originalDownload }()

			config, cleanup, err := setupConfig(tt.category, tt.customConfig)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			defer cleanup()

			// Verify config structure
			assert.NotEmpty(t, config.TempDir)
			assert.NotEmpty(t, config.ConfigPath)
			assert.NotEmpty(t, config.BaseURL)

			// Verify config file exists
			_, err = fs.Stat(config.ConfigPath)
			assert.NoError(t, err)

			// Verify temp directory exists
			info, err := fs.Stat(config.TempDir)
			assert.NoError(t, err)
			assert.True(t, info.IsDir())
		})
	}
}
