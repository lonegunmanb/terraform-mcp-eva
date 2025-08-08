# TFLint MCP Tool Development Action Plan

> NOTE (Enhancement In Progress): A new feature "RemoteConfigUrl" for the `tflint_scan` tool is being planned (see next section). The original completed plan is retained below for historical context.

---

## Enhancement: Support Custom Remote TFLint Configuration via `RemoteConfigUrl`

### Current Progress Snapshot (Aug 2025 - Updated After GetFile Implementation)
Status legend: ✅ done | 🔄 in-progress | ⏭ not started

| Item | Status | Notes |
|------|--------|-------|
| Param `RemoteConfigUrl` added to `ScanParam` | ✅ | Field + jsonschema tag present |
| Mutual exclusivity validation (category vs remote) | ✅ | Error returned early in `Scan` |
| Temp dir scaffolding | ✅ | `setupRemoteConfig` implemented |
| Remote getter abstraction | ✅ | `remote_getter.go` with real go-getter GetFile impl |
| Single file enforcement | ✅ | Switched to direct `GetFile` -> always `remote.tflint.hcl` |
| Directory/root repo error | ✅ | Git root heuristic maintained (root w/out .git//path rejected) |
| Timeout env var parsing | ✅ | Fallback 60s; test covers invalid -> fallback |
| Scheme whitelist | ❌ (removed) | Allow whatever go-getter supports |
| Unsupported scheme test | ❌ (removed) | Behavior delegated to go-getter |
| Git root heuristic (.git//) | ✅ | Still enforced |
| Custom config merge (remote) | ❌ (dropped) | Scope reduced: remote merging removed per refactor |
| Custom config merge (category) | ❌ (dropped) | `setupConfig` no longer merges custom config |
| Tests: directory error | ✅ | Updated expectation (root heuristic message) |
| Tests: git subfolder (no .hcl) | ✅ | Zero file case handled before switch; now always single file path -> heuristic still tested |
| Tests: happy path single file | ✅ | Uses mock getter writing remote.tflint.hcl |
| Tests: timeout (env invalid fallback) | ✅ | Covered (no enforced timeout scenario yet) |
| Tests: multiple .hcl files error | ✅ (legacy scenario) | Multi file test now simulates dual write pre-discovery; with GetFile path always single, test ensures earlier logic still guards multi writes |
| Tests: network / fetch failure propagation | ✅ | Mock returns error -> wrapped message asserted |
| Tests: ignored rules precedence | ✅ | Ignored rules applied via CLI flags test passes |
| README & tool schema update | ⏭ | Pending exposure of `remote_config_url` param |
| Invalid HCL sanitization | N/A (merge removed) | Sanitization helper remains but unused currently |
| Error message content privacy | ✅ (current scope) | No raw content surfaced without merge parsing |

### Adjusted Decisions
* Removed prior requirement to restrict schemes; we now defer to go-getter's supported protocols (git::, http(s), s3, gcs, etc.). Responsibility shifts to callers to use safe sources.
* Unsupported scheme test removed accordingly.
* Timeout env var implemented earlier than original step ordering.
* Custom config merge now handled fully inside setup functions; duplicate merge in `Scan` removed.

### Goal
Add an optional `RemoteConfigUrl` parameter to `TFLintScanParam` allowing callers to specify an arbitrary remote TFLint configuration fetched using [go-getter](https://github.com/hashicorp/go-getter/tree/v2.2.3). This must be **mutually exclusive** with the existing `Category` parameter. If `RemoteConfigUrl` is set, we skip category-based resolution and download the configuration via go-getter into the temp config directory before proceeding with rule ignoring / merging logic.

### Motivation
- Provide greater flexibility beyond the two predefined category configs.
- Enable experimentation with organization-specific or branch-specific configs.
- Reuse complex shared configs stored in Git, HTTP, S3, or other go-getter supported sources.

### Functional Requirements (Revised)
1. (✅) New optional field: `RemoteConfigUrl string` (JSON name: `remote_config_url`).
2. (✅) Mutual exclusivity validation: error if BOTH `Category` and `RemoteConfigUrl` are set. If both empty, default category = `reusable`.
3. (✅) When `RemoteConfigUrl` provided, skip category resolution.
4. (🔄) Support all go-getter supported schemes transparently (no internal whitelist). Real fetch pending.
5. (✅ initial / 🔄 final) Result MUST yield exactly one `.hcl` file. Current implementation enforces post-download discovery; real getter integration pending.
6. (✅) Preserve category behavior when `remote_config_url` absent.
7. (🔄) Clear error messages for: unreachable URL (after real getter), directory (done), multi / zero `.hcl` (done for zero; pending multi test), invalid HCL (pending sanitization), timeout (baseline implemented, need explicit timeout test).
8. (⏭) Update README & tool schema docs: describe single-file requirement, timeout env var, precedence of custom config & ignored rules.
9. (🔄) Ensure error sanitization: invalid HCL should not echo file contents (only URL context).
10. (⏭) Precedence tests: remote base < custom config override < ignored rules (ignored rules final).

### Non-Functional Requirements
- Maintain test coverage & follow TDD.
- Minimal disruption to existing public API (additive only).
- Add dependency: `github.com/hashicorp/go-getter/v2` with version pin in `go.mod`.
- Ensure deterministic temp paths & cleanup unaffected.

### Security / Safety Considerations
* Timeout enforced (default 60s) with env override `TFLINT_REMOTE_CONFIG_TIMEOUT_SECONDS` (>0). Invalid / zero / negative => fallback. Need test for successful override & for timeout triggering cancellation.
* Relying on go-getter introduces broader protocol surface (git, s3, gcs, etc.); caller responsibility to vet sources. No internal allowlist.
* Planned: sanitize invalid HCL parse errors to avoid leaking remote file content.

### Edge Cases
- Both `Category` and `RemoteConfigUrl` set -> validation error.
- Only `RemoteConfigUrl` set -> use it.
- Empty `RemoteConfigUrl` & empty `Category` -> default Category = `reusable`.
- Remote resolves to a directory (e.g., repo root) -> error (must specify path to a single file).
- Remote HCL invalid -> error reports remote URL only (no file content echoed).
- Network failure / 404 -> error with original URL and wrapped cause.
- Git repo with subdir file: Support syntax `git::https://...//path/to/config.tflint.hcl?ref=tag`.

### High-Level Design Changes (Status Update)
Modules impacted & implementation status:
- ✅ `pkg/tflint/types.go`: `RemoteConfigUrl` field added to `ScanParam` with jsonschema tag.
- ✅ `pkg/tflint/scanner.go`: Mutual exclusivity validation implemented; branches to `setupRemoteConfig` vs `setupConfig`.
- ✅ `pkg/tflint/config.go`: `setupRemoteConfig` function implemented with timeout, discovery, merge logic.
- ✅ `pkg/tflint/remote_getter.go`: `RemoteGetter` interface + noop default implementation.
- ❌ `pkg/tflint/utils.go`: Scheme validation removed (deferred to go-getter).
- 🔄 Real go-getter implementation: Interface ready, concrete implementation pending.
- ⏭ `pkg/tool/tflint_scan.go`: Tool schema update pending (add remote_config_url param).

### Updated Forward Plan (Post-Refactor)
1. Expose `remote_config_url` in MCP tool param & registry schema (mutual exclusivity note with `category`).
2. Update README: document remote mode, single-file constraint (always saved as `remote.tflint.hcl`), timeout env var, ignored rules via CLI flags.
3. Decide whether to re-introduce custom config merging (currently removed) or formally drop from spec; if keeping removed, clean old references and adjust public docs.
4. (Optional) Add explicit timeout trigger test (mock sleeps > configured small timeout) to assert cancellation path.
5. (Optional) Remove unused sanitization helper or wire it where future parsing/validation might occur.
6. (Stretch) Logging & potential caching of remote file (URL+ref).

Testing Principle: Focus tests on observable behavior and externally visible outcomes, not on trivial data assignments or mere struct field presence (the compiler already guarantees field existence).

Assertion Guideline: Prefer github.com/stretchr/testify/assert or require for all test validations (choose require for fatal conditions, assert for non-fatal) instead of manual if err != nil / t.Fatalf patterns.

#### Original Step Breakdown (with current status):
1. ✅ Test Scaffold – Param Validation: Added `TestScanParamCategoryAndRemoteConfigUrlMutualExclusion`.
2. ✅ Implement Param Struct Update & validation logic: Field in `types.go`, validation in `scanner.go`.
3. ✅ Tests: Default Behavior Unchanged: Existing category tests remain green.
4. ✅ Tests: Remote URL Happy Path: `TestScanRemoteConfigSingleFileSuccess` with mock getter.
5. 🔄 Implement go-getter integration: Interface ready, real implementation pending.
6. ✅ Tests: Remote URL Directory Result: `TestScanRemoteConfigDirectoryError` (git root heuristic).
7. ✅ Implement directory detection & error: Git root heuristic + single `.hcl` discovery.
8. ❌ Tests: Unsupported Scheme: Removed (all go-getter schemes now allowed).
9. ❌ Implement scheme validation: Removed to defer to go-getter capabilities.
10. ⏭ Tests: Custom Config Merge Still Applies: Pending (merge logic implemented but not specifically tested for remote path).
11. ✅ Implement merge ordering: `mergeOptionalCustomConfig` used in both category & remote paths.
12. ⏭ Tests: Ignored Rules Override Merged Config: Pending precedence test.
13. ✅ Implement ignored-rules-last logic: Already in place via existing ignored rules parameter handling.
14. ⏭ Tests: Invalid HCL error sanitization: Pending.
15. ⏭ Implement error wrapping ensuring content not included: Pending.
16. ⏭ Tests: Network Failure: Pending (after real getter).
17. ✅ Implement context timeout & error propagation: Done with env var override.
18. ✅ Refactor: Extract download logic: `setupRemoteConfig`, `discoverSingleHCLFile`, helpers extracted.
19. ⏭ Tests: Timeout Env Var Override: Basic env parsing tested; need explicit timeout trigger test.
20. ⏭ Documentation Update Tests: Optional.
21. ⏭ Update README & tool descriptions: Pending.
22. ⏭ Final Lint / Vet / Race checks: Pending.

### Test Utilities / Mocks (Status Update)
- ✅ Interface wrapper over go-getter: `RemoteGetter` interface in `remote_getter.go`.
- ✅ Provide default production implementation: `noopGetter` as placeholder; real implementation pending.
- ✅ Global variable override capability: `remoteConfigGetter` variable allows test stubbing via gostub.
- ✅ Mock implementation created: `mockRemoteGetter` in tests with configurable file creation behavior.

### Revised Acceptance Criteria
- [ ] Tool layer exposes `remote_config_url` & docs updated.
- [x] Mutual exclusivity enforced.
- [x] Category default behavior intact when remote absent.
- [x] Real go-getter `GetFile` implemented.
- [x] Single file always produced (`remote.tflint.hcl`).
- [x] Git root heuristic error for repo root without file path.
- [ ] Timeout cancellation test (explicit) (optional).
- [x] Ignored rules applied via CLI flags.
- [x] Network failure surfaced with wrapped error.
- [ ] README updated (remote usage, timeout env var, removed custom merge note).
- [ ] Sanitization helper decision (remove or document N/A).

### Open Questions / Decisions (Revised)
- Retain removal of custom config merging or reintroduce? (Currently removed — update README accordingly.)
- Add explicit timeout failure test? (Not critical but increases confidence.)
- Keep sanitization helper or prune?

### Next Steps (Immediate Priority)
1. **Add go-getter dependency**: `go get github.com/hashicorp/go-getter/v2` and implement real `RemoteGetter`.
2. **Missing tests**: Multiple `.hcl` files error, invalid HCL sanitization, explicit timeout trigger, precedence validation.
3. **Tool schema update**: Add `remote_config_url` parameter to `pkg/tool/tflint_scan.go`.
4. **Documentation**: Update README with new parameter, constraints, and examples.

---

## Project Overview
✅ **COMPLETED** - Develop a new MCP tool that enables AI agents to execute TFLint scanning on Terraform code in specified directories. The tool simulates the behavior of the AVM script `run-tflint.sh` with a simplified parameter structure.

## Implementation Summary

### Successfully Implemented Features
- ✅ **Complete TFLint Package Structure** - Created comprehensive package with proper separation of concerns
- ✅ **Configuration Management** - HCL config download, manipulation, and temporary setup
- ✅ **Scanning Engine** - TFLint command execution and JSON output parsing  
- ✅ **MCP Tool Integration** - Full Model Context Protocol integration with proper schema
- ✅ **Comprehensive Testing** - Extensive unit tests with mocking and integration tests
- ✅ **Registry Integration** - Tool registered and available in MCP server

### Technical Implementation Details

#### Package Structure (Enhanced)
```
pkg/tflint/
├── config.go              # ✅ TFLint configuration management with HCL support
├── config_test.go          # ✅ Configuration tests with afero filesystem abstraction
├── scanner.go              # ✅ Core scanning logic and command execution
├── scanner_test.go         # ✅ Scanner tests with comprehensive mocking
├── types.go                # ✅ Data structure definitions with JSON schema + RemoteConfigUrl
├── types_remote_config_test.go # ✅ Remote config specific tests (mutual exclusivity, directory error, happy path)
├── remote_getter.go        # ✅ RemoteGetter interface and noop implementation
├── utils.go                # ✅ Utility functions for defaults and validation
└── utils_test.go           # ✅ Utility function tests

pkg/tool/
├── tflint_scan.go          # 🔄 MCP tool wrapper and integration (needs remote_config_url param)
└── tflint_scan_test.go     # ✅ MCP tool tests with mocking
```

#### Core Features Implemented
1. **Category-Based Configuration** - Maps "reusable" and "example" to appropriate TFLint configs
2. **Dynamic Configuration Download** - Downloads configs from Azure/tfmod-scaffold repository
3. ✅ **Remote Configuration Support** - Added `RemoteConfigUrl` parameter with mutual exclusivity validation
4. ✅ **Remote Getter Abstraction** - Interface-based design allowing real go-getter integration and test mocking
5. ✅ **Single File Enforcement** - Validates remote downloads result in exactly one `.hcl` file
6. ✅ **Timeout Management** - Configurable via `TFLINT_REMOTE_CONFIG_TIMEOUT_SECONDS` environment variable
3. **HCL Configuration Manipulation** - Uses hclwrite and hclmerge for rule modifications
4. **Command Execution Abstraction** - Mockable command execution for testing
5. **Comprehensive Error Handling** - Detailed error reporting and recovery
6. **JSON Output Parsing** - Structured parsing of TFLint JSON output
7. **Filesystem Abstraction** - Uses afero for testable filesystem operations

#### Parameters (All Optional)
- **category**: "reusable" (default) or "example" - determines configuration type
- **target_directory**: Directory to scan (defaults to current directory)  
- **custom_config_file**: Path to custom TFLint configuration file
- **ignored_rule_ids**: Array of rule IDs to disable during scanning
- ✅ **remote_config_url**: URL to remote TFLint configuration (mutually exclusive with category; supports all go-getter schemes)

#### Output Structure
```json
{
  "success": true,
  "category": "reusable", 
  "target_path": "/path/to/terraform",
  "issues": [
    {
      "rule": "terraform_unused_declarations",
      "severity": "warning", 
      "message": "variable \"unused_var\" is declared but not used",
      "range": {
        "filename": "variables.tf",
        "start": {"line": 1, "column": 1},
        "end": {"line": 1, "column": 20}
      }
    }
  ],
  "output": "Init: TFLint initialized successfully\nScan: {...}",
  "summary": {
    "total_issues": 1,
    "error_count": 0, 
    "warning_count": 1,
    "info_count": 0
  }
}
```

## Testing Strategy Applied
- **Test-Driven Development** - Tests written before implementation
- **Comprehensive Unit Testing** - All functions and edge cases covered
- **Integration Testing** - End-to-end workflow validation
- **Dependency Injection** - Global variables for easy mocking
- **Filesystem Abstraction** - afero for testable file operations
- **Command Mocking** - Pattern-based command execution mocking

## Tool Registration
The TFLint tool is now registered in the MCP server registry (`pkg/registry.go`) with:
- **Tool Name**: `tflint_scan`
- **Description**: Comprehensive description of capabilities and use cases
- **Schema**: Complete JSON schema for all parameters
- **Annotations**: Proper MCP annotations for tool behavior

## Verification
All tests passing:
- ✅ TFLint package tests (types, utils, config, scanner)
- ✅ Tool package tests (MCP integration)
- ✅ Integration tests (complete workflow)
- ✅ Build verification (successful compilation)

The TFLint MCP tool is now fully functional and ready for use by AI agents to perform static analysis of Terraform code.

---

## Original Development Plan (Completed)

### Phase 1: Basic Structure Setup ✅
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
