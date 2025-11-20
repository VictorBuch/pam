# Feature: Multi-Architecture Package Search

## Goal

Allow users to search for packages across multiple system architectures simultaneously, useful for developers working with both Linux and macOS systems.

## Current Behavior

```bash
pam install vim
# Only searches current system (or --system flag value)
# If you want both Linux and macOS, you need to run twice
```

## Desired Behavior

```bash
pam install vim --search-systems=x86_64-linux,aarch64-darwin
# Or configured in config.yaml

Results shown:
> vim (9.0.2116) - x86_64-linux
  vim (9.0.2116) - aarch64-darwin
  vim-plugins.vim-sensible (1.0) - x86_64-linux
  ...
```

---

## Design Decisions

### 1. How Should Users Specify Systems?

**Option A: CLI Flag (Array)**

```bash
pam install vim --search-systems=x86_64-linux,aarch64-darwin
```

**Option B: Config File**

```yaml
search_systems:
  - x86_64-linux
  - aarch64-darwin
```

**Option C: Both (with priority)**
CLI flag overrides config, or use config if flag not provided.

**Question**: What's most user-friendly? Should there be a shortcut like `--all-systems`?

### 2. How to Handle Duplicates?

Same package might exist for multiple architectures:

```
vim-9.0.2116 for x86_64-linux
vim-9.0.2116 for aarch64-darwin
```

**Should you**:

- Show both as separate entries? (Current plan - user chooses which)
- Group them together? "vim-9.0.2116 (available for: linux, darwin)"
- Show a merged version with all systems?

**Trade-off**: Separate entries = more control, grouped = less clutter

### 3. Performance Consideration

Running multiple `nix search` commands takes time:

- Sequential: Slow but simple (search1, then search2, then...)
- Parallel: Fast but more complex (search all at once)

**Hint**: Go's goroutines make parallel execution easy!

---

## Implementation Hints

### Phase 1: Understanding Current Search

Look at `searchPackages()` in install.go (line 36):

```go
func searchPackages(packageName string, system string) (SearchResult, error)
```

Currently searches for ONE system. How can you extend this?

**Option A: Modify existing function**

```go
func searchPackages(packageName string, systems []string) (SearchResult, error)
    // Loop or parallelize
}
```

**Option B: New function that calls existing**

```go
func searchMultipleArchitectures(packageName string, systems []string) (SearchResult, error) {
    // Call searchPackages() for each system
    // Merge results
}
```

**Question**: Which is cleaner? Option B keeps backward compatibility if needed elsewhere.

### Phase 2: Parallel Execution

Go makes parallel execution easy with goroutines. Consider this pattern:

```go
func searchMultipleSystems(pkg string, systems []string) (SearchResult, error) {
    // What data structure do you need to collect results?
    // How do you wait for all goroutines to finish?
    // How do you handle errors from any goroutine?

    for _, system := range systems {
        go func(sys string) {
            // Search for this system
            // How do you send results back?
        }(system)
    }

    // Wait for all to complete
    // Merge results
    // Return combined result
}
```

**Helpful Go concepts**:

- **Goroutines**: `go func() { ... }()`
- **Channels**: `make(chan ResultType, bufferSize)`
- **WaitGroups**: `sync.WaitGroup` to wait for all goroutines
- **Error handling**: How to collect errors from multiple goroutines?

### Phase 3: Merging Results

Each search returns `SearchResult` which is `map[string]types.Package`.

Example results:

```go
// From x86_64-linux search:
{
  "legacyPackages.x86_64-linux.vim": Package{...},
  "legacyPackages.x86_64-linux.vim-plugins.sensible": Package{...},
}

// From aarch64-darwin search:
{
  "legacyPackages.aarch64-darwin.vim": Package{...},
  "legacyPackages.aarch64-darwin.vim-plugins.sensible": Package{...},
}
```

**Question**: How do you merge these maps?

- Simple append: Add all entries (duplicates by system will have different keys)
- Deduplicate: Check package name and version, keep both architectures
- Group: Combine packages with same name/version

**Hint**: The current code already extracts system from the key (line 61-64). This should work naturally!

### Phase 4: User Selection

After merge, `filteredPkgs` will have entries for multiple systems.

Current package selector shows:

```
vim (9.0.2116) - x86_64-linux
```

With multi-arch:

```
vim (9.0.2116) - x86_64-linux
vim (9.0.2116) - aarch64-darwin
```

**The user needs to pick which architecture they want!** This is important.

**Question**: Should you auto-detect current system and pre-select it?

---

## Goroutines Pattern

### Basic Pattern with WaitGroup

```go
var wg sync.WaitGroup
results := make(map[string]types.Package)
var mu sync.Mutex // To safely write to shared map

for _, system := range systems {
    wg.Add(1)
    go func(sys string) {
        defer wg.Done()

        result, err := searchPackages(packageName, sys)
        if err != nil {
            // How to handle error?
            return
        }

        // Safely merge into results
        mu.Lock()
        for k, v := range result {
            results[k] = v
        }
        mu.Unlock()
    }(system)
}

wg.Wait() // Wait for all searches to complete
return results, nil
```

**Key concepts**:

- `sync.WaitGroup`: Coordinates multiple goroutines
- `sync.Mutex`: Protects shared data (the results map)
- `defer wg.Done()`: Always signal completion
- Closure captures `system` - need to pass as parameter!

### Pattern with Channels

```go
type searchResult struct {
    data SearchResult
    err  error
}

resultChan := make(chan searchResult, len(systems))

for _, system := range systems {
    go func(sys string) {
        result, err := searchPackages(packageName, sys)
        resultChan <- searchResult{data: result, err: err}
    }(system)
}

// Collect results
allResults := make(map[string]types.Package)
for i := 0; i < len(systems); i++ {
    res := <-resultChan
    if res.err != nil {
        // Handle error
        continue
    }
    // Merge res.data into allResults
}

return allResults, nil
```

**Question**: Which pattern do you prefer? WaitGroup or channels?

---

## Edge Cases to Handle

### 1. One Search Fails

You search 3 systems, one fails:

- Should you return an error?
- Or return partial results with a warning?
- What's most user-friendly?

### 2. Empty Results

All searches return no packages:

- Same as current behavior (no packages found)

### 3. Different Versions Available

```
vim 9.0.2116 for Linux
vim 9.0.2100 for macOS
```

Both should be shown! User can see version differences.

### 4. Very Slow Searches

One system search takes 10 seconds:

- Parallel execution helps here
- Could add a timeout?
- Progress indicator?

---

## Integration Points

### Where to Modify

**1. Update `searchPackages()` or create new function** (line 36)

```go
// Option A: Modify signature
func searchPackages(packageName string, systems []string) (SearchResult, error)

// Option B: New function
func searchMultiArchitectures(packageName string, systems []string) (SearchResult, error)
```

**2. Update install command** (line 210)

```go
// Instead of:
packages, err := searchPackages(packageName, targetSystem)

// Do:
systems := getSystemsToSearch() // From config or flag
packages, err := searchMultiArchitectures(packageName, systems)
```

**3. Add flag to installCmd** (line 355)

```go
installCmd.Flags().StringSliceVarP(&searchSystems, "search-systems", "", []string{}, "Systems to search (e.g., x86_64-linux,aarch64-darwin)")
```

### Determining Systems to Search

Priority:

1. `--search-systems` flag (if provided)
2. `search_systems` from config (if set)
3. Current system only (fallback to current behavior)

```go
func getSystemsToSearch(config *Config, flagValue []string) []string {
    if len(flagValue) > 0 {
        return flagValue
    }
    if len(config.SearchSystems) > 0 {
        return config.SearchSystems
    }
    // Fallback: detect current system or use targetSystem flag
    return []string{getCurrentSystem()}
}
```

---

## Testing Strategy

### Test Case 1: Single System (Backward Compatibility)

```bash
pam install vim
# Should work exactly as before
```

### Test Case 2: Two Systems

```bash
pam install vim --search-systems=x86_64-linux,aarch64-darwin
# Should show results from both
```

### Test Case 3: System from Config

```yaml
# config.yaml
search_systems:
  - x86_64-linux
  - aarch64-darwin
```

```bash
pam install vim
# Should automatically search both systems
```

### Test Case 4: One System Fails

Mock one search to fail, verify:

- Other results still shown OR
- Appropriate error message

---

## Common Nix Systems

Here are the common system architectures to consider:

**Linux**:

- `x86_64-linux` - 64-bit Intel/AMD Linux
- `aarch64-linux` - 64-bit ARM Linux (Raspberry Pi 4, servers)
- `i686-linux` - 32-bit Linux (rare)

**macOS**:

- `x86_64-darwin` - Intel Mac
- `aarch64-darwin` - Apple Silicon (M1/M2/M3)

**Others** (less common):

- `armv7l-linux` - 32-bit ARM Linux (older Raspberry Pi)

**Hint**: Most users care about `x86_64-linux`, `aarch64-darwin`, and maybe `aarch64-linux`.

---

## Performance Optimization Ideas

### 1. Caching

Cache search results temporarily:

```go
// If searching same package for multiple systems in succession
// Cache to avoid duplicate work
```

### 2. Timeout Per Search

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
// Pass context to search function
```

### 3. Progress Indicator

For slow searches, show user what's happening:

```
Searching x86_64-linux... ✓
Searching aarch64-darwin... ⏳
```

**Hint**: This is advanced! Maybe save for after v1.

---

## Config File Integration

Update `Config` struct in config.md implementation:

```go
type Config struct {
    FlakePath      string   `yaml:"flake_path"`
    DefaultEditor  string   `yaml:"default_editor"`
    DefaultSystem  string   `yaml:"default_system"`
    SearchSystems  []string `yaml:"search_systems,omitempty"`  // NEW
}
```

Example config:

```yaml
flake_path: "/home/user/nixos-config"
search_systems:
  - x86_64-linux
  - aarch64-darwin
```

---

## Questions to Think About

1. **Error strategy**: Fail fast or collect partial results?
2. **Performance**: Is parallel search worth the complexity for 2-3 systems?
3. **User experience**: How to make it obvious which architecture they're selecting?
4. **Default behavior**: Should single-system search stay the default?
5. **Concurrency safety**: How do you safely merge results from goroutines?

---

## Go Concurrency Resources

- **Tour of Go - Concurrency**: https://go.dev/tour/concurrency/1
- **Goroutines**: https://gobyexample.com/goroutines
- **Channels**: https://gobyexample.com/channels
- **WaitGroups**: https://gobyexample.com/waitgroups
- **Mutexes**: https://gobyexample.com/mutexes

---

## Implementation Steps

### Step 1: Multi-System Search Function

1. Create function to search multiple systems
2. Decide: sequential or parallel?
3. Test with 2 systems

### Step 2: Result Merging

1. Combine search results from all systems
2. Ensure no data loss
3. Test deduplication if implemented

### Step 3: CLI Flag

1. Add `--search-systems` flag
2. Parse comma-separated values
3. Test flag works

### Step 4: Config Integration

1. Add `search_systems` to config struct
2. Load from config file
3. Implement priority: flag > config > default

### Step 5: User Selection

1. Verify package selector shows all architectures
2. Test that correct package is selected
3. Ensure downstream code handles architecture correctly

---

## Bonus Challenges (Optional)

1. **Smart filtering**: If searching from macOS, show darwin packages first
2. **System detection**: Auto-include current system in multi-arch search
3. **Visual grouping**: Group packages by name, show available systems
4. **Fast mode**: `--quick` only searches current system even if config has multiple
5. **System aliases**: Allow `linux` → `x86_64-linux`, `macos` → `aarch64-darwin`

Good luck! Parallel search is a great feature that showcases Go's concurrency strengths. 🚀
