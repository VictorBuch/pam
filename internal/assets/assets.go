package assets

import (
	_ "embed"
	"strings"

	"pam/internal/types"
)

//go:embed templates/packageTemplate.nix
var packageTemplate string

//go:embed templates/mkApp.nix
var mkApp string

func FillPackageTemplate(pkg *types.Package, useHomebrew bool) string {
	var linuxPackage string
	var darwinPackage string
	var homebrewPackage string

	if strings.Contains(pkg.System, "linux") {
		linuxPackage = "pkgs." + pkg.FullPath
	} else if strings.Contains(pkg.System, "darwin") {
		if useHomebrew {
			homebrewPackage = pkg.PName
		} else {
			darwinPackage = "pkgs." + pkg.FullPath
		}
	}
	replacer := strings.NewReplacer("LinuxPackage", linuxPackage, "DarwinPackage", darwinPackage, "HomebrewPackage", homebrewPackage, "PackageName", pkg.PName, "PackageDescription", pkg.Description)
	filledTemplate := replacer.Replace(packageTemplate)
	// Clean up double spaces that occur when placeholders are replaced with empty strings
	filledTemplate = strings.ReplaceAll(filledTemplate, "  ", " ")
	return filledTemplate
}

func GetPackageTemplate() string {
	return packageTemplate
}

func GetMkApp() string {
	return mkApp
}
