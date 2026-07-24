<!-- Thanks for contributing! Fill in the sections below. -->

## Summary

<!-- What does this PR change and why? One or two sentences. -->

## Motivation

<!-- Link the issue this closes (e.g. "Closes #123"), or describe the problem. -->

## Changes

<!-- Bullet list of what changed. Note any breaking changes explicitly. -->

-

## Checklist

- [ ] Tests pass: `export GOEXPERIMENT=jsonv2 && go test -race -count=1 ./...`
- [ ] Lint passes: `golangci-lint run ./...` (or `nix run .#lint`)
- [ ] New code is covered by tests
- [ ] No deprecated APIs introduced (the API is frozen since `v1.0.0`)
- [ ] Breaking change? If yes, it requires a major version bump — flag it here
- [ ] Public API changes documented (godoc examples / README as needed)

## Notes for review

<!-- Anything reviewers should pay attention to, edge cases, or tradeoffs. -->
