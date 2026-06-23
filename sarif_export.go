package finding

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

func sarifResultsFromFindings(findings []Finding, minSeverity Severity) []sarifResult {
	results := make([]sarifResult, 0, len(findings))

	for _, f := range findings {
		if f.IsSuppressed() || f.Severity.LessThan(minSeverity) {
			continue
		}

		results = append(results, findingToSARIF(f))
	}

	return results
}

func sarifDriverFromReport(r *Report) sarifDriver {
	return sarifDriver{Name: r.Tool.Name, Version: r.Tool.Version}
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
// The context is checked for cancellation before encoding begins.
func (r *Report) WriteSARIF(ctx context.Context, w io.Writer) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("writing SARIF: %w", err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	err = enc.Encode(r.sarifLog())
	if err != nil {
		return fmt.Errorf("encoding SARIF: %w", err)
	}

	return nil
}

// WriteSARIFFiltered writes non-suppressed findings with severity >= minSeverity
// in SARIF 2.1.0 format directly to w.
// Streams via json.Encoder, avoiding the intermediate []byte buffer.
// The context is checked for cancellation before encoding begins.
func (r *Report) WriteSARIFFiltered(ctx context.Context, w io.Writer, minSeverity Severity) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("writing SARIF filtered (minSeverity=%s): %w", minSeverity, err)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	err = enc.Encode(r.sarifLogFiltered(minSeverity))
	if err != nil {
		return fmt.Errorf("encoding SARIF filtered (minSeverity=%s): %w", minSeverity, err)
	}

	return nil
}

// WriteTo writes the report in SARIF 2.1.0 format to w and returns the bytes written.
// Implements io.WriterTo, enabling use with io.Copy for streaming SARIF output.
//
// For context-aware cancellation, prefer WriteSARIF directly.
func (r *Report) WriteTo(w io.Writer) (int64, error) {
	cw := &countingWriter{w: w}

	err := r.WriteSARIF(context.Background(), cw)
	if err != nil {
		return cw.n, fmt.Errorf("writing SARIF: %w", err)
	}

	return cw.n, nil
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)

	return n, err //nolint:wrapcheck // passthrough writer — wrapping would be misleading
}

func (r *Report) sarifLog() sarifLog {
	return r.buildsarifLog(sarifResultsFromFindings(r.readFindings(), SeverityInfo))
}

func (r *Report) sarifLogFiltered(severity Severity) sarifLog {
	return r.buildsarifLog(sarifResultsFromFindings(r.readFindings(), severity))
}

func (r *Report) buildsarifLog(results []sarifResult) sarifLog {
	return sarifLog{
		Version: sarifVersion,
		Schema:  sarifSchema,
		Runs: []sarifRun{
			{
				Tool:    sarifTool{Driver: sarifDriverFromReport(r)},
				Results: results,
			},
		},
	}
}

func findingToSARIF(f Finding) sarifResult {
	result := sarifResult{
		RuleID:     string(f.Rule),
		Level:      severityToSARIFLevel(f.Severity),
		Message:    sarifMessage{Text: f.Message},
		Locations:  sarifLocations(f),
		Fixes:      sarifFixes(f),
		Related:    sarifRelatedLocs(f),
		Properties: sarifProperties(f),
	}

	if f.Confidence > 0 {
		result.Rank = float64(f.NormalizedConfidence()) * sarifConfidenceScale
	}

	return result
}

func sarifLocations(f Finding) []sarifLocation {
	return []sarifLocation{{
		PhysicalLocation: sarifPhysicalLocation{
			ArtifactLocation: sarifArtifactLocation{URI: f.Position.File},
			Region:           findingRegion(f),
		},
	}}
}

func findingRegion(f Finding) *sarifRegion {
	region := &sarifRegion{
		StartLine:   f.Position.Line,
		StartColumn: f.Position.Column,
	}

	if f.Range != nil && f.Range.HasEnd() {
		region.EndLine = f.Range.End.Line
		region.EndColumn = f.Range.End.Column
	}

	if f.Snippet != "" {
		region.Snippet = f.Snippet
	}

	return region
}

func findingFixRegion(f Finding) sarifRegion {
	region := sarifRegion{
		StartLine:   f.Position.Line,
		StartColumn: f.Position.Column,
	}

	if f.Range != nil && f.Range.HasEnd() {
		region.EndLine = f.Range.End.Line
		region.EndColumn = f.Range.End.Column
	} else {
		region.EndLine = f.Position.Line
		region.EndColumn = f.Position.Column
	}

	return region
}

func sarifFixes(f Finding) []sarifFix {
	if f.HasFix() {
		region := findingFixRegion(f)

		return []sarifFix{{
			Description: sarifMessage{Text: f.Suggestion},
			Changes: []sarifArtifactChange{{
				ArtifactLocation: sarifArtifactLocation{URI: f.Position.File},
				Replacements: []sarifReplacement{{
					DeletedRegion: region,
					InsertedText:  sarifMessage{Text: f.AfterCode},
				}},
			}},
		}}
	}

	if f.HasSuggestion() {
		return []sarifFix{{
			Description: sarifMessage{Text: f.Suggestion},
		}}
	}

	return nil
}

func sarifRelatedLocs(f Finding) []sarifRelatedLoc {
	if len(f.Related) == 0 {
		return nil
	}

	related := make([]sarifRelatedLoc, 0, len(f.Related))

	for _, rel := range f.Related {
		region := &sarifRegion{
			StartLine:   rel.Position.Line,
			StartColumn: rel.Position.Column,
		}
		if rel.Range != nil && rel.Range.HasEnd() {
			region.EndLine = rel.Range.End.Line
			region.EndColumn = rel.Range.End.Column
		}

		sarifRel := sarifRelatedLoc{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: rel.Position.File},
				Region:           region,
			},
			Message: sarifMessage{Text: string(rel.Relation)},
		}

		if rel.FindingID != "" {
			sarifRel.Properties = map[string]any{sarifPropID: string(rel.FindingID)}
		}

		related = append(related, sarifRel)
	}

	return related
}

func sarifProperties(f Finding) map[string]any {
	props := map[string]any{
		sarifPropID:          string(f.ID),
		sarifPropSeverity:    string(f.Severity),
		sarifPropFixStrategy: string(f.FixStrategy),
		sarifPropToolName:    string(f.ToolName),
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

	if f.AfterCode != "" {
		props[sarifPropAfterCode] = f.AfterCode
	}

	for k, v := range f.Metadata {
		props[sarifPropMetaPrefix+k] = v
	}

	return props
}
