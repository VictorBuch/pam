package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pam/internal"
	"pam/internal/types"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

type SearchResult map[string]types.Package

// FIXME: i think this will break when distributed since the file will not exist on other peoples computers
const (
	NIX_PKGS_MODULE_TEMPLATE = "./mkApp.txt"
)

var FLAKE_ROOT, NIX_APPS_DIR, NIX_HOSTS_DIR string

var (
	showAll         bool
	targetSystem    string
	installWithBrew bool
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
		fullPathParts := strings.Split(fullPath, ".")
		if len(fullPathParts) > 1 {
			pkg.System = fullPathParts[1]
		}
		if len(fullPathParts) > 2 {
			pkg.FullPath = strings.Join(fullPathParts[2:], ".")
		}
		if len(fullPathParts) > 3 {
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

func enablePackage(content string, packageName string) string {
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
	packageContent := "\n      " + packageName + ".enable = true;\n    "

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

func getDirNames(path string) ([]string, error) {
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

func selectFolderRecursively(path string) (string, error) {
	currentPath := ""
	for {
		fullPath := filepath.Join(path, currentPath)
		subdirs, err := getDirNames(fullPath)
		if err != nil {
			return "", err
		}
		if len(subdirs) == 0 {
			if currentPath == "" {
				return "", fmt.Errorf("no folders found in %s", path)
			}
			return currentPath, nil
		}

		var options []huh.Option[string]
		title := "Select a folder"
		if currentPath != "" {
			title = fmt.Sprintf("Select a folder (current: %s)", currentPath)
			options = append(options, huh.NewOption("Use this folder", ""))
		}
		for _, dir := range subdirs {
			options = append(options, huh.NewOption(dir, dir))
		}

		var selected string
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(title).
					Options(options...).
					Value(&selected),
			),
		).Run()
		if err != nil {
			return "", err
		}

		if selected == "" {
			return currentPath, nil
		}
		currentPath = filepath.Join(currentPath, selected)

	}
}

func replacePkgContent(data []byte, selectedPkg *types.Package) string {
	var linuxPackage string
	var darwinPackage string
	var homebrewPackage string

	if strings.Contains(selectedPkg.System, "linux") {
		linuxPackage = "pkgs." + selectedPkg.FullPath
	} else if strings.Contains(selectedPkg.System, "darwin") {
		if installWithBrew {
			homebrewPackage = selectedPkg.PName
		} else {
			darwinPackage = "pkgs." + selectedPkg.FullPath
		}
	}
	replacer := strings.NewReplacer("LinuxPackage", linuxPackage, "DarwinPackage", darwinPackage, "HomebrewPackage", homebrewPackage, "PackageName", selectedPkg.PName, "PackageDescription", selectedPkg.Description)
	return replacer.Replace(string(data))
}

func initialSetup() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	FLAKE_ROOT = cfg.FlakePath
	NIX_HOSTS_DIR = filepath.Join(FLAKE_ROOT, cfg.DefaultHostDir)
	NIX_APPS_DIR = filepath.Join(FLAKE_ROOT, cfg.DefaultModuleDir)

	// TODO: make this also check the default.nix for lib dir
	// make it also add the registration for the lib functions in the flake.nix
	libPath := filepath.Join(FLAKE_ROOT, "lib")
	mkAppPath := filepath.Join(libPath, "mkApp.nix")

	err = os.MkdirAll(libPath, 0o755)
	if err != nil {
		fmt.Printf("Error creating lib directory: %v\n", err)
		return
	}

	if _, err := os.Stat(mkAppPath); err == nil {
		return
	}
	data, err := os.ReadFile(NIX_PKGS_MODULE_TEMPLATE)
	if err != nil {
		fmt.Printf("Error reading template file: %v\n", err)
		return
	}

	err = os.WriteFile(mkAppPath, data, 0o644)
	if err != nil {
		fmt.Printf("Error writing mkApp.nix: %v\n", err)
		return
	}
	fmt.Println("Created lib/mkApp.nix from template")
}

func install(cmd *cobra.Command, args []string) {
	initialSetup()

	packageName := args[0]

	var packages SearchResult
	var searchErr error

	err := spinner.New().
		Title("Searching nix pkgs...").
		Action(func() {
			packages, searchErr = searchPackages(packageName, targetSystem)
		}).
		Run()
	if err != nil {
		fmt.Println("Error running spinner: ", err)
		return
	}

	if searchErr != nil {
		fmt.Println("Error: ", searchErr)
		return
	}

	filteredPkgs := filterAndPrioritizePackages(packages, showAll)

	if len(filteredPkgs) == 0 {
		fmt.Println("No packages found")
		return
	}

	pkgOptions := make([]huh.Option[*types.Package], len(filteredPkgs))
	for i := range filteredPkgs {
		pkg := &filteredPkgs[i]
		label := fmt.Sprintf("%s (%s) - %s", pkg.PName, pkg.Version, pkg.System)
		pkgOptions[i] = huh.NewOption(label, pkg)
	}

	hostDirs, err := getDirNames(NIX_HOSTS_DIR)
	if err != nil {
		fmt.Println("Failed to read nix modules directory: ", err)
		return
	}

	hostOptions := huh.NewOptions(hostDirs...)

	var selectedPkg *types.Package
	var selectedHosts []string
	var openAfterWriting bool

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[*types.Package]().
				Title("Select a package to install").
				Options(pkgOptions...).
				Value(&selectedPkg),
		)).Run()
	if err != nil {
		fmt.Println("Form cancelled or error: ", err)
		return
	}

	selectedFolder, err := selectFolderRecursively(NIX_APPS_DIR)
	if err != nil {
		fmt.Println("Selecting folders failed, error: ", err)
		return
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select hosts").
				Description("Space to toggle, Enter to confirm").
				Options(hostOptions...).
				Value(&selectedHosts),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Do you want to edit the module after adding it?").
				Value(&openAfterWriting),
		),
	).Run()
	if err != nil {
		fmt.Println("Form cancelled or error: ", err)
		return
	}

	data, err := os.ReadFile(NIX_PKGS_MODULE_TEMPLATE)
	if err != nil {
		fmt.Println("Error reading template file: ", err)
		return
	}

	modulePackage := replacePkgContent(data, selectedPkg)
	modulePath := filepath.Join(NIX_APPS_DIR, selectedFolder)
	moduleFilePath := filepath.Join(modulePath, packageName) + ".nix"

	err = os.WriteFile(moduleFilePath, []byte(modulePackage), 0o644)
	if err != nil {
		fmt.Println("could not write file: ", err)
		return
	}

	for _, host := range selectedHosts {
		fullHostPath := filepath.Join(NIX_HOSTS_DIR, host, "configuration.nix")
		data, err = os.ReadFile(fullHostPath)
		if err != nil {
			fmt.Println("Could not read the host configuration.nix, error: ", err)
			return
		}
		content := string(data)

		if categoryExists(content, selectedFolder) {
			if packageExistsInCategory(content, selectedFolder, selectedPkg.PName) {
				content = enablePackage(content, selectedPkg.PName)
			} else {
				content = addPackageToCategory(content, selectedFolder, selectedPkg.PName)
			}
		} else {
			content = createCategory(content, selectedFolder, selectedPkg.PName)
		}

		err = os.WriteFile(fullHostPath, []byte(content), 0o644)
		if err != nil {
			fmt.Println("could not write file: ", err)
			return
		}
		fmt.Printf("\nDone! please run: nixos-rebuild switch --flake %s#%s", FLAKE_ROOT, host)
	}

	if openAfterWriting {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nvim"
		}
		editorCmd := exec.Command(editor, moduleFilePath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		err := editorCmd.Run()
		if err != nil {
			fmt.Println("Error opening editor: ", err)
		}
	}
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
	installCmd.Flags().BoolVarP(&installWithBrew, "brew", "b", false, "Use Homebrew cask instead of nix package for Darwin")
}
