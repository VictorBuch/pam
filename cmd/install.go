package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

type Package struct {
	PName       string `json:"pname"`
	Version     string `json:"version"`
	Description string `json:"description"`
	FullPath    string
}

type SearchResult map[string]Package

func searchPackages(packageName string) (SearchResult, error) {
	cmd := exec.Command("nix", "search", "nixpkgs", packageName, "--json")
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

func install(cmd *cobra.Command, args []string) {
	packageName := args[0]
	packages, err := searchPackages(packageName)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	if len(packages) == 0 {
		fmt.Println("No packages found")
		return
	}

	fmt.Printf("\nFound %d package(s):\n\n", len(packages))

	i := 1
	for fullPath, pkg := range packages {
		fmt.Printf("%d. %s (v%s)\n", i, pkg.PName, pkg.Version)
		fmt.Printf("   %s\n", pkg.Description)
		fmt.Printf("   Path: %s\n\n", fullPath)
		i++
	}

	var choice uint16
	fmt.Print("Select a package number to install: ")
	fmt.Scan(&choice)
	fmt.Printf("You selected: %d", choice)
}

var installCmd = &cobra.Command{
	Use:   "install [package]",
	Short: "Install a nix package to your system",
	Args:  cobra.MinimumNArgs(1),
	Run:   install,
}

func init() {
	rootCmd.AddCommand(installCmd)
}
