package conftest

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/prashantv/gostub"
)

func TestResolvePolicyUrls(t *testing.T) {
	tests := []struct {
		name                string
		predefinedAlias     string
		customUrls          []string
		wantErr             bool
		errMsg              string
		expectedUrlCount    int
		expectedUrls        []string
	}{
		{
			name:                "should resolve APRL predefined policies",
			predefinedAlias:     "aprl",
			wantErr:             false,
			expectedUrlCount:    1,
			expectedUrls:        []string{"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2"},
		},
		{
			name:                "should resolve AVMSEC predefined policies",
			predefinedAlias:     "avmsec",
			wantErr:             false,
			expectedUrlCount:    1,
			expectedUrls:        []string{"git::https://github.com/Azure/policy-library-avm.git//policy/avmsec"},
		},
		{
			name:                "should resolve ALL predefined policies",
			predefinedAlias:     "all",
			wantErr:             false,
			expectedUrlCount:    2,
			expectedUrls:        []string{
				"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2",
				"git::https://github.com/Azure/policy-library-avm.git//policy/avmsec",
			},
		},
		{
			name:                "should resolve custom policy URLs",
			customUrls:          []string{"git::https://example.com/policies.git", "https://example.com/policy.zip"},
			wantErr:             false,
			expectedUrlCount:    2,
			expectedUrls:        []string{"git::https://example.com/policies.git", "https://example.com/policy.zip"},
		},
		{
			name:                "should default to ALL when both empty",
			wantErr:             false,
			expectedUrlCount:    2,
			expectedUrls:        []string{
				"git::https://github.com/Azure/policy-library-avm.git//policy/Azure-Proactive-Resiliency-Library-v2",
				"git::https://github.com/Azure/policy-library-avm.git//policy/avmsec",
			},
		},
		{
			name:                "should fail with mutually exclusive parameters",
			predefinedAlias:     "aprl",
			customUrls:          []string{"git::https://example.com/policies.git"},
			wantErr:             true,
			errMsg:              "predefined_policy_library_alias and custom_urls are mutually exclusive",
		},
		{
			name:                "should fail with invalid predefined alias",
			predefinedAlias:     "invalid",
			wantErr:             true,
			errMsg:              "invalid predefined_policy_library_alias: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup memory filesystem (minimal setup since URL resolution doesn't need files)
			memFs := afero.NewMemMapFs()
			stubs := gostub.Stub(&fs, memFs)
			defer stubs.Reset()

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
				if tt.expectedUrls != nil {
					assert.Equal(t, tt.expectedUrls, urls)
				}
			}
		})
	}
}
