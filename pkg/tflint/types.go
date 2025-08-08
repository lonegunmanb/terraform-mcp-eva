package tflint

// ScanParam represents the input parameters for TFLint scanning
type ScanParam struct {
	Category        string   `json:"category,omitempty" jsonschema:"enum=reusable,example;description=Type of Terraform code to scan: 'reusable' for reusable modules, 'example' for example code. Defaults to 'reusable'"`
	RemoteConfigUrl string   `json:"remote_config_url,omitempty" jsonschema:"description=Optional remote TFLint configuration URL (HTTP(S) or git go-getter syntax) mutually exclusive with 'category'; must point to a single file"`
	TargetPath      string   `json:"target_path,omitempty" jsonschema:"description=Path to the directory containing Terraform code to scan. Defaults to current directory"`
	ConfigFile      string   `json:"config_file,omitempty" jsonschema:"description=Optional path to custom TFLint configuration file"`
	IgnoredRules    []string `json:"ignored_rules,omitempty" jsonschema:"description=Optional list of TFLint rule IDs to ignore during scanning"`
}

// ScanResult represents the result of a TFLint scan
type ScanResult struct {
	Success    bool        `json:"success"`
	Category   string      `json:"category"`
	TargetPath string      `json:"target_path"`
	Issues     []Issue     `json:"issues,omitempty"`
	Output     string      `json:"output"`
	Summary    ScanSummary `json:"summary"`
}

// Issue represents a single issue found by TFLint
type Issue struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Range    Range  `json:"range"`
}

// Range represents the location of an issue in the source code
type Range struct {
	Filename string `json:"filename"`
	Start    Point  `json:"start"`
	End      Point  `json:"end"`
}

// Point represents a specific position in source code
type Point struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// ScanSummary provides a summary of scan results
type ScanSummary struct {
	TotalIssues  int `json:"total_issues"`
	ErrorCount   int `json:"error_count"`
	WarningCount int `json:"warning_count"`
	InfoCount    int `json:"info_count"`
}

// ConfigData represents TFLint configuration data
type ConfigData struct {
	TempDir    string `json:"temp_dir"`
	ConfigPath string `json:"config_path"`
}
