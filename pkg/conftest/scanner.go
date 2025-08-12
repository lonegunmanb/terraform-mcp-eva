package conftest

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// CommandExecutor interface for executing system commands (following tflint pattern)
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
		if err != nil {
			return string(stdoutBytes), string(exitError.Stderr), err
		}
		return string(stdoutBytes), "", err
	}

	return string(stdoutBytes), "", nil
}

// Global command executor for testing (following tflint pattern)
var commandExecutor CommandExecutor = &RealCommandExecutor{}

// validateTargetPlan validates that the plan file exists and is a file
func validateTargetPlan(planFile string) error {
	info, err := fs.Stat(planFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plan file does not exist: %s", planFile)
		}
		return fmt.Errorf("failed to stat plan file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("plan path is not a file: %s", planFile)
	}

	return nil
}

// buildConftestCommand builds the conftest command with policy sources and options
func buildConftestCommand(planFile string, policySources []PolicySource, namespaces []string) string {
	parts := []string{"conftest", "test"}

	// Add namespace flags
	if len(namespaces) > 0 {
		for _, ns := range namespaces {
			parts = append(parts, "--namespace", ns)
		}
	} else {
		parts = append(parts, "--all-namespaces")
	}

	// Add policy source flags
	for _, source := range policySources {
		parts = append(parts, "-p", source.ResolvedPath)
	}

	// Add plan file
	parts = append(parts, planFile)

	return strings.Join(parts, " ")
}

// executeConftestScan executes the conftest command and returns the output
func executeConftestScan(workingDir, command string) (string, error) {
	stdout, stderr, err := commandExecutor.ExecuteCommand(workingDir, command)
	if err != nil {
		// Conftest may exit with non-zero status when violations are found, but still provide valid output
		if stdout != "" {
			// Try to parse the output, if successful, treat as non-fatal
			var result ConftestRawOutput
			parseErr := json.Unmarshal([]byte(stdout), &result)
			if parseErr == nil {
				return stdout, nil
			}
		}
		return stdout, fmt.Errorf("conftest scan failed: %w, stderr: %s", err, stderr)
	}

	return stdout, nil
}

// ConftestRawOutput represents the raw JSON output from conftest
type ConftestRawOutput struct {
	Warnings  []ConftestNamespaceResult `json:"warnings"`
	Failures  []ConftestNamespaceResult `json:"failures"`
	Successes int                       `json:"successes"`
}

// ConftestNamespaceResult represents results for a specific namespace
type ConftestNamespaceResult struct {
	Filename  string                  `json:"filename"`
	Namespace string                  `json:"namespace"`
	Successes int                     `json:"successes,omitempty"`
	Failures  []ConftestFailureDetail `json:"failures,omitempty"`
	Warnings  []ConftestWarningDetail `json:"warnings,omitempty"`
}

// ConftestFailureDetail represents a single failure
type ConftestFailureDetail struct {
	Message string `json:"msg"`
}

// ConftestWarningDetail represents a single warning
type ConftestWarningDetail struct {
	Message string `json:"msg"`
}

// parseConftestOutput parses conftest JSON output into violations and warnings
func parseConftestOutput(output string) ([]PolicyViolation, []PolicyWarning, error) {
	var rawOutput ConftestRawOutput
	if err := json.Unmarshal([]byte(output), &rawOutput); err != nil {
		return nil, nil, fmt.Errorf("failed to parse conftest output: %w", err)
	}

	var violations []PolicyViolation
	var warnings []PolicyWarning

	// Parse failures as violations
	for _, failure := range rawOutput.Failures {
		namespace, rule := parseNamespace(failure.Namespace)
		for _, detail := range failure.Failures {
			violations = append(violations, PolicyViolation{
				Policy:    rule,
				Rule:      rule,
				Message:   detail.Message,
				Namespace: namespace,
				Severity:  "error",
				Resource:  "",
			})
		}
	}

	// Parse warnings
	for _, warning := range rawOutput.Warnings {
		namespace, rule := parseNamespace(warning.Namespace)
		for _, detail := range warning.Warnings {
			warnings = append(warnings, PolicyWarning{
				Policy:    rule,
				Rule:      rule,
				Message:   detail.Message,
				Namespace: namespace,
				Resource:  "",
			})
		}
	}

	return violations, warnings, nil
}

// parseNamespace extracts namespace and rule from conftest namespace format (e.g., "main.storage_account_https_only")
func parseNamespace(fullNamespace string) (namespace, rule string) {
	parts := strings.SplitN(fullNamespace, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return fullNamespace, fullNamespace
}

// createIgnoreConfig creates ignore configuration files for ignored policies
func createIgnoreConfig(ignoredPolicies []IgnoredPolicy, tempDir string) ([]string, error) {
	if len(ignoredPolicies) == 0 {
		return []string{}, nil
	}

	// Group ignored policies by namespace
	namespaceRules := make(map[string][]string)
	for _, policy := range ignoredPolicies {
		namespaceRules[policy.Namespace] = append(namespaceRules[policy.Namespace], policy.Name)
	}

	var exceptionPaths []string
	for namespace, rules := range namespaceRules {
		// Create separate directory for each namespace
		namespaceDir := filepath.Join(tempDir, fmt.Sprintf("exceptions_%s", strings.ToLower(namespace)))
		err := fs.MkdirAll(namespaceDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create namespace directory %s: %w", namespaceDir, err)
		}

		// Create namespace-specific exception file in its own directory
		content := generateIgnoreRegoContent(namespace, rules)
		filename := fmt.Sprintf("exceptions_%s.rego", strings.ToLower(namespace))
		exceptionPath := filepath.Join(namespaceDir, filename)

		err = writeFile(exceptionPath, content)
		if err != nil {
			return nil, fmt.Errorf("failed to create exception file %s: %w", filename, err)
		}

		// Return the namespace directory for -p flag
		exceptionPaths = append(exceptionPaths, namespaceDir)
	}

	return exceptionPaths, nil
}

// generateIgnoreRegoContent generates Rego content for ignoring specific policies
func generateIgnoreRegoContent(namespace string, rules []string) string {
	var quotedRules []string
	for _, rule := range rules {
		quotedRules = append(quotedRules, fmt.Sprintf(`"%s"`, rule))
	}

	content := fmt.Sprintf(`package %s

import rego.v1

exception contains rules if {
    rules = [%s]
}`, namespace, strings.Join(quotedRules, ", "))

	return content
}

// writeFile writes content to a file (helper function for testing)
func writeFile(filename, content string) error {
	return afero.WriteFile(fs, filename, []byte(content), 0644)
}

// Scan performs a conftest scan with the given parameters
func Scan(param ConftestScanParam) (*ConftestScanResult, error) {
	// Validate parameters
	if err := param.Validate(); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Validate plan file
	if err := validateTargetPlan(param.PlanFile); err != nil {
		return nil, fmt.Errorf("plan file validation failed: %w", err)
	}

	// Create temporary directory for all conftest operations
	tempDir, err := afero.TempDir(fs, "", "conftest-scan-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer fs.RemoveAll(tempDir) // Ensure cleanup

	// Resolve policy URLs
	urls, err := resolvePolicyUrls(param.PreDefinedPolicyLibraryAlias, param.PolicyUrls)
	if err != nil {
		return nil, fmt.Errorf("policy URL resolution failed: %w", err)
	}

	// Mock policy sources for now (in real implementation, these would be downloaded)
	var policySources []PolicySource
	for _, url := range urls {
		policySources = append(policySources, PolicySource{
			OriginalURL:  url,
			ResolvedPath: "/tmp/mock-policies", // Mock path for testing
			Type:         "directory",
			PolicyCount:  5,
		})
	}

	// Handle ignored policies if any
	if len(param.IgnoredPolicies) > 0 {
		exceptionPaths, err := createIgnoreConfig(param.IgnoredPolicies, tempDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create ignore config: %w", err)
		}

		// Add exception paths as additional policy sources
		for _, path := range exceptionPaths {
			policySources = append(policySources, PolicySource{
				OriginalURL:  "ignore-config",
				ResolvedPath: path,
				Type:         "directory",
				PolicyCount:  1, // Each exception file counts as 1 policy
			})
		}
	}

	// Build conftest command
	command := buildConftestCommand(param.PlanFile, policySources, param.Namespaces)

	// Execute conftest scan
	output, err := executeConftestScan("", command)
	if err != nil {
		return nil, fmt.Errorf("conftest execution failed: %w", err)
	}

	// Parse output
	violations, warnings, err := parseConftestOutput(output)
	if err != nil {
		return nil, fmt.Errorf("output parsing failed: %w", err)
	}

	// Build result
	result := &ConftestScanResult{
		Success:       true,
		PlanFile:      param.PlanFile,
		PolicySources: policySources,
		Violations:    violations,
		Warnings:      warnings,
		Output:        output,
		Summary: ConftestSummary{
			TotalViolations: len(violations),
			ErrorCount:      len(violations),
			WarningCount:    len(warnings),
			InfoCount:       0,
			PoliciesRun:     len(policySources),
		},
	}

	return result, nil
}
