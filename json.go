package finding

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

// Sentinel errors for JSON validation.
var (
	ErrInvalidFinding = errors.New("invalid finding: missing required fields")
	ErrInvalidReport  = errors.New("invalid report: missing tool name")
)

// FilterInvalid returns true if the finding is invalid (has missing required fields).
func FilterInvalid(f Finding) bool {
	return !f.IsValid()
}

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
		return nil, ErrInvalidFinding
	}

	return &f, nil
}

// ReportFromJSON parses a Report from JSON and validates required fields.
// Invalid findings are silently dropped. Use the returned count to detect data loss.
func ReportFromJSON(data []byte) (*Report, int, error) {
	var r Report

	err := json.Unmarshal(data, &r)
	if err != nil {
		return nil, 0, fmt.Errorf("unmarshal report: %w", err)
	}

	if r.Tool.Name == "" {
		return nil, 0, ErrInvalidReport
	}

	before := len(r.Findings)
	r.Findings = slices.DeleteFunc(r.Findings, FilterInvalid)

	return &r, before - len(r.Findings), nil
}

// FindingsFromJSON parses a slice of Findings from JSON and validates each one.
// Invalid findings are silently dropped. Use the returned count to detect data loss.
func FindingsFromJSON(data []byte) ([]Finding, int, error) {
	var findings []Finding

	err := json.Unmarshal(data, &findings)
	if err != nil {
		return nil, 0, fmt.Errorf("unmarshal findings: %w", err)
	}

	before := len(findings)
	findings = slices.DeleteFunc(findings, FilterInvalid)

	return findings, before - len(findings), nil
}

// LineJSON returns compact JSON (single line).
func (f Finding) LineJSON() (string, error) {
	bytes, err := json.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("marshaling finding: %w", err)
	}

	return string(bytes), nil
}
