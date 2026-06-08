package finding

import (
	"errors"
	"fmt"
)

func (f Finding) Validate() error {
	var errs []error

	if f.ID == "" {
		errs = append(errs, NewValidationError("finding.ID is required", nil))
	}

	if f.Rule == "" {
		errs = append(errs, NewValidationError("finding.Rule is required", nil))
	}

	if f.ToolName == "" {
		errs = append(errs, NewValidationError("finding.ToolName is required", nil))
	}

	if f.Message == "" {
		errs = append(errs, NewValidationError("finding.Message is required", nil))
	}

	if !f.Severity.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Severity %q is invalid", f.Severity), nil,
		))
	}

	if !f.Position.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Position %+v is invalid", f.Position), nil,
		))
	}

	if f.FixStrategy != "" && !f.FixStrategy.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.FixStrategy %q is invalid", f.FixStrategy), nil,
		))
	}

	if !f.Confidence.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Confidence %s must be in [0.0, 1.0]", f.Confidence), nil,
		))
	}

	if f.BeforeCode == "" && f.AfterCode == "" && f.FixStrategy == FixStrategyDirect {
		errs = append(errs, NewValidationError(
			"finding.FixStrategyDirect requires BeforeCode or AfterCode", nil,
		))
	}

	for i, tag := range f.Tags {
		if !tag.IsValid() {
			errs = append(errs, NewValidationError(
				fmt.Sprintf("finding.Tags[%d] %q is invalid: tags must be non-empty", i, tag), nil,
			))
		}
	}

	for i, ref := range f.Related {
		if !ref.IsValid() {
			errs = append(errs, NewValidationError(
				fmt.Sprintf("finding.Related[%d] is invalid: missing required fields", i), nil,
			))
		}

		if ref.Range != nil && ref.Range.IsInverted() {
			errs = append(errs, NewValidationError(
				fmt.Sprintf(
					"finding.Related[%d].Range is invalid: End (%+v) before Start (%+v)",
					i, ref.Range.End, ref.Range.Start,
				),
				nil,
			))
		}
	}

	if f.Suppression != nil && !f.Suppression.IsValid() {
		errs = append(errs, NewValidationError(
			"finding.Suppression is invalid: missing Kind or Rule", nil,
		))
	}

	if f.Range != nil && f.Range.IsInverted() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf(
				"finding.Range is invalid: End (%+v) before Start (%+v)",
				f.Range.End,
				f.Range.Start,
			),
			nil,
		))
	}

	return errors.Join(errs...)
}

// IsValid returns true if the finding has required fields set.
func (f Finding) IsValid() bool {
	return f.ID != "" && f.Rule != "" && f.ToolName != "" &&
		f.Message != "" && f.Position.IsValid() && f.Severity.IsValid()
}

// Key returns a stable identifier for the finding.
// If ID is set, it is returned; otherwise a deterministic key is built
// from ToolName, Position.File, Rule, and Message using keySeparator.
func (f Finding) Key() string {
	if f.ID != "" {
		return f.ID
	}

	return f.ToolName + keySeparator + f.Position.File + keySeparator + f.Rule + keySeparator + f.Message
}
