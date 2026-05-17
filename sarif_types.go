package finding

// SARIF types for Report generation.
// These are simplified representations of SARIF 2.1.0.

// SARIF confidence scale: Confidence is 0-1, SARIF rank is 0-100.
const sarifConfidenceScale = 100.0

const (
	sarifVersion = "2.1.0"
	sarifSchema  = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"
)

const (
	sarifPropID          = "go-finding/id"
	sarifPropSeverity    = "go-finding/severity"
	sarifPropFixStrategy = "go-finding/fixStrategy"
	sarifPropToolName    = "go-finding/toolName"
	sarifPropCategory    = "go-finding/category"

	sarifPropTags       = "go-finding/tags"
	sarifPropConfidence = "go-finding/confidence"
	sarifPropSuggestion = "go-finding/suggestion"
	sarifPropSnippet    = "go-finding/snippet"
	sarifPropBeforeCode = "go-finding/beforeCode"
	sarifPropPrefix     = "go-finding/"
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
	Properties       map[string]any        `json:"properties,omitempty"`
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
