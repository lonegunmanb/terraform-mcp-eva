package gophon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLatestResourceCreateSourceCode(t *testing.T) {
	code, err := GetTerraformSourceCode("resource", "azurerm_resource_group", "create", "")
	require.NoError(t, err)
	assert.Contains(t, code, "func resourceResourceGroupCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error")
}

func TestGetTagVersionResourceCreateSourceCode(t *testing.T) {
	code, err := GetTerraformSourceCode("resource", "azurerm_resource_group", "create", "v4.25.0")
	require.NoError(t, err)
	assert.Contains(t, code, "func resourceResourceGroupCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error")
}

func TestGetTagVersionEphemeralOpenSourceCode(t *testing.T) {
	code, err := GetTerraformSourceCode("ephemeral", "azurerm_key_vault_secret", "open", "v4.25.0")
	require.NoError(t, err)
	assert.Contains(t, code, "func (e *KeyVaultSecretEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {")
}
