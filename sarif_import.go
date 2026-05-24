package finding

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FindingsFromSARIF parses SARIF JSON and returns Findings.
// It extracts go-finding-specific properties for round-trip fidelity
// (severity, ID, tool name, etc.) and falls back to SARIF fields otherwise.
func FindingsFromSARIF(data []byte) ([]Finding, error) {
	var log SarifLog

	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("parsing SARIF: %w", err)
	}

	var findings []Finding

	for _, run := range log.Runs {
		toolName := run.Tool.Driver.Name

		for _, r := range run.Results {
			f := findingFromSarResult(r, toolName)
			findings = append(findings, f)
		}
	}

	return findings, nil
}

// findingFromSarResult converts a single SarifResult into a Finding.
func findingFromSarResult(r SarifResult, toolName string) Finding {
	f := Finding{
		Rule:     r.RuleID,
		Severity: FromSARIFLevel(r.Level),
		Message:  r.Message.Text,
		ToolName: toolName,
	}

	applySarifPosition(&f, r)

	if r.Rank > 0 {
		f.Confidence = Confidence(r.Rank / sarifConfidenceScale)
	}

	if len(r.Fixes) > 0 {
		f.Suggestion = r.Fixes[0].Description.Text

		if len(r.Fixes[0].Changes) > 0 && len(r.Fixes[0].Changes[0].Replacements) > 0 {
			f.AfterCode = r.Fixes[0].Changes[0].Replacements[0].InsertedText.Text
			f.FixStrategy = FixStrategySuggest
		}
	}

	for _, rel := range r.Related {
		pos := Position{File: rel.PhysicalLocation.ArtifactLocation.URI}
		if rel.PhysicalLocation.Region != nil {
			pos.Line = rel.PhysicalLocation.Region.StartLine
			pos.Column = rel.PhysicalLocation.Region.StartColumn
		}

		ref := RelatedRef{
			Relation: rel.Message.Text,
			Position: pos,
		}
		if rel.Properties != nil {
			if v, ok := rel.Properties[sarifPropID].(string); ok {
				ref.FindingID = v
			}
		}

		f.Related = append(f.Related, ref)
	}

	if r.Properties != nil {
		applySarifProperties(&f, r.Properties)
	}

	if f.ID == "" {
		f.ID = GenerateID(f.ToolName, f.Rule, f.Position)
	}

	return f
}

// applySarifPosition sets the Position and Range fields from SARIF locations.
func applySarifPosition(f *Finding, r SarifResult) {
	if len(r.Locations) == 0 {
		return
	}

	loc := r.Locations[0]
	region := loc.PhysicalLocation.Region

	fileURI := loc.PhysicalLocation.ArtifactLocation.URI
	if region == nil {
		f.Position = Position{File: fileURI}

		return
	}

	f.Position = Position{
		File:   fileURI,
		Line:   region.StartLine,
		Column: region.StartColumn,
	}

	if region.EndLine > 0 ||
		region.EndColumn > 0 {
		f.Range = &Range{
			Start: f.Position,
			End: Position{
				File:   fileURI,
				Line:   region.EndLine,
				Column: region.EndColumn,
			},
		}
	}
}

// applySarifProperties restores go-finding-specific properties for round-trip fidelity.
func applySarifProperties(f *Finding, props map[string]any) {
	if v, ok := stringProp(props, sarifPropID); ok {
		f.ID = v
	}

	if v, ok := stringProp(props, sarifPropSeverity); ok {
		if s := Severity(v); s.IsValid() {
			f.Severity = s
		}
	}

	if v, ok := stringProp(props, sarifPropFixStrategy); ok {
		if fs := FixStrategy(v); fs.IsValid() {
			f.FixStrategy = fs
		}
	}

	if v, ok := stringProp(props, sarifPropToolName); ok {
		f.ToolName = v
	}

	if v, ok := stringProp(props, sarifPropCategory); ok {
		f.Category = Category(v)
	}

	if v, ok := props[sarifPropTags].([]any); ok {
		f.Tags = make([]Tag, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				f.Tags = append(f.Tags, Tag(s))
			}
		}
	}

	if v, ok := props[sarifPropConfidence].(float64); ok {
		f.Confidence = Confidence(v)
	}

	if v, ok := stringProp(props, sarifPropSuggestion); ok {
		f.Suggestion = v
	}

	if v, ok := stringProp(props, sarifPropSnippet); ok {
		f.Snippet = v
	}

	if v, ok := stringProp(props, sarifPropBeforeCode); ok {
		f.BeforeCode = v
	}

	if v, ok := stringProp(props, sarifPropAfterCode); ok {
		f.AfterCode = v
	}

	f.Metadata = sarifMetadataFromProps(props)
	if len(f.Metadata) == 0 {
		f.Metadata = nil
	}
}

// stringProp extracts a string property from a SARIF property bag.
func stringProp(props map[string]any, key string) (string, bool) {
	v, ok := props[key].(string)
	return v, ok
}

// sarifMetadataFromProps extracts non-go-finding properties as metadata.
// go-finding/edit/* properties are preserved for FixEdit round-tripping.
func sarifMetadataFromProps(props map[string]any) map[string]string {
	meta := make(map[string]string)

	for k, v := range props {
		if strings.HasPrefix(k, sarifPropPrefix) && !strings.HasPrefix(k, sarifPropEditPrefix) {
			continue
		}

		meta[k] = fmt.Sprintf("%v", v)
	}

	return meta
}
