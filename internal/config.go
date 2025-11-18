package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"
)

type Config struct {
	FlakePath        string `yaml:"flake_path"`
	DefaultSystem    string `yaml:"default_system"`
	DefaultModuleDir string `yaml:"default_module_dir"`
	DefaultHostDir   string `yaml:"default_host_dir"`
}

func (c *Config) Validate() error {
	if c.FlakePath == "" {
		return fmt.Errorf("flake_path is required. Please run setup or edit ~/.config/pam/config.yaml")
	}
	if _, err := os.Stat(c.FlakePath); os.IsNotExist(err) {
		return fmt.Errorf("flake_path '%s' does not exist", c.FlakePath)
	}
	return nil
}

func (c *Config) Save() error {
	path := getConfigPath()
	yaml, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(path)
	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, []byte(yaml), 0o644)
	if err != nil {
		return err
	}
	return nil
}

func interactiveSetup(cfg *Config) error {
	fmt.Println("\n🔧 Welcome to pam setup! Let's configure your flake path.")

	var flakePath string
	var system string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Where is your NixOs/nix-darwin flake located?").Placeholder("~/nixos").Value(&flakePath).Validate(
				func(s string) error {
					if s == "" {
						return fmt.Errorf("Flake path cannot be empty")
					}
					return nil
				},
			),
		),
		huh.NewGroup(
			huh.NewInput().Title("Input your system architecture").Placeholder("x86_64-linux or aarch64-darwin etc...").Value(&system),
		),
	)
	err := form.Run()
	if err != nil {
		return err
	}
	cfg.FlakePath = flakePath
	cfg.DefaultSystem = system
	return nil
}

func Default() *Config {
	return &Config{
		FlakePath:        "",
		DefaultSystem:    "",
		DefaultModuleDir: "modules/apps",
		DefaultHostDir:   "hosts",
	}
}

func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return homeDir
	}

	return strings.Replace(path, "~", homeDir, 1)
}

func getOrCreateConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// file does not exist

			configPath := filepath.Dir(path)
			err := os.MkdirAll(configPath, 0o755)
			if err != nil {
				return nil, err
			}

			defaults := Default()
			err = interactiveSetup(defaults)
			if err != nil {
				return nil, err
			}
			err = defaults.Save()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	configYaml, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	config, err := getOrCreateConfig(configPath)
	if err != nil {
		return nil, err
	}

	if config.FlakePath == "" {
		err = interactiveSetup(config)
		if err != nil {
			return nil, err
		}
		err = config.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save config: %w",
				err)
		}
	}

	config.FlakePath = expandPath(config.FlakePath)
	err = config.Validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Sprintf("Error getting home dir: %s", err)
	}
	configPath := filepath.Join(homeDir, ".config", "pam", "config.yaml")
	return configPath
}
