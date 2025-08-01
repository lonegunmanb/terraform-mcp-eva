package gophon

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryTypeWithTag(t *testing.T) {
	code, err := GetGolangSourceCode("github.com/hashicorp/terraform-provider-azurerm/internal/clients", "type", "", "Client", "v4.25.0")
	require.NoError(t, err)
	assert.Contains(t, code, "type Client struct {")
}

func TestQueryMethodWithTag(t *testing.T) {
	code, err := GetGolangSourceCode("github.com/hashicorp/terraform-provider-azurerm/internal/services/containerapps", "method", "ContainerAppResource", "Create", "v4.25.0")
	require.NoError(t, err)
	assert.Contains(t, code, "func (r ContainerAppResource) Create() sdk.ResourceFunc {")
}
