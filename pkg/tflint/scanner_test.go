package tflint

import (
	"strings"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTargetDirectory(t *testing.T) {
	tests := []struct {
		name      string
		setupFs   func(fs afero.Fs) string // returns target path
		wantErr   bool
		errString string
	}{
		{
			name: "should validate existing directory",
			setupFs: func(fs afero.Fs) string {
				dir := "/test/valid/dir"
				require.NoError(t, fs.MkdirAll(dir, 0755))
				return dir
			},
			wantErr: false,
		},
		{
			name: "should error on non-existent directory",
			setupFs: func(fs afero.Fs) string {
				return "/test/nonexistent"
			},
			wantErr:   true,
			errString: "target directory does not exist",
		},
		{
			name: "should error when path is a file, not directory",
			setupFs: func(fs afero.Fs) string {
				file := "/test/file.txt"
				require.NoError(t, fs.MkdirAll("/test", 0755))
				require.NoError(t, afero.WriteFile(fs, file, []byte("content"), 0644))
				return file
			},
			wantErr:   true,
			errString: "target path is not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use memory filesystem for testing
			memFs := afero.NewMemMapFs()
			stubs := gostub.Stub(&fs, memFs)
			defer stubs.Reset()

			targetPath := tt.setupFs(fs)

			err := validateTargetDirectory(targetPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteTFLintInit(t *testing.T) {
	tests := []struct {
		name        string
		targetPath  string
		configPath  string
		expectError bool
	}{
		{
			name:        "should execute tflint init successfully",
			targetPath:  "/test/terraform",
			configPath:  "/test/config/.tflint.hcl",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the command executor
			mockExecutor := &MockCommandExecutor{
				commands: make(map[string]*MockCommandResult),
			}

			// Set expected command and result
			expectedCmd := "tflint --init --config=" + tt.configPath
			mockExecutor.commands[expectedCmd] = &MockCommandResult{
				stdout: "TFLint initialized successfully",
				stderr: "",
				err:    nil,
			}

			stubs := gostub.Stub(&commandExecutor, mockExecutor)
			defer stubs.Reset()

			output, err := executeTFLintInit(tt.targetPath, tt.configPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, "TFLint initialized")
			}
		})
	}
}

func TestExecuteTFLintScan(t *testing.T) {
	tests := []struct {
		name         string
		targetPath   string
		configPath   string
		ignoredRules []string
		expectError  bool
	}{
		{
			name:        "should execute tflint scan successfully",
			targetPath:  "/test/terraform",
			configPath:  "/test/config/.tflint.hcl",
			expectError: false,
		},
		{
			name:         "should execute tflint scan with ignored rules",
			targetPath:   "/test/terraform",
			configPath:   "/test/config/.tflint.hcl",
			ignoredRules: []string{"terraform_unused_declarations"},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the command executor
			mockExecutor := &MockCommandExecutor{
				commands: make(map[string]*MockCommandResult),
			}

			// Build expected command
			expectedCmd := "tflint --format=json --config=" + tt.configPath
			for _, rule := range tt.ignoredRules {
				expectedCmd += " --disable-rule=" + rule
			}

			// Set expected result
			mockOutput := `{
				"issues": [
					{
						"rule": {
							"name": "terraform_deprecated_syntax",
							"severity": "warning"
						},
						"message": "Deprecated syntax found",
						"range": {
							"filename": "main.tf",
							"start": {"line": 1, "column": 1},
							"end": {"line": 1, "column": 10}
						}
					}
				],
				"errors": []
			}`

			mockExecutor.commands[expectedCmd] = &MockCommandResult{
				stdout: mockOutput,
				stderr: "",
				err:    nil,
			}

			stubs := gostub.Stub(&commandExecutor, mockExecutor)
			defer stubs.Reset()

			output, err := executeTFLintScan(tt.targetPath, tt.configPath, tt.ignoredRules)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, output, "terraform_deprecated_syntax")
			}
		})
	}
}

func TestParseTFLintOutput(t *testing.T) {
	tests := []struct {
		name           string
		jsonOutput     string
		expectedIssues int
		expectedErrors int
		wantErr        bool
	}{
		{
			name: "should parse valid TFLint JSON output",
			jsonOutput: `{
				"issues": [
					{
						"rule": {
							"name": "terraform_unused_declarations",
							"severity": "warning"
						},
						"message": "Variable is declared but not used",
						"range": {
							"filename": "variables.tf",
							"start": {"line": 5, "column": 1},
							"end": {"line": 5, "column": 20}
						}
					},
					{
						"rule": {
							"name": "terraform_deprecated_syntax",
							"severity": "error"
						},
						"message": "Deprecated syntax found",
						"range": {
							"filename": "main.tf",
							"start": {"line": 10, "column": 5},
							"end": {"line": 10, "column": 15}
						}
					}
				],
				"errors": []
			}`,
			expectedIssues: 2,
			expectedErrors: 1, // One error-severity issue
			wantErr:        false,
		},
		{
			name: "should parse output with errors",
			jsonOutput: `{
				"issues": [],
				"errors": [
					{
						"message": "Failed to load configuration",
						"range": {
							"filename": "main.tf",
							"start": {"line": 1, "column": 1},
							"end": {"line": 1, "column": 1}
						}
					}
				]
			}`,
			expectedIssues: 1, // Errors are converted to issues
			expectedErrors: 1, // This should track the error count in summary
			wantErr:        false,
		},
		{
			name:       "should error on invalid JSON",
			jsonOutput: `{invalid json`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseScanOutput(tt.jsonOutput, "reusable", "/test/path", "Init output")

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, result.Success)
				return
			}

			require.NoError(t, err)
			assert.True(t, result.Success)
			assert.Equal(t, "reusable", result.Category)
			assert.Equal(t, "/test/path", result.TargetPath)
			assert.Len(t, result.Issues, tt.expectedIssues)
			assert.Equal(t, tt.expectedIssues, result.Summary.TotalIssues)
			assert.Equal(t, tt.expectedErrors, result.Summary.ErrorCount)

			if tt.expectedIssues > 0 {
				// Verify first issue structure
				assert.NotEmpty(t, result.Issues[0].Rule)
				assert.NotEmpty(t, result.Issues[0].Severity)
				assert.NotEmpty(t, result.Issues[0].Message)
			}
		})
	}
}

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

func TestRunTFLintScan_Integration(t *testing.T) {
	tests := []struct {
		name           string
		param          ScanParam
		setupFs        func(fs afero.Fs)
		setupCommands  func(mock *MockCommandExecutor)
		expectedResult func(t *testing.T, result *ScanResult, err error)
	}{
		{
			name: "should run complete scan successfully",
			param: ScanParam{
				Category:   "reusable",
				TargetPath: "/test/terraform", // Use absolute path
			},
			setupFs: func(fs afero.Fs) {
				// Create target directory
				require.NoError(t, fs.MkdirAll("/test/terraform", 0755))
				// Create a terraform file
				require.NoError(t, afero.WriteFile(fs, "/test/terraform/main.tf", []byte("# test file"), 0644))
			},
			setupCommands: func(mock *MockCommandExecutor) {
				// Use pattern-based matching for commands with temp paths
				mock.patterns = make(map[string]*MockCommandResult)
				mock.patterns["tflint --init --config="] = &MockCommandResult{
					stdout: "TFLint initialized successfully",
					stderr: "",
					err:    nil,
				}
				mock.patterns["tflint --format=json --config="] = &MockCommandResult{
					stdout: `{
						"issues": [
							{
								"rule": {
									"name": "terraform_unused_declarations",
									"severity": "warning"
								},
								"message": "Variable is declared but not used",
								"range": {
									"filename": "variables.tf",
									"start": {"line": 5, "column": 1},
									"end": {"line": 5, "column": 20}
								}
							}
						],
						"errors": []
					}`,
					stderr: "",
					err:    nil,
				}
			},
			expectedResult: func(t *testing.T, result *ScanResult, err error) {
				require.NoError(t, err)
				assert.True(t, result.Success)
				assert.Equal(t, "reusable", result.Category)
				assert.Len(t, result.Issues, 1)
				assert.Equal(t, 1, result.Summary.TotalIssues)
				assert.Equal(t, 0, result.Summary.ErrorCount)
				assert.Equal(t, 1, result.Summary.WarningCount)
			},
		},
		{
			name: "should handle scan with ignored rules",
			param: ScanParam{
				Category:     "example",
				TargetPath:   "/test/terraform",
				IgnoredRules: []string{"terraform_unused_declarations"},
			},
			setupFs: func(fs afero.Fs) {
				require.NoError(t, fs.MkdirAll("/test/terraform", 0755))
			},
			setupCommands: func(mock *MockCommandExecutor) {
				// Use pattern-based matching for commands with temp paths
				mock.patterns = make(map[string]*MockCommandResult)
				mock.patterns["tflint --init --config="] = &MockCommandResult{
					stdout: "TFLint initialized successfully",
					stderr: "",
					err:    nil,
				}
				mock.patterns["tflint --format=json --config="] = &MockCommandResult{
					stdout: `{"issues": [], "errors": []}`,
					stderr: "",
					err:    nil,
				}
			},
			expectedResult: func(t *testing.T, result *ScanResult, err error) {
				require.NoError(t, err)
				assert.True(t, result.Success)
				assert.Equal(t, "example", result.Category)
				assert.Len(t, result.Issues, 0)
				assert.Equal(t, 0, result.Summary.TotalIssues)
			},
		},
		{
			name: "should handle non-existent target directory",
			param: ScanParam{
				TargetPath: "/test/nonexistent",
			},
			setupFs: func(fs afero.Fs) {
				// Don't create the directory
			},
			setupCommands: func(mock *MockCommandExecutor) {
				// No commands should be executed
			},
			expectedResult: func(t *testing.T, result *ScanResult, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "target directory does not exist")
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use memory filesystem for testing
			memFs := afero.NewMemMapFs()
			fsStubs := gostub.Stub(&fs, memFs)
			defer fsStubs.Reset()

			// Setup filesystem
			tt.setupFs(fs)

			// Mock the command executor
			mockExecutor := &MockCommandExecutor{
				commands: make(map[string]*MockCommandResult),
			}
			tt.setupCommands(mockExecutor)
			cmdStubs := gostub.Stub(&commandExecutor, mockExecutor)
			defer cmdStubs.Reset()

			// Mock the download function to return test content
			downloadStubs := gostub.Stub(&downloadConfigContent, func(url string) (string, error) {
				return `rule "terraform_deprecated_syntax" { enabled = true }`, nil
			})
			defer downloadStubs.Reset()

			// Mock getDefaultTargetPath to return the target path as-is
			pathStubs := gostub.Stub(&getDefaultTargetPath, func(targetPath string) (string, error) {
				if targetPath == "" {
					return "/default/path", nil
				}
				return targetPath, nil
			})
			defer pathStubs.Reset()

			// Override the temp dir creation to return predictable paths
			originalSetupTempConfigDir := setupTempConfigDir
			setupTempConfigDir = func() (string, func(), error) {
				tempDir := "/tmp/tflint-config-123"
				if tt.param.Category == "example" {
					tempDir = "/tmp/tflint-config-456"
				}
				require.NoError(t, fs.MkdirAll(tempDir, 0755))
				cleanup := func() { _ = fs.RemoveAll(tempDir) }
				return tempDir, cleanup, nil
			}
			defer func() { setupTempConfigDir = originalSetupTempConfigDir }()

			// Run the test
			result, err := Scan(tt.param)

			// Verify results
			tt.expectedResult(t, result, err)
		})
	}
}
