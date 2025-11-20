package ui

import (
	"os"
	"path/filepath"
	"testing"

	"pam/internal/types"
)

// TestGetDirNames tests the directory listing functionality
func TestGetDirNames(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create some test directories
	dirs := []string{"dir1", "dir2", "dir3"}
	for _, dir := range dirs {
		err := os.Mkdir(filepath.Join(tmpDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	// Create some files (should be ignored)
	files := []string{"file1.txt", "file2.nix"}
	for _, file := range files {
		err := os.WriteFile(filepath.Join(tmpDir, file), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test GetDirNames
	got, err := GetDirNames(tmpDir)
	if err != nil {
		t.Fatalf("GetDirNames() error = %v", err)
	}

	if len(got) != len(dirs) {
		t.Errorf("GetDirNames() returned %d directories, want %d", len(got), len(dirs))
	}

	// Check that all expected directories are present
	dirMap := make(map[string]bool)
	for _, d := range got {
		dirMap[d] = true
	}

	for _, expectedDir := range dirs {
		if !dirMap[expectedDir] {
			t.Errorf("GetDirNames() missing directory %q", expectedDir)
		}
	}
}

func TestGetDirNames_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	got, err := GetDirNames(tmpDir)
	if err != nil {
		t.Fatalf("GetDirNames() error = %v", err)
	}

	if len(got) != 0 {
		t.Errorf("GetDirNames() returned %d directories for empty dir, want 0", len(got))
	}
}

func TestGetDirNames_NonexistentPath(t *testing.T) {
	_, err := GetDirNames("/nonexistent/path/xyz")
	if err == nil {
		t.Error("GetDirNames() expected error for nonexistent path, got nil")
	}
}

// TestFormatPackageOption tests the package option formatting for selection
func TestFormatPackageOption(t *testing.T) {
	tests := []struct {
		name string
		pkg  *types.Package
		want string
	}{
		{
			name: "standard package",
			pkg: &types.Package{
				PName:   "firefox",
				Version: "120.0",
				System:  "x86_64-linux",
			},
			want: "firefox (120.0) - x86_64-linux",
		},
		{
			name: "darwin package",
			pkg: &types.Package{
				PName:   "vim",
				Version: "9.0.1",
				System:  "aarch64-darwin",
			},
			want: "vim (9.0.1) - aarch64-darwin",
		},
		{
			name: "package with long version",
			pkg: &types.Package{
				PName:   "nodejs",
				Version: "20.10.0-beta.1",
				System:  "x86_64-linux",
			},
			want: "nodejs (20.10.0-beta.1) - x86_64-linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPackageOption(tt.pkg)
			if got != tt.want {
				t.Errorf("FormatPackageOption() = %q, want %q", got, tt.want)
			}
		})
	}
}
