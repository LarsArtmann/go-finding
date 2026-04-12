package finding

// FixStrategy indicates how a finding can be remediated.
type FixStrategy string

const (
	FixStrategyNone    FixStrategy = "none"    // No fix available
	FixStrategySuggest FixStrategy = "suggest" // Human-readable suggestion, not machine-applicable
	FixStrategyDirect  FixStrategy = "direct"  // Deterministic code transformation
	FixStrategyAI      FixStrategy = "ai"      // Requires AI/LLM to generate fix
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
