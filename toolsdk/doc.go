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
package toolsdk
