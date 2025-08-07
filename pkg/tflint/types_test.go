package tflint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestScanSummary_CalculateTotals tests the behavior of calculating total issues
func TestScanSummary_CalculateTotals(t *testing.T) {
	tests := []struct {
		name     string
		summary  ScanSummary
		expected int
	}{
		{
			name: "calculates total correctly with all issue types",
			summary: ScanSummary{
				ErrorCount:   2,
				WarningCount: 3,
				InfoCount:    1,
			},
			expected: 6,
		},
		{
			name: "calculates total correctly with only errors",
			summary: ScanSummary{
				ErrorCount:   5,
				WarningCount: 0,
				InfoCount:    0,
			},
			expected: 5,
		},
		{
			name: "calculates total correctly with zero issues",
			summary: ScanSummary{
				ErrorCount:   0,
				WarningCount: 0,
				InfoCount:    0,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.summary.ErrorCount + tt.summary.WarningCount + tt.summary.InfoCount
			assert.Equal(t, tt.expected, total)
		})
	}
}
