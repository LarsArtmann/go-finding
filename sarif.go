package finding

import (
	"encoding/json"
)

// SARIF types for Report generation.
// These are simplified representations of SARIF 2.1.0.

type SarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SarifRun `json:"runs"`
}

type SarifRun struct {
	Tool    SarifTool     `json:"tool"`
	Results []SarifResult `json:"results"`
}

type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

type SarifDriver struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

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

type SarifMessage struct {
	Text string `json:"text"`
}

type SarifLocation struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
}

type SarifPhysicalLocation struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
	Region           *SarifRegion          `json:"region,omitempty"`
}

type SarifArtifactLocation struct {
	URI string `json:"uri"`
}

type SarifRegion struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

type SarifFix struct {
	Description SarifMessage          `json:"description"`
	Changes     []SarifArtifactChange `json:"artifactChanges"`
}

type SarifArtifactChange struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
	Replacements     []SarifReplacement    `json:"replacements"`
}

type SarifReplacement struct {
	DeletedRegion SarifRegion  `json:"deletedRegion"`
	InsertedText  SarifMessage `json:"insertedText"`
}

type SarifRelatedLoc struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
	Message          SarifMessage          `json:"message,omitempty"`
}

// ToSARIF converts a Report to SARIF 2.1.0 format.
func (r *Report) ToSARIF() ([]byte, error) {
	log := SarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []SarifRun{
			{
				Tool: SarifTool{
					Driver: SarifDriver{
						Name:    r.Tool.Name,
						Version: r.Tool.Version,
					},
				},
				Results: make([]SarifResult, 0, len(r.Findings)),
			},
		},
	}

	for _, f := range r.Findings {
		if f.IsSuppressed() {
			continue // Skip suppressed findings
		}

		result := findingToSARIF(f)
		log.Runs[0].Results = append(log.Runs[0].Results, result)
	}

	return json.MarshalIndent(log, "", "  ")
}

// ToSARIFFiltered converts only non-suppressed findings.
func (r *Report) ToSARIFFiltered(severity Severity) ([]byte, error) {
	log := SarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []SarifRun{
			{
				Tool: SarifTool{
					Driver: SarifDriver{
						Name:    r.Tool.Name,
						Version: r.Tool.Version,
					},
				},
				Results: make([]SarifResult, 0),
			},
		},
	}

	for _, f := range r.Findings {
		if f.IsSuppressed() || f.Severity.LessThan(severity) {
			continue
		}

		result := findingToSARIF(f)
		log.Runs[0].Results = append(log.Runs[0].Results, result)
	}

	return json.MarshalIndent(log, "", "  ")
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
	if f.Range != nil && f.Range.End.Line > 0 {
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
		if f.Range != nil && f.Range.End.Line > 0 {
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
