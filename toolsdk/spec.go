package toolsdk

import (
	"context"

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
	ModuleFanOut bool

	// Inputs are the file patterns this tool reads. Used to derive data-flow
	// edges and to gate the tool on file presence. Empty = always file-relevant.
	Inputs []string

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
	// "rust", or "" for language-agnostic.
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

// Repairer applies fixes to files. Tools implement this when they can
// auto-repair the issues they detect.
type Repairer interface {
	// Repair applies fixes. The working directory is available via
	// finding.WorkingDirFromContext(ctx). Returns a description of what was
	// done (BuildFlow does NOT trust self-reported fix counts — it re-runs
	// Detect and measures the finding delta).
	Repair(ctx context.Context) (RepairResult, error)
}

// RepairResult describes the outcome of a repair run. Only Description is
// meaningful to BuildFlow; fix counts are measured by re-detection, not
// self-reported (structural anti-lie design).
type RepairResult struct {
	Description string
}

// RepairerFunc adapts a function to the Repairer interface.
type RepairerFunc func(ctx context.Context) (RepairResult, error)

// Repair implements Repairer.
func (f RepairerFunc) Repair(ctx context.Context) (RepairResult, error) { return f(ctx) }
