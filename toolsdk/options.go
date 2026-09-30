package toolsdk

import (
	"context"
	"errors"
	"fmt"
	"maps"
)

// Sentinel errors for option declarations and per-run option values. The
// wrapped detail names the offending option; match with errors.Is.
var (
	// ErrInvalidOption reports a malformed Option declaration: empty name,
	// empty or unknown kind, or a Default that does not match the kind.
	ErrInvalidOption = errors.New("toolsdk: invalid option declaration")
	// ErrDuplicateOption reports a Spec declaring the same option name twice.
	ErrDuplicateOption = errors.New("toolsdk: duplicate option declaration")
	// ErrUnknownOption reports consumer-supplied values naming an option the
	// Spec does not declare — usually a config typo.
	ErrUnknownOption = errors.New("toolsdk: unknown option")
	// ErrOptionKindMismatch reports a consumer-supplied value whose Go type
	// does not match the declared OptionKind.
	ErrOptionKindMismatch = errors.New("toolsdk: option kind mismatch")
)

// OptionKind classifies the value shape of a declared Spec option. The SDK
// checks kind only — range and semantic validation stay with the tool
// implementation, which owns the option's meaning.
type OptionKind string

// Supported option kinds. A kind mismatch between a consumer-supplied value
// and the declared kind fails ValidateOptions before the tool ever runs.
const (
	OptionKindInt    OptionKind = "int"
	OptionKindString OptionKind = "string"
	OptionKindBool   OptionKind = "bool"
)

// Option declares one tunable knob a Spec's Detect/Repair implementation
// reads from the context (see OptionsFromContext). Declarations power
// consumer-side validation (Spec.ValidateOptions) and documentation
// (BuildFlow shows them next to the tool in --list output); the VALUES ride
// the context and are set by the consumer per run via WithOptions.
type Option struct {
	// Name is the config key consumers set (e.g. "threshold"). Unique within
	// a Spec's Options list.
	Name string
	// Kind is the value shape; consumer values of a different kind are
	// rejected by ValidateOptions.
	Kind OptionKind
	// Default documents the tool's built-in fallback. Informational: the
	// SDK does not inject it — when no value is set the tool applies its
	// own default. Must match Kind when non-nil.
	Default any
	// Description is a human-readable explanation shown to consumers.
	Description string
}

// Validate checks the declaration itself: non-empty name, known kind, and a
// default (when set) that matches the declared kind. Returns an error
// matching ErrInvalidOption.
func (opt Option) Validate() error {
	switch {
	case opt.Name == "":
		return fmt.Errorf("%w: empty Name", ErrInvalidOption)
	case opt.Kind == "":
		return fmt.Errorf("%w: option %q: empty Kind", ErrInvalidOption, opt.Name)
	}

	var defaultMatchesKind bool

	switch opt.Kind {
	case OptionKindInt:
		_, defaultMatchesKind = opt.Default.(int)
	case OptionKindString:
		_, defaultMatchesKind = opt.Default.(string)
	case OptionKindBool:
		_, defaultMatchesKind = opt.Default.(bool)
	default:
		return fmt.Errorf("%w: option %q: unknown Kind %q", ErrInvalidOption, opt.Name, opt.Kind)
	}

	switch {
	case opt.Default == nil:
		return nil
	case !defaultMatchesKind:
		return fmt.Errorf("%w: option %q: Default %T does not match Kind %q",
			ErrInvalidOption, opt.Name, opt.Default, opt.Kind)
	}

	return nil
}

// optionsCtxKey carries per-run option values for one Detect/Repair call.
type optionsCtxKey struct{}

// OptionValues are per-run option values a consumer resolved for one tool,
// keyed by the declared Option.Name. Treated as read-only once handed to
// WithOptions, which snapshots the map.
type OptionValues map[string]any

// WithOptions returns a context carrying values for the tool's declared
// options. BuildFlow's execution layer calls it when constructing the
// context for a Detect/Repair call; tools read them back via
// OptionsFromContext. The map is snapshotted: later mutation of the caller's
// map does not affect the run. Nil or empty values CLEAR any options
// inherited from a parent context — each run carries exactly the values it
// set, and a run that sets none runs on the tool's defaults.
func WithOptions(ctx context.Context, values OptionValues) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	frozen := make(OptionValues, len(values))
	maps.Copy(frozen, values)

	return context.WithValue(ctx, optionsCtxKey{}, frozen)
}

// OptionsFromContext reads the per-run option values from the context.
// Returns (nil, false) when none were set (or an empty set was set); the
// tool then applies its own defaults.
func OptionsFromContext(ctx context.Context) (OptionValues, bool) {
	if ctx == nil {
		return nil, false
	}

	values, ok := ctx.Value(optionsCtxKey{}).(OptionValues)
	if !ok || len(values) == 0 {
		return nil, false
	}

	return values, true
}

// ValidateOptions checks consumer-supplied values against the spec's
// declared Options. Unknown names are rejected (a config typo fails loudly
// instead of silently running on defaults) and value kinds must match the
// declarations. Missing names are fine — the tool applies its own default.
// A spec with no declared Options rejects any value: declaring knobs is
// what makes them settable.
func (s Spec) ValidateOptions(values OptionValues) error {
	if len(values) == 0 {
		return nil
	}

	declared := make(map[string]Option, len(s.Options))
	for _, opt := range s.Options {
		if err := opt.Validate(); err != nil {
			return err
		}

		if _, dup := declared[opt.Name]; dup {
			return fmt.Errorf("%w: spec %q declares option %q twice",
				ErrDuplicateOption, s.Name, opt.Name)
		}

		declared[opt.Name] = opt
	}

	for name, value := range values {
		decl, known := declared[name]
		if !known {
			return fmt.Errorf("%w: spec %q does not declare option %q",
				ErrUnknownOption, s.Name, name)
		}

		var matchesKind bool

		switch decl.Kind {
		case OptionKindInt:
			_, matchesKind = value.(int)
		case OptionKindString:
			_, matchesKind = value.(string)
		case OptionKindBool:
			_, matchesKind = value.(bool)
		}

		if !matchesKind {
			return fmt.Errorf("%w: spec %q option %q: got %T, want kind %q",
				ErrOptionKindMismatch, s.Name, name, value, decl.Kind)
		}
	}

	return nil
}
