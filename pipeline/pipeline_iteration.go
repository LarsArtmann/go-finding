package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/larsartmann/go-finding"
)

// runIteration executes one detect → triage → apply cycle.
// Returns (true, nil) when the pipeline should stop (no findings found).
func (p *Pipeline) runIteration(ctx context.Context, result *PipelineResult) (bool, error) {
	iter := Iteration{Number: p.iterations + 1} //nolint:exhaustruct

	detectDone := p.stageTiming(StageDetect)

	err := p.fireStageHook(ctx, StageBefore, StageDetect, iter.Number, nil, 0, 0)
	if err != nil {
		return false, fmt.Errorf("iteration %d: before detect: %w", p.iterations+1, err)
	}

	detResult, err := p.detect(ctx)

	detectDone()

	if err != nil {
		return false, fmt.Errorf("iteration %d: detect: %w", p.iterations+1, err)
	}

	hookErr := p.fireStageHook(ctx, StageAfter, StageDetect, iter.Number, detResult.Findings, 0, 0)
	if hookErr != nil {
		return false, fmt.Errorf("iteration %d: after detect: %w", p.iterations+1, hookErr)
	}

	findings := detResult.Findings

	for _, proc := range p.config.Processors {
		findings, err = proc.Transform(ctx, findings)
		if err != nil {
			return false, fmt.Errorf(
				"iteration %d: transformer %s: %w",
				p.iterations+1,
				proc.Name(),
				err,
			)
		}
	}

	if len(p.config.Processors) > 0 {
		hookErr := p.fireStageHook(ctx, StageAfter, StageProcess, iter.Number, findings, 0, 0)
		if hookErr != nil {
			return false, fmt.Errorf("iteration %d: after process: %w", p.iterations+1, hookErr)
		}
	}

	for name, detErr := range detResult.Errors {
		if result.PartialErrors == nil {
			result.PartialErrors = make(map[string]error)
		}

		result.PartialErrors[name] = detErr
	}

	iter.FindingsFound = len(findings)
	iter.findings = findings
	p.findings = append(p.findings, findings...)

	if len(findings) == 0 {
		result.Reason = ReasonStable
		result.Iterations = append(result.Iterations, iter)

		p.log(
			ctx, "iteration complete: stable (no findings)",
			slog.Int("iteration", iter.Number),
		)

		return true, nil
	}

	triage := p.triage(findings)
	iter.DirectFixes = len(triage.Direct)
	iter.SuggestFixes = len(triage.Suggest)
	iter.suggest = triage.Suggest
	iter.NoFix = len(triage.None)

	p.log(
		ctx, "triage complete",
		slog.Int("iteration", iter.Number),
		slog.Int("direct", len(triage.Direct)),
		slog.Int("suggest", len(triage.Suggest)),
		slog.Int("none", len(triage.None)),
	)

	hookErr = p.fireStageHook(ctx, StageAfter, StageTriage, iter.Number, findings, 0, 0)
	if hookErr != nil {
		return false, fmt.Errorf("iteration %d: after triage: %w", p.iterations+1, hookErr)
	}

	if !p.config.DryRun {
		err := p.fireStageHook(ctx, StageBefore, StageApply, iter.Number, triage.Direct, 0, 0)
		if err != nil {
			return false, fmt.Errorf("iteration %d: before apply: %w", p.iterations+1, err)
		}

		applyDone := p.stageTiming(StageApply)

		err = p.applyTriage(ctx, triage.Direct, &iter, result)
		if err != nil {
			applyDone()

			return false, fmt.Errorf("iteration %d: %w", p.iterations+1, err)
		}

		applyDone()

		hookErr := p.fireStageHook(
			ctx,
			StageAfter,
			StageApply,
			iter.Number,
			triage.Direct,
			iter.Applied,
			iter.Conflicts,
		)
		if hookErr != nil {
			applyDone()

			return false, fmt.Errorf("iteration %d: after apply: %w", p.iterations+1, hookErr)
		}
	}

	result.Iterations = append(result.Iterations, iter)
	p.iterations++

	if p.config.OnIteration != nil {
		p.config.OnIteration(p.iterations, findings)
	}

	return false, nil
}

// collectAllFindings gathers all findings from all iterations for verification.
func (*Pipeline) collectAllFindings(
	result *PipelineResult,
) []finding.Finding {
	seen := make(map[finding.ID]struct{})

	var all []finding.Finding

	for _, iter := range result.Iterations {
		for _, f := range iter.findings {
			if _, exists := seen[f.ID]; !exists {
				seen[f.ID] = struct{}{}
				all = append(all, f)
			}
		}
	}

	return all
}
