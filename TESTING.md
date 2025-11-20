# Testing Guide for PAM

This document explains how to run tests for the PAM package manager refactoring.

## Overview

Tests have been written for all proposed packages **before implementation**. This is Test-Driven Development (TDD) - the tests define the expected behavior, and you'll know your implementation is complete when all tests pass.

## Quick Start

```bash
# Run all tests
./test.sh

# Or use Go directly
go test ./internal/...
```

## Test Structure

```
internal/
  nixconfig/
    editor_test.go      ✅ 13 tests - Nix config parsing & manipulation
  search/
    search_test.go      ✅ 8 tests - Package search & filtering
  assets/
    render_test.go      ✅ 9 tests - Template rendering
  ui/
    forms_test.go       ✅ 11 tests - UI components & helpers
  setup/
    setup_test.go       ✅ 10 tests - Initialization & setup

testdata/
  sample_config.nix     - Sample NixOS configuration
  whitespace_config.nix - Config with various whitespace
  minimal_config.nix    - Minimal config for testing
```

**Total: 51 test cases** covering all proposed functionality.

## Running Tests

### Run All Tests
```bash
./test.sh
```

This script:
- Runs all tests with verbose output
- Generates coverage report
- Creates HTML coverage visualization
- Shows colored pass/fail status
- Saves logs to `/tmp/*_test.log`

### Run Specific Package Tests
```bash
# Test only nixconfig
go test -v ./internal/nixconfig

# Test only search
go test -v ./internal/search

# Test only assets
go test -v ./internal/assets

# Test only UI
go test -v ./internal/ui

# Test only setup
go test -v ./internal/setup
```

### Run Specific Test
```bash
# Run a single test by name
go test -v ./internal/nixconfig -run TestEditor_CategoryExists

# Run all tests matching a pattern
go test -v ./internal/nixconfig -run TestEditor_
```

### Watch Mode (requires entr)
```bash
# Re-run tests when files change
ls internal/**/*.go | entr -c go test ./internal/...
```

### With Coverage
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./internal/...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Benchmarks
```bash
# Run benchmark tests
go test -bench=. ./internal/...

# Benchmark specific package
go test -bench=. ./internal/search
```

## Test Organization

### 1. internal/nixconfig/editor_test.go
Tests the Nix configuration editor for robustness and correctness.

**Key tests:**
- `TestEditor_CategoryExists` - Tests category detection with various whitespace
- `TestEditor_PackageExistsInCategory` - Tests package finding within categories
- `TestEditor_EnablePackage` - Tests enabling disabled packages
- `TestEditor_AddPackageToCategory` - Tests adding new packages
- `TestEditor_CreateCategory` - Tests creating new categories
- `TestEditor_AddOrEnablePackage` - Tests high-level convenience method
- `TestEditor_WithRealConfigFile` - Integration test with actual config
- `TestEditor_WithWhitespaceVariations` - Tests regex parsing flexibility

**Expected API:**
```go
editor := nixconfig.NewEditor(content)
exists := editor.CategoryExists("browsers")
err := editor.AddOrEnablePackage("browsers", "firefox")
content := editor.Content()
```

### 2. internal/search/search_test.go
Tests package searching, filtering, and parsing.

**Key tests:**
- `TestFilterTopLevel` - Tests filtering nested packages
- `TestParseSearchResults` - Tests JSON parsing from nix search
- `TestExtractSystemAndPath` - Tests path parsing logic
- `TestFilterAndPrioritize` - Tests prioritization with showAll flag
- `BenchmarkFilterTopLevel` - Performance benchmark

**Expected API:**
```go
searcher := search.NewSearcher("x86_64-linux")
packages, err := searcher.Search("firefox")
filtered := search.FilterTopLevel(packages)
```

### 3. internal/assets/render_test.go
Tests template rendering with various package types.

**Key tests:**
- `TestRenderMkApp` - Tests rendering for Linux, Darwin, Homebrew
- `TestGetTemplate` - Tests template retrieval
- `TestRenderMkApp_AllPlaceholdersReplaced` - Ensures no leftover placeholders
- `TestRenderMkApp_PreservesTemplateStructure` - Validates Nix syntax
- `TestRenderMkApp_SpecialCharactersInDescription` - Tests edge cases
- `BenchmarkRenderMkApp` - Performance benchmark

**Expected API:**
```go
content := assets.RenderMkApp(pkg, useHomebrew)
template := assets.GetTemplate()
```

### 4. internal/ui/forms_test.go
Tests UI helper functions and directory operations.

**Key tests:**
- `TestGetDirNames` - Tests directory listing
- `TestFormatPackageOption` - Tests option formatting
- `TestBuildPackageOptions` - Tests option builder
- `TestValidateHostSelection` - Tests validation logic
- `TestBuildHostOptions` - Tests host options
- `TestFolderNavigationPath` - Tests path building
- `TestSelectFolder_Integration` - Integration test with filesystem

**Expected API:**
```go
dirs, err := ui.GetDirNames(path)
options := ui.BuildPackageOptions(packages)
label := ui.FormatPackageOption(pkg)
err := ui.ValidateHostSelection(hosts)
```

### 5. internal/setup/setup_test.go
Tests initialization and setup processes.

**Key tests:**
- `TestInitializer_EnsureLibDirectory` - Tests lib dir creation
- `TestInitializer_EnsureMkAppNix` - Tests mkApp.nix creation
- `TestInitializer_Run` - Tests full initialization
- `TestInitializer_InvalidFlakePath` - Tests error handling
- `TestInitializer_PermissionsCorrect` - Tests file permissions
- `TestInitializer_WithExistingFiles` - Tests idempotency
- `BenchmarkInitializer_Run` - Performance benchmark

**Expected API:**
```go
init := setup.NewInitializer(config)
err := init.Run()
err := init.EnsureLibDirectory()
err := init.EnsureMkAppNix()
```

## Implementation Workflow

### Phase 1: Create Package Structure
```bash
mkdir -p internal/{nixconfig,search,assets,ui,setup}
```

### Phase 2: Implement One Package at a Time

For each package (nixconfig, search, assets, ui, setup):

1. **Read the test file** to understand expected behavior
2. **Create the implementation file** (e.g., `editor.go`)
3. **Run tests** for that package:
   ```bash
   go test -v ./internal/nixconfig
   ```
4. **Iterate** until all tests pass
5. **Check coverage**:
   ```bash
   go test -cover ./internal/nixconfig
   ```

### Recommended Order:
1. ✅ **internal/assets** (easiest - template rendering)
2. ✅ **internal/nixconfig** (critical - config manipulation)
3. ✅ **internal/search** (moderate - search logic)
4. ✅ **internal/ui** (moderate - helper functions)
5. ✅ **internal/setup** (uses assets - initialization)

### Phase 3: Refactor cmd/install.go
Once all packages are implemented and tested:
1. Import new packages in `cmd/install.go`
2. Replace inline code with package calls
3. Remove old function definitions
4. Test manually with real Nix setup

## Success Criteria

Your implementation is complete when:

✅ All 51 tests pass
✅ Coverage is >70%
✅ No test failures in `./test.sh`
✅ `go build` succeeds
✅ Manual testing with real Nix config works

## Troubleshooting

### Tests won't compile
```bash
# Check for missing imports
go mod tidy

# Verify package structure
tree internal/
```

### Tests fail on actual implementation
- Read the test carefully - it shows expected behavior
- Use `t.Logf()` in tests to debug
- Check test output for specific assertion failures

### Coverage is low
```bash
# See which lines are uncovered
go test -coverprofile=coverage.out ./internal/nixconfig
go tool cover -html=coverage.out
```

### Integration tests fail
- Check testdata files exist
- Verify file permissions
- Look at `/tmp/*_test.log` for details

## Continuous Testing

### With entr (recommended)
```bash
# Install entr (if not installed)
# macOS: brew install entr
# Linux: apt install entr / pacman -S entr

# Watch and re-run tests
ls internal/**/*.go | entr -c ./test.sh
```

### With Go's built-in watcher (Go 1.23+)
```bash
go test -v ./internal/... -watch
```

## Test Coverage Goals

| Package           | Target Coverage | Tests |
|-------------------|----------------|-------|
| internal/nixconfig | >80%          | 13    |
| internal/search    | >75%          | 8     |
| internal/assets    | >90%          | 9     |
| internal/ui        | >70%          | 11    |
| internal/setup     | >85%          | 10    |
| **Overall**        | **>75%**      | **51**|

## Additional Resources

- [Go Testing Documentation](https://go.dev/doc/tutorial/add-a-test)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Test Coverage](https://go.dev/blog/cover)

## Questions?

If tests are unclear or seem incorrect:
1. Read the test code - it's self-documenting
2. Check the `feats/*.md` files for context
3. Look at similar tests for patterns
4. The tests define the contract - implement to match them

---

**Remember:** Green tests = you're done! 🎉

The tests are your specification. When they all pass, you've successfully refactored the codebase.
