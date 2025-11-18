# Road to v1

## Completed Features ✅

- [x] Make the UX better
- [x] Make the host selection a multiple choice selection
- [x] Ask the user if they want to edit the new module once finished with creation
- [x] Migrate to huh for better form handling

## In Progress Features 🚧

- [ ] Support having a config file like other great software in /home/<usr>/.config/nap/config (YAML or TOML recommended - could configure flake path, default editor, etc.)
- [ ] Support searching multiple system architectures (e.g., both x86_64-linux and aarch64-darwin)
- [ ] Complete initial setup: lib/mkApp.nix integration
  - [x] Create lib/mkApp.nix if missing
  - [ ] Create/update lib/default.nix to export mkApp (install.go:178-179)
  - [ ] Check flake.nix for lib registration
  - [ ] Add lib to flake outputs if missing
- [ ] Rescan selected folder to see if nested dirs exist and reprompt (install.go:239)

## Code Quality TODOs 🔧

- [x] Fix critical bug: content variable in install loop (was using template instead of host config)
- [x] Move moduleFilePath write outside of host loop
- [x] Add newlines to error messages
- [x] Fix file permission octal notation
- [ ] Make NIXOS_ROOT configurable (env var or config file)
- [ ] Add unit tests for string manipulation functions
- [ ] Better error handling (use errors.Is/As)

## Future Enhancements 💡

- [ ] Add --dry-run flag to preview changes
- [ ] Support removing/disabling packages
- [ ] Add package search with filters (category, license, etc.)
- [ ] Generate flake.lock diff after install
- [ ] Add rollback functionality
