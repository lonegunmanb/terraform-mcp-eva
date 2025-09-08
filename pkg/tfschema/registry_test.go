package tfschema

import (
	"encoding/json"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Common provider request for testing - using azurerm as an example
var testProviderReq = ProviderRequest{
	ProviderNamespace: "hashicorp",
	ProviderName:      "azurerm",
	ProviderVersion:   "4.39.0",
}

func TestQuerySchema_AzurermResourceGroup_EmptyPath(t *testing.T) {
	// Test querying azurerm_resource_group resource with empty path
	result, err := QuerySchema("resource", "azurerm_resource_group", "", testProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error")
	require.NotEmpty(t, result, "QuerySchema should not return empty result")

	// Verify the result is valid JSON
	var schema tfjson.Schema
	err = json.Unmarshal([]byte(result), &schema)
	require.NoError(t, err, "Result should be valid JSON")

	// Verify it's a proper schema structure
	require.NotNil(t, schema.Block, "Schema block should not be nil")

	// Check for expected attributes in azurerm_resource_group
	expectedAttributes := []string{"name", "location"}
	for _, attr := range expectedAttributes {
		assert.Contains(t, schema.Block.Attributes, attr, "Expected attribute %s should be found in schema", attr)
	}
}

// Test cases for path parameter using azurerm_kubernetes_cluster
func TestQuerySchema_AzurermKubernetesCluster_RootLevelAttribute(t *testing.T) {
	// Test querying a root-level attribute
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "name", testProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error for root-level attribute")
	require.NotEmpty(t, result, "QuerySchema should not return empty result")

	// Verify the result is valid JSON representing an attribute
	var attr tfjson.SchemaAttribute
	err = json.Unmarshal([]byte(result), &attr)
	require.NoError(t, err, "Result should be valid JSON for attribute")

	// The name attribute should be required
	assert.True(t, attr.Required, "Name attribute should be required")
}

func TestQuerySchema_AzurermKubernetesCluster_NestedBlock(t *testing.T) {
	// Test querying a nested block (default_node_pool)
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "default_node_pool", testProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error for nested block")
	require.NotEmpty(t, result, "QuerySchema should not return empty result")

	// Verify the result is valid JSON representing a nested block
	var nestedBlock tfjson.SchemaBlockType
	err = json.Unmarshal([]byte(result), &nestedBlock)
	require.NoError(t, err, "Result should be valid JSON for nested block")

	// default_node_pool should have a block structure
	require.NotNil(t, nestedBlock.Block, "Nested block should have a block structure")

	// Check for expected attributes in default_node_pool
	expectedAttributes := []string{"name", "node_count", "vm_size"}
	for _, attr := range expectedAttributes {
		assert.Contains(t, nestedBlock.Block.Attributes, attr, "Expected attribute %s should be found in default_node_pool", attr)
	}
}

func TestQuerySchema_AzurermKubernetesCluster_DeepNestedPath(t *testing.T) {
	// Test querying a deep nested path (default_node_pool.upgrade_settings)
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "default_node_pool.upgrade_settings", testProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error for deep nested path")
	require.NotEmpty(t, result, "QuerySchema should not return empty result")

	// Verify the result is valid JSON
	var nestedBlock tfjson.SchemaBlockType
	err = json.Unmarshal([]byte(result), &nestedBlock)
	require.NoError(t, err, "Result should be valid JSON for deep nested block")

	require.NotNil(t, nestedBlock.Block, "Deep nested block should have a block structure")

	// upgrade_settings should have specific attributes
	expectedAttributes := []string{"max_surge"}
	for _, attr := range expectedAttributes {
		assert.Contains(t, nestedBlock.Block.Attributes, attr, "Expected attribute %s should be found in upgrade_settings", attr)
	}
}

func TestQuerySchema_AzurermKubernetesCluster_AttributeInNestedBlock(t *testing.T) {
	// Test querying a specific attribute within a nested block
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "default_node_pool.name", testProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error for attribute in nested block")
	require.NotEmpty(t, result, "QuerySchema should not return empty result")

	// Verify the result is valid JSON representing an attribute
	var attr tfjson.SchemaAttribute
	err = json.Unmarshal([]byte(result), &attr)
	require.NoError(t, err, "Result should be valid JSON for nested attribute")

	// The name attribute in default_node_pool should be required
	assert.True(t, attr.Required, "default_node_pool.name attribute should be required")
}

func TestQuerySchema_AzurermKubernetesCluster_ComplexNestedBlock(t *testing.T) {
	// Test querying the identity block which is commonly used
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "identity", testProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error for identity block")
	require.NotEmpty(t, result, "QuerySchema should not return empty result")

	// Verify the result is valid JSON representing a nested block
	var nestedBlock tfjson.SchemaBlockType
	err = json.Unmarshal([]byte(result), &nestedBlock)
	require.NoError(t, err, "Result should be valid JSON for identity block")

	require.NotNil(t, nestedBlock.Block, "Identity block should have a block structure")

	// identity should have type attribute
	assert.Contains(t, nestedBlock.Block.Attributes, "type", "Identity block should have 'type' attribute")
}

func TestQuerySchema_InvalidCategory(t *testing.T) {
	// Test with invalid category
	_, err := QuerySchema("invalid", "azurerm_resource_group", "", testProviderReq)

	require.Error(t, err, "Should return error for invalid category")

	expectedError := "unknown schema category, must be one of 'resource', 'data', 'ephemeral', or 'function'"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected")
}

func TestQuerySchema_NonExistentResource(t *testing.T) {
	// Test with non-existent resource
	_, err := QuerySchema("resource", "non_existent_resource", "", testProviderReq)

	require.Error(t, err, "Should return error for non-existent resource")
	assert.Contains(t, err.Error(), "failed to get resource schema", "Error message should contain appropriate error")
}

func TestQuerySchema_DataSource(t *testing.T) {
	// Test querying a data source
	result, err := QuerySchema("data", "azurerm_resource_group", "", testProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error for data source")
	require.NotEmpty(t, result, "QuerySchema should not return empty result for data source")

	// Verify the result is valid JSON
	var schema tfjson.Schema
	err = json.Unmarshal([]byte(result), &schema)
	require.NoError(t, err, "Data source result should be valid JSON")
}

func TestQuerySchema_FunctionSupport(t *testing.T) {
	// Test querying a function schema from Azure azapi provider
	azapiProviderReq := ProviderRequest{
		ProviderNamespace: "Azure",
		ProviderName:      "azapi",
		ProviderVersion:   "2.6.1",
	}

	result, err := QuerySchema("function", "build_resource_id", "", azapiProviderReq)

	require.NoError(t, err, "QuerySchema should not return an error for function")
	require.NotEmpty(t, result, "QuerySchema should not return empty result")

	// Verify the result is valid JSON representing a function signature
	var functionSig tfjson.FunctionSignature
	err = json.Unmarshal([]byte(result), &functionSig)
	require.NoError(t, err, "Result should be valid JSON for function signature")

	// Verify the function signature structure
	assert.Equal(t, "string", functionSig.ReturnType.FriendlyName(), "Function should return string")
	assert.Equal(t, 3, len(functionSig.Parameters), "Function should have 3 parameters")

	// Verify parameter names and types
	expectedParams := []string{"parent_id", "resource_type", "name"}
	for i, expectedParam := range expectedParams {
		assert.Equal(t, expectedParam, functionSig.Parameters[i].Name, "Parameter %d should be named %s", i, expectedParam)
		assert.Equal(t, "string", functionSig.Parameters[i].Type.FriendlyName(), "Parameter %s should be string type", expectedParam)
	}
}

func TestQuerySchema_Function_PathNotSupported(t *testing.T) {
	// Test that path queries are not supported for functions
	azapiProviderReq := ProviderRequest{
		ProviderNamespace: "Azure",
		ProviderName:      "azapi",
		ProviderVersion:   "2.6.1",
	}

	_, err := QuerySchema("function", "build_resource_id", "some.path", azapiProviderReq)

	require.Error(t, err, "Should return error for function with path")
	assert.Equal(t, "path queries are not supported for function schemas", err.Error(), "Error message should match expected")
}

// Tests for ListItems function
func TestListItems_Resources(t *testing.T) {
	// Test listing resources for azurerm provider
	items, err := ListItems("resource", testProviderReq)

	require.NoError(t, err, "ListItems should not return an error")
	require.NotEmpty(t, items, "ListItems should return at least one resource")

	// Check that common azurerm resources are in the list
	expectedResources := []string{"azurerm_resource_group", "azurerm_virtual_network", "azurerm_storage_account"}
	for _, expected := range expectedResources {
		assert.Contains(t, items, expected, "Expected resource %s should be in the list", expected)
	}
}

func TestListItems_DataSources(t *testing.T) {
	// Test listing data sources for azurerm provider
	items, err := ListItems("data", testProviderReq)

	require.NoError(t, err, "ListItems should not return an error")
	require.NotEmpty(t, items, "ListItems should return at least one data source")

	// Check that common azurerm data sources are in the list
	expectedDataSources := []string{"azurerm_resource_group", "azurerm_client_config", "azurerm_subscription"}
	for _, expected := range expectedDataSources {
		assert.Contains(t, items, expected, "Expected data source %s should be in the list", expected)
	}
}

func TestListItems_Functions(t *testing.T) {
	// Test listing functions for azapi provider (known to have functions)
	azapiProviderReq := ProviderRequest{
		ProviderNamespace: "Azure",
		ProviderName:      "azapi",
		ProviderVersion:   "2.6.1",
	}

	items, err := ListItems("function", azapiProviderReq)

	require.NoError(t, err, "ListItems should not return an error")
	require.NotEmpty(t, items, "ListItems should return at least one function")

	// Check that known azapi functions are in the list
	expectedFunctions := []string{"build_resource_id"}
	for _, expected := range expectedFunctions {
		assert.Contains(t, items, expected, "Expected function %s should be in the list", expected)
	}
}

func TestListItems_Ephemeral(t *testing.T) {
	// Test listing ephemeral resources for azurerm provider
	items, err := ListItems("ephemeral", testProviderReq)

	require.NoError(t, err, "ListItems should not return an error")
	// Note: ephemeral resources might be empty for some providers, so we don't require items
	// but we should verify the result is a valid slice
	assert.IsType(t, []string{}, items, "ListItems should return a slice of strings")
}

func TestListItems_InvalidCategory(t *testing.T) {
	// Test invalid category
	_, err := ListItems("invalid_category", testProviderReq)

	require.Error(t, err, "ListItems should return error for invalid category")
	assert.Equal(t, "unknown category, must be one of 'resource', 'data', 'ephemeral', or 'function'", err.Error(), "Error message should match expected")
}

func TestListItems_NonExistentProvider(t *testing.T) {
	// Test listing for non-existent provider
	_, err := ListItems("resource", ProviderRequest{
		ProviderNamespace: "nonexistent",
		ProviderName:      "invalid-provider",
		ProviderVersion:   "1.0.0",
	})

	require.Error(t, err, "ListItems should return error for non-existent provider")
}
