package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lonegunmanb/terraform-mcp-eva/pkg/tflint"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TFLintScanParam struct {
	Category         string   `json:"category,omitempty" jsonschema:"Category type for TFLint configuration. Supported values: 'reusable' (default), 'example'. Determines which predefined configuration to use."`
	TargetDirectory  string   `json:"target_directory,omitempty" jsonschema:"Target directory to scan. If not specified, current working directory will be used. Can be absolute or relative path."`
	CustomConfigFile string   `json:"custom_config_file,omitempty" jsonschema:"Path to custom TFLint configuration file. If specified, this will be used instead of the category-based configuration."`
	IgnoredRuleIDs   []string `json:"ignored_rule_ids,omitempty" jsonschema:"List of TFLint rule IDs to ignore during scanning. These rules will be disabled in the configuration."`
}

func TFLintScan(_ context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[TFLintScanParam]) (*mcp.CallToolResultFor[any], error) {
	// Convert the MCP parameters to TFLint scan parameters
	scanParams := tflint.ScanParam{
		Category:     params.Arguments.Category,
		TargetPath:   params.Arguments.TargetDirectory,
		ConfigFile:   params.Arguments.CustomConfigFile,
		IgnoredRules: params.Arguments.IgnoredRuleIDs,
	}

	// Execute the TFLint scan
	result, err := tflint.Scan(scanParams)
	if err != nil {
		return nil, fmt.Errorf("TFLint scan failed: %w", err)
	}

	// Convert the result to compact JSON for AI agent efficiency
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scan result: %w", err)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: string(jsonResult),
			},
		},
	}, nil
}
