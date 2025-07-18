package azapi

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryAzapiSchemaDesc_EnumAsPossibleValues(t *testing.T) {
	m, err := GetResourceSchemaDescription("Microsoft.CognitiveServices/accounts", "2025-06-01", "")
	require.NoError(t, err)
	descriptions, ok := m.(map[string]any)
	require.True(t, ok)
	require.Contains(t, descriptions, "properties")
	properties := descriptions["properties"].(map[string]any)
	require.Contains(t, properties, "publicNetworkAccess")
	publicNetworkAccess, ok := properties["publicNetworkAccess"].(string)
	require.True(t, ok)
	assert.Contains(t, publicNetworkAccess, "Possible values: Enabled,Disabled")
}

func TestQueryAzapiSchemaDesc_WithPathToProperty(t *testing.T) {
	descriptions, err := GetResourceSchemaDescription("Microsoft.CognitiveServices/accounts", "2025-06-01", "properties.publicNetworkAccess")
	require.NoError(t, err)
	desc, ok := descriptions.(string)
	require.True(t, ok)
	assert.Contains(t, desc, "Whether or not public endpoint access is allowed for this account.")
}

func TestQueryAzapiSchemaDesc_WithPathToObject(t *testing.T) {
	descriptions, err := GetResourceSchemaDescription("Microsoft.CognitiveServices/accounts", "2025-06-01", "properties.encryption")
	require.NoError(t, err)
	desc, ok := descriptions.(map[string]any)
	require.True(t, ok)
	assert.Contains(t, desc, "keyVaultProperties")
	keyVaultProperties, ok := desc["keyVaultProperties"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, keyVaultProperties, "keyName")
	keyName, ok := keyVaultProperties["keyName"].(string)
	require.True(t, ok)
	assert.Equal(t, "Name of the Key from KeyVault", keyName)
}

func TestQueryAzapiSchemaDesc_Readonly(t *testing.T) {
	dateCreated, err := GetResourceSchemaDescription("Microsoft.CognitiveServices/accounts", "2025-06-01", "properties.dateCreated")
	require.NoError(t, err)
	desc, ok := dateCreated.(string)
	require.True(t, ok)
	assert.Contains(t, desc, "ReadOnly")
}
