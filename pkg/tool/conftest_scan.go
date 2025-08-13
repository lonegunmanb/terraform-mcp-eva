package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/conftest"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ConftestScanParam struct {
	PreDefinedPolicyLibraryAlias string                  `json:"predefined_policy_library_alias,omitempty" jsonschema:"Predefined policy library alias. Supported values: 'aprl' (Azure Proactive Resiliency Library), 'avmsec' (AVM Security policies), or 'all' (both libraries, default). Mutually exclusive with 'policy_urls' (cannot set both). If neither is set, defaults to 'all'."`
	PolicyUrls                   []string                `json:"policy_urls,omitempty" jsonschema:"Array of policy URLs in go-getter format (e.g., git::https://github.com/org/repo.git//policy/path, https://example.com/policies.zip, file:///local/path). Mutually exclusive with 'predefined_policy_library_alias'. Supports git repositories, HTTP/HTTPS URLs, local files, and archive formats."`
	TargetFile                   string                  `json:"target_file" jsonschema:"Required path to target file (Terraform plan file in JSON format or state file). Use relative paths relative to the current Terraform workspace (e.g., './plan.json' for root, './examples/default/plan.json' for AVM module examples). For plan files, generate using: 'terraform plan -out=plan.tfplan && terraform show -json plan.tfplan > plan.json'. For state files, use 'terraform.tfstate' or pull from backend using: 'terraform state pull'."`
	IgnoredPolicies              []ConftestIgnoredPolicy `json:"ignored_policies,omitempty" jsonschema:"Array of policies to ignore during scanning. Each policy must specify both 'namespace' and 'name' for precise identification (e.g., namespace: 'avmsec', name: 'storage_account_https_only')."`
	Namespaces                   []string                `json:"namespaces,omitempty" jsonschema:"Specific policy namespaces to test. If not specified, all namespaces will be tested. Use this to limit scanning to specific policy categories."`
	IncludeDefaultAVMExceptions  *bool                   `json:"include_default_avm_exceptions,omitempty" jsonschema:"Whether to include default Azure Verified Modules (AVM) exceptions. Defaults to true. When true, downloads and includes standard AVM policy exceptions from the official policy library."`
}

type ConftestIgnoredPolicy struct {
	Namespace string `json:"namespace" jsonschema:"Required policy namespace (e.g., 'avmsec', 'aprl', 'custom'). Used together with 'name' to uniquely identify the policy to ignore."`
	Name      string `json:"name" jsonschema:"Required policy rule name (e.g., 'storage_account_https_only', 'vm_backup_enabled'). Used together with 'namespace' to uniquely identify the policy to ignore."`
}

func ConftestScan(_ context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[ConftestScanParam]) (*mcp.CallToolResultFor[any], error) {
	// Convert MCP parameters to conftest scan parameters
	var ignoredPolicies []conftest.IgnoredPolicy
	for _, policy := range params.Arguments.IgnoredPolicies {
		ignoredPolicies = append(ignoredPolicies, conftest.IgnoredPolicy{
			Namespace: policy.Namespace,
			Name:      policy.Name,
		})
	}

	// Set default value for IncludeDefaultAVMExceptions if not explicitly provided
	includeAVMExceptions := true // Default to true
	if params.Arguments.IncludeDefaultAVMExceptions != nil {
		includeAVMExceptions = *params.Arguments.IncludeDefaultAVMExceptions
	}

	scanParams := conftest.ScanParam{
		PreDefinedPolicyLibraryAlias: params.Arguments.PreDefinedPolicyLibraryAlias,
		PolicyUrls:                   params.Arguments.PolicyUrls,
		TargetFile:                   params.Arguments.TargetFile,
		IgnoredPolicies:              ignoredPolicies,
		Namespaces:                   params.Arguments.Namespaces,
		IncludeDefaultAVMExceptions:  includeAVMExceptions,
	}

	// Execute the conftest scan
	result, err := conftest.Scan(scanParams)
	if err != nil {
		return nil, fmt.Errorf("conftest scan failed: %w", err)
	}

	// Convert the result to compact JSON for AI agent efficiency
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scan result: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonResult),
			},
		},
	}, nil
}
