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
      vendorHash = "sha256-Ee5TL0UpUWExhSX5YHQ4AV9k+Ss7s09bzaC7NCUcyPU=";
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
          meta = {
            description = "Code quality finding framework for Go";
            homepage = "https://github.com/LarsArtmann/go-finding";
            license = lib.licenses.mit;
            maintainers = [ lib.maintainers.larsartmann ];
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
              license = pkgs.lib.licenses.mit;
              platforms = pkgs.lib.platforms.unix;
              maintainers = [ pkgs.lib.maintainers.larsartmann ];
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
          };

          checks = {
            format = config.treefmt.build.check self;
            build = config.packages.default;
          };

          apps = {
            test = mkApp "test" "Run all tests" ''
              go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" "Run all tests with race detector" ''
              go test ./... -race -count=1 "$@"
            '';

            bench = mkApp "bench" "Run benchmarks" ''
              go test ./... -bench=. -benchmem "$@"
            '';

            build = mkApp "build" "Build all packages" ''
              go build ./...
            '';

            vet = mkApp "vet" "Run go vet" ''
              go vet ./...
            '';

            lint = mkApp "lint" "Run golangci-lint" ''
              golangci-lint run ./...
            '';

            coverage = mkApp "coverage" "Run tests with coverage report" ''
              go test ./... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';

            art-dupl = mkApp "art-dupl" "Check code duplication with art-dupl (requires art-dupl in PATH)" ''
              if ! command -v art-dupl &>/dev/null; then
                echo "art-dupl not found. Install: go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest" >&2
                exit 1
              fi
              art-dupl . -t 50 "$@"
            '';

            clean = mkApp "clean" "Clean build and test artifacts" ''
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
