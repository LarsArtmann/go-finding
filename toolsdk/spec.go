package toolsdk

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/larsartmann/go-finding"
)

// Spec is the declarative description of a tool that BuildFlow consumes.
// Every field is a plain value (string, []string) or a go-finding interface —
// no BuildFlow-internal types leak here.
type Spec struct {
	// Name is the tool's unique identifier (e.g. "branching-flow").
	// Becomes domain.ToolName inside BuildFlow.
	Name string

	// Description is a human-readable one-liner shown in --list output.
	Description string

	// Trigger declares when the tool is relevant (file patterns + language).
	Trigger Trigger

	// DependsOn lists tool names that must run before this one (DAG ordering).
	// These are plain strings mapped to domain.ToolName at conversion time.
	DependsOn []string

	// ModuleFanOut declares that the tool runs once per Go module in a
	// multi-module workspace rather than once at the repo root. BuildFlow maps
	// it to its DAGTopology.ModuleFanOut at conversion time; a spec that omits
	// the field silently loses per-module fan-out there. Field-for-field
	// parity with BuildFlow's domain/tool.DAGTopology — the same parity rule
	// as Trigger.NotRequires.
	//
	// Consumer contract (enforced by BuildFlow's conversion, not by specs):
	// when ModuleFanOut is set, the converter wraps Detect/Repair so the
	// input's WorkDir wins over the context's, the resolved directory is
	// written back onto the context, and per-module progress is emitted —
	// specs must NOT self-wrap with their own WorkDir/progress adapters, and
	// callbacks MUST analyze the working directory from the context
	// (finding.WorkingDirFromContext), never the process working directory.
	ModuleFanOut bool

	// SwitchCases declares conditional repair paths keyed on the diagnose
	// output's finding rule IDs (R6: gotcha-#194 parity). BuildFlow maps
	// them to DAGTopology.SwitchCases verbatim; a spec that omits the field
	// keeps the plain detect→repair topology. Field-for-field parity with
	// BuildFlow's domain/tool.SwitchCase.
	//
	// Mutual exclusion is the SPEC AUTHOR'S responsibility: go-workflow runs
	// EVERY case whose predicate matches, in parallel, so cases gating
	// mutating repairs on one diagnose gate must encode precedence via
	// ExcludedRuleIDs (at most one case runs per diagnosis) or lose writes
	// to racing siblings.
	SwitchCases []SwitchCase

	// Inputs are the file patterns this tool reads. Used to derive data-flow
	// edges and to gate the tool on file presence. Empty = always file-relevant.
	Inputs []string

	// Options declares the tunable knobs this tool reads from the context
	// via OptionsFromContext. Declarations power consumer-side validation
	// (ValidateOptions) and documentation; the VALUES are set by the
	// consumer per run via WithOptions. Kind checking only — range and
	// semantic validation stay with the tool implementation. Empty = the
	// tool takes no options (the common case).
	Options []Option

	// Detect is the canonical go-finding Detector. nil = this tool does not
	// detect (it may only repair, generate, or execute commands).
	// The working directory is read via finding.WorkingDirFromContext(ctx).
	Detect finding.Detector

	// Repair, if non-nil, applies fixes. The working directory is read via
	// finding.WorkingDirFromContext(ctx).
	Repair Repairer

	// HealthCheck, if non-nil, verifies external dependencies (e.g. a binary
	// is installed) before the pipeline runs. nil = no external dep (always OK).
	HealthCheck func(ctx context.Context) error
}

// Trigger declares when a tool is relevant to a project: file patterns that
// must be present and the target language.
type Trigger struct {
	// Files are glob patterns that activate this tool (e.g. "**/*.go").
	// Empty means "always file-relevant".
	Files []string

	// Language is the target language code: "go", "js", "ts", "python", "nix",
	// "rust", or "" for language-agnostic (see ValidLanguages). Register
	// rejects any other value at startup.
	Language string

	// Requires are prerequisite file patterns the tool needs to function.
	// At least one pattern must match (OR within group). Example: a Go module
	// tool might require "go.mod" or "go.work". Empty means no prerequisites.
	Requires []string

	// NotRequires are disqualifying file patterns: if ANY of them matches,
	// the tool does not run. This expresses ownership deference: when a repo
	// carries its own toolchain config (e.g. treefmt.toml), a standalone
	// formatter for the same file types would fight it on every run, so it
	// defers instead. Field-for-field parity with BuildFlow's
	// domain/tool.Trigger.NotRequires — a spec that omits the field silently
	// loses the deference behavior at conversion time.
	NotRequires []string
}

// validLanguages is the fixed Trigger.Language vocabulary Register accepts,
// in documented order. Never mutated; exported read-only via ValidLanguages.
var validLanguages = []string{"go", "js", "ts", "python", "nix", "rust"} //nolint:gochecknoglobals // fixed language vocabulary, cloned on export

// ValidLanguages returns every Trigger.Language code Register accepts, in
// documented order, as a fresh copy. The empty string (language-agnostic)
// is the zero value and valid without being listed.
func ValidLanguages() []string {
	return slices.Clone(validLanguages)
}

// ErrInvalidTrigger reports a malformed Trigger declaration: an unknown
// Language code or an empty pattern in Files/Requires/NotRequires. The
// wrapped detail names the offending field; match with errors.Is.
var ErrInvalidTrigger = errors.New("toolsdk: invalid trigger declaration")

// Validate checks the trigger declaration: Language must be "" (language-
// agnostic) or one of ValidLanguages, and none of Files, Requires, or
// NotRequires may contain an empty pattern (an empty glob matches nothing
// and is always a spec-authoring bug). Register panics on the first error.
func (t Trigger) Validate() error {
	if t.Language != "" && !slices.Contains(validLanguages, t.Language) {
		return fmt.Errorf("%w: Language %q (want \"\" or one of: %s)",
			ErrInvalidTrigger, t.Language, strings.Join(validLanguages, ", "))
	}

	for _, field := range []struct {
		name     string
		patterns []string
	}{
		{"Files", t.Files},
		{"Requires", t.Requires},
		{"NotRequires", t.NotRequires},
	} {
		if slices.Contains(field.patterns, "") {
			return fmt.Errorf("%w: %s contains an empty pattern (an empty glob matches nothing)",
				ErrInvalidTrigger, field.name)
		}
	}

	return nil
}

// SwitchCase declares one conditional repair path gated on the diagnose
// step's finding rule IDs. Field-for-field mirror of BuildFlow's
// domain/tool.SwitchCase: Suffix names the DAG node, FindingRuleIDs select
// the case (any match runs it), ExcludedRuleIDs suppress it (any match
// skips it) — encode case precedence there so at most one mutating case
// runs per diagnosis.
type SwitchCase struct {
	// Suffix is appended to the tool name to form the DAG node name
	// (e.g. "fix-hash-mismatch" → "nix-hash-fix:fix-hash-mismatch").
	Suffix string

	// FindingRuleIDs lists the finding rule IDs that trigger this case.
	// The case runs when ANY of these rule IDs appears in the diagnose output.
	FindingRuleIDs []string

	// ExcludedRuleIDs suppresses this case when ANY of these rule IDs
	// appears in the diagnose output.
	ExcludedRuleIDs []string
}

// Repairer applies fixes to files. Tools implement this when they can
// auto-repair the issues they detect.
type Repairer interface {
	// Repair applies fixes. The working directory is available via
	// finding.WorkingDirFromContext(ctx). Returns a description of what was
	// done (BuildFlow does NOT trust self-reported fix counts — it re-runs
	// Detect and measures the finding delta).
	Repair(ctx context.Context) (RepairResult, error)
}

// Fact is one tool-defined provenance pair carried in RepairResult.Facts.
type Fact struct {
	// Key names the provenance fact (e.g. "baseline-failed").
	Key string

	// Value carries the fact's value (e.g. "true").
	Value string
}

// RepairResult describes the outcome of a repair run. Only Description is
// meaningful to BuildFlow's fix measurement: fix counts are measured by
// re-detection, not self-reported (structural anti-lie design).
//
// ModifiedFiles and Facts are a structured provenance channel, additive and
// non-gating: they record what the run did (files written; tool-defined
// key/value pairs such as "baseline-failed" = "true") so hosts can surface
// context without string-matching Description. Facts describe the run, they
// never claim its effect — nothing downstream gates on a fact, and Detect
// re-measures the actual fix.
type RepairResult struct {
	// Description is a human-readable summary of what the repair did.
	Description string

	// ModifiedFiles lists files the repair wrote (or would write under
	// dry-run), typically as paths relative to the run's working directory
	// (finding.WorkingDirFromContext). Informational: no consumer gates on
	// the list. Nil or empty when the repair touched nothing.
	ModifiedFiles []string

	// Facts are tool-defined provenance key/value pairs. Nil or empty when
	// the tool has nothing to record.
	Facts []Fact
}

// RepairerFunc adapts a function to the Repairer interface.
type RepairerFunc func(ctx context.Context) (RepairResult, error)

// Repair implements Repairer.
func (f RepairerFunc) Repair(ctx context.Context) (RepairResult, error) { return f(ctx) }
