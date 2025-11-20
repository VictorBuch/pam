package nixconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	content := "test content"
	editor := NewConfig(content)

	if editor == nil {
		t.Fatal("NewConfig returned nil")
	}

	if editor.Content() != content {
		t.Errorf("Content() = %q, want %q", editor.Content(), content)
	}
}

func TestConfig_CategoryExists(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		category string
		want     bool
	}{
		{
			name:     "standard format",
			content:  "apps = {",
			category: "apps",
			want:     true,
		},
		{
			name:     "no spaces around equals",
			content:  "apps={",
			category: "apps",
			want:     true,
		},
		{
			name:     "extra spaces",
			content:  "apps  =  {",
			category: "apps",
			want:     true,
		},
		{
			name:     "category not found",
			content:  "other = {",
			category: "apps",
			want:     false,
		},
		{
			name:     "nested category",
			content:  "  browsers = {\n    firefox.enable = true;\n  };",
			category: "browsers",
			want:     true,
		},
		{
			name:     "multiple categories",
			content:  "apps = {\n  browsers = {\n  };\n}",
			category: "browsers",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewConfig(tt.content)
			got := editor.CategoryExists(tt.category)
			if got != tt.want {
				t.Errorf("CategoryExists(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

func TestConfig_PackageExistsInCategory(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		category    string
		packageName string
		want        bool
	}{
		{
			name: "package exists enabled",
			content: `apps = {
  browsers = {
    firefox.enable = true;
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			want:        true,
		},
		{
			name: "package exists disabled",
			content: `apps = {
  browsers = {
    firefox.enable = false;
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			want:        true,
		},
		{
			name: "package not in category",
			content: `apps = {
  browsers = {
    firefox.enable = true;
  };
}`,
			category:    "browsers",
			packageName: "chrome",
			want:        false,
		},
		{
			name: "category not found",
			content: `apps = {
  browsers = {
  };
}`,
			category:    "editors",
			packageName: "vim",
			want:        false,
		},
		{
			name: "package in different category",
			content: `apps = {
  browsers = {
    firefox.enable = true;
  };
  editors = {
    vim.enable = true;
  };
}`,
			category:    "browsers",
			packageName: "vim",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewConfig(tt.content)
			got := editor.PackageExistsInCategory(tt.category, tt.packageName)
			if got != tt.want {
				t.Errorf("PackageExistsInCategory(%q, %q) = %v, want %v",
					tt.category, tt.packageName, got, tt.want)
			}
		})
	}
}

func TestConfig_EnablePackage(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		packageName string
		want        bool
		wantContent string
	}{
		{
			name:        "enable disabled package",
			content:     "firefox.enable = false;",
			packageName: "firefox",
			want:        true,
			wantContent: "firefox.enable = true;",
		},
		{
			name:        "enable with extra spaces",
			content:     "firefox.enable  =  false  ;",
			packageName: "firefox",
			want:        true,
			wantContent: "firefox.enable = true;",
		},
		{
			name:        "package already enabled",
			content:     "firefox.enable = true;",
			packageName: "firefox",
			want:        false,
			wantContent: "firefox.enable = true;",
		},
		{
			name:        "package not found",
			content:     "chrome.enable = false;",
			packageName: "firefox",
			want:        false,
			wantContent: "chrome.enable = false;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewConfig(tt.content)
			got := editor.EnablePackage(tt.packageName)
			if got != tt.want {
				t.Errorf("EnablePackage(%q) = %v, want %v", tt.packageName, got, tt.want)
			}
			if editor.Content() != tt.wantContent {
				t.Errorf("Content after EnablePackage = %q, want %q",
					editor.Content(), tt.wantContent)
			}
		})
	}
}

func TestConfig_AddPackageToCategory(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		category    string
		packageName string
		wantErr     bool
		wantContain string
	}{
		{
			name: "add package to existing category",
			content: `apps = {
  browsers = {
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: "firefox.enable = true;",
		},
		{
			name: "add to category with existing packages",
			content: `apps = {
  browsers = {
    chrome.enable = true;
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: "firefox.enable = true;",
		},
		{
			name:        "category not found",
			content:     "apps = {\n}",
			category:    "browsers",
			packageName: "firefox",
			wantErr:     true,
			wantContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewConfig(tt.content)
			err := editor.AddPackageToCategory(tt.category, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPackageToCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.wantContain != "" {
				if !strings.Contains(editor.Content(), tt.wantContain) {
					t.Errorf("Content after AddPackageToCategory doesn't contain %q\nGot: %q",
						tt.wantContain, editor.Content())
				}
			}
		})
	}
}

func TestConfig_CreateCategory(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		category    string
		packageName string
		wantErr     bool
		wantContain []string
	}{
		{
			name: "create new category in empty apps",
			content: `apps = {
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: []string{"browsers = {", "firefox.enable = true;"},
		},
		{
			name: "create category with existing categories",
			content: `apps = {
  editors = {
    vim.enable = true;
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: []string{"browsers = {", "firefox.enable = true;"},
		},
		{
			name:        "apps section not found",
			content:     "{ }",
			category:    "browsers",
			packageName: "firefox",
			wantErr:     true,
			wantContain: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewConfig(tt.content)
			err := editor.CreateCategory(tt.category, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				content := editor.Content()
				for _, want := range tt.wantContain {
					if !strings.Contains(content, want) {
						t.Errorf("Content after CreateCategory doesn't contain %q\nGot: %q",
							want, content)
					}
				}
			}
		})
	}
}

func TestConfig_AddOrEnablePackage(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		category    string
		packageName string
		wantErr     bool
		wantContain []string
	}{
		{
			name: "category doesn't exist - create it",
			content: `apps = {
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: []string{"browsers = {", "firefox.enable = true;"},
		},
		{
			name: "category exists, package doesn't - add it",
			content: `apps = {
  browsers = {
    chrome.enable = true;
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: []string{"firefox.enable = true;", "chrome.enable = true;"},
		},
		{
			name: "package exists disabled - enable it",
			content: `apps = {
  browsers = {
    firefox.enable = false;
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: []string{"firefox.enable = true;"},
		},
		{
			name: "package already enabled - no change",
			content: `apps = {
  browsers = {
    firefox.enable = true;
  };
}`,
			category:    "browsers",
			packageName: "firefox",
			wantErr:     false,
			wantContain: []string{"firefox.enable = true;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewConfig(tt.content)
			err := editor.AddOrEnablePackage(tt.category, tt.packageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrEnablePackage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				content := editor.Content()
				for _, want := range tt.wantContain {
					if !strings.Contains(content, want) {
						t.Errorf("Content doesn't contain %q\nGot: %q", want, content)
					}
				}
			}
		})
	}
}

func TestConfig_WithRealConfigFile(t *testing.T) {
	// Test with actual sample config file
	testdataPath := filepath.Join("..", "..", "testdata", "sample_config.nix")
	data, err := os.ReadFile(testdataPath)
	if err != nil {
		t.Skipf("Skipping test with real config file: %v", err)
		return
	}

	editor := NewConfig(string(data))

	// Test that we can find existing categories
	if !editor.CategoryExists("browsers") {
		t.Error("Expected browsers category to exist")
	}

	if !editor.CategoryExists("editors") {
		t.Error("Expected editors category to exist")
	}

	// Test that we can find existing packages
	if !editor.PackageExistsInCategory("browsers", "firefox") {
		t.Error("Expected firefox to exist in browsers")
	}

	// Test adding a new package to existing category
	err = editor.AddPackageToCategory("browsers", "brave")
	if err != nil {
		t.Errorf("Failed to add brave to browsers: %v", err)
	}

	if !strings.Contains(editor.Content(), "brave.enable = true;") {
		t.Error("brave was not added to config")
	}
}

func TestConfig_WithWhitespaceVariations(t *testing.T) {
	testdataPath := filepath.Join("..", "..", "testdata", "whitespace_config.nix")
	data, err := os.ReadFile(testdataPath)
	if err != nil {
		t.Skipf("Skipping test with whitespace config: %v", err)
		return
	}

	editor := NewConfig(string(data))

	// Should handle no spaces: apps={
	if !editor.CategoryExists("apps") {
		t.Error("Failed to find apps with no spaces around =")
	}

	// Should handle extra spaces: browsers  =  {
	if !editor.CategoryExists("browsers") {
		t.Error("Failed to find browsers with extra spaces")
	}
}
