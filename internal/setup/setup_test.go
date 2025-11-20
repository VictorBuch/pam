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
