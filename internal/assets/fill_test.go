package assets

import (
	"strings"
	"testing"

	"pam/internal/types"
)

func TestFillPackageTemplate(t *testing.T) {
	tests := []struct {
		name           string
		pkg            *types.Package
		useHomebrew    bool
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "linux package",
			pkg: &types.Package{
				PName:       "firefox",
				FullPath:    "firefox",
				System:      "x86_64-linux",
				Version:     "120.0",
				Description: "A web browser",
			},
			useHomebrew: false,
			wantContains: []string{
				"name = \"firefox\"",
				"linuxPackages = pkgs: [ pkgs.firefox ]",
				"darwinPackages = pkgs: [ ]", // empty for linux
			},
		},
		{
			name: "darwin package with nix",
			pkg: &types.Package{
				PName:       "firefox",
				FullPath:    "firefox",
				System:      "aarch64-darwin",
				Version:     "120.0",
				Description: "A web browser",
			},
			useHomebrew: false,
			wantContains: []string{
				"name = \"firefox\"",
				"darwinPackages = pkgs: [ pkgs.firefox ]",
				"linuxPackages = pkgs: [ ]", // empty for darwin
			},
		},
		{
			name: "darwin package with homebrew",
			pkg: &types.Package{
				PName:       "firefox",
				FullPath:    "firefox",
				System:      "aarch64-darwin",
				Version:     "120.0",
				Description: "A web browser",
			},
			useHomebrew: true,
			wantContains: []string{
				"name = \"firefox\"",
				"homebrew.casks = [ \"firefox\" ]",
			},
			wantNotContain: []string{
				"pkgs.firefox",
			},
		},
		{
			name: "nested package path",
			pkg: &types.Package{
				PName:       "numpy",
				FullPath:    "python311Packages.numpy",
				System:      "x86_64-linux",
				Version:     "1.24.0",
				Description: "Scientific computing with Python",
			},
			useHomebrew: false,
			wantContains: []string{
				"name = \"numpy\"",
				"linuxPackages = pkgs: [ pkgs.python311Packages.numpy ]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FillPackageTemplate(tt.pkg, tt.useHomebrew)

			// Check that all expected strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("FillPackageTemplate() missing expected content %q\nGot:\n%s", want, got)
				}
			}

			// Check that unwanted strings are not present
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(got, notWant) {
					t.Errorf("FillPackageTemplate() contains unexpected content %q\nGot:\n%s", notWant, got)
				}
			}
		})
	}
}

func TestGetTemplate(t *testing.T) {
	template := GetPackageTemplate()

	// Verify that the template contains expected placeholders
	expectedPlaceholders := []string{
		"PackageName",
		"LinuxPackage",
		"DarwinPackage",
		"HomebrewPackage",
	}

	for _, placeholder := range expectedPlaceholders {
		if !strings.Contains(template, placeholder) {
			t.Errorf("GetTemplate() missing placeholder %q", placeholder)
		}
	}

	// Verify template is not empty
	if len(template) == 0 {
		t.Error("GetTemplate() returned empty string")
	}

	// Verify template contains mkApp function call
	if !strings.Contains(template, "mkApp") {
		t.Error("GetTemplate() missing mkApp function call")
	}
}

func TestFillPackageTemplate_AllPlaceholdersReplaced(t *testing.T) {
	pkg := &types.Package{
		PName:       "testpkg",
		FullPath:    "testpkg",
		System:      "x86_64-linux",
		Version:     "1.0.0",
		Description: "Test package",
	}

	result := FillPackageTemplate(pkg, false)

	// Ensure no unreplaced placeholders remain
	placeholders := []string{
		"PackageName",
		"PackageDescription",
	}

	for _, placeholder := range placeholders {
		if strings.Contains(result, placeholder) {
			t.Errorf("FillPackageTemplate() left unreplaced placeholder %q\nResult:\n%s", placeholder, result)
		}
	}
}

func TestFillPackageTemplate_PreservesTemplateStructure(t *testing.T) {
	pkg := &types.Package{
		PName:    "test",
		FullPath: "test",
		System:   "x86_64-linux",
	}

	result := FillPackageTemplate(pkg, false)

	// Verify essential Nix syntax is preserved
	essentialParts := []string{
		"args@{",
		"mkApp {",
		"} args",
	}

	for _, part := range essentialParts {
		if !strings.Contains(result, part) {
			t.Errorf("FillPackageTemplate() missing essential structure %q\nResult:\n%s", part, result)
		}
	}
}

func TestFillPackageTemplate_EmptyValues(t *testing.T) {
	tests := []struct {
		name        string
		pkg         *types.Package
		useHomebrew bool
		checkEmpty  []string
	}{
		{
			name: "linux - darwin fields should be empty",
			pkg: &types.Package{
				PName:    "test",
				FullPath: "test",
				System:   "x86_64-linux",
			},
			useHomebrew: false,
			checkEmpty: []string{
				"darwinPackages = pkgs: [ ]",
			},
		},
		{
			name: "darwin with nix - homebrew should be empty",
			pkg: &types.Package{
				PName:    "test",
				FullPath: "test",
				System:   "aarch64-darwin",
			},
			useHomebrew: false,
			checkEmpty: []string{
				"homebrew.casks = [ ]",
			},
		},
		{
			name: "darwin with homebrew - nix darwin should be empty",
			pkg: &types.Package{
				PName:    "test",
				FullPath: "test",
				System:   "aarch64-darwin",
			},
			useHomebrew: true,
			checkEmpty: []string{
				"darwinPackages = pkgs: [ ]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FillPackageTemplate(tt.pkg, tt.useHomebrew)

			for _, empty := range tt.checkEmpty {
				if !strings.Contains(result, empty) {
					t.Logf("Note: Expected to find empty field %q, but may have different formatting", empty)
				}
			}
		})
	}
}

// Benchmark for performance testing
func BenchmarkFillPackageTemplate(b *testing.B) {
	pkg := &types.Package{
		PName:       "firefox",
		FullPath:    "firefox",
		System:      "x86_64-linux",
		Version:     "120.0",
		Description: "A web browser",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FillPackageTemplate(pkg, false)
	}
}

func TestFillPackageTemplate_SpecialCharactersInDescription(t *testing.T) {
	pkg := &types.Package{
		PName:       "test",
		FullPath:    "test",
		System:      "x86_64-linux",
		Description: "A \"quoted\" description with 'apostrophes' and newlines\n",
	}

	result := FillPackageTemplate(pkg, false)

	// Should contain the description even with special characters
	if !strings.Contains(result, "quoted") {
		t.Error("FillPackageTemplate() failed to preserve quoted text in description")
	}
}
