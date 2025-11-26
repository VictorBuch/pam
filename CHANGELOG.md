# Changelog

All notable changes to pam will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-01-26

### Added

- 🎉 Initial release of pam (Package Application Manager)
- Interactive package search with fuzzy matching
- Multi-host NixOS configuration support
- Automatic module generation with cross-platform support (Linux/Darwin)
- Smart lib/ directory setup with mkApp helper
- Config file support (~/.config/pam/config)
- Folder recursion for organizing packages by category
- Optional editor integration (auto-open after install)
- Homebrew cask support for Darwin packages
- Embedded templates (no external dependencies)
- Comprehensive test coverage

### Features

**Package Management:**
- Search packages with `nix search` integration
- Filter by system architecture
- Support for stable and unstable channels
- Automatic platform-specific package handling

**Code Generation:**
- Generate NixOS modules from templates
- mkApp helper for minimal boilerplate
- Auto-detect module paths from file locations
- Cross-platform package declarations

**Library Setup:**
- Automatic lib/ directory initialization
- Smart merge for lib/default.nix
- Interactive prompts for modifications
- Flake integration detection and guidance

**Configuration:**
- YAML-based config file support
- Configurable paths (NIXOS_ROOT, MODULES_ROOT)
- Default editor preferences
- System architecture overrides

### Documentation

- Comprehensive README with examples
- Feature documentation in feats/
- Test coverage for all major components

[0.1.0]: https://github.com/VictorBuch/pam/releases/tag/v0.1.0
