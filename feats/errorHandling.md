# Error Handling Improvement

## Status
Proposed

## Problem
Currently, errors are handled inconsistently throughout the codebase:
- Using `fmt.Errorf()` directly everywhere
- Printing errors with `fmt.Println()` and returning
- No structured error types
- Hard to distinguish between different error categories
- Difficult to test error conditions

## Proposed Solution
Create an `internal/errors` package with custom error types:

### Custom Error Types
1. **SearchError** - Package search failures
2. **TemplateError** - Template rendering issues
3. **NixConfigError** - Configuration manipulation failures
4. **SetupError** - Initialization problems

### Sentinel Errors
- `ErrNixSearchFailed`
- `ErrNoPackagesFound`
- `ErrTemplateNotFound`
- `ErrConfigNotFound`
- `ErrInvalidNixConfig`

### Benefits
- Better error messages with context
- Testable error conditions
- Consistent error handling patterns
- Easier to debug issues
- Can use `errors.Is()` and `errors.As()` for error checking

## Implementation Priority
Low - Can be added after core refactoring is complete

## Example Usage
```go
// Instead of:
if err != nil {
    fmt.Println("Error: ", err)
    return
}

// Use:
if err != nil {
    return &errors.SearchError{
        PackageName: packageName,
        System: targetSystem,
        Err: err,
    }
}
```
