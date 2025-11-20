# Improved Nix Configuration Parsing

## Status
Proposed

## Problem
Current Nix configuration manipulation uses fragile string matching that's prone to breakage:

```go
// Current approach - breaks with whitespace variations
func categoryExists(content string, category string) bool {
    pattern := category + " = {"
    return strings.Contains(content, pattern)
}
```

Issues:
- **Whitespace sensitive**: Fails if user has `category={` or `category  =  {`
- **No error handling**: Silent failures or incorrect modifications
- **Hard to test**: String manipulation logic spread across multiple functions
- **Comment blind**: Doesn't handle commented-out sections
- **Fragile updates**: Easy to corrupt configuration files

Example failures:
```nix
# This works:
apps = {

# These don't:
apps={
apps  =  {
apps = { # comment
```

## Proposed Solution
Create an `internal/nixconfig` package with regex-based parsing for robustness.

### Package Structure
```
internal/
  nixconfig/
    editor.go           # Configuration editor with regex patterns
    editor_test.go      # Comprehensive tests
```

### Key Improvements

#### 1. Editor Pattern
Encapsulate configuration content in an `Editor` struct:
```go
type Editor struct {
    content string
}

func NewEditor(content string) *Editor
func (e *Editor) Content() string
func (e *Editor) AddOrEnablePackage(category, pkg string) error
```

#### 2. Regex-Based Parsing
Use regex for flexible matching:
```go
// Handles whitespace variations
pattern := regexp.QuoteMeta(category) + `\s*=\s*\{`
```

#### 3. Better API
```go
// Instead of multiple functions with confusing logic:
if categoryExists(content, category) {
    if packageExistsInCategory(content, category, pkg) {
        content = enablePackage(content, pkg)
    } else {
        content = addPackageToCategory(content, category, pkg)
    }
} else {
    content = createCategory(content, category, pkg)
}

// Use simple, clear API:
editor := nixconfig.NewEditor(content)
err := editor.AddOrEnablePackage(category, pkg)
updatedContent := editor.Content()
```

### Methods to Implement

1. **CategoryExists(category string) bool**
   - Check if category section exists
   - Regex handles whitespace

2. **PackageExistsInCategory(category, pkg string) bool**
   - Search within category bounds
   - Return true if package is defined

3. **EnablePackage(pkg string) bool**
   - Change `pkg.enable = false` → `pkg.enable = true`
   - Return whether change was made

4. **AddPackageToCategory(category, pkg string) error**
   - Add new package to existing category
   - Return error if category doesn't exist

5. **CreateCategory(category, pkg string) error**
   - Create new category with package
   - Return error if apps section missing

6. **AddOrEnablePackage(category, pkg string) error**
   - High-level method combining all logic
   - Automatically decides what to do

### Testing Strategy
```go
func TestEditor_CategoryExists(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        category string
        want     bool
    }{
        {"standard format", "apps = {", "apps", true},
        {"no spaces", "apps={", "apps", true},
        {"extra spaces", "apps  =  {", "apps", true},
        {"not found", "other = {", "apps", false},
    }
    // ...
}
```

## Benefits
1. **Robust**: Handles whitespace variations
2. **Testable**: Pure functions, easy to unit test
3. **Maintainable**: All config logic in one place
4. **Error handling**: Return errors instead of silent failures
5. **Reusable**: Can be used for other config operations
6. **Type safe**: Editor pattern prevents accidental string corruption

## Limitations
Still not a full Nix parser (would need AST parsing for that), but good enough for:
- Standard configuration.nix patterns
- Simple enable/disable operations
- Adding new packages

Won't handle:
- Complex Nix expressions
- Multi-line strings
- Nested attribute sets (beyond basic category structure)
- Comments (might break with comments in wrong places)

## Future Enhancements
- Use a proper Nix parser/AST library
- Support more complex configuration patterns
- Validate Nix syntax after modifications
- Dry-run mode to preview changes

## Implementation Priority
High - Critical for reliability and maintainability

## Migration Path
1. Create `internal/nixconfig` package with `Editor`
2. Write comprehensive tests
3. Update `cmd/install.go` to use new package
4. Remove old string manipulation functions
5. Add integration tests with real config files

## Related Issues
- Currently no error messages if config manipulation fails
- Hard to debug when things go wrong
- Users might have different formatting styles
