package finding

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SARIF types for Report generation.
// These are simplified representations of SARIF 2.1.0.

// SARIF confidence scale: Confidence is 0-1, SARIF rank is 0-100.
const sarifConfidenceScale = 100.0

const (
	sarifPropID          = "go-finding/id"
	sarifPropSeverity    = "go-finding/severity"
	sarifPropFixStrategy = "go-finding/fixStrategy"
	sarifPropToolName    = "go-finding/toolName"
	sarifPropCategory    = "go-finding/category"
	sarifPropTag         = "go-finding/tag"
	sarifPropConfidence  = "go-finding/confidence"
	sarifPropSuggestion  = "go-finding/suggestion"
	sarifPropSnippet     = "go-finding/snippet"
	sarifPropPrefix      = "go-finding/"
)

// SarifLog represents a SARIF log file containing run results.
type SarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SarifRun `json:"runs"`
}

// SarifRun represents a single analysis run in a SARIF log.
type SarifRun struct {
	Tool    SarifTool     `json:"tool"`
	Results []SarifResult `json:"results"`
}

// SarifTool defines the static analysis tool that generated the results.
type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

// SarifDriver represents the main driver tool with version information.
type SarifDriver struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// SarifResult represents a single finding in SARIF format.
type SarifResult struct {
	RuleID     string            `json:"ruleId"`
	Level      string            `json:"level"`
	Message    SarifMessage      `json:"message"`
	Locations  []SarifLocation   `json:"locations"`
	Fixes      []SarifFix        `json:"fixes,omitempty"`
	Related    []SarifRelatedLoc `json:"relatedLocations,omitempty"`
	Rank       float64           `json:"rank,omitempty"`
	Properties map[string]any    `json:"properties,omitempty"`
}

// SarifMessage represents a message in SARIF format.
type SarifMessage struct {
	Text string `json:"text"`
}

// SarifLocation represents a location in SARIF format.
type SarifLocation struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
}

// SarifPhysicalLocation represents physical details of a location.
type SarifPhysicalLocation struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
	Region           *SarifRegion          `json:"region,omitempty"`
}

// SarifArtifactLocation represents the artifact URI.
type SarifArtifactLocation struct {
	URI string `json:"uri"`
}

// SarifRegion represents a code region in a text document.
type SarifRegion struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// SarifFix represents a fix to be applied to the artifact.
type SarifFix struct {
	Description SarifMessage          `json:"description"`
	Changes     []SarifArtifactChange `json:"artifactChanges"`
}

// SarifArtifactChange represents a change to an artifact.
type SarifArtifactChange struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
	Replacements     []SarifReplacement    `json:"replacements"`
}

// SarifReplacement represents a replacement of text in an artifact.
type SarifReplacement struct {
	DeletedRegion SarifRegion  `json:"deletedRegion"`
	InsertedText  SarifMessage `json:"insertedText"`
}

// SarifRelatedLoc represents a related location in SARIF.
type SarifRelatedLoc struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
	Message          SarifMessage          `json:"message"`
}

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
		Rank: f.NormalizedConfidence() * sarifConfidenceScale, // SARIF uses 0-100
	}

	// Add end position if available
	if f.Range != nil && f.Range.HasEnd() {
		result.Locations[0].PhysicalLocation.Region.EndLine = f.Range.End.Line
		result.Locations[0].PhysicalLocation.Region.EndColumn = f.Range.End.Column
	}

	// Add fix or suggestion if available.
	// Export fix description even when only a suggestion exists (no code replacement).
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
								EndLine:     f.Position.Line,   // Default to single line
								EndColumn:   f.Position.Column, // Default to single position
							},
							InsertedText: SarifMessage{Text: f.AfterCode},
						},
					},
				},
			},
		}
		// Override with actual range if available
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

	// Add related locations
	for _, rel := range f.Related {
		result.Related = append(result.Related, SarifRelatedLoc{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: rel.Position.File},
				Region: &SarifRegion{ //nolint:exhaustruct
					StartLine:   rel.Position.Line,
					StartColumn: rel.Position.Column,
				},
			},
			Message: SarifMessage{Text: rel.Relation},
		})
	}

	// Preserve all non-standard fields in properties for round-trip fidelity.
	props := make(map[string]any)
	props[sarifPropID] = f.ID
	props[sarifPropSeverity] = string(f.Severity)
	props[sarifPropFixStrategy] = string(f.FixStrategy)
	props[sarifPropToolName] = f.ToolName

	if f.Category != "" {
		props[sarifPropCategory] = string(f.Category)
	}

	if f.Tag != "" {
		props[sarifPropTag] = f.Tag
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

	for k, v := range f.Metadata {
		props[k] = v
	}

	result.Properties = props

	return result
}

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
	f := Finding{ //nolint:exhaustruct
		Rule:     r.RuleID,
		Severity: FromSARIFLevel(r.Level),
		Message:  r.Message.Text,
		ToolName: toolName,
	}

	applySarifPosition(&f, r)

	if r.Rank > 0 {
		f.Confidence = r.Rank / sarifConfidenceScale
	}

	if len(r.Fixes) > 0 && len(r.Fixes[0].Changes) > 0 &&
		len(r.Fixes[0].Changes[0].Replacements) > 0 {
		f.Suggestion = r.Fixes[0].Description.Text
		f.AfterCode = r.Fixes[0].Changes[0].Replacements[0].InsertedText.Text
		f.FixStrategy = FixStrategySuggest
	}

	for _, rel := range r.Related {
		pos := Position{File: rel.PhysicalLocation.ArtifactLocation.URI} //nolint:exhaustruct
		if rel.PhysicalLocation.Region != nil {
			pos.Line = rel.PhysicalLocation.Region.StartLine
			pos.Column = rel.PhysicalLocation.Region.StartColumn
		}

		f.Related = append(f.Related, RelatedRef{ //nolint:exhaustruct
			Relation: rel.Message.Text,
			Position: pos,
		})
	}

	if r.Properties != nil {
		applySarifProperties(&f, r.Properties)
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
		f.Position = Position{File: fileURI} //nolint:exhaustruct

		return
	}

	f.Position = Position{ //nolint:exhaustruct
		File:   fileURI,
		Line:   region.StartLine,
		Column: region.StartColumn,
	}

	if region.EndLine > 0 ||
		region.EndColumn > 0 {
		f.Range = &Range{
			Start: f.Position,
			End: Position{ //nolint:exhaustruct
				File:   fileURI,
				Line:   region.EndLine,
				Column: region.EndColumn,
			},
		}
	}
}

// applySarifProperties restores go-finding-specific properties for round-trip fidelity.
func applySarifProperties(f *Finding, props map[string]any) {
	if v, ok := props[sarifPropID].(string); ok {
		f.ID = v
	}

	if v, ok := props[sarifPropSeverity].(string); ok {
		if s := Severity(v); s.IsValid() {
			f.Severity = s
		}
	}

	if v, ok := props[sarifPropFixStrategy].(string); ok {
		if fs := FixStrategy(v); fs.IsValid() {
			f.FixStrategy = fs
		}
	}

	if v, ok := props[sarifPropToolName].(string); ok {
		f.ToolName = v
	}

	if v, ok := props[sarifPropCategory].(string); ok {
		f.Category = Category(v)
	}

	if v, ok := props[sarifPropTag].(string); ok {
		f.Tag = v
	}

	if v, ok := props[sarifPropConfidence].(float64); ok {
		f.Confidence = v
	}

	if v, ok := props[sarifPropSuggestion].(string); ok {
		f.Suggestion = v
	}

	if v, ok := props[sarifPropSnippet].(string); ok {
		f.Snippet = v
	}

	f.Metadata = sarifMetadataFromProps(props)
	if len(f.Metadata) == 0 {
		f.Metadata = nil
	}
}

// sarifMetadataFromProps extracts non-go-finding properties as metadata.
func sarifMetadataFromProps(props map[string]any) map[string]string {
	meta := make(map[string]string)

	for k, v := range props {
		if strings.HasPrefix(k, sarifPropPrefix) {
			continue
		}

		meta[k] = fmt.Sprintf("%v", v)
	}

	return meta
}

// severityToSARIFLevel converts a Severity to a SARIF level string.
//
// Known limitation: SeverityCritical maps to "error" because SARIF 2.1.0 does not
// have a "critical" level. The original severity is preserved in the result's
// Properties["go-finding/severity"] for round-trip fidelity. Use FromSARIFLevel
// only when Properties are not available; otherwise prefer reading the property.
func severityToSARIFLevel(
	s Severity,
) string {
	switch s {
	case SeverityInfo:
		return "note"
	case SeverityWarning:
		return "warning" //nolint:goconst // SARIF "warning" != SeverityWarning
	case SeverityError, SeverityCritical:
		return "error"
	default:
		return "warning"
	}
}

// FromSARIFLevel converts a SARIF level back to Severity.
// Lossy: both SeverityCritical and SeverityError map to SARIF "error",
// so FromSARIFLevel("error") returns SeverityError. For full fidelity,
// read the "go-finding/severity" property from the result instead.
func FromSARIFLevel(level string) Severity {
	switch level {
	case "note":
		return SeverityInfo
	case "warning":
		return SeverityWarning
	case "error":
		return SeverityError
	default:
		return SeverityWarning
	}
}
