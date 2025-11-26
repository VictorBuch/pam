package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pam/internal"
)

func TestInitializer_EnsureLibDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	err := init.EnsureLibDirectory()
	if err != nil {
		t.Fatalf("EnsureLibDirectory() error = %v", err)
	}

	// Verify lib directory was created
	libPath := filepath.Join(tmpDir, "lib")
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		t.Error("lib directory was not created")
	}

	// Verify permissions
	info, err := os.Stat(libPath)
	if err != nil {
		t.Fatalf("Failed to stat lib directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("lib is not a directory")
	}

	// Test idempotency - should not error if directory already exists
	err = init.EnsureLibDirectory()
	if err != nil {
		t.Errorf("EnsureLibDirectory() failed on second call: %v", err)
	}
}

func TestInitializer_EnsureMkAppNix(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	// First ensure lib directory exists
	libPath := filepath.Join(tmpDir, "lib")
	err := os.MkdirAll(libPath, 0o755)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Test creating mkApp.nix
	err = init.EnsureMkAppNix()
	if err != nil {
		t.Fatalf("EnsureMkAppNix() error = %v", err)
	}

	// Verify mkApp.nix was created
	mkAppPath := filepath.Join(libPath, "mkApp.nix")
	if _, err := os.Stat(mkAppPath); os.IsNotExist(err) {
		t.Error("mkApp.nix was not created")
	}

	// Verify file content
	content, err := os.ReadFile(mkAppPath)
	if err != nil {
		t.Fatalf("Failed to read mkApp.nix: %v", err)
	}

	// Check for essential content
	essentialParts := []string{
		"lib",
		"config",
		"pkgs",
	}

	for _, part := range essentialParts {
		if !strings.Contains(string(content), part) {
			t.Errorf("mkApp.nix missing essential content %q", part)
		}
	}

	// Test idempotency - should not overwrite existing file
	originalContent := string(content)
	err = init.EnsureMkAppNix()
	if err != nil {
		t.Fatalf("EnsureMkAppNix() failed on second call: %v", err)
	}

	newContent, err := os.ReadFile(mkAppPath)
	if err != nil {
		t.Fatalf("Failed to read mkApp.nix after second call: %v", err)
	}

	if string(newContent) != originalContent {
		t.Error("EnsureMkAppNix() modified existing file")
	}
}

func TestInitializer_Run(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	// Run full initialization
	err := init.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify all expected files/directories exist
	expectedPaths := []string{
		filepath.Join(tmpDir, "lib"),
		filepath.Join(tmpDir, "lib", "mkApp.nix"),
	}

	for _, path := range expectedPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected path %q does not exist", path)
		}
	}

	// Test idempotency - running again should succeed
	err = init.Run()
	if err != nil {
		t.Errorf("Run() failed on second call: %v", err)
	}
}

func TestInitializer_InvalidFlakePath(t *testing.T) {
	cfg := &internal.Config{
		FlakePath: "/nonexistent/invalid/path",
	}

	init := NewInitializer(cfg)

	err := init.Run()
	if err == nil {
		t.Error("Run() expected error for invalid flake path, got nil")
	}
}

func TestInitializer_VerifyMkAppTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	// Create lib directory
	err := os.MkdirAll(filepath.Join(tmpDir, "lib"), 0o755)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Create mkApp.nix
	err = init.EnsureMkAppNix()
	if err != nil {
		t.Fatalf("EnsureMkAppNix() error = %v", err)
	}

	// Read and verify the template structure
	mkAppPath := filepath.Join(tmpDir, "lib", "mkApp.nix")
	content, err := os.ReadFile(mkAppPath)
	if err != nil {
		t.Fatalf("Failed to read mkApp.nix: %v", err)
	}

	// Verify expected sections exist
	requiredSections := []string{
		"config",
		"pkgs",
		"lib",
		"isLinux",
		"linuxPackages",
		"darwinPackages",
		"optionPath",
	}

	for _, section := range requiredSections {
		if !strings.Contains(string(content), section) {
			t.Errorf("mkApp.nix missing required section %q", section)
		}
	}
}

func TestInitializer_PermissionsCorrect(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	err := init.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Check lib directory permissions
	libInfo, err := os.Stat(filepath.Join(tmpDir, "lib"))
	if err != nil {
		t.Fatalf("Failed to stat lib directory: %v", err)
	}

	// Directory should be readable/writable/executable by owner
	if libInfo.Mode().Perm()&0o700 != 0o700 {
		t.Errorf("lib directory permissions = %o, want at least 0700", libInfo.Mode().Perm())
	}

	// Check mkApp.nix permissions
	mkAppInfo, err := os.Stat(filepath.Join(tmpDir, "lib", "mkApp.nix"))
	if err != nil {
		t.Fatalf("Failed to stat mkApp.nix: %v", err)
	}

	// File should be readable/writable by owner
	if mkAppInfo.Mode().Perm()&0o600 != 0o600 {
		t.Errorf("mkApp.nix permissions = %o, want at least 0600", mkAppInfo.Mode().Perm())
	}
}

func TestInitializer_WithExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create lib directory with custom file
	libPath := filepath.Join(tmpDir, "lib")
	err := os.MkdirAll(libPath, 0o755)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	customFile := filepath.Join(libPath, "custom.nix")
	err = os.WriteFile(customFile, []byte("# Custom content"), 0o644)
	if err != nil {
		t.Fatalf("Failed to create custom file: %v", err)
	}

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	// Run initialization
	err = init.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify custom file still exists
	if _, err := os.Stat(customFile); os.IsNotExist(err) {
		t.Error("Run() deleted existing custom file")
	}

	// Verify custom file content unchanged
	content, err := os.ReadFile(customFile)
	if err != nil {
		t.Fatalf("Failed to read custom file: %v", err)
	}

	if string(content) != "# Custom content" {
		t.Error("Run() modified existing custom file")
	}
}

// Test helper functions
func TestNewInitializer(t *testing.T) {
	cfg := &internal.Config{
		FlakePath: "/test/path",
	}

	init := NewInitializer(cfg)

	if init == nil {
		t.Fatal("NewInitializer() returned nil")
	}

	if init.config != cfg {
		t.Error("NewInitializer() did not store config correctly")
	}
}

// Tests for lib/default.nix functionality

func TestInitializer_EnsureLibDefault_CreateWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	// Create lib directory
	err := os.MkdirAll(filepath.Join(tmpDir, "lib"), 0o755)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	// Ensure lib/default.nix
	err = init.EnsureLibDefault()
	if err != nil {
		t.Fatalf("EnsureLibDefault() error = %v", err)
	}

	// Verify lib/default.nix was created
	libDefaultPath := filepath.Join(tmpDir, "lib", "default.nix")
	if _, err := os.Stat(libDefaultPath); os.IsNotExist(err) {
		t.Error("lib/default.nix was not created")
	}

	// Verify file content has mkApp
	content, err := os.ReadFile(libDefaultPath)
	if err != nil {
		t.Fatalf("Failed to read lib/default.nix: %v", err)
	}

	if !strings.Contains(string(content), "mkApp") {
		t.Error("lib/default.nix does not contain mkApp export")
	}
}

func TestInitializer_EnsureLibDefault_SkipWhenHasMkApp(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	// Create lib directory and default.nix with mkApp
	libPath := filepath.Join(tmpDir, "lib")
	err := os.MkdirAll(libPath, 0o755)
	if err != nil {
		t.Fatalf("Failed to create lib directory: %v", err)
	}

	existingContent := `{ lib }:
{
  mkApp = import ./mkApp.nix { inherit lib; };
  customFunction = x: x + 1;
}`

	libDefaultPath := filepath.Join(libPath, "default.nix")
	err = os.WriteFile(libDefaultPath, []byte(existingContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create lib/default.nix: %v", err)
	}

	// Run EnsureLibDefault
	err = init.EnsureLibDefault()
	if err != nil {
		t.Fatalf("EnsureLibDefault() error = %v", err)
	}

	// Verify content unchanged
	newContent, err := os.ReadFile(libDefaultPath)
	if err != nil {
		t.Fatalf("Failed to read lib/default.nix: %v", err)
	}

	if string(newContent) != existingContent {
		t.Error("EnsureLibDefault() modified file that already had mkApp")
	}
}

// Tests for Nix parser helpers

func TestHasExport_Positive(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		export  string
		want    bool
	}{
		{
			name: "simple export",
			content: `{
  mkApp = import ./mkApp.nix;
}`,
			export: "mkApp",
			want:   true,
		},
		{
			name: "export with spaces",
			content: `{
  mkApp    =    import ./mkApp.nix;
}`,
			export: "mkApp",
			want:   true,
		},
		{
			name: "export with other functions",
			content: `{
  foo = 1;
  mkApp = import ./mkApp.nix;
  bar = 2;
}`,
			export: "mkApp",
			want:   true,
		},
		{
			name: "no export",
			content: `{
  foo = 1;
  bar = 2;
}`,
			export: "mkApp",
			want:   false,
		},
		{
			name: "export in comment",
			content: `{
  # mkApp = import ./mkApp.nix;
  foo = 1;
}`,
			export: "mkApp",
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := hasExport(tc.content, tc.export)
			if got != tc.want {
				t.Errorf("hasExport() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanSafelyModify(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name: "simple attrset",
			content: `{ lib }:
{
  mkApp = import ./mkApp.nix;
}`,
			want: true,
		},
		{
			name: "with let expression",
			content: `{ lib }:
let
  foo = 1;
in {
  mkApp = import ./mkApp.nix;
}`,
			want: false,
		},
		{
			name: "recursive attrset",
			content: `{ lib }:
rec {
  mkApp = import ./mkApp.nix;
}`,
			want: false,
		},
		{
			name:    "no braces",
			content: `import ./something.nix`,
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := canSafelyModify(tc.content)
			if got != tc.want {
				t.Errorf("canSafelyModify() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAddExportToAttrSet(t *testing.T) {
	content := `{ lib }:

{
  customFunction = x: x + 1;
}`

	newContent, err := addExportToAttrSet(content, "mkApp", "import ./mkApp.nix { inherit lib; }")
	if err != nil {
		t.Fatalf("addExportToAttrSet() error = %v", err)
	}

	// Verify mkApp was added
	if !strings.Contains(newContent, "mkApp = import ./mkApp.nix { inherit lib; };") {
		t.Error("addExportToAttrSet() did not add mkApp export")
	}

	// Verify existing content preserved
	if !strings.Contains(newContent, "customFunction") {
		t.Error("addExportToAttrSet() lost existing content")
	}
}

// Tests for flake detection

func TestDetectFlakeLibIntegration_Positive(t *testing.T) {
	testCases := []struct {
		name    string
		flake   string
		pattern string
	}{
		{
			name: "outputs.lib pattern",
			flake: `{
  outputs = { self, nixpkgs }: {
    lib = import ./lib;
  };
}`,
			pattern: "outputs.lib",
		},
		{
			name: "customLib pattern",
			flake: `{
  outputs = { self, nixpkgs }: let
    customLib = import ./lib { lib = nixpkgs.lib; };
  in {
    nixosConfigurations = {};
  };
}`,
			pattern: "let-customLib",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			flakePath := filepath.Join(tmpDir, "flake.nix")
			err := os.WriteFile(flakePath, []byte(tc.flake), 0o644)
			if err != nil {
				t.Fatalf("Failed to create flake.nix: %v", err)
			}

			cfg := &internal.Config{
				FlakePath: tmpDir,
			}
			init := NewInitializer(cfg)

			integrated, pattern := init.DetectFlakeLibIntegration()
			if !integrated {
				t.Error("DetectFlakeLibIntegration() = false, want true")
			}
			if pattern != tc.pattern {
				t.Errorf("DetectFlakeLibIntegration() pattern = %v, want %v", pattern, tc.pattern)
			}
		})
	}
}

func TestDetectFlakeLibIntegration_Negative(t *testing.T) {
	tmpDir := t.TempDir()
	flakePath := filepath.Join(tmpDir, "flake.nix")

	flakeContent := `{
  outputs = { self, nixpkgs }: {
    nixosConfigurations = {};
  };
}`

	err := os.WriteFile(flakePath, []byte(flakeContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create flake.nix: %v", err)
	}

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}
	init := NewInitializer(cfg)

	integrated, _ := init.DetectFlakeLibIntegration()
	if integrated {
		t.Error("DetectFlakeLibIntegration() = true, want false")
	}
}

func TestDetectFlakeLibIntegration_NoFlake(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}
	init := NewInitializer(cfg)

	integrated, pattern := init.DetectFlakeLibIntegration()
	if integrated {
		t.Error("DetectFlakeLibIntegration() = true, want false")
	}
	if pattern != "no-flake" {
		t.Errorf("DetectFlakeLibIntegration() pattern = %v, want no-flake", pattern)
	}
}

// Integration test for Run() with new functionality

func TestInitializer_Run_CreatesLibDefault(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &internal.Config{
		FlakePath: tmpDir,
	}

	init := NewInitializer(cfg)

	// Run initialization - note: this will print setup instructions
	// since we don't have a flake.nix, but should not error
	err := init.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify lib/default.nix was created
	libDefaultPath := filepath.Join(tmpDir, "lib", "default.nix")
	if _, err := os.Stat(libDefaultPath); os.IsNotExist(err) {
		t.Error("Run() did not create lib/default.nix")
	}
}

// Benchmark tests
func BenchmarkInitializer_Run(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tmpDir := b.TempDir()
		cfg := &internal.Config{
			FlakePath: tmpDir,
		}
		init := NewInitializer(cfg)
		_ = init.Run()
	}
}
