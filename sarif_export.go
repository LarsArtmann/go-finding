package finding

import (
	"encoding/json"
	"fmt"
	"io"
)

func sarifResultsFromFindings(findings []Finding) []SarifResult {
	results := make([]SarifResult, 0, len(findings))
	for _, f := range findings {
		if f.IsSuppressed() {
			continue
		}

		results = append(results, findingToSARIF(f))
	}

	return results
}

func sarifResultsFromFindingsFiltered(findings []Finding, minSeverity Severity) []SarifResult {
	results := make([]SarifResult, 0, len(findings))

	for _, f := range findings {
		if f.IsSuppressed() || f.Severity.LessThan(minSeverity) {
			continue
		}

		results = append(results, findingToSARIF(f))
	}

	return results
}

func sarifDriverFromReport(r *Report) SarifDriver {
	return SarifDriver{Name: r.Tool.Name, Version: r.Tool.Version}
}

// ToSARIF converts a Report to SARIF 2.1.0 format.
//
// Round-trip losses: SARIF export→import does not preserve:
//   - Suppression data (suppressed findings are excluded from export)
//
// All other fields are preserved via the "properties" bag or related
// location properties.
func (r *Report) ToSARIF() ([]byte, error) {
	data, err := json.MarshalIndent(r.sarifLog(), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling SARIF: %w", err)
	}

	return data, nil
}

// ToSARIFFiltered converts non-suppressed findings with severity >= minSeverity
// to SARIF 2.1.0 format. It filters by BOTH suppression status and severity.
func (r *Report) ToSARIFFiltered(minSeverity Severity) ([]byte, error) {
	data, err := json.MarshalIndent(r.sarifLogFiltered(minSeverity), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling SARIF: %w", err)
	}

	return data, nil
}

// WriteSARIF writes the report in SARIF 2.1.0 format directly to w.
// Streams via json.Encoder, avoiding the intermediate []byte buffer of ToSARIF.
func (r *Report) WriteSARIF(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(r.sarifLog()); err != nil {
		return fmt.Errorf("encoding SARIF: %w", err)
	}

	return nil
}

// WriteSARIFFiltered writes non-suppressed findings with severity >= minSeverity
// in SARIF 2.1.0 format directly to w.
// Streams via json.Encoder, avoiding the intermediate []byte buffer.
func (r *Report) WriteSARIFFiltered(w io.Writer, minSeverity Severity) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(r.sarifLogFiltered(minSeverity)); err != nil {
		return fmt.Errorf("encoding SARIF: %w", err)
	}

	return nil
}

func (r *Report) sarifLog() SarifLog {
	return r.buildSarifLog(sarifResultsFromFindings(r.Findings))
}

func (r *Report) sarifLogFiltered(severity Severity) SarifLog {
	return r.buildSarifLog(sarifResultsFromFindingsFiltered(r.Findings, severity))
}

func (r *Report) buildSarifLog(results []SarifResult) SarifLog {
	return SarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []SarifRun{
			{
				Tool:    SarifTool{Driver: sarifDriverFromReport(r)},
				Results: results,
			},
		},
	}
}

func findingToSARIF(f Finding) SarifResult {
	result := SarifResult{ //nolint:exhaustruct
		RuleID:  f.Rule,
		Level:   severityToSARIFLevel(f.Severity),
		Message: SarifMessage{Text: f.Message},
		Locations: []SarifLocation{{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: f.Position.File},
				Region: &SarifRegion{ //nolint:exhaustruct
					StartLine:   f.Position.Line,
					StartColumn: f.Position.Column,
				},
			},
		}},
		Rank: float64(f.NormalizedConfidence()) * sarifConfidenceScale,
	}

	if f.Range != nil && f.Range.HasEnd() {
		result.Locations[0].PhysicalLocation.Region.EndLine = f.Range.End.Line
		result.Locations[0].PhysicalLocation.Region.EndColumn = f.Range.End.Column
	}

	if f.HasFix() {
		fix := SarifFix{
			Description: SarifMessage{Text: f.Suggestion},
			Changes: []SarifArtifactChange{
				{
					ArtifactLocation: SarifArtifactLocation{URI: f.Position.File},
					Replacements: []SarifReplacement{
						{
							DeletedRegion: SarifRegion{
								StartLine:   f.Position.Line,
								StartColumn: f.Position.Column,
								EndLine:     f.Position.Line,
								EndColumn:   f.Position.Column,
							},
							InsertedText: SarifMessage{Text: f.AfterCode},
						},
					},
				},
			},
		}
		if f.Range != nil && f.Range.HasEnd() {
			fix.Changes[0].Replacements[0].DeletedRegion.EndLine = f.Range.End.Line
			fix.Changes[0].Replacements[0].DeletedRegion.EndColumn = f.Range.End.Column
		}

		result.Fixes = append(result.Fixes, fix)
	} else if f.HasSuggestion() {
		result.Fixes = append(
			result.Fixes,
			SarifFix{ //nolint:exhaustruct // suggestion-only fix has no changes
				Description: SarifMessage{Text: f.Suggestion},
			},
		)
	}

	for _, rel := range f.Related {
		//nolint:exhaustruct // Properties set conditionally below
		sarifRel := SarifRelatedLoc{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: rel.Position.File},
				Region: &SarifRegion{ //nolint:exhaustruct
					StartLine:   rel.Position.Line,
					StartColumn: rel.Position.Column,
				},
			},
			Message: SarifMessage{Text: rel.Relation},
		}
		if rel.FindingID != "" {
			if sarifRel.Properties == nil {
				sarifRel.Properties = make(map[string]any)
			}
			sarifRel.Properties[sarifPropID] = rel.FindingID
		}
		result.Related = append(result.Related, sarifRel)
	}

	props := make(map[string]any)
	props[sarifPropID] = f.ID
	props[sarifPropSeverity] = string(f.Severity)
	props[sarifPropFixStrategy] = string(f.FixStrategy)
	props[sarifPropToolName] = f.ToolName

	if f.Category != "" {
		props[sarifPropCategory] = string(f.Category)
	}

	if len(f.Tags) > 0 {
		props[sarifPropTags] = f.Tags
	}

	if f.Confidence > 0 {
		props[sarifPropConfidence] = f.Confidence
	}

	if f.Suggestion != "" {
		props[sarifPropSuggestion] = f.Suggestion
	}

	if f.Snippet != "" {
		props[sarifPropSnippet] = f.Snippet
	}

	if f.BeforeCode != "" {
		props[sarifPropBeforeCode] = f.BeforeCode
	}

	for k, v := range f.Metadata {
		props[k] = v
	}

	result.Properties = props

	return result
}
