# TODO: Add Conftest Tool Implementation

## Current Status (August 12, 2025)
- ✅ **Phase 1 Complete**: Core types and infrastructure implemented
- ✅ **Phase 2 Complete**: Configuration management (policy URL resolution) following TDD
- 📍 **Next**: Implement `scanner_test.go` → `scanner.go` following TDD (no implementation without tests)

## Overview
Add a `conftest_scan` tool to terraform-mcp-eva that provides policy testing capabilities for Terraform plans using Open Policy Agent (OPA) Conftest, following the existing `tflint_scan` tool architecture and TDD principles.

## Analysis of Current Implementation Pattern

### TFLint Tool Architecture Reference
- **MCP Tool Interface Layer**: `pkg/tool/tflint_scan.go`
- **Core Package**: `pkg/tflint/` with modular files:
  - `scanner.go` - Core scanning logic
  - `types.go` - Type definitions
  - `config.go` - Configuration management
  - `remote_getter.go` - Remote resource fetching
  - `utils.go` - Utility functions
  - Comprehensive test files for each component

### Conftest Usage Reference
Based on Azure TFMod Scaffold:
```bash
conftest test --all-namespaces --update \
  git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2 \
  -p policy/aprl -p policy/default_exceptions -p exceptions \
  tfplan.json
```

## Implementation Plan

### Test-Driven Development Mandate
**ALL implementation MUST follow TDD principles:**
1. **Red**: Write failing tests first
2. **Green**: Write minimal code to make tests pass
3. **Refactor**: Improve code while keeping tests green

### Testing Standards
Follow existing project patterns from `pkg/tflint/`:
- Use `github.com/stretchr/testify/{assert,require}`
- Use `github.com/spf13/afero` with `afero.NewMemMapFs()` for file system testing
- Use `github.com/prashantv/gostub` for dependency injection
- Mock external dependencies (command execution, downloads, network)
- Focus on behavior testing, not implementation details

## Phase 1: Core Types and Infrastructure

### 1.1 Type Definitions (`pkg/conftest/types.go`)

#### Input Parameters
```go
type ConftestScanParam struct {
    PreDefinedPolicyLibraryAlias string          `json:"predefined_policy_library_alias,omitempty"` // "aprl", "avmsec", "all"
    PolicyUrls                   []string        `json:"policy_urls,omitempty"`                     // go-getter URLs
    PlanFile                     string          `json:"plan_file"`                                 // Required JSON plan file
    LocalPolicyDirs              []string        `json:"local_policy_dirs,omitempty"`               // Local policy directories
    IgnoredPolicies              []IgnoredPolicy `json:"ignored_policies,omitempty"`                // Policies to ignore
    Namespaces                   []string        `json:"namespaces,omitempty"`                      // Specific namespaces
    IncludeDefaultAVMExceptions  bool            `json:"include_default_avm_exceptions,omitempty"`  // Include AVM exceptions
}

type IgnoredPolicy struct {
    Namespace string `json:"namespace"` // Required: policy namespace
    Name      string `json:"name"`      // Required: policy rule name
}
```

#### Output Types
```go
type ConftestScanResult struct {
    Success       bool                `json:"success"`
    PlanFile      string              `json:"plan_file"`
    PolicySources []PolicySource      `json:"policy_sources"`
    Violations    []PolicyViolation   `json:"violations,omitempty"`
    Warnings      []PolicyWarning     `json:"warnings,omitempty"`
    Output        string              `json:"output"`
    Summary       ConftestSummary     `json:"summary"`
}

type PolicySource struct {
    OriginalURL  string `json:"original_url"`    // Public: original go-getter URL
    PolicyCount  int    `json:"policy_count"`    // Public: number of policies found
    ResolvedPath string `json:"-"`               // Internal: local download path
    Type         string `json:"-"`               // Internal: source type
}

type PolicyViolation struct {
    Policy    string `json:"policy"`
    Rule      string `json:"rule"`
    Message   string `json:"message"`
    Namespace string `json:"namespace"`
    Severity  string `json:"severity"`
    Resource  string `json:"resource,omitempty"`
}

type ConftestSummary struct {
    TotalViolations int `json:"total_violations"`
    ErrorCount      int `json:"error_count"`
    WarningCount    int `json:"warning_count"`
    InfoCount       int `json:"info_count"`
    PoliciesRun     int `json:"policies_run"`
}
```

### 1.2 TDD Implementation Order

1. **`types_test.go`** → **`types.go`** (✅ Complete)
   - ✅ Parameter validation (mutually exclusive options, required fields)
   - ✅ JSON serialization (public vs internal fields)
   - ✅ Ignored policy validation

2. **`config_test.go`** → **`config.go`** (✅ Complete)
   - ✅ Policy URL resolution (predefined vs custom)
   - ✅ Test coverage and validation
   - ✅ Cleaned up unused functions (following TDD principles)

3. **`scanner_test.go`** → **`scanner.go`** (Pending)
   - Command execution mocking
   - Output parsing
   - Error handling

4. **`utils_test.go`** → **`utils.go`** (Pending)
   - Helper functions
   - Validation utilities

## Phase 2: Configuration Management

### 2.1 Policy Resolution (`pkg/conftest/config.go`)

#### Predefined Policy Categories
```go
var predefinedPolicyConfigs = map[string][]string{
    "aprl":   {"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2"},
    "avmsec": {"git::https://github.com/Azure/policy-library-avm.git//policy/avmsec"},
    "all":    {"aprl", "avmsec"}, // Expands to both
}
```

#### Key Functions
- `resolvePolicyUrls(alias, customUrls) ([]string, error)`
- `downloadPolicySource(url, tempDir) (*PolicySource, error)`
- `createIgnoreConfig(ignoredPolicies, tempDir) ([]string, error)`
- `validateIgnoredPolicies(policies) error`

### 2.2 Remote Resource Management
- Leverage existing `go-getter` patterns from TFLint
- Support git, HTTP/HTTPS, local files, archives
- Automatic cleanup with defer patterns
- Timeout and error handling

## Phase 3: Core Scanning Logic

### 3.1 Scanner Implementation (`pkg/conftest/scanner.go`)

#### Primary Function
```go
func Scan(param ConftestScanParam) (*ConftestScanResult, error)
```

#### Key Functions
- `executeConftestScan()` - Command execution
- `parseConftestOutput()` - JSON output parsing
- `buildConftestCommand()` - Command construction
- `validateTargetPlan()` - Plan file validation

### 3.2 Command Execution Pattern
```go
type CommandExecutor interface {
    ExecuteCommand(dir, command string) (stdout, stderr string, err error)
}
```

## Phase 4: MCP Integration

### 4.1 Tool Interface (`pkg/tool/conftest_scan.go`)
- MCP tool parameter handling
- Integration with registry
- Error response formatting

### 4.2 Advanced Features
- Multiple policy source handling with `-p` flags
- Policy ignoring via generated exception files
- Namespace filtering
- Performance optimization for large policy sets

## Testing Strategy

### Mock Patterns
```go
type MockCommandExecutor struct {
    commands []string
    outputs  map[string]string
    errors   map[string]error
}

type MockGetter struct {
    downloadedSources map[string]string
    downloadErrors    map[string]error
}
```

### Test Categories
1. **Unit Tests**: Individual function behaviors (>90% coverage)
2. **Integration Tests**: End-to-end workflows
3. **Error Tests**: Failure scenarios and recovery
4. **Performance Tests**: Large policy sets and timeouts

### Behavior-Focused Testing
- Test outcomes, not implementation details
- Use table-driven tests with comprehensive scenarios
- Mock external dependencies consistently
- Verify side effects and interactions

## Success Criteria

1. ✅ Tool executes conftest scans with multiple policy sources
2. ✅ Support for all go-getter URL formats
3. ✅ AVM policy repository integration
4. ✅ Structured JSON output for AI consumption
5. ✅ Comprehensive error handling
6. ✅ Parameter validation (mutually exclusive options)
7. ✅ Policy ignoring with namespace.name format
8. ✅ >90% test coverage with TDD approach
9. ✅ Consistent with existing codebase patterns

## Dependencies

### Go Modules
- Existing `go-getter` (reuse from TFLint)
- JSON parsing for conftest output
- File system utilities

### External Requirements
- `conftest` binary installation
- Pre-generated Terraform plan files (JSON format)
- Network access for policy downloads

## Migration Strategy

1. **Phase 1**: Core infrastructure without breaking existing functionality ✅
2. **Phase 2**: Configuration management with comprehensive testing ✅
3. **Phase 3**: Scanner implementation with mocked dependencies 🔄
4. **Phase 4**: MCP integration and advanced features (Pending)
5. **Phase 5**: Documentation and final testing (Pending)

## Notes

- Follow established TFLint patterns for consistency
- Support all go-getter URL formats for flexibility
- Each policy URL generates separate `-p` flag in conftest command
- Strict validation for ignored policies (require namespace.name)
- Consider policy source caching for performance
- Maintain backwards compatibility
        "git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2",
        "git::https://github.com/Azure/policy-library-avm.git//policy/avmsec",
    },
}

// PolicySource represents a resolved policy source
type PolicySource struct {
    OriginalURL   string
    ResolvedPath  string  // Internal use only - not exposed in JSON
    Type          string  // Internal use only - not exposed in JSON
    PolicyCount   int
}

// Policy resolution and download functions
func resolvePolicyUrls(category string, customUrls []string) ([]string, error)
func downloadPolicySource(url string, tempDir string) (*PolicySource, error)
func detectSourceType(path string) (string, int, error) // Returns type and policy count - internal use only
func downloadDefaultExceptions(tempDir string) (string, error)
func validatePolicyUrls(urls []string) error
func createIgnoreConfig(ignoredPolicies []IgnoredPolicy, tempDir string) ([]string, error)
func formatIgnoredPolicy(policy IgnoredPolicy) string // Returns namespace.name format
func validateIgnoredPolicies(policies []IgnoredPolicy) error // Validates required fields
```

#### 2.2 Remote Resource Management
- Leverage the `go-getter` pattern from TFLint implementation for downloading policies
- Support for all go-getter URL formats:
  - Git: `git::https://github.com/org/repo.git//path/to/policies?ref=tag`
  - HTTP/HTTPS: `https://example.com/policy.rego` or `https://example.com/policies.zip`
  - Local files: `file:///path/to/local/policies`
  - Archive formats: `.zip`, `.tar.gz`, etc.
- Auto-detection of whether URL points to a directory or single file
- Timeout handling and error management
- Temporary directory management with automatic cleanup

#### 2.3 Policy Source Resolution Algorithm
```go
// Pseudocode for policy resolution:
func resolvePolicySources(param ConftestScanParam) ([]PolicySource, func(), error) {
    var urls []string
    
    // Determine policy URLs
    if param.PreDefinedPolicyLibraryAlias != "" && len(param.PolicyUrls) > 0 {
        return nil, nil, fmt.Errorf("predefined_policy_library_alias and policy_urls are mutually exclusive")
    }
    
    if param.PreDefinedPolicyLibraryAlias != "" {
        urls = predefinedPolicyConfigs[param.PreDefinedPolicyLibraryAlias]
    } else if len(param.PolicyUrls) > 0 {
        urls = param.PolicyUrls
    } else {
        urls = predefinedPolicyConfigs["all"] // Default
    }
    
    // Create temp directory for all downloads
    tempDir := createTempDir()
    cleanup := func() { removeTempDir(tempDir) }
    defer cleanup()
    
    var sources []PolicySource
    for _, url := range urls {
        source, err := downloadPolicySource(url, tempDir)
        if err != nil {
            return nil, nil, err
        }
        sources = append(sources, *source)
    }
    
    // Download default exceptions if requested
    if param.IncludeDefaultAVMExceptions {
        exceptionsPath, err := downloadDefaultExceptions(tempDir)
        if err == nil {
            sources = append(sources, PolicySource{
                OriginalURL:  "default-avm-exceptions",
                ResolvedPath: exceptionsPath, // Internal use only
                Type:        "directory",     // Internal use only
                PolicyCount:  countPolicies(exceptionsPath),
            })
        }
    }
    
    return sources, cleanup, nil
}

// Policy ignoring implementation
func createIgnoreConfig(ignoredPolicies []IgnoredPolicy, tempDir string) ([]string, error) {
    // Create namespace-specific exception files that follow AVM pattern
    // Returns list of exception file paths to include with -p flags
    
    if len(ignoredPolicies) == 0 {
        return nil, nil
    }
    
    exceptionFiles := generateIgnoreRegoContent(ignoredPolicies)
    var exceptionPaths []string
    
    for filename, content := range exceptionFiles {
        exceptionPath := filepath.Join(tempDir, filename)
        if err := writeFile(exceptionPath, content); err != nil {
            return nil, fmt.Errorf("failed to create exception file %s: %w", filename, err)
        }
        exceptionPaths = append(exceptionPaths, filepath.Dir(exceptionPath)) // Add directory for -p flag
    }
    
    return exceptionPaths, nil
}

func formatIgnoredPolicy(policy IgnoredPolicy) string {
    if policy.Namespace == "" || policy.Name == "" {
        // This should be caught by validation before reaching here
        return ""
    }
    return fmt.Sprintf("%s.%s", policy.Namespace, policy.Name)
}

func validateIgnoredPolicies(policies []IgnoredPolicy) error {
    for i, policy := range policies {
        if policy.Namespace == "" {
            return fmt.Errorf("ignored_policies[%d]: namespace is required", i)
        }
        if policy.Name == "" {
            return fmt.Errorf("ignored_policies[%d]: name is required", i)
        }
    }
    return nil
}
```

### Phase 3: MCP Tool Interface

#### 3.1 Tool Implementation (`pkg/tool/conftest_scan.go`)
```go
type ConftestScanParam struct {
    PreDefinedPolicyLibraryAlias string          `json:"predefined_policy_library_alias,omitempty"`    // "aprl", "avmsec", "all" - mutually exclusive with policy_urls
    PolicyUrls        []string        `json:"policy_urls,omitempty"`        // Array of go-getter compatible URLs
    PlanFile          string          `json:"plan_file"`                    // Required pre-generated plan file (JSON format)
    LocalPolicyDirs   []string        `json:"local_policy_dirs,omitempty"`  // Local policy directories
    LocalPolicyDirs   []string        `json:"local_policy_dirs,omitempty"`  // Local policy directories
    IgnoredPolicies   []IgnoredPolicy `json:"ignored_policies,omitempty"`   // Policies to ignore with namespace.name
    Namespaces        []string        `json:"namespaces,omitempty"`         // Specific namespaces to test
    IncludeDefaultAVMExceptions bool  `json:"include_default_avm_exceptions,omitempty"` // Include default AVM exceptions
}

type IgnoredPolicy struct {
    Namespace string `json:"namespace"` // Required policy namespace (e.g., "avmsec", "aprl")
    Name      string `json:"name"`      // Required policy name (e.g., "storage_account_https_only")
}

func ConftestScan(_ context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[ConftestScanParam]) (*mcp.CallToolResultFor[any], error)
```

#### 3.2 Registry Integration (`pkg/registry.go`)
Add tool registration with proper JSON schema and descriptions.

### Phase 4: Advanced Features

#### 4.1 Multiple Policy Source Handling
- Download all policy sources to isolated temporary directories
- Build conftest command with multiple `-p` flags:
  ```bash
  conftest test --all-namespaces \
    -p /tmp/policies/source1 \
    -p /tmp/policies/source2 \
    -p /tmp/policies/source3 \
    plan.json
  ```
- Handle both directory and file policy sources appropriately
- Support for local policy directories via additional `-p` flags

#### 4.2 Policy Exception Handling
- Download default AVM exceptions from policy-library-avm when `include_default_avm_exceptions` is true
- Integrate exceptions as additional policy sources with proper precedence

#### 4.3 Advanced Conftest Features
- Support for namespace filtering via `--namespace` flag
- Policy ignoring through multiple mechanisms:
  1. **Command-line exclusion**: Generate temporary ignore policies for specified namespace.name combinations
  2. **Policy modification**: Dynamically modify downloaded policies to skip ignored rules
  3. **Result filtering**: Post-process conftest output to exclude ignored policy violations
- Output format control (JSON for structured parsing)
- Error handling for policy compilation and execution errors

#### 4.4 Ignored Policy Implementation Strategy
Based on AVM exception pattern, generate Rego exception files per namespace:

```go
// Primary Strategy: Generate namespace-specific exception files
func generateIgnoreRegoContent(ignoredPolicies []IgnoredPolicy) map[string]string {
    // Group ignored policies by namespace
    namespaceRules := make(map[string][]string)
    for _, policy := range ignoredPolicies {
        namespaceRules[policy.Namespace] = append(namespaceRules[policy.Namespace], policy.Name)
    }
    
    // Generate one exception file per namespace
    exceptionFiles := make(map[string]string)
    for namespace, rules := range namespaceRules {
        content := fmt.Sprintf(`package %s

import rego.v1

exception contains rules if {
    rules = %s
}`, namespace, formatRulesArray(rules))
        
        filename := fmt.Sprintf("exceptions_%s.rego", strings.ToLower(namespace))
        exceptionFiles[filename] = content
    }
    
    return exceptionFiles
}

func formatRulesArray(rules []string) string {
    var quotedRules []string
    for _, rule := range rules {
        quotedRules = append(quotedRules, fmt.Sprintf(`"%s"`, rule))
    }
    return fmt.Sprintf("[%s]", strings.Join(quotedRules, ", "))
}

// Example generated exception file for namespace "Azure_Proactive_Resiliency_Library_v2":
/*
package Azure_Proactive_Resiliency_Library_v2

import rego.v1

exception contains rules if {
    rules = ["use_nat_gateway_instead_of_outbound_rules_for_production_load_balancer", "storage_accounts_are_zone_or_region_redundant"]
}
*/

// Alternative Strategy: Post-processing result filtering (fallback)
func filterIgnoredViolations(result *ConftestScanResult, ignoredPolicies []IgnoredPolicy) {
    ignoredSet := make(map[string]bool)
    for _, policy := range ignoredPolicies {
        ignoredSet[formatIgnoredPolicy(policy)] = true
    }
    
    var filteredViolations []PolicyViolation
    for _, violation := range result.Violations {
        policyId := fmt.Sprintf("%s.%s", violation.Namespace, violation.Rule)
        if !ignoredSet[policyId] {
            filteredViolations = append(filteredViolations, violation)
        }
    }
    result.Violations = filteredViolations
}
```

### Phase 5: Test-Driven Development and Documentation

#### 5.1 TDD Implementation Approach
**Write tests first, then implement functionality to make tests pass**

#### 5.2 Behavior-Driven Unit Tests
Focus on testing behaviors and outcomes, not internal data structures:

**Core Scanning Behavior Tests:**
- `TestScan_WithValidPlanAndPolicies_ReturnsSuccessfulResult()`: Verify scan completes successfully with valid inputs
- `TestScan_WithInvalidPlanFile_ReturnsError()`: Verify appropriate error when plan file doesn't exist
- `TestScan_WithEmptyPolicyUrls_UsesDefaultPolicies()`: Verify default "all" policies are used when no URLs provided
- `TestScan_WithMutuallyExclusiveParams_ReturnsError()`: Verify error when both predefined alias and custom URLs are provided

**Policy Resolution Behavior Tests:**
- `TestResolvePolicySources_WithGitURL_DownloadsAndCountsPolicies()`: Verify git URLs are downloaded and policy count is correct
- `TestResolvePolicySources_WithMultipleURLs_DownloadsAllSources()`: Verify all policy sources are downloaded in parallel
- `TestResolvePolicySources_WithInvalidURL_ReturnsError()`: Verify proper error handling for invalid go-getter URLs
- `TestResolvePolicySources_WithNetworkFailure_CleansUpAndReturnsError()`: Verify cleanup occurs on download failures

**Command Building Behavior Tests:**
- `TestBuildConftestCommand_WithMultiplePolicySources_GeneratesCorrectFlags()`: Verify multiple `-p` flags are generated
- `TestBuildConftestCommand_WithIgnoredPolicies_ExcludesSpecifiedPolicies()`: Verify ignored policies are properly excluded
- `TestBuildConftestCommand_WithNamespaceFilter_UsesNamespaceFlag()`: Verify namespace filtering is applied correctly
- `TestBuildConftestCommand_WithLocalPolicyDirs_IncludesLocalPaths()`: Verify local policy directories are included

**Policy Ignoring Behavior Tests:**
- `TestIgnoredPolicyHandling_WithValidNamespaceAndName_ExcludesFromResults()`: Verify policies are excluded from results
- `TestIgnoredPolicyHandling_WithMissingNamespace_ReturnsValidationError()`: Verify validation requires namespace
- `TestIgnoredPolicyHandling_WithMissingName_ReturnsValidationError()`: Verify validation requires policy name
- `TestIgnoredPolicyHandling_WithPartialMatch_OnlyExcludesExactMatches()`: Verify only exact namespace.name matches are ignored

**Output Parsing Behavior Tests:**
- `TestParseConftestOutput_WithViolations_ExtractsCorrectViolationData()`: Verify violations are parsed correctly
- `TestParseConftestOutput_WithWarnings_CategorizesProperly()`: Verify warnings vs errors are categorized correctly
- `TestParseConftestOutput_WithEmptyResult_ReturnsSuccessWithZeroCounts()`: Verify clean runs return success
- `TestParseConftestOutput_WithMalformedJSON_ReturnsParseError()`: Verify error handling for invalid conftest output

**Exception Handling Behavior Tests:**
- `TestDefaultAVMExceptions_WhenEnabled_DownloadsAndIncludesExceptions()`: Verify AVM exceptions are downloaded when requested
- `TestDefaultAVMExceptions_WhenDisabled_DoesNotDownloadExceptions()`: Verify no download when not requested
- `TestDefaultAVMExceptions_WhenDownloadFails_ContinuesWithoutExceptions()`: Verify graceful degradation on download failure

#### 5.3 Integration Behavior Tests
Test end-to-end scenarios focusing on user workflows:

**Complete Workflow Tests:**
- `TestConftestScan_EndToEnd_WithAPRLPolicies_ReturnsValidResults()`: Full scan with APRL policies
- `TestConftestScan_EndToEnd_WithCustomPolicyURL_DownloadsAndScans()`: Full scan with custom policy URL
- `TestConftestScan_EndToEnd_WithIgnoredPolicies_ExcludesViolations()`: Full scan with policy exclusions
- `TestConftestScan_EndToEnd_WithMultiplePolicySources_CombinesResults()`: Full scan with multiple policy sources

**Error Recovery Tests:**
- `TestConftestScan_WithConftestNotInstalled_ReturnsHelpfulError()`: Verify clear error when conftest binary missing
- `TestConftestScan_WithNetworkTimeout_ReturnsTimeoutError()`: Verify timeout handling for policy downloads
- `TestConftestScan_WithCorruptedPolicyFile_ReturnsValidationError()`: Verify handling of corrupted policy files

**Performance Behavior Tests:**
- `TestConftestScan_WithLargePolicySet_CompletesWithinTimeout()`: Verify performance with large policy sets
- `TestConftestScan_WithParallelDownloads_IsFasterThanSequential()`: Verify parallel downloading improves performance

#### 5.4 Mock Strategy for TDD (Following Project Patterns)
**Based on `pkg/tflint/` Testing Patterns:**

**File System Mocking (Primary Pattern):**
```go
func TestWithFileSystem(t *testing.T) {
    // Use memory filesystem for all file operations
    memFs := afero.NewMemMapFs()
    stubs := gostub.Stub(&fs, memFs)
    defer stubs.Reset()
    
    // Create test files and directories
    require.NoError(t, fs.MkdirAll("/test/policies", 0755))
    require.NoError(t, afero.WriteFile(fs, "/test/policies/policy1.rego", []byte("package test"), 0644))
    
    // Test your functionality
    result, err := functionUnderTest("/test/policies")
    assert.NoError(t, err)
}
```

**Command Execution Mocking:**
```go
type MockCommandExecutor struct {
    commands   []string
    outputs    map[string]string // command -> output mapping
    errors     map[string]error  // command -> error mapping
    callCount  int
}

func (m *MockCommandExecutor) ExecuteCommand(dir, command string) (stdout, stderr string, err error) {
    // Record command for verification (following tflint pattern)
    m.commands = append(m.commands, command)
    m.callCount++
    
    // Return predetermined response based on command
    if output, exists := m.outputs[command]; exists {
        return output, "", m.errors[command]
    }
    
    return "", "", fmt.Errorf("unexpected command: %s", command)
}

// Usage in tests
func TestConftestExecution_WithMockedCommand(t *testing.T) {
    mockExecutor := &MockCommandExecutor{
        outputs: map[string]string{
            "conftest test --all-namespaces -p /tmp/policies plan.json": `{"violations": []}`,
        },
        errors: map[string]error{
            "conftest test --all-namespaces -p /tmp/policies plan.json": nil,
        },
    }
    
    // Inject mock
    stubs := gostub.Stub(&commandExecutor, mockExecutor)
    defer stubs.Reset()
    
    // Test
    result, err := executeConftestScan("plan.json", policySources, ignoredPolicies)
    assert.NoError(t, err)
    assert.Contains(t, mockExecutor.commands[0], "conftest test")
}
```

**Go-Getter Mocking (For Policy Downloads):**
```go
type MockGetter struct {
    downloadedSources map[string]string // URL -> local path mapping
    downloadErrors    map[string]error  // URL -> error mapping for failure scenarios
    downloadDelay     time.Duration     // simulate download time
}

func (m *MockGetter) Get(dst, src string) error {
    // Simulate download delay if configured
    if m.downloadDelay > 0 {
        time.Sleep(m.downloadDelay)
    }
    
    // Check for configured errors
    if err, exists := m.downloadErrors[src]; exists {
        return err
    }
    
    // Simulate successful download by creating files
    if localPath, exists := m.downloadedSources[src]; exists {
        // Copy from mock source to destination
        return copyMockPolicies(localPath, dst)
    }
    
    return fmt.Errorf("mock: unknown source %s", src)
}

// Test with mocked downloads
func TestDownloadPolicySource_WithGitURL_CountsPoliciesCorrectly(t *testing.T) {
    memFs := afero.NewMemMapFs()
    stubs := gostub.Stub(&fs, memFs)
    defer stubs.Reset()
    
    // Setup mock files
    require.NoError(t, fs.MkdirAll("/mock/policies", 0755))
    require.NoError(t, afero.WriteFile(fs, "/mock/policies/policy1.rego", []byte("package test1"), 0644))
    require.NoError(t, afero.WriteFile(fs, "/mock/policies/policy2.rego", []byte("package test2"), 0644))
    
    mockGetter := &MockGetter{
        downloadedSources: map[string]string{
            "git::https://example.com/policies.git": "/mock/policies",
        },
    }
    
    // Inject mock getter
    originalGetter := getter
    getter = mockGetter
    defer func() { getter = originalGetter }()
    
    source, err := downloadPolicySource("git::https://example.com/policies.git", "/tmp")
    
    assert.NoError(t, err)
    assert.Equal(t, 2, source.PolicyCount) // Behavior: correctly counts .rego files
    assert.Equal(t, "git::https://example.com/policies.git", source.OriginalURL) // Behavior: preserves original URL
}
```

**Network and Timeout Mocking:**
```go
type MockHTTPClient struct {
    responses map[string]*http.Response
    errors    map[string]error
    delay     time.Duration
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
    if m.delay > 0 {
        time.Sleep(m.delay)
    }
    
    if err, exists := m.errors[url]; exists {
        return nil, err
    }
    
    if resp, exists := m.responses[url]; exists {
        return resp, nil
    }
    
    return nil, fmt.Errorf("mock: unknown URL %s", url)
}
```

#### 5.5 Test Data Strategy (Updated for Testify and Project Patterns)
**Focus on Behavior Scenarios, Not Data Structure Validation:**

**AVOID:** Testing data structure fields directly
```go
// DON'T DO THIS - testing data structure
func TestPolicySource_HasCorrectFields(t *testing.T) {
    source := PolicySource{OriginalURL: "test", PolicyCount: 5}
    assert.Equal(t, "test", source.OriginalURL)
    assert.Equal(t, 5, source.PolicyCount)
}
```

**PREFER:** Testing behavior with proper setup and mocking
```go
// DO THIS - testing behavior with project patterns
func TestDownloadPolicySource_WithGitURL_CountsPoliciesCorrectly(t *testing.T) {
    // Setup memory filesystem (following project pattern)
    memFs := afero.NewMemMapFs()
    stubs := gostub.Stub(&fs, memFs)
    defer stubs.Reset()
    
    // Create mock policy files
    policyDir := "/tmp/policies"
    require.NoError(t, fs.MkdirAll(policyDir, 0755))
    require.NoError(t, afero.WriteFile(fs, filepath.Join(policyDir, "policy1.rego"), []byte("package test1"), 0644))
    require.NoError(t, afero.WriteFile(fs, filepath.Join(policyDir, "policy2.rego"), []byte("package test2"), 0644))
    require.NoError(t, afero.WriteFile(fs, filepath.Join(policyDir, "policy3.rego"), []byte("package test3"), 0644))
    
    // Mock go-getter download
    mockGetter := &MockGetter{
        downloadedSources: map[string]string{
            "git::https://example.com/policies.git": policyDir,
        },
    }
    originalGetter := getter
    getter = mockGetter
    defer func() { getter = originalGetter }()
    
    // Execute function under test
    source, err := downloadPolicySource("git::https://example.com/policies.git", "/tmp")
    
    // Assert behavior (using testify)
    assert.NoError(t, err)
    assert.Equal(t, 3, source.PolicyCount, "should correctly count .rego files")
    assert.Equal(t, "git::https://example.com/policies.git", source.OriginalURL, "should preserve original URL")
    
    // Verify internal fields are set but not exposed in JSON
    assert.NotEmpty(t, source.ResolvedPath, "internal ResolvedPath should be set")
    assert.Equal(t, "directory", source.Type, "internal Type should be set")
    
    // Verify JSON serialization excludes internal fields
    data, err := json.Marshal(source)
    require.NoError(t, err)
    
    var jsonMap map[string]interface{}
    require.NoError(t, json.Unmarshal(data, &jsonMap))
    
    assert.Contains(t, jsonMap, "original_url")
    assert.Contains(t, jsonMap, "policy_count")
    assert.NotContains(t, jsonMap, "resolved_path", "internal field should be omitted from JSON")
    assert.NotContains(t, jsonMap, "type", "internal field should be omitted from JSON")
}
```

**Table-Driven Tests with Proper Setup:**
```go
func TestResolvePolicyUrls_WithVariousInputs(t *testing.T) {
    tests := []struct {
        name            string
        setupFs         func(fs afero.Fs)
        predefinedAlias string
        customUrls      []string
        wantErr         bool
        errMsg          string
        wantUrlCount    int
        wantUrls        []string
    }{
        {
            name:            "should resolve APRL predefined policies",
            setupFs:         func(fs afero.Fs) {}, // no setup needed
            predefinedAlias: "aprl",
            wantErr:         false,
            wantUrlCount:    1,
            wantUrls:        []string{"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2"},
        },
        {
            name:            "should resolve custom policy URLs",
            setupFs:         func(fs afero.Fs) {}, // no setup needed
            customUrls:      []string{"git::https://example.com/policies.git", "https://example.com/policy.zip"},
            wantErr:         false,
            wantUrlCount:    2,
            wantUrls:        []string{"git::https://example.com/policies.git", "https://example.com/policy.zip"},
        },
        {
            name:            "should fail with invalid predefined alias",
            setupFs:         func(fs afero.Fs) {}, // no setup needed
            predefinedAlias: "invalid",
            wantErr:         true,
            errMsg:          "invalid predefined_policy_library_alias: invalid",
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
            urls, err := resolvePolicyUrls(tt.predefinedAlias, tt.customUrls)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Equal(t, tt.errMsg, err.Error())
                }
            } else {
                assert.NoError(t, err)
                assert.Len(t, urls, tt.wantUrlCount)
                if tt.wantUrls != nil {
                    assert.Equal(t, tt.wantUrls, urls)
                }
            }
        })
    }
}
```

**Error Testing with Testify:**
```go
func TestCreateIgnoreConfig_WithValidationErrors(t *testing.T) {
    memFs := afero.NewMemMapFs()
    stubs := gostub.Stub(&fs, memFs)
    defer stubs.Reset()

    invalidPolicies := []IgnoredPolicy{
        {Namespace: "", Name: "test"}, // missing namespace
        {Namespace: "test", Name: ""}, // missing name
    }

    paths, err := createIgnoreConfig(invalidPolicies, "/tmp")
    
    require.Error(t, err)
    assert.Nil(t, paths)
    assert.Contains(t, err.Error(), "namespace is required", "should validate namespace")
}
```

#### 5.6 Test Coverage Goals
- **Unit Tests**: >90% coverage focusing on individual function behaviors
- **Integration Tests**: Cover all major user workflows and error scenarios
- **Behavior Verification**: Each test should verify a specific behavior or outcome
- **Mock Verification**: Verify interactions with external dependencies (command execution, file downloads)
- **Error Path Testing**: Comprehensive testing of error conditions and recovery

#### 5.7 Key Testing Principles Summary

**Follow Project Conventions:**
- Use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/require`
- Use `github.com/spf13/afero` with `afero.NewMemMapFs()` for file system testing
- Use `github.com/prashantv/gostub` for stubbing global variables
- Follow the table-driven test pattern established in `pkg/tflint/` tests

**TDD Workflow:**
1. Write failing test first (Red)
2. Write minimal code to pass (Green)
3. Refactor while keeping tests green
4. Focus on testing behavior, not implementation details

**Assertion Guidelines:**
- `require.*` for critical assertions that should stop test execution
- `assert.*` for non-critical assertions
- Always include descriptive error messages in assertions
- Use specific assertion methods (`assert.Len`, `assert.Contains`, etc.)

**Mock Strategy:**
- Mock external dependencies (file system, command execution, network calls)
- Use dependency injection patterns for testability
- Verify mock interactions to ensure correct behavior
- Prefer behavior verification over state verification

#### 5.8 Documentation Updates
- Update README.md with conftest_scan tool documentation
- Add usage examples
- Document policy categories and configuration options

## Implementation Considerations

### 0. Test-Driven Development Mandate
**ALL implementation must follow TDD principles:**
1. **Red**: Write a failing test that describes the desired behavior
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Improve code while keeping tests green
4. **Repeat**: Continue cycle for each new behavior

**Test Behavior, Not Implementation:**
- Focus on what the function should do, not how it does it

### 0.1 Testing Framework and Assertion Patterns
**Follow Existing Project Conventions:**

Based on the established patterns in `pkg/tflint/` tests, use:

```go
import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/spf13/afero"
    "github.com/prashantv/gostub"
)
```

#### Assertion Guidelines:
- **`require.NoError(t, err)`**: For critical assertions that should stop the test if they fail
- **`assert.NoError(t, err)`**: For non-critical error checks
- **`require.Error(t, err)`**: When expecting an error and test should stop if no error occurs
- **`assert.Equal(t, expected, actual)`**: For value comparisons
- **`assert.Len(t, slice, expectedLength)`**: For slice/map length assertions
- **`assert.Contains(t, container, item)`**: For membership checks
- **`assert.NotContains(t, container, item)`**: For exclusion checks

#### File System Testing Pattern:
```go
func TestExampleWithFileSystem(t *testing.T) {
    // Use memory filesystem for testing
    memFs := afero.NewMemMapFs()
    stubs := gostub.Stub(&fs, memFs)
    defer stubs.Reset()
    
    // Create test files/directories
    require.NoError(t, fs.MkdirAll("/test/dir", 0755))
    require.NoError(t, afero.WriteFile(fs, "/test/file.txt", []byte("content"), 0644))
    
    // Run test logic
    result, err := functionUnderTest("/test/dir")
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
}
```

#### Mock Strategy Pattern:
```go
type MockCommandExecutor struct {
    commands []string
    outputs  []string
    errors   []error
    callCount int
}

func (m *MockCommandExecutor) ExecuteCommand(dir, command string) (stdout, stderr string, err error) {
    // Record command for verification
    m.commands = append(m.commands, command)
    
    // Return predetermined response for test scenarios
    if m.callCount < len(m.outputs) {
        output := m.outputs[m.callCount]
        err := m.errors[m.callCount]
        m.callCount++
        return output, "", err
    }
    return "", "", fmt.Errorf("unexpected call")
}
```

#### Test Structure Pattern:
```go
func TestFunction_WithCondition_ExpectedBehavior(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(fs afero.Fs) // filesystem setup
        input    InputType
        wantErr  bool
        errMsg   string
        expected ExpectedType
    }{
        {
            name: "should succeed with valid input",
            setup: func(fs afero.Fs) {
                require.NoError(t, fs.MkdirAll("/test", 0755))
            },
            input:    validInput,
            wantErr:  false,
            expected: expectedOutput,
        },
        {
            name:    "should fail with invalid input",
            setup:   func(fs afero.Fs) {}, // no setup needed
            input:   invalidInput,
            wantErr: true,
            errMsg:  "expected error message",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            memFs := afero.NewMemMapFs()
            stubs := gostub.Stub(&fs, memFs)
            defer stubs.Reset()
            
            if tt.setup != nil {
                tt.setup(fs)
            }

            // Execute
            result, err := functionUnderTest(tt.input)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Equal(t, tt.errMsg, err.Error())
                }
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

#### Configuration Test Pattern (Following `pkg/tflint/config_test.go`):
```go
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
        })
    }
}
```

### 0.2 Behavior-Focused Test Examples for Conftest

#### Updated TDD Test Cases for Types (`pkg/conftest/types_test.go`):
```go
func TestConftestScanParam_Validation(t *testing.T) {
    tests := []struct {
        name    string
        param   ConftestScanParam
        wantErr bool
        errMsg  string
    }{
        {
            name: "mutually exclusive - both predefined alias and policy urls provided",
            param: ConftestScanParam{
                PreDefinedPolicyLibraryAlias: "aprl",
                PolicyUrls:                   []string{"git::https://example.com/policies.git"},
                PlanFile:                     "plan.json",
            },
            wantErr: true,
            errMsg:  "predefined_policy_library_alias and policy_urls are mutually exclusive",
        },
        // ... other test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.param.Validate()
            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Equal(t, tt.errMsg, err.Error())
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestPolicySource_InternalVsPublicFields(t *testing.T) {
    source := PolicySource{
        OriginalURL:  "git::https://example.com/policies.git",
        ResolvedPath: "/tmp/policies", // Internal field
        Type:         "directory",     // Internal field
        PolicyCount:  5,
    }

    // Test JSON marshaling - internal fields should be omitted
    data, err := json.Marshal(source)
    require.NoError(t, err)

    // Parse back to map to check which fields are present
    var jsonMap map[string]interface{}
    err = json.Unmarshal(data, &jsonMap)
    require.NoError(t, err)

    // Check that public fields are present
    assert.Contains(t, jsonMap, "original_url")
    assert.Contains(t, jsonMap, "policy_count")

    // Check that internal fields are omitted
    assert.NotContains(t, jsonMap, "resolved_path", "resolved_path should be omitted from JSON (internal field)")
    assert.NotContains(t, jsonMap, "type", "type should be omitted from JSON (internal field)")
}
```

#### Updated TDD Test Cases for Configuration (`pkg/conftest/config_test.go` - WRITE FIRST):
```go
func TestResolvePolicyUrls_WithMemoryFs(t *testing.T) {
    tests := []struct {
        name                string
        setupFs             func(fs afero.Fs)
        predefinedAlias     string
        customUrls          []string
        wantErr             bool
        errMsg              string
        expectedUrlCount    int
    }{
        {
            name:                "should resolve APRL predefined policies",
            predefinedAlias:     "aprl",
            wantErr:             false,
            expectedUrlCount:    1,
        },
        {
            name:                "should resolve all predefined policies by default",
            predefinedAlias:     "",
            wantErr:             false,
            expectedUrlCount:    2, // both APRL and AVMSEC
        },
        {
            name:       "should resolve custom policy URLs",
            customUrls: []string{"git::https://example.com/policies.git"},
            wantErr:    false,
            expectedUrlCount: 1,
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
            urls, err := resolvePolicyUrls(tt.predefinedAlias, tt.customUrls)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                if tt.errMsg != "" {
                    assert.Equal(t, tt.errMsg, err.Error())
                }
            } else {
                assert.NoError(t, err)
                assert.Len(t, urls, tt.expectedUrlCount)
            }
        })
    }
}
```
- Test inputs and outputs, not internal data structures
- Verify side effects and interactions with external systems
- Mock external dependencies to isolate behavior under test

### 1. Command Execution Pattern
Follow the same dependency injection pattern as TFLint:
```go
type CommandExecutor interface {
    ExecuteCommand(dir, command string) (stdout, stderr string, err error)
}
```

### 2. Error Handling Strategy
- Distinguish between setup errors (invalid URLs, download failures) and policy violations
- Handle conftest's exit codes appropriately (non-zero may indicate violations, not errors)
- Provide meaningful error messages for policy download and compilation failures
- Validate go-getter URLs before attempting downloads

### 3. Resource Management
- Temporary directory management for policies and plans
- Cleanup functions with defer patterns for multiple policy sources
- Resource leak prevention with proper error handling
- Isolation of policy sources to prevent conflicts

### 4. Configuration Flexibility
- Support both predefined AVM policy categories and custom policy URL arrays
- Allow local policy directory inclusion via additional `-p` flags
- Enable selective policy ignoring through required namespace.name composition format
- Mutually exclusive validation between `predefined_policy_library_alias` and `policy_urls`
- Strict validation requiring both namespace and name for ignored policies

### 5. Performance Considerations
- Parallel policy downloads where possible
- Efficient temporary file management with isolated directories
- Minimize redundant downloads through URL validation
- Optimize conftest execution with appropriate flags

## Expected Tool Interface

### Parameters:
- `predefined_policy_library_alias`: "aprl", "avmsec", "all" (default) - mutually exclusive with `policy_urls`
- `policy_urls`: Array of go-getter compatible URLs pointing to policy directories or files
- `plan_file`: Required path to pre-generated Terraform plan file in JSON format
- `local_policy_dirs`: Additional local policy directories to include
- `ignored_policies`: Array of objects with required `namespace` and `name` fields for precise policy ignoring
- `namespaces`: Specific namespaces to test (default: all namespaces)
- `include_default_avm_exceptions`: Boolean to include default AVM exceptions

### Output:
JSON-formatted result with:
- Success status and plan file path information
- List of resolved policy sources with metadata (URL, policy count)
- Policy violations categorized by severity
- Summary statistics (total violations, counts by severity, policies executed)
- Raw conftest output for debugging

### Example Usage:
```json
{
  "policy_urls": [
    "git::https://github.com/Azure/policy-library-avm.git//policy/avmsec",
    "https://example.com/custom-policies.zip",
    "file:///local/path/to/policies"
  ],
  "plan_file": "./tfplan.json",
  "include_default_avm_exceptions": true,
  "local_policy_dirs": ["./local-policies"],
  "namespaces": ["main", "security"],
  "ignored_policies": [
    {
      "namespace": "avmsec",
      "name": "storage_account_https_only"
    },
    {
      "namespace": "aprl", 
      "name": "vm_backup_enabled"
    }
  ]
}
```

Alternatively using predefined policy library alias:
```json
{
  "predefined_policy_library_alias": "all",
  "plan_file": "./tfplan.json",
  "include_default_avm_exceptions": true,
  "ignored_policies": [
    {
      "namespace": "avmsec",
      "name": "storage_account_https_only"
    }
  ]
}
```

## Dependencies to Add

### Go Modules:
- Leverage existing `go-getter` from TFLint implementation for policy downloads
- JSON parsing libraries for conftest output processing
- File system utilities for policy source detection and management

### External Dependencies:
- `conftest` binary (document installation requirements and version compatibility)
- Pre-generated Terraform plan files in JSON format (users must generate these separately)
- Network access for policy repository downloads via go-getter
- Support for various archive formats (.zip, .tar.gz) via go-getter

## Migration Strategy

1. **Phase 1**: Implement core infrastructure without breaking existing functionality ✅
2. **Phase 2**: Add configuration management with thorough testing ✅
3. **Phase 3**: Integrate MCP tool interface 🔄
4. **Phase 4**: Add advanced features incrementally (Pending)
5. **Phase 5**: Comprehensive testing and documentation (Pending)

## Success Criteria

1. ✅ Tool successfully executes conftest scans with multiple policy sources via individual `-p` flags
2. ✅ Support for all go-getter URL formats (git, HTTP, local files, archives)
3. ✅ Automatic detection of policy source types (directory vs. file)
4. ✅ Proper integration with AVM policy repositories and exception handling
5. ✅ Structured JSON output with policy source metadata for AI agent consumption
6. ✅ Comprehensive error handling for URL validation, downloads, and conftest execution
7. ✅ Mutually exclusive parameter validation (predefined_policy_library_alias vs. policy_urls)
8. ✅ Strict ignored policy specification with required namespace.name composition support
9. ✅ Unit and integration test coverage >80% with mock policy sources
10. ✅ Documentation updated with multiple policy URL usage examples including ignored policies
11. ✅ Consistent with existing codebase patterns and quality standards

## Notes

- Follow the established patterns from TFLint implementation for consistency
- Ensure proper separation of concerns between MCP interface and core logic
- Support all go-getter URL formats for maximum flexibility in policy source specification
- Each policy URL should result in a separate `-p` flag in the conftest command
- Implement robust URL validation and error handling for policy downloads
- Support strict ignored policy specification with required namespace.name composition (e.g., "avmsec.storage_account_https_only")
- Validate that both namespace and name are provided for ignored policies to ensure precision
- Remove support for namespace-less policy names to enforce clear policy identification
- Consider multiple strategies for policy ignoring (CLI flags, generated ignore policies, result filtering)
- Consider policy source caching for repeated URLs to improve performance
- Maintain backwards compatibility with existing tools
- Document policy URL format requirements and provide clear examples with ignored policies
- Handle edge cases like empty policy directories, invalid Rego files, and network failures
