{
  description = "Credit Card Statement Analyzer TUI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        # Allow unfree packages (required for claude-code)
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };

        # Package definition
        statements = pkgs.buildGoModule {
          pname = "statements";
          version = "0.1.0";

          src = ./.;

          vendorHash = "sha256-XijV+d9f4TagHGC2z0zaoZ1Mk1Ax9J2BC6WT/jNZgOA=";

          # Build flags
          ldflags = [
            "-s"
            "-w"
          ];

          meta = with pkgs.lib; {
            description = "Terminal UI for analyzing credit card statements";
            homepage = "https://github.com/BirkhoffLee/credit-card-statement-tui";
            license = licenses.mit;
            maintainers = [ ];
            mainProgram = "statements";
          };
        };

        # Helper function to build the package with specific GOOS and GOARCH
        buildForTarget = goos: goarch:
          pkgs.buildGoModule {
            pname = "statements";
            version = "0.1.0";
            src = ./.;
            vendorHash = "sha256-XijV+d9f4TagHGC2z0zaoZ1Mk1Ax9J2BC6WT/jNZgOA=";
            ldflags = [ "-s" "-w" ];

            # Set Go cross-compilation environment variables
            env = {
              GOOS = goos;
              GOARCH = goarch;
            };
          };

      in
      {
        # Default package
        packages = {
          default = statements;
          statements = statements;

          # Cross-platform builds using Go's built-in cross-compilation
          statements-linux-amd64 = buildForTarget "linux" "amd64";
          statements-linux-arm64 = buildForTarget "linux" "arm64";
          statements-darwin-amd64 = buildForTarget "darwin" "amd64";
          statements-darwin-arm64 = buildForTarget "darwin" "arm64";
          statements-windows-amd64 = buildForTarget "windows" "amd64";
        };

        # Development shell
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go development
            go_1_25
            gopls
            gotools
            go-tools

            # Claude Code
            claude-code

            # Utilities
            git
            jq
          ];

          shellHook = ''
            echo "🚀 Welcome to statements development environment"
            echo ""
            echo "Available commands:"
            echo "  go build -o statements       - Build the application"
            echo "  go run .                     - Run the application directly"
            echo "  nix build                    - Build with Nix"
            echo "  nix run                      - Run with Nix"
            echo ""
            echo "📦 Go version: $(go version)"
            echo "🤖 Claude Code: $(claude --version)"
            echo ""

            # Set up Go environment
            export GOPATH="$HOME/go"
            export PATH="$GOPATH/bin:$PATH"
          '';
        };

        # App for easy running
        apps.default = {
          type = "app";
          program = "${statements}/bin/statements";
        };
      }
    );
}
