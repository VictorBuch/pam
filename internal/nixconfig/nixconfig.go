package nixconfig

import (
	"fmt"
	"regexp"
	"strings"
)

type Config struct {
	content string
}

func NewConfig(content string) *Config {
	return &Config{content: content}
}

func (c *Config) Content() string {
	return c.content
}

func (c *Config) CategoryExists(category string) bool {
	pattern := regexp.QuoteMeta(category) + `\s*=\s*\{`
	matched, _ := regexp.MatchString(pattern, c.content)
	return matched
}

func (c *Config) PackageExistsInCategory(category string, packageName string) bool {
	// Use regex to find category with flexible whitespace
	categoryStart := regexp.QuoteMeta(category) + `\s*=\s*\{`
	re := regexp.MustCompile(categoryStart)
	startLoc := re.FindStringIndex(c.content)

	if startLoc == nil {
		return false
	}

	startPos := startLoc[1]
	endPos := strings.Index(c.content[startPos:], "};")
	if endPos == -1 {
		return false
	}

	categorySection := c.content[startPos : startPos+endPos]
	// Use regex to match package.enable pattern
	packagePattern := regexp.QuoteMeta(packageName) + `\.enable`
	matched, _ := regexp.MatchString(packagePattern, categorySection)
	return matched
}

func (c *Config) EnablePackage(packageName string) bool {
	// Use regex to match flexible whitespace around equals and semicolon
	oldPattern := regexp.QuoteMeta(packageName) + `\.enable\s*=\s*false\s*;`
	newPattern := packageName + ".enable = true;"

	re := regexp.MustCompile(oldPattern)
	if !re.MatchString(c.content) {
		return false // Pattern not found
	}

	c.content = re.ReplaceAllString(c.content, newPattern)
	return true
}

func (c *Config) AddPackageToCategory(category string, packageName string) error {
	// Use regex to find category with flexible whitespace
	categoryStart := regexp.QuoteMeta(category) + `\s*=\s*\{`
	re := regexp.MustCompile(categoryStart)
	startLoc := re.FindStringIndex(c.content)

	if startLoc == nil {
		return fmt.Errorf("category '%s' not found in configuration", category)
	}

	startPos := startLoc[1]
	endPos := strings.Index(c.content[startPos:], "};")
	if endPos == -1 {
		return fmt.Errorf("category '%s' closing brace not found", category)
	}

	absoluteEndPos := startPos + endPos
	packageContent := "\n      " + packageName + ".enable = true;\n    "

	c.content = c.content[:absoluteEndPos] + packageContent + c.content[absoluteEndPos:]
	return nil
}

func (c *Config) CreateCategory(category string, packageName string) error {
	// Use regex to find apps section with flexible whitespace
	appStart := `apps\s*=\s*\{`
	re := regexp.MustCompile(appStart)
	startLoc := re.FindStringIndex(c.content)

	if startLoc == nil {
		return fmt.Errorf("'apps' section not found in configuration")
	}

	insertPos := startLoc[1]
	newCategory := fmt.Sprintf("\n    %s = {\n      %s.enable = true;\n    };\n", category, packageName)

	c.content = c.content[:insertPos] + newCategory + c.content[insertPos:]
	return nil
}

func (c *Config) EnsureAppsSectionExists() error {
	if c.CategoryExists("apps") {
		return nil
	}

	// Find the last closing brace in the file
	lastBrace := strings.LastIndex(c.content, "}")
	if lastBrace == -1 {
		return fmt.Errorf("no closing brace found in configuration")
	}

	// Insert apps section before the last closing brace
	appsSection := "\n  apps = {\n  };\n\n"
	c.content = c.content[:lastBrace] + appsSection + c.content[lastBrace:]
	return nil
}

func (c *Config) AddOrEnablePackage(category, packageName string) error {
	if c.CategoryExists(category) {
		if c.PackageExistsInCategory(category, packageName) {
			c.EnablePackage(packageName)
			return nil
		} else {
			c.AddPackageToCategory(category, packageName)
		}
	} else {
		return c.CreateCategory(category, packageName)
	}
	return nil
}
