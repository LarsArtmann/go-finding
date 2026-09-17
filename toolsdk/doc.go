// Package toolsdk is the self-registration plugin contract for BuildFlow
// providers: a tool repo calls Register(Spec{...}) at package init, and
// BuildFlow discovers it via a blank import plus ToolsFromSDK.
//
// Boundary: this package is the STABLE surface third-party tools compile
// against. It deliberately exposes plain strings, go-finding types, and
// function values only — no BuildFlow-internal types (domain/tool, config),
// so a tool never imports BuildFlow itself. The one-way conversion from this
// contract to BuildFlow's branded types lives exclusively in BuildFlow
// (tools/providers/sdk_registry.go::ToolFromSpec); nothing flows back.
//
// Registration rules (validated eagerly by Register, gotcha #21): Name and a
// trimmed non-empty Description are required, and the Spec must carry a
// capability (Detect or Repair). Tests that mutate the process-global
// registry must snapshot with SnapshotForTest and restore via
// t.Cleanup(RestoreForTest) — the registry is shared for the whole process
// and the workspace test gate runs shuffled.
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
