package finding

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
// Includes all findings, including suppressed ones.
func (r *Report) PrettyJSON() (string, error) {
	r.mu.RLock()
	bytes, err := json.MarshalIndent(r, "", "  ")
	r.mu.RUnlock()

	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}

	return string(bytes), nil
}

// PrettyJSONFiltered returns a formatted JSON representation with only
// active (non-suppressed) findings. Unlike PrettyJSON, this excludes
// suppressed findings from the output.
func (r *Report) PrettyJSONFiltered() (string, error) {
	r.mu.RLock()

	filtered := &Report{ //nolint:exhaustruct
		Tool:     r.Tool,
		Findings: make([]Finding, 0, len(r.Findings)),
		Summary:  Summary{}, //nolint:exhaustruct
	}
	for _, f := range r.Findings {
		if !f.IsSuppressed() {
			filtered.Findings = append(filtered.Findings, f)
		}
	}

	r.mu.RUnlock()

	filtered.ComputeSummary()

	bytes, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling filtered JSON: %w", err)
	}

	return string(bytes), nil
}

// FromJSON parses a Finding from JSON and validates required fields.
func FromJSON(data []byte) (Finding, error) {
	var f Finding

	err := json.Unmarshal(data, &f)
	if err != nil {
		return Finding{}, fmt.Errorf("unmarshal finding: %w", err)
	}

	err = f.Validate()
	if err != nil {
		return Finding{}, fmt.Errorf("%w: %w", ErrInvalidFinding, err)
	}

	return f, nil
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

// WriteJSON writes compact JSON directly to w.
// Avoids the intermediate string allocation of LineJSON.
func (f Finding) WriteJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(f)
	if err != nil {
		return fmt.Errorf("encoding finding JSON: %w", err)
	}

	return nil
}

// WriteJSON writes pretty-printed JSON directly to w.
// Avoids the intermediate string allocation of PrettyJSON.
// Safe for concurrent use.
func (r *Report) WriteJSON(w io.Writer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	err := enc.Encode(r)
	if err != nil {
		return fmt.Errorf("encoding report JSON: %w", err)
	}

	return nil
}
