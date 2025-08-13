package conftest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanParam_Validation(t *testing.T) {
	tests := []struct {
		name    string
		param   ScanParam
		wantErr bool
		errMsg  string
	}{
		{
			name: "mutually exclusive - both predefined alias and policy urls provided",
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				PolicyUrls:                   []string{"git::https://example.com/policies.git"},
				TargetFile:                   "plan.json",
			},
			wantErr: true,
			errMsg:  "predefined_policy_library_alias and policy_urls are mutually exclusive",
		},
		{
			name: "valid predefined alias only",
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				TargetFile:                   "plan.json",
			},
			wantErr: false,
		},
		{
			name: "valid policy urls only",
			param: ScanParam{
				PolicyUrls: []string{"git::https://example.com/policies.git"},
				TargetFile: "plan.json",
			},
			wantErr: false,
		},
		{
			name: "missing plan file",
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
			},
			wantErr: true,
			errMsg:  "target_file is required",
		},
		{
			name: "invalid predefined alias",
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "invalid",
				TargetFile:                   "plan.json",
			},
			wantErr: true,
			errMsg:  "invalid predefined_policy_library_alias",
		},
		{
			name: "valid ignored policies with namespace and name",
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				TargetFile:                   "plan.json",
				IgnoredPolicies: []IgnoredPolicy{
					{Namespace: "avmsec", Name: "storage_account_https_only"},
					{Namespace: "aprl", Name: "vm_backup_enabled"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid ignored policy - missing namespace",
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				TargetFile:                   "plan.json",
				IgnoredPolicies: []IgnoredPolicy{
					{Name: "storage_account_https_only"},
				},
			},
			wantErr: true,
			errMsg:  "ignored_policies[0]: namespace is required",
		},
		{
			name: "invalid ignored policy - missing name",
			param: ScanParam{
				PreDefinedPolicyLibraryAlias: "aprl",
				TargetFile:                   "plan.json",
				IgnoredPolicies: []IgnoredPolicy{
					{Namespace: "avmsec"},
				},
			},
			wantErr: true,
			errMsg:  "ignored_policies[0]: name is required",
		},
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

func TestIgnoredPolicy_Validation(t *testing.T) {
	tests := []struct {
		name    string
		policy  IgnoredPolicy
		wantErr bool
		errMsg  string
	}{
		{
			name:   "valid policy with namespace and name",
			policy: IgnoredPolicy{Namespace: "avmsec", Name: "storage_account_https_only"},
		},
		{
			name:    "missing namespace",
			policy:  IgnoredPolicy{Name: "storage_account_https_only"},
			wantErr: true,
			errMsg:  "namespace is required",
		},
		{
			name:    "missing name",
			policy:  IgnoredPolicy{Namespace: "avmsec"},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "empty namespace and name",
			policy:  IgnoredPolicy{},
			wantErr: true,
			errMsg:  "namespace is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
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

func TestScanResult_JSONSerialization(t *testing.T) {
	result := ScanResult{
		Success:    true,
		TargetFile: "plan.json",
		PolicySources: []PolicySource{
			{OriginalURL: "git::https://example.com/policies.git", PolicyCount: 5},
		},
		Violations: []PolicyViolation{
			{
				Policy:    "test_policy",
				Rule:      "test_rule",
				Message:   "Test violation",
				Namespace: "test",
				Severity:  "error",
				Resource:  "test_resource",
			},
		},
		Output: "conftest output",
		Summary: Summary{
			TotalViolations: 1,
			ErrorCount:      1,
			WarningCount:    0,
			InfoCount:       0,
			PoliciesRun:     5,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled ScanResult
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify key fields
	assert.Equal(t, result.Success, unmarshaled.Success)
	assert.Equal(t, result.TargetFile, unmarshaled.TargetFile)
	assert.Len(t, unmarshaled.PolicySources, len(result.PolicySources))
	assert.Len(t, unmarshaled.Violations, len(result.Violations))
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
