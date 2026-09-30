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
		Detect:      finding.NamedDetectorFunc("tool", func(context.Context) ([]finding.Finding, error) { return nil, nil }),
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

	if _, ok := OptionsFromContext(nil); ok { //nolint:staticcheck // SA1012: nil-context handling is the documented contract
		t.Fatal("OptionsFromContext on nil context must report absent")
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
				Detect:      finding.NamedDetectorFunc("dup-options-tool", func(context.Context) ([]finding.Finding, error) { return nil, nil }),
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
				Detect:      finding.NamedDetectorFunc("bad-option-tool", func(context.Context) ([]finding.Finding, error) { return nil, nil }),
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
