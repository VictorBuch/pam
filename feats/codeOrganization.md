# Code Organization & Package Extraction

## Status
Proposed

## Problem
The `cmd/install.go` file has grown to 418 lines and violates Single Responsibility Principle:

**Current Responsibilities (all in one file):**
1. Package searching via Nix CLI
2. Filtering and prioritizing search results
3. Nix configuration file parsing/manipulation
4. Interactive UI forms and prompts
5. File system operations
6. Template rendering
7. Initial setup/bootstrapping
8. Command orchestration

**Issues:**
- Hard to test individual components
- Difficult to find specific functionality
- Can't reuse logic in other commands
- High coupling between concerns
- 145-line `install()` function doing too much
- Global state (FLAKE_ROOT, NIX_APPS_DIR, etc.)

## Proposed Solution
Extract logical components into focused internal packages.

### Package Extraction Plan

#### 1. `internal/search` - Package Search
**Purpose**: Handle Nix package searching and filtering

**Functions to extract from install.go:**
- `searchPackages()` (lines 34-53)
- `filterAndPrioritizePackages()` (lines 55-78)
- `SearchResult` type

**New API:**
```go
package search

type Searcher struct {
    system string
}

func NewSearcher(system string) *Searcher
func (s *Searcher) Search(packageName string) ([]types.Package, error)
func FilterTopLevel(packages []types.Package) []types.Package
```

**Benefits:**
- Can mock for testing
- Reusable for future commands (list, update, etc.)
- Isolated from UI concerns

---

#### 2. `internal/ui` - Interactive UI Components
**Purpose**: Reusable UI forms and prompts

**Functions to extract:**
- `selectFolderRecursively()` (lines 157-201)
- `getDirNames()` (lines 143-155)
- Package selection form logic (lines 310-320)
- Host selection form logic (lines 328-345)

**New API:**
```go
package ui

func SelectPackage(packages []types.Package) (*types.Package, error)
func SelectHosts(hostDir string) ([]string, error)
func SelectFolder(rootPath string) (string, error)
func ConfirmAction(message string) (bool, error)
```

**Benefits:**
- Consistent UI patterns across commands
- Easier to test UI logic
- Can switch to different UI library if needed

---

#### 3. `internal/setup` - Initialization
**Purpose**: Handle initial setup and bootstrapping

**Functions to extract:**
- `initialSetup()` (lines 221-258)
- mkApp.nix creation logic

**New API:**
```go
package setup

type Initializer struct {
    config *internal.Config
}

func NewInitializer(cfg *internal.Config) *Initializer
func (i *Initializer) EnsureLibDirectory() error
func (i *Initializer) EnsureMkAppNix() error
func (i *Initializer) Run() error
```

**Benefits:**
- Can expand setup logic without cluttering install.go
- Testable setup steps
- Future: could add setup verification

---

### Refactored `cmd/install.go` Structure

**Before:** 418 lines with everything mixed

**After:** ~150 lines, focused on orchestration

```go
package cmd

import (
    "pam/internal/search"
    "pam/internal/ui"
    "pam/internal/nixconfig"
    "pam/internal/assets"
    "pam/internal/setup"
)

func install(cmd *cobra.Command, args []string) {
    // 1. Setup
    setupInit := setup.NewInitializer(config)
    if err := setupInit.Run(); err != nil {
        handleError(err)
        return
    }

    // 2. Search
    searcher := search.NewSearcher(targetSystem)
    packages, err := searcher.Search(args[0])
    if err != nil {
        handleError(err)
        return
    }

    // 3. User selections
    selectedPkg, err := ui.SelectPackage(packages)
    if err != nil {
        handleError(err)
        return
    }

    selectedFolder, err := ui.SelectFolder(config.DefaultModuleDir)
    if err != nil {
        handleError(err)
        return
    }

    selectedHosts, err := ui.SelectHosts(config.DefaultHostDir)
    if err != nil {
        handleError(err)
        return
    }

    // 4. Create module
    if err := createModuleFile(selectedPkg, selectedFolder, args[0]); err != nil {
        handleError(err)
        return
    }

    // 5. Update host configurations
    if err := updateHostConfigs(selectedHosts, selectedFolder, selectedPkg); err != nil {
        handleError(err)
        return
    }

    // 6. Optional: open editor
    if confirmEdit, _ := ui.ConfirmAction("Edit module?"); confirmEdit {
        openEditor(moduleFilePath)
    }
}
```

**Result:** Clear flow, easy to understand, testable components

---

### Global Variables → Context Pattern

**Problem:**
```go
var FLAKE_ROOT, NIX_APPS_DIR, NIX_HOSTS_DIR string
```

**Solution:** Pass config through context or parameters
```go
type InstallContext struct {
    Config      *internal.Config
    FlakeRoot   string
    AppsDir     string
    HostsDir    string
}

func newInstallContext() (*InstallContext, error) {
    cfg, err := internal.LoadConfig()
    if err != nil {
        return nil, err
    }

    return &InstallContext{
        Config:    cfg,
        FlakeRoot: cfg.FlakePath,
        AppsDir:   filepath.Join(cfg.FlakePath, cfg.DefaultModuleDir),
        HostsDir:  filepath.Join(cfg.FlakePath, cfg.DefaultHostDir),
    }, nil
}
```

---

## File Structure After Refactoring

```
cmd/
  install.go          (~150 lines, orchestration only)
  root.go             (existing)

internal/
  config.go           (existing)

  nixconfig/          (see nixconfigParsing.md)
    editor.go
    editor_test.go

  search/             (NEW)
    search.go         - Package searching & filtering
    search_test.go

  ui/                 (NEW)
    forms.go          - Interactive UI components
    forms_test.go

  setup/              (NEW)
    setup.go          - Initial setup logic
    setup_test.go

  assets/             (see embeddedAssets.md)
    assets.go
    render.go
    render_test.go

  types/
    package.go        (existing)
```

---

## Benefits

### Maintainability
- Each package has single, clear responsibility
- Easy to locate specific functionality
- Changes isolated to relevant package

### Testability
- Can test each package independently
- Mock external dependencies (nix CLI, file system)
- Table-driven tests for each component

### Reusability
- UI components reusable in other commands
- Search logic available for `list`, `update` commands
- Config editor useful for `remove`, `disable` commands

### Onboarding
- New contributors can understand each package quickly
- Clear separation of concerns
- Self-documenting structure

### Future Features
With this structure, adding new commands is easy:
```go
// Future: cmd/remove.go
func remove(cmd *cobra.Command, args []string) {
    editor := nixconfig.NewEditor(content)
    ui.SelectHosts()  // Reuse!
    editor.DisablePackage(pkg)
}
```

---

## Migration Strategy

### Phase 1: Create new packages (no breaking changes)
1. Create `internal/search` with new code
2. Create `internal/ui` with new code
3. Create `internal/setup` with new code
4. Add tests for each

### Phase 2: Refactor install.go
1. Import new packages
2. Replace inline code with package calls
3. Remove old function definitions
4. Remove global variables

### Phase 3: Cleanup
1. Run full test suite
2. Test manually with real Nix setup
3. Update documentation
4. Remove old commented code

---

## Metrics

### Before
- `cmd/install.go`: 418 lines
- Packages: 1 (cmd)
- Test coverage: 0%
- Functions in install.go: 12

### After
- `cmd/install.go`: ~150 lines (-64%)
- Packages: 5 (cmd + 4 internal)
- Test coverage: >70% target
- Functions in install.go: ~6 (orchestration only)

---

## Implementation Priority
High - Foundation for all other improvements

## Dependencies
- Should be done alongside `nixconfig` and `assets` extractions
- Enables adding error handling package later
- Makes testing infrastructure easier to add

## Risk Mitigation
- Keep original install.go backed up
- Test each extraction independently
- Manual testing with real Nix configurations
- Consider feature flag for new code path initially
