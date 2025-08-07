package tflint

import (
	"os"
	"path/filepath"
)

// getDefaultCategory returns the default category if the provided category is empty or invalid
func getDefaultCategory(category string) string {
	if category == "reusable" || category == "example" {
		return category
	}
	return "reusable"
}

// getConfigURL returns the appropriate configuration URL based on category
func getConfigURL(category string) string {
	switch category {
	case "example":
		return "https://raw.githubusercontent.com/Azure/avm-terraform-governance/refs/heads/main/tflint-configs/avm.tflint_example.hcl"
	default:
		return "https://raw.githubusercontent.com/Azure/avm-terraform-governance/refs/heads/main/tflint-configs/avm.tflint.hcl"
	}
}

// getDefaultTargetPath returns the current working directory if targetPath is empty
var getDefaultTargetPath = func(targetPath string) (string, error) {
	if targetPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return cwd, nil
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

// validateCategory checks if the provided category is valid
func validateCategory(category string) bool {
	return category == "reusable" || category == "example"
}
