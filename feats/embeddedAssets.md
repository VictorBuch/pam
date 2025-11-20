# Embedded Assets / Template System

## Status
Proposed

## Problem
Currently facing the hardcoded path issue (FIXME on line 21-23 in cmd/install.go):
```go
const (
    NIX_PKGS_MODULE_TEMPLATE = "./mkApp.txt"  // Won't work when distributed!
)
```

When the binary is distributed, `mkApp.txt` won't be available at this relative path, causing the program to fail.

## Proposed Solution
Use Go 1.16+ `embed` package to embed template files directly into the binary.

### Package Naming Options
Since there will likely be multiple templates/assets, consider these names:

1. **`internal/assets`** ✅ RECOMMENDED
   - Generic enough for all embedded files
   - Can contain templates, default configs, etc.
   - Clear purpose

2. **`internal/templates`** (plural)
   - If we only embed templates
   - More specific but limiting

3. **`internal/embed`**
   - Very clear what it does
   - Might be confused with the Go embed package

### Implementation Structure

```go
// internal/assets/assets.go
package assets

import _ "embed"

// Nix module templates
//go:embed templates/mkApp.nix
var MkAppTemplate string

//go:embed templates/mkModule.nix
var MkModuleTemplate string

// Future: could add more assets
//go:embed configs/default.yaml
var DefaultConfig string

// RenderMkApp renders the mkApp template with package data
func RenderMkApp(pkg *types.Package, useHomebrew bool) string {
    // Template rendering logic here
}
```

### Directory Structure
```
internal/
  assets/
    assets.go           # Main embed file
    render.go           # Template rendering functions
    render_test.go      # Tests
    templates/
      mkApp.nix         # Moved from root mkApp.txt
      mkModule.nix      # Future templates
    configs/
      default.yaml      # Future embedded configs
```

### Benefits
1. **Distribution ready** - Works anywhere, no external file dependencies
2. **Version controlled** - Templates are versioned with code
3. **Atomic deploys** - Binary contains everything it needs
4. **Faster startup** - No disk I/O for template loading
5. **Extensible** - Easy to add more embedded assets

### Migration Path
1. Create `internal/assets` package
2. Move `mkApp.txt` → `internal/assets/templates/mkApp.nix`
3. Embed using `//go:embed`
4. Update `initialSetup()` to use embedded content
5. Update `install()` to use template renderer
6. Keep old `mkApp.txt` in root for backwards compatibility (can remove later)

### Example Usage
```go
// Before:
data, err := os.ReadFile(NIX_PKGS_MODULE_TEMPLATE)
if err != nil {
    fmt.Println("Error reading template file: ", err)
    return
}

// After:
moduleContent := assets.RenderMkApp(selectedPkg, installWithBrew)
```

## Implementation Priority
High - Fixes critical FIXME that prevents distribution

## Related Issues
- FIXME comment in cmd/install.go:21
- Makes the project distribution-ready
- Enables future template features (multiple module types, etc.)

## Additional Considerations
- Could add template validation at compile time
- Could support template variables/functions with `text/template` package
- Could cache rendered templates if performance becomes an issue
