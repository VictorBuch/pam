package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pam/internal"
	"pam/internal/assets"
	"pam/internal/ui"
)

type Initializer struct {
	config *internal.Config
}

func NewInitializer(cfg *internal.Config) *Initializer {
	return &Initializer{config: cfg}
}

func (i *Initializer) EnsureLibDirectory() error {
	libPath := filepath.Join(i.config.FlakePath, "lib")
	return os.MkdirAll(libPath, 0o755)
}

func (i *Initializer) EnsureMkAppNix() error {
	mkAppPath := filepath.Join(i.config.FlakePath, "lib", "mkApp.nix")

	if _, err := os.Stat(mkAppPath); err == nil {
		return nil // File exists, don't overwrite
	}

	template := assets.GetMkApp()
	return os.WriteFile(mkAppPath, []byte(template), 0o644)
}

func (i *Initializer) EnsureLibDefault() error {
	libDefaultPath := filepath.Join(i.config.FlakePath, "lib", "default.nix")

	// Check if file exists
	existingContent, err := os.ReadFile(libDefaultPath)
	if err != nil {
		// File doesn't exist - create it from template
		if os.IsNotExist(err) {
			template := assets.GetLibDefault()
			return os.WriteFile(libDefaultPath, []byte(template), 0o644)
		}
		return fmt.Errorf("failed to read lib/default.nix: %w", err)
	}

	// File exists - check if it exports mkApp
	content := string(existingContent)
	if hasExport(content, "mkApp") {
		// mkApp is already exported, nothing to do
		return nil
	}

	// mkApp is missing - ask user if they want to add it
	fmt.Println("\n⚠️  Found lib/default.nix but it doesn't export mkApp")
	fmt.Println("mkApp is required for pam-generated modules to work.")

	if !ui.Confirm("Would you like to add mkApp export to lib/default.nix?") {
		fmt.Println("\nℹ️  Skipping. Please add this to your lib/default.nix manually:")
		fmt.Println("  mkApp = import ./mkApp.nix { inherit lib; };")
		return nil
	}

	// Try to add mkApp to the file
	newContent, err := addExportToAttrSet(content, "mkApp", "import ./mkApp.nix { inherit lib; }")
	if err != nil {
		fmt.Printf("\n⚠️  Could not automatically modify lib/default.nix: %v\n", err)
		fmt.Println("\nPlease add this to your lib/default.nix manually:")
		fmt.Println("  mkApp = import ./mkApp.nix { inherit lib; };")
		return nil
	}

	// Write the modified content
	if err := os.WriteFile(libDefaultPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write lib/default.nix: %w", err)
	}

	fmt.Println("✓ Added mkApp export to lib/default.nix")
	return nil
}

func (i *Initializer) DetectFlakeLibIntegration() (bool, string) {
	flakePath := filepath.Join(i.config.FlakePath, "flake.nix")

	content, err := os.ReadFile(flakePath)
	if err != nil {
		// Flake doesn't exist or can't be read
		return false, "no-flake"
	}

	return detectFlakeLibPatterns(string(content))
}

func (i *Initializer) PrintSetupInstructions(pattern string) {
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("⚠️  MANUAL SETUP REQUIRED")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("\nYour lib/ directory is set up, but your flake.nix needs to be updated")
	fmt.Println("to make the lib functions available to your NixOS modules.")

	if pattern == "no-flake" {
		fmt.Println("\n❌ Could not find flake.nix in your configuration directory.")
		fmt.Println("pam requires a flake-based NixOS configuration.")
		return
	}

	fmt.Println("\n📝 Please add the following to your flake.nix:")

	fmt.Println("1. Import your lib directory in a 'let' binding:")
	fmt.Println("   outputs = { self, nixpkgs, ... }: let")
	fmt.Println("     customLib = import ./lib { lib = nixpkgs.lib; };")
	fmt.Println("   in { ... }")

	fmt.Println("\n2. Pass mkApp (and other lib functions) via specialArgs:")
	fmt.Println("   nixosConfigurations.yourhost = nixpkgs.lib.nixosSystem {")
	fmt.Println("     specialArgs = {")
	fmt.Println("       inherit inputs system;")
	fmt.Println("       inherit (customLib) mkApp;  # Add this line")
	fmt.Println("     };")
	fmt.Println("     modules = [ ... ];")
	fmt.Println("   };")

	fmt.Println("\n💡 Alternative: You can also export lib in outputs:")
	fmt.Println("   outputs = { ... }: {")
	fmt.Println("     lib = import ./lib { lib = nixpkgs.lib; };")
	fmt.Println("     # ... rest of outputs")
	fmt.Println("   }")

	fmt.Println("\n📚 For more details, see your existing flake.nix or NixOS flakes docs.")
	fmt.Println(strings.Repeat("=", 70))
}

func (i *Initializer) Run() error {
	if err := i.EnsureLibDirectory(); err != nil {
		return err
	}

	if err := i.EnsureMkAppNix(); err != nil {
		return err
	}

	if err := i.EnsureLibDefault(); err != nil {
		return err
	}

	// Check if flake.nix properly integrates lib
	integrated, pattern := i.DetectFlakeLibIntegration()
	if !integrated {
		i.PrintSetupInstructions(pattern)
	}

	return nil
}
