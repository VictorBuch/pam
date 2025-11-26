package setup

import (
	"fmt"
	"regexp"
	"strings"
)

// hasExport checks if a Nix attrset exports a specific name
// Looks for patterns like "name = " in the content
func hasExport(content, exportName string) bool {
	// Split into lines and check each line
	lines := strings.Split(content, "\n")

	// Create regex pattern that handles any amount of whitespace
	// Pattern: optional whitespace, exportName, optional whitespace, =
	pattern := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(exportName) + `\s*=`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check if this line exports the name
		if pattern.MatchString(line) {
			return true
		}
	}

	return false
}

// canSafelyModify checks if a Nix attrset file has a simple structure
// that we can safely modify by adding an export
func canSafelyModify(content string) bool {
	// Look for a simple attrset structure: { ... }
	// Should have opening { and closing }
	openBrace := strings.Contains(content, "{")
	closeBrace := strings.Contains(content, "}")

	if !openBrace || !closeBrace {
		return false
	}

	// Check that it's not doing complex things like:
	// - Multiple attrsets
	// - Let expressions
	// - Imports at the root level (beyond the function arg)
	complexPatterns := []string{
		`let\s+`,       // let expressions
		`inherit\s+\(`, // inherit from other attrsets
		`rec\s+{`,      // recursive attrsets
	}

	for _, pattern := range complexPatterns {
		matched, _ := regexp.MatchString(pattern, content)
		if matched {
			return false
		}
	}

	return true
}

// addExportToAttrSet adds a new export to a Nix attrset
// Assumes the file has format: { lib }: { exports... }
func addExportToAttrSet(content, exportName, importPath string) (string, error) {
	if !canSafelyModify(content) {
		return "", fmt.Errorf("file structure is too complex to safely modify")
	}

	// Find the position to insert - look for the closing }
	// We'll insert before the last }
	lines := strings.Split(content, "\n")

	// Find the last non-empty, non-comment line with }
	var insertIdx int
	var foundClosingBrace bool
	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, "}") {
			insertIdx = i
			foundClosingBrace = true
			break
		}
	}

	if !foundClosingBrace {
		return "", fmt.Errorf("could not find closing brace in attrset")
	}

	// Create the new export line
	newExport := fmt.Sprintf("  %s = %s;", exportName, importPath)

	// Check if there are other exports - if so, add a newline before
	hasOtherExports := false
	for i := 0; i < insertIdx; i++ {
		if strings.Contains(lines[i], " = ") && !strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			hasOtherExports = true
			break
		}
	}

	// Insert the new export
	var result []string
	if hasOtherExports {
		// Add before the closing brace with proper spacing
		result = append(lines[:insertIdx], newExport)
		result = append(result, lines[insertIdx:]...)
	} else {
		// First export - add with a blank line before closing brace
		result = append(lines[:insertIdx], newExport, "")
		result = append(result, lines[insertIdx:]...)
	}

	return strings.Join(result, "\n"), nil
}

// detectFlakeLibPatterns searches for common patterns indicating lib is integrated
func detectFlakeLibPatterns(content string) (integrated bool, pattern string) {
	patterns := map[string]string{
		`lib\s*=\s*import\s+\./lib`:                 "outputs.lib",
		`customLib\s*=\s*import\s+\./lib`:           "let-customLib",
		`inherit\s*\([^)]*customLib[^)]*\)\s*mkApp`: "specialArgs-inherit",
		`mkApp\s*=`: "direct-mkApp",
	}

	for regex, patternName := range patterns {
		matched, _ := regexp.MatchString(regex, content)
		if matched {
			return true, patternName
		}
	}

	return false, "unknown"
}
