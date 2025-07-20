Terraform MCP Eva
# Terraform MCP Eva

Eva stands for "tErraform deVeloper Assistant", which is a tool designed to help Terraform developers by providing schema query, policy validation, code formatting, and other features.

To try this mcp in vscode:

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
                "-e",
                "TRANSPORT_MODE=stdio",
                "--pull=always",
                "ghcr.io/lonegunmanb/terraform-mcp-eva"
            ]
        }
    }
}
```