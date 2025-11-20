# Feature: Subdirectory Scanning and Selection

## Goal
Allow users to navigate nested folder structures when selecting where to place a package module, instead of only showing the top-level folders.

## Current Behavior
```
User selects from: ["dev-tools", "media", "system-tools"]
Result: Package goes in apps/dev-tools/
```

## Desired Behavior
```
Step 1: Select from ["dev-tools", "media", "system-tools"]
  User picks: "dev-tools"

Step 2: Found subdirs! Select from ["[Use current folder]", "editors", "version-control"]
  User picks: "editors"

Step 3: Found subdirs! Select from ["[Use current folder]", "vim", "emacs"]
  User picks: "[Use current folder]"

Result: Package goes in apps/dev-tools/editors/
```

---

## Implementation Hints

### 1. The Loop Structure

You mentioned a while loop - that's a great instinct! Think about:
- **What condition keeps the loop going?** (Hint: as long as the user wants to keep navigating deeper OR there are subdirectories to explore)
- **What makes the loop stop?** (Hint: user chooses to stop, or no more subdirectories exist)

```go
// Pseudocode structure
currentPath := ""
for /* some condition */ {
    // 1. Where are we now?
    // 2. What folders exist here?
    // 3. Show options to user
    // 4. Handle user's choice
    // 5. Update currentPath or break
}
return currentPath
```

### 2. Building the Current Path

As the user navigates deeper, you need to build up a path like:
- Start: `""`
- After selecting "dev-tools": `"dev-tools"`
- After selecting "editors": `"dev-tools/editors"`

**Hint**: Go has a function perfect for this! Check `filepath.Join()` - it handles path separators correctly across OS.

### 3. Reading Subdirectories

You already have `getDirNames(path string)` which returns directory names. Consider:
- You need the **full path** to read: `filepath.Join(basePath, currentPath)`
- But you only want to **display and track** the relative path

### 4. The Options List

Think about what options to show the user:
- Should `"[Use current folder]"` always be available? (Hint: What if we're at the root level with no selection yet?)
- How do you combine the special option with the subdirectory list?

```go
// One approach:
var options []huh.Option[string]
if /* some condition */ {
    options = append(options, huh.NewOption("[Use current folder]", ???))
}
// Then add subdirectory options...
```

**Key question**: What value should you store for "[Use current folder]"?
- The display text is `"[Use current folder]"`
- But what value do you want back? (Hint: Think about how you'll detect this choice later)

### 5. Handling the User's Choice

After the user selects something, you need to decide:
- Did they pick a subdirectory? → Navigate deeper (update `currentPath`)
- Did they pick "[Use current folder]"? → Stop and return current path
- Are there no subdirectories? → Stop and return current path

### 6. Edge Cases to Consider

- **Empty folders**: What if `NIX_APPS_DIR` has no folders at all?
- **No subdirectories**: What if a folder has no subfolders? Should you auto-return it?
- **User cancels**: `huh.NewForm().Run()` returns an error if user presses Ctrl+C - handle this!

---

## Integration Points

### Where to add your function
Look at around line 157 in `install.go` - right after `getDirNames()` would be a good spot.

### Where to call it
Around line 230-237, you currently have:
```go
folders, err := getDirNames(NIX_APPS_DIR)
// ... error handling
folderOptions := huh.NewOptions(folders...)
// ... then later in the form
```

You'll need to replace this section to call your new function instead.

### Function Signature Suggestion
```go
func selectFolderRecursively(basePath string) (string, error)
```
- **Input**: The base directory to start from (e.g., `NIX_APPS_DIR`)
- **Output**: The selected path relative to basePath (e.g., `"dev-tools/editors"`)
- **Error**: If something goes wrong (directory doesn't exist, user cancels, etc.)

---

## Testing Strategy

### Test Case 1: Flat Structure
```
apps/
  ├── dev-tools/
  ├── media/
  └── system-tools/
```
Expected: Shows three options, no "[Use current folder]" at root, returns selected folder

### Test Case 2: Nested Structure
```
apps/
  └── dev-tools/
      ├── editors/
      │   ├── vim/
      │   └── emacs/
      └── version-control/
```
Expected: Can navigate multiple levels deep, "[Use current folder]" appears after first selection

### Test Case 3: No Subdirectories
```
apps/
  └── dev-tools/
      └── editors/   (empty, no subdirs)
```
Expected: After selecting editors, automatically returns (or shows "[Use current folder]" only)

---

## Helpful Go Patterns

### Infinite Loop with Break
```go
for {
    // Do stuff
    if someCondition {
        break // Exit loop
    }
}
```

### Building Paths
```go
base := "/apps"
relative := "dev-tools"
full := filepath.Join(base, relative) // "/apps/dev-tools"
```

### Checking Empty Slices
```go
dirs, err := getDirNames(path)
if len(dirs) == 0 {
    // No subdirectories!
}
```

---

## Questions to Think About

1. **Loop condition**: Should you use `for {}` (infinite loop with breaks) or `for someCondition {}` (conditional loop)?

2. **State tracking**: What variable(s) do you need to track as you navigate? Just `currentPath`? Anything else?

3. **Breadcrumbs**: How can you show the user where they are? (Hint: The form title!)

4. **Value vs Label**: In `huh.NewOption(label, value)`, what should be the value for each option?

5. **Return early**: Can you simplify by returning early in some cases? (e.g., if no folders exist at all)

---

## Bonus Challenges (Optional)

- Add a "Go back" option to navigate up a level
- Show the full path breadcrumb (e.g., "apps → dev-tools → editors")
- Add a description showing how many subdirectories exist at each level

---

## Need a Hint?

If you get stuck, think about these questions:
- What information do you need at each step of the loop?
- What decisions does the user need to make?
- When should the loop end?

Good luck! Take your time to think through the problem. The best way to learn is by working through it yourself. 🚀
