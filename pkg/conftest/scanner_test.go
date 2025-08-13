package conftest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCommandExecutor for testing command execution
type MockCommandExecutor struct {
	commands map[string]*MockCommandResult
	patterns map[string]*MockCommandResult // For pattern-based matching
}

type MockCommandResult struct {
	stdout string
	stderr string
	err    error
}

func (m *MockCommandExecutor) ExecuteCommand(dir, command string) (string, string, error) {
	// First try exact match
	result, exists := m.commands[command]
	if exists {
		return result.stdout, result.stderr, result.err
	}

	// Then try pattern-based matching
	if m.patterns != nil {
		for pattern, result := range m.patterns {
			if strings.Contains(command, pattern) {
				return result.stdout, result.stderr, result.err
			}
		}
	}

	return "", "", assert.AnError
}

// MockPolicyDownloader for testing policy downloading
type MockPolicyDownloader struct {
	downloads        map[string]*MockDownloadResult
	setupPolicyFiles func(fs afero.Fs, destDir string, url string) error
}

type MockDownloadResult struct {
	err error
}

func (m *MockPolicyDownloader) DownloadPolicy(url, destDir string) error {
	// Check if there's a specific result for this URL
	if m.downloads != nil {
		if result, exists := m.downloads[url]; exists {
			if result.err != nil {
				return result.err
			}
		}
	}

	// If no error, setup mock policy files
	if m.setupPolicyFiles != nil {
		return m.setupPolicyFiles(fs, destDir, url)
	}

	// Default behavior - create some mock .rego files
	return m.createDefaultMockPolicies(destDir)
}

func (m *MockPolicyDownloader) createDefaultMockPolicies(destDir string) error {
	// Create some default mock policy files
	policyContent := `package main

import rego.v1

deny[msg] {
	msg := "Test policy violation"
}`

	return afero.WriteFile(fs, destDir+"/test_policy.rego", []byte(policyContent), 0644)
}

func TestValidateTargetFile(t *testing.T) {
	tests := []struct {
		name      string
		setupFs   func(fs afero.Fs) string // returns target file path
		wantErr   bool
		errString string
	}{
		{
			name: "should validate existing target file",
			setupFs: func(fs afero.Fs) string {
				planFile := "/test/plan.json"
				require.NoError(t, fs.MkdirAll("/test", 0755))
				require.NoError(t, afero.WriteFile(fs, planFile, []byte(`{"terraform_version": "1.0.0"}`), 0644))
				return planFile
			},
			wantErr: false,
		},
		{
			name: "should error on non-existent target file",
			setupFs: func(fs afero.Fs) string {
				return "/test/nonexistent.json"
			},
			wantErr:   true,
			errString: "target file does not exist",
		},
		{
			name: "should error when path is a directory, not file",
			setupFs: func(fs afero.Fs) string {
				dir := "/test/plan"
				require.NoError(t, fs.MkdirAll(dir, 0755))
				return dir
			},
			wantErr:   true,
			errString: "target path is not a file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup memory filesystem
			memFs := afero.NewMemMapFs()
			stubs := gostub.Stub(&fs, memFs)
			defer stubs.Reset()

			planPath := tt.setupFs(fs)

			// Execute
			err := validateTargetFile(planPath)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildConftestCommand(t *testing.T) {
	tests := []struct {
		name          string
		planFile      string
		policySources []PolicySource
		namespaces    []string
		expectedCmd   string
		shouldContain []string
	}{
		{
			name:     "should build basic conftest command",
			planFile: "/test/plan.json",
			policySources: []PolicySource{
				{ResolvedPath: "/tmp/policies1"},
			},
			expectedCmd: "conftest test --no-color -o json --all-namespaces -p /tmp/policies1 /test/plan.json",
		},
		{
			name:     "should build command with multiple policy sources",
			planFile: "/test/plan.json",
			policySources: []PolicySource{
				{ResolvedPath: "/tmp/policies1"},
				{ResolvedPath: "/tmp/policies2"},
			},
			shouldContain: []string{
				"conftest test",
				"--no-color",
				"-o json",
				"--all-namespaces",
				"-p /tmp/policies1",
				"-p /tmp/policies2",
				"/test/plan.json",
			},
		},
		{
			name:     "should build command with namespace filtering",
			planFile: "/test/plan.json",
			policySources: []PolicySource{
				{ResolvedPath: "/tmp/policies1"},
			},
			namespaces: []string{"main", "security"},
			shouldContain: []string{
				"conftest test",
				"--no-color",
				"-o json",
				"--namespace main",
				"--namespace security",
				"-p /tmp/policies1",
				"/test/plan.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			cmd := buildConftestCommand(tt.planFile, tt.policySources, tt.namespaces)

			// Assert
			if tt.expectedCmd != "" {
				assert.Equal(t, tt.expectedCmd, cmd)
			}
			if tt.shouldContain != nil {
				for _, expected := range tt.shouldContain {
					assert.Contains(t, cmd, expected)
				}
			}
		})
	}
}

func TestExecuteConftestScan(t *testing.T) {
	tests := []struct {
		name        string
		planFile    string
		command     string
		expectError bool
		mockStdout  string
		mockStderr  string
		mockErr     error
	}{
		{
			name:     "should execute conftest scan successfully",
			planFile: "/test/plan.json",
			command:  "conftest test --all-namespaces -p /tmp/policies /test/plan.json",
			mockStdout: `{
				"warnings": [],
				"failures": [],
				"successes": 5
			}`,
			expectError: false,
		},
		{
			name:     "should execute conftest scan with violations",
			planFile: "/test/plan.json",
			command:  "conftest test --all-namespaces -p /tmp/policies /test/plan.json",
			mockStdout: `{
				"warnings": [],
				"failures": [
					{
						"filename": "/test/plan.json",
						"namespace": "main.storage_account_https_only",
						"successes": 0,
						"failures": [
							{
								"msg": "Storage account should enforce HTTPS"
							}
						]
					}
				],
				"successes": 4
			}`,
			expectError: false,
		},
		{
			name:        "should handle conftest command failure",
			planFile:    "/test/plan.json",
			command:     "conftest test --all-namespaces -p /tmp/policies /test/plan.json",
			mockStderr:  "conftest: command not found",
			mockErr:     assert.AnError,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the command executor
			mockExecutor := &MockCommandExecutor{
				commands: map[string]*MockCommandResult{
					tt.command: {
						stdout: tt.mockStdout,
						stderr: tt.mockStderr,
						err:    tt.mockErr,
					},
				},
			}

			stubs := gostub.Stub(&commandExecutor, mockExecutor)
			defer stubs.Reset()

			// Execute
			output, err := executeConftestScan("", tt.command)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockStdout, output)
			}
		})
	}
}

func TestParseConftestOutput(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedViolations int
		expectedWarnings   int
		expectError        bool
	}{
		{
			name: "should parse clean output with no violations",
			input: `[
				{
					"filename": "/test/plan.json",
					"namespace": "main",
					"successes": 5,
					"failures": [],
					"warnings": []
				}
			]`,
			expectedViolations: 0,
			expectedWarnings:   0,
			expectError:        false,
		},
		{
			name: "should parse output with violations",
			input: `[
				{
					"filename": "/test/plan.json",
					"namespace": "main",
					"successes": 0,
					"failures": [
						{
							"msg": "Storage account should enforce HTTPS"
						}
					],
					"warnings": []
				},
				{
					"filename": "/test/plan.json", 
					"namespace": "security",
					"successes": 0,
					"failures": [
						{
							"msg": "VM should have backup enabled"
						}
					],
					"warnings": []
				}
			]`,
			expectedViolations: 2,
			expectedWarnings:   0,
			expectError:        false,
		},
		{
			name: "should parse output with warnings",
			input: `[
				{
					"filename": "/test/plan.json",
					"namespace": "main",
					"successes": 4,
					"failures": [],
					"warnings": [
						{
							"msg": "Consider using premium storage"
						}
					]
				}
			]`,
			expectedViolations: 0,
			expectedWarnings:   1,
			expectError:        false,
		},
		{
			name:        "should handle malformed JSON",
			input:       `{"invalid": json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			violations, warnings, err := parseConftestOutput(tt.input)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, violations, tt.expectedViolations)
				assert.Len(t, warnings, tt.expectedWarnings)
			}
		})
	}
}

func TestScan_EndToEnd(t *testing.T) {
	tests := []struct {
		name                   string
		setupFs                func(fs afero.Fs)
		param                  ScanParam
		setupCommands          func(mock *MockCommandExecutor)
		expectError            bool
		expectedViolationCount int
	}{
		{
			name: "should perform successful scan with no violations",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/test", 0755))
				require.NoError(t, afero.WriteFile(fs, "/test/plan.json", []byte(`{"terraform_version": "1.0.0"}`), 0644))
			},
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				TargetFile:                   "/test/plan.json",
			},
			setupCommands: func(mock *MockCommandExecutor) {
				mock.patterns = map[string]*MockCommandResult{
					"conftest test": {
						stdout: `[
							{
								"filename": "/test/plan.json",
								"namespace": "main",
								"successes": 5,
								"failures": [],
								"warnings": []
							}
						]`,
						stderr: "",
						err:    nil,
					},
				}
			},
			expectError:            false,
			expectedViolationCount: 0,
		},
		{
			name: "should perform scan with violations",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/test", 0755))
				require.NoError(t, afero.WriteFile(fs, "/test/plan.json", []byte(`{"terraform_version": "1.0.0"}`), 0644))
			},
			param: ScanParam{
				PolicyUrls: []string{"git::https://example.com/policies.git"},
				TargetFile: "/test/plan.json",
			},
			setupCommands: func(mock *MockCommandExecutor) {
				mock.patterns = map[string]*MockCommandResult{
					"conftest test": {
						stdout: `[
							{
								"filename": "/test/plan.json",
								"namespace": "main",
								"successes": 0,
								"failures": [
									{
										"msg": "Storage account should enforce HTTPS"
									}
								],
								"warnings": []
							}
						]`,
						stderr: "",
						err:    nil,
					},
				}
			},
			expectError:            false,
			expectedViolationCount: 1,
		},
		{
			name: "should handle validation errors",
			setupFs: func(fs afero.Fs) {
				// Don't create the target file
			},
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				TargetFile:                   "/test/nonexistent.json",
			},
			setupCommands: func(mock *MockCommandExecutor) {},
			expectError:   true,
		},
		{
			name: "should cleanup temporary directories after scan with ignored policies",
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/test", 0755))
				require.NoError(t, afero.WriteFile(fs, "/test/plan.json", []byte(`{"terraform_version": "1.0.0"}`), 0644))
			},
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				TargetFile:                   "/test/plan.json",
				IgnoredPolicies: []IgnoredPolicy{
					{Namespace: "avmsec", Name: "storage_account_https_only"},
				},
			},
			setupCommands: func(mock *MockCommandExecutor) {
				mock.patterns = map[string]*MockCommandResult{
					"conftest test": {
						stdout: `[
							{
								"filename": "/test/plan.json",
								"namespace": "main",
								"successes": 5,
								"failures": [],
								"warnings": []
							}
						]`,
						stderr: "",
						err:    nil,
					},
				}
			},
			expectError:            false,
			expectedViolationCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup memory filesystem
			memFs := afero.NewMemMapFs()
			stubs := gostub.Stub(&fs, memFs)
			defer stubs.Reset()

			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			// Mock command executor
			mockExecutor := &MockCommandExecutor{
				commands: make(map[string]*MockCommandResult),
			}
			if tt.setupCommands != nil {
				tt.setupCommands(mockExecutor)
			}

			commandStubs := gostub.Stub(&commandExecutor, mockExecutor)
			defer commandStubs.Reset()

			// Mock policy downloader
			mockDownloader := &MockPolicyDownloader{
				downloads: make(map[string]*MockDownloadResult),
				setupPolicyFiles: func(fs afero.Fs, destDir string, url string) error {
					// Create mock policy files for any URL
					require.NoError(t, fs.MkdirAll(destDir, 0755))
					policyContent := `package main

import rego.v1

deny[msg] {
	msg := "Test policy from ` + url + `"
}`
					return afero.WriteFile(fs, destDir+"/mock_policy.rego", []byte(policyContent), 0644)
				},
			}

			downloaderStubs := gostub.Stub(&policyDownloader, mockDownloader)
			defer downloaderStubs.Reset()

			// Execute
			result, err := Scan(tt.param)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Violations, tt.expectedViolationCount)
				assert.Equal(t, tt.param.TargetFile, result.TargetFile)

				// For cleanup test, verify no temporary directories remain
				if tt.name == "should cleanup temporary directories after scan with ignored policies" {
					// Verify temp directory count is back to original
					// Since we're using a memory filesystem, we can check the root directory
					tempDirsAfter, err := afero.ReadDir(fs, "/")
					require.NoError(t, err)

					// Count directories that start with "conftest-scan-" pattern
					confTestDirCount := 0
					for _, dir := range tempDirsAfter {
						if dir.IsDir() && strings.HasPrefix(dir.Name(), "conftest-scan-") {
							confTestDirCount++
						}
					}

					assert.Equal(t, 0, confTestDirCount, "All temporary conftest-scan-* directories should be cleaned up")
				}
			}
		})
	}
}

func TestCreateIgnoreConfig(t *testing.T) {
	tests := []struct {
		name            string
		ignoredPolicies []IgnoredPolicy
		setupFs         func(fs afero.Fs)
		wantErr         bool
		expectedFiles   int
		expectedContent map[string][]string // filename -> expected content lines
	}{
		{
			name:            "should handle empty ignored policies",
			ignoredPolicies: []IgnoredPolicy{},
			setupFs:         func(fs afero.Fs) {},
			wantErr:         false,
			expectedFiles:   0,
		},
		{
			name: "should create ignore config for single namespace",
			ignoredPolicies: []IgnoredPolicy{
				{Namespace: "avmsec", Name: "storage_account_https_only"},
				{Namespace: "avmsec", Name: "vm_backup_enabled"},
			},
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/tmp", 0755))
			},
			wantErr:       false,
			expectedFiles: 1,
			expectedContent: map[string][]string{
				"exceptions_avmsec/exceptions_avmsec.rego": {
					"package avmsec",
					"import rego.v1",
					"exception contains rules if {",
					`rules = ["storage_account_https_only", "vm_backup_enabled"]`,
					"}",
				},
			},
		},
		{
			name: "should create ignore config for multiple namespaces",
			ignoredPolicies: []IgnoredPolicy{
				{Namespace: "avmsec", Name: "storage_account_https_only"},
				{Namespace: "aprl", Name: "vm_backup_enabled"},
			},
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/tmp", 0755))
			},
			wantErr:       false,
			expectedFiles: 2, // Two separate namespace directories
			expectedContent: map[string][]string{
				"exceptions_avmsec/exceptions_avmsec.rego": {
					"package avmsec",
					"import rego.v1",
					"exception contains rules if {",
					`rules = ["storage_account_https_only"]`,
					"}",
				},
				"exceptions_aprl/exceptions_aprl.rego": {
					"package aprl",
					"import rego.v1",
					"exception contains rules if {",
					`rules = ["vm_backup_enabled"]`,
					"}",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup memory filesystem
			memFs := afero.NewMemMapFs()
			stubs := gostub.Stub(&fs, memFs)
			defer stubs.Reset()

			if tt.setupFs != nil {
				tt.setupFs(fs)
			}

			// Execute
			paths, err := createIgnoreConfig(tt.ignoredPolicies, "/tmp")

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, paths, tt.expectedFiles)

				// Verify file content if expected
				if tt.expectedContent != nil {
					for filename, expectedLines := range tt.expectedContent {
						filePath := "/tmp/" + filename

						// Check that file exists
						exists, err := afero.Exists(fs, filePath)
						require.NoError(t, err)
						assert.True(t, exists, "Expected file %s to exist", filename)

						// Read and verify content
						content, err := afero.ReadFile(fs, filePath)
						require.NoError(t, err)

						contentStr := string(content)
						for _, expectedLine := range expectedLines {
							assert.Contains(t, contentStr, expectedLine,
								"File %s should contain line: %s", filename, expectedLine)
						}
					}
				}
			}
		})
	}
}

// TestDownloadPolicy_Integration tests the actual policy downloading functionality
// This is an integration test that requires network access and downloads real policies
func TestDownloadPolicy_Integration(t *testing.T) {
	// Skip this test in short mode or if running in CI without network access
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use real filesystem for this integration test
	realFs := afero.NewOsFs()
	stubs := gostub.Stub(&fs, realFs)
	defer stubs.Reset()

	// Create a temporary directory for the test
	tempDir, err := afero.TempDir(fs, "", "conftest-integration-test")
	require.NoError(t, err)
	defer func() {
		_ = fs.RemoveAll(tempDir) // Cleanup after test
	}()

	// Create real policy downloader instance
	downloader := &RealPolicyDownloader{}

	// Test URL: Azure policy library AVM policies
	testURL := "git::https://github.com/Azure/policy-library-avm.git//policy"

	// Execute the download
	err = downloader.DownloadPolicy(testURL, tempDir)

	// Assertions
	assert.NoError(t, err, "Policy download should succeed")

	// Verify that files were actually downloaded
	files, err := afero.ReadDir(fs, tempDir)
	require.NoError(t, err)
	assert.True(t, len(files) > 0, "Downloaded directory should contain files")

	// Verify that .rego policy files exist
	regoFileCount := 0
	err = afero.Walk(fs, tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".rego") {
			regoFileCount++
		}
		return nil
	})
	require.NoError(t, err)
	assert.True(t, regoFileCount > 0, "Should find at least one .rego policy file")

	t.Logf("Successfully downloaded %d policy files to %s", regoFileCount, tempDir)
}

func TestDownloadDefaultAVMExceptions(t *testing.T) {
	// Setup
	mockFs := afero.NewMemMapFs()
	originalFs := fs
	fs = mockFs
	defer func() { fs = originalFs }()

	// Create temp directory
	tempDir, err := afero.TempDir(fs, "", "test-*")
	require.NoError(t, err)

	// Mock the policy downloader
	mockDownloader := &MockPolicyDownloader{
		downloads: map[string]*MockDownloadResult{
			"https://raw.githubusercontent.com/Azure/policy-library-avm/refs/heads/main/policy/avmsec/avm_exceptions.rego.bak": {err: nil},
		},
		setupPolicyFiles: func(fs afero.Fs, destDir string, url string) error {
			// Create the expected file structure for default AVM exceptions
			exceptionsContent := `package avmsec

import rego.v1

exception contains ["test_exception"] if {
    true
}`
			return afero.WriteFile(fs, filepath.Join(destDir, "avmsec_exceptions.rego"), []byte(exceptionsContent), 0644)
		},
	}
	originalPolicyDownloader := policyDownloader
	policyDownloader = mockDownloader
	defer func() { policyDownloader = originalPolicyDownloader }()

	// Test the function
	source, err := downloadDefaultAVMExceptions(tempDir)

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, source)
	assert.Equal(t, "https://raw.githubusercontent.com/Azure/policy-library-avm/refs/heads/main/policy/avmsec/avm_exceptions.rego.bak", source.OriginalURL)
	assert.Equal(t, filepath.Join(tempDir, "default_exceptions"), source.ResolvedPath)
	assert.Equal(t, "directory", source.Type)
	assert.Equal(t, 1, source.PolicyCount)

	// Verify the exceptions file was created
	exceptionsFilePath := filepath.Join(tempDir, "default_exceptions", "avmsec_exceptions.rego")
	exists, err := afero.Exists(fs, exceptionsFilePath)
	require.NoError(t, err)
	assert.True(t, exists, "Exceptions file should exist")
}

func TestResolvePolicySources_WithDefaultAVMExceptions(t *testing.T) {
	// Setup
	mockFs := afero.NewMemMapFs()
	originalFs := fs
	fs = mockFs
	defer func() { fs = originalFs }()

	// Create temp directory
	tempDir, err := afero.TempDir(fs, "", "test-*")
	require.NoError(t, err)

	// Mock the policy downloader
	mockDownloader := &MockPolicyDownloader{
		downloads: map[string]*MockDownloadResult{
			"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2":               {err: nil},
			"git::https://github.com/Azure/policy-library-avm.git//policy/avmsec":                                              {err: nil},
			"https://raw.githubusercontent.com/Azure/policy-library-avm/refs/heads/main/policy/avmsec/avm_exceptions.rego.bak": {err: nil},
		},
		setupPolicyFiles: func(fs afero.Fs, destDir string, url string) error {
			// Create sample .rego files for different policy sources
			return afero.WriteFile(fs, filepath.Join(destDir, "sample.rego"), []byte("package test"), 0644)
		},
	}
	originalPolicyDownloader := policyDownloader
	policyDownloader = mockDownloader
	defer func() { policyDownloader = originalPolicyDownloader }()

	// Test parameter with IncludeDefaultAVMExceptions enabled
	param := ScanParam{
		PreDefinedPolicyLibraryAlias: "all",
		TargetFile:                   "plan.json",
		IncludeDefaultAVMExceptions:  true,
	}

	// Execute
	sources, err := resolvePolicySources(param, tempDir)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, sources, 3) // 2 from "all" predefined + 1 default exceptions

	// Check that default AVM exceptions source is included
	found := false
	for _, source := range sources {
		if source.OriginalURL == "https://raw.githubusercontent.com/Azure/policy-library-avm/refs/heads/main/policy/avmsec/avm_exceptions.rego.bak" {
			found = true
			assert.Equal(t, "directory", source.Type)
			assert.Equal(t, 1, source.PolicyCount)
			break
		}
	}
	assert.True(t, found, "Default AVM exceptions source should be included")
}

func TestResolvePolicySources_WithoutDefaultAVMExceptions(t *testing.T) {
	// Setup
	mockFs := afero.NewMemMapFs()
	originalFs := fs
	fs = mockFs
	defer func() { fs = originalFs }()

	// Create temp directory
	tempDir, err := afero.TempDir(fs, "", "test-*")
	require.NoError(t, err)

	// Mock the policy downloader
	mockDownloader := &MockPolicyDownloader{
		downloads: map[string]*MockDownloadResult{
			"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2": {err: nil},
			"git::https://github.com/Azure/policy-library-avm.git//policy/avmsec":                                {err: nil},
		},
		setupPolicyFiles: func(fs afero.Fs, destDir string, url string) error {
			// Create sample .rego files for different policy sources
			return afero.WriteFile(fs, filepath.Join(destDir, "sample.rego"), []byte("package test"), 0644)
		},
	}
	originalPolicyDownloader := policyDownloader
	policyDownloader = mockDownloader
	defer func() { policyDownloader = originalPolicyDownloader }()

	// Test parameter with IncludeDefaultAVMExceptions disabled (default)
	param := ScanParam{
		PreDefinedPolicyLibraryAlias: "all",
		TargetFile:                   "plan.json",
		IncludeDefaultAVMExceptions:  false,
	}

	// Execute
	sources, err := resolvePolicySources(param, tempDir)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, sources, 2) // Only 2 from "all" predefined, no default exceptions

	// Check that default AVM exceptions source is NOT included
	for _, source := range sources {
		assert.NotEqual(t, "https://raw.githubusercontent.com/Azure/policy-library-avm/refs/heads/main/policy/avmsec/avm_exceptions.rego.bak", source.OriginalURL)
	}
}
