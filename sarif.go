package finding

import (
	"encoding/json"
)

// SARIF confidence constant.
const (
	ConfidenceScale = 100.0 // Confidence values are expressed as percentage (0-100)
)

// SARIF types for Report generation.
// These are simplified representations of SARIF 2.1.0.

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
	RuleID     string                 `json:"ruleId"`
	Level      string                 `json:"level"`
	Message    SarifMessage           `json:"message"`
	Location   SarifLocation          `json:"location"`
	Fixes      []SarifFix             `json:"fixes,omitempty"`
	Related    []SarifRelatedLoc      `json:"relatedLocations,omitempty"`
	Rank       float64                `json:"rank,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
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
	Message          SarifMessage          `json:"message,omitempty"`
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
	results := make([]SarifResult, 0)

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
	return json.MarshalIndent(r.sarifLog(), "", "  ")
}

// ToSARIFFiltered converts only non-suppressed findings.
func (r *Report) ToSARIFFiltered(severity Severity) ([]byte, error) {
	return json.MarshalIndent(r.sarifLogFiltered(severity), "", "  ")
}

func (r *Report) sarifLog() SarifLog {
	return SarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []SarifRun{
			{
				Tool:    SarifTool{Driver: sarifDriverFromReport(r)},
				Results: sarifResultsFromFindings(r.Findings),
			},
		},
	}
}

func (r *Report) sarifLogFiltered(severity Severity) SarifLog {
	return SarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []SarifRun{
			{
				Tool:    SarifTool{Driver: sarifDriverFromReport(r)},
				Results: sarifResultsFromFindingsFiltered(r.Findings, severity),
			},
		},
	}
}

func findingToSARIF(f Finding) SarifResult {
	result := SarifResult{
		RuleID:  f.Rule,
		Level:   severityToSARIFLevel(f.Severity),
		Message: SarifMessage{Text: f.Message},
		Location: SarifLocation{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: f.Position.File},
				Region: &SarifRegion{
					StartLine:   f.Position.Line,
					StartColumn: f.Position.Column,
				},
			},
		},
		Rank: f.Confidence * 100.0, // SARIF uses 0-100
	}

	// Add end position if available
	if f.Range != nil && f.Range.HasEnd() {
		result.Location.PhysicalLocation.Region.EndLine = f.Range.End.Line
		result.Location.PhysicalLocation.Region.EndColumn = f.Range.End.Column
	}

	// Add fix if available
	if f.FixStrategy != FixStrategyNone && f.AfterCode != "" {
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
	}

	// Add related locations
	for _, rel := range f.Related {
		result.Related = append(result.Related, SarifRelatedLoc{
			PhysicalLocation: SarifPhysicalLocation{
				ArtifactLocation: SarifArtifactLocation{URI: rel.Position.File},
				Region: &SarifRegion{
					StartLine:   rel.Position.Line,
					StartColumn: rel.Position.Column,
				},
			},
			Message: SarifMessage{Text: rel.Relation},
		})
	}

	// Add metadata as properties
	if len(f.Metadata) > 0 {
		result.Properties = make(map[string]interface{})
		for k, v := range f.Metadata {
			result.Properties[k] = v
		}

		result.Properties["toolName"] = f.ToolName
		result.Properties["category"] = f.Category
		result.Properties["tag"] = f.Tag
	}

	return result
}

func severityToSARIFLevel(s Severity) string {
	switch s {
	case SeverityInfo:
		return "note"
	case SeverityWarning:
		return "warning"
	case SeverityError, SeverityCritical:
		return "error"
	default:
		return "warning"
	}
}

// FromSARIFLevel converts SARIF level back to Severity.
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
