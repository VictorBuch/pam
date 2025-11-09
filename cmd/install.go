package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"nap/internal/tui"
	"nap/internal/types"

	"github.com/spf13/cobra"
)

type SearchResult map[string]types.Package

const (
	NIXOS_ROOT               = "/Users/victorbuch/serenityOs"
	NIX_PKGS_MODULE_TEMPLATE = "./mkApp.txt"
)

var (
	NIX_APPS_DIR  = filepath.Join(NIXOS_ROOT, "modules", "apps")
	NIX_HOSTS_DIR = filepath.Join(NIXOS_ROOT, "hosts")
)

var (
	showAll      bool
	targetSystem string
)

func searchPackages(packageName string, system string) (SearchResult, error) {
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

func filterAndPrioritizePackages(packages SearchResult, showAll bool) []types.Package {
	var topLevel []types.Package
	var plugins []types.Package

	for fullPath, pkg := range packages {
		pkg.FullPath = fullPath
		parts := strings.Split(fullPath, ".")
		if len(parts) > 3 {
			plugins = append(plugins, pkg)
		} else {
			topLevel = append(topLevel, pkg)
		}
	}
	if showAll {
		return append(topLevel, plugins...)
	} else {
		return topLevel
	}
}

func categoryExists(content string, category string) bool {
	pattern := category + " = {"
	return strings.Contains(content, pattern)
}

func packageExistsInCategory(content string, category string, packageName string) bool {
	categoryStart := category + " = {"
	startPos := strings.Index(content, categoryStart)
	if startPos == -1 {
		return false
	}
	endPos := strings.Index(content[startPos:], "};")
	if endPos == -1 {
		return false
	}
	categorySection := content[startPos : startPos+endPos]
	packagePattern := packageName + ".enable"
	return strings.Contains(categorySection, packagePattern)
}

func enablePackage(content string, category string, packageName string) string {
	oldPattern := packageName + ".enable = false;"
	newPattern := packageName + ".enable = true;"

	return strings.Replace(content, oldPattern, newPattern, 1)
}

func addPackageToCategory(content string, category string, packageName string) string {
	categoryStart := category + " = {"
	startPos := strings.Index(content, categoryStart)
	if startPos == -1 {
		fmt.Println("Error: 'start position' section not found in configuration.nix")
		return content
	}
	endPos := strings.Index(content[startPos:], "};")
	if endPos == -1 {
		fmt.Println("Error: 'end position' section not found in configuration.nix")
		return content
	}

	absoluteEndPos := startPos + endPos

	fmt.Println("startPos:", startPos, "endPos:", endPos,
		"absoluteEndPos:", absoluteEndPos)
	packageContent := "\n        " + packageName + ".enable = true;"

	before := content[:absoluteEndPos]
	after := content[absoluteEndPos:]

	return before + packageContent + after
}

func createCategory(content string, category string, packageName string) string {
	appStart := "apps = {"
	startPos := strings.Index(content, appStart)
	if startPos == -1 {
		fmt.Println("Error: 'apps' section not found in configuration.nix")
		return content
	}
	insertPos := startPos + len(appStart)
	newCategory := fmt.Sprintf("\n    %s = {\n      %s.enable = true;\n    };\n", category, packageName)
	before := content[:insertPos]
	after := content[insertPos:]
	return before + newCategory + after
}

func install(cmd *cobra.Command, args []string) {
	packageName := args[0]
	packages, err := searchPackages(packageName, targetSystem)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	filteredPkgs := filterAndPrioritizePackages(packages, showAll)

	if len(filteredPkgs) == 0 {
		fmt.Println("No packages found")
		return
	}

	selectedPkg, err := tui.ShowPackageSelector(filteredPkgs)
	if err != nil {
		fmt.Println("Failed to select package: ", err)
		return
	}

	// Scan the apps directory to ask where to add this package
	files, err := os.ReadDir(NIX_APPS_DIR)
	if err != nil {
		fmt.Println("Failed to read nix modules directory: ", err)
		return
	}

	var folders []string
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file.Name())
		}
	}
	// ask user to select a folder
	for i, f := range folders {
		fmt.Printf("\n[%d] - %s", i+1, f)
	}
	var choice uint16
	fmt.Print("\nSelect a folder to put the package module: ")
	fmt.Scan(&choice)
	fmt.Printf("\nYou selected: %d", choice)

	selectedFolder := folders[choice-1]
	fullModulePath := filepath.Join(NIX_APPS_DIR, selectedFolder)
	fmt.Println(fullModulePath)

	// TODO: We should also rescan the selected folder to see if any nested dirs exist and reprompt - do later

	data, err := os.ReadFile(NIX_PKGS_MODULE_TEMPLATE)
	if err != nil {
		fmt.Println("Error reading template file: ", err)
		return
	}
	replacer := strings.NewReplacer("PackageToReplace", selectedPkg.FullPath, "PackageName", selectedPkg.PName, "PackageDescription", selectedPkg.Description)
	modulePackage := replacer.Replace(string(data))
	modulePackageBytes := []byte(modulePackage)

	err = os.WriteFile(filepath.Join(fullModulePath, packageName)+".nix", modulePackageBytes, 0o666)
	if err != nil {
		fmt.Println("could not write file: ", err)
		return
	}

	// TODO: Ask for if they want to modify the package themselves, then open using $EDITOR

	hosts, err := os.ReadDir(NIX_HOSTS_DIR)
	if err != nil {
		fmt.Println("Failed to read nix modules directory: ", err)
		return
	}
	var hostDirs []string
	for _, host := range hosts {
		if host.IsDir() {
			hostDirs = append(hostDirs, host.Name())
		}
	}

	for i, h := range hostDirs {
		fmt.Printf("\n[%d] - %s", i+1, h)
	}
	fmt.Print("\nSelect a host to enable the pkgs in: ")
	fmt.Scan(&choice)
	fmt.Printf("\nYou selected: %d", choice)

	selectedHost := hostDirs[choice-1]
	fullHostPath := filepath.Join(NIX_HOSTS_DIR, selectedHost, "configuration.nix")
	fmt.Println(fullHostPath)

	data, err = os.ReadFile(fullHostPath)
	if err != nil {
		fmt.Println("Could not read the host condifuration.nix, error: ", err)
		return
	}
	content := string(data)
	fmt.Println(content)
	fmt.Println("category: ", selectedFolder)

	if categoryExists(content, selectedFolder) {
		if packageExistsInCategory(content, selectedFolder, selectedPkg.PName) {
			fmt.Println("Package already exists. Enabling it...")
			content = enablePackage(content, selectedFolder, selectedPkg.PName)
		} else {
			fmt.Println("Category exists but package does not. Adding it to category...")
			content = addPackageToCategory(content, selectedFolder, selectedPkg.PName)
		}
	} else {
		fmt.Println("No category exists. Adding category...")
		content = createCategory(content, selectedFolder, selectedPkg.PName)

	}
	err = os.WriteFile(fullHostPath, []byte(content), 0o666)
	if err != nil {
		fmt.Println("could not write file: ", err)
		return
	}
	fmt.Printf("Done! please run: nixos-rebuild switch --flake %s#%s", NIXOS_ROOT, selectedHost)
}

var installCmd = &cobra.Command{
	Use:   "install [package]",
	Short: "Install a nix package to your system",
	Args:  cobra.MinimumNArgs(1),
	Run:   install,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVarP(&showAll, "show-all", "a", false, "Show all packages including plugins")
	installCmd.Flags().StringVarP(&targetSystem, "system", "s", "", "Target system architecture (e.g., x86_64-linux, aarch64-darwin)")
}
