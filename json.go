package finding

import (
	"encoding/json"
	"fmt"
)

// JSON helpers for serialization.

// MarshalJSON serializes a Finding to JSON.
func (f Finding) MarshalJSON() ([]byte, error) {
	// Use the default serialization
	type FindingAlias Finding
	return json.Marshal((*FindingAlias)(&f))
}

// UnmarshalJSON deserializes JSON to a Finding.
func (f *Finding) UnmarshalJSON(data []byte) error {
	type FindingAlias Finding
	aux := (*FindingAlias)(f)
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

// MarshalJSON serializes a Report to JSON.
func (r Report) MarshalJSON() ([]byte, error) {
	type ReportAlias Report
	return json.Marshal((*ReportAlias)(&r))
}

// PrettyJSON returns a formatted JSON representation of the report.
func (r *Report) PrettyJSON() (string, error) {
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON parses a Finding from JSON.
func FromJSON(data []byte) (*Finding, error) {
	var f Finding
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("unmarshal finding: %w", err)
	}
	return &f, nil
}

// ReportFromJSON parses a Report from JSON.
func ReportFromJSON(data []byte) (*Report, error) {
	var r Report
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("unmarshal report: %w", err)
	}
	return &r, nil
}

// FindingsFromJSON parses a slice of Findings from JSON.
func FindingsFromJSON(data []byte) ([]Finding, error) {
	var findings []Finding
	if err := json.Unmarshal(data, &findings); err != nil {
		return nil, fmt.Errorf("unmarshal findings: %w", err)
	}
	return findings, nil
}

// LineJSON returns compact JSON (single line).
func (f Finding) LineJSON() (string, error) {
	bytes, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
