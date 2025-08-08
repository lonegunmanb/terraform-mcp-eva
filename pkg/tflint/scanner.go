package tflint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CommandExecutor interface for executing system commands
type CommandExecutor interface {
	ExecuteCommand(dir, command string) (stdout, stderr string, err error)
}

// RealCommandExecutor implements CommandExecutor using exec.Command
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) ExecuteCommand(dir, command string) (stdout, stderr string, err error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", "", fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = dir

	stdoutBytes, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return string(stdoutBytes), string(exitError.Stderr), err
		}
		return string(stdoutBytes), "", err
	}

	return string(stdoutBytes), "", nil
}

// Global command executor for testing
var commandExecutor CommandExecutor = &RealCommandExecutor{}

// validateTargetDirectory validates that the target path exists and is a directory
func validateTargetDirectory(targetPath string) error {
	info, err := fs.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target directory does not exist: %s", targetPath)
		}
		return fmt.Errorf("failed to stat target directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("target path is not a directory: %s", targetPath)
	}

	return nil
}

// executeTFLintInit runs tflint --init in the target directory
func executeTFLintInit(targetPath, configPath string) (string, error) {
	command := fmt.Sprintf("tflint --init --config=%s", configPath)

	stdout, stderr, err := commandExecutor.ExecuteCommand(targetPath, command)
	if err != nil {
		return "", fmt.Errorf("tflint init failed: %w, stderr: %s", err, stderr)
	}

	return stdout, nil
}

// executeTFLintScan runs tflint scan in the target directory
func executeTFLintScan(targetPath, configPath string, ignoredRules []string) (string, error) {
	command := fmt.Sprintf("tflint --format=json --config=%s", configPath)

	// Add disable-rule flags for ignored rules
	for _, rule := range ignoredRules {
		command += fmt.Sprintf(" --disable-rule=%s", rule)
	}

	stdout, stderr, err := commandExecutor.ExecuteCommand(targetPath, command)
	if err != nil {
		// TFLint may exit with non-zero status when issues are found, but still provide valid output
		if stdout != "" {
			// Try to parse the output, if successful, treat as non-fatal
			var output Output
			parseErr := json.Unmarshal([]byte(stdout), &output)
			if parseErr == nil {
				return stdout, nil
			}
		}
		return stdout, fmt.Errorf("tflint scan failed: %w, stderr: %s", err, stderr)
	}

	return stdout, nil
}

// Output represents the structure of TFLint JSON output
type Output struct {
	Issues []RawIssue `json:"issues"`
	Errors []RawError `json:"errors"`
}

// RawIssue represents a raw issue from TFLint output
type RawIssue struct {
	Rule struct {
		Name     string `json:"name"`
		Severity string `json:"severity"`
	} `json:"rule"`
	Message string `json:"message"`
	Range   struct {
		Filename string `json:"filename"`
		Start    struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"start"`
		End struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"end"`
	} `json:"range"`
}

// RawError represents a raw error from TFLint output
type RawError struct {
	Message string `json:"message"`
	Range   struct {
		Filename string `json:"filename"`
		Start    struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"start"`
		End struct {
			Line   int `json:"line"`
			Column int `json:"column"`
		} `json:"end"`
	} `json:"range"`
}

// parseScanOutput parses TFLint JSON output into a ScanResult
func parseScanOutput(jsonOutput string, category string, targetPath string, initOutput string) (*ScanResult, error) {
	var output Output
	err := json.Unmarshal([]byte(jsonOutput), &output)
	if err != nil {
		return &ScanResult{
			Success:    false,
			Category:   category,
			TargetPath: targetPath,
			Issues:     nil,
			Output:     fmt.Sprintf("Init: %s\nParse Error: %s", initOutput, err.Error()),
			Summary:    ScanSummary{},
		}, fmt.Errorf("failed to parse TFLint output: %w", err)
	}

	var issues []Issue
	var errorCount, warningCount, infoCount int

	// Convert raw issues to our format
	for _, rawIssue := range output.Issues {
		issue := Issue{
			Rule:     rawIssue.Rule.Name,
			Severity: rawIssue.Rule.Severity,
			Message:  rawIssue.Message,
			Range: Range{
				Filename: rawIssue.Range.Filename,
				Start: Point{
					Line:   rawIssue.Range.Start.Line,
					Column: rawIssue.Range.Start.Column,
				},
				End: Point{
					Line:   rawIssue.Range.End.Line,
					Column: rawIssue.Range.End.Column,
				},
			},
		}

		issues = append(issues, issue)

		// Count by severity
		switch strings.ToLower(issue.Severity) {
		case "error":
			errorCount++
		case "warning":
			warningCount++
		case "info":
			infoCount++
		}
	}

	// Convert errors to issues with error severity
	for _, rawError := range output.Errors {
		issue := Issue{
			Rule:     "tflint_error",
			Severity: "error",
			Message:  rawError.Message,
			Range: Range{
				Filename: rawError.Range.Filename,
				Start: Point{
					Line:   rawError.Range.Start.Line,
					Column: rawError.Range.Start.Column,
				},
				End: Point{
					Line:   rawError.Range.End.Line,
					Column: rawError.Range.End.Column,
				},
			},
		}

		issues = append(issues, issue)
		errorCount++
	}

	summary := ScanSummary{
		TotalIssues:  len(issues),
		ErrorCount:   errorCount,
		WarningCount: warningCount,
		InfoCount:    infoCount,
	}

	return &ScanResult{
		Success:    true,
		Category:   category,
		TargetPath: targetPath,
		Issues:     issues,
		Output:     fmt.Sprintf("Init: %s\nScan: %s", initOutput, jsonOutput),
		Summary:    summary,
	}, nil
}

// Scan executes a complete TFLint scan
func Scan(param ScanParam) (*ScanResult, error) {
	// Validate mutual exclusivity between Category and RemoteConfigUrl
	if param.Category != "" && param.RemoteConfigUrl != "" {
		return nil, fmt.Errorf("category and remote_config_url are mutually exclusive; set only one")
	}
	// Apply defaults
	category := getDefaultCategory(param.Category)
	targetPath, err := getDefaultTargetPath(param.TargetPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Validate target directory
	err = validateTargetDirectory(targetPath)
	if err != nil {
		return nil, err
	}

	var config *ConfigData
	var cleanup func()
	if param.RemoteConfigUrl != "" {
		config, cleanup, err = setupRemoteConfig(param.RemoteConfigUrl)
	} else {
		config, cleanup, err = setupConfig(category)
	}
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to setup config: %w", err)
	}

	// (custom config already merged in setup functions)

	// Initialize TFLint
	initOutput, err := executeTFLintInit(targetPath, config.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize TFLint: %w", err)
	}

	// Run TFLint scan
	scanOutput, err := executeTFLintScan(targetPath, config.ConfigPath, param.IgnoredRules)
	if err != nil {
		return &ScanResult{
			Success:    false,
			Category:   category,
			TargetPath: targetPath,
			Issues:     nil,
			Output:     fmt.Sprintf("Init: %s\nScan Error: %s", initOutput, err.Error()),
			Summary:    ScanSummary{},
		}, err
	}

	// Parse scan results
	result, err := parseScanOutput(scanOutput, category, targetPath, initOutput)
	if err != nil {
		return result, err
	}

	return result, nil
}
