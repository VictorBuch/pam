package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"pam/internal"
	"pam/internal/assets"
	"pam/internal/nixconfig"
	"pam/internal/search"
	"pam/internal/setup"
	"pam/internal/types"
	"pam/internal/ui"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var NIX_APPS_DIR, NIX_HOSTS_DIR string

var (
	showAll         bool
	targetSystem    string
	installWithBrew bool
)

func selectFolderRecursively(path string) (string, error) {
	currentPath := ""
	for {
		fullPath := filepath.Join(path, currentPath)
		subdirs, err := ui.GetDirNames(fullPath)
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

func install(cmd *cobra.Command, args []string) {
	cfg, err := internal.LoadConfig()
	if err != nil {
		fmt.Printf("Loading config failed. error: %v", err)
		return
	}
	NIX_HOSTS_DIR = filepath.Join(cfg.FlakePath, cfg.DefaultHostDir)
	NIX_APPS_DIR = filepath.Join(cfg.FlakePath, cfg.DefaultModuleDir)

	init := setup.NewInitializer(cfg)
	err = init.Run()
	if err != nil {
		fmt.Printf("Setup failed. error: %v", err)
		return
	}

	packageName := args[0]

	var packages search.SearchResult
	var searchErr error

	err = spinner.New().
		Title("Searching nix pkgs...").
		Action(func() {
			packages, searchErr = search.SearchPackages(packageName, targetSystem)
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

	filteredPkgs := search.FilterAndPrioritizePackages(packages, showAll)

	if len(filteredPkgs) == 0 {
		fmt.Println("No packages found")
		return
	}

	pkgOptions := make([]huh.Option[*types.Package], len(filteredPkgs))
	for i := range filteredPkgs {
		pkg := &filteredPkgs[i]
		label := ui.FormatPackageOption(pkg)
		pkgOptions[i] = huh.NewOption(label, pkg)
	}

	hostDirs, err := ui.GetDirNames(NIX_HOSTS_DIR)
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

	modulePackage := assets.FillPackageTemplate(selectedPkg, installWithBrew)
	modulePath := filepath.Join(NIX_APPS_DIR, selectedFolder)
	moduleFilePath := filepath.Join(modulePath, packageName) + ".nix"

	err = os.WriteFile(moduleFilePath, []byte(modulePackage), 0o644)
	if err != nil {
		fmt.Println("could not write file: ", err)
		return
	}

	for _, host := range selectedHosts {
		fullHostPath := filepath.Join(NIX_HOSTS_DIR, host, "configuration.nix")
		data, err := os.ReadFile(fullHostPath)
		if err != nil {
			fmt.Println("Could not read the host configuration.nix, error: ", err)
			return
		}

		nixcfg := nixconfig.NewConfig(string(data))

		err = nixcfg.EnsureAppsSectionExists()
		if err != nil {
			fmt.Println("Error ensuring apps section: ", err)
			return
		}

		err = nixcfg.AddOrEnablePackage(selectedFolder, selectedPkg.PName)
		if err != nil {
			fmt.Println("Error updating config: ", err)
			return
		}

		err = os.WriteFile(fullHostPath, []byte(nixcfg.Content()), 0o644)
		if err != nil {
			fmt.Println("could not write file: ", err)
			return
		}

		fmt.Printf("\nDone! please run: nixos-rebuild switch --flake %s#%s", cfg.FlakePath, host)
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
