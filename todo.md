# PAM - Road to v1

## 🎯 What's Needed for v1

### Critical: Code Refactoring (Make tests pass)

**Status:** 51 tests written → Need implementation

Implement these 5 packages (see `IMPLEMENTATION_CHECKLIST.md` for details):

1. **`internal/assets/assets.go`** - Embed templates, fix hardcoded path FIXME
2. **`internal/nixconfig/editor.go`** - Robust config parsing with regex
3. **`internal/search/search.go`** - Package search logic
4. **`internal/ui/forms.go`** - UI helper functions
5. **`internal/setup/setup.go`** - Initialization logic

Then: 6. **Refactor `cmd/install.go`** - Use new packages (418 → ~150 lines)

**Run:** `test` to see what needs implementing

**Goal:** All 51 tests passing, >70% coverage

---

### Optional for v1: Lib Setup Completion

- [ ] Create/update `lib/default.nix` to export mkApp
- [ ] Check `flake.nix` for lib registration
- [ ] Add lib to flake outputs if missing

---

## ✅ Already Complete

**Core Features:**

- [x] Interactive UX with huh forms
- [x] Multi-host selection
- [x] Config file support (~/.config/pam/config)
- [x] Folder recursion/selection
- [x] Optional editor opening after install

**Code Quality:**

- [x] Critical bugs fixed
- [x] File permissions corrected
- [x] Configurable paths (no hardcoded NIXOS_ROOT)

**Testing & Documentation:**

- [x] 51 tests written (TDD approach)
- [x] 4 feature proposals (`feats/` directory)
- [x] Implementation guides (TESTING.md, IMPLEMENTATION_CHECKLIST.md)
- [x] Test infrastructure (test.sh, testdata/, devenv scripts)

---

## 🚀 Post-v1 Future Enhancements

**New Features:**

- [ ] Multi-architecture search (search x86_64-linux + aarch64-darwin simultaneously)
- [ ] Package removal/disabling command
- [ ] Dry-run mode (`--dry-run` flag)
- [ ] Advanced search filters (category, license, etc.)
- [ ] Generate flake.lock diff after install
- [ ] Rollback functionality

**Code Quality:**

- [ ] Structured error handling (`internal/errors` package)
- [ ] Better error messages with context
