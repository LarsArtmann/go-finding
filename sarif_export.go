package finding

import (
	"encoding/json"
	"fmt"
	"io"
)

func sarifResultsFromFindings(findings []Finding, minSeverity Severity) []SarifResult {
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
		return nil, fmt.Errorf("marshaling SARIF filtered (minSeverity=%s): %w", minSeverity, err)
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
		return fmt.Errorf("encoding SARIF filtered (minSeverity=%s): %w", minSeverity, err)
	}

	return nil
}

func (r *Report) sarifLog() SarifLog {
	return r.buildSarifLog(sarifResultsFromFindings(r.Findings, SeverityInfo))
}

func (r *Report) sarifLogFiltered(severity Severity) SarifLog {
	return r.buildSarifLog(sarifResultsFromFindings(r.Findings, severity))
}

func (r *Report) buildSarifLog(results []SarifResult) SarifLog {
	return SarifLog{
		Version: sarifVersion,
		Schema:  sarifSchema,
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
		RuleID:     f.Rule,
		Level:      severityToSARIFLevel(f.Severity),
		Message:    SarifMessage{Text: f.Message},
		Locations:  sarifLocations(f),
		Rank:       float64(f.NormalizedConfidence()) * sarifConfidenceScale,
		Fixes:      sarifFixes(f),
		Related:    sarifRelatedLocs(f),
		Properties: sarifProperties(f),
	}

	return result
}

func sarifLocations(f Finding) []SarifLocation {
	region := &SarifRegion{ //nolint:exhaustruct
		StartLine:   f.Position.Line,
		StartColumn: f.Position.Column,
	}

	if f.Range != nil && f.Range.HasEnd() {
		region.EndLine = f.Range.End.Line
		region.EndColumn = f.Range.End.Column
	}

	return []SarifLocation{{
		PhysicalLocation: SarifPhysicalLocation{
			ArtifactLocation: SarifArtifactLocation{URI: f.Position.File},
			Region:           region,
		},
	}}
}

func sarifFixes(f Finding) []SarifFix {
	if f.HasFix() {
		region := SarifRegion{ //nolint:exhaustruct
			StartLine:   f.Position.Line,
			StartColumn: f.Position.Column,
			EndLine:     f.Position.Line,
			EndColumn:   f.Position.Column,
		}
		if f.Range != nil && f.Range.HasEnd() {
			region.EndLine = f.Range.End.Line
			region.EndColumn = f.Range.End.Column
		}

		return []SarifFix{{
			Description: SarifMessage{Text: f.Suggestion},
			Changes: []SarifArtifactChange{{
				ArtifactLocation: SarifArtifactLocation{URI: f.Position.File},
				Replacements: []SarifReplacement{{
					DeletedRegion: region,
					InsertedText:  SarifMessage{Text: f.AfterCode},
				}},
			}},
		}}
	}

	if f.HasSuggestion() {
		return []SarifFix{{ //nolint:exhaustruct
			Description: SarifMessage{Text: f.Suggestion},
		}}
	}

	return nil
}

func sarifRelatedLocs(f Finding) []SarifRelatedLoc {
	if len(f.Related) == 0 {
		return nil
	}

	related := make([]SarifRelatedLoc, 0, len(f.Related))

	for _, rel := range f.Related {
		sarifRel := SarifRelatedLoc{ //nolint:exhaustruct
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
			sarifRel.Properties = map[string]any{sarifPropID: rel.FindingID}
		}

		related = append(related, sarifRel)
	}

	return related
}

func sarifProperties(f Finding) map[string]any {
	props := map[string]any{
		sarifPropID:          f.ID,
		sarifPropSeverity:    string(f.Severity),
		sarifPropFixStrategy: string(f.FixStrategy),
		sarifPropToolName:    f.ToolName,
	}

	if f.Category != "" {
		props[sarifPropCategory] = string(f.Category)
	}

	if len(f.Tags) > 0 {
		props[sarifPropTags] = f.Tags
	}

	if f.Confidence > 0 {
		props[sarifPropConfidence] = float64(f.Confidence)
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

	return props
}
