Terraform MCP Eva
# Terraform MCP Eva

![](https://raw.githubusercontent.com/lonegunmanb/terraform-mcp-eva/refs/heads/main/eva.png)

Eva stands for "tErraform deVeloper Assistant", which is a Model Context Protocol (MCP) server designed to help Terraform developers by providing comprehensive access to Terraform provider source code, schema information, and Azure API documentation.

## Overview

This MCP server provides AI agents with powerful tools to query and analyze Terraform providers, particularly focusing on:
- **Terraform Provider Source Code Analysis**: Deep dive into how Terraform resources are implemented
- **Schema Documentation**: Access fine-grained schema information for Terraform resources
- **Azure API Integration**: Query Azure resource schemas and API versions
- **Golang Source Code Exploration**: Analyze the underlying Go code of Terraform providers

## Installation & Usage

### Using with VS Code

Add this configuration to your VS Code MCP settings(make sure that you have permission to mount the current working dir into the container):

```json
{
    "servers": {
        "terraform-mcp-eva": {
            "type": "stdio",
            "command": "docker",
            "args": [
                "run",
                "-i",
                "--rm",
                "-v",
                "${workspaceFolder}:/workspace",
                "-e",
                "TRANSPORT_MODE=stdio",
                "--pull=always",
                "ghcr.io/lonegunmanb/terraform-mcp-eva"
            ]
        }
    }
}
```

## Available Tools

### � Code Quality & Linting

#### `tflint_scan`
**Parameters** (all optional):
- `category`: Category type for TFLint configuration - "reusable" (default) or "example"
- `target_directory`: Target directory to scan (defaults to current working directory)
- `custom_config_file`: Path to custom TFLint configuration file
- `ignored_rule_ids`: Array of TFLint rule IDs to ignore during scanning

**Description**: Execute TFLint scanning on Terraform code with configurable parameters. This tool performs static analysis of Terraform code using TFLint with predefined configurations for different code types.

**Returns**: Detailed scan results including:
- List of issues found with severity levels (error, warning, info)
- File locations and line numbers for each issue
- Scan summary statistics
- Raw TFLint output for debugging

**Use Cases**:
- Validate Terraform code quality and best practices
- Identify potential issues in Terraform configurations
- Perform automated code review of Terraform modules
- Check compliance with Terraform coding standards
- Integrate with CI/CD pipelines for quality gates

### �🔍 Golang Source Code Analysis

#### `golang_source_code_server_get_supported_golang_namespaces`
**Parameters**: None  
**Description**: Get all indexed golang namespaces available for source code retrieval.  Agent an read golang source code under supported namespaces.
**Returns**: JSON array of namespace strings like `['github.com/hashicorp/terraform-provider-azurerm/internal']`  
**Use Cases**:
- Discover what golang projects/packages have been indexed
- Find available namespaces before querying specific code symbols
- Understand the scope of indexed golang codebases

#### `golang_source_code_server_get_supported_tags`
**Parameters**:
- `namespace` (required): The golang namespace to get tags for (e.g. 'github.com/hashicorp/terraform-provider-azurerm/internal')

**Description**: Get all supported tags/versions for a specific golang namespace.  
**Returns**: JSON array of version tags like `['v4.20.0', 'v4.21.0']`  
**Use Cases**:
- Discover available versions/tags for a specific golang namespace
- Find the latest or specific versions before analyzing code
- Understand version history for indexed golang projects

#### `query_golang_source_code`
**Parameters**:
- `namespace` (required): The golang namespace to query
- `symbol` (required): The symbol type - one of: `func`, `method`, `type`, `var`(global variables and constants)
- `name` (required): The name of the function, method, type or variable
- `receiver` (optional): The type of method receiver (only for methods)
- `tag` (optional): Tag version (defaults to latest if not specified)

**Description**: Read golang source code for given type, variable, constant, function or method definition.  
**Use Cases**:
- See function, method, type, or variable definitions while reading golang source code
- Understand how Terraform providers expand or flatten structs, maps schema to API
- Debug issues related to specific Terraform resources

### 🏗️ Terraform Provider Analysis

#### `terraform_source_code_query_get_supported_providers`
**Parameters**: None  
**Description**: Get all supported Terraform provider names available for source code query.  
**Returns**: JSON array of provider names like `['azurerm']`  
**Use Cases**:
- Discover what Terraform providers have been indexed
- Find available providers before querying specific functions or methods

#### `query_terraform_block_implementation_source_code`
**Parameters**:
- `block_type` (required): The terraform block type (e.g. 'resource', 'data', 'ephemeral')
- `terraform_type` (required): The terraform type (e.g. 'azurerm_resource_group')
- `entrypoint_name` (required): The function or method name you want to read
  - For 'resource': 'create', 'read', 'update', 'delete', 'schema', 'attribute'
  - For 'data': 'read', 'schema', 'attribute'
  - For 'ephemeral': 'open', 'close', 'renew', 'schema'
- `tag` (optional): Tag version (defaults to latest if not specified)

**Description**: Read Terraform provider source code for a given Terraform block.  
**Use Cases**:
- Read the source code of specific Terraform functions or methods
- Understand how a Terraform Provider calls APIs
- Debug issues related to specific Terraform resources

### 📋 Schema Documentation

#### `query_terraform_fine_grained_document`
**Parameters**:
- `category` (required): Terraform block type - one of: `resource`, `data`, `ephemeral`
- `type` (required): Terraform block type like 'azurerm_resource_group'
- `path` (optional): JSON path to query specific schema parts (e.g. 'default_node_pool.upgrade_settings')

**Description**: Query fine-grained Terraform resource schema information.  
**Returns**: JSON string representing the resource schema with attribute descriptions  
**Supported Providers**: `azurerm`, `azuread`, `aws`, `awscc`, `google`  
**Use Cases**:
- Get schema information about specific attributes or nested blocks
- Understand resource structure and attribute descriptions
- Validate Terraform configuration requirements

### ☁️ Azure API Integration

#### `list_azapi_api_versions`
**Parameters**:
- `resource_type` (required): Azure resource type (e.g. 'Microsoft.Compute/virtualMachines')

**Description**: Query Azure API versions by resource type.  
**Returns**: List of API versions for the specified resource type  
**Use Cases**:
- Discover available API versions for Azure resources
- Find the latest API version before querying schemas

#### `query_azapi_resource_schema`
**Parameters**:
- `resource_type` (required): Azure resource type (e.g. 'Microsoft.Compute/virtualMachines')
- `api_version` (required): Azure resource api-version (e.g. '2024-11-01')
- `path` (optional): JSON path to query specific schema parts

**Description**: Query fine-grained AzAPI resource schema information.  
**Returns**: Go type string representation of the resource schema  
**Use Cases**:
- Get precise type information for Azure resources
- Understand resource structure for Go code development
- Validate AzAPI resource configurations

#### `query_azapi_resource_document`
**Parameters**:
- `resource_type` (required): Azure resource type (e.g. 'Microsoft.Compute/virtualMachines')
- `api_version` (required): Azure resource api-version (e.g. '2024-11-01')
- `path` (optional): JSON path to query specific property descriptions

**Description**: Query fine-grained AzAPI resource descriptions and documentation.  
**Returns**: Property descriptions or JSON object with property documentation  
**Use Cases**:
- Learn whether properties are read-only, write-only, or required
- Understand possible values for properties
- Get detailed documentation for Azure resource properties

## Workflow Examples

### Analyzing a Terraform Resource Implementation

1. **Discover available providers**:
   ```
   Use: terraform_source_code_query_get_supported_providers
   ```

2. **Find the resource implementation**:
   ```
   Use: query_terraform_block_implementation_source_code
   Parameters: { "block_type": "resource", "terraform_type": "azurerm_resource_group", "entrypoint_name": "create" }
   ```

3. **Explore related Go functions**:
   ```
   Use: query_golang_source_code
   Parameters: { "namespace": "github.com/hashicorp/terraform-provider-azurerm/internal/services/resource", "symbol": "func", "name": "resourceGroupCreateFunc" }
   ```

### Getting Azure Resource Schema Information

1. **List available API versions**:
   ```
   Use: list_azapi_api_versions
   Parameters: { "resource_type": "Microsoft.Compute/virtualMachines" }
   ```

2. **Get resource schema**:
   ```
   Use: query_azapi_resource_schema
   Parameters: { "resource_type": "Microsoft.Compute/virtualMachines", "api_version": "2024-11-01" }
   ```

3. **Get detailed property documentation**:
   ```
   Use: query_azapi_resource_document
   Parameters: { "resource_type": "Microsoft.Compute/virtualMachines", "api_version": "2024-11-01", "path": "body.properties.osProfile" }
   ```

## Key Features

- **Comprehensive Provider Coverage**: Deep indexing of major Terraform providers (AzureRM, AWS, etc.)
- **Version-Aware**: Support for querying specific versions of providers and APIs
- **Fine-Grained Queries**: Ability to query specific paths within schemas and resources
- **Source Code Integration**: Direct access to the actual Go implementation code
- **Azure-Specific Tools**: Specialized tools for Azure resource management and AzAPI provider

This MCP server is ideal for Terraform developers, DevOps engineers, and AI assistants working with infrastructure as code, providing unprecedented visibility into how Terraform providers work under the hood.