<!-- proposed title:
Register validates Options but not Trigger — a typo'd Language registers fine and then silently never matches
labels: bug
-->

## Why

`Register` panics on malformed registrations — empty Name, missing capability, malformed Options (`toolsdk/registry.go:26-29`, `mustValidate` at `registry.go:88-113`). `Trigger` is not validated. `Trigger.Language` documents exactly six valid values plus `""` (`spec.go:89-90`), but `Language: "golang"` (the most natural typo) registers cleanly and then fails silently downstream: BuildFlow's conversion maps it verbatim into `OnlyLanguage(config.Language(t.Language))` (BuildFlow `tools/providers/sdk_registry.go:194-204`), and a matcher built on a language nothing ever detects never matches — the tool just never runs. No error, no warning, empty pipeline.

That is exactly the failure mode the Register panic convention exists to prevent ("a malformed registration is a programming error that should surface at startup, not at runtime" — `registry.go:26-29`). It just doesn't cover Trigger.

## Source-level cause

```go
// toolsdk/registry.go:88-99 (v1.15.0, go-finding @ 91a216f)
func mustValidate(s Spec) {
	if s.Name == "" {
		panic("toolsdk.Register: Spec.Name is required")
	}
	if strings.TrimSpace(s.Description) == "" {
		panic("toolsdk.Register: Spec " + s.Name + " has no Description")
	}
	if s.Detect == nil && s.Repair == nil {
		panic("toolsdk.Register: Spec " + s.Name + " has no Detect or Repair capability")
	}
	// Options validated below — Trigger: nothing
}
```

## Proposal

- [ ] Validate Trigger in `mustValidate`: Language ∈ `{"", "go", "js", "ts", "python", "nix", "rust"}` — export the set (e.g. `ValidLanguages`) so consumers and docs stay in sync
- [ ] Panic message names the field and the valid set, same style as the Option panics
- [ ] Optionally reject empty-string entries in `Trigger.Files`/`Requires`/`NotRequires` (a `""` pattern matches nothing useful)
- [ ] `spec_test.go`: table cases per invalid value, including `"golang"`

## To verify

`Register` with `Language: "golang"` panics at startup; the 11 specs BuildFlow blank-imports today pass unchanged; `spec_test.go` covers the new panics.

💘 Generated with Crush
