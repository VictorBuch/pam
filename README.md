# PAM - Package Application Manager

[![Release](https://img.shields.io/github/v/release/VictorBuch/pam)](https://github.com/VictorBuch/pam/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/VictorBuch/pam)](go.mod)
[![Build Status](https://github.com/VictorBuch/pam/actions/workflows/go.yml/badge.svg)](https://github.com/VictorBuch/pam/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A fast, interactive CLI tool for searching and managing Nix packages across multiple NixOS/nix-darwin machines with automatic module generation.

## ✨ Features

- 🔍 **Fast Package Search** - Search nixpkgs with instant results
- 📦 **Smart Installation** - Automatically creates module files and updates host configurations
- 🖥️ **Multi-Host Support** - Install packages on multiple machines simultaneously
- 🎨 **Interactive UI** - Beautiful terminal interface powered by Bubble Tea
- ⚙️ **Flexible Configuration** - YAML-based config with interactive first-time setup
- 🏗️ **Module Generation** - Auto-generates Nix modules from templates
- 🍎 **macOS Support** - Optional Homebrew integration for Darwin systems
- 🎯 **Architecture-Aware** - Filter packages by system architecture

## 📋 Prerequisites

- Nix with flakes enabled
- A NixOS or nix-darwin flake-based configuration

## 🚀 Installation

### Download Pre-built Binary (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/VictorBuch/pam/releases):

**Linux (x86_64):**

```bash
wget https://github.com/VictorBuch/pam/releases/latest/download/pam_0.1.0_linux_amd64.tar.gz
tar -xzf pam_0.1.0_linux_amd64.tar.gz
sudo mv pam /usr/local/bin/
```

**Linux (ARM64):**

```bash
wget https://github.com/VictorBuch/pam/releases/latest/download/pam_0.1.0_linux_arm64.tar.gz
tar -xzf pam_0.1.0_linux_arm64.tar.gz
sudo mv pam /usr/local/bin/
```

**macOS (Intel):**

```bash
wget https://github.com/VictorBuch/pam/releases/latest/download/pam_0.1.0_darwin_amd64.tar.gz
tar -xzf pam_0.1.0_darwin_amd64.tar.gz
sudo mv pam /usr/local/bin/
```

**macOS (Apple Silicon):**

```bash
wget https://github.com/VictorBuch/pam/releases/latest/download/pam_0.1.0_darwin_arm64.tar.gz
tar -xzf pam_0.1.0_darwin_arm64.tar.gz
sudo mv pam /usr/local/bin/
```

### From Source (For Development)

```bash
git clone https://github.com/VictorBuch/pam.git
cd pam
go build -o pam
sudo mv pam /usr/local/bin/
```

### First Run

On first run, PAM will guide you through an interactive setup to configure your flake path:

```bash
pam install neovim
```

You'll be prompted to enter:

- Your NixOS/nix-darwin flake location (e.g., `~/nixos-config`)
- Default system architecture (optional, e.g., `x86_64-linux` or `aarch64-darwin`)

## ⚙️ Configuration

PAM uses a YAML configuration file located at `~/.config/pam/config.yaml`.

### Configuration File Structure

```yaml
# Path to your NixOS/nix-darwin flake (required)
flake_path: "~/nixos-config"

# Default system architecture for package search (optional)
default_system: "x86_64-linux"

# Module directory relative to flake_path (default: modules/apps)
default_module_dir: "modules/apps"

# Hosts directory relative to flake_path (default: hosts)
default_host_dir: "hosts"
```

### Configuration Options

| Field                | Required | Description                           | Example                              |
| -------------------- | -------- | ------------------------------------- | ------------------------------------ |
| `flake_path`         | ✅ Yes   | Path to your Nix flake directory      | `~/nixos-config` or `/home/user/nix` |
| `default_system`     | ❌ No    | Default system architecture to search | `x86_64-linux`, `aarch64-darwin`     |
| `default_module_dir` | ❌ No    | Where to store generated modules      | `modules/apps` (default)             |
| `default_host_dir`   | ❌ No    | Where your host configurations live   | `hosts` (default)                    |

### Manual Configuration

You can manually create or edit the config file:

```bash
mkdir -p ~/.config/pam
nano ~/.config/pam/config.yaml
```

**Note:** The `flake_path` supports tilde (`~`) expansion for home directory references.

## 📖 Usage

### Basic Package Installation

Search and install a package:

```bash
pam install neovim
```

This will:

1. Search nixpkgs for "neovim"
2. Let you select the package from search results
3. Choose which module category to place it in
4. Select which hosts to enable it on
5. Generate a Nix module file
6. Update your host configurations

### Advanced Options

```bash
# Show all packages including plugins
pam install neovim --show-all

# Search for specific system architecture
pam install htop --system x86_64-linux

# Use Homebrew for macOS packages (Darwin only)
pam install firefox --brew
```

### Command Flags

- `-a, --show-all` - Show all packages including plugins and nested packages
- `-s, --system <arch>` - Target specific system architecture
- `-b, --brew` - Use Homebrew cask instead of Nix package (macOS only)

## 🏗️ How It Works

1. **Package Search**: Uses `nix search` to find packages in nixpkgs
2. **Library Setup**: Automatically sets up `lib/` directory with `mkApp` helper function
3. **Module Generation**: Creates Nix modules based on embedded templates
4. **Configuration Update**: Automatically updates `configuration.nix` in selected hosts
5. **Category Management**: Organizes packages into categories (e.g., development, utilities)
6. **Multi-System Support**: Handles both Linux and Darwin packages intelligently

### Project Structure

Your Nix flake should follow this structure:

```
~/nixos-config/
├── flake.nix
├── lib/
│   └── mkApp.nix          # Generated from template
├── modules/
│   └── apps/
│       ├── development/
│       │   └── neovim.nix  # Generated modules
│       └── utilities/
│           └── htop.nix
└── hosts/
    ├── desktop/
    │   └── configuration.nix
    └── laptop/
        └── configuration.nix
```

## 📚 Documentation

- **[CHANGELOG.md](CHANGELOG.md)** - Version history and changes

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- UI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Huh](https://github.com/charmbracelet/huh)
- YAML parsing with [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
- Automated releases with [GoReleaser](https://goreleaser.com/)

**Made for the NixOS community**
