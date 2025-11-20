package setup

import (
	"os"
	"path/filepath"

	"pam/internal"
	"pam/internal/assets"
)

type Initializer struct {
	config *internal.Config
}

func NewInitializer(cfg *internal.Config) *Initializer {
	return &Initializer{config: cfg}
}

func (i *Initializer) EnsureLibDirectory() error {
	libPath := filepath.Join(i.config.FlakePath, "lib")
	return os.MkdirAll(libPath, 0o755)
}

func (i *Initializer) EnsureMkAppNix() error {
	mkAppPath := filepath.Join(i.config.FlakePath, "lib", "mkApp.nix")

	if _, err := os.Stat(mkAppPath); err == nil {
		return nil // File exists, don't overwrite
	}

	template := assets.GetMkApp()
	return os.WriteFile(mkAppPath, []byte(template), 0o644)
}

func (i *Initializer) Run() error {
	if err := i.EnsureLibDirectory(); err != nil {
		return err
	}

	if err := i.EnsureMkAppNix(); err != nil {
		return err
	}

	// TODO: make this also check the default.nix for lib dir
	// make it also add the registration for the lib functions in the flake.nix
	return nil
}
