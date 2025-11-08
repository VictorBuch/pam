{
  pkgs,
  lib,
  config,
  ...
}:

{
  # Go Language Configuration
  languages.go = {
    enable = true;
    package = pkgs.go; # Latest stable Go version
  };

  # Development Tools
  packages = with pkgs; [
    # Core Go tools
    gotools # goimports, guru, godoc, etc.
    gopls # Go Language Server
    delve # Go debugger (dlv)
    golangci-lint # Comprehensive Go linter

    # Database tooling
    sqlc # Generate type-safe Go from SQL

    # Utilities
    jq # JSON processing
  ];

  # Environment Variables
  env = {
    # Add custom env vars here if needed
  };

  # Pre-commit Hooks
  pre-commit.hooks = {
    gofmt = {
      enable = true;
      name = "gofmt";
      entry = "${pkgs.go}/bin/gofmt -l -w";
      files = "\\.go$";
      language = "system";
    };
  };

  # Custom Scripts
  scripts = {
    dev.exec = ''
      if [ -f "main.go" ]; then
        echo "Running go run ."
        go run .
      elif [ -f "cmd/main.go" ]; then
        echo "Running go run ./cmd"
        go run ./cmd
      else
        echo "No main.go found. Please specify your entry point:"
        echo "  go run <path>"
      fi
    '';

    build.exec = ''
      mkdir -p bin
      if [ -f "main.go" ]; then
        go build -o bin/app .
      elif [ -f "cmd/main.go" ]; then
        go build -o bin/app ./cmd
      else
        echo "No main.go found. Building all packages:"
        go build ./...
      fi
      echo "Build complete!"
    '';

    test.exec = ''
      echo "Running tests..."
      go test ./... -v
    '';

    test-coverage.exec = ''
      echo "Running tests with coverage..."
      go test ./... -coverprofile=coverage.out
      go tool cover -html=coverage.out -o coverage.html
      echo "Coverage report generated: coverage.html"
    '';

    lint.exec = ''
      echo "Running golangci-lint..."
      golangci-lint run ./...
    '';

    mod-tidy.exec = ''
      echo "Tidying go.mod..."
      go mod tidy
      echo "Done!"
    '';

    mod-init.exec = ''
      if [ -f "go.mod" ]; then
        echo "go.mod already exists!"
      else
        echo "Enter module name (e.g., github.com/user/project):"
        read module_name
        go mod init "$module_name"
        echo "Initialized go.mod with module: $module_name"
      fi
    '';
  };

  # Optional: PostgreSQL Database Service
  # Uncomment to enable PostgreSQL for your project
  # services.postgres = {
  #   enable = true;
  #   package = pkgs.postgresql_16;
  #   initialDatabases = [{ name = "go_db"; }];
  #   listen_addresses = "127.0.0.1";
  #   port = 5432;
  # };

  # Welcome Message
  enterShell = ''
    echo ""
    echo "🚀 Go Development Environment"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Go version:     $(go version | cut -d' ' -f3)"
    echo "gopls version:  $(gopls version | head -n1)"
    echo ""
    echo "📦 Available Tools:"
    echo "  - gopls (LSP)"
    echo "  - delve (debugger)"
    echo "  - golangci-lint"
    echo "  - sqlc"
    echo ""
    echo "🛠️  Custom Scripts:"
    echo "  dev              - Run your Go application"
    echo "  build            - Build binary to bin/app"
    echo "  test             - Run all tests"
    echo "  test-coverage    - Generate coverage report"
    echo "  lint             - Run golangci-lint"
    echo "  mod-tidy         - Tidy go.mod dependencies"
    echo "  mod-init         - Initialize new Go module"
    echo ""
    echo "💡 Tip: gofmt runs automatically on commit"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
  '';
}
