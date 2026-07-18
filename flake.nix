{
  description = "go-finding — Code quality finding framework for Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      treefmt-nix,
      systems,
    }:
    let
      inherit (nixpkgs) lib;

      version = self.rev or self.dirtyRev or "dev";
      vendorHash = "sha256-F6l+hilDt4vBSdW7QNRQnOy3yiQU7fWvwTsPLspA0Kw=";
      proxyVendor = true;

      goSrc = lib.fileset.toSource {
        root = ./.;
        fileset = lib.fileset.gitTracked ./.;
      };

      mkGoFinding =
        buildGoModule:
        buildGoModule {
          pname = "go-finding";
          inherit version vendorHash proxyVendor;
          src = goSrc;
          # Multi-module: build from cmd/go-finding module.
          # Remove go.work so buildGoModule uses replace directives (GOWORK=off).
          postPatch = "rm -f go.work";
          modRoot = "cmd/go-finding";
          subPackages = [ "." ];
          ldflags = [
            "-s"
            "-w"
          ];
          env = {
            GOEXPERIMENT = "jsonv2";
          };
          meta = {
            description = "Code quality finding framework for Go";
            homepage = "https://github.com/LarsArtmann/go-finding";
            license = lib.licenses.mit;
            maintainers = [
              {
                name = "Lars Artmann";
                github = "LarsArtmann";
              }
            ];
            mainProgram = "go-finding";
          };
        };
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          ...
        }:
        let
          goPkg = pkgs.go_1_26;

          mkApp = name: description: script: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name;
                runtimeInputs = [
                  goPkg
                  pkgs.golangci-lint
                  pkgs.trash-cli
                ];
                text = script;
              }
            }/bin/${name}";
            meta = {
              description = "Unified data model and pipeline for static analysis tools";
              mainProgram = name;
              homepage = "https://github.com/larsartmann/go-finding";
              license = lib.licenses.mit;
              platforms = lib.platforms.unix;
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
            };
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              golines = {
                enable = true;
                maxLength = 120;
              };
              nixfmt.enable = true;
            };
          };

          packages.default = mkGoFinding pkgs.buildGoModule;

          devShells.default = pkgs.mkShell {
            packages = [
              goPkg
              pkgs.golangci-lint
              pkgs.gofumpt
              pkgs.golines
              pkgs.gopls
              pkgs.gotools
              pkgs.trash-cli
            ];

            env = {
              GOEXPERIMENT = "jsonv2";
            };

            shellHook = ''
              echo "go-finding dev shell — $(go version)"
              echo "Multi-module workspace active (go.work)"
            '';
          };

          devShells.ci = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
            ];

            env = {
              GOEXPERIMENT = "jsonv2";
            };
          };

          checks = {
            format = config.treefmt.build.check self;
            build = config.packages.default;
          };

          apps = {
            test = mkApp "test" "Run all tests" ''
              export GOEXPERIMENT=jsonv2
              go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" "Run all tests with race detector" ''
              export GOEXPERIMENT=jsonv2
              go test ./... -race -count=1 "$@"
            '';

            bench = mkApp "bench" "Run benchmarks" ''
              export GOEXPERIMENT=jsonv2
              go test ./... -bench=. -benchmem "$@"
            '';

            build = mkApp "build" "Build all packages" ''
              export GOEXPERIMENT=jsonv2
              go build ./...
            '';

            vet = mkApp "vet" "Run go vet" ''
              export GOEXPERIMENT=jsonv2
              go vet ./...
            '';

            lint = mkApp "lint" "Run golangci-lint" ''
              export GOEXPERIMENT=jsonv2
              golangci-lint run ./...
            '';

            coverage = mkApp "coverage" "Run tests with coverage report" ''
              export GOEXPERIMENT=jsonv2
              go test ./... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';

            art-dupl = mkApp "art-dupl" "Check code duplication with art-dupl (requires art-dupl in PATH)" ''
              export GOEXPERIMENT=jsonv2
              if ! command -v art-dupl &>/dev/null; then
                echo "art-dupl not found. Install: go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest" >&2
                exit 1
              fi
              art-dupl . -t 50 "$@"
            '';

            clean = mkApp "clean" "Clean build and test artifacts" ''
              export GOEXPERIMENT=jsonv2
              trash-put coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };

      flake.overlays.default = final: _prev: {
        go-finding = mkGoFinding final.buildGoModule;
      };
    };
}
