package toolsdk

import (
	"context"
	"fmt"
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
// default (when set) that matches the declared kind.
func (o Option) Validate() error {
	switch {
	case o.Name == "":
		return fmt.Errorf("toolsdk option declaration invalid: empty Name")
	case o.Kind == "":
		return fmt.Errorf("toolsdk option %q invalid: empty Kind", o.Name)
	}

	var defaultOK bool
	switch o.Kind {
	case OptionKindInt:
		_, defaultOK = o.Default.(int)
	case OptionKindString:
		_, defaultOK = o.Default.(string)
	case OptionKindBool:
		_, defaultOK = o.Default.(bool)
	default:
		return fmt.Errorf("toolsdk option %q invalid: unknown Kind %q", o.Name, o.Kind)
	}

	switch {
	case o.Default == nil:
		return nil
	case !defaultOK:
		return fmt.Errorf("toolsdk option %q invalid: Default %T does not match Kind %q",
			o.Name, o.Default, o.Kind)
	}

	return nil
}

// optionsCtxKey carries per-run option values for one Detect/Repair call.
type optionsCtxKey struct{}

// OptionValues are per-run option values a consumer resolved for one tool,
// keyed by the declared Option.Name.
type OptionValues map[string]any

// WithOptions returns a context carrying values for the tool's declared
// options. BuildFlow's execution layer calls it when constructing the
// context for a Detect/Repair call; tools read them back via
// OptionsFromContext. Nil or empty values are equivalent to not calling it.
func WithOptions(ctx context.Context, values OptionValues) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(values) == 0 {
		return ctx
	}

	return context.WithValue(ctx, optionsCtxKey{}, values)
}

// OptionsFromContext reads the per-run option values from the context.
// Returns (nil, false) when none were set; the tool then applies its own
// defaults.
func OptionsFromContext(ctx context.Context) (OptionValues, bool) {
	if ctx == nil {
		return nil, false
	}

	v, ok := ctx.Value(optionsCtxKey{}).(OptionValues)

	return v, ok
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
	for _, o := range s.Options {
		if err := o.Validate(); err != nil {
			return err
		}

		if _, dup := declared[o.Name]; dup {
			return fmt.Errorf("toolsdk spec %q declares option %q twice", s.Name, o.Name)
		}

		declared[o.Name] = o
	}

	for name, value := range values {
		o, known := declared[name]
		if !known {
			return fmt.Errorf("toolsdk spec %q does not declare option %q", s.Name, name)
		}

		var ok bool
		switch o.Kind {
		case OptionKindInt:
			_, ok = value.(int)
		case OptionKindString:
			_, ok = value.(string)
		case OptionKindBool:
			_, ok = value.(bool)
		}

		if !ok {
			return fmt.Errorf("toolsdk spec %q option %q: got %T, want kind %q",
				s.Name, name, value, o.Kind)
		}
	}

	return nil
}
