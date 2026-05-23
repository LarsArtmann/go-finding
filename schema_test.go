package finding

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestFindingJSONSchema_IsValidJSON(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	data, err := os.ReadFile("docs/schemas/finding.schema.json")
	g.Expect(err).NotTo(HaveOccurred())

	var schema map[string]any
	g.Expect(json.Unmarshal(data, &schema)).NotTo(HaveOccurred())
	g.Expect(schema["$schema"]).NotTo(BeEmpty())
	g.Expect(schema["title"]).To(Equal("Finding"))
	g.Expect(schema["required"]).
		To(ContainElements("id", "rule", "toolName", "message", "severity", "position", "fixStrategy"))
}

func TestFindingJSONSchema_RoundTrip(t *testing.T) {
	g := NewWithT(t)
	t.Parallel()

	f := Finding{
		ID:          "tool:rule:file.go:10:5",
		Rule:        "SA1000",
		ToolName:    "tool",
		Message:     "test finding",
		Severity:    SeverityError,
		Position:    Position{File: "file.go", Line: 10, Column: 5},
		Category:    CategorySecurity,
		Tags:        []Tag{"security", "injection"},
		FixStrategy: FixStrategyDirect,
		Suggestion:  "fix it",
		BeforeCode:  "old()",
		AfterCode:   "new()",
		Range:       NewRangePtr("file.go", 10, 5, 10, 10),
		Snippet:     "old()",
		Confidence:  0.95,
		Related: []RelatedRef{
			{
				FindingID: "other:1",
				Relation:  "clone-of",
				Position:  Position{File: "other.go", Line: 5},
			},
		},
		Suppression: &Suppression{
			Kind:   SuppressionInSource,
			Rule:   "SA1000",
			Reason: "false positive",
		},
		Metadata: map[string]string{"custom": "value"},
	}

	data, err := json.Marshal(f)
	g.Expect(err).NotTo(HaveOccurred())

	var parsed map[string]any
	g.Expect(json.Unmarshal(data, &parsed)).NotTo(HaveOccurred())

	g.Expect(parsed).To(HaveKey("id"))
	g.Expect(parsed).To(HaveKey("rule"))
	g.Expect(parsed).To(HaveKey("toolName"))
	g.Expect(parsed).To(HaveKey("message"))
	g.Expect(parsed).To(HaveKey("severity"))
	g.Expect(parsed).To(HaveKey("position"))
	g.Expect(parsed).To(HaveKey("fixStrategy"))
	g.Expect(parsed).To(HaveKey("category"))
	g.Expect(parsed).To(HaveKey("tags"))
	g.Expect(parsed).To(HaveKey("range"))
	g.Expect(parsed).To(HaveKey("confidence"))
	g.Expect(parsed).To(HaveKey("related"))
	g.Expect(parsed).To(HaveKey("suppression"))
	g.Expect(parsed).To(HaveKey("metadata"))
}
