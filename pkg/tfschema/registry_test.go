package tfschema

import (
	"encoding/json"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerySchema_AzurermResourceGroup_EmptyPath(t *testing.T) {
	// Test querying azurerm_resource_group resource with empty path
	result, err := QuerySchema("resource", "azurerm_resource_group", "")

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
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "name")

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
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "default_node_pool")

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
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "default_node_pool.upgrade_settings")

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
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "default_node_pool.name")

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
	result, err := QuerySchema("resource", "azurerm_kubernetes_cluster", "identity")

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
	_, err := QuerySchema("invalid", "azurerm_resource_group", "")

	require.Error(t, err, "Should return error for invalid category")

	expectedError := "unknown schema category, must be one of 'resource', 'data_source', or 'ephemeral'"
	assert.Equal(t, expectedError, err.Error(), "Error message should match expected")
}

func TestQuerySchema_NonExistentResource(t *testing.T) {
	// Test with non-existent resource
	_, err := QuerySchema("resource", "non_existent_resource", "")

	require.Error(t, err, "Should return error for non-existent resource")
	assert.Contains(t, err.Error(), "not found", "Error message should contain 'not found'")
}

func TestQuerySchema_DataSource(t *testing.T) {
	// Test querying a data source
	result, err := QuerySchema("data_source", "azurerm_resource_group", "")

	require.NoError(t, err, "QuerySchema should not return an error for data source")
	require.NotEmpty(t, result, "QuerySchema should not return empty result for data source")

	// Verify the result is valid JSON
	var schema tfjson.Schema
	err = json.Unmarshal([]byte(result), &schema)
	require.NoError(t, err, "Data source result should be valid JSON")
}
