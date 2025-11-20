# Implementation Checklist

This checklist shows exactly which files you need to create to make all tests pass.

## 📋 Files to Create

### ✅ Already Done (Tests & Documentation)
- [x] `testdata/sample_config.nix`
- [x] `testdata/whitespace_config.nix`
- [x] `testdata/minimal_config.nix`
- [x] `internal/nixconfig/editor_test.go` (13 tests)
- [x] `internal/search/search_test.go` (8 tests)
- [x] `internal/assets/render_test.go` (9 tests)
- [x] `internal/ui/forms_test.go` (11 tests)
- [x] `internal/setup/setup_test.go` (10 tests)
- [x] `test.sh` (test runner script)
- [x] `TESTING.md` (testing guide)
- [x] `feats/errorHandling.md`
- [x] `feats/embeddedAssets.md`
- [x] `feats/nixconfigParsing.md`
- [x] `feats/codeOrganization.md`

### 🔨 To Implement

#### 1. internal/nixconfig/editor.go
Create the Nix configuration editor.

**Required exports:**
```go
type Editor struct { ... }
func NewEditor(content string) *Editor
func (e *Editor) Content() string
func (e *Editor) CategoryExists(category string) bool
func (e *Editor) PackageExistsInCategory(category, pkg string) bool
func (e *Editor) EnablePackage(pkg string) bool
func (e *Editor) AddPackageToCategory(category, pkg string) error
func (e *Editor) CreateCategory(category, pkg string) error
func (e *Editor) AddOrEnablePackage(category, pkg string) error
```

**Tests:** Run `go test -v ./internal/nixconfig`

**Reference:** `feats/nixconfigParsing.md`

---

#### 2. internal/search/search.go
Create the package search functionality.

**Required exports:**
```go
type Searcher struct { ... }
func NewSearcher(system string) *Searcher
func (s *Searcher) Search(packageName string) ([]types.Package, error)
func FilterTopLevel(packages []types.Package) []types.Package
```

**Helper functions (can be private):**
- Parse search results from JSON
- Extract system and path from Nix package path
- Filter by depth (top-level vs plugins)

**Tests:** Run `go test -v ./internal/search`

**Note:** This will execute actual `nix search` commands, so you may want to mock the command execution in tests.

---

#### 3. internal/assets/assets.go
Create the template system with embedded mkApp.txt.

**Required exports:**
```go
//go:embed templates/mkApp.nix
var mkAppTemplate string

func RenderMkApp(pkg *types.Package, useHomebrew bool) string
func GetTemplate() string
```

**File structure:**
```
internal/assets/
  assets.go                    # Main file with embed
  templates/
    mkApp.nix                  # Move mkApp.txt here, rename to .nix
```

**Tests:** Run `go test -v ./internal/assets`

**Reference:** `feats/embeddedAssets.md`

---

#### 4. internal/ui/forms.go
Create UI helper functions.

**Required exports:**
```go
func GetDirNames(path string) ([]string, error)
func FormatPackageOption(pkg *types.Package) string
func BuildPackageOptions(packages []types.Package) []PackageOption
func BuildHostOptions(hosts []string) []HostOption
func ValidateHostSelection(hosts []string) error
func BuildNavigationPath(basePath, currentPath string) string
func ShouldShowUseCurrentOption(currentPath string) bool

type PackageOption struct {
    Label string
    Value *types.Package
}

type HostOption struct {
    Label string
    Value string
}
```

**Optional (for actual UI - not tested):**
```go
func SelectPackage(packages []types.Package) (*types.Package, error)
func SelectHosts(hostDir string) ([]string, error)
func SelectFolder(rootPath string) (string, error)
```

**Tests:** Run `go test -v ./internal/ui`

---

#### 5. internal/setup/setup.go
Create initialization logic.

**Required exports:**
```go
type Initializer struct { ... }
func NewInitializer(cfg *internal.Config) *Initializer
func (i *Initializer) EnsureLibDirectory() error
func (i *Initializer) EnsureMkAppNix() error
func (i *Initializer) Run() error
```

**Tests:** Run `go test -v ./internal/setup`

**Note:** This package should use the `internal/assets` package to get the template content.

---

#### 6. cmd/install.go (Refactor)
Refactor the existing install.go to use new packages.

**Changes needed:**
1. Import new packages
2. Replace inline functions with package calls:
   - Search → use `internal/search`
   - Config manipulation → use `internal/nixconfig`
   - Template rendering → use `internal/assets`
   - UI helpers → use `internal/ui`
   - Setup → use `internal/setup`
3. Remove old function definitions
4. Eliminate global variables
5. Break down `install()` function

**Goal:** Reduce from 418 lines → ~150 lines

**No automated tests for this** - manual testing required with real Nix setup.

---

## 🎯 Implementation Order (Recommended)

### Phase 1: Independent Packages (No Dependencies)
```bash
1. internal/assets      ← Start here (simplest)
2. internal/ui          ← Pure helper functions
```

### Phase 2: Core Functionality
```bash
3. internal/nixconfig   ← Critical, uses regex
4. internal/search      ← Calls nix command
```

### Phase 3: Setup (Depends on Assets)
```bash
5. internal/setup       ← Uses internal/assets
```

### Phase 4: Integration
```bash
6. cmd/install.go       ← Refactor to use all packages
```

## 🚀 Quick Start

⚠️ **Note:** Tests will fail until you create the implementation files. This is expected!

```bash
# 1. Create package directories
mkdir -p internal/nixconfig internal/search internal/assets/templates internal/ui internal/setup

# 2. Move template file
cp mkApp.txt internal/assets/templates/mkApp.nix

# 3. Start with assets (easiest)
touch internal/assets/assets.go
# Implement based on assets_test.go
go test -v ./internal/assets

# 4. Continue with each package
# When all tests pass, you're done!
test  # or ./test.sh
```

## 📝 Running Tests (devenv)

```bash
# In your devenv shell:
test                      # Run all tests with test.sh
test-coverage            # Generate coverage report
test-package nixconfig   # Test specific package
```

## ✅ Success Criteria

You'll know you're done when:

- [ ] All 51 tests pass
- [ ] `./test.sh` shows green checkmarks
- [ ] Coverage is >70%
- [ ] `go build` succeeds
- [ ] cmd/install.go is under 200 lines
- [ ] Manual testing works with real Nix config

## 📊 Progress Tracking

Track your progress by running tests:

```bash
# Check overall status
./test.sh

# Check specific package
go test -v ./internal/nixconfig
```

When you see all green, you're done! 🎉

## 💡 Tips

1. **Read the tests first** - They show exactly what to implement
2. **Start simple** - Get one test passing, then move to the next
3. **Use the tests as documentation** - They define the API
4. **Don't over-engineer** - Implement just enough to pass tests
5. **Run tests frequently** - Fast feedback loop
6. **Check coverage** - `go test -cover ./internal/nixconfig`

## 🆘 If You Get Stuck

1. Read the corresponding `feats/*.md` file for design details
2. Look at the test code - it shows expected usage
3. Start with the simplest function in each package
4. Use `t.Logf()` in tests to debug
5. Check TESTING.md for troubleshooting tips

---

**Remember:** The tests are your specification. When they pass, you're done! ✨
