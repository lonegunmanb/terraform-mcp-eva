package conftest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	getter "github.com/hashicorp/go-getter/v2"
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
		if errors.As(err, &exitError) {
			return string(stdoutBytes), string(exitError.Stderr), err
		}
	}

	return string(stdoutBytes), "", err
}

// Global command executor for testing (following tflint pattern)
var commandExecutor CommandExecutor = &RealCommandExecutor{}

// PolicyDownloader interface for downloading policy sources (following tflint pattern)
type PolicyDownloader interface {
	DownloadPolicy(url, destDir string) error
}

// RealPolicyDownloader implements PolicyDownloader using go-getter
type RealPolicyDownloader struct{}

func (r *RealPolicyDownloader) DownloadPolicy(url, destDir string) error {
	// Apply timeout with env var override (default 60s, override via CONFTEST_POLICY_DOWNLOAD_TIMEOUT_SECONDS)
	timeout := 60 * time.Second
	if v := os.Getenv("CONFTEST_POLICY_DOWNLOAD_TIMEOUT_SECONDS"); v != "" {
		if secs, parseErr := strconv.Atoi(v); parseErr == nil && secs > 0 {
			timeout = time.Duration(secs) * time.Second
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use go-getter to download to the destination directory
	// GetAny supports both files and directories, which is what we need for policy sources
	if _, err := getter.GetAny(ctx, destDir, url); err != nil {
		return fmt.Errorf("go-getter GetAny failed for URL %s: %w", url, err)
	}

	return nil
}

// Global policy downloader for testing (following tflint pattern)
var policyDownloader PolicyDownloader = &RealPolicyDownloader{}

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
	parts := []string{"conftest", "test", "--no-color", "-o", "json"}

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
			var result RawOutput
			parseErr := json.Unmarshal([]byte(stdout), &result)
			if parseErr == nil {
				return stdout, nil
			}
		}
		return stdout, fmt.Errorf("conftest scan failed: %w, stderr: %s", err, stderr)
	}

	return stdout, nil
}

// RawOutput represents the raw JSON output from conftest (array of namespace results)
type RawOutput []NamespaceResult

// NamespaceResult represents results for a specific namespace
type NamespaceResult struct {
	Filename  string          `json:"filename"`
	Namespace string          `json:"namespace"`
	Successes int             `json:"successes,omitempty"`
	Failures  []FailureDetail `json:"failures,omitempty"`
	Warnings  []WarningDetail `json:"warnings,omitempty"`
}

// FailureDetail represents a single failure
type FailureDetail struct {
	Message  string            `json:"msg"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// WarningDetail represents a single warning
type WarningDetail struct {
	Message  string            `json:"msg"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ParseConftestOutput parses conftest JSON output into violations and warnings (exported for testing)
func ParseConftestOutput(output string) ([]PolicyViolation, []PolicyWarning, error) {
	return parseConftestOutput(output)
}

// parseConftestOutput parses conftest JSON output into violations and warnings
func parseConftestOutput(output string) ([]PolicyViolation, []PolicyWarning, error) {
	var rawOutput RawOutput
	if err := json.Unmarshal([]byte(output), &rawOutput); err != nil {
		return nil, nil, fmt.Errorf("failed to parse conftest output: %w", err)
	}

	var violations []PolicyViolation
	var warnings []PolicyWarning

	// Iterate through each namespace result
	for _, namespaceResult := range rawOutput {
		// Parse failures as violations
		for _, detail := range namespaceResult.Failures {
			violations = append(violations, PolicyViolation{
				Policy:    namespaceResult.Namespace,
				Rule:      extractRuleFromMessage(detail.Message),
				Message:   detail.Message,
				Namespace: namespaceResult.Namespace,
				Severity:  "error",
				Resource:  extractResourceFromMessage(detail.Message),
			})
		}

		// Parse warnings
		for _, detail := range namespaceResult.Warnings {
			warnings = append(warnings, PolicyWarning{
				Policy:    namespaceResult.Namespace,
				Rule:      extractRuleFromMessage(detail.Message),
				Message:   detail.Message,
				Namespace: namespaceResult.Namespace,
				Resource:  extractResourceFromMessage(detail.Message),
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

// extractRuleFromMessage extracts the rule name from the policy violation message
func extractRuleFromMessage(message string) string {
	// Try to extract rule from the beginning of the message (format: namespace/rule:)
	parts := strings.SplitN(message, ":", 2)
	if len(parts) >= 2 {
		rulePart := strings.TrimSpace(parts[0])
		// Extract just the rule name after the last slash
		if slashIndex := strings.LastIndex(rulePart, "/"); slashIndex != -1 {
			return rulePart[slashIndex+1:]
		}
		return rulePart
	}
	return "unknown"
}

// extractResourceFromMessage extracts the resource reference from the policy violation message
func extractResourceFromMessage(message string) string {
	// Look for patterns like 'module.disk.azurerm_storage_account.example' or 'azurerm_linux_virtual_machine.example'
	// These are typically enclosed in single quotes
	if startQuote := strings.Index(message, "'"); startQuote != -1 {
		remaining := message[startQuote+1:]
		if endQuote := strings.Index(remaining, "'"); endQuote != -1 {
			resource := remaining[:endQuote]
			// Check if this looks like a Terraform resource reference
			if strings.Contains(resource, ".") && (strings.Contains(resource, "azurerm_") || strings.Contains(resource, "module.")) {
				return resource
			}
		}
	}
	return ""
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

// resolvePredefinedPolicyLibrary resolves predefined policy library aliases to URLs
func resolvePredefinedPolicyLibrary(alias string) ([]string, error) {
	urls, exists := predefinedPolicyConfigs[alias]
	if !exists {
		return nil, fmt.Errorf("invalid predefined_policy_library_alias: %s", alias)
	}
	return urls, nil
}

// downloadPolicySource downloads a policy source from a URL and returns a PolicySource
func downloadPolicySource(url, tempDir string) (*PolicySource, error) {
	// Create a unique subdirectory for this policy source
	policyDir, err := afero.TempDir(fs, tempDir, fmt.Sprintf("policy-%d", rand.Int63()))
	if err != nil {
		return nil, fmt.Errorf("failed to create policy directory: %w", err)
	}

	// Download the policy source using go-getter
	if err := downloadPolicyToDirectory(url, policyDir); err != nil {
		return nil, fmt.Errorf("failed to download policy from %s: %w", url, err)
	}

	// Count the number of policy files in the downloaded directory
	policyCount, err := countPolicyFiles(policyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to count policy files in %s: %w", policyDir, err)
	}

	return &PolicySource{
		OriginalURL:  url,
		ResolvedPath: policyDir,
		Type:         "directory",
		PolicyCount:  policyCount,
	}, nil
}

// downloadPolicyToDirectory downloads a policy source to a directory using go-getter
func downloadPolicyToDirectory(url, destDir string) error {
	return policyDownloader.DownloadPolicy(url, destDir)
}

// countPolicyFiles counts the number of .rego files in a directory recursively
func countPolicyFiles(dir string) (int, error) {
	count := 0
	err := afero.Walk(fs, dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".rego") {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// downloadDefaultAVMExceptions downloads the default AVM exceptions from the Azure policy library
func downloadDefaultAVMExceptions(tempDir string) (*PolicySource, error) {
	const defaultAVMExceptionsURL = "https://raw.githubusercontent.com/Azure/policy-library-avm/refs/heads/main/policy/avmsec/avm_exceptions.rego.bak"
	const exceptionsFileName = "avmsec_exceptions.rego"

	// Create a dedicated directory for default exceptions
	exceptionsDir := filepath.Join(tempDir, "default_exceptions")
	err := fs.MkdirAll(exceptionsDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create default exceptions directory: %w", err)
	}

	// Download the exceptions file directly using go-getter
	exceptionsFilePath := filepath.Join(exceptionsDir, exceptionsFileName)
	if err := downloadPolicyToDirectory(defaultAVMExceptionsURL, exceptionsFilePath); err != nil {
		return nil, fmt.Errorf("failed to download default AVM exceptions from %s: %w", defaultAVMExceptionsURL, err)
	}

	return &PolicySource{
		OriginalURL:  defaultAVMExceptionsURL,
		ResolvedPath: exceptionsDir,
		Type:         "directory",
		PolicyCount:  1, // Single exceptions file
	}, nil
}

// resolvePolicySources resolves predefined policy aliases and creates policy sources
func resolvePolicySources(param ScanParam, tempDir string) ([]PolicySource, error) {
	var allUrls []string

	// First, process predefined policy libraries if specified
	if param.PreDefinedPolicyLibraryAlias != "" {
		predefinedUrls, err := resolvePredefinedPolicyLibrary(param.PreDefinedPolicyLibraryAlias)
		if err != nil {
			return nil, fmt.Errorf("predefined policy library resolution failed: %w", err)
		}
		allUrls = append(allUrls, predefinedUrls...)
	}

	// Then, process custom policy URLs
	if len(param.PolicyUrls) > 0 {
		allUrls = append(allUrls, param.PolicyUrls...)
	}

	// If neither predefined nor custom URLs are provided, use default "all"
	if len(allUrls) == 0 {
		predefinedUrls, err := resolvePredefinedPolicyLibrary("all")
		if err != nil {
			return nil, fmt.Errorf("default policy library resolution failed: %w", err)
		}
		allUrls = append(allUrls, predefinedUrls...)
	}

	// Download and create policy sources
	var policySources []PolicySource
	for _, url := range allUrls {
		source, err := downloadPolicySource(url, tempDir)
		if err != nil {
			return nil, fmt.Errorf("failed to download policy source %s: %w", url, err)
		}
		policySources = append(policySources, *source)
	}

	// Handle default AVM exceptions if requested
	if param.IncludeDefaultAVMExceptions {
		defaultExceptionsSource, err := downloadDefaultAVMExceptions(tempDir)
		if err != nil {
			return nil, fmt.Errorf("failed to download default AVM exceptions: %w", err)
		}
		policySources = append(policySources, *defaultExceptionsSource)
	}

	// Handle ignored policies if any
	if len(param.IgnoredPolicies) > 0 {
		exceptionPaths, err := createIgnoreConfig(param.IgnoredPolicies, tempDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create ignore config: %w", err)
		}

		// Add exception paths as additional policy sources
		for _, path := range exceptionPaths {
			source := PolicySource{
				OriginalURL:  "ignore-config",
				ResolvedPath: path,
				Type:         "directory",
				PolicyCount:  1, // Each exception file counts as 1 policy
			}
			policySources = append(policySources, source)
		}
	}

	return policySources, nil
}

// Scan performs a conftest scan with the given parameters
func Scan(param ScanParam) (*ScanResult, error) {
	// Validate parameters
	if err := param.Validate(); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Validate plan file
	if err := validateTargetPlan(param.TargetFile); err != nil {
		return nil, fmt.Errorf("plan file validation failed: %w", err)
	}

	// Create temporary directory for all conftest operations
	tempDir, err := afero.TempDir(fs, "", fmt.Sprintf("conftest-scan-%d", rand.Int63()))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer fs.RemoveAll(tempDir) // Ensure cleanup

	// Resolve and prepare policy sources
	policySources, err := resolvePolicySources(param, tempDir)
	if err != nil {
		return nil, fmt.Errorf("policy source resolution failed: %w", err)
	}

	// Build conftest command
	command := buildConftestCommand(param.TargetFile, policySources, param.Namespaces)

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
	result := &ScanResult{
		Success:       true,
		TargetFile:    param.TargetFile,
		PolicySources: policySources,
		Violations:    violations,
		Warnings:      warnings,
		Output:        output,
		Summary: Summary{
			TotalViolations: len(violations),
			ErrorCount:      len(violations),
			WarningCount:    len(warnings),
			InfoCount:       0,
			PoliciesRun:     len(policySources),
		},
	}

	return result, nil
}
