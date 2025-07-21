package azapi

import (
	"github.com/zclconf/go-cty/cty"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAzAPIType_WithoutJsonPath(t *testing.T) {
	resourceType, err := getSwaggerResourceType("Microsoft.Resources/resourcegroups", "2024-07-01")
	require.NoError(t, err)
	require.True(t, resourceType.IsObjectType())
	require.True(t, resourceType.HasAttribute("location"))
	locationType := resourceType.AttributeType("location")
	require.Equal(t, cty.String, locationType)
}

func TestGetAzAPIType_WithPath(t *testing.T) {
	cases := []struct {
		desc         string
		resourceType string
		apiVersion   string
		path         string
		expectedType string
	}{
		{
			resourceType: "Microsoft.MachineLearningServices/workspaces",
			apiVersion:   "2025-06-01",
			path:         "body.properties.encryption.identity.userAssignedIdentity",
			expectedType: `String`,
		},
		{
			resourceType: "Microsoft.MachineLearningServices/workspaces",
			apiVersion:   "2025-06-01",
			path:         "body.properties.publicNetworkAccess",
			expectedType: `String`,
		},
		{
			desc:         "path in array",
			resourceType: "Microsoft.Compute/virtualMachines",
			apiVersion:   "2024-11-01",
			path:         "body.properties.osProfile.secrets.sourceVault.id",
			expectedType: `String`,
		},
		{
			desc:         "object",
			resourceType: "Microsoft.Compute/virtualMachines",
			apiVersion:   "2024-11-01",
			path:         "body.properties.osProfile.secrets.sourceVault",
			expectedType: `ObjectWithOptionalAttrs(map[string]Type{"id":String}, []string{"id"})`,
		},
	}
	for _, c := range cases {
		caseName := c.desc
		if c.desc == "" {
			caseName = strings.Join([]string{c.resourceType, c.apiVersion, c.path}, "-")
		}
		t.Run(caseName, func(t *testing.T) {
			schema, err := GetResourceSchema(c.resourceType, c.apiVersion, c.path)
			require.NoError(t, err)
			assert.Equal(t, c.expectedType, schema)
		})
	}
}

func TestGetAzAPIType_WithPathToAzAPIProperty(t *testing.T) {
	cases := []struct {
		desc         string
		resourceType string
		apiVersion   string
		path         string
		expectedType string
	}{
		{
			resourceType: "Microsoft.MachineLearningServices/workspaces",
			apiVersion:   "2025-06-01",
			path:         "retry.interval_seconds",
			expectedType: `Number`,
		},
	}
	for _, c := range cases {
		caseName := c.desc
		if c.desc == "" {
			caseName = strings.Join([]string{c.resourceType, c.apiVersion, c.path}, "-")
		}
		t.Run(caseName, func(t *testing.T) {
			schema, err := GetResourceSchema(c.resourceType, c.apiVersion, c.path)
			require.NoError(t, err)
			assert.Equal(t, c.expectedType, schema)
		})
	}
}
