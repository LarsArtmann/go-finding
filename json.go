package finding

import (
	"encoding/json"
	"fmt"
)

// PrettyJSON returns a formatted JSON representation of the report.
func (r *Report) PrettyJSON() (string, error) {
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}

	return string(bytes), nil
}

// FromJSON parses a Finding from JSON and validates required fields.
func FromJSON(data []byte) (*Finding, error) {
	var f Finding

	err := json.Unmarshal(data, &f)
	if err != nil {
		return nil, fmt.Errorf("unmarshal finding: %w", err)
	}

	if !f.IsValid() {
		return nil, fmt.Errorf("invalid finding: missing required fields (id, rule, toolName, message, position, severity)")
	}

	return &f, nil
}

// ReportFromJSON parses a Report from JSON and validates required fields.
func ReportFromJSON(data []byte) (*Report, error) {
	var r Report

	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal report: %w", err)
	}

	if r.Tool.Name == "" {
		return nil, fmt.Errorf("invalid report: missing tool name")
	}

	return &r, nil
}

// FindingsFromJSON parses a slice of Findings from JSON.
func FindingsFromJSON(data []byte) ([]Finding, error) {
	var findings []Finding

	err := json.Unmarshal(data, &findings)
	if err != nil {
		return nil, fmt.Errorf("unmarshal findings: %w", err)
	}

	return findings, nil
}

// LineJSON returns compact JSON (single line).
func (f Finding) LineJSON() (string, error) {
	bytes, err := json.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("marshaling finding: %w", err)
	}

	return string(bytes), nil
}
