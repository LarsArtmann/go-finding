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
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          system,
          ...
        }:
        let
          goPkg = pkgs.go_1_26;

          version = self.rev or self.dirtyRev or "dev";

          src = pkgs.lib.fileset.toSource {
            root = ./.;
            fileset = pkgs.lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./doc.go
              ./version.go
              ./category.go
              ./confidence.go
              ./diff.go
              ./errors.go
              ./filter.go
              ./fix_strategy.go
              ./format.go
              ./id.go
              ./json.go
              ./lsp.go
              ./merge.go
              ./position.go
              ./report.go
              ./sarif_export.go
              ./sarif_import.go
              ./sarif_types.go
              ./severity.go
              ./suppression.go
              ./tag.go
              ./analysis
              ./cmd
              ./internal
              ./pipeline
            ];
          };

          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          mkApp = name: description: script: {
            type = "app";
            program = "${pkgs.writeShellApplication {
              inherit name;
              runtimeInputs = [ goPkg pkgs.golangci-lint pkgs.trash-cli ];
              text = script;
            }}/bin/${name}";
            meta = { inherit description; };
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              golines.enable = true;
              nixfmt.enable = true;
            };
          };

          packages.default = pkgs.buildGoModule {
            pname = "go-finding";
            inherit version src vendorHash;
            ldflags = [
              "-s"
              "-w"
            ];
            meta = with pkgs.lib; {
              description = "Code quality finding framework for Go";
              homepage = "https://github.com/LarsArtmann/go-finding";
              license = licenses.mit;
              mainProgram = "go-finding";
            };
          };

          devShells.default = pkgs.mkShell {
            packages = [
              goPkg
              pkgs.golangci-lint
              pkgs.gofumpt
              pkgs.golines
              pkgs.gotools
              pkgs.trash-cli
            ];

            GOWORK = "off";

            shellHook = ''
              echo "go-finding dev shell — $(go version)"
            '';
          };

          checks = {
            build = config.packages.default;
            test = config.packages.default.overrideAttrs (_: { doCheck = true; });
          };

          apps = {
            test = mkApp "test" "Run all tests" ''
              go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" "Run all tests with race detector" ''
              go test ./... -race -count=1 "$@"
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

            clean = mkApp "clean" "Clean build and test artifacts" ''
              trash-put coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };

      flake.overlays.default = final: prev: {
        go-finding = final.callPackage ./package.nix { };
      };
    };
}
