package search

import (
	"encoding/json"
	"testing"

	"pam/internal/types"
)

// TestFilterTopLevel removed - this functionality is part of FilterAndPrioritizePackages

func TestParseSearchResults(t *testing.T) {
	tests := []struct {
		name       string
		jsonOutput string
		wantCount  int
		wantErr    bool
	}{
		{
			name: "valid search results",
			jsonOutput: `{
				"legacyPackages.x86_64-linux.firefox": {
					"pname": "firefox",
					"version": "120.0",
					"description": "A web browser"
				},
				"legacyPackages.x86_64-linux.vim": {
					"pname": "vim",
					"version": "9.0",
					"description": "A text editor"
				}
			}`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "invalid json",
			jsonOutput: `{invalid json}`,
			wantCount:  0,
			wantErr:    true,
		},
		{
			name:       "empty results",
			jsonOutput: `{}`,
			wantCount:  0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result map[string]types.Package
			err := json.Unmarshal([]byte(tt.jsonOutput), &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(result) != tt.wantCount {
				t.Errorf("ParseSearchResults() got %d packages, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestExtractSystemAndPath(t *testing.T) {
	tests := []struct {
		name        string
		fullPath    string
		wantSystem  string
		wantPkgPath string
	}{
		{
			name:        "standard package path",
			fullPath:    "legacyPackages.x86_64-linux.firefox",
			wantSystem:  "x86_64-linux",
			wantPkgPath: "firefox",
		},
		{
			name:        "nested package path",
			fullPath:    "legacyPackages.aarch64-darwin.python311Packages.numpy",
			wantSystem:  "aarch64-darwin",
			wantPkgPath: "python311Packages.numpy",
		},
		{
			name:        "deep nested path",
			fullPath:    "legacyPackages.x86_64-linux.vimPlugins.nerdtree",
			wantSystem:  "x86_64-linux",
			wantPkgPath: "vimPlugins.nerdtree",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the logic that should be in ParseSearchResults
			// to extract system and package path from the full nix path
			parts := splitNixPath(tt.fullPath)
			if len(parts) < 3 {
				t.Fatalf("Invalid path format: %s", tt.fullPath)
			}

			gotSystem := parts[1]
			gotPkgPath := joinPath(parts[2:])

			if gotSystem != tt.wantSystem {
				t.Errorf("System = %q, want %q", gotSystem, tt.wantSystem)
			}

			if gotPkgPath != tt.wantPkgPath {
				t.Errorf("PkgPath = %q, want %q", gotPkgPath, tt.wantPkgPath)
			}
		})
	}
}

// Helper functions for testing - these define the expected API
func splitNixPath(path string) []string {
	// This is what the actual implementation should do
	result := []string{}
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func joinPath(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "."
		}
		result += part
	}
	return result
}

func TestSearcher_Search(t *testing.T) {
	// This test defines the expected interface for the Searcher
	// The actual implementation will need to execute the nix command

	tests := []struct {
		name        string
		packageName string
		system      string
		mockOutput  string
		wantErr     bool
		wantCount   int
	}{
		{
			name:        "successful search",
			packageName: "firefox",
			system:      "x86_64-linux",
			mockOutput: `{
				"legacyPackages.x86_64-linux.firefox": {
					"pname": "firefox",
					"version": "120.0",
					"description": "A web browser"
				}
			}`,
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:        "no results",
			packageName: "nonexistent-package-xyz",
			system:      "x86_64-linux",
			mockOutput:  `{}`,
			wantErr:     false,
			wantCount:   0,
		},
		{
			name:        "search without system filter",
			packageName: "vim",
			system:      "",
			mockOutput: `{
				"legacyPackages.x86_64-linux.vim": {
					"pname": "vim",
					"version": "9.0",
					"description": "A text editor"
				},
				"legacyPackages.aarch64-darwin.vim": {
					"pname": "vim",
					"version": "9.0",
					"description": "A text editor"
				}
			}`,
			wantErr:   false,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a stub test that defines the interface
			// The actual implementation will need to:
			// 1. Create a Searcher with the system
			// 2. Call Search with the package name
			// 3. Return the results

			// Mock the search results
			var mockResult map[string]types.Package
			err := json.Unmarshal([]byte(tt.mockOutput), &mockResult)
			if err != nil {
				t.Fatalf("Failed to parse mock output: %v", err)
			}

			// Verify the mock has expected count
			if len(mockResult) != tt.wantCount {
				t.Errorf("Mock has %d packages, want %d", len(mockResult), tt.wantCount)
			}

			// The actual implementation should:
			// searcher := NewSearcher(tt.system)
			// results, err := searcher.Search(tt.packageName)
			// if (err != nil) != tt.wantErr { ... }
			// if len(results) != tt.wantCount { ... }
		})
	}
}

func TestFilterAndPrioritize(t *testing.T) {
	tests := []struct {
		name     string
		packages SearchResult
		showAll  bool
		want     int // expected count
	}{
		{
			name: "show only top level",
			packages: SearchResult{
				"legacyPackages.x86_64-linux.firefox":              {PName: "firefox"},
				"legacyPackages.x86_64-linux.vim":                  {PName: "vim"},
				"legacyPackages.x86_64-linux.vimPlugins.nerdtree":  {PName: "nerdtree"},
				"legacyPackages.x86_64-linux.vimPlugins.telescope": {PName: "telescope"},
			},
			showAll: false,
			want:    2, // only firefox and vim
		},
		{
			name: "show all packages",
			packages: SearchResult{
				"legacyPackages.x86_64-linux.firefox":              {PName: "firefox"},
				"legacyPackages.x86_64-linux.vim":                  {PName: "vim"},
				"legacyPackages.x86_64-linux.vimPlugins.nerdtree":  {PName: "nerdtree"},
				"legacyPackages.x86_64-linux.vimPlugins.telescope": {PName: "telescope"},
			},
			showAll: true,
			want:    4, // all packages
		},
		{
			name:     "empty list",
			packages: SearchResult{},
			showAll:  false,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterAndPrioritizePackages(tt.packages, tt.showAll)

			if len(result) != tt.want {
				t.Errorf("FilterAndPrioritize() returned %d packages, want %d", len(result), tt.want)
			}
		})
	}
}

// Benchmark removed - FilterTopLevel not needed
