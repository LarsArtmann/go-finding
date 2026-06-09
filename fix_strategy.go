package finding

// FixStrategy indicates how a finding can be remediated.
type FixStrategy string

const (
	// FixStrategyNone indicates no fix is available.
	FixStrategyNone FixStrategy = "none"
	// FixStrategySuggest provides a human-readable suggestion.
	FixStrategySuggest FixStrategy = "suggest"
	// FixStrategyDirect can be automatically applied.
	FixStrategyDirect FixStrategy = "direct"
	// FixStrategyAI requires AI assistance.
	// Pipeline triage groups this with FixStrategySuggest (no auto-apply).
	// NeedsAI() is defined but no AI backend exists yet. Reserve this value
	// for future AI-powered remediation — do not remove.
	FixStrategyAI FixStrategy = "ai"
)

// IsValid returns true if the fix strategy is a valid value.
func (f FixStrategy) IsValid() bool {
	switch f {
	case FixStrategyNone, FixStrategySuggest, FixStrategyDirect, FixStrategyAI:
		return true
	}

	return false
}

// CanAutoApply returns true if this fix strategy can be automatically applied.
func (f FixStrategy) CanAutoApply() bool {
	return f == FixStrategyDirect
}

// NeedsAI returns true if this fix strategy requires AI assistance.
func (f FixStrategy) NeedsAI() bool {
	return f == FixStrategyAI
}

// String returns the string representation of the fix strategy.
func (f FixStrategy) String() string {
	return string(f)
}

// FixStrategyResolver determines whether a finding can be auto-applied.
// Implement this interface to customize fix strategy behavior beyond the
// built-in FixStrategy constants.
type FixStrategyResolver interface {
	// CanAutoApply reports whether findings with the given strategy can be
	// automatically applied by the pipeline.
	CanAutoApply(strategy FixStrategy) bool
}

// DefaultResolver is the standard FixStrategyResolver that matches
// [FixStrategy.CanAutoApply].
type DefaultResolver struct{}

// CanAutoApply returns true only for [FixStrategyDirect].
func (DefaultResolver) CanAutoApply(strategy FixStrategy) bool {
	return strategy.CanAutoApply()
}
