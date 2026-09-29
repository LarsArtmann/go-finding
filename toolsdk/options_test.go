package toolsdk

import (
	"context"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding/finding"
)

func TestOptionValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opt     Option
		wantErr string
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
			name:    "empty name",
			opt:     Option{Kind: OptionKindInt},
			wantErr: "empty Name",
		},
		{
			name:    "empty kind",
			opt:     Option{Name: "x"},
			wantErr: "empty Kind",
		},
		{
			name:    "default kind mismatch",
			opt:     Option{Name: "x", Kind: OptionKindInt, Default: "five"},
			wantErr: `Default string does not match Kind "int"`,
		},
		{
			name:    "unknown kind",
			opt:     Option{Name: "x", Kind: OptionKind("float")},
			wantErr: `unknown Kind "float"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.opt.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestWithOptionsRoundTrip(t *testing.T) {
	t.Parallel()

	if _, ok := OptionsFromContext(context.Background()); ok {
		t.Fatal("OptionsFromContext on a plain context must report absent")
	}

	if _, ok := OptionsFromContext(nil); ok {
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

func TestWithOptionsEmptyIsNoop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	if got := WithOptions(ctx, nil); got != ctx {
		t.Fatal("WithOptions with nil values must return the context unchanged")
	}

	if got := WithOptions(ctx, OptionValues{}); got != ctx {
		t.Fatal("WithOptions with empty values must return the context unchanged")
	}
}

func TestValidateOptions(t *testing.T) {
	t.Parallel()

	spec := Spec{
		Name:        "tool",
		Description: "test tool",
		Detect:      stubDetector{},
		Options: []Option{
			{Name: "threshold", Kind: OptionKindInt, Default: 5},
			{Name: "mode", Kind: OptionKindString},
		},
	}

	tests := []struct {
		name    string
		values  OptionValues
		wantErr string
	}{
		{name: "no values always valid", values: nil},
		{name: "declared int", values: OptionValues{"threshold": 10}},
		{name: "declared string", values: OptionValues{"mode": "strict"}},
		{name: "both declared", values: OptionValues{"threshold": 1, "mode": "x"}},
		{
			name:    "unknown name is rejected",
			values:  OptionValues{"threashold": 10},
			wantErr: `does not declare option "threashold"`,
		},
		{
			name:    "kind mismatch is rejected",
			values:  OptionValues{"threshold": "10"},
			wantErr: `got string, want kind "int"`,
		},
		{
			name:    "no declared options rejects any value",
			values:  OptionValues{"threshold": 10},
			wantErr: `does not declare option "threshold"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := spec
			if tt.name == "no declared options rejects any value" {
				target.Options = nil
			}

			err := target.ValidateOptions(tt.values)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateOptions() = %v, want nil", err)
				}

				return
			}

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ValidateOptions() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterRejectsBadOptions(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Register with duplicate option names must panic")
		}

		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "declares option \"threshold\" twice") {
			t.Fatalf("panic = %v, want duplicate-option message", r)
		}
	}()

	Register(Spec{
		Name:        "dup-options-tool",
		Description: "must fail registration",
		Detect:      stubDetector{},
		Options: []Option{
			{Name: "threshold", Kind: OptionKindInt, Default: 5},
			{Name: "threshold", Kind: OptionKindInt, Default: 6},
		},
	})
}

// stubDetector satisfies the finding.Detector interface for spec tests.
type stubDetector struct{}

func (stubDetector) Name() string { return "stub" }

func (stubDetector) Detect(context.Context) ([]finding.Finding, error) {
	return nil, nil
}
