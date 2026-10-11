package toolsdk

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
)

func testOptionSpec(t *testing.T) Spec {
	t.Helper()

	return Spec{
		Name:        "tool",
		Description: "test tool",
		Detect: finding.NamedDetectorFunc(
			"tool",
			func(context.Context) ([]finding.Finding, error) { return nil, nil },
		),
		Options: []Option{
			{Name: "threshold", Kind: OptionKindInt, Default: 5},
			{Name: "mode", Kind: OptionKindString},
		},
	}
}

func TestOptionValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opt      Option
		wantErr  error
		contains string
	}{
		{
			name: "valid int option with default",
			opt:  Option{Name: "threshold", Kind: OptionKindInt, Default: 5, Description: "min statements"},
		},
		{
			name: "valid option without default",
			opt:  Option{Name: "mode", Kind: OptionKindString},
		},
		{
			name:     "empty name",
			opt:      Option{Kind: OptionKindInt},
			wantErr:  ErrInvalidOption,
			contains: "empty Name",
		},
		{
			name:     "empty kind",
			opt:      Option{Name: "x"},
			wantErr:  ErrInvalidOption,
			contains: "empty Kind",
		},
		{
			name:     "default kind mismatch",
			opt:      Option{Name: "x", Kind: OptionKindInt, Default: "five"},
			wantErr:  ErrInvalidOption,
			contains: `Default string does not match Kind "int"`,
		},
		{
			name:     "unknown kind",
			opt:      Option{Name: "x", Kind: OptionKind("float")},
			wantErr:  ErrInvalidOption,
			contains: `unknown Kind "float"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.opt.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) || !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("Validate() = %v, want %v containing %q", err, tt.wantErr, tt.contains)
			}
		})
	}
}

func TestWithOptionsRoundTrip(t *testing.T) {
	t.Parallel()

	if _, ok := OptionsFromContext(context.Background()); ok {
		t.Fatal("OptionsFromContext on a plain context must report absent")
	}

	nilValues, nilOK := OptionsFromContext(nil) //nolint:staticcheck // SA1012: nil ctx is the contract under test
	if nilOK {
		t.Fatal("OptionsFromContext on nil context must report absent")
	}

	if nilValues != nil {
		t.Fatal("OptionsFromContext on nil context must return nil values")
	}

	values := OptionValues{"threshold": 10}
	ctx := WithOptions(context.Background(), values)

	got, ok := OptionsFromContext(ctx)
	if !ok {
		t.Fatal("OptionsFromContext must find values set via WithOptions")
	}

	if got["threshold"] != 10 {
		t.Fatalf("threshold = %v, want 10", got["threshold"])
	}
}

func TestWithOptionsSnapshotsValues(t *testing.T) {
	t.Parallel()

	values := OptionValues{"threshold": 10}
	ctx := WithOptions(context.Background(), values)

	values["threshold"] = 999
	delete(values, "threshold")

	got, ok := OptionsFromContext(ctx)
	if !ok {
		t.Fatal("OptionsFromContext must find values set via WithOptions")
	}

	if got["threshold"] != 10 {
		t.Fatalf("threshold = %v, want snapshotted 10 (later mutation must not leak)", got["threshold"])
	}
}

func TestWithOptionsEmptyClearsInherited(t *testing.T) {
	t.Parallel()

	base := WithOptions(context.Background(), OptionValues{"threshold": 10})

	if _, ok := OptionsFromContext(base); !ok {
		t.Fatal("precondition: base context must carry values")
	}

	cleared := WithOptions(base, nil)
	if _, ok := OptionsFromContext(cleared); ok {
		t.Fatal("WithOptions with nil values must clear inherited values")
	}

	clearedEmpty := WithOptions(base, OptionValues{})
	if _, ok := OptionsFromContext(clearedEmpty); ok {
		t.Fatal("WithOptions with empty values must clear inherited values")
	}
}

func TestValidateOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		noDecl   bool
		values   OptionValues
		wantErr  error
		contains string
	}{
		{name: "no values always valid", values: nil},
		{name: "declared int", values: OptionValues{"threshold": 10}},
		{name: "declared string", values: OptionValues{"mode": "strict"}},
		{name: "both declared", values: OptionValues{"threshold": 1, "mode": "x"}},
		{
			name:     "unknown name is rejected",
			values:   OptionValues{"threashold": 10},
			wantErr:  ErrUnknownOption,
			contains: `does not declare option "threashold"`,
		},
		{
			name:     "kind mismatch is rejected",
			values:   OptionValues{"threshold": "10"},
			wantErr:  ErrOptionKindMismatch,
			contains: `got string, want kind "int"`,
		},
		{
			name:     "no declared options rejects any value",
			noDecl:   true,
			values:   OptionValues{"threshold": 10},
			wantErr:  ErrUnknownOption,
			contains: `does not declare option "threshold"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			spec := testOptionSpec(t)
			if tt.noDecl {
				spec.Options = nil
			}

			err := spec.ValidateOptions(tt.values)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("ValidateOptions() = %v, want nil", err)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) || !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("ValidateOptions() = %v, want %v containing %q", err, tt.wantErr, tt.contains)
			}
		})
	}
}

func TestRegisterRejectsBadOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		spec    Spec
		wantMsg string
	}{
		{
			name: "duplicate option names",
			spec: Spec{
				Name:        "dup-options-tool",
				Description: "must fail registration",
				Detect: finding.NamedDetectorFunc(
					"dup-options-tool",
					func(context.Context) ([]finding.Finding, error) { return nil, nil },
				),
				Options: []Option{
					{Name: "threshold", Kind: OptionKindInt, Default: 5},
					{Name: "threshold", Kind: OptionKindInt, Default: 6},
				},
			},
			wantMsg: "declares option threshold twice",
		},
		{
			name: "invalid option declaration",
			spec: Spec{
				Name:        "bad-option-tool",
				Description: "must fail registration",
				Detect: finding.NamedDetectorFunc(
					"bad-option-tool",
					func(context.Context) ([]finding.Finding, error) { return nil, nil },
				),
				Options: []Option{
					{Name: "threshold", Kind: OptionKind("float")},
				},
			},
			wantMsg: `unknown Kind "float"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("Register with malformed options must panic")
				}

				msg, ok := r.(string)
				if !ok || !strings.Contains(msg, tt.wantMsg) {
					t.Fatalf("panic = %v, want message containing %q", r, tt.wantMsg)
				}
			}()

			Register(tt.spec)
		})
	}
}

// TestTypedOptionAccessorsReadValues covers the happy path (issue #43): each
// accessor returns the typed per-run value without error.
func TestTypedOptionAccessorsReadValues(t *testing.T) {
	t.Parallel()

	ctx := WithOptions(context.Background(), OptionValues{
		"mode":      "strict",
		"threshold": 7,
		"verbose":   true,
	})

	if got, err := StringFromContext(ctx, "mode", "all"); err != nil || got != "strict" {
		t.Fatalf("StringFromContext = (%q, %v), want (strict, nil)", got, err)
	}

	if got, err := IntFromContext(ctx, "threshold", 5); err != nil || got != 7 {
		t.Fatalf("IntFromContext = (%d, %v), want (7, nil)", got, err)
	}

	if got, err := BoolFromContext(ctx, "verbose", false); err != nil || !got {
		t.Fatalf("BoolFromContext = (%t, %v), want (true, nil)", got, err)
	}
}

// TestTypedOptionAccessorsFallback covers every miss shape (issue #43):
// unset name, empty set, nil context, and unknown name all yield the
// tool's fallback with no error.
func TestTypedOptionAccessorsFallback(t *testing.T) {
	t.Parallel()

	set := WithOptions(context.Background(), OptionValues{"mode": "strict"})
	empty := WithOptions(context.Background(), OptionValues{})

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "plain context", ctx: context.Background()},
		{name: "nil context", ctx: nil},
		{name: "empty option set", ctx: empty},
		{name: "unknown option name", ctx: set},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, err := StringFromContext(tt.ctx, "missing", "all"); err != nil || got != "all" {
				t.Fatalf("StringFromContext = (%q, %v), want fallback (all, nil)", got, err)
			}

			if got, err := IntFromContext(tt.ctx, "missing", 5); err != nil || got != 5 {
				t.Fatalf("IntFromContext = (%d, %v), want fallback (5, nil)", got, err)
			}

			if got, err := BoolFromContext(tt.ctx, "missing", true); err != nil || !got {
				t.Fatalf("BoolFromContext = (%t, %v), want fallback (true, nil)", got, err)
			}
		})
	}
}

// TestTypedOptionAccessorsNilValueFallsBack pins that a nil stored value is
// treated as "nothing set": the accessor yields the fallback, no error.
func TestTypedOptionAccessorsNilValueFallsBack(t *testing.T) {
	t.Parallel()

	ctx := WithOptions(context.Background(), OptionValues{"mode": nil})

	got, err := StringFromContext(ctx, "mode", "all")
	if err != nil || got != "all" {
		t.Fatalf("StringFromContext with nil value = (%q, %v), want fallback (all, nil)", got, err)
	}
}

// TestTypedOptionAccessorsKindMismatch covers the belt-and-braces check
// (issue #43): a present value of the wrong Go type fails with
// ErrOptionKindMismatch naming the option and both kinds.
func TestTypedOptionAccessorsKindMismatch(t *testing.T) {
	t.Parallel()

	ctx := WithOptions(context.Background(), OptionValues{
		"mode":      42,
		"threshold": "7",
		"verbose":   "yes",
	})

	tests := []struct {
		name     string
		read     func() error
		contains string
	}{
		{
			name: "string accessor sees int value",
			read: func() error {
				_, err := StringFromContext(ctx, "mode", "all")

				return err
			},
			contains: `option "mode": got int, want kind "string"`,
		},
		{
			name: "int accessor sees string value",
			read: func() error {
				_, err := IntFromContext(ctx, "threshold", 5)

				return err
			},
			contains: `option "threshold": got string, want kind "int"`,
		},
		{
			name: "bool accessor sees string value",
			read: func() error {
				_, err := BoolFromContext(ctx, "verbose", true)

				return err
			},
			contains: `option "verbose": got string, want kind "bool"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.read()
			if !errors.Is(err, ErrOptionKindMismatch) || !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("err = %v, want ErrOptionKindMismatch containing %q", err, tt.contains)
			}
		})
	}
}
