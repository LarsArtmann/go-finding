// Package toolsdk is the stable contract that external tools implement so
// BuildFlow can consume them as DAG providers WITHOUT each tool importing
// BuildFlow internals or writing per-tool glue.
//
// A tool declares a single Spec (typically in a package-level var) describing
// its name, trigger, file inputs, and detection/repair capabilities expressed
// against the canonical go-finding types. The tool then registers it:
//
//	var Provider = toolsdk.Register(toolsdk.Spec{
//	    Name:        "branching-flow",
//	    Description: "Semantic analysis (14 analyzers)",
//	    Trigger:     toolsdk.OnGoFiles(),
//	    Inputs:      []string{"**/*.go"},
//	    DependsOn:   []string{"go-fix", "workspace-build-verify"},
//	    Detect:      analysis.NewDetector(),  // a finding.Detector
//	})
//
// BuildFlow discovers registered specs via toolsdk.All() and converts each into
// a domain.Tool via its internal ToolFromSpec converter. Adding a new tool is a
// one-file change in the tool's own repo — zero files in BuildFlow.
//
// Design constraint: this package depends ONLY on go-finding (the ecosystem
// hub). It must never import BuildFlow's domain/execution/tools packages, so
// tools that target this SDK are not coupled to BuildFlow's release cycle.
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
