# Migration to Nix Flakes — Proposal

**Status:** Draft | **Date:** 2026-04-21 | **Project:** go-finding

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Why Nix Flakes](#why-nix-flakes)
4. [Proposed Architecture](#proposed-architecture)
5. [File-by-File Specification](#file-by-file-specification)
6. [Dependency Inventory](#dependency-inventory)
7. [Migration Phases](#migration-phases)
8. [CI/CD Integration](#cicd-integration)
9. [Developer Experience](#developer-experience)
10. [Rollback Plan](#rollback-plan)
11. [Risks and Mitigations](#risks-and-mitigations)
12. [Decision Record](#decision-record)

---

## Executive Summary

**go-finding** currently relies on ad-hoc tool installation (Go, golangci-lint, goreleaser, just, govulncheck, goimports) with no reproducibility guarantees across developer machines or CI runners. This proposal introduces **Nix Flakes** as a single, declarative source of truth for the entire development and build environment.

### What This Solves

| Problem                | Current State                                                       | With Nix Flakes                |
| ---------------------- | ------------------------------------------------------------------- | ------------------------------ |
| Tool version drift     | "Install Go 1.26+" in CONTRIBUTING.md                               | Pinned exactly in `flake.lock` |
| Dev environment setup  | Manual: go, golangci-lint, goreleaser, just, goimports, govulncheck | `nix develop` — one command    |
| CI reproducibility     | `actions/setup-go@v5` + `golangci-lint-action@v6` (version drift)   | Same flake in CI and local     |
| Cross-platform builds  | goreleaser handles it post-build                                    | Nix handles it natively        |
| Lint version pinning   | `golangci-lint-action@v6` with `version: v2.10.1`                   | `flake.lock` hash              |
| Vulnerability scanning | `govulncheck` (manual install)                                      | Available in dev shell         |

### What This Does NOT Solve

- Go dependency management (handled by `go.mod` — Nix does not replace this)
- SARIF/LSP output formats (unrelated to build system)
- Pipeline logic (pure Go code, unaffected)

---

## Current State Analysis

### Build & Development Tools

| Tool              | Version     | Source                           | How Installed                      |
| ----------------- | ----------- | -------------------------------- | ---------------------------------- |
| **Go**            | 1.26.0      | `go.mod`, `.golangci.yml`, CI    | Manual / `actions/setup-go@v5`     |
| **golangci-lint** | v2.10.1     | `.golangci.yml:61`, CI           | Manual / `golangci-lint-action@v6` |
| **goreleaser**    | latest      | `.goreleaser.yml`, CI            | Manual / `goreleaser-action@v6`    |
| **just**          | unspecified | `justfile`                       | Manual                             |
| **goimports**     | unspecified | `justfile:28` (`goimports -w .`) | Manual (`golang.org/x/tools`)      |
| **govulncheck**   | unspecified | `justfile:46`                    | Manual                             |
| **gofumpt**       | unspecified | `.golangci.yml:241` formatter    | Manual                             |
| **gci**           | unspecified | `.golangci.yml:239` formatter    | Manual                             |
| **golines**       | unspecified | `.golangci.yml:242` formatter    | Manual                             |

### Go Dependencies (from `go.mod`)

| Dependency           | Version | Purpose                         |
| -------------------- | ------- | ------------------------------- |
| `golang.org/x/sync`  | v0.20.0 | errgroup for parallel detection |
| `golang.org/x/tools` | v0.44.0 | go/analysis framework           |
| `gopkg.in/yaml.v3`   | v3.0.1  | YAML config parsing (CLI only)  |

### CI Workflows

**CI** (`.github/workflows/ci.yml`):

- Matrix: Go 1.26 on ubuntu-latest + macos-latest
- Steps: `go mod download` → `go build` → `go vet` → `go test -race` → coverage enforcement (≥75%)
- Coverage job: codecov upload
- Lint job: `golangci-lint-action@v6` with `version: v2.10.1`, `--timeout=5m`

**Release** (`.github/workflows/release.yml`):

- Trigger: `v*` tags
- Steps: test → vet → build → goreleaser release
- Cross-compile: linux/darwin/windows, amd64/arm64 (via goreleaser config)

### Justfile Recipes

| Recipe   | Commands                                  | Tools Required               |
| -------- | ----------------------------------------- | ---------------------------- |
| `test`   | `go test -v -race ./...`                  | go                           |
| `bench`  | `go test -bench=. -benchmem ./...`        | go                           |
| `cover`  | `go test -coverprofile` + `go tool cover` | go                           |
| `lint`   | `golangci-lint run ./...`                 | golangci-lint                |
| `fmt`    | `go fmt` + `goimports -w .`               | go, goimports                |
| `deps`   | `go mod download` + `go mod tidy`         | go                           |
| `clean`  | `rm -f` + `go clean -cache`               | go                           |
| `build`  | `go build ./...`                          | go                           |
| `vuln`   | `govulncheck ./...`                       | govulncheck                  |
| `check`  | `fmt` → `lint` → `test`                   | go, golangci-lint, goimports |
| `update` | `go get -u` + `go mod tidy`               | go                           |
| `docs`   | `go doc -all`                             | go                           |
| `ci`     | `deps` → `check`                          | all                          |

### golangci-lint Configuration

The project uses an extensive linter configuration (`.golangci.yml`) with **85+ enabled linters** and 4 formatters:

- **Linters:** `staticcheck`, `govet`, `gosec`, `revive`, `errcheck`, `gocritic`, `unparam`, `unused`, and 77 more
- **Formatters:** `gci`, `goimports`, `gofumpt`, `golines`
- **Build tags:** `goexperiment.arenas`, `goexperiment.goroutineleakprofile`, `goexperiment.jsonv2`, `goexperiment.runtimesecret`, `goexperiment.simd`
- **Timeout:** 5 minutes
- **Go version:** 1.26.0

### goreleaser Configuration

- **Binary:** `go-finding` from `./cmd/go-finding`
- **Platforms:** linux/darwin/windows × amd64/arm64
- **LD flags:** `-s -w -X main.version={{.Version}}`
- **Archives:** tar.gz (zip for windows)
- **Changelog:** auto-generated, filtered

---

## Why Nix Flakes

### Benefits for This Project

1. **Single source of truth** — `flake.nix` + `flake.lock` define every tool version precisely
2. **Reproducible builds** — Same inputs produce same outputs, bit-for-bit (modulo build tags)
3. **Zero-friction onboarding** — `nix develop` gives a complete dev environment
4. **CI/CD alignment** — Same flake used locally and in GitHub Actions
5. **Cross-platform** — Works on macOS and Linux (matches current CI matrix)
6. **Hermetic** — No "works on my machine" — tools don't leak in from the host
7. **Parallel to existing workflow** — `just` recipes continue to work inside the nix shell

### Trade-offs

| Concern                  | Mitigation                                                              |
| ------------------------ | ----------------------------------------------------------------------- |
| Nix learning curve       | `flake.nix` is a single file; justfile still drives daily work          |
| CI cold-start time       | Nix installer is ~10s; dependency caching via `nix-community/setup-nix` |
| IDE integration          | `direnv` + `nix-direnv` auto-loads the shell; LSP tools included        |
| goreleaser compatibility | goreleaser runs inside the dev shell; no changes to `.goreleaser.yml`   |
| Contributor friction     | CONTRIBUTING.md updated with both Nix and non-Nix paths                 |

### Alternatives Considered

| Alternative              | Why Not                                                       |
| ------------------------ | ------------------------------------------------------------- |
| **Makefile only**        | No version pinning; tool installation still manual            |
| **Docker dev container** | Heavy; slow startup; doesn't help CI without Docker-in-Docker |
| **asdf/mise**            | Partial coverage; no hermetic builds; plugin quality varies   |
| **Bazel/Rules Go**       | Overkill for a library; massive migration cost                |
| **Devbox (jetpack.io)**  | Built on Nix but adds a layer; less control; vendor lock-in   |

---

## Proposed Architecture

### File Structure (New Files)

```
go-finding/
├── flake.nix              # Primary flake definition
├── flake.lock             # Pinned dependency hashes (auto-generated)
├── .envrc                 # direnv integration (optional)
└── (existing files unchanged)
```

### High-Level `flake.nix` Design

```
┌─────────────────────────────────────────────────────┐
│  flake.nix                                          │
│                                                     │
│  inputs:                                            │
│  ├── nixpkgs (nixos-unstable for Go 1.26)          │
│  ├── flake-utils (system enumeration)              │
│  └── gitignore (auto-filter src)                   │
│                                                     │
│  outputs:                                           │
│  ├── packages.<system>                              │
│  │   ├── go-finding     (CLI binary)                │
│  │   └── default        → go-finding                │
│  ├── devShells.<system>                             │
│  │   └── default        (go, lint, tools, just)     │
│  ├── checks.<system>                                │
│  │   ├── test           (go test -race)             │
│  │   ├── lint           (golangci-lint)             │
│  │   ├── vet            (go vet)                    │
│  │   └── vuln           (govulncheck)               │
│  ├── apps.<system>                                  │
│  │   ├── lint           → golangci-lint run         │
│  │   ├── vuln           → govulncheck               │
│  │   └── fmt            → goimports + gofumpt       │
│  └── formatter.<system>                             │
│      └── default        (nixpkgs-fmt + golines)     │
└─────────────────────────────────────────────────────┘
```

---

## File-by-File Specification

### 1. `flake.nix`

```nix
{
  description = "go-finding: unified data model and pipeline for static analysis tools";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };

          # Go version matching go.mod
          go = pkgs.go_1_26;

          # Build the Go binary
          go-finding = pkgs.buildGoModule {
            pname = "go-finding";
            version = "0.0.0-unstable"; # overridden by goreleaser in CI

            src = pkgs.lib.cleanSource self;

            vendorHash = ""; # placeholder — run `nix build` to get the correct hash

            nativeBuildInputs = [ go ];

            subPackages = [ "cmd/go-finding" ];

            ldflags = [
              "-s"
              "-w"
            ];

            # Run tests during build
            checkFlags = [ "-race" ];

            meta = {
              description = "Unified data model and pipeline for static analysis tools";
              homepage = "https://github.com/larsartmann/go-finding";
              license = pkgs.lib.licenses.mit;
              mainProgram = "go-finding";
            };
          };

          # Shared development tooling
          devTools = with pkgs; [
            go

            # Linting & analysis
            golangci-lint
            govulncheck

            # Formatting
            gotools          # goimports, guru, etc.
            gofumpt
            golines

            # Build & release
            goreleaser

            # Task runner
            just

            # Nix tooling
            nixpkgs-fmt
          ];

        in
        {
          # --- Packages ---
          packages = {
            default = go-finding;
            inherit go-finding;
          };

          # --- Development Shell ---
          devShells.default = pkgs.mkShell {
            buildInputs = devTools;

            shellHook = ''
              echo "go-finding dev shell"
              echo "  Go:          $(go version)"
              echo "  golangci-lint: $(golangci-lint version --format short 2>/dev/null || echo 'N/A')"
              echo "  just:          $(just --version 2>/dev/null || echo 'N/A')"
              echo ""
              echo "Run 'just' to see available recipes."
            '';
          };

          # --- Checks (nix flake check) ---
          checks = {
            build = go-finding;

            test = pkgs.runCommand "go-finding-test" { }
              ''
                cd ${go-finding.src}
                ${go}/bin/go test -race -count=1 ./...
                touch $out
              '';

            vet = pkgs.runCommand "go-finding-vet" { }
              ''
                cd ${go-finding.src}
                ${go}/bin/go vet ./...
                touch $out
              '';

            lint = pkgs.runCommand "go-finding-lint" { }
              ''
                cd ${go-finding.src}
                ${pkgs.golangci-lint}/bin/golangci-lint run --timeout=5m ./...
                touch $out
              '';
          };

          # --- Apps (nix run .#<name>) ---
          apps = {
            default = {
              type = "app";
              program = "${go-finding}/bin/go-finding";
            };

            lint = {
              type = "app";
              program = "${pkgs.golangci-lint}/bin/golangci-lint";
            };

            vuln = {
              type = "app";
              program = "${pkgs.govulncheck}/bin/govulncheck";
            };
          };

          # --- Formatter (nix fmt) ---
          formatter = pkgs.nixpkgs-fmt;
        }
      );
}
```

#### Key Design Decisions

| Decision                                   | Rationale                                                                           |
| ------------------------------------------ | ----------------------------------------------------------------------------------- |
| `nixos-unstable` for nixpkgs               | Go 1.26 is recent; unstable has it. Pin via `flake.lock` for stability.             |
| `buildGoModule` (not `buildGoApplication`) | Standard approach; `buildGoApplication` requires `gomod2nix` overhead.              |
| `vendorHash = ""`                          | Initial placeholder; first `nix build` fails with the correct hash to paste in.     |
| `flake-utils.lib.eachDefaultSystem`        | Supports `x86_64-linux`, `aarch64-linux`, `x86_64-darwin`, `aarch64-darwin`.        |
| No overlay for golangci-lint               | nixpkgs-unstable ships recent golangci-lint; override only if version drift occurs. |
| `checkFlags = [ "-race" ]`                 | Matches CI and justfile behavior.                                                   |
| `pkgs.lib.cleanSource self`                | Excludes `.git`, `docs/status/`, and other non-source files from the build input.   |
| `shellHook` with version info              | Immediate feedback that the correct tools are available.                            |

### 2. `.envrc` (Optional — for direnv users)

```bash
# .envrc
if ! has nix_direnv_version || ! nix_direnv_version 2.3.0; then
  source_url "https://raw.githubusercontent.com/nix-community/nix-direnv/2.3.0/direnvrc" "sha256-Dmd+j63L84wuzgyjITIfSxSD57Tx7vfcMLlulXO1mWQ="
fi

use flake
```

### 3. Updated `.gitignore` Additions

```gitignore
# Nix
result
result-*
.direnv/
```

### 4. `flake.lock` (Auto-Generated)

Running `nix flake update` generates `flake.lock` with:

- Pinned `nixpkgs` commit hash
- Pinned `flake-utils` commit hash
- Reproducible across all machines

---

## Dependency Inventory

### Runtime Dependencies (Go modules — managed by `go.mod`)

| Dependency           | Version | Notes              |
| -------------------- | ------- | ------------------ |
| `golang.org/x/sync`  | v0.20.0 | errgroup           |
| `golang.org/x/tools` | v0.44.0 | go/analysis        |
| `gopkg.in/yaml.v3`   | v3.0.1  | CLI config parsing |

These are **not** managed by Nix. `buildGoModule` fetches them via `go mod download` during build and hashes the vendor directory.

### Development Dependencies (Managed by Nix)

| Tool            | Package in nixpkgs   | Purpose               |
| --------------- | -------------------- | --------------------- |
| `go_1_26`       | `pkgs.go_1_26`       | Go compiler           |
| `golangci-lint` | `pkgs.golangci-lint` | Linting (85+ linters) |
| `goreleaser`    | `pkgs.goreleaser`    | Release automation    |
| `just`          | `pkgs.just`          | Task runner           |
| `gotools`       | `pkgs.gotools`       | goimports, guru       |
| `gofumpt`       | `pkgs.gofumpt`       | Go formatter          |
| `golines`       | `pkgs.golines`       | Line-length formatter |
| `govulncheck`   | `pkgs.govulncheck`   | Vulnerability scanner |
| `nixpkgs-fmt`   | `pkgs.nixpkgs-fmt`   | Nix file formatter    |

---

## Migration Phases

### Phase 0: Prerequisites (15 minutes)

**Goal:** Ensure Nix is available and the team understands the basics.

- [ ] Install Nix with flakes enabled: https://nixos.org/download
- [ ] Verify: `nix --version` (≥ 2.20)
- [ ] Verify flakes: `nix flake --help`
- [ ] (Optional) Install direnv + nix-direnv for auto-shell loading

```bash
# macOS (determinate systems installer — recommended)
curl --proto '=https' --tlsv1.2 -sSf -L \
  https://install.determinate.systems/nix | sh -s -- install

# Or classic multi-user
sh <(curl -L https://nixos.org/nix/install)

# Enable flakes (if not using Determinate installer)
mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf
```

### Phase 1: Create the Flake (30 minutes)

**Goal:** Add `flake.nix` with a working dev shell.

**Steps:**

1. Create `flake.nix` (see [specification above](#1-flakenix))
2. Initialize the lock file:

```bash
nix flake update
```

3. Enter the dev shell:

```bash
nix develop
```

4. Verify tools are available:

```bash
go version          # go1.26.x
golangci-lint version  # v2.x
just --version      # 1.x.x
govulncheck           # runs
goimports            # runs
```

5. Run the full test suite inside the shell:

```bash
just ci
```

6. Get the correct `vendorHash`:

```bash
# First build will fail with the expected hash
nix build
# Copy the "got: sha256-..." line into vendorHash in flake.nix
# Then rebuild
nix build
```

**Verification:**

- [ ] `nix develop` opens a shell with all tools
- [ ] `just ci` passes inside the nix shell
- [ ] `nix build` produces the `go-finding` binary
- [ ] `./result/bin/go-finding` runs

### Phase 2: CI Integration (1 hour)

**Goal:** Use the flake in GitHub Actions.

**Updated `.github/workflows/ci.yml`:**

```yaml
name: CI

on:
  push:
    branches: [master]
    tags: ["v*"]
  pull_request:
    branches: [master]

permissions:
  contents: read

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4

      - uses: DeterminateSystems/nix-installer-action@v16

      - uses: DeterminateSystems/magic-nix-cache-action@v9

      - name: Run tests
        run: nix develop --command just test

      - name: Run vet
        run: nix develop --command go vet ./...

      - name: Run coverage
        run: |
          nix develop --command bash -c '
            go test -race -count=1 -coverprofile=coverage.out ./...
            COV=$(go tool cover -func=coverage.out | tail -1 | awk "{print \$3}" | sed "s/%//")
            echo "Total coverage: ${COV}%"
            if [ "$(echo "$COV < 75" | bc -l)" -eq 1 ]; then
              echo "::error::Coverage ${COV}% is below 75% threshold"
              exit 1
            fi
          '

      - uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          fail_ci_if_error: false

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: DeterminateSystems/nix-installer-action@v16

      - uses: DeterminateSystems/magic-nix-cache-action@v9

      - name: Run linter
        run: nix develop --command golangci-lint run --timeout=5m ./...

  flake-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: DeterminateSystems/nix-installer-action@v16

      - uses: DeterminateSystems/magic-nix-cache-action@v9

      - name: Check flake
        run: nix flake check --no-build
```

**Updated `.github/workflows/release.yml`:**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: DeterminateSystems/nix-installer-action@v16

      - name: Run tests
        run: nix develop --command just test

      - name: Run vet
        run: nix develop --command go vet ./...

      - name: Build all
        run: nix develop --command go build -v ./...

  release:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: DeterminateSystems/nix-installer-action@v16

      - name: Run GoReleaser
        run: nix develop --command goreleaser release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Verification:**

- [ ] CI passes on push
- [ ] `flake-check` job validates `flake.nix`
- [ ] Release workflow still produces cross-platform binaries
- [ ] Coverage enforcement still works

### Phase 3: Developer Experience (30 minutes)

**Goal:** Make daily development seamless.

**Steps:**

1. Create `.envrc` for direnv integration:

```bash
echo 'if ! has nix_direnv_version || ! nix_direnv_version 2.3.0; then
  source_url "https://raw.githubusercontent.com/nix-community/nix-direnv/2.3.0/direnvrc" "sha256-Dmd+j63L84wuzgyjITIfSxSD57Tx7vfcMLlulXO1mWQ="
fi

use flake' > .envrc
direnv allow
```

2. Update `.gitignore`:

```gitignore
# Nix
result
result-*
.direnv/
```

3. Update `CONTRIBUTING.md` — add Nix path:

````markdown
### Using Nix (recommended)

If you have [Nix with flakes](https://nixos.org/download) installed:

```bash
nix develop      # Enter dev shell with all tools
just ci          # Run all checks
```
````

With [direnv](https://direnv.net/) installed, the shell loads automatically when you `cd` into the project.

````

4. Update `AGENTS.md` — add Nix notes:

```markdown
### Nix

```bash
nix develop            # Enter dev shell
nix build              # Build binary
nix run                # Run binary
nix flake check        # Run all checks
nix fmt                # Format Nix files
nix flake update       # Update all flake inputs
````

````

**Verification:**

- [ ] `direnv allow` loads the shell automatically
- [ ] `just` recipes work in the nix shell
- [ ] `.gitignore` excludes `result` and `.direnv`

### Phase 4: nix flake check Integration (30 minutes)

**Goal:** Make `nix flake check` the single quality gate.

The `checks` defined in `flake.nix` already cover:
- `build` — compilation succeeds
- `test` — `go test -race ./...` passes
- `vet` — `go vet ./...` passes
- `lint` — `golangci-lint run ./...` passes

**Verification:**

```bash
nix flake check
````

All checks should pass. This becomes the canonical "is the project healthy?" command.

### Phase 5: Polish & Harden (1 hour)

**Goal:** Production-readiness.

- [ ] Pin `nixpkgs` to a specific commit in `flake.lock` (not just channel)
- [ ] Verify `vendorHash` is correct and stable
- [ ] Add `CONTRIBUTING.md` section for non-Nix users (keep manual path)
- [ ] Test on fresh macOS machine
- [ ] Test on fresh Linux machine (or CI)
- [ ] Verify goreleaser still produces correct binaries
- [ ] Add `.github/dependabot.yml` or Renovate for `flake.lock` updates (optional)
- [ ] Consider adding a `nix/overlay.nix` if other projects consume go-finding as a Nix dependency

---

## CI/CD Integration

### Strategy: Dual-Path (Transition Period)

During migration, keep both paths working:

```
┌─────────────────────────────────────────────────┐
│  CI Pipeline (Phase 2)                          │
│                                                 │
│  Option A: Nix path (new)                       │
│  ├── nix-installer-action                       │
│  ├── magic-nix-cache-action                     │
│  └── nix develop --command just test            │
│                                                 │
│  Option B: Traditional path (existing)          │
│  ├── actions/setup-go@v5                        │
│  ├── golangci/golangci-lint-action@v6           │
│  └── go test -race ./...                        │
│                                                 │
│  → Converge to Option A once stable             │
└─────────────────────────────────────────────────┘
```

### Caching Strategy

| Layer           | Mechanism                                |
| --------------- | ---------------------------------------- |
| Nix store       | `magic-nix-cache-action` (or `cachix`)   |
| Go module cache | Handled by Nix (inside buildGoModule)    |
| Go build cache  | Not shared across Nix builds (by design) |
| CI runner cache | `magic-nix-cache-action` handles this    |

### CI Timing Estimates

| Step              | Traditional | Nix (cold) | Nix (cached) |
| ----------------- | ----------- | ---------- | ------------ |
| Environment setup | ~30s        | ~60s       | ~15s         |
| Dependencies      | ~20s        | Included   | Included     |
| Test              | ~30s        | ~30s       | ~30s         |
| Lint              | ~60s        | ~60s       | ~60s         |
| **Total**         | ~2m20s      | ~2m30s     | ~1m45s       |

---

## Developer Experience

### Daily Workflow (Post-Migration)

```bash
# Enter dev environment (automatic with direnv)
cd go-finding

# Run checks (same as before)
just test
just lint
just check
just ci

# Nix-specific commands
nix build                  # Build binary → ./result/bin/go-finding
nix run                    # Build and run
nix flake check            # Run all checks
nix develop                # Enter dev shell manually
nix flake update           # Update all dependencies

# goreleaser (unchanged)
goreleaser release --clean
```

### Non-Nix Users

The traditional path remains fully functional. `CONTRIBUTING.md` continues to list manual installation steps. The `justfile` works identically with or without Nix.

### IDE Integration

**VS Code:**

- Install `mkhl.direnv` extension
- The nix shell provides `gopls`, `goimports`, `gofumpt` — LSP works out of the box

**Neovim/Vim:**

- direnv integration via plugin
- LSP tools available in PATH via the nix shell

**GoLand:**

- Set "Go SDK" to the Go from `nix develop` (or let direnv handle it)

---

## Rollback Plan

Nix is **additive** — adding `flake.nix` does not modify or remove any existing files or workflows.

### Rollback Steps (If Needed)

1. Delete `flake.nix`, `flake.lock`, `.envrc`
2. Remove Nix entries from `.gitignore`
3. Revert CI workflow changes
4. Remove Nix sections from `CONTRIBUTING.md` / `AGENTS.md`
5. Commit

**No data loss risk. No build artifact changes. No dependency changes.**

---

## Risks and Mitigations

| Risk                                    | Severity | Likelihood | Mitigation                                                                                                        |
| --------------------------------------- | -------- | ---------- | ----------------------------------------------------------------------------------------------------------------- |
| `go_1_26` not yet in nixpkgs            | High     | Low        | Use `go_1_25` as fallback or overlay with custom Go; `nixos-unstable` typically has latest within days of release |
| golangci-lint v2 not in nixpkgs         | Medium   | Medium     | Override with custom `buildGoModule` derivation (see headscape example)                                           |
| `vendorHash` breaks on `go.mod` changes | Low      | High       | Document: run `nix build`, copy new hash. Or use `vendorHash = pkgs.lib.fakeHash;` during development             |
| CI slower due to Nix overhead           | Low      | Medium     | `magic-nix-cache-action` provides aggressive caching                                                              |
| Contributor refuses to install Nix      | Low      | High       | Traditional path kept; no Nix required to contribute                                                              |
| Build tags (`goexperiment.*`)           | Medium   | Low        | May need `tags = [ "goexperiment.arenas" ... ];` in `buildGoModule`                                               |
| `golines` / `gci` not in nixpkgs        | Low      | Low        | Both are in nixpkgs; if not, custom `buildGoModule` override                                                      |

### Open Questions

1. **Go 1.26 availability in nixpkgs** — Verify `go_1_26` package exists in current nixos-unstable. If not, may need `overlays` to build from source.
2. **golangci-lint v2 compatibility** — nixpkgs may still ship v1.x. May need a custom derivation (see headscape's approach in the reference).
3. **Build tags** — The `.golangci.yml` uses 5 experimental build tags. These may need explicit `buildFlags = [ "-tags" "goexperiment.arenas,goexperiment.jsonv2" ];` in the derivation.
4. **goreleaser in Nix** — goreleaser works fine inside `nix develop`. No changes to `.goreleaser.yml` needed.

---

## Decision Record

### Recommendations

| #   | Decision               | Recommendation                                                                                   |
| --- | ---------------------- | ------------------------------------------------------------------------------------------------ |
| 1   | **Flake inputs**       | `nixpkgs` + `flake-utils` only. No `gomod2nix`, no `dream2nix`. Keep it minimal.                 |
| 2   | **Go module hashing**  | Use `vendorHash` in `buildGoModule`. Update manually on `go.mod` changes.                        |
| 3   | **System support**     | `eachDefaultSystem` — covers x86_64-linux, aarch64-linux, x86_64-darwin, aarch64-darwin.         |
| 4   | **CI migration**       | Full replacement of `setup-go` / `golangci-lint-action` with Nix. Keep as backup during Phase 2. |
| 5   | **direnv**             | Recommended but optional. Documented in Phase 3.                                                 |
| 6   | **Cachix**             | Not needed initially. `magic-nix-cache-action` suffices for CI.                                  |
| 7   | **Justfile**           | Unchanged. `just` is available in the nix shell.                                                 |
| 8   | **goreleaser**         | Unchanged. Runs inside the nix shell.                                                            |
| 9   | **Non-Nix path**       | Always maintained. `CONTRIBUTING.md` documents both paths.                                       |
| 10  | **Flake lock updates** | Manual (`nix flake update`) or via Dependabot/Renovate.                                          |

### Out of Scope

- NixOS module for go-finding (it's a library + CLI, not a service)
- Docker image builds via Nix (goreleaser handles distribution)
- Home Manager integration (user-level, not project-level)
- Binary cache hosting (Cachix) — reconsider if team grows

---

## References

- [Nix Flakes - NixOS Wiki](https://wiki.nixos.org/wiki/Flakes)
- [nix.dev - Getting Started with Flakes](https://nix.dev/concepts/flakes)
- [buildGoModule - nixpkgs manual](https://nixos.org/manual/nixpkgs/unstable/#sec-language-go)
- [flake-utils](https://github.com/numtide/flake-utils)
- [Determinate Systems Nix Installer](https://github.com/DeterminateSystems/nix-installer)
- [magic-nix-cache-action](https://github.com/DeterminateSystems/magic-nix-cache-action)
- [nix-direnv](https://github.com/nix-community/nix-direnv)
- [headscale flake.nix](https://github.com/juanfont/headscale/blob/main/flake.nix) — reference Go project with golangci-lint v2 + Go 1.26 overlay
