# TFLint MCP Tool Development Action Plan

## Project Overview
Develop a new MCP tool that enables AI agents to execute TFLint scanning on Terraform code in specified directories. The tool will simulate the behavior of the AVM script `run-tflint.sh` with a simplified parameter structure.

## Current Status Analysis

### Completed
- ✅ TFLint binary installation added to Dockerfile
- ✅ Complete MCP framework and tool registration mechanism in place
- ✅ Existing tool patterns available as reference (`pkg/tool/` directory implementations)

### Development Methodology
- **TDD (Test-Driven Development)**: Follow TDD approach with test-first development
- **Small Steps**: Implement functionality incrementally with small, safe changes
- **Unit Testing Focus**: Prioritize unit tests over integration tests for faster feedback
- **Safety First**: Ensure each step is tested and working before proceeding to next

### Requirements Based on User Input
The new tool should have the following simplified parameter structure (all parameters are optional):
1. **Category Type**: Either "reusable" (corresponding to root) or "example", defaults to "reusable"
2. **Target Directory**: Path to the directory to scan, defaults to current directory
3. **Optional TFLint Config File Path**: Custom tflint configuration file
4. **Optional Ignored Rule IDs**: List of tflint rule IDs to ignore

## Implementation Plan

### Phase 1: Basic Structure Setup
1. **Create TFLint Package Structure**
   ```
   pkg/tflint/
   ├── config.go          # TFLint configuration management
   ├── scanner.go         # Scanning logic implementation
   ├── types.go           # Data structure definitions
   └── utils.go           # Utility functions
   ```

2. **Define Data Structures**
   ```go
   type TFLintScanParam struct {
       Category      string   `json:"category,omitempty"`          // "reusable" or "example", defaults to "reusable"
       TargetPath    string   `json:"target_path,omitempty"`       // Directory to scan, defaults to current directory
       ConfigFile    string   `json:"config_file,omitempty"`       // Optional custom config file path
       IgnoredRules  []string `json:"ignored_rules,omitempty"`     // Optional list of rule IDs to ignore
   }
   
   type TFLintScanResult struct {
       Success   bool           `json:"success"`
       Issues    []TFLintIssue  `json:"issues,omitempty"`
       Output    string         `json:"output"`
       Summary   ScanSummary    `json:"summary"`
   }
   ```

### Phase 2: Core Functionality Implementation
3. **Configuration Management Module (`config.go`)**
   - `getDefaultConfigURL()` - Get appropriate config URL based on category
   - `downloadConfig()` - Download default configuration file
   - `modifyConfigForIgnoredRules()` - Modify config to set ignored rules' enabled = false
   - `mergeCustomConfig()` - Merge custom configuration file if provided
   - `setupConfig()` - Setup final configuration with all overrides and ignored rules

4. **Scanning Engine (`scanner.go`)**
   - `setupTempConfigDir()` - Create temporary directory for TFLint configuration files
   - `downloadConfigToTemp()` - Download configuration file to temporary directory
   - `executeInitialization()` - Run tflint --init in target directory with temp config
   - `executeScan()` - Execute TFLint scan in target directory with temp config
   - `parseResults()` - Parse TFLint output into structured format
   - `cleanupTempConfig()` - Clean up temporary configuration files

5. **Utility Functions (`utils.go`)**
   - File operation utilities
   - Command execution utilities
   - Error handling utilities

### Phase 3: MCP Tool Integration
6. **Create MCP Tool (`pkg/tool/tflint_scan.go`)**
   - Define tool parameter structure
   - Implement tool handler function
   - Integrate with MCP server

7. **Register Tool in Server (`pkg/registry.go`)**
   - Add tool definition and parameter schema
   - Set tool annotations and descriptions

### Phase 4: Testing and Documentation
8. **Write Tests (TDD Approach)**
   - **Test-First Development**: Write unit tests before implementing each function
   - **Small Increments**: Test and implement one function at a time
   - **Unit Test Focus**: Prioritize isolated unit tests with mocks/stubs
   - **Coverage Areas**:
     - Configuration parsing and modification (`config_test.go`)
     - Rule enabling/disabling logic (`config_test.go`)
     - Command execution utilities (`utils_test.go`)
     - Result parsing functions (`scanner_test.go`)
   - **Limited Integration Tests**: Only for end-to-end MCP tool validation

9. **Update Documentation**
   - Update `readme.md` with new tool description
   - Add usage examples

## Technical Implementation Details

### MCP Tool Parameter Design
```go
type TFLintScanParam struct {
    Category     string   `json:"category,omitempty" jsonschema:"enum=reusable,example;description=Type of Terraform code to scan: 'reusable' for reusable modules, 'example' for example code. Defaults to 'reusable'"`
    TargetPath   string   `json:"target_path,omitempty" jsonschema:"description=Path to the directory containing Terraform code to scan. Defaults to current directory"`
    ConfigFile   string   `json:"config_file,omitempty" jsonschema:"description=Optional path to custom TFLint configuration file"`
    IgnoredRules []string `json:"ignored_rules,omitempty" jsonschema:"description=Optional list of TFLint rule IDs to ignore during scanning"`
}
```

### Configuration URL Mapping
- **Reusable** → `https://raw.githubusercontent.com/Azure/tfmod-scaffold/main/avm.tflint.hcl`
- **Example** → `https://raw.githubusercontent.com/Azure/tfmod-scaffold/main/avm.tflint_example.hcl`

### Scan Result Structure
```go
type TFLintScanResult struct {
    Success    bool            `json:"success"`
    Category   string          `json:"category"`
    TargetPath string          `json:"target_path"`
    Issues     []TFLintIssue   `json:"issues,omitempty"`
    Output     string          `json:"output"`
    Summary    ScanSummary     `json:"summary"`
}

type TFLintIssue struct {
    Rule     string `json:"rule"`
    Severity string `json:"severity"`
    Message  string `json:"message"`
    Range    Range  `json:"range"`
}

type ScanSummary struct {
    TotalIssues   int `json:"total_issues"`
    ErrorCount    int `json:"error_count"`
    WarningCount  int `json:"warning_count"`
    InfoCount     int `json:"info_count"`
}
```

### Key Implementation Points
1. **Configuration Isolation**: Use temporary directory for TFLint configuration files to avoid polluting target directory
2. **Direct Scanning**: Scan target directory directly without copying code files
3. **Error Handling**: Detailed error information and recovery mechanisms
4. **Configuration Flexibility**: Support custom configuration files and rule ignoring
5. **Output Format**: Structured JSON output for AI comprehension
6. **Resource Cleanup**: Proper cleanup of temporary configuration files

### Configuration Processing Flow
1. **Download Base Config**: Download the appropriate config file based on category
2. **Apply Ignored Rules**: For each rule in `ignored_rules`, ensure it's set to `enabled = false`
3. **Merge Custom Config**: If a custom config file is provided, merge it with the base config
4. **Final Config Generation**: Generate the final TFLint configuration file in temp directory

### Ignored Rules Implementation
When `ignored_rules` contains rule IDs (e.g., `["terraform_unused_declarations", "terraform_deprecated_syntax"]`), the tool will:
1. Parse the downloaded base configuration file
2. For each rule in the `ignored_rules` list:
   - If the rule exists in the config, set `enabled = false`
   - If the rule doesn't exist, add it with `enabled = false`
3. Write the modified configuration to the temporary directory

Example configuration modification:
```hcl
# Original config
rule "terraform_unused_declarations" {
  enabled = true
}

# After applying ignored_rules = ["terraform_unused_declarations"]
rule "terraform_unused_declarations" {
  enabled = false
}
```

### Dependencies
- Ensure `tflint` binary is available in container
- Network access required for downloading configuration files
- File system read/write permissions needed

## Expected Outcomes
Upon completion, AI agents will be able to:
1. Execute AVM-standard TFLint scans on Terraform projects based on category
2. Receive structured scan results and issue reports
3. Use custom configurations and ignore specific rules
4. Get actionable feedback for code improvement

## Acceptance Criteria
- [ ] Successfully scan "reusable" category Terraform projects
- [ ] Successfully scan "example" category Terraform projects
- [ ] Support custom TFLint configuration files
- [ ] Support ignoring specified rule IDs
- [ ] Return structured scan results
- [ ] Comprehensive error handling and logging
- [ ] Tool works properly in VS Code MCP environment
- [ ] Complete documentation updates

## Tool Usage Examples

### Minimal Scan (using all defaults)
```json
{
}
```
This will scan the current directory using "reusable" category.

### Basic Reusable Module Scan
```json
{
  "target_path": "/workspace/terraform-module"
}
```

### Example Category Scan
```json
{
  "category": "example",
  "target_path": "/workspace/examples/basic"
}
```

### Scan with Custom Config
```json
{
  "target_path": "/workspace/examples/basic",
  "config_file": "/workspace/.tflint.hcl"
}
```

### Scan with Ignored Rules
```json
{
  "category": "reusable",
  "target_path": "/workspace/modules/network",
  "ignored_rules": ["terraform_unused_declarations", "terraform_deprecated_syntax"]
}
```
