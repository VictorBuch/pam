package search

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"pam/internal/types"
)

type SearchResult map[string]types.Package

func SearchPackages(packageName string, system string) (SearchResult, error) {
	args := []string{"search", "nixpkgs", packageName, "--json"}

	if system != "" {
		args = append(args, "--system", system)
	}

	cmd := exec.Command("nix", args...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, fmt.Errorf("Search failed: %w", err)
	}
	var result SearchResult
	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

func FilterAndPrioritizePackages(packages SearchResult, showAll bool) []types.Package {
	var topLevel []types.Package
	var plugins []types.Package

	for fullPath, pkg := range packages {
		fullPathParts := strings.Split(fullPath, ".")
		if len(fullPathParts) > 1 {
			pkg.System = fullPathParts[1]
		}
		if len(fullPathParts) > 2 {
			pkg.FullPath = strings.Join(fullPathParts[2:], ".")
		}
		if len(fullPathParts) == 3 {
			topLevel = append(topLevel, pkg)
		} else if len(fullPathParts) > 3 {
			plugins = append(plugins, pkg)
		}
		// Paths with < 3 parts are malformed, skip them
	}
	if showAll {
		return append(topLevel, plugins...)
	} else {
		return topLevel
	}
}
