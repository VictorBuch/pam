package ui

import (
	"fmt"
	"os"

	"pam/internal/types"
)

// GetDirNames returns a list of directory names in the given path
func GetDirNames(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

// FormatPackageOption formats a package for display in selection UI
func FormatPackageOption(pkg *types.Package) string {
	return fmt.Sprintf("%s (%s) - %s", pkg.PName, pkg.Version, pkg.System)
}
