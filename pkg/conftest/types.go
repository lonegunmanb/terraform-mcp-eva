package conftest

import (
	"fmt"
)

// ConftestScanParam - Input parameters for conftest scanning
type ConftestScanParam struct {
	PreDefinedPolicyLibraryAlias string          `json:"predefined_policy_library_alias,omitempty"` // "aprl", "avmsec", "all" - mutually exclusive with PolicyUrls
	PolicyUrls                   []string        `json:"policy_urls,omitempty"`                     // Array of policy URLs in go-getter format
	PlanFile                     string          `json:"plan_file"`                                 // Required pre-generated plan file path (JSON format)
	LocalPolicyDirs              []string        `json:"local_policy_dirs,omitempty"`               // Local policy directories to include
	IgnoredPolicies              []IgnoredPolicy `json:"ignored_policies,omitempty"`                // Policies to ignore with namespace and name
	Namespaces                   []string        `json:"namespaces,omitempty"`                      // Specific namespaces to test (default: all)
	IncludeDefaultAVMExceptions  bool            `json:"include_default_avm_exceptions,omitempty"`  // Whether to download and include default AVM exceptions
}

// Validate validates the ConftestScanParam
func (p *ConftestScanParam) Validate() error {
	// Check mutually exclusive parameters
	if p.PreDefinedPolicyLibraryAlias != "" && len(p.PolicyUrls) > 0 {
		return fmt.Errorf("predefined_policy_library_alias and policy_urls are mutually exclusive")
	}

	// Check required fields
	if p.PlanFile == "" {
		return fmt.Errorf("plan_file is required")
	}

	// Validate predefined alias if provided
	if p.PreDefinedPolicyLibraryAlias != "" {
		validAliases := []string{"aprl", "avmsec", "all"}
		valid := false
		for _, alias := range validAliases {
			if p.PreDefinedPolicyLibraryAlias == alias {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid predefined_policy_library_alias")
		}
	}

	// Validate ignored policies
	for i, policy := range p.IgnoredPolicies {
		if err := policy.Validate(); err != nil {
			return fmt.Errorf("ignored_policies[%d]: %w", i, err)
		}
	}

	return nil
}

// IgnoredPolicy represents a policy to be ignored during conftest execution
type IgnoredPolicy struct {
	Namespace string `json:"namespace"` // Required policy namespace (e.g., "avmsec", "aprl", "custom")
	Name      string `json:"name"`      // Policy name/rule name (e.g., "storage_account_https_only")
}

// Validate validates the IgnoredPolicy
func (p *IgnoredPolicy) Validate() error {
	if p.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// ConftestScanResult - Output structure
type ConftestScanResult struct {
	Success       bool              `json:"success"`
	PlanFile      string            `json:"plan_file"`
	PolicySources []PolicySource    `json:"policy_sources"` // Details of resolved policy sources
	Violations    []PolicyViolation `json:"violations,omitempty"`
	Warnings      []PolicyWarning   `json:"warnings,omitempty"`
	Output        string            `json:"output"`
	Summary       ConftestSummary   `json:"summary"`
}

// PolicySource - Information about a resolved policy source
type PolicySource struct {
	OriginalURL  string `json:"original_url"` // Original go-getter URL
	ResolvedPath string `json:"-"`            // Internal use only - not exposed in JSON
	Type         string `json:"-"`            // Internal use only - not exposed in JSON
	PolicyCount  int    `json:"policy_count"` // Number of policies found
}

// PolicyViolation - Individual policy violation
type PolicyViolation struct {
	Policy    string `json:"policy"`
	Rule      string `json:"rule"`
	Message   string `json:"message"`
	Namespace string `json:"namespace"`
	Severity  string `json:"severity"`
	Resource  string `json:"resource,omitempty"`
}

// PolicyWarning - Individual policy warning
type PolicyWarning struct {
	Policy    string `json:"policy"`
	Rule      string `json:"rule"`
	Message   string `json:"message"`
	Namespace string `json:"namespace"`
	Resource  string `json:"resource,omitempty"`
}

// ConftestSummary - Scan summary statistics
type ConftestSummary struct {
	TotalViolations int `json:"total_violations"`
	ErrorCount      int `json:"error_count"`
	WarningCount    int `json:"warning_count"`
	InfoCount       int `json:"info_count"`
	PoliciesRun     int `json:"policies_run"`
}
