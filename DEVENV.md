# Go Development Environment

A complete Go development environment using devenv.sh with essential tooling for building Go applications, APIs, and CLI tools.

## What's Included

### Go Toolchain
- **Go**: Latest stable version from nixpkgs-unstable
- **gopls**: Official Go Language Server for IDE features
- **delve**: Go debugger (`dlv`)
- **golangci-lint**: Comprehensive linter with multiple analyzers

### Development Tools
- **gotools**: Collection of Go tools (goimports, guru, godoc, etc.)
- **gomodifytags**: Add, update, or remove struct tags
- **impl**: Generate method stubs for interfaces
- **sqlc**: Generate type-safe Go code from SQL
- **jq**: JSON processor for API testing

### Git Hooks
- **gofmt**: Automatically format Go code on commit

### Custom Scripts
- `dev` - Run your Go application (auto-detects entry point)
- `build` - Build binary to `bin/app`
- `test` - Run all tests with verbose output
- `test-coverage` - Generate HTML coverage report
- `lint` - Run golangci-lint on entire project
- `mod-tidy` - Clean up go.mod dependencies
- `mod-init` - Initialize new Go module interactively

## Quick Start

### For New Projects

```bash
# 1. Copy this template to your project directory
cp -r ~/serenityOs/devenvs/go/* my-go-project/
cd my-go-project

# 2. Allow direnv to activate the environment
direnv allow

# 3. Initialize your Go module
mod-init

# 4. Create your main.go
cat > main.go <<EOF
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
EOF

# 5. Run your app
dev
```

### For Existing Projects

```bash
# 1. Copy template files to your project root
cd your-existing-project
cp ~/serenityOs/devenvs/go/devenv.nix .
cp ~/serenityOs/devenvs/go/devenv.yaml .
cp ~/serenityOs/devenvs/go/.envrc .
cp ~/serenityOs/devenvs/go/.gitignore .

# 2. Activate the environment
direnv allow

# 3. Install dependencies
go mod download

# 4. Run your app
dev
```

## Available Scripts

All scripts are available as commands once the environment is activated:

### Development
```bash
dev                 # Run your application (detects main.go or cmd/main.go)
build               # Build binary to bin/app
```

### Testing
```bash
test                # Run all tests
test-coverage       # Generate coverage report (coverage.html)
```

### Code Quality
```bash
lint                # Run golangci-lint
# Note: gofmt runs automatically on git commit
```

### Module Management
```bash
mod-tidy            # Clean up go.mod
mod-init            # Initialize new Go module (interactive)
```

## Works With

This environment supports all Go projects, including popular frameworks:

### Web Frameworks
- **Gin**: `go get -u github.com/gin-gonic/gin`
- **Echo**: `go get -u github.com/labstack/echo/v4`
- **Fiber**: `go get -u github.com/gofiber/fiber/v2`
- **Chi**: `go get -u github.com/go-chi/chi/v5`
- **Gorilla Mux**: `go get -u github.com/gorilla/mux`

### CLI Tools
- **Cobra**: `go get -u github.com/spf13/cobra`
- **Viper**: `go get -u github.com/spf13/viper`
- **Bubble Tea**: `go get github.com/charmbracelet/bubbletea`

### Database Libraries
- **GORM**: `go get -u gorm.io/gorm`
- **sqlx**: `go get -u github.com/jmoiron/sqlx`
- **pgx**: `go get -u github.com/jackc/pgx/v5`

## PostgreSQL Database Support

This template includes optional PostgreSQL support. To enable:

1. Edit `devenv.nix` and uncomment the PostgreSQL section:

```nix
services.postgres = {
  enable = true;
  package = pkgs.postgresql_16;
  initialDatabases = [{ name = "mydb"; }];
  listen_addresses = "127.0.0.1";
  port = 5432;
};
```

2. Reload the environment:
```bash
direnv reload
```

3. Start PostgreSQL:
```bash
devenv up
```

The database will be available at `postgresql://localhost:5432/mydb`. Data persists in `.devenv/state/postgres/`.

### Using sqlc

For type-safe database access, use sqlc:

1. Create `sqlc.yaml`:
```yaml
version: "2"
sql:
  - schema: "schema.sql"
    queries: "queries.sql"
    engine: "postgresql"
    gen:
      go:
        package: "db"
        out: "internal/db"
```

2. Generate code:
```bash
sqlc generate
```

## Customization

### Adding More Tools

Edit `devenv.nix` and add packages to the `packages` list:

```nix
packages = with pkgs; [
  # ... existing packages ...
  go-migrate      # Database migrations
  air            # Live reload
  goose          # Another migration tool
];
```

### Changing Go Version

To pin a specific Go version, edit `devenv.nix`:

```nix
languages.go = {
  enable = true;
  package = pkgs.go_1_22;  # Pin to Go 1.22
};
```

Available versions: `go` (latest), `go_1_23`, `go_1_22`, `go_1_21`

### Custom Environment Variables

Add environment variables in `devenv.nix`:

```nix
env = {
  DATABASE_URL = "postgresql://localhost:5432/mydb";
  API_KEY = "your-api-key";
  LOG_LEVEL = "debug";
};
```

### Additional Git Hooks

Add more pre-commit hooks in `devenv.nix`:

```nix
pre-commit.hooks = {
  gofmt = { /* ... existing ... */ };

  golangci-lint = {
    enable = true;
    name = "golangci-lint";
    entry = "${pkgs.golangci-lint}/bin/golangci-lint run --fix";
    files = "\\.go$";
    language = "system";
  };
};
```

## Common Workflows

### API Development with Gin

```bash
# Initialize project
mod-init

# Install Gin
go get -u github.com/gin-gonic/gin

# Create main.go
cat > main.go <<EOF
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    r.Run()
}
EOF

# Run with hot reload (add air if needed)
dev
```

### CLI Tool with Cobra

```bash
# Initialize project
mod-init

# Install Cobra
go get -u github.com/spf13/cobra

# Initialize Cobra structure
go run github.com/spf13/cobra-cli@latest init

# Build and test
build
./bin/app
```

### Database-Backed API

```bash
# Enable PostgreSQL in devenv.nix
# Start database
devenv up

# In another terminal
# Create schema and queries
# Run sqlc generate
sqlc generate

# Develop your API
dev
```

## Troubleshooting

### Environment Not Activating

```bash
# Ensure direnv is installed and hooked into your shell
direnv --version

# Check direnv status
direnv status

# Re-allow the directory
direnv allow
```

### Go Version Issues

```bash
# Verify Go version
go version

# If wrong version, update devenv.nix and reload
direnv reload
```

### Database Connection Issues

```bash
# Check if PostgreSQL is running
devenv processes

# View database logs
cat .devenv/state/postgres/postgresql.log

# Restart database
devenv down
devenv up
```

### gopls Not Working in Editor

```bash
# Verify gopls is available
gopls version

# Rebuild gopls cache
gopls workspace_configuration

# Check your editor's Go plugin configuration
```

## Learn More

- [devenv.sh Documentation](https://devenv.sh)
- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [golangci-lint](https://golangci-lint.run/)
- [sqlc](https://sqlc.dev/)
